<template>
  <div class="auth-page">
    <div class="auth-container card">
      <Transition name="auth-flip" mode="out-in">
        <div v-if="mode === 'login'" key="login">
          <h1>Se connecter</h1>
          <LoginComponent />
          
          <div class="security-msg">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="info-icon">
              <circle cx="12" cy="12" r="10"></circle>
              <line x1="12" y1="16" x2="12" y2="12"></line>
              <line x1="12" y1="8" x2="12.01" y2="8"></line>
            </svg>
            <p>
              Pour votre sécurité, nous recommandons l'utilisation d'un gestionnaire de mots de passe comme <strong>Bitwarden</strong> ou <strong>KeePass</strong>.
            </p>
          </div>

          <div class="oauth-unavailable">
            <small>⚠️ La connexion via des fournisseurs tiers (Google/GitHub) est désactivée car elle est incompatible avec notre chiffrement de bout en bout (Zéro-Knowledge).</small>
          </div>

          <div class="auth-links">
            <p>
              Pas encore de compte ? <a href="#" @click.prevent="mode = 'register'">S'inscrire</a>
            </p>
            <p>
              Mot de passe oublié ? <a href="#" @click.prevent="mode = 'recovery'">Récupérer mon compte</a>
            </p>
          </div>
        </div>
        <div v-else-if="mode === 'register'" key="register">
          <h1>Créer un compte</h1>
          <RegisterComponent />

          <div class="security-msg warning">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="info-icon">
              <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
              <line x1="12" y1="9" x2="12" y2="13"></line>
              <line x1="12" y1="17" x2="12.01" y2="17"></line>
            </svg>
            <p>
              <strong>Attention :</strong> Si vous perdez votre mot de passe, vos fichiers seront définitivement perdus. Nous ne pouvons PAS le réinitialiser.
            </p>
          </div>
          <div class="auth-links">
            <p>
              Déjà un compte ? <a href="#" @click.prevent="mode = 'login'">Se connecter</a>
            </p>
          </div>
        </div>
        <div v-else-if="mode === 'recovery'" key="recovery">
          <RecoveryComponent @cancel="mode = 'login'" @success="mode = 'login'" />
        </div>
      </Transition>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import LoginComponent from '../components/auth/loginComponent.vue'
import RegisterComponent from '../components/auth/registerComponent.vue'
import RecoveryComponent from '../components/auth/recoveryComponent.vue'
import { useFileStore } from '../stores/files'
import { useAuthStore } from '../stores/auth'

const mode = ref('login') // 'login', 'register', 'recovery'
const fileStore = useFileStore()
const authStore = useAuthStore()

onMounted(() => {
  // Nettoyage des données de la session précédente pour éviter les fuites de données entre comptes
  fileStore.recentFolders = []
  fileStore.recentFiles = []
  fileStore.folders = []
  fileStore.files = []
  fileStore.currentPath = '/'

  // Force la suppression du cache local pour éviter la persistance des données (fichiers suggérés) après un rafraîchissement
  localStorage.removeItem('files')
  localStorage.removeItem('file')

  // Sécurité : Suppression explicite des clés et tokens de la mémoire et du stockage
  authStore.privateKey = null
  authStore.publicKey = null
  authStore.masterKey = null
  authStore.user = null
  localStorage.removeItem('auth')
})
</script>

<style scoped>
.auth-page {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: calc(100vh - 60px); /* Adjust based on navbar height */
  background-color: var(--background-color);
}

.auth-container {
  width: 80%;
  max-width: 800px;
  padding: 2.5rem;
  border-radius: 12px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  text-align: center;
  perspective: 1000px;
}

.auth-flip-enter-active,
.auth-flip-leave-active {
  transition: all 0.4s ease-in-out;
}

.auth-flip-enter-from {
  opacity: 0;
  transform: rotateY(-90deg);
}

.auth-flip-leave-to {
  opacity: 0;
  transform: rotateY(90deg);
}

h1 {
  font-size: 1.8rem;
  margin-bottom: 2rem;
  color: var(--main-text-color);
  font-weight: 600;
}

.auth-links {
  margin-top: 1.5rem;
  font-size: 0.9rem;
  color: var(--secondary-text-color);
}

.auth-links p {
  margin: 0.5rem 0;
}

a {
  color: var(--primary-color);
  font-weight: 500;
  transition: color 0.2s;
}

a:hover {
  color: var(--accent-color);
  text-decoration: underline;
}

.security-msg {
  background-color: rgba(52, 152, 219, 0.1);
  border: 1px solid rgba(52, 152, 219, 0.3);
  color: var(--main-text-color);
  padding: 0.75rem;
  border-radius: 8px;
  font-size: 0.85rem;
  margin-top: 1.5rem;
  display: flex;
  align-items: flex-start;
  gap: 10px;
  text-align: left;
}

.security-msg.warning {
  background-color: rgba(231, 76, 60, 0.1);
  border-color: rgba(231, 76, 60, 0.3);
  color: var(--danger-color, #e74c3c);
}

.info-icon {
  flex-shrink: 0;
  margin-top: 2px;
}

.security-msg p {
  margin: 0;
}

.oauth-unavailable {
  margin-top: 1rem;
  font-size: 1rem;
  color: var(--secondary-text-color);
  opacity: 0.8;
  font-style: italic;
  padding: 0 1rem;
}
</style>