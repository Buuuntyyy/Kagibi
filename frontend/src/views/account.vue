<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="account-page">
    <div class="page-header">
      <div class="header-content">
        <button class="btn-back" @click="router.go(-1)">
          <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M19 12H5M12 19l-7-7 7-7"/>
          </svg>
          Retour
        </button>
        <h1>{{ t('account.title') }}</h1>
      </div>
      <p class="subtitle">{{ t('account.subtitle') }}</p>
    </div>

    <!-- Plan Banner -->
    <div class="plan-banner" v-if="!loading && authStore.user && billingStore.showSubscriptionUI">
      <div class="plan-content">
        <span class="plan-icon">🌟</span>
        <div class="plan-details">
          <span class="plan-title">{{ t('account.currentPlan') }}</span>
          <span class="plan-value">{{ formatPlanName(authStore.user?.plan) }}</span>
        </div>
      </div>
      <button class="btn-upgrade" @click="openUpgradeInfoPopup">{{ t('account.upgrade') }}</button>
    </div>

    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>{{ t('account.loading') }}</p>
    </div>

    <div v-else class="content-grid">
      <!-- Left Column: User Profile -->
      <div class="user-card">
        <div class="avatar-container">
          <AvatarSelector v-model="selectedAvatar" />
        </div>
        <div class="user-info">
          <h2>{{ authStore.user?.name || t('account.username') }}</h2>
          <p class="email">{{ authStore.user?.email || 'email@exemple.com' }}</p>
          <div class="divider"></div>
          <p class="joined-date">
             {{ t('account.memberSince') }} {{ formatDate(authStore.user?.created_at) }}
          </p>
        </div>
      </div>

      <!-- Right Column: Settings -->
      <div class="settings-container">

        <!-- Account Settings -->
        <section class="settings-section">
          <div class="section-header">
            <h3>{{ t('account.profile') }}</h3>
          </div>
          <div class="section-body">
            <div class="form-row">
              <div class="input-group">
                <label>
                  {{ t('account.username') }}
                  <input
                    type="text"
                    v-model="usernameForm.newName"
                    :placeholder="authStore.user?.name"
                  />
                </label>
              </div>
              <button class="btn-secondary" @click="handleUpdateUsername" :disabled="updatingUsername">
                {{ updatingUsername ? t('account.updating') : t('account.modify') }}
              </button>
            </div>
            <div class="form-divider"></div>
            <div class="form-row">
              <div class="input-group">
                <label>
                  {{ t('account.newEmail') }}
                  <input
                    type="email"
                    v-model="emailForm.newEmail"
                    :placeholder="authStore.user?.email"
                  />
                </label>
              </div>
              <div class="input-group">
                <label>
                  {{ t('account.currentPasswordForEmail') }}
                  <div class="password-input-wrapper">
                    <input
                      :type="showEmailPassword ? 'text' : 'password'"
                      v-model="emailForm.password"
                      placeholder="••••••••"
                    />
                    <button
                      type="button"
                      class="toggle-password-btn"
                      @click="showEmailPassword = !showEmailPassword"
                      :title="t('account.showHide')"
                    >
                      <svg v-if="!showEmailPassword" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
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
              <button class="btn-secondary" @click="handleUpdateEmail" :disabled="updatingEmail">
                {{ updatingEmail ? t('account.updatingEmail') : t('account.updateEmail') }}
              </button>
            </div>
          </div>
        </section>

        <section class="settings-section">
          <div class="section-header">
            <h3>{{ t('account.security') }}</h3>
          </div>
          <div class="section-body">
            <form @submit.prevent="handleUpdatePassword" class="password-form">
              <div class="input-group password-with-toggle">
                <label>
                  {{ t('account.currentPassword') }}
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
                      :title="t('account.showHide')"
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
              <div class="password-stack">
                <div class="input-group password-with-toggle">
                  <label>
                    {{ t('account.newPassword') }}
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
                        :title="t('account.showHide')"
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
                <PasswordCriteria :password="passwordForm.new" />
                <div class="input-group password-with-toggle">
                  <label>
                    {{ t('account.confirmPassword') }}
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
                        :title="t('account.showHide')"
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
                  {{ updatingPassword ? t('account.updatingPassword') : t('account.updatePassword') }}
                </button>
              </div>
            </form>
          </div>
        </section>

        <!-- MFA Settings -->
        <section class="settings-section">
          <div class="section-header">
            <h3>{{ t('account.mfaTitle') }}</h3>
          </div>
          <div class="section-body">
            <MFASettings />
          </div>
        </section>

        <section class="settings-section">
          <div class="section-header">
             <h3>{{ t('account.preferences') }}</h3>
          </div>
          <div class="section-body">
             <div class="pref-list">
               <div class="pref-item">
                  <div class="pref-text">
                     <span class="pref-title">{{ t('account.contextMenu') }}</span>
                     <span class="pref-desc">{{ t('account.contextMenuDesc') }}</span>
                  </div>
                  <label class="toggle-switch">
                     <input type="checkbox" v-model="preferenceStore.enableContextMenu">
                     <span class="slider"></span>
                  </label>
               </div>
               <div class="pref-item">
                  <div class="pref-text">
                     <span class="pref-title">{{ t('account.toolbar') }}</span>
                     <span class="pref-desc">{{ t('account.toolbarDesc') }}</span>
                  </div>
                  <label class="toggle-switch">
                     <input type="checkbox" v-model="preferenceStore.showToolBar">
                     <span class="slider"></span>
                  </label>
               </div>
              <div class="pref-item">
                <div class="pref-text">
                  <span class="pref-title">{{ t('account.folderSizes') }}</span>
                  <span class="pref-desc">{{ t('account.folderSizesDesc') }}</span>
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
             <h3>{{ t('account.legalInfo') }}</h3>
           </div>
           <div class="section-body">
             <div class="legal-links">
                <router-link to="/cgu" class="legal-link">{{ t('account.termsOfService') }}</router-link>
                <router-link to="/privacy" class="legal-link">{{ t('account.privacyPolicy') }}</router-link>
                <router-link to="/credits" class="legal-link">{{ t('account.legalCredits') }}</router-link>
             </div>
           </div>
        </section>

        <!-- Portabilité - RGPD Article 20 -->
        <section class="settings-section">
          <div class="section-header">
            <h3>{{ t('account.dataPortability') }}</h3>
          </div>
          <div class="section-body">
            <div class="portability-item">
              <div class="portability-info">
                <p class="portability-desc">
                  {{ t('account.dataPortabilityDesc') }}
                </p>
                <p class="portability-details">
                  {{ t('account.dataPortabilityDetails') }}
                </p>
              </div>
              <div class="portability-actions">
                <button
                  class="btn-primary"
                  @click="handleExportData"
                  :disabled="exportingData"
                >
                  <svg v-if="!exportingData" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 0.5rem; vertical-align: middle;">
                    <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
                    <polyline points="7 10 12 15 17 10"/>
                    <line x1="12" y1="15" x2="12" y2="3"/>
                  </svg>
                  {{ exportingData ? t('account.exportingData') : t('account.exportData') }}
                </button>
              </div>
            </div>
          </div>
        </section>

        <!-- Danger Zone - RGPD Article 17 -->
        <section class="settings-section danger-zone">
          <div class="section-header">
            <h3>{{ t('account.dangerZone') }}</h3>
          </div>
          <div class="section-body danger-zone-body">
            <div class="danger-zone-item">
              <div class="danger-zone-info">
                <h4>{{ t('account.deleteAccount') }}</h4>
                <p>{{ t('account.deleteAccountDesc') }}</p>
              </div>
              <button @click="showDeleteModal = true" class="btn-danger-outline">
                {{ t('account.deleteAccount') }}
              </button>
            </div>
          </div>
        </section>
      </div>
    </div>

    <DeleteAccountDialog
      v-model="showDeleteModal"
      v-model:deleteConfirmationText="deleteConfirmationText"
      :isDeletingAccount="isDeletingAccount"
      @confirm="handleDeleteAccount"
      @close="closeDeleteModal"
    />

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
          <button class="btn-primary" @click="closeErrorModal">{{ t('account.close') }}</button>
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
          <button class="btn-primary" @click="closeSuccessModal">{{ t('account.close') }}</button>
        </div>
      </div>
    </div>

    <!-- MFA Challenge Modal -->
    <MFAChallengeModal
      v-model="showMFAChallenge"
      :context="mfaChallengeContext"
      @verified="onMFAVerified"
      @cancelled="onMFACancelled"
    />

    <div v-if="upgradeInfoModal.show" class="success-modal-overlay" @click="closeUpgradeInfoPopup">
      <div class="success-modal" @click.stop>
        <div class="success-modal-header">
          <svg viewBox="0 0 24 24" width="24" height="24" fill="currentColor" class="success-icon">
            <circle cx="12" cy="12" r="10"/>
            <text x="12" y="16" text-anchor="middle" fill="white" font-size="12" font-weight="bold">i</text>
          </svg>
          <h3>Abonnements bientôt disponibles</h3>
          <button class="btn-close-modal" @click="closeUpgradeInfoPopup">×</button>
        </div>
        <div class="success-modal-body">
          <p>
            Les abonnements ne sont pas encore disponibles. En attendant, vous pouvez soutenir le projet via Buy me a coffee.
          </p>
        </div>
        <div class="success-modal-footer">
          <a
            v-if="buyMeACoffeeUrl"
            :href="buyMeACoffeeUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="btn-primary"
          >
            ☕ Buy me a coffee
          </a>
          <button class="btn-secondary" @click="closeUpgradeInfoPopup">{{ t('account.close') }}</button>
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

    <!-- MFA Challenge Modal -->
    <MFAChallengeModal
      v-model="showMFAChallenge"
      :context="mfaChallengeContext"
      @verified="onMFAVerified"
      @cancelled="onMFACancelled"
    />
  </div>
</template>

<script setup>
import { ref, onMounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../stores/auth'
import { useBillingStore } from '../stores/billing'
import { usePreferencesStore } from '../stores/preferences'
import { useMFA } from '../utils/useMFA'
import api from '../api'
import AvatarSelector from '../components/AvatarSelector.vue'
import DeleteAccountDialog from '../components/DeleteAccountDialog.vue'
import MFASettings from '../components/MFASettings.vue'
import MFAChallengeModal from '../components/MFAChallengeModal.vue'
import PasswordCriteria from '../components/auth/PasswordCriteria.vue'
import { checkPasswordCriteria, getPasswordErrors } from '../utils/passwordStrength'

const { t } = useI18n()

const router = useRouter()
const authStore = useAuthStore()
const billingStore = useBillingStore()
const preferenceStore = usePreferencesStore()
const { isMFARequired } = useMFA()

// MFA Challenge state
const showMFAChallenge = ref(false)
const mfaChallengeContext = ref('destructive')
const pendingAction = ref(null) // Will store the action to execute after MFA verification

const loading = ref(true)
const selectedAvatar = ref('/avatars/default.png')
const updatingAvatar = ref(false)

const usernameForm = ref({
  newName: ''
})

const emailForm = ref({
  newEmail: '',
  password: ''
})

const passwordForm = ref({
  current: '',
  new: '',
  confirm: ''
})

const showCurrentPassword = ref(false)
const showNewPassword = ref(false)
const showConfirmPassword = ref(false)
const showEmailPassword = ref(false)

const updatingUsername = ref(false)
const updatingEmail = ref(false)
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

const upgradeInfoModal = ref({
  show: false,
})

const showDeleteModal = ref(false)
const deleteConfirmationText = ref('')
const isDeletingAccount = ref(false)
const exportingData = ref(false)

const closeErrorModal = () => {
  errorModal.value.show = false
}

const closeSuccessModal = () => {
  successModal.value.show = false
}

const closeDeleteModal = () => {
  if (!isDeletingAccount.value) {
    showDeleteModal.value = false
    deleteConfirmationText.value = ''
  }
}

const showError = (title, message) => {
  errorModal.value = { show: true, title, message }
}

const showSuccess = (title, message) => {
  successModal.value = { show: true, title, message }
}

const buyMeACoffeeUrl = computed(() => {
  const runtimeUrl = typeof window !== 'undefined' ? window.__APP_CONFIG__?.buyMeACoffeeUrl : ''
  return runtimeUrl || import.meta.env.VITE_BUY_ME_A_COFFEE_URL || ''
})

const openUpgradeInfoPopup = () => {
  upgradeInfoModal.value.show = true
}

const closeUpgradeInfoPopup = () => {
  upgradeInfoModal.value.show = false
}

const onMFAVerified = async () => {
  showMFAChallenge.value = false
  if (pendingAction.value) {
    const action = pendingAction.value
    pendingAction.value = null
    try {
      await action()
    } catch (error) {
      console.error('Error executing pending action after MFA:', error)
      const errorMessage = error.response?.data?.error || error.message || 'Erreur lors de l\'exécution de l\'action.'
      showError('Erreur', errorMessage)
    }
  }
}

const onMFACancelled = () => {
  showMFAChallenge.value = false
  pendingAction.value = null
  showError('Action annulée', 'La vérification MFA a été annulée. Votre action n\'a pas été exécutée.')
}

onMounted(async () => {
  try {
    await authStore.fetchUser()
    if (authStore.user) {
      usernameForm.value.newName = authStore.user.name
      selectedAvatar.value = authStore.user.avatar_url || '/avatars/default.png'
    }
  } catch (e) {
    console.error("Error loading profile", e)
  } finally {
    loading.value = false
  }
})

// Watch for user changes to update selectedAvatar
watch(() => authStore.user?.avatar_url, (newAvatar) => {
  if (newAvatar) {
    selectedAvatar.value = newAvatar
  }
})

// Auto-save avatar when selection changes
watch(selectedAvatar, async (newAvatar, oldAvatar) => {
  // Only save if avatar actually changed and it's different from current user avatar
  if (newAvatar && oldAvatar && newAvatar !== oldAvatar && authStore.user?.avatar_url !== newAvatar) {
    try {
      updatingAvatar.value = true
      await authStore.updateAvatar(newAvatar)
      showSuccess('Succès', 'Votre avatar a été mis à jour avec succès !')
    } catch (error) {
      console.error('Failed to update avatar:', error)
      const errorMessage = error.response?.data?.error || error.message || 'Erreur lors de la mise à jour de l\'avatar.'
      showError('Erreur', errorMessage)
      // Revert to previous avatar on error
      selectedAvatar.value = oldAvatar
    } finally {
      updatingAvatar.value = false
    }
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

const executeEmailUpdate = async () => {
  const newEmail = emailForm.value.newEmail.trim()
  updatingEmail.value = true
  try {
    await authStore.updateEmail(newEmail, emailForm.value.password)
    showSuccess('Succès', 'Votre adresse email a été mise à jour avec succès.')
    emailForm.value.newEmail = ''
    emailForm.value.password = ''
    showEmailPassword.value = false
  } catch (error) {
    console.error('Failed to update email:', error)
    const errorMessage = error.message || 'Erreur lors de la mise à jour de l\'email.'
    showError('Erreur', errorMessage)
  } finally {
    updatingEmail.value = false
  }
}

const handleUpdateEmail = async () => {
  const newEmail = emailForm.value.newEmail.trim()
  if (!newEmail) {
    showError('Erreur', 'Veuillez entrer une nouvelle adresse email.')
    return
  }

  if (newEmail === authStore.user?.email) {
    showError('Erreur', 'Veuillez entrer une adresse email différente de l\'adresse actuelle.')
    return
  }

  if (!emailForm.value.password) {
    showError('Erreur', 'Veuillez entrer votre mot de passe pour confirmer.')
    return
  }

  // Check if MFA is required for email change
  try {
    const mfaRequired = await isMFARequired('email_change')
    if (mfaRequired) {
      pendingAction.value = async () => { await executeEmailUpdate() }
      mfaChallengeContext.value = 'email_change'
      showMFAChallenge.value = true
      return
    }
  } catch (err) {
    console.error('Error checking MFA requirement:', err)
  }

  await executeEmailUpdate()
}

const handleUpdatePassword = async () => {
  if (passwordForm.value.new !== passwordForm.value.confirm) {
    showError('Erreur', 'Les nouveaux mots de passe ne correspondent pas.')
    return
  }

  const { valid } = checkPasswordCriteria(passwordForm.value.new)
  if (!valid) {
    const errors = getPasswordErrors(passwordForm.value.new)
    showError('Erreur', errors[0])
    return
  }

  // Check if MFA is required for destructive actions
  try {
    const mfaRequired = await isMFARequired('destructive')
    if (mfaRequired) {
      // Store the action to execute after MFA verification
      pendingAction.value = async () => {
        await executePasswordUpdate()
      }
      mfaChallengeContext.value = 'destructive'
      showMFAChallenge.value = true
      return
    }
  } catch (err) {
    console.error('Error checking MFA requirement:', err)
    // Continue without MFA if check fails (not critical)
  }

  // Execute password update directly if MFA not required
  await executePasswordUpdate()
}

const executePasswordUpdate = async () => {
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

const handleUpdateAvatar = async () => {
  if (!selectedAvatar.value) {
    showError('Erreur', 'Veuillez sélectionner un avatar.')
    return
  }

  if (selectedAvatar.value === authStore.user?.avatar_url) {
    showError('Erreur', 'Veuillez sélectionner un avatar différent.')
    return
  }

  updatingAvatar.value = true
  try {
    await authStore.updateAvatar(selectedAvatar.value)
    showSuccess('Succès', 'Votre avatar a été mis à jour avec succès !')
  } catch (error) {
    console.error("Failed to update avatar:", error)
    const errorMessage = error.response?.data?.error || error.message || 'Erreur lors de la mise à jour de l\'avatar.'
    showError('Erreur', errorMessage)
  } finally {
    updatingAvatar.value = false
  }
}

// RGPD Article 20 - Droit a la portabilite
const handleExportData = async () => {
  exportingData.value = true
  try {
    const response = await api.get('/users/export', { responseType: 'blob' })
    const blob = new Blob([response.data], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    const date = new Date().toISOString().split('T')[0]
    link.href = url
    link.download = `kagibi-export-${date}.json`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
    showSuccess('Export termine', 'Vos donnees ont ete exportees avec succes au format JSON.')
  } catch (error) {
    console.error('Failed to export data:', error)
    const errorMessage = error.response?.data?.error || error.message || 'Erreur lors de l\'export des donnees.'
    showError('Erreur', errorMessage)
  } finally {
    exportingData.value = false
  }
}

const handleDeleteAccount = async () => {
  if (deleteConfirmationText.value !== 'SUPPRIMER DEFINITIVEMENT') {
    showError('Erreur', 'Veuillez taper "SUPPRIMER DEFINITIVEMENT" pour confirmer la suppression irréversible')
    return
  }

  // Confirmation supplémentaire pour éviter les suppressions accidentelles
  const finalConfirm = confirm(
    'DERNIERE CONFIRMATION\n\n' +
    'Votre compte et TOUTES vos donnees seront SUPPRIMES IMMEDIATEMENT.\n\n' +
    'Cette action est DEFINITIVEMENT IRREVERSIBLE.\n\n' +
    'AUCUNE RECUPERATION ne sera possible.\n\n' +
    'Etes-vous absolument certain(e) ?'
  )

  if (!finalConfirm) {
    return
  }

  // Check if MFA is required for destructive actions
  try {
    const mfaRequired = await isMFARequired('destructive')
    if (mfaRequired) {
      // Store the action to execute after MFA verification
      pendingAction.value = async () => {
        await executeDeleteAccount()
      }
      mfaChallengeContext.value = 'destructive'
      showMFAChallenge.value = true
      return
    }
  } catch (err) {
    console.error('Error checking MFA requirement:', err)
    // Continue without MFA if check fails (not critical)
  }

  // Execute deletion directly if MFA not required
  await executeDeleteAccount()
}

const executeDeleteAccount = async () => {
  isDeletingAccount.value = true

  try {
    await authStore.deleteAccount('SUPPRIMER')

    // Redirection vers la page d'accueil
    router.push('/')

    // Notification de succès
    alert(
      'Votre compte a ete definitivement supprime.\n\n' +
      'Toutes vos donnees sont irrecuperables.\n\n' +
      'Conformement au RGPD (Article 17), vos donnees personnelles ont ete effacees.'
    )

  } catch (error) {
    console.error('Failed to delete account:', error)
    const errorMessage = error.response?.data?.error || error.message || 'Erreur lors de la suppression du compte.'
    showError('Erreur', errorMessage)
  } finally {
    isDeletingAccount.value = false
    closeDeleteModal()
  }
}
</script>

<style scoped>
.account-page {
  padding: 2rem;
  background-color: var(--background-color);
  height: 100%;
  overflow-y: auto;
  overflow-x: hidden;
  box-sizing: border-box;
  max-width: 100%;
}

.page-header {
  margin-bottom: 2rem;
}

.header-content { display: flex; align-items: center; gap: 16px; margin-bottom: 8px; }

.btn-back {
  background: none; border: none; display: flex; align-items: center; justify-content: center;
  gap: 6px; color: var(--secondary-text-color); cursor: pointer; font-size: 0.9rem;
  padding: 6px 12px; border-radius: 8px; transition: all 0.2s;
}
.btn-back:hover { background-color: var(--hover-background-color); color: var(--primary-color); }

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

@media (min-width: 901px) {
  .user-card {
    position: sticky;
    top: 2rem;
  }
}

@media (max-width: 900px) {
  .content-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 600px) {
  .account-page {
    padding: 1rem;
  }

  .page-header {
    margin-bottom: 1.25rem;
  }

  .header-content h1 {
    font-size: 1.4rem;
  }

  .content-grid {
    gap: 1rem;
  }

  .user-card {
    padding: 1.25rem;
  }

  .section-header,
  .section-body {
    padding: 1rem;
  }

  .form-row {
    flex-direction: column;
    align-items: stretch;
    gap: 0.75rem;
  }

  .form-row .input-group {
    width: 100%;
  }

  .form-row button {
    width: 100%;
  }

  .password-form {
    gap: 0.75rem;
  }

  .password-row {
    flex-direction: column;
  }

  .danger-zone-item {
    flex-direction: column;
    align-items: flex-start;
  }

  .danger-zone-item .btn-danger-outline {
    width: 100%;
  }

  .form-actions {
    justify-content: stretch;
  }

  .form-actions button,
  .portability-actions button {
    width: 100%;
  }

  .portability-actions {
    justify-content: stretch;
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
}

.avatar-container {
  margin-bottom: 1.5rem;
  display: flex;
  justify-content: center;
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

.settings-section.danger-zone .section-header {
  background: var(--card-color);
}

.settings-section.danger-zone .section-header h3 {
  color: var(--main-text-color);
}

.danger-zone-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
}

.danger-zone-info h4 {
  margin: 0 0 0.5rem 0;
  color: var(--main-text-color);
}

.danger-zone-info p {
  margin: 0;
  color: var(--secondary-text-color);
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

.form-divider {
  border: none;
  border-top: 1px solid var(--border-color, #e2e8f0);
  margin: 1rem 0;
}

.password-row {
  display: flex;
  gap: 1rem;
}

.password-row .input-group {
  flex: 1;
}

.password-stack {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
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
  width: 100%;
  box-sizing: border-box;
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

.btn-danger-outline {
  background-color: transparent;
  border: 1px solid var(--error-color);
  color: var(--error-color);
  padding: 0.8rem 1.5rem;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 600;
  transition: background-color 0.2s, color 0.2s, border-color 0.2s;
}

.btn-danger-outline:hover {
  background-color: rgba(239, 68, 68, 0.08);
  border-color: var(--error-color);
  color: var(--error-color);
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

/* Portability Section */
.portability-item {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.portability-info {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.portability-desc {
  color: var(--secondary-text-color);
  line-height: 1.6;
  margin: 0;
}

.portability-details {
  color: var(--secondary-text-color);
  font-size: 0.9rem;
  line-height: 1.5;
  margin: 0;
}

.portability-actions {
  display: flex;
  justify-content: flex-end;
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

/* Avatar Section */
.avatar-section {
  display: flex;

/* Danger Zone Section */
.danger-content {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.warning-text-subtle {
  color: var(--secondary-text-color);
  line-height: 1.6;
  margin: 0 0 0.5rem 0;
}

.danger-details {
  border: 1px solid var(--border-color);
  border-radius: 6px;
  padding: 0;
  margin: 0.5rem 0;
}

.danger-summary {
  padding: 0.75rem 1rem;
  cursor: pointer;
  user-select: none;
  color: var(--secondary-text-color);
  font-weight: 500;
  list-style: none;
  transition: background-color 0.2s;
}

.danger-summary::-webkit-details-marker {
  display: none;
}

.danger-summary::before {
  content: "▶";
  display: inline-block;
  margin-right: 0.5rem;
  transition: transform 0.2s;
  font-size: 0.75rem;
}

details[open] .danger-summary::before {
  transform: rotate(90deg);
}

.danger-summary:hover {
  background-color: rgba(0, 0, 0, 0.02);
}

.warning-list-subtle {
  list-style: none;
  padding: 0 1.5rem 0.5rem 2rem;
  margin: 0;
}

.warning-list-subtle li {
  padding: 0.3rem 0;
  color: var(--secondary-text-color);
  line-height: 1.5;
  position: relative;
}

.warning-list-subtle li::before {
  content: "•";
  position: absolute;
  left: -1rem;
  color: #dc3545;
}

.rgpd-note {
  padding: 0 1.5rem 1rem 1.5rem;
  margin: 0;
  font-size: 0.875rem;
  color: var(--secondary-text-color);
  font-style: italic;
}

.btn-delete-account {
  background-color: transparent;
  color: #dc3545;
  border: 1px solid #dc3545;
  padding: 0.75rem 1.25rem;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 500;
  font-size: 0.95rem;
  transition: all 0.2s;
  align-self: flex-start;
  margin-top: 0.5rem;
}

.btn-delete-account:hover {
  background-color: #dc3545;
  color: white;
}

  flex-direction: column;
  gap: 1.5rem;
}

.current-avatar-display {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem;
  background: var(--background-color);
  border-radius: 8px;
}

.avatar-preview {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--primary-color), var(--accent-color));
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  box-shadow: 0 4px 10px rgba(0,0,0,0.1);
  position: relative;
}

.avatar-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.avatar-preview .avatar-initials {
  color: white;
  font-size: 2rem;
  font-weight: bold;
}

.avatar-hint {
  font-size: 0.85rem;
  color: var(--secondary-text-color);
  margin: 0;
}
</style>
