<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="search-bar" ref="searchBarRef">
    <!-- Disabled: filename encryption active -->
    <div v-if="filenamesAreEncrypted"
         class="search-wrapper search-wrapper--disabled"
         :title="t('search.encryptedDisabledTooltip', 'La recherche est désactivée car les noms de fichiers sont chiffrés')">
      <div class="icon-wrapper search-icon">
        <svg focusable="false" viewBox="0 0 24 24" height="22px" width="22px" fill="currentColor">
          <path d="M15.5 14h-.79l-.28-.27A6.471 6.471 0 0 0 16 9.5 6.5 6.5 0 1 0 9.5 16c1.61 0 3.09-.59 4.23-1.57l.27.28v.79l5 4.99L20.49 19l-4.99-5zm-6 0C7.01 14 5 11.99 5 9.5S7.01 5 9.5 5 14 7.01 14 9.5 11.99 14 9.5 14z"/>
        </svg>
      </div>
      <span class="search-disabled-label">{{ t('search.encryptedDisabled', 'Recherche désactivée (noms chiffrés)') }}</span>
    </div>

    <!-- Normal search -->
    <div v-else class="search-wrapper" :class="{ focused: isFocused || showDropdown }">
      <div class="icon-wrapper search-icon">
        <svg focusable="false" viewBox="0 0 24 24" height="22px" width="22px" fill="currentColor">
          <path d="M15.5 14h-.79l-.28-.27A6.471 6.471 0 0 0 16 9.5 6.5 6.5 0 1 0 9.5 16c1.61 0 3.09-.59 4.23-1.57l.27.28v.79l5 4.99L20.49 19l-4.99-5zm-6 0C7.01 14 5 11.99 5 9.5S7.01 5 9.5 5 14 7.01 14 9.5 11.99 14 9.5 14z"/>
        </svg>
      </div>

      <input
        type="text"
        v-model="searchQuery"
        :placeholder="t('search.placeholder')"
        @input="handleInput"
        @focus="handleFocus"
        @keydown.down.prevent="navigateResults(1)"
        @keydown.up.prevent="navigateResults(-1)"
        @keydown.enter.prevent="handleEnter"
      />

      <!-- Active filters pill (visible when query + filters) -->
      <button
        v-if="searchQuery && activeFilterCount > 0"
        class="active-filters-pill"
        @click.stop="clearAllFilters"
        title="Effacer les filtres"
      >
        <span class="active-filters-dot" />
        {{ activeFilterCount }} filtre{{ activeFilterCount > 1 ? 's' : '' }}
        <svg viewBox="0 0 24 24" width="12" height="12" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
      </button>

      <!-- Clear search -->
      <div v-if="searchQuery" class="icon-wrapper clear-icon" @click="clearSearch" title="Effacer">
        <svg focusable="false" viewBox="0 0 24 24" height="20px" width="20px" fill="currentColor">
          <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/>
        </svg>
      </div>
    </div>

    <!-- Committed filter bar: shown after pressing Enter (dropdown closed) -->
    <div
      class="committed-filter-bar"
      v-if="searchCommitted && searchQuery && !showDropdown && (searchResults.files.length > 0 || searchResults.folders.length > 0)"
    >
      <!-- Type chips -->
      <button
        class="filter-chip"
        :class="{ active: fileStore.searchFilterType === 'file' }"
        @click.stop="toggleType('file')"
      >
        <svg viewBox="0 0 24 24" width="13" height="13" fill="currentColor" style="flex-shrink:0">
          <path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm-1 7V3.5L18.5 9H13z"/>
        </svg>
        Fichiers <span class="chip-count">{{ searchResults.files.length }}</span>
      </button>
      <button
        class="filter-chip"
        :class="{ active: fileStore.searchFilterType === 'folder' }"
        @click.stop="toggleType('folder')"
      >
        <svg viewBox="0 0 24 24" width="13" height="13" fill="currentColor" style="flex-shrink:0">
          <path d="M10 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z"/>
        </svg>
        Répertoires <span class="chip-count">{{ searchResults.folders.length }}</span>
      </button>

      <span v-if="availableExtensions.length > 0" class="strip-divider" />

      <!-- Extensions -->
      <button
        v-for="{ ext, count } in availableExtensions"
        :key="'ext-' + ext"
        class="filter-chip"
        :class="{ active: fileStore.searchFilterExtensions.includes(ext) }"
        @click.stop="toggleExtension(ext)"
      >.{{ ext }} <span class="chip-count">{{ count }}</span></button>

      <span v-if="availableExtensions.length > 0 && availableTags.length > 0" class="strip-divider" />

      <!-- Tags -->
      <button
        v-for="tag in availableTags"
        :key="'tag-' + tag.name"
        class="filter-chip tag-chip"
        :class="{ active: fileStore.searchFilterTags.includes(tag.name) }"
        :style="fileStore.searchFilterTags.includes(tag.name)
          ? { background: tag.color, borderColor: tag.color, color: '#fff' }
          : { borderColor: tag.color, color: tag.color }"
        @click.stop="toggleTag(tag.name)"
      >
        <span class="tag-dot" :style="{ background: fileStore.searchFilterTags.includes(tag.name) ? '#fff' : tag.color }" />
        {{ tag.name }}
      </button>
    </div>

    <!-- Unified dropdown: filters + results -->
    <div class="search-dropdown" v-if="showDropdown && searchResults.files.length + searchResults.folders.length > 0">

      <!-- ── Filter strip ─────────────────────────────────────────── -->
      <div
        class="filter-strip"
        v-if="availableExtensions.length > 0 || availableTags.length > 0 || searchResults.files.length > 0 || searchResults.folders.length > 0"
      >
        <!-- Type chips -->
        <button
          class="filter-chip"
          :class="{ active: fileStore.searchFilterType === 'file' }"
          @click.stop="toggleType('file')"
        >
          <svg viewBox="0 0 24 24" width="13" height="13" fill="currentColor" style="flex-shrink:0">
            <path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm-1 7V3.5L18.5 9H13z"/>
          </svg>
          Fichiers <span class="chip-count">{{ searchResults.files.length }}</span>
        </button>
        <button
          class="filter-chip"
          :class="{ active: fileStore.searchFilterType === 'folder' }"
          @click.stop="toggleType('folder')"
        >
          <svg viewBox="0 0 24 24" width="13" height="13" fill="currentColor" style="flex-shrink:0">
            <path d="M10 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z"/>
          </svg>
          Répertoires <span class="chip-count">{{ searchResults.folders.length }}</span>
        </button>

        <!-- Divider between type and extensions -->
        <span
          v-if="availableExtensions.length > 0"
          class="strip-divider"
        />

        <!-- Extensions -->
        <button
          v-for="{ ext, count } in availableExtensions"
          :key="'ext-' + ext"
          class="filter-chip"
          :class="{ active: fileStore.searchFilterExtensions.includes(ext) }"
          @click.stop="toggleExtension(ext)"
        >.{{ ext }} <span class="chip-count">{{ count }}</span></button>

        <!-- Divider between extensions and tags -->
        <span
          v-if="availableExtensions.length > 0 && availableTags.length > 0"
          class="strip-divider"
        />

        <!-- Tags -->
        <button
          v-for="tag in availableTags"
          :key="'tag-' + tag.name"
          class="filter-chip tag-chip"
          :class="{ active: fileStore.searchFilterTags.includes(tag.name) }"
          :style="fileStore.searchFilterTags.includes(tag.name)
            ? { background: tag.color, borderColor: tag.color, color: '#fff' }
            : { borderColor: tag.color, color: tag.color }"
          @click.stop="toggleTag(tag.name)"
        >
          <span class="tag-dot" :style="{ background: fileStore.searchFilterTags.includes(tag.name) ? '#fff' : tag.color }" />
          {{ tag.name }}
        </button>
      </div>

      <!-- ── Results ──────────────────────────────────────────────── -->
      <template v-if="hasFilteredResults">
        <!-- Folders -->
        <div v-if="filteredFolders.length > 0" class="result-group">
          <div class="group-title">{{ t('file.folders') }}</div>
          <div
            v-for="(folder, index) in filteredFolders"
            :key="'folder-' + folder.ID"
            class="result-item"
            :class="{ active: activeIndex === index }"
            @click="openItem(folder, 'folder')"
            @mouseenter="activeIndex = index"
          >
            <div class="item-icon">
              <svg viewBox="0 0 24 24" width="20" height="20" fill="#5f6368">
                <path d="M10 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z"/>
              </svg>
            </div>
            <div class="item-details">
              <span class="item-name">{{ folder.Name }}</span>
              <span class="item-path">{{ folder.Path }}</span>
            </div>
          </div>
        </div>

        <!-- Files -->
        <div v-if="filteredFiles.length > 0" class="result-group">
          <div class="group-title">
            {{ t('file.files') }}
            <span
              v-if="activeFilterCount > 0 && filteredFiles.length < searchResults.files.length"
              class="results-count-badge"
            >{{ filteredFiles.length }} / {{ searchResults.files.length }}</span>
          </div>
          <div
            v-for="(file, index) in filteredFiles"
            :key="'file-' + file.ID"
            class="result-item"
            :class="{ active: activeIndex === (filteredFolders.length + index) }"
            @click="openItem(file, 'file')"
            @mouseenter="activeIndex = filteredFolders.length + index"
          >
            <div class="item-icon">
              <svg viewBox="0 0 24 24" width="20" height="20" fill="#5f6368">
                <path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z"/>
              </svg>
            </div>
            <div class="item-details">
              <span class="item-name">{{ file.Name }}</span>
              <span class="item-path">{{ file.Path }}</span>
            </div>
            <div class="item-actions">
              <button class="action-btn preview-btn" @click.stop="previewFile(file)" title="Visualiser">
                <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                  <path d="M12 4.5C7 4.5 2.73 7.61 1 12c1.73 4.39 6 7.5 11 7.5s9.27-3.11 11-7.5c-1.73-4.39-6-7.5-11-7.5zM12 17c-2.76 0-5-2.24-5-5s2.24-5 5-5 5 2.24 5 5-2.24 5-5 5zm0-8c-1.66 0-3 1.34-3 3s1.34 3 3 3 3-1.34 3-3-1.34-3-3-3z"/>
                </svg>
              </button>
              <button class="action-btn download-btn" @click.stop="downloadFile(file)" title="Télécharger">
                <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                  <path d="M19 9h-4V3H9v6H5l7 7 7-7zM5 18v2h14v-2H5z"/>
                </svg>
              </button>
            </div>
          </div>
        </div>
      </template>

      <!-- No results after filtering -->
      <div v-else class="no-results">
        <svg viewBox="0 0 24 24" width="32" height="32" fill="currentColor" style="opacity:.25;margin-bottom:6px">
          <path d="M15.5 14h-.79l-.28-.27A6.471 6.471 0 0 0 16 9.5 6.5 6.5 0 1 0 9.5 16c1.61 0 3.09-.59 4.23-1.57l.27.28v.79l5 4.99L20.49 19l-4.99-5zm-6 0C7.01 14 5 11.99 5 9.5S7.01 5 9.5 5 14 7.01 14 9.5 11.99 14 9.5 14z"/>
        </svg>
        <span>Aucun résultat pour ces filtres</span>
        <button class="no-results-reset" @click.stop="clearAllFilters">Réinitialiser les filtres</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useFileStore } from '../../stores/files'
import { useAuthStore } from '../../stores/auth'
import { useTagStore } from '../../stores/tags'
import api from '../../api'
import { debounce } from 'lodash'

const { t } = useI18n()
const router = useRouter()
const fileStore = useFileStore()
const authStore = useAuthStore()
const tagStore = useTagStore()

const searchQuery = ref('')
const isFocused = ref(false)
const searchResults = ref({ folders: [], files: [] })
const showDropdown = ref(false)
const activeIndex = ref(-1)
const searchBarRef = ref(null)
const searchCommitted = ref(false)

const filenamesAreEncrypted = computed(() => authStore.user?.encrypt_filenames === true)

// ── Filters ──────────────────────────────────────────────────────────────────

const activeFilterCount = computed(
  () =>
    fileStore.searchFilterExtensions.length +
    fileStore.searchFilterTags.length +
    (fileStore.searchFilterType ? 1 : 0)
)

const toggleExtension = (ext) => {
  const idx = fileStore.searchFilterExtensions.indexOf(ext)
  if (idx >= 0) fileStore.searchFilterExtensions.splice(idx, 1)
  else fileStore.searchFilterExtensions.push(ext)
}

const toggleTag = (name) => {
  const idx = fileStore.searchFilterTags.indexOf(name)
  if (idx >= 0) fileStore.searchFilterTags.splice(idx, 1)
  else fileStore.searchFilterTags.push(name)
}

const toggleType = (type) => {
  fileStore.searchFilterType = fileStore.searchFilterType === type ? null : type
}

const clearAllFilters = () => {
  fileStore.searchFilterExtensions = []
  fileStore.searchFilterTags = []
  fileStore.searchFilterType = null
}

// Extensions present in the current search results
const availableExtensions = computed(() => {
  const counts = {}
  for (const file of searchResults.value.files) {
    const name = file.Name || file.name || ''
    const dot = name.lastIndexOf('.')
    if (dot > 0) {
      const ext = name.slice(dot + 1).toLowerCase()
      counts[ext] = (counts[ext] || 0) + 1
    }
  }
  return Object.entries(counts)
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([ext, count]) => ({ ext, count }))
})

// Tags present in the current search results (matched against tagStore for colors)
const availableTags = computed(() => {
  const nameSet = new Set()
  for (const file of searchResults.value.files) {
    if (file.Tags) file.Tags.forEach(t => nameSet.add(t))
  }
  for (const folder of searchResults.value.folders) {
    if (folder.Tags) folder.Tags.forEach(t => nameSet.add(t))
  }
  if (!nameSet.size) return []
  return [...nameSet]
    .map(name => tagStore.tags.find(t => t.name === name) || { name, color: 'var(--secondary-text-color)' })
    .sort((a, b) => a.name.localeCompare(b.name))
})

// Filtered results for display in the dropdown
const filteredFiles = computed(() => {
  if (fileStore.searchFilterType === 'folder') return []
  let files = searchResults.value.files
  if (fileStore.searchFilterExtensions.length > 0) {
    files = files.filter(file => {
      const name = file.Name || file.name || ''
      const dot = name.lastIndexOf('.')
      const ext = dot > 0 ? name.slice(dot + 1).toLowerCase() : ''
      return fileStore.searchFilterExtensions.includes(ext)
    })
  }
  if (fileStore.searchFilterTags.length > 0) {
    files = files.filter(file => {
      const tags = file.Tags || []
      return fileStore.searchFilterTags.every(t => tags.includes(t))
    })
  }
  return files
})

const filteredFolders = computed(() => {
  if (fileStore.searchFilterType === 'file') return []
  let folders = searchResults.value.folders
  if (fileStore.searchFilterTags.length > 0) {
    folders = folders.filter(folder => {
      const tags = folder.Tags || []
      return fileStore.searchFilterTags.every(t => tags.includes(t))
    })
  }
  return folders
})

const hasFilteredResults = computed(
  () => filteredFolders.value.length > 0 || filteredFiles.value.length > 0
)

// ── Path helpers ─────────────────────────────────────────────────────────────

// file.Path = "/dir/file.txt" → "/dir/"
// folder.Path = "/dir/sub" → "/dir/"
const getParentPath = (path) => {
  if (!path || path === '/') return '/'
  const p = path.replace(/\/$/, '')
  const idx = p.lastIndexOf('/')
  return idx <= 0 ? '/' : p.slice(0, idx + 1)
}

// ── Search ────────────────────────────────────────────────────────────────────

const debouncedSearch = debounce(async (query) => {
  if (!query?.trim()) {
    searchResults.value = { folders: [], files: [] }
    showDropdown.value = false
    return
  }
  try {
    const { data } = await api.get('/files/search', { params: { q: query } })
    searchResults.value = { folders: data.folders || [], files: data.files || [] }
    showDropdown.value = true
    activeIndex.value = -1
  } catch {
    searchResults.value = { folders: [], files: [] }
  }
}, 300)

const handleInput = () => {
  searchCommitted.value = false
  debouncedSearch(searchQuery.value)
  fileStore.setSearchQuery(searchQuery.value)
}

const handleFocus = () => {
  isFocused.value = true
  if (searchQuery.value && (searchResults.value.folders.length || searchResults.value.files.length)) {
    showDropdown.value = true
    searchCommitted.value = false
  }
}

const clearSearch = () => {
  searchQuery.value = ''
  searchResults.value = { folders: [], files: [] }
  showDropdown.value = false
  searchCommitted.value = false
  fileStore.setSearchQuery('')
}

// ── Navigation ────────────────────────────────────────────────────────────────

const navigateResults = (direction) => {
  if (!showDropdown.value) return
  const total = filteredFolders.value.length + filteredFiles.value.length
  if (!total) return
  activeIndex.value = (activeIndex.value + direction + total) % total
}

const handleEnter = () => {
  if (activeIndex.value !== -1) {
    // An item is highlighted: open it
    const fLen = filteredFolders.value.length
    if (activeIndex.value < fLen) openItem(filteredFolders.value[activeIndex.value], 'folder')
    else openItem(filteredFiles.value[activeIndex.value - fLen], 'file')
  } else {
    // No item selected: commit search, close dropdown, show filter bar
    showDropdown.value = false
    searchCommitted.value = true
  }
}

const openItem = (item, type = 'file') => {
  const parentPath = getParentPath(item.Path || item.path)
  fileStore.pendingHighlight = { id: item.ID || item.id, type }
  fileStore.clearSearch()
  if (router.currentRoute.value.name !== 'MyFiles') {
    fileStore.pendingNavigatePath = parentPath
    router.push({ name: 'MyFiles' })
  } else {
    fileStore.fetchItems(parentPath)
  }
  showDropdown.value = false
  searchCommitted.value = false
  searchQuery.value = ''
}

const previewFile = (file) => {
  fileStore.downloadFile(
    file.ID || file.id, file.Name || file.name,
    file.MimeType || file.mime_type || 'application/octet-stream',
    true, file.EncryptedKey || file.encrypted_key
  )
  showDropdown.value = false
  searchCommitted.value = false
  searchQuery.value = ''
}

const downloadFile = (file) => {
  fileStore.downloadFile(
    file.ID || file.id, file.Name || file.name,
    file.MimeType || file.mime_type || 'application/octet-stream',
    false, file.EncryptedKey || file.encrypted_key
  )
  showDropdown.value = false
  searchCommitted.value = false
  searchQuery.value = ''
}

// Sync local input when search is cleared externally (e.g. from fileList navigation)
watch(() => fileStore.searchQuery, (val) => {
  if (!val && searchQuery.value) {
    searchQuery.value = ''
    searchCommitted.value = false
    searchResults.value = { folders: [], files: [] }
    showDropdown.value = false
  }
})

// ── Click outside ─────────────────────────────────────────────────────────────

const handleClickOutside = (e) => {
  if (searchBarRef.value && !searchBarRef.value.contains(e.target)) {
    showDropdown.value = false
    isFocused.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  tagStore.fetchTags()
})
onUnmounted(() => document.removeEventListener('click', handleClickOutside))
</script>

<style scoped>
.search-bar {
  flex: 1 1 0;
  min-width: 0;
  display: flex;
  justify-content: center;
  margin: 0 2rem;
  max-width: 720px;
  position: relative;
}

/* ── Search wrapper ───────────────────────────────────────────── */
.search-wrapper {
  display: flex;
  align-items: center;
  background-color: var(--hover-background-color);
  border: 1px solid var(--border-color);
  border-radius: 24px;
  padding: 0 8px;
  width: 100%;
  max-width: 700px;
  transition: background-color 0.1s, box-shadow 0.1s, border-color 0.1s;
  height: 40px;
  gap: 2px;
}

.search-wrapper--disabled {
  opacity: 0.55;
  cursor: not-allowed;
  pointer-events: none;
}

.search-wrapper.focused,
.search-wrapper:focus-within {
  background-color: var(--card-color);
  border-color: var(--primary-color);
  box-shadow: 0 1px 1px 0 rgba(65,69,73,.3), 0 1px 3px 1px rgba(65,69,73,.15);
}

.icon-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 8px;
  border-radius: 50%;
  cursor: pointer;
  flex-shrink: 0;
  color: var(--secondary-text-color);
}
.icon-wrapper:hover { background-color: var(--hover-background-color); }

.search-icon { margin-left: 4px; }
.clear-icon  { margin-right: 2px; }
.clear-icon:hover { color: var(--main-text-color); }

.search-disabled-label {
  flex-grow: 1;
  padding: 0 8px;
  font-size: 14px;
  color: var(--secondary-text-color);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.search-bar input {
  flex-grow: 1;
  border: none;
  background: transparent;
  padding: 0 4px;
  font-size: 15px;
  color: var(--main-text-color);
  outline: none;
  height: 100%;
  min-width: 0;
}
.search-bar input::placeholder { color: var(--secondary-text-color); }

/* Active filters pill inside the input row */
.active-filters-pill {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--primary-color) 15%, transparent);
  border: 1px solid color-mix(in srgb, var(--primary-color) 40%, transparent);
  color: var(--primary-color);
  font-size: 0.72rem;
  font-weight: 600;
  white-space: nowrap;
  cursor: pointer;
  flex-shrink: 0;
  transition: background 0.15s;
}
.active-filters-pill:hover {
  background: color-mix(in srgb, var(--primary-color) 25%, transparent);
}
.active-filters-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--primary-color);
  flex-shrink: 0;
}

/* ── Committed filter bar ─────────────────────────────────────── */
.committed-filter-bar {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 4px;
  padding: 6px 10px;
  background: var(--card-color);
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(32,33,36,.14), 0 1px 4px rgba(32,33,36,.08);
  border: 1px solid var(--border-color);
  z-index: 2000;
}

.committed-filter-bar .filter-chip {
  font-size: 0.72rem;
  padding: 3px 8px;
}

.committed-filter-bar .strip-divider {
  height: 14px;
}

/* ── Dropdown ─────────────────────────────────────────────────── */
.search-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  background: var(--card-color);
  border-radius: 12px;
  box-shadow: 0 8px 24px rgba(32,33,36,.18), 0 2px 6px rgba(32,33,36,.1);
  border: 1px solid var(--border-color);
  z-index: 2000;
  overflow: hidden;
}

/* ── Filter strip ─────────────────────────────────────────────── */
.filter-strip {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  padding: 10px 14px;
  border-bottom: 1px solid var(--border-color);
  background: var(--hover-background-color);
}

.filter-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  border-radius: 999px;
  border: 1px solid var(--border-color);
  background: var(--card-color);
  color: var(--secondary-text-color);
  font-size: 0.78rem;
  font-weight: 500;
  cursor: pointer;
  transition: border-color 0.15s, background 0.15s, color 0.15s;
  white-space: nowrap;
  line-height: 1.4;
}
.filter-chip:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
}
.filter-chip.active {
  background: var(--primary-color);
  border-color: var(--primary-color);
  color: #fff;
}
.filter-chip.active .chip-count { opacity: 0.75; }

.chip-count {
  font-size: 0.7rem;
  opacity: 0.6;
}

.strip-divider {
  width: 1px;
  height: 18px;
  background: var(--border-color);
  flex-shrink: 0;
  margin: 0 2px;
}

.tag-chip { font-weight: 500; }
.tag-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
}

/* ── Results ──────────────────────────────────────────────────── */
.result-group { padding: 6px 0; }

.result-group:not(:first-child) {
  border-top: 1px solid var(--border-color);
}

.group-title {
  padding: 6px 16px 4px;
  font-size: 0.75rem;
  font-weight: 700;
  color: var(--secondary-text-color);
  text-transform: uppercase;
  letter-spacing: .04em;
  display: flex;
  align-items: center;
  gap: 8px;
}

.results-count-badge {
  background: color-mix(in srgb, var(--primary-color) 15%, transparent);
  color: var(--primary-color);
  font-size: 0.7rem;
  font-weight: 700;
  padding: 1px 7px;
  border-radius: 999px;
  text-transform: none;
  letter-spacing: 0;
}

.result-item {
  display: flex;
  align-items: center;
  padding: 8px 16px;
  cursor: pointer;
  transition: background-color 0.1s;
  position: relative;
}
.result-item:hover,
.result-item.active { background-color: var(--hover-background-color); }
.result-item:hover .item-actions {
  opacity: 1;
  visibility: visible;
}

.item-icon { margin-right: 12px; display: flex; align-items: center; }

.item-details {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  flex: 1;
}
.item-name {
  font-size: 0.9rem;
  color: var(--main-text-color);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.item-path {
  font-size: 0.75rem;
  color: var(--secondary-text-color);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.item-actions {
  display: flex;
  gap: 4px;
  margin-left: 8px;
  opacity: 0;
  visibility: hidden;
  transition: opacity 0.2s, visibility 0.2s;
}
.action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border: none;
  background: transparent;
  border-radius: 50%;
  cursor: pointer;
  color: var(--secondary-text-color);
  transition: background-color 0.2s, color 0.2s;
  padding: 0;
}
.action-btn:hover   { background-color: var(--hover-background-color); }
.preview-btn:hover  { color: #1a73e8; }
.download-btn:hover { color: #34a853; }
.action-btn:active  { transform: scale(0.95); }

/* ── No results ───────────────────────────────────────────────── */
.no-results {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 24px 16px;
  color: var(--secondary-text-color);
  font-size: 0.85rem;
  gap: 8px;
}
.no-results-reset {
  background: none;
  border: 1px solid var(--border-color);
  border-radius: 999px;
  padding: 4px 14px;
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--primary-color);
  cursor: pointer;
  transition: background 0.15s;
}
.no-results-reset:hover { background: var(--hover-background-color); }
</style>
