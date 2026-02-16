<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="close">
    <div class="modal-content">
      <div class="modal-header">
        <h3>{{ t('dialogs.move.title') }}</h3>
        <button @click="close" class="btn-close">×</button>
      </div>
      <div class="modal-body">
        <p>{{ t('dialogs.move.selectDestination') }}</p>
        <div class="path-display">
          <span @click="goUp" class="back-arrow" :class="{ 'disabled': currentPath === '/' }">↑</span>
          <span>{{ currentPath }}</span>
        </div>
        <div class="folder-list">
          <div v-if="loading" class="loading-spinner">Chargement...</div>
          <div v-else-if="folders.length === 0 && currentPath !== '/'">Aucun sous-dossier.</div>
          <div v-for="folder in folders" :key="folder.ID" class="folder-item" @click="navigateTo(folder.Name)">
            📁 {{ folder.Name }}
          </div>
        </div>
      </div>
      <div class="modal-footer">
        <button @click="close">{{ t('dialogs.move.cancel') }}</button>
        <button @click="confirmMove" class="btn-primary">{{ t('dialogs.move.moveHere') }}</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import api from '../api';

const { t } = useI18n();

const props = defineProps({
  isOpen: Boolean,
});

const emit = defineEmits(['close', 'move-to']);

const loading = ref(false);
const currentPath = ref('/');
const folders = ref([]);

const fetchFolders = async (path) => {
  loading.value = true;
  try {
    const response = await api.get(`/files/list${path}`);
    folders.value = response.data.folders || [];
  } catch (error) {
    console.error('Error fetching folders:', error);
    folders.value = [];
  } finally {
    loading.value = false;
  }
};

watch(() => props.isOpen, (newVal) => {
  if (newVal) {
    currentPath.value = '/';
    fetchFolders('/');
  }
});

const navigateTo = (folderName) => {
  let newPath = currentPath.value;
  if (newPath.endsWith('/')) {
    newPath += folderName;
  } else {
    newPath += `/${folderName}`;
  }
  currentPath.value = newPath;
  fetchFolders(newPath);
};

const goUp = () => {
  if (currentPath.value === '/') return;
  const parts = currentPath.value.split('/').filter(p => p);
  parts.pop();
  const newPath = '/' + parts.join('/');
  currentPath.value = newPath;
  fetchFolders(newPath);
};

const confirmMove = () => {
  emit('move-to', currentPath.value);
  close();
};

const close = () => {
  emit('close');
};
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.6);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  padding: 20px;
  border-radius: 8px;
  width: 400px;
  max-width: 90%;
  box-shadow: 0 4px 15px rgba(0,0,0,0.2);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #eee;
  padding-bottom: 10px;
  margin-bottom: 15px;
}

.modal-header h3 {
  margin: 0;
  font-size: 1.2rem;
}

.btn-close {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
}

.path-display {
  background-color: #f5f5f5;
  padding: 8px;
  border-radius: 4px;
  margin-bottom: 10px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.back-arrow {
  cursor: pointer;
  font-weight: bold;
}

.back-arrow.disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.folder-list {
  height: 200px;
  overflow-y: auto;
  border: 1px solid #ddd;
  padding: 5px;
  border-radius: 4px;
}

.folder-item {
  padding: 8px;
  cursor: pointer;
  border-radius: 4px;
}

.folder-item:hover {
  background-color: #f0f0f0;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 20px;
  padding-top: 10px;
  border-top: 1px solid #eee;
}

button {
  padding: 8px 15px;
  border-radius: 4px;
  border: 1px solid #ccc;
  cursor: pointer;
}

.btn-primary {
  background-color: #42b983;
  color: white;
  border-color: #42b983;
}
</style>
