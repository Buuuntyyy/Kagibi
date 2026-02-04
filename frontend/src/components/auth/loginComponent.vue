<template>
  <form @submit.prevent="login" class="auth-form">
    <div class="form-group">
      <label>
        Email
      </label>
      <input v-model="email" type="email" required class="form-control" placeholder="votre@email.com" />
    </div>
    <div class="form-group">
      <label>
        Mot de passe
      </label>
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
      <span v-else>Se connecter</span>
    </button>
    <p v-if="error" class="error-message">{{ error }}</p>
  </form>
</template>

<script setup>
import { ref } from 'vue'
import { useAuthStore } from '../../stores/auth'
import { useRouter } from 'vue-router'

const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)
const showPassword = ref(false)
const authStore = useAuthStore()
const router = useRouter()

const login = async () => {
  error.value = ''
  loading.value = true
  try {
    const success = await authStore.login({email: email.value, password: password.value})
    if (success) {
      router.push({ name: 'Home' })
    } else {
      error.value = 'Identifiants invalides'
    }
  } catch (e) {
    console.error("Login error details:", e)
    // Afficher le message d'erreur détaillé du store
    if (e.message) {
      error.value = e.message
    } else {
      error.value = 'Une erreur est survenue. Veuillez réessayer.'
    }
  } finally {
    loading.value = false
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
}

.btn-submit:hover:not(:disabled) {
  background-color: var(--accent-color);
}

.btn-submit:disabled {
  opacity: 0.7;
  cursor: not-allowed;
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
</style>
<!-- frontend/src/components/Auth/Login.vue -->
