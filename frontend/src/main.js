import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import i18n from './i18n.js'
import './style.css'
import './styles/avatar-animations.css'
import { useAuthStore } from './stores/auth'
import { useRealtimeStore } from './stores/realtime'
import { useFriendStore } from './stores/friends'
import { useFileStore } from './stores/files'
import { useP2PStore } from './stores/p2p'
import { initSecurityMonitoring, getSecurityMonitor } from './utils/securityMonitoring'
import { detectXSSAttempts, setupXSSMonitoring } from './utils/secureCrypto'
import { createHead } from '@unhead/vue'

const pinia = createPinia()
const head = createHead()

const app = createApp(App)

app.use(pinia)
app.use(i18n)
app.use(head)
const authStore = useAuthStore(pinia);
const realtimeStore = useRealtimeStore(pinia);

// Register Supabase Realtime event handlers
realtimeStore.onEvent('storage_update', (payload) => {
  //console.log('[Main] Storage update received:', payload);
  if (authStore.user && payload.storage_used !== undefined) {
    authStore.updateUserStorage(payload.storage_used, payload.storage_limit);
  }
  // If action indicates share update, refetch files
  const shareActions = ['share_created', 'share_revoked', 'share_received', 'share_revoked_by_recipient', 'share_removed_from_imported'];
  if (shareActions.includes(payload.action)) {
    const fs = useFileStore(pinia);
    fs.fetchItems(fs.currentPath);
    fs.notifyShareUpdate();
  }
});

realtimeStore.onEvent('friend_update', (payload) => {
  //console.log('[Main] Friend update received:', payload);
  const friendStore = useFriendStore(pinia);
  friendStore.fetchFriends();
});

realtimeStore.onEvent('p2p_signal', (payload) => {
  //console.log('[Main] P2P signal received:', payload);
  const p2pStore = useP2PStore(pinia);
  // Transform to expected format
  p2pStore.handleSignal({
    sender_id: payload.from,
    type: payload.type,
    data: payload.payload
  });
});

// Watch presence changes and update friend online status
setInterval(() => {
  const friendStore = useFriendStore(pinia);
  if (friendStore.friends.length > 0) {
    friendStore.friends.forEach(friend => {
      const wasOnline = friend.online;
      const isOnline = realtimeStore.isUserOnline(friend.id);
      if (wasOnline !== isOnline) {
        friend.online = isOnline;
      }
    });
  }
}, 2000); // Check every 2 seconds

// Initialiser le monitoring de sécurité
initSecurityMonitoring();

// Vérifier les injections XSS
window.addEventListener('load', () => {
  const isSecure = detectXSSAttempts();
  if (!isSecure) {
    getSecurityMonitor().logSecurityEvent(
      'XSS_DETECTED_AT_STARTUP',
      'high',
      { timestamp: new Date().toISOString() }
    );
  }
  
  // Configurer le monitoring continu des injections
  setupXSSMonitoring();
});

app.use(router)
app.mount('#app')
