package auth

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"safercloud/backend/pkg/authprovider"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Identifiants invalides"})
			return
		}

		if err := lp.CheckPassword(au.PasswordHash, req.Password); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Identifiants invalides"})
			return
		}

		token, err := lp.GenerateToken(au.ID, au.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la génération du token"})
			return
		}

		c.JSON(http.StatusOK, sessionResponse(token, au.ID, au.Email))
	}
}

// LocalSignupHandler handles POST /api/v1/auth/signup (public).
// Creates only the auth_users record. Profile creation is done separately via POST /auth/register.
func LocalSignupHandler(provider authprovider.AuthProvider) gin.HandlerFunc {
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

// LocalRefreshHandler handles POST /api/v1/auth/refresh (public — current token in Authorization header).
// Issues a new 7-day token if the current token signature is valid (even if near expiry).
// Checks Redis for token revocation (e.g. post password-change) before issuing a new token.
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

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenUnverifiable
			}
			return lp.GetJWTSecret(), nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token invalide ou expiré"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Claims invalides"})
			return
		}

		userID, _ := claims["sub"].(string)
		email, _ := claims["email"].(string)
		aal, _ := claims["aal"].(string)
		if aal == "" {
			aal = "aal1"
		}

		// Reject refresh if the token was issued before a password change or MFA disable.
		if redisClient != nil {
			if iatFloat, ok := claims["iat"].(float64); ok {
				ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
				revokeStr, rErr := redisClient.Get(ctx, "token_revoke:"+userID).Result()
				cancel()
				if rErr == nil {
					if revokeTs, parseErr := strconv.ParseInt(revokeStr, 10, 64); parseErr == nil && int64(iatFloat) < revokeTs {
						c.JSON(http.StatusUnauthorized, gin.H{"error": "Token révoqué"})
						return
					}
				}
			}
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
