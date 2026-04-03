import PocketBase from 'pocketbase'

// Use runtime configuration from window.__APP_CONFIG__ (injected by nginx)
// or fallback to import.meta.env (for local dev builds)
const pbUrl = (
  typeof window !== 'undefined' && window.__APP_CONFIG__?.pocketbaseUrl
) ? window.__APP_CONFIG__.pocketbaseUrl : import.meta.env.VITE_POCKETBASE_URL

// Return null if PocketBase URL is not configured (e.g. Supabase mode)
export const pb = pbUrl ? new PocketBase(pbUrl) : null
