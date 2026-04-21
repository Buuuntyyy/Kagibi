<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div v-if="modelValue" class="mfa-modal-overlay" @click="cancel">
    <div class="mfa-modal-content" @click.stop>
      <div class="mfa-modal-header">
        <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M12 2L2 7v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V7l-10-5z"/>
          <polyline points="9 12 11 14 15 10"/>
        </svg>
        <h3>Vérification MFA requise</h3>
        <button class="btn-close" @click="cancel">×</button>
      </div>

      <div class="mfa-modal-body">
        <p class="mfa-description">{{ contextMessage }}</p>

        <div class="code-input-wrapper">
          <input
            ref="codeInput"
            type="text"
            v-model="code"
            placeholder="000000"
            maxlength="6"
            pattern="[0-9]*"
            inputmode="numeric"
            class="mfa-code-input"
            @input="filterNumeric"
            @keydown.enter="verify"
          />
        </div>

        <p v-if="error" class="error-message">{{ error }}</p>
      </div>

      <div class="mfa-modal-footer">
        <button class="btn-secondary" @click="cancel">Annuler</button>
        <button
          class="btn-primary"
          @click="verify"
          :disabled="verifying || code.length !== 6"
        >
          {{ verifying ? 'Vérification...' : 'Vérifier' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick } from 'vue'
import { useMFA } from '../utils/useMFA'

const props = defineProps({
  modelValue: Boolean,
  context: {
    type: String,
    default: 'general',
    validator: (value) => ['login', 'destructive', 'download', 'email_change', 'general'].includes(value)
  }
})

const emit = defineEmits(['update:modelValue', 'verified', 'cancelled'])

const { createChallenge, verifyChallenge, error } = useMFA()

const code = ref('')
const verifying = ref(false)
const challengeId = ref(null)
const factorId = ref(null)
const codeInput = ref(null)

const contextMessages = {
  login: 'Entrez votre code MFA pour vous connecter.',
  destructive: 'Cette action nécessite une vérification MFA pour des raisons de sécurité.',
  download: 'L\'accès à ce fichier nécessite une vérification MFA.',
  email_change: 'La modification de l\'adresse email nécessite une vérification MFA.',
  general: 'Veuillez entrer votre code MFA pour continuer.'
}

const contextMessage = computed(() => contextMessages[props.context])

// Auto-focus input when modal opens
watch(() => props.modelValue, async (isOpen) => {
  if (isOpen) {
    code.value = ''
    error.value = null

    // Create challenge when modal opens
    try {
      const challenge = await createChallenge()
      challengeId.value = challenge.challengeId
      factorId.value = challenge.factorId
    } catch (err) {
      console.error('Failed to create MFA challenge:', err)
      error.value = 'Impossible de créer le challenge MFA'
    }

    await nextTick()
    codeInput.value?.focus()
  }
})

function filterNumeric(event) {
  event.target.value = event.target.value.replace(/\D/g, '')
  code.value = event.target.value
}

async function verify() {
  if (code.value.length !== 6) return

  verifying.value = true

  try {
    await verifyChallenge(challengeId.value, factorId.value, code.value)
    emit('verified')
    emit('update:modelValue', false)
  } catch (err) {
    console.error('MFA verification failed:', err)
    // error.value is already set by useMFA
  } finally {
    verifying.value = false
  }
}

function cancel() {
  emit('cancelled')
  emit('update:modelValue', false)
}
</script>

<style scoped>
.mfa-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
  backdrop-filter: blur(4px);
}

.mfa-modal-content {
  background: var(--card-color);
  border-radius: 16px;
  max-width: 440px;
  width: 90%;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  animation: slideUp 0.3s ease-out;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.mfa-modal-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1.5rem;
  border-bottom: 1px solid var(--border-color);
}

.mfa-modal-header svg {
  color: var(--primary-color);
  flex-shrink: 0;
}

.mfa-modal-header h3 {
  flex: 1;
  margin: 0;
  font-size: 1.25rem;
  color: var(--main-text-color);
}

.btn-close {
  background: none;
  border: none;
  font-size: 1.75rem;
  color: var(--secondary-text-color);
  cursor: pointer;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  transition: all 0.2s;
}

.btn-close:hover {
  background: var(--hover-background-color);
  color: var(--main-text-color);
}

.mfa-modal-body {
  padding: 2rem 1.5rem;
}

.mfa-description {
  margin: 0 0 1.5rem 0;
  color: var(--secondary-text-color);
  text-align: center;
  line-height: 1.5;
}

.code-input-wrapper {
  display: flex;
  justify-content: center;
}

.mfa-code-input {
  width: 100%;
  max-width: 240px;
  padding: 1.25rem;
  border: 2px solid var(--border-color);
  border-radius: 12px;
  font-size: 2rem;
  text-align: center;
  letter-spacing: 0.75rem;
  font-family: 'Courier New', monospace;
  background: var(--background-color);
  color: var(--main-text-color);
  transition: all 0.2s;
}

.mfa-code-input:focus {
  outline: none;
  border-color: var(--primary-color);
  box-shadow: 0 0 0 4px rgba(250, 114, 104, 0.1);
}

.error-message {
  margin-top: 1rem;
  padding: 0.75rem;
  background: rgba(231, 76, 60, 0.1);
  border: 1px solid var(--error-color);
  border-radius: 8px;
  color: var(--error-color);
  font-size: 0.9rem;
  text-align: center;
}

.mfa-modal-footer {
  display: flex;
  gap: 1rem;
  padding: 1.5rem;
  border-top: 1px solid var(--border-color);
}

.btn-secondary,
.btn-primary {
  flex: 1;
  padding: 0.875rem;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  font-size: 1rem;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-secondary {
  background: transparent;
  border: 1px solid var(--border-color);
  color: var(--main-text-color);
}

.btn-secondary:hover {
  background: var(--hover-background-color);
}

.btn-primary {
  background: var(--primary-color);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: var(--accent-color);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(250, 114, 104, 0.3);
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
