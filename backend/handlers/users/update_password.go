// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package users

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"kagibi/backend/pkg"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type UpdatePasswordRequest struct {
	NewSalt               string `json:"new_salt" binding:"required"`
	NewEncryptedMasterKey string `json:"new_encrypted_master_key" binding:"required"`
}

// validateCryptoKeys validates the format and length of cryptographic keys
func validateCryptoKeys(salt, encryptedKey string) error {
	// 1. Validate hex format of salt
	saltBytes, err := hex.DecodeString(salt)
	if err != nil {
		return fmt.Errorf("invalid salt format: must be hex-encoded")
	}

	// 2. Check salt length (16 bytes minimum)
	if len(saltBytes) < 16 {
		return fmt.Errorf("salt too short: minimum 16 bytes required")
	}

	if len(saltBytes) > 64 {
		return fmt.Errorf("salt too long: maximum 64 bytes")
	}

	// 3. Validate base64 format of encrypted key
	// Accepter les 4 variantes : le web produit RawURLEncoding (sodium URLSAFE_NO_PADDING),
	// d'autres clients peuvent produire StdEncoding ou les variantes sans padding.
	var keyBytes []byte
	var b64Err error
	for _, enc := range []interface{ DecodeString(string) ([]byte, error) }{
		base64.RawURLEncoding,
		base64.URLEncoding,
		base64.StdEncoding,
		base64.RawStdEncoding,
	} {
		keyBytes, b64Err = enc.DecodeString(encryptedKey)
		if b64Err == nil {
			break
		}
	}
	if b64Err != nil {
		return fmt.Errorf("invalid encrypted key format: must be base64-encoded")
	}

	// 4. Check structure (IV + Encrypted Data + Auth Tag)
	// For AES-GCM: IV (12 bytes) + Data (variable) + Tag (16 bytes)
	if len(keyBytes) < 28 { // 12 + 0 + 16
		return fmt.Errorf("encrypted key too short")
	}

	// 5. Reasonable maximum limit
	if len(keyBytes) > 1024 {
		return fmt.Errorf("encrypted key too large")
	}

	return nil
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

	// CRITICAL: Validate cryptographic keys
	if err := validateCryptoKeys(req.NewSalt, req.NewEncryptedMasterKey); err != nil {
		log.Printf("SECURITY: Invalid crypto keys - UserID: %s, Error: %v", userID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cryptographic keys: " + err.Error()})
		return
	}

	// Fetch user
	user, err := pkg.FindUserByID(db, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
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

	log.Printf("INFO: Encryption keys updated - UserID: %s", userID)
	c.JSON(http.StatusOK, gin.H{"message": "Encryption keys updated successfully after password change"})
}
