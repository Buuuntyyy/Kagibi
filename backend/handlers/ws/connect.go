package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"safercloud/backend/pkg"
	wsPkg "safercloud/backend/pkg/ws"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
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
func ConnectHandler(c *gin.Context, manager *wsPkg.Manager, redisClient *redis.Client, db *bun.DB, jwks keyfunc.Keyfunc) {
	userID, err := authenticateWebSocketUser(c, jwks)
	if err != nil {
		// Error logging handled in authenticate or implicit by return status
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WS: Failed to upgrade: %v", err)
		return
	}

	manager.Register(userID, conn)
	notifyPresence(userID, true, manager, db)

	go handleWebSocketConnection(conn, manager, db, userID)
}

func authenticateWebSocketUser(c *gin.Context, jwks keyfunc.Keyfunc) (string, error) {
	// Try getting from Context (Middleware)
	if val, exists := c.Get("user_id"); exists {
		return val.(string), nil
	}

	// Fallback: Query Param
	tokenString := c.Query("token")
	if tokenString == "" {
		log.Println("WS: No token provided")
		c.AbortWithStatus(http.StatusUnauthorized)
		return "", fmt.Errorf("No token")
	}

	return validateToken(tokenString, jwks)
}

func validateToken(tokenString string, jwks keyfunc.Keyfunc) (string, error) {
	secret := os.Getenv("SUPABASE_JWT_SECRET")
	if secret == "" {
		log.Println("WS: Server misconfiguration (missing jwt secret)")
		return "", fmt.Errorf("Missing secret")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); ok {
			return []byte(secret), nil
		}
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); ok && jwks != nil {
			return jwks.Keyfunc(token)
		}
		return nil, jwt.ErrTokenUnverifiable
	})

	if err != nil || !token.Valid {
		log.Printf("WS: Invalid token: %v", err)
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims["sub"].(string), nil
	}
	return "", fmt.Errorf("Invalid claims")
}

func handleWebSocketConnection(conn *websocket.Conn, manager *wsPkg.Manager, db *bun.DB, userID string) {
	defer func() {
		manager.Unregister(userID, conn)
		notifyPresence(userID, false, manager, db)
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
		handleMessage(messageData, manager, userID)
	}
}

func handleMessage(messageData []byte, manager *wsPkg.Manager, userID string) {
	var msg struct {
		Type    string      `json:"type"`
		Payload interface{} `json:"payload"`
	}
	if err := json.Unmarshal(messageData, &msg); err != nil {
		log.Printf("WS: Invalid JSON: %v", err)
		return
	}

	if msg.Type == "p2p_signal" {
		processP2PSignal(manager, userID, msg.Payload)
	}
}

func processP2PSignal(manager *wsPkg.Manager, userID string, payload interface{}) {
	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		return
	}
	// Expecting: { target_id: "...", type: "offer/answer/ice", data: ... }
	targetID, _ := payloadMap["target_id"].(string)
	signalType, _ := payloadMap["type"].(string)
	data := payloadMap["data"]

	if targetID != "" {
		manager.SendSignal(userID, targetID, signalType, data)
	}
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
