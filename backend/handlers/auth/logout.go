package auth

import (
	"context"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func LogoutHandler(c *gin.Context, redisClient *redis.Client) {
	// 1. Récupérer l'ID de session depuis le cookie
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		// S'il n'y a pas de cookie, l'utilisateur n'est pas connecté.
		// On peut simplement renvoyer OK.
		c.JSON(http.StatusOK, gin.H{"message": "Déjà déconnecté"})
		return
	}

	// 2. Supprimer la session de Redis
	redisClient.Del(context.Background(), sessionID)

	domain := os.Getenv("COOKIE_DOMAIN")
	if domain == "" {
		domain = "localhost"
	}

	// 3. Dire au navigateur de supprimer le cookie en lui donnant une date d'expiration passée
	c.SetCookie("session_id", "", -1, "/", domain, true, true)

	c.JSON(http.StatusOK, gin.H{"message": "Déconnexion réussie"})
}
