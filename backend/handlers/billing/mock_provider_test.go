package billing

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockPaymentProvider_CreateCustomer(t *testing.T) {
	provider := NewMockPaymentProvider()

	ctx := context.Background()
	customer, err := provider.CreateCustomer(ctx, CreateCustomerRequest{
		ExternalID: "user-123",
		Email:      "test@example.com",
		Name:       "Test User",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, customer.ID)
	assert.Equal(t, "user-123", customer.ExternalID)
	assert.Equal(t, "test@example.com", customer.Email)
	assert.Equal(t, "Test User", customer.Name)
}

func TestMockPaymentProvider_CreateCustomerFailure(t *testing.T) {
	provider := NewMockPaymentProvider()
	provider.ShouldFailCreateCustomer = true

	ctx := context.Background()
	_, err := provider.CreateCustomer(ctx, CreateCustomerRequest{
		ExternalID: "user-123",
		Email:      "test@example.com",
		Name:       "Test User",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "simulated")
}

func TestMockPaymentProvider_CreatePaymentLink(t *testing.T) {
	provider := NewMockPaymentProvider()

	ctx := context.Background()
	link, err := provider.CreatePaymentLink(ctx, CreatePaymentLinkRequest{
		CustomerID:  "cust-123",
		InvoiceID:   "inv-456",
		Amount:      1500, // 15.00 EUR
		Currency:    "EUR",
		Description: "Test Invoice",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, link.ID)
	assert.NotEmpty(t, link.URL)
	assert.Equal(t, int64(1500), link.Amount)
	assert.Equal(t, "EUR", link.Currency)
	assert.Equal(t, "pending", link.Status)
	assert.Equal(t, "inv-456", link.InvoiceID)
}

func TestMockPaymentProvider_GetPaymentStatus(t *testing.T) {
	provider := NewMockPaymentProvider()

	ctx := context.Background()

	// First create a payment
	link, err := provider.CreatePaymentLink(ctx, CreatePaymentLinkRequest{
		CustomerID: "cust-123",
		InvoiceID:  "inv-456",
		Amount:     1500,
		Currency:   "EUR",
	})
	require.NoError(t, err)

	// Then get its status
	status, err := provider.GetPaymentStatus(ctx, link.ID)
	require.NoError(t, err)
	assert.Equal(t, link.ID, status.ID)
	assert.Equal(t, "pending", status.Status)
}

func TestMockPaymentProvider_SimulatePaymentSuccess(t *testing.T) {
	provider := NewMockPaymentProvider()
	ctx := context.Background()

	link, _ := provider.CreatePaymentLink(ctx, CreatePaymentLinkRequest{
		InvoiceID: "inv-789",
		Amount:    2000,
		Currency:  "EUR",
	})

	err := provider.SimulatePaymentSuccess(link.ID)
	require.NoError(t, err)

	status, _ := provider.GetPaymentStatus(ctx, link.ID)
	assert.Equal(t, "paid", status.Status)
	assert.NotEmpty(t, status.PaidAt)
}

func TestMockPaymentProvider_SimulatePaymentFailure(t *testing.T) {
	provider := NewMockPaymentProvider()
	ctx := context.Background()

	link, _ := provider.CreatePaymentLink(ctx, CreatePaymentLinkRequest{
		InvoiceID: "inv-789",
		Amount:    2000,
		Currency:  "EUR",
	})

	err := provider.SimulatePaymentFailure(link.ID)
	require.NoError(t, err)

	status, _ := provider.GetPaymentStatus(ctx, link.ID)
	assert.Equal(t, "failed", status.Status)
}

func TestMockPaymentProvider_Reset(t *testing.T) {
	provider := NewMockPaymentProvider()
	ctx := context.Background()

	provider.CreatePaymentLink(ctx, CreatePaymentLinkRequest{
		InvoiceID: "inv-1",
		Amount:    1000,
		Currency:  "EUR",
	})

	assert.Equal(t, 1, len(provider.GetAllPayments()))

	provider.Reset()

	assert.Equal(t, 0, len(provider.GetAllPayments()))
}

func TestMockPaymentProvider_ContextCancellation(t *testing.T) {
	provider := NewMockPaymentProvider()
	provider.SimulatedLatency = 100 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := provider.CreateCustomer(ctx, CreateCustomerRequest{
		ExternalID: "user-123",
	})

	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}
