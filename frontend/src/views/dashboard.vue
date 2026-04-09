<template>
  <div class="dashboard-container">
    <LeftBar />
    <div class="main-content">
      <router-view />
    </div>

    <FilePreview
      :visible="fileStore.preview.show"
      :fileUrl="fileStore.preview.url"
      :fileName="fileStore.preview.name"
      :mimeType="fileStore.preview.type"
      :loading="fileStore.preview.loading"
      :status="fileStore.preview.status"
      @close="fileStore.preview.show = false"
    />
  </div>
</template>

<script setup>
import LeftBar from '../components/bar/leftBar.vue'
import FilePreview from '../components/file/FilePreview.vue'
import { useFileStore } from '../stores/files'

const fileStore = useFileStore()
</script>

<style scoped>
.dashboard-container {
  display: flex;
  height: 100%;
  width: 100%;
  box-sizing: border-box;
  background-color: var(--background-color);
}

.friends-wrapper {
  overflow: hidden;
  flex-shrink: 0;
  position: relative;
  display: flex;
}

.friends-wrapper.show {
  /* width handled by style binding */
}

.resizer {
  width: 5px;
  cursor: ew-resize;
  background-color: transparent;
  position: absolute;
  right: 0;
  top: 0;
  bottom: 0;
  z-index: 10;
  transition: background-color 0.2s;
}

.resizer:hover, .friends-wrapper.resizing .resizer {
  background-color: var(--primary-color);
}

.main-content {
  flex-grow: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  border-top-left-radius: 30px;
  background-color: var(--card-color);
}
</style>
