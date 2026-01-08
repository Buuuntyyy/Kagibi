<template>
  <div class="shared-with-me-container">
    <div v-if="loading" class="loading">
      <div class="spinner"></div> Chargement...
    </div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="items.length === 0" class="empty">
      <p>Aucun fichier partagé avec vous.</p>
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
        <span class="file-link" @click.stop="handleOpenFile(item)" :title="item.name">{{ item.name }}</span>
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
              <span class="menu-icon">⬇️</span> Télécharger (Déchiffrer)
            </div>
             <div class="menu-divider"></div>
            <div class="menu-item delete" @click="handleContextAction('delete')">
              <span class="menu-icon">🗑️</span> Retirer ce partage
            </div>
        </template>
    </ContextMenu>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, watch } from 'vue';
import { useRouter } from 'vue-router';
import FileTable from './file/FileTable.vue';
import { formatSize, formatDate } from '../utils/format';
import api from '../api';
import ContextMenu from './file/ContextMenu.vue';
import { useFileStore } from '../stores/files';
import { useAuthStore } from '../stores/auth';
import { decryptKeyWithPrivateKey, importKeyFromPEM, decryptChunkedFileWorker } from '../utils/crypto';
import sodium from 'libsodium-wrappers-sumo';

const fileStore = useFileStore();
const authStore = useAuthStore();
const router = useRouter();
const items = ref([]);
const loading = ref(false);
const error = ref(null);

watch(() => fileStore.shareUpdateTrigger, () => {
    fetchSharedWithMe();
});

const contextMenu = ref({
  visible: false,
  x: 0,
  y: 0,
  item: null
});

const columns = [
  { key: 'icon', label: '', headerClass: 'icon-col', cellClass: 'icon-col' },
  { key: 'shared_name', label: 'Nom', cellClass: 'name-cell' },
  { key: 'owner', label: 'Propriétaire' },
  { key: 'shared_at', label: 'Partagé le' },
  { key: 'size', label: 'Taille' },
  { key: 'actions', label: 'Actions' }
]

const sharedFolders = computed(() => items.value.filter(i => i.type === 'folder'))
const sharedFiles = computed(() => items.value.filter(i => i.type === 'file'))

const fetchSharedWithMe = async () => {
  loading.value = true;
  try {
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
        owner_name: share.owner_name || 'Inconnu',
        // Determine if direct share for context actions
        is_direct: !!share.file_id || !!share.folder_id,
        // Helper specifically for download
        resource_id: share.file_id || share.folder_id
    }));
  } catch (err) {
    console.error("Error fetching shared with me:", err);
    error.value = "Impossible de charger les fichiers partagés avec vous.";
  } finally {
    loading.value = false;
  }
};

const handleContextMenu = (event, item, type) => {
  event.preventDefault();
  contextMenu.value = {
    visible: true,
    x: event.clientX,
    y: event.clientY,
    item: { ...item, type }
  };
};

const closeContextMenu = () => {
    contextMenu.value.visible = false;
}

const handleContextAction = async (action) => {
    const item = contextMenu.value.item;
    if (!item) return;

    if (action === 'download') {
        if (item.type === 'file') {
             await downloadSharedFile(item);
        }
    } else if (action === 'delete') {
         if (confirm("Voulez-vous retirer ce partage de votre liste ?")) {
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
                 alert("Erreur lors de la suppression");
             }
         }
    }
    closeContextMenu();
}

const handleOpenFolder = (folderName) => {
    // Navigate to shared folder view? 
    // Usually implies recursively listing contents.
    // For now: Alert not implemented
    alert("Ouverture de dossier partagé pas encore implémentée complètement.");
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
        // The item.encrypted_key is the FileKey encrypted with MY Public Key.
        if (!item.encrypted_key) {
            throw new Error("Clé de chiffrement manquante pour ce partage. S'il s'agit d'un ancien partage, veuillez demander à l'expéditeur de le partager à nouveau.");
        }
        
        await sodium.ready;
        
        // Decrypt User Private Key first if not ready (AuthStore handles this usually)
        if (!authStore.privateKey) {
             throw new Error("Clé privée non disponible.");
        }

        // Decrypt the Shared Key using my Private Key
        // item.encrypted_key is base64.
        // We use our RSA Private Key to decrypt it.
        const encryptedKeyBytes = sodium.from_base64(item.encrypted_key);
        
        // Use WebCrypto for RSA-OAEP decryption
        // Need to import private key to WebCrypto format if stored as PEM/other
        // authStore.privateKey is usually the PEM string or CryptoKey? 
        // Let's assume authStore stores the CryptoKey object for private key if loaded, 
        // OR we re-import it. 
        // Checked authStore: it stores 'privateKey' as ... wait, let's check authStore usage.
        // Actually authStore keeps encrypted_private_key string.
        // The decrypted private key is usually kept in memory or session?
        // Let's assume we use the helper 'decryptKeyWithPrivateKey' if available or do it manually.
        
        // wait, `decryptKeyWithPrivateKey` is for symmetric.
        // We need RSA decryption.
        
        // In DirectShareModal we encrpyted with:
        // window.crypto.subtle.encrypt({ name: "RSA-OAEP" }, publicKey, rawData)
        
        // So here we need:
        // window.crypto.subtle.decrypt({ name: "RSA-OAEP" }, privateKey, encryptedData)
        
        // We need the private key object.
        // If authStore doesn't expose it, we might need to derive it again or hopefully it is cached.
        
        // RE-CHECK: AuthStore usually keeps 'masterKey' (AES-GCM).
        // The RSA Private Key is stored ENCRYPTED in DB.
        // on login, we decrypt RSA Priv Key using Master Key.
        
        const rsaPrivateKey = authStore.privateKey; // Uses the correct store property
        if (!rsaPrivateKey) {
             throw new Error("Clé RSA privée non chargée. Reconnectez-vous.");
        }

        const fileKeyRawBuffer = await window.crypto.subtle.decrypt(
            { name: "RSA-OAEP" },
            rsaPrivateKey,
            encryptedKeyBytes
        );

        // 2. We have the File Key (Raw). Now download and decrypt the file content.
        // We can reuse fileStore.downloadFile but we need to pass the key explicitly 
        // OR manually fetch blob and decrypt.
        // fileStore.downloadFile expects file ID and looks up key in store... which won't work here.
        
        // We need a custom download function for shared files where we supply the key.
        
        const response = await api.get(`/files/download/${item.resource_id}`, { responseType: 'blob' });
        
        // 3. Decrypt Content
        // File content is encrypted with params (IV + Data).
        // Key is fileKeyRawBuffer.
        
        // Convert Blob to ArrayBuffer
        const encryptedFileBytes = await response.data.arrayBuffer();
        
        // Extract Nonce (24 bytes for XChaCha20, or 12 for AES-GCM?)
        // Backend uses: stream.NewEncrypter(key, make([]byte, 12)) for AES-GCM usually?
        // Let's check backend pkg/s3storage or upload handler.
        // Assumptions: AES-GCM standard (12 byte IV appended or prepended).
        // Let's assume standard format used in project: IV + Ciphertext.
        
        // WAIT: Browser decryption usually uses window.crypto.subtle (AES-GCM).
        
        // Let's import the File Key for AES-GCM
        const fileKeyCrypto = await window.crypto.subtle.importKey(
            "raw", 
            fileKeyRawBuffer,
            "AES-GCM",
            true,
            ["decrypt"]
        );
        
        // 4. Decrypt Content using Worker (Handles chunked files correctly)
        // Convert ArrayBuffer to Blob for the worker helper
        const encryptedBlob = new Blob([encryptedFileBytes]);
        
        const decryptedBlob = await decryptChunkedFileWorker(encryptedBlob, fileKeyCrypto, item.mime_type || 'application/octet-stream');
        
        // 5. Trigger Download
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
        alert("Erreur lors du téléchargement/déchiffrement : " + e.message);
    }
}

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
</style>