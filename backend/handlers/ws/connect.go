package ws

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"safercloud/backend/pkg"
	wsPkg "safercloud/backend/pkg/ws"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/uptrace/bun"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Autoriser toutes les origines pour le développement (à restreindre en prod)
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ConnectHandler gère l'upgrade HTTP -> WebSocket
func ConnectHandler(c *gin.Context, manager *wsPkg.Manager, redisClient *redis.Client, db *bun.DB) {
	// 1. Récupérer le cookie de session manuellement
	userID, exists := c.Get("userID")
	if !exists {
		sessionID, err := c.Cookie("session_id")
		if err != nil {
			log.Println("WS: No session cookie")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		id, err := redisClient.Get(context.Background(), sessionID).Result()
		if err != nil {
			log.Println("WS: Invalid session")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		userID = id
	}

	uidStr := userID.(string)

	// 2. Upgrade connection
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WS: Failed to upgrade: %v", err)
		return
	}

	// 3. Register user
	manager.Register(uidStr, conn)

	// --- PRESENCE: Notify friends that I am Online ---
	notifyPresence(uidStr, true, manager, db)

	// 4. Listen for messages
	go func() {
		defer func() {
			manager.Unregister(uidStr, conn)
			// --- PRESENCE: Notify friends that I am Offline ---
			notifyPresence(uidStr, false, manager, db)
			conn.Close()
		}()

		for {
			_, messageData, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WS: Error reading: %v", err)
				}
				break
			}

			// Handle Signaling
			var msg struct {
				Type    string      `json:"type"`
				Payload interface{} `json:"payload"`
			}
			if err := json.Unmarshal(messageData, &msg); err != nil {
				log.Printf("WS: Invalid JSON: %v", err)
				continue
			}

			if msg.Type == "p2p_signal" {
				payloadMap, ok := msg.Payload.(map[string]interface{})
				if !ok {
					continue
				}
				// Expecting: { target_id: "...", type: "offer/answer/ice", data: ... }
				targetID, _ := payloadMap["target_id"].(string)
				signalType, _ := payloadMap["type"].(string)
				data := payloadMap["data"]

				if targetID != "" {
					manager.SendSignal(uidStr, targetID, signalType, data)
				}
			}
		}
	}()
}

func notifyPresence(userID string, isOnline bool, manager *wsPkg.Manager, db *bun.DB) {
	// Fetch accepted friends
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var friendships []pkg.Friendship
	err := db.NewSelect().Model(&friendships).
		Where("status = ?", "accepted").
		WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("user_id_1 = ?", userID).WhereOr("user_id_2 = ?", userID)
		}).Scan(ctx)

	if err != nil {
		log.Printf("WS Error: Failed to fetch friends for presence: %v", err)
		return
	}

	for _, f := range friendships {
		friendID := f.UserID1
		if friendID == userID {
			friendID = f.UserID2
		}

		isFriendOnline := manager.IsUserOnline(friendID)

		if isFriendOnline {
			// 1. Inform neighbor that I am here (or leaving)
			manager.SendToUser(friendID, wsPkg.MsgPresenceUpdate, map[string]interface{}{
				"user_id":   userID,
				"online":    isOnline,
				"timestamp": time.Now(),
			})
		}

		// 2. If I'm arriving, I need to know if neighbor is here
		if isOnline && isFriendOnline {
			manager.SendToUser(userID, wsPkg.MsgPresenceUpdate, map[string]interface{}{
				"user_id":   friendID, // The friend is online
				"online":    true,
				"timestamp": time.Now(),
			})
		}
	}
}
