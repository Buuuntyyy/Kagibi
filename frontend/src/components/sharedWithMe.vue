<template>
  <div class="shared-with-me-container">
    <div v-if="loading" class="loading">
      <div class="spinner"></div> Chargement...
    </div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="items.length === 0" class="empty">
      <p>Aucun fichier partagé avec vous.</p>
    </div>
    <FileTable 
      v-else 
      :folders="sharedFolders"
      :files="sharedFiles"
      :columns="columns"
    >
      <template #shared_name="{ item }">
        <span :title="item.name">{{ item.name }}</span>
      </template>

      <template #owner="{ item }">
        {{ item.owner_name }}
      </template>

      <template #shared_at="{ item }">
        {{ formatDate(item.shared_at) }}
      </template>

      <template #size="{ item }">
        {{ formatSize(item.size) }}
      </template>

      <template #actions="{ item }">
        <!-- Actions futures (ex: télécharger, supprimer de ma liste) -->
      </template>
    </FileTable>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import FileTable from './file/FileTable.vue';
import { formatSize, formatDate } from '../utils/format';

// Mock data for now as backend implementation is pending
const items = ref([]);
const loading = ref(false);
const error = ref(null);

const columns = [
  { key: 'icon', label: '', headerClass: 'icon-col', cellClass: 'icon-col' },
  { key: 'shared_name', label: 'Nom', cellClass: 'name-cell' },
  { key: 'owner', label: 'Propriétaire' },
  { key: 'shared_at', label: 'Partagé le' },
  { key: 'size', label: 'Taille' },
  { key: 'actions', label: 'Actions' }
]

const sharedFolders = computed(() => items.value.filter(i => i.type === 'folder'))
const sharedFiles = computed(() => items.value.filter(i => i.type === 'file'))

const fetchSharedWithMe = async () => {
  loading.value = true;
  try {
    const response = await api.get('/shares/with-me');
    items.value = response.data;
  } catch (err) {
    console.error("Error fetching shared with me:", err);
    error.value = "Impossible de charger les fichiers partagés avec vous.";
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  fetchSharedWithMe();
});
</script>

<style scoped>
.shared-with-me-container {
  height: 100%;
  width: 100%;
}

.loading, .error, .empty {
  text-align: center;
  padding: 20px;
  color: #888;
}

.error {
  color: #ef5350;
}

.spinner {
  border: 3px solid rgba(255, 255, 255, 0.1);
  border-radius: 50%;
  border-top: 3px solid #64b5f6;
  width: 24px;
  height: 24px;
  animation: spin 1s linear infinite;
  display: inline-block;
  vertical-align: middle;
  margin-right: 10px;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}
</style>