// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '../api'
import { useAuthStore } from './auth'
import {
  generateOrgKey,
  encryptOrgKeyForUser,
  decryptOrgKey,
  wrapFileKey,
  unwrapFileKey,
  decryptFileFromOrg,
  encryptOrgName,
  decryptOrgName,
  CHUNK_SIZE,
} from '../utils/orgCrypto.js'
import {
  encryptChunkWorker,
  generateBaseNonce,
  NONCE_LENGTH,
  TAG_LENGTH_BYTES,
} from '../utils/crypto.js'

// CryptoKey objects must NOT enter Vue's reactive system (Proxy breaks them).
// This module-level Map is the canonical cache for decrypted org keys.
const orgKeyCache = new Map() // orgID (number) -> CryptoKey (AES-256-GCM)

export const useOrgStore = defineStore('organizations', () => {
  const orgs = ref([])
  const currentOrg = ref(null)
  const members = ref([])
  const invitations = ref([])
  const currentItems = ref({ folders: [], files: [], current_path: '/' })
  const permissions = ref([])
  const groups = ref([])
  const myGroups = ref([])
  const auditLog = ref([])
  const auditSummary = ref({})
  const orgStats = ref(null)
  // Maps encrypted folder/file name segment → decrypted display name.
  // Populated during fetchItems so breadcrumb and file-list displays stay in sync.
  const folderNameCache = ref({})
  // Search index cache: { orgID, items: OrgItemResult[] with decrypted_name populated }
  const searchCache = ref(null)
  // Org tags: { id, encrypted_name, color, name (decrypted) }[]
  const orgTags = ref([])
  // Activity feed: OrgAuditLog[] with detail_plain (decrypted for file/folder actions)
  const orgActivity = ref([])
  // Favorites: OrgFavorite[] enriched with _name and _path from cache
  const favorites = ref([])
  // Trash: TrashItem[] with _name decrypted
  const trash = ref([])
  // Org shares: OrgShareItem[] with _file_name decrypted
  const orgShares = ref([])
  const loading = ref(false)
  const error = ref(null)
  // Upload conflict dialog: null when idle, { fileName, resolve } when pending
  const orgConflictState = ref(null)

  // ── Pinned orgs (localStorage-backed) ────────────────────────────────────
  const PINNED_KEY = 'kagibi_pinned_orgs'
  const _loadPinned = () => {
    try { return new Set(JSON.parse(localStorage.getItem(PINNED_KEY) || '[]')) }
    catch { return new Set() }
  }
  const pinnedOrgIDs = ref(_loadPinned())

  function togglePin(orgID) {
    const s = new Set(pinnedOrgIDs.value)
    if (s.has(orgID)) s.delete(orgID); else s.add(orgID)
    pinnedOrgIDs.value = s
    localStorage.setItem(PINNED_KEY, JSON.stringify([...s]))
  }

  function isPinned(orgID) {
    return pinnedOrgIDs.value.has(orgID)
  }

  // ── Org key management ────────────────────────────────────────────────────

  /**
   * Retrieve the decrypted org key for orgID.
   * Checks in-memory cache first; otherwise decrypts from currentOrg.my_encrypted_org_key.
   * Throws if the key is not available (user joined via link and admin hasn't provisioned yet).
   */
  async function getOrgKey(orgID) {
    if (orgKeyCache.has(orgID)) return orgKeyCache.get(orgID)

    const encryptedOrgKey = currentOrg.value?.my_encrypted_org_key
    if (!encryptedOrgKey) {
      throw new Error('Clé org indisponible — un administrateur doit vous provisionner l\'accès.')
    }

    const authStore = useAuthStore()
    if (!authStore.privateKey) {
      throw new Error('Clé privée introuvable. Reconnectez-vous.')
    }

    const orgKey = await decryptOrgKey(encryptedOrgKey, authStore.privateKey)
    orgKeyCache.set(orgID, orgKey)
    return orgKey
  }

  // ── Orgs ──────────────────────────────────────────────────────────────────

  async function fetchOrgs() {
    loading.value = true
    error.value = null
    try {
      const { data } = await api.get('/orgs')
      orgs.value = data || []
    } catch (e) {
      error.value = e.response?.data?.error || e.message
    } finally {
      loading.value = false
    }
  }

  async function fetchOrg(orgID) {
    loading.value = true
    error.value = null
    try {
      const { data } = await api.get(`/orgs/${orgID}`)
      currentOrg.value = data
      // Pre-warm the org key cache so first upload/download is instant.
      // Silently ignore failures (user may not have a key yet).
      if (data.my_encrypted_org_key) {
        try { await getOrgKey(orgID) } catch (_) { /* will surface on first use */ }
      }
      return data
    } catch (e) {
      error.value = e.response?.data?.error || e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  /**
   * Create an org.
   * Generates an OrgKey and encrypts it with the creator's RSA public key.
   */
  async function createOrg(name, description, storageQuotaMB) {
    const authStore = useAuthStore()

    const orgKey = await generateOrgKey()
    const rsaPublicKey = authStore.publicKey || authStore.user?.public_key
    if (!rsaPublicKey) throw new Error('Clé publique manquante. Reconnectez-vous.')

    const encryptedOrgKey = await encryptOrgKeyForUser(orgKey, rsaPublicKey)

    const { data } = await api.post('/orgs', {
      name,
      description,
      storage_quota_mb: storageQuotaMB,
      encrypted_org_key: encryptedOrgKey,
    })

    // The API wraps the org inside {organization: {...}, my_role: "owner"}.
    // Flatten to the same shape as ListOrgs / GetOrg responses.
    const org = data.organization || data
    const entry = { ...org, my_role: data.my_role ?? 'owner', my_encrypted_org_key: encryptedOrgKey }

    // Cache the freshly generated key immediately — no need to re-decrypt.
    orgKeyCache.set(entry.id, orgKey)

    orgs.value.push(entry)
    return entry
  }

  async function updateOrg(orgID, payload) {
    const { data } = await api.patch(`/orgs/${orgID}`, payload)
    currentOrg.value = { ...currentOrg.value, ...data }
    const idx = orgs.value.findIndex(o => o.id === orgID)
    if (idx !== -1) orgs.value[idx] = { ...orgs.value[idx], ...data }
    return data
  }

  async function uploadOrgLogo(orgID, file) {
    const form = new FormData()
    form.append('logo', file)
    const { data } = await api.put(`/orgs/${orgID}/logo`, form, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    if (currentOrg.value?.id === orgID) currentOrg.value = { ...currentOrg.value, logo_url: data.logo_url, logo_path: data.logo_path }
    const idx = orgs.value.findIndex(o => o.id === orgID)
    if (idx !== -1) orgs.value[idx] = { ...orgs.value[idx], logo_url: data.logo_url, logo_path: data.logo_path }
    return data
  }

  async function deleteOrgLogo(orgID) {
    await api.delete(`/orgs/${orgID}/logo`)
    if (currentOrg.value?.id === orgID) currentOrg.value = { ...currentOrg.value, logo_url: '', logo_path: '' }
    const idx = orgs.value.findIndex(o => o.id === orgID)
    if (idx !== -1) orgs.value[idx] = { ...orgs.value[idx], logo_url: '', logo_path: '' }
  }

  async function deleteOrg(orgID) {
    await api.delete(`/orgs/${orgID}`)
    orgKeyCache.delete(orgID)
    orgs.value = orgs.value.filter(o => o.id !== orgID)
    if (currentOrg.value?.id === orgID) currentOrg.value = null
  }

  // ── Members ───────────────────────────────────────────────────────────────

  async function fetchMembers(orgID) {
    const { data } = await api.get(`/orgs/${orgID}/members`)
    members.value = data || []
    return data
  }

  // memberID is the OrgMember row ID (member.id), not the user UUID
  async function updateMemberRole(orgID, memberID, role) {
    await api.patch(`/orgs/${orgID}/members/${memberID}`, { role })
    const m = members.value.find(m => m.id === memberID)
    if (m) m.role = role
  }

  async function setMemberQuota(orgID, memberID, quotaBytes) {
    const m = members.value.find(m => m.id === memberID)
    const role = m?.role ?? 'member'
    const { data } = await api.patch(`/orgs/${orgID}/members/${memberID}`, { role, quota_bytes: quotaBytes })
    if (m) m.quota_bytes = data.quota_bytes ?? quotaBytes
  }

  // memberID is the OrgMember row ID (member.id), not the user UUID
  async function removeMember(orgID, memberID) {
    await api.delete(`/orgs/${orgID}/members/${memberID}`)
    members.value = members.value.filter(m => m.id !== memberID)
  }

  /**
   * Provision the org key for a member who joined via link (has no encrypted key yet).
   * Encrypts the org key with the target member's RSA public key and stores it.
   */
  async function provisionMemberKey(orgID, member) {
    if (!member.public_key) throw new Error('Clé publique du membre introuvable.')

    const orgKey = await getOrgKey(orgID)
    const encryptedOrgKey = await encryptOrgKeyForUser(orgKey, member.public_key)

    await api.patch(`/orgs/${orgID}/members/${member.id}/key`, {
      encrypted_org_key: encryptedOrgKey,
    })

    // Update local state so the UI refreshes immediately
    const m = members.value.find(m => m.id === member.id)
    if (m) m.encrypted_org_key = encryptedOrgKey
  }

  /**
   * Provision the org key for every member who is missing it and has a public key.
   * Decrypts the org key once, then wraps it for each member in sequence.
   * Returns the count of successfully provisioned members.
   */
  async function provisionAllMissingKeys(orgID) {
    const orgKey = await getOrgKey(orgID)
    const targets = members.value.filter(m => !m.encrypted_org_key && m.public_key)
    let count = 0
    for (const member of targets) {
      const encryptedOrgKey = await encryptOrgKeyForUser(orgKey, member.public_key)
      await api.patch(`/orgs/${orgID}/members/${member.id}/key`, { encrypted_org_key: encryptedOrgKey })
      const m = members.value.find(x => x.id === member.id)
      if (m) m.encrypted_org_key = encryptedOrgKey
      count++
    }
    return count
  }

  // Low-level setMemberKey (raw, for JoinView owner flow)
  async function setMemberKey(orgID, memberID, encryptedOrgKey) {
    await api.patch(`/orgs/${orgID}/members/${memberID}/key`, { encrypted_org_key: encryptedOrgKey })
  }

  // ── Invitations ───────────────────────────────────────────────────────────

  async function fetchInvitations(orgID) {
    const { data } = await api.get(`/orgs/${orgID}/invitations`)
    invitations.value = data || []
    return data
  }

  /**
   * Create an invitation.
   * For direct invites (target_user_id set), pre-encrypts the org key for the target.
   */
  async function createInvitation(orgID, payload) {
    let enrichedPayload = { ...payload }

    if (payload.target_user_id) {
      // Find target member's public key (they must already be in the system via friends or lookup)
      const target = members.value.find(m => m.user_id === payload.target_user_id)
      if (target?.public_key) {
        const orgKey = await getOrgKey(orgID)
        enrichedPayload.encrypted_org_key = await encryptOrgKeyForUser(orgKey, target.public_key)
      }
    }

    const { data } = await api.post(`/orgs/${orgID}/invitations`, enrichedPayload)
    invitations.value.unshift(data)
    return data
  }

  async function revokeInvitation(orgID, inviteID) {
    await api.delete(`/orgs/${orgID}/invitations/${inviteID}`)
    invitations.value = invitations.value.filter(i => i.id !== inviteID)
  }

  // Public — no org context needed, token carries everything
  async function getInvitation(token) {
    const { data } = await api.get(`/org-invitations/${token}`)
    return data
  }

  async function acceptInvitation(token, encryptedOrgKey = '') {
    const { data } = await api.post(`/org-invitations/${token}/accept`, {
      encrypted_org_key: encryptedOrgKey,
    })
    return data
  }

  // ── File system ───────────────────────────────────────────────────────────

  async function fetchItems(orgID, folderPath = '/') {
    loading.value = true
    try {
      const encodedPath = encodeURIComponent(folderPath)
      const { data } = await api.get(`/orgs/${orgID}/fs/list/${encodedPath}`)

      // Decrypt folder and file names with the OrgKey.
      // Fallback: if the key is unavailable or the name is plaintext (legacy),
      // decryptOrgName returns the input unchanged — no data loss.
      try {
        const orgKey = await getOrgKey(orgID)
        for (const folder of data.folders || []) {
          const plain = await decryptOrgName(folder.name, orgKey)
          folderNameCache.value[folder.name] = plain
          folder.name = plain
        }
        for (const file of data.files || []) {
          file.name = await decryptOrgName(file.name, orgKey)
        }
      } catch (_) { /* key not yet provisioned — names stay as-is */ }

      currentItems.value = {
        ...data,
        folders: data.folders || [],
        files: data.files || [],
      }
      return currentItems.value
    } catch (e) {
      error.value = e.response?.data?.error || e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  async function createFolder(orgID, name, parentPath, encryptedKey = '') {
    let apiName = name
    try {
      const orgKey = await getOrgKey(orgID)
      apiName = await encryptOrgName(name, orgKey)
      folderNameCache.value[apiName] = name
    } catch (_) { /* key unavailable — fall back to plaintext */ }

    const { data } = await api.post(`/orgs/${orgID}/fs/folder`, {
      name: apiName,
      parent_path: parentPath,
      encrypted_key: encryptedKey,
    })
    // Server echoes back the encrypted name; replace with plaintext for immediate display.
    data.name = name
    currentItems.value.folders = [...(currentItems.value.folders || []), data]
    searchCache.value = null
    return data
  }

  async function deleteFolder(orgID, folderID) {
    await api.delete(`/orgs/${orgID}/fs/folder/${folderID}`)
    currentItems.value.folders = currentItems.value.folders.filter(f => f.id !== folderID)
    searchCache.value = null
  }

  /**
   * Encrypt a file and upload it via multipart to S3.
   *
   * Flow:
   *   1. Get org key → decrypt from member record
   *   2. Generate per-file AES-256-GCM key
   *   3. Encrypt all chunks via Web Worker pool (NIST SP 800-38D nonces)
   *   4. Initiate S3 multipart upload
   *   5. PUT encrypted chunks to presigned URLs
   *   6. Complete upload with wrapped file key
   *
   * @param {number} orgID
   * @param {File} file
   * @param {string} folderPath
   * @param {(progress: number) => void} [onProgress]  0–100
   */
  function showOrgConflict(fileName) {
    return new Promise(resolve => {
      orgConflictState.value = { fileName, resolve }
    })
  }

  function resolveOrgConflict(choice) {
    if (orgConflictState.value?.resolve) {
      orgConflictState.value.resolve(choice)
      orgConflictState.value = null
    }
  }

  async function uploadOrgFile(orgID, file, folderPath, onProgress) {
    // Conflict check: is there a file with the same name in this folder?
    const conflict = (currentItems.value.files || []).some(f => f.name === file.name)
    if (conflict) {
      const choice = await showOrgConflict(file.name)
      if (choice === 'cancel') return null
      // keepBoth: find a free name
      let base = file.name, ext = ''
      const dot = file.name.lastIndexOf('.')
      if (dot > 0) { base = file.name.slice(0, dot); ext = file.name.slice(dot) }
      let n = 1
      while ((currentItems.value.files || []).some(f => f.name === `${base} (${n})${ext}`)) n++
      file = new File([file], `${base} (${n})${ext}`, { type: file.type })
    }

    const orgKey = await getOrgKey(orgID)

    let encryptedFileName = file.name
    try {
      encryptedFileName = await encryptOrgName(file.name, orgKey)
    } catch (_) { /* fall back to plaintext */ }

    // Generate a per-file key and wrap it with the org key.
    const fileKey = await generateOrgKey()
    const encryptedFileKey = await wrapFileKey(fileKey, orgKey)
    const baseNonce = generateBaseNonce()

    // Compute total encrypted size deterministically before touching S3.
    const numChunks = file.size === 0 ? 1 : Math.ceil(file.size / CHUNK_SIZE)
    const totalEncryptedSize = file.size + numChunks * (NONCE_LENGTH + TAG_LENGTH_BYTES)

    const { data: initData } = await api.post(`/orgs/${orgID}/fs/multipart/initiate`, {
      file_name: encryptedFileName,
      file_path: folderPath,
      content_type: 'application/octet-stream',
      total_size: totalEncryptedSize,
      total_parts: numChunks,
      encrypted_key: encryptedFileKey,
    })

    // Encrypt one chunk → upload → move on.  Peak memory ≈ 1 chunk (≈ 10 MB).
    const parts = []
    let offset = 0
    for (let i = 0; i < initData.presigned_urls.length; i++) {
      const { part_number, url } = initData.presigned_urls[i]
      // arrayBuffer() is transferred to the worker (detached in main thread → no copy).
      const chunkBuf = await file.slice(offset, offset + CHUNK_SIZE).arrayBuffer()
      const encryptedChunk = await encryptChunkWorker(chunkBuf, fileKey, i, baseNonce)
      offset += CHUNK_SIZE
      const res = await fetch(url, { method: 'PUT', body: encryptedChunk })
      // encryptedChunk goes out of scope here; eligible for GC immediately.
      const etag = res.headers.get('ETag') || ''
      parts.push({ part_number, etag })
      if (onProgress) onProgress(Math.round(((i + 1) / numChunks) * 100))
    }

    const { data: result } = await api.post(`/orgs/${orgID}/fs/multipart/complete`, {
      upload_id: initData.upload_id,
      key: initData.key,
      parts,
      file_name: encryptedFileName,
      file_path: folderPath,
      total_size: totalEncryptedSize,
      content_type: 'application/octet-stream',
      encrypted_key: encryptedFileKey,
    })

    if (result.file) {
      result.file.name = file.name
      currentItems.value.files = [...(currentItems.value.files || []), result.file]
      if (currentOrg.value) {
        currentOrg.value.storage_used_bytes = (currentOrg.value.storage_used_bytes || 0) + result.file.size
      }
    }
    return result
  }

  /**
   * Download and decrypt an org file.
   *
   * Flow:
   *   1. Get the file's wrapped key (from listing cache or dedicated endpoint)
   *   2. Get org key → decrypt from member record
   *   3. Unwrap file key with org key
   *   4. Download encrypted blob
   *   5. Decrypt blob via Worker pool
   *   6. Trigger browser download
   */
  async function downloadFile(orgID, fileID, fileName, mimeType = '') {
    // Prefer key from listing cache to avoid an extra round-trip
    const cachedFile = currentItems.value.files.find(f => f.id === fileID)
    const encryptedFileKey = cachedFile?.encrypted_key || (await getFileKey(orgID, fileID))

    const orgKey = await getOrgKey(orgID)

    const response = await api.get(`/orgs/${orgID}/fs/file/${fileID}/download`, {
      responseType: 'blob',
    })

    const decryptedBlob = await decryptFileFromOrg(
      response.data,
      encryptedFileKey,
      orgKey,
      mimeType || cachedFile?.mime_type || 'application/octet-stream',
    )

    const url = URL.createObjectURL(decryptedBlob)
    const a = document.createElement('a')
    a.href = url
    a.download = fileName
    a.click()
    URL.revokeObjectURL(url)
  }

  async function deleteFile(orgID, fileID) {
    const file = currentItems.value.files.find(f => f.id === fileID)
    await api.delete(`/orgs/${orgID}/fs/file/${fileID}`)
    currentItems.value.files = currentItems.value.files.filter(f => f.id !== fileID)
    if (currentOrg.value && file) {
      currentOrg.value.storage_used_bytes = Math.max(
        0,
        (currentOrg.value.storage_used_bytes || 0) - file.size,
      )
    }
    searchCache.value = null
  }

  async function getFileKey(orgID, fileID) {
    const { data } = await api.get(`/orgs/${orgID}/fs/file/${fileID}/key`)
    return data.encrypted_key
  }

  async function getFileBlob(orgID, fileID, mimeType) {
    const cachedFile = currentItems.value.files.find(f => f.id === fileID)
    const encryptedFileKey = cachedFile?.encrypted_key || (await getFileKey(orgID, fileID))
    const orgKey = await getOrgKey(orgID)
    const response = await api.get(`/orgs/${orgID}/fs/file/${fileID}/download`, { responseType: 'blob' })
    return decryptFileFromOrg(
      response.data,
      encryptedFileKey,
      orgKey,
      mimeType || cachedFile?.mime_type || 'application/octet-stream',
    )
  }

  // Low-level initiate/complete/abort kept for any direct caller
  async function initiateUpload(orgID, payload) {
    const { data } = await api.post(`/orgs/${orgID}/fs/multipart/initiate`, payload)
    return data
  }

  async function completeUpload(orgID, payload) {
    const { data } = await api.post(`/orgs/${orgID}/fs/multipart/complete`, payload)
    if (data.file) {
      currentItems.value.files = [...(currentItems.value.files || []), data.file]
      if (currentOrg.value) {
        currentOrg.value.storage_used_bytes =
          (currentOrg.value.storage_used_bytes || 0) + data.file.size
      }
    }
    searchCache.value = null
    return data
  }

  async function abortUpload(orgID, uploadID, key) {
    await api.post(`/orgs/${orgID}/fs/multipart/abort`, { upload_id: uploadID, key })
  }

  // ── Search ────────────────────────────────────────────────────────────────

  /**
   * Full-text search over all org files and folders.
   * Fetches the full item list once (per orgID) and decrypts names client-side;
   * subsequent calls reuse the cache until a mutation invalidates it.
   */
  async function searchOrgItems(orgID, query) {
    if (!searchCache.value || searchCache.value.orgID !== orgID) {
      const { data } = await api.get(`/orgs/${orgID}/fs/all-items`)
      let items = data || []
      try {
        const orgKey = await getOrgKey(orgID)
        for (const item of items) {
          const plain = await decryptOrgName(item.name, orgKey)
          item.decrypted_name = plain
          if (item.type === 'folder') folderNameCache.value[item.name] = plain

          // Decrypt path segments for display in search results
          const decryptSegs = async (p) => {
            if (!p || p === '/') return p
            const segs = p.split('/')
            const dec = await Promise.all(segs.map(s => s ? decryptOrgName(s, orgKey).catch(() => s) : Promise.resolve(s)))
            return dec.join('/')
          }
          item.decrypted_path = await decryptSegs(item.path)
          item.decrypted_parent_path = await decryptSegs(item.parent_path)
        }
      } catch (_) {
        for (const item of items) {
          item.decrypted_name = item.name
          item.decrypted_path = item.path
          item.decrypted_parent_path = item.parent_path
        }
      }
      searchCache.value = { orgID, items }
    }

    const q = query.trim().toLowerCase()
    if (!q) return []
    return searchCache.value.items.filter(item =>
      item.decrypted_name.toLowerCase().includes(q)
    )
  }

  // ── Groups ────────────────────────────────────────────────────────────────

  async function fetchGroups(orgID) {
    const { data } = await api.get(`/orgs/${orgID}/groups`)
    groups.value = data || []
    return data
  }

  async function fetchMyGroups(orgID) {
    const { data } = await api.get(`/orgs/${orgID}/groups/me`)
    myGroups.value = data || []
    return myGroups.value
  }

  async function createGroup(orgID, name, description) {
    const { data } = await api.post(`/orgs/${orgID}/groups`, { name, description })
    groups.value.push(data)
    return data
  }

  async function updateGroup(orgID, groupID, payload) {
    const { data } = await api.patch(`/orgs/${orgID}/groups/${groupID}`, payload)
    const idx = groups.value.findIndex(g => g.id === groupID)
    if (idx !== -1) groups.value[idx] = { ...groups.value[idx], ...data }
    return data
  }

  async function deleteGroup(orgID, groupID) {
    await api.delete(`/orgs/${orgID}/groups/${groupID}`)
    groups.value = groups.value.filter(g => g.id !== groupID)
  }

  async function addGroupMember(orgID, groupID, userID, role = 'member') {
    const { data } = await api.post(`/orgs/${orgID}/groups/${groupID}/members`, { user_id: userID, role })
    return data
  }

  async function removeGroupMember(orgID, groupID, memberID) {
    await api.delete(`/orgs/${orgID}/groups/${groupID}/members/${memberID}`)
  }

  async function updateGroupMemberRole(orgID, groupID, memberID, role) {
    const { data } = await api.patch(`/orgs/${orgID}/groups/${groupID}/members/${memberID}`, { role })
    return data
  }

  async function fetchGroupMembers(orgID, groupID) {
    const { data } = await api.get(`/orgs/${orgID}/groups/${groupID}/members`)
    return data || []
  }

  async function setGroupPermission(orgID, groupID, payload) {
    const { data } = await api.put(`/orgs/${orgID}/groups/${groupID}/permissions`, payload)
    return data
  }

  async function deleteGroupPermission(orgID, groupID, folderPath) {
    await api.delete(`/orgs/${orgID}/groups/${groupID}/permissions`, {
      data: { folder_path: folderPath },
    })
  }

  async function fetchGroupPermissions(orgID, groupID) {
    const { data } = await api.get(`/orgs/${orgID}/groups/${groupID}/permissions`)
    return data || []
  }

  // ── Permissions ───────────────────────────────────────────────────────────

  async function fetchFolderAccess(orgID, folderPath) {
    const { data } = await api.get(`/orgs/${orgID}/permissions/folder`, { params: { path: folderPath } })
    return data // { users: [...], groups: [{group, permission}, ...] }
  }

  async function fetchPermissions(orgID) {
    const { data } = await api.get(`/orgs/${orgID}/permissions`)
    permissions.value = data || []
    return data
  }

  async function setPermission(orgID, payload) {
    const { data } = await api.put(`/orgs/${orgID}/permissions`, payload)
    const idx = permissions.value.findIndex(
      p => p.user_id === payload.user_id && p.folder_path === (payload.folder_path || '/'),
    )
    if (idx !== -1) permissions.value[idx] = data
    else permissions.value.push(data)
    return data
  }

  async function deletePermission(orgID, userID, folderPath) {
    await api.delete(`/orgs/${orgID}/permissions`, {
      data: { user_id: userID, folder_path: folderPath },
    })
    permissions.value = permissions.value.filter(
      p => !(p.user_id === userID && p.folder_path === folderPath),
    )
  }

  // ── Audit log ─────────────────────────────────────────────────────────────

  // Scans a free-text detail string for encrypted path segments (/base64url)
  // and replaces each with its decrypted name. Used for permission/group actions
  // where the detail is a sentence containing embedded encrypted paths.
  async function decryptPathSegmentsInText(text, orgKey) {
    if (!text) return text
    const regex = /\/([A-Za-z0-9_\-]{20,})/g
    const segments = new Map()
    let match
    while ((match = regex.exec(text)) !== null) {
      const seg = match[1]
      if (!segments.has(seg)) {
        try { segments.set(seg, await decryptOrgName(seg, orgKey)) }
        catch (_) { segments.set(seg, seg) }
      }
    }
    if (segments.size === 0) return text
    let result = text
    for (const [enc, plain] of segments) result = result.replaceAll(`/${enc}`, `/${plain}`)
    return result
  }

  async function fetchAuditLog(orgID, page = 1) {
    const { data } = await api.get(`/orgs/${orgID}/audit`, { params: { page } })
    const entries = data || []
    try {
      const orgKey = await getOrgKey(orgID)
      for (const entry of entries) {
        if (FILE_FOLDER_AUDIT_ACTIONS.has(entry.action) && entry.detail) {
          try { entry.detail_plain = await decryptOrgName(entry.detail, orgKey) }
          catch (_) { entry.detail_plain = await decryptPathSegmentsInText(entry.detail, orgKey) }
        } else {
          entry.detail_plain = await decryptPathSegmentsInText(entry.detail, orgKey)
        }
      }
    } catch (_) {
      for (const entry of entries) entry.detail_plain = entry.detail
    }
    if (page === 1) {
      auditLog.value = entries
    } else {
      auditLog.value = [...auditLog.value, ...entries]
    }
    return entries
  }

  async function fetchAuditSummary(orgID) {
    const { data } = await api.get(`/orgs/${orgID}/audit/summary`)
    auditSummary.value = data?.days || {}
  }

  async function deleteAuditLog(orgID, payload) {
    const { data } = await api.delete(`/orgs/${orgID}/audit`, { data: payload })
    return data
  }

  async function exportAuditLog(orgID) {
    const response = await api.get(`/orgs/${orgID}/audit/export`, { responseType: 'blob' })
    const url = URL.createObjectURL(response.data)
    const a = document.createElement('a')
    a.href = url
    const today = new Date().toISOString().slice(0, 10)
    a.download = `audit-org${orgID}-${today}.csv`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }

  async function moveOrgFile(orgID, fileID, newFolderPath) {
    const { data } = await api.patch(`/orgs/${orgID}/fs/file/${fileID}/move`, {
      new_folder_path: newFolderPath,
    })
    currentItems.value.files = currentItems.value.files.filter(f => f.id !== fileID)
    searchCache.value = null
    return data
  }

  async function moveOrgFolder(orgID, folderID, newParentPath) {
    const { data } = await api.patch(`/orgs/${orgID}/fs/folder/${folderID}/move`, {
      new_parent_path: newParentPath,
    })
    currentItems.value.folders = currentItems.value.folders.filter(f => f.id !== folderID)
    searchCache.value = null
    return data
  }

  async function getAllOrgFolders(orgID) {
    if (!searchCache.value || searchCache.value.orgID !== orgID) {
      const { data } = await api.get(`/orgs/${orgID}/fs/all-items`)
      const items = data || []
      try {
        const orgKey = await getOrgKey(orgID)
        for (const item of items) {
          const plain = await decryptOrgName(item.name, orgKey)
          item.decrypted_name = plain
          if (item.type === 'folder') folderNameCache.value[item.name] = plain
        }
      } catch (_) {
        for (const item of items) item.decrypted_name = item.name
      }
      searchCache.value = { orgID, items }
    }
    return (searchCache.value.items || [])
      .filter(item => item.type === 'folder')
      .map(item => ({ id: item.id, name: item.decrypted_name, path: item.path, parent_path: item.parent_path }))
  }

  async function fetchOrgTags(orgID) {
    const { data } = await api.get(`/orgs/${orgID}/tags`)
    const tags = data || []
    try {
      const orgKey = await getOrgKey(orgID)
      for (const tag of tags) tag.name = await decryptOrgName(tag.encrypted_name, orgKey)
    } catch (_) {
      for (const tag of tags) tag.name = tag.encrypted_name
    }
    orgTags.value = tags
    return tags
  }

  async function createOrgTag(orgID, plainName, color) {
    const orgKey = await getOrgKey(orgID)
    const encryptedName = await encryptOrgName(plainName, orgKey)
    const { data } = await api.post(`/orgs/${orgID}/tags`, { encrypted_name: encryptedName, color })
    const tag = { ...data, name: plainName }
    orgTags.value = [...orgTags.value, tag]
    return tag
  }

  async function updateOrgTag(orgID, tagID, plainName, color) {
    const orgKey = await getOrgKey(orgID)
    const body = {}
    if (plainName) body.encrypted_name = await encryptOrgName(plainName, orgKey)
    if (color) body.color = color
    const { data } = await api.patch(`/orgs/${orgID}/tags/${tagID}`, body)
    orgTags.value = orgTags.value.map(t =>
      t.id === tagID ? { ...data, name: plainName || t.name } : t
    )
    return data
  }

  async function deleteOrgTag(orgID, tagID) {
    await api.delete(`/orgs/${orgID}/tags/${tagID}`)
    orgTags.value = orgTags.value.filter(t => t.id !== tagID)
    currentItems.value.files = currentItems.value.files.map(f => ({
      ...f, tag_ids: (f.tag_ids || []).filter(id => id !== tagID),
    }))
    currentItems.value.folders = currentItems.value.folders.map(f => ({
      ...f, tag_ids: (f.tag_ids || []).filter(id => id !== tagID),
    }))
  }

  async function setFileTags(orgID, fileID, tagIDs) {
    await api.put(`/orgs/${orgID}/fs/file/${fileID}/tags`, { tag_ids: tagIDs })
    const file = currentItems.value.files.find(f => f.id === fileID)
    if (file) file.tag_ids = tagIDs
  }

  async function setFolderTags(orgID, folderID, tagIDs) {
    await api.put(`/orgs/${orgID}/fs/folder/${folderID}/tags`, { tag_ids: tagIDs })
    const folder = currentItems.value.folders.find(f => f.id === folderID)
    if (folder) folder.tag_ids = tagIDs
  }

  const FILE_FOLDER_AUDIT_ACTIONS = new Set([
    'file_uploaded', 'file_deleted', 'file_downloaded', 'file_renamed',
    'file_moved', 'file_shared_public',
    'folder_created', 'folder_deleted', 'folder_renamed', 'folder_moved',
  ])

  async function fetchOrgActivity(orgID) {
    const { data } = await api.get(`/orgs/${orgID}/activity`)
    const entries = data || []
    try {
      const orgKey = await getOrgKey(orgID)
      for (const entry of entries) {
        if (FILE_FOLDER_AUDIT_ACTIONS.has(entry.action) && entry.detail) {
          try { entry.detail_plain = await decryptOrgName(entry.detail, orgKey) }
          catch (_) { entry.detail_plain = await decryptPathSegmentsInText(entry.detail, orgKey) }
        } else {
          entry.detail_plain = await decryptPathSegmentsInText(entry.detail, orgKey)
        }
      }
    } catch (_) {
      for (const entry of entries) entry.detail_plain = entry.detail
    }
    orgActivity.value = entries
    return entries
  }

  async function fetchFavorites(orgID) {
    // Warm search cache so we can resolve encrypted names
    if (!searchCache.value || searchCache.value.orgID !== orgID) {
      try {
        const { data } = await api.get(`/orgs/${orgID}/fs/all-items`)
        const items = data || []
        try {
          const orgKey = await getOrgKey(orgID)
          for (const item of items) {
            const plain = await decryptOrgName(item.name, orgKey)
            item.decrypted_name = plain
            if (item.type === 'folder') folderNameCache.value[item.name] = plain
          }
        } catch (_) {
          for (const item of items) item.decrypted_name = item.name
        }
        searchCache.value = { orgID, items }
      } catch (_) {}
    }

    const { data } = await api.get(`/orgs/${orgID}/favorites`)
    const favs = data || []

    if (searchCache.value?.orgID === orgID) {
      for (const fav of favs) {
        const found = searchCache.value.items.find(i => i.id === fav.item_id && i.type === fav.item_type)
        if (found) {
          fav._name = found.decrypted_name
          fav._path = found.path
          fav._parent_path = found.parent_path
        }
      }
    }

    favorites.value = favs
    return favs
  }

  async function addFavorite(orgID, itemID, itemType) {
    const { data } = await api.post(`/orgs/${orgID}/favorites`, { item_id: itemID, item_type: itemType })
    if (data.id) {
      if (searchCache.value?.orgID === orgID) {
        const found = searchCache.value.items.find(i => i.id === itemID && i.type === itemType)
        if (found) { data._name = found.decrypted_name; data._path = found.path; data._parent_path = found.parent_path }
      }
      if (!favorites.value.find(f => f.item_id === itemID && f.item_type === itemType)) {
        favorites.value.push(data)
      }
    }
    return data
  }

  async function removeFavorite(orgID, itemID, itemType) {
    await api.delete(`/orgs/${orgID}/favorites/${itemType}/${itemID}`)
    favorites.value = favorites.value.filter(f => !(f.item_id === itemID && f.item_type === itemType))
  }

  async function fetchTrash(orgID) {
    const { data } = await api.get(`/orgs/${orgID}/trash`)
    const items = data || []
    try {
      const orgKey = await getOrgKey(orgID)
      for (const item of items) {
        item._name = await decryptOrgName(item.name, orgKey)
        if (item.path) {
          const segs = item.path.split('/')
          const plain = await Promise.all(
            segs.map(s => s ? decryptOrgName(s, orgKey).catch(() => s) : Promise.resolve(s))
          )
          item._path = plain.join('/')
        } else {
          item._path = item.path
        }
      }
    } catch (_) {
      for (const item of items) { item._name = item.name; item._path = item.path }
    }
    trash.value = items
    return items
  }

  async function restoreTrashItem(orgID, itemType, itemID) {
    await api.post(`/orgs/${orgID}/trash/${itemType}/${itemID}/restore`)
    trash.value = trash.value.filter(i => !(i.id === itemID && i.item_type === itemType))
    searchCache.value = null
  }

  async function permanentDeleteTrashItem(orgID, itemType, itemID) {
    await api.delete(`/orgs/${orgID}/trash/${itemType}/${itemID}`)
    trash.value = trash.value.filter(i => !(i.id === itemID && i.item_type === itemType))
  }

  async function emptyTrash(orgID) {
    await api.delete(`/orgs/${orgID}/trash`)
    trash.value = []
  }

  async function fetchOrgShares(orgID) {
    const { data } = await api.get(`/orgs/${orgID}/shares`)
    const items = data || []
    try {
      const orgKey = await getOrgKey(orgID)
      for (const item of items) item._file_name = await decryptOrgName(item.file_name, orgKey)
    } catch (_) {
      for (const item of items) item._file_name = item.file_name
    }
    orgShares.value = items
    return items
  }

  async function revokeOrgShare(orgID, shareID) {
    await api.delete(`/orgs/${orgID}/shares/${shareID}`)
    orgShares.value = orgShares.value.filter(s => s.id !== shareID)
  }

  // Internal: collect { zipPath: Uint8Array } entries from file items + folder subtrees.
  async function _collectZipEntries(orgID, fileItems, folderInfos, onProgress) {
    if (!searchCache.value || searchCache.value.orgID !== orgID) {
      const { data } = await api.get(`/orgs/${orgID}/fs/all-items`)
      const items = data || []
      try {
        const orgKey = await getOrgKey(orgID)
        for (const item of items) {
          const plain = await decryptOrgName(item.name, orgKey)
          item.decrypted_name = plain
          if (item.type === 'folder') folderNameCache.value[item.name] = plain
        }
      } catch (_) {
        for (const item of items) item.decrypted_name = item.name
      }
      searchCache.value = { orgID, items }
    }

    const seenIDs = new Set()
    const zipList = []

    for (const f of fileItems) {
      if (seenIDs.has(f.id)) continue
      seenIDs.add(f.id)
      zipList.push({ id: f.id, mime_type: f.mime_type, zipPath: f.name })
    }

    for (const fi of folderInfos) {
      const prefix = fi.path === '/' ? '/' : fi.path + '/'
      const matched = (searchCache.value.items || []).filter(item => {
        if (item.type !== 'file') return false
        return fi.path === '/' ? true : (item.path === fi.path || item.path.startsWith(prefix))
      })
      for (const item of matched) {
        if (seenIDs.has(item.id)) continue
        seenIDs.add(item.id)
        const relDir = item.path.slice(prefix.length)
        const zipPath = relDir
          ? `${fi.name}/${relDir}/${item.decrypted_name}`
          : `${fi.name}/${item.decrypted_name}`
        zipList.push({ id: item.id, mime_type: item.mime_type, zipPath })
      }
    }

    if (zipList.length === 0) return {}

    const allKeys = await fetchAllFileKeys(orgID)
    const keyMap = {}
    for (const k of allKeys) if (k.encrypted_key) keyMap[k.id] = k.encrypted_key

    const orgKey = await getOrgKey(orgID)
    const entries = {}

    for (let i = 0; i < zipList.length; i++) {
      const { id, mime_type, zipPath } = zipList[i]
      const encKey = keyMap[id] || (await getFileKey(orgID, id))
      const res = await api.get(`/orgs/${orgID}/fs/file/${id}/download`, { responseType: 'blob' })
      const blob = await decryptFileFromOrg(res.data, encKey, orgKey, mime_type || 'application/octet-stream')
      entries[zipPath] = new Uint8Array(await blob.arrayBuffer())
      if (onProgress) onProgress(i + 1, zipList.length)
    }

    return entries
  }

  async function downloadFolderAsZip(orgID, folderPath, folderName, onProgress) {
    const { zipSync } = await import('fflate')
    const entries = await _collectZipEntries(orgID, [], [{ path: folderPath, name: folderName }], onProgress)
    if (Object.keys(entries).length === 0) return 0
    const zip = zipSync(entries)
    const url = URL.createObjectURL(new Blob([zip], { type: 'application/zip' }))
    const a = document.createElement('a')
    a.href = url; a.download = `${folderName}.zip`; a.click()
    URL.revokeObjectURL(url)
    return Object.keys(entries).length
  }

  async function downloadSelectionAsZip(orgID, fileItems, folderItems, onProgress) {
    const { zipSync } = await import('fflate')
    const folderInfos = folderItems.map(f => ({ path: f.path, name: f.name }))
    const entries = await _collectZipEntries(orgID, fileItems, folderInfos, onProgress)
    if (Object.keys(entries).length === 0) return 0
    const zip = zipSync(entries)
    const url = URL.createObjectURL(new Blob([zip], { type: 'application/zip' }))
    const a = document.createElement('a')
    a.href = url; a.download = 'selection.zip'; a.click()
    URL.revokeObjectURL(url)
    return Object.keys(entries).length
  }

  async function renameOrgFile(orgID, fileID, newPlainName) {
    const orgKey = await getOrgKey(orgID)
    const encryptedName = await encryptOrgName(newPlainName, orgKey)
    const { data } = await api.patch(`/orgs/${orgID}/fs/file/${fileID}/rename`, {
      encrypted_name: encryptedName,
    })
    const file = currentItems.value.files.find(f => f.id === fileID)
    if (file) {
      file.name = newPlainName
      file.path = data.new_path
    }
    searchCache.value = null
    return data
  }

  async function renameOrgFolder(orgID, folderID, newPlainName) {
    const orgKey = await getOrgKey(orgID)
    const encryptedName = await encryptOrgName(newPlainName, orgKey)
    const { data } = await api.patch(`/orgs/${orgID}/fs/folder/${folderID}/rename`, {
      encrypted_name: encryptedName,
    })
    const folder = currentItems.value.folders.find(f => f.id === folderID)
    if (folder) {
      folderNameCache.value[encryptedName] = newPlainName
      folder.name = newPlainName
      folder.path = data.new_path
    }
    searchCache.value = null
    return data
  }

  /**
   * Create a public share link for an org file.
   * The caller must have already:
   *   1. Unwrapped the file key with the org key
   *   2. Re-wrapped it with a fresh share key
   *   3. Encoded the share key as a URL-safe base64 string (used as the #fragment)
   *
   * Returns { token, link } from the server.
   */
  async function createOrgFileShare(orgID, fileID, { encryptedKey, expiresAt, password, singleUse }) {
    const { data } = await api.post(`/orgs/${orgID}/fs/file/${fileID}/share`, {
      encrypted_key: encryptedKey,
      expires_at: expiresAt ?? null,
      password: password || '',
      single_use: !!singleUse,
    })
    return data
  }

  async function fetchOrgStats(orgID) {
    const { data } = await api.get(`/orgs/${orgID}/stats`)
    orgStats.value = data
    return data
  }

  // ── Key rotation ──────────────────────────────────────────────────────────

  async function fetchAllFileKeys(orgID) {
    const { data } = await api.get(`/orgs/${orgID}/fs/all-keys`)
    return data || []
  }

  /**
   * Rotate the OrgKey for the organization.
   *
   * Flow (entirely client-side crypto — server never sees any plaintext key):
   *   1. Decrypt current OrgKey with own RSA private key
   *   2. Generate a new AES-256-GCM OrgKey
   *   3. Re-wrap every file's encrypted_key: unwrap(old) → wrap(new)
   *   4. Encrypt the new OrgKey for every member who already has a key
   *   5. POST all new keys atomically to /rotate-key
   *   6. Update cache
   *
   * @param {number} orgID
   */
  async function rotateOrgKey(orgID) {
    const oldOrgKey = await getOrgKey(orgID)
    const newOrgKey = await generateOrgKey()

    // Members must be fetched first (includes public_key for admin callers)
    if (!members.value.length) await fetchMembers(orgID)

    // Re-wrap all file keys
    const fileKeys = await fetchAllFileKeys(orgID)
    const newFileKeys = await Promise.all(fileKeys.map(async (fk) => {
      if (!fk.encrypted_key) return { file_id: fk.id, encrypted_key: '' }
      const fileKey = await unwrapFileKey(fk.encrypted_key, oldOrgKey)
      return { file_id: fk.id, encrypted_key: await wrapFileKey(fileKey, newOrgKey) }
    }))

    // Encrypt new OrgKey for each member who already has a key provisioned
    const authStore = useAuthStore()
    const newMemberKeys = await Promise.all(
      members.value
        .filter(m => m.public_key && m.encrypted_org_key)
        .map(async (m) => ({
          member_id: m.id,
          encrypted_org_key: await encryptOrgKeyForUser(newOrgKey, m.public_key),
        })),
    )

    await api.post(`/orgs/${orgID}/rotate-key`, {
      member_keys: newMemberKeys,
      file_keys: newFileKeys,
    })

    // Update cache and local state
    orgKeyCache.delete(orgID)
    orgKeyCache.set(orgID, newOrgKey)

    // Update my_encrypted_org_key on currentOrg so next load doesn't re-fetch
    const myRsaKey = authStore.publicKey || authStore.user?.public_key
    if (myRsaKey && currentOrg.value) {
      const myNewKey = await encryptOrgKeyForUser(newOrgKey, myRsaKey)
      currentOrg.value = { ...currentOrg.value, my_encrypted_org_key: myNewKey }
    }
  }

  async function transferOwnership(orgID, targetMemberID) {
    const orgKey = await getOrgKey(orgID)
    if (!members.value.length) await fetchMembers(orgID)
    const target = members.value.find(m => m.id === targetMemberID)
    if (!target?.public_key) throw new Error('Clé publique du membre introuvable.')

    const encryptedOrgKey = await encryptOrgKeyForUser(orgKey, target.public_key)
    await api.post(`/orgs/${orgID}/transfer-ownership`, {
      target_member_id: targetMemberID,
      encrypted_org_key: encryptedOrgKey,
    })

    const authStore = useAuthStore()
    const myID = authStore.user?.id || authStore.user?.user_id
    if (currentOrg.value) currentOrg.value = { ...currentOrg.value, my_role: 'admin' }
    const callerMember = members.value.find(m => m.user_id === myID)
    if (callerMember) callerMember.role = 'admin'
    const targetMember = members.value.find(m => m.id === targetMemberID)
    if (targetMember) targetMember.role = 'owner'
  }

  /**
   * Initialize the org key for the current user when they have none (e.g. org created before
   * encryption was introduced). Generates a fresh OrgKey, encrypts it with the caller's RSA
   * public key, and stores it via SetMemberKey. Safe only when no files are yet encrypted —
   * which is always the case here because uploading requires the key to already exist.
   */
  async function initializeOrgKey(orgID) {
    const authStore = useAuthStore()
    // Prefer the pre-imported CryptoKey so we skip the PEM re-import step,
    // which can fail with NotSupportedError on some browser/key combinations.
    const rsaPublicKey = authStore.publicKey || authStore.user?.public_key
    if (!rsaPublicKey) throw new Error('Clé publique manquante. Reconnectez-vous.')

    if (!members.value.length) await fetchMembers(orgID)
    const myID = authStore.user?.id || authStore.user?.user_id
    const myMember = members.value.find(m => m.user_id === myID)
    if (!myMember) throw new Error('Membre introuvable dans cette organisation.')

    let newOrgKey
    try { newOrgKey = await generateOrgKey() } catch (e) {
      throw new Error(`Génération clé org échouée: ${e.message}`)
    }

    let encryptedOrgKey
    try { encryptedOrgKey = await encryptOrgKeyForUser(newOrgKey, rsaPublicKey) } catch (e) {
      throw new Error(`Chiffrement clé org échoué (clé RSA ${typeof rsaPublicKey}): ${e.message}`)
    }

    await api.patch(`/orgs/${orgID}/members/${myMember.id}/key`, { encrypted_org_key: encryptedOrgKey })

    orgKeyCache.set(orgID, newOrgKey)
    if (currentOrg.value) currentOrg.value = { ...currentOrg.value, my_encrypted_org_key: encryptedOrgKey }
    if (myMember) myMember.encrypted_org_key = encryptedOrgKey
  }

  function $reset() {
    orgs.value = []
    currentOrg.value = null
    members.value = []
    invitations.value = []
    currentItems.value = { folders: [], files: [], current_path: '/' }
    permissions.value = []
    groups.value = []
    myGroups.value = []
    auditLog.value = []
    auditSummary.value = {}
    orgStats.value = null
    searchCache.value = null
    folderNameCache.value = {}
    orgActivity.value = []
    favorites.value = []
    trash.value = []
    orgShares.value = []
    loading.value = false
    error.value = null
    orgKeyCache.clear()
  }

  return {
    orgs, currentOrg, members, invitations, currentItems, permissions, groups, myGroups, auditLog, auditSummary, orgStats, folderNameCache, loading, error,
    fetchOrgs, fetchOrg, createOrg, updateOrg, deleteOrg, uploadOrgLogo, deleteOrgLogo,
    fetchMembers, updateMemberRole, setMemberQuota, removeMember, provisionMemberKey, provisionAllMissingKeys, setMemberKey,
    fetchInvitations, createInvitation, revokeInvitation, getInvitation, acceptInvitation,
    fetchItems, createFolder, deleteFolder, deleteFile, downloadFile, getFileKey, getFileBlob,
    renameOrgFile, renameOrgFolder,
    moveOrgFile, moveOrgFolder, getAllOrgFolders,
    orgTags, fetchOrgTags, createOrgTag, updateOrgTag, deleteOrgTag, setFileTags, setFolderTags,
    orgActivity, fetchOrgActivity,
    favorites, fetchFavorites, addFavorite, removeFavorite,
    trash, fetchTrash, restoreTrashItem, permanentDeleteTrashItem, emptyTrash,
    orgShares, fetchOrgShares, revokeOrgShare,
    downloadFolderAsZip, downloadSelectionAsZip,
    uploadOrgFile, initiateUpload, completeUpload, abortUpload,
    searchOrgItems,
    fetchFolderAccess, fetchPermissions, setPermission, deletePermission,
    fetchGroups, fetchMyGroups, createGroup, updateGroup, deleteGroup,
    addGroupMember, removeGroupMember, updateGroupMemberRole, fetchGroupMembers,
    setGroupPermission, deleteGroupPermission, fetchGroupPermissions,
    fetchAuditLog, fetchAuditSummary, deleteAuditLog, exportAuditLog, fetchAllFileKeys, rotateOrgKey, initializeOrgKey, transferOwnership,
    fetchOrgStats, createOrgFileShare,
    orgConflictState, resolveOrgConflict,
    pinnedOrgIDs, togglePin, isPinned,
    $reset,
  }
})
