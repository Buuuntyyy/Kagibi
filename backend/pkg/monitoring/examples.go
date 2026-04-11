// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package monitoring

/*
EXEMPLE D'UTILISATION DES MÉTRIQUES PROMETHEUS DANS VOS HANDLERS

Ce fichier contient des exemples d'utilisation des métriques personnalisées
dans différentes parties de votre application Kagibi.
*/

import (
	"time"
)

// Exemple 1: Enregistrer un upload de fichier
func ExampleRecordFileUpload() {
	// Après un upload réussi
	RecordFileUpload()

	// Les métriques suivantes seront automatiquement incrémentées :
	// - kagibi_file_uploads_total +1
}

// Exemple 2: Mesurer la latence d'une opération de chiffrement
func ExampleRecordEncryption() {
	start := time.Now()

	// Votre code de chiffrement ici
	// encryptedData := encryptFile(fileData)

	duration := time.Since(start)
	RecordEncryption(duration)

	// La métrique kagibi_encryption_duration_seconds observera la durée
}

// Exemple 3: Enregistrer une erreur d'authentification
func ExampleRecordAuthError() {
	// Token expiré
	RecordAuthError("expired_token")

	// Token invalide
	RecordAuthError("invalid_token")

	// Token manquant
	RecordAuthError("missing_token")

	// Credentials invalides
	RecordAuthError("invalid_credentials")

	// La métrique kagibi_auth_errors_total{type="..."} sera incrémentée
}

// Exemple 4: Enregistrer une vérification MFA
func ExampleRecordMFAVerification() {
	// Après vérification du code TOTP
	isValid := true // ou false si le code est invalide
	RecordMFAVerification(isValid)

	// La métrique kagibi_mfa_verifications_total{status="success"|"failure"} sera incrémentée
}

// Exemple 5: Enregistrer une requête S3
func ExampleRecordS3Request() {
	// Après un upload vers S3
	success := true
	RecordS3Request("put", success)

	// Après un download depuis S3
	RecordS3Request("get", true)

	// Après une suppression S3
	RecordS3Request("delete", false) // en cas d'erreur

	// La métrique kagibi_s3_requests_total{operation="...",status="..."} sera incrémentée
}

// Exemple 6: Utilisation complète dans un handler de fichier
func ExampleCompleteFileUploadHandler() {
	// 1. Incrémenter les connexions actives (déjà fait par le middleware)
	// IncrementActiveConnections() - automatique
	// defer DecrementActiveConnections() - automatique

	// 2. Mesurer le temps de chiffrement
	encryptStart := time.Now()
	// encryptedData := encryptFile(fileData)
	RecordEncryption(time.Since(encryptStart))

	// 3. Enregistrer l'upload du fichier
	RecordFileUpload()

	// 4. Enregistrer l'upload vers S3
	s3Success := true
	RecordS3Request("put", s3Success)

	// Les métriques HTTP sont automatiquement enregistrées par le middleware
	// RequestsTotal et RequestDuration seront mis à jour automatiquement
}

// Exemple 7: Handler avec gestion d'erreur
func ExampleHandlerWithErrorTracking() {
	// Vérifier l'authentification
	tokenValid := false
	if !tokenValid {
		RecordAuthError("invalid_token")
		// return unauthorized error
		return
	}

	// Vérifier le MFA si requis
	mfaValid := false // validateMFA(mfaCode)
	RecordMFAVerification(mfaValid)

	if !mfaValid {
		// return mfa error
		return
	}

	// Télécharger le fichier
	RecordFileDownload()

	// Mesurer le temps de déchiffrement
	decryptStart := time.Now()
	// decryptedData := decryptFile(encryptedData)
	RecordDecryption(time.Since(decryptStart))

	// Enregistrer la requête S3
	RecordS3Request("get", true)
}

/*
DASHBOARD GRAFANA RECOMMANDÉ

Voici les queries Prometheus utiles pour votre dashboard:

1. Taux de requêtes par endpoint:
   rate(kagibi_http_requests_total[5m])

2. Latence moyenne par endpoint (P95):
   histogram_quantile(0.95, rate(kagibi_http_request_duration_seconds_bucket[5m]))

3. Taux d'erreurs HTTP:
   sum(rate(kagibi_http_requests_total{status=~"5.."}[5m])) by (endpoint)

4. Uploads de fichiers par minute:
   rate(kagibi_file_uploads_total[1m]) * 60

5. Taille moyenne des uploads:
   rate(kagibi_file_upload_size_bytes_sum[5m]) / rate(kagibi_file_upload_size_bytes_count[5m])

6. Taux d'erreurs d'authentification:
   rate(kagibi_auth_errors_total[5m])

7. Taux de succès MFA:
   rate(kagibi_mfa_verifications_total{status="success"}[5m]) / rate(kagibi_mfa_verifications_total[5m])

8. Connexions actives:
   kagibi_active_connections

9. Utilisation mémoire Go:
   go_memstats_alloc_bytes

10. Nombre de goroutines:
    go_goroutines

CONFIGURATION PROMETHEUS (prometheus.yml):

scrape_configs:
  - job_name: 'kagibi'
    static_configs:
      - targets: ['localhost:9090']
    scrape_interval: 15s
    scrape_timeout: 10s

ALERTES RECOMMANDÉES:

groups:
  - name: kagibi_alerts
    rules:
      - alert: HighErrorRate
        expr: sum(rate(kagibi_http_requests_total{status=~"5.."}[5m])) > 0.05
        for: 5m
        annotations:
          summary: "Taux d'erreur HTTP élevé"

      - alert: HighMFAFailureRate
        expr: rate(kagibi_mfa_verifications_total{status="failure"}[5m]) / rate(kagibi_mfa_verifications_total[5m]) > 0.5
        for: 2m
        annotations:
          summary: "Taux d'échec MFA élevé"

      - alert: HighAuthErrorRate
        expr: rate(kagibi_auth_errors_total[5m]) > 10
        for: 5m
        annotations:
          summary: "Taux d'erreur d'authentification élevé"
*/
