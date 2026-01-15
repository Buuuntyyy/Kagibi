<template>
  <div class="auth-page">
    <div class="auth-container card">
      <div v-if="mode === 'login'">
        <h1>Se connecter</h1>
        <LoginComponent />
        <div class="auth-links">
          <p>
            Pas encore de compte ? <a href="#" @click.prevent="mode = 'register'">S'inscrire</a>
          </p>
          <p>
            Mot de passe oublié ? <a href="#" @click.prevent="mode = 'recovery'">Récupérer mon compte</a>
          </p>
        </div>
      </div>
      <div v-else-if="mode === 'register'">
        <h1>Créer un compte</h1>
        <RegisterComponent />
        <div class="auth-links">
          <p>
            Déjà un compte ? <a href="#" @click.prevent="mode = 'login'">Se connecter</a>
          </p>
        </div>
      </div>
      <div v-else-if="mode === 'recovery'">
        <RecoveryComponent @cancel="mode = 'login'" @success="mode = 'login'" />
      </div>
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
  width: 100%;
  max-width: 400px;
  padding: 2.5rem;
  border-radius: 12px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  text-align: center;
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
</style>