<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
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

    <!-- FAB -->
    <button class="mobile-nav-fab" @click="toggleFabMenu">
      <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" width="28" height="28"
           :style="{ transform: showFabMenu ? 'rotate(45deg)' : 'none', transition: 'transform 0.2s' }">
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
      <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" class="mobile-nav-icon" stroke="currentColor" stroke-width="2">
        <line x1="22" y1="2" x2="11" y2="13"></line>
        <polygon points="22 2 15 22 11 13 2 9 22 2" fill="currentColor" stroke="none"/>
      </svg>
      <span>P2P</span>
    </button>

    <!-- FAB backdrop -->
    <Transition name="fab-backdrop">
      <div v-if="showFabMenu" class="fab-backdrop" @click="showFabMenu = false" />
    </Transition>

    <!-- FAB menu -->
    <Transition name="fab-menu">
      <div v-if="showFabMenu" class="fab-menu">
        <button class="fab-menu-item" @click="navigateTo('/dashboard/friends'); showFabMenu = false">
          <div class="fab-menu-icon friends-icon">
            <svg viewBox="0 0 24 24" width="22" height="22" fill="currentColor">
              <path d="M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z"/>
            </svg>
          </div>
          <span>Mes amis</span>
        </button>
        <button class="fab-menu-item" @click="triggerUpload">
          <div class="fab-menu-icon upload-icon">
            <svg viewBox="0 0 24 24" width="22" height="22" fill="currentColor">
              <path d="M19.35 10.04A7.49 7.49 0 0 0 12 4C9.11 4 6.6 5.64 5.35 8.04A5.994 5.994 0 0 0 0 14c0 3.31 2.69 6 6 6h13c2.76 0 5-2.24 5-5 0-2.64-2.05-4.78-4.65-4.96zM14 13v4h-4v-4H7l5-5 5 5h-3z"/>
            </svg>
          </div>
          <span>Importer un fichier</span>
        </button>
        <button class="fab-menu-item" @click="triggerCreateFolder">
          <div class="fab-menu-icon folder-icon">
            <svg viewBox="0 0 24 24" width="22" height="22" fill="currentColor">
              <path d="M20 6h-8l-2-2H4c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2zm-1 8h-3v3h-2v-3h-3v-2h3V9h2v3h3v2z"/>
            </svg>
          </div>
          <span>Créer un dossier</span>
        </button>
      </div>
    </Transition>
  </nav>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUIStore } from '../../stores/ui'

const router = useRouter()
const route = useRoute()
const uiStore = useUIStore()

const showFabMenu = ref(false)

const navigateTo = (path) => router.push(path)
const isActive = (path) => route.path.startsWith(path)

const toggleFabMenu = () => {
  showFabMenu.value = !showFabMenu.value
}

const triggerUpload = async () => {
  showFabMenu.value = false
  if (!route.path.startsWith('/dashboard/files')) {
    await router.push('/dashboard/files')
  }
  uiStore.pendingMobileAction = 'upload'
}

const triggerCreateFolder = async () => {
  showFabMenu.value = false
  if (!route.path.startsWith('/dashboard/files')) {
    await router.push('/dashboard/files')
  }
  uiStore.pendingMobileAction = 'createFolder'
}

</script>

<style scoped>
.mobile-bottom-nav {
  display: none;
}

@media (max-width: 768px) {
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
    z-index: 910;
    position: relative;
  }

  .mobile-nav-fab:active {
    transform: scale(0.95);
  }

  /* Backdrop */
  .fab-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.35);
    z-index: 905;
  }

  .fab-backdrop-enter-active,
  .fab-backdrop-leave-active {
    transition: opacity 0.2s;
  }
  .fab-backdrop-enter-from,
  .fab-backdrop-leave-to {
    opacity: 0;
  }

  /* FAB menu */
  .fab-menu {
    position: fixed;
    bottom: 80px;
    left: 50%;
    transform: translateX(-50%);
    display: flex;
    flex-direction: column;
    gap: 12px;
    z-index: 910;
    width: 220px;
  }

  .fab-menu-enter-active,
  .fab-menu-leave-active {
    transition: opacity 0.2s, transform 0.2s;
  }
  .fab-menu-enter-from,
  .fab-menu-leave-to {
    opacity: 0;
    transform: translateX(-50%) translateY(16px);
  }

  .fab-menu-item {
    display: flex;
    align-items: center;
    gap: 14px;
    background: var(--card-color);
    border: 1px solid var(--border-color);
    border-radius: 14px;
    padding: 14px 18px;
    cursor: pointer;
    font-size: 0.92rem;
    font-weight: 600;
    color: var(--main-text-color);
    box-shadow: 0 4px 16px rgba(0,0,0,0.12);
    transition: background 0.15s;
    text-align: left;
  }

  .fab-menu-item:active {
    background: var(--hover-background-color);
  }

  .fab-menu-icon {
    width: 42px;
    height: 42px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .upload-icon {
    background: rgba(99, 102, 241, 0.12);
    color: var(--primary-color);
  }

  .folder-icon {
    background: rgba(245, 158, 11, 0.12);
    color: #f59e0b;
  }

  .friends-icon {
    background: rgba(99, 102, 241, 0.12);
    color: #6366f1;
  }

}
</style>
