package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"safercloud/backend/handlers/auth"
	"safercloud/backend/handlers/files"
	"safercloud/backend/handlers/folders"
	"safercloud/backend/handlers/users"
	"safercloud/backend/middleware"
	"safercloud/backend/pkg"
)

func main() {

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

	// Configure et utilise le middleware CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173"} // Remplacez par l'URL de votre frontend si nécessaire
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	router.Use(cors.New(config))

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

		// ROUTES FICHIERS
		fileRoutes := protectedRoutes.Group("/files")
		{
			fileRoutes.POST("/upload", func(c *gin.Context) { files.UploadHandler(c, db) })
			fileRoutes.GET("/list/*path", func(c *gin.Context) { files.ListFilesHandler(c, db) })
			fileRoutes.DELETE("/file/:fileID", func(c *gin.Context) { files.DeleteFileHandler(c, db) })
			fileRoutes.DELETE("/folder/:folderID", func(c *gin.Context) { files.DeleteFolderHandler(c, db) })
			fileRoutes.GET("/download/:fileID", func(c *gin.Context) { files.DownloadFileHandler(c, db) })
		}

		// ROUTES DOSSIERS
		folderRoutes := protectedRoutes.Group("/folders")
		{
			folderRoutes.POST("/create", func(c *gin.Context) { folders.CreateHandler(c, db) })
		}
	}	

	// Définis une route GET
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello, Gin!",
		})
	})

	// Démarre le serveur sur le port 8080
	router.Run(":8080")
}