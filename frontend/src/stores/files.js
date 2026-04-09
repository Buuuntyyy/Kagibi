import { defineStore } from 'pinia'
import api from '../api'
import { useAuthStore } from './auth'
import { usePreferencesStore } from './preferences'
import { encryptFile, decryptFile, generateMasterKey, wrapMasterKey, unwrapMasterKey, deriveKeyFromToken, encryptFileName, decryptFileName } from '../utils/crypto'
import { encryptChunkWorker, decryptChunkedFileWorker, CHUNK_SIZE } from '../utils/crypto'
import { generatePreview } from '../utils/previewGenerator'
import { MultipartUploadManager, PART_SIZE, UploadState } from '../utils/multipartUpload'
import sodium from 'libsodium-wrappers-sumo'

// Compress image for preview display (reduces memory, not bandwidth)
const PREVIEW_MAX_WIDTH = 1920;
const PREVIEW_QUALITY = 0.85;

function compressImageForPreview(blob) {
    return new Promise((resolve, reject) => {
        const img = new Image();
        const url = URL.createObjectURL(blob);
        img.onload = () => {
            URL.revokeObjectURL(url);
            
            // Skip compression if image is small enough
            if (img.width <= PREVIEW_MAX_WIDTH && blob.size < 2 * 1024 * 1024) {
                resolve(null); // Return null to signal no compression needed
                return;
            }
            
            const canvas = document.createElement('canvas');
            let width = img.width;
            let height = img.height;
            
            if (width > PREVIEW_MAX_WIDTH) {
                height = Math.round(height * (PREVIEW_MAX_WIDTH / width));
                width = PREVIEW_MAX_WIDTH;
            }
            
            canvas.width = width;
            canvas.height = height;
            const ctx = canvas.getContext('2d');
            ctx.drawImage(img, 0, 0, width, height);
            
            canvas.toBlob((compressedBlob) => {
                if (compressedBlob) resolve(compressedBlob);
                else resolve(null);
            }, 'image/jpeg', PREVIEW_QUALITY);
        };
        img.onerror = () => {
            URL.revokeObjectURL(url);
            resolve(null); // Don't fail, just use original
        };
        img.src = url;
    });
}

/**
 * Corrects a generic MIME type based on the file extension.
 * Returns the original mimeType if no correction is needed.
 */
function correctMimeType(mimeType, fileName) {
  if ((!mimeType || mimeType.includes('application/octet-stream')) && fileName) {
    const ext = fileName.split('.').pop().toLowerCase();
    const map = {
      pdf: 'application/pdf',
      jpg: 'image/jpeg',
      jpeg: 'image/jpeg',
      png: 'image/png',
      bmp: 'image/bmp',
      svg: 'image/svg+xml',
      gif: 'image/gif',
      webp: 'image/webp',
      txt: 'text/plain',
    };
    if (map[ext]) return map[ext];
  }
  return mimeType;
}

/**
 * Triggers a browser download for a given Blob.
 */
function triggerBlobDownload(blob, fileName) {
  const url = window.URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.setAttribute('download', fileName);
  document.body.appendChild(link);
  link.click();
  setTimeout(() => { link.remove(); window.URL.revokeObjectURL(url); }, 100);
}

/**
 * Decrypts a file's encrypted key using a folder/shared key (AES-GCM).
 * Returns the decrypted CryptoKey.
 */
async function decryptFileKeyWithFolderKey(encryptedKeyB64, folderKey) {
  const encryptedBytes = sodium.from_base64(encryptedKeyB64);
  const iv = encryptedBytes.slice(0, 12);
  const data = encryptedBytes.slice(12);
  const fileKeyRaw = await window.crypto.subtle.decrypt({ name: 'AES-GCM', iv }, folderKey, data);
  return window.crypto.subtle.importKey('raw', fileKeyRaw, 'AES-GCM', true, ['decrypt']);
}

/**
 * Builds a map of shareID → encryptedFileKey for all active shares on the given path.
 */
async function buildShareKeysMap(uploadPath, fileKey) {
  const shareKeysMap = {};
  try {
    const shareRes = await api.get('/shares/check-path', { params: { path: uploadPath } });
    const activeShares = shareRes.data.shares || [];
    for (const share of activeShares) {
      const shareKey = await deriveKeyFromToken(share.Token);
      shareKeysMap[share.ID] = await wrapMasterKey(fileKey, shareKey);
    }
  } catch (e) {
    console.error('Error checking active shares:', e);
  }
  return shareKeysMap;
}

/**
 * Encrypts all file chunks using the worker-based AES-GCM encryption.
 * Returns an array of encrypted Blobs and the total encrypted size.
 * Updates uploadProgress (0–30%) via the provided setter.
 */
async function encryptFileChunks(file, fileKey, onProgress) {
  const totalParts = Math.ceil(file.size / PART_SIZE);
  const encryptedChunks = [];
  let offset = 0;
  let chunkIndex = 0;
  while (offset < file.size) {
    let chunkBlob = file.slice(offset, offset + PART_SIZE);
    let chunkArrayBuffer = await chunkBlob.arrayBuffer();
    const encryptedChunkBlob = await encryptChunkWorker(chunkArrayBuffer, fileKey, chunkIndex);
    encryptedChunks.push(encryptedChunkBlob);
    chunkArrayBuffer = null;
    chunkBlob = null;
    offset += PART_SIZE;
    chunkIndex++;
    if (onProgress) onProgress(Math.round((chunkIndex / totalParts) * 30));
  }
  const totalEncryptedSize = encryptedChunks.reduce((sum, chunk) => sum + (chunk.size || chunk.byteLength || 0), 0);
  return { encryptedChunks, totalEncryptedSize };
}

/**
 * Generates a URL-safe random token for share links.
 */
function generateShareToken() {
  const tokenBytes = window.crypto.getRandomValues(new Uint8Array(32));
  return btoa(String.fromCodePoint(...tokenBytes))
    .replaceAll('+', '-').replaceAll('/', '_').replaceAll(/=+$/g, '');
}

/**
 * Encrypts the file key with a derived share key for a single-file share.
 * Returns the encrypted key string, or "" on failure.
 */
async function prepareFileShareKey(file, masterKey, token) {
  if (!file?.EncryptedKey) return '';
  try {
    const fileKey = await unwrapMasterKey(file.EncryptedKey, masterKey);
    const shareKey = await deriveKeyFromToken(token);
    return wrapMasterKey(fileKey, shareKey);
  } catch (e) {
    console.error('Encryption error for share:', e);
    throw new Error('Failed to prepare encryption for share.');
  }
}

/**
 * Builds the fileKeys map for a folder share: encrypts each file's key with the share key.
 * Returns the map { fileID: encryptedKey }.
 */
async function prepareFolderShareKeys(folder, masterKey, token) {
  const fileKeys = {};
  try {
    const res = await api.get('/files/list-recursive', { params: { path: folder.Path } });
    const filesInFolder = res.data.files || [];
    const shareKey = await deriveKeyFromToken(token);
    let missingKeysCount = 0;
    for (const f of filesInFolder) {
      if (f.EncryptedKey) {
        const k = await unwrapMasterKey(f.EncryptedKey, masterKey);
        fileKeys[f.ID] = await wrapMasterKey(k, shareKey);
      } else {
        missingKeysCount++;
      }
    }
    if (missingKeysCount > 0) {
      console.warn(`${missingKeysCount} files missing encryption keys.`);
      alert(`Attention : ${missingKeysCount} fichiers dans ce dossier n'ont pas de clé de chiffrement (anciens fichiers ?). Ils ne seront pas lisibles via le partage.`);
    }
  } catch (e) {
    console.error('Error preparing folder share:', e);
  }
  return fileKeys;
}

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
    uploadSpeed: 0,
    isUploading: false,
    uploadingFileName: '',
    uploadState: 'idle', // idle, encrypting, uploading, completing, error
    currentUploadManager: null, // MultipartUploadManager instance
    lastUploadedBytes: 0,
    lastUploadTimestamp: 0,
    heartbeatInterval: null, // Interval for session keepalive during long operations
    
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
    recentFolders: [],
    recentFiles: [],
    // Used to coordinate navigation from Suggestions
    pendingNavigatePath: null,
    // Maps encrypted folder path → decrypted display name (populated when encrypt_filenames=true)
    folderNameCache: {},
  }),
  actions: {
    async fetchRecents() {
        try {
            const res = await api.get('/users/recent')
            // Backend returns list of { type: 'file'|'folder', file: {...}, folder: {...}, ... }
            const items = res.data
            
            this.recentFiles = items
                .filter(i => i.type === 'file')
                .map(i => ({ ...i.file, type: 'file', displayName: i.file.Name }))
                
            this.recentFolders = items
                .filter(i => i.type === 'folder')
                .map(i => ({ ...i.folder, type: 'folder', displayName: i.folder.Name, path: i.folder.Path }))
                
        } catch (err) {
            console.error("Failed to fetch recent history", err)
        }
    },

    addToHistory(item) {
        // Optimistic UI Update first
        if (item.type === 'folder') {
            const path = item.path || item.Path
            this.recentFolders = this.recentFolders.filter(f => (f.path || f.Path) !== path)
            this.recentFolders.unshift(item)
            if (this.recentFolders.length > 10) this.recentFolders.pop()
        } else {
            const id = item.ID || item.id
            this.recentFiles = this.recentFiles.filter(f => (f.ID || f.id) !== id)
            this.recentFiles.unshift(item)
            if (this.recentFiles.length > 10) this.recentFiles.pop()
        }
        
        // Persist to Backend
        // Don't await strictly to not block UI
        const recentId = item.ID || item.id
        if (!recentId) {
          console.warn("No ID found for history item:", item)
          return
        }
        
        // Validate that recentId is a number (convert if needed)
        const numId = typeof recentId === 'string' ? parseInt(recentId, 10) : recentId
        if (isNaN(numId)) {
          console.warn("Invalid ID for history item:", recentId)
          return
        }
        
        api.post('/users/recent', {
          id: numId,
          type: item.type === 'folder' ? 'folder' : 'file'
        }).catch(err => console.error("Failed to save history", err))
    },

    notifyShareUpdate() {
        this.shareUpdateTrigger++;
    },
    
    startHeartbeat() {
      if (this.heartbeatInterval) {
        clearInterval(this.heartbeatInterval);
      }
      // Send heartbeat every 2.5 minutes (150 seconds)
      // Session timeout is 5 minutes, so this keeps it alive
      this.heartbeatInterval = setInterval(async () => {
        try {
          await api.get('/heartbeat');
          //console.log('[Upload/Download] Heartbeat sent to prevent session timeout');
        } catch (err) {
          console.error('[Upload/Download] Heartbeat failed:', err);
        }
      }, 150000);
    },
    
    stopHeartbeat() {
      if (this.heartbeatInterval) {
        clearInterval(this.heartbeatInterval);
        this.heartbeatInterval = null;
        //console.log('[Upload/Download] Heartbeat stopped');
      }
    },
    
        setSearchQuery(query) {
        this.searchQuery = query;
        this.searchFiles(query);
    },
    async fetchItems(path) {
      const preferenceStore = usePreferencesStore()
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
        const params = preferenceStore.showFolderSizes ? { include_folder_sizes: '1' } : undefined
        const response = await api.get(`/files/list${safePath}`, { params })
        this.files = response.data.files || []
        this.folders = response.data.folders || []
        this.currentPath = safePath

        // Decrypt names client-side when the user opted into filename encryption.
        // Names are stored as opaque base64url AES-GCM blobs; we decrypt them in
        // place so all existing UI code reading item.Name continues to work unchanged.
        const authStore = useAuthStore()
        if (authStore.user?.encrypt_filenames && authStore.masterKey) {
          for (const file of this.files) {
            file.Name = await decryptFileName(file.Name, authStore.masterKey)
          }
          for (const folder of this.folders) {
            const decrypted = await decryptFileName(folder.Name, authStore.masterKey)
            // Cache encrypted path → decrypted name so breadcrumbs can display correctly.
            this.folderNameCache[folder.Path] = decrypted
            folder.Name = decrypted
          }
        }

        // Record History
        const fid = response.data.current_folder_id;
        if (fid && fid !== 0) {
            this.addToHistory({ 
                type: 'folder', 
                id: fid, 
                name: safePath.split('/').pop(), 
                path: safePath 
            });
        }
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

    /** Shared-mode download: decrypts the file key with the folder key, then downloads and decrypts. */
    async _downloadSharedFile(fileId, fileName, mimeType, preview) {
      await sodium.ready;
      if (preview) this.preview.status = 'Récupération de la clé...';
      const file = this.files.find(f => f.ID === fileId);
      if (!file) { alert('Fichier introuvable.'); if (preview) this.preview.show = false; return; }
      if (!this.sharedKey) { alert('Clé de déchiffrement manquante.'); if (preview) this.preview.show = false; return; }
      if (!file.EncryptedKey) { alert('Clé de fichier manquante'); if (preview) this.preview.show = false; return; }

      try {
        const fileKeyCrypto = await decryptFileKeyWithFolderKey(file.EncryptedKey, this.sharedKey);
        if (preview) this.preview.status = 'Téléchargement du fichier chiffré...';
        const response = await api.get(`/files/download/${fileId}`, { responseType: 'blob' });
        if (preview) this.preview.status = 'Déchiffrement (Client-Side)...';
        const encryptedBlob = new Blob([await response.data.arrayBuffer()]);
        const decryptedBlob = await decryptChunkedFileWorker(encryptedBlob, fileKeyCrypto, mimeType);
        const url = window.URL.createObjectURL(decryptedBlob);
        if (preview) {
          this.preview = { show: true, url, type: mimeType, name: fileName, loading: false, status: '' };
        } else {
          const a = document.createElement('a');
          a.href = url; a.download = fileName;
          document.body.appendChild(a); a.click();
          window.URL.revokeObjectURL(url); document.body.removeChild(a);
        }
      } catch (e) {
        console.error('Shared download error', e);
        alert('Erreur téléchargement partagé: ' + e.message);
        if (preview) this.preview.show = false;
      }
    },

    /** Resolves which file ID, encrypted key and MIME type to use for a download/preview. */
    _resolveDownloadTarget(fileId, mimeType, encryptedKey, preview) {
      const file = this.files.find(f => f.ID === fileId);
      let targetFileId = fileId;
      let targetEncryptedKey = encryptedKey || (file ? file.EncryptedKey : null);
      let finalMimeType = mimeType;
      if (preview && file && file.preview) {
        targetFileId = file.preview.ID;
        targetEncryptedKey = file.preview.EncryptedKey;
        finalMimeType = 'image/jpeg';
      }
      return { file, targetFileId, targetEncryptedKey, finalMimeType };
    },

    async downloadFile(fileId, fileName, mimeType = 'application/octet-stream', preview = false, encryptedKey = null) {
      const authStore = useAuthStore();
      if (!preview) this.startHeartbeat();

      mimeType = correctMimeType(mimeType, fileName);

      if (preview) {
        this.preview = { show: true, url: null, type: mimeType, name: fileName, loading: true, status: 'Initialisation...' };
      }

      if (this.viewMode === 'shared') {
        try {
          await this._downloadSharedFile(fileId, fileName, mimeType, preview);
        } finally {
          if (!preview) this.stopHeartbeat();
        }
        return;
      }

      if (!authStore.masterKey) {
        alert('Erreur d\'authentification (Clé manquante). Veuillez vous reconnecter.');
        if (preview) this.preview.show = false;
        if (!preview) this.stopHeartbeat();
        return;
      }

      const { file, targetFileId, targetEncryptedKey, finalMimeType: resolvedMime } = this._resolveDownloadTarget(fileId, mimeType, encryptedKey, preview);
      let finalMimeType = resolvedMime;

      let fileKey = authStore.masterKey;
      if (targetEncryptedKey) {
        if (preview) this.preview.status = 'Préparation de la clé...';
        try {
          fileKey = await unwrapMasterKey(targetEncryptedKey, authStore.masterKey);
        } catch (e) {
          console.error('Failed to decrypt file key', e);
          alert('Erreur de déchiffrement de la clé du fichier.');
          if (preview) this.preview.show = false;
          if (!preview) this.stopHeartbeat();
          return;
        }
      }

      try {
        if (preview) this.preview.status = 'Téléchargement du contenu chiffré...';
        const endpoint = (preview && file && file.preview) ? `/files/preview/${targetFileId}` : `/files/download/${targetFileId}`;
        const response = await api.get(endpoint, { responseType: 'blob' });

        if (preview) this.preview.status = 'Déchiffrement local...';
        let decryptedBlob = await decryptChunkedFileWorker(response.data, fileKey, finalMimeType);

        if (preview && finalMimeType.startsWith('image/') && !finalMimeType.includes('svg')) {
          this.preview.status = 'Optimisation pour affichage...';
          try {
            const compressedBlob = await compressImageForPreview(decryptedBlob);
            if (compressedBlob) { decryptedBlob = compressedBlob; finalMimeType = 'image/jpeg'; }
          } catch (e) { console.warn('Image compression failed, using original', e); }
        }

        const url = window.URL.createObjectURL(decryptedBlob);
        if (preview) {
          this.preview = { show: true, url, type: finalMimeType, name: fileName, loading: false, status: '' };
        } else {
          triggerBlobDownload(decryptedBlob, fileName);
        }

        this.addToHistory({ id: fileId, type: 'file', displayName: fileName, MimeType: mimeType, EncryptedKey: encryptedKey });
      } catch (error) {
        console.error('Erreur download:', error);
        alert('Erreur lors du téléchargement.');
        if (preview) this.preview.show = false;
      } finally {
        if (!preview) this.stopHeartbeat();
      }
    },
    async uploadFile(file, isPreview = false, previewID = null, previewPath = null) {
      const authStore = useAuthStore()
      if (!authStore.isAuthenticated || !authStore.masterKey) {
        console.error('User not authenticated. Cannot upload file.')
        return
      }

      // Use preview path if provided, otherwise use current path
      const uploadPath = previewPath !== null ? previewPath : this.currentPath;

      // Generate Preview if main file and supported (images/PDFs)
      if (!isPreview && !previewID) {
          const previewBlob = await generatePreview(file);
          if (previewBlob) {
             const safeName = (file.name || "file").replaceAll(/[^a-zA-Z0-9.-]/g, '_');
             const previewName = "preview_" + safeName + ".jpg";
             const previewFile = new File([previewBlob], previewName, { type: "image/jpeg" });
             try {
               const previewResult = await this.uploadFile(previewFile, true, null, uploadPath);
                if (previewResult && previewResult.ID) {
                    previewID = previewResult.ID;
                } else {
                    console.warn("Preview uploaded but no ID returned");
                }
             } catch (e) {
                 console.error("Failed to upload preview:", e);
             }
          }
      }

      // Generate a unique key for this file
      const fileKey = await generateMasterKey();
      // Encrypt this key with the user's master key
      const encryptedFileKey = await wrapMasterKey(fileKey, authStore.masterKey);

      // Check for active shares on this path
      const shareKeysMap = await buildShareKeysMap(uploadPath, fileKey);

      // Setup upload state
      this.isUploading = true;
      this.uploadProgress = 0;
      this.uploadSpeed = 0;
      this.uploadingFileName = file.name;
      this.uploadState = 'encrypting';
      this.lastUploadedBytes = 0;
      this.lastUploadTimestamp = 0;
      
      // Start heartbeat to prevent session timeout during long uploads
      if (!isPreview) {
        this.startHeartbeat();
      }

      // Create multipart upload manager
      const uploadManager = new MultipartUploadManager({
        onProgress: (percent, uploaded, total) => {
          this.uploadProgress = percent;

          const now = Date.now();
          if (this.lastUploadTimestamp === 0) {
            this.lastUploadTimestamp = now;
            this.lastUploadedBytes = uploaded;
            return;
          }

          const elapsedMs = now - this.lastUploadTimestamp;
          const uploadedDelta = uploaded - this.lastUploadedBytes;
          if (elapsedMs >= 250 && uploadedDelta >= 0) {
            const instantSpeed = uploadedDelta / (elapsedMs / 1000);
            this.uploadSpeed = this.uploadSpeed > 0
              ? (this.uploadSpeed * 0.7) + (instantSpeed * 0.3)
              : instantSpeed;
            this.lastUploadTimestamp = now;
            this.lastUploadedBytes = uploaded;
          }
        },
        onStateChange: (state) => {
          if (state === UploadState.UPLOADING) {
            this.uploadState = 'uploading';
          } else if (state === UploadState.COMPLETED) {
            this.uploadState = 'completing';
          } else if (state === UploadState.FAILED || state === UploadState.ABORTED) {
            this.uploadState = 'error';
          }
        },
        onError: (error, partNumber) => {
          console.error(`Upload error on part ${partNumber}:`, error);
        }
      });
      this.currentUploadManager = uploadManager;

      try {
        // Encrypt all chunks first (client-side ZK encryption), updating progress 0–30%
        const { encryptedChunks, totalEncryptedSize } = await encryptFileChunks(
          file, fileKey, (pct) => { this.uploadProgress = pct; }
        );

        this.uploadState = 'uploading';

        // Encrypt filename if the user opted into client-side filename encryption.
        // The encrypted name is a base64url string that passes backend validation
        // and is stored opaquely in both PostgreSQL and S3.
        let uploadFileName = file.name;
        if (authStore.user?.encrypt_filenames && authStore.masterKey) {
          uploadFileName = await encryptFileName(file.name, authStore.masterKey);
        }

        // Initiate multipart upload with backend
        await uploadManager.initiate(
          uploadFileName,
          uploadPath,
          'application/octet-stream',
          totalEncryptedSize,
          encryptedFileKey
        );

        // Upload parts directly to S3 (parallel with retry)
        const completedParts = await uploadManager.uploadParts(encryptedChunks);

        // Help GC
        encryptedChunks.length = 0;

        this.uploadState = 'completing';

        // Complete the multipart upload
        const result = await uploadManager.complete(completedParts, {
          fileName: uploadFileName,
          filePath: uploadPath,
          totalSize: totalEncryptedSize,
          contentType: 'application/octet-stream',
          encryptedKey: encryptedFileKey,
          shareKeys: Object.keys(shareKeysMap).length > 0 ? JSON.stringify(shareKeysMap) : '',
          previewId: previewID,
          isPreview: isPreview
        });

        if (!isPreview) {
            this.fetchItems(this.currentPath);
            // Update user quota
            await authStore.fetchUser();

            // Add to history
            if (result?.file) {
                this.addToHistory({ 
                    ...result.file, 
                    type: 'file', 
                    displayName: result.file.Name 
                });
            }
        }
        
        // Reset progress after a short delay
        setTimeout(() => {
            if (!isPreview) {
                this.isUploading = false;
                this.uploadProgress = 0;
                this.uploadSpeed = 0;
                this.uploadState = 'idle';
                this.currentUploadManager = null;
                this.lastUploadedBytes = 0;
                this.lastUploadTimestamp = 0;
            }
        }, 1000);

        return result?.file;

      } catch (error) {
        console.error("Erreur upload multipart:", error);
        
        // Attempt to abort the upload
        if (uploadManager) {
          try {
            await uploadManager.abort();
          } catch (abortError) {
            console.error("Error aborting upload:", abortError);
          }
        }
        
        alert("Erreur lors de l'envoi du fichier: " + error.message);
        this.isUploading = false;
        this.uploadProgress = 0;
        this.uploadSpeed = 0;
        this.uploadState = 'idle';
        this.currentUploadManager = null;
        this.lastUploadedBytes = 0;
        this.lastUploadTimestamp = 0;
        
        // Stop heartbeat on error
        if (!isPreview) {
          this.stopHeartbeat();
        }
        
        throw error;
      } finally {
        // Always stop heartbeat when upload completes
        if (!isPreview) {
          this.stopHeartbeat();
        }
      }
    },

    /**
     * Cancel the current upload
     */
    async cancelUpload() {
      if (this.currentUploadManager) {
        await this.currentUploadManager.abort();
        this.isUploading = false;
        this.uploadProgress = 0;
        this.uploadSpeed = 0;
        this.uploadState = 'idle';
        this.currentUploadManager = null;
        this.lastUploadedBytes = 0;
        this.lastUploadTimestamp = 0;
      }
    },
    async createFolder(folderName) {
      try {
        const authStore = useAuthStore()
        let nameToSend = folderName;
        if (authStore.user?.encrypt_filenames && authStore.masterKey) {
          nameToSend = await encryptFileName(folderName, authStore.masterKey);
        }
        await api.post('/folders/create', {
          name: nameToSend,
          path: this.currentPath,
        })
        this.fetchItems(this.currentPath)
      } catch (error) {
        console.error('Error creating folder:', error)
      }
    },
    async deleteFiles(fileIDs) {
    try {
      const authStore = useAuthStore()
      await api.post('/files/bulk-delete', { file_ids: fileIDs })
      this.fetchItems(this.currentPath)
      await authStore.fetchUser()
    } catch (error) {
      console.error('Error deleting items:', error)
    }
    },
    async moveItems(items, destinationPath) {
      try {
        const promises = items.map(item => api.post('/files/move', {
          id: item.ID || item.id,
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
        const authStore = useAuthStore()
        let nameToSend = newName;
        if (authStore.user?.encrypt_filenames && authStore.masterKey) {
          // Encrypt the complete new name (including extension) before sending.
          // Backend extension auto-append is bypassed for encrypted names (no visible ext).
          nameToSend = await encryptFileName(newName, authStore.masterKey);
        }
        await api.post('/files/rename', {
          id: id,
          type: type,
          new_name: nameToSend
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
      const token = generateShareToken();

      let encryptedKeyForShare = '';
      let fileKeys = {};

      if (resourceType === 'file') {
        const file = this.files.find(f => f.ID === resourceId);
        encryptedKeyForShare = await prepareFileShareKey(file, authStore.masterKey, token);
      } else if (resourceType === 'folder') {
        const folder = this.folders.find(f => f.ID === resourceId);
        if (folder) {
          fileKeys = await prepareFolderShareKeys(folder, authStore.masterKey, token);
        }
      }

      try {
        const response = await api.post('/shares/link', {
          resource_id: resourceId,
          resource_type: resourceType,
          expires_at: expiresAt,
          token,
          encrypted_key: encryptedKeyForShare,
          file_keys: fileKeys,
        });
        return response.data;
      } catch (error) {
        console.error('Error creating share link:', error);
        throw error;
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
  },
})