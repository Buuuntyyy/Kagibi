<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="avatar-selector">
    <!-- Vue compacte : juste l'avatar actuel + bouton -->
    <div class="current-avatar-compact" @click="showModal = true">
      <img :src="normalizedModelValue" :alt="getCurrentAvatarName()" class="compact-avatar" />
      <div class="change-overlay">
        <svg viewBox="0 0 24 24" fill="white" width="20" height="20">
          <path d="M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25zM20.71 7.04c.39-.39.39-1.02 0-1.41l-2.34-2.34c-.39-.39-1.02-.39-1.41 0l-1.83 1.83 3.75 3.75 1.83-1.83z"/>
        </svg>
        <span>Changer</span>
      </div>
    </div>

    <!-- Modal de sélection -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showModal" class="modal-backdrop" @click="showModal = false">
          <div class="modal-content" @click.stop>
            <div class="modal-header">
              <h3>Choisir un avatar</h3>
              <button class="close-btn" @click="showModal = false">
                <svg viewBox="0 0 24 24" fill="currentColor" width="24" height="24">
                  <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/>
                </svg>
              </button>
            </div>
            <div class="avatar-grid">
              <div
                v-for="avatar in avatars"
                :key="avatar.url"
                class="avatar-option"
                :class="{ selected: normalizedModelValue === avatar.url }"
                @click="selectAvatar(avatar.url)"
                :title="avatar.name"
              >
                <img
                  :src="avatar.url"
                  :alt="avatar.name"
                  loading="lazy"
                />
                <div v-if="normalizedModelValue === avatar.url" class="check-badge">
                  <svg viewBox="0 0 24 24" fill="white" width="16" height="16">
                    <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/>
                  </svg>
                </div>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'

const showModal = ref(false)

const props = defineProps({
  modelValue: {
    type: String,
    default: '/avatars/default.png'
  }
})

const emit = defineEmits(['update:modelValue'])

// Liste des avatars disponibles (correspondant aux fichiers réels)
const avatars = computed(() => [
  { url: '/avatars/default.png', name: 'Avatar par défaut' },
  { url: '/avatars/boy.png', name: 'Garçon' },
  { url: '/avatars/girl.png', name: 'Fille' },
  { url: '/avatars/cat.png', name: 'Chat' },
  { url: '/avatars/dog.png', name: 'Chien' },
  { url: '/avatars/fox.png', name: 'Renard' },
  { url: '/avatars/rabbit.png', name: 'Lapin' },
  { url: '/avatars/panda.png', name: 'Panda' },
  { url: '/avatars/gorilla.png', name: 'Gorille' },
  { url: '/avatars/chicken.png', name: 'Poulet' }
])

const selectAvatar = (url) => {
  emit('update:modelValue', url)
  showModal.value = false
}

// Normalize the current avatar URL to match the list
const normalizedModelValue = computed(() => {
  if (!props.modelValue) return '/avatars/default.png'

  const url = props.modelValue

  // If URL starts with http, return as-is
  if (url.startsWith('http')) return url

  // If URL already starts with /avatars/, return as-is
  if (url.startsWith('/avatars/')) return url

  // If URL starts with /, remove it
  const cleanUrl = url.startsWith('/') ? url.substring(1) : url

  // Add /avatars/ prefix
  return `/avatars/${cleanUrl}`
})

const getCurrentAvatarName = () => {
  const current = avatars.value.find(a => a.url === normalizedModelValue.value)
  return current ? current.name : 'Avatar'
}
</script>

<style scoped>
/* Vue compacte */
.current-avatar-compact {
  position: relative;
  width: 140px; /* Increased to match height of 2 inputs + gap roughly */
  height: 140px;
  border-radius: 12px;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.3s ease;
  /* border: 3px solid var(--border-color); Removed border */
}

.current-avatar-compact:hover {
  transform: scale(1.05);
  border-color: var(--primary-color);
  box-shadow: 0 4px 16px rgba(99, 102, 241, 0.2);
}

.current-avatar-compact:hover .change-overlay {
  opacity: 1;
}

.compact-avatar {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.change-overlay {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  opacity: 0;
  transition: opacity 0.3s ease;
}

.change-overlay span {
  color: white;
  font-size: 0.85rem;
  font-weight: 600;
}

/* Modal */
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.modal-content {
  background: var(--card-color);
  border-radius: 16px;
  padding: 1.5rem;
  max-width: 500px;
  width: 100%;
  max-height: 80vh;
  overflow-y: auto;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1.5rem;
}

.modal-header h3 {
  margin: 0;
  color: var(--main-text-color);
  font-size: 1.25rem;
}

.close-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--secondary-text-color);
  padding: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  transition: all 0.2s;
}

.close-btn:hover {
  background: var(--background-color);
  color: var(--main-text-color);
}

/* Grille d'avatars dans le modal */
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

/* Transitions */
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.3s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-active .modal-content,
.modal-leave-active .modal-content {
  transition: transform 0.3s ease;
}

.modal-enter-from .modal-content {
  transform: scale(0.9);
}

.modal-leave-to .modal-content {
  transform: scale(0.9);
}

@media (max-width: 768px) {
  .avatar-grid {
    grid-template-columns: repeat(4, 1fr);
    gap: 10px;
  }

  .modal-content {
    max-width: 100%;
    max-height: 90vh;
  }

  .current-avatar-compact {
    width: 80px;
    height: 80px;
  }
}
</style>
