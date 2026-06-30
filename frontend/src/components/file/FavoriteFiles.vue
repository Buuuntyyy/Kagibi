<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="favorite-files" v-if="hasFavorites">
    <div class="accordion-header" @click="toggleAccordion">
      <span class="chevron" :class="{ 'open': isOpen }">
        <svg xmlns="http://www.w3.org/2000/svg" height="24px" viewBox="0 0 24 24" width="24px" fill="currentColor"><path d="M0 0h24v24H0V0z" fill="none"/><path d="M10 6L8.59 7.41 13.17 12l-4.58 4.59L10 18l6-6z"/></svg>
      </span>
      <h4 class="section-title">{{ t('home.favorites') }}</h4>
    </div>

    <div v-show="isOpen" class="accordion-content">
      <div class="cards-row">
        <div
          v-for="(item, index) in combinedFavorites"
          :key="index"
          class="recent-card"
          :class="item.type"
          @click="openItem(item)"
          @contextmenu.prevent.stop="handleContextMenu($event, item)"
        >
          <div class="icon-wrapper">
            <svg v-if="item.type === 'folder'" viewBox="0 0 24 24" width="24" height="24" fill="currentColor">
              <path d="M10 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z"></path>
            </svg>
            <svg v-else viewBox="0 0 24 24" width="24" height="24" fill="currentColor">
              <path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z"/>
            </svg>
          </div>

          <div class="card-details">
            <span class="item-name" :title="item.displayName">{{ item.displayName }}</span>
            <span class="item-type-text">{{ item.type === 'folder' ? t('file.folderType') : t('file.fileType') }}</span>
          </div>

          <button class="remove-fav-btn" :title="t('file.removeFavorite')" @click.stop="removeFavorite(item)">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
              <path d="M12 17.27L18.18 21l-1.64-7.03L22 9.24l-7.19-.61L12 2 9.19 8.63 2 9.24l5.46 4.73L5.82 21z"/>
            </svg>
          </button>
        </div>
      </div>
    </div>

    <ContextMenu
      v-if="contextMenu.visible"
      :x="contextMenu.x"
      :y="contextMenu.y"
      :item="contextMenu.item"
      @close="closeContextMenu"
    >
      <template #custom-actions>
        <div class="menu-item" @click="handleContextAction('preview')" v-if="contextMenu.item.type === 'file'">
          <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
            <path d="M12 4.5C7 4.5 2.73 7.61 1 12c1.73 4.39 6 7.5 11 7.5s9.27-3.11 11-7.5c-1.73-4.39-6-7.5-11-7.5zM12 17c-2.76 0-5-2.24-5-5s2.24-5 5-5 5 2.24 5 5-2.24 5-5 5zm0-8c-1.66 0-3 1.34-3 3s1.34 3 3 3 3-1.34 3-3-1.34-3-3-3z"/>
          </svg>
          {{ t('file.preview') }}
        </div>
        <div class="menu-item" @click="handleContextAction('download')" v-if="contextMenu.item.type === 'file'">
          <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
            <path d="M19 9h-4V3H9v6H5l7 7 7-7zM5 18v2h14v-2H5z"/>
          </svg>
          {{ t('file.download') }}
        </div>
        <div class="menu-item" @click="handleContextAction('remove-favorite')">
          <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
            <path d="M12 17.27L18.18 21l-1.64-7.03L22 9.24l-7.19-.61L12 2 9.19 8.63 2 9.24l5.46 4.73L5.82 21z"/>
          </svg>
          {{ t('file.removeFavorite') }}
        </div>
      </template>
    </ContextMenu>
  </div>
</template>

<script setup>
import { computed, ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useFileStore } from '../../stores/files'
import { useRouter } from 'vue-router'
import ContextMenu from './ContextMenu.vue'

const fileStore = useFileStore()
const { t } = useI18n()
const router = useRouter()
const isOpen = ref(true)

const contextMenu = ref({
  visible: false,
  x: 0,
  y: 0,
  item: null
})

onMounted(() => {
  fileStore.fetchUserFavorites()
})

const toggleAccordion = () => {
  isOpen.value = !isOpen.value
}

const hasFavorites = computed(() => {
  return fileStore.favoriteFolders.length > 0 || fileStore.favoriteFiles.length > 0
})

const combinedFavorites = computed(() => {
  const folders = fileStore.favoriteFolders.map(f => ({
    ...f,
    type: 'folder',
    displayName: f.displayName || f.Name || f.name
  }))
  const files = fileStore.favoriteFiles.map(f => ({
    ...f,
    type: 'file',
    displayName: f.displayName || f.Name || f.name
  }))
  return [...folders, ...files]
})

const openItem = async (item) => {
  if (item.type === 'folder') {
    fileStore.pendingNavigatePath = item.path || item.Path
    if (router.currentRoute.value.path !== '/dashboard/files') {
      await router.push('/dashboard/files')
    } else {
      fileStore.fetchItems(item.path || item.Path)
    }
    fileStore.addToHistory(item)
  } else {
    const fileId = item.ID || item.id
    const fileName = item.displayName || item.Name || item.name
    const mimeType = item.MimeType || item.mime_type
    const encryptedKey = item.EncryptedKey || item.encrypted_key
    fileStore.downloadFile(fileId, fileName, mimeType, false, encryptedKey)
    fileStore.addToHistory(item)
  }
}

const removeFavorite = (item) => {
  fileStore.removeUserFavorite(item)
}

const handleContextMenu = (event, item) => {
  contextMenu.value = {
    visible: true,
    x: event.clientX,
    y: event.clientY,
    item: item
  }
}

const closeContextMenu = () => {
  contextMenu.value.visible = false
}

const handleContextAction = (action) => {
  const item = contextMenu.value.item
  const fileId = item.ID || item.id
  const fileName = item.displayName || item.Name || item.name
  const mimeType = item.MimeType || item.mime_type
  const encryptedKey = item.EncryptedKey || item.encrypted_key

  switch (action) {
    case 'preview':
      if (item.type === 'file') {
        fileStore.downloadFile(fileId, fileName, mimeType, true, encryptedKey)
      }
      break
    case 'download':
      if (item.type === 'file') {
        fileStore.downloadFile(fileId, fileName, mimeType, false, encryptedKey)
      }
      break
    case 'remove-favorite':
      fileStore.removeUserFavorite(item)
      break
  }

  closeContextMenu()
}
</script>

<style scoped>
.favorite-files {
  padding: 0.5rem 1rem 0 1rem;
  background-color: var(--card-color);
  margin-bottom: 0.5rem;
}

.accordion-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  padding: 0.5rem;
  margin-left: -0.5rem;
  user-select: none;
  border-radius: 15px;
  transition: background-color 0.2s;
  width: fit-content;
}

.accordion-header:hover {
  background-color: var(--hover-background-color);
}

.section-title {
  margin: 0;
  font-size: 1.2rem;
  color: var(--main-text-color);
  font-weight: 550;
}

.chevron {
  display: flex;
  align-items: center;
  transition: transform 0.3s ease;
  color: var(--secondary-text-color);
}

.chevron.open {
  transform: rotate(90deg);
}

.cards-row {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  grid-auto-rows: 48px;
  gap: 12px;
  overflow: hidden;
  padding: 0.5rem 2px 1rem 2px;
}

.recent-card {
  background-color: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  width: 100%;
  height: 48px;
  padding: 0 8px 0 12px;
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  transition: box-shadow 0.2s, border-color 0.2s;
  box-sizing: border-box;
}

.recent-card:hover {
  box-shadow: 0 1px 3px rgba(0,0,0,0.12);
  border-color: transparent;
  background-color: var(--hover-background-color);
}

.icon-wrapper {
  color: var(--secondary-text-color);
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

.recent-card.folder .icon-wrapper svg {
  fill: var(--secondary-text-color);
}
.recent-card.file .icon-wrapper svg {
  fill: var(--primary-color);
}

.card-details {
  display: flex;
  flex-direction: column;
  justify-content: center;
  overflow: hidden;
  flex: 1;
}

.item-name {
  font-size: 0.85rem;
  font-weight: 500;
  color: var(--main-text-color);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  text-align: left;
}

.item-type-text {
  font-size: 0.7rem;
  color: var(--secondary-text-color);
  text-align: left;
}

.remove-fav-btn {
  background: none;
  border: none;
  cursor: pointer;
  padding: 4px;
  display: flex;
  align-items: center;
  border-radius: 50%;
  color: #f4b400;
  flex-shrink: 0;
  opacity: 0;
  transition: opacity 0.15s, background-color 0.15s;
}

.recent-card:hover .remove-fav-btn {
  opacity: 1;
}

.remove-fav-btn:hover {
  background-color: var(--hover-background-color);
}
</style>
