import { defineStore } from 'pinia'
import api from '../api'
import { useAuthStore } from './auth'
import { encryptFile, decryptFile, generateMasterKey, wrapMasterKey, unwrapMasterKey, deriveKeyFromToken } from '../utils/crypto'
import { encryptChunkWorker, decryptChunkedFileWorker, CHUNK_SIZE } from '../utils/crypto'

export const useFileStore = defineStore('files', {
  state: () => ({
    files: [],
    folders: [],
    currentPath: '/',
    uploadProgress: 0,
    isUploading: false,
    uploadingFileName: '',
    searchQuery: '',
    shareUpdateTrigger: 0,
    recentFolders: JSON.parse(localStorage.getItem('recentFolders')) || [],
    recentFiles: JSON.parse(localStorage.getItem('recentFiles')) || [],
  }),
  actions: {
    addToHistory(item) {
        if (item.type === 'folder') {
            // Remove existing if present
            this.recentFolders = this.recentFolders.filter(f => f.path !== item.path)
            // Add to top
            this.recentFolders.unshift(item)
            // Limit to 5
            if (this.recentFolders.length > 5) this.recentFolders.pop()
            localStorage.setItem('recentFolders', JSON.stringify(this.recentFolders))
        } else {
            // Remove existing if present (check ID)
            this.recentFiles = this.recentFiles.filter(f => f.ID !== item.ID)
            this.recentFiles.unshift(item)
            if (this.recentFiles.length > 5) this.recentFiles.pop()
            localStorage.setItem('recentFiles', JSON.stringify(this.recentFiles))
        }
    },
    notifyShareUpdate() {
        this.shareUpdateTrigger++;
    },
    setSearchQuery(query) {
      this.searchQuery = query
    },
    async fetchItems(path) {
      this.searchQuery = ''; // Clear search query when navigating
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
    async performSearch(query) {
      this.searchQuery = query;
      if (!query || query.trim() === '') {
        return this.fetchItems(this.currentPath);
      }
      
      try {
        const response = await api.get('/files/search', {
          params: { q: query }
        });
        
        this.files = response.data.files || [];
        this.folders = response.data.folders || [];
      } catch (err) {
        console.error("Search error:", err);
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

      // Find the file in the store to get encrypted_key
      const file = this.files.find(f => f.ID === fileId);
      let fileKey = authStore.masterKey; // Default to masterKey for old files

      if (file && file.EncryptedKey) {
          // Decrypt the file key
          try {
              fileKey = await unwrapMasterKey(file.EncryptedKey, authStore.masterKey);
          } catch (e) {
              console.error("Failed to decrypt file key", e);
              alert("Erreur de déchiffrement de la clé du fichier.");
              return;
          }
      }

      try {
        // 1. Télécharger le blob chiffré
        const response = await api.get(`/files/download/${fileId}`, { responseType: 'blob' });
        
        // 2. Déchiffrer via Worker
        const decryptedBlob = await decryptChunkedFileWorker(response.data, fileKey, mimeType);

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

      // Generate a unique key for this file
      const fileKey = await generateMasterKey();
      // Encrypt this key with the user's master key
      const encryptedFileKey = await wrapMasterKey(fileKey, authStore.masterKey);

      // Check for active shares on this path
      let shareKeysMap = {};
      try {
          const shareRes = await api.get('/shares/check-path', { params: { path: this.currentPath } });
          const activeShares = shareRes.data.shares || [];
          
          for (const share of activeShares) {
              // Derive Share Key from Token
              const shareKey = await deriveKeyFromToken(share.Token);
              // Encrypt File Key with Share Key
              const encryptedForShare = await wrapMasterKey(fileKey, shareKey);
              shareKeysMap[share.ID] = encryptedForShare;
          }
      } catch (e) {
          console.error("Error checking active shares:", e);
      }

      const totalChunks = Math.ceil(file.size / CHUNK_SIZE);
      // Calculate estimated encrypted size: original size + overhead (16 bytes salt + 12 bytes IV) per chunk
      const overheadPerChunk = 28; 
      const totalEncryptedSize = file.size + (totalChunks * overheadPerChunk);

      let chunkIndex = 0;
      let offset = 0;
      this.isUploading = true;
      this.uploadProgress = 0;
      this.uploadingFileName = file.name;

      try {
        while(offset < file.size) {
          let chunkBlob = file.slice(offset, offset + CHUNK_SIZE);
          let chunkArrayBuffer = await chunkBlob.arrayBuffer();

          let encryptedChunkBlob = await encryptChunkWorker(chunkArrayBuffer, fileKey, chunkIndex);
          
          // Help GC: Release source buffer immediately
          chunkArrayBuffer = null;
          chunkBlob = null;

          let encryptedFile = new File([encryptedChunkBlob], file.name, { type: 'application/octet-stream' });
          
          // Help GC: Release encrypted blob reference
          encryptedChunkBlob = null;

          const formData = new FormData()
          formData.append('file', encryptedFile)
          formData.append('path', this.currentPath)
          formData.append('chunk_index', chunkIndex)
          formData.append('total_chunks', totalChunks)
          formData.append('total_file_size', totalEncryptedSize) // Send total size for quota check
          formData.append('encrypted_key', encryptedFileKey)
          
          // Send share keys only with the last chunk (or every chunk, but backend only uses it on commit)
          // To be safe and simple, send with every chunk, backend ignores until commit.
          if (Object.keys(shareKeysMap).length > 0) {
              formData.append('share_keys', JSON.stringify(shareKeysMap));
          }

          console.log(`Uploading chunk ${chunkIndex + 1} / ${totalChunks}...`);

          await api.post('/files/upload', formData, {
            headers: {
              'Content-Type': 'multipart/form-data',
            },
            onUploadProgress: (progressEvent) => {
                // Calculate global progress
                const currentChunkProgress = progressEvent.loaded;
                const previousProgress = chunkIndex * CHUNK_SIZE;
                const totalProgress = previousProgress + currentChunkProgress;
                this.uploadProgress = Math.min(Math.round((totalProgress / totalEncryptedSize) * 100), 100);
            }
          });
          
          // Help GC: Release file object
          encryptedFile = null;

          offset += CHUNK_SIZE;
          chunkIndex += 1;

          // this.uploadProgress is updated in onUploadProgress
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
          id: item.ID,
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
    },
    async renameItem(id, type, newName) {
      try {
        await api.post('/files/rename', {
          id: id,
          type: type,
          new_name: newName
        })
        this.fetchItems(this.currentPath)
      } catch (error) {
        console.error('Error renaming item:', error)
        throw error
      }
    },
    async updateTags(id, type, tags) {
      try {
        await api.post('/files/tags', {
          id: id,
          type: type,
          tags: tags
        })
        this.fetchItems(this.currentPath)
      } catch (error) {
        console.error('Error updating tags:', error)
        throw error
      }
    },
    async createShareLink(resourceId, resourceType, expiresAt = null) {
      const authStore = useAuthStore();
      
      // 1. Generate Token
      const tokenBytes = window.crypto.getRandomValues(new Uint8Array(32));
      const token = btoa(String.fromCharCode(...tokenBytes))
        .replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');

      let encryptedKeyForShare = "";
      let fileKeys = {};

      if (resourceType === 'file') {
          const file = this.files.find(f => f.ID === resourceId);
          if (file && file.EncryptedKey) {
              try {
                  // 2. Decrypt File Key
                  const fileKey = await unwrapMasterKey(file.EncryptedKey, authStore.masterKey);
                  
                  // 3. Derive Share Key from Token
                  const shareKey = await deriveKeyFromToken(token);

                  // 4. Encrypt File Key with Share Key
                  encryptedKeyForShare = await wrapMasterKey(fileKey, shareKey);
              } catch (e) {
                  console.error("Encryption error for share:", e);
                  throw new Error("Failed to prepare encryption for share.");
              }
          }
      } else if (resourceType === 'folder') {
          const folder = this.folders.find(f => f.ID === resourceId);
          if (folder) {
              try {
                  // Fetch ALL files in the folder recursively to get their keys
                  console.log(`Fetching recursive files for path: ${folder.Path}`);
                  const res = await api.get(`/files/list-recursive`, { params: { path: folder.Path } });
                  const filesInFolder = res.data.files || [];
                  console.log(`Found ${filesInFolder.length} files in folder.`);
                  
                  const shareKey = await deriveKeyFromToken(token);
                  
                  let missingKeysCount = 0;
                  for (const f of filesInFolder) {
                      console.log(`Processing file ${f.ID} (${f.Name}). Has Key: ${!!f.EncryptedKey}`);
                      if (f.EncryptedKey) {
                          const k = await unwrapMasterKey(f.EncryptedKey, authStore.masterKey);
                          const sk = await wrapMasterKey(k, shareKey);
                          fileKeys[f.ID] = sk;
                      } else {
                          missingKeysCount++;
                      }
                  }
                  console.log(`Prepared keys for ${Object.keys(fileKeys).length} files.`);

                  if (missingKeysCount > 0) {
                      console.warn(`${missingKeysCount} files in this folder are missing encryption keys.`);
                      alert(`Attention : ${missingKeysCount} fichiers dans ce dossier n'ont pas de clé de chiffrement (anciens fichiers ?). Ils ne seront pas lisibles via le partage.`);
                  }
              } catch (e) {
                  console.error("Error preparing folder share:", e);
                  // Continue anyway, maybe some files won't be readable
              }
          }
      }

      try {
        const response = await api.post('/shares/link', {
          resource_id: resourceId,
          resource_type: resourceType,
          expires_at: expiresAt,
          token: token,
          encrypted_key: encryptedKeyForShare,
          file_keys: fileKeys
        })
        return response.data
      } catch (error) {
        console.error('Error creating share link:', error)
        throw error
      }
    },
    async searchFiles(query) {
        if (!query || query.trim() === '') {
            // Si la recherche est vide, on recharge le dossier courant
            return this.fetchItems(this.currentPath);
        }

        try {
            const response = await api.get(`/files/search?q=${encodeURIComponent(query)}`);
            
            // Mise à jour des fichiers et dossiers affichés
            this.folders = response.data.folders || [];
            this.files = response.data.files || [];
        } catch (error) {
            console.error("Search error:", error);
        }
    },
    
    setSearchQuery(query) {
        this.searchQuery = query;
        this.searchFiles(query);
    }
  },
})