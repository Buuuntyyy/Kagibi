<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="upgrade-page">
    <div class="page-header">
      <div class="header-content">
        <button class="btn-back" @click="router.go(-1)">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M19 12H5M12 19l-7-7 7-7"/>
          </svg>
          {{ t('common.back') }}
        </button>
        <h1>{{ t('upgrade.title') }}</h1>
      </div>
      <p class="subtitle">{{ t('upgrade.subtitle') }}</p>
    </div>

    <!-- Coming soon notice -->
    <div class="coming-soon-banner">
      <div class="cs-icon">
        <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
          <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
        </svg>
      </div>
      <div class="cs-text">
        <strong>{{ t('upgrade.comingSoonTitle') }}</strong>
        <span>{{ t('upgrade.comingSoonDesc') }}</span>
      </div>
      <a v-if="buyMeACoffeeUrl" :href="buyMeACoffeeUrl" target="_blank" rel="noopener noreferrer" class="btn-coffee">
        ☕ {{ t('upgrade.supportUs') }}
      </a>
    </div>

    <!-- Current plan -->
    <div v-if="billingStore.currentPlan" class="current-plan-row">
      <span class="current-plan-label">{{ t('upgrade.currentPlan') }}</span>
      <span class="current-plan-badge">{{ billingStore.currentPlan.name || t('billing.free') }}</span>
    </div>

    <!-- Plans grid -->
    <div class="plans-grid">
      <div
        v-for="plan in billingStore.plans"
        :key="plan.code"
        class="plan-card"
        :class="{ current: currentPlanCode === plan.code, featured: plan.code === 'business' }"
      >
        <div v-if="plan.code === 'business'" class="plan-popular-badge">{{ t('upgrade.popular') }}</div>
        <div v-if="currentPlanCode === plan.code" class="plan-current-tag">{{ t('upgrade.currentPlan') }}</div>
        <h3 class="plan-name">{{ plan.name }}</h3>
        <div class="plan-price">
          <span class="price-amount">{{ formatPlanPrice(plan) }}</span>
          <span class="price-period">{{ formatPlanPeriod(plan) }}</span>
        </div>
        <ul class="plan-features">
          <li v-for="f in (plan.features || [])" :key="f">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41L9 16.17z"/></svg>
            {{ f }}
          </li>
        </ul>
        <button
          class="btn-plan"
          :class="plan.code === 'business' ? 'btn-plan-primary' : 'btn-plan-outline'"
          disabled
        >
          {{ currentPlanCode === plan.code ? t('upgrade.currentLabel') : t('upgrade.comingSoonBtn') }}
        </button>
      </div>
    </div>

    <!-- Self-hosted note -->
    <p v-if="billingStore.isSelfHosted" class="selfhosted-note">
      {{ t('upgrade.selfHostedNote') }}
    </p>

    <!-- FAQ -->
    <div class="faq-section">
      <h2>{{ t('upgrade.faqTitle') }}</h2>
      <div class="faq-grid">
        <div v-for="faq in faqs" :key="faq.q" class="faq-item">
          <h4>{{ faq.q }}</h4>
          <p>{{ faq.a }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useBillingStore } from '../stores/billing'

const { t, locale } = useI18n()
const router = useRouter()
const billingStore = useBillingStore()

const currentPlanCode = computed(() => billingStore.currentPlan?.code || 'free')

const buyMeACoffeeUrl = computed(() => {
  const raw = (typeof window !== 'undefined' ? window.__APP_CONFIG__?.buyMeACoffeeUrl : '')
    || import.meta.env.VITE_BUY_ME_A_COFFEE_URL || ''
  if (!raw) return ''
  return /^https?:\/\//i.test(raw) ? raw : `https://${raw}`
})

const CURRENCY_SYMBOLS = { EUR: '€', USD: '$', GBP: '£' }

const formatPlanPrice = (plan) => {
  const symbol = CURRENCY_SYMBOLS[plan.currency] || plan.currency
  const amount = (plan.price_ttc_cents / 100).toLocaleString(locale.value, {
    minimumFractionDigits: plan.price_ttc_cents % 100 === 0 ? 0 : 2,
    maximumFractionDigits: 2,
  })
  return `${amount}${symbol}`
}

const formatPlanPeriod = (plan) => {
  if (plan.billing_model === 'payg') return `/To/${t('upgrade.month')}`
  return `/${t('upgrade.month')}`
}

const faqs = computed(() => [
  { q: t('upgrade.faq1q'), a: t('upgrade.faq1a') },
  { q: t('upgrade.faq2q'), a: t('upgrade.faq2a') },
  { q: t('upgrade.faq3q'), a: t('upgrade.faq3a') },
  { q: t('upgrade.faq4q'), a: t('upgrade.faq4a') },
])

onMounted(async () => {
  await Promise.all([
    billingStore.currentPlan ? Promise.resolve() : billingStore.fetchCurrentPlan(),
    billingStore.fetchPlans(),
  ])
})
</script>

<style scoped>
.upgrade-page {
  padding: 2rem;
  max-width: 1100px;
  margin: 0 auto;
  overflow-y: auto;
  height: 100%;
}

/* Header */
.page-header {
  margin-bottom: 1.5rem;
}

.header-content {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 0.4rem;
}

.btn-back {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  background: none;
  border: none;
  color: var(--secondary-text-color, #aaa);
  cursor: pointer;
  font-size: 0.875rem;
  padding: 0.3rem 0.5rem;
  border-radius: 6px;
  transition: color 0.15s, background 0.15s;
}
.btn-back:hover {
  color: var(--main-text-color);
  background: var(--hover-background-color);
}

h1 {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--main-text-color);
}

.subtitle {
  margin: 0;
  color: var(--secondary-text-color);
  font-size: 0.9rem;
}

/* Coming soon banner */
.coming-soon-banner {
  display: flex;
  align-items: center;
  gap: 0.875rem;
  background: var(--hover-background-color);
  border: 1px solid var(--primary-color);
  border-radius: 10px;
  padding: 0.875rem 1.125rem;
  margin-bottom: 1.5rem;
}

.cs-icon {
  color: var(--primary-color);
  flex-shrink: 0;
  display: flex;
}

.cs-text {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  flex: 1;
}

.cs-text strong {
  font-size: 0.875rem;
  color: var(--main-text-color);
}

.cs-text span {
  font-size: 0.8rem;
  color: var(--secondary-text-color);
}

.btn-coffee {
  flex-shrink: 0;
  padding: 0.45rem 1rem;
  border-radius: 7px;
  background: var(--card-color);
  border: 1px solid var(--primary-color);
  color: var(--main-text-color);
  font-size: 0.8rem;
  font-weight: 500;
  text-decoration: none;
  transition: background 0.15s, color 0.15s;
}
.btn-coffee:hover {
  background: var(--primary-color);
  color: #fff;
}

/* Current plan row */
.current-plan-row {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  margin-bottom: 1.25rem;
  font-size: 0.875rem;
}

.current-plan-label {
  color: var(--secondary-text-color);
}

.current-plan-badge {
  background: var(--hover-background-color);
  border: 1px solid var(--border-color);
  border-radius: 20px;
  padding: 0.2rem 0.7rem;
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--main-text-color);
}

/* Plans grid */
.plans-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 1.25rem;
  margin-bottom: 2rem;
}

.plan-card {
  position: relative;
  background: var(--card-color);
  border: 1.5px solid var(--border-color);
  border-radius: 12px;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  transition: border-color 0.2s;
}

.plan-card.featured {
  border-color: var(--primary-color);
  box-shadow: 0 0 0 1px var(--primary-color), 0 8px 24px rgba(0, 0, 0, 0.12);
}

.plan-card.current {
  border-color: var(--primary-color);
}

.plan-popular-badge {
  position: absolute;
  top: -12px;
  left: 50%;
  transform: translateX(-50%);
  background: var(--primary-color);
  color: #fff;
  font-size: 0.7rem;
  font-weight: 700;
  padding: 0.2rem 0.75rem;
  border-radius: 20px;
  white-space: nowrap;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.plan-current-tag {
  position: absolute;
  top: 0.6rem;
  right: 0.75rem;
  font-size: 0.7rem;
  font-weight: 600;
  color: var(--secondary-text-color);
  background: var(--hover-background-color);
  border-radius: 4px;
  padding: 0.15rem 0.45rem;
}

.plan-name {
  margin: 0;
  font-size: 1rem;
  font-weight: 700;
  color: var(--main-text-color);
}

.plan-price {
  display: flex;
  align-items: baseline;
  gap: 0.2rem;
}

.price-amount {
  font-size: 2rem;
  font-weight: 800;
  color: var(--main-text-color);
  line-height: 1;
}

.price-period {
  font-size: 0.8rem;
  color: var(--secondary-text-color, #aaa);
}

.plan-features {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  flex: 1;
}

.plan-features li {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  font-size: 0.825rem;
  color: var(--secondary-text-color, #ccc);
  line-height: 1.4;
}

.plan-features li svg {
  flex-shrink: 0;
  color: var(--primary-color, #667eea);
  margin-top: 1px;
}

.btn-plan {
  width: 100%;
  padding: 0.6rem;
  border-radius: 8px;
  font-size: 0.85rem;
  font-weight: 600;
  cursor: not-allowed;
  transition: opacity 0.15s;
  opacity: 0.7;
}

.btn-plan-outline {
  background: none;
  border: 1.5px solid var(--border-color, #444);
  color: var(--secondary-text-color, #aaa);
}

.btn-plan-primary {
  background: var(--primary-color);
  border: none;
  color: #fff;
}

/* Self-hosted note */
.selfhosted-note {
  text-align: center;
  font-size: 0.8rem;
  color: var(--secondary-text-color);
  margin-bottom: 2rem;
  padding: 0.75rem;
  background: var(--hover-background-color);
  border-radius: 8px;
  border: 1px solid var(--border-color);
}

/* FAQ */
.faq-section {
  margin-top: 2rem;
  padding-top: 2rem;
  border-top: 1px solid var(--border-color);
}

.faq-section h2 {
  margin: 0 0 1.25rem 0;
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--main-text-color);
}

.faq-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 1rem;
}

.faq-item {
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 10px;
  padding: 1rem 1.125rem;
}

.faq-item h4 {
  margin: 0 0 0.4rem 0;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--main-text-color);
}

.faq-item p {
  margin: 0;
  font-size: 0.825rem;
  color: var(--secondary-text-color, #aaa);
  line-height: 1.55;
}

@media (max-width: 768px) {
  .upgrade-page {
    padding: 1.25rem;
  }

  .plans-grid {
    grid-template-columns: 1fr;
  }

  .coming-soon-banner {
    flex-wrap: wrap;
  }
}
</style>
