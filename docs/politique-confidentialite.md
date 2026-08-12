# Politique de confidentialité — Kagibi

**Version** : 1.1  
**Date d'entrée en vigueur** : 2026-06-25  
**Responsable de traitement** : [Nom de la société / Nom du responsable], [Adresse], [Email de contact]  
**DPO / Contact RGPD** : [dpo@votre-domaine.fr]

---

## 1. Objet

La présente politique de confidentialité décrit la manière dont Kagibi collecte, utilise, stocke et protège vos données à caractère personnel, conformément au **Règlement (UE) 2016/679** (RGPD) et à la **loi Informatique et Libertés** modifiée.

---

## 2. Identité du responsable de traitement

| Champ | Valeur |
|-------|--------|
| Dénomination | [À compléter] |
| Forme juridique | [À compléter] |
| Siège social | [À compléter] |
| Email de contact | [À compléter] |
| Email DPO | [À compléter] |

---

## 3. Données collectées et finalités

### 3.1 Compte utilisateur

| Donnée | Finalité | Base légale | Durée |
|--------|----------|-------------|-------|
| Adresse e-mail | Authentification, notifications | Exécution du contrat (art. 6.1.b) | Durée du compte + 30 jours |
| Mot de passe (haché bcrypt) | Authentification | Exécution du contrat | Durée du compte |
| Nom d'affichage | Identification dans l'interface | Exécution du contrat | Durée du compte |
| Clé publique RSA | Chiffrement de bout en bout | Exécution du contrat | Durée du compte |
| Date d'inscription | Administration | Intérêt légitime (art. 6.1.f) | Durée du compte + 5 ans |
| Paramètres TOTP (MFA) | Sécurité du compte | Exécution du contrat | Durée du compte |

### 3.2 Fichiers et données chiffrées

Vos fichiers sont **chiffrés de bout en bout** côté client avant transmission. Kagibi ne dispose d'aucun accès au contenu en clair de vos fichiers. Les métadonnées suivantes sont conservées :

| Donnée | Finalité | Durée |
|--------|----------|-------|
| Nom de fichier (chiffré) | Affichage dans l'interface | Durée du compte |
| Taille du fichier | Quota de stockage | Durée du compte |
| Date de création/modification | Historique | Durée du compte |
| Identifiant S3 | Récupération technique | Durée du compte |

### 3.3 Journaux techniques (logs)

Conformément aux **recommandations de la CNIL** et à l'**article L34-1 du CPCE**, les logs suivants sont collectés automatiquement :

| Type de log | Données | Finalité | Durée de conservation |
|-------------|---------|----------|----------------------|
| Logs d'accès HTTP | Méthode, chemin, statut HTTP, durée, ID de requête, ID utilisateur, **IP tronquée** (3 derniers octets conservés) | Diagnostic, sécurité, statistiques | **1 an** |
| Logs d'authentification | Tentatives de connexion (succès/échec), IP complète, ID utilisateur | Détection de fraude, audit de sécurité | **1 an** |
| Logs MFA | Actions TOTP (enrôlement, vérification, suppression), IP, ID utilisateur | Audit de sécurité | **1 an** |
| Logs de sécurité | Accès refusés, rate limiting, activités suspectes, IP complète | Sécurité informatique | **1 an** |
| Logs LDAP | Résultats de synchronisation, erreurs (sans contenu des attributs sensibles) | Maintenance | **90 jours** |
| Logs d'erreurs applicatives | Messages d'erreur internes (sans données personnelles) | Diagnostic | **90 jours** |

> **Note IP** : Sauf pour les logs d'authentification et de sécurité (où l'IP complète est nécessaire à la détection de fraude et conservée 1 an), l'adresse IP est tronquée à 3 octets (`192.168.1.x`) dans les logs applicatifs courants. Cette mesure est conforme à la recommandation CNIL.

### 3.4 Organisations et partage

| Donnée | Finalité | Durée |
|--------|----------|-------|
| Membership d'organisation | Gestion des droits | Durée du membership |
| Rôle (owner/admin/member/viewer) | Contrôle d'accès | Durée du membership |
| Journal d'audit de l'organisation | Traçabilité des actions admin | 1 an (configurable) |
| Clé d'organisation chiffrée | Chiffrement E2E des fichiers de l'org | Durée du membership |

---

## 4. Destinataires des données

Vos données ne sont **jamais vendues ni cédées** à des tiers à des fins commerciales.

| Destinataire | Données | Raison |
|-------------|---------|--------|
| Hébergeur (OVH / infrastructure Rancher) | Toutes les données chiffrées | Infrastructure technique |
| Fournisseur de stockage objet (OVH Object Storage) | Fichiers chiffrés E2E | Stockage des fichiers |
| Fournisseur Redis (cache) | Sessions actives, rate limiting | Sécurité et performance |
| Service d'e-mail transactionnel | Adresse e-mail uniquement | Envoi de notifications |

Tous les sous-traitants sont soumis à un **accord de traitement des données (DPA)** et traitent les données exclusivement dans l'**Union Européenne**.

---

## 5. Transferts hors UE

Aucun transfert de données à caractère personnel n'est effectué hors de l'Union Européenne. L'ensemble de l'infrastructure est hébergée dans des datacentres situés en France et en Europe.

---

## 6. Sécurité

Kagibi met en œuvre les mesures techniques et organisationnelles suivantes :

- **Chiffrement de bout en bout** : vos fichiers sont chiffrés côté client (libsodium / XSalsa20-Poly1305) — Kagibi ne peut pas accéder à leur contenu
- **Hachage des mots de passe** : bcrypt avec facteur de coût adaptatif
- **Authentification à deux facteurs** (TOTP/RFC 6238) disponible
- **TLS 1.2+ obligatoire** sur tous les endpoints
- **Révocation de session** instantanée via Redis
- **Rate limiting** sur tous les endpoints sensibles
- **Journalisation de sécurité** avec conservation 1 an
- **Audits réguliers** du code et des dépendances

---

## 7. Vos droits

Conformément au RGPD (articles 15 à 22), vous disposez des droits suivants :

| Droit | Description | Comment l'exercer |
|-------|-------------|-------------------|
| **Accès** (art. 15) | Obtenir une copie de vos données | Via l'interface → Paramètres → Exporter mes données |
| **Rectification** (art. 16) | Corriger des données inexactes | Via l'interface → Paramètres → Profil |
| **Effacement** (art. 17) | Supprimer votre compte et vos données | Via l'interface → Paramètres → Supprimer mon compte |
| **Portabilité** (art. 20) | Recevoir vos données dans un format structuré | Via l'interface → Paramètres → Exporter mes données |
| **Opposition** (art. 21) | Vous opposer à certains traitements | Par e-mail à [dpo@votre-domaine.fr] |
| **Limitation** (art. 18) | Demander la limitation du traitement | Par e-mail à [dpo@votre-domaine.fr] |

**Délai de réponse** : 30 jours maximum, sans frais.

Pour exercer vos droits : **[dpo@votre-domaine.fr]**

En cas de réclamation non résolue, vous pouvez saisir la **CNIL** : [www.cnil.fr/fr/plaintes](https://www.cnil.fr/fr/plaintes)

---

## 8. Cookies et traceurs

Kagibi n'utilise **aucun cookie de tracking ou publicitaire**. Les seuls éléments stockés côté client sont :

| Élément | Type | Finalité | Durée |
|---------|------|----------|-------|
| Token JWT | localStorage | Session authentifiée | Durée de la session |
| Préférences d'interface (thème, etc.) | localStorage | Confort d'utilisation | Illimitée (effaçable) |

Aucune bannière de cookies n'est requise car aucun traceur non essentiel n'est utilisé.

---

## 9. Modifications de la politique

Toute modification substantielle sera notifiée par e-mail et/ou notification dans l'interface au moins **30 jours avant** son entrée en vigueur. La date de version figure en haut du présent document.

---

## 10. Contact

Pour toute question relative à la présente politique ou au traitement de vos données personnelles :

- **E-mail DPO** : [dpo@votre-domaine.fr]
- **Courrier** : [Adresse postale du responsable de traitement]
