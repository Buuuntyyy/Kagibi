package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

const wsTokenTTL = 30 * time.Second
const wsTokenPrefix = "ws_token:"

// WsTokenHandler issues a short-lived (30 s), single-use token that the frontend
// can pass as a ?ws_token= query parameter when opening the WebSocket connection.
// This is safer than the Sec-WebSocket-Protocol header trick, which some CDNs and
// reverse proxies strip or mangle.
//
// The token is stored in Redis as ws_token:<token> = <userID> with a 30 s TTL.
// The WebSocket handler consumes and deletes the key on first use.
//
// Route: POST /api/v1/auth/ws-token  (requires valid JWT via AuthMiddleware)
func WsTokenHandler(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}

		rawToken := make([]byte, 24)
		if _, err := rand.Read(rawToken); err != nil {
			log.Printf("[WsToken] Failed to generate random token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
			return
		}
		token := hex.EncodeToString(rawToken)

		key := wsTokenPrefix + token
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		if err := redisClient.Set(ctx, key, userID.(string), wsTokenTTL).Err(); err != nil {
			log.Printf("[WsToken] Failed to store token in Redis: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": token})
	}
}

// ConsumeWsToken validates and atomically deletes a single-use WebSocket token from Redis.
// Returns the userID associated with the token, or an error if invalid/expired.
func ConsumeWsToken(ctx context.Context, redisClient *redis.Client, token string) (string, error) {
	key := wsTokenPrefix + token
	userID, err := redisClient.GetDel(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return userID, nil
}
