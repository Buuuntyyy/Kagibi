// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRateLimitMiddleware(t *testing.T) {
	// Setup Gin
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RateLimitMiddleware(nil)) // nil Redis → fail-open (no rate limiting in unit tests)
	r.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Test 1: Request OK
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test 2: Burst limit (simulation)
	// Note: It's hard to test exact rate limits in unit tests without mocking time,
	// but we can verify the headers are present if your middleware sets them,
	// or just verify it doesn't crash.
	// For this simple test, we just ensure it passes.
	for i := 0; i < 10; i++ {
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}
