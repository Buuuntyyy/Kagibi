<template>
  <div class="transfer-page">
    <LandingNav />
    <section class="transfer-hero">
      <div class="container">
        <h1>Transfert P2P Ultra-Rapide</h1>
        <p>Envoyez jusqu'à 10 Go en direct, sans upload. Vitesse maximale, zéro trace.</p>
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
            <h2>Comment ça fonctionne ?</h2>
            <div class="steps">
              <div class="step">
                <div class="step-number">1</div>
                <div class="step-content">
                  <h4>Sélectionnez vos fichiers</h4>
                  <p>Choisissez les fichiers à envoyer (10 Go max)</p>
                </div>
              </div>
              <div class="step">
                <div class="step-number">2</div>
                <div class="step-content">
                  <h4>Obtenez un lien unique</h4>
                  <p>Un lien sécurisé est généré instantanément</p>
                </div>
              </div>
              <div class="step">
                <div class="step-number">3</div>
                <div class="step-content">
                  <h4>Partagez et transférez</h4>
                  <p>Connexion P2P directe via serveur TURN</p>
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
              <h4>Vitesse Maximale</h4>
              <p>Pas d'upload serveur, transfert direct P2P à pleine bande passante</p>
            </div>

            <div class="feature-box">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
                <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
              </svg>
              <h4>Chiffrement E2E</h4>
              <p>Vos fichiers sont chiffrés de bout en bout pendant le transfert</p>
            </div>

            <div class="feature-box">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"></polyline>
              </svg>
              <h4>Zero-Trace</h4>
              <p>Aucun fichier n'est stocké sur nos serveurs, tranquillité absolue</p>
            </div>

            <div class="feature-box">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"></circle>
                <polyline points="12 6 12 12 16 14"></polyline>
              </svg>
              <h4>Disponible 24/7</h4>
              <p>Service disponible en permanence, aucune limite d'utilisation</p>
            </div>
          </div>

          <!-- CTA Card -->
          <div class="cta-card">
            <h3>Prêt à transférer ?</h3>
            <p>Connectez-vous pour commencer à utiliser le transfert P2P instantané</p>
            <router-link to="/login" class="btn btn-primary btn-large">
              Accéder au transfert
            </router-link>
            <div class="cta-note">
              Gratuit et illimité pour tous les utilisateurs
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Technical Details -->
    <section class="tech-section">
      <div class="container">
        <h2>Technologie WebRTC & TURN</h2>
        <div class="tech-grid">
          <div class="tech-card">
            <h4>WebRTC</h4>
            <p>
              Protocole de communication peer-to-peer standardisé par le W3C. 
              Permet une connexion directe entre navigateurs sans intermédiaire.
            </p>
          </div>
          <div class="tech-card">
            <h4>Serveur TURN</h4>
            <p>
              Infrastructure de relais pour les connexions impossibles en direct. 
              Garantit un transfert réussi même derrière un NAT strict.
            </p>
          </div>
          <div class="tech-card">
            <h4>Chiffrement AES</h4>
            <p>
              Chiffrement côté client avant transfert. Les clés sont échangées via 
              signaling sécurisé, jamais stockées.
            </p>
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
}
</style>
