<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div v-if="uiStore.deleteDialog.visible" class="modal-overlay" @click.self="cancel">
    <div class="modal-content">
      <div class="modal-header">
        <h3>{{ uiStore.deleteDialog.title || (itemsCount > 1 ? t('common.deleteItems') : t('common.deleteItem')) }}</h3>
        <button @click="cancel" class="btn-close">×</button>
      </div>
      
      <div class="modal-body">
        <p v-if="uiStore.deleteDialog.message">{{ uiStore.deleteDialog.message }}</p>
        <template v-else>
          <p v-if="itemsCount === 1">
            {{ t('messages.confirmDelete', { name: uiStore.deleteDialog.itemName }) }}
          </p>
          <p v-else>
            {{ t('messages.confirmDeleteMultiple', { count: itemsCount }) }}
          </p>
        </template>
        <p class="sub-text">{{ t('messages.irreversible') }}</p>
      </div>

      <div class="modal-footer">
        <button @click="cancel" class="btn-secondary">{{ t('common.cancel') }}</button>
        <button @click="confirm" class="btn-delete">{{ t('common.delete') }}</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useUIStore } from '../stores/ui'

const { t } = useI18n()

const uiStore = useUIStore()

const itemsCount = computed(() => uiStore.deleteDialog.itemsCount || 1)

const confirm = () => {
  if (uiStore.deleteDialog.onConfirm) {
      uiStore.deleteDialog.onConfirm()
  }
  uiStore.closeDeleteDialog()
};

const cancel = () => {
  uiStore.closeDeleteDialog()
};
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.6);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 321000;
  animation: fadeIn 0.2s ease;
  backdrop-filter: blur(2px);
}

.modal-content {
  background: var(--card-color, #1e1e1e);
  padding: 0;
  border-radius: 12px;
  width: 400px;
  max-width: 90%;
  box-shadow: 0 10px 25px rgba(0,0,0,0.3);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid var(--border-color, #333);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  border-bottom: 1px solid var(--border-color, #333);
  background: rgba(255, 82, 82, 0.05); /* Slight red tint for delete context */
}

.modal-header h3 {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--main-text-color, #eee);
}

.btn-close {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: var(--secondary-text-color, #aaa);
  padding: 0;
  line-height: 1;
}

.modal-body {
  padding: 24px;
  text-align: center;
  color: var(--main-text-color, #eee);
}

.sub-text {
  color: var(--secondary-text-color, #aaa);
  font-size: 0.9rem;
  margin-top: 1rem;
}

.modal-footer {
  padding: 16px 24px;
  border-top: 1px solid var(--border-color, #333);
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  background-color: var(--background-color, #121212);
}

button {
  padding: 8px 16px;
  border-radius: 6px;
  border: 1px solid transparent;
  cursor: pointer;
  font-weight: 500;
  font-size: 0.9rem;
  transition: all 0.2s;
}

.btn-secondary {
  background-color: transparent;
  border: 1px solid var(--border-color, #444);
  color: var(--main-text-color, #ccc);
}

.btn-secondary:hover {
  background-color: rgba(255,255,255,0.05);
  border-color: #666;
}

.btn-delete {
  background-color: #f44336; /* Error Red */
  color: white;
  border: 1px solid #d32f2f;
}

.btn-delete:hover {
  background-color: #d32f2f;
  box-shadow: 0 4px 12px rgba(244, 67, 54, 0.3);
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}
</style>
