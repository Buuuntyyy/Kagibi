<template>
  <div class="friends-preview-list">
      <div v-if="friendStore.loading" class="loading">Chargement...</div>
      <div v-else-if="!hasFriendsOrRequests" class="empty-state">
        <p class="empty-text">Aucun ami</p>
      </div>
      <div v-else class="friends-list-compact">
        <!-- PENDING REQUESTS RECEIVED -->
        <div v-if="friendStore.pendingReceived.length > 0" class="pending-section">
           <div class="section-label">DEMANDES REÇUES ({{ friendStore.pendingReceived.length }})</div>
           <div v-for="req in friendStore.pendingReceived" :key="req.id" class="pending-item">
              <div class="pending-header">
                  <div class="avatar-compact small">
                      <img 
                        v-if="req.avatar_url" 
                        :src="req.avatar_url" 
                        :alt="req.name"
                        class="avatar-image"
                        @error="(e) => e.target.style.display = 'none'"
                      />
                      <span v-else class="avatar-initials">{{ getInitials(req.name) }}</span>
                  </div>
                  <div class="pending-info">
                      <span class="pending-name">{{ req.name }}</span>
                      <span class="pending-sub">Vous demande en ami</span>
                  </div>
              </div>
              <div class="pending-actions">
                  <button @click="friendStore.acceptRequest(req.requestId)" class="btn-xs accept">Accepter</button>
                  <button @click="friendStore.rejectRequest(req.requestId)" class="btn-xs reject">Refuser</button>
              </div>
           </div>
           <div class="divider-mini"></div>
        </div>

        <!-- PENDING REQUESTS SENT -->
        <div v-if="friendStore.pendingSent.length > 0" class="pending-section">
           <div class="section-label">DEMANDES ENVOYÉES ({{ friendStore.pendingSent.length }})</div>
           <div v-for="req in friendStore.pendingSent" :key="req.id" class="pending-item sent">
              <div class="pending-header">
                  <div class="avatar-compact small">
                      <img 
                        v-if="req.avatar_url" 
                        :src="req.avatar_url" 
                        :alt="req.name"
                        class="avatar-image"
                        @error="(e) => e.target.style.display = 'none'"
                      />
                      <span v-else class="avatar-initials">{{ getInitials(req.name) }}</span>
                  </div>
                  <div class="pending-info">
                      <span class="pending-name">{{ req.name }}</span>
                      <span class="pending-sub">En attente de réponse</span>
                  </div>
              </div>
              <div class="pending-status">
                  <span class="status-badge">En attente</span>
              </div>
           </div>
           <div class="divider-mini"></div>
        </div>

        <!-- ACCEPTED FRIENDS -->
        <div v-for="friend in friendStore.acceptedFriends" :key="friend.id" class="friend-item-compact">
          <div class="avatar-compact">
              <img 
                v-if="friend.avatar_url" 
                :src="friend.avatar_url" 
                :alt="friend.name"
                class="avatar-image"
                @error="(e) => e.target.style.display = 'none'"
              />
              <span v-else class="avatar-initials">{{ getInitials(friend.name) }}</span>
              <span v-if="friend.online" class="status-dot"></span>
          </div>
          <span class="name-compact">{{ friend.name }}</span>
          <button class="delete-friend-btn" @click.stop="deleteFriend(friend)" title="Supprimer cet ami">
             <svg viewBox="0 0 24 24" class="cross-icon" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <line x1="18" y1="6" x2="6" y2="18"></line>
                <line x1="6" y1="6" x2="18" y2="18"></line>
             </svg>
          </button>
        </div>
      </div>
  </div>
</template>

<script setup>
import { computed, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useFriendStore } from '../stores/friends'
import { useUIStore } from '../stores/ui'

const authStore = useAuthStore()
const friendStore = useFriendStore()
const uiStore = useUIStore()

onMounted(() => {
    friendStore.fetchFriends();
})

const hasFriendsOrRequests = computed(() => {
    return friendStore.acceptedFriends.length > 0 || 
           friendStore.pendingReceived.length > 0 || 
           friendStore.pendingSent.length > 0
})

const getInitials = (name) => {
  if (!name) return '?'
  return name.substring(0, 2).toUpperCase()
}

const deleteFriend = (friend) => {
    uiStore.requestDeleteConfirmation({
        title: 'Suppression ami',
        message: `Voulez-vous vraiment supprimer ${friend.name} de votre liste d'amis ?`,
        onConfirm: async () => {
             await friendStore.removeFriend(friend.id)
        }
    })
}
</script>

<style scoped>
.friends-preview-list {
  padding: 8px 16px;
  max-height: 300px;
  overflow-y: auto;
  scrollbar-width: thin;
}

.friends-list-compact {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.pending-section {
    margin-bottom: 8px;
    display: flex;
    flex-direction: column;
    gap: 6px;
}

.section-label {
    font-size: 0.7rem;
    font-weight: 700;
    color: var(--secondary-text-color);
    padding: 0 4px;
    margin-top: 4px;
}

.pending-item {
    background-color: var(--card-color);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    padding: 8px;
    display: flex;
    flex-direction: column;
    gap: 8px;
}

.pending-header {
    display: flex;
    align-items: center;
    gap: 8px;
}

.avatar-compact.small {
    width: 24px;
    height: 24px;
    font-size: 0.75rem;
}

.pending-info {
    display: flex;
    flex-direction: column;
    overflow: hidden;
}

.pending-name {
    font-size: 0.85rem;
    font-weight: 500;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.pending-sub {
    font-size: 0.7rem;
    color: var(--secondary-text-color);
}

.pending-actions {
    display: flex;
    gap: 6px;
}

.btn-xs {
    flex: 1;
    border: none;
    border-radius: 4px;
    font-size: 0.75rem;
    padding: 4px 0;
    cursor: pointer;
    font-weight: 500;
}

.btn-xs.accept {
    background-color: var(--primary-color);
    color: white;
}

.btn-xs.reject {
    background-color: #f1f3f4;
    color: #333;
}

.pending-item.sent {
    background-color: #fffbf0;
    border-color: #ffd966;
}

.pending-status {
    display: flex;
    justify-content: flex-end;
}

.status-badge {
    background-color: #ffd966;
    color: #333;
    padding: 4px 8px;
    border-radius: 4px;
    font-size: 0.7rem;
    font-weight: 500;
}

.divider-mini {
    height: 1px;
    background-color: var(--border-color);
    margin: 4px 8px;
}

.friend-item-compact {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 12px;
  border-radius: 20px;
  font-size: 0.9rem;
  color: var(--main-text-color);
  cursor: pointer;
  transition: background-color 0.2s;
}

.friend-item-compact:hover {
  background-color: var(--hover-background-color);
}

.friend-item-compact:hover .delete-friend-btn {
  opacity: 1;
}

.delete-friend-btn {
  opacity: 0;
  background: none;
  border: none;
  cursor: pointer;
  margin-left: auto;
  padding: 4px;
  color: #000;
  display: flex;
  align-items: center;
  border-radius: 50%;
  transition: opacity 0.2s, background-color 0.2s;
}

.delete-friend-btn:hover {
  background-color: rgba(0, 0, 0, 0.1);
}

.cross-icon {
    width: 16px;
    height: 16px;
}

.avatar-compact {
  width: 28px;
  height: 28px;
  background: linear-gradient(135deg, #f1f3f4 0%, #e1e3e6 100%);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.85rem;
  font-weight: 500;
  position: relative;
  flex-shrink: 0;
  overflow: hidden;
  transition: transform 0.3s ease;
}

.avatar-compact:hover {
  transform: scale(1.05);
}

.avatar-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: opacity 0.3s ease;
}

.avatar-initials {
  color: #5f6368;
  font-size: 0.85rem;
  font-weight: 500;
}

.status-dot {
  position: absolute;
  bottom: -1px;
  right: -1px;
  width: 10px;
  height: 10px;
  background-color: #34a853; 
  border-radius: 50%;
  border: 2px solid white;
}

.name-compact {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 140px;
  font-weight: 400;
}

.empty-state {
  text-align: center;
  padding: 10px;
}

.empty-text {
  font-size: 0.85rem;
  color: var(--secondary-text-color);
}

.loading {
  font-size: 0.85rem;
  color: var(--secondary-text-color);
  padding: 8px;
  text-align: left;
}
</style>
