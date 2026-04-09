import { createI18n } from 'vue-i18n'
import fr from './locales/fr.json'
import en from './locales/en.json'

const i18n = createI18n({
  legacy: false, // Use Composition API mode
  locale: localStorage.getItem('language') || 'fr',
  fallbackLocale: 'fr',
  messages: {
    fr,
    en
  }
})

export default i18n
