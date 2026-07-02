<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="profile-menu-wrap" ref="wrapRef">
    <!-- Avatar trigger -->
    <button
      class="avatar-btn"
      :title="authStore.user?.name || 'Mon profil'"
      @click="toggle"
      aria-haspopup="true"
      :aria-expanded="open"
    >
      <div class="user-avatar" :class="{ 'has-unread': notifStore.unreadCount > 0 }">
        <img
          v-if="authStore.user?.avatar_url"
          :src="authStore.user.avatar_url"
          :alt="authStore.user?.name"
          class="avatar-image"
          @error="onImgError"
        />
        <div v-else class="avatar-fallback">{{ initials }}</div>
        <span v-if="notifStore.unreadCount > 0" class="notif-badge">
          {{ notifStore.unreadCount > 99 ? '99+' : notifStore.unreadCount }}
        </span>
      </div>
    </button>

    <!-- Dropdown panel -->
    <Transition name="menu-drop">
      <div v-if="open" class="menu-panel" role="dialog" aria-label="Centre de notifications">

        <!-- User identity header -->
        <div class="menu-header">
          <div class="menu-avatar">{{ initials }}</div>
          <div class="menu-user-info">
            <span class="menu-user-name">{{ authStore.user?.name || '—' }}</span>
            <router-link to="/account" class="menu-account-link" @click="open = false">
              Mon compte
            </router-link>
          </div>
        </div>

        <!-- Notifications section -->
        <div class="notif-section">
          <div class="notif-section-header">
            <span class="notif-section-title">
              Notifications
              <span v-if="notifStore.unreadCount > 0" class="unread-chip">
                {{ notifStore.unreadCount }} non lu{{ notifStore.unreadCount !== 1 ? 'e' : '' }}{{ notifStore.unreadCount !== 1 ? 's' : '' }}
              </span>
            </span>
            <button
              v-if="notifStore.unreadCount > 0"
              class="mark-all-btn"
              @click="notifStore.markAllRead()"
            >
              Tout marquer lu
            </button>
          </div>

          <div v-if="notifStore.loading" class="notif-state">Chargement…</div>

          <div v-else-if="!notifStore.notifications.length" class="notif-state">
            <BellOff :size="32" class="notif-empty-icon" />
            <p>Aucune notification</p>
          </div>

          <div v-else class="notif-list">
            <div
              v-for="notif in notifStore.notifications"
              :key="notif.id"
              class="notif-item"
              :class="{ unread: !notif.is_read }"
              @click="handleNotifClick(notif)"
            >
              <div class="notif-icon-wrap">
                <Reply v-if="notif.type === 'reply_added'" :size="15" />
                <MessageSquare v-else :size="15" />
              </div>
              <div class="notif-body">
                <p class="notif-text">
                  <strong>{{ notif.actor_name }}</strong>
                  {{ notif.type === 'reply_added' ? 'a répondu à votre commentaire sur' : 'a commenté' }}
                  <strong>{{ notif.resource_name }}</strong>
                </p>
                <span class="notif-date">{{ relativeDate(notif.created_at) }}</span>
              </div>
              <span v-if="!notif.is_read" class="notif-dot" />
              <button
                class="notif-delete-btn"
                title="Supprimer"
                @click.stop="notifStore.deleteNotification(notif.id)"
              >
                <X :size="13" />
              </button>
            </div>
          </div>
        </div>

        <!-- Footer actions -->
        <div class="menu-footer">
          <button class="footer-btn logout-btn" @click="logout">
            <LogOut :size="15" />
            Déconnexion
          </button>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../../stores/auth'
import { useNotificationStore } from '../../stores/notifications'
import { useCommentStore } from '../../stores/comments'
import { useFileStore } from '../../stores/files'
import { MessageSquare, BellOff, LogOut, Reply, X } from 'lucide-vue-next'

const authStore = useAuthStore()
const notifStore = useNotificationStore()
const commentStore = useCommentStore()
const fileStore = useFileStore()
const router = useRouter()

const open = ref(false)
const wrapRef = ref(null)

const initials = computed(() => {
  const name = authStore.user?.name
  if (!name) return '?'
  return name.split(' ').map(w => w[0]).join('').slice(0, 2).toUpperCase()
})

function toggle() {
  open.value = !open.value
  // Refresh the full list when opening the panel so the user sees up-to-date content.
  if (open.value) notifStore.fetchNotifications()
}

async function handleNotifClick(notif) {
  // Mark as read
  if (!notif.is_read) await notifStore.markRead(notif.id)

  // Close the menu
  open.value = false

  if (notif.resource_type === 'file') {
    // Store pending navigation so fileList can open the panel once the file loads
    commentStore.setPendingNav(notif.resource_id, notif.resource_path || '/', 'file', null)

    // Navigate to the files view; fileList will consume the pending nav
    if (router.currentRoute.value.name !== 'MyFiles') {
      await router.push({ name: 'MyFiles' })
    }
    // If already on MyFiles, navigate the store to the right folder directly
    else if (notif.resource_path) {
      await fileStore.fetchItems(notif.resource_path)
    }
  } else if (notif.resource_type === 'org_file' && notif.org_id) {
    commentStore.setPendingNav(notif.resource_id, notif.resource_path || '/', 'org_file', notif.org_id)
    if (router.currentRoute.value.name !== 'OrgDetail' || String(router.currentRoute.value.params.orgID) !== String(notif.org_id)) {
      await router.push({ name: 'OrgDetail', params: { orgID: notif.org_id } })
    }
  }
}

async function logout() {
  open.value = false
  await authStore.logout()
  router.push({ name: 'Login' })
}

function onImgError(e) {
  e.target.style.display = 'none'
  const fallback = e.target.nextElementSibling
  if (fallback) fallback.style.display = 'flex'
}

function relativeDate(iso) {
  if (!iso) return ''
  const d = new Date(iso)
  const diff = (Date.now() - d) / 1000
  if (diff < 60) return 'à l\'instant'
  if (diff < 3600) return `il y a ${Math.floor(diff / 60)} min`
  if (diff < 86400) return `il y a ${Math.floor(diff / 3600)} h`
  return d.toLocaleDateString('fr-FR', { day: 'numeric', month: 'short' })
}

function onClickOutside(e) {
  if (wrapRef.value && !wrapRef.value.contains(e.target)) open.value = false
}

onMounted(() => document.addEventListener('click', onClickOutside))
onUnmounted(() => document.removeEventListener('click', onClickOutside))
</script>

<style scoped>
.profile-menu-wrap {
  position: relative;
  display: inline-flex;
  align-items: center;
}

/* ── Avatar button ────────────────────────────────────────────────────────── */
.avatar-btn {
  background: none;
  border: none;
  padding: 0;
  cursor: pointer;
  display: flex;
  align-items: center;
}

.user-avatar {
  position: relative;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  overflow: visible;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform 0.2s, box-shadow 0.2s;
  border: 2px solid var(--border-color);
}

.avatar-btn:hover .user-avatar {
  transform: scale(1.05);
  box-shadow: 0 4px 12px rgba(107, 127, 215, 0.3);
  border-color: var(--primary-color);
}

.user-avatar.has-unread {
  border-color: var(--primary-color);
}

.avatar-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 50%;
}

.avatar-fallback {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-weight: 600;
  font-size: 0.875rem;
  letter-spacing: 0.5px;
  background: linear-gradient(135deg, var(--primary-color, #6B7FD7) 0%, var(--secondary-color, #9370DB) 100%);
  border-radius: 50%;
}

.notif-badge {
  position: absolute;
  top: -4px;
  right: -4px;
  background: #ef4444;
  color: #fff;
  font-size: 0.6rem;
  font-weight: 700;
  min-width: 17px;
  height: 17px;
  border-radius: 99px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 3px;
  line-height: 1;
  border: 2px solid var(--background-color);
  box-sizing: border-box;
}

/* ── Dropdown panel ───────────────────────────────────────────────────────── */
.menu-panel {
  position: absolute;
  top: calc(100% + 10px);
  right: 0;
  width: 360px;
  max-height: 540px;
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.18);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  z-index: 3000;
}

/* Transition */
.menu-drop-enter-active,
.menu-drop-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}
.menu-drop-enter-from,
.menu-drop-leave-to {
  opacity: 0;
  transform: translateY(-6px) scale(0.98);
}

/* ── Header ───────────────────────────────────────────────────────────────── */
.menu-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem 1.1rem 0.85rem;
  border-bottom: 1px solid var(--border-color);
  flex-shrink: 0;
}

.menu-avatar {
  width: 38px;
  height: 38px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--primary-color, #6B7FD7) 0%, var(--secondary-color, #9370DB) 100%);
  color: #fff;
  font-size: 0.8rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.menu-user-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.menu-user-name {
  font-weight: 600;
  font-size: 0.9rem;
  color: var(--main-text-color);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.menu-account-link {
  font-size: 0.78rem;
  color: var(--primary-color);
  text-decoration: none;
}
.menu-account-link:hover { text-decoration: underline; }

/* ── Notification section ─────────────────────────────────────────────────── */
.notif-section {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.notif-section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.6rem 1.1rem 0.4rem;
  flex-shrink: 0;
}

.notif-section-title {
  font-weight: 600;
  font-size: 0.82rem;
  color: var(--secondary-text-color);
  text-transform: uppercase;
  letter-spacing: 0.4px;
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.unread-chip {
  background: color-mix(in srgb, var(--primary-color) 15%, transparent);
  color: var(--primary-color);
  font-size: 0.68rem;
  font-weight: 700;
  padding: 1px 6px;
  border-radius: 99px;
  text-transform: none;
  letter-spacing: 0;
}

.mark-all-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 0.75rem;
  color: var(--primary-color);
  padding: 0;
  white-space: nowrap;
}
.mark-all-btn:hover { text-decoration: underline; }

.notif-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 1.5rem 1rem;
  color: var(--secondary-text-color);
  font-size: 0.85rem;
  gap: 0.5rem;
}

.notif-empty-icon { opacity: 0.3; }

.notif-list {
  overflow-y: auto;
}

.notif-item {
  display: flex;
  align-items: flex-start;
  gap: 0.65rem;
  padding: 0.65rem 1.1rem;
  cursor: pointer;
  border-bottom: 1px solid var(--border-color);
  position: relative;
}
.notif-item:last-child { border-bottom: none; }
.notif-item:hover { background: var(--hover-background-color); }
.notif-item.unread {
  background: color-mix(in srgb, var(--primary-color) 5%, var(--card-color));
}

.notif-icon-wrap {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: var(--hover-background-color);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  color: var(--primary-color);
  margin-top: 1px;
}

.notif-body {
  flex: 1;
  min-width: 0;
}

.notif-text {
  margin: 0 0 2px;
  font-size: 0.82rem;
  color: var(--main-text-color);
  line-height: 1.45;
  word-break: break-word;
}

.notif-date {
  font-size: 0.72rem;
  color: var(--secondary-text-color);
}

.notif-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--primary-color);
  flex-shrink: 0;
  margin-top: 5px;
}

.notif-delete-btn {
  flex-shrink: 0;
  background: none;
  border: none;
  cursor: pointer;
  color: var(--secondary-text-color);
  display: none;
  align-items: center;
  justify-content: center;
  padding: 3px;
  border-radius: 4px;
  margin-left: 2px;
}
.notif-item:hover .notif-delete-btn { display: flex; }
.notif-delete-btn:hover { background: #fee2e2; color: #dc2626; }

/* ── Footer ───────────────────────────────────────────────────────────────── */
.menu-footer {
  flex-shrink: 0;
  padding: 0.5rem 0.75rem;
  border-top: 1px solid var(--border-color);
  display: flex;
  justify-content: flex-end;
}

.footer-btn {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  background: none;
  border: none;
  cursor: pointer;
  font-size: 0.82rem;
  color: var(--secondary-text-color);
  padding: 0.35rem 0.65rem;
  border-radius: 6px;
}
.footer-btn:hover { background: var(--hover-background-color); color: var(--main-text-color); }
.logout-btn:hover { color: #dc2626; background: #fee2e2; }

/* ── Mobile ───────────────────────────────────────────────────────────────── */
@media (max-width: 480px) {
  .menu-panel {
    position: fixed;
    top: 60px;
    left: 0;
    right: 0;
    width: 100%;
    border-radius: 0 0 12px 12px;
    max-height: calc(100vh - 60px);
  }
}
</style>
