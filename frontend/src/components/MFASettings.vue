<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="mfa-settings">
    <!-- MFA Status Header -->
    <div v-if="settingsLoaded" class="mfa-status-header" :class="{ active: isMFAVerified }">
      <div class="status-indicator" :class="{ active: isMFAVerified }">
        <svg v-if="isMFAVerified" viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2">
          <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
          <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
        </svg>
        <svg v-else viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M12 2L2 7v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V7l-10-5z"/>
        </svg>
      </div>
      <div class="status-text">
        <h4>Authentification à deux facteurs (MFA)</h4>
        <p class="status-desc">
          {{ isMFAVerified
            ? 'Votre compte est protégé par MFA'
            : 'Ajoutez une couche de sécurité supplémentaire'
          }}
        </p>
      </div>
      <button
        v-if="!isMFAEnabled"
        class="btn-enable-mfa"
        @click="startEnrollment"
        :disabled="enrolling"
      >
        {{ enrolling ? 'Configuration...' : 'Activer le MFA' }}
      </button>
    </div>

    <!-- Enrollment Flow Modal -->
    <div v-if="showEnrollmentFlow" class="modal-overlay" @click="handleOverlayClick">
      <div class="enrollment-modal" @click.stop">
      <!-- Step 1: QR Code Display -->
      <div v-if="!enrollmentVerified" class="enrollment-step">
        <h5>Étape 1 : Scannez le QR Code</h5>
        <p class="step-desc">
          Utilisez une application d'authentification (Google Authenticator, Authy, etc.)
          pour scanner ce QR code.
        </p>

        <!-- QR Code Container -->
        <div class="qr-code-container">
          <canvas ref="qrCanvas" class="qr-canvas"></canvas>
        </div>

        <!-- Manual Secret (fallback) -->
        <div class="secret-fallback">
          <p class="secret-label">Impossible de scanner ? Entrez ce code manuellement :</p>
          <div class="secret-code">
            <code>{{ secret }}</code>
            <button class="btn-copy" @click="copySecret" title="Copier">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
                <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
              </svg>
            </button>
          </div>
        </div>

        <!-- Step 2: Verify Code -->
        <div class="verification-input">
          <h5>Étape 2 : Vérifiez avec un code</h5>
          <p class="step-desc">Entrez le code à 6 chiffres généré par votre application.</p>
          <div class="code-input-group">
            <input
              type="text"
              v-model="verificationCode"
              placeholder="000000"
              maxlength="6"
              pattern="[0-9]*"
              inputmode="numeric"
              class="code-input"
              @input="filterNumericInput"
            />
            <button
              class="btn-verify"
              @click="verifyEnrollment"
              :disabled="verifying || verificationCode.length !== 6"
            >
              {{ verifying ? 'Vérification...' : 'Vérifier' }}
            </button>
          </div>
          <p v-if="error" class="error-message">{{ error }}</p>
        </div>

        <button class="btn-cancel" @click="cancelEnrollment">Annuler</button>
      </div>

      <!-- Success State -->
      <div v-else class="enrollment-success">
        <div class="success-icon">
          <svg viewBox="0 0 24 24" width="48" height="48" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="20 6 9 17 4 12"/>
          </svg>
        </div>
        <h5>MFA activé avec succès !</h5>
        <p>Votre compte est maintenant protégé par l'authentification à deux facteurs.</p>
        <button class="btn-primary" @click="finishEnrollment">Continuer</button>
      </div>
      </div>
    </div>

    <!-- MFA Restrictions (only shown when MFA is verified) -->
    <details v-if="isMFAVerified && !showEnrollmentFlow && settingsLoaded" class="mfa-restrictions" open>
      <summary class="restrictions-summary">
        <h5>Quand exiger le code MFA ?</h5>
        <svg class="chevron-icon" viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="6 9 12 15 18 9"/>
        </svg>
      </summary>
      <div class="restrictions-content">
        <p class="restrictions-desc">
          Choisissez quand vous souhaitez être invité à entrer votre code MFA.
        </p>

      <div class="restriction-list">
        <!-- Login Restriction -->
        <div class="restriction-item">
          <div class="restriction-info">
            <div class="restriction-icon">
              <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
                <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
              </svg>
            </div>
            <div class="restriction-text">
              <span class="restriction-title">Connexion au compte</span>
              <span class="restriction-desc">Exige le MFA à chaque connexion pour accéder au tableau de bord</span>
            </div>
          </div>
          <label class="toggle-switch">
            <input
              type="checkbox"
              v-model="localSettings.require_mfa_on_login"
              @change="saveRestriction('require_mfa_on_login')"
            >
            <span class="slider"></span>
          </label>
        </div>

        <!-- Destructive Actions Restriction -->
        <div class="restriction-item">
          <div class="restriction-info">
            <div class="restriction-icon">
              <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="3 6 5 6 21 6"></polyline>
                <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
              </svg>
            </div>
            <div class="restriction-text">
              <span class="restriction-title">Actions destructives</span>
              <span class="restriction-desc">Exige le MFA pour les suppressions et modifications sensibles</span>
            </div>
          </div>
          <label class="toggle-switch">
            <input
              type="checkbox"
              v-model="localSettings.require_mfa_on_destructive_actions"
              @change="saveRestriction('require_mfa_on_destructive_actions')"
            >
            <span class="slider"></span>
          </label>
        </div>

        <!-- Download Restriction -->
        <div class="restriction-item">
          <div class="restriction-info">
            <div class="restriction-icon">
              <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                <polyline points="7 10 12 15 17 10"></polyline>
                <line x1="12" y1="15" x2="12" y2="3"></line>
              </svg>
            </div>
            <div class="restriction-text">
              <span class="restriction-title">Téléchargements</span>
              <span class="restriction-desc">Exige le MFA avant de télécharger des fichiers sensibles</span>
            </div>
          </div>
          <label class="toggle-switch">
            <input
              type="checkbox"
              v-model="localSettings.require_mfa_on_downloads"
              @change="saveRestriction('require_mfa_on_downloads')"
            >
            <span class="slider"></span>
          </label>
        </div>
      </div>

      <!-- Disable MFA -->
      <div class="mfa-danger-zone">
        <button class="btn-disable-mfa" @click="showDisableConfirm = true">
          Désactiver le MFA
        </button>
      </div>
      </div>
    </details>

    <!-- Disable MFA Confirmation Modal -->
    <div v-if="showDisableConfirm" class="modal-overlay" @click="showDisableConfirm = false">
      <div class="modal-content" @click.stop>
        <h4>Désactiver le MFA</h4>
        <p>Entrez un code MFA pour confirmer la désactivation :</p>
        <input
          type="text"
          v-model="disableCode"
          placeholder="000000"
          maxlength="6"
          pattern="[0-9]*"
          inputmode="numeric"
          class="code-input"
        />
        <div class="modal-actions">
          <button class="btn-secondary" @click="showDisableConfirm = false">Annuler</button>
          <button
            class="btn-danger"
            @click="disableMFA"
            :disabled="unenrolling || disableCode.length !== 6"
          >
            {{ unenrolling ? 'Désactivation...' : 'Désactiver' }}
          </button>
        </div>
        <p v-if="error" class="error-message">{{ error }}</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { useMFA } from '../utils/useMFA'
import QRCode from 'qrcode'

const {
  enrolling,
  verifying,
  unenrolling,
  settingsLoaded,
  qrCode,
  secret,
  securitySettings,
  error,
  isMFAEnabled,
  isMFAVerified,
  enrollMFA,
  verifyAndEnableMFA,
  unenrollMFA,
  fetchSecuritySettings,
  updateSecuritySettings
} = useMFA()

// Local state
const showEnrollmentFlow = ref(false)
const enrollmentVerified = ref(false)
const verificationCode = ref('')
const disableCode = ref('')
const showDisableConfirm = ref(false)
const qrCanvas = ref(null)

// Local copy of settings for immediate UI updates
const localSettings = ref({
  require_mfa_on_login: false,
  require_mfa_on_destructive_actions: false,
  require_mfa_on_downloads: false
})

// Sync local settings with fetched settings
watch(securitySettings, (newSettings) => {
  localSettings.value = {
    require_mfa_on_login: newSettings.require_mfa_on_login,
    require_mfa_on_destructive_actions: newSettings.require_mfa_on_destructive_actions,
    require_mfa_on_downloads: newSettings.require_mfa_on_downloads
  }
}, { deep: true, immediate: true })

onMounted(async () => {
  await fetchSecuritySettings()
})

// Function to render QR code to canvas
async function renderQRCode() {
  //console.log('[MFASettings] renderQRCode called')
  //console.log('[MFASettings] qrCode.value:', qrCode.value)
  //console.log('[MFASettings] qrCanvas.value:', qrCanvas.value)

  if (!qrCode.value) {
    console.error('[MFASettings] Cannot render: qrCode.value is null/undefined')
    return
  }

  if (!qrCanvas.value) {
    console.error('[MFASettings] Cannot render: qrCanvas.value is null/undefined')
    return
  }

  await nextTick()

  try {
    //console.log('[MFASettings] Calling QRCode.toCanvas...')
    await QRCode.toCanvas(qrCanvas.value, qrCode.value, {
      width: 200,
      margin: 1,
      color: {
        dark: '#2D1B22',
        light: '#FFFFFF'
      }
    })
    //console.log('[MFASettings] ✓ QR Code rendered successfully!')
  } catch (err) {
    console.error('[MFASettings] ✗ Failed to generate QR code:', err)
  }
}

// Watch for QR code and try to render it
watch(qrCode, async (newQrCode, oldQrCode) => {
  //console.log('[MFASettings] qrCode watcher triggered:', { old: oldQrCode, new: newQrCode })
  if (newQrCode) {
    await renderQRCode()
  }
})

// Watch for canvas becoming available and render if we have a QR code
watch(qrCanvas, async (newCanvas, oldCanvas) => {
  //console.log('[MFASettings] qrCanvas watcher triggered:', { old: oldCanvas, new: newCanvas })
  if (newCanvas && qrCode.value) {
    //console.log('[MFASettings] Both canvas and qrCode available, rendering...')
    await renderQRCode()
  }
})

async function startEnrollment() {
  //console.log('[MFASettings] startEnrollment called')
  try {
    //console.log('[MFASettings] Calling enrollMFA()...')
    const enrollResult = await enrollMFA()
    //console.log('[MFASettings] enrollMFA() returned:', enrollResult)

    showEnrollmentFlow.value = true
    enrollmentVerified.value = false

    //console.log('[MFASettings] Modal opened, waiting for DOM...')
    // Wait for modal and canvas to be mounted
    await nextTick()
    await nextTick()

    //console.log('[MFASettings] DOM ready, checking values...')
    //console.log('[MFASettings] qrCode.value:', qrCode.value)
    //console.log('[MFASettings] qrCanvas.value:', qrCanvas.value)
    //console.log('[MFASettings] secret.value:', secret.value)

    // Force render after DOM is ready
    await renderQRCode()
  } catch (err) {
    console.error('[MFASettings] Enrollment failed:', err)
    error.value = err.message || 'Erreur lors de l\'activation du MFA'
  }
}

async function verifyEnrollment() {
  try {
    await verifyAndEnableMFA(verificationCode.value)
    enrollmentVerified.value = true
    verificationCode.value = ''
  } catch (err) {
    console.error('Verification failed:', err)
  }
}

function finishEnrollment() {
  showEnrollmentFlow.value = false
  enrollmentVerified.value = false
}

function cancelEnrollment() {
  showEnrollmentFlow.value = false
  enrollmentVerified.value = false
  verificationCode.value = ''
}

function handleOverlayClick() {
  // Only allow closing if enrollment is complete (success state)
  // This prevents accidental closure during QR scan/verification
  if (enrollmentVerified.value) {
    finishEnrollment()
  }
  // Otherwise, user must click "Annuler" button
}

async function saveRestriction(key) {
  try {
    // Always preserve MFA status when updating restrictions
    await updateSecuritySettings({
      mfa_enabled: securitySettings.value.mfa_enabled,
      mfa_verified: securitySettings.value.mfa_verified,
      [key]: localSettings.value[key]
    })
    // Re-fetch to ensure everything stays synced
    await fetchSecuritySettings()
  } catch (err) {
    console.error('Failed to save restriction:', err)
    // Revert on error
    localSettings.value[key] = !localSettings.value[key]
  }
}

async function disableMFA() {
  try {
    await unenrollMFA(disableCode.value)
    showDisableConfirm.value = false
    disableCode.value = ''
  } catch (err) {
    console.error('Failed to disable MFA:', err)
  }
}

function copySecret() {
  if (secret.value) {
    navigator.clipboard.writeText(secret.value)
  }
}

function filterNumericInput(event) {
  event.target.value = event.target.value.replace(/\D/g, '')
  verificationCode.value = event.target.value
}
</script>

<style scoped>
.mfa-settings {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.mfa-status-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1.5rem;
  background: var(--hover-background-color);
  border-radius: 12px;
  border: 1px solid var(--border-color);
  transition: all 0.3s ease;
}

.mfa-status-header.active {
  background: rgba(34, 197, 94, 0.08);
  border-color: rgba(34, 197, 94, 0.3);
}

.mfa-status-header.active .status-indicator {
  background: rgba(34, 197, 94, 0.15);
  color: rgb(22, 163, 74);
}

.status-indicator {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--border-color);
  color: var(--secondary-text-color);
  flex-shrink: 0;
}

.status-indicator.active {
  background: var(--success-color);
  color: white;
}

.status-text {
  flex: 1;
}

.status-text h4 {
  margin: 0 0 0.25rem 0;
  font-size: 1.1rem;
  color: var(--main-text-color);
}

.status-desc {
  margin: 0;
  font-size: 0.9rem;
  color: var(--secondary-text-color);
}

.btn-enable-mfa {
  padding: 0.75rem 1.5rem;
  background: var(--primary-color);
  color: white;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-enable-mfa:hover:not(:disabled) {
  background: var(--accent-color);
}

.btn-enable-mfa:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.enrollment-modal {
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 16px;
  padding: 1.5rem;
  max-width: 550px;
  width: 90%;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  animation: slideUp 0.3s ease-out;
  display: flex;
  flex-direction: column;
}

.enrollment-modal::-webkit-scrollbar {
  width: 8px;
}

.enrollment-modal::-webkit-scrollbar-track {
  background: var(--hover-background-color);
  border-radius: 4px;
}

.enrollment-modal::-webkit-scrollbar-thumb {
  background: var(--primary-color);
  border-radius: 4px;
}

.enrollment-modal::-webkit-scrollbar-thumb:hover {
  background: var(--accent-color);
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

.enrollment-step h5,
.enrollment-success h5 {
  margin: 0 0 0.5rem 0;
  font-size: 1.1rem;
  color: var(--main-text-color);
}

.step-desc {
  margin: 0 0 0.75rem 0;
  color: var(--secondary-text-color);
  font-size: 0.9rem;
}

.qr-code-container {
  display: flex;
  justify-content: center;
  margin: 1rem 0;
  padding: 1rem;
  background: white;
  border-radius: 12px;
}

.qr-canvas {
  border-radius: 8px;
}

.secret-fallback {
  margin: 1rem 0;
  padding: 0.875rem;
  background: var(--hover-background-color);
  border-radius: 8px;
}

.secret-label {
  margin: 0 0 0.5rem 0;
  font-size: 0.85rem;
  color: var(--secondary-text-color);
}

.secret-code {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.secret-code code {
  flex: 1;
  padding: 0.625rem;
  background: var(--background-color);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  font-family: 'Courier New', monospace;
  font-size: 0.9rem;
  color: var(--primary-color);
  letter-spacing: 1px;
  word-break: break-all;
}

.btn-copy {
  padding: 0.625rem;
  background: var(--primary-color);
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-copy:hover {
  background: var(--accent-color);
}

.verification-input {
  margin-top: 1rem;
}

.code-input-group {
  display: flex;
  gap: 0.75rem;
  margin-top: 0.75rem;
}

.code-input {
  flex: 1;
  padding: 0.875rem;
  border: 2px solid var(--border-color);
  border-radius: 8px;
  font-size: 1.25rem;
  text-align: center;
  letter-spacing: 0.4rem;
  font-family: 'Courier New', monospace;
  background: var(--background-color);
  color: var(--main-text-color);
  transition: border-color 0.2s;
}

.code-input:focus {
  outline: none;
  border-color: var(--primary-color);
}

.btn-verify {
  padding: 0.875rem 1.5rem;
  background: var(--success-color);
  color: white;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.btn-verify:hover:not(:disabled) {
  opacity: 0.9;
  transform: translateY(-1px);
}

.btn-verify:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-cancel {
  margin-top: 0.75rem;
  padding: 0.625rem 1.25rem;
  background: transparent;
  border: 1px solid var(--border-color);
  color: var(--secondary-text-color);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-cancel:hover {
  background: var(--hover-background-color);
}

.enrollment-success {
  text-align: center;
  padding: 2rem;
}

.success-icon {
  width: 80px;
  height: 80px;
  margin: 0 auto 1rem;
  border-radius: 50%;
  background: var(--success-color);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
}

.enrollment-success p {
  color: var(--secondary-text-color);
  margin-bottom: 2rem;
}

.btn-primary {
  padding: 1rem 2rem;
  background: var(--primary-color);
  color: white;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary:hover {
  background: var(--accent-color);
}

.mfa-restrictions {
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 0;
  overflow: hidden;
}

.restrictions-summary {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1.5rem 2rem;
  cursor: pointer;
  user-select: none;
  list-style: none;
  transition: background-color 0.2s;
}

.restrictions-summary::-webkit-details-marker {
  display: none;
}

.restrictions-summary:hover {
  background: var(--hover-background-color);
}

.restrictions-summary h5 {
  margin: 0;
  font-size: 1.1rem;
  color: var(--main-text-color);
}

.chevron-icon {
  flex-shrink: 0;
  transition: transform 0.3s ease;
  color: var(--secondary-text-color);
}

.mfa-restrictions[open] .chevron-icon {
  transform: rotate(180deg);
}

.restrictions-content {
  padding: 0 2rem 2rem 2rem;
}

.mfa-restrictions h5 {
  margin: 0 0 0.5rem 0;
  font-size: 1.1rem;
  color: var(--main-text-color);
}

.restrictions-desc {
  margin: 0 0 1.5rem 0;
  color: var(--secondary-text-color);
  font-size: 0.95rem;
}

.restriction-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.restriction-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1.25rem;
  background: var(--hover-background-color);
  border-radius: 8px;
  border: 1px solid var(--border-color);
}

.restriction-info {
  display: flex;
  align-items: center;
  gap: 1rem;
  flex: 1;
}

.restriction-icon {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  background: var(--primary-color);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.restriction-text {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.restriction-title {
  font-weight: 600;
  color: var(--main-text-color);
  font-size: 0.95rem;
}

.restriction-desc {
  font-size: 0.85rem;
  color: var(--secondary-text-color);
}

.toggle-switch {
  position: relative;
  display: inline-block;
  width: 50px;
  height: 28px;
  flex-shrink: 0;
}

.toggle-switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: var(--border-color);
  transition: 0.3s;
  border-radius: 28px;
}

.slider:before {
  position: absolute;
  content: "";
  height: 20px;
  width: 20px;
  left: 4px;
  bottom: 4px;
  background-color: white;
  transition: 0.3s;
  border-radius: 50%;
}

input:checked + .slider {
  background-color: var(--primary-color);
}

input:checked + .slider:before {
  transform: translateX(22px);
}

.mfa-danger-zone {
  margin-top: 2rem;
  padding-top: 2rem;
  border-top: 1px solid var(--border-color);
}

.btn-disable-mfa {
  padding: 0.75rem 1.5rem;
  background: transparent;
  border: 1px solid var(--error-color);
  color: var(--error-color);
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-disable-mfa:hover {
  background: var(--error-color);
  color: white;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: var(--card-color);
  border-radius: 12px;
  padding: 2rem;
  max-width: 400px;
  width: 90%;
}

.modal-content h4 {
  margin: 0 0 1rem 0;
  color: var(--main-text-color);
}

.modal-content p {
  margin: 0 0 1rem 0;
  color: var(--secondary-text-color);
}

.modal-actions {
  display: flex;
  gap: 1rem;
  margin-top: 1.5rem;
}

.btn-secondary {
  flex: 1;
  padding: 0.75rem;
  background: transparent;
  border: 1px solid var(--border-color);
  color: var(--main-text-color);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-secondary:hover {
  background: var(--hover-background-color);
}

.btn-danger {
  flex: 1;
  padding: 0.75rem;
  background: var(--error-color);
  color: white;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-danger:hover:not(:disabled) {
  opacity: 0.9;
}

.btn-danger:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.error-message {
  margin-top: 1rem;
  padding: 0.75rem;
  background: rgba(231, 76, 60, 0.1);
  border: 1px solid var(--error-color);
  border-radius: 6px;
  color: var(--error-color);
  font-size: 0.9rem;
  text-align: center;
}
</style>
