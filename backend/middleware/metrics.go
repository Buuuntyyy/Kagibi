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

// MetricsMiddleware enregistre les métriques Prometheus et logue chaque requête
// HTTP avec slog (JSON structuré). Toutes les requêtes sont loguées ; les 5xx
// sont loguées au niveau Error, les 4xx au niveau Warn, le reste en Info.
// L'adresse IP est anonymisée conformément à la délibération CNIL 2021-122.
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
		}

		userID := c.GetString("user_id")
		if userID == "" {
			userID = "unauthenticated"
		}

		level := slog.LevelInfo
		if status >= 500 {
			level = slog.LevelError
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
			"user_agent", c.Request.UserAgent(),
		)
	}
}
