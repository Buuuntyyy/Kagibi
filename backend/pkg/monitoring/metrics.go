package monitoring

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Métriques globales de l'application
var (
	// Counter: Nombre total de requêtes HTTP traitées
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kagibi_http_requests_total",
			Help: "Nombre total de requêtes HTTP traitées par l'application",
		},
		[]string{"method", "endpoint", "status"},
	)

	// Histogram: Latence des requêtes HTTP
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kagibi_http_request_duration_seconds",
			Help:    "Latence des requêtes HTTP en secondes",
			Buckets: prometheus.DefBuckets, // 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10
		},
		[]string{"method", "endpoint"},
	)

	// Counter: Nombre d'uploads de fichiers
	FileUploadsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "kagibi_file_uploads_total",
			Help: "Nombre total d'uploads de fichiers",
		},
	)

	// Counter: Nombre de téléchargements de fichiers
	FileDownloadsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "kagibi_file_downloads_total",
			Help: "Nombre total de téléchargements de fichiers",
		},
	)

	// Gauge: Nombre de connexions actives
	ActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "kagibi_active_connections",
			Help: "Nombre de connexions HTTP actives",
		},
	)

	// Histogram: Taille des fichiers uploadés
	FileUploadSize = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "kagibi_file_upload_size_bytes",
			Help:    "Taille des fichiers uploadés en octets",
			Buckets: []float64{1024, 10240, 102400, 1048576, 10485760, 104857600, 1073741824}, // 1KB, 10KB, 100KB, 1MB, 10MB, 100MB, 1GB
		},
	)

	// Counter: Erreurs d'authentification
	AuthErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kagibi_auth_errors_total",
			Help: "Nombre total d'erreurs d'authentification",
		},
		[]string{"type"}, // "invalid_token", "expired_token", "missing_token", etc.
	)

	// Counter: Vérifications MFA
	MFAVerificationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kagibi_mfa_verifications_total",
			Help: "Nombre total de vérifications MFA",
		},
		[]string{"status"}, // "success", "failure"
	)

	// Histogram: Durée des opérations de chiffrement
	EncryptionDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "kagibi_encryption_duration_seconds",
			Help:    "Durée des opérations de chiffrement en secondes",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
		},
	)

	// Histogram: Durée des opérations de déchiffrement
	DecryptionDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "kagibi_decryption_duration_seconds",
			Help:    "Durée des opérations de déchiffrement en secondes",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
		},
	)

	// Gauge: Nombre total d'utilisateurs actifs (identifiés par requêtes récentes)
	ActiveUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "kagibi_active_users",
			Help: "Nombre d'utilisateurs uniques actifs dans les 5 dernières minutes",
		},
	)

	// Counter: Requêtes vers S3
	S3RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kagibi_s3_requests_total",
			Help: "Nombre total de requêtes vers S3",
		},
		[]string{"operation", "status"}, // operation: "put", "get", "delete", status: "success", "error"
	)
)

// RecordRequestMetrics enregistre les métriques pour une requête HTTP
func RecordRequestMetrics(method, endpoint string, statusCode int, duration time.Duration) {
	RequestsTotal.WithLabelValues(method, endpoint, string(rune(statusCode))).Inc()
	RequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// RecordEncryption enregistre le temps d'une opération de chiffrement
func RecordEncryption(duration time.Duration) {
	EncryptionDuration.Observe(duration.Seconds())
}

// RecordDecryption enregistre le temps d'une opération de déchiffrement
func RecordDecryption(duration time.Duration) {
	DecryptionDuration.Observe(duration.Seconds())
}

// RecordFileUpload enregistre un upload de fichier
func RecordFileUpload(sizeBytes int64) {
	FileUploadsTotal.Inc()
	FileUploadSize.Observe(float64(sizeBytes))
}

// RecordFileDownload enregistre un téléchargement de fichier
func RecordFileDownload() {
	FileDownloadsTotal.Inc()
}

// RecordAuthError enregistre une erreur d'authentification
func RecordAuthError(errorType string) {
	AuthErrorsTotal.WithLabelValues(errorType).Inc()
}

// RecordMFAVerification enregistre une tentative de vérification MFA
func RecordMFAVerification(success bool) {
	status := "failure"
	if success {
		status = "success"
	}
	MFAVerificationsTotal.WithLabelValues(status).Inc()
}

// RecordS3Request enregistre une requête S3
func RecordS3Request(operation string, success bool) {
	status := "error"
	if success {
		status = "success"
	}
	S3RequestsTotal.WithLabelValues(operation, status).Inc()
}

// IncrementActiveConnections incrémente le nombre de connexions actives
func IncrementActiveConnections() {
	ActiveConnections.Inc()
}

// DecrementActiveConnections décrémente le nombre de connexions actives
func DecrementActiveConnections() {
	ActiveConnections.Dec()
}
