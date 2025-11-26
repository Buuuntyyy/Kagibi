package auth

import (
	"net/http"
	"safercloud/backend/pkg"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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
}

func RegisterHandler(c *gin.Context, db *bun.DB) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Valide les données
	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hache le mot de passe
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de hasher le mot de passe"})
		return
	}

	user := &pkg.User{
		Name:                       req.Name,
		Email:                      req.Email,
		PasswordHash:               string(hashedPassword),
		Salt:                       req.Salt,
		EncryptedMasterKey:         req.EncryptedMasterKey,
		EncryptedMasterKeyRecovery: req.EncryptedMasterKeyRecovery,
		RecoveryHash:               req.RecoveryHash,
		RecoverySalt:               req.RecoverySalt,
	}

	// Crée l'utilisateur
	if err := pkg.CreateUser(db, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Utilisateur créé avec succès"})
}
