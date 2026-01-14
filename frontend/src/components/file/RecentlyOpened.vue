<template>
  <div class="recently-opened">
    <div class="accordion-header" @click="toggleAccordion">
      <span class="chevron" :class="{ 'open': isOpen }">
        <svg xmlns="http://www.w3.org/2000/svg" height="24px" viewBox="0 0 24 24" width="24px" fill="currentColor"><path d="M0 0h24v24H0V0z" fill="none"/><path d="M10 6L8.59 7.41 13.17 12l-4.58 4.59L10 18l6-6z"/></svg>
      </span>
      <h4 class="section-title">Suggestions</h4>
    </div>
    
    <div v-show="isOpen" class="accordion-content">
      <div v-if="hasRecents" class="cards-row">
        <div 
          v-for="(item, index) in combinedRecents" 
          :key="index" 
          class="recent-card"
          :class="item.type"
          @click="openItem(item)"
        >
          <div class="icon-wrapper">
             <!-- Folder Icon -->
            <svg v-if="item.type === 'folder'" viewBox="0 0 24 24" width="24" height="24" fill="currentColor">
              <path d="M10 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z"></path>
            </svg>
             <!-- File Icon -->
            <svg v-else viewBox="0 0 24 24" width="24" height="24" fill="currentColor">
               <path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z"/>
            </svg>
          </div>
          
          <div class="card-details">
            <span class="item-name" :title="item.displayName">{{ item.displayName }}</span>
            <span class="item-type-text">{{ item.type === 'folder' ? 'Dossier' : 'Fichier' }}</span>
          </div>
        </div>
      </div>
      
      <div v-else class="empty-state">
        <span class="empty-icon">🕒</span>
        <span>Les éléments récemment ouverts apparaîtront ici</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useFileStore } from '../../stores/files'

const fileStore = useFileStore()
const isOpen = ref(true)

const toggleAccordion = () => {
  isOpen.value = !isOpen.value
}

const hasRecents = computed(() => {
  return fileStore.recentFolders.length > 0 || fileStore.recentFiles.length > 0
})

const combinedRecents = computed(() => {
  const folders = fileStore.recentFolders.map(f => ({ ...f, type: 'folder', displayName: f.name }));
  const files = fileStore.recentFiles.map(f => ({ ...f, type: 'file', displayName: f.Name }));
  
  // Combine and maybe slice to a reasonable number if needed, keeping separate lists structure 
  // but displaying them together. We could interleave them or just concat.
  // Given user wants "one line", concat is easiest.
  return [...folders, ...files];
})

import { useRouter } from 'vue-router'
const router = useRouter()

const openItem = async (item) => {
  if (item.type === 'folder') {
    // Set a pending navigation path so FileList.vue will use it after mount
    fileStore.pendingNavigatePath = item.path;
    if (router.currentRoute.value.path !== '/dashboard/files') {
      await router.push('/dashboard/files')
    } else {
      // If already on files, trigger navigation immediately
      fileStore.fetchItems(item.path)
    }
    fileStore.addToHistory(item)
  } else {
    fileStore.downloadFile(item.ID, item.Name)
    fileStore.addToHistory(item)
  }
}
</script>

<style scoped>
.recently-opened {
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
  display: flex;
  gap: 12px;
  overflow-x: auto;
  padding: 0.5rem 2px 1rem 2px; /* Bottom padding for shadow visibility */
  scrollbar-width: thin;
  scrollbar-color: var(--border-color) transparent;
}

.cards-row::-webkit-scrollbar {
  height: 6px;
}
.cards-row::-webkit-scrollbar-thumb {
  background-color: #ccc;
  border-radius: 3px;
}

.recent-card {
  background-color: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  min-width: 180px;
  max-width: 180px;
  height: 48px; /* Compact height like a chip/card row */
  padding: 0 12px;
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  transition: box-shadow 0.2s, border-color 0.2s;
  flex-shrink: 0;
}

.recent-card:hover {
  box-shadow: 0 1px 3px rgba(0,0,0,0.12);
  border-color: transparent;
}

.icon-wrapper {
  color: var(--secondary-text-color);
  display: flex;
  align-items: center;
}

.recent-card.folder .icon-wrapper {
  color: var(--secondary-text-color); /* Google folders are often grey in recent, or colored. Let's keep specific color if needed. */
}
/* Optional: Folder icon specific color */
.recent-card.folder .icon-wrapper svg {
    fill: #5f6368;
}
.recent-card.file .icon-wrapper svg {
    fill: #3498db; /* Blue for files */
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

.empty-state {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 1rem 0;
  color: var(--secondary-text-color);
  font-size: 0.9rem;
}

.empty-icon {
  font-size: 1.2rem;
  opacity: 0.7;
}
</style>
