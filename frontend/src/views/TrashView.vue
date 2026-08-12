<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<!-- Corbeille personnelle — même modèle que l'onglet corbeille des organisations. -->
<template>
  <div class="trash-view">
    <div class="section-header">
      <h3>{{ t('trash.title') }}</h3>
      <div class="section-header-actions">
        <button class="btn-sm" @click="loadTrash" :disabled="fileStore.trashLoading">
          <span v-if="fileStore.trashLoading" class="spinner-sm"></span>
          <span v-else>{{ t('trash.refresh') }}</span>
        </button>
        <button v-if="fileStore.trash.length > 0" class="btn-sm btn-danger-sm" @click="handleEmptyTrash">
          {{ t('trash.emptyTrash') }}
        </button>
      </div>
    </div>

    <div v-if="fileStore.trashLoading && fileStore.trash.length === 0" class="loading-center" style="padding:40px 0">
      <div class="spinner"></div>
    </div>

    <div v-else-if="fileStore.trash.length === 0" class="empty-tab">
      <svg viewBox="0 0 24 24" width="40" height="40" fill="currentColor" style="opacity:.3"><path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"/></svg>
      <p>{{ t('trash.empty') }}</p>
    </div>

    <div v-else class="trash-list">
      <div v-for="item in fileStore.trash" :key="item.item_type + item.id" class="trash-row">
        <div class="trash-row-icon">
          <svg v-if="item.item_type === 'folder'" viewBox="0 0 24 24" width="18" height="18" fill="currentColor"><path d="M10 4H4c-1.11 0-2 .89-2 2L2 18c0 1.11.89 2 2 2h16c1.11 0 2-.89 2-2V8c0-1.11-.89-2-2-2h-8l-2-2z"/></svg>
          <svg v-else viewBox="0 0 24 24" width="18" height="18" fill="currentColor"><path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z"/></svg>
        </div>
        <div class="trash-row-info">
          <span class="trash-row-name">{{ item._name || item.name }}</span>
          <span class="trash-row-path">{{ item._path || item.path }}</span>
        </div>
        <div class="trash-row-meta">
          <span class="trash-row-date" :title="item.deleted_at">{{ t('trash.deletedOn', { date: formatTrashDate(item.deleted_at) }) }}</span>
          <span v-if="item.item_type === 'file' && item.size" class="trash-row-by">{{ formatSize(item.size) }}</span>
        </div>
        <div class="trash-row-actions">
          <button class="btn-sm" @click="handleRestore(item)" :title="t('trash.restore')">{{ t('trash.restore') }}</button>
          <button class="btn-sm btn-danger-sm" @click="handlePermanentDelete(item)" :title="t('trash.permanentDelete')">{{ t('trash.permanentDelete') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useFileStore } from '../stores/files'
import { useUIStore } from '../stores/ui'
import { formatSize, formatDateOnly } from '../utils/format'

const { t } = useI18n()
const fileStore = useFileStore()
const uiStore = useUIStore()

onMounted(loadTrash)

async function loadTrash() {
  try { await fileStore.fetchTrash() } catch (_) {}
}

const formatTrashDate = formatDateOnly

async function handleRestore(item) {
  try {
    await fileStore.restoreTrashItem(item)
    uiStore.showToast(t('trash.itemRestored'))
  } catch (e) {
    uiStore.showToast(e.response?.data?.error || e.message, 'error')
  }
}

async function handlePermanentDelete(item) {
  const confirmed = await uiStore.showConfirm({
    title: t('trash.permanentDelete'),
    message: t('trash.confirmPermanentDelete', { name: item._name || item.name }),
    confirmLabel: t('trash.permanentDelete'),
  })
  if (!confirmed) return
  try {
    await fileStore.permanentDeleteTrashItem(item)
    uiStore.showToast(t('trash.permanentDeleted'))
  } catch (e) {
    uiStore.showToast(e.response?.data?.error || e.message, 'error')
  }
}

async function handleEmptyTrash() {
  const confirmed = await uiStore.showConfirm({
    title: t('trash.emptyTrash'),
    message: t('trash.confirmEmptyTrash'),
    confirmLabel: t('trash.emptyTrash'),
  })
  if (!confirmed) return
  try {
    await fileStore.emptyTrash()
    uiStore.showToast(t('trash.trashEmptied'))
  } catch (e) {
    uiStore.showToast(e.response?.data?.error || e.message, 'error')
  }
}
</script>

<style scoped>
.trash-view {
  width: 100%;
  height: 100%;
  box-sizing: border-box;
  padding: 20px 24px;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.section-header h3 {
  margin: 0;
  font-size: 1.1rem;
  color: var(--main-text-color);
}

.section-header-actions {
  display: flex;
  gap: 8px;
}

.btn-sm {
  padding: 5px 12px;
  font-size: 0.8rem;
  border-radius: 6px;
  border: 1px solid var(--border-color);
  background: var(--card-color);
  color: var(--main-text-color);
  cursor: pointer;
  transition: background 0.12s;
}

.btn-sm:hover:not(:disabled) {
  background: var(--hover-background-color);
}

.btn-sm:disabled {
  opacity: 0.6;
  cursor: default;
}

.btn-danger-sm {
  color: #e05555;
  border-color: rgba(224, 85, 85, 0.4);
}

.btn-danger-sm:hover:not(:disabled) {
  background: rgba(224, 85, 85, 0.08);
}

.loading-center {
  display: flex;
  justify-content: center;
}

.empty-tab {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  padding: 48px 0;
  color: var(--subtle-text-color, #8a8a8a);
}

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

.spinner,
.spinner-sm {
  border: 2px solid var(--border-color);
  border-top-color: var(--main-text-color);
  border-radius: 50%;
  animation: trash-spin 0.7s linear infinite;
}

.spinner {
  width: 22px;
  height: 22px;
}

.spinner-sm {
  display: inline-block;
  width: 12px;
  height: 12px;
  vertical-align: middle;
}

@keyframes trash-spin {
  to { transform: rotate(360deg); }
}

@media (max-width: 640px) {
  .trash-view {
    padding: 14px 12px;
  }

  .trash-row-meta {
    display: none;
  }
}
</style>
