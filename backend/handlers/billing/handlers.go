package billing

import (
	"net/http"
	"os"

	billingpkg "safercloud/backend/pkg/billing"

	"github.com/gin-gonic/gin"
)

// GetBillingStatusHandler retourne si le billing est activé
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
		// Détecter le type de provider
		switch provider.(type) {
		case *billingpkg.DisabledProvider:
			providerType = "disabled"
		case *billingpkg.MockProvider:
			providerType = "mock"
		case *billingpkg.WebhookProvider:
			providerType = "webhook"
		default:
			providerType = "unknown"
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"enabled":       enabled,
		"provider_type": providerType,
		"features": gin.H{
			"subscriptions": enabled && providerType != "disabled",
			"quotas":        enabled && providerType != "disabled",
			"invoices":      enabled && providerType == "webhook",
		},
	})
}

// GetCurrentPlanHandler retourne le plan actuel de l'utilisateur
// GET /api/billing/plan
func GetCurrentPlanHandler(c *gin.Context) {
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

	plan, err := provider.GetUserPlan(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get plan"})
		return
	}

	c.JSON(http.StatusOK, plan)
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

// GetUsageHandler retourne l'utilisation actuelle de l'utilisateur
// GET /api/billing/usage
func GetUsageHandler(c *gin.Context) {
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

	usage, err := provider.GetCurrentUsage(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get usage"})
		return
	}

	c.JSON(http.StatusOK, usage)
}

// CheckQuotaHandler vérifie si une opération est autorisée
// POST /api/billing/quota/check
func CheckQuotaHandler(c *gin.Context) {
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

	provider := billingpkg.GetProvider()
	if provider == nil {
		// Pas de provider = autorisé par défaut
		c.JSON(http.StatusOK, gin.H{
			"allowed": true,
		})
		return
	}

	result, err := provider.CheckQuota(c.Request.Context(), userID, req.RequestedBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check quota"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetInvoicesHandler retourne les factures de l'utilisateur
// GET /api/billing/invoices
func GetInvoicesHandler(c *gin.Context) {
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

	invoices, err := provider.GetInvoices(c.Request.Context(), userID, 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get invoices"})
		return
	}

	c.JSON(http.StatusOK, invoices)
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

// UpgradePlanHandler initie un changement de plan
// POST /api/billing/upgrade
func UpgradePlanHandler(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		PlanCode       string `json:"plan_code" binding:"required"`
		IdempotencyKey string `json:"idempotency_key" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Plan code and idempotency key required"})
		return
	}

	provider := billingpkg.GetProvider()
	if provider == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Billing service unavailable"})
		return
	}

	subscription, err := provider.UpdateSubscription(c.Request.Context(), userID, req.PlanCode, req.IdempotencyKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade plan"})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

// RegisterRoutes enregistre toutes les routes billing
func RegisterRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	billing := router.Group("/billing")
	{
		// Routes publiques
		billing.GET("/status", GetBillingStatusHandler)
		billing.GET("/plans", GetPlansHandler)

		// Routes authentifiées
		authenticated := billing.Group("")
		authenticated.Use(authMiddleware)
		{
			authenticated.GET("/plan", GetCurrentPlanHandler)
			authenticated.GET("/subscription", GetSubscriptionHandler)
			authenticated.GET("/usage", GetUsageHandler)
			authenticated.POST("/quota/check", CheckQuotaHandler)
			authenticated.GET("/invoices", GetInvoicesHandler)
			authenticated.GET("/invoices/:id/payment-link", GetPaymentLinkHandler)
			authenticated.POST("/upgrade", UpgradePlanHandler)
		}
	}
}
