# Organisations — Kagibi

### Overview

The **Organisations** feature lets multiple users share a common end-to-end encrypted storage space. Every file is encrypted client-side before being sent; the server only receives opaque blobs. The cryptographic architecture is described in detail in [`org-e2e-encryption.md`](./org-e2e-encryption).

---

### Feature access

#### Cloud (kagibi.cloud)

Organisations are a **premium feature**. They require a Pro or Business subscription. Users on the free plan receive an HTTP 402 response on all `/orgs` routes.

The `/organisations` page displays an upgrade screen for non-subscribers, listing the included capabilities.

#### Self-hosted instance

On any instance deployed with `BILLING_ENABLED=false` in the backend environment, the feature is **freely accessible** to all users regardless of their plan. No subscription check is performed.

---

### Roles

Each organisation member holds one of four roles:

| Role | Default access level | Description |
|------|---------------------|-------------|
| `owner` | manage | Sole owner. Full control: can delete the org and rotate the key. |
| `admin` | manage | Can manage members (except other admins), invitations, and permissions. |
| `member` | write | Can read and write (create, move) files. Cannot delete unless granted by permission override. |
| `viewer` | read | Read-only access. Can browse and download files. |

**Key rules:**
- Only the owner can invite admins.
- Only the owner can delete the organisation.
- Only the owner can rotate the encryption key.
- The `owner` role cannot be changed via API (no ownership transfer).
- Admins cannot change another admin's role or remove another admin.

---

### Per-folder permissions

Beyond global roles, **permission overrides** can be set per folder and per user. Only admins and the owner can configure these overrides.

#### Permission levels

| Level | Read | Create | Move | Delete |
|-------|------|--------|------|--------|
| `none` | ✗ | ✗ | ✗ | ✗ |
| `read` | ✓ | ✗ | ✗ | ✗ |
| `write` | ✓ | ✓ | ✓ | ✗ |
| `manage` | ✓ | ✓ | ✓ | ✓ |

#### Granular flags

Each override also exposes fine-grained flags:

| Field | Type | Description |
|-------|------|-------------|
| `perm_create` | bool | Can upload files and create sub-folders |
| `perm_delete` | bool | Can delete files |
| `perm_download` | bool | Can download files (default: `true`) |
| `perm_move` | bool | Can rename and move elements |

#### Hierarchical resolution

Permissions are resolved **from the most-specific path up to the root** (`/`):
1. If an override exists for `/projects/secret`, it is applied.
2. Otherwise, walk up to `/projects`.
3. Otherwise, walk up to `/`.
4. If no override exists, the role's default level is used.

Owners and admins cannot have their download access restricted by permission overrides.

---

### Invitations

Two invitation types allow adding members:

#### 1. Link invite

- A public token not tied to any specific user.
- Usable by anyone with the link.
- The new member joins **without an encryption key** (`encrypted_org_key` is empty).
- An admin/owner must then **provision the key** via the Members tab ("Provision access" button).
- Can be limited by number of uses (`max_uses`) and duration (`expires_at`).

#### 2. Direct invite

- Linked to a specific `target_user_id`.
- The inviter **encrypts the OrgKey** with the target's RSA public key before creating the invitation.
- The new member joins with their key already available — no provisioning needed.
- The invitation is automatically revoked after acceptance (single-use).

#### Invitation lifecycle

```
Created (status = "active")
    │
    ├─ Expired (expires_at in the past) → rejected at acceptance
    ├─ Exhausted (uses >= max_uses)     → rejected at acceptance
    ├─ Manually revoked                 → status = "revoked"
    └─ Accepted (direct invite)         → status = "revoked" after acceptance
```

---

### File operations

#### Multipart upload

1. `POST /orgs/:id/fs/multipart/initiate` — declares the file, receives presigned S3 URLs.
2. Upload chunks directly to S3 (never through the backend).
3. `POST /orgs/:id/fs/multipart/complete` — validates the upload, stores the file's `encrypted_key`.
4. `POST /orgs/:id/fs/multipart/abort` — cancels an in-progress upload.

Each chunk is encrypted client-side (AES-256-GCM) before being sent. The file's `encrypted_key` is an AES-GCM wrap of the FileKey with the OrgKey.

#### Download

1. `GET /orgs/:id/fs/file/:fileID/download` — returns a presigned S3 URL.
2. The client retrieves the OrgKey (from cache or via RSA decryption), decrypts the FileKey, then decrypts chunks in streaming fashion.

#### Listing and navigation

`GET /orgs/:id/fs/list/*path` — lists the contents of a folder (wildcard path parameter).

#### Folders

- `POST /orgs/:id/fs/folder` — create a folder.
- `DELETE /orgs/:id/fs/folder/:folderID` — delete a folder (and its contents).

---

### Key management

#### Organisation key initialisation

When an organisation is created, the owner generates an **OrgKey** (AES-256-GCM) client-side. It is immediately encrypted with their RSA public key and stored in `OrgMember.encrypted_org_key`.

For members joining via a public link, the key must be provisioned manually:

```
Admin sees the red badge on the member with no key
    │
    ├─ 1. Decrypts OrgKey with their RSA private key (from cache or authStore)
    ├─ 2. Encrypts OrgKey with the member's RSA public key
    └─ 3. PATCH /orgs/:id/members/:memberID/key { encrypted_org_key: "..." }
```

#### Key rotation (owner only)

Rotation generates a new OrgKey and re-wraps all FileKeys in a single **atomic transaction**:

```
Owner (frontend)
    │
    ├─ 1. GET /orgs/:id/fs/all-keys          → list of {file_id, encrypted_key}[]
    │
    ├─ 2. Generate a new OrgKey (AES-256-GCM)
    │
    ├─ 3. For each member: encrypt the new OrgKey with the member's RSA key
    │
    ├─ 4. For each file:
    │      a. Decrypt the old FileKey (unwrap with old OrgKey)
    │      b. Re-encrypt the FileKey (wrap with new OrgKey)
    │
    └─ 5. POST /orgs/:id/rotate-key {
              member_keys: [{member_id, encrypted_org_key}...],
              file_keys:   [{file_id, encrypted_key}...]
           }
           → Atomic DB transaction (all keys or none)
```

---

### Audit log

The audit log is accessible to admins and the owner via `GET /orgs/:id/audit` (50 entries per page, reverse chronological order). Only entries from the **last 12 months** are returned; older entries are automatically hidden.

#### Log deletion

Admins and the owner can delete entries via `DELETE /orgs/:id/audit` using three modes:

| Mode | JSON body | Effect |
|------|-----------|--------|
| `"all"` | `{ "mode": "all" }` | Deletes all entries for this organisation |
| `"months"` | `{ "mode": "months", "months": ["2026-01", "2026-02"] }` | Deletes entries for the specified YYYY-MM months |
| `"days"` | `{ "mode": "days", "days": ["2026-01-15"] }` | Deletes entries for the specified YYYY-MM-DD days |

The deletion itself is recorded in the log (`audit_cleared`).

The per-day summary (used to build the calendar UI) is available via `GET /orgs/:id/audit/summary` — returns `{ "days": { "YYYY-MM-DD": count } }` for the past year.

#### Recorded actions

| Action | Trigger |
|--------|---------|
| `file_uploaded` | Multipart upload completed |
| `file_downloaded` | File download |
| `file_deleted` | File deletion |
| `member_joined` | Invitation accepted |
| `member_removed` | Member removed (by admin or self-removal) |
| `role_changed` | Member role modified |
| `key_provisioned` | Org key provisioned for a member |
| `key_rotated` | Organisation key rotation |
| `invitation_created` | Invitation created |
| `invitation_revoked` | Invitation revoked |
| `permission_set` | Permission override created or updated |
| `permission_removed` | Permission override deleted |
| `audit_cleared` | Audit log entries deleted |

---

### LDAP / Active Directory Sync

Each organization can connect to an **LDAP** or **Active Directory** server to automatically import members and groups. Configuration is restricted to `admin` and `owner` roles.

#### Sync flow

1. The scheduler runs a full cycle at the configured interval (default: 60 min).
2. An LDAP connection is opened and the service account performs a `bind`.
3. A user search is performed under `user_base_dn` with `user_filter`.
4. **Safeguard 1**: if the result count < `min_expected_users`, the sync is aborted.
5. **Safeguard 2**: if > 20% of existing LDAP members disappear at once, the sync is aborted.
6. For each LDAP user:
   - If they already have a Kagibi account and are already a member → `source` and `ldap_uid` updated.
   - If they have an account but are not yet a member → a direct invitation is sent.
   - If they have no account → a link invitation is sent by email.
7. Members with `source = 'ldap'` absent from the LDAP result are **suspended** (`suspended_at` set).
8. Members suspended for longer than `auto_deprovision_days` are **removed** from the org.
9. If `group_base_dn` is set, LDAP groups are created or updated as Kagibi groups (`source = 'ldap'`).

#### Network constraint: direct outbound connection

**The current model requires the Kagibi backend to reach the client's LDAP/AD server directly.** `client.Dial()` (`backend/internal/ldap/client.go`) opens an outbound TCP connection to the configured URL — there is no agent or intermediary tunnel on the client side.

Implications for a real-world deployment (a client whose directory sits on their internal network, outside Kagibi's own network):

- **The client's LDAP/AD server must be publicly reachable** on the configured port (389/636), at least from Kagibi's outbound IP — typically via a NAT/port-forward on the client's firewall.
- **`ldaps://` (implicit TLS) is strongly recommended** for any traffic outside a local network: the bind DN and password travel in plain text over bare `ldap://`. The client's `StartTLS` fallback (`client.go`) fails silently (server-side log only) if the target doesn't support it — nothing on the Kagibi side prevents an effectively unencrypted connection.
- **IP restriction on the client side**: to avoid exposing the directory to the whole internet, the client should ideally restrict access to Kagibi's outbound IP. That requires a **stable, documented outbound IP** on Kagibi's side (a dedicated NAT gateway) — not guaranteed by default depending on backend hosting infrastructure.
- **Expected client friction**: most enterprise security/IT teams refuse to expose an internal directory (AD in particular) to the internet, even IP-restricted — it's an actively scanned attack vector. This model suits smaller organizations or ones already internet-facing (LDAP already cloud-hosted, existing VPN), but is a hard blocker for clients with strict security requirements.

The standard industry alternative (Okta, Azure AD Connect, JumpCloud…) is an **agent installed inside the client's network**, which queries the directory locally and opens an **outbound** connection to the SaaS itself (HTTPS/WebSocket) — no inbound port required on the client side. Kagibi does not offer this mode today.

#### Bind password encryption

The service account password is encrypted with **AES-256-GCM** via `emailcrypto.Encrypt` (key derived from `EMAIL_ENCRYPTION_KEY`) before storage. It is never transmitted in plain text.

#### Deprovisioning

| Phase | Trigger | Effect |
|-------|---------|--------|
| Suspension | Absent from LDAP on the next cycle | `suspended_at` recorded, org access removed |
| Removal | `suspended_at` + `auto_deprovision_days` exceeded | Permissions, groups, and membership deleted |

If `auto_deprovision_days = 0`, removal is manual only.

---

### Data models

#### Organization

| Field | Type | Description |
|-------|------|-------------|
| `id` | int64 | Unique identifier |
| `name` | string | Organisation name |
| `description` | string | Optional description |
| `owner_id` | string | Owner's user ID |
| `storage_quota_mb` | int64 | Storage quota in MB (default: 10 240 MB) |
| `created_at` | timestamp | Creation date |
| `updated_at` | timestamp | Last modification date |

#### OrgMember

| Field | Type | Description |
|-------|------|-------------|
| `id` | int64 | Identifier |
| `org_id` | int64 | Reference to the organisation |
| `user_id` | string | Reference to the user |
| `role` | string | `owner` / `admin` / `member` / `viewer` |
| `encrypted_org_key` | string | RSA-OAEP encrypted OrgKey (base64) |
| `source` | string | `internal` (manual) / `ldap` (directory sync) |
| `ldap_uid` | string | LDAP UID of the user (empty if `source = internal`) |
| `suspended_at` | timestamp? | LDAP suspension date (nil if active) |
| `joined_at` | timestamp | Join date |

#### OrgLDAPConfig

LDAP configuration for an organisation (one row per org).

| Field | Type | Description |
|-------|------|-------------|
| `org_id` | int64 | Reference to the organisation (unique) |
| `enabled` | bool | Whether sync is active |
| `url` | string | Server URL (`ldap://` or `ldaps://`) |
| `bind_dn` | string | Service account DN |
| `bind_password_enc` | string | AES-256-GCM encrypted password |
| `user_base_dn` | string | Root for user searches |
| `user_filter` | string | LDAP filter (default: `(objectClass=person)`) |
| `group_base_dn` | string | Root for group searches (empty = disabled) |
| `group_filter` | string | Group filter (default: `(objectClass=groupOfNames)`) |
| `attr_email` | string | Email attribute (default: `mail`) |
| `attr_display_name` | string | Display name attribute (default: `cn`) |
| `attr_uid` | string | Unique UID attribute (default: `uid`) |
| `tls_skip_verify` | bool | Skip TLS certificate verification |
| `sync_interval_minutes` | int | Sync interval in minutes (min. 5) |
| `auto_deprovision_days` | int | Grace period before removal (0 = manual) |
| `min_expected_users` | int | Anti-wipe safety threshold |
| `last_sync_at` | timestamp? | Timestamp of the last sync |
| `last_sync_error` | string | Error message from the last sync (empty if OK) |
| `last_sync_stats` | jsonb | Last sync statistics (`LDAPSyncStats`) |

#### OrgInvitation

| Field | Type | Description |
|-------|------|-------------|
| `id` | int64 | Identifier |
| `org_id` | int64 | Target organisation |
| `invited_by` | string | Inviter's user ID |
| `token` | string | Invitation token (32 char hex) |
| `target_user_id` | string? | Direct invite target (nil = link) |
| `encrypted_org_key` | string | Pre-encrypted OrgKey for direct invites |
| `role` | string | Role assigned upon acceptance |
| `max_uses` | int | Max number of uses (0 = unlimited) |
| `uses` | int | Current use count |
| `expires_at` | timestamp? | Expiry date |
| `status` | string | `active` / `revoked` |

#### OrgFolderPermission

| Field | Type | Description |
|-------|------|-------------|
| `org_id` | int64 | Organisation |
| `user_id` | string | Target member |
| `folder_path` | string | Normalised path (e.g. `/projects/secret`) |
| `level` | string | `none` / `read` / `write` / `manage` |
| `perm_create` | bool | Can create |
| `perm_delete` | bool | Can delete |
| `perm_download` | bool | Can download |
| `perm_move` | bool | Can move |

#### OrgAuditLog

| Field | Type | Description |
|-------|------|-------------|
| `id` | int64 | Identifier |
| `org_id` | int64 | Organisation |
| `actor_id` | string | User who triggered the action |
| `action` | string | Action type (see table above) |
| `target_id` | string | ID of the affected resource |
| `target_type` | string | `user` / `file` / `invitation` / `permission` / `org` |
| `detail` | string | Human-readable detail (e.g. `member → admin`) |
| `created_at` | timestamp | Timestamp |

---

### API reference

All routes (except `GET /org-invitations/:token`) require a valid JWT in the `Authorization: Bearer <token>` header.

On cloud, routes marked ★ return HTTP 402 for users on the free plan.

#### Organisations

| Method | Route | Auth | Description |
|--------|-------|------|-------------|
| `POST` | `/api/v1/orgs` | JWT ★ | Create an organisation |
| `GET` | `/api/v1/orgs` | JWT ★ | List own organisations |
| `GET` | `/api/v1/orgs/:orgID` | JWT | Get organisation details |
| `PATCH` | `/api/v1/orgs/:orgID` | JWT (admin+) | Update name / description / quota |
| `DELETE` | `/api/v1/orgs/:orgID` | JWT (owner) | Delete the organisation |

#### Members

| Method | Route | Auth | Description |
|--------|-------|------|-------------|
| `GET` | `/api/v1/orgs/:orgID/members` | JWT (member+) | List members |
| `PATCH` | `/api/v1/orgs/:orgID/members/:memberID` | JWT (admin+) | Change a member's role |
| `DELETE` | `/api/v1/orgs/:orgID/members/:memberID` | JWT | Remove a member (self or admin+) |
| `PATCH` | `/api/v1/orgs/:orgID/members/:memberID/key` | JWT (admin+) | Provision a member's org key |

#### Invitations

| Method | Route | Auth | Description |
|--------|-------|------|-------------|
| `POST` | `/api/v1/orgs/:orgID/invitations` | JWT (admin+) | Create an invitation |
| `GET` | `/api/v1/orgs/:orgID/invitations` | JWT (admin+) | List active invitations |
| `DELETE` | `/api/v1/orgs/:orgID/invitations/:invID` | JWT (admin+) | Revoke an invitation |
| `GET` | `/api/v1/org-invitations/:token` | Public | Preview invitation info |
| `POST` | `/api/v1/org-invitations/:token/accept` | JWT ★ | Accept an invitation |

#### File system

| Method | Route | Auth | Description |
|--------|-------|------|-------------|
| `GET` | `/api/v1/orgs/:orgID/fs/list/*path` | JWT (member+) | List a folder |
| `POST` | `/api/v1/orgs/:orgID/fs/folder` | JWT (write+) | Create a folder |
| `DELETE` | `/api/v1/orgs/:orgID/fs/folder/:folderID` | JWT (manage) | Delete a folder |
| `GET` | `/api/v1/orgs/:orgID/fs/file/:fileID/download` | JWT (member+) | Download a file |
| `GET` | `/api/v1/orgs/:orgID/fs/file/:fileID/key` | JWT (member+) | Retrieve a file's encrypted key |
| `DELETE` | `/api/v1/orgs/:orgID/fs/file/:fileID` | JWT (manage) | Delete a file |
| `POST` | `/api/v1/orgs/:orgID/fs/multipart/initiate` | JWT (write+) | Initiate a multipart upload |
| `POST` | `/api/v1/orgs/:orgID/fs/multipart/complete` | JWT (write+) | Complete a multipart upload |
| `POST` | `/api/v1/orgs/:orgID/fs/multipart/abort` | JWT (write+) | Abort a multipart upload |

#### Permissions

| Method | Route | Auth | Description |
|--------|-------|------|-------------|
| `GET` | `/api/v1/orgs/:orgID/permissions` | JWT (admin+) | List all permission overrides |
| `PUT` | `/api/v1/orgs/:orgID/permissions` | JWT (admin+) | Create or update an override |
| `DELETE` | `/api/v1/orgs/:orgID/permissions` | JWT (admin+) | Delete an override |
| `GET` | `/api/v1/orgs/:orgID/permissions/me` | JWT (member+) | Caller's effective permission |

#### Administration

| Method | Route | Auth | Description |
|--------|-------|------|-------------|
| `GET` | `/api/v1/orgs/:orgID/audit` | JWT (admin+) | Audit log (50 per page, last year) |
| `GET` | `/api/v1/orgs/:orgID/audit/summary` | JWT (admin+) | Per-day entry counts for the last year |
| `DELETE` | `/api/v1/orgs/:orgID/audit` | JWT (admin+) | Delete entries (all / months / days) |
| `GET` | `/api/v1/orgs/:orgID/fs/all-keys` | JWT (admin+) | All file keys (for rotation) |
| `POST` | `/api/v1/orgs/:orgID/rotate-key` | JWT (owner) | Rotate the organisation key |
| `GET` | `/api/v1/orgs/:orgID/ldap` | JWT (admin+) | Get LDAP configuration |
| `PUT` | `/api/v1/orgs/:orgID/ldap` | JWT (admin+) | Save LDAP configuration |
| `POST` | `/api/v1/orgs/:orgID/ldap/test` | JWT (admin+) | Test LDAP connection |
| `POST` | `/api/v1/orgs/:orgID/ldap/sync` | JWT (admin+) | Trigger a manual sync |
| `GET` | `/api/v1/orgs/:orgID/ldap/suspended` | JWT (admin+) | List suspended LDAP members |

---

### Implementation — where to find the code

| Component | File |
|-----------|------|
| Main handler + permission resolution | `backend/handlers/organizations/handler.go` |
| Organisation creation | `backend/handlers/organizations/create.go` |
| List / detail | `backend/handlers/organizations/list.go` |
| Update / delete | `backend/handlers/organizations/update.go` |
| Member management | `backend/handlers/organizations/members.go` |
| Invitations | `backend/handlers/organizations/invitations.go` |
| File system (list, download, delete) | `backend/handlers/organizations/orgfiles.go` |
| Multipart upload | `backend/handlers/organizations/orgmultipart.go` |
| Directory listing (fs/list) | `backend/handlers/organizations/orgfs.go` |
| Per-folder permissions | `backend/handlers/organizations/permissions.go` |
| Audit log + all-keys | `backend/handlers/organizations/orgaudit.go` |
| Key rotation | `backend/handlers/organizations/orgrotatekey.go` |
| LDAP API handlers | `backend/handlers/organizations/ldap.go` |
| LDAP client (dial, bind, search) | `backend/internal/ldap/client.go` |
| LDAP sync engine | `backend/internal/ldap/sync.go` |
| LDAP scheduler (per-org goroutines) | `backend/internal/ldap/scheduler.go` |
| Org crypto primitives | `frontend/src/utils/orgCrypto.js` |
| Pinia store (state + actions) | `frontend/src/stores/organizations.js` |
| Organisations list page | `frontend/src/views/OrganizationsView.vue` |
| Organisation detail page | `frontend/src/views/OrgDetailView.vue` |
| Invitation acceptance page | `frontend/src/views/JoinView.vue` |
| LDAP configuration panel (Vue) | `frontend/src/components/organizations/OrgLDAPPanel.vue` |
| Detailed encryption model | [`org-e2e-encryption.md`](./org-e2e-encryption) |
