<template>
  <div class="public-browse-container">
    <div v-if="store.isLoading" class="loading-spinner">Chargement...</div>
    <div v-else-if="store.error" class="error-message">{{ store.error }}</div>
    
    <div v-else class="browse-content">
      <header class="browse-header">
        <h1>Contenu partagé : {{ store.resourceName }}</h1>
        <p v-if="store.ownerEmail">Partagé par : {{ store.ownerEmail }}</p>
      </header>
      <PublicFileList />
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue';
import { useRoute } from 'vue-router';
import { usePublicFileStore } from '../stores/publicFileStore';
import PublicFileList from '../components/PublicFileList.vue';

const route = useRoute();
const store = usePublicFileStore();

onMounted(() => {
  const token = route.params.token;
  let subpath = '/';
  if (route.params.subpath) {
    if (Array.isArray(route.params.subpath)) {
      subpath = `/${route.params.subpath.join('/')}`;
    } else {
      subpath = `/${route.params.subpath}`;
    }
  }
  store.fetchItems(token, subpath);
});
</script>

<style scoped>
.public-browse-container {
  padding: 2rem;
  width: 60%;
  margin: 0 auto;
  box-sizing: border-box;
}

@media (max-width: 1200px) {
  .public-browse-container {
    width: 90%;
  }
}

.loading-spinner, .error-message {
  text-align: center;
  font-size: 1.2rem;
  padding: 2rem;
}

.error-message {
  color: #dc3545;
}

.browse-header {
  margin-bottom: 1.5rem;
  border-bottom: 1px solid #eee;
  padding-bottom: 1rem;
}

.browse-header h1 {
  margin: 0;
  font-size: 1.8rem;
}
</style>
