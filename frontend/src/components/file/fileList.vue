<template>
  <h2 class="dashboard-title">Dashboard</h2>
  <div class="file-list-container"
       @dragover.prevent="onDragOver"
       @dragleave.prevent="onDragLeave"
       @drop.prevent="onDrop">
       
    <div v-if="isDragging" class="drag-overlay">
      <div class="drag-content">
        <span class="drag-icon">☁️</span>
        <span class="drag-text">Déposez vos fichiers ici</span>
      </div>
    </div>

    <div class="toolbar">
      <div class="toolbar-left">
        <button @click="triggerFileInput" class="btn-add-file">Ajouter un fichier</button>
        <button @click="createNewFolder" class="btn-add-file">Créer un dossier</button>
        <input type="file" ref="fileInput" @change="handleFileUpload" style="display: none" />
      </div>
      <div class="toolbar-right">
        <button @click="downloadSelectedFiles" :disabled="selectedFiles.length === 0" class="btn-download">
          Télécharger
        </button>
        <button @click="deleteSelectedItems" :disabled="selectedFiles.length === 0" class="btn-delete">
          Supprimer
        </button>
      </div>
    </div>
    <div class="path-banner">
      <span @click="goUp" class="back-arrow" :class="{ 'disabled': fileStore.currentPath === '/' }">←</span>
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
    
    <!-- Upload Progress Popup -->
    <div v-if="fileStore.isUploading" class="upload-popup">
      <div class="popup-header">
        <span class="popup-title">Upload en cours...</span>
        <button @click="closeUploadPopup" class="btn-close">×</button>
      </div>
      <div class="popup-content">
        <div class="file-name" :title="fileStore.uploadingFileName">{{ fileStore.uploadingFileName }}</div>
        <div class="progress-container-popup">
          <div class="progress-bar">
            <div class="progress-fill" :style="{ width: fileStore.uploadProgress + '%' }"></div>
          </div>
          <span class="progress-text">{{ fileStore.uploadProgress }}%</span>
        </div>
      </div>
    </div>

    <div class="list-area">
      <!-- Folders -->
      <div v-for="folder in fileStore.folders" :key="folder.ID" class="list-item folder-item" @click="openFolder(folder.Name)">
        <span class="icon">📁</span>
        <span class="name">{{ folder.Name }}</span>
      </div>
      <!-- Files -->
      <div v-for="file in fileStore.files" :key="file.ID" 
          class="list-item"
          :class="{ selected: isSelected(file) }"
          @click="selectFile(file, $event)"
          @dblclick="downloadFile(file)">
        <span class="icon">📄</span>
        <span class="name">{{ file.Name }}</span>
        <span class="size">{{ formatSize(file.Size) }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref, computed } from 'vue'
import { useFileStore } from '../../stores/files'

const fileStore = useFileStore()
const selectedFiles = ref([])
const fileInput = ref(null)
const isDragging = ref(false)

const pathSegments = computed(() => {
  const path = fileStore.currentPath
  const segments = [{ name: 'Home', path: '/' }]
  
  if (path === '/') return segments

  const parts = path.split('/').filter(p => p)
  let currentBuild = ''
  
  parts.forEach(part => {
    currentBuild += '/' + part
    segments.push({ name: part, path: currentBuild })
  })
  
  return segments
})

const navigateToPath = (path) => {
  if (path === fileStore.currentPath) return
  selectedFiles.value = []
  fileStore.fetchItems(path)
}

onMounted(() => {
  fileStore.fetchItems('/')
})

const selectFile = (file) => {
  const isSelected = selectedFiles.value.some(f => f.ID === file.ID);
  if (!event.ctrlKey && !event.metaKey) { // si ctrl ou cmd n'est pas enfoncé
    selectedFiles.value = isSelected ? [] : [file]; // Select only this file
  } else { // si ctrl ou cmd est enfoncé
    if (isSelected) {
      selectedFiles.value = selectedFiles.value.filter(f => f.ID !== file.ID); // Deselect file
    } else {
      selectedFiles.value.push(file); // Add to selection
    }
  }
}

// Helper pour vérifier si un item est sélectionné (utile pour le template)
const isSelected = (file) => {
  return selectedFiles.value.some(f => f.ID === file.ID);
}

const openFolder = (folderName) => {
  selectedFiles.value = [] // Deselect file when navigating
  fileStore.navigateTo(folderName)
}

const goUp = () => {
  if (fileStore.currentPath !== '/') {
    selectedFiles.value = [] // Deselect file when navigating up
    fileStore.navigateUp()
  }
}

const downloadSelectedFiles = () => {
  if (selectedFiles.value.length === 0) return;

  if (selectedFiles.value.length === 1) {
    const file = selectedFiles.value[0];
    fileStore.downloadFile(file.ID, file.Name);
  } else {
    // Logic for downloading multiple files, e.g., zipping them first
    alert("Le téléchargement de plusieurs fichiers en une fois (ex: zip) n'est pas encore implémenté. Les fichiers seront téléchargés individuellement.");
    selectedFiles.value.forEach(file => {
      fileStore.downloadFile(file.ID, file.Name);
    });
  }
}

const deleteSelectedItems = async () => {
  if (selectedFiles.value.length === 0) return;

  const confirmDelete = confirm(`Êtes-vous sûr de vouloir supprimer les ${selectedFiles.value.length} élément(s) sélectionné(s) ?`);
  if (confirmDelete) {
    const fileIDs = selectedFiles.value.map(file => file.ID);
    await fileStore.deleteFiles(fileIDs);
    selectedFiles.value = [] // Clear selection after deletion
  }
}

const downloadFile = (file) => {
  fileStore.downloadFile(file.ID, file.Name, file.MimeType);
}

const triggerFileInput = () => {
  fileInput.value.click()
}

const handleFileUpload = async (event) => {
  const file = event.target.files[0]
  if (file) {
    await fileStore.uploadFile(file)
    event.target.value = '' // Reset file input
  }
}

const createNewFolder = async () => {
  const folderName = prompt("Entrez le nom du nouveau dossier :")
  if (folderName) {
    await fileStore.createFolder(folderName)
  }
}

const formatSize = (bytes) => {
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB']
  const k = 1024
  if (bytes === 0) return '0 Byte'
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const closeUploadPopup = () => {
  fileStore.isUploading = false
}

const onDragOver = (e) => {
  isDragging.value = true
}

const onDragLeave = (e) => {
  // Empêche le clignotement quand on passe sur les enfants
  if (e.currentTarget.contains(e.relatedTarget)) return;
  isDragging.value = false
}

const onDrop = async (e) => {
  isDragging.value = false
  const files = e.dataTransfer.files
  if (files.length > 0) {
    // On upload les fichiers un par un
    for (const file of files) {
      await fileStore.uploadFile(file)
    }
  }
}
</script>

<style scoped>
.file-list-container {
  position: relative;
  border: 1px solid #ccc;
  border-radius: 8px;
  background-color: var(--background-color);
  height: 60vh;
  width: 80vw;
  display: flex;
  flex-direction: column;
}

.path-banner {
  padding: 0.5rem 1rem;
  background-color: var(--background-color);
  border-bottom: 1px solid #ccc;
  display: flex;
  justify-content: flex-start;
  align-items: center;
  gap: 1rem;
  border-top-left-radius: 8px;
  border-top-right-radius: 8px;
}

.back-arrow {
  cursor: pointer;
  font-size: 1.5rem;
  font-weight: bold;
  padding: 0.2rem 0.8rem;
  color: #333;
  border: 1px solid #ccc;
  border-radius: 5px;
  background-color: #f0f0f0;
  transition: background-color 0.2s;
  line-height: 1;
}

.back-arrow:not(.disabled):hover {
  background-color: var(--hover-background-color);
}

.back-arrow.disabled {
  color: #ccc;
  cursor: not-allowed;
  background-color: var(--background-color);
  border-color: #eee;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  border-bottom: 1px solid #eee;
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

.list-area {
  margin-top: 1rem;
  overflow-y: auto;
  flex-grow: 1;
  padding: 0 1rem;
}

.list-item {
  display: flex;
  align-items: center;
  padding: 0.5rem;
  cursor: pointer;
  border-radius: 4px;
  transition: background-color 0.2s;
  user-select: none;
}

.list-item:hover {
  background-color: var(--hover-background-color);
}

.list-item .icon {
  margin-right: 0.5rem;
}
.name {
  flex-grow: 1;
  text-align: left;
}
.size {
  color: #666;
  font-size: 0.9em;
}

.list-item.selected {
  background-color: #42b983;
  color: white;
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
  font-size: 1rem;
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
</style>
