<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<!-- "What's new" modal, driven entirely by the repo-root CHANGELOG.md / CHANGELOG.en.md
     (mirrored into public/ at build time by vite.config.js's syncChangelogPlugin). Only the
     latest (topmost) version section is parsed and shown — there is nothing to hand-edit
     here per release: write the changelog entry as usual, deploy, and the modal reflects it
     automatically, once per version per browser (tracked in localStorage). -->

<template>
  <transition name="modal-fade">
    <div v-if="visible" class="changelog-overlay" @click.self="dismiss">
      <div class="changelog-box">
        <div class="changelog-header">
          <div class="changelog-title-group">
            <span v-if="entry" class="version-badge">v{{ entry.version }}</span>
            <h3>{{ t('changelog.title') }}</h3>
          </div>
          <button class="close-btn" @click="dismiss" :aria-label="t('common.close')">✕</button>
        </div>

        <div class="changelog-content">
          <div v-for="section in entry?.sections ?? []" :key="section.title" class="changelog-section">
            <h4 class="section-title">{{ section.title }}</h4>
            <ul>
              <li v-for="(item, idx) in section.items" :key="idx">
                <template v-if="item.label"><strong>{{ item.label }}</strong><template v-if="item.text"> — {{ item.text }}</template></template>
                <template v-else>{{ item.text }}</template>
              </li>
            </ul>
          </div>
        </div>

        <div class="changelog-footer">
          <button class="confirm-btn" @click="dismiss">{{ t('changelog.confirm') }}</button>
        </div>
      </div>
    </div>
  </transition>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { parseLatestChangelogEntry } from '../utils/changelog'

const { t, locale } = useI18n()

const visible = ref(false)
const entry = ref(null)

function storageKey(lang) {
  return `kagibi_changelog_seen_v_${lang}`
}

async function loadAndMaybeShow() {
  const file = locale.value === 'en' ? '/CHANGELOG.en.md' : '/CHANGELOG.md'
  try {
    const res = await fetch(file, { cache: 'no-cache' })
    if (!res.ok) return
    const text = await res.text()
    const parsed = parseLatestChangelogEntry(text)
    if (!parsed) return

    entry.value = parsed
    const seen = localStorage.getItem(storageKey(locale.value))
    if (seen !== parsed.version) {
      visible.value = true
    }
  } catch (e) {
    // Non-blocking — worst case the "what's new" modal just doesn't appear.
    console.warn('[Changelog] Failed to load changelog:', e)
  }
}

const dismiss = () => {
  visible.value = false
  if (entry.value) {
    localStorage.setItem(storageKey(locale.value), entry.value.version)
  }
}

onMounted(loadAndMaybeShow)

// Re-check when the user switches language while the app is open (e.g. a French user
// who already dismissed v2.31 switches to English before CHANGELOG.en.md catches up —
// per-language storage keys mean this correctly evaluates the English file on its own).
watch(locale, () => {
  visible.value = false
  loadAndMaybeShow()
})
</script>

<style scoped>
.changelog-overlay {
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

.changelog-box {
  background: var(--card-color, #1e1e1e);
  width: 100%;
  max-width: 540px;
  border-radius: 14px;
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.4);
  border: 1px solid var(--border-color, #333);
  overflow: hidden;
  animation: popIn 0.3s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}

.changelog-header {
  padding: 1rem 1.25rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  background: rgba(255, 255, 255, 0.02);
}

.changelog-title-group {
  display: flex;
  align-items: center;
  gap: 0.6rem;
}

.version-badge {
  background: var(--primary-color, #667eea);
  color: #fff;
  font-size: 0.7rem;
  font-weight: 700;
  padding: 0.2rem 0.5rem;
  border-radius: 20px;
  letter-spacing: 0.03em;
}

.changelog-header h3 {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-color, #fff);
}

.close-btn {
  background: none;
  border: none;
  color: var(--secondary-text-color, #aaa);
  font-size: 1rem;
  cursor: pointer;
  padding: 0.25rem 0.4rem;
  border-radius: 6px;
  line-height: 1;
  transition: background 0.15s, color 0.15s;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.08);
  color: var(--text-color, #fff);
}

.changelog-content {
  padding: 1.25rem;
  max-height: 320px;
  overflow-y: auto;
}

.changelog-section {
  margin-bottom: 1rem;
}

.changelog-section:last-child {
  margin-bottom: 0;
}

.section-title {
  margin: 0 0 0.75rem 0;
  font-size: 0.85rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--primary-color, #667eea);
}

.changelog-section ul {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}

.changelog-section li {
  font-size: 0.875rem;
  line-height: 1.55;
  color: var(--secondary-text-color, #ccc);
  padding-left: 0.75rem;
  border-left: 2px solid var(--border-color, #333);
}

.changelog-section li strong {
  color: var(--text-color, #fff);
  font-weight: 600;
}

.changelog-footer {
  padding: 1rem 1.25rem;
  display: flex;
  justify-content: flex-end;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
  background: rgba(0, 0, 0, 0.08);
}

.confirm-btn {
  padding: 0.55rem 1.5rem;
  border: none;
  border-radius: 8px;
  background: var(--primary-color, #667eea);
  color: #fff;
  font-size: 0.875rem;
  font-weight: 600;
  cursor: pointer;
  transition: transform 0.1s, filter 0.1s;
}

.confirm-btn:hover {
  filter: brightness(1.1);
  transform: scale(1.02);
}

.confirm-btn:active {
  transform: scale(0.97);
}

@keyframes popIn {
  from {
    opacity: 0;
    transform: scale(0.92) translateY(12px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: opacity 0.2s ease;
}

.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}

@media (max-width: 600px) {
  .changelog-box {
    margin: 1rem;
    max-width: 100%;
  }
}
</style>
