package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/uptrace/bun"
)

func LoginHandler(c *gin.Context, db *bun.DB, redisClient *redis.Client) {
	// Deprecated: Login is now handled by Supabase Auth on the client side.
	c.JSON(http.StatusGone, gin.H{"error": "Login endpoint is deprecated. Use Supabase Auth."})
}
