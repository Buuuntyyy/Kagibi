<template>
  <div class="auth-container">
    <div v-if="mode === 'login'">
      <h1>Se connecter</h1>
      <LoginComponent />
      <p>
        Pas encore de compte ? <a href="#" @click.prevent="mode = 'register'">S'inscrire</a>
      </p>
    </div>
    <div v-else>
      <h1>Créer un compte</h1>
      <RegisterComponent />
      <p>
        Déjà un compte ? <a href="#" @click.prevent="mode = 'login'">Se connecter</a>
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import LoginComponent from '../components/auth/loginComponent.vue'
import RegisterComponent from '../components/auth/registerComponent.vue'

const mode = ref('login') // 'login' or 'register'
</script>

<style scoped>
.auth-container {
  max-width: 400px;
  margin: 2rem auto;
  padding: 2rem;
  border: 1px solid #ccc;
  border-radius: 5px;
}
a {
    color: #42b983;
}
</style>