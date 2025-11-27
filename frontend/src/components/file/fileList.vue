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
        <input type="file" ref="fileInput" @change="handleFileUpload" style="display: none" />
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
    <div class="path-banner">
      <span @click="goUp" class="back-arrow" :class="{ 'disabled': fileStore.currentPath === '/' }"
            @drop.stop="onDropOnParent"
            @dragover.prevent="onFolderDragOver"
            @dragleave="onFolderDragLeave">←</span>
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

    <div class="list-header">
      <span class="header-icon"></span>
      <span class="header-name">Nom</span>
      <span class="header-tags">Tags</span>
      <span class="header-date">Créé le</span>
      <span class="header-date">Modifié le</span>
      <span class="header-size">Taille</span>
    </div>

    <div class="list-area">
      <!-- Folders -->
      <div v-for="folder in filteredFolders" :key="folder.ID" 
           class="list-item folder-item" 
           :class="{ selected: isSelected(folder, 'folder') }"
           @click="selectItem(folder, 'folder', $event)"
           @dblclick="openFolder(folder.Name)"
           @contextmenu.prevent="openContextMenu($event, folder, 'folder')"
           draggable="true"
           @dragstart="onDragStart(folder, 'folder', $event)"
           @drop.stop="onDropOnFolder(folder, $event)"
           @dragover.prevent="onFolderDragOver"
           @dragleave="onFolderDragLeave">
        <span class="icon">📁</span>
        <span class="name">{{ folder.Name }}</span>
        <span class="tags-column">
          <span v-if="folder.Tags && folder.Tags.length" class="tags-container">
            <span v-for="tag in folder.Tags" :key="tag" class="tag-badge" :style="getTagStyle(tag)">
              {{ tag }}
              <span class="remove-tag" @click.stop="removeTag(folder, 'folder', tag)">×</span>
            </span>
          </span>
        </span>
        <span class="date-column"></span>
        <span class="date-column">{{ formatDate(folder.CreatedAt) }}</span>
        <span class="size">-</span>
      </div>
      <!-- Files -->
      <div v-for="file in filteredFiles" :key="file.ID" 
          class="list-item"
          :class="{ selected: isSelected(file, 'file') }"
          @click="selectItem(file, 'file', $event)"
          @dblclick="downloadFile(file)"
          @contextmenu.prevent="openContextMenu($event, file, 'file')"
      >
        <span class="icon">📄</span>
        <span class="name">{{ file.Name }}</span>
        <span class="tags-column">
          <span v-if="file.Tags && file.Tags.length" class="tags-container">
            <span v-for="tag in file.Tags" :key="tag" class="tag-badge" :style="getTagStyle(tag)">
              {{ tag }}
              <span class="remove-tag" @click.stop="removeTag(file, 'file', tag)">×</span>
            </span>
          </span>
        </span>
        <span class="date-column">{{ formatDate(file.CreatedAt) }}</span>
        <span class="date-column">{{ formatDate(file.UpdatedAt) }}</span>
        <span class="size">{{ formatSize(file.Size) }}</span>
      </div>
    </div>

    <div v-if="contextMenu.visible" 
         class="context-menu" 
         :style="{ top: contextMenu.y + 'px', left: contextMenu.x + 'px' }">
      <div class="menu-item" @click="handleContextAction('download')" v-if="contextMenu.item.type === 'file'">
        📥 Télécharger
      </div>
      <div class="menu-item" @click="handleContextAction('rename')">
        ✏️ Renommer
      </div>
      <div class="menu-item" @click="handleContextAction('tags')">
        🏷️ Tags
      </div>
      <div class="menu-item delete" @click="handleContextAction('delete')">
        🗑️ Supprimer
      </div>
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
  </div>
</template>

<script setup>
import { onMounted, ref, computed, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useFileStore } from '../../stores/files'
import { useAuthStore } from '../../stores/auth'
import { usePreferencesStore } from '../../stores/preferences'
import { useTagStore } from '../../stores/tags'
import InputDialog from '../InputDialog.vue'
import TagDialog from '../TagDialog.vue'

const router = useRouter()
const authStore = useAuthStore()
const fileStore = useFileStore()
const preferenceStore = usePreferencesStore()
const tagStore = useTagStore()
const selectedItems = ref([])
const fileInput = ref(null)
const isDragging = ref(false)

const filteredFolders = computed(() => {
  if (!fileStore.searchQuery) return fileStore.folders
  const query = fileStore.searchQuery.toLowerCase()
  return fileStore.folders.filter(folder => folder.Name.toLowerCase().includes(query))
})

const filteredFiles = computed(() => {
  if (!fileStore.searchQuery) return fileStore.files
  const query = fileStore.searchQuery.toLowerCase()
  return fileStore.files.filter(file => file.Name.toLowerCase().includes(query))
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

const getTagStyle = (tagName) => {
    const tag = tagStore.tags.find(t => t.name === tagName)
    if (tag) {
        return {
            backgroundColor: tag.color,
            color: getContrastColor(tag.color),
            borderColor: tag.color
        }
    }
    return {}
}

const getContrastColor = (hexcolor) => {
    if (!hexcolor || hexcolor[0] !== '#') return 'black';
    var r = parseInt(hexcolor.substr(1,2),16);
    var g = parseInt(hexcolor.substr(3,2),16);
    var b = parseInt(hexcolor.substr(5,2),16);
    var yiq = ((r*299)+(g*587)+(b*114))/1000;
    return (yiq >= 128) ? 'black' : 'white';
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
  const path = fileStore.currentPath
  const segments = [{ name: 'Home', path: '/' }]
  
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
  if (path === fileStore.currentPath) return
  selectedItems.value = []
  fileStore.fetchItems(path)
}

onMounted(() => {
  fileStore.fetchItems('/')
  tagStore.fetchTags()
  document.addEventListener('click', closeContextMenu)
})

onUnmounted(() => {
  document.removeEventListener('click', closeContextMenu)
})

const openContextMenu = (event, item, type) => {
  if (!preferenceStore.enableContextMenu) return;
  if(!isSelected(item, type)) {
    selectItem.value = [{...item, type}]
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
  const isSelected = selectedItems.value.some(i => i.ID === item.ID && i.type === type);
  const itemWithType = { ...item, type };
  
  if (!event.ctrlKey && !event.metaKey) { // si ctrl ou cmd n'est pas enfoncé
    selectedItems.value = isSelected ? [] : [itemWithType]; // Select only this item
  } else { // si ctrl ou cmd est enfoncé
    if (isSelected) {
      selectedItems.value = selectedItems.value.filter(i => !(i.ID === item.ID && i.type === type)); // Deselect item
    } else {
      selectedItems.value.push(itemWithType); // Add to selection
    }
  }
}

// Helper pour vérifier si un item est sélectionné (utile pour le template)
const isSelected = (item, type) => {
  return selectedItems.value.some(i => i.ID === item.ID && i.type === type);
}

const openFolder = (folderName) => {
  selectedItems.value = [] // Deselect items when navigating
  fileStore.navigateTo(folderName)
}

const goUp = () => {
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

const deleteSelectedItems = async () => {
  if (selectedItems.value.length === 0) return;

  const confirmDelete = confirm(`Êtes-vous sûr de vouloir supprimer les ${selectedItems.value.length} élément(s) sélectionné(s) ?`);
  if (confirmDelete) {
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
  }
}

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

const formatSize = (bytes) => {
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB']
  const k = 1024
  if (bytes === 0) return '0 Byte'
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatDate = (dateString) => {
  if (!dateString) return '-'
  const date = new Date(dateString)
  return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
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
  border-bottom: 1px solid #ccc;
  display: flex;
  justify-content: flex-start;
  align-items: center;
  gap: 1rem;
}

.back-arrow {
  cursor: pointer;
  font-size: 1.5rem;
  font-weight: bold;
  padding: 0.2rem 0.8rem;
  color: #333;
  border: 1px solid #ccc;
  border-radius: 5px;
  background-color: #f0f0f0;
  transition: background-color 0.2s;
  line-height: 1;
}

.back-arrow:not(.disabled):hover {
  background-color: var(--hover-background-color);
}

.back-arrow.disabled {
  color: #ccc;
  cursor: not-allowed;
  background-color: var(--background-color);
  border-color: #eee;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  border-bottom: 1px solid #eee;
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

.list-header {
  display: grid;
  grid-template-columns: 40px 2fr 1fr 150px 150px 100px;
  padding: 0.5rem;
  font-weight: bold;
  border-bottom: 1px solid #ccc;
  background-color: var(--card-color);
  align-items: center;
}

.list-item {
  display: grid;
  grid-template-columns: 40px 2fr 1fr 150px 150px 100px;
  align-items: center;
  padding: 0.5rem;
  cursor: pointer;
  border-radius: 4px;
  transition: background-color 0.2s;
  user-select: none;
  border-bottom: 1px solid #ccc;
}

.list-item:hover {
  background-color: var(--hover-background-color);
}

.list-item .icon {
  margin-right: 0.5rem;
  display: flex;
  justify-content: center;
}
.name {
  text-align: left;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  padding-right: 1rem;
}
.size {
  color: #5c5c5c;
  font-size: 0.9em;
  text-align: right;
}

.list-item.selected {
  background-color: #42b983;
  color: white;
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
  border-radius: 4px;
  transition: all 0.2s ease;
}

.breadcrumb-link:hover {
  text-decoration: underline;
  background-color: rgba(66, 185, 131, 0.1);
  transform: translateY(-1px);
}

.breadcrumb-link.current {
  color: #333;
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

.drag-over-target {
  background-color: rgba(66, 185, 131, 0.2) !important;
  border: 2px dashed var(--primary-color, #42b983);
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

.tag-badge {
  background-color: #e0e0e0;
  color: #333;
  font-size: 0.75rem;
  padding: 2px 6px;
  border-radius: 10px;
  border: 1px solid #ccc;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.remove-tag {
  cursor: pointer;
  color: #666;
  font-weight: bold;
  line-height: 1;
  display: inline-block;
}

.remove-tag:hover {
  color: #dc3545;
}

.tags-column {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.date-column {
  font-size: 0.9em;
  color: #666;
}
</style>
