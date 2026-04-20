<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <Teleport to="body">
    <div v-if="visible" class="invite-overlay" @click.self="close">
      <div class="invite-dialog">

        <!-- Step 1: email + options -->
        <template v-if="step === 'form'">
          <h3 class="invite-title">{{ t('p2p.invite.title') }}</h3>
          <p class="invite-sub">{{ t('p2p.invite.subtitle') }}</p>

          <div class="invite-field">
            <label>{{ t('p2p.invite.recipientEmail') }} <span class="optional-tag">{{ t('common.optional') }}</span></label>
            <input
              v-model="email"
              type="email"
              :placeholder="t('p2p.invite.emailPlaceholder')"
              class="invite-input"
              @keyup.enter="create"
            />
            <p class="field-hint">{{ t('p2p.invite.emailHint') }}</p>
          </div>

          <div class="invite-field">
            <label>{{ t('p2p.invite.file') }}</label>
            <div class="file-summary">
              <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"/><polyline points="13 2 13 9 20 9"/></svg>
              <span class="file-name">{{ props.file?.name }}</span>
              <span class="file-size">{{ formatSize(props.file?.size) }}</span>
            </div>
          </div>

          <label class="send-email-toggle">
            <input type="checkbox" v-model="sendEmailOption" />
            <span>{{ t('p2p.invite.sendEmailOption') }}</span>
          </label>

          <label class="legal-consent-toggle">
            <input type="checkbox" v-model="legalConsent" />
            <span>
              Je certifie que le fichier partagé est légal, ne contient pas de contenu illicite, et j'accepte d'en être seul responsable conformément aux <router-link to="/terms" target="_blank" class="legal-link">CGU</router-link>.
            </span>
          </label>

          <p v-if="error" class="invite-error">{{ error }}</p>

          <div class="invite-actions">
            <button class="btn-secondary" @click="close" :disabled="loading">{{ t('common.cancel') }}</button>
            <button class="btn-primary" @click="create" :disabled="loading || !legalConsent">
              <span v-if="loading" class="spinner"></span>
              {{ loading ? t('common.loading') : t('p2p.invite.generate') }}
            </button>
          </div>
        </template>

        <!-- Step 2: link ready -->
        <template v-else-if="step === 'ready'">
          <h3 class="invite-title">{{ t('p2p.invite.readyTitle') }}</h3>
          <p class="invite-sub">{{ t('p2p.invite.readyDesc', { name: recipientName }) }}</p>

          <div class="link-box">
            <input class="link-input" :value="inviteUrl" readonly @click="copyLink" />
            <button class="copy-btn" @click="copyLink">
              <svg v-if="!copied" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
              <svg v-else xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"/></svg>
              {{ copied ? t('common.copied') : t('common.copy') }}
            </button>
          </div>

          <p v-if="emailSent" class="email-sent-notice">
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"/></svg>
            {{ t('p2p.invite.emailSent') }}
          </p>

          <div class="waiting-notice">
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
            {{ t('p2p.invite.waitingNotice') }}
          </div>

          <div class="invite-actions" style="justify-content: flex-end">
            <button class="btn-secondary" @click="close">{{ t('common.close') }}</button>
          </div>
        </template>

      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { authClient } from '../../auth-client'
import { API_BASE_URL } from '../../api'
import { useP2PStore } from '../../stores/p2p'
import { generateMasterKey } from '../../utils/crypto'

const { t } = useI18n()
const p2pStore = useP2PStore()

const props = defineProps({
  visible: { type: Boolean, default: false },
  file: { type: File, default: null },
  transferId: { type: String, default: '' },
})

const emit = defineEmits(['close', 'invited'])

const step = ref('form')
const email = ref('')
const error = ref('')
const loading = ref(false)
const copied = ref(false)
const emailSent = ref(false)
const sendEmailOption = ref(false)
const legalConsent = ref(false)

const inviteToken = ref('')
const recipientName = ref('')
const recipientId = ref('')
const recipientPublicKey = ref('')

const inviteUrl = computed(() => {
  if (!inviteToken.value) return ''
  const base = window.location.origin
  // Invite recipients land on the root path of the P2P subdomain (P2PSubdomainView),
  // not on /p2p which is the main-domain P2PView.
  return `${base}/?invite=${inviteToken.value}`
})

function formatSize(bytes) {
  if (!bytes) return ''
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`
}

async function create() {
  if (!props.file) return
  error.value = ''
  loading.value = true
  let _step = 'init'
  try {
    _step = 'getToken'
    const token = await authClient.getToken()
    console.log('[Invite] token present:', !!token)

    _step = 'fetch'
    const res = await fetch(`${API_BASE_URL}p2p/invite`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      body: JSON.stringify({
        recipient_email: email.value.trim(),
        file_name: props.file.name,
        file_size: props.file.size,
        transfer_id: props.transferId,
        send_email: sendEmailOption.value,
      }),
    })

    _step = 'json'
    let data
    try { data = await res.json() } catch (je) {
      console.error('[Invite] JSON parse failed:', je)
      data = {}
    }
    console.log('[Invite] HTTP', res.status, data)

    if (!res.ok) {
      error.value = data.error || t('p2p.invite.createError')
      return
    }

    _step = 'storeFields'
    inviteToken.value = data.token
    recipientName.value = data.recipient_name || 'Guest'
    recipientId.value = data.recipient_id
    recipientPublicKey.value = data.recipient_public_key || null

    _step = 'generateKey'
    const fileKey = await generateMasterKey()

    _step = 'setPendingInvite'
    p2pStore.setPendingInvite({
      transferId: props.transferId,
      file: props.file,
      fileKey,
      recipientId: data.recipient_id,
      recipientPublicKey: data.recipient_public_key || null,
      recipientName: data.recipient_name || 'Guest',
    })

    emailSent.value = sendEmailOption.value
    step.value = 'ready'
    emit('invited', { token: data.token, recipientId: data.recipient_id })
  } catch (e) {
    console.error(`[Invite] create failed at step="${_step}":`, e.name, e.message, e)
    error.value = `[${_step}] ${e.message}`
  } finally {
    loading.value = false
  }
}

async function copyLink() {
  try {
    await navigator.clipboard.writeText(inviteUrl.value)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch {
    // fallback: select input
  }
}

function close() {
  if (step.value === 'form') {
    step.value = 'form'
    email.value = ''
    error.value = ''
  }
  emit('close')
}
</script>

<style scoped>
.invite-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.invite-dialog {
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 14px;
  padding: 2rem;
  width: 440px;
  max-width: calc(100vw - 2rem);
  box-shadow: 0 8px 40px rgba(0,0,0,0.15);
}

.invite-title {
  margin: 0 0 0.4rem;
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--text-color);
}

.invite-sub {
  margin: 0 0 1.4rem;
  color: var(--text-secondary);
  font-size: 0.88rem;
}

.invite-field {
  margin-bottom: 1.1rem;
}

.invite-field label {
  display: block;
  font-size: 0.82rem;
  font-weight: 500;
  color: var(--text-secondary);
  margin-bottom: 0.4rem;
}

.optional-tag {
  font-weight: 400;
  font-size: 0.75rem;
  color: var(--text-secondary);
  opacity: 0.7;
  margin-left: 0.3rem;
}

.field-hint {
  font-size: 0.75rem;
  color: var(--text-secondary);
  margin: 0.3rem 0 0;
  opacity: 0.8;
}

.invite-input {
  width: 100%;
  box-sizing: border-box;
  padding: 0.6rem 0.8rem;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--background-color);
  color: var(--text-color);
  font-size: 0.9rem;
}

.invite-input:focus {
  outline: none;
  border-color: var(--primary-color);
}

.file-summary {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.8rem;
  background: var(--background-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  font-size: 0.88rem;
  color: var(--text-color);
}

.file-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-size {
  color: var(--text-secondary);
  white-space: nowrap;
}

.invite-error {
  color: var(--error-color, #ef4444);
  font-size: 0.85rem;
  margin: -0.5rem 0 0.8rem;
}

.invite-actions {
  display: flex;
  gap: 0.6rem;
  justify-content: flex-end;
  margin-top: 1.2rem;
}

.btn-primary {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.55rem 1.1rem;
  background: var(--primary-color);
  color: #fff;
  border: none;
  border-radius: 8px;
  font-size: 0.88rem;
  font-weight: 500;
  cursor: pointer;
}

.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }

.btn-secondary {
  padding: 0.55rem 1.1rem;
  background: transparent;
  border: 1px solid var(--border-color);
  color: var(--text-color);
  border-radius: 8px;
  font-size: 0.88rem;
  cursor: pointer;
}

.link-box {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 1rem;
}

.link-input {
  flex: 1;
  padding: 0.55rem 0.8rem;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--background-color);
  color: var(--text-color);
  font-size: 0.82rem;
  cursor: pointer;
}

.copy-btn {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.55rem 0.9rem;
  background: var(--primary-color);
  color: #fff;
  border: none;
  border-radius: 8px;
  font-size: 0.82rem;
  cursor: pointer;
  white-space: nowrap;
}

.send-email-toggle {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.85rem;
  color: var(--text-secondary);
  cursor: pointer;
  margin-bottom: 1rem;
  user-select: none;
}

.send-email-toggle input[type="checkbox"] {
  width: 15px;
  height: 15px;
  cursor: pointer;
  accent-color: var(--primary-color);
}

.legal-consent-toggle {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  font-size: 0.82rem;
  color: var(--text-secondary);
  cursor: pointer;
  margin-bottom: 1rem;
  user-select: none;
  line-height: 1.4;
  padding: 0.6rem 0.8rem;
  background: var(--background-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
}

.legal-consent-toggle input[type="checkbox"] {
  width: 15px;
  height: 15px;
  cursor: pointer;
  accent-color: var(--primary-color);
  flex-shrink: 0;
  margin-top: 0.15rem;
}

.legal-link {
  color: var(--primary-color);
  text-decoration: underline;
}

.email-sent-notice {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.82rem;
  color: var(--success-color, #27ae60);
  margin-bottom: 0.75rem;
}

.waiting-notice {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.8rem;
  color: var(--text-secondary);
  background: var(--background-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 0.6rem 0.8rem;
  margin-bottom: 0.5rem;
}

.spinner {
  display: inline-block;
  width: 12px;
  height: 12px;
  border: 2px solid rgba(255,255,255,0.3);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin { to { transform: rotate(360deg); } }
</style>
