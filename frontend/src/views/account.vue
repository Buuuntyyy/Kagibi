<template>
  <div class="account-container">
    <div class="header">
      <h2 class="page-title">Mon Compte</h2>
      <button @click="router.push('/dashboard')" class="btn-back">Retour au Dashboard</button>
    </div>

    <!-- Top Card: Personal Information -->
    <div class="card profile-card">
      <div class="card-header">
        <h3>Informations Personnelles</h3>
      </div>
      <div class="card-body profile-body">
         <div class="profile-avatar">
            {{ getInitials(authStore.user?.name) }}
         </div>
         <div class="profile-details">
            <div class="info-row">
              <span class="label">Nom d'utilisateur</span>
              <span class="value">{{ authStore.user?.name || 'Chargement...' }}</span>
            </div>
            <div class="info-row">
              <span class="label">Email</span>
              <span class="value">{{ authStore.user?.email || 'Chargement...' }}</span>
            </div>
            <div class="info-row">
              <span class="label">Membre depuis</span>
              <span class="value">{{ formatDate(authStore.user?.created_at) }}</span>
            </div>
         </div>
      </div>
    </div>

    <!-- Accordion Sections -->
    <div class="accordion-container">
      
      <!-- Change Username -->
      <div class="accordion-item" :class="{ 'active': activeSection === 'username' }">
        <button class="accordion-header" @click="toggleSection('username')">
          <span class="accordion-title">Modifier le nom d'utilisateur</span>
          <span class="accordion-icon">▼</span>
        </button>
        <div class="accordion-content" v-show="activeSection === 'username'">
           <div class="content-wrapper">
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
      </div>

      <!-- Change Password -->
      <div class="accordion-item" :class="{ 'active': activeSection === 'password' }">
        <button class="accordion-header" @click="toggleSection('password')">
          <span class="accordion-title">Changer le mot de passe</span>
          <span class="accordion-icon">▼</span>
        </button>
        <div class="accordion-content" v-show="activeSection === 'password'">
          <div class="content-wrapper">
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
      </div>

      <!-- Preferences -->
      <div class="accordion-item" :class="{ 'active': activeSection === 'preferences' }">
        <button class="accordion-header" @click="toggleSection('preferences')">
          <span class="accordion-title">Préférences d'Interface</span>
          <span class="accordion-icon">▼</span>
        </button>
        <div class="accordion-content" v-show="activeSection === 'preferences'">
           <div class="content-wrapper">
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
      </div>

      <!-- Legal -->
      <div class="accordion-item" :class="{ 'active': activeSection === 'legal' }">
        <button class="accordion-header" @click="toggleSection('legal')">
          <span class="accordion-title">Informations Légales</span>
          <span class="accordion-icon">▼</span>
        </button>
        <div class="accordion-content" v-show="activeSection === 'legal'">
           <div class="content-wrapper">
              <div class="legal-links">
                <router-link to="/cgu" class="legal-link">Conditions Générales d'Utilisation (CGU)</router-link>
                <router-link to="/privacy" class="legal-link">Politique de Confidentialité</router-link>
              </div>
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

const activeSection = ref(null)

const toggleSection = (section) => {
  activeSection.value = activeSection.value === section ? null : section
}

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

const getInitials = (name) => {
  if (!name) return '?'
  return name.substring(0, 2).toUpperCase()
}

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
  max-width: 800px;
  margin: 0 auto;
  padding: 2rem;
  color: var(--main-text-color);
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.page-title {
  font-size: 2rem;
  color: var(--main-text-color);
  margin: 0;
}

.btn-back {
  padding: 0.5rem 1rem;
  background-color: var(--card-color);
  border: 1px solid var(--border-color);
  color: var(--main-text-color);
  border-radius: 4px;
  cursor: pointer;
  font-weight: 500;
  transition: all 0.2s;
}

.btn-back:hover {
  background-color: var(--hover-background-color);
  border-color: var(--secondary-text-color);
}

/* Card Styling */
.card {
  background: var(--card-color);
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.05);
  overflow: hidden;
  border: 1px solid var(--border-color);
  margin-bottom: 2rem;
}

.card-header {
  background-color: var(--background-color);
  padding: 1rem 1.5rem;
  border-bottom: 1px solid var(--border-color);
}

.card-header h3 {
  margin: 0;
  font-size: 1.1rem;
  color: var(--main-text-color);
}

.profile-body {
  display: flex;
  align-items: center;
  padding: 2rem;
  gap: 2rem;
}

.profile-avatar {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background-color: var(--primary-color);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 2rem;
  font-weight: bold;
}

.profile-details {
  flex: 1;
}

.info-row {
  display: flex;
  justify-content: space-between;
  padding: 0.8rem 0;
  border-bottom: 1px solid var(--border-color);
}

.info-row:last-child {
  border-bottom: none;
}

.info-row .label {
  color: var(--secondary-text-color);
  font-weight: 500;
}

.info-row .value {
  color: var(--main-text-color);
  font-weight: 600;
}

/* Accordion Styling */
.accordion-container {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.accordion-item {
  background: var(--card-color);
  border-radius: 8px;
  border: 1px solid var(--border-color);
  overflow: hidden;
  transition: box-shadow 0.2s;
}

.accordion-item.active {
  box-shadow: 0 4px 12px rgba(0,0,0,0.05);
  border-color: var(--primary-color);
}

.accordion-header {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  background: none;
  border: none;
  cursor: pointer;
  color: var(--main-text-color);
  font-size: 1rem;
  font-weight: 600;
}

.accordion-header:hover {
  background-color: var(--hover-background-color);
}

.accordion-icon {
  font-size: 0.8rem;
  transition: transform 0.3s;
}

.accordion-item.active .accordion-icon {
  transform: rotate(180deg);
}

.accordion-content {
  border-top: 1px solid var(--border-color);
  background-color: var(--background-color);
  animation: slideDown 0.3s ease-out;
}

.content-wrapper {
  padding: 1.5rem;
}

@keyframes slideDown {
  from { opacity: 0; transform: translateY(-10px); }
  to { opacity: 1; transform: translateY(0); }
}

/* Form Shared Styles */
.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  color: var(--secondary-text-color);
  font-weight: 500;
}

.form-group input {
  width: 100%;
  padding: 0.8rem;
  background-color: var(--card-color);
  color: var(--main-text-color);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  font-size: 1rem;
  box-sizing: border-box;
}

.form-group input:focus {
  border-color: var(--primary-color);
  outline: none;
}

.btn-primary {
  width: 100%;
  padding: 0.8rem;
  background-color: var(--primary-color);
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 1rem;
  font-weight: bold;
  cursor: pointer;
  transition: background-color 0.2s;
}

.btn-primary:hover {
  background-color: var(--accent-color);
}

/* Prefs */
.pref-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 0;
  border-bottom: 1px solid var(--border-color);
}

.pref-item:last-child {
  border-bottom: none;
}

.pref-info label {
  display: block;
  font-weight: 600;
  color: var(--main-text-color);
}

.pref-desc {
  font-size: 0.85rem;
  color: var(--secondary-text-color);
}

/* Switch */
.switch {
  position: relative;
  display: inline-block;
  width: 50px;
  height: 24px;
  flex-shrink: 0;
}

.switch input { opacity: 0; width: 0; height: 0; }

.slider {
  position: absolute;
  cursor: pointer;
  top: 0; left: 0; right: 0; bottom: 0;
  background-color: var(--secondary-text-color);
  transition: .4s;
  border-radius: 34px;
}

.slider:before {
  position: absolute;
  content: "";
  height: 16px; width: 16px;
  left: 4px; bottom: 4px;
  background-color: white;
  transition: .4s;
  border-radius: 50%;
}

input:checked + .slider { background-color: var(--primary-color); }
input:checked + .slider:before { transform: translateX(26px); }

/* Legal Links */
.legal-links {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.legal-link {
  padding: 1rem;
  background-color: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  color: var(--primary-color);
  font-weight: 500;
  display: flex;
  align-items: center;
  text-decoration: none;
}

.legal-link:hover {
  background-color: var(--hover-background-color);
}

@media (max-width: 600px) {
  .profile-body {
    flex-direction: column;
    text-align: center;
  }
}
</style>
