<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="close">
    <div class="modal-content">
      <div class="modal-header">
        <h3>Aide & Support</h3>
        <button class="close-btn" @click="close">&times;</button>
      </div>
      <div class="modal-body">
        <p class="description">Vous avez trouvé un bug ou vous avez une idée géniale ? Sélectionnez une option ci-dessous pour nous en faire part sur GitHub.</p>
        
        <div class="action-buttons">
          <a :href="bugReportUrl" target="_blank" rel="noopener noreferrer" class="action-btn bug-btn" @click="close">
            <Bug class="btn-icon" :size="32" :stroke-width="2" />
            <div class="btn-text">
                <span class="btn-title">Signaler un bug</span>
                <span class="btn-desc">Quelque chose ne fonctionne pas ? Dites-le nous !</span>
            </div>
          </a>

          <a :href="featureRequestUrl" target="_blank" rel="noopener noreferrer" class="action-btn feature-btn" @click="close">
            <Lightbulb class="btn-icon" :size="32" :stroke-width="2" />
            <div class="btn-text">
                <span class="btn-title">Proposer une fonctionnalité</span>
                <span class="btn-desc">Vous avez une idée d'amélioration ?</span>
            </div>
          </a>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { Bug, Lightbulb } from 'lucide-vue-next'

const props = defineProps({
  isOpen: Boolean
})

const emit = defineEmits(['update:isOpen'])

const close = () => {
  emit('update:isOpen', false)
}

// GitHub issue template URLs
const repoUrl = "https://github.com/buuuntyyy/SaferCloud"
const bugReportUrl = repoUrl + "/issues/new?assignees=&labels=bug&template=bug_report.md&title=%5BBUG%5D+"
const featureRequestUrl = repoUrl + "/issues/new?assignees=&labels=enhancement&template=feature_request.md&title=%5BFEATURE%5D+"
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
  z-index: 3000;
  backdrop-filter: blur(2px);
}

.modal-content {
  background: var(--card-color, #ffffff);
  color: var(--main-text-color, #333333);
  border-radius: 12px;
  width: 90%;
  max-width: 480px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
  overflow: hidden;
  border: 1px solid var(--border-color, #eeeeee);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color, #eeeeee);
  background-color: rgba(0,0,0,0.02);
}

h3 {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 600;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  color: var(--secondary-text-color, #666);
  cursor: pointer;
  line-height: 1;
  padding: 0 4px;
  border-radius: 4px;
  transition: color 0.2s, background-color 0.2s;
}

.close-btn:hover {
  color: #ff4d4d;
  background-color: rgba(255, 77, 77, 0.1);
}

.modal-body {
  padding: 20px;
}

.description {
  margin: 0 0 20px 0;
  font-size: 0.95rem;
  color: var(--secondary-text-color, #666);
  line-height: 1.5;
}

.action-buttons {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.action-btn {
  display: flex;
  align-items: center;
  padding: 16px;
  border-radius: 8px;
  text-decoration: none;
  color: var(--main-text-color, #333);
  border: 1px solid var(--border-color, #ddd);
  transition: transform 0.2s, box-shadow 0.2s, border-color 0.2s;
  background: var(--background-color, #fafafa);
}

.action-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
}

.bug-btn:hover {
  border-color: #e74c3c;
}

.bug-btn:hover .btn-icon {
  color: #e74c3c;
}

.feature-btn:hover {
  border-color: #3498db;
}

.feature-btn:hover .btn-icon {
  color: #3498db;
}

.btn-icon {
  width: 32px;
  height: 32px;
  margin-right: 16px;
  color: var(--secondary-text-color, #888);
  transition: color 0.2s;
}

.btn-text {
  display: flex;
  flex-direction: column;
}

.btn-title {
  font-weight: 600;
  font-size: 1.05rem;
  margin-bottom: 4px;
}

.btn-desc {
  font-size: 0.85rem;
  color: var(--secondary-text-color, #666);
}
</style>
