<template>
  <nav>
    <router-link to="/" class="brand">SaferCloud</router-link>
    <SearchBar v-if="authStore.isAuthenticated" />
    <div class="nav-links">
      <button @click="themeStore.toggleTheme" class="theme-toggle" :title="themeStore.theme === 'light' ? 'Mode sombre' : 'Mode clair'">
        <svg v-if="themeStore.theme === 'light'" class="icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M12 3c-4.97 0-9 4.03-9 9s4.03 9 9 9 9-4.03 9-9c0-.46-.04-.92-.1-1.36-.98 1.37-2.58 2.26-4.4 2.26-2.98 0-5.4-2.42-5.4-5.4 0-1.81.89-3.42 2.26-4.4-.44-.02-.9-.02-1.36-.02z" fill="currentColor"/>
        </svg>
        <svg v-else class="icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M6.76 4.84l-1.8-1.79-1.41 1.41 1.79 1.79 1.42-1.41zM4 10.5H1v2h3v-2zm9-9.95h-2V3.5h2V.55zm7.45 3.91l-1.41-1.41-1.79 1.79 1.41 1.41 1.79-1.79zm-3.21 13.7l1.79 1.8 1.41-1.41-1.8-1.79-1.4 1.4zM20 10.5v2h3v-2h-3zm-8-5c-3.31 0-6 2.69-6 6s2.69 6 6 6 6-2.69 6-6-2.69-6-6-6zm-1 16.95h2V19.5h-2v2.95zm-7.45-3.91l1.41 1.41 1.79-1.8-1.41-1.41-1.79 1.8z" fill="currentColor"/>
        </svg>
      </button>
      <router-link v-if="!authStore.isAuthenticated" to="/login">Connexion / Inscription</router-link>
      <template v-else>
        <router-link to="/account">Mon Compte</router-link>
        <a @click.prevent="logout" href="#">Se déconnecter</a>
      </template>
    </div>
  </nav>
</template>

<script setup>
import { useAuthStore } from '../../stores/auth'
import { useThemeStore } from '../../stores/theme'
import { useRouter } from 'vue-router'
import SearchBar from '../bar/searchBar.vue'

const authStore = useAuthStore()
const themeStore = useThemeStore()
const router = useRouter()

const logout = async () => {
  await authStore.logout()
  router.push({ name: 'Login' })
}
</script>

<style scoped>
nav {
  height: 60px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 1rem;
  background-color: var(--background-color);
  color: var(--main-text-color);
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 1000;
  box-sizing: border-box;
  /* border-bottom: 1px solid var(--border-color); */
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
  cursor: pointer;
  margin-left: 1rem;
  color: var(--main-text-color);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 8px;
  border-radius: 50%;
}

.theme-toggle:hover {
  background-color: rgba(60,64,67,0.08);
}

.icon-svg {
  width: 24px;
  height: 24px;
}
</style>
