<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div v-if="isOpen" class="modal-overlay">
    <div class="modal-content">
      <h3>{{ t('dialogs.share.title', { name: item?.name || '' }) }}</h3>
      
      <div v-if="!generatedLink" class="share-form">
        <div class="form-group">
          <label>
            <input type="text" />
            {{ t('dialogs.share.expiration') }}
          </label>
          <input type="datetime-local" v-model="expiresAt" class="date-input" />
        </div>
        
        <div class="modal-actions">
          <button @click="close">{{ t('dialogs.share.cancel') }}</button>
          <button @click="generateLink" class="btn-primary" :disabled="loading">
            {{ loading ? t('common.loading') : t('dialogs.share.generate') }}
          </button>
        </div>
      </div>

      <div v-else class="share-result">
        <p>{{ t('dialogs.share.shareLink') }} :</p>
        <div class="link-display">
          <input type="text" :value="generatedLink" readonly ref="linkInput" />
          <button @click="copyLink" class="btn-copy">
            <span v-if="copied">{{ t('dialogs.share.copied') }}</span>
            <span v-else>{{ t('dialogs.share.copy') }}</span>
          </button>
        </div>
        
        <div class="modal-actions">
          <button @click="close" class="btn-primary">{{ t('dialogs.share.close') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useFileStore } from '../stores/files'

const { t } = useI18n()

const props = defineProps({
  isOpen: Boolean,
  item: Object // { id, type, name }
})

const emit = defineEmits(['close'])

const fileStore = useFileStore()
const expiresAt = ref('')
const generatedLink = ref('')
const loading = ref(false)
const copied = ref(false)

watch(() => props.isOpen, (newVal) => {
  if (newVal) {
    // Reset state when opening
    expiresAt.value = ''
    generatedLink.value = ''
    copied.value = false
    loading.value = false
  }
})

const close = () => {
  emit('close')
}

const generateLink = async () => {
  if (!props.item) return
  
  loading.value = true
  try {
    // Convert local datetime to ISO string if present
    let expiration = null
    if (expiresAt.value) {
      expiration = new Date(expiresAt.value).toISOString()
    }
    
    const result = await fileStore.createShareLink(props.item.id, props.item.type, expiration)
    
    // Construct full URL
    // Assuming the frontend is served at the same origin or we know the base URL
    // For now, let's assume window.location.origin + '/s/' + token
    generatedLink.value = `${window.location.origin}/s/${result.token}`
    
  } catch (error) {
    console.error("Failed to generate link", error)
    alert("Erreur lors de la génération du lien")
  } finally {
    loading.value = false
  }
}

const copyLink = () => {
  navigator.clipboard.writeText(generatedLink.value).then(() => {
    copied.value = true
    setTimeout(() => {
      copied.value = false
    }, 2000)
  })
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal-content {
  background-color: white;
  padding: 20px;
  border-radius: 8px;
  width: 400px;
  max-width: 90%;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
}

h3 {
  margin-top: 0;
  margin-bottom: 20px;
  color: #333;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
}

.date-input {
  width: 100%;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
  box-sizing: border-box;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 20px;
}

button {
  padding: 8px 16px;
  border: 1px solid var(--border-color);
  background-color: var(--card-color);
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.2s;
  color: var(--main-text-color);
}

button:hover {
  background-color: var(--hover-background-color);
}

.btn-primary {
  background-color: var(--primary-color);
  color: white;
  border: none;
}

.btn-primary:hover {
  background-color: var(--accent-color);
}

.btn-primary:disabled {
  background-color: var(--border-color);
  cursor: not-allowed;
}

.link-display {
  display: flex;
  gap: 10px;
  margin-bottom: 10px;
}

.link-display input {
  flex: 1;
  padding: 8px;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  background-color: var(--hover-background-color);
  color: var(--main-text-color);
}

.btn-copy {
  min-width: 80px;
}
</style>
