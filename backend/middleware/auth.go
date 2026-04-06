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

		userIDClaim := provider.GetUserIDClaim()
		userIDRaw, exists := claims[userIDClaim]
		if !exists {
			log.Printf("[Auth/%s] Missing user ID claim '%s' in token", provider.Name(), userIDClaim)
			c.AbortWithStatusJSON(401, gin.H{"error": "Claims invalides"})
			return
		}

		userID, ok := userIDRaw.(string)
		if !ok || userID == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "User ID invalide"})
			return
		}

		// Token revocation check — rejects tokens issued before a password change or MFA disable.
		// On those events, "token_revoke:<userID>" is set to the Unix timestamp of the change.
		// Any token whose iat (issued-at) is older than that timestamp is rejected.
		if redisClient != nil {
			if iatFloat, ok := claims["iat"].(float64); ok {
				revokeKey := "token_revoke:" + userID
				ctx2, cancel2 := context.WithTimeout(context.Background(), 500*time.Millisecond)
				revokeStr, rErr := redisClient.Get(ctx2, revokeKey).Result()
				cancel2()
				if rErr == nil {
					if revokeTs, parseErr := strconv.ParseInt(revokeStr, 10, 64); parseErr == nil && int64(iatFloat) < revokeTs {
						c.AbortWithStatusJSON(401, gin.H{"error": "Token révoqué"})
						return
					}
				} else if rErr != redis.Nil {
					// Redis error (timeout, connection failure) — fail-closed.
					// Allowing revoked tokens through during Redis unavailability is a security regression.
					log.Printf("[Auth] Redis revocation check failed for user %s: %v", userID, rErr)
					c.AbortWithStatusJSON(503, gin.H{"error": "Service temporarily unavailable"})
					return
				}
			}
		}

		c.Set("user_id", userID)

		// Propagate AAL claim so MFA-guarded handlers can verify assurance level
		aal := "aal1"
		if aalClaim, ok := claims["aal"].(string); ok && aalClaim != "" {
			aal = aalClaim
		}
		c.Set("aal", aal)

		// Track active session in Redis (TTL 15 minutes) for monitoring
		if redisClient != nil {
			go func(uid string) {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				redisClient.Set(ctx, "active_user:"+uid, "1", 15*time.Minute)
			}(userID)
		}

		c.Next()
	}
}
