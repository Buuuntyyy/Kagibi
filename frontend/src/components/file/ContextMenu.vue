<template>
  <div 
    class="context-menu" 
    :style="{ top: y + 'px', left: x + 'px' }"
    @click.stop
    v-click-outside="close"
  >
    <slot name="custom-actions">
      <!-- Default actions if no slot provided, but we mainly use slots or events -->
    </slot>
  </div>
</template>

<script setup>
import { onMounted, onUnmounted } from 'vue';

const props = defineProps({
  x: Number,
  y: Number,
  item: Object
});

const emit = defineEmits(['close', 'action']);

const close = () => {
  emit('close');
};

// Simple click-outside directive logic or event listener
const handleClickOutside = (event) => {
    // If click is not inside the context menu
    const menu = document.querySelector('.context-menu');
    if (menu && !menu.contains(event.target)) {
        close();
    }
}

onMounted(() => {
    setTimeout(() => {
        document.addEventListener('click', handleClickOutside);
        document.addEventListener('contextmenu', handleClickOutside); // Close on other right click
    }, 0);
});

onUnmounted(() => {
    document.removeEventListener('click', handleClickOutside);
    document.removeEventListener('contextmenu', handleClickOutside);
});
</script>

<style scoped>
.context-menu {
  position: fixed;
  background: var(--card-color);
  border: 1px solid var(--border-color);
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
  border-radius: 8px;
  padding: 6px 0;
  z-index: 9999;
  min-width: 180px;
  overflow: hidden;
}

:deep(.menu-item) {
  padding: 10px 16px;
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  font-size: 0.9rem;
  color: var(--main-text-color);
  transition: background-color 0.1s;
}

:deep(.menu-item:hover) {
  background-color: var(--hover-background-color);
}

:deep(.menu-item.delete) {
  color: var(--error-color);
}

:deep(.menu-item.delete:hover) {
  background-color: #fee2e2;
}
</style>