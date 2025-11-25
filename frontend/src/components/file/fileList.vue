<template>
  <h2 class="dashboard-title">Dashboard</h2>
  <div class="file-list-container">
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
      <span>{{ fileStore.currentPath }}</span>
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
        <span>{{ file.Name }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useFileStore } from '../../stores/files'

const fileStore = useFileStore()
const selectedFiles = ref([])
const fileInput = ref(null)

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
  fileStore.downloadFile(file.ID, file.Name);
}

const triggerFileInput = () => {
  fileInput.value.click()
}

const handleFileUpload = (event) => {
  const file = event.target.files[0]
  if (file) {
    fileStore.uploadFile(file)
  }
}

const createNewFolder = () => {
  const folderName = prompt("Entrez le nom du nouveau dossier :")
  if (folderName) {
    fileStore.createFolder(folderName)
  }
}
</script>

<style scoped>
.file-list-container {
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
</style>
