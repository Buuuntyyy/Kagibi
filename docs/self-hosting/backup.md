# Sauvegarde

Une instance Kagibi protège deux types de données bien distincts, avec deux stratégies différentes :

| Donnée | Où elle vit | Criticité | Approche |
|---|---|---|---|
| Métadonnées + clés chiffrées | PostgreSQL | **Critique** — perte = compte inutilisable même si les fichiers survivent | Dump régulier vers un bucket S3 externe |
| Contenu des fichiers | S3 externe ou Garage (self-hosted) | Important mais volumineux | Réplication côté fournisseur (S3 externe) ou synchronisation programmée (Garage) |

Point important : les fichiers sont **déjà chiffrés côté client** avant d'être envoyés (chiffrement de bout en bout). Une sauvegarde n'a donc besoin d'aucune gestion de clé — copier les objets tels quels suffit, ils restent illisibles sans les clés qui, elles, vivent dans la base de données. **C'est justement pour ça que la sauvegarde de la base est la priorité absolue** : sans elle, une sauvegarde des fichiers seule ne sert à rien (des blobs chiffrés sans leurs clés).

## 1. Base de données (priorité)

Le principe : dump régulier, compressé, poussé vers un bucket S3 — idéalement **différent** de celui qui stocke déjà vos fichiers (principe 3-2-1 : une sauvegarde sur le même support que l'original ne protège de rien en cas de perte de ce support).

### Script

```bash
#!/bin/bash
set -euo pipefail
cd /opt/kagibi   # adapter au chemin réel de votre installation

set -a; source .env; set +a

DUMP_FILE="kagibi-db-$(date +%Y%m%d-%H%M%S).sql.gz"
TMP_PATH="/tmp/$DUMP_FILE"

docker compose exec -T db pg_dump -U "${DB_USER:-user}" "${DB_NAME:-mydb}" | gzip > "$TMP_PATH"

# Un dump vide ou tronqué n'est pas une sauvegarde utilisable — ne rien pousser dans ce cas.
if [ ! -s "$TMP_PATH" ] || ! gzip -t "$TMP_PATH" 2>/dev/null; then
  echo "❌ Dump invalide ou vide — sauvegarde annulée." >&2
  rm -f "$TMP_PATH"
  exit 1
fi

rclone copy "$TMP_PATH" backup:kagibi-backups/db/
rm -f "$TMP_PATH"

echo "✅ Sauvegarde DB envoyée : $DUMP_FILE"
```

[rclone](https://rclone.org/) est utilisé ici plutôt qu'un outil propre à un fournisseur : il parle le protocole S3 (et beaucoup d'autres), donc le même script fonctionne quel que soit le bucket de destination — OVH, AWS, Scaleway, Cloudflare R2, Backblaze B2, un autre Garage...

### Configuration de rclone

```bash
rclone config
# Type de stockage : S3 (Amazon S3 Compliant Storage Providers)
# Provider : le vôtre (ou "Other" pour un stockage S3-compatible générique)
# access_key_id / secret_access_key : identifiants du bucket de DESTINATION (pas ceux de Kagibi)
# endpoint : URL S3 du fournisseur de destination
```
Ceci crée un remote nommé `backup` (ou le nom choisi) dans `~/.config/rclone/rclone.conf`, réutilisé par le script ci-dessus (`backup:kagibi-backups/db/`).

### Planification (cron)

```bash
# crontab -e
0 3 * * * /opt/kagibi/scripts/backup-db.sh >> /var/log/kagibi-backup.log 2>&1
```

### Rétention

Plutôt que de scripter la suppression des anciens dumps (risque réel de bug qui efface tout d'un coup), configurez une **règle de cycle de vie (lifecycle rule)** sur le bucket de destination lui-même, côté fournisseur — "supprimer les objets sous `kagibi-backups/db/` après 30 jours" par exemple. C'est une fonctionnalité standard de tout stockage S3-compatible, gérée en dehors de tout script qui pourrait avoir un bug.

## 2. Fichiers

### Option A — S3 externe (OVH, AWS, Scaleway, R2...)

Vos fichiers sont déjà dans un bucket géré par ce fournisseur. La bonne approche est la **réplication inter-bucket native** proposée par ces fournisseurs (ex. Cross-Region Replication chez AWS, équivalents chez OVH/Scaleway) plutôt qu'un script Kagibi : c'est plus fiable, ça ne consomme pas de bande passante/CPU sur votre instance, et c'est nativement cohérent avec le stockage source. Consultez la documentation "réplication" ou "cross-region replication" de votre fournisseur — la configuration se fait entièrement de leur côté, aucun changement à faire dans Kagibi.

### Option B — Garage (stockage local self-hosted)

**Ne sauvegardez pas en copiant directement les dossiers `garage/data`/`garage/meta` pendant que Garage tourne** — c'est un magasin de données actif, une copie de fichiers à chaud pendant qu'il écrit peut produire une sauvegarde incohérente. La bonne méthode est de synchroniser via l'**API S3 que Garage expose lui-même**, qui garantit de ne lire que des objets complets :

```bash
rclone config
# Créer un remote "garage_local" :
#   Type : S3, Provider : Other
#   access_key_id / secret_access_key : ceux générés dans garage/credentials/s3.env
#   endpoint : http://localhost:3900 (ou l'IP/domaine configuré selon votre setup)
```

```bash
rclone sync garage_local:kagibi backup:kagibi-backups/files --transfers 8 --progress
```

À planifier de la même façon que le dump DB (cron), à une fréquence adaptée au volume de données (une synchronisation quotidienne suffit dans la plupart des cas — `rclone sync` ne retransfère que ce qui a changé).

## 3. Tester la restauration

Une sauvegarde jamais restaurée n'est pas une sauvegarde vérifiée. Périodiquement (trimestriellement, par exemple) :

```bash
# Récupérer un dump depuis le bucket de sauvegarde
rclone copy backup:kagibi-backups/db/kagibi-db-<date>.sql.gz .

# Restaurer sur une base de test (JAMAIS sur la prod pour ce test)
gunzip -c kagibi-db-<date>.sql.gz | docker compose exec -T db psql -U "${DB_USER:-user}" "${DB_NAME:-mydb}_test_restore"
```

Voir aussi la page [Mise à jour](./upgrade) pour la procédure de sauvegarde à effectuer spécifiquement avant une mise à niveau de version (légèrement différente : ponctuelle et locale, plutôt que récurrente vers un bucket distant).
