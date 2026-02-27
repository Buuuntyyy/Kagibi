package billing

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// MockProvider implémente BillingProvider pour le développement/tests
// C'est le provider par défaut dans la version open-source
type MockProvider struct {
	mu          sync.RWMutex
	users       map[string]*mockUser
	idempotency map[string]interface{} // Cache pour idempotence
	defaultPlan *Plan
}

type mockUser struct {
	subscription *Subscription
	usage        *Usage
}

// NewMockProvider crée un nouveau mock billing provider
func NewMockProvider() *MockProvider {
	return &MockProvider{
		users:       make(map[string]*mockUser),
		idempotency: make(map[string]interface{}),
		defaultPlan: &Plan{
			Code:             "free",
			Name:             "Plan Gratuit",
			Description:      "5 Go de stockage gratuit",
			StorageLimitGB:   5,
			BandwidthLimitGB: 10,
			PriceMonthly:     0,
			Currency:         "EUR",
			Interval:         "monthly",
			Features: map[string]interface{}{
				"max_file_size_mb": 100,
				"p2p_enabled":      true,
			},
		},
	}
}

// === Lifecycle Events ===

func (m *MockProvider) OnUserCreated(ctx context.Context, event UserCreatedEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.Printf("[MockBilling] User created: %s (%s)", event.UserID, event.Email)

	now := time.Now()
	m.users[event.UserID] = &mockUser{
		subscription: &Subscription{
			ID:                 fmt.Sprintf("mock_sub_%s", event.UserID),
			UserID:             event.UserID,
			PlanCode:           "free",
			Status:             "active",
			CurrentPeriodStart: now,
			CurrentPeriodEnd:   now.AddDate(0, 1, 0),
		},
		usage: &Usage{
			UserID:      event.UserID,
			PeriodStart: now,
			PeriodEnd:   now.AddDate(0, 1, 0),
		},
	}

	return nil
}

func (m *MockProvider) OnUserDeleted(ctx context.Context, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.Printf("[MockBilling] User deleted: %s", userID)
	delete(m.users, userID)
	return nil
}

// === Subscription Management (avec idempotence) ===

func (m *MockProvider) CreateSubscription(ctx context.Context, userID, planCode, idempotencyKey string) (*Subscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Vérifier idempotence
	if cached, ok := m.idempotency[idempotencyKey]; ok {
		log.Printf("[MockBilling] Idempotent hit for CreateSubscription: %s", idempotencyKey)
		return cached.(*Subscription), nil
	}

	log.Printf("[MockBilling] Creating subscription: user=%s plan=%s", userID, planCode)

	now := time.Now()
	sub := &Subscription{
		ID:                 fmt.Sprintf("mock_sub_%s_%d", userID, now.Unix()),
		UserID:             userID,
		PlanCode:           planCode,
		Status:             "active",
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   now.AddDate(0, 1, 0),
	}

	// Mettre en cache pour idempotence
	m.idempotency[idempotencyKey] = sub

	// Mettre à jour l'utilisateur
	if user, ok := m.users[userID]; ok {
		user.subscription = sub
	} else {
		m.users[userID] = &mockUser{
			subscription: sub,
			usage: &Usage{
				UserID:      userID,
				PeriodStart: now,
				PeriodEnd:   now.AddDate(0, 1, 0),
			},
		}
	}

	return sub, nil
}

func (m *MockProvider) GetSubscription(ctx context.Context, userID string) (*Subscription, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, exists := m.users[userID]
	if !exists {
		// Auto-créer pour simplicité en dev
		now := time.Now()
		return &Subscription{
			ID:                 fmt.Sprintf("mock_sub_%s", userID),
			UserID:             userID,
			PlanCode:           "free",
			Status:             "active",
			CurrentPeriodStart: now,
			CurrentPeriodEnd:   now.AddDate(0, 1, 0),
		}, nil
	}

	return user.subscription, nil
}

func (m *MockProvider) UpdateSubscription(ctx context.Context, userID, newPlanCode, idempotencyKey string) (*Subscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Vérifier idempotence
	if cached, ok := m.idempotency[idempotencyKey]; ok {
		log.Printf("[MockBilling] Idempotent hit for UpdateSubscription: %s", idempotencyKey)
		return cached.(*Subscription), nil
	}

	log.Printf("[MockBilling] Updating subscription: user=%s newPlan=%s", userID, newPlanCode)

	user, exists := m.users[userID]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", userID)
	}

	user.subscription.PlanCode = newPlanCode

	// Mettre en cache pour idempotence
	m.idempotency[idempotencyKey] = user.subscription

	return user.subscription, nil
}

func (m *MockProvider) CancelSubscription(ctx context.Context, userID, idempotencyKey string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Vérifier idempotence
	if _, ok := m.idempotency[idempotencyKey]; ok {
		log.Printf("[MockBilling] Idempotent hit for CancelSubscription: %s", idempotencyKey)
		return nil
	}

	log.Printf("[MockBilling] Canceling subscription: user=%s", userID)

	user, exists := m.users[userID]
	if exists && user.subscription != nil {
		now := time.Now()
		user.subscription.Status = "canceled"
		user.subscription.CanceledAt = &now
	}

	// Marquer comme traité
	m.idempotency[idempotencyKey] = true

	return nil
}

// === Plan Information ===

func (m *MockProvider) GetPlan(ctx context.Context, planCode string) (*Plan, error) {
	plans := m.getAvailablePlans()
	for _, plan := range plans {
		if plan.Code == planCode {
			return &plan, nil
		}
	}
	return nil, fmt.Errorf("plan not found: %s", planCode)
}

func (m *MockProvider) GetUserPlan(ctx context.Context, userID string) (*Plan, error) {
	sub, err := m.GetSubscription(ctx, userID)
	if err != nil {
		return m.defaultPlan, nil
	}
	return m.GetPlan(ctx, sub.PlanCode)
}

func (m *MockProvider) ListPlans(ctx context.Context) ([]Plan, error) {
	return m.getAvailablePlans(), nil
}

func (m *MockProvider) getAvailablePlans() []Plan {
	return []Plan{
		{
			Code:             "free",
			Name:             "Gratuit",
			Description:      "5 Go de stockage chiffré",
			StorageLimitGB:   5,
			BandwidthLimitGB: 10,
			PriceMonthly:     0,
			Currency:         "EUR",
			Interval:         "monthly",
			Features: map[string]interface{}{
				"max_file_size_mb": 100,
				"p2p_enabled":      true,
				"p2p_limit_gb":     2,
			},
		},
		{
			Code:             "personal",
			Name:             "Personnel",
			Description:      "100 Go de stockage",
			StorageLimitGB:   100,
			BandwidthLimitGB: 500,
			PriceMonthly:     500, // 5€
			Currency:         "EUR",
			Interval:         "monthly",
			Features: map[string]interface{}{
				"max_file_size_mb": 500,
				"p2p_enabled":      true,
				"p2p_limit_gb":     50,
			},
		},
		{
			Code:             "expert",
			Name:             "Expert",
			Description:      "1 To de stockage",
			StorageLimitGB:   1024,
			BandwidthLimitGB: 2048,
			PriceMonthly:     1500, // 15€
			Currency:         "EUR",
			Interval:         "monthly",
			Features: map[string]interface{}{
				"max_file_size_mb": 5000,
				"p2p_enabled":      true,
				"p2p_limit_gb":     -1,
			},
		},
		{
			Code:             "enterprise",
			Name:             "Enterprise",
			Description:      "3 To de stockage",
			StorageLimitGB:   3072,
			BandwidthLimitGB: 10240,
			PriceMonthly:     4900, // 49€
			Currency:         "EUR",
			Interval:         "monthly",
			Features: map[string]interface{}{
				"max_file_size_mb": 10000,
				"p2p_enabled":      true,
				"p2p_limit_gb":     -1,
			},
		},
	}
}

// === Usage Tracking ===

func (m *MockProvider) TrackUsage(ctx context.Context, event UsageEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Vérifier idempotence si clé fournie
	if event.IdempotencyKey != "" {
		if _, ok := m.idempotency[event.IdempotencyKey]; ok {
			log.Printf("[MockBilling] Idempotent hit for TrackUsage: %s", event.IdempotencyKey)
			return nil
		}
		m.idempotency[event.IdempotencyKey] = true
	}

	log.Printf("[MockBilling] Usage tracked: user=%s type=%s bytes=%d",
		event.UserID, event.EventType, event.Bytes)

	user, exists := m.users[event.UserID]
	if !exists {
		return nil // Ignorer silencieusement
	}

	gb := float64(event.Bytes) / (1024 * 1024 * 1024)

	switch event.EventType {
	case "storage_add":
		user.usage.StorageUsedGB += gb
	case "storage_remove":
		user.usage.StorageUsedGB -= gb
		if user.usage.StorageUsedGB < 0 {
			user.usage.StorageUsedGB = 0
		}
	case "bandwidth":
		user.usage.BandwidthUsedGB += gb
	case "p2p_transfer":
		user.usage.P2PTransferGB += gb
	}

	return nil
}

func (m *MockProvider) GetCurrentUsage(ctx context.Context, userID string) (*Usage, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, exists := m.users[userID]
	if !exists {
		now := time.Now()
		return &Usage{
			UserID:      userID,
			PeriodStart: now,
			PeriodEnd:   now.AddDate(0, 1, 0),
		}, nil
	}

	return user.usage, nil
}

// === Quota Enforcement ===

func (m *MockProvider) CheckQuota(ctx context.Context, userID string, requestedBytes int64) (*QuotaCheckResult, error) {
	plan, _ := m.GetUserPlan(ctx, userID)
	usage, _ := m.GetCurrentUsage(ctx, userID)

	limitBytes := plan.StorageLimitGB * 1024 * 1024 * 1024
	currentBytes := int64(usage.StorageUsedGB * 1024 * 1024 * 1024)
	remaining := limitBytes - currentBytes

	result := &QuotaCheckResult{
		CurrentUsage:   currentBytes,
		Limit:          limitBytes,
		RemainingBytes: remaining,
	}

	if requestedBytes > remaining {
		result.Allowed = false
		result.Reason = fmt.Sprintf("Quota dépassé. Restant: %.2f Go", float64(remaining)/(1024*1024*1024))
	} else {
		result.Allowed = true
	}

	return result, nil
}

// === Invoices ===

func (m *MockProvider) GetInvoices(ctx context.Context, userID string, limit int) ([]Invoice, error) {
	// Mock: retourne liste vide (plan gratuit = pas de factures)
	return []Invoice{}, nil
}

func (m *MockProvider) GetPaymentLink(ctx context.Context, invoiceID string) (string, error) {
	return "", fmt.Errorf("no payment required in free plan")
}

// === Stripe Checkout (Mock) ===

func (m *MockProvider) CreateCheckoutSession(ctx context.Context, userID, planCode, successURL, cancelURL string) (string, error) {
	log.Printf("[MockBilling] CreateCheckoutSession: user=%s plan=%s (mock - redirecting to success)", userID, planCode)
	return successURL, nil
}

func (m *MockProvider) CreatePortalSession(ctx context.Context, userID, returnURL string) (string, error) {
	log.Printf("[MockBilling] CreatePortalSession: user=%s (mock - redirecting to return)", userID)
	return returnURL, nil
}
