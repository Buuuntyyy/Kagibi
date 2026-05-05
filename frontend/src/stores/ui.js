import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useUIStore = defineStore('ui', () => {
    const pendingMobileAction = ref(null) // 'upload' | 'createFolder' | null

    const alert = ref({
        visible: false,
        title: '',
        message: '',
        type: 'error' // error, warning, info
    })

    // --- Delete Confirmation State ---
    const deleteDialog = ref({
        visible: false,
        title: '',
        message: '',
        itemName: '',
        itemsCount: 1,
        onConfirm: null
    })

    const showError = (message, title = 'Erreur') => {
        alert.value = {
            visible: true,
            title,
            message,
            type: 'error'
        }
    }

    const showWarning = (message, title = 'Attention') => {
        alert.value = {
            visible: true,
            title,
            message,
            type: 'warning'
        }
    }

    const showInfo = (message, title = 'Information') => {
        alert.value = {
            visible: true,
            title,
            message,
            type: 'info'
        }
    }

    const requestDeleteConfirmation = ({ title, message, itemName, itemsCount = 1, onConfirm }) => {
        deleteDialog.value = {
            visible: true,
            title,
            message,
            itemName,
            itemsCount,
            onConfirm
        }
    }

    const closeAlert = () => {
        alert.value.visible = false
    }

    const closeDeleteDialog = () => {
        deleteDialog.value.visible = false
        deleteDialog.value.onConfirm = null
    }

    return {
        alert,
        pendingMobileAction,
        deleteDialog, // Export state
        showError,
        showWarning,
        showInfo,
        closeAlert,
        requestDeleteConfirmation, // Export action
        closeDeleteDialog // Export action
    }
})
