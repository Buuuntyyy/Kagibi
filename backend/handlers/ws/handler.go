// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package ws

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"kagibi/backend/middleware"
	"kagibi/backend/pkg/authprovider"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/uptrace/bun"
)

const (
	wsTokenPrefix     = "ws_token:"
	secWebSocketProto = "Sec-WebSocket-Protocol"
)

// WebSocketHandler upgrades an HTTP connection to WebSocket.
// Authentication is read from (in order of preference):
//  1. "Authorization: Bearer <JWT>" header (non-browser / server-side clients)
//  2. "?ws_token=<single-use-token>" query parameter — token issued by POST /auth/ws-token,
//     stored in Redis for 30 s, consumed on first use (preferred for browser clients)
//  3. "Sec-WebSocket-Protocol: token, <JWT>" header — legacy browser workaround,
//     kept for backwards compatibility
func WebSocketHandler(provider authprovider.AuthProvider, redisClient *redis.Client, db *bun.DB, allowedOrigins []string) gin.HandlerFunc {
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
		var claims jwt.MapClaims
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
			userID, claims, err = validateToken(provider, tokenStr)
			if err != nil {
				log.Printf("[WS] Token validation failed: %v", err)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
				return
			}
		}

		// Reject JWTs issued before a revocation event (password change, MFA disable,
		// account recovery) — mirrors the HTTP AuthMiddleware so a revoked session
		// cannot keep a live realtime channel open until natural expiry.
		if wsTokenRevoked(redisClient, userID, claims) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token revoked"})
			return
		}

		// Reject an aal1 session that must still complete MFA step-up — mirrors
		// EnforceMFAOnLogin so a stolen pre-step-up JWT cannot open a realtime channel
		// and passively observe activity (friend/org/storage/presence events) as a way
		// to bypass the HTTP API's MFA gate. Only applies to methods 2 & 3 (claims != nil):
		// the ws_token path (method 1) already passed through EnforceMFAOnLogin when the
		// token was issued via the protected, MFA-gated POST /auth/ws-token route.
		if claims != nil {
			aal, _ := claims["aal"].(string)
			if aal == "" {
				aal = "aal1"
			}
			mfaClaim, _ := claims["mfa"].(string)
			if middleware.MFALoginStepUpRequired(c.Request.Context(), db, userID, aal, mfaClaim) {
				c.JSON(http.StatusForbidden, gin.H{"error": "mfa_required"})
				return
			}
		}

		// Reject once the user is already at the connection cap — checked before the
		// upgrade so an over-budget attempt doesn't waste a handshake.
		if !GlobalHub.CanAcceptConnection(userID) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many active connections"})
			return
		}

		// When the client used the Sec-WebSocket-Protocol trick, echo "token" back
		// so the browser does not close the connection due to a protocol mismatch.
		var responseHeader http.Header
		if proto := c.Request.Header.Get(secWebSocketProto); proto != "" {
			responseHeader = http.Header{secWebSocketProto: []string{"token"}}
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, responseHeader)
		if err != nil {
			log.Printf("[WS] Upgrade failed for user=%s: %v", userID, err)
			return
		}

		client := &Client{
			id:     newConnID(),
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

// newConnID generates a random per-connection identifier, used as the member key
// for this connection's slot in the Redis connection-cap sorted set.
func newConnID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand failure is effectively unrecoverable; fall back to a
		// timestamp-based ID so the connection cap degrades to "less precise"
		// rather than panicking the connection.
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return hex.EncodeToString(b)
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
	if proto := r.Header.Get(secWebSocketProto); proto != "" {
		parts := strings.SplitN(proto, ",", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[0]) == "token" {
			return strings.TrimSpace(parts[1])
		}
	}
	return ""
}

// validateToken parses and validates the JWT, returning the user ID (sub claim)
// and the parsed claims (for downstream checks such as token revocation).
func validateToken(provider authprovider.AuthProvider, tokenStr string) (string, jwt.MapClaims, error) {
	secret := provider.GetJWTSecret()
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return secret, nil
	})
	if err != nil || !token.Valid {
		return "", nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", nil, jwt.ErrTokenInvalidClaims
	}

	userIDClaim := provider.GetUserIDClaim()
	userID, ok := claims[userIDClaim].(string)
	if !ok || userID == "" {
		return "", nil, jwt.ErrTokenInvalidClaims
	}

	return userID, claims, nil
}

// wsTokenRevoked reports whether a JWT (identified by its iat claim) was issued
// before a revocation event recorded in Redis for the user. Returns false when
// Redis is unavailable, no claims are present (single-use ws_token path), or no
// revocation has been recorded — matching the semantics of the HTTP middleware.
func wsTokenRevoked(redisClient *redis.Client, userID string, claims jwt.MapClaims) bool {
	if redisClient == nil || claims == nil {
		return false
	}
	iatFloat, ok := claims["iat"].(float64)
	if !ok {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	revokeStr, err := redisClient.Get(ctx, "token_revoke:"+userID).Result()
	if err != nil {
		// redis.Nil (no key) or a transient error — do not block the connection.
		return false
	}
	revokeTs, perr := strconv.ParseInt(revokeStr, 10, 64)
	return perr == nil && int64(iatFloat) < revokeTs
}
