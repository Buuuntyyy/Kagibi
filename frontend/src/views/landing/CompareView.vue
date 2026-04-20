<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="landing-compare">
    <LandingNav />

    <section class="compare-hero">
      <div class="container">
        <div class="hero-eyebrow">{{ t('landing.compare.eyebrow') }}</div>
        <h1>{{ t('landing.compare.title') }}</h1>
        <p class="hero-subtitle">{{ t('landing.compare.subtitle') }}</p>
      </div>
    </section>

    <section class="compare-table-section">
      <div class="container">
        <div class="table-wrapper">
          <table class="compare-table">
            <thead>
              <tr>
                <th class="feature-col">{{ t('landing.compare.tableFeature') }}</th>
                <th class="kagibi-col">
                  <div class="th-brand kagibi">
                    <img src="/Logo.png" alt="Kagibi" width="24" />
                    Kagibi
                  </div>
                </th>
                <th>
                  <div class="th-brand">Proton Drive</div>
                </th>
                <th>
                  <div class="th-brand">Nextcloud</div>
                </th>
                <th>
                  <div class="th-brand">OpenCloud</div>
                </th>
                <th>
                  <div class="th-brand">Google Drive</div>
                </th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="row in rows" :key="row.feature" :class="{ 'row-highlight': row.highlight }">
                <td class="feature-cell">
                  <span class="feature-name">{{ row.feature }}</span>
                  <span v-if="row.tooltip" class="feature-hint">{{ row.tooltip }}</span>
                </td>
                <td class="kagibi-col">
                  <span :class="cellClass(row.kagibi)">{{ cellText(row.kagibi) }}</span>
                </td>
                <td><span :class="cellClass(row.proton)">{{ cellText(row.proton) }}</span></td>
                <td><span :class="cellClass(row.nextcloud)">{{ cellText(row.nextcloud) }}</span></td>
                <td><span :class="cellClass(row.opencloud)">{{ cellText(row.opencloud) }}</span></td>
                <td><span :class="cellClass(row.google)">{{ cellText(row.google) }}</span></td>
              </tr>
            </tbody>
          </table>
        </div>

        <p class="disclaimer">{{ t('landing.compare.disclaimer') }}</p>
      </div>
    </section>

    <!-- Technical deep-dive -->
    <section class="tech-section">
      <div class="container">
        <div class="tech-header">
          <span class="tech-badge">+</span>
          <div>
            <h2>{{ t('landing.compare.techTitle') }}</h2>
            <p class="tech-subtitle">{{ t('landing.compare.techSubtitle') }}</p>
          </div>
        </div>

        <div class="tech-legend">
          <span class="tag green">{{ t('landing.compare.legendColorGreen') }}</span> {{ t('landing.compare.legendAdvantage') }}
          <span class="tag yellow">{{ t('landing.compare.legendColorYellow') }}</span> {{ t('landing.compare.legendPartial') }}
          <span class="tag red">{{ t('landing.compare.legendColorRed') }}</span> {{ t('landing.compare.legendRisk') }}
          <span class="tag gray">{{ t('landing.compare.legendColorGray') }}</span> {{ t('landing.compare.legendNeutral') }}
        </div>

        <div class="tech-grid">

          <div class="tech-card kagibi-card">
            <div class="tech-card-header">
              <img src="/Logo.png" alt="Kagibi" width="22" />
              <span>Kagibi</span>
            </div>
            <ul class="tech-list">
              <li><span class="tag green">AES-256-GCM</span> {{ t('landing.compare.techKagibi1') }}</li>
              <li><span class="tag green">RSA-OAEP / ECDH</span> {{ t('landing.compare.techKagibi2') }}</li>
              <li><span class="tag green">PBKDF2 / Argon2</span> {{ t('landing.compare.techKagibi3') }}</li>
              <li><span class="tag green">WebCrypto API</span> {{ t('landing.compare.techKagibi4') }}</li>
              <li><span class="tag green">Zero-Knowledge</span> {{ t('landing.compare.techKagibi5') }}</li>
              <li><span class="tag green">WebRTC DataChannel</span> {{ t('landing.compare.techKagibi6') }}</li>
              <li><span class="tag green">AGPLv3</span> {{ t('landing.compare.techKagibi7') }}</li>
              <li><span class="tag yellow">JWT + Redis</span> {{ t('landing.compare.techKagibi8') }}</li>
              <li><span class="tag yellow">MFA TOTP</span> {{ t('landing.compare.techKagibi9') }}</li>
            </ul>
          </div>

          <div class="tech-card">
            <div class="tech-card-header">
              <span>Proton Drive</span>
            </div>
            <ul class="tech-list">
              <li><span class="tag green">AES-256</span> {{ t('landing.compare.techProton1') }}</li>
              <li><span class="tag green">OpenPGP</span> {{ t('landing.compare.techProton2') }}</li>
              <li><span class="tag green">SRP</span> {{ t('landing.compare.techProton3') }}</li>
              <li><span class="tag green">Zero-Knowledge</span> {{ t('landing.compare.techProton4') }}</li>
              <li><span class="tag gray">Propriétaire</span> {{ t('landing.compare.techProton5') }}</li>
              <li><span class="tag gray">Suisse</span> {{ t('landing.compare.techProton6') }}</li>
              <li><span class="tag gray">Pas de P2P</span> {{ t('landing.compare.techProton7') }}</li>
            </ul>
          </div>

          <div class="tech-card">
            <div class="tech-card-header">
              <span>Nextcloud</span>
            </div>
            <ul class="tech-list">
              <li><span class="tag yellow">AES-128/256</span> {{ t('landing.compare.techNextcloud1') }}</li>
              <li><span class="tag red">Côté serveur</span> {{ t('landing.compare.techNextcloud2') }}</li>
              <li><span class="tag green">Open Source</span> {{ t('landing.compare.techNextcloud3') }}</li>
              <li><span class="tag green">Auto-hébergeable</span> {{ t('landing.compare.techNextcloud4') }}</li>
              <li><span class="tag gray">E2EE expérimental</span> {{ t('landing.compare.techNextcloud5') }}</li>
              <li><span class="tag gray">Dépend de l'hébergeur</span> {{ t('landing.compare.techNextcloud6') }}</li>
              <li><span class="tag gray">Pas de P2P</span> {{ t('landing.compare.techNextcloud7') }}</li>
            </ul>
          </div>

          <div class="tech-card">
            <div class="tech-card-header">
              <span>OpenCloud</span>
            </div>
            <ul class="tech-list">
              <li><span class="tag yellow">AES-256</span> {{ t('landing.compare.techOpencloud1') }}</li>
              <li><span class="tag red">Côté serveur</span> {{ t('landing.compare.techOpencloud2') }}</li>
              <li><span class="tag green">Open Source</span> {{ t('landing.compare.techOpencloud3') }}</li>
              <li><span class="tag green">Auto-hébergeable</span> {{ t('landing.compare.techOpencloud4') }}</li>
              <li><span class="tag gray">Pas d'E2EE natif</span> {{ t('landing.compare.techOpencloud5') }}</li>
              <li><span class="tag gray">Allemagne</span> {{ t('landing.compare.techOpencloud6') }}</li>
              <li><span class="tag gray">Pas de P2P</span> {{ t('landing.compare.techOpencloud7') }}</li>
            </ul>
          </div>

          <div class="tech-card">
            <div class="tech-card-header">
              <span>Google Drive</span>
            </div>
            <ul class="tech-list">
              <li><span class="tag yellow">AES-256</span> {{ t('landing.compare.techGoogle1') }}</li>
              <li><span class="tag red">Côté serveur</span> {{ t('landing.compare.techGoogle2') }}</li>
              <li><span class="tag red">Code fermé</span> {{ t('landing.compare.techGoogle3') }}</li>
              <li><span class="tag red">Cloud Act</span> {{ t('landing.compare.techGoogle4') }}</li>
              <li><span class="tag red">Analyse des données</span> {{ t('landing.compare.techGoogle5') }}</li>
              <li><span class="tag red">USA</span> {{ t('landing.compare.techGoogle6') }}</li>
              <li><span class="tag gray">Pas de P2P</span> {{ t('landing.compare.techGoogle7') }}</li>
            </ul>
          </div>

        </div>
      </div>
    </section>

    <section class="compare-cta">
      <div class="container">
        <h2>{{ t('landing.compare.ctaTitle') }}</h2>
        <p>{{ t('landing.compare.ctaSubtitle') }}</p>
        <div class="cta-buttons">
          <router-link to="/login" class="btn btn-primary btn-large">
            {{ t('landing.compare.ctaCreate') }}
          </router-link>
          <router-link to="/" class="btn btn-outline btn-large">
            {{ t('landing.compare.ctaMore') }}
          </router-link>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import LandingNav from '../../components/landing/LandingNav.vue'

const { t } = useI18n()

function cellClass(v) {
  if (v === true)       return 'cell yes'
  if (v === false)      return 'cell no'
  if (v === 'partial')  return 'cell partial'
  if (v === 'roadmap')  return 'cell roadmap'
  return 'cell text'
}

function cellText(v) {
  if (v === true)       return '✓'
  if (v === false)      return '✗'
  if (v === 'partial')  return '~'
  if (v === 'roadmap')  return t('landing.compare.inDev')
  return v
}

const rows = computed(() => [
  {
    feature: t('landing.compare.r1'),
    tooltip: t('landing.compare.r1tip'),
    highlight: true,
    kagibi: true, proton: true, nextcloud: false, opencloud: false, google: false,
  },
  {
    feature: t('landing.compare.r2'),
    tooltip: t('landing.compare.r2tip'),
    kagibi: true, proton: false, nextcloud: false, opencloud: false, google: false,
  },
  {
    feature: t('landing.compare.r3'),
    kagibi: true, proton: 'partial', nextcloud: true, opencloud: true, google: false,
  },
  {
    feature: t('landing.compare.r4'),
    kagibi: true, proton: true, nextcloud: 'partial', opencloud: 'partial', google: false,
  },
  {
    feature: t('landing.compare.r5'),
    tooltip: t('landing.compare.r5tip'),
    kagibi: true, proton: true, nextcloud: 'partial', opencloud: 'partial', google: false,
  },
  {
    feature: t('landing.compare.r6'),
    kagibi: true, proton: true, nextcloud: true, opencloud: true, google: 'partial',
  },
  {
    feature: t('landing.compare.r7'),
    kagibi: true, proton: true, nextcloud: true, opencloud: true, google: 'partial',
  },
  {
    feature: t('landing.compare.r8'),
    tooltip: t('landing.compare.r8tip'),
    kagibi: 'roadmap', proton: false, nextcloud: true, opencloud: true, google: false,
  },
  {
    feature: t('landing.compare.r9'),
    tooltip: t('landing.compare.r9tip'),
    kagibi: true, proton: false, nextcloud: 'partial', opencloud: 'partial', google: false,
  },
  {
    feature: t('landing.compare.r10'),
    kagibi: true, proton: true, nextcloud: true, opencloud: true, google: true,
  },
  {
    feature: t('landing.compare.r11'),
    kagibi: false, proton: true, nextcloud: true, opencloud: true, google: true,
  },
  {
    feature: t('landing.compare.r12'),
    kagibi: 'roadmap', proton: true, nextcloud: true, opencloud: true, google: true,
  },
  {
    feature: t('landing.compare.r13'),
    kagibi: true, proton: true, nextcloud: false, opencloud: false, google: false,
  },
  {
    feature: t('landing.compare.r14'),
    highlight: true,
    kagibi: true, proton: true, nextcloud: false, opencloud: false, google: false,
  },
])
</script>

<style scoped>
.landing-compare {
  background: var(--card-color);
  color: var(--main-text-color);
  overflow-y: auto;
  height: 100vh;
}

/* Hero */
.compare-hero {
  padding: 6rem 0 4rem;
  background: linear-gradient(135deg, var(--background-color) 0%, var(--card-color) 100%);
  text-align: center;
}

.hero-eyebrow {
  display: inline-block;
  font-size: 0.8rem;
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--primary-color);
  background: color-mix(in srgb, var(--primary-color) 12%, transparent);
  border: 1px solid color-mix(in srgb, var(--primary-color) 30%, transparent);
  padding: 0.35rem 1rem;
  border-radius: 999px;
  margin-bottom: 1.5rem;
}

.compare-hero h1 {
  font-size: clamp(2rem, 5vw, 3.2rem);
  font-weight: 800;
  margin-bottom: 1rem;
  letter-spacing: -0.02em;
}

.hero-subtitle {
  font-size: 1.1rem;
  color: var(--secondary-text-color);
  max-width: 600px;
  margin: 0 auto;
  line-height: 1.7;
}

/* Table section */
.compare-table-section {
  padding: 4rem 0 3rem;
  background: var(--background-color);
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 2rem;
}

.table-wrapper {
  overflow-x: auto;
  border-radius: 12px;
  border: 1px solid var(--border-color);
  box-shadow: 0 4px 20px rgba(0,0,0,0.06);
}

.compare-table {
  width: 100%;
  border-collapse: collapse;
  background: var(--card-color);
  font-size: 0.95rem;
}

.compare-table thead tr {
  background: var(--background-color);
  border-bottom: 2px solid var(--border-color);
}

.compare-table th {
  padding: 1.25rem 1rem;
  text-align: center;
  font-weight: 600;
  color: var(--secondary-text-color);
  font-size: 0.85rem;
  white-space: nowrap;
}

.feature-col {
  text-align: left !important;
  width: 28%;
  min-width: 200px;
}

.kagibi-col {
  background: color-mix(in srgb, var(--primary-color) 6%, transparent);
}

.th-brand {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--main-text-color);
}

.th-brand.kagibi {
  color: var(--primary-color);
}

.compare-table tbody tr {
  border-bottom: 1px solid var(--border-color);
  transition: background 0.15s;
}

.compare-table tbody tr:last-child {
  border-bottom: none;
}

.compare-table tbody tr:hover {
  background: var(--hover-background-color);
}

.compare-table tbody tr.row-highlight {
  background: color-mix(in srgb, var(--primary-color) 4%, transparent);
}

.compare-table tbody tr.row-highlight:hover {
  background: color-mix(in srgb, var(--primary-color) 8%, transparent);
}

.compare-table td {
  padding: 1rem;
  text-align: center;
  vertical-align: middle;
}

.feature-cell {
  text-align: left !important;
  padding-left: 1.25rem !important;
}

.feature-name {
  display: block;
  font-weight: 500;
  color: var(--main-text-color);
}

.feature-hint {
  display: block;
  font-size: 0.78rem;
  color: var(--secondary-text-color);
  margin-top: 0.2rem;
  line-height: 1.4;
}

/* Cells */
.cell {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 1rem;
  width: 28px;
  height: 28px;
  border-radius: 50%;
}

.cell.yes {
  color: #16a34a;
  background: rgba(22, 163, 74, 0.12);
}

.cell.no {
  color: #dc2626;
  background: rgba(220, 38, 38, 0.1);
}

.cell.partial {
  color: #d97706;
  background: rgba(217, 119, 6, 0.1);
  font-size: 1.2rem;
}

.cell.roadmap {
  width: auto;
  background: rgba(99, 102, 241, 0.1);
  color: var(--primary-color);
  font-size: 0.72rem;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--primary-color) 30%, transparent);
  letter-spacing: 0.02em;
}

.cell.text {
  width: auto;
  background: none;
  font-size: 0.82rem;
  font-weight: 500;
  color: var(--secondary-text-color);
  border-radius: 0;
}

.disclaimer {
  margin-top: 1.5rem;
  font-size: 0.8rem;
  color: var(--secondary-text-color);
  text-align: center;
}

/* CTA */
.compare-cta {
  padding: 6rem 0;
  background: linear-gradient(135deg, var(--background-color) 0%, var(--card-color) 100%);
  text-align: center;
}

.compare-cta h2 {
  font-size: clamp(1.6rem, 3.5vw, 2.4rem);
  font-weight: 800;
  margin-bottom: 1rem;
  color: var(--main-text-color);
  max-width: 700px;
  margin-left: auto;
  margin-right: auto;
}

.compare-cta p {
  font-size: 1.1rem;
  color: var(--secondary-text-color);
  margin-bottom: 2.5rem;
}

.cta-buttons {
  display: flex;
  gap: 1rem;
  justify-content: center;
  flex-wrap: wrap;
}

.btn {
  display: inline-flex;
  align-items: center;
  padding: 1rem 2rem;
  border-radius: 8px;
  font-weight: 600;
  font-size: 1rem;
  text-decoration: none;
  transition: all 0.3s ease;
  border: 2px solid transparent;
}

.btn-primary {
  background: var(--primary-color);
  color: var(--card-color);
}

.btn-primary:hover {
  background: var(--accent-color);
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(99, 102, 241, 0.3);
}

.btn-outline {
  background: transparent;
  border: 2px solid var(--border-color);
  color: var(--main-text-color);
}

.btn-outline:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
}

.btn-large {
  padding: 1.1rem 2.5rem;
  font-size: 1.05rem;
}

/* Technical section */
.tech-section {
  padding: 5rem 0;
  background: var(--card-color);
}

.tech-header {
  display: flex;
  align-items: flex-start;
  gap: 1.25rem;
  margin-bottom: 3rem;
}

.tech-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.5rem;
  height: 2.5rem;
  flex-shrink: 0;
  background: var(--primary-color);
  color: white;
  font-size: 1.4rem;
  font-weight: 800;
  border-radius: 10px;
  margin-top: 0.2rem;
}

.tech-header h2 {
  font-size: clamp(1.5rem, 3vw, 2rem);
  font-weight: 700;
  color: var(--main-text-color);
  margin-bottom: 0.4rem;
}

.tech-subtitle {
  color: var(--secondary-text-color);
  font-size: 1rem;
}

.tech-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 1.25rem;
}

.tech-card {
  background: var(--background-color);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 1.5rem;
}

.tech-card.kagibi-card {
  border-color: var(--primary-color);
  box-shadow: 0 0 0 1px color-mix(in srgb, var(--primary-color) 30%, transparent);
}

.tech-card-header {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  font-weight: 700;
  font-size: 1rem;
  color: var(--main-text-color);
  margin-bottom: 1.25rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid var(--border-color);
}

.kagibi-card .tech-card-header {
  color: var(--primary-color);
}

.tech-list {
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
}

.tech-list li {
  display: flex;
  align-items: baseline;
  gap: 0.5rem;
  font-size: 0.88rem;
  color: var(--secondary-text-color);
  line-height: 1.4;
}

.tag {
  display: inline-block;
  flex-shrink: 0;
  font-size: 0.7rem;
  font-weight: 700;
  padding: 1px 7px;
  border-radius: 4px;
  letter-spacing: 0.02em;
}

.tag.green  { background: rgba(22, 163, 74, 0.12);  color: #16a34a; border: 1px solid rgba(22, 163, 74, 0.25); }
.tag.yellow { background: rgba(217, 119, 6, 0.12);  color: #d97706; border: 1px solid rgba(217, 119, 6, 0.25); }
.tag.red    { background: rgba(220, 38, 38, 0.10);  color: #dc2626; border: 1px solid rgba(220, 38, 38, 0.25); }
.tag.gray   { background: var(--hover-background-color); color: var(--secondary-text-color); border: 1px solid var(--border-color); }

.tech-legend {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
  margin-bottom: 2rem;
  padding: 0.75rem 1rem;
  background: var(--background-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  font-size: 0.85rem;
  color: var(--secondary-text-color);
}

@media (max-width: 768px) {
  .compare-hero {
    padding: 4rem 0 2rem;
  }

  .compare-hero h1 {
    font-size: clamp(1.6rem, 5vw, 2.5rem);
  }

  .table-wrapper {
    border-radius: 8px;
  }

  .compare-table th,
  .compare-table td {
    padding: 10px 8px;
    font-size: 0.82rem;
  }

  .feature-cell {
    min-width: 120px;
  }

  .tech-grid {
    grid-template-columns: 1fr;
  }

  .cta-buttons {
    flex-direction: column;
    align-items: center;
  }

  .btn {
    width: 100%;
    justify-content: center;
  }
}

@media (max-width: 480px) {
  .container {
    padding: 0 1rem;
  }

  .compare-table th,
  .compare-table td {
    padding: 8px 6px;
    font-size: 0.75rem;
  }

  .th-brand {
    font-size: 0.75rem;
  }

  .feature-hint {
    display: none; /* hide tooltips on very small screens to save space */
  }
}
</style>
