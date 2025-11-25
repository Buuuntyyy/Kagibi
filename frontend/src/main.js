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

authStore.checkAuth().then(() => {
  app.use(router)
  app.mount('#app')
})
