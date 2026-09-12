# Organisations — Kagibi

### Vue d'ensemble

La fonctionnalité **Organisations** permet à plusieurs utilisateurs de partager un espace de stockage chiffré de bout en bout commun. Chaque fichier est chiffré côté client avant envoi ; le serveur ne reçoit que des blobs opaques. L'architecture cryptographique est décrite en détail dans [`org-e2e-encryption.md`](./org-e2e-encryption).

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

### Synchronisation LDAP / Active Directory

Chaque organisation peut se connecter à un annuaire **LDAP** ou **Active Directory** pour importer automatiquement les membres et les groupes. La configuration est réservée aux rôles `admin` et `owner`.

#### Flux de synchronisation

1. Le scheduler lance un cycle complet à l'intervalle configuré (défaut : 60 min).
2. Une connexion LDAP est ouverte et le compte de service effectue un `bind`.
3. Une recherche d'utilisateurs est effectuée sous `user_base_dn` avec `user_filter`.
4. **Garde-fou 1** : si le nombre de résultats < `min_expected_users`, la sync est annulée.
5. **Garde-fou 2** : si > 20 % des membres LDAP existants disparaissent d'un coup, la sync est annulée.
6. Pour chaque utilisateur LDAP :
   - S'il a déjà un compte Kagibi et est membre → `source` et `ldap_uid` mis à jour.
   - S'il a un compte mais n'est pas encore membre → une invitation directe lui est envoyée.
   - S'il n'a pas de compte → une invitation par lien lui est envoyée par e-mail.
7. Les membres avec `source = 'ldap'` absents du résultat LDAP sont **suspendus** (`suspended_at`).
8. Les membres suspendus depuis plus de `auto_deprovision_days` jours sont **retirés** de l'org.
9. Si `group_base_dn` est défini, les groupes LDAP sont créés ou mis à jour comme groupes Kagibi (`source = 'ldap'`).

#### Contrainte réseau : connexion directe sortante

**Le modèle actuel exige que le backend Kagibi puisse joindre directement le serveur LDAP/AD du client.** `client.Dial()` (`backend/internal/ldap/client.go`) ouvre une connexion TCP sortante vers l'URL configurée — il n'existe aucun agent ni tunnel intermédiaire côté client.

Implications pour un déploiement réel (client dont l'annuaire est sur son réseau interne, hors du réseau où tourne Kagibi) :

- **Le serveur LDAP/AD du client doit être exposé publiquement** sur le port configuré (389/636), au moins vers l'IP de sortie de Kagibi — typiquement via un NAT/port-forward sur le pare-feu du client.
- **`ldaps://` (TLS implicite) est fortement recommandé** pour tout trafic hors réseau local : le bind DN et le mot de passe transitent en clair sur `ldap://` simple. Le fallback `StartTLS` du client (`client.go`) échoue silencieusement côté serveur (log uniquement) si la cible ne le supporte pas — rien n'empêche côté Kagibi une connexion effectivement non chiffrée.
- **Restriction par IP côté client** : pour ne pas exposer l'annuaire à tout Internet, le client doit idéalement restreindre l'accès à l'IP de sortie de Kagibi. Cela suppose une IP de sortie **stable et documentée** côté Kagibi (passerelle NAT dédiée) — non garanti par défaut selon l'infrastructure d'hébergement du backend.
- **Friction attendue côté client** : la majorité des équipes sécurité/IT en entreprise refusent d'exposer un annuaire interne (AD en particulier) sur Internet, même restreint par IP — c'est un vecteur d'attaque activement scanné. Ce modèle convient à des structures plus petites ou déjà ouvertes sur Internet (LDAP déjà hébergé dans un cloud, VPN existant), mais constitue un blocage dur pour les clients avec des exigences de sécurité strictes.

L'alternative standard du marché (Okta, Azure AD Connect, JumpCloud…) est un **agent installé dans le réseau du client**, qui interroge l'annuaire en local et ouvre lui-même une connexion **sortante** vers le SaaS (HTTPS/WebSocket) — aucun port entrant à ouvrir côté client. Kagibi ne dispose pas de ce mode aujourd'hui.

#### Chiffrement du mot de passe Bind

Le mot de passe du compte de service est chiffré **AES-256-GCM** via `emailcrypto.Encrypt` (clé dérivée de `EMAIL_ENCRYPTION_KEY`) avant stockage en base. Il n'est jamais transmis en clair.

#### Déprovisionnement

| Phase | Déclencheur | Effet |
|-------|-------------|-------|
| Suspension | Absent du LDAP au cycle suivant | `suspended_at` enregistré, accès org retiré |
| Suppression | `suspended_at` + `auto_deprovision_days` dépassé | Permissions, groupes et adhésion supprimés |

Si `auto_deprovision_days = 0`, la suppression reste manuelle.

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
| `source` | string | `internal` (manuel) / `ldap` (sync annuaire) |
| `ldap_uid` | string | UID LDAP de l'utilisateur (vide si `source = internal`) |
| `suspended_at` | timestamp? | Date de suspension par LDAP (nil si actif) |
| `joined_at` | timestamp | Date d'adhésion |

#### OrgLDAPConfig

Configuration LDAP d'une organisation (une ligne par org).

| Champ | Type | Description |
|-------|------|-------------|
| `org_id` | int64 | Référence à l'organisation (unique) |
| `enabled` | bool | Sync activée ou non |
| `url` | string | URL du serveur (`ldap://` ou `ldaps://`) |
| `bind_dn` | string | DN du compte de service |
| `bind_password_enc` | string | Mot de passe chiffré AES-256-GCM |
| `user_base_dn` | string | Racine de la recherche utilisateurs |
| `user_filter` | string | Filtre LDAP (défaut : `(objectClass=person)`) |
| `group_base_dn` | string | Racine des groupes (vide = désactivé) |
| `group_filter` | string | Filtre groupes (défaut : `(objectClass=groupOfNames)`) |
| `attr_email` | string | Attribut e-mail (défaut : `mail`) |
| `attr_display_name` | string | Attribut nom affiché (défaut : `cn`) |
| `attr_uid` | string | Attribut UID unique (défaut : `uid`) |
| `tls_skip_verify` | bool | Ignorer la vérification TLS |
| `sync_interval_minutes` | int | Intervalle de sync en minutes (min. 5) |
| `auto_deprovision_days` | int | Délai de grâce avant suppression (0 = manuel) |
| `min_expected_users` | int | Seuil de sécurité anti-vidage |
| `last_sync_at` | timestamp? | Horodatage de la dernière sync |
| `last_sync_error` | string | Message d'erreur de la dernière sync (vide si OK) |
| `last_sync_stats` | jsonb | Statistiques de la dernière sync (`LDAPSyncStats`) |

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

#### LDAP / Active Directory

| Méthode | Route | Auth | Description |
|---------|-------|------|-------------|
| `GET` | `/api/v1/orgs/:orgID/ldap` | JWT (admin+) | Lire la configuration LDAP |
| `PUT` | `/api/v1/orgs/:orgID/ldap` | JWT (admin+) | Créer ou mettre à jour la configuration |
| `POST` | `/api/v1/orgs/:orgID/ldap/test` | JWT (admin+) | Tester la connexion et le bind |
| `POST` | `/api/v1/orgs/:orgID/ldap/sync` | JWT (admin+) | Déclencher une sync immédiate |
| `GET` | `/api/v1/orgs/:orgID/ldap/suspended` | JWT (admin+) | Lister les membres suspendus par la sync |

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
| **LDAP : handlers HTTP** | `backend/handlers/organizations/ldap.go` |
| **LDAP : connexion et recherche** | `backend/internal/ldap/client.go` |
| **LDAP : logique de sync** | `backend/internal/ldap/sync.go` |
| **LDAP : scheduler par org** | `backend/internal/ldap/scheduler.go` |
| Primitives crypto org | `frontend/src/utils/orgCrypto.js` |
| Store Pinia (état + actions) | `frontend/src/stores/organizations.js` |
| Page liste des organisations | `frontend/src/views/OrganizationsView.vue` |
| Page détail d'une organisation | `frontend/src/views/OrgDetailView.vue` |
| **Panneau LDAP (admin UI)** | `frontend/src/components/organizations/OrgLDAPPanel.vue` |
| Page acceptation d'invitation | `frontend/src/views/JoinView.vue` |
| Modèle de chiffrement détaillé | [`org-e2e-encryption.md`](./org-e2e-encryption) |
