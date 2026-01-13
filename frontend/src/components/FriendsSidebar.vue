<template>
  <div class="friends-sidebar">
    <div class="sidebar-header">
      <h3>Amis</h3>
      <button class="close-btn" @click="$emit('close')">
        <svg viewBox="0 0 24 24" width="20" height="20" stroke="currentColor" stroke-width="2" fill="none">
          <line x1="18" y1="6" x2="6" y2="18"></line>
          <line x1="6" y1="6" x2="18" y2="18"></line>
        </svg>
      </button>
    </div>

    <div class="sidebar-tabs">
      <button :class="{ active: activeTab === 'list' }" @click="activeTab = 'list'" title="Mes Amis">
        <svg viewBox="0 0 24 24" width="20" height="20" xmlns="http://www.w3.org/2000/svg" fill="none" stroke="currentColor" stroke-width="2">
           <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
           <circle cx="9" cy="7" r="4"></circle>
           <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
           <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
        </svg>
      </button>
      <button :class="{ active: activeTab === 'add' }" @click="activeTab = 'add'" title="Ajouter">
         <svg viewBox="0 0 24 24" width="20" height="20" xmlns="http://www.w3.org/2000/svg" fill="none" stroke="currentColor" stroke-width="2">
             <line x1="12" y1="5" x2="12" y2="19"></line>
             <line x1="5" y1="12" x2="19" y2="12"></line>
         </svg>
      </button>
      <button :class="{ active: activeTab === 'pending' }" @click="activeTab = 'pending'" title="En attente" class="pending-tab">
        <svg viewBox="0 0 24 24" width="20" height="20" xmlns="http://www.w3.org/2000/svg" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="22 17 13.5 8.5 8.5 13.5 2 7"></polyline>
            <polyline points="16 17 22 17 22 11"></polyline>
        </svg>
        <span class="badge-dot" v-if="friendStore.pendingReceived.length > 0"></span>
      </button>
    </div>

    <!-- LIST TAB -->
    <div v-if="activeTab === 'list'" class="tab-content">
      <div v-if="friendStore.loading" class="loading">Chargement...</div>
      <div v-else-if="friendStore.acceptedFriends.length === 0" class="empty-state">
        <p>Aucun ami.</p>
        <button @click="activeTab = 'add'" class="btn-text">Ajouter</button>
      </div>
      <div v-else class="friends-list">
        <div v-for="friend in friendStore.acceptedFriends" :key="friend.id" class="friend-item">
          <div class="friend-avatar">
              {{ getInitials(friend.name) }}
              <span v-if="friend.online" class="status-indicator"></span>
          </div>
          <div class="friend-info">
            <div class="name">{{ friend.name }}</div>
            <div class="email" :title="friend.email">{{ friend.email }}</div>
          </div>
          <div class="actions">
            <button v-if="friend.online" @click="triggerP2PSend(friend)" class="btn-icon p2p-btn" title="Envoyer fichier (P2P)">
                 <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><line x1="22" y1="2" x2="11" y2="13"></line><polygon points="22 2 15 22 11 13 2 9 22 2"></polygon></svg>
            </button>
            <button @click="confirmRemove(friend)" class="btn-icon delete" title="Supprimer">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M18 6L6 18M6 6l12 12"></path>
                </svg>
            </button>
          </div>
        </div>
      </div>
      <input type="file" ref="p2pFileInput" style="display: none" @change="handleP2PFileSelect" />
    </div>

    <!-- ADD TAB -->
    <div v-if="activeTab === 'add'" class="tab-content">
      <div class="add-section">
        <h4>Mon Code</h4>
        <div class="code-box">
          <span>{{ authStore.user?.friend_code || '...' }}</span>
          <button @click="copyCode" class="btn-icon">
             <svg v-if="!copied" viewBox="0 0 24 24" width="16" height="16" stroke="currentColor" fill="none"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path></svg>
             <span v-else>Ok</span>
          </button>
        </div>
        
        <div class="divider"></div>
        
        <h4>Ajouter</h4>
        <form @submit.prevent="submitAddFriend">
          <input v-model="friendCodeInput" type="text" placeholder="Code ami (ex: A1B2)" class="input-sidebar" required />
          <button type="submit" class="btn-full" :disabled="submitting">
            {{ submitting ? '...' : 'Envoyer' }}
          </button>
        </form>
        <div v-if="addMessage" :class="['msg', addMessageType]">{{ addMessage }}</div>
      </div>
    </div>

    <!-- PENDING TAB -->
    <div v-if="activeTab === 'pending'" class="tab-content">
      <div v-if="friendStore.pendingReceived.length > 0">
        <div class="section-label">Reçues</div>
        <div class="pending-list">
           <div v-for="req in friendStore.pendingReceived" :key="req.id" class="req-item col">
             <div class="req-header">
                <div class="friend-avatar small">{{ getInitials(req.name) }}</div>
                <span class="name">{{ req.name }}</span>
             </div>
             <div class="req-actions">
                <button @click="friendStore.acceptRequest(req.requestId)" class="btn-small accept">Accepter</button>
                <button @click="friendStore.rejectRequest(req.requestId)" class="btn-small reject">Refuser</button>
             </div>
           </div>
        </div>
      </div>

      <div v-if="friendStore.pendingSent.length > 0">
         <div class="section-label mt">Envoyées</div>
         <div class="pending-list">
            <div v-for="req in friendStore.pendingSent" :key="req.id" class="req-item">
               <span class="name">{{ req.name }}</span>
               <span class="status">En attente</span>
            </div>
         </div>
      </div>
      
      <div v-if="friendStore.pendingReceived.length === 0 && friendStore.pendingSent.length === 0" class="empty-state">
          Aucune demande.
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useFriendStore } from '../stores/friends'
import { useAuthStore } from '../stores/auth'
import { useP2PStore } from '../stores/p2p'

defineEmits(['close'])

const friendStore = useFriendStore()
const authStore = useAuthStore()
const p2pStore = useP2PStore()

const activeTab = ref('list')
const friendCodeInput = ref('')
const submitting = ref(false)
const addMessage = ref('')
const addMessageType = ref('')
const copied = ref(false)
const p2pFileInput = ref(null)
const selectedFriendForP2P = ref(null)

onMounted(() => {
  friendStore.fetchFriends()
  if (!authStore.user?.friend_code) {
      authStore.fetchUser()
  }
})

const triggerP2PSend = (friend) => {
    selectedFriendForP2P.value = friend;
    p2pFileInput.value.click();
}

const handleP2PFileSelect = (event) => {
    const file = event.target.files[0];
    if (file && selectedFriendForP2P.value) {
        p2pStore.startTransfer(selectedFriendForP2P.value, file);
        // Reset input
        event.target.value = '';
    }
}

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
    addMessage.value = "Envoyé !"
    addMessageType.value = "success"
    friendCodeInput.value = ''
  } catch (err) {
    addMessage.value = "Erreur"
    addMessageType.value = "error"
  } finally {
    submitting.value = false
  }
}

const confirmRemove = (friend) => {
  if (confirm(`Supprimer ${friend.name} ?`)) {
    friendStore.removeFriend(friend.id)
  }
}
</script>

<style scoped>
.friends-sidebar {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--background-color);
  border-right: 1px solid var(--border-color);
  box-sizing: border-box;
  font-family: inherit;
  color: var(--main-text-color);
}

.sidebar-header {
  padding: 1rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid var(--border-color);
}

.sidebar-header h3 {
  margin: 0;
  font-size: 1.1rem;
  color: var(--main-text-color);
}

.close-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--secondary-text-color);
  padding: 4px;
  border-radius: 50%;
  transition: background-color 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.close-btn:hover {
  background-color: var(--hover-background-color);
  color: var(--main-text-color);
}

.sidebar-tabs {
  display: flex;
  border-bottom: 1px solid var(--border-color);
}

.sidebar-tabs button {
  flex: 1;
  padding: 0.8rem;
  background: none;
  border: none;
  border-radius: 0; /* Override global button radius */
  cursor: pointer;
  color: var(--secondary-text-color);
  border-bottom: 2px solid transparent;
  position: relative;
  transition: all 0.2s;
}

.sidebar-tabs button:hover {
    background-color: var(--hover-background-color);
    color: var(--main-text-color);
}

.sidebar-tabs button.active {
  color: var(--primary-color);
  border-bottom-color: var(--primary-color);
  background-color: transparent;
}

.badge-dot {
  position: absolute;
  top: 8px;
  right: 25%;
  width: 8px;
  height: 8px;
  background: var(--primary-color);
  border-radius: 50%;
}

.tab-content {
  flex: 1;
  overflow-y: auto;
  padding: 1rem;
}

/* LIST */
.friends-list {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.friend-item {
  display: flex;
  align-items: center;
  gap: 0.8rem;
  padding: 0.5rem;
  border-radius: 8px;
  transition: background-color 0.2s;
}

.friend-item:hover {
  background: var(--hover-background-color);
}

.friend-avatar {
  background: var(--primary-color);
  color: white;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: 0.9rem;
  flex-shrink: 0;
}

.friend-info {
  flex: 1;
  overflow: hidden;
}

.name {
  font-weight: 500;
  font-size: 0.9rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  color: var(--main-text-color);
}

.email {
  font-size: 0.75rem;
  color: var(--secondary-text-color);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.btn-icon.delete {
    color: var(--error-color);
    opacity: 0;
    background: none;
    border: none;
    cursor: pointer;
    padding: 4px;
    border-radius: 4px;
    transition: opacity 0.2s, background-color 0.2s;
}

.friend-item:hover .btn-icon.delete {
    opacity: 1;
}

.btn-icon.delete:hover {
  background-color: rgba(231, 76, 60, 0.1);
}

/* ADD */
.code-box {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: var(--card-color);
  padding: 0.5rem 0.8rem;
  border-radius: 6px;
  border: 1px solid var(--border-color);
  margin-bottom: 1rem;
  font-family: monospace;
  color: var(--main-text-color);
}

.divider {
  height: 1px;
  background: var(--border-color);
  margin: 1.5rem 0;
}

.input-sidebar {
  width: 100%;
  padding: 0.6rem;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  background: var(--card-color);
  color: var(--main-text-color);
  margin-bottom: 0.5rem;
  box-sizing: border-box;
  font-family: inherit;
}

.input-sidebar:focus {
  outline: 2px solid var(--primary-color);
  border-color: transparent;
}

.btn-full {
  width: 100%;
  padding: 0.6rem;
  background: var(--primary-color);
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  transition: opacity 0.2s;
}

.btn-full:hover {
  opacity: 0.9;
}

.btn-full:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.msg {
  font-size: 0.8rem;
  margin-top: 0.5rem;
}
.msg.success { color: var(--success-color); }
.msg.error { color: var(--error-color); }

/* PENDING */
.section-label {
    font-size: 0.75rem;
    text-transform: uppercase;
    color: var(--secondary-text-color);
    margin-bottom: 0.5rem;
    font-weight: bold;
    letter-spacing: 0.05em;
}

.mt { margin-top: 1.5rem; }

.req-item {
    background: var(--card-color);
    padding: 0.8rem;
    border: 1px solid var(--border-color);
    border-radius: 8px;
    margin-bottom: 0.5rem;
}

.req-item.col {
    display: flex;
    flex-direction: column;
    gap: 0.8rem;
}

.req-header {
    display: flex;
    align-items: center;
    gap: 0.8rem;
}

.req-actions {
    display: flex;
    gap: 0.5rem;
}

.btn-small {
    flex: 1;
    border: none;
    padding: 0.4rem;
    border-radius: 4px;
    font-size: 0.8rem;
    cursor: pointer;
    font-weight: 500;
    transition: opacity 0.2s;
}

.btn-small:hover {
  opacity: 0.9;
}

.btn-small.accept { background: var(--success-color); color: white; }
.btn-small.reject { background: var(--error-color); color: white; }

.empty-state {
    text-align: center;
    color: var(--secondary-text-color);
    padding: 2rem 0;
    font-size: 0.9rem;
}

.btn-text {
  background: none;
  border: none;
  color: var(--primary-color);
  cursor: pointer;
  text-decoration: underline;
  padding: 0;
  font-size: inherit;
}

h4 {
  margin-top: 0;
  margin-bottom: 0.8rem;
  color: var(--main-text-color);
  font-size: 0.95rem;
}

/* Scrollbar styling for sidebar */
.tab-content::-webkit-scrollbar {
  width: 6px;
}
.tab-content::-webkit-scrollbar-thumb {
  background-color: var(--border-color);
  border-radius: 3px;
}
.tab-content::-webkit-scrollbar-track {
  background-color: transparent;
}

.friend-avatar {
    position: relative;
}

.status-indicator {
    position: absolute;
    bottom: -2px;
    right: -2px;
    width: 12px;
    height: 12px;
    background-color: #42b983;
    border-radius: 50%;
    border: 2px solid white;
}

.actions {
    display: flex;
    align-items: center;
    gap: 4px;
}

.p2p-btn {
    color: var(--primary-color);
}

.p2p-btn:hover {
    background-color: rgba(66, 185, 131, 0.1);
}
</style>
