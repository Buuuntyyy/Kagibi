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

export const useNotificationStore = defineStore('notifications', () => {
  const notifications = ref([])
  const unreadCount = ref(0)
  const loading = ref(false)

  // Internal handles for realtime subscription and polling fallback.
  let _unsubRealtime = null
  let _pollTimer = null

  async function fetchNotifications() {
    loading.value = true
    try {
      const token = await getToken()
      const res = await fetch(`${apiBase()}/notifications`, {
        headers: { Authorization: `Bearer ${token}` }
      })
      if (!res.ok) return
      const data = await res.json()
      notifications.value = data.notifications ?? []
      unreadCount.value = notifications.value.filter(n => !n.is_read).length
    } catch (e) {
      console.error('[Notifications] fetch error:', e)
    } finally {
      loading.value = false
    }
  }

  async function markRead(id) {
    const token = await getToken()
    await fetch(`${apiBase()}/notifications/${id}/read`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${token}` }
    })
    const idx = notifications.value.findIndex(n => n.id === id)
    if (idx !== -1 && !notifications.value[idx].is_read) {
      notifications.value[idx] = { ...notifications.value[idx], is_read: true }
      unreadCount.value = Math.max(0, unreadCount.value - 1)
    }
  }

  async function markAllRead() {
    const token = await getToken()
    await fetch(`${apiBase()}/notifications/read-all`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${token}` }
    })
    notifications.value = notifications.value.map(n => ({ ...n, is_read: true }))
    unreadCount.value = 0
  }

  async function deleteNotification(id) {
    const token = await getToken()
    await fetch(`${apiBase()}/notifications/${id}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${token}` }
    })
    const idx = notifications.value.findIndex(n => n.id === id)
    if (idx !== -1) {
      if (!notifications.value[idx].is_read) unreadCount.value = Math.max(0, unreadCount.value - 1)
      notifications.value.splice(idx, 1)
    }
  }

  /**
   * Wire up realtime updates: WebSocket subscription + 30 s polling fallback.
   * Must be called once after authentication (App.vue). Idempotent.
   */
  function connectRealtime(realtimeStore) {
    // WebSocket: instant delivery when connection is alive.
    if (_unsubRealtime) {
      _unsubRealtime()
      _unsubRealtime = null
    }
    _unsubRealtime = realtimeStore.onEvent('notification_update', () => {
      // Single authoritative fetch — avoids the race between manual ++ and
      // the async response that could overwrite the optimistic increment.
      fetchNotifications()
    })

    // Polling fallback: catches events missed during WS reconnection gaps,
    // HMR reloads in dev, or any other delivery failure.
    _startPolling()
  }

  /** Remove realtime subscription and stop polling on logout. */
  function disconnectRealtime() {
    if (_unsubRealtime) {
      _unsubRealtime()
      _unsubRealtime = null
    }
    _stopPolling()
    notifications.value = []
    unreadCount.value = 0
  }

  function _startPolling() {
    _stopPolling()
    // Poll every 30 s. Silent background update — only updates the count,
    // not the full notification list (that loads on panel open).
    _pollTimer = setInterval(async () => {
      try {
        const token = await getToken()
        const res = await fetch(`${apiBase()}/notifications/unread-count`, {
          headers: { Authorization: `Bearer ${token}` }
        })
        if (!res.ok) return
        const { count } = await res.json()
        // Only update if the value actually changed to avoid unnecessary re-renders.
        if (count !== unreadCount.value) unreadCount.value = count
      } catch {
        // Silent — polling is a best-effort fallback.
      }
    }, 30_000)
  }

  function _stopPolling() {
    if (_pollTimer) {
      clearInterval(_pollTimer)
      _pollTimer = null
    }
  }

  return {
    notifications, unreadCount, loading,
    fetchNotifications, markRead, markAllRead, deleteNotification,
    connectRealtime, disconnectRealtime,
  }
})
