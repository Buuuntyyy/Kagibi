/**
 * Upload Queue Store - Pinia Store for Multi-File Upload Management
 * Manages concurrent uploads with progress tracking and state management
 */
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

/**
 * Upload item status enum
 */
export const UploadStatus = {
  PENDING: 'pending',       // Waiting in queue
  ENCRYPTING: 'encrypting', // Client-side encryption in progress
  UPLOADING: 'uploading',   // Uploading to S3
  COMPLETING: 'completing', // Finalizing with backend
  COMPLETED: 'completed',   // Successfully finished
  FAILED: 'failed',         // Error occurred
  CANCELLED: 'cancelled'    // User cancelled
}

/**
 * Generate unique ID for upload items
 */
let uploadIdCounter = 0
function generateUploadId() {
  return `upload_${Date.now()}_${++uploadIdCounter}`
}

export const useUploadStore = defineStore('uploads', {
  state: () => ({
    /** @type {Map<string, UploadItem>} */
    uploads: new Map(),

    /** Global settings */
    maxConcurrentFiles: 3,
    maxConcurrentChunksPerFile: 3,

    /** Queue processing state */
    isProcessing: false,

    /** Show/hide upload manager panel */
    showManager: false,

    /** Folder creation phase (before file uploads begin) */
    folderCreation: {
      active: false,
      total: 0,
      done: 0
    }
  }),

  getters: {
    /**
     * All uploads as array (for reactivity in Vue)
     */
    uploadList: (state) => Array.from(state.uploads.values()),
    
    /**
     * Overall progress (0-100)
     */
    overallProgress: (state) => {
      const uploads = Array.from(state.uploads.values())
      if (uploads.length === 0) return 0
      
      const totalProgress = uploads.reduce((sum, u) => sum + u.progress, 0)
      return Math.round(totalProgress / uploads.length)
    },
    
    /**
     * Has any active uploads
     */
    hasActiveUploads: (state) => 
      Array.from(state.uploads.values()).some(u => 
        [UploadStatus.PENDING, UploadStatus.ENCRYPTING, UploadStatus.UPLOADING, UploadStatus.COMPLETING].includes(u.status)
      ),
    
    /**
     * Count by status
     */
    counts: (state) => {
      const uploads = Array.from(state.uploads.values())
      return {
        total: uploads.length,
        pending: uploads.filter(u => u.status === UploadStatus.PENDING).length,
        active: uploads.filter(u => [UploadStatus.ENCRYPTING, UploadStatus.UPLOADING, UploadStatus.COMPLETING].includes(u.status)).length,
        completed: uploads.filter(u => u.status === UploadStatus.COMPLETED).length,
        failed: uploads.filter(u => u.status === UploadStatus.FAILED).length
      }
    }
  },

  actions: {
    /**
     * Add a file to the upload queue
     * @param {File} file - The file to upload
     * @param {string} targetPath - Destination path
     * @returns {string} Upload ID
     */
    addToQueue(file, targetPath) {
      const id = generateUploadId()
      
      const uploadItem = {
        id,
        file,
        fileName: file.name,
        fileSize: file.size,
        mimeType: file.type || 'application/octet-stream',
        targetPath,
        status: UploadStatus.PENDING,
        progress: 0,
        uploadedBytes: 0,
        totalBytes: file.size,
        encryptedSize: 0,
        error: null,
        startTime: null,
        endTime: null,
        manager: null, // MultipartUploadManager reference
        abortController: null
      }
      
      this.uploads.set(id, uploadItem)
      
      // Auto-show manager when adding files
      if (!this.showManager) {
        this.showManager = true
      }
      
      return id
    },
    
    /**
     * Add multiple files to queue
     * @param {FileList|File[]} files 
     * @param {string} targetPath 
     * @returns {string[]} Upload IDs
     */
    addMultipleToQueue(files, targetPath) {
      const ids = []
      for (const file of files) {
        ids.push(this.addToQueue(file, targetPath))
      }
      return ids
    },
    
    /**
     * Update upload item state
     * Mutates in place to avoid spreading large objects (File, manager) on every progress tick.
     */
    updateUpload(id, updates) {
      const upload = this.uploads.get(id)
      if (upload) {
        Object.assign(upload, updates)
        this.uploads.set(id, upload)
      }
    },
    
    /**
     * Set upload status
     */
    setStatus(id, status) {
      this.updateUpload(id, { status })
    },
    
    /**
     * Set upload progress
     * @param {string} id 
     * @param {number} progress - 0-100
     * @param {number} uploadedBytes 
     */
    setProgress(id, progress, uploadedBytes = 0) {
      this.updateUpload(id, { progress, uploadedBytes })
    },
    
    /**
     * Mark upload as failed
     */
    setFailed(id, error) {
      this.updateUpload(id, { 
        status: UploadStatus.FAILED, 
        error: error?.message || error || 'Unknown error',
        endTime: Date.now()
      })
    },
    
    /**
     * Mark upload as completed. Releases File and manager references to free memory.
     */
    setCompleted(id, result = null) {
      this.updateUpload(id, {
        status: UploadStatus.COMPLETED,
        progress: 100,
        endTime: Date.now(),
        result,
        file: null,    // release File DOM reference
        manager: null  // release MultipartUploadManager + all part blobs
      })
    },
    
    /**
     * Cancel a specific upload
     */
    async cancelUpload(id) {
      const upload = this.uploads.get(id)
      if (!upload) return
      
      // If pending, just remove from queue
      if (upload.status === UploadStatus.PENDING) {
        this.updateUpload(id, { status: UploadStatus.CANCELLED })
        return
      }
      
      // If active, abort the manager
      if (upload.manager) {
        try {
          await upload.manager.abort()
        } catch (e) {
          console.error('Error aborting upload:', e)
        }
      }
      
      if (upload.abortController) {
        upload.abortController.abort()
      }
      
      this.updateUpload(id, { status: UploadStatus.CANCELLED, endTime: Date.now() })
    },
    
    /**
     * Cancel all pending and active uploads
     */
    async cancelAll() {
      const activeIds = Array.from(this.uploads.values())
        .filter(u => [UploadStatus.PENDING, UploadStatus.ENCRYPTING, UploadStatus.UPLOADING].includes(u.status))
        .map(u => u.id)
      
      await Promise.all(activeIds.map(id => this.cancelUpload(id)))
    },
    
    /**
     * Retry a failed upload
     */
    retryUpload(id) {
      const upload = this.uploads.get(id)
      if (!upload || upload.status !== UploadStatus.FAILED) return
      
      this.updateUpload(id, {
        status: UploadStatus.PENDING,
        progress: 0,
        uploadedBytes: 0,
        error: null,
        startTime: null,
        endTime: null,
        manager: null
      })
    },
    
    /**
     * Remove upload from list (completed/failed/cancelled only)
     */
    removeUpload(id) {
      const upload = this.uploads.get(id)
      if (!upload) return
      
      if ([UploadStatus.COMPLETED, UploadStatus.FAILED, UploadStatus.CANCELLED].includes(upload.status)) {
        this.uploads.delete(id)
      }
    },
    
    /**
     * Clear all completed uploads
     */
    clearCompleted() {
      for (const [id, upload] of this.uploads.entries()) {
        if (upload.status === UploadStatus.COMPLETED) {
          this.uploads.delete(id)
        }
      }
    },
    
    /**
     * Clear all uploads (including active - will cancel them)
     */
    async clearAll() {
      await this.cancelAll()
      this.uploads.clear()
    },
    
    /**
     * Toggle manager visibility
     */
    toggleManager() {
      this.showManager = !this.showManager
    },

    /**
     * Folder creation phase progress tracking
     */
    startFolderCreation(total) {
      this.folderCreation = { active: true, total, done: 0 }
      if (!this.showManager) this.showManager = true
    },
    incrementFolderCreation() {
      this.folderCreation.done++
    },
    endFolderCreation() {
      this.folderCreation.active = false
    },
    
    /**
     * Get next pending upload (for queue processor)
     */
    getNextPending() {
      for (const upload of this.uploads.values()) {
        if (upload.status === UploadStatus.PENDING) {
          return upload
        }
      }
      return null
    },
    
    /**
     * Count currently active uploads
     */
    getActiveCount() {
      let count = 0
      for (const upload of this.uploads.values()) {
        if ([UploadStatus.ENCRYPTING, UploadStatus.UPLOADING, UploadStatus.COMPLETING].includes(upload.status)) {
          count++
        }
      }
      return count
    }
  }
})
