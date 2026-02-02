<template>
  <Transition name="slide-up">
    <div v-if="uploadStore.showManager && uploadStore.uploadList.length > 0" class="upload-manager">
      <!-- Header -->
      <div class="manager-header" @click="toggleCollapsed">
        <div class="header-left">
          <span class="header-icon">
            <svg v-if="hasActiveUploads" class="spinning" xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
            </svg>
            <svg v-else xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
              <polyline points="17 8 12 3 7 8"/>
              <line x1="12" y1="3" x2="12" y2="15"/>
            </svg>
          </span>
          <span class="header-title">
            {{ headerTitle }}
          </span>
        </div>
        <div class="header-actions">
          <button v-if="hasActiveUploads" @click.stop="cancelAll" class="btn-action btn-cancel" title="Annuler tout">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="3" width="18" height="18" rx="2" ry="2"/>
            </svg>
          </button>
          <button @click.stop="toggleCollapsed" class="btn-action btn-collapse" :title="isCollapsed ? 'Développer' : 'Réduire'">
            <svg :class="{ rotated: !isCollapsed }" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="6 9 12 15 18 9"/>
            </svg>
          </button>
          <button @click.stop="closeManager" class="btn-action btn-close" title="Fermer">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"/>
              <line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>
      </div>

      <!-- Progress Bar (always visible) -->
      <div class="overall-progress" v-if="hasActiveUploads">
        <div class="progress-bar">
          <div class="progress-fill" :style="{ width: uploadStore.overallProgress + '%' }"></div>
        </div>
      </div>

      <!-- Upload List -->
      <Transition name="expand">
        <div v-show="!isCollapsed" class="upload-list">
          <div 
            v-for="upload in sortedUploads" 
            :key="upload.id" 
            class="upload-item"
            :class="statusClass(upload.status)"
          >
            <!-- File Icon -->
            <div class="item-icon">
              <svg v-if="upload.status === 'completed'" class="icon-success" xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
                <polyline points="22 4 12 14.01 9 11.01"/>
              </svg>
              <svg v-else-if="upload.status === 'failed'" class="icon-error" xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"/>
                <line x1="15" y1="9" x2="9" y2="15"/>
                <line x1="9" y1="9" x2="15" y2="15"/>
              </svg>
              <svg v-else-if="upload.status === 'cancelled'" class="icon-cancelled" xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"/>
                <line x1="8" y1="12" x2="16" y2="12"/>
              </svg>
              <div v-else class="icon-uploading">
                <svg class="spinning" xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
                </svg>
              </div>
            </div>

            <!-- File Info -->
            <div class="item-info">
              <div class="item-name" :title="upload.fileName">{{ upload.fileName }}</div>
              <div class="item-details">
                <span class="item-size">{{ formatSize(upload.fileSize) }}</span>
                <span class="item-status">{{ statusLabel(upload.status) }}</span>
                <span v-if="upload.error" class="item-error" :title="upload.error">{{ truncateError(upload.error) }}</span>
              </div>
              <!-- Individual Progress Bar -->
              <div v-if="isActive(upload.status)" class="item-progress">
                <div class="progress-bar small">
                  <div class="progress-fill" :style="{ width: upload.progress + '%' }"></div>
                </div>
                <span class="progress-text">{{ upload.progress }}%</span>
              </div>
            </div>

            <!-- Actions -->
            <div class="item-actions">
              <button 
                v-if="isActive(upload.status) || upload.status === 'pending'" 
                @click="cancelUpload(upload.id)" 
                class="btn-item btn-cancel"
                title="Annuler"
              >
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <line x1="18" y1="6" x2="6" y2="18"/>
                  <line x1="6" y1="6" x2="18" y2="18"/>
                </svg>
              </button>
              <button 
                v-if="upload.status === 'failed'" 
                @click="retryUpload(upload.id)" 
                class="btn-item btn-retry"
                title="Réessayer"
              >
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="23 4 23 10 17 10"/>
                  <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
                </svg>
              </button>
              <button 
                v-if="['completed', 'failed', 'cancelled'].includes(upload.status)" 
                @click="removeUpload(upload.id)" 
                class="btn-item btn-remove"
                title="Retirer de la liste"
              >
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <line x1="18" y1="6" x2="6" y2="18"/>
                  <line x1="6" y1="6" x2="18" y2="18"/>
                </svg>
              </button>
            </div>
          </div>
        </div>
      </Transition>

      <!-- Footer Actions -->
      <div v-if="!isCollapsed && (uploadStore.counts.completed > 0 || uploadStore.counts.failed > 0)" class="manager-footer">
        <button v-if="uploadStore.counts.completed > 0" @click="clearCompleted" class="btn-footer">
          Effacer terminés ({{ uploadStore.counts.completed }})
        </button>
      </div>
    </div>
  </Transition>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useUploadStore, UploadStatus } from '../../stores/uploads'
import uploadQueueManager from '../../utils/uploadQueueManager'

const uploadStore = useUploadStore()
const isCollapsed = ref(false)

// Computed
const hasActiveUploads = computed(() => uploadStore.hasActiveUploads)

const headerTitle = computed(() => {
  const counts = uploadStore.counts
  if (counts.active > 0) {
    return `Upload en cours (${counts.active}/${counts.total})`
  }
  if (counts.failed > 0 && counts.completed === 0) {
    return `${counts.failed} échec${counts.failed > 1 ? 's' : ''}`
  }
  if (counts.completed > 0) {
    return `${counts.completed} terminé${counts.completed > 1 ? 's' : ''}`
  }
  return `${counts.total} fichier${counts.total > 1 ? 's' : ''}`
})

const sortedUploads = computed(() => {
  // Sort: active first, then pending, then completed, then failed
  const order = {
    [UploadStatus.ENCRYPTING]: 0,
    [UploadStatus.UPLOADING]: 0,
    [UploadStatus.COMPLETING]: 0,
    [UploadStatus.PENDING]: 1,
    [UploadStatus.COMPLETED]: 2,
    [UploadStatus.FAILED]: 3,
    [UploadStatus.CANCELLED]: 4
  }
  return [...uploadStore.uploadList].sort((a, b) => {
    return (order[a.status] ?? 5) - (order[b.status] ?? 5)
  })
})

// Methods
function toggleCollapsed() {
  isCollapsed.value = !isCollapsed.value
}

function closeManager() {
  uploadStore.showManager = false
}

async function cancelAll() {
  if (confirm('Annuler tous les uploads en cours ?')) {
    await uploadStore.cancelAll()
  }
}

async function cancelUpload(id) {
  await uploadStore.cancelUpload(id)
}

function retryUpload(id) {
  uploadStore.retryUpload(id)
  uploadQueueManager.startProcessing()
}

function removeUpload(id) {
  uploadStore.removeUpload(id)
}

function clearCompleted() {
  uploadStore.clearCompleted()
}

function isActive(status) {
  return [UploadStatus.ENCRYPTING, UploadStatus.UPLOADING, UploadStatus.COMPLETING].includes(status)
}

function statusClass(status) {
  return `status-${status}`
}

function statusLabel(status) {
  const labels = {
    [UploadStatus.PENDING]: 'En attente',
    [UploadStatus.ENCRYPTING]: 'Chiffrement...',
    [UploadStatus.UPLOADING]: 'Envoi...',
    [UploadStatus.COMPLETING]: 'Finalisation...',
    [UploadStatus.COMPLETED]: 'Terminé',
    [UploadStatus.FAILED]: 'Échec',
    [UploadStatus.CANCELLED]: 'Annulé'
  }
  return labels[status] || status
}

function formatSize(bytes) {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

function truncateError(error) {
  if (!error) return ''
  return error.length > 30 ? error.substring(0, 30) + '...' : error
}
</script>

<style scoped>
.upload-manager {
  position: fixed;
  bottom: 20px;
  right: 20px;
  width: 380px;
  max-height: 60vh;
  background: var(--card-color, #fff);
  border: 1px solid var(--border-color, #e0e0e0);
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15);
  z-index: 1000;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* Header */
.manager-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: var(--background-color, #f5f5f5);
  border-bottom: 1px solid var(--border-color, #e0e0e0);
  cursor: pointer;
  user-select: none;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.header-icon {
  display: flex;
  color: var(--primary-color, #3498db);
}

.header-title {
  font-weight: 600;
  font-size: 14px;
  color: var(--main-text-color, #333);
}

.header-actions {
  display: flex;
  gap: 4px;
}

.btn-action {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: none;
  background: transparent;
  border-radius: 6px;
  cursor: pointer;
  color: var(--secondary-text-color, #666);
  transition: all 0.2s;
}

.btn-action:hover {
  background: var(--hover-color, rgba(0, 0, 0, 0.05));
  color: var(--main-text-color, #333);
}

.btn-cancel:hover {
  color: var(--error-color, #e74c3c);
}

.btn-collapse svg {
  transition: transform 0.3s ease;
}

.btn-collapse svg.rotated {
  transform: rotate(180deg);
}

/* Overall Progress */
.overall-progress {
  padding: 0 16px 12px;
  background: var(--background-color, #f5f5f5);
}

.progress-bar {
  height: 4px;
  background: var(--border-color, #e0e0e0);
  border-radius: 2px;
  overflow: hidden;
}

.progress-bar.small {
  height: 3px;
}

.progress-fill {
  height: 100%;
  background: var(--primary-color, #3498db);
  border-radius: 2px;
  transition: width 0.3s ease;
}

/* Upload List */
.upload-list {
  flex: 1;
  overflow-y: auto;
  max-height: 300px;
}

.upload-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-color, #e0e0e0);
  transition: background 0.2s;
}

.upload-item:last-child {
  border-bottom: none;
}

.upload-item:hover {
  background: var(--hover-color, rgba(0, 0, 0, 0.02));
}

/* Status colors */
.status-completed {
  background: rgba(46, 204, 113, 0.05);
}

.status-failed {
  background: rgba(231, 76, 60, 0.05);
}

.status-cancelled {
  background: rgba(149, 165, 166, 0.05);
}

/* Item Icon */
.item-icon {
  flex-shrink: 0;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.icon-success {
  color: var(--success-color, #2ecc71);
}

.icon-error {
  color: var(--error-color, #e74c3c);
}

.icon-cancelled {
  color: var(--secondary-text-color, #95a5a6);
}

.icon-uploading {
  color: var(--primary-color, #3498db);
}

/* Item Info */
.item-info {
  flex: 1;
  min-width: 0;
}

.item-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--main-text-color, #333);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.item-details {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 4px;
  font-size: 11px;
  color: var(--secondary-text-color, #666);
}

.item-size {
  color: var(--secondary-text-color, #888);
}

.item-status {
  font-weight: 500;
}

.status-encrypting .item-status,
.status-uploading .item-status,
.status-completing .item-status {
  color: var(--primary-color, #3498db);
}

.status-completed .item-status {
  color: var(--success-color, #2ecc71);
}

.status-failed .item-status {
  color: var(--error-color, #e74c3c);
}

.item-error {
  color: var(--error-color, #e74c3c);
  font-style: italic;
}

.item-progress {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 6px;
}

.item-progress .progress-bar {
  flex: 1;
}

.progress-text {
  font-size: 11px;
  color: var(--secondary-text-color, #666);
  min-width: 32px;
  text-align: right;
}

/* Item Actions */
.item-actions {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
}

.btn-item {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border: none;
  background: transparent;
  border-radius: 4px;
  cursor: pointer;
  color: var(--secondary-text-color, #999);
  transition: all 0.2s;
}

.btn-item:hover {
  background: var(--hover-color, rgba(0, 0, 0, 0.05));
}

.btn-cancel:hover {
  color: var(--error-color, #e74c3c);
}

.btn-retry:hover {
  color: var(--primary-color, #3498db);
}

.btn-remove:hover {
  color: var(--secondary-text-color, #666);
}

/* Footer */
.manager-footer {
  padding: 8px 16px;
  border-top: 1px solid var(--border-color, #e0e0e0);
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.btn-footer {
  font-size: 12px;
  padding: 6px 12px;
  border: none;
  background: var(--background-color, #f5f5f5);
  color: var(--secondary-text-color, #666);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-footer:hover {
  background: var(--hover-color, #e0e0e0);
  color: var(--main-text-color, #333);
}

/* Animations */
.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* Transitions */
.slide-up-enter-active,
.slide-up-leave-active {
  transition: all 0.3s ease;
}

.slide-up-enter-from,
.slide-up-leave-to {
  opacity: 0;
  transform: translateY(20px);
}

.expand-enter-active,
.expand-leave-active {
  transition: all 0.3s ease;
  overflow: hidden;
}

.expand-enter-from,
.expand-leave-to {
  max-height: 0;
  opacity: 0;
}

.expand-enter-to,
.expand-leave-from {
  max-height: 300px;
  opacity: 1;
}

/* Scrollbar */
.upload-list::-webkit-scrollbar {
  width: 6px;
}

.upload-list::-webkit-scrollbar-track {
  background: transparent;
}

.upload-list::-webkit-scrollbar-thumb {
  background: var(--border-color, #ddd);
  border-radius: 3px;
}

.upload-list::-webkit-scrollbar-thumb:hover {
  background: var(--secondary-text-color, #bbb);
}
</style>
