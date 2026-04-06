package billing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

// StripeWebhookProvider communique avec votre service privé Stripe via API REST.
// Aucune dépendance directe à Stripe SDK — toute la logique Stripe
// est dans votre repo privé. Celui-ci est un simple client HTTP.
type StripeWebhookProvider struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client

	// Cache local pour réduire les appels API
	planCache     map[string]*Plan
	planCacheTTL  time.Duration
	planCacheTime time.Time
	cacheMu       sync.RWMutex
}

// NewStripeWebhookProvider crée un nouveau provider depuis l'URL du service billing privé
func NewStripeWebhookProvider(baseURL, apiKey string) *StripeWebhookProvider {
	return &StripeWebhookProvider{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		planCache:    make(map[string]*Plan),
		planCacheTTL: 5 * time.Minute,
	}
}

// NewStripeWebhookProviderFromEnv crée un provider depuis les variables d'environnement
func NewStripeWebhookProviderFromEnv() (*StripeWebhookProvider, error) {
	baseURL := os.Getenv("STRIPE_BILLING_SERVICE_URL")
	apiKey := os.Getenv("STRIPE_BILLING_SERVICE_KEY")

	if baseURL == "" || apiKey == "" {
		return nil, fmt.Errorf("STRIPE_BILLING_SERVICE_URL and STRIPE_BILLING_SERVICE_KEY are required")
	}

	return NewStripeWebhookProvider(baseURL, apiKey), nil
}

// === HTTP Helpers ===

func (s *StripeWebhookProvider) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	url := fmt.Sprintf("%s%s", s.baseURL, path)
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))
	req.Header.Set("X-Service", "kagibi-core")

	return s.httpClient.Do(req)
}

func (s *StripeWebhookProvider) doJSON(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	resp, err := s.doRequest(ctx, method, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("billing service error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}

// === Lifecycle Events ===

func (s *StripeWebhookProvider) OnUserCreated(ctx context.Context, event UserCreatedEvent) error {
	return s.doJSON(ctx, "POST", "/api/v1/customers", event, nil)
}

func (s *StripeWebhookProvider) OnUserDeleted(ctx context.Context, userID string) error {
	return s.doJSON(ctx, "DELETE", fmt.Sprintf("/api/v1/customers/%s", userID), nil, nil)
}

// === Subscription Management ===

func (s *StripeWebhookProvider) CreateSubscription(ctx context.Context, userID, planCode, idempotencyKey string) (*Subscription, error) {
	var sub Subscription
	err := s.doJSON(ctx, "POST", fmt.Sprintf("/api/v1/customers/%s/subscriptions", userID), map[string]string{
		"plan_code":       planCode,
		"idempotency_key": idempotencyKey,
	}, &sub)
	return &sub, err
}

func (s *StripeWebhookProvider) GetSubscription(ctx context.Context, userID string) (*Subscription, error) {
	var sub Subscription
	err := s.doJSON(ctx, "GET", fmt.Sprintf("/api/v1/customers/%s/subscription", userID), nil, &sub)
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (s *StripeWebhookProvider) UpdateSubscription(ctx context.Context, userID, newPlanCode, idempotencyKey string) (*Subscription, error) {
	var sub Subscription
	err := s.doJSON(ctx, "PUT", fmt.Sprintf("/api/v1/customers/%s/subscription", userID), map[string]string{
		"plan_code":       newPlanCode,
		"idempotency_key": idempotencyKey,
	}, &sub)
	return &sub, err
}

func (s *StripeWebhookProvider) CancelSubscription(ctx context.Context, userID, idempotencyKey string) error {
	return s.doJSON(ctx, "DELETE", fmt.Sprintf("/api/v1/customers/%s/subscription", userID), map[string]string{
		"idempotency_key": idempotencyKey,
	}, nil)
}

// === Plan Information ===

func (s *StripeWebhookProvider) GetPlan(ctx context.Context, planCode string) (*Plan, error) {
	// Check cache
	s.cacheMu.RLock()
	if p, ok := s.planCache[planCode]; ok && time.Since(s.planCacheTime) < s.planCacheTTL {
		s.cacheMu.RUnlock()
		return p, nil
	}
	s.cacheMu.RUnlock()

	var plan Plan
	err := s.doJSON(ctx, "GET", fmt.Sprintf("/api/v1/plans/%s", planCode), nil, &plan)
	if err != nil {
		return nil, err
	}

	// Update cache
	s.cacheMu.Lock()
	s.planCache[planCode] = &plan
	s.planCacheTime = time.Now()
	s.cacheMu.Unlock()

	return &plan, nil
}

func (s *StripeWebhookProvider) GetUserPlan(ctx context.Context, userID string) (*Plan, error) {
	var plan Plan
	err := s.doJSON(ctx, "GET", fmt.Sprintf("/api/v1/customers/%s/plan", userID), nil, &plan)
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (s *StripeWebhookProvider) ListPlans(ctx context.Context) ([]Plan, error) {
	var plans []Plan
	err := s.doJSON(ctx, "GET", "/api/v1/plans", nil, &plans)
	return plans, err
}

// === Usage Tracking ===

func (s *StripeWebhookProvider) TrackUsage(ctx context.Context, event UsageEvent) error {
	// Fire-and-forget: on ne bloque pas si le service billing est down
	go func() {
		_ = s.doJSON(context.Background(), "POST", "/api/v1/usage", event, nil)
	}()
	return nil
}

func (s *StripeWebhookProvider) GetCurrentUsage(ctx context.Context, userID string) (*Usage, error) {
	var usage Usage
	err := s.doJSON(ctx, "GET", fmt.Sprintf("/api/v1/customers/%s/usage", userID), nil, &usage)
	if err != nil {
		return nil, err
	}
	return &usage, nil
}

// === Quota Enforcement ===

func (s *StripeWebhookProvider) CheckQuota(ctx context.Context, userID string, requestedBytes int64) (*QuotaCheckResult, error) {
	var result QuotaCheckResult
	err := s.doJSON(ctx, "POST", fmt.Sprintf("/api/v1/customers/%s/quota/check", userID), map[string]int64{
		"requested_bytes": requestedBytes,
	}, &result)
	if err != nil {
		// Fail-open: si le service billing est down, on autorise
		return &QuotaCheckResult{Allowed: true, Reason: "billing service unavailable"}, nil
	}
	return &result, nil
}

// === Invoices ===

func (s *StripeWebhookProvider) GetInvoices(ctx context.Context, userID string, limit int) ([]Invoice, error) {
	var invoices []Invoice
	err := s.doJSON(ctx, "GET", fmt.Sprintf("/api/v1/customers/%s/invoices?limit=%d", userID, limit), nil, &invoices)
	if err != nil {
		return []Invoice{}, nil
	}
	return invoices, nil
}

func (s *StripeWebhookProvider) GetPaymentLink(ctx context.Context, invoiceID string) (string, error) {
	var result struct {
		PaymentURL string `json:"payment_url"`
	}
	err := s.doJSON(ctx, "GET", fmt.Sprintf("/api/v1/invoices/%s/payment-link", invoiceID), nil, &result)
	if err != nil {
		return "", err
	}
	return result.PaymentURL, nil
}

// === Stripe Checkout ===

func (s *StripeWebhookProvider) CreateCheckoutSession(ctx context.Context, userID, planCode, successURL, cancelURL string) (string, error) {
	var result struct {
		CheckoutURL string `json:"checkout_url"`
	}
	err := s.doJSON(ctx, "POST", "/api/v1/checkout/sessions", map[string]string{
		"user_id":     userID,
		"plan_code":   planCode,
		"success_url": successURL,
		"cancel_url":  cancelURL,
	}, &result)
	if err != nil {
		return "", err
	}
	return result.CheckoutURL, nil
}

func (s *StripeWebhookProvider) CreatePortalSession(ctx context.Context, userID, returnURL string) (string, error) {
	var result struct {
		PortalURL string `json:"portal_url"`
	}
	err := s.doJSON(ctx, "POST", "/api/v1/portal/sessions", map[string]string{
		"user_id":    userID,
		"return_url": returnURL,
	}, &result)
	if err != nil {
		return "", err
	}
	return result.PortalURL, nil
}
