# Upgrading

Kagibi publishes versioned, signed Docker images (`vX.Y.Z`) on every release — check the [CHANGELOG](https://github.com/Buuuntyyy/Kagibi/blob/main/CHANGELOG.md) or [GitHub Releases](https://github.com/Buuuntyyy/Kagibi/releases) to see what changed before upgrading.

Database migrations run automatically when the backend starts and are **cumulative and idempotent**: jumping straight from an old version to a recent one (without stopping at intermediate versions) is safe, unless a specific release's notes explicitly say otherwise.

## 1. Back up before touching anything

An upgrade that goes wrong must stay reversible. Always back up first:

```bash
# Load variables from .env (DB_USER, DB_NAME, DB_PASSWORD...)
set -a; source .env; set +a

mkdir -p backups
BACKUP_FILE="backups/kagibi-$(date +%Y%m%d-%H%M%S).sql.gz"

docker compose exec -T db pg_dump -U "${DB_USER:-user}" "${DB_NAME:-mydb}" | gzip > "$BACKUP_FILE"

# Minimal sanity check: an empty or truncated dump isn't a usable backup.
if [ ! -s "$BACKUP_FILE" ] || ! gzip -t "$BACKUP_FILE" 2>/dev/null; then
  echo "❌ Backup invalid or empty — DO NOT proceed with the upgrade." >&2
  exit 1
fi
echo "✅ Backup OK: $BACKUP_FILE ($(du -h "$BACKUP_FILE" | cut -f1))"
```

If you're running **local storage (Option B, Garage)**, also back up the actual file data — the database only holds metadata and encrypted keys, not the files themselves:

```bash
tar czf "backups/garage-$(date +%Y%m%d-%H%M%S).tar.gz" garage/data garage/meta
```

With **external S3** (OVH, AWS, Scaleway, R2...), object durability/backup is your provider's responsibility — check their policy if you haven't already.

## 2. Pin the target version

In `.env`:

```bash
KAGIBI_VERSION=v2.32.0   # replace with your target version
```

Never leave `KAGIBI_VERSION=latest` on a long-running production install: without a pinned version, a plain `docker compose pull` can move you to a version you didn't deliberately choose to run at that moment.

## 3. (Recommended) Verify image signatures

Images are signed with [cosign](https://docs.sigstore.dev/) (keyless mode, via GitHub Actions) — verifiable without fetching any key:

```bash
cosign verify \
  --certificate-identity-regexp "^https://github\.com/Buuuntyyy/Kagibi/\.github/workflows/release\.yml@refs/heads/main$" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  ghcr.io/bunnntyyy/kagibi-backend:v2.32.0
```

(swap the tag for the equivalent `kagibi-frontend` command). A successful verification confirms the image genuinely came from this repository's official pipeline and wasn't tampered with between publishing and your `pull`.

## 4. Apply the upgrade

```bash
docker compose pull
docker compose up -d
```

Migrations run automatically as the backend starts.

## 5. Verify everything went well

```bash
# The backend started and responds
curl -f http://localhost:${BACKEND_PORT:-8080}/health

# Migrations completed without error
docker compose logs backend | grep "\[migrate\] Schema OK"
```

The second command should print a line like:

```
[migrate] Schema OK — 28 étapes appliquées (version applicative v2.32.0)
```

The version number in that line should match `KAGIBI_VERSION`. If the line never appears, the migration likely failed partway through — check the preceding lines (`Warning: ...` or a fatal error) in the same logs.

## 6. If something goes wrong — roll back

```bash
# 1. Revert to the previous version in .env
KAGIBI_VERSION=v2.31.0   # the version that was working before

# 2. Restart on that version
docker compose up -d

# 3. Only if the database was actually altered by the failed attempt, restore the
#    dump taken in step 1 (migrations being additive, reverting the image alone is
#    enough in the vast majority of cases):
gunzip -c backups/kagibi-<timestamp>.sql.gz | docker compose exec -T db psql -U "${DB_USER:-user}" "${DB_NAME:-mydb}"
```

## What makes this procedure safe

- **Signed, versioned images**: a pinned `KAGIBI_VERSION` guarantees you know exactly what's running, and rolling back means pointing at a specific tag, not rebuilding an old commit.
- **Idempotent, tracked migrations**: every migration step is recorded in the `schema_migrations` table along with the app version — the log line in step 5 is direct, checkable proof of that.
- **No destructive drops across versions**: replaced columns/tables are added alongside the old ones instead of overwriting them immediately, which keeps rolling back realistic.
