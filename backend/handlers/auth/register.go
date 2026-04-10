// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package auth

import (
	"crypto/rand"
	"kagibi/backend/pkg"
	"kagibi/backend/pkg/authprovider"
	"log"
	"math/big"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/uptrace/bun"
)

var validate = validator.New()

type RegisterRequest struct {
	Name      string `json:"name" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	AvatarURL string `json:"avatar_url"`
	// Password is handled by the auth provider (Supabase or PocketBase), not the backend
	Salt                       string `json:"salt" validate:"required"`
	EncryptedMasterKey         string `json:"encrypted_master_key" validate:"required"`
	EncryptedMasterKeyRecovery string `json:"encrypted_master_key_recovery" validate:"required"`
	RecoveryHash               string `json:"recovery_hash" validate:"required"`
	RecoverySalt               string `json:"recovery_salt" validate:"required"`
	PublicKey                  string `json:"public_key"`
	EncryptedPrivateKey        string `json:"encrypted_private_key"`
	EncryptFilenames           bool   `json:"encrypt_filenames"`
}

func generateFriendCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, 8)
	for i := range code {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		code[i] = charset[num.Int64()]
	}
	return "#" + string(code)
}

// normalizeAvatarURL ensures the avatar URL has a valid prefix.
// Empty URLs fall back to the default avatar. Bare filenames are prefixed with /avatars/.
func normalizeAvatarURL(raw string) string {
	if raw == "" {
		return "/avatars/default.png"
	}
	if !strings.HasPrefix(raw, "/avatars/") && !strings.HasPrefix(raw, "http") && !strings.Contains(raw, "/") {
		return "/avatars/" + raw
	}
	return raw
}

// cleanupAuthProviderUser removes the auth provider user after a profile creation failure.
func cleanupAuthProviderUser(provider authprovider.AuthProvider, userID string) {
	if err := provider.DeleteUser(userID); err != nil {
		log.Printf("[Register] Failed to cleanup auth provider user after profile creation failure: %v", err)
	} else {
		log.Printf("[Register] Cleaned up orphaned auth provider user %s", userID)
	}
}

// RegisterHandler creates the Kagibi profile for a user already authenticated with the auth provider.
// The user ID comes from the JWT (set by AuthMiddleware) and matches the ID in the auth provider.
func RegisterHandler(c *gin.Context, db *bun.DB, provider authprovider.AuthProvider) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides"})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: JWT required to create profile"})
		return
	}

	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation échouée"})
		return
	}

	user := &pkg.User{
		ID:                         userID,
		Name:                       req.Name,
		Email:                      req.Email,
		AvatarURL:                  normalizeAvatarURL(req.AvatarURL),
		Salt:                       req.Salt,
		EncryptedMasterKey:         req.EncryptedMasterKey,
		EncryptedMasterKeyRecovery: req.EncryptedMasterKeyRecovery,
		RecoveryHash:               req.RecoveryHash,
		RecoverySalt:               req.RecoverySalt,
		FriendCode:                 generateFriendCode(),
		PublicKey:                  req.PublicKey,
		EncryptedPrivateKey:        req.EncryptedPrivateKey,
		EncryptFilenames:           req.EncryptFilenames,
	}

	log.Printf("[Register/%s] Creating profile for email: %s", provider.Name(), req.Email)

	if err := pkg.CreateUser(db, user); err != nil {
		log.Printf("[Register] Error creating profile: %v", err)
		if strings.Contains(err.Error(), "unique constraint") || strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusConflict, gin.H{"error": "Profil déjà existant"})
			return
		}
		log.Printf("[Register] Profile creation failed, cleaning up auth provider user %s", userID)
		cleanupAuthProviderUser(provider, userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la création du profil. Veuillez réessayer."})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Profil créé avec succès"})
}
