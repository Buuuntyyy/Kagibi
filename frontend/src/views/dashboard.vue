<template>
  <div class="dashboard-container">
    <LeftBar @toggle-friends="toggleFriends" :isFriendsOpen="showFriends" />
    
    <div 
      class="friends-wrapper" 
      :class="{ show: showFriends, resizing: isResizing }"
      :style="wrapperStyle"
    >
      <FriendsSidebar @close="showFriends = false" />
      <div class="resizer" @mousedown.prevent="startResize"></div>
    </div>

    <div class="main-content">
      <router-view />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onUnmounted } from 'vue'
import LeftBar from '../components/bar/leftBar.vue'
import FriendsSidebar from '../components/FriendsSidebar.vue'

const showFriends = ref(false)
const sidebarWidth = ref(350)
const isResizing = ref(false)

const toggleFriends = () => {
  showFriends.value = !showFriends.value
}

const wrapperStyle = computed(() => {
  if (!showFriends.value) return { width: '0px' }
  return { 
    width: `${sidebarWidth.value}px`,
    transition: isResizing.value ? 'none' : 'width 0.3s cubic-bezier(0.4, 0, 0.2, 1)'
  }
})

const startResize = () => {
  isResizing.value = true
  document.addEventListener('mousemove', handleResize)
  document.addEventListener('mouseup', stopResize)
  // Disable selection during resize
  document.body.style.userSelect = 'none'
  document.body.style.cursor = 'ew-resize'
}

const handleResize = (e) => {
  if (!isResizing.value) return
  
  // Calculate new width relative to the left sidebar
  // LeftBar is typically 256px wide + padding.
  // Better to calculate based on movement or absolute position relative to wrapper start.
  // But wrapper start is fixed after LeftBar.
  // Let's assume LeftBar width is constant or we can get it via ref if needed,
  // but simpler is to use movementX or just absolute clientX minus left offset.
  
  // Actually, since the wrapper starts after LeftBar, we can just use the mouse position minus the LeftBar width.
  // Assuming LeftBar is roughly 256px + padding. 
  // Let's be dynamic: Sidebar width = MouseX - LeftBarWidth.
  
  // Safer approach: delta update.
  const newWidth = sidebarWidth.value + e.movementX
  
  // Clamp values
  if (newWidth > 200 && newWidth < 800) {
    sidebarWidth.value = newWidth
  }
}

const stopResize = () => {
  isResizing.value = false
  document.removeEventListener('mousemove', handleResize)
  document.removeEventListener('mouseup', stopResize)
  document.body.style.userSelect = ''
  document.body.style.cursor = ''
}

onUnmounted(() => {
  stopResize()
})
</script>

<style scoped>
.dashboard-container {
  display: flex;
  height: 100%;
  width: 100%;
  box-sizing: border-box;
  background-color: var(--background-color);
}

.friends-wrapper {
  overflow: hidden;
  flex-shrink: 0;
  position: relative;
  display: flex;
}

.friends-wrapper.show {
  /* width handled by style binding */
}

.resizer {
  width: 5px;
  cursor: ew-resize;
  background-color: transparent;
  position: absolute;
  right: 0;
  top: 0;
  bottom: 0;
  z-index: 10;
  transition: background-color 0.2s;
}

.resizer:hover, .friends-wrapper.resizing .resizer {
  background-color: var(--primary-color);
}

.main-content {
  flex-grow: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  border-top-left-radius: 30px;
  background-color: var(--card-color);
}
</style>
