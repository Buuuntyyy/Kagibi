# SaferCloud

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
