<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="search-bar" ref="searchBarRef">
    <!-- Disabled state when filename encryption is active -->
    <div v-if="filenamesAreEncrypted" class="search-wrapper search-wrapper--disabled" :title="t('search.encryptedDisabledTooltip', 'La recherche est désactivée car les noms de fichiers sont chiffrés')">
      <div class="icon-wrapper search-icon">
        <svg focusable="false" viewBox="0 0 24 24" height="24px" width="24px" fill="#5f6368">
          <path d="M15.5 14h-.79l-.28-.27A6.471 6.471 0 0 0 16 9.5 6.5 6.5 0 1 0 9.5 16c1.61 0 3.09-.59 4.23-1.57l.27.28v.79l5 4.99L20.49 19l-4.99-5zm-6 0C7.01 14 5 11.99 5 9.5S7.01 5 9.5 5 14 7.01 14 9.5 11.99 14 9.5 14z"></path>
        </svg>
      </div>
      <span class="search-disabled-label">{{ t('search.encryptedDisabled', 'Recherche désactivée (noms chiffrés)') }}</span>
      <div class="icon-wrapper filter-icon">
        <svg focusable="false" viewBox="0 0 24 24" height="24px" width="24px" fill="#5f6368">
          <path d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4zm0 10.99h7c-.53 4.12-3.28 7.79-7 8.94V12H5V6.3l7-3.11v8.8z"/>
        </svg>
      </div>
    </div>

    <!-- Normal search -->
    <div v-else class="search-wrapper" :class="{ focused: isFocused || showDropdown }">
      <div class="icon-wrapper search-icon">
        <svg focusable="false" viewBox="0 0 24 24" height="24px" width="24px" fill="#5f6368">
          <path d="M15.5 14h-.79l-.28-.27A6.471 6.471 0 0 0 16 9.5 6.5 6.5 0 1 0 9.5 16c1.61 0 3.09-.59 4.23-1.57l.27.28v.79l5 4.99L20.49 19l-4.99-5zm-6 0C7.01 14 5 11.99 5 9.5S7.01 5 9.5 5 14 7.01 14 9.5 11.99 14 9.5 14z"></path>
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
        @keydown.enter.prevent="openSelectedResult"
      />
      <div class="icon-wrapper filter-icon" v-if="searchQuery" @click="clearSearch">
        <svg focusable="false" viewBox="0 0 24 24" height="24px" width="24px" fill="#5f6368">
          <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"></path>
        </svg>
      </div>
      <div class="icon-wrapper filter-icon" v-else>
        <svg focusable="false" viewBox="0 0 24 24" height="24px" width="24px" fill="#5f6368">
          <path d="M3 17v2h6v-2H3zM3 5v2h10V5H3zm10 16v-2h8v-2h-8v-2h-2v6h2zM7 9v2H3v2h4v2h2V9H7zm14 4v-2H11v2h10zm-6-4h2V7h4V5h-4V3h-2v6z"></path>
        </svg>
      </div>
    </div>

    <!-- Dropdown Results -->
    <div class="search-dropdown" v-if="showDropdown && hasResults">
      <div v-if="searchResults.folders.length > 0" class="result-group">
        <div class="group-title">{{ t('file.folders') }}</div>
        <div 
          v-for="(folder, index) in searchResults.folders" 
          :key="'folder-' + folder.ID"
          class="result-item"
          :class="{ active: activeIndex === index }"
          @click="openItem(folder, 'folder')"
          @mouseenter="activeIndex = index"
        >
          <div class="item-icon">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="#5f6368"><path d="M10 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z"/></svg>
          </div>
          <div class="item-details">
            <span class="item-name">{{ folder.Name }}</span>
            <span class="item-path">{{ folder.Path }}</span>
          </div>
        </div>
      </div>

      <div v-if="searchResults.files.length > 0" class="result-group">
        <div class="group-title">{{ t('file.files') }}</div>
        <div 
          v-for="(file, index) in searchResults.files" 
          :key="'file-' + file.ID"
          class="result-item"
          :class="{ active: activeIndex === (searchResults.folders.length + index) }"
          @click="openItem(file, 'file')"
          @mouseenter="activeIndex = searchResults.folders.length + index"
        >
          <div class="item-icon">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="#5f6368"><path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z"/></svg>
          </div>
          <div class="item-details">
            <span class="item-name">{{ file.Name }}</span>
            <span class="item-path">{{ file.Path }}</span>
          </div>
          <div class="item-actions">
            <button 
              class="action-btn preview-btn" 
              @click.stop="previewFile(file)"
              title="Visualiser"
            >
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                <path d="M12 4.5C7 4.5 2.73 7.61 1 12c1.73 4.39 6 7.5 11 7.5s9.27-3.11 11-7.5c-1.73-4.39-6-7.5-11-7.5zM12 17c-2.76 0-5-2.24-5-5s2.24-5 5-5 5 2.24 5 5-2.24 5-5 5zm0-8c-1.66 0-3 1.34-3 3s1.34 3 3 3 3-1.34 3-3-1.34-3-3-3z"/>
              </svg>
            </button>
            <button 
              class="action-btn download-btn" 
              @click.stop="downloadFile(file)"
              title="Télécharger"
            >
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                <path d="M19 9h-4V3H9v6H5l7 7 7-7zM5 18v2h14v-2H5z"/>
              </svg>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useFileStore } from '../../stores/files'
import { useAuthStore } from '../../stores/auth'
import api from '../../api'
import { debounce } from 'lodash';

const router = useRouter();
const fileStore = useFileStore();
const authStore = useAuthStore();
const searchQuery = ref('');
const isFocused = ref(false);

// Search is disabled when the user opted into client-side filename encryption.
const filenamesAreEncrypted = computed(() => authStore.user?.encrypt_filenames === true);
const searchResults = ref({ folders: [], files: [] });
const showDropdown = ref(false);
const activeIndex = ref(-1);
const searchBarRef = ref(null);

const hasResults = computed(() => searchResults.value.folders.length > 0 || searchResults.value.files.length > 0);

// Debounce pour éviter de spammer l'API à chaque frappe
const debouncedSearch = debounce(async (query) => {
  if (!query || query.trim() === '') {
    searchResults.value = { folders: [], files: [] };
    showDropdown.value = false;
    return;
  }

  try {
    // Recherche globale via API (découplée de la vue actuelle)
    const response = await api.get('/files/search', { params: { q: query } });
    const data = response.data;
    
    // Adaptation selon la structure de réponse (supposée { folders: [], files: [] })
    searchResults.value = {
      folders: data.folders || [],
      files: data.files || []
    };
    
    showDropdown.value = true;
    activeIndex.value = -1; // Reset selection
  } catch (error) {
    console.error("Erreur de recherche:", error);
    searchResults.value = { folders: [], files: [] };
  }
}, 300);

const handleInput = () => {
  debouncedSearch(searchQuery.value);
};

const handleFocus = () => {
  isFocused.value = true;
  if (searchQuery.value && hasResults.value) {
    showDropdown.value = true;
  }
};

const clearSearch = () => {
  searchQuery.value = '';
  searchResults.value = { folders: [], files: [] };
  showDropdown.value = false;
};

const navigateResults = (direction) => {
  if (!showDropdown.value) return;
  const total = searchResults.value.folders.length + searchResults.value.files.length;
  activeIndex.value = (activeIndex.value + direction + total) % total;
};

const openSelectedResult = () => {
  if (activeIndex.value === -1) return;
  
  const folders = searchResults.value.folders;
  const files = searchResults.value.files;
  
  if (activeIndex.value < folders.length) {
    openItem(folders[activeIndex.value], 'folder');
  } else {
    openItem(files[activeIndex.value - folders.length], 'file');
  }
};

const openItem = (item, type) => {
  if (type === 'folder') {
    // Navigation vers le dossier
    fileStore.pendingNavigatePath = item.Path || item.path;
    if (router.currentRoute.value.name !== 'MyFiles') {
      router.push({ name: 'MyFiles' });
    } else {
      fileStore.fetchItems(item.Path || item.path);
    }
  } else {
    // Ouverture/Aperçu du fichier
    const fileId = item.ID || item.id;
    const fileName = item.Name || item.name;
    const mimeType = item.MimeType || item.mime_type || 'application/octet-stream';
    const encryptedKey = item.EncryptedKey || item.encrypted_key;
    
    fileStore.downloadFile(fileId, fileName, mimeType, true, encryptedKey);
  }
  showDropdown.value = false;
  searchQuery.value = '';
};

const previewFile = (file) => {
  const fileId = file.ID || file.id;
  const fileName = file.Name || file.name;
  const mimeType = file.MimeType || file.mime_type || 'application/octet-stream';
  const encryptedKey = file.EncryptedKey || file.encrypted_key;
  
  fileStore.downloadFile(fileId, fileName, mimeType, true, encryptedKey);
  showDropdown.value = false;
  searchQuery.value = '';
};

const downloadFile = (file) => {
  const fileId = file.ID || file.id;
  const fileName = file.Name || file.name;
  const mimeType = file.MimeType || file.mime_type || 'application/octet-stream';
  const encryptedKey = file.EncryptedKey || file.encrypted_key;
  
  fileStore.downloadFile(fileId, fileName, mimeType, false, encryptedKey);
  showDropdown.value = false;
  searchQuery.value = '';
};

const handleClickOutside = (event) => {
  if (searchBarRef.value && !searchBarRef.value.contains(event.target)) {
    showDropdown.value = false;
  }
};

onMounted(() => {
  document.addEventListener('click', handleClickOutside);
});

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside);
});
</script>

<style scoped>
.search-bar {
  flex-grow: 1;
  display: flex;
  justify-content: center;
  margin: 0 2rem;
  max-width: 720px;
  width: 100%;
  position: relative; /* Pour le positionnement du dropdown */
}

.search-wrapper {
  display: flex;
  align-items: center;
  background-color: var(--hover-background-color);
  border-radius: 24px;
  padding: 0 8px;
  width: 100%;
  max-width: 700px;
  transition: background-color 0.1s, box-shadow 0.1s;
  height: 40px;
}

.search-wrapper--disabled {
  opacity: 0.55;
  cursor: not-allowed;
  pointer-events: none;
}

.search-disabled-label {
  flex-grow: 1;
  padding: 0 8px;
  font-size: 14px;
  color: var(--secondary-text-color);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.search-wrapper.focused,
.search-wrapper:focus-within {
  background-color: var(--card-color);
  box-shadow: 0 1px 1px 0 rgba(65,69,73,0.3), 0 1px 3px 1px rgba(65,69,73,0.15);
}

.icon-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 8px;
  border-radius: 50%;
  cursor: pointer;
}

.icon-wrapper:hover {
  background-color: var(--hover-background-color);
}

.search-icon {
  margin-left: 4px;
}

.filter-icon {
  margin-right: 4px;
}

.search-bar input {
  flex-grow: 1;
  border: none;
  background: transparent;
  padding: 0 8px;
  font-size: 16px;
  color: var(--main-text-color);
  outline: none;
  height: 100%;
}

.search-bar input::placeholder {
  color: var(--secondary-text-color);
}

/* Dropdown Styles */
.search-dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  background: var(--card-color);
  border-radius: 0 0 8px 8px;
  margin-top: 2px;
  box-shadow: 0 4px 6px rgba(32,33,36,0.28);
  z-index: 2000;
  max-height: 400px;
  overflow-y: auto;
  padding: 8px 0;
  border-top: 1px solid var(--border-color);
}

.result-group {
  padding-bottom: 8px;
}

.group-title {
  padding: 8px 16px 4px;
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--secondary-text-color);
  text-transform: uppercase;
}

.result-item {
  display: flex;
  align-items: center;
  padding: 8px 16px;
  cursor: pointer;
  transition: background-color 0.1s;
  position: relative;
}

.result-item:hover, .result-item.active {
  background-color: var(--hover-background-color);
}

.result-item:hover .item-actions {
  opacity: 1;
  visibility: visible;
}

.item-icon {
  margin-right: 12px;
  display: flex;
  align-items: center;
}

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

/* Action Buttons */
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
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  border-radius: 50%;
  cursor: pointer;
  color: var(--secondary-text-color);
  transition: background-color 0.2s, color 0.2s;
  padding: 0;
}

.action-btn:hover {
  background-color: var(--hover-background-color);
  color: var(--main-text-color);
}

.preview-btn:hover {
  color: #1a73e8;
}

.download-btn:hover {
  color: #34a853;
}

.action-btn:active {
  transform: scale(0.95);
}
</style>
