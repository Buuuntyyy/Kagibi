package auth

import (
	"crypto/rand"
	"log"
	"math/big"
	"net/http"
	"safercloud/backend/pkg"
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
	// Password removed (handled by Supabase)
	Salt                       string `json:"salt" validate:"required"`
	EncryptedMasterKey         string `json:"encrypted_master_key" validate:"required"`
	EncryptedMasterKeyRecovery string `json:"encrypted_master_key_recovery" validate:"required"`
	RecoveryHash               string `json:"recovery_hash" validate:"required"`
	RecoverySalt               string `json:"recovery_salt" validate:"required"`
	PublicKey                  string `json:"public_key"`
	EncryptedPrivateKey        string `json:"encrypted_private_key"`
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

// Rename this to CreateProfileHandler to reflect its new purpose
func RegisterHandler(c *gin.Context, db *bun.DB) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides"})
		return
	}

	// Get UserID from JWT (set by auth middleware)
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: JWT required to create profile"})
		return
	}

	// Valide les données
	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation échouée"})
		return
	}

	// Set default avatar if not provided
	avatarURL := req.AvatarURL
	if avatarURL == "" {
		avatarURL = "/avatars/default.png"
	}

	user := &pkg.User{
		ID:        userID, // Use Supabase ID
		Name:      req.Name,
		Email:     req.Email,
		AvatarURL: avatarURL,
		// PasswordHash: removed
		Salt:                       req.Salt,
		EncryptedMasterKey:         req.EncryptedMasterKey,
		EncryptedMasterKeyRecovery: req.EncryptedMasterKeyRecovery,
		RecoveryHash:               req.RecoveryHash,
		RecoverySalt:               req.RecoverySalt,
		FriendCode:                 generateFriendCode(),
		PublicKey:                  req.PublicKey,
		EncryptedPrivateKey:        req.EncryptedPrivateKey,
	}

	log.Printf("[Register] Creating user profile for email: %s", req.Email)

	// Crée l'utilisateur
	if err := pkg.CreateUser(db, user); err != nil {
		log.Printf("Error creating user profile: %v", err)
		if strings.Contains(err.Error(), "unique constraint") || strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusConflict, gin.H{"error": "Profil déjà existant"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la création du profil"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Profil créé avec succès"})
}
