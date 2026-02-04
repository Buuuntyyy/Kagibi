package main

import (
	"context"
	"log"
	"os"
	"safercloud/backend/handlers/auth"
	"safercloud/backend/handlers/billing"
	"safercloud/backend/handlers/files"
	"safercloud/backend/handlers/folders"
	"safercloud/backend/handlers/friends"
	"safercloud/backend/handlers/keys"
	"safercloud/backend/handlers/security"
	"safercloud/backend/handlers/shares"
	"safercloud/backend/handlers/tags"
	"safercloud/backend/handlers/users"
	"safercloud/backend/middleware"
	"safercloud/backend/pkg"
	"safercloud/backend/pkg/s3storage"
	"safercloud/backend/pkg/workers"
	"strings"

	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/uptrace/bun"
)

func main() {
	log.Println("Starting SaferCloud Backend v3.0 (Supabase Realtime)...")

	loadEnv()
	initS3()

	db := pkg.NewDB()
	migrateDB(db)

	redisClient := initRedis()

	// Start Workers
	workers.StartWorker(redisClient)
	workers.StartCleanupWorker(db)

	// Initialize Handlers (no more WebSocket Manager)
	friendHandler := friends.NewFriendHandler(db)
	jwks, jwtSecret := initAuth()

	// Setup Server
	router := setupRouter()

	registerRoutes(router, db, redisClient, jwks, jwtSecret, friendHandler)

	startServer(router)
}

func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
}

func initS3() {
	if err := s3storage.InitS3(); err != nil {
		log.Printf("Warning: S3 not configured: %v", err)
	} else {
		log.Println("S3 Storage initialized successfully")
	}
}

func migrateDB(db *bun.DB) {
	err := pkg.Migrate(db)
	if err != nil {
		log.Printf("Failed to migrate: %v", err)
	}
	log.Println("Migrations executed successfully!")
}

func initRedis() *redis.Client {
	redisURL := os.Getenv("REDIS_URL")
	var redisOptions *redis.Options

	if redisURL != "" {
		var err error
		redisOptions, err = redis.ParseURL(redisURL)
		if err != nil {
			log.Fatalf("Invalid REDIS_URL: %v", err)
		}
	} else {
		redisOptions = &redis.Options{Addr: "localhost:6379"}
	}

	client := redis.NewClient(redisOptions)
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		log.Fatalf("Impossible de se connecter à Redis: %v", err)
	}
	return client
}

func initAuth() (keyfunc.Keyfunc, string) {
	jwtSecret := os.Getenv("SUPABASE_JWT_SECRET")
	supabaseURL := os.Getenv("SUPABASE_URL")
	var jwks keyfunc.Keyfunc

	if supabaseURL != "" {
		jwksURL := supabaseURL + "/auth/v1/.well-known/jwks.json"
		var err error
		jwks, err = keyfunc.NewDefault([]string{jwksURL})
		if err != nil {
			log.Printf("Attention: Impossible d'initialiser JWKS: %v. Les tokens ES256 échoueront.", err)
		} else {
			log.Println("JWKS initialisé avec succès pour validation ES256")
		}
	}
	return jwks, jwtSecret
}

func setupRouter() *gin.Engine {
	router := gin.Default()
	config := cors.DefaultConfig()

	allowedOriginsEnv := os.Getenv("ALLOWED_ORIGINS")
	if allowedOriginsEnv != "" {
		config.AllowOrigins = strings.Split(allowedOriginsEnv, ",")
	} else {
		config.AllowOrigins = []string{"http://localhost:5173", "http://localhost:3000"}
	}

	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	config.ExposeHeaders = []string{"Content-Length"}
	config.AllowCredentials = true
	config.MaxAge = 12 * time.Hour
	router.Use(cors.New(config))

	router.Use(middleware.SecureHeaders())
	router.Use(middleware.RateLimitMiddleware())

	return router
}

func registerRoutes(router *gin.Engine, db *bun.DB, redisClient *redis.Client, jwks keyfunc.Keyfunc, jwtSecret string, friendHandler *friends.FriendHandler) {
	api := router.Group("/api/v1")

	// Authentication Routes (Public + Protected)
	authGroup := api.Group("/auth")
	authGroup.POST("/recovery/init", func(c *gin.Context) { auth.RecoveryInitHandler(c, db) })
	authGroup.POST("/recovery/finish", func(c *gin.Context) { auth.RecoveryFinishHandler(c, db) })

	// Public Share Routes
	publicShareRoutes := api.Group("/public/share")
	publicShareRoutes.GET("/:token", func(c *gin.Context) { shares.GetShareLinkHandler(c, db) })
	publicShareRoutes.GET("/:token/download", func(c *gin.Context) { shares.DownloadSharedFileHandler(c, db) })
	publicShareRoutes.GET("/:token/download/file/:file_id", func(c *gin.Context) { shares.DownloadFileFromSharedFolderHandler(c, db) })
	publicShareRoutes.GET("/:token/browse/*subpath", func(c *gin.Context) { shares.BrowseSharedFolderHandler(c, db) })

	// Protected Routes
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(jwks, jwtSecret))

	registerUserRoutes(protected, db, redisClient)
	registerFileRoutes(protected, db, redisClient)
	registerFolderRoutes(protected, db)
	registerTagRoutes(protected, db)
	registerFriendRoutes(protected, friendHandler)
	registerShareRoutes(protected, db)
	registerSecurityRoutes(protected)
	registerBillingRoutes(router, api, protected, db)
	registerP2PRoutes(protected, db)

	// System
	router.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	router.GET("/api/v1/ping", func(c *gin.Context) { c.JSON(200, gin.H{"message": "pong", "version": "3.0"}) })
}

func registerUserRoutes(g *gin.RouterGroup, db *bun.DB, redisClient *redis.Client) {
	g.POST("/auth/register", func(c *gin.Context) { auth.RegisterHandler(c, db) })
	g.GET("/auth/keys", func(c *gin.Context) { auth.GetUserKeys(c, db) })
	g.POST("/auth/logout", func(c *gin.Context) { auth.LogoutHandler(c, redisClient) })

	usersG := g.Group("/users")
	usersG.GET("/", func(c *gin.Context) { users.ListUsersHandler(c, db) })
	usersG.GET("/me", func(c *gin.Context) { users.MeHandler(c, db) })
	usersG.POST("/change-password", func(c *gin.Context) { users.UpdatePasswordHandler(c, db) })
	usersG.PUT("/profile", func(c *gin.Context) { users.UpdateProfileHandler(c, db) })
	usersG.PUT("/avatar", func(c *gin.Context) { users.UpdateAvatarHandler(c, db) })
	usersG.POST("/recent", func(c *gin.Context) { users.AddRecentActivityHandler(c, db) })
	usersG.GET("/recent", func(c *gin.Context) { users.GetRecentActivityHandler(c, db) })
	usersG.POST("/keys", func(c *gin.Context) { keys.UpdateKeysHandler(c, db) })
}

func registerFileRoutes(g *gin.RouterGroup, db *bun.DB, redisClient *redis.Client) {
	filesG := g.Group("/files")
	filesG.POST("/upload", func(c *gin.Context) { files.UploadHandler(c, db, redisClient) })
	filesG.GET("/list-recursive", func(c *gin.Context) { files.ListAllFilesRecursiveHandler(c, db) })
	filesG.GET("/list/*path", func(c *gin.Context) { files.ListFilesHandler(c, db) })
	filesG.POST("/bulk-delete", func(c *gin.Context) { files.BulkDeleteHandler(c, db) })
	filesG.DELETE("/file/:fileID", func(c *gin.Context) { files.DeleteFileHandler(c, db) })
	filesG.DELETE("/folder/:folderID", func(c *gin.Context) { files.DeleteFolderHandler(c, db) })
	filesG.POST("/move", func(c *gin.Context) { files.MoveHandler(c, db, redisClient) })
	filesG.POST("/rename", func(c *gin.Context) { files.RenameHandler(c, db, redisClient) })
	filesG.POST("/tags", func(c *gin.Context) { files.UpdateTagsHandler(c, db) })
	filesG.GET("/download/:fileID", func(c *gin.Context) { files.DownloadFileHandler(c, db) })
	filesG.GET("/preview/:fileID", func(c *gin.Context) { files.PreviewFileHandler(c, db) })
	filesG.GET("/search", func(c *gin.Context) { files.SearchFilesHandler(c, db) })

	// Direct-to-S3 Multipart Upload Routes
	filesG.POST("/multipart/initiate", func(c *gin.Context) { files.InitiateMultipartHandler(c, db) })
	filesG.POST("/multipart/complete", func(c *gin.Context) { files.CompleteMultipartHandler(c, db) })
	filesG.POST("/multipart/abort", func(c *gin.Context) { files.AbortMultipartHandler(c, db) })
	filesG.POST("/multipart/refresh-url", func(c *gin.Context) { files.RefreshPresignedURLsHandler(c, db) })
	filesG.GET("/download/:fileID/presigned", func(c *gin.Context) { files.GetPresignedDownloadHandler(c, db) })

	// Batch Presign Routes (for ZIP download)
	filesG.POST("/batch-presign", func(c *gin.Context) { files.BatchPresignDownloadHandler(c, db) })
	filesG.POST("/selection-tree", func(c *gin.Context) { files.GetSelectionTreeHandler(c, db) })
}

func registerFolderRoutes(g *gin.RouterGroup, db *bun.DB) {
	foldersG := g.Group("/folders")
	foldersG.POST("/create", func(c *gin.Context) { folders.CreateHandler(c, db) })
	foldersG.PUT("/:id/key", func(c *gin.Context) { folders.UpdateFolderKeyHandler(c, db) })
	foldersG.GET("/:id/tree", func(c *gin.Context) { folders.GetFolderTreeHandler(c, db) })
}

func registerTagRoutes(g *gin.RouterGroup, db *bun.DB) {
	tagsG := g.Group("/tags")
	tagsG.GET("/", tags.ListTagsHandler(db))
	tagsG.POST("/", tags.CreateTagHandler(db))
	tagsG.DELETE("/:id", tags.DeleteTagHandler(db))
}

func registerFriendRoutes(g *gin.RouterGroup, h *friends.FriendHandler) {
	friendsG := g.Group("/friends")
	friendsG.GET("", h.ListFriends)
	friendsG.POST("", h.AddFriend)
	friendsG.DELETE("/:id", h.RemoveFriend)
	friendsG.PUT("/:id/accept", h.AcceptFriend)
	friendsG.DELETE("/:id/reject", h.RejectFriend)
}

func registerShareRoutes(g *gin.RouterGroup, db *bun.DB) {
	sharesG := g.Group("/shares")
	sharesG.GET("/list", func(c *gin.Context) { shares.ListSharesHandler(c, db) })
	sharesG.POST("/link", func(c *gin.Context) { shares.CreateShareLinkHandler(c, db) })
	sharesG.POST("/direct", func(c *gin.Context) { shares.CreateDirectShareHandler(c, db) })
	sharesG.GET("/direct", func(c *gin.Context) { shares.ListDirectSharesForResourceHandler(c, db) })
	sharesG.DELETE("/direct", func(c *gin.Context) { shares.RemoveDirectShareHandler(c, db) })
	sharesG.GET("/check-path", func(c *gin.Context) { shares.GetActiveSharesForPathHandler(c, db) })
	sharesG.GET("/file/:fileID", func(c *gin.Context) { shares.GetShareForResourceHandler(c, db) })
	sharesG.GET("/direct/folder/:folderID/content", func(c *gin.Context) { shares.GetSharedFolderContentHandler(c, db) })
	sharesG.DELETE("/link/:shareID", func(c *gin.Context) { shares.DeleteShareLinkHandler(c, db) })
	sharesG.GET("/with-me", func(c *gin.Context) { shares.ListImportedSharesHandler(c, db) })
	sharesG.POST("/with-me", func(c *gin.Context) { shares.ImportShareHandler(c, db) })
	sharesG.DELETE("/with-me/:id", func(c *gin.Context) { shares.RemoveImportedShareHandler(c, db) })
}

func registerSecurityRoutes(g *gin.RouterGroup) {
	securityG := g.Group("/security")
	securityG.POST("/report", func(c *gin.Context) { security.ReportSecurityEvent(c) })
	securityG.GET("/events", func(c *gin.Context) { security.GetSecurityEvents(c) })
}

// P2P Signaling Routes (replaces WebSocket signaling)
func registerP2PRoutes(g *gin.RouterGroup, db *bun.DB) {
	p2pG := g.Group("/p2p")

	// Send a P2P signal to another user
	p2pG.POST("/signal", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		var req struct {
			TargetUserID string                 `json:"target_user_id" binding:"required"`
			SignalType   string                 `json:"signal_type" binding:"required"`
			Payload      map[string]interface{} `json:"payload" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		signal := &pkg.P2PSignal{
			SenderID:   userID.(string),
			TargetID:   req.TargetUserID,
			SignalType: req.SignalType,
			Payload:    req.Payload,
		}

		if _, err := db.NewInsert().Model(signal).Exec(c.Request.Context()); err != nil {
			c.JSON(500, gin.H{"error": "Failed to send signal"})
			return
		}

		c.JSON(200, gin.H{"status": "sent", "signal_id": signal.ID})
	})

	// Poll for pending P2P signals
	p2pG.GET("/signals", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		var signals []pkg.P2PSignal

		err := db.NewSelect().
			Model(&signals).
			Where("target_id = ? AND consumed = false", userID).
			Order("created_at ASC").
			Scan(c.Request.Context())

		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to fetch signals"})
			return
		}

		// Mark fetched signals as consumed
		if len(signals) > 0 {
			var ids []int64
			for _, s := range signals {
				ids = append(ids, s.ID)
			}
			db.NewUpdate().Model((*pkg.P2PSignal)(nil)).
				Set("consumed = true").
				Where("id IN (?)", bun.In(ids)).
				Exec(c.Request.Context())
		}

		c.JSON(200, gin.H{"signals": signals})
	})

	// ICE Configuration (TURN/STUN servers)
	p2pG.GET("/ice-config", func(c *gin.Context) {
		turnURL := os.Getenv("TURN_SERVER_URL")
		turnUser := os.Getenv("TURN_USERNAME")
		turnCred := os.Getenv("TURN_CREDENTIAL")

		iceServers := []map[string]interface{}{
			{"urls": []string{"stun:stun.l.google.com:19302"}},
		}

		if turnURL != "" {
			iceServers = append(iceServers, map[string]interface{}{
				"urls":       []string{turnURL},
				"username":   turnUser,
				"credential": turnCred,
			})
		}

		c.JSON(200, gin.H{"iceServers": iceServers})
	})
}

func registerBillingRoutes(router *gin.Engine, api *gin.RouterGroup, protected *gin.RouterGroup, db *bun.DB) {
	// Initialize payment provider (use mock in test mode, Mollie in production)
	var paymentProvider billing.PaymentProvider
	if os.Getenv("BILLING_TEST_MODE") == "true" {
		log.Println("[Billing] Using MockPaymentProvider (test mode)")
		paymentProvider = billing.NewMockPaymentProvider()
	} else {
		mollieProvider, err := billing.NewMolliePaymentProviderFromEnv()
		if err != nil {
			log.Printf("[Billing] Warning: Mollie not configured: %v", err)
			paymentProvider = billing.NewMockPaymentProvider()
		} else {
			log.Println("[Billing] Using MolliePaymentProvider")
			paymentProvider = mollieProvider
		}
	}

	// Initialize usage collector
	usageCollector := billing.NewUsageCollectorFromEnv()
	usageCollector.Start()

	// Initialize services
	webhookHandler := billing.NewLagoWebhookHandlerFromEnv(db, paymentProvider)
	billingService := billing.NewBillingServiceFromEnv(db, paymentProvider, usageCollector)

	// Public webhook endpoint (Lago calls this)
	api.POST("/webhooks/lago", webhookHandler.HandleWebhook)

	// Protected billing endpoints
	billingG := protected.Group("/billing")
	billingG.GET("/plan", billingService.GetCurrentPlan)
	billingG.GET("/invoices", billingService.GetInvoices)
	billingG.GET("/invoices/:invoiceID/payment-link", billingService.GetPaymentLink)
	billingG.GET("/pending-invoices", webhookHandler.GetPendingInvoices)
}

func startServer(router *gin.Engine) {
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
