<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="public-share-wrapper">
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
    </header>

    <div class="content-container">
      <div v-if="loading" class="loading-state">
         <div class="spinner"></div>
         <p>Chargement du partage...</p>
      </div>
      
      <div v-else-if="error" class="error-state">
        <div class="error-icon">
          <svg viewBox="0 0 24 24" width="64" height="64" fill="none" stroke="var(--error-color)" stroke-width="2">
            <circle cx="12" cy="12" r="10"></circle>
            <line x1="12" y1="8" x2="12" y2="12"></line>
            <line x1="12" y1="16" x2="12.01" y2="16"></line>
          </svg>
        </div>
        <h2>Lien introuvable ou expiré</h2>
        <p>{{ error }}</p>
      </div>
      
      <div v-else class="share-card glass-panel">
        <div class="file-preview-section">
          <div class="file-icon-wrapper">
             <span v-if="shareInfo.resource_type === 'folder'">
              <svg class="icon-svg large" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M10 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z" fill="#5f6368"/>
              </svg>
            </span>
            <span v-else>
               <!-- Generic File Icon -->
               <svg class="icon-svg large" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z" fill="var(--primary-color)"/>
               </svg>
            </span>
          </div>
          <div class="file-info-header">
            <h1 class="file-name" :title="shareInfo.resource_name">{{ shareInfo.resource_name }}</h1>
            <p class="share-meta">
              Partagé par <span class="owner-highlight">{{ shareInfo.owner_email }}</span>
            </p>
          </div>
        </div>

        <div class="file-details-grid">
           <div class="detail-item" v-if="shareInfo.resource_type === 'file'">
              <span class="label">Taille</span>
              <span class="value">{{ formatSize(shareInfo.file_size) }}</span>
           </div>
           <div class="detail-item">
              <span class="label">Expire le</span>
              <span class="value">{{ shareInfo.expires_at ? new Date(shareInfo.expires_at).toLocaleDateString() : 'Jamais' }}</span>
           </div>
        </div>
        
        <div class="action-footer">
          <button @click="downloadFile" class="btn-primary-lg">
            <svg viewBox="0 0 24 24" width="24" height="24" fill="currentColor">
              <path d="M19 9h-4V3H9v6H5l7 7 7-7zm-8 2V5h2v6h1.17L12 13.17 9.83 11H11zm-6 7h14v2H5v-2z"/>
            </svg>
            <span>Télécharger</span>
          </button>
        </div>
      </div>
    </div>
    
    <footer class="public-footer">
      <p>Hébergé et sécurisé par <strong>Kagibi</strong></p>
    </footer>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '../api'
import { deriveKeyFromToken, unwrapMasterKey, decryptChunkedFileWorker } from '../utils/crypto'

const route = useRoute()
const router = useRouter()
const shareInfo = ref(null)
const loading = ref(true)
const error = ref(null)

const formatSize = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Number.parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

onMounted(async () => {
  try {
    const token = route.params.token
    const response = await api.get(`/public/share/${token}`)
    shareInfo.value = response.data

    if (shareInfo.value.resource_type === 'folder') {
      router.replace({ name: 'PublicBrowse', params: { token: token, subpath: [] } })
    }

  } catch (err) {
    error.value = err.response?.data?.error || 'Lien invalide ou expiré.'
  } finally {
    loading.value = false
  }
})

const downloadFile = async () => {
  if (!shareInfo.value) return;
  const token = route.params.token;

  try {
    // 1. Derive Share Key from Token
    const shareKey = await deriveKeyFromToken(token);

    // 2. Decrypt File Key
    let fileKey;
    if (shareInfo.value.encrypted_key) {
        try {
          fileKey = await unwrapMasterKey(shareInfo.value.encrypted_key, shareKey);
        } catch(e) {
            console.error("Decryption error:", e);
            alert("Impossible de déchiffrer la clé du fichier.");
            return;
        }
    } else {
        alert("Ce fichier ne peut pas être déchiffré (clé manquante).");
        return;
    }

    // 3. Download Encrypted Blob
    const response = await api.get(`/public/share/${token}/download`, { responseType: 'blob' });
    
    // 4. Decrypt Blob
    const mimeType = shareInfo.value.mime_type || 'application/octet-stream';
    const decryptedBlob = await decryptChunkedFileWorker(response.data, fileKey, mimeType);
    
    // 5. Save
    const url = window.URL.createObjectURL(decryptedBlob);
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', shareInfo.value.resource_name);
    document.body.appendChild(link);
    link.click();
    setTimeout(() => { link.remove(); window.URL.revokeObjectURL(url); }, 100);
    
  } catch (err) {
    console.error("Download failed", err)
    alert("Erreur lors du téléchargement")
  }
}
</script>

<style scoped>
.public-share-wrapper {
  min-height: 100vh;
  background-color: var(--background-color);
  display: flex;
  flex-direction: column;
  align-items: center;
  font-family: 'Segoe UI', system-ui, sans-serif;
  color: var(--main-text-color);
}

.public-header {
  width: 100%;
  padding: 24px 40px;
  display: flex;
  align-items: center;
  background: var(--card-color);
  box-shadow: 0 1px 2px rgba(0,0,0,0.05);
}

.brand {
  display: flex;
  align-items: center;
  gap: 12px;
}

.brand-logo {
  width: 32px;
  height: 32px;
}

.brand-name {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--main-text-color);
  letter-spacing: -0.02em;
}

.content-container {
  flex: 1;
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100%;
  padding: 20px;
}

.share-card {
  background: var(--card-color);
  border-radius: 16px;
  box-shadow: 0 10px 40px -10px rgba(0,0,0,0.1);
  padding: 48px;
  width: 100%;
  max-width: 480px;
  display: flex;
  flex-direction: column;
  gap: 32px;
  border: 1px solid var(--border-color);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.share-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 20px 40px -10px rgba(0,0,0,0.15);
}

.file-preview-section {
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20px;
}

.icon-svg.large {
  width: 96px;
  height: 96px;
  filter: drop-shadow(0 4px 6px rgba(0,0,0,0.05));
}

.file-name {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--main-text-color);
  word-break: break-word;
}

.share-meta {
  margin: 0;
  color: var(--secondary-text-color);
  font-size: 0.95rem;
}

.owner-highlight {
  color: var(--primary-color);
  font-weight: 500;
}

.file-details-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  background: var(--background-color);
  padding: 16px;
  border-radius: 12px;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.detail-item .label {
  font-size: 0.8rem;
  text-transform: uppercase;
  color: var(--secondary-text-color);
  font-weight: 600;
  letter-spacing: 0.05em;
}

.detail-item .value {
  font-size: 1rem;
  font-weight: 500;
  color: var(--main-text-color);
}

.btn-primary-lg {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 14px 24px;
  background: var(--primary-color);
  color: white;
  border: none;
  border-radius: 12px;
  font-size: 1.05rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-primary-lg:hover {
  background: var(--accent-color);
  transform: translateY(-1px);
}

.btn-primary-lg:active {
  transform: translateY(0);
}

.public-footer {
  padding: 24px;
  color: var(--secondary-text-color);
  font-size: 0.9rem;
}

/* Loading & Error States */
.loading-state, .error-state {
  text-align: center;
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

.error-state svg {
  margin-bottom: 20px;
}
</style>
