package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"safercloud/backend/pkg"
	"time"

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

	// Update Supabase password first (Admin API)
	if err := updateSupabasePassword(user.ID, req.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update auth password"})
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

func updateSupabasePassword(userID, newPassword string) error {
	supabaseURL := os.Getenv("SUPABASE_URL")
	adminKey := os.Getenv("SUPABASE_ADMIN_KEY")
	if adminKey == "" {
		adminKey = os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	}

	if supabaseURL == "" || adminKey == "" {
		return fmt.Errorf("supabase credentials not configured")
	}

	payload := map[string]string{"password": newPassword}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/auth/v1/admin/users/%s", supabaseURL, userID)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminKey))
	req.Header.Set("apikey", adminKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("supabase admin API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
