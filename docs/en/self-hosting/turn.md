# TURN server (P2P)

## Do you actually need one?

Kagibi's P2P transfer uses WebRTC: in most cases, a direct connection between the two devices is enough, or failing that a plain **STUN** server (already configured by default, free, provided by Google). A **TURN** server — which relays all the traffic instead of just helping establish the connection — is only needed when **both** devices are behind strict or symmetric NAT that STUN can't traverse (common on some corporate networks, some mobile carriers, or double NAT/CGNAT at the ISP level).

**Only deploy this service if you've actually seen P2P transfers fail to connect** — otherwise, it isn't needed.

## Two options

| | Effort | When |
|---|---|---|
| **A — Use an existing TURN server** | Minimal — just 3 variables to fill in | You already have a TURN server (Twilio, a coturn you already run elsewhere...) |
| **B — Deploy coturn** (this page) | One more Docker service, ports to open | You want to self-host everything |

### Option A — Existing TURN server

In `.env`:
```bash
TURN_SERVER_URL=turn:turn.your-domain.example:3478
TURN_USERNAME=...
TURN_CREDENTIAL=...
```
Nothing else to do — skip ahead to [Verify](#verify) below.

### Option B — Deploy coturn

**Linux only.** `docker-compose.turn.yml` uses `network_mode: host` — this is coturn's own documented approach for Docker: a TURN relay allocates a UDP port per active connection from a wide range, which Docker's usual per-port mapping doesn't scale to. Host networking isn't reliably supported by Docker Desktop on Mac/Windows; that's fine — a service that needs to relay real traffic belongs on a dedicated Linux server anyway, not a dev machine.

#### 1. Fill in `.env`

```bash
# Required — this server's PUBLIC IP (not a LAN IP)
TURN_EXTERNAL_IP=203.0.113.10

# Required — a single shared credential pair used by all users
TURN_USERNAME=...   # openssl rand -hex 24
TURN_CREDENTIAL=...  # openssl rand -hex 24

# Optional
# TURN_REALM=kagibi
# TURN_MIN_PORT=49160
# TURN_MAX_PORT=49200
```
Do **not** also fill in option A's `TURN_SERVER_URL`/`TURN_USERNAME`/`TURN_CREDENTIAL` — `docker-compose.turn.yml` derives them automatically for the backend from the variables above (the same credentials are used by both coturn and the backend, one place to change).

#### 2. Open the firewall

| Port | Protocol | Purpose |
|---|---|---|
| 3478 | UDP + TCP | TURN signaling |
| `TURN_MIN_PORT`–`TURN_MAX_PORT` (default 49160–49200) | UDP | Actual media relay traffic |

Without this, coturn starts up fine but no external client can actually reach it.

#### 3. Launch

```bash
docker compose -f docker-compose.yaml -f docker-compose.turn.yml up -d
```

## Hardening included by default

A poorly configured TURN relay can become an **open relay into your own internal network** (SSRF) — a malicious client asks it to relay traffic not to another WebRTC peer, but to an internal IP on your network, or even to the cloud metadata endpoint (`169.254.169.254`, a classic way to steal a cloud VM's IAM credentials). `docker-compose.turn.yml` already blocks this by default:

- **Private/special-use IP ranges denied as relay targets** (`--denied-peer-ip`): RFC1918 (`10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`), loopback, link-local (`169.254.0.0/16`, i.e. cloud metadata), CGNAT (`100.64.0.0/10`), and a few other special-use ranges. Loopback is denied by coturn by default anyway.
- **`--no-software-attribute`**: hides coturn's version in its responses (avoids making it easier to target a known CVE for a specific version).
- **`--no-cli`**: disables coturn's telnet admin interface (never needed for this use case).
- **Quotas** (`TURN_USER_QUOTA`/`TURN_TOTAL_QUOTA`, default 50): caps the number of concurrent relayed sessions. Important to understand: Kagibi uses a **single credential pair shared by all users** (no per-session temporary credentials) — with `--lt-cred-mech`, this "per-user" quota is therefore effectively a global cap on this relay. It bounds the damage if these credentials leak or get used outside Kagibi, without impacting normal usage.

**About the shared credentials**: `TURN_USERNAME`/`TURN_CREDENTIAL` are sent to the browser of any authenticated Kagibi user (via `GET /ice-config`, visible in the browser's dev tools) — these aren't as confidential as a server-side secret. Treat them as "accessible to any logged-in user," not as an infrastructure secret, and rotate them periodically (`openssl rand -hex 24`, then update `.env` and restart) if you want to limit the exposure window.

## Verify

From two different networks (e.g. your home connection plus your phone's mobile hotspot, to make sure you're going through two separate NATs), start a P2P transfer in the Kagibi UI. If it succeeds where it used to fail, TURN is working.

For a more direct test, [webrtc.github.io's trickle-ice sample](https://webrtc.github.io/samples/src/content/peerconnection/trickle-ice/) lets you enter your TURN server (`turn:your-ip:3478`, with your credentials) and check that a `relay`-type ICE candidate actually gets generated.

## Going further

`docker-compose.turn.yml` covers the common case (single-node relay, shared static credentials, baseline hardening). For anything more advanced — TLS/TURNS, IPv6, temporary REST-API credentials, multi-node clustering, fine-grained performance tuning — refer to the official coturn project documentation: **[github.com/coturn/coturn](https://github.com/coturn/coturn)** (README + [wiki](https://github.com/coturn/coturn/wiki)), which covers every `turnserver` option.

## Troubleshooting

| Symptom | Likely cause |
|---|---|
| No `relay` candidate generated (trickle-ice test) | Firewall blocking port 3478 or the relay port range — test from outside your network, not just locally |
| coturn starts then immediately exits | One of the required variables (`TURN_EXTERNAL_IP`, `TURN_USERNAME`, `TURN_CREDENTIAL`) is missing from `.env` — `docker compose logs turn` shows the exact error |
| Works locally but not from outside | `TURN_EXTERNAL_IP` must be the server's **public** IP, not a LAN IP (e.g. `192.168.x.x`) |
