<template>
  <div class="public-share-container">
    <div v-if="loading" class="loading">
      Chargement...
    </div>
    
    <div v-else-if="error" class="error-message">
      <h2>Oups !</h2>
      <p>{{ error }}</p>
    </div>
    
    <div v-else class="share-card">
      <div class="file-icon">
        <span v-if="shareInfo.resource_type === 'folder'">
          <svg class="icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M10 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z" fill="#5f6368"/>
          </svg>
        </span>
        <span v-else>
          <svg class="icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z" fill="#1976d2"/>
          </svg>
        </span>
      </div>
      
      <h2 class="file-name">{{ shareInfo.resource_name }}</h2>
      
      <div class="file-details">
        <p v-if="shareInfo.resource_type === 'file'">
          Taille : {{ formatSize(shareInfo.file_size) }}
        </p>
        <p>Partagé par : {{ shareInfo.owner_email }}</p>
        <p v-if="shareInfo.expires_at">
          Expire le : {{ new Date(shareInfo.expires_at).toLocaleString() }}
        </p>
      </div>
      
      <button @click="downloadFile" class="btn-download">
        Télécharger
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '../api'
import { deriveKeyFromToken, unwrapMasterKey, decryptChunkedFileWorker } from '../utils/crypto'

const route = useRoute()
const router = useRouter()
const shareInfo = ref(null)
const loading = ref(true)
const error = ref(null)

const formatSize = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

onMounted(async () => {
  try {
    const token = route.params.token
    const response = await api.get(`/public/share/${token}`)
    shareInfo.value = response.data

    if (shareInfo.value.resource_type === 'folder') {
      router.replace({ name: 'PublicBrowse', params: { token: token, subpath: [] } })
    }

  } catch (err) {
    error.value = err.response?.data?.error || 'Lien invalide ou expiré.'
  } finally {
    loading.value = false
  }
})

const downloadFile = async () => {
  if (!shareInfo.value) return;
  const token = route.params.token;

  try {
    // 1. Derive Share Key from Token
    const shareKey = await deriveKeyFromToken(token);

    // 2. Decrypt File Key
    let fileKey;
    if (shareInfo.value.encrypted_key) {
        try {
          fileKey = await unwrapMasterKey(shareInfo.value.encrypted_key, shareKey);
        } catch(e) {
            console.error("Decryption error:", e);
            alert("Impossible de déchiffrer la clé du fichier.");
            return;
        }
    } else {
        alert("Ce fichier ne peut pas être déchiffré (clé manquante).");
        return;
    }

    // 3. Download Encrypted Blob
    const response = await api.get(`/public/share/${token}/download`, { responseType: 'blob' });
    
    // 4. Decrypt Blob
    const mimeType = shareInfo.value.mime_type || 'application/octet-stream';
    const decryptedBlob = await decryptChunkedFileWorker(response.data, fileKey, mimeType);
    
    // 5. Save
    const url = window.URL.createObjectURL(decryptedBlob);
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', shareInfo.value.resource_name);
    document.body.appendChild(link);
    link.click();
    setTimeout(() => { link.remove(); window.URL.revokeObjectURL(url); }, 100);
    
  } catch (err) {
    console.error("Download failed", err)
    alert("Erreur lors du téléchargement")
  }
}
</script>

<style scoped>
.public-share-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background-color: var(--background-color);
  font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
  color: var(--main-text-color);
}

.share-card {
  background: var(--card-color);
  padding: 40px;
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
  text-align: center;
  max-width: 500px;
  width: 90%;
  border: 1px solid var(--border-color);
}

.file-icon {
  margin-bottom: 20px;
  display: flex;
  justify-content: center;
}

.icon-svg {
  width: 64px;
  height: 64px;
}

.file-name {
  margin: 0 0 20px;
  color: var(--main-text-color);
  word-break: break-all;
}

.file-details {
  color: var(--secondary-text-color);
  margin-bottom: 30px;
  text-align: left;
  background: var(--background-color);
  padding: 15px;
  border-radius: 8px;
  border: 1px solid var(--border-color);
}

.file-details p {
  margin: 8px 0;
}

.btn-download {
  background-color: var(--primary-color);
  color: white;
  border: none;
  padding: 12px 30px;
  font-size: 16px;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.3s, transform 0.2s;
  width: 100%;
}

.btn-download:hover {
  background-color: var(--accent-color);
  transform: translateY(-2px);
}

.error-message {
  text-align: center;
  color: var(--error-color);
}
</style>
