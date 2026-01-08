<template>
  <div class="left-bar">
    <div class="action-section">
      <button class="btn-new" @click.stop="toggleNewMenu">
        <svg class="plus-icon" width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z" fill="currentColor"/>
        </svg>
        <span>Nouveau</span>
      </button>
      <div v-if="showNewMenu" class="new-menu-dropdown" @click.stop>
        <div class="dropdown-item" @click="triggerUpload">
          <svg class="icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z" fill="currentColor"/>
          </svg>
          <span>Fichier</span>
        </div>
        <div class="dropdown-item" @click="triggerCreateFolder">
          <svg class="icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M10 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z" fill="currentColor"/>
          </svg>
          <span>Dossier</span>
        </div>
      </div>
    </div>

    <div class="menu-section">
      <div class="menu-item" :class="{ active: isActive('/dashboard') }" @click="navigateTo('/dashboard')">
        <svg class="icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M20 6h-8l-2-2H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2zm0 12H4V8h16v10z" fill="currentColor"/>
        </svg>
        <span>Mes fichiers</span>
      </div>
      <div class="menu-item" :class="{ active: isActive('/dashboard/shares') }" @click="navigateTo('/dashboard/shares')">
        <svg class="icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M15 8c0-1.42-.5-2.73-1.33-3.76.42-.14.86-.24 1.33-.24 2.21 0 4 1.79 4 4s-1.79 4-4 4c-.43 0-.84-.09-1.23-.21-.03-.01-.06-.02-.1-.03A5.98 5.98 0 0 0 15 8zm1.66 5.13C18.03 14.06 19 15.32 19 17v3h4v-3c0-2.18-3.57-3.47-6.34-3.87zM9 6c-1.1 0-2 .9-2 2s.9 2 2 2 2-.9 2-2-.9-2-2-2m0 9c-2.7 0-5.8 1.29-6 4.02V21h12v-1.98c-.2-2.72-3.3-4.02-6-4.02z" fill="currentColor"/>
        </svg>
        <span>Fichiers partagés</span>
      </div>
      <div class="menu-item" :class="{ active: isFriendsOpen }" @click="$emit('toggle-friends')">
        <svg class="icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z" fill="currentColor"/>
        </svg>
        <span>Amis</span>
      </div>
      <div class="menu-item">
        <svg class="icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 17.93c-3.95-.49-7-3.85-7-7.93 0-.62.08-1.21.21-1.79L9 15v1c0 1.1.9 2 2 2v1.93zm6.9-2.54c-.26-.81-1-1.39-1.9-1.39h-1v-3c0-.55-.45-1-1-1H8v-2h2c.55 0 1-.45 1-1V7h2c1.1 0 2-.9 2-2v-.41c2.93 1.19 5 4.06 5 7.41 0 2.08-.8 3.97-2.1 5.39z" fill="currentColor"/>
        </svg>
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

const props = defineProps({
  isFriendsOpen: Boolean
})

const emit = defineEmits(['toggle-friends'])

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
  background-color: var(--card-color);
  border: none;
  border-radius: 16px;
  padding: 0 16px;
  height: 56px;
  box-shadow: 0 1px 2px 0 rgba(60,64,67,0.3), 0 1px 3px 1px rgba(60,64,67,0.15);
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 14px;
  font-weight: 500;
  color: var(--main-text-color);
  width: 100%;
}

.btn-new:hover {
  box-shadow: 0 4px 8px 3px rgba(60,64,67,0.15);
  background-color: var(--hover-background-color);
}

.plus-icon {
  min-width: 24px;
  fill: currentColor;
}

.new-menu-dropdown {
  position: absolute;
  top: 60px;
  left: 0;
  background: var(--card-color);
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
  color: var(--main-text-color);
}

.dropdown-item:hover {
  background-color: var(--hover-background-color);
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
  height: 32px;
  padding: 0 12px 0 16px;
  border-radius: 16px;
  cursor: pointer;
  color: var(--main-text-color);
  font-size: 14px;
  font-weight: 500;
  transition: background-color 0.1s;
  text-decoration: none;
}

.menu-item:hover {
  background-color: var(--hover-background-color);
}

.menu-item.active {
  background-color: #c2e7ff;
  color: #001d35;
}

.menu-item.active:hover {
  background-color: #c2e7ff;
}

.icon-svg {
  width: 20px;
  height: 20px;
  margin-right: 12px;
  fill: currentColor;
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
  color: var(--secondary-text-color);
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
