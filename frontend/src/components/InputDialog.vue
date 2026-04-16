<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="cancel">
    <div class="modal-content">
      <h3>{{ title }}</h3>
      <input
        ref="inputRef"
        v-model="inputValue"
        :placeholder="placeholder"
        :class="['modal-input', { 'input-error': validationError }]"
        @keyup.enter="confirm"
        @keyup.esc="cancel"
      />
      <p v-if="validationError" class="error-label">{{ validationError }}</p>
      <div class="modal-actions">
        <button @click="cancel" class="btn-cancel">{{ t('common.cancel') }}</button>
        <button @click="confirm" class="btn-confirm" :disabled="!!validationError">{{ t('common.confirm') }}</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps({
  isOpen: Boolean,
  title: String,
  defaultValue: String,
  placeholder: String
})

const emit = defineEmits(['update:isOpen', 'confirm', 'cancel'])
const { t } = useI18n()

const inputValue = ref('')
const inputRef = ref(null)

// Mirrors the backend regex: Unicode letters, numbers, spaces, - . _
// The `u` flag enables Unicode property escapes (\p{L}, \p{N}).
const VALID_NAME_RE = /^[\p{L}\p{N}\s\-\._]+$/u

const validationError = computed(() => {
  const v = inputValue.value
  if (!v.trim()) return null // empty is handled by backend required binding
  if (!VALID_NAME_RE.test(v)) return t('dialogs.rename.invalidName')
  return null
})

watch(() => props.isOpen, (newVal) => {
  if (newVal) {
    inputValue.value = props.defaultValue || ''
    nextTick(() => {
      inputRef.value?.focus()
      inputRef.value?.select()
    })
  }
})

const confirm = () => {
  if (validationError.value) return
  emit('confirm', inputValue.value)
  emit('update:isOpen', false)
}

const cancel = () => {
  emit('cancel')
  emit('update:isOpen', false)
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 2000;
}

.modal-content {
  background: white;
  padding: 20px;
  border-radius: 8px;
  width: 400px;
  box-shadow: 0 4px 15px rgba(0, 0, 0, 0.2);
}

h3 {
  margin-top: 0;
  margin-bottom: 15px;
  color: #333;
}

.modal-input {
  width: 100%;
  padding: 10px;
  margin-bottom: 6px;
  border: 1px solid #ccc;
  border-radius: 4px;
  box-sizing: border-box;
  font-size: 1rem;
  transition: border-color 0.2s;
}

.modal-input.input-error {
  border-color: #e53935;
  outline-color: #e53935;
}

.error-label {
  margin: 0 0 14px 0;
  font-size: 0.82rem;
  color: #e53935;
  line-height: 1.4;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

button {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 500;
}

.btn-cancel {
  background-color: #f0f0f0;
  color: #333;
}

.btn-cancel:hover {
  background-color: #e0e0e0;
}

.btn-confirm {
  background-color: #42b983;
  color: white;
}

.btn-confirm:hover:not(:disabled) {
  background-color: #3aa876;
}

.btn-confirm:disabled {
  background-color: #a5d6bc;
  cursor: not-allowed;
}
</style>
