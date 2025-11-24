<template>
  <div class="file-list-container">
    <div class="toolbar">
      <button class="btn-add-file">Ajouter un fichier</button>
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
  padding: 1rem;
  background-color: #f9f9f9;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 1rem;
  border-bottom: 1px solid #eee;
}

.list-area {
  margin-top: 1rem;
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
</style>
