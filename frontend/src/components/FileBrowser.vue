<template>
  <div>
    <h2>Explorateur de fichiers</h2>
    <p>Chemin actuel : {{ fileStore.currentPath }}</p>
    <button @click="goUp" :disabled="fileStore.currentPath === '/'">Dossier parent</button>
    <file-list />
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useFileStore } from '../stores/files'
import FileList from './file/fileList.vue'

const fileStore = useFileStore()

onMounted(() => {
  fileStore.fetchItems('/')
})

function goUp() {
  fileStore.navigateUp()
}
</script>

<style scoped>
button {
  margin-bottom: 1rem;
  padding: 0.5rem 1rem;
  background-color: #f0f0f0;
  border: 1px solid #ccc;
  border-radius: 4px;
  cursor: pointer;
}

button:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}
</style>
