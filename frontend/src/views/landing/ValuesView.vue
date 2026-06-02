<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="values-page">
    <LandingNav />

    <!-- ── Hero ── -->
    <section class="sec-hero">
      <div class="container">
        <div class="hero-eyebrow">{{ t('landing.values.eyebrow') }}</div>
        <h1>
          {{ t('landing.values.title1') }}<br>
          <span class="highlight">{{ t('landing.values.titleHighlight') }}</span>
        </h1>
        <p class="hero-sub">{{ t('landing.values.heroSub') }}</p>
      </div>
    </section>

    <!-- ── Valeurs fondamentales ── -->
    <section class="section">
      <div class="container">
        <div class="section-label">{{ t('landing.values.sec1Label') }}</div>
        <h2>{{ t('landing.values.sec1Title') }}</h2>
        <p class="section-sub">{{ t('landing.values.sec1Sub') }}</p>

        <div class="values-grid">

          <div class="value-card vc-primary">
            <div class="value-icon-wrap vi-primary">
              <svg viewBox="0 0 24 24" width="22" height="22" fill="currentColor">
                <path d="M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zm-6 9c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2zm3.1-9H8.9V6c0-1.71 1.39-3.1 3.1-3.1 1.71 0 3.1 1.39 3.1 3.1v2z"/>
              </svg>
            </div>
            <h3>{{ t('landing.values.v1.title') }}</h3>
            <p>{{ t('landing.values.v1.desc') }}</p>
          </div>

          <div class="value-card vc-blue">
            <div class="value-icon-wrap vi-blue">
              <svg viewBox="0 0 24 24" width="22" height="22" fill="currentColor">
                <path d="M9.4 16.6L4.8 12l4.6-4.6L8 6l-6 6 6 6 1.4-1.4zm5.2 0l4.6-4.6-4.6-4.6L16 6l6 6-6 6-1.4-1.4z"/>
              </svg>
            </div>
            <h3>{{ t('landing.values.v2.title') }}</h3>
            <p>{{ t('landing.values.v2.desc') }}</p>
          </div>

          <div class="value-card vc-purple">
            <div class="value-icon-wrap vi-purple">
              <svg viewBox="0 0 24 24" width="22" height="22" fill="currentColor">
                <path d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4z"/>
              </svg>
            </div>
            <h3>{{ t('landing.values.v3.title') }}</h3>
            <p>{{ t('landing.values.v3.desc') }}</p>
          </div>

          <div class="value-card vc-green">
            <div class="value-icon-wrap vi-green">
              <svg viewBox="0 0 24 24" width="22" height="22" fill="currentColor">
                <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
              </svg>
            </div>
            <h3>{{ t('landing.values.v4.title') }}</h3>
            <p>{{ t('landing.values.v4.desc') }}</p>
          </div>

        </div>
      </div>
    </section>

    <!-- ── Engagements concrets ── -->
    <section class="section dark-section">
      <div class="container">
        <div class="section-label">{{ t('landing.values.sec2Label') }}</div>
        <h2>{{ t('landing.values.sec2Title') }}</h2>
        <p class="section-sub">{{ t('landing.values.sec2Sub') }}</p>

        <div class="commit-grid">

          <div class="commit-card">
            <div class="commit-top">
              <div class="commit-icon-wrap ci-primary">
                <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                  <path d="M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zm-6 9c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2zm3.1-9H8.9V6c0-1.71 1.39-3.1 3.1-3.1 1.71 0 3.1 1.39 3.1 3.1v2z"/>
                </svg>
              </div>
              <span class="commit-badge badge-primary">{{ t('landing.values.c1.badge') }}</span>
            </div>
            <h4>{{ t('landing.values.c1.title') }}</h4>
            <p>{{ t('landing.values.c1.desc') }}</p>
          </div>

          <div class="commit-card">
            <div class="commit-top">
              <div class="commit-icon-wrap ci-blue">
                <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                  <path d="M12 2C8.13 2 5 5.13 5 9c0 5.25 7 13 7 13s7-7.75 7-13c0-3.87-3.13-7-7-7zm0 9.5c-1.38 0-2.5-1.12-2.5-2.5s1.12-2.5 2.5-2.5 2.5 1.12 2.5 2.5-1.12 2.5-2.5 2.5z"/>
                </svg>
              </div>
              <span class="commit-badge badge-blue">{{ t('landing.values.c2.badge') }}</span>
            </div>
            <h4>{{ t('landing.values.c2.title') }}</h4>
            <p>{{ t('landing.values.c2.desc') }}</p>
          </div>

          <div class="commit-card">
            <div class="commit-top">
              <div class="commit-icon-wrap ci-teal">
                <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                  <path d="M9.4 16.6L4.8 12l4.6-4.6L8 6l-6 6 6 6 1.4-1.4zm5.2 0l4.6-4.6-4.6-4.6L16 6l6 6-6 6-1.4-1.4z"/>
                </svg>
              </div>
              <span class="commit-badge badge-teal">{{ t('landing.values.c3.badge') }}</span>
            </div>
            <h4>{{ t('landing.values.c3.title') }}</h4>
            <p>{{ t('landing.values.c3.desc') }}</p>
          </div>

          <div class="commit-card">
            <div class="commit-top">
              <div class="commit-icon-wrap ci-orange">
                <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                  <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
                </svg>
              </div>
              <span class="commit-badge badge-orange">{{ t('landing.values.c4.badge') }}</span>
            </div>
            <h4>{{ t('landing.values.c4.title') }}</h4>
            <p>{{ t('landing.values.c4.desc') }}</p>
          </div>

          <div class="commit-card">
            <div class="commit-top">
              <div class="commit-icon-wrap ci-purple">
                <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                  <path d="M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm-7 3c1.93 0 3.5 1.57 3.5 3.5S13.93 13 12 13s-3.5-1.57-3.5-3.5S10.07 6 12 6zm7 13H5v-.23c0-.62.28-1.2.76-1.58C7.47 15.82 9.64 15 12 15s4.53.82 6.24 2.19c.48.38.76.97.76 1.58V19z"/>
                </svg>
              </div>
              <span class="commit-badge badge-purple">{{ t('landing.values.c5.badge') }}</span>
            </div>
            <h4>{{ t('landing.values.c5.title') }}</h4>
            <p>{{ t('landing.values.c5.desc') }}</p>
          </div>

          <div class="commit-card">
            <div class="commit-top">
              <div class="commit-icon-wrap ci-green">
                <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                  <path d="M17 12h-5v5h5v-5zM16 1v2H8V1H6v2H5c-1.11 0-1.99.9-1.99 2L3 19c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2h-1V1h-2zm3 18H5V8h14v11z"/>
                </svg>
              </div>
              <span class="commit-badge badge-green">{{ t('landing.values.c6.badge') }}</span>
            </div>
            <h4>{{ t('landing.values.c6.title') }}</h4>
            <p>{{ t('landing.values.c6.desc') }}</p>
          </div>

        </div>
      </div>
    </section>

    <!-- ── Souveraineté : 3 prismes ── -->
    <section class="section sec-sov">
      <div class="container">
        <div class="section-label">{{ t('landing.values.sovLabel') }}</div>
        <h2>{{ t('landing.values.sovTitle') }}</h2>
        <p class="section-sub">{{ t('landing.values.sovSub') }}</p>

        <div class="sov-grid">

          <!-- Prisme 1 : Géographique -->
          <div class="sov-card sov-card-blue">
            <div class="sov-num">01</div>
            <div class="sov-icon-wrap si-blue">
              <svg viewBox="0 0 24 24" width="22" height="22" fill="currentColor">
                <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 17.93c-3.95-.49-7-3.85-7-7.93 0-.62.08-1.21.21-1.79L9 15v1c0 1.1.9 2 2 2v1.93zm6.9-2.54c-.26-.81-1-1.39-1.9-1.39h-1v-3c0-.55-.45-1-1-1H8v-2h2c.55 0 1-.45 1-1V7h2c1.1 0 2-.9 2-2v-.41c2.93 1.19 5 4.06 5 7.41 0 2.08-.8 3.97-2.1 5.39z"/>
              </svg>
            </div>
            <span class="sov-tag st-blue">{{ t('landing.values.sov1.tag') }}</span>
            <h3>{{ t('landing.values.sov1.title') }}</h3>
            <p>{{ t('landing.values.sov1.desc') }}</p>
            <ul class="sov-bullets">
              <li>{{ t('landing.values.sov1.b1') }}</li>
              <li>{{ t('landing.values.sov1.b2') }}</li>
              <li>{{ t('landing.values.sov1.b3') }}</li>
            </ul>
          </div>

          <!-- Prisme 2 : Nationale & Juridique -->
          <div class="sov-card sov-card-teal">
            <div class="sov-num">02</div>
            <div class="sov-icon-wrap si-teal">
              <svg viewBox="0 0 24 24" width="22" height="22" fill="currentColor">
                <path d="M12 3L1 9l11 6 9-4.91V17h2V9L12 3zM5 13.18v4L12 21l7-3.82v-4L12 17l-7-3.82z"/>
              </svg>
            </div>
            <span class="sov-tag st-teal">{{ t('landing.values.sov2.tag') }}</span>
            <h3>{{ t('landing.values.sov2.title') }}</h3>
            <p>{{ t('landing.values.sov2.desc') }}</p>
            <ul class="sov-bullets">
              <li>{{ t('landing.values.sov2.b1') }}</li>
              <li>{{ t('landing.values.sov2.b2') }}</li>
              <li>{{ t('landing.values.sov2.b3') }}</li>
            </ul>
          </div>

          <!-- Prisme 3 : Personnelle -->
          <div class="sov-card sov-card-primary">
            <div class="sov-num">03</div>
            <div class="sov-icon-wrap si-primary">
              <svg viewBox="0 0 24 24" width="22" height="22" fill="currentColor">
                <path d="M12.65 10C11.83 7.67 9.61 6 7 6c-3.31 0-6 2.69-6 6s2.69 6 6 6c2.61 0 4.83-1.67 5.65-4H17v4h4v-4h2v-4H12.65zM7 14c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2z"/>
              </svg>
            </div>
            <span class="sov-tag st-primary">{{ t('landing.values.sov3.tag') }}</span>
            <h3>{{ t('landing.values.sov3.title') }}</h3>
            <p>{{ t('landing.values.sov3.desc') }}</p>
            <ul class="sov-bullets">
              <li>{{ t('landing.values.sov3.b1') }}</li>
              <li>{{ t('landing.values.sov3.b2') }}</li>
              <li>{{ t('landing.values.sov3.b3') }}</li>
            </ul>
          </div>

        </div>
      </div>
    </section>

    <!-- ── Manifeste ── -->
    <section class="section sec-manifesto">
      <div class="container">
        <div class="section-label">{{ t('landing.values.manifestoLabel') }}</div>
        <h2>{{ t('landing.values.manifestoTitle') }}</h2>
        <div class="manifesto-body">
          <blockquote class="manifesto-quote">
            <span class="quote-mark">"</span>
            {{ t('landing.values.manifestoQuote') }}
            <span class="quote-mark">"</span>
          </blockquote>
          <p>{{ t('landing.values.manifestoP1') }}</p>
          <p>{{ t('landing.values.manifestoP2') }}</p>
          <p>{{ t('landing.values.manifestoP3') }}</p>
          <div class="manifesto-sig">
            <div class="sig-line"></div>
            <span>{{ t('landing.values.manifestoSig') }}</span>
          </div>
        </div>
      </div>
    </section>

    <!-- ── CTA ── -->
    <section class="sec-cta">
      <div class="container">
        <h2>{{ t('landing.values.ctaTitle') }}</h2>
        <p>{{ t('landing.values.ctaSub') }}</p>
        <div class="cta-row">
          <router-link to="/login" class="btn btn-primary">{{ t('landing.values.ctaBtn1') }}</router-link>
          <router-link to="/security" class="btn btn-outline">{{ t('landing.values.ctaBtn2') }}</router-link>
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
/* ── Layout ── */
.values-page {
  height: 100vh;
  overflow-y: auto;
  background: var(--background-color);
}

.container {
  max-width: 1100px;
  margin: 0 auto;
  padding: 0 2rem;
}

.section {
  padding: 5rem 0;
}

.dark-section {
  background: var(--card-color);
}

.section-label {
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--primary-color);
  margin-bottom: 0.75rem;
}

.section-sub {
  color: var(--secondary-text-color);
  font-size: 1.05rem;
  max-width: 680px;
  line-height: 1.7;
  margin-bottom: 2.5rem;
}

h2 {
  font-size: clamp(1.6rem, 3.5vw, 2.4rem);
  font-weight: 700;
  line-height: 1.2;
  margin-bottom: 1rem;
  color: var(--main-text-color);
}

/* ── Hero ── */
.sec-hero {
  padding: 7rem 0 5rem;
  text-align: center;
}

.hero-eyebrow {
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--primary-color);
  margin-bottom: 1rem;
}

.sec-hero h1 {
  font-size: clamp(2rem, 5vw, 3.5rem);
  line-height: 1.15;
  margin-bottom: 1.5rem;
  color: var(--main-text-color);
}

.highlight {
  background: linear-gradient(135deg, var(--primary-color), var(--secondary-color));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.hero-sub {
  color: var(--secondary-text-color);
  font-size: 1.1rem;
  max-width: 640px;
  margin: 0 auto;
  line-height: 1.75;
}

/* ── Values grid ── */
.values-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1.5rem;
}

@media (max-width: 768px) {
  .values-grid { grid-template-columns: 1fr; }
}

.value-card {
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 16px;
  padding: 2rem;
  border-left: 3px solid transparent;
  transition: transform 0.2s, box-shadow 0.2s;
}

.value-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.1);
}

.vc-primary { border-left-color: var(--primary-color); }
.vc-blue    { border-left-color: #3b82f6; }
.vc-purple  { border-left-color: #8b5cf6; }
.vc-green   { border-left-color: #22c55e; }

.value-icon-wrap {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1.25rem;
}

.vi-primary { background: color-mix(in srgb, var(--primary-color) 14%, transparent); color: var(--primary-color); }
.vi-blue    { background: rgba(59,130,246,0.12); color: #60a5fa; }
.vi-purple  { background: rgba(139,92,246,0.12); color: #a78bfa; }
.vi-green   { background: rgba(34,197,94,0.12); color: #4ade80; }

.value-card h3 {
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--main-text-color);
  margin: 0 0 0.6rem;
}

.value-card p {
  font-size: 0.9rem;
  color: var(--secondary-text-color);
  line-height: 1.7;
  margin: 0;
}

/* ── Commitments grid ── */
.commit-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1.25rem;
}

@media (max-width: 960px) {
  .commit-grid { grid-template-columns: repeat(2, 1fr); }
}

@media (max-width: 560px) {
  .commit-grid { grid-template-columns: 1fr; }
}

.commit-card {
  background: var(--background-color);
  border: 1px solid var(--border-color);
  border-radius: 14px;
  padding: 1.5rem;
  transition: border-color 0.2s;
}

.commit-card:hover { border-color: var(--primary-color); }

.commit-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1rem;
}

.commit-icon-wrap {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.ci-primary { background: color-mix(in srgb, var(--primary-color) 14%, transparent); color: var(--primary-color); }
.ci-blue    { background: rgba(59,130,246,0.12); color: #60a5fa; }
.ci-teal    { background: rgba(20,184,166,0.12); color: #2dd4bf; }
.ci-orange  { background: rgba(249,115,22,0.12); color: #fb923c; }
.ci-purple  { background: rgba(139,92,246,0.12); color: #a78bfa; }
.ci-green   { background: rgba(34,197,94,0.12); color: #4ade80; }

.commit-badge {
  font-size: 0.68rem;
  font-weight: 700;
  padding: 0.25rem 0.65rem;
  border-radius: 99px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.badge-primary { background: color-mix(in srgb, var(--primary-color) 12%, transparent); color: var(--primary-color); }
.badge-blue    { background: rgba(59,130,246,0.12); color: #60a5fa; }
.badge-teal    { background: rgba(20,184,166,0.12); color: #2dd4bf; }
.badge-orange  { background: rgba(249,115,22,0.12); color: #fb923c; }
.badge-purple  { background: rgba(139,92,246,0.12); color: #a78bfa; }
.badge-green   { background: rgba(34,197,94,0.12); color: #4ade80; }

.commit-card h4 {
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--main-text-color);
  margin: 0 0 0.5rem;
}

.commit-card p {
  font-size: 0.85rem;
  color: var(--secondary-text-color);
  line-height: 1.65;
  margin: 0;
}

/* ── Manifesto ── */
.sec-manifesto {
  background: var(--background-color);
}

.manifesto-body {
  max-width: 720px;
  margin-top: 2rem;
}

.manifesto-quote {
  font-size: clamp(1.1rem, 2.5vw, 1.45rem);
  font-style: italic;
  font-weight: 500;
  color: var(--main-text-color);
  line-height: 1.6;
  margin: 0 0 2rem;
  padding: 0;
  border: none;
  position: relative;
}

.quote-mark {
  font-size: 2rem;
  font-style: normal;
  color: var(--primary-color);
  line-height: 0;
  vertical-align: -0.4em;
  font-family: Georgia, serif;
  font-weight: 400;
}

.manifesto-body p {
  font-size: 1rem;
  color: var(--secondary-text-color);
  line-height: 1.8;
  margin: 0 0 1.25rem;
}

.manifesto-body p:last-of-type {
  margin-bottom: 2rem;
}

.manifesto-sig {
  display: flex;
  align-items: center;
  gap: 1rem;
  color: var(--secondary-text-color);
  font-size: 0.875rem;
  font-weight: 500;
}

.sig-line {
  width: 40px;
  height: 2px;
  background: var(--primary-color);
  flex-shrink: 0;
}

/* ── CTA ── */
.sec-cta {
  padding: 5rem 0;
  background: var(--card-color);
  text-align: center;
}

.sec-cta h2 {
  font-size: clamp(1.6rem, 3.5vw, 2.4rem);
  margin-bottom: 1rem;
}

.sec-cta p {
  color: var(--secondary-text-color);
  font-size: 1.05rem;
  max-width: 480px;
  margin: 0 auto 2.5rem;
  line-height: 1.7;
}

.cta-row {
  display: flex;
  gap: 1rem;
  justify-content: center;
  flex-wrap: wrap;
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1.75rem;
  border-radius: 10px;
  font-weight: 600;
  font-size: 0.95rem;
  text-decoration: none;
  transition: all 0.2s;
}

.btn-primary {
  background: var(--primary-color);
  color: var(--card-color);
}

.btn-primary:hover {
  opacity: 0.9;
  transform: translateY(-1px);
}

.btn-outline {
  background: transparent;
  color: var(--main-text-color);
  border: 1px solid var(--border-color);
}

.btn-outline:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
}

/* ── Sovereignty section ── */
.sec-sov {
  background: var(--card-color);
}

.sov-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  border: 1px solid var(--border-color);
  border-radius: 16px;
  overflow: hidden;
}

.sov-card {
  padding: 2rem;
  background: var(--background-color);
  border-right: 1px solid var(--border-color);
  border-top: 3px solid transparent;
  position: relative;
  transition: background 0.2s;
}

.sov-card:last-child { border-right: none; }
.sov-card:hover { background: var(--hover-background-color); }

.sov-card-blue    { border-top-color: #3b82f6; }
.sov-card-teal    { border-top-color: #2dd4bf; }
.sov-card-primary { border-top-color: var(--primary-color); }

.sov-num {
  font-size: 2.5rem;
  font-weight: 800;
  font-family: monospace;
  color: var(--border-color);
  line-height: 1;
  margin-bottom: 1.25rem;
  letter-spacing: -0.02em;
}

.sov-icon-wrap {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 0.75rem;
}

.si-blue    { background: rgba(59,130,246,0.12); color: #60a5fa; }
.si-teal    { background: rgba(20,184,166,0.12); color: #2dd4bf; }
.si-primary { background: color-mix(in srgb, var(--primary-color) 14%, transparent); color: var(--primary-color); }

.sov-tag {
  display: inline-block;
  font-size: 0.62rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  padding: 0.2rem 0.6rem;
  border-radius: 4px;
  margin-bottom: 0.9rem;
}

.st-blue    { background: rgba(59,130,246,0.12); color: #60a5fa; }
.st-teal    { background: rgba(20,184,166,0.12); color: #2dd4bf; }
.st-primary { background: color-mix(in srgb, var(--primary-color) 12%, transparent); color: var(--primary-color); }

.sov-card h3 {
  font-size: 1.05rem;
  font-weight: 700;
  color: var(--main-text-color);
  margin: 0 0 0.65rem;
  line-height: 1.3;
}

.sov-card > p {
  font-size: 0.875rem;
  color: var(--secondary-text-color);
  line-height: 1.7;
  margin: 0 0 1.25rem;
}

.sov-bullets {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
  border-top: 1px solid var(--border-color);
  padding-top: 1rem;
}

.sov-bullets li {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  font-size: 0.81rem;
  color: var(--secondary-text-color);
  line-height: 1.45;
}

.sov-bullets li::before {
  content: '→';
  color: var(--primary-color);
  font-weight: 700;
  flex-shrink: 0;
  font-size: 0.75rem;
  margin-top: 0.1em;
}

@media (max-width: 860px) {
  .sov-grid {
    grid-template-columns: 1fr;
  }
  .sov-card {
    border-right: none;
    border-bottom: 1px solid var(--border-color);
    border-top: none;
    border-left: 3px solid transparent;
  }
  .sov-card:last-child { border-bottom: none; }
  .sov-card-blue    { border-left-color: #3b82f6; }
  .sov-card-teal    { border-left-color: #2dd4bf; }
  .sov-card-primary { border-left-color: var(--primary-color); }
}

@media (max-width: 768px) {
  .container { padding: 0 1rem; }
  .section { padding: 3.5rem 0; }
  .sec-hero { padding: 5rem 0 3.5rem; }
  .sec-cta { padding: 3.5rem 0; }
}
</style>
