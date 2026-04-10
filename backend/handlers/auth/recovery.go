// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package auth

import (
	"context"
	"crypto/hmac"
	"kagibi/backend/pkg"
	"kagibi/backend/pkg/authprovider"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
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
}

func RecoveryInitHandler(c *gin.Context, db *bun.DB) {
	var req RecoveryInitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := pkg.FindUserByEmail(db, req.Email)
	if err != nil {
		// Don't reveal whether the user exists
		c.JSON(http.StatusOK, gin.H{"message": "If the account exists, recovery data has been sent."})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"encrypted_master_key_recovery": user.EncryptedMasterKeyRecovery,
		"salt":                          user.RecoverySalt,
	})
}

func RecoveryFinishHandler(c *gin.Context, db *bun.DB, provider authprovider.AuthProvider, redisClient *redis.Client) {
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

	// Constant-time comparison to prevent timing side-channel on recovery hash.
	if !hmac.Equal([]byte(user.RecoveryHash), []byte(req.RecoveryHash)) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid recovery code"})
		return
	}

	// Update the password in the auth provider (Supabase or PocketBase)
	if err := provider.UpdateUserPassword(user.ID, req.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update auth password"})
		return
	}

	// In local mode: wipe TOTP so a compromised-account attacker can't keep their 2FA factor
	// active after the victim recovers. Also syncs user_security_settings accordingly.
	if lp, ok := getLocalProvider(provider); ok {
		if err := lp.DisableTOTP(user.ID); err != nil {
			log.Printf("[recovery] failed to disable TOTP for user=%s: %v", user.ID, err)
		}
		if err := lp.SyncMFAStatus(user.ID, false); err != nil {
			log.Printf("[recovery] failed to sync MFA status for user=%s: %v", user.ID, err)
		}
	}

	// Revoke all pre-recovery sessions — the old password is no longer valid.
	if redisClient != nil {
		go func(userID string) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			redisClient.Set(ctx, "token_revoke:"+userID, strconv.FormatInt(time.Now().Unix(), 10), 7*24*time.Hour)
		}(user.ID)
	}

	// Update the encrypted crypto keys on the backend
	user.Salt = req.NewSalt
	user.EncryptedMasterKey = req.NewEncryptedMasterKey

	_, err = db.NewUpdate().Model(user).Column("salt", "encrypted_master_key").Where("id = ?", user.ID).Exec(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Keys reset successfully"})
}
