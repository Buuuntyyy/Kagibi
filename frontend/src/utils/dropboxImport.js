// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

/**
 * Dropbox Import — zero-knowledge pipeline
 *
 * Data visibility contract:
 *   - Kagibi BACKEND never receives: file content (plaintext), Dropbox OAuth token
 *   - Kagibi FRONTEND (browser RAM only, never persisted): OAuth access token,
 *     file/folder names, file sizes, plaintext content (chunk by chunk only)
 *   - Kagibi servers receive: same encrypted blobs as any regular upload
 *   - Dropbox receives: OAuth consent event with Kagibi's app key; files.list_folder /
 *     files.download API calls
 *
 * Like OneDrive, Dropbox uses a direct PKCE flow for public clients: the browser
 * exchanges the authorization code with Dropbox's token endpoint itself — there is no
 * backend proxy/desktop-token endpoint, since Dropbox does not require a client secret
 * for PKCE (confirmed by Dropbox's own JS SDK, which performs this exchange from the
 * browser with fetch()).
 *
 * Architecture:
 *   Dropbox content API (streaming fetch) → buffer to chunkSize → encryptChunkWorker
 *     → MultipartUploadManager → S3 (encrypted blobs only)
 */

import api from '../api'
import { generateMasterKey, wrapMasterKey, generateBaseNonce, encryptChunkWorker } from './crypto'
import { MultipartUploadManager, pickChunkSize } from './multipartUpload'
import { useAuthStore } from '../stores/auth'
import {
  calcEncryptedSize,
  buildIdPathMap,
  resolveItemPath,
  generateRandomString,
  generateCodeChallenge,
  encryptStream,
  ensureFoldersBatch,
  formatBytes,
} from './importShared'

const API_BASE = 'https://api.dropboxapi.com/2'
const CONTENT_BASE = 'https://content.dropboxapi.com/2'
const AUTH_ENDPOINT = 'https://www.dropbox.com/oauth2/authorize'
const TOKEN_ENDPOINT = 'https://api.dropboxapi.com/oauth2/token'
const SCOPE = 'files.metadata.read files.content.read'

// Number of files processed concurrently (download + encrypt + upload in parallel).
// Each slot uses ≤ MAX_CONCURRENT_WORKERS × PART_SIZE ≈ 30 MB of RAM.
const CONCURRENT_FILES = 3

export class DropboxImport {
  constructor() {
    this._clientId = null
    this._accessToken = null
    this._aborted = false
  }

  // Fetch app key from backend. Must be called once before authenticate().
  async init() {
    const res = await api.get('/import/dropbox/config')
    this._clientId = res.data.app_key
  }

  get isConfigured() {
    return !!this._clientId
  }

  // Opens the Dropbox OAuth consent popup (PKCE, public client) and stores the
  // short-lived access token in memory only. The token is never sent to Kagibi's backend,
  // and the code exchange happens directly against Dropbox — no backend proxy involved.
  authenticate() {
    return new Promise((resolve, reject) => {
      let popup, pollTimer
      const messageHandler = async (event) => {
        if (event.origin !== window.location.origin) return
        if (!event.data || event.data.state !== state) return
        cleanup()
        popup?.close()
        if (event.data.error) {
          reject(new Error(event.data.error_description || event.data.error))
          return
        }
        try {
          this._accessToken = await this._exchangeCode(event.data.code, verifier, redirectUri)
          resolve()
        } catch (err) {
          reject(err)
        }
      }
      const cleanup = () => {
        window.removeEventListener('message', messageHandler)
        if (pollTimer) clearInterval(pollTimer)
      }

      const verifier = generateRandomString(64)
      const state = generateRandomString(32)
      const redirectUri = `${window.location.origin}/dropbox-callback.html`

      generateCodeChallenge(verifier).then((challenge) => {
        const params = new URLSearchParams({
          client_id: this._clientId,
          response_type: 'code',
          redirect_uri: redirectUri,
          token_access_type: 'online',
          scope: SCOPE,
          state,
          code_challenge: challenge,
          code_challenge_method: 'S256',
        })

        popup = window.open(`${AUTH_ENDPOINT}?${params}`, 'kagibi-dropbox-oauth', 'width=500,height=650')
        if (!popup) {
          reject(new Error("Impossible d'ouvrir la fenêtre d'authentification (bloquée par le navigateur ?)"))
          return
        }

        window.addEventListener('message', messageHandler)
        pollTimer = setInterval(() => {
          if (popup.closed) {
            cleanup()
            reject(new Error('Authentification Dropbox annulée'))
          }
        }, 500)
      }).catch((err) => {
        cleanup()
        reject(err)
      })
    })
  }

  async _exchangeCode(code, verifier, redirectUri) {
    const res = await fetch(TOKEN_ENDPOINT, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({
        client_id: this._clientId,
        grant_type: 'authorization_code',
        code,
        redirect_uri: redirectUri,
        code_verifier: verifier,
      })
    })
    const data = await res.json().catch(() => ({}))
    if (!res.ok || data.error) {
      throw new Error(data.error_description || data.error_summary || data.error || `Échange du token Dropbox échoué (${res.status})`)
    }
    return data.access_token
  }

  // Fetches the whole account tree via files/list_folder (recursive, paginated) and
  // splits it into folders/files. Unlike OneDrive/Drive, Dropbox entries carry their full
  // path directly (path_lower) instead of a parent id — parentId is derived by mapping
  // each folder's path_lower to its id, then looking up the dirname of every item's path.
  // include_non_downloadable_files:false excludes Dropbox Paper docs and similar items
  // that have no /files/download endpoint, so no separate unsupported-type filtering is
  // needed here (unlike Google Workspace docs or OneNote packages for the other providers).
  async listAllItems(onProgress) {
    const rawEntries = await this._listFolder(onProgress)

    const folderIdByPath = new Map()
    for (const e of rawEntries) {
      if (e['.tag'] === 'folder') folderIdByPath.set(e.path_lower, e.id)
    }
    const parentIdOf = (pathLower) => {
      const idx = pathLower.lastIndexOf('/')
      const parentPath = idx > 0 ? pathLower.slice(0, idx) : ''
      return parentPath ? (folderIdByPath.get(parentPath) ?? null) : null
    }

    const folders = []
    const files = []
    for (const e of rawEntries) {
      const parentId = parentIdOf(e.path_lower)
      if (e['.tag'] === 'folder') {
        folders.push({ id: e.id, name: e.name, parentId, isFolder: true })
      } else if (e['.tag'] === 'file') {
        files.push({ id: e.id, name: e.name, size: e.size ?? 0, parentId, isFolder: false })
      }
    }

    files.sort((a, b) => a.name.localeCompare(b.name))
    return { folders, files }
  }

  async _listFolder(onProgress) {
    const entries = []
    let data = await this._fetchWithRetry(`${API_BASE}/files/list_folder`, {
      path: '',
      recursive: true,
      include_non_downloadable_files: false,
    })
    entries.push(...(data.entries ?? []))
    onProgress?.(entries.length)

    while (data.has_more) {
      data = await this._fetchWithRetry(`${API_BASE}/files/list_folder/continue`, { cursor: data.cursor })
      entries.push(...(data.entries ?? []))
      onProgress?.(entries.length)
    }
    return entries
  }

  async _fetchWithRetry(url, body, attempt = 0) {
    const res = await fetch(url, {
      method: 'POST',
      headers: { Authorization: `Bearer ${this._accessToken}`, 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    })
    if (res.status === 429 && attempt < 5) {
      let retryAfter = parseInt(res.headers.get('Retry-After') ?? '', 10)
      if (!retryAfter) {
        const errData = await res.json().catch(() => ({}))
        retryAfter = errData?.error?.retry_after ?? 2
      }
      await new Promise(r => setTimeout(r, Math.max(1, retryAfter) * 1000))
      return this._fetchWithRetry(url, body, attempt + 1)
    }
    if (!res.ok) {
      const errData = await res.json().catch(() => ({}))
      throw new Error(errData?.error_summary || `Dropbox API: ${res.status}`)
    }
    return res.json()
  }

  // Builds a Map<folderId, sanitizedKagibiPath> from folders only (much fewer than files).
  async buildPathMap(folders) {
    return buildIdPathMap(folders)
  }

  // Signal the current import to stop after the current file finishes.
  abort() {
    this._aborted = true
  }

  /**
   * Import a list of selected Dropbox files into Kagibi.
   *
   * @param {Array}    selectedFiles  normalized file items (no folders) — see listAllItems
   * @param {Map}      pathMap        dropboxId → kagibi path (from buildPathMap)
   * @param {string}   targetPath     Destination root in Kagibi (e.g. '/' or '/Imports')
   * @param {Object}   callbacks
   * @param {Function} callbacks.onTotal(n)              Total file count
   * @param {Function} callbacks.onFileStart(name, i)    File starting
   * @param {Function} callbacks.onFileDone(name, i)     File completed
   * @param {Function} callbacks.onFileError(name, msg)  File failed
   */
  async importItems(selectedFiles, pathMap, targetPath, callbacks) {
    this._aborted = false
    const authStore = useAuthStore()
    if (!authStore.masterKey) throw new Error('Clé maître non disponible')

    const { onTotal, onFileStart, onFileDone, onFileError, onBytesProgress, onFolderProgress } = callbacks
    onTotal(selectedFiles.length)

    // Pre-compute total encrypted bytes for the progress bar.
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

    // Build the exact set of folders to create: direct parents of selected files
    // AND all their ancestors. Only folders actually needed are touched.
    const neededDirs = new Set()
    for (const file of selectedFiles) {
      const { kagibiDir } = resolveItemPath(file, pathMap, targetPath)
      if (kagibiDir === targetPath) continue
      const logicalPath = targetPath === '/' ? kagibiDir : kagibiDir.slice(targetPath.length)
      const segments = logicalPath.split('/').filter(Boolean)
      for (let depth = 1; depth <= segments.length; depth++) {
        neededDirs.add('/' + segments.slice(0, depth).join('/'))
      }
    }
    const sortedFolders = [...neededDirs]
      .sort((a, b) => a.split('/').length - b.split('/').length)

    if (this._aborted) return
    const { invalidPaths } = await ensureFoldersBatch(sortedFolders, targetPath, authStore.masterKey, onFolderProgress)

    // Import files with bounded concurrency: CONCURRENT_FILES pipelines run in parallel.
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
        const { kagibiDir, fileName } = resolveItemPath(file, pathMap, targetPath)
        onFileStart(file.name, fileIdx)
        if (invalidPaths.has(kagibiDir)) {
          onFileError(file.name, invalidPaths.get(kagibiDir))
          return
        }
        try {
          await this._importFile(file, fileName, kagibiDir, authStore.masterKey, makeFileProgress(file.id))
          onFileDone(file.name, fileIdx)
        } catch (err) {
          onFileError(file.name, err.message ?? String(err))
        }
      }).finally(releaseSlot)

      filePromises.push(p)
    }
    await Promise.all(filePromises)
  }

  // Regular binary file: streams from Dropbox → encrypts chunk by chunk → uploads to S3.
  // No export step exists for Dropbox — every listed file is already downloadable as-is
  // (Paper docs etc. were excluded server-side via include_non_downloadable_files:false).
  // Peak browser RAM ≈ MAX_CONCURRENT_WORKERS × PART_SIZE (each part freed immediately after upload).
  async _importFile(item, fileName, kagibiPath, masterKey, onFileProgress) {
    const fileKey = await generateMasterKey()
    const encryptedFileKey = await wrapMasterKey(fileKey, masterKey)
    const mgrProgress = onFileProgress ? (_, uploaded) => onFileProgress(uploaded) : undefined
    const plainSize = parseInt(item.size ?? 0, 10)

    const chunkSize = pickChunkSize(plainSize)
    const totalEncryptedSize = calcEncryptedSize(plainSize, chunkSize)
    const manager = new MultipartUploadManager({ onProgress: mgrProgress })
    await manager.initiate(fileName, kagibiPath, "application/octet-stream", totalEncryptedSize, encryptedFileKey, chunkSize)

    if (plainSize === 0) {
      const parts = await manager.uploadParts([await this._emptyChunk(fileKey)])
      await this._completeUpload(manager, parts, fileName, kagibiPath, totalEncryptedSize, encryptedFileKey)
      return
    }

    // Dropbox-API-Arg identifies the file by its stable Dropbox id ("id:xxxx") rather than
    // its path. HTTP headers must be ASCII-only, and Dropbox's own docs require an
    // ASCII-safe JSON encoding for this header when a path contains non-ASCII characters
    // (e.g. accented filenames) — using the id sidesteps that entirely, since ids are
    // already plain ASCII.
    const res = await fetch(`${CONTENT_BASE}/files/download`, {
      headers: {
        Authorization: `Bearer ${this._accessToken}`,
        'Dropbox-API-Arg': JSON.stringify({ path: item.id }),
      }
    })
    if (!res.ok) throw new Error(`Téléchargement échoué (${res.status})`)

    const parts = await manager.uploadPartsStreamed(encryptStream(res, fileKey, chunkSize, () => this._aborted))
    await this._completeUpload(manager, parts, fileName, kagibiPath, totalEncryptedSize, encryptedFileKey)
  }

  // Empty file: produce the standard 28-byte encrypted blob (nonce + auth tag only).
  async _emptyChunk(fileKey) {
    const baseNonce = generateBaseNonce()
    return encryptChunkWorker(new ArrayBuffer(0), fileKey, 0, baseNonce)
  }

  async _completeUpload(manager, parts, fileName, filePath, totalSize, encryptedKey) {
    return manager.complete(parts, {
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
}

// ── Helpers exported for use in the dialog component ──

export function getImportableFiles(items) {
  return items.filter(item => !item.isFolder)
}

export function getFolders(items) {
  return items.filter(item => item.isFolder)
}

export { formatBytes }
