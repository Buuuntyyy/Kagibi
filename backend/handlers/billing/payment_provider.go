package billing

import "context"

// PaymentProvider defines the interface for payment gateway integrations
// This follows the Dependency Inversion Principle for clean architecture
type PaymentProvider interface {
	// CreateCustomer creates a customer in the payment gateway
	CreateCustomer(ctx context.Context, req CreateCustomerRequest) (*Customer, error)

	// CreatePaymentLink generates a payment link for an invoice
	CreatePaymentLink(ctx context.Context, req CreatePaymentLinkRequest) (*PaymentLink, error)

	// GetPaymentStatus retrieves the status of a payment
	GetPaymentStatus(ctx context.Context, paymentID string) (*PaymentStatus, error)
}

// CreateCustomerRequest contains data needed to create a customer
type CreateCustomerRequest struct {
	ExternalID string
	Email      string
	Name       string
	Metadata   map[string]string
}

// Customer represents a customer in the payment system
type Customer struct {
	ID         string
	ExternalID string
	Email      string
	Name       string
}

// CreatePaymentLinkRequest contains data needed to create a payment link
type CreatePaymentLinkRequest struct {
	CustomerID  string
	InvoiceID   string
	Amount      int64  // Amount in cents
	Currency    string // ISO 4217 currency code (e.g., "EUR")
	Description string
	RedirectURL string
	WebhookURL  string
	Metadata    map[string]string
}

// PaymentLink represents a generated payment link
type PaymentLink struct {
	ID        string
	URL       string
	Amount    int64
	Currency  string
	Status    string
	ExpiresAt string
	InvoiceID string
}

// PaymentStatus represents the current status of a payment
type PaymentStatus struct {
	ID        string
	Status    string // pending, paid, failed, expired, canceled
	PaidAt    string
	Amount    int64
	Currency  string
	InvoiceID string
}

// PaymentProviderError represents errors from the payment provider
type PaymentProviderError struct {
	Code    string
	Message string
	Err     error
}

func (e *PaymentProviderError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *PaymentProviderError) Unwrap() error {
	return e.Err
}
