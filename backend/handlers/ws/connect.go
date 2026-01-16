package ws

import (
	"context"
	"encoding/json"
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
	// 1. Authentification via JWT Supabase
	// Le token est passé soit via le middleware (c.Get) si la route est protégée,
	// soit via Query Param `?token=...` car les websockets ne supportent pas les headers custom facilement lors du handshake initial.

	userID := ""

	// Essayer de récupérer depuis le contexte (cas où un middleware Auth serait passé avant)
	if val, exists := c.Get("user_id"); exists {
		userID = val.(string)
	} else {
		// Fallback: Récupérer le token depuis l'URL query string
		tokenString := c.Query("token")
		if tokenString == "" {
			log.Println("WS: No token provided")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		secret := os.Getenv("SUPABASE_JWT_SECRET")
		if secret == "" {
			log.Println("WS: Server misconfiguration (missing jwt secret)")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Si le token est signé avec HMAC (HS256), on utilise le secret
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); ok {
				return []byte(secret), nil
			}

			// Si le token est signé avec ECDSA (ES256) et qu'on a le JWKS, on utilise le JWKS
			if _, ok := token.Method.(*jwt.SigningMethodECDSA); ok && jwks != nil {
				return jwks.Keyfunc(token)
			}

			// Sinon, méthode Inconnue
			return nil, jwt.ErrTokenUnverifiable
		})

		if err != nil || !token.Valid {
			log.Printf("WS: Invalid token: %v", err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			userID = claims["sub"].(string)
		} else {
			log.Println("WS: Invalid claims")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}

	uidStr := userID

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
