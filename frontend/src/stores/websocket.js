import { defineStore } from 'pinia'
import { useAuthStore } from './auth'
import { useFriendStore } from './friends'
import { useFileStore } from './files'

export const useWebSocketStore = defineStore('websocket', {
  state: () => ({
    socket: null,
    isConnected: false,
    reconnectInterval: 1000,
    maxReconnectInterval: 30000,
  }),
  actions: {
    connect() {
      if (this.socket && (this.socket.readyState === WebSocket.OPEN || this.socket.readyState === WebSocket.CONNECTING)) {
        return;
      }

      let url;
      // 1. Priorité : Variable d'environnement explicite pour le WebSocket
      if (import.meta.env.VITE_WS_URL) {
          url = import.meta.env.VITE_WS_URL;
      } else {
          // 2. Fallback Dev : Localhost
          const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
          const host = 'localhost:8080'; 
          url = `${protocol}//${host}/ws`;
      }

      console.log(`Connecting to WebSocket at ${url}...`);
      
      this.socket = new WebSocket(url);

      this.socket.onopen = () => {
        console.log('WebSocket connected');
        this.isConnected = true;
        this.reconnectInterval = 1000; // Reset reconnect interval
      };

      this.socket.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data);
          this.handleMessage(message);
        } catch (e) {
          console.error('WebSocket message error:', e);
        }
      };

      this.socket.onclose = (event) => {
        console.log('WebSocket disconnected', event.code, event.reason);
        this.isConnected = false;
        this.socket = null;

        // Attempt reconnect if not closed cleanly
        if (event.code !== 1000) {
          setTimeout(() => {
            this.reconnectInterval = Math.min(this.reconnectInterval * 2, this.maxReconnectInterval);
            this.connect();
          }, this.reconnectInterval);
        }
      };

      this.socket.onerror = (error) => {
        console.error('WebSocket error:', error);
        this.socket.close();
      };
    },

    disconnect() {
      if (this.socket) {
        this.socket.close(1000, "User logged out");
        this.socket = null;
        this.isConnected = false;
      }
    },

    async handleMessage(message) {
      const authStore = useAuthStore();
      
      switch (message.type) {
        case 'storage_update':
          console.log('Storage update received:', message.payload);
          if (authStore.user && message.payload.storage_used !== undefined) {
             authStore.user.storage_used = message.payload.storage_used;
          }
          // If action indicates share update, refetch files
          if (message.payload.action === 'share_created' || message.payload.action === 'share_revoked' || message.payload.action === 'share_received' || message.payload.action === 'share_revoked_by_recipient') {
             // We need to refresh the current file list if we are inspecting files
             import('./files').then(m => {
                 const fs = m.useFileStore();
                 fs.fetchItems(fs.currentPath);
                 fs.notifyShareUpdate();
             });
          }
          break;
        case 'friend_update':
          console.log('Friend update received:', message.payload);
          const friendStore = useFriendStore();
          friendStore.fetchFriends();
          break;
        case 'presence_update':
          // Payload: { user_id, online }
          // We can update FriendStore directly
          const fs = useFriendStore();
          fs.updatePresence(message.payload);
          break;
        case 'p2p_signal':
          // { sender_id, type, data }
          import('./p2p').then(m => {
              const p2pStore = m.useP2PStore();
              p2pStore.handleSignal(message.payload);
          });
          break;
        default:
          console.warn('Unknown WebSocket message type:', message.type);
      }
    },
    
    sendSignal(targetId, type, data) {
        if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
            console.error("Cannot send signal: WebSocket not connected");
            return;
        }
        this.socket.send(JSON.stringify({
            type: "p2p_signal",
            payload: {
                target_id: targetId,
                type: type,
                data: data
            }
        }));
    }
  }
})
