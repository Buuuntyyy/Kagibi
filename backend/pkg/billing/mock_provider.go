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
	subscription  *Subscription
	storageUsedGB float64
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
			Description:    "20 Go de stockage",
			StorageLimitGB: 20,
			PriceMonthly:   0, PriceYearly: 0, Currency: "EUR",
		},
		{
			Code: "personal", Name: "Personnel",
			Description:    "200 Go de stockage, versioning",
			StorageLimitGB: 200,
			PriceMonthly:   400, PriceYearly: 4000, Currency: "EUR",
			Features: map[string]interface{}{"versioning": true},
		},
		{
			Code: "business", Name: "Business",
			Description:    "1 To de stockage, versioning, organisations",
			StorageLimitGB: 1024,
			PriceMonthly:   1400, PriceYearly: 14000, Currency: "EUR",
			Features: map[string]interface{}{"versioning": true, "orgs": true, "priority_support": true},
		},
		{
			Code: "payg", Name: "Pay as you go",
			Description:    "Stockage à la demande (15 €/To/mois)",
			StorageLimitGB: -1,
			PriceMonthly:   1500, Currency: "EUR",
			Features: map[string]interface{}{"versioning": true, "orgs": true, "billing_model": "payg"},
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
