# 🔐 SaferCloud - Zero-Knowledge Cloud Storage

**Architecture Zero-Knowledge End-to-End** | **AGPLv3 License** | **Production-Ready**

SaferCloud est une plateforme de stockage cloud sécurisée où toutes les données sont chiffrées côté client avant d'atteindre le serveur. Le backend ne possède **aucune clé de déchiffrement** et ne peut accéder au contenu des fichiers.

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
│     ├── Dérivation Argon2id → KEK (Key Encryption Key)          │
│     ├── Génération AES-GCM 256 bits → MasterKey (RAM uniquement)│
│     ├── MasterKey chiffre fichiers (AES-GCM)                    │
│     └── KEK enveloppe MasterKey → EncryptedMasterKey            │
│                                                                  │
│  2. TRANSMISSION                                                 │
│     ├── Fichiers chiffrés + EncryptedMasterKey                  │
│     └── Backend reçoit UNIQUEMENT des données chiffrées         │
│                                                                  │
│  3. BACKEND (Go)                                                 │
│     ├── Stocke EncryptedMasterKey (inutilisable sans password)  │
│     ├── Stocke fichiers chiffrés dans S3/MinIO                  │
│     └── ❌ AUCUNE CLÉ DE DÉCHIFFREMENT                          │
│                                                                  │
│  4. RÉCUPÉRATION                                                 │
│     ├── Client récupère EncryptedMasterKey + fichiers chiffrés  │
│     ├── Dérive KEK depuis mot de passe                          │
│     ├── Déchiffre MasterKey                                     │
│     └── Déchiffre fichiers localement                           │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Garanties ZK

| Composant | Garantie |
|-----------|----------|
| **MasterKey** | Jamais envoyée au backend, stockée en RAM uniquement |
| **Fichiers** | Chiffrés en AES-GCM avant upload avec chunks de 10MB |
| **Partage** | RSA-OAEP 4096 bits pour chiffrer clés symétriques |
| **Métadonnées** | Noms de fichiers chiffrés, seules tailles/timestamps en clair |
| **Recovery** | Code de récupération avec hash SHA-256 + dérivation Argon2id |

---

## 🚀 Fonctionnalités

### Gestion de Fichiers
*   **Upload & Téléchargement** : Support des gros fichiers avec chiffrement à la volée
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
git clone https://github.com/votre-org/safercloud.git
cd safercloud
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
DATABASE_URL=postgresql://user:password@localhost:5432/safercloud

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
S3_BUCKET=safercloud-files

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

---

## 🔒 Sécurité

### Conformité

- ✅ **RGPD** - Architecture zero-knowledge conforme
- ✅ **OWASP Top 10** - Toutes vulnérabilités corrigées (voir [SECURITY_AUDIT_REPORT.md](SECURITY_AUDIT_REPORT.md))
- ✅ **Rate Limiting** - Protection DDoS avec `sync.Map`
- ✅ **CORS** - Origines configurables
- ✅ **CSP** - Content Security Policy strict
- ✅ **Timing Attack Mitigation** - Délais constants sur endpoints sensibles

### Cryptographie

| Algorithme | Usage | Paramètres |
|-----------|-------|------------|
| **Argon2id** | Dérivation de clé depuis password | 64MB RAM, 4 itérations |
| **AES-GCM** | Chiffrement symétrique fichiers | 256 bits, IV 12 octets |
| **RSA-OAEP** | Chiffrement asymétrique (partage) | 4096 bits, SHA-256 |
| **SHA-256** | Hachage codes de récupération | Standard |

### Logs Sécurisés

Tous les événements de sécurité sont enregistrés dans `backend/logs/security.log` avec structure:
- Timestamp, Type d'événement, UserID, IP, Succès/Échec
- **Jamais de clés cryptographiques**

---

## 📁 Structure du Projet

```
safercloud/
├── backend/                    # API Go
│   ├── handlers/               # Contrôleurs HTTP
│   │   ├── auth/               # Authentification, récupération
│   │   ├── files/              # Upload, download, delete
│   │   ├── folders/            # Gestion dossiers
│   │   ├── friends/            # Système d'amis
│   │   ├── shares/             # Partages public/privé
│   │   ├── users/              # Profils utilisateur
│   │   └── ws/                 # WebSocket temps réel
│   ├── middleware/             # Auth, rate limit, sécurité
│   ├── pkg/                    # Logique métier
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
│   │   │   ├── files.js        # Gestion fichiers
│   │   │   ├── friends.js      # Amis
│   │   │   ├── p2p.js          # WebRTC
│   │   │   └── websocket.js    # Client WS
│   │   ├── utils/              # Utilitaires crypto
│   │   │   ├── crypto.js       # Fonctions AES, RSA
│   │   │   ├── secureCrypto.js # XSS monitoring, rate limit
│   │   │   └── securityMonitoring.js
│   │   ├── views/              # Pages
│   │   ├── workers/            # Web Workers (chiffrement)
│   │   └── router/             # Vue Router
│   ├── public/
│   │   └── sw-crypto.js        # Service Worker (session)
│   └── index.html
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
```

### Frontend
```bash
cd frontend
npm run test          # Vitest
npm run test:coverage # Coverage
```

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

- **Issues**: [GitHub Issues](https://github.com/votre-org/safercloud/issues)
- **Documentation**: Ce README + sous-dossiers
- **Security**: Voir [SECURITY_AUDIT_REPORT.md](SECURITY_AUDIT_REPORT.md)

---

**Développé avec ❤️ et 🔐 pour la vie privée**
