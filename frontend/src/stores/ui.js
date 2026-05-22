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

    // --- Toast State ---
    const toasts = ref([])
    let toastIdCounter = 0

    const showToast = (message, type = 'success') => {
        const id = ++toastIdCounter
        toasts.value.push({ id, message, type })
        setTimeout(() => {
            toasts.value = toasts.value.filter(t => t.id !== id)
        }, 3000)
    }

    const dismissToast = (id) => {
        toasts.value = toasts.value.filter(t => t.id !== id)
    }

    // --- General Confirm Dialog ---
    const confirmDialog = ref({
        visible: false,
        title: '',
        message: '',
        confirmLabel: 'Confirmer',
        cancelLabel: 'Annuler',
        confirmClass: 'primary', // 'primary' | 'danger'
        resolve: null
    })

    const showConfirm = ({ title, message, confirmLabel = 'Confirmer', cancelLabel = 'Annuler', confirmClass = 'danger' } = {}) => {
        return new Promise((resolve) => {
            confirmDialog.value = { visible: true, title, message, confirmLabel, cancelLabel, confirmClass, resolve }
        })
    }

    const resolveConfirm = (result) => {
        if (confirmDialog.value.resolve) confirmDialog.value.resolve(result)
        confirmDialog.value = { ...confirmDialog.value, visible: false, resolve: null }
    }

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
        deleteDialog,
        toasts,
        confirmDialog,
        showError,
        showWarning,
        showInfo,
        closeAlert,
        requestDeleteConfirmation,
        closeDeleteDialog,
        showToast,
        dismissToast,
        showConfirm,
        resolveConfirm,
    }
})
