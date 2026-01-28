package users

import (
	"net/http"
	"safercloud/backend/pkg"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type UpdatePasswordRequest struct {
	NewSalt               string `json:"new_salt" binding:"required"`
	NewEncryptedMasterKey string `json:"new_encrypted_master_key" binding:"required"`
}

func UpdatePasswordHandler(c *gin.Context, db *bun.DB) {
	var req UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	userIDInterface, _ := c.Get("user_id")
	userID, _ := userIDInterface.(string)

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	// Fetch user
	user, err := pkg.FindUserByID(db, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Verify new salt and encrypted master key are not empty
	if req.NewSalt == "" || req.NewEncryptedMasterKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Salt and encrypted master key cannot be empty"})
		return
	}

	// Update user with new salt and encrypted master key
	// With Supabase, we trust the JWT token. The client handles the password verification/update with Supabase.
	// This endpoint only updates the rotation of the Master Key wrapping.
	user.Salt = req.NewSalt
	user.EncryptedMasterKey = req.NewEncryptedMasterKey

	_, err = db.NewUpdate().Model(user).Column("salt", "encrypted_master_key").Where("id = ?", userID).Exec(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update encryption keys: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Encryption keys updated successfully after password change"})
}
