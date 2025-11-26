<template>
  <div v-if="!recoveryCode">
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
      <button type="submit" :disabled="loading">{{ loading ? 'Création...' : "S'inscrire" }}</button>
      <p v-if="error" class="error">{{ error }}</p>
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
    <button @click="copyCode">Copier le code</button>
    <button class="btn-continue" @click="finishRegistration">Continuer vers le Dashboard</button>
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
    error.value = 'Erreur lors de l\'inscription'
  } finally {
    loading.value = false
  }
}

const copyCode = () => {
  navigator.clipboard.writeText(recoveryCode.value)
  alert("Code copié !")
}

const finishRegistration = async () => {
    // Après une inscription réussie, connecter automatiquement l'utilisateur
    const loginSuccess = await authStore.login({email: email.value, password: password.value})
    if (loginSuccess) {
      router.push({ name: 'Dashboard' })
    } else {
      error.value = 'Erreur lors de la connexion automatique après inscription.'
    }
}
</script>

<style scoped>
.error {
  color: red;
}
.recovery-display {
  text-align: center;
}
.alert-box {
  background-color: #fff3cd;
  color: #856404;
  padding: 1rem;
  border: 1px solid #ffeeba;
  border-radius: 4px;
  margin-bottom: 1rem;
  font-size: 0.9rem;
}
.code-box {
  background-color: #f8f9fa;
  border: 1px solid #dee2e6;
  padding: 1rem;
  font-family: monospace;
  font-size: 1.1rem;
  word-break: break-all;
  margin-bottom: 1rem;
  border-radius: 4px;
  user-select: all;
}
.btn-continue {
  background-color: #42b983;
  color: white;
  margin-top: 1rem;
  width: 100%;
}
</style>
