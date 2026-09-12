# Réseau — vue d'ensemble

Cette page rassemble **tous** les ports et besoins réseau de la stack en un seul endroit. Pour le détail de mise en œuvre de chaque point, voir [Nginx](./nginx), [Stockage](./storage) et [Serveur TURN](./turn).

## Tous les ports, en un tableau

| Service | Port | Protocole | Exposition requise | Fichier compose |
|---|---|---|---|---|
| `frontend` | `${FRONTEND_PORT:-80}` | HTTP | Derrière un reverse proxy TLS (jamais directement sur Internet) | `docker-compose.yaml` |
| `backend` | `${BACKEND_PORT:-8080}` | HTTP + WebSocket (`/api/v1/ws`) | Derrière un reverse proxy TLS, avec upgrade WebSocket | `docker-compose.yaml` |
| `db` (PostgreSQL) | 5432 | TCP | **Jamais** — aucun `ports:` publié par défaut, joignable uniquement par les autres conteneurs du réseau Docker interne | `docker-compose.yaml` |
| `redis-db` | 6379 | TCP | **Jamais** — même principe, protégé en plus par mot de passe (`requirepass`) | `docker-compose.yaml` |
| `garage` — API S3 | `${GARAGE_S3_PORT:-3900}` | HTTP(S) | **Option B uniquement.** Doit être joignable directement par le **navigateur** (uploads/téléchargements présignés) — localhost, IP LAN, ou domaine derrière TLS | `docker-compose.garage.yml` |
| `garage` — RPC interne | 3901 | TCP | **Jamais** — non publié, réservé à la communication entre nœuds Garage (sans objet en mono-nœud) | `docker-compose.garage.yml` |
| `garage` — API admin | 3903 | HTTP | **Jamais** — non publié par défaut ; ne l'exposez que si vous administrez le cluster à distance, et alors uniquement derrière un accès restreint | `docker-compose.garage.yml` |
| `turn` (coturn) — signalisation | 3478 | UDP + TCP | **Add-on TURN uniquement.** Direct sur Internet (pas de reverse proxy possible pour UDP) | `docker-compose.turn.yml` |
| `turn` (coturn) — relais média | `${TURN_MIN_PORT:-49160}`–`${TURN_MAX_PORT:-49200}` | UDP | **Add-on TURN uniquement.** Direct sur Internet, toute la plage | `docker-compose.turn.yml` |

**Règle simple** : tout ce qui n'apparaît pas dans ce tableau comme "à exposer" doit rester injoignable depuis l'extérieur du réseau Docker interne. `db` et `redis-db` en particulier ne doivent **jamais** être publiés sur Internet, même temporairement pour du débogage (une ligne commentée existe dans `docker-compose.yaml` pour un accès local ponctuel, liée explicitement à `127.0.0.1` — ne la modifiez pas pour écouter sur `0.0.0.0`).

## Modèle réseau Docker

- Par défaut, tous les services d'un même fichier `docker-compose` partagent un réseau bridge interne créé automatiquement — c'est ce qui permet à `backend` de joindre `db:5432` ou `redis-db:6379` par leur nom de service, sans jamais passer par l'hôte.
- Seuls les services listés avec une exposition ci-dessus utilisent `ports:` pour publier un port vers l'hôte.
- Exception : le service `turn` utilise `network_mode: host` (pas de réseau bridge du tout) — nécessaire pour la plage de ports de relais UDP, voir [Serveur TURN](./turn) pour le détail. C'est pour cette raison que l'add-on TURN est **Linux uniquement**.

## DNS

| Besoin | Quand |
|---|---|
| Un enregistrement A/AAAA pointant vers ce serveur, pour le domaine principal | Toujours, dès que vous exposez Kagibi au-delà de `localhost` (voir [Nginx](./nginx)) |
| Un second enregistrement A/AAAA, sous-domaine dédié (ex. `s3.votre-domaine.example`) | Option B (Garage) uniquement, si vous choisissez d'exposer Garage via un domaine plutôt qu'une IP LAN ou `localhost` (voir [Stockage](./storage)) |
| Aucun — `TURN_EXTERNAL_IP` attend une IP, pas un nom de domaine | Add-on TURN |

## Cloudflare Tunnel

[Cloudflare Tunnel](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/) (`cloudflared`) établit une connexion sortante depuis votre serveur vers Cloudflare — aucun port entrant à ouvrir, pas d'IP publique fixe nécessaire. Pertinent selon le service :

| Service | Compatible Cloudflare Tunnel ? |
|---|---|
| `frontend` / `backend` (HTTP + WebSocket) | **Oui** — cas d'usage standard, y compris l'upgrade WebSocket sur `/api/v1/ws`. |
| `garage` — API S3 (Option B) | **Avec une réserve** — ça fonctionne, mais le plan Free/Pro de Cloudflare plafonne à **100 Mo par requête** tout le trafic qui transite par le tunnel. Au-delà, les gros fichiers échoueraient. Si vous avez besoin de fichiers volumineux, n'exposez pas Garage via le tunnel — utilisez une IP LAN/publique directe ou votre propre reverse proxy TLS à la place (voir [Stockage](./storage)). |
| `turn` (coturn) | **Non — même sur les plans payants standards.** Cloudflare Tunnel ne relaie que du TCP/HTTP/WebSocket ; le relais média de TURN, lui, est en UDP, que le Tunnel classique ne prend pas en charge. Seul [Cloudflare Spectrum](https://developers.cloudflare.com/spectrum/) (offre **Enterprise**, payante) proxifie de l'UDP arbitraire. En pratique : continuez à exposer coturn directement (port-forwarding classique, voir le pare-feu ci-dessous) — le Tunnel ne résout pas ce cas. |

**Alternative pour éviter d'auto-héberger coturn** : Cloudflare propose son propre service TURN managé ([Cloudflare Realtime TURN](https://developers.cloudflare.com/realtime/turn/)) — un produit distinct, pas un tunnel vers votre coturn, mais un serveur TURN hébergé par Cloudflare que vous configurez directement dans `TURN_SERVER_URL`/`TURN_USERNAME`/`TURN_CREDENTIAL`. Si vous ne voulez pas gérer l'exposition réseau de coturn, c'est l'option à regarder plutôt que de forcer le passage par un Tunnel.

## Pare-feu — synthèse

- **Toujours** : 443 (et 80 pour la redirection HTTPS) ouverts sur votre reverse proxy.
- **Option B (Garage)** : le port `GARAGE_S3_PORT` (défaut 3900) ouvert si Garage est exposé sur une IP LAN/publique plutôt que `localhost`.
- **Add-on TURN** : 3478 (UDP+TCP) et toute la plage `TURN_MIN_PORT`–`TURN_MAX_PORT` en UDP.
- **Jamais** : 5432 (Postgres), 6379 (Redis), 3901/3903 (Garage RPC/admin).
