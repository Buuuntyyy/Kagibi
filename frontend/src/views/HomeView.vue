<template>
  <div class="home-view-container">
    <div class="welcome-banner">
      <h2>Bienvenue sur votre Cloud Sécurisé</h2>
    </div>
    
    <div class="home-sections">
      <RecentlyOpened />
      <FileShared />
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import RecentlyOpened from '../components/file/RecentlyOpened.vue'
import FileShared from '../components/file/FileShared.vue'
import { useAuthStore } from '../stores/auth'

const authStore = useAuthStore()

onMounted(async () => {
  // Refresh user info (storage usage, etc.)
  await authStore.checkAuth();
})
</script>

<style scoped>
.home-view-container {
  padding: 1rem;
  height: 100%;
  box-sizing: border-box;
  overflow-y: auto;
}

.welcome-banner {
  margin-bottom: 2rem;
  padding: 1rem;
}

.welcome-banner h2 {
  margin: 0;
  color: var(--text-color);
  font-weight: 500;
}

.home-sections {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
</style>
