<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="close">
    <div class="modal-wrapper" :class="{ 'with-side-panel': restrictionsPanelOpen && isShared && item?.type === 'folder' }">

      <!-- ── Left panel (main dialog) ── -->
      <div class="modal-content">
        <div class="modal-header">
          <h3>{{ t('share.title', { name: item?.Name || item?.name }) }}</h3>
          <button @click="close" class="btn-close">×</button>
        </div>

        <div class="modal-body">

          <!-- === FRIENDS SECTION === -->
          <div class="friends-section">
              <h4 class="section-title">{{ t('share.withFriends') }}</h4>

              <div v-if="friends.length === 0" class="empty-friends">
                  {{ t('friends.noFriends') }}
                  <br>
                  <router-link to="/friends">{{ t('friends.addFriend') }}</router-link>
              </div>

              <div v-else class="friends-list">
                   <div v-for="friend in friends" :key="friend.id" class="friend-entry">
                      <div class="friend-item">
                         <div class="friend-info">
                            <div class="friend-avatar">
                              {{ friend.name.charAt(0).toUpperCase() }}
                            </div>
                            <div>
                              <p class="friend-name">{{ friend.name }}</p>
                              <p class="friend-email">{{ friend.email }}</p>
                            </div>
                         </div>

                         <div v-if="!friend.public_key" class="key-missing" :title="t('share.keyMissing')">
                            <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor" style="vertical-align:middle;margin-right:4px;flex-shrink:0"><path d="M1 21h22L12 2 1 21zm12-3h-2v-2h2v2zm0-4h-2v-4h2v4z"/></svg>{{ t('share.noKey') }}
                         </div>

                         <button v-else
                             @click="shareWithFriend(friend)"
                             :disabled="sharing[friend.id]"
                             class="btn-sm"
                             :class="[isFriendShared(friend.id) ? 'btn-danger' : 'btn-outline']">
                             <span v-if="sharing[friend.id]">...</span>
                             <span v-else-if="isFriendShared(friend.id)">{{ t('share.stop') }}</span>
                             <span v-else>{{ t('share.send') }}</span>
                         </button>
                      </div>

                      <!-- Per-friend permissions (only when shared) -->
                      <div v-if="isFriendShared(friend.id)" class="friend-perms">
                         <button
                             class="perm-friend-chip"
                             :class="{ active: directShareStatus[friend.id]?.perm_download }"
                             @click="toggleDirectPerm(friend, 'download')"
                             title="Téléchargement">
                             <svg viewBox="0 0 24 24" width="11" height="11" fill="currentColor"><path d="M19 9h-4V3H9v6H5l7 7 7-7zm-8 2V5h2v6h1.17L12 13.17 9.83 11H11zm-6 7h14v2H5v-2z"/></svg>
                             Téléchargement
                         </button>
                         <template v-if="item?.type === 'folder'">
                            <button
                                class="perm-friend-chip"
                                :class="{ active: directShareStatus[friend.id]?.perm_create }"
                                @click="toggleDirectPerm(friend, 'create')"
                                title="Création">
                                <svg viewBox="0 0 24 24" width="11" height="11" fill="currentColor"><path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z"/></svg>
                                Création
                            </button>
                            <button
                                class="perm-friend-chip"
                                :class="{ active: directShareStatus[friend.id]?.perm_delete }"
                                @click="toggleDirectPerm(friend, 'delete')"
                                title="Suppression">
                                <svg viewBox="0 0 24 24" width="11" height="11" fill="currentColor"><path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM15.5 4l-1-1h-5l-1 1H5v2h14V4z"/></svg>
                                Suppression
                            </button>
                            <button
                                class="perm-friend-chip"
                                :class="{ active: directShareStatus[friend.id]?.perm_move }"
                                @click="toggleDirectPerm(friend, 'move')"
                                title="Déplacement">
                                <svg viewBox="0 0 24 24" width="11" height="11" fill="currentColor"><path d="M10 9h4V6h3l-5-5-5 5h3v3zm-1 1H6V7l-5 5 5 5v-3h3v-4zm14 2-5-5v3h-3v4h3v3l5-5zm-9 3h-4v3H7l5 5 5-5h-3v-3z"/></svg>
                                Déplacement
                            </button>
                         </template>
                      </div>
                   </div>
              </div>
          </div>

          <div class="section-divider"></div>

          <!-- === LINK SECTION === -->
          <div class="link-section-wrapper">
              <h4 class="section-title">Partage via lien public</h4>

              <div v-if="loading" class="loading-state">
                  <div class="spinner"></div> Traitement en cours...
              </div>

              <!-- Not Shared State -->
              <div v-else-if="!isShared" class="not-shared-state">
                  <div class="illustration">
                    <svg viewBox="0 0 24 24" width="40" height="40" fill="currentColor" style="opacity:0.35"><path d="M3.9 12c0-1.71 1.39-3.1 3.1-3.1h4V7H7c-2.76 0-5 2.24-5 5s2.24 5 5 5h4v-1.9H7c-1.71 0-3.1-1.39-3.1-3.1zM8 13h8v-2H8v2zm9-6h-4v1.9h4c1.71 0 3.1 1.39 3.1 3.1s-1.39 3.1-3.1 3.1h-4V17h4c2.76 0 5-2.24 5-5s-2.24-5-5-5z"/></svg>
                  </div>
                  <p>Ce {{ item?.type === 'folder' ? 'dossier' : 'fichier' }} n'est pas encore partagé par lien.</p>
                  <p class="sub-text">Créez un lien pour le partager avec d'autres personnes.</p>

                  <div class="form-group">
                      <label for="expiresAt">Expiration (optionnel)</label>
                      <input type="datetime-local" id="expiresAt" v-model="expiresAt" class="form-control" />
                  </div>

                  <div class="perm-group">
                      <label class="perm-label">Droits accordés</label>
                      <div class="perm-chips">
                          <label class="perm-chip" :class="{ active: permissions.download }">
                              <input type="checkbox" v-model="permissions.download" />
                              <svg viewBox="0 0 24 24" width="13" height="13" fill="currentColor"><path d="M19 9h-4V3H9v6H5l7 7 7-7zm-8 2V5h2v6h1.17L12 13.17 9.83 11H11zm-6 7h14v2H5v-2z"/></svg>
                              Téléchargement
                          </label>
                          <template v-if="item?.type === 'folder'">
                          <label class="perm-chip" :class="{ active: permissions.create }">
                              <input type="checkbox" v-model="permissions.create" />
                              <svg viewBox="0 0 24 24" width="13" height="13" fill="currentColor"><path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z"/></svg>
                              Création
                          </label>
                          <label class="perm-chip" :class="{ active: permissions.delete }">
                              <input type="checkbox" v-model="permissions.delete" />
                              <svg viewBox="0 0 24 24" width="13" height="13" fill="currentColor"><path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zm3.46-7.12 1.41-1.41L12 11.59l1.12-1.12 1.41 1.41L13.41 13l1.12 1.12-1.41 1.41L12 14.41l-1.12 1.12-1.41-1.41L10.59 13l-1.13-1.12zM15.5 4l-1-1h-5l-1 1H5v2h14V4z"/></svg>
                              Suppression
                          </label>
                          <label class="perm-chip" :class="{ active: permissions.move }">
                              <input type="checkbox" v-model="permissions.move" />
                              <svg viewBox="0 0 24 24" width="13" height="13" fill="currentColor"><path d="M10 9h4V6h3l-5-5-5 5h3v3zm-1 1H6V7l-5 5 5 5v-3h3v-4zm14 2-5-5v3h-3v4h3v3l5-5zm-9 3h-4v3H7l5 5 5-5h-3v-3z"/></svg>
                              Déplacement
                          </label>
                          </template>
                      </div>
                  </div>

                  <button @click="createShare" class="btn-primary">Créer un lien de partage</button>
              </div>

              <!-- Shared State (link share or direct share) -->
              <div v-else-if="isShared || isDirectShare" class="shared-state">
                  <div v-if="isShared && !isDirectShare" class="link-section">
                      <label>Lien de partage</label>
                      <div class="link-container">
                          <input type="text" :value="shareUrl" readonly ref="shareLinkInput" @click="selectAll" />
                          <button @click="copyLink" class="btn-copy" :class="{ copied: linkCopied }">
                              {{ linkCopied ? 'Copié !' : 'Copier' }}
                          </button>
                      </div>
                  </div>
                  <div v-else-if="isDirectShare" class="link-section" style="text-align:center;padding:0.5rem 0;color:var(--secondary-text-color);font-size:0.9rem;">
                      Partage direct avec un ami
                  </div>

                  <!-- Visible rights card -->
                  <div class="rights-card">
                      <div class="rights-card-title">
                          <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4zm0 4l5 2.18V11c0 3.5-2.33 6.79-5 7.93-2.67-1.14-5-4.43-5-7.93V7.18L12 5z"/></svg>
                          Droits accordés
                      </div>
                      <div class="rights-rows">
                          <div class="rights-row clickable" :class="{ active: localPermissions.download }" @click="togglePermission('download')">
                              <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M19 9h-4V3H9v6H5l7 7 7-7zm-8 2V5h2v6h1.17L12 13.17 9.83 11H11zm-6 7h14v2H5v-2z"/></svg>
                              <span>Téléchargement</span>
                              <span class="rights-toggle">
                                <svg v-if="localPermissions.download" viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/></svg>
                                <svg v-else viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
                              </span>
                          </div>
                          <template v-if="item?.type === 'folder'">
                          <div class="rights-row clickable" :class="{ active: localPermissions.create }" @click="togglePermission('create')">
                              <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z"/></svg>
                              <span>Création</span>
                              <span class="rights-toggle">
                                <svg v-if="localPermissions.create" viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/></svg>
                                <svg v-else viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
                              </span>
                          </div>
                          <div class="rights-row clickable" :class="{ active: localPermissions.delete }" @click="togglePermission('delete')">
                              <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM15.5 4l-1-1h-5l-1 1H5v2h14V4z"/></svg>
                              <span>Suppression</span>
                              <span class="rights-toggle">
                                <svg v-if="localPermissions.delete" viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/></svg>
                                <svg v-else viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
                              </span>
                          </div>
                          <div class="rights-row clickable" :class="{ active: localPermissions.move }" @click="togglePermission('move')">
                              <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M10 9h4V6h3l-5-5-5 5h3v3zm-1 1H6V7l-5 5 5 5v-3h3v-4zm14 2-5-5v3h-3v4h3v3l5-5zm-9 3h-4v3H7l5 5 5-5h-3v-3z"/></svg>
                              <span>Déplacement</span>
                              <span class="rights-toggle">
                                <svg v-if="localPermissions.move" viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/></svg>
                                <svg v-else viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
                              </span>
                          </div>
                          </template>
                      </div>
                      <button v-if="item?.type === 'folder'" class="btn-manage-restrictions" @click="toggleRestrictionsPanel">
                          <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M19.14 12.94c.04-.3.06-.61.06-.94 0-.32-.02-.64-.07-.94l2.03-1.58c.18-.14.23-.41.12-.61l-1.92-3.32c-.12-.22-.37-.29-.59-.22l-2.39.96c-.5-.38-1.03-.7-1.62-.94l-.36-2.54c-.04-.24-.24-.41-.48-.41h-3.84c-.24 0-.43.17-.47.41l-.36 2.54c-.59.24-1.13.57-1.62.94l-2.39-.96c-.22-.08-.47 0-.59.22L2.74 8.87c-.12.21-.08.47.12.61l2.03 1.58c-.05.3-.09.63-.09.94s.02.64.07.94l-2.03 1.58c-.18.14-.23.41-.12.61l1.92 3.32c.12.22.37.29.59.22l2.39-.96c.5.38 1.03.7 1.62.94l.36 2.54c.05.24.24.41.48.41h3.84c.24 0 .44-.17.47-.41l.36-2.54c.59-.24 1.13-.56 1.62-.94l2.39.96c.22.08.47 0 .59-.22l1.92-3.32c.12-.22.07-.47-.12-.61l-2.01-1.58zM12 15.6c-1.98 0-3.6-1.62-3.6-3.6s1.62-3.6 3.6-3.6 3.6 1.62 3.6 3.6-1.62 3.6-3.6 3.6z"/></svg>
                          Gérer les restrictions par élément
                          <span class="arrow">{{ restrictionsPanelOpen ? '←' : '→' }}</span>
                      </button>
                  </div>

                  <div v-if="!isDirectShare" class="share-info">
                      <p v-if="localExpiresAt"><svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor" style="vertical-align:middle;margin-right:4px;flex-shrink:0"><path d="M11.99 2C6.47 2 2 6.48 2 12s4.47 10 9.99 10C17.52 22 22 17.52 22 12S17.52 2 11.99 2zM12 20c-4.42 0-8-3.58-8-8s3.58-8 8-8 8 3.58 8 8-3.58 8-8 8zm.5-13H11v6l5.25 3.15.75-1.23-4.5-2.67V7z"/></svg>Ce lien expirera le : <b>{{ formattedExpiration }}</b></p>
                      <p><svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor" style="vertical-align:middle;margin-right:4px;flex-shrink:0"><path d="M1 21h22L12 2 1 21zm12-3h-2v-2h2v2zm0-4h-2v-4h2v4z"/></svg>Toute personne disposant de ce lien pourra accéder au contenu <b>déchiffré</b>.</p>
                  </div>
              </div>
          </div>

        </div>

        <div class="modal-footer">
          <button v-if="isShared && !isDirectShare" @click="deleteShare" class="btn-delete">Arrêter le lien</button>
          <button @click="close" class="btn-secondary">Fermer</button>
        </div>
      </div>

      <!-- ── Right panel (restrictions tree) ── -->
      <div v-if="restrictionsPanelOpen && isShared && item?.type === 'folder'" class="modal-side-panel">
        <div class="side-panel-header">
          <span>Restrictions par élément</span>
          <button class="btn-close" @click="restrictionsPanelOpen = false">×</button>
        </div>

        <!-- Bulk controls -->
        <div class="bulk-controls">
          <div class="bulk-row">
            <span class="bulk-label">Dossiers — Accès</span>
            <div class="bulk-btns">
              <button v-for="level in ['full','readonly','none']" :key="level"
                  class="access-opt"
                  @click="applyBulkOverride('folder', { access_level: level })"
                  :title="{ full: 'Accès complet', readonly: 'Lecture seule', none: 'Masqué' }[level]">
                <svg v-if="level === 'full'" viewBox="0 0 24 24" width="11" height="11" fill="currentColor"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41L9 16.17z"/></svg>
                <svg v-else-if="level === 'readonly'" viewBox="0 0 24 24" width="11" height="11" fill="currentColor"><path d="M12 4.5C7 4.5 2.73 7.61 1 12c1.73 4.39 6 7.5 11 7.5s9.27-3.11 11-7.5c-1.73-4.39-6-7.5-11-7.5zM12 17c-2.76 0-5-2.24-5-5s2.24-5 5-5 5 2.24 5 5-2.24 5-5 5zm0-8c-1.66 0-3 1.34-3 3s1.34 3 3 3 3-1.34 3-3-1.34-3-3-3z"/></svg>
                <svg v-else viewBox="0 0 24 24" width="11" height="11" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12 19 6.41z"/></svg>
              </button>
            </div>
          </div>
          <div class="bulk-row">
            <span class="bulk-label">Dossiers — Suppr.</span>
            <div class="bulk-btns">
              <button class="access-opt" @click="applyBulkOverride('folder', { access_level: 'full', can_delete: true })" title="Tous supprimables">
                <svg viewBox="0 0 24 24" width="11" height="11" fill="currentColor"><path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM15.5 4l-1-1h-5l-1 1H5v2h14V4z"/></svg>
              </button>
              <button class="access-opt" @click="applyBulkOverride('folder', { access_level: 'full', can_delete: false })" title="Tous protégés">
                <svg viewBox="0 0 24 24" width="11" height="11" fill="currentColor"><path d="M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zM12 17c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2zm3.1-9H8.9V6c0-1.71 1.39-3.1 3.1-3.1 1.71 0 3.1 1.39 3.1 3.1v2z"/></svg>
              </button>
            </div>
          </div>
          <div class="bulk-row">
            <span class="bulk-label">Fichiers — Suppression</span>
            <div class="bulk-btns">
              <button class="access-opt" @click="applyBulkOverride('file', { can_delete: true })" title="Tous supprimables">
                <svg viewBox="0 0 24 24" width="11" height="11" fill="currentColor"><path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM15.5 4l-1-1h-5l-1 1H5v2h14V4z"/></svg>
              </button>
              <button class="access-opt" @click="applyBulkOverride('file', { can_delete: false })" title="Tous protégés">
                <svg viewBox="0 0 24 24" width="11" height="11" fill="currentColor"><path d="M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zM12 17c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2zm3.1-9H8.9V6c0-1.71 1.39-3.1 3.1-3.1 1.71 0 3.1 1.39 3.1 3.1v2z"/></svg>
              </button>
            </div>
          </div>
          <div class="bulk-row">
            <span class="bulk-label">Fichiers — Téléchargement</span>
            <div class="bulk-btns">
              <button class="access-opt" @click="applyBulkOverride('file', { can_download: true })" title="Téléchargement autorisé pour tous">
                <svg viewBox="0 0 24 24" width="11" height="11" fill="currentColor"><path d="M19 9h-4V3H9v6H5l7 7 7-7zm-8 2V5h2v6h1.17L12 13.17 9.83 11H11zm-6 7h14v2H5v-2z"/></svg>
              </button>
              <button class="access-opt" @click="applyBulkOverride('file', { can_download: false })" title="Téléchargement bloqué pour tous">
                <svg viewBox="0 0 24 24" width="11" height="11" fill="currentColor"><path d="M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zM12 17c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2zm3.1-9H8.9V6c0-1.71 1.39-3.1 3.1-3.1 1.71 0 3.1 1.39 3.1 3.1v2z"/></svg>
              </button>
            </div>
          </div>
        </div>

        <div v-if="treeLoading" class="tree-loading">Chargement...</div>
        <template v-else>
          <div class="tree-nav">
            <button v-for="(seg, i) in treePath" :key="i"
                @click="navigateTree(i)"
                class="tree-crumb"
                :class="{ active: i === treePath.length - 1 }">
              {{ seg.name }}{{ i < treePath.length - 1 ? ' /' : '' }}
            </button>
          </div>
          <div class="tree-items">
            <div v-for="folder in treeItems.folders" :key="folder.ID" class="tree-item">
              <button class="tree-item-name" @click="drillDown(folder)">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="#5f6368"><path d="M10 4H4c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z"/></svg>
                {{ folder.Name }}
              </button>
              <div class="folder-controls">
                <div class="access-toggle">
                  <button v-for="level in ['full','readonly','none']" :key="level"
                      class="access-opt"
                      :class="{ active: (folder.access_level || 'full') === level }"
                      @click="setFolderAccess(folder, level)"
                      :title="{ full: 'Accès complet', readonly: 'Lecture seule', none: 'Masqué' }[level]">
                    <svg v-if="level === 'full'" viewBox="0 0 24 24" width="11" height="11" fill="currentColor"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41L9 16.17z"/></svg>
                    <svg v-else-if="level === 'readonly'" viewBox="0 0 24 24" width="11" height="11" fill="currentColor"><path d="M12 4.5C7 4.5 2.73 7.61 1 12c1.73 4.39 6 7.5 11 7.5s9.27-3.11 11-7.5c-1.73-4.39-6-7.5-11-7.5zM12 17c-2.76 0-5-2.24-5-5s2.24-5 5-5 5 2.24 5 5-2.24 5-5 5zm0-8c-1.66 0-3 1.34-3 3s1.34 3 3 3 3-1.34 3-3-1.34-3-3-3z"/></svg>
                    <svg v-else viewBox="0 0 24 24" width="11" height="11" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12 19 6.41z"/></svg>
                  </button>
                </div>
                <div class="folder-extra-toggle">
                  <button class="access-opt"
                      :class="{ active: folder.can_delete !== false }"
                      @click="setFolderDelete(folder, true)"
                      title="Dossier supprimable (cascade sur le contenu)">
                    <svg viewBox="0 0 24 24" width="10" height="10" fill="currentColor"><path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM15.5 4l-1-1h-5l-1 1H5v2h14V4z"/></svg>
                  </button>
                  <button class="access-opt"
                      :class="{ active: folder.can_delete === false }"
                      @click="setFolderDelete(folder, false)"
                      title="Dossier protégé contre la suppression (cascade)">
                    <svg viewBox="0 0 24 24" width="10" height="10" fill="currentColor"><path d="M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zM12 17c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2zm3.1-9H8.9V6c0-1.71 1.39-3.1 3.1-3.1 1.71 0 3.1 1.39 3.1 3.1v2z"/></svg>
                  </button>
                </div>
              </div>
            </div>
            <div v-for="file in treeItems.files" :key="file.ID" class="tree-item">
              <span class="tree-item-name static">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="#888"><path d="M14 2H6c-1.1 0-2 .9-2 2v16c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z"/></svg>
                {{ file.Name }}
              </span>
              <div class="file-controls">
                <div class="download-toggle" title="Téléchargement / Prévisualisation">
                  <button class="access-opt"
                      :class="{ active: file.can_download !== false }"
                      @click="setFileDownload(file, true)"
                      title="Téléchargement autorisé">
                    <svg viewBox="0 0 24 24" width="10" height="10" fill="currentColor"><path d="M19 9h-4V3H9v6H5l7 7 7-7zm-8 2V5h2v6h1.17L12 13.17 9.83 11H11zm-6 7h14v2H5v-2z"/></svg>
                  </button>
                  <button class="access-opt"
                      :class="{ active: file.can_download === false }"
                      @click="setFileDownload(file, false)"
                      title="Téléchargement bloqué (preview aussi)">
                    <svg viewBox="0 0 24 24" width="10" height="10" fill="currentColor"><path d="M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zM12 17c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2zm3.1-9H8.9V6c0-1.71 1.39-3.1 3.1-3.1 1.71 0 3.1 1.39 3.1 3.1v2z"/></svg>
                  </button>
                </div>
                <div class="delete-toggle">
                  <button class="access-opt"
                      :class="{ active: file.can_delete !== false }"
                      @click="setFileDelete(file, true)"
                      title="Supprimable">
                    <svg viewBox="0 0 24 24" width="10" height="10" fill="currentColor"><path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM15.5 4l-1-1h-5l-1 1H5v2h14V4z"/></svg>
                  </button>
                  <button class="access-opt"
                      :class="{ active: file.can_delete === false }"
                      @click="setFileDelete(file, false)"
                      title="Protégé contre la suppression">
                    <svg viewBox="0 0 24 24" width="10" height="10" fill="currentColor"><path d="M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zM12 17c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2zm3.1-9H8.9V6c0-1.71 1.39-3.1 3.1-3.1 1.71 0 3.1 1.39 3.1 3.1v2z"/></svg>
                  </button>
                </div>
              </div>
            </div>
            <div v-if="!treeItems.folders.length && !treeItems.files.length" class="tree-empty">
              Dossier vide
            </div>
          </div>
        </template>
      </div>

    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { useFileStore } from '../stores/files';
import { useFriendStore } from '../stores/friends';
import { useAuthStore } from '../stores/auth';
import { useUIStore } from '../stores/ui';
import api from '../api';
import { decryptKeyWithPrivateKey, importKeyFromPEM, encryptKeyWithPublicKey, generateMasterKey } from '../utils/crypto';
import sodium from 'libsodium-wrappers-sumo';

const { t } = useI18n();

const props = defineProps({
  isOpen: Boolean,
  item: Object,
  initialTab: {
    type: String,
    default: 'link'
  }
});

const emit = defineEmits(['close', 'share-deleted', 'share-created']);
const fileStore = useFileStore();
const friendStore = useFriendStore();
const authStore = useAuthStore();
const uiStore = useUIStore();

// UI State
const loading = ref(false);

// Link Share State
const linkCopied = ref(false);
const shareLinkInput = ref(null);
const expiresAt = ref(null);

// Permissions (for folder shares, set at creation time)
const permissions = ref({ download: true, create: true, delete: false, move: false });

// Restrictions panel (post-creation, owner can set per-item overrides)
const restrictionsPanelOpen = ref(false);
const treeLoading = ref(false);
const treeItems = ref({ folders: [], files: [] });
const treePath = ref([]); // [{name, subpath}]
const localPermissions = ref({ download: true, create: false, delete: false, move: false });

// Friends Share State
// directShareStatus[friendId] = { shareId, perm_download, perm_create, perm_delete, perm_move }
const sharing = ref({});
const directShareStatus = ref({});
const friends = computed(() => friendStore.acceptedFriends);

// Local state to handle immediate updates without waiting for parent refresh
const localShareToken = ref(null);
const localShareId = ref(null);
const localExpiresAt = ref(null);

const resourceType = computed(() =>
    (props.item?.type === 'folder' || props.item?.is_dir) ? 'folder' : 'file'
);

const fetchDirectShares = async () => {
    if (!props.item) return;
    try {
        const resourceId = props.item.ID || props.item.id;
        const response = await api.get('/shares/direct', {
            params: { resource_id: resourceId, resource_type: resourceType.value }
        });
        const list = response.data.shared_with || [];
        const map = {};
        list.forEach(info => {
            map[info.user_id] = {
                shareId:      info.share_id,
                perm_download: info.perm_download,
                perm_create:   info.perm_create  ?? false,
                perm_delete:   info.perm_delete  ?? false,
                perm_move:     info.perm_move    ?? false,
            };
        });
        directShareStatus.value = map;
    } catch (e) {
        console.error("Error fetching direct shares:", e);
    }
};

// Reset local state when item changes
watch(() => props.item, (newItem) => {
    if (newItem) {
        localShareToken.value = newItem.share_token || newItem.ShareToken;
        localShareId.value = newItem.share_id || newItem.ShareID;
        localExpiresAt.value = newItem.expires_at || newItem.ExpiresAt;
        localPermissions.value = {
            download: newItem.perm_download ?? true,
            create: newItem.perm_create ?? false,
            delete: newItem.perm_delete ?? false,
            move: newItem.perm_move ?? false,
        };

        directShareStatus.value = {};
        restrictionsPanelOpen.value = false;
        treeItems.value = { folders: [], files: [] };
        treePath.value = [];
        fetchDirectShares();
    }
}, { immediate: true });

onMounted(() => {
    if (friends.value.length === 0) {
        friendStore.fetchFriends();
    }
});

const isShared = computed(() => !!localShareToken.value);
const isDirectShare = computed(() => props.item?.share_token === 'DIRECT');

const formattedExpiration = computed(() => {
  if (!localExpiresAt.value) return null;
  return new Date(localExpiresAt.value).toLocaleString();
});

const shareUrl = computed(() => {
  if (localShareToken.value) {
    return `${window.location.origin}/s/${localShareToken.value}`;
  }
  return '';
});

const selectAll = (e) => {
    e.target.select();
}

// --- Link Sharing Methods ---

const createShare = async () => {
    if (!props.item) return;
    loading.value = true;
    try {
        const itemId = props.item.ID || props.item.id;
        
        // Convert expiresAt to ISO string if present
        let expirationDate = null;
        if (expiresAt.value) {
            const selectedDate = new Date(expiresAt.value);
            if (selectedDate <= new Date()) {
                alert("La date d'expiration doit être dans le futur.");
                loading.value = false;
                return;
            }
            expirationDate = selectedDate.toISOString();
        }

        const result = await fileStore.createShareLink(itemId, props.item.type, expirationDate, permissions.value);

        localShareToken.value = result.token;
        localShareId.value = result.id;
        localExpiresAt.value = expirationDate;
        localPermissions.value = { ...permissions.value };
        
        emit('share-created'); 
        
    } catch (error) {
        console.error("Create share error:", error);
        alert("Erreur lors de la création du partage.");
    } finally {
        loading.value = false;
    }
};

const copyLink = () => {
  if (shareLinkInput.value) {
    shareLinkInput.value.select();
    navigator.clipboard.writeText(shareUrl.value).then(() => {
      linkCopied.value = true;
      setTimeout(() => linkCopied.value = false, 2000);
    }).catch(err => {
      console.error('Impossible de copier le lien:', err);
    });
  }
};

const deleteShare = async () => {
  const idToDelete = localShareId.value || props.item.share_id || props.item.ShareID;
  
  if (!idToDelete) {
      alert("Impossible de supprimer le partage (ID manquant). Veuillez rafraîchir la page.");
      return;
  }
  
  uiStore.requestDeleteConfirmation({
      title: "Arrêter le partage",
      message: "Êtes-vous sûr de vouloir arrêter le partage ? Le lien ne fonctionnera plus.",
      onConfirm: async () => {
        loading.value = true;
        try {
            await api.delete(`/shares/link/${idToDelete}`);
            localShareToken.value = null;
            localShareId.value = null;
            emit('share-deleted');
        } catch (error) {
            console.error('Erreur lors de la suppression du partage:', error);
            alert('Impossible de supprimer le partage.');
        } finally {
            loading.value = false;
        }
      }
  });
};

const togglePermission = async (key) => {
    const shareId = localShareId.value;
    if (!shareId) return;
    const newValue = !localPermissions.value[key];
    localPermissions.value[key] = newValue;
    try {
        const endpoint = isDirectShare.value
            ? `/shares/direct/${shareId}/permissions`
            : `/shares/link/${shareId}/permissions`;
        await api.patch(endpoint, { [`perm_${key}`]: newValue });
    } catch (e) {
        console.error('Failed to update permission:', e);
        localPermissions.value[key] = !newValue; // revert
    }
};

// --- Friends Sharing Methods ---

const isFriendShared = (friendId) => !!directShareStatus.value[friendId];

const toggleDirectPerm = async (friend, key) => {
    const info = directShareStatus.value[friend.id];
    if (!info?.shareId) return;
    const newValue = !info[`perm_${key}`];
    info[`perm_${key}`] = newValue;
    try {
        await api.patch(`/shares/direct/${info.shareId}/permissions`, {
            resource_type: resourceType.value,
            [`perm_${key}`]: newValue,
        });
    } catch (e) {
        console.error('Failed to update direct share permission:', e);
        info[`perm_${key}`] = !newValue;
    }
};

// Decrypt a key encrypted with master key (AES-GCM, IV prepended)
async function decryptWithMasterKey(encryptedB64, masterKey) {
    const encBytes = sodium.from_base64(encryptedB64);
    const iv = encBytes.slice(0, 12);
    const data = encBytes.slice(12);
    return window.crypto.subtle.decrypt({ name: "AES-GCM", iv }, masterKey, data);
}

// Encrypt a raw key buffer with a CryptoKey, returning IV+ciphertext as base64
async function encryptWithKey(rawKeyBuffer, cryptoKey) {
    const iv = window.crypto.getRandomValues(new Uint8Array(12));
    const encrypted = await window.crypto.subtle.encrypt({ name: "AES-GCM", iv }, cryptoKey, rawKeyBuffer);
    const combined = new Uint8Array(iv.byteLength + encrypted.byteLength);
    combined.set(iv);
    combined.set(new Uint8Array(encrypted), iv.byteLength);
    return sodium.to_base64(combined);
}

// Encrypt child file and folder keys with the parent folder key
async function encryptChildKeys(files, subFolders, folderKeyCrypto, masterKey) {
    const folderFileKeys = {};
    for (const file of files) {
        const fEncKey = file.EncryptedKey || file.encrypted_key;
        if (!fEncKey) { console.warn(`File ${file.Name} has no key, skipping.`); continue; }
        try {
            const fileKeyRaw = await decryptWithMasterKey(fEncKey, masterKey);
            folderFileKeys[file.ID || file.id] = await encryptWithKey(fileKeyRaw, folderKeyCrypto);
        } catch(err) { console.warn(`Failed to process key for file ${file.Name}:`, err); }
    }
    const folderFolderKeys = {};
    for (const folder of subFolders) {
        const fEncKey = folder.EncryptedKey || folder.encrypted_key;
        if (!fEncKey) { console.warn(`Folder ${folder.Name} has no key, skipping.`); continue; }
        try {
            const subFolderKeyRaw = await decryptWithMasterKey(fEncKey, masterKey);
            folderFolderKeys[folder.ID || folder.id] = await encryptWithKey(subFolderKeyRaw, folderKeyCrypto);
        } catch(err) { console.warn(`Failed to process key for folder ${folder.Name}:`, err); }
    }
    return { folderFileKeys, folderFolderKeys };
}

// Get or create a folder key, persisting a new one to the backend if needed
async function getOrCreateFolderKey(itemId, existingEncKey, masterKey, item) {
    if (existingEncKey) {
        const folderKeyRaw = await decryptWithMasterKey(existingEncKey, masterKey);
        const folderKeyCrypto = await window.crypto.subtle.importKey(
            "raw", folderKeyRaw, { name: "AES-GCM" }, false, ["encrypt", "decrypt"]
        );
        return { folderKeyRaw, folderKeyCrypto };
    }
    const folderKeyCrypto = await generateMasterKey();
    const folderKeyRaw = await window.crypto.subtle.exportKey("raw", folderKeyCrypto);
    const encryptedKeyBase64 = await encryptWithKey(folderKeyRaw, masterKey);
    await api.put(`/folders/${itemId}/key`, { encrypted_key: encryptedKeyBase64 });
    if (item) {
        item.encrypted_key = encryptedKeyBase64;
        item.EncryptedKey = encryptedKeyBase64;
    }
    return { folderKeyRaw, folderKeyCrypto };
}

// Resolve the full folder path from item properties
function resolveFolderPath(item) {
    const itemName = item.Name || item.name;
    let itemParentPath = item.Path || item.path || '/';
    if (!itemParentPath) itemParentPath = '/';
    if (itemParentPath.endsWith('/' + itemName)) {
        console.warn("Detected Path might be full path. Using as is:", itemParentPath);
        return itemParentPath;
    }
    return (itemParentPath === '/' ? '' : itemParentPath) + '/' + itemName;
}

const shareWithFriend = async (friend) => {
    if (!props.item || !friend.public_key) return;

    if (isFriendShared(friend.id)) {
        uiStore.requestDeleteConfirmation({
           title: "Arrêter le partage",
           message: `Arrêter le partage avec ${friend.name} ?`,
           onConfirm: async () => {
             sharing.value[friend.id] = true;
             try {
                const resourceId = props.item.ID || props.item.id;
                await api.delete(`/shares/direct`, {
                    params: { resource_id: resourceId, resource_type: resourceType.value, friend_id: friend.id }
                });
                delete directShareStatus.value[friend.id];
             } catch(e) {
                console.error("Revoke failed:", e);
                if (e.response && e.response.status === 404) {
                     delete directShareStatus.value[friend.id];
                } else {
                     alert("Erreur lors de la suppression du partage.");
                }
             } finally {
                sharing.value[friend.id] = false;
             }
           }
        });
        return;
    }

    sharing.value[friend.id] = true;
    try {
        await sodium.ready;

        if (resourceType.value === 'file') {
             const itemEncryptedKey = props.item.EncryptedKey || props.item.encrypted_key;
             if (!itemEncryptedKey) throw new Error("Clé du fichier manquante. Impossible de partager.");
             if (!authStore.masterKey) throw new Error("Clé Maître non disponible (Session expirée ?). Veuillez vous reconnecter.");
             if (!authStore.privateKey) throw new Error("Clé privée non disponible. Veuillez vous reconnecter.");

             const fileKeyRawBuffer = await decryptWithMasterKey(itemEncryptedKey, authStore.masterKey);
             const friendPublicKey = await importKeyFromPEM(friend.public_key, 'spki');
             const encryptedKeyForFriend = await encryptKeyWithPublicKey(fileKeyRawBuffer, friendPublicKey);
             await api.post('/shares/direct', {
                resource_id: props.item.ID || props.item.id,
                resource_type: resourceType.value,
                friend_id: friend.id,
                encrypted_key: encryptedKeyForFriend,
                permission: 'read',
                perm_download: true,
             });
             await fetchDirectShares();

        } else if (resourceType.value === 'folder') {
            if (!authStore.masterKey) throw new Error("Clé Maître non disponible (Session expirée ?). Veuillez vous reconnecter.");

            const itemId = props.item.ID || props.item.id;
            const existingEncKey = props.item.EncryptedKey || props.item.encrypted_key;
            const { folderKeyRaw, folderKeyCrypto } = await getOrCreateFolderKey(
                itemId, existingEncKey, authStore.masterKey, props.item
            );

            const friendPublicKey = await importKeyFromPEM(friend.public_key, 'spki');
            const encryptedKeyForFriend = await encryptKeyWithPublicKey(folderKeyRaw, friendPublicKey);

            const folderPath = resolveFolderPath(props.item);
            const listRes = await api.get(`/files/list-recursive?path=${encodeURIComponent(folderPath)}`);
            const files = listRes.data.files || [];
            const subFolders = listRes.data.folders || [];

            const { folderFileKeys, folderFolderKeys } = await encryptChildKeys(
                files, subFolders, folderKeyCrypto, authStore.masterKey
            );

            await api.post('/shares/direct', {
                resource_id: props.item.ID || props.item.id,
                resource_type: resourceType.value,
                friend_id: friend.id,
                encrypted_key: encryptedKeyForFriend,
                permission: 'read',
                perm_download: true,
                perm_create: false,
                perm_delete: false,
                perm_move: false,
                folder_file_keys: folderFileKeys,
                folder_folder_keys: folderFolderKeys
            });
            await fetchDirectShares();
        }

    } catch (e) {
        console.error("Partage échoué:", e);
        alert("Erreur: " + e.message);
    } finally {
        sharing.value[friend.id] = false;
    }
}


// --- Restrictions panel methods ---

const applyBulkOverride = async (itemType, options) => {
    const shareId = localShareId.value;
    if (!shareId) return;
    treeLoading.value = true;
    try {
        await api.post(`/shares/link/${shareId}/overrides/bulk`, {
            item_type: itemType,
            ...options,
        });
        await loadTreeLevel(treePath.value[treePath.value.length - 1]?.subpath || '/');
    } catch (e) {
        console.error('Bulk override failed:', e);
    } finally {
        treeLoading.value = false;
    }
};

const toggleRestrictionsPanel = async () => {
    restrictionsPanelOpen.value = !restrictionsPanelOpen.value;
    if (restrictionsPanelOpen.value && !treeItems.value.folders.length && !treeItems.value.files.length) {
        const itemName = props.item?.Name || props.item?.name || '';
        treePath.value = [{ name: itemName, subpath: '/' }];
        await loadTreeLevel('/');
    }
};

const loadTreeLevel = async (subpath) => {
    const shareId = localShareId.value;
    if (!shareId) return;
    treeLoading.value = true;
    try {
        const path = subpath === '/' ? '/' : subpath;
        const response = await api.get(`/shares/link/${shareId}/browse${path}`);
        treeItems.value = {
            folders: response.data.folders || [],
            files: response.data.files || [],
        };
    } catch (e) {
        console.error('Tree load failed:', e);
    } finally {
        treeLoading.value = false;
    }
};

const drillDown = async (folder) => {
    const parentSubpath = treePath.value[treePath.value.length - 1].subpath;
    const newSubpath = parentSubpath === '/' ? `/${folder.Name}` : `${parentSubpath}/${folder.Name}`;
    treePath.value.push({ name: folder.Name, subpath: newSubpath });
    await loadTreeLevel(newSubpath);
};

const navigateTree = async (index) => {
    if (index === treePath.value.length - 1) return;
    treePath.value = treePath.value.slice(0, index + 1);
    await loadTreeLevel(treePath.value[index].subpath);
};

const setFolderAccess = async (folder, level) => {
    const shareId = localShareId.value;
    if (!shareId) return;
    try {
        await api.put(`/shares/link/${shareId}/overrides`, {
            item_path: folder.Path,
            item_type: 'folder',
            access_level: level,
            can_delete: folder.can_delete !== false,
            can_download: folder.can_download !== false,
        });
        folder.access_level = level;
    } catch (e) {
        console.error('Failed to set folder access:', e);
    }
};

const setFolderDelete = async (folder, canDelete) => {
    const shareId = localShareId.value;
    if (!shareId) return;
    try {
        await api.put(`/shares/link/${shareId}/overrides`, {
            item_path: folder.Path,
            item_type: 'folder',
            access_level: folder.access_level || 'full',
            can_delete: canDelete,
            can_download: folder.can_download !== false,
        });
        folder.can_delete = canDelete;
    } catch (e) {
        console.error('Failed to set folder delete:', e);
    }
};

const setFileDelete = async (file, canDelete) => {
    const shareId = localShareId.value;
    if (!shareId) return;
    try {
        await api.put(`/shares/link/${shareId}/overrides`, {
            item_path: file.Path,
            item_type: 'file',
            access_level: 'full',
            can_delete: canDelete,
            can_download: file.can_download !== false,
        });
        file.can_delete = canDelete;
    } catch (e) {
        console.error('Failed to set file delete:', e);
    }
};

const setFileDownload = async (file, canDownload) => {
    const shareId = localShareId.value;
    if (!shareId) return;
    try {
        await api.put(`/shares/link/${shareId}/overrides`, {
            item_path: file.Path,
            item_type: 'file',
            access_level: 'full',
            can_delete: file.can_delete !== false,
            can_download: canDownload,
        });
        file.can_download = canDownload;
    } catch (e) {
        console.error('Failed to set file download:', e);
    }
};

const close = () => {
  emit('close');
  directShareStatus.value = {};
  restrictionsPanelOpen.value = false;
  treeItems.value = { folders: [], files: [] };
  treePath.value = [];
  permissions.value = { download: true, create: true, delete: false, move: false };
};
</script>

<style scoped>
.form-group {
    margin-bottom: 1rem;
    text-align: left;
    width: 100%;
    max-width: 300px;
    margin-left: auto;
    margin-right: auto;
}

.form-group label {
    display: block;
    margin-bottom: 0.5rem;
    font-size: 0.9rem;
    color: var(--secondary-text-color);
}

.form-control {
    width: 100%;
    padding: 8px 12px;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    font-size: 0.9rem;
    background-color: var(--card-color);
    color: var(--main-text-color);
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.6);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
  animation: fadeIn 0.2s ease;
}

.modal-wrapper {
  display: flex;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 10px 25px rgba(0,0,0,0.2);
  max-width: 92vw;
  max-height: 90vh;
}

.modal-content {
  background: var(--card-color);
  padding: 0;
  width: 480px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}

/* Side panel */
.modal-side-panel {
  width: 360px;
  flex-shrink: 0;
  border-left: 1px solid var(--border-color);
  background: var(--background-color);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.side-panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px;
  border-bottom: 1px solid var(--border-color);
  background: var(--card-color);
  font-weight: 600;
  font-size: 0.9rem;
  color: var(--main-text-color);
  flex-shrink: 0;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 20px;
  border-bottom: 1px solid var(--border-color);
}

.modal-header h3 {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--main-text-color);
}

.btn-close {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: var(--secondary-text-color);
  padding: 0;
  line-height: 1;
}

.modal-body {
  padding: 14px 20px;
  min-height: 100px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: var(--secondary-text-color);
}

.not-shared-state {
  text-align: center;
}

.illustration {
  font-size: 2rem;
  margin-bottom: 0.4rem;
}

.sub-text {
  color: var(--secondary-text-color);
  margin-bottom: 0.75rem;
  font-size: 0.9rem;
}

.shared-state {
  display: flex;
  flex-direction: column;
  gap: 0.7rem;
}

.link-section label {
  display: block;
  font-size: 0.85rem;
  font-weight: 500;
  color: var(--secondary-text-color);
  margin-bottom: 0.5rem;
}

.link-container {
  display: flex;
  gap: 10px;
}

.link-container input {
  flex-grow: 1;
  padding: 10px 12px;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  background-color: var(--background-color);
  color: var(--main-text-color);
  font-size: 0.9rem;
  outline: none;
}

.link-container input:focus {
  border-color: var(--primary-color);
  background-color: var(--card-color);
}

.share-info {
  background-color: var(--background-color);
  color: var(--primary-color);
  padding: 12px;
  border-radius: 4px;
  font-size: 0.85rem;
  display: flex;
  align-items: center;
  border: 1px solid var(--primary-color);
}

.share-info p {
  margin: 0;
}

.modal-footer {
  padding: 10px 20px;
  border-top: 1px solid var(--border-color);
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  background-color: var(--background-color);
}

button {
  padding: 8px 16px;
  border-radius: 4px;
  border: 1px solid transparent;
  cursor: pointer;
  font-weight: 500;
  font-size: 0.9rem;
  transition: background-color 0.2s;
}

.btn-primary {
  background-color: var(--primary-color);
  color: white;
}

.btn-primary:hover {
  background-color: var(--accent-color);
  box-shadow: 0 1px 2px rgba(60,64,67,0.3);
}

.btn-secondary {
  background-color: var(--card-color);
  border: 1px solid var(--border-color);
  color: var(--main-text-color);
}

.btn-secondary:hover {
  background-color: var(--hover-background-color);
  border-color: var(--border-color);
}

.btn-copy {
  background-color: var(--card-color);
  border: 1px solid var(--border-color);
  color: var(--primary-color);
  min-width: 80px;
}

.btn-copy:hover {
  background-color: var(--hover-background-color);
}

.btn-copy.copied {
  background-color: var(--success-color);
  color: white;
  border-color: transparent;
}

.btn-delete {
  background-color: transparent;
  color: var(--error-color);
  margin-right: auto; /* Push to left */
}

.btn-delete:hover {
  background-color: var(--hover-background-color);
}

.btn-danger {
  background-color: var(--error-color);
  color: white;
  border: 1px solid var(--error-color);
}

.btn-danger:hover {
  filter: brightness(0.9);
}

.spinner {
  border: 3px solid var(--border-color);
  border-radius: 50%;
  border-top: 3px solid var(--primary-color);
  width: 20px;
  height: 20px;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.section-title {
  margin: 0 0 8px 0;
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--main-text-color);
}

.section-divider {
  height: 1px;
  background-color: var(--border-color);
  margin: 10px 0;
}

.link-section-wrapper {
  margin-top: 10px;
}

.friends-list {
  max-height: 280px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.friend-entry {
  border: 1px solid var(--border-color);
  border-radius: 6px;
  overflow: hidden;
}

.friend-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 5px 8px;
}

.friend-entry:hover .friend-item {
  background-color: var(--hover-background-color);
}

.friend-perms {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  padding: 5px 8px;
  border-top: 1px solid var(--border-color);
  background-color: var(--background-color);
}

.perm-friend-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 8px;
  border-radius: 12px;
  border: 1px solid var(--border-color);
  font-size: 0.75rem;
  font-weight: 500;
  cursor: pointer;
  background: transparent;
  color: var(--secondary-text-color);
  transition: all 0.15s;
  user-select: none;
}

.perm-friend-chip.active {
  background: rgba(22, 163, 74, 0.1);
  border-color: #16a34a;
  color: #16a34a;
}

.perm-friend-chip:not(.active) {
  background: rgba(220, 38, 38, 0.05);
  border-color: rgba(220, 38, 38, 0.25);
  color: #b91c1c;
}

.perm-friend-chip:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
  background: rgba(99, 102, 241, 0.08);
}

.friend-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.friend-avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background-color: var(--primary-color);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: 0.8rem;
}

.friend-name {
  font-weight: 500;
  margin: 0;
  color: var(--main-text-color);
}

.friend-email {
  font-size: 0.8rem;
  color: var(--secondary-text-color);
  margin: 0;
}

.btn-sm {
  padding: 4px 10px;
  font-size: 0.8rem;
}

.btn-success {
  background-color: var(--success-color);
  color: white;
  cursor: default;
}

.btn-outline {
  background-color: transparent;
  border: 1px solid var(--primary-color);
  color: var(--primary-color);
}

.btn-outline:hover {
  background-color: var(--primary-color);
  color: white;
}

.warning-box {
  background-color: var(--card-color);
  color: var(--warning-color);
  padding: 10px;
  border-radius: 4px;
  margin-bottom: 15px;
  font-size: 0.9rem;
  border: 1px solid var(--warning-color);
}

.empty-friends {
  text-align: center;
  color: var(--secondary-text-color);
  padding: 20px;
}

/* Permission chips (creation form) */
.perm-group {
  margin-bottom: 1.25rem;
  text-align: left;
  width: 100%;
  max-width: 300px;
  margin-left: auto;
  margin-right: auto;
}

.perm-label {
  display: block;
  margin-bottom: 0.5rem;
  font-size: 0.9rem;
  color: var(--secondary-text-color);
}

.perm-chips {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.perm-chip {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  border-radius: 20px;
  border: 1px solid var(--border-color);
  font-size: 0.82rem;
  font-weight: 500;
  cursor: pointer;
  color: var(--secondary-text-color);
  background: var(--background-color);
  transition: all 0.15s;
  user-select: none;
}

.perm-chip input[type="checkbox"] {
  display: none;
}

.perm-chip.active {
  background: rgba(22, 163, 74, 0.1);
  border-color: #16a34a;
  color: #16a34a;
}

.perm-chip:not(.active) {
  background: rgba(220, 38, 38, 0.06);
  border-color: rgba(220, 38, 38, 0.3);
  color: #b91c1c;
}

.perm-chip:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
  background: rgba(99, 102, 241, 0.1);
}

/* Rights card (visible section in shared state) */
.rights-card {
  border: 1px solid var(--border-color);
  border-radius: 8px;
  overflow: hidden;
  background: var(--background-color);
}

.rights-card-title {
  display: flex;
  align-items: center;
  gap: 7px;
  padding: 10px 14px;
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--main-text-color);
  background: var(--card-color);
  border-bottom: 1px solid var(--border-color);
}

.rights-rows {
  padding: 4px 0;
}

.rights-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 14px;
  font-size: 0.85rem;
  color: var(--secondary-text-color);
}

.rights-row.active {
  color: #16a34a;
  background: rgba(22, 163, 74, 0.07);
  border-radius: 6px;
}

.rights-row svg {
  flex-shrink: 0;
  opacity: 0.5;
}

.rights-row.active svg {
  opacity: 1;
  color: #16a34a;
}

.rights-row.clickable:not(.active) {
  color: #b91c1c;
  background: rgba(220, 38, 38, 0.05);
}

.rights-row.clickable:not(.active) svg {
  color: #b91c1c;
  opacity: 0.7;
}

.rights-row span:first-of-type {
  flex: 1;
}

.rights-row.clickable {
  cursor: pointer;
  border-radius: 6px;
  transition: background 0.15s;
}

.rights-row.clickable:hover {
  background: var(--hover-background-color);
  color: var(--main-text-color);
}

.rights-row.clickable:hover svg {
  color: var(--primary-color);
  opacity: 1;
}

.rights-toggle {
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

.rights-row:not(.active) .rights-toggle {
  color: #b91c1c;
  opacity: 0.7;
}

.rights-row.active .rights-toggle {
  color: #16a34a;
}

.btn-manage-restrictions {
  display: flex;
  align-items: center;
  gap: 7px;
  width: 100%;
  padding: 10px 14px;
  background: none;
  border: none;
  border-top: 1px solid var(--border-color);
  color: var(--primary-color);
  font-size: 0.85rem;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.15s;
  text-align: left;
}

.btn-manage-restrictions:hover {
  background: rgba(99, 102, 241, 0.06);
}

.btn-manage-restrictions .arrow {
  margin-left: auto;
  font-size: 0.9rem;
}

.tree-loading {
  padding: 12px;
  font-size: 0.85rem;
  color: var(--secondary-text-color);
  text-align: center;
}

.tree-nav {
  display: flex;
  flex-wrap: wrap;
  gap: 2px;
  padding: 8px 10px;
  border-bottom: 1px solid var(--border-color);
  background: var(--card-color);
  flex-shrink: 0;
}

.tree-crumb {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 0.8rem;
  color: var(--secondary-text-color);
  padding: 2px 4px;
  border-radius: 4px;
}

.tree-crumb.active {
  color: var(--main-text-color);
  font-weight: 600;
  cursor: default;
}

.tree-crumb:hover:not(.active) {
  color: var(--primary-color);
  background: var(--hover-background-color);
}

.tree-items {
  flex: 1;
  overflow-y: auto;
}

.tree-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 10px;
  border-bottom: 1px solid var(--border-color);
  gap: 8px;
}

.tree-item:last-child {
  border-bottom: none;
}

.tree-item-name {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 0.82rem;
  color: var(--main-text-color);
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
  text-align: left;
}

.tree-item-name.static {
  cursor: default;
}

.tree-item-name:hover:not(.static) {
  color: var(--primary-color);
}

.access-toggle, .delete-toggle, .download-toggle, .folder-extra-toggle {
  display: flex;
  gap: 2px;
  flex-shrink: 0;
}

.folder-controls {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
}

.file-controls {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
}

.access-opt {
  background: none;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  padding: 2px 7px;
  font-size: 0.75rem;
  cursor: pointer;
  color: var(--secondary-text-color);
  transition: all 0.12s;
}

.access-opt:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
}

.access-opt.active {
  background: rgba(99, 102, 241, 0.12);
  border-color: var(--primary-color);
  color: var(--primary-color);
}

.tree-empty {
  padding: 16px;
  text-align: center;
  font-size: 0.82rem;
  color: var(--secondary-text-color);
  font-style: italic;
}

.bulk-controls {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 8px 10px;
  border-bottom: 1px solid var(--border-color);
  background: var(--card-color);
  flex-shrink: 0;
}

.bulk-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.bulk-label {
  font-size: 0.78rem;
  color: var(--secondary-text-color);
  font-weight: 500;
}

.bulk-btns {
  display: flex;
  gap: 3px;
}
</style>
