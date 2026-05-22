# Organisation E2E Encryption — Kagibi

---

🇫🇷 [Français](#français) · 🇬🇧 [English](#english)

---

## Français

### Vue d'ensemble

Chaque organisation Kagibi bénéficie d'un chiffrement de bout en bout (E2E) identique à celui des espaces personnels. Les fichiers sont chiffrés **sur le client** avant d'être envoyés vers S3. Le serveur ne stocke que des blobs opaques — il ne peut reconstituer aucune clé, ni lire aucun fichier.

Les primitives cryptographiques respectent les recommandations NIST SP 800-38D, ANSSI et OWASP.

---

### Hiérarchie des clés

```
Mot de passe utilisateur
    │
    └─ Argon2id (64 Mo RAM, 4 passes, sel 16B)
            │
            └─ KEK (AES-256-GCM)
                    │
                    └─ déchiffre ──► MasterKey (AES-256-GCM)
                                            │
                                            └─ déchiffre ──► RSA Private Key (RSA-4096, PKCS8)
                                                                    │
                                                                    └─ RSA-OAEP SHA-256
                                                                            │
                                                                            ▼
                                                                   OrgKey (AES-256-GCM)
                                                                   une par organisation
                                                                            │
                                                                   AES-GCM key wrap
                                                                            │
                                                                            ▼
                                                                   FileKey (AES-256-GCM)
                                                                   une par fichier
                                                                            │
                                                                   AES-GCM, chunks 10 Mo
                                                                   Nonce NIST SP 800-38D
                                                                            │
                                                                            ▼
                                                                   Contenu du fichier (S3)
```

---

### Algorithmes et paramètres

| Composant | Algorithme | Taille | Standard |
|---|---|---|---|
| Chiffrement contenu | AES-GCM | 256 bits | NIST SP 800-38D |
| Chiffrement/wrapping de clés | AES-GCM | 256 bits | NIST SP 800-38D |
| Distribution de clé org | RSA-OAEP | 4096 bits, SHA-256 | PKCS#1 v2.2 |
| Dérivation de clé (mot de passe) | Argon2id | 32 octets | RFC 9106 |
| Tag d'authentification GCM | — | 128 bits | NIST recommandation |
| Nonce (IV) | CSPRNG | 96 bits (12 octets) | NIST SP 800-38D §8.2.2 |
| Nonce de chunk | [8B aléatoire] + [4B compteur] | 96 bits | Déterministe, unique par fichier |
| Taille des chunks | — | 10 Mo | — |

Le nonce de chunk est construit comme suit : les 8 premiers octets sont aléatoires et communs à tous les chunks du fichier ; les 4 derniers encodent l'index du chunk en little-endian. Cela garantit l'unicité sans état persistant, et supporte jusqu'à 2³² chunks par fichier (≈ 40 Po à 10 Mo par chunk).

---

### Cycle de vie de l'OrgKey

#### Naissance

| Scénario | Moment | Qui génère la clé |
|---|---|---|
| Création via l'application | À la création de l'org | L'utilisateur créateur |
| Org provisionnée par le CLI admin | À l'acceptation de l'invitation owner | Le premier propriétaire, sur la page `/join/:token` |

Dans les deux cas, l'OrgKey est générée entièrement côté client (Web Crypto API). Elle n'est jamais envoyée en clair au serveur.

#### Distribution aux membres

```
Invitation directe (target_user_id connu)
    └─ L'invitant chiffre l'OrgKey avec la clé publique RSA de la cible
    └─ Stocké dans OrgInvitation.encrypted_org_key
    └─ Copié dans OrgMember.encrypted_org_key lors de l'acceptation

Invitation par lien (token public)
    └─ Le membre rejoint sans clé (encrypted_org_key vide)
    └─ Un admin/owner doit "Activer l'accès" dans l'onglet Membres
    └─ L'admin chiffre l'OrgKey avec la clé publique RSA du membre
    └─ PATCH /api/v1/orgs/:id/members/:memberID/key
```

Chaque membre possède **sa propre copie** de l'OrgKey, chiffrée avec **sa propre clé publique RSA**. Le serveur stocke N blobs RSA distincts pour N membres — il ne peut en dériver l'OrgKey depuis aucun d'eux.

---

### Flux d'upload

```
Client                                    Backend / S3
  │                                            │
  ├─ 1. getOrgKey(orgID)                       │
  │      └─ lit my_encrypted_org_key           │
  │      └─ RSA-OAEP déchiffre avec            │
  │         authStore.privateKey               │
  │      └─ cache en mémoire (hors Vue)        │
  │                                            │
  ├─ 2. generateMasterKey()   ← FileKey        │
  │                                            │
  ├─ 3. wrapMasterKey(fileKey, orgKey)         │
  │      → encrypted_key (base64)             │
  │                                            │
  ├─ 4. Pour chaque chunk de 10 Mo :           │
  │      encryptChunkWorker(chunk, fileKey,    │
  │        index, baseNonce)                   │
  │      → [Nonce 12B][Ciphertext][Tag 16B]    │
  │                                            │
  ├─ 5. POST /fs/multipart/initiate ──────────►│ retourne presigned URLs
  │      { encrypted_key, total_size, ... }    │
  │                                            │
  ├─ 6. PUT presigned_url ────────────────────►│ S3 (jamais en clair)
  │      (chunk chiffré)                       │
  │                                            │
  └─ 7. POST /fs/multipart/complete ──────────►│ stocke encrypted_key en base
         { encrypted_key, parts, ... }         │
```

---

### Flux de téléchargement

```
Client                                    Backend / S3
  │                                            │
  ├─ 1. encrypted_key ← listing ou            │
  │      GET /fs/file/:id/key                 │
  │                                            │
  ├─ 2. getOrgKey(orgID)   ← cache ou RSA     │
  │                                            │
  ├─ 3. unwrapMasterKey(encrypted_key, orgKey) │
  │      → FileKey (AES-256-GCM)              │
  │                                            │
  ├─ 4. GET /fs/file/:id/download ────────────►│ S3 retourne blob chiffré
  │                                            │
  ├─ 5. decryptChunkedFileWorker(blob, fileKey)│
  │      → Blob plaintext (Worker Pool)        │
  │                                            │
  └─ 6. URL.createObjectURL → download         │
```

---

### Ce que le serveur voit (et ne voit pas)

| Donnée stockée | Format | Le serveur peut-il déchiffrer ? |
|---|---|---|
| `OrgMember.encrypted_org_key` | RSA-OAEP, base64 | ✗ — nécessite la clé privée RSA de l'utilisateur |
| `OrgFile.encrypted_key` | AES-GCM wrap, base64 | ✗ — nécessite l'OrgKey |
| `OrgInvitation.encrypted_org_key` | RSA-OAEP, base64 | ✗ |
| Contenu des fichiers (S3) | Chunks AES-GCM | ✗ — nécessite la FileKey |
| `User.public_key` | PEM RSA-4096 | ✓ (clé publique par définition) |
| `User.encrypted_private_key` | AES-GCM, base64 | ✗ — nécessite la MasterKey |
| `User.encrypted_master_key` | AES-GCM, base64 | ✗ — nécessite la KEK dérivée du mot de passe |

La chaîne de confiance ne peut être remontée sans le mot de passe utilisateur, qui n'est jamais transmis au serveur.

---

### Provisioning de clé (invitation par lien)

Quand un membre rejoint via un lien public, son `encrypted_org_key` est vide. L'onglet **Membres** affiche un badge rouge et un bouton **Activer l'accès** visible uniquement par les admins et le propriétaire.

Lorsque l'admin clique sur ce bouton :
1. Son client récupère l'OrgKey depuis le cache (ou la déchiffre avec sa clé privée RSA).
2. Il chiffre l'OrgKey avec la **clé publique RSA du membre** (`member.public_key`, renvoyée par `ListMembers`).
3. Il envoie le résultat via `PATCH /api/v1/orgs/:id/members/:memberID/key`.

Jusqu'à ce provisioning, le membre peut voir la liste des fichiers mais ne peut pas les télécharger ni en uploader.

---

### Implémentation — où trouver le code

| Composant | Fichier |
|---|---|
| Primitives crypto org | `frontend/src/utils/orgCrypto.js` |
| Primitives crypto communes | `frontend/src/utils/crypto.js` |
| Chiffrement en Worker | `frontend/src/workers/crypto.worker.js` |
| Pool de Workers | `frontend/src/workers/cryptoWorkerPool.js` |
| Store Pinia avec intégration crypto | `frontend/src/stores/organizations.js` |
| Page d'acceptation d'invitation | `frontend/src/views/JoinView.vue` |
| Vue org avec upload/download | `frontend/src/views/OrgDetailView.vue` |
| Backend — membres (clé publique) | `backend/handlers/organizations/members.go` |
| Backend — org (clé chiffrée membre) | `backend/handlers/organizations/list.go` |
| Backend — provisioning de clé | `backend/handlers/organizations/members.go` (`SetMemberKey`) |

---

---

## English

### Overview

Every Kagibi organisation benefits from end-to-end encryption (E2E) identical to personal file spaces. Files are encrypted **on the client** before being sent to S3. The server only stores opaque blobs — it cannot reconstruct any key, nor read any file.

The cryptographic primitives follow NIST SP 800-38D, ANSSI and OWASP recommendations.

---

### Key hierarchy

```
User password
    │
    └─ Argon2id (64 MB RAM, 4 passes, 16B salt)
            │
            └─ KEK (AES-256-GCM)
                    │
                    └─ decrypts ──► MasterKey (AES-256-GCM)
                                            │
                                            └─ decrypts ──► RSA Private Key (RSA-4096, PKCS8)
                                                                    │
                                                                    └─ RSA-OAEP SHA-256
                                                                            │
                                                                            ▼
                                                                   OrgKey (AES-256-GCM)
                                                                   one per organisation
                                                                            │
                                                                   AES-GCM key wrap
                                                                            │
                                                                            ▼
                                                                   FileKey (AES-256-GCM)
                                                                   one per file
                                                                            │
                                                                   AES-GCM, 10 MB chunks
                                                                   NIST SP 800-38D nonce
                                                                            │
                                                                            ▼
                                                                   File content (S3)
```

---

### Algorithms and parameters

| Component | Algorithm | Size | Standard |
|---|---|---|---|
| Content encryption | AES-GCM | 256 bits | NIST SP 800-38D |
| Key encryption / wrapping | AES-GCM | 256 bits | NIST SP 800-38D |
| Org key distribution | RSA-OAEP | 4096 bits, SHA-256 | PKCS#1 v2.2 |
| Key derivation (password) | Argon2id | 32 bytes | RFC 9106 |
| GCM authentication tag | — | 128 bits | NIST recommendation |
| Nonce (IV) | CSPRNG | 96 bits (12 bytes) | NIST SP 800-38D §8.2.2 |
| Chunk nonce | [8B random] + [4B counter] | 96 bits | Deterministic, unique per file |
| Chunk size | — | 10 MB | — |

The chunk nonce is constructed as follows: the first 8 bytes are random and shared across all chunks of the file; the last 4 encode the chunk index in little-endian order. This guarantees uniqueness without persistent state and supports up to 2³² chunks per file (≈ 40 PB at 10 MB per chunk).

---

### OrgKey lifecycle

#### Birth

| Scenario | When | Who generates the key |
|---|---|---|
| Created via the application | At org creation | The creating user |
| Org provisioned by admin CLI | When owner accepts the invitation | The first owner, on the `/join/:token` page |

In both cases, the OrgKey is generated entirely client-side (Web Crypto API). It is never sent to the server in plaintext.

#### Distribution to members

```
Direct invitation (target_user_id known)
    └─ The inviter encrypts the OrgKey with the target's RSA public key
    └─ Stored in OrgInvitation.encrypted_org_key
    └─ Copied into OrgMember.encrypted_org_key upon acceptance

Link invitation (public token)
    └─ The member joins without a key (encrypted_org_key empty)
    └─ An admin/owner must click "Provision access" in the Members tab
    └─ The admin encrypts the OrgKey with the member's RSA public key
    └─ PATCH /api/v1/orgs/:id/members/:memberID/key
```

Each member holds **their own copy** of the OrgKey, encrypted with **their own RSA public key**. The server stores N distinct RSA blobs for N members — it cannot derive the OrgKey from any of them.

---

### Upload flow

```
Client                                    Backend / S3
  │                                            │
  ├─ 1. getOrgKey(orgID)                       │
  │      └─ reads my_encrypted_org_key         │
  │      └─ RSA-OAEP decrypts with             │
  │         authStore.privateKey               │
  │      └─ caches in memory (outside Vue)     │
  │                                            │
  ├─ 2. generateMasterKey()   ← FileKey        │
  │                                            │
  ├─ 3. wrapMasterKey(fileKey, orgKey)         │
  │      → encrypted_key (base64)             │
  │                                            │
  ├─ 4. For each 10 MB chunk:                 │
  │      encryptChunkWorker(chunk, fileKey,    │
  │        index, baseNonce)                   │
  │      → [Nonce 12B][Ciphertext][Tag 16B]    │
  │                                            │
  ├─ 5. POST /fs/multipart/initiate ──────────►│ returns presigned URLs
  │      { encrypted_key, total_size, ... }    │
  │                                            │
  ├─ 6. PUT presigned_url ────────────────────►│ S3 (never plaintext)
  │      (encrypted chunk)                     │
  │                                            │
  └─ 7. POST /fs/multipart/complete ──────────►│ stores encrypted_key in DB
         { encrypted_key, parts, ... }         │
```

---

### Download flow

```
Client                                    Backend / S3
  │                                            │
  ├─ 1. encrypted_key ← listing or            │
  │      GET /fs/file/:id/key                 │
  │                                            │
  ├─ 2. getOrgKey(orgID)  ← cache or RSA      │
  │                                            │
  ├─ 3. unwrapMasterKey(encrypted_key, orgKey) │
  │      → FileKey (AES-256-GCM)              │
  │                                            │
  ├─ 4. GET /fs/file/:id/download ────────────►│ S3 returns encrypted blob
  │                                            │
  ├─ 5. decryptChunkedFileWorker(blob, fileKey)│
  │      → plaintext Blob (Worker Pool)        │
  │                                            │
  └─ 6. URL.createObjectURL → download         │
```

---

### What the server sees (and does not see)

| Stored data | Format | Can the server decrypt it? |
|---|---|---|
| `OrgMember.encrypted_org_key` | RSA-OAEP, base64 | ✗ — requires the user's RSA private key |
| `OrgFile.encrypted_key` | AES-GCM wrap, base64 | ✗ — requires the OrgKey |
| `OrgInvitation.encrypted_org_key` | RSA-OAEP, base64 | ✗ |
| File content (S3) | AES-GCM chunks | ✗ — requires the FileKey |
| `User.public_key` | PEM RSA-4096 | ✓ (public key by definition) |
| `User.encrypted_private_key` | AES-GCM, base64 | ✗ — requires the MasterKey |
| `User.encrypted_master_key` | AES-GCM, base64 | ✗ — requires the KEK derived from the password |

The trust chain cannot be traced back without the user's password, which is never transmitted to the server.

---

### Key provisioning (link invitation)

When a member joins via a public link, their `encrypted_org_key` is empty. The **Members** tab displays a red badge and a **Provision access** button, visible only to admins and the owner.

When the admin clicks the button:
1. Their client retrieves the OrgKey from cache (or decrypts it with their RSA private key).
2. It encrypts the OrgKey with the **member's RSA public key** (`member.public_key`, returned by `ListMembers`).
3. It submits the result via `PATCH /api/v1/orgs/:id/members/:memberID/key`.

Until this provisioning step, the member can see the file listing but cannot download or upload files.

---

### Implementation — where to find the code

| Component | File |
|---|---|
| Org crypto primitives | `frontend/src/utils/orgCrypto.js` |
| Shared crypto primitives | `frontend/src/utils/crypto.js` |
| Worker encryption | `frontend/src/workers/crypto.worker.js` |
| Worker pool | `frontend/src/workers/cryptoWorkerPool.js` |
| Pinia store with crypto integration | `frontend/src/stores/organizations.js` |
| Invitation acceptance page | `frontend/src/views/JoinView.vue` |
| Org view (upload / download) | `frontend/src/views/OrgDetailView.vue` |
| Backend — members (public key) | `backend/handlers/organizations/members.go` |
| Backend — org (member encrypted key) | `backend/handlers/organizations/list.go` |
| Backend — key provisioning | `backend/handlers/organizations/members.go` (`SetMemberKey`) |
