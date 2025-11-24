package main

import (
	"github.com/gin-gonic/gin"
	"safercloud/backend/pkg"
	"log"
	"net/http"
)

func main() {
	db := pkg.NewDB()

	// Exécute les migrations
	err := pkg.Migrate(db)
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
		var user pkg.User  // <-- Utilise le modèle pkg.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Appelle CreateUser avec le bon modèle
		err := pkg.CreateUser(db, &user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, user)
	})

	// Route pour lister les utilisateurs
	router.GET("/users", func(c *gin.Context) {
		users, err := pkg.ListUsers(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, users)
	})

	// Démarre le serveur sur le port 8080
	router.Run(":8080")
}
