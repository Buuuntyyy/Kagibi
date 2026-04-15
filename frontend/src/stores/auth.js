import { defineStore } from 'pinia'
import api from '../api'
import router from '../router'
import { authClient, IS_POCKETBASE } from '../auth-client'
import { isP2PSubdomain } from '../composables/useSubdomain'
import { useFriendStore } from './friends'
import {
  deriveKeyFromPassword, generateSalt, wrapMasterKey, unwrapMasterKey,
  hashRecoveryCode, deriveKeyFromRecoveryCode,
  generateRSAKeyPair, exportKeyToPEM, importKeyFromPEM, encryptPrivateKey, decryptPrivateKey
} from '../utils/crypto'
import sodium from 'libsodium-wrappers-sumo'
import { useMFA } from '../utils/useMFA'

const isLocalAuthBypassEnabled =
  import.meta.env.DEV && String(import.meta.env.VITE_LOCAL_BYPASS_AUTH).toLowerCase() === 'true'

const localDevBypassUser = {
  id: 'local-dev-user',
  name: 'Local Dev',
  email: 'local@dev.local',
  avatar_url: '/avatars/default.png',
  storage_used: 0,
  storage_limit: 20 * 1024 * 1024 * 1024,
  plan: 'free',
  p2p_max_exchanges: -1,
  p2p_exchanges_used: 0,
}

export const useAuthStore = defineStore('auth', {
  state: () => {
    let cachedUser = null
    try {
      const stored = localStorage.getItem('kagibi_user')
      if (stored) cachedUser = JSON.parse(stored)
    } catch (e) {
      console.error('Failed to restore user from localStorage:', e)
    }
    return {
      isAuthenticated: false,
      user: cachedUser,
      masterKey: null,
      privateKey: null,
      publicKey: null,
      sessionTimeoutId: null,
      pendingMFAVerification: false,
    }
  },
  actions: {
    async ensureRSAKeys(masterKey) {
      if (!this.user) return
      await sodium.ready

      if (!this.user.public_key || !this.user.encrypted_private_key) {
        //console.log('Generating RSA keys for user...')
        const keyPair = await generateRSAKeyPair()
        const publicKeyPEM = await exportKeyToPEM(keyPair.publicKey, 'spki')
        const encryptedPrivateKey = await encryptPrivateKey(keyPair.privateKey, masterKey)
        await api.post('/users/keys', { public_key: publicKeyPEM, encrypted_private_key: encryptedPrivateKey })
        this.user.public_key = publicKeyPEM
        this.user.encrypted_private_key = encryptedPrivateKey
        this.privateKey = keyPair.privateKey
        this.publicKey = keyPair.publicKey
      } else {
        try {
          this.privateKey = await decryptPrivateKey(this.user.encrypted_private_key, masterKey)
          this.publicKey = await importKeyFromPEM(this.user.public_key, 'spki')
        } catch (e) {
          console.error('Failed to decrypt RSA Private Key:', e)
        }
      }
    },

    async login(credentials) {
      try {
        // 1. Authenticate with the current provider (Supabase or PocketBase)
        const { data, error } = await authClient.signIn(credentials.email, credentials.password)
        if (error) throw error

        // 2. Fetch encryption keys from the backend (token injected by api.js interceptor)
        let keysResponse
        try {
          keysResponse = await api.get('/auth/keys')
        } catch (err) {
          console.error('Failed to fetch keys from backend:', err)
          throw new Error('Impossible de récupérer les clés de chiffrement du serveur.')
        }

        const { salt, encrypted_master_key } = keysResponse.data

        if (salt && encrypted_master_key) {
          try {
            await sodium.ready
            const saltBytes = sodium.from_hex(salt)
            const kek = await deriveKeyFromPassword(credentials.password, saltBytes)
            this.masterKey = await unwrapMasterKey(encrypted_master_key, kek)
            this.setupSessionTimeout()
            await this.fetchUser()
            await this.ensureRSAKeys(this.masterKey)
          } catch (decryptError) {
            console.error('Decryption failed during login:', decryptError)
            throw new Error('Impossible de déchiffrer vos clés. Votre mot de passe est-il correct ?')
          }
        } else {
          this.masterKey = null
          throw new Error('Pas de clés de chiffrement trouvées sur le serveur. Veuillez contacter le support.')
        }

        this.isAuthenticated = true
        await this.fetchUser()
        this.persistUserToStorage()

        // Check if MFA is required (Supabase only)
        if (authClient.isMFASupported) {
          const mfa = useMFA()
          try {
            const mfaRequired = await mfa.isMFARequired('login')
            if (mfaRequired) {
              //console.log('[Auth] MFA required for login, setting pending state')
              this.pendingMFAVerification = true
              return 'mfa_required'
            }
          } catch (mfaErr) {
            console.warn('[Auth] Failed to check MFA requirement:', mfaErr)
          }
        }

        router.push(isP2PSubdomain ? '/' : { name: 'Home' })
        return true
      } catch (error) {
        console.error('Login failed:', error)
        this.isAuthenticated = false
        this.user = null
        throw error
      }
    },

    async register(username, email, password, avatarUrl = '/avatars/default.png', encryptFilenames = false) {
      // Offload all key generation to a Web Worker to keep UI responsive during Argon2id (~1s)
      const WORKER_TIMEOUT_MS = 30000
      const keyMaterial = await new Promise((resolve, reject) => {
        const worker = new Worker(
          new URL('../workers/registration.worker.js', import.meta.url),
          { type: 'module' }
        )
        let settled = false
        const settle = (fn, value) => {
          if (settled) return
          settled = true
          worker.terminate()
          fn(value)
        }
        const timeoutId = setTimeout(() => {
          settle(reject, new Error('La génération des clés a expiré. Veuillez réessayer.'))
        }, WORKER_TIMEOUT_MS)

        worker.onmessage = ({ data }) => {
          clearTimeout(timeoutId)
          if (data.type === 'REGISTER_KEYS_RESULT') {
            settle(resolve, data.payload)
          } else if (data.type === 'REGISTER_KEYS_ERROR') {
            settle(reject, new Error('Erreur lors de la génération des clés. Veuillez réessayer.'))
          }
        }
        worker.onerror = (err) => {
          clearTimeout(timeoutId)
          settle(reject, new Error(err.message || 'Erreur lors de la génération des clés. Veuillez réessayer.'))
        }
        worker.postMessage({ type: 'REGISTER_KEYS', password })
      })

      const {
        saltHex,
        wrappedMasterKey,
        wrappedMasterKeyRecovery,
        recoveryHash,
        recoveryCode,
        publicKeyPEM,
        encryptedPrivateKey,
        masterKeyRaw,
      } = keyMaterial

      try {
        // Re-import master key from raw bytes on the main thread
        const masterKey = await window.crypto.subtle.importKey(
          'raw',
          masterKeyRaw,
          { name: 'AES-GCM' },
          true,
          ['encrypt', 'decrypt']
        )

        // 1. Create account in the auth provider (Supabase or PocketBase)
        const { data, error } = await authClient.signUp(email, password, {
          data: { name: username }
        })
        if (error) throw error

        if (!data.session) {
          throw new Error(
            IS_POCKETBASE
              ? 'Erreur: Impossible de créer une session après l\'inscription PocketBase.'
              : 'L\'inscription nécessite que la confirmation d\'email soit DÉSACTIVÉE dans Supabase. Les clés de chiffrement générées ne peuvent pas être sauvegardées sans session active.'
          )
        }

        // 2. Create the encrypted profile on the backend
        const accessToken = data.session.access_token
        const config = accessToken ? { headers: { Authorization: `Bearer ${accessToken}` } } : {}

        try {
          await api.post('/auth/register', {
            name: username,
            email,
            avatar_url: avatarUrl,
            salt: saltHex,
            encrypted_master_key: wrappedMasterKey,
            encrypted_master_key_recovery: wrappedMasterKeyRecovery,
            recovery_hash: recoveryHash,
            recovery_salt: saltHex,
            public_key: publicKeyPEM,
            encrypted_private_key: encryptedPrivateKey,
            encrypt_filenames: encryptFilenames,
          }, config)
        } catch (backendErr) {
          // Auth provider account was created but backend profile failed.
          // Surface a specific error so the user can contact support.
          // The auth provider account exists; retrying with the same email will fail signUp.
          const msg = backendErr.response?.data?.error || backendErr.message || 'Erreur serveur'
          throw new Error(
            `Votre compte a été créé mais la configuration a échoué (${msg}). ` +
            'Veuillez contacter le support en indiquant votre adresse e-mail.'
          )
        }

        // 3. Set auth state only after all backend operations succeed
        await this.fetchUser()
        this.masterKey = masterKey
        this.isAuthenticated = true
        this.setupSessionTimeout()

      } catch (err) {
        if (err.message) throw new Error(err.message)
        else throw new Error('Erreur lors de l\'inscription')
      }
      return recoveryCode
    },

    async logout() {
      try {
        await api.post('/auth/logout')
        await authClient.signOut()
      } catch (error) {
        console.error('Logout failed:', error)
      } finally {
        useFriendStore().cleanup()
        this.isAuthenticated = false
        this.user = null
        this.masterKey = null
        localStorage.removeItem('kagibi_user')
        router.push({ name: 'Login' })
      }
    },

    async deleteAccount(confirmation) {
      if (!this.isAuthenticated) throw new Error('Non authentifié')
      if (confirmation !== 'SUPPRIMER') throw new Error("Confirmation invalide. Tapez 'SUPPRIMER' pour confirmer.")

      try {
        const response = await api.delete('/auth/account', { data: { confirmation: 'SUPPRIMER' } })
        this.masterKey = null
        this.privateKey = null
        this.publicKey = null
        this.user = null
        this.isAuthenticated = false
        localStorage.removeItem('kagibi_user')
        //console.log('[RGPD] ✅ Account deleted successfully')
        return response.data
      } catch (error) {
        console.error('[RGPD] Account deletion failed:', error)
        throw new Error(error.response?.data?.error || 'Erreur lors de la suppression du compte')
      }
    },

    async fetchUser() {
      try {
        const token = await authClient.getToken()
        const headers = token ? { Authorization: `Bearer ${token}` } : {}
        const response = await api.get('/users/me', { headers })
        this.user = response.data
        this.persistUserToStorage()
      } catch (error) {
        console.error('Failed to fetch user:', error)
      }
    },

    async updatePassword(currentPassword, newPassword) {
      if (!this.masterKey) throw new Error('Master key not available. Please re-login.')
      await sodium.ready

      try {
        // Update password in the auth provider
        await authClient.updateUser({
          password: newPassword,
          oldPassword: currentPassword, // Required by PocketBase; ignored by Supabase
        })

        // Update the encrypted master key on the backend (crypto key depends on password)
        const newSalt = generateSalt()
        const newSaltHex = sodium.to_hex(newSalt)
        const newKek = await deriveKeyFromPassword(newPassword, newSalt)
        const newEncryptedMasterKey = await wrapMasterKey(this.masterKey, newKek)

        await api.post('/users/change-password', {
          new_salt: newSaltHex,
          new_encrypted_master_key: newEncryptedMasterKey
        })
      } catch (error) {
        console.error('Password update failed:', error)
        throw error
      }
    },

    async updateUsername(newName) {
      if (!newName || newName.trim().length === 0) throw new Error('Le nom d\'utilisateur ne peut pas être vide.')

      try {
        // Update metadata in the auth provider
        await authClient.updateUser({ data: { name: newName.trim() } })

        // Update our backend profile
        const response = await api.put('/users/profile', { name: newName.trim() })
        this.user = response.data
        this.persistUserToStorage()
        return response.data
      } catch (error) {
        console.error('Username update failed:', error)
        throw error
      }
    },

    async checkAuth() {
      if (isLocalAuthBypassEnabled) {
        this.isAuthenticated = true
        if (!this.user) {
          this.user = { ...localDevBypassUser }
          this.persistUserToStorage()
        }
        return true
      }

      const { data: { session } } = await authClient.getSession()

      if (session?.access_token) {
        // SECURITY: MasterKey is NOT persisted — user must re-login after page reload
        if (!this.masterKey) {
          console.warn('Session found but MasterKey missing (page reload or new tab). Redirecting to login.')
          return false
        }

        this.isAuthenticated = true

        if (!this.user || this.user.id !== session.user?.id) {
          await this.fetchUser()
        }
        if (this.user && this.masterKey) {
          await this.ensureRSAKeys(this.masterKey)
        }
      } else {
        this.isAuthenticated = false
        this.user = null
        this.masterKey = null
      }
      return this.isAuthenticated
    },

    async recoverAccount(email, recoveryCode, newPassword) {
      await sodium.ready

      const initResponse = await api.post('/auth/recovery/init', { email })
      const { encrypted_master_key_recovery, salt } = initResponse.data

      if (!encrypted_master_key_recovery) throw new Error('Recovery not available for this account.')

      const saltBytes = sodium.from_hex(salt)
      const recoveryKek = await deriveKeyFromRecoveryCode(recoveryCode, saltBytes)

      let masterKey
      try {
        masterKey = await unwrapMasterKey(encrypted_master_key_recovery, recoveryKek)
      } catch (e) {
        throw new Error('Invalid recovery code.')
      }

      const newSalt = generateSalt()
      const newSaltHex = sodium.to_hex(newSalt)
      const newKek = await deriveKeyFromPassword(newPassword, newSalt)
      const newEncryptedMasterKey = await wrapMasterKey(masterKey, newKek)
      const recoveryHash = await hashRecoveryCode(recoveryCode)

      await api.post('/auth/recovery/finish', {
        email,
        recovery_hash: recoveryHash,
        new_password: newPassword,
        new_salt: newSaltHex,
        new_encrypted_master_key: newEncryptedMasterKey
      })

      this.setupSessionTimeout()
      return true
    },

    persistUserToStorage() {
      if (this.user) {
        try {
          localStorage.setItem('kagibi_user', JSON.stringify(this.user))
        } catch (e) {
          console.error('Failed to persist user data to localStorage', e)
        }
      }
    },

    updateUserStorage(storageUsed, storageLimit = undefined) {
      if (!this.user || storageUsed === undefined) return
      this.user = {
        ...this.user,
        storage_used: storageUsed,
        ...(storageLimit !== undefined ? { storage_limit: storageLimit } : {})
      }
      this.persistUserToStorage()
    },

    restoreUserFromStorage() {
      try {
        const storedUser = localStorage.getItem('kagibi_user')
        if (storedUser) {
          this.user = JSON.parse(storedUser)
          return true
        }
      } catch (e) {
        console.error('Failed to restore user data from localStorage', e)
      }
      return false
    },

    setupSessionTimeout() {
      if (this.sessionTimeoutId) clearTimeout(this.sessionTimeoutId)
      const THIRTY_MINUTES = 30 * 60 * 1000
      this.sessionTimeoutId = setTimeout(() => {
        console.warn('[Security] Session timeout - logging out')
        this.logout()
      }, THIRTY_MINUTES)
      //console.log('[Security] Session timeout set to 30 minutes')
    },

    async updateAvatar(avatarUrl) {
      try {
        await api.put('/users/avatar', { avatar_url: avatarUrl })
        if (this.user) {
          this.user.avatar_url = avatarUrl
          this.persistUserToStorage()
        }
      } catch (error) {
        console.error('Failed to update avatar:', error)
        throw error
      }
    }
  },
})
