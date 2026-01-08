package auth

import (
	"log"
	"net/http"
	"safercloud/backend/pkg"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"golang.org/x/crypto/bcrypt"
)

var validate = validator.New()

type RegisterRequest struct {
	Name                       string `json:"name" validate:"required"`
	Email                      string `json:"email" validate:"required,email"`
	Password                   string `json:"password" validate:"required,min=8"`
	Salt                       string `json:"salt" validate:"required"`
	EncryptedMasterKey         string `json:"encrypted_master_key" validate:"required"`
	EncryptedMasterKeyRecovery string `json:"encrypted_master_key_recovery" validate:"required"`
	RecoveryHash               string `json:"recovery_hash" validate:"required"`
	RecoverySalt               string `json:"recovery_salt" validate:"required"`
	PublicKey                  string `json:"public_key"`
	EncryptedPrivateKey        string `json:"encrypted_private_key"`
}

func RegisterHandler(c *gin.Context, db *bun.DB) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides"})
		return
	}

	// Valide les données
	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation échouée"})
		return
	}

	// Hache le mot de passe
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur interne"})
		return
	}

	user := &pkg.User{
		ID:                         uuid.New().String(),
		Name:                       req.Name,
		Email:                      req.Email,
		PasswordHash:               string(hashedPassword),
		Salt:                       req.Salt,
		EncryptedMasterKey:         req.EncryptedMasterKey,
		EncryptedMasterKeyRecovery: req.EncryptedMasterKeyRecovery,
		RecoveryHash:               req.RecoveryHash,
		RecoverySalt:               req.RecoverySalt,
		PublicKey:                  req.PublicKey,
		EncryptedPrivateKey:        req.EncryptedPrivateKey,
	}

	// Crée l'utilisateur
	if err := pkg.CreateUser(db, user); err != nil {
		log.Printf("Error creating user: %v", err)
		if strings.Contains(err.Error(), "unique constraint") || strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusConflict, gin.H{"error": "Cet email est déjà utilisé"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la création du compte"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Utilisateur créé avec succès"})
}
