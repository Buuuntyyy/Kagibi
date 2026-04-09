<template>
  <transition name="modal-fade">
    <div v-if="uiStore.alert.visible" class="warn-overlay" @click.self="uiStore.closeAlert">
      <div class="warn-box" :class="uiStore.alert.type">
        <div class="warn-header">
            <span class="warn-icon" v-if="uiStore.alert.type === 'error'">
                <svg viewBox="0 0 24 24" width="24" height="24" stroke="currentColor" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
            </span>
            <span class="warn-icon" v-else-if="uiStore.alert.type === 'warning'">
                <svg viewBox="0 0 24 24" width="24" height="24" stroke="currentColor" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path><line x1="12" y1="9" x2="12" y2="13"></line><line x1="12" y1="17" x2="12.01" y2="17"></line></svg>
            </span>
            <span class="warn-icon" v-else>
               <svg viewBox="0 0 24 24" width="24" height="24" stroke="currentColor" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="16" x2="12" y2="12"></line><line x1="12" y1="8" x2="12.01" y2="8"></line></svg>
            </span>
            
            <h3>{{ uiStore.alert.title }}</h3>
        </div>
        
        <div class="warn-content">
            <p>{{ uiStore.alert.message }}</p>
        </div>
        
        <div class="warn-footer">
            <button class="warn-btn" @click="uiStore.closeAlert">{{ t('common.ok') }}</button>
        </div>
      </div>
    </div>
  </transition>
</template>

<script setup>
import { useI18n } from 'vue-i18n'
import { useUIStore } from '../stores/ui'

const { t } = useI18n()
const uiStore = useUIStore()
</script>

<style scoped>
.warn-overlay {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    background: rgba(0, 0, 0, 0.6);
    backdrop-filter: blur(4px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 9999;
}

.warn-box {
    background: var(--card-color, #1e1e1e);
    min-width: 320px;
    max-width: 450px;
    border-radius: 12px;
    box-shadow: 0 10px 30px rgba(0,0,0,0.3);
    border: 1px solid var(--border-color, #333);
    overflow: hidden;
    animation: popIn 0.3s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}

.warn-header {
    padding: 1rem;
    display: flex;
    align-items: center;
    gap: 10px;
    border-bottom: 1px solid rgba(255,255,255,0.05);
}

.warn-header h3 {
    margin: 0;
    font-size: 1.1rem;
    font-weight: 600;
}

.warn-content {
    padding: 1.5rem 1rem;
    color: var(--secondary-text-color, #ccc);
    line-height: 1.5;
    text-align: center;
}

.warn-footer {
    padding: 1rem;
    display: flex;
    justify-content: center;
    background: rgba(0,0,0,0.1);
}

.warn-btn {
    padding: 0.6rem 2rem;
    border: none;
    border-radius: 6px;
    background: var(--primary-color, #667eea);
    color: white;
    font-weight: 500;
    cursor: pointer;
    transition: transform 0.1s;
}

.warn-btn:hover {
    transform: scale(1.03);
    filter: brightness(1.1);
}

.warn-btn:active {
    transform: scale(0.97);
}

/* Modifier styles */
.warn-box.error .warn-header {
    background: rgba(255, 82, 82, 0.1);
    color: #ff5252;
}

.warn-box.warning .warn-header {
    background: rgba(255, 179, 0, 0.1);
    color: #ffb300;
}

.warn-box.info .warn-header {
    background: rgba(33, 150, 243, 0.1);
    color: #2196f3;
}

/* Animations */
@keyframes popIn {
    from { opacity: 0; transform: scale(0.9) translateY(10px); }
    to { opacity: 1; transform: scale(1) translateY(0); }
}

.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: opacity 0.2s ease;
}

.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}
</style>
