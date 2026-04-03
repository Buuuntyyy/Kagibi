import axios from 'axios'
import { authClient } from './auth-client'

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
    return Promise.reject(error)
  }
)

export default api
