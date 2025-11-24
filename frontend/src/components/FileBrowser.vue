<template>
  <div>
    <h2>File Browser</h2>
    <p>Current Path: {{ fileStore.currentPath }}</p>
    <button @click="goUp" :disabled="fileStore.currentPath === '/'">Go Up</button>
    <ul>
      <li v-for="folder in fileStore.folders" :key="folder.id" @click="openFolder(folder.name)">
        [D] {{ folder.name }}
      </li>
      <li v-for="file in fileStore.files" :key="file.id">
        [F] {{ file.name }}
      </li>
    </ul>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useFileStore } from '../stores/files'

const fileStore = useFileStore()

onMounted(() => {
  fileStore.fetchItems('/')
})

function openFolder(folderName) {
  fileStore.navigateTo(folderName)
}

function goUp() {
  fileStore.navigateUp()
}
</script>

<style scoped>
ul {
  list-style-type: none;
  padding: 0;
}
li {
  cursor: pointer;
  padding: 5px;
}
li:hover {
  background-color: #f0f0f0;
}
</style>
