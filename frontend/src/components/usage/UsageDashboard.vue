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
        <h1>Mon Utilisation</h1>
      </div>
      <p class="subtitle">Consultez en temps réel l'utilisation de votre espace et vos transferts P2P.</p>
    </div>

    <!-- Loading State -->
    <div v-if="billingStore.loading && !billingStore.currentPlan" class="loading-state">
      <div class="spinner"></div>
      <p>Chargement des statistiques...</p>
    </div>

    <div v-else class="content-grid">

      <!-- Usage Section -->
      <section class="settings-section">
          <div class="section-header">
              <h3>Statistiques d'utilisation</h3>
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
                          <span class="stat-label">Stockage Utilisé</span>
                          <span class="stat-value">{{ billingStore.storageUsageGB.toFixed(2) }} <small>Go / {{ maxStorageGB }} Go</small></span>
                      </div>
                      <div class="stat-bar">
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
                          <span class="stat-label">Bande passante consommée</span>
                          <span class="stat-value">{{ billingStore.bandwidthUsageGB.toFixed(2) }} <small>Go</small></span>
                      </div>
                  </div>

                  <div class="stat-card">
                       <div class="stat-icon">
                          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                              <polyline points="16 16 12 12 8 16"></polyline>
                              <line x1="12" y1="12" x2="12" y2="21"></line>
                              <path d="M20.39 18.39A5 5 0 0 0 18 9h-1.26A8 8 0 1 0 3 16.3"></path>
                          </svg>
                      </div>
                      <div class="stat-info">
                          <span class="stat-label">Taille max par fichier (+ Partages)</span>
                          <span class="stat-value">{{ maxFileSizeText }} </span>
                      </div>
                  </div>
              </div>
          </div>
      </section>

    </div>
  </div>
</template>

<script setup>
import { computed, onMounted } from 'vue'
import { useBillingStore } from '../../stores/billing'
import { useRouter } from 'vue-router'

const router = useRouter()
const billingStore = useBillingStore()

const currentPlan = computed(() => billingStore.currentPlan)

const maxStorageGB = computed(() => {
  return currentPlan.value?.storage_limit_gb || Number(import.meta.env.VITE_DEFAULT_STORAGE_GB) || 20;
})

const maxFileSizeText = computed(() => {
   const val = currentPlan.value?.features?.max_file_size_mb || 1024;
   if(val >= 1024) return (val/1024) + ' Go';
   return val + ' Mo';
})

const storagePercent = computed(() => {
  const limit = maxStorageGB.value
  const used = billingStore.storageUsageGB
  return Math.min((used / limit) * 100, 100)
})

onMounted(async () => {
  // Fetch billing data in parallel
  await Promise.all([
    billingStore.fetchBillingStatus(),
    billingStore.fetchCurrentPlan(),
    billingStore.fetchUsage()
  ])
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

.section-header {
  padding: 20px 24px; border-bottom: 1px solid var(--border-color);
  display: flex; justify-content: space-between; align-items: center;
}
.section-header h3 { margin: 0; font-size: 1.1rem; font-weight: 600; color: var(--main-text-color); }
.section-body { padding: 24px; }

/* Stats */
.usage-stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 20px; }
.stat-card { display: flex; flex-wrap: wrap; align-items: stretch; gap: 16px; padding: 16px; background: var(--hover-background-color); border-radius: 10px; }
.stat-icon { width: 40px; height: 40px; border-radius: 8px; background: var(--card-color); display: flex; align-items: center; justify-content: center; color: var(--primary-color); }
.stat-icon svg { width: 20px; height: 20px; }
.stat-info { display: flex; flex-direction: column; flex: 1; min-width: 150px; justify-content: center;}
.stat-label { font-size: 0.85rem; color: var(--secondary-text-color); margin-bottom: 2px; }
.stat-value { font-size: 1.25rem; font-weight: 600; color: var(--main-text-color); }
.stat-value small { font-size: 0.85rem; color: var(--secondary-text-color); font-weight: normal; margin-left: 5px; }

/* Usage Progress Bar */
.stat-bar { width: 100%; flex: 1 1 100%; height: 6px; background: var(--border-color); border-radius: 3px; overflow: hidden; margin-top: 8px; }
.stat-bar-fill { height: 100%; background: var(--primary-color); border-radius: 3px; transition: width 0.5s ease; }
.stat-bar-fill.bar-warning { background: var(--warning-color); }
.stat-bar-fill.bar-danger { background: var(--error-color); }

@media (max-width: 768px) {
  .account-page {
    padding: 1.5rem 1rem;
  }

  .page-header h1 {
    font-size: 1.5rem;
  }

  .usage-stats {
    grid-template-columns: 1fr;
  }

  .section-body {
    padding: 16px;
  }

  .section-header {
    padding: 14px 16px;
  }
}

@media (max-width: 480px) {
  .account-page {
    padding: 1rem 0.75rem;
  }

  .stat-card {
    flex-direction: column;
  }
}

</style>
