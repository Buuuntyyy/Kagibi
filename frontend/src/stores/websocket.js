import { defineStore } from 'pinia'
import { useAuthStore } from './auth'

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

      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const host = 'localhost:8080'; // En prod, utiliser window.location.host ou une variable d'env
      const url = `${protocol}//${host}/ws`;

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

    handleMessage(message) {
      const authStore = useAuthStore();
      
      switch (message.type) {
        case 'storage_update':
          console.log('Storage update received:', message.payload);
          if (authStore.user) {
            authStore.user.storage_used = message.payload.storage_used;
          }
          break;
        default:
          console.warn('Unknown WebSocket message type:', message.type);
      }
    }
  }
})
