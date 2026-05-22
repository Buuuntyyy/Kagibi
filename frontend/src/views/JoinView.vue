<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="join-wrapper">
    <header class="join-header">
      <div class="brand">
        <svg class="brand-logo" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M19.35 10.04C18.67 6.59 15.64 4 12 4 9.11 4 6.6 5.64 5.35 8.04 2.34 8.36 0 10.91 0 14c0 3.31 2.69 6 6 6h13c2.76 0 5-2.24 5-5 0-2.64-2.05-4.78-4.65-4.96z" fill="url(#joinGrad)"/>
          <defs>
            <linearGradient id="joinGrad" x1="0" y1="0" x2="24" y2="24" gradientUnits="userSpaceOnUse">
              <stop offset="0%" stop-color="var(--primary-color)" />
              <stop offset="100%" stop-color="var(--accent-color)" />
            </linearGradient>
          </defs>
        </svg>
        <span class="brand-name">Kagibi</span>
      </div>
    </header>

    <div class="join-content">
      <!-- Loading -->
      <div v-if="state === 'loading'" class="join-card glass-panel">
        <div class="spinner-lg"></div>
        <p class="hint">{{ t('join.loading') }}</p>
      </div>

      <!-- Error states -->
      <div v-else-if="state === 'error'" class="join-card glass-panel">
        <div class="icon-wrap icon-error">
          <svg viewBox="0 0 24 24" width="48" height="48" fill="currentColor"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/></svg>
        </div>
        <h2>{{ t('join.invalid') }}</h2>
        <p class="hint">{{ errorMessage }}</p>
        <router-link to="/dashboard/home" class="btn-primary">{{ t('join.goToDashboard') }}</router-link>
      </div>

      <!-- Already a member -->
      <div v-else-if="state === 'already-member'" class="join-card glass-panel">
        <div class="icon-wrap icon-success">
          <svg viewBox="0 0 24 24" width="48" height="48" fill="currentColor"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/></svg>
        </div>
        <h2>{{ t('join.alreadyMember') }}</h2>
        <p class="hint">{{ t('join.alreadyMemberHint', { org: invitation?.org_name }) }}</p>
        <router-link :to="`/dashboard/organizations/${invitation?.org_id}`" class="btn-primary">
          {{ t('join.goToOrg') }}
        </router-link>
      </div>

      <!-- Success -->
      <div v-else-if="state === 'joined'" class="join-card glass-panel">
        <div class="icon-wrap icon-success">
          <svg viewBox="0 0 24 24" width="48" height="48" fill="currentColor"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/></svg>
        </div>
        <h2>{{ t('join.joined') }}</h2>
        <p class="hint">{{ t('join.joinedHint', { org: invitation?.org_name }) }}</p>
        <router-link :to="`/dashboard/organizations/${joinedOrgID}`" class="btn-primary">
          {{ t('join.goToOrg') }}
        </router-link>
      </div>

      <!-- Not authenticated -->
      <div v-else-if="state === 'unauthenticated'" class="join-card glass-panel">
        <div class="icon-wrap icon-info">
          <svg viewBox="0 0 24 24" width="48" height="48" fill="currentColor"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-6h2v6zm0-8h-2V7h2v2z"/></svg>
        </div>
        <div class="org-preview" v-if="invitation">
          <div class="org-avatar-lg">{{ invitation.org_name?.charAt(0)?.toUpperCase() }}</div>
          <div>
            <h2>{{ invitation.org_name }}</h2>
            <p class="role-hint">{{ t('join.youreInvitedAs', { role: t(`orgs.${invitation.role}`) }) }}</p>
          </div>
        </div>
        <p class="hint">{{ t('join.loginRequired') }}</p>
        <a :href="loginURL" class="btn-primary">{{ t('join.goToLogin') }}</a>
      </div>

      <!-- Confirm join -->
      <div v-else-if="state === 'confirm'" class="join-card glass-panel">
        <div class="org-preview">
          <div class="org-avatar-lg">{{ invitation?.org_name?.charAt(0)?.toUpperCase() }}</div>
          <div>
            <h2>{{ invitation?.org_name }}</h2>
            <p class="role-hint">{{ t('join.youreInvitedAs', { role: t(`orgs.${invitation?.role}`) }) }}</p>
          </div>
        </div>

        <p v-if="joinError" class="form-error">{{ joinError }}</p>

        <div class="join-actions">
          <router-link to="/dashboard/home" class="btn-secondary">{{ t('join.decline') }}</router-link>
          <button class="btn-primary" @click="doAccept" :disabled="joining">
            <span v-if="joining" class="spinner-sm"></span>
            {{ t('join.accept') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useOrgStore } from '../stores/organizations'
import { useAuthStore } from '../stores/auth'
import { generateOrgKey, encryptOrgKeyForUser } from '../utils/orgCrypto.js'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const orgStore = useOrgStore()
const authStore = useAuthStore()

const token = computed(() => route.params.token)
const state = ref('loading') // loading | error | unauthenticated | confirm | already-member | joined
const invitation = ref(null)
const errorMessage = ref('')
const joinError = ref('')
const joining = ref(false)
const joinedOrgID = ref(null)

const loginURL = computed(() => `/login?redirect=${encodeURIComponent(route.fullPath)}`)

onMounted(async () => {
  // Fetch invitation info (public endpoint — no auth needed)
  try {
    invitation.value = await orgStore.getInvitation(token.value)
  } catch (e) {
    const status = e.response?.status
    if (status === 404) {
      errorMessage.value = t('join.notFound')
    } else if (status === 410) {
      errorMessage.value = e.response?.data?.error || t('join.expired')
    } else {
      errorMessage.value = e.response?.data?.error || e.message
    }
    state.value = 'error'
    return
  }

  // Check authentication
  const isAuthenticated = await authStore.checkAuth()
  if (!isAuthenticated) {
    state.value = 'unauthenticated'
    return
  }

  state.value = 'confirm'
})

const doAccept = async () => {
  joining.value = true
  joinError.value = ''
  try {
    let encryptedOrgKey = ''

    // Owner invitations come from the admin CLI — no OrgKey exists yet.
    // The first owner generates it and encrypts it with their own RSA public key.
    if (invitation.value?.role === 'owner') {
      const publicKeyPEM = authStore.user?.public_key
      if (!publicKeyPEM) throw new Error('Clé publique introuvable. Reconnectez-vous.')
      const orgKey = await generateOrgKey()
      encryptedOrgKey = await encryptOrgKeyForUser(orgKey, publicKeyPEM)
    }
    // For link invites (non-owner): join without a key.
    // An admin must later call provisionMemberKey to distribute the org key.

    const result = await orgStore.acceptInvitation(token.value, encryptedOrgKey)
    joinedOrgID.value = result.org_id
    state.value = 'joined'
  } catch (e) {
    const status = e.response?.status
    if (status === 409) {
      state.value = 'already-member'
    } else if (status === 410) {
      errorMessage.value = e.response?.data?.error || t('join.expired')
      state.value = 'error'
    } else {
      joinError.value = e.response?.data?.error || e.message
    }
  } finally {
    joining.value = false
  }
}
</script>

<style scoped>
.join-wrapper {
  min-height: 100vh;
  background: var(--background-color);
  display: flex;
  flex-direction: column;
}

.join-header {
  padding: 20px 32px;
  border-bottom: 1px solid var(--border-color);
}

.brand {
  display: flex;
  align-items: center;
  gap: 10px;
}

.brand-logo {
  width: 32px;
  height: 32px;
}

.brand-name {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--main-text-color);
  letter-spacing: -0.02em;
}

.join-content {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
}

.join-card {
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 16px;
  padding: 48px 40px;
  width: 100%;
  max-width: 440px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20px;
  text-align: center;
}

.icon-wrap {
  width: 72px;
  height: 72px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.icon-error { background: rgba(239, 68, 68, 0.12); color: #ef4444; }
.icon-success { background: rgba(34, 197, 94, 0.12); color: #22c55e; }
.icon-info { background: rgba(99, 102, 241, 0.12); color: var(--primary-color); }

.join-card h2 {
  font-size: 1.3rem;
  font-weight: 700;
  color: var(--main-text-color);
  margin: 0;
}

.hint {
  font-size: 0.9rem;
  color: var(--secondary-text-color);
  margin: 0;
  line-height: 1.5;
}

.org-preview {
  display: flex;
  align-items: center;
  gap: 16px;
  text-align: left;
  width: 100%;
  padding: 16px;
  background: var(--hover-background-color);
  border-radius: 12px;
  border: 1px solid var(--border-color);
}

.org-preview h2 {
  font-size: 1.1rem;
  margin-bottom: 4px;
}

.org-avatar-lg {
  width: 52px;
  height: 52px;
  border-radius: 12px;
  background: var(--primary-color);
  color: white;
  font-size: 1.4rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.role-hint {
  font-size: 0.85rem;
  color: var(--secondary-text-color);
  margin: 0;
}

.join-actions {
  display: flex;
  gap: 12px;
  width: 100%;
  justify-content: flex-end;
}

.form-error {
  color: #ef4444;
  font-size: 0.85rem;
  margin: 0;
}

.btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: var(--primary-color);
  color: white;
  border: none;
  border-radius: 8px;
  padding: 10px 20px;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  text-decoration: none;
  transition: opacity 0.15s;
}

.btn-primary:hover:not(:disabled) { opacity: 0.85; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }

.btn-secondary {
  display: inline-flex;
  align-items: center;
  background: none;
  color: var(--secondary-text-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 10px 20px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  text-decoration: none;
  transition: background 0.15s;
}

.btn-secondary:hover { background: var(--hover-background-color); }

.spinner-lg {
  width: 40px;
  height: 40px;
  border: 3px solid var(--border-color);
  border-top-color: var(--primary-color);
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

.spinner-sm {
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255,255,255,0.4);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

@keyframes spin { to { transform: rotate(360deg); } }
</style>
