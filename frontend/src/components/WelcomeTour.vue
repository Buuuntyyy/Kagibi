<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <Transition name="wizard">
    <div v-if="visible" class="wizard-overlay" @click.self="handleSkip">
      <div class="wizard-modal" role="dialog" aria-modal="true" aria-labelledby="tour-title">

        <button class="wizard-close" @click="handleSkip" :aria-label="t('common.close')">
          <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
            <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/>
          </svg>
        </button>

        <!-- Progress dots -->
        <div class="steps-bar" aria-hidden="true">
          <template v-for="i in STEPS" :key="i">
            <div class="step-dot" :class="{ active: step === i, done: step > i }">
              <svg v-if="step > i" viewBox="0 0 24 24" width="9" height="9" fill="currentColor">
                <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41L9 16.17z"/>
              </svg>
            </div>
            <div v-if="i < STEPS" class="step-line" :class="{ filled: step > i }"></div>
          </template>
        </div>

        <Transition name="slide" mode="out-in">

          <!-- Step 1: Welcome -->
          <div v-if="step === 1" key="s1" class="wizard-body">
            <div class="tour-logo">
              <svg viewBox="0 0 24 24" width="34" height="34" fill="currentColor">
                <path d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4z"/>
              </svg>
            </div>
            <h2 class="wizard-title" id="tour-title">{{ t('tour.step1.title') }}</h2>
            <p class="wizard-desc">{{ t('tour.step1.desc') }}</p>
            <div class="feature-grid">
              <div class="feature-card">
                <svg viewBox="0 0 24 24" width="17" height="17" fill="currentColor" class="fc-icon">
                  <path d="M2 20h20v-4H2v4zm2-3h2v2H4v-2zM2 4v4h20V4H2zm4 3H4V5h2v2zm-4 7h20v-4H2v4zm2-3h2v2H4v-2z"/>
                </svg>
                <span>{{ t('tour.step1.feat1') }}</span>
              </div>
              <div class="feature-card">
                <svg viewBox="0 0 24 24" width="17" height="17" fill="currentColor" class="fc-icon">
                  <path d="M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z"/>
                </svg>
                <span>{{ t('tour.step1.feat2') }}</span>
              </div>
              <div class="feature-card">
                <svg viewBox="0 0 24 24" width="17" height="17" fill="currentColor" class="fc-icon">
                  <path d="M9.01 14H2v2h7.01v3L13 15l-3.99-4v3zm5.98-1v-3H22V8h-7.01V5L11 9l3.99 4z"/>
                </svg>
                <span>{{ t('tour.step1.feat3') }}</span>
              </div>
              <div class="feature-card">
                <svg viewBox="0 0 24 24" width="17" height="17" fill="currentColor" class="fc-icon">
                  <path d="M18 16.08c-.76 0-1.44.3-1.96.77L8.91 12.7c.05-.23.09-.46.09-.7s-.04-.47-.09-.7l7.05-4.11c.54.5 1.25.81 2.04.81 1.66 0 3-1.34 3-3s-1.34-3-3-3-3 1.34-3 3c0 .24.04.47.09.7L8.04 9.81C7.5 9.31 6.79 9 6 9c-1.66 0-3 1.34-3 3s1.34 3 3 3c.79 0 1.5-.31 2.04-.81l7.12 4.16c-.05.21-.08.43-.08.65 0 1.61 1.31 2.92 2.92 2.92 1.61 0 2.92-1.31 2.92-2.92s-1.31-2.92-2.92-2.92z"/>
                </svg>
                <span>{{ t('tour.step1.feat4') }}</span>
              </div>
              <div class="feature-card feature-card-wide">
                <svg viewBox="0 0 24 24" width="17" height="17" fill="currentColor" class="fc-icon">
                  <path d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4zm-2 16l-4-4 1.41-1.41L10 14.17l6.59-6.59L18 9l-8 8z"/>
                </svg>
                <span>{{ t('tour.step1.feat5') }}</span>
              </div>
            </div>
          </div>

          <!-- Step 2: Storage -->
          <div v-else-if="step === 2" key="s2" class="wizard-body">
            <div class="step-icon-circle">
              <svg viewBox="0 0 24 24" width="26" height="26" fill="currentColor">
                <path d="M2 20h20v-4H2v4zm2-3h2v2H4v-2zM2 4v4h20V4H2zm4 3H4V5h2v2zm-4 7h20v-4H2v4zm2-3h2v2H4v-2z"/>
              </svg>
            </div>
            <h2 class="wizard-title">{{ t('tour.step2.title') }}</h2>
            <p class="wizard-desc">{{ t('tour.step2.desc') }}</p>
            <div class="info-pills">
              <span class="pill">{{ t('tour.step2.pill1') }}</span>
              <span class="pill">{{ t('tour.step2.pill2') }}</span>
              <span class="pill">{{ t('tour.step2.pill3') }}</span>
            </div>
          </div>

          <!-- Step 3: Organizations -->
          <div v-else-if="step === 3" key="s3" class="wizard-body">
            <div class="step-icon-circle">
              <svg viewBox="0 0 24 24" width="26" height="26" fill="currentColor">
                <path d="M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z"/>
              </svg>
            </div>
            <h2 class="wizard-title">{{ t('tour.step3.title') }}</h2>
            <p class="wizard-desc">{{ t('tour.step3.desc') }}</p>
            <div class="info-pills">
              <span class="pill">{{ t('tour.step3.pill1') }}</span>
              <span class="pill">{{ t('tour.step3.pill2') }}</span>
              <span class="pill">{{ t('tour.step3.pill3') }}</span>
            </div>
          </div>

          <!-- Step 4: P2P -->
          <div v-else-if="step === 4" key="s4" class="wizard-body">
            <div class="step-icon-circle">
              <svg viewBox="0 0 24 24" width="26" height="26" fill="currentColor">
                <path d="M9.01 14H2v2h7.01v3L13 15l-3.99-4v3zm5.98-1v-3H22V8h-7.01V5L11 9l3.99 4z"/>
              </svg>
            </div>
            <h2 class="wizard-title">{{ t('tour.step4.title') }}</h2>
            <p class="wizard-desc">{{ t('tour.step4.desc') }}</p>
            <div class="info-pills">
              <span class="pill">{{ t('tour.step4.pill1') }}</span>
              <span class="pill">{{ t('tour.step4.pill2') }}</span>
              <span class="pill">{{ t('tour.step4.pill3') }}</span>
            </div>
          </div>

          <!-- Step 5: File Sharing -->
          <div v-else-if="step === 5" key="s5" class="wizard-body">
            <div class="step-icon-circle">
              <svg viewBox="0 0 24 24" width="26" height="26" fill="currentColor">
                <path d="M18 16.08c-.76 0-1.44.3-1.96.77L8.91 12.7c.05-.23.09-.46.09-.7s-.04-.47-.09-.7l7.05-4.11c.54.5 1.25.81 2.04.81 1.66 0 3-1.34 3-3s-1.34-3-3-3-3 1.34-3 3c0 .24.04.47.09.7L8.04 9.81C7.5 9.31 6.79 9 6 9c-1.66 0-3 1.34-3 3s1.34 3 3 3c.79 0 1.5-.31 2.04-.81l7.12 4.16c-.05.21-.08.43-.08.65 0 1.61 1.31 2.92 2.92 2.92 1.61 0 2.92-1.31 2.92-2.92s-1.31-2.92-2.92-2.92z"/>
              </svg>
            </div>
            <h2 class="wizard-title">{{ t('tour.step5.title') }}</h2>
            <p class="wizard-desc">{{ t('tour.step5.desc') }}</p>
            <div class="info-pills">
              <span class="pill">{{ t('tour.step5.pill1') }}</span>
              <span class="pill">{{ t('tour.step5.pill2') }}</span>
              <span class="pill">{{ t('tour.step5.pill3') }}</span>
            </div>
          </div>

          <!-- Step 6: MFA -->
          <div v-else-if="step === 6" key="s6" class="wizard-body">
            <div class="step-icon-circle">
              <svg viewBox="0 0 24 24" width="26" height="26" fill="currentColor">
                <path d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4zm-2 16l-4-4 1.41-1.41L10 14.17l6.59-6.59L18 9l-8 8z"/>
              </svg>
            </div>
            <h2 class="wizard-title">{{ t('tour.step6.title') }}</h2>
            <p class="wizard-desc">{{ t('tour.step6.desc') }}</p>
            <div class="info-pills">
              <span class="pill">{{ t('tour.step6.pill1') }}</span>
              <span class="pill">{{ t('tour.step6.pill2') }}</span>
            </div>
          </div>

          <!-- Step 7: Done -->
          <div v-else-if="step === 7" key="s7" class="wizard-body wizard-body-done">
            <div class="checkmark-wrap">
              <svg viewBox="0 0 52 52" class="checkmark-svg" aria-hidden="true">
                <circle class="ck-circle" cx="26" cy="26" r="25" fill="none"/>
                <path class="ck-check" fill="none" d="M14.1 27.2l7.1 7.2 16.7-16.8"/>
              </svg>
            </div>
            <h2 class="wizard-title">{{ t('tour.done.title') }}</h2>
            <p class="wizard-desc">{{ t('tour.done.desc') }}</p>
          </div>

        </Transition>

        <!-- Footer -->
        <div class="wizard-footer">
          <button v-if="step < STEPS" class="btn-skip" @click="handleSkip">
            {{ t('tour.skip') }}
          </button>
          <div v-else class="footer-spacer"></div>

          <div class="footer-right">
            <button v-if="step < STEPS" class="btn-primary" @click="nextStep">
              {{ step === STEPS - 1 ? t('tour.finish') : t('tour.next') }}
              <svg viewBox="0 0 24 24" width="15" height="15" fill="currentColor">
                <path d="M8.59 16.59L13.17 12 8.59 7.41 10 6l6 6-6 6-1.41-1.41z"/>
              </svg>
            </button>
            <button v-else class="btn-primary btn-start" @click="handleDone">
              {{ t('tour.startUsing') }}
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
import { useRoute } from 'vue-router'
import { usePreferencesStore } from '../stores/preferences'
import { useAuthStore } from '../stores/auth'

const { t } = useI18n()
const route = useRoute()
const prefsStore = usePreferencesStore()
const authStore = useAuthStore()

const STEPS = 7
const step = ref(1)

const isLandingPage = computed(() =>
  ['LandingHome', 'Pricing', 'Transfer', 'Compare', 'Security'].includes(route.name)
)

const visible = computed(() =>
  authStore.isAuthenticated && !prefsStore.tutorialDone && !isLandingPage.value
)

function nextStep() {
  if (step.value < STEPS) step.value++
}

function handleSkip() {
  prefsStore.tutorialDone = true
}

function handleDone() {
  prefsStore.tutorialDone = true
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
  padding: 24px 28px 0;
}

.step-dot {
  width: 18px;
  height: 18px;
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
  font-size: 1.2rem;
  font-weight: 700;
  color: var(--main-text-color);
  margin: 0;
  line-height: 1.3;
}

.wizard-desc {
  font-size: 0.88rem;
  color: var(--secondary-text-color);
  margin: 0;
  line-height: 1.65;
  max-width: 360px;
}

/* ── Step 1: logo + feature grid ── */
.tour-logo {
  width: 64px;
  height: 64px;
  border-radius: 16px;
  background: var(--primary-color);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.feature-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  width: 100%;
  max-width: 340px;
  margin-top: 4px;
}

.feature-card {
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--background-color);
  border: 1px solid var(--border-color);
  border-radius: 10px;
  padding: 10px 12px;
  font-size: 0.81rem;
  color: var(--main-text-color);
  text-align: left;
}

.feature-card-wide {
  grid-column: 1 / -1;
}

.fc-icon {
  color: var(--primary-color);
  flex-shrink: 0;
}

/* ── Steps 2–6: feature icon ── */
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

/* ── Info pills ── */
.info-pills {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: center;
  margin-top: 4px;
}

.pill {
  background: color-mix(in srgb, var(--primary-color) 10%, transparent);
  color: var(--primary-color);
  border: 1px solid color-mix(in srgb, var(--primary-color) 25%, transparent);
  border-radius: 20px;
  padding: 5px 14px;
  font-size: 0.81rem;
  font-weight: 500;
}

/* ── Step 7: done screen ── */
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
.btn-start { padding: 11px 22px; }

/* ── Transitions ── */
.wizard-enter-active, .wizard-leave-active { transition: opacity 0.2s; }
.wizard-enter-active .wizard-modal,
.wizard-leave-active .wizard-modal { transition: transform 0.2s; }
.wizard-enter-from, .wizard-leave-to { opacity: 0; }
.wizard-enter-from .wizard-modal { transform: scale(0.96) translateY(10px); }
.wizard-leave-to .wizard-modal   { transform: scale(0.96) translateY(10px); }

.slide-enter-active { transition: opacity 0.18s ease, transform 0.18s ease; }
.slide-leave-active { transition: opacity 0.12s ease, transform 0.12s ease; }
.slide-enter-from   { opacity: 0; transform: translateX(20px); }
.slide-leave-to     { opacity: 0; transform: translateX(-20px); }

@media (max-width: 520px) {
  .wizard-body    { padding: 20px 20px 8px; }
  .steps-bar      { padding: 20px 20px 0; }
  .wizard-footer  { padding: 14px 20px 18px; }
  .feature-grid   { grid-template-columns: 1fr; }
  .feature-card-wide { grid-column: 1; }
}
</style>
