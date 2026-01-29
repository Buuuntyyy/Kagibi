# Tests CI/CD - SaferCloud

## 📋 Vue d'ensemble

Le pipeline CI/CD de SaferCloud a été enrichi avec des tests complets pour vérifier toutes les fonctionnalités du site.

## 🧪 Jobs de Tests

### 1. **backend-test** (Tests Backend)
- ✅ Vérification du formatage du code Go (`gofmt`)
- ✅ Tests unitaires (utils, middleware, handlers, models)
- ✅ Tests avec couverture de code et race detection
- ✅ Upload des métriques de couverture vers Codecov
- ✅ Build du backend

**Services**: PostgreSQL 16, Redis 7

### 2. **backend-integration** (Tests d'intégration Backend)
- ✅ Tests d'intégration avec base de données réelle
- ✅ Tests des interactions entre composants
- ✅ Vérification de la logique métier end-to-end

**Services**: PostgreSQL 16, Redis 7

### 3. **frontend-test** (Tests Frontend)
- ✅ Linting du code (si configuré)
- ✅ Tests unitaires des composants Vue.js
- ✅ Tests avec couverture de code
- ✅ Upload des métriques de couverture vers Codecov
- ✅ Build de production
- ✅ Vérification de la taille du bundle

### 4. **e2e-tests** (Tests End-to-End)
- ✅ Démarrage complet du backend
- ✅ Build et déploiement du frontend
- ✅ Health check de l'API
- ✅ Tests de l'application complète

**Dépendances**: backend-test, frontend-test

### 5. **security-audit** (Audit de Sécurité)
- ✅ Scan de sécurité Go avec `gosec`
- ✅ Audit des vulnérabilités npm
- ✅ Détection de mots de passe en dur
- ✅ Détection de clés API exposées

### 6. **code-quality** (Qualité du Code)
- ✅ Vérification avec `go vet`
- ✅ Analyse statique avec `staticcheck`
- ✅ Détection des TODO/FIXME

## 📁 Nouveaux Fichiers de Tests

### Backend
- `backend/handlers/auth/login_test.go` - Tests du login
- `backend/handlers/auth/register_test.go` - Tests de l'inscription
- `backend/pkg/models_test.go` - Tests des modèles de données

### Frontend
- `frontend/src/stores/__tests__/store.spec.js` - Tests Pinia
- `frontend/src/utils/__tests__/api.spec.js` - Tests utilitaires API
- `frontend/src/router/__tests__/router.spec.js` - Tests du routeur

## 🛠️ Scripts Utiles

### Backend
```bash
# Tests unitaires
go test -v ./...

# Tests avec couverture
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Tests d'intégration
go test -v ./... -tags=integration
```

### Frontend
```bash
# Tests unitaires
npm run test

# Tests avec couverture
npm run test -- --coverage

# Build
npm run build
```

## 🔒 Sécurité

Le pipeline vérifie automatiquement:
- Vulnérabilités des dépendances
- Secrets/passwords en dur
- Clés API exposées
- Failles de sécurité dans le code

## 📊 Couverture de Code

Les rapports de couverture sont automatiquement envoyés à Codecov pour:
- Backend (Go)
- Frontend (Vue.js)

## 🚀 Déclenchement

Les tests sont exécutés automatiquement sur:
- Chaque `push` sur les branches `main` ou `master`
- Chaque `pull request` vers ces branches

## ⚙️ Configuration Locale

Pour exécuter les tests d'intégration localement:

```bash
# Démarrer les services de test
docker-compose -f docker-compose.test.yml up -d

# Exécuter les tests
./scripts/integration-tests.sh

# Nettoyer
docker-compose -f docker-compose.test.yml down
```

## 📝 Notes

- Les tests E2E nécessitent que le backend et le frontend soient fonctionnels
- Les services PostgreSQL et Redis sont provisionnés automatiquement dans GitHub Actions
- La couverture de code cible minimum recommandée : 70%
