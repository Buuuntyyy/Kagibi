import { defineStore } from 'pinia'
import api from '../api'
import { useAuthStore } from './auth'
import { encryptFile, decryptFile, generateMasterKey, wrapMasterKey, unwrapMasterKey, deriveKeyFromToken } from '../utils/crypto'
import { encryptChunkWorker, decryptChunkedFileWorker, CHUNK_SIZE } from '../utils/crypto'
import { generatePreview } from '../utils/previewGenerator'
import sodium from 'libsodium-wrappers-sumo'

export const useFileStore = defineStore('files', {
  state: () => ({
    files: [],
    folders: [],
    currentPath: '/',
    // Shared Mode State
    viewMode: 'drive', // 'drive' or 'shared'
    sharedKey: null, // CryptoKey (AES-GCM) for the current shared root
    sharedBreadcrumbs: [], // Array of { id, name }
    
    uploadProgress: 0,
    isUploading: false,
    uploadingFileName: '',
    
    // Preview State
    preview: {
      show: false,
      url: null,
      type: null,
      name: null,
      loading: false,
      status: ''
    },

    searchQuery: '',
    shareUpdateTrigger: 0,
    recentFolders: JSON.parse(localStorage.getItem('recentFolders')) || [],
    recentFiles: JSON.parse(localStorage.getItem('recentFiles')) || [],
    // Used to coordinate navigation from Suggestions
    pendingNavigatePath: null,
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
        // If a pending navigation is set, consume it and use that path
        if (this.pendingNavigatePath) {
          path = this.pendingNavigatePath;
          this.pendingNavigatePath = null;
        }
      // If we are in shared mode and fetchItems('/') is called, switch back to drive mode
      if (path === '/' && this.viewMode === 'shared') {
          this.viewMode = 'drive';
          this.sharedKey = null;
          this.sharedBreadcrumbs = [];
      }

      // If in shared mode, ignore normal fetch (or maybe redirect to fetchShared?)
      // But typically fetchItems is called with a path string. 
      // Shared navigation uses IDs.
      if (this.viewMode === 'shared') {
          // If path is passed, it might be an error or legacy call. 
          // We'll ignore it unless it's strictly root which we handled above.
          // Or maybe we treat it as "refresh current".
          if (this.sharedBreadcrumbs.length > 0) {
              const current = this.sharedBreadcrumbs[this.sharedBreadcrumbs.length - 1];
              return this.fetchSharedFolderContent(current.id);
          }
          return;
      }

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
    
    // --- SHARED MODE ACTIONS ---
    async openSharedRoot(shareItem) {
        const authStore = useAuthStore();
        this.viewMode = 'shared';
        this.sharedBreadcrumbs = [{ id: shareItem.resource_id, name: shareItem.name || shareItem.Name }];
        
        try {
            await sodium.ready;
            
            // Decrypt the Shared Folder Key
            // Assuming direct share logic (RSA) similar to SharedWithMe.vue
            if (!shareItem.encrypted_key) {
                console.error("Shared folder has no key");
                alert("Dossier partagé verrouillé (clé manquante).");
                return;
            }

             // We only support Direct Share (RSA) for now at root
            if (!authStore.privateKey) {
                 throw new Error("Clé privée non disponible.");
            }

            const encryptedKeyBytes = sodium.from_base64(shareItem.encrypted_key);
            const rsaPrivateKey = authStore.privateKey; // CryptoKey

            const folderKeyRaw = await window.crypto.subtle.decrypt(
                { name: "RSA-OAEP" },
                rsaPrivateKey,
                encryptedKeyBytes
            );
            
            // Import as AES-GCM for file decryption later
            this.sharedKey = await window.crypto.subtle.importKey(
                "raw", 
                folderKeyRaw,
                "AES-GCM",
                true,
                ["decrypt"]
            );

            // Fetch Content
            await this.fetchSharedFolderContent(shareItem.resource_id);

        } catch (e) {
            console.error("Error opening shared folder:", e);
            alert("Erreur lors de l'ouverture du dossier partagé.");
            this.viewMode = 'drive';
        }
    },

    async navigateShared(folderId, folderName) {
        if (this.viewMode !== 'shared') return;
        
        // Push to breadcrumbs
        this.sharedBreadcrumbs.push({ id: folderId, name: folderName });
        await this.fetchSharedFolderContent(folderId);
    },

    async navigateSharedUp() {
        if (this.viewMode !== 'shared') return;
        if (this.sharedBreadcrumbs.length <= 1) {
            // Exit shared mode if asking to go up from root
            this.viewMode = 'drive';
            this.fetchItems('/');
            return;
        }

        this.sharedBreadcrumbs.pop();
        const current = this.sharedBreadcrumbs[this.sharedBreadcrumbs.length - 1];
        await this.fetchSharedFolderContent(current.id);
    },

    async navigateSharedTo(index) {
         if (this.viewMode !== 'shared') return;
         // Slice breadcrumbs to index+1
         this.sharedBreadcrumbs = this.sharedBreadcrumbs.slice(0, index + 1);
         const current = this.sharedBreadcrumbs[this.sharedBreadcrumbs.length - 1];
         await this.fetchSharedFolderContent(current.id);
    },

    async fetchSharedFolderContent(folderID) {
        try {
            const response = await api.get(`/shares/direct/folder/${folderID}/content`);
            const data = response.data;
            
            // Normalize to match standard file structure (PascalCase for component compat)
            this.files = (data.files || []).map(f => ({
                ...f,
                EncryptedKey: f.encrypted_key, // Key specifically for this file (encrypted with FolderKey)
                Name: f.Name || f.name,
                ID: f.ID || f.id,
                Size: f.Size || f.size,
                file_id: f.ID || f.id // Ensure we have something component might look for
            }));
            
            this.folders = (data.folders || []).map(f => ({
                ...f,
                Name: f.Name,
                ID: f.ID
            }));
            
            // We do NOT update currentPath string because it is path-based and we are ID-based
            // FileList.vue will need to read sharedBreadcrumbs instead
            
        } catch (error) {
            console.error("Error fetching shared folder content:", error);
            alert("Erreur de chargement du contenu.");
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
        if (this.viewMode === 'shared') {
             return; // ID-based navigation handled by components
        }
        let newPath = this.currentPath
        if (newPath.endsWith('/')) {
            newPath += folderName
        } else {
            newPath += `/${folderName}`
        }
        this.fetchItems(newPath)
    },
    navigateUp() {
        if (this.viewMode === 'shared') {
            this.navigateSharedUp();
            return;
        }
        if (this.currentPath === '/') return
        const parts = this.currentPath.split('/').filter(p => p)
        parts.pop()
        const newPath = '/' + parts.join('/')
        this.fetchItems(newPath)
    },
    async downloadFile(fileId, fileName, mimeType='application/octet-stream', preview = false) {
      const authStore = useAuthStore();
      
      // Attempt to correct MIME type based on extension if generic
      if ((!mimeType || mimeType.includes('application/octet-stream')) && fileName) {
          const ext = fileName.split('.').pop().toLowerCase();
          if (ext === 'pdf') mimeType = 'application/pdf';
          else if (['jpg', 'jpeg'].includes(ext)) mimeType = 'image/jpeg';
          else if (ext === 'png') mimeType = 'image/png';
          else if (ext === 'bmp') mimeType = 'image/bmp';
          else if (ext === 'svg') mimeType = 'image/svg+xml';
          else if (ext === 'gif') mimeType = 'image/gif';
          else if (ext === 'webp') mimeType = 'image/webp';
          else if (ext === 'txt') mimeType = 'text/plain';
      }

      // Reset preview state if starting a new preview
      if (preview) {
        this.preview = { 
            show: true, 
            url: null, 
            type: mimeType, 
            name: fileName,
            loading: true,
            status: 'Initialisation...'
        };
      }

      // SHARED MODE DOWNLOAD
      if (this.viewMode === 'shared') {
          await sodium.ready;
          if (preview) this.preview.status = 'Récupération de la clé...';
          const file = this.files.find(f => f.ID === fileId);
          if (!file || !this.sharedKey) return;
          
          if (!file.EncryptedKey) {
             alert("Clé de fichier manquante");
             return;
          }

          try {
             // Decrypt File Key (using Folder Key)
             const encryptedBytes = sodium.from_base64(file.EncryptedKey);
             // Assume IV is first 12 bytes
             const iv = encryptedBytes.slice(0, 12);
             const data = encryptedBytes.slice(12);
             
             const fileKeyRaw = await window.crypto.subtle.decrypt(
                { name: "AES-GCM", iv: iv },
                this.sharedKey, // The Folder Key
                data
             );
             
             const fileKeyCrypto = await window.crypto.subtle.importKey("raw", fileKeyRaw, "AES-GCM", true, ["decrypt"]);
             
             // Download content
             if (preview) this.preview.status = 'Téléchargement du fichier chiffré...';
             const response = await api.get(`/files/download/${fileId}`, { responseType: 'blob' });
             
             // Decrypt content
             if (preview) this.preview.status = 'Déchiffrement (Client-Side)...';
             const encryptedFileBytes = await response.data.arrayBuffer();
             const encryptedBlob = new Blob([encryptedFileBytes]);
             const decryptedBlob = await decryptChunkedFileWorker(encryptedBlob, fileKeyCrypto, mimeType);
             
             // Save or Preview
             const url = window.URL.createObjectURL(decryptedBlob);
             if (preview) {
                 this.preview = {
                    show: true,
                    url: url,
                    type: mimeType,
                    name: fileName,
                    loading: false, // Done
                    status: ''
                 };
             } else {
                 const a = document.createElement('a');
                 a.href = url;
                 a.download = fileName;
                 document.body.appendChild(a);
                 a.click();
                 window.URL.revokeObjectURL(url);
                 document.body.removeChild(a);
             }
          } catch (e) {
              console.error("Shared download error", e);
              alert("Erreur téléchargement partagé: " + e.message);
              if (preview) this.preview.show = false;
          }
          return;
      }

      if (!authStore.masterKey) return;

      // Find the file in the store to get encrypted_key
      const file = this.files.find(f => f.ID === fileId);
      
      let targetFileId = fileId;
      let targetEncryptedKey = file ? file.EncryptedKey : null;
      let finalMimeType = mimeType;

      // Use Preview file if available and requested
      if (preview && file && file.preview) {
          console.log("Using optimized preview:", file.preview.ID);
          targetFileId = file.preview.ID;
          targetEncryptedKey = file.preview.EncryptedKey;
          // Previews are usually JPEGs
          finalMimeType = file.preview.MimeType || 'image/jpeg';
      }

      let fileKey = authStore.masterKey; // Default to masterKey for old files

      if (targetEncryptedKey) {
          // Decrypt the file key
          if (preview) this.preview.status = 'Préparation de la clé...';
          try {
              fileKey = await unwrapMasterKey(targetEncryptedKey, authStore.masterKey);
          } catch (e) {
              console.error("Failed to decrypt file key", e);
              alert("Erreur de déchiffrement de la clé du fichier.");
              if (preview) this.preview.show = false;
              return;
          }
      }

      try {
        // 1. Télécharger le blob chiffré
        if (preview) this.preview.status = 'Téléchargement du contenu chiffré...';
        const response = await api.get(`/files/download/${targetFileId}`, { responseType: 'blob' });
        
        // 2. Déchiffrer via Worker
        if (preview) this.preview.status = 'Déchiffrement local...';
        const decryptedBlob = await decryptChunkedFileWorker(response.data, fileKey, finalMimeType);

        // 3. Sauvegarder ou Prévisualiser
        const url = window.URL.createObjectURL(decryptedBlob);
        
        if (preview) {
             this.preview = {
                show: true,
                url: url,
                type: finalMimeType,
                name: fileName,
                loading: false,
                status: ''
             };
        } else {
             const link = document.createElement('a');
             link.href = url;
             link.setAttribute('download', fileName);
             document.body.appendChild(link);
             link.click();
             setTimeout(() => { link.remove(); window.URL.revokeObjectURL(url); }, 100);
        }

      } catch (error) {
        console.error("Erreur download:", error);
        alert("Erreur lors du téléchargement.");
        if (preview) this.preview.show = false;
      }
    },
    async uploadFile(file, isPreview = false, previewID = null) {
      const authStore = useAuthStore()
      if (!authStore.isAuthenticated || !authStore.masterKey) {
        console.error('User not authenticated. Cannot upload file.')
        return
      }

      // 1. Generate Preview if main file and supported
      if (!isPreview && !previewID) {
          const previewBlob = await generatePreview(file);
          if (previewBlob) {
             console.log("Uploading preview...");
             const previewName = "(preview) " + (file.name || "file") + ".jpg";
             const previewFile = new File([previewBlob], previewName, { type: "image/jpeg" });
             try {
                const previewResult = await this.uploadFile(previewFile, true);
                if (previewResult && previewResult.ID) {
                    previewID = previewResult.ID;
                }
             } catch (e) {
                 console.warn("Failed to upload preview", e);
             }
          }
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
      let lastResponse = null;
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

          if (isPreview) {
              formData.append('is_preview', 'true');
          }
          if (previewID) {
              formData.append('preview_id', previewID.toString());
          }

          console.log(`Uploading chunk ${chunkIndex + 1} / ${totalChunks}...`);

          lastResponse = await api.post('/files/upload', formData, {
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

        if (!isPreview) {
            this.fetchItems(this.currentPath)
        }
        
        // Reset progress after a short delay
        setTimeout(() => {
            if (!isPreview) { // Only reset if main file
                this.isUploading = false;
                this.uploadProgress = 0;
            }
        }, 1000);

        return lastResponse?.data?.file;

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