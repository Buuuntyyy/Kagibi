import { defineStore } from 'pinia'
import api from '../api'
import { useAuthStore } from './auth'
import { encryptFile, decryptFile } from '../utils/crypto'

export const useFileStore = defineStore('files', {
  state: () => ({
    files: [],
    folders: [],
    currentPath: '/',
  }),
  actions: {
    async fetchItems(path) {
      try {
        const safePath = path.startsWith('/') ? path : `/${path}`
        const response = await api.get(`/files/list${safePath}`)
        this.files = response.data.files || []
        this.folders = response.data.folders || []
        this.currentPath = safePath
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
      const authStore = useAuthStore()
      if (!authStore.masterKey) {
        console.error('Cipher key not available. Cannot download file.')
        return
      }
      try {
        // Téléchargement du blob chiffré depuis le serveur
        const response = await api.get(`/files/download/${fileId}`, {
          responseType: 'blob', // Important pour recevoir le fichier
        })

        const decryptedBlob = await decryptFile(response.data, authStore.masterKey, MimeType);

        const url = window.URL.createObjectURL(decryptedBlob)
        const link = document.createElement('a')
        link.href = url;
        link.setAttribute('download', fileName) // Nom du fichier
        document.body.appendChild(link)
        link.click()
        
        setTimeout(() => {
          link.remove()
          window.URL.revokeObjectURL(url)
        }, 100);

      } catch (error) {
        console.error("Erreur download:", error);
        if (error.message && error.message.includes("déchiffrement")) {
            alert("Erreur d'intégrité : Le fichier est corrompu ou votre mot de passe est incorrect.");
        } else {
            alert("Impossible de télécharger le fichier.");
        }
      }
    },
    async uploadFile(file) {
      const authStore = useAuthStore()
      if (!authStore.isAuthenticated || !authStore.masterKey) {
        console.error('User not authenticated. Cannot upload file.')
        return
      }
      const masterKey = authStore.masterKey;
      try {
        const encryptedBlob = await encryptFile(file, masterKey);
        const encryptedFile = new File([encryptedBlob], file.name, { type: 'application/octet-stream' });
        
        //Création d'un fichier virtuel chiffré pour l'envoie (nom original conservé)
        const formData = new FormData()
        formData.append('file', encryptedFile)
        formData.append('path', this.currentPath)

        //Envoie au serveur
        await api.post('/files/upload', formData, {
          headers: {
            'Content-Type': 'multipart/form-data',
          },
        })

        // Rafraîchir la liste des fichiers après l'upload
        this.fetchItems(this.currentPath)

      } catch (error) {
        console.error("Erreur upload:", error);
        if (error.message && error.message.includes("chiffrement")) {
            alert("Erreur critique : Impossible de chiffrer le fichier localement.");
        } else {
            alert("Erreur lors de l'envoi du fichier au serveur.");
        }
      }
    },
    async createFolder(folderName) {
      try {
        await api.post('/folders/create', {
          name: folderName,
          path: this.currentPath,
        })
        this.fetchItems(this.currentPath)
      } catch (error) {
        console.error('Error creating folder:', error)
      }
    },
    async deleteFiles(fileIDs) {
    try {
      await api.post('/files/bulk-delete', { file_ids: fileIDs })
      this.fetchItems(this.currentPath)
    } catch (error) {
      console.error('Error deleting items:', error)
    }
  }
  },
})