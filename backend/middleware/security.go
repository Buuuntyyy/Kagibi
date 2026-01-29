package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

// generateNonce generates a cryptographically secure nonce
func generateNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func SecureHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate nonce for each request
		nonce := generateNonce()
		c.Set("csp_nonce", nonce)

		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Strict CSP with nonce + Service Worker support
		csp := fmt.Sprintf(
			"default-src 'none'; "+
				"script-src 'self' 'nonce-%s'; "+
				"style-src 'self' 'nonce-%s'; "+
				"img-src 'self' data: blob:; "+
				"font-src 'self'; "+
				"connect-src 'self' ws: wss:; "+
				"media-src 'self' blob:; "+
				"worker-src 'self'; "+ // Service Worker allowed
				"object-src 'none'; "+
				"base-uri 'self'; "+
				"form-action 'self'; "+
				"frame-ancestors 'none'; "+
				"upgrade-insecure-requests;",
			nonce, nonce,
		)
		c.Header("Content-Security-Policy", csp)

		// HSTS
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

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
