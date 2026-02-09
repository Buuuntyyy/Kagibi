# Système de Monitoring Prometheus - SaferCloud

## Vue d'ensemble

Ce système de monitoring utilise **Prometheus** et expose des métriques détaillées sur:
- Performance HTTP (latence, taux de requêtes)
- Uploads/Downloads de fichiers
- Authentification et MFA
- Opérations de chiffrement/déchiffrement
- Requêtes S3
- Métriques runtime Go (mémoire, GC, goroutines)

## Installation

### 1. Installer Prometheus (macOS/Linux)

```bash
# macOS avec Homebrew
brew install prometheus

# Linux (Ubuntu/Debian)
sudo apt-get install prometheus

# Ou télécharger depuis https://prometheus.io/download/
```

### 2. Installer les dépendances Go

```bash
cd backend
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promhttp
go get github.com/prometheus/client_golang/prometheus/promauto
```

### 3. Configuration Prometheus

Créer un fichier `prometheus.yml`:

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'safercloud'
    static_configs:
      - targets: ['localhost:9090']
    scrape_interval: 10s
    scrape_timeout: 5s

  - job_name: 'safercloud-api'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 10s
```

### 4. Démarrer Prometheus

```bash
prometheus --config.file=prometheus.yml
```

Accéder à l'interface: http://localhost:9090

## Endpoints disponibles

Le serveur de métriques écoute sur le **port 9090** (par défaut):

- `http://localhost:9090/metrics` - Métriques Prometheus
- `http://localhost:9090/health` - Health check
- `http://localhost:9090/ready` - Readiness check

## Métriques disponibles

### Métriques HTTP

| Métrique | Type | Description |
|----------|------|-------------|
| `safercloud_http_requests_total` | Counter | Nombre total de requêtes HTTP |
| `safercloud_http_request_duration_seconds` | Histogram | Latence des requêtes HTTP |
| `safercloud_active_connections` | Gauge | Connexions HTTP actives |

Labels: `method`, `endpoint`, `status`

### Métriques Fichiers

| Métrique | Type | Description |
|----------|------|-------------|
| `safercloud_file_uploads_total` | Counter | Nombre d'uploads |
| `safercloud_file_downloads_total` | Counter | Nombre de téléchargements |
| `safercloud_file_upload_size_bytes` | Histogram | Taille des fichiers uploadés |

### Métriques Authentification

| Métrique | Type | Description |
|----------|------|-------------|
| `safercloud_auth_errors_total` | Counter | Erreurs d'authentification |
| `safercloud_mfa_verifications_total` | Counter | Vérifications MFA |

Labels: `type` (auth), `status` (MFA)

### Métriques Chiffrement

| Métrique | Type | Description |
|----------|------|-------------|
| `safercloud_encryption_duration_seconds` | Histogram | Durée du chiffrement |
| `safercloud_decryption_duration_seconds` | Histogram | Durée du déchiffrement |

### Métriques S3

| Métrique | Type | Description |
|----------|------|-------------|
| `safercloud_s3_requests_total` | Counter | Requêtes vers S3 |

Labels: `operation` (put/get/delete), `status` (success/error)

### Métriques Runtime Go (automatiques)

- `go_goroutines` - Nombre de goroutines
- `go_memstats_alloc_bytes` - Mémoire allouée
- `go_gc_duration_seconds` - Durée du Garbage Collector
- `process_cpu_seconds_total` - Temps CPU
- `process_resident_memory_bytes` - Mémoire RSS

## Utilisation dans le code

### Instrumentation automatique (middleware)

Les requêtes HTTP sont automatiquement instrumentées via le middleware:

```go
// Déjà configuré dans main.go
router.Use(middleware.MetricsMiddleware())
```

### Instrumentation manuelle

```go
import "safercloud/backend/pkg/monitoring"

// Upload de fichier
func HandleFileUpload(c *gin.Context) {
    // ... votre logique d'upload ...

    monitoring.RecordFileUpload(fileSize)
}

// Chiffrement
func EncryptFile(data []byte) []byte {
    start := time.Now()

    encrypted := performEncryption(data)

    monitoring.RecordEncryption(time.Since(start))
    return encrypted
}

// Erreur d'authentification
func ValidateToken(token string) error {
    if !isValid {
        monitoring.RecordAuthError("invalid_token")
        return errors.New("invalid token")
    }
    return nil
}

// Vérification MFA
func VerifyMFA(code string) bool {
    valid := checkTOTP(code)
    monitoring.RecordMFAVerification(valid)
    return valid
}

// Requête S3
func UploadToS3(key string, data []byte) error {
    err := s3Client.PutObject(...)
    monitoring.RecordS3Request("put", err == nil)
    return err
}
```

## Queries Prometheus utiles

### Performance globale

```promql
# Taux de requêtes par seconde
rate(safercloud_http_requests_total[5m])

# Latence P95
histogram_quantile(0.95, rate(safercloud_http_request_duration_seconds_bucket[5m]))

# Taux d'erreurs HTTP 5xx
sum(rate(safercloud_http_requests_total{status=~"5.."}[5m]))
```

### Fichiers

```promql
# Uploads par minute
rate(safercloud_file_uploads_total[1m]) * 60

# Taille moyenne des uploads
rate(safercloud_file_upload_size_bytes_sum[5m]) / rate(safercloud_file_upload_size_bytes_count[5m])

# Ratio upload/download
rate(safercloud_file_uploads_total[5m]) / rate(safercloud_file_downloads_total[5m])
```

### Authentification & Sécurité

```promql
# Taux d'erreurs d'auth
rate(safercloud_auth_errors_total[5m])

# Taux de succès MFA
rate(safercloud_mfa_verifications_total{status="success"}[5m]) / rate(safercloud_mfa_verifications_total[5m])

# Top erreurs d'auth
topk(5, sum by (type) (rate(safercloud_auth_errors_total[5m])))
```

### Performance système

```promql
# Utilisation mémoire
go_memstats_alloc_bytes / 1024 / 1024

# Nombre de goroutines
go_goroutines

# Connexions actives
safercloud_active_connections
```

## Configuration Grafana

### Importer le datasource Prometheus

1. Aller dans Configuration > Data Sources
2. Ajouter Prometheus
3. URL: `http://localhost:9090`
4. Cliquer sur "Save & Test"

### Dashboards recommandés

**Dashboard: SaferCloud Overview**

- Taux de requêtes (par endpoint)
- Latence P50, P95, P99
- Taux d'erreurs
- Connexions actives
- Uploads/Downloads

**Dashboard: Performance**

- Latence de chiffrement/déchiffrement
- Performance S3
- Utilisation CPU/Mémoire
- Goroutines

**Dashboard: Sécurité**

- Erreurs d'authentification
- Succès/échecs MFA
- Tentatives suspectes

## Alerting

Exemple de règles d'alerte (`alerts.yml`):

```yaml
groups:
  - name: safercloud_critical
    rules:
      - alert: HighErrorRate
        expr: sum(rate(safercloud_http_requests_total{status=~"5.."}[5m])) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Taux d'erreur HTTP élevé (> 5%)"
          description: "{{ $value }} erreurs/sec"

      - alert: HighLatency
        expr: histogram_quantile(0.95, rate(safercloud_http_request_duration_seconds_bucket[5m])) > 2
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Latence P95 > 2s"

      - alert: MFAFailureSpike
        expr: rate(safercloud_mfa_verifications_total{status="failure"}[5m]) > 10
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "Pic d'échecs MFA détecté"

      - alert: AuthErrorSpike
        expr: rate(safercloud_auth_errors_total[5m]) > 20
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Pic d'erreurs d'authentification"

      - alert: HighMemoryUsage
        expr: go_memstats_alloc_bytes > 1e9
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Utilisation mémoire > 1GB"
```

## Docker Compose (Prometheus + Grafana)

```yaml
version: '3.8'

services:
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9091:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus-data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3001:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana-data:/var/lib/grafana
    depends_on:
      - prometheus

volumes:
  prometheus-data:
  grafana-data:
```

Démarrer:
```bash
docker-compose up -d
```

## Configuration avancée

### Changer le port du serveur de métriques

Dans `main.go`:
```go
metricsServer := monitoring.NewServer(9090) // Changer le port ici
```

### Désactiver le monitoring

Commenter ces lignes dans `main.go`:
```go
// metricsServer := monitoring.NewServer(9090)
// if err := metricsServer.Start(); err != nil {
//     log.Printf("Warning: Failed to start metrics server: %v", err)
// }
```

Et retirer le middleware:
```go
// router.Use(middleware.MetricsMiddleware())
```

## Ressources

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Prometheus Client Go](https://github.com/prometheus/client_golang)
- [Grafana Dashboards](https://grafana.com/grafana/dashboards/)
- [PromQL Tutorial](https://prometheus.io/docs/prometheus/latest/querying/basics/)

## Troubleshooting

**Problème: Métriques non visibles**
- Vérifier que le port 9090 n'est pas bloqué
- Vérifier les logs du serveur: `Serveur de métriques Prometheus démarré sur le port 9090`
- Tester manuellement: `curl http://localhost:9090/metrics`

**Problème: Prometheus ne scrape pas**
- Vérifier le fichier `prometheus.yml`
- Vérifier le statut dans Prometheus UI: Status > Targets
- Vérifier les logs Prometheus

**Problème: Shutdown ne fonctionne pas**
- Le serveur utilise SIGINT (Ctrl+C) et SIGTERM
- Timeout de 30 secondes configuré
- Vérifier les logs: `Arrêt gracieux en cours...`

## Bonnes pratiques

1. **Ne jamais bloquer** le middleware avec des opérations longues
2. **Limiter les cardinalités** des labels (éviter user_id, request_id, etc.)
3. **Utiliser des buckets appropriés** pour les histogrammes
4. **Monitorer la croissance** des métriques (éviter l'explosion)
5. **Tester les queries** avant de les mettre en production
6. **Configurer des alertes** sur les métriques critiques
7. **Documenter** les métriques personnalisées

## Résultat

Avec ce système, vous avez maintenant:
- Monitoring temps réel de votre application
- Métriques détaillées sur la performance
- Détection proactive des problèmes
- Capacité d'analyse post-mortem
- Base pour l'auto-scaling
- Conformité production-ready

---

**Auteur**: Expert Backend Go
**Version**: 1.0
**Date**: 2026
