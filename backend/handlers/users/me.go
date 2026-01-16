package users

import (
	"net/http"
	"safercloud/backend/pkg"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// Structure pour la réponse JSON, filtrée pour la sécurité et la confidentialité
// Ne renvoie que les informations nécessaires pour :
// - Upload/Download (id, storage_used, storage_limit, plan)
// - Partage (public_key, encrypted_private_key, friend_code, id)
// - Affichage UI (name, email, created_at)
type UserResponse struct {
	ID                   string    `json:"id"`
	Name                 string    `json:"name"`
	Email                string    `json:"email"`
	StorageUsed          int64     `json:"storage_used"`
	StorageLimit         int64     `json:"storage_limit"`
	Plan                 string    `json:"plan"`
	FriendCode           string    `json:"friend_code"`
	PublicKey            string    `json:"public_key"`
	EncryptedPrivateKey  string    `json:"encrypted_private_key"`
	CreatedAt            time.Time `json:"created_at"`
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

	// 4. Construire une réponse filtrée avec UNIQUEMENT les champs nécessaires
	// EXCLUT : password_hash, salt, encrypted_master_key, encrypted_master_key_recovery, recovery_hash, recovery_salt
	response := UserResponse{
		ID:                  user.ID,
		Name:                user.Name,
		Email:               user.Email,
		StorageUsed:         user.StorageUsed,
		StorageLimit:        user.StorageLimit,
		Plan:                user.Plan,
		FriendCode:          user.FriendCode,
		PublicKey:           user.PublicKey,
		EncryptedPrivateKey: user.EncryptedPrivateKey,
		CreatedAt:           user.CreatedAt,
	}

	c.JSON(http.StatusOK, response)
}
