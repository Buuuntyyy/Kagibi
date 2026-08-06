import axios from 'axios'
import { authClient } from './auth-client'
import { requestStepUp } from './utils/mfaStepUp'

// Use runtime configuration from window.__APP_CONFIG__
// which is injected by nginx from Kubernetes environment variables
export const API_BASE_URL = (
  typeof window !== 'undefined' && window.__APP_CONFIG__?.apiUrl
) ? window.__APP_CONFIG__.apiUrl : (import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1')

const api = axios.create({
  baseURL: API_BASE_URL,
  withCredentials: true,
})

// Inject the auth token from the current provider (Supabase or PocketBase)
api.interceptors.request.use(async (config) => {
  const token = await authClient.getToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
}, (error) => {
  return Promise.reject(error)
})

// On 401, attempt a token refresh once and retry the original request.
// Handles both Supabase (silent refresh) and PocketBase (authRefresh) modes.
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const original = error.config
    if (error.response?.status === 401 && !original._retried) {
      original._retried = true
      const refreshed = await authClient.refreshSession()
      if (refreshed) {
        const token = await authClient.getToken()
        if (token) {
          original.headers.Authorization = `Bearer ${token}`
        }
        return api(original)
      }
    }

    // The account requires MFA for this action and the session is still aal1.
    // Prompt for step-up once, then replay the original request with the new
    // aal2 token. If the user cancels, the original error propagates unchanged.
    if (
      error.response?.status === 403 &&
      error.response?.data?.error === 'mfa_required' &&
      original &&
      !original._mfaRetried
    ) {
      original._mfaRetried = true
      try {
        await requestStepUp()
      } catch {
        throw error
      }
      const token = await authClient.getToken()
      if (token) {
        original.headers.Authorization = `Bearer ${token}`
      }
      return api(original)
    }

    throw error
  }
)

export default api
