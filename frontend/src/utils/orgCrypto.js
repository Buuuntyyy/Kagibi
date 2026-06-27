// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// Organisation E2E encryption — same standards as individual user crypto.
//
// Key hierarchy:
//   OrgKey (AES-256-GCM, one per org)
//     └─ Encrypted with each member's RSA-4096 public key (RSA-OAEP SHA-256)
//        → stored in OrgMember.encrypted_org_key
//   FileKey (AES-256-GCM, one per file)
//     └─ Wrapped with OrgKey (AES-GCM key-wrap, same as wrapMasterKey)
//        → stored in OrgFile.encrypted_key
//   File content
//     └─ Encrypted with FileKey, chunked at 10 MB
//        Chunk format: [Nonce 12B] + [Ciphertext] + [GCM Tag 16B]
//        Nonce: [8B random base] + [4B chunk counter] — NIST SP 800-38D
//
// This mirrors the individual-user encryption model exactly.
// No new primitives — reuses existing NIST-compliant functions from crypto.js.

import {
  generateMasterKey,
  wrapMasterKey,
  unwrapMasterKey,
  encryptKeyWithPublicKey,
  decryptKeyWithPrivateKey,
  importKeyFromPEM,
  encryptChunkWorker,
  decryptChunkedFileWorker,
  generateBaseNonce,
  CHUNK_SIZE,
  ENCRYPTED_CHUNK_SIZE,
  encryptFileName,
  decryptFileName,
} from './crypto.js'

export { CHUNK_SIZE, ENCRYPTED_CHUNK_SIZE, decryptChunkedFileWorker }

// ── Org key generation ────────────────────────────────────────────────────────

/**
 * Generate a new organisation key.
 * AES-256-GCM, extractable (needed for RSA wrapping).
 */
export function generateOrgKey() {
  return generateMasterKey()
}

// ── Org key distribution (RSA-OAEP) ──────────────────────────────────────────

/**
 * Encrypt the org key for a specific member using their RSA-4096 public key.
 * Output: base64-encoded RSA-OAEP ciphertext — stored in OrgMember.encrypted_org_key.
 *
 * @param {CryptoKey} orgKey  AES-256-GCM org key
 * @param {string} publicKeyPEM  RSA-4096 public key in PEM/SPKI format
 * @returns {Promise<string>} base64-encoded encrypted key
 */
export async function encryptOrgKeyForUser(orgKey, publicKeyOrPEM) {
  const rawKey = await crypto.subtle.exportKey('raw', orgKey)
  const rsaPublicKey = typeof publicKeyOrPEM === 'string'
    ? await importKeyFromPEM(publicKeyOrPEM, 'spki')
    : publicKeyOrPEM
  return encryptKeyWithPublicKey(rawKey, rsaPublicKey)
}

/**
 * Decrypt the org key using own RSA-4096 private key.
 * Input: base64-encoded RSA-OAEP ciphertext from OrgMember.encrypted_org_key.
 *
 * @param {string} encryptedOrgKeyB64  base64-encoded RSA-OAEP ciphertext
 * @param {CryptoKey} rsaPrivateKey  own RSA-4096 private key
 * @returns {Promise<CryptoKey>} AES-256-GCM org key (extractable — required for member provisioning and key rotation)
 */
export async function decryptOrgKey(encryptedOrgKeyB64, rsaPrivateKey) {
  const rawKey = await decryptKeyWithPrivateKey(encryptedOrgKeyB64, rsaPrivateKey)
  return crypto.subtle.importKey('raw', rawKey, { name: 'AES-GCM' }, true, ['encrypt', 'decrypt'])
}

// ── Folder / file name encryption (AES-256-GCM) ──────────────────────────────

/**
 * Encrypt a folder or file name with the OrgKey.
 * Output: base64url (no padding) — safe as a path segment and S3 key component.
 *
 * @param {string} name  Plaintext folder or file name
 * @param {CryptoKey} orgKey  AES-256-GCM org key
 * @returns {Promise<string>} Base64url-encoded encrypted name
 */
export function encryptOrgName(name, orgKey) {
  return encryptFileName(name, orgKey)
}

/**
 * Decrypt a folder or file name with the OrgKey.
 * Falls back to the raw input if decryption fails (e.g., legacy plaintext name).
 *
 * @param {string} encryptedName  Base64url-encoded encrypted name (or plaintext)
 * @param {CryptoKey} orgKey  AES-256-GCM org key
 * @returns {Promise<string>} Decrypted name, or the input unchanged on error
 */
export function decryptOrgName(encryptedName, orgKey) {
  return decryptFileName(encryptedName, orgKey)
}

// ── File key wrapping (AES-GCM) ───────────────────────────────────────────────

/**
 * Wrap a file key with the org key.
 * Format: base64( [Nonce 12B] + [AES-GCM(fileKey)] )
 * Stored in OrgFile.encrypted_key.
 */
export function wrapFileKey(fileKey, orgKey) {
  return wrapMasterKey(fileKey, orgKey)
}

/**
 * Unwrap a file key with the org key.
 */
export function unwrapFileKey(wrappedKeyB64, orgKey) {
  return unwrapMasterKey(wrappedKeyB64, orgKey)
}

// ── File encryption / decryption ──────────────────────────────────────────────

/**
 * Encrypt a File for org storage.
 *
 * Steps:
 *   1. Generate per-file AES-256-GCM key
 *   2. Encrypt all chunks (10 MB each) via Web Worker pool
 *   3. Wrap file key with org key
 *
 * @param {File|Blob} file
 * @param {CryptoKey} orgKey  AES-256-GCM org key
 * @param {(progress: number) => void} [onProgress]  0–100
 * @returns {{ encryptedChunks: Blob[], encryptedFileKey: string, totalEncryptedSize: number }}
 */
export async function encryptFileForOrg(file, orgKey, onProgress) {
  const fileKey = await generateMasterKey()
  const encryptedFileKey = await wrapFileKey(fileKey, orgKey)

  const baseNonce = generateBaseNonce()
  const encryptedChunks = []
  let totalEncryptedSize = 0
  let chunkIndex = 0
  let offset = 0
  const totalChunks = Math.ceil(file.size / CHUNK_SIZE)

  while (offset < file.size) {
    const chunkBlob = file.slice(offset, offset + CHUNK_SIZE)
    const chunkBuf = await chunkBlob.arrayBuffer()
    const encryptedChunk = await encryptChunkWorker(chunkBuf, fileKey, chunkIndex, baseNonce)
    encryptedChunks.push(encryptedChunk)
    totalEncryptedSize += encryptedChunk.size
    offset += CHUNK_SIZE
    chunkIndex++
    if (onProgress) onProgress(Math.round((chunkIndex / totalChunks) * 50))
  }

  return { encryptedChunks, encryptedFileKey, totalEncryptedSize }
}

/**
 * Decrypt a downloaded encrypted file blob.
 *
 * @param {Blob} encryptedBlob  raw encrypted bytes from S3
 * @param {string} encryptedFileKeyB64  wrapped file key from OrgFile.encrypted_key
 * @param {CryptoKey} orgKey  AES-256-GCM org key
 * @param {string} [mimeType]
 * @returns {Promise<Blob>} plaintext file
 */
export async function decryptFileFromOrg(encryptedBlob, encryptedFileKeyB64, orgKey, mimeType, compression = '') {
  const fileKey = await unwrapFileKey(encryptedFileKeyB64, orgKey)
  return decryptChunkedFileWorker(encryptedBlob, fileKey, mimeType || 'application/octet-stream', compression)
}
