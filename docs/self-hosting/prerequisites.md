# Prérequis

Ces prérequis sont communs aux deux options de stockage — S3 externe ou 100 % local (Garage), voir la page [Stockage](./storage).

## Ce qu'il vous faut

- **Docker Engine** avec le plugin **Compose v2** (`docker compose version` doit fonctionner — pas l'ancien binaire Python `docker-compose` v1, obsolète).

## Étapes

1. Clonez le dépôt :
   ```bash
   git clone https://github.com/Buuuntyyy/Kagibi.git
   cd Kagibi
   ```

2. Créez le fichier d'environnement racine :
   ```bash
   cp .env.example .env
   ```

3. Générez les quatre secrets obligatoires et collez-les dans `.env` :

   | Variable | Commande | Rôle |
   |---|---|---|
   | `JWT_SECRET` | `openssl rand -hex 32` | Signe les jetons de session |
   | `EMAIL_ENCRYPTION_KEY` | `openssl rand -hex 32` | Chiffre les emails au repos — **ne jamais changer après la première utilisation** |
   | `DB_PASSWORD` | `openssl rand -hex 24` | Mot de passe PostgreSQL |
   | `REDIS_PASSWORD` | `openssl rand -hex 24` | Mot de passe Redis (requirepass) |

4. Adaptez les URLs à votre déploiement dans `.env` :
   - `VITE_API_URL` — URL par laquelle le **navigateur** atteint le backend (`http://localhost:8080/api/v1` en local, `https://votre-domaine.example/api/v1` en production).
   - `ALLOWED_ORIGINS` — origine du frontend, doit correspondre exactement (`http://localhost` en local, `https://votre-domaine.example` en production). Jamais `*` : incompatible avec l'envoi de cookies/identifiants.

`BILLING_ENABLED=false` est déjà la valeur par défaut de `.env.example` — rien à faire, stockage illimité et aucune UI de facturation, ce qui est recommandé en self-hosted.

## Prochaine étape

Une fois ces étapes terminées, rendez-vous sur la page [Stockage](./storage) pour choisir et configurer votre backend.
