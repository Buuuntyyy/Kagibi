# Docker - SaferCloud

Ce répertoire contient les fichiers Docker pour construire les images du backend et du frontend de SaferCloud.

## Objectif

Ce dossier sert uniquement à construire les images Docker optimisées pour le backend Go et le frontend Vue.js. Les images construites peuvent ensuite être déployées dans différents environnements (développement, staging, production).

## Structure

```
docker/
└── deployment/
    ├── backend/
    │   └── Dockerfile       # Image multi-stage pour le backend Go
    |   └── grafana
    |       └── provisioning
    |           └── dashboards
    |           |   └── dashboard.yaml
    |           └── datasources
    |               └── prometheus.yaml
    ├── frontend/
    │   └── Dockerfile       # Image multi-stage pour le frontend Vue.js + nginx
    └── docker-compose.prod.yml

```

## Images disponibles

### Backend (API Go)

**Dockerfile:** [deployment/backend/Dockerfile](deployment/backend/Dockerfile)

**Caractéristiques:**
- Build multi-stage avec Go 1.25
- Image finale basée sur Alpine Linux (légère)
- Utilisateur non-root pour la sécurité
- Binary optimisé avec ldflags
- Inclut les migrations de base de données
- Expose les ports 8080 (API) et 9090 (métriques Prometheus)

**Construction:**

```bash
# Depuis la racine du projet
docker build -f docker/deployment/backend/Dockerfile -t safercloud-backend:latest ./backend

# Ou depuis le dossier backend
cd backend
docker build -f ../docker/deployment/backend/Dockerfile -t safercloud-backend:latest .
```

**Tags recommandés:**

```bash
# Version spécifique
docker build -f docker/deployment/backend/Dockerfile -t safercloud-backend:1.0.0 ./backend

# Latest
docker build -f docker/deployment/backend/Dockerfile -t safercloud-backend:latest ./backend
```

### Frontend (Vue.js + nginx)

**Dockerfile:** [deployment/frontend/Dockerfile](deployment/frontend/Dockerfile)

**Caractéristiques:**
- Build multi-stage avec Node.js 22
- Serveur nginx 1.27 pour production
- Build-time variables d'environnement (Vite)
- Configuration nginx personnalisée
- Image finale optimisée et légère

**Construction:**

```bash
# Depuis la racine du projet
docker build -f docker/deployment/frontend/Dockerfile \
  --build-arg VITE_API_URL=https://api.example.com \
  --build-arg VITE_SUPABASE_URL=https://project.supabase.co \
  --build-arg VITE_SUPABASE_KEY=your-anon-key \
  -t safercloud-frontend:latest \
  ./frontend

# Ou depuis le dossier frontend
cd frontend
docker build -f ../docker/deployment/frontend/Dockerfile \
  --build-arg VITE_API_URL=https://api.example.com \
  --build-arg VITE_SUPABASE_URL=https://project.supabase.co \
  --build-arg VITE_SUPABASE_KEY=your-anon-key \
  -t safercloud-frontend:latest \
  .
```

**Build arguments requis:**
- `VITE_API_URL`: URL de l'API backend
- `VITE_SUPABASE_URL`: URL du projet Supabase
- `VITE_SUPABASE_KEY`: Clé anonyme Supabase

## Construction des deux images

Pour construire les deux images en une seule commande :

```bash
# Backend
docker build -f docker/deployment/backend/Dockerfile -t safercloud-backend:latest ./backend

# Frontend
docker build -f docker/deployment/frontend/Dockerfile \
  --build-arg VITE_API_URL=${VITE_API_URL} \
  --build-arg VITE_SUPABASE_URL=${VITE_SUPABASE_URL} \
  --build-arg VITE_SUPABASE_KEY=${VITE_SUPABASE_KEY} \
  -t safercloud-frontend:latest \
  ./frontend
```

## Test des images en local

### Tester le backend :

```bash
docker run --rm -p 8080:8080 -p 9090:9090 \
  -e DATABASE_URL="postgres://user:pass@host:5432/db" \
  -e SUPABASE_URL="https://project.supabase.co" \
  -e SUPABASE_JWT_SECRET="your-jwt-secret" \
  safercloud-backend:latest
```

### Tester le frontend :

```bash
docker run --rm -p 80:80 safercloud-frontend:latest
```

Accéder à http://localhost

## Push vers un registry

### Docker Hub

```bash
# Tag
docker tag safercloud-backend:latest username/safercloud-backend:latest
docker tag safercloud-frontend:latest username/safercloud-frontend:latest

# Push
docker push username/safercloud-backend:latest
docker push username/safercloud-frontend:latest
```

### GitHub Container Registry

```bash
# Login
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# Tag
docker tag safercloud-backend:latest ghcr.io/username/safercloud-backend:latest
docker tag safercloud-frontend:latest ghcr.io/username/safercloud-frontend:latest

# Push
docker push ghcr.io/username/safercloud-backend:latest
docker push ghcr.io/username/safercloud-frontend:latest
```

### Registry privé

```bash
# Tag
docker tag safercloud-backend:latest registry.example.com/safercloud-backend:latest
docker tag safercloud-frontend:latest registry.example.com/safercloud-frontend:latest

# Push
docker push registry.example.com/safercloud-backend:latest
docker push registry.example.com/safercloud-frontend:latest
```

## Optimisations

Les Dockerfiles utilisent des pratiques optimales :

1. **Multi-stage builds** : séparation build/runtime pour réduire la taille
2. **Layer caching** : copie des fichiers de dépendances avant le code source
3. **Images Alpine** : images de base légères
4. **Non-root user** : sécurité renforcée (backend)
5. **Build optimisé** : flags de compilation pour réduire la taille du binary
6. **Health checks** : vérification de l'état des conteneurs

## Taille des images

Tailles approximatives après construction :

- **Backend** : ~20-30 MB (grâce à Alpine et binary optimisé)
- **Frontend** : ~30-40 MB (nginx + assets statiques)

## Variables d'environnement

### Backend (runtime)

Consultez la documentation du backend pour la liste complète. Principales variables :

- `DATABASE_URL`: Connexion PostgreSQL
- `REDIS_URL`: Connexion Redis
- `SUPABASE_URL`: URL Supabase
- `SUPABASE_JWT_SECRET`: Secret JWT Supabase
- `S3_ENDPOINT`: Endpoint S3
- `S3_BUCKET`: Nom du bucket
- `AWS_ACCESS_KEY_ID`: Clé d'accès AWS
- `AWS_SECRET_ACCESS_KEY`: Secret AWS

### Frontend (build-time)

Variables passées via `--build-arg` :

- `VITE_API_URL`: URL de l'API
- `VITE_SUPABASE_URL`: URL Supabase
- `VITE_SUPABASE_KEY`: Clé anonyme Supabase

## Notes importantes

1. Les Dockerfiles sont optimisés pour la production
2. Le contexte de build doit être le dossier `backend/` ou `frontend/`, pas la racine
3. Les variables d'environnement du frontend sont intégrées au build (Vite)
4. Le backend nécessite des variables d'environnement au runtime
5. Les images ne contiennent pas de données sensibles
6. Utilisez toujours des tags versionnés en production

## Commandes utiles

```bash
# Lister les images
docker images | grep safercloud

# Voir la taille d'une image
docker image inspect safercloud-backend:latest -f '{{.Size}}' | numfmt --to=iec

# Supprimer les images
docker rmi safercloud-backend:latest
docker rmi safercloud-frontend:latest

# Nettoyer les images intermédiaires
docker image prune -f

# Inspecter les layers d'une image
docker history safercloud-backend:latest
```

## Ressources

- [Docker Multi-stage builds](https://docs.docker.com/build/building/multi-stage/)
- [Docker Best practices](https://docs.docker.com/develop/dev-best-practices/)
- [Alpine Linux](https://alpinelinux.org/)
- [nginx Docker](https://hub.docker.com/_/nginx)
