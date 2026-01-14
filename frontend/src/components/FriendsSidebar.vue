<template>
  <div class="friends-preview-list">
      <div v-if="friendStore.loading" class="loading">Chargement...</div>
      <div v-else-if="friendsCount === 0" class="empty-state">
        <p class="empty-text">Aucun ami</p>
      </div>
      <div v-else class="friends-list-compact">
        <div v-for="friend in friendStore.acceptedFriends" :key="friend.id" class="friend-item-compact">
          <div class="avatar-compact">
              {{ getInitials(friend.name) }}
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
      
      <DeleteConfirmDialog 
        :isOpen="showDeleteConfirm" 
        :title="'Suppression ami'"
        :message="friendToDelete ? `Voulez-vous vraiment supprimer ${friendToDelete.name} de votre liste d'amis ?` : ''"
        @confirm="confirmDelete"
        @cancel="cancelDelete" 
      />
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useFriendStore } from '../stores/friends'
import DeleteConfirmDialog from './DeleteConfirmDialog.vue'

const authStore = useAuthStore()
const friendStore = useFriendStore()

const showDeleteConfirm = ref(false)
const friendToDelete = ref(null)

onMounted(() => {
    friendStore.fetchFriends();
})

const friendsCount = computed(() => friendStore.acceptedFriends.length)

const getInitials = (name) => {
  if (!name) return '?'
  return name.substring(0, 2).toUpperCase()
}

const deleteFriend = (friend) => {
    friendToDelete.value = friend
    showDeleteConfirm.value = true
}

const confirmDelete = async () => {
    if (friendToDelete.value) {
        await friendStore.removeFriend(friendToDelete.value.id)
    }
    cancelDelete()
}

const cancelDelete = () => {
    showDeleteConfirm.value = false
    friendToDelete.value = null
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
  background-color: #f1f3f4; 
  color: #5f6368;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.85rem;
  font-weight: 500;
  position: relative;
  flex-shrink: 0;
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
