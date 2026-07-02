# Référentiel des événements à journaliser — Kagibi

Chaque événement est émis en JSON via slog avec les champs : `time`, `level`, `service`, `env`, `msg` (nom de l'événement), et les attributs métier listés ci-dessous.

---

## Catégorie 1 — Authentification (`component: security`, `event_type: AUTH_*`)

| Événement (`msg`) | Niveau | Champs obligatoires | Déjà implémenté |
|-------------------|--------|---------------------|-----------------|
| `auth.attempt` | INFO/WARN | `user_id`, `ip`, `success`, `reason` | ✅ |
| `auth.token_invalid` | WARN | `provider`, `err` | ✅ |
| `auth.token_revoked` | INFO | `user_id`, `reason` | ✅ |
| `auth.missing_user_id_claim` | WARN | `provider` | ✅ |
| `auth.revocation_check_failed` | ERROR | `user_id`, `err` | ✅ |
| `auth.session_expired` | INFO | `user_id` | ⬜ |
| `auth.logout` | INFO | `user_id`, `ip` | ⬜ |
| `auth.signup` | INFO | `user_id`, `ip`, `provider` | ⬜ |
| `auth.signup_failed` | WARN | `ip`, `reason` | ⬜ |
| `auth.token_refreshed` | DEBUG | `user_id` | ⬜ |

---

## Catégorie 2 — MFA (`event_type: MFA_*`)

| Événement (`msg`) | Niveau | Champs obligatoires | Déjà implémenté |
|-------------------|--------|---------------------|-----------------|
| `auth.mfa` (action=ENROLL) | INFO | `user_id`, `ip`, `success` | ✅ |
| `auth.mfa` (action=VERIFY) | INFO/WARN | `user_id`, `ip`, `success` | ✅ |
| `auth.mfa` (action=UNENROLL) | INFO | `user_id`, `ip`, `success` | ✅ |
| `auth.mfa` (action=CHALLENGE) | DEBUG | `user_id` | ⬜ |
| `auth.mfa_bruteforce_detected` | ERROR | `user_id`, `ip`, `attempts` | ⬜ |

---

## Catégorie 3 — Compte utilisateur (`event_type: ACCOUNT_*`)

| Événement (`msg`) | Niveau | Champs obligatoires | Déjà implémenté |
|-------------------|--------|---------------------|-----------------|
| `auth.password_change` | INFO | `user_id`, `ip` | ✅ |
| `account.deleted` | INFO | `user_id`, `ip` | ✅ |
| `account.email_changed` | INFO | `user_id`, `ip`, `old_email_hash` | ⬜ |
| `account.profile_updated` | DEBUG | `user_id`, `ip`, `fields_changed` | ⬜ |
| `account.export_requested` | INFO | `user_id`, `ip` | ⬜ |
| `account.recovery_initiated` | INFO | `ip`, `email_hash` | ⬜ |
| `account.recovery_completed` | INFO | `user_id`, `ip` | ⬜ |

---

## Catégorie 4 — Accès HTTP (`msg: http_request`)

| Événement | Niveau | Champs obligatoires | Déjà implémenté |
|-----------|--------|---------------------|-----------------|
| Toute requête HTTP | INFO (2xx/3xx), WARN (4xx), ERROR (5xx) | `request_id`, `method`, `path`, `status`, `duration_ms`, `user_id`, `ip_anon` | ✅ |

---

## Catégorie 5 — Fichiers (`event_type: FILE_*`)

| Événement (`msg`) | Niveau | Champs obligatoires | Déjà implémenté |
|-------------------|--------|---------------------|-----------------|
| `file.upload` | INFO | `user_id`, `file_id`, `size_bytes`, `ip_anon` | ⬜ |
| `file.download` | INFO | `user_id`, `file_id`, `ip_anon` | ⬜ |
| `file.access` (denied) | WARN | `user_id`, `file_id`, `ip`, `success=false` | ✅ |
| `file.delete` | INFO | `user_id`, `file_id`, `ip_anon` | ⬜ |
| `file.rename` | DEBUG | `user_id`, `file_id`, `ip_anon` | ⬜ |
| `file.move` | DEBUG | `user_id`, `file_id`, `destination`, `ip_anon` | ⬜ |
| `file.share_created` | INFO | `user_id`, `file_id`, `share_type`, `ip_anon` | ⬜ |
| `file.share_accessed` | INFO | `file_id`, `share_token_hash`, `ip_anon` | ⬜ |
| `file.share_revoked` | INFO | `user_id`, `file_id`, `ip_anon` | ⬜ |
| `file.version_restored` | INFO | `user_id`, `file_id`, `version_id`, `ip_anon` | ⬜ |
| `file.presign_suspicious` | WARN | `user_id`, `file_id`, `ip`, `reason` | ⬜ |

---

## Catégorie 6 — Organisations (`event_type: ORG_*`)

| Événement (`msg`) | Niveau | Champs obligatoires | Déjà implémenté |
|-------------------|--------|---------------------|-----------------|
| `org.created` | INFO | `user_id`, `org_id`, `org_name` | ⬜ |
| `org.deleted` | INFO | `user_id`, `org_id` | ⬜ |
| `org.member_added` | INFO | `actor_id`, `org_id`, `target_user_id`, `role` | ⬜ |
| `org.member_removed` | INFO | `actor_id`, `org_id`, `target_user_id` | ⬜ |
| `org.member_role_changed` | INFO | `actor_id`, `org_id`, `target_user_id`, `old_role`, `new_role` | ⬜ |
| `org.invitation_created` | INFO | `actor_id`, `org_id`, `role`, `token_hash` | ⬜ |
| `org.invitation_accepted` | INFO | `user_id`, `org_id`, `token_hash` | ⬜ |
| `org.invitation_revoked` | INFO | `actor_id`, `org_id`, `token_hash` | ⬜ |
| `org.key_rotated` | INFO | `actor_id`, `org_id` | ⬜ |
| `org.audit_cleared` | INFO | `actor_id`, `org_id`, `mode` | ⬜ |

---

## Catégorie 7 — LDAP (`event_type: LDAP_*`)

| Événement (`msg`) | Niveau | Champs obligatoires | Déjà implémenté |
|-------------------|--------|---------------------|-----------------|
| `ldap.sync_ok` | INFO | `org_id`, `duration_ms`, `users_found`, `users_added`, `users_suspended`, `users_removed` | ✅ |
| `ldap.sync_error` | ERROR | `org_id`, `err` | ✅ |
| `ldap.user_suspended` | INFO | `org_id`, `user_id` | ✅ |
| `ldap.user_removed` | INFO | `org_id`, `user_id`, `reason` | ✅ |
| `ldap.invite_email_failed` | WARN | `email`, `err` | ✅ |
| `ldap.safeguard_triggered` | ERROR | `org_id`, `reason`, `users_found`, `threshold` | ⬜ |
| `ldap.config_saved` | INFO | `actor_id`, `org_id`, `enabled` | ⬜ |
| `ldap.test_connection` | INFO | `actor_id`, `org_id`, `success`, `users_found` | ⬜ |

---

## Catégorie 8 — Sécurité réseau (`event_type: SECURITY_*`)

| Événement (`msg`) | Niveau | Champs obligatoires | Déjà implémenté |
|-------------------|--------|---------------------|-----------------|
| `ratelimit.exceeded` | WARN | `ip`, `endpoint` | ✅ |
| `ratelimit.redis_error` | WARN | `key`, `err` | ✅ |
| `access.unauthorized` | WARN | `user_id`, `resource`, `ip` | ✅ |
| `security.suspicious` | WARN | `user_id`, `activity`, `ip` | ✅ |
| `security.csp_violation` | WARN | `ip_anon`, `document_uri`, `violated_directive` | ⬜ |
| `security.invalid_content_type` | WARN | `user_id`, `ip`, `endpoint` | ⬜ |

---

## Catégorie 9 — Infrastructure (`component: infra`)

| Événement (`msg`) | Niveau | Champs obligatoires | Déjà implémenté |
|-------------------|--------|---------------------|-----------------|
| `server.started` | INFO | `port`, `gin_mode`, `auth_provider` | ⬜ |
| `server.shutdown` | INFO | `duration_ms` | ⬜ |
| `migrations_ok` | INFO | — | ✅ |
| `migration_failed` | ERROR | `err` | ✅ |
| `redis.connected` | INFO | — | ⬜ |
| `redis.disconnected` | ERROR | `err` | ⬜ |
| `s3.initialized` | INFO | — | ⬜ |
| `s3.error` | ERROR | `operation`, `err` | ⬜ |
| `worker.started` | INFO | `worker_type` | ⬜ |
| `worker.panic_recovered` | ERROR | `worker_type`, `err` | ⬜ |

---

## Catégorie 10 — RGPD / Conformité (`event_type: GDPR_*`)

| Événement (`msg`) | Niveau | Champs obligatoires | Déjà implémenté |
|-------------------|--------|---------------------|-----------------|
| `account.deleted` | INFO | `user_id`, `ip` | ✅ |
| `gdpr.data_export_requested` | INFO | `user_id`, `ip` | ⬜ |
| `gdpr.data_export_delivered` | INFO | `user_id` | ⬜ |
| `gdpr.erasure_s3_files` | INFO | `user_id`, `files_deleted` | ⬜ |
| `gdpr.erasure_db_complete` | INFO | `user_id` | ⬜ |
| `gdpr.retention_purge` | INFO | `log_type`, `entries_deleted`, `older_than_days` | ⬜ |

---

## Règles de nommage

- **`msg`** : `domaine.action` en snake_case — ex : `auth.login_failed`, `file.download`
- **`event_type`** : majuscules avec underscore — ex : `AUTH_ATTEMPT`, `FILE_DOWNLOAD` (pour filtrage Loki)
- **`level`** :
  - `DEBUG` — événement technique fin, non conservé en production
  - `INFO` — événement métier normal
  - `WARN` — événement anormal non critique (tentative échouée, dégradation)
  - `ERROR` — erreur nécessitant une action
- Ne jamais logger : mots de passe, tokens complets, clés de chiffrement, contenu de fichiers

---

## Requêtes LogQL utiles (Grafana / Loki)

```logql
# Toutes les erreurs d'authentification des 24h
{service="kagibi-backend"} | json | event_type="AUTH_ATTEMPT" | success="false" | __error__=""

# Erreurs 5xx par endpoint
sum by (path) (
  count_over_time({service="kagibi-backend"} | json | status >= 500 [1h])
)

# Activités suspectes (rate limit + accès refusés)
{service="kagibi-backend", component="security"} | json | level="warn"

# Synchronisations LDAP en erreur
{service="kagibi-backend"} | json | event_type="LDAP_SYNC" | error != ""

# Volume de logs par niveau sur 24h
sum by (level) (count_over_time({service="kagibi-backend"}[24h]))
```
