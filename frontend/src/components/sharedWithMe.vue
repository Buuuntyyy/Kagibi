<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="shared-with-me-container">
    <div v-if="loading" class="loading">
      <div class="spinner"></div> {{ t('shared.loadingShort') }}
    </div>
    
    <div v-if="currentFolder" class="folder-header">
         <button @click="navigateUp" class="back-btn">⬅ {{ t('shared.back') }}</button>
         <span class="current-path">{{ t('shared.folderPrefix') }} {{ currentFolder.name }}</span>
    </div>

    <div v-if="error" class="error">{{ error }}</div>
    <div v-else-if="!loading && items.length === 0" class="empty">
      <p v-if="currentFolder">{{ t('shared.emptyFolder') }}</p>
      <p v-else>{{ t('shared.emptySharedWithMe') }}</p>
    </div>
    
    <FileTable 
      v-else 
      :folders="sharedFolders"
      :files="sharedFiles"
      :columns="columns"
      @context-menu="handleContextMenu"
      @open-folder="handleOpenFolder"
      @open-file="handleOpenFile"
    >
      <template #shared_name="{ item }">
        <span class="file-link" @click.stop="handleItemClick(item)" :title="item.Name || item.name">{{ item.Name || item.name }}</span>
      </template>


      <template #owner="{ item }">
        {{ item.owner_name }}
      </template>

      <template #shared_at="{ item }">
        {{ formatDate(item.shared_at) }}
      </template>

      <template #size="{ item }">
        {{ formatSize(item.size) }}
      </template>

      <template #actions="{ item }">
        <!-- Actions futures (ex: télécharger, supprimer de ma liste) -->
      </template>
    </FileTable>

    <!-- Context Menu -->
    <ContextMenu
      v-if="contextMenu.visible"
      :x="contextMenu.x"
      :y="contextMenu.y"
      :item="contextMenu.item"
      @close="closeContextMenu"
      @action="handleContextAction"
    >
        <template #custom-actions>
            <div class="menu-item" @click="handleContextAction('download')" v-if="contextMenu.item.type === 'file'">
              <span class="menu-icon">⬇️</span> {{ t('shared.downloadDecrypt') }}
            </div>
            <div class="menu-item" @click="handleContextAction('download-zip')" v-if="contextMenu.item.type === 'folder' && !isZipping">
              <span class="menu-icon">🗜️</span> {{ t('shared.downloadAsZip') }}
            </div>
            <div class="menu-item disabled" v-if="contextMenu.item.type === 'folder' && isZipping">
              <span class="menu-icon">⏳</span> {{ zipProgress }}/{{ zipTotal }}
            </div>
             <div class="menu-divider"></div>
            <div class="menu-item delete" @click="handleContextAction('delete')">
              <span class="menu-icon">🗑️</span> {{ t('shared.removeShare') }}
            </div>
        </template>
    </ContextMenu>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, watch } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import FileTable from './file/FileTable.vue';
import { formatSize, formatDate } from '../utils/format';
import api from '../api';
import ContextMenu from './file/ContextMenu.vue';
import { useFileStore } from '../stores/files';
import { useAuthStore } from '../stores/auth';
import { decryptChunkedFileWorker } from '../utils/crypto';
import { zipSync } from 'fflate';
import sodium from 'libsodium-wrappers-sumo';

const fileStore = useFileStore();
const authStore = useAuthStore();
const { t } = useI18n();

const isZipping = ref(false);
const zipProgress = ref(0);
const zipTotal = ref(0);
const router = useRouter();
const route = useRoute();
const items = ref([]);
const loading = ref(false);
const error = ref(null);

const currentFolder = ref(null);

watch(() => fileStore.shareUpdateTrigger, () => {
    if (currentFolder.value) {
        fetchFolderContent(currentFolder.value.resource_id);
    } else {
        fetchSharedWithMe();
    }
});

const contextMenu = ref({
  visible: false,
  x: 0,
  y: 0,
  item: null
});

const columns = [
  { key: 'icon', label: '', headerClass: 'icon-col', cellClass: 'icon-col' },
  { key: 'shared_name', label: t('file.columnName'), cellClass: 'name-cell' },
  { key: 'owner', label: t('shared.owner') },
  { key: 'size', label: t('file.columnSize') },
  { key: 'actions', label: t('shared.actions') }
]

const sharedFolders = computed(() => {
    return items.value
        .filter(i => i.type === 'folder')
        .map(f => ({
            ...f,
            // Ensure standard props for FileTable if coming from fetchSharedWithMe
            Name: f.Name || f.name,
            ID: f.ID || f.resource_id,
            shared: false
        }));
});

const sharedFiles = computed(() => {
    return items.value
        .filter(i => i.type === 'file')
        .map(f => ({
            ...f,
            Name: f.Name || f.name,
            ID: f.ID || f.resource_id,
            Size: f.Size || f.size,
            MimeType: f.MimeType || f.mime_type
        }));
});

const fetchSharedWithMe = async () => {
  loading.value = true;
  try {
    // Ensure RSA keys are available before fetching shared files
    // This is crucial for decrypting shared folder/file keys
    if (authStore.masterKey && !authStore.privateKey) {
      try {
        await authStore.ensureRSAKeys(authStore.masterKey);
      } catch (e) {
        console.error("Failed to ensure RSA keys:", e);
        error.value = t('shared.decryptKeysLoadError');
        return;
      }
    }
    
    const response = await api.get('/shares/with-me');
    // Map response to match FileTable expectation
    items.value = (response.data || []).map(share => ({
        ...share,
        // Backend now returns 'name', 'size', 'type' directly
        // share.type is 'file' or 'folder'
        name: share.name,
        type: share.type, 
        size: share.size || 0,
        shared_at: share.shared_at,
        owner_name: share.owner_name || t('shared.unknownOwner'),
        // Determine if direct share for context actions
        is_direct: !!share.file_id || !!share.folder_id,
        // Helper specifically for download
        resource_id: share.file_id || share.folder_id,
        // Map Name/ID early too
        Name: share.name,
        ID: share.file_id || share.folder_id
    }));

    // Auto-open folder if requested via query param
    const targetFolderId = route.query.folderId;
    if (targetFolderId && !currentFolder.value) {
        const target = items.value.find(i => i.resource_id === targetFolderId && i.type === 'folder');
        if (target) {
            handleOpenFolder(target);
        }
    }

  } catch (err) {
    console.error("Error fetching shared with me:", err);
    error.value = t('shared.loadSharedWithMeError');
  } finally {
    loading.value = false;
  }
};

const fetchFolderContent = async (folderID) => {
    loading.value = true;
    try {
        const response = await api.get(`/shares/direct/folder/${folderID}/content`);
        const data = response.data;
        
        const files = (data.files || [])
          .map(f => ({
            ...f,
            type: 'file',
            resource_id: f.ID, // Standardize ID
            encrypted_key: f.encrypted_key,
            Name: f.Name,
            Size: f.Size,
            ID: f.ID
        }));
        
        const folders = (data.folders || []).map(f => ({
            ...f,
            type: 'folder',
            resource_id: f.ID,
            Name: f.Name,
            ID: f.ID
        }));
        
        items.value = [...folders, ...files];
    } catch (err) {
        console.error("Error fetching folder content:", err);
      error.value = t('shared.openFolderError');
    } finally {
        loading.value = false;
    }
}

const handleItemClick = (item) => {
    if (item.type === 'folder') {
        handleOpenFolder(item);
    } else {
        handleOpenFile(item);
    }
}

const navigateUp = () => {
    currentFolder.value = null;
    fetchSharedWithMe();
}

const handleOpenFolder = async (folder) => {
    if (!folder) {
        console.warn("handleOpenFolder called with undefined folder");
        return;
    }
    // Navigate to /dashboard/files and open the shared folder there
    await fileStore.openSharedRoot({
        id: folder.id || folder.ID,
        resource_id: folder.resource_id || folder.ID || folder.id,
        name: folder.name || folder.Name,
        encrypted_key: folder.encrypted_key,
    });
    router.push('/dashboard/files');
}

const handleContextAction = async (action) => {
    const item = contextMenu.value.item;
    if (!item) return;

    if (action === 'download') {
        if (item.type === 'file') {
            await downloadSharedFile(item);
        }
    } else if (action === 'download-zip') {
        if (item.type === 'folder') {
            await downloadFolderAsZip(item);
        }
    }
        else if (action === 'delete') {
          if (confirm(t('shared.removeShareConfirm'))) {
             try {
                let url = `/shares/with-me/${item.id}`;
                // Determine type query param
                if (item.file_id) {
                    url += '?type=direct_file';
                } else if (item.folder_id) {
                    url += '?type=direct_folder';
                } else {
                    url += '?type=imported';
                }

                await api.delete(url);
                await fetchSharedWithMe();
             } catch (e) {
                 console.error(e);
               alert(t('shared.deleteError'));
             }
         }
    }
    closeContextMenu();
}

const handleOpenFile = (item) => {
    // If it's a link share (Imported), navigate to the public link view
    if (item.link) {
        // Warning: Public links normally require the #key if it was in the hash. 
        // Imported shares typically only store the token. 
        // If the original link had a hash key, it might be lost unless we stored it.
        // For now, let's assume we just open the link.
        router.push(item.link);
        return;
    }

    // Preview if possible, or download (Direct Share)
    if (item.type === 'file') downloadSharedFile(item);
}

const downloadSharedFile = async (item) => {
    try {
        // If it's a link share, we can't use this method
        if (item.link) {
            window.open(item.link, '_blank');
            return;
        }

        // 1. Decrypt the Share Key
        // The item.encrypted_key is the FileKey encrypted with MY Public Key (if direct share)
        // OR encrypted with the Folder Key (if inside a shared folder)
        
        await sodium.ready;
        let fileKeyCrypto;

        if (currentFolderKey.value) {
             // We are inside a shared folder
             if (!item.encrypted_key) throw new Error(t('shared.missingFileKey'));
             
             // Decrypt using AES-GCM (FolderKey)
             const encryptedBytes = sodium.from_base64(item.encrypted_key);
             // Assume IV is first 12 bytes
             const iv = encryptedBytes.slice(0, 12);
             const data = encryptedBytes.slice(12);
             
             try {
                const fileKeyRaw = await window.crypto.subtle.decrypt(
                    { name: "AES-GCM", iv: iv },
                    currentFolderKey.value,
                    data
                );
                fileKeyCrypto = await window.crypto.subtle.importKey("raw", fileKeyRaw, "AES-GCM", true, ["decrypt"]);
             } catch (e) {
               throw new Error(t('shared.decryptFileKeyError'));
             }

        } else {
            // Root Share (Direct)
            if (!item.encrypted_key) {
            throw new Error(t('shared.missingEncryptionKey'));
            }
            
            // Ensure RSA keys are loaded before attempting decryption
            if (!authStore.privateKey && authStore.masterKey) {
                await authStore.ensureRSAKeys(authStore.masterKey);
            }
            
            // Decrypt User Private Key first if not ready
            if (!authStore.privateKey) {
                throw new Error(t('shared.privateKeyUnavailable'));
            }

            const encryptedKeyBytes = sodium.from_base64(item.encrypted_key);
            const rsaPrivateKey = authStore.privateKey; // CryptoKey

            const fileKeyRawBuffer = await window.crypto.subtle.decrypt(
                { name: "RSA-OAEP" },
                rsaPrivateKey,
                encryptedKeyBytes
            );
            
            fileKeyCrypto = await window.crypto.subtle.importKey(
                "raw", 
                fileKeyRawBuffer,
                "AES-GCM",
                true,
                ["decrypt"]
            );
        }

        // 2. Download and Decrypt Content
        const response = await api.get(`/files/download/${item.resource_id}`, { responseType: 'blob' });
        const encryptedFileBytes = await response.data.arrayBuffer(); // This is the encrypted file
        const encryptedBlob = new Blob([encryptedFileBytes]);
        
        const decryptedBlob = await decryptChunkedFileWorker(encryptedBlob, fileKeyCrypto, item.mime_type || 'application/octet-stream');
        
        // 3. Trigger Download
        const url = window.URL.createObjectURL(decryptedBlob);
        const a = document.createElement('a');
        a.href = url;
        a.download = item.name;
        document.body.appendChild(a);
        a.click();
        window.URL.revokeObjectURL(url);
        document.body.removeChild(a);

    } catch (e) {
        console.error("Download error:", e);
    alert(`${t('shared.downloadDecryptError')}: ${e.message}`);
    }
}

const downloadFolderAsZip = async (folder) => {
    const folderID = folder.resource_id || folder.ID || folder.id;
    if (!folderID) return;

    if (!authStore.privateKey && authStore.masterKey) {
        await authStore.ensureRSAKeys(authStore.masterKey);
    }
    if (!authStore.privateKey) {
        alert(t('shared.privateKeyUnavailable'));
        return;
    }

    isZipping.value = true;
    zipProgress.value = 0;
    zipTotal.value = 0;

    try {
        await sodium.ready;

        const res = await api.get(`/shares/direct/folder/${folderID}/files-recursive`);
        const files = res.data.files || [];
        const rootEncryptedKey = res.data.root_encrypted_key;

        if (files.length === 0) {
            alert(t('shared.emptyFolder'));
            return;
        }

        // Decrypt root folder key with RSA private key
        const encryptedKeyBytes = sodium.from_base64(rootEncryptedKey);
        const rootFolderKeyRaw = await window.crypto.subtle.decrypt(
            { name: 'RSA-OAEP' },
            authStore.privateKey,
            encryptedKeyBytes
        );
        const rootFolderKey = await window.crypto.subtle.importKey(
            'raw', rootFolderKeyRaw, 'AES-GCM', false, ['decrypt']
        );

        zipTotal.value = files.length;
        const zipData = {};

        for (let i = 0; i < files.length; i++) {
            const file = files[i];
            try {
                // Decrypt file key with root folder key (AES-GCM, IV = first 12 bytes)
                const encFileKeyBytes = sodium.from_base64(file.encrypted_key);
                const iv = encFileKeyBytes.slice(0, 12);
                const ciphertext = encFileKeyBytes.slice(12);
                const fileKeyRaw = await window.crypto.subtle.decrypt(
                    { name: 'AES-GCM', iv },
                    rootFolderKey,
                    ciphertext
                );
                const fileKey = await window.crypto.subtle.importKey(
                    'raw', fileKeyRaw, 'AES-GCM', false, ['decrypt']
                );

                const response = await api.get(`/files/download/${file.id}`, { responseType: 'blob' });
                const decryptedBlob = await decryptChunkedFileWorker(
                    response.data,
                    fileKey,
                    file.mime_type || 'application/octet-stream'
                );
                const buffer = await decryptedBlob.arrayBuffer();
                zipData[file.relative_path] = new Uint8Array(buffer);
            } catch (e) {
                console.error(`ZIP: skipping ${file.name}:`, e);
            }
            zipProgress.value = i + 1;
        }

        if (Object.keys(zipData).length === 0) {
            alert(t('shared.zipNoFiles'));
            return;
        }

        const zipped = zipSync(zipData, { level: 0 });
        const blob = new Blob([zipped], { type: 'application/zip' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `${folder.name || folder.Name || 'dossier'}.zip`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        setTimeout(() => URL.revokeObjectURL(url), 5000);
    } catch (e) {
        console.error('ZIP download error:', e);
        alert(`${t('shared.downloadDecryptError')}: ${e.message}`);
    } finally {
        isZipping.value = false;
        zipProgress.value = 0;
        zipTotal.value = 0;
    }
};

onMounted(() => {
  fetchSharedWithMe();
});
</script>

<style scoped>
.shared-with-me-container {
  height: 100%;
  width: 100%;
}

.loading, .error, .empty {
  text-align: center;
  padding: 20px;
  color: #888;
  background-color: var(--background-color);
}

.error {
  color: #ef5350;
}

.spinner {
  border: 3px solid rgba(255, 255, 255, 0.1);
  border-radius: 50%;
  border-top: 3px solid #64b5f6;
  width: 24px;
  height: 24px;
  animation: spin 1s linear infinite;
  display: inline-block;
  vertical-align: middle;
  margin-right: 10px;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.folder-header {
  display: flex;
  align-items: center;
  padding: 10px;
  background: var(--surface-color);
  border-bottom: 1px solid var(--border-color);
  margin-bottom: 10px;
}

.back-btn {
  background: none;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  padding: 5px 10px;
  cursor: pointer;
  margin-right: 15px;
  color: var(--text-color);
}

.back-btn:hover {
  background: var(--hover-color);
}

.current-path {
  font-weight: bold;
}

.menu-item.disabled {
  opacity: 0.5;
  cursor: default;
  pointer-events: none;
}
</style>