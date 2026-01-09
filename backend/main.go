package main

import (
	"context"
	"log"
	"os"
	"safercloud/backend/handlers/auth"
	"safercloud/backend/handlers/files"
	"safercloud/backend/handlers/folders"
	"safercloud/backend/handlers/friends"
	"safercloud/backend/handlers/keys"
	"safercloud/backend/handlers/shares"
	"safercloud/backend/handlers/tags"
	"safercloud/backend/handlers/users"
	"safercloud/backend/handlers/ws"
	"safercloud/backend/middleware"
	"safercloud/backend/pkg"
	"safercloud/backend/pkg/s3storage"
	"safercloud/backend/pkg/workers"
	wsPkg "safercloud/backend/pkg/ws"

	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

func main() {
	log.Println("Starting SaferCloud Backend v2.1 (With Share Keys Fix)...")

	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize S3
	if err := s3storage.InitS3(); err != nil {
		log.Printf("Warning: S3 not configured: %v", err)
	} else {
		log.Println("S3 Storage initialized successfully")
	}

	// Initialise la connexion à la base de données
	db := pkg.NewDB()

	// Exécute les migrations
	err := pkg.Migrate(db)
	if err != nil {
		log.Printf("Failed to migrate: %v", err) // Printf instead of Fatalf strictly for safety in this context
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

	redisURL := os.Getenv("REDIS_URL")
	var redisOptions *redis.Options

	if redisURL != "" {
		// Si une URL complète est fournie (ex: rediss://user:pass@host:port)
		var err error
		redisOptions, err = redis.ParseURL(redisURL)
		if err != nil {
			log.Fatalf("Invalid REDIS_URL: %v", err)
		}
	} else {
		// Fallback local
		redisOptions = &redis.Options{
			Addr: "localhost:6379",
		}
	}

	redisClient := redis.NewClient(redisOptions)

	if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
		log.Fatalf("Impossible de se connecter à Redis: %v", err)
	}

	// Start S3 Worker
	workers.StartWorker(redisClient)

	// Start Cleanup Worker (Expired shares, etc.)
	workers.StartCleanupWorker(db)

	// Initialize WebSocket Manager
	wsManager := wsPkg.NewManager()

	friendHandler := friends.NewFriendHandler(db, wsManager)

	api := router.Group("/api/v1")
	// ROUTES PUBLIQUES (Non protégées par l'authentification)
	publicRoutes := api.Group("/auth")
	{
		publicRoutes.POST("/register", func(c *gin.Context) { auth.RegisterHandler(c, db) })
		publicRoutes.POST("/login", func(c *gin.Context) { auth.LoginHandler(c, db, redisClient) })
		publicRoutes.POST("/recovery/init", func(c *gin.Context) { auth.RecoveryInitHandler(c, db) })
		publicRoutes.POST("/recovery/finish", func(c *gin.Context) { auth.RecoveryFinishHandler(c, db) })
	}

	// ROUTES PUBLIQUES DE PARTAGE
	publicShareRoutes := api.Group("/public/share")
	{
		publicShareRoutes.GET("/:token", func(c *gin.Context) { shares.GetShareLinkHandler(c, db) })
		publicShareRoutes.GET("/:token/download", func(c *gin.Context) { shares.DownloadSharedFileHandler(c, db) })
		publicShareRoutes.GET("/:token/download/file/:file_id", func(c *gin.Context) { shares.DownloadFileFromSharedFolderHandler(c, db) })
		publicShareRoutes.GET("/:token/browse/*subpath", func(c *gin.Context) { shares.BrowseSharedFolderHandler(c, db) })
	}

	// ROUTES PROTÉGÉES (Protégées par l'authentification JWT)
	protectedRoutes := api.Group("")
	protectedRoutes.Use(middleware.AuthMiddleware(redisClient))
	{
		protectedRoutes.POST("/auth/logout", func(c *gin.Context) { auth.LogoutHandler(c, redisClient) })

		userRoutes := protectedRoutes.Group("/users")
		// ROUTES UTILISATEURS
		userRoutes.GET("/", func(c *gin.Context) { users.ListUsersHandler(c, db) })
		userRoutes.GET("/me", func(c *gin.Context) { users.MeHandler(c, db) })
		userRoutes.POST("/change-password", func(c *gin.Context) { users.UpdatePasswordHandler(c, db) })
		userRoutes.POST("/keys", func(c *gin.Context) {
			// Lazy load to avoid cycle, or just call handler directly
			// assuming 'keys' pkg doesn't import main
			// Actually need to import it. I will fix imports later if needed.
			// For now let's assume 'keys' package is available.
			// Ideally should be "safercloud/backend/handlers/keys"
			keys.UpdateKeysHandler(c, db)
		})

		// ROUTES FICHIERS
		fileRoutes := protectedRoutes.Group("/files")
		{
			fileRoutes.POST("/upload", func(c *gin.Context) { files.UploadHandler(c, db, redisClient, wsManager) })
			fileRoutes.GET("/list-recursive", func(c *gin.Context) { files.ListAllFilesRecursiveHandler(c, db) })
			fileRoutes.GET("/list/*path", func(c *gin.Context) { files.ListFilesHandler(c, db) })
			fileRoutes.POST("/bulk-delete", func(c *gin.Context) { files.BulkDeleteHandler(c, db, wsManager) })
			fileRoutes.DELETE("/file/:fileID", func(c *gin.Context) { files.DeleteFileHandler(c, db, wsManager) })
			fileRoutes.DELETE("/folder/:folderID", func(c *gin.Context) { files.DeleteFolderHandler(c, db, wsManager) })
			fileRoutes.POST("/move", func(c *gin.Context) { files.MoveHandler(c, db, redisClient) })
			fileRoutes.POST("/rename", func(c *gin.Context) { files.RenameHandler(c, db, redisClient) })
			fileRoutes.POST("/tags", func(c *gin.Context) { files.UpdateTagsHandler(c, db) })
			fileRoutes.GET("/download/:fileID", func(c *gin.Context) { files.DownloadFileHandler(c, db) })
			fileRoutes.GET("/search", func(c *gin.Context) { files.SearchFilesHandler(c, db) })
		}

		// ROUTES DOSSIERS
		folderRoutes := protectedRoutes.Group("/folders")
		{
			folderRoutes.POST("/create", func(c *gin.Context) { folders.CreateHandler(c, db) })
			folderRoutes.PUT("/:id/key", func(c *gin.Context) { folders.UpdateFolderKeyHandler(c, db) })
		}

		// ROUTES TAGS
		tagRoutes := protectedRoutes.Group("/tags")
		{
			tagRoutes.GET("/", tags.ListTagsHandler(db))
			tagRoutes.POST("/", tags.CreateTagHandler(db))
			tagRoutes.DELETE("/:id", tags.DeleteTagHandler(db))
		}

		// ROUTES AMIS
		friendRoutes := protectedRoutes.Group("/friends")
		{
			friendRoutes.GET("", friendHandler.ListFriends)
			friendRoutes.POST("", friendHandler.AddFriend)
			friendRoutes.DELETE("/:id", friendHandler.RemoveFriend)
			friendRoutes.PUT("/:id/accept", friendHandler.AcceptFriend)
			friendRoutes.DELETE("/:id/reject", friendHandler.RejectFriend)
		}

		// ROUTES PARTAGE
		shareRoutes := protectedRoutes.Group("/shares")
		{
			shareRoutes.GET("/list", func(c *gin.Context) { shares.ListSharesHandler(c, db) })
			shareRoutes.POST("/link", func(c *gin.Context) {
				log.Println("DEBUG: Hit /shares/link route")
				shares.CreateShareLinkHandler(c, db)
			})
			shareRoutes.POST("/direct", func(c *gin.Context) { shares.CreateDirectShareHandler(c, db, wsManager) })
			shareRoutes.GET("/direct", func(c *gin.Context) { shares.ListDirectSharesForResourceHandler(c, db) })
			shareRoutes.DELETE("/direct", func(c *gin.Context) { shares.RemoveDirectShareHandler(c, db, wsManager) })
			shareRoutes.GET("/check-path", func(c *gin.Context) { shares.GetActiveSharesForPathHandler(c, db) })
			shareRoutes.GET("/file/:fileID", func(c *gin.Context) { shares.GetShareForResourceHandler(c, db) })
			shareRoutes.GET("/direct/folder/:folderID/content", func(c *gin.Context) { shares.GetSharedFolderContentHandler(c, db) })
			shareRoutes.DELETE("/link/:shareID", func(c *gin.Context) { shares.DeleteShareLinkHandler(c, db) })

			// Shared With Me Routes
			shareRoutes.GET("/with-me", func(c *gin.Context) { shares.ListImportedSharesHandler(c, db) })
			shareRoutes.POST("/with-me", func(c *gin.Context) { shares.ImportShareHandler(c, db) })
			shareRoutes.DELETE("/with-me/:id", func(c *gin.Context) { shares.RemoveImportedShareHandler(c, db, wsManager) })
		}
	}

	// Route WebSocket (Racine)
	router.GET("/ws", func(c *gin.Context) { ws.ConnectHandler(c, wsManager, redisClient) })

	// Route de debug
	router.GET("/api/v1/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong", "version": "2.2"})
	})

	// Démarre le serveur sur le port 8080
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
