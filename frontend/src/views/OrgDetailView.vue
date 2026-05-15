<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="org-detail">
    <!-- Header -->
    <header class="org-header">
      <button class="btn-back" @click="router.push('/dashboard/organizations')">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20v-2z"/></svg>
        {{ t('orgs.back') }}
      </button>

      <div v-if="orgStore.currentOrg" class="org-identity">
        <!-- Logo / avatar — click to upload (admins only) -->
        <div class="org-avatar-wrap" :class="{ 'is-admin': canManage }" @click="canManage && logoInputRef?.click()">
          <img v-if="orgStore.currentOrg.logo_url" class="org-avatar org-avatar-img" :src="orgStore.currentOrg.logo_url" :alt="orgStore.currentOrg.name" />
          <div v-else class="org-avatar">{{ orgStore.currentOrg.name.charAt(0).toUpperCase() }}</div>
          <div v-if="canManage" class="org-avatar-overlay">
            <span v-if="uploadingLogo" class="spinner-sm"></span>
            <svg v-else viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M12 15.2A3.2 3.2 0 0 1 8.8 12 3.2 3.2 0 0 1 12 8.8 3.2 3.2 0 0 1 15.2 12 3.2 3.2 0 0 1 12 15.2M20 4h-3.17L15 2H9L7.17 4H4c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V6c0-1.1-.9-2-2-2z"/></svg>
          </div>
          <button v-if="canManage && orgStore.currentOrg.logo_url" class="logo-remove-btn" @click.stop="handleRemoveLogo" :title="t('orgs.removeLogo')">
            <svg viewBox="0 0 24 24" width="10" height="10" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
          </button>
          <input ref="logoInputRef" type="file" accept="image/jpeg,image/png,image/gif,image/webp,image/svg+xml" style="display:none" @change="handleLogoChange" />
        </div>
        <div class="org-identity-body">
          <div class="org-identity-top">
            <h1 class="org-name">{{ orgStore.currentOrg.name }}</h1>
            <span class="role-badge" :class="orgStore.currentOrg.my_role" v-if="orgStore.currentOrg.my_role">
              {{ t(`orgs.${orgStore.currentOrg.my_role}`) }}
            </span>
          </div>
          <p v-if="orgStore.currentOrg.description" class="org-desc">{{ orgStore.currentOrg.description }}</p>
          <div class="storage-indicator">
            <div class="storage-track">
              <div class="storage-fill" :class="storageClass" :style="{ width: storagePercent + '%' }"></div>
            </div>
            <span class="storage-text">{{ formatSize(orgStore.currentOrg.storage_used_bytes) }} / {{ formatSize(orgStore.currentOrg.storage_quota_mb * 1024 * 1024) }}</span>
          </div>
        </div>
      </div>
    </header>

    <div v-if="orgStore.loading && !orgStore.currentOrg" class="loading-center">
      <div class="spinner"></div>
    </div>

    <template v-else-if="orgStore.currentOrg">
      <!-- Tabs -->
      <div class="tabs-bar">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          class="tab-btn"
          :class="{ active: activeTab === tab.key }"
          @click="switchTab(tab.key)"
        >
          <component :is="tab.icon" class="tab-icon" />
          {{ tab.label }}
          <span v-if="tab.count" class="tab-count">{{ tab.count }}</span>
        </button>
      </div>

      <!-- TAB: FILES -->
      <div v-if="activeTab === 'files'" class="tab-content">
        <div class="fs-toolbar">
          <div class="breadcrumb">
            <button class="bc-item" @click="navigateToPath('/')">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M10 20v-6h4v6h5v-8h3L12 3 2 12h3v8z"/></svg>
            </button>
            <template v-for="(seg, idx) in pathSegments" :key="idx">
              <span class="bc-sep">/</span>
              <button class="bc-item" @click="navigateToPath(buildPath(idx))">{{ orgStore.folderNameCache[seg] || seg }}</button>
            </template>
          </div>
          <div class="fs-actions" v-if="canWrite">
            <button class="btn-sm" @click="showNewFolderModal = true">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M10 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z"/></svg>
              {{ t('orgs.newFolder') }}
            </button>
            <label class="btn-sm btn-upload">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z"/></svg>
              {{ t('orgs.uploadFile') }}
              <input type="file" multiple style="display:none" @change="handleFileUpload" />
            </label>
          </div>
        </div>

        <!-- Key not yet initialized for this owner (org created before encryption was added) -->
        <div v-if="!orgStore.currentOrg?.my_encrypted_org_key && canManage" class="key-init-banner">
          <div class="key-init-icon">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor"><path d="M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zm-6 9c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2zm3.1-9H8.9V6c0-1.71 1.39-3.1 3.1-3.1 1.71 0 3.1 1.39 3.1 3.1v2z"/></svg>
          </div>
          <div class="key-init-text">
            <strong>{{ t('orgs.keyNotInitialized') }}</strong>
            <span>{{ t('orgs.keyNotInitializedHint') }}</span>
          </div>
          <button class="btn-init-key" @click="handleInitKey" :disabled="initializingKey">
            <span v-if="initializingKey" class="spinner-sm"></span>
            <span v-else>{{ t('orgs.initKey') }}</span>
          </button>
        </div>

        <!-- Active upload progress -->
        <div v-if="Object.keys(uploadProgress).length > 0" class="upload-queue">
          <div v-for="(pct, name) in uploadProgress" :key="name" class="upload-row">
            <span class="upload-name">{{ name }}</span>
            <div class="upload-bar-track">
              <div class="upload-bar-fill" :style="{ width: pct + '%' }"></div>
            </div>
            <span class="upload-pct">{{ pct }}%</span>
          </div>
        </div>

        <div v-if="orgStore.loading" class="loading-inline">
          <div class="spinner-sm-dark"></div>
        </div>
        <div v-else>
          <div v-if="orgStore.currentItems.folders.length === 0 && orgStore.currentItems.files.length === 0" class="empty-folder">
            <svg viewBox="0 0 24 24" width="48" height="48" fill="currentColor" style="opacity:0.25"><path d="M20 6h-8l-2-2H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2zm0 12H4V8h16v10z"/></svg>
            <p>{{ t('orgs.emptyFolder') }}</p>
          </div>

          <div class="items-list">
            <!-- Folders -->
            <div
              v-for="folder in orgStore.currentItems.folders"
              :key="'f-' + folder.id"
              class="item-row folder-row"
              @click="navigateToPath(folder.path)"
            >
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor" class="item-icon folder-icon"><path d="M10 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z"/></svg>
              <span class="item-name">{{ folder.name }}</span>
              <span class="item-meta">{{ formatDate(folder.created_at) }}</span>
              <div class="item-actions">
                <button v-if="canManage || isGroupAdmin" class="btn-icon" @click.stop="openAccessDialog(folder)" :title="t('orgs.manageAccess')">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zm-6 9c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2zm3.1-9H8.9V6c0-1.71 1.39-3.1 3.1-3.1 1.71 0 3.1 1.39 3.1 3.1v2z"/></svg>
                </button>
                <button v-if="canWrite" class="btn-icon-danger" @click.stop="confirmDeleteFolder(folder)" :title="t('orgs.deleteFolder')">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"/></svg>
                </button>
              </div>
            </div>

            <!-- Files -->
            <div
              v-for="file in orgStore.currentItems.files"
              :key="'file-' + file.id"
              class="item-row"
            >
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor" class="item-icon file-icon"><path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z"/></svg>
              <span class="item-name">{{ file.name }}</span>
              <span class="item-meta">{{ formatSize(file.size) }} · {{ formatDate(file.created_at) }}</span>
              <div class="item-actions">
                <button class="btn-icon" @click.stop="handleDownload(file)" :title="t('file.download')">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M19 9h-4V3H9v6H5l7 7 7-7zM5 18v2h14v-2H5z"/></svg>
                </button>
                <button v-if="canWrite" class="btn-icon-danger" @click.stop="confirmDeleteFile(file)" :title="t('orgs.deleteFile')">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"/></svg>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- TAB: GROUPS -->
      <div v-if="activeTab === 'groups'" class="tab-content">
        <OrgGroupsPanel :orgID="orgID" />
      </div>

      <!-- TAB: PROFILE -->
      <div v-if="activeTab === 'profile'" class="tab-content">
        <div class="section-header">
          <h3>{{ t('orgs.myProfile') }}</h3>
        </div>
        <div class="profile-section">
          <div class="profile-label">{{ t('orgs.myRole') }}</div>
          <div class="profile-role-row">
            <div class="member-avatar">{{ (authStore.user?.name || authStore.user?.email || '?').charAt(0).toUpperCase() }}</div>
            <div class="member-info">
              <div class="member-name">{{ authStore.user?.name || authStore.user?.email }}</div>
              <div class="member-email">{{ authStore.user?.email }}</div>
            </div>
            <span class="role-badge" :class="orgStore.currentOrg.my_role">{{ t(`orgs.${orgStore.currentOrg.my_role}`) }}</span>
          </div>
        </div>
        <div class="profile-section">
          <div class="profile-label">{{ t('orgs.myGroups') }}</div>
          <div v-if="orgStore.loading" class="loading-inline"><div class="spinner-sm-dark"></div></div>
          <div v-else-if="orgStore.myGroups.length === 0" class="profile-empty">{{ t('orgs.notInAnyGroup') }}</div>
          <div v-else class="profile-groups-list">
            <div v-for="g in orgStore.myGroups" :key="g.id" class="profile-group-row">
              <div class="group-avatar-sm">{{ g.name.charAt(0).toUpperCase() }}</div>
              <div class="member-info">
                <div class="member-name">{{ g.name }}</div>
                <div v-if="g.description" class="member-email">{{ g.description }}</div>
              </div>
              <span class="role-badge" :class="'group-' + g.my_role">
                {{ t(`orgs.group${capitalize(g.my_role)}`) }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- TAB: MEMBERS -->
      <div v-if="activeTab === 'members'" class="tab-content">
        <div class="section-header">
          <h3>{{ t('orgs.members') }}</h3>
        </div>
        <div v-if="orgStore.loading" class="loading-inline"><div class="spinner-sm-dark"></div></div>
        <div v-else class="members-list">
          <div v-for="m in orgStore.members" :key="m.user_id" class="member-row">
            <div class="member-avatar">{{ (m.name || m.email || '?').charAt(0).toUpperCase() }}</div>
            <div class="member-info">
              <div class="member-name">{{ m.name || m.email }}</div>
              <div class="member-email">{{ m.email }}</div>
            </div>
            <div class="member-meta">
              <!-- Key missing indicator — visible to admins/owner only -->
              <span
                v-if="canManage && !m.encrypted_org_key && m.user_id !== myUserID"
                class="key-missing-badge"
                :title="t('orgs.keyMissing')"
              >
                <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zm-6 9c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2zm3.1-9H8.9V6c0-1.71 1.39-3.1 3.1-3.1 1.71 0 3.1 1.39 3.1 3.1v2z"/></svg>
                {{ t('orgs.needAdminKey') }}
              </span>
              <button
                v-if="canManage && !m.encrypted_org_key && m.public_key && m.user_id !== myUserID"
                class="btn-sm btn-provision"
                @click="handleProvisionKey(m)"
                :disabled="provisioningKey === m.id"
                :title="t('orgs.provisionKey')"
              >
                <span v-if="provisioningKey === m.id" class="spinner-sm-dark"></span>
                <span v-else>{{ t('orgs.provisionKey') }}</span>
              </button>
              <select
                v-if="canManage && m.role !== 'owner' && m.user_id !== myUserID"
                class="role-select"
                :value="m.role"
                @change="handleRoleChange(m, $event.target.value)"
              >
                <option value="admin">{{ t('orgs.admin') }}</option>
                <option value="member">{{ t('orgs.member') }}</option>
                <option value="viewer">{{ t('orgs.viewer') }}</option>
              </select>
              <span v-else class="role-badge" :class="m.role">{{ t(`orgs.${m.role}`) }}</span>
              <button
                v-if="(canManage && m.role !== 'owner' && m.user_id !== myUserID) || m.user_id === myUserID"
                class="btn-icon-danger"
                @click="handleRemoveMember(m)"
                :title="t('orgs.removeFromOrg')"
              >
                <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- TAB: INVITATIONS -->
      <div v-if="activeTab === 'invitations'" class="tab-content">
        <div class="section-header">
          <h3>{{ t('orgs.invitations') }}</h3>
          <button v-if="canManage" class="btn-primary btn-sm-action" @click="showInviteModal = true">
            + {{ t('orgs.createInvite') }}
          </button>
        </div>

        <div v-if="orgStore.loading" class="loading-inline"><div class="spinner-sm-dark"></div></div>
        <div v-else-if="orgStore.invitations.length === 0" class="empty-tab">
          <p>{{ t('orgs.noInvitations') }}</p>
        </div>
        <div v-else class="invitations-list">
          <div v-for="inv in orgStore.invitations" :key="inv.id" class="invite-row">
            <div class="invite-info">
              <div class="invite-token">
                <code>{{ inviteURL(inv.token) }}</code>
                <button class="btn-copy" @click="copyInviteLink(inv.token)" :title="t('orgs.copyInviteLink')">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M16 1H4c-1.1 0-2 .9-2 2v14h2V3h12V1zm3 4H8c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h11c1.1 0 2-.9 2-2V7c0-1.1-.9-2-2-2zm0 16H8V7h11v14z"/></svg>
                </button>
              </div>
              <div class="invite-meta">
                <span class="role-badge" :class="inv.role">{{ t(`orgs.${inv.role}`) }}</span>
                <span class="invite-detail">{{ inv.uses }}/{{ inv.max_uses || '∞' }} uses</span>
                <span v-if="inv.expires_at" class="invite-detail">expires {{ formatDate(inv.expires_at) }}</span>
              </div>
            </div>
            <button v-if="canManage" class="btn-icon-danger" @click="handleRevokeInvite(inv)" :title="t('orgs.revokeInvite')">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
            </button>
          </div>
        </div>
      </div>

      <!-- TAB: PERMISSIONS -->
      <div v-if="activeTab === 'permissions'" class="tab-content">
        <div class="section-header">
          <h3>{{ t('orgs.permissions') }}</h3>
          <button v-if="canManage" class="btn-primary btn-sm-action" @click="showPermModal = true">
            + {{ t('orgs.setPermission') }}
          </button>
        </div>

        <div v-if="orgStore.loading" class="loading-inline"><div class="spinner-sm-dark"></div></div>
        <div v-else-if="orgStore.permissions.length === 0" class="empty-tab">
          <p>{{ t('orgs.noPermissions') }}</p>
        </div>
        <div v-else class="permissions-list">
          <div v-for="perm in orgStore.permissions" :key="perm.id" class="perm-row">
            <div class="perm-info">
              <code class="perm-path">{{ perm.folder_path }}</code>
              <span class="perm-user">{{ orgStore.members.find(m => m.user_id === perm.user_id)?.name || orgStore.members.find(m => m.user_id === perm.user_id)?.email || perm.user_id }}</span>
            </div>
            <div class="perm-level">
              <span class="level-badge" :class="perm.level">{{ t(`orgs.perm${capitalize(perm.level)}`) }}</span>
            </div>
            <button v-if="canManage" class="btn-icon-danger" @click="handleDeletePerm(perm)" :title="t('orgs.deletePermission')">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"/></svg>
            </button>
          </div>
        </div>
      </div>

      <!-- TAB: AUDIT LOG -->
      <div v-if="activeTab === 'audit'" class="tab-content">
        <div class="section-header">
          <h3>{{ t('orgs.auditLog') }}</h3>
          <div style="display:flex;gap:8px;align-items:center">
            <span class="audit-retention-note">{{ t('orgs.auditRetentionNote') }}</span>
            <button class="btn-sm" @click="refreshAudit">{{ t('orgs.refresh') }}</button>
            <button v-if="canManage" class="btn-sm btn-clean" @click="openCleanModal">
              <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"/></svg>
              {{ t('orgs.cleanAudit') }}
            </button>
          </div>
        </div>
        <div v-if="orgStore.auditLog.length === 0" class="empty-tab">
          <p>{{ t('orgs.noAuditEvents') }}</p>
        </div>
        <div v-else>
          <div class="audit-list">
            <div v-for="entry in orgStore.auditLog" :key="entry.id" class="audit-row">
              <div class="audit-action">
                <span class="audit-badge" :class="entry.action">{{ t(`orgs.audit_${entry.action}`, entry.action) }}</span>
                <span v-if="entry.detail" class="audit-detail">{{ entry.detail }}</span>
              </div>
              <div class="audit-meta">
                <span class="audit-actor" :title="entry.actor_id">{{ entry.actor_id.slice(0, 8) }}</span>
                <span class="audit-time">{{ formatDate(entry.created_at) }}</span>
              </div>
            </div>
          </div>
          <div class="audit-load-more" v-if="auditHasMore">
            <button class="btn-sm" @click="loadMoreAudit" :disabled="loadingMoreAudit">
              <span v-if="loadingMoreAudit" class="spinner-sm-dark" style="width:14px;height:14px"></span>
              <span v-else>{{ t('orgs.loadMore') }}</span>
            </button>
          </div>
        </div>
      </div>

      <!-- TAB: SETTINGS -->
      <div v-if="activeTab === 'settings'" class="tab-content">
        <div class="settings-section">
          <h3>{{ t('orgs.settings') }}</h3>
          <div class="form-group">
            <label>{{ t('orgs.orgName') }}</label>
            <input v-model="settingsForm.name" class="input-field" type="text" />
          </div>
          <div class="form-group">
            <label>{{ t('orgs.orgDesc') }}</label>
            <input v-model="settingsForm.description" class="input-field" type="text" />
          </div>
          <div class="form-group" v-if="isOwner">
            <label>{{ t('orgs.storageQuotaMB') }}</label>
            <input v-model.number="settingsForm.storageQuotaMB" class="input-field" type="number" min="100" />
          </div>
          <p v-if="settingsError" class="form-error">{{ settingsError }}</p>
          <button class="btn-primary" @click="handleSaveSettings" :disabled="savingSettings">
            <span v-if="savingSettings" class="spinner-sm"></span>
            {{ t('orgs.saveSettings') }}
          </button>
        </div>

        <div class="danger-zone" v-if="isOwner">
          <h4>{{ t('orgs.dangerZone') }}</h4>
          <div class="danger-action">
            <div>
              <strong>{{ t('orgs.rotateKeyTitle') }}</strong>
              <p class="hint-sm">{{ t('orgs.rotateKeyDesc') }}</p>
            </div>
            <button class="btn-danger-outline" @click="handleRotateKey" :disabled="rotatingKey">
              <span v-if="rotatingKey" class="spinner-sm"></span>
              {{ t('orgs.rotateKey') }}
            </button>
          </div>
          <div class="danger-action">
            <div>
              <strong>{{ t('orgs.deleteOrg') }}</strong>
              <p class="hint-sm">{{ t('orgs.deleteOrgConfirmHint') }}</p>
            </div>
            <button class="btn-danger" @click="handleDeleteOrg">{{ t('orgs.deleteOrg') }}</button>
          </div>
        </div>
        <div class="danger-zone" v-else-if="orgStore.currentOrg.my_role !== 'owner'">
          <h4>{{ t('orgs.dangerZone') }}</h4>
          <button class="btn-danger" @click="handleLeaveOrg">{{ t('orgs.leaveOrg') }}</button>
        </div>
      </div>
    </template>

    <!-- ── Modals ─────────────────────────────────────────────────────────── -->

    <!-- New folder modal -->
    <Transition name="modal">
      <div v-if="showNewFolderModal" class="modal-overlay" @click.self="showNewFolderModal = false">
        <div class="modal modal-sm">
          <div class="modal-header">
            <h3>{{ t('orgs.newFolder') }}</h3>
            <button class="btn-close" @click="showNewFolderModal = false">
              <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
            </button>
          </div>
          <div class="modal-body">
            <input v-model="newFolderName" class="input-field" type="text" :placeholder="t('orgs.enterFolderName')" @keydown.enter="handleCreateFolder" />
            <p v-if="folderError" class="form-error">{{ folderError }}</p>
          </div>
          <div class="modal-footer">
            <button class="btn-secondary" @click="showNewFolderModal = false">{{ t('orgs.cancel') }}</button>
            <button class="btn-primary" @click="handleCreateFolder" :disabled="!newFolderName">{{ t('orgs.create') }}</button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Create invite modal -->
    <Transition name="modal">
      <div v-if="showInviteModal" class="modal-overlay" @click.self="showInviteModal = false">
        <div class="modal">
          <div class="modal-header">
            <h3>{{ t('orgs.inviteUser') }}</h3>
            <button class="btn-close" @click="showInviteModal = false">
              <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
            </button>
          </div>
          <div class="modal-body">
            <div class="form-group">
              <label>{{ t('orgs.inviteRole') }}</label>
              <select v-model="inviteForm.role" class="input-field">
                <option value="viewer">{{ t('orgs.viewer') }}</option>
                <option value="member">{{ t('orgs.member') }}</option>
                <option v-if="isOwner" value="admin">{{ t('orgs.admin') }}</option>
              </select>
            </div>
            <div class="form-group">
              <label>{{ t('orgs.inviteMaxUses') }}</label>
              <input v-model.number="inviteForm.maxUses" class="input-field" type="number" min="0" />
            </div>
            <div class="form-group">
              <label>{{ t('orgs.inviteExpiry') }}</label>
              <input v-model="inviteForm.expiresAt" class="input-field" type="datetime-local" />
            </div>
            <p v-if="inviteError" class="form-error">{{ inviteError }}</p>
          </div>
          <div class="modal-footer">
            <button class="btn-secondary" @click="showInviteModal = false">{{ t('orgs.cancel') }}</button>
            <button class="btn-primary" @click="handleCreateInvite" :disabled="creatingInvite">
              <span v-if="creatingInvite" class="spinner-sm"></span>
              {{ t('orgs.createInvite') }}
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
            <h3>{{ t('orgs.setPermission') }}</h3>
            <button class="btn-close" @click="showPermModal = false">
              <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
            </button>
          </div>
          <div class="modal-body">
            <div class="form-group">
              <label>{{ t('orgs.targetUser') }}</label>
              <select v-model="permForm.userID" class="input-field">
                <option value="">— {{ t('orgs.targetUser') }} —</option>
                <option v-for="m in orgStore.members.filter(m => m.role !== 'owner')" :key="m.user_id" :value="m.user_id">
                  {{ m.name || m.email }}
                </option>
              </select>
            </div>
            <div class="form-group">
              <label>{{ t('orgs.folderPath') }}</label>
              <input v-model="permForm.folderPath" class="input-field" type="text" placeholder="/ (racine)" />
            </div>
            <div class="form-group">
              <label>{{ t('orgs.permissions') }}</label>
              <select v-model="permForm.level" class="input-field">
                <option value="none">{{ t('orgs.permNone') }}</option>
                <option value="read">{{ t('orgs.permRead') }}</option>
                <option value="write">{{ t('orgs.permWrite') }}</option>
                <option value="manage">{{ t('orgs.permManage') }}</option>
              </select>
            </div>
            <p v-if="permError" class="form-error">{{ permError }}</p>
          </div>
          <div class="modal-footer">
            <button class="btn-secondary" @click="showPermModal = false">{{ t('orgs.cancel') }}</button>
            <button class="btn-primary" @click="handleSetPerm" :disabled="!permForm.userID || settingPerm">
              <span v-if="settingPerm" class="spinner-sm"></span>
              {{ t('common.save') }}
            </button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Clean audit modal -->
    <Transition name="modal">
      <div v-if="showCleanModal" class="modal-overlay" @click.self="showCleanModal = false">
        <div class="modal modal-clean">
          <div class="modal-header">
            <h3>{{ t('orgs.cleanAuditTitle') }}</h3>
            <button class="btn-close" @click="showCleanModal = false">
              <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
            </button>
          </div>
          <div class="modal-body">
            <!-- Mode tabs -->
            <div class="clean-tabs">
              <button class="clean-tab" :class="{ active: cleanMode === 'all' }" @click="cleanMode = 'all'">{{ t('orgs.cleanAll') }}</button>
              <button class="clean-tab" :class="{ active: cleanMode === 'months' }" @click="cleanMode = 'months'">{{ t('orgs.cleanByMonth') }}</button>
              <button class="clean-tab" :class="{ active: cleanMode === 'days' }" @click="cleanMode = 'days'">{{ t('orgs.cleanByDay') }}</button>
            </div>

            <!-- All mode -->
            <div v-if="cleanMode === 'all'" class="clean-panel">
              <p class="clean-warn">{{ t('orgs.confirmDeleteAllAudit') }}</p>
            </div>

            <!-- By month mode -->
            <div v-else-if="cleanMode === 'months'" class="clean-panel">
              <div v-if="availableMonths.length === 0" class="clean-empty">{{ t('orgs.noAuditMonth') }}</div>
              <div v-else class="month-grid">
                <button
                  v-for="m in availableMonths"
                  :key="m"
                  class="month-chip"
                  :class="{ selected: selectedMonths.includes(m) }"
                  @click="toggleMonth(m)"
                >
                  <span class="month-label">{{ formatMonthLabel(m) }}</span>
                  <span class="month-count">{{ monthCount(m) }}</span>
                </button>
              </div>
              <p v-if="selectedMonths.length" class="clean-selection-hint">{{ t('orgs.selectedMonths', { count: selectedMonths.length }) }}</p>
            </div>

            <!-- By day mode -->
            <div v-else class="clean-panel">
              <div class="cal-nav">
                <button class="cal-nav-btn" :disabled="!canGoPrevMonth" @click="prevCalMonth">
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M15.41 7.41L14 6l-6 6 6 6 1.41-1.41L10.83 12z"/></svg>
                </button>
                <span class="cal-month-label">{{ calMonthLabel }}</span>
                <button class="cal-nav-btn" :disabled="!canGoNextMonth" @click="nextCalMonth">
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M10 6L8.59 7.41 13.17 12l-4.58 4.59L10 18l6-6z"/></svg>
                </button>
              </div>
              <div class="cal-grid">
                <div class="cal-dow" v-for="d in calDowLabels" :key="d">{{ d }}</div>
                <div
                  v-for="(cell, i) in calGrid"
                  :key="i"
                  class="cal-cell"
                  :class="{
                    'cal-empty': !cell,
                    'cal-has': cell && cell.count > 0,
                    'cal-selected': cell && selectedDays.includes(cell.dateStr),
                    'cal-none': cell && cell.count === 0,
                  }"
                  @click="cell && toggleDay(cell.dateStr)"
                >
                  <span v-if="cell" class="cal-day-num">{{ cell.day }}</span>
                  <span v-if="cell && cell.count" class="cal-day-count">{{ cell.count }}</span>
                </div>
              </div>
              <p v-if="selectedDays.length" class="clean-selection-hint">{{ t('orgs.selectedDays', { count: selectedDays.length }) }}</p>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn-secondary" @click="showCleanModal = false">{{ t('orgs.cancel') }}</button>
            <button
              class="btn-danger"
              :disabled="cleaningAudit || (cleanMode === 'months' && !selectedMonths.length) || (cleanMode === 'days' && !selectedDays.length)"
              @click="handleCleanAudit"
            >
              <span v-if="cleaningAudit" class="spinner-sm"></span>
              {{ t('orgs.cleanDeleteBtn') }}
            </button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Toast notification -->
    <Transition name="toast">
      <div v-if="toast" class="toast" :class="toast.type">{{ toast.message }}</div>
    </Transition>

    <!-- Folder access dialog -->
    <OrgFolderAccessDialog
      v-model="showAccessDialog"
      :orgID="orgID"
      :folder="accessDialogFolder"
      :canManage="canManage"
    />

    <!-- Onboarding wizard (first org creation only) -->
    <OrgOnboardingWizard
      v-if="showOnboardingWizard && orgStore.currentOrg"
      :orgId="orgID"
      :orgName="orgStore.currentOrg.name"
      @close="showOnboardingWizard = false"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch, h } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useOrgStore } from '../stores/organizations'
import { useAuthStore } from '../stores/auth'
import { useRealtimeStore } from '../stores/realtime'
import OrgGroupsPanel from '../components/organizations/OrgGroupsPanel.vue'
import OrgFolderAccessDialog from '../components/organizations/OrgFolderAccessDialog.vue'
import OrgOnboardingWizard from '../components/organizations/OrgOnboardingWizard.vue'

const { t, locale } = useI18n()
const route = useRoute()
const router = useRouter()
const orgStore = useOrgStore()
const authStore = useAuthStore()
const realtimeStore = useRealtimeStore()

const activeTab = ref('files')
const currentPath = ref('/')
const toast = ref(null)

const showAccessDialog = ref(false)
const accessDialogFolder = ref(null)

// ── Audit log pagination & cleanup ───────────────────────────────────────────

const auditPage = ref(1)
const auditHasMore = ref(false)
const loadingMoreAudit = ref(false)

const showCleanModal = ref(false)
const cleanMode = ref('all')
const selectedMonths = ref([])
const selectedDays = ref([])
const cleaningAudit = ref(false)
const calYear = ref(new Date().getFullYear())
const calMonth = ref(new Date().getMonth())

// ── Tab config ────────────────────────────────────────────────────────────────

const TabIcon = (paths) => ({ render: () => h('svg', { viewBox: '0 0 24 24', width: 18, height: 18, fill: 'currentColor' }, paths.map(d => h('path', { d }))) })

const tabs = computed(() => {
  const base = [
    { key: 'files', label: t('orgs.files'), icon: TabIcon(['M20 6h-8l-2-2H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2zm0 12H4V8h16v10z']) },
    { key: 'members', label: t('orgs.members'), icon: TabIcon(['M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z']), count: orgStore.members.length || null },
    { key: 'profile', label: t('orgs.myProfile'), icon: TabIcon(['M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z']) },
  ]
  if (canManage.value || isGroupAdmin.value) {
    base.push(
      { key: 'groups', label: t('orgs.groups'), icon: TabIcon(['M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z', 'M23 14h-2v-2h-2v2h-2v2h2v2h2v-2h2z']) },
    )
  }
  if (canManage.value) {
    base.push(
      { key: 'invitations', label: t('orgs.invitations'), icon: TabIcon(['M20 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V6c0-1.1-.9-2-2-2zm0 4l-8 5-8-5V6l8 5 8-5v2z']) },
      { key: 'permissions', label: t('orgs.permissions'), icon: TabIcon(['M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zm-6 9c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2zm3.1-9H8.9V6c0-1.71 1.39-3.1 3.1-3.1 1.71 0 3.1 1.39 3.1 3.1v2z']) },
      { key: 'audit', label: t('orgs.auditLog'), icon: TabIcon(['M9 11H7v2h2v-2zm4 0h-2v2h2v-2zm4 0h-2v2h2v-2zm2-7h-1V2h-2v2H8V2H6v2H5c-1.11 0-1.99.9-1.99 2L3 20c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2V6c0-1.1-.9-2-2-2zm0 16H5V9h14v11z']) },
      { key: 'settings', label: t('orgs.settings'), icon: TabIcon(['M19.14,12.94c0.04-0.3,0.06-0.61,0.06-0.94c0-0.32-0.02-0.64-0.07-0.94l2.03-1.58c0.18-0.14,0.23-0.41,0.12-0.61 l-1.92-3.32c-0.12-0.22-0.37-0.29-0.59-0.22l-2.39,0.96c-0.5-0.38-1.03-0.7-1.62-0.94L14.4,2.81c-0.04-0.24-0.24-0.41-0.48-0.41 h-3.84c-0.24,0-0.43,0.17-0.47,0.41L9.25,5.35C8.66,5.59,8.12,5.92,7.63,6.29L5.24,5.33c-0.22-0.08-0.47,0-0.59,0.22L2.74,8.87 C2.62,9.08,2.66,9.34,2.86,9.48l2.03,1.58C4.84,11.36,4.8,11.69,4.8,12s0.02,0.64,0.07,0.94l-2.03,1.58 c-0.18,0.14-0.23,0.41-0.12,0.61l1.92,3.32c0.12,0.22,0.37,0.29,0.59,0.22l2.39-0.96c0.5,0.38,1.03,0.7,1.62,0.94l0.36,2.54 c0.05,0.24,0.24,0.41,0.48,0.41h3.84c0.24,0,0.44-0.17,0.47-0.41l0.36-2.54c0.59-0.24,1.13-0.56,1.62-0.94l2.39,0.96 c0.22,0.08,0.47,0,0.59-0.22l1.92-3.32c0.12-0.22,0.07-0.47-0.12-0.61L19.14,12.94z M12,15.6c-1.98,0-3.6-1.62-3.6-3.6 s1.62-3.6,3.6-3.6s3.6,1.62,3.6,3.6S13.98,15.6,12,15.6z']) },
    )
  }
  return base
})

// ── Computed ──────────────────────────────────────────────────────────────────

const orgID = computed(() => parseInt(route.params.orgID))
const myUserID = computed(() => authStore.user?.id || authStore.user?.user_id)
const isOwner = computed(() => orgStore.currentOrg?.my_role === 'owner')
const canManage = computed(() => ['owner', 'admin'].includes(orgStore.currentOrg?.my_role))
const canWrite = computed(() => ['owner', 'admin', 'member'].includes(orgStore.currentOrg?.my_role))
const isGroupAdmin = computed(() => orgStore.currentOrg?.is_group_admin === true)

const storagePercent = computed(() => {
  const org = orgStore.currentOrg
  if (!org) return 0
  const quota = org.storage_quota_mb * 1024 * 1024
  if (!quota) return 0
  return Math.min(100, (org.storage_used_bytes / quota) * 100)
})

const storageClass = computed(() => {
  const p = storagePercent.value
  if (p >= 90) return 'critical'
  if (p >= 75) return 'warning'
  return 'ok'
})

const pathSegments = computed(() => {
  if (currentPath.value === '/') return []
  return currentPath.value.replace(/^\//, '').split('/').filter(s => s)
})

// ── Audit calendar computed ───────────────────────────────────────────────────

const auditSummaryDays = computed(() => orgStore.auditSummary || {})

const availableMonths = computed(() => {
  const months = new Set()
  for (const day of Object.keys(auditSummaryDays.value)) {
    months.add(day.slice(0, 7))
  }
  return [...months].sort().reverse()
})

const calMonthLabel = computed(() => {
  return new Date(calYear.value, calMonth.value, 1).toLocaleDateString(locale.value, { month: 'long', year: 'numeric' })
})

const calDowLabels = computed(() => {
  // Monday-based day names from locale
  const days = []
  for (let i = 1; i <= 7; i++) {
    const d = new Date(2024, 0, i) // Jan 1 2024 is Monday
    days.push(d.toLocaleDateString(locale.value, { weekday: 'short' }).slice(0, 2))
  }
  return days
})

const calGrid = computed(() => {
  const year = calYear.value
  const month = calMonth.value
  const daysInMonth = new Date(year, month + 1, 0).getDate()
  let startDow = (new Date(year, month, 1).getDay() + 6) % 7 // Monday = 0
  const cells = []
  for (let i = 0; i < startDow; i++) cells.push(null)
  for (let d = 1; d <= daysInMonth; d++) {
    const ds = `${year}-${String(month + 1).padStart(2, '0')}-${String(d).padStart(2, '0')}`
    cells.push({ day: d, dateStr: ds, count: auditSummaryDays.value[ds] || 0 })
  }
  return cells
})

const canGoPrevMonth = computed(() => {
  const limit = new Date()
  limit.setFullYear(limit.getFullYear() - 1)
  return new Date(calYear.value, calMonth.value, 1) > limit
})

const canGoNextMonth = computed(() => {
  const now = new Date()
  return !(calYear.value === now.getFullYear() && calMonth.value === now.getMonth())
})

const monthCount = (m) =>
  Object.entries(auditSummaryDays.value)
    .filter(([day]) => day.startsWith(m))
    .reduce((sum, [, c]) => sum + c, 0)

const formatMonthLabel = (m) => {
  const [year, month] = m.split('-')
  return new Date(parseInt(year), parseInt(month) - 1, 1).toLocaleDateString(locale.value, { month: 'long', year: 'numeric' })
}

// ── Init ──────────────────────────────────────────────────────────────────────

let _unsubOrgUpdate = null

const showOnboardingWizard = ref(false)

onMounted(async () => {
  await orgStore.fetchOrg(orgID.value)
  await orgStore.fetchItems(orgID.value, '/')
  await orgStore.fetchMembers(orgID.value)

  settingsForm.value = {
    name: orgStore.currentOrg.name,
    description: orgStore.currentOrg.description,
    storageQuotaMB: orgStore.currentOrg.storage_quota_mb,
  }

  const pendingOrgId = localStorage.getItem('kagibi_org_onboarding')
  if (pendingOrgId && parseInt(pendingOrgId) === orgID.value) {
    localStorage.removeItem('kagibi_org_onboarding')
    showOnboardingWizard.value = true
  }

  // Refresh members list when someone joins or leaves this org
  _unsubOrgUpdate = realtimeStore.onEvent('org_update', (payload) => {
    if (payload?.org_id !== orgID.value) return
    orgStore.fetchMembers(orgID.value)
  })
})

onUnmounted(() => {
  if (_unsubOrgUpdate) _unsubOrgUpdate()
})

const orgAdminTabs = new Set(['invitations', 'permissions', 'audit', 'settings'])

const switchTab = async (tab) => {
  if (orgAdminTabs.has(tab) && !canManage.value) return
  if (tab === 'groups' && !canManage.value && !isGroupAdmin.value) return
  activeTab.value = tab
  if (tab === 'profile') await orgStore.fetchMyGroups(orgID.value)
  if (tab === 'invitations' && orgStore.invitations.length === 0) await orgStore.fetchInvitations(orgID.value)
  if (tab === 'permissions') await orgStore.fetchPermissions(orgID.value)
  if (tab === 'audit') {
    auditPage.value = 1
    auditHasMore.value = false
    const entries = await orgStore.fetchAuditLog(orgID.value, 1)
    auditHasMore.value = entries.length === 50
  }
}

// ── File system ───────────────────────────────────────────────────────────────

const navigateToPath = async (path) => {
  const prevPath = currentPath.value
  currentPath.value = path || '/'
  try {
    await orgStore.fetchItems(orgID.value, currentPath.value)
  } catch (e) {
    currentPath.value = prevPath
    if (e.response?.status === 403) {
      showToast(t('orgs.noReadAccess'), 'error')
    }
  }
}

const openAccessDialog = (folder) => {
  accessDialogFolder.value = { path: folder.path, name: folder.name }
  showAccessDialog.value = true
}

const buildPath = (idx) => {
  const segs = pathSegments.value.slice(0, idx + 1)
  return '/' + segs.join('/')
}

// New folder
const showNewFolderModal = ref(false)
const newFolderName = ref('')
const folderError = ref('')

const handleCreateFolder = async () => {
  if (!newFolderName.value) return
  folderError.value = ''
  try {
    await orgStore.createFolder(orgID.value, newFolderName.value, currentPath.value)
    showNewFolderModal.value = false
    newFolderName.value = ''
    showToast(t('orgs.folderCreated'))
  } catch (e) {
    folderError.value = e.response?.data?.error || e.message
  }
}

// Upload
const uploadProgress = ref({}) // fileName -> 0-100

const handleFileUpload = async (event) => {
  const files = Array.from(event.target.files)
  event.target.value = ''
  for (const file of files) {
    await uploadFile(file)
  }
}

const uploadFile = async (file) => {
  uploadProgress.value[file.name] = 0
  try {
    await orgStore.uploadOrgFile(orgID.value, file, currentPath.value, (p) => {
      uploadProgress.value[file.name] = p
    })
    showToast(file.name + ' importé')
  } catch (e) {
    showToast(e.response?.data?.error || e.message, 'error')
  } finally {
    delete uploadProgress.value[file.name]
  }
}

// Delete file
const confirmDeleteFile = async (file) => {
  if (!confirm(t('orgs.confirmDeleteFile', { name: file.name }))) return
  try {
    await orgStore.deleteFile(orgID.value, file.id)
    showToast(t('orgs.fileDeleted'))
  } catch (e) {
    showToast(e.response?.data?.error || e.message, 'error')
  }
}

// Delete folder
const confirmDeleteFolder = async (folder) => {
  if (!confirm(t('orgs.confirmDeleteFolder', { name: folder.name }))) return
  try {
    await orgStore.deleteFolder(orgID.value, folder.id)
    showToast(t('orgs.folderDeleted'))
  } catch (e) {
    showToast(e.response?.data?.error || e.message, 'error')
  }
}

// Download
const handleDownload = async (file) => {
  try {
    await orgStore.downloadFile(orgID.value, file.id, file.name, file.mime_type)
  } catch (e) {
    showToast(e.response?.data?.error || e.message, 'error')
  }
}

// ── Members ───────────────────────────────────────────────────────────────────

// ── Logo upload ───────────────────────────────────────────────────────────────

const logoInputRef = ref(null)
const uploadingLogo = ref(false)

const handleLogoChange = async (event) => {
  const file = event.target.files?.[0]
  event.target.value = ''
  if (!file) return
  uploadingLogo.value = true
  try {
    await orgStore.uploadOrgLogo(orgID.value, file)
    showToast(t('orgs.logoUpdated'))
  } catch (e) {
    showToast(e.response?.data?.error || e.message, 'error')
  } finally {
    uploadingLogo.value = false
  }
}

const handleRemoveLogo = async () => {
  if (!confirm(t('orgs.removeLogoConfirm'))) return
  try {
    await orgStore.deleteOrgLogo(orgID.value)
    showToast(t('orgs.logoRemoved'))
  } catch (e) {
    showToast(e.response?.data?.error || e.message, 'error')
  }
}

// ── Key initialization (owner with no key) ────────────────────────────────────

const initializingKey = ref(false)

const handleInitKey = async () => {
  initializingKey.value = true
  try {
    await orgStore.initializeOrgKey(orgID.value)
    showToast(t('orgs.keyInitialized'))
  } catch (e) {
    showToast(e.response?.data?.error || e.message, 'error')
  } finally {
    initializingKey.value = false
  }
}

// ── Key provisioning ──────────────────────────────────────────────────────────

const provisioningKey = ref(null) // member.id being provisioned

const handleProvisionKey = async (member) => {
  provisioningKey.value = member.id
  try {
    await orgStore.provisionMemberKey(orgID.value, member)
    showToast(t('orgs.keyProvisioned'))
  } catch (e) {
    showToast(e.response?.data?.error || e.message, 'error')
  } finally {
    provisioningKey.value = null
  }
}

// ── Members ───────────────────────────────────────────────────────────────────

const handleRoleChange = async (member, role) => {
  try {
    await orgStore.updateMemberRole(orgID.value, member.id, role)
    showToast(t('orgs.role') + ' mis à jour')
  } catch (e) {
    showToast(e.response?.data?.error || e.message, 'error')
  }
}

const handleRemoveMember = async (member) => {
  if (!confirm(t('orgs.confirmRemoveMember'))) return
  try {
    await orgStore.removeMember(orgID.value, member.id)
    showToast(t('orgs.memberRemoved'))
    if (member.user_id === myUserID.value) {
      router.push('/dashboard/organizations')
    }
  } catch (e) {
    showToast(e.response?.data?.error || e.message, 'error')
  }
}

// ── Invitations ───────────────────────────────────────────────────────────────

const showInviteModal = ref(false)
const creatingInvite = ref(false)
const inviteError = ref('')
const inviteForm = ref({ role: 'member', maxUses: 0, expiresAt: '' })

const handleCreateInvite = async () => {
  creatingInvite.value = true
  inviteError.value = ''
  try {
    const payload = {
      role: inviteForm.value.role,
      max_uses: inviteForm.value.maxUses,
    }
    if (inviteForm.value.expiresAt) {
      payload.expires_at = new Date(inviteForm.value.expiresAt).toISOString()
    }
    await orgStore.createInvitation(orgID.value, payload)
    showInviteModal.value = false
    inviteForm.value = { role: 'member', maxUses: 0, expiresAt: '' }
    showToast(t('orgs.inviteCreated'))
  } catch (e) {
    inviteError.value = e.response?.data?.error || e.message
  } finally {
    creatingInvite.value = false
  }
}

const handleRevokeInvite = async (inv) => {
  try {
    await orgStore.revokeInvitation(orgID.value, inv.id)
    showToast(t('orgs.inviteRevoked'))
  } catch (e) {
    showToast(e.response?.data?.error || e.message, 'error')
  }
}

const inviteURL = (token) => {
  return `${window.location.origin}/join/${token}`
}

const copyInviteLink = (token) => {
  navigator.clipboard.writeText(inviteURL(token))
  showToast(t('orgs.inviteLinkCopied'))
}

// ── Permissions ───────────────────────────────────────────────────────────────

const showPermModal = ref(false)
const settingPerm = ref(false)
const permError = ref('')
const permForm = ref({ userID: '', folderPath: '/', level: 'read' })

const handleSetPerm = async () => {
  if (!permForm.value.userID) return
  settingPerm.value = true
  permError.value = ''
  try {
    await orgStore.setPermission(orgID.value, {
      user_id: permForm.value.userID,
      folder_path: permForm.value.folderPath || '/',
      level: permForm.value.level,
    })
    showPermModal.value = false
    permForm.value = { userID: '', folderPath: '/', level: 'read' }
    showToast(t('common.success'))
  } catch (e) {
    permError.value = e.response?.data?.error || e.message
  } finally {
    settingPerm.value = false
  }
}

const handleDeletePerm = async (perm) => {
  try {
    await orgStore.deletePermission(orgID.value, perm.user_id, perm.folder_path)
    showToast(t('common.success'))
  } catch (e) {
    showToast(e.response?.data?.error || e.message, 'error')
  }
}

// ── Settings ──────────────────────────────────────────────────────────────────

const settingsForm = ref({ name: '', description: '', storageQuotaMB: 10240 })
const savingSettings = ref(false)
const settingsError = ref('')

const handleSaveSettings = async () => {
  savingSettings.value = true
  settingsError.value = ''
  try {
    const payload = { name: settingsForm.value.name, description: settingsForm.value.description }
    if (isOwner.value) payload.storage_quota_mb = settingsForm.value.storageQuotaMB
    await orgStore.updateOrg(orgID.value, payload)
    showToast(t('common.success'))
  } catch (e) {
    settingsError.value = e.response?.data?.error || e.message
  } finally {
    savingSettings.value = false
  }
}

const handleLeaveOrg = async () => {
  if (!confirm(t('orgs.leaveOrgConfirm', { name: orgStore.currentOrg.name }))) return
  const myMember = orgStore.members.find(m => m.user_id === myUserID.value)
  if (!myMember) return
  try {
    await orgStore.removeMember(orgID.value, myMember.id)
    showToast(t('orgs.orgLeft'))
    router.push('/dashboard/organizations')
  } catch (e) {
    showToast(e.response?.data?.error || e.message, 'error')
  }
}

const rotatingKey = ref(false)

const handleRotateKey = async () => {
  if (!confirm(t('orgs.rotateKeyConfirm'))) return
  rotatingKey.value = true
  try {
    await orgStore.rotateOrgKey(orgID.value)
    showToast(t('orgs.keyRotated'))
  } catch (e) {
    showToast(e.response?.data?.error || e.message, 'error')
  } finally {
    rotatingKey.value = false
  }
}

const handleDeleteOrg = async () => {
  if (!confirm(t('orgs.deleteOrgConfirm', { name: orgStore.currentOrg.name }))) return
  try {
    await orgStore.deleteOrg(orgID.value)
    showToast(t('orgs.orgDeleted'))
    router.push('/dashboard/organizations')
  } catch (e) {
    showToast(e.response?.data?.error || e.message, 'error')
  }
}

// ── Audit log actions ─────────────────────────────────────────────────────────

const refreshAudit = async () => {
  auditPage.value = 1
  auditHasMore.value = false
  const entries = await orgStore.fetchAuditLog(orgID.value, 1)
  auditHasMore.value = entries.length === 50
}

const loadMoreAudit = async () => {
  loadingMoreAudit.value = true
  try {
    auditPage.value++
    const entries = await orgStore.fetchAuditLog(orgID.value, auditPage.value)
    auditHasMore.value = entries.length === 50
  } finally {
    loadingMoreAudit.value = false
  }
}

const openCleanModal = async () => {
  cleanMode.value = 'all'
  selectedMonths.value = []
  selectedDays.value = []
  calYear.value = new Date().getFullYear()
  calMonth.value = new Date().getMonth()
  await orgStore.fetchAuditSummary(orgID.value)
  showCleanModal.value = true
}

const toggleMonth = (m) => {
  const idx = selectedMonths.value.indexOf(m)
  if (idx >= 0) selectedMonths.value.splice(idx, 1)
  else selectedMonths.value.push(m)
}

const toggleDay = (dateStr) => {
  if (!auditSummaryDays.value[dateStr]) return
  const idx = selectedDays.value.indexOf(dateStr)
  if (idx >= 0) selectedDays.value.splice(idx, 1)
  else selectedDays.value.push(dateStr)
}

const prevCalMonth = () => {
  if (!canGoPrevMonth.value) return
  if (calMonth.value === 0) { calMonth.value = 11; calYear.value-- }
  else calMonth.value--
}

const nextCalMonth = () => {
  if (!canGoNextMonth.value) return
  if (calMonth.value === 11) { calMonth.value = 0; calYear.value++ }
  else calMonth.value++
}

const handleCleanAudit = async () => {
  let payload
  if (cleanMode.value === 'all') {
    payload = { mode: 'all' }
  } else if (cleanMode.value === 'months') {
    if (!selectedMonths.value.length) return
    payload = { mode: 'months', months: [...selectedMonths.value] }
  } else {
    if (!selectedDays.value.length) return
    payload = { mode: 'days', days: [...selectedDays.value] }
  }
  cleaningAudit.value = true
  try {
    const res = await orgStore.deleteAuditLog(orgID.value, payload)
    showCleanModal.value = false
    showToast(t('orgs.auditDeleted', { count: res.deleted }))
    await refreshAudit()
  } catch (e) {
    showToast(e.response?.data?.error || e.message, 'error')
  } finally {
    cleaningAudit.value = false
  }
}

// ── Helpers ───────────────────────────────────────────────────────────────────

const showToast = (message, type = 'success') => {
  toast.value = { message, type }
  setTimeout(() => { toast.value = null }, 3000)
}

const formatSize = (bytes) => {
  if (!bytes) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Number.parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleDateString()
}

const capitalize = (s) => s ? s.charAt(0).toUpperCase() + s.slice(1) : ''
</script>

<style scoped>
.org-detail {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

/* ── Header ─────────────────────────────────────────────────────────────── */
.org-header {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 14px 24px 16px;
  border-bottom: 1px solid var(--border-color);
  background: var(--card-color);
  flex-shrink: 0;
}

.btn-back {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  background: none;
  border: none;
  color: var(--secondary-text-color);
  font-size: 0.78rem;
  font-weight: 500;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 5px;
  transition: background 0.15s, color 0.15s;
  align-self: flex-start;
}
.btn-back:hover { background: var(--hover-background-color); color: var(--main-text-color); }

.org-identity {
  display: flex;
  align-items: center;
  gap: 14px;
}

.org-avatar-wrap {
  position: relative;
  width: 44px;
  height: 44px;
  flex-shrink: 0;
}
.org-avatar-wrap.is-admin { cursor: pointer; }
.org-avatar-wrap.is-admin:hover .org-avatar-overlay { opacity: 1; }
.org-avatar-wrap.is-admin:hover .logo-remove-btn { opacity: 1; }

.org-avatar {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  background: linear-gradient(135deg, var(--primary-color), var(--secondary-color));
  color: white;
  font-size: 1.3rem;
  font-weight: 800;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 2px 8px color-mix(in srgb, var(--primary-color) 40%, transparent);
}

.org-avatar-img {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  object-fit: cover;
  background: var(--card-color);
  box-shadow: 0 2px 8px color-mix(in srgb, var(--primary-color) 20%, transparent);
}

.org-avatar-overlay {
  position: absolute;
  inset: 0;
  border-radius: 12px;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  opacity: 0;
  transition: opacity 0.15s;
  pointer-events: none;
}

.logo-remove-btn {
  position: absolute;
  top: -5px;
  right: -5px;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: var(--error-color);
  color: white;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.15s;
  z-index: 1;
  padding: 0;
}

.org-identity-body { flex: 1; min-width: 0; }

.org-identity-top {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 2px;
}

.org-name {
  font-size: 1.05rem;
  font-weight: 700;
  color: var(--main-text-color);
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.org-desc {
  font-size: 0.78rem;
  color: var(--secondary-text-color);
  margin: 0 0 8px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* Storage bar */
.storage-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
}

.storage-track {
  flex: 1;
  max-width: 180px;
  height: 4px;
  background: var(--border-color);
  border-radius: 2px;
  overflow: hidden;
}

.storage-fill {
  height: 100%;
  border-radius: 2px;
  transition: width 0.3s ease;
}
.storage-fill.ok       { background: var(--success-color); }
.storage-fill.warning  { background: var(--warning-color); }
.storage-fill.critical { background: var(--error-color); }

.storage-text {
  font-size: 0.72rem;
  color: var(--secondary-text-color);
  white-space: nowrap;
}

/* legacy alias kept for any refs that remain */
.storage-pill {
  display: flex;
  align-items: center;
  gap: 6px;
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 20px;
  padding: 5px 12px;
  font-size: 0.78rem;
  color: var(--secondary-text-color);
  flex-shrink: 0;
}

/* ── Tabs ───────────────────────────────────────────────────────────────── */
.tabs-bar {
  display: flex;
  gap: 0;
  border-bottom: 1px solid var(--border-color);
  padding: 0 24px;
  overflow-x: auto;
  flex-shrink: 0;
  scrollbar-width: none;
}
.tabs-bar::-webkit-scrollbar { display: none; }

.tab-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 12px 14px;
  border: none;
  background: none;
  cursor: pointer;
  color: var(--secondary-text-color);
  font-size: 0.84rem;
  font-weight: 500;
  transition: color 0.15s;
  white-space: nowrap;
  position: relative;
}
.tab-btn:hover { color: var(--main-text-color); }
.tab-btn.active { color: var(--primary-color); font-weight: 600; }
.tab-btn.active::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 14px;
  right: 14px;
  height: 2px;
  background: var(--primary-color);
  border-radius: 2px 2px 0 0;
}

.tab-icon { flex-shrink: 0; }

.tab-count {
  font-size: 0.68rem;
  font-weight: 700;
  padding: 1px 6px;
  border-radius: 10px;
  background: var(--hover-background-color);
  color: var(--secondary-text-color);
  min-width: 18px;
  text-align: center;
}
.tab-btn.active .tab-count {
  background: color-mix(in srgb, var(--primary-color) 15%, transparent);
  color: var(--primary-color);
}

/* Tab content */
.tab-content {
  flex: 1;
  overflow-y: auto;
  padding: 20px 24px;
}

/* File system */
.fs-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  gap: 12px;
  flex-wrap: wrap;
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 4px;
  flex: 1;
  min-width: 0;
  overflow: hidden;
}

.bc-item {
  background: none;
  border: none;
  color: var(--primary-color);
  cursor: pointer;
  font-size: 0.88rem;
  font-weight: 500;
  padding: 4px 6px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  transition: background 0.15s;
  white-space: nowrap;
}

.bc-item:hover { background: var(--hover-background-color); }
.bc-sep { color: var(--secondary-text-color); font-size: 0.85rem; }

.fs-actions { display: flex; gap: 8px; }

.btn-sm, .btn-sm-action {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 7px 14px;
  border: 1px solid var(--border-color);
  border-radius: 7px;
  background: var(--card-color);
  color: var(--main-text-color);
  font-size: 0.82rem;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.15s;
  white-space: nowrap;
}

.btn-sm:hover, .btn-sm-action:hover { background: var(--hover-background-color); }

.btn-sm-action {
  background: var(--primary-color);
  color: white;
  border-color: transparent;
}

.btn-sm-action:hover { opacity: 0.9; }

.btn-upload { position: relative; }

.items-list {
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.item-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 9px 10px;
  border-radius: 7px;
  cursor: default;
  transition: background 0.1s;
  position: relative;
}
.item-row:hover { background: var(--hover-background-color); }

.folder-row { cursor: pointer; }
.folder-row:hover .item-name { color: var(--primary-color); }

.item-icon { flex-shrink: 0; }
.folder-icon { color: var(--warning-color); }
.file-icon { color: var(--secondary-text-color); opacity: 0.7; }

.item-name {
  flex: 1;
  font-size: 0.88rem;
  font-weight: 500;
  color: var(--main-text-color);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  transition: color 0.12s;
}

.item-meta {
  font-size: 0.73rem;
  color: var(--secondary-text-color);
  white-space: nowrap;
  flex-shrink: 0;
}

.item-actions {
  display: flex;
  gap: 2px;
  opacity: 0;
  transition: opacity 0.15s;
  flex-shrink: 0;
}
.item-row:hover .item-actions { opacity: 1; }

.btn-icon {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--secondary-text-color);
  padding: 5px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  transition: background 0.15s, color 0.15s;
}

.btn-icon:hover { background: var(--hover-background-color); color: var(--primary-color); }

.btn-icon-danger {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--secondary-text-color);
  padding: 5px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  transition: background 0.15s, color 0.15s;
}

.btn-icon-danger:hover { background: rgba(239,68,68,0.1); color: #ef4444; }

.empty-folder {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 64px 24px;
  color: var(--secondary-text-color);
  text-align: center;
}
.empty-folder svg { opacity: 0.18; }
.empty-folder p { margin: 0; font-size: 0.88rem; line-height: 1.5; }

/* Members */
.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.section-header h3 {
  font-size: 1rem;
  font-weight: 700;
  color: var(--main-text-color);
  margin: 0;
}

/* Profile tab */
.profile-section {
  margin-bottom: 24px;
}

.profile-label {
  font-size: 0.78rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--secondary-text-color);
  margin-bottom: 10px;
}

.profile-role-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 10px;
}

.profile-groups-list { display: flex; flex-direction: column; gap: 8px; }

.profile-group-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 10px;
}

.group-avatar-sm {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background: var(--primary-color);
  color: white;
  font-weight: 700;
  font-size: 0.9rem;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.profile-empty {
  color: var(--secondary-text-color);
  font-size: 0.88rem;
  padding: 16px 0;
}

.members-list { display: flex; flex-direction: column; gap: 8px; }

.member-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 10px;
}

.member-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: var(--primary-color);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  flex-shrink: 0;
  font-size: 0.9rem;
}

.member-info { flex: 1; min-width: 0; }

.member-name {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--main-text-color);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.member-email {
  font-size: 0.78rem;
  color: var(--secondary-text-color);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.member-meta { display: flex; align-items: center; gap: 8px; }

.role-select {
  background: var(--background-color);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  padding: 4px 8px;
  font-size: 0.8rem;
  color: var(--main-text-color);
  cursor: pointer;
}

/* Invitations */
.invitations-list { display: flex; flex-direction: column; gap: 8px; }

.invite-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 10px;
}

.invite-info { flex: 1; min-width: 0; }

.invite-token {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}

.invite-token code {
  font-size: 0.75rem;
  color: var(--secondary-text-color);
  background: var(--background-color);
  padding: 3px 8px;
  border-radius: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 340px;
}

.btn-copy {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--secondary-text-color);
  padding: 3px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  transition: color 0.15s;
}

.btn-copy:hover { color: var(--primary-color); }

.invite-meta { display: flex; align-items: center; gap: 8px; }

.invite-detail { font-size: 0.75rem; color: var(--secondary-text-color); }

/* Permissions */
.permissions-list { display: flex; flex-direction: column; gap: 8px; }

.perm-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 10px;
}

.perm-info { flex: 1; min-width: 0; }

.perm-path {
  display: block;
  font-size: 0.85rem;
  color: var(--main-text-color);
  margin-bottom: 2px;
}

.perm-user {
  font-size: 0.75rem;
  color: var(--secondary-text-color);
}

.perm-level { flex-shrink: 0; }

.level-badge {
  font-size: 0.72rem;
  font-weight: 600;
  padding: 3px 8px;
  border-radius: 20px;
}

.level-badge.none   { background: color-mix(in srgb, var(--secondary-text-color) 10%, transparent); color: var(--secondary-text-color); }
.level-badge.read   { background: color-mix(in srgb, var(--success-color) 12%, transparent);        color: var(--success-color); }
.level-badge.write  { background: color-mix(in srgb, var(--primary-color) 12%, transparent);        color: var(--primary-color); }
.level-badge.manage { background: color-mix(in srgb, var(--secondary-color) 12%, transparent);      color: var(--secondary-color); }

/* Settings */
.settings-section {
  max-width: 480px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.settings-section h3 {
  font-size: 1rem;
  font-weight: 700;
  color: var(--main-text-color);
  margin: 0 0 16px 0;
}

.danger-zone {
  margin-top: 40px;
  padding: 20px;
  border: 1px solid rgba(239,68,68,0.3);
  border-radius: 10px;
  max-width: 480px;
}

.danger-zone h4 {
  font-size: 0.9rem;
  font-weight: 700;
  color: #ef4444;
  margin: 0 0 14px 0;
}

.btn-danger {
  background: rgba(239,68,68,0.1);
  color: #ef4444;
  border: 1px solid rgba(239,68,68,0.3);
  border-radius: 8px;
  padding: 9px 18px;
  font-size: 0.88rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s;
}

.btn-danger:hover { background: rgba(239,68,68,0.18); }

.btn-danger-outline {
  background: none;
  color: #ef4444;
  border: 1px solid rgba(239,68,68,0.5);
  border-radius: 8px;
  padding: 9px 18px;
  font-size: 0.88rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}
.btn-danger-outline:hover:not(:disabled) { background: rgba(239,68,68,0.08); }
.btn-danger-outline:disabled { opacity: 0.5; cursor: not-allowed; }

.danger-action {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 0;
  border-bottom: 1px solid rgba(239,68,68,0.15);
}
.danger-action:last-child { border-bottom: none; padding-bottom: 0; }
.danger-action > div { flex: 1; }
.danger-action strong { font-size: 0.9rem; color: var(--main-text-color); }
.hint-sm { font-size: 0.8rem; color: var(--secondary-text-color); margin: 4px 0 0 0; }

/* Audit log */
.audit-list { display: flex; flex-direction: column; gap: 1px; }

.audit-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  border-radius: 8px;
  transition: background 0.1s;
}
.audit-row:hover { background: var(--hover-background-color); }

.audit-action { display: flex; align-items: center; gap: 10px; flex: 1; min-width: 0; }

.audit-badge {
  font-size: 0.72rem;
  font-weight: 700;
  padding: 3px 8px;
  border-radius: 4px;
  white-space: nowrap;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  background: var(--hover-background-color);
  color: var(--secondary-text-color);
}
.audit-badge.file_uploaded, .audit-badge.file_deleted { background: color-mix(in srgb, var(--primary-color) 12%, transparent); color: var(--primary-color); }
.audit-badge.file_downloaded  { background: color-mix(in srgb, var(--success-color) 10%, transparent);       color: var(--success-color); }
.audit-badge.member_joined    { background: color-mix(in srgb, var(--success-color) 12%, transparent);       color: var(--success-color); }
.audit-badge.member_removed   { background: color-mix(in srgb, var(--error-color) 12%, transparent);         color: var(--error-color); }
.audit-badge.role_changed     { background: color-mix(in srgb, var(--warning-color) 12%, transparent);       color: var(--warning-color); }
.audit-badge.key_rotated, .audit-badge.key_provisioned     { background: color-mix(in srgb, var(--secondary-color) 12%, transparent); color: var(--secondary-color); }
.audit-badge.permission_set, .audit-badge.permission_removed { background: color-mix(in srgb, var(--warning-color) 10%, transparent); color: var(--warning-color); }

.audit-detail { font-size: 0.83rem; color: var(--secondary-text-color); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.audit-meta { display: flex; align-items: center; gap: 12px; flex-shrink: 0; }
.audit-actor { font-size: 0.78rem; font-family: monospace; color: var(--secondary-text-color); }
.audit-time { font-size: 0.78rem; color: var(--secondary-text-color); }

/* Audit log — load more / retention */
.audit-load-more {
  display: flex;
  justify-content: center;
  padding: 16px 0 4px;
}

.audit-retention-note {
  font-size: 0.72rem;
  color: var(--secondary-text-color);
  white-space: nowrap;
}

.btn-clean {
  color: #ef4444;
  border-color: rgba(239,68,68,0.3);
}
.btn-clean:hover { background: rgba(239,68,68,0.07); }

/* Clean modal */
.modal-clean { max-width: 520px; }

.clean-tabs {
  display: flex;
  gap: 4px;
  margin-bottom: 16px;
  background: var(--background-color);
  border-radius: 8px;
  padding: 4px;
}

.clean-tab {
  flex: 1;
  padding: 7px 10px;
  border: none;
  background: none;
  border-radius: 6px;
  font-size: 0.82rem;
  font-weight: 500;
  color: var(--secondary-text-color);
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}
.clean-tab.active {
  background: var(--card-color);
  color: var(--main-text-color);
  box-shadow: 0 1px 4px rgba(0,0,0,0.08);
}

.clean-panel { min-height: 180px; }

.clean-warn {
  font-size: 0.88rem;
  color: #ef4444;
  background: rgba(239,68,68,0.07);
  border: 1px solid rgba(239,68,68,0.2);
  border-radius: 8px;
  padding: 12px 16px;
  margin: 0;
}

.clean-empty {
  padding: 40px 0;
  text-align: center;
  color: var(--secondary-text-color);
  font-size: 0.85rem;
}

.clean-selection-hint {
  font-size: 0.78rem;
  color: var(--secondary-text-color);
  margin: 10px 0 0;
  text-align: center;
}

/* Month grid */
.month-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.month-chip {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 14px;
  border-radius: 20px;
  border: 1px solid var(--border-color);
  background: var(--background-color);
  cursor: pointer;
  font-size: 0.82rem;
  font-weight: 500;
  color: var(--main-text-color);
  transition: background 0.15s, border-color 0.15s;
}
.month-chip:hover { background: var(--hover-background-color); }
.month-chip.selected {
  background: rgba(239,68,68,0.08);
  border-color: rgba(239,68,68,0.4);
  color: #ef4444;
}

.month-label { white-space: nowrap; }

.month-count {
  font-size: 0.7rem;
  font-weight: 700;
  padding: 1px 6px;
  border-radius: 10px;
  background: var(--border-color);
  color: var(--secondary-text-color);
}
.month-chip.selected .month-count {
  background: rgba(239,68,68,0.2);
  color: #ef4444;
}

/* Calendar */
.cal-nav {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.cal-month-label {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--main-text-color);
  text-transform: capitalize;
}

.cal-nav-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--secondary-text-color);
  padding: 4px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  transition: background 0.15s, color 0.15s;
}
.cal-nav-btn:hover:not(:disabled) { background: var(--hover-background-color); color: var(--main-text-color); }
.cal-nav-btn:disabled { opacity: 0.3; cursor: default; }

.cal-grid {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 2px;
}

.cal-dow {
  font-size: 0.7rem;
  font-weight: 600;
  color: var(--secondary-text-color);
  text-align: center;
  padding: 4px 0;
  text-transform: uppercase;
}

.cal-cell {
  aspect-ratio: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  cursor: default;
  position: relative;
  gap: 1px;
  transition: background 0.1s;
}

.cal-cell.cal-empty { visibility: hidden; }

.cal-cell.cal-none {
  color: var(--secondary-text-color);
  opacity: 0.4;
}

.cal-cell.cal-has {
  cursor: pointer;
  background: rgba(99,102,241,0.07);
  color: var(--primary-color);
}
.cal-cell.cal-has:hover { background: rgba(99,102,241,0.15); }

.cal-cell.cal-selected {
  background: rgba(239,68,68,0.12) !important;
  color: #ef4444 !important;
}
.cal-cell.cal-selected:hover { background: rgba(239,68,68,0.2) !important; }

.cal-day-num {
  font-size: 0.78rem;
  font-weight: 600;
  line-height: 1;
}

.cal-day-count {
  font-size: 0.58rem;
  font-weight: 700;
  opacity: 0.75;
}

/* Key provisioning */
.key-missing-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 0.7rem;
  font-weight: 600;
  padding: 3px 8px;
  border-radius: 20px;
  background: rgba(239,68,68,0.1);
  color: #ef4444;
}

.btn-provision {
  font-size: 0.75rem;
  padding: 5px 12px;
  background: color-mix(in srgb, var(--primary-color) 12%, transparent);
  color: var(--primary-color);
  border: 1px solid color-mix(in srgb, var(--primary-color) 30%, transparent);
  border-radius: 6px;
  cursor: pointer;
  font-weight: 600;
  transition: background 0.15s;
  white-space: nowrap;
}
.btn-provision:hover:not(:disabled) { background: color-mix(in srgb, var(--primary-color) 20%, transparent); }
.btn-provision:disabled { opacity: 0.5; cursor: not-allowed; }

/* ── Shared badges ───────────────────────────────────────────────────────── */
.role-badge {
  font-size: 0.69rem;
  font-weight: 700;
  padding: 3px 9px;
  border-radius: 20px;
  letter-spacing: 0.02em;
  white-space: nowrap;
  flex-shrink: 0;
}

/* Org roles — Kagibi palette */
.role-badge.owner   { background: color-mix(in srgb, var(--primary-color) 18%, transparent);   color: var(--primary-color); }
.role-badge.admin   { background: color-mix(in srgb, var(--secondary-color) 15%, transparent); color: var(--secondary-color); }
.role-badge.member  { background: color-mix(in srgb, var(--success-color) 14%, transparent);   color: var(--success-color); }
.role-badge.viewer  { background: color-mix(in srgb, var(--secondary-text-color) 10%, transparent); color: var(--secondary-text-color); }

/* Group roles */
.role-badge.group-admin  { background: color-mix(in srgb, var(--accent-color) 14%, transparent); color: var(--accent-color); }
.role-badge.group-member { background: color-mix(in srgb, var(--secondary-text-color) 8%, transparent); color: var(--secondary-text-color); }

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
}

.modal-sm { max-width: 360px; }

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px 16px;
  border-bottom: 1px solid var(--border-color);
}

.modal-header h3 {
  margin: 0;
  font-size: 1rem;
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
  margin-bottom: 14px;
}

.form-group label {
  font-size: 0.82rem;
  font-weight: 500;
  color: var(--secondary-text-color);
}

.input-field {
  background: var(--background-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 10px 12px;
  font-size: 0.88rem;
  color: var(--main-text-color);
  width: 100%;
  box-sizing: border-box;
  transition: border-color 0.15s;
}

.input-field:focus { outline: none; border-color: var(--primary-color); }

.form-error { color: #ef4444; font-size: 0.8rem; margin: 0; }

.btn-primary {
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--primary-color);
  color: white;
  border: none;
  border-radius: 8px;
  padding: 10px 18px;
  font-size: 0.88rem;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.15s;
}

.btn-primary:hover { opacity: 0.9; }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }

.btn-secondary {
  background: var(--card-color);
  color: var(--main-text-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 10px 18px;
  font-size: 0.88rem;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.15s;
}

.btn-secondary:hover { background: var(--hover-background-color); }

.empty-tab {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 60px 24px;
  color: var(--secondary-text-color);
  font-size: 0.88rem;
}

/* Spinners */
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

.spinner-sm-dark {
  display: inline-block;
  width: 22px;
  height: 22px;
  border: 2px solid var(--border-color);
  border-top-color: var(--primary-color);
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

.loading-center {
  display: flex;
  justify-content: center;
  align-items: center;
  flex: 1;
  padding: 80px;
}

.loading-inline {
  display: flex;
  justify-content: center;
  padding: 40px;
}

@keyframes spin { to { transform: rotate(360deg); } }

/* Transitions */
.modal-enter-active, .modal-leave-active { transition: opacity 0.2s; }
.modal-enter-from, .modal-leave-to { opacity: 0; }

/* ── Key init banner ────────────────────────────────────────────────────── */
.key-init-banner {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  margin-bottom: 16px;
  background: color-mix(in srgb, var(--warning-color) 8%, var(--card-color));
  border: 1px solid color-mix(in srgb, var(--warning-color) 30%, transparent);
  border-radius: 10px;
}

.key-init-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: color-mix(in srgb, var(--warning-color) 15%, transparent);
  color: var(--warning-color);
  flex-shrink: 0;
}

.key-init-text {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.key-init-text strong { font-size: 0.84rem; color: var(--main-text-color); font-weight: 600; }
.key-init-text span   { font-size: 0.75rem; color: var(--secondary-text-color); }

.btn-init-key {
  background: var(--warning-color);
  color: white;
  border: none;
  border-radius: 7px;
  padding: 7px 14px;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  white-space: nowrap;
  transition: opacity 0.15s;
  flex-shrink: 0;
}
.btn-init-key:hover:not(:disabled) { opacity: 0.87; }
.btn-init-key:disabled { opacity: 0.6; cursor: not-allowed; }

/* Upload progress queue */
.upload-queue {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 12px;
  padding: 10px 14px;
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
}

.upload-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.upload-name {
  flex: 1;
  font-size: 0.82rem;
  color: var(--main-text-color);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
}

.upload-bar-track {
  width: 120px;
  flex-shrink: 0;
  height: 4px;
  background: var(--border-color);
  border-radius: 2px;
  overflow: hidden;
}

.upload-bar-fill {
  height: 100%;
  background: var(--primary-color);
  border-radius: 2px;
  transition: width 0.2s ease;
}

.upload-pct {
  font-size: 0.75rem;
  color: var(--secondary-text-color);
  width: 30px;
  text-align: right;
  flex-shrink: 0;
}

/* Toast */
.toast {
  position: fixed;
  bottom: 80px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 3000;
  background: #1e293b;
  color: white;
  padding: 10px 20px;
  border-radius: 8px;
  font-size: 0.88rem;
  font-weight: 500;
  box-shadow: 0 4px 16px rgba(0,0,0,0.2);
  white-space: nowrap;
}

.toast.error { background: #dc2626; }
.toast.success { background: #16a34a; }

.toast-enter-active, .toast-leave-active { transition: opacity 0.25s, transform 0.25s; }
.toast-enter-from, .toast-leave-to { opacity: 0; transform: translateX(-50%) translateY(12px); }

/* Upload progress queue */
.upload-queue {
  margin: 8px 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.upload-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 10px;
  background: var(--hover-background-color);
  border-radius: 8px;
  font-size: 0.82rem;
}

.upload-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--main-text-color);
}

.upload-bar-track {
  width: 120px;
  height: 5px;
  background: var(--border-color);
  border-radius: 3px;
  overflow: hidden;
  flex-shrink: 0;
}

.upload-bar-fill {
  height: 100%;
  background: var(--primary-color);
  border-radius: 3px;
  transition: width 0.3s ease;
}

.upload-pct {
  width: 36px;
  text-align: right;
  color: var(--secondary-text-color);
  font-size: 0.78rem;
  flex-shrink: 0;
}

@media (max-width: 768px) {
  .org-header  { padding: 10px 16px 12px; }
  .tabs-bar    { padding: 0 16px; }
  .tab-content { padding: 14px 16px; }
  .org-name    { font-size: 0.96rem; }
  .storage-indicator { display: none; }
  .tab-btn     { padding: 10px 10px; font-size: 0.8rem; gap: 4px; }
  .tab-count   { display: none; }
}
</style>
