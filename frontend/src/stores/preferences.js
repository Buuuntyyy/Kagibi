import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

export const usePreferencesStore = defineStore('preferences', () => {
  // Valeurs par défaut
  const enableContextMenu = ref(true)
  const showToolBar = ref(true)

  // Initialisation depuis le LocalStorage
  const init = () => {
    const stored = localStorage.getItem('user_preferences')
    if (stored) {
      try {
        const parsed = JSON.parse(stored)
        // On vérifie que les clés existent pour éviter d'écraser avec undefined
        if (parsed.enableContextMenu !== undefined) enableContextMenu.value = parsed.enableContextMenu
        if (parsed.showToolBar !== undefined) showToolBar.value = parsed.showToolBar
      } catch (e) {
        console.error("Erreur lors du chargement des préférences", e)
      }
    }
  }

  // Sauvegarde automatique à chaque changement
  watch([enableContextMenu, showToolBar], () => {
    localStorage.setItem('user_preferences', JSON.stringify({
      enableContextMenu: enableContextMenu.value,
      showToolBar: showToolBar.value
    }))
  })

  // Lancer l'init à la création du store
  init()

  return {
    enableContextMenu,
    showToolBar
  }
})