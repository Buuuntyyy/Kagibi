package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"kagibi/backend/pkg/authprovider"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

// testProvider is a minimal AuthProvider for testing purposes.
type testProvider struct {
	secret       []byte
	userIDClaim  string
}

func (p *testProvider) Name() string                                     { return "test" }
func (p *testProvider) GetUserIDClaim() string                           { return p.userIDClaim }
func (p *testProvider) GetJWTSecret() []byte                             { return p.secret }
func (p *testProvider) DeleteUser(_ string) error                        { return nil }
func (p *testProvider) UpdateUserPassword(_, _ string) error             { return nil }

var _ authprovider.AuthProvider = (*testProvider)(nil)

func newTestProvider(secret string) *testProvider {
	return &testProvider{secret: []byte(secret), userIDClaim: "sub"}
}

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("No Cookie", func(t *testing.T) {
		_, _ = redismock.NewClientMock()
		r := gin.New()
		r.Use(AuthMiddleware(newTestProvider("test-secret"), nil))
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
		r.Use(AuthMiddleware(newTestProvider("test-secret"), client))
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
		// No active revocation for this user
		mock.ExpectGet("token_revoke:user-123").SetErr(redis.Nil)

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": "user-123",
			"exp": time.Now().Add(time.Hour).Unix(),
			"iat": time.Now().Unix(),
		})
		tokenString, _ := token.SignedString([]byte("test-secret"))

		r := gin.New()
		r.Use(AuthMiddleware(newTestProvider("test-secret"), client))
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
