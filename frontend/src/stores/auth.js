import { defineStore } from 'pinia'
import api from '../api'
import router from '../router'
import { deriveKeyFromPassword, generateSalt } from '../utils/crypto'
import sodium from 'libsodium-wrappers-sumo'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    isAuthenticated: false,
    user: null,
  }),
  actions: {
    async login(credentials) {
      try {
        const authentication_response = await api.post('/auth/login', credentials, {
          headers: {
            'Content-Type': 'application/json'
          }
        });

        const { salt } = authentication_response.data;

        if (salt) {
          await sodium.ready;
          const saltBytes = sodium.from_hex(salt);

          this.masterKey = await deriveKeyFromPassword(credentials.password, saltBytes);
        } else {
          this.masterKey = null;
          console.warn("No salt received from server during login.");
        }

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
      await sodium.ready;
      const salt = generateSalt();
      const saltHex = sodium.to_hex(salt);

      const payload = {
        name: username,
        email: email,
        password: password,
        salt: saltHex
      };
      await api.post('/auth/register', payload)
    },
    async logout() {
      try {
        await api.post('/auth/logout');
      } catch (error) {
        console.error("Logout failed:", error)
      } finally {
        this.isAuthenticated = false;
        this.user = null;
        this.masterKey = null;
        router.push({ name: 'Login' });
      }
    },
    async checkAuth() {
      try {
        const response = await api.get('/users/me');
        this.isAuthenticated = true;
        this.user = response.data;
        // Note: Au refresh de la page, masterKey est perdu (c'est voulu pour la sécurité).
        // L'utilisateur devra peut-être se reconnecter ou retaper son mot de passe pour déchiffrer.
        return true

      } catch (error) {
        this.isAuthenticated = false;
        this.user = null;
        this.masterKey = null;
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
