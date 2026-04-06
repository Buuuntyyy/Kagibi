//go:build !production

package main

import (
	"log"
	"os"

	billingpkg "kagibi/backend/pkg/billing"
)

// setupBillingProvider initialise le provider de facturation pour le build OSS.
// Le build production (repo privé) remplace ce fichier pour activer Stripe.
//
//	BILLING_ENABLED=false  → DisabledProvider (self-hosted, sans limites)
//	(défaut)               → MockProvider     (dev, limites simulées)
func setupBillingProvider() {
	if os.Getenv("BILLING_ENABLED") == "false" {
		billingpkg.SetProvider(billingpkg.NewDisabledProvider())
		log.Println("[Billing] DISABLED — Self-hosted mode (unlimited storage)")
		return
	}
	billingpkg.SetProvider(billingpkg.NewMockProvider())
	log.Println("[Billing] MockProvider initialized — dev mode (simulated limits)")
}
