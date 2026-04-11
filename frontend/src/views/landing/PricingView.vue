<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="pricing-page">
    <LandingNav />
    <section class="pricing-hero">
      <div class="container">
        <h1>Tarification Simple et Transparente</h1>
        <p>Commencez gratuitement, évoluez selon vos besoins. Aucun engagement.</p>
      </div>
    </section>

    <section class="pricing-section">
      <div class="container">
        <div class="pricing-grid-full">
          <!-- Free Plan -->
          <div class="pricing-card-full">
            <h3 class="plan-name">Gratuit</h3>
            <div class="plan-price">
              <span class="amount">0€</span>
              <span class="period">/mois</span>
            </div>
            <p class="plan-description">Parfait pour découvrir Kagibi</p>
            <ul class="plan-features">
              <li>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="20 6 9 17 4 12"></polyline>
                </svg>
                20 Go de stockage chiffré
              </li>
              <li>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="20 6 9 17 4 12"></polyline>
                </svg>
                Transferts P2P illimités
              </li>
              <li>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="20 6 9 17 4 12"></polyline>
                </svg>
                Chiffrement E2E (AES-256)
              </li>
              <li>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="20 6 9 17 4 12"></polyline>
                </svg>
                Partage sécurisé par lien
              </li>
            </ul>
            <router-link to="/login" class="btn btn-outline">Commencer</router-link>
          </div>

          <!-- Personal Plan -->
          <div class="pricing-card-full">
            <h3 class="plan-name">Personnel</h3>
            <div class="plan-price">
              <span class="amount">5€</span>
              <span class="period">/mois</span>
            </div>
            <p class="plan-description">Pour un usage personnel intensif</p>
            <ul class="plan-features">
              <li>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="20 6 9 17 4 12"></polyline>
                </svg>
                100 Go de stockage
              </li>
              <li>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="20 6 9 17 4 12"></polyline>
                </svg>
                Transferts P2P illimités
              </li>
              <li>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="20 6 9 17 4 12"></polyline>
                </svg>
                Historique d'activité 90 jours
              </li>
              <li>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="20 6 9 17 4 12"></polyline>
                </svg>
                Support prioritaire
              </li>
            </ul>
            <router-link to="/login" class="btn btn-outline">Choisir ce plan</router-link>
          </div>

          <!-- Expert Plan (Featured) -->
          <div class="pricing-card-full featured">
            <div class="featured-badge">Le plus populaire</div>
            <h3 class="plan-name">Expert</h3>
            <div class="plan-price">
              <span class="amount">15€</span>
              <span class="period">/mois</span>
            </div>
            <p class="plan-description">Pour les professionnels exigeants</p>
            <ul class="plan-features">
              <li>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="20 6 9 17 4 12"></polyline>
                </svg>
                1 To de stockage
              </li>
              <li>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="20 6 9 17 4 12"></polyline>
                </svg>
                Transferts P2P illimités
              </li>
              <li>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="20 6 9 17 4 12"></polyline>
                </svg>
                Historique illimité
              </li>
              <li>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="20 6 9 17 4 12"></polyline>
                </svg>
                Support prioritaire 24/7
              </li>
              <li>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="20 6 9 17 4 12"></polyline>
                </svg>
                Partages avancés
              </li>
            </ul>
            <router-link to="/login" class="btn btn-primary">Choisir ce plan</router-link>
          </div>

          <!-- Business Plan -->
          <div class="pricing-card-full">
            <h3 class="plan-name">Business</h3>
            <div class="plan-price">
              <span class="amount">49€</span>
              <span class="period">/mois</span>
            </div>
            <p class="plan-description">Pour les équipes et entreprises</p>
            <ul class="plan-features">
              <li>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="20 6 9 17 4 12"></polyline>
                </svg>
                3 To de stockage
              </li>
              <li>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="20 6 9 17 4 12"></polyline>
                </svg>
                Comptes multi-utilisateurs
              </li>
              <li>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="20 6 9 17 4 12"></polyline>
                </svg>
                Administration centralisée
              </li>
              <li>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="20 6 9 17 4 12"></polyline>
                </svg>
                SLA 99.9%
              </li>
              <li>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="20 6 9 17 4 12"></polyline>
                </svg>
                Support dédié
              </li>
            </ul>
            <router-link to="/login" class="btn btn-outline">Choisir ce plan</router-link>
          </div>
        </div>

        <!-- Pay-As-You-Go -->
        <div class="payg-section">
          <h3>Option Flexible : Pay-As-You-Go</h3>
          <p>
            Besoin de plus d'espace ? Ajoutez du stockage à la demande pour <strong>0.05€/Go/mois</strong>. 
            Parfait pour les besoins évolutifs sans engagement.
          </p>
        </div>
      </div>
    </section>

    <!-- FAQ Section -->
    <section class="faq-section">
      <div class="container">
        <h2>Questions Fréquentes</h2>
        <div class="faq-grid">
          <div class="faq-item">
            <h4>Puis-je changer de plan à tout moment ?</h4>
            <p>Oui, vous pouvez upgrader ou downgrader votre plan à tout moment. Les changements sont appliqués immédiatement.</p>
          </div>
          <div class="faq-item">
            <h4>Que se passe-t-il si je dépasse mon quota ?</h4>
            <p>Vous serez notifié et pourrez soit passer à un plan supérieur, soit ajouter du stockage Pay-As-You-Go.</p>
          </div>
          <div class="faq-item">
            <h4>Les transferts P2P sont-ils vraiment illimités ?</h4>
            <p>Oui, aucune limite de taille ou de nombre de transferts. Utilisez toute votre bande passante.</p>
          </div>
          <div class="faq-item">
            <h4>Comment fonctionne le chiffrement Zero-Knowledge ?</h4>
            <p>Vos fichiers sont chiffrés avec votre clé avant même de quitter votre appareil. Nous ne pouvons pas les déchiffrer.</p>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import LandingNav from '../../components/landing/LandingNav.vue'
</script>

<style scoped>
.pricing-page {
  background: var(--card-color);
  color: var(--main-text-color);
  min-height: 100vh;
  overflow-y: auto;
  height: 100vh;
}

.pricing-hero {
  padding: 6rem 2rem 4rem;
  text-align: center;
  background: linear-gradient(135deg, var(--background-color) 0%, var(--card-color) 100%);
}

.pricing-hero h1 {
  font-size: clamp(2.5rem, 5vw, 4rem);
  margin-bottom: 1rem;
  font-weight: 700;
  color: var(--main-text-color);
}

.pricing-hero p {
  font-size: 1.25rem;
  color: var(--secondary-text-color);
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 2rem;
}

.pricing-section {
  padding: 4rem 0;
}

.pricing-grid-full {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 2rem;
}

.pricing-card-full {
  background: var(--card-color);
  border: 2px solid var(--border-color);
  border-radius: 8px;
  padding: 2.5rem;
  position: relative;
  transition: all 0.3s ease;
}

.pricing-card-full:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 30px rgba(0, 0, 0, 0.08);
}

.pricing-card-full.featured {
  border-color: var(--primary-color);
  background: linear-gradient(135deg, var(--card-color) 0%, var(--hover-background-color) 100%);
}

.featured-badge {
  position: absolute;
  top: -12px;
  left: 50%;
  transform: translateX(-50%);
  background: var(--primary-color);
  color: var(--card-color);
  padding: 0.25rem 1rem;
  border-radius: 8px;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
}

.plan-name {
  font-size: 1.75rem;
  font-weight: 600;
  margin-bottom: 1rem;
  color: var(--main-text-color);
}

.plan-price {
  margin-bottom: 1rem;
}

.amount {
  font-size: 3rem;
  font-weight: 700;
  color: var(--primary-color);
}

.period {
  color: var(--secondary-text-color);
  font-size: 1rem;
}

.plan-description {
  color: var(--secondary-text-color);
  margin-bottom: 2rem;
  font-size: 0.95rem;
}

.plan-features {
  list-style: none;
  margin-bottom: 2rem;
}

.plan-features li {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 0;
  color: var(--secondary-text-color);
  border-bottom: 1px solid var(--border-color);
}

.plan-features li:last-child {
  border-bottom: none;
}

.plan-features svg {
  width: 20px;
  height: 20px;
  color: var(--primary-color);
  flex-shrink: 0;
}

.btn {
  display: block;
  width: 100%;
  padding: 1rem;
  text-align: center;
  border-radius: 8px;
  font-weight: 600;
  text-decoration: none;
  transition: all 0.3s ease;
}

.btn-primary {
  background: var(--primary-color);
  color: var(--card-color);
  border: 2px solid var(--primary-color);
}

.btn-primary:hover {
  background: var(--accent-color);
  border-color: var(--accent-color);
  transform: translateY(-2px);
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

.payg-section {
  margin-top: 4rem;
  padding: 3rem;
  background: var(--card-color);
  border: 2px solid var(--border-color);
  border-radius: 8px;
  text-align: center;
}

.payg-section h3 {
  font-size: 1.75rem;
  margin-bottom: 1rem;
  color: var(--main-text-color);
}

.payg-section p {
  color: var(--secondary-text-color);
  font-size: 1.125rem;
  line-height: 1.7;
}

.faq-section {
  padding: 6rem 0;
  background: var(--background-color);
}

.faq-section h2 {
  text-align: center;
  font-size: 2.5rem;
  margin-bottom: 3rem;
  color: var(--main-text-color);
}

.faq-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 2rem;
}

.faq-item {
  background: var(--card-color);
  padding: 2rem;
  border-radius: 8px;
  border: 1px solid var(--border-color);
}

.faq-item h4 {
  font-size: 1.25rem;
  margin-bottom: 1rem;
  color: var(--primary-color);
}

.faq-item p {
  color: var(--secondary-text-color);
  line-height: 1.7;
}

@media (max-width: 768px) {
  .pricing-grid-full {
    grid-template-columns: 1fr;
  }
}
</style>
