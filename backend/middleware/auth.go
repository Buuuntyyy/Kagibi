// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package middleware

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"kagibi/backend/pkg/authprovider"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
)

// checkTokenRevocation verifies the token has not been revoked via Redis.
// Returns false (not revoked) if redisClient is nil.
// Returns an error only on a hard Redis failure (not on redis.Nil / key absent).
func checkTokenRevocation(redisClient *redis.Client, userID string, claims jwt.MapClaims) (revoked bool, redisErr error) {
	if redisClient == nil {
		return false, nil
	}
	iatFloat, ok := claims["iat"].(float64)
	if !ok {
		return false, nil
	}
	revokeKey := "token_revoke:" + userID
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	revokeStr, rErr := redisClient.Get(ctx, revokeKey).Result()
	cancel()
	if rErr == redis.Nil {
		return false, nil
	}
	if rErr != nil {
		return false, rErr
	}
	revokeTs, parseErr := strconv.ParseInt(revokeStr, 10, 64)
	return parseErr == nil && int64(iatFloat) < revokeTs, nil
}

// extractUserID parses the user ID from the token claims using the provider-specific claim key.
func extractUserID(provider authprovider.AuthProvider, claims jwt.MapClaims) (string, bool) {
	userIDClaim := provider.GetUserIDClaim()
	userIDRaw, exists := claims[userIDClaim]
	if !exists {
		return "", false
	}
	userID, ok := userIDRaw.(string)
	return userID, ok && userID != ""
}

// trackActiveSession records the user as active in Redis with a 15-minute TTL (best-effort).
func trackActiveSession(redisClient *redis.Client, userID string) {
	if redisClient == nil {
		return
	}
	go func(uid string) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		redisClient.Set(ctx, "active_user:"+uid, "1", 15*time.Minute)
	}(userID)
}

// AuthMiddleware validates JWT tokens using the configured auth provider (Supabase or PocketBase).
// It supports HS256 tokens and extracts the user ID from the provider-specific claim.
func AuthMiddleware(provider authprovider.AuthProvider, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(401, gin.H{"error": "Token manquant"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenUnverifiable
			}
			return provider.GetJWTSecret(), nil
		})

		if err != nil || !token.Valid {
			log.Printf("[Auth/%s] Token validation error: %v", provider.Name(), err)
			c.AbortWithStatusJSON(401, gin.H{"error": "Token invalide"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(401, gin.H{"error": "Claims invalides"})
			return
		}

		userID, valid := extractUserID(provider, claims)
		if !valid {
			log.Printf("[Auth/%s] Missing or invalid user ID claim in token", provider.Name())
			c.AbortWithStatusJSON(401, gin.H{"error": "Claims invalides"})
			return
		}

		// Token revocation check — rejects tokens issued before a password change or MFA disable.
		revoked, redisErr := checkTokenRevocation(redisClient, userID, claims)
		if redisErr != nil {
			// Redis error (timeout, connection failure) — fail-closed.
			log.Printf("[Auth] Redis revocation check failed for user %s: %v", userID, redisErr)
			c.AbortWithStatusJSON(503, gin.H{"error": "Service temporarily unavailable"})
			return
		}
		if revoked {
			c.AbortWithStatusJSON(401, gin.H{"error": "Token révoqué"})
			return
		}

		c.Set("user_id", userID)

		// Propagate AAL claim so MFA-guarded handlers can verify assurance level
		aal := "aal1"
		if aalClaim, ok := claims["aal"].(string); ok && aalClaim != "" {
			aal = aalClaim
		}
		c.Set("aal", aal)
		c.Set("is_guest", aal == "guest")

		trackActiveSession(redisClient, userID)

		c.Next()
	}
}

// BlockGuest rejects requests that carry a guest JWT (aal=guest).
// Apply this after AuthMiddleware on any route group that must be inaccessible
// to ephemeral P2P guest sessions (file storage, account management, etc.).
func BlockGuest() gin.HandlerFunc {
	return func(c *gin.Context) {
		if isGuest, _ := c.Get("is_guest"); isGuest == true {
			c.AbortWithStatusJSON(403, gin.H{"error": "Guest sessions cannot access this resource"})
			return
		}
		c.Next()
	}
}
