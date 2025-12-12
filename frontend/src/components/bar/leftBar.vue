<template>
  <div class="left-bar">
    <div class="action-section">
      <button class="btn-new" @click.stop="toggleNewMenu">
        <svg class="plus-icon" width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z" fill="#ea4335"/>
        </svg>
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
      <div class="menu-item" :class="{ active: isActive('/dashboard') }" @click="navigateTo('/dashboard')">
        <span class="icon">📁</span>
        <span>Mes fichiers</span>
      </div>
      <div class="menu-item" :class="{ active: isActive('/dashboard/shares') }" @click="navigateTo('/dashboard/shares')">
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
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../../stores/auth'
import { useFileStore } from '../../stores/files'
import InputDialog from '../InputDialog.vue'

const authStore = useAuthStore()
const fileStore = useFileStore()
const router = useRouter()
const route = useRoute()

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

const navigateTo = (path) => {
  router.push(path)
}

const isActive = (path) => {
  if (path === '/dashboard') {
    return (route.path === '/' || route.path.startsWith('/dashboard')) && !route.path.startsWith('/dashboard/shares')
  }
  return route.path.startsWith(path)
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
  width: 256px;
  background-color: var(--background-color);
  display: flex;
  flex-direction: column;
  padding: 8px 16px;
  height: 100%;
  box-sizing: border-box;
  font-family: 'Roboto', 'Segoe UI', sans-serif;
}

.action-section {
  padding: 8px 0 16px 0;
  position: relative;
}

.btn-new {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  background-color: white;
  border: none;
  border-radius: 16px;
  padding: 0 16px;
  height: 56px;
  box-shadow: 0 1px 2px 0 rgba(60,64,67,0.3), 0 1px 3px 1px rgba(60,64,67,0.15);
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 18px;
  font-weight: 500;
  color: #3c4043;
  width: 100%;
}

.btn-new:hover {
  box-shadow: 0 4px 8px 3px rgba(60,64,67,0.15);
  background-color: #fafafa;
}

.plus-icon {
  min-width: 24px;
}

.new-menu-dropdown {
  position: absolute;
  top: 60px;
  left: 0;
  background: white;
  border-radius: 4px;
  box-shadow: 0 2px 10px rgba(0,0,0,0.2);
  z-index: 100;
  min-width: 200px;
  padding: 8px 0;
}

.dropdown-item {
  padding: 8px 16px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 14px;
  color: #3c4043;
}

.dropdown-item:hover {
  background-color: #f1f3f4;
}

.menu-section {
  flex-grow: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.menu-item {
  display: flex;
  align-items: center;
  height: 40px;
  padding: 0 16px;
  border-radius: 16px;
  cursor: pointer;
  color: #3c4043;
  font-size: 14px;
  font-weight: 500;
  transition: background-color 0.1s;
  text-decoration: none;
}

.menu-item:hover {
  background-color: #f1f3f4;
}

.menu-item.active {
  background-color: #c2e7ff;
  color: #001d35;
}

.menu-item.active:hover {
  background-color: #c2e7ff;
}

.icon {
  margin-right: 12px;
  font-size: 18px;
  width: 20px;
  text-align: center;
  display: flex;
  justify-content: center;
}

.storage-section {
  margin-top: 16px;
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
</style>
