package main

import (
	"github.com/gin-gonic/gin"
	"github.com/buuuntyyy/safercloud/backend/pkg"
	"context"
	"log"
	"net/http"
)

func main() {
	db := internal.NewDB()

	// Exécute les migrations
	err := internal.Migrate(db)
	if err != nil {
		log.Fatalf("Failed to migrate: %v", err)
	}

	log.Println("Migrations executed successfully!")

	// Crée une instance de Gin (equivalent à Express.js en Node.js)
	router := gin.Default()

	// Définis une route GET
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello, Gin!",
		})
	})

	// Route POST avec un paramètre JSON
	router.POST("/users", func(c *gin.Context) {
		var user struct {
			Name  string `json:"name" binding:"required"`
			Email string `json:"email" binding:"required"`
		}

		// Parse le JSON de la requête
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Retourne une réponse
		c.JSON(http.StatusCreated, gin.H{
			"user": user,
		})
	})

	// Route pour lister les utilisateurs
	router.GET("/users", func(c *gin.Context) {
		users, err := internal.ListUsers(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, users)
	})

	// Démarre le serveur sur le port 8080
	router.Run(":8080")
}
