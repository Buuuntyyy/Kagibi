# SaferCloud
## Suivi de projet
[Notion du projet](https://www.notion.so/Projet-Cloud-Files-29273e3ff6e780f986e6eeb280237ff3?source=copy_link)

## Déploiement
### Pré-requis:
- Docker

### Démarche:
#### Docker
Déployer la base de données (postgresql) et le cache (redis)
```
docker-compose up -d
```
#### Backend
Démarrage du serveur Go
```
cd backend
go mod tidy
go run main.go
```

#### Frontend
Démarrage du serveur frontend
```
cd frontend
npm install
npm run dev
```

### Utilisation
1. Créer un utilisateur afin d'accéder aux routes protégées
2. Importer des fichiers
3. Créer des répertoires

## Frontend
Vue3 :
Authentification (inscription/connexion, JWT).  
Upload/téléchargement de fichiers.  
Gestion des dossiers (création, suppression, renommage).  
Affichage des fichiers (liste, grille, prévisualisation).  
Partage de fichiers (liens publics, permissions).  
Recherche de fichiers.  

### Architecture
Framework : Vue 3 (Composition API).  
Router : Vue Router pour la navigation.  
State Management : Pinia (remplace Vuex).  
UI Components : PrimeVue, Quasar, ou Vuetify.  
HTTP Client : Axios pour les requêtes API.  
Authentification : Vueuse pour gérer les tokens JWT.  
Build Tool : Vite (plus rapide que Webpack).  
Tests : Vitest ou Jest.

### Initialiser Vue
```
cd mon-drive/frontend
npm create vue@latest
# Choisir : TypeScript, JSX, Vue Router, Pinia, ESLint, Vitest
cd frontend
npm install axios primevue primeicons pinia vue-router
npm install -D @vitejs/plugin-vue
```

## Backend
Golang :  
API REST pour gérer les fichiers et dossiers.  
Stockage des fichiers (local ou cloud : S3, Google Cloud Storage).  
Gestion des utilisateurs et authentification (JWT, OAuth).  
Logique de partage et permissions.  
Base de données pour les métadonnées (PostgreSQL, MongoDB).  

### Architecture
Framework Web : Gin ou Fiber (inspiré d’Express).  
Base de données :  

SQL : PostgreSQL (avec GORM ou sqlx).  
NoSQL : MongoDB (avec mongo-go-driver).  


Stockage de fichiers :  

Local (pour le développement).  
Cloud : AWS S3, Google Cloud Storage (avec AWS SDK for Go).  


Authentification : JWT ou OAuth2.  
Validation : Ozzo Validation ou Validator.  
Tests : Testify ou Ginkgo.

### Initialisation backend
```
cd mon-drive/backend
go mod init github.com/ton-username/mon-drive
go get -u github.com/gin-gonic/gin
go get -u gorm.io/gorm
go get -u gorm.io/driver/postgres
go get -u github.com/golang-jwt/jwt/v5
```

### Créer le docker-compose.yaml de postgresql
```
version: "3.8"
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
      POSTGRES_DB: mon_drive
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
  redis:
    image: redis:7
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
volumes:
  postgres_data:
  redis_data:
```
Lancer le docker avec : docker-compose up -d (dans le répertoire du fichier docker-compose.yaml

### Base de données
#### Installation de Bun (équivalent prisma, mais pour Go)
Dans le dossier backend:
```
go get -u github.com/uptrace/bun
go get -u github.com/uptrace/bun/dialect/pgdialect
go get -u github.com/uptrace/bun/driver/pgdriver

```
Créer le dossier /backend/internal puis créer la connexion postgresql
```
// internal/database.go
package internal

import (
	"database/sql"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func NewDB() *bun.DB {
	// Remplace les valeurs par celles de ton docker-compose ou de ta configuration locale
	dsn := "postgres://user:password@localhost:5432/mon_drive?sslmode=disable"

	// Ouvre la connexion SQL
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

	// Crée une instance Bun
	db := bun.NewDB(sqldb, pgdialect.New())

	return db
}
```
Créer le fichier /backend/internal/models.go et configurer la structure de la bdd
```
// internal/models.go
package internal

import "time"

type User struct {
	ID        int64     `bun:"id,pk,autoincrement"`
	Name      string    `bun:"name,notnull"`
	Email     string    `bun:"email,unique,notnull"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`
}

type File struct {
	ID        int64     `bun:"id,pk,autoincrement"`
	Name      string    `bun:"name,notnull"`
	Path      string    `bun:"path,notnull"`
	Size      int64     `bun:"size,notnull"`
	UserID    int64     `bun:"user_id,notnull"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
}
```
Bun ne gère pas les migrations automatiquement comme GORM, mais tu peux utiliser son API pour créer les tables. Crée le fichier /backend/internal/migrate.go :
```
// internal/migrate.go
package internal

import (
	"context"
	"fmt"
)

func Migrate(db *bun.DB) error {
	ctx := context.Background()

	// Crée les tables si elles n'existent pas
	models := []interface{}{(*User)(nil), (*File)(nil)}

	for _, model := range models {
		_, err := db.NewSelect().Model(model).Exec(ctx)
		if err != nil {
			_, err = db.NewCreateTable().Model(model).IfNotExists().Exec(ctx)
			if err != nil {
				return fmt.Errorf("failed to create table: %w", err)
			}
			fmt.Printf("Table created: %T\n", model)
		}
	}

	return nil
}
```
Pour effectuer les migrations, il faut ajouter le code suivant dans main.go:
```
// main.go
package main

import (
	"log"
	"ton-projet/backend/internal"
)

func main() {
	db := internal.NewDB()

	// Exécute les migrations
	err := internal.Migrate(db)
	if err != nil {
		log.Fatalf("Failed to migrate: %v", err)
	}

	log.Println("Migrations executed successfully!")
}
```

Schémas :  
```
users (id, email, password_hash, created_at).  
files (id, user_id, name, path, size, mime_type, created_at).  
folders (id, user_id, name, path, created_at).  
shares (id, file_id, user_id, permission, created_at).
```

### Stockage des fichiers
Local : Dossier /uploads dans le projet (pour le développement).  
Cloud : AWS S3 ou Google Cloud Storage (pour la production).

## Structure du workspace (visual studio code)
```
mon-drive/
├── frontend/          # Vue 3
│   ├── public/
│   ├── src/
│   │   ├── assets/
│   │   ├── components/
│   │   ├── stores/    # Pinia
│   │   ├── views/
│   │   ├── router/
│   │   ├── App.vue
│   │   └── main.js
│   ├── package.json
│   ├── vite.config.js
│   └── ...
├── backend/           # Golang
│   ├── cmd/
│   │   └── server/
│   ├── internal/
│   │   ├── controllers/
│   │   ├── models/
│   │   ├── routes/
│   │   ├── services/
│   │   └── utils/
│   ├── pkg/
│   ├── go.mod
│   ├── go.sum
│   └── main.go
├── docker-compose.yml # Pour les services (PostgreSQL, Redis, etc.)
├── .gitignore
└── README.md
```

## CI/CD et Tests

Le projet intègre un pipeline d'Intégration Continue et de Déploiement Continu (CI/CD) via **GitHub Actions**, ainsi qu'une suite de tests unitaires pour le backend et le frontend.

### Pipeline GitHub Actions
Le workflow est défini dans `.github/workflows/ci.yml` et s'exécute à chaque `push` ou `pull_request` sur les branches principales. Il effectue les actions suivantes :

*   **Backend (Go)** :
    *   Installation de Go (v1.21).
    *   Vérification du formatage du code (`go fmt`).
    *   Exécution des tests unitaires (`go test`).
    *   Vérification de la compilation (`go build`).
*   **Frontend (Vue.js)** :
    *   Installation de Node.js (v22).
    *   Installation des dépendances (`npm ci`).
    *   Exécution des tests unitaires (`vitest`).
    *   Build de l'application (`npm run build`).

### Tests Backend (Go)
Les tests backend couvrent la logique métier, la sécurité et les interactions avec la base de données (mockée).

*   **Technologies** : `testing` (std lib), `testify/assert`, `go-sqlmock` (pour mocker PostgreSQL), `redismock` (pour mocker Redis).
*   **Couverture** :
    *   **Sécurité** : Middleware d'authentification, Rate Limiting, Protection contre le Path Traversal (`utils.SecureJoin`).
    *   **Fonctionnalités** : Création de dossiers, Renommage de fichiers (avec gestion des transactions et conflits).

**Lancer les tests backend en local :**
```bash
cd backend
go test -v ./...
```

### Tests Frontend (Vue.js)
Les tests frontend vérifient le bon rendu des composants et la logique d'interface.

*   **Technologies** : `Vitest`, `@vue/test-utils`, `jsdom`.
*   **Configuration** : Environnement simulé via `jsdom` pour tester les composants Vue sans navigateur.

**Lancer les tests frontend en local :**
```bash
cd frontend
npm run test
```
