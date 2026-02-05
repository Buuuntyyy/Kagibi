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
	"time"

	"github.com/gin-gonic/gin"
)

// WebhookReceiver gère les webhooks entrants du service de facturation
type WebhookReceiver struct {
	secret string
}

// NewWebhookReceiver crée un nouveau receiver de webhooks
func NewWebhookReceiver() *WebhookReceiver {
	return &WebhookReceiver{
		secret: os.Getenv("BILLING_SERVICE_SECRET"),
	}
}

// WebhookEvent représente un événement webhook entrant
type WebhookEvent struct {
	Type      string          `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

// SubscriptionUpdatedData données pour subscription.updated
type SubscriptionUpdatedData struct {
	UserID   string `json:"user_id"`
	PlanCode string `json:"plan_code"`
	Status   string `json:"status"`
}

// PaymentData données pour payment.*
type PaymentData struct {
	UserID    string  `json:"user_id"`
	InvoiceID string  `json:"invoice_id"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Status    string  `json:"status"`
}

// QuotaExceededData données pour quota.exceeded
type QuotaExceededData struct {
	UserID       string `json:"user_id"`
	QuotaType    string `json:"quota_type"` // storage, bandwidth
	CurrentUsage int64  `json:"current_usage"`
	Limit        int64  `json:"limit"`
}

// HandleWebhook gère les webhooks entrants
// POST /api/billing/webhooks
func (wr *WebhookReceiver) HandleWebhook(c *gin.Context) {
	// Lire le body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("[Webhook] Failed to read body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}

	// Vérifier la signature
	if !wr.verifySignature(body, c.Request.Header) {
		log.Printf("[Webhook] Invalid signature")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
		return
	}

	// Parser l'événement
	var event WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("[Webhook] Failed to parse event: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event format"})
		return
	}

	log.Printf("[Webhook] Received event: %s", event.Type)

	// Traiter l'événement selon son type
	switch event.Type {
	case "subscription.updated":
		wr.handleSubscriptionUpdated(event.Data)
	case "subscription.canceled":
		wr.handleSubscriptionCanceled(event.Data)
	case "payment.succeeded":
		wr.handlePaymentSucceeded(event.Data)
	case "payment.failed":
		wr.handlePaymentFailed(event.Data)
	case "quota.exceeded":
		wr.handleQuotaExceeded(event.Data)
	default:
		log.Printf("[Webhook] Unknown event type: %s", event.Type)
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}

// verifySignature vérifie la signature HMAC du webhook
func (wr *WebhookReceiver) verifySignature(body []byte, headers http.Header) bool {
	if wr.secret == "" {
		log.Printf("[Webhook] Warning: No secret configured, skipping verification")
		return true
	}

	timestamp := headers.Get("X-Billing-Timestamp")
	signature := headers.Get("X-Billing-Signature")

	if timestamp == "" || signature == "" {
		return false
	}

	// Vérifier que le timestamp n'est pas trop vieux (5 minutes)
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}
	if time.Now().Unix()-ts > 300 {
		log.Printf("[Webhook] Timestamp too old")
		return false
	}

	// Calculer la signature attendue
	message := fmt.Sprintf("%s.%s", timestamp, string(body))
	mac := hmac.New(sha256.New, []byte(wr.secret))
	mac.Write([]byte(message))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSig))
}

// === Event Handlers ===

func (wr *WebhookReceiver) handleSubscriptionUpdated(data json.RawMessage) {
	var d SubscriptionUpdatedData
	if err := json.Unmarshal(data, &d); err != nil {
		log.Printf("[Webhook] Failed to parse subscription.updated: %v", err)
		return
	}

	log.Printf("[Webhook] Subscription updated: user=%s plan=%s status=%s",
		d.UserID, d.PlanCode, d.Status)

	// TODO: Mettre à jour les quotas de l'utilisateur dans la base locale
	// Exemple: pkg.DB.NewUpdate().Model(&models.User{}).
	//   Where("id = ?", d.UserID).
	//   Set("plan_code = ?", d.PlanCode).
	//   Exec(context.Background())
}

func (wr *WebhookReceiver) handleSubscriptionCanceled(data json.RawMessage) {
	var d SubscriptionUpdatedData
	if err := json.Unmarshal(data, &d); err != nil {
		log.Printf("[Webhook] Failed to parse subscription.canceled: %v", err)
		return
	}

	log.Printf("[Webhook] Subscription canceled: user=%s", d.UserID)

	// TODO: Rétrograder l'utilisateur vers le plan gratuit
	// et éventuellement gérer les données excédentaires
}

func (wr *WebhookReceiver) handlePaymentSucceeded(data json.RawMessage) {
	var d PaymentData
	if err := json.Unmarshal(data, &d); err != nil {
		log.Printf("[Webhook] Failed to parse payment.succeeded: %v", err)
		return
	}

	log.Printf("[Webhook] Payment succeeded: user=%s invoice=%s amount=%.2f%s",
		d.UserID, d.InvoiceID, d.Amount/100, d.Currency)

	// TODO: Envoyer un email de confirmation de paiement
}

func (wr *WebhookReceiver) handlePaymentFailed(data json.RawMessage) {
	var d PaymentData
	if err := json.Unmarshal(data, &d); err != nil {
		log.Printf("[Webhook] Failed to parse payment.failed: %v", err)
		return
	}

	log.Printf("[Webhook] Payment failed: user=%s invoice=%s", d.UserID, d.InvoiceID)

	// TODO: Envoyer un email de relance de paiement
	// TODO: Après plusieurs échecs, suspendre le compte
}

func (wr *WebhookReceiver) handleQuotaExceeded(data json.RawMessage) {
	var d QuotaExceededData
	if err := json.Unmarshal(data, &d); err != nil {
		log.Printf("[Webhook] Failed to parse quota.exceeded: %v", err)
		return
	}

	log.Printf("[Webhook] Quota exceeded: user=%s type=%s usage=%d limit=%d",
		d.UserID, d.QuotaType, d.CurrentUsage, d.Limit)

	// TODO: Envoyer une notification à l'utilisateur
	// TODO: Bloquer les uploads si quota de stockage dépassé
}

// RegisterWebhookRoute enregistre la route webhook
func RegisterWebhookRoute(router *gin.RouterGroup) {
	receiver := NewWebhookReceiver()
	router.POST("/billing/webhooks", receiver.HandleWebhook)
}
