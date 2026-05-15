<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <Transition name="modal">
    <div v-if="modelValue" class="modal-overlay" @click.self="close">
      <div class="access-dialog">

        <!-- Header -->
        <div class="dialog-header">
          <div class="dialog-title">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor" class="dialog-icon">
              <path d="M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zm-6 9c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2zm3.1-9H8.9V6c0-1.71 1.39-3.1 3.1-3.1 1.71 0 3.1 1.39 3.1 3.1v2z"/>
            </svg>
            <div>
              <h3 class="dialog-heading">{{ t('orgs.manageAccess') }}</h3>
              <p class="dialog-subtitle">{{ folderDisplayName }}</p>
            </div>
          </div>
          <button class="btn-close" @click="close">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
          </button>
        </div>

        <!-- Body -->
        <div class="dialog-body">
          <div v-if="loading" class="dialog-loading">
            <div class="spinner"></div>
          </div>
          <template v-else>

            <!-- ── Users ─────────────────────────────────────────────────── -->
            <div v-if="canManage" class="access-section">
              <div class="access-section-header">
                <span class="access-section-title">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z"/></svg>
                  {{ t('orgs.members') }}
                </span>
                <button class="btn-add-access" @click="showUserPicker = !showUserPicker">
                  <svg viewBox="0 0 24 24" width="13" height="13" fill="currentColor"><path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z"/></svg>
                  {{ t('orgs.addUser') }}
                </button>
              </div>

              <!-- User picker -->
              <div v-if="showUserPicker" class="add-access-form">
                <select v-model="newUserID" class="access-select">
                  <option value="">— {{ t('orgs.members') }} —</option>
                  <option v-for="m in availableMembers" :key="m.user_id" :value="m.user_id">
                    {{ m.name || m.email }}
                  </option>
                </select>
                <select v-model="newUserLevel" class="access-select level-select">
                  <option value="read">{{ t('orgs.permRead') }}</option>
                  <option value="write">{{ t('orgs.permWrite') }}</option>
                  <option value="manage">{{ t('orgs.permManage') }}</option>
                  <option value="none">{{ t('orgs.permNone') }}</option>
                </select>
                <button class="btn-confirm" @click="addUserPerm" :disabled="!newUserID || savingUser">
                  <span v-if="savingUser" class="spinner-sm"></span>
                  <svg v-else viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41L9 16.17z"/></svg>
                </button>
                <button class="btn-cancel-add" @click="showUserPicker = false; newUserID = ''">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
                </button>
              </div>

              <!-- User rows -->
              <div v-if="userPerms.length === 0" class="access-empty">{{ t('orgs.noUserAccessOnFolder') }}</div>
              <div v-else class="access-rows">
                <div v-for="up in userPerms" :key="up.user_id" class="access-row">
                  <div class="access-avatar">{{ memberInitial(up.user_id) }}</div>
                  <span class="access-name">{{ memberName(up.user_id) }}</span>
                  <select
                    class="access-select level-select"
                    :value="up.level"
                    @change="updateUserPerm(up, $event.target.value)"
                  >
                    <option value="read">{{ t('orgs.permRead') }}</option>
                    <option value="write">{{ t('orgs.permWrite') }}</option>
                    <option value="manage">{{ t('orgs.permManage') }}</option>
                    <option value="none">{{ t('orgs.permNone') }}</option>
                  </select>
                  <button class="btn-icon-danger sm" @click="removeUserPerm(up)">
                    <svg viewBox="0 0 24 24" width="13" height="13" fill="currentColor"><path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"/></svg>
                  </button>
                </div>
              </div>
            </div>

            <!-- ── Groups ─────────────────────────────────────────────────── -->
            <div class="access-section">
              <div class="access-section-header">
                <span class="access-section-title">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z"/></svg>
                  {{ t('orgs.groups') }}
                </span>
                <button v-if="canManage" class="btn-add-access" @click="showGroupPicker = !showGroupPicker">
                  <svg viewBox="0 0 24 24" width="13" height="13" fill="currentColor"><path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z"/></svg>
                  {{ t('orgs.addGroup') }}
                </button>
              </div>

              <!-- Group picker -->
              <div v-if="showGroupPicker" class="add-access-form">
                <select v-model="newGroupID" class="access-select">
                  <option value="">— {{ t('orgs.groups') }} —</option>
                  <option v-for="g in availableGroups" :key="g.id" :value="g.id">{{ g.name }}</option>
                </select>
                <select v-model="newGroupLevel" class="access-select level-select">
                  <option value="read">{{ t('orgs.permRead') }}</option>
                  <option value="write">{{ t('orgs.permWrite') }}</option>
                  <option value="manage">{{ t('orgs.permManage') }}</option>
                  <option value="none">{{ t('orgs.permNone') }}</option>
                </select>
                <button class="btn-confirm" @click="addGroupPerm" :disabled="!newGroupID || savingGroup">
                  <span v-if="savingGroup" class="spinner-sm"></span>
                  <svg v-else viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41L9 16.17z"/></svg>
                </button>
                <button class="btn-cancel-add" @click="showGroupPicker = false; newGroupID = ''">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
                </button>
              </div>

              <!-- Group rows -->
              <div v-if="groupPerms.length === 0" class="access-empty">{{ t('orgs.noGroupAccessOnFolder') }}</div>
              <div v-else class="access-rows">
                <div v-for="gp in groupPerms" :key="gp.group.id" class="access-row">
                  <div class="access-avatar group-avatar">{{ gp.group.name.charAt(0).toUpperCase() }}</div>
                  <span class="access-name">{{ gp.group.name }}</span>
                  <select
                    class="access-select level-select"
                    :value="gp.permission.level"
                    @change="updateGroupPerm(gp, $event.target.value)"
                  >
                    <option value="read">{{ t('orgs.permRead') }}</option>
                    <option value="write">{{ t('orgs.permWrite') }}</option>
                    <option value="manage">{{ t('orgs.permManage') }}</option>
                    <option value="none">{{ t('orgs.permNone') }}</option>
                  </select>
                  <button class="btn-icon-danger sm" @click="removeGroupPerm(gp)">
                    <svg viewBox="0 0 24 24" width="13" height="13" fill="currentColor"><path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"/></svg>
                  </button>
                </div>
              </div>
            </div>

          </template>
        </div>

      </div>
    </div>
  </Transition>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useOrgStore } from '../../stores/organizations'

const props = defineProps({
  modelValue: { type: Boolean, required: true },
  orgID:      { type: Number,  required: true },
  folder:     { type: Object,  default: null }, // { path, name }
  canManage:  { type: Boolean, default: false }, // org admin or owner
})
const emit = defineEmits(['update:modelValue'])

const { t } = useI18n()
const orgStore = useOrgStore()

const loading    = ref(false)
const userPerms  = ref([])
const groupPerms = ref([])

const showUserPicker  = ref(false)
const newUserID       = ref('')
const newUserLevel    = ref('read')
const savingUser      = ref(false)

const showGroupPicker = ref(false)
const newGroupID      = ref('')
const newGroupLevel   = ref('read')
const savingGroup     = ref(false)

// ── helpers ───────────────────────────────────────────────────────────────────

const folderDisplayName = computed(() => {
  if (!props.folder) return ''
  return props.folder.name || props.folder.path
})

function memberName(userID) {
  const m = orgStore.members.find(m => m.user_id === userID)
  return m?.name || m?.email || userID
}
function memberInitial(userID) {
  return memberName(userID).charAt(0).toUpperCase()
}

const availableMembers = computed(() => {
  const alreadySet = new Set(userPerms.value.map(p => p.user_id))
  return orgStore.members.filter(m => !alreadySet.has(m.user_id))
})

const availableGroups = computed(() => {
  const alreadySet = new Set(groupPerms.value.map(gp => gp.group.id))
  return orgStore.groups.filter(g => !alreadySet.has(g.id))
})

// ── load / close ──────────────────────────────────────────────────────────────

async function load() {
  if (!props.folder) return
  loading.value = true
  showUserPicker.value  = false
  showGroupPicker.value = false
  newUserID.value       = ''
  newGroupID.value      = ''
  try {
    // Ensure members and groups lists are available for pickers
    const fetches = []
    if (!orgStore.members.length) fetches.push(orgStore.fetchMembers(props.orgID))
    if (!orgStore.groups.length)  fetches.push(orgStore.fetchGroups(props.orgID))
    if (fetches.length) await Promise.all(fetches)

    const data = await orgStore.fetchFolderAccess(props.orgID, props.folder.path)
    userPerms.value  = data.users  || []
    groupPerms.value = data.groups || []
  } finally {
    loading.value = false
  }
}

function close() { emit('update:modelValue', false) }

watch(() => props.modelValue, (open) => { if (open) load() })

// ── user perm actions ─────────────────────────────────────────────────────────

async function addUserPerm() {
  if (!newUserID.value) return
  savingUser.value = true
  try {
    const perm = await orgStore.setPermission(props.orgID, {
      user_id:     newUserID.value,
      folder_path: props.folder.path,
      level:       newUserLevel.value,
    })
    userPerms.value.push(perm)
    showUserPicker.value = false
    newUserID.value      = ''
    newUserLevel.value   = 'read'
  } finally {
    savingUser.value = false
  }
}

async function updateUserPerm(up, level) {
  await orgStore.setPermission(props.orgID, {
    user_id:     up.user_id,
    folder_path: props.folder.path,
    level,
  })
  const idx = userPerms.value.findIndex(p => p.user_id === up.user_id)
  if (idx !== -1) userPerms.value[idx] = { ...userPerms.value[idx], level }
}

async function removeUserPerm(up) {
  await orgStore.deletePermission(props.orgID, up.user_id, props.folder.path)
  userPerms.value = userPerms.value.filter(p => p.user_id !== up.user_id)
}

// ── group perm actions ────────────────────────────────────────────────────────

async function addGroupPerm() {
  if (!newGroupID.value) return
  savingGroup.value = true
  try {
    const perm = await orgStore.setGroupPermission(props.orgID, newGroupID.value, {
      folder_path: props.folder.path,
      level:       newGroupLevel.value,
    })
    const group = orgStore.groups.find(g => g.id === Number(newGroupID.value))
    if (group) groupPerms.value.push({ group, permission: perm })
    showGroupPicker.value = false
    newGroupID.value      = ''
    newGroupLevel.value   = 'read'
  } finally {
    savingGroup.value = false
  }
}

async function updateGroupPerm(gp, level) {
  await orgStore.setGroupPermission(props.orgID, gp.group.id, {
    folder_path: props.folder.path,
    level,
  })
  const idx = groupPerms.value.findIndex(g => g.group.id === gp.group.id)
  if (idx !== -1) groupPerms.value[idx] = { ...groupPerms.value[idx], permission: { ...groupPerms.value[idx].permission, level } }
}

async function removeGroupPerm(gp) {
  await orgStore.deleteGroupPermission(props.orgID, gp.group.id, props.folder.path)
  groupPerms.value = groupPerms.value.filter(g => g.group.id !== gp.group.id)
}
</script>

<style scoped>
/* ── Dialog shell ─────────────────────────────────────────────────────────── */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.access-dialog {
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 16px;
  width: 500px;
  max-width: calc(100vw - 32px);
  max-height: 80vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.25);
}

/* ── Header ───────────────────────────────────────────────────────────────── */
.dialog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 18px 20px 14px;
  border-bottom: 1px solid var(--border-color);
  flex-shrink: 0;
}

.dialog-title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.dialog-icon {
  color: var(--primary-color);
  flex-shrink: 0;
}

.dialog-heading {
  font-size: 1rem;
  font-weight: 700;
  color: var(--main-text-color);
  margin: 0;
}

.dialog-subtitle {
  font-size: 0.78rem;
  color: var(--secondary-text-color);
  margin: 2px 0 0;
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
}
.btn-close:hover { background: var(--hover-background-color); color: var(--main-text-color); }

/* ── Body ─────────────────────────────────────────────────────────────────── */
.dialog-body {
  flex: 1;
  overflow-y: auto;
  padding: 0;
}

.dialog-loading {
  display: flex;
  justify-content: center;
  padding: 48px 0;
}

/* ── Sections ─────────────────────────────────────────────────────────────── */
.access-section {
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
}
.access-section:last-child { border-bottom: none; }

.access-section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.access-section-title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.8rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--secondary-text-color);
}

.btn-add-access {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 0.78rem;
  font-weight: 600;
  color: var(--primary-color);
  background: none;
  border: 1px solid var(--primary-color);
  border-radius: 6px;
  padding: 3px 10px;
  cursor: pointer;
  transition: background 0.15s;
}
.btn-add-access:hover { background: color-mix(in srgb, var(--primary-color) 10%, transparent); }

/* ── Add form ─────────────────────────────────────────────────────────────── */
.add-access-form {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 10px;
  padding: 10px 12px;
  background: var(--hover-background-color);
  border-radius: 8px;
}

.access-select {
  flex: 1;
  background: var(--background-color);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  padding: 5px 8px;
  font-size: 0.82rem;
  color: var(--main-text-color);
  cursor: pointer;
}

.level-select {
  flex: 0 0 auto;
  min-width: 90px;
}

.btn-confirm {
  background: var(--primary-color);
  border: none;
  border-radius: 6px;
  color: white;
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  flex-shrink: 0;
}
.btn-confirm:disabled { opacity: 0.6; cursor: not-allowed; }

.btn-cancel-add {
  background: none;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  color: var(--secondary-text-color);
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  flex-shrink: 0;
}
.btn-cancel-add:hover { background: var(--hover-background-color); }

/* ── Access rows ──────────────────────────────────────────────────────────── */
.access-rows { display: flex; flex-direction: column; gap: 6px; }

.access-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 8px;
  border-radius: 8px;
  transition: background 0.1s;
}
.access-row:hover { background: var(--hover-background-color); }

.access-avatar {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  background: var(--primary-color);
  color: white;
  font-size: 0.78rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.access-avatar.group-avatar { border-radius: 7px; }

.access-name {
  flex: 1;
  font-size: 0.87rem;
  color: var(--main-text-color);
  min-width: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.access-empty {
  font-size: 0.82rem;
  color: var(--secondary-text-color);
  text-align: center;
  padding: 8px 0;
}

/* ── Danger button ────────────────────────────────────────────────────────── */
.btn-icon-danger {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--secondary-text-color);
  padding: 4px;
  border-radius: 5px;
  display: flex;
  align-items: center;
  flex-shrink: 0;
  transition: color 0.15s, background 0.15s;
}
.btn-icon-danger:hover { color: #ef4444; background: rgba(239, 68, 68, 0.08); }

/* ── Spinner ──────────────────────────────────────────────────────────────── */
.spinner {
  width: 28px;
  height: 28px;
  border: 3px solid var(--border-color);
  border-top-color: var(--primary-color);
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}
.spinner-sm {
  width: 12px;
  height: 12px;
  border: 2px solid rgba(255,255,255,0.4);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

/* ── Transitions ──────────────────────────────────────────────────────────── */
.modal-enter-active, .modal-leave-active { transition: opacity 0.2s; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
.modal-enter-active .access-dialog, .modal-leave-active .access-dialog { transition: transform 0.2s; }
.modal-enter-from .access-dialog, .modal-leave-to .access-dialog { transform: scale(0.96) translateY(8px); }
</style>
