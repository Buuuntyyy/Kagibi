import { createClient } from '@supabase/supabase-js'

// Use runtime configuration from window.__APP_CONFIG__ (injected by nginx)
// or fallback to import.meta.env (for local dev builds)
const supabaseUrl = (
  typeof window !== 'undefined' && window.__APP_CONFIG__?.supabaseUrl
) ? window.__APP_CONFIG__.supabaseUrl : import.meta.env.VITE_SUPABASE_URL

const supabaseKey = (
  typeof window !== 'undefined' && window.__APP_CONFIG__?.supabaseAnonKey
) ? window.__APP_CONFIG__.supabaseAnonKey : import.meta.env.VITE_SUPABASE_ANON_KEY

// Return null instead of throwing when Supabase is not configured (e.g. PocketBase mode).
// Consumers should use authClient from auth-client.js instead of importing supabase directly.
export const supabase = (supabaseUrl && supabaseKey)
  ? createClient(supabaseUrl, supabaseKey)
  : null
