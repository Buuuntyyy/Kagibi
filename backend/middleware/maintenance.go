// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

const MaintenanceKey = "kagibi:maintenance"

// MaintenanceMiddleware bloque toutes les requêtes (sauf /healthz) quand la clé
// Redis kagibi:maintenance est présente. Retourne 503 avec Retry-After.
func MaintenanceMiddleware(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/healthz" {
			c.Next()
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		exists, err := rdb.Exists(ctx, MaintenanceKey).Result()
		if err == nil && exists > 0 {
			c.Header("Retry-After", "60")
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "Le service est temporairement indisponible pour une mise à jour. Veuillez réessayer dans quelques minutes.",
				"code":  "maintenance",
			})
			return
		}

		c.Next()
	}
}

// SetMaintenance active le mode maintenance dans Redis.
func SetMaintenance(rdb *redis.Client, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return rdb.Set(ctx, MaintenanceKey, "1", ttl).Err()
}

// ClearMaintenance désactive le mode maintenance.
func ClearMaintenance(rdb *redis.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return rdb.Del(ctx, MaintenanceKey).Err()
}
