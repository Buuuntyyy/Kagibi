<template>
  <div class="p2p-container" ref="containerRef">
    <div class="p2p-header-minimal">
        <h2>Transfert P2P</h2>
    </div>

    <!-- Canvas for Particles -->
    <canvas ref="canvasRef" class="particle-canvas"></canvas>

    <div class="p2p-layout">
        <!-- LEFT: Current User -->
        <div class="user-zone" ref="userZoneRef">
            <div class="zone-label">MON COMPTE</div>
            <div class="large-avatar pulse-effect">
                {{ getInitials(authStore.user?.name) }}
            </div>
            <p class="user-name">{{ authStore.user?.name || 'Moi' }}</p>
        </div>

        <!-- CENTER: Action Button -->
        <div class="center-zone" ref="centerZoneRef">
             <!-- Spacer for alignment -->
             <div class="zone-label" style="opacity: 0">ACTION</div>
            <input type="file" id="p2p-file-input" @change="handleFileSelect" style="display: none" />
            
            <div style="position: relative">
                <div 
                    class="action-circle" 
                    :class="{ 'has-file': !!selectedFile, 'ready-to-send': canSend }"
                    @click="handleActionClick"
                    @dragover.prevent="dragOver = true" 
                    @dragleave.prevent="dragOver = false" 
                    @drop.prevent="handleDrop"
                >
                    <div v-if="!selectedFile" class="plus-icon">+</div>
                    
                    <div v-else-if="!canSend" class="file-state">
                        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" class="file-icon">
                            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" stroke="currentColor" stroke-width="2"/>
                            <polyline points="14 2 14 8 20 8" stroke="currentColor" stroke-width="2"/>
                        </svg>
                        <span class="filename">{{ truncate(selectedFile.name) }}</span>
                    </div>

                    <div v-else class="send-state">
                        <span class="send-text">ENVOYER</span>
                        <svg width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <line x1="22" y1="2" x2="11" y2="13"></line>
                            <polygon points="22 2 15 22 11 13 2 9 22 2"></polygon>
                        </svg>
                    </div>
                </div>

                <transition name="fade">
                    <button v-if="selectedFile" class="close-file" @click.stop="removeFile" title="Changer de fichier">
                         <svg viewBox="0 0 24 24" width="21" height="21" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">
                            <line x1="18" y1="6" x2="6" y2="18"></line>
                            <line x1="6" y1="6" x2="18" y2="18"></line>
                        </svg>
                    </button>
                </transition>
            </div>

            <p class="action-label" v-if="!selectedFile">Sélectionner un fichier</p>
            <p class="action-label" v-else>
                 {{ canSend ? `Prêt à envoyer à ${selectedFriend.name}` : 'En attente d\'un destinataire...' }}
            </p>
        </div>

        <!-- RIGHT: Friends List or Selected Friend -->
        <div class="friends-zone" ref="friendsZoneRef">
            
            <!-- Case: Friend Selected -->
            <div v-if="selectedFriend" class="selected-friend-view">
                <div class="zone-label">DESTINATAIRE</div>
                <div style="position: relative;">
                    <div class="large-avatar friend-avatar pulse-effect">
                        {{ getInitials(selectedFriend.name) }}
                    </div>
                    <button class="close-friend" @click="deselectFriend" title="Fermer">
                         <svg viewBox="0 0 24 24" width="21" height="21" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">
                            <line x1="18" y1="6" x2="6" y2="18"></line>
                            <line x1="6" y1="6" x2="18" y2="18"></line>
                        </svg>
                    </button>
                </div>
                <p class="user-name">{{ selectedFriend.name }}</p>
            </div>

            <!-- Case: Friends List -->
            <div v-else class="friends-list-box">
                <div class="list-header zone-label">
                    <span>AMIS EN LIGNE</span>
                    <span class="count">{{ onlineFriends.length }}</span>
                </div>
                
                <div v-if="onlineFriends.length === 0" class="empty-list">
                    <p>Personne en vue...</p>
                </div>
                
                <div class="scrollable-list" v-else>
                     <div 
                        v-for="friend in onlineFriends" 
                        :key="friend.id" 
                        class="friend-row"
                        @click="selectFriend(friend)"
                     >
                        <div class="mini-avatar">{{ getInitials(friend.name) }}</div>
                        <span class="mini-name">{{ friend.name }}</span>
                        <div class="mini-status"></div>
                     </div>
                </div>
            </div>

        </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useFriendStore } from '../stores/friends'
import { useAuthStore } from '../stores/auth'
import { useP2PStore } from '../stores/p2p'

const friendStore = useFriendStore()
const p2pStore = useP2PStore()
const authStore = useAuthStore()

const containerRef = ref(null)
const canvasRef = ref(null)
const userZoneRef = ref(null)
const centerZoneRef = ref(null)
const friendsZoneRef = ref(null)

const selectedFriend = ref(null)
const selectedFile = ref(null)
const dragOver = ref(false)

const canSend = computed(() => !!selectedFile.value && !!selectedFriend.value)

// --- Lifecycle ---
onMounted(() => {
  friendStore.fetchFriends()
  startAnimation()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  cancelAnimationFrame(animFrameId)
  window.removeEventListener('resize', handleResize)
})

const onlineFriends = computed(() => {
  return friendStore.acceptedFriends.filter(f => f.online)
})

const getInitials = (name) => {
  if (!name) return '?'
  return name.substring(0, 2).toUpperCase()
}

const truncate = (str) => {
    if (str.length > 15) return str.substring(0, 12) + '...'
    return str
}

// --- Interaction Logic ---
const selectFriend = (friend) => {
  selectedFriend.value = friend
}

const deselectFriend = () => {
    selectedFriend.value = null
}

const removeFile = () => {
    selectedFile.value = null
    const input = document.getElementById('p2p-file-input')
    if(input) input.value = ''
}

const handleActionClick = () => {
    if (canSend.value) {
        startTransfer()
    } else {
        document.getElementById('p2p-file-input').click()
    }
}

const handleFileSelect = (event) => {
  if (event.target.files.length > 0) {
    selectedFile.value = event.target.files[0]
  }
}

const handleDrop = (event) => {
  dragOver.value = false
  if (event.dataTransfer.files.length > 0) {
    selectedFile.value = event.dataTransfer.files[0]
  }
}

const startTransfer = async () => {
  if (!canSend.value) return
  
  try {
     await p2pStore.startTransfer(selectedFriend.value, selectedFile.value)
     // Keep file selected for repeated transfers
     // alert("Transfert démarré !")
  } catch (e) {
     console.error("Transfer failed", e)
     alert("Erreur lors du démarrage du transfert")
  }
}

// --- Particle Engine ---
let particles = []
let animFrameId
let lastTime = 0

// Configuration Colors
const COLOR_RED_DARK = 'hsla(0, 40%, 65%, 0.6)'
const COLOR_GREEN_DARK = 'hsla(140, 40%, 55%, 0.8)'

class Particle {
    constructor(x, y, targetX, targetY, color, type) {
        this.x = x
        this.y = y
        this.startX = x
        this.startY = y
        this.targetX = targetX
        this.targetY = targetY
        this.color = color
        
        this.progress = 0
        this.speed = 0.005 + Math.random() * 0.01 // Speed of travel (0 to 1)
        
        // Curve Control Point (offset from midpoint)
        // Perpendicular offset
        const midX = (x + targetX) / 2
        const midY = (y + targetY) / 2
        const dx = targetX - x
        const dy = targetY - y
        const dist = Math.sqrt(dx*dx + dy*dy)
        
        // Random curve direction (up or down relative to line)
        const offset = (Math.random() - 0.5) * (dist * 0.5) 
        
        // Normal vector (-dy, dx)
        this.cpX = midX - dy * (offset / dist)
        this.cpY = midY + dx * (offset / dist)
        
        this.radius = 2 + Math.random() * 4
        this.wobblePhase = Math.random() * Math.PI * 2
    }
    
    update() {
        this.progress += this.speed
        if (this.progress > 1) {
             return false // Dead
        }
        
        // Quadratic Bezier Formula
        // (1-t)^2 * P0 + 2(1-t)t * P1 + t^2 * P2
        const t = this.progress
        const invT = 1 - t
        
        this.x = (invT * invT * this.startX) + (2 * invT * t * this.cpX) + (t * t * this.targetX)
        this.y = (invT * invT * this.startY) + (2 * invT * t * this.cpY) + (t * t * this.targetY)
        
        return true
    }
    
    draw(ctx) {
        ctx.fillStyle = this.color
        ctx.beginPath()
        ctx.arc(this.x, this.y, this.radius, 0, Math.PI * 2)
        ctx.fill()
    }
}

const getRandomPastelDark = () => {
    const hue = Math.floor(Math.random() * 360)
    return `hsla(${hue}, 40%, 60%, 0.6)`
}

const spawnParticle = (sourceRect, targetRect, color, type) => {
    // Random point on source border or center? Let's say center area
    const sx = sourceRect.left + sourceRect.width/2 + (Math.random() - 0.5) * 20
    const sy = sourceRect.top + sourceRect.height/2 + (Math.random() - 0.5) * 20
    
    // Target is center of destination
    const tx = targetRect.left + targetRect.width/2 + (Math.random() - 0.5) * 10
    const ty = targetRect.top + targetRect.height/2 + (Math.random() - 0.5) * 10
    
    return new Particle(sx, sy, tx, ty, color, type)
}

const updateParticles = () => {
    if (!canvasRef.value || !userZoneRef.value || !centerZoneRef.value || !friendsZoneRef.value) return
    
    const ctx = canvasRef.value.getContext('2d')
    const width = canvasRef.value.width
    const height = canvasRef.value.height
    
    ctx.clearRect(0, 0, width, height)
    
    // Rects
    const containerRect = containerRef.value.getBoundingClientRect()
    
    // Helper to get relative coords inside canvas
    const getRelRect = (el) => {
        const rect = el.getBoundingClientRect()
        return {
            left: rect.left - containerRect.left,
            top: rect.top - containerRect.top,
            width: rect.width,
            height: rect.height
        }
    }
    
    const userRect = getRelRect(userZoneRef.value.querySelector('.large-avatar') || userZoneRef.value)
    const centerRect = getRelRect(centerZoneRef.value.querySelector('.action-circle'))
    
    // -- SPAWNING LOGIC --
    
    // 1. LEFT STREAM (User -> Center)
    if (Math.random() < 0.15) { // Spawn rate
        const color = selectedFile.value ? COLOR_GREEN_DARK : COLOR_RED_DARK
        particles.push(spawnParticle(userRect, centerRect, color, 'left'))
    }
    
    // 2. RIGHT STREAM (Friends -> Center)
    if (Math.random() < 0.15) {
        if (selectedFriend.value) {
            // Spawn from Selected Friend Avatar
            const friendAvatarEl = friendsZoneRef.value.querySelector('.friend-avatar')
            if (friendAvatarEl) {
                 const friendRect = getRelRect(friendAvatarEl)
                 particles.push(spawnParticle(friendRect, centerRect, COLOR_GREEN_DARK, 'right'))
            }
        } else {
            // Spawn from List (Randomly along the right box)
            const boxEl = friendsZoneRef.value.querySelector('.friends-list-box')
            if (boxEl) {
                const boxRect = getRelRect(boxEl)
                // Spawn source: Left edge of the box
                const sy = boxRect.top + Math.random() * boxRect.height
                const sx = boxRect.left + 20 
                
                // Target: Center
                const tx = centerRect.left + centerRect.width/2
                const ty = centerRect.top + centerRect.height/2
                
                // Multicolor
                particles.push(new Particle(sx, sy, tx, ty, getRandomPastelDark(), 'right-list'))
            }
        }
    }
    
    // -- UPDATE & DRAW --
    for (let i = particles.length - 1; i >= 0; i--) {
        const p = particles[i]
        const alive = p.update()
        if (alive) {
            p.draw(ctx)
        } else {
            particles.splice(i, 1)
        }
    }
    
    animFrameId = requestAnimationFrame(updateParticles)
}

const startAnimation = () => {
     handleResize() // Init size
     animFrameId = requestAnimationFrame(updateParticles)
}

const handleResize = () => {
    if (canvasRef.value && containerRef.value) {
        // Match canvas size to container
        canvasRef.value.width = containerRef.value.offsetWidth
        canvasRef.value.height = containerRef.value.offsetHeight
    }
}

// Watchers
watch(onlineFriends, () => {
    // Reactive update if friends list changes size/pos
    setTimeout(handleResize, 100)
})

</script>

<style scoped>
.p2p-container {
    padding: 1rem;
    height: 100%;
    position: relative;
    overflow: hidden; /* For particles */
    display: flex;
    flex-direction: column;
}

.particle-canvas {
    position: absolute;
    top: 0;
    left: 0;
    pointer-events: none; /* Let clicks pass through */
    z-index: 1;
}

.p2p-header-minimal {
    text-align: center;
    margin-bottom: 2rem;
    z-index: 2;
}

.p2p-layout {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: space-evenly;
    position: relative;
    z-index: 2;
}

/* --- ZONES --- */
.user-zone, .center-zone, .friends-zone {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-width: 200px;
}

/* User Zone */
.large-avatar {
    width: 120px;
    height: 120px;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 3rem;
    font-weight: bold;
    box-shadow: 0 4px 15px rgba(0,0,0,0.2);
    margin-bottom: 1rem;
    position: relative;
}

.pulse-effect::after {
    content: '';
    position: absolute;
    top: -5px; left: -5px; right: -5px; bottom: -5px;
    border-radius: 50%;
    border: 2px solid rgba(118, 75, 162, 0.4);
    animation: pulse 2s infinite;
}

.friend-avatar.pulse-effect::after {
    border-color: rgba(253, 160, 133, 0.6);
}

@keyframes pulse {
    0% { transform: scale(1); opacity: 1; }
    100% { transform: scale(1.2); opacity: 0; }
}

.user-name {
    font-weight: 600;
    color: var(--main-text-color);
}

/* Center Zone */
.action-circle {
    width: 180px;
    height: 180px;
    border-radius: 50%;
    background: var(--card-color);
    border: 4px dashed var(--border-color);
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: all 0.3s ease;
    box-shadow: 0 4px 10px rgba(0,0,0,0.05);
}

.action-circle:hover {
    border-color: var(--primary-color);
    transform: scale(1.05);
}

.plus-icon {
    font-size: 4.5rem;
    color: var(--secondary-text-color);
}

.action-circle.has-file {
    border-style: solid;
    border-color: #42b983;
}

.action-circle.ready-to-send {
    background: #42b983;
    border-color: #42b983;
    color: white;
}

.file-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    padding: 10px;
    text-align: center;
}

.send-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    color: white;
}

.action-label {
    margin-top: 1rem;
    color: var(--secondary-text-color);
    font-size: 0.9rem;
    min-height: 1.5em;
}

.filename {
    font-size: 0.8rem;
    word-break: break-all;
    max-width: 100px;
}

/* Friends Zone */
.friends-list-box {
    width: 280px;
    display: flex;
    flex-direction: column;
}
zone-label {
    font-weight: 600;
    color: var(--secondary-text-color);
    font-size: 0.8rem;
    text-transform: uppercase;
    letter-spacing: 1px;
    margin-bottom: 12px;
    width: 100%;
    text-align: center;
}

.list-header {
    padding: 0 4px; /* Reduced padding since margin is on zone-label */
    display: flex;
    justify-content: space-between;
    align-items: centercase;
    letter-spacing: 0.5px;
}

.count {
    background: rgba(0,0,0,0.05);
    color: var(--main-text-color);
    padding: 2px 8px;
    border-radius: 12px;
    font-size: 0.75rem;
    font-weight: 700;
}

.scrollable-list {
    max-height: 400px;
    overflow-y: auto;
    padding: 10px; /* More padding for hover effects */
    display: flex;
    flex-direction: column;
    gap: 4px; /* Reduced gap since rows have margins or padding */
}

.friend-row {
    display: flex;
    align-items: center;
    padding: 6px;
    padding-right: 20px;
    background: var(--card-color);
    border-radius: 99px; /* Pill shape */
    cursor: pointer;
    transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
    border: 1px solid transparent;
    box-shadow: 0 2px 8px rgba(0,0,0,0.03);
    margin-bottom: 2px;
}

.friend-row:hover {
    transform: scale(1.03) translateX(5px);
    box-shadow: 0 8px 20px rgba(0,0,0,0.08);
    background: white; 
    z-index: 10;
}

.mini-avatar {
    width: 63px;
    height: 63px;
    border-radius: 50%;
    background: #f1f3f4;
    color: #444;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1.35rem;
    font-weight: 600;
    margin-right: 12px;
    flex-shrink: 0;
    border: 2px solid white; /* Border to separate from row bg */
    box-shadow: 0 2px 5px rgba(0,0,0,0.05);
}

.mini-name {
    flex: 1;
    font-size: 0.95rem;
    color: var(--main-text-color);
    font-weight: 500;
}

.mini-status {
    width: 8px;
    height: 8px;
    background: #4caf50;
    border-radius: 50%;
    box-shadow: 0 0 0 2px var(--card-color);
}

.selected-friend-view {
    display: flex;
    flex-direction: column;
    align-items: center;
    position: relative;
    animation: fadeIn 0.3s;
    /* Removed card styling to match User zone exactly */
}

.close-friend {
    position: absolute;
    top: 0;
    right: 0;
    transform: translate(0, 0); /* Positioned relative to avatar wrapper */
    width: 36px;
    height: 36px;
    border-radius: 50%;
    background: #ff5252;
    color: white;
    border: 2px solid var(--background-color, #fff);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0;
    box-shadow: 0 2px 5px rgba(0,0,0,0.2);
    transition: transform 0.2s;
    z-index: 5;
}

.close-friend:hover {
    transform: scale(1.1); /* Simple scale on hover */
    background: #ff1744;
}

.friend-avatar {
    background: linear-gradient(135deg, #f6d365 0%, #fda085 100%);
    /* Inherits size 80px from .large-avatar */
}

.status-indicator {
    position: absolute;
    bottom: 0;
    right: 0;
    width: 20px;
    height: 20px;
    background: #4caf50;
    border: 3px solid var(--background-color);
    border-radius: 50%;
}

.close-file {
    position: absolute;
    top: 0;
    right: 0;
    width: 42px;
    height: 42px;
    border-radius: 50%;
    background: #e0e0e0;
    color: #555;
    border: 2px solid var(--background-color, #fff);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0;
    box-shadow: 0 2px 5px rgba(0,0,0,0.1);
    transition: all 0.2s;
    z-index: 10;
}

.close-file:hover {
    background: #d6d6d6;
    transform: scale(1.1);
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* Responsive */
@media (max-width: 900px) {
    .p2p-layout {
        flex-direction: column;
        gap: 2rem;
    }
    .user-zone { order: 1; }
    .center-zone { order: 2; margin: 2rem 0; }
    .friends-zone { order: 3; }
    
    .friends-list-box {
        height: 200px;
        width: 100%;
        min-width: 280px;
    }
}

@keyframes fadeIn {
    from { opacity: 0; transform: scale(0.9); }
    to { opacity: 1; transform: scale(1); }
}
</style>