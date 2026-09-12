# Network — overview

This page brings together **every** port and networking requirement of the stack in one place. For implementation details on each point, see [Nginx](./nginx), [Storage](./storage), and [TURN server](./turn).

## All ports, in one table

| Service | Port | Protocol | Required exposure | Compose file |
|---|---|---|---|---|
| `frontend` | `${FRONTEND_PORT:-80}` | HTTP | Behind a TLS reverse proxy (never directly on the Internet) | `docker-compose.yaml` |
| `backend` | `${BACKEND_PORT:-8080}` | HTTP + WebSocket (`/api/v1/ws`) | Behind a TLS reverse proxy, with WebSocket upgrade | `docker-compose.yaml` |
| `db` (PostgreSQL) | 5432 | TCP | **Never** — no `ports:` published by default, reachable only by other containers on the internal Docker network | `docker-compose.yaml` |
| `redis-db` | 6379 | TCP | **Never** — same principle, additionally password-protected (`requirepass`) | `docker-compose.yaml` |
| `garage` — S3 API | `${GARAGE_S3_PORT:-3900}` | HTTP(S) | **Option B only.** Must be directly reachable by the **browser** (presigned uploads/downloads) — localhost, LAN IP, or a domain behind TLS | `docker-compose.garage.yml` |
| `garage` — internal RPC | 3901 | TCP | **Never** — not published, reserved for communication between Garage nodes (not relevant with a single node) | `docker-compose.garage.yml` |
| `garage` — admin API | 3903 | HTTP | **Never** — not published by default; only expose it if you administer the cluster remotely, and then only behind restricted access | `docker-compose.garage.yml` |
| `turn` (coturn) — signaling | 3478 | UDP + TCP | **TURN add-on only.** Direct on the Internet (a reverse proxy can't handle UDP) | `docker-compose.turn.yml` |
| `turn` (coturn) — media relay | `${TURN_MIN_PORT:-49160}`–`${TURN_MAX_PORT:-49200}` | UDP | **TURN add-on only.** Direct on the Internet, the whole range | `docker-compose.turn.yml` |

**Simple rule**: anything not listed above as needing exposure should stay unreachable from outside the internal Docker network. `db` and `redis-db` in particular must **never** be published to the Internet, even temporarily for debugging (a commented-out line exists in `docker-compose.yaml` for one-off local access, explicitly bound to `127.0.0.1` — don't change it to listen on `0.0.0.0`).

## Docker networking model

- By default, all services in the same `docker-compose` file share an automatically created internal bridge network — that's what lets `backend` reach `db:5432` or `redis-db:6379` by service name, never going through the host.
- Only the services listed above with a required exposure use `ports:` to publish a port to the host.
- Exception: the `turn` service uses `network_mode: host` (no bridge network at all) — required for the UDP relay port range, see [TURN server](./turn) for details. That's why the TURN add-on is **Linux only**.

## DNS

| Requirement | When |
|---|---|
| An A/AAAA record pointing to this server, for the main domain | Always, as soon as you expose Kagibi beyond `localhost` (see [Nginx](./nginx)) |
| A second A/AAAA record, a dedicated subdomain (e.g. `s3.your-domain.example`) | Option B (Garage) only, if you choose to expose Garage via a domain rather than a LAN IP or `localhost` (see [Storage](./storage)) |
| None — `TURN_EXTERNAL_IP` expects an IP, not a domain name | TURN add-on |

## Cloudflare Tunnel

[Cloudflare Tunnel](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/) (`cloudflared`) opens an outbound connection from your server to Cloudflare — no inbound port to open, no static public IP needed. How well it fits depends on the service:

| Service | Works with Cloudflare Tunnel? |
|---|---|
| `frontend` / `backend` (HTTP + WebSocket) | **Yes** — standard use case, including the WebSocket upgrade on `/api/v1/ws`. |
| `garage` — S3 API (Option B) | **With a caveat** — it works, but Cloudflare's Free/Pro plans cap **any request routed through the tunnel at 100MB**. Larger files would fail. If you need large file support, don't route Garage through the tunnel — use a direct LAN/public IP or your own TLS reverse proxy instead (see [Storage](./storage)). |
| `turn` (coturn) | **No — even on standard paid plans.** Cloudflare Tunnel only proxies TCP/HTTP/WebSocket; TURN's actual media relay path is UDP, which the standard Tunnel doesn't support. Only [Cloudflare Spectrum](https://developers.cloudflare.com/spectrum/) (an **Enterprise**, paid product) proxies arbitrary UDP. In practice: keep exposing coturn directly (plain port-forwarding, see the firewall summary below) — the Tunnel doesn't solve this case. |

**Alternative to self-hosting coturn at all**: Cloudflare offers its own managed TURN service ([Cloudflare Realtime TURN](https://developers.cloudflare.com/realtime/turn/)) — a separate product, not a tunnel to your own coturn, but a TURN server hosted by Cloudflare that you configure directly via `TURN_SERVER_URL`/`TURN_USERNAME`/`TURN_CREDENTIAL`. If you'd rather not manage coturn's network exposure, that's the option to look at instead of forcing it through a Tunnel.

## Firewall — summary

- **Always**: 443 (and 80 for the HTTPS redirect) open on your reverse proxy.
- **Option B (Garage)**: the `GARAGE_S3_PORT` port (default 3900) open if Garage is exposed on a LAN/public IP rather than `localhost`.
- **TURN add-on**: 3478 (UDP+TCP) and the whole `TURN_MIN_PORT`–`TURN_MAX_PORT` range in UDP.
- **Never**: 5432 (Postgres), 6379 (Redis), 3901/3903 (Garage RPC/admin).
