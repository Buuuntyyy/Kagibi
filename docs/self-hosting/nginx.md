# Exposer Kagibi sur Internet (reverse proxy)

Kagibi ne gère pas le TLS lui-même : il faut un **reverse proxy** devant l'application pour l'exposer sur Internet via un nom de domaine, en HTTPS. Les points ci-dessous s'appliquent quel que soit le proxy choisi (nginx, Traefik, Caddy, Nginx Proxy Manager...).

## Ce qui doit être exposé derrière le proxy

| Service | Port interne (docker-compose) | Chemin | Nécessaire pour |
|---|---|---|---|
| Frontend | `${FRONTEND_PORT:-80}` | `/` | L'application elle-même |
| Backend API | `${BACKEND_PORT:-8080}` | `/api/v1/*` | L'application elle-même, **y compris** `/api/v1/ws` (WebSocket temps réel — notifications, présence, signalisation P2P) |
| Garage S3 (Option B uniquement) | `${GARAGE_S3_PORT:-3900}` | tout le chemin, sur un **sous-domaine séparé** (ex. `s3.votre-domaine.example`) | Les uploads/téléchargements directs entre le navigateur et le stockage — c'est la valeur de `GARAGE_PUBLIC_ENDPOINT` |

## Points obligatoires, quel que soit le proxy choisi

1. **TLS partout, y compris sur l'endpoint Garage (Option B).** Si l'app est servie en `https://`, un navigateur bloque par défaut tout appel vers une adresse en `http://` (contenu mixte) — `GARAGE_PUBLIC_ENDPOINT` doit donc lui aussi être en `https://` dès que l'app n'est plus en local. C'est différent de l'exemple `http://192.168.1.50:3900` donné plus haut, qui ne convient que pour un accès LAN sans nom de domaine.
2. **Support de l'upgrade WebSocket sur `/api/v1/ws`.** Un reverse proxy ne relaie pas les connexions WebSocket par défaut — il faut explicitement transmettre les en-têtes `Upgrade`/`Connection`. Sans ça, l'app continue de fonctionner (un mécanisme de repli par polling existe déjà côté frontend) mais avec une latence dégradée sur les notifications, la présence en ligne et le transfert P2P.
3. **Transmission de l'IP réelle du visiteur** (`X-Real-IP` / `X-Forwarded-For`) **et déclaration du proxy dans `.env`** :
   ```bash
   TRUSTED_PROXIES=127.0.0.1/32   # ou l'IP/CIDR du conteneur proxy si le proxy tourne lui-même dans Docker
   ```
   Sans cette variable, le backend ignore les en-têtes `X-Forwarded-For` (mesure de sécurité anti-usurpation déjà en place dans le code) et applique le rate-limiting à l'IP du proxy — donc à **tous les visiteurs confondus**, ce qui bloque le service dès qu'un seul utilisateur dépasse la limite.
4. **`ALLOWED_ORIGINS` et `VITE_API_URL` dans `.env` doivent correspondre exactement** au domaine public choisi (schéma + domaine, ex. `https://kagibi.votre-domaine.example`), pas à `localhost`.
5. **Timeout de connexion long sur `/api/v1/ws`** — le backend envoie un ping toutes les ~54 secondes pour garder la connexion vivante ; un timeout de proxy par défaut (souvent 60s) coupe la connexion à la limite. Fixez un timeout d'au moins quelques minutes sur cette route spécifiquement.
6. **Taille de requête suffisante pour les endpoints qui ne passent pas par les URLs présignées** (avatar, logo d'organisation) — une limite de 20 Mo convient largement. Le transfert des fichiers eux-mêmes ne passe **jamais** par le reverse proxy : le navigateur téléverse et télécharge directement vers S3/Garage via des URLs présignées, donc cette limite n'affecte pas les gros fichiers.

## Exemple complet — nginx

```nginx
map $http_upgrade $connection_upgrade {
    default upgrade;
    ''      close;
}

# Redirection HTTP → HTTPS
server {
    listen 80;
    server_name kagibi.votre-domaine.example;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name kagibi.votre-domaine.example;

    ssl_certificate     /etc/letsencrypt/live/kagibi.votre-domaine.example/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/kagibi.votre-domaine.example/privkey.pem;

    client_max_body_size 20m;

    # Frontend (SPA)
    location / {
        proxy_pass http://127.0.0.1:80;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Backend API
    location /api/v1/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # WebSocket temps réel — upgrade + timeout long (voir point 2 et 5 ci-dessus)
    location /api/v1/ws {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $connection_upgrade;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_read_timeout 3600s;
    }
}
```

Si vous êtes en **Option B**, ajoutez un second bloc `server` pour l'endpoint Garage, sur son propre sous-domaine :

```nginx
server {
    listen 443 ssl http2;
    server_name s3.kagibi.votre-domaine.example;

    ssl_certificate     /etc/letsencrypt/live/s3.kagibi.votre-domaine.example/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/s3.kagibi.votre-domaine.example/privkey.pem;

    client_max_body_size 0;   # pas de limite — les parts sont déjà découpées par le navigateur

    location / {
        proxy_pass http://127.0.0.1:3900;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```
Puis `GARAGE_PUBLIC_ENDPOINT=https://s3.kagibi.votre-domaine.example` dans `.env` (remplace l'exemple `http://192.168.1.50:3900` donné plus haut).

Ces blocs supposent que nginx tourne sur la même machine que Docker et que les ports sont publiés sur `127.0.0.1` côté hôte (`docker-compose.yaml`/`docker-compose.garage.yml` les publient sur `0.0.0.0` par défaut — si nginx tourne sur le même hôte, il est recommandé de restreindre ces `ports:` à `127.0.0.1:8080:8080` etc. pour que le backend ne soit plus joignable directement, en contournant le proxy). Si votre reverse proxy tourne lui-même dans un conteneur sur le même réseau Docker, remplacez `127.0.0.1:<port>` par le nom du service (`http://backend:8080`, `http://frontend:80`, `http://garage:3900`).

## Traefik (labels Docker Compose)

Traefik gère l'upgrade WebSocket automatiquement, sans configuration supplémentaire — seul le routage est à déclarer :

```yaml
  backend:
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.kagibi-api.rule=Host(`kagibi.votre-domaine.example`) && PathPrefix(`/api/v1`)"
      - "traefik.http.routers.kagibi-api.tls.certresolver=letsencrypt"
      - "traefik.http.services.kagibi-api.loadbalancer.server.port=8080"
  frontend:
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.kagibi-web.rule=Host(`kagibi.votre-domaine.example`)"
      - "traefik.http.routers.kagibi-web.tls.certresolver=letsencrypt"
      - "traefik.http.services.kagibi-web.loadbalancer.server.port=80"
```
Les points 1, 3, 4, 5, 6 ci-dessus restent applicables (TLS, `TRUSTED_PROXIES`, `ALLOWED_ORIGINS`/`VITE_API_URL`, timeouts, taille de requête) — seul le point 2 (upgrade WebSocket) est géré nativement par Traefik.

## Caddy

Caddy gère lui aussi le TLS (Let's Encrypt automatique) et l'upgrade WebSocket sans configuration explicite :

```
kagibi.votre-domaine.example {
    reverse_proxy /api/v1/* 127.0.0.1:8080
    reverse_proxy 127.0.0.1:80
}
```

## Portainer et Docker Swarm

Les deux options fonctionnent à l'identique via **Portainer en mode Standalone** (stack Docker Compose classique) — Portainer appelle le même moteur Compose en interne. Pour l'option B, ajouter les deux fichiers Compose (`docker-compose.yaml` et `docker-compose.garage.yml`) au stack, et réaliser les étapes 1-3 dans le répertoire de travail du stack sur l'hôte avant le premier démarrage.

**Ne fonctionne pas en mode Docker Swarm** (`docker stack deploy`) : Swarm ignore les conditions `depends_on` (notamment `condition: service_completed_successfully`), donc `garage-init` et `backend` démarreraient sans ordre garanti.
