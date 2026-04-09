<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="close">
    <div class="modal-content">
      <div class="modal-header">
        <h3>{{ t('share.title', { name: item?.Name || item?.name }) }}</h3>
        <button @click="close" class="btn-close">×</button>
      </div>

      <div class="modal-body">

        <!-- === FRIENDS SECTION === -->
        <div class="friends-section">
            <h4 class="section-title">{{ t('share.withFriends') }}</h4>
            
            <div v-if="friends.length === 0" class="empty-friends">
                {{ t('friends.noFriends') }}
                <br>
                <router-link to="/friends">{{ t('friends.addFriend') }}</router-link>
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

                    <div v-if="!friend.public_key" class="key-missing" :title="t('share.keyMissing')">
                       ⚠️ {{ t('share.noKey') }}
                    </div>

                    <button v-else 
                        @click="shareWithFriend(friend)"
                        :disabled="sharing[friend.id]"
                        class="btn-sm"
                        :class="[
                          isFriendShared(friend.id) 
                            ? 'btn-danger' 
                            : 'btn-outline'
                        ]">
                        <span v-if="sharing[friend.id]">...</span>
                        <span v-else-if="isFriendShared(friend.id)">{{ t('share.stop') }}</span>
                        <span v-else>{{ t('share.send') }}</span>
                    </button>
                 </div>
            </div>
        </div>

        <div class="section-divider"></div>

        <!-- === LINK SECTION === -->
        <div class="link-section-wrapper">
            <h4 class="section-title">Partage via lien public</h4>

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
                    <label>
                      Lien de partage
                      <input type="text"/>
                    </label>
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
import { useI18n } from 'vue-i18n';
import { useFileStore } from '../stores/files';
import { useFriendStore } from '../stores/friends';
import { useAuthStore } from '../stores/auth';
import { useUIStore } from '../stores/ui';
import api from '../api';
import { decryptKeyWithPrivateKey, importKeyFromPEM, encryptKeyWithPublicKey, generateMasterKey } from '../utils/crypto';
import sodium from 'libsodium-wrappers-sumo';

const { t } = useI18n();

const props = defineProps({
  isOpen: Boolean,
  item: Object,
  initialTab: {
    type: String,
    default: 'link'
  }
});

const emit = defineEmits(['close', 'share-deleted', 'share-created']);
const fileStore = useFileStore();
const friendStore = useFriendStore();
const authStore = useAuthStore();
const uiStore = useUIStore();

// UI State
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

const fetchDirectShares = async () => {
    if (!props.item) return;
    try {
        const resourceType = (props.item.type === 'folder' || props.item.is_dir) ? 'folder' : 'file';
        const resourceId = props.item.ID || props.item.id;
        
        const response = await api.get('/shares/direct', {
            params: {
                resource_id: resourceId,
                resource_type: resourceType
            }
        });
        
        const sharedWithArg = response.data.shared_with || [];
        const statusMap = {};
        sharedWithArg.forEach(uid => {
            statusMap[uid] = true;
        });
        sharedStatus.value = statusMap;
        
    } catch (e) {
        console.error("Error fetching direct shares:", e);
    }
}

// Reset local state when item changes
watch(() => props.item, (newItem) => {
    if (newItem) {
        localShareToken.value = newItem.share_token || newItem.ShareToken;
        localShareId.value = newItem.share_id || newItem.ShareID;
        localExpiresAt.value = newItem.expires_at || newItem.ExpiresAt; // Populate expiration
        
        sharedStatus.value = {};
        fetchDirectShares();
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
        localShareId.value = result.id; // Capture ID for subsequent deletion
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
  
  uiStore.requestDeleteConfirmation({
      title: "Arrêter le partage",
      message: "Êtes-vous sûr de vouloir arrêter le partage ? Le lien ne fonctionnera plus.",
      onConfirm: async () => {
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
  });
};

// --- Friends Sharing Methods ---

const isFriendShared = (friendId) => {
    return sharedStatus.value[friendId];
}

// Decrypt a key encrypted with master key (AES-GCM, IV prepended)
async function decryptWithMasterKey(encryptedB64, masterKey) {
    const encBytes = sodium.from_base64(encryptedB64);
    const iv = encBytes.slice(0, 12);
    const data = encBytes.slice(12);
    return window.crypto.subtle.decrypt({ name: "AES-GCM", iv }, masterKey, data);
}

// Encrypt a raw key buffer with a CryptoKey, returning IV+ciphertext as base64
async function encryptWithKey(rawKeyBuffer, cryptoKey) {
    const iv = window.crypto.getRandomValues(new Uint8Array(12));
    const encrypted = await window.crypto.subtle.encrypt({ name: "AES-GCM", iv }, cryptoKey, rawKeyBuffer);
    const combined = new Uint8Array(iv.byteLength + encrypted.byteLength);
    combined.set(iv);
    combined.set(new Uint8Array(encrypted), iv.byteLength);
    return sodium.to_base64(combined);
}

// Encrypt child file and folder keys with the parent folder key
async function encryptChildKeys(files, subFolders, folderKeyCrypto, masterKey) {
    const folderFileKeys = {};
    for (const file of files) {
        const fEncKey = file.EncryptedKey || file.encrypted_key;
        if (!fEncKey) { console.warn(`File ${file.Name} has no key, skipping.`); continue; }
        try {
            const fileKeyRaw = await decryptWithMasterKey(fEncKey, masterKey);
            folderFileKeys[file.ID || file.id] = await encryptWithKey(fileKeyRaw, folderKeyCrypto);
        } catch(err) { console.warn(`Failed to process key for file ${file.Name}:`, err); }
    }
    const folderFolderKeys = {};
    for (const folder of subFolders) {
        const fEncKey = folder.EncryptedKey || folder.encrypted_key;
        if (!fEncKey) { console.warn(`Folder ${folder.Name} has no key, skipping.`); continue; }
        try {
            const subFolderKeyRaw = await decryptWithMasterKey(fEncKey, masterKey);
            folderFolderKeys[folder.ID || folder.id] = await encryptWithKey(subFolderKeyRaw, folderKeyCrypto);
        } catch(err) { console.warn(`Failed to process key for folder ${folder.Name}:`, err); }
    }
    return { folderFileKeys, folderFolderKeys };
}

// Get or create a folder key, persisting a new one to the backend if needed
async function getOrCreateFolderKey(itemId, existingEncKey, masterKey, item) {
    if (existingEncKey) {
        const folderKeyRaw = await decryptWithMasterKey(existingEncKey, masterKey);
        const folderKeyCrypto = await window.crypto.subtle.importKey(
            "raw", folderKeyRaw, { name: "AES-GCM" }, false, ["encrypt", "decrypt"]
        );
        return { folderKeyRaw, folderKeyCrypto };
    }
    const folderKeyCrypto = await generateMasterKey();
    const folderKeyRaw = await window.crypto.subtle.exportKey("raw", folderKeyCrypto);
    const encryptedKeyBase64 = await encryptWithKey(folderKeyRaw, masterKey);
    await api.put(`/folders/${itemId}/key`, { encrypted_key: encryptedKeyBase64 });
    if (item) {
        item.encrypted_key = encryptedKeyBase64;
        item.EncryptedKey = encryptedKeyBase64;
    }
    return { folderKeyRaw, folderKeyCrypto };
}

// Resolve the full folder path from item properties
function resolveFolderPath(item) {
    const itemName = item.Name || item.name;
    let itemParentPath = item.Path || item.path || '/';
    if (!itemParentPath) itemParentPath = '/';
    if (itemParentPath.endsWith('/' + itemName)) {
        console.warn("Detected Path might be full path. Using as is:", itemParentPath);
        return itemParentPath;
    }
    return (itemParentPath === '/' ? '' : itemParentPath) + '/' + itemName;
}

const shareWithFriend = async (friend) => {
    if (!props.item || !friend.public_key) return;

    if (isFriendShared(friend.id)) {
        uiStore.requestDeleteConfirmation({
           title: "Arrêter le partage",
           message: `Arrêter le partage avec ${friend.name} ?`,
           onConfirm: async () => {
             sharing.value[friend.id] = true;
             try {
                const resourceType = (props.item.type === 'folder' || props.item.is_dir) ? 'folder' : 'file';
                const resourceId = props.item.ID || props.item.id;
                await api.delete(`/shares/direct`, {
                    params: { resource_id: resourceId, resource_type: resourceType, friend_id: friend.id }
                });
                sharedStatus.value[friend.id] = false;
             } catch(e) {
                console.error("Revoke failed:", e);
                if (e.response && e.response.status === 404) {
                     sharedStatus.value[friend.id] = false;
                } else {
                     alert("Erreur lors de la suppression du partage.");
                }
             } finally {
                sharing.value[friend.id] = false;
             }
           }
        });
        return;
    }

    sharing.value[friend.id] = true;
    try {
        await sodium.ready;
        const resourceType = (props.item.type === 'folder' || props.item.is_dir) ? 'folder' : 'file';

        if (resourceType === 'file') {
             const itemEncryptedKey = props.item.EncryptedKey || props.item.encrypted_key;
             if (!itemEncryptedKey) throw new Error("Clé du fichier manquante. Impossible de partager.");
             if (!authStore.masterKey) throw new Error("Clé Maître non disponible (Session expirée ?). Veuillez vous reconnecter.");
             if (!authStore.privateKey) throw new Error("Clé privée non disponible. Veuillez vous reconnecter.");

             const fileKeyRawBuffer = await decryptWithMasterKey(itemEncryptedKey, authStore.masterKey);
             const friendPublicKey = await importKeyFromPEM(friend.public_key, 'spki');
             const encryptedKeyForFriend = await encryptKeyWithPublicKey(fileKeyRawBuffer, friendPublicKey);
             await api.post('/shares/direct', {
                resource_id: props.item.ID || props.item.id,
                resource_type: resourceType,
                friend_id: friend.id,
                encrypted_key: encryptedKeyForFriend,
                permission: 'read',
             });
             sharedStatus.value[friend.id] = true;

        } else if (resourceType === 'folder') {
            if (!authStore.masterKey) throw new Error("Clé Maître non disponible (Session expirée ?). Veuillez vous reconnecter.");

            const itemId = props.item.ID || props.item.id;
            const existingEncKey = props.item.EncryptedKey || props.item.encrypted_key;
            const { folderKeyRaw, folderKeyCrypto } = await getOrCreateFolderKey(
                itemId, existingEncKey, authStore.masterKey, props.item
            );

            const friendPublicKey = await importKeyFromPEM(friend.public_key, 'spki');
            const encryptedKeyForFriend = await encryptKeyWithPublicKey(folderKeyRaw, friendPublicKey);

            const folderPath = resolveFolderPath(props.item);
            const listRes = await api.get(`/files/list-recursive?path=${encodeURIComponent(folderPath)}`);
            const files = listRes.data.files || [];
            const subFolders = listRes.data.folders || [];

            const { folderFileKeys, folderFolderKeys } = await encryptChildKeys(
                files, subFolders, folderKeyCrypto, authStore.masterKey
            );

            await api.post('/shares/direct', {
                resource_id: props.item.ID || props.item.id,
                resource_type: resourceType,
                friend_id: friend.id,
                encrypted_key: encryptedKeyForFriend,
                permission: 'read',
                folder_file_keys: folderFileKeys,
                folder_folder_keys: folderFolderKeys
            });
            sharedStatus.value[friend.id] = true;
        }

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

.btn-danger {
  background-color: var(--error-color);
  color: white;
  border: 1px solid var(--error-color);
}

.btn-danger:hover {
  filter: brightness(0.9);
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

.section-title {
  margin: 0 0 15px 0;
  font-size: 1rem;
  font-weight: 600;
  color: var(--main-text-color);
}

.section-divider {
  height: 1px;
  background-color: var(--border-color);
  margin: 20px 0;
}

.link-section-wrapper {
  margin-top: 10px;
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
  background-color: var(--card-color);
  color: var(--warning-color);
  padding: 10px;
  border-radius: 4px;
  margin-bottom: 15px;
  font-size: 0.9rem;
  border: 1px solid var(--warning-color);
}

.empty-friends {
  text-align: center;
  color: var(--secondary-text-color);
  padding: 20px;
}
</style>
