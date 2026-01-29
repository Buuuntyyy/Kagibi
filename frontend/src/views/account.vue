<template>
  <div class="account-page">
    <div class="page-header">
      <h1>Mon Compte</h1>
      <p class="subtitle">Gérez vos informations personnelles et vos préférences.</p>
    </div>

    <!-- Plan Banner -->
    <div class="plan-banner" v-if="!loading && authStore.user">
      <div class="plan-content">
        <span class="plan-icon">🌟</span>
        <div class="plan-details">
          <span class="plan-title">Votre plan actuel</span>
          <span class="plan-value">{{ formatPlanName(authStore.user?.plan) }}</span>
        </div>
      </div>
      <button class="btn-upgrade">Mettre à niveau</button>
    </div>

    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>Chargement du profil...</p>
    </div>

    <div v-else class="content-grid">
      <!-- Left Column: User Profile -->
      <div class="user-card">
        <div class="avatar-large">
          {{ getInitials(authStore.user?.name) }}
        </div>
        <div class="user-info">
          <h2>{{ authStore.user?.name || 'Utilisateur' }}</h2>
          <p class="email">{{ authStore.user?.email || 'email@exemple.com' }}</p>
          <div class="divider"></div>
          <p class="joined-date">
             Membre depuis le {{ formatDate(authStore.user?.created_at) }}
          </p>
        </div>
      </div>

      <!-- Right Column: Settings -->
      <div class="settings-container">
        
        <!-- Account Settings -->
        <section class="settings-section">
          <div class="section-header">
            <h3>Profil</h3>
          </div>
          <div class="section-body">
            <div class="form-row">
              <div class="input-group">
                <label>
                  Nom d'utilisateur
                  <input 
                    type="text" 
                    v-model="usernameForm.newName" 
                    :placeholder="authStore.user?.name" 
                  />
                </label>
              </div>
              <button class="btn-secondary" @click="handleUpdateUsername" :disabled="updatingUsername">
                {{ updatingUsername ? 'Mise à jour...' : 'Modifier' }}
              </button>
            </div>
          </div>
        </section>

        <section class="settings-section">
          <div class="section-header">
            <h3>Sécurité</h3>
          </div>
          <div class="section-body">
            <form @submit.prevent="handleUpdatePassword" class="password-form">
              <div class="input-group password-with-toggle">
                <label>
                  Mot de passe actuel
                  <div class="password-input-wrapper">
                    <input 
                      :type="showCurrentPassword ? 'text' : 'password'" 
                      v-model="passwordForm.current" 
                      required 
                      placeholder="••••••••" 
                    />
                    <button 
                      type="button" 
                      class="toggle-password-btn" 
                      @click="showCurrentPassword = !showCurrentPassword"
                      title="Afficher/Masquer"
                    >
                      <svg v-if="!showCurrentPassword" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
                        <circle cx="12" cy="12" r="3"/>
                      </svg>
                      <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"/>
                        <line x1="1" y1="1" x2="23" y2="23"/>
                      </svg>
                    </button>
                  </div>
                </label>
              </div>
              <div class="password-row">
                 <div class="input-group password-with-toggle">
                  <label>
                    Nouveau mot de passe
                    <div class="password-input-wrapper">
                      <input 
                        :type="showNewPassword ? 'text' : 'password'" 
                        v-model="passwordForm.new" 
                        required 
                        placeholder="••••••••" 
                      />
                      <button 
                        type="button" 
                        class="toggle-password-btn" 
                        @click="showNewPassword = !showNewPassword"
                        title="Afficher/Masquer"
                      >
                        <svg v-if="!showNewPassword" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                          <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
                          <circle cx="12" cy="12" r="3"/>
                        </svg>
                        <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                          <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"/>
                          <line x1="1" y1="1" x2="23" y2="23"/>
                        </svg>
                      </button>
                    </div>
                  </label>
                </div>
                <div class="input-group password-with-toggle">
                  <label>
                    Confirmer
                    <div class="password-input-wrapper">
                      <input 
                        :type="showConfirmPassword ? 'text' : 'password'" 
                        v-model="passwordForm.confirm" 
                        required 
                        placeholder="••••••••" 
                      />
                      <button 
                        type="button" 
                        class="toggle-password-btn" 
                        @click="showConfirmPassword = !showConfirmPassword"
                        title="Afficher/Masquer"
                      >
                        <svg v-if="!showConfirmPassword" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                          <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
                          <circle cx="12" cy="12" r="3"/>
                        </svg>
                        <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                          <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"/>
                          <line x1="1" y1="1" x2="23" y2="23"/>
                        </svg>
                      </button>
                    </div>
                  </label>
                </div>
              </div>
              <div class="form-actions">
                <button type="submit" class="btn-primary" :disabled="updatingPassword">
                  {{ updatingPassword ? 'Mise à jour...' : 'Mettre à jour le mot de passe' }}
                </button>
              </div>
            </form>
          </div>
        </section>

        <section class="settings-section">
          <div class="section-header">
             <h3>Préférences</h3>
          </div>
          <div class="section-body">
             <div class="pref-list">
               <div class="pref-item">
                  <div class="pref-text">
                     <span class="pref-title">Menu Contextuel</span>
                     <span class="pref-desc">Afficher un menu d'actions au clic-droit sur les fichiers</span>
                  </div>
                  <label class="toggle-switch">
                     <input type="checkbox" v-model="preferenceStore.enableContextMenu">
                     <span class="slider"></span>
                  </label>
               </div>
               <div class="pref-item">
                  <div class="pref-text">
                     <span class="pref-title">Barre d'outils</span>
                     <span class="pref-desc">Afficher la barre d'actions au-dessus de la liste de fichiers</span>
                  </div>
                  <label class="toggle-switch">
                     <input type="checkbox" v-model="preferenceStore.showToolBar">
                     <span class="slider"></span>
                  </label>
               </div>
              <div class="pref-item">
                <div class="pref-text">
                  <span class="pref-title">Taille des dossiers</span>
                  <span class="pref-desc">Afficher la taille des dossiers (calcul immédiat, hors previews)</span>
                </div>
                <label class="toggle-switch">
                  <input type="checkbox" v-model="preferenceStore.showFolderSizes">
                  <span class="slider"></span>
                </label>
              </div>
            </div>
          </div>
        </section>

        <section class="settings-section">
           <div class="section-header">
             <h3>Informations légales</h3>
           </div>
           <div class="section-body">
             <div class="legal-links">
                <router-link to="/cgu" class="legal-link">Conditions Générales d'Utilisation</router-link>
                <router-link to="/privacy" class="legal-link">Politique de Confidentialité</router-link>
             </div>
           </div>
        </section>
      </div>
    </div>

    <!-- Error Modal -->
    <div v-if="errorModal.show" class="error-modal-overlay" @click="closeErrorModal">
      <div class="error-modal" @click.stop>
        <div class="error-modal-header">
          <svg viewBox="0 0 24 24" width="24" height="24" fill="currentColor" class="error-icon">
            <circle cx="12" cy="12" r="10"/>
            <text x="12" y="16" text-anchor="middle" fill="white" font-size="12" font-weight="bold">!</text>
          </svg>
          <h3>{{ errorModal.title }}</h3>
          <button class="btn-close-modal" @click="closeErrorModal">×</button>
        </div>
        <div class="error-modal-body">
          <p>{{ errorModal.message }}</p>
        </div>
        <div class="error-modal-footer">
          <button class="btn-primary" @click="closeErrorModal">Fermer</button>
        </div>
      </div>
    </div>

    <!-- Success Modal -->
    <div v-if="successModal.show" class="success-modal-overlay" @click="closeSuccessModal">
      <div class="success-modal" @click.stop>
        <div class="success-modal-header">
          <svg viewBox="0 0 24 24" width="24" height="24" fill="currentColor" class="success-icon">
            <circle cx="12" cy="12" r="10"/>
            <path d="M10 14.5l3 3 5-5.5" stroke="white" stroke-width="2" fill="none"/>
          </svg>
          <h3>{{ successModal.title }}</h3>
          <button class="btn-close-modal" @click="closeSuccessModal">×</button>
        </div>
        <div class="success-modal-body">
          <p>{{ successModal.message }}</p>
        </div>
        <div class="success-modal-footer">
          <button class="btn-primary" @click="closeSuccessModal">Fermer</button>
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
const loading = ref(true)

const usernameForm = ref({
  newName: ''
})

const passwordForm = ref({
  current: '',
  new: '',
  confirm: ''
})

const showCurrentPassword = ref(false)
const showNewPassword = ref(false)
const showConfirmPassword = ref(false)

const updatingUsername = ref(false)
const updatingPassword = ref(false)

const errorModal = ref({
  show: false,
  title: '',
  message: ''
})

const successModal = ref({
  show: false,
  title: '',
  message: ''
})

const closeErrorModal = () => {
  errorModal.value.show = false
}

const closeSuccessModal = () => {
  successModal.value.show = false
}

const showError = (title, message) => {
  errorModal.value = { show: true, title, message }
}

const showSuccess = (title, message) => {
  successModal.value = { show: true, title, message }
}

onMounted(async () => {
  try {
    await authStore.fetchUser()
    if (authStore.user) {
      usernameForm.value.newName = authStore.user.name
    }
  } catch (e) {
    console.error("Error loading profile", e)
  } finally {
    loading.value = false
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

const formatPlanName = (plan) => {
  const plans = {
    'free': 'Gratuit',
    'basic': 'Basique',
    'pro': 'Professionnel',
    'enterprise': 'Entreprise'
  }
  return plans[plan] || plan || 'Gratuit'
}

const handleUpdateUsername = async () => {
  if (!usernameForm.value.newName.trim()) {
    showError('Erreur', 'Le nom d\'utilisateur ne peut pas être vide.')
    return
  }

  if (usernameForm.value.newName === authStore.user?.name) {
    showError('Erreur', 'Veuillez entrer un nouveau nom différent du nom actuel.')
    return
  }

  updatingUsername.value = true
  try {
    await authStore.updateUsername(usernameForm.value.newName)
    showSuccess('Succès', 'Votre nom d\'utilisateur a été mis à jour avec succès.')
    // Update local user
    authStore.user.name = usernameForm.value.newName
  } catch (error) {
    console.error("Failed to update username:", error)
    const errorMessage = error.response?.data?.error || error.message || 'Erreur lors de la mise à jour du nom d\'utilisateur.'
    showError('Erreur', errorMessage)
  } finally {
    updatingUsername.value = false
  }
}

const handleUpdatePassword = async () => {
  if (passwordForm.value.new !== passwordForm.value.confirm) {
    showError('Erreur', 'Les nouveaux mots de passe ne correspondent pas.')
    return
  }
  
  if (passwordForm.value.new.length < 20) {
    showError('Erreur', 'Le nouveau mot de passe doit contenir au moins 20 caractères.')
    return
  }
  
  updatingPassword.value = true
  try {
    await authStore.updatePassword(passwordForm.value.current, passwordForm.value.new)
    showSuccess('Succès', 'Votre mot de passe a été mis à jour avec succès !')
    
    // Reset the form
    passwordForm.value.current = ''
    passwordForm.value.new = ''
    passwordForm.value.confirm = ''
    showCurrentPassword.value = false
    showNewPassword.value = false
    showConfirmPassword.value = false
  } catch (error) {
    console.error("Failed to update password:", error)
    
    let errorMessage = 'Erreur lors de la mise à jour du mot de passe.'
    
    if (error.response) {
      if (error.response.status === 401) {
        errorMessage = 'Mot de passe actuel incorrect.'
      } else if (error.response.data && error.response.data.error) {
        errorMessage = error.response.data.error
      }
    } else if (error.message) {
      errorMessage = error.message
    }
    
    showError('Erreur', errorMessage)
  } finally {
    updatingPassword.value = false
  }
}
</script>

<style scoped>
.account-page {
  padding: 2rem;
  background-color: var(--background-color);
  height: 100%;
  overflow-y: auto;
  box-sizing: border-box;
}

.page-header {
  margin-bottom: 2rem;
}

.page-header h1 {
  font-size: 2rem;
  margin: 0;
  color: var(--main-text-color);
}

.subtitle {
  color: var(--secondary-text-color);
  margin-top: 0.5rem;
}

.plan-banner {
  background: linear-gradient(135deg, var(--primary-color), var(--accent-color));
  border-radius: 12px;
  padding: 1.5rem;
  color: white;
  margin-bottom: 2rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  box-shadow: 0 4px 15px rgba(0,0,0,0.1);
}

.plan-content {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.plan-icon {
  font-size: 2rem;
  background: rgba(255,255,255,0.2);
  width: 50px;
  height: 50px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
}

.plan-details {
  display: flex;
  flex-direction: column;
}

.plan-title {
  font-size: 0.9rem;
  opacity: 0.9;
}

.plan-value {
  font-size: 1.5rem;
  font-weight: bold;
  text-transform: capitalize;
}

.btn-upgrade {
  background: white;
  color: var(--primary-color);
  border: none;
  padding: 0.8rem 1.5rem;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  transition: transform 0.2s;
}

.btn-upgrade:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 8px rgba(0,0,0,0.2);
}

@media (max-width: 600px) {
  .plan-banner {
    flex-direction: column;
    gap: 1rem;
    text-align: center;
  }
  
  .plan-content {
    flex-direction: column;
  }

  .btn-upgrade {
    width: 100%;
  }
}

.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 200px;
  color: var(--secondary-text-color);
}

.content-grid {
  display: grid;
  grid-template-columns: 300px 1fr;
  gap: 2rem;
  align-items: start;
}

@media (max-width: 900px) {
  .content-grid {
    grid-template-columns: 1fr;
  }
}

/* User Card */
.user-card {
  background: var(--card-color);
  padding: 2rem;
  border-radius: 12px;
  border: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  position: sticky;
  top: 2rem;
}

.avatar-large {
  width: 100px;
  height: 100px;
  border-radius: 50%;
  background-color: var(--primary-color);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 2.5rem;
  font-weight: bold;
  margin-bottom: 1.5rem;
  box-shadow: 0 4px 10px rgba(0,0,0,0.1);
}

.user-info h2 {
  margin: 0;
  font-size: 1.5rem;
  color: var(--main-text-color);
}

.user-info .email {
  color: var(--secondary-text-color);
  margin: 0.5rem 0 1.5rem 0;
}

.divider {
  height: 1px;
  background-color: var(--border-color);
  width: 100%;
  margin: 1.5rem 0;
}

.user-info .joined-date {
  font-size: 0.9rem;
  color: var(--secondary-text-color);
}

/* Settings Sections */
.settings-container {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.settings-section {
  background: var(--card-color);
  border-radius: 12px;
  border: 1px solid var(--border-color);
  overflow: hidden;
}

.section-header {
  padding: 1.5rem;
  border-bottom: 1px solid var(--border-color);
}

.section-header h3 {
  margin: 0;
  font-size: 1.2rem;
  color: var(--main-text-color);
}

.section-body {
  padding: 1.5rem;
}

/* Forms */
.form-row {
  display: flex;
  gap: 1rem;
  align-items: flex-end;
}

.password-row {
  display: flex;
  gap: 1rem;
}

.password-row .input-group {
  flex: 1;
}

.input-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  flex: 1;
  margin-bottom: 1rem;
}

.input-group label {
  font-size: 0.9rem;
  font-weight: 500;
  color: var(--secondary-text-color);
}

.password-input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

input {
  padding: 0.8rem 1rem;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--background-color);
  color: var(--main-text-color);
  font-size: 1rem;
  transition: border-color 0.2s;
}

input:focus {
  outline: none;
  border-color: var(--primary-color);
}

.password-input-wrapper input {
  flex: 1;
  padding-right: 40px;
}

.toggle-password-btn {
  position: absolute;
  right: 8px;
  background: none;
  border: none;
  cursor: pointer;
  color: var(--secondary-text-color);
  padding: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  transition: all 0.2s;
}

.toggle-password-btn:hover {
  color: var(--primary-color);
  background: rgba(99, 102, 241, 0.1);
}

.toggle-password-btn svg {
  width: 18px;
  height: 18px;
}

.btn-primary {
  background-color: var(--primary-color);
  color: white;
  border: none;
  padding: 0.8rem 1.5rem;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 600;
  transition: background-color 0.2s;
}

.btn-primary:hover {
  background-color: var(--accent-color);
}

.btn-secondary {
  background-color: transparent;
  border: 1px solid var(--border-color);
  color: var(--main-text-color);
  padding: 0.8rem 1.5rem;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 500;
  margin-bottom: 1rem; /* alignment fix for form-row */
}

.btn-secondary:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 1rem;
}

/* Preferences */
.pref-list {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.pref-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.pref-text {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.pref-title {
  font-weight: 600;
  color: var(--main-text-color);
}

.pref-desc {
  font-size: 0.9rem;
  color: var(--secondary-text-color);
}

/* Toggle Switch */
.toggle-switch {
  position: relative;
  display: inline-block;
  width: 50px;
  height: 26px;
  flex-shrink: 0;
}

.toggle-switch input { 
  opacity: 0; 
  width: 0; 
  height: 0; 
}

.slider {
  position: absolute;
  cursor: pointer;
  top: 0; left: 0; right: 0; bottom: 0;
  background-color: var(--border-color);
  transition: .4s;
  border-radius: 34px;
}

.slider:before {
  position: absolute;
  content: "";
  height: 18px; width: 18px;
  left: 4px; bottom: 4px;
  background-color: white;
  transition: .4s;
  border-radius: 50%;
}

input:checked + .slider {
  background-color: var(--success-color);
}

input:checked + .slider:before {
  transform: translateX(24px);
}

/* Legal Links */
.legal-links {
  display: flex;
  flex-direction: column;
  gap: 0.8rem;
}

.legal-link {
  color: var(--primary-color);
  text-decoration: none;
  font-weight: 500;
}

.legal-link:hover {
  text-decoration: underline;
}

/* Spinner */
.spinner {
  width: 40px;
  height: 40px;
  border: 4px solid var(--border-color);
  border-top-color: var(--primary-color);
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 1rem;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Error & Success Modals */
.error-modal-overlay, .success-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  animation: fadeIn 0.2s ease;
}

.error-modal, .success-modal {
  background: var(--card-color);
  border-radius: 12px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.15);
  max-width: 400px;
  width: 90%;
  animation: slideUp 0.3s ease;
  border: 1px solid var(--border-color);
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

.error-modal-header, .success-modal-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 24px 24px 16px;
  border-bottom: 1px solid var(--border-color);
}

.error-modal-header h3, .success-modal-header h3 {
  margin: 0;
  flex: 1;
  font-size: 1.1rem;
  color: var(--main-text-color);
}

.error-icon {
  color: var(--error-color);
  flex-shrink: 0;
}

.success-icon {
  color: var(--success-color);
  flex-shrink: 0;
}

.btn-close-modal {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: var(--secondary-text-color);
  padding: 0;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: color 0.2s;
}

.btn-close-modal:hover {
  color: var(--main-text-color);
}

.error-modal-body, .success-modal-body {
  padding: 20px 24px;
  color: var(--main-text-color);
  line-height: 1.5;
}

.error-modal-body p, .success-modal-body p {
  margin: 0;
}

.error-modal-footer, .success-modal-footer {
  padding: 16px 24px;
  border-top: 1px solid var(--border-color);
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
