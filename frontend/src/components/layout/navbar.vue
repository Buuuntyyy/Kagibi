<template>
  <nav>
    <router-link to="/" class="brand">SaferCloud</router-link>
    <div class="nav-links">
      <button @click="themeStore.toggleTheme" class="theme-toggle">
        {{ themeStore.theme === 'light' ? '🌙' : '☀️' }}
      </button>
      <router-link v-if="!authStore.isAuthenticated" to="/login">Connexion / Inscription</router-link>
      <a v-else @click="logout" href="#">Se déconnecter</a>
    </div>
  </nav>
</template>

<script setup>
import { useAuthStore } from '../../stores/auth'
import { useThemeStore } from '../../stores/theme'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const themeStore = useThemeStore()
const router = useRouter()

const logout = () => {
  authStore.logout()
  router.push({ name: 'Login' })
}
</script>

<style scoped>
nav {
  height: 5vh;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  background-color: var(--card-color);
  color: var(--main-text-color);
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 1000;
  border-bottom: 1px solid var(--border-color);
}

.brand {
  font-weight: bold;
  font-size: 1.5rem;
  color: var(--main-text-color);
  text-decoration: none;
}

.nav-links {
  display: flex;
  align-items: center;
}

.nav-links a {
  color: var(--main-text-color);
  text-decoration: none;
  margin-left: 1rem;
}

.nav-links a:hover {
  text-decoration: underline;
}

.theme-toggle {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  margin-left: 1rem;
  color: var(--main-text-color);
}
</style>
