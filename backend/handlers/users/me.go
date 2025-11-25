package users

import (
	"net/http"
	"safercloud/backend/pkg"
	"strconv"

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
    userIDStr, ok := userIDInterface.(string)
    if !ok {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Format d'ID utilisateur invalide"})
        return
    }

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
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
    response := UserResponse{
        ID:    strconv.FormatInt(user.ID, 10),
        Name:  user.Name,
        Email: user.Email,
    }

    c.JSON(http.StatusOK, response)
}