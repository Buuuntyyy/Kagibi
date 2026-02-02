/**
 * Upload Queue Manager Service
 * Handles concurrent file uploads with encryption, S3 multipart, and progress tracking
 * 
 * Architecture:
 * - Limited concurrency (default 3 files simultaneously)
 * - Each file has its own encryption → upload pipeline
 * - Memory-efficient: chunks are freed after upload
 * - Fault-tolerant: one file failure doesn't stop others
 */

import { useUploadStore, UploadStatus } from '../stores/uploads'
import { useFileStore } from '../stores/files'
import { useAuthStore } from '../stores/auth'
import { MultipartUploadManager, PART_SIZE, UploadState } from './multipartUpload'
import { encryptChunkWorker, generateBaseNonce } from './crypto'
import { generateMasterKey, wrapMasterKey, deriveKeyFromToken } from './crypto'
import { generatePreview } from './previewGenerator'
import api from '../api'

// Configuration
const MAX_CONCURRENT_FILES = 3
const ENCRYPTION_PROGRESS_WEIGHT = 0.3 // 30% of progress is encryption

/**
 * Upload Queue Manager - Singleton Service
 */
class UploadQueueManager {
  constructor() {
    this.isProcessing = false
    this.processingPromise = null
  }

  /**
   * Add files to queue and start processing
   * @param {FileList|File[]} files 
   * @param {string} targetPath 
   */
  async addFiles(files, targetPath) {
    const uploadStore = useUploadStore()
    
    // Add all files to queue
    const ids = uploadStore.addMultipleToQueue(files, targetPath)
    
    // Start processing if not already running
    this.startProcessing()
    
    return ids
  }

  /**
   * Add single file to queue
   * @param {File} file 
   * @param {string} targetPath 
   */
  async addFile(file, targetPath) {
    return this.addFiles([file], targetPath)
  }

  /**
   * Start the queue processor
   */
  startProcessing() {
    if (this.isProcessing) return
    
    this.isProcessing = true
    this.processingPromise = this.processQueue()
  }

  /**
   * Main queue processing loop
   * Maintains MAX_CONCURRENT_FILES active uploads
   */
  async processQueue() {
    const uploadStore = useUploadStore()
    
    while (true) {
      // Check if we can start more uploads
      const activeCount = uploadStore.getActiveCount()
      const availableSlots = MAX_CONCURRENT_FILES - activeCount
      
      if (availableSlots <= 0) {
        // Wait a bit before checking again
        await this.sleep(100)
        continue
      }
      
      // Get next pending uploads
      const pendingUploads = []
      for (let i = 0; i < availableSlots; i++) {
        const next = uploadStore.getNextPending()
        if (next) {
          // Mark as starting to prevent re-picking
          uploadStore.setStatus(next.id, UploadStatus.ENCRYPTING)
          pendingUploads.push(next)
        }
      }
      
      if (pendingUploads.length === 0 && activeCount === 0) {
        // No more work to do
        break
      }
      
      if (pendingUploads.length > 0) {
        // Start processing these uploads in parallel (fire and forget)
        for (const upload of pendingUploads) {
          this.processUpload(upload).catch(err => {
            console.error(`Upload ${upload.id} failed:`, err)
          })
        }
      }
      
      // Small delay before next iteration
      await this.sleep(50)
    }
    
    this.isProcessing = false
    this.processingPromise = null
  }

  /**
   * Process a single file upload
   * @param {Object} uploadItem 
   */
  async processUpload(uploadItem) {
    const uploadStore = useUploadStore()
    const fileStore = useFileStore()
    const authStore = useAuthStore()
    
    const { id, file, targetPath } = uploadItem
    
    try {
      // Verify authentication
      if (!authStore.isAuthenticated || !authStore.masterKey) {
        throw new Error('Not authenticated')
      }

      uploadStore.updateUpload(id, { startTime: Date.now() })

      // === STEP 1: Generate Preview (if applicable) ===
      let previewID = null
      const previewBlob = await generatePreview(file)
      if (previewBlob) {
        try {
          const safeName = (file.name || 'file').replace(/[^a-zA-Z0-9.-]/g, '_')
          const previewName = `preview_${safeName}.jpg`
          const previewFile = new File([previewBlob], previewName, { type: 'image/jpeg' })
          
          // Upload preview silently (doesn't count toward main progress)
          const previewResult = await this.uploadSingleFile(previewFile, targetPath, authStore.masterKey, {
            isPreview: true,
            silent: true
          })
          
          if (previewResult?.ID) {
            previewID = previewResult.ID
          }
        } catch (e) {
          console.warn('Preview upload failed:', e)
        }
      }

      // === STEP 2: Generate File Key ===
      const fileKey = await generateMasterKey()
      const encryptedFileKey = await wrapMasterKey(fileKey, authStore.masterKey)

      // === STEP 3: Check Active Shares ===
      let shareKeysMap = {}
      try {
        const shareRes = await api.get('/shares/check-path', { params: { path: targetPath } })
        const activeShares = shareRes.data.shares || []
        
        for (const share of activeShares) {
          const shareKey = await deriveKeyFromToken(share.Token)
          const encryptedForShare = await wrapMasterKey(fileKey, shareKey)
          shareKeysMap[share.ID] = encryptedForShare
        }
      } catch (e) {
        console.warn('Error checking shares:', e)
      }

      // === STEP 4: Encrypt Chunks ===
      uploadStore.setStatus(id, UploadStatus.ENCRYPTING)
      
      const totalParts = Math.ceil(file.size / PART_SIZE)
      const encryptedChunks = []
      const baseNonce = generateBaseNonce()
      
      let offset = 0
      let chunkIndex = 0
      
      while (offset < file.size) {
        // Check if cancelled
        const currentUpload = uploadStore.uploads.get(id)
        if (currentUpload?.status === UploadStatus.CANCELLED) {
          throw new Error('Upload cancelled')
        }
        
        const chunkBlob = file.slice(offset, offset + PART_SIZE)
        const chunkArrayBuffer = await chunkBlob.arrayBuffer()
        
        // Encrypt chunk
        const encryptedChunkBlob = await encryptChunkWorker(chunkArrayBuffer, fileKey, chunkIndex, baseNonce)
        encryptedChunks.push(encryptedChunkBlob)
        
        offset += PART_SIZE
        chunkIndex++
        
        // Update progress (0-30% is encryption)
        const encryptProgress = Math.round((chunkIndex / totalParts) * ENCRYPTION_PROGRESS_WEIGHT * 100)
        uploadStore.setProgress(id, encryptProgress, 0)
      }

      // Calculate total encrypted size
      const totalEncryptedSize = encryptedChunks.reduce((sum, chunk) => {
        return sum + (chunk.size || chunk.byteLength || 0)
      }, 0)
      
      uploadStore.updateUpload(id, { encryptedSize: totalEncryptedSize, totalBytes: totalEncryptedSize })

      // === STEP 5: Initiate & Upload to S3 ===
      uploadStore.setStatus(id, UploadStatus.UPLOADING)
      
      const manager = new MultipartUploadManager({
        onProgress: (percent, uploaded, total) => {
          // Map 0-100% upload progress to 30-95% total progress
          const uploadProgress = ENCRYPTION_PROGRESS_WEIGHT * 100 + (percent * (1 - ENCRYPTION_PROGRESS_WEIGHT - 0.05))
          uploadStore.setProgress(id, Math.round(uploadProgress), uploaded)
        },
        onStateChange: (state) => {
          if (state === UploadState.FAILED) {
            uploadStore.setStatus(id, UploadStatus.FAILED)
          }
        },
        onError: (error, partNumber) => {
          console.error(`Part ${partNumber} error:`, error)
        }
      })
      
      uploadStore.updateUpload(id, { manager })
      
      // Initiate multipart upload
      await manager.initiate(
        file.name,
        targetPath,
        'application/octet-stream',
        totalEncryptedSize,
        encryptedFileKey
      )

      // Upload parts
      const completedParts = await manager.uploadParts(encryptedChunks)
      
      // Free memory
      encryptedChunks.length = 0

      // === STEP 6: Complete Upload ===
      uploadStore.setStatus(id, UploadStatus.COMPLETING)
      uploadStore.setProgress(id, 95, totalEncryptedSize)
      
      const result = await manager.complete(completedParts, {
        fileName: file.name,
        filePath: targetPath,
        totalSize: totalEncryptedSize,
        contentType: 'application/octet-stream',
        encryptedKey: encryptedFileKey,
        shareKeys: Object.keys(shareKeysMap).length > 0 ? JSON.stringify(shareKeysMap) : '',
        previewId: previewID,
        isPreview: false
      })

      // === SUCCESS ===
      uploadStore.setCompleted(id, result)
      
      // Refresh file list
      fileStore.fetchItems(fileStore.currentPath)
      
      // Update user quota
      authStore.fetchUser()
      
      // Add to history
      if (result?.file) {
        fileStore.addToHistory({
          ...result.file,
          type: 'file',
          displayName: result.file.Name
        })
      }

      return result

    } catch (error) {
      console.error(`Upload ${id} failed:`, error)
      
      // Attempt cleanup
      const upload = uploadStore.uploads.get(id)
      if (upload?.manager) {
        try {
          await upload.manager.abort()
        } catch (e) {
          console.warn('Abort cleanup failed:', e)
        }
      }
      
      uploadStore.setFailed(id, error)
      throw error
    }
  }

  /**
   * Upload a single file (used for previews)
   * @param {File} file 
   * @param {string} targetPath 
   * @param {CryptoKey} masterKey 
   * @param {Object} options 
   */
  async uploadSingleFile(file, targetPath, masterKey, options = {}) {
    const { isPreview = false, silent = false } = options
    
    // Generate file key
    const fileKey = await generateMasterKey()
    const encryptedFileKey = await wrapMasterKey(fileKey, masterKey)
    
    // Encrypt
    const totalParts = Math.ceil(file.size / PART_SIZE)
    const encryptedChunks = []
    const baseNonce = generateBaseNonce()
    
    let offset = 0
    let chunkIndex = 0
    
    while (offset < file.size) {
      const chunkBlob = file.slice(offset, offset + PART_SIZE)
      const chunkArrayBuffer = await chunkBlob.arrayBuffer()
      const encryptedChunkBlob = await encryptChunkWorker(chunkArrayBuffer, fileKey, chunkIndex, baseNonce)
      encryptedChunks.push(encryptedChunkBlob)
      offset += PART_SIZE
      chunkIndex++
    }
    
    const totalEncryptedSize = encryptedChunks.reduce((sum, chunk) => sum + (chunk.size || 0), 0)
    
    // Upload
    const manager = new MultipartUploadManager({})
    
    await manager.initiate(
      file.name,
      targetPath,
      'application/octet-stream',
      totalEncryptedSize,
      encryptedFileKey
    )
    
    const completedParts = await manager.uploadParts(encryptedChunks)
    
    const result = await manager.complete(completedParts, {
      fileName: file.name,
      filePath: targetPath,
      totalSize: totalEncryptedSize,
      contentType: 'application/octet-stream',
      encryptedKey: encryptedFileKey,
      shareKeys: '',
      previewId: null,
      isPreview: isPreview
    })
    
    return result?.file
  }

  /**
   * Utility sleep function
   */
  sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms))
  }
}

// Singleton instance
export const uploadQueueManager = new UploadQueueManager()

// Export for direct import
export default uploadQueueManager
