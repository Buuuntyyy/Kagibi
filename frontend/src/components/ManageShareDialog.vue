<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="close">
    <div class="modal-content">
      <div class="modal-header">
        <h3>Gérer le partage</h3>
        <button @click="close" class="btn-close">×</button>
      </div>
      <div class="modal-body">
        <p>Ce fichier est partagé. Vous pouvez copier le lien ou arrêter le partage.</p>
        <div class="link-container">
          <input type="text" :value="shareUrl" readonly ref="shareLinkInput" />
          <button @click="copyLink">Copier</button>
        </div>
      </div>
      <div class="modal-footer">
        <button @click="deleteShare" class="btn-delete">Supprimer le partage</button>
        <button @click="close">Fermer</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue';
import api from '../api';

const props = defineProps({
  isOpen: Boolean,
  item: Object,
});

const emit = defineEmits(['close', 'share-deleted']);

const shareLinkInput = ref(null);

const shareUrl = computed(() => {
  if (props.item && props.item.share_token) {
    return `${window.location.origin}/s/${props.item.share_token}`;
  }
  return '';
});

const copyLink = () => {
  if (shareLinkInput.value) {
    shareLinkInput.value.select();
    navigator.clipboard.writeText(shareUrl.value).then(() => {
      alert('Lien copié dans le presse-papiers !');
    }).catch(err => {
      console.error('Impossible de copier le lien:', err);
      alert('Erreur lors de la copie du lien.');
    });
  }
};

const deleteShare = async () => {
  if (!props.item || !props.item.share_id) return;
  if (confirm('Êtes-vous sûr de vouloir supprimer ce partage ? Le lien ne sera plus accessible.')) {
    try {
      await api.delete(`/shares/link/${props.item.share_id}`);
      alert('Le partage a été supprimé.');
      emit('share-deleted');
      close();
    } catch (error) {
      console.error('Erreur lors de la suppression du partage:', error);
      alert('Impossible de supprimer le partage.');
    }
  }
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
  width: 500px;
  max-width: 90%;
  box-shadow: 0 4px 15px rgba(0,0,0,0.2);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #eee;
  padding-bottom: 10px;
  margin-bottom: 20px;
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

.modal-body p {
  margin-bottom: 15px;
}

.link-container {
  display: flex;
  gap: 10px;
}

.link-container input {
  flex-grow: 1;
  padding: 8px;
  border: 1px solid #ccc;
  border-radius: 4px;
  background-color: #f9f9f9;
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
  border: 1px solid transparent;
  cursor: pointer;
}

.btn-delete {
  background-color: #dc3545;
  color: white;
  border-color: #dc3545;
}
</style>
