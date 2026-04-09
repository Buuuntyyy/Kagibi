import { defineStore } from 'pinia'
import api from '../api'
import { useRealtimeStore } from './realtime'

export const useFriendStore = defineStore('friends', {
  state: () => ({
    friends: [],
    loading: false,
    error: null,
    _unsubscribers: []
  }),

  getters: {
    acceptedFriends: (state) => state.friends.filter(f => f.status === 'accepted'),
    pendingSent:     (state) => state.friends.filter(f => f.status === 'pending_sent'),
    pendingReceived: (state) => state.friends.filter(f => f.status === 'pending_received')
  },

  actions: {
    async fetchFriends() {
      this.loading = true
      try {
        const response = await api.get('/friends')
        // API response already includes correct `online` field from hub state
        this.friends = response.data || []
        this._subscribeRealtime()
      } catch (err) {
        this.error = err.message
      } finally {
        this.loading = false
      }
    },

    // Subscribe to realtime events once; safe to call multiple times.
    _subscribeRealtime() {
      if (this._unsubscribers.length > 0) return

      const realtimeStore = useRealtimeStore()

      const unsubFriend = realtimeStore.onEvent('friend_update', () => {
        this.fetchFriends()
      })

      const unsubPresence = realtimeStore.onEvent('presence_update', (payload) => {
        this.updatePresence(payload)
      })

      this._unsubscribers = [unsubFriend, unsubPresence]
    },

    // Unsubscribe from realtime events (call on sign-out).
    cleanup() {
      this._unsubscribers.forEach(fn => fn())
      this._unsubscribers = []
    },

    async sendRequest(friendCode) {
      try {
        await api.post('/friends', { friendCode })
        await this.fetchFriends()
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
    },

    updatePresence(payload) {
      const friend = this.friends.find(f => f.id === payload.user_id)
      if (friend) {
        friend.online = payload.online
      }
    }
  }
})
