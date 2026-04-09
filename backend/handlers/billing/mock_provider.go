package billing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

const errPaymentNotFound = "payment %s not found"

// MockPaymentProvider implements PaymentProvider for testing purposes
type MockPaymentProvider struct {
	mu           sync.RWMutex
	customers    map[string]*Customer
	paymentLinks map[string]*PaymentLink
	payments     map[string]*PaymentStatus

	// Configurable behavior for testing
	ShouldFailCreateCustomer    bool
	ShouldFailCreatePaymentLink bool
	ShouldFailGetPaymentStatus  bool
	SimulatedLatency            time.Duration
}

// NewMockPaymentProvider creates a new mock payment provider
func NewMockPaymentProvider() *MockPaymentProvider {
	return &MockPaymentProvider{
		customers:    make(map[string]*Customer),
		paymentLinks: make(map[string]*PaymentLink),
		payments:     make(map[string]*PaymentStatus),
	}
}

// CreateCustomer creates a mock customer
func (m *MockPaymentProvider) CreateCustomer(ctx context.Context, req CreateCustomerRequest) (*Customer, error) {
	if m.SimulatedLatency > 0 {
		select {
		case <-time.After(m.SimulatedLatency):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if m.ShouldFailCreateCustomer {
		return nil, &PaymentProviderError{
			Code:    "MOCK_ERROR",
			Message: "simulated create customer failure",
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	customer := &Customer{
		ID:         "cust_mock_" + uuid.New().String()[:8],
		ExternalID: req.ExternalID,
		Email:      req.Email,
		Name:       req.Name,
	}

	m.customers[customer.ID] = customer
	return customer, nil
}

// CreatePaymentLink creates a mock payment link
func (m *MockPaymentProvider) CreatePaymentLink(ctx context.Context, req CreatePaymentLinkRequest) (*PaymentLink, error) {
	if m.SimulatedLatency > 0 {
		select {
		case <-time.After(m.SimulatedLatency):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if m.ShouldFailCreatePaymentLink {
		return nil, &PaymentProviderError{
			Code:    "MOCK_ERROR",
			Message: "simulated create payment link failure",
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	paymentID := "pay_mock_" + uuid.New().String()[:8]
	link := &PaymentLink{
		ID:        paymentID,
		URL:       fmt.Sprintf("https://mock-payment.example.com/pay/%s", paymentID),
		Amount:    req.Amount,
		Currency:  req.Currency,
		Status:    "pending",
		ExpiresAt: time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		InvoiceID: req.InvoiceID,
	}

	m.paymentLinks[link.ID] = link
	m.payments[paymentID] = &PaymentStatus{
		ID:        paymentID,
		Status:    "pending",
		Amount:    req.Amount,
		Currency:  req.Currency,
		InvoiceID: req.InvoiceID,
	}

	return link, nil
}

// GetPaymentStatus retrieves the mock payment status
func (m *MockPaymentProvider) GetPaymentStatus(ctx context.Context, paymentID string) (*PaymentStatus, error) {
	if m.SimulatedLatency > 0 {
		select {
		case <-time.After(m.SimulatedLatency):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if m.ShouldFailGetPaymentStatus {
		return nil, &PaymentProviderError{
			Code:    "MOCK_ERROR",
			Message: "simulated get payment status failure",
		}
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	status, exists := m.payments[paymentID]
	if !exists {
		return nil, &PaymentProviderError{
			Code:    "NOT_FOUND",
			Message: fmt.Sprintf(errPaymentNotFound, paymentID),
		}
	}

	return status, nil
}

// SimulatePaymentSuccess simulates a successful payment (for testing webhooks)
func (m *MockPaymentProvider) SimulatePaymentSuccess(paymentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	status, exists := m.payments[paymentID]
	if !exists {
		return fmt.Errorf(errPaymentNotFound, paymentID)
	}

	status.Status = "paid"
	status.PaidAt = time.Now().Format(time.RFC3339)

	if link, ok := m.paymentLinks[paymentID]; ok {
		link.Status = "paid"
	}

	return nil
}

// SimulatePaymentFailure simulates a failed payment (for testing)
func (m *MockPaymentProvider) SimulatePaymentFailure(paymentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	status, exists := m.payments[paymentID]
	if !exists {
		return fmt.Errorf(errPaymentNotFound, paymentID)
	}

	status.Status = "failed"

	if link, ok := m.paymentLinks[paymentID]; ok {
		link.Status = "failed"
	}

	return nil
}

// GetAllPayments returns all payments (for testing assertions)
func (m *MockPaymentProvider) GetAllPayments() map[string]*PaymentStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*PaymentStatus)
	for k, v := range m.payments {
		result[k] = v
	}
	return result
}

// Reset clears all mock data
func (m *MockPaymentProvider) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.customers = make(map[string]*Customer)
	m.paymentLinks = make(map[string]*PaymentLink)
	m.payments = make(map[string]*PaymentStatus)
	m.ShouldFailCreateCustomer = false
	m.ShouldFailCreatePaymentLink = false
	m.ShouldFailGetPaymentStatus = false
	m.SimulatedLatency = 0
}
