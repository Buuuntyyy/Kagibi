<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="close">
    <div class="modal-content">
      <div class="modal-header">
        <h3>Partager "{{ item?.Name || item?.name }}"</h3>
        <button @click="close" class="btn-close">×</button>
      </div>

      <div class="modal-tabs">
        <button 
          :class="['tab-btn', { active: activeTab === 'link' }]"
          @click="activeTab = 'link'"
        >
          Lien Public
        </button>
        <button 
          :class="['tab-btn', { active: activeTab === 'friends' }]"
          @click="activeTab = 'friends'"
        >
          Amis (Sécurisé)
        </button>
      </div>
      
      <div class="modal-body">
        
        <!-- === LINK SHARE TAB === -->
        <div v-if="activeTab === 'link'">
            <!-- Loading State -->
            <div v-if="loading" class="loading-state">
                <div class="spinner"></div> Traitement en cours...
            </div>

            <!-- Not Shared State -->
            <div v-else-if="!isShared" class="not-shared-state">
                <div class="illustration">
                    🔗
                </div>
                <p>Ce {{ item?.type === 'folder' ? 'dossier' : 'fichier' }} n'est pas encore partagé par lien.</p>
                <p class="sub-text">Créez un lien pour le partager avec d'autres personnes.</p>
                
                <div class="form-group">
                    <label for="expiresAt">Expiration (optionnel)</label>
                    <input type="datetime-local" id="expiresAt" v-model="expiresAt" class="form-control" />
                </div>

                <button @click="createShare" class="btn-primary">Créer un lien de partage</button>
            </div>

            <!-- Shared State -->
            <div v-else class="shared-state">
                <div class="link-section">
                    <label>Lien de partage</label>
                    <div class="link-container">
                        <input type="text" :value="shareUrl" readonly ref="shareLinkInput" @click="selectAll" />
                        <button @click="copyLink" class="btn-copy" :class="{ copied: linkCopied }">
                            {{ linkCopied ? 'Copié !' : 'Copier' }}
                        </button>
                    </div>
                </div>
                
                <div class="share-info">
                    <p v-if="localExpiresAt">⏳ Ce lien expirera le : <b>{{ formattedExpiration }}</b></p>
                    <p>⚠️ Toute personne disposant de ce lien pourra accéder au contenu <b>déchiffré</b>.</p>
                </div>
            </div>
        </div>

        <!-- === FRIENDS SHARE TAB === -->
        <div v-else-if="activeTab === 'friends'" class="friends-tab">
            
            <div v-if="!item?.encrypted_key && (item?.type === 'folder' || item?.is_dir)" class="warning-box">
                <strong>Partiel :</strong> Le partage de dossier partage l'accès mais pas les clés des fichiers internes pour le moment.
            </div>

            <div v-if="friends.length === 0" class="empty-friends">
                Vous n'avez pas encore d'amis. 
                <br>
                <router-link to="/friends">Ajouter des amis</router-link>
            </div>

            <div v-else class="friends-list">
                 <div v-for="friend in friends" :key="friend.id" class="friend-item">
                    <div class="friend-info">
                       <div class="friend-avatar">
                         {{ friend.name.charAt(0).toUpperCase() }}
                       </div>
                       <div>
                         <p class="friend-name">{{ friend.name }}</p>
                         <p class="friend-email">{{ friend.email }}</p>
                       </div>
                    </div>

                    <div v-if="!friend.public_key" class="key-missing" title="Clé manquante">
                       ⚠️ Pas de clé
                    </div>

                    <button v-else 
                        @click="shareWithFriend(friend)"
                        :disabled="sharing[friend.id]"
                        class="btn-sm"
                        :class="[
                          isFriendShared(friend.id) 
                            ? 'btn-success' 
                            : 'btn-outline'
                        ]">
                        <span v-if="sharing[friend.id]">...</span>
                        <span v-else-if="isFriendShared(friend.id)">Partagé</span>
                        <span v-else>Envoyer</span>
                    </button>
                 </div>
            </div>
        </div>

      </div>

      <div class="modal-footer">
        <button v-if="activeTab === 'link' && isShared" @click="deleteShare" class="btn-delete">Arrêter le lien</button>
        <button @click="close" class="btn-secondary">Fermer</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue';
import { useFileStore } from '../stores/files';
import { useFriendStore } from '../stores/friends';
import { useAuthStore } from '../stores/auth';
import api from '../api';
import { decryptKeyWithPrivateKey, importKeyFromPEM, encryptKeyWithPublicKey } from '../utils/crypto';
import sodium from 'libsodium-wrappers-sumo';

const props = defineProps({
  isOpen: Boolean,
  item: Object,
});

const emit = defineEmits(['close', 'share-deleted', 'share-created']);
const fileStore = useFileStore();
const friendStore = useFriendStore();
const authStore = useAuthStore();

// UI State
const activeTab = ref('link'); // 'link' or 'friends'
const loading = ref(false);

// Link Share State
const linkCopied = ref(false);
const shareLinkInput = ref(null);
const expiresAt = ref(null);

// Friends Share State
const sharing = ref({}); // Map [friendId] -> boolean (loading state)
const sharedStatus = ref({}); // Map [friendId] -> boolean (success state)
const friends = computed(() => friendStore.acceptedFriends);

// Local state to handle immediate updates without waiting for parent refresh
const localShareToken = ref(null);
const localShareId = ref(null);
const localExpiresAt = ref(null);

// Reset local state when item changes
watch(() => props.item, (newItem) => {
    if (newItem) {
        localShareToken.value = newItem.share_token || newItem.ShareToken;
        localShareId.value = newItem.share_id || newItem.ShareID;
        localExpiresAt.value = newItem.expires_at || newItem.ExpiresAt; // Populate expiration
        
        // Reset tab
        activeTab.value = 'link';
        sharedStatus.value = {};
    }
}, { immediate: true });

onMounted(() => {
    if (friends.value.length === 0) {
        friendStore.fetchFriends();
    }
});

const isShared = computed(() => !!localShareToken.value);

const formattedExpiration = computed(() => {
  if (!localExpiresAt.value) return null;
  return new Date(localExpiresAt.value).toLocaleString();
});

const shareUrl = computed(() => {
  if (localShareToken.value) {
    return `${window.location.origin}/s/${localShareToken.value}`;
  }
  return '';
});

const selectAll = (e) => {
    e.target.select();
}

// --- Link Sharing Methods ---

const createShare = async () => {
    if (!props.item) return;
    loading.value = true;
    try {
        const itemId = props.item.ID || props.item.id;
        
        // Convert expiresAt to ISO string if present
        let expirationDate = null;
        if (expiresAt.value) {
            const selectedDate = new Date(expiresAt.value);
            if (selectedDate <= new Date()) {
                alert("La date d'expiration doit être dans le futur.");
                loading.value = false;
                return;
            }
            expirationDate = selectedDate.toISOString();
        }

        const result = await fileStore.createShareLink(itemId, props.item.type, expirationDate);
        
        localShareToken.value = result.token;
        localExpiresAt.value = expirationDate;
        
        emit('share-created'); 
        
    } catch (error) {
        console.error("Create share error:", error);
        alert("Erreur lors de la création du partage.");
    } finally {
        loading.value = false;
    }
};

const copyLink = () => {
  if (shareLinkInput.value) {
    shareLinkInput.value.select();
    navigator.clipboard.writeText(shareUrl.value).then(() => {
      linkCopied.value = true;
      setTimeout(() => linkCopied.value = false, 2000);
    }).catch(err => {
      console.error('Impossible de copier le lien:', err);
    });
  }
};

const deleteShare = async () => {
  const idToDelete = localShareId.value || props.item.share_id || props.item.ShareID;
  
  if (!idToDelete) {
      alert("Impossible de supprimer le partage (ID manquant). Veuillez rafraîchir la page.");
      return;
  }
  
  if (confirm('Êtes-vous sûr de vouloir arrêter le partage ? Le lien ne fonctionnera plus.')) {
    loading.value = true;
    try {
      await api.delete(`/shares/link/${idToDelete}`);
      localShareToken.value = null;
      localShareId.value = null;
      emit('share-deleted');
    } catch (error) {
      console.error('Erreur lors de la suppression du partage:', error);
      alert('Impossible de supprimer le partage.');
    } finally {
        loading.value = false;
    }
  }
};

// --- Friends Sharing Methods ---

const isFriendShared = (friendId) => {
    return sharedStatus.value[friendId];
}

const shareWithFriend = async (friend) => {
    if (!props.item || !friend.public_key) return;
    
    sharing.value[friend.id] = true;
    try {
        await sodium.ready;
        
        let encryptedKeyForFriend = "";
        // props.item can be file or folder. Check type.
        const resourceType = (props.item.type === 'folder' || props.item.is_dir) ? 'folder' : 'file';

        if (resourceType === 'file') {
             // Handle case sensitivity from Go backend (PascalCase) vs potentially camelCase
             const itemEncryptedKey = props.item.EncryptedKey || props.item.encrypted_key;

             if (!itemEncryptedKey) {
                  throw new Error("Clé du fichier manquante. Impossible de partager.");
             }
             
             if (!authStore.privateKey) {
                  throw new Error("Clé privée non disponible. Veuillez vous reconnecter.");
             }

             const fileKeyEncryptedBytes = sodium.from_base64(itemEncryptedKey);
             const iv = fileKeyEncryptedBytes.slice(0, 12);
             const data = fileKeyEncryptedBytes.slice(12);
             
             const fileKeyRawBuffer = await window.crypto.subtle.decrypt(
                 { name: "AES-GCM", iv: iv },
                 authStore.masterKey,
                 data
             );

             const friendPublicKey = await importKeyFromPEM(friend.public_key, 'spki');
             encryptedKeyForFriend = await encryptKeyWithPublicKey(fileKeyRawBuffer, friendPublicKey);
        }
        
        await api.post('/shares/direct', {
            resource_id: props.item.ID || props.item.id,
            resource_type: resourceType,
            friend_id: friend.id,
            encrypted_key: encryptedKeyForFriend,
            permission: 'read'
        });

        sharedStatus.value[friend.id] = true;

    } catch (e) {
        console.error("Partage échoué:", e);
        alert("Erreur: " + e.message);
    } finally {
        sharing.value[friend.id] = false;
    }
}

const close = () => {
  emit('close');
  // Clean up
  sharedStatus.value = {};
};
</script>

<style scoped>
.form-group {
    margin-bottom: 1rem;
    text-align: left;
    width: 100%;
    max-width: 300px;
    margin-left: auto;
    margin-right: auto;
}

.form-group label {
    display: block;
    margin-bottom: 0.5rem;
    font-size: 0.9rem;
    color: var(--secondary-text-color);
}

.form-control {
    width: 100%;
    padding: 8px 12px;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    font-size: 0.9rem;
    background-color: var(--card-color);
    color: var(--main-text-color);
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.6);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
  animation: fadeIn 0.2s ease;
}

.modal-content {
  background: var(--card-color);
  padding: 0;
  border-radius: 12px;
  width: 480px;
  max-width: 90%;
  box-shadow: 0 10px 25px rgba(0,0,0,0.2);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  border-bottom: 1px solid var(--border-color);
}

.modal-header h3 {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--main-text-color);
}

.btn-close {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: var(--secondary-text-color);
  padding: 0;
  line-height: 1;
}

.modal-body {
  padding: 24px;
  min-height: 150px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: var(--secondary-text-color);
}

.not-shared-state {
  text-align: center;
}

.illustration {
  font-size: 3rem;
  margin-bottom: 1rem;
}

.sub-text {
  color: var(--secondary-text-color);
  margin-bottom: 1.5rem;
  font-size: 0.9rem;
}

.shared-state {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.link-section label {
  display: block;
  font-size: 0.85rem;
  font-weight: 500;
  color: var(--secondary-text-color);
  margin-bottom: 0.5rem;
}

.link-container {
  display: flex;
  gap: 10px;
}

.link-container input {
  flex-grow: 1;
  padding: 10px 12px;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  background-color: var(--background-color);
  color: var(--main-text-color);
  font-size: 0.9rem;
  outline: none;
}

.link-container input:focus {
  border-color: var(--primary-color);
  background-color: var(--card-color);
}

.share-info {
  background-color: var(--background-color);
  color: var(--primary-color);
  padding: 12px;
  border-radius: 4px;
  font-size: 0.85rem;
  display: flex;
  align-items: center;
  border: 1px solid var(--primary-color);
}

.share-info p {
  margin: 0;
}

.modal-footer {
  padding: 16px 24px;
  border-top: 1px solid var(--border-color);
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  background-color: var(--background-color);
}

button {
  padding: 8px 16px;
  border-radius: 4px;
  border: 1px solid transparent;
  cursor: pointer;
  font-weight: 500;
  font-size: 0.9rem;
  transition: background-color 0.2s;
}

.btn-primary {
  background-color: var(--primary-color);
  color: white;
}

.btn-primary:hover {
  background-color: var(--accent-color);
  box-shadow: 0 1px 2px rgba(60,64,67,0.3);
}

.btn-secondary {
  background-color: var(--card-color);
  border: 1px solid var(--border-color);
  color: var(--main-text-color);
}

.btn-secondary:hover {
  background-color: var(--hover-background-color);
  border-color: var(--border-color);
}

.btn-copy {
  background-color: var(--card-color);
  border: 1px solid var(--border-color);
  color: var(--primary-color);
  min-width: 80px;
}

.btn-copy:hover {
  background-color: var(--hover-background-color);
}

.btn-copy.copied {
  background-color: var(--success-color);
  color: white;
  border-color: transparent;
}

.btn-delete {
  background-color: transparent;
  color: var(--error-color);
  margin-right: auto; /* Push to left */
}

.btn-delete:hover {
  background-color: var(--hover-background-color);
}

.spinner {
  border: 3px solid var(--border-color);
  border-radius: 50%;
  border-top: 3px solid var(--primary-color);
  width: 20px;
  height: 20px;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

/* Tabs */
.modal-tabs {
  display: flex;
  border-bottom: 1px solid var(--border-color);
  background-color: var(--background-color);
}

.tab-btn {
  flex: 1;
  padding: 12px;
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  color: var(--secondary-text-color);
  cursor: pointer;
  transition: all 0.2s;
}

.tab-btn:hover {
  background-color: var(--hover-background-color);
  color: var(--main-text-color);
}

.tab-btn.active {
  color: var(--primary-color);
  border-bottom-color: var(--primary-color);
  font-weight: 600;
  background-color: var(--card-color);
}

.friends-list {
  max-height: 300px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.friend-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px;
  border-radius: 6px;
  border: 1px solid var(--border-color);
}

.friend-item:hover {
  background-color: var(--hover-background-color);
}

.friend-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.friend-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background-color: var(--primary-color);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
}

.friend-name {
  font-weight: 500;
  margin: 0;
  color: var(--main-text-color);
}

.friend-email {
  font-size: 0.8rem;
  color: var(--secondary-text-color);
  margin: 0;
}

.btn-sm {
  padding: 4px 10px;
  font-size: 0.8rem;
}

.btn-success {
  background-color: var(--success-color);
  color: white;
  cursor: default;
}

.btn-outline {
  background-color: transparent;
  border: 1px solid var(--primary-color);
  color: var(--primary-color);
}

.btn-outline:hover {
  background-color: var(--primary-color);
  color: white;
}

.warning-box {
  background-color: #fff3cd;
  color: #856404;
  padding: 10px;
  border-radius: 4px;
  margin-bottom: 15px;
  font-size: 0.9rem;
  border: 1px solid #ffeeba;
}

.empty-friends {
  text-align: center;
  color: var(--secondary-text-color);
  padding: 20px;
}
</style>
