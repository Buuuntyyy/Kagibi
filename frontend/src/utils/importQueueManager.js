// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

/**
 * Import Queue Manager — singleton orchestrating a background cloud-import
 * transfer (Google Drive / OneDrive / Dropbox), mirroring zipDownloadManager's
 * shape. Consolidates what used to be duplicated inside each import dialog's
 * doImport(): dedicated-folder creation, the heartbeat/file-list refresh
 * around the transfer, and forwarding importItems()'s callback contract to
 * whatever store (useImportStore) called init().
 */
import api from '../api'
import { useFileStore } from '../stores/files'

class ImportQueueManager {
  constructor() {
    this._callbacks = null
    this._importer = null
  }

  // Wires the store's handlers. Safe to call more than once (idempotent) —
  // the widget calls this from onMounted every time it (re)mounts.
  init(callbacks) {
    this._callbacks = callbacks
  }

  /**
   * @param {Object} opts
   * @param {string}   opts.namespace            locale namespace ('gdImport' | 'odImport' | 'dbImport')
   * @param {Object}   opts.importer              GoogleDriveImport | OneDriveImport | DropboxImport instance
   * @param {Array}    opts.filesToImport         selected file items (deduped by the dialog)
   * @param {Map}      opts.pathMap               providerId → kagibi path (from buildPathMap)
   * @param {boolean}  opts.useDedicatedFolder
   * @param {string}   opts.dedicatedFolderName
   */
  async start({ importer, filesToImport, pathMap, useDedicatedFolder, dedicatedFolderName }) {
    this._importer = importer

    const fileStore = useFileStore()
    fileStore.startHeartbeat()

    // Create the dedicated root folder if requested, then derive the target path —
    // identical to what each dialog's doImport() did locally before this refactor.
    let targetPath = '/'
    if (useDedicatedFolder && dedicatedFolderName) {
      try {
        await api.post('/folders/create', { name: dedicatedFolderName, path: '/' })
      } catch (err) {
        if (err?.response?.status !== 409) {
          // 409 = already exists, perfectly fine; anything else is a real error
          const msg = err?.response?.data?.error ?? err.message ?? String(err)
          fileStore.stopHeartbeat()
          this._callbacks?.onFatalError?.(msg)
          this._callbacks?.onDone?.()
          return
        }
      }
      targetPath = '/' + dedicatedFolderName
    }

    try {
      await importer.importItems(filesToImport, pathMap, targetPath, {
        onTotal: (n) => this._callbacks?.onTotal?.(n),
        onFileStart: (name) => this._callbacks?.onFileStart?.(name),
        onFileDone: (name) => this._callbacks?.onFileDone?.(name),
        onFileError: (name, msg) => this._callbacks?.onFileError?.(name, msg),
        onBytesProgress: (uploaded, total) => this._callbacks?.onBytesProgress?.(uploaded, total),
        onFolderProgress: (done, total) => this._callbacks?.onFolderProgress?.(done, total),
      })
    } catch (err) {
      this._callbacks?.onFatalError?.(err.message ?? String(err))
    } finally {
      fileStore.stopHeartbeat()
    }

    fileStore.fetchItems(fileStore.currentPath)
    this._callbacks?.onDone?.()
  }

  abort() {
    this._importer?.abort()
  }
}

export const importQueueManager = new ImportQueueManager()
export default importQueueManager
