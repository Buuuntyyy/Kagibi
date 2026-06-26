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
		"ip_anon", logger.AnonymiseIP(ip),
	)
}

// LogUnauthorizedAccess enregistre un accès refusé.
func LogUnauthorizedAccess(ctx context.Context, userID, resource, ip string) {
	seclog().Log(ctx, slog.LevelWarn, "access.denied",
		"event_type", "access.denied",
		"user_id", userID,
		"resource", resource,
		"ip_anon", logger.AnonymiseIP(ip),
	)
}

// LogSuspiciousActivity enregistre une activité suspecte.
func LogSuspiciousActivity(ctx context.Context, userID, activity, ip string) {
	seclog().Log(ctx, slog.LevelWarn, "security.suspicious",
		"event_type", "security.suspicious",
		"user_id", userID,
		"activity", activity,
		"ip_anon", logger.AnonymiseIP(ip),
	)
}

// LogFileAccess enregistre un accès à un fichier.
func LogFileAccess(ctx context.Context, userID, fileID, ip string, success bool) {
	level := slog.LevelDebug
	if !success {
		level = slog.LevelWarn
	}
	seclog().Log(ctx, level, "file.access",
		"event_type", "file.access",
		"user_id", userID,
		"file_id", fileID,
		"ip_anon", logger.AnonymiseIP(ip),
		"success", success,
	)
}

// LogProfileUpdate enregistre une mise à jour de profil.
func LogProfileUpdate(ctx context.Context, userID, ip string) {
	seclog().Log(ctx, slog.LevelInfo, "user.profile_updated",
		"event_type", "user.profile_updated",
		"user_id", userID,
		"ip_anon", logger.AnonymiseIP(ip),
	)
}

// LogRateLimitExceeded enregistre un dépassement de limite de débit.
func LogRateLimitExceeded(ctx context.Context, ip, endpoint string) {
	seclog().Log(ctx, slog.LevelWarn, "ratelimit.exceeded",
		"event_type", "ratelimit.exceeded",
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
		"ip_anon", logger.AnonymiseIP(ip),
	)
}
