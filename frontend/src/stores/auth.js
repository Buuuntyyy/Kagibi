import { defineStore } from 'pinia'
import api from '../api'
import router from '../router'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    isAuthenticated: false,
    user: null,
  }),
  actions: {
    async login(credentials) {
      try {
        await api.post('/auth/login', credentials, {
          headers: {
            'Content-Type': 'application/json'
          }
        });

        this.isAuthenticated = true;

        await this.fetchUser();
        router.push({ name: 'Dashboard' });
        return true
      } catch (error) {
        console.error("Login failed:", error)
        this.isAuthenticated = false;
        this.user = null;
        return false
      }
    },
    async register(username, email, password) {
      await api.post('/auth/register', { name: username, email: email, password: password })
    },
    async logout() {
      try {
        await api.post('/auth/logout');
      } catch (error) {
        console.error("Logout failed:", error)
      } finally {
        this.isAuthenticated = false;
        this.user = null;
        router.push({ name: 'Login' });
      }
    },
    async checkAuth() {
      try {
        const response = await api.get('/users/me');
        this.isAuthenticated = true;
        this.user = response.data;
        return true

      } catch (error) {
        this.isAuthenticated = false;
        this.user = null;
        return false
      }
    },
    async fetchUser() {
      try {
        const response = await api.get('/users/me');
        this.user = response.data;
      } catch (error) {
        console.error("Failed to fetch user:", error)
        this.logout();
      }
    },
  },
})
