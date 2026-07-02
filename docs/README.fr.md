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
- **Upload multipart** — les fichiers volumineux sont découpés en fragments de 5 à 100 Mo, chiffrés individuellement avec AES-256-GCM dans le navigateur, puis envoyés **en parallèle** directement vers S3 via des URLs présignées (TTL 3 min). Le backend orchestre les URLs présignées et finalise l'opération multipart, mais ne touche jamais le contenu brut.
- **Téléchargement** — déchiffrement en streaming côté client : le fichier n'est jamais reconstruit en clair en mémoire avant d'être écrit sur disque.
- **Organisation** — création de dossiers, renommage, déplacement, suppression (simple ou récursive).
- **Tags** — étiquetez vos dossiers pour les retrouver plus facilement via la recherche et les filtres.
- **Prévisualisation** — aperçu des images et PDF directement dans le navigateur, sans téléchargement.
- **Favoris** — marquez n'importe quel fichier ou dossier comme favori via le bouton ★ dans le tableau ou le menu contextuel. Les favoris apparaissent dans une section dédiée sur la page d'accueil et peuvent servir de critère de tri dans le navigateur de fichiers.

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
- **Page FAQ** (`/faq`) — accessible publiquement, sans compte, depuis la landing page et le menu **Aide & Support** de la navbar du dashboard. Couvre les questions générales (souveraineté, sécurité, chiffrement), les fonctionnalités (P2P, partages, Organisations, Amis) et les valeurs de Kagibi.

### Page d'accueil

La page d'accueil (`/dashboard/home`) présente trois sections accordéon :

1. **Fichiers favoris** — affiche tous les fichiers et dossiers marqués en favori, en grille 5 colonnes. La section n'apparaît que si au moins un favori existe. Cliquer sur une carte ouvre directement l'élément ; l'étoile dorée visible au survol retire l'élément des favoris.
2. **Récemment ouverts** — les 10 derniers fichiers ou dossiers consultés, avec navigation rapide en un clic.
3. **Partagés avec moi** — fichiers et dossiers partagés par d'autres utilisateurs.

---

## Organisations

Les organisations sont des espaces collaboratifs chiffrés de bout en bout. Tous les fichiers, noms de dossiers et métadonnées stockés dans une organisation sont chiffrés avec une clé partagée que seuls les membres détiennent — le serveur n'a aucun accès au contenu.

### Rôles et accès

Chaque membre d'une organisation possède l'un des quatre rôles suivants :

| Rôle | Droits |
|------|--------|
| Propriétaire (owner) | Contrôle total : gestion des membres, rôles, quota, suppression de l'org |
| Admin | Gestion des membres, provisionnement, audit, invitations |
| Membre (member) | Lecture et écriture des fichiers selon les permissions de dossier |
| Lecteur (viewer) | Accès en lecture seule aux dossiers autorisés |

### Chiffrement de bout en bout

Chaque organisation utilise une **OrgKey** dédiée (AES-256). Cette clé est générée une fois par le propriétaire, puis re-chiffrée individuellement pour chaque membre avec sa clé publique RSA-4096 avant d'être stockée côté serveur. Concrètement :

- Le serveur ne détient jamais l'OrgKey en clair.
- L'ajout d'un nouveau membre nécessite qu'un admin **provisionne** sa clé : l'admin déchiffre l'OrgKey localement, puis la re-chiffre avec la clé publique du nouveau membre.
- Les membres ayant rejoint via un lien d'invitation ne peuvent pas déchiffrer le contenu de l'organisation tant qu'un admin n'a pas provisionné leur clé.

```
MasterKey du propriétaire
       │
       ▼
  OrgKey (AES-256)
       │
  ┌────┴─────────────┐
  │                  │
  RSA-OAEP(membre1)  RSA-OAEP(membre2) ...
  stocké côté serveur par membre
```

### Groupes

Au sein d'une organisation, les **groupes** permettent de regrouper des membres pour leur assigner des permissions de façon collective :

- Créer, renommer et supprimer des groupes.
- Ajouter ou retirer des membres d'un groupe.
- Les membres d'un groupe héritent d'un rôle au sein du groupe : **admin** ou **membre**.
- Les permissions de dossier peuvent être assignées à un groupe entier en une seule opération.

#### Chiffrement par groupe (GroupKey)

Chaque groupe peut disposer d'une **clé de chiffrement dédiée** (GroupKey, AES-256-GCM), indépendante de l'OrgKey. Les fichiers stockés dans un dossier lié au groupe sont chiffrés avec cette GroupKey — seuls les membres du groupe (ayant reçu leur clé provisionnée) peuvent les déchiffrer, même si d'autres membres de l'organisation ont accès au reste du contenu.

- **Initialisation** : un admin génère la GroupKey côté client, la wrappe avec l'OrgKey (backup admin) et la chiffre avec la clé publique RSA de chaque membre du groupe.
- **Provisionnement d'un nouveau membre** : après l'ajout d'un membre au groupe, un admin peut lui provisionner la GroupKey en un clic depuis l'onglet Groupes.
- **Rotation** : la rotation remplace la GroupKey, re-wrappe tous les fichiers du groupe avec la nouvelle clé, et reprovisionnne tous les membres — opération atomique côté serveur.
- **Révocation** : retirer un membre du groupe supprime immédiatement son entrée de clé ; une rotation est recommandée pour garantir la rupture cryptographique complète.
- **Rétrocompatibilité** : les fichiers sans GroupKey (`group_id = NULL`) continuent d'utiliser l'OrgKey sans migration nécessaire.

### Synchronisation LDAP / Active Directory

Les organisations peuvent se connecter à un annuaire d'entreprise **LDAP ou Active Directory** pour synchroniser automatiquement leurs membres et groupes, sans avoir à gérer manuellement les invitations :

- **Provisionnement automatique** — les nouveaux utilisateurs du LDAP reçoivent une invitation par e-mail et rejoignent l'organisation dès qu'ils l'acceptent.
- **Synchronisation des groupes** — les groupes LDAP sont recréés comme groupes Kagibi et leurs membres mis à jour à chaque cycle.
- **Déprovisionnement en deux phases** — un utilisateur qui quitte l'annuaire est d'abord suspendu, puis retiré automatiquement après un délai de grâce configurable (ou manuellement par un admin).
- **Garde-fous** — la sync est annulée si le LDAP retourne un résultat vide ou si plus de 20 % des membres existants disparaissent d'un coup, protégeant contre les erreurs de filtre et les pannes réseau.
- **Chiffrement du mot de passe Bind** — le mot de passe du compte de service est chiffré AES-256-GCM avant stockage.

La configuration s'effectue dans l'onglet **Administration → LDAP / AD** de l'organisation (réservé aux admins et owners). Voir la [documentation complète LDAP](../desktop-app/DOCUMENTATION.md#10-annuaire-ldap--active-directory) pour tous les détails de configuration et de fonctionnement.

### Assistant d'initialisation (Onboarding Wizard)

Quand un utilisateur crée sa première organisation, un assistant pas-à-pas le guide :

1. Nommage de l'organisation et saisie d'une description.
2. Explication du modèle de chiffrement et du flux de provisionnement de clés.
3. Création du premier lien d'invitation pour les membres de l'équipe.

### Gestion des fichiers dans les organisations

Le navigateur de fichiers d'une organisation propose une interface complète pour le travail collaboratif :

- **Tri** — par nom, taille ou date (ascendant / descendant).
- **Filtrage** — par catégorie de type (images, documents, vidéos, audio, archives) ou par tag d'organisation.
- **Navigation par fil d'Ariane** — chemin cliquable avec support du glisser-déposer pour déplacer des éléments entre dossiers.
- **Taille des dossiers** — la taille totale récursive est calculée et affichée pour chaque dossier.
- **Glisser-déposer** — faites glisser des fichiers ou dossiers vers un autre dossier ou segment du fil d'Ariane pour les déplacer ; faites glisser depuis l'OS pour uploader.
- **Sélection multiple** — sélectionnez plusieurs éléments via des cases à cocher ou shift-clic, puis téléchargez, déplacez ou supprimez en masse.
- **Renommage inline** — renommage directement dans la liste, avec gestion au clavier (Entrée / Échap) et perte de focus.
- **Prévisualisation** — aperçu dans le navigateur des images, PDF, fichiers audio et vidéo, sans téléchargement.
- **Téléchargement ZIP** — téléchargez un dossier entier ou une sélection de fichiers/dossiers sous forme d'archive ZIP.
- **Tags** — des tags à l'échelle de l'organisation (avec couleur) peuvent être appliqués à tout fichier ou dossier ; filtrage par tag dans le navigateur de fichiers.
- **Favoris (épingles)** — étoilez les fichiers et dossiers fréquemment consultés ; ils apparaissent dans une bande d'accès rapide en haut du navigateur.
- **Corbeille** — les éléments supprimés sont déplacés dans la corbeille, où ils peuvent être restaurés individuellement ou supprimés définitivement ; les admins peuvent vider toute la corbeille.
- **Recherche** — recherche plein texte sur les noms de fichiers et dossiers de l'organisation, avec déchiffrement des noms chiffrés avant la correspondance.
- **Progression d'upload** — barres de progression par fichier visibles pendant le chiffrement et l'envoi.

### Contrôle d'accès par dossier

Les admins et les admins de groupe peuvent définir des permissions par dossier pour des utilisateurs individuels ou des groupes :

| Niveau | Description |
|--------|-------------|
| manage | Peut lire, écrire et modifier les permissions du dossier |
| write | Peut uploader, renommer, déplacer et supprimer dans le dossier |
| read | Peut naviguer et télécharger depuis le dossier |
| none | Aucun accès ; le dossier est invisible |

Les permissions s'accumulent : le niveau d'accès effectif d'un utilisateur à un dossier est le niveau le plus élevé accordé directement ou via l'un des groupes auxquels il appartient.

#### Dossiers verrouillés et demandes d'accès

Un dossier dont l'accès effectif est `none` est affiché **verrouillé** (icône cadenas) dans le navigateur de fichiers. Le membre peut cliquer sur **Demander l'accès** pour soumettre une demande à l'admin. Les admins voient toutes les demandes en attente dans l'onglet **Demandes d'accès** du panneau de gestion et peuvent les approuver ou refuser d'un clic.

#### Accès effectif

- L'onglet **Profil** permet à chaque membre de consulter son propre accès effectif dossier par dossier.
- Dans l'onglet **Membres**, les admins peuvent afficher l'accès effectif de n'importe quel membre en ligne, sans quitter la liste.

### Liens de partage d'organisation

Les fichiers d'une organisation peuvent être partagés via des liens publics, indépendamment du système de partage personnel :

- **Générer un lien de partage** pour tout fichier ou dossier au sein de l'organisation.
- **Protection par mot de passe** — protection optionnelle par mot de passe pour accéder au lien.
- **Option à usage unique** — le lien est automatiquement révoqué après son premier accès réussi.
- **Gestion des partages** — liste de tous les liens de partage actifs de l'org et révocation à la demande.
- **Page d'accès public** — les destinataires accèdent au contenu via une URL dédiée ; le déchiffrement s'effectue côté client dans leur navigateur.

### Journal d'audit

Chaque action effectuée au sein d'une organisation est enregistrée dans un journal d'audit immuable :

- Les événements incluent : upload, téléchargement, suppression, renommage, déplacement de fichier, adhésion/départ de membre, changement de rôle, changement de permission, création/révocation de partage, provisionnement de clé.
- Les **champs chiffrés** (noms de fichiers, chemins) sont déchiffrés côté client avant affichage.
- **Export** — les admins peuvent exporter le journal d'audit complet sous forme de fichier.
- **Gestion de la rétention** — les admins peuvent supprimer les entrées d'audit antérieures à une date choisie.
- **Pagination** — bouton "charger plus" pour les historiques volumineux.

### Obligation MFA

Les propriétaires et admins d'une organisation peuvent exiger que tous les membres aient activé le MFA avant d'accéder à l'organisation. Les membres sans MFA actif voient un écran de blocage et sont redirigés vers leurs paramètres de compte.

### Tableau de bord et statistiques

Le tableau de bord de l'organisation offre une vue d'ensemble aux admins :

- Nombre total de membres, fichiers, dossiers.
- Activité sur les 7 derniers jours.
- Nombre de liens de partage actifs.
- Alerte quand des membres n'ont pas encore de clé d'organisation provisionnée.
- Accès direct au flux de provisionnement.

### CLI d'administration

Un outil en ligne de commande (`admin`) est disponible pour la gestion des organisations côté serveur :

```bash
./admin org create --name "Acme" --owner <user-id> --quota 10240
./admin org list
./admin org quota --id <org-id> --quota 20480
./admin org delete --id <org-id>
```

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
| Lire le contenu d'une organisation | Non — l'OrgKey n'est jamais stockée en clair |

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

Les permissions sont **modifiables à tout moment** après création du partage : un clic sur le chip bascule le droit et synchronise immédiatement avec le serveur.

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

### Restrictions par élément dans un partage par lien

Pour les dossiers partagés via lien public, il est possible de définir des droits **par sous-élément** indépendamment des permissions globales du lien. Un panneau latéral dans la boîte de dialogue de gestion permet de naviguer dans l'arborescence du dossier partagé et de configurer chaque entrée individuellement.

#### Niveaux d'accès pour les sous-dossiers

| Niveau | Comportement |
|--------|-------------|
| Accès complet | Le visiteur voit et peut interagir avec le dossier selon les droits globaux |
| Lecture seule | Le visiteur peut consulter le contenu mais ne peut pas y écrire |
| Masqué | Le dossier est invisible pour le visiteur |

Pour chaque fichier, deux droits supplémentaires sont réglables indépendamment :
- **Téléchargement** : autoriser ou bloquer le téléchargement (et la prévisualisation) de ce fichier spécifique.
- **Suppression** : autoriser ou protéger ce fichier contre la suppression.

#### Contrôles en masse

Des boutons de contrôle en masse permettent d'appliquer un réglage uniforme à tous les dossiers ou tous les fichiers du niveau courant en un seul clic, puis d'affiner élément par élément.

#### Navigation dans l'arborescence

Le panneau affiche un fil d'Ariane cliquable. Il est possible de descendre dans n'importe quel sous-dossier pour y configurer les restrictions, puis de remonter via le fil d'Ariane.

---

### Vue de gestion des partages

La page "Partages" centralise tous vos partages actifs en deux sections repliables :

- **Mes partages** — liste dédupliquée de vos ressources partagées avec : type (fichier / dossier), compteur de vues, date de création, date d'expiration, copie du lien en un clic, accès direct au dossier partagé dans l'arborescence, et gestion des droits.
- **Partagés avec moi** — liste des ressources que d'autres utilisateurs ont partagées avec vous.

---

### 3. Transfert P2P (appareil à appareil)

Le transfert P2P envoie des fichiers directement d'un appareil à un autre, chiffré de bout en bout, sans stockage serveur intermédiaire. **Il n'y a pas de limite de taille de fichier.**

WebRTC tente toujours une **connexion directe en premier** (LAN ou traversée NAT via STUN). Ce n'est que si aucun chemin direct ne peut être établi — à cause d'un NAT ou d'un pare-feu trop restrictif — que le transfert bascule sur un **relais TURN opéré par Kagibi**. Ce relais est un simple commutateur de flux : les données entrent et sortent en temps réel sans être écrites sur disque. **Le serveur TURN ne produit aucun log et ne peut pas accéder au contenu**, qui reste chiffré AES-256-GCM de bout en bout tout au long du transfert.

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
- **Quitter manuellement** — l'expéditeur ou le destinataire peut fermer la connexion à tout moment.

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
| Clé de fichier (`EncryptedKey`) | Chiffré (MasterKey ou OrgKey) |

### Données d'organisation

| Donnée | Format |
|--------|--------|
| Nom de l'organisation | Clair |
| Liste des membres et rôles | Clair |
| OrgKey par membre | Chiffré (clé publique RSA du membre) |
| GroupKey par groupe (backup admin) | Chiffré (OrgKey, AES-256-GCM) |
| GroupKey par membre de groupe | Chiffré (clé publique RSA du membre) |
| Noms de fichiers et dossiers de l'org | Chiffré (OrgKey, AES-256-GCM) |
| Contenu des fichiers sans groupe | Chiffré (FileKey wrappée avec OrgKey, AES-256-GCM) |
| Contenu des fichiers de groupe | Chiffré (FileKey wrappée avec GroupKey, AES-256-GCM) |
| Entrées du journal d'audit | Actions en clair ; chemins/noms chiffrés déchiffrés côté client |
| Permissions de dossier | Clair (IDs utilisateur/groupe + niveau d'accès) |
| Demandes d'accès | Clair (ID membre, ID dossier, statut, message optionnel) |

### Données sociales et de partage

- Liste d'amis et statut (en attente / accepté)
- Partages actifs : identifiant de ressource + clé chiffrée + permissions
- Liens publics : token + clé chiffrée + expiration + hash de mot de passe optionnel
- Invitations P2P : token + nom du fichier + taille + date d'expiration (contenu non stocké)
- Liens de partage d'organisation : token + clé chiffrée + hash de mot de passe optionnel + indicateur à usage unique

### Journaux de connexion (conformité LCEN)

Conformément à la loi française (LCEN article 6 II et décret 2021-1363), les données techniques suivantes sont conservées **1 an** :

| Événement journalisé | Données enregistrées |
|---|---|
| Création / suppression de compte | ID utilisateur, IP complète, IP anonymisée, user-agent, horodatage |
| Tentatives de connexion (succès / échec) | ID utilisateur, IP complète, IP anonymisée, horodatage |
| Changement de mot de passe / révocation de token | ID utilisateur, IP complète, IP anonymisée, horodatage |
| Accès aux fichiers | ID utilisateur, ID fichier, IP complète, IP anonymisée, horodatage |
| Création / révocation d'un lien de partage public | ID utilisateur, ressource, token, IP complète, user-agent, horodatage |
| Création d'un partage direct | ID propriétaire, ID destinataire, ressource, IP complète, user-agent, horodatage |
| Requêtes HTTP | IP anonymisée uniquement (CNIL 2021-122), user-agent, statut, durée |

**Politique IP** : l'IP complète est conservée uniquement dans les journaux d'événements de sécurité (cycle de vie du compte, auth, partages). Les journaux HTTP applicatifs contiennent uniquement l'IP anonymisée (dernier octet IPv4 / 80 bits IPv6 masqués), conformément à la délibération CNIL 2021-122. Les journaux sont centralisés dans Grafana Cloud Loki et conservés 1 an.

Ces données peuvent être communiquées aux autorités judiciaires ou administratives sur réquisition légale. Elles sont soumises à contrôle d'accès et ne contiennent jamais de clé de déchiffrement.

### Ce qui n'est pas collecté

- Contenu des fichiers (jamais en clair sur le serveur)
- Historique de navigation ou de recherche dans vos fichiers
- IP complète dans les logs HTTP standards (anonymisée uniquement — CNIL 2021-122)

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
