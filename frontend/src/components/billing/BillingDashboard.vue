<template>
  <div class="billing-dashboard">
    <!-- Loading State -->
    <div v-if="billingStore.loading" class="loading-container">
      <div class="spinner"></div>
      <p>Chargement des informations de facturation...</p>
    </div>

    <!-- Error Banner -->
    <div v-if="billingStore.error" class="error-banner">
      <span class="error-icon">⚠️</span>
      <span>{{ billingStore.error }}</span>
      <button @click="billingStore.clearError" class="btn-dismiss">✕</button>
    </div>

    <!-- Main Content -->
    <div v-if="!billingStore.loading" class="billing-content">
      
      <!-- Current Plan Card -->
      <div class="card plan-card">
        <div class="card-header">
          <h2>Mon abonnement</h2>
          <span :class="['plan-badge', planBadgeClass]">{{ currentPlan?.name || 'Free' }}</span>
        </div>
        
        <div class="card-body">
          <div class="plan-details">
            <div class="plan-price">
              <span class="price-amount">{{ formatPrice(currentPlan?.amount_cents) }}</span>
              <span class="price-interval">/ {{ intervalLabel }}</span>
            </div>
            
            <div class="plan-features" v-if="currentPlan">
              <div class="feature" v-if="currentPlan.code === 'free'">
                <span class="feature-icon">📦</span>
                <span>5 Go de stockage</span>
              </div>
              <div class="feature" v-if="currentPlan.code === 'pro'">
                <span class="feature-icon">📦</span>
                <span>100 Go de stockage</span>
              </div>
              <div class="feature" v-if="currentPlan.code === 'pro'">
                <span class="feature-icon">🔄</span>
                <span>Transfert P2P illimité</span>
              </div>
              <div class="feature" v-if="currentPlan.code === 'enterprise'">
                <span class="feature-icon">♾️</span>
                <span>Stockage illimité</span>
              </div>
            </div>
          </div>

          <!-- Subscription Status -->
          <div v-if="activeSubscription" class="subscription-info">
            <div class="info-row">
              <span class="label">Statut</span>
              <span :class="['status-badge', subscriptionStatusClass]">
                {{ subscriptionStatusLabel }}
              </span>
            </div>
            <div class="info-row">
              <span class="label">Depuis le</span>
              <span>{{ formatDate(activeSubscription.started_at) }}</span>
            </div>
            <div class="info-row" v-if="activeSubscription.ending_at">
              <span class="label">Se termine le</span>
              <span class="text-warning">{{ formatDate(activeSubscription.ending_at) }}</span>
            </div>
          </div>
        </div>

        <div class="card-footer" v-if="currentPlan?.code === 'free'">
          <button class="btn btn-primary btn-upgrade" @click="showUpgradeOptions">
            <span class="btn-icon">⬆️</span>
            Passer à Pro
          </button>
        </div>
      </div>

      <!-- Current Usage Card -->
      <div class="card usage-card">
        <div class="card-header">
          <h2>Consommation actuelle</h2>
          <span class="period-label" v-if="billingStore.currentUsage">
            {{ formatPeriod(billingStore.currentUsage.from_datetime, billingStore.currentUsage.to_datetime) }}
          </span>
        </div>

        <div class="card-body">
          <div class="usage-metrics">
            <!-- Storage Usage -->
            <div class="metric">
              <div class="metric-header">
                <span class="metric-icon">💾</span>
                <span class="metric-name">Stockage</span>
              </div>
              <div class="metric-value">
                <span class="value">{{ billingStore.storageUsageGB.toFixed(2) }}</span>
                <span class="unit">Go</span>
              </div>
              <div class="metric-cost">
                {{ formatPrice(getChargeAmount('storage_gb')) }}
              </div>
            </div>

            <!-- P2P Transfer Usage -->
            <div class="metric">
              <div class="metric-header">
                <span class="metric-icon">🔄</span>
                <span class="metric-name">Transfert P2P</span>
              </div>
              <div class="metric-value">
                <span class="value">{{ billingStore.p2pUsageMB.toFixed(0) }}</span>
                <span class="unit">Mo</span>
              </div>
              <div class="metric-cost">
                {{ formatPrice(getChargeAmount('p2p_mb')) }}
              </div>
            </div>
          </div>

          <!-- Total Usage -->
          <div class="usage-total">
            <span class="total-label">Total estimé ce mois</span>
            <span class="total-amount">{{ formatPrice(billingStore.currentUsageAmount) }}</span>
          </div>
        </div>
      </div>

      <!-- Pending Payment Alert -->
      <div v-if="billingStore.hasPendingPayment" class="card payment-alert-card">
        <div class="card-body">
          <div class="payment-alert">
            <div class="alert-icon">💳</div>
            <div class="alert-content">
              <h3>Paiement en attente</h3>
              <p>
                Facture {{ billingStore.pendingInvoice.number }} - 
                {{ formatPrice(billingStore.pendingInvoice.total_amount_cents, billingStore.pendingInvoice.currency) }}
              </p>
            </div>
            <button class="btn btn-primary" @click="payPendingInvoice" :disabled="paymentLoading">
              <span v-if="paymentLoading" class="spinner-small"></span>
              <span v-else>Payer maintenant</span>
            </button>
          </div>
        </div>
      </div>

      <!-- Invoice History -->
      <div class="card invoices-card">
        <div class="card-header">
          <h2>Historique des factures</h2>
          <button class="btn btn-secondary btn-sm" @click="refreshInvoices" :disabled="billingStore.loadingInvoices">
            <span v-if="billingStore.loadingInvoices" class="spinner-small"></span>
            <span v-else>🔄 Actualiser</span>
          </button>
        </div>

        <div class="card-body">
          <div v-if="billingStore.invoices.length === 0" class="empty-state">
            <span class="empty-icon">📄</span>
            <p>Aucune facture pour le moment</p>
          </div>

          <table v-else class="invoices-table">
            <thead>
              <tr>
                <th>Numéro</th>
                <th>Date</th>
                <th>Montant</th>
                <th>Statut</th>
                <th>Action</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="invoice in billingStore.invoices" :key="invoice.id">
                <td>{{ invoice.invoice_number }}</td>
                <td>{{ formatDate(invoice.issuing_date) }}</td>
                <td>{{ formatPrice(invoice.total_amount_cents, invoice.currency) }}</td>
                <td>
                  <span :class="['status-badge', getPaymentStatusClass(invoice.payment_status)]">
                    {{ getPaymentStatusLabel(invoice.payment_status) }}
                  </span>
                </td>
                <td>
                  <button
                    v-if="invoice.payment_status !== 'succeeded' && invoice.payment_link_url"
                    class="btn btn-sm btn-primary"
                    @click="payInvoice(invoice.lago_invoice_id)"
                  >
                    Payer
                  </button>
                  <span v-else-if="invoice.payment_status === 'succeeded'" class="text-success">✓</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useBillingStore } from '../../stores/billing'

const billingStore = useBillingStore()
const paymentLoading = ref(false)

// Computed properties
const currentPlan = computed(() => billingStore.currentPlan)

const activeSubscription = computed(() => {
  return billingStore.subscriptions.find(sub => sub.status === 'active')
})

const intervalLabel = computed(() => {
  const interval = currentPlan.value?.interval
  if (interval === 'monthly') return 'mois'
  if (interval === 'yearly') return 'an'
  return 'mois'
})

const planBadgeClass = computed(() => {
  const code = currentPlan.value?.code
  if (code === 'free') return 'badge-free'
  if (code === 'pro') return 'badge-pro'
  if (code === 'enterprise') return 'badge-enterprise'
  return 'badge-free'
})

const subscriptionStatusClass = computed(() => {
  const status = activeSubscription.value?.status
  if (status === 'active') return 'status-active'
  if (status === 'pending') return 'status-pending'
  if (status === 'terminated') return 'status-terminated'
  return 'status-unknown'
})

const subscriptionStatusLabel = computed(() => {
  const status = activeSubscription.value?.status
  const labels = {
    active: 'Actif',
    pending: 'En attente',
    terminated: 'Résilié',
    canceled: 'Annulé'
  }
  return labels[status] || status
})

// Methods
function formatPrice(cents, currency = 'EUR') {
  if (!cents && cents !== 0) return '-'
  return billingStore.formatAmount(cents, currency)
}

function formatDate(dateString) {
  if (!dateString) return '-'
  const date = new Date(dateString)
  return date.toLocaleDateString('fr-FR', {
    day: 'numeric',
    month: 'long',
    year: 'numeric'
  })
}

function formatPeriod(from, to) {
  if (!from || !to) return ''
  const fromDate = new Date(from)
  const toDate = new Date(to)
  return `${fromDate.toLocaleDateString('fr-FR', { day: 'numeric', month: 'short' })} - ${toDate.toLocaleDateString('fr-FR', { day: 'numeric', month: 'short' })}`
}

function getChargeAmount(code) {
  const charge = billingStore.currentUsage?.charges?.find(c => c.code === code)
  return charge?.amount_cents || 0
}

function getPaymentStatusClass(status) {
  const classes = {
    succeeded: 'status-success',
    pending: 'status-pending',
    failed: 'status-error'
  }
  return classes[status] || 'status-unknown'
}

function getPaymentStatusLabel(status) {
  const labels = {
    succeeded: 'Payée',
    pending: 'En attente',
    failed: 'Échec'
  }
  return labels[status] || status
}

async function payPendingInvoice() {
  paymentLoading.value = true
  try {
    await billingStore.payPendingInvoice()
  } finally {
    paymentLoading.value = false
  }
}

async function payInvoice(invoiceId) {
  await billingStore.payInvoice(invoiceId)
}

function refreshInvoices() {
  billingStore.fetchInvoices()
}

function showUpgradeOptions() {
  // TODO: Open upgrade modal or redirect to pricing page
  console.log('Show upgrade options')
}

// Lifecycle
onMounted(() => {
  billingStore.fetchCurrentPlan()
  billingStore.fetchInvoices()
})
</script>

<style scoped>
.billing-dashboard {
  max-width: 900px;
  margin: 0 auto;
  padding: 1.5rem;
}

/* Loading */
.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem;
  color: var(--secondary-text-color);
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid var(--border-color);
  border-top-color: var(--primary-color);
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 1rem;
}

.spinner-small {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255,255,255,0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  display: inline-block;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Error Banner */
.error-banner {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  background-color: rgba(239, 68, 68, 0.1);
  border: 1px solid var(--error-color);
  border-radius: 8px;
  margin-bottom: 1.5rem;
  color: var(--error-color);
}

.error-banner .btn-dismiss {
  margin-left: auto;
  background: none;
  border: none;
  color: var(--error-color);
  cursor: pointer;
  padding: 0.25rem;
}

/* Cards */
.billing-content {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.card {
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  overflow: hidden;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid var(--border-color);
  background: var(--hover-background-color);
}

.card-header h2 {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--main-text-color);
}

.card-body {
  padding: 1.5rem;
}

.card-footer {
  padding: 1rem 1.5rem;
  border-top: 1px solid var(--border-color);
  background: var(--hover-background-color);
}

/* Plan Card */
.plan-badge {
  padding: 0.25rem 0.75rem;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: 600;
  text-transform: uppercase;
}

.badge-free {
  background: var(--hover-background-color);
  color: var(--secondary-text-color);
}

.badge-pro {
  background: linear-gradient(135deg, var(--primary-color), #818cf8);
  color: white;
}

.badge-enterprise {
  background: linear-gradient(135deg, #f59e0b, #d97706);
  color: white;
}

.plan-details {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.plan-price {
  display: flex;
  align-items: baseline;
  gap: 0.25rem;
}

.price-amount {
  font-size: 2rem;
  font-weight: 700;
  color: var(--main-text-color);
}

.price-interval {
  color: var(--secondary-text-color);
}

.plan-features {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.feature {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: var(--main-text-color);
}

.feature-icon {
  font-size: 1.1rem;
}

.subscription-info {
  margin-top: 1.5rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--border-color);
}

.info-row {
  display: flex;
  justify-content: space-between;
  padding: 0.5rem 0;
}

.info-row .label {
  color: var(--secondary-text-color);
}

/* Status badges */
.status-badge {
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  font-size: 0.8rem;
  font-weight: 500;
}

.status-active, .status-success {
  background: rgba(34, 197, 94, 0.15);
  color: var(--success-color);
}

.status-pending {
  background: rgba(245, 158, 11, 0.15);
  color: var(--warning-color);
}

.status-terminated, .status-error {
  background: rgba(239, 68, 68, 0.15);
  color: var(--error-color);
}

/* Usage Card */
.usage-metrics {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1.5rem;
}

.metric {
  padding: 1rem;
  background: var(--hover-background-color);
  border-radius: 8px;
}

.metric-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
}

.metric-icon {
  font-size: 1.2rem;
}

.metric-name {
  color: var(--secondary-text-color);
  font-size: 0.9rem;
}

.metric-value {
  display: flex;
  align-items: baseline;
  gap: 0.25rem;
}

.metric-value .value {
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--main-text-color);
}

.metric-value .unit {
  color: var(--secondary-text-color);
}

.metric-cost {
  margin-top: 0.5rem;
  color: var(--secondary-text-color);
  font-size: 0.85rem;
}

.usage-total {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 1.5rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--border-color);
}

.total-label {
  color: var(--secondary-text-color);
  font-weight: 500;
}

.total-amount {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--primary-color);
}

.period-label {
  color: var(--secondary-text-color);
  font-size: 0.85rem;
}

/* Payment Alert */
.payment-alert-card {
  border-color: var(--warning-color);
  background: rgba(245, 158, 11, 0.05);
}

.payment-alert {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.alert-icon {
  font-size: 2rem;
}

.alert-content {
  flex: 1;
}

.alert-content h3 {
  margin: 0 0 0.25rem 0;
  font-size: 1rem;
  color: var(--main-text-color);
}

.alert-content p {
  margin: 0;
  color: var(--secondary-text-color);
  font-size: 0.9rem;
}

/* Invoices Table */
.invoices-table {
  width: 100%;
  border-collapse: collapse;
}

.invoices-table th,
.invoices-table td {
  padding: 0.75rem 1rem;
  text-align: left;
  border-bottom: 1px solid var(--border-color);
}

.invoices-table th {
  color: var(--secondary-text-color);
  font-weight: 500;
  font-size: 0.85rem;
}

.invoices-table td {
  color: var(--main-text-color);
}

.invoices-table tr:last-child td {
  border-bottom: none;
}

/* Empty state */
.empty-state {
  text-align: center;
  padding: 2rem;
  color: var(--secondary-text-color);
}

.empty-icon {
  font-size: 2.5rem;
  display: block;
  margin-bottom: 0.5rem;
}

/* Buttons */
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.6rem 1.2rem;
  border-radius: 8px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  border: none;
}

.btn-primary {
  background: var(--primary-color);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: var(--accent-color);
}

.btn-secondary {
  background: var(--hover-background-color);
  color: var(--main-text-color);
  border: 1px solid var(--border-color);
}

.btn-secondary:hover:not(:disabled) {
  background: var(--border-color);
}

.btn-sm {
  padding: 0.4rem 0.8rem;
  font-size: 0.85rem;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-upgrade {
  width: 100%;
}

.btn-icon {
  font-size: 1.1rem;
}

/* Utility classes */
.text-success {
  color: var(--success-color);
}

.text-warning {
  color: var(--warning-color);
}
</style>
