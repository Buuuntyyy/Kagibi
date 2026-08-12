<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<!-- Demande de fichiers : crée un lien de dépôt seul sur un dossier. -->
<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="close">
    <div class="modal-content">
      <h3>{{ t('fileRequest.dialogTitle', { name: item?.name || '' }) }}</h3>

      <div v-if="!generatedLink" class="request-form">
        <p class="request-hint">{{ t('fileRequest.dialogHint') }}</p>

        <div class="form-group">
          <label>{{ t('fileRequest.labelField') }}</label>
          <input
            type="text"
            v-model="label"
            class="text-input"
            maxlength="200"
            :placeholder="t('fileRequest.labelPlaceholder')"
          />
        </div>

        <div class="form-group">
          <label>{{ t('fileRequest.expiration') }}</label>
          <input type="datetime-local" v-model="expiresAt" class="text-input" />
        </div>

        <div class="form-group">
          <label>{{ t('fileRequest.password') }}</label>
          <input
            type="text"
            v-model="password"
            class="text-input"
            autocomplete="off"
            :placeholder="t('fileRequest.passwordPlaceholder')"
          />
        </div>

        <div class="modal-actions">
          <button @click="close">{{ t('fileRequest.cancel') }}</button>
          <button @click="generateLink" class="btn-primary" :disabled="loading">
            {{ loading ? t('common.loading') : t('fileRequest.generate') }}
          </button>
        </div>
      </div>

      <div v-else class="request-result">
        <p v-if="isExisting" class="existing-note">{{ t('fileRequest.existingNote') }}</p>
        <p v-else>{{ t('fileRequest.linkReady') }}</p>
        <div class="link-display">
          <input type="text" :value="generatedLink" readonly />
          <button @click="copyLink" class="btn-copy">
            <span v-if="copied">{{ t('fileRequest.copied') }}</span>
            <span v-else>{{ t('fileRequest.copy') }}</span>
          </button>
        </div>

        <div class="modal-actions">
          <button v-if="shareId" class="btn-danger" @click="revoke" :disabled="loading">
            {{ t('fileRequest.revoke') }}
          </button>
          <button @click="close" class="btn-primary">{{ t('fileRequest.close') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useFileStore } from '../stores/files'
import { useUIStore } from '../stores/ui'
import api from '../api'

const { t } = useI18n()

const props = defineProps({
  isOpen: Boolean,
  item: Object, // dossier : { id, name }
})

const emit = defineEmits(['close'])

const fileStore = useFileStore()
const uiStore = useUIStore()

const label = ref('')
const expiresAt = ref('')
const password = ref('')
const generatedLink = ref('')
const shareId = ref(null)
const isExisting = ref(false)
const loading = ref(false)
const copied = ref(false)

watch(() => props.isOpen, (open) => {
  if (open) {
    label.value = ''
    expiresAt.value = ''
    password.value = ''
    generatedLink.value = ''
    shareId.value = null
    isExisting.value = false
    copied.value = false
    loading.value = false
  }
})

const close = () => emit('close')

const generateLink = async () => {
  if (!props.item) return
  loading.value = true
  try {
    const expiration = expiresAt.value ? new Date(expiresAt.value).toISOString() : null
    const result = await fileStore.createFileRequestLink(props.item.id, {
      expiresAt: expiration,
      password: password.value,
      label: label.value,
    })
    generatedLink.value = `${window.location.origin}/r/${result.token}`
    shareId.value = result.id || null
    isExisting.value = !!result.existing
  } catch (e) {
    console.error('Failed to create file request link', e)
    uiStore.showToast(t('fileRequest.createError'), 'error')
  } finally {
    loading.value = false
  }
}

const revoke = async () => {
  if (!shareId.value) return
  const confirmed = await uiStore.showConfirm({
    title: t('fileRequest.revoke'),
    message: t('fileRequest.confirmRevoke'),
    confirmLabel: t('fileRequest.revoke'),
  })
  if (!confirmed) return
  loading.value = true
  try {
    await api.delete(`/shares/link/${shareId.value}`)
    uiStore.showToast(t('fileRequest.revoked'))
    generatedLink.value = ''
    shareId.value = null
    isExisting.value = false
  } catch (e) {
    console.error('Failed to revoke file request link', e)
    uiStore.showToast(t('fileRequest.revokeError'), 'error')
  } finally {
    loading.value = false
  }
}

const copyLink = () => {
  navigator.clipboard.writeText(generatedLink.value).then(() => {
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  })
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: var(--card-color, #fff);
  color: var(--main-text-color, #222);
  border: 1px solid var(--border-color, #ddd);
  border-radius: 12px;
  padding: 22px;
  width: min(440px, calc(100vw - 32px));
}

.modal-content h3 {
  margin: 0 0 10px;
  font-size: 1.05rem;
}

.request-hint {
  font-size: 0.82rem;
  color: var(--subtle-text-color, #8a8a8a);
  margin: 0 0 14px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 12px;
}

.form-group label {
  font-size: 0.8rem;
  color: var(--subtle-text-color, #8a8a8a);
}

.text-input {
  padding: 8px 10px;
  border-radius: 8px;
  border: 1px solid var(--border-color, #ddd);
  background: var(--background-color, #fff);
  color: var(--main-text-color, #222);
  font-size: 0.85rem;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 16px;
}

.modal-actions button {
  padding: 7px 14px;
  border-radius: 8px;
  border: 1px solid var(--border-color, #ddd);
  background: var(--card-color, #fff);
  color: var(--main-text-color, #222);
  cursor: pointer;
  font-size: 0.85rem;
}

.btn-primary {
  background: var(--primary-color, #4f6ef7) !important;
  border-color: var(--primary-color, #4f6ef7) !important;
  color: #fff !important;
}

.btn-danger {
  color: #e05555 !important;
  border-color: rgba(224, 85, 85, 0.4) !important;
}

.existing-note {
  font-size: 0.82rem;
  color: var(--subtle-text-color, #8a8a8a);
}

.link-display {
  display: flex;
  gap: 8px;
  margin-top: 8px;
}

.link-display input {
  flex: 1;
  padding: 8px 10px;
  border-radius: 8px;
  border: 1px solid var(--border-color, #ddd);
  background: var(--background-color, #fff);
  color: var(--main-text-color, #222);
  font-size: 0.8rem;
}

.btn-copy {
  padding: 7px 12px;
  border-radius: 8px;
  border: 1px solid var(--border-color, #ddd);
  background: var(--card-color, #fff);
  color: var(--main-text-color, #222);
  cursor: pointer;
  white-space: nowrap;
}
</style>
