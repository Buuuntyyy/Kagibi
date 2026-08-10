// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package auth

import (
	"net/http"
	"time"

	"kagibi/backend/pkg"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type recoveryVerifyBackupRequest struct {
	RecoveryHash string `json:"recovery_hash" binding:"required"`
}

// VerifyRecoveryBackupHandler laisse l'utilisateur connecté prouver qu'il possède
// bien son code de récupération (kit sauvegardé). En cas de succès,
// recovery_verified_at est horodaté — utilisé par l'UI pour cesser les rappels.
func VerifyRecoveryBackupHandler(c *gin.Context, db *bun.DB) {
	userID := c.GetString("user_id")

	var req recoveryVerifyBackupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	user, err := pkg.FindUserByID(db, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	verified, needsUpgrade := verifyRecoveryHash(user.RecoveryHash, req.RecoveryHash)
	if !verified {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid recovery code"})
		return
	}

	now := time.Now().UTC()
	user.RecoveryVerifiedAt = &now
	columns := []string{"recovery_verified_at"}
	if needsUpgrade {
		user.RecoveryHash = hashRecoveryVerifier(req.RecoveryHash)
		columns = append(columns, "recovery_hash")
	}
	if _, err := db.NewUpdate().Model(user).Column(columns...).
		Where("id = ?", userID).Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update verification status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "recovery_verified_at": now})
}

type recoveryRotateRequest struct {
	RecoveryHash               string `json:"recovery_hash" binding:"required"`
	RecoverySalt               string `json:"recovery_salt" binding:"required"`
	EncryptedMasterKeyRecovery string `json:"encrypted_master_key_recovery" binding:"required"`
}

// RotateRecoveryHandler remplace le matériel de récupération par celui d'un
// nouveau code généré côté client (master key re-wrappée avec la nouvelle KEK).
// L'ancien code devient immédiatement inutilisable. Le nouveau code n'est pas
// considéré vérifié tant que l'utilisateur n'a pas refait la preuve de possession.
func RotateRecoveryHandler(c *gin.Context, db *bun.DB) {
	userID := c.GetString("user_id")

	var req recoveryRotateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	user, err := pkg.FindUserByID(db, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.RecoveryHash = hashRecoveryVerifier(req.RecoveryHash)
	user.RecoverySalt = req.RecoverySalt
	user.EncryptedMasterKeyRecovery = req.EncryptedMasterKeyRecovery
	user.RecoveryVerifiedAt = nil

	if _, err := db.NewUpdate().Model((*pkg.User)(nil)).
		Set("recovery_hash = ?, recovery_salt = ?, encrypted_master_key_recovery = ?, recovery_verified_at = NULL",
			user.RecoveryHash, user.RecoverySalt, user.EncryptedMasterKeyRecovery).
		Where("id = ?", userID).Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to rotate recovery material"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
