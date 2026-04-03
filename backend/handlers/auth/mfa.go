package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"strconv"
	"time"

	"safercloud/backend/pkg/authprovider"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// MFAListFactorsHandler handles GET /api/v1/auth/mfa/factors (protected).
// Returns the user's TOTP factor(s), compatible with the Supabase MFA factors shape.
func MFAListFactorsHandler(provider authprovider.AuthProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		lp, ok := getLocalProvider(provider)
		if !ok {
			c.JSON(http.StatusNotImplemented, gin.H{"error": "MFA only available in local auth mode"})
			return
		}

		userID := c.GetString("user_id")
		au, err := lp.GetAuthUserByID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Utilisateur introuvable"})
			return
		}

		factors := []gin.H{}
		if au.TOTPFactorID != "" {
			status := "unverified"
			if au.TOTPEnabled {
				status = "verified"
			}
			factors = append(factors, gin.H{
				"id":            au.TOTPFactorID,
				"status":        status,
				"friendly_name": au.TOTPFriendlyName,
			})
		}

		log.Printf("[MFA] list_factors user=%s count=%d", userID, len(factors))
		c.JSON(http.StatusOK, gin.H{"totp": factors})
	}
}

// MFAEnrollHandler handles POST /api/v1/auth/mfa/enroll (protected).
// Generates a new TOTP secret and stores it unverified.
func MFAEnrollHandler(provider authprovider.AuthProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		lp, ok := getLocalProvider(provider)
		if !ok {
			c.JSON(http.StatusNotImplemented, gin.H{"error": "MFA only available in local auth mode"})
			return
		}

		userID := c.GetString("user_id")

		var req struct {
			FactorType   string `json:"factor_type"`
			FriendlyName string `json:"friendly_name"`
		}
		_ = c.ShouldBindJSON(&req)
		if req.FriendlyName == "" {
			req.FriendlyName = "SaferCloud Authenticator"
		}

		// Get user email for the OTP URI issuer label
		au, err := lp.GetAuthUserByID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Utilisateur introuvable"})
			return
		}

		// Block re-enrollment if already verified
		if au.TOTPEnabled {
			log.Printf("[MFA] enroll_conflict user=%s already_verified=true", userID)
			c.JSON(http.StatusConflict, gin.H{"error": "MFA already active. Disable it before re-enrolling."})
			return
		}

		// Idempotency: return existing pending factor instead of silently overwriting it
		if au.TOTPFactorID != "" {
			log.Printf("[MFA] enroll_idempotent user=%s factor=%s", userID, au.TOTPFactorID)
			c.JSON(http.StatusOK, gin.H{
				"id": au.TOTPFactorID,
				"totp": gin.H{
					"secret": au.TOTPSecret,
				},
			})
			return
		}

		factorID, _, secret, err := lp.StartTOTPEnrollment(userID, au.Email, req.FriendlyName)
		if err != nil {
			log.Printf("[MFA] enroll_error user=%s err=%v", userID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start MFA enrollment"})
			return
		}

		log.Printf("[MFA] enrolled user=%s factor=%s", userID, factorID)
		c.JSON(http.StatusOK, gin.H{
			"id": factorID,
			"totp": gin.H{
				"secret": secret,
			},
		})
	}
}

// MFAChallengeHandler handles POST /api/v1/auth/mfa/challenge (protected).
// Returns a random challenge ID. TOTP is time-based so no server state is needed.
func MFAChallengeHandler(provider authprovider.AuthProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := getLocalProvider(provider); !ok {
			c.JSON(http.StatusNotImplemented, gin.H{"error": "MFA only available in local auth mode"})
			return
		}

		b := make([]byte, 16)
		rand.Read(b)
		challengeID := hex.EncodeToString(b)

		log.Printf("[MFA] challenge_created user=%s", c.GetString("user_id"))
		c.JSON(http.StatusOK, gin.H{"id": challengeID})
	}
}

// MFAVerifyHandler handles POST /api/v1/auth/mfa/verify (protected).
// Validates a TOTP code, activates the factor if unverified, and returns an AAL2 JWT.
func MFAVerifyHandler(provider authprovider.AuthProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		lp, ok := getLocalProvider(provider)
		if !ok {
			c.JSON(http.StatusNotImplemented, gin.H{"error": "MFA only available in local auth mode"})
			return
		}

		userID := c.GetString("user_id")

		var req struct {
			FactorID    string `json:"factor_id"`
			ChallengeID string `json:"challenge_id"`
			Code        string `json:"code" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Code requis"})
			return
		}

		// Validate code format: must be exactly 6 decimal digits
		if len(req.Code) != 6 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Le code doit être composé de 6 chiffres"})
			return
		}
		for _, ch := range req.Code {
			if ch < '0' || ch > '9' {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Le code doit être composé de 6 chiffres"})
				return
			}
		}

		au, err := lp.GetAuthUserByID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Utilisateur introuvable"})
			return
		}

		if err := lp.ValidateTOTPCode(userID, req.Code); err != nil {
			log.Printf("[MFA] verify_failed user=%s err=%v", userID, err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Code invalide"})
			return
		}

		// Activate the factor if this is the enrollment verification
		if !au.TOTPEnabled {
			if err := lp.ActivateTOTP(userID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to activate MFA"})
				return
			}
			// Keep user_security_settings in sync — mark MFA as enabled/verified
			if err := lp.SyncMFAStatus(userID, true); err != nil {
				log.Printf("[MFA] sync_security_settings_error user=%s err=%v", userID, err)
			}
			log.Printf("[MFA] activated user=%s factor=%s", userID, au.TOTPFactorID)
		}

		// Issue a new AAL2 token
		token, err := lp.GenerateTokenWithAAL(userID, au.Email, "aal2")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		log.Printf("[MFA] verify_success user=%s aal=aal2", userID)
		c.JSON(http.StatusOK, sessionResponse(token, userID, au.Email))
	}
}

// MFAUnenrollHandler handles DELETE /api/v1/auth/mfa/unenroll (protected, requires AAL2).
// Removes the TOTP factor. The caller must have verified TOTP via /mfa/verify first.
// All previously issued tokens are revoked because their AAL2 claim is no longer valid.
func MFAUnenrollHandler(provider authprovider.AuthProvider, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		lp, ok := getLocalProvider(provider)
		if !ok {
			c.JSON(http.StatusNotImplemented, gin.H{"error": "MFA only available in local auth mode"})
			return
		}

		userID := c.GetString("user_id")

		au, err := lp.GetAuthUserByID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Utilisateur introuvable"})
			return
		}

		// Removing an active (verified) factor requires AAL2 — the caller must have
		// completed /mfa/verify immediately before. Removing a pending (unverified)
		// factor is always allowed (cleanup during enrollment).
		if au.TOTPEnabled && c.GetString("aal") != "aal2" {
			c.JSON(http.StatusForbidden, gin.H{"error": "MFA verification required before unenrolling"})
			return
		}

		factorID := au.TOTPFactorID

		if err := lp.DisableTOTP(userID); err != nil {
			log.Printf("[MFA] unenroll_error user=%s err=%v", userID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disable MFA"})
			return
		}

		// Keep user_security_settings in sync — mark MFA as disabled
		if err := lp.SyncMFAStatus(userID, false); err != nil {
			log.Printf("[MFA] sync_security_settings_error user=%s err=%v", userID, err)
		}

		// Revoke all previously issued tokens — any AAL2 claim they carry is no longer valid.
		if redisClient != nil {
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				redisClient.Set(ctx, "token_revoke:"+userID, strconv.FormatInt(time.Now().Unix(), 10), 7*24*time.Hour)
			}()
		}

		log.Printf("[MFA] unenrolled user=%s factor=%s", userID, factorID)
		c.JSON(http.StatusOK, gin.H{"id": factorID})
	}
}
