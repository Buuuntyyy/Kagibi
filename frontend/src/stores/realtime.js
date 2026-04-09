// frontend/src/stores/realtime.js
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { authClient } from '../auth-client'

export const useRealtimeStore = defineStore('realtime', () => {
  // State
  const isConnected = ref(false)
  const lastEventId = ref(null)
  const presenceState = ref({})

  // Polling timer (fallback for WebSocket reconnect gaps)
  let _pollP2PTimer = null
  const _seenSignalIds = new Set()

  // WebSocket state
  let _ws = null
  let _wsReconnectTimer = null
  let _wsReconnectDelay = 1000 // ms, doubles on each failure up to 30 s
  let _wsIntentionalClose = false

  // Event handlers registered by other stores/components
  const eventHandlers = {
    storage_update: [],
    friend_update: [],
    p2p_signal: [],
    presence_update: []
  }

  function onEvent(eventType, handler) {
    if (eventHandlers[eventType]) {
      eventHandlers[eventType].push(handler)
    }
    return () => {
      const idx = eventHandlers[eventType]?.indexOf(handler)
      if (idx > -1) eventHandlers[eventType].splice(idx, 1)
    }
  }

  function dispatchEvent(eventType, payload) {
    //console.log(`[Realtime] Dispatching ${eventType}:`, payload)
    eventHandlers[eventType]?.forEach(handler => {
      try { handler(payload) } catch (e) {
        console.error(`[Realtime] Error in ${eventType} handler:`, e)
      }
    })
  }

  // ── WebSocket ────────────────────────────────────────────────────────────────

  function _wsUrl() {
    const apiUrl = (
      typeof window !== 'undefined' && window.__APP_CONFIG__?.apiUrl
    ) || import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1'
    const base = apiUrl.replace(/\/$/, '').replace(/^http/, 'ws')
    return `${base}/ws`
  }

  async function _wsUrlWithToken() {
    // Prefer single-use ws_token (more proxy-friendly than Sec-WebSocket-Protocol trick)
    try {
      const apiUrl = (
        typeof window !== 'undefined' && window.__APP_CONFIG__?.apiUrl
      ) || import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1'
      const base = apiUrl.replace(/\/$/, '')
      const token = await authClient.getToken()
      if (!token) return null

      const res = await fetch(`${base}/auth/ws-token`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` }
      })
      if (res.ok) {
        const { token: wsToken } = await res.json()
        return `${_wsUrl()}?ws_token=${wsToken}`
      }
    } catch (e) {
      console.warn('[Realtime] Failed to get ws_token, falling back to Sec-WebSocket-Protocol:', e)
    }
    return null // caller falls back to legacy method
  }

  function _connectWS() {
    if (_ws && (_ws.readyState === WebSocket.CONNECTING || _ws.readyState === WebSocket.OPEN)) return
    _wsIntentionalClose = false
    _openWebSocket()
  }

  function _openWebSocket() {
    if (_ws && (_ws.readyState === WebSocket.CONNECTING || _ws.readyState === WebSocket.OPEN)) return

    _wsUrlWithToken().then(urlWithToken => {
      if (_wsIntentionalClose) return

      let ws
      if (urlWithToken) {
        // ws_token path: clean URL, no subprotocol trick needed
        ws = new WebSocket(urlWithToken)
      } else {
        // Legacy path: encode JWT in Sec-WebSocket-Protocol header
        authClient.getToken().then(token => {
          if (!token || _wsIntentionalClose) return
          _ws = new WebSocket(_wsUrl(), ['token', token])
          _attachWsHandlers(_ws)
        })
        return
      }

      _ws = ws
      _attachWsHandlers(_ws)
    })
  }

  function _attachWsHandlers(ws) {
    ws.onopen = () => {
      //console.log('[Realtime] WebSocket connected')
      isConnected.value = true
      _wsReconnectDelay = 1000
      if (_wsReconnectTimer) { clearTimeout(_wsReconnectTimer); _wsReconnectTimer = null }
    }

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        _handleWSMessage(msg)
      } catch (e) {
        console.error('[Realtime] Failed to parse WebSocket message:', e)
      }
    }

    ws.onclose = (event) => {
      isConnected.value = false
      _ws = null
      if (_wsIntentionalClose) return
      console.warn(`[Realtime] WebSocket closed (code=${event.code}), reconnecting in ${_wsReconnectDelay}ms...`)
      _wsReconnectTimer = setTimeout(() => {
        _wsReconnectDelay = Math.min(_wsReconnectDelay * 2, 30000)
        _openWebSocket()
      }, _wsReconnectDelay)
    }

    ws.onerror = (err) => {
      console.error('[Realtime] WebSocket error:', err)
      // onclose fires next and handles reconnect
    }
  }

  function _handleWSMessage(msg) {
    if (msg.type === 'event') {
      if (msg.id) lastEventId.value = msg.id
      dispatchEvent(msg.event_type, msg.payload)
    } else if (msg.type === 'p2p_signal') {
      if (_seenSignalIds.has(msg.id)) return
      if (msg.id) _seenSignalIds.add(msg.id)
      dispatchEvent('p2p_signal', {
        from: msg.from,
        type: msg.signal_type,
        payload: msg.payload
      })
    } else if (msg.type === 'presence_update') {
      if (msg.online) {
        presenceState.value[msg.user_id] = { online: true, lastSeen: Date.now() }
      } else if (presenceState.value[msg.user_id]) {
        presenceState.value[msg.user_id].online = false
        presenceState.value[msg.user_id].lastSeen = Date.now()
      }
      dispatchEvent('presence_update', { user_id: msg.user_id, online: msg.online })
    }
  }

  function _disconnectWS() {
    _wsIntentionalClose = true
    if (_wsReconnectTimer) { clearTimeout(_wsReconnectTimer); _wsReconnectTimer = null }
    if (_ws) {
      _ws.close(1000, 'user disconnect')
      _ws = null
    }
    isConnected.value = false
  }

  // ── P2P polling fallback ─────────────────────────────────────────────────────
  // Kept as a resilience layer: catches signals delivered during WS reconnect gaps.

  async function pollP2PSignals() {
    try {
      const token = await authClient.getToken()
      if (!token) return []

      const response = await fetch('/api/v1/p2p/signals', {
        headers: { Authorization: `Bearer ${token}` }
      })
      if (!response.ok) return []

      const data = await response.json()
      const signals = data.signals || []

      signals.forEach(signal => {
        if (_seenSignalIds.has(signal.id)) return
        _seenSignalIds.add(signal.id)
        dispatchEvent('p2p_signal', {
          from: signal.sender_id,
          type: signal.signal_type,
          payload: typeof signal.payload === 'string' ? JSON.parse(signal.payload) : signal.payload
        })
      })
      return signals
    } catch (error) {
      console.error('[Realtime] Error polling P2P signals:', error)
      return []
    }
  }

  // ── Public API ──────────────────────────────────────────────────────────────

  async function connect() {
    const session = await authClient.getSession()
    const userId = session?.data?.session?.user?.id
    if (!userId) {
      console.warn('[Realtime] No user session, skipping connection')
      return
    }

    //console.log(`[Realtime] Connecting for user ${userId}`)

    _connectWS()

    // P2P polling fallback — catches signals during WS reconnect gaps
    if (!_pollP2PTimer) {
      _pollP2PTimer = setInterval(() => pollP2PSignals(), 2500)
    }
  }

  async function disconnect() {
    //console.log('[Realtime] Disconnecting...')

    if (_pollP2PTimer) { clearInterval(_pollP2PTimer); _pollP2PTimer = null }
    _seenSignalIds.clear()

    _disconnectWS()

    presenceState.value = {}
  }

  async function sendP2PSignal(targetUserId, signalType, payload) {
    try {
      const token = await authClient.getToken()
      const response = await fetch('/api/v1/p2p/signal', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`
        },
        body: JSON.stringify({ target_user_id: targetUserId, signal_type: signalType, payload })
      })
      if (!response.ok) throw new Error(`Failed to send P2P signal: ${response.status}`)
      return await response.json()
    } catch (error) {
      console.error('[Realtime] Error sending P2P signal:', error)
      throw error
    }
  }

  function isUserOnline(userId) {
    return presenceState.value[userId]?.online ?? false
  }

  return {
    isConnected,
    lastEventId,
    presenceState,
    connect,
    disconnect,
    onEvent,
    isUserOnline,
    sendP2PSignal,
    pollP2PSignals,
  }
})
