// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"time"

	"kagibi/backend/pkg"
	"kagibi/backend/pkg/mailer"

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

// --- Rotation du code de récupération : durcie par un code de confirmation email ---
//
// La rotation ne peut pas être confirmée par un simple lien cliquable : le nouveau
// matériel cryptographique (master key re-enveloppée avec la nouvelle KEK) est généré
// et détenu UNIQUEMENT en mémoire côté client au moment de la demande (architecture
// zero-knowledge — le serveur ne voit jamais la master key en clair). Un lien ouvert
// depuis un autre appareil ou un autre onglet n'aurait donc pas ce matériel prêt à
// soumettre. Le code de confirmation est donc un secret court à ressaisir dans la
// MÊME session, exactement comme un challenge MFA — juste livré par email plutôt que
// par TOTP.
//
// Défense en profondeur : une session volée (token dérobé, XSS, etc.) qui parviendrait
// à appeler /auth/recovery/rotate sans jamais passer par le kit de récupération légitime
// pourrait sinon planter une porte dérobée durable (un nouveau code de récupération connu
// de l'attaquant, permettant de reprendre le compte des semaines plus tard même après un
// changement de mot de passe). Le code email garantit que l'attaquant a aussi accès à la
// boîte mail de la victime, en plus de la session et (si activé) du second facteur.

const (
	recoveryRotationCodeTTL      = 15 * time.Minute
	recoveryRotationCodeCooldown = 60 * time.Second
	recoveryRotationCodeDigits   = 6
)

// recoveryRotationChallenge est un code de confirmation à usage unique pour la
// rotation du code de récupération. Un seul actif à la fois par utilisateur.
type recoveryRotationChallenge struct {
	bun.BaseModel `bun:"table:recovery_rotation_challenges,alias:rrc"`

	ID        int64      `bun:"id,pk,autoincrement"`
	UserID    string     `bun:"user_id,notnull"`
	CodeHash  string     `bun:"code_hash,notnull"`
	ExpiresAt time.Time  `bun:"expires_at,notnull"`
	UsedAt    *time.Time `bun:"used_at"`
	CreatedAt time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp"`
}

func generateNumericCode(digits int) (string, error) {
	max := int64(1)
	for range digits {
		max *= 10
	}
	n, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%0*d", digits, n.Int64()), nil
}

func hashRecoveryRotationCode(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

type requestRecoveryRotationCodeRequest struct {
	Lang string `json:"lang"`
}

// RequestRecoveryRotationCodeHandler génère un code de confirmation à 6 chiffres,
// invalide tout challenge en attente pour l'utilisateur, et l'envoie par email.
// POST /auth/recovery/rotate/request-code
func RequestRecoveryRotationCodeHandler(c *gin.Context, db *bun.DB) {
	userID := c.GetString("user_id")
	ctx := c.Request.Context()

	var recentCount int
	_ = db.NewSelect().Model((*recoveryRotationChallenge)(nil)).
		ColumnExpr("COUNT(*)").
		Where("user_id = ? AND created_at > ?", userID, time.Now().Add(-recoveryRotationCodeCooldown)).
		Scan(ctx, &recentCount)
	if recentCount > 0 {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Veuillez patienter avant de redemander un code"})
		return
	}

	user, err := pkg.FindUserByID(db, userID)
	if err != nil || user.Email == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load user"})
		return
	}

	code, err := generateNumericCode(recoveryRotationCodeDigits)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate code"})
		return
	}

	// Un seul challenge actif à la fois : invalider tout précédent.
	_, _ = db.NewDelete().Model((*recoveryRotationChallenge)(nil)).
		Where("user_id = ? AND used_at IS NULL", userID).Exec(ctx)

	challenge := &recoveryRotationChallenge{
		UserID:    userID,
		CodeHash:  hashRecoveryRotationCode(code),
		ExpiresAt: time.Now().Add(recoveryRotationCodeTTL),
	}
	if _, err := db.NewInsert().Model(challenge).Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create challenge"})
		return
	}

	var req requestRecoveryRotationCodeRequest
	_ = c.ShouldBindJSON(&req)
	lang := req.Lang
	if lang != "en" && lang != "fr" {
		lang = "fr"
	}

	if err := mailer.SendRecoveryRotationCode(user.Email, user.Name, code, lang); err != nil {
		log.Printf("[recovery-rotate] failed to send confirmation email to user=%s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send confirmation email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "expires_in": int(recoveryRotationCodeTTL.Seconds())})
}

// verifyRecoveryRotationEmailCode consomme (usage unique) un challenge email valide et
// non expiré pour cet utilisateur. Retourne false si absent/expiré/déjà utilisé/incorrect.
func verifyRecoveryRotationEmailCode(ctx context.Context, db *bun.DB, userID, code string) bool {
	var challenge recoveryRotationChallenge
	err := db.NewSelect().Model(&challenge).
		Where("user_id = ? AND code_hash = ? AND used_at IS NULL AND expires_at > ?",
			userID, hashRecoveryRotationCode(code), time.Now()).
		Scan(ctx)
	if err != nil {
		return false
	}
	now := time.Now().UTC()
	_, _ = db.NewUpdate().Model((*recoveryRotationChallenge)(nil)).
		Set("used_at = ?", now).
		Where("id = ?", challenge.ID).
		Exec(ctx)
	return true
}

type recoveryRotateRequest struct {
	RecoveryHash               string `json:"recovery_hash" binding:"required"`
	RecoverySalt               string `json:"recovery_salt" binding:"required"`
	EncryptedMasterKeyRecovery string `json:"encrypted_master_key_recovery" binding:"required"`
	EmailCode                  string `json:"email_code" binding:"required"`
}

// RotateRecoveryHandler remplace le matériel de récupération par celui d'un
// nouveau code généré côté client (master key re-wrappée avec la nouvelle KEK).
// L'ancien code devient immédiatement inutilisable. Le nouveau code n'est pas
// considéré vérifié tant que l'utilisateur n'a pas refait la preuve de possession.
// Double gate : MFA (middleware.RequireMFAForAction, obligatoire si le compte a la
// MFA activée — cf. middleware/mfa.go) + code de confirmation email (ci-dessus).
func RotateRecoveryHandler(c *gin.Context, db *bun.DB) {
	userID := c.GetString("user_id")

	var req recoveryRotateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if !verifyRecoveryRotationEmailCode(c.Request.Context(), db, userID, req.EmailCode) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Code de confirmation invalide ou expiré"})
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
