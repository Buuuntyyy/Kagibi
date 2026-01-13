package ws

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Manager gère les connexions WebSocket actives
type Manager struct {
	clients map[string][]*websocket.Conn // Map UserID -> Liste de connexions
	lock    sync.RWMutex
}

// NewManager crée une nouvelle instance du gestionnaire
func NewManager() *Manager {
	return &Manager{
		clients: make(map[string][]*websocket.Conn),
	}
}

// Register ajoute une nouvelle connexion pour un utilisateur
func (m *Manager) Register(userID string, conn *websocket.Conn) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.clients[userID]; !ok {
		m.clients[userID] = make([]*websocket.Conn, 0)
	}
	m.clients[userID] = append(m.clients[userID], conn)
	log.Printf("WS: User %s connected. Total connections: %d", userID, len(m.clients[userID]))
}

// Unregister supprime une connexion
func (m *Manager) Unregister(userID string, conn *websocket.Conn) {
	m.lock.Lock()
	defer m.lock.Unlock()

	conns, ok := m.clients[userID]
	if !ok {
		return
	}

	for i, c := range conns {
		if c == conn {
			// Remove element at index i
			m.clients[userID] = append(conns[:i], conns[i+1:]...)
			break
		}
	}

	if len(m.clients[userID]) == 0 {
		delete(m.clients, userID)
	}
	log.Printf("WS: User %s disconnected.", userID)
}

// MessageType définit le type de message envoyé
type MessageType string

const (
	MsgStorageUpdate MessageType = "storage_update"
	MsgFriendUpdate  MessageType = "friend_update"
)

// Message structure pour les données JSON
type Message struct {
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload"`
}

// IsUserOnline vérifie si un utilisateur est connecté
func (m *Manager) IsUserOnline(userID string) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	conns, ok := m.clients[userID]
	return ok && len(conns) > 0
}

// SendSignal route un message de signalisation P2P d'un utilisateur à un autre
func (m *Manager) SendSignal(senderID, targetID string, signalType string, payload interface{}) {
	m.lock.RLock()
	conns, ok := m.clients[targetID]
	m.lock.RUnlock()

	if !ok || len(conns) == 0 {
		return // Target offline
	}

	msg := Message{
		Type: "p2p_signal",
		Payload: map[string]interface{}{
			"sender_id": senderID,
			"type":      signalType,
			"data":      payload,
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("WS Error: Failed to marshal signal: %v", err)
		return
	}

	for _, conn := range conns {
		conn.WriteMessage(websocket.TextMessage, data)
	}
}

// SendToUser envoie un message à toutes les connexions actives d'un utilisateur
func (m *Manager) SendToUser(userID string, msgType MessageType, payload interface{}) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	conns, ok := m.clients[userID]
	if !ok {
		return
	}

	msg := Message{
		Type:    msgType,
		Payload: payload,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("WS Error: Failed to marshal message: %v", err)
		return
	}

	for _, conn := range conns {
		err := conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Printf("WS Error: Failed to write message to user %s: %v", userID, err)
			// Note: On pourrait fermer la connexion ici, mais on laisse le handler de lecture le faire
			conn.Close()
		}
	}
}
