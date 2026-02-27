package billing

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// StripeWebhookReceiver gère les webhooks entrants du service billing privé Stripe.
// Ce n'est PAS un webhook Stripe direct — c'est votre service privé qui relaie
// les événements Stripe vers SaferCloud après les avoir traités.
type StripeWebhookReceiver struct {
	secret string
}

// NewStripeWebhookReceiver crée un nouveau receiver
func NewStripeWebhookReceiver() *StripeWebhookReceiver {
	return &StripeWebhookReceiver{
		secret: os.Getenv("STRIPE_WEBHOOK_SECRET"),
	}
}

// StripeWebhookEvent représente un événement relayé par le service billing privé
type StripeWebhookEvent struct {
	Type      string          `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

// SubscriptionUpdatedData données pour subscription.updated / subscription.created
type SubscriptionUpdatedData struct {
	UserID   string `json:"user_id"`
	PlanCode string `json:"plan_code"`
	Status   string `json:"status"`
}

// SubscriptionCancelledData données pour subscription.cancelled
type SubscriptionCancelledData struct {
	UserID string `json:"user_id"`
}

// PaymentData données pour payment.*
type PaymentData struct {
	UserID    string `json:"user_id"`
	InvoiceID string `json:"invoice_id"`
	Amount    int64  `json:"amount_cents"`
	Currency  string `json:"currency"`
	Status    string `json:"status"`
}

// QuotaUpdatedData données pour quota.updated (poussé par le service billing)
type QuotaUpdatedData struct {
	UserID         string `json:"user_id"`
	StorageLimitGB int64  `json:"storage_limit_gb"`
	PlanCode       string `json:"plan_code"`
}

// HandleStripeWebhook gère les webhooks relayés par le service billing Stripe privé
func (r *StripeWebhookReceiver) HandleStripeWebhook(c *gin.Context) {
	// Lire le body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("[StripeWebhook] Failed to read body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read body"})
		return
	}

	// Vérifier la signature HMAC
	if r.secret != "" {
		signature := c.GetHeader("X-Webhook-Signature")
		timestamp := c.GetHeader("X-Webhook-Timestamp")

		if !r.verifySignature(body, signature, timestamp) {
			log.Printf("[StripeWebhook] Invalid signature")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
			return
		}
	}

	// Parser l'événement
	var event StripeWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("[StripeWebhook] Failed to parse event: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event format"})
		return
	}

	log.Printf("[StripeWebhook] Event received: %s", event.Type)

	// Router l'événement
	switch event.Type {
	case "subscription.created", "subscription.updated":
		r.handleSubscriptionUpdated(event.Data)
	case "subscription.cancelled", "subscription.deleted":
		r.handleSubscriptionCancelled(event.Data)
	case "payment.succeeded":
		r.handlePaymentSucceeded(event.Data)
	case "payment.failed":
		r.handlePaymentFailed(event.Data)
	case "quota.updated":
		r.handleQuotaUpdated(event.Data)
	default:
		log.Printf("[StripeWebhook] Unhandled event type: %s", event.Type)
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}

func (r *StripeWebhookReceiver) verifySignature(body []byte, signature, timestamp string) bool {
	if signature == "" || timestamp == "" {
		return false
	}

	// Vérifier que le timestamp n'est pas trop vieux (5 min)
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}
	if time.Since(time.Unix(ts, 0)) > 5*time.Minute {
		log.Printf("[StripeWebhook] Timestamp too old")
		return false
	}

	// HMAC-SHA256(secret, timestamp.body)
	payload := fmt.Sprintf("%s.%s", timestamp, string(body))
	mac := hmac.New(sha256.New, []byte(r.secret))
	mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(strings.ToLower(signature)), []byte(strings.ToLower(expected)))
}

func (r *StripeWebhookReceiver) handleSubscriptionUpdated(data json.RawMessage) {
	var sub SubscriptionUpdatedData
	if err := json.Unmarshal(data, &sub); err != nil {
		log.Printf("[StripeWebhook] Failed to parse subscription data: %v", err)
		return
	}
	log.Printf("[StripeWebhook] Subscription updated: user=%s plan=%s status=%s", sub.UserID, sub.PlanCode, sub.Status)
	// TODO: Mettre à jour le plan en base
	// db.NewUpdate().Model(&pkg.User{}).Set("plan = ?", sub.PlanCode).Where("id = ?", sub.UserID).Exec(ctx)
}

func (r *StripeWebhookReceiver) handleSubscriptionCancelled(data json.RawMessage) {
	var sub SubscriptionCancelledData
	if err := json.Unmarshal(data, &sub); err != nil {
		log.Printf("[StripeWebhook] Failed to parse cancellation data: %v", err)
		return
	}
	log.Printf("[StripeWebhook] Subscription cancelled: user=%s → downgrade to free", sub.UserID)
	// TODO: Rétrograder au plan gratuit en base
}

func (r *StripeWebhookReceiver) handlePaymentSucceeded(data json.RawMessage) {
	var payment PaymentData
	if err := json.Unmarshal(data, &payment); err != nil {
		log.Printf("[StripeWebhook] Failed to parse payment data: %v", err)
		return
	}
	log.Printf("[StripeWebhook] Payment succeeded: user=%s invoice=%s amount=%d%s",
		payment.UserID, payment.InvoiceID, payment.Amount, payment.Currency)
}

func (r *StripeWebhookReceiver) handlePaymentFailed(data json.RawMessage) {
	var payment PaymentData
	if err := json.Unmarshal(data, &payment); err != nil {
		log.Printf("[StripeWebhook] Failed to parse payment data: %v", err)
		return
	}
	log.Printf("[StripeWebhook] Payment failed: user=%s invoice=%s", payment.UserID, payment.InvoiceID)
	// TODO: Envoyer notification à l'utilisateur, grace period, etc.
}

func (r *StripeWebhookReceiver) handleQuotaUpdated(data json.RawMessage) {
	var quota QuotaUpdatedData
	if err := json.Unmarshal(data, &quota); err != nil {
		log.Printf("[StripeWebhook] Failed to parse quota data: %v", err)
		return
	}
	log.Printf("[StripeWebhook] Quota updated: user=%s storage=%dGB plan=%s",
		quota.UserID, quota.StorageLimitGB, quota.PlanCode)
	// TODO: Mettre à jour storage_limit en base
	// newLimit := quota.StorageLimitGB * 1024 * 1024 * 1024
	// db.NewUpdate().Model(&pkg.User{}).Set("storage_limit = ?", newLimit).Where("id = ?", quota.UserID).Exec(ctx)
}

// RegisterWebhookRoute enregistre la route webhook du service billing Stripe privé
func RegisterWebhookRoute(router *gin.RouterGroup) {
	receiver := NewStripeWebhookReceiver()
	router.POST("/webhooks/stripe", receiver.HandleStripeWebhook)
}
