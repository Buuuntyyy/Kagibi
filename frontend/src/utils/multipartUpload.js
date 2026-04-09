/**
 * Direct-to-S3 Multipart Upload Manager
 * Handles parallel uploads with retry logic and progress tracking
 */
import api from '../api'

// Configuration
const MAX_CONCURRENT_WORKERS = 3
const MAX_RETRIES = 3
const INITIAL_RETRY_DELAY = 1000 // 1 second
const PART_SIZE = 10 * 1024 * 1024 // 10MB per part (matches crypto CHUNK_SIZE)

/**
 * Upload state enum
 */
export const UploadState = {
  PENDING: 'pending',
  UPLOADING: 'uploading',
  COMPLETED: 'completed',
  FAILED: 'failed',
  ABORTED: 'aborted'
}

/**
 * Represents a single part upload task
 */
class PartUploadTask {
  constructor(partNumber, url, data, size) {
    this.partNumber = partNumber
    this.url = url
    this.data = data
    this.size = size
    this.etag = null
    this.state = UploadState.PENDING
    this.retryCount = 0
    this.progress = 0
    this.controller = null
  }
}

/**
 * Multipart Upload Manager
 * Orchestrates parallel uploads with retry and progress tracking
 */
export class MultipartUploadManager {
  constructor(options = {}) {
    this.uploadId = null
    this.key = null
    this.parts = []
    this.totalSize = 0
    this.uploadedBytes = 0
    this.state = UploadState.PENDING
    this.onProgress = options.onProgress || (() => {})
    this.onStateChange = options.onStateChange || (() => {})
    this.onError = options.onError || (() => {})
    this.abortController = new AbortController()
  }

  /**
   * Initialize multipart upload with the backend
   */
  async initiate(fileName, filePath, contentType, totalSize, encryptedKey) {
    const totalParts = Math.ceil(totalSize / PART_SIZE)
    
    try {
      const response = await api.post('/files/multipart/initiate', {
        file_name: fileName,
        file_path: filePath,
        content_type: contentType,
        total_size: totalSize,
        total_parts: totalParts,
        encrypted_key: encryptedKey
      })

      this.uploadId = response.data.upload_id
      this.key = response.data.key
      this.totalSize = totalSize
      
      // Create part tasks with presigned URLs
      this.parts = response.data.presigned_urls.map(p => 
        new PartUploadTask(p.part_number, p.url, null, 0)
      )
      
      this.state = UploadState.UPLOADING
      this.onStateChange(this.state)
      
      return {
        uploadId: this.uploadId,
        key: this.key,
        presignedUrls: response.data.presigned_urls
      }
    } catch (error) {
      this.state = UploadState.FAILED
      this.onStateChange(this.state)
      throw error
    }
  }

  /**
   * Upload encrypted chunks directly to S3
   * @param {Array<Blob>} encryptedChunks - Array of encrypted data blobs
   */
  async uploadParts(encryptedChunks) {
    if (this.state === UploadState.ABORTED) {
      throw new Error('Upload was aborted')
    }

    // Assign data to parts
    encryptedChunks.forEach((chunk, index) => {
      if (this.parts[index]) {
        this.parts[index].data = chunk
        this.parts[index].size = chunk.size || chunk.byteLength
      }
    })

    // Create upload worker pool
    const pendingParts = [...this.parts]
    const activeTasks = new Map()
    const completedParts = []

    return new Promise((resolve, reject) => {
      const processNext = () => {
        // Check for abort
        if (this.state === UploadState.ABORTED) {
          activeTasks.forEach((_, controller) => controller.abort())
          reject(new Error('Upload aborted'))
          return
        }

        // Start new tasks if pool has capacity
        while (activeTasks.size < MAX_CONCURRENT_WORKERS && pendingParts.length > 0) {
          const part = pendingParts.shift()
          const controller = new AbortController()
          activeTasks.set(part.partNumber, controller)
          part.controller = controller

          this.uploadSinglePart(part, controller.signal)
            .then(etag => {
              part.etag = etag
              part.state = UploadState.COMPLETED
              completedParts.push({
                part_number: part.partNumber,
                etag: etag
              })
              activeTasks.delete(part.partNumber)
              
              // Update progress
              this.uploadedBytes += part.size
              this.updateProgress()
              
              // Process next or complete
              if (completedParts.length === this.parts.length) {
                resolve(completedParts)
              } else {
                processNext()
              }
            })
            .catch(error => {
              activeTasks.delete(part.partNumber)
              if (this.state === UploadState.ABORTED) return

              if (part.retryCount < MAX_RETRIES) {
                part.retryCount++
                part.state = UploadState.PENDING
                const delay = INITIAL_RETRY_DELAY * Math.pow(2, part.retryCount - 1)
                console.warn(`Part ${part.partNumber} failed, retrying in ${delay}ms (attempt ${part.retryCount}/${MAX_RETRIES})`)
                this.schedulePartRetry(part, delay, pendingParts, processNext, reject)
              } else {
                part.state = UploadState.FAILED
                this.state = UploadState.FAILED
                this.onStateChange(this.state)
                this.onError(error, part.partNumber)
                reject(new Error(`Part ${part.partNumber} failed after ${MAX_RETRIES} retries`))
              }
            })
        }
      }

      processNext()
    })
  }

  /**
   * Schedule a retry for a failed part after the given delay, refreshing its URL first.
   */
  schedulePartRetry(part, delay, pendingParts, processNext, reject) {
    setTimeout(() => {
      this.refreshPartUrl(part)
        .then(() => {
          pendingParts.unshift(part)
          processNext()
        })
        .catch(refreshError => {
          console.error('Failed to refresh URL:', refreshError)
          reject(new Error(`Part ${part.partNumber} failed after ${MAX_RETRIES} retries`))
        })
    }, delay)
  }

  /**
   * Upload a single part to S3 with progress tracking
   */
  async uploadSinglePart(part, signal) {
    return new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest()
      
      xhr.open('PUT', part.url, true)
      
      // Track upload progress for this part
      xhr.upload.onprogress = (event) => {
        if (event.lengthComputable) {
          part.progress = event.loaded
          this.updateProgress()
        }
      }

      xhr.onload = () => {
        if (xhr.status >= 200 && xhr.status < 300) {
          // CRITICAL: Capture ETag from response headers
          let etag = xhr.getResponseHeader('ETag')
          if (etag) {
            // Remove quotes if present
            etag = etag.replace(/"/g, '')
            resolve(etag)
          } else {
            reject(new Error('ETag not found in response'))
          }
        } else {
          reject(new Error(`Upload failed with status ${xhr.status}`))
        }
      }

      xhr.onerror = () => {
        reject(new Error('Network error during upload'))
      }

      xhr.ontimeout = () => {
        reject(new Error('Upload timed out'))
      }

      // Handle abort
      if (signal) {
        signal.addEventListener('abort', () => {
          xhr.abort()
          reject(new Error('Upload aborted'))
        })
      }

      // Send the data
      xhr.send(part.data)
    })
  }

  /**
   * Refresh presigned URL for a part (for retries)
   */
  async refreshPartUrl(part) {
    const response = await api.post('/files/multipart/refresh-url', {
      upload_id: this.uploadId,
      key: this.key,
      part_number: part.partNumber,
      part_size: part.size
    })
    part.url = response.data.url
  }

  /**
   * Calculate and emit overall progress
   */
  updateProgress() {
    // Sum uploaded bytes from completed parts + current progress of in-flight parts
    let totalProgress = 0
    for (const part of this.parts) {
      if (part.state === UploadState.COMPLETED) {
        totalProgress += part.size
      } else if (part.state === UploadState.UPLOADING) {
        totalProgress += part.progress
      }
    }
    
    const percentage = this.totalSize > 0 
      ? Math.min(Math.round((totalProgress / this.totalSize) * 100), 100)
      : 0
    
    this.onProgress(percentage, totalProgress, this.totalSize)
  }

  /**
   * Complete the multipart upload
   */
  async complete(completedParts, metadata) {
    try {
      const response = await api.post('/files/multipart/complete', {
        upload_id: this.uploadId,
        key: this.key,
        parts: completedParts,
        file_name: metadata.fileName,
        file_path: metadata.filePath,
        total_size: metadata.totalSize,
        content_type: metadata.contentType,
        encrypted_key: metadata.encryptedKey,
        share_keys: metadata.shareKeys || '',
        preview_id: metadata.previewId || null,
        is_preview: metadata.isPreview || false
      })

      this.state = UploadState.COMPLETED
      this.onStateChange(this.state)
      
      return response.data
    } catch (error) {
      this.state = UploadState.FAILED
      this.onStateChange(this.state)
      throw error
    }
  }

  /**
   * Abort the multipart upload
   */
  async abort() {
    this.state = UploadState.ABORTED
    this.onStateChange(this.state)
    this.abortController.abort()

    // Cancel all in-flight requests
    for (const part of this.parts) {
      if (part.controller) {
        part.controller.abort()
      }
    }

    // Notify backend to cleanup
    if (this.uploadId && this.key) {
      try {
        await api.post('/files/multipart/abort', {
          upload_id: this.uploadId,
          key: this.key
        })
      } catch (error) {
        console.error('Failed to abort upload on backend:', error)
      }
    }
  }
}

/**
 * Helper function for simple uploads
 * Handles the full flow: encrypt -> initiate -> upload -> complete
 */
export async function uploadFileMultipart(
  file,
  filePath,
  encryptFunction,
  fileKey,
  encryptedFileKey,
  options = {}
) {
  const {
    onProgress = () => {},
    onStateChange = () => {},
    onError = () => {},
    shareKeys = {},
    previewId = null,
    isPreview = false
  } = options

  const manager = new MultipartUploadManager({
    onProgress,
    onStateChange,
    onError
  })

  try {
    // Calculate parts
    const totalParts = Math.ceil(file.size / PART_SIZE)
    
    // Encrypt all chunks first
    onStateChange('encrypting')
    const encryptedChunks = []
    let offset = 0
    let chunkIndex = 0
    
    while (offset < file.size) {
      const chunkBlob = file.slice(offset, offset + PART_SIZE)
      const chunkArrayBuffer = await chunkBlob.arrayBuffer()
      const encryptedChunk = await encryptFunction(chunkArrayBuffer, fileKey, chunkIndex)
      encryptedChunks.push(encryptedChunk)
      offset += PART_SIZE
      chunkIndex++
    }

    // Calculate total encrypted size
    const totalEncryptedSize = encryptedChunks.reduce((sum, chunk) => sum + chunk.size, 0)

    // Initiate multipart upload
    await manager.initiate(
      file.name,
      filePath,
      'application/octet-stream', // Always octet-stream for encrypted data
      totalEncryptedSize,
      encryptedFileKey
    )

    // Upload parts
    const completedParts = await manager.uploadParts(encryptedChunks)

    // Complete upload
    const result = await manager.complete(completedParts, {
      fileName: file.name,
      filePath: filePath,
      totalSize: totalEncryptedSize,
      contentType: 'application/octet-stream',
      encryptedKey: encryptedFileKey,
      shareKeys: Object.keys(shareKeys).length > 0 ? JSON.stringify(shareKeys) : '',
      previewId: previewId,
      isPreview: isPreview
    })

    return result
  } catch (error) {
    // Attempt to abort on error
    try {
      await manager.abort()
    } catch (abortError) {
      console.error('Error during abort:', abortError)
    }
    throw error
  }
}

export { PART_SIZE }
