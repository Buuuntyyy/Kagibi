import axios from 'axios'
import { useAuthStore } from './stores/auth'

const api = axios.create({
  baseURL: 'http://localhost:8080/api/v1',
  withCredentials: true,
})

export default api
