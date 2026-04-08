<template>
  <div v-if="!recoveryCode">
    <form @submit.prevent="submit" class="auth-form">
      <div class="form-row-top">
        <div class="avatar-sidebar">
          <label class="field-label">Avatar</label>
          <AvatarSelector v-model="selectedAvatar" />
        </div>

        <div class="identity-fields">
          <div class="form-group">
            <label>Nom d'utilisateur</label>
            <input v-model="username" type="text" required class="form-control" placeholder="Votre nom" />
          </div>
          <div class="form-group">
            <label>Email</label>
            <input v-model="email" type="email" required class="form-control" placeholder="votre@email.com" />
          </div>
        </div>
      </div>

      <div class="form-group">
        <label>Mot de passe</label>
        <div class="password-input-wrapper">
          <input
            v-model="password"
            :type="showPassword ? 'text' : 'password'"
            required
            class="form-control"
            placeholder="••••••••"
            @focus="passwordFocused = true"
            @blur="handlePasswordBlur"
          />
          <button type="button" class="toggle-password-btn" @click="showPassword = !showPassword" tabindex="-1">
            <svg v-if="!showPassword" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
              <circle cx="12" cy="12" r="3"></circle>
            </svg>
            <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"></path>
              <line x1="1" y1="1" x2="23" y2="23"></line>
            </svg>
          </button>
        </div>
        <PasswordCriteria :password="password" :show="passwordFocused" />
      </div>

      <!-- Filename encryption opt-in -->
      <div class="option-group">
        <label class="option-toggle" :class="{ active: encryptFilenames }">
          <input type="checkbox" v-model="encryptFilenames" />
          <span class="toggle-track"><span class="toggle-thumb"></span></span>
          <span class="option-label">
            Chiffrer les noms de fichiers et dossiers
            <span class="option-hint">
              Les noms sont stockés chiffrés sur le serveur — la barre de recherche sera désactivée.
            </span>
          </span>
        </label>
      </div>

      <!-- Trust copy — zero-knowledge guarantee (UX-DR12) -->
      <div class="zk-trust-note">
        <svg class="zk-shield-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
        </svg>
        <p>Vos fichiers sont chiffrés avant de quitter votre appareil. Nous ne pouvons pas les lire.</p>
      </div>

      <button type="submit" class="btn-submit" :disabled="loading">
        <span v-if="loading" class="spinner"></span>
        <span v-else>S'inscrire</span>
      </button>
      <p v-if="error" class="error-message">{{ error }}</p>
    </form>
  </div>

  <!-- Recovery code display — UX-DR13: alertdialog with focus lock -->
  <div
    v-else
    class="recovery-display"
    role="alertdialog"
    aria-modal="true"
    aria-labelledby="recovery-title"
    @keydown="trapFocus"
    ref="recoveryDisplayRef"
  >
    <h3 id="recovery-title">Compte créé avec succès !</h3>
    <div class="alert-box">
      <strong>IMPORTANT :</strong> Voici votre code de récupération. Conservez-le en lieu sûr. C'est le SEUL moyen de récupérer votre compte si vous perdez votre mot de passe.
    </div>
    <div class="code-box">
      {{ recoveryCode }}
    </div>
    <div class="actions">
      <button @click="copyCode" class="btn-secondary">Copier le code</button>

      <div class="copy-confirm">
        <input
          type="checkbox"
          id="copied-checkbox"
          v-model="codeCheckboxChecked"
          autofocus
        />
        <label for="copied-checkbox">J'ai copié ce code en lieu sûr</label>
      </div>

      <button class="btn-submit" @click="showConfirmModal = true" :disabled="!codeCheckboxChecked">
        Continuer vers le Dashboard
      </button>
    </div>

    <!-- Toast Notification -->
    <div v-if="showNotification" class="toast-notification">
      <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="toast-icon">
        <polyline points="20 6 9 17 4 12"></polyline>
      </svg>
      <span>Code copié avec succès !</span>
    </div>

    <!-- Confirmation Modal -->
    <div v-if="showConfirmModal" class="modal-overlay" @click="showConfirmModal = false">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="modal-icon">
            <circle cx="12" cy="12" r="10"></circle>
            <line x1="12" y1="16" x2="12" y2="12"></line>
            <line x1="12" y1="8" x2="12.01" y2="8"></line>
          </svg>
          <h3>Conseil de sécurité</h3>
        </div>
        <div class="modal-body">
          <p><strong>Avez-vous sauvegardé votre code de récupération ?</strong></p>
          <p>Nous vous recommandons d'utiliser un gestionnaire de mots de passe (comme <strong>Bitwarden</strong>, <strong>1Password</strong>, ou <strong>KeePass</strong>) pour sauvegarder ce code de récupération ainsi que vos identifiants de connexion.</p>
          <p class="warning-text">
            <svg viewBox="0 0 24 24" fill="none" class="inline-alert-icon" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/>
              <line x1="12" y1="9" x2="12" y2="13"/>
              <line x1="12" y1="17" x2="12.01" y2="17"/>
            </svg>
            <b>Sans ce code, vous ne pourrez pas récupérer votre compte si vous oubliez votre mot de passe.</b>
          </p>
        </div>
        <div class="modal-footer">
          <button class="btn-cancel" @click="showConfirmModal = false">Annuler</button>
          <button class="btn-confirm" @click="finishRegistration">Continuer</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useAuthStore } from '../../stores/auth'
import { useRouter } from 'vue-router'
import AvatarSelector from '../AvatarSelector.vue'
import PasswordCriteria from './PasswordCriteria.vue'
import { checkPasswordCriteria, getPasswordErrors } from '../../utils/passwordStrength'

const username = ref('')
const email = ref('')
const password = ref('')
const selectedAvatar = ref('/avatars/default.png')
const encryptFilenames = ref(false)
const error = ref('')
const recoveryCode = ref('')
const loading = ref(false)
const showPassword = ref(false)
const showConfirmModal = ref(false)
const codeCheckboxChecked = ref(false)
const showNotification = ref(false)
const passwordFocused = ref(false)
const recoveryDisplayRef = ref(null)
const authStore = useAuthStore()
const router = useRouter()

const handlePasswordBlur = () => {
  const { valid } = checkPasswordCriteria(password.value)
  if (valid) passwordFocused.value = false
  // Keep showing criteria if password is still invalid so user sees what's missing
}

const submit = async () => {
  error.value = ''

  // Validate password strength before sending to backend
  const { valid } = checkPasswordCriteria(password.value)
  if (!valid) {
    const errors = getPasswordErrors(password.value)
    error.value = errors[0]
    return
  }

  loading.value = true
  try {
    const code = await authStore.register(username.value, email.value, password.value, selectedAvatar.value, encryptFilenames.value)
    recoveryCode.value = code
  } catch (err) {
    console.error(err)
    // Show backend error message if available
    error.value = err.message || "Erreur lors de l'inscription"
  } finally {
    loading.value = false
  }
}

const copyCode = async () => {
  try {
    await navigator.clipboard.writeText(recoveryCode.value)
    showNotification.value = true
    setTimeout(() => { showNotification.value = false }, 3000)
  } catch {
    error.value = 'Impossible de copier le code. Copiez-le manuellement.'
  }
}

// Focus trap for recovery display section (UX-DR13)
const trapFocus = (event) => {
  if (!recoveryDisplayRef.value) return

  // Escape: keep user in dialog (alertdialog cannot be dismissed until checkbox is checked)
  if (event.key === 'Escape') {
    event.preventDefault()
    return
  }

  if (event.key !== 'Tab') return

  const focusableSelectors = 'button, input, [tabindex]:not([tabindex="-1"])'
  const focusable = Array.from(recoveryDisplayRef.value.querySelectorAll(focusableSelectors))
  if (focusable.length === 0) return

  const first = focusable[0]
  const last = focusable[focusable.length - 1]

  if (event.shiftKey) {
    if (document.activeElement === first) {
      event.preventDefault()
      last.focus()
    }
  } else {
    if (document.activeElement === last) {
      event.preventDefault()
      first.focus()
    }
  }
}

// When recovery display becomes visible, programmatically focus the checkbox (P-10)
watch(recoveryCode, (newCode) => {
  if (newCode) {
    // Use nextTick equivalent — wait one frame for v-else to render
    setTimeout(() => {
      const checkbox = recoveryDisplayRef.value?.querySelector('#copied-checkbox')
      checkbox?.focus()
    }, 50)
  }
})

const finishRegistration = async () => {
    showConfirmModal.value = false
    // After successful registration, user is already authenticated
    // The register() function already set isAuthenticated = true and loaded the MasterKey
    // Just ensure RSA keys are loaded
    try {
      await authStore.ensureRSAKeys(authStore.masterKey);
      router.push({ name: 'MyFiles' })
    } catch (err) {
      console.error('Error ensuring RSA keys after registration:', err);
      error.value = 'Erreur lors de la finalisation de l\'inscription: ' + err.message;
    }
}
</script>

<style scoped>
.auth-form {
  display: flex;
  flex-direction: column;
  gap: 1.2rem;
  text-align: left;
}

.form-row-top {
  display: flex;
  gap: 1.25rem;
  align-items: flex-start;
}

.avatar-sidebar {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.identity-fields {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

/* Ensure consistent label styling */
label,
.field-label {
  font-size: 0.9rem;
  font-weight: 500;
  color: var(--secondary-text-color);
}

.form-control {
  padding: 14px 16px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background-color: var(--background-color);
  color: var(--main-text-color);
  font-size: 1.05rem;
  transition: all 0.2s ease;
}

.form-control:focus {
  outline: none;
  border-color: var(--primary-color);
  box-shadow: 0 0 0 4px rgba(52, 152, 219, 0.1);
  background-color: var(--card-color);
}

/* Filename encryption toggle */
.option-group {
  margin-bottom: 0.5rem;
}

.option-toggle {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  cursor: pointer;
  user-select: none;
}

.option-toggle input[type="checkbox"] {
  position: absolute;
  opacity: 0;
  width: 0;
  height: 0;
}

.toggle-track {
  flex-shrink: 0;
  position: relative;
  width: 36px;
  height: 20px;
  background-color: var(--border-color, #ccc);
  border-radius: 10px;
  transition: background-color 0.2s;
  margin-top: 2px;
}

.option-toggle.active .toggle-track {
  background-color: #1a73e8;
}

.toggle-thumb {
  position: absolute;
  top: 2px;
  left: 2px;
  width: 16px;
  height: 16px;
  background: #fff;
  border-radius: 50%;
  transition: transform 0.2s;
  box-shadow: 0 1px 3px rgba(0,0,0,0.2);
}

.option-toggle.active .toggle-thumb {
  transform: translateX(16px);
}

.option-label {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  font-size: 0.9rem;
  color: var(--main-text-color);
}

.option-hint {
  font-size: 0.75rem;
  color: var(--secondary-text-color);
  line-height: 1.4;
}

/* Zero-knowledge trust note (UX-DR12) */
.zk-trust-note {
  display: flex;
  align-items: flex-start;
  gap: 0.6rem;
  background: rgba(52, 152, 219, 0.06);
  border: 1px solid rgba(52, 152, 219, 0.25);
  border-radius: 8px;
  padding: 0.75rem 1rem;
  color: var(--main-text-color);
}

.zk-shield-icon {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
  margin-top: 1px;
  color: var(--primary-color, #3498db);
}

.zk-trust-note p {
  margin: 0;
  font-size: 0.88rem;
  line-height: 1.45;
  color: var(--secondary-text-color, #64748b);
}

.btn-submit {
  margin-top: 0.5rem;
  background-color: var(--primary-color);
  color: white;
  border: none;
  padding: 12px;
  font-size: 1rem;
  font-weight: 600;
  border-radius: 8px;
  cursor: pointer;
  transition: background-color 0.2s;
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100%;
}

.btn-submit:hover:not(:disabled) {
  background-color: var(--accent-color);
}

.btn-submit:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.btn-secondary {
  background-color: transparent;
  border: 1px solid var(--border-color);
  color: var(--main-text-color);
  padding: 10px;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 500;
  transition: background-color 0.2s;
  width: 100%;
}

.btn-secondary:hover {
  background-color: var(--background-color);
}

.error-message {
  color: var(--error-color);
  font-size: 0.9rem;
  text-align: center;
  margin: 0;
  padding: 8px;
  background-color: rgba(231, 76, 60, 0.1);
  border-radius: 4px;
}

.spinner {
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-radius: 50%;
  border-top: 2px solid white;
  width: 16px;
  height: 16px;
  animation: spin 1s linear infinite;
}

/* Recovery Display Styles */
.recovery-display {
  text-align: center;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  width: 80%;
  max-width: 800px;
  margin: 0 auto;
}

.recovery-display h3 {
  color: var(--success-color, #2ecc71);
  margin: 0;
}

.alert-box {
  background-color: rgba(241, 196, 15, 0.1);
  border: 1px solid rgba(241, 196, 15, 0.3);
  color: var(--main-text-color);
  padding: 1rem;
  border-radius: 8px;
  font-size: 0.9rem;
  line-height: 1.5;
  text-align: left;
}

.code-box {
  background-color: var(--background-color);
  border: 2px dashed var(--border-color);
  padding: 1rem;
  font-family: 'JetBrains Mono', monospace;
  font-size: 1.2rem;
  font-weight: bold;
  letter-spacing: 1px;
  border-radius: 8px;
  word-break: break-all;
  color: var(--primary-color);
  user-select: all;
}

.actions {
  display: flex;
  flex-direction: column;
  gap: 0.8rem;
}

/* Checkbox gate for recovery code acknowledgment */
.copy-confirm {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  text-align: left;
  font-size: 0.92rem;
  color: var(--main-text-color);
}

.copy-confirm input[type="checkbox"] {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
  cursor: pointer;
  accent-color: var(--primary-color, #3498db);
}

.copy-confirm label {
  cursor: pointer;
  font-size: 0.92rem;
  color: var(--main-text-color);
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.password-input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.password-input-wrapper input {
  flex: 1;
  padding-right: 40px;
}

.toggle-password-btn {
  position: absolute;
  right: 8px;
  background: none;
  border: none;
  cursor: pointer;
  color: var(--secondary-text-color);
  padding: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  transition: all 0.2s;
}

.toggle-password-btn:hover {
  color: var(--primary-color);
  background: rgba(52, 152, 219, 0.1);
}

.toggle-password-btn svg {
  width: 18px;
  height: 18px;
}

/* Modal Styles */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  animation: fadeIn 0.2s ease;
}

.modal-content {
  background: var(--card-color);
  border-radius: 12px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.2);
  max-width: 500px;
  width: 90%;
  animation: slideUp 0.3s ease;
  border: 1px solid var(--border-color);
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

.modal-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 1.5rem;
  border-bottom: 1px solid var(--border-color);
}

.modal-icon {
  color: var(--primary-color);
  flex-shrink: 0;
}

.modal-header h3 {
  margin: 0;
  font-size: 1.2rem;
  color: var(--main-text-color);
}

.modal-body {
  padding: 1.5rem;
  color: var(--main-text-color);
  line-height: 1.6;
}

.modal-body p {
  margin: 0 0 1rem 0;
}

.modal-body p:last-child {
  margin-bottom: 0;
}

.warning-text {
  color: var(--error-color, #e74c3c);
  font-weight: 500;
  background-color: rgba(231, 76, 60, 0.1);
  padding: 0.75rem;
  border-radius: 6px;
  margin-top: 1rem;
  display: flex;
  gap: 0.75rem;
  align-items: flex-start;
}

.inline-alert-icon {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
  margin-top: 2px;
}

.modal-footer {
  padding: 1rem 1.5rem;
  border-top: 1px solid var(--border-color);
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
}

.btn-cancel {
  background-color: transparent;
  border: 1px solid var(--border-color);
  color: var(--main-text-color);
  padding: 0.7rem 1.5rem;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 500;
  transition: all 0.2s;
}

.btn-cancel:hover {
  background-color: var(--background-color);
  border-color: var(--secondary-text-color);
}

.btn-confirm {
  background-color: var(--primary-color);
  color: white;
  border: none;
  padding: 0.7rem 1.5rem;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 600;
  transition: background-color 0.2s;
}

.btn-confirm:hover {
  background-color: var(--accent-color);
}

/* Toast Notification */
.toast-notification {
  position: fixed;
  bottom: 30px;
  left: 50%;
  transform: translateX(-50%);
  background: linear-gradient(135deg, #2ecc71, #27ae60);
  color: white;
  padding: 1rem 1.5rem;
  border-radius: 12px;
  box-shadow: 0 10px 30px rgba(46, 204, 113, 0.4);
  display: flex;
  align-items: center;
  gap: 12px;
  font-weight: 600;
  font-size: 1rem;
  z-index: 2000;
  animation: slideUpToast 0.4s cubic-bezier(0.68, -0.55, 0.265, 1.55);
}

.toast-icon {
  flex-shrink: 0;
  stroke-width: 3;
}

@keyframes slideUpToast {
  from {
    opacity: 0;
    transform: translate(-50%, 100px);
  }
  to {
    opacity: 1;
    transform: translate(-50%, 0);
  }
}

@media (max-width: 768px) {
  .recovery-display {
    width: 95%;
  }

  .toast-notification {
    width: 90%;
    left: 5%;
    transform: none;
  }

  @keyframes slideUpToast {
    from {
      opacity: 0;
      transform: translateY(100px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
}
</style>
