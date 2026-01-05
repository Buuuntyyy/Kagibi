<template>
  <div class="recovery-container">
    <h2>Récupération de compte</h2>
    <p class="info-text">
      Entrez votre email et votre code de récupération pour réinitialiser votre mot de passe.
    </p>
    
    <form @submit.prevent="handleRecovery" class="auth-form">
      <div class="form-group">
        <label>Email</label>
        <input type="email" v-model="email" required class="form-control" placeholder="votre@email.com" />
      </div>
      
      <div class="form-group">
        <label>Code de récupération</label>
        <textarea v-model="recoveryCode" required class="form-control" placeholder="Collez votre code ici..." rows="3"></textarea>
      </div>
      
      <div class="form-group">
        <label>Nouveau mot de passe</label>
        <input type="password" v-model="newPassword" required minlength="8" class="form-control" placeholder="••••••••" />
      </div>
      
      <div class="form-group">
        <label>Confirmer le mot de passe</label>
        <input type="password" v-model="confirmPassword" required minlength="8" class="form-control" placeholder="••••••••" />
      </div>
      
      <button type="submit" class="btn-submit" :disabled="loading">
        <span v-if="loading" class="spinner"></span>
        <span v-else>Réinitialiser le mot de passe</span>
      </button>
      
      <p v-if="error" class="error-message">{{ error }}</p>
    </form>
    
    <button class="btn-secondary" @click="$emit('cancel')">Annuler</button>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useAuthStore } from '../../stores/auth'

const emit = defineEmits(['cancel', 'success'])
const authStore = useAuthStore()

const email = ref('')
const recoveryCode = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const loading = ref(false)
const error = ref('')

const handleRecovery = async () => {
  error.value = ''
  if (newPassword.value !== confirmPassword.value) {
    error.value = "Les mots de passe ne correspondent pas."
    return
  }
  
  loading.value = true
  try {
    await authStore.recoverAccount(email.value, recoveryCode.value.trim(), newPassword.value)
    alert("Compte récupéré avec succès ! Vous pouvez maintenant vous connecter avec votre nouveau mot de passe.")
    emit('success')
  } catch (err) {
    console.error(err)
    error.value = err.message || "Erreur lors de la récupération."
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.recovery-container {
  text-align: left;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

h2 {
  margin: 0;
  font-size: 1.5rem;
  color: var(--main-text-color);
}

.info-text {
  font-size: 0.9rem;
  color: var(--secondary-text-color);
  margin: 0;
  line-height: 1.4;
}

.auth-form {
  display: flex;
  flex-direction: column;
  gap: 1.2rem;
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
  width: 100%;
  box-sizing: border-box;
}

.form-control:focus {
  outline: none;
  border-color: var(--primary-color);
  box-shadow: 0 0 0 3px rgba(52, 152, 219, 0.1);
  background-color: var(--card-color);
}

textarea.form-control {
  resize: vertical;
  font-family: monospace;
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
  color: var(--secondary-text-color);
  padding: 10px;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 500;
  transition: all 0.2s;
  width: 100%;
}

.btn-secondary:hover {
  background-color: var(--background-color);
  color: var(--main-text-color);
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

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}
</style>
