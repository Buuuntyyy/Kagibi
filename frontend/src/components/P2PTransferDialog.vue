<template>
  <div v-if='visible' class='p2p-notification-container'>
    <div class='p2p-card'>
      <div class='card-header'>
        <h3 class='header-title'>
          <svg xmlns='http://www.w3.org/2000/svg' width='18' height='18' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'><circle cx='12' cy='12' r='10'></circle><line x1='2' y1='12' x2='22' y2='12'></line><path d='M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z'></path></svg>
          Transfert P2P
        </h3>
        <button v-if='canClose' @click='close' class='close-icon'>&times;</button>
      </div>
      
      <!-- INCOMING REQUEST -->
      <div v-if='p2pStore.incomingOffer' class='notification-body'>
         <p class='request-text'>
            <strong>{{ p2pStore.incomingOffer.senderId.substring(0,8) }}...</strong> souhaite vous envoyer un fichier.
         </p>
         <div class='file-preview'>
            <div class='file-icon-box'>
                <svg xmlns='http://www.w3.org/2000/svg' width='24' height='24' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'><path d='M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z'></path><polyline points='13 2 13 9 20 9'></polyline></svg>
            </div>
            <div class='file-info'>
                <div class='f-name' :title='p2pStore.incomingOffer.name'>{{ p2pStore.incomingOffer.name }}</div>
                <div class='f-size'>{{ formatSize(p2pStore.incomingOffer.size) }}</div>
            </div>
         </div>
         <div class='actions-grid'>
            <button @click='reject' class='btn btn-secondary'>Refuser</button>
            <button @click='accept' class='btn btn-primary'>Recevoir</button>
         </div>
      </div>

      <!-- ACTIVE TRANSFER -->
      <div v-else-if='p2pStore.activeTransfer' class='notification-body'>
         <div class='status-header'>
             <span class='status-label'>{{ statusText }}</span>
             <span class='pct-badge'>{{ p2pStore.activeTransfer.progress }}%</span>
         </div>
         <div class='progress-track'>
             <div class='progress-fill' :style='{ width: p2pStore.activeTransfer.progress + "%" }'></div>
         </div>
         <p class='filename-display' :title='p2pStore.activeTransfer.fileName'>{{ p2pStore.activeTransfer.fileName }}</p>
         
         <div class='actions-grid single' v-if='isDone'>
            <button @click='close' class='btn btn-primary'>Fermer</button>
         </div>
         <div class='actions-grid single' v-else>
             <button @click='cancel' class='btn btn-danger-text'>Annuler</button>
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
.p2p-notification-container {
    position: fixed;
    bottom: 2rem;
    right: 2rem;
    z-index: 9999;
    /* Ensure no full height taking */
    height: auto;
    width: auto;
}

.p2p-card {
    background: var(--card-color, #ffffff);
    width: 340px;
    border-radius: 12px;
    box-shadow: 0 8px 30px rgba(0,0,0,0.12);
    border: 1px solid var(--border-color, #e0e0e0);
    overflow: hidden;
    font-family: inherit;
    animation: slideIn 0.3s ease-out;
}

@keyframes slideIn {
    from { transform: translateY(20px); opacity: 0; }
    to { transform: translateY(0); opacity: 1; }
}

.card-header {
    background: var(--background-color, #f8f9fa);
    padding: 12px 16px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-bottom: 1px solid var(--border-color, #eaeaea);
}

.header-title {
    margin: 0;
    font-size: 0.95rem;
    font-weight: 600;
    color: var(--main-text-color, #333);
    display: flex;
    align-items: center;
    gap: 8px;
}
.header-title svg { color: var(--primary-color, #3498db); }

.close-icon {
    background: none;
    border: none;
    font-size: 1.4rem;
    line-height: 1;
    cursor: pointer;
    color: var(--secondary-text-color, #888);
    padding: 0;
}
.close-icon:hover { color: var(--main-text-color, #333); }

.notification-body {
    padding: 16px;
}

.request-text {
    margin: 0 0 12px 0;
    font-size: 0.9rem;
    color: var(--main-text-color, #444);
    line-height: 1.4;
}

.file-preview {
    display: flex;
    align-items: center;
    gap: 12px;
    background: var(--hover-background-color, #f4f6f8);
    padding: 10px;
    border-radius: 8px;
    margin-bottom: 16px;
}

.file-icon-box {
    color: var(--secondary-text-color, #666);
    display: flex;
}

.file-info {
    flex: 1;
    overflow: hidden;
}

.f-name {
    font-weight: 600;
    font-size: 0.9rem;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    color: var(--main-text-color, #333);
}

.f-size {
    font-size: 0.75rem;
    color: var(--secondary-text-color, #888);
}

.actions-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
}
.actions-grid.single {
    grid-template-columns: 1fr;
}

.btn {
    padding: 8px;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    font-size: 0.9rem;
    font-weight: 600;
    transition: all 0.2s;
}

.btn-primary {
    background: var(--primary-color, #3498db);
    color: white;
}
.btn-primary:hover {
    filter: brightness(1.1);
}

.btn-secondary {
    background: transparent;
    border: 1px solid var(--border-color, #ddd);
    color: var(--secondary-text-color, #666);
}
.btn-secondary:hover {
    background: var(--hover-background-color, #f5f5f5);
    color: var(--main-text-color, #333);
}

.btn-danger-text {
    background: none;
    color: var(--error-color, #e74c3c);
    text-decoration: underline;
}

/* Progress styles */
.progress-track {
    height: 6px;
    background: var(--border-color, #eee);
    border-radius: 3px;
    overflow: hidden;
    margin: 8px 0 12px 0;
}
.progress-fill {
    height: 100%;
    background: var(--success-color, #2ecc71);
    transition: width 0.3s ease;
}

.status-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 0.85rem;
    color: var(--secondary-text-color, #666);
    margin-bottom: 4px;
}
.pct-badge {
    font-weight: 700;
    color: var(--primary-color, #3498db);
}
.filename-display {
    font-size: 0.85rem;
    color: var(--main-text-color, #333);
    margin: 0 0 16px 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}
</style>
