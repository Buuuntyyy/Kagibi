<template>
  <form @submit.prevent="submit">
    <div>
      <label>Nom d'utilisateur</label>
      <input v-model="username" type="text" required />
    </div>
    <div>
      <label>Email</label>
      <input v-model="email" type="email" required />
    </div>
    <div>
      <label>Mot de passe</label>
      <input v-model="password" type="password" required />
    </div>
    <button type="submit">S'inscrire</button>
    <p v-if="error" class="error">{{ error }}</p>
  </form>
</template>

<script setup>
import { ref } from 'vue'
import { useAuthStore } from '../../stores/auth'
import { useRouter } from 'vue-router'

const username = ref('')
const email = ref('')
const password = ref('')
const error = ref('')
const authStore = useAuthStore()
const router = useRouter()

const submit = async () => {
  error.value = ''
  try {
    const success = await authStore.login({email: email.value, password: password.value})
    if (success) {
      router.push({ name: 'Dashboard' })
    } else {
      error.value = 'Erreur lors de la connexion automatique après inscription.'
    }
  } catch (err) {
    error.value = 'Erreur lors de l\'inscription'
  }
}
</script>

<style scoped>
.error {
  color: red;
}
</style>
