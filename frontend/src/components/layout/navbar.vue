<template>
  <nav>
    <router-link to="/dashboard" class="brand">
      <img src="/Logo.png" alt="SkyDrive Logo" class="brand-logo"/>
      <span>FoxEarth</span>
    </router-link>
    <SearchBar v-if="authStore.isAuthenticated" />
    <div class="nav-links">
      <a
        v-if="buyMeACoffeeUrl"
        :href="buyMeACoffeeUrl"
        target="_blank"
        rel="noopener noreferrer"
        class="support-link"
      >
        {{ t('file.supportProject') }}
      </a>
      <button @click="showHelpDialog = true" class="theme-toggle" title="Aide & Support">
        <HelpCircle class="icon-svg" :size="24" :stroke-width="2" />
      </button>
      <LanguageSwitcher />
      <button @click="themeStore.toggleTheme" class="theme-toggle" :title="themeStore.theme === 'light' ? t('nav.darkMode') : t('nav.lightMode')">
        <svg v-if="themeStore.theme === 'light'" class="icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M12 3c-4.97 0-9 4.03-9 9s4.03 9 9 9 9-4.03 9-9c0-.46-.04-.92-.1-1.36-.98 1.37-2.58 2.26-4.4 2.26-2.98 0-5.4-2.42-5.4-5.4 0-1.81.89-3.42 2.26-4.4-.44-.02-.9-.02-1.36-.02z" fill="currentColor"/>
        </svg>
        <svg v-else class="icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M6.76 4.84l-1.8-1.79-1.41 1.41 1.79 1.79 1.42-1.41zM4 10.5H1v2h3v-2zm9-9.95h-2V3.5h2V.55zm7.45 3.91l-1.41-1.41-1.79 1.79 1.41 1.41 1.79-1.79zm-3.21 13.7l1.79 1.8 1.41-1.41-1.8-1.79-1.4 1.4zM20 10.5v2h3v-2h-3zm-8-5c-3.31 0-6 2.69-6 6s2.69 6 6 6 6-2.69 6-6-2.69-6-6-6zm-1 16.95h2V19.5h-2v2.95zm-7.45-3.91l1.41 1.41 1.79-1.8-1.41-1.41-1.79 1.8z" fill="currentColor"/>
        </svg>
      </button>
      <router-link v-if="!authStore.isAuthenticated" to="/login">{{ t('nav.login') }}</router-link>
      <template v-else>
        <router-link to="/account" class="user-avatar-link" :title="authStore.user?.name || t('nav.myAccount')">
          <div class="user-avatar">
            <img
              v-if="authStore.user?.avatar_url"
              :src="authStore.user.avatar_url"
              :alt="authStore.user?.name"
              class="avatar-image"
              @error="handleImageError"
            />
            <div class="avatar-fallback" :style="{ display: authStore.user?.avatar_url ? 'none' : 'flex' }">
              {{ getInitials(authStore.user?.name) }}
            </div>
          </div>
        </router-link>
        <a @click.prevent="logout" href="#">{{ t('nav.logout') }}</a>
      </template>
    </div>
    <HelpDialog v-model:isOpen="showHelpDialog" />
  </nav>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useAuthStore } from '../../stores/auth'
import { useThemeStore } from '../../stores/theme'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import SearchBar from '../bar/searchBar.vue'
import LanguageSwitcher from '../LanguageSwitcher.vue'
import HelpDialog from '../HelpDialog.vue'
import { HelpCircle } from 'lucide-vue-next'

const { t } = useI18n()

const authStore = useAuthStore()
const themeStore = useThemeStore()
const router = useRouter()

const showHelpDialog = ref(false)

const buyMeACoffeeUrl = computed(() => {
  const runtimeUrl = typeof window !== 'undefined' ? window.__APP_CONFIG__?.buyMeACoffeeUrl : ''
  return runtimeUrl || import.meta.env.VITE_BUY_ME_A_COFFEE_URL || ''
})

const logout = async () => {
  await authStore.logout()
  router.push({ name: 'Login' })
}

const getInitials = (name) => {
  if (!name) return '?'
  return name.substring(0, 2).toUpperCase()
}

const handleImageError = (event) => {
  // Hide the broken image and show fallback initials
  event.target.style.display = 'none'
  const fallback = event.target.nextElementSibling
  if (fallback) fallback.style.display = 'flex'
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
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  width: 256px;
  padding: 16px;
  flex-shrink: 0;
  box-sizing: border-box;
  margin-left: -1rem;
}

.brand-logo {
  height: 36px;
  width: auto;
}

.nav-links {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.nav-links a {
  color: var(--main-text-color);
  text-decoration: none;
}

.support-link {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.3rem 0.65rem;
  border-radius: 999px;
  border: 1px solid var(--border-color);
  background: var(--card-color);
  font-size: 0.82rem;
  font-weight: 600;
  white-space: nowrap;
}

.nav-links a:hover {
  text-decoration: underline;
}

.support-link:hover {
  text-decoration: none !important;
  background: var(--hover-background-color);
}

.theme-toggle {
  background: none;
  border: none;
  cursor: pointer;
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

.user-avatar-link {
  text-decoration: none !important;
}

.user-avatar-link:hover {
  text-decoration: none !important;
  opacity: 0.9;
}

.user-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: var(--card-color, #f0f0f0);
  transition: transform 0.3s ease, box-shadow 0.3s ease;
  cursor: pointer;
  border: 2px solid var(--border-color, #e0e0e0);
}

.user-avatar:hover {
  transform: scale(1.05);
  box-shadow: 0 4px 12px rgba(107, 127, 215, 0.3);
  border-color: var(--primary-color, #6B7FD7);
}

.avatar-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: opacity 0.3s ease;
}

.avatar-fallback {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: 600;
  font-size: 0.875rem;
  letter-spacing: 0.5px;
  background: linear-gradient(135deg, var(--primary-color, #6B7FD7) 0%, var(--secondary-color, #9370DB) 100%);
  border-radius: 50%;
}
</style>
