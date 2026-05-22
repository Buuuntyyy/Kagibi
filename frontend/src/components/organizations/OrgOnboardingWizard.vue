<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <Transition name="wizard">
    <div class="wizard-overlay" @click.self="$emit('close')">
      <div class="wizard-modal" role="dialog" aria-modal="true">

        <button class="wizard-close" @click="$emit('close')" :aria-label="t('common.close')">
          <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
        </button>

        <!-- Step progress bar -->
        <div class="steps-bar" aria-hidden="true">
          <template v-for="i in STEPS" :key="i">
            <div class="step-dot" :class="{ active: step === i, done: step > i }">
              <svg v-if="step > i" viewBox="0 0 24 24" width="10" height="10" fill="currentColor"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41L9 16.17z"/></svg>
            </div>
            <div v-if="i < STEPS" class="step-line" :class="{ filled: step > i }"></div>
          </template>
        </div>

        <!-- Step content -->
        <Transition name="slide" mode="out-in">

          <!-- 1 · Welcome -->
          <div v-if="step === 1" key="s1" class="wizard-body">
            <div class="org-avatar-lg">{{ orgName.charAt(0).toUpperCase() }}</div>
            <h2 class="wizard-title">{{ t('orgs.wizard.welcomeTitle', { name: orgName }) }}</h2>
            <p class="wizard-desc">{{ t('orgs.wizard.welcomeDesc') }}</p>
            <div class="upcoming-steps">
              <div class="upcoming-item">
                <span class="upcoming-num">1</span>
                <span>{{ t('orgs.wizard.step2Title') }}</span>
              </div>
              <div class="upcoming-item">
                <span class="upcoming-num">2</span>
                <span>{{ t('orgs.wizard.step3Title') }}</span>
              </div>
            </div>
          </div>

          <!-- 2 · Invite -->
          <div v-else-if="step === 2" key="s2" class="wizard-body">
            <div class="step-icon-circle">
              <svg viewBox="0 0 24 24" width="26" height="26" fill="currentColor">
                <path d="M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z"/>
              </svg>
            </div>
            <h2 class="wizard-title">{{ t('orgs.wizard.step2Title') }}</h2>
            <p class="wizard-desc">{{ t('orgs.wizard.step2Desc') }}</p>

            <div v-if="!inviteLink" class="step-action">
              <button class="btn-primary btn-full" @click="generateInvite" :disabled="generatingInvite">
                <span v-if="generatingInvite" class="spinner-sm"></span>
                <svg v-else viewBox="0 0 24 24" width="15" height="15" fill="currentColor"><path d="M13 7h-2v4H7v2h4v4h2v-4h4v-2h-4V7zm-1-5C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8z"/></svg>
                {{ t('orgs.wizard.generateInvite') }}
              </button>
              <p v-if="inviteError" class="form-error">{{ inviteError }}</p>
            </div>

            <div v-else class="invite-result">
              <span class="success-tag">
                <svg viewBox="0 0 24 24" width="12" height="12" fill="currentColor"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41L9 16.17z"/></svg>
                {{ t('orgs.wizard.inviteGenerated') }}
              </span>
              <div class="link-row">
                <input class="input-field link-input" :value="inviteLink" readonly />
                <button class="btn-copy" @click="copyInvite">
                  <svg v-if="!copied" viewBox="0 0 24 24" width="15" height="15" fill="currentColor"><path d="M16 1H4c-1.1 0-2 .9-2 2v14h2V3h12V1zm3 4H8c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h11c1.1 0 2-.9 2-2V7c0-1.1-.9-2-2-2zm0 16H8V7h11v14z"/></svg>
                  <svg v-else viewBox="0 0 24 24" width="15" height="15" fill="currentColor" style="color:var(--primary-color)"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41L9 16.17z"/></svg>
                </button>
              </div>
              <p class="invite-hint">{{ t('orgs.wizard.inviteHint') }}</p>
            </div>
          </div>

          <!-- 3 · First folder -->
          <div v-else-if="step === 3" key="s3" class="wizard-body">
            <div class="step-icon-circle">
              <svg viewBox="0 0 24 24" width="26" height="26" fill="currentColor">
                <path d="M20 6h-8l-2-2H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2z"/>
              </svg>
            </div>
            <h2 class="wizard-title">{{ t('orgs.wizard.step3Title') }}</h2>
            <p class="wizard-desc">{{ t('orgs.wizard.step3Desc') }}</p>

            <div v-if="!folderCreated">
              <div class="chip-row">
                <button
                  v-for="s in folderSuggestions"
                  :key="s"
                  class="chip"
                  :class="{ active: folderName === s }"
                  @click="folderName = s"
                >{{ s }}</button>
              </div>
              <div class="folder-row">
                <input
                  v-model="folderName"
                  class="input-field"
                  :placeholder="t('orgs.wizard.folderPlaceholder')"
                  @keyup.enter="folderName && createFirstFolder()"
                />
                <button class="btn-primary" @click="createFirstFolder" :disabled="!folderName || creatingFolder">
                  <span v-if="creatingFolder" class="spinner-sm"></span>
                  <span v-else>{{ t('orgs.wizard.createFolder') }}</span>
                </button>
              </div>
              <p v-if="folderError" class="form-error">{{ folderError }}</p>
            </div>

            <div v-else class="folder-done">
              <svg viewBox="0 0 24 24" width="34" height="34" fill="currentColor" class="folder-done-check">
                <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41L9 16.17z"/>
              </svg>
              <p>{{ t('orgs.wizard.folderCreatedMsg', { name: folderName }) }}</p>
            </div>
          </div>

          <!-- 4 · Done -->
          <div v-else-if="step === 4" key="s4" class="wizard-body wizard-body-done">
            <div class="checkmark-wrap">
              <svg viewBox="0 0 52 52" class="checkmark-svg" aria-hidden="true">
                <circle class="ck-circle" cx="26" cy="26" r="25" fill="none"/>
                <path class="ck-check" fill="none" d="M14.1 27.2l7.1 7.2 16.7-16.8"/>
              </svg>
            </div>
            <h2 class="wizard-title">{{ t('orgs.wizard.doneTitle') }}</h2>
            <p class="wizard-desc">{{ t('orgs.wizard.doneDesc') }}</p>
            <ul class="done-list">
              <li>
                <div class="check-dot done"><svg viewBox="0 0 24 24" width="11" height="11" fill="currentColor"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41L9 16.17z"/></svg></div>
                {{ t('orgs.wizard.checkKey') }}
              </li>
              <li>
                <div class="check-dot" :class="{ done: !!inviteLink }">
                  <svg v-if="inviteLink" viewBox="0 0 24 24" width="11" height="11" fill="currentColor"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41L9 16.17z"/></svg>
                </div>
                {{ t('orgs.wizard.checkInvite') }}
              </li>
              <li>
                <div class="check-dot" :class="{ done: folderCreated }">
                  <svg v-if="folderCreated" viewBox="0 0 24 24" width="11" height="11" fill="currentColor"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41L9 16.17z"/></svg>
                </div>
                {{ t('orgs.wizard.checkFolder') }}
              </li>
            </ul>
          </div>

        </Transition>

        <!-- Footer navigation -->
        <div class="wizard-footer">
          <button v-if="step === 2 || step === 3" class="btn-skip" @click="advanceStep">
            {{ t('orgs.wizard.skip') }}
          </button>
          <div v-else class="footer-spacer"></div>

          <div class="footer-right">
            <button v-if="step < 4" class="btn-primary" @click="advanceStep">
              {{ step === 3 ? t('orgs.wizard.finish') : t('orgs.wizard.next') }}
              <svg viewBox="0 0 24 24" width="15" height="15" fill="currentColor"><path d="M8.59 16.59L13.17 12 8.59 7.41 10 6l6 6-6 6-1.41-1.41z"/></svg>
            </button>
            <button v-else class="btn-primary btn-start" @click="$emit('close')">
              {{ t('orgs.wizard.startUsing') }}
            </button>
          </div>
        </div>

      </div>
    </div>
  </Transition>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useOrgStore } from '../../stores/organizations'

const props = defineProps({
  orgId:   { type: Number, required: true },
  orgName: { type: String, required: true },
})
defineEmits(['close'])

const { t } = useI18n()
const orgStore = useOrgStore()

const STEPS = 4
const step = ref(1)

// Step 2 state
const inviteLink     = ref('')
const generatingInvite = ref(false)
const inviteError    = ref('')
const copied         = ref(false)

// Step 3 state
const folderName    = ref('')
const creatingFolder = ref(false)
const folderCreated  = ref(false)
const folderError   = ref('')

const folderSuggestions = computed(() =>
  t('orgs.wizard.folderSuggestions').split(',')
)

function advanceStep() {
  if (step.value < STEPS) step.value++
}

async function generateInvite() {
  generatingInvite.value = true
  inviteError.value = ''
  try {
    const inv = await orgStore.createInvitation(props.orgId, { role: 'member', max_uses: 0 })
    inviteLink.value = `${window.location.origin}/join/${inv.token}`
  } catch (e) {
    inviteError.value = e.response?.data?.error || e.message
  } finally {
    generatingInvite.value = false
  }
}

async function copyInvite() {
  try {
    await navigator.clipboard.writeText(inviteLink.value)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch (_) { /* clipboard denied */ }
}

async function createFirstFolder() {
  if (!folderName.value || creatingFolder.value) return
  creatingFolder.value = true
  folderError.value = ''
  try {
    await orgStore.createFolder(props.orgId, folderName.value, '/')
    folderCreated.value = true
  } catch (e) {
    folderError.value = e.response?.data?.error || e.message
  } finally {
    creatingFolder.value = false
  }
}
</script>

<style scoped>
/* ── Overlay ── */
.wizard-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.55);
  backdrop-filter: blur(3px);
  z-index: 2100;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
}

/* ── Modal shell ── */
.wizard-modal {
  background: var(--card-color);
  border-radius: 16px;
  box-shadow: 0 24px 64px rgba(0, 0, 0, 0.28);
  width: 100%;
  max-width: 480px;
  max-height: 92vh;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  position: relative;
}

.wizard-close {
  position: absolute;
  top: 14px;
  right: 14px;
  background: none;
  border: none;
  cursor: pointer;
  color: var(--secondary-text-color);
  padding: 6px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  transition: background 0.15s;
  z-index: 1;
}
.wizard-close:hover { background: var(--hover-background-color); }

/* ── Progress bar ── */
.steps-bar {
  display: flex;
  align-items: center;
  padding: 24px 48px 0;
}

.step-dot {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  border: 2px solid var(--border-color);
  background: var(--card-color);
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: border-color 0.25s, background 0.25s;
  color: white;
}
.step-dot.active {
  border-color: var(--primary-color);
  background: var(--primary-color);
}
.step-dot.done {
  border-color: var(--primary-color);
  background: var(--primary-color);
}

.step-line {
  flex: 1;
  height: 2px;
  background: var(--border-color);
  transition: background 0.3s;
}
.step-line.filled { background: var(--primary-color); }

/* ── Body ── */
.wizard-body {
  padding: 28px 32px 8px;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  flex: 1;
}

.wizard-title {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--main-text-color);
  margin: 0;
  line-height: 1.3;
}

.wizard-desc {
  font-size: 0.9rem;
  color: var(--secondary-text-color);
  margin: 0;
  line-height: 1.6;
  max-width: 360px;
}

/* Step 1 – org avatar */
.org-avatar-lg {
  width: 64px;
  height: 64px;
  border-radius: 16px;
  background: var(--primary-color);
  color: white;
  font-size: 1.8rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

/* Step 1 – upcoming list */
.upcoming-steps {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
  max-width: 300px;
  margin-top: 4px;
  text-align: left;
}
.upcoming-item {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 0.875rem;
  color: var(--main-text-color);
}
.upcoming-num {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background: color-mix(in srgb, var(--primary-color) 14%, transparent);
  color: var(--primary-color);
  font-size: 0.78rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

/* Steps 2 & 3 – feature icon */
.step-icon-circle {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  background: color-mix(in srgb, var(--primary-color) 12%, transparent);
  color: var(--primary-color);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

/* Step 2 – action block */
.step-action { width: 100%; }

.btn-full {
  width: 100%;
  justify-content: center;
  gap: 8px;
  padding: 12px 20px;
}

.invite-result {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 8px;
  align-items: flex-start;
}
.success-tag {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  background: color-mix(in srgb, #2A9D8F 12%, transparent);
  color: #2A9D8F;
  font-size: 0.78rem;
  font-weight: 600;
  padding: 3px 10px;
  border-radius: 20px;
}

.link-row {
  display: flex;
  gap: 6px;
  width: 100%;
}
.link-input {
  flex: 1;
  font-size: 0.78rem;
  padding: 8px 10px;
  background: var(--background-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  color: var(--secondary-text-color);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.btn-copy {
  flex-shrink: 0;
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 8px 10px;
  cursor: pointer;
  display: flex;
  align-items: center;
  color: var(--secondary-text-color);
  transition: background 0.15s;
}
.btn-copy:hover { background: var(--hover-background-color); }

.invite-hint {
  font-size: 0.78rem;
  color: var(--secondary-text-color);
  margin: 0;
  line-height: 1.5;
}

/* Step 3 – chips */
.chip-row {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  justify-content: center;
  margin-bottom: 10px;
}
.chip {
  background: var(--background-color);
  border: 1px solid var(--border-color);
  border-radius: 20px;
  padding: 5px 14px;
  font-size: 0.82rem;
  color: var(--main-text-color);
  cursor: pointer;
  transition: border-color 0.15s, background 0.15s;
}
.chip:hover { border-color: var(--primary-color); }
.chip.active {
  border-color: var(--primary-color);
  background: color-mix(in srgb, var(--primary-color) 10%, transparent);
  color: var(--primary-color);
  font-weight: 600;
}

.folder-row {
  display: flex;
  gap: 8px;
  width: 100%;
}
.folder-row .input-field { flex: 1; }

.folder-done {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}
.folder-done-check { color: #2A9D8F; }
.folder-done p { font-size: 0.9rem; color: var(--main-text-color); margin: 0; font-weight: 500; }

/* Step 4 – done screen */
.wizard-body-done { padding-top: 24px; }

.checkmark-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
}
.checkmark-svg {
  width: 72px;
  height: 72px;
  stroke: var(--primary-color);
  stroke-width: 2;
}
.ck-circle {
  stroke-dasharray: 157;
  stroke-dashoffset: 157;
  stroke: color-mix(in srgb, var(--primary-color) 22%, transparent);
  animation: ck-stroke 0.55s cubic-bezier(0.65, 0, 0.45, 1) forwards;
}
.ck-check {
  stroke-dasharray: 34;
  stroke-dashoffset: 34;
  stroke-width: 3;
  stroke-linecap: round;
  stroke-linejoin: round;
  animation: ck-stroke 0.3s cubic-bezier(0.65, 0, 0.45, 1) 0.48s forwards;
}
@keyframes ck-stroke { to { stroke-dashoffset: 0; } }

.done-list {
  list-style: none;
  padding: 0;
  margin: 4px 0 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
  text-align: left;
  width: 100%;
  max-width: 280px;
}
.done-list li {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 0.875rem;
  color: var(--main-text-color);
}

.check-dot {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  border: 2px solid var(--border-color);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: border-color 0.2s, background 0.2s;
}
.check-dot.done {
  border-color: var(--primary-color);
  background: var(--primary-color);
  color: white;
}

/* ── Footer ── */
.wizard-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 24px 20px;
  border-top: 1px solid var(--border-color);
  margin-top: 12px;
}

.footer-spacer { flex: 1; }
.footer-right { margin-left: auto; }

.btn-skip {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--secondary-text-color);
  font-size: 0.875rem;
  padding: 8px 4px;
  transition: color 0.15s;
}
.btn-skip:hover { color: var(--main-text-color); }

.btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: var(--primary-color);
  color: white;
  border: none;
  border-radius: 8px;
  padding: 10px 18px;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.2s, transform 0.1s;
}
.btn-primary:hover { opacity: 0.9; }
.btn-primary:active { transform: scale(0.98); }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }
.btn-start { padding: 11px 22px; }

/* ── Form elements ── */
.input-field {
  background: var(--background-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 10px 12px;
  font-size: 0.9rem;
  color: var(--main-text-color);
  transition: border-color 0.15s;
  box-sizing: border-box;
}
.input-field:focus { outline: none; border-color: var(--primary-color); }

.form-error {
  color: #ef4444;
  font-size: 0.8rem;
  margin: 4px 0 0;
  text-align: left;
}

.spinner-sm {
  display: inline-block;
  width: 13px;
  height: 13px;
  border: 2px solid rgba(255,255,255,0.4);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

/* ── Transitions ── */
.wizard-enter-active, .wizard-leave-active { transition: opacity 0.2s; }
.wizard-enter-active .wizard-modal,
.wizard-leave-active .wizard-modal { transition: transform 0.2s; }
.wizard-enter-from, .wizard-leave-to { opacity: 0; }
.wizard-enter-from .wizard-modal { transform: scale(0.96) translateY(10px); }
.wizard-leave-to .wizard-modal    { transform: scale(0.96) translateY(10px); }

.slide-enter-active { transition: opacity 0.18s ease, transform 0.18s ease; }
.slide-leave-active { transition: opacity 0.12s ease, transform 0.12s ease; }
.slide-enter-from   { opacity: 0; transform: translateX(20px); }
.slide-leave-to     { opacity: 0; transform: translateX(-20px); }

@media (max-width: 520px) {
  .wizard-body { padding: 20px 20px 8px; }
  .steps-bar   { padding: 20px 32px 0; }
  .wizard-footer { padding: 14px 20px 18px; }
}
</style>
