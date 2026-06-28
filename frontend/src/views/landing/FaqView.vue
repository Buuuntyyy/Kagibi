<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="faq-page">
    <LandingNav />

    <!-- Hero -->
    <section class="faq-hero">
      <div class="container">
        <div class="hero-eyebrow">{{ t('landing.faq.eyebrow') }}</div>
        <h1>
          {{ t('landing.faq.title1') }}<br>
          <span class="highlight">{{ t('landing.faq.titleHighlight') }}</span>
        </h1>
        <p class="hero-sub">{{ t('landing.faq.heroSub') }}</p>
      </div>
    </section>

    <!-- FAQ Categories -->
    <section class="faq-section">
      <div class="container">

        <!-- Catégorie 1: Général & Souveraineté -->
        <div class="faq-category">
          <div class="cat-header">
            <div class="cat-icon cat-icon--blue">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="20" height="20">
                <circle cx="12" cy="12" r="10"/><path d="M12 8v4l3 3"/>
              </svg>
            </div>
            <div>
              <div class="cat-label">{{ t('landing.faq.cat1Label') }}</div>
              <h2 class="cat-title">{{ t('landing.faq.cat1Title') }}</h2>
            </div>
          </div>

          <div class="faq-list">
            <div
              v-for="(item, i) in faqGeneral"
              :key="'g'+i"
              class="faq-item"
              :class="{ open: openItems['g'+i] }"
              @click="toggle('g'+i)"
            >
              <div class="faq-question">
                <span>{{ item.q }}</span>
                <svg class="faq-chevron" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                  <polyline points="6 9 12 15 18 9"/>
                </svg>
              </div>
              <div class="faq-answer">
                <p>{{ item.a }}</p>
              </div>
            </div>
          </div>
        </div>

        <!-- Catégorie 2: Sécurité & Chiffrement -->
        <div class="faq-category">
          <div class="cat-header">
            <div class="cat-icon cat-icon--green">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="20" height="20">
                <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
              </svg>
            </div>
            <div>
              <div class="cat-label">{{ t('landing.faq.cat2Label') }}</div>
              <h2 class="cat-title">{{ t('landing.faq.cat2Title') }}</h2>
            </div>
          </div>

          <div class="faq-list">
            <div
              v-for="(item, i) in faqSecurity"
              :key="'s'+i"
              class="faq-item"
              :class="{ open: openItems['s'+i] }"
              @click="toggle('s'+i)"
            >
              <div class="faq-question">
                <span>{{ item.q }}</span>
                <svg class="faq-chevron" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                  <polyline points="6 9 12 15 18 9"/>
                </svg>
              </div>
              <div class="faq-answer">
                <p>{{ item.a }}</p>
              </div>
            </div>
          </div>
        </div>

        <!-- Catégorie 3: Fonctionnalités -->
        <div class="faq-category">
          <div class="cat-header">
            <div class="cat-icon cat-icon--purple">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="20" height="20">
                <rect x="2" y="3" width="20" height="14" rx="2"/><path d="M8 21h8m-4-4v4"/>
              </svg>
            </div>
            <div>
              <div class="cat-label">{{ t('landing.faq.cat3Label') }}</div>
              <h2 class="cat-title">{{ t('landing.faq.cat3Title') }}</h2>
            </div>
          </div>

          <div class="faq-list">
            <div
              v-for="(item, i) in faqFeatures"
              :key="'f'+i"
              class="faq-item"
              :class="{ open: openItems['f'+i] }"
              @click="toggle('f'+i)"
            >
              <div class="faq-question">
                <span>{{ item.q }}</span>
                <svg class="faq-chevron" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                  <polyline points="6 9 12 15 18 9"/>
                </svg>
              </div>
              <div class="faq-answer">
                <p>{{ item.a }}</p>
              </div>
            </div>
          </div>
        </div>

        <!-- Catégorie 4: Valeurs & Engagements -->
        <div class="faq-category">
          <div class="cat-header">
            <div class="cat-icon cat-icon--orange">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="20" height="20">
                <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"/>
              </svg>
            </div>
            <div>
              <div class="cat-label">{{ t('landing.faq.cat4Label') }}</div>
              <h2 class="cat-title">{{ t('landing.faq.cat4Title') }}</h2>
            </div>
          </div>

          <div class="faq-list">
            <div
              v-for="(item, i) in faqValues"
              :key="'v'+i"
              class="faq-item"
              :class="{ open: openItems['v'+i] }"
              @click="toggle('v'+i)"
            >
              <div class="faq-question">
                <span>{{ item.q }}</span>
                <svg class="faq-chevron" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                  <polyline points="6 9 12 15 18 9"/>
                </svg>
              </div>
              <div class="faq-answer">
                <p>{{ item.a }}</p>
              </div>
            </div>
          </div>
        </div>

      </div>
    </section>

    <!-- CTA -->
    <section class="faq-cta">
      <div class="container">
        <h2>{{ t('landing.faq.ctaTitle') }}</h2>
        <p>{{ t('landing.faq.ctaSub') }}</p>
        <div class="cta-buttons">
          <router-link to="/dashboard" class="btn-primary">{{ t('landing.faq.ctaBtn1') }}</router-link>
          <router-link to="/security" class="btn-ghost">{{ t('landing.faq.ctaBtn2') }}</router-link>
        </div>
      </div>
    </section>

  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import LandingNav from '../../components/landing/LandingNav.vue'

const { t } = useI18n()

const openItems = ref({})

function toggle(key) {
  openItems.value[key] = !openItems.value[key]
}

const faqGeneral = computed(() => [
  { q: t('landing.faq.g1q'), a: t('landing.faq.g1a') },
  { q: t('landing.faq.g2q'), a: t('landing.faq.g2a') },
  { q: t('landing.faq.g3q'), a: t('landing.faq.g3a') },
  { q: t('landing.faq.g4q'), a: t('landing.faq.g4a') },
  { q: t('landing.faq.g5q'), a: t('landing.faq.g5a') },
])

const faqSecurity = computed(() => [
  { q: t('landing.faq.s1q'), a: t('landing.faq.s1a') },
  { q: t('landing.faq.s2q'), a: t('landing.faq.s2a') },
  { q: t('landing.faq.s3q'), a: t('landing.faq.s3a') },
  { q: t('landing.faq.s4q'), a: t('landing.faq.s4a') },
  { q: t('landing.faq.s5q'), a: t('landing.faq.s5a') },
])

const faqFeatures = computed(() => [
  { q: t('landing.faq.f1q'), a: t('landing.faq.f1a') },
  { q: t('landing.faq.f2q'), a: t('landing.faq.f2a') },
  { q: t('landing.faq.f3q'), a: t('landing.faq.f3a') },
  { q: t('landing.faq.f4q'), a: t('landing.faq.f4a') },
  { q: t('landing.faq.f5q'), a: t('landing.faq.f5a') },
])

const faqValues = computed(() => [
  { q: t('landing.faq.v1q'), a: t('landing.faq.v1a') },
  { q: t('landing.faq.v2q'), a: t('landing.faq.v2a') },
  { q: t('landing.faq.v3q'), a: t('landing.faq.v3a') },
  { q: t('landing.faq.v4q'), a: t('landing.faq.v4a') },
])
</script>

<style scoped>
/* ── Shared layout ── */
.faq-page {
  min-height: 100vh;
  background: var(--background-color);
  color: var(--main-text-color);
}

.container {
  max-width: 860px;
  margin: 0 auto;
  padding: 0 2rem;
}

.highlight {
  color: var(--primary-color);
}

/* ── Hero ── */
.faq-hero {
  padding: 5rem 0 3.5rem;
  text-align: center;
  background: linear-gradient(180deg, var(--card-color) 0%, var(--background-color) 100%);
  border-bottom: 1px solid var(--border-color);
}

.hero-eyebrow {
  font-size: 0.8rem;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--primary-color);
  margin-bottom: 1rem;
}

.faq-hero h1 {
  font-size: clamp(2rem, 5vw, 3rem);
  font-weight: 800;
  line-height: 1.15;
  margin: 0 0 1.25rem;
}

.hero-sub {
  max-width: 600px;
  margin: 0 auto;
  color: var(--secondary-text-color);
  line-height: 1.7;
  font-size: 1.05rem;
}

/* ── FAQ Section ── */
.faq-section {
  padding: 4rem 0 5rem;
}

.faq-category {
  margin-bottom: 3.5rem;
}

.cat-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.cat-icon {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.cat-icon--blue   { background: rgba(59, 130, 246, 0.12); color: #3b82f6; }
.cat-icon--green  { background: rgba(34, 197, 94, 0.12);  color: #22c55e; }
.cat-icon--purple { background: rgba(168, 85, 247, 0.12); color: #a855f7; }
.cat-icon--orange { background: rgba(249, 115, 22, 0.12); color: #f97316; }

.cat-label {
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--secondary-text-color);
  margin-bottom: 0.2rem;
}

.cat-title {
  font-size: 1.4rem;
  font-weight: 700;
  margin: 0;
  color: var(--main-text-color);
}

/* ── Accordion items ── */
.faq-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.faq-item {
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  overflow: hidden;
  cursor: pointer;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.faq-item:hover {
  border-color: var(--primary-color);
  box-shadow: 0 2px 12px rgba(0, 80, 255, 0.07);
}

.faq-item.open {
  border-color: var(--primary-color);
}

.faq-question {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 1.1rem 1.4rem;
  font-weight: 600;
  font-size: 0.97rem;
  color: var(--main-text-color);
  user-select: none;
}

.faq-chevron {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
  color: var(--secondary-text-color);
  transition: transform 0.25s ease;
}

.faq-item.open .faq-chevron {
  transform: rotate(180deg);
  color: var(--primary-color);
}

.faq-answer {
  max-height: 0;
  overflow: hidden;
  transition: max-height 0.3s ease, padding 0.3s ease;
}

.faq-item.open .faq-answer {
  max-height: 500px;
  padding-bottom: 1.2rem;
}

.faq-answer p {
  margin: 0;
  padding: 0 1.4rem;
  color: var(--secondary-text-color);
  line-height: 1.75;
  font-size: 0.93rem;
}

/* ── CTA ── */
.faq-cta {
  background: var(--card-color);
  border-top: 1px solid var(--border-color);
  padding: 4rem 0;
  text-align: center;
}

.faq-cta h2 {
  font-size: 1.8rem;
  font-weight: 800;
  margin: 0 0 0.75rem;
}

.faq-cta p {
  color: var(--secondary-text-color);
  margin: 0 0 2rem;
  font-size: 1rem;
}

.cta-buttons {
  display: flex;
  gap: 1rem;
  justify-content: center;
  flex-wrap: wrap;
}

.btn-primary {
  padding: 0.75rem 2rem;
  background: var(--primary-color);
  color: #fff;
  text-decoration: none;
  border-radius: 10px;
  font-weight: 700;
  font-size: 0.97rem;
  transition: background 0.2s ease, transform 0.2s ease;
}

.btn-primary:hover {
  background: var(--accent-color);
  transform: translateY(-2px);
}

.btn-ghost {
  padding: 0.75rem 2rem;
  border: 1.5px solid var(--border-color);
  color: var(--main-text-color);
  text-decoration: none;
  border-radius: 10px;
  font-weight: 600;
  font-size: 0.97rem;
  transition: border-color 0.2s ease, color 0.2s ease;
}

.btn-ghost:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
}

/* ── Responsive ── */
@media (max-width: 640px) {
  .faq-hero {
    padding: 3.5rem 0 2.5rem;
  }

  .cat-title {
    font-size: 1.2rem;
  }

  .faq-question {
    font-size: 0.9rem;
    padding: 1rem 1.1rem;
  }

  .faq-answer p {
    padding: 0 1.1rem;
  }

  .cta-buttons {
    flex-direction: column;
    align-items: center;
  }
}
</style>
