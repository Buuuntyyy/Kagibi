# 🔐 Kagibi - Zero-Knowledge Cloud Storage

**Architecture Zero-Knowledge End-to-End** | **AGPLv3 License** | **Production-Ready**

Kagibi est une plateforme de stockage cloud sécurisée où toutes les données sont chiffrées côté client avant d'atteindre le serveur. Le backend ne possède **aucune clé de déchiffrement** et ne peut accéder au contenu des fichiers.

---

## 📋 Table des Matières

- [Architecture Zero-Knowledge](#️-architecture-zero-knowledge)
- [Stack Technique](#️-stack-technique)
- [Démarrage Rapide](#-démarrage-rapide)
- [Configuration](#️-configuration)
- [Sécurité](#-sécurité)
- [Structure du Projet](#-structure-du-projet)
- [License](#-license)

---

## 🏗️ Architecture Zero-Knowledge

```
┌─────────────────────────────────────────────────────────────────┐
│                      FLUX DE CHIFFREMENT                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  1. CLIENT (Vue.js)                                              │
│     ├── Mot de passe utilisateur                                │
│     ├── Dérivation Argon2id (64MB, 4 passes) → KEK              │
│     ├── Génération AES-GCM 256 bits → MasterKey (RAM uniquement)│
│     ├── Génération Base Nonce (8B CSPRNG) par fichier           │
│     ├── MasterKey chiffre fichiers (AES-GCM)                    │
│     │   └─ Chunks 10MB avec Nonce = Base + Counter (12B total)  │
│     └── KEK enveloppe MasterKey → EncryptedMasterKey            │
│                                                                  │
│  2. UPLOAD MULTIPART DIRECT-TO-S3                                │
│     ├── Backend génère URLs présignées (180s TTL)               │
│     ├── Client upload parallèle chunks chiffrés → S3            │
│     ├── Format: [Nonce 12B][Ciphertext][Tag 16B] par chunk     │
│     ├── Backend reçoit ETags + Complete Multipart               │
│     └── ❌ Backend ne voit JAMAIS le contenu                    │
│                                                                  │
│  3. BACKEND (Go)                                                 │
│     ├── Stocke EncryptedMasterKey (inutilisable sans password)  │
│     ├── Orchestration S3 (initiate/complete/abort)              │
│     ├── Métadonnées PostgreSQL (tailles, timestamps)            │
│     └── ❌ AUCUNE CLÉ DE DÉCHIFFREMENT                          │
│                                                                  │
│  4. DOWNLOAD STREAMING                                           │
│     ├── Backend génère URL présignée S3 (5min TTL)              │
│     ├── Client fetch() → ReadableStream                         │
│     ├── TransformStream déchiffre chunks à la volée             │
│     │   └─ Parse [Nonce][Ciphertext][Tag] → AES-GCM decrypt    │
│     ├── FileSystemWritableFileStream ou Blob                    │
│     └── Aucun stockage déchiffré en mémoire (backpressure)      │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Garanties ZK

| Composant | Garantie |
|-----------|----------|
| **MasterKey** | Jamais envoyée au backend, stockée en RAM uniquement |
| **Fichiers** | Chiffrés en AES-GCM avant upload avec chunks de 10MB |
| **Upload** | Direct-to-S3 multipart avec URLs présignées (backend ne touche jamais le contenu) |
| **Download** | Streaming avec décryptage à la volée (pas de stockage temporaire déchiffré) |
| **Nonces** | Génération CSPRNG conforme NIST SP 800-38D avec détection de réutilisation |
| **Partage** | RSA-OAEP 4096 bits pour chiffrer clés symétriques |
| **Métadonnées** | Noms de fichiers chiffrés, seules tailles/timestamps en clair |
| **Recovery** | Code de récupération avec hash SHA-256 + dérivation Argon2id |

---

## 🚀 Fonctionnalités

### Gestion de Fichiers
*   **Upload Multi-Fichiers avec Queue** : Gestionnaire de file d'attente pour uploads simultanés
    - 3 fichiers uploadés en parallèle avec priorité FIFO
    - Chiffrement AES-GCM par Web Worker (non-bloquant UI)
    - Progress bar individuel et global avec vitesse/ETA
    - Retry automatique (3 tentatives, backoff exponentiel)
    - Annulation individuelle ou globale
*   **Upload Direct-to-S3 Multipart** : Upload parallélisé avec chunks de 10MB directement vers S3 via URLs présignées (180s TTL)
    - 3 workers parallèles par fichier pour upload des parts
    - Capture des ETags via XHR pour validation S3
    - Annulation d'upload avec nettoyage des parts incomplètes
*   **Download Multi-Fichiers ZIP Streaming** : Téléchargement de dossiers/sélections en ZIP Zero-Knowledge
    - Service Worker pour assemblage ZIP en streaming (fflate level 0)
    - 4 fichiers téléchargés en parallèle avec pool de connexions
    - Décryptage AES-GCM streaming avec backpressure
    - Batch presigned URLs (jusqu'à 500 fichiers, génération parallèle)
    - Progress tracking avec ETA et vitesse
    - Fallback in-memory pour navigateurs sans Service Worker
*   **Streaming Download avec Décryptage** : Téléchargement fichier unique mémoire-efficient
    - Décryptage à la volée par chunks (pas de buffer complet en RAM)
    - Support FileSystem Access API avec fallback Blob
    - Gestion du backpressure pour éviter la saturation mémoire
*   **Organisation** : Création de dossiers, déplacement, renommage et suppression
*   **Navigation** : Interface fluide type "Google Drive" avec vue liste et grille
*   **Recherche** : Barre de recherche rapide pour filtrer les fichiers
*   **Tags** : Système de tags pour classer les fichiers

### Partage & Collaboration
*   **Partage Utilisateur** : Partage sécurisé de fichiers avec d'autres utilisateurs de la plateforme
*   **Liens Publics** : Génération de liens de partage accessibles publiquement (avec expiration optionnelle et mot de passe optionnel)
*   **Permissions** : Gestion fine des droits d'accès
*   **WebRTC P2P** : Transfert direct entre utilisateurs (optionnel)

### Interface Utilisateur (UI/UX)
*   **Design Moderne** : Interface épurée avec icônes SVG dynamiques selon le type de fichier
*   **Thème Sombre/Clair** : Support natif du Dark Mode via CSS variables
*   **Responsive** : Adapté aux écrans de bureau et tablettes
*   **Feedback Visuel** : Dialogues de confirmation personnalisés, notifications toast, et indicateurs de chargement

---

## 🚀 Architecture Upload/Download Optimisée

### Upload Multipart Direct-to-S3

```
┌─────────────────────────────────────────────────────────────┐
│  FLUX UPLOAD MULTIPART                                       │
├─────────────────────────────────────────────────────────────┤
│  1. Client → Backend: POST /multipart/initiate              │
│     ├── filename, size, encryptedName, encryptedMasterKey   │
│     └── Backend → DB: Création record file (état: uploading)│
│                                                              │
│  2. Backend → Client: { uploadID, presignedURLs[] }         │
│     ├── 1 URL présignée par part (180s TTL)                 │
│     ├── Content-Length verrouillé (anti-tampering)          │
│     └── Région OVH/AWS avec retry policy                    │
│                                                              │
│  3. Client → S3: PUT parallèle (3 workers max)              │
│     ├── [Nonce 12B][Ciphertext][Tag 16B] par chunk         │
│     ├── ETag capture via XMLHttpRequest                     │
│     ├── Retry automatique (3 tentatives, backoff exp.)      │
│     └── Progress tracking temps réel                        │
│                                                              │
│  4. Client → Backend: POST /multipart/complete              │
│     ├── { uploadID, parts: [{ PartNumber, ETag }] }        │
│     ├── Backend → S3: CompleteMultipartUpload               │
│     └── Backend → DB: État = 'completed'                    │
│                                                              │
│  ⚠️  Annulation: POST /multipart/abort                       │
│     ├── Backend → S3: AbortMultipartUpload                  │
│     └── Backend → DB: Suppression record                    │
└─────────────────────────────────────────────────────────────┘
```

**Avantages**:
- ✅ Backend ne touche jamais les données (zero-knowledge preserved)
- ✅ Pas de réassemblage côté serveur (économie CPU/RAM)
- ✅ Support fichiers jusqu'à 5TB (limite S3)
- ✅ Résistance aux coupures réseau (retry + resume)

### Download Streaming avec Décryptage

```
┌─────────────────────────────────────────────────────────────┐
│  FLUX DOWNLOAD STREAMING                                     │
├─────────────────────────────────────────────────────────────┤
│  1. Client → Backend: GET /download/:id/presigned           │
│     ├── Backend vérifie ownership                           │
│     ├── Backend → S3: GeneratePresignedURL (5min TTL)       │
│     └── Backend → Client: { url, metadata }                 │
│                                                              │
│  2. Client → S3: fetch(presignedURL)                        │
│     ├── Response.body → ReadableStream<Uint8Array>          │
│     └── Pas de buffer complet en RAM                        │
│                                                              │
│  3. TransformStream Pipeline:                               │
│     ┌──────────────────────────────────────────┐            │
│     │ S3 Stream → Chunk Parser                 │            │
│     │   ├── Buffer incomplet jusqu'à complet   │            │
│     │   └── Extract [Nonce][Ciphertext][Tag]  │            │
│     ├────────────────────────────────────────────           │
│     │ Chunk Parser → AES-GCM Decryptor         │            │
│     │   ├── crypto.subtle.decrypt()            │            │
│     │   └── Plaintext chunks                   │            │
│     ├────────────────────────────────────────────           │
│     │ Decryptor → FileSystemWritableFileStream │            │
│     │   ├── Écriture directe sur disque        │            │
│     │   └── Backpressure handling              │            │
│     └──────────────────────────────────────────┘            │
│                                                              │
│  4. Fallback si FileSystem API indisponible:                │
│     ├── Accumulation en Blob                                │
│     └── Download via <a download>                           │
└─────────────────────────────────────────────────────────────┘
```

**Avantages**:
- ✅ Consommation mémoire constante (~20-40MB max)
- ✅ Support fichiers multi-GB sans crash
- ✅ Début du téléchargement immédiat (pas d'attente fin décryptage)
- ✅ Annulation propre avec cleanup

---

## 🛠️ Stack Technique

### Frontend

| Technologie | Version | Usage |
|-------------|---------|-------|
| **Vue.js** | 3.5.24 | Framework réactif |
| **Pinia** | 3.0.4 | State management |
| **Web Crypto API** | Native | Chiffrement AES-GCM, RSA-OAEP |
| **libsodium-wrappers** | 0.7.15 | Argon2id, utilitaires crypto |
| **Axios** | 1.13.2 | Client HTTP |
| **Supabase Client** | 2.90.1 | Authentification JWT |
| **PrimeVue** | 4.4.1 | Composants UI |
| **Vite** | 7.2.4 | Build tool |

### Backend

| Technologie | Version | Usage |
|-------------|---------|-------|
| **Go** | 1.21+ | Runtime serveur |
| **Gin** | 1.11.0 | Framework HTTP |
| **Bun ORM** | 1.2.16 | ORM PostgreSQL |
| **PostgreSQL** | 16+ | Base de données |
| **Redis** | 7+ | Cache & rate limiting |
| **AWS SDK v2** | 1.40.0 | Interface S3/MinIO |
| **MinIO** | Compatible S3 | Stockage objet |
| **Gorilla WebSocket** | 1.5.3 | WebSocket temps réel |
| **JWT** | 5.3.0 | Validation tokens |

---

## 🚀 Démarrage Rapide

### Prérequis

```bash
# Frontend
Node.js 18+ ou Bun
npm ou bun

# Backend
Go 1.21+
PostgreSQL 16+
Redis 7+
MinIO ou AWS S3
```

### Installation

1. **Cloner le repository**
```bash
git clone https://github.com/votre-org/kagibi.git
cd kagibi
```

2. **Démarrer l'infrastructure (Docker)**
```bash
docker-compose up -d
# Démarre: PostgreSQL, Redis, MinIO
```

3. **Lancer Backend**
```bash
cd backend
go mod download
go run main.go
# → API disponible sur http://localhost:8080
```

4. **Lancer Frontend**
```bash
cd frontend
npm install  # ou: bun install
npm run dev  # ou: bun run dev
# → UI disponible sur http://localhost:5173
```

Consultez [`frontend/README.md`](frontend/README.md) et [`backend/README.md`](backend/README.md) pour plus de détails.

---

## ⚙️ Configuration

### Variables d'Environnement Backend

Créez `backend/.env`:

```bash
# Base de données
DATABASE_URL=postgresql://user:password@localhost:5432/kagibi

# Redis
REDIS_URL=redis://localhost:6379

# Supabase (Authentification)
SUPABASE_URL=https://xxx.supabase.co
SUPABASE_JWT_SECRET=your-jwt-secret

# Stockage S3/MinIO
S3_ENDPOINT=http://localhost:9000  # MinIO local
S3_REGION=us-east-1
S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=minioadmin
S3_BUCKET=kagibi-files

# CORS
ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000

# TURN (optionnel - WebRTC P2P)
TURN_URLS=turn:your-turn-server.com:3478
TURN_SECRET=your-turn-secret
```

### Variables d'Environnement Frontend

Créez `frontend/.env`:

```bash
# Backend API
VITE_API_URL=http://localhost:8080/api/v1

# Supabase (Authentification)
VITE_SUPABASE_URL=https://xxx.supabase.co
VITE_SUPABASE_ANON_KEY=your-anon-key

# WebSocket (optionnel)
VITE_WS_URL=ws://localhost:8080/ws
```

### Configuration S3 pour Multipart Upload

Appliquez les configurations CORS et Lifecycle depuis `scripts/`:

```bash
# CORS (expose ETags pour validation client)
aws s3api put-bucket-cors \
  --bucket kagibi-files \
  --cors-configuration file://scripts/s3-cors-config.json \
  --endpoint-url http://localhost:9000  # MinIO local

# Lifecycle (cleanup uploads incomplets après 24h)
aws s3api put-bucket-lifecycle-configuration \
  --bucket kagibi-files \
  --lifecycle-configuration file://scripts/s3-lifecycle-policy.json \
  --endpoint-url http://localhost:9000
```

**Contenu `s3-cors-config.json`**:
```json
{
  "CORSRules": [
    {
      "AllowedOrigins": ["http://localhost:5173", "https://votre-domaine.com"],
      "AllowedMethods": ["GET", "PUT", "POST", "HEAD"],
      "AllowedHeaders": ["*"],
      "ExposeHeaders": ["ETag", "Content-Length"],
      "MaxAgeSeconds": 3600
    }
  ]
}
```

**Contenu `s3-lifecycle-policy.json`**:
```json
{
  "Rules": [
    {
      "ID": "cleanup-incomplete-multipart-uploads",
      "Status": "Enabled",
      "Filter": { "Prefix": "" },
      "AbortIncompleteMultipartUpload": {
        "DaysAfterInitiation": 1
      }
    }
  ]
}
```

---

## 🔒 Sécurité

### Conformité

- ✅ **NIST SP 800-38D** - Modes GCM conformes (nonce 96 bits, CSPRNG, limite 2^32 invocations)
- ✅ **NIST SP 800-90A** - Générateurs aléatoires cryptographiques (crypto/rand, Web Crypto API)
- ✅ **RFC 9106** - Argon2id pour dérivation de clés (64MB, 4 passes)
- ✅ **RGPD** - Architecture zero-knowledge conforme
- ✅ **OWASP Top 10** - Toutes vulnérabilités corrigées (voir [SECURITY_AUDIT_REPORT.md](SECURITY_AUDIT_REPORT.md))
- ✅ **ANSSI** - Guide de sélection d'algorithmes cryptographiques (2021)
- ✅ **Rate Limiting** - Protection DDoS avec `sync.Map`
- ✅ **CORS** - Origines configurables
- ✅ **CSP** - Content Security Policy strict
- ✅ **Timing Attack Mitigation** - Délais constants sur endpoints sensibles

### Cryptographie

| Algorithme | Usage | Paramètres | Standard |
|-----------|-------|------------|----------|
| **Argon2id** | Dérivation de clé depuis password | 64MB RAM, 4 itérations | RFC 9106 |
| **AES-GCM** | Chiffrement symétrique fichiers | 256 bits, nonce 96 bits | NIST SP 800-38D |
| **Nonce/IV** | Vecteur d'initialisation unique | 12 octets (96 bits), CSPRNG | NIST SP 800-38D §8 |
| **RSA-OAEP** | Chiffrement asymétrique (partage) | 4096 bits, SHA-256 | PKCS#1 v2.2 |
| **SHA-256** | Hachage codes de récupération | Standard | FIPS 180-4 |

#### Gestion des Nonces (IV) - Conformité NIST SP 800-38D

**Architecture hybride** : `[8 octets base random] + [4 octets counter]`

| Composant | Description |
|-----------|-------------|
| **Base Nonce** | 8 octets aléatoires via CSPRNG (crypto/rand Go, crypto.getRandomValues() JS) |
| **Chunk Nonce** | Nonce déterministe = Base (8B) + Index little-endian (4B) = 12 octets total |
| **Unicité inter-fichiers** | 2^64 combinaisons (base aléatoire) |
| **Unicité intra-fichier** | 2^32 chunks max (~42.9 TB par fichier à 10MB/chunk) |
| **Limite NIST** | < 2^32 invocations par clé (respectée) |
| **Format fil** | `[Nonce 12B] ‖ [Ciphertext] ‖ [Tag 16B]` |
| **Détection réutilisation** | Defense-in-depth avec Set tracking (10,000 limite/session) |

**Sources** :
- NIST SP 800-38D §8.2.1 - Recommandation 96 bits pour IV déterministes
- NIST SP 800-38D §8.3 - Limite 2^32 invocations
- ANSSI - Guide de sélection d'algorithmes cryptographiques (2021)

### Logs Sécurisés

Tous les événements de sécurité sont enregistrés dans `backend/logs/security.log` avec structure:
- Timestamp, Type d'événement, UserID, IP, Succès/Échec
- **Jamais de clés cryptographiques**

---

## 📁 Structure du Projet

```
kagibi/
├── backend/                    # API Go
│   ├── handlers/               # Contrôleurs HTTP
│   │   ├── auth/               # Authentification, récupération
│   │   ├── files/              # Upload, download, delete, multipart
│   │   ├── folders/            # Gestion dossiers
│   │   ├── friends/            # Système d'amis
│   │   ├── shares/             # Partages public/privé
│   │   ├── users/              # Profils utilisateur
│   │   └── ws/                 # WebSocket temps réel
│   ├── middleware/             # Auth, rate limit, sécurité
│   ├── pkg/                    # Logique métier
│   │   ├── crypto/             # Génération nonces NIST SP 800-38D
│   │   ├── database.go         # Opérations DB
│   │   ├── models.go           # Structures Bun
│   │   ├── s3storage/          # Client S3/MinIO
│   │   └── workers/            # Tâches async
│   ├── migrations/             # Migrations SQL
│   └── main.go                 # Point d'entrée
│
├── frontend/                   # SPA Vue.js
│   ├── src/
│   │   ├── components/         # Composants réutilisables
│   │   ├── stores/             # Pinia stores
│   │   │   ├── auth.js         # Authentification + crypto
│   │   │   ├── files.js        # Gestion fichiers (multipart upload/streaming download)
│   │   │   ├── uploads.js      # Queue multi-fichiers upload avec progress
│   │   │   ├── downloads.js    # Download multi-fichiers ZIP avec progress
│   │   │   ├── friends.js      # Amis
│   │   │   ├── p2p.js          # WebRTC
│   │   │   └── websocket.js    # Client WS
│   │   ├── utils/              # Utilitaires crypto
│   │   │   ├── crypto.js       # Fonctions AES, RSA, nonces
│   │   │   ├── multipartUpload.js  # Upload multipart avec retry
│   │   │   ├── uploadQueueManager.js  # Gestionnaire queue multi-fichiers
│   │   │   ├── streamingDownload.js  # Download streaming avec décryptage
│   │   │   ├── zipDownloadManager.js  # Download multi-fichiers ZIP
│   │   │   ├── secureCrypto.js # XSS monitoring, rate limit
│   │   │   └── securityMonitoring.js
│   │   ├── views/              # Pages
│   │   ├── workers/            # Web Workers (chiffrement avec tracking nonces)
│   │   └── router/             # Vue Router
│   ├── public/
│   │   ├── sw-crypto.js        # Service Worker (session)
│   │   └── download-worker.js  # Service Worker ZIP streaming (fflate)
│   └── index.html
│
├── scripts/                    # Scripts configuration
│   ├── s3-cors-config.json     # CORS S3 (ETag exposure)
│   └── s3-lifecycle-policy.json # Cleanup multipart uploads 24h
│
├── db_data/                    # Données PostgreSQL (gitignore)
├── redis-data/                 # Données Redis (gitignore)
├── docker-compose.yaml         # Stack complète
├── .github/workflows/ci.yml    # CI/CD
├── SECURITY_AUDIT_REPORT.md    # Audit sécurité détaillé
├── SECURITY_FIXES_SUMMARY.md   # Résumé correctifs
└── CHANGELOG.md                # Historique versions
```

---

## 🧪 Tests

### Backend
```bash
cd backend
go test ./... -v
go test ./handlers/... -cover

# Tests spécifiques nonces NIST
go test ./pkg/crypto/... -v
# ✅ 10,000 unicité, thread-safety, sérialisation
```

### Frontend
```bash
cd frontend
npm run test          # Vitest
npm run test:coverage # Coverage

# Tests spécifiques nonces
npm test -- nonce.test.js --run
# ✅ 10,000 unicité, entropie, round-trip serialization
```

### Tests de Conformité NIST SP 800-38D

| Test | Backend (Go) | Frontend (JS) | Standard |
|------|--------------|---------------|----------|
| Longueur nonce 96 bits | ✅ | ✅ | NIST SP 800-38D §5.2.1.1 |
| CSPRNG (crypto/rand) | ✅ | ✅ | NIST SP 800-90A |
| Unicité 10,000 samples | ✅ | ✅ | Birthday paradox test |
| Détection réutilisation | ✅ | ✅ | Defense-in-depth |
| Format fil standardisé | ✅ | ✅ | `[N‖C‖T]` 12+n+16 bytes |
| Counter little-endian | ✅ | ✅ | IEEE Std 1003.1 |
| Thread-safety | ✅ | N/A | sync.Mutex Go |

### CI/CD

GitHub Actions exécute automatiquement:
- Tests unitaires (Go + Vue)
- Analyse sécurité (go vet, npm audit)
- Build production

---

## 📚 Documentation

| Document | Description |
|----------|-------------|
| [`frontend/README.md`](frontend/README.md) | Guide frontend détaillé |
| [`backend/README.md`](backend/README.md) | Guide backend détaillé |
| [SECURITY_AUDIT_REPORT.md](SECURITY_AUDIT_REPORT.md) | Audit sécurité complet (28 Jan 2026) |
| [SECURITY_FIXES_SUMMARY.md](SECURITY_FIXES_SUMMARY.md) | Résumé des correctifs |
| [CHANGELOG.md](CHANGELOG.md) | Historique des versions |

---

## ❓ Pourquoi pas de connexion OAuth ?

### TL;DR
**L'authentification via Google, Facebook, Apple, etc. est incompatible avec l'architecture Zero-Knowledge de Kagibi.**

### Explication détaillée

#### Le problème fondamental

Kagibi utilise un **chiffrement Zero-Knowledge** où :
1. Votre **mot de passe** est la **seule source** pour dériver votre clé de chiffrement (via Argon2id)
2. Cette clé **n'existe que dans votre navigateur** et n'est **jamais envoyée** au serveur
3. Le serveur ne peut **jamais déchiffrer** vos fichiers

#### Pourquoi OAuth ne fonctionne pas ?

Avec OAuth (Google/Facebook/Apple) :
- ✅ Vous vous connectez facilement
- ❌ **Mais vous n'avez pas de mot de passe Kagibi**
- ❌ Sans mot de passe → **Impossible de dériver la clé de chiffrement**
- ❌ Sans clé → **Vos fichiers restent inaccessibles**

```
┌─────────────────────────────────────────────────────────┐
│ ARCHITECTURE ZERO-KNOWLEDGE                              │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  Mot de passe utilisateur                               │
│         ↓                                                │
│  Argon2id (64MB RAM, 4 passes)                          │
│         ↓                                                │
│  KEK (Key Encryption Key)                               │
│         ↓                                                │
│  Déchiffre MasterKey stockée côté serveur               │
│         ↓                                                │
│  MasterKey déchiffre tous les fichiers                  │
│                                                          │
│  ⚠️  SANS MOT DE PASSE → AUCUNE CLÉ POSSIBLE           │
└─────────────────────────────────────────────────────────┘
```

#### Solutions envisagées (et pourquoi elles échouent)

##### Option 1 : Générer un mot de passe aléatoire
**Problème** : L'utilisateur doit le sauvegarder quelque part → perd l'avantage d'OAuth

##### Option 2 : Utiliser le token OAuth comme clé
**Problème** :
- Les tokens OAuth changent à chaque connexion
- Si Google révoque le token → **Tous vos fichiers sont perdus**
- Le token transite par le serveur → **Ce n'est plus Zero-Knowledge**

##### Option 3 : Stocker la clé sur le serveur
**Problème** :
- Le serveur peut déchiffrer vos fichiers
- **Ce n'est plus Zero-Knowledge**
- Vous devez nous faire confiance (contraire à la philosophie du projet)

#### Comparaison avec d'autres services

| Service | OAuth | Zero-Knowledge | Explication |
|---------|-------|----------------|-------------|
| **Google Drive** | ✅ | ❌ | Google peut lire vos fichiers |
| **Dropbox** | ✅ | ❌ | Dropbox peut lire vos fichiers |
| **OneDrive** | ✅ | ❌ | Microsoft peut lire vos fichiers |
| **ProtonDrive** | ❌ | ✅ | Mot de passe obligatoire (comme nous) |
| **Tresorit** | ❌ | ✅ | Mot de passe obligatoire |
| **Kagibi** | ❌ | ✅ | **Vie privée > Commodité** |

#### Recommandation : Utilisez un gestionnaire de mots de passe

Pour concilier sécurité et commodité :
- **Bitwarden** (open-source, gratuit, sync cloud)
- **KeePassXC** (open-source, gratuit, local)
- **1Password** (payant, très bon UX)

**Avantages** :
- Connexion en 1 clic (auto-remplissage)
- Génération de mots de passe forts (20+ caractères)
- Synchronisation multi-appareils
- Compatible avec notre architecture Zero-Knowledge

#### Pourquoi nous ne changerons pas d'avis

> **"Si c'est gratuit, c'est vous le produit."**

Kagibi est conçu pour que :
- Nous ne puissions **jamais** lire vos fichiers
- Une fuite de notre base de données soit **inutile** aux attaquants
- Une ordonnance judiciaire ne puisse **pas** nous forcer à révéler vos données

**OAuth briserait cette garantie.**

#### Cas d'usage : Récupération de compte

Même sans OAuth, Kagibi offre une solution de secours :
- **Code de récupération** généré à l'inscription (256 bits)
- Stockez-le dans votre gestionnaire de mots de passe
- Permet de régénérer votre clé en cas d'oubli du mot de passe

---

## 🤝 Contribution

Ce projet est sous licence **AGPLv3**. Toute modification doit être partagée sous la même licence.

### Guidelines
1. Respecter l'architecture Zero-Knowledge
2. Aucune donnée sensible dans les logs
3. Tests unitaires obligatoires pour nouveaux features
4. Code review par 2 développeurs minimum

---

## 📄 License

**GNU Affero General Public License v3.0 (AGPLv3)**

Ce logiciel est libre et open-source. Vous êtes libre de:
- ✅ Utiliser commercialement
- ✅ Modifier le code
- ✅ Distribuer
- ✅ Utiliser à titre privé

**Obligations:**
- ⚠️ Divulgation du code source (même pour utilisation réseau)
- ⚠️ Même licence pour modifications
- ⚠️ Indication des changements

Voir [LICENSE](LICENSE) pour le texte complet.

---

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/votre-org/kagibi/issues)
- **Documentation**: Ce README + sous-dossiers
- **Security**: Voir [SECURITY_AUDIT_REPORT.md](SECURITY_AUDIT_REPORT.md)

---

**Développé avec ❤️ et 🔐 pour la vie privée**

---

## 📡 API Endpoints Multipart

### Upload Multipart

| Endpoint | Méthode | Auth | Description |
|----------|---------|------|-------------|
| `/api/v1/multipart/initiate` | POST | ✅ | Initie upload multipart, retourne uploadID + URLs présignées |
| `/api/v1/multipart/complete` | POST | ✅ | Finalise upload avec ETags, déclenche CompleteMultipartUpload S3 |
| `/api/v1/multipart/abort` | POST | ✅ | Annule upload, cleanup parts S3 |
| `/api/v1/multipart/refresh-url` | POST | ✅ | Régénère URLs présignées expirées (180s TTL) |

**Exemple `POST /multipart/initiate`**:
```json
{
  "filename": "document.pdf.enc",
  "size": 104857600,
  "encryptedName": "base64_encrypted_name",
  "encryptedMasterKey": "base64_encrypted_key",
  "folderID": "uuid-optional"
}
```

**Réponse**:
```json
{
  "uploadID": "aws-s3-upload-id",
  "fileID": "uuid",
  "presignedURLs": [
    { "partNumber": 1, "url": "https://s3.../part1?X-Amz-Signature=..." },
    { "partNumber": 2, "url": "https://s3.../part2?X-Amz-Signature=..." }
  ]
}
```

### Download Streaming

| Endpoint | Méthode | Auth | Description |
|----------|---------|------|-------------|
| `/api/v1/download/:fileID/presigned` | GET | ✅ | Génère URL présignée S3 (5min TTL) + métadonnées |

**Réponse**:
```json
{
  "url": "https://s3.../file?X-Amz-Signature=...",
  "metadata": {
    "size": 104857600,
    "encryptedName": "base64...",
    "encryptedMasterKey": "base64...",
    "mimeType": "application/pdf"
  }
}
```

### Download Multi-Fichiers (ZIP)

| Endpoint | Méthode | Auth | Description |
|----------|---------|------|-------------|
| `/api/v1/folders/:id/tree` | GET | ✅ | Arborescence complète d'un dossier avec chemins relatifs |
| `/api/v1/files/batch-presign` | POST | ✅ | Génère URLs présignées en batch (max 500) |
| `/api/v1/files/selection-tree` | POST | ✅ | Arborescence mixte fichiers + dossiers sélectionnés |

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
      "relative_path": "Documents/report.pdf",
      "size": 1048576,
      "encrypted_key": "base64..."
    }
  ],
  "encrypted_keys": { "folder_id": "encrypted_folder_key" }
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
