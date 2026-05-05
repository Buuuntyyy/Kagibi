// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package ws

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"

	"kagibi/backend/pkg/monitoring"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024 // 512 KB

	// presenceGracePeriod is how long we wait before broadcasting "offline" after a
	// user's last connection drops. Reconnects within this window are transparent.
	presenceGracePeriod = 8 * time.Second

	redisChanPrefix     = "ws:user:"
	redisPresencePrefix = "ws:presence:"
	redisBroadcastChan  = "ws:broadcast"
	// presenceTTL must outlive the grace period so a pod crash doesn't leave stale keys forever.
	// It must also be longer than pingPeriod (54 s) so the key is still alive when renewPresence
	// is called from the pong handler. 5 minutes gives plenty of headroom.
	presenceTTL = 5 * time.Minute
)

// Client represents a single WebSocket connection from an authenticated user.
type Client struct {
	userID string
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
}

// Hub maintains the set of active clients and routes messages to them.
//
// Single-pod mode (rdb == nil): messages are delivered in-process only.
// Multi-pod mode (InitRedis called): SendToUser publishes to Redis so every pod
// can deliver the message to its locally connected clients, making the hub
// horizontally scalable with no change to callers.
type Hub struct {
	mu      sync.RWMutex
	clients map[string][]*Client // userID → local connections

	// localPresence tracks users that are considered online in single-pod mode.
	// It is set on first connect and cleared only when the grace-period timer fires,
	// so IsConnected returns true during the reconnect grace window even though
	// h.clients is temporarily empty after Unregister.
	localPresence map[string]struct{}

	// Redis fields — nil when running without Redis (single-pod / dev mode).
	rdb        *redis.Client
	pubsub     *redis.PubSub
	subMu      sync.Mutex
	subscribed map[string]struct{} // userIDs this pod is currently subscribed to
}

// GlobalHub is the singleton hub used throughout the application.
var GlobalHub = &Hub{
	clients:       make(map[string][]*Client),
	localPresence: make(map[string]struct{}),
}

// pendingDisconnects holds timers that fire the DisconnectHook after the grace period.
// If the user reconnects before the timer fires, the timer is cancelled.
var pendingDisconnects sync.Map // map[userID string]*time.Timer

// ConnectHook is called (in a goroutine) when a user's FIRST WebSocket connection opens
// AND there is no pending disconnect timer (i.e. not a quick reconnect).
var ConnectHook func(userID string)

// DisconnectHook is called when a user's LAST connection has been gone for presenceGracePeriod.
var DisconnectHook func(userID string)

// InitRedis wires up Redis Pub/Sub for cross-pod message delivery and cross-pod presence.
// Must be called once at startup, before any clients connect.
// Without this call the hub works in single-pod mode (zero regression).
func (h *Hub) InitRedis(rdb *redis.Client) {
	h.rdb = rdb
	h.subscribed = make(map[string]struct{})
	// Start with no channel subscriptions; channels are added dynamically as users connect.
	h.pubsub = rdb.Subscribe(context.Background(), redisBroadcastChan)
	go h.redisListener()
	log.Printf("[WS] Redis Pub/Sub enabled — hub is horizontally scalable")
}

// redisListener runs in its own goroutine and delivers incoming Redis messages
// to the locally connected clients of the target user.
func (h *Hub) redisListener() {
	ch := h.pubsub.Channel()
	for msg := range ch {
		if msg.Channel == redisBroadcastChan {
			h.localBroadcast([]byte(msg.Payload))
			continue
		}
		userID := strings.TrimPrefix(msg.Channel, redisChanPrefix)
		if userID == msg.Channel {
			continue // unrecognised channel prefix, skip
		}
		h.localSend(userID, []byte(msg.Payload))
	}
	log.Printf("[WS] Redis listener stopped")
}

// localBroadcast delivers msg to every locally connected client.
func (h *Hub) localBroadcast(msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, clients := range h.clients {
		for _, c := range clients {
			safeSend(c.send, msg)
		}
	}
}

// BroadcastToAll sends a message to every connected client across all pods.
func (h *Hub) BroadcastToAll(eventType string, payload map[string]any) {
	msg, err := json.Marshal(map[string]any{
		"type":    eventType,
		"payload": payload,
	})
	if err != nil {
		log.Printf("[WS] BroadcastToAll marshal error: %v", err)
		return
	}

	if h.rdb != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := h.rdb.Publish(ctx, redisBroadcastChan, msg).Err(); err != nil {
			log.Printf("[WS] BroadcastToAll Redis publish failed: %v — falling back to local", err)
			h.localBroadcast(msg)
		}
		return
	}
	h.localBroadcast(msg)
}

// subscribeUser adds a Redis Pub/Sub subscription for userID on this pod.
func (h *Hub) subscribeUser(userID string) {
	h.subMu.Lock()
	defer h.subMu.Unlock()
	if _, ok := h.subscribed[userID]; ok {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := h.pubsub.Subscribe(ctx, redisChanPrefix+userID); err != nil {
		log.Printf("[WS] Redis subscribe failed for user=%s: %v", userID, err)
		return
	}
	h.subscribed[userID] = struct{}{}
}

// unsubscribeUser removes the Redis Pub/Sub subscription for userID on this pod.
func (h *Hub) unsubscribeUser(userID string) {
	h.subMu.Lock()
	defer h.subMu.Unlock()
	if _, ok := h.subscribed[userID]; !ok {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := h.pubsub.Unsubscribe(ctx, redisChanPrefix+userID); err != nil {
		log.Printf("[WS] Redis unsubscribe failed for user=%s: %v", userID, err)
	}
	delete(h.subscribed, userID)
}

// setPresence marks the user as online in Redis, visible to all pods.
func (h *Hub) setPresence(userID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := h.rdb.Set(ctx, redisPresencePrefix+userID, "1", presenceTTL).Err(); err != nil {
		log.Printf("[WS] Redis setPresence failed for user=%s: %v", userID, err)
	}
}

// clearPresence removes the user's online marker from Redis.
func (h *Hub) clearPresence(userID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := h.rdb.Del(ctx, redisPresencePrefix+userID).Err(); err != nil {
		log.Printf("[WS] Redis clearPresence failed for user=%s: %v", userID, err)
	}
}

// renewPresence resets the TTL of the user's presence key so it doesn't expire
// while the WebSocket connection is still alive. Called on every pong received.
func (h *Hub) renewPresence(userID string) {
	if h.rdb == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := h.rdb.Expire(ctx, redisPresencePrefix+userID, presenceTTL).Err(); err != nil {
		log.Printf("[WS] Redis renewPresence failed for user=%s: %v", userID, err)
	}
}

// Register adds a client to the hub.
func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	firstConn := len(h.clients[c.userID]) == 0
	h.clients[c.userID] = append(h.clients[c.userID], c)
	log.Printf("[WS] Client registered: user=%s (total: %d)", c.userID, len(h.clients[c.userID]))
	h.mu.Unlock()
	monitoring.IncrementWSConnections()

	// If there was a pending offline timer, cancel it — the user reconnected in time.
	// The Redis subscription and presence key are still active, nothing else to do.
	if pendingTimer, wasPending := pendingDisconnects.LoadAndDelete(c.userID); wasPending {
		pendingTimer.(*time.Timer).Stop()
		log.Printf("[Presence] Reconnect within grace period for user=%s — no presence change", c.userID)
		return
	}

	// Truly first connection: set up presence, then fire hook.
	if firstConn {
		if h.rdb != nil {
			h.subscribeUser(c.userID)
			h.setPresence(c.userID)
		} else {
			h.mu.Lock()
			h.localPresence[c.userID] = struct{}{}
			h.mu.Unlock()
		}
		if ConnectHook != nil {
			go ConnectHook(c.userID)
		}
	}
}

// Unregister removes a client from the hub and schedules an offline broadcast
// after presenceGracePeriod (cancelled if the user reconnects in time).
func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	list := h.clients[c.userID]
	for i, client := range list {
		if client == c {
			h.clients[c.userID] = append(list[:i], list[i+1:]...)
			close(c.send)
			break
		}
	}
	lastConn := len(h.clients[c.userID]) == 0
	if lastConn {
		delete(h.clients, c.userID)
	}
	log.Printf("[WS] Client unregistered: user=%s (remaining: %d)", c.userID, len(h.clients[c.userID]))
	h.mu.Unlock()

	monitoring.DecrementWSConnections()
	if !lastConn {
		return
	}

	// Schedule the offline cleanup after the grace period.
	userID := c.userID
	useRedis := h.rdb != nil
	timer := time.AfterFunc(presenceGracePeriod, func() {
		pendingDisconnects.Delete(userID)
		log.Printf("[Presence] Grace period elapsed, broadcasting offline for user=%s", userID)
		if useRedis {
			h.unsubscribeUser(userID)
			h.clearPresence(userID)
		} else {
			h.mu.Lock()
			delete(h.localPresence, userID)
			h.mu.Unlock()
		}
		if DisconnectHook != nil {
			DisconnectHook(userID)
		}
	})
	pendingDisconnects.Store(userID, timer)
}

// localSend delivers msg directly to all local connections of userID.
func (h *Hub) localSend(userID string, msg []byte) {
	h.mu.RLock()
	clients := make([]*Client, len(h.clients[userID]))
	copy(clients, h.clients[userID])
	h.mu.RUnlock()

	for _, c := range clients {
		if !safeSend(c.send, msg) {
			log.Printf("[WS] Send buffer full or channel closed for user=%s, dropping message", userID)
		}
	}
}

// SendToUser sends a raw JSON message to all connections belonging to userID.
// In multi-pod mode the message is published to Redis so every pod delivers it locally.
// Falls back to local delivery if Redis publish fails.
func (h *Hub) SendToUser(userID string, msg []byte) {
	if h.rdb != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := h.rdb.Publish(ctx, redisChanPrefix+userID, msg).Err(); err != nil {
			log.Printf("[WS] Redis publish failed for user=%s: %v — falling back to local", userID, err)
			h.localSend(userID, msg)
		}
		return
	}
	h.localSend(userID, msg)
}

// safeSend sends msg to ch without blocking, recovering from a panic if ch was
// already closed by Unregister racing with SendToUser.
func safeSend(ch chan []byte, msg []byte) (sent bool) {
	defer func() {
		if recover() != nil {
			sent = false
		}
	}()
	select {
	case ch <- msg:
		return true
	default:
		return false
	}
}

// SendEventToUser marshals and delivers a structured event message.
func (h *Hub) SendEventToUser(userID, eventType string, id int64, payload map[string]any) {
	msg, err := json.Marshal(map[string]any{
		"type":       "event",
		"event_type": eventType,
		"id":         id,
		"payload":    payload,
	})
	if err != nil {
		log.Printf("[WS] Failed to marshal event: %v", err)
		return
	}
	h.SendToUser(userID, msg)
}

// SendP2PSignalToUser delivers a P2P signal over WebSocket.
// signalID is the database ID of the signal — the frontend uses it to deduplicate
// against the polling fallback so that a WS-delivered signal is not processed again
// 2.5 s later when the polling loop also picks it up.
func (h *Hub) SendP2PSignalToUser(targetUserID, senderID, signalType string, signalID int64, payload map[string]any) {
	msg, err := json.Marshal(map[string]any{
		"type":        "p2p_signal",
		"id":          signalID,
		"from":        senderID,
		"signal_type": signalType,
		"payload":     payload,
	})
	if err != nil {
		log.Printf("[WS] Failed to marshal p2p signal: %v", err)
		return
	}
	h.SendToUser(targetUserID, msg)
}

// IsConnected returns true if the user has at least one active WebSocket connection
// on any pod (Redis mode) or locally (single-pod mode).
// In single-pod mode, returns true during the reconnect grace period even if
// h.clients is temporarily empty after Unregister.
func (h *Hub) IsConnected(userID string) bool {
	if h.rdb != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		exists, err := h.rdb.Exists(ctx, redisPresencePrefix+userID).Result()
		if err == nil {
			return exists > 0
		}
		log.Printf("[WS] Redis IsConnected check failed for user=%s: %v — falling back to local", userID, err)
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, present := h.localPresence[userID]
	return present
}

// writePump pumps messages from the send channel to the WebSocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump handles incoming messages (pong frames keep the connection alive).
func (c *Client) readPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		c.hub.renewPresence(c.userID)
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("[WS] Unexpected close for user=%s: %v", c.userID, err)
			}
			break
		}
	}
}
