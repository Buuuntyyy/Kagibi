<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="transfer-page">
    <LandingNav />
    <section class="transfer-hero">
      <div class="container">
        <h1>{{ t('landing.transfer.heroTitle') }}</h1>
        <p>{{ t('landing.transfer.heroSubtitle') }}</p>
      </div>
    </section>

    <section class="transfer-section">
      <div class="container">
        <div class="transfer-content">
          <!-- Info Card -->
          <div class="info-card">
            <div class="info-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="22" y1="2" x2="11" y2="13"></line>
                <polygon points="22 2 15 22 11 13 2 9 22 2"></polygon>
              </svg>
            </div>
            <h2>{{ t('landing.transfer.howTitle') }}</h2>
            <div class="steps">
              <div class="step">
                <div class="step-number">1</div>
                <div class="step-content">
                  <h4>{{ t('landing.transfer.step1Title') }}</h4>
                  <p>{{ t('landing.transfer.step1Desc') }}</p>
                </div>
              </div>
              <div class="step">
                <div class="step-number">2</div>
                <div class="step-content">
                  <h4>{{ t('landing.transfer.step2Title') }}</h4>
                  <p>{{ t('landing.transfer.step2Desc') }}</p>
                </div>
              </div>
              <div class="step">
                <div class="step-number">3</div>
                <div class="step-content">
                  <h4>{{ t('landing.transfer.step3Title') }}</h4>
                  <p>{{ t('landing.transfer.step3Desc') }}</p>
                </div>
              </div>
            </div>
          </div>

          <!-- Features Grid -->
          <div class="transfer-features">
            <div class="feature-box">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"></path>
              </svg>
              <h4>{{ t('landing.transfer.feat1Title') }}</h4>
              <p>{{ t('landing.transfer.feat1Desc') }}</p>
            </div>

            <div class="feature-box">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
                <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
              </svg>
              <h4>{{ t('landing.transfer.feat2Title') }}</h4>
              <p>{{ t('landing.transfer.feat2Desc') }}</p>
            </div>

            <div class="feature-box">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"></polyline>
              </svg>
              <h4>{{ t('landing.transfer.feat3Title') }}</h4>
              <p>{{ t('landing.transfer.feat3Desc') }}</p>
            </div>

            <div class="feature-box">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"></circle>
                <polyline points="12 6 12 12 16 14"></polyline>
              </svg>
              <h4>{{ t('landing.transfer.feat4Title') }}</h4>
              <p>{{ t('landing.transfer.feat4Desc') }}</p>
            </div>
          </div>

          <!-- CTA Card -->
          <div class="cta-card">
            <h3>{{ t('landing.transfer.ctaTitle') }}</h3>
            <p>{{ t('landing.transfer.ctaSubtitle') }}</p>
            <router-link to="/login" class="btn btn-primary btn-large">
              {{ t('landing.transfer.ctaBtn') }}
            </router-link>
            <div class="cta-note">
              {{ t('landing.transfer.ctaNote') }}
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Technical Details -->
    <section class="tech-section">
      <div class="container">
        <h2>{{ t('landing.transfer.techTitle') }}</h2>
        <div class="tech-grid">
          <div class="tech-card">
            <h4>{{ t('landing.transfer.tech1Title') }}</h4>
            <p>{{ t('landing.transfer.tech1Desc') }}</p>
          </div>
          <div class="tech-card">
            <h4>{{ t('landing.transfer.tech2Title') }}</h4>
            <p>{{ t('landing.transfer.tech2Desc') }}</p>
          </div>
          <div class="tech-card">
            <h4>{{ t('landing.transfer.tech3Title') }}</h4>
            <p>{{ t('landing.transfer.tech3Desc') }}</p>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { useI18n } from 'vue-i18n'
import LandingNav from '../../components/landing/LandingNav.vue'

const { t } = useI18n()
</script>

<style scoped>
.transfer-page {
  background: var(--card-color);
  color: var(--main-text-color);
  min-height: 100vh;
  overflow-y: auto;
  height: 100vh;
}

.transfer-hero {
  padding: 6rem 2rem 4rem;
  text-align: center;
  background: linear-gradient(135deg, var(--background-color) 0%, var(--card-color) 100%);
}

.transfer-hero h1 {
  font-size: clamp(2.5rem, 5vw, 4rem);
  margin-bottom: 1rem;
  font-weight: 700;
  color: var(--main-text-color);
}

.transfer-hero p {
  font-size: 1.25rem;
  color: var(--secondary-text-color);
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 2rem;
}

.transfer-section {
  padding: 4rem 0;
}

.transfer-content {
  display: grid;
  gap: 3rem;
}

/* Info Card */
.info-card {
  background: linear-gradient(135deg, var(--card-color) 0%, var(--hover-background-color) 100%);
  border: 2px solid var(--primary-color);
  border-radius: 8px;
  padding: 3rem;
  text-align: center;
}

.info-icon {
  width: 80px;
  height: 80px;
  background: var(--hover-background-color);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 2rem;
}

.info-icon svg {
  width: 40px;
  height: 40px;
  color: var(--primary-color);
}

.info-card h2 {
  font-size: 2rem;
  margin-bottom: 3rem;
  color: var(--main-text-color);
}

.steps {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 2rem;
  text-align: left;
}

.step {
  display: flex;
  gap: 1rem;
  align-items: flex-start;
}

.step-number {
  width: 40px;
  height: 40px;
  background: var(--primary-color);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 1.25rem;
  flex-shrink: 0;
  color: var(--card-color);
}

.step-content h4 {
  font-size: 1.125rem;
  margin-bottom: 0.5rem;
  color: var(--main-text-color);
}

.step-content p {
  color: var(--secondary-text-color);
  font-size: 0.95rem;
}

/* Features Grid */
.transfer-features {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 2rem;
}

.feature-box {
  background: var(--card-color);
  border: 2px solid var(--border-color);
  border-radius: 8px;
  padding: 2rem;
  transition: all 0.3s ease;
}

.feature-box:hover {
  border-color: var(--primary-color);
  transform: translateY(-4px);
  box-shadow: 0 12px 30px rgba(0, 0, 0, 0.08);
}

.feature-box svg {
  width: 40px;
  height: 40px;
  color: var(--primary-color);
  margin-bottom: 1rem;
}

.feature-box h4 {
  font-size: 1.25rem;
  margin-bottom: 0.75rem;
  color: var(--main-text-color);
}

.feature-box p {
  color: var(--secondary-text-color);
  line-height: 1.6;
}

/* CTA Card */
.cta-card {
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--accent-color) 100%);
  border-radius: 8px;
  padding: 4rem 3rem;
  text-align: center;
}

.cta-card h3 {
  font-size: 2.5rem;
  margin-bottom: 1rem;
  color: var(--card-color);
}

.cta-card p {
  font-size: 1.125rem;
  margin-bottom: 2rem;
  color: var(--card-color);
}

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 1rem 2rem;
  border-radius: 8px;
  font-weight: 600;
  text-decoration: none;
  transition: all 0.3s ease;
}

.btn-primary {
  background: var(--card-color);
  color: var(--primary-color);
  border: 2px solid var(--card-color);
}

.btn-primary:hover {
  background: transparent;
  color: var(--card-color);
  border-color: var(--card-color);
  transform: translateY(-2px);
}

.btn-large {
  padding: 1.25rem 3rem;
  font-size: 1.125rem;
}

.cta-note {
  margin-top: 1.5rem;
  font-size: 0.95rem;
  color: var(--card-color);
}

/* Technical Section */
.tech-section {
  padding: 6rem 0;
  background: var(--background-color);
}

.tech-section h2 {
  text-align: center;
  font-size: 2.5rem;
  margin-bottom: 3rem;
  color: var(--main-text-color);
}

.tech-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 2rem;
}

.tech-card {
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 2.5rem;
}

.tech-card h4 {
  font-size: 1.5rem;
  color: var(--primary-color);
  margin-bottom: 1rem;
}

.tech-card p {
  color: var(--secondary-text-color);
  line-height: 1.7;
}

@media (max-width: 768px) {
  .steps,
  .transfer-features,
  .tech-grid {
    grid-template-columns: 1fr;
  }

  .cta-card {
    padding: 3rem 2rem;
  }

  .transfer-hero {
    padding: 3rem 1rem 2rem;
  }

  .transfer-hero h1 {
    font-size: clamp(1.6rem, 5vw, 3rem);
  }

  .info-card {
    padding: 1.5rem;
  }
}

@media (max-width: 480px) {
  .transfer-hero {
    padding: 2rem 1rem;
  }

  .cta-card {
    padding: 2rem 1rem;
  }

  .container {
    padding: 0 1rem;
  }
}
</style>
