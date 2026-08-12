# Registre des activités de traitement — Kagibi

**Article 30 RGPD** — Tenu par le Responsable de traitement  
**Dernière mise à jour** : 2026-06-25

---

## Fiche 1 — Gestion des comptes utilisateurs

| Champ | Valeur |
|-------|--------|
| **Finalité** | Création, authentification et gestion des comptes |
| **Base légale** | Exécution du contrat (art. 6.1.b) |
| **Catégories de personnes** | Utilisateurs inscrits |
| **Catégories de données** | E-mail, mot de passe haché, nom d'affichage, clé publique RSA, date d'inscription |
| **Destinataires** | Infrastructure interne uniquement |
| **Transfert hors UE** | Non |
| **Durée de conservation** | Durée du compte + 30 jours après suppression |
| **Mesures de sécurité** | Hachage bcrypt, TLS, authentification JWT, révocation Redis |

---

## Fiche 2 — Authentification et sécurité de session

| Champ | Valeur |
|-------|--------|
| **Finalité** | Sécurisation des accès, prévention de la fraude |
| **Base légale** | Intérêt légitime (art. 6.1.f) — sécurité informatique |
| **Catégories de personnes** | Tous les visiteurs et utilisateurs |
| **Catégories de données** | Adresse IP (complète), ID utilisateur, timestamp, résultat (succès/échec), token JWT (haché) |
| **Destinataires** | Infrastructure interne, équipe sécurité |
| **Transfert hors UE** | Non |
| **Durée de conservation** | **1 an** (CNIL délibération 2021-122) |
| **Mesures de sécurité** | Accès restreint aux logs, chiffrement au repos, intégrité garantie |

---

## Fiche 3 — Journaux d'accès HTTP (logs applicatifs)

| Champ | Valeur |
|-------|--------|
| **Finalité** | Diagnostic technique, monitoring des performances, détection d'anomalies |
| **Base légale** | Intérêt légitime (art. 6.1.f) — maintenance et sécurité du SI |
| **Catégories de personnes** | Tous les utilisateurs |
| **Catégories de données** | IP tronquée (3 octets), méthode HTTP, chemin, statut, durée, ID utilisateur, ID de requête |
| **Destinataires** | Équipe technique interne, Grafana/Loki (hébergé en UE) |
| **Transfert hors UE** | Non |
| **Durée de conservation** | **1 an** |
| **Mesures de sécurité** | Pseudonymisation de l'IP, accès restreint via Grafana avec authentification |

---

## Fiche 4 — Stockage de fichiers chiffrés

| Champ | Valeur |
|-------|--------|
| **Finalité** | Stockage et partage sécurisé de fichiers |
| **Base légale** | Exécution du contrat (art. 6.1.b) |
| **Catégories de personnes** | Utilisateurs inscrits |
| **Catégories de données** | Fichiers chiffrés E2E, métadonnées (nom chiffré, taille, dates, ID S3) |
| **Destinataires** | Fournisseur de stockage objet (OVH, UE) |
| **Transfert hors UE** | Non |
| **Durée de conservation** | Durée du compte utilisateur |
| **Mesures de sécurité** | Chiffrement de bout en bout côté client, aucun accès serveur au contenu |

---

## Fiche 5 — Organisations et collaboration

| Champ | Valeur |
|-------|--------|
| **Finalité** | Gestion des espaces collaboratifs, partage de fichiers entre membres |
| **Base légale** | Exécution du contrat (art. 6.1.b) |
| **Catégories de personnes** | Membres d'organisations |
| **Catégories de données** | ID utilisateur, rôle, clé d'organisation chiffrée, date d'adhésion, source (interne/LDAP) |
| **Destinataires** | Infrastructure interne |
| **Transfert hors UE** | Non |
| **Durée de conservation** | Durée du membership |
| **Mesures de sécurité** | Clé d'organisation chiffrée RSA-OAEP par membre |

---

## Fiche 6 — Journal d'audit des organisations

| Champ | Valeur |
|-------|--------|
| **Finalité** | Traçabilité des actions administratives dans une organisation |
| **Base légale** | Intérêt légitime (art. 6.1.f) — obligation de sécurité de l'administrateur |
| **Catégories de personnes** | Membres d'organisations |
| **Catégories de données** | ID acteur, action, ID cible, type de cible, détail, timestamp |
| **Destinataires** | Administrateurs de l'organisation concernée |
| **Transfert hors UE** | Non |
| **Durée de conservation** | **1 an** (configurable par l'administrateur, minimum recommandé) |
| **Mesures de sécurité** | Accès restreint aux rôles admin/owner, export CSV disponible |

---

## Fiche 7 — Synchronisation LDAP / Active Directory

| Champ | Valeur |
|-------|--------|
| **Finalité** | Provisionnement automatique des membres d'une organisation depuis un annuaire d'entreprise |
| **Base légale** | Intérêt légitime (art. 6.1.f) + exécution du contrat pour les membres concernés |
| **Catégories de personnes** | Employés de l'organisation synchronisée |
| **Catégories de données** | E-mail, nom d'affichage, UID LDAP, appartenance aux groupes |
| **Destinataires** | Infrastructure interne |
| **Transfert hors UE** | Non |
| **Durée de conservation** | Durée du membership ; logs de sync : 90 jours |
| **Mesures de sécurité** | Mot de passe de liaison chiffré AES-256-GCM, connexion LDAP via StartTLS/LDAPS |
| **Information des personnes** | Les utilisateurs provisionnés reçoivent une invitation par e-mail |

---

## Fiche 8 — Communications par e-mail

| Champ | Valeur |
|-------|--------|
| **Finalité** | Envoi d'invitations, notifications de sécurité, e-mail de bienvenue |
| **Base légale** | Exécution du contrat (art. 6.1.b) |
| **Catégories de personnes** | Utilisateurs inscrits et invités |
| **Catégories de données** | Adresse e-mail, nom d'affichage, lien d'invitation |
| **Destinataires** | Fournisseur SMTP (hébergé en UE) |
| **Transfert hors UE** | Non |
| **Durée de conservation** | Logs d'envoi : 90 jours |
| **Mesures de sécurité** | TLS sur SMTP, pas de stockage du contenu des e-mails |

---

## Fiche 9 — Métriques et monitoring (Prometheus/Grafana)

| Champ | Valeur |
|-------|--------|
| **Finalité** | Surveillance des performances et de la disponibilité du service |
| **Base légale** | Intérêt légitime (art. 6.1.f) |
| **Catégories de personnes** | Aucune — métriques agrégées uniquement |
| **Catégories de données** | Compteurs agrégés (requêtes/s, latence p99, erreurs) — **aucune donnée individuelle** |
| **Destinataires** | Équipe technique interne via Grafana |
| **Transfert hors UE** | Non |
| **Durée de conservation** | 90 jours (rétention Prometheus par défaut) |
| **Mesures de sécurité** | Endpoint métriques non exposé publiquement (accès réseau interne uniquement) |

---

## Annexe — Sous-traitants (art. 28 RGPD)

| Sous-traitant | Données traitées | Localisation | DPA signé |
|---------------|-----------------|--------------|-----------|
| OVH SAS | Fichiers chiffrés, DB, Redis | France (UE) | Oui |
| [Fournisseur SMTP] | E-mails transactionnels | [À compléter] | Oui |

---

*Ce registre doit être tenu à jour à chaque nouveau traitement ou modification substantielle. Il est mis à disposition de la CNIL sur demande (art. 30.4 RGPD).*
