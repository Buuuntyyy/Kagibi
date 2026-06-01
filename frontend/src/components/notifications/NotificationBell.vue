<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="notif-wrap" ref="wrapRef">
    <button
      class="notif-btn"
      :class="{ 'has-unread': notifStore.unreadCount > 0 }"
      :title="`Notifications (${notifStore.unreadCount} non lu${notifStore.unreadCount !== 1 ? 'e' : 's'})`"
      @click="togglePanel"
    >
      <Bell :size="22" :stroke-width="2" />
      <span v-if="notifStore.unreadCount > 0" class="notif-badge">
        {{ notifStore.unreadCount > 99 ? '99+' : notifStore.unreadCount }}
      </span>
    </button>

    <Transition name="notif-dropdown">
      <div v-if="open" class="notif-panel">
        <div class="notif-header">
          <span class="notif-title">Notifications</span>
          <button
            v-if="notifStore.unreadCount > 0"
            class="mark-all-btn"
            @click="notifStore.markAllRead()"
          >
            Tout marquer comme lu
          </button>
        </div>

        <div v-if="notifStore.loading" class="notif-loading">Chargement...</div>

        <div v-else-if="!notifStore.notifications.length" class="notif-empty">
          Aucune notification
        </div>

        <div v-else class="notif-list">
          <div
            v-for="notif in notifStore.notifications"
            :key="notif.id"
            class="notif-item"
            :class="{ unread: !notif.is_read }"
            @click="handleNotifClick(notif)"
          >
            <div class="notif-icon">
              <MessageSquare :size="16" />
            </div>
            <div class="notif-body">
              <p class="notif-text">
                <strong>{{ notif.actor_name }}</strong>
                a commenté sur
                <strong>{{ notif.resource_name }}</strong>
              </p>
              <span class="notif-date">{{ formatDate(notif.created_at) }}</span>
            </div>
            <div v-if="!notif.is_read" class="notif-dot" />
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { Bell, MessageSquare } from 'lucide-vue-next'
import { useNotificationStore } from '../../stores/notifications'
import { useRealtimeStore } from '../../stores/realtime'

const notifStore = useNotificationStore()
const realtimeStore = useRealtimeStore()

const open = ref(false)
const wrapRef = ref(null)

function togglePanel() {
  open.value = !open.value
  if (open.value && !notifStore.notifications.length) {
    notifStore.fetchNotifications()
  }
}

async function handleNotifClick(notif) {
  if (!notif.is_read) await notifStore.markRead(notif.id)
}

function formatDate(iso) {
  if (!iso) return ''
  const d = new Date(iso)
  const now = new Date()
  const diff = (now - d) / 1000
  if (diff < 60) return 'à l\'instant'
  if (diff < 3600) return `il y a ${Math.floor(diff / 60)} min`
  if (diff < 86400) return `il y a ${Math.floor(diff / 3600)} h`
  return d.toLocaleDateString('fr-FR')
}

function onClickOutside(e) {
  if (wrapRef.value && !wrapRef.value.contains(e.target)) {
    open.value = false
  }
}

let unsubscribe = null

onMounted(() => {
  notifStore.fetchNotifications()
  document.addEventListener('click', onClickOutside)

  // Listen for real-time notification updates
  unsubscribe = realtimeStore.onEvent('notification_update', () => {
    notifStore.incrementUnread()
    if (open.value) notifStore.fetchNotifications()
  })
})

onUnmounted(() => {
  document.removeEventListener('click', onClickOutside)
  if (unsubscribe) unsubscribe()
})
</script>

<style scoped>
.notif-wrap {
  position: relative;
  display: inline-flex;
  align-items: center;
}

.notif-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--main-text-color);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.3rem;
  border-radius: 6px;
  position: relative;
}
.notif-btn:hover { background: var(--hover-background-color); }
.notif-btn.has-unread { color: var(--primary-color); }

.notif-badge {
  position: absolute;
  top: -2px;
  right: -4px;
  background: #ef4444;
  color: #fff;
  font-size: 0.6rem;
  font-weight: 700;
  min-width: 16px;
  height: 16px;
  border-radius: 99px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 3px;
  line-height: 1;
}

/* Dropdown panel */
.notif-panel {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  width: 340px;
  max-height: 480px;
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 10px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  z-index: 3000;
}

.notif-dropdown-enter-active,
.notif-dropdown-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}
.notif-dropdown-enter-from,
.notif-dropdown-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}

.notif-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--border-color);
  flex-shrink: 0;
}

.notif-title {
  font-weight: 600;
  font-size: 0.9rem;
  color: var(--main-text-color);
}

.mark-all-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 0.75rem;
  color: var(--primary-color);
  padding: 0;
}
.mark-all-btn:hover { text-decoration: underline; }

.notif-loading,
.notif-empty {
  padding: 2rem 1rem;
  text-align: center;
  color: var(--secondary-text-color);
  font-size: 0.88rem;
}

.notif-list {
  overflow-y: auto;
  flex: 1;
}

.notif-item {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  cursor: pointer;
  border-bottom: 1px solid var(--border-color);
  position: relative;
}
.notif-item:last-child { border-bottom: none; }
.notif-item:hover { background: var(--hover-background-color); }
.notif-item.unread { background: color-mix(in srgb, var(--primary-color) 6%, var(--card-color)); }

.notif-icon {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: var(--hover-background-color);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  color: var(--primary-color);
}

.notif-body {
  flex: 1;
  min-width: 0;
}

.notif-text {
  margin: 0 0 2px;
  font-size: 0.83rem;
  color: var(--main-text-color);
  line-height: 1.4;
  word-break: break-word;
}

.notif-date {
  font-size: 0.73rem;
  color: var(--secondary-text-color);
}

.notif-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--primary-color);
  flex-shrink: 0;
  margin-top: 4px;
}
</style>
