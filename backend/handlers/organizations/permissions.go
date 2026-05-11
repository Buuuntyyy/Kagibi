// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package organizations

import (
	"net/http"
	"strconv"

	"kagibi/backend/pkg"

	"github.com/gin-gonic/gin"
)

// ListPermissions returns all folder permission overrides for an org (admins/owners only).
func (h *OrgHandler) ListPermissions(c *gin.Context) {
	callerID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}
	ctx := c.Request.Context()

	callerRole, err := h.memberRole(ctx, orgID, callerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check membership"})
		return
	}
	if !canManage(callerRole) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	var perms []pkg.OrgFolderPermission
	if err := h.DB.NewSelect().Model(&perms).
		Where("org_id = ?", orgID).
		OrderExpr("user_id ASC, folder_path ASC").
		Scan(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list permissions"})
		return
	}
	c.JSON(http.StatusOK, perms)
}

// SetPermission creates or replaces a folder-level permission override for a specific user.
func (h *OrgHandler) SetPermission(c *gin.Context) {
	callerID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}

	var req struct {
		UserID       string `json:"user_id" binding:"required"`
		FolderPath   string `json:"folder_path"`
		Level        string `json:"level" binding:"required"` // read | write | manage | none
		PermCreate   bool   `json:"perm_create"`
		PermDelete   bool   `json:"perm_delete"`
		PermDownload *bool  `json:"perm_download"`
		PermMove     bool   `json:"perm_move"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validLevels := map[string]bool{"read": true, "write": true, "manage": true, "none": true}
	if !validLevels[req.Level] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid level: must be read, write, manage, or none"})
		return
	}

	ctx := c.Request.Context()
	callerRole, err := h.memberRole(ctx, orgID, callerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check membership"})
		return
	}
	if !canManage(callerRole) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	// Can only set permissions on actual members
	targetRole, err := h.memberRole(ctx, orgID, req.UserID)
	if err != nil || targetRole == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "target user is not a member of this organization"})
		return
	}
	// Admins cannot restrict owners
	if targetRole == "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot set permissions on the organization owner"})
		return
	}
	// Admins cannot set permissions on other admins
	if callerRole == "admin" && targetRole == "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admins cannot set permissions on other admins"})
		return
	}

	folderPath := normPath(req.FolderPath)

	permDownload := true
	if req.PermDownload != nil {
		permDownload = *req.PermDownload
	}

	// Derive perm flags from level if not explicitly set
	switch req.Level {
	case "manage":
		req.PermCreate = true
		req.PermDelete = true
		req.PermMove = true
	case "write":
		req.PermCreate = true
		req.PermMove = true
	}

	perm := &pkg.OrgFolderPermission{
		OrgID:        orgID,
		UserID:       req.UserID,
		FolderPath:   folderPath,
		Level:        req.Level,
		PermCreate:   req.PermCreate,
		PermDelete:   req.PermDelete,
		PermDownload: permDownload,
		PermMove:     req.PermMove,
	}

	if _, err := h.DB.NewInsert().Model(perm).
		On("CONFLICT (org_id, user_id, folder_path) DO UPDATE").
		Set("level = EXCLUDED.level, perm_create = EXCLUDED.perm_create, perm_delete = EXCLUDED.perm_delete, perm_download = EXCLUDED.perm_download, perm_move = EXCLUDED.perm_move").
		Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set permission"})
		return
	}
	h.logAudit(ctx, orgID, callerID, "permission_set", req.UserID, "permission",
		req.Level+" on "+folderPath)
	c.JSON(http.StatusOK, perm)
}

// DeletePermission removes a specific folder permission override.
func (h *OrgHandler) DeletePermission(c *gin.Context) {
	callerID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}

	var req struct {
		UserID     string `json:"user_id" binding:"required"`
		FolderPath string `json:"folder_path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	callerRole, err := h.memberRole(ctx, orgID, callerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check membership"})
		return
	}
	if !canManage(callerRole) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	folderPath := normPath(req.FolderPath)

	res, err := h.DB.NewDelete().Model((*pkg.OrgFolderPermission)(nil)).
		Where("org_id = ? AND user_id = ? AND folder_path = ?", orgID, req.UserID, folderPath).
		Exec(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete permission"})
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "permission override not found"})
		return
	}
	h.logAudit(ctx, orgID, callerID, "permission_removed", req.UserID, "permission", folderPath)
	c.JSON(http.StatusOK, gin.H{"message": "permission removed"})
}

// GetMyPermission returns the caller's effective permission level on a given folder path.
func (h *OrgHandler) GetMyPermission(c *gin.Context) {
	userID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}

	folderPath := normPath(c.Query("path"))
	ctx := c.Request.Context()

	perm, err := h.resolvePermission(ctx, orgID, userID, folderPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve permission"})
		return
	}

	role, _ := h.memberRole(ctx, orgID, userID)

	c.JSON(http.StatusOK, gin.H{
		"level":        levelToString(perm),
		"role":         role,
		"folder_path":  folderPath,
		"can_read":     perm >= PermRead,
		"can_write":    perm >= PermWrite,
		"can_manage":   perm >= PermManage,
	})
}
