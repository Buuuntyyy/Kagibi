<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="file-list-container" @click="closeContextMenu">
    <!-- Breadcrumbs & Navigation -->
    <div class="path-banner">
      <button @click="goUp" class="btn-icon back-btn" :disabled="store.currentPath === '/'" title="Remonter">
        <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M15 18l-6-6 6-6"/>
        </svg>
      </button>

      <div class="breadcrumbs">
        <div v-for="(segment, index) in pathSegments" :key="index" class="breadcrumb-item">
          <span
            class="breadcrumb-link"
            :class="{ 'active': index === pathSegments.length - 1 }"
            @click="navigateToPath(segment.path)"
          >
            {{ segment.name }}
          </span>
          <span v-if="index < pathSegments.length - 1" class="separator">/</span>
        </div>
      </div>
    </div>

    <!-- Toolbar (create / rename / upload) -->
    <div v-if="store.permissions.create || store.permissions.move" class="action-toolbar">
      <button v-if="store.permissions.create" class="toolbar-btn" @click="promptCreateFolder" title="Nouveau dossier">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
          <path d="M20 6h-8l-2-2H4c-1.11 0-1.99.89-1.99 2L2 18c0 1.11.89 2 2 2h16c1.11 0 2-.89 2-2V8c0-1.11-.89-2-2-2zm-1 8h-3v3h-2v-3h-3v-2h3V9h2v3h3v2z"/>
        </svg>
        Nouveau dossier
      </button>
      <button v-if="store.permissions.create" class="toolbar-btn" @click="triggerFileInput" :disabled="store.isUploading" title="Ajouter des fichiers">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
          <path d="M9 16h6v-6h4l-7-7-7 7h4zm-4 2h14v2H5z"/>
        </svg>
        {{ store.isUploading ? `Envoi... ${store.uploadProgress}%` : 'Ajouter des fichiers' }}
      </button>
      <button v-if="store.permissions.move && selectedItems.length === 1" class="toolbar-btn" @click="promptRenameSelected" title="Renommer">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
          <path d="M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25zm17.71-10.21c.39-.39.39-1.02 0-1.41l-2.34-2.34c-.39-.39-1.02-.39-1.41 0l-1.83 1.83 3.75 3.75 1.83-1.83z"/>
        </svg>
        Renommer
      </button>
      <input ref="fileInputRef" type="file" multiple style="display:none" @change="handleFileInputChange" />
    </div>

    <!-- Table -->
    <div class="table-responsive">
      <table class="files-table">
        <thead>
          <tr>
            <th class="col-check">
              <input type="checkbox" class="item-checkbox" :checked="allSelected" :indeterminate="someSelected" @change="toggleSelectAll" />
            </th>
            <th class="col-icon"></th>
            <th class="col-name sortable" @click="toggleSort('name')">
              Nom <span class="sort-arrow">{{ sortArrow('name') }}</span>
            </th>
            <th class="col-date sortable" @click="toggleSort('date')">
              Modifié le <span class="sort-arrow">{{ sortArrow('date') }}</span>
            </th>
            <th class="col-size sortable" @click="toggleSort('size')">
              Taille <span class="sort-arrow">{{ sortArrow('size') }}</span>
            </th>
            <th class="col-actions"></th>
          </tr>
        </thead>
        <tbody>
          <!-- Empty State -->
          <tr v-if="sortedFolders.length === 0 && sortedFiles.length === 0">
             <td colspan="6" class="empty-state">
                <p>Ce dossier est vide.</p>
             </td>
          </tr>

          <!-- Folders -->
          <tr v-for="folder in sortedFolders" :key="'f-' + folder.ID"
              class="list-item folder-item"
              :class="{ selected: isSelected(folder.ID, 'folder') }"
              @click="handleRowClick($event, folder, 'folder')"
              @dblclick="openFolder(folder.Name)"
              @contextmenu.prevent="openContextMenu($event, folder, 'folder')">
            <td class="col-check" @click.stop>
              <input type="checkbox" class="item-checkbox"
                :checked="isSelected(folder.ID, 'folder')"
                @change="toggleSelection(folder, 'folder')" />
            </td>
            <td class="col-icon">
              <svg class="file-icon folder-icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M10 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z" fill="#5f6368"/>
              </svg>
            </td>
            <td class="col-name">
              <span class="name-text">{{ folder.Name }}</span>
              <span v-if="folder.access_level === 'readonly'" class="access-badge readonly" title="Lecture seule">lecture</span>
            </td>
            <td class="col-date">{{ formatDate(folder.CreatedAt) }}</td>
            <td class="col-size">{{ formatSize(folder.SizeBytes) }}</td>
            <td class="col-actions" @click.stop>
              <button class="action-btn more-btn" @click.stop="openContextMenu($event, folder, 'folder')" title="Plus d'actions">
                <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                  <path d="M12 8c1.1 0 2-.9 2-2s-.9-2-2-2-2 .9-2 2 .9 2 2 2zm0 2c-1.1 0-2 .9-2 2s.9 2 2 2 2-.9 2-2-.9-2-2-2zm0 6c-1.1 0-2 .9-2 2s.9 2 2 2 2-.9 2-2-.9-2-2-2z"/>
                </svg>
              </button>
            </td>
          </tr>

          <!-- Files -->
          <tr v-for="file in sortedFiles" :key="'file-' + file.ID"
              class="list-item file-item"
              :class="{ selected: isSelected(file.ID, 'file') }"
              @click="handleRowClick($event, file, 'file')"
              @dblclick="openFile(file)"
              @contextmenu.prevent="openContextMenu($event, file, 'file')">
            <td class="col-check" @click.stop>
              <input type="checkbox" class="item-checkbox"
                :checked="isSelected(file.ID, 'file')"
                @change="toggleSelection(file, 'file')" />
            </td>
            <td class="col-icon">
              <span class="file-icon-wrapper">
                 <svg class="file-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z" fill="#888"/>
                 </svg>
              </span>
            </td>
            <td class="col-name">
              <span class="name-text">{{ file.Name }}</span>
              <span v-if="file.can_download === false" class="access-badge no-download" title="Téléchargement désactivé">
                <svg viewBox="0 0 24 24" width="9" height="9" fill="currentColor" style="margin-right:2px"><path d="M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zm-6 9c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2zm3.1-9H8.9V6c0-1.71 1.39-3.1 3.1-3.1 1.71 0 3.1 1.39 3.1 3.1v2z"/></svg>privé
              </span>
              <span v-else-if="store.permissions.delete && file.can_delete === false" class="access-badge locked" title="Non supprimable">protégé</span>
            </td>
            <td class="col-date">{{ formatDate(file.UpdatedAt) }}</td>
            <td class="col-size">{{ formatSize(file.Size) }}</td>
            <td class="col-actions" @click.stop>
              <button class="action-btn more-btn" @click.stop="openContextMenu($event, file, 'file')" title="Plus d'actions">
                <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                  <path d="M12 8c1.1 0 2-.9 2-2s-.9-2-2-2-2 .9-2 2 .9 2 2 2zm0 2c-1.1 0-2 .9-2 2s.9 2 2 2 2-.9 2-2-.9-2-2-2zm0 6c-1.1 0-2 .9-2 2s.9 2 2 2 2-.9 2-2-.9-2-2-2z"/>
                </svg>
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Floating Selection Action Bar -->
    <div class="selection-gap" :class="{ 'has-content': selectedItems.length > 0 }">
      <Transition name="selection-bar" mode="out-in">
        <div v-if="selectedItems.length > 0" key="selection-bar" class="selection-action-bar">
          <div class="selection-actions">
            <button v-if="store.permissions.download && selectedDownloadableFiles.length > 0"
                class="action-btn download-action"
                @click.stop="downloadSelected"
                title="Télécharger la sélection">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                <path d="M19 9h-4V3H9v6H5l7 7 7-7zm-8 2V5h2v6h1.17L12 13.17 9.83 11H11zm-6 7h14v2H5v-2z"/>
              </svg>
              Télécharger
            </button>
            <button v-if="store.permissions.delete && deletableSelected.length > 0"
                class="action-btn delete-action"
                @click.stop="deleteSelected"
                title="Supprimer la sélection">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                <path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM15.5 4l-1-1h-5l-1 1H5v2h14V4z"/>
              </svg>
              Supprimer
            </button>
          </div>
          <span class="selection-count">{{ selectedItems.length }} élément{{ selectedItems.length > 1 ? 's' : '' }} sélectionné{{ selectedItems.length > 1 ? 's' : '' }}</span>
          <button class="deselect-btn" @click="selectedItems = []" title="Désélectionner tout">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
          </button>
        </div>
      </Transition>
    </div>

    <!-- Floating Context Menu -->
    <Teleport to="body">
      <div v-if="contextMenu.visible"
           class="context-menu"
           :style="{ top: contextMenu.top + 'px', left: contextMenu.left + 'px' }"
           @click.stop>
        <template v-if="contextMenu.type === 'file'">
          <button v-if="canDownloadFile(contextMenu.item)"
              class="ctx-item"
              @click="downloadFile(contextMenu.item); closeContextMenu()">
            <svg viewBox="0 0 24 24" width="15" height="15" fill="currentColor"><path d="M19 9h-4V3H9v6H5l7 7 7-7zm-8 2V5h2v6h1.17L12 13.17 9.83 11H11zm-6 7h14v2H5v-2z"/></svg>
            Télécharger / Prévisualiser
          </button>
          <button v-if="store.permissions.move"
              class="ctx-item"
              @click="promptRename(contextMenu.item, 'file'); closeContextMenu()">
            <svg viewBox="0 0 24 24" width="15" height="15" fill="currentColor"><path d="M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25zm17.71-10.21c.39-.39.39-1.02 0-1.41l-2.34-2.34c-.39-.39-1.02-.39-1.41 0l-1.83 1.83 3.75 3.75 1.83-1.83z"/></svg>
            Renommer
          </button>
          <button v-if="canDeleteFile(contextMenu.item)"
              class="ctx-item ctx-danger"
              @click="deleteFile(contextMenu.item); closeContextMenu()">
            <svg viewBox="0 0 24 24" width="15" height="15" fill="currentColor"><path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM15.5 4l-1-1h-5l-1 1H5v2h14V4z"/></svg>
            Supprimer
          </button>
          <span v-if="!canDownloadFile(contextMenu.item) && !store.permissions.move && !canDeleteFile(contextMenu.item)" class="ctx-empty">
            Aucune action disponible
          </span>
        </template>
        <template v-else-if="contextMenu.type === 'folder'">
          <button class="ctx-item"
              @click="openFolder(contextMenu.item.Name); closeContextMenu()">
            <svg viewBox="0 0 24 24" width="15" height="15" fill="currentColor"><path d="M10 4H4c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z"/></svg>
            Ouvrir
          </button>
          <button v-if="store.permissions.move"
              class="ctx-item"
              @click="promptRename(contextMenu.item, 'folder'); closeContextMenu()">
            <svg viewBox="0 0 24 24" width="15" height="15" fill="currentColor"><path d="M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25zm17.71-10.21c.39-.39.39-1.02 0-1.41l-2.34-2.34c-.39-.39-1.02-.39-1.41 0l-1.83 1.83 3.75 3.75 1.83-1.83z"/></svg>
            Renommer
          </button>
          <button v-if="canDeleteFolder(contextMenu.item)"
              class="ctx-item ctx-danger"
              @click="deleteFolderAction(contextMenu.item); closeContextMenu()">
            <svg viewBox="0 0 24 24" width="15" height="15" fill="currentColor"><path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM15.5 4l-1-1h-5l-1 1H5v2h14V4z"/></svg>
            Supprimer le dossier
          </button>
        </template>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { computed, ref, onMounted, onUnmounted } from 'vue';
import { usePublicFileStore } from '../stores/publicFileStore';

const store = usePublicFileStore();
const fileInputRef = ref(null);

const triggerFileInput = () => {
  if (fileInputRef.value) fileInputRef.value.click();
};

const handleFileInputChange = async (event) => {
  const files = event.target.files;
  if (!files || files.length === 0) return;
  await store.uploadFiles(files);
  event.target.value = '';
};

// Sorting state
const sortKey = ref('name');
const sortDir = ref('asc');

const toggleSort = (key) => {
  if (sortKey.value === key) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc';
  } else {
    sortKey.value = key;
    sortDir.value = 'asc';
  }
};

const sortArrow = (key) => {
  if (sortKey.value !== key) return '↕';
  return sortDir.value === 'asc' ? '↑' : '↓';
};

const sortItems = (items, getDate, getSize) => {
  return [...items].sort((a, b) => {
    let cmp = 0;
    if (sortKey.value === 'name') {
      cmp = (a.Name || '').localeCompare(b.Name || '', 'fr', { sensitivity: 'base' });
    } else if (sortKey.value === 'date') {
      cmp = new Date(getDate(a)) - new Date(getDate(b));
    } else if (sortKey.value === 'size') {
      cmp = (getSize(a) || 0) - (getSize(b) || 0);
    }
    return sortDir.value === 'asc' ? cmp : -cmp;
  });
};

const sortedFolders = computed(() =>
  sortItems(store.folders, f => f.CreatedAt, f => f.SizeBytes)
);

const sortedFiles = computed(() =>
  sortItems(store.files, f => f.UpdatedAt, f => f.Size)
);

// Selection
const selectedItems = ref([]);

const isSelected = (id, type) =>
  selectedItems.value.some(s => s.id === id && s.type === type);

const toggleSelection = (item, type) => {
  const idx = selectedItems.value.findIndex(s => s.id === item.ID && s.type === type);
  if (idx >= 0) {
    selectedItems.value.splice(idx, 1);
  } else {
    selectedItems.value.push({ id: item.ID, type, name: item.Name, item });
  }
};

const allSelectable = computed(() => [
  ...sortedFolders.value.map(f => ({ id: f.ID, type: 'folder' })),
  ...sortedFiles.value.map(f => ({ id: f.ID, type: 'file' })),
]);

const allSelected = computed(() =>
  allSelectable.value.length > 0 &&
  allSelectable.value.every(s => isSelected(s.id, s.type))
);

const someSelected = computed(() =>
  selectedItems.value.length > 0 && !allSelected.value
);

const toggleSelectAll = () => {
  if (allSelected.value) {
    selectedItems.value = [];
  } else {
    selectedItems.value = allSelectable.value.map(s => ({
      id: s.id,
      type: s.type,
      name: s.type === 'folder'
        ? sortedFolders.value.find(f => f.ID === s.id)?.Name
        : sortedFiles.value.find(f => f.ID === s.id)?.Name,
      item: s.type === 'folder'
        ? sortedFolders.value.find(f => f.ID === s.id)
        : sortedFiles.value.find(f => f.ID === s.id),
    }));
  }
};

const selectedDownloadableFiles = computed(() =>
  selectedItems.value.filter(s => {
    if (s.type !== 'file') return false;
    const file = store.files.find(f => f.ID === s.id);
    return file && canDownloadFile(file);
  })
);

const deletableSelected = computed(() =>
  selectedItems.value.filter(s => {
    if (s.type === 'file') {
      const file = store.files.find(f => f.ID === s.id);
      return file && canDeleteFile(file);
    }
    if (s.type === 'folder') {
      const folder = store.folders.find(f => f.ID === s.id);
      return folder && canDeleteFolder(folder);
    }
    return false;
  })
);

// Permission helpers
const canDownloadFile = (file) =>
  store.permissions.download && file.can_download !== false;

const canDeleteFile = (file) =>
  store.permissions.delete && file.can_delete !== false;

const canDeleteFolder = (folder) =>
  store.permissions.delete && folder.can_delete !== false;

// Context menu
const contextMenu = ref({ visible: false, item: null, type: null, top: 0, left: 0 });

const openContextMenu = (event, item, type) => {
  event.stopPropagation();
  const vw = window.innerWidth;
  const vh = window.innerHeight;
  const menuW = 220;
  const menuH = 120;

  let left = event.clientX;
  let top = event.clientY;

  if (left + menuW > vw) left = vw - menuW - 8;
  if (top + menuH > vh) top = vh - menuH - 8;

  contextMenu.value = { visible: true, item, type, top, left };
};

const closeContextMenu = () => {
  contextMenu.value.visible = false;
};

const handleKeyDown = (e) => {
  if (e.key === 'Escape') closeContextMenu();
};

onMounted(() => document.addEventListener('keydown', handleKeyDown));
onUnmounted(() => document.removeEventListener('keydown', handleKeyDown));

const handleRowClick = (event, item, type) => {
  if (event.target.type === 'checkbox') return;
  if (type === 'folder') return;
};

const downloadSelected = () => {
  for (const s of selectedDownloadableFiles.value) {
    store.downloadFile(s.id, s.name);
  }
};

const deleteSelected = async () => {
  const names = deletableSelected.value.map(s => s.name).join(', ');
  if (!confirm(`Supprimer ${deletableSelected.value.length} élément(s) :\n${names}\n\nCette action est irréversible.`)) return;
  for (const s of [...deletableSelected.value]) {
    try {
      if (s.type === 'file') {
        await store.deleteFile(s.id);
      } else {
        await store.deleteFolder(s.id);
      }
    } catch {
      store.showToast(`Impossible de supprimer "${s.name}".`);
    }
  }
  selectedItems.value = selectedItems.value.filter(s => {
    if (s.type === 'file') return store.files.some(f => f.ID === s.id);
    if (s.type === 'folder') return store.folders.some(f => f.ID === s.id);
    return false;
  });
};

const pathSegments = computed(() => {
  const path = store.currentPath;
  const segments = [{ name: store.resourceName || 'Racine', path: '/' }];

  if (path === '/') return segments;

  const parts = path.split('/').filter(p => p);
  let currentBuild = '';

  parts.forEach(part => {
    currentBuild += '/' + part;
    segments.push({ name: part, path: currentBuild });
  });

  return segments;
});

const navigateToPath = (path) => {
  if (path === store.currentPath) return;
  selectedItems.value = [];
  store.fetchItems(store.shareToken, path);
};

const openFolder = (folderName) => {
  selectedItems.value = [];
  store.navigateTo(folderName);
};

const openFile = (file) => {
  if (!canDownloadFile(file)) return;
  store.downloadFile(file.ID, file.Name);
};

const goUp = () => {
  selectedItems.value = [];
  store.navigateUp();
};

const downloadFile = (file) => {
  store.downloadFile(file.ID, file.Name);
};

const deleteFile = async (file) => {
  if (!confirm(`Supprimer "${file.Name}" ? Cette action est irréversible.`)) return;
  try {
    await store.deleteFile(file.ID);
  } catch {
    store.showToast('Impossible de supprimer ce fichier.');
  }
};

const deleteFolderAction = async (folder) => {
  if (!confirm(`Supprimer le dossier "${folder.Name}" et tout son contenu ? Cette action est irréversible.`)) return;
  try {
    await store.deleteFolder(folder.ID);
    selectedItems.value = selectedItems.value.filter(s => !(s.type === 'folder' && s.id === folder.ID));
  } catch {
    store.showToast('Impossible de supprimer ce dossier.');
  }
};

const promptCreateFolder = async () => {
  const name = prompt('Nom du nouveau dossier :');
  if (!name || !name.trim()) return;
  try {
    await store.createFolder(name.trim());
  } catch {
    // store already shows toast
  }
};

const promptRenameSelected = () => {
  if (selectedItems.value.length !== 1) return;
  const s = selectedItems.value[0];
  promptRename(s.item, s.type);
};

const promptRename = async (item, type) => {
  const current = item.Name || item.name || '';
  const newName = prompt('Nouveau nom :', current);
  if (!newName || !newName.trim() || newName.trim() === current) return;
  try {
    await store.renameItem(item.ID || item.id, type, newName.trim());
  } catch {
    // store already shows toast
  }
};

const formatSize = (bytes) => {
  if (!bytes || bytes === 0) return '–';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return Number.parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

const formatDate = (dateString) => {
  if (!dateString) return '-';
  const date = new Date(dateString);
  return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
};
</script>

<style scoped>
.file-list-container {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
  overflow: hidden;
}

/* Path Banner */
.path-banner {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px 24px;
  border-bottom: 1px solid var(--border-color);
  background: var(--card-color);
}

.back-btn {
  background: none;
  border: none;
  color: var(--secondary-text-color);
  cursor: pointer;
  padding: 4px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s, color 0.2s;
}

.back-btn:hover:not(:disabled) {
  background: var(--hover-background-color);
  color: var(--main-text-color);
}

.back-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.breadcrumbs {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
}

.breadcrumb-item {
  display: flex;
  align-items: center;
}

.breadcrumb-link {
  color: var(--secondary-text-color);
  cursor: pointer;
  font-weight: 500;
  padding: 4px 6px;
  border-radius: 4px;
  transition: color 0.2s, background 0.2s;
}

.breadcrumb-link:hover {
  color: var(--primary-color);
  background: var(--hover-background-color);
}

.breadcrumb-link.active {
  color: var(--main-text-color);
  cursor: default;
  font-weight: 600;
  background: none;
}

.separator {
  margin: 0 4px;
  color: var(--border-color);
}

/* Toolbar */
.action-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border-bottom: 1px solid var(--border-color);
  background: var(--card-color);
}

.toolbar-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  background: none;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  padding: 5px 12px;
  font-size: 0.82rem;
  color: var(--main-text-color);
  cursor: pointer;
  transition: background 0.15s, border-color 0.15s;
}

.toolbar-btn:hover {
  background: var(--hover-background-color);
  border-color: var(--primary-color);
  color: var(--primary-color);
}

/* Table */
.table-responsive {
  overflow-x: auto;
  overflow-y: auto;
  flex: 1;
  width: 100%;
}

.files-table {
  width: 100%;
  border-collapse: collapse;
  min-width: 600px;
}

.files-table th {
  text-align: left;
  padding: 12px 16px;
  font-weight: 600;
  color: var(--secondary-text-color);
  border-bottom: 1px solid var(--border-color);
  background: var(--background-color);
  font-size: 0.85rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  position: sticky;
  top: 0;
  white-space: nowrap;
}

.files-table th.sortable {
  cursor: pointer;
  user-select: none;
}

.files-table th.sortable:hover {
  color: var(--primary-color);
}

.sort-arrow {
  opacity: 0.5;
  font-size: 0.75rem;
  margin-left: 4px;
}

th.sortable:hover .sort-arrow {
  opacity: 1;
}

.files-table td {
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-color);
  color: var(--main-text-color);
  vertical-align: middle;
}

.list-item {
  cursor: pointer;
  transition: background-color 0.15s;
}

.list-item:hover {
  background-color: var(--hover-background-color);
}

.list-item.selected {
  background-color: rgba(99, 102, 241, 0.07);
}

.list-item:last-child td {
  border-bottom: none;
}

/* Columns */
.col-check {
  width: 40px;
  text-align: center;
  padding: 12px 8px !important;
}

.col-icon {
  width: 48px;
  text-align: center;
}

.col-name {
  width: 40%;
}

.col-date, .col-size {
  width: 20%;
  color: var(--secondary-text-color);
  font-size: 0.95rem;
}

.col-actions {
  width: 48px;
  text-align: right;
}

.item-checkbox {
  cursor: pointer;
  width: 16px;
  height: 16px;
  accent-color: var(--primary-color);
}

/* Icons & text */
.file-icon {
  width: 24px;
  height: 24px;
  display: block;
}

.name-text {
  font-weight: 500;
}

.folder-item .name-text {
  font-weight: 600;
}

/* Action Buttons */
.action-btn {
  background: none;
  border: none;
  color: var(--secondary-text-color);
  cursor: pointer;
  padding: 6px;
  border-radius: 50%;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.more-btn:hover {
  color: var(--main-text-color);
  background: var(--hover-background-color);
}

/* Access restriction badges */
.access-badge {
  display: inline-flex;
  align-items: center;
  margin-left: 8px;
  padding: 1px 6px;
  border-radius: 10px;
  font-size: 0.7rem;
  font-weight: 600;
  letter-spacing: 0.03em;
  vertical-align: middle;
  text-transform: uppercase;
}

.access-badge.readonly {
  background: rgba(234, 179, 8, 0.12);
  color: #ca8a04;
  border: 1px solid rgba(234, 179, 8, 0.3);
}

.access-badge.locked {
  background: rgba(99, 102, 241, 0.1);
  color: var(--primary-color);
  border: 1px solid rgba(99, 102, 241, 0.25);
}

.access-badge.no-download {
  background: rgba(107, 114, 128, 0.1);
  color: var(--secondary-text-color);
  border: 1px solid rgba(107, 114, 128, 0.25);
}

/* Empty State */
.empty-state {
  text-align: center;
  padding: 40px !important;
  color: var(--secondary-text-color);
  font-style: italic;
}

/* Floating Selection Action Bar */
.selection-gap {
  height: 0;
  overflow: visible;
  position: relative;
}

.selection-action-bar {
  position: fixed;
  bottom: 28px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  align-items: center;
  gap: 12px;
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 10px 16px;
  box-shadow: 0 8px 24px rgba(0,0,0,0.15);
  z-index: 200;
  white-space: nowrap;
}

.selection-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

.selection-action-bar .action-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 7px 14px;
  border-radius: 8px;
  font-size: 0.85rem;
  font-weight: 500;
}

.selection-action-bar .download-action {
  color: var(--primary-color);
  background: rgba(99, 102, 241, 0.08);
}

.selection-action-bar .download-action:hover {
  background: rgba(99, 102, 241, 0.16);
}

.selection-action-bar .delete-action {
  color: var(--error-color, #ef4444);
  background: rgba(239, 68, 68, 0.07);
}

.selection-action-bar .delete-action:hover {
  background: rgba(239, 68, 68, 0.14);
}

.selection-count {
  font-size: 0.85rem;
  color: var(--secondary-text-color);
  font-weight: 500;
  padding: 0 4px;
  border-left: 1px solid var(--border-color);
  padding-left: 12px;
}

.deselect-btn {
  background: none;
  border: none;
  color: var(--secondary-text-color);
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  transition: color 0.2s;
}

.deselect-btn:hover {
  color: var(--main-text-color);
}

/* Selection bar transitions */
.selection-bar-enter-active,
.selection-bar-leave-active {
  transition: opacity 0.2s, transform 0.2s;
}

.selection-bar-enter-from,
.selection-bar-leave-to {
  opacity: 0;
  transform: translateX(-50%) translateY(12px);
}

/* Floating Context Menu */
:global(.context-menu) {
  position: fixed;
  min-width: 200px;
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 4px 0;
  box-shadow: 0 8px 24px rgba(0,0,0,0.15);
  z-index: 9999;
  animation: ctx-fade 0.1s ease;
}

@keyframes ctx-fade {
  from { opacity: 0; transform: scale(0.97); }
  to { opacity: 1; transform: scale(1); }
}

:global(.ctx-item) {
  display: flex;
  align-items: center;
  gap: 9px;
  width: 100%;
  padding: 9px 16px;
  background: none;
  border: none;
  cursor: pointer;
  font-size: 0.875rem;
  color: var(--main-text-color);
  text-align: left;
  transition: background 0.12s;
}

:global(.ctx-item:hover) {
  background: var(--hover-background-color);
}

:global(.ctx-item.ctx-danger) {
  color: var(--error-color, #ef4444);
}

:global(.ctx-item.ctx-danger:hover) {
  background: rgba(239, 68, 68, 0.07);
}

:global(.ctx-empty) {
  display: block;
  padding: 9px 16px;
  font-size: 0.82rem;
  color: var(--secondary-text-color);
  font-style: italic;
}
</style>
