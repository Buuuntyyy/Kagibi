package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"net/http"
	"time"
)

func SecureHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	}
}

func RateLimiter() gin.HandlerFunc {
	rate := limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  100, // 100 requêtes par minute par IP
	}
	store := memory.NewStore()
	limiterMiddleware := limiter.New(store, rate)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		_, err := limiterMiddleware.Get(c.Request.Context(), ip)
		if err != nil {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Trop de requêtes"})
			c.Abort()
			return
		}
		c.Next()
	}
}
