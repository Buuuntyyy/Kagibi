<template>
  <div class="file-browser-wrapper">
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
</script>

<style scoped>
.file-browser-wrapper {
  width: 100%;
  height: 100%;
}
</style>
