// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package organizations

import (
	"net/http"
	"strconv"
	"time"

	"kagibi/backend/pkg"

	"github.com/gin-gonic/gin"
)

func (h *OrgHandler) UpdateOrg(c *gin.Context) {
	userID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}
	ctx := c.Request.Context()

	role, err := h.memberRole(ctx, orgID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check membership"})
		return
	}
	if !canManage(role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	var req struct {
		Name           *string `json:"name"`
		Description    *string `json:"description"`
		StorageQuotaMB *int64  `json:"storage_quota_mb"` // owner only
		RequireMFA     *bool   `json:"require_mfa"`      // admin/owner
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var org pkg.Organization
	if err := h.DB.NewSelect().Model(&org).Where("id = ?", orgID).Scan(ctx); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		return
	}

	if req.Name != nil {
		org.Name = *req.Name
	}
	if req.Description != nil {
		org.Description = *req.Description
	}
	if req.StorageQuotaMB != nil && isOwner(role) {
		org.StorageQuotaMB = *req.StorageQuotaMB
	}
	if req.RequireMFA != nil {
		org.RequireMFA = *req.RequireMFA
	}
	org.UpdatedAt = time.Now()

	if _, err := h.DB.NewUpdate().Model(&org).WherePK().
		Column("name", "description", "storage_quota_mb", "require_mfa", "updated_at").
		Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update organization"})
		return
	}
	c.JSON(http.StatusOK, OrgResponse{Organization: org, MyRole: role})
}

func (h *OrgHandler) DeleteOrg(c *gin.Context) {
	userID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}
	ctx := c.Request.Context()

	role, err := h.memberRole(ctx, orgID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check membership"})
		return
	}
	if !isOwner(role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the owner can delete an organization"})
		return
	}

	org := &pkg.Organization{ID: orgID}
	if _, err := h.DB.NewDelete().Model(org).WherePK().Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete organization"})
		return
	}

	// Revoke all active invitations so pending join links stop working
	_, _ = h.DB.NewUpdate().Model((*pkg.OrgInvitation)(nil)).
		Set("status = ?", "revoked").
		Where("org_id = ? AND status = 'active'", orgID).
		Exec(ctx)

	c.JSON(http.StatusOK, gin.H{"message": "organization deleted"})
}
