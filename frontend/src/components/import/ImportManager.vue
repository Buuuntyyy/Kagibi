<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<!-- Floating background-transfer widget for cloud imports (Google Drive / OneDrive /
     Dropbox) — structural twin of DownloadManager.vue. Lets a migration keep running
     after the import dialog (GoogleDriveImportDialog.vue etc.) closes: the dialog only
     drives the connect/folders/select wizard, then hands off to useImportStore, which
     this widget renders regardless of what page the user navigates to afterwards. -->

<template>
  <Teleport to="body">
    <Transition name="slide-up">
      <div
        v-if="importStore.showManager"
        class="import-manager"
        :class="{ minimized: importStore.minimized }"
      >
        <!-- Header -->
        <div class="im-header" @click="importStore.toggleMinimize">
          <div class="im-header-left">
            <svg class="im-icon" viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
              <polyline points="7 10 12 15 17 10"/>
              <line x1="12" y1="15" x2="12" y2="3"/>
            </svg>
            <span class="im-title">
              {{ importStore.minimized ? `${importStore.progressPercent}%` : title }}
            </span>
          </div>
          <div class="im-header-right">
            <button
              v-if="importStore.isInProgress"
              class="im-btn im-btn-cancel"
              @click.stop="importStore.abort"
              :title="t(`${importStore.namespace}.abort`)"
            >
              <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
                <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/>
              </svg>
            </button>
            <button
              v-if="!importStore.isInProgress"
              class="im-btn im-btn-close"
              @click.stop="importStore.close"
              :title="t(`${importStore.namespace}.close`)"
            >
              <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
                <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/>
              </svg>
            </button>
            <button class="im-btn im-btn-toggle" @click.stop="importStore.toggleMinimize">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
                <path v-if="importStore.minimized" d="M7.41 15.41L12 10.83l4.59 4.58L18 14l-6-6-6 6z"/>
                <path v-else d="M7.41 8.59L12 13.17l4.59-4.58L18 10l-6 6-6-6z"/>
              </svg>
            </button>
          </div>
        </div>

        <!-- Content (when not minimized) -->
        <Transition name="collapse">
          <div v-if="!importStore.minimized" class="im-content">

            <!-- Folder creation progress -->
            <div v-if="importStore.folderPhaseTotal > 0 && importStore.phase === 'progress'" class="im-folder-phase">
              <div class="im-folder-phase-label">
                {{ t(`${importStore.namespace}.creatingFolders`, { done: importStore.folderPhaseDone, total: importStore.folderPhaseTotal }) }}
              </div>
              <div class="im-bytes-bar-track">
                <div class="im-bytes-bar-fill" :style="{ width: importStore.folderPhasePercent + '%' }"></div>
              </div>
            </div>

            <!-- File count -->
            <div class="im-status">
              <span class="im-status-text">
                {{ t(`${importStore.namespace}.progressTitle`, { done: importStore.progressDone, total: importStore.progressTotal }) }}
              </span>
            </div>

            <!-- Byte-level progress -->
            <div v-if="importStore.totalImportBytes > 0" class="im-bytes-section">
              <div class="im-bytes-bar-track">
                <div class="im-bytes-bar-fill" :style="{ width: importStore.bytesPercent + '%' }"></div>
              </div>
              <div class="im-bytes-label">
                {{ formatBytes(importStore.importedBytes) }} / {{ formatBytes(importStore.totalImportBytes) }}
                <span class="im-bytes-pct">({{ importStore.bytesPercent }}%)</span>
              </div>
            </div>
            <div v-else-if="importStore.importedBytes > 0" class="im-bytes-section">
              <div class="im-bytes-label">{{ formatBytes(importStore.importedBytes) }} {{ t(`${importStore.namespace}.transferred`) }}</div>
            </div>

            <!-- Log (scrollable) -->
            <div class="im-log" ref="logEl">
              <div
                v-for="entry in importStore.log"
                :key="entry.id"
                class="im-log-row"
                :class="'im-log-' + entry.status"
              >
                <span class="im-log-icon">
                  <svg v-if="entry.status === 'done'" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M20 6L9 17l-5-5"/></svg>
                  <svg v-else-if="entry.status === 'error'" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M18 6L6 18M6 6l12 12"/></svg>
                  <span v-else class="im-spinner"></span>
                </span>
                <span class="im-log-name" :title="entry.name">{{ entry.name }}</span>
                <span v-if="entry.error" class="im-log-error" :title="entry.error">{{ entry.error }}</span>
              </div>
            </div>

            <!-- Fatal error -->
            <div v-if="importStore.error" class="im-error">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
                <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
              </svg>
              <span>{{ t(`${importStore.namespace}.fatalError`) }}: {{ importStore.error }}</span>
            </div>

            <!-- Done summary -->
            <div v-if="importStore.phase === 'done'" class="im-done-summary">
              <p v-if="importStore.errorCount === 0">
                {{ t(`${importStore.namespace}.doneSuccess`, { n: importStore.progressTotal }) }}
              </p>
              <p v-else>
                {{ t(`${importStore.namespace}.donePartial`, { ok: importStore.progressTotal - importStore.errorCount, errors: importStore.errorCount }) }}
              </p>
            </div>

          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { computed, nextTick, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useImportStore } from '../../stores/imports'
import { formatBytes } from '../../utils/importShared'

const { t } = useI18n()
const importStore = useImportStore()
const logEl = ref(null)

importStore.init()

const title = computed(() =>
  importStore.namespace ? t(`${importStore.namespace}.title`) : ''
)

watch(() => importStore.log.length, async () => {
  await nextTick()
  if (logEl.value) logEl.value.scrollTop = logEl.value.scrollHeight
})
</script>

<style scoped>
.import-manager {
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

.import-manager.minimized {
  width: 220px;
}

/* Header */
.im-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: var(--primary-color, #4a90d9);
  color: white;
  cursor: pointer;
  user-select: none;
}

.im-header-left {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.im-icon {
  flex-shrink: 0;
}

.im-title {
  font-weight: 600;
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.im-header-right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.im-btn {
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

.im-btn:hover {
  background: rgba(255, 255, 255, 0.3);
}

.im-btn-cancel:hover {
  background: rgba(255, 100, 100, 0.5);
}

/* Content */
.im-content {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.im-status-text {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-color, #333);
}

/* Folder phase + bytes bars */
.im-folder-phase {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.im-folder-phase-label {
  font-size: 11px;
  color: var(--text-secondary, #666);
}

.im-bytes-section {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.im-bytes-bar-track {
  height: 6px;
  background: var(--background-color, #f5f5f5);
  border-radius: 3px;
  overflow: hidden;
}

.im-bytes-bar-fill {
  height: 100%;
  background: var(--primary-color, #4a90d9);
  border-radius: 3px;
  transition: width 0.3s ease;
}

.im-bytes-label {
  font-size: 11px;
  color: var(--text-secondary, #666);
  text-align: right;
}

.im-bytes-pct {
  color: var(--text-secondary, #999);
}

/* Log */
.im-log {
  max-height: 150px;
  overflow-y: auto;
  border-top: 1px solid var(--border-color, #e0e0e0);
  padding-top: 8px;
}

.im-log-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 5px 0;
  font-size: 12px;
}

.im-log-icon { flex-shrink: 0; display: flex; align-items: center; }
.im-log-done .im-log-icon { color: #4caf50; }
.im-log-error .im-log-icon { color: #f44336; }
.im-log-importing .im-log-icon { color: var(--primary-color, #4a90d9); }

.im-log-name {
  flex: 1;
  min-width: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  color: var(--text-color, #333);
}

.im-log-error {
  font-size: 11px;
  color: #f44336;
  flex-shrink: 0;
  max-width: 130px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.im-spinner {
  width: 12px;
  height: 12px;
  border: 2px solid var(--border-color, #d1d5db);
  border-top-color: var(--primary-color, #4a90d9);
  border-radius: 50%;
  animation: im-spin 0.8s linear infinite;
}

@keyframes im-spin {
  to { transform: rotate(360deg); }
}

/* Error */
.im-error {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: #ffebee;
  border-radius: 6px;
  color: #c62828;
  font-size: 12px;
}

/* Done summary */
.im-done-summary {
  text-align: center;
  font-size: 13px;
  color: var(--text-color, #333);
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
  max-height: 500px;
  opacity: 1;
}

/* Dark mode support */
@media (prefers-color-scheme: dark) {
  .import-manager {
    background: #1e1e1e;
    border-color: #333;
  }

  .im-bytes-bar-track {
    background: #333;
  }

  .im-log {
    border-color: #333;
  }

  .im-error {
    background: #4a1515;
    color: #ff8a80;
  }
}
</style>
