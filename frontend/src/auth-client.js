/**
 * auth-client.js — Local authentication client
 *
 * Handles all authentication against the SaferCloud backend local auth stack.
 * The multi-provider abstraction (Supabase/PocketBase) has been removed following
 * the completion of the Supabase → local auth migration.
 *
 * Configure the API base URL via:
 *   - Build time: VITE_API_URL
 *   - Runtime (Docker/K8s): window.__APP_CONFIG__.apiUrl
 */

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

// ── Token storage ─────────────────────────────────────────────────────────────

const LOCAL_TOKEN_KEY = 'safercloud_local_token'

function _localSaveToken(token) {
  if (token) localStorage.setItem(LOCAL_TOKEN_KEY, token)
  else localStorage.removeItem(LOCAL_TOKEN_KEY)
}

function _localGetToken() {
  const token = localStorage.getItem(LOCAL_TOKEN_KEY)
  if (!token) return null
  try {
    const payload = JSON.parse(atob(token.split('.')[1]))
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
    const payload = JSON.parse(atob(token.split('.')[1]))
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
    const data = await _localFetch('/auth/login', { email, password })
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
    const data = await _localFetch('/auth/signup', { email, password })
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
    const token = _localGetToken()
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
    return _localGetToken()
  },

  async updateUser(updates) {
    if (updates.password) {
      const token = _localGetToken()
      return _localFetch('/auth/update-password', {
        old_password: updates.oldPassword,
        new_password: updates.password,
      }, token)
    }
    // Name/metadata updates are handled by the backend /users/profile endpoint
    return {}
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
