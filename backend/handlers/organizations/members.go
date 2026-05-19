// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package organizations

import (
	"net/http"
	"strconv"
	"time"

	"kagibi/backend/pkg"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type MemberResponse struct {
	ID              int64     `json:"id"`
	UserID          string    `json:"user_id"`
	Name            string    `json:"name"`
	Email           string    `json:"email,omitempty"`
	AvatarURL       string    `json:"avatar_url"`
	PublicKey       string    `json:"public_key,omitempty"`
	Role            string    `json:"role"`
	EncryptedOrgKey string    `json:"encrypted_org_key,omitempty"`
	QuotaBytes      int64     `json:"quota_bytes"`
	JoinedAt        time.Time `json:"joined_at"`
}

func (h *OrgHandler) ListMembers(c *gin.Context) {
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

	var members []pkg.OrgMember
	if err := h.DB.NewSelect().Model(&members).
		Where("org_id = ?", orgID).
		OrderExpr("joined_at ASC").
		Scan(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list members"})
		return
	}

	userIDs := make([]string, len(members))
	for i, m := range members {
		userIDs[i] = m.UserID
	}
	var users []pkg.User
	if err := h.DB.NewSelect().Model(&users).
		Where("id IN (?)", bun.In(userIDs)).
		Scan(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch member profiles"})
		return
	}
	// Decrypt emails before indexing so the map holds plaintext values.
	for i := range users {
		_ = pkg.DecryptUserEmail(&users[i])
	}
	userByID := make(map[string]pkg.User, len(users))
	for _, u := range users {
		userByID[u.ID] = u
	}

	isAdmin := canManage(role)
	result := make([]MemberResponse, len(members))
	for i, m := range members {
		u := userByID[m.UserID]
		res := MemberResponse{
			ID:         m.ID,
			UserID:     m.UserID,
			Name:       u.Name,
			Email:      u.Email,
			AvatarURL:  u.AvatarURL,
			Role:       m.Role,
			QuotaBytes: m.QuotaBytes,
			JoinedAt:   m.JoinedAt,
		}
		// Key material and raw public keys are only exposed to admins/owners.
		// Regular members and viewers have no legitimate use for this data.
		if isAdmin {
			res.PublicKey = u.PublicKey
			res.EncryptedOrgKey = m.EncryptedOrgKey
		}
		result[i] = res
	}
	c.JSON(http.StatusOK, result)
}

func (h *OrgHandler) UpdateMemberRole(c *gin.Context) {
	callerID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}
	memberID, err := strconv.ParseInt(c.Param("memberID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid member id"})
		return
	}

	var req struct {
		Role       string `json:"role" binding:"required"`
		QuotaBytes *int64 `json:"quota_bytes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Role == "owner" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ownership transfer is not supported via this endpoint"})
		return
	}
	validRoles := map[string]bool{"admin": true, "member": true, "viewer": true}
	if !validRoles[req.Role] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role: must be admin, member, or viewer"})
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

	var target pkg.OrgMember
	if err := h.DB.NewSelect().Model(&target).
		Where("id = ? AND org_id = ?", memberID, orgID).
		Scan(ctx); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "member not found"})
		return
	}
	if target.Role == "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot change the owner's role"})
		return
	}
	if callerRole == "admin" && target.Role == "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admins cannot change other admins' roles"})
		return
	}

	previousRole := target.Role
	target.Role = req.Role
	columns := []string{"role"}
	if req.QuotaBytes != nil {
		target.QuotaBytes = *req.QuotaBytes
		columns = append(columns, "quota_bytes")
	}
	if _, err := h.DB.NewUpdate().Model(&target).WherePK().Column(columns...).Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update member"})
		return
	}
	h.logAudit(ctx, orgID, callerID, "role_changed", target.UserID, "user",
		previousRole+" → "+req.Role)
	c.JSON(http.StatusOK, gin.H{"id": target.ID, "role": target.Role, "quota_bytes": target.QuotaBytes})
}

func (h *OrgHandler) RemoveMember(c *gin.Context) {
	callerID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}
	memberID, err := strconv.ParseInt(c.Param("memberID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid member id"})
		return
	}
	ctx := c.Request.Context()

	callerRole, err := h.memberRole(ctx, orgID, callerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check membership"})
		return
	}
	if callerRole == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member of this organization"})
		return
	}

	var target pkg.OrgMember
	if err := h.DB.NewSelect().Model(&target).
		Where("id = ? AND org_id = ?", memberID, orgID).
		Scan(ctx); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "member not found"})
		return
	}
	if target.Role == "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot remove the organization owner"})
		return
	}

	isSelf := target.UserID == callerID
	if !isSelf && !canManage(callerRole) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}
	if !isSelf && callerRole == "admin" && target.Role == "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admins cannot remove other admins"})
		return
	}

	if _, err := h.DB.NewDelete().Model(&target).WherePK().Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove member"})
		return
	}
	detail := "self-removal"
	if !isSelf {
		detail = "removed by " + callerID
	}
	h.logAudit(ctx, orgID, callerID, "member_removed", target.UserID, "user", detail)
	c.JSON(http.StatusOK, gin.H{"message": "member removed"})
}

// SetMemberKey lets an admin provide the org key encrypted with a specific member's public key.
// This is called after a member joins via link invite (where no pre-encrypted key exists).
func (h *OrgHandler) SetMemberKey(c *gin.Context) {
	callerID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}
	memberID, err := strconv.ParseInt(c.Param("memberID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid member id"})
		return
	}

	var req struct {
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
	if !canManage(callerRole) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	var target pkg.OrgMember
	if err := h.DB.NewSelect().Model(&target).
		Where("id = ? AND org_id = ?", memberID, orgID).
		Scan(ctx); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "member not found"})
		return
	}
	if _, err = h.DB.NewUpdate().Model((*pkg.OrgMember)(nil)).
		Set("encrypted_org_key = ?", req.EncryptedOrgKey).
		Where("id = ? AND org_id = ?", memberID, orgID).
		Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set member key"})
		return
	}
	h.logAudit(ctx, orgID, callerID, "key_provisioned", target.UserID, "user", "")
	c.JSON(http.StatusOK, gin.H{"message": "key updated"})
}
