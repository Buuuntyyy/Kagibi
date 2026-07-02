<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <Teleport to="body">
    <Transition name="panel-slide">
      <div v-if="open" class="versions-panel-overlay" @click.self="$emit('close')">
        <div class="versions-panel">

          <!-- Header -->
          <div class="panel-header">
            <div class="panel-title">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75"
                stroke-linecap="round" stroke-linejoin="round" class="panel-icon">
                <path d="M12 8v4l3 3"/>
                <circle cx="12" cy="12" r="10"/>
              </svg>
              <span>Historique des versions</span>
              <span v-if="file" class="panel-filename">— {{ file.name }}</span>
            </div>
            <button class="panel-close" @click="$emit('close')">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                stroke-linecap="round" stroke-linejoin="round" width="18" height="18">
                <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
              </svg>
            </button>
          </div>

          <!-- Body -->
          <div class="panel-body">
            <div v-if="loading" class="panel-state">Chargement…</div>

            <div v-else-if="!versions.length" class="panel-state empty">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"
                stroke-linecap="round" stroke-linejoin="round" class="empty-icon">
                <path d="M12 8v4l3 3"/><circle cx="12" cy="12" r="10"/>
              </svg>
              <p>Aucune version sauvegardée.</p>
              <p class="empty-hint">Les versions sont créées automatiquement quand vous modifiez un fichier existant.</p>
            </div>

            <div v-else class="versions-list">
              <!-- Current version info -->
              <div class="current-badge">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75"
                  stroke-linecap="round" stroke-linejoin="round" width="13" height="13">
                  <polyline points="20 6 9 17 4 12"/>
                </svg>
                Version actuelle · {{ formatBytes(file?.size ?? 0) }}
              </div>

              <div v-for="v in versions" :key="v.id" class="version-item">
                <div class="version-meta">
                  <span class="version-number">v{{ v.version_number }}</span>
                  <span class="version-date">{{ formatDate(v.created_at) }}</span>
                  <span class="version-size">{{ formatBytes(v.size) }}</span>
                </div>
                <div class="version-actions">
                  <button class="btn-action" @click="downloadVersion(v)" :disabled="busy === v.id" title="Télécharger cette version">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75"
                      stroke-linecap="round" stroke-linejoin="round" width="14" height="14">
                      <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
                      <polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/>
                    </svg>
                  </button>
                  <button class="btn-action btn-restore" @click="confirmRestore(v)"
                    :disabled="busy === v.id" title="Restaurer cette version">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75"
                      stroke-linecap="round" stroke-linejoin="round" width="14" height="14">
                      <path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/>
                      <path d="M3 3v5h5"/>
                    </svg>
                    Restaurer
                  </button>
                  <button class="btn-action btn-delete" @click="deleteVersion(v)"
                    :disabled="busy === v.id" title="Supprimer cette version">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75"
                      stroke-linecap="round" stroke-linejoin="round" width="14" height="14">
                      <polyline points="3 6 5 6 21 6"/>
                      <path d="M19 6l-1 14H6L5 6"/><path d="M10 11v6"/><path d="M14 11v6"/>
                      <path d="M9 6V4h6v2"/>
                    </svg>
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- Footer hint -->
          <div class="panel-footer" v-if="versions.length">
            <span class="footer-hint">
              {{ versions.length }} version{{ versions.length > 1 ? 's' : '' }} ·
              {{ formatBytes(totalVersionStorage) }} utilisés
            </span>
          </div>

        </div>
      </div>
    </Transition>
  </Teleport>

  <!-- Restore confirmation dialog -->
  <Teleport to="body">
    <div v-if="restoreTarget" class="confirm-overlay" @click.self="restoreTarget = null">
      <div class="confirm-modal">
        <h3>Restaurer la version v{{ restoreTarget?.version_number }} ?</h3>
        <p>Le fichier actuel sera sauvegardé comme une nouvelle version avant d'être remplacé.</p>
        <div class="confirm-actions">
          <button class="btn btn-secondary" @click="restoreTarget = null">Annuler</button>
          <button class="btn btn-primary" @click="doRestore">Restaurer</button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import api from '../../api'

const props = defineProps({
  open: Boolean,
  file: Object,
})

const emit = defineEmits(['close', 'restored'])

const versions = ref([])
const loading = ref(false)
const busy = ref(null)
const restoreTarget = ref(null)

const totalVersionStorage = computed(() =>
  versions.value.reduce((acc, v) => acc + v.size, 0)
)

watch(() => props.open, async (val) => {
  if (val && props.file) await loadVersions()
  else versions.value = []
})

async function loadVersions() {
  loading.value = true
  try {
    const { data } = await api.get(`/files/${props.file.id}/versions`)
    versions.value = data.versions ?? []
  } catch (e) {
    console.error('[FileVersionsPanel] loadVersions error:', e)
  } finally {
    loading.value = false
  }
}

async function downloadVersion(v) {
  busy.value = v.id
  try {
    const { data } = await api.get(`/files/${props.file.id}/versions/${v.id}/presigned`)
    // Reuse the same streaming download pattern as the main download flow
    const link = document.createElement('a')
    link.href = data.url
    link.download = props.file.name
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  } catch (e) {
    console.error('[FileVersionsPanel] downloadVersion error:', e)
  } finally {
    busy.value = null
  }
}

function confirmRestore(v) {
  restoreTarget.value = v
}

async function doRestore() {
  const v = restoreTarget.value
  if (!v) return
  restoreTarget.value = null
  busy.value = v.id
  try {
    await api.post(`/files/${props.file.id}/versions/${v.id}/restore`)
    await loadVersions()
    emit('restored')
  } catch (e) {
    console.error('[FileVersionsPanel] restore error:', e)
  } finally {
    busy.value = null
  }
}

async function deleteVersion(v) {
  if (!confirm(`Supprimer définitivement la version v${v.version_number} ?`)) return
  busy.value = v.id
  try {
    await api.delete(`/files/${props.file.id}/versions/${v.id}`)
    versions.value = versions.value.filter(x => x.id !== v.id)
  } catch (e) {
    console.error('[FileVersionsPanel] deleteVersion error:', e)
  } finally {
    busy.value = null
  }
}

function formatBytes(bytes) {
  const b = Number(bytes || 0)
  if (b < 1024) return `${b} B`
  if (b < 1024 ** 2) return `${(b / 1024).toFixed(1)} KB`
  if (b < 1024 ** 3) return `${(b / 1024 ** 2).toFixed(1)} MB`
  return `${(b / 1024 ** 3).toFixed(2)} GB`
}

function formatDate(iso) {
  if (!iso) return ''
  return new Date(iso).toLocaleString('fr-FR', {
    day: '2-digit', month: '2-digit', year: 'numeric',
    hour: '2-digit', minute: '2-digit',
  })
}
</script>

<style scoped>
/* Panel slide transition (matches CommentPanel) */
.panel-slide-enter-active,
.panel-slide-leave-active { transition: transform 0.22s ease, opacity 0.22s ease; }
.panel-slide-enter-from,
.panel-slide-leave-to { transform: translateX(100%); opacity: 0; }

.versions-panel-overlay {
  position: fixed;
  inset: 0;
  z-index: 40;
  background: rgba(0, 0, 0, 0.35);
  display: flex;
  justify-content: flex-end;
}

.versions-panel {
  width: 360px;
  max-width: 100%;
  height: 100%;
  background: var(--bg-surface);
  border-left: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px;
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}

.panel-title {
  display: flex;
  align-items: center;
  gap: 7px;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  min-width: 0;
}

.panel-icon { flex-shrink: 0; color: var(--accent); }

.panel-filename {
  color: var(--text-muted);
  font-weight: 400;
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.panel-close {
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  border-radius: 6px;
  padding: 4px;
  display: flex;
  align-items: center;
}
.panel-close:hover { background: var(--bg-hover); color: var(--text-primary); }

.panel-body {
  flex: 1;
  overflow-y: auto;
  padding: 12px 16px;
}

.panel-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  height: 180px;
  color: var(--text-muted);
  font-size: 13px;
  text-align: center;
}

.empty-icon { opacity: 0.35; }

.empty-hint {
  font-size: 12px;
  color: var(--text-very-dim);
  max-width: 240px;
  line-height: 1.5;
}

.current-badge {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--accent);
  font-weight: 500;
  background: color-mix(in srgb, var(--accent) 12%, transparent);
  border: 1px solid color-mix(in srgb, var(--accent) 30%, transparent);
  border-radius: 6px;
  padding: 6px 10px;
  margin-bottom: 12px;
}

.versions-list { display: flex; flex-direction: column; gap: 8px; }

.version-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  padding: 10px 12px;
  transition: border-color 0.15s;
}
.version-item:hover { border-color: var(--border); }

.version-meta {
  display: flex;
  flex-direction: column;
  gap: 3px;
  min-width: 0;
}

.version-number {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.version-date {
  font-size: 11px;
  color: var(--text-muted);
}

.version-size {
  font-size: 11px;
  color: var(--text-very-dim);
}

.version-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
}

.btn-action {
  display: flex;
  align-items: center;
  gap: 5px;
  background: var(--bg-btn-secondary);
  border: 1px solid var(--border-input);
  border-radius: 6px;
  padding: 5px 8px;
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
  cursor: pointer;
  transition: background 0.15s, color 0.15s, border-color 0.15s;
  white-space: nowrap;
}
.btn-action:disabled { opacity: 0.45; cursor: not-allowed; }
.btn-action:not(:disabled):hover { background: var(--bg-btn-secondary-h); color: var(--text-primary); }

.btn-restore:not(:disabled):hover { border-color: var(--accent); color: var(--accent); }

.btn-delete:not(:disabled):hover {
  background: rgba(239, 68, 68, 0.1);
  border-color: rgba(239, 68, 68, 0.4);
  color: #ef4444;
}

.panel-footer {
  padding: 10px 16px;
  border-top: 1px solid var(--border);
  flex-shrink: 0;
}

.footer-hint {
  font-size: 11px;
  color: var(--text-very-dim);
}

/* Restore confirm modal */
.confirm-overlay {
  position: fixed;
  inset: 0;
  z-index: 60;
  background: rgba(0,0,0,0.5);
  display: flex;
  align-items: center;
  justify-content: center;
}

.confirm-modal {
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 24px;
  max-width: 380px;
  width: 100%;
  box-shadow: 0 8px 32px rgba(0,0,0,0.3);
}

.confirm-modal h3 {
  margin: 0 0 10px;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.confirm-modal p {
  margin: 0 0 20px;
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.5;
}

.confirm-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.btn { padding: 8px 16px; border: none; border-radius: 7px; font-size: 13px; font-weight: 500; cursor: pointer; transition: opacity 0.15s; }
.btn-secondary { background: var(--bg-btn-secondary); color: var(--text-secondary); border: 1px solid var(--border); }
.btn-secondary:hover { background: var(--bg-btn-secondary-h); }
.btn-primary { background: var(--accent); color: #fff; }
.btn-primary:hover { opacity: 0.88; }
</style>
