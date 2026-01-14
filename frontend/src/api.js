import axios from 'axios'
import { useAuthStore } from './stores/auth'

export const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1'

const api = axios.create({
  baseURL: API_BASE_URL,
  withCredentials: true,
})

export default api
