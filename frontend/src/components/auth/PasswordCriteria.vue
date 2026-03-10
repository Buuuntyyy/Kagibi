<template>
  <div class="password-criteria" :class="{ 'criteria-visible': show }">
    <!-- Security tip: recommend a password manager -->
    <div class="criteria-tip">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="12" cy="12" r="10"/>
        <line x1="12" y1="16" x2="12" y2="12"/>
        <line x1="12" y1="8" x2="12.01" y2="8"/>
      </svg>
      <span>
        Pour votre sécurité, utilisez un gestionnaire de mots de passe
        (<strong>Bitwarden</strong>, <strong>1Password</strong>, <strong>KeePass</strong>)
        pour générer et stocker ce mot de passe.
      </span>
    </div>

    <!-- Criteria list -->
    <p class="criteria-title">Critères requis :</p>
    <ul class="criteria-list">
      <li :class="{ met: criteria.length, unmet: !criteria.length }">
        <span class="criteria-icon">
          <svg v-if="criteria.length" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="20 6 9 17 4 12"/>
          </svg>
          <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <line x1="18" y1="6" x2="6" y2="18"/>
            <line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </span>
        <span>20 caractères minimum <span class="criteria-count">({{ currentLength }}/20)</span></span>
      </li>
      <li :class="{ met: criteria.digits, unmet: !criteria.digits }">
        <span class="criteria-icon">
          <svg v-if="criteria.digits" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="20 6 9 17 4 12"/>
          </svg>
          <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <line x1="18" y1="6" x2="6" y2="18"/>
            <line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </span>
        <span>Au moins 2 chiffres (0–9) <span class="criteria-count">({{ currentDigits }}/2)</span></span>
      </li>
      <li :class="{ met: criteria.specials, unmet: !criteria.specials }">
        <span class="criteria-icon">
          <svg v-if="criteria.specials" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="20 6 9 17 4 12"/>
          </svg>
          <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <line x1="18" y1="6" x2="6" y2="18"/>
            <line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </span>
        <span>
          Au moins 2 caractères spéciaux non ambigus
          <span class="criteria-count">({{ currentSpecials }}/2)</span>
          <span class="specials-hint">! @ # $ % ^ &amp; * ( ) - _ = + [ ] { } : ; &lt; &gt; , . ? / ~</span>
        </span>
      </li>
    </ul>

    <!-- Strength bar -->
    <div class="strength-bar-wrapper" v-if="password.length > 0">
      <div class="strength-bar">
        <div class="strength-fill" :class="strengthClass" :style="{ width: strengthPercent + '%' }"></div>
      </div>
      <span class="strength-label" :class="strengthClass">{{ strengthLabel }}</span>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { checkPasswordCriteria } from '../../utils/passwordStrength'

const props = defineProps({
  password: { type: String, default: '' },
  show: { type: Boolean, default: true },
})

const result = computed(() => checkPasswordCriteria(props.password))
const criteria = computed(() => result.value.criteria)
const currentLength = computed(() => result.value.currentLength)
const currentDigits = computed(() => result.value.currentDigits)
const currentSpecials = computed(() => result.value.currentSpecials)

// Strength meter: 0-3 criteria met
const metCount = computed(() => Object.values(criteria.value).filter(Boolean).length)
const strengthPercent = computed(() => (metCount.value / 3) * 100)
const strengthClass = computed(() => {
  if (metCount.value === 0) return 'strength-weak'
  if (metCount.value === 1) return 'strength-weak'
  if (metCount.value === 2) return 'strength-medium'
  return 'strength-strong'
})
const strengthLabel = computed(() => {
  if (metCount.value === 3) return 'Fort ✓'
  if (metCount.value === 2) return 'Moyen'
  return 'Faible'
})
</script>

<style scoped>
.password-criteria {
  background: var(--background-color, #f8f9fa);
  border: 1px solid var(--border-color, #e2e8f0);
  border-radius: 8px;
  padding: 0.9rem 1rem;
  font-size: 0.85rem;
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}

.criteria-tip {
  display: flex;
  gap: 0.5rem;
  align-items: flex-start;
  color: var(--secondary-text-color, #64748b);
  background: rgba(52, 152, 219, 0.07);
  border-radius: 6px;
  padding: 0.55rem 0.7rem;
  line-height: 1.45;
}

.criteria-tip svg {
  width: 15px;
  height: 15px;
  flex-shrink: 0;
  margin-top: 2px;
  color: var(--primary-color, #3498db);
}

.criteria-title {
  margin: 0;
  font-weight: 600;
  color: var(--main-text-color, #1e293b);
}

.criteria-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.criteria-list li {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  line-height: 1.4;
}

.criteria-icon {
  width: 16px;
  height: 16px;
  flex-shrink: 0;
  margin-top: 1px;
}

.criteria-icon svg {
  width: 16px;
  height: 16px;
}

.met .criteria-icon { color: #27ae60; }
.unmet .criteria-icon { color: var(--error-color, #e74c3c); }
.met { color: #27ae60; }
.unmet { color: var(--error-color, #e74c3c); }

.criteria-count {
  font-weight: 700;
  margin-left: 2px;
  opacity: 0.85;
}

.specials-hint {
  display: block;
  font-size: 0.75rem;
  opacity: 0.65;
  font-family: monospace;
  margin-top: 2px;
  color: var(--secondary-text-color, #64748b);
}

/* Strength bar */
.strength-bar-wrapper {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  margin-top: 0.1rem;
}

.strength-bar {
  flex: 1;
  height: 5px;
  background: var(--border-color, #e2e8f0);
  border-radius: 99px;
  overflow: hidden;
}

.strength-fill {
  height: 100%;
  border-radius: 99px;
  transition: width 0.3s ease, background-color 0.3s ease;
}

.strength-weak  { background-color: var(--error-color, #e74c3c); color: var(--error-color, #e74c3c); }
.strength-medium { background-color: #f39c12; color: #f39c12; }
.strength-strong { background-color: #27ae60; color: #27ae60; }

.strength-label {
  font-size: 0.78rem;
  font-weight: 600;
  white-space: nowrap;
}
</style>
