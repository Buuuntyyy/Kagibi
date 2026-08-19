// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

/**
 * Shared helpers for third-party cloud import (OneDrive, Dropbox, ...).
 *
 * Google Drive import (googleDriveImport.js) predates this module and keeps its own
 * copies — its item shape (mimeType sentinel for folders, parents as an array) differs
 * enough that forcing it through this generic id/parentId model isn't worth the risk to
 * an already-shipped feature. OneDrive and Dropbox both use the same
 * {id, parentId, name, isFolder} shape, so their provider-agnostic pieces live here.
 */

import api from '../api'
import { NONCE_LENGTH, TAG_LENGTH_BYTES, generateBaseNonce, encryptChunkWorker, generateMasterKey, wrapMasterKey } from './crypto'
import { PART_SIZE } from './multipartUpload'

// Max folders per /folders/batch-create request (mirrors the backend's binding tag).
const FOLDER_BATCH_SIZE = 500

// Kagibi backend folder name validation: /^[\p{L}\p{N}\s\-\._‘’’]+$/u
// Characters outside this set are replaced by underscore.
// "." and ".." are reserved path names rejected by the backend even after sanitization.
export function sanitizeName(name) {
  const s = name.replace(/[^\p{L}\p{N}\s\-._‘’']/gu, "_").trim() || "Import"
  return (s === "." || s === "..") ? "Import" : s
}

export function formatBytes(bytes) {
  if (!bytes || bytes === 0) return '—'
  if (bytes < 1024) return `${bytes} o`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} Ko`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} Mo`
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} Go`
}

// Returns total encrypted byte count for a given plaintext size and chunk size.
// AES-256-GCM adds (NONCE_LENGTH + TAG_LENGTH_BYTES) = 28 bytes per chunk.
// Empty files produce exactly one chunk of 28 bytes (nonce + auth tag, no ciphertext).
export function calcEncryptedSize(plainSize, chunkSize = PART_SIZE) {
  if (plainSize === 0) return NONCE_LENGTH + TAG_LENGTH_BYTES
  const numChunks = Math.ceil(plainSize / chunkSize)
  return plainSize + numChunks * (NONCE_LENGTH + TAG_LENGTH_BYTES)
}

export function isDescendantOf(path, parentSet) {
  for (const p of parentSet) {
    if (path.startsWith(p + "/")) return true
  }
  return false
}

export function getAncestorValue(path, map) {
  for (const [k, v] of map) {
    if (path.startsWith(k + "/")) return v
  }
  return ""
}

// Builds a Map<id, sanitizedKagibiPath> from a flat folder list where each folder has
// {id, parentId, name}. File paths are resolved later via resolveItemPath().
// Uses iterative BFS to avoid call-stack overflow on deep hierarchies.
export async function buildIdPathMap(folders) {
  const byId = new Map(folders.map(f => [f.id, f]))
  const pathCache = new Map()

  const children = new Map()
  for (const folder of folders) {
    const parentId = folder.parentId
    if (parentId && byId.has(parentId)) {
      if (!children.has(parentId)) children.set(parentId, [])
      children.get(parentId).push(folder.id)
    }
  }

  const queue = []
  for (const folder of folders) {
    const parentId = folder.parentId
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

// Creates every folder in sortedLogicalPaths (shallow-to-deep, e.g. from a sort by
// segment count) via the batched /folders/batch-create endpoint instead of one request
// per folder. Conflicts (folder already exists) are always a silent no-op — Kagibi never
// asks the user to choose merge/skip, so a folder that already exists is simply reused.
// Also generates + wraps a folder key for every newly-created folder, so imported folders
// are readable in the file browser afterwards (previously only the desktop importer did this).
// Returns { invalidPaths: Map<fullPath, errorMessage> } for folders the backend rejected
// (bad name after sanitization, etc.) — callers skip files destined for those paths.
export async function ensureFoldersBatch(sortedLogicalPaths, targetPath, masterKey, onFolderProgress) {
  const invalidPaths = new Map()
  const total = sortedLogicalPaths.length
  onFolderProgress?.(0, total)
  if (total === 0) return { invalidPaths }

  let done = 0
  for (let i = 0; i < sortedLogicalPaths.length; i += FOLDER_BATCH_SIZE) {
    const chunk = sortedLogicalPaths.slice(i, i + FOLDER_BATCH_SIZE)
    const items = []
    for (const logicalPath of chunk) {
      const segments = logicalPath.split('/').filter(Boolean)
      const name = segments[segments.length - 1]
      const parentSegments = segments.slice(0, -1)
      const parent = parentSegments.length > 0
        ? (targetPath === '/' ? '/' + parentSegments.join('/') : targetPath + '/' + parentSegments.join('/'))
        : targetPath
      const fullPath = targetPath === '/' ? logicalPath : targetPath + logicalPath
      const folderKey = await generateMasterKey()
      const encryptedKey = await wrapMasterKey(folderKey, masterKey)
      items.push({ name, path: parent, encryptedKey, fullPath })
    }

    const res = await api.post('/folders/batch-create', {
      folders: items.map(it => ({ name: it.name, path: it.path, encrypted_key: it.encryptedKey })),
    })

    res.data.results.forEach((r, idx) => {
      if (r.status === 'invalid') {
        invalidPaths.set(items[idx].fullPath, r.error || 'Nom de dossier invalide')
      }
    })

    done += chunk.length
    onFolderProgress?.(done, total)
  }

  return { invalidPaths }
}

// Resolves an item's Kagibi destination directory from its parentId.
export function resolveItemPath(item, pathMap, targetPath) {
  const parentId = item.parentId
  const folderPath = parentId && pathMap.has(parentId) ? pathMap.get(parentId) : ''
  const kagibiDir = folderPath
    ? (targetPath === '/' ? folderPath : targetPath + folderPath)
    : targetPath
  return { kagibiDir, fileName: sanitizeName(item.name) }
}

// ─── PKCE helpers (OneDrive/Dropbox use hand-rolled PKCE; Google uses Google Identity
// Services' popup token client instead, so it has no PKCE code of its own) ──

export function generateRandomString(length = 64) {
  const array = new Uint8Array(length)
  crypto.getRandomValues(array)
  return Array.from(array, b => ('0' + b.toString(16)).slice(-2)).join('').slice(0, length)
}

export function base64UrlEncode(buffer) {
  const bytes = new Uint8Array(buffer)
  let binary = ''
  for (const b of bytes) binary += String.fromCharCode(b)
  return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
}

export async function generateCodeChallenge(verifier) {
  const digest = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(verifier))
  return base64UrlEncode(digest)
}

// ─── Streaming encrypt (OneDrive/Dropbox both stream a fetch Response through this;
// Google buffers Workspace exports as a Blob instead, so it keeps its own variant) ──

// Async generator: buffers a fetch Response stream into chunkSize chunks, encrypts each
// chunk, and yields the resulting Blobs one at a time. Peak RAM ≈ 2 × chunkSize regardless
// of file size. isAborted() is polled between chunks so an in-flight import can be stopped.
export async function* encryptStream(response, fileKey, chunkSize = PART_SIZE, isAborted = () => false) {
  const baseNonce = generateBaseNonce()
  let chunkIdx = 0
  for await (const chunkBuffer of bufferStream(response, chunkSize)) {
    if (isAborted()) throw new Error('Import annulé')
    yield await encryptChunkWorker(chunkBuffer, fileKey, chunkIdx++, baseNonce)
  }
}

// Reads a fetch Response as a stream and yields ArrayBuffers aligned to chunkSize.
export async function* bufferStream(response, chunkSize = PART_SIZE) {
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
