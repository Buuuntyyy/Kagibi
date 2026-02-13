import axios from 'axios'
import { supabase } from './supabase'

// Use runtime configuration from window.__APP_CONFIG__ 
// which is injected by nginx from Kubernetes environment variables
export const API_BASE_URL = (
  typeof window !== 'undefined' && window.__APP_CONFIG__?.apiUrl
) ? window.__APP_CONFIG__.apiUrl : (import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1')

const api = axios.create({
  baseURL: API_BASE_URL,
  withCredentials: true,
})

// Add a request interceptor to inject the Supabase Token
api.interceptors.request.use(async (config) => {
  const { data: { session } } = await supabase.auth.getSession()
  
  if (session?.access_token) {
    config.headers.Authorization = `Bearer ${session.access_token}`
  }
  
  return config
}, (error) => {
  return Promise.reject(error)
})

export default api
