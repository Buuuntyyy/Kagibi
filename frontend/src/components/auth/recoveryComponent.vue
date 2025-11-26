<template>
  <div class="recovery-container">
    <h2>Récupération de compte</h2>
    <p class="info-text">
      Entrez votre email et votre code de récupération pour réinitialiser votre mot de passe.
    </p>
    
    <form @submit.prevent="handleRecovery">
      <div class="form-group">
        <label>Email</label>
        <input type="email" v-model="email" required />
      </div>
      
      <div class="form-group">
        <label>Code de récupération</label>
        <textarea v-model="recoveryCode" required placeholder="Collez votre code ici..." rows="3"></textarea>
      </div>
      
      <div class="form-group">
        <label>Nouveau mot de passe</label>
        <input type="password" v-model="newPassword" required minlength="8" />
      </div>
      
      <div class="form-group">
        <label>Confirmer le mot de passe</label>
        <input type="password" v-model="confirmPassword" required minlength="8" />
      </div>
      
      <button type="submit" :disabled="loading">
        {{ loading ? 'Traitement...' : 'Réinitialiser le mot de passe' }}
      </button>
    </form>
    
    <button class="btn-cancel" @click="$emit('cancel')">Annuler</button>
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

const handleRecovery = async () => {
  if (newPassword.value !== confirmPassword.value) {
    alert("Les mots de passe ne correspondent pas.")
    return
  }
  
  loading.value = true
  try {
    await authStore.recoverAccount(email.value, recoveryCode.value.trim(), newPassword.value)
    alert("Compte récupéré avec succès ! Vous pouvez maintenant vous connecter avec votre nouveau mot de passe.")
    emit('success')
  } catch (error) {
    console.error(error)
    alert(error.message || "Erreur lors de la récupération.")
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.recovery-container {
  text-align: left;
}
.info-text {
  font-size: 0.9rem;
  color: #666;
  margin-bottom: 1rem;
}
.form-group {
  margin-bottom: 1rem;
}
label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: bold;
}
input, textarea {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid #ccc;
  border-radius: 4px;
  box-sizing: border-box;
}
button {
  width: 100%;
  padding: 0.75rem;
  background-color: #42b983;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-weight: bold;
  margin-top: 1rem;
}
button:disabled {
  background-color: #a8d8c4;
  cursor: not-allowed;
}
.btn-cancel {
  background-color: transparent;
  color: #666;
  border: 1px solid #ccc;
  margin-top: 0.5rem;
}
</style>
