package ws

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"kagibi/backend/pkg/authprovider"
)

const wsTokenPrefix = "ws_token:"

// WebSocketHandler upgrades an HTTP connection to WebSocket.
// Authentication is read from (in order of preference):
//  1. "Authorization: Bearer <JWT>" header (non-browser / server-side clients)
//  2. "?ws_token=<single-use-token>" query parameter — token issued by POST /auth/ws-token,
//     stored in Redis for 30 s, consumed on first use (preferred for browser clients)
//  3. "Sec-WebSocket-Protocol: token, <JWT>" header — legacy browser workaround,
//     kept for backwards compatibility
func WebSocketHandler(provider authprovider.AuthProvider, redisClient *redis.Client, allowedOrigins []string) gin.HandlerFunc {
	originSet := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		if trimmed := strings.TrimSpace(o); trimmed != "" {
			originSet[trimmed] = true
		}
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			if origin == "" {
				return true
			}
			return originSet[origin]
		},
	}

	return func(c *gin.Context) {
		var userID string
		var err error

		// Method 1: single-use ws_token query parameter (preferred for browsers)
		if wsToken := c.Query("ws_token"); wsToken != "" && redisClient != nil {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			userID, err = consumeWsToken(ctx, redisClient, wsToken)
			cancel()
			if err != nil {
				log.Printf("[WS] Invalid or expired ws_token: %v", err)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired ws_token"})
				return
			}
		} else {
			// Methods 2 & 3: JWT from Authorization header or Sec-WebSocket-Protocol trick
			tokenStr := tokenFromRequest(c.Request)
			if tokenStr == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
				return
			}
			userID, err = validateToken(provider, tokenStr)
			if err != nil {
				log.Printf("[WS] Token validation failed: %v", err)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
				return
			}
		}

		// When the client used the Sec-WebSocket-Protocol trick, echo "token" back
		// so the browser does not close the connection due to a protocol mismatch.
		var responseHeader http.Header
		if proto := c.Request.Header.Get("Sec-WebSocket-Protocol"); proto != "" {
			responseHeader = http.Header{"Sec-WebSocket-Protocol": []string{"token"}}
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, responseHeader)
		if err != nil {
			log.Printf("[WS] Upgrade failed for user=%s: %v", userID, err)
			return
		}

		client := &Client{
			userID: userID,
			hub:    GlobalHub,
			conn:   conn,
			send:   make(chan []byte, 256),
		}

		GlobalHub.Register(client)

		go client.writePump()
		client.readPump() // blocks until connection closes
	}
}

// consumeWsToken validates and atomically deletes a single-use WebSocket token from Redis.
func consumeWsToken(ctx context.Context, redisClient *redis.Client, token string) (string, error) {
	key := wsTokenPrefix + token
	return redisClient.GetDel(ctx, key).Result()
}

// tokenFromRequest extracts the bearer token from:
//  1. "Authorization: Bearer <token>" header (non-browser / server-side clients), or
//  2. "Sec-WebSocket-Protocol: token, <JWT>" header (browser WebSocket clients).
func tokenFromRequest(r *http.Request) string {
	if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	// Browser WebSocket API cannot send custom headers, so the client encodes the
	// token as the second subprotocol: new WebSocket(url, ['token', jwt]).
	// The browser sends "Sec-WebSocket-Protocol: token, <JWT>".
	if proto := r.Header.Get("Sec-WebSocket-Protocol"); proto != "" {
		parts := strings.SplitN(proto, ",", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[0]) == "token" {
			return strings.TrimSpace(parts[1])
		}
	}
	return ""
}

// validateToken parses and validates the JWT, returning the user ID (sub claim).
func validateToken(provider authprovider.AuthProvider, tokenStr string) (string, error) {
	secret := provider.GetJWTSecret()
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return secret, nil
	})
	if err != nil || !token.Valid {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", jwt.ErrTokenInvalidClaims
	}

	userIDClaim := provider.GetUserIDClaim()
	userID, ok := claims[userIDClaim].(string)
	if !ok || userID == "" {
		return "", jwt.ErrTokenInvalidClaims
	}

	return userID, nil
}
