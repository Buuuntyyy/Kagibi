// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"time"

	"kagibi/backend/pkg/monitoring"

	"github.com/gin-gonic/gin"
)

func generateRequestID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// MetricsMiddleware est un middleware Gin qui enregistre automatiquement
// les métriques pour chaque requête HTTP et logue les erreurs 5xx
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := generateRequestID()
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		monitoring.IncrementActiveConnections()
		defer monitoring.DecrementActiveConnections()

		start := time.Now()
		c.Next()
		duration := time.Since(start)

		status := c.Writer.Status()
		method := c.Request.Method
		endpoint := c.FullPath()

		monitoring.RecordRequestMetrics(method, endpoint, status, duration)

		if status >= 500 {
			monitoring.InternalErrorsTotal.WithLabelValues(method, endpoint).Inc()

			userID := c.GetString("user_id")
			if userID == "" {
				userID = "unauthenticated"
			}

			log.Printf("[ERROR_500] request_id=%s method=%s endpoint=%s status=%d duration=%s user_id=%s ip=%s",
				requestID,
				method,
				endpoint,
				status,
				duration.Round(time.Millisecond),
				userID,
				c.ClientIP(),
			)
		}
	}
}
