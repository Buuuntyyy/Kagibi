import { createClient } from '@supabase/supabase-js'

// Use runtime configuration from window.__APP_CONFIG__ (injected by nginx)
// or fallback to import.meta.env (for local dev builds)
const supabaseUrl = (
  typeof window !== 'undefined' && window.__APP_CONFIG__?.supabaseUrl
) ? window.__APP_CONFIG__.supabaseUrl : (import.meta.env.VITE_SUPABASE_URL || 'https://msshzlznpgrvdnowefbb.supabase.co')

const supabaseKey = (
  typeof window !== 'undefined' && window.__APP_CONFIG__?.supabaseAnonKey
) ? window.__APP_CONFIG__.supabaseAnonKey : (import.meta.env.VITE_SUPABASE_ANON_KEY || 'sb_publishable_zz8DVqEf2Ewr5GL8cOzvnw_Il3LTMBc')

export const supabase = createClient(supabaseUrl, supabaseKey)
