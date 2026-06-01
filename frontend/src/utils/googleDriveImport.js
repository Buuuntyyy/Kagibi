// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

/**
 * Google Drive Import — zero-knowledge pipeline
 *
 * Data visibility contract:
 *   - Kagibi BACKEND never receives: file content (plaintext), Google OAuth token
 *   - Kagibi FRONTEND (browser RAM only, never persisted): OAuth access token,
 *     file/folder names, file sizes, MIME types, plaintext content (chunk by chunk only)
 *   - Kagibi servers receive: same encrypted blobs as any regular upload
 *   - Google receives: OAuth consent event with Kagibi's client ID; Drive read API calls
 *
 * Architecture:
 *   Google Drive (streaming fetch) → buffer to PART_SIZE → encryptChunkWorker
 *     → MultipartUploadManager → S3 (encrypted blobs only)
 */

import api from '../api'
import {
  generateMasterKey,
  wrapMasterKey,
  generateBaseNonce,
  encryptChunkWorker,
  NONCE_LENGTH,
  TAG_LENGTH_BYTES
} from './crypto'
import { MultipartUploadManager, PART_SIZE, pickChunkSize } from './multipartUpload'
import { useAuthStore } from '../stores/auth'

const DRIVE_API = 'https://www.googleapis.com/drive/v3'

// Number of files processed concurrently (download + encrypt + upload in parallel).
// Each slot uses ≤ MAX_CONCURRENT_WORKERS × PART_SIZE ≈ 30 MB of RAM.
const CONCURRENT_FILES = 3

// Google Workspace documents cannot be downloaded directly — they must be exported.
// Each type maps to a compatible Office format preserving content as faithfully as possible.
const WORKSPACE_EXPORTS = {
  'application/vnd.google-apps.document':
    { mime: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document', ext: '.docx' },
  'application/vnd.google-apps.spreadsheet':
    { mime: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet', ext: '.xlsx' },
  'application/vnd.google-apps.presentation':
    { mime: 'application/vnd.openxmlformats-officedocument.presentationml.presentation', ext: '.pptx' },
  'application/vnd.google-apps.drawing':
    { mime: 'image/png', ext: '.png' },
  'application/vnd.google-apps.form':
    { mime: 'application/pdf', ext: '.pdf' },
}

// Types that have no exportable format and are silently skipped
const UNSUPPORTED_TYPES = new Set([
  'application/vnd.google-apps.folder',
  'application/vnd.google-apps.shortcut',
  'application/vnd.google-apps.map',
  'application/vnd.google-apps.site',
  'application/vnd.google-apps.jam',
  'application/vnd.google-apps.script',
])

// Kagibi backend folder name validation: /^[\p{L}\p{N}\s\-\._‘’’]+$/u
// Characters outside this set are replaced by underscore.
// "." and ".." are reserved path names rejected by the backend even after sanitization.
function sanitizeName(name) {
  const s = name.replace(/[^\p{L}\p{N}\s\-._\u2018\u2019']/gu, "_").trim() || "Import"
  return (s === "." || s === "..") ? "Import" : s
}

function _isDescendantOf(path, parentSet) {
  for (const p of parentSet) {
    if (path.startsWith(p + "/")) return true
  }
  return false
}

function _getAncestorValue(path, map) {
  for (const [k, v] of map) {
    if (path.startsWith(k + "/")) return v
  }
  return ""
}

// Returns total encrypted byte count for a given plaintext size and chunk size.
// AES-256-GCM adds (NONCE_LENGTH + TAG_LENGTH_BYTES) = 28 bytes per chunk.
// Empty files produce exactly one chunk of 28 bytes (nonce + auth tag, no ciphertext).
function calcEncryptedSize(plainSize, chunkSize = PART_SIZE) {
  if (plainSize === 0) return NONCE_LENGTH + TAG_LENGTH_BYTES
  const numChunks = Math.ceil(plainSize / chunkSize)
  return plainSize + numChunks * (NONCE_LENGTH + TAG_LENGTH_BYTES)
}

// Load Google Identity Services script once.
function loadGISScript() {
  if (window.google?.accounts?.oauth2) return Promise.resolve()
  return new Promise((resolve, reject) => {
    if (document.querySelector('script[src*="accounts.google.com/gsi/client"]')) {
      const poll = () => window.google?.accounts?.oauth2 ? resolve() : setTimeout(poll, 60)
      poll()
      return
    }
    const s = document.createElement('script')
    s.src = 'https://accounts.google.com/gsi/client'
    s.async = true
    s.onload = resolve
    s.onerror = () => reject(new Error('Impossible de charger Google Identity Services'))
    document.head.appendChild(s)
  })
}

export class GoogleDriveImport {
  constructor() {
    this._clientId = null
    this._accessToken = null
    this._aborted = false
  }

  // Fetch client ID from backend and load GIS. Must be called once before authenticate().
  async init() {
    const res = await api.get('/import/google/config')
    this._clientId = res.data.client_id
    await loadGISScript()
  }

  get isConfigured() {
    return !!this._clientId
  }

  // Opens the Google OAuth consent popup and stores the short-lived access token
  // in memory only. The token is never sent to Kagibi's backend.
  authenticate() {
    return new Promise((resolve, reject) => {
      const client = window.google.accounts.oauth2.initTokenClient({
        client_id: this._clientId,
        scope: 'https://www.googleapis.com/auth/drive.readonly',
        callback: (response) => {
          if (response.error) {
            reject(new Error(response.error_description || response.error))
          } else {
            this._accessToken = response.access_token
            resolve()
          }
        },
        error_callback: (err) => {
          reject(new Error(err?.message || 'Authentification Google annulée'))
        }
      })
      client.requestAccessToken()
    })
  }

  // Fetches folders and files in parallel, strips the Google Photos/Computers subtrees
  // client-side ('in ancestors' is not a valid Drive API v3 query operator).
  // orderBy is intentionally omitted: the Drive API sorts server-side before paginating,
  // which dramatically slows down large libraries. Sorting is done locally instead.
  async listAllItems(onProgress) {
    const [allFolders, allFiles] = await Promise.all([
      this._fetchPage(
        "mimeType='application/vnd.google-apps.folder' and trashed=false and 'me' in owners",
        'id,name,parents',
        onProgress ? (n) => onProgress('folders', n) : null
      ),
      this._fetchPage(
        "mimeType!='application/vnd.google-apps.folder' and trashed=false and 'me' in owners",
        'id,name,mimeType,size,parents',
        onProgress ? (n) => onProgress('files', n) : null
      )
    ])

    const excludedIds = this._buildSystemExcludeSet(allFolders)
    const folders = excludedIds.size > 0
      ? allFolders.filter(f => !excludedIds.has(f.id))
      : allFolders
    const files = excludedIds.size > 0
      ? allFiles.filter(f => !excludedIds.has((f.parents ?? [])[0]))
      : allFiles

    files.sort((a, b) => a.name.localeCompare(b.name))
    return { folders, files }
  }

  // Finds Drive-managed system folders that are NOT visible in the "My Drive" tab
  // ("Google Photos" for photo backups, "Computers" for desktop sync) and returns
  // a Set of their IDs plus all descendant folder IDs to exclude from the import.
  // These folders are root-level (parent absent from our folder list) with known names.
  _buildSystemExcludeSet(folders) {
    const SYSTEM_NAMES = new Set(['Google Photos', 'Computers'])
    const folderIds = new Set(folders.map(f => f.id))

    const systemRoots = folders.filter(
      f => SYSTEM_NAMES.has(f.name) && !folderIds.has((f.parents ?? [])[0])
    )
    if (systemRoots.length === 0) return new Set()

    const childrenOf = new Map()
    for (const folder of folders) {
      const pid = (folder.parents ?? [])[0]
      if (pid) {
        if (!childrenOf.has(pid)) childrenOf.set(pid, [])
        childrenOf.get(pid).push(folder.id)
      }
    }

    const excluded = new Set(systemRoots.map(f => f.id))
    const queue = [...excluded]
    let head = 0
    while (head < queue.length) {
      const pid = queue[head++]
      for (const childId of (childrenOf.get(pid) ?? [])) {
        if (!excluded.has(childId)) {
          excluded.add(childId)
          queue.push(childId)
        }
      }
    }
    return excluded
  }

  async _fetchPage(q, fileFields, onCount) {
    const items = []
    let pageToken = null
    do {
      const params = new URLSearchParams({
        fields: `nextPageToken,files(${fileFields})`,
        pageSize: '1000',
        q,
        spaces: 'drive'   // restrict to "My Drive" only — excludes Photos space, appDataFolder
      })
      if (pageToken) params.set('pageToken', pageToken)

      const res = await fetch(`${DRIVE_API}/files?${params}`, {
        headers: { Authorization: `Bearer ${this._accessToken}` }
      })
      if (!res.ok) {
        const body = await res.json().catch(() => ({}))
        throw new Error(body?.error?.message || `Google Drive API: ${res.status}`)
      }
      const data = await res.json()
      items.push(...(data.files ?? []))
      if (onCount) onCount(items.length)
      pageToken = data.nextPageToken
    } while (pageToken)
    return items
  }

  // Builds a Map<folderId, sanitizedKagibiPath> from folders only (much fewer than files).
  // File paths are resolved later as: folderPath + '/' + fileName.
  // Uses iterative BFS to avoid call-stack overflow on deep hierarchies.
  async buildPathMap(folders) {
    const byId = new Map(folders.map(f => [f.id, f]))
    const pathCache = new Map()

    const children = new Map()
    for (const folder of folders) {
      const parentId = (folder.parents ?? [])[0]
      if (parentId && byId.has(parentId)) {
        if (!children.has(parentId)) children.set(parentId, [])
        children.get(parentId).push(folder.id)
      }
    }

    const queue = []
    for (const folder of folders) {
      const parentId = (folder.parents ?? [])[0]
      if (!parentId || !byId.has(parentId)) {
        pathCache.set(folder.id, '/' + sanitizeName(folder.name))
        queue.push(folder.id)
      }
    }

    let head = 0
    let processed = 0
    while (head < queue.length) {
      const parentId = queue[head++]
      const parentPath = pathCache.get(parentId)
      for (const childId of (children.get(parentId) ?? [])) {
        const child = byId.get(childId)
        pathCache.set(childId, parentPath + '/' + sanitizeName(child.name))
        queue.push(childId)
      }
      if (++processed % 500 === 0) {
        await new Promise(resolve => setTimeout(resolve, 0))
      }
    }

    return pathCache
  }

  // Signal the current import to stop after the current file finishes.
  abort() {
    this._aborted = true
  }

  /**
   * Import a list of selected Drive files into Kagibi.
   *
   * @param {Array}    selectedFiles  Drive file objects to import (no folders)
   * @param {Map}      pathMap        driveId → kagibi path (from buildPathMap)
   * @param {string}   targetPath     Destination root in Kagibi (e.g. '/' or '/Imports')
   * @param {Object}   callbacks
   * @param {Function} callbacks.onTotal(n)              Total file count
   * @param {Function} callbacks.onFileStart(name, i)    File starting
   * @param {Function} callbacks.onFileDone(name, i)     File completed
   * @param {Function} callbacks.onFileError(name, msg)  File failed
   */
  // Resolve a file's destination folder path in Kagibi.
  // pathMap maps folderId → kagibi path. A file's parent folder gives its directory.
  _resolveFilePath(file, pathMap, targetPath) {
    const parentId = (file.parents ?? [])[0]
    const folderPath = parentId && pathMap.has(parentId) ? pathMap.get(parentId) : ''
    const kagibiDir = folderPath
      ? (targetPath === '/' ? folderPath : targetPath + folderPath)
      : targetPath
    return { kagibiDir, fileName: sanitizeName(file.name) }
  }

  async importItems(selectedFiles, pathMap, targetPath, callbacks) {
    this._aborted = false
    const authStore = useAuthStore()
    if (!authStore.masterKey) throw new Error('Clé maître non disponible')

    const { onTotal, onFileStart, onFileDone, onFileError, onFileSkipped, onFolderConflict, onBytesProgress, onFolderProgress } = callbacks
    onTotal(selectedFiles.length)

    // Pre-compute total encrypted bytes for the progress bar.
    // Workspace files (Docs/Sheets/…) have no Drive metadata size — they count as 28 bytes
    // (1 empty AES-GCM chunk) until the real export size is discovered via onSizeDiscovered.
    let totalImportBytes = 0
    for (const file of selectedFiles) totalImportBytes += calcEncryptedSize(parseInt(file.size ?? 0, 10))
    let globalUploadedBytes = 0
    const fileByteTracker = new Map()   // file.id → last uploaded bytes reported by manager

    const makeFileProgress = (fileId) => (uploaded) => {
      const prev = fileByteTracker.get(fileId) || 0
      const delta = uploaded - prev
      if (delta > 0) {
        fileByteTracker.set(fileId, uploaded)
        globalUploadedBytes += delta
        onBytesProgress?.(globalUploadedBytes, totalImportBytes)
      }
    }

    // When a Workspace file's actual export size is known (after fetching the blob),
    // replace its initial 28-byte estimate in totalImportBytes with the real encrypted size.
    const makeFileSizeUpdate = (estimatedSize) => (actualSize) => {
      totalImportBytes += actualSize - estimatedSize
      onBytesProgress?.(globalUploadedBytes, totalImportBytes)
    }

    // Build the exact set of folders to create: direct parents of selected files
    // AND all their ancestors. Only folders actually needed are touched.
    const neededDirs = new Set()
    for (const file of selectedFiles) {
      const { kagibiDir } = this._resolveFilePath(file, pathMap, targetPath)
      if (kagibiDir === targetPath) continue
      const logicalPath = targetPath === '/' ? kagibiDir : kagibiDir.slice(targetPath.length)
      const segments = logicalPath.split('/').filter(Boolean)
      for (let depth = 1; depth <= segments.length; depth++) {
        neededDirs.add('/' + segments.slice(0, depth).join('/'))
      }
    }
    const sortedFolders = [...neededDirs]
      .sort((a, b) => a.split('/').length - b.split('/').length)

    const skippedPaths = new Set()   // user chose to skip these folder paths
    const invalidPaths = new Map()   // folder path → error message (400 from backend)

    let foldersCreated = 0
    onFolderProgress?.(0, sortedFolders.length)
    for (const logicalPath of sortedFolders) {
      if (this._aborted) return
      const fullPath = targetPath === '/' ? logicalPath : targetPath + logicalPath
      if (_isDescendantOf(fullPath, skippedPaths)) { skippedPaths.add(fullPath); continue }
      if (_isDescendantOf(fullPath, invalidPaths)) {
        invalidPaths.set(fullPath, _getAncestorValue(fullPath, invalidPaths))
        continue
      }
      const result = await this._ensureFolder(logicalPath, targetPath, onFolderConflict)
      if (result === 'skip') skippedPaths.add(fullPath)
      else if (typeof result === 'string' && result.startsWith('invalid:')) {
        invalidPaths.set(fullPath, result.slice(8))
      }
      foldersCreated++
      onFolderProgress?.(foldersCreated, sortedFolders.length)
    }

    // Import files with bounded concurrency: CONCURRENT_FILES pipelines run in parallel.
    // Each pipeline overlaps its download/encrypt/upload with the others, hiding network
    // latency and idle time that sequential processing would waste.
    let activeSlots = 0
    const slotWaiters = []
    const acquireSlot = () => {
      if (activeSlots < CONCURRENT_FILES) { activeSlots++; return Promise.resolve() }
      return new Promise(r => slotWaiters.push(r))
    }
    const releaseSlot = () => {
      activeSlots--
      if (slotWaiters.length > 0) { activeSlots++; slotWaiters.shift()() }
    }

    const filePromises = []
    let idx = 0
    for (const file of selectedFiles) {
      if (this._aborted) break
      const fileIdx = idx++

      const p = acquireSlot().then(async () => {
        if (this._aborted) return
        const { kagibiDir, fileName } = this._resolveFilePath(file, pathMap, targetPath)
        onFileStart(file.name, fileIdx)
        if (skippedPaths.has(kagibiDir)) {
          onFileSkipped?.(file.name)
          return
        }
        if (invalidPaths.has(kagibiDir)) {
          onFileError(file.name, invalidPaths.get(kagibiDir))
          return
        }
        try {
          const estimatedEncSize = calcEncryptedSize(parseInt(file.size ?? 0, 10))
          await this._importFile(file, fileName, kagibiDir, authStore.masterKey, makeFileProgress(file.id), makeFileSizeUpdate(estimatedEncSize))
          onFileDone(file.name, fileIdx)
        } catch (err) {
          onFileError(file.name, err.message ?? String(err))
        }
      }).finally(releaseSlot)

      filePromises.push(p)
    }
    await Promise.all(filePromises)
  }

  // Creates a Kagibi folder at (targetPath + logicalPath).
  // Returns 'created', 'merged', 'skip', or 'invalid:<msg>' (on 400).
  // onConflict(name, fullPath) → Promise<'merge'|'skip'> is called on 409.
  async _ensureFolder(logicalPath, targetPath, onConflict) {
    const segments = logicalPath.split('/').filter(Boolean)
    const name = segments[segments.length - 1]
    const parentSegments = segments.slice(0, -1)
    const parent = parentSegments.length > 0
      ? (targetPath === '/' ? '/' + parentSegments.join('/') : targetPath + '/' + parentSegments.join('/'))
      : targetPath

    try {
      await api.post('/folders/create', { name, path: parent })
      return 'created'
    } catch (err) {
      if (err?.response?.status === 409) {
        if (onConflict) {
          const fullPath = targetPath === '/' ? logicalPath : targetPath + logicalPath
          return await onConflict(name, fullPath)
        }
        return 'merged'
      }
      if (err?.response?.status === 400) {
        return `invalid:${err?.response?.data?.error ?? 'Nom de dossier invalide'}`
      }
      throw err
    }
  }

  // Dispatches to workspace export or regular streaming import.
  // onFileProgress(uploadedBytes) is called as bytes are sent to S3.
  // onSizeDiscovered(actualEncryptedSize) is called once a Workspace file's true size is known.
  async _importFile(driveFile, fileName, kagibiPath, masterKey, onFileProgress, onSizeDiscovered) {
    if (UNSUPPORTED_TYPES.has(driveFile.mimeType)) return

    const fileKey = await generateMasterKey()
    const encryptedFileKey = await wrapMasterKey(fileKey, masterKey)
    const exportInfo = WORKSPACE_EXPORTS[driveFile.mimeType]
    const mgrProgress = onFileProgress
      ? (_, uploaded) => onFileProgress(uploaded)
      : undefined

    if (exportInfo) {
      await this._importWorkspace(driveFile, fileName + exportInfo.ext, kagibiPath, fileKey, encryptedFileKey, exportInfo.mime, mgrProgress, onSizeDiscovered)
    } else {
      const plainSize = parseInt(driveFile.size ?? "0", 10)
      await this._importRegular(driveFile, fileName, plainSize, kagibiPath, fileKey, encryptedFileKey, mgrProgress, onSizeDiscovered)
    }
  }

  // Regular binary file: streams from Drive → encrypts chunk by chunk → uploads to S3.
  // Peak browser RAM ≈ MAX_CONCURRENT_WORKERS × PART_SIZE (each part freed immediately after upload).
  async _importRegular(driveFile, fileName, plainSize, kagibiPath, fileKey, encryptedFileKey, onProgress, onSizeDiscovered) {
    const chunkSize = pickChunkSize(plainSize)
    const totalEncryptedSize = calcEncryptedSize(plainSize, chunkSize)
    const manager = new MultipartUploadManager({ onProgress })
    await manager.initiate(fileName, kagibiPath, "application/octet-stream", totalEncryptedSize, encryptedFileKey, chunkSize)

    if (plainSize === 0) {
      // Empty file: produce the standard 28-byte encrypted blob (nonce + auth tag only)
      const baseNonce = generateBaseNonce()
      const blob = await encryptChunkWorker(new ArrayBuffer(0), fileKey, 0, baseNonce)
      const parts = await manager.uploadParts([blob])
      const backendSize = await this._completeUpload(manager, parts, fileName, kagibiPath, totalEncryptedSize, encryptedFileKey)
      onSizeDiscovered?.(backendSize)
      return
    }

    const res = await fetch(`${DRIVE_API}/files/${driveFile.id}?alt=media`, {
      headers: { Authorization: `Bearer ${this._accessToken}` }
    })
    if (!res.ok) throw new Error(`Téléchargement échoué (${res.status})`)

    const parts = await manager.uploadPartsStreamed(this._encryptStream(res, fileKey, chunkSize))
    const backendSize = await this._completeUpload(manager, parts, fileName, kagibiPath, totalEncryptedSize, encryptedFileKey)
    onSizeDiscovered?.(backendSize)
  }

  // Google Workspace file: export is buffered once (needed for size), then encrypted
  // and uploaded via the streaming pipeline — no encrypted-chunk array accumulation.
  async _importWorkspace(driveFile, fileName, kagibiPath, fileKey, encryptedFileKey, exportMime, onProgress, onSizeDiscovered) {
    const res = await fetch(
      `${DRIVE_API}/files/${driveFile.id}/export?mimeType=${encodeURIComponent(exportMime)}`,
      { headers: { Authorization: `Bearer ${this._accessToken}` } }
    )
    if (!res.ok) throw new Error(`Export Workspace échoué (${res.status})`)

    const blob = await res.blob()
    // chunk size is chosen after the blob is downloaded (blob.size is the true plaintext size)
    const chunkSize = pickChunkSize(blob.size)
    const totalEncryptedSize = calcEncryptedSize(blob.size, chunkSize)
    const manager = new MultipartUploadManager({ onProgress })
    await manager.initiate(fileName, kagibiPath, "application/octet-stream", totalEncryptedSize, encryptedFileKey, chunkSize)

    // Empty export edge case: produce the standard 28-byte AES-GCM blob (nonce + auth tag).
    if (blob.size === 0) {
      const baseNonce = generateBaseNonce()
      const emptyChunk = await encryptChunkWorker(new ArrayBuffer(0), fileKey, 0, baseNonce)
      const parts = await manager.uploadParts([emptyChunk])
      const backendSize = await this._completeUpload(manager, parts, fileName, kagibiPath, totalEncryptedSize, encryptedFileKey)
      onSizeDiscovered?.(backendSize)
      return
    }

    const parts = await manager.uploadPartsStreamed(this._encryptBlobStream(blob, fileKey, chunkSize))
    const backendSize = await this._completeUpload(manager, parts, fileName, kagibiPath, totalEncryptedSize, encryptedFileKey)
    onSizeDiscovered?.(backendSize)
  }

  async _completeUpload(manager, parts, fileName, filePath, totalSize, encryptedKey) {
    const result = await manager.complete(parts, {
      fileName,
      filePath,
      totalSize,
      contentType: 'application/octet-stream',
      encryptedKey,
      shareKeys: '',
      previewId: null,
      isPreview: false
    })
    // The backend stores the HeadObject-confirmed actual size; use it to keep the
    // prediction in sync with what is truly written to S3 (Drive metadata can diverge).
    return result?.file?.size ?? totalSize
  }

  // Async generator: buffers the Drive response stream into chunkSize chunks,
  // encrypts each chunk, and yields the resulting Blobs one at a time.
  // Peak RAM ≈ 2 × chunkSize regardless of file size.
  async *_encryptStream(response, fileKey, chunkSize = PART_SIZE) {
    const baseNonce = generateBaseNonce()
    let chunkIdx = 0
    for await (const chunkBuffer of this._bufferStream(response, chunkSize)) {
      if (this._aborted) throw new Error('Import annulé')
      yield await encryptChunkWorker(chunkBuffer, fileKey, chunkIdx++, baseNonce)
    }
  }

  // Async generator for Workspace blobs: slices into chunkSize pieces one at a time,
  // encrypts each, and yields Blobs — same bounded RAM profile as _encryptStream.
  async *_encryptBlobStream(blob, fileKey, chunkSize = PART_SIZE) {
    const baseNonce = generateBaseNonce()
    let offset = 0, chunkIdx = 0
    while (offset < blob.size) {
      if (this._aborted) throw new Error("Import annulé")
      const chunk = await blob.slice(offset, offset + chunkSize).arrayBuffer()
      yield await encryptChunkWorker(chunk, fileKey, chunkIdx++, baseNonce)
      offset += chunkSize
    }
  }

  // Reads a fetch Response as a stream and yields ArrayBuffers aligned to chunkSize.
  async *_bufferStream(response, chunkSize = PART_SIZE) {
    const reader = response.body.getReader()
    let buf = new Uint8Array(chunkSize)
    let fill = 0
    try {
      while (true) {
        const { done, value } = await reader.read()
        if (done) {
          // IMPORTANT: buf.buffer is the full chunkSize ArrayBuffer; slice to exact fill length.
          // (buf.subarray(0, fill).buffer would return the full underlying buffer — wrong.)
          if (fill > 0) yield buf.buffer.slice(0, fill)
          return
        }
        let srcOff = 0
        while (srcOff < value.length) {
          const toCopy = Math.min(chunkSize - fill, value.length - srcOff)
          buf.set(value.subarray(srcOff, srcOff + toCopy), fill)
          fill += toCopy
          srcOff += toCopy
          if (fill === chunkSize) {
            yield buf.buffer.slice(0)
            buf = new Uint8Array(chunkSize)
            fill = 0
          }
        }
      }
    } finally {
      reader.releaseLock()
    }
  }
}

// ── Helpers exported for use in the dialog component ──

export function getImportableFiles(items) {
  return items.filter(item =>
    item.mimeType !== 'application/vnd.google-apps.folder' &&
    !UNSUPPORTED_TYPES.has(item.mimeType)
  )
}

export function getFolders(items) {
  return items.filter(item => item.mimeType === 'application/vnd.google-apps.folder')
}

export function isWorkspaceFile(mimeType) {
  return !!WORKSPACE_EXPORTS[mimeType]
}

export function workspaceExtension(mimeType) {
  return WORKSPACE_EXPORTS[mimeType]?.ext ?? ''
}

export function formatBytes(bytes) {
  if (!bytes || bytes === 0) return '—'
  if (bytes < 1024) return `${bytes} o`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} Ko`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} Mo`
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} Go`
}
