// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package organizations

import (
	"database/sql"
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

// TransferOwnership atomically transfers org ownership from the caller to a target member.
// The caller must be the current owner. The client pre-encrypts the org key with the
// target's RSA public key so E2E encryption is maintained.
// POST /orgs/:orgID/transfer-ownership
func (h *OrgHandler) TransferOwnership(c *gin.Context) {
	callerID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}

	var req struct {
		TargetMemberID  int64  `json:"target_member_id" binding:"required"`
		EncryptedOrgKey string `json:"encrypted_org_key" binding:"required"`
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
	if !isOwner(callerRole) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the organization owner can transfer ownership"})
		return
	}

	var target pkg.OrgMember
	if err := h.DB.NewSelect().Model(&target).
		Where("id = ? AND org_id = ?", req.TargetMemberID, orgID).
		Scan(ctx); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "target member not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch target member"})
		return
	}
	if target.UserID == callerID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot transfer ownership to yourself"})
		return
	}

	tx, err := h.DB.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to begin transaction"})
		return
	}
	defer tx.Rollback()

	if _, err := tx.NewUpdate().Model((*pkg.OrgMember)(nil)).
		Set("role = 'owner'").
		Set("encrypted_org_key = ?", req.EncryptedOrgKey).
		Where("id = ? AND org_id = ?", req.TargetMemberID, orgID).
		Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to promote new owner"})
		return
	}

	if _, err := tx.NewUpdate().Model((*pkg.OrgMember)(nil)).
		Set("role = 'admin'").
		Where("org_id = ? AND user_id = ?", orgID, callerID).
		Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to demote caller"})
		return
	}

	if _, err := tx.NewUpdate().Model((*pkg.Organization)(nil)).
		Set("owner_id = ?", target.UserID).
		Where("id = ?", orgID).
		Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update organization owner"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit ownership transfer"})
		return
	}

	h.logAudit(ctx, orgID, callerID, "ownership_transferred", target.UserID, "user",
		"ownership transferred to "+target.UserID)

	c.JSON(http.StatusOK, gin.H{"message": "ownership transferred"})
}
