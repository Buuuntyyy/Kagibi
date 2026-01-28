<template>
  <div class="file-list-container">
    <!-- Breadcrumbs & Navigation -->
    <div class="path-banner">
      <button @click="goUp" class="btn-icon back-btn" :disabled="store.currentPath === '/'" title="Remonter">
        <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M15 18l-6-6 6-6"/>
        </svg>
      </button>
      
      <div class="breadcrumbs">
        <div v-for="(segment, index) in pathSegments" :key="index" class="breadcrumb-item">
          <span 
            class="breadcrumb-link" 
            :class="{ 'active': index === pathSegments.length - 1 }"
            @click="navigateToPath(segment.path)"
          >
            {{ segment.name }}
          </span>
          <span v-if="index < pathSegments.length - 1" class="separator">/</span>
        </div>
      </div>
    </div>

    <!-- Table -->
    <div class="table-responsive">
      <table class="files-table">
        <thead>
          <tr>
            <th class="col-icon"></th>
            <th class="col-name">Nom</th>
            <th class="col-date">Modifié le</th>
            <th class="col-size">Taille</th>
            <th class="col-actions"></th>
          </tr>
        </thead>
        <tbody>
          <!-- Empty State -->
          <tr v-if="store.folders.length === 0 && store.files.length === 0">
             <td colspan="5" class="empty-state">
                <p>Ce dossier est vide.</p>
             </td>
          </tr>

          <!-- Folders -->
          <tr v-for="folder in store.folders" :key="folder.ID" 
              class="list-item folder-item" 
              @dblclick="openFolder(folder.Name)">
            <td class="col-icon">
              <!-- Folder Icon -->
              <svg class="file-icon folder-icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M10 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z" fill="#5f6368"/>
              </svg>
            </td>
            <td class="col-name">
              <span class="name-text">{{ folder.Name }}</span>
            </td>
            <td class="col-date">{{ formatDate(folder.CreatedAt) }}</td>
            <td class="col-size">-</td>
             <td class="col-actions"></td>
          </tr>

          <!-- Files -->
          <tr v-for="file in store.files" :key="file.ID" 
              class="list-item file-item"
              @dblclick="openFile(file)">
            <td class="col-icon">
              <span class="file-icon-wrapper">
                 <!-- Simple File Icon -->
                 <svg class="file-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z" fill="#888"/>
                 </svg>
              </span>
            </td>
            <td class="col-name">
              <span class="name-text">{{ file.Name }}</span>
            </td>
            <td class="col-date">{{ formatDate(file.UpdatedAt) }}</td>
            <td class="col-size">{{ formatSize(file.Size) }}</td>
            <td class="col-actions">
              <button @click.stop="downloadFile(file)" class="action-btn download-btn" title="Télécharger">
                <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
                   <path d="M19 9h-4V3H9v6H5l7 7 7-7zm-8 2V5h2v6h1.17L12 13.17 9.83 11H11zm-6 7h14v2H5v-2z"/>
                </svg>
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { usePublicFileStore } from '../stores/publicFileStore';

const store = usePublicFileStore();

const pathSegments = computed(() => {
  const path = store.currentPath;
  const segments = [{ name: 'Racine', path: '/' }];

  if (path === '/') return segments;

  const parts = path.split('/').filter(p => p);
  let currentBuild = '';

  parts.forEach(part => {
    currentBuild += '/' + part;
    segments.push({ name: part, path: currentBuild });
  });

  return segments;
});

const navigateToPath = (path) => {
  if (path === store.currentPath) return;
  store.fetchItems(store.shareToken, path);
};

const openFolder = (folderName) => {
  store.navigateTo(folderName);
};

const openFile = (file) => {
  store.downloadFile(file.ID, file.Name);
};

const goUp = () => {
  store.navigateUp();
};

const downloadFile = (file) => {
  store.downloadFile(file.ID, file.Name);
};

const formatSize = (bytes) => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return Number.parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

const formatDate = (dateString) => {
  if (!dateString) return '-';
  const date = new Date(dateString);
  return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
};
</script>

<style scoped>
.file-list-container {
  display: flex;
  flex-direction: column;
  width: 100%;
}

/* Path Banner */
.path-banner {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px 24px;
  border-bottom: 1px solid var(--border-color);
  background: var(--card-color);
}

.back-btn {
  background: none;
  border: none;
  color: var(--secondary-text-color);
  cursor: pointer;
  padding: 4px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s, color 0.2s;
}

.back-btn:hover:not(:disabled) {
  background: var(--hover-background-color);
  color: var(--main-text-color);
}

.back-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.breadcrumbs {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
}

.breadcrumb-item {
  display: flex;
  align-items: center;
}

.breadcrumb-link {
  color: var(--secondary-text-color);
  cursor: pointer;
  font-weight: 500;
  padding: 4px 6px;
  border-radius: 4px;
  transition: color 0.2s, background 0.2s;
}

.breadcrumb-link:hover {
  color: var(--primary-color);
  background: var(--hover-background-color);
}

.breadcrumb-link.active {
  color: var(--main-text-color);
  cursor: default;
  font-weight: 600;
  background: none;
}

.separator {
  margin: 0 4px;
  color: var(--border-color);
}

/* Table */
.table-responsive {
  overflow-x: auto;
  width: 100%;
}

.files-table {
  width: 100%;
  border-collapse: collapse;
  min-width: 600px;
}

.files-table th {
  text-align: left;
  padding: 12px 16px;
  font-weight: 600;
  color: var(--secondary-text-color);
  border-bottom: 1px solid var(--border-color);
  background: var(--background-color);
  font-size: 0.85rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  position: sticky;
  top: 0;
}

.files-table td {
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-color);
  color: var(--main-text-color);
  vertical-align: middle;
}

.list-item {
  cursor: pointer;
  transition: background-color 0.15s;
}

.list-item:hover {
  background-color: var(--hover-background-color);
}

.list-item:last-child td {
  border-bottom: none;
}

/* Columns */
.col-icon {
  width: 48px;
  text-align: center;
}

.col-name {
  width: 40%;
}

.col-date, .col-size {
  width: 20%;
  color: var(--secondary-text-color);
  font-size: 0.95rem;
}

.col-actions {
  width: 60px;
  text-align: right;
}

/* Icons & text */
.file-icon {
  width: 24px;
  height: 24px;
  display: block;
}

.name-text {
  font-weight: 500;
}

.folder-item .name-text {
  font-weight: 600;
}

/* Action Buttons */
.action-btn {
  background: none;
  border: none;
  color: var(--secondary-text-color);
  cursor: pointer;
  padding: 8px;
  border-radius: 50%;
  transition: all 0.2s;
}

.download-btn:hover {
  color: var(--primary-color);
  background: rgba(99, 102, 241, 0.1);
}

/* Empty State */
.empty-state {
  text-align: center;
  padding: 40px !important;
  color: var(--secondary-text-color);
  font-style: italic;
}
</style>
