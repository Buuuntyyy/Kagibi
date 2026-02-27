import { createClient } from '@supabase/supabase-js'

// Use runtime configuration from window.__APP_CONFIG__ (injected by nginx)
// or fallback to import.meta.env (for local dev builds)
const supabaseUrl = (
  typeof window !== 'undefined' && window.__APP_CONFIG__?.supabaseUrl
) ? window.__APP_CONFIG__.supabaseUrl : import.meta.env.VITE_SUPABASE_URL

const supabaseKey = (
  typeof window !== 'undefined' && window.__APP_CONFIG__?.supabaseAnonKey
) ? window.__APP_CONFIG__.supabaseAnonKey : import.meta.env.VITE_SUPABASE_ANON_KEY

if (!supabaseUrl || !supabaseKey) {
  throw new Error('Supabase configuration is missing. Set SUPABASE_URL/VITE_SUPABASE_URL and SUPABASE_ANON_KEY/VITE_SUPABASE_ANON_KEY.')
}

export const supabase = createClient(supabaseUrl, supabaseKey)
