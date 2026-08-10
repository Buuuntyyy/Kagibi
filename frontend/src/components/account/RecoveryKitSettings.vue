<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<!-- Kit de récupération : statut de vérification, preuve de possession du code,
     régénération d'un nouveau code (MFA requis si activé). -->
<template>
  <div class="recovery-kit">
    <!-- Statut -->
    <div class="status-row">
      <span class="status-badge" :class="verified ? 'ok' : 'warn'">
        <svg v-if="verified" viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/></svg>
        <svg v-else viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M1 21h22L12 2 1 21zm12-3h-2v-2h2v2zm0-4h-2v-4h2v4z"/></svg>
        {{ verified ? t('recoveryKit.statusVerified', { date: verifiedDate }) : t('recoveryKit.statusUnverified') }}
      </span>
    </div>
    <p class="hint">{{ t('recoveryKit.sectionHint') }}</p>

    <!-- Vérification : coller le code -->
    <div v-if="!showNewCode && !showEmailCodeStep" class="verify-block">
      <label class="field-label">{{ t('recoveryKit.verifyLabel') }}</label>
      <form class="verify-row" @submit.prevent="handleVerify">
        <input
          v-model="codeInput"
          type="text"
          class="text-input mono"
          autocomplete="off"
          spellcheck="false"
          :placeholder="t('recoveryKit.verifyPlaceholder')"
        />
        <button type="submit" class="btn-primary" :disabled="verifying || !codeInput.trim()">
          {{ verifying ? t('common.loading') : t('recoveryKit.verifyButton') }}
        </button>
      </form>
      <p v-if="verifyError" class="error-text">{{ t('recoveryKit.verifyError') }}</p>
      <p v-if="justVerified" class="success-text">
        {{ t('recoveryKit.verifySuccess') }}
        <button class="btn-link" @click="downloadKitFromInput">{{ t('recoveryKit.downloadKit') }}</button>
      </p>
    </div>

    <!-- Nouveau code généré -->
    <div v-if="showNewCode" class="new-code-block">
      <div class="alert-box">{{ t('recoveryKit.newCodeWarning') }}</div>
      <div class="code-box">{{ newCode }}</div>
      <div class="new-code-actions">
        <button class="btn-secondary" @click="copyNewCode">{{ copied ? t('recoveryKit.copied') : t('recoveryKit.copy') }}</button>
        <button class="btn-secondary" @click="downloadKitForNewCode">{{ t('recoveryKit.downloadKit') }}</button>
        <button class="btn-primary" @click="dismissNewCode">{{ t('recoveryKit.newCodeDone') }}</button>
      </div>
    </div>

    <!-- Étape : confirmation par email (obligatoire) -->
    <div v-else-if="showEmailCodeStep" class="rotate-block">
      <p class="hint">{{ t('recoveryKit.emailCodeSentHint') }}</p>
      <form class="verify-row" @submit.prevent="submitEmailCode">
        <input
          v-model="emailCodeInput"
          type="text"
          class="text-input mono"
          maxlength="6"
          inputmode="numeric"
          autocomplete="off"
          :placeholder="t('recoveryKit.emailCodePlaceholder')"
          @input="e => emailCodeInput = e.target.value.replace(/\D/g, '')"
        />
        <button type="submit" class="btn-primary" :disabled="rotating || emailCodeInput.length !== 6">
          {{ rotating ? t('common.loading') : t('recoveryKit.emailCodeConfirm') }}
        </button>
      </form>
      <p v-if="rotateError" class="error-text">{{ rotateError }}</p>
      <div class="new-code-actions">
        <button class="btn-secondary" :disabled="requestingCode" @click="handleRotate">
          {{ requestingCode ? t('common.loading') : t('recoveryKit.emailCodeResend') }}
        </button>
        <button class="btn-secondary" @click="cancelRotate">{{ t('recoveryKit.cancel') }}</button>
      </div>
    </div>

    <!-- Régénération -->
    <div v-else class="rotate-block">
      <button class="btn-outline-danger" @click="handleRotate" :disabled="requestingCode">
        {{ requestingCode ? t('common.loading') : t('recoveryKit.rotateButton') }}
      </button>
      <p class="hint">{{ t('recoveryKit.rotateHint') }}</p>
      <p v-if="requestError" class="error-text">{{ requestError }}</p>
    </div>

    <MFAChallengeModal
      v-model="showMFAChallenge"
      context="recovery_change"
      @verified="onMFAVerified"
      @cancelled="pendingAction = null"
    />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../../stores/auth'
import { useUIStore } from '../../stores/ui'
import { useMFA } from '../../utils/useMFA'
import { downloadRecoveryKit } from '../../utils/recoveryKit'
import MFAChallengeModal from '../MFAChallengeModal.vue'

const { t } = useI18n()
const authStore = useAuthStore()
const uiStore = useUIStore()
const { isMFARequired } = useMFA()

const codeInput = ref('')
const verifying = ref(false)
const verifyError = ref(false)
const justVerified = ref(false)

const rotating = ref(false)
const newCode = ref('')
const showNewCode = ref(false)
const copied = ref(false)

const requestingCode = ref(false)
const requestError = ref('')
const showEmailCodeStep = ref(false)
const emailCodeInput = ref('')
const rotateError = ref('')

const showMFAChallenge = ref(false)
const pendingAction = ref(null)

const verified = computed(() => !!authStore.user?.recovery_verified_at)
const verifiedDate = computed(() => {
  const d = authStore.user?.recovery_verified_at
  return d ? new Date(d).toLocaleDateString() : ''
})

async function handleVerify() {
  verifying.value = true
  verifyError.value = false
  justVerified.value = false
  try {
    await authStore.verifyRecoveryBackup(codeInput.value)
    justVerified.value = true
  } catch (e) {
    verifyError.value = true
  } finally {
    verifying.value = false
  }
}

function downloadKitFromInput() {
  downloadRecoveryKit({ code: codeInput.value.trim(), email: authStore.user?.email })
}

// doRotate finalise la rotation : le code de confirmation email a déjà été saisi et
// vérifié côté formulaire (submitEmailCode) ; il ne reste que le step-up MFA éventuel
// (toujours exigé côté serveur si la MFA est activée — cf. middleware/mfa.go) avant
// d'envoyer la requête effective.
async function doRotate() {
  rotating.value = true
  rotateError.value = ''
  try {
    newCode.value = await authStore.rotateRecoveryCode(emailCodeInput.value)
    showNewCode.value = true
    showEmailCodeStep.value = false
    justVerified.value = false
    codeInput.value = ''
    emailCodeInput.value = ''
  } catch (e) {
    rotateError.value = e.response?.data?.error || e.message || t('recoveryKit.rotateError')
  } finally {
    rotating.value = false
  }
}

// handleRotate déclenche l'envoi du code de confirmation par email — première étape
// obligatoire de la rotation, avant même le step-up MFA (qui n'intervient qu'à l'appel
// final /auth/recovery/rotate, une fois le code email saisi).
async function handleRotate() {
  if (!showEmailCodeStep.value) {
    const confirmed = await uiStore.showConfirm({
      title: t('recoveryKit.rotateButton'),
      message: t('recoveryKit.confirmRotate'),
      confirmLabel: t('recoveryKit.rotateConfirmLabel'),
    })
    if (!confirmed) return
  }

  requestingCode.value = true
  requestError.value = ''
  try {
    await authStore.requestRecoveryRotationEmailCode()
    showEmailCodeStep.value = true
    emailCodeInput.value = ''
    rotateError.value = ''
  } catch (e) {
    requestError.value = e.response?.data?.error || e.message || t('recoveryKit.requestCodeError')
  } finally {
    requestingCode.value = false
  }
}

function cancelRotate() {
  showEmailCodeStep.value = false
  emailCodeInput.value = ''
  rotateError.value = ''
}

async function submitEmailCode() {
  if (emailCodeInput.value.length !== 6) return
  rotateError.value = ''
  try {
    const mfaRequired = await isMFARequired('recovery_change')
    if (mfaRequired) {
      pendingAction.value = doRotate
      showMFAChallenge.value = true
      return
    }
  } catch (_) { /* en cas de doute, tenter directement — le serveur tranchera */ }
  await doRotate()
}

async function onMFAVerified() {
  showMFAChallenge.value = false
  const action = pendingAction.value
  pendingAction.value = null
  if (action) await action()
}

function copyNewCode() {
  navigator.clipboard.writeText(newCode.value).then(() => {
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  })
}

function downloadKitForNewCode() {
  downloadRecoveryKit({ code: newCode.value, email: authStore.user?.email })
}

function dismissNewCode() {
  showNewCode.value = false
  newCode.value = ''
}
</script>

<style scoped>
.recovery-kit {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.status-row {
  display: flex;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  border-radius: 999px;
  font-size: 0.8rem;
  font-weight: 500;
}

.status-badge.ok {
  background: rgba(46, 158, 91, 0.12);
  color: #2e7d32;
}

.status-badge.warn {
  background: rgba(240, 160, 40, 0.14);
  color: #c07b12;
}

.hint {
  font-size: 0.8rem;
  color: var(--subtle-text-color, #8a8a8a);
  margin: 0;
}

.field-label {
  font-size: 0.8rem;
  color: var(--subtle-text-color, #8a8a8a);
  display: block;
  margin-bottom: 4px;
}

.verify-row {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.text-input {
  flex: 1;
  min-width: 220px;
  padding: 8px 10px;
  border-radius: 8px;
  border: 1px solid var(--border-color, #ddd);
  background: var(--background-color, #fff);
  color: var(--main-text-color, #222);
  font-size: 0.85rem;
}

.mono {
  font-family: monospace;
}

.btn-primary {
  padding: 8px 14px;
  border-radius: 8px;
  border: none;
  background: var(--primary-color, #4f6ef7);
  color: #fff;
  cursor: pointer;
  font-size: 0.85rem;
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: default;
}

.btn-secondary {
  padding: 8px 14px;
  border-radius: 8px;
  border: 1px solid var(--border-color, #ddd);
  background: var(--card-color, #fff);
  color: var(--main-text-color, #222);
  cursor: pointer;
  font-size: 0.85rem;
}

.btn-outline-danger {
  padding: 8px 14px;
  border-radius: 8px;
  border: 1px solid rgba(224, 85, 85, 0.4);
  background: transparent;
  color: #e05555;
  cursor: pointer;
  font-size: 0.85rem;
  align-self: flex-start;
}

.btn-link {
  border: none;
  background: none;
  color: var(--primary-color, #4f6ef7);
  cursor: pointer;
  text-decoration: underline;
  font-size: inherit;
  padding: 0;
}

.error-text {
  color: #e05555;
  font-size: 0.8rem;
  margin: 6px 0 0;
}

.success-text {
  color: #2e7d32;
  font-size: 0.8rem;
  margin: 6px 0 0;
}

.alert-box {
  background: rgba(240, 160, 40, 0.12);
  border: 1px solid rgba(240, 160, 40, 0.35);
  color: var(--main-text-color, #222);
  border-radius: 8px;
  padding: 10px 12px;
  font-size: 0.82rem;
}

.code-box {
  font-family: monospace;
  font-size: 0.85rem;
  word-break: break-all;
  background: var(--background-color, #f5f6f8);
  border: 1px solid var(--border-color, #ddd);
  border-radius: 8px;
  padding: 12px;
  margin-top: 8px;
  user-select: all;
}

.new-code-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-top: 10px;
}

.rotate-block {
  border-top: 1px solid var(--border-color, #eee);
  padding-top: 12px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
</style>
