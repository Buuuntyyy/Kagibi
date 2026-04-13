<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="public-browse-wrapper">
    <!-- Branding Header -->
    <header class="public-header">
      <div class="brand">
        <svg class="brand-logo" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
           <path d="M19.35 10.04C18.67 6.59 15.64 4 12 4 9.11 4 6.6 5.64 5.35 8.04 2.34 8.36 0 10.91 0 14c0 3.31 2.69 6 6 6h13c2.76 0 5-2.24 5-5 0-2.64-2.05-4.78-4.65-4.96z" fill="url(#brandGradient)"/>
           <defs>
             <linearGradient id="brandGradient" x1="0" y1="0" x2="24" y2="24" gradientUnits="userSpaceOnUse">
               <stop offset="0%" stop-color="var(--primary-color)" />
               <stop offset="100%" stop-color="var(--accent-color)" />
             </linearGradient>
           </defs>
        </svg>
        <span class="brand-name">Kagibi</span>
      </div>
      <div class="header-actions" v-if="store.resourceName">
         <span class="shared-badge">Dossier Partagé</span>
      </div>
    </header>

    <main class="browse-container centered-layout">
      <div v-if="store.isLoading" class="loading-state">
         <div class="spinner"></div>
         <p>Chargement du dossier...</p>
      </div>
      <div v-else-if="store.error" class="error-message">{{ store.error }}</div>
      
      <div v-else class="browse-content fade-in">
        <header class="content-header">
          <div class="header-info">
            <h1>{{ store.resourceName }}</h1>
            <p v-if="store.ownerEmail" class="owner-meta">Partagé par {{ store.ownerEmail }}</p>
          </div>
          <div class="stats" v-if="store.files.length || store.folders.length">
            {{ store.folders.length }} dossiers, {{ store.files.length }} fichiers
          </div>
        </header>

        <div class="file-list-card">
          <PublicFileList />
        </div>
      </div>
    </main>
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

.public-header {
  height: 64px;
  padding: 0 40px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--card-color);
  border-bottom: 1px solid var(--border-color);
  position: sticky;
  top: 0;
  z-index: 10;
}

.brand {
  display: flex;
  align-items: center;
  gap: 12px;
}

.brand-logo {
  width: 28px;
  height: 28px;
}

.brand-name {
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--main-text-color);
}

.shared-badge {
  background: rgba(99, 102, 241, 0.1);
  color: var(--primary-color);
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: 600;
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
  align-items: flex-end;
  margin-bottom: 24px;
}

.header-info h1 {
  margin: 0 0 8px 0;
  font-size: 2rem;
  color: var(--main-text-color);
}

.owner-meta {
  margin: 0;
  color: var(--secondary-text-color);
}

.stats {
  color: var(--secondary-text-color);
  font-size: 0.9rem;
}

.file-list-card {
  background: var(--card-color);
  border-radius: 12px;
  border: 1px solid var(--border-color);
  overflow: hidden;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.05);
}

.fade-in {
  animation: fadeIn 0.4s ease-out;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

@media (max-width: 768px) {
  .public-header {
    padding: 0 20px;
  }
  .content-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .content-title {
    font-size: clamp(1.2rem, 4vw, 2rem);
  }
}

@media (max-width: 480px) {
  .public-header {
    padding: 0 12px;
  }

  .share-info-card {
    padding: 12px;
  }

  .download-all-btn {
    width: 100%;
    justify-content: center;
  }
}
</style>
