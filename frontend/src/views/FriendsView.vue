<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="friends-container">
    <div class="header">
      <h2>{{ t('nav.friends') }}</h2>
      <div class="tabs">
        <button 
          :class="{ active: activeTab === 'list' }" 
          @click="activeTab = 'list'"
        >
          {{ t('friends.myFriends') }}
          <span class="count" v-if="friendStore.acceptedFriends.length">{{ friendStore.acceptedFriends.length }}</span>
        </button>
        <button 
          :class="{ active: activeTab === 'add' }" 
          @click="activeTab = 'add'"
        >
          {{ t('friends.add') }}
        </button>
        <button 
          :class="{ active: activeTab === 'pending' }" 
          @click="activeTab = 'pending'"
        >
          {{ t('friends.pending') }}
          <span class="count badge" v-if="friendStore.pendingReceived.length > 0">{{ friendStore.pendingReceived.length }}</span>
        </button>
      </div>
    </div>

    <!-- TAB: LIST FRIENDS -->
    <div v-if="activeTab === 'list'" class="tab-content">
      <div v-if="friendStore.loading" class="loading">{{ t('common.loading') }}</div>
      <div v-else-if="friendStore.acceptedFriends.length === 0" class="empty-state">
        <div class="empty-icon">👥</div>
        <h3>{{ t('friends.noFriends') }}</h3>
        <p>{{ t('friends.addFirstFriend') }}</p>
        <button @click="activeTab = 'add'" class="btn-primary">{{ t('friends.addFriend') }}</button>
      </div>
      <div v-else class="friends-grid">
        <div v-for="friend in friendStore.acceptedFriends" :key="friend.id" class="friend-card">
          <div class="friend-avatar-wrapper">
            <div class="friend-avatar">{{ getInitials(friend.name) }}</div>
            <div class="online-indicator" :class="{ online: friend.online }" :title="friend.online ? t('friends.online') : t('friends.offline')"></div>
          </div>
          <div class="friend-info">
            <div class="friend-name">{{ friend.name }}</div>
            <div class="friend-email">{{ friend.email }}</div>
          </div>
          <button @click="confirmRemove(friend)" class="btn-icon delete" :title="t('friends.removeFriend')">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M18 6L6 18M6 6l12 12"></path>
            </svg>
          </button>
        </div>
      </div>
    </div>

    <!-- TAB: ADD FRIEND -->
    <div v-if="activeTab === 'add'" class="tab-content narrow">
      <div class="add-section">
        <h3>{{ t('friends.myCode') }}</h3>
        <div class="my-code-box">
          <span class="code">{{ authStore.user?.friend_code || t('common.loading') }}</span>
          
          <div class="info-container">
             <button class="btn-info" @click.stop="showInfoTooltip = !showInfoTooltip">
                 <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor" class="info-icon">
                    <path d="M11 18h2v-2h-2v2zm1-16C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8zm0-14c-2.21 0-4 1.79-4 4h2c0-1.1.9-2 2-2s2 .9 2 2c0 2-3 1.75-3 5h2c0-2.25 3-2.5 3-5 0-2.21-1.79-4-4-4z"/>
                </svg>
             </button>
             <div v-if="showInfoTooltip" class="info-tooltip">
                <p>Ceci est votre identifiant unique. Partagez-le avec vos connaissances pour qu'ils puissent vous ajouter à leurs amis et partager des fichiers de manière sécurisée.</p>
             </div>
          </div>

          <button @click="copyCode" class="btn-copy">
            {{ copied ? t('common.copied') : t('friends.copyCode') }}
          </button>
        </div>
        <p class="description">Partagez ce code avec un ami pour qu'il puisse vous ajouter.</p>
      </div>

      <div class="divider"></div>

      <div class="add-section">
        <h3>{{ t('friends.addFriend') }}</h3>
        <form @submit.prevent="submitAddFriend" class="add-form">
          <input 
            v-model="friendCodeInput" 
            type="text" 
            :placeholder="t('friends.enterCode')" 
            required 
            class="code-input"
          />
          <button type="submit" class="btn-primary" :disabled="submitting">
            {{ submitting ? t('common.loading') : t('friends.sendRequest') }}
          </button>
        </form>
        <div v-if="addMessage" :class="['message', addMessageType]">{{ addMessage }}</div>
      </div>
    </div>

    <!-- TAB: PENDING -->
    <div v-if="activeTab === 'pending'" class="tab-content">
      
      <!-- Received -->
      <div class="section-title">Reçues ({{ friendStore.pendingReceived.length }})</div>
      <div v-if="friendStore.pendingReceived.length === 0" class="empty-sub">Aucune demande reçue</div>
      <div class="requests-list">
        <div v-for="req in friendStore.pendingReceived" :key="req.id" class="request-item">
          <div class="req-user">
            <div class="friend-avatar small">{{ getInitials(req.name) }}</div>
            <span>{{ req.name }} veut vous ajouter</span>
          </div>
          <div class="req-actions">
            <button @click="friendStore.acceptRequest(req.requestId)" class="btn-small primary">Accepter</button>
            <button @click="friendStore.rejectRequest(req.requestId)" class="btn-small secondary">Refuser</button>
          </div>
        </div>
      </div>

      <!-- Sent -->
      <div class="section-title mt-4">Envoyées ({{ friendStore.pendingSent.length }})</div>
      <div v-if="friendStore.pendingSent.length === 0" class="empty-sub">Aucune demande en attente</div>
      <div class="requests-list">
        <div v-for="req in friendStore.pendingSent" :key="req.id" class="request-item">
           <div class="req-user">
            <div class="friend-avatar small">{{ getInitials(req.name) }}</div>
            <span>En attente de {{ req.name }}</span>
          </div>
           <div class="req-status">Envoyée</div>
        </div>
      </div>

    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useFriendStore } from '../stores/friends'
import { useAuthStore } from '../stores/auth'

const { t } = useI18n()

const friendStore = useFriendStore()
const authStore = useAuthStore()

const activeTab = ref('list')
const friendCodeInput = ref('')
const submitting = ref(false)
const addMessage = ref('')
const addMessageType = ref('')
const copied = ref(false)
const showInfoTooltip = ref(false)

const closeInfoTooltip = () => {
  if(showInfoTooltip.value) showInfoTooltip.value = false;
}

onMounted(() => {
  friendStore.fetchFriends()
  // Ensure we have the latest user data (for friend_code)
  if (!authStore.user?.friend_code) {
      authStore.fetchUser()
  }
  window.addEventListener('click', closeInfoTooltip)
})

onUnmounted(() => {
  window.removeEventListener('click', closeInfoTooltip)
})

const getInitials = (name) => {
  return name ? name.substring(0, 2).toUpperCase() : '?'
}

const copyCode = () => {
  navigator.clipboard.writeText(authStore.user?.friend_code || '')
  copied.value = true
  setTimeout(() => copied.value = false, 2000)
}

const submitAddFriend = async () => {
  submitting.value = true
  addMessage.value = ''
  try {
    await friendStore.sendRequest(friendCodeInput.value)
    addMessage.value = "Demande envoyée avec succès ! Consultez l'onglet 'En attente' pour voir vos demandes."
    addMessageType.value = "success"
    friendCodeInput.value = ''
    // Auto-switch to pending tab after 2 seconds
    setTimeout(() => {
      if (addMessageType.value === 'success') {
        activeTab.value = 'pending'
      }
    }, 2000)
  } catch (err) {
    addMessage.value = err
    addMessageType.value = "error"
  } finally {
    submitting.value = false
  }
}

const confirmRemove = (friend) => {
  if (confirm(`Êtes-vous sûr de vouloir supprimer ${friend.name} de vos amis ?`)) {
    friendStore.removeFriend(friend.id)
  }
}
</script>

<style scoped>
.friends-container {
  padding: 2rem;
  max-width: 1000px;
  margin: 0 auto;
  color: var(--main-text-color);
  height: 100%;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
}

.header {
  margin-bottom: 2rem;
}

.header h2 {
  font-size: 1.8rem;
  margin-bottom: 1.5rem;
}

.tabs {
  display: flex;
  gap: 1rem;
  border-bottom: 1px solid var(--border-color);
}

.tabs button {
  padding: 0.8rem 1.5rem;
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  cursor: pointer;
  font-size: 1rem;
  font-weight: 500;
  color: var(--secondary-text-color);
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.tabs button.active {
  color: var(--primary-color);
  border-bottom-color: var(--primary-color);
}

.tabs button:hover {
  color: var(--main-text-color);
}

.count {
  background: var(--border-color);
  padding: 0.1rem 0.4rem;
  border-radius: 10px;
  font-size: 0.8rem;
}

.count.badge {
  background: var(--primary-color);
  color: white;
}

.tab-content {
  flex: 1;
  overflow-y: auto;
}

.tab-content.narrow {
    max-width: 600px;
    margin: 0 auto;
}

/* List Grid */
.friends-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 1.5rem;
}

.friend-card {
  background: var(--background-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 1.5rem;
  display: flex;
  align-items: center;
  gap: 1rem;
  transition: transform 0.2s, box-shadow 0.2s;
}

.friend-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0,0,0,0.05);
}

.friend-avatar-wrapper {
  position: relative;
  flex-shrink: 0;
}

.friend-avatar {
  background: var(--primary-color);
  color: white;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: 1.2rem;
  flex-shrink: 0;
}

.online-indicator {
  position: absolute;
  bottom: 2px;
  right: 2px;
  width: 14px;
  height: 14px;
  background: #666;
  border: 2px solid var(--background-color);
  border-radius: 50%;
  transition: background 0.3s;
}

.online-indicator.online {
  background: #4ade80;
  box-shadow: 0 0 8px rgba(74, 222, 128, 0.5);
}

.friend-avatar.small {
    width: 36px;
    height: 36px;
    font-size: 0.9rem;
}

.friend-info {
  flex: 1;
  overflow: hidden;
}

.friend-name {
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.friend-email {
  font-size: 0.85rem;
  color: var(--secondary-text-color);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.btn-icon {
  background: none;
  border: none;
  color: var(--secondary-text-color);
  cursor: pointer;
  padding: 0.5rem;
  border-radius: 4px;
}

.btn-icon:hover.delete {
  background-color: rgba(220, 53, 69, 0.1);
  color: #dc3545;
}

/* Empty State */
.empty-state {
  text-align: center;
  padding: 4rem 2rem;
  color: var(--secondary-text-color);
}

.empty-icon {
  font-size: 4rem;
  margin-bottom: 1rem;
  opacity: 0.5;
}

/* Add Friend Section */
.add-section {
  background: var(--background-color);
  padding: 2rem;
  border-radius: 8px;
  border: 1px solid var(--border-color);
  text-align: center;
}

.my-code-box {
  background: var(--card-color);
  border: 2px dashed var(--primary-color);
  padding: 1rem;
  border-radius: 8px;
  margin: 1rem 0;
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 1rem;
}

.code {
  font-family: monospace;
  font-size: 1.5rem;
  letter-spacing: 2px;
  font-weight: bold;
}

.btn-copy {
  border: 1px solid var(--border-color);
  background: white;
  padding: 0.3rem 0.8rem;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.9rem;
}

.description {
  color: var(--secondary-text-color);
  font-size: 0.9rem;
}

.divider {
  height: 1px;
  background: var(--border-color);
  margin: 2rem 0;
}

.add-form {
  display: flex;
  gap: 1rem;
  margin-top: 1.5rem;
}

.code-input {
  flex: 1;
  padding: 0.8rem;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  font-size: 1rem;
}

.btn-primary {
  background: var(--primary-color);
  color: white;
  border: none;
  padding: 0.8rem 1.5rem;
  border-radius: 4px;
  font-weight: bold;
  cursor: pointer;
}

.btn-primary:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.message {
  margin-top: 1rem;
  padding: 0.8rem;
  border-radius: 4px;
  font-size: 0.9rem;
}

.message.success {
  background-color: rgba(66, 185, 131, 0.1);
  color: #2e7d32;
}

.message.error {
  background-color: rgba(220, 53, 69, 0.1);
  color: #c62828;
}

/* Pending Requests */
.section-title {
    font-weight: 600;
    margin-bottom: 1rem;
    color: var(--secondary-text-color);
    font-size: 0.9rem;
    text-transform: uppercase;
    letter-spacing: 0.5px;
}

.mt-4 { margin-top: 2rem; }

.request-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem;
    background: var(--background-color);
    border: 1px solid var(--border-color);
    border-radius: 6px;
    margin-bottom: 0.8rem;
}

.req-user {
    display: flex;
    align-items: center;
    gap: 1rem;
    font-weight: 500;
}

.req-actions {
    display: flex;
    gap: 0.5rem;
}

.btn-small {
    padding: 0.4rem 0.8rem;
    border-radius: 4px;
    font-size: 0.85rem;
    cursor: pointer;
    border: none;
}

.btn-small.primary {
    background: var(--primary-color);
    color: white;
}

.btn-small.secondary {
    background: #e0e0e0;
    color: #333;
}

.req-status {
    font-size: 0.9rem;
    color: var(--secondary-text-color);
    font-style: italic;
}

.empty-sub {
    color: var(--secondary-text-color);
    font-style: italic;
    font-size: 0.9rem;
}

.info-container {
    position: relative;
    display: flex;
    align-items: center;
}

.btn-info {
    background: none;
    border: none;
    cursor: pointer;
    color: var(--secondary-text-color);
    padding: 0;
    display: flex;
    align-items: center;
    transition: color 0.2s;
}

.btn-info:hover {
    color: var(--primary-color);
}

.info-tooltip {
  position: absolute;
  top: 140%; /* Below the icon */
  left: 50%;
  transform: translateX(-50%);
  width: 260px;
  background-color: #333;
  color: white;
  padding: 12px;
  border-radius: 6px;
  font-size: 0.85rem;
  z-index: 100;
  text-align: center;
  box-shadow: 0 4px 12px rgba(0,0,0,0.2);
  line-height: 1.4;
}

.info-tooltip::after {
  content: "";
  position: absolute;
  bottom: 100%;
  left: 50%;
  margin-left: -6px;
  border-width: 6px;
  border-style: solid;
  border-color: transparent transparent #333 transparent;
}
</style>
