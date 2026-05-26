<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <Teleport to="body">
    <div v-if="modelValue" class="gdi-overlay" @click.self="handleClose">
      <div class="gdi-dialog" role="dialog" aria-modal="true" :aria-labelledby="'gdi-title-' + uid">

        <!-- Header -->
        <div class="gdi-header">
          <div class="gdi-header-left">
            <svg class="gdi-gdrive-icon" viewBox="0 0 87.3 78" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
              <path d="m6.6 66.85 3.85 6.65c.8 1.4 1.95 2.5 3.3 3.3l13.75-23.8h-27.5c0 1.55.4 3.1 1.2 4.5z" fill="#0066da"/>
              <path d="m43.65 25-13.75-23.8c-1.35.8-2.5 1.9-3.3 3.3l-25.4 44a9.06 9.06 0 0 0 -1.2 4.5h27.5z" fill="#00ac47"/>
              <path d="m73.55 76.8c1.35-.8 2.5-1.9 3.3-3.3l1.6-2.75 7.65-13.25c.8-1.4 1.2-2.95 1.2-4.5h-27.502l5.852 11.5z" fill="#ea4335"/>
              <path d="m43.65 25 13.75-23.8c-1.35-.8-2.9-1.2-4.5-1.2h-18.5c-1.6 0-3.15.45-4.5 1.2z" fill="#00832d"/>
              <path d="m59.8 53h-32.3l-13.75 23.8c1.35.8 2.9 1.2 4.5 1.2h50.8c1.6 0 3.15-.45 4.5-1.2z" fill="#2684fc"/>
              <path d="m73.4 26.5-12.7-22c-.8-1.4-1.95-2.5-3.3-3.3l-13.75 23.8 16.15 27h27.45c0-1.55-.4-3.1-1.2-4.5z" fill="#ffba00"/>
            </svg>
            <h2 :id="'gdi-title-' + uid">{{ t('gdImport.title') }}</h2>
          </div>
          <button class="gdi-close" @click="handleClose" :aria-label="t('gdImport.close')">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M18 6L6 18M6 6l12 12"/>
            </svg>
          </button>
        </div>

        <!-- ── Phase 1: Connect ── -->
        <div v-if="phase === 'connect'" class="gdi-body">
          <p class="gdi-intro">{{ t('gdImport.connectIntro') }}</p>

          <div class="gdi-info-box">
            <div class="gdi-info-row">
              <svg class="gdi-info-icon gdi-icon-ok" viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 6L9 17l-5-5"/></svg>
              <span>{{ t('gdImport.infoE2E') }}</span>
            </div>
            <div class="gdi-info-row">
              <svg class="gdi-info-icon gdi-icon-ok" viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 6L9 17l-5-5"/></svg>
              <span>{{ t('gdImport.infoReadOnly') }}</span>
            </div>
            <div class="gdi-info-row">
              <svg class="gdi-info-icon gdi-icon-ok" viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 6L9 17l-5-5"/></svg>
              <span>{{ t('gdImport.infoTokenNotStored') }}</span>
            </div>
            <div class="gdi-info-row">
              <svg class="gdi-info-icon gdi-icon-info" viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
              <span>{{ t('gdImport.infoWorkspace') }}</span>
            </div>
          </div>

          <div v-if="connecting" class="gdi-connect-status">
            <span class="gdi-spinner gdi-spinner-sm"></span>
            <span>{{ connectStatusLabel }}</span>
          </div>

          <p v-if="connectError" class="gdi-error">{{ connectError }}</p>

          <div class="gdi-actions">
            <button class="gdi-btn-secondary" @click="handleClose" :disabled="connecting">{{ t('gdImport.cancel') }}</button>
            <button class="gdi-btn-primary" @click="doConnect" :disabled="connecting">
              <span v-if="connecting" class="gdi-spinner"></span>
              {{ connecting ? t('gdImport.connecting') : t('gdImport.connectBtn') }}
            </button>
          </div>
        </div>

        <!-- ── Phase 2: Select ── -->
        <div v-else-if="phase === 'select'" class="gdi-body">
          <div class="gdi-select-toolbar">
            <span class="gdi-count-label">
              {{ t('gdImport.fileCount', { n: allFiles.length }) }}
            </span>
            <div class="gdi-select-actions">
              <button class="gdi-btn-text" @click="selectAll">{{ t('gdImport.selectAll') }}</button>
              <button class="gdi-btn-text" @click="deselectAll">{{ t('gdImport.deselectAll') }}</button>
            </div>
          </div>

          <div class="gdi-file-list" role="list">
            <label
              v-for="file in allFiles"
              :key="file.id"
              class="gdi-file-row"
              role="listitem"
            >
              <input type="checkbox" :value="file.id" v-model="selectedIds" class="gdi-checkbox" />
              <span class="gdi-file-icon" :title="file.mimeType">{{ mimeIcon(file.mimeType) }}</span>
              <span class="gdi-file-name">
                {{ file.name }}
                <span v-if="isWorkspaceFile(file.mimeType)" class="gdi-badge-workspace">
                  → {{ workspaceExtension(file.mimeType) }}
                </span>
              </span>
              <span class="gdi-file-size">{{ formatBytes(parseInt(file.size || '0', 10)) }}</span>
            </label>
          </div>

          <div class="gdi-select-summary">
            {{ t('gdImport.selectedSummary', {
              n: selectedIds.length,
              size: formatBytes(selectedTotalSize)
            }) }}
          </div>

          <p v-if="selectError" class="gdi-error">{{ selectError }}</p>

          <div class="gdi-actions">
            <button class="gdi-btn-secondary" @click="handleClose">{{ t('gdImport.cancel') }}</button>
            <button
              class="gdi-btn-primary"
              @click="doImport"
              :disabled="selectedIds.length === 0"
            >
              {{ t('gdImport.importBtn', { n: selectedIds.length }) }}
            </button>
          </div>
        </div>

        <!-- ── Phase 3: Progress ── -->
        <div v-else-if="phase === 'progress'" class="gdi-body">
          <div class="gdi-progress-header">
            <span>{{ t('gdImport.progressTitle', { done: progressDone, total: progressTotal }) }}</span>
            <button
              v-if="!importDone"
              class="gdi-btn-text gdi-btn-abort"
              @click="doAbort"
            >
              {{ t('gdImport.abort') }}
            </button>
          </div>

          <div class="gdi-progress-bar-track">
            <div
              class="gdi-progress-bar-fill"
              :style="{ width: progressPercent + '%' }"
            ></div>
          </div>

          <div class="gdi-log" ref="logEl">
            <div
              v-for="entry in log"
              :key="entry.id"
              class="gdi-log-row"
              :class="'gdi-log-' + entry.status"
            >
              <span class="gdi-log-icon">
                <svg v-if="entry.status === 'done'" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M20 6L9 17l-5-5"/></svg>
                <svg v-else-if="entry.status === 'error'" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M18 6L6 18M6 6l12 12"/></svg>
                <span v-else class="gdi-spinner gdi-spinner-sm"></span>
              </span>
              <span class="gdi-log-name">{{ entry.name }}</span>
              <span v-if="entry.error" class="gdi-log-error">{{ entry.error }}</span>
            </div>
          </div>

          <div v-if="importDone" class="gdi-done-summary">
            <p v-if="errorCount === 0">{{ t('gdImport.doneSuccess', { n: progressTotal }) }}</p>
            <p v-else>{{ t('gdImport.donePartial', { ok: progressDone - errorCount, errors: errorCount }) }}</p>
          </div>

          <div class="gdi-actions">
            <button class="gdi-btn-primary" @click="handleClose" :disabled="!importDone">
              {{ importDone ? t('gdImport.close') : t('gdImport.importing') }}
            </button>
          </div>
        </div>

      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, computed, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { useFileStore } from '../../stores/files'
import {
  GoogleDriveImport,
  getImportableFiles,
  isWorkspaceFile,
  workspaceExtension,
  formatBytes
} from '../../utils/googleDriveImport'

const { t } = useI18n()

const props = defineProps({
  modelValue: { type: Boolean, default: false }
})
const emit = defineEmits(['update:modelValue'])

// Unique ID for aria-labelledby (avoids collision if dialog is mounted twice)
const uid = Math.random().toString(36).slice(2, 8)

// ── State ──
const phase       = ref('connect')      // 'connect' | 'select' | 'progress'
const connecting  = ref(false)
const connectError = ref('')
const selectError = ref('')

const allFiles    = ref([])
const pathMap     = ref(new Map())
const selectedIds = ref([])

const progressTotal = ref(0)
const progressDone  = ref(0)
const errorCount    = ref(0)
const importDone    = ref(false)
const log           = ref([])  // [{ id, name, status: 'importing'|'done'|'error', error? }]
const logEl         = ref(null)

let logCounter = 0
const importer = new GoogleDriveImport()

// ── Computed ──
const progressPercent = computed(() =>
  progressTotal.value > 0
    ? Math.round((progressDone.value / progressTotal.value) * 100)
    : 0
)

const selectedTotalSize = computed(() => {
  const idSet = new Set(selectedIds.value)
  return allFiles.value
    .filter(f => idSet.has(f.id))
    .reduce((sum, f) => sum + parseInt(f.size || '0', 10), 0)
})

// ── Reset on open ──
watch(() => props.modelValue, (open) => {
  if (open) resetDialog()
})

function resetDialog() {
  phase.value        = 'connect'
  connecting.value   = false
  connectStatus.value = 'auth'
  connectError.value = ''
  selectError.value  = ''
  allFiles.value    = []
  pathMap.value     = new Map()
  selectedIds.value = []
  progressTotal.value = 0
  progressDone.value  = 0
  errorCount.value    = 0
  importDone.value    = false
  log.value           = []
}

// ── Phase 1: Connect ──
const connectStatus = ref('auth')  // 'auth' | 'listing' | 'building'
const connectStatusLabel = computed(() => ({
  auth:     t('gdImport.statusAuth'),
  listing:  t('gdImport.statusListing'),
  building: t('gdImport.statusBuilding'),
}[connectStatus.value] ?? t('gdImport.connecting')))

async function doConnect() {
  connecting.value = true
  connectError.value = ''
  connectStatus.value = 'auth'
  try {
    await importer.init()
    if (!importer.isConfigured) {
      connectError.value = t('gdImport.errorNotConfigured')
      return
    }
    await importer.authenticate()
    connectStatus.value = 'listing'
    const { folders, files } = await importer.listAllItems()
    connectStatus.value = 'building'
    pathMap.value = await importer.buildPathMap(folders)
    allFiles.value = getImportableFiles(files)
    selectedIds.value = allFiles.value.map(f => f.id)
    phase.value = 'select'
  } catch (err) {
    connectError.value = err.message || t('gdImport.errorConnect')
  } finally {
    connecting.value = false
  }
}

// ── Phase 2: Select ──
function selectAll() { selectedIds.value = allFiles.value.map(f => f.id) }
function deselectAll() { selectedIds.value = [] }

async function doImport() {
  if (selectedIds.value.length === 0) return
  const idSet = new Set(selectedIds.value)
  const filesToImport = allFiles.value.filter(f => idSet.has(f.id))
  phase.value = 'progress'

  const fileStore = useFileStore()

  try {
    await importer.importItems(filesToImport, pathMap.value, '/', {
      onTotal(n) {
        progressTotal.value = n
      },
      onFileStart(name) {
        log.value.push({ id: logCounter++, name, status: 'importing' })
        scrollLog()
      },
      onFileDone(name) {
        progressDone.value++
        const entry = [...log.value].reverse().find(e => e.name === name && e.status === 'importing')
        if (entry) entry.status = 'done'
        scrollLog()
      },
      onFileError(name, msg) {
        progressDone.value++
        errorCount.value++
        const entry = [...log.value].reverse().find(e => e.name === name && e.status === 'importing')
        if (entry) { entry.status = 'error'; entry.error = msg }
        scrollLog()
      }
    })
  } catch (err) {
    log.value.push({ id: logCounter++, name: t('gdImport.fatalError'), status: 'error', error: err.message ?? String(err) })
    scrollLog()
  }

  importDone.value = true
  fileStore.fetchItems(fileStore.currentPath)
}

// ── Abort ──
function doAbort() {
  importer.abort()
  importDone.value = true
}

// ── Close ──
function handleClose() {
  if (!importDone.value && phase.value === 'progress') {
    importer.abort()
  }
  emit('update:modelValue', false)
}

// ── Helpers ──
async function scrollLog() {
  await nextTick()
  if (logEl.value) logEl.value.scrollTop = logEl.value.scrollHeight
}

function mimeIcon(mimeType) {
  if (mimeType === 'application/vnd.google-apps.document')      return '📄'
  if (mimeType === 'application/vnd.google-apps.spreadsheet')   return '📊'
  if (mimeType === 'application/vnd.google-apps.presentation')  return '📽'
  if (mimeType === 'application/vnd.google-apps.drawing')       return '🎨'
  if (mimeType === 'application/vnd.google-apps.form')          return '📋'
  if (mimeType.startsWith('image/'))   return '🖼'
  if (mimeType.startsWith('video/'))   return '🎬'
  if (mimeType.startsWith('audio/'))   return '🎵'
  if (mimeType === 'application/pdf')  return '📑'
  if (mimeType.includes('zip') || mimeType.includes('compressed')) return '🗜'
  return '📁'
}

</script>

<style scoped>
.gdi-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.55);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.gdi-dialog {
  background: var(--card-background, #fff);
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.25);
  width: 100%;
  max-width: 560px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* Header */
.gdi-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1.25rem 1.5rem 1rem;
  border-bottom: 1px solid var(--border-color, #e5e7eb);
}

.gdi-header-left {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.gdi-gdrive-icon {
  width: 28px;
  height: 28px;
  flex-shrink: 0;
}

.gdi-header h2 {
  font-size: 1.1rem;
  font-weight: 600;
  margin: 0;
  color: var(--text-color, #1a1a1a);
}

.gdi-close {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--secondary-text-color, #6b7280);
  padding: 0.25rem;
  border-radius: 4px;
  display: flex;
  align-items: center;
  transition: color 0.15s;
}
.gdi-close:hover { color: var(--text-color, #1a1a1a); }

/* Body */
.gdi-body {
  padding: 1.5rem;
  overflow-y: auto;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.gdi-intro {
  color: var(--secondary-text-color, #6b7280);
  line-height: 1.6;
  margin: 0;
  font-size: 0.95rem;
}

/* Connect status */
.gdi-connect-status {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  font-size: 0.875rem;
  color: var(--secondary-text-color, #6b7280);
  padding: 0.5rem 0;
}

/* Info box */
.gdi-info-box {
  background: var(--background-color, #f8fafc);
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 8px;
  padding: 1rem 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}

.gdi-info-row {
  display: flex;
  align-items: flex-start;
  gap: 0.6rem;
  font-size: 0.88rem;
  color: var(--secondary-text-color, #6b7280);
  line-height: 1.5;
}

.gdi-info-icon { flex-shrink: 0; margin-top: 1px; }
.gdi-icon-ok { color: #16a34a; }
.gdi-icon-info { color: #2563eb; }

/* Select phase */
.gdi-select-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.gdi-count-label {
  font-size: 0.9rem;
  color: var(--secondary-text-color, #6b7280);
}

.gdi-select-actions {
  display: flex;
  gap: 0.75rem;
}

.gdi-file-list {
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 8px;
  max-height: 280px;
  overflow-y: auto;
}

.gdi-file-row {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.55rem 0.9rem;
  cursor: pointer;
  transition: background-color 0.1s;
  border-bottom: 1px solid var(--border-color, #f0f0f0);
  font-size: 0.875rem;
}
.gdi-file-row:last-child { border-bottom: none; }
.gdi-file-row:hover { background: var(--background-color, #f8fafc); }

.gdi-checkbox { flex-shrink: 0; cursor: pointer; accent-color: var(--primary-color, #42b983); }
.gdi-file-icon { font-size: 1rem; flex-shrink: 0; }

.gdi-file-name {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--text-color, #1a1a1a);
}

.gdi-badge-workspace {
  font-size: 0.75rem;
  background: #dbeafe;
  color: #1d4ed8;
  border-radius: 4px;
  padding: 0 5px;
  margin-left: 0.3rem;
  white-space: nowrap;
}

.gdi-file-size {
  font-size: 0.8rem;
  color: var(--secondary-text-color, #9ca3af);
  white-space: nowrap;
  flex-shrink: 0;
}

.gdi-select-summary {
  font-size: 0.88rem;
  color: var(--secondary-text-color, #6b7280);
  text-align: right;
}

/* Progress phase */
.gdi-progress-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.9rem;
  color: var(--secondary-text-color, #6b7280);
}

.gdi-progress-bar-track {
  height: 6px;
  background: var(--border-color, #e5e7eb);
  border-radius: 3px;
  overflow: hidden;
}

.gdi-progress-bar-fill {
  height: 100%;
  background: var(--primary-color, #42b983);
  border-radius: 3px;
  transition: width 0.3s ease;
}

.gdi-log {
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 8px;
  max-height: 220px;
  overflow-y: auto;
  font-size: 0.82rem;
}

.gdi-log-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.4rem 0.75rem;
  border-bottom: 1px solid var(--border-color, #f0f0f0);
}
.gdi-log-row:last-child { border-bottom: none; }

.gdi-log-icon { flex-shrink: 0; display: flex; align-items: center; }
.gdi-log-done .gdi-log-icon { color: #16a34a; }
.gdi-log-error .gdi-log-icon { color: #dc2626; }
.gdi-log-importing .gdi-log-icon { color: var(--primary-color, #42b983); }

.gdi-log-name {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--text-color, #1a1a1a);
}

.gdi-log-error {
  font-size: 0.78rem;
  color: #dc2626;
  flex-shrink: 0;
  max-width: 160px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.gdi-done-summary {
  text-align: center;
  font-size: 0.95rem;
  color: var(--text-color, #1a1a1a);
  padding: 0.5rem 0;
}

/* Shared buttons */
.gdi-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  margin-top: auto;
  padding-top: 0.25rem;
}

.gdi-btn-primary {
  background: var(--primary-color, #42b983);
  color: #fff;
  border: none;
  border-radius: 8px;
  padding: 0.6rem 1.25rem;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  transition: opacity 0.15s;
}
.gdi-btn-primary:disabled { opacity: 0.5; cursor: default; }
.gdi-btn-primary:not(:disabled):hover { opacity: 0.88; }

.gdi-btn-secondary {
  background: transparent;
  color: var(--secondary-text-color, #6b7280);
  border: 1px solid var(--border-color, #d1d5db);
  border-radius: 8px;
  padding: 0.6rem 1.25rem;
  font-size: 0.9rem;
  cursor: pointer;
  transition: background-color 0.15s;
}
.gdi-btn-secondary:hover { background: var(--background-color, #f8fafc); }

.gdi-btn-text {
  background: none;
  border: none;
  color: var(--primary-color, #42b983);
  font-size: 0.85rem;
  cursor: pointer;
  padding: 0.2rem 0.4rem;
  border-radius: 4px;
  transition: background-color 0.15s;
}
.gdi-btn-text:hover { background: rgba(66, 185, 131, 0.08); }
.gdi-btn-abort { color: #dc2626; }
.gdi-btn-abort:hover { background: rgba(220, 38, 38, 0.08); }

/* Error */
.gdi-error {
  color: #dc2626;
  font-size: 0.875rem;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 6px;
  padding: 0.6rem 0.9rem;
  margin: 0;
}

/* Spinner */
.gdi-spinner {
  display: inline-block;
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.4);
  border-top-color: #fff;
  border-radius: 50%;
  animation: gdi-spin 0.7s linear infinite;
}
.gdi-spinner-sm {
  width: 12px;
  height: 12px;
  border: 2px solid var(--border-color, #d1d5db);
  border-top-color: var(--primary-color, #42b983);
}

@keyframes gdi-spin {
  to { transform: rotate(360deg); }
}

@media (max-width: 600px) {
  .gdi-dialog { border-radius: 8px; max-height: 95vh; }
  .gdi-body { padding: 1rem; }
  .gdi-header { padding: 1rem; }
}
</style>
