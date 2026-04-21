// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package users

import (
	"kagibi/backend/pkg"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// Structure pour la réponse JSON, filtrée pour la sécurité et la confidentialité
// Ne renvoie que les informations nécessaires pour :
// - Upload/Download (id, storage_used, storage_limit, plan)
// - Partage (public_key, encrypted_private_key, friend_code, id)
// - Affichage UI (name, email, avatar_url, created_at)
type UserResponse struct {
	ID                  string    `json:"id"`
	Name                string    `json:"name"`
	Email               string    `json:"email"`
	AvatarURL           string    `json:"avatar_url"`
	StorageUsed         int64     `json:"storage_used"`
	StorageLimit        int64     `json:"storage_limit"`
	Plan                string    `json:"plan"`
	P2PMaxExchanges     int       `json:"p2p_max_exchanges"`
	P2PExchangesUsed    int       `json:"p2p_exchanges_used"`
	FriendCode          string    `json:"friend_code"`
	PublicKey           string    `json:"public_key"`
	EncryptedPrivateKey string    `json:"encrypted_private_key"`
	EncryptFilenames    bool      `json:"encrypt_filenames"`
	CreatedAt           time.Time `json:"created_at"`
}

func MeHandler(c *gin.Context, db *bun.DB) {
	// 1. Récupérer l'ID utilisateur placé dans le contexte par le middleware
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Contexte utilisateur non trouvé"})
		return
	}

	// 2. Convertir l'ID en string
	userID, ok := userIDInterface.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Format d'ID utilisateur invalide"})
		return
	}

	// 3. Récupérer les informations de l'utilisateur depuis la base de données
	user, err := pkg.FindUserByID(db, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Utilisateur non trouvé"})
		return
	}

	activeP2P, _ := pkg.CountUserActiveP2PExchanges(db, userID)

	planState, err := pkg.FindUserPlanByUserID(db, userID)
	if err != nil || planState == nil {
		planState = &pkg.UserPlan{
			UserID:           user.ID,
			Plan:             pkg.PlanFree,
			StorageLimit:     pkg.StorageFree,
			StorageUsed:      0,
			P2PMaxExchanges:  pkg.P2PLimitFree,
			P2PExchangesUsed: activeP2P,
		}
		_ = pkg.UpsertUserPlan(db, planState)
	} else {
		planState.P2PExchangesUsed = activeP2P
		_ = pkg.UpsertUserPlan(db, planState)
	}

	// 4. Construire une réponse filtrée avec UNIQUEMENT les champs nécessaires
	// EXCLUT : password_hash, salt, encrypted_master_key, encrypted_master_key_recovery, recovery_hash, recovery_salt
	response := UserResponse{
		ID:                  user.ID,
		Name:                user.Name,
		Email:               user.Email,
		AvatarURL:           user.AvatarURL,
		StorageUsed:         planState.StorageUsed,
		StorageLimit:        planState.StorageLimit,
		Plan:                planState.Plan,
		P2PMaxExchanges:     planState.P2PMaxExchanges,
		P2PExchangesUsed:    planState.P2PExchangesUsed,
		FriendCode:          user.FriendCode,
		PublicKey:           user.PublicKey,
		EncryptedPrivateKey: user.EncryptedPrivateKey,
		EncryptFilenames:    user.EncryptFilenames,
		CreatedAt:           user.CreatedAt,
	}

	c.JSON(http.StatusOK, response)
}
