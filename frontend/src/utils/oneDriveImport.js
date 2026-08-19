// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

/**
 * OneDrive Import — zero-knowledge pipeline
 *
 * Data visibility contract:
 *   - Kagibi BACKEND never receives: file content (plaintext), Microsoft OAuth token
 *   - Kagibi FRONTEND (browser RAM only, never persisted): OAuth access token,
 *     file/folder names, file sizes, MIME types, plaintext content (chunk by chunk only)
 *   - Kagibi servers receive: same encrypted blobs as any regular upload
 *   - Microsoft receives: OAuth consent event with Kagibi's client ID; Graph read API calls
 *
 * Unlike Google Drive import, OneDrive uses a direct PKCE flow for public clients:
 * the browser exchanges the authorization code with Microsoft's token endpoint itself —
 * there is no backend proxy/desktop-token endpoint, since Microsoft does not require a
 * client secret for a "Single-page application" platform registration.
 *
 * Architecture:
 *   Microsoft Graph (streaming fetch) → buffer to chunkSize → encryptChunkWorker
 *     → MultipartUploadManager → S3 (encrypted blobs only)
 */

import api from '../api'
import { generateMasterKey, wrapMasterKey, generateBaseNonce, encryptChunkWorker } from './crypto'
import { MultipartUploadManager, pickChunkSize } from './multipartUpload'
import { useAuthStore } from '../stores/auth'
import {
  sanitizeName,
  calcEncryptedSize,
  buildIdPathMap,
  resolveItemPath,
  generateRandomString,
  generateCodeChallenge,
  encryptStream,
  ensureFoldersBatch,
  formatBytes,
} from './importShared'

const GRAPH_API = 'https://graph.microsoft.com/v1.0'
const AUTH_ENDPOINT = 'https://login.microsoftonline.com/common/oauth2/v2.0/authorize'
const TOKEN_ENDPOINT = 'https://login.microsoftonline.com/common/oauth2/v2.0/token'
const SCOPE = 'Files.Read.All'

// Number of files processed concurrently (download + encrypt + upload in parallel).
// Each slot uses ≤ MAX_CONCURRENT_WORKERS × PART_SIZE ≈ 30 MB of RAM.
const CONCURRENT_FILES = 3

// "Package" items (OneNote notebooks) have no /content endpoint — GET returns 400.
// Unlike Google Workspace docs, no export format exists for these; they're skipped.
const UNSUPPORTED_PACKAGE_TYPES = new Set(['oneNote'])

export class OneDriveImport {
  constructor() {
    this._clientId = null
    this._accessToken = null
    this._aborted = false
  }

  // Fetch client ID from backend. Must be called once before authenticate().
  async init() {
    const res = await api.get('/import/onedrive/config')
    this._clientId = res.data.client_id
  }

  get isConfigured() {
    return !!this._clientId
  }

  // Opens the Microsoft OAuth consent popup (PKCE, public client) and stores the
  // short-lived access token in memory only. The token is never sent to Kagibi's backend,
  // and the code exchange happens directly against Microsoft — no backend proxy involved.
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
      const redirectUri = `${window.location.origin}/onedrive-callback.html`

      generateCodeChallenge(verifier).then((challenge) => {
        const params = new URLSearchParams({
          client_id: this._clientId,
          response_type: 'code',
          redirect_uri: redirectUri,
          response_mode: 'query',
          scope: SCOPE,
          state,
          code_challenge: challenge,
          code_challenge_method: 'S256',
        })

        popup = window.open(`${AUTH_ENDPOINT}?${params}`, 'kagibi-onedrive-oauth', 'width=500,height=650')
        if (!popup) {
          reject(new Error("Impossible d'ouvrir la fenêtre d'authentification (bloquée par le navigateur ?)"))
          return
        }

        window.addEventListener('message', messageHandler)
        pollTimer = setInterval(() => {
          if (popup.closed) {
            cleanup()
            reject(new Error('Authentification OneDrive annulée'))
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
        scope: SCOPE,
      })
    })
    const data = await res.json().catch(() => ({}))
    if (!res.ok || data.error) {
      throw new Error(data.error_description || data.error || `Échange du token OneDrive échoué (${res.status})`)
    }
    return data.access_token
  }

  // Fetches the whole drive tree via the Graph delta endpoint (paginated) and splits it
  // into folders/files. Unlike Drive's query API, delta always returns the entire tree in
  // one flat paginated stream — there is no server-side "files only" filter, so items are
  // split client-side by facet presence (item.folder vs item.file), same principle as the
  // Google importer's client-side filtering.
  async listAllItems(onProgress) {
    const rawItems = await this._fetchDelta(onProgress)

    // Delta can return the same item more than once across pages if it changes while the
    // crawl is still in progress (e.g. a file being actively synced) — Microsoft's own
    // docs say later occurrences supersede earlier ones. Collapse by id, keeping only the
    // last state, before building folders/files: without this, a single duplicated
    // ancestor folder can multiply its entire subtree during buildPathMap's BFS, turning
    // into a runaway client-side loop on large/active drives.
    const byId = new Map()
    for (const item of rawItems) byId.set(item.id, item)

    const folders = []
    const files = []

    for (const item of byId.values()) {
      // Delta includes tombstones for previously-seen deleted items, and the drive root
      // item itself (identified by the "root" facet) — both must be skipped.
      if (item.deleted || item.root) continue
      const parentId = item.parentReference?.id ?? null

      if (item.folder) {
        folders.push({ id: item.id, name: item.name, parentId, isFolder: true })
      } else if (item.file) {
        files.push({
          id: item.id,
          name: item.name,
          mimeType: item.file.mimeType,
          size: item.size ?? 0,
          parentId,
          isFolder: false,
          packageType: item.package?.type ?? null,
        })
      }
    }

    files.sort((a, b) => a.name.localeCompare(b.name))
    return { folders, files }
  }

  async _fetchDelta(onProgress) {
    const items = []
    let url = `${GRAPH_API}/me/drive/root/delta?$select=id,name,size,file,folder,parentReference,deleted,package,root`
    while (url) {
      const data = await this._fetchWithRetry(url)
      items.push(...(data.value ?? []))
      onProgress?.(items.length)
      // @odata.nextLink is a complete, ready-to-fetch URL (already carries an opaque
      // delta token) — unlike Drive's pageToken, it must be called as-is, not rebuilt.
      url = data['@odata.nextLink'] ?? null
    }
    return items
  }

  // Microsoft Graph throttles bulk listing more aggressively than Drive's per-user quota;
  // 429/503/509 responses carry a Retry-After (seconds) header that must be honored.
  async _fetchWithRetry(url, attempt = 0) {
    const res = await fetch(url, { headers: { Authorization: `Bearer ${this._accessToken}` } })
    if ((res.status === 429 || res.status === 503 || res.status === 509) && attempt < 5) {
      const retryAfter = parseInt(res.headers.get('Retry-After') ?? '2', 10)
      await new Promise(r => setTimeout(r, Math.max(1, retryAfter) * 1000))
      return this._fetchWithRetry(url, attempt + 1)
    }
    if (!res.ok) {
      const body = await res.json().catch(() => ({}))
      throw new Error(body?.error?.message || `Microsoft Graph API: ${res.status}`)
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
   * Import a list of selected OneDrive files into Kagibi.
   *
   * @param {Array}    selectedFiles  normalized file items (no folders) — see listAllItems
   * @param {Map}      pathMap        oneDriveId → kagibi path (from buildPathMap)
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

  // Regular binary file: streams from OneDrive → encrypts chunk by chunk → uploads to S3.
  // No export step exists for OneDrive (unlike Google Workspace docs) — every file is
  // already a plain binary, so this is the only import path.
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

    // Graph responds with a 302 to a pre-authenticated, time-limited download URL.
    // fetch() follows redirects by default; the Authorization header is dropped by the
    // spec on cross-origin redirects, which is exactly what the storage URL expects.
    const res = await fetch(`${GRAPH_API}/me/drive/items/${item.id}/content`, {
      headers: { Authorization: `Bearer ${this._accessToken}` }
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
  return items.filter(item => !item.isFolder && !UNSUPPORTED_PACKAGE_TYPES.has(item.packageType))
}

export function getFolders(items) {
  return items.filter(item => item.isFolder)
}

export { formatBytes }
