<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="orgs-container">
    <div class="header">
      <h2>{{ t('orgs.title') }}</h2>
      <button class="btn-primary" @click="showCreateModal = true">
        <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor"><path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z"/></svg>
        {{ t('orgs.createOrg') }}
      </button>
    </div>

    <div v-if="orgStore.loading" class="loading-state">
      <div class="spinner"></div>
      <span>{{ t('common.loading') }}</span>
    </div>

    <div v-else-if="orgStore.error" class="error-state">
      <svg viewBox="0 0 24 24" width="40" height="40" fill="currentColor" class="error-icon"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/></svg>
      <p>{{ orgStore.error }}</p>
    </div>

    <div v-else-if="orgStore.orgs.length === 0" class="empty-state">
      <div class="empty-icon">
        <svg viewBox="0 0 24 24" width="64" height="64" fill="currentColor"><path d="M12 7V3H2v18h20V7H12zM6 19H4v-2h2v2zm0-4H4v-2h2v2zm0-4H4V9h2v2zm0-4H4V5h2v2zm4 12H8v-2h2v2zm0-4H8v-2h2v2zm0-4H8V9h2v2zm0-4H8V5h2v2zm10 12h-8v-2h2v-2h-2v-2h2v-2h-2V9h8v10zm-2-8h-2v2h2v-2zm0 4h-2v2h2v-2z"/></svg>
      </div>
      <h3>{{ t('orgs.noOrgs') }}</h3>
      <p>{{ t('orgs.noOrgsDesc') }}</p>
      <button class="btn-primary" @click="showCreateModal = true">{{ t('orgs.createOrg') }}</button>
    </div>

    <div v-else class="orgs-grid">
      <div
        v-for="org in orgStore.orgs"
        :key="org.id"
        class="org-card"
        @click="router.push(`/dashboard/organizations/${org.id}`)"
      >
        <div class="org-card-header">
          <div class="org-avatar">{{ org.name.charAt(0).toUpperCase() }}</div>
          <div class="org-info">
            <h3 class="org-name">{{ org.name }}</h3>
            <p class="org-desc">{{ org.description || '—' }}</p>
          </div>
          <div class="role-badge" :class="org.my_role">{{ t(`orgs.${org.my_role}`) }}</div>
        </div>

        <div class="org-storage">
          <div class="storage-row">
            <span class="storage-label">{{ t('orgs.storageUsed') }}</span>
            <span class="storage-value">{{ formatSize(org.storage_used_bytes) }} / {{ formatSize(org.storage_quota_mb * 1024 * 1024) }}</span>
          </div>
          <div class="storage-bar">
            <div class="storage-fill" :style="{ width: storagePercent(org) + '%', background: storageColor(org) }"></div>
          </div>
        </div>
      </div>
    </div>

    <!-- Create org modal -->
    <Transition name="modal">
      <div v-if="showCreateModal" class="modal-overlay" @click.self="showCreateModal = false">
        <div class="modal">
          <div class="modal-header">
            <h3>{{ t('orgs.createOrg') }}</h3>
            <button class="btn-close" @click="showCreateModal = false">
              <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
            </button>
          </div>
          <div class="modal-body">
            <div class="form-group">
              <label>{{ t('orgs.orgName') }} *</label>
              <input v-model="form.name" type="text" :placeholder="t('orgs.orgName')" class="input-field" />
            </div>
            <div class="form-group">
              <label>{{ t('orgs.orgDesc') }}</label>
              <input v-model="form.description" type="text" :placeholder="t('orgs.orgDesc')" class="input-field" />
            </div>
            <div class="form-group">
              <label>{{ t('orgs.storageQuotaMB') }}</label>
              <input v-model.number="form.storageQuotaMB" type="number" min="100" class="input-field" />
            </div>
            <p v-if="createError" class="form-error">{{ createError }}</p>
          </div>
          <div class="modal-footer">
            <button class="btn-secondary" @click="showCreateModal = false">{{ t('orgs.cancel') }}</button>
            <button class="btn-primary" @click="handleCreate" :disabled="creating || !form.name">
              <span v-if="creating" class="spinner-sm"></span>
              {{ creating ? t('common.loading') : t('orgs.create') }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useOrgStore } from '../stores/organizations'

const { t } = useI18n()
const router = useRouter()
const orgStore = useOrgStore()

const showCreateModal = ref(false)
const creating = ref(false)
const createError = ref('')

const form = ref({
  name: '',
  description: '',
  storageQuotaMB: 10240,
})

onMounted(() => orgStore.fetchOrgs())

const handleCreate = async () => {
  if (!form.value.name) return
  creating.value = true
  createError.value = ''
  try {
    const org = await orgStore.createOrg(form.value.name, form.value.description, form.value.storageQuotaMB)
    showCreateModal.value = false
    form.value = { name: '', description: '', storageQuotaMB: 10240 }
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
</script>

<style scoped>
.orgs-container {
  padding: 24px;
  max-width: 1200px;
  margin: 0 auto;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
}

.header h2 {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--main-text-color);
  margin: 0;
}

.btn-primary {
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--primary-color);
  color: white;
  border: none;
  border-radius: 8px;
  padding: 10px 18px;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.2s, transform 0.1s;
}

.btn-primary:hover { opacity: 0.9; }
.btn-primary:active { transform: scale(0.98); }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }

.btn-secondary {
  background: var(--card-color);
  color: var(--main-text-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 10px 18px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.2s;
}

.btn-secondary:hover { background: var(--hover-background-color); }

.loading-state, .error-state, .empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  padding: 80px 24px;
  text-align: center;
  color: var(--secondary-text-color);
}

.empty-icon {
  color: var(--border-color);
  opacity: 0.6;
}

.empty-state h3 {
  font-size: 1.2rem;
  font-weight: 600;
  color: var(--main-text-color);
  margin: 0;
}

.empty-state p {
  margin: 0;
  font-size: 0.9rem;
}

.error-icon { color: #ef4444; }

.spinner {
  width: 32px;
  height: 32px;
  border: 3px solid var(--border-color);
  border-top-color: var(--primary-color);
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

.spinner-sm {
  display: inline-block;
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255,255,255,0.4);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

@keyframes spin { to { transform: rotate(360deg); } }

.orgs-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 16px;
}

.org-card {
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 20px;
  cursor: pointer;
  transition: transform 0.15s, box-shadow 0.15s;
}

.org-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 16px rgba(0,0,0,0.1);
}

.org-card-header {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 16px;
}

.org-avatar {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  background: var(--primary-color);
  color: white;
  font-size: 1.3rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.org-info { flex: 1; min-width: 0; }

.org-name {
  font-size: 1rem;
  font-weight: 700;
  color: var(--main-text-color);
  margin: 0 0 4px 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.org-desc {
  font-size: 0.8rem;
  color: var(--secondary-text-color);
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.role-badge {
  font-size: 0.7rem;
  font-weight: 600;
  padding: 3px 8px;
  border-radius: 20px;
  flex-shrink: 0;
}

.role-badge.owner { background: rgba(245, 158, 11, 0.15); color: #f59e0b; }
.role-badge.admin { background: rgba(99, 102, 241, 0.15); color: #6366f1; }
.role-badge.member { background: rgba(34, 197, 94, 0.15); color: #22c55e; }
.role-badge.viewer { background: rgba(107, 114, 128, 0.12); color: #6b7280; }

.org-storage { margin-top: 4px; }

.storage-row {
  display: flex;
  justify-content: space-between;
  font-size: 0.78rem;
  color: var(--secondary-text-color);
  margin-bottom: 6px;
}

.storage-bar {
  height: 5px;
  background: var(--border-color);
  border-radius: 3px;
  overflow: hidden;
}

.storage-fill {
  height: 100%;
  border-radius: 3px;
  transition: width 0.3s ease;
}

/* Modal */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.5);
  z-index: 2000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
}

.modal {
  background: var(--card-color);
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0,0,0,0.25);
  width: 100%;
  max-width: 480px;
  overflow: hidden;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px 16px;
  border-bottom: 1px solid var(--border-color);
}

.modal-header h3 {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--main-text-color);
}

.btn-close {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--secondary-text-color);
  padding: 4px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  transition: background 0.15s;
}

.btn-close:hover { background: var(--hover-background-color); }

.modal-body { padding: 20px 24px; }

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 24px 20px;
  border-top: 1px solid var(--border-color);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 16px;
}

.form-group label {
  font-size: 0.85rem;
  font-weight: 500;
  color: var(--secondary-text-color);
}

.input-field {
  background: var(--background-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 10px 12px;
  font-size: 0.9rem;
  color: var(--main-text-color);
  transition: border-color 0.15s;
  width: 100%;
  box-sizing: border-box;
}

.input-field:focus {
  outline: none;
  border-color: var(--primary-color);
}

.form-error {
  color: #ef4444;
  font-size: 0.82rem;
  margin: 0;
}

.modal-enter-active, .modal-leave-active { transition: opacity 0.2s; }
.modal-enter-active .modal, .modal-leave-active .modal { transition: transform 0.2s; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
.modal-enter-from .modal { transform: scale(0.95) translateY(8px); }
.modal-leave-to .modal { transform: scale(0.95) translateY(8px); }

@media (max-width: 768px) {
  .orgs-container { padding: 16px; }
  .orgs-grid { grid-template-columns: 1fr; }
  .header { flex-wrap: wrap; gap: 12px; }
}
</style>
