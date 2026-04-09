package main

import (
	"context"
	"encoding/json"
	"kagibi/backend/handlers/auth"
	billinghandlers "kagibi/backend/handlers/billing"
	"kagibi/backend/handlers/files"
	"kagibi/backend/handlers/folders"
	"kagibi/backend/handlers/friends"
	"kagibi/backend/handlers/keys"
	"kagibi/backend/handlers/security"
	"kagibi/backend/handlers/shares"
	"kagibi/backend/handlers/tags"
	"kagibi/backend/handlers/users"
	wshandler "kagibi/backend/handlers/ws"
	"kagibi/backend/middleware"
	"kagibi/backend/pkg"
	"kagibi/backend/pkg/authprovider"
	"kagibi/backend/pkg/monitoring"
	"kagibi/backend/pkg/s3storage"
	"kagibi/backend/pkg/workers"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/uptrace/bun"
)

func main() {
	loadEnv()

	// DB must be initialized before auth so LocalProvider can access auth_users
	db := pkg.NewDB()
	migrateDB(db)

	// Register the WebSocket hub so EmitRealtimeEvent can push live events
	pkg.SetWSHub(wshandler.GlobalHub)

	provider := initAuth(db)
	log.Printf("Starting Kagibi Backend v3.0 (auth provider: %s)...", provider.Name())

	initS3()
	setupBillingProvider()

	redisClient := initRedis()

	// Start Workers
	workers.StartWorker(redisClient)
	workers.StartCleanupWorker(db)
	workers.StartAccountCleanupWorker(db) // RGPD Article 17

	friendHandler := friends.NewFriendHandler(db, wshandler.GlobalHub.IsConnected)
	setupPresenceHooks(db)

	metricsServer := monitoring.NewServer(9090)
	monitoring.StartSessionMonitor(redisClient)
	if err := metricsServer.Start(); err != nil {
		log.Printf("Warning: Failed to start metrics server: %v", err)
	}

	router := setupRouter(redisClient)
	registerRoutes(router, db, redisClient, provider, friendHandler)
	startServerWithGracefulShutdown(router, metricsServer)
}

func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	} else {
		log.Println("✓ .env file loaded successfully")
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
	if err := pkg.Migrate(db); err != nil {
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

// initAuth creates the auth provider based on the AUTH_PROVIDER environment variable.
// Supported values: "local" (default), "supabase", "pocketbase"
func initAuth(db *bun.DB) authprovider.AuthProvider {
	switch os.Getenv("AUTH_PROVIDER") {
	case "supabase":
		return authprovider.NewSupabaseProvider()
	case "pocketbase":
		p := authprovider.NewPocketBaseProvider()
		if err := p.SetupJWTSecret(); err != nil {
			log.Fatalf("Fatal: PocketBase JWT secret configuration failed — backend and PocketBase would use different signing keys, causing all token validations to fail: %v", err)
		}
		return p
	default: // "local" or unset
		return authprovider.NewLocalProvider(db)
	}
}

func setupRouter(redisClient *redis.Client) *gin.Engine {
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
	router.Use(middleware.RateLimitMiddleware(redisClient))
	router.Use(middleware.MetricsMiddleware())

	return router
}

func registerRoutes(router *gin.Engine, db *bun.DB, redisClient *redis.Client, provider authprovider.AuthProvider, friendHandler *friends.FriendHandler) {
	api := router.Group("/api/v1")

	// Public auth routes (no JWT required)
	authGroup := api.Group("/auth")
	authGroup.POST("/login", auth.LocalLoginHandler(provider))
	authGroup.POST("/signup", auth.LocalSignupHandler(provider))
	authGroup.POST("/refresh", auth.LocalRefreshHandler(provider, redisClient))
	authGroup.POST("/recovery/init", func(c *gin.Context) { auth.RecoveryInitHandler(c, db) })
	authGroup.POST("/recovery/finish", func(c *gin.Context) { auth.RecoveryFinishHandler(c, db, provider, redisClient) })

	// Public share routes
	publicShareRoutes := api.Group("/public/share")
	publicShareRoutes.GET("/:token", func(c *gin.Context) { shares.GetShareLinkHandler(c, db) })
	publicShareRoutes.GET("/:token/download", func(c *gin.Context) { shares.DownloadSharedFileHandler(c, db) })
	publicShareRoutes.GET("/:token/download/file/:file_id", func(c *gin.Context) { shares.DownloadFileFromSharedFolderHandler(c, db) })
	publicShareRoutes.GET("/:token/browse/*subpath", func(c *gin.Context) { shares.BrowseSharedFolderHandler(c, db) })

	// MFA routes — protected by JWT, registered on the auth group
	mfaGroup := authGroup.Group("/mfa")
	mfaGroup.Use(middleware.AuthMiddleware(provider, redisClient))
	mfaGroup.GET("/factors", auth.MFAListFactorsHandler(provider))
	mfaGroup.POST("/enroll", auth.MFAEnrollHandler(provider))
	mfaGroup.POST("/challenge", auth.MFAChallengeHandler(provider))
	mfaGroup.POST("/verify", auth.MFAVerifyHandler(provider))
	mfaGroup.DELETE("/unenroll", auth.MFAUnenrollHandler(provider, redisClient))

	// Protected routes (JWT required)
	authMW := middleware.AuthMiddleware(provider, redisClient)
	protected := api.Group("")
	protected.Use(authMW)

	registerUserRoutes(protected, db, redisClient, provider)
	registerFileRoutes(protected, db, redisClient)
	registerFolderRoutes(protected, db)
	registerTagRoutes(protected, db)
	registerFriendRoutes(protected, friendHandler)
	registerShareRoutes(protected, db)
	registerSecurityRoutes(protected)
	registerBillingRoutes(api, protected, authMW, db)
	registerP2PRoutes(protected, db)
	registerEventRoutes(protected, db)

	// WebSocket endpoint — authenticated via Authorization header or Sec-WebSocket-Protocol trick.
	wsAllowedOrigins := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")
	if len(wsAllowedOrigins) == 0 || wsAllowedOrigins[0] == "" {
		wsAllowedOrigins = []string{"http://localhost:5173", "http://localhost:3000"}
	}
	api.GET("/ws", wshandler.WebSocketHandler(provider, redisClient, wsAllowedOrigins))

	router.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	router.GET("/api/v1/ping", func(c *gin.Context) { c.JSON(200, gin.H{"message": "pong", "version": "3.0"}) })

	protected.GET("/heartbeat", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "alive", "timestamp": time.Now().Unix()})
	})
}

func registerUserRoutes(g *gin.RouterGroup, db *bun.DB, redisClient *redis.Client, provider authprovider.AuthProvider) {
	g.POST("/auth/register", func(c *gin.Context) { auth.RegisterHandler(c, db, provider) })
	g.GET("/auth/keys", func(c *gin.Context) { auth.GetUserKeys(c, db) })
	g.POST("/auth/logout", func(c *gin.Context) { auth.LogoutHandler(c, redisClient) })
	g.POST("/auth/ws-token", auth.WsTokenHandler(redisClient))
	g.POST("/auth/update-password", auth.LocalUpdatePasswordHandler(provider, redisClient))
	g.DELETE("/auth/account", auth.DeleteAccount(db, provider))

	usersG := g.Group("/users")
	usersG.GET("/", func(c *gin.Context) { users.ListUsersHandler(c, db) })
	usersG.GET("/me", func(c *gin.Context) { users.MeHandler(c, db) })
	usersG.POST("/change-password", func(c *gin.Context) { users.UpdatePasswordHandler(c, db) })
	usersG.PUT("/profile", func(c *gin.Context) { users.UpdateProfileHandler(c, db) })
	usersG.PUT("/avatar", func(c *gin.Context) { users.UpdateAvatarHandler(c, db) })
	usersG.POST("/recent", func(c *gin.Context) { users.AddRecentActivityHandler(c, db) })
	usersG.GET("/recent", func(c *gin.Context) { users.GetRecentActivityHandler(c, db) })
	usersG.POST("/keys", func(c *gin.Context) { keys.UpdateKeysHandler(c, db) })
	usersG.GET("/export", func(c *gin.Context) { users.ExportUserDataHandler(c, db) })
	usersG.GET("/security-settings", func(c *gin.Context) { users.GetSecuritySettingsHandler(c, db) })
	usersG.PUT("/security-settings", func(c *gin.Context) { users.UpdateSecuritySettingsHandler(c, db) })
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
	filesG.POST("/multipart/initiate", func(c *gin.Context) { files.InitiateMultipartHandler(c, db) })
	filesG.POST("/multipart/complete", func(c *gin.Context) { files.CompleteMultipartHandler(c, db) })
	filesG.POST("/multipart/abort", func(c *gin.Context) { files.AbortMultipartHandler(c, db) })
	filesG.POST("/multipart/refresh-url", func(c *gin.Context) { files.RefreshPresignedURLsHandler(c, db) })
	filesG.GET("/download/:fileID/presigned", func(c *gin.Context) { files.GetPresignedDownloadHandler(c, db) })
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

func registerP2PRoutes(g *gin.RouterGroup, db *bun.DB) {
	p2pG := g.Group("/p2p")

	p2pG.POST("/signal", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		var req struct {
			TargetUserID string         `json:"target_user_id" binding:"required"`
			SignalType   string         `json:"signal_type" binding:"required"`
			Payload      map[string]any `json:"payload" binding:"required"`
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

		// Push signal over WebSocket for immediate delivery
		wshandler.GlobalHub.SendP2PSignalToUser(req.TargetUserID, userID.(string), req.SignalType, req.Payload)

		c.JSON(200, gin.H{"status": "sent", "signal_id": signal.ID})
	})

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

	p2pG.GET("ice-config", func(c *gin.Context) {
		turnURL := os.Getenv("TURN_SERVER_URL")
		turnUser := os.Getenv("TURN_USERNAME")
		turnCred := os.Getenv("TURN_CREDENTIAL")

		iceServers := []map[string]any{
			{"urls": []string{"stun:stun.l.google.com:19302", "stun:stun1.l.google.com:19302"}},
		}

		if turnURL != "" {
			turnURLs := []string{turnURL}
			if !strings.Contains(turnURL, "?transport=") {
				turnURLs = append(turnURLs, turnURL+"?transport=tcp")
			}
			iceServers = append(iceServers, map[string]any{
				"urls":       turnURLs,
				"username":   turnUser,
				"credential": turnCred,
			})
		}
		c.JSON(200, gin.H{"iceServers": iceServers})
	})
}

// registerEventRoutes adds a polling endpoint for realtime events.
// Used by the frontend when Supabase Realtime is not available (PocketBase mode).
func registerEventRoutes(g *gin.RouterGroup, db *bun.DB) {
	g.GET("/events/poll", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		sinceID := c.DefaultQuery("since_id", "0")

		var events []pkg.RealtimeEvent
		err := db.NewSelect().
			Model(&events).
			Where("user_id = ? AND id > ?", userID, sinceID).
			Order("id ASC").
			Limit(50).
			Scan(c.Request.Context())

		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to fetch events"})
			return
		}
		c.JSON(200, gin.H{"events": events})
	})
}

func registerBillingRoutes(api *gin.RouterGroup, protected *gin.RouterGroup, authMW gin.HandlerFunc, db *bun.DB) {
	billinghandlers.RegisterWebhookRoute(api, db)
	billinghandlers.RegisterRoutes(protected, authMW, db)
}

// setupPresenceHooks wires WebSocket connect/disconnect events to broadcast
// online/offline presence updates to each connected friend.
func setupPresenceHooks(db *bun.DB) {
	wshandler.ConnectHook = func(userID string) {
		broadcastPresence(db, userID, true)
	}
	wshandler.DisconnectHook = func(userID string) {
		broadcastPresence(db, userID, false)
	}
}

func broadcastPresence(db *bun.DB, userID string, online bool) {
	var friendships []pkg.Friendship
	if err := db.NewSelect().
		Model(&friendships).
		Where("(user_id_1 = ? OR user_id_2 = ?) AND status = 'accepted'", userID, userID).
		Scan(context.Background()); err != nil {
		log.Printf("[Presence] Failed to fetch friends for user=%s: %v", userID, err)
		return
	}

	// Message announcing userID's status to all friends
	selfMsg, _ := json.Marshal(map[string]any{
		"type":    "presence_update",
		"user_id": userID,
		"online":  online,
	})

	for _, f := range friendships {
		friendID := f.UserID2
		if f.UserID2 == userID {
			friendID = f.UserID1
		}

		// Tell each friend about userID's new status
		wshandler.GlobalHub.SendToUser(friendID, selfMsg)

		// Bootstrap: when userID just came online, also tell userID about each friend's
		// current status — they may have connected before userID and would never
		// otherwise receive a presence update.
		if online {
			friendOnline := wshandler.GlobalHub.IsConnected(friendID)
			bootstrapMsg, _ := json.Marshal(map[string]any{
				"type":    "presence_update",
				"user_id": friendID,
				"online":  friendOnline,
			})
			wshandler.GlobalHub.SendToUser(userID, bootstrapMsg)
		}
	}
}

func startServerWithGracefulShutdown(router *gin.Engine, metricsServer *monitoring.Server) {
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Println("Serveur principal démarré sur le port 8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to run server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("\nSignal d'arrêt reçu, arrêt gracieux en cours...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := metricsServer.Shutdown(ctx); err != nil {
		log.Printf("Erreur lors de l'arrêt du serveur de métriques: %v", err)
	}
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Erreur lors de l'arrêt du serveur principal: %v", err)
	} else {
		log.Println("✓ Serveur arrêté gracieusement")
	}
}
