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

// ── Groups ────────────────────────────────────────────────────────────────────

// ListGroups returns all groups for the organization. Visible to all members.
func (h *OrgHandler) ListGroups(c *gin.Context) {
	callerID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}
	ctx := c.Request.Context()

	role, err := h.memberRole(ctx, orgID, callerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check membership"})
		return
	}
	if role == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member of this organization"})
		return
	}

	var groups []pkg.OrgGroup
	if err := h.DB.NewSelect().Model(&groups).
		Where("org_id = ?", orgID).
		OrderExpr("name ASC").
		Scan(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list groups"})
		return
	}
	c.JSON(http.StatusOK, groups)
}

// ListMyGroups returns the groups the current user belongs to, including their role in each group.
func (h *OrgHandler) ListMyGroups(c *gin.Context) {
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
	if role == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member of this organization"})
		return
	}

	type MyGroupEntry struct {
		ID          int64  `bun:"id"          json:"id"`
		OrgID       int64  `bun:"org_id"      json:"org_id"`
		Name        string `bun:"name"        json:"name"`
		Description string `bun:"description" json:"description,omitempty"`
		Source      string `bun:"source"      json:"source,omitempty"`
		MyRole      string `bun:"my_role"     json:"my_role"`
	}

	var groups []MyGroupEntry
	if err := h.DB.NewSelect().
		TableExpr("org_groups og").
		ColumnExpr("og.id, og.org_id, og.name, og.description, og.source, ogm.role AS my_role").
		Join("JOIN org_group_members ogm ON ogm.group_id = og.id").
		Where("og.org_id = ? AND ogm.user_id = ?", orgID, userID).
		OrderExpr("og.name ASC").
		Scan(ctx, &groups); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list groups"})
		return
	}
	if groups == nil {
		groups = []MyGroupEntry{}
	}
	c.JSON(http.StatusOK, groups)
}

// GetGroup returns a single group with its member list.
func (h *OrgHandler) GetGroup(c *gin.Context) {
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
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check membership"})
		return
	}
	if role == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member of this organization"})
		return
	}

	var group pkg.OrgGroup
	if err := h.DB.NewSelect().Model(&group).
		Where("id = ? AND org_id = ?", groupID, orgID).
		Scan(ctx); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch group"})
		return
	}

	var members []pkg.OrgGroupMember
	if err := h.DB.NewSelect().Model(&members).
		Where("group_id = ?", groupID).
		OrderExpr("joined_at ASC").
		Scan(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch group members"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"group": group, "members": members})
}

// CreateGroup creates a new group within an organization. Requires admin or owner.
func (h *OrgHandler) CreateGroup(c *gin.Context) {
	callerID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
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

	group := &pkg.OrgGroup{
		OrgID:       orgID,
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   callerID,
		Source:      "internal",
	}
	if _, err := h.DB.NewInsert().Model(group).Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create group"})
		return
	}

	h.logAudit(ctx, orgID, callerID, "group_created", strconv.FormatInt(group.ID, 10), "group", req.Name)
	c.JSON(http.StatusCreated, group)
}

// UpdateGroup renames or updates a group's description. Requires admin or owner.
func (h *OrgHandler) UpdateGroup(c *gin.Context) {
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
		Name        *string `json:"name"`
		Description *string `json:"description"`
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

	var group pkg.OrgGroup
	if err := h.DB.NewSelect().Model(&group).
		Where("id = ? AND org_id = ?", groupID, orgID).
		Scan(ctx); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch group"})
		return
	}
	if group.Source == "ldap" {
		c.JSON(http.StatusForbidden, gin.H{"error": "LDAP-synced groups cannot be edited manually"})
		return
	}

	if req.Name != nil {
		group.Name = *req.Name
	}
	if req.Description != nil {
		group.Description = *req.Description
	}
	group.UpdatedAt = time.Now()

	if _, err := h.DB.NewUpdate().Model(&group).
		Column("name", "description", "updated_at").
		Where("id = ?", groupID).
		Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update group"})
		return
	}

	h.logAudit(ctx, orgID, callerID, "group_updated", strconv.FormatInt(groupID, 10), "group", group.Name)
	c.JSON(http.StatusOK, group)
}

// DeleteGroup deletes a group and all its memberships and permission overrides.
//
// If the group has encrypted files (group_id set on org_files), the caller must supply
// the file keys re-wrapped with the OrgKey in the request body, otherwise the files would
// become permanently undecryptable (data loss).  The endpoint returns 409 with the file
// count when this re-wrap payload is missing.
func (h *OrgHandler) DeleteGroup(c *gin.Context) {
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

	callerRole, err := h.memberRole(ctx, orgID, callerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check membership"})
		return
	}
	if !canManage(callerRole) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	var group pkg.OrgGroup
	if err := h.DB.NewSelect().Model(&group).
		Where("id = ? AND org_id = ?", groupID, orgID).
		Scan(ctx); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch group"})
		return
	}

	// Optional body: re-wrapped file keys (required when the group has encrypted files).
	var req struct {
		FileKeys []struct {
			FileID       int64  `json:"file_id"`
			EncryptedKey string `json:"encrypted_key"`
		} `json:"file_keys"`
	}
	// ShouldBindJSON may fail on a body-less DELETE — that is acceptable.
	_ = c.ShouldBindJSON(&req)

	// Count non-deleted files that are still encrypted with this group's key.
	var groupFileCount int
	if err := h.DB.NewSelect().TableExpr("org_files").
		ColumnExpr("COUNT(*)").
		Where("group_id = ? AND deleted_at IS NULL", groupID).
		Scan(ctx, &groupFileCount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count group files"})
		return
	}
	if groupFileCount > 0 && len(req.FileKeys) == 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error":      "group has encrypted files; re-wrap their keys with the org key before deleting",
			"file_count": groupFileCount,
		})
		return
	}

	// Cascade: remove members, permission overrides and the group itself in a single
	// transaction so a partial failure never leaves orphaned rows behind.
	tx, err := h.DB.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start transaction"})
		return
	}
	defer tx.Rollback()

	// Re-wrap file keys with the OrgKey before unlinking them from the group,
	// so they remain decryptable after the GroupKey is gone.
	for _, fk := range req.FileKeys {
		if _, err := tx.NewUpdate().Model((*pkg.OrgFile)(nil)).
			Set("encrypted_key = ?", fk.EncryptedKey).
			Where("id = ? AND org_id = ? AND group_id = ?", fk.FileID, orgID, groupID).
			Exec(ctx); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update file key"})
			return
		}
	}

	if _, err := tx.NewDelete().Model((*pkg.OrgGroupMember)(nil)).
		Where("group_id = ?", groupID).Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete group members"})
		return
	}
	if _, err := tx.NewDelete().Model((*pkg.OrgGroupPermission)(nil)).
		Where("group_id = ?", groupID).Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete group permissions"})
		return
	}
	// org_group_keys has ON DELETE CASCADE from the group_id FK, so they are removed automatically.
	// Nullify group_id on files/folders — keys are now wrapped with OrgKey again.
	if _, err := tx.NewUpdate().Model((*pkg.OrgFile)(nil)).
		Set("group_id = NULL").
		Where("group_id = ?", groupID).
		Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unlink group files"})
		return
	}
	if _, err := tx.NewUpdate().Model((*pkg.OrgFolder)(nil)).
		Set("group_id = NULL").
		Where("group_id = ?", groupID).
		Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unlink group folders"})
		return
	}
	if _, err := tx.NewDelete().Model(&group).WherePK().Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete group"})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit deletion"})
		return
	}

	h.logAudit(ctx, orgID, callerID, "group_deleted", strconv.FormatInt(groupID, 10), "group", group.Name)
	c.JSON(http.StatusOK, gin.H{"message": "group deleted"})
}

// ── Group members ─────────────────────────────────────────────────────────────

// ListGroupMembers returns the members of a group.
func (h *OrgHandler) ListGroupMembers(c *gin.Context) {
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
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check membership"})
		return
	}
	if role == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member of this organization"})
		return
	}

	var members []pkg.OrgGroupMember
	if err := h.DB.NewSelect().Model(&members).
		Where("group_id = ?", groupID).
		OrderExpr("joined_at ASC").
		Scan(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list group members"})
		return
	}
	c.JSON(http.StatusOK, members)
}

// AddGroupMember adds an org member to a group.
// Requires org admin/owner, or group admin (who can only add with role "member").
func (h *OrgHandler) AddGroupMember(c *gin.Context) {
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
		UserID string `json:"user_id" binding:"required"`
		Role   string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Role == "" {
		req.Role = "member"
	}
	if req.Role != "admin" && req.Role != "member" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role must be admin or member"})
		return
	}

	ctx := c.Request.Context()
	callerOrgRole, err := h.memberRole(ctx, orgID, callerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check membership"})
		return
	}
	callerGroupRole, err := h.groupRole(ctx, groupID, callerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check group membership"})
		return
	}

	isOrgAdmin := canManage(callerOrgRole)
	isGroupAdmin := callerGroupRole == "admin"
	if !isOrgAdmin && !isGroupAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}
	// Group admins can only add members, not promote to group admin.
	if !isOrgAdmin && isGroupAdmin && req.Role == "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "group admins cannot grant admin role"})
		return
	}

	targetRole, err := h.memberRole(ctx, orgID, req.UserID)
	if err != nil || targetRole == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "target user is not a member of this organization"})
		return
	}

	var group pkg.OrgGroup
	if err := h.DB.NewSelect().Model(&group).
		Where("id = ? AND org_id = ?", groupID, orgID).
		Scan(ctx); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch group"})
		return
	}

	gm := &pkg.OrgGroupMember{
		GroupID: groupID,
		UserID:  req.UserID,
		Role:    req.Role,
		AddedBy: callerID,
	}
	if _, err := h.DB.NewInsert().Model(gm).
		On("CONFLICT (group_id, user_id) DO UPDATE SET role = EXCLUDED.role, added_by = EXCLUDED.added_by").
		Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add group member"})
		return
	}

	h.logAudit(ctx, orgID, callerID, "group_member_added", req.UserID, "group", group.Name)
	c.JSON(http.StatusCreated, gm)
}

// RemoveGroupMember removes a user from a group.
// Requires org admin/owner, or group admin (who cannot remove other group admins).
func (h *OrgHandler) RemoveGroupMember(c *gin.Context) {
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
	memberID, err := strconv.ParseInt(c.Param("memberID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid member id"})
		return
	}

	ctx := c.Request.Context()
	callerOrgRole, err := h.memberRole(ctx, orgID, callerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check membership"})
		return
	}
	callerGroupRole, err := h.groupRole(ctx, groupID, callerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check group membership"})
		return
	}
	if !canManage(callerOrgRole) && callerGroupRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	var gm pkg.OrgGroupMember
	if err := h.DB.NewSelect().Model(&gm).
		Where("id = ? AND group_id = ?", memberID, groupID).
		Scan(ctx); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "group member not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch group member"})
		return
	}

	// Group admins cannot remove other group admins — only org admins/owners can.
	if !canManage(callerOrgRole) && gm.Role == "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "group admins cannot remove other group admins"})
		return
	}

	var group pkg.OrgGroup
	h.DB.NewSelect().Model(&group).Where("id = ? AND org_id = ?", groupID, orgID).Scan(ctx)

	// Remove the member and revoke their group key in a single transaction so
	// a DB failure cannot leave the key entry behind (silent access leak — B5 fix).
	tx, err := h.DB.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start transaction"})
		return
	}
	defer tx.Rollback()

	if _, err := tx.NewDelete().Model(&gm).WherePK().Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove group member"})
		return
	}
	if _, err := tx.NewDelete().Model((*pkg.OrgGroupKey)(nil)).
		Where("group_id = ? AND user_id = ?", groupID, gm.UserID).
		Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke group key"})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit"})
		return
	}

	h.logAudit(ctx, orgID, callerID, "group_member_removed", gm.UserID, "group", group.Name)
	c.JSON(http.StatusOK, gin.H{"message": "member removed from group"})
}

// SetGroupMemberRole changes the role of a group member (admin | member).
// Org admin/owner can set any role. Group admin can only set "member".
func (h *OrgHandler) SetGroupMemberRole(c *gin.Context) {
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
	memberID, err := strconv.ParseInt(c.Param("memberID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid member id"})
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Role != "admin" && req.Role != "member" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role must be admin or member"})
		return
	}

	ctx := c.Request.Context()
	callerOrgRole, err := h.memberRole(ctx, orgID, callerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check membership"})
		return
	}
	callerGroupRole, err := h.groupRole(ctx, groupID, callerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check group membership"})
		return
	}
	isOrgAdmin := canManage(callerOrgRole)
	if !isOrgAdmin && callerGroupRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}
	if !isOrgAdmin && req.Role == "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "group admins cannot grant admin role"})
		return
	}

	var gm pkg.OrgGroupMember
	if err := h.DB.NewSelect().Model(&gm).
		Where("id = ? AND group_id = ?", memberID, groupID).
		Scan(ctx); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "group member not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch group member"})
		return
	}

	gm.Role = req.Role
	if _, err := h.DB.NewUpdate().Model(&gm).Column("role").WherePK().Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update group member role"})
		return
	}

	c.JSON(http.StatusOK, gm)
}

// ── Group permissions ─────────────────────────────────────────────────────────

// ListGroupPermissions returns all permission overrides for a group.
func (h *OrgHandler) ListGroupPermissions(c *gin.Context) {
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

	callerRole, err := h.memberRole(ctx, orgID, callerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check membership"})
		return
	}
	if !canManage(callerRole) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	var perms []pkg.OrgGroupPermission
	if err := h.DB.NewSelect().Model(&perms).
		Where("org_id = ? AND group_id = ?", orgID, groupID).
		OrderExpr("folder_path ASC").
		Scan(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list group permissions"})
		return
	}
	c.JSON(http.StatusOK, perms)
}

// SetGroupPermission creates or replaces a folder-level permission for a group.
func (h *OrgHandler) SetGroupPermission(c *gin.Context) {
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
		FolderPath       string `json:"folder_path"`
		Level            string `json:"level" binding:"required"`
		PermCreate       bool   `json:"perm_create"`
		PermDelete       bool   `json:"perm_delete"`
		PermDownload     *bool  `json:"perm_download"`
		PermMove         bool   `json:"perm_move"`
		RestrictToGroups bool   `json:"restrict_to_groups"`
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
	caps, err := h.resolveCallerCaps(ctx, orgID, callerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check permissions"})
		return
	}
	if caps.OrgRole == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member of this organization"})
		return
	}

	folderPath := normPath(req.FolderPath)

	if !caps.IsOrgAdmin() {
		if !caps.AdminGroupIDs[groupID] {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}
		// Group admins can only modify the level of an existing permission — not grant new folder access.
		var existing int
		if err := h.DB.NewSelect().TableExpr("org_group_permissions").
			ColumnExpr("COUNT(*)").
			Where("org_id = ? AND group_id = ? AND folder_path = ?", orgID, groupID, folderPath).
			Scan(ctx, &existing); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check existing permission"})
			return
		}
		if existing == 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "group admins cannot grant new folder access; contact an org admin"})
			return
		}
	}

	// Group must belong to this org.
	var count int
	if err := h.DB.NewSelect().TableExpr("org_groups").
		ColumnExpr("COUNT(*)").
		Where("id = ? AND org_id = ?", groupID, orgID).
		Scan(ctx, &count); err != nil || count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		return
	}

	permDownload := true
	if req.PermDownload != nil {
		permDownload = *req.PermDownload
	}

	switch req.Level {
	case "manage":
		req.PermCreate, req.PermDelete, req.PermMove = true, true, true
	case "write":
		req.PermCreate, req.PermMove = true, true
	}

	perm := &pkg.OrgGroupPermission{
		OrgID:            orgID,
		GroupID:          groupID,
		FolderPath:       folderPath,
		Level:            req.Level,
		PermCreate:       req.PermCreate,
		PermDelete:       req.PermDelete,
		PermDownload:     permDownload,
		PermMove:         req.PermMove,
		RestrictToGroups: req.RestrictToGroups,
	}

	if _, err := h.DB.NewInsert().Model(perm).
		On("CONFLICT (org_id, group_id, folder_path) DO UPDATE").
		Set("level = EXCLUDED.level, perm_create = EXCLUDED.perm_create, perm_delete = EXCLUDED.perm_delete, perm_download = EXCLUDED.perm_download, perm_move = EXCLUDED.perm_move, restrict_to_groups = EXCLUDED.restrict_to_groups").
		Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set group permission"})
		return
	}

	h.logAudit(ctx, orgID, callerID, "group_permission_set",
		strconv.FormatInt(groupID, 10), "group", req.Level+" on "+folderPath)
	c.JSON(http.StatusOK, perm)
}

// DeleteGroupPermission removes a folder-level permission override from a group.
func (h *OrgHandler) DeleteGroupPermission(c *gin.Context) {
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
		FolderPath string `json:"folder_path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	caps, err := h.resolveCallerCaps(ctx, orgID, callerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check permissions"})
		return
	}
	if caps.OrgRole == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member of this organization"})
		return
	}
	if !caps.IsOrgAdmin() && !caps.AdminGroupIDs[groupID] {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	folderPath := normPath(req.FolderPath)
	res, err := h.DB.NewDelete().Model((*pkg.OrgGroupPermission)(nil)).
		Where("org_id = ? AND group_id = ? AND folder_path = ?", orgID, groupID, folderPath).
		Exec(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete group permission"})
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "permission override not found"})
		return
	}

	h.logAudit(ctx, orgID, callerID, "group_permission_removed",
		strconv.FormatInt(groupID, 10), "group", folderPath)
	c.JSON(http.StatusOK, gin.H{"message": "group permission removed"})
}
