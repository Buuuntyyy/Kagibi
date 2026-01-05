<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="cancel">
    <div class="modal-content">
      <div class="modal-header">
        <h3>Supprimer {{ itemsCount > 1 ? 'les éléments' : 'l\'élément' }} ?</h3>
        <button @click="cancel" class="btn-close">×</button>
      </div>
      
      <div class="modal-body">
        <div class="warning-icon">
          ⚠️
        </div>
        <p v-if="itemsCount === 1">
          Êtes-vous sûr de vouloir supprimer <strong>"{{ itemName }}"</strong> ?
        </p>
        <p v-else>
          Êtes-vous sûr de vouloir supprimer ces <strong>{{ itemsCount }} éléments</strong> ?
        </p>
        <p class="sub-text">Cette action est irréversible.</p>
      </div>

      <div class="modal-footer">
        <button @click="cancel" class="btn-secondary">Annuler</button>
        <button @click="confirm" class="btn-delete">Supprimer</button>
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps({
  isOpen: Boolean,
  itemsCount: {
    type: Number,
    default: 1
  },
  itemName: {
    type: String,
    default: ''
  }
});

const emit = defineEmits(['confirm', 'cancel']);

const confirm = () => {
  emit('confirm');
};

const cancel = () => {
  emit('cancel');
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
  animation: fadeIn 0.2s ease;
}

.modal-content {
  background: var(--card-color);
  padding: 0;
  border-radius: 12px;
  width: 400px;
  max-width: 90%;
  box-shadow: 0 10px 25px rgba(0,0,0,0.2);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  border-bottom: 1px solid var(--border-color);
}

.modal-header h3 {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--main-text-color);
}

.btn-close {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: var(--secondary-text-color);
  padding: 0;
  line-height: 1;
}

.modal-body {
  padding: 24px;
  text-align: center;
  color: var(--main-text-color);
}

.warning-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
}

.sub-text {
  color: var(--secondary-text-color);
  font-size: 0.9rem;
  margin-top: 0.5rem;
}

.modal-footer {
  padding: 16px 24px;
  border-top: 1px solid var(--border-color);
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  background-color: var(--background-color);
}

button {
  padding: 8px 16px;
  border-radius: 4px;
  border: 1px solid transparent;
  cursor: pointer;
  font-weight: 500;
  font-size: 0.9rem;
  transition: background-color 0.2s;
}

.btn-secondary {
  background-color: var(--card-color);
  border: 1px solid var(--border-color);
  color: var(--main-text-color);
}

.btn-secondary:hover {
  background-color: var(--hover-background-color);
  border-color: var(--border-color);
}

.btn-delete {
  background-color: var(--error-color);
  color: white;
}

.btn-delete:hover {
  background-color: #c0392b; /* Darker red */
  box-shadow: 0 1px 2px rgba(60,64,67,0.3);
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}
</style>
