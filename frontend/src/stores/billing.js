import { defineStore } from 'pinia'
import api from '../api'

export const useBillingStore = defineStore('billing', {
  state: () => ({
    // Billing status
    enabled: true,
    providerType: 'mock',
    features: {
      subscriptions: true,
      quotas: true,
      invoices: false
    },

    // Current plan info
    currentPlan: null,
    subscriptions: [],
    currentUsage: null,

    // Invoices
    invoices: [],
    pendingInvoice: null,

    // Loading states
    loading: false,
    loadingInvoices: false,

    // Error states
    error: null,

    // Customer info
    customer: {
      id: null,
      externalId: null,
      email: null,
      name: null
    }
  }),

  getters: {
    // Check if billing is enabled
    isBillingEnabled: (state) => state.enabled,

    // Check if self-hosted mode (no billing)
    isSelfHosted: (state) => !state.enabled || state.providerType === 'disabled',

    // Should show subscription UI
    showSubscriptionUI: (state) => state.enabled && state.features.subscriptions,

    // Should show invoices
    showInvoices: (state) => state.enabled && state.features.invoices,

    // Has quota enforcement
    hasQuotas: (state) => state.enabled && state.features.quotas,

    // Get the current plan code
    planCode: (state) => state.currentPlan?.code || 'free',

    // Check if user has a paid plan
    isPaidPlan: (state) => {
      return state.currentPlan && state.currentPlan.price_monthly_cents > 0
    },

    // Check if user has any active subscription
    hasActiveSubscription: (state) => {
      return state.subscriptions.some(sub => sub.status === 'active')
    },

    // Get total usage amount for current period
    currentUsageAmount: (state) => {
      return state.currentUsage?.total_amount_cents || 0
    },

    // Format amount in euros
    formatAmount: () => (cents, currency = 'EUR') => {
      return new Intl.NumberFormat('fr-FR', {
        style: 'currency',
        currency: currency
      }).format(cents / 100)
    },

    // Check if there's a pending payment
    hasPendingPayment: (state) => {
      return state.pendingInvoice && state.pendingInvoice.payment_link_url
    },

    // Get storage usage in GB
    storageUsageGB: (state) => {
      if (!state.currentUsage?.charges) return 0
      const storageCharge = state.currentUsage.charges.find(c => c.code === 'storage_gb')
      return storageCharge ? parseFloat(storageCharge.units) : 0
    },

    // Get P2P usage in MB
    p2pUsageMB: (state) => {
      if (!state.currentUsage?.charges) return 0
      const p2pCharge = state.currentUsage.charges.find(c => c.code === 'p2p_mb')
      return p2pCharge ? parseFloat(p2pCharge.units) : 0
    }
  },

  actions: {
    // Fetch billing status (enabled/disabled)
    async fetchBillingStatus() {
      try {
        const response = await api.get('/billing/status')
        this.enabled = response.data.enabled
        this.providerType = response.data.provider_type
        this.features = response.data.features
      } catch (error) {
        console.error('[BillingStore] Failed to fetch billing status:', error)
        // Par défaut, considérer comme activé
        this.enabled = true
        this.providerType = 'mock'
      }
    },

    // Fetch current plan and subscription info
    async fetchCurrentPlan() {
      this.loading = true
      this.error = null

      try {
        const response = await api.get('/billing/plan')
        const data = response.data

        this.currentPlan = data

      } catch (err) {
        console.error('[BillingStore] Failed to fetch plan:', err)
        this.error = err.response?.data?.error || 'Impossible de charger les informations de facturation'

        // Set default free plan on error
        this.currentPlan = {
          code: 'free',
          name: 'Plan Gratuit',
          price_monthly_cents: 0,
          currency: 'EUR',
          interval: 'monthly'
        }
      } finally {
        this.loading = false
      }
    },

    // Fetch all invoices
    async fetchInvoices() {
      if (!this.showInvoices) {
        this.invoices = []
        return
      }

      this.loadingInvoices = true

      try {
        const response = await api.get('/billing/invoices')
        this.invoices = response.data || []
      } catch (err) {
        console.error('[BillingStore] Failed to fetch invoices:', err)
        this.invoices = []
      } finally {
        this.loadingInvoices = false
      }
    },

    // Fetch pending invoices only
    async fetchPendingInvoices() {
      if (!this.showInvoices) {
        return
      }

      try {
        const response = await api.get('/billing/invoices')
        const invoices = response.data || []
        const pending = invoices.find(inv => inv.status === 'pending')
        if (pending) {
          this.pendingInvoice = pending
        }
      } catch (err) {
        console.error('[BillingStore] Failed to fetch pending invoices:', err)
      }
    },

    // Get payment link for an invoice and redirect
    async payInvoice(invoiceId) {
      try {
        const response = await api.get(`/billing/invoices/${invoiceId}/payment-link`)
        const paymentUrl = response.data.payment_url

        if (paymentUrl) {
          // Open payment in new tab
          window.open(paymentUrl, '_blank')
          return true
        }

        return false
      } catch (err) {
        console.error('[BillingStore] Failed to get payment link:', err)
        this.error = err.response?.data?.error || 'Impossible de récupérer le lien de paiement'
        return false
      }
    },

    // Pay pending invoice directly
    async payPendingInvoice() {
      if (!this.pendingInvoice?.id) {
        return false
      }
      return this.payInvoice(this.pendingInvoice.id)
    },

    // Clear error
    clearError() {
      this.error = null
    },

    // Reset store
    reset() {
      this.enabled = true
      this.providerType = 'mock'
      this.features = {
        subscriptions: true,
        quotas: true,
        invoices: false
      }
      this.currentPlan = null
      this.subscriptions = []
      this.currentUsage = null
      this.invoices = []
      this.pendingInvoice = null
      this.loading = false
      this.loadingInvoices = false
      this.error = null
      this.customer = {
        id: null,
        externalId: null,
        email: null,
        name: null
      }
    }
  }
})
