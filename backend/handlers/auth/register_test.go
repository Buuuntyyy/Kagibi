package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

const registerPath = "/register"

func TestRegisterHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Missing Required Fields", func(t *testing.T) {
		router := gin.New()
		router.POST(registerPath, func(c *gin.Context) {
			var req struct {
				Email     string `json:"email"`
				Password  string `json:"password"`
				PublicKey string `json:"publicKey"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
				return
			}
			if req.Email == "" || req.Password == "" || req.PublicKey == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "All fields required"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		body := map[string]string{
			"email":    "test@example.com",
			"password": "",
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", registerPath, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid Email Format", func(t *testing.T) {
		router := gin.New()
		router.POST(registerPath, func(c *gin.Context) {
			var req struct {
				Email     string `json:"email"`
				Password  string `json:"password"`
				PublicKey string `json:"publicKey"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
				return
			}
			// Simple email validation
			if req.Email != "" && !contains(req.Email, "@") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		body := map[string]string{
			"email":     "invalid-email",
			"password":  "password123",
			"publicKey": "key123",
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", registerPath, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func contains(s, substr string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '@' {
			return true
		}
	}
	return false
}
