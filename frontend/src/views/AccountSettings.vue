<template>
  <div class="settings-page">
    <div class="settings-container">
      <h1>Paramètres du Compte</h1>

      <!-- Informations utilisateur -->
      <section class="settings-section">
        <h2>Informations</h2>
        <div class="info-row">
          <span class="label">Email :</span>
          <span class="value">{{ authStore.user?.email || '-' }}</span>
        </div>
        <div class="info-row">
          <span class="label">Nom :</span>
          <span class="value">{{ authStore.user?.name || '-' }}</span>
        </div>
        <div class="info-row">
          <span class="label">Stockage utilisé :</span>
          <span class="value">{{ formatStorageUsed() }}</span>
        </div>
      </section>

      <!-- Zone Dangereuse -->
      <section class="danger-zone">
        <h2>⚠️ Zone Dangereuse</h2>
        <p class="warning-text">
          La suppression de votre compte est <strong>IMMÉDIATE</strong> et <strong>DÉFINITIVEMENT IRRÉVERSIBLE</strong>.
        </p>
        <p class="warning-text critical">
          <strong>🚨 ATTENTION IRRÉVERSIBLE :</strong>
        </p>
        <ul class="warning-list">
          <li>❌ Toutes vos données seront <strong>SUPPRIMÉES IMMÉDIATEMENT</strong></li>
          <li>❌ Vos fichiers seront <strong>DÉFINITIVEMENT EFFACÉS</strong></li>
          <li>❌ Vos clés de chiffrement seront <strong>DÉTRUITES</strong></li>
          <li>❌ <strong>AUCUNE RÉCUPÉRATION POSSIBLE</strong> après validation</li>
          <li>❌ <strong>AUCUN DÉLAI DE GRÂCE</strong> : suppression immédiate</li>
        </ul>
        <p class="warning-text critical">
          🚨 Conformément au RGPD (Article 17), cette suppression est <strong>DÉFINITIVE</strong>.
        </p>

        <button @click="showDeleteModal = true" class="btn-danger">
          🗑️ Supprimer Mon Compte
        </button>
      </section>
    </div>

    <!-- Modal de confirmation -->
    <div v-if="showDeleteModal" class="modal-overlay" @click.self="closeModal">
      <div class="modal-card">
        <div class="modal-header">
          <svg class="warning-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/>
            <line x1="12" y1="9" x2="12" y2="13"/>
            <line x1="12" y1="17" x2="12.01" y2="17"/>
          </svg>
          <h3>Confirmer la Suppression</h3>
        </div>

        <div class="modal-body">
          <p><strong>🚨 CETTE ACTION EST IMMÉDIATE ET DÉFINITIVEMENT IRRÉVERSIBLE !</strong></p>
          <p><strong>En cliquant sur "Supprimer Définitivement" :</strong></p>
          <ul>
            <li>❌ Votre compte sera <strong>SUPPRIMÉ IMMÉDIATEMENT</strong></li>
            <li>❌ Vos fichiers seront <strong>EFFACÉS IMMÉDIATEMENT</strong></li>
            <li>❌ Vos clés de chiffrement seront <strong>DÉTRUITES IMMÉDIATEMENT</strong></li>
            <li>❌ Vos partages seront <strong>RÉVOQUÉS IMMÉDIATEMENT</strong></li>
            <li>❌ Toutes vos données seront <strong>IRRÉCUPÉRABLES</strong></li>
          </ul>
          <div class="critical-warning">
            <p><strong>🚨 AUCUNE RÉCUPÉRATION POSSIBLE</strong></p>
            <p>Il n'y a <strong>AUCUN DÉLAI DE GRÂCE</strong>. La suppression est <strong>INSTANTANÉE</strong> et <strong>DÉFINITIVE</strong>.</p>
          </div>

          <div class="confirmation-section">
            <label for="confirmInput">
              Tapez <code>SUPPRIMER DEFINITIVEMENT</code> pour confirmer :
            </label>
            <input
              id="confirmInput"
              v-model="confirmationText"
              type="text"
              placeholder="SUPPRIMER DEFINITIVEMENT"
              class="confirmation-input"
              @keyup.enter="handleDelete"
              autocomplete="off"
            />
          </div>
        </div>

        <div class="modal-footer">
          <button @click="closeModal" class="btn-secondary" :disabled="isDeleting">
            Annuler
          </button>
          <button
            @click="handleDelete"
            class="btn-danger"
            :disabled="confirmationText !== 'SUPPRIMER DEFINITIVEMENT' || isDeleting"
          >
            <span v-if="isDeleting">Suppression IRRÉVERSIBLE...</span>
            <span v-else>⚠️ Supprimer DÉFINITIVEMENT (IRRÉVERSIBLE)</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()

const showDeleteModal = ref(false)
const confirmationText = ref('')
const isDeleting = ref(false)

const formatStorageUsed = () => {
  const used = authStore.user?.storage_used || 0
  const limit = authStore.user?.storage_limit || 5368709120 // 5GB default

  const formatBytes = (bytes) => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  return `${formatBytes(used)} / ${formatBytes(limit)}`
}

const closeModal = () => {
  if (!isDeleting.value) {
    showDeleteModal.value = false
    confirmationText.value = ''
  }
}

const handleDelete = async () => {
  if (confirmationText.value !== 'SUPPRIMER DEFINITIVEMENT') {
    alert('⚠️ Veuillez taper "SUPPRIMER DEFINITIVEMENT" pour confirmer la suppression irréversible');
    return
  }

  // Confirmation supplémentaire pour éviter les suppressions accidentelles
  const finalConfirm = confirm(
    '🚨 DERNIÈRE CONFIRMATION 🚨\n\n' +
    'Votre compte et TOUTES vos données seront SUPPRIMÉS IMMÉDIATEMENT.\n\n' +
    'Cette action est DÉFINITIVEMENT IRRÉVERSIBLE.\n\n' +
    'AUCUNE RÉCUPÉRATION ne sera possible.\n\n' +
    'Êtes-vous ABSOLUMENT certain(e) ?'
  )

  if (!finalConfirm) {
    return
  }

  isDeleting.value = true

  try {
    await authStore.deleteAccount('SUPPRIMER')

    // Redirection vers la page d'accueil
    router.push('/')

    // Notification de succès
    alert(
      '✅ Votre compte a été DÉFINITIVEMENT supprimé.\n\n' +
      '❌ Toutes vos données sont IRRÉCUPÉRABLES.\n\n' +
      'Conformément au RGPD (Article 17), vos données personnelles ont été effacées.'
    )

  } catch (error) {
    alert('❌ Erreur lors de la suppression : ' + error.message)
  } finally {
    isDeleting.value = false
    closeModal()
  }
}
</script>

<style scoped>
.settings-page {
  min-height: 100vh;
  background-color: var(--background-color);
  padding: 2rem;
}

.settings-container {
  max-width: 800px;
  margin: 0 auto;
}

h1 {
  margin-bottom: 2rem;
  color: var(--text-color);
}

.settings-section {
  background-color: var(--card-color);
  border-radius: 8px;
  padding: 1.5rem;
  margin-bottom: 2rem;
}

.settings-section h2 {
  margin-bottom: 1rem;
  color: var(--text-color);
  font-size: 1.25rem;
}

.info-row {
  display: flex;
  padding: 0.75rem 0;
  border-bottom: 1px solid var(--border-color);
}

.info-row:last-child {
  border-bottom: none;
}

.label {
  flex: 0 0 150px;
  color: var(--text-secondary);
  font-weight: 500;
}

.value {
  color: var(--text-color);
}

.danger-zone {
  background-color: rgba(220, 53, 69, 0.05);
  border: 2px solid #dc3545;
  border-radius: 8px;
  padding: 2rem;
}

.danger-zone h2 {
  color: #dc3545;
  margin-bottom: 1rem;
}

.warning-text {
  color: var(--text-secondary);
  margin-bottom: 1rem;
  line-height: 1.6;
}

.warning-text.critical {
  color: #742a2a;
  font-weight: 600;
  background: rgba(220, 53, 69, 0.1);
  padding: 0.75rem;
  border-radius: 6px;
  border-left: 4px solid #dc3545;
}

.warning-list {
  list-style: none;
  padding-left: 0;
  margin: 1rem 0;
}

.warning-list li {
  padding: 0.5rem 0;
  color: #742a2a;
  font-weight: 500;
  line-height: 1.5;
}

.critical-warning {
  background: #fff5f5;
  border: 2px solid #dc3545;
  border-radius: 8px;
  padding: 1rem;
  margin: 1.5rem 0;
}

.critical-warning p {
  color: #c53030;
  font-weight: 600;
  margin: 0.5rem 0;
}

.btn-danger {
  background-color: #dc3545;
  color: white;
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 600;
  font-size: 1rem;
  transition: background-color 0.2s, opacity 0.2s;
}

.btn-danger:hover:not(:disabled) {
  background-color: #c82333;
}

.btn-danger:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-secondary {
  background-color: #6c757d;
  color: white;
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  margin-right: 1rem;
  transition: background-color 0.2s;
}

.btn-secondary:hover:not(:disabled) {
  background-color: #5a6268;
}

.btn-secondary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Modal */
.modal-overlay {
  position: fixed;
  inset: 0;
  background-color: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.modal-card {
  background-color: var(--card-color);
  border-radius: 12px;
  max-width: 500px;
  width: 100%;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  animation: modalSlideIn 0.2s ease-out;
}

@keyframes modalSlideIn {
  from {
    opacity: 0;
    transform: translateY(-20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.modal-header {
  padding: 1.5rem 2rem;
  border-bottom: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  gap: 1rem;
}

.warning-icon {
  width: 48px;
  height: 48px;
  color: #dc3545;
  flex-shrink: 0;
}

.modal-header h3 {
  margin: 0;
  color: var(--text-color);
}

.modal-body {
  padding: 1.5rem 2rem;
}

.modal-body p {
  margin-bottom: 1rem;
  color: var(--text-color);
}

.modal-body ul {
  margin: 1rem 0;
  padding-left: 1.5rem;
  color: var(--text-secondary);
}

.modal-body li {
  margin-bottom: 0.5rem;
  line-height: 1.5;
}

.confirmation-section {
  margin-top: 1.5rem;
}

.confirmation-section label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 600;
  color: var(--text-color);
}

.confirmation-section code {
  background-color: rgba(220, 53, 69, 0.2);
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  color: #c53030;
  font-weight: bold;
  font-size: 0.875rem;
}

.confirmation-input {
  width: 100%;
  padding: 0.75rem;
  border: 2px solid var(--border-color);
  border-radius: 6px;
  font-size: 1rem;
  background-color: var(--background-color);
  color: var(--text-color);
  transition: border-color 0.2s;
}

.confirmation-input:focus {
  outline: none;
  border-color: #dc3545;
}

.modal-footer {
  padding: 1.5rem 2rem;
  border-top: 1px solid var(--border-color);
  display: flex;
  justify-content: flex-end;
}

/* Responsive */
@media (max-width: 600px) {
  .settings-page {
    padding: 1rem;
  }

  .info-row {
    flex-direction: column;
    gap: 0.25rem;
  }

  .label {
    flex: none;
  }

  .modal-footer {
    flex-direction: column-reverse;
    gap: 0.75rem;
  }

  .btn-secondary {
    margin-right: 0;
    width: 100%;
  }

  .btn-danger {
    width: 100%;
  }
}
</style>
