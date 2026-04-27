import { defineStore } from 'pinia';
import api from '../api';
import { deriveKeyFromToken, unwrapMasterKey, decryptChunkedFileWorker, generateMasterKey, wrapMasterKey, encryptChunkWorker } from '../utils/crypto';
import { PART_SIZE } from '../utils/multipartUpload';

export const usePublicFileStore = defineStore('publicFiles', {
  state: () => ({
    files: [],
    folders: [],
    isLoading: false,
    error: null,
    currentPath: '/',
    shareToken: null,
    ownerEmail: null,
    ownerName: null,
    resourceName: null,
    permissions: { download: true, create: false, delete: false, move: false },
    toast: { visible: false, message: '', type: 'error' },
    isUploading: false,
    uploadProgress: 0,
    uploadingFileName: null,
  }),
  actions: {
    showToast(message, type = 'error') {
      this.toast = { visible: true, message, type };
      setTimeout(() => { this.toast.visible = false; }, 3500);
    },

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
        this.ownerName = response.data.owner_name;
        this.resourceName = response.data.resource_name;
        this.permissions = response.data.permissions || { download: true, create: false, delete: false, move: false };
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

    async deleteFile(fileId) {
      try {
        await api.delete(`/public/share/${this.shareToken}/file/${fileId}`);
        this.files = this.files.filter(f => f.ID !== fileId);
      } catch (error) {
        console.error('Delete error:', error);
        throw error;
      }
    },

    async deleteFolder(folderId) {
      try {
        await api.delete(`/public/share/${this.shareToken}/folder/${folderId}`);
        this.folders = this.folders.filter(f => f.ID !== folderId);
      } catch (error) {
        console.error('Delete folder error:', error);
        throw error;
      }
    },

    async createFolder(name) {
      if (!this.permissions.create) {
        this.showToast("Vous n'avez pas l'autorisation de créer des dossiers dans ce partage.");
        return;
      }
      try {
        const response = await api.post(`/public/share/${this.shareToken}/folder`, {
          name,
          parent_path: this.currentPath,
        });
        this.folders = [...this.folders, { ID: response.data.id, Name: response.data.name, Path: response.data.path }];
      } catch (e) {
        console.error('Create folder error:', e);
        this.showToast("Impossible de créer le dossier.");
        throw e;
      }
    },

    async renameItem(id, type, newName) {
      if (!this.permissions.move) {
        this.showToast("Vous n'avez pas l'autorisation de renommer dans ce partage.");
        return;
      }
      try {
        await api.post(`/public/share/${this.shareToken}/rename`, { id, type, new_name: newName });
        if (type === 'file') {
          const f = this.files.find(f => f.ID === id);
          if (f) f.Name = newName;
        } else {
          const f = this.folders.find(f => f.ID === id);
          if (f) f.Name = newName;
        }
      } catch (e) {
        console.error('Rename error:', e);
        this.showToast("Impossible de renommer l'élément.");
        throw e;
      }
    },

    async downloadFile(fileId, fileName) {
        if (!this.permissions.download) {
            this.showToast("Vous n'avez pas l'autorisation de télécharger des fichiers depuis ce partage.");
            return;
        }

        const file = this.files.find(f => f.ID === fileId);
        if (file && file.can_download === false) {
            this.showToast("Le téléchargement de ce fichier n'est pas autorisé.");
            return;
        }
        if (!file) {
             console.error("File not found in store");
             return;
        }

        let fileKey = null;
        if (file.EncryptedKey) {
            try {
                const shareKey = await deriveKeyFromToken(this.shareToken);
                fileKey = await unwrapMasterKey(file.EncryptedKey, shareKey);
            } catch (e) {
                console.error("Failed to decrypt file key", e);
                this.showToast("Erreur de déchiffrement du fichier.");
                return;
            }
        } else {
             console.error("Missing EncryptedKey for file:", file);
             this.showToast(`Ce fichier (${file.Name}) ne peut pas être déchiffré (clé manquante).`);
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
            this.showToast("Impossible de télécharger le fichier.");
        }
    },

    async uploadFiles(fileList) {
        if (!this.permissions.create) {
            this.showToast("Vous n'avez pas l'autorisation d'ajouter des fichiers dans ce partage.");
            return;
        }
        const files = Array.from(fileList);
        if (files.length === 0) return;

        this.isUploading = true;
        this.uploadProgress = 0;

        try {
            const tokenKey = await deriveKeyFromToken(this.shareToken);

            for (let fi = 0; fi < files.length; fi++) {
                const file = files[fi];
                this.uploadingFileName = file.name;

                // Encrypt chunks
                const fileKey = await generateMasterKey();
                const totalParts = Math.max(1, Math.ceil(file.size / PART_SIZE));
                const encryptedChunks = [];
                let offset = 0;
                let chunkIndex = 0;
                while (offset < file.size) {
                    const chunkBlob = file.slice(offset, offset + PART_SIZE);
                    const chunkBuf = await chunkBlob.arrayBuffer();
                    const encChunk = await encryptChunkWorker(chunkBuf, fileKey, chunkIndex);
                    encryptedChunks.push(encChunk);
                    offset += PART_SIZE;
                    chunkIndex++;
                }
                const totalEncryptedSize = encryptedChunks.reduce((s, c) => s + (c.size || c.byteLength || 0), 0);
                const encryptedFileKey = await wrapMasterKey(fileKey, tokenKey);

                // Initiate
                const initiateRes = await api.post(`/public/share/${this.shareToken}/multipart/initiate`, {
                    file_name: file.name,
                    file_path: this.currentPath,
                    content_type: 'application/octet-stream',
                    total_size: totalEncryptedSize,
                    total_parts: encryptedChunks.length,
                    encrypted_key: encryptedFileKey,
                });
                const { upload_id: uploadId, key: s3Key, presigned_urls: presignedURLs } = initiateRes.data;

                // Upload parts
                const completedParts = [];
                for (let i = 0; i < presignedURLs.length; i++) {
                    const resp = await fetch(presignedURLs[i].url, { method: 'PUT', body: encryptedChunks[i] });
                    if (!resp.ok) {
                        await api.post(`/public/share/${this.shareToken}/multipart/abort`, { upload_id: uploadId, key: s3Key }).catch(() => {});
                        throw new Error(`Part ${i + 1} upload failed (HTTP ${resp.status})`);
                    }
                    const etag = resp.headers.get('ETag') || '';
                    completedParts.push({ part_number: i + 1, etag });
                    this.uploadProgress = Math.round(((fi + (i + 1) / presignedURLs.length) / files.length) * 100);
                }

                // Complete
                await api.post(`/public/share/${this.shareToken}/multipart/complete`, {
                    upload_id: uploadId,
                    key: s3Key,
                    parts: completedParts,
                    file_name: file.name,
                    file_path: this.currentPath,
                    total_size: totalEncryptedSize,
                    content_type: 'application/octet-stream',
                    encrypted_key: encryptedFileKey,
                });
            }

            this.uploadProgress = 100;
            this.showToast(files.length === 1 ? 'Fichier ajouté avec succès.' : `${files.length} fichiers ajoutés avec succès.`, 'success');
            await this.fetchItems(this.shareToken, this.currentPath);
        } catch (error) {
            console.error('Public share upload error:', error);
            this.showToast('Erreur lors de l\'envoi du fichier : ' + error.message);
        } finally {
            this.isUploading = false;
            this.uploadProgress = 0;
            this.uploadingFileName = null;
        }
    },
  },
});
