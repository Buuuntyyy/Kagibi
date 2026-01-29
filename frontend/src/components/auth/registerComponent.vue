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
        <input v-model="password" type="password" required class="form-control" placeholder="••••••••" />
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
      <button class="btn-submit" @click="finishRegistration">Continuer vers le Dashboard</button>
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
  alert("Code copié !")
}

const finishRegistration = async () => {
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
</style>
