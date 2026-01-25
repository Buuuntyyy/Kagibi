<template>
  <div v-if="visible" class="p2p-dialog-overlay">
    <div class="p2p-dialog">
      <div class="header">
        <h3>Transfert P2P</h3>
        <button v-if="canClose" @click="close" class="close-btn">&times;</button>
      </div>
      
      <!-- INCOMING REQUEST -->
      <div v-if="p2pStore.incomingOffer" class="content">
         <p><strong>{{ p2pStore.incomingOffer.senderId.substring(0,8) }}...</strong> souhaite envoyer un fichier.</p>
         <div class="file-details">
            <div>📄 {{ p2pStore.incomingOffer.name }}</div>
            <div>📦 {{ formatSize(p2pStore.incomingOffer.size) }}</div>
         </div>
         <div class="actions-row">
            <button @click="accept" class="btn accept">Recevoir</button>
            <button @click="reject" class="btn reject">Refuser</button>
         </div>
      </div>

      <!-- ACTIVE TRANSFER -->
      <div v-else-if="p2pStore.activeTransfer" class="content">
         <div class="status-row">
             <span>{{ statusText }}</span>
             <span class="pct">{{ p2pStore.activeTransfer.progress }}%</span>
         </div>
         <div class="progress-bar">
             <div class="fill" :style="{ width: p2pStore.activeTransfer.progress + '%' }"></div>
         </div>
         <p class="filename">{{ p2pStore.activeTransfer.fileName }}</p>
         
         <div class="actions-row" v-if="isDone">
            <button @click="close" class="btn">Fermer</button>
         </div>
         <div class="actions-row" v-else>
             <button @click="cancel" class="btn text-danger">Annuler</button>
         </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { useP2PStore } from '../stores/p2p';

const p2pStore = useP2PStore();

const visible = computed(() => !!p2pStore.incomingOffer || !!p2pStore.activeTransfer);
const isDone = computed(() => p2pStore.activeTransfer?.status === 'Done' || p2pStore.activeTransfer?.status === 'Complete');
const canClose = computed(() => isDone.value || !!p2pStore.incomingOffer);

const statusText = computed(() => {
    if (!p2pStore.activeTransfer) return '';
    return p2pStore.activeTransfer.status;
});

const formatSize = (bytes) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Number.parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

const accept = () => p2pStore.acceptTransfer();
const reject = () => p2pStore.rejectTransfer();
const cancel = () => p2pStore.cancelTransfer();
const close = () => {
    if(p2pStore.incomingOffer) p2pStore.rejectTransfer();
    else if(p2pStore.activeTransfer && isDone.value) p2pStore.activeTransfer = null;
    else p2pStore.cancelTransfer(); 
};
</script>

<style scoped>
.p2p-dialog-overlay {
    position: fixed;
    bottom: 20px;
    right: 20px;
    z-index: 2000;
}
.p2p-dialog {
    background: white;
    width: 320px;
    border-radius: 8px;
    box-shadow: 0 4px 15px rgba(0,0,0,0.25);
    border: 1px solid #ddd;
    overflow: hidden;
    font-family: sans-serif;
}
.header {
    background: #f5f5f5;
    padding: 10px 15px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-bottom: 1px solid #eee;
}
.header h3 { margin: 0; font-size: 1rem; color: #333; }
.close-btn { background: none; border: none; font-size: 1.2rem; cursor: pointer; color: #666; }

.content { padding: 15px; }
.file-details {
    background: #f9f9f9;
    padding: 10px;
    border-radius: 4px;
    margin: 10px 0;
    font-size: 0.85rem;
    color: #555;
    border: 1px solid #eee;
}
.actions-row { display: flex; gap: 10px; justify-content: flex-end; margin-top: 10px; }
.btn { padding: 8px 16px; border: none; border-radius: 4px; cursor: pointer; font-size: 0.9rem; font-weight: 500; }
.accept { background: #42b983; color: white; }
.reject { background: #ff5252; color: white; }
.text-danger { color: #ff5252; background: none; text-decoration: underline; padding: 5px; }

.progress-bar { height: 8px; background: #eee; border-radius: 4px; overflow: hidden; margin: 10px 0; }
.fill { height: 100%; background: #42b983; transition: width 0.2s; }
.status-row { display: flex; justify-content: space-between; font-size: 0.85rem; color: #666; }
.filename { font-weight: bold; margin: 5px 0; font-size: 0.9rem; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; color: #333; }
</style>