<template>
  <form @submit.prevent="login">
    <div>
      <label>Email</label>
      <input v-model="email" type="email" required />
    </div>
    <div>
      <label>Mot de passe</label>
      <input v-model="password" type="password" required />
    </div>
    <button type="submit">Se connecter</button>
    <p v-if="error" class="error">{{ error }}</p>
  </form>
</template>

<script setup>
import { ref } from 'vue'
import { useAuthStore } from '../../stores/auth'
import { useRouter } from 'vue-router'

const email = ref('')
const password = ref('')
const error = ref('')
const authStore = useAuthStore()
const router = useRouter()

const login = async () => {
  error.value = ''
  const success = await authStore.login({email: email.value, password: password.value})
  if (success) {
    router.push({ name: 'Dashboard' })
  } else {
    error.value = 'Identifiants invalides'
  }
}
</script>
<!-- frontend/src/components/Auth/Login.vue -->
