package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"safercloud/backend/pkg"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/uptrace/bun"
	"golang.org/x/crypto/bcrypt"
)

func LoginHandler(c *gin.Context, db *bun.DB, redisClient *redis.Client) {
	var credentials struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := pkg.FindUserByEmail(db, credentials.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Identifiants invalides"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(credentials.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Identifiants invalides"})
		return
	}

	sessionIDBytes := make([]byte, 32)
	if _, err := rand.Read(sessionIDBytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de créer la session"})
		return
	}
	sessionID := hex.EncodeToString(sessionIDBytes)

	err = redisClient.Set(context.Background(), sessionID, user.ID, time.Hour*24).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de créer la session"})
		return
	}

	domain := os.Getenv("COOKIE_DOMAIN")
	if domain == "" {
		domain = "localhost"
	}

	c.SetCookie("session_id", sessionID, 3600*24, "/", domain, true, true)

	c.JSON(http.StatusOK, gin.H{
		"message": "Connexion réussie",
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
		"salt":                 user.Salt,
		"encrypted_master_key": user.EncryptedMasterKey,
	})
}
