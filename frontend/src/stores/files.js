import { defineStore } from 'pinia'
import api from '../api'
import { useAuthStore } from './auth'
import { encryptFile, decryptFile } from '../utils/crypto'
import { encryptChunkWorker, decryptChunkedFileWorker, CHUNK_SIZE } from '../utils/crypto'

export const useFileStore = defineStore('files', {
  state: () => ({
    files: [],
    folders: [],
    currentPath: '/',
    uploadProgress: 0,
    isUploading: false,
    uploadingFileName: '',
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
    async downloadFile(fileId, fileName, mimeType='application/octet-stream') {
      const authStore = useAuthStore();
      if (!authStore.masterKey) return;

      try {
        // 1. Télécharger le blob chiffré
        const response = await api.get(`/files/download/${fileId}`, { responseType: 'blob' });
        
        // 2. Déchiffrer via Worker
        const decryptedBlob = await decryptChunkedFileWorker(response.data, authStore.masterKey, mimeType);

        // 3. Sauvegarder
        const url = window.URL.createObjectURL(decryptedBlob);
        const link = document.createElement('a');
        link.href = url;
        link.setAttribute('download', fileName);
        document.body.appendChild(link);
        link.click();
        setTimeout(() => { link.remove(); window.URL.revokeObjectURL(url); }, 100);

      } catch (error) {
        console.error("Erreur download:", error);
        alert("Erreur lors du téléchargement.");
      }
    },
    async uploadFile(file) {
      const authStore = useAuthStore()
      if (!authStore.isAuthenticated || !authStore.masterKey) {
        console.error('User not authenticated. Cannot upload file.')
        return
      }

      const totalChunks = Math.ceil(file.size / CHUNK_SIZE);
      let chunkIndex = 0;
      let offset = 0;
      this.isUploading = true;
      this.uploadProgress = 0;
      this.uploadingFileName = file.name;

      try {
        while(offset < file.size) {
          const chunkBlob = file.slice(offset, offset + CHUNK_SIZE);
          const chunkArrayBuffer = await chunkBlob.arrayBuffer();

          const encryptedChunkBlob = await encryptChunkWorker(chunkArrayBuffer, authStore.masterKey, chunkIndex);

          const encryptedFile = new File([encryptedChunkBlob], file.name, { type: 'application/octet-stream' });

          const formData = new FormData()
          formData.append('file', encryptedFile)
          formData.append('path', this.currentPath)
          formData.append('chunk_index', chunkIndex)
          formData.append('total_chunks', totalChunks)

          console.log(`Uploading chunk ${chunkIndex + 1} / ${totalChunks}...`);

          await api.post('/files/upload', formData, {
            headers: {
              'Content-Type': 'multipart/form-data',
            },
          });

          offset += CHUNK_SIZE;
          chunkIndex += 1;

          this.uploadProgress = Math.round((chunkIndex / totalChunks) * 100);
          console.log(`Uploaded chunk ${chunkIndex} / ${totalChunks}`);
        } 

        this.fetchItems(this.currentPath)
        
        // Reset progress after a short delay
        setTimeout(() => {
            this.isUploading = false;
            this.uploadProgress = 0;
        }, 1000);

      } catch (error) {
        console.error("Erreur upload:", error);
        alert("Erreur lors de l'envoi du fichier au serveur.");
        this.isUploading = false;
        this.uploadProgress = 0;
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
  },
    async moveItems(items, destinationPath) {
      try {
        const promises = items.map(item => api.post('/files/move', {
          id: item.id,
          type: item.type,
          destinationPath: destinationPath
        }))
        
        await Promise.all(promises)
        this.fetchItems(this.currentPath)
      } catch (error) {
        console.error('Error moving items:', error)
        alert("Erreur lors du déplacement de certains éléments.")
        this.fetchItems(this.currentPath)
      }
    },
    async moveItem(itemId, type, destinationPath) {
      await this.moveItems([{ id: itemId, type: type }], destinationPath)
    }
  },
})