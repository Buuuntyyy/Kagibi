<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <nav class="landing-nav">
    <div class="nav-container">
      <router-link to="/" class="nav-logo">
        <img src="/Logo.png" alt="Kagibi Logo" width="40rem"/>
        <span>Kagibi</span>
      </router-link>
      
      <div class="nav-links">
        <router-link to="/" class="nav-link">Accueil</router-link>
        <router-link to="/transfer" class="nav-link">Transfert P2P</router-link>
        <router-link to="/compare" class="nav-link">Comparatif</router-link>
        <router-link to="/login" class="nav-btn">Connexion</router-link>
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
      <router-link to="/" class="mobile-link" @click="closeMenu">Accueil</router-link>
      <router-link to="/transfer" class="mobile-link" @click="closeMenu">Transfert P2P</router-link>
      <router-link to="/compare" class="mobile-link" @click="closeMenu">Comparatif</router-link>
      <router-link to="/login" class="mobile-btn" @click="closeMenu">Connexion</router-link>
    </div>
  </nav>
</template>

<script setup>
import { ref } from 'vue'

const isMenuOpen = ref(false)

const toggleMenu = () => {
  isMenuOpen.value = !isMenuOpen.value
}

const closeMenu = () => {
  isMenuOpen.value = false
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

.nav-logo svg {
  width: 28px;
  height: 28px;
  color: var(--primary-color);
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
  position: relative;
}

.nav-link:hover {
  color: var(--primary-color);
}

.nav-link.router-link-active {
  color: var(--primary-color);
}

.nav-link.router-link-active::after {
  content: '';
  position: absolute;
  bottom: -8px;
  left: 0;
  right: 0;
  height: 2px;
  background: var(--primary-color);
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
}

.mobile-link:hover,
.mobile-link.router-link-active {
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
