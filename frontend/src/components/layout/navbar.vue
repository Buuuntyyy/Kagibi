<template>
  <nav>
    <router-link to="/" class="brand">SaferCloud</router-link>
    <div class="nav-links">
      <router-link v-if="!authStore.isAuthenticated" to="/login">Connexion / Inscription</router-link>
      <a v-else @click="logout" href="#">Se déconnecter</a>
    </div>
  </nav>
</template>

<script setup>
import { useAuthStore } from '../../stores/auth'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()

const logout = () => {
  authStore.logout()
  router.push('/login')
}
</script>

<style scoped>
nav {
  height: 5vh;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  background-color: #333;
  color: white;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 1000;
}

.brand {
  font-weight: bold;
  font-size: 1.5rem;
  color: white;
  text-decoration: none;
}

.nav-links a {
  color: white;
  text-decoration: none;
  margin-left: 1rem;
}

.nav-links a:hover {
  text-decoration: underline;
}
</style>
