<template>
  <div class="account-container">
    <div class="header">
      <h2 class="page-title">Mon Compte</h2>
      <button @click="router.push('/dashboard')" class="btn-back">Retour au Dashboard</button>
    </div>

    <div class="content-grid">
      <!-- Profile Info Card -->
      <div class="card profile-card">
        <div class="card-header">
          <h3>Informations Personnelles</h3>
        </div>
        <div class="card-body">
          <div class="info-group">
            <label>Nom d'utilisateur</label>
            <div class="info-value">{{ authStore.user?.name || 'Chargement...' }}</div>
          </div>
          <div class="info-group">
            <label>Email</label>
            <div class="info-value">{{ authStore.user?.email || 'Chargement...' }}</div>
          </div>
          <div class="info-group">
            <label>Membre depuis</label>
            <div class="info-value">{{ formatDate(authStore.user?.created_at) }}</div>
          </div>
        </div>
      </div>

      <!-- Change Username Card -->
      <div class="card">
        <div class="card-header">
          <h3>Modifier le nom d'utilisateur</h3>
        </div>
        <div class="card-body">
          <form @submit.prevent="handleUpdateUsername">
            <div class="form-group">
              <label for="newUsername">Nouveau nom d'utilisateur</label>
              <input 
                type="text" 
                id="newUsername" 
                v-model="usernameForm.newName" 
                placeholder="Entrez votre nouveau nom"
                required
              />
            </div>
            <button type="submit" class="btn-primary">Sauvegarder</button>
          </form>
        </div>
      </div>

      <!-- Change Password Card -->
      <div class="card">
        <div class="card-header">
          <h3>Changer le mot de passe</h3>
        </div>
        <div class="card-body">
          <form @submit.prevent="handleUpdatePassword">
            <div class="form-group">
              <label for="currentPassword">Mot de passe actuel</label>
              <input 
                type="password" 
                id="currentPassword" 
                v-model="passwordForm.current" 
                required
              />
            </div>
            <div class="form-group">
              <label for="newPassword">Nouveau mot de passe</label>
              <input 
                type="password" 
                id="newPassword" 
                v-model="passwordForm.new" 
                required
              />
            </div>
            <div class="form-group">
              <label for="confirmPassword">Confirmer le nouveau mot de passe</label>
              <input 
                type="password" 
                id="confirmPassword" 
                v-model="passwordForm.confirm" 
                required
              />
            </div>
            <button type="submit" class="btn-primary">Mettre à jour le mot de passe</button>
          </form>
        </div>
      </div>

      <div class="card profile-card"> <!-- J'utilise profile-card pour qu'elle prenne toute la largeur -->
        <div class="card-header">
          <h3>Préférences d'Interface</h3>
        </div>
        <div class="card-body">
          <div class="prefs-grid">
            
            <div class="pref-item">
              <div class="pref-info">
                <label class="pref-label">Menu Contextuel (Clic-droit)</label>
                <span class="pref-desc">Affiche un menu d'actions rapides au clic-droit sur un fichier.</span>
              </div>
              <label class="switch">
                <input type="checkbox" v-model="preferenceStore.enableContextMenu">
                <span class="slider round"></span>
              </label>
            </div>

            <div class="pref-item">
              <div class="pref-info">
                <label class="pref-label">Barre d'actions</label>
                <span class="pref-desc">Affiche la barre d'outils (boutons Renommer, Supprimer...) au dessus de la liste.</span>
              </div>
              <label class="switch">
                <input type="checkbox" v-model="preferenceStore.showToolBar">
                <span class="slider round"></span>
              </label>
            </div>

          </div>
        </div>
      </div>

      <!-- Legal Information Card -->
      <div class="card profile-card">
        <div class="card-header">
          <h3>Informations Légales</h3>
        </div>
        <div class="card-body">
          <div class="legal-links">
            <router-link to="/cgu" class="legal-link">Conditions Générales d'Utilisation (CGU)</router-link>
            <router-link to="/privacy" class="legal-link">Politique de Confidentialité</router-link>
          </div>
        </div>
      </div>

    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { usePreferencesStore } from '../stores/preferences'

const router = useRouter()
const authStore = useAuthStore()
const preferenceStore = usePreferencesStore()

const usernameForm = ref({
  newName: ''
})

const passwordForm = ref({
  current: '',
  new: '',
  confirm: ''
})

onMounted(async () => {
  if (!authStore.user) {
    await authStore.fetchUser()
  }
  // Pre-fill username if available
  if (authStore.user) {
    usernameForm.value.newName = authStore.user.name
  }
})

const formatDate = (dateString) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleDateString('fr-FR', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })
}

const handleUpdateUsername = async () => {
  // TODO: Implement backend call
  alert("La fonctionnalité de changement de nom d'utilisateur sera implémentée prochainement.")
  console.log('Update username to:', usernameForm.value.newName)
}

const handleUpdatePassword = async () => {
  if (passwordForm.value.new !== passwordForm.value.confirm) {
    alert("Les nouveaux mots de passe ne correspondent pas.")
    return
  }
  
  try {
    await authStore.updatePassword(passwordForm.value.current, passwordForm.value.new)
    alert("Mot de passe mis à jour avec succès !")
    
    // Reset sensitive fields
    passwordForm.value.current = ''
    passwordForm.value.new = ''
    passwordForm.value.confirm = ''
  } catch (error) {
    console.error("Failed to update password:", error)
    if (error.response && error.response.data && error.response.data.error) {
        alert("Erreur: " + error.response.data.error)
    } else {
        alert("Erreur lors de la mise à jour du mot de passe.")
    }
  }
}
</script>

<style scoped>
.account-container {
  max-width: 1000px;
  margin: 0 auto;
  padding: 2rem;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.page-title {
  font-size: 2rem;
  color: #333;
  margin: 0;
}

.btn-back {
  padding: 0.5rem 1rem;
  background-color: #f0f0f0;
  border: 1px solid #ccc;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 500;
  transition: background-color 0.2s;
}

.btn-back:hover {
  background-color: #e0e0e0;
}

.content-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 2rem;
}

@media (min-width: 768px) {
  .content-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  
  .profile-card {
    grid-column: span 2;
  }
}

.card {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  overflow: hidden;
  border: 1px solid #eee;
}

.card-header {
  background-color: #f8f9fa;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid #eee;
}

.card-header h3 {
  margin: 0;
  font-size: 1.1rem;
  color: #444;
}

.card-body {
  padding: 1.5rem;
}

.info-group {
  margin-bottom: 1rem;
}

.info-group label {
  display: block;
  font-size: 0.85rem;
  color: #666;
  margin-bottom: 0.25rem;
}

.info-value {
  font-size: 1.1rem;
  color: #333;
  font-weight: 500;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  color: #555;
  font-weight: 500;
}

.form-group input {
  width: 100%;
  padding: 0.6rem;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 1rem;
  box-sizing: border-box; /* Important for padding not to affect width */
}

.form-group input:focus {
  border-color: var(--primary-color, #42b983);
  outline: none;
  box-shadow: 0 0 0 2px rgba(66, 185, 131, 0.2);
}

.btn-primary {
  width: 100%;
  padding: 0.75rem;
  background-color: var(--primary-color, #42b983);
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 1rem;
  font-weight: bold;
  cursor: pointer;
  transition: background-color 0.2s;
}

.btn-primary:hover {
  background-color: #3aa876;
}

.prefs-grid {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.pref-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 1rem;
  border-bottom: 1px solid #f0f0f0;
}

.pref-item:last-child {
  border-bottom: none;
  padding-bottom: 0;
}

.pref-info {
  display: flex;
  flex-direction: column;
}

.pref-label {
  font-weight: 500;
  color: #333;
  margin-bottom: 0.25rem;
}

.pref-desc {
  font-size: 0.85rem;
  color: #888;
}

.switch {
  position: relative;
  display: inline-block;
  width: 50px;
  height: 24px;
  flex-shrink: 0;
}

.switch input { 
  opacity: 0;
  width: 0;
  height: 0;
}

.slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: #ccc;
  transition: .4s;
}

.slider:before {
  position: absolute;
  content: "";
  height: 16px;
  width: 16px;
  left: 4px;
  bottom: 4px;
  background-color: white;
  transition: .4s;
}

input:checked + .slider {
  background-color: var(--primary-color, #42b983);
}

input:focus + .slider {
  box-shadow: 0 0 1px var(--primary-color, #42b983);
}

input:checked + .slider:before {
  transform: translateX(26px);
}

.slider.round {
  border-radius: 34px;
}

.slider.round:before {
  border-radius: 50%;
}

.legal-links {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.legal-link {
  color: var(--primary-color, #42b983);
  text-decoration: none;
  font-weight: 500;
  padding: 0.5rem;
  border-radius: 4px;
  transition: background-color 0.2s;
  display: flex;
  align-items: center;
}

.legal-link:hover {
  background-color: #f0fdf4;
  text-decoration: underline;
}

.legal-link::before {
  content: "📄";
  margin-right: 10px;
}
</style>
