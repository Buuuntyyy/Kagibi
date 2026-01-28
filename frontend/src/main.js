import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import './style.css'
import { useAuthStore } from './stores/auth'
import { initSecurityMonitoring, getSecurityMonitor } from './utils/securityMonitoring'
import { detectXSSAttempts, setupXSSMonitoring } from './utils/secureCrypto'

const pinia = createPinia()

const app = createApp(App)

app.use(pinia)
const authStore = useAuthStore(pinia);

// Initialiser le monitoring de sécurité
initSecurityMonitoring();

// Vérifier les injections XSS
window.addEventListener('load', () => {
  const isSecure = detectXSSAttempts();
  if (!isSecure) {
    console.error('[App] XSS attempts detected at startup');
    getSecurityMonitor().logSecurityEvent(
      'XSS_DETECTED_AT_STARTUP',
      'critical',
      { timestamp: new Date().toISOString() }
    );
  }
  
  // Configurer le monitoring continu des injections
  setupXSSMonitoring();
});

// Initialiser le Service Worker crypto au démarrage
authStore.initCrypto().then(() => {
  console.log('[App] Secure crypto initialized');
}).catch(err => {
  console.error('[App] Failed to init secure crypto:', err);
  getSecurityMonitor().logSecurityEvent(
    'CRYPTO_INIT_FAILED',
    'high',
    { error: err.message }
  );
});

app.use(router)
app.mount('#app')
