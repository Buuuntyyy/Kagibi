<template>
  <Teleport to="body">
    <Transition name="slide-up">
      <div 
        v-if="downloadStore.showManager" 
        class="download-manager"
        :class="{ minimized: downloadStore.minimized }"
      >
        <!-- Header -->
        <div class="dm-header" @click="downloadStore.toggleMinimize">
          <div class="dm-header-left">
            <svg class="dm-icon" viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
              <path d="M19 9h-4V3H9v6H5l7 7 7-7zM5 18v2h14v-2H5z"/>
            </svg>
            <span class="dm-title">
              {{ downloadStore.minimized ? `${downloadStore.percent}%` : 'Téléchargement ZIP' }}
            </span>
          </div>
          <div class="dm-header-right">
            <span v-if="!downloadStore.minimized && downloadStore.isInProgress" class="dm-eta">
              {{ downloadStore.formattedEta }} restant
            </span>
            <button 
              v-if="downloadStore.canCancel" 
              class="dm-btn dm-btn-cancel"
              @click.stop="downloadStore.cancel"
              title="Annuler"
            >
              <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
                <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/>
              </svg>
            </button>
            <button 
              v-if="!downloadStore.isInProgress"
              class="dm-btn dm-btn-close"
              @click.stop="downloadStore.close"
              title="Fermer"
            >
              <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
                <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/>
              </svg>
            </button>
            <button class="dm-btn dm-btn-toggle" @click.stop="downloadStore.toggleMinimize">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
                <path v-if="downloadStore.minimized" d="M7.41 15.41L12 10.83l4.59 4.58L18 14l-6-6-6 6z"/>
                <path v-else d="M7.41 8.59L12 13.17l4.59-4.58L18 10l-6 6-6-6z"/>
              </svg>
            </button>
          </div>
        </div>
        
        <!-- Content (when not minimized) -->
        <Transition name="collapse">
          <div v-if="!downloadStore.minimized" class="dm-content">
            <!-- Status -->
            <div class="dm-status">
              <span class="dm-status-text" :class="statusClass">
                {{ downloadStore.statusText }}
              </span>
              <span class="dm-file-count">
                {{ downloadStore.processedFiles }} / {{ downloadStore.totalFiles }} fichiers
              </span>
            </div>
            
            <!-- Zip Name -->
            <div class="dm-zip-name" :title="downloadStore.zipName">
              📦 {{ downloadStore.zipName }}
            </div>
            
            <!-- Progress Bar -->
            <div class="dm-progress-wrapper">
              <div class="dm-progress-bar">
                <div 
                  class="dm-progress-fill" 
                  :class="progressClass"
                  :style="{ width: `${downloadStore.percent}%` }"
                ></div>
              </div>
              <div class="dm-progress-info">
                <span class="dm-progress-percent">{{ downloadStore.percent }}%</span>
                <span class="dm-progress-speed" v-if="downloadStore.isInProgress">
                  {{ downloadStore.formattedSpeed }}
                </span>
              </div>
            </div>
            
            <!-- Size info -->
            <div class="dm-size-info">
              {{ downloadStore.formattedDownloaded }} / {{ downloadStore.formattedTotalSize }}
            </div>
            
            <!-- Error message -->
            <div v-if="downloadStore.error" class="dm-error">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
                <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
              </svg>
              <span>{{ downloadStore.error }}</span>
            </div>
            
            <!-- File list (scrollable) -->
            <div v-if="downloadStore.files.length > 0" class="dm-file-list">
              <div 
                v-for="file in displayedFiles" 
                :key="file.name" 
                class="dm-file-item"
                :class="file.status"
              >
                <span class="dm-file-icon">
                  <svg v-if="file.status === 'completed'" viewBox="0 0 24 24" width="14" height="14" fill="currentColor">
                    <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/>
                  </svg>
                  <svg v-else-if="file.status === 'error'" viewBox="0 0 24 24" width="14" height="14" fill="currentColor">
                    <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/>
                  </svg>
                  <span v-else-if="file.status === 'downloading'" class="dm-spinner"></span>
                  <span v-else class="dm-pending-dot"></span>
                </span>
                <span class="dm-file-name" :title="file.name">{{ file.name }}</span>
              </div>
              <div v-if="downloadStore.files.length > maxDisplayedFiles" class="dm-more-files">
                +{{ downloadStore.files.length - maxDisplayedFiles }} autres fichiers
              </div>
            </div>
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useDownloadStore, DownloadStatus } from '../../stores/downloads'

const downloadStore = useDownloadStore()
const maxDisplayedFiles = 5

// Initialize download manager on mount
onMounted(async () => {
  await downloadStore.init()
})

// Computed classes
const statusClass = computed(() => ({
  'status-progress': downloadStore.isInProgress,
  'status-completed': downloadStore.status === DownloadStatus.COMPLETED,
  'status-error': downloadStore.status === DownloadStatus.ERROR,
  'status-aborted': downloadStore.status === DownloadStatus.ABORTED
}))

const progressClass = computed(() => ({
  'progress-active': downloadStore.isInProgress,
  'progress-completed': downloadStore.status === DownloadStatus.COMPLETED,
  'progress-error': downloadStore.status === DownloadStatus.ERROR
}))

const displayedFiles = computed(() => {
  return downloadStore.files.slice(0, maxDisplayedFiles)
})
</script>

<style scoped>
.download-manager {
  position: fixed;
  bottom: 20px;
  right: 20px;
  width: 360px;
  background: var(--card-color, #ffffff);
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15);
  z-index: 10000;
  overflow: hidden;
  border: 1px solid var(--border-color, #e0e0e0);
  transition: all 0.3s ease;
}

.download-manager.minimized {
  width: 220px;
}

/* Header */
.dm-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: var(--primary-color, #4a90d9);
  color: white;
  cursor: pointer;
  user-select: none;
}

.dm-header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.dm-icon {
  flex-shrink: 0;
}

.dm-title {
  font-weight: 600;
  font-size: 14px;
}

.dm-header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.dm-eta {
  font-size: 12px;
  opacity: 0.9;
}

.dm-btn {
  background: rgba(255, 255, 255, 0.2);
  border: none;
  border-radius: 4px;
  padding: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  transition: background 0.2s;
}

.dm-btn:hover {
  background: rgba(255, 255, 255, 0.3);
}

.dm-btn-cancel:hover {
  background: rgba(255, 100, 100, 0.5);
}

/* Content */
.dm-content {
  padding: 16px;
}

.dm-status {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.dm-status-text {
  font-size: 13px;
  font-weight: 500;
}

.dm-status-text.status-progress {
  color: var(--primary-color, #4a90d9);
}

.dm-status-text.status-completed {
  color: #4caf50;
}

.dm-status-text.status-error {
  color: #f44336;
}

.dm-status-text.status-aborted {
  color: #ff9800;
}

.dm-file-count {
  font-size: 12px;
  color: var(--text-secondary, #666);
}

.dm-zip-name {
  font-size: 13px;
  color: var(--text-color, #333);
  margin-bottom: 12px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Progress */
.dm-progress-wrapper {
  margin-bottom: 8px;
}

.dm-progress-bar {
  height: 8px;
  background: var(--background-color, #f5f5f5);
  border-radius: 4px;
  overflow: hidden;
}

.dm-progress-fill {
  height: 100%;
  background: var(--primary-color, #4a90d9);
  border-radius: 4px;
  transition: width 0.3s ease;
}

.dm-progress-fill.progress-active {
  background: linear-gradient(90deg, var(--primary-color, #4a90d9), #64b5f6);
}

.dm-progress-fill.progress-completed {
  background: #4caf50;
}

.dm-progress-fill.progress-error {
  background: #f44336;
}

.dm-progress-info {
  display: flex;
  justify-content: space-between;
  margin-top: 4px;
}

.dm-progress-percent {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-color, #333);
}

.dm-progress-speed {
  font-size: 12px;
  color: var(--text-secondary, #666);
}

.dm-size-info {
  font-size: 12px;
  color: var(--text-secondary, #666);
  margin-bottom: 12px;
}

/* Error */
.dm-error {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: #ffebee;
  border-radius: 6px;
  color: #c62828;
  font-size: 12px;
  margin-bottom: 12px;
}

/* File list */
.dm-file-list {
  max-height: 150px;
  overflow-y: auto;
  border-top: 1px solid var(--border-color, #e0e0e0);
  padding-top: 12px;
}

.dm-file-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 0;
  font-size: 12px;
}

.dm-file-item.completed .dm-file-icon {
  color: #4caf50;
}

.dm-file-item.error .dm-file-icon {
  color: #f44336;
}

.dm-file-item.downloading .dm-file-icon {
  color: var(--primary-color, #4a90d9);
}

.dm-file-icon {
  flex-shrink: 0;
  width: 14px;
  height: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.dm-spinner {
  width: 12px;
  height: 12px;
  border: 2px solid var(--primary-color, #4a90d9);
  border-top-color: transparent;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.dm-pending-dot {
  width: 6px;
  height: 6px;
  background: var(--text-secondary, #999);
  border-radius: 50%;
}

.dm-file-name {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  color: var(--text-color, #333);
}

.dm-more-files {
  font-size: 11px;
  color: var(--text-secondary, #666);
  text-align: center;
  padding: 8px;
  font-style: italic;
}

/* Animations */
.slide-up-enter-active,
.slide-up-leave-active {
  transition: all 0.3s ease;
}

.slide-up-enter-from,
.slide-up-leave-to {
  transform: translateY(100%);
  opacity: 0;
}

.collapse-enter-active,
.collapse-leave-active {
  transition: all 0.2s ease;
  overflow: hidden;
}

.collapse-enter-from,
.collapse-leave-to {
  max-height: 0;
  opacity: 0;
  padding-top: 0;
  padding-bottom: 0;
}

.collapse-enter-to,
.collapse-leave-from {
  max-height: 400px;
  opacity: 1;
}

/* Dark mode support */
@media (prefers-color-scheme: dark) {
  .download-manager {
    background: #1e1e1e;
    border-color: #333;
  }
  
  .dm-progress-bar {
    background: #333;
  }
  
  .dm-error {
    background: #4a1515;
    color: #ff8a80;
  }
  
  .dm-file-list {
    border-color: #333;
  }
}
</style>
