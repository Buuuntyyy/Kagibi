// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package middleware — security_logger.go expose des helpers de log de
// sécurité structurés via slog. Les événements sont émis avec l'attribut
// component="security" pour faciliter le filtrage dans Loki / Grafana.
//
// Exemples de requêtes LogQL :
//
//	{service="kagibi-backend"} | json | component="security"
//	{service="kagibi-backend"} | json | event_type="auth.failed"
package middleware

import (
	"context"
	"log/slog"

	"kagibi/backend/pkg/logger"
)

// seclog retourne toujours le logger par défaut courant avec l'attribut
// component="security". On ne le cache pas en var globale pour éviter de
// capturer le logger avant que applogger.Init() l'ait configuré.
func seclog() *slog.Logger { return slog.Default().With("component", "security") }

// LogAuthAttempt enregistre une tentative d'authentification.
func LogAuthAttempt(ctx context.Context, userID, ip string, success bool, reason string) {
	level := slog.LevelInfo
	if !success {
		level = slog.LevelWarn
	}
	seclog().Log(ctx, level, "auth.attempt",
		"event_type", "auth.attempt",
		"user_id", userID,
		"ip", ip,
		"ip_anon", logger.AnonymiseIP(ip),
		"success", success,
		"reason", reason,
	)
}

// LogPasswordChange enregistre un changement de mot de passe.
func LogPasswordChange(ctx context.Context, userID, ip string) {
	seclog().Log(ctx, slog.LevelInfo, "auth.password_changed",
		"event_type", "auth.password_changed",
		"user_id", userID,
		"ip", ip,
		"ip_anon", logger.AnonymiseIP(ip),
	)
}

// LogUnauthorizedAccess enregistre un accès refusé.
func LogUnauthorizedAccess(ctx context.Context, userID, resource, ip string) {
	seclog().Log(ctx, slog.LevelWarn, "access.denied",
		"event_type", "access.denied",
		"user_id", userID,
		"resource", resource,
		"ip", ip,
		"ip_anon", logger.AnonymiseIP(ip),
	)
}

// LogSuspiciousActivity enregistre une activité suspecte.
func LogSuspiciousActivity(ctx context.Context, userID, activity, ip string) {
	seclog().Log(ctx, slog.LevelWarn, "security.suspicious",
		"event_type", "security.suspicious",
		"user_id", userID,
		"activity", activity,
		"ip", ip,
		"ip_anon", logger.AnonymiseIP(ip),
	)
}

// LogFileAccess enregistre un accès à un fichier.
// Le niveau est Info (succès) ou Warn (échec) — jamais Debug, pour satisfaire
// l'obligation de conservation des logs d'accès (LCEN, 1 an).
func LogFileAccess(ctx context.Context, userID, fileID, ip string, success bool) {
	level := slog.LevelInfo
	if !success {
		level = slog.LevelWarn
	}
	seclog().Log(ctx, level, "file.access",
		"event_type", "file.access",
		"user_id", userID,
		"file_id", fileID,
		"ip", ip,
		"ip_anon", logger.AnonymiseIP(ip),
		"success", success,
	)
}

// LogProfileUpdate enregistre une mise à jour de profil.
func LogProfileUpdate(ctx context.Context, userID, ip string) {
	seclog().Log(ctx, slog.LevelInfo, "user.profile_updated",
		"event_type", "user.profile_updated",
		"user_id", userID,
		"ip", ip,
		"ip_anon", logger.AnonymiseIP(ip),
	)
}

// LogRateLimitExceeded enregistre un dépassement de limite de débit.
func LogRateLimitExceeded(ctx context.Context, ip, endpoint string) {
	seclog().Log(ctx, slog.LevelWarn, "ratelimit.exceeded",
		"event_type", "ratelimit.exceeded",
		"ip", ip,
		"ip_anon", logger.AnonymiseIP(ip),
		"endpoint", endpoint,
	)
}

// LogLDAPSync enregistre le résultat d'une synchronisation LDAP.
func LogLDAPSync(ctx context.Context, orgID int64, usersFound, added, suspended, removed int, syncErr string) {
	level := slog.LevelInfo
	if syncErr != "" {
		level = slog.LevelError
	}
	seclog().Log(ctx, level, "ldap.sync",
		"event_type", "ldap.sync",
		"org_id", orgID,
		"users_found", usersFound,
		"users_added", added,
		"users_suspended", suspended,
		"users_removed", removed,
		"error", syncErr,
	)
}

// LogTokenRevoked enregistre la révocation d'un token (déconnexion, MFA unenroll…).
func LogTokenRevoked(ctx context.Context, userID, reason, ip string) {
	seclog().Log(ctx, slog.LevelInfo, "auth.token_revoked",
		"event_type", "auth.token_revoked",
		"user_id", userID,
		"reason", reason,
		"ip", ip,
		"ip_anon", logger.AnonymiseIP(ip),
	)
}

// LogAccountCreated enregistre la création d'un compte utilisateur.
func LogAccountCreated(ctx context.Context, userID, ip, userAgent string) {
	seclog().Log(ctx, slog.LevelInfo, "account.created",
		"event_type", "account.created",
		"user_id", userID,
		"ip", ip,
		"ip_anon", logger.AnonymiseIP(ip),
		"user_agent", userAgent,
	)
}

// LogAccountDeleted enregistre la suppression définitive d'un compte (RGPD art. 17).
func LogAccountDeleted(ctx context.Context, userID, ip, userAgent string) {
	seclog().Log(ctx, slog.LevelInfo, "account.deleted",
		"event_type", "account.deleted",
		"user_id", userID,
		"ip", ip,
		"ip_anon", logger.AnonymiseIP(ip),
		"user_agent", userAgent,
	)
}

// LogShareCreated enregistre la création d'un lien de partage public.
func LogShareCreated(ctx context.Context, userID, resourceType string, resourceID int64, token, ip, userAgent string) {
	seclog().Log(ctx, slog.LevelInfo, "share.created",
		"event_type", "share.created",
		"user_id", userID,
		"resource_type", resourceType,
		"resource_id", resourceID,
		"token", token,
		"ip", ip,
		"ip_anon", logger.AnonymiseIP(ip),
		"user_agent", userAgent,
	)
}

// LogShareRevoked enregistre la suppression d'un lien de partage public.
func LogShareRevoked(ctx context.Context, userID, shareID, ip, userAgent string) {
	seclog().Log(ctx, slog.LevelInfo, "share.revoked",
		"event_type", "share.revoked",
		"user_id", userID,
		"share_id", shareID,
		"ip", ip,
		"ip_anon", logger.AnonymiseIP(ip),
		"user_agent", userAgent,
	)
}

// LogDirectShareCreated enregistre la création d'un partage direct entre utilisateurs.
func LogDirectShareCreated(ctx context.Context, ownerID, recipientID, resourceType string, resourceID int64, ip, userAgent string) {
	seclog().Log(ctx, slog.LevelInfo, "share.direct_created",
		"event_type", "share.direct_created",
		"owner_id", ownerID,
		"recipient_id", recipientID,
		"resource_type", resourceType,
		"resource_id", resourceID,
		"ip", ip,
		"ip_anon", logger.AnonymiseIP(ip),
		"user_agent", userAgent,
	)
}
