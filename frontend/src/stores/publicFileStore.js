import { defineStore } from 'pinia';
import api from '../api';
import { deriveKeyFromToken, unwrapMasterKey, decryptChunkedFileWorker } from '../utils/crypto';

export const usePublicFileStore = defineStore('publicFiles', {
  state: () => ({
    files: [],
    folders: [],
    isLoading: false,
    error: null,
    currentPath: '/',
    shareToken: null,
    ownerEmail: null,
    resourceName: null,
  }),
  actions: {
    async fetchItems(token, subpath = '/') {
      if (!token) {
        this.error = "Token de partage manquant.";
        return;
      }
      this.isLoading = true;
      this.error = null;
      this.shareToken = token;
      this.currentPath = subpath;

      try {
        const response = await api.get(`/public/share/${token}/browse${subpath}`);
        this.files = response.data.files || [];
        this.folders = response.data.folders || [];
        this.ownerEmail = response.data.owner_email;
        this.resourceName = response.data.resource_name;
      } catch (error) {
        console.error('Erreur lors de la récupération des éléments partagés:', error);
        this.error = "Impossible de charger le contenu du partage. Le lien est peut-être invalide ou a expiré.";
        this.files = [];
        this.folders = [];
      } finally {
        this.isLoading = false;
      }
    },

    navigateTo(folderName) {
      const newPath = this.currentPath === '/' ? `/${folderName}` : `${this.currentPath}/${folderName}`;
      this.fetchItems(this.shareToken, newPath);
    },

    navigateUp() {
      if (this.currentPath === '/') return;
      const parts = this.currentPath.split('/').filter(p => p);
      parts.pop();
      const newPath = parts.length > 0 ? '/' + parts.join('/') : '/';
      this.fetchItems(this.shareToken, newPath);
    },

    async downloadFile(fileId, fileName) {
        const file = this.files.find(f => f.ID === fileId);
        if (!file) {
             console.error("File not found in store");
             return;
        }

        let fileKey = null;
        //console.log("Attempting to download file:", file);
        if (file.EncryptedKey) {
            try {
                const shareKey = await deriveKeyFromToken(this.shareToken);
                fileKey = await unwrapMasterKey(file.EncryptedKey, shareKey);
            } catch (e) {
                console.error("Failed to decrypt file key", e);
                alert("Erreur de déchiffrement.");
                return;
            }
        } else {
             console.error("Missing EncryptedKey for file:", file);
             alert(`Ce fichier (${file.Name}) ne peut pas être déchiffré (clé manquante). ID: ${file.ID}`);
             return;
        }

        try {
            const response = await api.get(`/public/share/${this.shareToken}/download/file/${fileId}`, {
                responseType: 'blob',
            });
            
            const decryptedBlob = await decryptChunkedFileWorker(response.data, fileKey, file.MimeType || 'application/octet-stream');

            const url = window.URL.createObjectURL(decryptedBlob);
            const link = document.createElement('a');
            link.href = url;
            link.setAttribute('download', fileName);
            document.body.appendChild(link);
            link.click();
            setTimeout(() => { link.remove(); window.URL.revokeObjectURL(url); }, 100);
        } catch (error) {
            console.error('Download error:', error);
            alert("Impossible de télécharger le fichier.");
        }
    }
  },
});
