<template>
  <div class="file-shared">
    <div class="accordion-header" @click="toggleAccordion">
      <span class="chevron" :class="{ 'open': isOpen }">
        <svg xmlns="http://www.w3.org/2000/svg" height="24px" viewBox="0 0 24 24" width="24px" fill="currentColor"><path d="M0 0h24v24H0V0z" fill="none"/><path d="M10 6L8.59 7.41 13.17 12l-4.58 4.59L10 18l6-6z"/></svg>
      </span>
      <h4 class="section-title">Partagés avec moi</h4>
    </div>
    
    <div v-show="isOpen" class="accordion-content">
      <div v-if="loading" class="loading-state">
        <span>Chargement...</span>
      </div>
      <div v-else-if="sharedItems.length > 0" class="cards-row">
        <div 
          v-for="(item, index) in sharedItems" 
          :key="index" 
          class="recent-card"
          :class="item.type"
          @click="openItem(item)"
        >
          <div class="icon-wrapper">
             <!-- Folder Icon -->
            <svg v-if="item.type === 'folder'" viewBox="0 0 24 24" width="24" height="24" fill="currentColor">
              <path d="M10 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z"></path>
            </svg>
             <!-- File Icon -->
            <svg v-else viewBox="0 0 24 24" width="24" height="24" fill="currentColor">
               <path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z"/>
            </svg>
          </div>
          
          <div class="card-details">
            <span class="item-name" :title="item.displayName">{{ item.displayName }}</span>
            <span class="item-info">{{ item.owner_name }}</span>
          </div>
        </div>
      </div>
      
      <div v-else class="empty-state">
        <span class="empty-icon">📂</span>
        <span>Aucun fichier partagé</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router';
import api from '../../api'
import { useAuthStore } from '../../stores/auth'
import { useFileStore } from '../../stores/files'
import sodium from 'libsodium-wrappers-sumo'
import { decryptChunkedFileWorker } from '../../utils/crypto'

const router = useRouter()
const authStore = useAuthStore()
const fileStore = useFileStore()
const isOpen = ref(true)
const loading = ref(true)
const sharedItems = ref([])

const toggleAccordion = () => {
  isOpen.value = !isOpen.value
}

const fetchSharedItems = async () => {
  loading.value = true;
  try {
    const response = await api.get('/shares/with-me');
    // Limit to 5 recent shares
    sharedItems.value = (response.data || []).slice(0, 5).map(share => ({
        ...share,
        displayName: share.name,
        // Ensure type is lowercase for class matching logic if any
        type: share.type || (share.file_id ? 'file' : 'folder'),
        resource_id: share.file_id || share.folder_id
    }));
  } catch (err) {
    console.error("Error fetching shared items:", err);
  } finally {
    loading.value = false;
  }
}

const openItem = async (item) => {
  if (item.type === 'folder') {
    // 1. Open in shared mode specifically
    await fileStore.openSharedRoot(item);
    
    // 2. Navigate to file list view to display it
    // We navigate to /dashboard/files BUT fileStore.viewMode won't be reset
    // because fetchItems('/') check in store only resets if path is '/'
    // We should make sure we land on nothing that resets it.
    router.push({ name: 'MyFiles' });
  } else {
    // Attempt download
    if (item.link) {
        window.open(item.link, '_blank');
        return;
    }
    await downloadSharedFile(item);
  }
}

const downloadSharedFile = async (item) => {
    try {
        await sodium.ready;
        
        let fileKeyCrypto;

        // Root Share (Direct)
        if (!item.encrypted_key) {
             console.error("Clé de chiffrement manquante.");
             alert("Impossible d'ouvrir le fichier : clé manquante.");
             return;
        }
        
        if (!authStore.privateKey) {
             console.error("Clé privée non disponible.");
             alert("Clé privée non disponible.");
             return;
        }

        const encryptedKeyBytes = sodium.from_base64(item.encrypted_key);
        const rsaPrivateKey = authStore.privateKey;

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

        // Download and Decrypt Content
        const response = await api.get(`/files/download/${item.resource_id}`, { responseType: 'blob' });
        const encryptedFileBytes = await response.data.arrayBuffer();
        const encryptedBlob = new Blob([encryptedFileBytes]);
        
        const decryptedBlob = await decryptChunkedFileWorker(encryptedBlob, fileKeyCrypto, item.mime_type || 'application/octet-stream');
        
        // Trigger Download
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
  fetchSharedItems()
})
</script>

<style scoped>
.file-shared {
  padding: 0.5rem 1rem 0 1rem;
  background-color: var(--card-color);
  margin-bottom: 0.5rem;
}

.accordion-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  padding: 0.5rem;
  margin-left: -0.5rem;
  user-select: none;
  border-radius: 15px;
  transition: background-color 0.2s;
  width: fit-content;
}

.accordion-header:hover {
  background-color: var(--hover-color, #f0f0f0);
}

.chevron {
  display: flex;
  align-items: center;
  transition: transform 0.3s ease;
  color: #666;
}

.chevron.open {
  transform: rotate(90deg);
}

.section-title {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-color);
}

.accordion-content {
  padding-top: 0.5rem;
  padding-bottom: 1rem;
  /* animation: slideDown 0.3s ease-out; */
}

.cards-row {
  display: flex;
  flex-wrap: wrap; /* Allow wrapping if many items */
  gap: 1rem;
}

.recent-card {
  display: flex;
  flex-direction: column;
  width: 140px;
  height: 120px;
  background-color: var(--background-color, #f9f9f9);
  border: 1px solid #eee;
  border-radius: 12px;
  padding: 0.8rem;
  cursor: pointer;
  transition: all 0.2s ease;
  position: relative;
  overflow: hidden;
}

.recent-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0,0,0,0.08);
  border-color: var(--primary-color, #42b983);
}

.icon-wrapper {
  flex-grow: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #5f6368;
  margin-bottom: 0.5rem;
}

.recent-card.folder .icon-wrapper {
  color: #5f6368;
}

.recent-card.file .icon-wrapper {
  color: var(--primary-color, #42b983);
}

.icon-wrapper svg {
  width: 48px;
  height: 48px;
}

.card-details {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.item-name {
  font-weight: 500;
  font-size: 0.9rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  color: var(--text-color);
}

.item-info {
  font-size: 0.75rem;
  color: #888;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.empty-state {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #888;
  font-style: italic;
  font-size: 0.9rem;
  padding: 1rem 0;
}

.loading-state {
    padding: 1rem 0;
    color: #888;
    font-size: 0.9rem;
}
</style>
