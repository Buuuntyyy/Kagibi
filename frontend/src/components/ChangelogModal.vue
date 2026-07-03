<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <transition name="modal-fade">
    <div v-if="visible" class="changelog-overlay" @click.self="dismiss">
      <div class="changelog-box">
        <div class="changelog-header">
          <div class="changelog-title-group">
            <span class="version-badge">v{{ CURRENT_VERSION }}</span>
            <h3>{{ t('changelog.title') }}</h3>
          </div>
          <button class="close-btn" @click="dismiss" :aria-label="t('common.close')">✕</button>
        </div>

        <div class="changelog-content">
          <div v-for="section in sections" :key="section.title" class="changelog-section">
            <h4 class="section-title">{{ section.title }}</h4>
            <ul>
              <li v-for="item in section.items" :key="item.label">
                <strong>{{ item.label }}</strong> — {{ item.text }}
              </li>
            </ul>
          </div>
        </div>

        <div class="changelog-footer">
          <button class="confirm-btn" @click="dismiss">{{ t('changelog.confirm') }}</button>
        </div>
      </div>
    </div>
  </transition>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'

const { t, locale } = useI18n()

const CURRENT_VERSION = '2.26.0'
const STORAGE_KEY = 'kagibi_changelog_seen_v'

const visible = ref(false)

const sectionsFr = [
  {
    title: '🚀 Nouvelles fonctionnalités',
    items: [
      {
        label: 'Demandes d\'accès',
        text: 'Les membres peuvent demander l\'accès à un dossier d\'organisation. Les admins reçoivent et traitent ces demandes depuis le panneau d\'administration.',
      },
      {
        label: 'Vue d\'accès effectif',
        text: 'Nouvelle vue synthétique affichant les permissions réelles d\'un utilisateur sur un dossier (droits directs + groupes).',
      },
      {
        label: 'Gestion du chiffrement par groupe',
        text: 'Les admins d\'organisation peuvent gérer la clé chiffrée par groupe — attribution et révocation sécurisée.',
      },
      {
        label: 'Héritage des permissions de dossier',
        text: 'Les permissions définies pour un groupe sur un dossier s\'héritent désormais correctement dans les sous-dossiers.',
      },
    ],
  },
]

const sectionsEn = [
  {
    title: '🚀 New Features',
    items: [
      {
        label: 'Access Requests',
        text: 'Members can now request access to an organization folder. Admins receive and handle these requests from the admin panel.',
      },
      {
        label: 'Effective Access View',
        text: 'New summary view showing a user\'s effective permissions on a folder, combining direct rights and group memberships.',
      },
      {
        label: 'Group Encryption Management',
        text: 'Organization admins can manage the encrypted key per group — granting and securely revoking encrypted access.',
      },
      {
        label: 'Folder Permission Inheritance',
        text: 'Permissions assigned to a group on a folder are now correctly inherited by all subfolders.',
      },
    ],
  },
]

const sections = computed(() => (locale.value === 'fr' ? sectionsFr : sectionsEn))

const dismiss = () => {
  visible.value = false
  localStorage.setItem(STORAGE_KEY, CURRENT_VERSION)
}

onMounted(() => {
  const seen = localStorage.getItem(STORAGE_KEY)
  if (seen !== CURRENT_VERSION) {
    visible.value = true
  }
})
</script>

<style scoped>
.changelog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
}

.changelog-box {
  background: var(--card-color, #1e1e1e);
  width: 100%;
  max-width: 540px;
  border-radius: 14px;
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.4);
  border: 1px solid var(--border-color, #333);
  overflow: hidden;
  animation: popIn 0.3s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}

.changelog-header {
  padding: 1rem 1.25rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  background: rgba(255, 255, 255, 0.02);
}

.changelog-title-group {
  display: flex;
  align-items: center;
  gap: 0.6rem;
}

.version-badge {
  background: var(--primary-color, #667eea);
  color: #fff;
  font-size: 0.7rem;
  font-weight: 700;
  padding: 0.2rem 0.5rem;
  border-radius: 20px;
  letter-spacing: 0.03em;
}

.changelog-header h3 {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-color, #fff);
}

.close-btn {
  background: none;
  border: none;
  color: var(--secondary-text-color, #aaa);
  font-size: 1rem;
  cursor: pointer;
  padding: 0.25rem 0.4rem;
  border-radius: 6px;
  line-height: 1;
  transition: background 0.15s, color 0.15s;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.08);
  color: var(--text-color, #fff);
}

.changelog-content {
  padding: 1.25rem;
  max-height: 320px;
  overflow-y: auto;
}

.changelog-section {
  margin-bottom: 1rem;
}

.changelog-section:last-child {
  margin-bottom: 0;
}

.section-title {
  margin: 0 0 0.75rem 0;
  font-size: 0.85rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--primary-color, #667eea);
}

.changelog-section ul {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}

.changelog-section li {
  font-size: 0.875rem;
  line-height: 1.55;
  color: var(--secondary-text-color, #ccc);
  padding-left: 0.75rem;
  border-left: 2px solid var(--border-color, #333);
}

.changelog-section li strong {
  color: var(--text-color, #fff);
  font-weight: 600;
}

.changelog-footer {
  padding: 1rem 1.25rem;
  display: flex;
  justify-content: flex-end;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
  background: rgba(0, 0, 0, 0.08);
}

.confirm-btn {
  padding: 0.55rem 1.5rem;
  border: none;
  border-radius: 8px;
  background: var(--primary-color, #667eea);
  color: #fff;
  font-size: 0.875rem;
  font-weight: 600;
  cursor: pointer;
  transition: transform 0.1s, filter 0.1s;
}

.confirm-btn:hover {
  filter: brightness(1.1);
  transform: scale(1.02);
}

.confirm-btn:active {
  transform: scale(0.97);
}

@keyframes popIn {
  from {
    opacity: 0;
    transform: scale(0.92) translateY(12px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: opacity 0.2s ease;
}

.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}

@media (max-width: 600px) {
  .changelog-box {
    margin: 1rem;
    max-width: 100%;
  }
}
</style>
