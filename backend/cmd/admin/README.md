# admin — CLI opérateur Kagibi / Kagibi Operator CLI

---

🇫🇷 [Français](#français) · 🇬🇧 [English](#english)

---

## Français

Outil en ligne de commande à usage interne pour provisionner les organisations clientes.  
Il se connecte directement à la base de données et **n'expose aucun port réseau**.  
À exécuter sur le serveur via SSH uniquement.

### Compilation

```bash
cd backend/
go build -o admin ./cmd/admin
```

Le binaire résultant (`admin`) est autonome et ne dépend que des variables d'environnement du serveur.

### Variables d'environnement requises

| Variable            | Description                                             |
|---------------------|---------------------------------------------------------|
| `DATABASE_URL`      | DSN PostgreSQL (ex : `postgresql://user:pass@host/db`) |
| `APP_URL`           | URL publique du frontend (ex : `https://kagibi.cloud`) |
| `MAIL_HOST`         | Serveur SMTP                                            |
| `MAIL_USERNAME`     | Login SMTP                                              |
| `MAIL_PASSWORD`     | Mot de passe SMTP                                       |
| `MAIL_FROM_ADDRESS` | Adresse expéditeur (ex : `no-reply@kagibi.cloud`)      |

`DATABASE_URL` est obligatoire. Les variables `MAIL_*` sont optionnelles : si absentes, le lien d'invitation est affiché dans le terminal sans qu'un email soit envoyé.

En développement local, un fichier `.env` à la racine de `backend/` est automatiquement chargé.

### Commandes

#### `org create` — Provisionner une organisation cliente

```bash
./admin org create \
  --name        "Acme Corp" \
  --desc        "Client enterprise — contrat 2025" \
  --quota       51200 \
  --owner-email cto@acme.com
```

| Flag            | Obligatoire | Description                            | Défaut          |
|-----------------|-------------|----------------------------------------|-----------------|
| `--name`        | Oui         | Nom de l'organisation                  | —               |
| `--desc`        | Non         | Description interne                    | `""`            |
| `--quota`       | Non         | Quota de stockage en Mo                | `10240` (10 Go) |
| `--owner-email` | Non         | Email du propriétaire — reçoit le lien | `""`            |

**Ce que fait la commande :**

1. Crée l'organisation en base avec `owner_id = "pending"`.
2. Génère une invitation owner à usage unique, valable **7 jours**.
3. Envoie un email au client avec le lien (si `--owner-email` et `MAIL_*` sont configurés).
4. Affiche le lien dans le terminal dans tous les cas.

**Exemple de sortie :**

```
  ✓ Organisation créée
  ID:              42
  Nom:             Acme Corp
  Quota:           51200 Mo

  Lien d'invitation owner (valable 7 jours, 1 usage) :
  https://kagibi.cloud/join/a3f8c1d2e4b5...

  Un email a été envoyé à cto@acme.com
```

**Côté client :** le lien redirige vers la page d'inscription/connexion. Une fois le compte créé et le lien accepté, l'utilisateur devient automatiquement propriétaire de l'organisation.

> **Lien expiré ?** Générez-en un nouveau via l'API (`POST /api/v1/orgs/:id/invitations`) depuis le compte d'un admin de l'org, ou relancez `org create` pour une nouvelle org.

---

#### `org list` — Lister toutes les organisations

```bash
./admin org list
```

```
ID  NOM          MEMBRES  UTILISÉ (Mo)  QUOTA (Mo)  CRÉÉ LE
1   Acme Corp    3        1240          51200        2025-06-01
2   Beta SAS     1        0             10240        2025-06-15
```

---

#### `org quota` — Modifier le quota de stockage

Utile lors d'un changement de plan ou d'une extension contractuelle.

```bash
./admin org quota --id 42 --quota 102400
```

| Flag      | Obligatoire | Description          |
|-----------|-------------|----------------------|
| `--id`    | Oui         | ID de l'organisation |
| `--quota` | Oui         | Nouveau quota en Mo  |

---

#### `org delete` — Supprimer une organisation

```bash
./admin org delete --id 42
```

Demande une confirmation interactive (`oui` pour valider). Passer `--yes` pour bypasser (scripts).

```bash
./admin org delete --id 42 --yes
```

**Ce que fait la commande :** révoque toutes les invitations actives, supprime tous les membres, puis supprime l'organisation (soft-delete).

> Les fichiers stockés sur S3 ne sont **pas** supprimés automatiquement. Nettoyez-les manuellement via la console S3 ou AWS CLI si nécessaire.

---

### Architecture — Réutilisation avec Stripe

Toute la logique métier est dans `internal/provisioning/org.go`, découplée du CLI et de HTTP.  
Lors de l'intégration Stripe, le webhook appellera les mêmes fonctions :

```
Aujourd'hui  →  CLI ssh          →  provisioning.CreateOrg()
Demain       →  Stripe webhook   →  provisioning.CreateOrg()
Si besoin    →  Panel admin web  →  provisioning.CreateOrg()
```

---

## English

An internal command-line tool for provisioning customer organisations.  
It connects directly to the database and **exposes no network surface**.  
Run it on the server over SSH only.

### Build

```bash
cd backend/
go build -o admin ./cmd/admin
```

The resulting binary (`admin`) is self-contained and only relies on server environment variables.

### Required environment variables

| Variable            | Description                                               |
|---------------------|-----------------------------------------------------------|
| `DATABASE_URL`      | PostgreSQL DSN (e.g. `postgresql://user:pass@host/db`)   |
| `APP_URL`           | Public frontend URL (e.g. `https://kagibi.cloud`)        |
| `MAIL_HOST`         | SMTP server                                               |
| `MAIL_USERNAME`     | SMTP login                                                |
| `MAIL_PASSWORD`     | SMTP password                                             |
| `MAIL_FROM_ADDRESS` | Sender address (e.g. `no-reply@kagibi.cloud`)            |

`DATABASE_URL` is required. The `MAIL_*` variables are optional: if absent, the invitation link is printed to the terminal without sending an email.

In local development, a `.env` file at the root of `backend/` is loaded automatically.

### Commands

#### `org create` — Provision a customer organisation

```bash
./admin org create \
  --name        "Acme Corp" \
  --desc        "Enterprise client — 2025 contract" \
  --quota       51200 \
  --owner-email cto@acme.com
```

| Flag            | Required | Description                              | Default         |
|-----------------|----------|------------------------------------------|-----------------|
| `--name`        | Yes      | Organisation name                        | —               |
| `--desc`        | No       | Internal description                     | `""`            |
| `--quota`       | No       | Storage quota in MB                      | `10240` (10 GB) |
| `--owner-email` | No       | Owner email — receives the invite link   | `""`            |

**What it does:**

1. Creates the organisation in the database with `owner_id = "pending"`.
2. Generates a single-use owner invitation valid for **7 days**.
3. Sends an email to the client with the link (if `--owner-email` and `MAIL_*` are set).
4. Prints the link to the terminal regardless.

**Sample output:**

```
  ✓ Organisation created
  ID:              42
  Name:            Acme Corp
  Quota:           51200 MB

  Owner invitation link (valid 7 days, 1 use):
  https://kagibi.cloud/join/a3f8c1d2e4b5...

  An email has been sent to cto@acme.com
```

**Client side:** the link redirects to the sign-up/login page. Once the account is created and the link is accepted, the user automatically becomes the organisation owner.

> **Link expired?** Generate a new one via the API (`POST /api/v1/orgs/:id/invitations`) from an org admin account, or re-run `org create` for a new org.

---

#### `org list` — List all organisations

```bash
./admin org list
```

```
ID  NAME         MEMBERS  USED (MB)  QUOTA (MB)  CREATED
1   Acme Corp    3        1240       51200        2025-06-01
2   Beta SAS     1        0          10240        2025-06-15
```

---

#### `org quota` — Update storage quota

Useful when a client upgrades their plan or requests an extension.

```bash
./admin org quota --id 42 --quota 102400
```

| Flag      | Required | Description          |
|-----------|----------|----------------------|
| `--id`    | Yes      | Organisation ID      |
| `--quota` | Yes      | New quota in MB      |

---

#### `org delete` — Delete an organisation

```bash
./admin org delete --id 42
```

Prompts for interactive confirmation (`oui` to confirm). Pass `--yes` to skip (scripts).

```bash
./admin org delete --id 42 --yes
```

**What it does:** revokes all active invitations, removes all members, then soft-deletes the organisation.

> Files stored on S3 are **not** deleted automatically. Clean them up manually via the S3 console or AWS CLI if needed.

---

### Architecture — Future Stripe integration

All business logic lives in `internal/provisioning/org.go`, decoupled from both the CLI and HTTP.  
When Stripe is integrated, the webhook will call the exact same functions:

```
Today         →  SSH CLI          →  provisioning.CreateOrg()
Tomorrow      →  Stripe webhook   →  provisioning.CreateOrg()
If needed     →  Admin web panel  →  provisioning.CreateOrg()
```

No code duplication required.
