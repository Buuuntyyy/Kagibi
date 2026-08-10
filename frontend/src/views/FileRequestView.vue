<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<!-- Page publique de dépôt (file request) : le déposant chiffre côté client et
     envoie ses fichiers dans le dossier du propriétaire, sans jamais voir son contenu. -->
<template>
  <div class="request-page">
    <div class="request-card">
      <!-- Chargement -->
      <div v-if="state === 'loading'" class="center-block">
        <div class="spinner"></div>
      </div>

      <!-- Lien invalide / expiré -->
      <div v-else-if="state === 'error'" class="center-block">
        <svg viewBox="0 0 24 24" width="44" height="44" fill="currentColor" style="opacity:.35"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/></svg>
        <p>{{ errorMessage }}</p>
      </div>

      <!-- Mot de passe requis -->
      <div v-else-if="state === 'password'" class="center-block">
        <h2>{{ t('fileRequest.passwordTitle') }}</h2>
        <p class="subtle">{{ t('fileRequest.passwordHint') }}</p>
        <form class="password-row" @submit.prevent="loadShare">
          <input
            v-model="password"
            type="password"
            class="text-input"
            :placeholder="t('fileRequest.passwordPlaceholderGuest')"
            autofocus
          />
          <button type="submit" class="btn-primary">{{ t('fileRequest.unlock') }}</button>
        </form>
        <p v-if="passwordError" class="password-error">{{ t('fileRequest.wrongPassword') }}</p>
      </div>

      <!-- Dépôt -->
      <template v-else>
        <div class="request-header">
          <h2>{{ t('fileRequest.pageTitle') }}</h2>
          <p v-if="requestLabel" class="request-label">« {{ requestLabel }} »</p>
          <p class="subtle">{{ t('fileRequest.pageHint') }}</p>
          <p class="e2ee-note">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zM9 6c0-1.66 1.34-3 3-3s3 1.34 3 3v2H9V6z"/></svg>
            {{ t('fileRequest.e2eeNote') }}
          </p>
        </div>

        <div
          class="dropzone"
          :class="{ dragging: isDragging, disabled: isUploading }"
          @dragover.prevent="isDragging = true"
          @dragleave.prevent="isDragging = false"
          @drop.prevent="onDrop"
          @click="!isUploading && fileInput.click()"
        >
          <svg viewBox="0 0 24 24" width="40" height="40" fill="currentColor" style="opacity:.4"><path d="M19.35 10.04A7.49 7.49 0 0 0 12 4C9.11 4 6.6 5.64 5.35 8.04A5.994 5.994 0 0 0 0 14c0 3.31 2.69 6 6 6h13c2.76 0 5-2.24 5-5 0-2.64-2.05-4.78-4.65-4.96zM14 13v4h-4v-4H7l5-5 5 5h-3z"/></svg>
          <p>{{ t('fileRequest.dropHere') }}</p>
          <input ref="fileInput" type="file" multiple hidden @change="onPick" />
        </div>

        <!-- Progression -->
        <div v-if="isUploading" class="upload-progress">
          <div class="progress-track">
            <div class="progress-fill" :style="{ width: uploadProgress + '%' }"></div>
          </div>
          <p class="subtle">{{ t('fileRequest.uploading', { name: uploadingFileName || '…' }) }}</p>
        </div>

        <!-- Fichiers envoyés cette session -->
        <div v-if="sentFiles.length > 0" class="sent-list">
          <div v-for="(f, i) in sentFiles" :key="i" class="sent-row">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="#2e9e5b"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/></svg>
            <span class="sent-name">{{ f.name }}</span>
            <span class="sent-size">{{ formatSize(f.size) }}</span>
          </div>
        </div>
      </template>
    </div>

    <p class="powered-by">
      {{ t('fileRequest.poweredBy') }} <a href="/" target="_blank" rel="noopener">Kagibi</a>
    </p>

    <div v-if="toast" class="toast" :class="toast.type">{{ toast.message }}</div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import api from '../api'
import { deriveKeyFromToken, generateMasterKey, wrapMasterKey, encryptChunkWorker } from '../utils/crypto'
import { PART_SIZE } from '../utils/multipartUpload'
import { formatSize } from '../utils/format'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()

const token = route.params.token
const state = ref('loading') // loading | password | ready | error
const errorMessage = ref('')
const password = ref('')
const passwordError = ref(false)
const requestLabel = ref('')

const fileInput = ref(null)
const isDragging = ref(false)
const isUploading = ref(false)
const uploadProgress = ref(0)
const uploadingFileName = ref('')
const sentFiles = ref([])

const toast = ref(null)
function showToast(message, type = 'success') {
  toast.value = { message, type }
  setTimeout(() => { toast.value = null }, 3500)
}

function authHeaders() {
  return password.value ? { 'X-Share-Password': password.value } : {}
}

async function loadShare() {
  passwordError.value = false
  try {
    const { data } = await api.get(`/public/share/${token}`, { headers: authHeaders() })
    if (!data.upload_only) {
      // Lien de partage classique ouvert sur la page de dépôt → rediriger
      router.replace(`/s/${token}`)
      return
    }
    requestLabel.value = data.request_label || ''
    state.value = 'ready'
  } catch (e) {
    const status = e.response?.status
    if (status === 401) {
      if (state.value === 'password') passwordError.value = true
      state.value = 'password'
      return
    }
    errorMessage.value = status === 410
      ? t('fileRequest.linkExpired')
      : t('fileRequest.linkNotFound')
    state.value = 'error'
  }
}

onMounted(loadShare)

function onDrop(e) {
  isDragging.value = false
  if (isUploading.value) return
  uploadFiles(e.dataTransfer?.files || [])
}

function onPick(e) {
  uploadFiles(e.target.files || [])
  e.target.value = ''
}

async function uploadFiles(fileList) {
  const files = Array.from(fileList)
  if (files.length === 0) return

  isUploading.value = true
  uploadProgress.value = 0

  try {
    const tokenKey = await deriveKeyFromToken(token)

    for (let fi = 0; fi < files.length; fi++) {
      const file = files[fi]
      uploadingFileName.value = file.name

      // Chiffrement AES-GCM par chunks côté client
      const fileKey = await generateMasterKey()
      const encryptedChunks = []
      let offset = 0
      let chunkIndex = 0
      while (offset < file.size || chunkIndex === 0) {
        const chunkBlob = file.slice(offset, offset + PART_SIZE)
        const chunkBuf = await chunkBlob.arrayBuffer()
        const encChunk = await encryptChunkWorker(chunkBuf, fileKey, chunkIndex)
        encryptedChunks.push(encChunk)
        offset += PART_SIZE
        chunkIndex++
        if (file.size === 0) break
      }
      const totalEncryptedSize = encryptedChunks.reduce((s, c) => s + (c.size || c.byteLength || 0), 0)
      const encryptedFileKey = await wrapMasterKey(fileKey, tokenKey)

      const initiateRes = await api.post(`/public/share/${token}/multipart/initiate`, {
        file_name: file.name,
        file_path: '/',
        content_type: 'application/octet-stream',
        total_size: totalEncryptedSize,
        total_parts: encryptedChunks.length,
        encrypted_key: encryptedFileKey,
      }, { headers: authHeaders() })
      const { upload_id: uploadId, key: s3Key, presigned_urls: presignedURLs } = initiateRes.data

      const completedParts = []
      for (let i = 0; i < presignedURLs.length; i++) {
        const resp = await fetch(presignedURLs[i].url, { method: 'PUT', body: encryptedChunks[i] })
        if (!resp.ok) {
          await api.post(`/public/share/${token}/multipart/abort`, { upload_id: uploadId, key: s3Key }, { headers: authHeaders() }).catch(() => {})
          throw new Error(`Part ${i + 1} upload failed (HTTP ${resp.status})`)
        }
        completedParts.push({ part_number: i + 1, etag: resp.headers.get('ETag') || '' })
        uploadProgress.value = Math.round(((fi + (i + 1) / presignedURLs.length) / files.length) * 100)
      }

      await api.post(`/public/share/${token}/multipart/complete`, {
        upload_id: uploadId,
        key: s3Key,
        parts: completedParts,
        file_name: file.name,
        file_path: '/',
        total_size: totalEncryptedSize,
        content_type: 'application/octet-stream',
        encrypted_key: encryptedFileKey,
      }, { headers: authHeaders() })

      sentFiles.value.push({ name: file.name, size: file.size })
    }

    uploadProgress.value = 100
    showToast(t('fileRequest.uploadSuccess'))
  } catch (e) {
    console.error('File request upload error:', e)
    showToast(t('fileRequest.uploadError'), 'error')
  } finally {
    isUploading.value = false
    uploadingFileName.value = ''
    uploadProgress.value = 0
  }
}
</script>

<style scoped>
.request-page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 14px;
  padding: 24px 16px;
  background: var(--background-color, #f5f6f8);
  color: var(--main-text-color, #222);
}

.request-card {
  width: min(520px, 100%);
  background: var(--card-color, #fff);
  border: 1px solid var(--border-color, #e2e2e2);
  border-radius: 16px;
  padding: 28px;
}

.center-block {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 24px 0;
  text-align: center;
}

.request-header {
  text-align: center;
  margin-bottom: 18px;
}

.request-header h2 {
  margin: 0 0 6px;
  font-size: 1.25rem;
}

.request-label {
  font-size: 0.95rem;
  font-weight: 500;
  margin: 0 0 6px;
}

.subtle {
  font-size: 0.82rem;
  color: var(--subtle-text-color, #8a8a8a);
  margin: 0;
}

.e2ee-note {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 0.75rem;
  color: var(--subtle-text-color, #8a8a8a);
  margin-top: 10px;
}

.dropzone {
  border: 2px dashed var(--border-color, #ccc);
  border-radius: 12px;
  padding: 36px 16px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  text-align: center;
  transition: border-color 0.15s, background 0.15s;
}

.dropzone.dragging {
  border-color: var(--primary-color, #4f6ef7);
  background: rgba(79, 110, 247, 0.06);
}

.dropzone.disabled {
  opacity: 0.6;
  cursor: default;
}

.upload-progress {
  margin-top: 16px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.progress-track {
  height: 6px;
  border-radius: 3px;
  background: var(--border-color, #e2e2e2);
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: var(--primary-color, #4f6ef7);
  transition: width 0.2s;
}

.sent-list {
  margin-top: 16px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.sent-row {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.85rem;
}

.sent-name {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sent-size {
  color: var(--subtle-text-color, #8a8a8a);
  font-size: 0.75rem;
}

.password-row {
  display: flex;
  gap: 8px;
  width: 100%;
  max-width: 320px;
}

.text-input {
  flex: 1;
  padding: 9px 11px;
  border-radius: 8px;
  border: 1px solid var(--border-color, #ddd);
  background: var(--background-color, #fff);
  color: var(--main-text-color, #222);
}

.btn-primary {
  padding: 9px 16px;
  border-radius: 8px;
  border: none;
  background: var(--primary-color, #4f6ef7);
  color: #fff;
  cursor: pointer;
}

.password-error {
  color: #e05555;
  font-size: 0.8rem;
  margin: 0;
}

.powered-by {
  font-size: 0.78rem;
  color: var(--subtle-text-color, #8a8a8a);
}

.powered-by a {
  color: inherit;
  font-weight: 600;
  text-decoration: none;
}

.spinner {
  width: 26px;
  height: 26px;
  border: 3px solid var(--border-color, #e2e2e2);
  border-top-color: var(--primary-color, #4f6ef7);
  border-radius: 50%;
  animation: fr-spin 0.7s linear infinite;
}

@keyframes fr-spin {
  to { transform: rotate(360deg); }
}

.toast {
  position: fixed;
  bottom: 22px;
  left: 50%;
  transform: translateX(-50%);
  padding: 10px 18px;
  border-radius: 8px;
  background: #2e9e5b;
  color: #fff;
  font-size: 0.85rem;
  z-index: 2000;
}

.toast.error {
  background: #e05555;
}
</style>
