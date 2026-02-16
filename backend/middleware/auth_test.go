package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("No Cookie", func(t *testing.T) {
		_, _ = redismock.NewClientMock()
		r := gin.New()
		r.Use(AuthMiddleware(nil, "test-secret", nil)) // Use a dummy secret for JWT and nil mock (will be ignored in test)
		r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Invalid Session", func(t *testing.T) {
		client, mock := redismock.NewClientMock()

		mock.ExpectGet("fake-session").SetErr(redis.Nil)

		r := gin.New()
		r.Use(AuthMiddleware(nil, "test-secret", client)) // Use a dummy secret for JWT and pass the mock
		r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "fake-session"})
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Valid Session", func(t *testing.T) {
		client, mock := redismock.NewClientMock()

		mock.ExpectGet("valid-session").SetVal("user-123")

		// Créer un token JWT valide pour le test
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": "user-123",
			"exp": time.Now().Add(time.Hour).Unix(),
		})
		tokenString, _ := token.SignedString([]byte("test-secret"))

		r := gin.New()
		r.Use(AuthMiddleware(nil, "test-secret", client)) // Passer le mock Redis
		r.GET("/", func(c *gin.Context) {
			userID, _ := c.Get("user_id")
			assert.Equal(t, "user-123", userID)
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
