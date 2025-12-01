<template>
  <div class="left-bar">
    <div class="action-section">
      <button class="btn-new" @click.stop="toggleNewMenu">
        <span class="plus-icon">+</span>
        <span>Nouveau</span>
      </button>
      <div v-if="showNewMenu" class="new-menu-dropdown" @click.stop>
        <div class="dropdown-item" @click="triggerUpload">
          <span class="icon">📄</span> Fichier
        </div>
        <div class="dropdown-item" @click="triggerCreateFolder">
          <span class="icon">📁</span> Dossier
        </div>
      </div>
    </div>

    <div class="menu-section">
      <div class="menu-item active">
        <span class="icon">📁</span>
        <span>Mes fichiers</span>
      </div>
      <div class="menu-item">
        <span class="icon">🔗</span>
        <span>Fichiers partagés</span>
      </div>
      <div class="menu-item">
        <span class="icon">🔄</span>
        <span>Peer-to-Peer</span>
      </div>
    </div>

    <div class="storage-section">
      <div class="storage-info">
        <span class="storage-label">Stockage</span>
        <span class="storage-value">{{ formattedStorageUsed }} / {{ formattedStorageLimit }}</span>
      </div>
      <div class="storage-bar">
        <div class="storage-fill" :style="{ width: storagePercentage + '%' }"></div>
      </div>
    </div>

    <input type="file" ref="fileInput" @change="handleFileUpload" style="display: none" />
    <InputDialog 
      v-model:isOpen="inputDialog.isOpen"
      :title="inputDialog.title"
      :defaultValue="inputDialog.defaultValue"
      :placeholder="inputDialog.placeholder"
      @confirm="handleInputConfirm"
      @cancel="handleInputCancel"
    />
  </div>
</template>

<script setup>
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useAuthStore } from '../../stores/auth'
import { useFileStore } from '../../stores/files'
import InputDialog from '../InputDialog.vue'

const authStore = useAuthStore()
const fileStore = useFileStore()

const showNewMenu = ref(false)
const fileInput = ref(null)
const inputDialog = ref({
  isOpen: false,
  title: '',
  defaultValue: '',
  placeholder: '',
  resolve: null
})

const toggleNewMenu = () => {
  showNewMenu.value = !showNewMenu.value
}

const closeNewMenu = () => {
  showNewMenu.value = false
}

onMounted(() => {
  document.addEventListener('click', closeNewMenu)
})

onUnmounted(() => {
  document.removeEventListener('click', closeNewMenu)
})

const triggerUpload = () => {
  fileInput.value.click()
  showNewMenu.value = false
}

const handleFileUpload = async (event) => {
  const file = event.target.files[0]
  if (file) {
    await fileStore.uploadFile(file)
    event.target.value = ''
  }
}

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

const triggerCreateFolder = async () => {
  showNewMenu.value = false
  const folderName = await openInputDialog("Entrez le nom du nouveau dossier :")
  if (folderName) {
    await fileStore.createFolder(folderName)
  }
}

const formatSize = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formattedStorageUsed = computed(() => {
  return formatSize(authStore.user?.storage_used || 0)
})

const formattedStorageLimit = computed(() => {
  return formatSize(authStore.user?.storage_limit || 10737418240) // Default 10GB
})

const storagePercentage = computed(() => {
  const used = authStore.user?.storage_used || 0
  const limit = authStore.user?.storage_limit || 10737418240
  return Math.min((used / limit) * 100, 100)
})
</script>

<style scoped>
.left-bar {
  width: 250px;
  background-color: var(--background-color);
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 1rem;
  padding-top: 0;
  height: 100%;
  box-sizing: border-box;
}

.menu-section {
  padding-top: 2rem;
  flex-grow: 1;
  overflow-y: auto;
}

.menu-item {
  display: flex;
  align-items: center;
  padding: 0.8rem 1rem;
  cursor: pointer;
  border-radius: 6px;
  margin-bottom: 0.5rem;
  color: #555;
  transition: background-color 0.2s;
}

.menu-item:hover {
  background-color: #e9ecef;
}

.menu-item.active {
  background-color: #e3f2fd;
  color: #0d6efd;
  font-weight: 500;
}

.icon {
  margin-right: 10px;
  font-size: 1.2rem;
}

.storage-section {
  background-color: var(--card-color);
  padding: 1rem;
  border-radius: 8px;
  border: 1px solid var(--border-color);
  box-shadow: 0 2px 4px rgba(0,0,0,0.05);
}

.storage-info {
  display: flex;
  justify-content: space-between;
  font-size: 0.85rem;
  margin-bottom: 0.5rem;
  color: #666;
}

.storage-bar {
  height: 6px;
  background-color: #e9ecef;
  border-radius: 3px;
  overflow: hidden;
}

.storage-fill {
  height: 100%;
  background-color: #42b983;
  border-radius: 3px;
}

.action-section {
  margin-bottom: 1rem;
  position: relative;
}

.btn-new {
  width: 100%;
  height: 120%;
  padding: 0.8rem;
  background-color: white;
  border: 1px solid #ddd;
  border-radius: 24px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  font-weight: bold;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
  transition: all 0.2s;
}

.btn-new:hover {
  background-color: #f8f9fa;
  box-shadow: 0 2px 5px rgba(0,0,0,0.15);
}

.plus-icon {
  font-size: 1.2rem;
  color: var(--primary-color, #42b983);
}

.new-menu-dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  width: 100%;
  background: white;
  border: 1px solid #ddd;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
  z-index: 100;
  margin-top: 5px;
  overflow: hidden;
}

.dropdown-item {
  padding: 10px 15px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 10px;
  transition: background 0.2s;
}

.dropdown-item:hover {
  background-color: #f0f0f0;
}
</style>
