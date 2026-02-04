<template>
  <div v-if="!recoveryCode">
    <form @submit.prevent="submit" class="auth-form">
      <div class="form-group">
        <label>Nom d'utilisateur</label>
        <input v-model="username" type="text" required class="form-control" placeholder="Votre nom" />
      </div>
      <div class="form-group">
        <label>Email</label>
        <input v-model="email" type="email" required class="form-control" placeholder="votre@email.com" />
      </div>
      <div class="form-group">
        <label>Mot de passe</label>
        <div class="password-input-wrapper">
          <input v-model="password" :type="showPassword ? 'text' : 'password'" required class="form-control" placeholder="••••••••" />
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
      </div>
      <button type="submit" class="btn-submit" :disabled="loading">
        <span v-if="loading" class="spinner"></span>
        <span v-else>S'inscrire</span>
      </button>
      <p v-if="error" class="error-message">{{ error }}</p>
    </form>
  </div>
  <div v-else class="recovery-display">
    <h3>Compte créé avec succès !</h3>
    <div class="alert-box">
      <strong>IMPORTANT :</strong> Voici votre code de récupération. Conservez-le en lieu sûr. C'est le SEUL moyen de récupérer votre compte si vous perdez votre mot de passe.
    </div>
    <div class="code-box">
      {{ recoveryCode }}
    </div>
    <div class="actions">
      <button @click="copyCode" class="btn-secondary">Copier le code</button>
      <button class="btn-submit" @click="showConfirmModal = true" :disabled="!codeCopied">Continuer vers le Dashboard</button>
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
          <p class="warning-text"><b>⚠️ Sans ce code, vous ne pourrez pas récupérer votre compte si vous oubliez votre mot de passe.</b></p>
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
import { ref } from 'vue'
import { useAuthStore } from '../../stores/auth'
import { useRouter } from 'vue-router'

const username = ref('')
const email = ref('')
const password = ref('')
const error = ref('')
const recoveryCode = ref('')
const loading = ref(false)
const showPassword = ref(false)
const showConfirmModal = ref(false)
const codeCopied = ref(false)
const showNotification = ref(false)
const authStore = useAuthStore()
const router = useRouter()

const submit = async () => {
  error.value = ''
  loading.value = true
  try {
    const code = await authStore.register(username.value, email.value, password.value)
    recoveryCode.value = code
  } catch (err) {
    console.error(err)
    // Show backend error message if available
    error.value = err.message || "Erreur lors de l'inscription"
  } finally {
    loading.value = false
  }
}

const copyCode = () => {
  navigator.clipboard.writeText(recoveryCode.value)
  codeCopied.value = true
  showNotification.value = true
  
  // Hide notification after 3 seconds
  setTimeout(() => {
    showNotification.value = false
  }, 3000)
}

const finishRegistration = async () => {
    showConfirmModal.value = false
    // After successful registration, user is already authenticated
    // The register() function already set isAuthenticated = true and loaded the MasterKey
    // Just ensure RSA keys are loaded
    try {
      await authStore.ensureRSAKeys(authStore.masterKey);
      router.push({ name: 'Home' })
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

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

label {
  font-size: 0.9rem;
  font-weight: 500;
  color: var(--secondary-text-color);
}

.form-control {
  padding: 10px 12px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background-color: var(--background-color);
  color: var(--main-text-color);
  font-size: 1rem;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.form-control:focus {
  outline: none;
  border-color: var(--primary-color);
  box-shadow: 0 0 0 3px rgba(52, 152, 219, 0.1);
  background-color: var(--card-color);
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
  font-family: monospace;
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
