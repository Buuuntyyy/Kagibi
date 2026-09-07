# Prerequisites

These prerequisites are common to both storage options — external S3 or fully local (Garage), see the [Storage](./storage) page.

## What you need

- **Docker Engine** with the **Compose v2** plugin (`docker compose version` must work — not the legacy Python `docker-compose` v1 binary).

## Steps

1. Clone the repository:
   ```bash
   git clone https://github.com/Buuuntyyy/Kagibi.git
   cd Kagibi
   ```

2. Create the root environment file:
   ```bash
   cp .env.example .env
   ```

3. Generate the four required secrets and paste them into `.env`:

   | Variable | Command | Purpose |
   |---|---|---|
   | `JWT_SECRET` | `openssl rand -hex 32` | Signs session tokens |
   | `EMAIL_ENCRYPTION_KEY` | `openssl rand -hex 32` | Encrypts emails at rest — **never change after first use** |
   | `DB_PASSWORD` | `openssl rand -hex 24` | PostgreSQL password |
   | `REDIS_PASSWORD` | `openssl rand -hex 24` | Redis password (requirepass) |

4. Adapt the URLs to your deployment in `.env`:
   - `VITE_API_URL` — URL the **browser** uses to reach the backend (`http://localhost:8080/api/v1` locally, `https://your-domain.example/api/v1` in production).
   - `ALLOWED_ORIGINS` — the frontend's origin, must match exactly (`http://localhost` locally, `https://your-domain.example` in production). Never `*` — incompatible with sending cookies/credentials.

`BILLING_ENABLED=false` is already `.env.example`'s default — nothing to do, unlimited storage and no billing UI, which is recommended for self-hosting.

## Next step

Once these steps are done, head to the [Storage](./storage) page to choose and configure your backend.
