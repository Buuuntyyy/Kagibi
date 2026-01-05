# SaferCloud

SaferCloud est une solution de stockage cloud sécurisée, open-source et auto-hébergeable, inspirée par Google Drive. Elle met l'accent sur la confidentialité des données grâce à une architecture "Zero-Knowledge" (chiffrement de bout en bout).

## 🚀 Fonctionnalités

### Gestion de Fichiers
*   **Upload & Téléchargement** : Support des gros fichiers avec chiffrement à la volée.
*   **Organisation** : Création de dossiers, déplacement, renommage et suppression.
*   **Navigation** : Interface fluide type "Google Drive" avec vue liste et grille.
*   **Recherche** : Barre de recherche rapide pour filtrer les fichiers.
*   **Tags** : Système de tags pour classer les fichiers.

### Partage & Collaboration
*   **Partage Utilisateur** : Partage sécurisé de fichiers avec d'autres utilisateurs de la plateforme.
*   **Liens Publics** : Génération de liens de partage accessibles publiquement (avec expiration optionnelle).
*   **Permissions** : Gestion fine des droits d'accès.

### Interface Utilisateur (UI/UX)
*   **Design Moderne** : Interface épurée avec icônes SVG dynamiques selon le type de fichier.
*   **Thème Sombre/Clair** : Support natif du Dark Mode via CSS variables.
*   **Responsive** : Adapté aux écrans de bureau et tablettes.
*   **Feedback Visuel** : Dialogues de confirmation personnalisés, notifications toast, et indicateurs de chargement.

### Sécurité
*   **Authentification** : Inscription, Connexion, et Récupération de compte sécurisée.
*   **Chiffrement** : Chiffrement côté client (voir section Sécurité).

---

## 🔒 Sécurité & Chiffrement (Zero-Knowledge)

SaferCloud utilise une architecture **Zero-Knowledge**, ce qui signifie que le serveur ne connaît jamais votre mot de passe en clair ni le contenu de vos fichiers. Tout est chiffré dans votre navigateur avant d'être envoyé.

### 1. Gestion des Clés
*   **Clé Maître (Master Key)** : Chaque utilisateur possède une clé maître AES-256 générée aléatoirement dans le navigateur.
*   **Dérivation de Clé (Argon2id)** : Votre mot de passe dérive une clé de chiffrement (KEK) via l'algorithme **Argon2id** (paramètres robustes : 64MB RAM, 4 passes).
*   **Stockage Sécurisé** : La clé maître est chiffrée avec votre KEK et stockée sur le serveur (`encrypted_master_key`). Le serveur ne peut pas la déchiffrer sans votre mot de passe.

### 2. Chiffrement des Fichiers
*   **Clé de Fichier** : Chaque fichier possède sa propre clé de chiffrement unique (AES-256).
*   **Chiffrement AES-GCM** : Le contenu du fichier est chiffré par blocs (chunks de 1MB) utilisant **AES-GCM** via l'API Web Crypto du navigateur.
*   **Web Workers** : Le chiffrement/déchiffrement s'effectue dans des threads séparés (Web Workers) pour ne pas bloquer l'interface.

### 3. Partage Sécurisé
*   Lorsqu'un fichier est partagé, la clé du fichier est déchiffrée par le propriétaire, puis rechiffrée avec la clé publique (ou clé maître partagée) du destinataire. Cela garantit que seul le destinataire légitime peut accéder au contenu.

---

## 🛠️ Stack Technique

### Frontend
*   **Framework** : Vue 3 (Composition API)
*   **Build Tool** : Vite
*   **State Management** : Pinia
*   **Styling** : CSS natif avec Variables (Theming), SVG Icons
*   **Crypto** : Web Crypto API, Libsodium (via `libsodium-wrappers-sumo`)

### Backend
*   **Langage** : Go (Golang)
*   **Framework Web** : Gin
*   **Base de Données** : PostgreSQL
*   **ORM** : Bun (Go module)
*   **Cache** : Redis (pour les sessions et limites de débit)
*   **Stockage** : Compatible S3 (MinIO, AWS S3, OVH Object Storage) ou Local

### Infrastructure
*   **Conteneurisation** : Docker & Docker Compose
*   **CI/CD** : GitHub Actions

---

## 📦 Installation & Démarrage

### Pré-requis
*   Docker & Docker Compose
*   Go 1.21+ (pour le développement local)
*   Node.js 20+ (pour le développement local)

### 1. Démarrer l'infrastructure (Base de données & Redis)
```bash
docker-compose up -d
```

### 2. Démarrer le Backend
```bash
cd backend
go mod tidy
go run main.go
```
Le serveur démarrera sur `http://localhost:8080`.

### 3. Démarrer le Frontend
```bash
cd frontend
npm install
npm run dev
```
L'application sera accessible sur `http://localhost:5173`.

---

## 🧪 Tests

Le projet inclut une suite de tests complète.

### Backend
```bash
cd backend
go test -v ./...
```

### Frontend
```bash
cd frontend
npm run test
```

---

## 📂 Structure du Projet

```
SaferCloud/
├── backend/                # API Go
│   ├── handlers/           # Contrôleurs API (Auth, Files, Shares...)
│   ├── middleware/         # Auth, RateLimit, Security
│   ├── pkg/                # Packages utilitaires (Database, S3, Models)
│   └── main.go             # Point d'entrée
├── frontend/               # Application Vue.js
│   ├── src/
│   │   ├── components/     # Composants UI (Dialogs, FileList...)
│   │   ├── stores/         # Stores Pinia (Auth, Files...)
│   │   ├── utils/          # Logique Crypto & Helpers
│   │   ├── views/          # Pages (Login, Dashboard...)
│   │   └── workers/        # Web Workers pour le chiffrement
│   └── ...
├── docker/                 # Configuration Docker
└── db_data/                # Persistance DB (ignoré par git)
```
