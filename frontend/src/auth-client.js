/**
 * auth-client.js — Local authentication client
 *
 * Handles all authentication against the Kagibi backend local auth stack.
 * The multi-provider abstraction (Supabase/PocketBase) has been removed following
 * the completion of the Supabase → local auth migration.
 *
 * Configure the API base URL via:
 *   - Build time: VITE_API_URL
 *   - Runtime (Docker/K8s): window.__APP_CONFIG__.apiUrl
 */

import sodium from 'libsodium-wrappers-sumo'

export const AUTH_PROVIDER = 'local'
export const IS_LOCAL      = true
export const IS_POCKETBASE = false
export const IS_SUPABASE   = false

// Base URL for direct auth API calls (avoids circular import with api.js)
const _apiBase = (() => {
  const url = (
    typeof window !== 'undefined' && window.__APP_CONFIG__?.apiUrl
  ) ? window.__APP_CONFIG__.apiUrl : (import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1')
  return url.replace(/\/$/, '')
})()

// ── Guest token (in-memory, never persisted) ──────────────────────────────────
// Set during a P2P guest session. Takes priority over localStorage token.

let _guestToken = null

// ── Token storage ─────────────────────────────────────────────────────────────

const LOCAL_TOKEN_KEY = 'kagibi_local_token'

function _localSaveToken(token) {
  if (token) localStorage.setItem(LOCAL_TOKEN_KEY, token)
  else localStorage.removeItem(LOCAL_TOKEN_KEY)
}

function _b64urlDecode(str) {
  // JWT uses base64url: replace URL-safe chars and restore padding
  const b64 = str.replace(/-/g, '+').replace(/_/g, '/')
  const pad = (4 - b64.length % 4) % 4
  return atob(b64 + '='.repeat(pad))
}

function _localGetToken() {
  const token = localStorage.getItem(LOCAL_TOKEN_KEY)
  if (!token) return null
  try {
    const payload = JSON.parse(_b64urlDecode(token.split('.')[1]))
    if (payload.exp * 1000 < Date.now()) {
      localStorage.removeItem(LOCAL_TOKEN_KEY)
      return null
    }
    return token
  } catch {
    localStorage.removeItem(LOCAL_TOKEN_KEY)
    return null
  }
}

function _localDecodeUser(token) {
  try {
    const payload = JSON.parse(_b64urlDecode(token.split('.')[1]))
    return { id: payload.sub, email: payload.email, aal: payload.aal || 'aal1' }
  } catch {
    return null
  }
}

async function _localFetch(path, body, token = null) {
  const headers = { 'Content-Type': 'application/json' }
  if (token) headers['Authorization'] = `Bearer ${token}`
  const res = await fetch(`${_apiBase}${path}`, {
    method: 'POST',
    headers,
    body: JSON.stringify(body),
  })
  const data = await res.json()
  if (!res.ok) throw new Error(data.error || `HTTP ${res.status}`)
  return data
}

// ── Solution A1 : secret d'authentification dérivé ─────────────────────────────
// Le mot de passe BRUT n'est jamais envoyé au backend. On envoie à la place
// authPassword = Argon2id(motDePasse, sel d'app fixe). À sens unique : le serveur ne voit
// jamais le mot de passe en clair. La dérivation du KEK de CHIFFREMENT (côté stores/auth.js)
// reste, elle, basée sur le mot de passe brut local — inchangée.
//
// ⚠️ DOIT rester strictement identique au desktop (Go crypto.DeriveAuthPassword) :
// mêmes paramètres Argon2id, même sel, même encodage base64. Vérifié par vecteurs de test.
async function _deriveAuthPassword(password) {
  await sodium.ready
  const authSalt = sodium.crypto_hash_sha256(sodium.from_string('kagibi-auth-pepper-v1')).slice(0, 16)
  const key = sodium.crypto_pwhash(
    32, password, authSalt,
    4,                 // OPSLIMIT (= Argon2 time)
    64 * 1024 * 1024,  // MEMLIMIT (= 64 MB)
    sodium.crypto_pwhash_ALG_ARGON2ID13,
  )
  return sodium.to_base64(key, sodium.base64_variants.ORIGINAL)
}

// _loginWithMigration se connecte avec le secret d'auth dérivé. Si le compte n'a pas encore
// été migré (hash backend = bcrypt(mot de passe brut)), il bascule de façon transparente
// le hash vers bcrypt(authPassword) puis renvoie une session fraîche. En cas d'échec de
// bascule, conserve la session legacy (aucune régression).
async function _loginWithMigration(email, password) {
  const authPwd = await _deriveAuthPassword(password)
  try {
    return await _localFetch('/auth/login', { email, password: authPwd })
  } catch (authErr) {
    let legacy
    try {
      legacy = await _localFetch('/auth/login', { email, password })
    } catch {
      throw authErr // vrai mauvais mot de passe
    }
    try {
      await _localFetch('/auth/update-password',
        { old_password: password, new_password: authPwd }, legacy.access_token)
    } catch {
      return legacy // bascule impossible → session legacy
    }
    return await _localFetch('/auth/login', { email, password: authPwd })
  }
}

// ── MFA helpers ───────────────────────────────────────────────────────────────

async function _mfaFetch(method, path, body = null) {
  const token = _localGetToken()
  const headers = { 'Content-Type': 'application/json' }
  if (token) headers['Authorization'] = `Bearer ${token}`
  const opts = { method, headers }
  if (body !== null) opts.body = JSON.stringify(body)
  const res = await fetch(`${_apiBase}${path}`, opts)
  const data = await res.json()
  if (!res.ok) return { data: null, error: { message: data.error || `HTTP ${res.status}` } }
  return { data, error: null }
}

const _mfa = {
  listFactors: async () => {
    const { data, error } = await _mfaFetch('GET', '/auth/mfa/factors')
    if (error) return { data: null, error }
    return { data: { totp: data.totp || [] }, error: null }
  },

  enroll: async ({ factorType, friendlyName }) => {
    return _mfaFetch('POST', '/auth/mfa/enroll', {
      factor_type: factorType,
      friendly_name: friendlyName,
    })
  },

  challenge: async ({ factorId }) => {
    return _mfaFetch('POST', '/auth/mfa/challenge', { factor_id: factorId })
  },

  verify: async ({ factorId, challengeId, code }) => {
    const result = await _mfaFetch('POST', '/auth/mfa/verify', {
      factor_id: factorId,
      challenge_id: challengeId,
      code,
    })
    if (result.data?.access_token) {
      _localSaveToken(result.data.access_token)
    }
    return result
  },

  unenroll: async ({ factorId }) => {
    return _mfaFetch('DELETE', '/auth/mfa/unenroll', { factor_id: factorId })
  },
}

// ── Unified auth interface ────────────────────────────────────────────────────

export const authClient = {
  provider: 'local',
  isMFASupported: true,

  async signIn(email, password) {
    const data = await _loginWithMigration(email, password)
    _localSaveToken(data.access_token)
    const user = _localDecodeUser(data.access_token) || data.user
    return {
      data: {
        session: {
          access_token: data.access_token,
          user: { id: user.id, email: user.email, aal: 'aal1' }
        }
      },
      error: null
    }
  },

  async signUp(email, password) {
    const authPwd = await _deriveAuthPassword(password)
    const data = await _localFetch('/auth/signup', { email, password: authPwd })
    _localSaveToken(data.access_token)
    const user = _localDecodeUser(data.access_token) || data.user
    return {
      data: {
        user: { id: user.id, email: user.email },
        session: {
          access_token: data.access_token,
          user: { id: user.id, email: user.email }
        }
      },
      error: null
    }
  },

  async signOut() {
    _localSaveToken(null)
  },

  async getSession() {
    const token = _guestToken || _localGetToken()
    if (!token) return { data: { session: null } }
    const user = _localDecodeUser(token)
    return {
      data: {
        session: {
          access_token: token,
          user: { id: user?.id, email: user?.email, aal: user?.aal || 'aal1' }
        }
      }
    }
  },

  async getToken() {
    if (_guestToken) return _guestToken
    return _localGetToken()
  },

  setGuestToken(token) {
    _guestToken = token
  },

  clearGuestToken() {
    _guestToken = null
  },

  isGuestSession() {
    return _guestToken !== null
  },

  async updateUser(updates) {
    if (updates.password) {
      const token = _localGetToken()
      const newAuth = await _deriveAuthPassword(updates.password)
      try {
        // Compte migré : ancien hash = bcrypt(authPwd(currentPassword)).
        return await _localFetch('/auth/update-password', {
          old_password: await _deriveAuthPassword(updates.oldPassword),
          new_password: newAuth,
        }, token)
      } catch {
        // Compte legacy : ancien hash = bcrypt(mot de passe brut) → repli.
        return await _localFetch('/auth/update-password', {
          old_password: updates.oldPassword,
          new_password: newAuth,
        }, token)
      }
    }
    // Name/metadata updates are handled by the backend /users/profile endpoint
    return {}
  },

  async updateEmail(newEmail, password) {
    const token = _localGetToken()
    const authPwd = await _deriveAuthPassword(password)
    const putEmail = async (pwd) => {
      const headers = { 'Content-Type': 'application/json' }
      if (token) headers['Authorization'] = `Bearer ${token}`
      const res = await fetch(`${_apiBase}/auth/update-email`, {
        method: 'PUT',
        headers,
        body: JSON.stringify({ new_email: newEmail, password: pwd }),
      })
      const data = await res.json()
      if (!res.ok) throw new Error(data.error || `HTTP ${res.status}`)
      return data
    }
    let data
    try {
      data = await putEmail(authPwd)           // compte migré
    } catch {
      data = await putEmail(password)          // compte legacy : repli mot de passe brut
    }
    if (data.access_token) _localSaveToken(data.access_token)
    return data
  },

  async refreshSession() {
    const token = _localGetToken()
    if (!token) return false
    try {
      const data = await _localFetch('/auth/refresh', {}, token)
      if (data.access_token) {
        _localSaveToken(data.access_token)
        return true
      }
      return false
    } catch {
      return false
    }
  },

  get mfa() {
    return _mfa
  },
}

export default authClient
