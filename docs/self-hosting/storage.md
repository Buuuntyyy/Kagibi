# Stockage

L'auto-hébergement de Kagibi repose sur un socle Docker Compose commun (`docker-compose.yaml` : PostgreSQL, Redis, backend, frontend), complété par un choix de **backend de stockage objet**, activé via un second fichier Compose optionnel :

| | Fichiers à lancer | Stockage | Emplacement des données |
|---|---|---|---|
| **Option A — S3 externe** | `docker-compose.yaml` seul | OVH Object Storage, AWS S3, Scaleway, Cloudflare R2, MinIO managé... | Chez le fournisseur S3 |
| **Option B — Local (Garage)** | `docker-compose.yaml` + `docker-compose.garage.yml` | [Garage](https://garagehq.deuxfleurs.fr/), S3-compatible, auto-hébergé | Sur votre propre disque/NAS |

Dans les deux cas, le backend Kagibi ne change pas : il parle toujours le protocole S3 (`backend/pkg/s3storage/client.go`). Seule la destination change.

**Quel choix faire ?** L'option A convient si vous voulez déléguer la durabilité des données à un fournisseur (réplication multi-AZ, SLA) sans gérer de disque vous-même. L'option B convient si vous voulez que les fichiers (déjà chiffrés de bout en bout côté client) ne quittent jamais votre propre matériel — typiquement un NAS avec du RAID/ZFS pour la durabilité.

## Option A — Stockage S3 externe (OVH Object Storage)

### 1. Créer un conteneur (bucket) S3 chez OVH

Dans le manager OVH : **Public Cloud → Object Storage → Créer un objet de stockage**, type **Standard** (ou High Performance selon vos besoins), puis choisissez une région (ex. `gra` — Gravelines, `sbg` — Strasbourg, `de` — Francfort, `waw` — Varsovie). Notez la région et le nom du conteneur.

### 2. Générer des identifiants S3

Dans **Object Storage → Gestion des utilisateurs** (ou "S3 Users"), créez un utilisateur avec des droits de lecture/écriture sur le conteneur créé. OVH fournit alors une **Access Key** et une **Secret Key** — c'est la seule fois où la Secret Key est affichée en clair, copiez-la immédiatement.

### 3. Configurer le CORS du bucket (obligatoire)

Le navigateur téléverse et télécharge les fichiers **directement** vers le bucket S3, via des URLs présignées générées par le backend — sans cette étape, ces requêtes cross-origin échouent silencieusement (erreurs CORS visibles uniquement dans la console du navigateur).

```bash
cp scripts/s3-cors-config.example.json scripts/s3-cors-config.json
# Éditer scripts/s3-cors-config.json : remplacer AllowedOrigins par votre(vos) domaine(s)
#   ex: ["https://votre-domaine.example"]

aws s3api put-bucket-cors \
  --endpoint-url https://s3.<region>.io.cloud.ovh.net \
  --bucket <nom-du-conteneur> \
  --cors-configuration file://scripts/s3-cors-config.json
```
(nécessite l'AWS CLI, configuré avec les identifiants de l'étape 2 — `aws configure` ou variables `AWS_ACCESS_KEY_ID`/`AWS_SECRET_ACCESS_KEY` d'environnement pour cette seule commande).

### 4. Renseigner `.env`

Rouvrez le fichier `.env` créé à l'étape des prérequis communs (celui qui contient déjà `JWT_SECRET`, `DB_PASSWORD`, etc.) et complétez les 5 variables S3 restées vides :

| Variable | Exemple OVH |
|---|---|
| `S3_ENDPOINT` | `https://s3.gra.io.cloud.ovh.net` |
| `S3_BUCKET` | `<nom-du-conteneur>` |
| `S3_REGION` | `gra` |
| `S3_ACCESS_KEY` | `<Access Key de l'étape 2>` |
| `S3_SECRET_KEY` | `<Secret Key de l'étape 2>` |

### 5. Lancer

```bash
docker compose up -d --build
```

### 6. Vérifier

```bash
curl http://localhost:8080/health          # {"status":"ok"}
docker compose logs -f backend             # pas d'erreur "S3 configuration is missing"
```
Ensuite, dans l'interface : créez un compte, envoyez un fichier puis téléchargez-le — cette séquence confirme que le chemin présigné (backend → navigateur → S3 → navigateur) fonctionne de bout en bout.

### Maintenance

- **Mise à jour** : `git pull && docker compose up -d --build`.
- **Sauvegarde** : la durabilité des fichiers est déléguée au fournisseur (réplication multi-AZ chez OVH). Seule la base PostgreSQL (`db_data/`) contient les métadonnées (arborescence, clés de fichier chiffrées, comptes) — à sauvegarder séparément (`pg_dump` régulier ou snapshot du volume).

### Dépannage

| Symptôme | Cause probable |
|---|---|
| `S3 configuration is missing` dans les logs backend | Une des 5 variables `S3_*` est vide dans `.env` |
| Upload bloqué, erreur CORS dans la console navigateur | Étape 3 non faite, ou `AllowedOrigins` du CORS ne correspond pas exactement à l'origine servie (schéma + domaine + port) |
| `403 Forbidden` sur les requêtes S3 | Access/Secret Key invalides, ou utilisateur S3 sans droit d'écriture sur le conteneur |

---

## Option B — Stockage 100% local (Garage)

### 0. Décider de l'adresse joignable par le navigateur

Contrairement au backend (qui peut communiquer avec Garage via le réseau Docker interne), les URLs présignées données au **navigateur** doivent pointer vers une adresse qu'il peut réellement résoudre :

- **Usage strictement local** (vous accédez à Kagibi via `http://localhost` sur la même machine que Docker, aucun reverse-proxy) : `http://localhost:3900` — le port est déjà publié sur l'hôte par `docker-compose.garage.yml` (`"${GARAGE_S3_PORT:-3900}:3900"`), inutile d'aller plus loin dans ce cas.
- **Accès depuis d'autres appareils sur le réseau local** : l'IP LAN de votre machine/NAS, ex. `http://192.168.1.50:3900`.
- **Accès depuis Internet** : un domaine derrière un reverse-proxy TLS, ex. `https://s3.votre-domaine.example` (voir la page [Nginx](./nginx)).

Décidez-le maintenant, cette adresse sert à l'étape 3.

### 1. Configurer Garage

```bash
cp garage/garage.toml.example garage/garage.toml
```
Éditez `garage/garage.toml` et remplacez les deux placeholders par des secrets générés :
```bash
openssl rand -hex 32   # → rpc_secret
openssl rand -hex 32   # → admin_token
```
`garage.toml` est ignoré par git (`.gitignore`) — ne jamais le committer avec de vrais secrets.

### 2. Créer le fichier de credentials (placeholder)

```bash
mkdir -p garage/credentials && touch garage/credentials/s3.env
```
Requis même vide : Docker Compose valide l'existence de tous les `env_file:` référencés avant de démarrer quoi que ce soit — sans ce fichier, `docker compose up` refuse de démarrer. Il sera rempli automatiquement au premier lancement par `garage-init`.

### 3. Compléter `.env`

Rouvrez le même fichier `.env` créé à l'étape des prérequis communs (celui qui contient déjà `JWT_SECRET`, `DB_PASSWORD`, etc.) et ajoutez-y :

```bash
# Obligatoire — voir étape 0. Exemple ci-dessous pour un accès réseau local (LAN) ;
# remplacez par http://localhost:3900 pour un usage strictement local sans reverse-proxy.
GARAGE_PUBLIC_ENDPOINT=http://192.168.1.50:3900

# Optionnel — valeurs par défaut déjà cohérentes entre garage.toml, bootstrap.sh et le backend
# GARAGE_S3_PORT=3900
# GARAGE_S3_REGION=garage
# GARAGE_BUCKET=kagibi
# GARAGE_KEY_NAME=kagibi-key
# GARAGE_CAPACITY=10G
```
Les 5 variables `S3_*` de la section précédente (`S3_ENDPOINT`, `S3_BUCKET`...) n'ont pas besoin d'être renseignées ici, même si elles sont encore vides dans `.env` : `docker-compose.garage.yml` les remplace automatiquement pour le service `backend` — inutile d'y toucher pour cette option.

### 4. Lancer

```bash
docker compose -f docker-compose.yaml -f docker-compose.garage.yml up -d --build
```
Au premier lancement, `garage-init` attend que le nœud Garage réponde, assigne le layout (obligatoire même à un seul nœud), crée le bucket et une clé S3, puis écrit `garage/credentials/s3.env` — le backend ne démarre qu'une fois cette étape terminée avec succès (`depends_on: condition: service_completed_successfully`).

### 5. Vérifier le bootstrap

```bash
docker compose -f docker-compose.yaml -f docker-compose.garage.yml ps garage-init
# → doit afficher "Exited (0)"

cat garage/credentials/s3.env
# → doit contenir S3_ACCESS_KEY=... et S3_SECRET_KEY=... (non vides)
```
Si `garage-init` a échoué (code de sortie ≠ 0) : `docker compose logs garage-init` — le script s'arrête au premier échec (`set -e`) avec un message explicite.

### 6. Configurer le CORS du bucket (obligatoire, comme pour l'option A)

```bash
cp scripts/s3-cors-config.example.json scripts/s3-cors-config.json
# Éditer AllowedOrigins avec votre domaine/IP frontend

aws s3api put-bucket-cors \
  --endpoint-url "$GARAGE_PUBLIC_ENDPOINT" \
  --bucket kagibi \
  --cors-configuration file://scripts/s3-cors-config.json
```
(AWS CLI configuré avec les `S3_ACCESS_KEY`/`S3_SECRET_KEY` lues dans `garage/credentials/s3.env` à l'étape 5).

### 7. Vérifier de bout en bout

```bash
curl http://localhost:8080/health
```
Ensuite, dans l'interface : créez un compte, envoyez un fichier, puis téléchargez-le. Si l'envoi reste bloqué, consultez le dépannage ci-dessous — c'est le symptôme typique d'un `GARAGE_PUBLIC_ENDPOINT` mal renseigné.

### Maintenance

- **Mise à jour** : `git pull && docker compose -f docker-compose.yaml -f docker-compose.garage.yml up -d --build`. `garage-init` se termine immédiatement au redémarrage (`garage/credentials/s3.env` déjà rempli), rien n'est réappliqué.
- **Sauvegarde** : sauvegarder `garage/data/` (les objets), `garage/meta/` (les métadonnées Garage) et `garage/garage.toml` (contient `rpc_secret`/`admin_token` — sans eux, un nœud restauré ne peut pas rejoindre son propre layout) — un snapshot ZFS régulier du dataset qui héberge ces dossiers couvre les trois. `db_data/` (PostgreSQL) reste à sauvegarder séparément, comme pour l'option A.

### Dépannage

| Symptôme | Cause probable |
|---|---|
| `docker compose up` refuse de démarrer, `env file ... s3.env not found` | Étape 2 non faite |
| `GARAGE_PUBLIC_ENDPOINT manquant dans .env` au lancement | Variable absente de `.env` (étape 3) |
| `garage-init` en erreur sur `layout apply` | Un layout a déjà été appliqué avec une version différente (ex. `garage/credentials/s3.env` supprimé manuellement après un premier bootstrap réussi) — vérifier `docker compose exec garage garage layout show` et ajuster manuellement si besoin |
| Upload/téléchargement bloqué depuis le navigateur, mais `curl` sur `/health` fonctionne | `GARAGE_PUBLIC_ENDPOINT` injoignable depuis le poste client (mauvaise IP, port non ouvert sur le pare-feu/routeur) ou CORS non configuré (étape 6) |
| `garage-init` bloqué indéfiniment sur "Attente que le nœud Garage réponde" | `garage.toml` absent ou mal rempli (étape 1) — vérifier `docker compose logs garage` |
