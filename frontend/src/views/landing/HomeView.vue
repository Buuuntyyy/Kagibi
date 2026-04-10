<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="landing-home">
    <LandingNav />

    <!-- Hero Section -->
    <section class="hero">
      <div class="container">
        <div class="hero-content">
          <div class="hero-eyebrow">Cloud souverain · Open Source · France</div>
          <h1 class="hero-title">
            Le seul cloud qui<br>
            <span class="highlight">ne peut pas lire vos données</span>
          </h1>
          <p class="hero-subtitle">
            Kagibi chiffre vos fichiers sur votre appareil avant tout envoi.
            Ni nous, ni personne ne peut y accéder — mathématiquement.
            Hébergé en France, conforme RGPD, immunisé au Cloud Act.
          </p>
          <div class="hero-cta">
            <router-link to="/dashboard" class="btn btn-primary">
              Accéder au Drive
            </router-link>
            <router-link to="/transfer" class="btn btn-secondary">
              Transfert P2P instantané
            </router-link>
          </div>
          <div class="trust-badges">
            <div class="badge">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"></path>
              </svg>
              Hébergé en France
            </div>
            <div class="badge">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="16 18 22 12 16 6"></polyline>
                <polyline points="8 6 2 12 8 18"></polyline>
              </svg>
              Open Source AGPLv3
            </div>
            <div class="badge">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
                <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
              </svg>
              Zero-Knowledge
            </div>
            <div class="badge">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path>
              </svg>
              Immunisé Cloud Act
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- How it works -->
    <section class="how-it-works">
      <div class="container">
        <h2 class="section-title">Comment ça fonctionne ?</h2>
        <p class="section-subtitle">Trois étapes, zéro compromis sur la vie privée.</p>
        <div class="steps-grid">
          <div class="step">
            <div class="step-number">1</div>
            <h3>Chiffrement local</h3>
            <p>Vos fichiers sont chiffrés avec AES-256-GCM directement dans votre navigateur. La clé ne quitte jamais votre appareil.</p>
          </div>
          <div class="step-arrow">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="5" y1="12" x2="19" y2="12"></line>
              <polyline points="12 5 19 12 12 19"></polyline>
            </svg>
          </div>
          <div class="step">
            <div class="step-number">2</div>
            <h3>Stockage sécurisé</h3>
            <p>Seul le flux chiffré arrive sur nos serveurs en France. Même avec un accès complet à l'infrastructure, personne ne peut déchiffrer vos fichiers.</p>
          </div>
          <div class="step-arrow">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="5" y1="12" x2="19" y2="12"></line>
              <polyline points="12 5 19 12 12 19"></polyline>
            </svg>
          </div>
          <div class="step">
            <div class="step-number">3</div>
            <h3>Accès exclusif</h3>
            <p>Seul vous pouvez déchiffrer et lire vos données. Partagez en toute sécurité via liens chiffrés ou transfert P2P direct.</p>
          </div>
        </div>
      </div>
    </section>

    <!-- Features Section -->
    <section class="features">
      <div class="container">
        <h2 class="section-title">Tout ce dont vous avez besoin</h2>
        <p class="section-subtitle">Une plateforme complète, construite sur des fondations cryptographiques solides.</p>
        <div class="features-grid">
          <div class="feature-card">
            <div class="feature-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
                <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
              </svg>
            </div>
            <h3>Zero-Knowledge garanti</h3>
            <p>
              Chiffrement AES-256-GCM côté client. Vos clés ne quittent jamais votre appareil.
              Nos serveurs ne voient que du bruit chiffré — impossible autrement.
            </p>
          </div>

          <div class="feature-card">
            <div class="feature-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="22" y1="2" x2="11" y2="13"></line>
                <polygon points="22 2 15 22 11 13 2 9 22 2"></polygon>
              </svg>
            </div>
            <h3>Transfert P2P jusqu'à 10 Go</h3>
            <p>
              Envoyez de gros fichiers en direct, de navigateur à navigateur.
              Aucun stockage intermédiaire, vitesse maximale, aucune trace sur nos serveurs.
            </p>
          </div>

          <div class="feature-card">
            <div class="feature-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"></path>
              </svg>
            </div>
            <h3>Souveraineté numérique</h3>
            <p>
              Infrastructure 100% OVH France (Gravelines). Hors juridiction américaine.
              Conformité RGPD stricte. Vos données ne quittent jamais l'Europe.
            </p>
          </div>

          <div class="feature-card">
            <div class="feature-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="16 18 22 12 16 6"></polyline>
                <polyline points="8 6 2 12 8 18"></polyline>
              </svg>
            </div>
            <h3>100% vérifiable</h3>
            <p>
              Code source intégralement publié sous licence AGPLv3.
              Auditez vous-même ce que fait Kagibi de vos données.
              La confiance ne devrait pas être aveugle.
            </p>
          </div>
        </div>
      </div>
    </section>

    <!-- Open Source Section -->
    <section class="opensource">
      <div class="container">
        <div class="opensource-inner">
          <div class="opensource-text">
            <div class="license-pill">AGPLv3</div>
            <h2>Un projet 100% open source</h2>
            <p>
              Kagibi est publié sous licence <strong>GNU AGPLv3</strong> — la licence copyleft la plus forte.
              Toute modification déployée en production doit être partagée publiquement.
              Aucune version propriétaire cachée, aucun backdoor possible.
            </p>
            <p>
              Le code est librement auditable, forçable et améliorable par tous.
              Contribuez à la sécurité, signalez des bugs, proposez des fonctionnalités —
              chaque Pull Request est la bienvenue.
            </p>
            <div class="opensource-actions">
              <a
                href="https://github.com/Buuuntyyy/Kagibi"
                target="_blank"
                rel="noopener noreferrer"
                class="btn-github"
              >
                <svg viewBox="0 0 24 24" fill="currentColor">
                  <path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0 0 24 12c0-6.63-5.37-12-12-12z"/>
                </svg>
                Voir le code source
              </a>
              <a
                href="https://github.com/Buuuntyyy/Kagibi"
                target="_blank"
                rel="noopener noreferrer"
                class="btn-star"
              >
                <svg viewBox="0 0 24 24" fill="currentColor">
                  <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
                </svg>
                Star on GitHub
              </a>
            </div>
          </div>
          <div class="opensource-stats">
            <div class="stat-card">
              <div class="stat-icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="16 18 22 12 16 6"></polyline>
                  <polyline points="8 6 2 12 8 18"></polyline>
                </svg>
              </div>
              <div class="stat-label">Licence</div>
              <div class="stat-value">AGPLv3</div>
            </div>
            <div class="stat-card">
              <div class="stat-icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <circle cx="12" cy="12" r="3"></circle>
                  <path d="M19.07 4.93A10 10 0 0 0 4.93 19.07"></path>
                  <path d="M2 12h2M20 12h2M12 2v2M12 20v2"></path>
                </svg>
              </div>
              <div class="stat-label">Contributions</div>
              <div class="stat-value">Ouvertes</div>
            </div>
            <div class="stat-card">
              <div class="stat-icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M9 19c-5 1.5-5-2.5-7-3m14 6v-3.87a3.37 3.37 0 0 0-.94-2.61c3.14-.35 6.44-1.54 6.44-7A5.44 5.44 0 0 0 20 4.77 5.07 5.07 0 0 0 19.91 1S18.73.65 16 2.48a13.38 13.38 0 0 0-7 0C6.27.65 5.09 1 5.09 1A5.07 5.07 0 0 0 5 4.77a5.44 5.44 0 0 0-1.5 3.78c0 5.42 3.3 6.61 6.44 7A3.37 3.37 0 0 0 9 18.13V22"></path>
                </svg>
              </div>
              <div class="stat-label">Plateforme</div>
              <div class="stat-value">GitHub</div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Final CTA -->
    <section class="final-cta">
      <div class="container">
        <h2>Prêt à reprendre le contrôle ?</h2>
        <p>Rejoignez Kagibi et stockez vos fichiers sans jamais compromettre votre vie privée.</p>
        <div class="cta-buttons">
          <router-link to="/login" class="btn btn-primary btn-large">
            Créer mon compte gratuitement
          </router-link>
          <router-link to="/transfer" class="btn btn-outline btn-large">
            Essayer le transfert P2P
          </router-link>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import LandingNav from '../../components/landing/LandingNav.vue'

onMounted(() => {
  const observerOptions = {
    threshold: 0.1,
    rootMargin: '0px 0px -80px 0px'
  }

  const observer = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        entry.target.classList.add('visible')
      }
    })
  }, observerOptions)

  document.querySelectorAll('.feature-card, .step').forEach(el => {
    observer.observe(el)
  })
})
</script>

<style scoped>
.landing-home {
  background: var(--card-color);
  color: var(--main-text-color);
  overflow-y: auto;
  height: 100vh;
}

/* Hero Section */
.hero {
  min-height: 90vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, var(--background-color) 0%, var(--card-color) 100%);
  position: relative;
  overflow: hidden;
}

.hero::before {
  content: '';
  position: absolute;
  top: 0;
  right: 0;
  width: 50%;
  height: 100%;
  background: radial-gradient(circle at center, var(--primary-color), transparent 70%);
  opacity: 0.08;
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 2rem;
  position: relative;
  color: var(--text-secondary);
}

.hero-content {
  text-align: center;
  max-width: 820px;
  margin: 0 auto;
  animation: fadeInUp 0.8s ease-out;
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
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

.hero-title {
  font-size: clamp(2.5rem, 6vw, 4.2rem);
  font-weight: 800;
  line-height: 1.15;
  margin-bottom: 1.5rem;
  letter-spacing: -0.03em;
  color: var(--main-text-color);
}

.highlight {
  color: var(--primary-color);
}

.hero-subtitle {
  font-size: 1.2rem;
  color: var(--secondary-text-color);
  line-height: 1.7;
  margin-bottom: 2.5rem;
  max-width: 680px;
  margin-left: auto;
  margin-right: auto;
}

.hero-cta {
  display: flex;
  gap: 1rem;
  justify-content: center;
  flex-wrap: wrap;
  margin-bottom: 3rem;
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

.btn-secondary {
  background: transparent;
  border-color: var(--border-color);
  color: var(--main-text-color);
}

.btn-secondary:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
  background: var(--hover-background-color);
}

.btn-outline {
  background: transparent;
  border: 2px solid var(--border-color);
  color: var(--main-text-color);
}

.btn-outline:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
  background: var(--hover-background-color);
}

.btn-large {
  padding: 1.25rem 3rem;
  font-size: 1.125rem;
}

.trust-badges {
  display: flex;
  gap: 1rem;
  justify-content: center;
  flex-wrap: wrap;
}

.badge {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.6rem 1rem;
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  font-size: 0.8rem;
  font-weight: 500;
  color: var(--secondary-text-color);
}

.badge svg {
  width: 16px;
  height: 16px;
  color: var(--primary-color);
  flex-shrink: 0;
}

/* How it works */
.how-it-works {
  padding: 6rem 0;
  background: var(--background-color);
}

.section-title {
  text-align: center;
  font-size: clamp(1.8rem, 4vw, 2.8rem);
  margin-bottom: 0.75rem;
  font-weight: 700;
  color: var(--main-text-color);
}

.section-subtitle {
  text-align: center;
  font-size: 1.1rem;
  color: var(--secondary-text-color);
  margin-bottom: 4rem;
}

.steps-grid {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  flex-wrap: wrap;
}

.step {
  flex: 1;
  min-width: 220px;
  max-width: 280px;
  text-align: center;
  padding: 2rem 1.5rem;
  opacity: 0;
  transform: translateY(20px);
  transition: opacity 0.5s ease, transform 0.5s ease;
}

.step.visible {
  opacity: 1;
  transform: translateY(0);
}

.step-number {
  width: 52px;
  height: 52px;
  background: var(--primary-color);
  color: var(--card-color);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.25rem;
  font-weight: 700;
  margin: 0 auto 1.25rem;
}

.step h3 {
  font-size: 1.1rem;
  font-weight: 600;
  margin-bottom: 0.75rem;
  color: var(--main-text-color);
}

.step p {
  font-size: 0.9rem;
  color: var(--secondary-text-color);
  line-height: 1.65;
}

.step-arrow {
  flex-shrink: 0;
  color: var(--secondary-text-color);
  opacity: 0.4;
}

.step-arrow svg {
  width: 28px;
  height: 28px;
}

/* Features Section */
.features {
  padding: 6rem 0;
  background: var(--card-color);
}

.features-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 1.5rem;
  margin-top: 0;
}

.feature-card {
  background: var(--background-color);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 2.25rem;
  transition: all 0.3s ease;
  opacity: 0;
  transform: translateY(20px);
}

.feature-card.visible {
  opacity: 1;
  transform: translateY(0);
}

.feature-card:hover {
  border-color: var(--primary-color);
  transform: translateY(-4px);
  box-shadow: 0 12px 30px rgba(0, 0, 0, 0.08);
}

.feature-icon {
  width: 52px;
  height: 52px;
  background: color-mix(in srgb, var(--primary-color) 12%, transparent);
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1.25rem;
}

.feature-icon svg {
  width: 26px;
  height: 26px;
  color: var(--primary-color);
}

.feature-card h3 {
  font-size: 1.2rem;
  margin-bottom: 0.75rem;
  font-weight: 600;
  color: var(--main-text-color);
}

.feature-card p {
  color: var(--secondary-text-color);
  line-height: 1.7;
  font-size: 0.95rem;
}

/* Open Source Section */
.opensource {
  padding: 6rem 0;
  background: var(--background-color);
}

.opensource-inner {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 4rem;
  align-items: center;
}

.license-pill {
  display: inline-block;
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: #16a34a;
  background: rgba(22, 163, 74, 0.1);
  border: 1px solid rgba(22, 163, 74, 0.3);
  padding: 0.3rem 0.9rem;
  border-radius: 999px;
  margin-bottom: 1.25rem;
}

.opensource-text h2 {
  font-size: clamp(1.8rem, 3.5vw, 2.5rem);
  font-weight: 800;
  margin-bottom: 1rem;
  color: var(--main-text-color);
}

.opensource-text p {
  color: var(--secondary-text-color);
  line-height: 1.75;
  font-size: 1rem;
  margin-bottom: 0.75rem;
}

.opensource-text strong {
  color: var(--main-text-color);
}

.opensource-actions {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
  margin-top: 2rem;
}

.btn-github {
  display: inline-flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.75rem 1.5rem;
  background: var(--main-text-color);
  color: var(--card-color);
  text-decoration: none;
  border-radius: 8px;
  font-weight: 600;
  font-size: 0.95rem;
  transition: all 0.3s ease;
}

.btn-github svg {
  width: 18px;
  height: 18px;
}

.btn-github:hover {
  opacity: 0.85;
  transform: translateY(-2px);
}

.btn-star {
  display: inline-flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.75rem 1.5rem;
  background: transparent;
  color: #b45309;
  border: 2px solid #d97706;
  text-decoration: none;
  border-radius: 8px;
  font-weight: 600;
  font-size: 0.95rem;
  transition: all 0.3s ease;
}

.btn-star svg {
  width: 16px;
  height: 16px;
  color: #f59e0b;
}

.btn-star:hover {
  background: rgba(245, 158, 11, 0.1);
  border-color: #f59e0b;
  transform: translateY(-2px);
}

.opensource-stats {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem 1.5rem;
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 10px;
  min-width: 200px;
}

.stat-icon {
  width: 36px;
  height: 36px;
  background: color-mix(in srgb, var(--primary-color) 12%, transparent);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.stat-icon svg {
  width: 18px;
  height: 18px;
  color: var(--primary-color);
}

.stat-label {
  font-size: 0.75rem;
  color: var(--secondary-text-color);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.stat-value {
  font-size: 1rem;
  font-weight: 700;
  color: var(--main-text-color);
}

/* Final CTA */
.final-cta {
  padding: 7rem 0;
  background: linear-gradient(135deg, var(--background-color) 0%, var(--card-color) 100%);
  text-align: center;
}

.final-cta h2 {
  font-size: clamp(2rem, 4vw, 3rem);
  font-weight: 800;
  margin-bottom: 1rem;
  color: var(--main-text-color);
}

.final-cta p {
  font-size: 1.15rem;
  color: var(--secondary-text-color);
  margin-bottom: 2.5rem;
  max-width: 520px;
  margin-left: auto;
  margin-right: auto;
}

.cta-buttons {
  display: flex;
  gap: 1rem;
  justify-content: center;
  flex-wrap: wrap;
}

/* Responsive */
@media (max-width: 768px) {
  .hero-title {
    font-size: 2.2rem;
  }

  .hero-subtitle {
    font-size: 1rem;
  }

  .hero-cta,
  .cta-buttons {
    flex-direction: column;
    align-items: center;
  }

  .btn {
    width: 100%;
    justify-content: center;
  }

  .trust-badges {
    flex-direction: column;
    align-items: center;
  }

  .features-grid {
    grid-template-columns: 1fr;
  }

  .opensource-inner {
    grid-template-columns: 1fr;
  }

  .opensource-stats {
    flex-direction: row;
    flex-wrap: wrap;
    justify-content: center;
  }

  .stat-card {
    min-width: 140px;
    flex: 1;
  }

  .steps-grid {
    flex-direction: column;
    align-items: center;
  }

  .step-arrow {
    transform: rotate(90deg);
  }
}
</style>
