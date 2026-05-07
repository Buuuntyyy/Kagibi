<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="p2p-subdomain">
    <P2PNav />

    <main class="p2p-main">

      <!-- Guest mode overlay — shown when arriving via an invite link -->
      <div v-if="isGuestMode" class="guest-overlay">
        <div class="guest-card">

          <!-- Consent screen — shown before auth starts -->
          <template v-if="guestState === 'consent'">
            <div class="guest-consent-header">
              <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"/><polyline points="13 2 13 9 20 9"/></svg>
              <div>
                <p class="guest-consent-title">Réception d'un fichier</p>
                <p class="guest-consent-subtitle">Quelqu'un souhaite vous envoyer un fichier via Kagibi P2P</p>
              </div>
            </div>

            <div class="guest-consent-info">
              <p>Le fichier sera transféré <strong>directement</strong> entre les appareils, chiffré de bout en bout. Kagibi n'a aucun accès au contenu.</p>
            </div>

            <label class="guest-consent-check">
              <input type="checkbox" v-model="guestLegalConsent" />
              <span>
                Je confirme que la réception de ce fichier est légale et j'accepte les
                <a href="/terms" target="_blank" class="guest-legal-link">Conditions d'Utilisation</a>.
              </span>
            </label>

            <button class="guest-accept-btn" :disabled="!guestLegalConsent" @click="startGuestAuth">
              Accepter et télécharger
            </button>
            <button class="guest-decline-btn" @click="isGuestMode = false">
              Refuser
            </button>

            <p class="guest-privacy-note">
              <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
              End-to-end encrypted · No account required · Powered by Kagibi
            </p>
          </template>

          <!-- Transfer in progress -->
          <template v-else>
            <!-- File info header -->
            <div v-if="guestInviteInfo" class="guest-file-info">
              <svg xmlns="http://www.w3.org/2000/svg" width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"/><polyline points="13 2 13 9 20 9"/></svg>
              <div>
                <p class="guest-filename">{{ guestInviteInfo.file_name }}</p>
                <p class="guest-filesender">{{ t('p2p.invite.incomingDesc', { sender: guestInviteInfo.sender_name, fileName: '', size: formatSize(guestInviteInfo.file_size) }).trim() }}</p>
              </div>
            </div>

            <!-- State indicator -->
            <div class="guest-state" :class="{ 'is-error': guestState === 'error', 'is-done': guestState === 'done' }">
              <div v-if="guestState !== 'error' && guestState !== 'done'" class="guest-spinner"></div>
              <svg v-else-if="guestState === 'done'" xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"/></svg>
              <svg v-else xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg>
              <span>{{ guestStateLabel }}</span>
            </div>

            <!-- Progress bar while transferring -->
            <div v-if="guestState === 'transferring' && p2pStore.activeTransfer" class="guest-progress-wrap">
              <div class="guest-progress-track">
                <div class="guest-progress-fill" :style="{ width: p2pStore.activeTransfer.progress + '%' }"></div>
              </div>
              <span class="guest-pct">{{ p2pStore.activeTransfer.progress }}%</span>
            </div>

            <!-- Manual leave button once transfer is complete -->
            <button v-if="guestState === 'done'" @click="guestLeave" class="btn btn-secondary guest-done-btn">
              {{ t('p2p.invite.guest.leave') }}
            </button>

            <p class="guest-privacy-note">
              <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
              End-to-end encrypted · No account required · Powered by Kagibi
            </p>
          </template>

        </div>
      </div>

      <div v-else class="p2p-container">
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

                <!-- Selection not made yet: show friends + invite option -->
                <div v-if="!selectedFriend && !inviteMode">
                  <p class="secondary-text">{{ t('p2p.selectFriendDesc') }}</p>
                  <div class="recipient-options">
                    <!-- Online friends -->
                    <div class="friends-grid">
                      <div v-if="onlineFriends.length === 0" class="no-friends">
                        {{ t('p2p.noOnlineFriends') }}
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

                    <!-- Invite option -->
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
    </main>

    <P2PTransferDialog />
    <P2PInviteDialog
      :visible="showInviteDialog"
      :file="selectedFile"
      :transfer-id="inviteTransferId"
      @close="showInviteDialog = false"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useFriendStore } from '../../stores/friends'
import { useP2PStore } from '../../stores/p2p'
import { useRealtimeStore } from '../../stores/realtime'
import { useAuthStore } from '../../stores/auth'
import { authClient } from '../../auth-client'
import { API_BASE_URL } from '../../api'
import { generateRSAKeyPair, exportKeyToPEM } from '../../utils/crypto'
import P2PNav from '../../components/landing/P2PNav.vue'
import P2PTransferDialog from '../../components/P2PTransferDialog.vue'
import P2PInviteDialog from '../../components/p2p/P2PInviteDialog.vue'

const { t } = useI18n()
const route = useRoute()
const friendStore = useFriendStore()
const p2pStore = useP2PStore()
const realtimeStore = useRealtimeStore()
const authStore = useAuthStore()

const selectedFriend = ref(null)
const selectedFile = ref(null)
const inviteMode = ref(false)
const showInviteDialog = ref(false)
const inviteTransferId = ref('')

// Guest flow state
const isGuestMode = ref(false)
const guestState = ref('') // 'consent' | 'authenticating' | 'generating-keys' | 'connecting' | 'waiting' | 'transferring' | 'done' | 'error'
const guestError = ref('')
const guestInviteInfo = ref(null) // { sender_name, file_name, file_size }
const guestLegalConsent = ref(false)
let pendingGuestToken = ''

// Non-guest invite acceptance state (kept for backward compat)
const incomingInvite = ref(null)
const inviteError = ref('')
const inviteAccepted = ref(false)
const inviteLoading = ref(false)

const onlineFriends = computed(() => {
  return friendStore.acceptedFriends.filter(f => f.online)
})

const canSend = computed(() => !!selectedFriend.value && !!selectedFile.value && !inviteMode.value)

const guestStateLabel = computed(() => {
  switch (guestState.value) {
    case 'consent':           return ''
    case 'authenticating':    return t('p2p.invite.guest.authenticating')
    case 'generating-keys':   return t('p2p.invite.guest.generatingKeys')
    case 'connecting':        return t('p2p.invite.guest.connecting')
    case 'waiting':           return t('p2p.invite.guest.waiting')
    case 'transferring':      return t('p2p.invite.guest.transferring')
    case 'done':              return t('p2p.invite.guest.done')
    case 'error':             return guestError.value || t('p2p.invite.guest.error')
    default:                  return t('common.loading')
  }
})

onMounted(async () => {
  const inviteToken = route.query.invite
  if (inviteToken) {
    isGuestMode.value = true
    pendingGuestToken = inviteToken
    guestState.value = 'consent'
  } else {
    friendStore.fetchFriends()
  }
})

async function startGuestAuth() {
  if (!guestLegalConsent.value || !pendingGuestToken) return
  await guestAutoAuth(pendingGuestToken)
}

// Close invite dialog automatically when the recipient accepted and transfer started
watch(() => p2pStore.inviteReady, (ready) => {
  if (ready && showInviteDialog.value) showInviteDialog.value = false
})

// Watch for active transfer starting (guest flow: auto-accept offer)
watch(() => p2pStore.incomingOffer, async (offer) => {
  if (!isGuestMode.value || !offer) return
  guestState.value = 'transferring'
  await p2pStore.acceptTransfer()
})

// Watch for transfer completing (guest flow)
watch(() => p2pStore.activeTransfer?.status, (status) => {
  if (!isGuestMode.value) return
  if (status === 'Complete' || status === 'Done') {
    guestState.value = 'done'
  }
})

function guestLeave() {
  authClient.clearGuestToken()
  window.location.href = 'https://kagibi.cloud'
}

async function guestAutoAuth(inviteToken) {
  guestError.value = ''
  try {
    // Step 1: Get guest JWT
    guestState.value = 'authenticating'
    const authRes = await fetch(`${API_BASE_URL}p2p/guest-auth`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ token: inviteToken }),
    })
    let authData
    try { authData = await authRes.json() } catch { authData = {} }
    if (!authRes.ok) {
      guestError.value = authRes.status === 410 ? t('p2p.invite.expired') : (authData.error || t('p2p.invite.guest.error'))
      guestState.value = 'error'
      return
    }
    authClient.setGuestToken(authData.jwt)
    guestInviteInfo.value = {
      sender_name: authData.sender_name,
      file_name: authData.file_name,
      file_size: authData.file_size,
    }

    // Step 2: Generate RSA keypair for this session
    guestState.value = 'generating-keys'
    const keyPair = await generateRSAKeyPair()
    const publicKeyPEM = await exportKeyToPEM(keyPair.publicKey, 'spki')
    // Inject private key into auth store so p2pStore.acceptTransfer() can decrypt the file key
    authStore.privateKey = keyPair.privateKey

    // Step 3: Connect WebSocket with guest JWT
    guestState.value = 'connecting'
    await realtimeStore.connect()

    // Step 4: Accept the invite, send public key
    const acceptRes = await fetch(`${API_BASE_URL}p2p/invite/${inviteToken}/accept`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${authData.jwt}`,
      },
      body: JSON.stringify({ public_key: publicKeyPEM }),
    })
    let acceptData
    try { acceptData = await acceptRes.json() } catch { acceptData = {} }
    if (!acceptRes.ok) {
      guestError.value = acceptData.error || t('p2p.invite.guest.error')
      guestState.value = 'error'
      return
    }

    // Step 5: Wait for sender to start the transfer (handled by watch above)
    guestState.value = 'waiting'
  } catch (e) {
    console.error('[Guest] Auto-auth failed:', e)
    guestError.value = e.message || t('p2p.invite.guest.error')
    guestState.value = 'error'
  }
}

async function openInviteDialog() {
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
    console.error('Transfer failed', e)
  }
}
</script>

<style scoped>
.p2p-subdomain {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: var(--background-color);
}

.p2p-main {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.p2p-container {
  flex: 1;
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding: 3rem 1rem;
}

.main-card {
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 16px;
  width: 600px;
  max-width: 100%;
  padding: 2.5rem;
  box-shadow: 0 4px 20px rgba(0,0,0,0.04);
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
.friend-chip:hover { border-color: var(--primary-color); }
.friend-chip.selected { background: var(--primary-color); color: white; }

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
  width: 8px; height: 8px;
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
.file-meta { flex: 1; display: flex; flex-direction: column; overflow: hidden; min-width: 0; }
.fname { font-weight: 600; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.fsize { font-size: 0.85rem; color: var(--secondary-text-color); }
.change-file-btn { font-size: 0.8rem; padding: 0.2rem 0.6rem; }
.close-btn { background: transparent; border: none; color: white; font-size: 1.2rem; margin-left: 0.5rem; cursor: pointer; }

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
.send-big-btn:hover:not(:disabled) { transform: translateY(-2px); }
.send-big-btn:disabled {
  background: var(--border-color);
  box-shadow: none;
  cursor: not-allowed;
  color: var(--secondary-text-color);
}

.recipient-options {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
  margin-top: 0.5rem;
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
  color: var(--text-secondary, var(--secondary-text-color));
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
  padding: 0.75rem 1rem;
}

.invite-banner {
  max-width: 600px;
  margin: 0 auto;
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

.guest-overlay {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem 1rem;
}

.guest-card {
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 16px;
  padding: 2.5rem 2rem;
  width: 440px;
  max-width: 100%;
  box-shadow: 0 4px 20px rgba(0,0,0,0.06);
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.guest-file-info {
  display: flex;
  align-items: center;
  gap: 1rem;
  color: var(--text-color);
}

.guest-filename {
  margin: 0;
  font-weight: 600;
  font-size: 1rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 300px;
}

.guest-filesender {
  margin: 0.2rem 0 0;
  font-size: 0.85rem;
  color: var(--text-secondary, var(--secondary-text-color));
}

.guest-state {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-size: 0.95rem;
  color: var(--text-color);
  background: var(--hover-background-color, var(--background-color));
  border: 1px solid var(--border-color);
  border-radius: 10px;
  padding: 0.9rem 1rem;
}

.guest-state.is-error {
  color: var(--error-color, #ef4444);
  border-color: var(--error-color, #ef4444);
}

.guest-state.is-done {
  color: var(--success-color, #27ae60);
  border-color: var(--success-color, #27ae60);
}

.guest-spinner {
  width: 18px;
  height: 18px;
  border: 2.5px solid var(--border-color);
  border-top-color: var(--primary-color);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  flex-shrink: 0;
}

.guest-progress-wrap {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.guest-progress-track {
  flex: 1;
  height: 6px;
  background: var(--border-color);
  border-radius: 3px;
  overflow: hidden;
}

.guest-progress-fill {
  height: 100%;
  background: var(--primary-color);
  border-radius: 3px;
  transition: width 0.3s ease;
}

.guest-pct {
  font-size: 0.85rem;
  color: var(--text-secondary, var(--secondary-text-color));
  min-width: 36px;
  text-align: right;
}

.guest-done-btn {
  width: 100%;
  margin-top: 0.5rem;
}

.guest-privacy-note {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.72rem;
  color: var(--text-secondary, var(--secondary-text-color));
  margin: 0;
  opacity: 0.7;
}

.guest-consent-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  color: var(--text-color);
}

.guest-consent-title {
  margin: 0;
  font-weight: 600;
  font-size: 1rem;
}

.guest-consent-subtitle {
  margin: 0.2rem 0 0;
  font-size: 0.85rem;
  color: var(--text-secondary, var(--secondary-text-color));
}

.guest-consent-info {
  font-size: 0.88rem;
  color: var(--text-secondary, var(--secondary-text-color));
  background: var(--hover-background-color, var(--background-color));
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 0.75rem 1rem;
  line-height: 1.5;
}

.guest-consent-check {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  font-size: 0.82rem;
  color: var(--text-secondary, var(--secondary-text-color));
  cursor: pointer;
  user-select: none;
  line-height: 1.4;
}

.guest-consent-check input[type="checkbox"] {
  width: 15px;
  height: 15px;
  flex-shrink: 0;
  margin-top: 0.15rem;
  cursor: pointer;
  accent-color: var(--primary-color);
}

.guest-legal-link {
  color: var(--primary-color);
  text-decoration: underline;
}

.guest-accept-btn {
  width: 100%;
  padding: 0.75rem;
  background: var(--primary-color);
  color: #fff;
  border: none;
  border-radius: 10px;
  font-size: 0.95rem;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.2s;
}

.guest-accept-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.guest-decline-btn {
  width: 100%;
  padding: 0.6rem;
  background: transparent;
  border: 1px solid var(--border-color);
  color: var(--text-secondary, var(--secondary-text-color));
  border-radius: 10px;
  font-size: 0.88rem;
  cursor: pointer;
}

@keyframes spin { to { transform: rotate(360deg); } }

@media (max-width: 768px) {
  .p2p-subdomain {
    height: 100dvh;
    height: 100svh; /* fallback for older Safari */
    overflow-y: auto;
    -webkit-overflow-scrolling: touch;
  }
  .p2p-main {
    overflow-y: auto;
    -webkit-overflow-scrolling: touch;
    flex: none;
    min-height: 0;
  }
  .p2p-container {
    padding: 1rem 1rem 2rem;
    align-items: flex-start;
    flex: none;
  }
  .main-card { padding: 1.5rem 1rem; }
  .invite-banner-content { flex-direction: column; align-items: flex-start; }
  .guest-card { padding: 1.5rem 1rem; }
}
</style>
