# Kagibi

**End-to-end encrypted cloud storage, without compromise.**

Kagibi is a cloud storage platform built around a simple principle: **what you store is yours alone**. The server cannot read your files — not because we promise not to, but because we are technically incapable of doing so.

This project was developed by [Buuuntyyy] with the assistance of AI for certain development and documentation tasks. The goal is to provide a secure, privacy-respecting, easy-to-use storage solution while being fully transparent about how it works internally.

---

## Philosophy

Most cloud solutions encrypt your data *on the server*, using keys the provider controls. In the event of a breach, a legal request, or internal abuse, your data is exposed.

Kagibi works differently. Your files are encrypted **on your device**, before being sent. The server only receives opaque blobs. Your decryption key never leaves your machine.

This *zero-knowledge* model comes with a trade-off: if you lose your password without a recovery code, your files cannot be recovered. This is an intentional design decision, not a bug.

Kagibi is released under the **AGPLv3** license: the code is auditable, and self-hosting is fully supported.

---

## Features

### File Management

- **File upload** — drag-and-drop or file picker, with real-time progress (encryption phase then upload phase).
- **Folder upload** — upload an entire directory tree in one operation. When a name conflict arises, three options are available: auto-rename, skip, or replace.
- **Multipart upload** — large files are split into parts (5 to 100 MB each), individually encrypted with AES-256-GCM in the browser, then uploaded **in parallel** directly to S3 via pre-signed URLs (TTL 3 min). The backend never handles the raw content — it only orchestrates presigned URLs and finalises the multipart operation.
- **Download** — client-side streaming decryption: the file is never fully reconstructed in plaintext in memory before being written to disk.
- **Organization** — create folders, rename, move, delete (single file or recursive folder).
- **Tags** — label your folders to find them quickly via search and filters.
- **Preview** — view images and PDFs directly in the browser, without downloading.

### Search and Filtering

The global search bar (shortcut **Ctrl+K**) searches across all your files and folders.

- **In-context results** — clicking a result navigates directly to the file's location in the tree, with visual highlighting.
- **Available filters**:
  - By category (All, Documents, Images, Archives)
  - By extension (e.g. `.pdf`, `.mp4`)
  - By tag (labels applied to folders)
  - By element type (file or folder)
- **Note**: search is disabled when filename encryption is active, since stored names are opaque to the server.

### Sharing

Three sharing mechanisms are available, described in detail in the [Three Sharing Systems](#the-three-sharing-systems) section.

- **Link sharing** — public link (no account required), with the ability for visitors to upload files into a publicly shared folder.
- **Friend sharing** — granular permissions (download, create, delete, move), visual green/red management UI. Files uploaded by a friend are always recoverable by the owner via the folder key chain.
- **P2P transfer** — no server storage, end-to-end encrypted.

### P2P Transfer

Send a file directly from one device to another, end-to-end encrypted, without intermediate storage on our servers. See the dedicated section for details.

### Friends and Presence

- **Friend code** system (8 alphanumeric characters, e.g. `#A7KD92XZ`) for finding other users without exposing email addresses.
- Send and accept friend requests.
- Real-time **presence indicator** (green dot) with an 8-second grace period on disconnection to prevent status flickering.
- Mutual friend removal (automatically revokes associated shares).

### Account Security

- **Two-factor authentication (MFA)** — TOTP (authenticator app), with a 15-minute lockout after 5 failed attempts.
- **Recovery code** — generated at registration, allows regaining access to the master key if the password is lost.
- **Session revocation** — instantly disconnect all devices.
- **AAL2 elevation** — sensitive actions (password change, account deletion) require MFA confirmation even in an active session.

### GDPR Compliance

- **Right to erasure (Art. 17)** — account deletion triggers an immediate logical deletion, followed by permanent physical deletion (S3 blobs + database rows) after 30 days.
- **Right to portability (Art. 20)** — export all your data on request.

### Interface and Ergonomics

- **Light / dark theme**, toggleable in one click.
- **Multilingual interface**: French and English, with persistent preference.
- **Keyboard navigation**: Ctrl+K for search, arrow keys in lists.
- **Responsive design**: mobile-adapted navigation with a bottom bar and bottom sheets.
- **Storage quota** displayed in real time in the sidebar (updated within 2 seconds of each operation).
- **FAQ page** (`/faq`) — publicly accessible, covering general questions (sovereignty, security, encryption), features (P2P, sharing, Organisations, Friends), and Kagibi's values. Also accessible from the **Help & Support** menu in the dashboard navbar.

---

## Organizations

Organizations are end-to-end encrypted collaborative spaces. All files, folder names, and metadata stored within an organization are encrypted with a shared key that only members hold — the server has zero access to the content.

### Roles and Access

Each organization member has one of four roles:

| Role | Rights |
|------|--------|
| Owner | Full control: manage members, roles, quota, delete org |
| Admin | Manage members, provisioning, audit, invites |
| Member | Read and write files according to folder permissions |
| Viewer | Read-only access to permitted folders |

### End-to-End Encryption

Each organization uses a dedicated **OrgKey** (AES-256). This key is generated once by the owner, then individually re-encrypted for each member using their RSA-4096 public key before being stored server-side. This means:

- The server never holds the OrgKey in plaintext.
- Adding a new member requires an admin to **provision** their key: the admin decrypts the OrgKey locally and re-encrypts it with the new member's public key.
- Members who joined via invitation link cannot decrypt organization content until an admin provisions their key.

```
Owner's MasterKey
       │
       ▼
  OrgKey (AES-256)
       │
  ┌────┴─────────────┐
  │                  │
  RSA-OAEP(member1)  RSA-OAEP(member2) ...
  stored server-side per member
```

### Groups

Within an organization, **groups** allow clustering members to assign permissions collectively:

- Create, rename, and delete groups.
- Add or remove members from a group.
- Group members inherit a role within the group: **admin** or **member**.
- Folder-level permissions can be assigned to an entire group at once.

### LDAP / Active Directory Sync

Organizations can connect to a corporate **LDAP or Active Directory** server to automatically synchronize their members and groups, without managing invitations manually:

- **Automatic provisioning** — new LDAP users receive an invitation email and join the organization as soon as they accept it.
- **Group synchronization** — LDAP groups are recreated as Kagibi groups and their membership is updated at every sync cycle.
- **Two-phase deprovisioning** — a user who leaves the directory is first suspended, then automatically removed after a configurable grace period (or manually by an admin).
- **Safeguards** — the sync is aborted if LDAP returns an empty result or if more than 20% of existing members disappear in one cycle, protecting against filter errors and network failures.
- **Bind password encryption** — the service account password is stored AES-256-GCM encrypted.

Configuration is done in the **Administration → LDAP / AD** tab of the organization (restricted to admins and owners). See the [full LDAP documentation](../desktop-app/DOCUMENTATION_EN.md#10-ldap--active-directory-integration) for complete configuration and operational details.

### Onboarding Wizard

When a user creates their first organization, a step-by-step wizard guides them through:

1. Naming the organization and setting a description.
2. Understanding the encryption model and key provisioning workflow.
3. Creating the first invitation link for team members.

### File Management in Organizations

The organization file browser provides a full-featured interface for collaborative work:

- **Sorting** — by name, size, or date (ascending / descending).
- **Filtering** — by file type category (images, documents, videos, audio, archives) or by org tag.
- **Breadcrumb navigation** — clickable path with drag-and-drop support for moving items between folders.
- **Folder sizes** — total recursive size is computed and displayed for each folder.
- **Drag-and-drop** — drag files or folders onto another folder or breadcrumb segment to move them; drag from the OS to upload.
- **Bulk selection** — select multiple items via checkboxes or shift-click, then bulk download, move, or delete.
- **Rename** — inline rename with keyboard (Enter / Escape) and blur support.
- **File preview** — in-browser preview for images, PDFs, audio, and video files, without downloading.
- **ZIP download** — download an entire folder or a selection of files/folders as a ZIP archive (decrypted server-side before zipping, then streamed to the client).
- **Tags** — organization-wide tags (with colour) can be applied to any file or folder; filter by tag in the file browser.
- **Pinned items (Favorites)** — star frequently-accessed files and folders; they appear in a quick-access strip at the top of the browser.
- **Trash** — deleted items are moved to the trash, where they can be restored individually or permanently deleted; admins can empty the entire trash.
- **Search** — full-text search across organization file and folder names, with encrypted-name decryption before matching.
- **Upload progress** — per-file progress bars visible while files are being encrypted and uploaded.

### Folder-Level Access Control

Admins and group admins can define per-folder permissions for individual users or groups:

| Level | Description |
|-------|-------------|
| manage | Can read, write, and change permissions on the folder |
| write | Can upload, rename, move, and delete within the folder |
| read | Can browse and download from the folder |
| none | No access; folder is invisible |

Permissions cascade: a user's effective access to a folder is the highest level granted either directly or via any group they belong to.

### Organization Share Links

Organization files can be shared via public links, independently of the personal sharing system:

- **Generate a share link** for any file or folder within the organization.
- **Password protection** — optionally require a password to access the link.
- **Single-use option** — the link is automatically revoked after its first successful access.
- **Share management** — list all active org share links and revoke any of them.
- **Public access page** — recipients access the content at a dedicated URL; decryption happens client-side in their browser.

### Audit Log

Every action performed within an organization is recorded in an immutable audit log:

- Events include: file upload, download, deletion, rename, move, member join/leave, role change, permission change, share creation/revocation, key provisioning.
- **Encrypted fields** (file names, paths) are decrypted on the client before display.
- **Export** — admins can export the full audit log as a file.
- **Retention management** — admins can delete audit entries older than a chosen date.
- **Pagination** — load-more button for large audit histories.

### MFA Enforcement

Organization owners and admins can require all members to have MFA enabled before accessing the organization. Members without an active MFA setup see an access gate and are directed to their account settings.

### Dashboard and Statistics

The organization dashboard provides an overview for admins:

- Total member count, file count, folder count.
- Activity over the past 7 days.
- Number of active share links.
- Alert when members are missing an org key (not yet provisioned).
- Quick navigation to the provisioning workflow.

### Admin CLI

A command-line tool (`admin`) is available for server-side organization management:

```bash
./admin org create --name "Acme" --owner <user-id> --quota 10240
./admin org list
./admin org quota --id <org-id> --quota 20480
./admin org delete --id <org-id>
```

---

## How Encryption Works

### Key Derivation

When you create an account or log in, Kagibi derives a **master key** (MasterKey) from your password:

```
Password + random salt (16 bytes)
        │
        ▼
   Argon2id (64 MB memory, 4 iterations)
        │
        ▼
   KEK (Key Encryption Key) — stays in RAM, never leaves the browser
        │
        ▼
   MasterKey — derived, stored in RAM only
```

The **MasterKey** then encrypts all your files and metadata. The **KEK** wraps the MasterKey so it can be stored server-side in encrypted form (`EncryptedMasterKey`) — unusable without your password.

### File Encryption (upload)

Each file is split into **10 MB chunks**, each individually encrypted with **AES-256-GCM**:

```
Original file
        │
        ▼
  Split into 10 MB chunks
        │
        ▼
  For each chunk:
    ├── Unique nonce (8 random bytes + 4 counter bytes)
    ├── AES-256-GCM encrypt
    └── Stored format: [Nonce 12B][Ciphertext][Tag 16B]
        │
        ▼
  Direct upload to S3 via pre-signed URLs (TTL 180s)
  The backend orchestrates, but never touches the content.
```

### Streaming Decryption (download)

Downloads never reconstruct the entire file in memory:

```
Pre-signed S3 URL (TTL 5 min)
        │
        ▼
  ReadableStream (fetch)
        │
        ▼
  TransformStream: parse [Nonce][Ciphertext][Tag] → AES-GCM decrypt
        │
        ▼
  FileSystemWritableFileStream or Blob
  (never temporarily stored in decrypted form)
```

### What the Server Cannot Do

| Operation | Possible for the server? |
|-----------|--------------------------|
| Read file content | No — opaque blobs on S3 |
| Read a file name | No if the option is enabled — see below |
| Decrypt share data | No — keys encrypted with RSA-OAEP |
| Access your master key | No — never transmitted to the backend |
| Read organization content | No — OrgKey never stored in plaintext |

### Filename Encryption (opt-in)

At registration, you can enable encryption of file and folder names. This option is independent of content encryption (always active).

**When disabled (default)**: names are stored in plaintext in the database and S3 bucket. The search bar is functional.

**When enabled**:

```
Filename (e.g. "report.pdf")
        │
        ▼
  AES-256-GCM with MasterKey
  Random IV (12 bytes, CSPRNG)
        │
        ▼
  base64url encoding (no padding)
  → "aB3xK7mQ..." (opaque, no special characters)
        │
  ┌─────┴─────┐
  │           │
  PostgreSQL  OVH S3
  name = "aB3xK7..."   users/{id}/enc_path/aB3xK7...
```

- The browser decrypts names locally on every directory load.
- The search bar is disabled: since stored names are opaque blobs, a server-side `ILIKE` search has no effect.
- This choice is permanent and set at account creation.

---

## The Three Sharing Systems

### 1. Link Sharing

You generate a public link that anyone can open, without an account.

**How it works:**

1. Kagibi generates a random `ShareKey`, then encrypts the file key with it.
2. A random token (32 bytes) is created and associated with the link.
3. The link can be protected by a password (bcrypt-hashed) and/or time-limited (1 to 30 days).
4. The recipient visits the link; Kagibi returns the encrypted blob and the `ShareKey`.
5. Their browser decrypts the file locally.

When the link targets a **folder**, the public page also allows visitors to **upload files** into that folder. Files sent by visitors are encrypted in their browser using the `FolderKey` before being transferred to the owner's S3 storage. The server never has access to plaintext content.

The server stores: the token, the ShareKey-encrypted file key, the optional password hash, and the expiration date. It cannot read the file.

---

### 2. Friend Sharing (user-to-user)

Direct sharing between accounts uses asymmetric cryptography to ensure only the recipient can decrypt.

**How it works:**

1. At account creation, each user generates a **4096-bit RSA-OAEP** key pair.
   - The public key is stored in plaintext on the server.
   - The private key is encrypted with the MasterKey, then stored on the server.

2. To add a friend, you use their **friend code** (8 alphanumeric characters, e.g. `#A7KD92XZ`), unique per account.

3. To share a **file**:
   - Kagibi retrieves the recipient's RSA public key.
   - The `FileKey` (the file's AES key) is encrypted with that public key.
   - The encrypted result is stored in the database, linked to the share.

4. When the recipient accesses the file:
   - They retrieve the encrypted `FileKey`.
   - Their browser decrypts it using their RSA private key (itself decrypted with their MasterKey).
   - The file is decrypted locally.

5. To share a **folder** (with granular permissions):
   - The owner generates a `FolderKey` (AES-256), encrypted with their own MasterKey and stored server-side.
   - They configure the permissions granted to the friend.
   - The friend accesses folder contents according to the granted rights.

#### Folder share permissions

| Permission | Grants |
|------------|--------|
| Download | Access and download files |
| Create | Upload files and create sub-folders |
| Delete | Delete files in the shared folder |
| Move | Rename and move elements |

Default permissions when creating a new share: **Download + Create**.

Permissions are displayed with color coding in the share management dialog: **green** = granted, **red** = denied. Any attempt to perform an action without the required permission triggers an explicit error message.

Permissions are **editable at any time** after the share is created: clicking a chip toggles the right and syncs immediately with the server.

#### Key chain for friend-uploaded files

When a friend uploads a file into your shared folder, the file is encrypted with a key derived from the `FolderKey`. A dedicated backend endpoint allows the owner to recover the file key:

```
Owner's MasterKey
        │
        ▼
  Unwrap folder.encrypted_key  →  FolderKey
        │
        ▼
  Unwrap folder_file_key.encrypted_key  →  FileKey
        │
        ▼
  Decrypt file content
```

This chain guarantees the owner can always access files uploaded by friends, while the zero-knowledge guarantee is preserved.

The server stores: the encrypted `FileKey` (unusable without the recipient's private key), friendship relations, and permissions.

---

### Per-element restrictions in link shares

For folders shared via public link, it is possible to define access rights **per sub-element**, independently of the link's global permissions. A side panel in the share management dialog lets you navigate the shared folder tree and configure each entry individually.

#### Access levels for sub-folders

| Level | Behavior |
|-------|----------|
| Full access | The visitor sees and can interact with the folder per the global link permissions |
| Read-only | The visitor can browse the folder's content but cannot write to it |
| Hidden | The folder is invisible to the visitor |

For each file, two additional rights are independently configurable:
- **Download**: allow or block download (and preview) of that specific file.
- **Delete**: allow or protect that specific file from deletion.

#### Bulk controls

Bulk control buttons let you apply a uniform setting to all folders or all files at the current level in one click, then fine-tune element by element.

#### Tree navigation

The panel shows a clickable breadcrumb trail. You can drill down into any sub-folder to configure its restrictions, then navigate back up via the breadcrumb.

---

### Shares management view

The "Shares" page centralises all your active shares in two collapsible sections:

- **My shares** — deduplicated list of your shared resources with: type (file / folder), view counter, creation date, expiry date, one-click link copy, direct navigation to the shared folder in the file tree, and rights management.
- **Shared with me** — list of resources that other users have shared with you.

---

### 3. P2P Transfer (device-to-device)

P2P transfer sends files directly from one device to another, end-to-end encrypted, without intermediate storage on our servers. **There is no file size limit.**

WebRTC always attempts a **direct connection first** (LAN or NAT traversal via STUN). Only if a direct path cannot be established — due to a restrictive NAT or firewall — does the transfer fall back to a **Kagibi-operated TURN relay**. This relay is a pure traffic switcher: data enters and exits in real time without being written to disk. **The TURN server produces no logs and cannot access the content**, which remains AES-256-GCM encrypted end-to-end throughout.

#### Two transfer modes

**Direct mode (between registered friends)**

1. The sender selects an online friend and a file, then starts the transfer.
2. A random AES-256 file key is generated, encrypted with the recipient's RSA public key.
3. The WebRTC connection is negotiated over WebSocket (signals stored in `p2p_signals`).
4. Once the DataChannel is open, the file is sent in 16 KB chunks, each encrypted with a distinct random nonce.
5. The recipient receives a sound + visual notification, accepts, and their browser decrypts and reassembles the file locally.

**Invite mode (no account required)**

1. The sender generates an **invitation link** from the P2P page.
2. The link can be shared manually or sent by email (in French or English, your choice).
3. The recipient opens the link on `send.kagibi.cloud` — **no account is needed**.
4. They generate an ephemeral RSA key pair in their browser (never stored).
5. The sender is notified of the acceptance and starts the WebRTC transfer.
6. The invitation link is **single-use** and expires after 24 hours.

#### Information displayed during transfer

- **Progress** as a percentage with a visual bar.
- **Transfer speed** (e.g. `4.2 MB/s`) calculated in real time.
- **Estimated time remaining** (e.g. `~1m 30s`).
- **Connection type**: direct (LAN), via STUN (NAT traversal), or via TURN relay.
- **Re-notify**: the sender can send another sound alert to the recipient (up to 3 times, 30 s cooldown).
- **Manual leave** — the sender or recipient can manually close the connection at any point.

---

## What Is Stored on the Server

### Account Data

| Data | Format | Why |
|------|--------|-----|
| Email address | Encrypted (AES-256-GCM) | Authentication, without plaintext exposure |
| Display name | Plaintext | User interface |
| Password | bcrypt (cost 12) | Login verification |
| Argon2id salt | Random (16 bytes) | Client-side KEK derivation |
| `EncryptedMasterKey` | Encrypted (KEK) | MasterKey restoration at login |
| RSA public key | Plaintext | Encrypting incoming shares |
| `EncryptedPrivateKey` | Encrypted (MasterKey) | Decrypting received shares |
| Recovery code | SHA-256 (hash) | Password-less reset |
| Friend code | Plaintext | Friend discovery |

### File Metadata

| Data | Format |
|------|--------|
| Filename | Plaintext (default) or AES-GCM encrypted if enabled at registration |
| Size (bytes) | Plaintext |
| MIME type | Plaintext |
| Creation/modification dates | Plaintext |
| File key (`EncryptedKey`) | Encrypted (MasterKey or OrgKey) |

### Organization Data

| Data | Format |
|------|--------|
| Organization name | Plaintext |
| Member list and roles | Plaintext |
| OrgKey per member | Encrypted (member's RSA public key) |
| File and folder names within org | Encrypted (OrgKey, AES-256-GCM) |
| File content within org | Encrypted (OrgKey-derived key, AES-256-GCM) |
| Audit log entries | Plaintext actions; encrypted paths/names decrypted client-side |
| Folder permissions | Plaintext (user/group IDs + access level) |

### Social and Sharing Data

- Friend list and status (pending / accepted)
- Active shares: resource identifier + encrypted key + permissions
- Public links: token + encrypted key + expiration + optional password hash
- P2P invitations: token + file name + size + expiration date (content never stored)
- Organization share links: token + encrypted key + optional password hash + single-use flag

### Connection Logs (LCEN compliance)

In accordance with French law (LCEN Article 6 II and Decree 2021-1363), the following technical data is retained for **1 year**:

| Event logged | Data recorded |
|---|---|
| Account creation / deletion | User ID, full IP, anonymised IP, user-agent, timestamp |
| Login attempts (success / failure) | User ID, full IP, anonymised IP, timestamp |
| Password / token changes | User ID, full IP, anonymised IP, timestamp |
| File access | User ID, file ID, full IP, anonymised IP, timestamp |
| Public share link creation / revocation | User ID, resource, token, full IP, user-agent, timestamp |
| Direct share creation | Owner ID, recipient ID, resource, full IP, user-agent, timestamp |
| HTTP requests | Anonymised IP only (CNIL 2021-122), user-agent, status, duration |

**IP address policy**: full IP is stored only in security event logs (account lifecycle, auth, shares). Standard HTTP request logs contain only the anonymised IP (last IPv4 octet / last 80 IPv6 bits masked), in accordance with CNIL deliberation 2021-122. Logs are shipped to Grafana Cloud Loki and retained for 1 year.

All log data may be disclosed to judicial or administrative authorities upon a valid legal request. These logs are access-controlled and never contain decryption keys.

### What Is Not Collected

- File content (never in plaintext on the server)
- Browsing or search history within your files
- Full IP in standard HTTP logs (anonymised only — CNIL 2021-122)

---

## Account Recovery

A recovery code is generated at registration. It is separate from the password and allows regaining access to the MasterKey if the password is lost.

```
Recovery code (8 characters)
        │
        ├── SHA-256(code) → stored as RecoveryHash (verification)
        │
        └── Argon2id(code, recovery_salt) → decrypts EncryptedMasterKeyRecovery
```

If the recovery code is also lost, the data is **permanently inaccessible**. This is not a bug — it is the zero-knowledge guarantee.

---

## Data Deletion

- Account deletion triggers an immediate **logical deletion** (marking `deleted_at`).
- An asynchronous cleanup process performs **permanent physical deletion** after 30 days: database rows, S3 blobs.
- GDPR compliance (Articles 17 and 20): right to erasure and data portability.

---

## Tech Stack

| Component | Technology |
|-----------|------------|
| Frontend | Vue 3.5, Vite 7, Pinia |
| Backend | Go 1.21+, Gin |
| Database | PostgreSQL 16+ |
| Cache / rate-limiting | Redis 7+ |
| Object storage | OVH S3 (AWS-compatible) |
| Encryption | AES-256-GCM, RSA-OAEP 4096, Argon2id |
| Authentication | JWT HS256, TOTP (MFA) |
| P2P | WebRTC DataChannel, TURN/STUN (Coturn) |
| Deployment | Docker Compose (dev), Kubernetes / Rancher (prod) |

---

## Self-Hosting

**Prerequisites:** [Docker + Docker Compose](https://docs.docker.com/get-docker/) · an S3-compatible bucket (AWS, OVH, Scaleway, Cloudflare R2, MinIO…)

### 3 steps

**1. Clone and configure**

```bash
git clone https://github.com/Buuuntyyy/Kagibi.git
cd Kagibi
cp .env.example .env
```

**2. Edit `.env` — required values**

| Variable | Description |
|---|---|
| `JWT_SECRET` | Random secret key — `openssl rand -hex 32` |
| `EMAIL_ENCRYPTION_KEY` | AES-256 key for email encryption at rest — `openssl rand -hex 32` ⚠ never change after first use (existing encrypted emails would become unreadable) |
| `S3_ENDPOINT` | Your S3 endpoint URL (e.g. `https://s3.amazonaws.com`) |
| `S3_BUCKET` / `S3_REGION` | Your bucket name and region |
| `S3_ACCESS_KEY` / `S3_SECRET_KEY` | S3 credentials |
| `VITE_API_URL` | URL the browser uses to reach the backend — `http://localhost:8080/api/v1` locally, `https://your-domain.com/api/v1` in production |
| `ALLOWED_ORIGINS` | Frontend URL for CORS — `http://localhost` locally, `https://your-domain.com` in production |

**3. Start**

```bash
docker compose up -d
```

Startup order is managed automatically: PostgreSQL and Redis start first, the backend waits for them to be healthy, the frontend waits for the backend. Database migrations run automatically on first startup.

| Service | Default URL |
|---|---|
| Frontend | `http://localhost` |
| Backend API | `http://localhost:8080` |

All data is persisted in named Docker volumes (`db_data`, `redis_data`).

### Optional settings

All optional variables have safe defaults for self-hosting:

| Variable | Default | Description |
|---|---|---|
| `DB_PASSWORD` | `postgres` | PostgreSQL password (change it in production) |
| `AUTH_PROVIDER` | `local` | `local` (recommended), `supabase`, or `pocketbase` |
| `BILLING_ENABLED` | `false` | `false` = unlimited storage, no billing UI |
| `TURN_SERVER_URL` | — | P2P TURN relay for devices behind strict NAT |
| `GOOGLE_OAUTH_CLIENT_ID` | — | Enable Google Drive import |
| `FRONTEND_PORT` / `BACKEND_PORT` | `80` / `8080` | Change the exposed ports |

### Verify the deployment

```bash
docker compose ps                    # all 4 services should show "healthy"
curl http://localhost:8080/health    # → {"status":"ok"}
```

Open [http://localhost](http://localhost) in your browser.

To follow logs live:

```bash
docker compose logs -f backend   # backend only
docker compose logs -f           # all services
```

### Production (reverse proxy + TLS)

**Prerequisites:**
- A domain name with a DNS **A record** pointing to your server's public IP
- Ports **80** and **443** open on your server's firewall

**1. Update `.env`**

Free port 80 for the reverse proxy by changing `FRONTEND_PORT`, and set your domain:

```env
VITE_API_URL=https://your-domain.com/api/v1
ALLOWED_ORIGINS=https://your-domain.com
DB_PASSWORD=<strong-password>
BILLING_ENABLED=false
FRONTEND_PORT=3000
```

**2. Rebuild the frontend** (needed because `VITE_API_URL` is baked in at build time):

```bash
docker compose up -d --build frontend
```

**3. Configure the reverse proxy**

The reverse proxy must route:
- `/api/*` and `/health` → backend on port `8080`
- `/ws` → backend on port `8080` (WebSocket — requires upgrade headers)
- everything else → frontend on port `3000`

---

**Option A — Caddy (recommended, automatic TLS via Let's Encrypt)**

[Install Caddy](https://caddyserver.com/docs/install), then create `/etc/caddy/Caddyfile`:

```
your-domain.com {
    reverse_proxy /api/* localhost:8080
    reverse_proxy /health localhost:8080
    reverse_proxy /ws localhost:8080
    reverse_proxy * localhost:3000
}
```

```bash
systemctl reload caddy
```

Caddy obtains and renews the TLS certificate automatically. No further configuration needed.

---

**Option B — Nginx + Certbot**

```nginx
# /etc/nginx/sites-available/kagibi
server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl;
    server_name your-domain.com;

    ssl_certificate     /etc/letsencrypt/live/your-domain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;

    # Backend API + health
    location ~ ^/(api|health) {
        proxy_pass http://localhost:8080;
        proxy_set_header Host              $host;
        proxy_set_header X-Real-IP         $remote_addr;
        proxy_set_header X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto https;
    }

    # WebSocket (real-time presence, P2P signalling)
    location /ws {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade    $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host       $host;
        proxy_read_timeout 3600s;
    }

    # Frontend
    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host            $host;
        proxy_set_header X-Real-IP       $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

```bash
ln -s /etc/nginx/sites-available/kagibi /etc/nginx/sites-enabled/
certbot --nginx -d your-domain.com   # issues and configures TLS automatically
nginx -s reload
```

---

## Local development

**Prerequisites:** Go 1.21+, Node.js 18+, Docker

```bash
git clone https://github.com/Buuuntyyy/Kagibi.git
cd Kagibi

# Infrastructure only (PostgreSQL + Redis)
docker compose up db redis -d

# Backend
cp backend/.env.example backend/.env   # fill in S3_*, JWT_SECRET, EMAIL_ENCRYPTION_KEY
cd backend && go run .

# Frontend (new terminal)
cd frontend && npm install && npm run dev
```

Frontend: `http://localhost:5173` — Backend: `http://localhost:8080`

---

## License

AGPLv3 — see [`LICENSE`](../LICENSE).

Any modification of the code, including in a SaaS context, must be published under the same license.
