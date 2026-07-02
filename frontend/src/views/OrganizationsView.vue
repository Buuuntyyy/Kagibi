<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="orgs-page">

    <!-- ── Page header ─────────────────────────────────────────────────── -->
    <div class="page-header">
      <div class="page-header-text">
        <h1>{{ t('orgs.title') }}</h1>
        <p class="page-subtitle">{{ t('orgs.pageSubtitle') }}</p>
      </div>
      <button
        v-if="isPremium && orgStore.orgs.length > 0"
        class="btn-create"
        @click="showCreateModal = true"
      >
        <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z"/></svg>
        {{ t('orgs.createOrg') }}
      </button>
    </div>

    <!-- ── Billing loading ─────────────────────────────────────────────── -->
    <div v-if="!billingReady" class="center-state">
      <div class="spinner"></div>
    </div>

    <!-- ── Paywall ─────────────────────────────────────────────────────── -->
    <div v-else-if="!isPremium" class="paywall-wrapper">
      <div class="paywall-card">
        <div class="paywall-glow"></div>
        <div class="paywall-icon-wrap">
          <svg viewBox="0 0 24 24" width="32" height="32" fill="currentColor"><path d="M12 7V3H2v18h20V7H12zM6 19H4v-2h2v2zm0-4H4v-2h2v2zm0-4H4V9h2v2zm0-4H4V5h2v2zm4 12H8v-2h2v2zm0-4H8v-2h2v2zm0-4H8V9h2v2zm0-4H8V5h2v2zm10 12h-8v-2h2v-2h-2v-2h2v-2h-2V9h8v10zm-2-8h-2v2h2v-2zm0 4h-2v2h2v-2z"/></svg>
        </div>
        <span class="paywall-badge">
          <svg viewBox="0 0 24 24" width="11" height="11" fill="currentColor"><path d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4zm-2 16l-4-4 1.41-1.41L10 14.17l6.59-6.59L18 9l-8 8z"/></svg>
          {{ t('orgs.premiumBadge') }}
        </span>
        <h2 class="paywall-title">{{ t('orgs.premiumTitle') }}</h2>
        <p class="paywall-desc">{{ t('orgs.premiumDesc') }}</p>
        <ul class="paywall-features">
          <li v-for="n in 5" :key="n">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41L9 16.17z"/></svg>
            {{ t(`orgs.premiumFeature${n}`) }}
          </li>
        </ul>
        <button class="btn-upgrade" @click="router.push('/dashboard/billing')">
          {{ t('orgs.upgradeCta') }}
          <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M8.59 16.59L13.17 12 8.59 7.41 10 6l6 6-6 6z"/></svg>
        </button>
        <p class="paywall-note">{{ t('orgs.upgradeNote') }}</p>
      </div>
    </div>

    <!-- ── Premium content ────────────────────────────────────────────── -->
    <template v-else>
      <div v-if="orgStore.loading" class="center-state">
        <div class="spinner"></div>
      </div>

      <div v-else-if="orgStore.error" class="center-state error">
        <svg viewBox="0 0 24 24" width="36" height="36" fill="currentColor"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/></svg>
        <p>{{ orgStore.error }}</p>
      </div>

      <!-- Empty state -->
      <div v-else-if="orgStore.orgs.length === 0" class="empty-state">
        <div class="empty-visual">
          <div class="empty-circle c1"></div>
          <div class="empty-circle c2"></div>
          <div class="empty-circle c3"></div>
          <svg viewBox="0 0 24 24" width="40" height="40" fill="currentColor" class="empty-icon"><path d="M12 7V3H2v18h20V7H12zM6 19H4v-2h2v2zm0-4H4v-2h2v2zm0-4H4V9h2v2zm0-4H4V5h2v2zm4 12H8v-2h2v2zm0-4H8v-2h2v2zm0-4H8V9h2v2zm0-4H8V5h2v2zm10 12h-8v-2h2v-2h-2v-2h2v-2h-2V9h8v10zm-2-8h-2v2h2v-2zm0 4h-2v2h2v-2z"/></svg>
        </div>
        <h3>{{ t('orgs.noOrgs') }}</h3>
        <p>{{ t('orgs.noOrgsDesc') }}</p>
        <button class="btn-create" @click="showCreateModal = true">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z"/></svg>
          {{ t('orgs.createOrg') }}
        </button>
      </div>

      <!-- Grid + pinned strip -->
      <div v-else>
      <!-- Pinned strip -->
      <div v-if="pinnedOrgs.length > 0" class="pinned-strip">
        <p class="pinned-strip-label">
          <svg viewBox="0 0 16 16" width="12" height="12" fill="currentColor" xmlns="http://www.w3.org/2000/svg">
            <path d="M4.146.146A.5.5 0 0 1 4.5 0h7a.5.5 0 0 1 .5.5c0 .68-.342 1.174-.646 1.479-.126.125-.25.224-.354.298v4.431l.078.048c.203.127.476.314.751.555C12.36 7.775 13 8.527 13 9.5a.5.5 0 0 1-.5.5h-4v4.5c0 .276-.224 1.5-.5 1.5s-.5-1.224-.5-1.5V10h-4a.5.5 0 0 1-.5-.5c0-.973.64-1.725 1.17-2.168.276-.241.549-.428.752-.555l.078-.048V2.277a2.77 2.77 0 0 1-.354-.298C3.342 1.674 3 1.179 3 .5a.5.5 0 0 1 .146-.354z"/>
          </svg>
          {{ t('orgs.pinned') }}
        </p>
        <div class="pinned-list">
          <button
            v-for="org in pinnedOrgs"
            :key="'pin-' + org.id"
            class="pinned-chip"
            @click="router.push(`/dashboard/organizations/${org.id}`)"
          >
            <span class="pinned-chip-dot" :style="{ background: orgAccent(org) }"></span>
            <span>{{ org.name }}</span>
          </button>
        </div>
      </div>

      <!-- Orgs grid -->
      <div class="orgs-grid">
        <article
          v-for="org in orgStore.orgs"
          :key="org.id"
          class="org-card"
          @click="router.push(`/dashboard/organizations/${org.id}`)"
          tabindex="0"
          @keydown.enter="router.push(`/dashboard/organizations/${org.id}`)"
        >
          <!-- Colored top accent band -->
          <div class="card-accent" :style="{ background: orgAccent(org) }"></div>

          <div class="card-body">
            <!-- Avatar + name row -->
            <div class="card-identity">
              <div class="org-avatar" :style="{ background: orgAccent(org) }">
                <img
                  v-if="org.logo_url"
                  :src="org.logo_url"
                  :alt="org.name"
                  class="org-avatar-img"
                  @error="e => e.target.style.display = 'none'"
                />
                <span v-else>{{ org.name.charAt(0).toUpperCase() }}</span>
              </div>
              <div class="org-meta">
                <h3 class="org-name">{{ org.name }}</h3>
                <p class="org-desc">{{ org.description || t('orgs.noDescription') }}</p>
              </div>
              <span class="role-badge" :class="org.my_role">
                {{ t(`orgs.${org.my_role}`) }}
              </span>
            </div>

            <!-- Stats row -->
            <div class="card-stats">
              <div class="stat-item" v-if="org.member_count">
                <svg viewBox="0 0 24 24" width="13" height="13" fill="currentColor"><path d="M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z"/></svg>
                <span>{{ org.member_count }} {{ t('orgs.members') }}</span>
              </div>
              <div class="stat-item">
                <svg viewBox="0 0 24 24" width="13" height="13" fill="currentColor"><path d="M20 6h-8l-2-2H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2zm0 12H4V8h16v10z"/></svg>
                <span>{{ formatSize(org.storage_used_bytes) }}</span>
              </div>
            </div>

            <!-- Storage progress -->
            <div class="storage-section">
              <div class="storage-labels">
                <span>{{ t('orgs.storageUsed') }}</span>
                <span :class="storagePercent(org) > 90 ? 'text-danger' : storagePercent(org) > 70 ? 'text-warn' : ''">
                  {{ storagePercent(org).toFixed(0) }}%
                </span>
              </div>
              <div class="storage-track">
                <div
                  class="storage-fill"
                  :style="{ width: storagePercent(org) + '%', background: storageColor(org) }"
                ></div>
              </div>
              <div class="storage-quota">{{ formatSize(org.storage_quota_mb * 1024 * 1024) }} {{ t('orgs.total') }}</div>
            </div>
          </div>

          <!-- Pin button -->
          <button
            class="btn-pin"
            :class="{ pinned: orgStore.isPinned(org.id) }"
            :title="orgStore.isPinned(org.id) ? t('orgs.unpin') : t('orgs.pin')"
            @click.stop="orgStore.togglePin(org.id)"
          >
            <!-- Pinned: vertical filled thumbtack -->
            <svg v-if="orgStore.isPinned(org.id)" viewBox="0 0 16 16" width="14" height="14" fill="currentColor" xmlns="http://www.w3.org/2000/svg">
              <path d="M4.146.146A.5.5 0 0 1 4.5 0h7a.5.5 0 0 1 .5.5c0 .68-.342 1.174-.646 1.479-.126.125-.25.224-.354.298v4.431l.078.048c.203.127.476.314.751.555C12.36 7.775 13 8.527 13 9.5a.5.5 0 0 1-.5.5h-4v4.5c0 .276-.224 1.5-.5 1.5s-.5-1.224-.5-1.5V10h-4a.5.5 0 0 1-.5-.5c0-.973.64-1.725 1.17-2.168.276-.241.549-.428.752-.555l.078-.048V2.277a2.77 2.77 0 0 1-.354-.298C3.342 1.674 3 1.179 3 .5a.5.5 0 0 1 .146-.354z"/>
            </svg>
            <!-- Not pinned: diagonal angled thumbtack -->
            <svg v-else viewBox="0 0 16 16" width="14" height="14" fill="currentColor" xmlns="http://www.w3.org/2000/svg">
              <path d="M9.828.722a.5.5 0 0 1 .354.146l4.95 4.95a.5.5 0 0 1 0 .707c-.48.48-1.072.588-1.503.588-.177 0-.335-.018-.46-.039l-3.134 3.134a5.927 5.927 0 0 1 .16 1.013c.046.702-.032 1.687-.72 2.375a.5.5 0 0 1-.707 0l-2.829-2.828-3.182 3.18-1.415 1.413-.707-.707 1.414-1.414 3.182-3.182-2.828-2.829a.5.5 0 0 1 0-.707c.688-.688 1.673-.767 2.375-.72a5.922 5.922 0 0 1 1.013.16l3.134-3.133a2.772 2.772 0 0 1-.04-.461c0-.43.108-1.022.589-1.503a.5.5 0 0 1 .353-.146z"/>
            </svg>
          </button>

          <!-- Hover arrow -->
          <div class="card-arrow">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M8.59 16.59L13.17 12 8.59 7.41 10 6l6 6-6 6z"/></svg>
          </div>
        </article>
      </div><!-- orgs-grid -->
      </div><!-- grid+pinned wrapper -->

    </template><!-- end v-else premium content -->

    <!-- ── Create org modal ───────────────────────────────────────────── -->
    <Transition name="modal">
      <div v-if="showCreateModal" class="modal-overlay" @click.self="showCreateModal = false">
        <div class="modal">
          <div class="modal-header">
            <div class="modal-header-icon">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor"><path d="M12 7V3H2v18h20V7H12zM6 19H4v-2h2v2zm0-4H4v-2h2v2zm0-4H4V9h2v2zm0-4H4V5h2v2zm4 12H8v-2h2v2zm0-4H8v-2h2v2zm0-4H8V9h2v2zm0-4H8V5h2v2zm10 12h-8v-2h2v-2h-2v-2h2v-2h-2V9h8v10zm-2-8h-2v2h2v-2zm0 4h-2v2h2v-2z"/></svg>
            </div>
            <div>
              <h3>{{ t('orgs.createOrg') }}</h3>
              <p class="modal-subtitle">{{ t('orgs.createOrgSubtitle') }}</p>
            </div>
            <button class="btn-close" @click="showCreateModal = false">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
            </button>
          </div>

          <div class="modal-body">
            <div class="form-group">
              <label>{{ t('orgs.orgName') }} <span class="required">*</span></label>
              <input
                v-model="form.name"
                type="text"
                :placeholder="t('orgs.orgNamePlaceholder')"
                class="input-field"
                autofocus
              />
            </div>
            <div class="form-group">
              <label>{{ t('orgs.orgDesc') }}</label>
              <input
                v-model="form.description"
                type="text"
                :placeholder="t('orgs.orgDescPlaceholder')"
                class="input-field"
              />
            </div>
            <div class="form-group">
              <label>{{ t('orgs.storageQuotaMB') }}</label>
              <div class="quota-input-wrap">
                <input v-model.number="form.storageQuotaMB" type="number" min="100" class="input-field" />
                <span class="quota-unit">MB</span>
              </div>
              <p class="field-hint">{{ formatSize(form.storageQuotaMB * 1024 * 1024) }}</p>
            </div>
            <p v-if="createError" class="form-error">
              <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/></svg>
              {{ createError }}
            </p>
          </div>

          <div class="modal-footer">
            <button class="btn-secondary" @click="showCreateModal = false">{{ t('orgs.cancel') }}</button>
            <button class="btn-create" @click="handleCreate" :disabled="creating || !form.name">
              <span v-if="creating" class="spinner-sm"></span>
              <svg v-else viewBox="0 0 24 24" width="15" height="15" fill="currentColor"><path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z"/></svg>
              {{ creating ? t('common.loading') : t('orgs.create') }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useOrgStore } from '../stores/organizations'
import { useAuthStore } from '../stores/auth'
import { useBillingStore } from '../stores/billing'


const { t } = useI18n()
const router = useRouter()
const orgStore = useOrgStore()
const authStore = useAuthStore()
const billingStore = useBillingStore()

const isPremium = computed(() =>
  billingStore.isSelfHosted || (authStore.user?.plan ?? 'free') !== 'free'
)

const billingReady = ref(false)
const showCreateModal = ref(false)
const creating = ref(false)
const createError = ref('')

const form = ref({ name: '', description: '', storageQuotaMB: 10240 })

onMounted(async () => {
  await billingStore.fetchBillingStatus()
  billingReady.value = true
  if (isPremium.value) orgStore.fetchOrgs()
})

const handleCreate = async () => {
  if (!form.value.name) return
  creating.value = true
  createError.value = ''
  try {
    const org = await orgStore.createOrg(form.value.name, form.value.description, form.value.storageQuotaMB)
    showCreateModal.value = false
    form.value = { name: '', description: '', storageQuotaMB: 10240 }
    localStorage.setItem('kagibi_org_onboarding', String(org.id))
    router.push(`/dashboard/organizations/${org.id}`)
  } catch (e) {
    createError.value = e.response?.data?.error || e.message
  } finally {
    creating.value = false
  }
}

const formatSize = (bytes) => {
  if (!bytes) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Number.parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

const storagePercent = (org) => {
  const quota = org.storage_quota_mb * 1024 * 1024
  return quota > 0 ? Math.min((org.storage_used_bytes / quota) * 100, 100) : 0
}

const storageColor = (org) => {
  const pct = storagePercent(org)
  if (pct > 90) return '#ef4444'
  if (pct > 70) return '#f59e0b'
  return 'var(--primary-color)'
}

const ACCENT_PALETTE = [
  'linear-gradient(135deg, #6366f1, #8b5cf6)',
  'linear-gradient(135deg, #0ea5e9, #6366f1)',
  'linear-gradient(135deg, #10b981, #0ea5e9)',
  'linear-gradient(135deg, #f59e0b, #ef4444)',
  'linear-gradient(135deg, #ec4899, #8b5cf6)',
  'linear-gradient(135deg, #14b8a6, #10b981)',
  'linear-gradient(135deg, #f97316, #f59e0b)',
]
const orgAccent = (org) => ACCENT_PALETTE[org.id % ACCENT_PALETTE.length]

const pinnedOrgs = computed(() =>
  orgStore.orgs.filter(o => orgStore.isPinned(o.id))
)
</script>

<style scoped>
/* ── Layout ──────────────────────────────────────────────────────────── */
.orgs-page {
  padding: 32px 32px 48px;
  max-width: 1280px;
  margin: 0 auto;
  min-height: 100%;
  box-sizing: border-box;
}

/* ── Page header ─────────────────────────────────────────────────────── */
.page-header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  margin-bottom: 32px;
  gap: 16px;
  flex-wrap: wrap;
}

.page-header-text h1 {
  font-size: 1.75rem;
  font-weight: 800;
  color: var(--main-text-color);
  margin: 0 0 4px;
  letter-spacing: -0.02em;
}

.page-subtitle {
  font-size: 0.9rem;
  color: var(--secondary-text-color);
  margin: 0;
}

/* ── Create button ───────────────────────────────────────────────────── */
.btn-create {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  background: var(--primary-color);
  color: #fff;
  border: none;
  border-radius: 10px;
  padding: 10px 20px;
  font-size: 0.875rem;
  font-weight: 600;
  cursor: pointer;
  transition: filter 0.15s, transform 0.1s;
  white-space: nowrap;
}
.btn-create:hover { filter: brightness(1.1); }
.btn-create:active { transform: scale(0.97); }
.btn-create:disabled { opacity: 0.55; cursor: not-allowed; filter: none; transform: none; }

/* ── Center states ───────────────────────────────────────────────────── */
.center-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 80px 0;
  color: var(--secondary-text-color);
  font-size: 0.9rem;
}
.center-state.error { color: #ef4444; }

.spinner {
  width: 30px; height: 30px;
  border: 3px solid var(--border-color);
  border-top-color: var(--primary-color);
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}
.spinner-sm {
  display: inline-block;
  width: 13px; height: 13px;
  border: 2px solid rgba(255,255,255,0.35);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

/* ── Empty state ─────────────────────────────────────────────────────── */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 14px;
  padding: 80px 0;
  text-align: center;
}

.empty-visual {
  position: relative;
  width: 96px;
  height: 96px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 4px;
}

.empty-circle {
  position: absolute;
  border-radius: 50%;
  opacity: 0.12;
}
.empty-circle.c1 { width: 96px; height: 96px; background: var(--primary-color); }
.empty-circle.c2 { width: 68px; height: 68px; background: var(--primary-color); opacity: 0.18; }
.empty-circle.c3 { width: 44px; height: 44px; background: var(--primary-color); opacity: 0.25; }

.empty-icon { color: var(--primary-color); position: relative; z-index: 1; }

.empty-state h3 {
  font-size: 1.2rem;
  font-weight: 700;
  color: var(--main-text-color);
  margin: 0;
}
.empty-state p {
  font-size: 0.875rem;
  color: var(--secondary-text-color);
  margin: 0;
  max-width: 340px;
  line-height: 1.6;
}

/* ── Orgs grid ───────────────────────────────────────────────────────── */
.orgs-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
}

/* ── Org card ────────────────────────────────────────────────────────── */
.org-card {
  position: relative;
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 14px;
  overflow: hidden;
  cursor: pointer;
  transition: transform 0.18s, box-shadow 0.18s, border-color 0.18s;
  outline: none;
}

.org-card:hover,
.org-card:focus-visible {
  transform: translateY(-3px);
  box-shadow: 0 8px 28px rgba(0,0,0,0.2);
  border-color: color-mix(in srgb, var(--primary-color) 40%, var(--border-color));
}

.card-accent {
  height: 4px;
  width: 100%;
}

.card-body {
  padding: 18px 20px 16px;
}

/* ── Identity ──────────────────────────────────────────────────────── */
.card-identity {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 14px;
}

.org-avatar {
  width: 46px;
  height: 46px;
  border-radius: 12px;
  color: #fff;
  font-size: 1.25rem;
  font-weight: 800;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  overflow: hidden;
  letter-spacing: -0.02em;
}

.org-avatar-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.org-meta { flex: 1; min-width: 0; }

.org-name {
  font-size: 0.975rem;
  font-weight: 700;
  color: var(--main-text-color);
  margin: 0 0 3px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.org-desc {
  font-size: 0.775rem;
  color: var(--secondary-text-color);
  margin: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-style: italic;
}

/* ── Role badge ────────────────────────────────────────────────────── */
.role-badge {
  font-size: 0.68rem;
  font-weight: 700;
  padding: 3px 9px;
  border-radius: 20px;
  flex-shrink: 0;
  align-self: flex-start;
  letter-spacing: 0.02em;
  text-transform: uppercase;
}
.role-badge.owner  { background: rgba(245,158,11,0.14); color: #f59e0b; }
.role-badge.admin  { background: rgba(99,102,241,0.14);  color: #818cf8; }
.role-badge.member { background: rgba(34,197,94,0.14);   color: #4ade80; }
.role-badge.viewer { background: rgba(107,114,128,0.12); color: #9ca3af; }

/* ── Stats row ─────────────────────────────────────────────────────── */
.card-stats {
  display: flex;
  gap: 16px;
  margin-bottom: 14px;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 0.775rem;
  color: var(--secondary-text-color);
}

/* ── Storage ───────────────────────────────────────────────────────── */
.storage-section { }

.storage-labels {
  display: flex;
  justify-content: space-between;
  font-size: 0.75rem;
  color: var(--secondary-text-color);
  margin-bottom: 5px;
}

.text-danger { color: #ef4444 !important; font-weight: 600; }
.text-warn   { color: #f59e0b !important; font-weight: 600; }

.storage-track {
  height: 4px;
  background: var(--border-color);
  border-radius: 4px;
  overflow: hidden;
  margin-bottom: 4px;
}

.storage-fill {
  height: 100%;
  border-radius: 4px;
  transition: width 0.4s ease;
}

.storage-quota {
  font-size: 0.72rem;
  color: var(--secondary-text-color);
  opacity: 0.7;
}

/* ── Pin button ────────────────────────────────────────────────────── */
.btn-pin {
  position: absolute;
  top: 10px;
  right: 10px;
  padding: 0;
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 7px;
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: var(--secondary-text-color);
  opacity: 0;
  transform: scale(0.85);
  transition: opacity 0.15s, transform 0.15s, color 0.15s, background 0.15s;
  z-index: 2;
}

.org-card:hover .btn-pin,
.org-card:focus-visible .btn-pin,
.btn-pin.pinned {
  opacity: 1;
  transform: scale(1);
}

.btn-pin.pinned {
  color: var(--primary-color);
  background: color-mix(in srgb, var(--primary-color) 10%, var(--card-color));
  border-color: color-mix(in srgb, var(--primary-color) 30%, var(--border-color));
}

.btn-pin:hover {
  color: var(--primary-color);
  background: color-mix(in srgb, var(--primary-color) 10%, var(--card-color));
}

/* ── Pinned strip ──────────────────────────────────────────────────── */
.pinned-strip {
  margin-bottom: 24px;
}

.pinned-strip-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.72rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.07em;
  color: var(--secondary-text-color);
  margin: 0 0 10px;
}

.pinned-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.pinned-chip {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 10px;
  padding: 6px 14px 6px 10px;
  font-size: 0.82rem;
  font-weight: 600;
  color: var(--main-text-color);
  cursor: pointer;
  transition: background 0.15s, border-color 0.15s, transform 0.1s;
}

.pinned-chip:hover {
  background: var(--hover-background-color);
  border-color: color-mix(in srgb, var(--primary-color) 35%, var(--border-color));
  transform: translateY(-1px);
}

.pinned-chip-dot {
  width: 10px;
  height: 10px;
  border-radius: 3px;
  flex-shrink: 0;
}

/* ── Card arrow ────────────────────────────────────────────────────── */
.card-arrow {
  position: absolute;
  bottom: 16px;
  right: 16px;
  color: var(--secondary-text-color);
  opacity: 0;
  transform: translateX(-4px);
  transition: opacity 0.18s, transform 0.18s;
}

.org-card:hover .card-arrow,
.org-card:focus-visible .card-arrow {
  opacity: 0.6;
  transform: translateX(0);
}

/* ── Paywall ─────────────────────────────────────────────────────────── */
.paywall-wrapper {
  display: flex;
  justify-content: center;
  padding: 40px 16px;
}

.paywall-card {
  position: relative;
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 20px;
  padding: 44px 48px;
  max-width: 500px;
  width: 100%;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  overflow: hidden;
}

.paywall-glow {
  position: absolute;
  top: -60px;
  left: 50%;
  transform: translateX(-50%);
  width: 280px;
  height: 180px;
  background: radial-gradient(ellipse, color-mix(in srgb, var(--primary-color) 20%, transparent), transparent 70%);
  pointer-events: none;
}

.paywall-icon-wrap {
  width: 68px;
  height: 68px;
  border-radius: 18px;
  background: color-mix(in srgb, var(--primary-color) 13%, transparent);
  color: var(--primary-color);
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid color-mix(in srgb, var(--primary-color) 25%, transparent);
  position: relative;
  z-index: 1;
}

.paywall-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  background: color-mix(in srgb, var(--primary-color) 13%, transparent);
  color: var(--primary-color);
  font-size: 0.72rem;
  font-weight: 700;
  padding: 4px 12px;
  border-radius: 20px;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}

.paywall-title {
  font-size: 1.4rem;
  font-weight: 800;
  color: var(--main-text-color);
  margin: 0;
  letter-spacing: -0.02em;
}

.paywall-desc {
  font-size: 0.875rem;
  color: var(--secondary-text-color);
  margin: 0;
  line-height: 1.65;
  max-width: 360px;
}

.paywall-features {
  list-style: none;
  padding: 0;
  margin: 0;
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 9px;
  text-align: left;
  background: var(--background-color);
  border-radius: 12px;
  padding: 16px 18px;
}

.paywall-features li {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 0.85rem;
  color: var(--main-text-color);
}
.paywall-features li svg { color: var(--primary-color); flex-shrink: 0; }

.btn-upgrade {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: var(--primary-color);
  color: #fff;
  border: none;
  border-radius: 10px;
  padding: 13px 32px;
  font-size: 0.925rem;
  font-weight: 700;
  cursor: pointer;
  transition: filter 0.15s, transform 0.1s;
  width: 100%;
  justify-content: center;
}
.btn-upgrade:hover { filter: brightness(1.1); }
.btn-upgrade:active { transform: scale(0.98); }

.paywall-note {
  font-size: 0.75rem;
  color: var(--secondary-text-color);
  margin: 0;
  line-height: 1.5;
}

/* ── Modal ───────────────────────────────────────────────────────────── */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.55);
  backdrop-filter: blur(3px);
  z-index: 2000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
}

.modal {
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 16px;
  box-shadow: 0 24px 64px rgba(0,0,0,0.3);
  width: 100%;
  max-width: 480px;
  overflow: hidden;
}

.modal-header {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 22px 24px 18px;
  border-bottom: 1px solid var(--border-color);
}

.modal-header-icon {
  width: 38px; height: 38px;
  border-radius: 10px;
  background: color-mix(in srgb, var(--primary-color) 13%, transparent);
  color: var(--primary-color);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.modal-header h3 {
  margin: 0 0 2px;
  font-size: 1rem;
  font-weight: 700;
  color: var(--main-text-color);
}

.modal-subtitle {
  margin: 0;
  font-size: 0.78rem;
  color: var(--secondary-text-color);
}

.btn-close {
  margin-left: auto;
  background: none;
  border: none;
  cursor: pointer;
  color: var(--secondary-text-color);
  padding: 4px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  transition: background 0.15s, color 0.15s;
  flex-shrink: 0;
}
.btn-close:hover { background: var(--hover-background-color); color: var(--main-text-color); }

.modal-body { padding: 20px 24px; }

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 16px 24px 20px;
  border-top: 1px solid var(--border-color);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 16px;
}
.form-group:last-of-type { margin-bottom: 0; }

.form-group label {
  font-size: 0.82rem;
  font-weight: 600;
  color: var(--secondary-text-color);
}

.required { color: var(--primary-color); }

.input-field {
  background: var(--background-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 10px 12px;
  font-size: 0.875rem;
  color: var(--main-text-color);
  transition: border-color 0.15s, box-shadow 0.15s;
  width: 100%;
  box-sizing: border-box;
}
.input-field:focus {
  outline: none;
  border-color: var(--primary-color);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--primary-color) 15%, transparent);
}

.quota-input-wrap { position: relative; }
.quota-input-wrap .input-field { padding-right: 44px; }
.quota-unit {
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 0.8rem;
  color: var(--secondary-text-color);
  pointer-events: none;
}

.field-hint {
  font-size: 0.75rem;
  color: var(--secondary-text-color);
  margin: 2px 0 0;
}

.form-error {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #ef4444;
  font-size: 0.82rem;
  margin: 10px 0 0;
  padding: 8px 12px;
  background: rgba(239,68,68,0.08);
  border-radius: 6px;
  border: 1px solid rgba(239,68,68,0.2);
}

.btn-secondary {
  background: transparent;
  color: var(--main-text-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 9px 18px;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.15s;
}
.btn-secondary:hover { background: var(--hover-background-color); }

/* ── Modal animation ─────────────────────────────────────────────────── */
.modal-enter-active, .modal-leave-active { transition: opacity 0.2s; }
.modal-enter-active .modal, .modal-leave-active .modal { transition: transform 0.2s; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
.modal-enter-from .modal, .modal-leave-to .modal { transform: scale(0.96) translateY(10px); }

/* ── Responsive ──────────────────────────────────────────────────────── */
@media (max-width: 768px) {
  .orgs-page { padding: 20px 16px 40px; }
  .page-header-text h1 { font-size: 1.4rem; }
  .orgs-grid { grid-template-columns: 1fr; gap: 14px; }
  .paywall-card { padding: 32px 20px; }
  .modal-header { flex-wrap: wrap; }
}
</style>
