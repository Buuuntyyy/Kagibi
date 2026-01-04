<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="close">
    <div class="modal-content">
      <div class="modal-header">
        <h3>Partager "{{ item?.Name || item?.name }}"</h3>
        <button @click="close" class="btn-close">×</button>
      </div>
      
      <div class="modal-body">
        <!-- Loading State -->
        <div v-if="loading" class="loading-state">
            <div class="spinner"></div> Traitement en cours...
        </div>

        <!-- Not Shared State -->
        <div v-else-if="!isShared" class="not-shared-state">
            <div class="illustration">
                🔗
            </div>
            <p>Ce {{ item?.type === 'folder' ? 'dossier' : 'fichier' }} n'est pas encore partagé.</p>
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
                <p>⚠️ Toute personne disposant de ce lien pourra accéder au contenu <b>déchiffré</b> de manière légitime.</p>
            </div>
        </div>
      </div>

      <div class="modal-footer">
        <button v-if="isShared" @click="deleteShare" class="btn-delete">Arrêter le partage</button>
        <button @click="close" class="btn-secondary">Terminé</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue';
import { useFileStore } from '../stores/files';
import api from '../api';

const props = defineProps({
  isOpen: Boolean,
  item: Object,
});

const emit = defineEmits(['close', 'share-deleted', 'share-created']);
const fileStore = useFileStore();

const loading = ref(false);
const linkCopied = ref(false);
const shareLinkInput = ref(null);
const expiresAt = ref(null);

// Local state to handle immediate updates without waiting for parent refresh
const localShareToken = ref(null);
const localShareId = ref(null);

// Reset local state when item changes
watch(() => props.item, (newItem) => {
    if (newItem) {
        localShareToken.value = newItem.share_token || newItem.ShareToken;
        localShareId.value = newItem.share_id || newItem.ShareID;
    }
}, { immediate: true });

const isShared = computed(() => !!localShareToken.value);

const shareUrl = computed(() => {
  if (localShareToken.value) {
    return `${window.location.origin}/s/${localShareToken.value}`;
  }
  return '';
});

const selectAll = (e) => {
    e.target.select();
}

const createShare = async () => {
    if (!props.item) return;
    loading.value = true;
    try {
        // Use the store action which handles encryption
        // Note: Go struct fields are capitalized (ID), so we must use item.ID if item.id is undefined
        const itemId = props.item.ID || props.item.id;
        
        // Convert expiresAt to ISO string if present
        let expirationDate = null;
        if (expiresAt.value) {
            expirationDate = new Date(expiresAt.value).toISOString();
        }

        const result = await fileStore.createShareLink(itemId, props.item.type, expirationDate);
        
        // Update local state
        localShareToken.value = result.token;
        // If backend returns ID, use it. If not, we might need to refresh.
        // Assuming result has token.
        
        emit('share-created'); // Parent should refresh list
        
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
      // If we don't have ID (e.g. just created and backend didn't return ID), 
      // we can't delete immediately.
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

const close = () => {
  emit('close');
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
    color: #5f6368;
}

.form-control {
    width: 100%;
    padding: 8px 12px;
    border: 1px solid #dadce0;
    border-radius: 4px;
    font-size: 0.9rem;
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
  background: white;
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
  border-bottom: 1px solid #eee;
}

.modal-header h3 {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 600;
  color: #202124;
}

.btn-close {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: #5f6368;
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
  color: #5f6368;
}

.not-shared-state {
  text-align: center;
}

.illustration {
  font-size: 3rem;
  margin-bottom: 1rem;
}

.sub-text {
  color: #5f6368;
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
  color: #5f6368;
  margin-bottom: 0.5rem;
}

.link-container {
  display: flex;
  gap: 10px;
}

.link-container input {
  flex-grow: 1;
  padding: 10px 12px;
  border: 1px solid #dadce0;
  border-radius: 4px;
  background-color: #f8f9fa;
  color: #3c4043;
  font-size: 0.9rem;
  outline: none;
}

.link-container input:focus {
  border-color: #1a73e8;
  background-color: white;
}

.share-info {
  background-color: #e8f0fe;
  color: #1967d2;
  padding: 12px;
  border-radius: 4px;
  font-size: 0.85rem;
  display: flex;
  align-items: center;
}

.share-info p {
  margin: 0;
}

.modal-footer {
  padding: 16px 24px;
  border-top: 1px solid #eee;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  background-color: #f8f9fa;
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
  background-color: #1a73e8;
  color: white;
}

.btn-primary:hover {
  background-color: #1765cc;
  box-shadow: 0 1px 2px rgba(60,64,67,0.3);
}

.btn-secondary {
  background-color: white;
  border: 1px solid #dadce0;
  color: #3c4043;
}

.btn-secondary:hover {
  background-color: #f1f3f4;
  border-color: #dadce0;
}

.btn-copy {
  background-color: white;
  border: 1px solid #dadce0;
  color: #1a73e8;
  min-width: 80px;
}

.btn-copy:hover {
  background-color: #f1f3f4;
}

.btn-copy.copied {
  background-color: #e6f4ea;
  color: #137333;
  border-color: transparent;
}

.btn-delete {
  background-color: transparent;
  color: #d93025;
  margin-right: auto; /* Push to left */
}

.btn-delete:hover {
  background-color: #fce8e6;
}

.spinner {
  border: 3px solid rgba(0, 0, 0, 0.1);
  border-radius: 50%;
  border-top: 3px solid #1a73e8;
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
</style>
