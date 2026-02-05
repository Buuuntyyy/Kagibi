<template>
  <div class="app-container">
    <Navbar v-if="!isLandingPage" />
    <main :class="isLandingPage ? 'landing-content' : 'content'">
      <router-view />
    </main>
    <P2PTransferDialog v-if="!isLandingPage" />
    <WarnDialog v-if="!isLandingPage" />
    <DeleteConfirmDialog v-if="!isLandingPage" />
    <UploadManager v-if="!isLandingPage" />
    <DownloadManager v-if="!isLandingPage" />
  </div>
</template>

<script setup>
import Navbar from './components/layout/navbar.vue'
import P2PTransferDialog from './components/P2PTransferDialog.vue'
import WarnDialog from './components/WarnDialog.vue'
import DeleteConfirmDialog from './components/DeleteConfirmDialog.vue'
import UploadManager from './components/upload/UploadManager.vue'
import DownloadManager from './components/download/DownloadManager.vue'
import { useThemeStore } from './stores/theme'
import { useAuthStore } from './stores/auth'
import { useBillingStore } from './stores/billing'
import { useRealtimeStore } from './stores/realtime'
import { watch, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'

const themeStore = useThemeStore()
const authStore = useAuthStore()
const billingStore = useBillingStore()
const realtimeStore = useRealtimeStore()
const route = useRoute()

// Check if current route is a landing page
const isLandingPage = computed(() => {
  return ['LandingHome', 'Pricing', 'Transfer'].includes(route.name)
})

// Fetch billing status on app mount
onMounted(() => {
  billingStore.fetchBillingStatus()
})

// Connect Supabase Realtime when authenticated
watch(() => authStore.isAuthenticated, (isAuthenticated) => {
  if (isAuthenticated) {
    realtimeStore.connect()
  } else {
    realtimeStore.disconnect()
  }
}, { immediate: true })
</script>

<style>
/* Variables CSS globales */
:root {
  --primary-color: #0050FF;
  --primary-dark: #0040CC;
  --bg-dark: #121212;
  --bg-card: #1a1a1a;
  --text-primary: #ffffff;
  --text-secondary: #a0a0a0;
  --border-color: #2a2a2a;
  --radius: 4px;
}

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
  background: var(--bg-dark);
  color: var(--text-primary);
  line-height: 1.6;
}

.app-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.content {
  padding-top: 60px; /* Hauteur de la navbar */
  flex-grow: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.landing-content {
  min-height: 100vh;
  width: 100%;
}
</style>
