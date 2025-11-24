import { defineStore } from 'pinia'
import api from '../api'

export const useFileStore = defineStore('files', {
  state: () => ({
    files: [],
    folders: [],
    currentPath: '/',
  }),
  actions: {
    async fetchItems(path) {
      try {
        const response = await api.get(`/files/list${path}`)
        this.files = response.data.files || []
        this.folders = response.data.folders || []
        this.currentPath = path
      } catch (error) {
        console.error('Error fetching items:', error)
      }
    },
    navigateTo(folderName) {
        let newPath = this.currentPath
        if (newPath.endsWith('/')) {
            newPath += folderName
        } else {
            newPath += `/${folderName}`
        }
        this.fetchItems(newPath)
    },
    navigateUp() {
        if (this.currentPath === '/') return
        const parts = this.currentPath.split('/').filter(p => p)
        parts.pop()
        const newPath = '/' + parts.join('/')
        this.fetchItems(newPath)
    },
    async downloadFile(fileId, fileName) {
      try {
        const response = await api.get(`/files/download/${fileId}`, {
          responseType: 'blob', // Important pour recevoir le fichier
        })

        // Créer un lien temporaire pour déclencher le téléchargement
        const url = window.URL.createObjectURL(new Blob([response.data]))
        const link = document.createElement('a')
        link.href = url
        link.setAttribute('download', fileName) // Nom du fichier
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
      } catch (error) {
        console.error('Error downloading file:', error)
      }
    },
    async uploadFile(file) {
      try {
        const formData = new FormData()
        formData.append('files', file)
        formData.append('path', this.currentPath)

        await api.post('/files/upload', formData, {
          headers: {
            'Content-Type': 'multipart/form-data',
          },
        })

        // Rafraîchir la liste des fichiers après l'upload
        this.fetchItems(this.currentPath)
      } catch (error) {
        console.error('Error uploading file:', error)
      }
    }
  },
})