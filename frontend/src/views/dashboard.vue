<template>
  <div class="dashboard-container">
    <FileBrowser />
  </div>
</template>

<script setup>
import FileBrowser from '../components/FileBrowser.vue'
</script>

<style scoped>
.dashboard-container {
  display: flex;
  justify-content: center;
  padding: 2rem;
  height: 100%;
  box-sizing: border-box;
}
</style>
