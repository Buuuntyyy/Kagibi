package main

import (
	"github.com/joho/godotenv"
	"safercloud/backend/middleware"
	"github.com/gin-gonic/gin"
	"safercloud/backend/pkg"
	"log"
	"net/http"
	"safercloud/backend/handlers/auth"
	"safercloud/backend/handlers/users"
	"safercloud/backend/handlers/files"
)

func main() {
	// Charge les variables d'environnement
	if err := godotenv.Load(); err != nil {
		log.Fatal("Erreur lors du chargement du fichier .env")
	}
	// Initialise la connexion à la base de données
	db := pkg.NewDB()

	// Exécute les migrations
	err := pkg.Migrate(db)
	if err != nil {
		log.Fatalf("Failed to migrate: %v", err)
	}

	log.Println("Migrations executed successfully!")


	// Crée une instance de Gin (equivalent à Express.js en Node.js)
	router := gin.Default()
	router.Use(middleware.SecureHeaders())
	router.Use(middleware.RateLimiter())

	// ROUTES PUBLIQUES (Non protégées par l'authentification)
	publicRoutes := router.Group("/api/v1")
	{
		publicRoutes.POST("/auth/register", func(c *gin.Context) { auth.RegisterHandler(c, db) })
		publicRoutes.POST("/auth/login", func(c *gin.Context) { auth.LoginHandler(c, db) })
	}

	// ROUTES PROTÉGÉES (Protégées par l'authentification JWT)
	protectedRoutes := router.Group("/api/v1")
	protectedRoutes.Use(middleware.AuthMiddleware())
	{
		// ROUTES UTILISATEURS
		protectedRoutes.GET("/users", func(c *gin.Context) { users.ListUsersHandler(c, db) })
	}

	// ROUTES FICHIERS
	fileRoutes := protectedRoutes.Group("/files")
	{
		fileRoutes.POST("/upload", func(c *gin.Context) { files.UploadHandler(c, db) })
		fileRoutes.GET("/", func(c *gin.Context) { files.ListFilesHandler(c, db) })
		fileRoutes.DELETE("/:fileID", func(c *gin.Context) { files.DeleteFileHandler(c, db) })
		fileRoutes.DELETE("/:folderID", func(c *gin.Context) { files.DeleteFolderHandler(c, db) })
		fileRoutes.GET("/download/:fileID", func(c *gin.Context) { files.DownloadFileHandler(c, db) })
	}	

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

	router.POST("/files", func(c *gin.Context) {
		var file pkg.File
		if err := c.ShouldBindJSON(&file); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// TODO: persist the file using pkg (e.g., pkg.CreateFile) if implemented
		c.JSON(http.StatusCreated, file)
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