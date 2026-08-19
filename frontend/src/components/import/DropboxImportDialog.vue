<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <Teleport to="body">
    <div v-if="modelValue" class="dbi-overlay" @click.self="handleClose">
      <div class="dbi-dialog" role="dialog" aria-modal="true" :aria-labelledby="'dbi-title-' + uid">

        <!-- Header -->
        <div class="dbi-header">
          <div class="dbi-header-left">
            <svg class="dbi-dropbox-icon" viewBox="0 0 43 40" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
              <path d="M12.5 0L0 8.1l8.7 6.9L21.2 7 12.5 0z" fill="#0061FF"/>
              <path d="M0 21.9l12.5 8.1 8.7-6.9-12.5-8.1L0 21.9z" fill="#0061FF"/>
              <path d="M21.2 23.1l8.7 6.9L42.4 21.9l-8.7-6.9-12.5 8.1z" fill="#0061FF"/>
              <path d="M42.4 8.1L29.9 0l-8.7 6.9 12.5 8.1 8.7-6.9z" fill="#0061FF"/>
              <path d="M21.3 24.6l-8.7 6.9-3.7-2.4v2.7L21.3 40l12.4-8.2v-2.7l-3.7 2.4-8.4-6.9z" fill="#0061FF"/>
            </svg>
            <h2 :id="'dbi-title-' + uid">{{ t('dbImport.title') }}</h2>
          </div>
          <button class="dbi-close" @click="handleClose" :aria-label="t('dbImport.close')">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M18 6L6 18M6 6l12 12"/>
            </svg>
          </button>
        </div>

        <!-- ── Phase 1: Connect ── -->
        <div v-if="phase === 'connect'" class="dbi-body">
          <p class="dbi-intro">{{ t('dbImport.connectIntro') }}</p>

          <div class="dbi-info-box">
            <div class="dbi-info-row">
              <svg class="dbi-info-icon dbi-icon-ok" viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 6L9 17l-5-5"/></svg>
              <span>{{ t('dbImport.infoE2E') }}</span>
            </div>
            <div class="dbi-info-row">
              <svg class="dbi-info-icon dbi-icon-ok" viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 6L9 17l-5-5"/></svg>
              <span>{{ t('dbImport.infoReadOnly') }}</span>
            </div>
            <div class="dbi-info-row">
              <svg class="dbi-info-icon dbi-icon-ok" viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 6L9 17l-5-5"/></svg>
              <span>{{ t('dbImport.infoTokenNotStored') }}</span>
            </div>
          </div>

          <div v-if="connecting" class="dbi-connect-status">
            <span class="dbi-spinner dbi-spinner-sm"></span>
            <span>{{ connectStatusLabel }}</span>
          </div>

          <p v-if="connectError" class="dbi-error">{{ connectError }}</p>

          <div class="dbi-actions">
            <button class="dbi-btn-secondary" @click="handleClose" :disabled="connecting">{{ t('dbImport.cancel') }}</button>
            <button class="dbi-btn-primary" @click="doConnect" :disabled="connecting">
              <span v-if="connecting" class="dbi-spinner"></span>
              {{ connecting ? t('dbImport.connecting') : t('dbImport.connectBtn') }}
            </button>
          </div>
        </div>

        <!-- ── Phase 1.5: Select Source Folders ── -->
        <div v-else-if="phase === 'folders'" class="dbi-body">
          <p class="dbi-intro">{{ t('dbImport.foldersIntro') }}</p>
          <div class="dbi-select-toolbar">
            <span class="dbi-count-label">{{ t('dbImport.rootFolderCount', { n: rootFolders.length }) }}</span>
            <div class="dbi-select-actions">
              <button class="dbi-btn-text" @click="selectAllRoots">{{ t('dbImport.selectAll') }}</button>
              <button class="dbi-btn-text" @click="deselectAllRoots">{{ t('dbImport.deselectAll') }}</button>
            </div>
          </div>
          <div class="dbi-file-list" role="list">
            <label v-for="folder in rootFolders" :key="folder.path" class="dbi-file-row" role="listitem">
              <input type="checkbox" :checked="selectedRoots.has(folder.path)" @change="toggleRoot(folder.path)" class="dbi-checkbox" />
              <span class="dbi-file-icon">📁</span>
              <span class="dbi-file-name">{{ folder.name }}</span>
            </label>
          </div>
          <!-- Dedicated root folder option -->
          <div class="dbi-dedicated-option">
            <label class="dbi-dedicated-checkbox">
              <input type="checkbox" v-model="useDedicatedFolder" class="dbi-checkbox" />
              <span>{{ t('dbImport.dedicatedFolder') }}</span>
            </label>
            <div v-if="useDedicatedFolder" class="dbi-dedicated-name-row">
              <span class="dbi-dedicated-prefix">Kagibi /</span>
              <input
                type="text"
                v-model.trim="dedicatedFolderName"
                class="dbi-dedicated-input"
                :placeholder="t('dbImport.dedicatedFolderPlaceholder')"
                maxlength="100"
                spellcheck="false"
              />
            </div>
          </div>

          <div class="dbi-actions">
            <button class="dbi-btn-secondary" @click="handleClose">{{ t('dbImport.cancel') }}</button>
            <button
              class="dbi-btn-primary"
              @click="confirmFolders"
              :disabled="selectedRoots.size === 0 || (useDedicatedFolder && !dedicatedFolderName)"
            >
              {{ t('dbImport.foldersConfirm', { n: selectedRoots.size }) }}
            </button>
          </div>
        </div>

        <!-- ── Phase 2: Select ── -->
        <div v-else-if="phase === 'select'" class="dbi-body">
          <div class="dbi-select-toolbar">
            <div class="dbi-toolbar-left">
              <button class="dbi-btn-text dbi-btn-back" @click="goBackToFolders">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M15 18l-6-6 6-6"/></svg>
                {{ t('dbImport.back') }}
              </button>
              <span class="dbi-count-label">{{ t('dbImport.fileCount', { n: allFiles.length }) }}</span>
              <span class="dbi-dest-hint" :title="t('dbImport.destinationHint')">
                <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 9l9-7 9 7v11a2 2 0 01-2 2H5a2 2 0 01-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/></svg>
                {{ targetPathDisplay }}
              </span>
            </div>
            <div class="dbi-select-actions">
              <button class="dbi-btn-text" @click="selectAll">{{ t('dbImport.selectAll') }}</button>
              <button class="dbi-btn-text" @click="deselectAll">{{ t('dbImport.deselectAll') }}</button>
            </div>
          </div>

          <div class="dbi-file-list" role="list">
            <template v-for="row in treeRows" :key="row.type === 'folder' ? 'd:' + row.path : 'f:' + row.file.id">

              <!-- Folder row -->
              <div v-if="row.type === 'folder'" class="dbi-tree-dir" :style="{ paddingLeft: (row.level * 18) + 'px' }">
                <button class="dbi-tree-toggle" @click.stop="toggleFolder(row.path)" :aria-expanded="row.isExpanded">
                  <svg class="dbi-tree-chevron" :class="{ 'dbi-tree-chevron-open': row.isExpanded }" viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2.5">
                    <path d="M9 18l6-6-6-6"/>
                  </svg>
                </button>
                <label class="dbi-tree-dir-label">
                  <input
                    type="checkbox"
                    class="dbi-checkbox"
                    :checked="folderCheckState(row.fileIds) !== 'none'"
                    :indeterminate="folderCheckState(row.fileIds) === 'some'"
                    @change="toggleFolderSelection(row.fileIds)"
                  />
                  <span class="dbi-file-icon">📁</span>
                  <span class="dbi-tree-dir-name">{{ row.name }}</span>
                  <span class="dbi-tree-dir-count">{{ row.fileIds.length }}</span>
                </label>
              </div>

              <!-- File row -->
              <label v-else class="dbi-file-row dbi-tree-file" :style="{ paddingLeft: (row.level * 18 + 22) + 'px' }">
                <input type="checkbox" :checked="selectedSet.has(row.file.id)" @change="toggleFile(row.file.id)" class="dbi-checkbox" />
                <span class="dbi-file-icon">{{ mimeIcon(row.file.name) }}</span>
                <span class="dbi-file-name">{{ row.file.name }}</span>
                <span class="dbi-file-size">{{ formatBytes(parseInt(row.file.size || '0', 10)) }}</span>
              </label>

            </template>
          </div>

          <div class="dbi-select-summary">
            {{ t('dbImport.selectedSummary', {
              n: selectedSet.size,
              size: formatBytes(selectedTotalSize)
            }) }}
          </div>

          <p v-if="selectError" class="dbi-error">{{ selectError }}</p>

          <div class="dbi-actions">
            <button class="dbi-btn-secondary" @click="handleClose">{{ t('dbImport.cancel') }}</button>
            <button
              class="dbi-btn-primary"
              @click="doImport"
              :disabled="selectedSet.size === 0"
            >
              {{ t('dbImport.importBtn', { n: selectedSet.size }) }}
            </button>
          </div>
        </div>

      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useImportStore } from '../../stores/imports'
import { useFileStore } from '../../stores/files'
import {
  DropboxImport,
  getImportableFiles,
  formatBytes
} from '../../utils/dropboxImport'

const { t } = useI18n()
const importStore = useImportStore()
const fileStore = useFileStore()

const props = defineProps({
  modelValue: { type: Boolean, default: false }
})
const emit = defineEmits(['update:modelValue'])

// Unique ID for aria-labelledby (avoids collision if dialog is mounted twice)
const uid = Math.random().toString(36).slice(2, 8)

// ── State ──
// Transfer progress lives in useImportStore/ImportManager.vue once doImport()
// hands off — this dialog only ever drives the wizard up to that point.
const phase       = ref('connect')      // 'connect' | 'folders' | 'select'
const connecting  = ref(false)
const connectError = ref('')
const selectError = ref('')

const allFiles    = ref([])
const pathMap     = ref(new Map())
const selectedSet = ref(new Set())  // O(1) lookup, avoids Array.includes on every checkbox render
const rootFolders = ref([])         // [{ path: '/Name', name: 'Name' }] — depth-1 Dropbox folders
const selectedRoots = ref(new Set())

const expandedFolders    = ref(new Set())   // set of kagibi folder paths currently expanded
const allFilesUnfiltered = ref([])          // full file list before confirmFolders filter

const useDedicatedFolder  = ref(true)         // import into a dedicated root folder
const dedicatedFolderName = ref('Dropbox')    // name of that folder

const importer = new DropboxImport()

// ── Computed ──
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
    const fp = file.parentId ? pathMap.value.get(file.parentId) : null
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
  expandedFolders.value    = new Set()
  allFilesUnfiltered.value = []
  useDedicatedFolder.value  = true
  dedicatedFolderName.value = 'Dropbox'
}

// ── Phase 1: Connect ──
const connectStatus = ref('auth')  // 'auth' | 'listing' | 'building'
const connectStatusLabel = computed(() => ({
  auth:     t('dbImport.statusAuth'),
  listing:  t('dbImport.statusListing'),
  building: t('dbImport.statusBuilding'),
}[connectStatus.value] ?? t('dbImport.connecting')))

async function doConnect() {
  connecting.value = true
  connectError.value = ''
  connectStatus.value = 'auth'
  // Listing a large drive can take well over the session's inactivity window — keep the
  // session (server + client-side timers) alive for the whole connect phase, not just
  // the transfer that useImportStore covers afterwards.
  fileStore.startHeartbeat()
  try {
    await importer.init()
    if (!importer.isConfigured) {
      connectError.value = t('dbImport.errorNotConfigured')
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

    // Detect files sitting directly at the Dropbox root (parent not in pathMap).
    // Expose them as an explicit selectable entry so users can include or exclude them.
    const hasRootFiles = allFiles.value.some(f => !f.parentId || !pathMap.value.has(f.parentId))
    if (hasRootFiles) {
      rootFolders.value.unshift({ path: '__root__', name: t('dbImport.driveRoot') })
    }

    if (rootFolders.value.length > 0) {
      selectedRoots.value = new Set(rootFolders.value.map(r => r.path))
      phase.value = 'folders'
    } else {
      // No sub-folders at all — all files are at the Dropbox root, skip folder selection.
      selectedSet.value = new Set(allFiles.value.map(f => f.id))
      phase.value = 'select'
    }
  } catch (err) {
    connectError.value = err.message || t('dbImport.errorConnect')
  } finally {
    connecting.value = false
    fileStore.stopHeartbeat()
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
    const folderPath = f.parentId ? pathMap.value.get(f.parentId) : null
    if (!folderPath) return includeRoot
    const rootPath = '/' + folderPath.split('/').filter(Boolean)[0]
    return selectedRoots.value.has(rootPath)
  })
  allFiles.value = filtered
  selectedSet.value = new Set(filtered.map(f => f.id))

  // Expand all folders that contain at least one file
  const expanded = new Set()
  for (const file of filtered) {
    const fp = file.parentId ? pathMap.value.get(file.parentId) : null
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

// Hands the transfer off to the background import store/widget (ImportManager.vue)
// and closes this dialog immediately — the migration keeps running regardless of
// what the user does afterwards.
function doImport() {
  if (selectedSet.value.size === 0) return

  // Deduplicate by file ID — the same Dropbox file can appear in allFiles more than once
  // if the user went back and re-confirmed.
  const seenIds = new Set()
  const filesToImport = []
  for (const f of allFiles.value) {
    if (selectedSet.value.has(f.id) && !seenIds.has(f.id)) {
      seenIds.add(f.id)
      filesToImport.push(f)
    }
  }

  const started = importStore.start({
    namespace: 'dbImport',
    importer,
    filesToImport,
    pathMap: pathMap.value,
    useDedicatedFolder: useDedicatedFolder.value,
    dedicatedFolderName: dedicatedFolderName.value,
  })

  if (!started) {
    selectError.value = t('dbImport.errorAlreadyRunning')
    return
  }

  emit('update:modelValue', false)
}

// ── Close ──
function handleClose() {
  emit('update:modelValue', false)
}

// Dropbox's API doesn't return MIME types in file metadata, unlike Drive/Graph — icons
// are picked from the file extension instead.
function mimeIcon(fileName) {
  const ext = (fileName.split('.').pop() || '').toLowerCase()
  if (['doc', 'docx'].includes(ext)) return '📄'
  if (['xls', 'xlsx', 'csv'].includes(ext)) return '📊'
  if (['ppt', 'pptx'].includes(ext)) return '📽'
  if (['png', 'jpg', 'jpeg', 'gif', 'webp', 'svg', 'bmp'].includes(ext)) return '🖼'
  if (['mp4', 'mov', 'avi', 'mkv', 'webm'].includes(ext)) return '🎬'
  if (['mp3', 'wav', 'flac', 'aac', 'ogg'].includes(ext)) return '🎵'
  if (ext === 'pdf') return '📑'
  if (['zip', 'rar', '7z', 'tar', 'gz'].includes(ext)) return '🗜'
  return '📃'
}

</script>

<style scoped>
.dbi-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.55);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.dbi-dialog {
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
.dbi-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1.25rem 1.5rem 1rem;
  border-bottom: 1px solid var(--border-color, #e5e7eb);
}

.dbi-header-left {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.dbi-dropbox-icon {
  width: 26px;
  height: 24px;
  flex-shrink: 0;
}

.dbi-header h2 {
  font-size: 1.1rem;
  font-weight: 600;
  margin: 0;
  color: var(--text-color, #1a1a1a);
}

.dbi-close {
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
.dbi-close:hover { color: var(--text-color, #1a1a1a); }

/* Body */
.dbi-body {
  padding: 1.5rem;
  overflow-y: auto;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.dbi-intro {
  color: var(--secondary-text-color, #6b7280);
  line-height: 1.6;
  margin: 0;
  font-size: 0.95rem;
}

/* Connect status */
.dbi-connect-status {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  font-size: 0.875rem;
  color: var(--secondary-text-color, #6b7280);
  padding: 0.5rem 0;
}

/* Info box */
.dbi-info-box {
  background: var(--background-color, #f8fafc);
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 8px;
  padding: 1rem 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}

.dbi-info-row {
  display: flex;
  align-items: flex-start;
  gap: 0.6rem;
  font-size: 0.88rem;
  color: var(--secondary-text-color, #6b7280);
  line-height: 1.5;
}

.dbi-info-icon { flex-shrink: 0; margin-top: 1px; }
.dbi-icon-ok { color: #16a34a; }
.dbi-icon-info { color: #2563eb; }

/* Select phase */
.dbi-select-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.dbi-count-label {
  font-size: 0.9rem;
  color: var(--secondary-text-color, #6b7280);
}

.dbi-select-actions {
  display: flex;
  gap: 0.75rem;
}

.dbi-file-list {
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 8px;
  max-height: 280px;
  overflow-y: auto;
}

.dbi-file-row {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.55rem 0.9rem;
  cursor: pointer;
  transition: background-color 0.1s;
  border-bottom: 1px solid var(--border-color, #f0f0f0);
  font-size: 0.875rem;
}
.dbi-file-row:last-child { border-bottom: none; }
.dbi-file-row:hover { background: var(--background-color, #f8fafc); }

.dbi-checkbox { flex-shrink: 0; cursor: pointer; accent-color: var(--primary-color, #42b983); }
.dbi-file-icon { font-size: 1rem; flex-shrink: 0; }

.dbi-file-name {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--text-color, #1a1a1a);
}

.dbi-file-size {
  font-size: 0.8rem;
  color: var(--secondary-text-color, #9ca3af);
  white-space: nowrap;
  flex-shrink: 0;
}

.dbi-select-summary {
  font-size: 0.88rem;
  color: var(--secondary-text-color, #6b7280);
  text-align: right;
}

/* Shared buttons */
.dbi-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  margin-top: auto;
  padding-top: 0.25rem;
}

.dbi-btn-primary {
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
.dbi-btn-primary:disabled { opacity: 0.5; cursor: default; }
.dbi-btn-primary:not(:disabled):hover { opacity: 0.88; }

.dbi-btn-secondary {
  background: transparent;
  color: var(--secondary-text-color, #6b7280);
  border: 1px solid var(--border-color, #d1d5db);
  border-radius: 8px;
  padding: 0.6rem 1.25rem;
  font-size: 0.9rem;
  cursor: pointer;
  transition: background-color 0.15s;
}
.dbi-btn-secondary:hover { background: var(--background-color, #f8fafc); }

.dbi-btn-text {
  background: none;
  border: none;
  color: var(--primary-color, #42b983);
  font-size: 0.85rem;
  cursor: pointer;
  padding: 0.2rem 0.4rem;
  border-radius: 4px;
  transition: background-color 0.15s;
}
.dbi-btn-text:hover { background: rgba(66, 185, 131, 0.08); }

/* Error */
.dbi-error {
  color: #dc2626;
  font-size: 0.875rem;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 6px;
  padding: 0.6rem 0.9rem;
  margin: 0;
}

/* Spinner */
.dbi-spinner {
  display: inline-block;
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.4);
  border-top-color: #fff;
  border-radius: 50%;
  animation: dbi-spin 0.7s linear infinite;
}
.dbi-spinner-sm {
  width: 12px;
  height: 12px;
  border: 2px solid var(--border-color, #d1d5db);
  border-top-color: var(--primary-color, #42b983);
}

@keyframes dbi-spin {
  to { transform: rotate(360deg); }
}

/* Dedicated folder option */
.dbi-dedicated-option {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding: 0.75rem 1rem;
  background: var(--background-color, #f8fafc);
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 8px;
  font-size: 0.88rem;
}

.dbi-dedicated-checkbox {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  color: var(--text-color, #1a1a1a);
  font-weight: 500;
}

.dbi-dedicated-name-row {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  margin-left: 1.5rem;
}

.dbi-dedicated-prefix {
  font-size: 0.85rem;
  color: var(--secondary-text-color, #6b7280);
  white-space: nowrap;
}

.dbi-dedicated-input {
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
.dbi-dedicated-input:focus { border-color: var(--primary-color, #42b983); }

/* Destination hint in select toolbar */
.dbi-dest-hint {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.78rem;
  color: var(--secondary-text-color, #9ca3af);
  font-family: monospace;
  white-space: nowrap;
}

/* Toolbar back button */
.dbi-toolbar-left {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.dbi-btn-back {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  color: var(--secondary-text-color, #6b7280);
  font-size: 0.85rem;
}
.dbi-btn-back:hover { background: rgba(107, 114, 128, 0.08); }

/* Tree view */
.dbi-tree-dir {
  display: flex;
  align-items: center;
  border-bottom: 1px solid var(--border-color, #f0f0f0);
  background: var(--background-color, #f8fafc);
  font-size: 0.875rem;
  min-height: 36px;
}

.dbi-tree-toggle {
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
.dbi-tree-toggle:hover { background: var(--border-color, #e5e7eb); }

.dbi-tree-chevron {
  transition: transform 0.15s;
  flex-shrink: 0;
}
.dbi-tree-chevron-open { transform: rotate(90deg); }

.dbi-tree-dir-label {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  flex: 1;
  cursor: pointer;
  padding: 0.35rem 0.75rem 0.35rem 0.25rem;
  min-width: 0;
}

.dbi-tree-dir-name {
  flex: 1;
  font-weight: 500;
  color: var(--text-color, #1a1a1a);
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dbi-tree-dir-count {
  font-size: 0.72rem;
  color: var(--secondary-text-color, #9ca3af);
  background: var(--border-color, #e5e7eb);
  border-radius: 10px;
  padding: 1px 6px;
  flex-shrink: 0;
}

.dbi-tree-file {
  /* dbi-file-row covers the rest; extra indent is applied inline */
}

@media (max-width: 600px) {
  .dbi-dialog { border-radius: 8px; max-height: 95vh; }
  .dbi-body { padding: 1rem; }
  .dbi-header { padding: 1rem; }
}
</style>
