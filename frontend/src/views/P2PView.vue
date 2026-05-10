<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="dashboard-container">
    <LeftBar />
    <div class="main-content">
      <div class="p2p-page">

        <!-- Incoming invite banner -->
        <div v-if="incomingInvite || inviteError || inviteLoading" class="invite-banner-wrap">
          <div class="invite-banner">
            <template v-if="inviteLoading">
              <span class="invite-banner-text">{{ t('common.loading') }}</span>
            </template>
            <template v-else-if="inviteError">
              <span class="invite-banner-error">{{ inviteError }}</span>
            </template>
            <template v-else-if="inviteAccepted">
              <span class="invite-banner-text">{{ t('p2p.invite.accepted') }}</span>
            </template>
            <template v-else-if="incomingInvite">
              <div class="invite-banner-content">
                <div class="invite-banner-info">
                  <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"/><polyline points="13 2 13 9 20 9"/></svg>
                  <span>{{ t('p2p.invite.incomingDesc', { sender: incomingInvite.sender_name, fileName: incomingInvite.file_name, size: formatSize(incomingInvite.file_size) }) }}</span>
                </div>
                <div class="invite-banner-actions">
                  <button class="invite-btn-decline" @click="incomingInvite = null" :disabled="inviteLoading">{{ t('p2p.invite.decline') }}</button>
                  <button class="invite-btn-accept" @click="acceptInvite" :disabled="inviteLoading">{{ t('p2p.invite.accept') }}</button>
                </div>
              </div>
            </template>
          </div>
        </div>

        <div class="p2p-container">
          <div class="main-card">
        <div class="card-header">
           <h2>{{ t('p2p.pageTitle') }}</h2>
           <p>{{ t('p2p.pageDesc') }}</p>
        </div>

        <div class="transfer-flow">
            <!-- STEP 1: DESTINATAIRE -->
            <div class="flow-step" :class="{ active: !selectedFriend && !inviteMode, completed: selectedFriend || inviteMode }">
                <div class="step-icon">1</div>
                <div class="step-content">
                    <h3>{{ t('p2p.selectFriend') }}</h3>

                    <!-- Selection not made yet -->
                    <div v-if="!selectedFriend && !inviteMode">
                        <p class="secondary-text">{{ t('p2p.selectFriendDesc') }}</p>
                        <div class="recipient-options">
                            <div class="friends-grid">
                                <div v-if="onlineFriends.length === 0" class="no-friends">
                                    {{ t('p2p.noOnlineFriends') }}
                                    <router-link to="/dashboard/friends" class="friends-page-link">{{ t('friends.myFriends') }} →</router-link>
                                </div>
                                <template v-else>
                                    <div
                                        v-for="friend in onlineFriends"
                                        :key="friend.id"
                                        class="friend-chip"
                                        @click="selectedFriend = friend"
                                    >
                                        <div class="avatar-mini-wrapper">
                                            <div class="avatar-mini">
                                                <img
                                                    v-if="normalizeAvatarUrl(friend.avatar_url)"
                                                    :src="normalizeAvatarUrl(friend.avatar_url)"
                                                    :alt="friend.name"
                                                    class="avatar-image"
                                                    @error="(e) => e.target.style.display = 'none'"
                                                />
                                                <span v-else class="avatar-initials-mini">{{ getInitials(friend.name) }}</span>
                                            </div>
                                            <span class="status-dot-mini"></span>
                                        </div>
                                        <span>{{ friend.name }}</span>
                                    </div>
                                </template>
                            </div>
                            <div class="invite-option-chip" @click="inviteMode = true">
                                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg>
                                {{ t('p2p.invite.title') }}
                            </div>
                        </div>
                    </div>

                    <!-- Friend selected -->
                    <div v-else-if="selectedFriend" class="selected-friend-display" style="width: fit-content">
                        <div class="friend-chip selected">
                            <div class="avatar-mini-wrapper">
                                <div class="avatar-mini">
                                    <img
                                        v-if="normalizeAvatarUrl(selectedFriend.avatar_url)"
                                        :src="normalizeAvatarUrl(selectedFriend.avatar_url)"
                                        :alt="selectedFriend.name"
                                        class="avatar-image"
                                        @error="(e) => e.target.style.display = 'none'"
                                    />
                                    <span v-else class="avatar-initials-mini">{{ getInitials(selectedFriend.name) }}</span>
                                </div>
                                <span class="status-dot-mini"></span>
                            </div>
                            <span>{{ selectedFriend.name }}</span>
                            <button class="close-btn" @click="selectedFriend = null">×</button>
                        </div>
                    </div>

                    <!-- Invite mode selected -->
                    <div v-else class="invite-mode-display">
                        <div class="invite-option-chip selected">
                            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg>
                            {{ t('p2p.invite.title') }}
                        </div>
                        <button class="close-mode-btn" @click="inviteMode = false">×</button>
                    </div>

                </div>
            </div>

            <!-- STEP 2: FICHIER -->
            <div class="flow-step" :class="{ active: !selectedFile, completed: selectedFile }">
                <div class="step-icon">2</div>
                <div class="step-content">
                    <h3>{{ t('p2p.selectFile') }}</h3>
                    <input type="file" id="p2p-file-input" @change="handleFileSelect" style="display: none" />

                    <div v-if="!selectedFile"
                         class="drop-area pulse"
                         @click="triggerFileSelect()"
                         @dragover.prevent
                         @drop.prevent="handleDrop"
                    >
                        <span>{{ t('p2p.dropFile') }}</span>
                    </div>

                    <div v-else class="file-display">
                        <div class="file-icon">
                            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"></path><polyline points="13 2 13 9 20 9"></polyline></svg>
                        </div>
                        <div class="file-meta">
                            <span class="fname" :title="selectedFile.name">{{ selectedFile.name }}</span>
                            <span class="fsize">{{ formatSize(selectedFile.size) }}</span>
                        </div>
                        <button class="change-file-btn" @click="selectedFile = null">{{ t('p2p.changeFile') }}</button>
                    </div>
                </div>
            </div>

            <!-- STEP 3: ACTION -->
            <div class="flow-step action-step">
                <!-- Legal consent for direct transfer -->
                <label v-if="!inviteMode && selectedFriend && selectedFile" class="direct-consent-toggle">
                  <input type="checkbox" v-model="directLegalConsent" />
                  <span>
                    Je certifie que ce fichier est légal et j'en assume l'entière responsabilité conformément aux <router-link to="/terms" target="_blank">CGU</router-link>.
                  </span>
                </label>
                <!-- Direct transfer mode -->
                <button v-if="!inviteMode" class="send-big-btn" :disabled="!canSend" @click="startTransfer">
                    <span v-if="!canSend">{{ t('p2p.selectFriendAndFile') }}</span>
                    <span v-else>
                        {{ t('p2p.sendFile') }}
                        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="margin-left:8px;"><line x1="22" y1="2" x2="11" y2="13"></line><polygon points="22 2 15 22 11 13 2 9 22 2"></polygon></svg>
                    </span>
                </button>
                <!-- Invite mode -->
                <button v-else class="send-big-btn invite-send-btn" :disabled="!selectedFile" @click="openInviteDialog">
                    <span v-if="!selectedFile">{{ t('p2p.selectFile') }}</span>
                    <span v-else>
                        {{ t('p2p.invite.generate') }}
                        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="margin-left:8px;"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg>
                    </span>
                </button>
            </div>
        </div>
      </div>
    </div>
    </div>
    </div>
    <!-- Mobile Bottom Navigation -->
    <MobileBottomNav />
  </div>

  <P2PInviteDialog
    :visible="showInviteDialog"
    :file="selectedFile"
    :transfer-id="inviteTransferId"
    @close="showInviteDialog = false"
  />
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useFriendStore } from '../stores/friends'
import { useAuthStore } from '../stores/auth'
import { useP2PStore } from '../stores/p2p'
import { authClient } from '../auth-client'
import { API_BASE_URL } from '../api'
import LeftBar from '../components/bar/leftBar.vue'
import MobileBottomNav from '../components/bar/MobileBottomNav.vue'
import P2PInviteDialog from '../components/p2p/P2PInviteDialog.vue'

const { t } = useI18n()
const route = useRoute()
const friendStore = useFriendStore()
const authStore = useAuthStore()
const p2pStore = useP2PStore()

const selectedFriend = ref(null)
const selectedFile = ref(null)
const inviteMode = ref(false)
const showInviteDialog = ref(false)
const inviteTransferId = ref('')

// Invite acceptance state
const incomingInvite = ref(null)
const inviteError = ref('')
const inviteAccepted = ref(false)
const inviteLoading = ref(false)

const onlineFriends = computed(() => {
  return friendStore.acceptedFriends.filter(f => f.online)
})

const directLegalConsent = ref(false)
const canSend = computed(() => !!selectedFriend.value && !!selectedFile.value && !inviteMode.value && directLegalConsent.value)

watch(() => p2pStore.inviteReady, (ready) => {
  if (ready && showInviteDialog.value) showInviteDialog.value = false
})

function resetWizard() {
  selectedFriend.value    = null
  selectedFile.value      = null
  inviteMode.value        = false
  showInviteDialog.value  = false
  inviteTransferId.value  = ''
  directLegalConsent.value = false
}

watch(() => p2pStore.activeTransfer, (current, previous) => {
  if (!current && previous?.type === 'send' && previous?.status === 'Done') {
    resetWizard()
  }
})

onMounted(async () => {
    await friendStore.fetchFriends()

    const friendId = route.query.friendId
    if (friendId) {
      const friend = friendStore.acceptedFriends.find(f => String(f.id) === String(friendId))
      if (friend) selectedFriend.value = friend
    }

    const token = route.query.invite
    if (token) {
      await loadInvite(token)
    }
})

async function loadInvite(token) {
  inviteLoading.value = true
  inviteError.value = ''
  try {
    const jwt = await authClient.getToken()
    const res = await fetch(`${API_BASE_URL}p2p/invite/${token}`, {
      headers: { Authorization: `Bearer ${jwt}` },
    })
    const data = await res.json()
    if (res.status === 410) { inviteError.value = t('p2p.invite.expired'); return }
    if (res.status === 403) { inviteError.value = t('p2p.invite.notForYou'); return }
    if (!res.ok) { inviteError.value = data.error || t('p2p.invite.createError'); return }
    incomingInvite.value = { ...data, token }
  } catch {
    inviteError.value = t('p2p.invite.createError')
  } finally {
    inviteLoading.value = false
  }
}

async function acceptInvite() {
  if (!incomingInvite.value) return
  inviteLoading.value = true
  inviteError.value = ''
  try {
    const jwt = await authClient.getToken()
    const res = await fetch(`${API_BASE_URL}p2p/invite/${incomingInvite.value.token}/accept`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${jwt}` },
    })
    const data = await res.json()
    if (!res.ok) { inviteError.value = data.error || t('p2p.invite.createError'); return }
    inviteAccepted.value = true
  } catch {
    inviteError.value = t('p2p.invite.createError')
  } finally {
    inviteLoading.value = false
  }
}

function openInviteDialog() {
  inviteTransferId.value = crypto.randomUUID()
  showInviteDialog.value = true
}

const getInitials = (name) => {
    if (!name) return '?'
    return name.substring(0, 2).toUpperCase()
}

const normalizeAvatarUrl = (url) => {
  if (!url) return null
  if (url.startsWith('http')) return url
  if (url.startsWith('/avatars/')) return url
  const cleanUrl = url.startsWith('/') ? url.substring(1) : url
  return `/avatars/${cleanUrl}`
}

const formatSize = (bytes) => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const triggerFileSelect = () => {
    document.getElementById('p2p-file-input').click()
}

const handleFileSelect = (e) => {
    if (e.target.files.length > 0) {
        selectedFile.value = e.target.files[0]
    }
}

const handleDrop = (e) => {
    if (e.dataTransfer.files.length > 0) {
        selectedFile.value = e.dataTransfer.files[0]
    }
}

const startTransfer = async () => {
    if (!canSend.value) return
    try {
        await p2pStore.startTransfer(selectedFriend.value, selectedFile.value)
    } catch (e) {
        console.error("Transfer failed", e)
        alert("Erreur: " + e.message)
    }
}
</script>

<style scoped>
.dashboard-container {
  display: flex;
  height: 100%;
  width: 100%;
  box-sizing: border-box;
  background-color: var(--background-color);
}

.main-content {
  flex-grow: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  border-top-left-radius: 30px;
  background-color: var(--card-color);
}

.p2p-page {
    background-color: var(--background-color);
    min-height: 100vh;
    display: flex;
    flex-direction: column;
}

.p2p-container {
    flex: 1;
    display: flex;
    justify-content: center;
    align-items: flex-start;
    padding-top: 2rem;
    padding-bottom: 2rem;
    overflow-y: auto;
}

.main-card {
    background: var(--card-color);
    border: 1px solid var(--border-color);
    border-radius: 16px;
    width: 600px;
    padding: 2.5rem;
    box-shadow: 0 4px 20px rgba(0,0,0,0.04);
    margin-bottom: 2rem;
}

.card-header {
    text-align: center;
    margin-bottom: 3rem;
}
.card-header h2 { margin: 0 0 0.5rem 0; color: var(--main-text-color); }
.card-header p { margin: 0; color: var(--secondary-text-color); }

.transfer-flow {
    display: flex;
    flex-direction: column;
    gap: 2rem;
}

.flow-step {
    display: flex;
    gap: 1.5rem;
    padding-bottom: 2rem;
    border-bottom: 1px dashed var(--border-color);
}
.flow-step:last-child {
    border-bottom: none;
    padding-bottom: 0;
}
.flow-step.disabled {
    opacity: 0.5;
    pointer-events: none;
}

.step-icon {
    width: 40px; height: 40px;
    background: var(--border-color);
    color: var(--secondary-text-color);
    font-weight: 700;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
}

.flow-step.active .step-icon {
    background: var(--primary-color);
    color: white;
    box-shadow: 0 0 0 4px rgba(52, 152, 219, 0.2);
}

.flow-step.completed .step-icon {
    background: var(--success-color);
    color: white;
}

.step-content {
    flex: 1;
    min-width: 0;
}
.step-content h3 {
    margin: 0 0 0.5rem 0;
    font-size: 1.1rem;
    font-weight: 600;
}

.secondary-text {
    font-size: 0.9rem;
    color: var(--secondary-text-color);
    margin-bottom: 0.5rem;
}

.friends-grid {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
}

.friend-chip {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.4rem 0.8rem;
    background: var(--hover-background-color);
    border-radius: 20px;
    cursor: pointer;
    border: 1px solid transparent;
    transition: all 0.2s;
}
.friend-chip:hover {
    border-color: var(--primary-color);
}
.friend-chip.selected {
    background: var(--primary-color);
    color: white;
}

.avatar-mini {
    width: 24px; height: 24px;
    background: rgba(255,255,255,0.3);
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.7rem;
    overflow: hidden;
    position: relative;
}

.avatar-mini-wrapper {
    position: relative;
    display: inline-block;
}

.avatar-image {
    width: 100%;
    height: 100%;
    object-fit: cover;
    border-radius: 50%;
}

.avatar-initials-mini {
    color: #5f6368;
    font-size: 0.7rem;
    font-weight: 500;
}

.status-dot-mini {
    position: absolute;
    bottom: -2px;
    right: -2px;
    width: 8px;
    height: 8px;
    background-color: #34a853;
    border-radius: 50%;
    border: 2px solid white;
    z-index: 1;
}

.drop-area {
    border: 2px dashed var(--border-color);
    border-radius: 8px;
    padding: 2rem;
    text-align: center;
    color: var(--secondary-text-color);
    cursor: pointer;
    transition: all 0.2s;
}
.drop-area:hover, .drop-area.pulse {
    border-color: var(--primary-color);
    background: var(--hover-background-color);
    color: var(--primary-color);
}

.file-display {
    display: flex;
    align-items: center;
    gap: 1rem;
    background: var(--hover-background-color);
    padding: 1rem;
    border-radius: 8px;
    border-left: 4px solid var(--primary-color);
}
.file-icon { font-size: 1.5rem; }
.file-meta { flex: 1; display:flex; flex-direction:column; overflow: hidden; min-width: 0; }
.fname { font-weight: 600; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.fsize { font-size: 0.85rem; color: var(--secondary-text-color); }
.change-file-btn { font-size: 0.8rem; padding: 0.2rem 0.6rem; }
.close-btn { background:transparent; border:none; color:white; font-size:1.2rem; margin-left:0.5rem; cursor:pointer;}

.direct-consent-toggle {
    display: flex;
    align-items: flex-start;
    gap: 0.5rem;
    font-size: 0.82rem;
    color: var(--text-secondary);
    cursor: pointer;
    user-select: none;
    line-height: 1.4;
    padding: 0.6rem 0.8rem;
    background: var(--background-color);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    margin-bottom: 0.75rem;
}

.direct-consent-toggle input[type="checkbox"] {
    width: 15px;
    height: 15px;
    cursor: pointer;
    accent-color: var(--primary-color);
    flex-shrink: 0;
    margin-top: 0.15rem;
}

.send-big-btn {
    width: 100%;
    padding: 1rem;
    font-size: 1.1rem;
    background: var(--success-color);
    color: white;
    border: none;
    border-radius: 12px;
    font-weight: 700;
    cursor: pointer;
    box-shadow: 0 4px 10px rgba(39, 174, 96, 0.3);
    transition: transform 0.1s;
}
.send-big-btn:hover:not(:disabled) {
    transform: translateY(-2px);
}
.send-big-btn:disabled {
    background: var(--border-color);
    box-shadow: none;
    cursor: not-allowed;
    color: var(--secondary-text-color);
}

@media (max-width: 768px) {
  .main-card {
    width: 100%;
    max-width: 100%;
    padding: 1.5rem 1rem;
  }

  .p2p-container {
    padding: 1rem;
  }

  .friends-grid {
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .drop-area {
    padding: 1.5rem 1rem;
  }
}

@media (max-width: 768px) {
  .dashboard-container {
    padding-bottom: 64px; /* space for bottom nav */
    flex-direction: column;
  }

  .main-content {
    overflow-y: auto;
    border-radius: 0;
  }

  .p2p-page {
    min-height: unset;
  }

  .p2p-container {
    overflow-y: unset;
  }
}

@media (max-width: 480px) {
  .p2p-page {
    padding-bottom: 0;
  }

  .card-header h2 {
    font-size: 1.2rem;
  }

  .friend-chip {
    font-size: 0.8rem;
    padding: 0.3rem 0.6rem;
  }
}

.recipient-options {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
  margin-top: 0.5rem;
}

.friends-page-link {
  display: inline-block;
  margin-top: 0.5rem;
  font-size: 0.85rem;
  color: var(--primary-color);
  text-decoration: none;
}

.friends-page-link:hover {
  text-decoration: underline;
}

.invite-option-chip {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.4rem 0.8rem;
  background: var(--hover-background-color);
  border: 1px dashed var(--border-color);
  border-radius: 20px;
  font-size: 0.85rem;
  color: var(--secondary-text-color);
  cursor: pointer;
  transition: all 0.2s;
}
.invite-option-chip:hover { border-color: var(--primary-color); color: var(--primary-color); }
.invite-option-chip.selected {
  background: var(--primary-color);
  color: #fff;
  border-style: solid;
  border-color: var(--primary-color);
}

.invite-mode-display {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.close-mode-btn {
  background: transparent;
  border: none;
  color: var(--secondary-text-color);
  font-size: 1.2rem;
  cursor: pointer;
  padding: 0 0.2rem;
  line-height: 1;
}

.invite-send-btn {
  background: var(--primary-color) !important;
  box-shadow: 0 4px 10px rgba(52, 152, 219, 0.3) !important;
}

.invite-banner-wrap {
  padding: 0.75rem 1rem 0;
  display: flex;
  justify-content: center;
}

.invite-banner {
  width: 600px;
  max-width: 100%;
  background: var(--card-color);
  border: 1px solid var(--primary-color);
  border-radius: 12px;
  padding: 1rem 1.2rem;
}

.invite-banner-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  flex-wrap: wrap;
}

.invite-banner-info {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  font-size: 0.9rem;
  color: var(--text-color);
}

.invite-banner-actions {
  display: flex;
  gap: 0.5rem;
  flex-shrink: 0;
}

.invite-btn-accept {
  padding: 0.4rem 1rem;
  background: var(--primary-color);
  color: #fff;
  border: none;
  border-radius: 8px;
  font-size: 0.85rem;
  cursor: pointer;
}
.invite-btn-accept:disabled { opacity: 0.6; cursor: not-allowed; }

.invite-btn-decline {
  padding: 0.4rem 1rem;
  background: transparent;
  border: 1px solid var(--border-color);
  color: var(--text-color);
  border-radius: 8px;
  font-size: 0.85rem;
  cursor: pointer;
}
.invite-btn-decline:disabled { opacity: 0.6; cursor: not-allowed; }

.invite-banner-text {
  font-size: 0.9rem;
  color: var(--text-secondary);
}

.invite-banner-error {
  font-size: 0.9rem;
  color: var(--error-color, #ef4444);
}
</style>
