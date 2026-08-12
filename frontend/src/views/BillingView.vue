<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="billing-page">
    <div class="coming-soon-card">
      <div class="cs-icon">
        <svg viewBox="0 0 24 24" width="48" height="48" fill="none" stroke="currentColor" stroke-width="1.5">
          <path stroke-linecap="round" stroke-linejoin="round"
            d="M12 6v6l4 2m6-2a10 10 0 1 1-20 0 10 10 0 0 1 20 0z" />
        </svg>
      </div>

      <h1 class="cs-title">Abonnements</h1>
      <p class="cs-subtitle">Bientôt disponible</p>

      <p class="cs-description">
        Nous travaillons sur nos offres d'abonnement pour vous donner accès à plus de stockage,
        des fonctionnalités avancées et un support prioritaire.
        Revenez bientôt&nbsp;!
      </p>

      <div class="cs-features">
        <div class="cs-feature">
          <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z" />
          </svg>
          <span>Stockage étendu</span>
        </div>
        <div class="cs-feature">
          <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
          </svg>
          <span>Chiffrement de bout en bout avancé</span>
        </div>
        <div class="cs-feature">
          <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
            <circle cx="9" cy="7" r="4" />
            <path d="M23 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75" />
          </svg>
          <span>Organisations sans limite de membres</span>
        </div>
        <div class="cs-feature">
          <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10" />
            <path d="M12 8v4l3 3" />
          </svg>
          <span>Support prioritaire</span>
        </div>
      </div>

      <button class="btn-back" @click="router.go(-1)">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M19 12H5M12 19l-7-7 7-7" />
        </svg>
        Retour
      </button>
    </div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'

const router = useRouter()
</script>

<style scoped>
.billing-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 80vh;
  padding: 2rem;
}

.coming-soon-card {
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 20px;
  padding: 3rem 2.5rem;
  max-width: 480px;
  width: 100%;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1.25rem;
}

.cs-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: color-mix(in srgb, var(--primary-color) 12%, transparent);
  color: var(--primary-color);
  margin-bottom: 0.25rem;
}

.cs-title {
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--main-text-color);
  margin: 0;
}

.cs-subtitle {
  font-size: 0.95rem;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--primary-color);
  margin: 0;
  padding: 0.25rem 0.75rem;
  background: color-mix(in srgb, var(--primary-color) 12%, transparent);
  border-radius: 999px;
}

.cs-description {
  font-size: 0.95rem;
  color: var(--secondary-text-color);
  line-height: 1.6;
  margin: 0;
}

.cs-features {
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
  width: 100%;
  margin-top: 0.25rem;
}

.cs-feature {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0.6rem 0.9rem;
  background: var(--hover-background-color);
  border-radius: 10px;
  font-size: 0.88rem;
  color: var(--secondary-text-color);
}

.cs-feature svg {
  flex-shrink: 0;
  color: var(--primary-color);
  opacity: 0.8;
}

.btn-back {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  margin-top: 0.5rem;
  padding: 0.55rem 1.25rem;
  background: transparent;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  color: var(--secondary-text-color);
  font-size: 0.88rem;
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}

.btn-back:hover {
  background: var(--hover-background-color);
  color: var(--main-text-color);
}
</style>
