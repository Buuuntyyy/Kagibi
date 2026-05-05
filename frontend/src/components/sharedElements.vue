<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="file-list-container">
    <div class="path-banner">
      <div class="breadcrumbs">
        <span class="breadcrumb-link current">{{ t('nav.shared') }}</span>
      </div>
    </div>

    <div class="scrollable-content">
      <!-- Section: Mes Partages -->
      <div class="accordion-item">
        <div class="accordion-header" @click="toggleSection('my-shares')" :class="{ active: sections['my-shares'] }">
          <span class="accordion-title">{{ t('shared.myShares') }} ({{ uniqueShares.length }})</span>
          <span class="accordion-icon">{{ sections['my-shares'] ? '▼' : '▶' }}</span>
        </div>
        
        <div v-show="sections['my-shares']" class="accordion-content">
          <div v-if="loading" class="loading">
            <div class="spinner"></div> {{ t('shared.loading') }}
          </div>
          <div v-else-if="error" class="error">{{ error }}</div>
          <div v-else-if="uniqueShares.length === 0" class="empty">
            <p>{{ t('shared.noShares') }}</p>
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
              <span v-if="item.resource_type === 'file'" class="badge file">{{ t('file.files') }}</span>
              <span v-else class="badge folder">{{ t('file.folders') }}</span>
            </template>

            <template #views="{ item }">
              {{ item.views }}
            </template>

            <template #created_at="{ item }">
              {{ formatDate(item.created_at) }}
            </template>

            <template #expires_at="{ item }">
              <span :class="{ 'expired': isExpired(item.expires_at) }">
                {{ item.expires_at ? formatDate(item.expires_at) : t('shared.expired') }}
              </span>
            </template>

            <template #link="{ item }">
              <div class="link-actions">
                <button v-if="item._hasPublicLink" @click.stop="copyLink(item.link)" class="icon-btn" :title="t('shared.copyLink')">
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M16 1H4c-1.1 0-2 .9-2 2v14h2V3h12V1zm3 4H8c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h11c1.1 0 2-.9 2-2V7c0-1.1-.9-2-2-2zm0 16H8V7h11v14z"/></svg>
                </button>
              </div>
            </template>

            <template #actions="{ item }">
              <div class="action-group">
                <button @click.stop="openManageDialog(item)" class="icon-btn" title="Gérer le partage">
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M19.14 12.94c.04-.3.06-.61.06-.94 0-.32-.02-.64-.07-.94l2.03-1.58c.18-.14.23-.41.12-.61l-1.92-3.32c-.12-.22-.37-.29-.59-.22l-2.39.96c-.5-.38-1.03-.7-1.62-.94l-.36-2.54c-.04-.24-.24-.41-.48-.41h-3.84c-.24 0-.43.17-.47.41l-.36 2.54c-.59.24-1.13.57-1.62.94l-2.39-.96c-.22-.08-.47 0-.59.22L2.74 8.87c-.12.21-.08.47.12.61l2.03 1.58c-.05.3-.09.63-.09.94s.02.64.07.94l-2.03 1.58c-.18.14-.23.41-.12.61l1.92 3.32c.12.22.37.29.59.22l2.39-.96c.5.38 1.03.7 1.62.94l.36 2.54c.05.24.24.41.48.41h3.84c.24 0 .44-.17.47-.41l.36-2.54c.59-.24 1.13-.56 1.62-.94l2.39.96c.22.08.47 0 .59-.22l1.92-3.32c.12-.22.07-.47-.12-.61l-2.01-1.58zM12 15.6c-1.98 0-3.6-1.62-3.6-3.6s1.62-3.6 3.6-3.6 3.6 1.62 3.6 3.6-1.62 3.6-3.6 3.6z"/></svg>
                </button>
                <button v-if="item.resource_type === 'folder'" @click.stop="navigateToFolder(item)" class="icon-btn" title="Naviguer vers le dossier">
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M10 4H4c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z"/></svg>
                </button>
                <button @click.stop="deleteShare(item.id, item)" class="delete-btn" :title="t('shared.deleteShare')">
                  <svg viewBox="0 0 24 24" width="13" height="13" fill="currentColor"><path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM15.5 4l-1-1h-5l-1 1H5v2h14V4z"/></svg>
                  {{ t('common.delete') }}
                </button>
              </div>
            </template>
          </FileTable>
        </div>
      </div>

      <!-- Section: Partagés avec moi -->
      <div class="accordion-item">
        <div class="accordion-header" @click="toggleSection('shared-with-me')" :class="{ active: sections['shared-with-me'] }">
          <span class="accordion-title">{{ t('shared.sharedWithMe') }}</span>
          <span class="accordion-icon">{{ sections['shared-with-me'] ? '▼' : '▶' }}</span>
        </div>
        <div v-show="sections['shared-with-me']" class="accordion-content">
          <SharedWithMe />
        </div>
      </div>
    </div>
  </div>

  <ManageShareDialog
    :isOpen="showManageDialog"
    :item="managingItem"
    @close="showManageDialog = false"
  />
</template>

<script setup>
import { ref, onMounted, computed, reactive, watch }from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import api from '../api';
import FileTable from './file/FileTable.vue';
import SharedWithMe from './sharedWithMe.vue';
import ManageShareDialog from './ManageShareDialog.vue';
import { useFileStore } from '../stores/files';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const fileStore = useFileStore();
const shares = ref([]);
const loading = ref(true);
const error = ref(null);
const showManageDialog = ref(false);
const managingItem = ref(null);

const sections = reactive({
  'my-shares': true,
  'shared-with-me': false
});

// Auto-open section if query param is present
const checkQueryAndOpenSection = () => {
  if (route.query.folderId || route.query.section === 'shared-with-me') {
    sections['my-shares'] = false;
    sections['shared-with-me'] = true;
  }
};


// Watch for file store updates (triggered via WebSocket)
watch(() => fileStore.shareUpdateTrigger, () => {
    fetchShares();
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

// Deduplicate: one entry per (resource_type, resource_id), preferring the public link entry
const uniqueShares = computed(() => {
  const map = new Map();
  for (const share of shares.value) {
    const key = `${share.resource_type}:${share.resource_id}`;
    if (!map.has(key)) {
      map.set(key, { ...share, _hasPublicLink: share.token !== 'DIRECT' });
    } else {
      const existing = map.get(key);
      if (share.token !== 'DIRECT' && !existing._hasPublicLink) {
        map.set(key, { ...share, _hasPublicLink: true });
      }
    }
  }
  return Array.from(map.values());
});

const sharedFolders = computed(() => uniqueShares.value.filter(s => s.resource_type === 'folder').map(s => ({...s, ID: s.id})))
const sharedFiles = computed(() => uniqueShares.value.filter(s => s.resource_type === 'file').map(s => ({...s, ID: s.id})))

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

const deleteShare = async (id, item) => {
  if (!confirm("Voulez-vous vraiment supprimer ce partage ?")) return;

  try {
    if (item._hasPublicLink) {
      await api.delete(`/shares/link/${id}`);
    } else {
      await api.delete(`/shares/direct`, {
        params: { id, resource_type: item.resource_type }
      });
    }
    // Remove all raw entries for this resource so uniqueShares updates correctly
    shares.value = shares.value.filter(
      s => !(s.resource_type === item.resource_type && s.resource_id === item.resource_id)
    );
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

const openManageDialog = (share) => {
  managingItem.value = {
    ID: share.resource_id,
    Name: share.resource_name,
    type: share.resource_type,
    share_token: share._hasPublicLink ? share.token : null,
    share_id: share._hasPublicLink ? share.id : null,
    perm_download: share.perm_download,
    perm_create: share.perm_create,
    perm_delete: share.perm_delete,
    perm_move: share.perm_move,
  };
  showManageDialog.value = true;
};

const navigateToFolder = (share) => {
  fileStore.currentPath = share.resource_path || '/';
  router.push('/dashboard/files');
};

onMounted(() => {
  checkQueryAndOpenSection();
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
.accordion-header {
  border: 1px solid #eee;
  border-radius: 8px;
  margin-bottom: 1rem;
  overflow: hidden;
  background-color: var(--card-color);
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  cursor: pointer;
  transition: background-color 0.2s;
  user-select: none;
}

.accordion-header:hover {
  background-color: var(--hover-background-color);
}

.accordion-header.active {
  background-color: var(--background-color);
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
  color: var(--text-color);
  cursor: default;
  font-weight: bold;
  text-decoration: none;
}

.action-group {
  display: flex;
  align-items: center;
  gap: 4px;
}
</style>
