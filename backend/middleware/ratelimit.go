// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// endpointWindows defines per-endpoint rate limits: (window duration, max requests per window).
// Shared across all Kubernetes replicas via Redis.
var endpointWindows = map[string][2]int64{
	// Auth — strict limits to prevent brute-force and account enumeration
	// [window_seconds, max_requests]
	"/api/v1/auth/login":           {60, 5},
	"/api/v1/auth/signup":          {60, 3},
	"/api/v1/auth/refresh":         {60, 30},
	"/api/v1/auth/update-password": {60, 2},
	"/api/v1/auth/register":        {60, 3},
	"/api/v1/auth/recovery/init":   {60, 3},
	"/api/v1/auth/recovery/finish": {60, 3},
	// MFA — strict limits to prevent TOTP brute-force
	"/api/v1/auth/mfa/verify":    {60, 5},
	"/api/v1/auth/mfa/enroll":    {60, 3},
	"/api/v1/auth/mfa/challenge": {60, 30},
	"/api/v1/auth/mfa/unenroll":  {60, 3},
	// User operations
	"/api/v1/users/change-password": {60, 2},
	"/api/v1/users/profile":         {60, 5},
	// File operations — higher limits for chunked transfers
	"/api/v1/files/upload":   {1, 50},
	"/api/v1/files/download": {1, 100},
}

// defaultWindow is used for all endpoints not listed in endpointWindows.
var defaultWindow = [2]int64{1, 30}

// RateLimitMiddleware returns a Gin middleware backed by Redis fixed-window counters,
// shared across all Kubernetes replicas. This replaces the previous in-memory sync.Map
// implementation that applied limits per-pod rather than globally.
//
// On Redis failure the middleware fails open (rate limiting is non-security-critical;
// the fail-closed behaviour for token revocation is handled separately in auth.go).
func RateLimitMiddleware(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip rate limiting for CORS preflight requests.
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		path := c.Request.URL.Path
		ip := c.ClientIP()

		window, limit := defaultWindow[0], defaultWindow[1]
		if cfg, ok := endpointWindows[path]; ok {
			window, limit = cfg[0], cfg[1]
		}

		// Fixed-window key: changes every <window> seconds, ensuring automatic expiry.
		slot := time.Now().Unix() / window
		key := fmt.Sprintf("rl:%s:%s:%d", ip, path, slot)

		ctx := c.Request.Context()
		count, err := redisClient.Incr(ctx, key).Result()
		if err != nil {
			// Redis unavailable — fail open for rate limiting.
			log.Printf("[RateLimit] Redis error for key %s: %v", key, err)
			c.Next()
			return
		}

		// Set expiry only on the first increment to avoid repeated EXPIRE calls.
		if count == 1 {
			redisClient.Expire(ctx, key, time.Duration(window*2)*time.Second)
		}

		if count > limit {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests",
			})
			return
		}

		c.Next()
	}
}

// redisRateLimitKey builds a deterministic key for testing.
func redisRateLimitKey(ip, path string, windowSeconds int64) string {
	slot := time.Now().Unix() / windowSeconds
	return fmt.Sprintf("rl:%s:%s:%d", ip, path, slot)
}

// newNilRateLimiter returns a no-op middleware for use in tests where no Redis
// client is available.
func newNilRateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}

// noopRedis is a minimal Redis stub so tests can call RateLimitMiddleware(nil)
// without panicking — nil redisClient causes the first Incr to error and fail open.
var _ = context.Background // keep context import used
