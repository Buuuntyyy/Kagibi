<template>
  <div class="file-list-container">
    <div class="path-banner">
      <span @click="goUp" class="back-arrow" :class="{ 'disabled': store.currentPath === '/' }">
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M19 12H5M12 19l-7-7 7-7"/>
        </svg>
      </span>
      <div class="breadcrumbs">
        <span v-for="(segment, index) in pathSegments" :key="index" class="breadcrumb-segment">
          <span 
            class="breadcrumb-link" 
            :class="{ 'current': index === pathSegments.length - 1 }"
            @click="navigateToPath(segment.path)"
          >
            {{ segment.name }}
          </span>
          <span v-if="index < pathSegments.length - 1" class="separator">/</span>
        </span>
      </div>
    </div>

    <div class="table-responsive">
      <table class="files-table">
        <thead>
          <tr>
            <th class="icon-col"></th>
            <th class="name-cell">Nom</th>
            <th>Modifié le</th>
            <th>Taille</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <!-- Folders -->
          <tr v-for="folder in store.folders" :key="folder.ID" 
              class="list-item folder-item" 
              @dblclick="openFolder(folder.Name)">
            <td class="icon-col">
              <span class="icon">
                <svg class="icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M10 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z" fill="#5f6368"/>
                </svg>
              </span>
            </td>
            <td class="name-cell">
              <span class="name">{{ folder.Name }}</span>
            </td>
            <td>{{ formatDate(folder.CreatedAt) }}</td>
            <td class="size">-</td>
            <td></td>
          </tr>

          <!-- Files -->
          <tr v-for="file in store.files" :key="file.ID" 
              class="list-item"
              @dblclick="downloadFile(file)">
            <td class="icon-col">
              <span class="icon">
                <!-- PDF -->
                <svg v-if="getFileType(file.Name) === 'pdf'" class="icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                   <path d="M20 2H8c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm-8.5 7.5c0 .83-.67 1.5-1.5 1.5H9v2H7.5V7H10c.83 0 1.5.67 1.5 1.5v1zm5 2c0 .83-.67 1.5-1.5 1.5h-2.5V7H15c.83 0 1.5.67 1.5 1.5v3zm4-3H19v1h1.5V11H19v2h-1.5V7h3v1.5zM9 9.5h1v-1H9v1zM4 6H2v14c0 1.1.9 2 2 2h14v-2H4V6zm10 5.5h1v-3h-1v3z" fill="#ea4335"/>
                </svg>
                <!-- Word -->
                <svg v-else-if="getFileType(file.Name) === 'word'" class="icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                   <path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z" fill="#4285f4"/>
                </svg>
                <!-- Excel -->
                <svg v-else-if="getFileType(file.Name) === 'excel'" class="icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                   <path d="M20 2H4c-1.1 0-2 .9-2 2v16c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zM8 20H4v-4h4v4zm0-6H4v-4h4v4zm0-6H4V4h4v4zm6 12h-4v-4h4v4zm0-6h-4v-4h4v4zm0-6h-4V4h4v4zm6 12h-4v-4h4v4zm0-6h-4v-4h4v4zm0-6h-4V4h4v4z" fill="#0f9d58"/>
                </svg>
                <!-- PowerPoint -->
                <svg v-else-if="getFileType(file.Name) === 'powerpoint'" class="icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                   <path d="M10 8v8l5-4-5-4zm9-5H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm0 16H5V5h14v14z" fill="#f4b400"/>
                </svg>
                <!-- Image -->
                <svg v-else-if="getFileType(file.Name) === 'image'" class="icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                   <path d="M21 19V5c0-1.1-.9-2-2-2H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2zM8.5 13.5l2.5 3.01L14.5 12l4.5 6H5l3.5-4.5z" fill="#db4437"/>
                </svg>
                <!-- Video -->
                <svg v-else-if="getFileType(file.Name) === 'video'" class="icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                   <path d="M18 4l2 4h-3l-2-4h-2l2 4h-3l-2-4H8l2 4H7L5 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V4h-4z" fill="#db4437"/>
                </svg>
                <!-- Text -->
                <svg v-else-if="getFileType(file.Name) === 'text'" class="icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                   <path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z" fill="#5f6368"/>
                </svg>
                <!-- Default -->
                <svg v-else class="icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                   <path d="M6 2c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6H6zm7 7V3.5L18.5 9H13z" fill="#5f6368"/>
                </svg>
              </span>
            </td>
            <td class="name-cell">
              <span class="name">{{ file.Name }}</span>
            </td>
            <td>{{ formatDate(file.UpdatedAt) }}</td>
            <td class="size">{{ formatSize(file.Size) }}</td>
            <td class="actions">
              <button @click.stop="downloadFile(file)" class="btn-download-icon" title="Télécharger">
                <svg xmlns="http://www.w3.org/2000/svg" height="24px" viewBox="0 0 24 24" width="24px" fill="#5f6368"><path d="M0 0h24v24H0V0z" fill="none"/><path d="M19 9h-4V3H9v6H5l7 7 7-7zm-8 2V5h2v6h1.17L12 13.17 9.83 11H11zm-6 7h14v2H5v-2z"/></svg>
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
  const segments = [{ name: 'Contenu partagé', path: '/' }];

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

const getFileType = (filename) => {
  if (!filename) return 'default'
  const ext = filename.split('.').pop().toLowerCase()
  if (['pdf'].includes(ext)) return 'pdf'
  if (['doc', 'docx', 'odt', 'rtf'].includes(ext)) return 'word'
  if (['xls', 'xlsx', 'csv', 'ods'].includes(ext)) return 'excel'
  if (['ppt', 'pptx', 'odp'].includes(ext)) return 'powerpoint'
  if (['jpg', 'jpeg', 'png', 'gif', 'webp', 'svg', 'bmp', 'tiff'].includes(ext)) return 'image'
  if (['mp4', 'avi', 'mov', 'mkv', 'webm', 'flv', 'wmv'].includes(ext)) return 'video'
  if (['txt', 'md', 'json', 'xml', 'log', 'ini', 'yaml', 'yml'].includes(ext)) return 'text'
  return 'default'
}
</script>

<style scoped>
.file-list-container {
  margin-top: 0.8rem;
  background-color: var(--card-color);
  width: 100%;
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  color: var(--main-text-color);
}

.table-responsive {
  overflow-x: auto;
  width: 100%;
}

.files-table {
  width: 100%;
  border-collapse: collapse;
}

.files-table th, .files-table td {
  padding: 0.75rem 1rem;
  text-align: left;
  border-bottom: 1px solid var(--border-color);
  white-space: nowrap;
}

.files-table th {
  font-weight: 500;
  color: var(--secondary-text-color);
  font-size: 0.9em;
}

.list-item {
  cursor: pointer;
  transition: background-color 0.2s;
}

.list-item:hover {
  background-color: var(--hover-background-color);
}

.icon-col {
  width: 40px;
  padding-right: 0.5rem;
  text-align: center;
}

.name-cell {
  width: 100%;
}

.icon-svg {
  width: 24px;
  height: 24px;
  vertical-align: middle;
}

.name {
  font-weight: 500;
  color: var(--main-text-color);
}

.size {
  color: var(--secondary-text-color);
  font-size: 0.9em;
}

.btn-download-icon {
  background: none;
  border: none;
  cursor: pointer;
  padding: 4px;
  border-radius: 50%;
  transition: background-color 0.2s;
}

.btn-download-icon:hover {
  background-color: rgba(0,0,0,0.1);
}

.path-banner {
  padding: 0.5rem 1rem;
  background-color: var(--card-color);
  border-bottom: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  gap: 1rem;
}

.back-arrow {
  cursor: pointer;
  color: var(--secondary-text-color);
  display: flex;
  align-items: center;
}

.back-arrow.disabled {
  color: var(--border-color);
  cursor: not-allowed;
}

.breadcrumbs {
  display: flex;
  align-items: center;
  font-size: 1.2rem;
}

.breadcrumb-link {
  cursor: pointer;
  color: var(--primary-color);
  text-decoration: none;
}

.breadcrumb-link.current {
  color: var(--main-text-color);
  cursor: default;
  font-weight: bold;
}

.separator {
  margin: 0 0.5rem;
  color: var(--secondary-text-color);
}
</style>

