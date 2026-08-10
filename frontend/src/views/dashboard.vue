<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="dashboard-container">
    <LeftBar />
    <div class="main-content">
      <!-- Rappel kit de récupération : visible tant que la preuve de possession
           du code n'a pas été faite (snooze 7 jours au clic sur ✕) -->
      <div v-if="showRecoveryReminder" class="recovery-reminder">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zM9 6c0-1.66 1.34-3 3-3s3 1.34 3 3v2H9V6z"/></svg>
        <span>{{ t('recoveryKit.reminderText') }}</span>
        <router-link to="/account#recovery-kit" class="reminder-link">{{ t('recoveryKit.reminderCta') }}</router-link>
        <button class="reminder-close" @click="snoozeRecoveryReminder" :title="t('recoveryKit.reminderDismiss')">✕</button>
      </div>
      <router-view />
    </div>

    <FilePreview
      :visible="fileStore.preview.show"
      :fileUrl="fileStore.preview.url"
      :fileName="fileStore.preview.name"
      :mimeType="fileStore.preview.type"
      :loading="fileStore.preview.loading"
      :status="fileStore.preview.status"
      @close="fileStore.preview.show = false"
    />

    <!-- Mobile Bottom Navigation -->
    <MobileBottomNav />

    <!-- Changelog popup -->
    <ChangelogModal />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import LeftBar from '../components/bar/leftBar.vue'
import MobileBottomNav from '../components/bar/MobileBottomNav.vue'
import FilePreview from '../components/file/FilePreview.vue'
import ChangelogModal from '../components/ChangelogModal.vue'
import { useFileStore } from '../stores/files'
import { useAuthStore } from '../stores/auth'

const { t } = useI18n()
const fileStore = useFileStore()
const authStore = useAuthStore()

const RECOVERY_SNOOZE_KEY = 'kagibi_recovery_reminder_snooze'
const SNOOZE_MS = 7 * 24 * 60 * 60 * 1000

const snoozedUntil = ref(Number(localStorage.getItem(RECOVERY_SNOOZE_KEY) || 0))

const showRecoveryReminder = computed(() =>
  authStore.user &&
  !authStore.user.recovery_verified_at &&
  Date.now() > snoozedUntil.value
)

const snoozeRecoveryReminder = () => {
  const until = Date.now() + SNOOZE_MS
  snoozedUntil.value = until
  try { localStorage.setItem(RECOVERY_SNOOZE_KEY, String(until)) } catch (_) {}
}
</script>

<style scoped>
.dashboard-container {
  display: flex;
  height: 100%;
  width: 100%;
  box-sizing: border-box;
  background-color: var(--background-color);
}

.main-content {
  flex-grow: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  border-top-left-radius: 30px;
  background-color: var(--card-color);
}

.recovery-reminder {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 16px;
  background: rgba(240, 160, 40, 0.12);
  border-bottom: 1px solid rgba(240, 160, 40, 0.3);
  color: var(--main-text-color);
  font-size: 0.82rem;
  flex-shrink: 0;
}

.recovery-reminder svg {
  flex-shrink: 0;
  color: #c07b12;
}

.recovery-reminder span {
  flex: 1;
  min-width: 0;
}

.reminder-link {
  color: var(--primary-color, #4f6ef7);
  font-weight: 600;
  text-decoration: none;
  white-space: nowrap;
}

.reminder-close {
  border: none;
  background: none;
  color: var(--subtle-text-color, #8a8a8a);
  cursor: pointer;
  font-size: 0.9rem;
  padding: 2px 6px;
}

@media (max-width: 768px) {
  .dashboard-container {
    flex-direction: column;
    padding-bottom: 64px; /* space for bottom nav */
  }

  .main-content {
    border-top-left-radius: 0;
    border-radius: 0;
    overflow-y: auto;
  }
}
</style>
