<template>
  <div class="home-view-container">
    <div class="welcome-banner">
      <h2>{{ t('home.welcome') }}</h2>
    </div>
    
    <div class="home-sections">
      <RecentlyOpened @open-share-dialog="handleOpenShareDialog" />
      <FileShared />
    </div>
    
    <!-- Share Dialog -->
    <ManageShareDialog
      :isOpen="shareDialog.isOpen"
      :item="shareDialog.item"
      @close="closeShareDialog"
    />
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import RecentlyOpened from '../components/file/RecentlyOpened.vue'
import FileShared from '../components/file/FileShared.vue'
import ManageShareDialog from '../components/ManageShareDialog.vue'
import { useAuthStore } from '../stores/auth'

const { t } = useI18n()

const authStore = useAuthStore()

const shareDialog = ref({
  isOpen: false,
  item: null
})

const handleOpenShareDialog = (item) => {
  shareDialog.value = {
    isOpen: true,
    item: item
  }
}

const closeShareDialog = () => {
  shareDialog.value.isOpen = false
  shareDialog.value.item = null
}

onMounted(async () => {
  // Refresh user info (storage usage, etc.) and ensure RSA keys are loaded
  await authStore.checkAuth();
  // Ensure RSA keys are available for shared folder decryption
  if (authStore.masterKey) {
    await authStore.ensureRSAKeys(authStore.masterKey);
  }
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
