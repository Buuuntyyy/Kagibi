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

      <!-- MFA required gate -->
      <div v-if="orgMFARequired" class="mfa-required-overlay">
        <svg viewBox="0 0 24 24" width="48" height="48" fill="currentColor"><path d="M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zm-6 9c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2zm3.1-9H8.9V6c0-1.71 1.39-3.1 3.1-3.1 1.71 0 3.1 1.39 3.1 3.1v2z"/></svg>
        <h3>{{ t('orgs.mfaRequired') }}</h3>
        <p>{{ t('orgs.mfaRequiredHint') }}</p>
        <router-link to="/account" class="btn-primary">{{ t('nav.account') }}</router-link>
      </div>

      <!-- TAB: FILES -->
      <div v-else-if="activeTab === 'files'" class="tab-content"
        :class="{ 'drop-zone-active': isDragOver }"
        @dragenter.prevent="onDragEnterZone"
        @dragleave="onDragLeaveZone"
        @dragover.prevent="onDragOverZone"
        @drop.prevent="onDropFiles"
      >
        <div v-if="isDragOver && canWrite" class="drop-overlay">
          <svg viewBox="0 0 24 24" width="48" height="48" fill="currentColor"><path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z"/></svg>
          <p>{{ t('orgs.dropToUpload') }}</p>
        </div>
        <!-- Search bar -->
        <div class="search-bar-wrap">
          <svg class="search-icon" viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M15.5 14h-.79l-.28-.27A6.471 6.471 0 0 0 16 9.5 6.5 6.5 0 1 0 9.5 16c1.61 0 3.09-.59 4.23-1.57l.27.28v.79l5 4.99L20.49 19l-4.99-5zm-6 0C7.01 14 5 11.99 5 9.5S7.01 5 9.5 5 14 7.01 14 9.5 11.99 14 9.5 14z"/></svg>
          <input
            ref="searchInputRef"
            v-model="searchQuery"
            class="search-input"
            type="text"
            :placeholder="t('orgs.searchPlaceholder')"
            @input="onSearchInput"
            @keydown.escape="clearSearch"
          />
          <button v-if="searchQuery" class="search-clear" @click="clearSearch">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
          </button>
          <span v-if="searchLoading" class="spinner-sm-dark" style="margin-left:6px;flex-shrink:0"></span>
        </div>

        <!-- Search results panel -->
        <div v-if="searchQuery && !searchLoading" class="search-results">
          <div v-if="searchResults.length === 0" class="search-empty">
            {{ t('orgs.searchNoResults', { q: searchQuery }) }}
          </div>
          <template v-else>
            <div class="search-count">{{ t('orgs.searchResultCount', { count: searchResults.length }) }}</div>
            <div
              v-for="item in searchResults"
              :key="item.type + '-' + item.id"
              class="search-result-row"
              @click="navigateToSearchResult(item)"
            >
              <svg v-if="item.type === 'folder'" viewBox="0 0 24 24" width="16" height="16" fill="currentColor" class="search-result-icon folder"><path d="M10 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z"/></svg>
              <svg v-else viewBox="0 0 24 24" width="16" height="16" fill="currentColor" class="search-result-icon file"><path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z"/></svg>
              <div class="search-result-body">
                <span class="search-result-name" v-html="highlightMatch(item.decrypted_name, searchQuery)"></span>
                <span class="search-result-path">{{ item.parent_path === '/' ? '/' : item.parent_path }}</span>
              </div>
              <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor" class="search-result-arrow"><path d="M10 6L8.59 7.41 13.17 12l-4.58 4.59L10 18l6-6z"/></svg>
            </div>
          </template>
        </div>

        <div class="fs-toolbar" v-show="!searchQuery">
          <div class="breadcrumb">
            <button class="bc-item" @click="navigateToPath('/')"
              :class="{ 'bc-drag-over': dragOverBcPath === '/' }"
              @dragover="onBcDragOver($event, '/')"
              @dragleave="onBcDragLeave('/')"
              @drop.prevent="onDropOnPath($event, '/')"
            >
              <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M10 20v-6h4v6h5v-8h3L12 3 2 12h3v8z"/></svg>
            </button>
            <template v-for="(seg, idx) in pathSegments" :key="idx">
              <span class="bc-sep">/</span>
              <button class="bc-item" @click="navigateToPath(buildPath(idx))"
                :class="{ 'bc-drag-over': dragOverBcPath === buildPath(idx) }"
                @dragover="onBcDragOver($event, buildPath(idx))"
                @dragleave="onBcDragLeave(buildPath(idx))"
                @drop.prevent="onDropOnPath($event, buildPath(idx))"
              >{{ orgStore.folderNameCache[seg] || seg }}</button>
            </template>
          </div>
          <div class="fs-actions">
            <button class="btn-sm btn-ghost-sm" @click="showTagManager = true">
              <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M21.41 11.58l-9-9C12.05 2.22 11.55 2 11 2H4c-1.1 0-2 .9-2 2v7c0 .55.22 1.05.59 1.42l9 9c.36.36.86.58 1.41.58s1.05-.22 1.41-.59l7-7c.37-.36.59-.86.59-1.41s-.23-1.06-.59-1.42zM5.5 7C4.67 7 4 6.33 4 5.5S4.67 4 5.5 4 7 4.67 7 5.5 6.33 7 5.5 7z"/></svg>
              {{ t('orgs.manageTags') }}
            </button>
            <template v-if="canWrite">
              <button class="btn-sm" @click="showNewFolderModal = true">
                <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M10 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z"/></svg>
                {{ t('orgs.newFolder') }}
              </button>
              <label class="btn-sm btn-upload">
                <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z"/></svg>
                {{ t('orgs.uploadFile') }}
                <input type="file" multiple style="display:none" @change="handleFileUpload" />
              </label>
            </template>
          </div>
        </div>

        <!-- Sort / filter bar -->
        <div class="sort-filter-bar" v-show="!searchQuery">
          <div class="sort-group">
            <button
              v-for="field in ['name', 'size', 'date']"
              :key="field"
              class="sort-btn"
              :class="{ active: sortBy === field }"
              @click="toggleSort(field)"
            >
              {{ t(`orgs.sort${capitalize(field)}`) }}
              <svg v-if="sortBy === field" viewBox="0 0 24 24" width="12" height="12" fill="currentColor">
                <path v-if="sortDir === 'asc'" d="M7 14l5-5 5 5z"/>
                <path v-else d="M7 10l5 5 5-5z"/>
              </svg>
            </button>
          </div>
          <div class="sort-filter-divider"></div>
          <div class="filter-group">
            <button
              v-for="type in ['all', 'images', 'documents', 'videos', 'audio', 'archives']"
              :key="type"
              class="filter-btn"
              :class="{ active: filterType === type }"
              @click="filterType = type"
            >
              {{ t(`orgs.filter${capitalize(type)}`) }}
            </button>
          </div>
          <template v-if="orgStore.orgTags.length > 0">
            <div class="sort-filter-divider"></div>
            <div class="tag-filter-group">
              <button
                v-for="tag in orgStore.orgTags"
                :key="tag.id"
                class="tag-filter-btn"
                :class="{ active: filterTagID === tag.id }"
                :style="{ '--tag-color': tag.color }"
                @click="filterTagID = filterTagID === tag.id ? null : tag.id"
              >
                <span class="tag-dot"></span>{{ tag.name }}
              </button>
            </div>
          </template>
        </div>

        <!-- Pinned / favorites quick-access strip -->
        <div v-if="orgStore.favorites.length > 0 && !searchQuery" class="pinned-section">
          <div class="pinned-label">
            <svg viewBox="0 0 24 24" width="13" height="13" fill="currentColor" style="flex-shrink:0"><path d="M12 17.27L18.18 21l-1.64-7.03L22 9.24l-7.19-.61L12 2 9.19 8.63 2 9.24l5.46 4.73L5.82 21z"/></svg>
            {{ t('orgs.pinned') }}
          </div>
          <div class="pinned-chips">
            <button
              v-for="fav in enrichedFavorites"
              :key="fav.id"
              class="pinned-chip"
              @click="onFavClick(fav)"
              :title="fav._path || fav.item_type"
            >
              <svg v-if="fav.item_type === 'folder'" viewBox="0 0 24 24" width="13" height="13" fill="currentColor"><path d="M10 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z"/></svg>
              <svg v-else viewBox="0 0 24 24" width="13" height="13" fill="currentColor"><path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z"/></svg>
              <span class="pinned-chip-name">{{ fav._name || `#${fav.item_id}` }}</span>
              <span class="pinned-chip-remove" @click.stop="toggleFavorite({ id: fav.item_id, name: fav._name }, fav.item_type)" :title="t('orgs.unpinItem')">
                <svg viewBox="0 0 24 24" width="11" height="11" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
              </span>
            </button>
          </div>
        </div>

        <!-- Member joined via link but admin hasn't provisioned their key yet -->
        <div v-if="!orgStore.currentOrg?.my_encrypted_org_key && !canManage" class="key-pending-banner">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/></svg>
          <div class="key-pending-text">
            <strong>{{ t('orgs.keyPending') }}</strong>
            <span>{{ t('orgs.keyPendingHint') }}</span>
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
          <div v-if="sortedFolders.length === 0 && sortedFiles.length === 0" class="empty-folder">
            <svg viewBox="0 0 24 24" width="48" height="48" fill="currentColor" style="opacity:0.25"><path d="M20 6h-8l-2-2H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2zm0 12H4V8h16v10z"/></svg>
            <p>{{ filterType !== 'all' ? t('orgs.filterNoResults') : t('orgs.emptyFolder') }}</p>
          </div>

          <!-- Bulk action bar -->
          <div v-if="hasSelection && canWrite" class="bulk-bar">
            <span class="bulk-count">{{ t('orgs.bulkSelected', { count: selectedCount }) }}</span>
            <div class="bulk-actions">
              <button class="btn-sm" @click="bulkDownload" :disabled="bulkLoading || (selectedFileItems.length === 0 && selectedFolderItems.length === 0)" :title="selectedFolderItems.length > 0 ? t('orgs.downloadZip') : t('orgs.bulkDownload')">
                <span v-if="bulkLoading" class="spinner-sm"></span>
                <svg v-else viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M19 9h-4V3H9v6H5l7 7 7-7zM5 18v2h14v-2H5z"/></svg>
                {{ selectedFolderItems.length > 0 ? t('orgs.downloadZip') : t('orgs.bulkDownload') }}
              </button>
              <button class="btn-sm" @click="openBulkMoveDialog" :disabled="bulkLoading">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M20 6h-8l-2-2H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2zm-2 8h-3v3h-2v-3h-3v-2h3V9h2v3h3v2z"/></svg>
                {{ t('orgs.bulkMove') }}
              </button>
              <button class="btn-sm btn-danger" @click="bulkDelete" :disabled="bulkLoading">
                <span v-if="bulkLoading" class="spinner-sm"></span>
                <svg v-else viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"/></svg>
                {{ t('orgs.bulkDelete') }}
              </button>
              <button class="btn-sm btn-ghost" @click="clearSelection">{{ t('orgs.bulkCancel') }}</button>
            </div>
          </div>

          <div class="items-list">
            <!-- Select-all row -->
            <div v-if="sortedFolders.length + sortedFiles.length > 0" class="select-all-row">
              <label class="checkbox-wrap" @click.stop>
                <input type="checkbox" :checked="allVisibleSelected" @change="toggleSelectAll" class="item-checkbox" />
              </label>
              <span class="select-all-label">{{ allVisibleSelected ? t('orgs.deselectAll') : t('orgs.selectAll') }}</span>
            </div>

            <!-- Folders -->
            <div
              v-for="folder in sortedFolders"
              :key="'f-' + folder.id"
              class="item-row folder-row"
              :class="{ selected: isSelected('folder', folder.id), 'drag-over': dragOverFolderID === folder.id, 'folder-active': activeFolderID === folder.id }"
              :draggable="canWrite"
              @dragstart="onItemDragStart($event, folder, 'folder')"
              @dragend="onItemDragEnd"
              @dragover="onFolderDragOver($event, folder)"
              @dragleave="onFolderDragLeave(folder)"
              @drop="onDropOnFolder($event, folder)"
              @click="renamingItem?.id !== folder.id && selectFolderRow(folder.id)"
              @dblclick="renamingItem?.id !== folder.id && navigateToPath(folder.path)"
            >
              <label class="checkbox-wrap" @click.stop>
                <input type="checkbox" :checked="isSelected('folder', folder.id)" @change="e => toggleSelect(e, 'folder', folder.id)" class="item-checkbox" />
              </label>
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor" class="item-icon folder-icon"><path d="M10 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z"/></svg>
              <input
                v-if="renamingItem?.id === folder.id"
                :id="`rename-input-${folder.id}`"
                class="item-rename-input"
                v-model="renamingItem.value"
                :placeholder="t('orgs.renamePlaceholder')"
                @keydown.enter.prevent="saveRename"
                @keydown.escape.prevent="cancelRename"
                @blur="saveRename"
                @click.stop
              />
              <span v-else class="item-name">{{ folder.name }}</span>
              <div v-if="renamingItem?.id !== folder.id && (folder.tag_ids || []).length > 0" class="item-tags">
                <span
                  v-for="tagID in (folder.tag_ids || []).slice(0, 3)"
                  :key="tagID"
                  class="item-tag-dot"
                  :style="{ background: tagColor(tagID) }"
                  :title="tagName(tagID)"
                ></span>
                <span v-if="(folder.tag_ids || []).length > 3" class="item-tag-more">+{{ (folder.tag_ids || []).length - 3 }}</span>
              </div>
              <span v-if="renamingItem?.id !== folder.id" class="item-meta">{{ formatDate(folder.created_at) }}</span>
              <div class="item-actions">
                <button class="btn-icon" @click.stop="openTagPopover($event, folder.id, 'folder')" :title="t('orgs.addTag')">
                  <svg viewBox="0 0 24 24" width="13" height="13" fill="currentColor"><path d="M21.41 11.58l-9-9C12.05 2.22 11.55 2 11 2H4c-1.1 0-2 .9-2 2v7c0 .55.22 1.05.59 1.42l9 9c.36.36.86.58 1.41.58s1.05-.22 1.41-.59l7-7c.37-.36.59-.86.59-1.41s-.23-1.06-.59-1.42zM5.5 7C4.67 7 4 6.33 4 5.5S4.67 4 5.5 4 7 4.67 7 5.5 6.33 7 5.5 7z"/></svg>
                </button>
                <button v-if="canManage || isGroupAdmin" class="btn-icon" @click.stop="openAccessDialog(folder)" :title="t('orgs.manageAccess')">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zm-6 9c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2zm3.1-9H8.9V6c0-1.71 1.39-3.1 3.1-3.1 1.71 0 3.1 1.39 3.1 3.1v2z"/></svg>
                </button>
                <button v-if="canWrite && renamingItem?.id !== folder.id" class="btn-icon" @click.stop="openMoveDialog(folder, 'folder')" :title="t('orgs.moveFolder')">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M20 6h-8l-2-2H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2zm-2 8h-3v3h-2v-3h-3v-2h3V9h2v3h3v2z"/></svg>
                </button>
                <button v-if="canWrite && renamingItem?.id !== folder.id" class="btn-icon" @click.stop="startRename(folder, 'folder')" :title="t('orgs.renameFolder')">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25zM20.71 7.04c.39-.39.39-1.02 0-1.41l-2.34-2.34a.9959.9959 0 0 0-1.41 0l-1.83 1.83 3.75 3.75 1.83-1.83z"/></svg>
                </button>
                <button class="btn-icon" @click.stop="handleFolderZipDownload(folder)" :title="t('orgs.downloadZip')" :disabled="!!zipDownloadStates[folder.id]">
                  <span v-if="zipDownloadStates[folder.id]" class="spinner-sm"></span>
                  <svg v-else viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M19 9h-4V3H9v6H5l7 7 7-7zM5 18v2h14v-2H5z"/></svg>
                </button>
                <button class="btn-icon" :class="{ 'btn-icon-pinned': isFavorite(folder.id, 'folder') }" @click.stop="toggleFavorite(folder, 'folder')" :title="isFavorite(folder.id, 'folder') ? t('orgs.unpinItem') : t('orgs.pinItem')">
                  <svg v-if="isFavorite(folder.id, 'folder')" viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M12 17.27L18.18 21l-1.64-7.03L22 9.24l-7.19-.61L12 2 9.19 8.63 2 9.24l5.46 4.73L5.82 21z"/></svg>
                  <svg v-else viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M22 9.24l-7.19-.62L12 2 9.19 8.63 2 9.24l5.46 4.73L5.82 21 12 17.27 18.18 21l-1.63-7.03L22 9.24zM12 15.4l-3.76 2.27 1-4.28-3.32-2.88 4.38-.38L12 6.1l1.71 4.04 4.38.38-3.32 2.88 1 4.28L12 15.4z"/></svg>
                </button>
                <button v-if="canWrite" class="btn-icon-danger" @click.stop="confirmDeleteFolder(folder)" :title="t('orgs.deleteFolder')">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"/></svg>
                </button>
              </div>
            </div>

            <!-- Files -->
            <div
              v-for="file in sortedFiles"
              :key="'file-' + file.id"
              class="item-row"
              :class="{ selected: isSelected('file', file.id), previewable: canPreview(file.mime_type) }"
              :draggable="canWrite"
              @dragstart="onItemDragStart($event, file, 'file')"
              @dragend="onItemDragEnd"
              @click="openPreview(file)"
            >
              <label class="checkbox-wrap" @click.stop>
                <input type="checkbox" :checked="isSelected('file', file.id)" @change="e => toggleSelect(e, 'file', file.id)" class="item-checkbox" />
              </label>
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor" class="item-icon file-icon"><path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z"/></svg>
              <input
                v-if="renamingItem?.id === file.id"
                :id="`rename-input-${file.id}`"
                class="item-rename-input"
                v-model="renamingItem.value"
                :placeholder="t('orgs.renamePlaceholder')"
                @keydown.enter.prevent="saveRename"
                @keydown.escape.prevent="cancelRename"
                @blur="saveRename"
                @click.stop
              />
              <span v-else class="item-name">{{ file.name }}</span>
              <div v-if="renamingItem?.id !== file.id && (file.tag_ids || []).length > 0" class="item-tags">
                <span
                  v-for="tagID in (file.tag_ids || []).slice(0, 3)"
                  :key="tagID"
                  class="item-tag-dot"
                  :style="{ background: tagColor(tagID) }"
                  :title="tagName(tagID)"
                ></span>
                <span v-if="(file.tag_ids || []).length > 3" class="item-tag-more">+{{ (file.tag_ids || []).length - 3 }}</span>
              </div>
              <span v-if="renamingItem?.id !== file.id" class="item-meta">{{ formatSize(file.size) }} · {{ formatDate(file.created_at) }}</span>
              <div class="item-actions">
                <button class="btn-icon" @click.stop="openTagPopover($event, file.id, 'file')" :title="t('orgs.addTag')">
                  <svg viewBox="0 0 24 24" width="13" height="13" fill="currentColor"><path d="M21.41 11.58l-9-9C12.05 2.22 11.55 2 11 2H4c-1.1 0-2 .9-2 2v7c0 .55.22 1.05.59 1.42l9 9c.36.36.86.58 1.41.58s1.05-.22 1.41-.59l7-7c.37-.36.59-.86.59-1.41s-.23-1.06-.59-1.42zM5.5 7C4.67 7 4 6.33 4 5.5S4.67 4 5.5 4 7 4.67 7 5.5 6.33 7 5.5 7z"/></svg>
                </button>
                <button class="btn-icon" @click.stop="handleDownload(file)" :title="t('file.download')">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M19 9h-4V3H9v6H5l7 7 7-7zM5 18v2h14v-2H5z"/></svg>
                </button>
                <button class="btn-icon" @click.stop="openShareModal(file)" :title="t('orgs.shareFile')">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M18 16.08c-.76 0-1.44.3-1.96.77L8.91 12.7c.05-.23.09-.46.09-.7s-.04-.47-.09-.7l7.05-4.11c.54.5 1.25.81 2.04.81 1.66 0 3-1.34 3-3s-1.34-3-3-3-3 1.34-3 3c0 .24.04.47.09.7L8.04 9.81C7.5 9.31 6.79 9 6 9c-1.66 0-3 1.34-3 3s1.34 3 3 3c.79 0 1.5-.31 2.04-.81l7.12 4.16c-.05.21-.08.43-.08.65 0 1.61 1.31 2.92 2.92 2.92 1.61 0 2.92-1.31 2.92-2.92s-1.31-2.92-2.92-2.92z"/></svg>
                </button>
                <button v-if="canWrite && renamingItem?.id !== file.id" class="btn-icon" @click.stop="openMoveDialog(file, 'file')" :title="t('orgs.moveFile')">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M20 6h-8l-2-2H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2zm-2 8h-3v3h-2v-3h-3v-2h3V9h2v3h3v2z"/></svg>
                </button>
                <button v-if="canWrite && renamingItem?.id !== file.id" class="btn-icon" @click.stop="startRename(file, 'file')" :title="t('orgs.renameFile')">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25zM20.71 7.04c.39-.39.39-1.02 0-1.41l-2.34-2.34a.9959.9959 0 0 0-1.41 0l-1.83 1.83 3.75 3.75 1.83-1.83z"/></svg>
                </button>
                <button class="btn-icon" :class="{ 'btn-icon-pinned': isFavorite(file.id, 'file') }" @click.stop="toggleFavorite(file, 'file')" :title="isFavorite(file.id, 'file') ? t('orgs.unpinItem') : t('orgs.pinItem')">
                  <svg v-if="isFavorite(file.id, 'file')" viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M12 17.27L18.18 21l-1.64-7.03L22 9.24l-7.19-.61L12 2 9.19 8.63 2 9.24l5.46 4.73L5.82 21z"/></svg>
                  <svg v-else viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M22 9.24l-7.19-.62L12 2 9.19 8.63 2 9.24l5.46 4.73L5.82 21 12 17.27 18.18 21l-1.63-7.03L22 9.24zM12 15.4l-3.76 2.27 1-4.28-3.32-2.88 4.38-.38L12 6.1l1.71 4.04 4.38.38-3.32 2.88 1 4.28L12 15.4z"/></svg>
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

        <!-- Provision-all banner — shown when ≥1 provisionable member exists -->
        <div v-if="canManage && membersNeedingKey.length > 0" class="provision-banner">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zm-6 9c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2zm3.1-9H8.9V6c0-1.71 1.39-3.1 3.1-3.1 1.71 0 3.1 1.39 3.1 3.1v2z"/></svg>
          <span>{{ t('orgs.nMembersNeedKey', { count: membersNeedingKey.length }) }}</span>
          <button
            class="btn-provision-all"
            @click="handleProvisionAll"
            :disabled="provisioningAll"
          >
            <span v-if="provisioningAll" class="spinner-sm"></span>
            <span v-else>{{ t('orgs.provisionAll') }}</span>
          </button>
        </div>

        <div v-if="orgStore.loading" class="loading-inline"><div class="spinner-sm-dark"></div></div>
        <div v-else class="members-list">
          <div v-for="m in sortedMembers" :key="m.user_id" class="member-row" :class="{ 'needs-key': canManage && !m.encrypted_org_key }">
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
              <span v-if="m.quota_bytes > 0" class="quota-badge" :title="t('orgs.memberQuotaLabel')">
                {{ formatBytes(m.quota_bytes) }}
              </span>
              <button
                v-if="canManage && m.role !== 'owner' && m.user_id !== myUserID"
                class="btn-sm"
                @click="openQuotaDialog(m)"
                :title="t('orgs.memberQuotaLabel')"
              >
                <svg viewBox="0 0 24 24" width="13" height="13" fill="currentColor"><path d="M19.14 12.94c.04-.3.06-.61.06-.94 0-.32-.02-.64-.07-.94l2.03-1.58a.49.49 0 0 0 .12-.61l-1.92-3.32a.49.49 0 0 0-.6-.22l-2.39.96a7.2 7.2 0 0 0-1.62-.94l-.36-2.54a.484.484 0 0 0-.48-.41h-3.84c-.24 0-.43.17-.47.41l-.36 2.54a7.37 7.37 0 0 0-1.62.94l-2.39-.96a.48.48 0 0 0-.6.22L2.74 8.87a.47.47 0 0 0 .12.61l2.03 1.58c-.05.3-.07.63-.07.94s.02.64.07.94l-2.03 1.58a.47.47 0 0 0-.12.61l1.92 3.32c.12.22.37.29.6.22l2.39-.96c.5.36 1.04.67 1.62.94l.36 2.54c.05.24.24.41.48.41h3.84c.24 0 .44-.17.47-.41l.36-2.54a7.37 7.37 0 0 0 1.62-.94l2.39.96c.23.09.48 0 .6-.22l1.92-3.32a.47.47 0 0 0-.12-.61l-2.03-1.58zM12 15.6c-1.98 0-3.6-1.62-3.6-3.6s1.62-3.6 3.6-3.6 3.6 1.62 3.6 3.6-1.62 3.6-3.6 3.6z"/></svg>
              </button>
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
                <span v-if="inv.email_notified" class="invite-email-badge">
                  <svg viewBox="0 0 24 24" width="11" height="11" fill="currentColor"><path d="M20 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V6c0-1.1-.9-2-2-2zm0 4l-8 5-8-5V6l8 5 8-5v2z"/></svg>
                  {{ t('orgs.inviteEmailSent') }}
                </span>
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
            <button v-if="canManage" class="btn-sm" @click="handleExportAudit" :title="t('orgs.exportAuditLogHint')">
              <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M19 9h-4V3H9v6H5l7 7 7-7zm-14 9v2h14v-2H5z"/></svg>
              {{ t('orgs.exportAuditLog') }}
            </button>
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

      <!-- TAB: DASHBOARD -->
      <div v-if="activeTab === 'dashboard'" class="tab-content">
        <div class="section-header">
          <h3>{{ t('orgs.dashboard') }}</h3>
          <button class="btn-sm" @click="refreshDashboard">{{ t('orgs.refresh') }}</button>
        </div>

        <div v-if="loadingStats" class="loading-center"><div class="spinner"></div></div>
        <template v-else-if="orgStore.orgStats">
          <!-- Alert: members without org key -->
          <div v-if="orgStore.orgStats.members_no_key > 0" class="dash-alert">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/></svg>
            {{ t('orgs.dashMembersNoKey', { count: orgStore.orgStats.members_no_key }) }}
            <button class="btn-link" @click="switchTab('members')">{{ t('orgs.dashGoProvision') }}</button>
          </div>

          <!-- KPI cards -->
          <div class="dash-kpis">
            <div class="dash-kpi">
              <div class="dash-kpi-value">{{ orgStore.members.length }}</div>
              <div class="dash-kpi-label">{{ t('orgs.dashMembers') }}</div>
            </div>
            <div class="dash-kpi">
              <div class="dash-kpi-value">{{ orgStore.orgStats.file_count }}</div>
              <div class="dash-kpi-label">{{ t('orgs.dashFiles') }}</div>
            </div>
            <div class="dash-kpi">
              <div class="dash-kpi-value">{{ orgStore.orgStats.folder_count }}</div>
              <div class="dash-kpi-label">{{ t('orgs.dashFolders') }}</div>
            </div>
            <div class="dash-kpi">
              <div class="dash-kpi-value">{{ orgStore.orgStats.activity_7d }}</div>
              <div class="dash-kpi-label">{{ t('orgs.dashActivity7d') }}</div>
            </div>
            <div class="dash-kpi">
              <div class="dash-kpi-value">{{ orgStore.invitations.length }}</div>
              <div class="dash-kpi-label">{{ t('orgs.dashActiveInvites') }}</div>
            </div>
          </div>

          <!-- Storage breakdown by member -->
          <div class="dash-section">
            <h4 class="dash-section-title">{{ t('orgs.dashStorageByMember') }}</h4>
            <div v-if="orgStore.orgStats.storage_by_member.length === 0" class="empty-tab">
              <p>{{ t('orgs.dashNoFiles') }}</p>
            </div>
            <div v-else class="dash-storage-list">
              <div
                v-for="stat in orgStore.orgStats.storage_by_member"
                :key="stat.user_id"
                class="dash-storage-row"
              >
                <div class="dash-storage-identity">
                  <div class="member-avatar small">{{ (stat.name || stat.user_id).charAt(0).toUpperCase() }}</div>
                  <div class="dash-storage-name">{{ stat.name || stat.user_id.slice(0, 8) }}</div>
                </div>
                <div class="dash-storage-bar-wrap">
                  <div
                    class="dash-storage-bar"
                    :style="{ width: memberStoragePercent(stat) + '%' }"
                  ></div>
                </div>
                <div class="dash-storage-meta">
                  <span>{{ formatSize(stat.storage_bytes) }}</span>
                  <span class="dash-file-count">{{ stat.file_count }} {{ t('orgs.dashFileCount') }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Recent audit activity -->
          <div class="dash-section" v-if="canManage">
            <h4 class="dash-section-title">{{ t('orgs.dashRecentActivity') }}</h4>
            <div v-if="orgStore.auditLog.length === 0" class="empty-tab">
              <p>{{ t('orgs.noAuditEvents') }}</p>
            </div>
            <div v-else class="audit-list">
              <div v-for="entry in orgStore.auditLog.slice(0, 8)" :key="entry.id" class="audit-row">
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
            <button class="btn-link dash-audit-link" @click="switchTab('audit')">{{ t('orgs.dashViewAllActivity') }} →</button>
          </div>
        </template>
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
          <div class="form-group">
            <label class="toggle-label">
              <input type="checkbox" v-model="settingsForm.requireMFA" class="toggle-checkbox" />
              <span>{{ t('orgs.requireMFA') }}</span>
            </label>
            <p class="hint-sm">{{ t('orgs.requireMFAHint') }}</p>
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

      <!-- TAB: ACTIVITY -->
      <div v-if="activeTab === 'activity'" class="tab-content">
        <div class="section-header">
          <h3>{{ t('orgs.activity') }}</h3>
          <button class="btn-sm" @click="loadActivity" :disabled="activityLoading">
            <span v-if="activityLoading" class="spinner-sm"></span>
            <span v-else>{{ t('orgs.refresh') }}</span>
          </button>
        </div>
        <div v-if="activityLoading && orgStore.orgActivity.length === 0" class="loading-center" style="padding:40px 0">
          <div class="spinner"></div>
        </div>
        <div v-else-if="orgStore.orgActivity.length === 0" class="empty-tab">
          <svg viewBox="0 0 24 24" width="40" height="40" fill="currentColor" style="opacity:.3"><path d="M13 3c-4.97 0-9 4.03-9 9H1l3.89 3.89.07.14L9 12H6c0-3.87 3.13-7 7-7s7 3.13 7 7-3.13 7-7 7c-1.93 0-3.68-.79-4.94-2.06l-1.42 1.42C8.27 19.99 10.51 21 13 21c4.97 0 9-4.03 9-9s-4.03-9-9-9zm-1 5v5l4.28 2.54.72-1.21-3.5-2.08V8H12z"/></svg>
          <p>{{ t('orgs.noActivity') }}</p>
        </div>
        <div v-else class="activity-feed">
          <div v-for="group in activityByDay" :key="group.day" class="activity-day-group">
            <div class="activity-day-label">{{ group.day }}</div>
            <div class="activity-entry" v-for="entry in group.entries" :key="entry.id">
              <div class="activity-icon">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor">
                  <path :d="getActivityIcon(entry.action)" />
                </svg>
              </div>
              <div class="activity-body">
                <span class="activity-actor">{{ actorDisplayName(entry.actor_id) }}</span>
                <span class="activity-desc">{{ activityDescription(entry) }}</span>
              </div>
              <span class="activity-time" :title="entry.created_at">{{ timeAgo(entry.created_at) }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- TAB: TRASH -->
      <div v-if="activeTab === 'trash'" class="tab-content">
        <div class="section-header">
          <h3>{{ t('orgs.trash') }}</h3>
          <div class="section-header-actions">
            <button class="btn-sm" @click="loadTrash" :disabled="trashLoading">
              <span v-if="trashLoading" class="spinner-sm"></span>
              <span v-else>{{ t('orgs.refresh') }}</span>
            </button>
            <button v-if="canManage && orgStore.trash.length > 0" class="btn-sm btn-danger-sm" @click="handleEmptyTrash">
              {{ t('orgs.emptyTrash') }}
            </button>
          </div>
        </div>
        <div v-if="trashLoading && orgStore.trash.length === 0" class="loading-center" style="padding:40px 0">
          <div class="spinner"></div>
        </div>
        <div v-else-if="orgStore.trash.length === 0" class="empty-tab">
          <svg viewBox="0 0 24 24" width="40" height="40" fill="currentColor" style="opacity:.3"><path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"/></svg>
          <p>{{ t('orgs.trashEmpty') }}</p>
        </div>
        <div v-else class="trash-list">
          <div v-for="item in orgStore.trash" :key="item.item_type + item.id" class="trash-row">
            <div class="trash-row-icon">
              <svg v-if="item.item_type === 'folder'" viewBox="0 0 24 24" width="18" height="18" fill="currentColor"><path d="M10 4H4c-1.11 0-2 .89-2 2L2 18c0 1.11.89 2 2 2h16c1.11 0 2-.89 2-2V8c0-1.11-.89-2-2-2h-8l-2-2z"/></svg>
              <svg v-else viewBox="0 0 24 24" width="18" height="18" fill="currentColor"><path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z"/></svg>
            </div>
            <div class="trash-row-info">
              <span class="trash-row-name">{{ item.name }}</span>
              <span class="trash-row-path">{{ item.path }}</span>
            </div>
            <div class="trash-row-meta">
              <span class="trash-row-date" :title="item.deleted_at">{{ t('orgs.deletedOn', { date: formatTrashDate(item.deleted_at) }) }}</span>
              <span v-if="item.deleted_by" class="trash-row-by">{{ t('orgs.deletedBy', { user: trashActorName(item.deleted_by) }) }}</span>
            </div>
            <div class="trash-row-actions">
              <button class="btn-sm" @click="handleRestore(item)" :title="t('orgs.restore')">{{ t('orgs.restore') }}</button>
              <button v-if="isAdminOrOwner" class="btn-sm btn-danger-sm" @click="handlePermanentDelete(item)" :title="t('orgs.permanentDelete')">{{ t('orgs.permanentDelete') }}</button>
            </div>
          </div>
        </div>
      </div>

      <!-- TAB: SHARES -->
      <div v-if="activeTab === 'shares'" class="tab-content">
        <div class="section-header">
          <h3>{{ t('orgs.shares') }}</h3>
          <button class="btn-sm" @click="loadShares" :disabled="sharesLoading">
            <span v-if="sharesLoading" class="spinner-sm"></span>
            <span v-else>{{ t('orgs.refresh') }}</span>
          </button>
        </div>
        <p class="shares-scope-hint" v-if="!canManage">{{ t('orgs.sharesMyOwn') }}</p>
        <div v-if="sharesLoading && orgStore.orgShares.length === 0" class="loading-center" style="padding:40px 0">
          <div class="spinner"></div>
        </div>
        <div v-else-if="orgStore.orgShares.length === 0" class="empty-tab">
          <svg viewBox="0 0 24 24" width="40" height="40" fill="currentColor" style="opacity:.3"><path d="M18 16.08c-.76 0-1.44.3-1.96.77L8.91 12.7c.05-.23.09-.46.09-.7s-.04-.47-.09-.7l7.05-4.11c.54.5 1.25.81 2.04.81 1.66 0 3-1.34 3-3s-1.34-3-3-3-3 1.34-3 3c0 .24.04.47.09.7L8.04 9.81C7.5 9.31 6.79 9 6 9c-1.66 0-3 1.34-3 3s1.34 3 3 3c.79 0 1.5-.31 2.04-.81l7.12 4.16c-.05.21-.08.43-.08.65 0 1.61 1.31 2.92 2.92 2.92 1.61 0 2.92-1.31 2.92-2.92s-1.31-2.92-2.92-2.92z"/></svg>
          <p>{{ t('orgs.noShares') }}</p>
        </div>
        <div v-else class="shares-list">
          <div v-for="share in orgStore.orgShares" :key="share.id" class="share-row">
            <div class="share-row-icon">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor"><path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z"/></svg>
            </div>
            <div class="share-row-info">
              <span class="share-row-name">{{ share._file_name || share.file_name }}</span>
              <span class="share-row-path">{{ share.file_path }}</span>
            </div>
            <div class="share-row-meta">
              <span class="share-row-creator">{{ t('orgs.sharedBy', { user: shareActorName(share.owner_id) }) }}</span>
              <span class="share-row-date">{{ t('orgs.sharedOn', { date: formatTrashDate(share.created_at) }) }}</span>
              <span v-if="share.expires_at" class="share-row-expiry" :class="{ 'share-expired': isExpired(share.expires_at) }">
                {{ isExpired(share.expires_at) ? t('orgs.shareExpired') : t('orgs.shareExpires', { date: formatTrashDate(share.expires_at) }) }}
              </span>
            </div>
            <div class="share-row-stats">
              <span class="share-views">
                <svg viewBox="0 0 24 24" width="13" height="13" fill="currentColor"><path d="M12 4.5C7 4.5 2.73 7.61 1 12c1.73 4.39 6 7.5 11 7.5s9.27-3.11 11-7.5c-1.73-4.39-6-7.5-11-7.5zM12 17c-2.76 0-5-2.24-5-5s2.24-5 5-5 5 2.24 5 5-2.24 5-5 5zm0-8c-1.66 0-3 1.34-3 3s1.34 3 3 3 3-1.34 3-3-1.34-3-3-3z"/></svg>
                {{ share.views }}
              </span>
              <span class="share-downloads" :title="t('orgs.shareDownloadCount', share.download_count ?? 0)">
                <svg viewBox="0 0 24 24" width="13" height="13" fill="currentColor"><path d="M19 9h-4V3H9v6H5l7 7 7-7zm-14 9v2h14v-2H5z"/></svg>
                {{ share.download_count ?? 0 }}
              </span>
              <span v-if="share.single_use" class="share-single-use">{{ t('orgs.shareSingleUse') }}</span>
            </div>
            <div class="share-row-actions">
              <button class="btn-sm btn-danger-sm" @click="handleRevokeShare(share)" :title="t('orgs.revokeShare')">
                {{ t('orgs.revokeShare') }}
              </button>
            </div>
          </div>
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
            <button class="btn-primary" @click="handleCreateFolder" :disabled="!newFolderName || folderCreating">{{ t('orgs.create') }}</button>
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

            <div class="invite-email-section">
              <label class="checkbox-label">
                <input type="checkbox" v-model="inviteForm.sendEmail" />
                <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M20 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V6c0-1.1-.9-2-2-2zm0 4l-8 5-8-5V6l8 5 8-5v2z"/></svg>
                {{ t('orgs.inviteEmailSection') }}
              </label>
              <template v-if="inviteForm.sendEmail">
                <div class="form-group" style="margin-bottom:0">
                  <input
                    v-model="inviteForm.recipientEmail"
                    class="input-field"
                    type="email"
                    :placeholder="t('orgs.inviteEmailPlaceholder')"
                  />
                </div>
                <div class="invite-lang-row">
                  <span class="invite-lang-label">{{ t('orgs.inviteEmailLang') }}</span>
                  <div class="lang-toggle">
                    <button
                      class="lang-btn"
                      :class="{ active: inviteForm.emailLang === 'fr' }"
                      @click="inviteForm.emailLang = 'fr'"
                    >🇫🇷 {{ t('orgs.inviteEmailLangFr') }}</button>
                    <button
                      class="lang-btn"
                      :class="{ active: inviteForm.emailLang === 'en' }"
                      @click="inviteForm.emailLang = 'en'"
                    >🇬🇧 {{ t('orgs.inviteEmailLangEn') }}</button>
                  </div>
                </div>
              </template>
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

    <!-- Member quota dialog -->
    <Transition name="modal">
      <div v-if="quotaDialogMember" class="modal-overlay" @click.self="quotaDialogMember = null">
        <div class="modal" style="max-width:380px">
          <div class="modal-header">
            <h3>{{ t('orgs.memberQuotaLabel') }}</h3>
            <button class="btn-close" @click="quotaDialogMember = null">
              <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
            </button>
          </div>
          <div class="modal-body">
            <p style="margin:0 0 12px;font-size:14px;color:var(--text-secondary)">
              {{ quotaDialogMember.name || quotaDialogMember.email }}
            </p>
            <input
              v-model.number="quotaDialogBytes"
              type="number"
              min="0"
              step="1073741824"
              class="input-field"
              :placeholder="t('orgs.memberQuotaLabel')"
            />
            <p class="hint-sm" style="margin-top:6px">0 = {{ t('orgs.memberQuotaLabel').split(',')[1] || 'unlimited' }}</p>
          </div>
          <div class="modal-footer">
            <button class="btn-secondary" @click="quotaDialogMember = null">{{ t('common.cancel') }}</button>
            <button class="btn-primary" @click="handleSaveQuota" :disabled="savingQuota">
              <span v-if="savingQuota" class="spinner-sm"></span>
              {{ t('common.save') }}
            </button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Clean audit modal -->
    <Transition name="modal">
      <!-- Upload conflict dialog (org files) -->
      <div v-if="orgStore.orgConflictState" class="modal-overlay" @click.self="orgStore.resolveOrgConflict('cancel')">
        <div class="modal" style="max-width:420px">
          <div class="modal-header">
            <h3>{{ t('orgs.uploadConflictTitle') }}</h3>
          </div>
          <div class="modal-body">
            <p style="font-size:14px;color:var(--text-secondary);margin:0 0 8px">{{ t('orgs.uploadConflictMsg', { name: orgStore.orgConflictState.fileName }) }}</p>
          </div>
          <div class="modal-footer" style="display:flex;gap:10px;justify-content:flex-end;padding:16px 20px">
            <button class="btn-secondary" @click="orgStore.resolveOrgConflict('cancel')">{{ t('orgs.uploadConflictCancel') }}</button>
            <button class="btn-primary" @click="orgStore.resolveOrgConflict('keepBoth')">{{ t('orgs.uploadConflictKeepBoth') }}</button>
          </div>
        </div>
      </div>

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

    <!-- File preview modal -->
    <Transition name="modal">
      <div v-if="previewFile" class="modal-overlay preview-overlay" @click.self="closePreview" @keydown.escape="closePreview">
        <div class="preview-modal">
          <div class="preview-header">
            <div class="preview-title">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor" class="preview-file-icon"><path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z"/></svg>
              <span class="preview-filename">{{ previewFile.name }}</span>
            </div>
            <div class="preview-header-actions">
              <button class="btn-sm" @click="handleDownload(previewFile)" :title="t('file.download')">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M19 9h-4V3H9v6H5l7 7 7-7zM5 18v2h14v-2H5z"/></svg>
                {{ t('file.download') }}
              </button>
              <button class="btn-close" @click="closePreview">
                <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
              </button>
            </div>
          </div>

          <div class="preview-body">
            <div v-if="previewLoading" class="preview-loading">
              <div class="spinner"></div>
              <p>{{ t('orgs.previewDecrypting') }}</p>
            </div>
            <template v-else-if="previewKind(previewFile.mime_type) === 'image'">
              <img v-if="previewUrl" :src="previewUrl" class="preview-image" :alt="previewFile.name" />
            </template>
            <template v-else-if="previewKind(previewFile.mime_type) === 'video'">
              <video v-if="previewUrl" :src="previewUrl" class="preview-video" controls />
            </template>
            <template v-else-if="previewKind(previewFile.mime_type) === 'audio'">
              <div class="preview-audio-wrap">
                <svg viewBox="0 0 24 24" width="64" height="64" fill="currentColor" style="opacity:0.3"><path d="M12 3v10.55c-.59-.34-1.27-.55-2-.55-2.21 0-4 1.79-4 4s1.79 4 4 4 4-1.79 4-4V7h4V3h-6z"/></svg>
                <audio v-if="previewUrl" :src="previewUrl" controls class="preview-audio" />
              </div>
            </template>
            <template v-else-if="previewKind(previewFile.mime_type) === 'pdf'">
              <iframe v-if="previewUrl" :src="previewUrl" class="preview-pdf" frameborder="0" />
            </template>
            <template v-else-if="previewKind(previewFile.mime_type) === 'text'">
              <pre v-if="previewText !== null" class="preview-text">{{ previewText }}</pre>
            </template>
            <div v-else class="preview-unsupported">
              <svg viewBox="0 0 24 24" width="48" height="48" fill="currentColor" style="opacity:0.25"><path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z"/></svg>
              <p>{{ t('orgs.previewUnsupported') }}</p>
            </div>
          </div>

          <div class="preview-footer">
            <span class="preview-meta">{{ previewFile.mime_type || t('orgs.unknownType') }}</span>
            <span class="preview-meta">{{ formatSize(previewFile.size) }}</span>
            <span class="preview-meta">{{ formatDate(previewFile.created_at) }}</span>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Tag popover -->
    <div v-if="tagPopoverID !== null" class="tag-popover"
      :style="tagPopoverStyle"
      @click.stop
    >
      <div v-if="orgStore.orgTags.length === 0" class="tag-popover-empty">{{ t('orgs.noTagsYet') }}</div>
      <label
        v-for="tag in orgStore.orgTags"
        :key="tag.id"
        class="tag-popover-item"
      >
        <input type="checkbox"
          :checked="tagPopoverCurrentIDs.includes(tag.id)"
          @change="toggleTagOnItem(tag.id)"
        />
        <span class="tag-popover-dot" :style="{ background: tag.color }"></span>
        <span class="tag-popover-name">{{ tag.name }}</span>
      </label>
      <div class="tag-popover-footer">
        <button class="tag-popover-manage" @click="showTagManager = true; closeTagPopover()">{{ t('orgs.manageTags') }}</button>
      </div>
    </div>
    <div v-if="tagPopoverID !== null" class="tag-popover-backdrop" @click="closeTagPopover"></div>

    <!-- Tag manager modal -->
    <Transition name="modal">
      <div v-if="showTagManager" class="modal-overlay" @click.self="showTagManager = false">
        <div class="modal modal-sm">
          <div class="modal-header">
            <h3>{{ t('orgs.manageTags') }}</h3>
            <button class="btn-close" @click="showTagManager = false">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
            </button>
          </div>
          <div class="modal-body">
            <!-- Existing tags -->
            <div v-for="tag in orgStore.orgTags" :key="tag.id" class="tag-mgr-row">
              <span class="tag-mgr-dot" :style="{ background: tag.color }"></span>
              <span v-if="editingTagID !== tag.id" class="tag-mgr-name" @dblclick="startEditTag(tag)">{{ tag.name }}</span>
              <input v-else class="tag-mgr-input" v-model="editTagName" @keydown.enter.prevent="saveEditTag(tag)" @keydown.escape.prevent="editingTagID = null" @blur="saveEditTag(tag)" autofocus />
              <div class="tag-mgr-colors">
                <button
                  v-for="col in TAG_COLORS"
                  :key="col"
                  class="tag-color-swatch"
                  :class="{ active: tag.color === col }"
                  :style="{ background: col }"
                  @click="recolorTag(tag, col)"
                ></button>
              </div>
              <button v-if="canManage" class="btn-icon-danger" @click="confirmDeleteTag(tag)" :title="t('orgs.deleteTag')">
                <svg viewBox="0 0 24 24" width="13" height="13" fill="currentColor"><path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"/></svg>
              </button>
            </div>
            <!-- Create new tag -->
            <div class="tag-mgr-create">
              <div class="tag-mgr-colors">
                <button
                  v-for="col in TAG_COLORS"
                  :key="col"
                  class="tag-color-swatch"
                  :class="{ active: newTagColor === col }"
                  :style="{ background: col }"
                  @click="newTagColor = col"
                ></button>
              </div>
              <input
                class="tag-mgr-input"
                v-model="newTagName"
                :placeholder="t('orgs.tagNamePlaceholder')"
                @keydown.enter.prevent="createTag"
              />
              <button class="btn-sm" @click="createTag" :disabled="!newTagName.trim() || tagMgrLoading">
                {{ t('orgs.createTag') }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Move file/folder modal -->
    <Transition name="modal">
      <div v-if="showMoveModal" class="modal-overlay" @click.self="closeMoveDialog">
        <div class="modal move-modal">
          <div class="modal-header">
            <h3>{{ t('orgs.moveTo') }}</h3>
            <button class="btn-close" @click="closeMoveDialog">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
            </button>
          </div>
          <div class="modal-body">
            <p class="move-item-label">{{ movingItem?.name }}</p>
            <input
              class="move-search-input"
              v-model="moveSearch"
              :placeholder="t('orgs.moveSearch')"
              autofocus
            />
            <div v-if="moveFetching" class="move-fetching">
              <div class="spinner-sm-dark"></div>
            </div>
            <div v-else class="move-folder-list">
              <button
                v-for="folder in filteredMoveDestinations"
                :key="folder.path"
                class="move-folder-item"
                :class="{
                  'is-selected': moveDestination?.path === folder.path,
                  'is-current': folder.path === movingItem?.currentPath
                }"
                @click="moveDestination = folder"
              >
                <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor" class="move-folder-icon"><path d="M10 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z"/></svg>
                <div class="move-folder-info">
                  <span class="move-folder-name">{{ folder.name }}</span>
                  <span class="move-folder-path">{{ displayFolderPath(folder) }}</span>
                </div>
                <span v-if="folder.path === movingItem?.currentPath" class="move-current-badge">{{ t('orgs.moveCurrentFolder') }}</span>
                <svg v-if="moveDestination?.path === folder.path" viewBox="0 0 24 24" width="14" height="14" fill="currentColor" class="move-check"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/></svg>
              </button>
              <p v-if="filteredMoveDestinations.length === 0" class="move-empty">{{ t('orgs.moveNoFolders') }}</p>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn-secondary" @click="closeMoveDialog">{{ t('orgs.cancel') }}</button>
            <button class="btn-primary" @click="confirmMove" :disabled="!moveDestination || moveLoading || moveDestination.path === movingItem?.currentPath">
              <span v-if="moveLoading" class="spinner-sm"></span>
              <span v-else>{{ t('orgs.moveHere') }}</span>
            </button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Share file modal -->
    <Transition name="modal">
      <div v-if="showShareModal" class="modal-overlay" @click.self="closeShareModal">
        <div class="modal">
          <div class="modal-header">
            <h3>{{ t('orgs.shareFileTitle') }}</h3>
            <button class="btn-close" @click="closeShareModal">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
            </button>
          </div>
          <div class="modal-body">
            <div v-if="shareLoading" class="share-loading">
              <span class="spinner"></span>
              <span>{{ t('orgs.shareGenerating') }}</span>
            </div>
            <div v-else-if="shareLink" class="share-result">
              <p class="share-hint">{{ t('orgs.shareHint') }}</p>
              <div class="share-link-row">
                <input class="share-link-input" readonly :value="shareLink" @focus="$event.target.select()" />
                <button class="btn-copy" @click="copyShareLink" :title="t('orgs.copyLink')">
                  <svg v-if="!shareCopied" viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M16 1H4c-1.1 0-2 .9-2 2v14h2V3h12V1zm3 4H8c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h11c1.1 0 2-.9 2-2V7c0-1.1-.9-2-2-2zm0 16H8V7h11v14z"/></svg>
                  <svg v-else viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/></svg>
                </button>
              </div>
              <p class="share-warning">{{ t('orgs.shareKeyWarning') }}</p>
            </div>
            <div v-else class="share-form">
              <div class="share-form-field">
                <label>{{ t('orgs.shareExpiry') }}</label>
                <input type="datetime-local" v-model="shareExpiresAt" class="share-form-input" />
              </div>
              <div class="share-form-field">
                <label>{{ t('orgs.sharePasswordLabel') }}</label>
                <input type="password" v-model="sharePassword" class="share-form-input" :placeholder="t('orgs.sharePasswordPlaceholder')" autocomplete="new-password" />
              </div>
              <label class="share-single-use-label">
                <input type="checkbox" v-model="shareSingleUse" />
                <span>{{ t('orgs.shareSingleUseLabel') }}</span>
              </label>
              <p class="share-single-use-hint">{{ t('orgs.shareSingleUseHint') }}</p>
              <p v-if="shareError" class="share-error-msg">{{ shareError }}</p>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn-secondary" @click="closeShareModal">{{ t('orgs.close') }}</button>
            <button v-if="!shareLink" class="btn-primary" @click="createOrgShare" :disabled="shareLoading">
              {{ t('orgs.createShareLink') }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch, h, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useOrgStore } from '../stores/organizations'
import { useAuthStore } from '../stores/auth'
import { useRealtimeStore } from '../stores/realtime'
import OrgGroupsPanel from '../components/organizations/OrgGroupsPanel.vue'
import OrgFolderAccessDialog from '../components/organizations/OrgFolderAccessDialog.vue'
import OrgOnboardingWizard from '../components/organizations/OrgOnboardingWizard.vue'
import { generateOrgKey, decryptOrgKey, unwrapFileKey, wrapFileKey } from '../utils/orgCrypto.js'
import api from '../api'

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
    { key: 'dashboard', label: t('orgs.dashboard'), icon: TabIcon(['M3 13h8V3H3v10zm0 8h8v-6H3v6zm10 0h8V11h-8v10zm0-18v6h8V3h-8z']) },
    { key: 'activity', label: t('orgs.activity'), icon: TabIcon(['M13 3c-4.97 0-9 4.03-9 9H1l3.89 3.89.07.14L9 12H6c0-3.87 3.13-7 7-7s7 3.13 7 7-3.13 7-7 7c-1.93 0-3.68-.79-4.94-2.06l-1.42 1.42C8.27 19.99 10.51 21 13 21c4.97 0 9-4.03 9-9s-4.03-9-9-9zm-1 5v5l4.28 2.54.72-1.21-3.5-2.08V8H12z']) },
    { key: 'trash', label: t('orgs.trash'), icon: TabIcon(['M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z']), count: orgStore.trash.length || null },
    { key: 'shares', label: t('orgs.shares'), icon: TabIcon(['M18 16.08c-.76 0-1.44.3-1.96.77L8.91 12.7c.05-.23.09-.46.09-.7s-.04-.47-.09-.7l7.05-4.11c.54.5 1.25.81 2.04.81 1.66 0 3-1.34 3-3s-1.34-3-3-3-3 1.34-3 3c0 .24.04.47.09.7L8.04 9.81C7.5 9.31 6.79 9 6 9c-1.66 0-3 1.34-3 3s1.34 3 3 3c.79 0 1.5-.31 2.04-.81l7.12 4.16c-.05.21-.08.43-.08.65 0 1.61 1.31 2.92 2.92 2.92 1.61 0 2.92-1.31 2.92-2.92s-1.31-2.92-2.92-2.92z']), count: orgStore.orgShares.length || null },
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
const orgMFARequired = computed(() =>
  orgStore.currentOrg?.require_mfa === true && !authStore.user?.mfa_enabled
)

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

// ── Dashboard ─────────────────────────────────────────────────────────────────

const loadingStats = ref(false)

const loadDashboard = async () => {
  loadingStats.value = true
  try {
    await Promise.all([
      orgStore.fetchOrgStats(orgID.value),
      canManage.value && orgStore.auditLog.length === 0
        ? orgStore.fetchAuditLog(orgID.value, 1)
        : Promise.resolve(),
      orgStore.invitations.length === 0
        ? orgStore.fetchInvitations(orgID.value)
        : Promise.resolve(),
    ])
  } finally {
    loadingStats.value = false
  }
}

const refreshDashboard = () => loadDashboard()

const maxMemberStorage = computed(() => {
  const s = orgStore.orgStats?.storage_by_member
  if (!s || s.length === 0) return 1
  return Math.max(...s.map(m => m.storage_bytes), 1)
})

const memberStoragePercent = (stat) => {
  return Math.round((stat.storage_bytes / maxMemberStorage.value) * 100)
}

// ── Init ──────────────────────────────────────────────────────────────────────

let _unsubOrgUpdate = null

const showOnboardingWizard = ref(false)

onMounted(async () => {
  document.addEventListener('keydown', _onKeydown)
  await orgStore.fetchOrg(orgID.value)
  await Promise.all([
    orgStore.fetchItems(orgID.value, '/'),
    orgStore.fetchMembers(orgID.value),
    orgStore.fetchOrgTags(orgID.value).catch(() => {}),
    orgStore.fetchFavorites(orgID.value).catch(() => {}),
  ])

  settingsForm.value = {
    name: orgStore.currentOrg.name,
    description: orgStore.currentOrg.description,
    storageQuotaMB: orgStore.currentOrg.storage_quota_mb,
    requireMFA: orgStore.currentOrg.require_mfa ?? false,
  }

  const pendingOrgId = localStorage.getItem('kagibi_org_onboarding')
  if (pendingOrgId && parseInt(pendingOrgId) === orgID.value) {
    localStorage.removeItem('kagibi_org_onboarding')
    showOnboardingWizard.value = true
  }

  // Refresh members list and notify admin when someone joins or leaves this org
  _unsubOrgUpdate = realtimeStore.onEvent('org_update', async (payload) => {
    if (payload?.org_id !== orgID.value) return
    await orgStore.fetchMembers(orgID.value)

    if (payload?.action === 'member_joined' && canManage.value) {
      const newMember = orgStore.members.find(m => m.user_id === payload.user_id)
      const displayName = newMember?.name || payload.user_id?.slice(0, 8) || '?'
      if (newMember && !newMember.encrypted_org_key && newMember.public_key) {
        showToast(t('orgs.memberJoinedNeedsKey', { name: displayName }), 'info')
      } else {
        showToast(t('orgs.memberJoined', { name: displayName }))
      }
    }
  })
})

const _onKeydown = (e) => {
  if (e.key === 'Escape') {
    if (previewFile.value) { closePreview(); return }
    if (hasSelection.value && !showMoveModal.value) clearSelection()
  }
}

onUnmounted(() => {
  if (_unsubOrgUpdate) _unsubOrgUpdate()
  document.removeEventListener('keydown', _onKeydown)
  stopActivityRefresh()
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
  if (tab === 'dashboard') await loadDashboard()
  if (tab === 'activity') { loadActivity(); startActivityRefresh() }
  else stopActivityRefresh()
  if (tab === 'trash') loadTrash()
  if (tab === 'shares') loadShares()
}

// ── Activity feed ─────────────────────────────────────────────────────────────

const activityLoading = ref(false)
let _activityRefreshTimer = null

function timeAgo(dateStr) {
  const diff = (Date.now() - new Date(dateStr).getTime()) / 1000
  if (diff < 60) return t('orgs.justNow')
  if (diff < 3600) return t('orgs.minutesAgo', { n: Math.floor(diff / 60) })
  if (diff < 86400) return t('orgs.hoursAgo', { n: Math.floor(diff / 3600) })
  return t('orgs.daysAgo', { n: Math.floor(diff / 86400) })
}

function actorDisplayName(actorID) {
  const m = orgStore.members.find(m => m.user_id === actorID)
  return m?.name || actorID?.slice(0, 8) || '?'
}

function activityDescription(entry) {
  const key = `orgs.act_${entry.action}`
  const detail = entry.detail_plain || entry.detail || ''
  return t(key, { detail })
}

function getActivityIcon(action) {
  if (action === 'file_uploaded') return 'M19 9h-4V3H9v6H5l7 7 7-7zm-8 2V5h2v6h1.17L12 13.17 9.83 11H11zm-6 8v2h14v-2H5z'
  if (action === 'file_deleted' || action === 'folder_deleted') return 'M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z'
  if (action === 'file_downloaded') return 'M19 9h-4V3H9v6H5l7 7 7-7zm-8 2V5h2v6h1.17L12 13.17 9.83 11H11zm-6 8v2h14v-2H5z'
  if (action === 'file_renamed' || action === 'folder_renamed') return 'M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25zM20.71 7.04c.39-.39.39-1.02 0-1.41l-2.34-2.34c-.39-.39-1.02-.39-1.41 0l-1.83 1.83 3.75 3.75 1.83-1.83z'
  if (action === 'file_moved' || action === 'folder_moved') return 'M20 6h-2.18c.07-.44.18-.88.18-1.3C18 2.55 16.15 1 14 1h-.37C12.14.41 10.73 0 9 0 5.13 0 2 3.13 2 7v3c0 1.7.55 3.26 1.47 4.53L1 17l1.41 1.41 2.08-2.08c.71.43 1.49.72 2.32.86L7 21h10l2-10h1c1.1 0 2-.9 2-2v-1c0-1.1-.9-2-2-2z'
  if (action === 'file_shared_public') return 'M18 16.08c-.76 0-1.44.3-1.96.77L8.91 12.7c.05-.23.09-.46.09-.7s-.04-.47-.09-.7l7.05-4.11c.54.5 1.25.81 2.04.81 1.66 0 3-1.34 3-3s-1.34-3-3-3-3 1.34-3 3c0 .24.04.47.09.7L8.04 9.81C7.5 9.31 6.79 9 6 9c-1.66 0-3 1.34-3 3s1.34 3 3 3c.79 0 1.5-.31 2.04-.81l7.12 4.16c-.05.21-.08.43-.08.65 0 1.61 1.31 2.92 2.92 2.92 1.61 0 2.92-1.31 2.92-2.92s-1.31-2.92-2.92-2.92z'
  if (action === 'folder_created') return 'M10 4H4c-1.11 0-2 .89-2 2L2 18c0 1.11.89 2 2 2h16c1.11 0 2-.89 2-2V8c0-1.11-.89-2-2-2h-8l-2-2z'
  if (action.startsWith('member_')) return 'M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z'
  if (action.startsWith('invitation_')) return 'M20 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V6c0-1.1-.9-2-2-2zm0 4l-8 5-8-5V6l8 5 8-5v2z'
  if (action.startsWith('permission_')) return 'M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zm-6 9c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2zm3.1-9H8.9V6c0-1.71 1.39-3.1 3.1-3.1 1.71 0 3.1 1.39 3.1 3.1v2z'
  if (action.startsWith('group_')) return 'M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z'
  return 'M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z'
}

const activityByDay = computed(() => {
  const days = []
  let lastDay = null
  for (const entry of orgStore.orgActivity) {
    const day = entry.created_at ? entry.created_at.slice(0, 10) : '?'
    if (day !== lastDay) { days.push({ day, entries: [] }); lastDay = day }
    days[days.length - 1].entries.push(entry)
  }
  return days
})

async function loadActivity() {
  activityLoading.value = true
  try { await orgStore.fetchOrgActivity(orgID.value) } catch (_) {}
  activityLoading.value = false
}

function startActivityRefresh() {
  stopActivityRefresh()
  _activityRefreshTimer = setInterval(() => {
    if (activeTab.value === 'activity') loadActivity()
  }, 30000)
}

function stopActivityRefresh() {
  if (_activityRefreshTimer) { clearInterval(_activityRefreshTimer); _activityRefreshTimer = null }
}

// ── Favorites ────────────────────────────────────────────────────────────────

const enrichedFavorites = computed(() => orgStore.favorites)

function isFavorite(itemID, itemType) {
  return orgStore.favorites.some(f => f.item_id === itemID && f.item_type === itemType)
}

async function toggleFavorite(item, type) {
  const id = item.id
  if (isFavorite(id, type)) {
    await orgStore.removeFavorite(orgID.value, id, type)
    showToast(t('orgs.itemUnpinned'))
  } else {
    await orgStore.addFavorite(orgID.value, id, type)
    showToast(t('orgs.itemPinned'))
  }
}

function onFavClick(fav) {
  if (fav.item_type === 'folder') {
    navigateToPath(fav._path || '/')
  } else {
    navigateToPath(fav._parent_path || '/')
  }
}

// ── Trash ────────────────────────────────────────────────────────────────────

const trashLoading = ref(false)

async function loadTrash() {
  trashLoading.value = true
  try { await orgStore.fetchTrash(orgID.value) } catch (_) {}
  trashLoading.value = false
}

function formatTrashDate(dateStr) {
  if (!dateStr) return '?'
  const d = new Date(dateStr)
  return d.toLocaleDateString()
}

function trashActorName(actorID) {
  const m = orgStore.members.find(m => m.user_id === actorID)
  return m?.name || actorID?.slice(0, 8) || actorID
}

async function handleRestore(item) {
  try {
    await orgStore.restoreTrashItem(orgID.value, item.item_type, item.id)
    showToast(t('orgs.itemRestored'))
  } catch (_) {}
}

async function handlePermanentDelete(item) {
  if (!confirm(t('orgs.confirmPermanentDelete', { name: item.name }))) return
  try {
    await orgStore.permanentDeleteTrashItem(orgID.value, item.item_type, item.id)
    showToast(t('orgs.permanentDeleted'))
  } catch (_) {}
}

async function handleEmptyTrash() {
  if (!confirm(t('orgs.confirmEmptyTrash'))) return
  try {
    await orgStore.emptyTrash(orgID.value)
    showToast(t('orgs.trashEmptied'))
  } catch (_) {}
}

// ── Shares ───────────────────────────────────────────────────────────────────

const sharesLoading = ref(false)

async function loadShares() {
  sharesLoading.value = true
  try { await orgStore.fetchOrgShares(orgID.value) } catch (_) {}
  sharesLoading.value = false
}

function shareActorName(actorID) {
  const m = orgStore.members.find(m => m.user_id === actorID)
  return m?.name || actorID?.slice(0, 8) || actorID
}

function isExpired(dateStr) {
  if (!dateStr) return false
  return new Date(dateStr) < new Date()
}

async function handleRevokeShare(share) {
  if (!confirm(t('orgs.confirmRevokeShare', { name: share._file_name || share.file_name }))) return
  try {
    await orgStore.revokeOrgShare(orgID.value, share.id)
    showToast(t('orgs.shareRevoked'))
  } catch (_) {}
}

// ── Search ────────────────────────────────────────────────────────────────────

const searchInputRef = ref(null)
const searchQuery = ref('')
const searchResults = ref([])
const searchLoading = ref(false)
let _searchDebounce = null

const onSearchInput = () => {
  clearTimeout(_searchDebounce)
  if (!searchQuery.value.trim()) {
    searchResults.value = []
    return
  }
  searchLoading.value = true
  _searchDebounce = setTimeout(async () => {
    try {
      searchResults.value = await orgStore.searchOrgItems(orgID.value, searchQuery.value)
    } catch (e) {
      searchResults.value = []
    } finally {
      searchLoading.value = false
    }
  }, 280)
}

const clearSearch = () => {
  searchQuery.value = ''
  searchResults.value = []
  searchLoading.value = false
}

const navigateToSearchResult = (item) => {
  clearSearch()
  if (item.type === 'folder') {
    navigateToPath(item.path)
  } else {
    navigateToPath(item.parent_path)
  }
}

const highlightMatch = (text, query) => {
  if (!query) return text
  const idx = text.toLowerCase().indexOf(query.toLowerCase())
  if (idx === -1) return text
  return (
    text.slice(0, idx) +
    '<mark class="search-highlight">' +
    text.slice(idx, idx + query.length) +
    '</mark>' +
    text.slice(idx + query.length)
  )
}

// ── File system ───────────────────────────────────────────────────────────────

const navigateToPath = async (path) => {
  const prevPath = currentPath.value
  currentPath.value = path || '/'
  activeFolderID.value = null
  clearSelection()
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
const folderCreating = ref(false)

const handleCreateFolder = async () => {
  if (!newFolderName.value || folderCreating.value) return
  folderError.value = ''
  folderCreating.value = true
  try {
    await orgStore.createFolder(orgID.value, newFolderName.value, currentPath.value)
    showNewFolderModal.value = false
    newFolderName.value = ''
    showToast(t('orgs.folderCreated'))
  } catch (e) {
    folderError.value = e.response?.data?.error || e.message
  } finally {
    folderCreating.value = false
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

// ── Public file share ─────────────────────────────────────────────────────────

const showShareModal = ref(false)
const shareLoading = ref(false)
const shareLink = ref('')
const shareError = ref('')
const shareCopied = ref(false)
const shareSelectedFile = ref(null)
// Share creation form fields
const shareExpiresAt = ref('')
const sharePassword = ref('')
const shareSingleUse = ref(false)

const openShareModal = (file) => {
  shareSelectedFile.value = file
  showShareModal.value = true
  shareLink.value = ''
  shareError.value = ''
  shareCopied.value = false
  shareExpiresAt.value = ''
  sharePassword.value = ''
  shareSingleUse.value = false
}

const createOrgShare = async () => {
  const file = shareSelectedFile.value
  if (!file) return
  shareLoading.value = true
  shareError.value = ''
  try {
    const { data: keyData } = await api.get(`/orgs/${orgID.value}/fs/file/${file.id}/key`)
    const encryptedFileKey = keyData.encrypted_key

    if (!authStore.privateKey) throw new Error('Clé privée introuvable. Reconnectez-vous.')
    const orgKey = await decryptOrgKey(
      orgStore.currentOrg?.my_encrypted_org_key,
      authStore.privateKey,
    )

    const fileKey = await unwrapFileKey(encryptedFileKey, orgKey)
    const shareKey = await generateOrgKey()
    const encryptedKeyForShare = await wrapFileKey(fileKey, shareKey)

    const shareKeyRaw = await crypto.subtle.exportKey('raw', shareKey)
    const shareKeyB64 = btoa(String.fromCharCode(...new Uint8Array(shareKeyRaw)))
      .replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '')

    let expiresAt = null
    if (shareExpiresAt.value) {
      const d = new Date(shareExpiresAt.value)
      if (d <= new Date()) throw new Error("La date d'expiration doit être dans le futur.")
      expiresAt = d.toISOString()
    }

    const result = await orgStore.createOrgFileShare(orgID.value, file.id, {
      encryptedKey: encryptedKeyForShare,
      expiresAt,
      password: sharePassword.value,
      singleUse: shareSingleUse.value,
    })

    shareLink.value = `${window.location.origin}/s/org/${result.token}#${shareKeyB64}`
    // Refresh shares list silently
    orgStore.fetchOrgShares(orgID.value).catch(() => {})
  } catch (e) {
    shareError.value = e.response?.data?.error || e.message
  } finally {
    shareLoading.value = false
  }
}

const closeShareModal = () => {
  showShareModal.value = false
  shareLink.value = ''
  shareError.value = ''
  shareCopied.value = false
  shareSelectedFile.value = null
}

const copyShareLink = async () => {
  try {
    await navigator.clipboard.writeText(shareLink.value)
    shareCopied.value = true
    setTimeout(() => { shareCopied.value = false }, 2000)
  } catch {
    // fallback
    const el = document.createElement('textarea')
    el.value = shareLink.value
    document.body.appendChild(el)
    el.select()
    document.execCommand('copy')
    document.body.removeChild(el)
    shareCopied.value = true
    setTimeout(() => { shareCopied.value = false }, 2000)
  }
}

// ── Inline rename ─────────────────────────────────────────────────────────────

const renamingItem = ref(null)  // { id, type: 'file'|'folder', value: string }
const renameLoading = ref(false)

function startRename(item, type) {
  renamingItem.value = { id: item.id, type, value: item.name }
  nextTick(() => {
    const el = document.getElementById(`rename-input-${item.id}`)
    if (el) { el.focus(); el.select() }
  })
}

function cancelRename() {
  renamingItem.value = null
}

async function saveRename() {
  if (!renamingItem.value) return
  const { id, type, value } = renamingItem.value
  const trimmed = value.trim()
  renamingItem.value = null
  if (!trimmed) return

  const items = orgStore.currentItems
  const existing = type === 'folder'
    ? items.folders.find(f => f.id === id)
    : items.files.find(f => f.id === id)
  if (!existing || existing.name === trimmed) return

  renameLoading.value = true
  try {
    if (type === 'folder') {
      await orgStore.renameOrgFolder(orgID.value, id, trimmed)
    } else {
      await orgStore.renameOrgFile(orgID.value, id, trimmed)
    }
    showToast(t('orgs.renameSuccess'))
  } catch (e) {
    showToast(e.response?.data?.error || e.message, 'error')
  } finally {
    renameLoading.value = false
  }
}

// ── Move dialog ───────────────────────────────────────────────────────────────

const showMoveModal = ref(false)
const movingItem = ref(null)  // { id, name, type, currentPath }
const moveDestination = ref(null)  // { id, name, path }
const moveSearch = ref('')
const allOrgFolders = ref([])
const moveLoading = ref(false)
const moveFetching = ref(false)

const filteredMoveDestinations = computed(() => {
  const root = { id: 0, name: '/ ' + t('orgs.moveRoot'), path: '/' }
  const all = [root, ...allOrgFolders.value]
  const q = moveSearch.value.trim().toLowerCase()
  if (!q) return all
  return all.filter(f =>
    f.name.toLowerCase().includes(q) ||
    displayFolderPath(f).toLowerCase().includes(q)
  )
})

function displayFolderPath(folder) {
  if (folder.path === '/') return '/'
  return folder.path.split('/').filter(s => s).map(seg =>
    orgStore.folderNameCache[seg] || seg.slice(0, 6) + '…'
  ).join(' / ')
}

async function openMoveDialog(item, type) {
  movingItem.value = {
    id: item.id,
    name: item.name,
    type,
    currentPath: type === 'file' ? item.folder_path : item.parent_path,
  }
  moveDestination.value = null
  moveSearch.value = ''
  showMoveModal.value = true
  moveFetching.value = true
  try {
    allOrgFolders.value = await orgStore.getAllOrgFolders(orgID.value)
  } catch (e) {
    showToast(e.response?.data?.error || e.message, 'error')
    showMoveModal.value = false
  } finally {
    moveFetching.value = false
  }
}

function closeMoveDialog() {
  showMoveModal.value = false
  movingItem.value = null
  moveDestination.value = null
  bulkMoveMode.value = false
}

async function confirmMove() {
  if (!moveDestination.value) return
  moveLoading.value = true
  try {
    if (bulkMoveMode.value) {
      let failed = 0
      for (const folder of selectedFolderItems.value) {
        try { await orgStore.moveOrgFolder(orgID.value, folder.id, moveDestination.value.path) }
        catch { failed++ }
      }
      for (const file of selectedFileItems.value) {
        try { await orgStore.moveOrgFile(orgID.value, file.id, moveDestination.value.path) }
        catch { failed++ }
      }
      const done = selectedCount.value - failed
      clearSelection()
      if (failed > 0) showToast(t('orgs.bulkMovePartial', { done, failed }), 'error')
      else showToast(t('orgs.bulkMoveSuccess', { count: done }))
    } else {
      if (!movingItem.value) return
      if (movingItem.value.type === 'file') {
        await orgStore.moveOrgFile(orgID.value, movingItem.value.id, moveDestination.value.path)
      } else {
        await orgStore.moveOrgFolder(orgID.value, movingItem.value.id, moveDestination.value.path)
      }
      showToast(t('orgs.moveSuccess'))
    }
    closeMoveDialog()
  } catch (e) {
    showToast(e.response?.data?.error || e.message, 'error')
  } finally {
    moveLoading.value = false
  }
}

// ── File preview ──────────────────────────────────────────────────────────────

const previewFile = ref(null)
const previewUrl = ref(null)
const previewText = ref(null)
const previewLoading = ref(false)
const TEXT_PREVIEW_LIMIT = 1024 * 512  // 512 KB

function canPreview(mimeType) {
  if (!mimeType) return false
  return mimeType.startsWith('image/') ||
    mimeType.startsWith('video/') ||
    mimeType.startsWith('audio/') ||
    mimeType.startsWith('text/') ||
    mimeType === 'application/pdf'
}

function previewKind(mimeType) {
  if (!mimeType) return 'none'
  if (mimeType.startsWith('image/')) return 'image'
  if (mimeType.startsWith('video/')) return 'video'
  if (mimeType.startsWith('audio/')) return 'audio'
  if (mimeType === 'application/pdf') return 'pdf'
  if (mimeType.startsWith('text/')) return 'text'
  return 'none'
}

async function openPreview(file) {
  if (renamingItem.value) return
  previewFile.value = file
  previewUrl.value = null
  previewText.value = null
  if (!canPreview(file.mime_type)) return
  previewLoading.value = true
  try {
    const blob = await orgStore.getFileBlob(orgID.value, file.id, file.mime_type)
    if (previewKind(file.mime_type) === 'text') {
      const slice = blob.size > TEXT_PREVIEW_LIMIT ? blob.slice(0, TEXT_PREVIEW_LIMIT) : blob
      previewText.value = await slice.text()
      if (blob.size > TEXT_PREVIEW_LIMIT) previewText.value += '\n\n[… truncated]'
    } else {
      previewUrl.value = URL.createObjectURL(blob)
    }
  } catch (e) {
    showToast(e.response?.data?.error || e.message, 'error')
    previewFile.value = null
  } finally {
    previewLoading.value = false
  }
}

function closePreview() {
  if (previewUrl.value) URL.revokeObjectURL(previewUrl.value)
  previewFile.value = null
  previewUrl.value = null
  previewText.value = null
  previewLoading.value = false
}

// ── Tags ──────────────────────────────────────────────────────────────────────

const TAG_COLORS = ['#ef4444','#f97316','#eab308','#22c55e','#06b6d4','#6366f1','#a855f7','#ec4899']

const filterTagID = ref(null)
const showTagManager = ref(false)
const newTagName = ref('')
const newTagColor = ref('#6366f1')
const tagMgrLoading = ref(false)
const editingTagID = ref(null)
const editTagName = ref('')

// Tag popover state
const tagPopoverID = ref(null)       // item id (file or folder)
const tagPopoverType = ref(null)     // 'file' | 'folder'
const tagPopoverStyle = ref({})
const tagPopoverCurrentIDs = ref([]) // tag_ids currently on the item

function tagColor(tagID) {
  return orgStore.orgTags.find(t => t.id === tagID)?.color || '#888'
}
function tagName(tagID) {
  return orgStore.orgTags.find(t => t.id === tagID)?.name || '?'
}

function openTagPopover(e, itemID, type) {
  e.stopPropagation()
  const btn = e.currentTarget
  const rect = btn.getBoundingClientRect()
  tagPopoverStyle.value = {
    top: (rect.bottom + window.scrollY + 4) + 'px',
    left: (rect.left + window.scrollX) + 'px',
  }
  tagPopoverID.value = itemID
  tagPopoverType.value = type
  const items = type === 'file' ? orgStore.currentItems.files : orgStore.currentItems.folders
  const item = items.find(i => i.id === itemID)
  tagPopoverCurrentIDs.value = [...(item?.tag_ids || [])]
}

function closeTagPopover() {
  tagPopoverID.value = null
  tagPopoverType.value = null
  tagPopoverCurrentIDs.value = []
}

async function toggleTagOnItem(tagID) {
  const ids = [...tagPopoverCurrentIDs.value]
  const idx = ids.indexOf(tagID)
  if (idx >= 0) ids.splice(idx, 1)
  else ids.push(tagID)
  tagPopoverCurrentIDs.value = ids
  try {
    if (tagPopoverType.value === 'file') await orgStore.setFileTags(orgID.value, tagPopoverID.value, ids)
    else await orgStore.setFolderTags(orgID.value, tagPopoverID.value, ids)
  } catch (e) {
    showToast(e.response?.data?.error || e.message, 'error')
  }
}

async function createTag() {
  const name = newTagName.value.trim()
  if (!name) return
  tagMgrLoading.value = true
  try {
    await orgStore.createOrgTag(orgID.value, name, newTagColor.value)
    newTagName.value = ''
    newTagColor.value = '#6366f1'
    showToast(t('orgs.tagCreated'))
  } catch (e) {
    showToast(e.response?.data?.error || e.message, 'error')
  } finally {
    tagMgrLoading.value = false
  }
}

function startEditTag(tag) {
  editingTagID.value = tag.id
  editTagName.value = tag.name
}

async function saveEditTag(tag) {
  const name = editTagName.value.trim()
  editingTagID.value = null
  if (!name || name === tag.name) return
  try {
    await orgStore.updateOrgTag(orgID.value, tag.id, name, null)
  } catch (e) {
    showToast(e.response?.data?.error || e.message, 'error')
  }
}

async function recolorTag(tag, color) {
  if (tag.color === color) return
  try {
    await orgStore.updateOrgTag(orgID.value, tag.id, null, color)
  } catch (e) {
    showToast(e.response?.data?.error || e.message, 'error')
  }
}

async function confirmDeleteTag(tag) {
  if (!confirm(t('orgs.deleteTagConfirm', { name: tag.name }))) return
  try {
    await orgStore.deleteOrgTag(orgID.value, tag.id)
    if (filterTagID.value === tag.id) filterTagID.value = null
    showToast(t('orgs.tagDeleted'))
  } catch (e) {
    showToast(e.response?.data?.error || e.message, 'error')
  }
}

// ── Drag & drop upload ────────────────────────────────────────────────────────

const isDragOver = ref(false)
let _dragDepth = 0

function onDragEnterZone(e) {
  if (!e.dataTransfer.types.includes('Files')) return
  _dragDepth++
  isDragOver.value = true
}

function onDragLeaveZone(e) {
  if (!e.dataTransfer.types.includes('Files')) return
  _dragDepth--
  if (_dragDepth <= 0) { _dragDepth = 0; isDragOver.value = false }
}

function onDragOverZone(e) {
  if (e.dataTransfer.types.includes('Files')) e.dataTransfer.dropEffect = 'copy'
}

function onDropFiles(e) {
  _dragDepth = 0
  isDragOver.value = false
  if (!canWrite.value) return
  const files = Array.from(e.dataTransfer.files)
  files.forEach(f => uploadFile(f))
}

// ── Drag & drop move ──────────────────────────────────────────────────────────

let _dragItem = null  // { id, type, path, currentPath }
const dragOverFolderID = ref(null)
const dragOverBcPath = ref(null)

function onItemDragStart(e, item, type) {
  _dragItem = {
    id: item.id,
    type,
    path: type === 'folder' ? item.path : null,
    currentPath: type === 'file' ? item.folder_path : item.parent_path,
  }
  e.dataTransfer.effectAllowed = 'move'
  e.dataTransfer.setData('text/plain', `${type}-${item.id}`)
}

function onItemDragEnd() {
  _dragItem = null
  dragOverFolderID.value = null
  dragOverBcPath.value = null
}

function _canDropOnPath(targetPath) {
  if (!_dragItem) return false
  if (targetPath === _dragItem.currentPath) return false
  if (_dragItem.type === 'folder' && _dragItem.path) {
    if (targetPath === _dragItem.path) return false
    if (targetPath.startsWith(_dragItem.path + '/')) return false
  }
  return true
}

function onFolderDragOver(e, folder) {
  if (!_canDropOnPath(folder.path)) return
  e.preventDefault()
  e.dataTransfer.dropEffect = 'move'
  dragOverFolderID.value = folder.id
}

function onFolderDragLeave(folder) {
  if (dragOverFolderID.value === folder.id) dragOverFolderID.value = null
}

async function onDropOnFolder(e, folder) {
  e.preventDefault()
  dragOverFolderID.value = null
  if (!_dragItem || !_canDropOnPath(folder.path)) return
  const { id, type } = _dragItem
  _dragItem = null
  try {
    if (type === 'file') await orgStore.moveOrgFile(orgID.value, id, folder.path)
    else await orgStore.moveOrgFolder(orgID.value, id, folder.path)
    showToast(t('orgs.moveSuccess'))
  } catch (err) {
    showToast(err.response?.data?.error || err.message, 'error')
  }
}

function onBcDragOver(e, path) {
  if (!_canDropOnPath(path)) return
  e.preventDefault()
  e.dataTransfer.dropEffect = 'move'
  dragOverBcPath.value = path
}

function onBcDragLeave(path) {
  if (dragOverBcPath.value === path) dragOverBcPath.value = null
}

async function onDropOnPath(e, path) {
  e.preventDefault()
  dragOverBcPath.value = null
  if (!_dragItem || !_canDropOnPath(path)) return
  const { id, type } = _dragItem
  _dragItem = null
  try {
    if (type === 'file') await orgStore.moveOrgFile(orgID.value, id, path)
    else await orgStore.moveOrgFolder(orgID.value, id, path)
    showToast(t('orgs.moveSuccess'))
  } catch (err) {
    showToast(err.response?.data?.error || err.message, 'error')
  }
}

// ── Bulk selection ────────────────────────────────────────────────────────────

const selectedIDs = ref(new Set())
const activeFolderID = ref(null)
const bulkLoading = ref(false)
const bulkMoveMode = ref(false)
const zipDownloadStates = ref({})

function selKey(type, id) { return `${type}-${id}` }
function isSelected(type, id) { return selectedIDs.value.has(selKey(type, id)) }

function toggleSelect(e, type, id) {
  e.stopPropagation()
  const key = selKey(type, id)
  const s = new Set(selectedIDs.value)
  if (s.has(key)) s.delete(key)
  else s.add(key)
  selectedIDs.value = s
}

function selectFolderRow(id) {
  activeFolderID.value = id
  const key = selKey('folder', id)
  if (!selectedIDs.value.has(key)) {
    const s = new Set(selectedIDs.value)
    s.add(key)
    selectedIDs.value = s
  }
}

const hasSelection = computed(() => selectedIDs.value.size > 0)
const selectedCount = computed(() => selectedIDs.value.size)

const allVisibleSelected = computed(() => {
  const total = sortedFolders.value.length + sortedFiles.value.length
  return total > 0 && selectedIDs.value.size === total
})

function toggleSelectAll() {
  if (allVisibleSelected.value) {
    selectedIDs.value = new Set()
  } else {
    const s = new Set()
    sortedFolders.value.forEach(f => s.add(selKey('folder', f.id)))
    sortedFiles.value.forEach(f => s.add(selKey('file', f.id)))
    selectedIDs.value = s
  }
}

function clearSelection() { selectedIDs.value = new Set() }

const selectedFileItems = computed(() =>
  sortedFiles.value.filter(f => isSelected('file', f.id))
)
const selectedFolderItems = computed(() =>
  sortedFolders.value.filter(f => isSelected('folder', f.id))
)

async function bulkDelete() {
  if (!confirm(t('orgs.bulkDeleteConfirm', { count: selectedCount.value }))) return
  bulkLoading.value = true
  let failed = 0
  for (const folder of selectedFolderItems.value) {
    try { await orgStore.deleteFolder(orgID.value, folder.id) }
    catch { failed++ }
  }
  for (const file of selectedFileItems.value) {
    try { await orgStore.deleteFile(orgID.value, file.id) }
    catch { failed++ }
  }
  const done = selectedCount.value - failed
  clearSelection()
  bulkLoading.value = false
  if (failed > 0) showToast(t('orgs.bulkDeletePartial', { done, failed }), 'error')
  else showToast(t('orgs.bulkDeleteSuccess', { count: done }))
}

async function openBulkMoveDialog() {
  bulkMoveMode.value = true
  const first = selectedFolderItems.value[0] || selectedFileItems.value[0]
  movingItem.value = {
    id: null,
    name: t('orgs.bulkSelected', { count: selectedCount.value }),
    type: 'bulk',
    currentPath: null,
  }
  moveDestination.value = null
  moveSearch.value = ''
  showMoveModal.value = true
  moveFetching.value = true
  try {
    allOrgFolders.value = await orgStore.getAllOrgFolders(orgID.value)
  } catch (e) {
    showToast(e.response?.data?.error || e.message, 'error')
    showMoveModal.value = false
  } finally {
    moveFetching.value = false
  }
}

async function bulkDownload() {
  if (selectedFolderItems.value.length > 0) {
    bulkLoading.value = true
    try {
      const count = await orgStore.downloadSelectionAsZip(
        orgID.value, selectedFileItems.value, selectedFolderItems.value
      )
      if (count === 0) showToast(t('orgs.zipEmpty'), 'info')
      else showToast(t('orgs.zipFilesDownloaded', { n: count }))
      clearSelection()
    } catch (e) {
      showToast(e.response?.data?.error || e.message, 'error')
    } finally {
      bulkLoading.value = false
    }
  } else {
    for (const file of selectedFileItems.value) await handleDownload(file)
  }
}

async function handleFolderZipDownload(folder) {
  if (zipDownloadStates.value[folder.id]) return
  zipDownloadStates.value[folder.id] = true
  showToast(t('orgs.zipBuilding'), 'info')
  try {
    const count = await orgStore.downloadFolderAsZip(orgID.value, folder.path, folder.name)
    if (count === 0) showToast(t('orgs.zipEmpty'), 'info')
    else showToast(t('orgs.zipDownloaded', { name: folder.name }))
  } catch (e) {
    showToast(e.response?.data?.error || e.message, 'error')
  } finally {
    delete zipDownloadStates.value[folder.id]
  }
}

// ── Sort & filter ─────────────────────────────────────────────────────────────

const sortBy = ref('name')
const sortDir = ref('asc')
const filterType = ref('all')

function capitalize(s) { return s.charAt(0).toUpperCase() + s.slice(1) }

function toggleSort(field) {
  if (sortBy.value === field) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortBy.value = field
    sortDir.value = 'asc'
  }
}

function fileMatchesFilter(mimeType) {
  if (filterType.value === 'all') return true
  if (!mimeType) return false
  switch (filterType.value) {
    case 'images': return mimeType.startsWith('image/')
    case 'videos': return mimeType.startsWith('video/')
    case 'audio': return mimeType.startsWith('audio/')
    case 'documents':
      return mimeType === 'application/pdf' ||
        mimeType.startsWith('text/') ||
        mimeType.startsWith('application/msword') ||
        mimeType.startsWith('application/vnd.openxmlformats-officedocument') ||
        mimeType.startsWith('application/vnd.oasis.opendocument')
    case 'archives':
      return ['application/zip','application/x-tar','application/gzip',
        'application/x-rar-compressed','application/x-7z-compressed',
        'application/x-bzip2'].some(t => mimeType.startsWith(t))
    default: return true
  }
}

const sortedFolders = computed(() => {
  let folders = [...(orgStore.currentItems.folders || [])]
  if (filterTagID.value !== null)
    folders = folders.filter(f => (f.tag_ids || []).includes(filterTagID.value))
  return folders.sort((a, b) => {
    const diff = sortBy.value === 'date'
      ? new Date(a.created_at) - new Date(b.created_at)
      : a.name.localeCompare(b.name)
    return sortDir.value === 'asc' ? diff : -diff
  })
})

const sortedFiles = computed(() => {
  let files = [...(orgStore.currentItems.files || [])].filter(f => fileMatchesFilter(f.mime_type))
  if (filterTagID.value !== null)
    files = files.filter(f => (f.tag_ids || []).includes(filterTagID.value))
  files.sort((a, b) => {
    let diff
    if (sortBy.value === 'size') diff = a.size - b.size
    else if (sortBy.value === 'date') diff = new Date(a.created_at) - new Date(b.created_at)
    else diff = a.name.localeCompare(b.name)
    return sortDir.value === 'asc' ? diff : -diff
  })
  return files
})

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
const provisioningAll = ref(false)

// Members who can be provisioned right now (have a public key but no org key yet).
const membersNeedingKey = computed(() =>
  orgStore.members.filter(m => !m.encrypted_org_key && m.public_key && m.user_id !== myUserID.value)
)

// Sort: members needing provisioning first, then alphabetically by name.
const sortedMembers = computed(() => {
  return [...orgStore.members].sort((a, b) => {
    const aNeedsKey = !a.encrypted_org_key ? 0 : 1
    const bNeedsKey = !b.encrypted_org_key ? 0 : 1
    if (aNeedsKey !== bNeedsKey) return aNeedsKey - bNeedsKey
    return (a.name || '').localeCompare(b.name || '')
  })
})

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

const handleProvisionAll = async () => {
  provisioningAll.value = true
  try {
    const count = await orgStore.provisionAllMissingKeys(orgID.value)
    showToast(t('orgs.allKeysProvisioned', { count }))
  } catch (e) {
    showToast(e.response?.data?.error || e.message, 'error')
  } finally {
    provisioningAll.value = false
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

const quotaDialogMember = ref(null)
const quotaDialogBytes = ref(0)
const savingQuota = ref(false)

const openQuotaDialog = (member) => {
  quotaDialogMember.value = member
  quotaDialogBytes.value = member.quota_bytes ?? 0
}

const handleSaveQuota = async () => {
  if (!quotaDialogMember.value) return
  savingQuota.value = true
  try {
    await orgStore.setMemberQuota(orgID.value, quotaDialogMember.value.id, quotaDialogBytes.value)
    showToast(t('orgs.memberQuotaSaved'))
    quotaDialogMember.value = null
  } catch (e) {
    showToast(e.response?.data?.error || e.message, 'error')
  } finally {
    savingQuota.value = false
  }
}

const formatBytes = (bytes) => {
  if (!bytes || bytes === 0) return '∞'
  const gb = bytes / (1024 ** 3)
  return gb >= 1 ? `${gb.toFixed(1)} GB` : `${(bytes / (1024 ** 2)).toFixed(0)} MB`
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
const inviteForm = ref({ role: 'member', maxUses: 0, expiresAt: '', sendEmail: false, recipientEmail: '', emailLang: locale.value === 'fr' ? 'fr' : 'en' })

const handleCreateInvite = async () => {
  creatingInvite.value = true
  inviteError.value = ''
  try {
    const payload = {
      role: inviteForm.value.role,
      max_uses: inviteForm.value.maxUses,
      send_email: inviteForm.value.sendEmail,
      recipient_email: inviteForm.value.recipientEmail,
      email_lang: inviteForm.value.emailLang,
    }
    if (inviteForm.value.expiresAt) {
      payload.expires_at = new Date(inviteForm.value.expiresAt).toISOString()
    }
    const inv = await orgStore.createInvitation(orgID.value, payload)
    showInviteModal.value = false
    inviteForm.value = { role: 'member', maxUses: 0, expiresAt: '', sendEmail: false, recipientEmail: '', emailLang: locale.value === 'fr' ? 'fr' : 'en' }
    if (inv.email_notified) {
      showToast(t('orgs.inviteCreatedWithEmail'))
    } else {
      showToast(t('orgs.inviteCreated'))
    }
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

const settingsForm = ref({ name: '', description: '', storageQuotaMB: 10240, requireMFA: false })
const savingSettings = ref(false)
const settingsError = ref('')

const handleSaveSettings = async () => {
  savingSettings.value = true
  settingsError.value = ''
  try {
    const payload = {
      name: settingsForm.value.name,
      description: settingsForm.value.description,
      require_mfa: settingsForm.value.requireMFA,
    }
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

const handleExportAudit = async () => {
  try {
    await orgStore.exportAuditLog(orgID.value)
  } catch (e) {
    showToast(e.response?.data?.error || e.message, 'error')
  }
}

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
  position: relative;
}
.tab-content.drop-zone-active { outline: 2px dashed var(--primary-color, #6366f1); outline-offset: -4px; border-radius: 8px; }
.drop-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  background: var(--primary-color-light, rgba(99,102,241,0.12));
  border-radius: 8px;
  z-index: 20;
  pointer-events: none;
  color: var(--primary-color, #6366f1);
  font-size: 1rem;
  font-weight: 600;
}

.folder-row.drag-over {
  background: var(--primary-color-light, rgba(99,102,241,0.15)) !important;
  outline: 1px dashed var(--primary-color, #6366f1);
  border-radius: 6px;
}
.bc-item.bc-drag-over {
  background: var(--primary-color-light, rgba(99,102,241,0.15));
  color: var(--primary-color, #6366f1);
  border-radius: 4px;
}
.item-row[draggable="true"] { cursor: grab; }
.item-row[draggable="true"]:active { cursor: grabbing; }
.item-row.previewable { cursor: pointer; }

.preview-overlay { align-items: center; justify-content: center; }
.preview-modal {
  background: var(--surface-color, #1e1e2e);
  border: 1px solid var(--border-color, rgba(255,255,255,0.1));
  border-radius: 14px;
  width: min(90vw, 900px);
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 24px 80px rgba(0,0,0,0.5);
}
.preview-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 18px;
  border-bottom: 1px solid var(--border-color, rgba(255,255,255,0.08));
  gap: 12px;
  flex-shrink: 0;
}
.preview-title {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}
.preview-file-icon { flex-shrink: 0; opacity: 0.6; }
.preview-filename {
  font-weight: 600;
  font-size: 0.95rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.preview-header-actions { display: flex; align-items: center; gap: 8px; flex-shrink: 0; }
.preview-body {
  flex: 1;
  overflow: auto;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 200px;
  background: var(--bg-color, #13131f);
}
.preview-loading { display: flex; flex-direction: column; align-items: center; gap: 12px; color: var(--secondary-text-color); }
.preview-image { max-width: 100%; max-height: 70vh; object-fit: contain; display: block; }
.preview-video { max-width: 100%; max-height: 70vh; display: block; }
.preview-audio-wrap { display: flex; flex-direction: column; align-items: center; gap: 20px; padding: 40px; }
.preview-audio { width: 320px; max-width: 100%; }
.preview-pdf { width: 100%; height: 70vh; border: none; display: block; }
.preview-text {
  width: 100%;
  height: 100%;
  min-height: 300px;
  max-height: 70vh;
  overflow: auto;
  padding: 20px;
  margin: 0;
  font-family: 'Fira Code', 'Cascadia Code', monospace;
  font-size: 0.82rem;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
  color: var(--text-color);
  background: var(--bg-color, #13131f);
  align-self: stretch;
}
.preview-unsupported { display: flex; flex-direction: column; align-items: center; gap: 12px; padding: 48px; color: var(--secondary-text-color); }
.preview-footer {
  display: flex;
  gap: 16px;
  padding: 10px 18px;
  border-top: 1px solid var(--border-color, rgba(255,255,255,0.08));
  font-size: 0.78rem;
  color: var(--secondary-text-color);
  flex-shrink: 0;
}
.preview-meta { white-space: nowrap; }

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

.item-rename-input {
  flex: 1;
  font-size: 0.88rem;
  font-weight: 500;
  color: var(--main-text-color);
  background: var(--input-background-color, var(--background-secondary-color));
  border: 1px solid var(--primary-color);
  border-radius: 4px;
  padding: 2px 6px;
  outline: none;
  min-width: 0;
}

.move-modal { max-width: 420px; width: 100%; }
.move-item-label {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--main-text-color);
  margin: 0 0 10px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.move-search-input {
  width: 100%;
  box-sizing: border-box;
  padding: 7px 10px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  background: var(--background-secondary-color);
  color: var(--main-text-color);
  font-size: 0.85rem;
  outline: none;
  margin-bottom: 8px;
}
.move-search-input:focus { border-color: var(--primary-color); }
.move-fetching { display: flex; justify-content: center; padding: 20px; }
.move-folder-list {
  max-height: 260px;
  overflow-y: auto;
  border: 1px solid var(--border-color);
  border-radius: 6px;
}
.move-folder-item {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 8px 10px;
  background: none;
  border: none;
  cursor: pointer;
  text-align: left;
  border-bottom: 1px solid var(--border-color);
  transition: background 0.12s;
}
.move-folder-item:last-child { border-bottom: none; }
.move-folder-item:hover { background: var(--hover-background-color); }
.move-folder-item.is-selected { background: color-mix(in srgb, var(--primary-color) 12%, transparent); }
.move-folder-item.is-current { opacity: 0.55; cursor: default; }
.move-folder-icon { flex-shrink: 0; color: var(--secondary-text-color); }
.move-folder-info { flex: 1; min-width: 0; }
.move-folder-name { display: block; font-size: 0.85rem; font-weight: 500; color: var(--main-text-color); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.move-folder-path { display: block; font-size: 0.72rem; color: var(--secondary-text-color); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.move-current-badge { font-size: 0.7rem; color: var(--secondary-text-color); white-space: nowrap; flex-shrink: 0; }
.move-check { flex-shrink: 0; color: var(--primary-color); }
.move-empty { text-align: center; padding: 16px; font-size: 0.85rem; color: var(--secondary-text-color); }

.bulk-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 8px 12px;
  background: var(--primary-color-light, rgba(99,102,241,0.12));
  border: 1px solid var(--primary-color, #6366f1);
  border-radius: 8px;
  margin-bottom: 8px;
}
.bulk-count { font-size: 0.85rem; font-weight: 600; color: var(--primary-color, #6366f1); white-space: nowrap; }
.bulk-actions { display: flex; gap: 6px; flex-wrap: wrap; }
.btn-danger {
  background: rgba(239,68,68,0.12) !important;
  border-color: #ef4444 !important;
  color: #ef4444 !important;
}
.btn-danger:hover { background: rgba(239,68,68,0.22) !important; }
.btn-ghost { background: none !important; border-color: transparent !important; }
.btn-ghost:hover { background: var(--hover-bg, rgba(255,255,255,0.06)) !important; }

.select-all-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 8px;
  font-size: 0.8rem;
  color: var(--secondary-text-color);
  border-bottom: 1px solid var(--border-color, rgba(255,255,255,0.06));
  margin-bottom: 2px;
}
.select-all-label { cursor: pointer; user-select: none; }

.checkbox-wrap {
  display: flex;
  align-items: center;
  flex-shrink: 0;
  cursor: pointer;
  padding: 2px 4px;
}
.item-checkbox {
  width: 15px;
  height: 15px;
  cursor: pointer;
  accent-color: var(--primary-color, #6366f1);
  opacity: 0;
  transition: opacity 0.15s;
}
.item-row:hover .item-checkbox,
.item-row.selected .item-checkbox { opacity: 1; }

.item-row.selected { background: var(--primary-color-light, rgba(99,102,241,0.08)); }
.item-row.folder-active { background: var(--primary-color-light, rgba(99,102,241,0.13)); outline: 1px solid var(--primary-color, #6366f1); outline-offset: -1px; }

.btn-ghost-sm {
  background: none !important;
  border-color: var(--border-color, rgba(255,255,255,0.1)) !important;
  color: var(--secondary-text-color) !important;
  font-size: 0.78rem !important;
  padding: 3px 8px !important;
}
.btn-ghost-sm:hover { background: var(--hover-bg, rgba(255,255,255,0.06)) !important; color: var(--text-color) !important; }

.tag-filter-group { display: flex; gap: 4px; align-items: center; flex-wrap: wrap; }
.tag-filter-btn {
  display: flex;
  align-items: center;
  gap: 5px;
  background: none;
  border: 1px solid transparent;
  border-radius: 12px;
  padding: 2px 8px;
  font-size: 0.78rem;
  cursor: pointer;
  color: var(--secondary-text-color);
  transition: background 0.15s, border-color 0.15s, color 0.15s;
  white-space: nowrap;
}
.tag-filter-btn:hover { background: var(--hover-bg, rgba(255,255,255,0.06)); color: var(--text-color); }
.tag-filter-btn.active { border-color: var(--tag-color); background: color-mix(in srgb, var(--tag-color) 15%, transparent); color: var(--text-color); }
.tag-dot { width: 8px; height: 8px; border-radius: 50%; background: var(--tag-color); flex-shrink: 0; }

.item-tags { display: flex; align-items: center; gap: 3px; flex-shrink: 0; }
.item-tag-dot { display: inline-block; width: 9px; height: 9px; border-radius: 50%; flex-shrink: 0; }
.item-tag-more { font-size: 0.7rem; color: var(--secondary-text-color); white-space: nowrap; }

.tag-popover-backdrop { position: fixed; inset: 0; z-index: 199; }
.tag-popover {
  position: fixed;
  z-index: 200;
  background: var(--surface-color, #1e1e2e);
  border: 1px solid var(--border-color, rgba(255,255,255,0.12));
  border-radius: 10px;
  padding: 6px;
  min-width: 180px;
  max-width: 240px;
  box-shadow: 0 8px 32px rgba(0,0,0,0.4);
}
.tag-popover-empty { font-size: 0.82rem; color: var(--secondary-text-color); padding: 6px 8px; }
.tag-popover-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 5px 8px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.85rem;
  transition: background 0.12s;
}
.tag-popover-item:hover { background: var(--hover-bg, rgba(255,255,255,0.06)); }
.tag-popover-item input[type="checkbox"] { accent-color: var(--primary-color, #6366f1); }
.tag-popover-dot { width: 10px; height: 10px; border-radius: 50%; flex-shrink: 0; }
.tag-popover-name { flex: 1; }
.tag-popover-footer { border-top: 1px solid var(--border-color, rgba(255,255,255,0.08)); margin-top: 4px; padding-top: 4px; }
.tag-popover-manage { background: none; border: none; cursor: pointer; font-size: 0.78rem; color: var(--secondary-text-color); padding: 4px 8px; width: 100%; text-align: left; border-radius: 4px; }
.tag-popover-manage:hover { background: var(--hover-bg); color: var(--text-color); }

.tag-mgr-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 4px;
  border-bottom: 1px solid var(--border-color, rgba(255,255,255,0.06));
}
.tag-mgr-dot { width: 12px; height: 12px; border-radius: 50%; flex-shrink: 0; }
.tag-mgr-name { flex: 1; font-size: 0.88rem; cursor: pointer; }
.tag-mgr-name:hover { text-decoration: underline; }
.tag-mgr-input { flex: 1; background: var(--input-bg, rgba(255,255,255,0.06)); border: 1px solid var(--border-color); border-radius: 5px; padding: 3px 8px; font-size: 0.85rem; color: var(--text-color); }
.tag-mgr-colors { display: flex; gap: 4px; flex-shrink: 0; }
.tag-color-swatch { width: 14px; height: 14px; border-radius: 50%; border: 2px solid transparent; cursor: pointer; padding: 0; transition: border-color 0.12s; }
.tag-color-swatch.active, .tag-color-swatch:hover { border-color: var(--text-color, #fff); }
.tag-mgr-create { display: flex; align-items: center; gap: 8px; padding-top: 10px; flex-wrap: wrap; }

.sort-filter-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 6px 0 8px;
  flex-wrap: wrap;
  border-bottom: 1px solid var(--border-color, rgba(255,255,255,0.08));
  margin-bottom: 4px;
}
.sort-group, .filter-group { display: flex; gap: 4px; align-items: center; }
.sort-btn, .filter-btn {
  background: none;
  border: 1px solid transparent;
  border-radius: 6px;
  padding: 3px 10px;
  font-size: 0.8rem;
  cursor: pointer;
  color: var(--secondary-text-color);
  display: flex;
  align-items: center;
  gap: 4px;
  transition: background 0.15s, color 0.15s, border-color 0.15s;
  white-space: nowrap;
}
.sort-btn:hover, .filter-btn:hover { background: var(--hover-bg, rgba(255,255,255,0.06)); color: var(--text-color); }
.sort-btn.active, .filter-btn.active {
  background: var(--primary-color-light, rgba(99,102,241,0.15));
  border-color: var(--primary-color, #6366f1);
  color: var(--primary-color, #6366f1);
}
.sort-filter-divider { width: 1px; height: 18px; background: var(--border-color, rgba(255,255,255,0.08)); }

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

.invite-meta { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }

.invite-detail { font-size: 0.75rem; color: var(--secondary-text-color); }

.invite-email-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 0.72rem;
  color: #2A9D8F;
  background: color-mix(in srgb, #2A9D8F 10%, transparent);
  padding: 2px 7px;
  border-radius: 10px;
  max-width: 180px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* Email section in invite modal */
.invite-email-section {
  border-top: 1px solid var(--border-color);
  padding-top: 14px;
  margin-top: 4px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.82rem;
  font-weight: 600;
  color: var(--secondary-text-color);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  cursor: pointer;
  user-select: none;
}

.invite-email-header {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.82rem;
  font-weight: 600;
  color: var(--secondary-text-color);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.invite-lang-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.invite-lang-label {
  font-size: 0.82rem;
  color: var(--secondary-text-color);
  white-space: nowrap;
}

.lang-toggle {
  display: flex;
  gap: 4px;
}

.lang-btn {
  background: var(--background-color);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  padding: 4px 10px;
  font-size: 0.8rem;
  cursor: pointer;
  color: var(--secondary-text-color);
  transition: border-color 0.15s, color 0.15s;
}
.lang-btn:hover { border-color: var(--primary-color); }
.lang-btn.active {
  border-color: var(--primary-color);
  color: var(--primary-color);
  font-weight: 600;
  background: color-mix(in srgb, var(--primary-color) 8%, transparent);
}

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

.toggle-label { display: flex; align-items: center; gap: 8px; cursor: pointer; font-size: 14px; }
.toggle-checkbox { width: 16px; height: 16px; cursor: pointer; }

/* MFA required overlay */
.mfa-required-overlay {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  padding: 60px 24px;
  text-align: center;
  color: var(--text-secondary);
}
.mfa-required-overlay h3 { margin: 0; font-size: 18px; color: var(--text-primary); }
.mfa-required-overlay p  { margin: 0; font-size: 14px; max-width: 400px; }

/* Quota badge on member row */
.quota-badge {
  font-size: 11px;
  padding: 2px 6px;
  border-radius: 4px;
  background: color-mix(in srgb, var(--primary-color) 12%, transparent);
  color: var(--primary-color);
  white-space: nowrap;
}

/* Download count on share row */
.share-downloads {
  display: flex;
  align-items: center;
  gap: 3px;
  font-size: 12px;
  color: var(--text-secondary);
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

/* Key-pending banner — for members waiting on admin provisioning */
.key-pending-banner {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  margin-bottom: 16px;
  background: color-mix(in srgb, var(--primary-color) 6%, var(--card-color));
  border: 1px solid color-mix(in srgb, var(--primary-color) 25%, transparent);
  border-radius: 10px;
  color: var(--secondary-text-color);
}

.key-pending-text {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.key-pending-text strong { font-size: 0.84rem; color: var(--main-text-color); font-weight: 600; }
.key-pending-text span   { font-size: 0.75rem; color: var(--secondary-text-color); }

/* Provision-all banner in members tab */
.provision-banner {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  margin-bottom: 14px;
  background: color-mix(in srgb, var(--primary-color) 8%, transparent);
  border: 1px solid color-mix(in srgb, var(--primary-color) 25%, transparent);
  border-radius: 8px;
  font-size: 0.83rem;
  color: var(--main-text-color);
}

.provision-banner svg { flex-shrink: 0; color: var(--primary-color); }

.btn-provision-all {
  margin-left: auto;
  background: var(--primary-color);
  color: white;
  border: none;
  border-radius: 6px;
  padding: 5px 14px;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  white-space: nowrap;
  transition: opacity 0.15s;
}
.btn-provision-all:hover:not(:disabled) { opacity: 0.87; }
.btn-provision-all:disabled { opacity: 0.5; cursor: not-allowed; }

/* Highlight member rows needing key provision */
.member-row.needs-key {
  border-left: 3px solid color-mix(in srgb, var(--primary-color) 60%, transparent);
  padding-left: calc(var(--member-row-padding, 12px) - 3px);
}

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
.toast.info { background: #2563eb; }

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

/* ── Search ─────────────────────────────────────────────────────────────────── */
.search-bar-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--hover-background-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 0 10px;
  margin-bottom: 12px;
  transition: border-color 0.15s, box-shadow 0.15s;
}
.search-bar-wrap:focus-within {
  border-color: var(--primary-color);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--primary-color) 18%, transparent);
}

.search-icon {
  color: var(--secondary-text-color);
  flex-shrink: 0;
}

.search-input {
  flex: 1;
  background: transparent;
  border: none;
  outline: none;
  padding: 9px 0;
  font-size: 0.88rem;
  color: var(--main-text-color);
}
.search-input::placeholder { color: var(--secondary-text-color); }

.search-clear {
  background: none;
  border: none;
  padding: 0;
  color: var(--secondary-text-color);
  cursor: pointer;
  display: flex;
  align-items: center;
  flex-shrink: 0;
}
.search-clear:hover { color: var(--main-text-color); }

.search-results {
  background: var(--card-background-color, var(--background-color));
  border: 1px solid var(--border-color);
  border-radius: 10px;
  overflow: hidden;
  margin-bottom: 12px;
}

.search-empty {
  padding: 20px 16px;
  text-align: center;
  font-size: 0.85rem;
  color: var(--secondary-text-color);
}

.search-count {
  padding: 7px 14px;
  font-size: 0.75rem;
  color: var(--secondary-text-color);
  border-bottom: 1px solid var(--border-color);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.search-result-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 9px 14px;
  cursor: pointer;
  transition: background 0.1s;
  border-bottom: 1px solid var(--border-color);
}
.search-result-row:last-child { border-bottom: none; }
.search-result-row:hover { background: var(--hover-background-color); }

.search-result-icon { flex-shrink: 0; }
.search-result-icon.folder { color: #f59e0b; }
.search-result-icon.file   { color: var(--secondary-text-color); }

.search-result-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.search-result-name {
  font-size: 0.86rem;
  font-weight: 500;
  color: var(--main-text-color);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.search-result-path {
  font-size: 0.74rem;
  color: var(--secondary-text-color);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.search-result-arrow { color: var(--secondary-text-color); flex-shrink: 0; }

:deep(.search-highlight) {
  background: color-mix(in srgb, var(--primary-color) 22%, transparent);
  color: var(--primary-color);
  border-radius: 2px;
  font-weight: 600;
  padding: 0 1px;
}

/* ── Dashboard ─────────────────────────────────────────────────────────────── */
.dash-alert {
  display: flex;
  align-items: center;
  gap: 8px;
  background: color-mix(in srgb, var(--warning-color, #f59e0b) 12%, transparent);
  border: 1px solid color-mix(in srgb, var(--warning-color, #f59e0b) 35%, transparent);
  border-radius: 8px;
  padding: 10px 14px;
  font-size: 0.85rem;
  color: var(--main-text-color);
  margin-bottom: 16px;
}

.btn-link {
  background: none;
  border: none;
  padding: 0;
  color: var(--primary-color);
  font-size: 0.85rem;
  cursor: pointer;
  text-decoration: underline;
  margin-left: 4px;
}

.dash-kpis {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(110px, 1fr));
  gap: 12px;
  margin-bottom: 24px;
}

.dash-kpi {
  background: var(--card-background-color, var(--hover-background-color));
  border: 1px solid var(--border-color);
  border-radius: 10px;
  padding: 16px 14px;
  text-align: center;
}

.dash-kpi-value {
  font-size: 1.7rem;
  font-weight: 700;
  color: var(--primary-color);
  line-height: 1;
}

.dash-kpi-label {
  font-size: 0.75rem;
  color: var(--secondary-text-color);
  margin-top: 6px;
}

.dash-section {
  margin-top: 24px;
}

.dash-section-title {
  font-size: 0.82rem;
  font-weight: 600;
  color: var(--secondary-text-color);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin: 0 0 12px;
}

.dash-storage-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.dash-storage-row {
  display: grid;
  grid-template-columns: 160px 1fr auto;
  align-items: center;
  gap: 12px;
}

.dash-storage-identity {
  display: flex;
  align-items: center;
  gap: 8px;
  overflow: hidden;
}

.member-avatar.small {
  width: 26px;
  height: 26px;
  min-width: 26px;
  font-size: 0.72rem;
}

.dash-storage-name {
  font-size: 0.84rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dash-storage-bar-wrap {
  height: 8px;
  background: var(--border-color);
  border-radius: 4px;
  overflow: hidden;
}

.dash-storage-bar {
  height: 100%;
  background: var(--primary-color);
  border-radius: 4px;
  transition: width 0.4s ease;
  min-width: 2px;
}

.dash-storage-meta {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 2px;
  font-size: 0.82rem;
  white-space: nowrap;
}

.dash-file-count {
  font-size: 0.75rem;
  color: var(--secondary-text-color);
}

.dash-audit-link {
  display: block;
  margin-top: 10px;
  font-size: 0.83rem;
}

/* ── Share modal ─────────────────────────────────────────────────────────── */

.share-loading {
  display: flex;
  align-items: center;
  gap: 10px;
  color: var(--secondary-text-color);
  font-size: 0.9rem;
}

.share-hint {
  font-size: 0.87rem;
  color: var(--secondary-text-color);
  margin: 0 0 12px;
}

.share-link-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.share-link-input {
  flex: 1;
  padding: 8px 10px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--input-background);
  color: var(--main-text-color);
  font-size: 0.8rem;
  font-family: monospace;
  min-width: 0;
}

.btn-copy {
  flex-shrink: 0;
  padding: 8px 10px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--card-color);
  color: var(--main-text-color);
  cursor: pointer;
  display: flex;
  align-items: center;
  transition: background 0.15s, color 0.15s;
}

.btn-copy:hover { background: var(--hover-background-color); color: var(--primary-color); }

.share-warning {
  margin: 10px 0 0;
  font-size: 0.78rem;
  color: var(--warning-color, #f59e0b);
}

.share-error-msg {
  color: var(--error-color);
  font-size: 0.87rem;
}

.share-form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.share-form-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.share-form-field label {
  font-size: 0.85rem;
  color: var(--secondary-text-color);
  font-weight: 500;
}

.share-form-input {
  padding: 8px 10px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--input-background, var(--card-color));
  color: var(--main-text-color);
  font-size: 0.9rem;
  width: 100%;
  box-sizing: border-box;
}

.share-single-use-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-size: 0.9rem;
  color: var(--main-text-color);
  font-weight: 500;
}

.share-single-use-label input[type="checkbox"] {
  accent-color: var(--primary-color);
  width: 15px;
  height: 15px;
}

.share-single-use-hint {
  font-size: 0.78rem;
  color: var(--secondary-text-color);
  margin: -4px 0 0 23px;
}

/* ── Pinned / favorites ─────────────────────────────────────────────────────── */

.pinned-section {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 0 6px;
  border-bottom: 1px solid var(--border-color, #e5e7eb);
  margin-bottom: 4px;
  flex-wrap: wrap;
}

.pinned-label {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 0.72rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-secondary, #6b7280);
  white-space: nowrap;
  flex-shrink: 0;
}

.pinned-chips {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.pinned-chip {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 3px 8px 3px 6px;
  border-radius: 20px;
  border: 1px solid var(--border-color, #e5e7eb);
  background: var(--bg-secondary, #f9fafb);
  font-size: 0.78rem;
  cursor: pointer;
  color: var(--text-primary, #111827);
  transition: background 0.15s, border-color 0.15s;
  max-width: 180px;
}

.pinned-chip:hover {
  background: var(--bg-hover, #f3f4f6);
  border-color: var(--accent-color, #6366f1);
}

.pinned-chip-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 120px;
}

.pinned-chip-remove {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 14px;
  height: 14px;
  border-radius: 50%;
  border: none;
  background: transparent;
  color: var(--text-secondary, #6b7280);
  cursor: pointer;
  padding: 0;
  flex-shrink: 0;
  transition: background 0.12s, color 0.12s;
}

.pinned-chip-remove:hover {
  background: var(--danger-light, #fee2e2);
  color: var(--danger-color, #ef4444);
}

.btn-icon-pinned {
  color: #f59e0b !important;
}

/* ── Activity feed ─────────────────────────────────────────────────────────── */

.activity-feed {
  display: flex;
  flex-direction: column;
  gap: 24px;
  max-width: 680px;
}

.activity-day-group {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.activity-day-label {
  font-size: 0.72rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--subtle-text-color, #8a8a8a);
  padding: 0 0 6px 0;
  border-bottom: 1px solid var(--border-color);
  margin-bottom: 4px;
}

.activity-entry {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 7px 10px;
  border-radius: 8px;
  transition: background 0.12s;
}

.activity-entry:hover {
  background: var(--hover-background-color);
}

.activity-icon {
  flex-shrink: 0;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: var(--card-color);
  border: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--primary-color);
}

.activity-body {
  flex: 1;
  min-width: 0;
  font-size: 0.85rem;
  line-height: 1.4;
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  align-items: baseline;
}

.activity-actor {
  font-weight: 600;
  color: var(--main-text-color);
  flex-shrink: 0;
}

.activity-desc {
  color: var(--secondary-text-color, #666);
  word-break: break-word;
}

.activity-time {
  flex-shrink: 0;
  font-size: 0.75rem;
  color: var(--subtle-text-color, #8a8a8a);
  white-space: nowrap;
}

/* ── Trash ─────────────────────────────────────────────────────────────────── */

.trash-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.trash-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  border-radius: 8px;
  background: var(--card-color);
  border: 1px solid var(--border-color);
  transition: background 0.12s;
}

.trash-row:hover {
  background: var(--hover-background-color);
}

.trash-row-icon {
  flex-shrink: 0;
  color: var(--subtle-text-color, #8a8a8a);
  display: flex;
  align-items: center;
}

.trash-row-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.trash-row-name {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--main-text-color);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.trash-row-path {
  font-size: 0.75rem;
  color: var(--subtle-text-color, #8a8a8a);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.trash-row-meta {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 1px;
  min-width: 110px;
}

.trash-row-date {
  font-size: 0.75rem;
  color: var(--subtle-text-color, #8a8a8a);
  white-space: nowrap;
}

.trash-row-by {
  font-size: 0.72rem;
  color: var(--subtle-text-color, #8a8a8a);
  white-space: nowrap;
}

.trash-row-actions {
  flex-shrink: 0;
  display: flex;
  gap: 6px;
}

.btn-danger-sm {
  background: transparent;
  border: 1px solid var(--danger-color, #ef4444);
  color: var(--danger-color, #ef4444);
  font-size: 0.78rem;
  padding: 3px 10px;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.12s, color 0.12s;
}

.btn-danger-sm:hover {
  background: var(--danger-color, #ef4444);
  color: #fff;
}

/* ── Shares ─────────────────────────────────────────────────────────────────── */

.shares-scope-hint {
  font-size: 0.8rem;
  color: var(--subtle-text-color, #8a8a8a);
  margin: 0 0 12px 0;
}

.shares-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.share-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  border-radius: 8px;
  background: var(--card-color);
  border: 1px solid var(--border-color);
  transition: background 0.12s;
}

.share-row:hover {
  background: var(--hover-background-color);
}

.share-row-icon {
  flex-shrink: 0;
  color: var(--subtle-text-color, #8a8a8a);
  display: flex;
  align-items: center;
}

.share-row-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.share-row-name {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--main-text-color);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.share-row-path {
  font-size: 0.75rem;
  color: var(--subtle-text-color, #8a8a8a);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.share-row-meta {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 1px;
  min-width: 130px;
}

.share-row-creator,
.share-row-date {
  font-size: 0.75rem;
  color: var(--subtle-text-color, #8a8a8a);
  white-space: nowrap;
}

.share-row-expiry {
  font-size: 0.72rem;
  color: var(--subtle-text-color, #8a8a8a);
  white-space: nowrap;
}

.share-expired {
  color: var(--danger-color, #ef4444);
}

.share-row-stats {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.share-views {
  display: flex;
  align-items: center;
  gap: 3px;
  font-size: 0.78rem;
  color: var(--subtle-text-color, #8a8a8a);
}

.share-single-use {
  font-size: 0.7rem;
  background: var(--primary-color-10, rgba(99,102,241,.12));
  color: var(--primary-color, #6366f1);
  border-radius: 4px;
  padding: 1px 6px;
  white-space: nowrap;
}

.share-row-actions {
  flex-shrink: 0;
}
</style>
