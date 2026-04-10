<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <transition name="delete-dialog-fade">
    <div
      v-if="modelValue"
      class="delete-dialog-overlay"
      @click.self="handleClose"
      role="presentation"
    >
      <div
        class="delete-dialog"
        role="dialog"
        aria-modal="true"
        aria-labelledby="delete-dialog-title"
      >
        <div class="delete-dialog-header">
          <svg class="warning-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
            <line x1="12" y1="9" x2="12" y2="13" />
            <line x1="12" y1="17" x2="12.01" y2="17" />
          </svg>
          <h3 id="delete-dialog-title">{{ t('dialogs.deleteAccount.title') }}</h3>
          <button class="btn-close" type="button" @click="handleClose" :disabled="isDeletingAccount">×</button>
        </div>
        <div class="delete-dialog-body">
          <p><strong>{{ t('dialogs.deleteAccount.warning') }}</strong></p>
          <p><strong>{{ t('dialogs.deleteAccount.description') }}</strong></p>
          <ul>
            <li>{{ t('dialogs.deleteAccount.point1') }}</li>
            <li>{{ t('dialogs.deleteAccount.point2') }}</li>
            <li>{{ t('dialogs.deleteAccount.point3') }}</li>
            <li>{{ t('dialogs.deleteAccount.point4') }}</li>
          </ul>
          <div class="critical-warning">
            <p><strong>{{ t('dialogs.deleteAccount.criticalWarning') }}</strong></p>
          </div>
          <div class="confirmation-section">
            <label for="confirmInput">
              {{ t('dialogs.deleteAccount.confirmText') }} <code>{{ t('dialogs.deleteAccount.deleteMyAccount') }}</code> :
            </label>
            <input
              id="confirmInput"
              v-model="confirmationText"
              type="text"
              :placeholder="t('dialogs.deleteAccount.deleteMyAccount')"
              class="confirmation-input"
              autocomplete="off"
            />
          </div>
        </div>
        <div class="delete-dialog-footer">
          <button class="btn-secondary" type="button" @click="handleClose" :disabled="isDeletingAccount">
            {{ t('dialogs.deleteAccount.cancel') }}
          </button>
          <button
            class="btn-danger"
            type="button"
            @click="$emit('confirm')"
            :disabled="confirmationText !== t('dialogs.deleteAccount.deleteMyAccount') || isDeletingAccount"
          >
            <span v-if="isDeletingAccount">{{ t('dialogs.deleteAccount.deleting') }}</span>
            <span v-else>{{ t('dialogs.deleteAccount.deleteButton') }}</span>
          </button>
        </div>
      </div>
    </div>
  </transition>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = defineProps({
  modelValue: {
    type: Boolean,
    required: true
  },
  deleteConfirmationText: {
    type: String,
    default: ''
  },
  isDeletingAccount: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue', 'update:deleteConfirmationText', 'confirm', 'close'])

const confirmationText = computed({
  get: () => props.deleteConfirmationText,
  set: (value) => emit('update:deleteConfirmationText', value)
})

const handleClose = () => {
  if (!props.isDeletingAccount) {
    emit('update:modelValue', false)
    emit('close')
  }
}
</script>

<style scoped>
.delete-dialog-overlay {
  position: fixed;
  inset: 0;
  background-color: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.delete-dialog {
  background-color: var(--card-color);
  border-radius: 12px;
  max-width: 520px;
  width: 100%;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  border: 1px solid var(--border-color);
  animation: deleteDialogSlideUp 0.3s ease;
}

.delete-dialog-header {
  padding: 1.5rem 2rem;
  border-bottom: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  gap: 1rem;
}

.warning-icon {
  width: 48px;
  height: 48px;
  color: var(--error-color);
  flex-shrink: 0;
}

.delete-dialog-header h3 {
  margin: 0;
  color: var(--main-text-color);
  flex: 1;
  font-size: 1.1rem;
}

.btn-close {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: var(--secondary-text-color);
  padding: 0;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: color 0.2s;
}

.btn-close:hover {
  color: var(--main-text-color);
}

.delete-dialog-body {
  padding: 1.5rem 2rem;
  color: var(--main-text-color);
}

.delete-dialog-body p {
  margin-bottom: 1rem;
}

.delete-dialog-body ul {
  margin: 1rem 0;
  padding-left: 1.5rem;
}

.delete-dialog-body li {
  margin-bottom: 0.5rem;
  line-height: 1.5;
}

.critical-warning {
  background: rgba(220, 53, 69, 0.05);
  border: 2px solid #dc3545;
  border-radius: 8px;
  padding: 1rem;
  margin: 1.5rem 0;
}

.critical-warning p {
  color: #c53030;
  font-weight: 600;
  margin: 0.5rem 0;
}

.confirmation-section {
  margin-top: 1.5rem;
}

.confirmation-section label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 600;
  color: var(--main-text-color);
}

.confirmation-section code {
  background-color: rgba(220, 53, 69, 0.2);
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  color: #c53030;
  font-weight: bold;
  font-size: 0.875rem;
}

.confirmation-input {
  width: 100%;
  padding: 0.75rem;
  border: 2px solid var(--border-color);
  border-radius: 6px;
  font-size: 1rem;
  background-color: var(--background-color);
  color: var(--main-text-color);
  transition: border-color 0.2s;
}

.confirmation-input:focus {
  outline: none;
  border-color: #dc3545;
}

.delete-dialog-footer {
  padding: 1.5rem 2rem;
  border-top: 1px solid var(--border-color);
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
}

.btn-secondary {
  background-color: transparent;
  border: 1px solid var(--border-color);
  color: var(--main-text-color);
  padding: 0.8rem 1.5rem;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 500;
}

.btn-secondary:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
}

.btn-danger {
  background-color: #dc3545;
  color: white;
  border: none;
  padding: 0.8rem 1.5rem;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 600;
  transition: all 0.2s;
}

.btn-danger:hover:not(:disabled) {
  background-color: #c82333;
}

.btn-danger:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.delete-dialog-fade-enter-active {
  animation: deleteDialogFadeIn 0.2s ease;
}

.delete-dialog-fade-leave-active {
  animation: deleteDialogFadeOut 0.2s ease;
}

@keyframes deleteDialogFadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes deleteDialogFadeOut {
  from { opacity: 1; }
  to { opacity: 0; }
}

@keyframes deleteDialogSlideUp {
  from { opacity: 0; transform: translateY(16px); }
  to { opacity: 1; transform: translateY(0); }
}

@media (max-width: 600px) {
  .delete-dialog-footer {
    flex-direction: column-reverse;
    gap: 0.75rem;
  }

  .btn-secondary,
  .btn-danger {
    width: 100%;
  }
}
</style>
