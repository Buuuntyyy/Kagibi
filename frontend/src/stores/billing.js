import { defineStore } from 'pinia'
import api from '../api'

export const useBillingStore = defineStore('billing', {
  state: () => ({
    // Billing status (kept for fallback compatibility in other components)
    enabled: true,
    providerType: 'mock',
    features: {
      quotas: true
    },

    // Current plan & usage info
    currentPlan: null,
    currentUsage: null,

    // States
    loading: false,
    error: null,
  }),

  getters: {
    // Check if billing/quotas are enabled
    isBillingEnabled: (state) => state.enabled,

    // Check if self-hosted mode (no billing)
    isSelfHosted: (state) => !state.enabled || state.providerType === 'disabled',

    // Has quota enforcement
    hasQuotas: (state) => state.enabled && state.features.quotas,

    // Get the current plan code
    planCode: (state) => state.currentPlan?.code || 'free',

    // Get storage usage in GB
    storageUsageGB: (state) => {
      return state.currentUsage?.storage_used_gb || 0
    },

    // Get bandwidth usage in GB
    bandwidthUsageGB: (state) => {
      return state.currentUsage?.bandwidth_used_gb || 0
    },

    // Get active P2P shares count
    p2pActiveShares: (state) => {
      return state.currentUsage?.p2p_shares_active || 0
    },

    // Get P2P shares limit from current plan
    p2pSharesLimit: (state) => {
      return state.currentPlan?.p2p_shares_limit || 5
    }
  },

  actions: {
    // Fetch billing status (enabled/disabled)
    async fetchBillingStatus() {
      try {
        const response = await api.get('/billing/status')
        this.enabled = response.data.enabled
        this.providerType = response.data.provider_type
        this.features = response.data.features || { quotas: true }
      } catch (error) {
        const status = error?.response?.status
        if (status === 401 || status === 404) {
          // Endpoint absent ou non authentifié : billing désactivé silencieusement.
          this.enabled = false
          this.providerType = 'disabled'
        } else {
          console.error('[BillingStore] Failed to fetch billing status:', error)
          this.enabled = false
          this.providerType = 'disabled'
        }
      }
    },

    // Fetch current plan and subscription info (for quotas limit)
    async fetchCurrentPlan() {
      this.loading = true
      this.error = null

      try {
        const response = await api.get('/billing/plan')
        this.currentPlan = response.data
      } catch (err) {
        console.error('[BillingStore] Failed to fetch plan:', err)
        this.error = err.response?.data?.error || 'Impossible de chager les informations de plan'

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

    // Fetch current usage
    async fetchUsage() {
      try {
        const response = await api.get('/billing/usage')
        this.currentUsage = response.data
      } catch (err) {
        console.error('[BillingStore] Failed to fetch usage:', err)
      }
    },

    // Check P2P quota before creating a share
    async checkP2PQuota() {
      try {
        const response = await api.get('/billing/quota/p2p')
        return response.data
      } catch (err) {
        // Fail-open
        return { allowed: true }
      }
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
        quotas: true
      }
      this.currentPlan = null
      this.currentUsage = null
      this.loading = false
      this.error = null
    }
  }
})
