package billing

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

const subscriptionPath = "/api/subscriptions/%s"

// WebhookProvider communique avec un service de facturation externe via API REST
// Utilisé en production pour se connecter au service privé de billing
type WebhookProvider struct {
	baseURL    string
	secret     string
	httpClient *http.Client

	// Cache local pour réduire les appels API
	planCache     map[string]*Plan
	planCacheTTL  time.Duration
	planCacheTime time.Time
	cacheMu       sync.RWMutex
}

// NewWebhookProvider crée un nouveau provider webhook
func NewWebhookProvider(baseURL, secret string) *WebhookProvider {
	return &WebhookProvider{
		baseURL: baseURL,
		secret:  secret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		planCache:    make(map[string]*Plan),
		planCacheTTL: 5 * time.Minute,
	}
}

// NewWebhookProviderFromEnv crée un provider depuis les variables d'environnement
func NewWebhookProviderFromEnv() (*WebhookProvider, error) {
	baseURL := os.Getenv("BILLING_SERVICE_URL")
	if baseURL == "" {
		return nil, fmt.Errorf("BILLING_SERVICE_URL not set")
	}

	secret := os.Getenv("BILLING_SERVICE_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("BILLING_SERVICE_SECRET not set")
	}

	return NewWebhookProvider(baseURL, secret), nil
}

// === Request Signing ===

func (w *WebhookProvider) signRequest(body []byte, timestamp string) string {
	message := fmt.Sprintf("%s.%s", timestamp, string(body))
	mac := hmac.New(sha256.New, []byte(w.secret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

func (w *WebhookProvider) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var bodyBytes []byte
	var err error

	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, w.baseURL+path, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Ajouter les headers de signature
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	signature := w.signRequest(bodyBytes, timestamp)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Billing-Timestamp", timestamp)
	req.Header.Set("X-Billing-Signature", signature)

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("billing service error (%d): %s", resp.StatusCode, string(respBody))
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}

// === Lifecycle Events ===

func (w *WebhookProvider) OnUserCreated(ctx context.Context, event UserCreatedEvent) error {
	return w.doRequest(ctx, "POST", "/api/users", event, nil)
}

func (w *WebhookProvider) OnUserDeleted(ctx context.Context, userID string) error {
	return w.doRequest(ctx, "DELETE", fmt.Sprintf("/api/users/%s", userID), nil, nil)
}

// === Subscription Management ===

type createSubscriptionRequest struct {
	UserID         string `json:"user_id"`
	PlanCode       string `json:"plan_code"`
	IdempotencyKey string `json:"idempotency_key"`
}

func (w *WebhookProvider) CreateSubscription(ctx context.Context, userID, planCode, idempotencyKey string) (*Subscription, error) {
	var result Subscription
	err := w.doRequest(ctx, "POST", "/api/subscriptions", createSubscriptionRequest{
		UserID:         userID,
		PlanCode:       planCode,
		IdempotencyKey: idempotencyKey,
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (w *WebhookProvider) GetSubscription(ctx context.Context, userID string) (*Subscription, error) {
	var result Subscription
	err := w.doRequest(ctx, "GET", fmt.Sprintf(subscriptionPath, userID), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

type updateSubscriptionRequest struct {
	PlanCode       string `json:"plan_code"`
	IdempotencyKey string `json:"idempotency_key"`
}

func (w *WebhookProvider) UpdateSubscription(ctx context.Context, userID, newPlanCode, idempotencyKey string) (*Subscription, error) {
	var result Subscription
	err := w.doRequest(ctx, "PUT", fmt.Sprintf(subscriptionPath, userID), updateSubscriptionRequest{
		PlanCode:       newPlanCode,
		IdempotencyKey: idempotencyKey,
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

type cancelSubscriptionRequest struct {
	IdempotencyKey string `json:"idempotency_key"`
}

func (w *WebhookProvider) CancelSubscription(ctx context.Context, userID, idempotencyKey string) error {
	return w.doRequest(ctx, "DELETE", fmt.Sprintf(subscriptionPath, userID), cancelSubscriptionRequest{
		IdempotencyKey: idempotencyKey,
	}, nil)
}

// === Plan Information ===

func (w *WebhookProvider) GetPlan(ctx context.Context, planCode string) (*Plan, error) {
	// Vérifier le cache
	w.cacheMu.RLock()
	if plan, ok := w.planCache[planCode]; ok && time.Since(w.planCacheTime) < w.planCacheTTL {
		w.cacheMu.RUnlock()
		return plan, nil
	}
	w.cacheMu.RUnlock()

	var result Plan
	err := w.doRequest(ctx, "GET", fmt.Sprintf("/api/plans/%s", planCode), nil, &result)
	if err != nil {
		return nil, err
	}

	// Mettre en cache
	w.cacheMu.Lock()
	w.planCache[planCode] = &result
	w.planCacheTime = time.Now()
	w.cacheMu.Unlock()

	return &result, nil
}

func (w *WebhookProvider) GetUserPlan(ctx context.Context, userID string) (*Plan, error) {
	var result Plan
	err := w.doRequest(ctx, "GET", fmt.Sprintf("/api/users/%s/plan", userID), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (w *WebhookProvider) ListPlans(ctx context.Context) ([]Plan, error) {
	var result []Plan
	err := w.doRequest(ctx, "GET", "/api/plans", nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// === Usage Tracking ===

func (w *WebhookProvider) TrackUsage(ctx context.Context, event UsageEvent) error {
	return w.doRequest(ctx, "POST", "/api/usage", event, nil)
}

func (w *WebhookProvider) GetCurrentUsage(ctx context.Context, userID string) (*Usage, error) {
	var result Usage
	err := w.doRequest(ctx, "GET", fmt.Sprintf("/api/usage/%s", userID), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// === Quota Enforcement ===

type checkQuotaRequest struct {
	UserID         string `json:"user_id"`
	RequestedBytes int64  `json:"requested_bytes"`
}

func (w *WebhookProvider) CheckQuota(ctx context.Context, userID string, requestedBytes int64) (*QuotaCheckResult, error) {
	var result QuotaCheckResult
	err := w.doRequest(ctx, "POST", "/api/quota/check", checkQuotaRequest{
		UserID:         userID,
		RequestedBytes: requestedBytes,
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// === Invoices ===

func (w *WebhookProvider) GetInvoices(ctx context.Context, userID string, limit int) ([]Invoice, error) {
	var result []Invoice
	err := w.doRequest(ctx, "GET", fmt.Sprintf("/api/users/%s/invoices?limit=%d", userID, limit), nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *WebhookProvider) GetPaymentLink(ctx context.Context, invoiceID string) (string, error) {
	var result struct {
		URL string `json:"url"`
	}
	err := w.doRequest(ctx, "GET", fmt.Sprintf("/api/invoices/%s/payment-link", invoiceID), nil, &result)
	if err != nil {
		return "", err
	}
	return result.URL, nil
}
