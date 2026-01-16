package middleware

import (
	"log"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware vérifie les tokens via JWKS (ES256) ou Secret (HS256)
func AuthMiddleware(jwks keyfunc.Keyfunc, secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(401, gin.H{"error": "Token manquant"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Si le token est signé avec HMAC (HS256), on utilise le secret
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); ok {
				return []byte(secret), nil
			}

			// Si le token est signé avec ECDSA (ES256) et qu'on a le JWKS, on utilise le JWKS
			if _, ok := token.Method.(*jwt.SigningMethodECDSA); ok && jwks != nil {
				return jwks.Keyfunc(token)
			}

			// Sinon, méthode Inconnue
			return nil, jwt.ErrTokenUnverifiable
		})

		if err != nil || !token.Valid {
			log.Printf("Auth Error: %v", err) // Log pour débugger
			c.AbortWithStatusJSON(401, gin.H{"error": "Token invalide"})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("user_id", claims["sub"])
		} else {
			c.AbortWithStatusJSON(401, gin.H{"error": "Claims invalides"})
		}

		c.Next()
	}
}
