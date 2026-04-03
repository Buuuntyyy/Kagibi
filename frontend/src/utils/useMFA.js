/**
 * useMFA.js
 * Composable for managing MFA (TOTP).
 *
 * Supabase mode: Full TOTP enrollment/verification via Supabase MFA API.
 * PocketBase mode: MFA is not supported — all functions return gracefully with no-op results.
 */

import { ref, computed } from 'vue'
import { authClient } from '../auth-client'
import api from '../api'

export function useMFA() {
  const enrolling = ref(false)
  const verifying = ref(false)
  const unenrolling = ref(false)
  const loadingSettings = ref(false)
  const settingsLoaded = ref(false)
  const challengeInProgress = ref(false)

  const qrCode = ref(null)
  const secret = ref(null)
  const factorId = ref(null)

  const securitySettings = ref({
    mfa_enabled: false,
    mfa_verified: false,
    require_mfa_on_login: false,
    require_mfa_on_destructive_actions: false,
    require_mfa_on_downloads: false
  })

  const error = ref(null)

  const isMFAEnabled = computed(() => securitySettings.value.mfa_enabled)
  const isMFAVerified = computed(() => securitySettings.value.mfa_verified)

  async function getAAL() {
    if (!authClient.isMFASupported) return 'aal1'
    try {
      const { data, error: sessionError } = await authClient.getSession()
      if (sessionError) throw sessionError
      return data.session?.user?.aal || 'aal1'
    } catch (err) {
      console.error('[useMFA] Failed to get AAL:', err)
      return null
    }
  }

  async function hasAAL2() {
    const aal = await getAAL()
    return aal === 'aal2'
  }

  async function syncMFAStatusFromFactors() {
    if (!authClient.isMFASupported) return false
    const { data: factors, error: listError } = await authClient.mfa.listFactors()
    if (listError) throw listError
    const hasVerifiedTotp = factors?.totp?.some((f) => f.status === 'verified') || false
    securitySettings.value = { ...securitySettings.value, mfa_enabled: hasVerifiedTotp, mfa_verified: hasVerifiedTotp }
    return hasVerifiedTotp
  }

  async function cleanupUnverifiedFactors() {
    if (!authClient.isMFASupported) return
    try {
      const { data: factors, error: listError } = await authClient.mfa.listFactors()
      if (listError) throw listError
      if (factors.totp) {
        for (const factor of factors.totp) {
          if (factor.status !== 'verified') {
            await authClient.mfa.unenroll({ factorId: factor.id })
          }
        }
      }
    } catch (err) {
      console.warn('[useMFA] Cleanup failed:', err)
    }
  }

  async function enrollMFA() {
    if (!authClient.isMFASupported) {
      error.value = 'MFA not supported with the current auth provider'
      throw new Error(error.value)
    }

    enrolling.value = true
    error.value = null

    try {
      const { data: existingFactors, error: listError } = await authClient.mfa.listFactors()
      if (listError) throw listError
      if (existingFactors?.totp?.some((f) => f.status === 'verified')) {
        const message = 'MFA deja active. Desactivez-le avant de le reconfigurer.'
        error.value = message
        throw new Error(message)
      }

      await cleanupUnverifiedFactors()

      const timestamp = Date.now()
      const { data, error: enrollError } = await authClient.mfa.enroll({
        factorType: 'totp',
        friendlyName: `SaferCloud Authenticator ${timestamp}`
      })
      if (enrollError) throw enrollError

      // Get user email for TOTP account name
      const session = await authClient.getSession()
      const userEmail = session?.data?.session?.user?.email || 'user@safercloud.app'

      const issuer = 'SaferCloud'
      const customUri = `otpauth://totp/${encodeURIComponent(issuer)}:${encodeURIComponent(userEmail)}?secret=${data.totp.secret}&issuer=${encodeURIComponent(issuer)}`

      qrCode.value = customUri
      secret.value = data.totp.secret
      factorId.value = data.id

      return { qrCode: customUri, secret: data.totp.secret, factorId: data.id }

    } catch (err) {
      console.error('[useMFA] Enrollment failed:', err)
      error.value = err.message || 'Erreur lors de l\'activation du MFA'
      throw err
    } finally {
      enrolling.value = false
    }
  }

  async function verifyAndEnableMFA(code) {
    if (!authClient.isMFASupported) throw new Error('MFA not supported')
    if (!factorId.value) throw new Error('No enrollment in progress. Call enrollMFA first.')

    verifying.value = true
    error.value = null

    try {
      const { data: challengeData, error: challengeError } = await authClient.mfa.challenge({ factorId: factorId.value })
      if (challengeError) throw challengeError

      const { error: verifyError } = await authClient.mfa.verify({
        factorId: factorId.value,
        challengeId: challengeData.id,
        code
      })
      if (verifyError) throw verifyError

      await updateSecuritySettings({ mfa_enabled: true, mfa_verified: true })
      await fetchSecuritySettings()
      secret.value = null
      qrCode.value = null
      return true

    } catch (err) {
      console.error('[useMFA] Verification failed:', err)
      error.value = err.message || 'Code invalide'
      throw err
    } finally {
      verifying.value = false
    }
  }

  async function unenrollMFA(confirmationCode) {
    if (!authClient.isMFASupported) throw new Error('MFA not supported')

    unenrolling.value = true
    error.value = null

    try {
      const { data: factors, error: listError } = await authClient.mfa.listFactors()
      if (listError) throw listError

      const totpFactor = factors.totp?.[0]
      if (!totpFactor) throw new Error('No TOTP factor found')

      const { data: challengeData, error: challengeError } = await authClient.mfa.challenge({ factorId: totpFactor.id })
      if (challengeError) throw challengeError

      const { error: verifyError } = await authClient.mfa.verify({
        factorId: totpFactor.id,
        challengeId: challengeData.id,
        code: confirmationCode
      })
      if (verifyError) throw verifyError

      const { error: unenrollError } = await authClient.mfa.unenroll({ factorId: totpFactor.id })
      if (unenrollError) throw unenrollError

      await updateSecuritySettings({
        mfa_enabled: false, mfa_verified: false,
        require_mfa_on_login: false, require_mfa_on_destructive_actions: false, require_mfa_on_downloads: false
      })
      await fetchSecuritySettings()
      return true

    } catch (err) {
      console.error('[useMFA] Unenrollment failed:', err)
      error.value = err.message || 'Erreur lors de la désactivation du MFA'
      throw err
    } finally {
      unenrolling.value = false
    }
  }

  async function createChallenge() {
    if (!authClient.isMFASupported) throw new Error('MFA not supported')

    challengeInProgress.value = true
    error.value = null

    try {
      const { data: factors, error: listError } = await authClient.mfa.listFactors()
      if (listError) throw listError

      const totpFactor = factors.totp?.[0]
      if (!totpFactor || totpFactor.status !== 'verified') throw new Error('MFA not configured or not verified')

      const { data, error: challengeError } = await authClient.mfa.challenge({ factorId: totpFactor.id })
      if (challengeError) throw challengeError

      return { challengeId: data.id, factorId: totpFactor.id }

    } catch (err) {
      console.error('[useMFA] Challenge creation failed:', err)
      error.value = err.message || 'Erreur lors de la création du challenge MFA'
      throw err
    } finally {
      challengeInProgress.value = false
    }
  }

  async function verifyChallenge(challengeId, factorId, code) {
    if (!authClient.isMFASupported) throw new Error('MFA not supported')

    verifying.value = true
    error.value = null

    try {
      const { error: verifyError } = await authClient.mfa.verify({ factorId, challengeId, code })
      if (verifyError) throw verifyError
      return true
    } catch (err) {
      console.error('[useMFA] Challenge verification failed:', err)
      error.value = err.message || 'Code invalide'
      throw err
    } finally {
      verifying.value = false
    }
  }

  async function fetchSecuritySettings() {
    loadingSettings.value = true
    error.value = null
    try {
      const response = await api.get('/users/security-settings')
      securitySettings.value = response.data
      try { await syncMFAStatusFromFactors() } catch (e) { /* ignore */ }
      settingsLoaded.value = true
      return securitySettings.value
    } catch (err) {
      console.error('[useMFA] Failed to fetch security settings:', err)
      error.value = err.message || 'Erreur lors du chargement des paramètres'
      try { await syncMFAStatusFromFactors() } catch (e) { /* ignore */ }
      settingsLoaded.value = true
      return securitySettings.value
    } finally {
      loadingSettings.value = false
    }
  }

  async function updateSecuritySettings(updates) {
    loadingSettings.value = true
    error.value = null
    try {
      const response = await api.put('/users/security-settings', updates)
      securitySettings.value = { ...securitySettings.value, ...response.data }
      return response.data
    } catch (err) {
      console.error('[useMFA] Failed to update security settings:', err)
      error.value = err.message || 'Erreur lors de la mise à jour des paramètres'
      throw err
    } finally {
      loadingSettings.value = false
    }
  }

  async function isMFARequired(context) {
    await fetchSecuritySettings()
    if (!authClient.isMFASupported) return false
    if (!securitySettings.value.mfa_enabled || !securitySettings.value.mfa_verified) return false
    if (await hasAAL2()) return false // session is already at AAL2, no re-prompt needed
    switch (context) {
      case 'login': return securitySettings.value.require_mfa_on_login
      case 'destructive': return securitySettings.value.require_mfa_on_destructive_actions
      case 'download': return securitySettings.value.require_mfa_on_downloads
      default: return false
    }
  }

  return {
    enrolling, verifying, unenrolling, loadingSettings, settingsLoaded, challengeInProgress,
    qrCode, secret, factorId, securitySettings, error,
    isMFAEnabled, isMFAVerified,
    getAAL, hasAAL2, enrollMFA, verifyAndEnableMFA, unenrollMFA,
    createChallenge, verifyChallenge, fetchSecuritySettings, updateSecuritySettings, isMFARequired
  }
}
