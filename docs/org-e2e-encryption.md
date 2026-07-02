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
                                                                   ┌──────────┴──────────┐
                                                                   │                     │
                                                          AES-GCM key wrap        AES-GCM key wrap
                                                                   │                     │
                                                                   ▼                     ▼
                                                          FileKey               GroupKey (AES-256-GCM)
                                                          (fichiers sans         une par groupe
                                                           groupe)               ┌───────┴──────────┐
                                                                                 │                  │
                                                                        RSA-OAEP (membre)   AES-GCM key wrap
                                                                        stocké par membre         │
                                                                                                   ▼
                                                                                          FileKey (AES-256-GCM)
                                                                                          une par fichier de groupe
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

### Algorithmes supplémentaires pour les GroupKeys

| Composant | Algorithme | Taille | Détail |
|---|---|---|---|
| GroupKey (backup admin) | AES-GCM key wrap | 256 bits | GroupKey wrappée avec l'OrgKey |
| GroupKey (par membre) | RSA-OAEP SHA-256 | 4096 bits | GroupKey chiffrée avec la clé publique RSA du membre |
| FileKey de groupe | AES-GCM key wrap | 256 bits | FileKey wrappée avec la GroupKey au lieu de l'OrgKey |

Stockage côté serveur :

| Table | Champ | Contenu |
|---|---|---|
| `org_groups` | `encrypted_group_key` | GroupKey wrappée avec l'OrgKey (backup admin) |
| `org_group_keys` | `encrypted_key` | GroupKey wrappée avec la clé publique RSA de chaque membre |
| `org_files` | `group_id` | NULL → FileKey wrappée avec OrgKey ; non NULL → FileKey wrappée avec GroupKey |

---

### Chiffrement par groupe (GroupKey)

Les **groupes** peuvent bénéficier d'un chiffrement dédié : chaque groupe dispose alors de sa propre **GroupKey** (AES-256-GCM). Les fichiers stockés dans un dossier lié au groupe sont chiffrés avec la GroupKey et non avec l'OrgKey. Cela isole cryptographiquement le contenu d'un groupe des autres membres de l'organisation.

#### Initialisation

1. Un admin org génère une GroupKey aléatoire via `generateOrgKey()` (Web Crypto API).
2. Il wrapp la GroupKey avec l'OrgKey (`wrapFileKey(groupKey, orgKey)`) → stocké dans `org_groups.encrypted_group_key` (backup admin).
3. Pour chaque membre du groupe ayant une clé publique RSA, il chiffre la GroupKey avec `encryptOrgKeyForUser(groupKey, member.public_key)` → stocké dans `org_group_keys.encrypted_key`.
4. Appel `POST /orgs/:orgID/groups/:groupID/key/init` pour persister le tout en transaction.

#### Distribution à un nouveau membre du groupe

Après l'ajout d'un membre au groupe, un admin doit provisionner sa GroupKey :

```
Admin                                   Backend
  │                                         │
  ├─ getGroupKey(orgID, groupID)             │
  │    └─ via org_group_keys (RSA path) ou  │
  │       via encrypted_group_key (admin)   │
  │                                         │
  ├─ encryptOrgKeyForUser(groupKey, member.public_key)
  │                                         │
  └─ PUT /orgs/:orgID/groups/:groupID/key/members/:userID
```

#### Accès admin sans entrée org_group_keys

L'admin peut toujours récupérer la GroupKey via le chemin OrgKey :

```
admin.orgKey ──► unwrapFileKey(group.encrypted_group_key, orgKey) ──► GroupKey
```

Cela permet à l'admin de provisionner de nouveaux membres sans avoir besoin d'une entrée personnelle dans `org_group_keys`.

#### Rotation de clé

La rotation remplace la GroupKey et re-wrappe tous les fichiers du groupe :

1. Génération d'une nouvelle GroupKey.
2. Re-wrapping de tous les fichiers du groupe : `unwrap(oldKey) → wrap(newKey)`.
3. Reprovisionnement de tous les membres (RSA path).
4. Appel atomique `POST /orgs/:orgID/groups/:groupID/key/rotate` (transaction DB).

Lors d'une rotation de la **OrgKey**, les fichiers avec `group_id != NULL` sont ignorés — ils sont gérés par la rotation de leur GroupKey respective.

#### Révocation d'accès

Retirer un membre du groupe supprime immédiatement son entrée `org_group_keys`. Il ne peut plus déchiffrer les nouveaux fichiers. Les anciens fichiers déjà téléchargés (s'ils sont en cache local) restent accessibles — il est recommandé de faire une rotation de GroupKey après retrait d'un membre sensible.

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
  ├─ 2. current_group_id ?                     │
  │      ├─ OUI : getGroupKey(orgID, groupID)  │
  │      │        (RSA path ou org-key path)   │
  │      │        wrappingKey = groupKey        │
  │      └─ NON : wrappingKey = orgKey         │
  │                                            │
  ├─ 3. generateMasterKey()   ← FileKey        │
  │                                            │
  ├─ 4. wrapFileKey(fileKey, wrappingKey)      │
  │      → encrypted_key (base64)             │
  │                                            │
  ├─ 5. Pour chaque chunk de 10 Mo :           │
  │      encryptChunkWorker(chunk, fileKey,    │
  │        index, baseNonce)                   │
  │      → [Nonce 12B][Ciphertext][Tag 16B]    │
  │                                            │
  ├─ 6. POST /fs/multipart/initiate ──────────►│ retourne presigned URLs
  │      { encrypted_key, group_id, ... }      │
  │                                            │
  ├─ 7. PUT presigned_url ────────────────────►│ S3 (jamais en clair)
  │      (chunk chiffré)                       │
  │                                            │
  └─ 8. POST /fs/multipart/complete ──────────►│ stocke encrypted_key + group_id
         { encrypted_key, group_id, parts, ... }│
```

---

### Flux de téléchargement

```
Client                                    Backend / S3
  │                                            │
  ├─ 1. encrypted_key + group_id ← listing    │
  │      ou GET /fs/file/:id/key              │
  │      (retourne encrypted_key + group_id)  │
  │                                            │
  ├─ 2. group_id ?                             │
  │      ├─ OUI : getGroupKey(orgID, group_id) │
  │      │        wrappingKey = groupKey       │
  │      └─ NON : getOrgKey(orgID)            │
  │               wrappingKey = orgKey        │
  │                                            │
  ├─ 3. unwrapFileKey(encrypted_key,           │
  │      wrappingKey) → FileKey (AES-256-GCM) │
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
| `OrgFile.encrypted_key` (sans groupe) | AES-GCM wrap, base64 | ✗ — nécessite l'OrgKey |
| `OrgFile.encrypted_key` (avec groupe) | AES-GCM wrap, base64 | ✗ — nécessite la GroupKey |
| `OrgGroup.encrypted_group_key` | AES-GCM wrap, base64 | ✗ — nécessite l'OrgKey |
| `OrgGroupKey.encrypted_key` | RSA-OAEP, base64 | ✗ — nécessite la clé privée RSA du membre |
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

### Provisioning de clé de groupe — résumé des endpoints

| Endpoint | Rôle |
|---|---|
| `GET /orgs/:id/groups/:groupID/key` | Retourne la GroupKey chiffrée (RSA ou admin fallback) |
| `POST /orgs/:id/groups/:groupID/key/init` | Initialise la GroupKey (admin, idempotent via 409) |
| `PUT /orgs/:id/groups/:groupID/key/members/:userID` | Provisionne un membre individuel |
| `GET /orgs/:id/groups/:groupID/key/members` | Liste les `user_id` ayant une clé provisionnée (admin) |
| `POST /orgs/:id/groups/:groupID/key/rotate` | Rotation atomique (membres + fichiers) |
| `GET /orgs/:id/groups/:groupID/key/files` | Retourne toutes les FileKeys du groupe (admin, pour rotation) |

### Contrôle d'accès effectif

| Endpoint | Rôle |
|---|---|
| `GET /orgs/:id/my-access` | Accès effectif de l'utilisateur courant (tous les dossiers) |
| `GET /orgs/:id/members/:userID/effective-access` | Accès effectif d'un membre vu par un admin |

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
| Panneau de gestion des groupes | `frontend/src/components/organizations/OrgGroupsPanel.vue` |
| Backend — membres (clé publique) | `backend/handlers/organizations/members.go` |
| Backend — org (clé chiffrée membre) | `backend/handlers/organizations/list.go` |
| Backend — provisioning de clé OrgKey | `backend/handlers/organizations/members.go` (`SetMemberKey`) |
| Backend — GroupKey (init/rotate/provision) | `backend/handlers/organizations/orggroupkey.go` |
| Backend — accès effectif | `backend/handlers/organizations/orgeffectiveaccess.go` |
| Backend — demandes d'accès | `backend/handlers/organizations/orgaccessrequests.go` |

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
                                                                   ┌──────────┴──────────┐
                                                                   │                     │
                                                          AES-GCM key wrap        AES-GCM key wrap
                                                                   │                     │
                                                                   ▼                     ▼
                                                          FileKey               GroupKey (AES-256-GCM)
                                                          (ungrouped files)      one per group
                                                                                 ┌───────┴──────────┐
                                                                                 │                  │
                                                                        RSA-OAEP (member)   AES-GCM key wrap
                                                                        stored per member          │
                                                                                                   ▼
                                                                                          FileKey (AES-256-GCM)
                                                                                          one per group file
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

### Additional algorithms for GroupKeys

| Component | Algorithm | Size | Detail |
|---|---|---|---|
| GroupKey (admin backup) | AES-GCM key wrap | 256 bits | GroupKey wrapped with OrgKey |
| GroupKey (per member) | RSA-OAEP SHA-256 | 4096 bits | GroupKey encrypted with member's RSA public key |
| Group FileKey | AES-GCM key wrap | 256 bits | FileKey wrapped with GroupKey instead of OrgKey |

Server-side storage:

| Table | Field | Content |
|---|---|---|
| `org_groups` | `encrypted_group_key` | GroupKey wrapped with OrgKey (admin backup) |
| `org_group_keys` | `encrypted_key` | GroupKey wrapped with each member's RSA public key |
| `org_files` | `group_id` | NULL → FileKey wrapped with OrgKey; non-NULL → FileKey wrapped with GroupKey |

---

### Group encryption (GroupKey)

**Groups** can have dedicated encryption: each group gets its own **GroupKey** (AES-256-GCM). Files in folders linked to that group are encrypted with the GroupKey rather than the OrgKey, cryptographically isolating the group's content from other org members.

#### Initialisation

1. An org admin generates a random GroupKey via `generateOrgKey()` (Web Crypto API).
2. The GroupKey is wrapped with the OrgKey (`wrapFileKey(groupKey, orgKey)`) → stored in `org_groups.encrypted_group_key` (admin backup).
3. For each group member with an RSA public key, the GroupKey is encrypted with `encryptOrgKeyForUser(groupKey, member.public_key)` → stored in `org_group_keys.encrypted_key`.
4. `POST /orgs/:orgID/groups/:groupID/key/init` persists everything in a single transaction.

#### Provisioning a new group member

After adding a member to a group, an admin must provision their GroupKey:

```
Admin                                   Backend
  │                                         │
  ├─ getGroupKey(orgID, groupID)             │
  │    └─ via org_group_keys (RSA path) or  │
  │       via encrypted_group_key (admin)   │
  │                                         │
  ├─ encryptOrgKeyForUser(groupKey, member.public_key)
  │                                         │
  └─ PUT /orgs/:orgID/groups/:groupID/key/members/:userID
```

#### Admin access without an org_group_keys entry

Admins can always recover the GroupKey via the OrgKey path:

```
admin.orgKey ──► unwrapFileKey(group.encrypted_group_key, orgKey) ──► GroupKey
```

This allows admins to provision new members without needing a personal entry in `org_group_keys`.

#### Key rotation

Rotation replaces the GroupKey and re-wraps all group files:

1. Generate a new GroupKey.
2. Re-wrap all group files: `unwrap(oldKey) → wrap(newKey)`.
3. Re-provision all members (RSA path).
4. Atomic `POST /orgs/:orgID/groups/:groupID/key/rotate` call (DB transaction).

During an **OrgKey** rotation, files with `group_id != NULL` are skipped — they are managed by their own GroupKey rotation.

#### Access revocation

Removing a member from a group immediately deletes their `org_group_keys` entry. They can no longer decrypt new files. Files already downloaded (local cache) remain accessible — a GroupKey rotation is recommended after removing a sensitive member.

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
  ├─ 2. current_group_id ?                     │
  │      ├─ YES : getGroupKey(orgID, groupID)  │
  │      │        (RSA path or admin path)     │
  │      │        wrappingKey = groupKey        │
  │      └─ NO  : wrappingKey = orgKey         │
  │                                            │
  ├─ 3. generateMasterKey()   ← FileKey        │
  │                                            │
  ├─ 4. wrapFileKey(fileKey, wrappingKey)      │
  │      → encrypted_key (base64)             │
  │                                            │
  ├─ 5. For each 10 MB chunk:                 │
  │      encryptChunkWorker(chunk, fileKey,    │
  │        index, baseNonce)                   │
  │      → [Nonce 12B][Ciphertext][Tag 16B]    │
  │                                            │
  ├─ 6. POST /fs/multipart/initiate ──────────►│ returns presigned URLs
  │      { encrypted_key, group_id, ... }      │
  │                                            │
  ├─ 7. PUT presigned_url ────────────────────►│ S3 (never plaintext)
  │      (encrypted chunk)                     │
  │                                            │
  └─ 8. POST /fs/multipart/complete ──────────►│ stores encrypted_key + group_id
         { encrypted_key, group_id, parts, ... }│
```

---

### Download flow

```
Client                                    Backend / S3
  │                                            │
  ├─ 1. encrypted_key + group_id ← listing    │
  │      or GET /fs/file/:id/key             │
  │      (returns encrypted_key + group_id)  │
  │                                            │
  ├─ 2. group_id ?                             │
  │      ├─ YES : getGroupKey(orgID, group_id) │
  │      │        wrappingKey = groupKey       │
  │      └─ NO  : getOrgKey(orgID)            │
  │               wrappingKey = orgKey        │
  │                                            │
  ├─ 3. unwrapFileKey(encrypted_key,           │
  │      wrappingKey) → FileKey (AES-256-GCM) │
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
| `OrgFile.encrypted_key` (ungrouped) | AES-GCM wrap, base64 | ✗ — requires the OrgKey |
| `OrgFile.encrypted_key` (grouped) | AES-GCM wrap, base64 | ✗ — requires the GroupKey |
| `OrgGroup.encrypted_group_key` | AES-GCM wrap, base64 | ✗ — requires the OrgKey |
| `OrgGroupKey.encrypted_key` | RSA-OAEP, base64 | ✗ — requires the member's RSA private key |
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

### Group key endpoints — summary

| Endpoint | Role |
|---|---|
| `GET /orgs/:id/groups/:groupID/key` | Returns the encrypted GroupKey (RSA or admin fallback) |
| `POST /orgs/:id/groups/:groupID/key/init` | Initialises the GroupKey (admin, 409 if already set) |
| `PUT /orgs/:id/groups/:groupID/key/members/:userID` | Provisions a single member |
| `GET /orgs/:id/groups/:groupID/key/members` | Lists provisioned `user_id`s (admin) |
| `POST /orgs/:id/groups/:groupID/key/rotate` | Atomic rotation (members + files) |
| `GET /orgs/:id/groups/:groupID/key/files` | Returns all group FileKeys (admin, for rotation) |

### Effective access endpoints

| Endpoint | Role |
|---|---|
| `GET /orgs/:id/my-access` | Caller's effective access across all folders |
| `GET /orgs/:id/members/:userID/effective-access` | Any member's effective access (admin view) |

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
| Group management panel | `frontend/src/components/organizations/OrgGroupsPanel.vue` |
| Backend — members (public key) | `backend/handlers/organizations/members.go` |
| Backend — org (member encrypted key) | `backend/handlers/organizations/list.go` |
| Backend — OrgKey provisioning | `backend/handlers/organizations/members.go` (`SetMemberKey`) |
| Backend — GroupKey (init/rotate/provision) | `backend/handlers/organizations/orggroupkey.go` |
| Backend — effective access | `backend/handlers/organizations/orgeffectiveaccess.go` |
| Backend — access requests | `backend/handlers/organizations/orgaccessrequests.go` |
