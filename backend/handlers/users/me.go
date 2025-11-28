package users

import (
	"net/http"
	"safercloud/backend/pkg"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// Structure pour la réponse JSON, pour ne pas exposer le mot de passe hashé
type UserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func MeHandler(c *gin.Context, db *bun.DB) {
	// 1. Récupérer l'ID utilisateur placé dans le contexte par le middleware
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Contexte utilisateur non trouvé"})
		return
	}

	// 2. Convertir l'ID en string
	userID, ok := userIDInterface.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Format d'ID utilisateur invalide"})
		return
	}

	// 3. Récupérer les informations de l'utilisateur depuis la base de données
	user, err := pkg.FindUserByID(db, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Utilisateur non trouvé"})
		return
	}

	// 4. Renvoyer une réponse formatée et sécurisée
	// Note: On renvoie l'objet user complet (sauf password qui n'est pas exporté en JSON si on utilisait le struct User directement, mais ici on mappe manuellement)
	// Pour l'instant on garde la structure UserResponse mais on pourrait renvoyer user directement si on veut tous les champs (storage, plan, etc)
	// Le frontend semble attendre l'objet complet pour le storage.
	// On va renvoyer l'objet user complet car le frontend l'utilise pour le storage.

	c.JSON(http.StatusOK, user)
}
