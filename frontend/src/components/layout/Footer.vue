<template>
  <footer class="site-footer">
    <div class="footer-content">
      <div class="footer-section">
        <p class="footer-text">© 2026 SaferCloud - Stockage sécurisé Zero-Knowledge</p>
      </div>
      <div class="footer-section credits">
        <button @click="showCredits = !showCredits" class="credits-toggle">
          <svg viewBox="0 0 24 24" class="icon">
            <path fill="currentColor" d="M12,2A10,10 0 0,0 2,12A10,10 0 0,0 12,22A10,10 0 0,0 22,12A10,10 0 0,0 12,2M12,4A8,8 0 0,1 20,12A8,8 0 0,1 12,20A8,8 0 0,1 4,12A8,8 0 0,1 12,4M11,16.5L6.5,12L7.91,10.59L11,13.67L16.59,8.09L18,9.5L11,16.5Z"/>
          </svg>
          Crédits
        </button>
        <Transition name="fade">
          <div v-if="showCredits" class="credits-panel">
            <h4>Attributions des icônes</h4>
            <div class="credits-list">
              <p class="credit-item">
                Icônes d'avatars (animaux, personnages) créées par 
                <a href="https://www.flaticon.com/free-icons/animals" target="_blank" rel="noopener noreferrer">Freepik - Flaticon</a>
              </p>
              <p class="credit-item">
                Icône de compte utilisateur créée par 
                <a href="https://www.flaticon.com/free-icons/user-account" target="_blank" rel="noopener noreferrer">Luch Phou - Flaticon</a>
              </p>
            </div>
          </div>
        </Transition>
      </div>
    </div>
  </footer>
</template>

<script setup>
import { ref } from 'vue'

const showCredits = ref(false)
</script>

<style scoped>
.site-footer {
  background-color: var(--background-color);
  border-top: 1px solid var(--border-color);
  padding: 1rem 2rem;
  margin-top: auto;
  width: 100%;
  position: relative;
}

.footer-content {
  max-width: 1400px;
  margin: 0 auto;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
}

.footer-section {
  display: flex;
  align-items: center;
}

.footer-text {
  font-size: 0.875rem;
  color: var(--secondary-text-color);
  margin: 0;
}

.credits {
  position: relative;
}

.credits-toggle {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  background: none;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  color: var(--secondary-text-color);
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.credits-toggle:hover {
  background-color: var(--hover-background-color);
  border-color: var(--primary-color);
  color: var(--main-text-color);
}

.credits-toggle .icon {
  width: 16px;
  height: 16px;
}

.credits-panel {
  position: absolute;
  bottom: 100%;
  right: 0;
  margin-bottom: 0.5rem;
  background-color: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  padding: 1rem;
  min-width: 320px;
  max-width: 400px;
  z-index: 100;
}

.credits-panel h4 {
  margin: 0 0 0.75rem 0;
  font-size: 0.95rem;
  color: var(--main-text-color);
  font-weight: 600;
}

.credits-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.credit-item {
  font-size: 0.8rem;
  color: var(--secondary-text-color);
  line-height: 1.5;
  margin: 0;
}

.credit-item a {
  color: var(--primary-color);
  text-decoration: none;
  font-weight: 500;
}

.credit-item a:hover {
  text-decoration: underline;
}

/* Animations */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.fade-enter-from {
  opacity: 0;
  transform: translateY(10px);
}

.fade-leave-to {
  opacity: 0;
  transform: translateY(10px);
}

/* Responsive */
@media (max-width: 768px) {
  .footer-content {
    flex-direction: column;
    text-align: center;
  }
  
  .credits-panel {
    right: auto;
    left: 50%;
    transform: translateX(-50%);
    min-width: 280px;
  }
  
  .fade-enter-from,
  .fade-leave-to {
    transform: translateX(-50%) translateY(10px);
  }
}
</style>
