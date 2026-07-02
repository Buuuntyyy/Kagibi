<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="groups-panel">
    <!-- Header -->
    <div class="panel-header">
      <h3 class="panel-title">{{ t('orgs.groups') }}</h3>
      <button v-if="isOrgAdmin" class="btn-primary" @click="openCreateModal">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z"/></svg>
        {{ t('orgs.createGroup') }}
      </button>
    </div>

    <!-- Empty state -->
    <div v-if="orgStore.groups.length === 0 && !loading" class="empty-state">
      <svg viewBox="0 0 24 24" width="48" height="48" fill="currentColor" class="empty-icon">
        <path d="M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z"/>
      </svg>
      <p>{{ t('orgs.noGroups') }}</p>
    </div>

    <div v-else-if="loading" class="loading-state">
      <div class="spinner"></div>
    </div>

    <!-- Group list -->
    <div v-else class="groups-list">
      <div
        v-for="group in orgStore.groups"
        :key="group.id"
        class="group-card"
        :class="{ active: selectedGroup?.id === group.id }"
        @click="selectGroup(group)"
      >
        <div class="group-card-left">
          <div class="group-avatar">{{ group.name.charAt(0).toUpperCase() }}</div>
          <div class="group-info">
            <span class="group-name">{{ group.name }}</span>
            <span v-if="group.description" class="group-desc">{{ group.description }}</span>
          </div>
        </div>
        <div class="group-card-right">
          <span v-if="group.source === 'ldap'" class="badge-ldap">LDAP</span>
          <button v-if="isOrgAdmin" class="btn-icon-danger" @click.stop="confirmDelete(group)" :title="t('orgs.deleteGroup')">
            <svg viewBox="0 0 24 24" width="15" height="15" fill="currentColor"><path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"/></svg>
          </button>
        </div>
      </div>
    </div>

    <!-- Group detail panel -->
    <Transition name="slide">
      <div v-if="selectedGroup" class="group-detail">
        <div class="detail-header">
          <div class="detail-title-row">
            <div class="group-avatar lg">{{ selectedGroup.name.charAt(0).toUpperCase() }}</div>
            <div>
              <h4 class="detail-name">{{ selectedGroup.name }}</h4>
              <p v-if="selectedGroup.description" class="detail-desc">{{ selectedGroup.description }}</p>
            </div>
          </div>
          <button class="btn-close" @click="selectedGroup = null">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
          </button>
        </div>

        <!-- Members sub-tab -->
        <div class="detail-section">
          <div class="section-header">
            <span class="section-title">{{ t('orgs.members') }}</span>
            <button v-if="canManageSelected" class="btn-sm" @click="openAddMemberModal">
              <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z"/></svg>
              {{ t('orgs.addMember') }}
            </button>
          </div>

          <div v-if="groupMembers.length === 0" class="section-empty">{{ t('orgs.noGroupMembers') }}</div>
          <div v-else class="member-rows">
            <div v-for="gm in groupMembers" :key="gm.id" class="member-row">
              <div class="member-avatar-sm">{{ memberName(gm.user_id).charAt(0).toUpperCase() }}</div>
              <span class="member-name">{{ memberName(gm.user_id) }}</span>
              <select
                v-if="canManageSelected"
                class="role-select-sm"
                :value="gm.role"
                @change="handleRoleChange(gm, $event.target.value)"
              >
                <option value="member">{{ t('orgs.groupMember') }}</option>
                <option value="admin">{{ t('orgs.groupAdmin') }}</option>
              </select>
              <span v-else class="role-badge-sm">{{ t(`orgs.group${capitalize(gm.role)}`) }}</span>
              <button v-if="canManageSelected" class="btn-icon-danger sm" @click="handleRemoveMember(gm)" :title="t('orgs.removeMember')">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
              </button>
            </div>
          </div>
        </div>

        <!-- Permissions sub-tab -->
        <div class="detail-section">
          <div class="section-header">
            <span class="section-title">{{ t('orgs.permissions') }}</span>
            <button v-if="isOrgAdmin" class="btn-sm" @click="openPermModal">
              <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z"/></svg>
              {{ t('orgs.addPermission') }}
            </button>
          </div>

          <div v-if="groupPermissions.length === 0" class="section-empty">{{ t('orgs.noGroupPermissions') }}</div>
          <div v-else class="perm-rows">
            <div v-for="perm in groupPermissions" :key="perm.id" class="perm-row">
              <code class="perm-path">{{ displayPath(perm.folder_path) }}</code>
              <span class="perm-level" :class="perm.level">{{ perm.level }}</span>
              <button v-if="canManageSelected" class="btn-icon-danger sm" @click="handleDeletePerm(perm)" :title="t('orgs.removePermission')">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
              </button>
            </div>
          </div>
        </div>

        <!-- Encryption section (admin only) -->
        <div v-if="isOrgAdmin" class="detail-section">
          <div class="section-header">
            <span class="section-title">Chiffrement de groupe</span>
            <div class="enc-actions">
              <button
                v-if="!groupEncInitialized"
                class="btn-sm btn-enc-init"
                :disabled="encLoading"
                @click="handleInitGroupKey"
              >
                <span v-if="encLoading" class="spinner-sm dark"></span>
                <svg v-else viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zm-6 9c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2zm3.1-9H8.9V6c0-1.71 1.39-3.1 3.1-3.1 1.71 0 3.1 1.39 3.1 3.1v2z"/></svg>
                Initialiser le chiffrement
              </button>
              <button
                v-else
                class="btn-sm"
                :disabled="encLoading"
                @click="handleRotateGroupKey"
              >
                <span v-if="encLoading" class="spinner-sm dark"></span>
                <svg v-else viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M12 6v3l4-4-4-4v3c-4.42 0-8 3.58-8 8 0 1.57.46 3.03 1.24 4.26L6.7 14.8c-.45-.83-.7-1.79-.7-2.8 0-3.31 2.69-6 6-6zm6.76 1.74L17.3 9.2c.44.84.7 1.79.7 2.8 0 3.31-2.69 6-6 6v-3l-4 4 4 4v-3c4.42 0 8-3.58 8-8 0-1.57-.46-3.03-1.24-4.26z"/></svg>
                Faire pivoter la clé
              </button>
            </div>
          </div>

          <!-- Key status -->
          <div class="enc-status-row">
            <span v-if="groupEncInitialized" class="enc-status-badge initialized">
              <svg viewBox="0 0 24 24" width="12" height="12" fill="currentColor"><path d="M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zm-6 9c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2zm3.1-9H8.9V6c0-1.71 1.39-3.1 3.1-3.1 1.71 0 3.1 1.39 3.1 3.1v2z"/></svg>
              Chiffrement actif
            </span>
            <span v-else class="enc-status-badge not-initialized">
              <svg viewBox="0 0 24 24" width="12" height="12" fill="currentColor"><path d="M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zm-6 9c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2zm3.1-9H8.9V6c0-1.71 1.39-3.1 3.1-3.1 1.71 0 3.1 1.39 3.1 3.1v2z"/></svg>
              Non initialisé
            </span>
          </div>

          <p v-if="encError" class="form-error" style="margin-top:6px;">{{ encError }}</p>

          <!-- Per-member provisioning status (only visible when key is initialized) -->
          <div v-if="groupEncInitialized && groupMembers.length" class="member-key-rows">
            <div v-for="gm in groupMembers" :key="gm.id" class="member-key-row">
              <div class="member-avatar-sm">{{ memberName(gm.user_id).charAt(0).toUpperCase() }}</div>
              <span class="member-name">{{ memberName(gm.user_id) }}</span>
              <span v-if="provisionedUserIDs.has(gm.user_id)" class="key-badge provisioned">Clé provisionnée</span>
              <span v-else class="key-badge not-provisioned">Non provisionnée</span>
              <button
                v-if="!provisionedUserIDs.has(gm.user_id)"
                class="btn-sm btn-provision"
                :disabled="encLoading"
                @click="handleProvisionMember(gm)"
              >Provisionner</button>
            </div>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Create group modal -->
    <Transition name="modal">
      <div v-if="showCreateModal" class="modal-overlay" @click.self="showCreateModal = false">
        <div class="modal">
          <div class="modal-header">
            <h3>{{ t('orgs.createGroup') }}</h3>
            <button class="btn-close" @click="showCreateModal = false">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
            </button>
          </div>
          <div class="modal-body">
            <div class="form-group">
              <label>{{ t('orgs.groupName') }} *</label>
              <input v-model="createForm.name" type="text" class="input-field" :placeholder="t('orgs.groupName')" />
            </div>
            <div class="form-group">
              <label>{{ t('orgs.orgDesc') }}</label>
              <input v-model="createForm.description" type="text" class="input-field" :placeholder="t('orgs.orgDesc')" />
            </div>
            <p v-if="createError" class="form-error">{{ createError }}</p>
          </div>
          <div class="modal-footer">
            <button class="btn-secondary" @click="showCreateModal = false">{{ t('orgs.cancel') }}</button>
            <button class="btn-primary" @click="handleCreate" :disabled="creating || !createForm.name">
              <span v-if="creating" class="spinner-sm"></span>
              {{ creating ? t('common.loading') : t('orgs.create') }}
            </button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Add member modal -->
    <Transition name="modal">
      <div v-if="showAddMemberModal" class="modal-overlay" @click.self="showAddMemberModal = false">
        <div class="modal">
          <div class="modal-header">
            <h3>{{ t('orgs.addMember') }}</h3>
            <button class="btn-close" @click="showAddMemberModal = false">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
            </button>
          </div>
          <div class="modal-body">
            <p class="modal-hint">{{ t('orgs.addMemberHint') }}</p>
            <div class="member-picker">
              <div
                v-for="m in availableMembers"
                :key="m.user_id"
                class="pick-row"
                :class="{ selected: selectedUserID === m.user_id }"
                @click="selectedUserID = m.user_id"
              >
                <div class="member-avatar-sm">{{ (m.name || m.email || '?').charAt(0).toUpperCase() }}</div>
                <span>{{ m.name || m.email }}</span>
                <svg v-if="selectedUserID === m.user_id" viewBox="0 0 24 24" width="16" height="16" fill="currentColor" class="check-icon"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41L9 16.17z"/></svg>
              </div>
              <p v-if="availableMembers.length === 0" class="section-empty">{{ t('orgs.allMembersInGroup') }}</p>
            </div>
            <div class="form-group" style="margin-top: 12px;">
              <label>{{ t('orgs.groupRole') }}</label>
              <select v-model="newMemberRole" class="input-field">
                <option value="member">{{ t('orgs.groupMember') }}</option>
                <option value="admin">{{ t('orgs.groupAdmin') }}</option>
              </select>
            </div>
            <p v-if="addMemberError" class="form-error">{{ addMemberError }}</p>
          </div>
          <div class="modal-footer">
            <button class="btn-secondary" @click="showAddMemberModal = false">{{ t('orgs.cancel') }}</button>
            <button class="btn-primary" @click="handleAddMember" :disabled="addingMember || !selectedUserID">
              <span v-if="addingMember" class="spinner-sm"></span>
              {{ addingMember ? t('common.loading') : t('orgs.addMember') }}
            </button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Set permission modal -->
    <Transition name="modal">
      <div v-if="showPermModal" class="modal-overlay" @click.self="showPermModal = false">
        <div class="modal">
          <div class="modal-header">
            <h3>{{ t('orgs.addPermission') }}</h3>
            <button class="btn-close" @click="showPermModal = false">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
            </button>
          </div>
          <div class="modal-body">
            <div class="form-group">
              <label>{{ t('orgs.folderPath') }}</label>
              <select v-if="folderOptions.length > 0" v-model="permForm.folder_path" class="input-field">
                <option value="/">/</option>
                <option v-for="f in folderOptions" :key="f.path" :value="f.path">{{ f.name }}</option>
              </select>
              <input v-else v-model="permForm.folder_path" type="text" class="input-field" placeholder="/" />
            </div>
            <div class="form-group">
              <label>{{ t('orgs.level') }}</label>
              <select v-model="permForm.level" class="input-field">
                <option value="read">read</option>
                <option value="write">write</option>
                <option value="manage">manage</option>
                <option value="none">none</option>
              </select>
            </div>
            <p v-if="permError" class="form-error">{{ permError }}</p>
          </div>
          <div class="modal-footer">
            <button class="btn-secondary" @click="showPermModal = false">{{ t('orgs.cancel') }}</button>
            <button class="btn-primary" @click="handleSetPerm" :disabled="settingPerm">
              <span v-if="settingPerm" class="spinner-sm"></span>
              {{ settingPerm ? t('common.loading') : t('orgs.save') }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useOrgStore } from '../../stores/organizations'
import { useUIStore } from '../../stores/ui'

const props = defineProps({
  orgID: { type: Number, required: true },
})

const { t } = useI18n()
const orgStore = useOrgStore()
const uiStore = useUIStore()
const loading = ref(false)
const selectedGroup = ref(null)
const groupMembers = ref([])
const groupPermissions = ref([])

// ── Group encryption ──────────────────────────────────────────────────────────
const encLoading = ref(false)
const encError = ref('')
const provisionedUserIDs = ref(new Set())

const groupEncInitialized = computed(() => !!selectedGroup.value?.encrypted_group_key)

async function loadEncryptionState() {
  if (!selectedGroup.value || !isOrgAdmin.value) return
  encError.value = ''
  try {
    const ids = await orgStore.fetchGroupKeyProvisionedMembers(props.orgID, selectedGroup.value.id)
    provisionedUserIDs.value = new Set(ids)
  } catch (_) {
    provisionedUserIDs.value = new Set()
  }
}

async function handleInitGroupKey() {
  encLoading.value = true
  encError.value = ''
  try {
    await orgStore.initializeGroupKey(props.orgID, selectedGroup.value.id)
    // Reflect the new state in the cached group object
    const g = orgStore.groups.find(g => g.id === selectedGroup.value.id)
    if (g) g.encrypted_group_key = '__initialized__'
    selectedGroup.value = { ...selectedGroup.value, encrypted_group_key: '__initialized__' }
    await loadEncryptionState()
  } catch (e) {
    encError.value = e.response?.data?.error || e.message
  } finally {
    encLoading.value = false
  }
}

async function handleRotateGroupKey() {
  if (!await uiStore.showConfirm({
    title: 'Faire pivoter la clé de groupe',
    message: 'Cette opération re-chiffre tous les fichiers du groupe avec une nouvelle clé. Continuer ?',
    confirmLabel: 'Faire pivoter',
  })) return
  encLoading.value = true
  encError.value = ''
  try {
    await orgStore.rotateGroupKey(props.orgID, selectedGroup.value.id)
    await loadEncryptionState()
  } catch (e) {
    encError.value = e.response?.data?.error || e.message
  } finally {
    encLoading.value = false
  }
}

async function handleProvisionMember(gm) {
  const member = orgStore.members.find(m => m.user_id === gm.user_id)
  if (!member?.public_key) {
    encError.value = `Clé publique introuvable pour ${memberName(gm.user_id)}.`
    return
  }
  encLoading.value = true
  encError.value = ''
  try {
    await orgStore.provisionGroupKeyForMember(props.orgID, selectedGroup.value.id, member)
    provisionedUserIDs.value = new Set([...provisionedUserIDs.value, gm.user_id])
  } catch (e) {
    encError.value = e.response?.data?.error || e.message
  } finally {
    encLoading.value = false
  }
}

// ── Authorization ─────────────────────────────────────────────────────────────

const isOrgAdmin = computed(() =>
  ['owner', 'admin'].includes(orgStore.currentOrg?.my_role),
)

// IDs des groupes où l'utilisateur est admin de groupe
const myAdminGroupIDs = computed(() => {
  const set = new Set()
  for (const g of orgStore.myGroups) {
    if (g.my_role === 'admin') set.add(g.id)
  }
  return set
})

// L'utilisateur peut gérer le groupe sélectionné (org admin OU admin du groupe)
const canManageSelected = computed(() =>
  isOrgAdmin.value || myAdminGroupIDs.value.has(selectedGroup.value?.id),
)

// Create group modal
const showCreateModal = ref(false)
const creating = ref(false)
const createError = ref('')
const createForm = ref({ name: '', description: '' })

// Add member modal
const showAddMemberModal = ref(false)
const addingMember = ref(false)
const addMemberError = ref('')
const selectedUserID = ref('')
const newMemberRole = ref('member')

// Permission modal
const showPermModal = ref(false)
const settingPerm = ref(false)
const permError = ref('')
const permForm = ref({ folder_path: '/', level: 'read', restrict_to_groups: false })
const folderOptions = ref([])

onMounted(async () => {
  loading.value = true
  try {
    const fetches = [orgStore.fetchGroups(props.orgID)]
    if (!orgStore.members.length) fetches.push(orgStore.fetchMembers(props.orgID))
    if (!orgStore.myGroups.length) fetches.push(orgStore.fetchMyGroups(props.orgID))
    await Promise.all(fetches)
  } finally {
    loading.value = false
  }
})

watch(selectedGroup, async (group) => {
  if (!group) {
    provisionedUserIDs.value = new Set()
    encError.value = ''
    return
  }
  groupMembers.value = await orgStore.fetchGroupMembers(props.orgID, group.id)
  groupPermissions.value = await orgStore.fetchGroupPermissions(props.orgID, group.id)
  await loadEncryptionState()
})

// Members not yet in the selected group
const availableMembers = computed(() => {
  const inGroup = new Set(groupMembers.value.map(gm => gm.user_id))
  return orgStore.members.filter(m => !inGroup.has(m.user_id))
})

function memberName(userID) {
  const m = orgStore.members.find(m => m.user_id === userID)
  return m?.name || m?.email || userID
}

function displayPath(encryptedPath) {
  if (!encryptedPath || encryptedPath === '/') return '/'
  return encryptedPath.split('/').map(seg => orgStore.folderNameCache[seg] || seg).join('/')
}

function openCreateModal() {
  createForm.value = { name: '', description: '' }
  createError.value = ''
  showCreateModal.value = true
}

async function handleCreate() {
  creating.value = true
  createError.value = ''
  try {
    const group = await orgStore.createGroup(props.orgID, createForm.value.name, createForm.value.description)
    try {
      const folder = await orgStore.createFolder(props.orgID, createForm.value.name, '/', '')
      await orgStore.setGroupPermission(props.orgID, group.id, {
        folder_path: folder.path,
        level: 'manage',
      })
    } catch (_) { /* non-fatal: group created, base folder setup failed */ }
    showCreateModal.value = false
    await selectGroup(group)
  } catch (e) {
    createError.value = e.response?.data?.error || e.message
  } finally {
    creating.value = false
  }
}

function openAddMemberModal() {
  selectedUserID.value = ''
  newMemberRole.value = 'member'
  addMemberError.value = ''
  showAddMemberModal.value = true
}

async function handleAddMember() {
  if (!selectedUserID.value) return
  addingMember.value = true
  addMemberError.value = ''
  try {
    const gm = await orgStore.addGroupMember(props.orgID, selectedGroup.value.id, selectedUserID.value, newMemberRole.value)
    groupMembers.value.push(gm)
    showAddMemberModal.value = false
  } catch (e) {
    addMemberError.value = e.response?.data?.error || e.message
  } finally {
    addingMember.value = false
  }
}

async function handleRemoveMember(gm) {
  await orgStore.removeGroupMember(props.orgID, selectedGroup.value.id, gm.id)
  groupMembers.value = groupMembers.value.filter(m => m.id !== gm.id)
}

async function handleRoleChange(gm, newRole) {
  const updated = await orgStore.updateGroupMemberRole(props.orgID, selectedGroup.value.id, gm.id, newRole)
  const idx = groupMembers.value.findIndex(m => m.id === gm.id)
  if (idx !== -1) groupMembers.value[idx] = { ...groupMembers.value[idx], role: updated.role }
}

async function openPermModal() {
  permForm.value = { folder_path: '/', level: 'read' }
  permError.value = ''
  folderOptions.value = []
  showPermModal.value = true
  try {
    const items = await orgStore.fetchItems(props.orgID, '/')
    folderOptions.value = (items.folders || []).map(f => ({ name: f.name, path: f.path }))
  } catch (_) { /* folder list unavailable — fall back to manual input */ }
}

async function handleSetPerm() {
  settingPerm.value = true
  permError.value = ''
  try {
    const perm = await orgStore.setGroupPermission(props.orgID, selectedGroup.value.id, permForm.value)
    const idx = groupPermissions.value.findIndex(p => p.folder_path === perm.folder_path)
    if (idx !== -1) groupPermissions.value[idx] = perm
    else groupPermissions.value.push(perm)
    showPermModal.value = false
  } catch (e) {
    permError.value = e.response?.data?.error || e.message
  } finally {
    settingPerm.value = false
  }
}

async function handleDeletePerm(perm) {
  await orgStore.deleteGroupPermission(props.orgID, selectedGroup.value.id, perm.folder_path)
  groupPermissions.value = groupPermissions.value.filter(p => p.id !== perm.id)
}

async function selectGroup(group) {
  selectedGroup.value = group
}

function capitalize(s) {
  return s ? s.charAt(0).toUpperCase() + s.slice(1) : ''
}

async function confirmDelete(group) {
  if (!await uiStore.showConfirm({ title: t('orgs.confirmDeleteGroup') || 'Supprimer le groupe', message: `${t('orgs.confirmDeleteGroup')} "${group.name}" ?`, confirmLabel: 'Supprimer' })) return
  if (selectedGroup.value?.id === group.id) selectedGroup.value = null
  await orgStore.deleteGroup(props.orgID, group.id)
}
</script>

<style scoped>
.groups-panel {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 0 16px 0;
  border-bottom: 1px solid var(--border-color);
  margin-bottom: 16px;
}

.panel-title {
  font-size: 1rem;
  font-weight: 700;
  color: var(--main-text-color);
  margin: 0;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 48px 0;
  color: var(--secondary-text-color);
  text-align: center;
}

.empty-icon { opacity: 0.3; color: var(--secondary-text-color); }

.loading-state {
  display: flex;
  justify-content: center;
  padding: 40px 0;
}

.groups-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  flex: 1;
}

.group-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 14px;
  border: 1px solid var(--border-color);
  border-radius: 10px;
  cursor: pointer;
  transition: background 0.15s, border-color 0.15s;
  background: var(--card-color);
}

.group-card:hover { background: var(--hover-background-color); }
.group-card.active { border-color: var(--primary-color); background: color-mix(in srgb, var(--primary-color) 6%, var(--card-color)); }

.group-card-left { display: flex; align-items: center; gap: 10px; }
.group-card-right { display: flex; align-items: center; gap: 8px; }

.group-avatar {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background: linear-gradient(135deg, var(--primary-color), var(--secondary-color));
  color: white;
  font-weight: 700;
  font-size: 1rem;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.group-avatar.lg { width: 44px; height: 44px; border-radius: 10px; font-size: 1.2rem; }

.group-info { display: flex; flex-direction: column; gap: 2px; }
.group-name { font-size: 0.9rem; font-weight: 600; color: var(--main-text-color); }
.group-desc { font-size: 0.78rem; color: var(--secondary-text-color); }

.badge-ldap {
  font-size: 0.65rem;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 4px;
  background: color-mix(in srgb, var(--secondary-color) 12%, transparent);
  color: var(--secondary-color);
  letter-spacing: 0.04em;
}

/* Group detail panel */
.group-detail {
  margin-top: 20px;
  border: 1px solid var(--border-color);
  border-radius: 12px;
  overflow: hidden;
  background: var(--card-color);
}

.detail-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  padding: 16px 18px;
  border-bottom: 1px solid var(--border-color);
}

.detail-title-row { display: flex; align-items: center; gap: 12px; }
.detail-name { font-size: 1rem; font-weight: 700; color: var(--main-text-color); margin: 0; }
.detail-desc { font-size: 0.82rem; color: var(--secondary-text-color); margin: 4px 0 0 0; }

.detail-section {
  padding: 14px 18px;
  border-bottom: 1px solid var(--border-color);
}
.detail-section:last-child { border-bottom: none; }

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}

.section-title { font-size: 0.82rem; font-weight: 700; color: var(--secondary-text-color); text-transform: uppercase; letter-spacing: 0.05em; }
.section-empty { font-size: 0.82rem; color: var(--secondary-text-color); text-align: center; padding: 10px 0; }

.member-rows, .perm-rows { display: flex; flex-direction: column; gap: 6px; }

.member-row, .perm-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  border-radius: 6px;
  background: var(--hover-background-color);
}

.member-avatar-sm {
  width: 26px;
  height: 26px;
  border-radius: 6px;
  background: var(--primary-color);
  color: white;
  font-size: 0.75rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.member-name { flex: 1; font-size: 0.85rem; color: var(--main-text-color); }

.role-badge-sm {
  font-size: 0.7rem;
  font-weight: 600;
  padding: 2px 7px;
  border-radius: 8px;
  background: color-mix(in srgb, var(--secondary-color) 12%, transparent);
  color: var(--secondary-color);
}

.role-select-sm {
  font-size: 0.75rem;
  padding: 2px 6px;
  border-radius: 6px;
  border: 1px solid var(--border-color);
  background: var(--background-color);
  color: var(--main-text-color);
  cursor: pointer;
}

.perm-path { flex: 1; font-size: 0.8rem; color: var(--secondary-text-color); background: none; }
.perm-level {
  font-size: 0.72rem;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: 10px;
}
.perm-level.read   { background: color-mix(in srgb, var(--success-color) 12%, transparent);    color: var(--success-color); }
.perm-level.write  { background: color-mix(in srgb, var(--primary-color) 12%, transparent);   color: var(--primary-color); }
.perm-level.manage { background: color-mix(in srgb, var(--secondary-color) 12%, transparent); color: var(--secondary-color); }
.perm-level.none   { background: color-mix(in srgb, var(--error-color) 12%, transparent);     color: var(--error-color); }

/* Member picker */
.modal-hint { font-size: 0.82rem; color: var(--secondary-text-color); margin: 0 0 10px 0; }
.member-picker { display: flex; flex-direction: column; gap: 4px; max-height: 240px; overflow-y: auto; }
.pick-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  border-radius: 8px;
  cursor: pointer;
  font-size: 0.875rem;
  color: var(--main-text-color);
  transition: background 0.1s;
}
.pick-row:hover { background: var(--hover-background-color); }
.pick-row.selected { background: color-mix(in srgb, var(--primary-color) 10%, transparent); }
.check-icon { margin-left: auto; color: var(--primary-color); }

/* Buttons */
.btn-primary {
  display: flex;
  align-items: center;
  gap: 6px;
  background: var(--primary-color);
  color: white;
  border: none;
  border-radius: 8px;
  padding: 8px 14px;
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.2s;
}
.btn-primary:hover { opacity: 0.9; }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }

.btn-secondary {
  background: var(--card-color);
  color: var(--main-text-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 8px 14px;
  font-size: 0.85rem;
  cursor: pointer;
  transition: background 0.15s;
}
.btn-secondary:hover { background: var(--hover-background-color); }

.btn-sm {
  display: flex;
  align-items: center;
  gap: 4px;
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  padding: 4px 10px;
  font-size: 0.78rem;
  color: var(--main-text-color);
  cursor: pointer;
  transition: background 0.15s;
}
.btn-sm:hover { background: var(--hover-background-color); }

.btn-icon-danger {
  background: none;
  border: none;
  cursor: pointer;
  color: #ef4444;
  padding: 4px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  opacity: 0.6;
  transition: opacity 0.15s;
}
.btn-icon-danger:hover { opacity: 1; }
.btn-icon-danger.sm { padding: 2px; margin-left: auto; }

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
  max-width: 440px;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 18px 22px 14px;
  border-bottom: 1px solid var(--border-color);
}

.modal-header h3 { margin: 0; font-size: 1rem; font-weight: 700; color: var(--main-text-color); }

.modal-body { padding: 18px 22px; }
.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 14px 22px 18px;
  border-top: 1px solid var(--border-color);
}

.form-group { display: flex; flex-direction: column; gap: 6px; margin-bottom: 14px; }
.form-group label { font-size: 0.82rem; font-weight: 500; color: var(--secondary-text-color); }
.form-group:last-child { margin-bottom: 0; }

.input-field {
  background: var(--background-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 9px 12px;
  font-size: 0.875rem;
  color: var(--main-text-color);
  transition: border-color 0.15s;
  width: 100%;
  box-sizing: border-box;
}
.input-field:focus { outline: none; border-color: var(--primary-color); }

.form-error { color: #ef4444; font-size: 0.82rem; margin: 0; }

.form-group-check { gap: 4px; }
.check-label { display: flex; align-items: center; gap: 8px; font-size: 0.875rem; color: var(--main-text-color); cursor: pointer; }
.check-input { width: 15px; height: 15px; accent-color: var(--primary-color); cursor: pointer; flex-shrink: 0; }
.check-hint { font-size: 0.75rem; color: var(--secondary-text-color); padding-left: 23px; }

/* Spinner */
.spinner {
  width: 28px;
  height: 28px;
  border: 3px solid var(--border-color);
  border-top-color: var(--primary-color);
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}
.spinner-sm {
  display: inline-block;
  width: 12px;
  height: 12px;
  border: 2px solid rgba(255,255,255,0.4);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

/* Encryption section */
.enc-actions { display: flex; gap: 8px; }

.btn-enc-init {
  color: var(--primary-color);
  border-color: var(--primary-color);
  font-weight: 600;
}

.enc-status-row { margin-bottom: 8px; }

.enc-status-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 0.75rem;
  font-weight: 600;
  padding: 3px 10px;
  border-radius: 10px;
}
.enc-status-badge.initialized {
  background: color-mix(in srgb, var(--success-color) 12%, transparent);
  color: var(--success-color);
}
.enc-status-badge.not-initialized {
  background: color-mix(in srgb, #f59e0b 12%, transparent);
  color: #f59e0b;
}

.member-key-rows { display: flex; flex-direction: column; gap: 5px; margin-top: 8px; }

.member-key-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 5px 8px;
  border-radius: 6px;
  background: var(--hover-background-color);
}

.key-badge {
  font-size: 0.7rem;
  font-weight: 600;
  padding: 2px 7px;
  border-radius: 8px;
}
.key-badge.provisioned {
  background: color-mix(in srgb, var(--success-color) 12%, transparent);
  color: var(--success-color);
}
.key-badge.not-provisioned {
  background: color-mix(in srgb, #f59e0b 12%, transparent);
  color: #f59e0b;
}

.btn-provision {
  margin-left: auto;
  color: var(--primary-color);
  border-color: color-mix(in srgb, var(--primary-color) 40%, transparent);
}

.spinner-sm.dark {
  border-color: color-mix(in srgb, var(--main-text-color) 20%, transparent);
  border-top-color: var(--main-text-color);
}

/* Transitions */
.slide-enter-active, .slide-leave-active { transition: opacity 0.2s, transform 0.2s; }
.slide-enter-from, .slide-leave-to { opacity: 0; transform: translateY(8px); }
.modal-enter-active, .modal-leave-active { transition: opacity 0.2s; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
</style>
