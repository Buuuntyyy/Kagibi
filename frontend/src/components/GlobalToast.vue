<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <TransitionGroup name="toast" tag="div" class="toast-container">
    <div
      v-for="toast in uiStore.toasts"
      :key="toast.id"
      class="toast-item"
      :class="toast.type"
      @click="uiStore.dismissToast(toast.id)"
    >
      {{ toast.message }}
    </div>
  </TransitionGroup>
</template>

<script setup>
import { useUIStore } from '../stores/ui'
const uiStore = useUIStore()
</script>

<style scoped>
.toast-container {
  position: fixed;
  bottom: 80px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  z-index: 99999;
  pointer-events: none;
}

.toast-item {
  padding: 10px 20px;
  border-radius: 8px;
  color: white;
  font-size: 0.9rem;
  font-weight: 500;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  cursor: pointer;
  pointer-events: auto;
  white-space: nowrap;
  max-width: 400px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.toast-item.success { background: #16a34a; }
.toast-item.error   { background: #dc2626; }
.toast-item.info    { background: #2563eb; }
.toast-item.warning { background: #d97706; }

.toast-enter-active, .toast-leave-active {
  transition: opacity 0.25s, transform 0.25s;
}
.toast-enter-from, .toast-leave-to {
  opacity: 0;
  transform: translateY(12px);
}
</style>
