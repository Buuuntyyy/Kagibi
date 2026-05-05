<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="public-browse-wrapper">
    <main class="browse-container centered-layout">
      <div v-if="store.isLoading" class="loading-state">
         <div class="spinner"></div>
         <p>Chargement du dossier...</p>
      </div>
      <div v-else-if="store.error" class="error-message">{{ store.error }}</div>
      
      <div v-else class="browse-content fade-in">
        <header class="content-header">
          <span class="header-name">{{ store.resourceName }}</span>
          <span class="header-stats">
            {{ store.folders.length }} dossier{{ store.folders.length !== 1 ? 's' : '' }}, {{ store.files.length }} fichier{{ store.files.length !== 1 ? 's' : '' }}
          </span>
          <span class="header-owner" v-if="store.ownerName || store.ownerEmail">
            Partagé par {{ store.ownerName || store.ownerEmail }}
          </span>
        </header>

        <div class="file-list-card">
          <PublicFileList />
        </div>
      </div>
    </main>

    <Transition name="toast">
      <div v-if="store.toast.visible" class="toast-bar" :class="store.toast.type">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor" style="flex-shrink:0">
          <path v-if="store.toast.type === 'error'" d="M1 21h22L12 2 1 21zm12-3h-2v-2h2v2zm0-4h-2v-4h2v4z"/>
          <path v-else d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
        </svg>
        {{ store.toast.message }}
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { onMounted } from 'vue';
import { useRoute } from 'vue-router';
import { usePublicFileStore } from '../stores/publicFileStore';
import PublicFileList from '../components/PublicFileList.vue';

const route = useRoute();
const store = usePublicFileStore();

onMounted(() => {
  const token = route.params.token;
  let subpath = '/';
  if (route.params.subpath) {
    if (Array.isArray(route.params.subpath)) {
      subpath = `/${route.params.subpath.join('/')}`;
    } else {
      subpath = `/${route.params.subpath}`;
    }
  }
  store.fetchItems(token, subpath);
});
</script>

<style scoped>
.public-browse-wrapper {
  min-height: 100vh;
  background-color: var(--background-color);
  font-family: 'Segoe UI', system-ui, sans-serif;
  color: var(--main-text-color);
}

.centered-layout {
  max-width: 1200px;
  margin: 0 auto;
  padding: 40px 20px;
}

.loading-state {
  text-align: center;
  padding: 60px;
  color: var(--secondary-text-color);
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid rgba(0,0,0,0.1);
  border-radius: 50%;
  border-top-color: var(--primary-color);
  animation: spin 1s ease-in-out infinite;
  margin: 0 auto 20px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.content-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.header-name,
.header-owner,
.header-stats {
  font-size: 0.9rem;
  color: var(--secondary-text-color);
  font-weight: 500;
}

.toast-bar {
  position: fixed;
  bottom: 28px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 20px;
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 500;
  box-shadow: 0 4px 16px rgba(0,0,0,0.18);
  z-index: 9999;
  white-space: nowrap;
}

.toast-bar.error {
  background: var(--error-color, #ef4444);
  color: #fff;
}

.toast-bar.success {
  background: var(--success-color, #22c55e);
  color: #fff;
}

.toast-enter-active, .toast-leave-active {
  transition: opacity 0.25s, transform 0.25s;
}

.toast-enter-from, .toast-leave-to {
  opacity: 0;
  transform: translateX(-50%) translateY(10px);
}

.file-list-card {
  background: var(--card-color);
  border-radius: 12px;
  border: 1px solid var(--border-color);
  overflow: hidden;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.05);
  max-height: 65vh;
  display: flex;
  flex-direction: column;
}

.fade-in {
  animation: fadeIn 0.4s ease-out;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

@media (max-width: 768px) {
  .content-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }
}

@media (max-width: 480px) {
  .share-info-card {
    padding: 12px;
  }

  .download-all-btn {
    width: 100%;
    justify-content: center;
  }
}
</style>
