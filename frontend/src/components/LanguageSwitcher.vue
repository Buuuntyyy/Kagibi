<template>
  <div class="language-switcher">
    <button 
      v-for="lang in languages" 
      :key="lang.code"
      @click="switchLanguage(lang.code)"
      :class="['lang-btn', { active: currentLanguage === lang.code }]"
      :title="`${lang.name}`"
    >
      {{ lang.flag }}
    </button>
  </div>
</template>

<script setup>
import { useI18n } from 'vue-i18n'
import { onMounted, ref } from 'vue'

const { locale } = useI18n()
const currentLanguage = ref(localStorage.getItem('language') || 'fr')

const languages = [
  { code: 'fr', name: 'Français', flag: '🇫🇷' },
  { code: 'en', name: 'English', flag: '🇺🇸' }
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
  gap: 0.5rem;
  align-items: center;
}

.lang-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  transition: background-color 0.2s;
  opacity: 0.6;
}

.lang-btn:hover {
  background-color: rgba(0, 0, 0, 0.05);
  opacity: 1;
}

.lang-btn.active {
  opacity: 1;
  background-color: rgba(0, 0, 0, 0.1);
}
</style>
