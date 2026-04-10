// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package billing

import (
	"fmt"
	"net/http"
	"os"

	"kagibi/backend/pkg"
	billingpkg "kagibi/backend/pkg/billing"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

const errFailedGetUserPlan = "Failed to get user plan"

func getUserPlanState(db *bun.DB, userID string) (*pkg.UserPlan, error) {
	planState, err := pkg.FindUserPlanByUserID(db, userID)
	if err == nil && planState != nil {
		return planState, nil
	}

	// Fallback self-heal: create a free default row if missing
	planState = &pkg.UserPlan{
		UserID:           userID,
		Plan:             pkg.PlanFree,
		StorageLimit:     pkg.StorageFree,
		StorageUsed:      0,
		P2PMaxExchanges:  pkg.P2PLimitFree,
		P2PExchangesUsed: 0,
	}
	if upsertErr := pkg.UpsertUserPlan(db, planState); upsertErr != nil {
		return nil, upsertErr
	}
	return planState, nil
}

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
		planState, err := getUserPlanState(db, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errFailedGetUserPlan})
			return
		}
		// Return plan info directly from DB — no billing provider required
		provider := billingpkg.GetProvider()
		if provider != nil {
			if plan, err := provider.GetPlan(c.Request.Context(), planState.Plan); err == nil {
				c.JSON(http.StatusOK, plan)
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"code":             planState.Plan,
			"name":             planState.Plan,
			"storage_limit_gb": float64(planState.StorageLimit) / (1024 * 1024 * 1024),
		})
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
		// Billing not configured — return no active subscription
		c.JSON(http.StatusOK, nil)
		return
	}
	subscription, err := provider.GetSubscription(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusOK, nil)
		return
	}
	c.JSON(http.StatusOK, subscription)
}

// GetPlansHandler retourne la liste des plans disponibles
// GET /api/billing/plans
func GetPlansHandler(c *gin.Context) {
	provider := billingpkg.GetProvider()
	if provider == nil {
		// Billing not configured — return empty list
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	plans, err := provider.ListPlans(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, []interface{}{})
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
		planState, err := getUserPlanState(db, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errFailedGetUserPlan})
			return
		}
		activeShares, _ := db.NewSelect().TableExpr("file_shares fs").
			Join("JOIN files f ON f.id = fs.file_id").
			Where("f.user_id = ?", userID).
			Count(c.Request.Context())
		_, _ = db.NewUpdate().Model((*pkg.UserPlan)(nil)).
			Set("p2p_exchanges_used = ?", activeShares).
			Where("user_id = ?", userID).
			Exec(c.Request.Context())
		c.JSON(http.StatusOK, gin.H{
			"storage_used_bytes": planState.StorageUsed,
			"storage_used_gb":    float64(planState.StorageUsed) / (1024 * 1024 * 1024),
			"storage_limit_gb":   float64(planState.StorageLimit) / (1024 * 1024 * 1024),
			"p2p_shares_active":  activeShares,
			"p2p_shares_limit":   planState.P2PMaxExchanges,
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
		planState, err := getUserPlanState(db, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errFailedGetUserPlan})
			return
		}
		remaining := planState.StorageLimit - planState.StorageUsed
		result := billingpkg.QuotaCheckResult{
			CurrentUsage:   planState.StorageUsed,
			Limit:          planState.StorageLimit,
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
		activeShares, _ := db.NewSelect().TableExpr("file_shares fs").
			Join("JOIN files f ON f.id = fs.file_id").
			Where("f.user_id = ?", userID).
			Count(c.Request.Context())

		// Billing disabled: P2P quota is not enforced
		result := billingpkg.P2PQuotaCheckResult{
			ActiveShares:    activeShares,
			Limit:           -1, // unlimited
			RemainingShares: -1,
			Allowed:         true,
		}
		c.JSON(http.StatusOK, result)
	}
}

// GetInvoicesHandler retourne les factures depuis Stripe
// GET /api/billing/invoices
func GetInvoicesHandler(db *bun.DB) gin.HandlerFunc {
	_ = db
	return func(c *gin.Context) {
		// TODO: implement when Stripe is enabled
		c.JSON(http.StatusOK, []interface{}{})
	}
}

// GetPaymentLinkHandler génère un lien de paiement pour une facture
// GET /api/billing/invoices/:id/payment-link
func GetPaymentLinkHandler(c *gin.Context) {
	provider := billingpkg.GetProvider()
	if provider == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Billing non configuré"})
		return
	}
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
	provider := billingpkg.GetProvider()
	if provider == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Billing non configuré"})
		return
	}
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
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "https://kagibi.cloud"
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
	_ = db
	return func(c *gin.Context) {
		// TODO: implement when Stripe is enabled
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Billing non activé"})
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
