package auth

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"safercloud/backend/pkg"
	"strings"
	"time"

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
	} else {
		// Ensure avatar URL has the correct path prefix
		if !strings.HasPrefix(avatarURL, "/avatars/") && !strings.HasPrefix(avatarURL, "http") {
			// If it's just the filename, prepend /avatars/
			if !strings.Contains(avatarURL, "/") {
				avatarURL = "/avatars/" + avatarURL
			}
		}
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

		// Si c'est une erreur de contrainte unique, l'utilisateur existe déjà
		// Pas besoin de supprimer de Supabase
		if strings.Contains(err.Error(), "unique constraint") || strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusConflict, gin.H{"error": "Profil déjà existant"})
			return
		}

		// Pour toute autre erreur, supprimer l'utilisateur de Supabase
		// pour éviter les comptes orphelins (dans auth.users mais pas dans profiles)
		log.Printf("[Register] Profile creation failed, cleaning up Supabase auth.users entry for user %s", userID)
		if cleanupErr := deleteUserFromSupabaseAuth(userID); cleanupErr != nil {
			log.Printf("[Register] Failed to cleanup Supabase user after profile creation failure: %v", cleanupErr)
		} else {
			log.Printf("[Register] Cleaned up orphaned Supabase user %s", userID)
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la création du profil. Veuillez réessayer."})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Profil créé avec succès"})
}

// deleteUserFromSupabaseAuth supprime un utilisateur de Supabase auth.users
// Utilisé pour nettoyer les comptes orphelins si la création du profil échoue
func deleteUserFromSupabaseAuth(userID string) error {
	supabaseURL := os.Getenv("SUPABASE_URL")
	adminKey := os.Getenv("SUPABASE_ADMIN_KEY")
	if adminKey == "" {
		adminKey = os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	}

	if supabaseURL == "" || adminKey == "" {
		log.Printf("[Register] SUPABASE_URL or admin key not set, cannot cleanup orphaned user")
		return fmt.Errorf("supabase credentials not configured")
	}

	// Endpoint Admin API pour supprimer un utilisateur
	url := fmt.Sprintf("%s/auth/v1/admin/users/%s", supabaseURL, userID)

	// Créer la requête DELETE
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Headers nécessaires
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminKey))
	req.Header.Set("apikey", adminKey)

	// Exécuter la requête
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute DELETE request: %w", err)
	}
	defer resp.Body.Close()

	// Lire la réponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Vérifier le code de statut
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("supabase admin API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
