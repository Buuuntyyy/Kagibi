<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <Transition name="modal-fade">
    <div v-if="uiStore.confirmDialog.visible" class="modal-overlay" @click.self="uiStore.resolveConfirm(false)">
      <div class="modal-content">
        <div class="modal-header" :class="uiStore.confirmDialog.confirmClass">
          <h3>{{ uiStore.confirmDialog.title }}</h3>
          <button class="btn-close" @click="uiStore.resolveConfirm(false)">×</button>
        </div>
        <div class="modal-body">
          <p>{{ uiStore.confirmDialog.message }}</p>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="uiStore.resolveConfirm(false)">
            {{ uiStore.confirmDialog.cancelLabel }}
          </button>
          <button
            :class="uiStore.confirmDialog.confirmClass === 'danger' ? 'btn-danger' : 'btn-primary'"
            @click="uiStore.resolveConfirm(true)"
          >
            {{ uiStore.confirmDialog.confirmLabel }}
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup>
import { useUIStore } from '../stores/ui'
const uiStore = useUIStore()
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.6);
  backdrop-filter: blur(2px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 99000;
}

.modal-content {
  background: var(--card-color, #1e1e1e);
  border-radius: 12px;
  width: 420px;
  max-width: 90%;
  overflow: hidden;
  border: 1px solid var(--border-color, #333);
  box-shadow: 0 10px 30px rgba(0,0,0,0.4);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 24px;
  border-bottom: 1px solid var(--border-color, #333);
}

.modal-header.danger { background: rgba(244,67,54,0.08); }
.modal-header.primary { background: rgba(0,80,255,0.08); }

.modal-header h3 {
  margin: 0;
  font-size: 1.05rem;
  font-weight: 600;
  color: var(--main-text-color, #eee);
}

.btn-close {
  background: none;
  border: none;
  font-size: 1.4rem;
  cursor: pointer;
  color: var(--secondary-text-color, #aaa);
  line-height: 1;
  padding: 0;
}

.modal-body {
  padding: 20px 24px;
  color: var(--main-text-color, #eee);
  line-height: 1.5;
}

.modal-footer {
  padding: 14px 24px;
  border-top: 1px solid var(--border-color, #333);
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  background: rgba(0,0,0,0.1);
}

button {
  padding: 8px 18px;
  border-radius: 6px;
  border: none;
  cursor: pointer;
  font-weight: 500;
  font-size: 0.9rem;
  transition: all 0.15s;
}

.btn-secondary {
  background: transparent;
  border: 1px solid var(--border-color, #444);
  color: var(--main-text-color, #ccc);
}
.btn-secondary:hover { background: rgba(255,255,255,0.05); }

.btn-danger { background: #f44336; color: white; }
.btn-danger:hover { background: #d32f2f; }

.btn-primary { background: var(--primary-color, #0050FF); color: white; }
.btn-primary:hover { filter: brightness(1.1); }

.modal-fade-enter-active, .modal-fade-leave-active { transition: opacity 0.2s; }
.modal-fade-enter-from, .modal-fade-leave-to { opacity: 0; }
</style>
