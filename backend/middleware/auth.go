package middleware

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func AuthMiddleware(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		sessionID, err := c.Cookie("session_id")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Non authorisé"})
			return
		}
		userdID, err := redisClient.Get(context.Background(), sessionID).Result()
		if err == redis.Nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Session invalide"})
			return
		} else if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Erreur serveur"})
			return
		}

		// Log Redis latency if it's slow (> 100ms)
		elapsed := time.Since(start)
		if elapsed > 100*time.Millisecond {
			log.Printf("SLOW REDIS AUTH: %v", elapsed)
		}

		c.Set("user_id", userdID)
		c.Next()
	}
}
