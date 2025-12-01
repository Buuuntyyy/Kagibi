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

    <div class="list-header">
      <span class="header-icon"></span>
      <span class="header-name">Nom</span>
      <span class="header-date">Modifié le</span>
      <span class="header-size">Taille</span>
    </div>

    <div class="list-area">
      <!-- Folders -->
      <div v-for="folder in store.folders" :key="folder.ID" 
           class="list-item folder-item" 
           @dblclick="openFolder(folder.Name)">
        <span class="icon">📁</span>
        <span class="name">{{ folder.Name }}</span>
        <span class="date-column">{{ formatDate(folder.CreatedAt) }}</span>
        <span class="size">-</span>
      </div>
      <!-- Files -->
      <div v-for="file in store.files" :key="file.ID" 
          class="list-item"
          @dblclick="downloadFile(file)">
        <span class="icon">📄</span>
        <span class="name">{{ file.Name }}</span>
        <span class="date-column">{{ formatDate(file.UpdatedAt) }}</span>
        <span class="size">{{ formatSize(file.Size) }}</span>
        <span class="actions">
          <button @click.stop="downloadFile(file)" class="btn-download" title="Télécharger">
            📥
          </button>
        </span>
      </div>
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
  if (bytes === 0) return '0 Byte';
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

const formatDate = (dateString) => {
  if (!dateString) return '-';
  const date = new Date(dateString);
  return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
};
</script>

<style scoped>
.file-list-container {
  margin-top: 0.8rem;
  background-color: var(--card-color);
  height: 100%;
  width: 100%;
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
}

.path-banner {
  padding: 0.5rem 1rem;
  background-color: var(--card-color);
  border-bottom: 1px solid #ccc;
  display: flex;
  align-items: center;
  gap: 1rem;
}

.back-arrow {
  cursor: pointer;
  color: #555;
}

.back-arrow.disabled {
  color: #ccc;
  cursor: not-allowed;
}

.breadcrumbs {
  display: flex;
  align-items: center;
  font-size: 1.2rem;
}

.breadcrumb-link {
  cursor: pointer;
  color: var(--primary-color, #42b983);
  text-decoration: none;
}

.breadcrumb-link.current {
  color: #333;
  cursor: default;
  font-weight: bold;
}

.separator {
  margin: 0 0.5rem;
  color: #999;
}

.list-header {
  display: grid;
  grid-template-columns: 40px 3fr 1fr 1fr 50px;
  padding: 0.5rem;
  font-weight: bold;
  border-bottom: 1px solid #ccc;
}

.list-item {
  display: grid;
  grid-template-columns: 40px 3fr 1fr 1fr 50px;
  align-items: center;
  padding: 0.5rem;
  cursor: pointer;
  border-bottom: 1px solid #eee;
}

.list-item:hover {
  background-color: var(--hover-background-color);
}

.icon {
  margin-right: 0.5rem;
}

.name {
  text-align: left;
}

.size, .date-column {
  color: #5c5c5c;
  font-size: 0.9em;
  text-align: right;
}

.actions {
  display: flex;
  justify-content: center;
}

.btn-download {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 1.2rem;
  padding: 0;
  opacity: 0.6;
  transition: opacity 0.2s;
}

.btn-download:hover {
  opacity: 1;
}
</style>
