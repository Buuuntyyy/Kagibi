<template>
  <div class="file-list-container">
    <div class="toolbar" v-if="preferenceStore.showToolBar">
      <div class="toolbar-left" style="visibility: hidden">
         <button class="btn-add-file">Spacer</button>
      </div>
    </div>
    <div class="path-banner">
      <div class="breadcrumbs">
        <span class="breadcrumb-link current">Partage</span>
      </div>
    </div>

    <div class="scrollable-content">
      <!-- Section: Mes Partages -->
      <div class="accordion-item">
        <div class="accordion-header" @click="toggleSection('my-shares')" :class="{ active: sections['my-shares'] }">
          <span class="accordion-title">Fichiers partagés ({{ shares.length }})</span>
          <span class="accordion-icon">{{ sections['my-shares'] ? '▼' : '▶' }}</span>
        </div>
        
        <div v-show="sections['my-shares']" class="accordion-content">
          <div v-if="loading" class="loading">
            <div class="spinner"></div> Chargement...
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
      </div>

      <!-- Section: Partagés avec moi -->
      <div class="accordion-item">
        <div class="accordion-header" @click="toggleSection('shared-with-me')" :class="{ active: sections['shared-with-me'] }">
          <span class="accordion-title">Partagés avec moi</span>
          <span class="accordion-icon">{{ sections['shared-with-me'] ? '▼' : '▶' }}</span>
        </div>
        <div v-show="sections['shared-with-me']" class="accordion-content">
          <SharedWithMe />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, reactive } from 'vue';
import api from '../api';
import FileTable from './file/FileTable.vue';
import SharedWithMe from './sharedWithMe.vue';
import { usePreferencesStore } from '../stores/preferences';

const preferenceStore = usePreferencesStore();
const shares = ref([]);
const loading = ref(true);
const error = ref(null);

const sections = reactive({
  'my-shares': true,
  'shared-with-me': false
});

const toggleSection = (section) => {
  sections[section] = !sections[section];
};

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

.badge {
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.85rem;
  font-weight: 500;
}

.badge.file {
  background-color: #e3f2fd;
  color: #1976d2;
}

.badge.folder {
  background-color: #fff3e0;
  color: #f57c00;
}

.link-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.open-link {
  color: #2196f3;
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
  font-size: 1.1rem;
  padding: 0.25rem;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.icon-btn:hover {
  background-color: #f0f0f0;
}

.delete-btn {
  background-color: #ffebee;
  color: #c62828;
  border: none;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.85rem;
  transition: background-color 0.2s;
}

.delete-btn:hover {
  background-color: #ffcdd2;
}

.expired {
  color: #d32f2f;
  font-weight: 500;
}

.loading, .error, .empty {
  padding: 2rem;
  text-align: center;
  color: #666;
}

.error {
  color: #d32f2f;
}

.spinner {
  display: inline-block;
  width: 20px;
  height: 20px;
  border: 3px solid rgba(0, 0, 0, 0.1);
  border-radius: 50%;
  border-top-color: #2196f3;
  animation: spin 1s ease-in-out infinite;
  margin-right: 0.5rem;
  vertical-align: middle;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Accordion Styles */
.scrollable-content {
  overflow-y: auto;
  flex-grow: 1;
  padding: 0 1rem;
}

.accordion-item {
  border: 1px solid #eee;
  border-radius: 8px;
  margin-bottom: 1rem;
  overflow: hidden;
  background-color: white;
}

.accordion-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  background-color: #f9f9f9;
  cursor: pointer;
  transition: background-color 0.2s;
  user-select: none;
}

.accordion-header:hover {
  background-color: #f0f0f0;
}

.accordion-header.active {
  background-color: #e3f2fd;
  color: #1976d2;
}

.accordion-title {
  font-weight: 600;
  font-size: 1.1rem;
}

.accordion-icon {
  font-size: 0.8rem;
  color: #666;
}

.accordion-content {
  padding: 0;
  border-top: 1px solid #eee;
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
</style>
