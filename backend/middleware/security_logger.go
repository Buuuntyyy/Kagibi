// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package middleware — security_logger.go fournit un logger slog dédié aux
// événements de sécurité. Tous les événements sont émis en JSON sur stdout
// avec le champ event_type pour faciliter le filtrage dans Loki/Grafana.
package middleware

import (
	"context"
	"log/slog"
)

// Security est le logger de sécurité applicatif.
// Il utilise le logger slog par défaut (initialisé dans pkg/logger) avec un
// attribut "component" fixe pour permettre le filtrage dans Loki via :
//   {service="kagibi-backend"} | json | component="security"
var Security = slog.Default().With("component", "security")

// LogAuthAttempt enregistre une tentative d'authentification.
func LogAuthAttempt(userID, ip string, success bool, reason string) {
	lvl := slog.LevelInfo
	if !success {
		lvl = slog.LevelWarn
	}
	Security.Log(context.TODO(), lvl, "auth.attempt",
		"event_type", "AUTH_ATTEMPT",
		"user_id", userID,
		"ip", ip,
		"success", success,
		"reason", reason,
	)
}

// LogTokenRevoked enregistre une révocation de token (déconnexion, changement de MDP).
func LogTokenRevoked(userID, reason string) {
	Security.Info("auth.token_revoked",
		"event_type", "TOKEN_REVOKED",
		"user_id", userID,
		"reason", reason,
	)
}

// LogPasswordChange enregistre un changement de mot de passe.
func LogPasswordChange(userID, ip string) {
	Security.Info("auth.password_change",
		"event_type", "PASSWORD_CHANGE",
		"user_id", userID,
		"ip", ip,
	)
}

// LogMFAEvent enregistre un événement MFA (enroll, verify, unenroll).
func LogMFAEvent(userID, ip, action string, success bool) {
	lvl := slog.LevelInfo
	if !success {
		lvl = slog.LevelWarn
	}
	Security.Log(context.TODO(), lvl, "auth.mfa",
		"event_type", "MFA_"+action,
		"user_id", userID,
		"ip", ip,
		"success", success,
	)
}

// LogUnauthorizedAccess enregistre un accès refusé à une ressource.
func LogUnauthorizedAccess(userID, resource, ip string) {
	Security.Warn("access.unauthorized",
		"event_type", "UNAUTHORIZED_ACCESS",
		"user_id", userID,
		"resource", resource,
		"ip", ip,
	)
}

// LogRateLimitExceeded enregistre un dépassement de limite de taux.
func LogRateLimitExceeded(ip, endpoint string) {
	Security.Warn("ratelimit.exceeded",
		"event_type", "RATE_LIMIT_EXCEEDED",
		"ip", ip,
		"endpoint", endpoint,
	)
}

// LogFileAccess enregistre un accès à un fichier (téléchargement, partage).
func LogFileAccess(userID, fileID, action, ip string, success bool) {
	lvl := slog.LevelInfo
	if !success {
		lvl = slog.LevelWarn
	}
	Security.Log(context.TODO(), lvl, "file.access",
		"event_type", "FILE_"+action,
		"user_id", userID,
		"file_id", fileID,
		"ip", ip,
		"success", success,
	)
}

// LogSuspiciousActivity enregistre une activité suspecte.
func LogSuspiciousActivity(userID, activity, ip string) {
	Security.Warn("security.suspicious",
		"event_type", "SUSPICIOUS_ACTIVITY",
		"user_id", userID,
		"activity", activity,
		"ip", ip,
	)
}

// LogAccountDeletion enregistre la suppression d'un compte (RGPD art. 17).
func LogAccountDeletion(userID, ip string) {
	Security.Info("account.deleted",
		"event_type", "ACCOUNT_DELETION",
		"user_id", userID,
		"ip", ip,
	)
}

// LogLDAPSync enregistre le résultat d'une synchronisation LDAP.
func LogLDAPSync(orgID int64, usersFound, added, suspended, removed int, syncErr string) {
	lvl := slog.LevelInfo
	if syncErr != "" {
		lvl = slog.LevelError
	}
	Security.Log(context.TODO(), lvl, "ldap.sync",
		"event_type", "LDAP_SYNC",
		"org_id", orgID,
		"users_found", usersFound,
		"users_added", added,
		"users_suspended", suspended,
		"users_removed", removed,
		"error", syncErr,
	)
}
