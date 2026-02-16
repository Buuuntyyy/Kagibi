<template>
  <div class="left-bar-container">
  <div class="left-bar" :class="{ 'collapsed': isCollapsed }">
    <div class="action-section">
      <button class="btn-new" @click.stop="toggleNewMenu">
        <svg class="plus-icon" width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z" fill="currentColor"/>
        </svg>
        <span>{{ t('sidebar.new') }}</span>
      </button>
      <div v-if="showNewMenu" class="new-menu-dropdown" @click.stop>
        <div class="dropdown-item" @click="triggerUpload">
          <svg class="icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z" fill="currentColor"/>
          </svg>
          <span>{{ t('nav.uploadFile') }}</span>
        </div>
        <div class="dropdown-item" @click="triggerCreateFolder">
          <svg class="icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M10 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z" fill="currentColor"/>
          </svg>
          <span>{{ t('nav.createFolder') }}</span>
        </div>
        <div class="dropdown-item" @click="triggerP2P">
           <svg class="icon-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="22" y1="2" x2="11" y2="13"></line>
              <polygon points="22 2 15 22 11 13 2 9 22 2"></polygon>
           </svg>
           <span>{{ t('sidebar.p2pTransfer') }}</span>
        </div>
      </div>
    </div>

    <div class="menu-section">
      <div class="menu-item" :class="{ active: isActive('/dashboard/home') }" @click="navigateTo('/dashboard/home')">
        <svg class="icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M10 20v-6h4v6h5v-8h3L12 3 2 12h3v8z" fill="currentColor"/>
        </svg>
        <span>{{ t('nav.dashboard') }}</span>
      </div>
      <div class="menu-item" :class="{ active: isActive('/dashboard/files') }" @click="navigateTo('/dashboard/files')">
        <svg class="icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M20 6h-8l-2-2H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2zm0 12H4V8h16v10z" fill="currentColor"/>
        </svg>
        <span>{{ t('nav.files') }}</span>
      </div>
      <div class="menu-item" :class="{ active: isActive('/dashboard/shares') }" @click="navigateTo('/dashboard/shares')">
        <svg class="icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M15 8c0-1.42-.5-2.73-1.33-3.76.42-.14.86-.24 1.33-.24 2.21 0 4 1.79 4 4s-1.79 4-4 4c-.43 0-.84-.09-1.23-.21-.03-.01-.06-.02-.1-.03A5.98 5.98 0 0 0 15 8zm1.66 5.13C18.03 14.06 19 15.32 19 17v3h4v-3c0-2.18-3.57-3.47-6.34-3.87zM9 6c-1.1 0-2 .9-2 2s.9 2 2 2 2-.9 2-2-.9-2-2-2m0 9c-2.7 0-5.8 1.29-6 4.02V21h12v-1.98c-.2-2.72-3.3-4.02-6-4.02z" fill="currentColor"/>
        </svg>
        <span>{{ t('nav.shared') }}</span>
      </div>
      <div class="menu-item" :class="{ active: isActive('/p2p') }" @click="navigateTo('/p2p')">
        <svg class="icon-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
             <line x1="22" y1="2" x2="11" y2="13"></line>
             <polygon points="22 2 15 22 11 13 2 9 22 2"></polygon>
        </svg>
        <span>{{ t('nav.p2pTransfer') }}</span>
      </div>

      <!-- Friends Accordion -->
      <div class="friends-accordion" :class="{ open: friendsOpen }">
        <div class="menu-item" @click="toggleFriendsAccordion">
            <svg class="icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z" fill="currentColor"/>
            </svg>
            <span>{{ t('sidebar.myFriends') }}</span>
            <svg class="arrow-icon" viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
                <path d="M7.41 8.59L12 13.17l4.59-4.58L18 10l-6 6-6-6 1.41-1.41z"/>
            </svg>
        </div>
        <div class="accordion-content" v-if="friendsOpen">
            <div class="friend-header" v-if="authStore.user && authStore.user.friend_code">
                <div class="code-wrapper">
                  <span class="my-code" @click="copyFriendCode" :title="t('friends.copyCode')">{{ authStore.user.friend_code }}</span>
                  <div class="info-container-left">
                     <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor" class="info-icon" @click.stop="toggleLeftInfo">
                        <path d="M11 18h2v-2h-2v2zm1-16C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8zm0-14c-2.21 0-4 1.79-4 4h2c0-1.1.9-2 2-2s2 .9 2 2c0 2-3 1.75-3 5h2c0-2.25 3-2.5 3-5 0-2.21-1.79-4-4-4z"/>
                    </svg>
                    <div v-if="showLeftInfo" class="info-tooltip-mini">
                        {{ t('friends.myCode') }}
                    </div>
                  </div>
                </div>

                <div class="header-actions">
                  <button class="add-friend-btn" @click.stop="toggleAddFriendMenu" :title="t('sidebar.addFriend')">
                      <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
                          <path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z"/>
                      </svg>
                  </button>
                </div>

                <div v-if="showAddFriendMenu" class="add-friend-popup" @click.stop>
                    <input id="add-friend-input" v-model="newFriendId" :placeholder="t('friends.enterCode')" @keydown.enter="addFriend" />
                    <button @click="addFriend" class="confirm-add">OK</button>
                </div>
            </div>
            <FriendsSidebar />
        </div>
      </div>
    </div>

    <div class="storage-section" v-if="billingStore.showSubscriptionUI" @click="navigateTo('/billing')" style="cursor: pointer;">
      <div class="storage-info" v-if="!isCollapsed">
        <span class="storage-label">{{ t('sidebar.storage') }}</span>
        <span class="storage-value">{{ formattedStorageUsed }} {{ t('sidebar.of') }} {{ formattedStorageLimit }}</span>
      </div>
      <div class="storage-info-collapsed" v-else :title="storagePercentageInt + '% utilisé sur ' + storageLimitGB">
             <svg class="progress-ring" viewBox="0 0 36 36">
                <path class="ring-bg"
                    d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                />
                <path class="ring-fill"
                    :stroke-dasharray="storagePercentage + ', 100'"
                    d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                />
                <text x="18" y="14" class="ring-text-pct">{{ storagePercentageInt }}%</text>
                <text x="18" y="24" class="ring-text-limit">{{ storageLimitGB }}</text>
             </svg>
      </div>
      <div class="storage-bar" v-if="!isCollapsed">
        <div class="storage-fill" :style="{ width: storagePercentage + '%' }"></div>
      </div>
    </div>

    <div class="collapse-toggle" @click="toggleCollapse" :title="isCollapsed ? 'Agrandir' : 'Réduire'">
        <svg class="toggle-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline v-if="!isCollapsed" points="15 18 9 12 15 6"></polyline>
            <polyline v-else points="9 18 15 12 9 6"></polyline>
        </svg>
    </div>
  </div>

    <input type="file" ref="fileInput" @change="handleFileUpload" style="display: none" multiple />
    <div class="dialogs">
      <InputDialog
        v-model:isOpen="inputDialog.isOpen"
        :title="inputDialog.title"
        :defaultValue="inputDialog.defaultValue"
        :placeholder="inputDialog.placeholder"
        @confirm="handleInputConfirm"
        @cancel="handleInputCancel"
      />
    </div>
  </div>
</template>

<script setup>
import { computed, ref, onMounted, onUnmounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../../stores/auth'
import { useFileStore } from '../../stores/files'
import { useFriendStore } from '../../stores/friends'
import { useBillingStore } from '../../stores/billing'
import { uploadQueueManager } from '../../utils/uploadQueueManager'
import InputDialog from '../InputDialog.vue'
import FriendsSidebar from '../FriendsSidebar.vue'

const { t } = useI18n()

const props = defineProps({
  // No props needed now for layout control
})

const authStore = useAuthStore()
const fileStore = useFileStore()
const friendStore = useFriendStore()
const billingStore = useBillingStore()
const router = useRouter()
const route = useRoute()

const showNewMenu = ref(false)
const friendsOpen = ref(true) // Default open or closed
const fileInput = ref(null)
const showAddFriendMenu = ref(false)
const newFriendId = ref('')
const showLeftInfo = ref(false)
const isCollapsed = ref(false)

// Auto-collapse menu on P2P page
watch(() => route.name, (newRouteName) => {
  isCollapsed.value = newRouteName === 'P2P'
}, { immediate: true })

const closeAddFriendMenu = () => {
    if (showAddFriendMenu.value) {
        showAddFriendMenu.value = false
    }
    if (showLeftInfo.value) {
        showLeftInfo.value = false
    }
}

const toggleLeftInfo = () => {
    showLeftInfo.value = !showLeftInfo.value
}

const toggleAddFriendMenu = () => {
  showAddFriendMenu.value = !showAddFriendMenu.value
  if (showAddFriendMenu.value) {
    setTimeout(() => document.getElementById('add-friend-input')?.focus(), 100)
  }
}

onMounted(() => {
    window.addEventListener('click', closeAddFriendMenu)
})

onUnmounted(() => {
    window.removeEventListener('click', closeAddFriendMenu)
})

const addFriend = async () => {
  if (!newFriendId.value) return;
  try {
    await friendStore.sendRequest(newFriendId.value)
    showAddFriendMenu.value = false
    newFriendId.value = ''
  } catch (e) {
    console.error(e)
  }
}

const copyFriendCode = () => {
    if (authStore.user?.friend_code) {
        navigator.clipboard.writeText(authStore.user.friend_code)
    }
}

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

const toggleFriendsAccordion = () => {
    if (isCollapsed.value) {
        isCollapsed.value = false
        friendsOpen.value = true
    } else {
        friendsOpen.value = !friendsOpen.value
    }
}

const toggleCollapse = () => {
  isCollapsed.value = !isCollapsed.value
  if (isCollapsed.value) {
    friendsOpen.value = false
  }
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
  const files = event.target.files
  if (files && files.length > 0) {
    // Use the queue manager for multi-file uploads
    await uploadQueueManager.addFiles(files, fileStore.currentPath)
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

const triggerP2P = () => {
    showNewMenu.value = false;
    router.push('/p2p')
}

const formatSize = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Number.parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
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

const storagePercentageInt = computed(() => Math.round(storagePercentage.value))

const storageLimitGB = computed(() => {
    const limit = authStore.user?.storage_limit || 10737418240
    return Math.round(limit / (1024 * 1024 * 1024)) + ' Go'
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
  position: relative;
  z-index: 5;
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
  background-color: var(--hover-background-color);
  color: var(--primary-color);
  font-weight: bold;
}

.menu-item.active:hover {
  background-color: var(--hover-background-color);
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
.storage-fill {
  height: 100%;
  background-color: var(--primary-color);
  transition: width 0.3s ease;
}

.storage-bar {
  height: 6px;
  background-color: #e9ecef;
  border-radius: 3px;
  overflow: hidden;
}

.friends-accordion {
  display: flex;
  flex-direction: column;
}

.friends-accordion .menu-item {
  justify-content: space-between;
}

.friends-accordion .menu-item span {
    flex-grow: 1;
}

.arrow-icon {
  width: 16px;
  height: 16px;
  transition: transform 0.3s ease;
}

.friends-accordion.open .arrow-icon {
  transform: rotate(180deg);
}

.accordion-content {
  overflow: visible;
  display: flex;
  flex-direction: column;
  animation: slideDown 0.3s ease-out;
}

.btn-text-small {
  background: none;
  border: none;
  color: var(--secondary-text-color); /* Was primary color, now more subtle */
  font-size: 0.8rem;
  cursor: pointer;
  padding: 8px 16px;
  text-align: left;
  width: 100%;
  margin-left: 4px;
  transition: color 0.2s;
}

.btn-text-small:hover {
  color: var(--main-text-color);
  text-decoration: none; /* Removed underline */
}

.accordion-actions {
    border-top: 1px solid var(--border-color);
    margin-top: 4px;
}

.friend-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 16px; /* Aligned with typical menu item padding */
    background-color: transparent; /* Cleaner look */
    font-size: 0.85rem;
    position: relative;
    border-bottom: 1px solid var(--border-color);
    margin-bottom: 4px;
}

.code-wrapper {
    display: flex;
    align-items: center;
    gap: 6px;
}

.info-container-left {
    position: relative;
    display: flex;
    align-items: center;
}

.info-icon {
    color: var(--secondary-text-color);
    cursor: pointer;
    opacity: 0.7;
    transition: opacity 0.2s;
}

.info-icon:hover {
    opacity: 1;
    color: var(--primary-color);
}

.info-tooltip-mini {
    position: absolute;
    top: 20px;
    left: 0;
    width: 180px;
    background-color: #333;
    color: white;
    padding: 8px;
    border-radius: 4px;
    font-size: 0.75rem;
    z-index: 200;
    line-height: 1.3;
    box-shadow: 0 4px 10px rgba(0,0,0,0.2);
}

.my-code {
    font-family: monospace;
    background: rgba(0,0,0,0.05);
    padding: 4px 8px;
    border-radius: 4px;
    cursor: pointer;
    user-select: all;
    color: var(--main-text-color);
    font-size: 0.8rem;
    border: 1px solid transparent;
    transition: all 0.2s;
}

.my-code:hover {
    background: rgba(0,0,0,0.08);
    border-color: var(--border-color);
}

.add-friend-btn {
    background: transparent;
    border: none;
    cursor: pointer;
    width: 24px;
    height: 24px;
    padding: 0;
    border-radius: 50%;
    color: var(--secondary-text-color);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1.2rem;
    line-height: 1;
    transition: all 0.2s;
    margin-left: 8px;
    flex-shrink: 0;
}

.add-friend-btn:hover {
    color: var(--primary-color);
    background-color: var(--hover-background-color, rgba(0,0,0,0.05));
}

.add-friend-popup {
    position: absolute;
    top: 36px; /* Just below the header */
    left: 10px; /* Align left */
    right: 10px; /* Stretch to right - padding */
    background: var(--card-color);
    border: 1px solid var(--border-color);
    box-shadow: 0 4px 16px rgba(0,0,0,0.2);
    padding: 12px;
    border-radius: 8px;
    z-index: 1000;
    display: flex;
    flex-direction: column;
    gap: 8px;
}

.add-friend-popup input {
    width: 100%;
    padding: 8px;
    font-size: 0.9rem;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    background: var(--input-background);
    color: var(--main-text-color);
    box-sizing: border-box; /* Important for width 100% */
}

.confirm-add {
    width: 100%;
    font-size: 0.85rem;
    padding: 6px 0;
    background: var(--primary-color);
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-weight: 500;
}



@keyframes slideDown {
  from { opacity: 0; transform: translateY(-10px); }
  to { opacity: 1; transform: translateY(0); }
}

.left-bar-container {
    height: 100%;
    display: flex;
    flex-direction: column;
}

.dialogs {
    position: absolute;
}

/* Collapsed State Styles */
.left-bar {
  transition: width 0.3s ease;
}

.left-bar.collapsed {
  width: 72px;
  padding: 8px;
}

.left-bar.collapsed .btn-new {
  padding: 0;
  justify-content: center;
}
.left-bar.collapsed .btn-new span { display: none; }
.left-bar.collapsed .plus-icon { margin: 0; }

.left-bar.collapsed .menu-item span,
.left-bar.collapsed .arrow-icon {
  display: none;
}

.left-bar.collapsed .menu-item {
  justify-content: center;
  padding: 0;
}
.left-bar.collapsed .icon-svg { margin: 0; }

.left-bar.collapsed .accordion-content {
  display: none;
}

.left-bar.collapsed .new-menu-dropdown {
  left: 10px;
}

.collapse-toggle {
    margin-top: 1rem;
    padding: 10px;
    cursor: pointer;
    display: flex;
    justify-content: center;
    color: var(--secondary-text-color);
    border-top: 1px solid var(--border-color);
    transition: all 0.2s;
}
.collapse-toggle:hover {
    color: var(--primary-color);
    background: var(--hover-background-color);
    border-radius: 8px;
}
.toggle-icon {
    width: 20px;
    height: 20px;
}

.left-bar.collapsed .storage-section {
    padding: 6px;
}

.storage-info-collapsed {
    display: flex;
    justify-content: center;
    align-items: center;
    width: 100%;
}

.progress-ring {
    width: 42px;
    height: 42px;
}

.ring-bg {
    fill: none;
    stroke: var(--border-color);
    stroke-width: 3;
}

.ring-fill {
    fill: none;
    stroke: var(--primary-color);
    stroke-width: 3;
    stroke-linecap: round;
    transition: stroke-dasharray 0.3s ease;
}

.ring-text-pct {
    font-size: 8px;
    font-weight: bold;
    fill: var(--main-text-color);
    text-anchor: middle;
}

.ring-text-limit {
    font-size: 6px;
    fill: var(--secondary-text-color);
    text-anchor: middle;
}

/* Hide line bar in collapsed mode, already handled by v-if in template but cleaning css */
.left-bar.collapsed .storage-bar {
    display: none;
}
</style>
