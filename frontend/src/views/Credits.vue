<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="credits-page">
    <div class="page-header">
      <router-link to="/account" class="back-link">
        <svg viewBox="0 0 24 24" fill="currentColor" width="20" height="20">
          <path d="M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20v-2z"/>
        </svg>
        Retour
      </router-link>
      <h1>Mentions légales et crédits</h1>
    </div>

    <div class="content-section">
      <section class="legal-section">
        <h2>Éditeur du site</h2>
        <div class="legal-content">
          <p><strong>Kagibi</strong></p>
          <p>Application de stockage sécurisé Zero-Knowledge</p>
          <p>© 2026 Kagibi - Tous droits réservés</p>
        </div>
      </section>

      <section class="legal-section">
        <h2>Hébergement</h2>
        <div class="legal-content">
          <p>L'application est hébergée conformément aux normes de sécurité et de protection des données.</p>
        </div>
      </section>

      <section class="legal-section">
        <h2>Crédits et attributions</h2>
        <div class="legal-content">
          <h3>Icônes et illustrations</h3>
          <p>Les icônes utilisées dans cette application proviennent de :</p>
          
          <div class="attribution-item">
            <h4>Avatars (animaux et personnages)</h4>
            <p>
              Icônes créées par <strong>Freepik</strong> depuis 
              <a href="https://www.flaticon.com" target="_blank" rel="noopener noreferrer">Flaticon</a>
            </p>
            <ul class="icon-list">
              <li>Icône Chat - <a href="https://www.flaticon.com/free-icons/cat" target="_blank" rel="noopener">Cat icons created by Freepik - Flaticon</a></li>
              <li>Icône Chien - <a href="https://www.flaticon.com/free-icons/dog" target="_blank" rel="noopener">Dog icons created by Freepik - Flaticon</a></li>
              <li>Icône Renard - <a href="https://www.flaticon.com/free-icons/fox" target="_blank" rel="noopener">Fox icons created by Freepik - Flaticon</a></li>
              <li>Icône Lapin - <a href="https://www.flaticon.com/free-icons/rabbit" target="_blank" rel="noopener">Rabbit icons created by Freepik - Flaticon</a></li>
              <li>Icône Panda - <a href="https://www.flaticon.com/free-icons/panda" target="_blank" rel="noopener">Panda icons created by Freepik - Flaticon</a></li>
              <li>Icône Gorille - <a href="https://www.flaticon.com/free-icons/gorilla" target="_blank" rel="noopener">Gorilla icons created by Freepik - Flaticon</a></li>
              <li>Icône Poulet - <a href="https://www.flaticon.com/free-icons/chicken" target="_blank" rel="noopener">Chicken icons created by Freepik - Flaticon</a></li>
              <li>Icône Garçon - <a href="https://www.flaticon.com/free-icons/boy" target="_blank" rel="noopener">Boy icons created by Freepik - Flaticon</a></li>
              <li>Icône Fille - <a href="https://www.flaticon.com/free-icons/girl" target="_blank" rel="noopener">Girl icons created by Freepik - Flaticon</a></li>
            </ul>
          </div>

          <div class="attribution-item">
            <h4>Icône de compte utilisateur</h4>
            <p>
              Icône créée par <strong>Luch Phou</strong> depuis 
              <a href="https://www.flaticon.com" target="_blank" rel="noopener noreferrer">Flaticon</a>
            </p>
            <p>
              <a href="https://www.flaticon.com/free-icons/user-account" target="_blank" rel="noopener">User account icons created by Luch Phou - Flaticon</a>
            </p>
          </div>

          <div class="legal-note">
            <svg viewBox="0 0 24 24" fill="currentColor" width="20" height="20">
              <path d="M12,2A10,10 0 0,0 2,12A10,10 0 0,0 12,22A10,10 0 0,0 22,12A10,10 0 0,0 12,2M12,4A8,8 0 0,1 20,12A8,8 0 0,1 12,20A8,8 0 0,1 4,12A8,8 0 0,1 12,4M11,16.5L6.5,12L7.91,10.59L11,13.67L16.59,8.09L18,9.5L11,16.5Z"/>
            </svg>
            <p>
              Ces attributions sont conformes aux <a href="https://support.flaticon.com/s/article/Attribution-How-when-and-where-FI?language=en_US" target="_blank" rel="noopener">conditions d'utilisation de Flaticon</a> 
              et à la législation française en matière de droits d'auteur.
            </p>
          </div>
        </div>
      </section>

      <section class="legal-section">
        <h2>Propriété intellectuelle</h2>
        <div class="legal-content">
          <p>
            L'ensemble du site Kagibi, incluant sa structure, son design, son code source et son contenu textuel, 
            constitue une œuvre protégée par le droit d'auteur. Toute reproduction ou représentation, totale ou partielle, 
            sans autorisation expresse est interdite et constituerait une contrefaçon sanctionnée par les articles L.335-2 
            et suivants du Code de la propriété intellectuelle.
          </p>
        </div>
      </section>

      <section class="legal-section">
        <h2>Liens utiles</h2>
        <div class="legal-links-grid">
          <router-link to="/cgu" class="legal-card">
            <svg viewBox="0 0 24 24" fill="currentColor" width="24" height="24">
              <path d="M14,2H6A2,2 0 0,0 4,4V20A2,2 0 0,0 6,22H18A2,2 0 0,0 20,20V8L14,2M18,20H6V4H13V9H18V20Z"/>
            </svg>
            <div>
              <h3>CGU</h3>
              <p>Conditions Générales d'Utilisation</p>
            </div>
          </router-link>
          
          <router-link to="/privacy" class="legal-card">
            <svg viewBox="0 0 24 24" fill="currentColor" width="24" height="24">
              <path d="M12,1L3,5V11C3,16.55 6.84,21.74 12,23C17.16,21.74 21,16.55 21,11V5L12,1Z"/>
            </svg>
            <div>
              <h3>Confidentialité</h3>
              <p>Politique de confidentialité</p>
            </div>
          </router-link>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup>
// Page statique, pas de logique nécessaire
</script>

<style scoped>
.credits-page {
  padding: 2rem;
  background-color: var(--background-color);
  min-height: 100vh;
  max-width: 900px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 2rem;
}

.back-link {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  color: var(--primary-color);
  text-decoration: none;
  font-size: 0.9rem;
  margin-bottom: 1rem;
  transition: all 0.2s;
}

.back-link:hover {
  gap: 0.75rem;
  color: var(--accent-color);
}

.page-header h1 {
  font-size: 2rem;
  margin: 0;
  color: var(--main-text-color);
}

.content-section {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.legal-section {
  background: var(--card-color);
  border-radius: 12px;
  border: 1px solid var(--border-color);
  padding: 2rem;
}

.legal-section h2 {
  margin: 0 0 1.5rem 0;
  font-size: 1.5rem;
  color: var(--main-text-color);
  border-bottom: 2px solid var(--primary-color);
  padding-bottom: 0.5rem;
}

.legal-section h3 {
  margin: 1.5rem 0 1rem 0;
  font-size: 1.2rem;
  color: var(--main-text-color);
}

.legal-section h4 {
  margin: 1rem 0 0.5rem 0;
  font-size: 1rem;
  color: var(--primary-color);
  font-weight: 600;
}

.legal-content {
  color: var(--secondary-text-color);
  line-height: 1.8;
}

.legal-content p {
  margin: 0.75rem 0;
}

.legal-content strong {
  color: var(--main-text-color);
}

.legal-content a {
  color: var(--primary-color);
  text-decoration: none;
  font-weight: 500;
}

.legal-content a:hover {
  text-decoration: underline;
}

.attribution-item {
  margin: 1.5rem 0;
  padding: 1.5rem;
  background: var(--background-color);
  border-radius: 8px;
  border-left: 4px solid var(--primary-color);
}

.icon-list {
  margin: 1rem 0;
  padding-left: 1.5rem;
  list-style: none;
}

.icon-list li {
  margin: 0.5rem 0;
  position: relative;
  padding-left: 1rem;
}

.icon-list li::before {
  content: "→";
  position: absolute;
  left: -0.5rem;
  color: var(--primary-color);
  font-weight: bold;
}

.legal-note {
  display: flex;
  gap: 1rem;
  margin-top: 2rem;
  padding: 1rem;
  background: rgba(99, 102, 241, 0.1);
  border-radius: 8px;
  border: 1px solid var(--primary-color);
}

.legal-note svg {
  flex-shrink: 0;
  color: var(--primary-color);
  margin-top: 0.2rem;
}

.legal-note p {
  margin: 0;
  font-size: 0.9rem;
  color: var(--main-text-color);
}

.legal-links-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1rem;
  margin-top: 1rem;
}

.legal-card {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1.5rem;
  background: var(--background-color);
  border-radius: 8px;
  border: 1px solid var(--border-color);
  text-decoration: none;
  transition: all 0.2s;
}

.legal-card:hover {
  border-color: var(--primary-color);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.legal-card svg {
  color: var(--primary-color);
  flex-shrink: 0;
}

.legal-card h3 {
  margin: 0;
  font-size: 1.1rem;
  color: var(--main-text-color);
}

.legal-card p {
  margin: 0.25rem 0 0 0;
  font-size: 0.85rem;
  color: var(--secondary-text-color);
}

@media (max-width: 768px) {
  .credits-page {
    padding: 1rem;
  }

  .page-header h1 {
    font-size: 1.5rem;
  }

  .legal-section {
    padding: 1.5rem;
  }

  .legal-section h2 {
    font-size: 1.3rem;
  }

  .attribution-item {
    padding: 1rem;
  }

  .legal-links-grid {
    grid-template-columns: 1fr;
  }
}
</style>
