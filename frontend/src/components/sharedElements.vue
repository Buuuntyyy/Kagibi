<template>
  <div class="file-list-container">
    <div class="toolbar" v-if="preferenceStore.showToolBar">
      <div class="toolbar-left" style="visibility: hidden">
         <button class="btn-add-file">Spacer</button>
      </div>
    </div>
    <div class="path-banner">
      <div class="breadcrumbs">
        <span class="breadcrumb-link current">Mes Partages</span>
      </div>
    </div>

    <div v-if="loading" class="loading">
      <div class="spinner"></div> Chargement des partages...
    </div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="shares.length === 0" class="empty">
      <p>Vous n'avez aucun partage actif.</p>
    </div>
    <FileTable 
      v-else 
      :folders="sharedFolders"
      :files="sharedFiles"
      :columns="columns"
    >
      <template #resource_name="{ item }">
        <span :title="item.resource_name">{{ item.resource_name }}</span>
      </template>

      <template #resource_type="{ item }">
        <span v-if="item.resource_type === 'file'" class="badge file">Fichier</span>
        <span v-else class="badge folder">Dossier</span>
      </template>

      <template #views="{ item }">
        {{ item.views }}
      </template>

      <template #created_at="{ item }">
        {{ formatDate(item.created_at) }}
      </template>

      <template #expires_at="{ item }">
        <span :class="{ 'expired': isExpired(item.expires_at) }">
          {{ item.expires_at ? formatDate(item.expires_at) : 'Jamais' }}
        </span>
      </template>

      <template #link="{ item }">
        <div class="link-actions">
          <a :href="item.link" target="_blank" class="open-link">Ouvrir</a>
          <button @click.stop="copyLink(item.link)" class="icon-btn" title="Copier le lien">
            📋
          </button>
        </div>
      </template>

      <template #actions="{ item }">
        <button @click.stop="deleteShare(item.id)" class="delete-btn" title="Supprimer le partage">
          🗑️ Supprimer
        </button>
      </template>
    </FileTable>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import api from '../api';
import FileTable from './file/FileTable.vue';
import { usePreferencesStore } from '../stores/preferences';

const preferenceStore = usePreferencesStore();
const shares = ref([]);
const loading = ref(true);
const error = ref(null);

const columns = [
  { key: 'resource_name', label: 'Nom', cellClass: 'name-cell' },
  { key: 'resource_type', label: 'Type' },
  { key: 'views', label: 'Vues' },
  { key: 'created_at', label: 'Créé le' },
  { key: 'expires_at', label: 'Expire le' },
  { key: 'link', label: 'Lien', cellClass: 'link-cell' },
  { key: 'actions', label: 'Actions' }
]

const sharedFolders = computed(() => shares.value.filter(s => s.resource_type === 'folder').map(s => ({...s, ID: s.id})))
const sharedFiles = computed(() => shares.value.filter(s => s.resource_type === 'file').map(s => ({...s, ID: s.id})))

const fetchShares = async () => {
  loading.value = true;
  error.value = null;
  try {
    const response = await api.get('/shares/list');
    shares.value = response.data.shares || [];
  } catch (err) {
    console.error("Error fetching shares:", err);
    error.value = "Impossible de charger la liste des partages.";
  } finally {
    loading.value = false;
  }
};

const deleteShare = async (id) => {
  if (!confirm("Voulez-vous vraiment supprimer ce lien de partage ? Le lien ne fonctionnera plus.")) return;
  
  try {
    await api.delete(`/shares/link/${id}`);
    shares.value = shares.value.filter(s => s.id !== id);
  } catch (err) {
    console.error("Error deleting share:", err);
    alert("Erreur lors de la suppression du partage.");
  }
};

const copyLink = (link) => {
  const fullLink = `${window.location.origin}${link}`;
  navigator.clipboard.writeText(fullLink).then(() => {
      // Could use a toast notification here
      alert("Lien copié dans le presse-papier !");
  }).catch(err => {
      console.error('Failed to copy: ', err);
  });
};

const formatDate = (dateString) => {
  if (!dateString) return '';
  return new Date(dateString).toLocaleDateString('fr-FR', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  });
};

const isExpired = (dateString) => {
  if (!dateString) return false;
  return new Date(dateString) < new Date();
};

onMounted(() => {
  fetchShares();
});
</script>

<style scoped>
.file-list-container {
  margin-top: 0.8rem;
  position: relative;
  background-color: var(--card-color);
  height: 100%;
  width: 100%;
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
}

.path-banner {
  padding: 0.5rem 1rem;
  padding-top: 0;
  background-color: var(--card-color);
  display: flex;
  justify-content: flex-start;
  align-items: center;
  gap: 1rem;
}

.back-arrow {
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  padding: 0;
  color: #555;
  border: 1px solid #ccc;
  border-radius: 50%;
  background-color: white;
  transition: all 0.2s;
}

.back-arrow:not(.disabled):hover {
  background-color: #f0f0f0;
  border-color: #bbb;
  color: #333;
}

.back-arrow.disabled {
  color: #ccc;
  cursor: not-allowed;
  background-color: #f9f9f9;
  border-color: #eee;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.dashboard-title {
  margin: 0;
  font-size: 1.5rem;
  text-align: left;
}

.dashboard-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
  width: 80vw; /* Match file-list-container width */
}

.header-actions {
  display: flex;
  gap: 1rem;
}

.btn-header {
  background-color: white;
  border: 1px solid #ccc;
  color: #333;
  padding: 0.5rem 1rem;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 500;
  transition: background-color 0.2s;
}

.btn-header:hover {
  background-color: #f0f0f0;
}

.btn-logout {
  color: #dc3545;
  border-color: #dc3545;
}

.btn-logout:hover {
  background-color: #dc3545;
  color: white;
}

.list-area {
  margin-top: 1rem;
  overflow-y: auto;
  flex-grow: 1;
  padding: 0 1rem;
}


.size {
  color: #5c5c5c;
  font-size: 0.9em;
  text-align: right;
}

.list-item.selected {
  background-color: rgba(66, 185, 131, 0.2); /* Light green selection */
}

button {
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-weight: bold;
}

.btn-add-file {
  background-color: var(--primary-color);
  color: white;
}

.btn-rename {
  background-color: #ffc107;
  color: #333;
  margin-right: 0.5rem;
}

.btn-rename:disabled {
  background-color: #e0e0e0;
  color: #999;
  cursor: not-allowed;
}

.btn-download {
  background-color: var(--primary-color);
  color: white;
}

.path-banner button {
    background-color: var(--background-color);
    border: 1px solid #ccc;
}

.path-banner button:disabled {
    cursor: not-allowed;
    opacity: 0.5;
}

.breadcrumbs {
  display: flex;
  align-items: center;
  font-size: 1.5rem;
  transform: translateY(-2px);
}

.breadcrumb-segment {
  display: flex;
  align-items: center;
}

.breadcrumb-link {
  cursor: pointer;
  color: var(--primary-color, #42b983);
  text-decoration: none;
  padding: 0.2rem 0.5rem;
  padding-top: 0.2rem;
  border-radius: 4px;
  transition: all 0.2s ease;
}

.breadcrumb-link:hover {
  text-decoration: underline;
  background-color: rgba(66, 185, 131, 0.1);
  transform: translateY(-1px);
}

.breadcrumb-link.current {
  color: #333;
  cursor: default;
  font-weight: bold;
  text-decoration: none;
}

.separator {
  margin: 0 0.5rem;
  color: #999;
}

.progress-container {
  padding: 0.5rem 1rem;
  display: flex;
  align-items: center;
  gap: 1rem;
  background-color: #f9f9f9;
  border-bottom: 1px solid #eee;
}

.progress-bar {
  flex-grow: 1;
  height: 10px;
  background-color: #e0e0e0;
  border-radius: 5px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background-color: var(--primary-color, #42b983);
  transition: width 0.3s ease;
}

.progress-text {
  font-size: 0.9rem;
  font-weight: bold;
  color: #555;
  min-width: 3rem;
  text-align: right;
}

.upload-popup {
  position: fixed;
  top: 20px;
  right: 20px;
  width: 320px;
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
  z-index: 1000;
  border: 1px solid #eee;
  overflow: hidden;
  animation: slideIn 0.3s ease;
}

@keyframes slideIn {
  from { transform: translateX(100%); opacity: 0; }
  to { transform: translateX(0); opacity: 1; }
}

.popup-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem 1rem;
  background-color: #f5f5f5;
  border-bottom: 1px solid #eee;
}

.popup-title {
  font-weight: bold;
  font-size: 0.9rem;
  color: #333;
}

.btn-close {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  padding: 0;
  line-height: 0.8;
  color: #999;
  width: auto;
  height: auto;
}

.btn-close:hover {
  color: #333;
  background: none;
}

.popup-content {
  padding: 1rem;
}

.file-name {
  margin-bottom: 0.8rem;
  font-size: 0.9rem;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  color: #333;
  text-align: left;
}

.progress-container-popup {
  display: flex;
  align-items: center;
  gap: 0.8rem;
}

.drag-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(255, 255, 255, 0.9);
  border: 3px dashed var(--primary-color);
  border-radius: 8px;
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 500;
  pointer-events: none; /* Permet de drop "au travers" de l'overlay */
}

.drag-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  color: var(--primary-color);
}

.drag-icon {
  font-size: 4rem;
  margin-bottom: 1rem;
}

.drag-text {
  font-size: 1.5rem;
  font-weight: bold;
}



.context-menu {
  position: fixed;
  background: white;
  border: 1px solid #ccc;
  border-radius: 4px;
  box-shadow: 0 2px 10px rgba(0,0,0,0.2);
  z-index: 1000;
  min-width: 150px;
  padding: 5px 0;
}

.menu-item {
  padding: 8px 15px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.9rem;
  color: #333;
  transition: background-color 0.2s;
  text-align: left;
}

.menu-item:hover {
  background-color: #f0f0f0;
}

.menu-item.delete {
  color: #dc3545;
  border-top: 1px solid #eee;
}



.date-column {
  font-size: 0.9em;
  color: #666;
}

.name-cell {
  font-weight: 500;
  color: var(--main-text-color);
  max-width: 250px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.badge {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
}

.badge.file {
  background-color: rgba(66, 165, 245, 0.2);
  color: #90caf9;
}

.badge.folder {
  background-color: rgba(255, 167, 38, 0.2);
  color: #ffcc80;
}

.expired {
  color: #ef5350;
  font-weight: bold;
}

.link-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.open-link {
  color: #64b5f6;
  text-decoration: none;
  font-size: 0.9rem;
}

.open-link:hover {
  text-decoration: underline;
}

.icon-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 1.2rem;
  padding: 4px;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.icon-btn:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

.delete-btn {
  background-color: transparent;
  color: #ef5350;
  border: 1px solid #ef5350;
  padding: 6px 12px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.85rem;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 5px;
}

.delete-btn:hover {
  background-color: #ef5350;
  color: white;
}

.loading, .error, .empty {
  text-align: center;
  padding: 60px 20px;
  background-color: #1e1e1e;
  border-radius: 8px;
  color: #888;
}

.error {
  color: #ef5350;
}

.spinner {
  border: 3px solid rgba(255, 255, 255, 0.1);
  border-radius: 50%;
  border-top: 3px solid #64b5f6;
  width: 24px;
  height: 24px;
  animation: spin 1s linear infinite;
  display: inline-block;
  vertical-align: middle;
  margin-right: 10px;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}
</style>
