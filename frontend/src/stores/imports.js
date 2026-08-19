// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

/**
 * Import Store — Pinia store for background cloud-import transfer state
 * (Google Drive / OneDrive / Dropbox). Single active session, same shape and
 * spirit as useDownloadStore: lets a migration survive the import dialog
 * being closed or the user navigating elsewhere, driven by importQueueManager.
 */
import { defineStore } from 'pinia'
import { importQueueManager } from '../utils/importQueueManager'

export const useImportStore = defineStore('imports', {
  state: () => ({
    isActive: false,
    namespace: '',          // 'gdImport' | 'odImport' | 'dbImport' — reuses each
                             // provider's existing locale keys (progressTitle,
                             // abort, doneSuccess, ...)

    phase: 'progress',      // 'progress' | 'done'
    progressTotal: 0,
    progressDone: 0,
    errorCount: 0,
    importedBytes: 0,
    totalImportBytes: 0,
    folderPhaseDone: 0,
    folderPhaseTotal: 0,

    log: [],                 // [{ id, name, status: 'importing'|'done'|'error', error? }]

    error: null,

    showManager: false,
    minimized: false,
  }),

  getters: {
    isInProgress: (state) => state.isActive && state.phase === 'progress',

    progressPercent: (state) =>
      state.progressTotal > 0 ? Math.round((state.progressDone / state.progressTotal) * 100) : 0,

    bytesPercent: (state) =>
      state.totalImportBytes > 0
        ? Math.min(Math.round((state.importedBytes / state.totalImportBytes) * 100), 100)
        : 0,

    folderPhasePercent: (state) =>
      state.folderPhaseTotal > 0
        ? Math.min(Math.round((state.folderPhaseDone / state.folderPhaseTotal) * 100), 100)
        : 0,
  },

  actions: {
    // Wires this store's handlers into the singleton manager. Idempotent —
    // safe to call from the widget's onMounted every time it (re)mounts.
    init() {
      importQueueManager.init({
        onTotal: this.handleTotal.bind(this),
        onFileStart: this.handleFileStart.bind(this),
        onFileDone: this.handleFileDone.bind(this),
        onFileError: this.handleFileError.bind(this),
        onBytesProgress: this.handleBytesProgress.bind(this),
        onFolderProgress: this.handleFolderProgress.bind(this),
        onDone: this.handleDone.bind(this),
        onFatalError: this.handleFatalError.bind(this),
      })
    },

    // Starts a background import. Returns false (and leaves existing state
    // untouched) if another import is already running — only one migration
    // can be active at a time, same constraint as the download manager.
    start(opts) {
      if (this.isActive) return false

      this.reset()
      this.isActive = true
      this.showManager = true
      this.namespace = opts.namespace
      this.phase = 'progress'

      importQueueManager.start(opts)
      return true
    },

    abort() {
      importQueueManager.abort()
    },

    toggleMinimize() {
      this.minimized = !this.minimized
    },

    close() {
      if (!this.isInProgress) {
        this.showManager = false
        this.reset()
      }
    },

    reset() {
      this.isActive = false
      this.namespace = ''
      this.phase = 'progress'
      this.progressTotal = 0
      this.progressDone = 0
      this.errorCount = 0
      this.importedBytes = 0
      this.totalImportBytes = 0
      this.folderPhaseDone = 0
      this.folderPhaseTotal = 0
      this.log = []
      this.error = null
      this.minimized = false
    },

    // ── Handlers matching importItems()'s callback contract ──

    handleTotal(n) {
      this.progressTotal = n
    },

    handleFileStart(name) {
      this.log.push({ id: crypto.randomUUID(), name, status: 'importing' })
    },

    handleFileDone(name) {
      this.progressDone++
      const entry = [...this.log].reverse().find(e => e.name === name && e.status === 'importing')
      if (entry) entry.status = 'done'
    },

    handleFileError(name, msg) {
      this.progressDone++
      this.errorCount++
      const entry = [...this.log].reverse().find(e => e.name === name && e.status === 'importing')
      if (entry) { entry.status = 'error'; entry.error = msg }
    },

    handleBytesProgress(uploaded, total) {
      this.importedBytes = uploaded
      this.totalImportBytes = total
    },

    handleFolderProgress(done, total) {
      this.folderPhaseDone = done
      this.folderPhaseTotal = total
    },

    // The label ("Erreur critique de l'import") is rendered by the widget via
    // t(`${namespace}.fatalError`) — the store only carries the raw detail.
    handleFatalError(message) {
      this.error = message
    },

    handleDone() {
      this.phase = 'done'

      // Auto-minimize after a delay, same as the download manager.
      setTimeout(() => {
        if (this.phase === 'done') {
          this.minimized = true
        }
      }, 3000)
    },
  },
})
