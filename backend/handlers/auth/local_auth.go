// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package auth

import (
	"context"
	"html"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"kagibi/backend/pkg"
	"kagibi/backend/pkg/authprovider"
	"kagibi/backend/pkg/monitoring"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	"github.com/uptrace/bun"
)

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type signupRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type updateAuthPasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type updateAuthEmailRequest struct {
	NewEmail string `json:"new_email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func getLocalProvider(provider authprovider.AuthProvider) (*authprovider.LocalProvider, bool) {
	lp, ok := provider.(*authprovider.LocalProvider)
	return lp, ok
}

func sessionResponse(token, userID, email string) gin.H {
	return gin.H{
		"access_token": token,
		"token_type":   "Bearer",
		"user": gin.H{
			"id":    userID,
			"email": email,
		},
	}
}

// LocalLoginHandler handles POST /api/v1/auth/login (public).
func LocalLoginHandler(provider authprovider.AuthProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		lp, ok := getLocalProvider(provider)
		if !ok {
			c.JSON(http.StatusNotImplemented, gin.H{"error": "Login endpoint only available in local auth mode"})
			return
		}

		var req loginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email et mot de passe requis"})
			return
		}

		au, err := lp.FindAuthUserByEmail(req.Email)
		if err != nil {
			monitoring.RecordUserLogin(false)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Identifiants invalides"})
			return
		}

		if err := lp.CheckPassword(au.PasswordHash, req.Password); err != nil {
			monitoring.RecordUserLogin(false)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Identifiants invalides"})
			return
		}

		token, err := lp.GenerateToken(au.ID, au.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la génération du token"})
			return
		}
		monitoring.UserLoginsTotal.WithLabelValues("success").Inc()
		monitoring.RecordUserLogin(true)
		c.JSON(http.StatusOK, sessionResponse(token, au.ID, au.Email))
	}
}

// LocalSignupHandler handles POST /api/v1/auth/signup (public).
// Creates only the auth_users record. Profile creation is done separately via POST /auth/register.
//
// Orphan recovery: if auth_users already exists for this email but no profile was ever created
// (e.g. the registration flow was interrupted), the handler verifies the password and returns
// a fresh token so the client can retry POST /auth/register — instead of returning 409 forever.
func LocalSignupHandler(provider authprovider.AuthProvider, db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		lp, ok := getLocalProvider(provider)
		if !ok {
			c.JSON(http.StatusNotImplemented, gin.H{"error": "Signup endpoint only available in local auth mode"})
			return
		}

		var req signupRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userID, err := lp.CreateAuthUser(req.Email, req.Password)
		if err != nil {
			if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
				// Check if this is an orphaned auth_users row (signup succeeded but
				// profile creation failed). If so, let the user retry registration
				// rather than blocking them with a permanent 409.
				if token, id, recovered := recoverOrphanedSignup(lp, db, req.Email, req.Password); recovered {
					c.JSON(http.StatusCreated, sessionResponse(token, id, req.Email))
					return
				}
				c.JSON(http.StatusConflict, gin.H{"error": "Un compte avec cet email existe déjà"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la création du compte"})
			return
		}

		token, err := lp.GenerateToken(userID, req.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la génération du token"})
			return
		}

		c.JSON(http.StatusCreated, sessionResponse(token, userID, req.Email))
	}
}

// recoverOrphanedSignup handles the case where auth_users exists but profiles does not.
// It verifies the password, then returns a fresh token so registration can be completed.
// Returns (token, userID, true) on success, ("", "", false) otherwise.
func recoverOrphanedSignup(lp *authprovider.LocalProvider, db *bun.DB, email, password string) (string, string, bool) {
	userID, token, err := lp.ReissueToken(email, password)
	if err != nil {
		return "", "", false
	}

	// Profile exists → real duplicate, not an orphan.
	if _, err := pkg.FindUserByID(db, userID); err == nil {
		return "", "", false
	}

	log.Printf("[Signup/local] Orphaned auth_users detected for email=%s — issuing recovery token", email)
	return token, userID, true
}

// LocalRefreshHandler handles POST /api/v1/auth/refresh (public — current token in Authorization header).
// Issues a new 7-day token if the current token signature is valid (even if near expiry).
// Checks Redis for token revocation (e.g. post password-change) before issuing a new token.
// isTokenRevoked checks Redis to determine whether the given token (identified
// by its iat claim) was issued before a revocation event for the user.
// Returns true if the token should be rejected. A missing Redis key means no
// revocation has been recorded — the token is considered valid.
func isTokenRevoked(redisClient *redis.Client, userID string, claims jwt.MapClaims) bool {
	if redisClient == nil {
		return false
	}
	iatFloat, ok := claims["iat"].(float64)
	if !ok {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	revokeStr, rErr := redisClient.Get(ctx, "token_revoke:"+userID).Result()
	cancel()
	if rErr != nil {
		return false
	}
	revokeTs, parseErr := strconv.ParseInt(revokeStr, 10, 64)
	return parseErr == nil && int64(iatFloat) < revokeTs
}

// parseLocalToken parses and validates a JWT using the LocalProvider secret,
// returning the token claims on success.
func parseLocalToken(lp interface{ GetJWTSecret() []byte }, tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenUnverifiable
		}
		return lp.GetJWTSecret(), nil
	})
	if err != nil || !token.Valid {
		return nil, jwt.ErrTokenSignatureInvalid
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}

func LocalRefreshHandler(provider authprovider.AuthProvider, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		lp, ok := getLocalProvider(provider)
		if !ok {
			c.JSON(http.StatusNotImplemented, gin.H{"error": "Refresh endpoint only available in local auth mode"})
			return
		}

		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token manquant"})
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := parseLocalToken(lp, tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token invalide ou expiré"})
			return
		}

		userID, _ := claims["sub"].(string)
		email, _ := claims["email"].(string)
		aal, _ := claims["aal"].(string)
		if aal == "" {
			aal = "aal1"
		}

		// Reject refresh if the token was issued before a password change or MFA disable.
		if isTokenRevoked(redisClient, userID, claims) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token révoqué"})
			return
		}

		newToken, err := lp.GenerateTokenWithAAL(userID, email, aal)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors du renouvellement du token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"access_token": newToken,
			"token_type":   "Bearer",
		})
	}
}

// LocalUpdatePasswordHandler handles POST /api/v1/auth/update-password (protected — JWT required).
// Verifies the old password before updating the hash, then revokes all previously issued tokens.
func LocalUpdatePasswordHandler(provider authprovider.AuthProvider, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		lp, ok := getLocalProvider(provider)
		if !ok {
			c.JSON(http.StatusNotImplemented, gin.H{"error": "Update password only available in local auth mode"})
			return
		}

		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Non authentifié"})
			return
		}

		var req updateAuthPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := lp.UpdateUserPasswordWithVerification(userID, req.OldPassword, req.NewPassword); err != nil {
			if err.Error() == "invalid current password" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Mot de passe actuel incorrect"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la mise à jour du mot de passe"})
			return
		}

		// Revoke all tokens issued before this moment — they authenticated with the old password.
		if redisClient != nil {
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				redisClient.Set(ctx, "token_revoke:"+userID, strconv.FormatInt(time.Now().Unix(), 10), 7*24*time.Hour)
			}()
		}

		c.JSON(http.StatusOK, gin.H{"message": "Mot de passe mis à jour avec succès"})
	}
}

// LocalUpdateEmailHandler handles PUT /api/v1/auth/update-email (protected — JWT required).
// Verifies the current password before updating the email in both auth_users and profiles tables,
// then issues a fresh JWT embedding the new email.
func LocalUpdateEmailHandler(provider authprovider.AuthProvider, db *bun.DB, redisClient *redis.Client) gin.HandlerFunc {
	emailRe := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return func(c *gin.Context) {
		lp, ok := getLocalProvider(provider)
		if !ok {
			c.JSON(http.StatusNotImplemented, gin.H{"error": "Update email only available in local auth mode"})
			return
		}

		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Non authentifié"})
			return
		}

		var req updateAuthEmailRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		newEmail := strings.ToLower(strings.TrimSpace(req.NewEmail))
		if !emailRe.MatchString(newEmail) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Adresse email invalide"})
			return
		}
		// Defense in depth: HTML-escape to neutralize any residual special chars
		// (the regex already excludes < > " ' ; but we keep this for consistency with username handling)
		newEmail = html.EscapeString(newEmail)

		// Verify password and update auth_users.email
		if err := lp.UpdateUserEmailWithVerification(userID, req.Password, newEmail); err != nil {
			if err.Error() == "invalid current password" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Mot de passe incorrect"})
				return
			}
			if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
				c.JSON(http.StatusConflict, gin.H{"error": "Cette adresse email est déjà utilisée"})
				return
			}
			log.Printf("ERROR: Failed to update auth email for user %s: %v", userID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la mise à jour de l'email"})
			return
		}

		// Update profiles.email to keep tables in sync
		if _, err := db.NewUpdate().TableExpr("profiles").
			Set("email = ?", newEmail).
			Where("id = ?", userID).
			Exec(c.Request.Context()); err != nil {
			log.Printf("ERROR: Failed to update profile email for user %s: %v", userID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la mise à jour du profil"})
			return
		}

		// Revoke old tokens and issue a fresh one with the new email
		if redisClient != nil {
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				redisClient.Set(ctx, "token_revoke:"+userID, strconv.FormatInt(time.Now().Unix(), 10), 7*24*time.Hour)
			}()
		}

		// Fetch the current AAL from the old token claims so we preserve MFA state
		aal := "aal1"
		if authHeader := c.GetHeader("Authorization"); strings.HasPrefix(authHeader, "Bearer ") {
			if claims, err := parseLocalToken(lp, strings.TrimPrefix(authHeader, "Bearer ")); err == nil {
				if v, ok2 := claims["aal"].(string); ok2 && v != "" {
					aal = v
				}
			}
		}

		newToken, err := lp.GenerateTokenWithAAL(userID, newEmail, aal)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la génération du token"})
			return
		}

		log.Printf("INFO: Email updated - UserID: %s, NewEmail: %s", userID, newEmail)
		c.JSON(http.StatusOK, gin.H{
			"message":      "Email mis à jour avec succès",
			"email":        newEmail,
			"access_token": newToken,
		})
	}
}
