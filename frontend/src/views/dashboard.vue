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
    <nav class="mobile-bottom-nav">
      <button class="mobile-nav-item" :class="{ active: isActive('/dashboard/home') }" @click="navigateTo('/dashboard/home')">
        <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" class="mobile-nav-icon">
          <path d="M10 20v-6h4v6h5v-8h3L12 3 2 12h3v8z" fill="currentColor"/>
        </svg>
        <span>Accueil</span>
      </button>
      <button class="mobile-nav-item" :class="{ active: isActive('/dashboard/files') }" @click="navigateTo('/dashboard/files')">
        <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" class="mobile-nav-icon">
          <path d="M20 6h-8l-2-2H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2zm0 12H4V8h16v10z" fill="currentColor"/>
        </svg>
        <span>Fichiers</span>
      </button>
      <button class="mobile-nav-fab" @click="navigateTo('/dashboard/files')">
        <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" width="28" height="28">
          <path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z" fill="currentColor"/>
        </svg>
      </button>
      <button class="mobile-nav-item" :class="{ active: isActive('/dashboard/shares') }" @click="navigateTo('/dashboard/shares')">
        <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" class="mobile-nav-icon">
          <path d="M15 8c0-1.42-.5-2.73-1.33-3.76.42-.14.86-.24 1.33-.24 2.21 0 4 1.79 4 4s-1.79 4-4 4c-.43 0-.84-.09-1.23-.21-.03-.01-.06-.02-.1-.03A5.98 5.98 0 0 0 15 8zm1.66 5.13C18.03 14.06 19 15.32 19 17v3h4v-3c0-2.18-3.57-3.47-6.34-3.87zM9 6c-1.1 0-2 .9-2 2s.9 2 2 2 2-.9 2-2-.9-2-2-2m0 9c-2.7 0-5.8 1.29-6 4.02V21h12v-1.98c-.2-2.72-3.3-4.02-6-4.02z" fill="currentColor"/>
        </svg>
        <span>Partages</span>
      </button>
      <button class="mobile-nav-item" :class="{ active: isActive('/p2p') }" @click="navigateTo('/p2p')">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="mobile-nav-icon">
          <line x1="22" y1="2" x2="11" y2="13"></line>
          <polygon points="22 2 15 22 11 13 2 9 22 2"></polygon>
        </svg>
        <span>P2P</span>
      </button>
    </nav>
  </div>
</template>

<script setup>
import LeftBar from '../components/bar/leftBar.vue'
import FilePreview from '../components/file/FilePreview.vue'
import { useFileStore } from '../stores/files'
import { useRouter, useRoute } from 'vue-router'

const fileStore = useFileStore()
const router = useRouter()
const route = useRoute()

const navigateTo = (path) => router.push(path)
const isActive = (path) => route.path.startsWith(path)
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

/* Mobile Bottom Navigation */
.mobile-bottom-nav {
  display: none;
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

  .mobile-bottom-nav {
    display: flex;
    align-items: center;
    justify-content: space-around;
    position: fixed;
    bottom: 0;
    left: 0;
    right: 0;
    height: 64px;
    background: var(--card-color);
    border-top: 1px solid var(--border-color);
    z-index: 900;
    padding: 0 0.25rem;
    box-shadow: 0 -2px 12px rgba(0,0,0,0.08);
  }

  .mobile-nav-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 2px;
    background: none;
    border: none;
    cursor: pointer;
    color: var(--secondary-text-color);
    font-size: 0.65rem;
    font-weight: 500;
    padding: 0.4rem 0.75rem;
    border-radius: 12px;
    transition: color 0.2s;
    flex: 1;
  }

  .mobile-nav-item.active {
    color: var(--primary-color);
  }

  .mobile-nav-icon {
    width: 22px;
    height: 22px;
  }

  .mobile-nav-fab {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 52px;
    height: 52px;
    border-radius: 50%;
    background: var(--primary-color);
    color: white;
    border: none;
    cursor: pointer;
    box-shadow: 0 4px 12px rgba(0, 80, 255, 0.35);
    margin-bottom: 8px;
    flex-shrink: 0;
    transition: transform 0.2s, box-shadow 0.2s;
  }

  .mobile-nav-fab:active {
    transform: scale(0.95);
  }
}
</style>
