// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	idempotency map[string]interface{}
}

type mockUser struct {
	subscription    *Subscription
	storageUsedGB   float64
	p2pSharesActive int
}

// NewMockProvider crée un nouveau mock billing provider
func NewMockProvider() *MockProvider {
	return &MockProvider{
		users:       make(map[string]*mockUser),
		idempotency: make(map[string]interface{}),
	}
}

// ProviderName implements the optional providerNamer interface.
func (m *MockProvider) ProviderName() string { return "mock" }

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

// === Subscription Management ===

func (m *MockProvider) CreateSubscription(ctx context.Context, userID, planCode, idempotencyKey string) (*Subscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if cached, ok := m.idempotency[idempotencyKey]; ok {
		return cached.(*Subscription), nil
	}
	now := time.Now()
	sub := &Subscription{
		ID:                 fmt.Sprintf("mock_sub_%s_%d", userID, now.Unix()),
		UserID:             userID,
		PlanCode:           planCode,
		Status:             "active",
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   now.AddDate(0, 1, 0),
	}
	m.idempotency[idempotencyKey] = sub
	if user, ok := m.users[userID]; ok {
		user.subscription = sub
	} else {
		m.users[userID] = &mockUser{subscription: sub}
	}
	return sub, nil
}

func (m *MockProvider) GetSubscription(ctx context.Context, userID string) (*Subscription, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	user, exists := m.users[userID]
	if !exists {
		now := time.Now()
		return &Subscription{
			ID: fmt.Sprintf("mock_sub_%s", userID), UserID: userID, PlanCode: "free",
			Status: "active", CurrentPeriodStart: now, CurrentPeriodEnd: now.AddDate(0, 1, 0),
		}, nil
	}
	return user.subscription, nil
}

func (m *MockProvider) UpdateSubscription(ctx context.Context, userID, newPlanCode, idempotencyKey string) (*Subscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if cached, ok := m.idempotency[idempotencyKey]; ok {
		return cached.(*Subscription), nil
	}
	user, exists := m.users[userID]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", userID)
	}
	user.subscription.PlanCode = newPlanCode
	m.idempotency[idempotencyKey] = user.subscription
	return user.subscription, nil
}

func (m *MockProvider) CancelSubscription(ctx context.Context, userID, idempotencyKey string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.idempotency[idempotencyKey]; ok {
		return nil
	}
	user, exists := m.users[userID]
	if exists && user.subscription != nil {
		now := time.Now()
		user.subscription.Status = "canceled"
		user.subscription.CanceledAt = &now
	}
	m.idempotency[idempotencyKey] = true
	return nil
}

// === Plan Information ===

func (m *MockProvider) getPlans() []Plan {
	return []Plan{
		{
			Code: "free", Name: "Gratuit",
			Description:    "5 Go de stockage, 5 partages P2P actifs",
			StorageLimitGB: 5, P2PSharesLimit: 5,
			PriceMonthly: 0, PriceYearly: 0, Currency: "EUR",
			Features: map[string]interface{}{"p2p_enabled": true},
		},
		{
			Code: "pro", Name: "Pro",
			Description:    "50 Go de stockage, 50 partages P2P actifs",
			StorageLimitGB: 50, P2PSharesLimit: 50,
			PriceMonthly: 500, PriceYearly: 5000, Currency: "EUR",
			Features: map[string]interface{}{"p2p_enabled": true},
		},
		{
			Code: "business", Name: "Business",
			Description:    "200 Go de stockage, 200 partages P2P actifs",
			StorageLimitGB: 200, P2PSharesLimit: 200,
			PriceMonthly: 1500, PriceYearly: 15000, Currency: "EUR",
			Features: map[string]interface{}{"p2p_enabled": true, "priority_support": true},
		},
	}
}

func (m *MockProvider) GetPlan(ctx context.Context, planCode string) (*Plan, error) {
	for _, p := range m.getPlans() {
		if p.Code == planCode {
			return &p, nil
		}
	}
	free := m.getPlans()[0]
	return &free, nil
}

func (m *MockProvider) GetUserPlan(ctx context.Context, userID string) (*Plan, error) {
	sub, _ := m.GetSubscription(ctx, userID)
	return m.GetPlan(ctx, sub.PlanCode)
}

func (m *MockProvider) ListPlans(ctx context.Context) ([]Plan, error) {
	return m.getPlans(), nil
}

// === Usage Tracking ===

func (m *MockProvider) TrackUsage(ctx context.Context, event UsageEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if event.IdempotencyKey != "" {
		if _, ok := m.idempotency[event.IdempotencyKey]; ok {
			return nil
		}
		m.idempotency[event.IdempotencyKey] = true
	}
	user, ok := m.users[event.UserID]
	if !ok {
		return nil
	}
	gb := float64(event.Bytes) / (1024 * 1024 * 1024)
	switch event.EventType {
	case "storage_add":
		user.storageUsedGB += gb
	case "storage_remove":
		user.storageUsedGB -= gb
		if user.storageUsedGB < 0 {
			user.storageUsedGB = 0
		}
	}
	return nil
}

func (m *MockProvider) GetCurrentUsage(ctx context.Context, userID string) (*Usage, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	now := time.Now()
	u := &Usage{UserID: userID, PeriodStart: now.AddDate(0, -1, 0), PeriodEnd: now}
	if user, ok := m.users[userID]; ok {
		u.StorageUsedGB = user.storageUsedGB
		u.P2PSharesActive = user.p2pSharesActive
	}
	return u, nil
}

// === Quota Enforcement ===

func (m *MockProvider) CheckQuota(ctx context.Context, userID string, requestedBytes int64) (*QuotaCheckResult, error) {
	plan, _ := m.GetUserPlan(ctx, userID)
	usage, _ := m.GetCurrentUsage(ctx, userID)
	limitBytes := plan.StorageLimitGB * 1024 * 1024 * 1024
	currentBytes := int64(usage.StorageUsedGB * 1024 * 1024 * 1024)
	remaining := limitBytes - currentBytes
	result := &QuotaCheckResult{CurrentUsage: currentBytes, Limit: limitBytes, RemainingBytes: remaining}
	if requestedBytes > remaining {
		result.Allowed = false
		result.Reason = fmt.Sprintf("Quota dépassé. Restant: %.2f Go", float64(remaining)/(1024*1024*1024))
	} else {
		result.Allowed = true
	}
	return result, nil
}

func (m *MockProvider) CheckP2PQuota(ctx context.Context, userID string, currentActiveShares int) (*P2PQuotaCheckResult, error) {
	plan, _ := m.GetUserPlan(ctx, userID)
	remaining := plan.P2PSharesLimit - currentActiveShares
	result := &P2PQuotaCheckResult{
		ActiveShares:    currentActiveShares,
		Limit:           plan.P2PSharesLimit,
		RemainingShares: remaining,
	}
	if remaining <= 0 {
		result.Allowed = false
		result.Reason = fmt.Sprintf("Limite de %d partages P2P actifs atteinte pour le plan %s", plan.P2PSharesLimit, plan.Name)
	} else {
		result.Allowed = true
	}
	return result, nil
}

// === Invoices ===

func (m *MockProvider) GetInvoices(ctx context.Context, userID string, limit int) ([]Invoice, error) {
	return []Invoice{}, nil
}

func (m *MockProvider) GetPaymentLink(ctx context.Context, invoiceID string) (string, error) {
	return "", fmt.Errorf("no payment required in free plan")
}

// === Stripe Checkout (Mock — redirect directly to success/return) ===

func (m *MockProvider) CreateCheckoutSession(ctx context.Context, userID, planCode, interval, successURL, cancelURL string) (string, error) {
	log.Printf("[MockBilling] CreateCheckoutSession: user=%s plan=%s interval=%s (mock)", userID, planCode, interval)
	return successURL, nil
}

func (m *MockProvider) CreatePortalSession(ctx context.Context, stripeCustomerID, returnURL string) (string, error) {
	log.Printf("[MockBilling] CreatePortalSession: customer=%s (mock)", stripeCustomerID)
	return returnURL, nil
}
