<template>
  <div class="avatar-selector">
    <div class="avatar-grid">
      <div 
        v-for="avatar in avatars" 
        :key="avatar.url"
        class="avatar-option"
        :class="{ selected: modelValue === avatar.url }"
        @click="selectAvatar(avatar.url)"
        :title="avatar.name"
      >
        <img 
          :src="avatar.url" 
          :alt="avatar.name"
          loading="lazy"
        />
        <div v-if="modelValue === avatar.url" class="check-badge">
          <svg viewBox="0 0 24 24" fill="white" width="16" height="16">
            <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/>
          </svg>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  modelValue: {
    type: String,
    default: '/avatars/default.svg'
  }
})

const emit = defineEmits(['update:modelValue'])

// Liste des avatars disponibles
const avatars = computed(() => {
  const avatarList = [
    { url: '/avatars/default.svg', name: 'Avatar par défaut' }
  ]
  
  // Générer avatars 1-9
  for (let i = 1; i <= 9; i++) {
    avatarList.push({
      url: `/avatars/avatar${i}.svg`,
      name: `Avatar ${i}`
    })
  }
  
  return avatarList
})

const selectAvatar = (url) => {
  emit('update:modelValue', url)
}
</script>

<style scoped>
.avatar-selector {
  width: 100%;
}

.avatar-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(80px, 1fr));
  gap: 12px;
  max-width: 100%;
}

.avatar-option {
  position: relative;
  aspect-ratio: 1;
  border-radius: 12px;
  overflow: hidden;
  cursor: pointer;
  border: 3px solid transparent;
  transition: all 0.2s ease;
  background: var(--background-color);
}

.avatar-option:hover {
  border-color: var(--primary-color);
  transform: scale(1.05);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.avatar-option.selected {
  border-color: var(--primary-color);
  box-shadow: 0 0 0 4px rgba(99, 102, 241, 0.1);
}

.avatar-option img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.check-badge {
  position: absolute;
  bottom: 4px;
  right: 4px;
  width: 24px;
  height: 24px;
  background: var(--primary-color);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

@media (max-width: 768px) {
  .avatar-grid {
    grid-template-columns: repeat(auto-fill, minmax(60px, 1fr));
    gap: 8px;
  }
  
  .check-badge {
    width: 20px;
    height: 20px;
  }
}
</style>
