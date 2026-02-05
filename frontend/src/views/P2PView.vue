<template>
  <div class="p2p-page">
    <div class="p2p-container">
      <div class="main-card">
        <div class="card-header">
           <h2>Transfert P2P Sécurisé</h2>
           <p>Envoyez des fichiers directement à vos amis connectés, sans passer par le serveur.</p>
        </div>

        <div class="transfer-flow">
            <!-- STEP 1: DESTINATAIRE -->
            <div class="flow-step" :class="{ active: !selectedFriend, completed: selectedFriend }">
                <div class="step-icon">1</div>
                <div class="step-content">
                    <h3>Destinataire</h3>
                    <div v-if="!selectedFriend">
                        <p class="secondary-text">Sélectionnez un ami en ligne</p>
                        <div class="friend-selector">
                           <div v-if="onlineFriends.length === 0" class="no-friends">
                               Aucun ami en ligne
                           </div>
                           <div v-else class="friends-grid">
                               <div 
                                  v-for="friend in onlineFriends" 
                                  :key="friend.id" 
                                  class="friend-chip"
                                  @click="selectedFriend = friend"
                               >
                                  <div class="avatar-mini-wrapper">
                                      <div class="avatar-mini">
                                          <img 
                                            v-if="normalizeAvatarUrl(friend.avatar_url)" 
                                            :src="normalizeAvatarUrl(friend.avatar_url)" 
                                            :alt="friend.name"
                                            class="avatar-image"
                                            @error="(e) => e.target.style.display = 'none'"
                                          />
                                          <span v-else class="avatar-initials-mini">{{ getInitials(friend.name) }}</span>
                                      </div>
                                      <span class="status-dot-mini"></span>
                                  </div>
                                  <span>{{ friend.name }}</span>
                               </div>
                           </div>
                        </div>
                    </div>
                    <div v-else class="selected-friend-display" style="width: fit-content">
                        <div class="friend-chip selected">
                            <div class="avatar-mini-wrapper">
                                <div class="avatar-mini">
                                    <img 
                                      v-if="normalizeAvatarUrl(selectedFriend.avatar_url)" 
                                      :src="normalizeAvatarUrl(selectedFriend.avatar_url)" 
                                      :alt="selectedFriend.name"
                                      class="avatar-image"
                                      @error="(e) => e.target.style.display = 'none'"
                                    />
                                    <span v-else class="avatar-initials-mini">{{ getInitials(selectedFriend.name) }}</span>
                                </div>
                                <span class="status-dot-mini"></span>
                            </div>
                            <span>{{ selectedFriend.name }}</span>
                            <button class="close-btn" @click="selectedFriend = null">×</button>
                        </div>
                    </div>
                </div>
            </div>

            <!-- STEP 2: FICHIER -->
            <div class="flow-step" :class="{ active: selectedFriend && !selectedFile, completed: selectedFile, disabled: !selectedFriend }">
                <div class="step-icon">2</div>
                <div class="step-content">
                    <h3>Fichier</h3>
                    <input type="file" id="p2p-file-input" @change="handleFileSelect" style="display: none" />
                    
                    <div v-if="!selectedFile" 
                         class="drop-area" 
                         :class="{ 'pulse': selectedFriend && !selectedFile }"
                         @click="selectedFriend && triggerFileSelect()"
                         @dragover.prevent
                         @drop.prevent="handleDrop"
                    >
                        <span v-if="!selectedFriend">Sélectionnez un destinataire d'abord</span>
                        <span v-else>Cliquez ou glissez un fichier ici</span>
                    </div>

                    <div v-else class="file-display">
                        <div class="file-icon">
                            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"></path><polyline points="13 2 13 9 20 9"></polyline></svg>
                        </div>
                        <div class="file-meta">
                            <span class="fname" :title="selectedFile.name">{{ selectedFile.name }}</span>
                            <span class="fsize">{{ formatSize(selectedFile.size) }}</span>
                        </div>
                        <button class="change-file-btn" @click="selectedFile = null">Changer</button>
                    </div>
                </div>
            </div>

            <!-- STEP 3: ACTION -->
            <div class="flow-step action-step">
                <button class="send-big-btn" :disabled="!canSend" @click="startTransfer">
                    <span v-if="!canSend">En attente...</span>
                    <span v-else>
                        Envoyer le fichier
                        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="margin-left:8px;"><line x1="22" y1="2" x2="11" y2="13"></line><polygon points="22 2 15 22 11 13 2 9 22 2"></polygon></svg>
                    </span>
                </button>
            </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useFriendStore } from '../stores/friends'
import { useAuthStore } from '../stores/auth'
import { useP2PStore } from '../stores/p2p'

const friendStore = useFriendStore()
const authStore = useAuthStore()
const p2pStore = useP2PStore()

const selectedFriend = ref(null)
const selectedFile = ref(null)

const onlineFriends = computed(() => {
  return friendStore.acceptedFriends.filter(f => f.online)
})

const canSend = computed(() => !!selectedFriend.value && !!selectedFile.value)

onMounted(() => {
    friendStore.fetchFriends()
})

const getInitials = (name) => {
    if (!name) return '?'
    return name.substring(0, 2).toUpperCase()
}

const normalizeAvatarUrl = (url) => {
  if (!url) return null
  if (url.startsWith('http')) return url
  if (url.startsWith('/avatars/')) return url
  const cleanUrl = url.startsWith('/') ? url.substring(1) : url
  return `/avatars/${cleanUrl}`
}

const formatSize = (bytes) => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const triggerFileSelect = () => {
    document.getElementById('p2p-file-input').click()
}

const handleFileSelect = (e) => {
    if (e.target.files.length > 0) {
        selectedFile.value = e.target.files[0]
    }
}

const handleDrop = (e) => {
    if (!selectedFriend.value) return
    if (e.dataTransfer.files.length > 0) {
        selectedFile.value = e.dataTransfer.files[0]
    }
}

const startTransfer = async () => {
    if (!canSend.value) return
    try {
        await p2pStore.startTransfer(selectedFriend.value, selectedFile.value)
        // Reset after send? Or keep?
        // Let's keep for now so user sees feedback or can send again
    } catch (e) {
        console.error("Transfer failed", e)
        alert("Erreur: " + e.message)
    }
}
</script>

<style scoped>
.p2p-page {
    background-color: var(--background-color);
    min-height: 100vh;
    display: flex;
    flex-direction: column;
}

.p2p-container {
    flex: 1;
    display: flex;
    justify-content: center;
    align-items: flex-start;
    padding-top: 2rem;
    padding-bottom: 2rem;
    overflow-y: auto;
}

.main-card {
    background: var(--card-color);
    border: 1px solid var(--border-color);
    border-radius: 16px;
    width: 600px;
    padding: 2.5rem;
    box-shadow: 0 4px 20px rgba(0,0,0,0.04);
    margin-bottom: 2rem;
}

.card-header {
    text-align: center;
    margin-bottom: 3rem;
}
.card-header h2 { margin: 0 0 0.5rem 0; color: var(--main-text-color); }
.card-header p { margin: 0; color: var(--secondary-text-color); }

.transfer-flow {
    display: flex;
    flex-direction: column;
    gap: 2rem;
}

.flow-step {
    display: flex;
    gap: 1.5rem;
    padding-bottom: 2rem;
    border-bottom: 1px dashed var(--border-color);
}
.flow-step:last-child {
    border-bottom: none;
    padding-bottom: 0;
}
.flow-step.disabled {
    opacity: 0.5;
    pointer-events: none;
}

.step-icon {
    width: 40px; height: 40px;
    background: var(--border-color);
    color: var(--secondary-text-color);
    font-weight: 700;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
}

.flow-step.active .step-icon {
    background: var(--primary-color);
    color: white;
    box-shadow: 0 0 0 4px rgba(52, 152, 219, 0.2);
}

.flow-step.completed .step-icon {
    background: var(--success-color);
    color: white;
}

.step-content {
    flex: 1;
    min-width: 0;
}
.step-content h3 {
    margin: 0 0 0.5rem 0;
    font-size: 1.1rem;
    font-weight: 600;
}

.secondary-text {
    font-size: 0.9rem;
    color: var(--secondary-text-color);
    margin-bottom: 0.5rem;
}

.friends-grid {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
}

.friend-chip {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.4rem 0.8rem;
    background: var(--hover-background-color);
    border-radius: 20px;
    cursor: pointer;
    border: 1px solid transparent;
    transition: all 0.2s;
}
.friend-chip:hover {
    border-color: var(--primary-color);
}
.friend-chip.selected {
    background: var(--primary-color);
    color: white;
}

.avatar-mini {
    width: 24px; height: 24px;
    background: rgba(255,255,255,0.3);
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.7rem;
    overflow: hidden;
    position: relative;
}

.avatar-mini-wrapper {
    position: relative;
    display: inline-block;
}

.avatar-image {
    width: 100%;
    height: 100%;
    object-fit: cover;
    border-radius: 50%;
}

.avatar-initials-mini {
    color: #5f6368;
    font-size: 0.7rem;
    font-weight: 500;
}

.status-dot-mini {
    position: absolute;
    bottom: -2px;
    right: -2px;
    width: 8px;
    height: 8px;
    background-color: #34a853;
    border-radius: 50%;
    border: 2px solid white;
    z-index: 1;
}

.drop-area {
    border: 2px dashed var(--border-color);
    border-radius: 8px;
    padding: 2rem;
    text-align: center;
    color: var(--secondary-text-color);
    cursor: pointer;
    transition: all 0.2s;
}
.drop-area:hover, .drop-area.pulse {
    border-color: var(--primary-color);
    background: var(--hover-background-color);
    color: var(--primary-color);
}

.file-display {
    display: flex;
    align-items: center;
    gap: 1rem;
    background: var(--hover-background-color);
    padding: 1rem;
    border-radius: 8px;
    border-left: 4px solid var(--primary-color);
}
.file-icon { font-size: 1.5rem; }
.file-meta { flex: 1; display:flex; flex-direction:column; overflow: hidden; min-width: 0; }
.fname { font-weight: 600; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.fsize { font-size: 0.85rem; color: var(--secondary-text-color); }
.change-file-btn { font-size: 0.8rem; padding: 0.2rem 0.6rem; }
.close-btn { background:transparent; border:none; color:white; font-size:1.2rem; margin-left:0.5rem; cursor:pointer;}

.send-big-btn {
    width: 100%;
    padding: 1rem;
    font-size: 1.1rem;
    background: var(--success-color);
    color: white;
    border: none;
    border-radius: 12px;
    font-weight: 700;
    cursor: pointer;
    box-shadow: 0 4px 10px rgba(39, 174, 96, 0.3);
    transition: transform 0.1s;
}
.send-big-btn:hover:not(:disabled) {
    transform: translateY(-2px);
}
.send-big-btn:disabled {
    background: var(--border-color);
    box-shadow: none;
    cursor: not-allowed;
    color: var(--secondary-text-color);
}
</style>