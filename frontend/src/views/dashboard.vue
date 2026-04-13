<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

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

    <!-- Mobile Bottom Navigation -->
    <MobileBottomNav />
  </div>
</template>

<script setup>
import LeftBar from '../components/bar/leftBar.vue'
import MobileBottomNav from '../components/bar/MobileBottomNav.vue'
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

.main-content {
  flex-grow: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  border-top-left-radius: 30px;
  background-color: var(--card-color);
}

@media (max-width: 768px) {
  .dashboard-container {
    flex-direction: column;
    padding-bottom: 64px; /* space for bottom nav */
  }

  .main-content {
    border-top-left-radius: 0;
    border-radius: 0;
    overflow-y: auto;
  }
}
</style>
