<template>
  <Navbar />
  <main class="content">
    <router-view />
  </main>
  <P2PTransferDialog />
  <WarnDialog />
  <DeleteConfirmDialog />
  <UploadManager />
</template>

<script setup>
import Navbar from './components/layout/navbar.vue'
import P2PTransferDialog from './components/P2PTransferDialog.vue'
import WarnDialog from './components/WarnDialog.vue'
import DeleteConfirmDialog from './components/DeleteConfirmDialog.vue'
import UploadManager from './components/upload/UploadManager.vue'
import { useThemeStore } from './stores/theme'
import { useAuthStore } from './stores/auth'
import { useWebSocketStore } from './stores/websocket'
import { watch } from 'vue'

const themeStore = useThemeStore()
const authStore = useAuthStore()
const wsStore = useWebSocketStore()

// Connect WebSocket when authenticated
watch(() => authStore.isAuthenticated, (isAuthenticated) => {
  if (isAuthenticated) {
    wsStore.connect()
  } else {
    wsStore.disconnect()
  }
}, { immediate: true })
</script>

<style>
.content {
  padding-top: 60px; /* Hauteur de la navbar */
  flex-grow: 1;
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
}
</style>
