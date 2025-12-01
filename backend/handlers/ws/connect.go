package ws

import (
	"context"
	"log"
	"net/http"
	wsPkg "safercloud/backend/pkg/ws"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
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
func ConnectHandler(c *gin.Context, manager *wsPkg.Manager, redisClient *redis.Client) {
	// 1. Récupérer le cookie de session manuellement car ce n'est pas une requête API standard
	// Note: Le middleware AuthMiddleware pourrait être utilisé, mais pour les WS c'est parfois plus simple de le faire ici
	// car l'upgrade doit se faire avant tout write.

	// Cependant, si on est derrière le middleware AuthMiddleware, c.Get("userID") devrait fonctionner
	// SI le client envoie les cookies lors de la connexion WS.

	userID, exists := c.Get("userID")
	if !exists {
		// Fallback: Check cookie manually if middleware didn't run or failed
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

	// 2. Upgrade connection
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WS: Failed to upgrade: %v", err)
		return
	}

	// 3. Register user
	uidStr := userID.(string)
	manager.Register(uidStr, conn)

	// 4. Listen for close (Keep-alive loop)
	// On ne lit pas vraiment de messages du client pour l'instant, mais on doit garder la boucle active
	// pour détecter la déconnexion.
	go func() {
		defer func() {
			manager.Unregister(uidStr, conn)
			conn.Close()
		}()

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WS: Error reading: %v", err)
				}
				break
			}
		}
	}()
}
