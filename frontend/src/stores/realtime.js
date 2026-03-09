// frontend/src/stores/realtime.js
// Supabase Realtime store - replaces WebSocket store
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { supabase } from '../supabase'

export const useRealtimeStore = defineStore('realtime', () => {
  // State
  const isConnected = ref(false)
  const lastEventId = ref(null)
  const presenceState = ref({}) // { [userId]: { online: true, lastSeen: timestamp } }
  const p2pChannel = ref(null)
  const eventsChannel = ref(null)
  const presenceChannel = ref(null)

  // Polling fallback pour les signaux P2P (au cas où Supabase Realtime échoue)
  let _pollTimer = null
  const _seenSignalIds = new Set()

  // Event handlers - these will be set by components/stores that need to react
  const eventHandlers = {
    storage_update: [],
    friend_update: [],
    p2p_signal: []
  }

  // Register event handler
  function onEvent(eventType, handler) {
    if (eventHandlers[eventType]) {
      eventHandlers[eventType].push(handler)
    }
    // Return unsubscribe function
    return () => {
      const idx = eventHandlers[eventType]?.indexOf(handler)
      if (idx > -1) {
        eventHandlers[eventType].splice(idx, 1)
      }
    }
  }

  // Dispatch event to all registered handlers
  function dispatchEvent(eventType, payload) {
    console.log(`[Realtime] Dispatching ${eventType}:`, payload)
    eventHandlers[eventType]?.forEach(handler => {
      try {
        handler(payload)
      } catch (e) {
        console.error(`[Realtime] Error in ${eventType} handler:`, e)
      }
    })
  }

  // Initialize Supabase Realtime subscriptions
  async function connect() {
    const session = await supabase.auth.getSession()
    const userId = session?.data?.session?.user?.id
    
    if (!userId) {
      console.warn('[Realtime] No user session, skipping connection')
      return
    }

    console.log('[Realtime] Connecting for user:', userId)

    // 1. Subscribe to realtime_events table for this user (postgres_changes)
    eventsChannel.value = supabase
      .channel(`events:${userId}`)
      .on(
        'postgres_changes',
        {
          event: 'INSERT',
          schema: 'public',
          table: 'realtime_events',
          filter: `user_id=eq.${userId}`
        },
        (payload) => {
          console.log('[Realtime] Event received:', payload.new)
          const event = payload.new
          lastEventId.value = event.id
          dispatchEvent(event.event_type, event.payload)
        }
      )
      .subscribe((status) => {
        console.log('[Realtime] Events channel status:', status)
        isConnected.value = status === 'SUBSCRIBED'
      })

    // 2. Subscribe to P2P signals table for signaling
    p2pChannel.value = supabase
      .channel(`p2p:${userId}`)
      .on(
        'postgres_changes',
        {
          event: 'INSERT',
          schema: 'public',
          table: 'p2p_signals',
          filter: `target_id=eq.${userId}`
        },
        (payload) => {
          console.log('[Realtime] P2P signal received:', payload.new)
          const signal = payload.new
          if (_seenSignalIds.has(signal.id)) return
          _seenSignalIds.add(signal.id)
          dispatchEvent('p2p_signal', {
            from: signal.sender_id,
            type: signal.signal_type,
            payload: typeof signal.payload === 'string' ? JSON.parse(signal.payload) : signal.payload
          })
        }
      )
      .subscribe((status) => {
        console.log('[Realtime] P2P channel status:', status)
      })

    // 3. Initialize Presence channel for online status
    presenceChannel.value = supabase.channel('presence:global', {
      config: {
        presence: {
          key: userId
        }
      }
    })

    presenceChannel.value
      .on('presence', { event: 'sync' }, () => {
        const state = presenceChannel.value.presenceState()
        console.log('[Realtime] Presence sync:', state)
        updatePresenceState(state)
      })
      .on('presence', { event: 'join' }, ({ key, newPresences }) => {
        console.log('[Realtime] User joined:', key, newPresences)
        presenceState.value[key] = { online: true, lastSeen: Date.now() }
      })
      .on('presence', { event: 'leave' }, ({ key, leftPresences }) => {
        console.log('[Realtime] User left:', key)
        if (presenceState.value[key]) {
          presenceState.value[key].online = false
          presenceState.value[key].lastSeen = Date.now()
        }
      })
      .subscribe(async (status) => {
        console.log('[Realtime] Presence channel status:', status)
        if (status === 'SUBSCRIBED') {
          // Track own presence
          await presenceChannel.value.track({
            user_id: userId,
            online_at: new Date().toISOString()
          })
        }
      })

    // Polling fallback — s'assure que les signaux P2P sont reçus même si
    // Supabase Realtime ne fonctionne pas (table absente de la publication).
    if (!_pollTimer) {
      _pollTimer = setInterval(() => pollP2PSignals(), 2500)
    }
  }

  // Update presence state from sync event
  function updatePresenceState(state) {
    const newState = {}
    for (const [userId, presences] of Object.entries(state)) {
      if (presences && presences.length > 0) {
        newState[userId] = { online: true, lastSeen: Date.now() }
      }
    }
    presenceState.value = newState
  }

  // Check if a user is online
  function isUserOnline(userId) {
    return presenceState.value[userId]?.online ?? false
  }

  // Disconnect all channels
  async function disconnect() {
    console.log('[Realtime] Disconnecting...')

    if (_pollTimer) { clearInterval(_pollTimer); _pollTimer = null }
    _seenSignalIds.clear()

    if (eventsChannel.value) {
      await supabase.removeChannel(eventsChannel.value)
      eventsChannel.value = null
    }
    
    if (p2pChannel.value) {
      await supabase.removeChannel(p2pChannel.value)
      p2pChannel.value = null
    }
    
    if (presenceChannel.value) {
      await presenceChannel.value.untrack()
      await supabase.removeChannel(presenceChannel.value)
      presenceChannel.value = null
    }
    
    isConnected.value = false
    presenceState.value = {}
  }

  // Send P2P signal via API (inserts into p2p_signals table)
  async function sendP2PSignal(targetUserId, signalType, payload) {
    try {
      const response = await fetch('/api/v1/p2p/signal', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${(await supabase.auth.getSession()).data.session?.access_token}`
        },
        body: JSON.stringify({
          target_user_id: targetUserId,
          signal_type: signalType,
          payload: payload
        })
      })
      
      if (!response.ok) {
        throw new Error(`Failed to send P2P signal: ${response.status}`)
      }
      
      return await response.json()
    } catch (error) {
      console.error('[Realtime] Error sending P2P signal:', error)
      throw error
    }
  }

  // Poll for pending P2P signals (fallback if realtime is slow)
  async function pollP2PSignals() {
    try {
      const response = await fetch('/api/v1/p2p/signals', {
        headers: {
          'Authorization': `Bearer ${(await supabase.auth.getSession()).data.session?.access_token}`
        }
      })
      
      if (!response.ok) return []
      
      const data = await response.json()
      const signals = data.signals || []
      
      // Process each signal (dedup avec le canal Realtime)
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

  return {
    // State
    isConnected,
    lastEventId,
    presenceState,
    
    // Actions
    connect,
    disconnect,
    onEvent,
    isUserOnline,
    sendP2PSignal,
    pollP2PSignals
  }
})
