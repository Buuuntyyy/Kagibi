import { defineStore } from 'pinia'
import api from '../api'

export const useFriendStore = defineStore('friends', {
  state: () => ({
    friends: [],
    loading: false,
    error: null
  }),

  getters: {
    acceptedFriends: (state) => state.friends.filter(f => f.status === 'accepted'),
    pendingSent: (state) => state.friends.filter(f => f.status === 'pending_sent'),
    pendingReceived: (state) => state.friends.filter(f => f.status === 'pending_received')
  },

  actions: {
    async fetchFriends() {
      this.loading = true
      try {
        const response = await api.get('/friends')
        this.friends = response.data || []
      } catch (err) {
        this.error = err.message
      } finally {
        this.loading = false
      }
    },

    async sendRequest(friendCode) {
      try {
        await api.post('/friends', { friendCode })
        await this.fetchFriends() // Refresh list
        return true
      } catch (err) {
        throw err.response?.data?.error || "Erreur lors de l'envoi"
      }
    },

    async acceptRequest(requestId) {
      try {
        await api.put(`/friends/${requestId}/accept`)
        await this.fetchFriends()
      } catch (err) {
        console.error(err)
      }
    },

    async rejectRequest(requestId) {
      try {
        await api.delete(`/friends/${requestId}/reject`)
        await this.fetchFriends()
      } catch (err) {
        console.error(err)
      }
    },

    async removeFriend(friendId) {
      try {
        await api.delete(`/friends/${friendId}`)
        await this.fetchFriends()
      } catch (err) {
        console.error(err)
      }
    }
  }
})
