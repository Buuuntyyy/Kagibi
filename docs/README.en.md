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

The server stores: the token, the ShareKey-encrypted file key, the optional password hash, and the expiration date. It cannot read the file.

---

### 2. Friend Sharing (user-to-user)

Direct sharing between accounts uses asymmetric cryptography to ensure only the recipient can decrypt.

**How it works:**

1. At account creation, each user generates a **4096-bit RSA-OAEP** key pair.
   - The public key is stored in plaintext on the server.
   - The private key is encrypted with the MasterKey, then stored on the server.

2. To add a friend, you use their **friend code** (8 alphanumeric characters, e.g. `#A7KD92XZ`), unique per account.

3. To share a file:
   - Kagibi retrieves the recipient's RSA public key.
   - The `FileKey` (the file's AES key) is encrypted with that public key.
   - The encrypted result is stored in the database, linked to the share.

4. When the recipient accesses the file:
   - They retrieve the encrypted `FileKey`.
   - Their browser decrypts it using their RSA private key (itself decrypted with their MasterKey).
   - The file is decrypted locally.

The server stores: the encrypted `FileKey` (unusable without the recipient's private key), friendship relations, and permissions.

---

### 3. P2P Transfer (device-to-device)

P2P transfer sends files directly from one device to another, bypassing server storage.
**However, modern internet networks often make direct connections impossible (NAT, firewalls). Kagibi uses a TURN server to relay data when necessary, while maintaining end-to-end encryption.**
In practice, devices establish a WebRTC DataChannel connection, and data is encrypted with AES-GCM before being sent. The server only sees encrypted data streams, even during TURN relay.

**How it works:**

1. Both devices establish a **WebRTC DataChannel** connection via a signaling server (WebSocket).
2. The backend relays only WebRTC negotiation messages (offer/answer/ICE candidates), temporarily stored in the `p2p_signals` table for offline delivery.
3. Once the peer-to-peer connection is established, data flows **directly** between devices, encrypted at the application layer with AES-GCM.
4. The server sees neither the transferred content nor the file metadata.

P2P transfers count against the user's plan (`p2p_max_exchanges`).

---

## What Is Stored on the Server

### Account Data

| Data | Format | Why |
|------|--------|-----|
| Email address | Plaintext | Authentication |
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
- Recent activity (accessed files — optional)

### What Is Not Collected

- File content (never in plaintext on the server)
- Browsing or search history
- IP addresses (except temporary logging for security/abuse purposes)
- Analytics or tracking pixels
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
cp backend/.env.example backend/.env # Fill in S3 variables, JWT_SECRET, etc.
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
