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

        <!-- ── Phase 1.5: Select Source Folders ── -->
        <div v-else-if="phase === 'folders'" class="gdi-body">
          <p class="gdi-intro">{{ t('gdImport.foldersIntro') }}</p>
          <div class="gdi-select-toolbar">
            <span class="gdi-count-label">{{ t('gdImport.rootFolderCount', { n: rootFolders.length }) }}</span>
            <div class="gdi-select-actions">
              <button class="gdi-btn-text" @click="selectAllRoots">{{ t('gdImport.selectAll') }}</button>
              <button class="gdi-btn-text" @click="deselectAllRoots">{{ t('gdImport.deselectAll') }}</button>
            </div>
          </div>
          <div class="gdi-file-list" role="list">
            <label v-for="folder in rootFolders" :key="folder.path" class="gdi-file-row" role="listitem">
              <input type="checkbox" :checked="selectedRoots.has(folder.path)" @change="toggleRoot(folder.path)" class="gdi-checkbox" />
              <span class="gdi-file-icon">📁</span>
              <span class="gdi-file-name">{{ folder.name }}</span>
            </label>
          </div>
          <!-- Dedicated root folder option -->
          <div class="gdi-dedicated-option">
            <label class="gdi-dedicated-checkbox">
              <input type="checkbox" v-model="useDedicatedFolder" class="gdi-checkbox" />
              <span>{{ t('gdImport.dedicatedFolder') }}</span>
            </label>
            <div v-if="useDedicatedFolder" class="gdi-dedicated-name-row">
              <span class="gdi-dedicated-prefix">Kagibi /</span>
              <input
                type="text"
                v-model.trim="dedicatedFolderName"
                class="gdi-dedicated-input"
                :placeholder="t('gdImport.dedicatedFolderPlaceholder')"
                maxlength="100"
                spellcheck="false"
              />
            </div>
          </div>

          <div class="gdi-actions">
            <button class="gdi-btn-secondary" @click="handleClose">{{ t('gdImport.cancel') }}</button>
            <button
              class="gdi-btn-primary"
              @click="confirmFolders"
              :disabled="selectedRoots.size === 0 || (useDedicatedFolder && !dedicatedFolderName)"
            >
              {{ t('gdImport.foldersConfirm', { n: selectedRoots.size }) }}
            </button>
          </div>
        </div>

        <!-- ── Phase 2: Select ── -->
        <div v-else-if="phase === 'select'" class="gdi-body">
          <div class="gdi-select-toolbar">
            <div class="gdi-toolbar-left">
              <button class="gdi-btn-text gdi-btn-back" @click="goBackToFolders">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M15 18l-6-6 6-6"/></svg>
                {{ t('gdImport.back') }}
              </button>
              <span class="gdi-count-label">{{ t('gdImport.fileCount', { n: allFiles.length }) }}</span>
              <span class="gdi-dest-hint" :title="t('gdImport.destinationHint')">
                <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 9l9-7 9 7v11a2 2 0 01-2 2H5a2 2 0 01-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/></svg>
                {{ targetPathDisplay }}
              </span>
            </div>
            <div class="gdi-select-actions">
              <button class="gdi-btn-text" @click="selectAll">{{ t('gdImport.selectAll') }}</button>
              <button class="gdi-btn-text" @click="deselectAll">{{ t('gdImport.deselectAll') }}</button>
            </div>
          </div>

          <div class="gdi-file-list" role="list">
            <template v-for="row in treeRows" :key="row.type === 'folder' ? 'd:' + row.path : 'f:' + row.file.id">

              <!-- Folder row -->
              <div v-if="row.type === 'folder'" class="gdi-tree-dir" :style="{ paddingLeft: (row.level * 18) + 'px' }">
                <button class="gdi-tree-toggle" @click.stop="toggleFolder(row.path)" :aria-expanded="row.isExpanded">
                  <svg class="gdi-tree-chevron" :class="{ 'gdi-tree-chevron-open': row.isExpanded }" viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2.5">
                    <path d="M9 18l6-6-6-6"/>
                  </svg>
                </button>
                <label class="gdi-tree-dir-label">
                  <input
                    type="checkbox"
                    class="gdi-checkbox"
                    :checked="folderCheckState(row.fileIds) !== 'none'"
                    :indeterminate="folderCheckState(row.fileIds) === 'some'"
                    @change="toggleFolderSelection(row.fileIds)"
                  />
                  <span class="gdi-file-icon">📁</span>
                  <span class="gdi-tree-dir-name">{{ row.name }}</span>
                  <span class="gdi-tree-dir-count">{{ row.fileIds.length }}</span>
                </label>
              </div>

              <!-- File row -->
              <label v-else class="gdi-file-row gdi-tree-file" :style="{ paddingLeft: (row.level * 18 + 22) + 'px' }">
                <input type="checkbox" :checked="selectedSet.has(row.file.id)" @change="toggleFile(row.file.id)" class="gdi-checkbox" />
                <span class="gdi-file-icon" :title="row.file.mimeType">{{ mimeIcon(row.file.mimeType) }}</span>
                <span class="gdi-file-name">
                  {{ row.file.name }}
                  <span v-if="isWorkspaceFile(row.file.mimeType)" class="gdi-badge-workspace">
                    → {{ workspaceExtension(row.file.mimeType) }}
                  </span>
                </span>
                <span class="gdi-file-size">{{ formatBytes(parseInt(row.file.size || '0', 10)) }}</span>
              </label>

            </template>
          </div>

          <div class="gdi-select-summary">
            {{ t('gdImport.selectedSummary', {
              n: selectedSet.size,
              size: formatBytes(selectedTotalSize)
            }) }}
          </div>

          <p v-if="selectError" class="gdi-error">{{ selectError }}</p>

          <div class="gdi-actions">
            <button class="gdi-btn-secondary" @click="handleClose">{{ t('gdImport.cancel') }}</button>
            <button
              class="gdi-btn-primary"
              @click="doImport"
              :disabled="selectedSet.size === 0"
            >
              {{ t('gdImport.importBtn', { n: selectedSet.size }) }}
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
            <div class="gdi-progress-bar-fill" :style="{ width: progressPercent + '%' }"></div>
          </div>

          <!-- Folder creation progress (visible while building the folder tree) -->
          <div v-if="folderPhaseTotal > 0" class="gdi-folder-phase">
            <div class="gdi-folder-phase-label">
              {{ t('gdImport.creatingFolders', { done: folderPhaseDone, total: folderPhaseTotal }) }}
            </div>
            <div class="gdi-bytes-bar-track">
              <div class="gdi-bytes-bar-fill" :style="{ width: folderPhasePercent + '%' }"></div>
            </div>
          </div>

          <!-- Byte-level progress bar (hidden when all files are Workspace with no known size) -->
          <div v-if="totalImportBytes > 0" class="gdi-bytes-section">
            <div class="gdi-bytes-bar-track">
              <div class="gdi-bytes-bar-fill" :style="{ width: bytesPercent + '%' }"></div>
            </div>
            <div class="gdi-bytes-label">
              {{ formatBytes(importedBytes) }} / {{ formatBytes(totalImportBytes) }}
              <span class="gdi-bytes-pct">({{ bytesPercent }}%)</span>
            </div>
          </div>
          <div v-else-if="importedBytes > 0" class="gdi-bytes-section">
            <div class="gdi-bytes-label">{{ formatBytes(importedBytes) }} {{ t('gdImport.transferred') }}</div>
          </div>

          <!-- Conflict resolution prompt (pauses import until user decides) -->
          <div v-if="conflictPending" class="gdi-conflict-prompt">
            <div class="gdi-conflict-header">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" flex-shrink="0"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
              <span class="gdi-conflict-title">{{ t('gdImport.conflictTitle') }}</span>
            </div>
            <p class="gdi-conflict-desc">{{ t('gdImport.conflictDesc', { name: conflictPending.folderName }) }}</p>
            <label class="gdi-conflict-apply-all">
              <input type="checkbox" v-model="conflictApplyToAll" />
              {{ t('gdImport.conflictApplyAll') }}
            </label>
            <div class="gdi-conflict-actions">
              <button class="gdi-btn-secondary gdi-btn-sm" @click="resolveConflict('skip')">{{ t('gdImport.conflictSkip') }}</button>
              <button class="gdi-btn-primary gdi-btn-sm" @click="resolveConflict('merge')">{{ t('gdImport.conflictMerge') }}</button>
            </div>
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
                <svg v-else-if="entry.status === 'skipped'" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M5 12h14"/></svg>
                <span v-else class="gdi-spinner gdi-spinner-sm"></span>
              </span>
              <span class="gdi-log-name">{{ entry.name }}</span>
              <span v-if="entry.status === 'skipped'" class="gdi-log-skipped-label">{{ t('gdImport.fileSkipped') }}</span>
              <span v-else-if="entry.error" class="gdi-log-error">{{ entry.error }}</span>
            </div>
          </div>

          <div v-if="importDone" class="gdi-done-summary">
            <p v-if="errorCount === 0 && skipCount === 0">{{ t('gdImport.doneSuccess', { n: progressTotal }) }}</p>
            <p v-else-if="errorCount === 0">{{ t('gdImport.doneWithSkips', { ok: progressTotal - skipCount, skips: skipCount }) }}</p>
            <p v-else-if="skipCount === 0">{{ t('gdImport.donePartial', { ok: progressTotal - errorCount, errors: errorCount }) }}</p>
            <p v-else>{{ t('gdImport.donePartialWithSkips', { ok: progressTotal - errorCount - skipCount, skips: skipCount, errors: errorCount }) }}</p>
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
import api from '../../api'
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
const selectedSet = ref(new Set())  // O(1) lookup, avoids Array.includes on every checkbox render
const rootFolders = ref([])         // [{ path: '/Name', name: 'Name' }] — depth-1 Drive folders
const selectedRoots = ref(new Set())

const progressTotal      = ref(0)
const progressDone       = ref(0)
const errorCount         = ref(0)
const skipCount          = ref(0)
const importedBytes      = ref(0)   // bytes uploaded so far (encrypted, but ~= plaintext)
const totalImportBytes   = ref(0)   // sum of plaintext sizes of selected regular files
const importDone         = ref(false)
const log           = ref([])  // [{ id, name, status: 'importing'|'done'|'error'|'skipped', error? }]
const logEl         = ref(null)

const conflictPending    = ref(null)   // { folderName, folderPath, resolve } | null
const conflictApplyToAll = ref(false)
const autoConflictChoice = ref(null)   // 'merge' | 'skip' | null (set when "apply to all")

const expandedFolders    = ref(new Set())   // set of kagibi folder paths currently expanded
const allFilesUnfiltered = ref([])          // full file list before confirmFolders filter

const useDedicatedFolder  = ref(true)              // import into a dedicated root folder
const dedicatedFolderName = ref('Google Drive')    // name of that folder

const folderPhaseDone  = ref(0)   // folders created so far in the tree-creation phase
const folderPhaseTotal = ref(0)   // total folders to create

let logCounter = 0
const importer = new GoogleDriveImport()

// ── Computed ──
const progressPercent = computed(() =>
  progressTotal.value > 0
    ? Math.round((progressDone.value / progressTotal.value) * 100)
    : 0
)

const bytesPercent = computed(() =>
  totalImportBytes.value > 0
    ? Math.min(Math.round((importedBytes.value / totalImportBytes.value) * 100), 100)
    : 0
)

const folderPhasePercent = computed(() =>
  folderPhaseTotal.value > 0
    ? Math.min(Math.round((folderPhaseDone.value / folderPhaseTotal.value) * 100), 100)
    : 0
)

const targetPathDisplay = computed(() =>
  useDedicatedFolder.value && dedicatedFolderName.value.trim()
    ? '/' + dedicatedFolderName.value.trim()
    : '/'
)

// Flat tree representation: [{type:'folder', path, name, level, isExpanded, fileIds} | {type:'file', file, level}]
// Recomputed when allFiles or expandedFolders changes.
const treeRows = computed(() => {
  // Group files by their kagibi directory path
  const dirFiles = new Map()
  for (const file of allFiles.value) {
    const parentId = (file.parents ?? [])[0]
    const fp = parentId ? pathMap.value.get(parentId) : null
    const dir = fp || '__root__'
    if (!dirFiles.has(dir)) dirFiles.set(dir, [])
    dirFiles.get(dir).push(file)
  }

  // Collect all ancestor folder paths implied by the file paths
  const allDirPaths = new Set()
  for (const dir of dirFiles.keys()) {
    if (dir === '__root__') continue
    const segs = dir.split('/').filter(Boolean)
    for (let i = 1; i <= segs.length; i++) allDirPaths.add('/' + segs.slice(0, i).join('/'))
  }

  // Build parent → [children] map
  const childFolders = new Map()
  for (const dir of allDirPaths) {
    const segs = dir.split('/').filter(Boolean)
    const parent = segs.length === 1 ? '__root__' : '/' + segs.slice(0, -1).join('/')
    if (!childFolders.has(parent)) childFolders.set(parent, [])
    childFolders.get(parent).push(dir)
  }

  // Recursively collect all file IDs in a directory subtree
  function getFileIds(dirPath) {
    const ids = (dirFiles.get(dirPath) || []).map(f => f.id)
    for (const sub of (childFolders.get(dirPath) || [])) ids.push(...getFileIds(sub))
    return ids
  }

  const rows = []
  function visit(dirPath, level) {
    if (dirPath !== '__root__') {
      const segs = dirPath.split('/').filter(Boolean)
      const name = segs[segs.length - 1]
      const isExpanded = expandedFolders.value.has(dirPath)
      rows.push({ type: 'folder', path: dirPath, name, level, isExpanded, fileIds: getFileIds(dirPath) })
      if (!isExpanded) return
    }
    const childLevel = dirPath === '__root__' ? 0 : level + 1
    const subs = (childFolders.get(dirPath) || []).sort((a, b) => a.localeCompare(b))
    for (const sub of subs) visit(sub, childLevel)
    const filesHere = (dirFiles.get(dirPath) || []).sort((a, b) => a.name.localeCompare(b.name))
    for (const file of filesHere) rows.push({ type: 'file', file, level: childLevel })
  }

  visit('__root__', -1)
  return rows
})

const selectedTotalSize = computed(() => {
  let sum = 0
  for (const f of allFiles.value) {
    if (selectedSet.value.has(f.id)) sum += parseInt(f.size || '0', 10)
  }
  return sum
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
  allFiles.value      = []
  pathMap.value       = new Map()
  selectedSet.value   = new Set()
  rootFolders.value   = []
  selectedRoots.value = new Set()
  progressTotal.value = 0
  progressDone.value  = 0
  errorCount.value         = 0
  skipCount.value          = 0
  importedBytes.value      = 0
  totalImportBytes.value   = 0
  importDone.value         = false
  log.value           = []
  conflictPending.value    = null
  conflictApplyToAll.value = false
  autoConflictChoice.value = null
  expandedFolders.value    = new Set()
  allFilesUnfiltered.value = []
  useDedicatedFolder.value  = true
  dedicatedFolderName.value = 'Google Drive'
  folderPhaseDone.value  = 0
  folderPhaseTotal.value = 0
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
    allFilesUnfiltered.value = allFiles.value   // keep original for back-navigation

    // Extract depth-1 folder paths from pathMap → let the user pick which roots to import.
    const rootPaths = new Map()
    for (const [, folderPath] of pathMap.value) {
      const segments = folderPath.split('/').filter(Boolean)
      if (segments.length >= 1) {
        const rootPath = '/' + segments[0]
        if (!rootPaths.has(rootPath)) rootPaths.set(rootPath, segments[0])
      }
    }
    rootFolders.value = [...rootPaths.entries()]
      .map(([path, name]) => ({ path, name }))
      .sort((a, b) => a.name.localeCompare(b.name))

    // Detect files sitting directly in My Drive root (parent not in pathMap).
    // Expose them as an explicit selectable entry so users can include or exclude them.
    const hasRootFiles = allFiles.value.some(f => {
      const parentId = (f.parents ?? [])[0]
      return !parentId || !pathMap.value.has(parentId)
    })
    if (hasRootFiles) {
      rootFolders.value.unshift({ path: '__root__', name: t('gdImport.driveRoot') })
    }

    if (rootFolders.value.length > 0) {
      selectedRoots.value = new Set(rootFolders.value.map(r => r.path))
      phase.value = 'folders'
    } else {
      // No sub-folders at all — all files are at Drive root, skip folder selection.
      selectedSet.value = new Set(allFiles.value.map(f => f.id))
      phase.value = 'select'
    }
  } catch (err) {
    connectError.value = err.message || t('gdImport.errorConnect')
  } finally {
    connecting.value = false
  }
}

// ── Phase 1.5: Folder selection ──
function selectAllRoots()  { selectedRoots.value = new Set(rootFolders.value.map(r => r.path)) }
function deselectAllRoots() { selectedRoots.value = new Set() }
function toggleRoot(path) {
  const s = new Set(selectedRoots.value)
  s.has(path) ? s.delete(path) : s.add(path)
  selectedRoots.value = s
}
function confirmFolders() {
  if (selectedRoots.value.size === 0) return
  const includeRoot = selectedRoots.value.has('__root__')
  // Always filter from the original full list so back+confirm works correctly
  const filtered = allFilesUnfiltered.value.filter(f => {
    const parentId = (f.parents ?? [])[0]
    const folderPath = parentId ? pathMap.value.get(parentId) : null
    if (!folderPath) return includeRoot
    const rootPath = '/' + folderPath.split('/').filter(Boolean)[0]
    return selectedRoots.value.has(rootPath)
  })
  allFiles.value = filtered
  selectedSet.value = new Set(filtered.map(f => f.id))

  // Expand all folders that contain at least one file
  const expanded = new Set()
  for (const file of filtered) {
    const parentId = (file.parents ?? [])[0]
    const fp = parentId ? pathMap.value.get(parentId) : null
    if (fp) {
      const segs = fp.split('/').filter(Boolean)
      for (let i = 1; i <= segs.length; i++) expanded.add('/' + segs.slice(0, i).join('/'))
    }
  }
  expandedFolders.value = expanded

  phase.value = 'select'
}

// ── Go back from select to folder picker ──
function goBackToFolders() {
  phase.value = 'folders'
  // allFiles will be re-filtered from allFilesUnfiltered on next confirmFolders()
}

// ── Tree view helpers ──
function toggleFolder(path) {
  const s = new Set(expandedFolders.value)
  s.has(path) ? s.delete(path) : s.add(path)
  expandedFolders.value = s
}

// 'all' | 'some' | 'none'
function folderCheckState(fileIds) {
  let n = 0
  for (const id of fileIds) if (selectedSet.value.has(id)) n++
  if (n === 0) return 'none'
  if (n === fileIds.length) return 'all'
  return 'some'
}

function toggleFolderSelection(fileIds) {
  const s = new Set(selectedSet.value)
  const allSelected = fileIds.every(id => s.has(id))
  if (allSelected) fileIds.forEach(id => s.delete(id))
  else             fileIds.forEach(id => s.add(id))
  selectedSet.value = s
}

// ── Phase 2: Select ──
function selectAll()  { selectedSet.value = new Set(allFiles.value.map(f => f.id)) }
function deselectAll() { selectedSet.value = new Set() }
function toggleFile(id) {
  const s = new Set(selectedSet.value)
  s.has(id) ? s.delete(id) : s.add(id)
  selectedSet.value = s
}

async function doImport() {
  if (selectedSet.value.size === 0) return

  // Deduplicate by file ID — the same Drive file can appear in allFiles more than once
  // if the user went back and re-confirmed, or if the Drive API returned duplicates.
  const seenIds = new Set()
  const filesToImport = []
  for (const f of allFiles.value) {
    if (selectedSet.value.has(f.id) && !seenIds.has(f.id)) {
      seenIds.add(f.id)
      filesToImport.push(f)
    }
  }

  phase.value = 'progress'

  const fileStore = useFileStore()
  fileStore.startHeartbeat()

  // Create dedicated root folder if requested, then derive the target path
  let targetPath = '/'
  if (useDedicatedFolder.value && dedicatedFolderName.value) {
    try {
      await api.post('/folders/create', { name: dedicatedFolderName.value, path: '/' })
    } catch (err) {
      if (err?.response?.status !== 409) {
        // 409 = already exists, perfectly fine; anything else is a real error
        const msg = err?.response?.data?.error ?? err.message ?? String(err)
        log.value.push({ id: logCounter++, name: dedicatedFolderName.value, status: 'error', error: msg })
        importDone.value = true
        fileStore.stopHeartbeat()
        return
      }
    }
    targetPath = '/' + dedicatedFolderName.value
  }

  try {
    await importer.importItems(filesToImport, pathMap.value, targetPath, {
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
      },
      onFileSkipped(name) {
        progressDone.value++
        skipCount.value++
        const entry = [...log.value].reverse().find(e => e.name === name && e.status === 'importing')
        if (entry) entry.status = 'skipped'
        scrollLog()
      },
      onFolderConflict(folderName, folderPath) {
        if (autoConflictChoice.value) return Promise.resolve(autoConflictChoice.value)
        return new Promise(resolve => {
          conflictPending.value = { folderName, folderPath, resolve }
        })
      },
      onBytesProgress(uploaded, total) {
        importedBytes.value = uploaded
        totalImportBytes.value = total
      },
      onFolderProgress(done, total) {
        folderPhaseDone.value = done
        folderPhaseTotal.value = total
      }
    })
  } catch (err) {
    log.value.push({ id: logCounter++, name: t('gdImport.fatalError'), status: 'error', error: err.message ?? String(err) })
    scrollLog()
  } finally {
    fileStore.stopHeartbeat()
  }

  importDone.value = true
  fileStore.fetchItems(fileStore.currentPath)
}

// ── Conflict resolution ──
function resolveConflict(choice) {
  if (conflictApplyToAll.value) autoConflictChoice.value = choice
  if (conflictPending.value) {
    conflictPending.value.resolve(choice)
    conflictPending.value = null
  }
  conflictApplyToAll.value = false
}

// ── Abort ──
function doAbort() {
  if (conflictPending.value) {
    conflictPending.value.resolve('skip')
    conflictPending.value = null
  }
  importer.abort()
  importDone.value = true
}

// ── Close ──
function handleClose() {
  if (!importDone.value && phase.value === 'progress') {
    if (conflictPending.value) {
      conflictPending.value.resolve('skip')
      conflictPending.value = null
    }
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

.gdi-load-more {
  display: flex;
  justify-content: center;
  padding: 0.5rem;
  border-top: 1px solid var(--border-color, #f0f0f0);
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

.gdi-folder-phase {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
}

.gdi-folder-phase-label {
  font-size: 0.78rem;
  color: var(--secondary-text-color, #6b7280);
}

.gdi-bytes-section {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
}

.gdi-bytes-bar-track {
  height: 4px;
  background: var(--border-color, #e5e7eb);
  border-radius: 2px;
  overflow: hidden;
}

.gdi-bytes-bar-fill {
  height: 100%;
  background: #60a5fa;
  border-radius: 2px;
  transition: width 0.4s ease;
}

.gdi-bytes-label {
  font-size: 0.78rem;
  color: var(--secondary-text-color, #6b7280);
  text-align: right;
}

.gdi-bytes-pct {
  color: var(--secondary-text-color, #9ca3af);
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
.gdi-log-skipped .gdi-log-icon { color: #9ca3af; }
.gdi-log-skipped .gdi-log-name { color: #9ca3af; }

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

.gdi-log-skipped-label {
  font-size: 0.75rem;
  color: #9ca3af;
  flex-shrink: 0;
}

/* Conflict resolution prompt */
.gdi-conflict-prompt {
  background: #fffbeb;
  border: 1px solid #fbbf24;
  border-radius: 8px;
  padding: 0.9rem 1.1rem;
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}

.gdi-conflict-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #92400e;
}

.gdi-conflict-title {
  font-weight: 600;
  font-size: 0.875rem;
}

.gdi-conflict-desc {
  font-size: 0.85rem;
  color: #78350f;
  margin: 0;
}

.gdi-conflict-apply-all {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.8rem;
  color: #92400e;
  cursor: pointer;
}

.gdi-conflict-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}

.gdi-btn-sm {
  padding: 0.4rem 0.9rem;
  font-size: 0.82rem;
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

/* Dedicated folder option */
.gdi-dedicated-option {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding: 0.75rem 1rem;
  background: var(--background-color, #f8fafc);
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 8px;
  font-size: 0.88rem;
}

.gdi-dedicated-checkbox {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  color: var(--text-color, #1a1a1a);
  font-weight: 500;
}

.gdi-dedicated-name-row {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  margin-left: 1.5rem;
}

.gdi-dedicated-prefix {
  font-size: 0.85rem;
  color: var(--secondary-text-color, #6b7280);
  white-space: nowrap;
}

.gdi-dedicated-input {
  flex: 1;
  border: 1px solid var(--border-color, #d1d5db);
  border-radius: 6px;
  padding: 0.3rem 0.6rem;
  font-size: 0.85rem;
  background: var(--card-background, #fff);
  color: var(--text-color, #1a1a1a);
  outline: none;
  min-width: 0;
}
.gdi-dedicated-input:focus { border-color: var(--primary-color, #42b983); }

/* Destination hint in select toolbar */
.gdi-dest-hint {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.78rem;
  color: var(--secondary-text-color, #9ca3af);
  font-family: monospace;
  white-space: nowrap;
}

/* Toolbar back button */
.gdi-toolbar-left {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.gdi-btn-back {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  color: var(--secondary-text-color, #6b7280);
  font-size: 0.85rem;
}
.gdi-btn-back:hover { background: rgba(107, 114, 128, 0.08); }

/* Tree view */
.gdi-tree-dir {
  display: flex;
  align-items: center;
  border-bottom: 1px solid var(--border-color, #f0f0f0);
  background: var(--background-color, #f8fafc);
  font-size: 0.875rem;
  min-height: 36px;
}

.gdi-tree-toggle {
  flex-shrink: 0;
  width: 22px;
  height: 22px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: none;
  border: none;
  cursor: pointer;
  color: var(--secondary-text-color, #6b7280);
  border-radius: 3px;
  margin-left: 6px;
  transition: background 0.1s;
}
.gdi-tree-toggle:hover { background: var(--border-color, #e5e7eb); }

.gdi-tree-chevron {
  transition: transform 0.15s;
  flex-shrink: 0;
}
.gdi-tree-chevron-open { transform: rotate(90deg); }

.gdi-tree-dir-label {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  flex: 1;
  cursor: pointer;
  padding: 0.35rem 0.75rem 0.35rem 0.25rem;
  min-width: 0;
}

.gdi-tree-dir-name {
  flex: 1;
  font-weight: 500;
  color: var(--text-color, #1a1a1a);
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.gdi-tree-dir-count {
  font-size: 0.72rem;
  color: var(--secondary-text-color, #9ca3af);
  background: var(--border-color, #e5e7eb);
  border-radius: 10px;
  padding: 1px 6px;
  flex-shrink: 0;
}

.gdi-tree-file {
  /* gdi-file-row covers the rest; extra indent is applied inline */
}

@media (max-width: 600px) {
  .gdi-dialog { border-radius: 8px; max-height: 95vh; }
  .gdi-body { padding: 1rem; }
  .gdi-header { padding: 1rem; }
}
</style>
