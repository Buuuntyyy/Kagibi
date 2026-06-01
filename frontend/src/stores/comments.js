// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

import { defineStore } from 'pinia'
import { ref } from 'vue'
import { authClient } from '../auth-client'

function apiBase() {
  return (typeof window !== 'undefined' && window.__APP_CONFIG__?.apiUrl)
    ? window.__APP_CONFIG__.apiUrl.replace(/\/$/, '')
    : (import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1').replace(/\/$/, '')
}

async function getToken() {
  return authClient.getToken()
}

export const useCommentStore = defineStore('comments', () => {
  // Unread comment counts per file: { [fileID]: number }
  const fileCounts = ref({})
  // Unread comment counts per org file: { [orgFileID]: number }
  const orgFileCounts = ref({})

  // Active panel state
  const panelFile = ref(null)
  const panelType = ref(null)   // 'file' | 'org_file'
  const panelOrgID = ref(null)
  const panelOpen = ref(false)

  // Comments currently displayed in the panel
  const comments = ref([])
  const loading = ref(false)

  // Pending navigation: navigate to a folder then open the comment panel for a file.
  // Set by notification click; consumed by fileList when the file appears in the listing.
  const pendingNav = ref(null) // { fileID, folderPath, type, orgID }

  // ── Count loading ────────────────────────────────────────────────────────────

  async function fetchCounts(fileIDs = [], orgFileIDs = []) {
    if (!fileIDs.length && !orgFileIDs.length) return
    try {
      const token = await getToken()
      const res = await fetch(`${apiBase()}/comments/batch-counts`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify({ file_ids: fileIDs, org_file_ids: orgFileIDs })
      })
      if (!res.ok) return
      const data = await res.json()
      if (data.file_counts) {
        Object.assign(fileCounts.value, data.file_counts)
      }
      if (data.org_file_counts) {
        Object.assign(orgFileCounts.value, data.org_file_counts)
      }
    } catch (e) {
      console.error('[Comments] fetchCounts error:', e)
    }
  }

  function getCount(id, type = 'file') {
    return type === 'org_file' ? (orgFileCounts.value[id] ?? 0) : (fileCounts.value[id] ?? 0)
  }

  // ── Panel ────────────────────────────────────────────────────────────────────

  async function openPanel(file, type, orgID = null) {
    panelFile.value = file
    panelType.value = type
    panelOrgID.value = orgID
    panelOpen.value = true
    await _loadComments(file.ID ?? file.id, type, orgID)
  }

  function closePanel() {
    panelOpen.value = false
    panelFile.value = null
    panelType.value = null
    panelOrgID.value = null
    comments.value = []
  }

  async function _loadComments(fileID, type, orgID) {
    loading.value = true
    comments.value = []
    try {
      const token = await getToken()
      const url = type === 'org_file'
        ? `${apiBase()}/orgs/${orgID}/fs/file/${fileID}/comments`
        : `${apiBase()}/comments/file/${fileID}`
      const res = await fetch(url, { headers: { Authorization: `Bearer ${token}` } })
      if (!res.ok) return
      const data = await res.json()
      comments.value = data.comments ?? []
    } catch (e) {
      console.error('[Comments] load error:', e)
    } finally {
      loading.value = false
    }
  }

  // ── CRUD ─────────────────────────────────────────────────────────────────────

  async function addComment(fileID, content, type = 'file', orgID = null, parentID = null) {
    const token = await getToken()
    const url = type === 'org_file'
      ? `${apiBase()}/orgs/${orgID}/fs/file/${fileID}/comments`
      : `${apiBase()}/comments/file/${fileID}`
    const body = { content }
    if (parentID != null) body.parent_id = parentID
    const res = await fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      body: JSON.stringify(body)
    })
    if (!res.ok) throw new Error('Failed to add comment')
    await _loadComments(fileID, type, orgID)
  }

  async function editComment(commentID, content) {
    const token = await getToken()
    const res = await fetch(`${apiBase()}/comments/${commentID}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      body: JSON.stringify({ content })
    })
    if (!res.ok) throw new Error('Failed to edit comment')
    const idx = comments.value.findIndex(c => c.id === commentID)
    if (idx !== -1) comments.value[idx] = { ...comments.value[idx], content }
  }

  async function deleteComment(commentID) {
    const token = await getToken()
    const res = await fetch(`${apiBase()}/comments/${commentID}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${token}` }
    })
    if (!res.ok) throw new Error('Failed to delete comment')
    comments.value = comments.value.filter(c => c.id !== commentID)
  }

  async function markRead(commentID, fileID, type = 'file', orgID = null) {
    const token = await getToken()
    await fetch(`${apiBase()}/comments/${commentID}/read`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${token}` }
    })
    const idx = comments.value.findIndex(c => c.id === commentID)
    if (idx !== -1) comments.value[idx] = { ...comments.value[idx], is_read: true }

    // Refresh the badge count for this file
    if (type === 'org_file') await fetchCounts([], [fileID])
    else await fetchCounts([fileID], [])
  }

  async function resolveComment(commentID, isResolved, fileID, type = 'file', orgID = null) {
    const token = await getToken()
    await fetch(`${apiBase()}/comments/${commentID}/resolve`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      body: JSON.stringify({ is_resolved: isResolved })
    })
    const idx = comments.value.findIndex(c => c.id === commentID)
    if (idx !== -1) comments.value[idx] = { ...comments.value[idx], is_resolved: isResolved }
    if (type === 'org_file') await fetchCounts([], [fileID])
    else await fetchCounts([fileID], [])
  }

  function setPendingNav(fileID, folderPath, type = 'file', orgID = null) {
    pendingNav.value = { fileID, folderPath, type, orgID }
  }

  function clearPendingNav() {
    pendingNav.value = null
  }

  return {
    fileCounts, orgFileCounts,
    panelFile, panelType, panelOrgID, panelOpen,
    comments, loading,
    pendingNav,
    fetchCounts, getCount,
    openPanel, closePanel,
    addComment, editComment, deleteComment,
    markRead, resolveComment,
    setPendingNav, clearPendingNav,
  }
})
