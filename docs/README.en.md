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
- **Multipart upload** — large files (> 10 MB) are split into 10 MB chunks, individually encrypted and uploaded in parallel.
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

### 3. P2P Transfer (device-to-device)

P2P transfer sends files directly from one device to another, end-to-end encrypted, without intermediate storage on our servers.

Modern networks sometimes make direct connections impossible (NAT, firewalls). In those cases, Kagibi uses a TURN relay server owned by Kagibi. **Data passing through this relay remains AES-GCM encrypted — the server sees only opaque streams, not the content.**

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
| File key (`EncryptedKey`) | Encrypted (MasterKey) |

### Social and Sharing Data

- Friend list and status (pending / accepted)
- Active shares: resource identifier + encrypted key + permissions
- Public links: token + encrypted key + expiration + optional password hash
- P2P invitations: token + file name + size + expiration date (content never stored)

### What Is Not Collected

- File content (never in plaintext on the server)
- Browsing or search history
- IP addresses (except temporary logging for security/abuse purposes)
- Device or browser information

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

## Quick Start (development)

**Prerequisites:** Docker, Docker Compose

```bash
git clone https://github.com/Buuuntyyy/Kagibi.git
cd Kagibi
cp backend/.env.example backend/.env   # Fill in S3 variables, JWT_SECRET, etc.
cp frontend/.env.example frontend/.env # Fill in VITE_BACKEND_URL=http://localhost:8080

cd backend
go run main.go

cd frontend
npm install
npm run dev
```

Frontend: `http://localhost` — Backend: `http://localhost:8080`

For detailed configuration (environment variables, S3, Kubernetes), see [`backend/README.md`](../backend/README.md).

---

## License

AGPLv3 — see [`LICENSE`](../LICENSE).

Any modification of the code, including in a SaaS context, must be published under the same license.
