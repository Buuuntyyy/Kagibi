<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <nav class="landing-nav">
    <div class="nav-container">
      <a href="https://kagibi.cloud" class="nav-logo">
        <img src="/Logo.png" alt="Kagibi Logo" width="40rem"/>
        <span>Kagibi</span>
      </a>

      <div class="nav-links">
        <a href="https://kagibi.cloud" class="nav-link">{{ t('nav.home') }}</a>
        <a href="https://kagibi.cloud/transfer" class="nav-link">{{ t('nav.transfer') }}</a>
        <a href="https://kagibi.cloud/security" class="nav-link">{{ t('nav.security') }}</a>

        <template v-if="authStore.isAuthenticated">
          <router-link to="/account" class="user-avatar-link" :title="authStore.user?.name || t('nav.myAccount')">
            <div class="user-avatar">
              <img
                v-if="authStore.user?.avatar_url"
                :src="authStore.user.avatar_url"
                :alt="authStore.user?.name"
                class="avatar-image"
                @error="(e) => e.target.style.display = 'none'"
              />
              <div v-else class="avatar-fallback">
                {{ getInitials(authStore.user?.name) }}
              </div>
            </div>
          </router-link>
          <a href="#" class="nav-link" @click.prevent="logout">{{ t('nav.logout') }}</a>
        </template>
        <template v-else>
          <router-link to="/login" class="nav-btn">{{ t('nav.login') }}</router-link>
        </template>
      </div>

      <!-- Mobile Menu Button -->
      <button class="mobile-menu-btn" @click="toggleMenu">
        <svg v-if="!isMenuOpen" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="3" y1="12" x2="21" y2="12"></line>
          <line x1="3" y1="6" x2="21" y2="6"></line>
          <line x1="3" y1="18" x2="21" y2="18"></line>
        </svg>
        <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="18" y1="6" x2="6" y2="18"></line>
          <line x1="6" y1="6" x2="18" y2="18"></line>
        </svg>
      </button>
    </div>

    <!-- Mobile Menu -->
    <div class="mobile-menu" :class="{ open: isMenuOpen }">
      <a href="https://kagibi.cloud" class="mobile-link" @click="closeMenu">{{ t('nav.home') }}</a>
      <a href="https://kagibi.cloud/transfer" class="mobile-link" @click="closeMenu">{{ t('nav.transfer') }}</a>
      <a href="https://kagibi.cloud/security" class="mobile-link" @click="closeMenu">{{ t('nav.security') }}</a>
      <template v-if="authStore.isAuthenticated">
        <router-link to="/account" class="mobile-link" @click="closeMenu">{{ t('nav.myAccount') }}</router-link>
        <a href="#" class="mobile-link" @click.prevent="logout; closeMenu()">{{ t('nav.logout') }}</a>
      </template>
      <template v-else>
        <router-link to="/login" class="mobile-btn" @click="closeMenu">{{ t('nav.login') }}</router-link>
      </template>
    </div>
  </nav>
</template>

<script setup>
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../../stores/auth'
import { useRouter } from 'vue-router'

const { t } = useI18n()
const authStore = useAuthStore()
const router = useRouter()
const isMenuOpen = ref(false)

const toggleMenu = () => { isMenuOpen.value = !isMenuOpen.value }
const closeMenu = () => { isMenuOpen.value = false }

const getInitials = (name) => {
  if (!name) return '?'
  return name.substring(0, 2).toUpperCase()
}

const logout = async () => {
  await authStore.logout()
  router.push('/login')
}
</script>

<style scoped>
.landing-nav {
  position: sticky;
  top: 0;
  background: var(--card-color);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid var(--border-color);
  z-index: 1000;
  opacity: 0.98;
}

.nav-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 1rem 2rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.nav-logo {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  text-decoration: none;
  color: var(--main-text-color);
  font-weight: 700;
  font-size: 1.25rem;
  transition: opacity 0.3s ease;
}

.nav-logo:hover {
  opacity: 0.7;
}

.nav-links {
  display: flex;
  align-items: center;
  gap: 2rem;
}

.nav-link {
  text-decoration: none;
  color: var(--secondary-text-color);
  font-weight: 500;
  font-size: 0.95rem;
  transition: color 0.3s ease;
  cursor: pointer;
}

.nav-link:hover {
  color: var(--primary-color);
}

.nav-btn {
  padding: 0.5rem 1.5rem;
  background: var(--primary-color);
  color: var(--card-color);
  text-decoration: none;
  border-radius: 8px;
  font-weight: 600;
  font-size: 0.95rem;
  transition: all 0.3s ease;
}

.nav-btn:hover {
  background: var(--accent-color);
  transform: translateY(-1px);
}

.user-avatar-link {
  text-decoration: none;
}

.user-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  overflow: hidden;
  border: 2px solid var(--border-color);
  cursor: pointer;
  transition: border-color 0.2s;
}

.user-avatar:hover {
  border-color: var(--primary-color);
}

.avatar-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.avatar-fallback {
  width: 100%;
  height: 100%;
  background: var(--primary-color);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.75rem;
  font-weight: 700;
}

.mobile-menu-btn {
  display: none;
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.5rem;
  color: var(--main-text-color);
}

.mobile-menu-btn svg {
  width: 24px;
  height: 24px;
}

.mobile-menu {
  display: none;
  flex-direction: column;
  gap: 0.5rem;
  padding: 0 2rem;
  background: var(--card-color);
  max-height: 0;
  overflow: hidden;
  transition: max-height 0.3s ease, padding 0.3s ease;
}

.mobile-menu.open {
  max-height: 300px;
  padding: 1rem 2rem;
  border-top: 1px solid var(--border-color);
}

.mobile-link {
  padding: 0.75rem;
  text-decoration: none;
  color: var(--secondary-text-color);
  font-weight: 500;
  border-radius: 8px;
  transition: all 0.3s ease;
  cursor: pointer;
}

.mobile-link:hover {
  color: var(--primary-color);
  background: var(--hover-background-color);
}

.mobile-btn {
  padding: 0.75rem;
  background: var(--primary-color);
  color: var(--card-color);
  text-decoration: none;
  border-radius: 8px;
  font-weight: 600;
  text-align: center;
  transition: background 0.3s ease;
}

.mobile-btn:hover {
  background: #0040CC;
}

@media (max-width: 768px) {
  .nav-links {
    display: none;
  }

  .mobile-menu-btn {
    display: block;
  }

  .mobile-menu {
    display: flex;
  }
}
</style>
