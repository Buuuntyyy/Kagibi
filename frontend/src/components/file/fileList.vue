<template>
  <div class="file-list-container">
    <div class="path-banner">
      <span>Chemin : {{ fileStore.currentPath }}</span>
      <button @click="goUp" :disabled="fileStore.currentPath === '/'">Dossier parent</button>
    </div>
    <div class="toolbar">
      <div class="toolbar-left">
        <h2 class="dashboard-title">Dashboard</h2>
        <button class="btn-add-file">Ajouter un fichier</button>
      </div>
      <button v-if="selectedFile" @click="downloadSelectedFile" class="btn-download">Télécharger</button>
    </div>
    <div class="list-area">
      <!-- Folders -->
      <div v-for="folder in fileStore.folders" :key="folder.id" class="list-item folder-item" @click="openFolder(folder.name)">
        <span class="icon">📁</span>
        <span class="name">{{ folder.name }}</span>
      </div>
      <!-- Files -->
      <div v-for="file in fileStore.files" :key="file.id" class="list-item file-item" :class="{ selected: selectedFile && selectedFile.id === file.id }" @click="selectFile(file)">
        <span class="icon">📄</span>
        <span class="name">{{ file.name }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useFileStore } from '../../stores/files'

const fileStore = useFileStore()
const selectedFile = ref(null)

const selectFile = (file) => {
  if (selectedFile.value && selectedFile.value.id === file.id) {
    selectedFile.value = null // Deselect if clicking the same file
  } else {
    selectedFile.value = file
  }
}

const openFolder = (folderName) => {
  selectedFile.value = null // Deselect file when navigating
  fileStore.navigateTo(folderName)
}

const goUp = () => {
  fileStore.navigateUp()
}

const downloadSelectedFile = () => {
  if (selectedFile.value) {
    fileStore.downloadFile(selectedFile.value.id, selectedFile.value.name)
  }
}
</script>

<style scoped>
.file-list-container {
  border: 1px solid #ccc;
  border-radius: 8px;
  background-color: #f9f9f9;
  height: 60vh;
  width: 80vw;
  display: flex;
  flex-direction: column;
}

.path-banner {
  padding: 0.5rem 1rem;
  background-color: #e9ecef;
  border-bottom: 1px solid #ccc;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-top-left-radius: 8px;
  border-top-right-radius: 8px;
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
}

.list-item:hover {
  background-color: #e9e9e9;
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
  background-color: #3498db;
  color: white;
}

.btn-download {
  background-color: #2ecc71;
  color: white;
}

.path-banner button {
    background-color: #f0f0f0;
    border: 1px solid #ccc;
}

.path-banner button:disabled {
    cursor: not-allowed;
    opacity: 0.5;
}
</style>
