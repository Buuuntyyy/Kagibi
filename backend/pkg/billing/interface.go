// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package billing

import (
	"context"
	"time"
)

// =============================================================================
// BILLING PROVIDER INTERFACE - Client API vers service externe
// =============================================================================

// BillingProvider définit l'interface client pour communiquer avec un service
// de facturation externe. Cette interface est conçue comme un CLIENT API,
// sans aucune dépendance au code source du service billing.
//
// La version open-source inclut un MockProvider pour le développement.
// En production, le WebhookProvider communique avec votre service privé.
type BillingProvider interface {
	// === Lifecycle Events (Fire-and-forget, async) ===

	// OnUserCreated notifie le service billing de la création d'un utilisateur
	OnUserCreated(ctx context.Context, event UserCreatedEvent) error

	// OnUserDeleted notifie le service billing de la suppression d'un utilisateur
	OnUserDeleted(ctx context.Context, userID string) error

	// === Subscription Management (avec idempotence) ===

	// CreateSubscription crée ou récupère un abonnement existant (idempotent)
	// idempotencyKey doit être unique par requête (ex: "sub_{userID}_{planCode}_{timestamp}")
	CreateSubscription(ctx context.Context, userID, planCode, idempotencyKey string) (*Subscription, error)

	// GetSubscription récupère l'abonnement actif d'un utilisateur
	GetSubscription(ctx context.Context, userID string) (*Subscription, error)

	// UpdateSubscription change le plan d'un utilisateur (idempotent)
	UpdateSubscription(ctx context.Context, userID, newPlanCode, idempotencyKey string) (*Subscription, error)

	// CancelSubscription annule l'abonnement d'un utilisateur
	CancelSubscription(ctx context.Context, userID, idempotencyKey string) error

	// === Plan Information ===

	// GetPlan récupère les détails d'un plan
	GetPlan(ctx context.Context, planCode string) (*Plan, error)

	// GetUserPlan récupère le plan actif d'un utilisateur
	GetUserPlan(ctx context.Context, userID string) (*Plan, error)

	// ListPlans liste tous les plans disponibles
	ListPlans(ctx context.Context) ([]Plan, error)

	// === Usage Tracking ===

	// TrackUsage enregistre un événement d'usage (async, fire-and-forget)
	TrackUsage(ctx context.Context, event UsageEvent) error

	// GetCurrentUsage récupère l'usage de la période en cours
	GetCurrentUsage(ctx context.Context, userID string) (*Usage, error)

	// === Quota Enforcement ===

	// CheckQuota vérifie si une opération est autorisée par le quota stockage
	CheckQuota(ctx context.Context, userID string, requestedBytes int64) (*QuotaCheckResult, error)

	// CheckP2PQuota vérifie si un nouveau partage P2P peut être créé
	CheckP2PQuota(ctx context.Context, userID string, currentActiveShares int) (*P2PQuotaCheckResult, error)

	// === Invoices (Read-only depuis le core) ===

	// GetInvoices récupère les factures d'un utilisateur
	GetInvoices(ctx context.Context, userID string, limit int) ([]Invoice, error)

	// GetPaymentLink génère un lien de paiement pour une facture
	GetPaymentLink(ctx context.Context, invoiceID string) (string, error)

	// === Stripe Checkout (upgrade de plan) ===

	// CreateCheckoutSession crée une session Stripe Checkout pour upgrade
	// interval: "monthly" ou "yearly"
	CreateCheckoutSession(ctx context.Context, userID, planCode, interval, successURL, cancelURL string) (string, error)

	// CreatePortalSession crée une session Stripe Customer Portal
	CreatePortalSession(ctx context.Context, userID, returnURL string) (string, error)
}

// =============================================================================
// EVENT TYPES
// =============================================================================

// UserCreatedEvent est émis lors de l'inscription d'un utilisateur
type UserCreatedEvent struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	Timestamp time.Time `json:"timestamp"`
}

// UsageEvent représente un événement d'usage facturable
type UsageEvent struct {
	UserID         string                 `json:"user_id"`
	EventType      string                 `json:"event_type"` // "storage_add", "storage_remove", "bandwidth", "p2p_transfer"
	Bytes          int64                  `json:"bytes"`
	Timestamp      time.Time              `json:"timestamp"`
	ResourceID     string                 `json:"resource_id,omitempty"`
	Description    string                 `json:"description,omitempty"`
	IdempotencyKey string                 `json:"idempotency_key,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// =============================================================================
// DATA MODELS
// =============================================================================

// Subscription représente un abonnement utilisateur
type Subscription struct {
	ID                 string     `json:"id"`
	UserID             string     `json:"user_id"`
	PlanCode           string     `json:"plan_code"`
	Status             string     `json:"status"` // "active", "canceled", "past_due", "trialing", "incomplete"
	CurrentPeriodStart time.Time  `json:"current_period_start"`
	CurrentPeriodEnd   time.Time  `json:"current_period_end"`
	CanceledAt         *time.Time `json:"canceled_at,omitempty"`
	TrialEndsAt        *time.Time `json:"trial_ends_at,omitempty"`
	StripeSubID        string     `json:"stripe_subscription_id,omitempty"`
}

// Plan représente un plan tarifaire avec quotas
type Plan struct {
	Code           string `json:"code"`
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
	StorageLimitGB int64  `json:"storage_limit_gb"`
	P2PSharesLimit int    `json:"p2p_shares_limit"` // max simultaneous active P2P shares
	PriceMonthly   int64  `json:"price_monthly_cents"`
	PriceYearly    int64  `json:"price_yearly_cents,omitempty"` // 0 = not available
	Currency       string `json:"currency"`
	// StripePriceIDMonthly / Yearly are the IDs from your Stripe dashboard
	StripePriceIDMonthly string                 `json:"stripe_price_id_monthly,omitempty"`
	StripePriceIDYearly  string                 `json:"stripe_price_id_yearly,omitempty"`
	Features             map[string]interface{} `json:"features,omitempty"`
}

// Usage représente l'usage de la période en cours
type Usage struct {
	UserID          string    `json:"user_id"`
	PeriodStart     time.Time `json:"period_start"`
	PeriodEnd       time.Time `json:"period_end"`
	StorageUsedGB   float64   `json:"storage_used_gb"`
	P2PSharesActive int       `json:"p2p_shares_active"` // current active P2P shares count
}

// QuotaCheckResult est le résultat d'une vérification de quota
type QuotaCheckResult struct {
	Allowed        bool   `json:"allowed"`
	Reason         string `json:"reason,omitempty"`
	CurrentUsage   int64  `json:"current_usage_bytes"`
	Limit          int64  `json:"limit_bytes"`
	RemainingBytes int64  `json:"remaining_bytes"`
}

// P2PQuotaCheckResult est le résultat d'une vérification du quota P2P
type P2PQuotaCheckResult struct {
	Allowed         bool   `json:"allowed"`
	Reason          string `json:"reason,omitempty"`
	ActiveShares    int    `json:"active_shares"`
	Limit           int    `json:"limit"`
	RemainingShares int    `json:"remaining_shares"`
}

// Invoice représente une facture
type Invoice struct {
	ID          string     `json:"id"`
	Number      string     `json:"number"`
	Status      string     `json:"status"` // "draft", "open", "paid", "void", "uncollectible"
	AmountCents int64      `json:"amount_cents"`
	Currency    string     `json:"currency"`
	IssuedAt    time.Time  `json:"issued_at"`
	DueAt       time.Time  `json:"due_at"`
	PaidAt      *time.Time `json:"paid_at,omitempty"`
	PaymentURL  string     `json:"payment_url,omitempty"`
	DownloadURL string     `json:"download_url,omitempty"`
}

// =============================================================================
// GLOBAL PROVIDER INSTANCE
// =============================================================================

var provider BillingProvider

// SetProvider configure le provider de facturation global
func SetProvider(p BillingProvider) {
	provider = p
}

// GetProvider retourne le provider de facturation actuel
// Retourne MockProvider si aucun n'est configuré
func GetProvider() BillingProvider {
	if provider == nil {
		provider = NewMockProvider()
	}
	return provider
}
