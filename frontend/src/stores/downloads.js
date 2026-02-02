/**
 * Download Store - Pinia store for multi-file download state
 */
import { defineStore } from 'pinia'
import { downloadManager, DownloadStatus } from '../utils/zipDownloadManager'

export { DownloadStatus }

export const useDownloadStore = defineStore('downloads', {
  state: () => ({
    // Current download session
    isActive: false,
    status: DownloadStatus.IDLE,
    downloadName: '',  // File name or ZIP name
    isSingleFile: false,  // True for single file download (no ZIP)
    
    // Progress tracking
    totalFiles: 0,
    processedFiles: 0,
    totalSize: 0,
    bytesDownloaded: 0,
    percent: 0,
    speed: 0, // bytes/sec
    eta: 0, // seconds remaining
    
    // Individual file states
    files: [],
    
    // UI state
    showManager: false,
    minimized: false,
    
    // Error handling
    error: null
  }),
  
  getters: {
    /**
     * Human readable speed
     */
    formattedSpeed: (state) => {
      if (state.speed <= 0) return '0 B/s'
      const units = ['B/s', 'KB/s', 'MB/s', 'GB/s']
      let speed = state.speed
      let unitIndex = 0
      while (speed >= 1024 && unitIndex < units.length - 1) {
        speed /= 1024
        unitIndex++
      }
      return `${speed.toFixed(1)} ${units[unitIndex]}`
    },
    
    /**
     * Human readable ETA
     */
    formattedEta: (state) => {
      if (state.eta <= 0) return '--'
      const minutes = Math.floor(state.eta / 60)
      const seconds = state.eta % 60
      if (minutes > 60) {
        const hours = Math.floor(minutes / 60)
        const mins = minutes % 60
        return `${hours}h ${mins}m`
      }
      if (minutes > 0) {
        return `${minutes}m ${seconds}s`
      }
      return `${seconds}s`
    },
    
    /**
     * Human readable total size
     */
    formattedTotalSize: (state) => {
      return formatBytes(state.totalSize)
    },
    
    /**
     * Human readable downloaded size
     */
    formattedDownloaded: (state) => {
      return formatBytes(state.bytesDownloaded)
    },
    
    /**
     * Status text for UI
     */
    statusText: (state) => {
      const statusTexts = {
        [DownloadStatus.IDLE]: 'En attente',
        [DownloadStatus.INITIALIZING]: 'Initialisation...',
        [DownloadStatus.FETCHING_TREE]: 'Récupération de la structure...',
        [DownloadStatus.GENERATING_URLS]: 'Génération des URLs...',
        [DownloadStatus.DOWNLOADING]: 'Téléchargement en cours...',
        [DownloadStatus.DECRYPTING]: 'Déchiffrement...',
        [DownloadStatus.FINALIZING]: state.isSingleFile ? 'Finalisation...' : 'Finalisation du ZIP...',
        [DownloadStatus.COMPLETED]: 'Terminé',
        [DownloadStatus.ERROR]: 'Erreur',
        [DownloadStatus.ABORTED]: 'Annulé'
      }
      return statusTexts[state.status] || state.status
    },
    
    /**
     * Check if download can be cancelled
     */
    canCancel: (state) => {
      return [
        DownloadStatus.INITIALIZING,
        DownloadStatus.FETCHING_TREE,
        DownloadStatus.GENERATING_URLS,
        DownloadStatus.DOWNLOADING
      ].includes(state.status)
    },
    
    /**
     * Check if download is in progress
     */
    isInProgress: (state) => {
      return ![
        DownloadStatus.IDLE,
        DownloadStatus.COMPLETED,
        DownloadStatus.ERROR,
        DownloadStatus.ABORTED
      ].includes(state.status)
    }
  },
  
  actions: {
    /**
     * Initialize the download manager
     */
    async init() {
      await downloadManager.init({
        onProgress: this.handleProgress.bind(this),
        onStatusChange: this.handleStatusChange.bind(this),
        onError: this.handleError.bind(this),
        onComplete: this.handleComplete.bind(this)
      })
    },
    
    /**
     * Download a folder as ZIP
     */
    async downloadFolder(folderId, folderName) {
      this.reset()
      this.isActive = true
      this.showManager = true
      this.downloadName = `${folderName}.zip`
      this.isSingleFile = false
      
      await downloadManager.downloadFolder(folderId, folderName)
    },
    
    /**
     * Download selected items as ZIP
     */
    async downloadSelection(fileIds, folderIds, zipName = 'selection.zip') {
      this.reset()
      this.isActive = true
      this.showManager = true
      this.downloadName = zipName
      this.isSingleFile = false
      
      await downloadManager.downloadSelection(fileIds, folderIds, zipName)
    },

    /**
     * Download a single file with progress tracking (no ZIP)
     */
    async downloadSingleFile(fileId, fileName, encryptedKey = null, fileSize = 0) {
      this.reset()
      this.isActive = true
      this.showManager = true
      this.downloadName = fileName
      this.isSingleFile = true
      
      await downloadManager.downloadSingleFile(fileId, fileName, encryptedKey, fileSize)
    },

    /**
     * Cancel current download
     */
    cancel() {
      downloadManager.abort()
      this.status = DownloadStatus.ABORTED
    },
    
    /**
     * Handle progress updates from manager
     * Progress is now bytes-based for smooth granular updates
     */
    handleProgress(progress) {
      this.totalFiles = progress.totalFiles
      this.processedFiles = progress.processedFiles
      this.totalSize = progress.totalSize
      this.bytesDownloaded = progress.bytesDownloaded
      this.percent = progress.percent
      this.speed = progress.speed
      this.eta = progress.eta
      this.isSingleFile = progress.isSingleFile ?? this.isSingleFile
      this.files = progress.files || []
    },
    
    /**
     * Handle status changes
     */
    handleStatusChange(status) {
      this.status = status
    },
    
    /**
     * Handle errors
     */
    handleError(errorMessage) {
      this.error = errorMessage
      this.status = DownloadStatus.ERROR
    },
    
    /**
     * Handle completion
     */
    handleComplete(result) {
      this.status = DownloadStatus.COMPLETED
      
      // Auto-hide after delay
      setTimeout(() => {
        if (this.status === DownloadStatus.COMPLETED) {
          this.minimized = true
        }
      }, 3000)
    },
    
    /**
     * Toggle minimized state
     */
    toggleMinimize() {
      this.minimized = !this.minimized
    },
    
    /**
     * Close the download manager UI
     */
    close() {
      if (!this.isInProgress) {
        this.showManager = false
        this.reset()
      }
    },
    
    /**
     * Reset state
     */
    reset() {
      this.isActive = false
      this.status = DownloadStatus.IDLE
      this.downloadName = ''
      this.isSingleFile = false
      this.totalFiles = 0
      this.processedFiles = 0
      this.totalSize = 0
      this.bytesDownloaded = 0
      this.percent = 0
      this.speed = 0
      this.eta = 0
      this.files = []
      this.error = null
      this.minimized = false
    }
  }
})

/**
 * Format bytes to human readable string
 */
function formatBytes(bytes) {
  if (bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const k = 1024
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + units[i]
}
