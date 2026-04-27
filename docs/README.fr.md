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

## Fonctionnalités

### Gestion des fichiers

- **Upload de fichiers** — glisser-déposer ou sélection classique, avec progression en temps réel (phase de chiffrement puis phase d'envoi).
- **Upload de dossiers** — téléversez une arborescence entière en une fois. En cas de conflit de nom, trois options s'offrent à vous : renommer automatiquement, ignorer ou remplacer.
- **Upload multipart** — les fichiers volumineux (> 10 Mo) sont découpés en fragments de 10 Mo, chiffrés individuellement et envoyés en parallèle.
- **Téléchargement** — déchiffrement en streaming côté client : le fichier n'est jamais reconstruit en clair en mémoire avant d'être écrit sur disque.
- **Organisation** — création de dossiers, renommage, déplacement, suppression (simple ou récursive).
- **Tags** — étiquetez vos dossiers pour les retrouver plus facilement via la recherche et les filtres.
- **Prévisualisation** — aperçu des images et PDF directement dans le navigateur, sans téléchargement.

### Recherche et filtrage

La barre de recherche globale (raccourci **Ctrl+K**) parcourt l'ensemble de vos fichiers et dossiers.

- **Résultats en contexte** — cliquer sur un résultat vous amène directement à l'emplacement du fichier dans l'arborescence, avec mise en surbrillance.
- **Filtres disponibles** :
  - Par catégorie de type (Tous, Documents, Images, Archives)
  - Par extension (ex. `.pdf`, `.mp4`)
  - Par tag (étiquettes posées sur les dossiers)
  - Par type d'élément (fichier ou dossier)
- **Note** : la recherche est désactivée si le chiffrement des noms de fichiers est activé, les noms stockés étant opaques pour le serveur.

### Partage

Trois mécanismes de partage coexistent, décrits en détail dans la section [Les trois systèmes de partage](#les-trois-systèmes-de-partage).

- **Partage par lien** — lien public (avec ou sans compte), possibilité de déposer des fichiers dans un dossier partagé publiquement.
- **Partage avec un ami** — permissions granulaires (téléchargement, création, suppression, déplacement), gestion visuelle en vert/rouge. Les fichiers déposés par l'ami sont récupérables par le propriétaire via une chaîne de clés dossier.
- **Transfert P2P** — aucun stockage serveur, chiffrement de bout en bout.

### Transfert P2P

Envoi direct d'un fichier d'un appareil à un autre, chiffré de bout en bout, sans stockage intermédiaire sur nos serveurs. Voir la section dédiée pour le détail du fonctionnement.

### Amis et présence

- Système de **code ami** (8 caractères alphanumériques, ex. `#A7KD92XZ`) pour trouver d'autres utilisateurs sans exposer l'adresse e-mail.
- Envoi et acceptation de demandes d'amitié.
- **Indicateur de présence** en temps réel (point vert) avec tolérance de 8 secondes à la déconnexion pour éviter les clignotements.
- Suppression mutuelle d'un ami (révoque automatiquement les partages associés).

### Sécurité du compte

- **Authentification à deux facteurs (MFA)** — TOTP (application d'authentification), avec verrouillage de 15 minutes après 5 tentatives échouées.
- **Code de récupération** — généré à l'inscription, permet de regagner l'accès à la clé maître si le mot de passe est perdu.
- **Révocation de sessions** — déconnexion de tous les appareils instantanément.
- **Élévation AAL2** — certaines actions sensibles (changement de mot de passe, suppression du compte) nécessitent une confirmation MFA même si la session est déjà active.

### Conformité RGPD

- **Droit à l'effacement (Art. 17)** — la suppression de compte déclenche une suppression logique immédiate, suivie d'une suppression physique définitive (fichiers S3 + lignes base de données) au bout de 30 jours.
- **Droit à la portabilité (Art. 20)** — export de toutes vos données sur demande.

### Interface et ergonomie

- Thème **clair / sombre**, bascule en un clic.
- Interface **multilingue** : français et anglais, avec persistance du choix.
- **Navigation au clavier** : Ctrl+K pour la recherche, touches fléchées dans les listes.
- **Design responsive** : navigation adaptée mobile avec barre inférieure et feuilles de bas de page.
- **Quota de stockage** affiché en temps réel dans la barre latérale (mis à jour en moins de 2 secondes après chaque opération).

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

Lorsque le lien porte sur un **dossier**, la page publique permet également de **déposer des fichiers** dans ce dossier. Les fichiers envoyés par des visiteurs sont chiffrés dans leur navigateur avec la `FolderKey`, puis téléversés vers le stockage S3 du propriétaire. Le serveur n'a à aucun moment accès au contenu en clair.

Le serveur stocke : le token, la clé chiffrée avec la ShareKey, le hash du mot de passe optionnel, la date d'expiration. Il ne peut pas lire le fichier.

---

### 2. Partage avec un ami (utilisateur à utilisateur)

Le partage direct entre comptes utilise la cryptographie asymétrique pour garantir que seul le destinataire peut déchiffrer.

**Fonctionnement :**

1. À la création de compte, chaque utilisateur génère une paire de clés **RSA-OAEP 4096 bits**.
   - La clé publique est stockée en clair sur le serveur.
   - La clé privée est chiffrée avec la MasterKey, puis stockée sur le serveur.

2. Pour ajouter un ami, on utilise son **code ami** (8 caractères alphanumériques, ex. `#A7KD92XZ`), unique par compte.

3. Pour partager un **fichier** :
   - Kagibi récupère la clé publique RSA du destinataire.
   - La `FileKey` (clé AES du fichier) est chiffrée avec cette clé publique.
   - Le résultat chiffré est stocké en base, rattaché au partage.

4. Quand le destinataire accède au fichier :
   - Il récupère la `FileKey` chiffrée.
   - Son navigateur la déchiffre avec sa clé privée RSA (déchiffrée elle-même avec sa MasterKey).
   - Le fichier est déchiffré localement.

5. Pour partager un **dossier** (avec permissions granulaires) :
   - Le propriétaire génère une `FolderKey` (AES-256), chiffrée avec sa propre MasterKey et stockée côté serveur.
   - Il définit les permissions accordées à l'ami.
   - L'ami accède au contenu du dossier selon les droits accordés.

#### Permissions de partage de dossier

| Permission | Accorde |
|------------|---------|
| Téléchargement | Accéder et télécharger les fichiers |
| Création | Déposer des fichiers et créer des sous-dossiers |
| Suppression | Supprimer des fichiers dans le dossier partagé |
| Déplacement | Renommer et déplacer des éléments |

Permissions accordées par défaut lors d'un nouveau partage : **Téléchargement + Création**.

Les permissions sont visualisées en couleur dans la boîte de dialogue de gestion du partage : **vert** = droit accordé, **rouge** = droit refusé. Toute tentative d'action sans le droit correspondant déclenche un message d'erreur explicite.

#### Chaîne de clés pour les fichiers déposés par un ami

Quand un ami dépose un fichier dans votre dossier partagé, le fichier est chiffré avec une clé dérivée de la `FolderKey`. Pour que le propriétaire puisse le télécharger, le backend expose un endpoint de récupération de clé :

```
MasterKey du propriétaire
        │
        ▼
  Déchiffre folder.encrypted_key  →  FolderKey
        │
        ▼
  Déchiffre folder_file_key.encrypted_key  →  FileKey
        │
        ▼
  Déchiffrement du contenu du fichier
```

Cette chaîne garantit que le propriétaire retrouve toujours accès à ses fichiers, même ceux déposés par des tiers, sans jamais exposer la MasterKey au serveur.

Le serveur stocke : la `FileKey` chiffrée (inutilisable sans la clé privée du destinataire), les relations d'amitié, les permissions.

---

### 3. Transfert P2P (appareil à appareil)

Le transfert P2P envoie des fichiers directement d'un appareil à un autre, chiffré de bout en bout, sans stockage serveur intermédiaire.

Les réseaux modernes rendent parfois les connexions directes impossibles (NAT, pare-feu). Dans ce cas, Kagibi utilise un serveur TURN (relais) qui appartient à Kagibi. **Les données transitant par ce relais restent chiffrées en AES-GCM — le serveur ne voit que des flux opaques, pas le contenu.**

#### Deux modes de transfert

**Mode direct (entre amis enregistrés)**

1. L'expéditeur sélectionne un ami en ligne et un fichier, puis lance le transfert.
2. Une clé de fichier AES-256 est générée aléatoirement, chiffrée avec la clé publique RSA du destinataire.
3. La connexion WebRTC est négociée via WebSocket (signaux stockés dans `p2p_signals`).
4. Une fois le canal DataChannel ouvert, le fichier est envoyé par fragments de 16 Ko, chacun chiffré avec un nonce aléatoire distinct.
5. Le destinataire reçoit une notification sonore + visuelle, accepte le transfert, et son navigateur déchiffre et reconstitue le fichier localement.

**Mode invitation (sans compte requis)**

1. L'expéditeur génère un **lien d'invitation** depuis la page P2P.
2. Le lien peut être partagé manuellement ou envoyé par e-mail (en français ou en anglais au choix).
3. Le destinataire ouvre le lien sur `send.kagibi.cloud` — **aucun compte n'est nécessaire**.
4. Il génère une paire de clés RSA éphémère dans son navigateur (non stockée).
5. L'expéditeur est notifié de l'acceptation et démarre le transfert WebRTC.
6. Le lien d'invitation est à **usage unique** et expire après 24 heures.

#### Informations affichées pendant le transfert

- **Progression** en pourcentage avec barre visuelle.
- **Vitesse de transfert** (ex. `4.2 MB/s`) calculée en temps réel.
- **Temps restant estimé** (ex. `~1m 30s`).
- **Type de connexion** : direct (LAN), via STUN (traversée NAT) ou via relais TURN.
- **Re-notification** : l'expéditeur peut relancer une alerte sonore au destinataire (jusqu'à 3 fois, cooldown 30 s).

---

## Ce qui est stocké sur le serveur

### Données de compte

| Donnée | Format | Pourquoi |
|--------|--------|---------|
| Adresse e-mail | Chiffré (AES-256-GCM) | Authentification, sans exposition en clair |
| Nom d'affichage | Clair | Interface utilisateur |
| Mot de passe | bcrypt (coût 12) | Vérification à la connexion |
| Sel Argon2id | Aléatoire (16 octets) | Dérivation de la KEK côté client |
| `EncryptedMasterKey` | Chiffré (KEK) | Restauration de la MasterKey à la connexion |
| Clé publique RSA | Clair | Chiffrement des partages entrants |
| `EncryptedPrivateKey` | Chiffré (MasterKey) | Déchiffrement des partages reçus |
| Code de récupération | SHA-256 (hash) | Réinitialisation sans e-mail |
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
- Invitations P2P : token + nom du fichier + taille + date d'expiration (contenu non stocké)

### Ce qui n'est pas collecté

- Contenu des fichiers (jamais en clair sur le serveur)
- Historique de navigation ou de recherche
- Adresses IP (sauf journalisation temporaire à des fins de sécurité/abus)
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
git clone https://github.com/Buuuntyyy/Kagibi.git
cd Kagibi
cp backend/.env.example backend/.env   # Configurer les variables S3, JWT_SECRET, etc.
cp frontend/.env.example frontend/.env # Configurer VITE_BACKEND_URL=http://localhost:8080

cd backend
go run main.go

cd frontend
npm install
npm run dev
```

Frontend : `http://localhost` — Backend : `http://localhost:8080`

Pour la configuration détaillée (variables d'environnement, S3, Kubernetes), voir [`backend/README.md`](../backend/README.md).

---

## Licence

AGPLv3 — voir [`LICENSE`](../LICENSE).

Toute modification du code, y compris dans un contexte SaaS, doit être publiée sous la même licence.
