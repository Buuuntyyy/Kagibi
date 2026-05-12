# Organisations — Kagibi

---

🇫🇷 [Français](#français) · 🇬🇧 [English](#english)

---

## Français

### Vue d'ensemble

La fonctionnalité **Organisations** permet à plusieurs utilisateurs de partager un espace de stockage chiffré de bout en bout commun. Chaque fichier est chiffré côté client avant envoi ; le serveur ne reçoit que des blobs opaques. L'architecture cryptographique est décrite en détail dans [`org-e2e-encryption.md`](./org-e2e-encryption.md).

---

### Accès à la fonctionnalité

#### Cloud (kagibi.cloud)

Les organisations sont une **fonctionnalité premium**. Elles nécessitent un abonnement Pro ou Business. Les utilisateurs avec un plan gratuit reçoivent une réponse HTTP 402 sur toutes les routes `/orgs`.

La page `/organisations` affiche un écran de mise à niveau pour les utilisateurs non abonnés, avec la liste des fonctionnalités incluses.

#### Instance auto-hébergée

Sur toute instance déployée avec `BILLING_ENABLED=false` dans l'environnement backend, la fonctionnalité est **librement accessible** à tous les utilisateurs, quel que soit leur plan. Aucune vérification d'abonnement n'est effectuée.

---

### Rôles

Chaque membre d'une organisation possède l'un des quatre rôles suivants :

| Rôle | Niveau d'accès par défaut | Description |
|------|--------------------------|-------------|
| `owner` | manage | Propriétaire unique. Peut tout faire, y compris supprimer l'org et faire pivoter la clé. |
| `admin` | manage | Peut gérer les membres (sauf autres admins), les invitations et les permissions. |
| `member` | write | Peut lire et écrire (créer, déplacer) des fichiers. Ne peut pas supprimer sauf sur ses propres dossiers. |
| `viewer` | read | Accès en lecture seule. Peut parcourir et télécharger les fichiers. |

**Règles importantes :**
- Seul le propriétaire peut inviter des admins.
- Seul le propriétaire peut supprimer l'organisation.
- Seul le propriétaire peut pivoter la clé de chiffrement.
- Le rôle `owner` ne peut pas être changé via l'API (pas de transfert de propriété).
- Les admins ne peuvent pas modifier le rôle d'autres admins, ni les retirer.

---

### Permissions par dossier

En plus des rôles globaux, des **overrides de permissions** peuvent être définis par dossier et par utilisateur. Seuls les admins et le propriétaire peuvent configurer ces overrides.

#### Niveaux de permission

| Niveau | Lecture | Création | Déplacement | Suppression |
|--------|---------|----------|-------------|-------------|
| `none` | ✗ | ✗ | ✗ | ✗ |
| `read` | ✓ | ✗ | ✗ | ✗ |
| `write` | ✓ | ✓ | ✓ | ✗ |
| `manage` | ✓ | ✓ | ✓ | ✓ |

#### Champs granulaires

Chaque override expose également des flags fins :

| Champ | Type | Description |
|-------|------|-------------|
| `perm_create` | bool | Peut uploader des fichiers et créer des dossiers |
| `perm_delete` | bool | Peut supprimer des fichiers |
| `perm_download` | bool | Peut télécharger des fichiers (défaut : `true`) |
| `perm_move` | bool | Peut déplacer/renommer des éléments |

#### Résolution hiérarchique

Les permissions sont résolues **du chemin le plus spécifique vers la racine** (`/`) :
1. Si un override existe pour `/projects/secret`, il est appliqué.
2. Sinon, on remonte vers `/projects`.
3. Sinon, on remonte vers `/`.
4. En l'absence d'override, le niveau par défaut du rôle est utilisé.

Les propriétaires et admins ne peuvent pas voir leur accès download restreint.

---

### Invitations

Deux types d'invitation permettent d'ajouter des membres :

#### 1. Invitation par lien (`link invite`)

- Token public non lié à un utilisateur particulier.
- Utilisable par n'importe qui disposant du lien.
- Le nouvel arrivant rejoint **sans clé de chiffrement** (`encrypted_org_key` vide).
- Un admin/owner doit ensuite **provisionner la clé** via l'onglet Membres (bouton « Activer l'accès »).
- Peut être limitée en nombre d'utilisations (`max_uses`) et en durée (`expires_at`).

#### 2. Invitation directe (`direct invite`)

- Liée à un `target_user_id` spécifique.
- L'invitant **chiffre l'OrgKey** avec la clé publique RSA de la cible avant de créer l'invitation.
- Le nouvel arrivant rejoint avec sa clé déjà disponible — aucun provisioning nécessaire.
- L'invitation est révoquée automatiquement à l'acceptation (single-use).

#### Cycle de vie d'une invitation

```
Création (status = "active")
    │
    ├─ Expirée (expires_at dépassé) → rejetée à l'acceptation
    ├─ Épuisée (uses >= max_uses)   → rejetée à l'acceptation
    ├─ Révoquée manuellement        → status = "revoked"
    └─ Acceptée (directe)           → status = "revoked" après acceptation
```

---

### Opérations sur les fichiers

#### Upload multipart

1. `POST /orgs/:id/fs/multipart/initiate` — déclare le fichier, reçoit des presigned URLs S3.
2. Upload des chunks directement vers S3 (jamais via le backend).
3. `POST /orgs/:id/fs/multipart/complete` — valide l'upload, stocke la `encrypted_key` du fichier.
4. `POST /orgs/:id/fs/multipart/abort` — annule l'upload en cours.

Chaque chunk est chiffré côté client (AES-256-GCM) avant envoi. La `encrypted_key` du fichier est un wrap AES-GCM de la FileKey par l'OrgKey.

#### Téléchargement

1. `GET /orgs/:id/fs/file/:fileID/download` — retourne une presigned URL S3.
2. Le client récupère l'OrgKey (depuis le cache ou via déchiffrement RSA), déchiffre la FileKey, puis déchiffre les chunks en streaming.

#### Liste et navigation

`GET /orgs/:id/fs/list/*path` — liste le contenu d'un dossier (paramètre de chemin wildcard).

#### Dossiers

- `POST /orgs/:id/fs/folder` — créer un dossier.
- `DELETE /orgs/:id/fs/folder/:folderID` — supprimer un dossier (et son contenu).

---

### Gestion des clés

#### Initialisation de la clé d'organisation

À la création d'une organisation, le propriétaire génère une **OrgKey** (AES-256-GCM) côté client. Elle est immédiatement chiffrée avec sa clé publique RSA et stockée dans `OrgMember.encrypted_org_key`.

Pour les membres rejoignant via un lien public, la clé doit être provisionnée manuellement :

```
Admin voit le badge rouge sur le membre sans clé
    │
    ├─ 1. Déchiffre l'OrgKey avec sa clé RSA privée (depuis le cache ou authStore)
    ├─ 2. Chiffre l'OrgKey avec la clé publique RSA du membre
    └─ 3. PATCH /orgs/:id/members/:memberID/key { encrypted_org_key: "..." }
```

#### Rotation de clé (`owner` seulement)

La rotation génère une nouvelle OrgKey et re-chiffre toutes les FileKeys en une **transaction atomique** :

```
Owner (frontend)
    │
    ├─ 1. GET /orgs/:id/fs/all-keys         → liste {file_id, encrypted_key}[]
    │
    ├─ 2. Génère une nouvelle OrgKey (AES-256-GCM)
    │
    ├─ 3. Pour chaque membre : chiffre la nouvelle OrgKey avec la clé RSA du membre
    │
    ├─ 4. Pour chaque fichier :
    │      a. Déchiffre l'ancienne FileKey (unwrap avec l'ancienne OrgKey)
    │      b. Re-chiffre la FileKey (wrap avec la nouvelle OrgKey)
    │
    └─ 5. POST /orgs/:id/rotate-key {
              member_keys: [{member_id, encrypted_org_key}...],
              file_keys:   [{file_id, encrypted_key}...]
           }
           → Transaction DB atomique (toutes les clés ou aucune)
```

---

### Journal d'audit

Le journal d'audit est accessible aux admins et propriétaires via `GET /orgs/:id/audit` (50 entrées par page, ordre chronologique inverse). Seules les entrées des **12 derniers mois** sont retournées ; les entrées plus anciennes sont automatiquement masquées.

#### Suppression des logs

Les admins et propriétaires peuvent supprimer des entrées via `DELETE /orgs/:id/audit` avec trois modes :

| Mode | Corps JSON | Effet |
|------|-----------|-------|
| `"all"` | `{ "mode": "all" }` | Supprime toutes les entrées de l'organisation |
| `"months"` | `{ "mode": "months", "months": ["2026-01", "2026-02"] }` | Supprime les entrées des mois YYYY-MM indiqués |
| `"days"` | `{ "mode": "days", "days": ["2026-01-15"] }` | Supprime les entrées des jours YYYY-MM-DD indiqués |

La suppression elle-même est tracée dans le journal (`audit_cleared`).

Le résumé par jour (pour construire l'interface calendrier) est disponible via `GET /orgs/:id/audit/summary` — retourne `{ "days": { "YYYY-MM-DD": count } }` pour la dernière année.

#### Actions enregistrées

| Action | Déclencheur |
|--------|------------|
| `file_uploaded` | Upload multipart complété |
| `file_downloaded` | Téléchargement d'un fichier |
| `file_deleted` | Suppression d'un fichier |
| `member_joined` | Acceptation d'une invitation |
| `member_removed` | Retrait d'un membre (par admin ou auto-retrait) |
| `role_changed` | Modification du rôle d'un membre |
| `key_provisioned` | Provisioning de la clé org pour un membre |
| `key_rotated` | Rotation de la clé d'organisation |
| `invitation_created` | Création d'une invitation |
| `invitation_revoked` | Révocation d'une invitation |
| `permission_set` | Création ou mise à jour d'un override de permission |
| `permission_removed` | Suppression d'un override de permission |
| `audit_cleared` | Suppression d'entrées du journal d'audit |

---

### Modèles de données

#### Organization

| Champ | Type | Description |
|-------|------|-------------|
| `id` | int64 | Identifiant unique |
| `name` | string | Nom de l'organisation |
| `description` | string | Description (optionnel) |
| `owner_id` | string | ID de l'utilisateur propriétaire |
| `storage_quota_mb` | int64 | Quota de stockage en Mo (défaut : 10 240 Mo) |
| `created_at` | timestamp | Date de création |
| `updated_at` | timestamp | Date de dernière modification |

#### OrgMember

| Champ | Type | Description |
|-------|------|-------------|
| `id` | int64 | Identifiant |
| `org_id` | int64 | Référence à l'organisation |
| `user_id` | string | Référence à l'utilisateur |
| `role` | string | `owner` / `admin` / `member` / `viewer` |
| `encrypted_org_key` | string | OrgKey chiffrée RSA-OAEP (base64) |
| `joined_at` | timestamp | Date d'adhésion |

#### OrgInvitation

| Champ | Type | Description |
|-------|------|-------------|
| `id` | int64 | Identifiant |
| `org_id` | int64 | Organisation concernée |
| `invited_by` | string | ID de l'invitant |
| `token` | string | Token d'invitation (32 char hex) |
| `target_user_id` | string? | Invitation directe (nil = lien) |
| `encrypted_org_key` | string | OrgKey pré-chiffrée pour les invitations directes |
| `role` | string | Rôle attribué à l'acceptation |
| `max_uses` | int | Nombre max d'utilisations (0 = illimité) |
| `uses` | int | Nombre d'utilisations actuelles |
| `expires_at` | timestamp? | Date d'expiration |
| `status` | string | `active` / `revoked` |

#### OrgFolderPermission

| Champ | Type | Description |
|-------|------|-------------|
| `org_id` | int64 | Organisation |
| `user_id` | string | Membre concerné |
| `folder_path` | string | Chemin normalisé (ex : `/projets/secret`) |
| `level` | string | `none` / `read` / `write` / `manage` |
| `perm_create` | bool | Peut créer |
| `perm_delete` | bool | Peut supprimer |
| `perm_download` | bool | Peut télécharger |
| `perm_move` | bool | Peut déplacer |

#### OrgAuditLog

| Champ | Type | Description |
|-------|------|-------------|
| `id` | int64 | Identifiant |
| `org_id` | int64 | Organisation |
| `actor_id` | string | Utilisateur ayant déclenché l'action |
| `action` | string | Type d'action (voir tableau ci-dessus) |
| `target_id` | string | ID de la ressource affectée |
| `target_type` | string | `user` / `file` / `invitation` / `permission` / `org` |
| `detail` | string | Détail lisible (ex : `member → admin`) |
| `created_at` | timestamp | Horodatage |

---

### Référence API

Toutes les routes (sauf `GET /org-invitations/:token`) nécessitent un JWT valide dans le header `Authorization: Bearer <token>`.

Sur le cloud, les routes marquées ★ renvoient HTTP 402 pour les utilisateurs avec un plan gratuit.

#### Organisations

| Méthode | Route | Auth | Description |
|---------|-------|------|-------------|
| `POST` | `/api/v1/orgs` | JWT ★ | Créer une organisation |
| `GET` | `/api/v1/orgs` | JWT ★ | Lister ses organisations |
| `GET` | `/api/v1/orgs/:orgID` | JWT | Détail d'une organisation |
| `PATCH` | `/api/v1/orgs/:orgID` | JWT (admin+) | Modifier nom / description / quota |
| `DELETE` | `/api/v1/orgs/:orgID` | JWT (owner) | Supprimer l'organisation |

#### Membres

| Méthode | Route | Auth | Description |
|---------|-------|------|-------------|
| `GET` | `/api/v1/orgs/:orgID/members` | JWT (member+) | Lister les membres |
| `PATCH` | `/api/v1/orgs/:orgID/members/:memberID` | JWT (admin+) | Changer le rôle d'un membre |
| `DELETE` | `/api/v1/orgs/:orgID/members/:memberID` | JWT | Retirer un membre (self ou admin+) |
| `PATCH` | `/api/v1/orgs/:orgID/members/:memberID/key` | JWT (admin+) | Provisionner la clé org d'un membre |

#### Invitations

| Méthode | Route | Auth | Description |
|---------|-------|------|-------------|
| `POST` | `/api/v1/orgs/:orgID/invitations` | JWT (admin+) | Créer une invitation |
| `GET` | `/api/v1/orgs/:orgID/invitations` | JWT (admin+) | Lister les invitations actives |
| `DELETE` | `/api/v1/orgs/:orgID/invitations/:invID` | JWT (admin+) | Révoquer une invitation |
| `GET` | `/api/v1/org-invitations/:token` | Public | Infos sur une invitation (aperçu) |
| `POST` | `/api/v1/org-invitations/:token/accept` | JWT ★ | Accepter une invitation |

#### Système de fichiers

| Méthode | Route | Auth | Description |
|---------|-------|------|-------------|
| `GET` | `/api/v1/orgs/:orgID/fs/list/*path` | JWT (member+) | Lister un dossier |
| `POST` | `/api/v1/orgs/:orgID/fs/folder` | JWT (write+) | Créer un dossier |
| `DELETE` | `/api/v1/orgs/:orgID/fs/folder/:folderID` | JWT (manage) | Supprimer un dossier |
| `GET` | `/api/v1/orgs/:orgID/fs/file/:fileID/download` | JWT (member+) | Télécharger un fichier |
| `GET` | `/api/v1/orgs/:orgID/fs/file/:fileID/key` | JWT (member+) | Récupérer la clé d'un fichier |
| `DELETE` | `/api/v1/orgs/:orgID/fs/file/:fileID` | JWT (manage) | Supprimer un fichier |
| `POST` | `/api/v1/orgs/:orgID/fs/multipart/initiate` | JWT (write+) | Initier un upload multipart |
| `POST` | `/api/v1/orgs/:orgID/fs/multipart/complete` | JWT (write+) | Finaliser un upload multipart |
| `POST` | `/api/v1/orgs/:orgID/fs/multipart/abort` | JWT (write+) | Annuler un upload multipart |

#### Permissions

| Méthode | Route | Auth | Description |
|---------|-------|------|-------------|
| `GET` | `/api/v1/orgs/:orgID/permissions` | JWT (admin+) | Lister tous les overrides |
| `PUT` | `/api/v1/orgs/:orgID/permissions` | JWT (admin+) | Créer / mettre à jour un override |
| `DELETE` | `/api/v1/orgs/:orgID/permissions` | JWT (admin+) | Supprimer un override |
| `GET` | `/api/v1/orgs/:orgID/permissions/me` | JWT (member+) | Permission effective du caller |

#### Administration

| Méthode | Route | Auth | Description |
|---------|-------|------|-------------|
| `GET` | `/api/v1/orgs/:orgID/audit` | JWT (admin+) | Journal d'audit (50/page, dernière année) |
| `GET` | `/api/v1/orgs/:orgID/audit/summary` | JWT (admin+) | Comptage par jour pour la dernière année |
| `DELETE` | `/api/v1/orgs/:orgID/audit` | JWT (admin+) | Supprimer des entrées (all / months / days) |
| `GET` | `/api/v1/orgs/:orgID/fs/all-keys` | JWT (admin+) | Toutes les FileKeys (pour rotation) |
| `POST` | `/api/v1/orgs/:orgID/rotate-key` | JWT (owner) | Pivoter la clé d'organisation |

---

### Implémentation — où trouver le code

| Composant | Fichier |
|-----------|---------|
| Handler principal + résolution des permissions | `backend/handlers/organizations/handler.go` |
| Création d'organisation | `backend/handlers/organizations/create.go` |
| Liste / détail | `backend/handlers/organizations/list.go` |
| Mise à jour / suppression | `backend/handlers/organizations/update.go` |
| Gestion des membres | `backend/handlers/organizations/members.go` |
| Invitations | `backend/handlers/organizations/invitations.go` |
| Système de fichiers (liste, download, delete) | `backend/handlers/organizations/orgfiles.go` |
| Upload multipart | `backend/handlers/organizations/orgmultipart.go` |
| Navigation (fs/list) | `backend/handlers/organizations/orgfs.go` |
| Permissions par dossier | `backend/handlers/organizations/permissions.go` |
| Journal d'audit + all-keys | `backend/handlers/organizations/orgaudit.go` |
| Rotation de clé | `backend/handlers/organizations/orgrotatekey.go` |
| Primitives crypto org | `frontend/src/utils/orgCrypto.js` |
| Store Pinia (état + actions) | `frontend/src/stores/organizations.js` |
| Page liste des organisations | `frontend/src/views/OrganizationsView.vue` |
| Page détail d'une organisation | `frontend/src/views/OrgDetailView.vue` |
| Page acceptation d'invitation | `frontend/src/views/JoinView.vue` |
| Modèle de chiffrement détaillé | `docs/org-e2e-encryption.md` |

---

---

## English

### Overview

The **Organisations** feature lets multiple users share a common end-to-end encrypted storage space. Every file is encrypted client-side before being sent; the server only receives opaque blobs. The cryptographic architecture is described in detail in [`org-e2e-encryption.md`](./org-e2e-encryption.md).

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
| `joined_at` | timestamp | Join date |

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
| Org crypto primitives | `frontend/src/utils/orgCrypto.js` |
| Pinia store (state + actions) | `frontend/src/stores/organizations.js` |
| Organisations list page | `frontend/src/views/OrganizationsView.vue` |
| Organisation detail page | `frontend/src/views/OrgDetailView.vue` |
| Invitation acceptance page | `frontend/src/views/JoinView.vue` |
| Detailed encryption model | `docs/org-e2e-encryption.md` |
