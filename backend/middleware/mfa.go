// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// mfaActionFreshness is how recently a session must have completed TOTP verification
// for a per-action MFA gate (downloads / destructive / email change) to be satisfied.
// A stale aal2 session is asked to step up again. Login-level enforcement
// (EnforceMFAOnLogin) is not subject to this window — being MFA-authenticated for the
// login session is a one-time requirement, not a per-action one.
const mfaActionFreshness = 10 * time.Minute

// mfaGates holds a user's MFA enforcement preferences for the current request.
type mfaGates struct {
	MFAEnabled        bool
	RequireOnLogin    bool
	RequireOnDestruct bool
	RequireOnDownload bool
	RequireOnEmail    bool
	RequireOnRecovery bool
}

// fetchMFAGates loads the user's MFA enforcement preferences, caching the result on
// the gin context so several MFA middlewares on the same request share one DB read.
// A missing settings row (or a read error) yields all-false gates — enforcement then
// simply passes, matching the "no preference set" default.
func fetchMFAGates(c *gin.Context, db *bun.DB, userID string) mfaGates {
	if cached, ok := c.Get("mfa_gates"); ok {
		return cached.(mfaGates)
	}

	var row struct {
		MFAEnabled        bool `bun:"mfa_enabled"`
		RequireOnLogin    bool `bun:"require_mfa_on_login"`
		RequireOnDestruct bool `bun:"require_mfa_on_destructive_actions"`
		RequireOnDownload bool `bun:"require_mfa_on_downloads"`
		RequireOnEmail    bool `bun:"require_mfa_on_email_change"`
		RequireOnRecovery bool `bun:"require_mfa_on_recovery_change"`
	}

	gates := mfaGates{}
	err := db.NewSelect().
		TableExpr("user_security_settings").
		ColumnExpr("mfa_enabled, require_mfa_on_login, require_mfa_on_destructive_actions, require_mfa_on_downloads, require_mfa_on_email_change, require_mfa_on_recovery_change").
		Where("user_id = ?", userID).
		Scan(c.Request.Context(), &row)
	if err == nil {
		gates = mfaGates{
			MFAEnabled:        row.MFAEnabled,
			RequireOnLogin:    row.RequireOnLogin,
			RequireOnDestruct: row.RequireOnDestruct,
			RequireOnDownload: row.RequireOnDownload,
			RequireOnEmail:    row.RequireOnEmail,
			RequireOnRecovery: row.RequireOnRecovery,
		}
	}

	c.Set("mfa_gates", gates)
	return gates
}

// mfaStepUpAllowed lists the protected-group paths that must stay reachable with an
// aal1 session while a user completes MFA step-up: identity, encryption keys, session
// management, and the settings the step-up UI reads. The TOTP challenge/verify
// endpoints live on a separate route group and are never gated here.
func mfaStepUpAllowed(path string) bool {
	switch path {
	case "/api/v1/auth/keys",
		"/api/v1/auth/logout",
		"/api/v1/auth/update-password",
		"/api/v1/auth/register",
		"/api/v1/users/me",
		"/api/v1/users/security-settings":
		return true
	}
	return false
}

// mfaEnrolled reports whether the request carries the signed "mfa":"enabled" claim.
func mfaEnrolled(c *gin.Context) bool {
	return c.GetString("mfa") == "enabled"
}

// mfaFresh reports whether the session completed TOTP verification within
// mfaActionFreshness, based on the signed "mfa_at" claim.
func mfaFresh(c *gin.Context) bool {
	v, ok := c.Get("mfa_at")
	if !ok {
		return false
	}
	at, ok := v.(int64)
	if !ok || at <= 0 {
		return false
	}
	return time.Since(time.Unix(at, 0)) < mfaActionFreshness
}

// EnforceMFAOnLogin blocks an aal1 session from protected resources when the user
// has MFA enabled and require_mfa_on_login is set; the session must first step up to
// aal2 via /auth/mfa/verify. It is a no-op (no DB lookup) for sessions already at
// aal2 and for users without an enrolled factor (no "mfa" claim) — i.e. the vast
// majority of requests — so the hot path is unaffected.
func EnforceMFAOnLogin(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetString("aal") == "aal2" || !mfaEnrolled(c) {
			c.Next()
			return
		}
		if mfaStepUpAllowed(c.Request.URL.Path) {
			c.Next()
			return
		}
		gates := fetchMFAGates(c, db, c.GetString("user_id"))
		if gates.MFAEnabled && gates.RequireOnLogin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "mfa_required"})
			return
		}
		c.Next()
	}
}

// RequireMFAForAction enforces a per-action MFA requirement for a specific route
// (action is "download", "destructive", "email_change" or "recovery_change"). Users without an enrolled
// factor pass without a DB lookup, as do sessions that completed a *recent* TOTP
// verification (aal2 within mfaActionFreshness). Otherwise the matching preference is
// consulted and a 403 {"error":"mfa_required"} is returned when set, prompting the
// client to step up and retry — so the requirement is genuinely per-action rather than
// once-per-7-day-session.
func RequireMFAForAction(db *bun.DB, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !mfaEnrolled(c) {
			c.Next()
			return
		}
		if c.GetString("aal") == "aal2" && mfaFresh(c) {
			c.Next()
			return
		}
		gates := fetchMFAGates(c, db, c.GetString("user_id"))
		if !gates.MFAEnabled {
			c.Next()
			return
		}
		required := false
		switch action {
		case "download":
			required = gates.RequireOnDownload
		case "destructive":
			required = gates.RequireOnDestruct
		case "email_change":
			required = gates.RequireOnEmail
		case "recovery_change":
			required = gates.RequireOnRecovery
		}
		if required {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "mfa_required"})
			return
		}
		c.Next()
	}
}
