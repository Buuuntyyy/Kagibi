// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"time"

	"kagibi/backend/pkg/logger"
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

		userID := c.GetString("user_id")
		if userID == "" {
			userID = "anon"
		}

		level := slog.LevelInfo
		if status >= 500 {
			level = slog.LevelError
			monitoring.InternalErrorsTotal.WithLabelValues(method, endpoint).Inc()
		} else if status >= 400 {
			level = slog.LevelWarn
		}

		slog.Log(c.Request.Context(), level, "http_request",
			"request_id", requestID,
			"method", method,
			"path", endpoint,
			"status", status,
			"duration_ms", duration.Milliseconds(),
			"user_id", userID,
			"ip_anon", logger.AnonymiseIP(c.ClientIP()),
		)
	}
}
