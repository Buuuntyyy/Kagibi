# Kagibi

**Stockage cloud chiffré de bout en bout, sans compromis.**

Kagibi est une plateforme de stockage cloud conçue autour d'un principe simple : **ce que vous stockez ne regarde que vous**. Le serveur ne peut pas lire vos fichiers. Pas parce que nous promettons de ne pas le faire — mais parce que nous n'en avons pas la capacité technique.
Ce projet a été développé par [Buuuntyyy] avec l'aide d'intelligence artificielle pour certaines tâches de développement et de documentation. L'objectif est de fournir une solution de stockage sécurisée, respectueuse de la vie privée, et facile à utiliser, tout en étant transparente sur son fonctionnement interne.

---

## Philosophie

La plupart des solutions cloud chiffrent vos données *sur le serveur*, avec des clés que le fournisseur contrôle. En cas de faille, de réquisition judiciaire, ou d'abus interne, vos données sont exposées.

Kagibi fonctionne différemment. Vos fichiers sont chiffrés **sur votre appareil**, avant d'être envoyés. Le serveur ne reçoit que des blocs opaques. Votre clé de déchiffrement ne quitte jamais votre machine.

Ce modèle dit *zero-knowledge* a un coût : il n'est pas possible de récupérer vos fichiers si vous perdez votre mot de passe sans code de récupération. C'est un compromis assumé, pas un bug.

Kagibi est publié sous licence **AGPLv3** : le code est auditable, le déploiement est autonome si vous le souhaitez.

---

## Comment fonctionne le chiffrement

### Dérivation des clés

Quand vous créez un compte ou vous connectez, Kagibi dérive une **clé maître** (MasterKey) à partir de votre mot de passe :

```
Mot de passe + sel aléatoire (16 octets)
        │
        ▼
   Argon2id (64 Mo mémoire, 4 passages)
        │
        ▼
   KEK (Key Encryption Key) — reste en RAM, ne quitte jamais le navigateur
        │
        ▼
   MasterKey — dérivée, stockée uniquement en RAM
```

La **MasterKey** chiffre ensuite tous vos fichiers et métadonnées. La **KEK** enveloppe la MasterKey pour la stocker côté serveur sous forme chiffrée (`EncryptedMasterKey`) — inutilisable sans votre mot de passe.

### Chiffrement des fichiers (upload)

Chaque fichier est découpé en **chunks de 10 Mo**, chiffrés individuellement avec **AES-256-GCM** :

```
Fichier original
        │
        ▼
  Découpage en chunks de 10 Mo
        │
        ▼
  Pour chaque chunk :
    ├── Nonce unique (8 octets aléatoires + 4 octets compteur)
    ├── AES-256-GCM encrypt
    └── Format stocké : [Nonce 12B][Ciphertext][Tag 16B]
        │
        ▼
  Upload direct vers S3 via URLs présignées (TTL 180s)
  Le backend orchestre, mais ne touche jamais le contenu.
```

### Déchiffrement en streaming (téléchargement)

Le téléchargement ne reconstitue jamais le fichier entier en mémoire :

```
URL présignée S3 (TTL 5 min)
        │
        ▼
  ReadableStream (fetch)
        │
        ▼
  TransformStream : parse [Nonce][Ciphertext][Tag] → AES-GCM decrypt
        │
        ▼
  FileSystemWritableFileStream ou Blob
  (jamais stocké déchiffré de façon temporaire)
```

### Ce que le serveur ne peut pas faire

| Opération | Possible pour le serveur ? |
|-----------|---------------------------|
| Lire le contenu d'un fichier | Non — blobs opaques sur S3 |
| Lire le nom d'un fichier | Non si l'option est activée — voir ci-dessous |
| Déchiffrer les données d'un partage | Non — clés chiffrées avec RSA-OAEP |
| Accéder à votre clé maître | Non — jamais transmise au backend |

### Chiffrement des noms de fichiers (opt-in)

Lors de l'inscription, il est possible d'activer le chiffrement des noms de fichiers et dossiers. Cette option est indépendante du chiffrement du contenu (toujours actif).

**Quand l'option est désactivée (défaut)** : les noms sont stockés en clair en base de données et dans le bucket S3. La barre de recherche est fonctionnelle.

**Quand l'option est activée** :

```
Nom du fichier (ex. "rapport.pdf")
        │
        ▼
  AES-256-GCM avec la MasterKey
  IV aléatoire (12 octets, CSPRNG)
        │
        ▼
  Encodage base64url (pas de padding)
  → "aB3xK7mQ..." (opaque, pas de caractère spécial)
        │
  ┌─────┴─────┐
  │           │
  PostgreSQL  S3 OVH
  name = "aB3xK7..."   users/{id}/enc_path/aB3xK7...
```

- Le navigateur déchiffre les noms localement à chaque chargement de répertoire.
- La barre de recherche est désactivée : les noms stockés étant des blobs opaques, une recherche `ILIKE` côté serveur est sans effet.
- Le choix est permanent à la création du compte.

---

## Les trois systèmes de partage

### 1. Partage par lien

Vous générez un lien public que n'importe qui peut ouvrir, sans compte.

**Fonctionnement :**

1. Kagibi génère une `ShareKey` aléatoire, puis chiffre la clé du fichier avec elle.
2. Un token aléatoire (32 octets) est créé et associé au lien.
3. Le lien peut être protégé par un mot de passe (haché en bcrypt) et/ou limité dans le temps (1 à 30 jours).
4. Le destinataire visite le lien, Kagibi lui retourne le blob chiffré et la `ShareKey`.
5. Son navigateur déchiffre le fichier localement.

Le serveur stocke : le token, la clé chiffrée avec la ShareKey, le hash du mot de passe optionnel, la date d'expiration. Il ne peut pas lire le fichier.

---

### 2. Partage avec un ami (utilisateur à utilisateur)

Le partage direct entre comptes utilise la cryptographie asymétrique pour garantir que seul le destinataire peut déchiffrer.

**Fonctionnement :**

1. À la création de compte, chaque utilisateur génère une paire de clés **RSA-OAEP 4096 bits**.
   - La clé publique est stockée en clair sur le serveur.
   - La clé privée est chiffrée avec la MasterKey, puis stockée sur le serveur.

2. Pour ajouter un ami, on utilise son **code ami** (8 caractères alphanumériques, ex. `#A7KD92XZ`), unique par compte.

3. Pour partager un fichier :
   - Kagibi récupère la clé publique RSA du destinataire.
   - La `FileKey` (clé AES du fichier) est chiffrée avec cette clé publique.
   - Le résultat chiffré est stocké en base, rattaché au partage.

4. Quand le destinataire accède au fichier :
   - Il récupère la `FileKey` chiffrée.
   - Son navigateur la déchiffre avec sa clé privée RSA (déchiffrée elle-même avec sa MasterKey).
   - Le fichier est déchiffré localement.

Le serveur stocke : la `FileKey` chiffrée (inutilisable sans la clé privée du destinataire), les relations d'amitié, les permissions.

---

### 3. Transfert P2P (appareil à appareil)

Le transfert P2P envoie des fichiers directement d'un appareil à un autre, sans passer par le stockage serveur.
**Toutefois, les réseaux internet modernes rendent souvent les connexions directes impossibles (NAT, pare-feu). Kagibi utilise un serveur TURN pour relayer les données quand nécessaire, mais le chiffrement de bout en bout est maintenu.**
Concrètement, les appareils établissent une connexion WebRTC DataChannel, et les données sont chiffrées en AES-GCM avant d'être envoyées. Le serveur ne voit que des flux de données chiffrés, même lors du relais TURN.

**Fonctionnement :**

1. Les deux appareils établissent une connexion **WebRTC DataChannel** via un serveur de signalisation (WebSocket).
2. Le backend relaie uniquement les messages de négociation WebRTC (offer/answer/ICE candidates), stockés temporairement dans la table `p2p_signals` pour livraison hors-ligne.
3. Une fois la connexion pair-à-pair établie, les données transitent **directement** entre les appareils, chiffrées en AES-GCM par la couche applicative.
4. Le serveur ne voit ni le contenu transféré, ni les métadonnées du fichier.

Les transferts P2P sont comptabilisés par plan utilisateur (`p2p_max_exchanges`).

---

## Ce qui est stocké sur le serveur

### Données de compte

| Donnée | Format | Pourquoi |
|--------|--------|---------|
| Adresse email | Clair | Authentification |
| Nom d'affichage | Clair | Interface utilisateur |
| Mot de passe | bcrypt (coût 12) | Vérification à la connexion |
| Sel Argon2id | Aléatoire (16 octets) | Dérivation de la KEK côté client |
| `EncryptedMasterKey` | Chiffré (KEK) | Restauration de la MasterKey à la connexion |
| Clé publique RSA | Clair | Chiffrement des partages entrants |
| `EncryptedPrivateKey` | Chiffré (MasterKey) | Déchiffrement des partages reçus |
| Code de récupération | SHA-256 (hash) | Réinitialisation sans email |
| Code ami | Clair | Recherche d'amis |

### Métadonnées de fichiers

| Donnée | Format |
|--------|--------|
| Nom du fichier | Clair (défaut) ou Chiffré AES-GCM si option activée à l'inscription |
| Taille (octets) | Clair |
| Type MIME | Clair |
| Dates de création/modification | Clair |
| Clé de fichier (`EncryptedKey`) | Chiffré (MasterKey) |

### Données sociales et de partage

- Liste d'amis et statut (en attente / accepté)
- Partages actifs : identifiant de ressource + clé chiffrée + permissions
- Liens publics : token + clé chiffrée + expiration + hash de mot de passe optionnel
- Activités récentes (fichiers accédés — optionnel)

### Ce qui n'est pas collecté

- Contenu des fichiers (jamais en clair sur le serveur)
- Historique de navigation ou de recherche
- Adresses IP (sauf journalisation temporaire à des fins de sécurité/abus)
- Données analytiques ou pixels de tracking
- Informations sur l'appareil ou le navigateur

---

## Récupération de compte

Un code de récupération est généré à l'inscription. Il est distinct du mot de passe et permet de retrouver l'accès à la MasterKey si le mot de passe est perdu.

```
Code de récupération (8 caractères)
        │
        ├── SHA-256(code) → stocké comme RecoveryHash (vérification)
        │
        └── Argon2id(code, recovery_salt) → déchiffre EncryptedMasterKeyRecovery
```

Si le code de récupération est également perdu, les données sont **définitivement inaccessibles**. Ce n'est pas un bug — c'est la garantie zero-knowledge.

---

## Suppression des données

- La suppression d'un compte déclenche une **suppression logique** immédiate (marquage `deleted_at`).
- Un processus de nettoyage asynchrone effectue la **suppression physique définitive** au bout de 30 jours : lignes en base, blobs S3.
- Conformité RGPD (articles 17 et 20) : droit à l'effacement et à la portabilité.

---

## Stack technique

| Composant | Technologie |
|-----------|-------------|
| Frontend | Vue 3.5, Vite 7, Pinia |
| Backend | Go 1.21+, Gin |
| Base de données | PostgreSQL 16+ |
| Cache / rate-limit | Redis 7+ |
| Stockage objet | OVH S3 (compatible AWS) |
| Chiffrement | AES-256-GCM, RSA-OAEP 4096, Argon2id |
| Authentification | JWT HS256, TOTP (MFA) |
| P2P | WebRTC DataChannel, TURN/STUN (Coturn) |
| Déploiement | Docker Compose (dev), Kubernetes / Rancher (prod) |

---

## Démarrage rapide (développement)

**Prérequis :** Docker, Docker Compose

```bash
git clone https://github.com/Buuuntyyy/SaferCloud.git
cd SaferCloud

cp backend/.env.example backend/.env
# Renseigner les variables S3, JWT_SECRET, etc.

docker compose up -d
```

Frontend : `http://localhost` — Backend : `http://localhost:8080`

Pour la configuration détaillée (variables d'environnement, S3, Kubernetes), voir [`backend/README.md`](./backend/README.md).

---

## Licence

AGPLv3 — voir [`LICENSE`](./LICENSE).

Toute modification du code, y compris dans un contexte SaaS, doit être publiée sous la même licence.
