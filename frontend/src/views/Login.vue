<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="auth-page">
    <div class="auth-layout">
      <!-- Left Column: Branding & Value Proposition -->
      <div class="auth-branding">
        <div class="brand-header">
          <div class="logo-placeholder" :class="{ 'logo-p2p': isP2PSubdomain }">
            <svg v-if="!isP2PSubdomain" viewBox="0 0 24 24" fill="none" class="logo-icon" stroke="currentColor" stroke-width="2">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
              <polyline points="7 10 12 15 17 10" />
              <line x1="12" y1="15" x2="12" y2="3" />
            </svg>
            <svg v-else viewBox="0 0 24 24" fill="none" class="logo-icon" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <line x1="22" y1="2" x2="11" y2="13"></line>
              <polygon points="22 2 15 22 11 13 2 9 22 2"></polygon>
            </svg>
          </div>
          <div class="brand-title">
            <h1>Kagibi<span v-if="isP2PSubdomain" class="brand-suffix"> Send</span></h1>
            <span v-if="isP2PSubdomain" class="brand-badge">Transfert P2P</span>
          </div>
        </div>

        <h2 class="auth-tagline">{{ isP2PSubdomain ? t('auth.p2pTagline') : t('auth.tagline') }}</h2>

        <div class="features-grid">
          <template v-if="!isP2PSubdomain">
            <div class="feature-item">
              <div class="feature-icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="icon-svg">
                  <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
                  <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
                </svg>
              </div>
              <div class="feature-text">
                <h3>{{ t('auth.e2eEncryption') }}</h3>
                <p>{{ t('auth.e2eDesc') }}</p>
              </div>
            </div>
            <div class="feature-item">
              <div class="feature-icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="icon-svg">
                  <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"></path>
                  <line x1="1" y1="1" x2="23" y2="23"></line>
                </svg>
              </div>
              <div class="feature-text">
                <h3>{{ t('auth.zeroKnowledge') }}</h3>
                <p>{{ t('auth.zeroKnowledgeDesc') }}</p>
              </div>
            </div>
            <div class="feature-item">
              <div class="feature-icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="icon-svg">
                  <line x1="22" y1="2" x2="11" y2="13"></line>
                  <polygon points="22 2 15 22 11 13 2 9 22 2"></polygon>
                </svg>
              </div>
              <div class="feature-text">
                <h3>{{ t('auth.secureSharing') }}</h3>
                <p>{{ t('auth.secureSharingDesc') }}</p>
              </div>
            </div>
          </template>
          <template v-else>
            <div class="feature-item">
              <div class="feature-icon feature-icon-p2p">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="icon-svg">
                  <polyline points="5 12 19 12"></polyline>
                  <polyline points="13 6 19 12 13 18"></polyline>
                </svg>
              </div>
              <div class="feature-text">
                <h3>{{ t('auth.p2pFeature1') }}</h3>
                <p>{{ t('auth.p2pFeature1Desc') }}</p>
              </div>
            </div>
            <div class="feature-item">
              <div class="feature-icon feature-icon-p2p">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="icon-svg">
                  <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
                  <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
                </svg>
              </div>
              <div class="feature-text">
                <h3>{{ t('auth.p2pFeature2') }}</h3>
                <p>{{ t('auth.p2pFeature2Desc') }}</p>
              </div>
            </div>
            <div class="feature-item">
              <div class="feature-icon feature-icon-p2p">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="icon-svg">
                  <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
                  <circle cx="9" cy="7" r="4"></circle>
                  <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
                  <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
                </svg>
              </div>
              <div class="feature-text">
                <h3>{{ t('auth.p2pFeature3') }}</h3>
                <p>{{ t('auth.p2pFeature3Desc') }}</p>
              </div>
            </div>
          </template>
        </div>
      </div>

      <!-- Right Column: Authentication Forms -->
      <div class="auth-interaction">
        <Transition name="fade-slide" mode="out-in">
          <div v-if="mode === 'login'" key="login" class="auth-form-container">
            <div class="form-header">
              <div class="form-header-titles">
                <h2>{{ t('auth.login') }}</h2>
                <a href="#" class="header-register-link" @click.prevent="mode = 'register'">{{ t('auth.register') }}</a>
              </div>
              <p>{{ t('auth.welcomeBack') }}</p>
            </div>

            <LoginComponent />

            <div class="auth-separator">
              <span>{{ t('auth.securityTitle') }}</span>
            </div>

            <div class="security-note">
              <p>{{ t('auth.passwordManagerReco') }}</p>
              <ul class="pwd-manager-list">
                <li><a href="https://keepass.fr/tutoriel-pour-keepass-le-guide-complet/" target="_blank" rel="noopener noreferrer">Tutoriel KeePass</a> (Gratuit & Local)</li>
                <li><a href="https://bitwarden.com/fr-fr/help/courses/password-manager-personal/" target="_blank" rel="noopener noreferrer">Tutoriel Bitwarden</a> (Cloud & Open Source)</li>
              </ul>
            </div>

            <div class="auth-footer">
              <p>
                {{ t('auth.noAccount') }}
                <a href="#" @click.prevent="mode = 'register'" class="action-link">{{ t('auth.createAccount') }}</a>
              </p>
              <p>
                <a href="#" @click.prevent="mode = 'recovery'" class="dimmed-link">{{ t('auth.forgotPassword') }}</a>
              </p>
            </div>
          </div>

          <div v-else-if="mode === 'register'" key="register" class="auth-form-container">
            <div class="form-header">
              <div class="form-header-titles">
                <a href="#" class="header-register-link" @click.prevent="mode = 'login'">{{ t('auth.login') }}</a>
                <h2>{{ t('auth.register') }}</h2>
              </div>
              <p>{{ t('auth.startSecuring') }}</p>
            </div>

            <RegisterComponent />

            <div class="security-warning">
              <div class="warning-icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="icon-svg-alert">
                   <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
                   <line x1="12" y1="9" x2="12" y2="13"></line>
                   <line x1="12" y1="17" x2="12.01" y2="17"></line>
                </svg>
              </div>
              <p>{{ t('auth.passwordWarning') }}</p>
            </div>

            <div class="auth-footer">
              <p>
                {{ t('auth.hasAccount') }}
                <a href="#" @click.prevent="mode = 'login'" class="action-link">{{ t('auth.signIn') }}</a>
              </p>
            </div>
          </div>

          <div v-else-if="mode === 'recovery'" key="recovery" class="auth-form-container">
            <div class="form-header">
              <h2>{{ t('auth.recovery') }}</h2>
              <p>{{ t('auth.useRecoveryCode') }}</p>
            </div>
            <RecoveryComponent @cancel="mode = 'login'" @success="mode = 'login'" />
          </div>
        </Transition>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import LoginComponent from '../components/auth/loginComponent.vue'
import RegisterComponent from '../components/auth/registerComponent.vue'
import RecoveryComponent from '../components/auth/recoveryComponent.vue'
import { useFileStore } from '../stores/files'
import { useAuthStore } from '../stores/auth'
import { isP2PSubdomain } from '../composables/useSubdomain'

const { t } = useI18n()

const mode = ref('login') // 'login', 'register', 'recovery'
const fileStore = useFileStore()
const authStore = useAuthStore()

onMounted(() => {
  // Reset states
  fileStore.recentFolders = []
  fileStore.recentFiles = []
  fileStore.folders = []
  fileStore.files = []
  fileStore.currentPath = '/'
  localStorage.removeItem('files')
  localStorage.removeItem('file')
  authStore.privateKey = null
  authStore.publicKey = null
  authStore.masterKey = null
  authStore.user = null
  localStorage.removeItem('auth')
})
</script>

<style scoped>
.auth-page {
  /* Removed min-height + align-items center which caused top-overflow clipping */
  height: 100vh;
  width: 100%;
  background-color: var(--background-color);
  overflow-y: auto; /* Enable page scrolling if content is taller than viewport */
  display: flex;
  flex-direction: column;
}

.auth-layout {
  /* Use margin: auto within the flex/block container to center vertically safe */
  margin: auto;
  padding: 2rem 2rem; /* Give some breathing room when scrolling */
  display: grid;
  grid-template-columns: 1fr 1fr;
  width: 100%;
  max-width: 80%;
  gap: 2rem;
  align-items: center;
}

/* Left Column: Branding */
.auth-branding {
  display: flex;
  flex-direction: column;
  gap: 2rem;
  color: var(--main-text-color);
}

.brand-header {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.brand-title {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.brand-suffix {
  color: var(--primary-color);
}

.brand-badge {
  display: inline-block;
  font-size: 0.75rem;
  font-weight: 600;
  padding: 0.2rem 0.6rem;
  background: rgba(0, 80, 255, 0.12);
  color: var(--primary-color);
  border-radius: 20px;
  letter-spacing: 0.05em;
  width: fit-content;
}

.logo-p2p {
  background: linear-gradient(135deg, #0050FF, #00b4d8);
}

.feature-icon-p2p {
  background-color: rgba(0, 80, 255, 0.08);
  color: var(--primary-color);
}

.logo-placeholder {
  width: 48px;
  height: 48px;
  background: linear-gradient(135deg, var(--primary-color), var(--accent-color));
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.logo-icon {
  width: 24px;
  height: 24px;
}

.auth-branding h1 {
  font-size: 2.5rem;
  font-weight: 700;
  margin: 0;
  background: linear-gradient(to right, var(--primary-color), var(--accent-color));
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
}

.auth-tagline {
  font-size: 2rem;
  font-weight: 600;
  line-height: 1.2;
  margin: 0;
  opacity: 0.9;
}

.features-grid {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  margin-top: 1.5rem;
}

.feature-item {
  display: flex;
  gap: 1.25rem;
  align-items: center;
  padding: 1rem;
  border-radius: 16px;
  background-color: var(--background-color); /* Matches page bg, but border gives shape */
  border: 1px solid var(--border-color);
  transition: all 0.3s cubic-bezier(0.25, 0.8, 0.25, 1);
  /* Ensure text doesn't center itself weirdly */
  justify-content: flex-start;
  height: auto;
}

.feature-item:hover {
  transform: translateY(-3px);
  border-color: var(--primary-color);
  box-shadow: 0 12px 24px -10px rgba(0, 0, 0, 0.1);
  background-color: var(--card-color);
}

.feature-icon {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12px;
  background-color: rgba(52, 152, 219, 0.08);
  color: var(--primary-color);
  flex-shrink: 0;
}

.icon-svg {
  width: 24px;
  height: 24px;
  stroke: currentColor;
  stroke-width: 2px;
}

.feature-text {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  text-align: left;
}

.feature-text h3 {
  font-size: 1.05rem;
  margin: 0;
  font-weight: 600;
  color: var(--main-text-color);
}

.feature-text p {
  margin: 0;
  font-size: 0.9rem;
  color: var(--secondary-text-color);
  line-height: 1.4;
}

/* Right Column: Interaction */
.auth-interaction {
  width: 100%;
  max-width: 480px;
  margin-left: auto;
}

.auth-form-container {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.form-header-titles {
  display: flex;
  align-items: baseline;
  gap: 1rem;
  margin-bottom: 0.5rem;
}

.form-header h2 {
  font-size: 2rem;
  margin: 0;
  color: var(--main-text-color);
}

.header-register-link {
  font-size: 1.2rem; /* 2rem * 0.6 = 40% smaller */
  color: var(--secondary-text-color);
  text-decoration: none;
  font-weight: 600;
  transition: color 0.2s;
}

.header-register-link:hover {
  color: var(--primary-color);
}

.form-header p {
  margin: 0;
  color: var(--secondary-text-color);
  font-size: 1.1rem;
}

/* Security Note & Warning */
.security-note {
  font-size: 0.9rem;
  color: var(--secondary-text-color);
  background: rgba(52, 152, 219, 0.05);
  padding: 1rem;
  border-radius: 8px;
  border-left: 3px solid var(--primary-color);
}

.security-note p { margin: 0; }

.pwd-manager-list {
  margin: 0.5rem 0 0 0;
  padding-left: 1.2rem;
  text-align: left;
}

.pwd-manager-list li {
  margin-bottom: 0.25rem;
  display: flex;
  align-items: center;
}
.icon-svg-alert {
  width: 24px;
  height: 24px;
  stroke: var(--danger-color);

}

.security-warning {
  display: flex;
  gap: 1rem;
  background: rgba(231, 76, 60, 0.05);
  padding: 1rem;
  border-radius: 8px;
  border: 1px solid rgba(231, 76, 60, 0.2);
  color: var(--danger-color, #e74c3c);
  font-size: 0.9rem;
  align-items: center;
}

.warning-icon { font-size: 1.5rem; }
.security-warning p { margin: 0; }

/* Footer Links */
.auth-separator {
  display: flex;
  align-items: center;
  text-align: center;
  color: var(--secondary-text-color);
  font-size: 0.85rem;
  margin: 0.5rem 0;
}

.auth-separator::before,
.auth-separator::after {
  content: '';
  flex: 1;
  border-bottom: 1px solid var(--border-color);
}

.auth-separator span {
  padding: 0 10px;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.auth-footer {
  text-align: center;
  display: flex;
  flex-direction: column;
  gap: 0.8rem;
  margin-top: 1rem;
}

.auth-footer p { margin: 0; }

.action-link {
  color: var(--primary-color);
  font-weight: 600;
  text-decoration: none;
  font-size: 1.05rem;
}

.action-link:hover { text-decoration: underline; }

.dimmed-link {
  color: var(--secondary-text-color);
  text-decoration: none;
  font-size: 0.9rem;
}

.dimmed-link:hover { color: var(--main-text-color); }

/* Transitions */
.fade-slide-enter-active,
.fade-slide-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.fade-slide-enter-from {
  opacity: 0;
  transform: translateX(20px);
}

.fade-slide-leave-to {
  opacity: 0;
  transform: translateX(-20px);
}

/* Responsive */
@media (max-width: 900px) {
  .auth-layout {
    grid-template-columns: 1fr;
    max-width: 500px;
    gap: 3rem;
  }

  .auth-branding {
    text-align: center;
    align-items: center;
  }

  .feature-item {
    text-align: left;
  }

  .auth-interaction {
    margin: 0 auto;
  }
}

@media (max-width: 600px) {
  .auth-layout {
    max-width: 100%;
    padding: 1.5rem 1rem;
    gap: 2rem;
  }

  .auth-branding h1 {
    font-size: clamp(1.6rem, 6vw, 2.5rem);
  }

  .auth-tagline {
    font-size: clamp(1.1rem, 4vw, 2rem);
  }

  .auth-branding {
    gap: 1rem;
  }

  .features-grid {
    gap: 0.75rem;
    margin-top: 0.75rem;
  }

  .feature-item {
    padding: 0.75rem;
    gap: 0.75rem;
  }

  .feature-icon {
    width: 36px;
    height: 36px;
    flex-shrink: 0;
  }
}

@media (max-width: 400px) {
  .auth-layout {
    padding: 1rem 0.75rem;
  }

  .logo-placeholder {
    width: 36px;
    height: 36px;
  }
}
</style>
