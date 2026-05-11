// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package organizations

import (
	"net/http"
	"strconv"

	"kagibi/backend/pkg"

	"github.com/gin-gonic/gin"
)

// ListAuditLog returns paginated audit events for an org.
// Restricted to admins and owners.
func (h *OrgHandler) ListAuditLog(c *gin.Context) {
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

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	const pageSize = 50
	offset := (page - 1) * pageSize

	var entries []pkg.OrgAuditLog
	if err := h.DB.NewSelect().Model(&entries).
		Where("org_id = ?", orgID).
		OrderExpr("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Scan(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch audit log"})
		return
	}

	c.JSON(http.StatusOK, entries)
}

// GetOrgAllFileKeys returns all {id, encrypted_key} pairs for non-deleted files in the org.
// Used by the frontend to re-wrap keys during OrgKey rotation.
// Restricted to admins and owners.
func (h *OrgHandler) GetOrgAllFileKeys(c *gin.Context) {
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

	type fileKeyEntry struct {
		ID           int64  `json:"id"`
		EncryptedKey string `json:"encrypted_key"`
	}

	var files []pkg.OrgFile
	if err := h.DB.NewSelect().Model(&files).
		Column("id", "encrypted_key").
		Where("org_id = ?", orgID).
		Scan(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch file keys"})
		return
	}

	result := make([]fileKeyEntry, len(files))
	for i, f := range files {
		result[i] = fileKeyEntry{ID: f.ID, EncryptedKey: f.EncryptedKey}
	}
	c.JSON(http.StatusOK, result)
}
