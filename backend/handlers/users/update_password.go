package users

import (
	"net/http"
	"safercloud/backend/pkg"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"golang.org/x/crypto/bcrypt"
)

type UpdatePasswordRequest struct {
	CurrentPassword       string `json:"current_password" binding:"required"`
	NewPassword           string `json:"new_password" binding:"required,min=8"`
	NewSalt               string `json:"new_salt" binding:"required"`
	NewEncryptedMasterKey string `json:"new_encrypted_master_key" binding:"required"`
}

func UpdatePasswordHandler(c *gin.Context, db *bun.DB) {
	var req UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDInterface, _ := c.Get("user_id")
	userID, _ := userIDInterface.(string)

	// Fetch user
	user, err := pkg.FindUserByID(db, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Mot de passe actuel incorrect"})
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Update user
	user.PasswordHash = string(hashedPassword)
	user.Salt = req.NewSalt
	user.EncryptedMasterKey = req.NewEncryptedMasterKey

	_, err = db.NewUpdate().Model(user).Column("password_hash", "salt", "encrypted_master_key").Where("id = ?", userID).Exec(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Mot de passe mis à jour avec succès"})
}
