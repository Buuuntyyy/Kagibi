package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"strings"
)

func AuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(401, gin.H{"error": "Token manquant"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Supabase utilise HS256 avec votre JWT SECRET
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(401, gin.H{"error": "Token invalide"})
			return
		}

		// L'ID utilisateur (sub) est dans les claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// UUID de l'utilisateur
			c.Set("userID", claims["sub"])
		} else {
			c.AbortWithStatusJSON(401, gin.H{"error": "Claims invalides"})
		}

		c.Next()
	}
}
