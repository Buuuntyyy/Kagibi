package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("No Cookie", func(t *testing.T) {
		db, _ := redismock.NewClientMock()
		r := gin.New()
		r.Use(AuthMiddleware(db))
		r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Invalid Session", func(t *testing.T) {
		db, mock := redismock.NewClientMock()

		mock.ExpectGet("fake-session").SetErr(redis.Nil)

		r := gin.New()
		r.Use(AuthMiddleware(db))
		r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "fake-session"})
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Valid Session", func(t *testing.T) {
		db, mock := redismock.NewClientMock()

		mock.ExpectGet("valid-session").SetVal("user-123")

		r := gin.New()
		r.Use(AuthMiddleware(db))
		r.GET("/", func(c *gin.Context) {
			userID, _ := c.Get("user_id")
			assert.Equal(t, "user-123", userID)
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid-session"})
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
