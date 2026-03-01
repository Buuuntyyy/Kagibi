package billing

import (
	"context"
	"fmt"
	"log"
)

// DisabledProvider est utilisé quand le billing est complètement désactivé
// Mode self-hosted : pas de limites, pas de facturation
type DisabledProvider struct{}

// NewDisabledProvider crée un provider désactivé (self-hosted mode)
func NewDisabledProvider() *DisabledProvider {
	return &DisabledProvider{}
}

// ProviderName implements the optional providerNamer interface.
func (d *DisabledProvider) ProviderName() string { return "disabled" }

// === Lifecycle Events ===

func (d *DisabledProvider) OnUserCreated(ctx context.Context, event UserCreatedEvent) error {
	log.Printf("[Billing] Disabled - User created: %s", event.UserID)
	return nil
}

func (d *DisabledProvider) OnUserDeleted(ctx context.Context, userID string) error {
	log.Printf("[Billing] Disabled - User deleted: %s", userID)
	return nil
}

// === Subscription Management ===

func (d *DisabledProvider) CreateSubscription(ctx context.Context, userID, planCode, idempotencyKey string) (*Subscription, error) {
	return d.GetSubscription(ctx, userID)
}

func (d *DisabledProvider) GetSubscription(ctx context.Context, userID string) (*Subscription, error) {
	// Retourne un abonnement illimité fictif
	return &Subscription{
		ID:       "disabled",
		UserID:   userID,
		PlanCode: "unlimited",
		Status:   "active",
	}, nil
}

func (d *DisabledProvider) UpdateSubscription(ctx context.Context, userID, newPlanCode, idempotencyKey string) (*Subscription, error) {
	return d.GetSubscription(ctx, userID)
}

func (d *DisabledProvider) CancelSubscription(ctx context.Context, userID, idempotencyKey string) error {
	return nil
}

// === Plan Information ===

func (d *DisabledProvider) GetPlan(ctx context.Context, planCode string) (*Plan, error) {
	return d.getUnlimitedPlan(), nil
}

func (d *DisabledProvider) GetUserPlan(ctx context.Context, userID string) (*Plan, error) {
	return d.getUnlimitedPlan(), nil
}

func (d *DisabledProvider) ListPlans(ctx context.Context) ([]Plan, error) {
	return []Plan{*d.getUnlimitedPlan()}, nil
}

func (d *DisabledProvider) getUnlimitedPlan() *Plan {
	return &Plan{
		Code:           "unlimited",
		Name:           "Self-Hosted",
		Description:    "Stockage illimité (mode auto-hébergé)",
		StorageLimitGB: 999999999,
		P2PSharesLimit: 999999999,
		PriceMonthly:   0,
		Currency:       "EUR",
		Features: map[string]interface{}{
			"self_hosted": true,
			"unlimited":   true,
		},
	}
}

// === Usage Tracking ===

func (d *DisabledProvider) TrackUsage(ctx context.Context, event UsageEvent) error {
	// Ne rien faire, pas de tracking nécessaire
	return nil
}

func (d *DisabledProvider) GetCurrentUsage(ctx context.Context, userID string) (*Usage, error) {
	return &Usage{
		UserID:        userID,
		StorageUsedGB: 0,
	}, nil
}

// === Quota Enforcement ===

func (d *DisabledProvider) CheckQuota(ctx context.Context, userID string, requestedBytes int64) (*QuotaCheckResult, error) {
	return &QuotaCheckResult{
		Allowed:        true,
		CurrentUsage:   0,
		Limit:          999999999 * 1024 * 1024 * 1024,
		RemainingBytes: 999999999 * 1024 * 1024 * 1024,
	}, nil
}

func (d *DisabledProvider) CheckP2PQuota(ctx context.Context, userID string, currentActiveShares int) (*P2PQuotaCheckResult, error) {
	return &P2PQuotaCheckResult{
		Allowed:         true,
		ActiveShares:    currentActiveShares,
		Limit:           999999999,
		RemainingShares: 999999999,
	}, nil
}

// === Invoices ===

func (d *DisabledProvider) GetInvoices(ctx context.Context, userID string, limit int) ([]Invoice, error) {
	return []Invoice{}, nil
}

func (d *DisabledProvider) GetPaymentLink(ctx context.Context, invoiceID string) (string, error) {
	return "", nil
}

// === Stripe Checkout (disabled) ===

func (d *DisabledProvider) CreateCheckoutSession(ctx context.Context, userID, planCode, interval, successURL, cancelURL string) (string, error) {
	return "", fmt.Errorf("billing disabled in self-hosted mode")
}

func (d *DisabledProvider) CreatePortalSession(ctx context.Context, userID, returnURL string) (string, error) {
	return "", fmt.Errorf("billing disabled in self-hosted mode")
}
