# Mise à jour

Kagibi publie des images Docker versionnées et signées (`vX.Y.Z`) à chaque release — voir le [CHANGELOG](https://github.com/Buuuntyyy/Kagibi/blob/main/CHANGELOG.md) ou les [GitHub Releases](https://github.com/Buuuntyyy/Kagibi/releases) pour savoir ce qui change avant de mettre à jour.

Les migrations de base de données tournent automatiquement au démarrage du backend et sont **cumulatives et idempotentes** : passer directement d'une ancienne version à une version récente (sans s'arrêter aux versions intermédiaires) est sûr, sauf mention contraire explicite dans les notes de release d'une version donnée.

## 1. Sauvegarder avant de toucher à quoi que ce soit

Une mise à jour qui tourne mal doit rester réversible. Sauvegardez systématiquement :

```bash
# Charge les variables du .env (DB_USER, DB_NAME, DB_PASSWORD...)
set -a; source .env; set +a

mkdir -p backups
BACKUP_FILE="backups/kagibi-$(date +%Y%m%d-%H%M%S).sql.gz"

docker compose exec -T db pg_dump -U "${DB_USER:-user}" "${DB_NAME:-mydb}" | gzip > "$BACKUP_FILE"

# Vérification minimale : un dump vide ou tronqué n'est pas une sauvegarde utilisable.
if [ ! -s "$BACKUP_FILE" ] || ! gzip -t "$BACKUP_FILE" 2>/dev/null; then
  echo "❌ Sauvegarde invalide ou vide — NE PAS continuer la mise à jour." >&2
  exit 1
fi
echo "✅ Sauvegarde OK : $BACKUP_FILE ($(du -h "$BACKUP_FILE" | cut -f1))"
```

Si vous êtes en **stockage local (Option B, Garage)**, sauvegardez aussi les fichiers eux-mêmes — la base ne contient que les métadonnées et clés chiffrées, pas les données :

```bash
tar czf "backups/garage-$(date +%Y%m%d-%H%M%S).tar.gz" garage/data garage/meta
```

Avec un **S3 externe** (OVH, AWS, Scaleway, R2...), la redondance/sauvegarde des objets relève du fournisseur — vérifiez sa politique si ce n'est pas déjà fait.

## 2. Épingler la nouvelle version

Dans `.env` :

```bash
KAGIBI_VERSION=v2.32.0   # remplacez par la version cible
```

Ne mettez jamais `KAGIBI_VERSION=latest` en production durable : sans version épinglée, un simple `docker compose pull` peut basculer vers une version que vous n'avez pas choisi de tester à ce moment précis.

## 3. (Recommandé) Vérifier la signature des images

Les images sont signées avec [cosign](https://docs.sigstore.dev/) (mode *keyless*, via GitHub Actions) — vérifiable sans clé à récupérer nulle part :

```bash
cosign verify \
  --certificate-identity-regexp "^https://github\.com/Buuuntyyy/Kagibi/\.github/workflows/release\.yml@refs/heads/main$" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  ghcr.io/buuuntyyy/kagibi-backend:v2.32.0
```

(remplacez le tag pour la commande `kagibi-frontend` équivalente). Une vérification réussie confirme que l'image provient bien du pipeline officiel de ce dépôt et n'a pas été altérée entre la publication et votre `pull`.

## 4. Appliquer la mise à jour

```bash
docker compose pull
docker compose up -d
```

Les migrations tournent automatiquement au démarrage du backend.

## 5. Vérifier que tout s'est bien passé

```bash
# Le backend a bien démarré et répond
curl -f http://localhost:${BACKEND_PORT:-8080}/health

# Les migrations se sont terminées sans erreur
docker compose logs backend | grep "\[migrate\] Schema OK"
```

La seconde commande doit afficher une ligne du type :

```
[migrate] Schema OK — 28 étapes appliquées (version applicative v2.32.0)
```

Le numéro de version dans cette ligne doit correspondre à `KAGIBI_VERSION`. Si la ligne n'apparaît pas du tout, la migration a probablement échoué en cours de route — consultez les lignes précédentes (`Warning: ...` ou une erreur fatale) dans les mêmes logs.

## 6. En cas de problème — revenir en arrière

```bash
# 1. Revenir à l'ancienne version dans .env
KAGIBI_VERSION=v2.31.0   # la version précédente, qui fonctionnait

# 2. Redémarrer sur cette version
docker compose up -d

# 3. Si la base de données a été altérée par la tentative de mise à jour, restaurer
#    le dump pris à l'étape 1 (à faire seulement si nécessaire — les migrations étant
#    additives, revenir à l'ancienne image suffit dans l'immense majorité des cas) :
gunzip -c backups/kagibi-<horodatage>.sql.gz | docker compose exec -T db psql -U "${DB_USER:-user}" "${DB_NAME:-mydb}"
```

## Ce qui rend cette procédure sûre

- **Images signées et versionnées** : `KAGIBI_VERSION` épinglé garantit que vous savez exactement ce qui tourne, et que revenir en arrière veut dire pointer sur un tag précis, pas re-builder un ancien commit.
- **Migrations idempotentes et tracées** : chaque étape de migration est enregistrée dans la table `schema_migrations` avec la version applicative — la ligne de log à l'étape 5 en est la preuve directement consultable.
- **Aucune suppression destructrice au fil des versions** : les colonnes/tables remplacées sont ajoutées en parallèle des anciennes plutôt que de les écraser immédiatement, ce qui garde un retour en arrière réaliste.
