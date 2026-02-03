import { defineStore } from 'pinia'
import api from '../api'

export const useBillingStore = defineStore('billing', {
  state: () => ({
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
    // Get the current plan code
    planCode: (state) => state.currentPlan?.code || 'free',
    
    // Check if user has a paid plan
    isPaidPlan: (state) => {
      return state.currentPlan && state.currentPlan.amount_cents > 0
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
    // Fetch current plan and subscription info
    async fetchCurrentPlan() {
      this.loading = true
      this.error = null
      
      try {
        const response = await api.get('/billing/plan')
        const data = response.data
        
        this.currentPlan = data.current_plan
        this.subscriptions = data.subscriptions || []
        this.currentUsage = data.current_usage
        this.pendingInvoice = data.pending_invoice
        
        this.customer = {
          id: data.customer_id,
          externalId: data.external_id,
          email: data.email,
          name: data.name
        }
        
      } catch (err) {
        console.error('[BillingStore] Failed to fetch plan:', err)
        this.error = err.response?.data?.error || 'Impossible de charger les informations de facturation'
        
        // Set default free plan on error
        this.currentPlan = {
          code: 'free',
          name: 'Plan Gratuit',
          amount_cents: 0,
          currency: 'EUR',
          interval: 'monthly'
        }
      } finally {
        this.loading = false
      }
    },

    // Fetch all invoices
    async fetchInvoices() {
      this.loadingInvoices = true
      
      try {
        const response = await api.get('/billing/invoices')
        this.invoices = response.data.invoices || []
      } catch (err) {
        console.error('[BillingStore] Failed to fetch invoices:', err)
        this.invoices = []
      } finally {
        this.loadingInvoices = false
      }
    },

    // Fetch pending invoices only
    async fetchPendingInvoices() {
      try {
        const response = await api.get('/billing/pending-invoices')
        const invoices = response.data.invoices || []
        if (invoices.length > 0) {
          this.pendingInvoice = invoices[0]
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
