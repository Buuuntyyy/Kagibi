import { defineStore } from 'pinia'
import api from '../api'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('token') || null,
    user: null,
  }),
  getters: {
    isAuthenticated: (state) => !!state.token,
  },
  actions: {
    async login(email, password) {
      try {
        const response = await api.post('/auth/login', { email, password })
        this.token = response.data.token
        localStorage.setItem('token', this.token)
        this.user = { email } 
        return true
      } catch (error) {
        console.error("Login failed:", error)
        return false
      }
    },
    async register(username, email, password) {
      await api.post('/auth/register', { name: username, email: email, password: password })
    },
    logout() {
      this.token = null
      this.user = null
      localStorage.removeItem('token')
    },
    async checkAuth() {
      if (!this.token) return false
      try {
        await api.get('/users')
        return true
      } catch (error) {
        this.logout()
        return false
      }
    },
  },
})
