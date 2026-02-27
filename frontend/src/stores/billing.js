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
      invoices: false,
      checkout: false,
      portal: false
    },

    // Current plan info
    currentPlan: null,
    subscriptions: [],
    currentUsage: null,

    // Available plans
    availablePlans: [],

    // Invoices
    invoices: [],
    pendingInvoice: null,

    // Loading states
    loading: false,
    loadingInvoices: false,
    loadingPlans: false,

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

    // Can use Stripe Checkout
    canCheckout: (state) => state.enabled && state.features.checkout,

    // Can use Stripe Customer Portal
    canUsePortal: (state) => state.enabled && state.features.portal,

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
      return state.pendingInvoice && state.pendingInvoice.payment_url
    },

    // Get storage usage in GB
    storageUsageGB: (state) => {
      return state.currentUsage?.storage_used_gb || 0
    },

    // Get bandwidth usage in GB
    bandwidthUsageGB: (state) => {
      return state.currentUsage?.bandwidth_used_gb || 0
    },

    // Get P2P usage in GB
    p2pUsageGB: (state) => {
      return state.currentUsage?.p2p_transfer_gb || 0
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
        this.currentPlan = response.data
      } catch (err) {
        console.error('[BillingStore] Failed to fetch plan:', err)
        this.error = err.response?.data?.error || 'Impossible de charger les informations de facturation'

        // Set default free plan on error
        this.currentPlan = {
          code: 'free',
          name: 'Gratuit',
          price_monthly_cents: 0,
          currency: 'EUR',
          interval: 'monthly'
        }
      } finally {
        this.loading = false
      }
    },

    // Fetch available plans
    async fetchPlans() {
      this.loadingPlans = true
      try {
        const response = await api.get('/billing/plans')
        this.availablePlans = response.data || []
        return this.availablePlans
      } catch (err) {
        console.error('[BillingStore] Failed to fetch plans:', err)
        this.availablePlans = []
        return []
      } finally {
        this.loadingPlans = false
      }
    },

    // Fetch current usage
    async fetchUsage() {
      try {
        const response = await api.get('/billing/usage')
        this.currentUsage = response.data
      } catch (err) {
        console.error('[BillingStore] Failed to fetch usage:', err)
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

        // Find pending invoice
        const pending = this.invoices.find(inv => inv.status === 'open' || inv.status === 'pending')
        if (pending) {
          this.pendingInvoice = pending
        }
      } catch (err) {
        console.error('[BillingStore] Failed to fetch invoices:', err)
        this.invoices = []
      } finally {
        this.loadingInvoices = false
      }
    },

    // Initiate Stripe Checkout for plan upgrade
    async initiateCheckout(planCode) {
      this.loading = true
      this.error = null

      try {
        const response = await api.post('/billing/checkout', {
          plan_code: planCode
        })

        const checkoutUrl = response.data.checkout_url
        if (checkoutUrl) {
          window.location.href = checkoutUrl
          return true
        }

        throw new Error('No checkout URL returned')
      } catch (err) {
        console.error('[BillingStore] Checkout failed:', err)
        this.error = err.response?.data?.error || 'Impossible de créer la session de paiement'
        return false
      } finally {
        this.loading = false
      }
    },

    // Open Stripe Customer Portal (manage subscription, payment methods, invoices)
    async openPortal() {
      this.loading = true
      this.error = null

      try {
        const response = await api.post('/billing/portal')

        const portalUrl = response.data.portal_url
        if (portalUrl) {
          window.location.href = portalUrl
          return true
        }

        throw new Error('No portal URL returned')
      } catch (err) {
        console.error('[BillingStore] Portal failed:', err)
        this.error = err.response?.data?.error || 'Impossible d\'ouvrir le portail de gestion'
        return false
      } finally {
        this.loading = false
      }
    },

    // Get payment link for an invoice and redirect
    async payInvoice(invoiceId) {
      try {
        const response = await api.get(`/billing/invoices/${invoiceId}/payment-link`)
        const paymentUrl = response.data.payment_url

        if (paymentUrl) {
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

    // Check quota before upload (called from upload logic)
    async checkQuota(requestedBytes) {
      try {
        const response = await api.post('/billing/quota/check', {
          requested_bytes: requestedBytes
        })
        return response.data
      } catch (err) {
        // Fail-open
        return { allowed: true }
      }
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
        invoices: false,
        checkout: false,
        portal: false
      }
      this.currentPlan = null
      this.subscriptions = []
      this.currentUsage = null
      this.availablePlans = []
      this.invoices = []
      this.pendingInvoice = null
      this.loading = false
      this.loadingInvoices = false
      this.loadingPlans = false
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
