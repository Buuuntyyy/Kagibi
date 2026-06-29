# Kagibi

**End-to-end encrypted cloud storage, without compromise.**  
**Stockage cloud chiffré de bout en bout, sans compromis.**

---

Read the documentation in your language:

- [English](./docs/README.en.md)
- [Français](./docs/README.fr.md)

---

## Features

### Security & Privacy
- **Zero-knowledge encryption** — files are encrypted on your device before upload; the server never sees plaintext content
- **AES-256-GCM** file encryption, split into 10 MB chunks with unique nonces
- **Argon2id** key derivation (64 MB memory, 4 iterations)
- **RSA-OAEP 4096-bit** key pair per user for asymmetric sharing
- **Optional filename encryption** — file and folder names can be encrypted client-side at registration
- **MFA (TOTP)** two-factor authentication support
- **Account recovery code** — password-independent recovery without server access to your keys

### File Management
- Upload, download, organize files and folders
- Streaming decryption — large files are never fully loaded into memory
- Recently opened files
- File search (disabled when filename encryption is active)

### Sharing
- **P2P transfer** — direct device-to-device file transfer over WebRTC DataChannel, encrypted with AES-GCM; a TURN relay (**no intermediate storage**, only relay) is used when a direct connection is not possible, without compromising encryption.
- **Public link sharing** — optional password protection and expiration (1–30 days)
- **Friend sharing (user-to-user)** — end-to-end encrypted via RSA; only the recipient can decrypt

### Social
- Friend system via unique friend codes (e.g. `#A7KD92XZ`)
- Real-time online presence
- P2P transfer notifications with sound and browser push

### Account & Settings
- Light / dark theme
- Multi-language interface (i18n)
- Usage dashboard (storage, P2P quota)
- GDPR-compliant data export and account deletion (30-day grace period)

---

## Roadmap

> This section lists planned and considered features. It is not a commitment to a release date.

<!-- Add roadmap items below. Example format:
- [ ] Feature name — short description
- [x] Already shipped feature
-->

---

## Self-hosting (Docker)

**Prerequisites:** [Docker + Docker Compose](https://docs.docker.com/get-docker/) · an S3-compatible bucket (AWS, OVH, Scaleway, Cloudflare R2, MinIO…)

**1. Clone and configure**

```bash
git clone https://github.com/Buuuntyyy/Kagibi.git
cd Kagibi
cp .env.example .env
```

**2. Edit `.env` — required values**

| Variable | Description |
|---|---|
| `JWT_SECRET` | Random secret — `openssl rand -hex 32` |
| `EMAIL_ENCRYPTION_KEY` | AES-256 key — `openssl rand -hex 32` ⚠ never change after first use |
| `S3_ENDPOINT` / `S3_BUCKET` / `S3_REGION` | Your S3 bucket coordinates |
| `S3_ACCESS_KEY` / `S3_SECRET_KEY` | S3 credentials |
| `VITE_API_URL` | Browser-facing backend URL — `http://localhost:8080/api/v1` locally, `https://your-domain.com/api/v1` in production |
| `ALLOWED_ORIGINS` | Frontend URL for CORS — `http://localhost` locally, `https://your-domain.com` in production |

**3. Start**

```bash
docker compose up -d
```

Startup order is managed automatically: PostgreSQL and Redis start first, the backend waits for them to be healthy, the frontend waits for the backend. Database migrations run automatically on first startup.

Frontend: `http://localhost` — Backend: `http://localhost:8080`

All data is persisted in named Docker volumes (`db_data`, `redis_data`).

**Optional settings** — all have safe defaults for self-hosting:

| Variable | Default | Description |
|---|---|---|
| `DB_PASSWORD` | `postgres` | PostgreSQL password (change in production) |
| `AUTH_PROVIDER` | `local` | `local` (recommended), `supabase`, or `pocketbase` |
| `BILLING_ENABLED` | `false` | `false` = unlimited storage, no billing UI |
| `TURN_SERVER_URL` | — | P2P relay for devices behind strict NAT |
| `GOOGLE_OAUTH_CLIENT_ID` | — | Enable Google Drive import |
| `FRONTEND_PORT` / `BACKEND_PORT` | `80` / `8080` | Exposed ports |

**Verify the deployment**

```bash
docker compose ps                    # all 4 services should show "healthy"
curl http://localhost:8080/health    # → {"status":"ok"}
```

Then open [http://localhost](http://localhost) in your browser.

**Production (reverse proxy + TLS)**

Update `.env` with your domain and free port 80 for the reverse proxy:

```env
VITE_API_URL=https://your-domain.com/api/v1
ALLOWED_ORIGINS=https://your-domain.com
FRONTEND_PORT=3000
```

Rebuild the frontend: `docker compose up -d --build frontend`

Then configure your reverse proxy — see [`docs/README.en.md`](./docs/README.en.md) for Caddy and Nginx examples.

## Local development

**Prerequisites:** [Go 1.21+](https://go.dev/dl/), [Docker](https://docs.docker.com/get-docker/), [Node.js 18+](https://nodejs.org/)

```bash
# Infrastructure (PostgreSQL + Redis)
docker compose up db redis -d

# Backend
cp backend/.env.example backend/.env   # edit S3_* and secrets
cd backend && go run .

# Frontend (new terminal)
cp frontend/.env.example frontend/.env
cd frontend && npm install && npm run dev
```

Frontend: `http://localhost:5173` — Backend: `http://localhost:8080`

## License

AGPLv3 — see [`LICENSE`](./LICENSE).
