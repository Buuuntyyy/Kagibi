<template>
  <form @submit.prevent="login" class="auth-form">
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
    error.value = 'Une erreur est survenue'
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
</style>
<!-- frontend/src/components/Auth/Login.vue -->
