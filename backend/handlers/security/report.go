package security

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// SecurityEvent représente un événement de sécurité signalé par le client
type SecurityEvent struct {
	Timestamp  string                 `json:"timestamp"`
	Type       string                 `json:"type"`
	Severity   string                 `json:"severity"`
	Details    map[string]interface{} `json:"details"`
	UserAgent  string                 `json:"userAgent"`
	UserID     string                 `json:"userId"`
	IP         string                 `json:"ip"`
	ReceivedAt time.Time              `json:"receivedAt"`
}

// ReportSecurityEvent traite un rapport d'événement de sécurité du client
func ReportSecurityEvent(c *gin.Context) {
	var event SecurityEvent

	// Parser le JSON
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Ajouter les métadonnées
	event.IP = c.ClientIP()
	event.ReceivedAt = time.Now()

	// Récupérer l'utilisateur depuis le contexte (si authentifié)
	if userID, exists := c.Get("userID"); exists {
		event.UserID = userID.(string)
	}

	// Logger l'événement
	logSecurityEvent(event)

	// Envoyer une alerte si criticité élevée
	if event.Severity == "critical" {
		sendSecurityAlert(event)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Event reported successfully",
	})
}

// logSecurityEvent log l'événement de sécurité
func logSecurityEvent(event SecurityEvent) {
	// Logger l'événement de sécurité
	message := fmt.Sprintf(
		"[SECURITY] %s - Severity: %s, UserID: %s, IP: %s, Details: %v",
		event.Type,
		event.Severity,
		event.UserID,
		event.IP,
		event.Details,
	)
	log.Println(message)
}

// sendSecurityAlert envoie une alerte en cas d'événement critique
func sendSecurityAlert(event SecurityEvent) {
	// TODO: Implémenter l'envoi d'alerte (email, webhook, etc.)
	// Pour l'instant, juste logger
	message := fmt.Sprintf(
		"[CRITICAL_SECURITY_EVENT] Type: %s, UserID: %s, IP: %s, Details: %v",
		event.Type,
		event.UserID,
		event.IP,
		event.Details,
	)
	log.Println(message)
}

// GetSecurityEvents récupère les événements de sécurité de l'utilisateur
func GetSecurityEvents(c *gin.Context) {
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	// TODO: Implémenter la récupération depuis la base de données
	// Pour l'instant, retourner une liste vide
	c.JSON(http.StatusOK, gin.H{
		"events": []SecurityEvent{},
		"total":  0,
	})
}

// LogSecurityEvent enregistre un événement de sécurité
func LogSecurityEvent(eventType, severity, userID, ip string, details map[string]interface{}) {
	message := "[SECURITY] " + eventType + " - Severity: " + severity

	if userID != "" {
		message += ", UserID: " + userID
	}

	if ip != "" {
		message += ", IP: " + ip
	}

	// Utiliser le logger existant
	// TODO: Envoyer à security_logger.go
	println(message)
}
