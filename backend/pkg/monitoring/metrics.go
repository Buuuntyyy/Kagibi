// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package monitoring

import (
	"strconv"
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

	// --- Métriques métier ---

	// Counter: Inscriptions utilisateurs
	UserRegistrationsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "kagibi_user_registrations_total",
			Help: "Nombre total de comptes créés avec succès",
		},
	)

	// Counter: Suppressions de compte
	UserDeletionsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "kagibi_user_deletions_total",
			Help: "Nombre total de comptes supprimés",
		},
	)

	// Counter: Connexions réussies
	UserLoginsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kagibi_user_logins_total",
			Help: "Nombre total de connexions",
		},
		[]string{"status"}, // "success", "failure"
	)

	// Counter: Transferts P2P initiés
	P2PTransfersTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "kagibi_p2p_transfers_total",
			Help: "Nombre total de transferts P2P initiés",
		},
	)

	// Gauge: Nombre total d'utilisateurs inscrits (mis à jour périodiquement)
	TotalUsersGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "kagibi_users_total",
			Help: "Nombre total de profils utilisateurs en base",
		},
	)

	// Gauge: Stockage total utilisé (octets)
	TotalStorageUsedBytes = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "kagibi_storage_used_bytes_total",
			Help: "Somme du stockage utilisé par tous les utilisateurs en octets",
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

	// Histogram: Latence des opérations S3
	S3OperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kagibi_s3_operation_duration_seconds",
			Help:    "Latence des opérations S3 en secondes",
			Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0},
		},
		[]string{"operation"}, // "put", "get", "delete"
	)

	// Counter: Hits du rate limiter
	RateLimitHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kagibi_rate_limit_hits_total",
			Help: "Nombre de requêtes bloquées par le rate limiter",
		},
		[]string{"endpoint"},
	)

	// Counter: Tentatives d'accès aux liens de partage
	ShareAccessAttemptsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kagibi_share_access_attempts_total",
			Help: "Tentatives d'accès aux liens de partage",
		},
		[]string{"result"}, // "success", "not_found", "expired", "forbidden"
	)

	// Counter: Liens de partage créés
	SharesCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "kagibi_shares_created_total",
			Help: "Nombre total de liens de partage créés",
		},
	)

	// Counter: Fichiers supprimés
	FilesDeletedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "kagibi_files_deleted_total",
			Help: "Nombre total de fichiers supprimés",
		},
	)

	// Gauge: Connexions WebSocket actives
	WSConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "kagibi_ws_connections_active",
			Help: "Nombre de connexions WebSocket actives",
		},
	)
)

// RecordRequestMetrics enregistre les métriques pour une requête HTTP
func RecordRequestMetrics(method, endpoint string, statusCode int, duration time.Duration) {
	RequestsTotal.WithLabelValues(method, endpoint, strconv.Itoa(statusCode)).Inc()
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
func RecordFileUpload() {
	FileUploadsTotal.Inc()
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

// RecordS3Duration enregistre la latence d'une opération S3
func RecordS3Duration(operation string, duration time.Duration) {
	S3OperationDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// RecordRateLimitHit enregistre un blocage du rate limiter
func RecordRateLimitHit(endpoint string) {
	RateLimitHitsTotal.WithLabelValues(endpoint).Inc()
}

// RecordShareAccess enregistre une tentative d'accès à un partage
func RecordShareAccess(result string) {
	ShareAccessAttemptsTotal.WithLabelValues(result).Inc()
}

// RecordShareCreated enregistre la création d'un lien de partage
func RecordShareCreated() {
	SharesCreatedTotal.Inc()
}

// RecordFileDeleted enregistre la suppression d'un fichier
func RecordFileDeleted() {
	FilesDeletedTotal.Inc()
}

// RecordUserRegistration enregistre une nouvelle inscription
func RecordUserRegistration() {
	UserRegistrationsTotal.Inc()
}

// RecordUserDeletion enregistre une suppression de compte
func RecordUserDeletion() {
	UserDeletionsTotal.Inc()
}

// RecordUserLogin enregistre une tentative de connexion
func RecordUserLogin(success bool) {
	status := "failure"
	if success {
		status = "success"
	}
	UserLoginsTotal.WithLabelValues(status).Inc()
}

// IncrementWSConnections incrémente le nombre de connexions WebSocket actives
func IncrementWSConnections() {
	WSConnectionsActive.Inc()
}

// DecrementWSConnections décrémente le nombre de connexions WebSocket actives
func DecrementWSConnections() {
	WSConnectionsActive.Dec()
}

// IncrementActiveConnections incrémente le nombre de connexions actives
func IncrementActiveConnections() {
	ActiveConnections.Inc()
}

// DecrementActiveConnections décrémente le nombre de connexions actives
func DecrementActiveConnections() {
	ActiveConnections.Dec()
}
