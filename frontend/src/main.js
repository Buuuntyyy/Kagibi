import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import './style.css'
import { useAuthStore } from './stores/auth'

const pinia = createPinia()

const app = createApp(App)

app.use(pinia)
const authStore = useAuthStore(pinia);

// Initialiser le Service Worker crypto au démarrage
authStore.initCrypto().then(() => {
  console.log('[App] Secure crypto initialized');
}).catch(err => {
  console.error('[App] Failed to init secure crypto:', err);
});

app.use(router)
app.mount('#app')
