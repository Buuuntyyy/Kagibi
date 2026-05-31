<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="app-container">
    <template v-if="!isP2PSubdomain">
      <Navbar v-if="!isLandingPage" />
      <main :class="isLandingPage ? 'landing-content' : 'content'">
        <router-view />
      </main>
      <P2PTransferDialog v-if="!isLandingPage" />
      <WarnDialog v-if="!isLandingPage" />
      <DeleteConfirmDialog v-if="!isLandingPage" />
      <ConfirmDialog v-if="!isLandingPage" />
      <GlobalToast />
      <UploadManager v-if="!isLandingPage" />
      <DownloadManager v-if="!isLandingPage" />
    </template>
    <template v-else>
      <main class="landing-content">
        <router-view />
      </main>
    </template>
  </div>
</template>

<script setup>
import Navbar from './components/layout/navbar.vue'
import P2PTransferDialog from './components/P2PTransferDialog.vue'
import WarnDialog from './components/WarnDialog.vue'
import DeleteConfirmDialog from './components/DeleteConfirmDialog.vue'
import UploadManager from './components/upload/UploadManager.vue'
import DownloadManager from './components/download/DownloadManager.vue'
import ConfirmDialog from './components/ConfirmDialog.vue'
import GlobalToast from './components/GlobalToast.vue'
import { useThemeStore } from './stores/theme'
import { useAuthStore } from './stores/auth'
import { useBillingStore } from './stores/billing'
import { useRealtimeStore } from './stores/realtime'
import { useNotificationStore } from './stores/notifications'
import { watch, computed } from 'vue'
import { useRoute } from 'vue-router'
import { isP2PSubdomain } from './composables/useSubdomain'

const themeStore = useThemeStore()
const authStore = useAuthStore()
const billingStore = useBillingStore()
const realtimeStore = useRealtimeStore()
const notifStore = useNotificationStore()
const route = useRoute()

// Check if current route is a landing page
const isLandingPage = computed(() => {
  return ['LandingHome', 'Pricing', 'Transfer', 'Compare', 'Security'].includes(route.name)
})

// Connect realtime services when authenticated.
// The notification subscription is wired here (not inside a navbar component)
// so it survives page navigation and Vite HMR reloads.
watch(() => authStore.isAuthenticated, (isAuthenticated) => {
  if (isAuthenticated) {
    realtimeStore.connect()
    billingStore.fetchBillingStatus()
    notifStore.fetchNotifications()
    notifStore.connectRealtime(realtimeStore)
  } else {
    realtimeStore.disconnect()
    notifStore.disconnectRealtime()
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
  overflow-y: auto;
  overflow-x: hidden;
}

.landing-content {
  min-height: 100vh;
  width: 100%;
}

/* Global mobile resets */
@media (max-width: 768px) {
  /* Prevent iOS from zooming on input focus — requires font-size >= 16px */
  input[type="text"],
  input[type="email"],
  input[type="password"],
  input[type="search"],
  textarea,
  select {
    font-size: 16px !important;
  }
}
</style>
