<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="language-switcher">
    <select
      :value="currentLanguage"
      class="lang-select"
      @change="switchLanguage($event.target.value)"
      aria-label="Language selector"
    >
      <option v-for="lang in languages" :key="lang.code" :value="lang.code">
        {{ lang.flag }}
      </option>
    </select>
  </div>
</template>

<script setup>
import { ref, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'

const { locale } = useI18n({ useScope: 'global' })

const currentLanguage = ref(locale.value)

const switchLanguage = async (lang) => {
  currentLanguage.value = lang
  await nextTick()
  locale.value = lang
  localStorage.setItem('language', lang)
}

const languages = [
  { code: 'fr', flag: '🇫🇷' },
  { code: 'en', flag: '🇬🇧' }
]
</script>

<style scoped>
.language-switcher {
  display: flex;
  align-items: center;
}

.lang-select {
  background-color: var(--background-color);
  color: var(--main-text-color);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  font-size: 1.25rem;
  cursor: pointer;
  padding: 0.1rem 0.3rem;
  outline: none;
}

.lang-select:hover {
  background-color: var(--card-color);
}

.lang-select:focus {
  border-color: var(--primary-color);
}
</style>
