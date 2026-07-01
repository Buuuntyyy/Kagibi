// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package organizations

import (
	"net/http"
	"strconv"

	"kagibi/backend/pkg"

	"github.com/gin-gonic/gin"
)

// EffectiveAccessEntry represents what level of access a user has on a folder.
type EffectiveAccessEntry struct {
	FolderPath string `json:"folder_path"`
	FolderName string `json:"folder_name"`
	Level      string `json:"level"` // "none" | "read" | "write" | "manage"
}

// GetMyEffectiveAccess returns the calling user's effective access level on every
// folder in the org. Intended for regular members who want to see what they can access.
func (h *OrgHandler) GetMyEffectiveAccess(c *gin.Context) {
	userID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}
	ctx := c.Request.Context()

	role, err := h.memberRole(ctx, orgID, userID)
	if err != nil || role == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member of this organization"})
		return
	}

	entries, err := h.computeEffectiveAccess(c, orgID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to compute effective access"})
		return
	}
	c.JSON(http.StatusOK, entries)
}

// GetMemberEffectiveAccess returns a specific member's effective access level on every
// folder in the org. Admin/owner only.
func (h *OrgHandler) GetMemberEffectiveAccess(c *gin.Context) {
	callerID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}
	targetUserID := c.Param("userID")
	ctx := c.Request.Context()

	callerRole, err := h.memberRole(ctx, orgID, callerID)
	if err != nil || !canManage(callerRole) {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	// Verify target is actually a member.
	targetRole, err := h.memberRole(ctx, orgID, targetUserID)
	if err != nil || targetRole == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "member not found"})
		return
	}

	entries, err := h.computeEffectiveAccess(c, orgID, targetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to compute effective access"})
		return
	}
	c.JSON(http.StatusOK, entries)
}

// computeEffectiveAccess fetches all org folders and resolves the effective permission
// for targetUserID on each one.
func (h *OrgHandler) computeEffectiveAccess(c *gin.Context, orgID int64, targetUserID string) ([]EffectiveAccessEntry, error) {
	ctx := c.Request.Context()

	var folders []pkg.OrgFolder
	if err := h.DB.NewSelect().Model(&folders).
		Where("org_id = ?", orgID).
		OrderExpr("path ASC").
		Scan(ctx); err != nil {
		return nil, err
	}

	entries := make([]EffectiveAccessEntry, 0, len(folders))
	for _, f := range folders {
		level, err := h.resolvePermission(ctx, orgID, targetUserID, f.Path)
		if err != nil {
			return nil, err
		}
		entries = append(entries, EffectiveAccessEntry{
			FolderPath: f.Path,
			FolderName: f.Name,
			Level:      levelToString(level),
		})
	}
	return entries, nil
}
