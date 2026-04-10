<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="file-browser-wrapper">
    <file-list />
  </div>
</template>

<script setup>
import { useFileStore } from '../stores/files'
import FileList from './file/fileList.vue'

const fileStore = useFileStore()

// FileList already handles fetching initial items
</script>

<style scoped>
.file-browser-wrapper {
  width: 100%;
  height: 100%;
}
</style>
