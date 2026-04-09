//go:build !production

package billing

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// HandleStripeWebhook is a no-op stub in the open-source build.
// The real Stripe implementation lives in the private repository
// and is compiled with -tags production.
func HandleStripeWebhook(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{
			"error": "Stripe webhook not available in this build.",
		})
	}
}

// RegisterWebhookRoute registers the webhook stub (always 501 in OSS builds).
func RegisterWebhookRoute(router *gin.RouterGroup, db *bun.DB) {
	router.POST("/billing/webhook", HandleStripeWebhook(db))
}

// --- stub types/functions below so the !production build compiles ---

// activatePlan is implemented in the production build only.
// This stub exists so handlers.go can reference types without errors.

// Ensure bun import is used.
var _ = (*bun.DB)(nil)
