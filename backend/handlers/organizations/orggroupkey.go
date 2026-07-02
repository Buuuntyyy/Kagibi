// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package organizations

import (
	"database/sql"
	"net/http"
	"strconv"

	"kagibi/backend/pkg"

	"github.com/gin-gonic/gin"
)

// GetGroupKey returns the group key material the caller needs to decrypt group files.
//
// Response:
//   - encrypted_key          — from org_group_keys (RSA-OAEP, for this caller).
//     Present only when a per-member entry exists.
//   - encrypted_group_key    — from org_groups (AES-GCM wrapped with org key).
//     Returned only to org admins/owners, so they can provision new members or
//     recover the group key via their own org key without needing a separate entry.
func (h *OrgHandler) GetGroupKey(c *gin.Context) {
	callerID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}
	groupID, err := strconv.ParseInt(c.Param("groupID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}

	ctx := c.Request.Context()
	role, err := h.memberRole(ctx, orgID, callerID)
	if err != nil || role == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member of this organization"})
		return
	}

	// Fetch caller's personal entry (RSA-wrapped group key).
	var gk pkg.OrgGroupKey
	callerKeyErr := h.DB.NewSelect().Model(&gk).
		Where("group_id = ? AND user_id = ?", groupID, callerID).
		Scan(ctx)

	resp := gin.H{}
	if callerKeyErr == nil {
		resp["encrypted_key"] = gk.EncryptedKey
	}

	// Admins/owners additionally receive the org-key-wrapped backup so they can
	// provision keys for members who don't have an entry yet.
	if canManage(role) {
		var grp pkg.OrgGroup
		if err := h.DB.NewSelect().Model(&grp).
			Where("id = ? AND org_id = ?", groupID, orgID).
			Scan(ctx); err == nil && grp.EncryptedGroupKey != "" {
			resp["encrypted_group_key"] = grp.EncryptedGroupKey
		}
	}

	if len(resp) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "group key not available for your account"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// InitializeGroupKey is called once by an admin to set up E2E encryption for a group.
//
// The client sends:
//   - encrypted_group_key  — group key wrapped with org key (AES-GCM), for admin recovery
//   - member_keys          — group key wrapped with each provisioned member's RSA public key
func (h *OrgHandler) InitializeGroupKey(c *gin.Context) {
	callerID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}
	groupID, err := strconv.ParseInt(c.Param("groupID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}

	var req struct {
		EncryptedGroupKey string `json:"encrypted_group_key" binding:"required"`
		MemberKeys        []struct {
			UserID       string `json:"user_id"       binding:"required"`
			EncryptedKey string `json:"encrypted_key" binding:"required"`
		} `json:"member_keys"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	role, err := h.memberRole(ctx, orgID, callerID)
	if err != nil || !canManage(role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	// Verify the group belongs to this org.
	var grp pkg.OrgGroup
	if err := h.DB.NewSelect().Model(&grp).
		Where("id = ? AND org_id = ?", groupID, orgID).
		Scan(ctx); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch group"})
		return
	}
	if grp.EncryptedGroupKey != "" {
		c.JSON(http.StatusConflict, gin.H{"error": "group key already initialized — use rotate-group-key to replace it"})
		return
	}

	tx, err := h.DB.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to begin transaction"})
		return
	}
	defer tx.Rollback()

	// Store the org-key-wrapped backup on the group.
	if _, err := tx.NewUpdate().Model((*pkg.OrgGroup)(nil)).
		Set("encrypted_group_key = ?", req.EncryptedGroupKey).
		Where("id = ?", groupID).
		Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store group key"})
		return
	}

	// Upsert per-member entries.
	for _, mk := range req.MemberKeys {
		entry := &pkg.OrgGroupKey{
			GroupID:      groupID,
			UserID:       mk.UserID,
			EncryptedKey: mk.EncryptedKey,
		}
		if _, err := tx.NewInsert().Model(entry).
			On("CONFLICT (group_id, user_id) DO UPDATE").
			Set("encrypted_key = EXCLUDED.encrypted_key").
			Exec(ctx); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store member key"})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit"})
		return
	}

	h.logAudit(ctx, orgID, callerID, "group_key_initialized", strconv.FormatInt(groupID, 10), "group", grp.Name)
	c.JSON(http.StatusOK, gin.H{"message": "group key initialized", "members_provisioned": len(req.MemberKeys)})
}

// ProvisionGroupKeyForMember adds or updates a group key entry for a single member.
// Admin/owner only.  Used after adding a new member to the group.
func (h *OrgHandler) ProvisionGroupKeyForMember(c *gin.Context) {
	callerID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}
	groupID, err := strconv.ParseInt(c.Param("groupID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}
	targetUserID := c.Param("userID")

	var req struct {
		EncryptedKey string `json:"encrypted_key" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	role, err := h.memberRole(ctx, orgID, callerID)
	if err != nil || !canManage(role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	// Verify target is a member of this group.
	var count int
	if err := h.DB.NewSelect().TableExpr("org_group_members").
		ColumnExpr("COUNT(*)").
		Where("group_id = ? AND user_id = ?", groupID, targetUserID).
		Scan(ctx, &count); err != nil || count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "user is not a member of this group"})
		return
	}

	entry := &pkg.OrgGroupKey{
		GroupID:      groupID,
		UserID:       targetUserID,
		EncryptedKey: req.EncryptedKey,
	}
	if _, err := h.DB.NewInsert().Model(entry).
		On("CONFLICT (group_id, user_id) DO UPDATE").
		Set("encrypted_key = EXCLUDED.encrypted_key").
		Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to provision key"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "group key provisioned"})
}

// RotateGroupKey replaces the group key for all members and re-wraps all group file keys.
// Admin/owner only.  All crypto happens client-side — this endpoint only stores the results.
func (h *OrgHandler) RotateGroupKey(c *gin.Context) {
	callerID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}
	groupID, err := strconv.ParseInt(c.Param("groupID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}

	var req struct {
		EncryptedGroupKey string `json:"encrypted_group_key" binding:"required"` // new group key wrapped with org key
		MemberKeys        []struct {
			UserID       string `json:"user_id"       binding:"required"`
			EncryptedKey string `json:"encrypted_key" binding:"required"`
		} `json:"member_keys" binding:"required"`
		FileKeys []struct {
			FileID       int64  `json:"file_id"       binding:"required"`
			EncryptedKey string `json:"encrypted_key" binding:"required"`
		} `json:"file_keys"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	role, err := h.memberRole(ctx, orgID, callerID)
	if err != nil || !canManage(role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	// Verify group belongs to org.
	var grp pkg.OrgGroup
	if err := h.DB.NewSelect().Model(&grp).
		Where("id = ? AND org_id = ?", groupID, orgID).
		Scan(ctx); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch group"})
		return
	}

	tx, err := h.DB.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to begin transaction"})
		return
	}
	defer tx.Rollback()

	// Update org-key-wrapped backup.
	if _, err := tx.NewUpdate().Model((*pkg.OrgGroup)(nil)).
		Set("encrypted_group_key = ?", req.EncryptedGroupKey).
		Where("id = ?", groupID).
		Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update group key"})
		return
	}

	// Replace per-member entries.
	for _, mk := range req.MemberKeys {
		entry := &pkg.OrgGroupKey{GroupID: groupID, UserID: mk.UserID, EncryptedKey: mk.EncryptedKey}
		if _, err := tx.NewInsert().Model(entry).
			On("CONFLICT (group_id, user_id) DO UPDATE").
			Set("encrypted_key = EXCLUDED.encrypted_key").
			Exec(ctx); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update member key"})
			return
		}
	}

	// Re-wrap file keys.
	for _, fk := range req.FileKeys {
		if _, err := tx.NewUpdate().Model((*pkg.OrgFile)(nil)).
			Set("encrypted_key = ?", fk.EncryptedKey).
			Where("id = ? AND org_id = ? AND group_id = ?", fk.FileID, orgID, groupID).
			Exec(ctx); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update file key"})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit"})
		return
	}

	h.logAudit(ctx, orgID, callerID, "group_key_rotated", strconv.FormatInt(groupID, 10), "group", grp.Name)
	c.JSON(http.StatusOK, gin.H{
		"message":             "group key rotated",
		"members_provisioned": len(req.MemberKeys),
		"file_keys_rewrapped": len(req.FileKeys),
	})
}

// GetGroupKeyProvisionedMembers returns the list of user IDs that have a key entry
// for this group, so the admin UI can display per-member provisioning status.
func (h *OrgHandler) GetGroupKeyProvisionedMembers(c *gin.Context) {
	callerID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}
	groupID, err := strconv.ParseInt(c.Param("groupID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}

	ctx := c.Request.Context()
	role, err := h.memberRole(ctx, orgID, callerID)
	if err != nil || !canManage(role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	var userIDs []string
	if err := h.DB.NewSelect().TableExpr("org_group_keys").
		ColumnExpr("user_id").
		Where("group_id = ?", groupID).
		Scan(ctx, &userIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch provisioned members"})
		return
	}
	if userIDs == nil {
		userIDs = []string{}
	}
	c.JSON(http.StatusOK, gin.H{"provisioned_user_ids": userIDs})
}

// GetGroupFileKeys returns all file IDs + encrypted keys for files in a group.
// Used by the client when rotating the group key.
func (h *OrgHandler) GetGroupFileKeys(c *gin.Context) {
	callerID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}
	groupID, err := strconv.ParseInt(c.Param("groupID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}

	ctx := c.Request.Context()
	role, err := h.memberRole(ctx, orgID, callerID)
	if err != nil || !canManage(role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	type fileKeyRow struct {
		ID           int64  `bun:"id" json:"id"`
		EncryptedKey string `bun:"encrypted_key" json:"encrypted_key"`
	}
	var rows []fileKeyRow
	if err := h.DB.NewSelect().TableExpr("org_files").
		ColumnExpr("id, encrypted_key").
		Where("org_id = ? AND group_id = ? AND deleted_at IS NULL", orgID, groupID).
		Scan(ctx, &rows); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch file keys"})
		return
	}
	if rows == nil {
		rows = []fileKeyRow{}
	}
	c.JSON(http.StatusOK, rows)
}
