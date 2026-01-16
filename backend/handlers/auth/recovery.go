package auth

import (
	"net/http"
	"safercloud/backend/pkg"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type RecoveryInitRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type RecoveryFinishRequest struct {
	Email                 string `json:"email" binding:"required,email"`
	RecoveryHash          string `json:"recovery_hash" binding:"required"`
	NewPassword           string `json:"new_password" binding:"required,min=8"`
	NewSalt               string `json:"new_salt" binding:"required"`
	NewEncryptedMasterKey string `json:"new_encrypted_master_key" binding:"required"`
	// Optionally rotate recovery key too, but for simplicity we keep it or require re-generation
}

func RecoveryInitHandler(c *gin.Context, db *bun.DB) {
	var req RecoveryInitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := pkg.FindUserByEmail(db, req.Email)
	if err != nil {
		// Don't reveal if user exists
		c.JSON(http.StatusOK, gin.H{"message": "If the account exists, recovery data has been sent."})
		return
	}

	// Return the encrypted recovery blob.
	// In a real app, we might want to verify something first, but here the security is the code itself.
	c.JSON(http.StatusOK, gin.H{
		"encrypted_master_key_recovery": user.EncryptedMasterKeyRecovery,
		"salt":                          user.RecoverySalt, // Use RecoverySalt here
	})
}

func RecoveryFinishHandler(c *gin.Context, db *bun.DB) {
	var req RecoveryFinishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := pkg.FindUserByEmail(db, req.Email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Verify Recovery Hash to prove they have the code
	// Note: In a real zero-knowledge system, the server shouldn't know the code.
	// Here we assume the client sends a HASH of the code (or derived key) that the server stored during register.
	// If the client sends the same hash, it proves they derived the same key.
	if user.RecoveryHash != req.RecoveryHash {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid recovery code"})
		return
	}

	// Update user
    // We only update the encryption keys. Password hash is no longer stored here.
	user.Salt = req.NewSalt
	user.EncryptedMasterKey = req.NewEncryptedMasterKey

	_, err = db.NewUpdate().Model(user).Column("salt", "encrypted_master_key").Where("id = ?", user.ID).Exec(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Keys reset successfully"})
}
