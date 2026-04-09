/**
 * Download Service Worker
 * Intercepts download requests and generates streaming ZIP files
 * 
 * This worker enables:
 * - Streaming ZIP creation without memory buildup
 * - Backpressure handling for disk write speed
 * - Real-time progress reporting via MessageChannel
 */

// Import fflate for streaming ZIP compression
importScripts('https://cdn.jsdelivr.net/npm/fflate@0.8.2/umd/index.min.js')

const DOWNLOAD_SCOPE = '/download-stream/'

// Active download sessions
const downloadSessions = new Map()

// Install and activate immediately
self.addEventListener('install', (event) => {
  //console.log('[DownloadWorker] Installing...')
  self.skipWaiting()
})

self.addEventListener('activate', (event) => {
  //console.log('[DownloadWorker] Activating...')
  event.waitUntil(self.clients.claim())
})

// Handle messages from main thread
self.addEventListener('message', async (event) => {
  // Verify message origin — only accept messages from the same origin
  if (event.origin && event.origin !== self.location.origin) {
    console.error('[DownloadWorker] Rejected message from untrusted origin:', event.origin)
    return
  }

  const { type, sessionId, data } = event.data
  const port = event.ports[0]

  switch (type) {
    case 'INIT_DOWNLOAD':
      handleInitDownload(sessionId, data, port)
      break
    case 'ADD_FILE':
      handleAddFile(sessionId, data, port)
      break
    case 'FINALIZE':
      handleFinalize(sessionId, port)
      break
    case 'ABORT':
      handleAbort(sessionId, port)
      break
    case 'PING':
      port?.postMessage({ type: 'PONG' })
      break
  }
})

// Intercept fetch requests for download streams
self.addEventListener('fetch', (event) => {
  const url = new URL(event.request.url)
  
  if (url.pathname.startsWith(DOWNLOAD_SCOPE)) {
    const sessionId = url.pathname.replace(DOWNLOAD_SCOPE, '')
    const session = downloadSessions.get(sessionId)
    
    if (session) {
      event.respondWith(createStreamResponse(session))
    } else {
      event.respondWith(new Response('Download session not found', { status: 404 }))
    }
  }
})

/**
 * Initialize a new download session
 */
function handleInitDownload(sessionId, data, port) {
  const { fileName, totalFiles, totalSize } = data
  
  // Create a TransformStream for the ZIP output
  const { readable, writable } = new TransformStream()
  
  // Create ZIP stream using fflate
  const zipStream = new fflate.Zip((err, chunk, final) => {
    if (err) {
      console.error('[DownloadWorker] ZIP error:', err)
      return
    }
    
    const session = downloadSessions.get(sessionId)
    if (!session) return
    
    if (chunk) {
      session.controller?.enqueue(chunk)
      session.bytesWritten += chunk.length
    }
    
    if (final) {
      session.controller?.close()
      downloadSessions.delete(sessionId)
    }
  })
  
  const session = {
    sessionId,
    fileName,
    totalFiles,
    totalSize,
    processedFiles: 0,
    bytesProcessed: 0,
    bytesWritten: 0,
    readable,
    writable,
    zipStream,
    controller: null,
    port,
    aborted: false,
    pendingFiles: [],
    isProcessing: false
  }
  
  downloadSessions.set(sessionId, session)
  
  port?.postMessage({ 
    type: 'INIT_SUCCESS',
    downloadUrl: `${self.registration.scope}download-stream/${sessionId}`
  })
  
  //console.log(`[DownloadWorker] Session ${sessionId} initialized for ${fileName}`)
}

/**
 * Add a decrypted file chunk to the ZIP
 */
function handleAddFile(sessionId, data, port) {
  const session = downloadSessions.get(sessionId)
  if (!session) {
    port?.postMessage({ type: 'ERROR', error: 'Session not found' })
    return
  }
  
  if (session.aborted) {
    port?.postMessage({ type: 'ERROR', error: 'Session aborted' })
    return
  }
  
  const { relativePath, chunk, isLast, fileSize } = data
  
  // Queue the file operation
  session.pendingFiles.push({ relativePath, chunk, isLast, fileSize, port })
  
  // Process queue if not already processing
  if (!session.isProcessing) {
    processFileQueue(session)
  }
}

/**
 * Process queued file chunks sequentially
 */
async function processFileQueue(session) {
  if (session.isProcessing || session.aborted) return
  session.isProcessing = true
  
  while (session.pendingFiles.length > 0 && !session.aborted) {
    const { relativePath, chunk, isLast, fileSize, port } = session.pendingFiles.shift()
    
    try {
      // Get or create file in ZIP
      let zipFile = session.currentFile
      
      if (!zipFile || session.currentFilePath !== relativePath) {
        // New file - create ZIP entry
        if (zipFile) {
          // Close previous file
          zipFile.push(new Uint8Array(0), true)
        }
        
        // Create new ZIP entry with streaming
        zipFile = new fflate.ZipDeflate(relativePath, {
          level: 0 // No compression for encrypted data (already incompressible)
        })
        
        session.zipStream.add(zipFile)
        session.currentFile = zipFile
        session.currentFilePath = relativePath
      }
      
      // Add chunk to ZIP
      if (chunk && chunk.byteLength > 0) {
        const uint8Chunk = chunk instanceof Uint8Array ? chunk : new Uint8Array(chunk)
        zipFile.push(uint8Chunk, isLast)
        session.bytesProcessed += uint8Chunk.length
      } else if (isLast) {
        zipFile.push(new Uint8Array(0), true)
      }
      
      if (isLast) {
        session.processedFiles++
        session.currentFile = null
        session.currentFilePath = null
        
        // Report progress
        const progress = {
          type: 'PROGRESS',
          processedFiles: session.processedFiles,
          totalFiles: session.totalFiles,
          bytesProcessed: session.bytesProcessed,
          totalSize: session.totalSize,
          percent: Math.round((session.bytesProcessed / session.totalSize) * 100)
        }
        
        session.port?.postMessage(progress)
        port?.postMessage({ type: 'FILE_ADDED', relativePath })
      }
      
    } catch (error) {
      console.error(`[DownloadWorker] Error adding file ${relativePath}:`, error)
      port?.postMessage({ type: 'ERROR', error: error.message })
    }
  }
  
  session.isProcessing = false
}

/**
 * Finalize the ZIP and close the stream
 */
function handleFinalize(sessionId, port) {
  const session = downloadSessions.get(sessionId)
  if (!session) {
    port?.postMessage({ type: 'ERROR', error: 'Session not found' })
    return
  }
  
  try {
    // Close current file if any
    if (session.currentFile) {
      session.currentFile.push(new Uint8Array(0), true)
    }
    
    // End the ZIP stream
    session.zipStream.end()
    
    port?.postMessage({ 
      type: 'FINALIZE_SUCCESS',
      totalBytes: session.bytesWritten
    })
    
    //console.log(`[DownloadWorker] Session ${sessionId} finalized, ${session.bytesWritten} bytes written`)
    
  } catch (error) {
    console.error(`[DownloadWorker] Finalize error:`, error)
    port?.postMessage({ type: 'ERROR', error: error.message })
  }
}

/**
 * Abort the download session
 */
function handleAbort(sessionId, port) {
  const session = downloadSessions.get(sessionId)
  if (session) {
    session.aborted = true
    session.controller?.error(new Error('Download aborted'))
    downloadSessions.delete(sessionId)
    //console.log(`[DownloadWorker] Session ${sessionId} aborted`)
  }
  port?.postMessage({ type: 'ABORT_SUCCESS' })
}

/**
 * Create the streaming response for the download
 */
function createStreamResponse(session) {
  const stream = new ReadableStream({
    start(controller) {
      session.controller = controller
    },
    cancel() {
      session.aborted = true
      downloadSessions.delete(session.sessionId)
    }
  })
  
  return new Response(stream, {
    headers: {
      'Content-Type': 'application/zip',
      'Content-Disposition': `attachment; filename="${session.fileName}"`,
      'Cache-Control': 'no-store'
    }
  })
}

//console.log('[DownloadWorker] Service Worker loaded')
