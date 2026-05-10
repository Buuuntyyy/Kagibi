// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '../api'

export const useOrgStore = defineStore('organizations', () => {
  const orgs = ref([])
  const currentOrg = ref(null)
  const members = ref([])
  const invitations = ref([])
  const currentItems = ref({ folders: [], files: [], current_path: '/' })
  const permissions = ref([])
  const loading = ref(false)
  const error = ref(null)

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
      return data
    } catch (e) {
      error.value = e.response?.data?.error || e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  async function createOrg(name, description, storageQuotaMB) {
    const { data } = await api.post('/orgs', { name, description, storage_quota_mb: storageQuotaMB })
    orgs.value.push(data)
    return data
  }

  async function updateOrg(orgID, payload) {
    const { data } = await api.put(`/orgs/${orgID}`, payload)
    currentOrg.value = { ...currentOrg.value, ...data }
    const idx = orgs.value.findIndex(o => o.id === orgID)
    if (idx !== -1) orgs.value[idx] = { ...orgs.value[idx], ...data }
    return data
  }

  async function deleteOrg(orgID) {
    await api.delete(`/orgs/${orgID}`)
    orgs.value = orgs.value.filter(o => o.id !== orgID)
    if (currentOrg.value?.id === orgID) currentOrg.value = null
  }

  // ── Members ───────────────────────────────────────────────────────────────

  async function fetchMembers(orgID) {
    const { data } = await api.get(`/orgs/${orgID}/members`)
    members.value = data || []
    return data
  }

  async function updateMemberRole(orgID, userID, role) {
    await api.put(`/orgs/${orgID}/members/${userID}/role`, { role })
    const m = members.value.find(m => m.user_id === userID)
    if (m) m.role = role
  }

  async function removeMember(orgID, userID) {
    await api.delete(`/orgs/${orgID}/members/${userID}`)
    members.value = members.value.filter(m => m.user_id !== userID)
  }

  // ── Invitations ───────────────────────────────────────────────────────────

  async function fetchInvitations(orgID) {
    const { data } = await api.get(`/orgs/${orgID}/invitations`)
    invitations.value = data || []
    return data
  }

  async function createInvitation(orgID, payload) {
    const { data } = await api.post(`/orgs/${orgID}/invitations`, payload)
    invitations.value.unshift(data)
    return data
  }

  async function revokeInvitation(orgID, inviteID) {
    await api.delete(`/orgs/${orgID}/invitations/${inviteID}`)
    invitations.value = invitations.value.filter(i => i.id !== inviteID)
  }

  // ── File system ───────────────────────────────────────────────────────────

  async function fetchItems(orgID, folderPath = '/') {
    loading.value = true
    try {
      const encodedPath = encodeURIComponent(folderPath)
      const { data } = await api.get(`/orgs/${orgID}/fs/${encodedPath}`)
      currentItems.value = data
      return data
    } catch (e) {
      error.value = e.response?.data?.error || e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  async function createFolder(orgID, name, parentPath, encryptedKey = '') {
    const { data } = await api.post(`/orgs/${orgID}/folders`, {
      name,
      parent_path: parentPath,
      encrypted_key: encryptedKey,
    })
    currentItems.value.folders = [...(currentItems.value.folders || []), data]
    return data
  }

  async function deleteFolder(orgID, folderID) {
    await api.delete(`/orgs/${orgID}/folders/${folderID}`)
    currentItems.value.folders = currentItems.value.folders.filter(f => f.id !== folderID)
  }

  async function deleteFile(orgID, fileID) {
    await api.delete(`/orgs/${orgID}/files/${fileID}`)
    currentItems.value.files = currentItems.value.files.filter(f => f.id !== fileID)
    // Update storage_used_bytes on currentOrg
    if (currentOrg.value) {
      const file = currentItems.value.files.find(f => f.id === fileID)
      if (file) currentOrg.value.storage_used_bytes = Math.max(0, (currentOrg.value.storage_used_bytes || 0) - file.size)
    }
  }

  async function downloadFile(orgID, fileID, fileName) {
    const response = await api.get(`/orgs/${orgID}/files/${fileID}/download`, { responseType: 'blob' })
    const url = URL.createObjectURL(response.data)
    const a = document.createElement('a')
    a.href = url
    a.download = fileName
    a.click()
    URL.revokeObjectURL(url)
  }

  async function getFileKey(orgID, fileID) {
    const { data } = await api.get(`/orgs/${orgID}/files/${fileID}/key`)
    return data.encrypted_key
  }

  // ── Multipart upload ──────────────────────────────────────────────────────

  async function initiateUpload(orgID, payload) {
    const { data } = await api.post(`/orgs/${orgID}/upload/initiate`, payload)
    return data
  }

  async function completeUpload(orgID, payload) {
    const { data } = await api.post(`/orgs/${orgID}/upload/complete`, payload)
    if (data.file) {
      currentItems.value.files = [...(currentItems.value.files || []), data.file]
      if (currentOrg.value) {
        currentOrg.value.storage_used_bytes = (currentOrg.value.storage_used_bytes || 0) + data.file.size
      }
    }
    return data
  }

  async function abortUpload(orgID, uploadID, key) {
    await api.post(`/orgs/${orgID}/upload/abort`, { upload_id: uploadID, key })
  }

  // ── Permissions ───────────────────────────────────────────────────────────

  async function fetchPermissions(orgID) {
    const { data } = await api.get(`/orgs/${orgID}/permissions`)
    permissions.value = data || []
    return data
  }

  async function setPermission(orgID, payload) {
    const { data } = await api.post(`/orgs/${orgID}/permissions`, payload)
    const idx = permissions.value.findIndex(
      p => p.user_id === payload.user_id && p.folder_path === (payload.folder_path || '/')
    )
    if (idx !== -1) permissions.value[idx] = data
    else permissions.value.push(data)
    return data
  }

  async function deletePermission(orgID, userID, folderPath) {
    await api.delete(`/orgs/${orgID}/permissions`, { data: { user_id: userID, folder_path: folderPath } })
    permissions.value = permissions.value.filter(
      p => !(p.user_id === userID && p.folder_path === folderPath)
    )
  }

  function $reset() {
    orgs.value = []
    currentOrg.value = null
    members.value = []
    invitations.value = []
    currentItems.value = { folders: [], files: [], current_path: '/' }
    permissions.value = []
    loading.value = false
    error.value = null
  }

  return {
    orgs, currentOrg, members, invitations, currentItems, permissions, loading, error,
    fetchOrgs, fetchOrg, createOrg, updateOrg, deleteOrg,
    fetchMembers, updateMemberRole, removeMember,
    fetchInvitations, createInvitation, revokeInvitation,
    fetchItems, createFolder, deleteFolder, deleteFile, downloadFile, getFileKey,
    initiateUpload, completeUpload, abortUpload,
    fetchPermissions, setPermission, deletePermission,
    $reset,
  }
})
