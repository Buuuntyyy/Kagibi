/**
 * Runtime Configuration
 * Injected by Kubernetes environment variables
 */
window.__APP_CONFIG__ = {
  apiUrl: window.__API_URL__ || 'http://localhost:8080/api/v1',
  supabaseUrl: window.__SUPABASE_URL__ || '',
  supabaseAnonKey: window.__SUPABASE_ANON_KEY__ || '',
}
