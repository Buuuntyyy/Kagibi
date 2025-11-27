package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"safercloud/backend/handlers/auth"
	"safercloud/backend/handlers/files"
	"safercloud/backend/handlers/folders"
	"safercloud/backend/handlers/tags"
	"safercloud/backend/handlers/users"
	"safercloud/backend/middleware"
	"safercloud/backend/pkg"

	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
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
	config.ExposeHeaders = []string{"Content-Length"}
	config.AllowCredentials = true
	config.MaxAge = 12 * time.Hour
	router.Use(cors.New(config))

	router.Use(middleware.SecureHeaders())
	router.Use(middleware.RateLimitMiddleware())

	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
		log.Fatalf("Impossible de se connecter à Redis: %v", err)
	}

	api := router.Group("/api/v1")
	// ROUTES PUBLIQUES (Non protégées par l'authentification)
	publicRoutes := api.Group("/auth")
	{
		publicRoutes.POST("/register", func(c *gin.Context) { auth.RegisterHandler(c, db) })
		publicRoutes.POST("/login", func(c *gin.Context) { auth.LoginHandler(c, db, redisClient) })
		publicRoutes.POST("/recovery/init", func(c *gin.Context) { auth.RecoveryInitHandler(c, db) })
		publicRoutes.POST("/recovery/finish", func(c *gin.Context) { auth.RecoveryFinishHandler(c, db) })
	}

	// ROUTES PROTÉGÉES (Protégées par l'authentification JWT)
	protectedRoutes := api.Group("/")
	protectedRoutes.Use(middleware.AuthMiddleware(redisClient))
	{
		protectedRoutes.POST("/auth/logout", func(c *gin.Context) { auth.LogoutHandler(c, redisClient) })

		userRoutes := protectedRoutes.Group("/users")
		// ROUTES UTILISATEURS
		userRoutes.GET("/", func(c *gin.Context) { users.ListUsersHandler(c, db) })
		userRoutes.GET("/me", func(c *gin.Context) { users.MeHandler(c, db) })
		userRoutes.POST("/change-password", func(c *gin.Context) { users.UpdatePasswordHandler(c, db) })

		// ROUTES FICHIERS
		fileRoutes := protectedRoutes.Group("/files")
		{
			fileRoutes.POST("/upload", func(c *gin.Context) { files.UploadHandler(c, db) })
			fileRoutes.GET("/list/*path", func(c *gin.Context) { files.ListFilesHandler(c, db) })
			fileRoutes.POST("/bulk-delete", func(c *gin.Context) { files.BulkDeleteHandler(c, db) })
			fileRoutes.DELETE("/file/:fileID", func(c *gin.Context) { files.DeleteFileHandler(c, db) })
			fileRoutes.DELETE("/folder/:folderID", func(c *gin.Context) { files.DeleteFolderHandler(c, db) })
			fileRoutes.POST("/move", func(c *gin.Context) { files.MoveHandler(c, db) })
			fileRoutes.POST("/rename", func(c *gin.Context) { files.RenameHandler(c, db) })
			fileRoutes.POST("/tags", func(c *gin.Context) { files.UpdateTagsHandler(c, db) })
			fileRoutes.GET("/download/:fileID", func(c *gin.Context) { files.DownloadFileHandler(c, db) })
		}

		// ROUTES DOSSIERS
		folderRoutes := protectedRoutes.Group("/folders")
		{
			folderRoutes.POST("/create", func(c *gin.Context) { folders.CreateHandler(c, db) })
		}

		// ROUTES TAGS
		tagRoutes := protectedRoutes.Group("/tags")
		{
			tagRoutes.GET("/", tags.ListTagsHandler(db))
			tagRoutes.POST("/", tags.CreateTagHandler(db))
			tagRoutes.DELETE("/:id", tags.DeleteTagHandler(db))
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
