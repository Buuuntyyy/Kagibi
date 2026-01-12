<template>
  <div class="file-list-container"
       @dragover.prevent="onDragOver"
       @dragleave.prevent="onDragLeave"
       @drop.prevent="onDrop">
       
    <div v-if="isDragging" class="drag-overlay">
      <div class="drag-content">
        <span class="drag-icon">☁️</span>
        <span class="drag-text">Déposez vos fichiers ici</span>
      </div>
    </div>
    
    <div class="toolbar" v-if="preferenceStore.showToolBar">
      <div class="toolbar-left">
        <button @click="triggerFileInput" class="btn-add-file">Ajouter un fichier</button>
        <button @click="createNewFolder" class="btn-add-file">Créer un dossier</button>
      </div>
      <div class="toolbar-right">
        <button @click="renameSelectedItem" :disabled="selectedItems.length !== 1" class="btn-rename">
          Renommer
        </button>
        <button @click="downloadSelectedFiles" :disabled="selectedItems.length === 0" class="btn-download">
          Télécharger
        </button>
        <button @click="deleteSelectedItems" :disabled="selectedItems.length === 0" class="btn-delete">
          Supprimer
        </button>
      </div>
    </div>
    <input type="file" ref="fileInput" @change="handleFileUpload" style="display: none" />
    <div class="path-banner">
      
      <div class="breadcrumbs">
        <span v-for="(segment, index) in pathSegments" :key="index" class="breadcrumb-segment">
          <span 
            class="breadcrumb-link" 
            :class="{ 'current': index === pathSegments.length - 1 }"
            @click="navigateToPath(segment.path)"
          >
            {{ segment.name }}
          </span>
          <span v-if="index < pathSegments.length - 1" class="separator">/</span>
        </span>
      </div>
    </div>
    
    <!-- Upload Progress Popup -->
    <div v-if="fileStore.isUploading" class="upload-popup">
      <div class="popup-header">
        <span class="popup-title">Upload en cours...</span>
        <button @click="closeUploadPopup" class="btn-close">×</button>
      </div>
      <div class="popup-content">
        <div class="file-name" :title="fileStore.uploadingFileName">{{ fileStore.uploadingFileName }}</div>
        <div class="progress-container-popup">
          <div class="progress-bar">
            <div class="progress-fill" :style="{ width: fileStore.uploadProgress + '%' }"></div>
          </div>
          <span class="progress-text">{{ fileStore.uploadProgress }}%</span>
        </div>
      </div>
    </div>
    <FileTable 
      :folders="filteredFolders"
      :files="filteredFiles"
      :selectedItems="selectedItems"
      :columns="columns"
      :sortKey="currentSortKey"
      :sortDirection="currentSortDirection"
      @sort-change="handleSortChange"
      @select-item="selectItem"
      @open-folder="openFolder"
      @open-file="downloadFile"
      @context-menu="openContextMenu"
      @drag-start="onDragStart"
      @drop-on-folder="onDropOnFolder"
      @folder-drag-over="onFolderDragOver"
      @folder-drag-leave="onFolderDragLeave"
      @manage-share="openManageShareDialog"
      @remove-tag="removeTag"
    />


  <div
    class="context-menu"
    v-if="contextMenu.visible"
    :style="{ top: contextMenu.y + 'px', left: contextMenu.x + 'px' }">
      <template v-if="contextMenu.item">
        <div class="menu-item" @click="handleContextAction('download')" v-if="contextMenu.item.type === 'file'">
          <span class="menu-icon"><svg xmlns="http://www.w3.org/2000/svg" height="20px" viewBox="0 0 24 24" width="20px" fill="#5f6368"><path d="M0 0h24v24H0V0z" fill="none"/><path d="M19 9h-4V3H9v6H5l7 7 7-7zm-8 2V5h2v6h1.17L12 13.17 9.83 11H11zm-6 7h14v2H5v-2z"/></svg></span> Télécharger
        </div>
        <div class="menu-item" @click="handleContextAction('rename')">
          <span class="menu-icon"><svg xmlns="http://www.w3.org/2000/svg" height="20px" viewBox="0 0 24 24" width="20px" fill="#5f6368"><path d="M0 0h24v24H0V0z" fill="none"/><path d="M14.06 9.02l.92.92L5.92 19H5v-.92l9.06-9.06M17.66 3c-.25 0-.51.1-.7.29l-1.83 1.83 3.75 3.75 1.83-1.83c.39-.39.39-1.02 0-1.41l-2.34-2.34c-.2-.2-.45-.29-.71-.29zm-3.6 3.19L3 17.25V21h3.75L17.81 9.94l-3.75-3.75z"/></svg></span> Renommer
        </div>
        <div class="menu-item" @click="handleContextAction('move')">
          <span class="menu-icon"><svg xmlns="http://www.w3.org/2000/svg" height="20px" viewBox="0 0 24 24" width="20px" fill="#5f6368"><path d="M0 0h24v24H0V0z" fill="none"/><path d="M20 6h-8l-2-2H4c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2zm-6 12v-3h-4v-4h4V8l5 5-5 5z"/></svg></span> Déplacer
        </div>
        <div class="menu-item" @click="handleContextAction('share')">
          <span class="menu-icon"><svg xmlns="http://www.w3.org/2000/svg" height="20px" viewBox="0 0 24 24" width="20px" fill="#5f6368"><path d="M0 0h24v24H0V0z" fill="none"/><path d="M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z"/></svg></span> Partager
        </div>
        <div class="menu-item" v-if="contextMenu.item && contextMenu.item.shared" @click="handleContextAction('get-share-link')">
          <span class="menu-icon"><svg xmlns="http://www.w3.org/2000/svg" height="20px" viewBox="0 0 24 24" width="20px" fill="#5f6368"><path d="M0 0h24v24H0V0z" fill="none"/><path d="M3.9 12c0-1.71 1.39-3.1 3.1-3.1h4V7H7c-2.76 0-5 2.24-5 5s2.24 5 5 5h4v-1.9H7c-1.71 0-3.1-1.39-3.1-3.1zM8 13h8v-2H8v2zm9-6h-4v1.9h4c1.71 0 3.1 1.39 3.1 3.1s-1.39 3.1-3.1 3.1h-4V17h4c2.76 0 5-2.24 5-5s-2.24-5-5-5z"/></svg></span> Voir le lien de partage
        </div>
        <div class="menu-item" @click="handleContextAction('tags')">
          <span class="menu-icon"><svg xmlns="http://www.w3.org/2000/svg" height="20px" viewBox="0 0 24 24" width="20px" fill="#5f6368"><path d="M0 0h24v24H0V0z" fill="none"/><path d="M17.63 5.84C17.27 5.33 16.67 5 16 5L5 5.01C3.9 5.01 3 5.9 3 7v10c0 1.1.9 1.99 2 1.99L16 19c.67 0 1.27-.33 1.63-.84L22 12l-4.37-6.16zM16 17H5V7h11l3.55 5L16 17z"/></svg></span> Tags
        </div>
        <div class="menu-item delete" @click="handleContextAction('delete')">
          <span class="menu-icon"><svg xmlns="http://www.w3.org/2000/svg" height="20px" viewBox="0 0 24 24" width="20px" fill="#5f6368"><path d="M0 0h24v24H0V0z" fill="none"/><path d="M16 9v10H8V9h8m-1.5-6h-5l-1 1H5v2h14V4h-3.5l-1-1zM18 7H6v12c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7z"/></svg></span> Supprimer
        </div>
      </template>
      <template v-else>
        <div class="menu-item" @click="handleContextAction('add-file')">
          <span class="menu-icon"><svg xmlns="http://www.w3.org/2000/svg" height="20px" viewBox="0 0 24 24" width="20px" fill="#5f6368"><path d="M0 0h24v24H0V0z" fill="none"/><path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm4 18H6V4h7v5h5v11zM8 15.01l1.41 1.41L11 14.83V19h2v-4.17l1.59 1.59L16 15.01 12.01 11 8 15.01z"/></svg></span> Ajouter un fichier
        </div>
        <div class="menu-item" @click="handleContextAction('create-folder')">
          <span class="menu-icon"><svg xmlns="http://www.w3.org/2000/svg" height="20px" viewBox="0 0 24 24" width="20px" fill="#5f6368"><path d="M0 0h24v24H0V0z" fill="none"/><path d="M20 6h-8l-2-2H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2zm0 12H4V6h5.17l2 2H20v10zm-8-4h2v2h2v-2h2v-2h-2v-2h-2v2h-2z"/></svg></span> Créer un dossier
        </div>
      </template>
    </div>

    <InputDialog 
      v-model:isOpen="inputDialog.isOpen"
      :title="inputDialog.title"
      :defaultValue="inputDialog.defaultValue"
      :placeholder="inputDialog.placeholder"
      @confirm="handleInputConfirm"
      @cancel="handleInputCancel"
    />
    <TagDialog 
      v-model:isOpen="tagDialog.isOpen"
      :initialTags="tagDialog.initialTags"
      @confirm="handleTagConfirm"
    />
    <ShareDialog
      :isOpen="shareDialog.isOpen"
      :item="shareDialog.item"
      @close="closeShareDialog"
    />
    <ManageShareDialog
      :isOpen="manageShareDialog.isOpen"
      :item="manageShareDialog.item"
      :initialTab="manageShareDialog.initialTab"
      @close="closeManageShareDialog"
      @share-deleted="onShareDeleted"
      @share-created="onShareCreated"
    />
    <MoveDialog
      v-if="moveDialog.isOpen"
      @close="closeMoveDialog"
      @move-to="onMoveTo"
    />
    <DeleteConfirmDialog
      :isOpen="deleteDialog.isOpen"
      :itemsCount="deleteDialog.itemsCount"
      :itemName="deleteDialog.itemName"
      @confirm="confirmDelete"
      @cancel="closeDeleteDialog"
    />
  </div>
</template>

<script setup>
import { onMounted, ref, computed, onUnmounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useFileStore } from '../../stores/files'
import { useAuthStore } from '../../stores/auth'
import { usePreferencesStore } from '../../stores/preferences'
import { useTagStore } from '../../stores/tags'
import InputDialog from '../InputDialog.vue'
import TagDialog from '../TagDialog.vue'
import ShareDialog from '../ShareDialog.vue'
import api from '../../api'
import MoveDialog from '../MoveDialog.vue';
import ManageShareDialog from '../ManageShareDialog.vue';
import DeleteConfirmDialog from '../DeleteConfirmDialog.vue';
import FileTable from './FileTable.vue';

const router = useRouter()
const authStore = useAuthStore()
const fileStore = useFileStore()
const preferenceStore = usePreferencesStore()
const tagStore = useTagStore()
const selectedItems = ref([])
const lastClickedIndex = ref(-1) // Pour la sélection avec Shift
const fileInput = ref(null)
const isDragging = ref(false)

const currentSortKey = ref('name');
const currentSortDirection = ref('asc');

const handleSortChange = (key) => {
  if (currentSortKey.value === key) {
    currentSortDirection.value = currentSortDirection.value === 'asc' ? 'desc' : 'asc';
  } else {
    currentSortKey.value = key;
    currentSortDirection.value = 'asc';
  }
};

const sortItems = (items) => {
  return [...items].sort((a, b) => {
    let valA, valB;

    switch (currentSortKey.value) {
      case 'name':
        valA = a.Name.toLowerCase();
        valB = b.Name.toLowerCase();
        break;
      case 'size':
        valA = a.Size || 0;
        valB = b.Size || 0;
        break;
      case 'created':
        valA = new Date(a.CreatedAt).getTime();
        valB = new Date(b.CreatedAt).getTime();
        break;
      case 'updated':
        valA = new Date(a.UpdatedAt).getTime();
        valB = new Date(b.UpdatedAt).getTime();
        break;
      default:
        return 0;
    }

    if (valA < valB) return currentSortDirection.value === 'asc' ? -1 : 1;
    if (valA > valB) return currentSortDirection.value === 'asc' ? 1 : -1;
    return 0;
  });
};

const columns = computed(() => {
  const cols = [
    { key: 'icon', label: '', headerClass: 'icon-col', cellClass: 'icon-col' },
    { key: 'name', label: 'Nom', cellClass: 'name-cell' },
  ];

  if (fileStore.searchQuery && fileStore.searchQuery.trim() !== '') {
    cols.push({ key: 'path', label: 'Chemin' });
  }

  cols.push(
    { key: 'tags', label: 'Tags' },
    { key: 'created', label: 'Créé le' },
    { key: 'updated', label: 'Modifié le' },
    { key: 'size', label: 'Taille' }
  );
  return cols;
})

const filteredFolders = computed(() => {
  let folders = fileStore.folders;
  if (fileStore.searchQuery) {
    const query = fileStore.searchQuery.toLowerCase()
    folders = folders.filter(folder => folder.Name.toLowerCase().includes(query))
  }
  return sortItems(folders);
})

const filteredFiles = computed(() => {
  let files = fileStore.files;
  if (fileStore.searchQuery) {
    const query = fileStore.searchQuery.toLowerCase()
    files = files.filter(file => file.Name.toLowerCase().includes(query))
  }
  return sortItems(files);
})

const allItems = computed(() => {
  return [
    ...filteredFolders.value.map(item => ({ ...item, type: 'folder' })),
    ...filteredFiles.value.map(item => ({ ...item, type: 'file' }))
  ]
})

const inputDialog = ref({
  isOpen: false,
  title: '',
  defaultValue: '',
  placeholder: '',
  resolve: null
})

const tagDialog = ref({
  isOpen: false,
  initialTags: [],
  resolve: null
})

const shareDialog = ref({
  isOpen: false,
  item: null
})

const manageShareDialog = ref({
  isOpen: false,
  item: null,
  initialTab: 'link'
});

const deleteDialog = ref({
  isOpen: false,
  itemsCount: 0,
  itemName: ''
});

const moveDialog = ref({
  isOpen: false,
});

const openManageShareDialog = (item, type = 'file', initialTab = 'link') => {
  manageShareDialog.value = {
    isOpen: true,
    item: { ...item, type: type },
    initialTab: initialTab
  };
};

const closeManageShareDialog = () => {
  manageShareDialog.value.isOpen = false;
  manageShareDialog.value.item = null;
};

const onShareCreated = () => {
  fileStore.fetchItems(fileStore.currentPath)
}

const onShareDeleted = () => {
  fileStore.fetchItems(fileStore.currentPath)
  closeManageShareDialog()
}

const openMoveDialog = () => {
  moveDialog.value.isOpen = true
}

const closeMoveDialog = () => {
  moveDialog.value.isOpen = false
}

const onMoveTo = (destinationPath) => {
  fileStore.moveItems(selectedItems.value, destinationPath)
  closeMoveDialog()
}

const onFileUploaded = () => {
  fileStore.fetchItems(fileStore.currentPath)
}



const openShareDialog = (item) => {
  openManageShareDialog(item, item.type);
};

const closeShareDialog = () => {
  shareDialog.value.isOpen = false;
  shareDialog.value.item = null;
};

const openInputDialog = (title, defaultValue = '', placeholder = '') => {
  return new Promise((resolve) => {
    inputDialog.value = {
      isOpen: true,
      title,
      defaultValue,
      placeholder,
      resolve
    }
  })
}

const handleInputConfirm = (value) => {
  if (inputDialog.value.resolve) {
    inputDialog.value.resolve(value)
  }
  inputDialog.value.resolve = null
}

const handleInputCancel = () => {
  if (inputDialog.value.resolve) {
    inputDialog.value.resolve(null)
  }
  inputDialog.value.resolve = null
}

const openTagDialog = (initialTags) => {
  return new Promise((resolve) => {
    tagDialog.value = {
      isOpen: true,
      initialTags,
      resolve
    }
  })
}

const handleTagConfirm = (tags) => {
  if (tagDialog.value.resolve) {
    tagDialog.value.resolve(tags)
  }
  tagDialog.value.resolve = null
}



const contextMenu = ref({
  visible: false,
  x: 0,
  y: 0,
  item: null
})

const closeContextMenu = () => {
  contextMenu.value.visible = false
}

const pathSegments = computed(() => {
  if (fileStore.viewMode === 'shared') {
      const segments = [{ name: 'Partagés avec moi', path: 'SHARE_ROOT' }];
      fileStore.sharedBreadcrumbs.forEach((crumb, index) => {
          segments.push({
              name: crumb.name,
              path: index, // Use index as identifier for navigation
              isShared: true
          });
      });
      return segments;
  }

  const path = fileStore.currentPath
  const segments = [{ name: 'Mon Drive', path: '/' }]
  
  if (path === '/') return segments

  const parts = path.split('/').filter(p => p)
  let currentBuild = ''
  
  parts.forEach(part => {
    currentBuild += '/' + part
    segments.push({ name: part, path: currentBuild })
  })
  
  return segments
})

const navigateToPath = (path) => {
  if (fileStore.viewMode === 'shared') {
       if (path === 'SHARE_ROOT') {
           fileStore.viewMode = 'drive';
           fileStore.fetchItems('/');
           return;
       }
       if (typeof path === 'number') {
           fileStore.navigateSharedTo(path);
       }
       return;
   }

  if (path === fileStore.currentPath) return
  selectedItems.value = []
  fileStore.fetchItems(path)
}

onMounted(() => {
  // Only fetch root if we are NOT in shared mode.
  // This prevents resetting viewMode when coming from HomeView -> FileShared
  if (fileStore.viewMode !== 'shared') {
    fileStore.fetchItems('/')
  }
  tagStore.fetchTags()
  document.addEventListener('click', closeContextMenu)
})

watch(() => fileStore.currentPath, () => {
  selectedItems.value = []
})

onUnmounted(() => {
  document.removeEventListener('click', closeContextMenu)
})

const openBackgroundContextMenu = (event) => {
  if (!preferenceStore.enableContextMenu) return;
  
  // Deselect items when clicking on background
  selectedItems.value = []

  const menuWidth = 150;
  const menuHeight = 100;
  let x = event.clientX;
  let y = event.clientY;

  if (x + menuWidth > window.innerWidth) {
    x -= menuWidth;
  }
  if (y + menuHeight > window.innerHeight) {
    y -= menuHeight;
  }

  contextMenu.value = {
    visible: true,
    x: x,
    y: y,
    item: null
  }
}

const openContextMenu = (event, item, type) => {
  if (!preferenceStore.enableContextMenu) return;
  if(!isSelected(item, type)) {
    selectedItems.value = [{...item, type}]
  }
  const menuWidth = 150; // Approximate width of context menu
  const menuHeight = 100; // Approximate height of context menu
  let x = event.clientX;
  let y = event.clientY;

  if (x + menuWidth > window.innerWidth) {
    x -= menuWidth;
  }
  if (y + menuHeight > window.innerHeight) {
    y -= menuHeight;
  }

  contextMenu.value = {
    visible: true,
    x: x,
    y: y,
    item: { ...item, type }
  }
}

const handleContextAction = (action) => {
  const item = contextMenu.value.item
  
  if (action === 'add-file') {
    triggerFileInput()
    closeContextMenu()
    return
  }
  if (action === 'create-folder') {
    createNewFolder()
    closeContextMenu()
    return
  }

  if (!item) return;

  switch (action){
    case 'download':
      if (item.type === 'file') {
        fileStore.downloadFile(item.ID, item.Name, item.MimeType)
      }
      break
    case 'rename':
      renameSelectedItem()
      break
    case 'move':
      openMoveDialog()
      break
    case 'share':
      openShareDialog(item)
      break
    case 'direct-share':
      openManageShareDialog(item, item.type, 'friends')
      break
    case 'get-share-link':
      if (item.type === 'file' && item.share_token) {
        const shareUrl = `${window.location.origin}/s/${item.share_token}`;
        navigator.clipboard.writeText(shareUrl).then(() => {
          alert('Lien de partage copié dans le presse-papiers !');
        }).catch(err => {
          alert('Impossible de copier le lien.');
          console.error('Could not copy text: ', err);
        });
      }
      break
    case 'tags':
      updateTags()
      break
    case 'delete':
      deleteSelectedItems()
      break
  }
  closeContextMenu()
}

const selectItem = (item, type, event) => {
  const currentIndex = allItems.value.findIndex(i => i.ID === item.ID && i.type === type);
  const itemWithType = { ...item, type };

  if (event.shiftKey && lastClickedIndex.value !== -1) {
    const start = Math.min(lastClickedIndex.value, currentIndex);
    const end = Math.max(lastClickedIndex.value, currentIndex);
    const rangeToSelect = allItems.value.slice(start, end + 1);
    
    selectedItems.value = rangeToSelect;

  } else if (event.ctrlKey || event.metaKey) {
    const isItemSelected = isSelected(item, type);
    if (isItemSelected) {
      selectedItems.value = selectedItems.value.filter(i => !(i.ID === item.ID && i.type === type));
    } else {
      selectedItems.value.push(itemWithType);
    }
    lastClickedIndex.value = currentIndex;
  } else {
    if (isSelected(item, type) && selectedItems.value.length === 1) {
      selectedItems.value = [];
      lastClickedIndex.value = -1;
    } else {
      selectedItems.value = [itemWithType];
      lastClickedIndex.value = currentIndex;
    }
  }
}

// Helper pour vérifier si un item est sélectionné (utile pour le template)
const isSelected = (item, type) => {
  return selectedItems.value.some(i => i.ID === item.ID && i.type === type);
}

const openFolder = (folder) => {
  const folderName = folder.Name || folder.name;
  
  if (fileStore.viewMode === 'shared') {
      fileStore.navigateShared(folder.ID, folderName);
      selectedItems.value = [];
      return;
  }

  // Add to history
  const fullPath = fileStore.currentPath === '/' ? '/' + folderName : fileStore.currentPath + '/' + folderName;
  fileStore.addToHistory({ 
      name: folderName, 
      path: fullPath, 
      type: 'folder' 
  });

  selectedItems.value = [] // Deselect items when navigating
  fileStore.navigateTo(folderName)
}

const goUp = () => {
  if (fileStore.viewMode === 'shared') {
      fileStore.navigateUp(); // Should handle shared logic
      selectedItems.value = [];
      return;
  }
  if (fileStore.currentPath !== '/') {
    selectedItems.value = [] // Deselect items when navigating up
    fileStore.navigateUp()
  }
}

const downloadSelectedFiles = () => {
  const files = selectedItems.value.filter(i => i.type === 'file');
  if (files.length === 0) return;

  if (files.length === 1) {
    const file = files[0];
    fileStore.downloadFile(file.ID, file.Name);
  } else {
    // Logic for downloading multiple files, e.g., zipping them first
    alert("Le téléchargement de plusieurs fichiers en une fois (ex: zip) n'est pas encore implémenté. Les fichiers seront téléchargés individuellement.");
    files.forEach(file => {
      fileStore.downloadFile(file.ID, file.Name);
    });
  }
}

const deleteSelectedItems = () => {
  if (selectedItems.value.length === 0) return;

  deleteDialog.value = {
    isOpen: true,
    itemsCount: selectedItems.value.length,
    itemName: selectedItems.value.length === 1 ? selectedItems.value[0].Name : ''
  };
};

const closeDeleteDialog = () => {
  deleteDialog.value.isOpen = false;
};

const confirmDelete = async () => {
  closeDeleteDialog();
  
  const fileIDs = selectedItems.value.filter(i => i.type === 'file').map(i => i.ID);
  const folderIDs = selectedItems.value.filter(i => i.type === 'folder').map(i => i.ID);
  
  if (fileIDs.length > 0) {
      await fileStore.deleteFiles(fileIDs);
  }
  
  // Delete folders one by one for now as bulk delete folders is not implemented
  for (const folderID of folderIDs) {
      await api.delete(`/files/folder/${folderID}`);
  }
  
  // Refresh list if we deleted folders manually (deleteFiles already refreshes)
  if (folderIDs.length > 0 && fileIDs.length === 0) {
      fileStore.fetchItems(fileStore.currentPath);
  }
  
  selectedItems.value = [] // Clear selection after deletion
};

const renameSelectedItem = async () => {
  if (selectedItems.value.length !== 1) return;
  
  const item = selectedItems.value[0];
  const newName = await openInputDialog("Entrez le nouveau nom :", item.Name);
  
  if (newName && newName !== item.Name) {
    try {
      await fileStore.renameItem(item.ID, item.type, newName);
      selectedItems.value = []; // Clear selection
    } catch (error) {
      alert("Erreur lors du renommage : " + (error.response?.data?.error || error.message));
    }
  }
}

const removeTag = async (item, type, tagToRemove) => {
  const currentTags = item.Tags || [];
  const newTags = currentTags.filter(t => t !== tagToRemove);
  try {
    await fileStore.updateTags(item.ID, type, newTags);
  } catch (error) {
    console.error(error);
    alert("Erreur lors de la suppression du tag.");
  }
}

const updateTags = async () => {
  if (selectedItems.value.length !== 1) return;
  
  const item = selectedItems.value[0];
  const currentTags = item.Tags || [];
  
  const newTags = await openTagDialog(currentTags);
  
  if (newTags !== null) {
    try {
      await fileStore.updateTags(item.ID, item.type, newTags);
      selectedItems.value = []; // Clear selection
    } catch (error) {
      alert("Erreur lors de la mise à jour des tags : " + (error.response?.data?.error || error.message));
    }
  }
}

const downloadFile = (file) => {
  fileStore.addToHistory({ ...file, type: 'file' });
  fileStore.downloadFile(file.ID, file.Name, file.MimeType);
}

const triggerFileInput = () => {
  fileInput.value.click()
}

const handleFileUpload = async (event) => {
  const file = event.target.files[0]
  if (file) {
    await fileStore.uploadFile(file)
    event.target.value = '' // Reset file input
  }
}

const createNewFolder = async () => {
  const folderName = await openInputDialog("Entrez le nom du nouveau dossier :")
  if (folderName) {
    await fileStore.createFolder(folderName)
  }
}



const closeUploadPopup = () => {
  fileStore.isUploading = false
}

const onDragOver = (e) => {
  // Only show overlay if dragging files from OS
  if (e.dataTransfer.types.includes('Files')) {
    isDragging.value = true
  }
}

const onDragLeave = (e) => {
  // Empêche le clignotement quand on passe sur les enfants
  if (e.currentTarget.contains(e.relatedTarget)) return;
  isDragging.value = false
}

const onDrop = async (e) => {
  isDragging.value = false
  const files = e.dataTransfer.files
  if (files.length > 0) {
    // On upload les fichiers un par un
    for (const file of files) {
      await fileStore.uploadFile(file)
    }
  }
}

const onDragStart = (item, type, event) => {
  event.dataTransfer.effectAllowed = 'move'
  event.dataTransfer.dropEffect = 'move'
  
  let itemsToDrag = []
  
  // Check if the dragged item is in the selection
  const isSelected = selectedItems.value.some(i => i.ID === item.ID && i.type === type)
  
  if (isSelected && selectedItems.value.length > 0) {
      itemsToDrag = selectedItems.value.map(i => ({ id: i.ID, type: i.type }))
  } else {
      itemsToDrag = [{ id: item.ID, type: type }]
  }

  event.dataTransfer.setData('application/json', JSON.stringify({ items: itemsToDrag }))
}

const onFolderDragOver = (event) => {
  event.currentTarget.classList.add('drag-over-target')
}

const onFolderDragLeave = (event) => {
  event.currentTarget.classList.remove('drag-over-target')
}

const onDropOnFolder = async (targetFolder, event) => {
  event.currentTarget.classList.remove('drag-over-target')
  isDragging.value = false
  const data = event.dataTransfer.getData('application/json')
  if (!data) return 

  try {
    const parsed = JSON.parse(data)
    const items = parsed.items || [parsed] // Handle potential backward compatibility or single item structure
    
    // Filter out the target folder itself if it's being dragged (cannot move folder into itself)
    const validItems = items.filter(item => !(item.id === targetFolder.ID && item.type === 'folder'))
    
    if (validItems.length > 0) {
        await fileStore.moveItems(validItems, targetFolder.Path)
    }
  } catch (e) {
    console.error("Invalid drag data", e)
  }
}

const onDropOnParent = async (event) => {
  event.currentTarget.classList.remove('drag-over-target')
  isDragging.value = false
  if (fileStore.currentPath === '/') return;

  const data = event.dataTransfer.getData('application/json')
  if (!data) return

  try {
    const parsed = JSON.parse(data)
    const items = parsed.items || [parsed]
    
    const parts = fileStore.currentPath.split('/').filter(p => p)
    parts.pop()
    const parentPath = parts.length > 0 ? '/' + parts.join('/') : '/'
    
    await fileStore.moveItems(items, parentPath)
  } catch (e) {
    console.error("Invalid drag data", e)
  }
}
</script>

<style scoped>
.file-list-container {
  margin-top: 0.8rem;
  position: relative;
  background-color: var(--card-color);
  height: 100%;
  width: 100%;
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
}

.path-banner {
  padding: 0.5rem 1rem;
  background-color: var(--card-color);
  display: flex;
  justify-content: flex-start;
  align-items: center;
  gap: 1rem;
}

.back-arrow {
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  padding: 0;
  color: #555;
  border: 1px solid #ccc;
  border-radius: 50%;
  background-color: white;
  transition: all 0.2s;
}

.back-arrow:not(.disabled):hover {
  background-color: #f0f0f0;
  border-color: #bbb;
  color: #333;
}

.back-arrow.disabled {
  color: #ccc;
  cursor: not-allowed;
  background-color: #f9f9f9;
  border-color: #eee;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.dashboard-title {
  margin: 0;
  font-size: 1.5rem;
  text-align: left;
}

.dashboard-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
  width: 80vw; /* Match file-list-container width */
}

.header-actions {
  display: flex;
  gap: 1rem;
}

.btn-header {
  background-color: white;
  border: 1px solid #ccc;
  color: #333;
  padding: 0.5rem 1rem;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 500;
  transition: background-color 0.2s;
}

.btn-header:hover {
  background-color: #f0f0f0;
}

.btn-logout {
  color: #dc3545;
  border-color: #dc3545;
}

.btn-logout:hover {
  background-color: #dc3545;
  color: white;
}

.list-area {
  margin-top: 1rem;
  overflow-y: auto;
  flex-grow: 1;
  padding: 0 1rem;
}


.size {
  color: #5c5c5c;
  font-size: 0.9em;
  text-align: right;
}

.list-item.selected {
  background-color: rgba(66, 185, 131, 0.2); /* Light green selection */
}

button {
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-weight: bold;
}

.btn-add-file {
  background-color: var(--primary-color);
  color: white;
}

.btn-rename {
  background-color: #ffc107;
  color: #333;
  margin-right: 0.5rem;
}

.btn-rename:disabled {
  background-color: #e0e0e0;
  color: #999;
  cursor: not-allowed;
}

.btn-download {
  background-color: var(--primary-color);
  color: white;
}

.path-banner button {
    background-color: var(--background-color);
    border: 1px solid #ccc;
}

.path-banner button:disabled {
    cursor: not-allowed;
    opacity: 0.5;
}

.breadcrumbs {
  display: flex;
  align-items: center;
  font-size: 1.5rem;
  transform: translateY(-2px);
}

.breadcrumb-segment {
  display: flex;
  align-items: center;
}

.breadcrumb-link {
  cursor: pointer;
  color: var(--primary-color, #42b983);
  text-decoration: none;
  padding: 0.2rem 0.5rem;
  padding-top: 0.2rem;
  border-radius: 4px;
  transition: all 0.2s ease;
}

.breadcrumb-link:hover {
  text-decoration: underline;
  background-color: rgba(66, 185, 131, 0.1);
  transform: translateY(-1px);
}

.breadcrumb-link.current {
  color: var(--text-color);
  cursor: default;
  font-weight: bold;
  text-decoration: none;
}

.separator {
  margin: 0 0.5rem;
  color: #999;
}

.progress-container {
  padding: 0.5rem 1rem;
  display: flex;
  align-items: center;
  gap: 1rem;
  background-color: #f9f9f9;
  border-bottom: 1px solid #eee;
}

.progress-bar {
  flex-grow: 1;
  height: 10px;
  background-color: #e0e0e0;
  border-radius: 5px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background-color: var(--primary-color, #42b983);
  transition: width 0.3s ease;
}

.progress-text {
  font-size: 0.9rem;
  font-weight: bold;
  color: #555;
  min-width: 3rem;
  text-align: right;
}

.upload-popup {
  position: fixed;
  top: 20px;
  right: 20px;
  width: 320px;
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
  z-index: 1000;
  border: 1px solid #eee;
  overflow: hidden;
  animation: slideIn 0.3s ease;
}

@keyframes slideIn {
  from { transform: translateX(100%); opacity: 0; }
  to { transform: translateX(0); opacity: 1; }
}

.popup-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem 1rem;
  background-color: #f5f5f5;
  border-bottom: 1px solid #eee;
}

.popup-title {
  font-weight: bold;
  font-size: 0.9rem;
  color: #333;
}

.btn-close {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  padding: 0;
  line-height: 0.8;
  color: #999;
  width: auto;
  height: auto;
}

.btn-close:hover {
  color: #333;
  background: none;
}

.popup-content {
  padding: 1rem;
}

.file-name {
  margin-bottom: 0.8rem;
  font-size: 0.9rem;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  color: #333;
  text-align: left;
}

.progress-container-popup {
  display: flex;
  align-items: center;
  gap: 0.8rem;
}

.drag-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(255, 255, 255, 0.9);
  border: 3px dashed var(--primary-color);
  border-radius: 8px;
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 500;
  pointer-events: none; /* Permet de drop "au travers" de l'overlay */
}

.drag-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  color: var(--primary-color);
}

.drag-icon {
  font-size: 4rem;
  margin-bottom: 1rem;
}

.drag-text {
  font-size: 1.5rem;
  font-weight: bold;
}



.context-menu {
  position: fixed;
  background: white;
  border: 1px solid #ccc;
  border-radius: 4px;
  box-shadow: 0 2px 10px rgba(0,0,0,0.2);
  z-index: 1000;
  min-width: 150px;
  padding: 5px 0;
}

.menu-item {
  padding: 8px 15px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.9rem;
  color: #333;
  transition: background-color 0.2s;
  text-align: left;
}

.menu-item:hover {
  background-color: #f0f0f0;
}

.menu-item.delete {
  color: #dc3545;
  border-top: 1px solid #eee;
}

.menu-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
}

.menu-item.delete .menu-icon svg {
  fill: #dc3545;
}

.date-column {
  font-size: 0.9em;
  color: #666;
}
</style>
