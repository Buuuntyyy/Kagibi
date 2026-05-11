// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package organizations

import (
	"context"
	"database/sql"
	"log"
	"path"
	"strings"

	"kagibi/backend/pkg"

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
	DB *bun.DB
}

func NewOrgHandler(db *bun.DB) *OrgHandler {
	return &OrgHandler{DB: db}
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

// resolvePermission returns the effective PermLevel for userID on folderPath within orgID.
// It walks the path hierarchy from most specific to root, using the first override found.
// If no override exists, the member's role default is applied.
func (h *OrgHandler) resolvePermission(ctx context.Context, orgID int64, userID, folderPath string) (PermLevel, error) {
	role, err := h.memberRole(ctx, orgID, userID)
	if err != nil {
		return PermNone, err
	}
	if role == "" {
		return PermNone, nil
	}

	p := normPath(folderPath)
	for {
		var perm pkg.OrgFolderPermission
		if err := h.DB.NewSelect().Model(&perm).
			Where("org_id = ? AND user_id = ? AND folder_path = ?", orgID, userID, p).
			Scan(ctx); err == nil {
			return levelFromString(perm.Level), nil
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
// It walks the path hierarchy (most-specific first) and returns the perm_download flag
// of the first matching override. If no override is found, download is allowed by default.
// Owners and admins cannot have their download access restricted (no perm records apply to them).
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
		var perm pkg.OrgFolderPermission
		if err := h.DB.NewSelect().Model(&perm).
			Where("org_id = ? AND user_id = ? AND folder_path = ?", orgID, userID, p).
			Scan(ctx); err == nil {
			return perm.PermDownload, nil
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
