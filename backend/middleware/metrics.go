package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"kagibi/backend/pkg/monitoring"
)

// MetricsMiddleware est un middleware Gin qui enregistre automatiquement
// les métriques pour chaque requête HTTP
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Incrémenter les connexions actives
		monitoring.IncrementActiveConnections()
		defer monitoring.DecrementActiveConnections()

		// Enregistrer le temps de début
		start := time.Now()

		// Traiter la requête
		c.Next()

		// Calculer la durée
		duration := time.Since(start)

		// Enregistrer les métriques
		monitoring.RecordRequestMetrics(
			c.Request.Method,
			c.FullPath(), // Utilise le chemin de route (avec paramètres) au lieu de l'URL complète
			c.Writer.Status(),
			duration,
		)
	}
}
