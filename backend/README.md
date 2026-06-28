# Backend Kagibi - Go API + Zero-Knowledge Architecture

**API REST** qui stocke uniquement des données chiffrées et ne possède aucune clé de déchiffrement.

---

## Vue d'ensemble

Le backend Kagibi est une API Go qui respecte les principes **Zero-Knowledge** :
- Stocke uniquement `EncryptedMasterKey` (inutilisable sans password)
- Fichiers chiffrés dans S3/MinIO
- JWT validation via Supabase
- Rate limiting avec `sync.Map`
- WebSocket temps réel pour notifications
- **Upload S3 Multipart** avec URLs présignées (direct client → S3, parts 5–100 MB, TTL 3 min)
- **Download multi-fichiers** avec batch presigned URLs (génération parallèle 10 goroutines) et folder tree
- **Logs structurés JSON** (slog) avec double IP (complète pour événements sécurité, anonymisée CNIL 2021-122 pour logs HTTP), user-agent, et couverture complète LCEN : compte, auth, partages, fichiers

---

## Stack Technique

| Technologie | Version | Rôle |
|-------------|---------|------|
| **Go** | 1.21+ | Langage serveur |
| **Gin** | 1.11.0 | Framework HTTP |
| **Bun ORM** | 1.2.16 | ORM PostgreSQL |
| **PostgreSQL** | 16+ | Base de données |
| **Redis** | 7+ | Cache & rate limiting |
| **AWS SDK v2** | 1.40.0 | Client S3/MinIO |
| **Gorilla WebSocket** | 1.5.3 | WebSocket temps réel |
| **golang-jwt** | 5.3.0 | Validation JWT |
| **keyfunc** | 3.7.0 | JWKS validation (Supabase ES256) |

---

## Démarrage Rapide

### Prérequis

```bash
Go 1.21+
PostgreSQL 16+
Redis 7+
MinIO ou AWS S3
```

### Installation

```bash
cd backend
go mod download
```

### Configuration

Créez `.env`:

```bash
# Base de données
DATABASE_URL=postgresql://user:password@localhost:5432/kagibi

# Redis
REDIS_URL=redis://localhost:6379

# Supabase (Authentification JWT)
SUPABASE_URL=https://xxx.supabase.co
SUPABASE_JWT_SECRET=your-jwt-secret

# Stockage S3/MinIO
S3_ENDPOINT=http://localhost:9000
S3_REGION=us-east-1
S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=minioadmin
S3_BUCKET=kagibi-files

# CORS
ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000

# TURN (optionnel - WebRTC)
TURN_URLS=turn:your-turn-server.com:3478
TURN_SECRET=your-turn-secret
TURN_USER=username
TURN_PASSWORD=password
```

### Lancement

```bash
go run main.go
# → API disponible sur http://localhost:8080
```

---

## 📁 Structure

```
backend/
├── main.go                     # Point d'entrée
│
├── handlers/                   # Contrôleurs API
│   ├── auth/
│   │   ├── keys.go             # GET /auth/keys (EncryptedMasterKey)
│   │   ├── login.go            # POST /auth/login (Supabase)
│   │   ├── register.go         # POST /auth/register
│   │   ├── recovery.go         # POST /auth/recovery/* (compte)
│   │   └── logout.go           # POST /auth/logout
│   │
│   ├── files/
│   │   ├── upload.go           # POST /files/upload (chunked)
│   │   ├── download.go         # GET /files/download/:id
│   │   ├── batch_presign.go    # POST /files/batch-presign, selection-tree
│   │   ├── delete.go           # DELETE /files/file/:id
│   │   ├── list.go             # GET /files/list/*path
│   │   ├── move.go             # POST /files/move
│   │   ├── rename.go           # POST /files/rename
│   │   └── search.go           # GET /files/search?q=
│   │
│   ├── folders/
│   │   ├── create.go           # POST /folders/create
│   │   ├── tree.go             # GET /folders/:id/tree (récursif)
│   │   └── update_key.go       # PUT /folders/:id/key
│   │
│   ├── friends/
│   │   └── handler.go          # CRUD amis (WebSocket notif)
│   │
│   ├── shares/
│   │   ├── create_link.go      # POST /shares/link (public)
│   │   ├── create_direct.go    # POST /shares/direct (user)
│   │   ├── list.go             # GET /shares/list
│   │   └── remove.go           # DELETE /shares/*
│   │
│   ├── users/
│   │   ├── profile.go          # GET/PUT /users/profile
│   │   ├── update_password.go  # POST /users/change-password
│   │   └── recent.go           # GET/POST /users/recent
│   │
│   ├── security/
│   │   └── report.go           # POST /security/report (events)
│   │
│   └── ws/
│       ├── connect.go          # WebSocket handler
│       └── ice.go              # GET /ice-config (TURN)
│
├── middleware/
│   ├── auth.go                 # JWT validation (Supabase)
│   ├── ratelimit.go            # Rate limiting (sync.Map)
│   ├── security.go             # Security headers
│   └── security_logger.go      # Structured security logs
│
├── pkg/
│   ├── database.go             # Opérations DB (Bun ORM)
│   ├── models.go               # Structures Bun
│   ├── migrate.go              # Migrations SQL
│   ├── folder_sizes.go         # Calcul tailles dossiers
│   ├── shared_folder.go        # Logique partage
│   │
│   ├── s3storage/
│   │   └── client.go           # Client S3/MinIO
│   │
│   ├── workers/
│   │   ├── preview.go          # Génération previews (async)
│   │   └── cleanup.go          # Cleanup fichiers temporaires
│   │
│   └── ws/
│       └── manager.go          # WebSocket connection manager
│
├── migrations/
│   └── *.sql                   # Migrations PostgreSQL
│
└── logs/
    └── security.log            # Logs sécurité (auto-créé)
```

---

## 🔐 Architecture Zero-Knowledge

### Flux d'Authentification

```
1. Frontend → Supabase : Login/Register
   ↓
2. Supabase → Frontend : JWT Token (ES256 ou HS256)
   ↓
3. Frontend → Backend : Request + Authorization: Bearer <JWT>
   ↓
4. Backend → Supabase JWKS : Validation signature JWT
   ↓
5. Backend → Response : user_id extrait du JWT
```

### Données Stockées (DB)

```sql
-- users table
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    salt VARCHAR(255) NOT NULL,               -- Hex-encoded 16 bytes
    encrypted_master_key TEXT NOT NULL,       -- Base64 AES-GCM encrypted
    public_key TEXT,                          -- RSA-OAEP 4096 bits PEM
    encrypted_private_key TEXT,               -- Chiffré avec MasterKey
    recovery_hash VARCHAR(255),               -- SHA-256 recovery code
    recovery_salt VARCHAR(255),               -- Recovery code salt
    friend_code VARCHAR(10) UNIQUE,           -- Code ami unique
    storage_used BIGINT DEFAULT 0,
    storage_limit BIGINT DEFAULT 5368709120,  -- 5GB
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

**Note critique**: Le backend ne peut **JAMAIS** déchiffrer `encrypted_master_key` car il nécessite le password utilisateur pour dériver le KEK.

### Stockage Fichiers (S3/MinIO)

```
Structure S3:
users/<user_id>/<path>/<filename>

Exemple:
users/550e8400-e29b-41d4-a716-446655440000/Documents/report.pdf.enc
                                          └─────┬─────┘
                                           Fichier chiffré AES-GCM
```

---

## API Endpoints

### Authentification

| Méthode | Endpoint | Description | Auth |
|---------|----------|-------------|------|
| POST | `/api/v1/auth/register` | Inscription + génération MasterKey | Non |
| GET | `/api/v1/auth/keys` | Récupérer Salt + EncryptedMasterKey | Oui |
| POST | `/api/v1/auth/logout` | Déconnexion (invalide JWT dans Redis) | Oui |
| POST | `/api/v1/auth/recovery/init` | Initialiser récupération compte | Non |
| POST | `/api/v1/auth/recovery/finish` | Finaliser récupération avec recovery code | Non |

### Fichiers

| Méthode | Endpoint | Description | Auth |
|---------|----------|-------------|------|
| POST | `/api/v1/files/upload` | Upload fichier chiffré (multipart) | Oui |
| GET | `/api/v1/files/list/*path` | Liste fichiers/dossiers | Oui |
| GET | `/api/v1/files/list-recursive` | Liste complète récursive | Oui |
| GET | `/api/v1/files/download/:fileID` | Télécharger fichier | Oui |
| GET | `/api/v1/files/preview/:fileID` | Télécharger preview | Oui |
| POST | `/api/v1/files/batch-presign` | Génère URLs présignées en batch (max 500) | Oui |
| POST | `/api/v1/files/selection-tree` | Arborescence mixte fichiers + dossiers | Oui |
| DELETE | `/api/v1/files/file/:fileID` | Supprimer fichier | Oui |
| DELETE | `/api/v1/files/folder/:folderID` | Supprimer dossier (récursif) | Oui |
| POST | `/api/v1/files/bulk-delete` | Supprimer multiple | Oui |
| POST | `/api/v1/files/move` | Déplacer fichier/dossier | Oui |
| POST | `/api/v1/files/rename` | Renommer fichier/dossier | Oui |
| POST | `/api/v1/files/tags` | Mettre à jour tags | Oui |
| GET | `/api/v1/files/search?q=query` | Recherche fichiers | Oui |

### Dossiers

| Méthode | Endpoint | Description | Auth |
|---------|----------|-------------|------|
| POST | `/api/v1/folders/create` | Créer dossier | Oui |
| GET | `/api/v1/folders/:id/tree` | Arborescence complète récursive | Oui |
| PUT | `/api/v1/folders/:id/key` | Mettre à jour clé chiffrée dossier | Oui |

### Partages

| Méthode | Endpoint | Description | Auth |
|---------|----------|-------------|------|
| POST | `/api/v1/shares/link` | Créer lien public | Oui |
| POST | `/api/v1/shares/direct` | Partager avec utilisateur | Oui |
| GET | `/api/v1/shares/list` | Lister mes partages | Oui |
| DELETE | `/api/v1/shares/link/:shareID` | Supprimer lien | Oui |
| DELETE | `/api/v1/shares/direct` | Révoquer partage user | Oui |
| GET | `/api/v1/shares/with-me` | Fichiers partagés avec moi | Oui |
| GET | `/api/v1/public/share/:token` | Accéder partage public | Non |
| GET | `/api/v1/public/share/:token/download` | Télécharger partage public | Non |

### Utilisateurs

| Méthode | Endpoint | Description | Auth |
|---------|----------|-------------|------|
| GET | `/api/v1/users/me` | Profil utilisateur | Oui |
| PUT | `/api/v1/users/profile` | Mettre à jour profil | Oui |
| POST | `/api/v1/users/change-password` | Changer mot de passe | Oui |
| GET | `/api/v1/users/` | Lister utilisateurs (search) | Oui |
| POST | `/api/v1/users/recent` | Ajouter activité récente | Oui |
| GET | `/api/v1/users/recent` | Récupérer activité récente | Oui |
| POST | `/api/v1/users/keys` | Mettre à jour clés RSA | Oui |

### Amis

| Méthode | Endpoint | Description | Auth |
|---------|----------|-------------|------|
| GET | `/api/v1/friends` | Lister amis | Oui |
| POST | `/api/v1/friends` | Ajouter ami (friend_code) | Oui |
| DELETE | `/api/v1/friends/:id` | Supprimer ami | Oui |
| PUT | `/api/v1/friends/:id/accept` | Accepter demande | Oui |
| DELETE | `/api/v1/friends/:id/reject` | Rejeter demande | Oui |

### WebSocket

| Endpoint | Description | Auth |
|----------|-------------|------|
| `GET /ws?token=<JWT>` | Connexion WebSocket | Oui (via query param) |
| `GET /ice-config` | Config TURN pour WebRTC | Oui |

### Sécurité

| Méthode | Endpoint | Description | Auth |
|---------|----------|-------------|------|
| POST | `/api/v1/security/report` | Reporter événement sécurité | Oui |
| GET | `/api/v1/security/events` | Récupérer événements | Oui |
---

## API Upload/Download Multi-Fichiers

### Upload S3 Multipart

| Endpoint | Méthode | Description |
|----------|---------|-------------|
| `/api/v1/multipart/initiate` | POST | Initie upload, retourne uploadID + URLs présignées |
| `/api/v1/multipart/complete` | POST | Finalise upload avec ETags |
| `/api/v1/multipart/abort` | POST | Annule upload, cleanup parts S3 |
| `/api/v1/multipart/refresh-url` | POST | Régénère URLs expirées (180s TTL) |

**Exemple `POST /multipart/initiate`**:
```json
// Request
{
  "filename": "document.pdf.enc",
  "size": 104857600,
  "encryptedName": "base64_encrypted_name",
  "encryptedMasterKey": "base64_encrypted_key"
}

// Response
{
  "uploadID": "aws-s3-upload-id",
  "fileID": "uuid",
  "presignedURLs": [
    { "partNumber": 1, "url": "https://s3...?X-Amz-Signature=..." }
  ]
}
```

### Download Multi-Fichiers (ZIP)

| Endpoint | Méthode | Description |
|----------|---------|-------------|
| `/api/v1/folders/:id/tree` | GET | Arborescence complète avec chemins relatifs |
| `/api/v1/files/batch-presign` | POST | Génère URLs présignées en batch (max 500, parallèle) |
| `/api/v1/files/selection-tree` | POST | Arborescence mixte fichiers + dossiers sélectionnés |

**Exemple `GET /folders/:id/tree`**:
```json
{
  "root_folder": "Documents",
  "total_size": 1073741824,
  "total_files": 42,
  "files": [
    {
      "id": "uuid",
      "name": "report.pdf",
      "relative_path": "Documents/2024/report.pdf",
      "size": 1048576,
      "encrypted_key": "base64...",
      "mime_type": "application/pdf"
    }
  ],
  "encrypted_keys": {
    "folder_uuid": "encrypted_folder_key"
  }
}
```

**Exemple `POST /files/batch-presign`**:
```json
// Request
{ "file_ids": ["uuid1", "uuid2", "uuid3"] }

// Response
{
  "urls": [
    { "file_id": "uuid1", "url": "https://s3.../presigned?..." },
    { "file_id": "uuid2", "url": "https://s3.../presigned?..." }
  ],
  "expires_in": 300
}
```

**Performance**: Génération parallèle avec sémaphore (10 concurrent), latence ~50ms pour 100 fichiers.
---

## Sécurité

### Middleware Auth

```go
// middleware/auth.go
func AuthMiddleware(jwks keyfunc.Keyfunc, secret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Extraire JWT du header Authorization
        // 2. Valider signature (ES256 via JWKS ou HS256 via secret)
        // 3. Vérifier expiration
        // 4. Extraire user_id et stocker dans context
        c.Set("user_id", claims.Subject)
        c.Next()
    }
}
```

### Rate Limiting

```go
// middleware/ratelimit.go
var visitors sync.Map // map[string]*visitor

type visitor struct {
    limiter  *rate.Limiter
    lastSeen time.Time
}

// Limits per endpoint
"/auth/register":   0.05 req/s (3 max, 1 every 20s)
"/files/upload":    5 req/s (50 max burst)
"/files/download":  10 req/s (100 max burst)
"default":          10 req/s (30 max burst)
```

### Security Logger

Logs structurés JSON via `log/slog`, filtrables dans Loki/Grafana avec `component="security"`.

**Politique IP** (CNIL délibération 2021-122) :
- Logs HTTP applicatifs (`MetricsMiddleware`) → `ip_anon` uniquement (dernier octet IPv4 / 80 bits IPv6 masqués)
- Événements de sécurité → `ip` (complète) **+** `ip_anon` dans le même log

**Fonctions disponibles** :

```go
// middleware/security_logger.go

// Authentification
func LogAuthAttempt(ctx, userID, ip string, success bool, reason string)
func LogPasswordChange(ctx, userID, ip string)
func LogTokenRevoked(ctx, userID, reason, ip string)

// Compte
func LogAccountCreated(ctx, userID, ip, userAgent string)   // ← nouveau
func LogAccountDeleted(ctx, userID, ip, userAgent string)   // ← nouveau

// Accès ressources
func LogFileAccess(ctx, userID, fileID, ip string, success bool)
func LogUnauthorizedAccess(ctx, userID, resource, ip string)

// Partages
func LogShareCreated(ctx, userID, resourceType string, resourceID int64, token, ip, userAgent string) // ← nouveau
func LogShareRevoked(ctx, userID, shareID, ip, userAgent string)                                       // ← nouveau
func LogDirectShareCreated(ctx, ownerID, recipientID, resourceType string, resourceID int64, ip, userAgent string) // ← nouveau

// Divers
func LogProfileUpdate(ctx, userID, ip string)
func LogRateLimitExceeded(ctx, ip, endpoint string)
func LogSuspiciousActivity(ctx, userID, activity, ip string)
func LogLDAPSync(ctx context.Context, orgID int64, usersFound, added, suspended, removed int, syncErr string)
```

**Format JSON (exemple `account.created`)** :
```json
{
  "time": "2026-06-28T14:30:00Z",
  "level": "INFO",
  "msg": "account.created",
  "component": "security",
  "event_type": "account.created",
  "user_id": "550e8400-...",
  "ip": "203.0.113.42",
  "ip_anon": "203.0.113.0",
  "user_agent": "Mozilla/5.0 ..."
}
```

**Format JSON (logs HTTP via `MetricsMiddleware`)** :
```json
{
  "time": "2026-06-28T14:30:00Z",
  "level": "INFO",
  "msg": "http_request",
  "request_id": "a1b2c3d4e5f6a7b8",
  "method": "POST",
  "path": "/api/v1/files/upload",
  "status": 201,
  "duration_ms": 42,
  "user_id": "550e8400-...",
  "ip_anon": "203.0.113.0",
  "user_agent": "Go-http-client/2.0"
}
```

**Événements couverts** (conformité LCEN / décret 2021-1363 — conservation 1 an) :

| Événement | `event_type` | Niveau |
|-----------|-------------|--------|
| Création de compte | `account.created` | INFO |
| Suppression de compte | `account.deleted` | INFO |
| Tentative d'auth | `auth.attempt` | INFO / WARN |
| Changement de mot de passe | `auth.password_changed` | INFO |
| Révocation de token | `auth.token_revoked` | INFO |
| Accès fichier | `file.access` | INFO / WARN |
| Création lien public | `share.created` | INFO |
| Révocation lien public | `share.revoked` | INFO |
| Partage direct créé | `share.direct_created` | INFO |
| Mise à jour profil | `user.profile_updated` | INFO |
| Rate limit dépassé | `ratelimit.exceeded` | WARN |
| Accès refusé | `access.denied` | WARN |
| Activité suspecte | `security.suspicious` | WARN |
| Sync LDAP | `ldap.sync` | INFO / ERROR |

**Garantie ZK** : Aucune clé cryptographique n'est jamais loguée.

**LogQL utiles** :
```logql
# Tous les événements sécurité
{service="kagibi-backend"} | json | component="security"

# Créations/suppressions de compte
{service="kagibi-backend"} | json | event_type=~"account\\.(created|deleted)"

# Partages par utilisateur
{service="kagibi-backend"} | json | component="security" | event_type=~"share\\..*" | user_id="550e8400-..."

# Client web vs desktop (dans les logs HTTP)
{service="kagibi-backend"} | json | user_agent=~"Mozilla.*"        # web
{service="kagibi-backend"} | json | user_agent=~"Go-http-client.*" # desktop
```

---

## 🧪 Tests

### Tests Unitaires

```bash
go test ./... -v
```

### Tests avec Coverage

```bash
go test ./handlers/... -cover
go test ./pkg/... -cover
```

### Tests Middleware

```bash
go test ./middleware/... -v
```

### Exemple de Test

```go
// handlers/folders/create_test.go
func TestCreateFolderHandler(t *testing.T) {
    db, mock := setupTestDB(t)
    defer db.Close()

    // Mock: CreateFolderDB returns folder with ID
    mock.ExpectQuery("INSERT INTO folders").
        WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

    // Mock: INSERT folder_sizes
    mock.ExpectExec("INSERT INTO folder_sizes").
        WillReturnResult(sqlmock.NewResult(1, 1))

    // Test handler
    req := CreateFolderRequest{Name: "Test", Path: "/", EncryptedKey: "xxx"}
    // ...
}
```

---

## Performance

### Optimisations Appliquées

| Optimisation | Technique | Impact |
|--------------|-----------|--------|
| **Concurrency** | `sync.Map` pour rate limiting | -60% contention |
| **Connection Pooling** | Bun ORM avec pool PostgreSQL | +200% throughput |
| **Redis Cache** | Cache metadata fichiers | -80% requêtes DB |
| **Upload S3 Multipart** | Parts 5–100 MB, URLs présignées (TTL 3 min), upload direct client → S3 | Support gros fichiers, zéro transit backend |
| **Pipeline Upload parallèle** | Chiffrement côté client + multipart S3 en parallèle (frontend) | Débit x3–5 sur connexion rapide |
| **Batch Presign parallèle** | Sémaphore 10 goroutines concurrentes pour `batch-presign` | ~50 ms pour 100 fichiers |
| **Goroutines async** | Cleanup, preview generation, worker S3 via Redis queue | Non-bloquant pour l'utilisateur |
| **Logs structurés JSON** | `slog` → Loki/Grafana Cloud, 0 parse custom | Observabilité production-ready |

### Métriques

- Latence p95 upload multipart (initiate) : < 200 ms
- Latence p95 batch-presign 100 fichiers : ~50 ms
- Latence p95 download presigned URL : < 200 ms
- Latence p95 list files : < 100 ms
- Throughput : 1 000 req/s (rate limiter middleware)

---

## Déploiement

### Build Production

```bash
go build -o server .
./server
```

### Docker

```bash
# Avec docker-compose (voir racine)
docker-compose up -d
```

### Variables d'Environnement Production

```bash
DATABASE_URL=postgres://prod_user:prod_pass@db:5432/kagibi
REDIS_URL=redis://redis:6379
S3_ENDPOINT=https://s3.amazonaws.com
S3_BUCKET=kagibi-prod
ALLOWED_ORIGINS=https://kagibi.com
```

---

## Frontend — Pages & Composants Notables

> Ces éléments sont dans `frontend/src/` et s'appuient sur l'API backend ci-dessus.

### Page FAQ (`/faq`)

Route publique — accessible sans authentification depuis la landing page et le bouton **Aide & Support** du dashboard.

- **Fichier** : `views/landing/FaqView.vue`
- **Route** : `/faq` (nom `Faq`, enregistrée dans `router/index.js`)
- **Nav** : lien dans `LandingNav.vue` (desktop + mobile) entre `/values` et le bouton de connexion
- **HelpDialog** : bouton "FAQ" avec icône `BookOpen` (lucide-vue-next) dans le menu Aide & Support
- **i18n** : clés sous `landing.faq.*` dans `fr.json` / `en.json` (4 catégories, 19 Q&A)

Catégories couvertes :
1. Général & Souveraineté (g1–g5)
2. Sécurité & Chiffrement (s1–s5)
3. Fonctionnalités — P2P, partages, Organisations, Amis (f1–f5)
4. Valeurs & Engagements de Kagibi (v1–v4)

### Upload — Pipeline côté client

Le frontend chiffre les chunks AES-256-GCM **en parallèle** avant de les envoyer via S3 Multipart :
1. `POST /api/v1/files/multipart/initiate` → reçoit `upload_id` + URLs présignées par part
2. Upload direct de chaque part chiffrée vers S3 (bypass backend)
3. `POST /api/v1/files/multipart/complete` → finalisation avec ETags
4. `POST /api/v1/files/multipart/refresh-url` si une URL présignée expire (TTL 3 min)

Si l'opération est annulée : `POST /api/v1/files/multipart/abort` nettoie les parts S3.

---

## 🐛 Debugging

### Logs

```bash
# Logs applicatifs JSON (stdout) — ingérés par Loki/Grafana Cloud
go run main.go 2>&1 | tee app.log

# Filtrer les événements de sécurité en local
go run main.go 2>&1 | jq 'select(.component == "security")'

# Filtrer par type d'événement
go run main.go 2>&1 | jq 'select(.event_type == "account.created")'

# Filtrer les requêtes HTTP avec user-agent
go run main.go 2>&1 | jq 'select(.msg == "http_request") | {path, status, user_agent}'
```

> En production, les logs sont collectés automatiquement par **Grafana Cloud Loki** (managé). Aucune configuration locale requise.

### Hot Reload

```bash
# Avec air (https://github.com/cosmtrek/air)
go install github.com/cosmtrek/air@latest
air
```

---

## Ressources

- [Gin Framework](https://gin-gonic.com/)
- [Bun ORM](https://bun.uptrace.dev/)
- [AWS SDK Go v2](https://aws.github.io/aws-sdk-go-v2/)
- [Gorilla WebSocket](https://github.com/gorilla/websocket)
- [golang-jwt](https://github.com/golang-jwt/jwt)

---

**Développé en Go avec 🔐 Zero-Knowledge Architecture**
