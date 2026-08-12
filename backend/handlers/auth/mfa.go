// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"strconv"
	"time"

	"kagibi/backend/pkg/authprovider"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

const (
	errMFALocalOnly = "MFA only available in local auth mode"
	errUserNotFound = "Utilisateur introuvable"
)

// MFAListFactorsHandler handles GET /api/v1/auth/mfa/factors (protected).
// Returns the user's TOTP factor(s), compatible with the Supabase MFA factors shape.
func MFAListFactorsHandler(provider authprovider.AuthProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		lp, ok := getLocalProvider(provider)
		if !ok {
			c.JSON(http.StatusNotImplemented, gin.H{"error": errMFALocalOnly})
			return
		}

		userID := c.GetString("user_id")
		au, err := lp.GetAuthUserByID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errUserNotFound})
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
			c.JSON(http.StatusNotImplemented, gin.H{"error": errMFALocalOnly})
			return
		}

		userID := c.GetString("user_id")

		var req struct {
			FactorType   string `json:"factor_type"`
			FriendlyName string `json:"friendly_name"`
		}
		_ = c.ShouldBindJSON(&req)
		if req.FriendlyName == "" {
			req.FriendlyName = "Kagibi Authenticator"
		}

		// Get user email for the OTP URI issuer label
		au, err := lp.GetAuthUserByID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errUserNotFound})
			return
		}

		// Block re-enrollment if already verified
		if au.TOTPEnabled {
			log.Printf("[MFA] enroll_conflict user=%s already_verified=true", userID)
			c.JSON(http.StatusConflict, gin.H{"error": "MFA already active. Disable it before re-enrolling."})
			return
		}

		// Idempotency: return existing pending factor instead of silently overwriting it.
		// The stored secret is encrypted at rest, so decrypt it for the QR re-display.
		if au.TOTPFactorID != "" {
			log.Printf("[MFA] enroll_idempotent user=%s factor=%s", userID, au.TOTPFactorID)
			c.JSON(http.StatusOK, gin.H{
				"id": au.TOTPFactorID,
				"totp": gin.H{
					"secret": lp.DecryptTOTPSecret(au.TOTPSecret),
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

// mfaChallengeTTL bounds how long an issued MFA challenge stays valid.
const mfaChallengeTTL = 10 * time.Minute

func mfaChallengeKey(userID, challengeID string) string {
	return "mfa_challenge:" + userID + ":" + challengeID
}

// peekMFAChallenge reports whether a challenge issued to userID is present.
// checked is false when the binding was skipped (no Redis, or a Redis error) — the
// caller then falls back to code-only verification to preserve availability. When
// Redis is reachable, an empty challenge ID is treated as an invalid (missing) one.
func peekMFAChallenge(redisClient *redis.Client, userID, challengeID string) (ok bool, checked bool) {
	if redisClient == nil {
		return false, false
	}
	if challengeID == "" {
		return false, true
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	n, err := redisClient.Exists(ctx, mfaChallengeKey(userID, challengeID)).Result()
	if err != nil {
		return false, false
	}
	return n > 0, true
}

// consumeMFAChallenge deletes a challenge so it cannot be reused after a successful verify.
func consumeMFAChallenge(redisClient *redis.Client, userID, challengeID string) {
	if redisClient == nil || challengeID == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = redisClient.Del(ctx, mfaChallengeKey(userID, challengeID)).Err()
}

// MFAChallengeHandler handles POST /api/v1/auth/mfa/challenge (protected).
// Issues a random challenge ID and binds it to the user in Redis with a short TTL, so
// /mfa/verify can require a freshly-issued challenge. Best-effort: a Redis outage does
// not block challenge creation (verification then falls back to code-only).
func MFAChallengeHandler(provider authprovider.AuthProvider, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := getLocalProvider(provider); !ok {
			c.JSON(http.StatusNotImplemented, gin.H{"error": errMFALocalOnly})
			return
		}

		userID := c.GetString("user_id")

		b := make([]byte, 16)
		rand.Read(b)
		challengeID := hex.EncodeToString(b)

		if redisClient != nil {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			if err := redisClient.Set(ctx, mfaChallengeKey(userID, challengeID), "1", mfaChallengeTTL).Err(); err != nil {
				log.Printf("[MFA] challenge_store_failed user=%s err=%v", userID, err)
			}
			cancel()
		}

		log.Printf("[MFA] challenge_created user=%s", userID)
		c.JSON(http.StatusOK, gin.H{"id": challengeID})
	}
}

// isSixDigitCode returns true if s is exactly 6 decimal digit characters.
func isSixDigitCode(s string) bool {
	if len(s) != 6 {
		return false
	}
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

// activateTOTPIfNeeded activates TOTP for the user if it is not already enabled.
func activateTOTPIfNeeded(lp *authprovider.LocalProvider, userID, factorID string, totpEnabled bool) error {
	if totpEnabled {
		return nil
	}
	if err := lp.ActivateTOTP(userID); err != nil {
		return err
	}
	if err := lp.SyncMFAStatus(userID, true); err != nil {
		log.Printf("[MFA] sync_security_settings_error user=%s err=%v", userID, err)
	}
	log.Printf("[MFA] activated user=%s factor=%s", userID, factorID)
	return nil
}

// MFAVerifyHandler handles POST /api/v1/auth/mfa/verify (protected).
// Validates a TOTP code against a freshly-issued challenge, activates the factor if
// unverified, and returns an AAL2 JWT.
func MFAVerifyHandler(provider authprovider.AuthProvider, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		lp, ok := getLocalProvider(provider)
		if !ok {
			c.JSON(http.StatusNotImplemented, gin.H{"error": errMFALocalOnly})
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

		if !isSixDigitCode(req.Code) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Le code doit être composé de 6 chiffres"})
			return
		}

		// Require a freshly-issued, user-scoped challenge. Peek (don't consume) so a
		// wrong code can be retried with the same challenge; it is consumed only on
		// success. If Redis is unavailable the binding is skipped (code-only fallback).
		challengeOK, challengeChecked := peekMFAChallenge(redisClient, userID, req.ChallengeID)
		if challengeChecked && !challengeOK {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Challenge invalide ou expiré"})
			return
		}

		au, err := lp.GetAuthUserByID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errUserNotFound})
			return
		}

		if err := lp.ValidateTOTPCode(userID, req.Code); err != nil {
			log.Printf("[MFA] verify_failed user=%s err=%v", userID, err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Code invalide"})
			return
		}

		// Code valid — consume the challenge so this (challenge, code) pair is single-use.
		if challengeChecked {
			consumeMFAChallenge(redisClient, userID, req.ChallengeID)
		}

		if err := activateTOTPIfNeeded(lp, userID, au.TOTPFactorID, au.TOTPEnabled); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to activate MFA"})
			return
		}

		// aal2 + mfa=enabled (so per-action gates keep enforcing once the elevation
		// goes stale) + mfa_at=now (step-up freshness anchor).
		token, err := lp.GenerateTokenWithClaims(userID, au.Email, "aal2", true, time.Now().Unix())
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
			c.JSON(http.StatusNotImplemented, gin.H{"error": errMFALocalOnly})
			return
		}

		userID := c.GetString("user_id")

		au, err := lp.GetAuthUserByID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errUserNotFound})
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
