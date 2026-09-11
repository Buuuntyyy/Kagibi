# Backup

A Kagibi instance protects two very different kinds of data, with two different strategies:

| Data | Where it lives | Criticality | Approach |
|---|---|---|---|
| Metadata + encrypted keys | PostgreSQL | **Critical** — losing it makes the account unusable even if the files survive | Regular dump pushed to an external S3 bucket |
| File content | External S3 or Garage (self-hosted) | Important but large | Provider-side replication (external S3) or scheduled sync (Garage) |

Important point: files are **already encrypted client-side** before being uploaded (end-to-end encryption). A backup therefore needs no key management at all — copying the objects as-is is enough, they stay unreadable without the keys, which live in the database. **This is exactly why backing up the database is the top priority**: without it, a files-only backup is useless (encrypted blobs with no keys to open them).

## 1. Database (priority)

The idea: a regular, compressed dump, pushed to an S3 bucket — ideally **different** from the one already storing your files (the 3-2-1 principle: a backup on the same medium as the original protects you against nothing if that medium is lost).

### Script

```bash
#!/bin/bash
set -euo pipefail
cd /opt/kagibi   # adjust to your actual install path

set -a; source .env; set +a

DUMP_FILE="kagibi-db-$(date +%Y%m%d-%H%M%S).sql.gz"
TMP_PATH="/tmp/$DUMP_FILE"

docker compose exec -T db pg_dump -U "${DB_USER:-user}" "${DB_NAME:-mydb}" | gzip > "$TMP_PATH"

# An empty or truncated dump isn't a usable backup — don't push it.
if [ ! -s "$TMP_PATH" ] || ! gzip -t "$TMP_PATH" 2>/dev/null; then
  echo "❌ Invalid or empty dump — backup aborted." >&2
  rm -f "$TMP_PATH"
  exit 1
fi

rclone copy "$TMP_PATH" backup:kagibi-backups/db/
rm -f "$TMP_PATH"

echo "✅ DB backup uploaded: $DUMP_FILE"
```

[rclone](https://rclone.org/) is used here instead of a provider-specific tool: it speaks the S3 protocol (and many others), so the same script works regardless of the destination bucket — OVH, AWS, Scaleway, Cloudflare R2, Backblaze B2, another Garage instance...

### Configuring rclone

```bash
rclone config
# Storage type: S3 (Amazon S3 Compliant Storage Providers)
# Provider: yours (or "Other" for a generic S3-compatible store)
# access_key_id / secret_access_key: credentials for the DESTINATION bucket (not Kagibi's own)
# endpoint: the destination provider's S3 URL
```
This creates a remote named `backup` (or whatever name you choose) in `~/.config/rclone/rclone.conf`, reused by the script above (`backup:kagibi-backups/db/`).

### Scheduling (cron)

```bash
# crontab -e
0 3 * * * /opt/kagibi/scripts/backup-db.sh >> /var/log/kagibi-backup.log 2>&1
```

### Retention

Rather than scripting the deletion of old dumps (a real risk of a bug wiping everything at once), configure a **lifecycle rule** on the destination bucket itself, on the provider's side — e.g. "delete objects under `kagibi-backups/db/` after 30 days." This is a standard feature of any S3-compatible storage, managed outside of any script that could have a bug.

## 2. Files

### Option A — External S3 (OVH, AWS, Scaleway, R2...)

Your files already live in a bucket managed by that provider. The right approach is that provider's **native cross-bucket replication** (e.g. Cross-Region Replication on AWS, equivalents on OVH/Scaleway) rather than a Kagibi-side script: it's more reliable, doesn't consume bandwidth/CPU on your instance, and is natively consistent with the source storage. Check your provider's "replication" or "cross-region replication" docs — the whole setup happens on their side, nothing to change in Kagibi.

### Option B — Garage (local self-hosted storage)

**Don't back up by directly copying the `garage/data`/`garage/meta` folders while Garage is running** — it's a live data store, and copying files from disk while it's writing can produce an inconsistent backup. The right method is to sync through the **S3 API that Garage itself exposes**, which guarantees only complete objects are read:

```bash
rclone config
# Create a "garage_local" remote:
#   Type: S3, Provider: Other
#   access_key_id / secret_access_key: the ones generated in garage/credentials/s3.env
#   endpoint: http://localhost:3900 (or whatever IP/domain your setup uses)
```

```bash
rclone sync garage_local:kagibi backup:kagibi-backups/files --transfers 8 --progress
```

Schedule it the same way as the DB dump (cron), at a frequency matching your data volume (a daily sync is enough in most cases — `rclone sync` only transfers what changed).

## 3. Test the restore

A backup that's never been restored isn't a verified backup. Periodically (quarterly, for example):

```bash
# Fetch a dump from the backup bucket
rclone copy backup:kagibi-backups/db/kagibi-db-<date>.sql.gz .

# Restore into a TEST database (NEVER onto production for this test)
gunzip -c kagibi-db-<date>.sql.gz | docker compose exec -T db psql -U "${DB_USER:-user}" "${DB_NAME:-mydb}_test_restore"
```

See also the [Upgrading](./upgrade) page for the backup step to take specifically before a version upgrade (slightly different: a one-off, local snapshot rather than a recurring backup to a remote bucket).
