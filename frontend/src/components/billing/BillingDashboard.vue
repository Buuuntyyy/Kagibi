<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="account-page">
    <div class="page-header">
      <div class="header-content">
        <button class="btn-back" @click="router.push('/dashboard')">
            <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M19 12H5M12 19l-7-7 7-7"/>
            </svg>
            Retour
        </button>
        <h1>Facturation</h1>
      </div>
      <p class="subtitle">Gérez votre abonnement et consultez votre historique de facturation.</p>
    </div>

    <!-- Loading State -->
    <div v-if="billingStore.loading" class="loading-state">
      <div class="spinner"></div>
      <p>Chargement des informations...</p>
    </div>

    <div v-else class="content-grid">

      <!-- Plan Section -->
      <section class="settings-section plan-section">
        <div class="section-header">
          <h3>Votre Abonnement</h3>
          <span :class="['plan-badge', planBadgeClass]">{{ currentPlan?.name || 'Gratuit' }}</span>
        </div>
        <div class="section-body">
            <div class="plan-details-grid">
                <div class="plan-info">
                    <div class="price-display">
                        <span class="amount">{{ formatPrice(currentPlan?.amount_cents) }}</span>
                        <span class="period" v-if="currentPlan?.amount_cents > 0">/ {{ intervalLabel }}</span>
                    </div>
                </div>
                <div class="plan-actions" v-if="currentPlan?.code === 'free'">
                     <button class="btn-primary" @click="showUpgradeOptions">Mettre à niveau</button>
                </div>
            </div>

            <div class="features-list">
                 <div class="feature-item" v-if="currentPlan?.code === 'free'">
                    <svg class="check-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <polyline points="20 6 9 17 4 12"></polyline>
                    </svg>
                    <span>20 Go de stockage</span>
                 </div>
                 <div class="feature-item" v-if="currentPlan?.code === 'pro'">
                    <svg class="check-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <polyline points="20 6 9 17 4 12"></polyline>
                    </svg>
                    <span>100 Go de stockage</span>
                 </div>
                 <div class="feature-item">
                    <svg class="check-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <polyline points="20 6 9 17 4 12"></polyline>
                    </svg>
                    <span>Chiffrement de bout en bout</span>
                 </div>
            </div>
        </div>
      </section>

      <!-- Usage Section -->
      <section class="settings-section">
          <div class="section-header">
              <h3>Consommation</h3>
              <span class="period-label" v-if="billingStore.currentUsage">
                {{ formatPeriod(billingStore.currentUsage.from_datetime, billingStore.currentUsage.to_datetime) }}
              </span>
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
                          <span class="stat-label">Stockage</span>
                          <span class="stat-value">{{ billingStore.storageUsageGB.toFixed(2) }} <small>Go</small></span>
                      </div>
                  </div>
                  <div class="stat-card">
                       <div class="stat-icon">
                          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                              <polyline points="16 16 12 12 8 16"></polyline>
                              <line x1="12" y1="12" x2="12" y2="21"></line>
                              <path d="M20.39 18.39A5 5 0 0 0 18 9h-1.26A8 8 0 1 0 3 16.3"></path>
                              <polyline points="16 16 12 12 8 16"></polyline>
                          </svg>
                      </div>
                      <div class="stat-info">
                          <span class="stat-label">Transfert P2P</span>
                          <span class="stat-value">{{ billingStore.p2pUsageMB.toFixed(0) }} <small>Mo</small></span>
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
                  <h4>Paiement en attente</h4>
                  <p>Facture {{ billingStore.pendingInvoice.number }} : {{ formatPrice(billingStore.pendingInvoice.total_amount_cents, billingStore.pendingInvoice.currency) }}</p>
              </div>
              <button class="btn-primary" @click="payPendingInvoice" :disabled="paymentLoading">
                  {{ paymentLoading ? '...' : 'Payer' }}
              </button>
          </div>
      </section>

      <!-- Invoices Section -->
      <section class="settings-section">
        <div class="section-header">
            <h3>Historique</h3>
        </div>
        <div class="section-body no-padding">
            <div v-if="billingStore.invoices.length === 0" class="empty-state">
                <p>Aucune facture disponible.</p>
            </div>
            <table v-else class="data-table">
                <thead>
                    <tr>
                        <th>Date</th>
                        <th>Numéro</th>
                        <th>Montant</th>
                        <th>Statut</th>
                        <th>Action</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="invoice in billingStore.invoices" :key="invoice.id">
                        <td>{{ formatDate(invoice.issuing_date) }}</td>
                        <td class="mono">{{ invoice.invoice_number }}</td>
                        <td class="font-medium">{{ formatPrice(invoice.total_amount_cents, invoice.currency) }}</td>
                        <td>
                            <span :class="['status-pill', getPaymentStatusClass(invoice.payment_status)]">
                                {{ getPaymentStatusLabel(invoice.payment_status) }}
                            </span>
                        </td>
                        <td>
                             <button
                                v-if="invoice.payment_status !== 'succeeded' && invoice.payment_link_url"
                                class="btn-sm btn-outline"
                                @click="payInvoice(invoice.lago_invoice_id)"
                              >
                                Payer
                              </button>
                              <a v-if="invoice.invoice_pdf_url" :href="invoice.invoice_pdf_url" target="_blank" class="download-link" title="Télécharger">
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
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useBillingStore } from '../../stores/billing'
import { useRouter } from 'vue-router'

const router = useRouter()
const billingStore = useBillingStore()
const paymentLoading = ref(false)

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
  if (code === 'pro') return 'badge-pro'
  if (code === 'enterprise') return 'badge-enterprise'
  return 'badge-free'
})

// Formatting helpers
function formatPrice(cents, currency = 'EUR') {
  if (!cents && cents !== 0) return '-'
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

function formatPeriod(from, to) {
  if (!from || !to) return ''
  const fromDate = new Date(from)
  const toDate = new Date(to)
  return `${fromDate.toLocaleDateString('fr-FR', { day: 'numeric', month: 'short' })} au ${toDate.toLocaleDateString('fr-FR', { day: 'numeric', month: 'short' })}`
}

function getPaymentStatusClass(status) {
  const classes = { succeeded: 'success', pending: 'warning', failed: 'error' }
  return classes[status] || 'unknown'
}

function getPaymentStatusLabel(status) {
  const labels = { succeeded: 'Payée', pending: 'En attente', failed: 'Échec' }
  return labels[status] || status
}

// Actions
async function payPendingInvoice() {
  paymentLoading.value = true
  try { await billingStore.payPendingInvoice() } 
  finally { paymentLoading.value = false }
}

async function payInvoice(invoiceId) {
  await billingStore.payInvoice(invoiceId)
}

function showUpgradeOptions() {
  // Logic to show upgrade modal
  console.log("Upgrade clicked")
}

onMounted(() => {
  billingStore.fetchCurrentPlan()
  billingStore.fetchInvoices()
})
</script>

<style scoped>
/* Page Layout (Matching Account.vue) */
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

.page-header {
  margin-bottom: 40px;
}

.header-content {
    display: flex;
    align-items: center;
    gap: 16px;
    margin-bottom: 8px;
}

.btn-back {
    background: none;
    border: none;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    color: var(--secondary-text-color);
    cursor: pointer;
    font-size: 0.9rem;
    padding: 6px 12px;
    border-radius: 8px;
    transition: all 0.2s;
}

.btn-back:hover {
    background-color: var(--hover-background-color);
    color: var(--primary-color);
}

.page-header h1 {
  font-size: 2rem;
  font-weight: 700;
  color: var(--main-text-color);
  margin: 0;
}

.subtitle {
  color: var(--secondary-text-color);
  margin: 0;
  font-size: 1.1rem;
}

/* Loading */
.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 400px;
  color: var(--secondary-text-color);
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid var(--border-color);
  border-top-color: var(--primary-color);
  border-radius: 50%;
  animation: spin 1s infinite linear;
  margin: 0 auto 16px;
}

@keyframes spin { to { transform: rotate(360deg); } }

/* Grid & Cards */
.content-grid {
  display: grid;
  gap: 24px;
}

.settings-section {
  background: var(--card-color);
  border-radius: 12px;
  border: 1px solid var(--border-color);
  overflow: hidden;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.05); /* Subtle shadow like Account */
}

.warning-section {
    border-color: var(--warning-color);
    background: #fffbf0; /* Very light yellow */
}
[data-theme="dark"] .warning-section {
    background: rgba(245, 158, 11, 0.1);
}

.section-header {
  padding: 20px 24px;
  border-bottom: 1px solid var(--border-color);
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: var(--hover-background-color); /* Light header bg found in Account cards ? or just clean */
  background: transparent; /* Account.vue seems to have transparent headers in sections inside cards? No, account.vue has sections. */
}

.section-header h3 {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--main-text-color);
}

.section-body {
  padding: 24px;
}

.section-body.no-padding {
    padding: 0;
}

/* Plan Badge */
.plan-badge {
  padding: 6px 12px;
  border-radius: 20px;
  font-size: 0.85rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
.badge-free { background: var(--hover-background-color); color: var(--secondary-text-color); }
.badge-pro { background: var(--primary-color); color: white; }
.badge-enterprise { background: linear-gradient(135deg, #f59e0b, #d97706); color: white; }

/* Plan Section Details */
.plan-details-grid {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 24px;
}

.price-display {
    display: flex;
    align-items: baseline;
}
.amount { font-size: 2.5rem; font-weight: 700; color: var(--main-text-color); }
.period { font-size: 1rem; color: var(--secondary-text-color); margin-left: 6px; }

.features-list {
    display: grid;
    gap: 12px;
    padding-top: 16px;
    border-top: 1px solid var(--border-color);
}

.feature-item {
    display: flex;
    align-items: center;
    gap: 12px;
    color: var(--main-text-color);
}

.check-icon {
    width: 20px;
    height: 20px;
    color: var(--success-color);
}

/* Stats */
.usage-stats {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 20px;
}

.stat-card {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 16px;
    background: var(--hover-background-color);
    border-radius: 10px;
}

.stat-icon {
    width: 40px;
    height: 40px;
    border-radius: 8px;
    background: var(--card-color);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--primary-color);
}
.stat-icon svg { width: 20px; height: 20px; }

.stat-info { display: flex; flex-direction: column; }
.stat-label { font-size: 0.85rem; color: var(--secondary-text-color); }
.stat-value { font-size: 1.25rem; font-weight: 600; color: var(--main-text-color); }
.stat-value small { font-size: 0.85rem; color: var(--secondary-text-color); }

/* Buttons */
.btn-primary {
  background: var(--primary-color);
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-primary:hover:not(:disabled) {
  background: var(--accent-color);
  transform: translateY(-1px);
}
.btn-primary:disabled { opacity: 0.7; cursor: not-allowed; }

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

.download-link { color: var(--secondary-text-color); margin-left: 10px; opacity: 0.7; transition: 0.2s; }
.download-link:hover { opacity: 1; color: var(--primary-color); }

.alert-body { display: flex; align-items: center; gap: 16px; }
.alert-icon-wrapper { color: var(--warning-color); }
.alert-icon-wrapper svg { width: 24px; height: 24px; }
.alert-content { flex: 1; }
.alert-content h4 { margin: 0 0 4px 0; color: var(--main-text-color); }
.alert-content p { margin: 0; color: var(--secondary-text-color); font-size: 0.9rem; }

.empty-state { text-align: center; padding: 40px; color: var(--secondary-text-color); font-style: italic; }
</style>
