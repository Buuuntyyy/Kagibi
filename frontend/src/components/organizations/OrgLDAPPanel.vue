<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="ldap-panel">
    <div class="panel-header">
      <h3 class="panel-title">Annuaire LDAP / Active Directory</h3>
      <span class="badge-ldap" v-if="config.enabled">Actif</span>
      <span class="badge-disabled" v-else>Désactivé</span>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="loading-state"><div class="spinner"></div></div>

    <template v-else>
      <!-- Status card -->
      <div class="status-card" v-if="config.last_sync_at">
        <div class="status-row">
          <span class="status-label">Dernière synchronisation</span>
          <span class="status-value">{{ formatDate(config.last_sync_at) }}</span>
        </div>
        <div class="status-row" v-if="config.last_sync_error">
          <span class="status-label error">Erreur</span>
          <span class="status-value error">{{ config.last_sync_error }}</span>
        </div>
        <div class="status-row" v-if="config.last_sync_stats">
          <span class="status-label">Stats</span>
          <span class="status-value stats">
            {{ config.last_sync_stats.users_found }} utilisateurs trouvés ·
            {{ config.last_sync_stats.users_invited }} invités ·
            {{ config.last_sync_stats.users_suspended }} suspendus ·
            {{ config.last_sync_stats.users_deleted }} supprimés ·
            {{ config.last_sync_stats.groups_found }} groupes ·
            {{ config.last_sync_stats.duration_ms }}ms
          </span>
        </div>
        <div class="status-actions">
          <button class="btn-secondary" @click="triggerSync" :disabled="syncing">
            <span v-if="syncing" class="spinner-sm"></span>
            <svg v-else viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M12 4V1L8 5l4 4V6c3.31 0 6 2.69 6 6 0 1.01-.25 1.97-.7 2.8l1.46 1.46C19.54 15.03 20 13.57 20 12c0-4.42-3.58-8-8-8zm0 14c-3.31 0-6-2.69-6-6 0-1.01.25-1.97.7-2.8L5.24 7.74C4.46 8.97 4 10.43 4 12c0 4.42 3.58 8 8 8v3l4-4-4-4v3z"/></svg>
            Synchroniser maintenant
          </button>
          <button class="btn-secondary" @click="loadSuspended">
            Membres suspendus ({{ suspendedCount }})
          </button>
        </div>
      </div>

      <!-- Suspended members -->
      <div v-if="showSuspended && suspended.length > 0" class="suspended-list">
        <h4 class="section-label">Membres suspendus en attente de suppression</h4>
        <div v-for="m in suspended" :key="m.id" class="suspended-row">
          <span class="suspended-uid">{{ m.ldap_uid }}</span>
          <span class="suspended-date">Suspendu le {{ formatDate(m.suspended_at) }}</span>
        </div>
      </div>

      <!-- Config form -->
      <form class="ldap-form" @submit.prevent="save">
        <div class="form-section">
          <h4 class="section-label">Connexion</h4>

          <label class="form-row toggle-row">
            <span>Synchronisation LDAP activée</span>
            <label class="toggle">
              <input type="checkbox" v-model="form.enabled" />
              <span class="toggle-slider"></span>
            </label>
          </label>

          <label class="form-row">
            <span>URL du serveur</span>
            <input class="form-input" v-model="form.url" placeholder="ldap://ldap.example.com ou ldaps://..." required />
          </label>

          <label class="form-row">
            <span>Bind DN</span>
            <input class="form-input" v-model="form.bind_dn" placeholder="cn=admin,dc=example,dc=com" />
          </label>

          <label class="form-row">
            <span>Mot de passe Bind <span class="hint">(laisser vide pour conserver l'existant)</span></span>
            <input class="form-input" type="password" v-model="form.bind_password" autocomplete="new-password" />
          </label>

          <label class="form-row toggle-row">
            <span>Ignorer la vérification TLS <span class="hint">(déconseillé en production)</span></span>
            <label class="toggle">
              <input type="checkbox" v-model="form.tls_skip_verify" />
              <span class="toggle-slider"></span>
            </label>
          </label>
        </div>

        <div class="form-section">
          <h4 class="section-label">Utilisateurs</h4>

          <label class="form-row">
            <span>Base DN utilisateurs</span>
            <input class="form-input" v-model="form.user_base_dn" placeholder="ou=users,dc=example,dc=com" />
          </label>

          <label class="form-row">
            <span>Filtre utilisateurs</span>
            <input class="form-input" v-model="form.user_filter" placeholder="(objectClass=person)" />
          </label>

          <label class="form-row">
            <span>Attribut e-mail</span>
            <input class="form-input" v-model="form.attr_email" placeholder="mail" />
          </label>

          <label class="form-row">
            <span>Attribut nom d'affichage</span>
            <input class="form-input" v-model="form.attr_display_name" placeholder="cn" />
          </label>

          <label class="form-row">
            <span>Attribut UID unique</span>
            <input class="form-input" v-model="form.attr_uid" placeholder="uid" />
          </label>
        </div>

        <div class="form-section">
          <h4 class="section-label">Groupes <span class="hint">(optionnel)</span></h4>

          <label class="form-row">
            <span>Base DN groupes</span>
            <input class="form-input" v-model="form.group_base_dn" placeholder="ou=groups,dc=example,dc=com" />
          </label>

          <label class="form-row">
            <span>Filtre groupes</span>
            <input class="form-input" v-model="form.group_filter" placeholder="(objectClass=groupOfNames)" />
          </label>
        </div>

        <div class="form-section">
          <h4 class="section-label">Planification</h4>

          <label class="form-row">
            <span>Intervalle de sync (minutes)</span>
            <input class="form-input narrow" type="number" min="5" v-model.number="form.sync_interval_minutes" />
          </label>

          <label class="form-row">
            <span>Jours avant suppression automatique <span class="hint">(0 = manuel uniquement)</span></span>
            <input class="form-input narrow" type="number" min="0" v-model.number="form.auto_deprovision_days" />
          </label>

          <label class="form-row">
            <span>Nombre minimal d'utilisateurs attendus <span class="hint">(protection anti-vidage)</span></span>
            <input class="form-input narrow" type="number" min="1" v-model.number="form.min_expected_users" />
          </label>
        </div>

        <p v-if="error" class="msg-error">{{ error }}</p>
        <p v-if="success" class="msg-success">{{ success }}</p>

        <div class="form-actions">
          <button type="button" class="btn-secondary" @click="testConnection" :disabled="testing">
            <span v-if="testing" class="spinner-sm"></span>
            Tester la connexion
          </button>
          <button type="submit" class="btn-primary" :disabled="saving">
            <span v-if="saving" class="spinner-sm"></span>
            Enregistrer
          </button>
        </div>

        <div v-if="testResult" class="test-result" :class="testResult.success ? 'success' : 'error'">
          <template v-if="testResult.success">
            Connexion réussie · {{ testResult.users_found }} utilisateur(s) trouvé(s)
          </template>
          <template v-else>
            {{ testResult.error }}
          </template>
        </div>
      </form>
    </template>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '../../api'

const props = defineProps({ orgID: { type: [Number, String], required: true } })

const loading = ref(true)
const saving  = ref(false)
const testing = ref(false)
const syncing = ref(false)
const error   = ref('')
const success = ref('')
const testResult = ref(null)
const showSuspended = ref(false)
const suspended = ref([])
const suspendedCount = ref(0)
const config = ref({})

const form = ref({
  enabled: false,
  url: '',
  bind_dn: '',
  bind_password: '',
  user_base_dn: '',
  user_filter: '(objectClass=person)',
  group_base_dn: '',
  group_filter: '(objectClass=groupOfNames)',
  attr_email: 'mail',
  attr_display_name: 'cn',
  attr_uid: 'uid',
  tls_skip_verify: false,
  sync_interval_minutes: 60,
  auto_deprovision_days: 30,
  min_expected_users: 1,
})

onMounted(load)

async function load() {
  loading.value = true
  try {
    const { data } = await api.get(`/orgs/${props.orgID}/ldap`)
    config.value = data
    Object.assign(form.value, {
      enabled: data.enabled,
      url: data.url,
      bind_dn: data.bind_dn,
      bind_password: '',
      user_base_dn: data.user_base_dn,
      user_filter: data.user_filter || '(objectClass=person)',
      group_base_dn: data.group_base_dn,
      group_filter: data.group_filter || '(objectClass=groupOfNames)',
      attr_email: data.attr_email || 'mail',
      attr_display_name: data.attr_display_name || 'cn',
      attr_uid: data.attr_uid || 'uid',
      tls_skip_verify: data.tls_skip_verify,
      sync_interval_minutes: data.sync_interval_minutes || 60,
      auto_deprovision_days: data.auto_deprovision_days ?? 30,
      min_expected_users: data.min_expected_users || 1,
    })
  } catch (e) {
    error.value = e.response?.data?.error ?? e.message
  } finally {
    loading.value = false
  }
}

async function save() {
  saving.value = true
  error.value = ''
  success.value = ''
  testResult.value = null
  try {
    const { data } = await api.put(`/orgs/${props.orgID}/ldap`, form.value)
    config.value = data
    form.value.bind_password = ''
    success.value = 'Configuration sauvegardée.'
    setTimeout(() => { success.value = '' }, 4000)
  } catch (e) {
    error.value = e.response?.data?.error ?? e.message
  } finally {
    saving.value = false
  }
}

async function testConnection() {
  testing.value = true
  testResult.value = null
  error.value = ''
  try {
    const { data } = await api.post(`/orgs/${props.orgID}/ldap/test`)
    testResult.value = { success: true, users_found: data.users_found }
  } catch (e) {
    testResult.value = { success: false, error: e.response?.data?.error ?? e.message }
  } finally {
    testing.value = false
  }
}

async function triggerSync() {
  syncing.value = true
  error.value = ''
  success.value = ''
  try {
    const { data } = await api.post(`/orgs/${props.orgID}/ldap/sync`)
    config.value.last_sync_stats = data.stats
    config.value.last_sync_at = new Date().toISOString()
    config.value.last_sync_error = ''
    success.value = 'Synchronisation terminée.'
    setTimeout(() => { success.value = '' }, 4000)
  } catch (e) {
    error.value = e.response?.data?.error ?? e.message
  } finally {
    syncing.value = false
  }
}

async function loadSuspended() {
  showSuspended.value = !showSuspended.value
  if (!showSuspended.value) return
  try {
    const { data } = await api.get(`/orgs/${props.orgID}/ldap/suspended`)
    suspended.value = data.members ?? []
    suspendedCount.value = suspended.value.length
  } catch (e) {
    error.value = e.response?.data?.error ?? e.message
  }
}

function formatDate(iso) {
  if (!iso) return '—'
  return new Date(iso).toLocaleString()
}
</script>

<style scoped>
.ldap-panel { padding: 24px; max-width: 720px; }

.panel-header { display: flex; align-items: center; gap: 12px; margin-bottom: 24px; }
.panel-title  { font-size: 1.1rem; font-weight: 600; margin: 0; }

.badge-ldap     { background: #2563eb; color: #fff; border-radius: 4px; padding: 2px 8px; font-size: .75rem; font-weight: 600; }
.badge-disabled { background: #6b7280; color: #fff; border-radius: 4px; padding: 2px 8px; font-size: .75rem; }

.status-card {
  background: var(--color-surface-raised, #f8fafc);
  border: 1px solid var(--color-border, #e5e7eb);
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 24px;
}
.status-row { display: flex; gap: 12px; margin-bottom: 6px; font-size: .9rem; }
.status-label { font-weight: 500; min-width: 140px; color: var(--color-text-muted, #6b7280); }
.status-label.error { color: #ef4444; }
.status-value.error { color: #ef4444; }
.status-value.stats { color: var(--color-text-muted, #6b7280); font-size: .85rem; }
.status-actions { display: flex; gap: 8px; margin-top: 12px; flex-wrap: wrap; }

.suspended-list { margin-bottom: 24px; }
.suspended-row  { display: flex; gap: 16px; padding: 8px 0; border-bottom: 1px solid var(--color-border, #e5e7eb); font-size: .9rem; }
.suspended-uid  { font-weight: 500; }
.suspended-date { color: var(--color-text-muted, #6b7280); }

.ldap-form   { display: flex; flex-direction: column; gap: 0; }
.form-section { margin-bottom: 28px; }
.section-label { font-size: .8rem; font-weight: 600; text-transform: uppercase; letter-spacing: .06em; color: var(--color-text-muted, #6b7280); margin: 0 0 14px; }
.hint { font-weight: 400; text-transform: none; letter-spacing: 0; font-size: .8rem; }

.form-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
  font-size: .9rem;
}
.form-row > span:first-child { min-width: 240px; color: var(--color-text, #111); }
.toggle-row > span:first-child { flex: 1; }

.form-input {
  flex: 1;
  border: 1px solid var(--color-border, #d1d5db);
  border-radius: 6px;
  padding: 7px 10px;
  font-size: .9rem;
  background: var(--color-input-bg, #fff);
  color: var(--color-text, #111);
}
.form-input.narrow { max-width: 100px; flex: none; }
.form-input:focus  { outline: none; border-color: var(--color-primary, #2563eb); box-shadow: 0 0 0 2px rgba(37,99,235,.15); }

.form-actions {
  display: flex;
  gap: 10px;
  margin-top: 8px;
  justify-content: flex-end;
}

.test-result {
  margin-top: 10px;
  padding: 10px 14px;
  border-radius: 6px;
  font-size: .9rem;
}
.test-result.success { background: #d1fae5; color: #065f46; }
.test-result.error   { background: #fee2e2; color: #991b1b; }

.msg-error   { color: #ef4444; font-size: .9rem; margin: 0 0 8px; }
.msg-success { color: #16a34a; font-size: .9rem; margin: 0 0 8px; }

.loading-state { display: flex; justify-content: center; padding: 40px; }

/* Re-use toggle styles from the app's global stylesheet */
.toggle { position: relative; display: inline-flex; align-items: center; cursor: pointer; }
.toggle input { opacity: 0; width: 0; height: 0; }
.toggle-slider {
  width: 36px; height: 20px; background: #d1d5db; border-radius: 20px;
  transition: background .2s;
  position: relative;
}
.toggle-slider::after {
  content: ''; position: absolute; top: 2px; left: 2px;
  width: 16px; height: 16px; border-radius: 50%; background: #fff;
  transition: transform .2s;
}
.toggle input:checked + .toggle-slider { background: #2563eb; }
.toggle input:checked + .toggle-slider::after { transform: translateX(16px); }

.btn-primary, .btn-secondary {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 7px 16px; border-radius: 6px; font-size: .9rem;
  cursor: pointer; border: none; font-weight: 500; transition: background .15s;
}
.btn-primary  { background: var(--color-primary, #2563eb); color: #fff; }
.btn-primary:hover  { background: #1d4ed8; }
.btn-primary:disabled, .btn-secondary:disabled { opacity: .6; cursor: not-allowed; }
.btn-secondary { background: var(--color-surface-raised, #f3f4f6); color: var(--color-text, #111); border: 1px solid var(--color-border, #d1d5db); }
.btn-secondary:hover { background: #e5e7eb; }

.spinner { width: 24px; height: 24px; border: 2px solid var(--color-border, #e5e7eb); border-top-color: var(--color-primary, #2563eb); border-radius: 50%; animation: spin .7s linear infinite; }
.spinner-sm { width: 14px; height: 14px; border: 2px solid rgba(255,255,255,.4); border-top-color: #fff; border-radius: 50%; animation: spin .7s linear infinite; display: inline-block; }
@keyframes spin { to { transform: rotate(360deg); } }
</style>
