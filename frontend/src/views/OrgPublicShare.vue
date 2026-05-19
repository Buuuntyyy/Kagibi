<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="public-share-wrapper">
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

      <div v-else-if="noKey" class="error-state">
        <div class="error-icon">
          <svg viewBox="0 0 24 24" width="64" height="64" fill="none" stroke="var(--warning-color, #f59e0b)" stroke-width="2">
            <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
            <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
          </svg>
        </div>
        <h2>Clé de déchiffrement manquante</h2>
        <p>Le lien est incomplet — la clé de déchiffrement doit être présente dans le fragment # de l'URL.</p>
      </div>

      <div v-else-if="passwordRequired" class="share-card glass-panel password-card">
        <div class="file-preview-section">
          <div class="file-icon-wrapper">
            <svg viewBox="0 0 24 24" width="64" height="64" fill="none" stroke="var(--primary-color)" stroke-width="1.5">
              <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
              <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
            </svg>
          </div>
          <h2 class="file-name">Partage protégé</h2>
          <p class="share-meta">Un mot de passe est requis pour accéder à ce partage.</p>
        </div>
        <form @submit.prevent="submitPassword" class="password-form">
          <input
            type="password"
            v-model="enteredPassword"
            placeholder="Mot de passe"
            class="password-input"
            :class="{ 'input-error': passwordError }"
            autofocus
            autocomplete="current-password"
          />
          <p v-if="passwordError" class="password-error-msg">Mot de passe incorrect.</p>
          <button type="submit" class="btn-primary-lg" :disabled="passwordLoading">
            <span v-if="passwordLoading">Vérification...</span>
            <span v-else>Déverrouiller</span>
          </button>
        </form>
      </div>

      <div v-else-if="shareInfo" class="share-card glass-panel">
        <div class="file-preview-section">
          <div class="file-icon-wrapper">
            <svg class="icon-svg large" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z" fill="var(--primary-color)"/>
            </svg>
          </div>
          <div class="file-info-header">
            <h1 class="file-name" :title="shareInfo.resource_name">{{ shareInfo.resource_name }}</h1>
            <p class="share-meta">Fichier partagé via Kagibi</p>
          </div>
        </div>

        <div class="file-details-grid">
          <div class="detail-item">
            <span class="label">Taille</span>
            <span class="value">{{ formatSize(shareInfo.file_size) }}</span>
          </div>
          <div class="detail-item" v-if="shareInfo.expires_at">
            <span class="label">Expire le</span>
            <span class="value">{{ formatDate(shareInfo.expires_at) }}</span>
          </div>
        </div>

        <div class="e2e-badge">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor">
            <path d="M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zm-6 9c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2zm3.1-9H8.9V6c0-1.71 1.39-3.1 3.1-3.1 1.71 0 3.1 1.39 3.1 3.1v2z"/>
          </svg>
          Chiffré de bout en bout — votre fichier sera déchiffré localement dans votre navigateur
        </div>

        <div v-if="downloadError" class="download-error">{{ downloadError }}</div>

        <div class="action-row">
          <button
            class="btn-primary-lg"
            :disabled="downloading"
            @click="downloadFile"
          >
            <span v-if="downloading">
              <span class="spinner-inline"></span>
              {{ downloadProgress > 0 ? `Déchiffrement… ${downloadProgress}%` : 'Téléchargement…' }}
            </span>
            <span v-else>
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor" style="vertical-align:middle;margin-right:6px">
                <path d="M19 9h-4V3H9v6H5l7 7 7-7zM5 18v2h14v-2H5z"/>
              </svg>
              Télécharger
            </span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import api from '../api'
import { decryptFileFromOrg } from '../utils/orgCrypto.js'

const route = useRoute()
const token = route.params.token

const loading = ref(true)
const error = ref('')
const noKey = ref(false)
const shareInfo = ref(null)
const downloading = ref(false)
const downloadProgress = ref(0)
const downloadError = ref('')

// Password gate state
const passwordRequired = ref(false)
const enteredPassword = ref('')
const passwordError = ref(false)
const passwordLoading = ref(false)
let confirmedPassword = ''

// The share key lives in the URL fragment (never sent to the server).
let shareKeyB64 = ''
let shareKeyCrypto = null

async function fetchShareInfo(password = '') {
  const headers = password ? { 'X-Share-Password': password } : {}
  const { data } = await api.get(`/public/org-share/${token}`, { headers })
  return data
}

onMounted(async () => {
  shareKeyB64 = window.location.hash.slice(1)
  if (!shareKeyB64) {
    noKey.value = true
    loading.value = false
    return
  }

  try {
    const raw = Uint8Array.from(atob(shareKeyB64.replace(/-/g, '+').replace(/_/g, '/')), c => c.charCodeAt(0))
    shareKeyCrypto = await crypto.subtle.importKey('raw', raw, { name: 'AES-GCM' }, true, ['encrypt', 'decrypt', 'wrapKey', 'unwrapKey'])
  } catch {
    error.value = 'Clé de déchiffrement invalide dans le fragment URL.'
    loading.value = false
    return
  }

  try {
    const data = await fetchShareInfo()
    shareInfo.value = data
  } catch (e) {
    if (e.response?.status === 401 && e.response?.data?.error === 'password_required') {
      passwordRequired.value = true
    } else {
      error.value = e.response?.data?.error || e.message
    }
  } finally {
    loading.value = false
  }
})

const submitPassword = async () => {
  if (!enteredPassword.value) return
  passwordLoading.value = true
  passwordError.value = false
  try {
    const data = await fetchShareInfo(enteredPassword.value)
    confirmedPassword = enteredPassword.value
    shareInfo.value = data
    passwordRequired.value = false
  } catch (e) {
    if (e.response?.status === 401) {
      passwordError.value = true
    } else {
      error.value = e.response?.data?.error || e.message
      passwordRequired.value = false
    }
  } finally {
    passwordLoading.value = false
  }
}

const downloadFile = async () => {
  if (!shareKeyCrypto || !shareInfo.value) return
  downloading.value = true
  downloadError.value = ''
  downloadProgress.value = 0

  try {
    const headers = confirmedPassword ? { 'X-Share-Password': confirmedPassword } : {}
    // Download the encrypted blob
    const response = await api.get(`/public/org-share/${token}/download`, {
      responseType: 'blob',
      headers,
      onDownloadProgress: (ev) => {
        if (ev.total) downloadProgress.value = Math.round((ev.loaded / ev.total) * 50)
      },
    })

    downloadProgress.value = 50

    // Decrypt the file
    const encryptedBlob = response.data
    const decryptedBlob = await decryptFileFromOrg(
      encryptedBlob,
      shareInfo.value.encrypted_key,
      shareKeyCrypto,
      shareInfo.value.mime_type,
    )

    downloadProgress.value = 100

    // Trigger browser download
    const url = URL.createObjectURL(decryptedBlob)
    const a = document.createElement('a')
    a.href = url
    a.download = shareInfo.value.resource_name
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    setTimeout(() => URL.revokeObjectURL(url), 500)
  } catch (e) {
    downloadError.value = e.message || 'Erreur lors du téléchargement ou du déchiffrement.'
  } finally {
    downloading.value = false
    downloadProgress.value = 0
  }
}

function formatSize(bytes) {
  if (!bytes) return '—'
  if (bytes < 1024) return bytes + ' o'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' Ko'
  if (bytes < 1024 * 1024 * 1024) return (bytes / 1024 / 1024).toFixed(1) + ' Mo'
  return (bytes / 1024 / 1024 / 1024).toFixed(2) + ' Go'
}

function formatDate(d) {
  if (!d) return '—'
  return new Date(d).toLocaleDateString()
}
</script>

<style scoped>
.public-share-wrapper {
  min-height: 100vh;
  background: var(--page-background, #f0f2f5);
  display: flex;
  flex-direction: column;
}

.public-header {
  padding: 16px 24px;
  background: var(--card-color);
  border-bottom: 1px solid var(--border-color);
  display: flex;
  align-items: center;
}

.brand {
  display: flex;
  align-items: center;
  gap: 8px;
}

.brand-logo { width: 28px; height: 28px; }

.brand-name {
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--main-text-color);
}

.content-container {
  flex: 1;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding: 40px 16px;
}

.loading-state,
.error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  padding: 48px 24px;
  text-align: center;
  color: var(--secondary-text-color);
}

.error-icon { margin-bottom: 8px; }

.share-card {
  width: 100%;
  max-width: 480px;
  background: var(--card-color);
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(0,0,0,0.1);
  padding: 32px 28px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.file-preview-section {
  display: flex;
  align-items: center;
  gap: 16px;
}

.file-icon-wrapper {
  flex-shrink: 0;
  width: 56px;
  height: 56px;
  background: color-mix(in srgb, var(--primary-color) 10%, transparent);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.icon-svg.large { width: 32px; height: 32px; }

.file-info-header { flex: 1; min-width: 0; }

.file-name {
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--main-text-color);
  margin: 0 0 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.share-meta {
  font-size: 0.85rem;
  color: var(--secondary-text-color);
  margin: 0;
}

.file-details-grid {
  display: flex;
  gap: 20px;
  flex-wrap: wrap;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.detail-item .label {
  font-size: 0.75rem;
  color: var(--secondary-text-color);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.detail-item .value {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--main-text-color);
}

.e2e-badge {
  display: flex;
  align-items: center;
  gap: 8px;
  background: color-mix(in srgb, var(--success-color, #22c55e) 8%, transparent);
  color: var(--success-color, #22c55e);
  border: 1px solid color-mix(in srgb, var(--success-color, #22c55e) 25%, transparent);
  border-radius: 8px;
  padding: 10px 14px;
  font-size: 0.82rem;
  line-height: 1.4;
}

.download-error {
  color: var(--error-color);
  font-size: 0.87rem;
  text-align: center;
}

.action-row {
  display: flex;
  justify-content: center;
}

.btn-primary-lg {
  width: 100%;
  padding: 14px 24px;
  border: none;
  border-radius: 10px;
  background: var(--primary-color);
  color: #fff;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  transition: opacity 0.15s;
}

.btn-primary-lg:disabled { opacity: 0.6; cursor: not-allowed; }
.btn-primary-lg:hover:not(:disabled) { opacity: 0.88; }

.spinner {
  width: 36px;
  height: 36px;
  border: 3px solid var(--border-color);
  border-top-color: var(--primary-color);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.spinner-inline {
  display: inline-block;
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255,255,255,0.4);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
  vertical-align: middle;
  margin-right: 4px;
}

@keyframes spin { to { transform: rotate(360deg); } }

.password-card {
  text-align: center;
}

.password-form {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-top: 8px;
}

.password-input {
  width: 100%;
  padding: 12px 14px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--card-color);
  color: var(--main-text-color);
  font-size: 1rem;
  box-sizing: border-box;
  transition: border-color 0.15s;
}

.password-input:focus { outline: none; border-color: var(--primary-color); }
.password-input.input-error { border-color: var(--error-color); }

.password-error-msg {
  color: var(--error-color);
  font-size: 0.85rem;
  margin: -4px 0 0;
}
</style>
