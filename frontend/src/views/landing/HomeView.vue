<template>
  <div class="landing-home">
    <LandingNav />
    <!-- Hero Section -->
    <section class="hero">
      <div class="container">
        <div class="hero-content">
          <h1 class="hero-title">
            Vos données, vos clés,<br>
            <span class="highlight">notre territoire</span>
          </h1>
          <p class="hero-subtitle">
            Le seul Cloud Drive Zero-Knowledge hébergé en France. 
            Chiffrement client-side, transferts P2P instantanés, immunité totale au Cloud Act.
          </p>
          <div class="hero-cta">
            <router-link to="/dashboard" class="btn btn-primary">
              Accéder au Drive
            </router-link>
            <router-link to="/transfer" class="btn btn-secondary">
              Transfert P2P Gratuit
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
          </div>
        </div>
      </div>
    </section>

    <!-- Features Section -->
    <section class="features">
      <div class="container">
        <h2 class="section-title">Sécurité Sans Compromis</h2>
        <div class="features-grid">
          <div class="feature-card">
            <div class="feature-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
                <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
              </svg>
            </div>
            <h3>Zero-Knowledge Absolu</h3>
            <p>
              Chiffrement AES-256-GCM côté client. Vos clés ne quittent jamais votre appareil. 
              Impossibilité mathématique de lire vos données.
            </p>
          </div>

          <div class="feature-card">
            <div class="feature-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="22" y1="2" x2="11" y2="13"></line>
                <polygon points="22 2 15 22 11 13 2 9 22 2"></polygon>
              </svg>
            </div>
            <h3>Transfert P2P Instantané</h3>
            <p>
              Envoyez jusqu'à 10 Go en direct via serveur TURN. Aucun upload/download, 
              vitesse maximale, aucune trace sur nos serveurs.
            </p>
          </div>

          <div class="feature-card">
            <div class="feature-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"></path>
              </svg>
            </div>
            <h3>Souveraineté Numérique</h3>
            <p>
              Infrastructure 100% OVH France (Gravelines). Immunité au Cloud Act. 
              Conformité RGPD stricte. Vos données restent européennes.
            </p>
          </div>
        </div>
      </div>
    </section>

    <!-- Final CTA -->
    <section class="final-cta">
      <div class="container">
        <h2>Reprenez le Contrôle de Vos Données</h2>
        <router-link to="/login" class="btn btn-primary btn-large">
          Créer mon compte
        </router-link>
      </div>
    </section>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import LandingNav from '../../components/landing/LandingNav.vue'

onMounted(() => {
  // Animations d'apparition au scroll
  const observerOptions = {
    threshold: 0.1,
    rootMargin: '0px 0px -100px 0px'
  }

  const observer = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        entry.target.classList.add('visible')
      }
    })
  }, observerOptions)

  document.querySelectorAll('.feature-card').forEach(el => {
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
  opacity: 0.1;
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 2rem;
  position: relative;
  color : var(--text-secondary);
}

.hero-content {
  text-align: center;
  max-width: 800px;
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

.hero-title {
  font-size: clamp(2.5rem, 6vw, 4rem);
  font-weight: 700;
  line-height: 1.2;
  margin-bottom: 1.5rem;
  letter-spacing: -0.02em;
}

.highlight {
  color: var(--primary-color);
  position: relative;
}

.hero-subtitle {
  font-size: 1.25rem;
  color: var(--secondary-text-color);
  line-height: 1.6;
  margin-bottom: 2.5rem;
  max-width: 700px;
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
  padding: 0.75rem 1.5rem;
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
  gap: 2rem;
  justify-content: center;
  flex-wrap: wrap;
}

.badge {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1.25rem;
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  font-size: 0.875rem;
  color: var(--secondary-text-color);
}

.badge svg {
  width: 18px;
  height: 18px;
  color: var(--primary-color);
}

/* Features Section */
.features {
  padding: 6rem 0;
  background: var(--background-color);
}

.section-title {
  text-align: center;
  font-size: clamp(2rem, 4vw, 3rem);
  margin-bottom: 1rem;
  font-weight: 700;
  color: var(--main-text-color);
}

.section-subtitle {
  text-align: center;
  font-size: 1.125rem;
  color: var(--secondary-text-color);
  margin-bottom: 4rem;
}

.features-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 2rem;
  margin-top: 4rem;
}

.feature-card {
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 2.5rem;
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
  width: 60px;
  height: 60px;
  background: var(--hover-background-color);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1.5rem;
}

.feature-icon svg {
  width: 30px;
  height: 30px;
  color: var(--primary-color);
}

.feature-card h3 {
  font-size: 1.5rem;
  margin-bottom: 1rem;
  font-weight: 600;
  color: var(--main-text-color);
}

.feature-card p {
  color: var(--secondary-text-color);
  line-height: 1.7;
}

/* Pricing Preview */
.pricing-preview {
  padding: 6rem 0;
  background: linear-gradient(180deg, var(--card-color) 0%, var(--background-color) 100%);
}

.pricing-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 1.5rem;
  margin-top: 3rem;
}

.pricing-card {
  background: var(--card-color);
  border: 2px solid var(--border-color);
  border-radius: 8px;
  padding: 2.5rem;
  text-align: center;
  position: relative;
  transition: all 0.3s ease;
  opacity: 0;
  transform: translateY(20px);
}

.pricing-card.visible {
  opacity: 1;
  transform: translateY(0);
}

.pricing-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 30px rgba(0, 0, 0, 0.08);
}

.pricing-card.featured {
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
  letter-spacing: 0.05em;
}

.plan-name {
  font-size: 1.5rem;
  font-weight: 600;
  margin-bottom: 1rem;
  color: var(--main-text-color);
}

.plan-price {
  margin-bottom: 2rem;
}

.amount {
  font-size: 3rem;
  font-weight: 700;
  color: var(--primary-color);
}

.amount-small {
  font-size: 2rem;
  font-weight: 700;
  color: var(--primary-color);
}

.period {
  color: var(--secondary-text-color);
  font-size: 1rem;
}

.plan-features {
  list-style: none;
  margin-bottom: 2rem;
  text-align: left;
}

.plan-features li {
  padding: 0.75rem 0;
  color: var(--secondary-text-color);
  border-bottom: 1px solid var(--border-color);
}

.plan-features li:last-child {
  border-bottom: none;
}

.pricing-note {
  text-align: center;
  margin-top: 3rem;
}

.link-primary {
  color: var(--primary-color);
  text-decoration: none;
  font-weight: 600;
  transition: opacity 0.3s ease;
}

.link-primary:hover {
  opacity: 0.8;
}

/* Final CTA */
.final-cta {
  padding: 6rem 0;
  background: var(--background-color);
  text-align: center;
}

.final-cta h2 {
  font-size: clamp(2rem, 4vw, 3rem);
  margin-bottom: 1rem;
  color: var(--main-text-color);
}

.final-cta p {
  font-size: 1.125rem;
  color: var(--secondary-text-color);
  margin-bottom: 2.5rem;
}

/* Responsive */
@media (max-width: 768px) {
  .hero-title {
    font-size: 2rem;
  }

  .hero-subtitle {
    font-size: 1rem;
  }

  .hero-cta {
    flex-direction: column;
  }

  .btn {
    width: 100%;
    justify-content: center;
  }

  .trust-badges {
    flex-direction: column;
    align-items: center;
  }

  .features-grid,
  .pricing-grid {
    grid-template-columns: 1fr;
  }
}
</style>
