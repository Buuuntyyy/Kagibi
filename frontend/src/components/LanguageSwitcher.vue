<template>
  <div class="language-switcher">
    <select
      v-model="currentLanguage"
      class="lang-select"
      @change="switchLanguage(currentLanguage)"
      aria-label="Language selector"
    >
      <option v-for="lang in languages" :key="lang.code" :value="lang.code">
        {{ lang.flag }} {{ lang.name }}
      </option>
    </select>
  </div>
</template>

<script setup>
import { useI18n } from 'vue-i18n'
import { onMounted, ref } from 'vue'

const { locale } = useI18n()
const currentLanguage = ref(localStorage.getItem('language') || 'fr')

const languages = [
  { code: 'fr', name: 'Français', flag: '🇫🇷' },
  { code: 'en', name: 'English', flag: '🇬🇧' }
]

const switchLanguage = (lang) => {
  locale.value = lang
  currentLanguage.value = lang
  localStorage.setItem('language', lang)
}

onMounted(() => {
  const saved = localStorage.getItem('language')
  if (saved) {
    locale.value = saved
    currentLanguage.value = saved
  }
})
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
  font-size: 0.9rem;
  cursor: pointer;
  padding: 0.35rem 0.5rem;
  outline: none;
}

.lang-select:hover {
  background-color: var(--card-color);
}

.lang-select:focus {
  border-color: var(--primary-color);
}
</style>
