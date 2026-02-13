/**
 * useMFA.js
 * Composable for managing MFA (TOTP) with Supabase
 *
 * Features:
 * - Enroll/Unenroll TOTP
 * - Challenge and verify TOTP codes
 * - Manage security settings (when MFA is required)
 * - Check AAL (Authentication Assurance Level)
 */

import { ref, computed } from 'vue'
import { supabase } from '../supabase'
import api from '../api'

export function useMFA() {
  // State
  const enrolling = ref(false)
  const verifying = ref(false)
  const unenrolling = ref(false)
  const loadingSettings = ref(false)
  const settingsLoaded = ref(false)
  const challengeInProgress = ref(false)

  const qrCode = ref(null)
  const secret = ref(null) // Only exposed during initial enrollment
  const factorId = ref(null)

  const securitySettings = ref({
    mfa_enabled: false,
    mfa_verified: false,
    require_mfa_on_login: false,
    require_mfa_on_destructive_actions: false,
    require_mfa_on_downloads: false
  })

  const error = ref(null)

  // Computed
  const isMFAEnabled = computed(() => securitySettings.value.mfa_enabled)
  const isMFAVerified = computed(() => securitySettings.value.mfa_verified)

  /**
   * Get current Authentication Assurance Level
   * @returns {Promise<'aal1'|'aal2'|null>}
   */
  async function getAAL() {
    try {
      const { data, error: sessionError } = await supabase.auth.getSession()
      if (sessionError) throw sessionError

      return data.session?.user?.aal || 'aal1'
    } catch (err) {
      console.error('[useMFA] Failed to get AAL:', err)
      return null
    }
  }

  /**
   * Check if current session meets AAL2 requirement
   * @returns {Promise<boolean>}
   */
  async function hasAAL2() {
    const aal = await getAAL()
    return aal === 'aal2'
  }

  /**
   * Sync MFA enabled/verified flags from Supabase factors
   * This keeps UI accurate even if backend settings are stale.
   * @returns {Promise<boolean>} true if a verified TOTP factor exists
   */
  async function syncMFAStatusFromFactors() {
    const { data: factors, error: listError } = await supabase.auth.mfa.listFactors()
    if (listError) throw listError

    const hasVerifiedTotp = factors?.totp?.some((factor) => factor.status === 'verified') || false
    securitySettings.value = {
      ...securitySettings.value,
      mfa_enabled: hasVerifiedTotp,
      mfa_verified: hasVerifiedTotp
    }

    return hasVerifiedTotp
  }

  /**
   * Clean up any unverified/incomplete MFA factors
   * This prevents "factor already exists" errors when re-enrolling
   * @returns {Promise<void>}
   */
  async function cleanupUnverifiedFactors() {
    try {
      const { data: factors, error: listError } = await supabase.auth.mfa.listFactors()
      if (listError) throw listError

      console.log('[useMFA] Current factors:', factors)

      // Unenroll any UNVERIFIED TOTP factors (verified ones cannot be unenrolled without code)
      if (factors.totp && factors.totp.length > 0) {
        for (const factor of factors.totp) {
          // Only try to unenroll if factor is NOT verified (status !== 'verified')
          if (factor.status !== 'verified') {
            console.log('[useMFA] Attempting to cleanup unverified factor:', factor.id)
            const { error: unenrollError } = await supabase.auth.mfa.unenroll({
              factorId: factor.id
            })

            if (unenrollError) {
              console.warn('[useMFA] Failed to cleanup factor:', factor.id, unenrollError)
            } else {
              console.log('[useMFA] Successfully cleaned up factor:', factor.id)
            }
          } else {
            console.log('[useMFA] Skipping verified factor:', factor.id)
          }
        }
      }
    } catch (err) {
      console.warn('[useMFA] Cleanup failed:', err)
      // Don't throw - we'll try enrollment anyway
    }
  }

  /**
   * Enroll TOTP MFA
   * @returns {Promise<{qrCode: string, secret: string, factorId: string}>}
   */
  async function enrollMFA() {
    enrolling.value = true
    error.value = null

    try {
      // Step -1: If a verified factor already exists, do not re-enroll
      const { data: existingFactors, error: listError } = await supabase.auth.mfa.listFactors()
      if (listError) throw listError
      const hasVerifiedTotp = existingFactors?.totp?.some((factor) => factor.status === 'verified')
      if (hasVerifiedTotp) {
        const message = 'MFA deja active. Desactivez-le avant de le reconfigurer.'
        error.value = message
        throw new Error(message)
      }

      // Step 0: Clean up any existing unverified factors
      await cleanupUnverifiedFactors()

      // Step 1: Enroll TOTP factor with unique name
      const timestamp = Date.now()
      const { data, error: enrollError } = await supabase.auth.mfa.enroll({
        factorType: 'totp',
        friendlyName: `SaferCloud Authenticator ${timestamp}`
      })

      console.log('[useMFA] Enrollment response:', { data, enrollError })

      if (enrollError) throw enrollError

      // Store enrollment data - use URI not qr_code
      qrCode.value = data.totp.uri  // This is the otpauth:// URI
      secret.value = data.totp.secret
      factorId.value = data.id

      console.log('[useMFA] MFA enrollment started')
      console.log('[useMFA] Factor ID:', data.id)
      console.log('[useMFA] TOTP URI:', data.totp.uri)
      console.log('[useMFA] Secret:', data.totp.secret)
      console.log('[useMFA] qrCode.value set to:', qrCode.value)

      return {
        qrCode: data.totp.uri,  // Return the URI to encode as QR
        secret: data.totp.secret,
        factorId: data.id
      }

    } catch (err) {
      console.error('[useMFA] Enrollment failed:', err)
      error.value = err.message || 'Erreur lors de l\'activation du MFA'
      throw err
    } finally {
      enrolling.value = false
    }
  }

  /**
   * Verify TOTP code and complete enrollment
   * @param {string} code - 6-digit TOTP code
   * @returns {Promise<boolean>}
   */
  async function verifyAndEnableMFA(code) {
    if (!factorId.value) {
      throw new Error('No enrollment in progress. Call enrollMFA first.')
    }

    verifying.value = true
    error.value = null

    try {
      // Step 2: Challenge the factor to verify
      const { data: challengeData, error: challengeError } = await supabase.auth.mfa.challenge({
        factorId: factorId.value
      })

      if (challengeError) throw challengeError

      // Step 3: Verify the TOTP code
      const { data: verifyData, error: verifyError } = await supabase.auth.mfa.verify({
        factorId: factorId.value,
        challengeId: challengeData.id,
        code: code
      })

      if (verifyError) throw verifyError

      console.log('[useMFA] MFA verified successfully')

      // Step 4: Update security settings in backend
      await updateSecuritySettings({
        mfa_enabled: true,
        mfa_verified: true
      })

      // Refresh to ensure sync
      await fetchSecuritySettings()

      // Clear sensitive data
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

  /**
   * Unenroll MFA (disable TOTP)
   * @param {string} confirmationCode - TOTP code to confirm unenrollment
   * @returns {Promise<boolean>}
   */
  async function unenrollMFA(confirmationCode) {
    unenrolling.value = true
    error.value = null

    try {
      // List all factors
      const { data: factors, error: listError } = await supabase.auth.mfa.listFactors()
      if (listError) throw listError

      const totpFactor = factors.totp?.[0]
      if (!totpFactor) {
        throw new Error('No TOTP factor found')
      }

      // Verify code before unenrolling (security measure)
      const { data: challengeData, error: challengeError } = await supabase.auth.mfa.challenge({
        factorId: totpFactor.id
      })

      if (challengeError) throw challengeError

      const { error: verifyError } = await supabase.auth.mfa.verify({
        factorId: totpFactor.id,
        challengeId: challengeData.id,
        code: confirmationCode
      })

      if (verifyError) throw verifyError

      // Unenroll the factor
      const { error: unenrollError } = await supabase.auth.mfa.unenroll({
        factorId: totpFactor.id
      })

      if (unenrollError) throw unenrollError

      console.log('[useMFA] MFA unenrolled successfully')

      // Update security settings in backend
      await updateSecuritySettings({
        mfa_enabled: false,
        mfa_verified: false,
        require_mfa_on_login: false,
        require_mfa_on_destructive_actions: false,
        require_mfa_on_downloads: false
      })

      // Refresh to ensure sync
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

  /**
   * Create a new MFA challenge (for AAL1 -> AAL2 elevation)
   * @returns {Promise<{challengeId: string, factorId: string}>}
   */
  async function createChallenge() {
    challengeInProgress.value = true
    error.value = null

    try {
      // Get the user's TOTP factor
      const { data: factors, error: listError } = await supabase.auth.mfa.listFactors()
      if (listError) throw listError

      const totpFactor = factors.totp?.[0]
      if (!totpFactor || totpFactor.status !== 'verified') {
        throw new Error('MFA not configured or not verified')
      }

      // Create challenge
      const { data, error: challengeError } = await supabase.auth.mfa.challenge({
        factorId: totpFactor.id
      })

      if (challengeError) throw challengeError

      return {
        challengeId: data.id,
        factorId: totpFactor.id
      }

    } catch (err) {
      console.error('[useMFA] Challenge creation failed:', err)
      error.value = err.message || 'Erreur lors de la création du challenge MFA'
      throw err
    } finally {
      challengeInProgress.value = false
    }
  }

  /**
   * Verify MFA code for an existing challenge (elevate to AAL2)
   * @param {string} challengeId
   * @param {string} factorId
   * @param {string} code
   * @returns {Promise<boolean>}
   */
  async function verifyChallenge(challengeId, factorId, code) {
    verifying.value = true
    error.value = null

    try {
      const { data, error: verifyError } = await supabase.auth.mfa.verify({
        factorId: factorId,
        challengeId: challengeId,
        code: code
      })

      if (verifyError) throw verifyError

      console.log('[useMFA] Challenge verified, AAL elevated to aal2')
      return true

    } catch (err) {
      console.error('[useMFA] Challenge verification failed:', err)
      error.value = err.message || 'Code invalide'
      throw err
    } finally {
      verifying.value = false
    }
  }

  /**
   * Fetch user security settings from backend
   * @returns {Promise<Object>}
   */
  async function fetchSecuritySettings() {
    loadingSettings.value = true
    error.value = null

    try {
      const response = await api.get('/users/security-settings')
      securitySettings.value = response.data
      try {
        await syncMFAStatusFromFactors()
      } catch (syncErr) {
        console.warn('[useMFA] Failed to sync MFA status from factors:', syncErr)
      }
      settingsLoaded.value = true
      return securitySettings.value
    } catch (err) {
      console.error('[useMFA] Failed to fetch security settings:', err)
      error.value = err.message || 'Erreur lors du chargement des paramètres'
      try {
        await syncMFAStatusFromFactors()
      } catch (syncErr) {
        console.warn('[useMFA] Failed to sync MFA status from factors:', syncErr)
      }
      settingsLoaded.value = true
      return securitySettings.value
    } finally {
      loadingSettings.value = false
    }
  }

  /**
   * Update user security settings in backend
   * @param {Object} updates - Partial security settings
   * @returns {Promise<Object>}
   */
  async function updateSecuritySettings(updates) {
    loadingSettings.value = true
    error.value = null

    try {
      const response = await api.put('/users/security-settings', updates)
      securitySettings.value = { ...securitySettings.value, ...response.data }
      console.log('[useMFA] Security settings updated:', response.data)
      return response.data
    } catch (err) {
      console.error('[useMFA] Failed to update security settings:', err)
      error.value = err.message || 'Erreur lors de la mise à jour des paramètres'
      throw err
    } finally {
      loadingSettings.value = false
    }
  }

  /**
   * Check if MFA is required based on settings and context
   * @param {'login'|'destructive'|'download'} context
   * @returns {Promise<boolean>}
   */
  async function isMFARequired(context) {
    await fetchSecuritySettings()

    if (!securitySettings.value.mfa_enabled || !securitySettings.value.mfa_verified) {
      return false
    }

    switch (context) {
      case 'login':
        return securitySettings.value.require_mfa_on_login
      case 'destructive':
        return securitySettings.value.require_mfa_on_destructive_actions
      case 'download':
        return securitySettings.value.require_mfa_on_downloads
      default:
        return false
    }
  }

  return {
    // State
    enrolling,
    verifying,
    unenrolling,
    loadingSettings,
    settingsLoaded,
    challengeInProgress,
    qrCode,
    secret,
    factorId,
    securitySettings,
    error,

    // Computed
    isMFAEnabled,
    isMFAVerified,

    // Methods
    getAAL,
    hasAAL2,
    enrollMFA,
    verifyAndEnableMFA,
    unenrollMFA,
    createChallenge,
    verifyChallenge,
    fetchSecuritySettings,
    updateSecuritySettings,
    isMFARequired
  }
}
