package billing

import (
	"fmt"
	"net/http"
	"os"

	"safercloud/backend/pkg"
	billingpkg "safercloud/backend/pkg/billing"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// GetBillingStatusHandler retourne le statut du billing
// GET /api/billing/status
func GetBillingStatusHandler(c *gin.Context) {
	enabled := os.Getenv("BILLING_ENABLED") != "false"
	provider := billingpkg.GetProvider()
	var providerType string
	if !enabled {
		providerType = "disabled"
	} else if provider == nil {
		providerType = "none"
	} else {
		// Use optional ProviderName() interface (avoids build-tag coupling)
		type providerNamer interface{ ProviderName() string }
		if n, ok := provider.(providerNamer); ok {
			providerType = n.ProviderName()
		} else {
			providerType = "unknown"
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"enabled":       enabled,
		"provider_type": providerType,
		"features": gin.H{
			"subscriptions": enabled && providerType != "disabled",
			"quotas":        enabled && providerType != "disabled",
			"invoices":      enabled && providerType == "stripe",
			"checkout":      enabled && providerType != "disabled",
			"portal":        enabled && providerType == "stripe",
		},
	})
}

// GetCurrentPlanHandler retourne le plan actuel de l'utilisateur depuis la DB
// GET /api/billing/plan
func GetCurrentPlanHandler(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		provider := billingpkg.GetProvider()
		if provider == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Billing service unavailable"})
			return
		}
		var user pkg.User
		if err := db.NewSelect().Model(&user).Where("id = ?", userID).Scan(c.Request.Context()); err != nil {
			user.Plan = "free"
		}
		plan, err := provider.GetPlan(c.Request.Context(), user.Plan)
		if err != nil {
			plan, _ = provider.GetPlan(c.Request.Context(), "free")
		}
		c.JSON(http.StatusOK, plan)
	}
}

// GetSubscriptionHandler retourne la souscription de l'utilisateur
// GET /api/billing/subscription
func GetSubscriptionHandler(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	provider := billingpkg.GetProvider()
	if provider == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Billing service unavailable"})
		return
	}
	subscription, err := provider.GetSubscription(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscription"})
		return
	}
	c.JSON(http.StatusOK, subscription)
}

// GetPlansHandler retourne la liste des plans disponibles
// GET /api/billing/plans
func GetPlansHandler(c *gin.Context) {
	provider := billingpkg.GetProvider()
	if provider == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Billing service unavailable"})
		return
	}
	plans, err := provider.ListPlans(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list plans"})
		return
	}
	c.JSON(http.StatusOK, plans)
}

// GetUsageHandler retourne l'utilisation actuelle depuis la DB
// GET /api/billing/usage
func GetUsageHandler(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		var user pkg.User
		if err := db.NewSelect().Model(&user).Where("id = ?", userID).Scan(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
			return
		}
		activeShares, _ := db.NewSelect().TableExpr("file_shares fs").
			Join("JOIN files f ON f.id = fs.file_id").
			Where("f.user_id = ?", userID).
			Count(c.Request.Context())

		plan, _ := billingpkg.GetProvider().GetPlan(c.Request.Context(), user.Plan)
		c.JSON(http.StatusOK, gin.H{
			"storage_used_bytes": user.StorageUsed,
			"storage_used_gb":    float64(user.StorageUsed) / (1024 * 1024 * 1024),
			"storage_limit_gb":   plan.StorageLimitGB,
			"p2p_shares_active":  activeShares,
			"p2p_shares_limit":   plan.P2PSharesLimit,
		})
	}
}

// CheckQuotaHandler vérifie si un upload est autorisé par le quota stockage
// POST /api/billing/quota/check
func CheckQuotaHandler(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		var req struct {
			RequestedBytes int64 `json:"requested_bytes"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		var user pkg.User
		if err := db.NewSelect().Model(&user).Where("id = ?", userID).Scan(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
			return
		}
		remaining := user.StorageLimit - user.StorageUsed
		result := billingpkg.QuotaCheckResult{
			CurrentUsage:   user.StorageUsed,
			Limit:          user.StorageLimit,
			RemainingBytes: remaining,
		}
		if req.RequestedBytes > remaining {
			result.Allowed = false
			result.Reason = fmt.Sprintf("Quota de stockage dépassé. Restant : %.2f Go", float64(remaining)/(1024*1024*1024))
		} else {
			result.Allowed = true
		}
		c.JSON(http.StatusOK, result)
	}
}

// CheckP2PQuotaHandler vérifie si un nouveau partage P2P peut être créé
// POST /api/billing/quota/p2p
func CheckP2PQuotaHandler(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		var user pkg.User
		if err := db.NewSelect().Model(&user).Where("id = ?", userID).Scan(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
			return
		}
		activeShares, _ := db.NewSelect().TableExpr("file_shares fs").
			Join("JOIN files f ON f.id = fs.file_id").
			Where("f.user_id = ?", userID).
			Count(c.Request.Context())

		p2pLimit := pkg.GetP2PLimit(user.Plan)
		remaining := p2pLimit - activeShares
		result := billingpkg.P2PQuotaCheckResult{
			ActiveShares:    activeShares,
			Limit:           p2pLimit,
			RemainingShares: remaining,
		}
		if remaining <= 0 {
			result.Allowed = false
			result.Reason = fmt.Sprintf("Limite de %d partages P2P actifs atteinte. Passez au plan supérieur.", p2pLimit)
		} else {
			result.Allowed = true
		}
		c.JSON(http.StatusOK, result)
	}
}

// GetInvoicesHandler retourne les factures depuis Stripe
// GET /api/billing/invoices
func GetInvoicesHandler(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		provider := billingpkg.GetProvider()
		if provider == nil {
			c.JSON(http.StatusOK, []interface{}{})
			return
		}
		var user pkg.User
		if err := db.NewSelect().Model(&user).Where("id = ?", userID).Scan(c.Request.Context()); err != nil {
			c.JSON(http.StatusOK, []interface{}{})
			return
		}
		if user.StripeCustomerID == "" {
			c.JSON(http.StatusOK, []interface{}{})
			return
		}
		invoices, err := provider.GetInvoices(c.Request.Context(), user.StripeCustomerID, 20)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get invoices"})
			return
		}
		c.JSON(http.StatusOK, invoices)
	}
}

// GetPaymentLinkHandler génère un lien de paiement pour une facture
// GET /api/billing/invoices/:id/payment-link
func GetPaymentLinkHandler(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	invoiceID := c.Param("id")
	if invoiceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invoice ID required"})
		return
	}
	provider := billingpkg.GetProvider()
	if provider == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Billing service unavailable"})
		return
	}
	link, err := provider.GetPaymentLink(c.Request.Context(), invoiceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get payment link"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": link})
}

// CreateCheckoutHandler crée une session Stripe Checkout pour upgrade de plan
// POST /api/billing/checkout
// Body: { "plan_code": "pro"|"business", "interval": "monthly"|"yearly" }
func CreateCheckoutHandler(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	var req struct {
		PlanCode string `json:"plan_code" binding:"required"`
		Interval string `json:"interval"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "plan_code is required"})
		return
	}
	if req.Interval != "yearly" {
		req.Interval = "monthly"
	}
	provider := billingpkg.GetProvider()
	if provider == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Billing service unavailable"})
		return
	}
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "https://kagibi.stratus.ovh"
	}
	successURL := fmt.Sprintf("%s/dashboard/billing?checkout=success&session_id={CHECKOUT_SESSION_ID}", frontendURL)
	cancelURL := fmt.Sprintf("%s/dashboard/billing?checkout=cancelled", frontendURL)

	checkoutURL, err := provider.CreateCheckoutSession(c.Request.Context(), userID, req.PlanCode, req.Interval, successURL, cancelURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"checkout_url": checkoutURL})
}

// CreatePortalHandler crée une session Stripe Customer Portal
// POST /api/billing/portal
func CreatePortalHandler(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		provider := billingpkg.GetProvider()
		if provider == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Billing service unavailable"})
			return
		}
		var user pkg.User
		if err := db.NewSelect().Model(&user).Where("id = ?", userID).Scan(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
			return
		}
		if user.StripeCustomerID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Aucun abonnement actif. Souscrivez à un plan d'abord."})
			return
		}
		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "https://kagibi.stratus.ovh"
		}
		returnURL := fmt.Sprintf("%s/dashboard/billing", frontendURL)
		portalURL, err := provider.CreatePortalSession(c.Request.Context(), user.StripeCustomerID, returnURL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"portal_url": portalURL})
	}
}

// RegisterRoutes enregistre toutes les routes billing
func RegisterRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc, db *bun.DB) {
	billing := router.Group("/billing")
	{
		billing.GET("/status", GetBillingStatusHandler)
		billing.GET("/plans", GetPlansHandler)

		authenticated := billing.Group("")
		authenticated.Use(authMiddleware)
		{
			authenticated.GET("/plan", GetCurrentPlanHandler(db))
			authenticated.GET("/subscription", GetSubscriptionHandler)
			authenticated.GET("/usage", GetUsageHandler(db))
			authenticated.POST("/quota/check", CheckQuotaHandler(db))
			authenticated.POST("/quota/p2p", CheckP2PQuotaHandler(db))
			authenticated.GET("/invoices", GetInvoicesHandler(db))
			authenticated.GET("/invoices/:id/payment-link", GetPaymentLinkHandler)
			authenticated.POST("/checkout", CreateCheckoutHandler)
			authenticated.POST("/portal", CreatePortalHandler(db))
		}
	}
}
