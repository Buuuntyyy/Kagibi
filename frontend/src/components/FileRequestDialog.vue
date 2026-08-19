<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<!-- Demande de fichiers : crée un lien de dépôt seul sur un dossier. -->
<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="close">
    <div class="modal-content">
      <h3>{{ t('fileRequest.dialogTitle', { name: item?.name || '' }) }}</h3>

      <div v-if="checkingExisting" class="request-form">
        <p class="request-hint">{{ t('common.loading') }}</p>
      </div>

      <div v-else-if="!generatedLink" class="request-form">
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

        <div class="form-group">
          <label>{{ t('fileRequest.recipientEmail') }} <span class="optional-tag">{{ t('common.optional') }}</span></label>
          <input
            type="email"
            v-model="recipientEmail"
            class="text-input"
            :placeholder="t('fileRequest.emailPlaceholder')"
          />
          <p class="field-hint">{{ t('fileRequest.emailHint') }}</p>
        </div>

        <label class="send-email-toggle">
          <input type="checkbox" v-model="sendEmailOption" />
          <span>{{ t('fileRequest.sendEmailOption') }}</span>
        </label>

        <div v-if="sendEmailOption" class="email-lang-row">
          <span class="email-lang-label">{{ t('fileRequest.emailLang') }}</span>
          <div class="lang-toggle">
            <button
              class="lang-btn"
              :class="{ active: emailLang === 'fr' }"
              @click="emailLang = 'fr'"
              type="button"
            >🇫🇷 Français</button>
            <button
              class="lang-btn"
              :class="{ active: emailLang === 'en' }"
              @click="emailLang = 'en'"
              type="button"
            >🇬🇧 English</button>
          </div>
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

        <div v-if="isExisting" class="existing-details">
          <p v-if="existingRequestLabel">{{ t('fileRequest.labelField') }} : « {{ existingRequestLabel }} »</p>
          <p v-if="existingExpiresAt">{{ t('fileRequest.expiration') }} : {{ new Date(existingExpiresAt).toLocaleString() }}</p>
          <p v-if="existingHasPassword">🔒 {{ t('fileRequest.password') }}</p>
        </div>

        <div class="link-display">
          <input type="text" :value="generatedLink" readonly />
          <button @click="copyLink" class="btn-copy">
            <span v-if="copied">{{ t('fileRequest.copied') }}</span>
            <span v-else>{{ t('fileRequest.copy') }}</span>
          </button>
        </div>

        <p v-if="emailSent" class="email-sent-notice">{{ t('fileRequest.emailSent') }}</p>

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
const recipientEmail = ref('')
const sendEmailOption = ref(false)
const emailLang = ref('fr')
const emailSent = ref(false)
const generatedLink = ref('')
const shareId = ref(null)
const isExisting = ref(false)
const loading = ref(false)
const copied = ref(false)
const checkingExisting = ref(false)
const existingRequestLabel = ref('')
const existingExpiresAt = ref(null)
const existingHasPassword = ref(false)

watch(() => props.isOpen, async (open) => {
  if (!open) return

  label.value = ''
  expiresAt.value = ''
  password.value = ''
  recipientEmail.value = ''
  sendEmailOption.value = false
  emailLang.value = 'fr'
  emailSent.value = false
  generatedLink.value = ''
  shareId.value = null
  isExisting.value = false
  copied.value = false
  loading.value = false
  existingRequestLabel.value = ''
  existingExpiresAt.value = null
  existingHasPassword.value = false

  if (!props.item) return

  // Un lien de demande de fichiers existe peut-être déjà pour ce dossier — on le
  // vérifie avant d'afficher le formulaire de création, pour l'afficher directement
  // au lieu de forcer l'utilisateur à tenter une création qui échouerait avec un 409.
  checkingExisting.value = true
  try {
    const existing = await fileStore.getExistingFileRequestLink(props.item.id)
    if (existing) {
      generatedLink.value = `${window.location.origin}/r/${existing.token}`
      shareId.value = existing.id
      isExisting.value = true
      existingRequestLabel.value = existing.request_label || ''
      existingExpiresAt.value = existing.expires_at || null
      existingHasPassword.value = !!existing.has_password
    }
  } catch (e) {
    console.error('Failed to check for an existing file request link', e)
  } finally {
    checkingExisting.value = false
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
      recipientEmail: recipientEmail.value.trim(),
      sendEmail: sendEmailOption.value,
      emailLang: emailLang.value,
    })
    generatedLink.value = `${window.location.origin}/r/${result.token}`
    shareId.value = result.id || null
    isExisting.value = !!result.existing
    emailSent.value = sendEmailOption.value && !!recipientEmail.value.trim()
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
    existingRequestLabel.value = ''
    existingExpiresAt.value = null
    existingHasPassword.value = false
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

.existing-details {
  margin-top: 6px;
  padding: 8px 10px;
  border-radius: 8px;
  background: var(--background-color, #f5f5f5);
  border: 1px solid var(--border-color, #e2e2e2);
}

.existing-details p {
  margin: 0;
  font-size: 0.82rem;
  color: var(--main-text-color, #333);
}

.existing-details p + p {
  margin-top: 4px;
}

.optional-tag {
  font-weight: 400;
  font-size: 0.75rem;
  color: var(--subtle-text-color, #8a8a8a);
  opacity: 0.8;
  margin-left: 0.2rem;
}

.field-hint {
  font-size: 0.75rem;
  color: var(--subtle-text-color, #8a8a8a);
  margin: 4px 0 0;
  opacity: 0.85;
}

.send-email-toggle {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.85rem;
  color: var(--subtle-text-color, #8a8a8a);
  cursor: pointer;
  margin-bottom: 12px;
  user-select: none;
}

.send-email-toggle input[type="checkbox"] {
  width: 15px;
  height: 15px;
  cursor: pointer;
  accent-color: var(--primary-color, #4f6ef7);
}

.email-lang-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: -6px 0 12px;
}

.email-lang-label {
  font-size: 0.8rem;
  color: var(--subtle-text-color, #8a8a8a);
  white-space: nowrap;
}

.lang-toggle {
  display: flex;
  gap: 5px;
}

.lang-btn {
  padding: 4px 10px;
  border: 1px solid var(--border-color, #ddd);
  border-radius: 6px;
  background: transparent;
  color: var(--subtle-text-color, #8a8a8a);
  font-size: 0.78rem;
  cursor: pointer;
  transition: all 0.15s;
}

.lang-btn.active {
  border-color: var(--primary-color, #4f6ef7);
  background: var(--primary-color, #4f6ef7);
  color: #fff;
}

.email-sent-notice {
  font-size: 0.82rem;
  color: #2e9e5b;
  margin: 10px 0 0;
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
