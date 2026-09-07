# Exposing Kagibi to the Internet (reverse proxy)

Kagibi doesn't terminate TLS itself: you need a **reverse proxy** in front of it to expose it on the Internet over a domain name, in HTTPS. The points below apply whichever proxy you choose (nginx, Traefik, Caddy, Nginx Proxy Manager...).

## What needs to sit behind the proxy

| Service | Internal port (docker-compose) | Path | Needed for |
|---|---|---|---|
| Frontend | `${FRONTEND_PORT:-80}` | `/` | The app itself |
| Backend API | `${BACKEND_PORT:-8080}` | `/api/v1/*` | The app itself, **including** `/api/v1/ws` (real-time WebSocket — notifications, presence, P2P signaling) |
| Garage S3 (Option B only) | `${GARAGE_S3_PORT:-3900}` | the whole path, on a **separate subdomain** (e.g. `s3.your-domain.example`) | Direct browser ↔ storage uploads/downloads — this is the value of `GARAGE_PUBLIC_ENDPOINT` |

## Required points, whichever proxy you pick

1. **TLS everywhere, including on the Garage endpoint (Option B).** If the app is served over `https://`, a browser blocks by default any call to an `http://` address (mixed content) — so `GARAGE_PUBLIC_ENDPOINT` must also be `https://` as soon as the app isn't local anymore. This differs from the `http://192.168.1.50:3900` example given earlier, which only suits a LAN-only setup without a domain name.
2. **WebSocket upgrade support on `/api/v1/ws`.** A reverse proxy doesn't forward WebSocket connections by default — the `Upgrade`/`Connection` headers must be explicitly passed through. Without this, the app still works (a polling fallback already exists on the frontend) but with degraded latency on notifications, online presence, and P2P transfer.
3. **Forward the visitor's real IP** (`X-Real-IP` / `X-Forwarded-For`) **and declare the proxy in `.env`**:
   ```bash
   TRUSTED_PROXIES=127.0.0.1/32   # or the proxy container's IP/CIDR if the proxy itself runs in Docker
   ```
   Without this variable, the backend ignores `X-Forwarded-For` headers (an anti-spoofing safeguard already in place in the code) and applies rate-limiting to the proxy's own IP — meaning **every visitor combined**, which locks out the whole service as soon as a single user hits the limit.
4. **`ALLOWED_ORIGINS` and `VITE_API_URL` in `.env` must exactly match** the public domain you chose (scheme + domain, e.g. `https://kagibi.your-domain.example`), not `localhost`.
5. **Long connection timeout on `/api/v1/ws`** — the backend sends a ping roughly every ~54 seconds to keep the connection alive; a default proxy timeout (often 60s) cuts it close. Set a timeout of at least a few minutes specifically on this route.
6. **Sufficient request size for the endpoints that don't go through presigned URLs** (avatar, organization logo) — a 20 MB limit is more than enough. The actual file transfers never go through the reverse proxy at all: the browser uploads/downloads directly to S3/Garage via presigned URLs, so this limit has no effect on large files.

## Full example — nginx

```nginx
map $http_upgrade $connection_upgrade {
    default upgrade;
    ''      close;
}

# HTTP → HTTPS redirect
server {
    listen 80;
    server_name kagibi.your-domain.example;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name kagibi.your-domain.example;

    ssl_certificate     /etc/letsencrypt/live/kagibi.your-domain.example/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/kagibi.your-domain.example/privkey.pem;

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

    # Real-time WebSocket — upgrade + long timeout (see points 2 and 5 above)
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

If you're on **Option B**, add a second `server` block for the Garage endpoint, on its own subdomain:

```nginx
server {
    listen 443 ssl http2;
    server_name s3.kagibi.your-domain.example;

    ssl_certificate     /etc/letsencrypt/live/s3.kagibi.your-domain.example/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/s3.kagibi.your-domain.example/privkey.pem;

    client_max_body_size 0;   # no limit — the browser already splits uploads into parts

    location / {
        proxy_pass http://127.0.0.1:3900;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```
Then set `GARAGE_PUBLIC_ENDPOINT=https://s3.kagibi.your-domain.example` in `.env` (replacing the `http://192.168.1.50:3900` example given earlier).

These blocks assume nginx runs on the same machine as Docker and that the ports are published on `127.0.0.1` on the host side (`docker-compose.yaml`/`docker-compose.garage.yml` publish them on `0.0.0.0` by default — if nginx runs on the same host, it's recommended to restrict these `ports:` to `127.0.0.1:8080:8080` etc. so the backend is no longer directly reachable, bypassing the proxy). If your reverse proxy itself runs in a container on the same Docker network, replace `127.0.0.1:<port>` with the service name instead (`http://backend:8080`, `http://frontend:80`, `http://garage:3900`).

## Traefik (Docker Compose labels)

Traefik handles the WebSocket upgrade automatically, no extra configuration needed — only the routing needs declaring:

```yaml
  backend:
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.kagibi-api.rule=Host(`kagibi.your-domain.example`) && PathPrefix(`/api/v1`)"
      - "traefik.http.routers.kagibi-api.tls.certresolver=letsencrypt"
      - "traefik.http.services.kagibi-api.loadbalancer.server.port=8080"
  frontend:
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.kagibi-web.rule=Host(`kagibi.your-domain.example`)"
      - "traefik.http.routers.kagibi-web.tls.certresolver=letsencrypt"
      - "traefik.http.services.kagibi-web.loadbalancer.server.port=80"
```
Points 1, 3, 4, 5, 6 above still apply (TLS, `TRUSTED_PROXIES`, `ALLOWED_ORIGINS`/`VITE_API_URL`, timeouts, request size) — only point 2 (WebSocket upgrade) is handled natively by Traefik.

## Caddy

Caddy also handles TLS (automatic Let's Encrypt) and the WebSocket upgrade without explicit configuration:

```
kagibi.your-domain.example {
    reverse_proxy /api/v1/* 127.0.0.1:8080
    reverse_proxy 127.0.0.1:80
}
```

## Portainer and Docker Swarm

Both options work identically via **Portainer in Standalone mode** (a plain Docker Compose stack) — Portainer calls the same Compose engine internally. For option B, add both Compose files (`docker-compose.yaml` and `docker-compose.garage.yml`) to the stack, and perform steps 1-3 inside the stack's working directory on the host before the first start.

**Does not work under Docker Swarm mode** (`docker stack deploy`): Swarm ignores `depends_on` conditions, so `garage-init` and `backend` would start with no guaranteed order. This isn't a Garage-specific limitation — `docker-compose.yaml` already uses `build:` for `backend`/`frontend`, a directive Swarm doesn't support at all.
