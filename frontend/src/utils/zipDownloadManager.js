/**
 * Zero-Knowledge Multi-File Download Manager
 * 
 * Handles batch downloading and streaming ZIP creation with:
 * - Concurrent downloads (4 files max)
 * - Streaming decryption
 * - Service Worker ZIP assembly
 * - Progress tracking with ETA calculation
 * - Backpressure handling
 */

import api from '../api'
import { useAuthStore } from '../stores/auth'
import { CHUNK_SIZE, IV_LENGTH } from './crypto'
import sodium from 'libsodium-wrappers-sumo'

// Constants
const MAX_CONCURRENT_DOWNLOADS = 4
const TAG_LENGTH = 16
const ENCRYPTED_CHUNK_OVERHEAD = IV_LENGTH + TAG_LENGTH
const ENCRYPTED_CHUNK_SIZE = CHUNK_SIZE + ENCRYPTED_CHUNK_OVERHEAD

/**
 * Download state enum
 */
export const DownloadStatus = {
  IDLE: 'idle',
  INITIALIZING: 'initializing',
  FETCHING_TREE: 'fetching_tree',
  GENERATING_URLS: 'generating_urls',
  DOWNLOADING: 'downloading',
  DECRYPTING: 'decrypting',
  FINALIZING: 'finalizing',
  COMPLETED: 'completed',
  ERROR: 'error',
  ABORTED: 'aborted'
}

/**
 * Single file download state
 */
class FileDownloadTask {
  constructor(fileInfo) {
    this.id = fileInfo.id
    this.name = fileInfo.name
    this.relativePath = fileInfo.relative_path || fileInfo.relativePath
    this.size = fileInfo.size
    this.encryptedKey = fileInfo.encrypted_key || fileInfo.encryptedKey
    this.mimeType = fileInfo.mime_type || fileInfo.mimeType
    this.presignedUrl = null
    this.status = 'pending'
    this.progress = 0
    this.bytesDownloaded = 0
    this.abortController = null
    this.error = null
  }
}

/**
 * Manages multi-file downloads with ZIP streaming
 */
/**
 * Throttle interval for progress updates (ms)
 * Using ~60fps for smooth UI updates without overwhelming the main thread
 */
const PROGRESS_THROTTLE_MS = 16

class ZipDownloadManager {
  constructor() {
    this.sessionId = null
    this.status = DownloadStatus.IDLE
    this.files = []
    this.totalSize = 0
    this.totalFiles = 0
    this.processedFiles = 0
    this.bytesDownloaded = 0
    this.bytesWritten = 0
    this.startTime = null
    this.worker = null
    this.messageChannel = null
    this.callbacks = {
      onProgress: () => {},
      onStatusChange: () => {},
      onError: () => {},
      onComplete: () => {}
    }
    this.activeDownloads = new Set()
    this.pendingQueue = []
    this.cryptoKey = null
    this.aborted = false
    this.isSingleFile = false  // True when downloading a single file (no ZIP)
    
    // Throttling state for progress updates
    this._lastProgressTime = 0
    this._progressScheduled = false
    this._rafId = null
  }

  /**
   * Initialize the download manager
   */
  async init(callbacks = {}) {
    this.callbacks = { ...this.callbacks, ...callbacks }
    
    // Register service worker if not already registered
    if ('serviceWorker' in navigator) {
      try {
        const registration = await navigator.serviceWorker.register('/download-worker.js', {
          scope: '/download-stream/'
        })
        await navigator.serviceWorker.ready
        this.worker = registration.active || registration.waiting || registration.installing
        //console.log('[DownloadManager] Service Worker registered')
      } catch (error) {
        console.warn('[DownloadManager] Service Worker registration failed, using fallback:', error)
        this.worker = null
      }
    }
  }

  /**
   * Reset all state for a new download
   */
  reset() {
    // Cancel any pending progress update
    if (this._rafId) {
      cancelAnimationFrame(this._rafId)
      this._rafId = null
    }
    
    this.sessionId = null
    this.status = DownloadStatus.IDLE
    this.files = []
    this.totalSize = 0
    this.totalFiles = 0
    this.processedFiles = 0
    this.bytesDownloaded = 0
    this.bytesWritten = 0
    this.startTime = null
    this.activeDownloads.clear()
    this.pendingQueue = []
    this.cryptoKey = null
    this.aborted = false
    this.isSingleFile = false
    
    // Reset throttling state
    this._lastProgressTime = 0
    this._progressScheduled = false
  }

  /**
   * Download a single folder as ZIP
   */
  async downloadFolder(folderId, folderName) {
    // Reset all state before starting new download
    this.reset()
    
    this.setStatus(DownloadStatus.INITIALIZING)
    this.sessionId = `folder-${folderId}-${Date.now()}`
    this.startTime = Date.now()
    this.aborted = false
    
    try {
      // Get folder tree
      this.setStatus(DownloadStatus.FETCHING_TREE)
      const treeResponse = await api.get(`/folders/${folderId}/tree`)
      const tree = treeResponse.data
      
      this.totalFiles = tree.total_files
      this.totalSize = tree.total_size
      this.files = tree.files.map(f => new FileDownloadTask(f))
      
      if (this.files.length === 0) {
        throw new Error('Le dossier est vide')
      }
      
      // Get crypto key
      const authStore = useAuthStore()
      this.cryptoKey = authStore.masterKey
      if (!this.cryptoKey) {
        throw new Error('Clé de déchiffrement non disponible')
      }
      
      // Generate batch presigned URLs
      this.setStatus(DownloadStatus.GENERATING_URLS)
      const fileIds = this.files.map(f => f.id)
      const presignResponse = await api.post('/files/batch-presign', { file_ids: fileIds })
      
      // Map URLs to files
      const urlMap = new Map(presignResponse.data.urls.map(u => [u.file_id, u]))
      for (const file of this.files) {
        const urlInfo = urlMap.get(file.id)
        if (urlInfo && !urlInfo.error) {
          file.presignedUrl = urlInfo.url
          // Use encrypted key from batch response if available
          if (urlInfo.encrypted_key) {
            file.encryptedKey = urlInfo.encrypted_key
          }
        } else {
          file.status = 'error'
          file.error = urlInfo?.error || 'URL not available'
        }
      }
      
      // Start download and ZIP assembly
      const zipFileName = `${folderName || tree.root_folder}.zip`
      await this.startZipDownload(zipFileName)
      
    } catch (error) {
      this.handleError(error)
    }
  }

  /**
   * Download multiple selected items as ZIP
   */
  async downloadSelection(fileIds, folderIds, zipName = 'selection.zip') {
    // Reset all state before starting new download
    this.reset()
    
    this.setStatus(DownloadStatus.INITIALIZING)
    this.sessionId = `selection-${Date.now()}`
    this.startTime = Date.now()
    this.aborted = false
    
    try {
      // Get selection tree
      this.setStatus(DownloadStatus.FETCHING_TREE)
      const treeResponse = await api.post('/files/selection-tree', {
        file_ids: fileIds || [],
        folder_ids: folderIds || []
      })
      const tree = treeResponse.data
      
      this.totalFiles = tree.total_files
      this.totalSize = tree.total_size
      this.files = tree.files.map(f => new FileDownloadTask({
        id: f.id,
        name: f.name,
        relative_path: f.relative_path,
        size: f.size,
        mime_type: f.mime_type,
        encrypted_key: f.encrypted_key
      }))
      
      if (this.files.length === 0) {
        throw new Error('Aucun fichier à télécharger')
      }
      
      // Get crypto key
      const authStore = useAuthStore()
      this.cryptoKey = authStore.masterKey
      if (!this.cryptoKey) {
        throw new Error('Clé de déchiffrement non disponible')
      }
      
      // Generate batch presigned URLs
      this.setStatus(DownloadStatus.GENERATING_URLS)
      const batchFileIds = this.files.map(f => f.id)
      const presignResponse = await api.post('/files/batch-presign', { file_ids: batchFileIds })
      
      // Map URLs to files
      const urlMap = new Map(presignResponse.data.urls.map(u => [u.file_id, u]))
      for (const file of this.files) {
        const urlInfo = urlMap.get(file.id)
        if (urlInfo && !urlInfo.error) {
          file.presignedUrl = urlInfo.url
          if (urlInfo.encrypted_key) {
            file.encryptedKey = urlInfo.encrypted_key
          }
        } else {
          file.status = 'error'
          file.error = urlInfo?.error || 'URL not available'
        }
      }
      
      await this.startZipDownload(zipName)
      
    } catch (error) {
      this.handleError(error)
    }
  }

  /**
   * Download a single file with streaming progress (no ZIP)
   * Uses the same UI as multi-file downloads for consistent UX
   */
  async downloadSingleFile(fileId, fileName, encryptedKey = null, fileSize = 0) {
    this.reset()
    this.isSingleFile = true
    
    this.setStatus(DownloadStatus.INITIALIZING)
    this.sessionId = `single-${fileId}-${Date.now()}`
    this.startTime = Date.now()
    this.aborted = false
    
    try {
      // Get crypto key
      const authStore = useAuthStore()
      this.cryptoKey = authStore.masterKey
      if (!this.cryptoKey) {
        throw new Error('Clé de déchiffrement non disponible')
      }
      
      // Generate presigned URL
      this.setStatus(DownloadStatus.GENERATING_URLS)
      const presignResponse = await api.post('/files/batch-presign', { file_ids: [fileId] })
      const urlInfo = presignResponse.data.urls[0]
      
      if (!urlInfo || urlInfo.error) {
        throw new Error(urlInfo?.error || 'Impossible de générer l\'URL de téléchargement')
      }
      
      // Create file task
      const file = new FileDownloadTask({
        id: fileId,
        name: fileName,
        relative_path: fileName,
        size: fileSize || 0,
        encrypted_key: encryptedKey || urlInfo.encrypted_key
      })
      file.presignedUrl = urlInfo.url
      
      this.files = [file]
      this.totalFiles = 1
      this.totalSize = fileSize || 0
      
      // Start streaming download
      this.setStatus(DownloadStatus.DOWNLOADING)
      await this.downloadAndDecryptSingleFile(file, fileName)
      
    } catch (error) {
      this.handleError(error)
    }
  }

  /**
   * Download, decrypt and save a single file with streaming progress
   */
  async downloadAndDecryptSingleFile(file, fileName) {
    file.status = 'downloading'
    file.abortController = new AbortController()
    
    try {
      // Import the file's encryption key
      const fileKey = await this.importFileKey(file.encryptedKey)
      
      // Fetch with streaming
      const response = await fetch(file.presignedUrl, {
        signal: file.abortController.signal
      })
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`)
      }
      
      // Get size from Content-Length if not known
      const contentLength = parseInt(response.headers.get('content-length') || '0', 10)
      if (contentLength > 0) {
        file.size = contentLength
        this.totalSize = contentLength
      }
      
      const reader = response.body.getReader()
      let buffer = new Uint8Array(0)
      let chunkIndex = 0
      const decryptedChunks = []
      
      file.bytesDownloaded = 0
      
      // Stream and collect encrypted data
      while (!this.aborted) {
        const { done, value } = await reader.read()
        
        if (this.aborted) {
          reader.cancel()
          return
        }
        
        if (value) {
          // Append to buffer
          const newBuffer = new Uint8Array(buffer.length + value.length)
          newBuffer.set(buffer)
          newBuffer.set(value, buffer.length)
          buffer = newBuffer
          
          file.bytesDownloaded += value.length
          this.bytesDownloaded += value.length
          
          // Update progress
          if (file.size > 0) {
            file.progress = Math.min(95, Math.round((file.bytesDownloaded / file.size) * 100))
          }
          
          this.scheduleProgressUpdate()
        }
        
        // Process complete encrypted chunks
        while (buffer.length >= ENCRYPTED_CHUNK_SIZE && !this.aborted) {
          const encryptedChunk = buffer.slice(0, ENCRYPTED_CHUNK_SIZE)
          buffer = buffer.slice(ENCRYPTED_CHUNK_SIZE)
          
          // Decrypt chunk
          const decrypted = await this.decryptChunk(encryptedChunk, fileKey, chunkIndex)
          decryptedChunks.push(decrypted)
          chunkIndex++
        }
        
        if (done) {
          // Process final partial chunk
          if (buffer.length > 0 && !this.aborted) {
            this.setStatus(DownloadStatus.DECRYPTING)
            const decrypted = await this.decryptChunk(buffer, fileKey, chunkIndex)
            decryptedChunks.push(decrypted)
          }
          break
        }
      }
      
      if (this.aborted) return
      
      // Combine all decrypted chunks
      this.setStatus(DownloadStatus.FINALIZING)
      const totalLength = decryptedChunks.reduce((sum, chunk) => sum + chunk.length, 0)
      const completeFile = new Uint8Array(totalLength)
      let offset = 0
      for (const chunk of decryptedChunks) {
        completeFile.set(chunk, offset)
        offset += chunk.length
      }
      
      // Create blob and trigger download
      const blob = new Blob([completeFile])
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = fileName
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      
      // Clean up after short delay
      setTimeout(() => URL.revokeObjectURL(url), 5000)
      
      file.status = 'completed'
      file.progress = 100
      this.processedFiles = 1
      
      this.setStatus(DownloadStatus.COMPLETED)
      this.callbacks.onComplete({
        totalFiles: 1,
        totalSize: totalLength,
        duration: Date.now() - this.startTime
      })
      
    } catch (error) {
      if (error.name === 'AbortError') {
        file.status = 'aborted'
      } else {
        file.status = 'error'
        file.error = error.message
        console.error(`[DownloadManager] Error downloading ${file.name}:`, error)
        this.handleError(error)
      }
    }
  }

  /**
   * Start the ZIP streaming download process
   */
  async startZipDownload(fileName) {
    this.setStatus(DownloadStatus.DOWNLOADING)
    
    // Filter files with valid URLs
    const validFiles = this.files.filter(f => f.presignedUrl && f.status !== 'error')
    this.pendingQueue = [...validFiles]
    
    if (this.worker) {
      // Use Service Worker for streaming ZIP
      await this.initServiceWorkerSession(fileName)
      await this.processDownloadQueue()
      await this.finalizeZip()
    } else {
      // Fallback: Use in-memory ZIP with fflate
      await this.downloadWithFallback(fileName)
    }
  }

  /**
   * Initialize Service Worker session
   */
  async initServiceWorkerSession(fileName) {
    return new Promise((resolve, reject) => {
      this.messageChannel = new MessageChannel()
      
      this.messageChannel.port1.onmessage = (event) => {
        const { type, downloadUrl, error, processedFiles, bytesProcessed, percent } = event.data
        
        switch (type) {
          case 'INIT_SUCCESS':
            // Trigger download by opening the streaming URL
            this.downloadUrl = downloadUrl
            const link = document.createElement('a')
            link.href = downloadUrl
            link.download = fileName
            document.body.appendChild(link)
            link.click()
            document.body.removeChild(link)
            resolve()
            break
          case 'PROGRESS':
            this.processedFiles = processedFiles
            this.bytesWritten = bytesProcessed
            this.reportProgress()
            break
          case 'ERROR':
            reject(new Error(error))
            break
        }
      }
      
      this.worker.postMessage({
        type: 'INIT_DOWNLOAD',
        sessionId: this.sessionId,
        data: {
          fileName,
          totalFiles: this.totalFiles,
          totalSize: this.totalSize
        }
      }, [this.messageChannel.port2])
    })
  }

  /**
   * Process download queue with concurrency control
   */
  async processDownloadQueue() {
    const downloadPromises = []
    
    while ((this.pendingQueue.length > 0 || this.activeDownloads.size > 0) && !this.aborted) {
      // Fill up to max concurrent downloads
      while (this.pendingQueue.length > 0 && this.activeDownloads.size < MAX_CONCURRENT_DOWNLOADS) {
        const file = this.pendingQueue.shift()
        const promise = this.downloadAndStreamFile(file)
        this.activeDownloads.add(file.id)
        
        promise.finally(() => {
          this.activeDownloads.delete(file.id)
        })
        
        downloadPromises.push(promise)
      }
      
      // Wait for any download to complete before continuing
      if (this.activeDownloads.size >= MAX_CONCURRENT_DOWNLOADS) {
        await Promise.race(downloadPromises.filter(Boolean))
      }
      
      // Small delay to prevent tight loop
      await new Promise(resolve => setTimeout(resolve, 10))
    }
    
    // Wait for all remaining downloads
    await Promise.allSettled(downloadPromises)
  }

  /**
   * Download a single file and stream decrypted chunks to ZIP
   */
  async downloadAndStreamFile(file) {
    file.status = 'downloading'
    file.abortController = new AbortController()
    
    try {
      // Import the file's encryption key
      const fileKey = await this.importFileKey(file.encryptedKey)
      
      // Fetch encrypted file stream
      const response = await fetch(file.presignedUrl, {
        signal: file.abortController.signal
      })
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`)
      }
      
      const reader = response.body.getReader()
      let buffer = new Uint8Array(0)
      let chunkIndex = 0
      let isFirst = true
      
      while (true) {
        const { done, value } = await reader.read()
        
        if (this.aborted) {
          reader.cancel()
          break
        }
        
        if (value) {
          // Append to buffer
          const newBuffer = new Uint8Array(buffer.length + value.length)
          newBuffer.set(buffer)
          newBuffer.set(value, buffer.length)
          buffer = newBuffer
          
          // Update byte counters
          file.bytesDownloaded += value.length
          this.bytesDownloaded += value.length
          
          // Update individual file progress (percentage based on encrypted size)
          if (file.size > 0) {
            file.progress = Math.min(99, Math.round((file.bytesDownloaded / file.size) * 100))
          }
          
          // Throttled progress report
          this.scheduleProgressUpdate()
        }
        
        // Process complete encrypted chunks
        while (buffer.length >= ENCRYPTED_CHUNK_SIZE && !this.aborted) {
          const encryptedChunk = buffer.slice(0, ENCRYPTED_CHUNK_SIZE)
          buffer = buffer.slice(ENCRYPTED_CHUNK_SIZE)
          
          // Decrypt chunk
          const decrypted = await this.decryptChunk(encryptedChunk, fileKey, chunkIndex)
          
          // Send to Service Worker
          await this.sendChunkToWorker(file.relativePath, decrypted, false, file.size)
          
          chunkIndex++
        }
        
        if (done) {
          // Process final partial chunk
          if (buffer.length > 0 && !this.aborted) {
            const decrypted = await this.decryptChunk(buffer, fileKey, chunkIndex)
            await this.sendChunkToWorker(file.relativePath, decrypted, true, file.size)
          } else {
            // Mark file as complete even if no final chunk
            await this.sendChunkToWorker(file.relativePath, new Uint8Array(0), true, file.size)
          }
          break
        }
      }
      
      file.status = 'completed'
      file.progress = 100
      
    } catch (error) {
      if (error.name === 'AbortError') {
        file.status = 'aborted'
      } else {
        file.status = 'error'
        file.error = error.message
        console.error(`[DownloadManager] Error downloading ${file.name}:`, error)
      }
    }
  }

  /**
   * Import file encryption key from wrapped format
   */
  async importFileKey(encryptedKeyBase64) {
    // The encrypted key is wrapped with the master key
    // Use sodium.from_base64 for URL-safe base64 compatibility
    await sodium.ready
    const encryptedKeyData = sodium.from_base64(encryptedKeyBase64)
    
    // Extract IV and encrypted key
    const iv = encryptedKeyData.slice(0, 12)
    const wrappedKey = encryptedKeyData.slice(12)
    
    // Decrypt the file key using master key
    const rawKey = await crypto.subtle.decrypt(
      { name: 'AES-GCM', iv },
      this.cryptoKey,
      wrappedKey
    )
    
    // Import as AES-GCM key
    return crypto.subtle.importKey(
      'raw',
      rawKey,
      { name: 'AES-GCM' },
      false,
      ['decrypt']
    )
  }

  /**
   * Decrypt a single chunk
   */
  async decryptChunk(encryptedData, cryptoKey, chunkIndex) {
    const iv = encryptedData.slice(0, IV_LENGTH)
    const ciphertext = encryptedData.slice(IV_LENGTH)
    
    const decrypted = await crypto.subtle.decrypt(
      {
        name: 'AES-GCM',
        iv,
        tagLength: TAG_LENGTH * 8
      },
      cryptoKey,
      ciphertext
    )
    
    return new Uint8Array(decrypted)
  }

  /**
   * Send decrypted chunk to Service Worker
   */
  async sendChunkToWorker(relativePath, chunk, isLast, fileSize) {
    return new Promise((resolve, reject) => {
      if (!this.worker) {
        reject(new Error('Service Worker not available'))
        return
      }
      
      const channel = new MessageChannel()
      
      channel.port1.onmessage = (event) => {
        if (event.data.type === 'FILE_ADDED' || event.data.type === 'PROGRESS') {
          resolve()
        } else if (event.data.type === 'ERROR') {
          reject(new Error(event.data.error))
        }
      }
      
      this.worker.postMessage({
        type: 'ADD_FILE',
        sessionId: this.sessionId,
        data: {
          relativePath,
          chunk: chunk.buffer,
          isLast,
          fileSize
        }
      }, [channel.port2, chunk.buffer])
    })
  }

  /**
   * Finalize ZIP stream
   */
  async finalizeZip() {
    this.setStatus(DownloadStatus.FINALIZING)
    
    return new Promise((resolve, reject) => {
      if (!this.worker) {
        resolve()
        return
      }
      
      const channel = new MessageChannel()
      
      channel.port1.onmessage = (event) => {
        if (event.data.type === 'FINALIZE_SUCCESS') {
          this.setStatus(DownloadStatus.COMPLETED)
          this.callbacks.onComplete({
            totalFiles: this.totalFiles,
            totalSize: this.totalSize,
            duration: Date.now() - this.startTime
          })
          resolve()
        } else if (event.data.type === 'ERROR') {
          reject(new Error(event.data.error))
        }
      }
      
      this.worker.postMessage({
        type: 'FINALIZE',
        sessionId: this.sessionId
      }, [channel.port2])
    })
  }

  /**
   * Fallback download using in-memory ZIP (for browsers without Service Worker)
   * Uses streaming progress tracking for smooth UI updates
   */
  async downloadWithFallback(fileName) {
    // Dynamic import fflate
    const fflate = await import('fflate')
    
    const zipData = {}
    
    for (const file of this.files.filter(f => f.presignedUrl)) {
      if (this.aborted) break
      
      try {
        file.status = 'downloading'
        file.bytesDownloaded = 0
        
        // Import file key
        const fileKey = await this.importFileKey(file.encryptedKey)
        
        // Download with streaming progress tracking
        const response = await fetch(file.presignedUrl)
        
        // Get content length for progress calculation
        const contentLength = parseInt(response.headers.get('content-length') || '0', 10)
        if (contentLength > 0 && file.size === 0) {
          file.size = contentLength
        }
        
        // Stream the response with progress tracking
        const reader = response.body.getReader()
        const chunks = []
        
        while (true) {
          const { done, value } = await reader.read()
          if (done) break
          
          chunks.push(value)
          file.bytesDownloaded += value.length
          this.bytesDownloaded += value.length
          
          // Update individual file progress
          if (file.size > 0) {
            file.progress = Math.min(99, Math.round((file.bytesDownloaded / file.size) * 100))
          }
          
          this.scheduleProgressUpdate()
        }
        
        // Combine chunks into single array
        const encryptedLength = chunks.reduce((sum, chunk) => sum + chunk.length, 0)
        const encryptedData = new Uint8Array(encryptedLength)
        let encPos = 0
        for (const chunk of chunks) {
          encryptedData.set(chunk, encPos)
          encPos += chunk.length
        }
        
        // Decrypt all chunks
        const decryptedParts = []
        let offset = 0
        let chunkIndex = 0
        
        while (offset < encryptedData.length) {
          const chunkSize = Math.min(ENCRYPTED_CHUNK_SIZE, encryptedData.length - offset)
          const chunk = encryptedData.slice(offset, offset + chunkSize)
          const decrypted = await this.decryptChunk(chunk, fileKey, chunkIndex)
          decryptedParts.push(decrypted)
          offset += chunkSize
          chunkIndex++
        }
        
        // Combine decrypted parts
        const decryptedLength = decryptedParts.reduce((sum, part) => sum + part.length, 0)
        const combined = new Uint8Array(decryptedLength)
        let decPos = 0
        for (const part of decryptedParts) {
          combined.set(part, decPos)
          decPos += part.length
        }
        
        // Add to ZIP structure
        zipData[file.relativePath] = combined
        
        file.status = 'completed'
        file.progress = 100
        this.processedFiles++
        this.reportProgress()
        
      } catch (error) {
        file.status = 'error'
        file.error = error.message
        console.error(`[DownloadManager] Fallback error for ${file.name}:`, error)
      }
    }
    
    // Create ZIP
    this.setStatus(DownloadStatus.FINALIZING)
    
    const zipped = fflate.zipSync(zipData, { level: 0 })
    
    // Trigger download
    const blob = new Blob([zipped], { type: 'application/zip' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = fileName
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    
    setTimeout(() => URL.revokeObjectURL(url), 10000)
    
    this.setStatus(DownloadStatus.COMPLETED)
    this.callbacks.onComplete({
      totalFiles: this.processedFiles,
      totalSize: this.totalSize,
      duration: Date.now() - this.startTime
    })
  }

  /**
   * Abort the download
   */
  abort() {
    this.aborted = true
    this.setStatus(DownloadStatus.ABORTED)
    
    // Abort all active downloads
    for (const file of this.files) {
      if (file.abortController) {
        file.abortController.abort()
      }
    }
    
    // Notify Service Worker
    if (this.worker) {
      const channel = new MessageChannel()
      this.worker.postMessage({
        type: 'ABORT',
        sessionId: this.sessionId
      }, [channel.port2])
    }
  }

  /**
   * Schedule a throttled progress update using requestAnimationFrame
   * This prevents overwhelming the UI with thousands of updates per second
   */
  scheduleProgressUpdate() {
    if (this._progressScheduled) return
    
    const now = performance.now()
    const timeSinceLastUpdate = now - this._lastProgressTime
    
    if (timeSinceLastUpdate >= PROGRESS_THROTTLE_MS) {
      // Enough time has passed, update immediately
      this._lastProgressTime = now
      this.reportProgress()
    } else {
      // Schedule update for next animation frame
      this._progressScheduled = true
      this._rafId = requestAnimationFrame(() => {
        this._progressScheduled = false
        this._lastProgressTime = performance.now()
        this.reportProgress()
      })
    }
  }

  /**
   * Report progress to callbacks
   * Calculates global progress based on total bytes downloaded vs total size
   */
  reportProgress() {
    const elapsed = Date.now() - this.startTime
    const elapsedSeconds = elapsed / 1000
    
    // Calculate speed with smoothing (exponential moving average)
    const instantSpeed = elapsedSeconds > 0 ? this.bytesDownloaded / elapsedSeconds : 0
    
    // Calculate remaining time
    const remaining = this.totalSize - this.bytesDownloaded
    const eta = instantSpeed > 0 ? Math.ceil(remaining / instantSpeed) : 0
    
    // Global progress: bytes-based, not file-based
    const globalPercent = this.totalSize > 0 
      ? Math.min(100, Math.round((this.bytesDownloaded / this.totalSize) * 100))
      : 0
    
    const progress = {
      status: this.status,
      totalFiles: this.totalFiles,
      processedFiles: this.processedFiles,
      totalSize: this.totalSize,
      bytesDownloaded: this.bytesDownloaded,
      percent: globalPercent,
      speed: instantSpeed,
      eta,
      isSingleFile: this.isSingleFile,
      files: this.files.map(f => ({
        name: f.name,
        status: f.status,
        progress: f.progress,
        bytesDownloaded: f.bytesDownloaded,
        size: f.size,
        error: f.error
      }))
    }
    
    this.callbacks.onProgress(progress)
  }

  /**
   * Set status and notify
   */
  setStatus(status) {
    this.status = status
    this.callbacks.onStatusChange(status)
    this.reportProgress()
  }

  /**
   * Handle errors
   */
  handleError(error) {
    console.error('[DownloadManager] Error:', error)
    this.setStatus(DownloadStatus.ERROR)
    this.callbacks.onError(error.message || error)
  }
}

// Export singleton instance
export const downloadManager = new ZipDownloadManager()

// Export class for testing
export { ZipDownloadManager }

/**
 * Utility function: Track stream progress with throttling
 * Wraps a ReadableStream to report progress without blocking the data flow
 * 
 * @param {ReadableStream} stream - The input stream to track
 * @param {Function} onProgress - Callback(bytesRead) called on each chunk
 * @param {number} throttleMs - Minimum ms between progress callbacks (default: 16ms ~60fps)
 * @returns {ReadableStream} - A new stream that passes through all data while tracking progress
 * 
 * @example
 * const response = await fetch(url)
 * const trackedStream = trackStreamProgress(
 *   response.body,
 *   (bytes) => //console.log(`Downloaded: ${bytes}`),
 *   16
 * )
 * const reader = trackedStream.getReader()
 */
export function trackStreamProgress(stream, onProgress, throttleMs = PROGRESS_THROTTLE_MS) {
  let totalBytesRead = 0
  let lastReportTime = 0
  let pendingReport = false
  let rafId = null
  
  const reader = stream.getReader()
  
  return new ReadableStream({
    async pull(controller) {
      try {
        const { done, value } = await reader.read()
        
        if (done) {
          // Final progress report
          if (rafId) cancelAnimationFrame(rafId)
          onProgress(totalBytesRead)
          controller.close()
          return
        }
        
        totalBytesRead += value.length
        controller.enqueue(value)
        
        // Throttled progress reporting
        const now = performance.now()
        if (now - lastReportTime >= throttleMs) {
          lastReportTime = now
          onProgress(totalBytesRead)
        } else if (!pendingReport) {
          pendingReport = true
          rafId = requestAnimationFrame(() => {
            pendingReport = false
            lastReportTime = performance.now()
            onProgress(totalBytesRead)
          })
        }
      } catch (error) {
        if (rafId) cancelAnimationFrame(rafId)
        controller.error(error)
      }
    },
    
    cancel(reason) {
      if (rafId) cancelAnimationFrame(rafId)
      reader.cancel(reason)
    }
  })
}
