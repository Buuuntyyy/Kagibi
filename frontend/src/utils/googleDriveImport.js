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
import { MultipartUploadManager, PART_SIZE } from './multipartUpload'
import { useAuthStore } from '../stores/auth'

const DRIVE_API = 'https://www.googleapis.com/drive/v3'

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

// Kagibi backend folder name validation: /^[\p{L}\p{N}\s\-\._'‘’]+$/u
// Characters outside this set are replaced by underscore.
function sanitizeName(name) {
  return name.replace(/[^\p{L}\p{N}\s\-._'‘’]/gu, '_').trim() || 'Import'
}

// Returns total encrypted byte count for a given plaintext size.
// AES-256-GCM adds (NONCE_LENGTH + TAG_LENGTH_BYTES) = 28 bytes per chunk.
// Empty files produce exactly one chunk of 28 bytes (nonce + auth tag, no ciphertext).
function calcEncryptedSize(plainSize) {
  if (plainSize === 0) return NONCE_LENGTH + TAG_LENGTH_BYTES
  const numParts = Math.ceil(plainSize / PART_SIZE)
  return plainSize + numParts * (NONCE_LENGTH + TAG_LENGTH_BYTES)
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

  // Fetches all folders first (typically far fewer than files), then all files.
  // Returns { folders, files } separately so buildPathMap only processes folders.
  // Google Photos albums are excluded: they are owned by the user (so 'me' in owners is true)
  // but they never appear in the "My Drive" tab of the Drive UI.
  async listAllItems(onProgress) {
    const gPhotosId = await this._getGooglePhotosFolderId()
    const exclude = gPhotosId
      ? ` and not '${gPhotosId}' in ancestors and id != '${gPhotosId}'`
      : ''

    const folders = await this._fetchPage(
      `mimeType='application/vnd.google-apps.folder' and trashed=false and 'root' in ancestors${exclude}`,
      'id,name,parents',
      onProgress ? (n) => onProgress('folders', n) : null
    )
    const files = await this._fetchPage(
      `mimeType!='application/vnd.google-apps.folder' and trashed=false and 'root' in ancestors${exclude}`,
      'id,name,mimeType,size,parents,modifiedTime',
      onProgress ? (n) => onProgress('files', n) : null
    )
    return { folders, files }
  }

  // Resolves the Drive-managed "Google Photos" root folder ID so it can be excluded
  // from listing queries. Returns null if the folder does not exist or cannot be found.
  async _getGooglePhotosFolderId() {
    const params = new URLSearchParams({
      q: "name='Google Photos' and mimeType='application/vnd.google-apps.folder' and 'root' in parents and 'me' in owners",
      fields: 'files(id)',
      pageSize: '1'
    })
    try {
      const res = await fetch(`${DRIVE_API}/files?${params}`, {
        headers: { Authorization: `Bearer ${this._accessToken}` }
      })
      if (!res.ok) return null
      const data = await res.json()
      return data.files?.[0]?.id ?? null
    } catch {
      return null
    }
  }

  async _fetchPage(q, fileFields, onCount) {
    const items = []
    let pageToken = null
    do {
      const params = new URLSearchParams({
        fields: `nextPageToken,files(${fileFields})`,
        pageSize: '1000',
        q,
        orderBy: 'name'
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

    const { onTotal, onFileStart, onFileDone, onFileError } = callbacks
    onTotal(selectedFiles.length)

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

    for (const logicalPath of sortedFolders) {
      if (this._aborted) return
      await this._ensureFolder(logicalPath, targetPath)
    }

    // Import files sequentially — one encryption pipeline at a time.
    let idx = 0
    for (const file of selectedFiles) {
      if (this._aborted) return
      const { kagibiDir, fileName } = this._resolveFilePath(file, pathMap, targetPath)
      onFileStart(file.name, idx)
      try {
        await this._importFile(file, fileName, kagibiDir, authStore.masterKey)
        onFileDone(file.name, idx)
      } catch (err) {
        onFileError(file.name, err.message ?? String(err))
      }
      idx++
    }
  }

  // Creates a Kagibi folder at (targetPath + logicalPath). Ignores 409 conflicts.
  async _ensureFolder(logicalPath, targetPath) {
    const segments = logicalPath.split('/').filter(Boolean)
    const name = segments[segments.length - 1]
    const parentSegments = segments.slice(0, -1)
    const parent = parentSegments.length > 0
      ? (targetPath === '/' ? '/' + parentSegments.join('/') : targetPath + '/' + parentSegments.join('/'))
      : targetPath

    try {
      await api.post('/folders/create', { name, path: parent })
    } catch (err) {
      if (err?.response?.status !== 409) throw err
    }
  }

  // Dispatches to workspace export or regular streaming import.
  async _importFile(driveFile, fileName, kagibiPath, masterKey) {
    if (UNSUPPORTED_TYPES.has(driveFile.mimeType)) return

    const fileKey = await generateMasterKey()
    const encryptedFileKey = await wrapMasterKey(fileKey, masterKey)
    const exportInfo = WORKSPACE_EXPORTS[driveFile.mimeType]

    if (exportInfo) {
      await this._importWorkspace(driveFile, fileName + exportInfo.ext, kagibiPath, fileKey, encryptedFileKey, exportInfo.mime)
    } else {
      const plainSize = parseInt(driveFile.size ?? '0', 10)
      await this._importRegular(driveFile, fileName, plainSize, kagibiPath, fileKey, encryptedFileKey)
    }
  }

  // Regular binary file: streams from Drive → encrypts chunk by chunk → uploads to S3.
  // Peak browser RAM usage ≈ 2–3 × PART_SIZE (one in-flight download chunk + one encrypted chunk).
  async _importRegular(driveFile, fileName, plainSize, kagibiPath, fileKey, encryptedFileKey) {
    const totalEncryptedSize = calcEncryptedSize(plainSize)
    const manager = new MultipartUploadManager({})
    await manager.initiate(fileName, kagibiPath, 'application/octet-stream', totalEncryptedSize, encryptedFileKey)

    if (plainSize === 0) {
      // Empty file: produce the standard 28-byte encrypted blob (nonce + auth tag only)
      const baseNonce = generateBaseNonce()
      const blob = await encryptChunkWorker(new ArrayBuffer(0), fileKey, 0, baseNonce)
      const parts = await manager.uploadParts([blob])
      await this._completeUpload(manager, parts, fileName, kagibiPath, totalEncryptedSize, encryptedFileKey)
      return
    }

    const res = await fetch(`${DRIVE_API}/files/${driveFile.id}?alt=media`, {
      headers: { Authorization: `Bearer ${this._accessToken}` }
    })
    if (!res.ok) throw new Error(`Téléchargement échoué (${res.status})`)

    const parts = await manager.uploadPartsStreamed(this._encryptStream(res, fileKey))
    await this._completeUpload(manager, parts, fileName, kagibiPath, totalEncryptedSize, encryptedFileKey)
  }

  // Google Workspace file: downloaded as an Office export (full blob in RAM, typically small),
  // then encrypted and uploaded using the standard pipeline.
  async _importWorkspace(driveFile, fileName, kagibiPath, fileKey, encryptedFileKey, exportMime) {
    const res = await fetch(
      `${DRIVE_API}/files/${driveFile.id}/export?mimeType=${encodeURIComponent(exportMime)}`,
      { headers: { Authorization: `Bearer ${this._accessToken}` } }
    )
    if (!res.ok) throw new Error(`Export Workspace échoué (${res.status})`)

    const blob = await res.blob()
    const totalEncryptedSize = calcEncryptedSize(blob.size)
    const manager = new MultipartUploadManager({})
    await manager.initiate(fileName, kagibiPath, 'application/octet-stream', totalEncryptedSize, encryptedFileKey)

    // Empty export edge case: produce the standard 28-byte AES-GCM blob (nonce + auth tag).
    if (blob.size === 0) {
      const baseNonce = generateBaseNonce()
      const emptyChunk = await encryptChunkWorker(new ArrayBuffer(0), fileKey, 0, baseNonce)
      const parts = await manager.uploadParts([emptyChunk])
      await this._completeUpload(manager, parts, fileName, kagibiPath, totalEncryptedSize, encryptedFileKey)
      return
    }

    const baseNonce = generateBaseNonce()
    const encryptedChunks = []
    let offset = 0, chunkIdx = 0
    while (offset < blob.size) {
      const chunk = await blob.slice(offset, offset + PART_SIZE).arrayBuffer()
      encryptedChunks.push(await encryptChunkWorker(chunk, fileKey, chunkIdx++, baseNonce))
      offset += PART_SIZE
    }

    const parts = await manager.uploadParts(encryptedChunks)
    await this._completeUpload(manager, parts, fileName, kagibiPath, totalEncryptedSize, encryptedFileKey)
  }

  async _completeUpload(manager, parts, fileName, filePath, totalSize, encryptedKey) {
    await manager.complete(parts, {
      fileName,
      filePath,
      totalSize,
      contentType: 'application/octet-stream',
      encryptedKey,
      shareKeys: '',
      previewId: null,
      isPreview: false
    })
  }

  // Async generator: buffers the Drive response stream into PART_SIZE chunks,
  // encrypts each chunk, and yields the resulting Blobs one at a time.
  // This keeps peak RAM at roughly 2 × PART_SIZE regardless of file size.
  async *_encryptStream(response, fileKey) {
    const baseNonce = generateBaseNonce()
    let chunkIdx = 0
    for await (const chunkBuffer of this._bufferStream(response)) {
      if (this._aborted) throw new Error('Import annulé')
      yield await encryptChunkWorker(chunkBuffer, fileKey, chunkIdx++, baseNonce)
    }
  }

  // Reads a fetch Response as a stream and yields ArrayBuffers aligned to PART_SIZE.
  async *_bufferStream(response) {
    const reader = response.body.getReader()
    let buf = new Uint8Array(PART_SIZE)
    let fill = 0
    try {
      while (true) {
        const { done, value } = await reader.read()
        if (done) {
          if (fill > 0) yield buf.subarray(0, fill).buffer.slice(0)
          return
        }
        let srcOff = 0
        while (srcOff < value.length) {
          const toCopy = Math.min(PART_SIZE - fill, value.length - srcOff)
          buf.set(value.subarray(srcOff, srcOff + toCopy), fill)
          fill += toCopy
          srcOff += toCopy
          if (fill === PART_SIZE) {
            yield buf.buffer.slice(0)
            buf = new Uint8Array(PART_SIZE)
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
