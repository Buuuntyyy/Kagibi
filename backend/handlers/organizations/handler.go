// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package organizations

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"kagibi/backend/pkg"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/uptrace/bun"
)

// PermLevel represents an effective access level on an org folder path.
type PermLevel int

const (
	PermNone   PermLevel = 0
	PermRead   PermLevel = 1
	PermWrite  PermLevel = 2 // implies read
	PermManage PermLevel = 3 // implies write
)

type OrgHandler struct {
	DB          *bun.DB
	RedisClient *redis.Client
}

func NewOrgHandler(db *bun.DB, redisClient *redis.Client) *OrgHandler {
	return &OrgHandler{DB: db, RedisClient: redisClient}
}

// CallerCaps holds pre-resolved authorization context for a caller within an org.
type CallerCaps struct {
	OrgRole       string
	AdminGroupIDs map[int64]bool
}

func (c CallerCaps) IsOrgAdmin() bool   { return canManage(c.OrgRole) }
func (c CallerCaps) IsGroupAdmin() bool { return len(c.AdminGroupIDs) > 0 }

// resolveCallerCaps fetches the caller's org role and the set of group IDs where they hold
// the "admin" role. Single source of truth for all permission checks in this package.
func (h *OrgHandler) resolveCallerCaps(ctx context.Context, orgID int64, callerID string) (CallerCaps, error) {
	caps := CallerCaps{AdminGroupIDs: make(map[int64]bool)}

	role, err := h.memberRole(ctx, orgID, callerID)
	if err != nil {
		return caps, err
	}
	caps.OrgRole = role

	type groupIDRow struct {
		GroupID int64 `bun:"group_id"`
	}
	var rows []groupIDRow
	if err := h.DB.NewSelect().
		TableExpr("org_group_members ogm").
		ColumnExpr("ogm.group_id").
		Join("JOIN org_groups og ON og.id = ogm.group_id").
		Where("og.org_id = ? AND ogm.user_id = ? AND ogm.role = 'admin'", orgID, callerID).
		Scan(ctx, &rows); err != nil {
		return caps, err
	}
	for _, r := range rows {
		caps.AdminGroupIDs[r.GroupID] = true
	}
	return caps, nil
}

// memberRole returns the caller's role in the org ("owner","admin","member","viewer"),
// or "" if the user is not a member. Returns an error only on DB failures.
func (h *OrgHandler) memberRole(ctx context.Context, orgID int64, userID string) (string, error) {
	var m pkg.OrgMember
	if err := h.DB.NewSelect().Model(&m).
		Where("org_id = ? AND user_id = ?", orgID, userID).
		Scan(ctx); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return m.Role, nil
}

// groupRole returns the caller's role within a specific group ("admin" | "member"), or "" if not a member.
func (h *OrgHandler) groupRole(ctx context.Context, groupID int64, userID string) (string, error) {
	var gm pkg.OrgGroupMember
	if err := h.DB.NewSelect().Model(&gm).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Scan(ctx); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return gm.Role, nil
}

// resolvePermission returns the effective PermLevel for userID on folderPath within orgID.
//
// Resolution order (most specific path wins, walking toward root):
//  1. Direct user override — takes absolute precedence at that path level.
//  2. Group overrides — most permissive among all groups (group "none" never hard-blocks).
//  3. Role default — applied when no override is found anywhere in the hierarchy.
//
// Owners and admins always receive PermManage regardless of any override.
func (h *OrgHandler) resolvePermission(ctx context.Context, orgID int64, userID, folderPath string) (PermLevel, error) {
	role, err := h.memberRole(ctx, orgID, userID)
	if err != nil {
		return PermNone, err
	}
	if role == "" {
		return PermNone, nil
	}
	if canManage(role) {
		return PermManage, nil
	}

	p := normPath(folderPath)
	for {
		// 1. Direct user override — highest priority at this path level.
		var directPerm pkg.OrgFolderPermission
		if err := h.DB.NewSelect().Model(&directPerm).
			Where("org_id = ? AND user_id = ? AND folder_path = ?", orgID, userID, p).
			Scan(ctx); err == nil {
			return levelFromString(directPerm.Level), nil
		}

		// 2. Group overrides — take the most permissive across all groups the user belongs to.
		var groupLevels []struct {
			Level string `bun:"level"`
		}
		userInGroupOnPath := false
		if err := h.DB.NewSelect().
			TableExpr("org_group_permissions ogp").
			ColumnExpr("ogp.level").
			Join("JOIN org_group_members ogm ON ogm.group_id = ogp.group_id").
			Where("ogp.org_id = ? AND ogm.user_id = ? AND ogp.folder_path = ?", orgID, userID, p).
			Scan(ctx, &groupLevels); err == nil && len(groupLevels) > 0 {
			userInGroupOnPath = true
			best := PermNone
			for _, gl := range groupLevels {
				if l := levelFromString(gl.Level); l > best {
					best = l
				}
			}
			if best > PermNone {
				return best, nil
			}
			// All groups had "none" — group none does not hard-block; continue walking up.
		}

		// 2b. If any group has a restrict_to_groups permission on this path and the
		// user belongs to none of them, deny access.
		// restrict_to_groups=false means the permission is additive, not exclusive:
		// members without a group permission still fall through to the role default.
		if !userInGroupOnPath {
			var groupPermCount int
			if err := h.DB.NewSelect().
				TableExpr("org_group_permissions").
				ColumnExpr("COUNT(*)").
				Where("org_id = ? AND folder_path = ? AND restrict_to_groups = true", orgID, p).
				Scan(ctx, &groupPermCount); err == nil && groupPermCount > 0 {
				return PermNone, nil
			}
		}

		if p == "/" {
			break
		}
		p = path.Dir(p)
	}
	return roleDefaultLevel(role), nil
}

func normPath(p string) string {
	p = strings.ReplaceAll(p, "\\", "/")
	p = path.Clean("/" + strings.TrimPrefix(p, "/"))
	if p == "." {
		return "/"
	}
	return p
}

func levelFromString(s string) PermLevel {
	switch s {
	case "read":
		return PermRead
	case "write":
		return PermWrite
	case "manage":
		return PermManage
	default:
		return PermNone
	}
}

func levelToString(l PermLevel) string {
	switch l {
	case PermRead:
		return "read"
	case PermWrite:
		return "write"
	case PermManage:
		return "manage"
	default:
		return "none"
	}
}

func roleDefaultLevel(role string) PermLevel {
	switch role {
	case "owner", "admin":
		return PermManage
	case "member":
		return PermWrite
	case "viewer":
		return PermRead
	default:
		return PermNone
	}
}

// resolveDownloadAllowed reports whether userID may download files from folderPath.
// Direct user override takes precedence; if absent, any group granting download is sufficient.
// Owners and admins always have download access.
func (h *OrgHandler) resolveDownloadAllowed(ctx context.Context, orgID int64, userID, folderPath string) (bool, error) {
	role, err := h.memberRole(ctx, orgID, userID)
	if err != nil {
		return false, err
	}
	if role == "" {
		return false, nil
	}
	if canManage(role) {
		return true, nil
	}

	p := normPath(folderPath)
	for {
		// Direct user override.
		var perm pkg.OrgFolderPermission
		if err := h.DB.NewSelect().Model(&perm).
			Where("org_id = ? AND user_id = ? AND folder_path = ?", orgID, userID, p).
			Scan(ctx); err == nil {
			return perm.PermDownload, nil
		}

		// Group overrides — allowed if any group permits download.
		var groupDL []struct {
			PermDownload bool `bun:"perm_download"`
		}
		if err := h.DB.NewSelect().
			TableExpr("org_group_permissions ogp").
			ColumnExpr("ogp.perm_download").
			Join("JOIN org_group_members ogm ON ogm.group_id = ogp.group_id").
			Where("ogp.org_id = ? AND ogm.user_id = ? AND ogp.folder_path = ?", orgID, userID, p).
			Scan(ctx, &groupDL); err == nil && len(groupDL) > 0 {
			for _, g := range groupDL {
				if g.PermDownload {
					return true, nil
				}
			}
			return false, nil
		}

		// Restricted path: user has no group permission here — deny download.
		var restrictionCount int
		if err := h.DB.NewSelect().
			TableExpr("org_group_permissions").
			ColumnExpr("COUNT(*)").
			Where("org_id = ? AND folder_path = ? AND restrict_to_groups = true", orgID, p).
			Scan(ctx, &restrictionCount); err == nil && restrictionCount > 0 {
			return false, nil
		}

		if p == "/" {
			break
		}
		p = path.Dir(p)
	}
	return true, nil
}

// logAudit inserts an audit entry. Errors are printed to stderr but never
// propagated — audit failures must not interrupt the main operation.
func (h *OrgHandler) logAudit(ctx context.Context, orgID int64, actorID, action, targetID, targetType, detail string) {
	entry := &pkg.OrgAuditLog{
		OrgID:      orgID,
		ActorID:    actorID,
		Action:     action,
		TargetID:   targetID,
		TargetType: targetType,
		Detail:     detail,
	}
	if _, err := h.DB.NewInsert().Model(entry).Exec(ctx); err != nil {
		log.Printf("AUDIT LOG ERROR org=%d action=%s: %v", orgID, action, err)
	}
}

func canManage(role string) bool { return role == "owner" || role == "admin" }
func isOwner(role string) bool   { return role == "owner" }

// checkMFAEnforcement returns true if the caller may proceed.
// If the org has require_mfa=true and the caller has not enrolled MFA,
// it writes a 403 response and returns false.
func (h *OrgHandler) checkMFAEnforcement(c *gin.Context, orgID int64, userID string) bool {
	ctx := c.Request.Context()

	var org pkg.Organization
	if err := h.DB.NewSelect().Model(&org).Column("require_mfa").Where("id = ?", orgID).Scan(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "mfa_check_failed",
			"message": "Unable to verify organization MFA requirements. Please try again.",
		})
		return false
	}
	if !org.RequireMFA {
		return true
	}

	var settings struct {
		MFAEnabled bool `bun:"mfa_enabled"`
	}
	err := h.DB.NewSelect().
		TableExpr("user_security_settings").
		ColumnExpr("mfa_enabled").
		Where("user_id = ?", userID).
		Scan(ctx, &settings)
	if err != nil || !settings.MFAEnabled {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "mfa_required",
			"message": "This organization requires two-factor authentication. Please enable MFA in your account settings.",
		})
		return false
	}
	return true
}

// hasOrgAccess reports whether userID may use the Organizations feature.
// Self-hosted instances (BILLING_ENABLED=false) grant access to everyone.
// On cloud, only paid-plan users (Pro / Business) may create, list or join orgs.
func (h *OrgHandler) hasOrgAccess(userID string) bool {
	if os.Getenv("BILLING_ENABLED") == "false" {
		return true
	}
	plan, err := pkg.FindUserPlanByUserID(h.DB, userID)
	if err != nil || plan == nil {
		return false
	}
	return plan.Plan != pkg.PlanFree
}
