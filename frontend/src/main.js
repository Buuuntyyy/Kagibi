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
import { createUnhead, headSymbol } from '@unhead/vue'

const pinia = createPinia()
const head = createUnhead()

const app = createApp(App)

app.use(pinia)
app.use(i18n)
app.provide(headSymbol, head)
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
  // If owner changed direct-share permissions, refresh active shared session
  if (payload.action === 'share_permissions_updated' && payload.share_id) {
    const fs = useFileStore(pinia);
    fs.refreshSharedPermissionsIfActive(payload.share_id);
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

// Sync presenceState → friend.online, but ONLY for users we have received
// a WS presence update for. Never override with the default false when
// presenceState has no entry — that would clobber the correct value from
// the initial HTTP /friends response before the WS bootstrap arrives.
setInterval(() => {
  const friendStore = useFriendStore(pinia);
  if (friendStore.friends.length > 0) {
    friendStore.friends.forEach(friend => {
      const entry = realtimeStore.presenceState[friend.id];
      if (entry !== undefined && friend.online !== entry.online) {
        friend.online = entry.online;
      }
    });
  }
}, 2000);

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
