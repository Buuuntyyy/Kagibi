/**
 * Streaming Decryption Download Manager
 * Zero-Knowledge streaming download with minimal memory footprint
 * 
 * Architecture:
 * S3 (presigned URL) -> fetch ReadableStream -> TransformStream (decrypt) -> StreamSaver/FileSystem
 */

import { unwrapMasterKey, CHUNK_SIZE } from './crypto'

// Constants matching backend encryption format
const IV_LENGTH = 12
const TAG_LENGTH = 16
const ENCRYPTED_CHUNK_OVERHEAD = IV_LENGTH + TAG_LENGTH
const ENCRYPTED_CHUNK_SIZE = CHUNK_SIZE + ENCRYPTED_CHUNK_OVERHEAD

/**
 * Download state enum
 */
export const DownloadState = {
  IDLE: 'idle',
  FETCHING_URL: 'fetching_url',
  DOWNLOADING: 'downloading',
  DECRYPTING: 'decrypting',
  SAVING: 'saving',
  COMPLETED: 'completed',
  ERROR: 'error',
  ABORTED: 'aborted'
}

/**
 * Creates a TransformStream that decrypts AES-GCM chunks on the fly
 * Handles backpressure automatically via Web Streams API
 */
function createDecryptionTransform(cryptoKey, onProgress) {
  let buffer = new Uint8Array(0)
  let chunkIndex = 0
  let totalDecrypted = 0

  return new TransformStream({
    async transform(chunk, controller) {
      // Append incoming data to buffer
      const newBuffer = new Uint8Array(buffer.length + chunk.length)
      newBuffer.set(buffer)
      newBuffer.set(chunk, buffer.length)
      buffer = newBuffer

      // Process complete encrypted chunks
      while (buffer.length >= ENCRYPTED_CHUNK_SIZE) {
        const encryptedChunk = buffer.slice(0, ENCRYPTED_CHUNK_SIZE)
        buffer = buffer.slice(ENCRYPTED_CHUNK_SIZE)

        try {
          const decrypted = await decryptChunk(encryptedChunk, cryptoKey, chunkIndex)
          controller.enqueue(decrypted)
          totalDecrypted += decrypted.byteLength
          chunkIndex++
          onProgress?.(totalDecrypted)
        } catch (error) {
          controller.error(new Error(`Decryption failed at chunk ${chunkIndex}: ${error.message}`))
          return
        }
      }
    },

    async flush(controller) {
      // Handle final partial chunk
      if (buffer.length > 0) {
        if (buffer.length < IV_LENGTH + TAG_LENGTH) {
          controller.error(new Error('Corrupted final chunk: insufficient data'))
          return
        }

        try {
          const decrypted = await decryptChunk(buffer, cryptoKey, chunkIndex)
          controller.enqueue(decrypted)
          totalDecrypted += decrypted.byteLength
          onProgress?.(totalDecrypted)
        } catch (error) {
          controller.error(new Error(`Final chunk decryption failed: ${error.message}`))
        }
      }
    }
  })
}

/**
 * Decrypt a single chunk using Web Crypto API
 * Format: [IV (12 bytes)][Ciphertext + Tag]
 */
async function decryptChunk(encryptedData, cryptoKey, chunkIndex) {
  const iv = encryptedData.slice(0, IV_LENGTH)
  const ciphertext = encryptedData.slice(IV_LENGTH)

  // IV is stored directly without chunk index modification
  // (matches the encryption in crypto.worker.js)
  const decrypted = await crypto.subtle.decrypt(
    {
      name: 'AES-GCM',
      iv: iv,
      tagLength: TAG_LENGTH * 8
    },
    cryptoKey,
    ciphertext
  )

  return new Uint8Array(decrypted)
}

/**
 * StreamSaver-like functionality using Service Worker or native File System Access API
 */
class StreamSaver {
  constructor(fileName, options = {}) {
    this.fileName = fileName
    this.mimeType = options.mimeType || 'application/octet-stream'
    this.fileHandle = null
    this.writer = null
    this.useNativeFS = 'showSaveFilePicker' in window && !options.forceBlob
  }

  async init() {
    if (this.useNativeFS) {
      try {
        this.fileHandle = await window.showSaveFilePicker({
          suggestedName: this.fileName,
          types: [{
            description: 'File',
            accept: { [this.mimeType]: [] }
          }]
        })
        const writable = await this.fileHandle.createWritable()
        this.writer = writable
        return true
      } catch (error) {
        if (error.name === 'AbortError') {
          throw error // User cancelled
        }
        // Fallback to blob method
        this.useNativeFS = false
      }
    }

    // Fallback: collect in memory then download
    this.chunks = []
    return true
  }

  async write(chunk) {
    if (this.useNativeFS && this.writer) {
      await this.writer.write(chunk)
    } else {
      this.chunks.push(chunk)
    }
  }

  async close() {
    if (this.useNativeFS && this.writer) {
      await this.writer.close()
    } else if (this.chunks) {
      // Create blob and trigger download
      const blob = new Blob(this.chunks, { type: this.mimeType })
      this.chunks = null // Release memory

      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = this.fileName
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      
      // Delay revoke to ensure download starts
      setTimeout(() => URL.revokeObjectURL(url), 10000)
    }
  }

  abort() {
    if (this.useNativeFS && this.writer) {
      this.writer.abort()
    }
    this.chunks = null
  }
}

/**
 * Main streaming download function
 * Downloads and decrypts a file with minimal memory usage
 */
export async function downloadAndDecryptStream(fileId, masterKey, options = {}) {
  const {
    onProgress = () => {},
    onStateChange = () => {},
    onError = () => {},
    api = null, // Axios or fetch wrapper
    forceBlob = false
  } = options

  let state = DownloadState.IDLE
  let abortController = new AbortController()
  let streamSaver = null

  const setState = (newState) => {
    state = newState
    onStateChange(newState)
  }

  try {
    // 1. Fetch presigned URL and metadata from backend
    setState(DownloadState.FETCHING_URL)
    
    const metaResponse = await api.get(`/files/download/${fileId}/presigned`)
    const {
      url: presignedUrl,
      file_name: fileName,
      file_size: fileSize,
      mime_type: mimeType,
      encrypted_key: encryptedKey
    } = metaResponse.data

    // 2. Decrypt the file key using master key
    let fileKey = masterKey
    if (encryptedKey) {
      fileKey = await unwrapMasterKey(encryptedKey, masterKey)
    }

    // 3. Initialize stream saver (prompts user for save location)
    setState(DownloadState.SAVING)
    streamSaver = new StreamSaver(fileName, { mimeType, forceBlob })
    await streamSaver.init()

    // 4. Fetch encrypted file as stream
    setState(DownloadState.DOWNLOADING)
    
    const response = await fetch(presignedUrl, {
      signal: abortController.signal,
      cache: 'no-store'
    })

    if (!response.ok) {
      throw new Error(`Download failed: ${response.status} ${response.statusText}`)
    }

    const reader = response.body

    if (!reader) {
      throw new Error('Streaming not supported by browser')
    }

    // 5. Create decryption transform stream
    let downloadedBytes = 0
    let decryptedBytes = 0

    const decryptTransform = createDecryptionTransform(fileKey, (bytes) => {
      decryptedBytes = bytes
      onProgress({
        downloadedBytes,
        decryptedBytes,
        totalBytes: fileSize,
        percentage: Math.round((downloadedBytes / fileSize) * 100)
      })
    })

    // 6. Create write stream
    const writeStream = new WritableStream({
      async write(chunk) {
        await streamSaver.write(chunk)
      },
      async close() {
        await streamSaver.close()
      },
      abort(reason) {
        streamSaver.abort()
      }
    })

    // 7. Track download progress
    const progressTransform = new TransformStream({
      transform(chunk, controller) {
        downloadedBytes += chunk.byteLength
        controller.enqueue(chunk)
        onProgress({
          downloadedBytes,
          decryptedBytes,
          totalBytes: fileSize,
          percentage: Math.round((downloadedBytes / fileSize) * 100)
        })
      }
    })

    // 8. Pipe: S3 -> progress tracker -> decryptor -> file writer
    setState(DownloadState.DECRYPTING)

    await reader
      .pipeThrough(progressTransform)
      .pipeThrough(decryptTransform)
      .pipeTo(writeStream)

    setState(DownloadState.COMPLETED)

    return {
      success: true,
      fileName,
      fileSize,
      decryptedBytes
    }

  } catch (error) {
    if (error.name === 'AbortError') {
      setState(DownloadState.ABORTED)
      streamSaver?.abort()
      return { success: false, aborted: true }
    }

    setState(DownloadState.ERROR)
    onError(error)
    streamSaver?.abort()
    throw error
  }

  // Return abort function
  return {
    abort: () => {
      abortController.abort()
      streamSaver?.abort()
    }
  }
}

/**
 * Simpler fallback for browsers without full streaming support
 * Downloads entire file, decrypts in chunks, saves
 */
export async function downloadAndDecryptFallback(fileId, masterKey, options = {}) {
  const {
    onProgress = () => {},
    onStateChange = () => {},
    onError = () => {},
    api = null
  } = options

  try {
    onStateChange(DownloadState.FETCHING_URL)

    // Get presigned URL
    const metaResponse = await api.get(`/files/download/${fileId}/presigned`)
    const { url: presignedUrl, file_name: fileName, mime_type: mimeType, encrypted_key: encryptedKey } = metaResponse.data

    // Decrypt file key
    let fileKey = masterKey
    if (encryptedKey) {
      fileKey = await unwrapMasterKey(encryptedKey, masterKey)
    }

    onStateChange(DownloadState.DOWNLOADING)

    // Download encrypted blob
    const response = await fetch(presignedUrl)
    const encryptedBlob = await response.blob()

    onStateChange(DownloadState.DECRYPTING)

    // Decrypt in chunks using existing worker-based approach
    const { decryptChunkedFileWorker } = await import('./crypto')
    const decryptedBlob = await decryptChunkedFileWorker(encryptedBlob, fileKey, mimeType)

    onStateChange(DownloadState.SAVING)

    // Trigger download
    const url = URL.createObjectURL(decryptedBlob)
    const a = document.createElement('a')
    a.href = url
    a.download = fileName
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)

    onStateChange(DownloadState.COMPLETED)

    return { success: true, fileName }

  } catch (error) {
    onStateChange(DownloadState.ERROR)
    onError(error)
    throw error
  }
}

/**
 * Auto-select best download method based on browser capabilities
 */
export async function smartDownload(fileId, masterKey, options = {}) {
  // Check for streaming support
  const supportsStreaming = 
    typeof ReadableStream !== 'undefined' &&
    typeof TransformStream !== 'undefined' &&
    typeof WritableStream !== 'undefined' &&
    'pipeThrough' in ReadableStream.prototype

  if (supportsStreaming) {
    return downloadAndDecryptStream(fileId, masterKey, options)
  } else {
    console.warn('Browser does not support streaming, using fallback method')
    return downloadAndDecryptFallback(fileId, masterKey, options)
  }
}
