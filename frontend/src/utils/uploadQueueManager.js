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
   * Dequeue up to `slots` pending items, marking them as in-progress.
   */
  dequeuePending(uploadStore, slots) {
    const pendingUploads = []
    for (let i = 0; i < slots; i++) {
      const next = uploadStore.getNextPending()
      if (!next) break
      uploadStore.setStatus(next.id, UploadStatus.ENCRYPTING)
      pendingUploads.push(next)
    }
    return pendingUploads
  }

  /**
   * Main queue processing loop
   * Maintains MAX_CONCURRENT_FILES active uploads
   */
  async processQueue() {
    const uploadStore = useUploadStore()

    while (true) {
      const activeCount = uploadStore.getActiveCount()
      const availableSlots = MAX_CONCURRENT_FILES - activeCount

      if (availableSlots <= 0) {
        await this.sleep(100)
        continue
      }

      const pendingUploads = this.dequeuePending(uploadStore, availableSlots)

      if (pendingUploads.length === 0 && activeCount === 0) break

      for (const upload of pendingUploads) {
        this.processUpload(upload).catch(err => {
          console.error(`Upload ${upload.id} failed:`, err)
        })
      }

      await this.sleep(50)
    }

    this.isProcessing = false
    this.processingPromise = null
  }

  /**
   * Upload a preview image for the given file, returning its ID or null.
   */
  async uploadPreview(file, targetPath, masterKey) {
    const previewBlob = await generatePreview(file)
    if (!previewBlob) return null
    try {
      const safeName = (file.name || 'file').replace(/[^a-zA-Z0-9.-]/g, '_')
      const previewFile = new File([previewBlob], `preview_${safeName}.jpg`, { type: 'image/jpeg' })
      const previewResult = await this.uploadSingleFile(previewFile, targetPath, masterKey, {
        isPreview: true, silent: true
      })
      return previewResult?.ID ?? null
    } catch (e) {
      console.warn('Preview upload failed:', e)
      return null
    }
  }

  /**
   * Build a share keys map: { shareID -> encryptedFileKey } for all active shares at path.
   */
  async buildShareKeysMap(fileKey, targetPath) {
    const shareKeysMap = {}
    try {
      const shareRes = await api.get('/shares/check-path', { params: { path: targetPath } })
      for (const share of (shareRes.data.shares || [])) {
        const shareKey = await deriveKeyFromToken(share.Token)
        shareKeysMap[share.ID] = await wrapMasterKey(fileKey, shareKey)
      }
    } catch (e) {
      console.warn('Error checking shares:', e)
    }
    return shareKeysMap
  }

  /**
   * Encrypt all chunks of a file and return the encrypted blobs array.
   */
  async encryptFileChunks(uploadStore, id, file, fileKey) {
    const totalParts = Math.ceil(file.size / PART_SIZE)
    const encryptedChunks = []
    const baseNonce = generateBaseNonce()
    let offset = 0
    let chunkIndex = 0
    while (offset < file.size) {
      const currentUpload = uploadStore.uploads.get(id)
      if (currentUpload?.status === UploadStatus.CANCELLED) throw new Error('Upload cancelled')
      const chunkArrayBuffer = await file.slice(offset, offset + PART_SIZE).arrayBuffer()
      encryptedChunks.push(await encryptChunkWorker(chunkArrayBuffer, fileKey, chunkIndex, baseNonce))
      offset += PART_SIZE
      chunkIndex++
      const encryptProgress = Math.round((chunkIndex / totalParts) * ENCRYPTION_PROGRESS_WEIGHT * 100)
      uploadStore.setProgress(id, encryptProgress, 0)
    }
    return encryptedChunks
  }

  /**
   * Create a MultipartUploadManager wired to uploadStore progress updates.
   */
  createUploadManager(uploadStore, id) {
    return new MultipartUploadManager({
      onProgress: (percent, uploaded) => {
        const uploadProgress = ENCRYPTION_PROGRESS_WEIGHT * 100 + (percent * (1 - ENCRYPTION_PROGRESS_WEIGHT - 0.05))
        uploadStore.setProgress(id, Math.round(uploadProgress), uploaded)
      },
      onStateChange: (state) => {
        if (state === UploadState.FAILED) uploadStore.setStatus(id, UploadStatus.FAILED)
      },
      onError: (error, partNumber) => {
        console.error(`Part ${partNumber} error:`, error)
      }
    })
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
      if (!authStore.isAuthenticated || !authStore.masterKey) throw new Error('Not authenticated')
      uploadStore.updateUpload(id, { startTime: Date.now() })

      const previewID = await this.uploadPreview(file, targetPath, authStore.masterKey)

      const fileKey = await generateMasterKey()
      const encryptedFileKey = await wrapMasterKey(fileKey, authStore.masterKey)
      const shareKeysMap = await this.buildShareKeysMap(fileKey, targetPath)

      uploadStore.setStatus(id, UploadStatus.ENCRYPTING)
      const encryptedChunks = await this.encryptFileChunks(uploadStore, id, file, fileKey)
      const totalEncryptedSize = encryptedChunks.reduce((sum, c) => sum + (c.size || c.byteLength || 0), 0)
      uploadStore.updateUpload(id, { encryptedSize: totalEncryptedSize, totalBytes: totalEncryptedSize })

      uploadStore.setStatus(id, UploadStatus.UPLOADING)
      const manager = this.createUploadManager(uploadStore, id)
      uploadStore.updateUpload(id, { manager })

      await manager.initiate(file.name, targetPath, 'application/octet-stream', totalEncryptedSize, encryptedFileKey)
      const completedParts = await manager.uploadParts(encryptedChunks)
      encryptedChunks.length = 0

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

      uploadStore.setCompleted(id, result)
      fileStore.fetchItems(fileStore.currentPath)
      authStore.fetchUser()
      if (result?.file) {
        fileStore.addToHistory({ ...result.file, type: 'file', displayName: result.file.Name })
      }
      return result

    } catch (error) {
      console.error(`Upload ${id} failed:`, error)
      const upload = uploadStore.uploads.get(id)
      if (upload?.manager) {
        try { await upload.manager.abort() } catch (e) { console.warn('Abort cleanup failed:', e) }
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
