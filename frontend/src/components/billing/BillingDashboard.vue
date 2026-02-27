<template>
  <div class="account-page">
    <div class="page-header">
      <div class="header-content">
        <button class="btn-back" @click="router.push('/dashboard')">
            <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M19 12H5M12 19l-7-7 7-7"/>
            </svg>
            {{ t('billing.backToAccount') }}
        </button>
        <h1>{{ t('billing.title') }}</h1>
      </div>
      <p class="subtitle">{{ t('billing.subtitle') }}</p>
    </div>

    <!-- Stripe return success banner -->
    <div v-if="stripeSuccess" class="stripe-banner success-banner">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="20" height="20">
        <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
        <polyline points="22 4 12 14.01 9 11.01"/>
      </svg>
      <span>{{ t('billing.checkoutSuccess') || 'Paiement réussi ! Votre abonnement est en cours d\'activation.' }}</span>
    </div>

    <!-- Loading State -->
    <div v-if="billingStore.loading && !billingStore.currentPlan" class="loading-state">
      <div class="spinner"></div>
      <p>{{ t('billing.loading') }}</p>
    </div>

    <div v-else class="content-grid">

      <!-- Plan Section -->
      <section class="settings-section plan-section">
        <div class="section-header">
          <h3>{{ t('billing.currentPlan') }}</h3>
          <span :class="['plan-badge', planBadgeClass]">{{ currentPlan?.name || t('billing.free') }}</span>
        </div>
        <div class="section-body">
            <div class="plan-details-grid">
                <div class="plan-info">
                    <div class="price-display">
                        <span class="amount">{{ formatPrice(currentPlan?.price_monthly_cents) }}</span>
                        <span class="period" v-if="currentPlan?.price_monthly_cents > 0">/ {{ intervalLabel }}</span>
                    </div>
                    <div class="storage-limit" v-if="currentPlan?.storage_limit_gb">
                        {{ currentPlan.storage_limit_gb >= 1000 ? (currentPlan.storage_limit_gb / 1000) + ' To' : currentPlan.storage_limit_gb + ' Go' }} {{ t('billing.storage') || 'de stockage' }}
                    </div>
                </div>
                <div class="plan-actions">
                     <button v-if="!billingStore.isPaidPlan && billingStore.canCheckout" class="btn-primary" @click="showUpgradeModal = true">
                       {{ t('billing.upgradePlan') }}
                     </button>
                     <button v-if="billingStore.isPaidPlan && billingStore.canUsePortal" class="btn-outline-primary" @click="openPortal" :disabled="portalLoading">
                       {{ portalLoading ? '...' : (t('billing.manageSubscription') || 'Gérer l\'abonnement') }}
                     </button>
                </div>
            </div>

            <div class="features-list">
                 <div class="feature-item">
                    <svg class="check-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <polyline points="20 6 9 17 4 12"></polyline>
                    </svg>
                    <span>{{ t('billing.e2eEncryption') }}</span>
                 </div>
                 <div class="feature-item" v-if="currentPlan?.features?.p2p_enabled">
                    <svg class="check-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <polyline points="20 6 9 17 4 12"></polyline>
                    </svg>
                    <span>{{ t('billing.p2pEnabled') || 'Transferts P2P activés' }}</span>
                 </div>
                 <div class="feature-item" v-if="currentPlan?.features?.max_file_size_mb">
                    <svg class="check-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <polyline points="20 6 9 17 4 12"></polyline>
                    </svg>
                    <span>{{ t('billing.maxFileSize') || 'Taille max par fichier' }}: {{ currentPlan.features.max_file_size_mb >= 1024 ? (currentPlan.features.max_file_size_mb / 1024) + ' Go' : currentPlan.features.max_file_size_mb + ' Mo' }}</span>
                 </div>
            </div>
        </div>
      </section>

      <!-- Usage Section -->
      <section class="settings-section">
          <div class="section-header">
              <h3>{{ t('billing.usage') }}</h3>
          </div>
          <div class="section-body">
              <div class="usage-stats">
                  <div class="stat-card">
                      <div class="stat-icon">
                          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                              <path d="M22 12h-4l-3 9L9 3l-3 9H2"/>
                          </svg>
                      </div>
                      <div class="stat-info">
                          <span class="stat-label">{{ t('billing.storageUsed') }}</span>
                          <span class="stat-value">{{ billingStore.storageUsageGB.toFixed(2) }} <small>Go</small></span>
                      </div>
                      <div class="stat-bar" v-if="currentPlan?.storage_limit_gb">
                          <div class="stat-bar-fill" :style="{ width: storagePercent + '%' }" :class="{ 'bar-warning': storagePercent > 80, 'bar-danger': storagePercent > 95 }"></div>
                      </div>
                  </div>
                  <div class="stat-card">
                       <div class="stat-icon">
                          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                              <path d="M12 2L2 7l10 5 10-5-10-5z"/>
                              <path d="M2 17l10 5 10-5"/>
                              <path d="M2 12l10 5 10-5"/>
                          </svg>
                      </div>
                      <div class="stat-info">
                          <span class="stat-label">{{ t('billing.bandwidthUsed') || 'Bande passante' }}</span>
                          <span class="stat-value">{{ billingStore.bandwidthUsageGB.toFixed(2) }} <small>Go</small></span>
                      </div>
                  </div>
                  <div class="stat-card" v-if="currentPlan?.features?.p2p_enabled">
                       <div class="stat-icon">
                          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                              <polyline points="16 16 12 12 8 16"></polyline>
                              <line x1="12" y1="12" x2="12" y2="21"></line>
                              <path d="M20.39 18.39A5 5 0 0 0 18 9h-1.26A8 8 0 1 0 3 16.3"></path>
                              <polyline points="16 16 12 12 8 16"></polyline>
                          </svg>
                      </div>
                      <div class="stat-info">
                          <span class="stat-label">{{ t('billing.p2pTransfers') }}</span>
                          <span class="stat-value">{{ billingStore.p2pUsageGB.toFixed(2) }} <small>Go</small></span>
                      </div>
                  </div>
              </div>
          </div>
      </section>

      <!-- Payment Alert -->
      <section v-if="billingStore.hasPendingPayment" class="settings-section warning-section">
          <div class="section-body alert-body">
              <div class="alert-icon-wrapper">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="12" cy="12" r="10"></circle>
                    <line x1="12" y1="8" x2="12" y2="12"></line>
                    <line x1="12" y1="16" x2="12.01" y2="16"></line>
                  </svg>
              </div>
              <div class="alert-content">
                  <h4>{{ t('billing.pending') }}</h4>
                  <p>{{ t('billing.invoiceNumber') }} {{ billingStore.pendingInvoice.number }} : {{ formatPrice(billingStore.pendingInvoice.amount_cents, billingStore.pendingInvoice.currency) }}</p>
              </div>
              <button class="btn-primary" @click="payPendingInvoice" :disabled="paymentLoading">
                  {{ paymentLoading ? '...' : t('billing.pay') }}
              </button>
          </div>
      </section>

      <!-- Invoices Section -->
      <section v-if="billingStore.showInvoices" class="settings-section">
        <div class="section-header">
            <h3>{{ t('billing.invoices') }}</h3>
        </div>
        <div class="section-body no-padding">
            <div v-if="billingStore.invoices.length === 0" class="empty-state">
                <p>{{ t('billing.noInvoices') }}</p>
            </div>
            <table v-else class="data-table">
                <thead>
                    <tr>
                        <th>{{ t('billing.date') || 'Date' }}</th>
                        <th>{{ t('billing.invoiceNumber') }}</th>
                        <th>{{ t('billing.amount') }}</th>
                        <th>{{ t('billing.status') }}</th>
                        <th>{{ t('billing.actions') }}</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="invoice in billingStore.invoices" :key="invoice.id">
                        <td>{{ formatDate(invoice.issued_at) }}</td>
                        <td class="mono">{{ invoice.number }}</td>
                        <td class="font-medium">{{ formatPrice(invoice.amount_cents, invoice.currency) }}</td>
                        <td>
                            <span :class="['status-pill', getPaymentStatusClass(invoice.status)]">
                                {{ getPaymentStatusLabel(invoice.status) }}
                            </span>
                        </td>
                        <td>
                             <button
                                v-if="invoice.status === 'open' && invoice.payment_url"
                                class="btn-sm btn-outline"
                                @click="payInvoice(invoice.id)"
                              >
                                {{ t('billing.pay') }}
                              </button>
                              <a v-if="invoice.download_url" :href="invoice.download_url" target="_blank" class="download-link" :title="t('billing.download')">
                                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16">
                                      <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                                      <polyline points="7 10 12 15 17 10"></polyline>
                                      <line x1="12" y1="15" x2="12" y2="3"></line>
                                  </svg>
                              </a>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
      </section>
    </div>

    <!-- Upgrade Modal -->
    <Teleport to="body">
      <div v-if="showUpgradeModal" class="modal-overlay" @click.self="showUpgradeModal = false">
        <div class="modal-content plans-modal">
          <div class="modal-header">
            <h2>{{ t('billing.choosePlan') || 'Choisir un plan' }}</h2>
            <button class="btn-close" @click="showUpgradeModal = false">&times;</button>
          </div>
          <div class="modal-body">
            <div v-if="billingStore.loadingPlans" class="loading-state" style="height: 200px;">
              <div class="spinner"></div>
            </div>
            <div v-else class="plans-grid">
              <div
                v-for="plan in billingStore.availablePlans"
                :key="plan.code"
                :class="['plan-card', { 'plan-featured': plan.code === 'expert', 'plan-current': plan.code === billingStore.planCode }]"
              >
                <div class="plan-card-badge" v-if="plan.code === 'expert'">Populaire</div>
                <h3 class="plan-card-name">{{ plan.name }}</h3>
                <div class="plan-card-price">
                  <span class="plan-card-amount">{{ formatPrice(plan.price_monthly_cents) }}</span>
                  <span class="plan-card-interval" v-if="plan.price_monthly_cents > 0">/mois</span>
                </div>
                <ul class="plan-card-features">
                  <li>{{ plan.storage_limit_gb >= 1000 ? (plan.storage_limit_gb / 1000) + ' To' : plan.storage_limit_gb + ' Go' }} de stockage</li>
                  <li v-if="plan.features?.max_file_size_mb">Fichiers jusqu'à {{ plan.features.max_file_size_mb >= 1024 ? (plan.features.max_file_size_mb / 1024) + ' Go' : plan.features.max_file_size_mb + ' Mo' }}</li>
                  <li v-if="plan.features?.p2p_enabled">P2P activé{{ plan.features.p2p_limit_gb ? ' (' + plan.features.p2p_limit_gb + ' Go/mois)' : '' }}</li>
                  <li v-else>P2P non disponible</li>
                  <li>Chiffrement E2E</li>
                </ul>
                <button
                  v-if="plan.code !== billingStore.planCode && plan.price_monthly_cents > 0"
                  class="btn-primary plan-card-btn"
                  :disabled="checkoutLoading"
                  @click="handleUpgrade(plan.code)"
                >
                  {{ checkoutLoading ? '...' : (t('billing.subscribe') || 'S\'abonner') }}
                </button>
                <span v-else-if="plan.code === billingStore.planCode" class="plan-card-current">
                  {{ t('billing.currentPlanLabel') || 'Plan actuel' }}
                </span>
              </div>
            </div>
            <p v-if="billingStore.error" class="error-text">{{ billingStore.error }}</p>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useBillingStore } from '../../stores/billing'
import { useRouter, useRoute } from 'vue-router'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()
const billingStore = useBillingStore()

const paymentLoading = ref(false)
const showUpgradeModal = ref(false)
const checkoutLoading = ref(false)
const portalLoading = ref(false)
const stripeSuccess = ref(false)

const currentPlan = computed(() => billingStore.currentPlan)

const intervalLabel = computed(() => {
  const interval = currentPlan.value?.interval
  if (interval === 'monthly') return 'mois'
  if (interval === 'yearly') return 'an'
  return 'mois'
})

const planBadgeClass = computed(() => {
  const code = currentPlan.value?.code
  if (code === 'free') return 'badge-free'
  if (code === 'personal') return 'badge-personal'
  if (code === 'expert') return 'badge-expert'
  if (code === 'enterprise') return 'badge-enterprise'
  return 'badge-free'
})

const storagePercent = computed(() => {
  const limit = currentPlan.value?.storage_limit_gb || 5
  const used = billingStore.storageUsageGB
  return Math.min((used / limit) * 100, 100)
})

// Formatting helpers
function formatPrice(cents, currency = 'EUR') {
  if (!cents && cents !== 0) return '0,00 €'
  return billingStore.formatAmount(cents, currency)
}

function formatDate(dateString) {
  if (!dateString) return '-'
  const date = new Date(dateString)
  return date.toLocaleDateString('fr-FR', {
    day: 'numeric',
    month: 'short',
    year: 'numeric'
  })
}

function getPaymentStatusClass(status) {
  const classes = { paid: 'success', open: 'warning', draft: 'info', void: 'muted', uncollectible: 'error' }
  return classes[status] || 'unknown'
}

function getPaymentStatusLabel(status) {
  const labels = { paid: 'Payée', open: 'En attente', draft: 'Brouillon', void: 'Annulée', uncollectible: 'Impayée' }
  return labels[status] || status
}

// Actions
async function handleUpgrade(planCode) {
  checkoutLoading.value = true
  billingStore.clearError()
  try {
    await billingStore.initiateCheckout(planCode)
    // If successful, user is redirected to Stripe Checkout
  } finally {
    checkoutLoading.value = false
  }
}

async function openPortal() {
  portalLoading.value = true
  try {
    await billingStore.openPortal()
    // If successful, user is redirected to Stripe Portal
  } finally {
    portalLoading.value = false
  }
}

async function payPendingInvoice() {
  paymentLoading.value = true
  try { await billingStore.payPendingInvoice() }
  finally { paymentLoading.value = false }
}

async function payInvoice(invoiceId) {
  await billingStore.payInvoice(invoiceId)
}

onMounted(async () => {
  // Detect Stripe checkout return
  if (route.query.checkout === 'success') {
    stripeSuccess.value = true
    // Clean URL
    router.replace({ query: {} })
    setTimeout(() => { stripeSuccess.value = false }, 8000)
  }

  // Fetch billing data in parallel
  await Promise.all([
    billingStore.fetchBillingStatus(),
    billingStore.fetchCurrentPlan()
  ])

  // After status is known, fetch additional data
  const promises = []
  if (billingStore.showInvoices) promises.push(billingStore.fetchInvoices())
  if (billingStore.canCheckout) promises.push(billingStore.fetchPlans())
  promises.push(billingStore.fetchUsage())
  await Promise.all(promises)
})
</script>

<style scoped>
/* Page Layout */
.account-page {
  width: 100%;
  height: 100%;
  margin: 0;
  padding: 40px 10%;
  overflow-y: auto;
  background-color: var(--background-color);
  animation: fadeIn 0.4s ease;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.page-header { margin-bottom: 40px; }
.header-content { display: flex; align-items: center; gap: 16px; margin-bottom: 8px; }

.btn-back {
  background: none; border: none; display: flex; align-items: center; justify-content: center;
  gap: 6px; color: var(--secondary-text-color); cursor: pointer; font-size: 0.9rem;
  padding: 6px 12px; border-radius: 8px; transition: all 0.2s;
}
.btn-back:hover { background-color: var(--hover-background-color); color: var(--primary-color); }

.page-header h1 { font-size: 2rem; font-weight: 700; color: var(--main-text-color); margin: 0; }
.subtitle { color: var(--secondary-text-color); margin: 0; font-size: 1.1rem; }

/* Stripe Success Banner */
.stripe-banner {
  display: flex; align-items: center; gap: 12px; padding: 14px 20px;
  border-radius: 10px; margin-bottom: 24px; font-weight: 500;
}
.success-banner {
  background: rgba(34, 197, 94, 0.1); color: var(--success-color);
  border: 1px solid rgba(34, 197, 94, 0.3);
}

/* Loading */
.loading-state { display: flex; flex-direction: column; align-items: center; justify-content: center; height: 400px; color: var(--secondary-text-color); }
.spinner { width: 40px; height: 40px; border: 3px solid var(--border-color); border-top-color: var(--primary-color); border-radius: 50%; animation: spin 1s infinite linear; margin: 0 auto 16px; }
@keyframes spin { to { transform: rotate(360deg); } }

/* Grid & Cards */
.content-grid { display: grid; gap: 24px; }

.settings-section {
  background: var(--card-color); border-radius: 12px; border: 1px solid var(--border-color);
  overflow: hidden; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.05);
}
.warning-section { border-color: var(--warning-color); background: #fffbf0; }
[data-theme="dark"] .warning-section { background: rgba(245, 158, 11, 0.1); }

.section-header {
  padding: 20px 24px; border-bottom: 1px solid var(--border-color);
  display: flex; justify-content: space-between; align-items: center;
}
.section-header h3 { margin: 0; font-size: 1.1rem; font-weight: 600; color: var(--main-text-color); }
.section-body { padding: 24px; }
.section-body.no-padding { padding: 0; }

/* Plan Badge */
.plan-badge { padding: 6px 12px; border-radius: 20px; font-size: 0.85rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.5px; }
.badge-free { background: var(--hover-background-color); color: var(--secondary-text-color); }
.badge-personal { background: var(--primary-color); color: white; }
.badge-expert { background: linear-gradient(135deg, #8b5cf6, #6d28d9); color: white; }
.badge-enterprise { background: linear-gradient(135deg, #f59e0b, #d97706); color: white; }

/* Plan Section Details */
.plan-details-grid { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.price-display { display: flex; align-items: baseline; }
.amount { font-size: 2.5rem; font-weight: 700; color: var(--main-text-color); }
.period { font-size: 1rem; color: var(--secondary-text-color); margin-left: 6px; }
.storage-limit { color: var(--secondary-text-color); font-size: 0.95rem; margin-top: 4px; }

.features-list { display: grid; gap: 12px; padding-top: 16px; border-top: 1px solid var(--border-color); }
.feature-item { display: flex; align-items: center; gap: 12px; color: var(--main-text-color); }
.check-icon { width: 20px; height: 20px; color: var(--success-color); }

/* Stats */
.usage-stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; }
.stat-card { display: flex; flex-wrap: wrap; align-items: center; gap: 16px; padding: 16px; background: var(--hover-background-color); border-radius: 10px; }
.stat-icon { width: 40px; height: 40px; border-radius: 8px; background: var(--card-color); display: flex; align-items: center; justify-content: center; color: var(--primary-color); }
.stat-icon svg { width: 20px; height: 20px; }
.stat-info { display: flex; flex-direction: column; }
.stat-label { font-size: 0.85rem; color: var(--secondary-text-color); }
.stat-value { font-size: 1.25rem; font-weight: 600; color: var(--main-text-color); }
.stat-value small { font-size: 0.85rem; color: var(--secondary-text-color); }

/* Usage Progress Bar */
.stat-bar { width: 100%; height: 6px; background: var(--border-color); border-radius: 3px; overflow: hidden; }
.stat-bar-fill { height: 100%; background: var(--primary-color); border-radius: 3px; transition: width 0.5s ease; }
.stat-bar-fill.bar-warning { background: var(--warning-color); }
.stat-bar-fill.bar-danger { background: var(--error-color); }

/* Buttons */
.btn-primary {
  background: var(--primary-color); color: white; border: none; padding: 10px 20px;
  border-radius: 8px; font-weight: 600; cursor: pointer; transition: all 0.2s;
}
.btn-primary:hover:not(:disabled) { background: var(--accent-color); transform: translateY(-1px); }
.btn-primary:disabled { opacity: 0.7; cursor: not-allowed; }

.btn-outline-primary {
  background: transparent; border: 1px solid var(--primary-color); color: var(--primary-color);
  padding: 10px 20px; border-radius: 8px; font-weight: 600; cursor: pointer; transition: all 0.2s;
}
.btn-outline-primary:hover:not(:disabled) { background: var(--primary-color); color: white; }
.btn-outline-primary:disabled { opacity: 0.7; cursor: not-allowed; }

.btn-sm { padding: 6px 12px; font-size: 0.85rem; border-radius: 6px; cursor: pointer; transition: 0.2s; }
.btn-outline { background: transparent; border: 1px solid var(--border-color); color: var(--main-text-color); }
.btn-outline:hover { border-color: var(--primary-color); color: var(--primary-color); }

/* Table */
.data-table { width: 100%; border-collapse: collapse; }
.data-table th, .data-table td { padding: 16px 24px; text-align: left; }
.data-table th { background: var(--hover-background-color); color: var(--secondary-text-color); font-weight: 600; font-size: 0.85rem; }
.data-table td { border-top: 1px solid var(--border-color); color: var(--main-text-color); }
.data-table tr:hover td { background-color: rgba(0,0,0,0.02); }

.mono { font-family: monospace; font-size: 0.9em; }
.font-medium { font-weight: 500; }

.status-pill { padding: 4px 10px; border-radius: 12px; font-size: 0.75rem; font-weight: 600; display: inline-block; }
.status-pill.success { background: rgba(34, 197, 94, 0.1); color: var(--success-color); }
.status-pill.warning { background: rgba(245, 158, 11, 0.1); color: var(--warning-color); }
.status-pill.error { background: rgba(239, 68, 68, 0.1); color: var(--error-color); }
.status-pill.info { background: rgba(59, 130, 246, 0.1); color: #3b82f6; }
.status-pill.muted { background: var(--hover-background-color); color: var(--secondary-text-color); }

.download-link { color: var(--secondary-text-color); margin-left: 10px; opacity: 0.7; transition: 0.2s; }
.download-link:hover { opacity: 1; color: var(--primary-color); }

.alert-body { display: flex; align-items: center; gap: 16px; }
.alert-icon-wrapper { color: var(--warning-color); }
.alert-icon-wrapper svg { width: 24px; height: 24px; }
.alert-content { flex: 1; }
.alert-content h4 { margin: 0 0 4px 0; color: var(--main-text-color); }
.alert-content p { margin: 0; color: var(--secondary-text-color); font-size: 0.9rem; }

.empty-state { text-align: center; padding: 40px; color: var(--secondary-text-color); font-style: italic; }

.error-text { color: var(--error-color); text-align: center; margin-top: 16px; font-size: 0.9rem; }

/* Modal */
.modal-overlay {
  position: fixed; inset: 0; background: rgba(0, 0, 0, 0.5); display: flex;
  align-items: center; justify-content: center; z-index: 1000; backdrop-filter: blur(4px);
}
.modal-content {
  background: var(--card-color); border-radius: 16px; max-width: 900px; width: 90%;
  max-height: 90vh; overflow-y: auto; box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
}
.modal-header {
  display: flex; justify-content: space-between; align-items: center; padding: 24px 28px;
  border-bottom: 1px solid var(--border-color);
}
.modal-header h2 { margin: 0; font-size: 1.4rem; font-weight: 700; color: var(--main-text-color); }
.btn-close { background: none; border: none; font-size: 1.8rem; cursor: pointer; color: var(--secondary-text-color); line-height: 1; padding: 0; }
.btn-close:hover { color: var(--main-text-color); }
.modal-body { padding: 28px; }

/* Plans Grid */
.plans-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 20px; }
.plan-card {
  position: relative; border: 2px solid var(--border-color); border-radius: 12px; padding: 28px 24px;
  text-align: center; transition: all 0.2s; display: flex; flex-direction: column;
}
.plan-card:hover { border-color: var(--primary-color); transform: translateY(-2px); box-shadow: 0 8px 16px rgba(0, 0, 0, 0.08); }
.plan-featured { border-color: #8b5cf6; }
.plan-current { opacity: 0.7; }
.plan-card-badge {
  position: absolute; top: -12px; left: 50%; transform: translateX(-50%);
  background: linear-gradient(135deg, #8b5cf6, #6d28d9); color: white;
  padding: 4px 16px; border-radius: 12px; font-size: 0.75rem; font-weight: 600; white-space: nowrap;
}
.plan-card-name { margin: 8px 0 16px; font-size: 1.2rem; font-weight: 700; color: var(--main-text-color); }
.plan-card-price { margin-bottom: 20px; }
.plan-card-amount { font-size: 2rem; font-weight: 700; color: var(--main-text-color); }
.plan-card-interval { color: var(--secondary-text-color); font-size: 0.9rem; }
.plan-card-features { list-style: none; padding: 0; margin: 0 0 24px; text-align: left; flex: 1; }
.plan-card-features li { padding: 6px 0; color: var(--secondary-text-color); font-size: 0.9rem; border-bottom: 1px solid var(--border-color); }
.plan-card-features li:last-child { border-bottom: none; }
.plan-card-btn { width: 100%; }
.plan-card-current { color: var(--secondary-text-color); font-weight: 600; font-size: 0.9rem; }
</style>
