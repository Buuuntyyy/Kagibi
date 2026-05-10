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
	AvatarURL       string    `json:"avatar_url"`
	Role            string    `json:"role"`
	EncryptedOrgKey string    `json:"encrypted_org_key,omitempty"`
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
	userByID := make(map[string]pkg.User, len(users))
	for _, u := range users {
		userByID[u.ID] = u
	}

	result := make([]MemberResponse, len(members))
	for i, m := range members {
		u := userByID[m.UserID]
		result[i] = MemberResponse{
			ID:              m.ID,
			UserID:          m.UserID,
			Name:            u.Name,
			AvatarURL:       u.AvatarURL,
			Role:            m.Role,
			EncryptedOrgKey: m.EncryptedOrgKey,
			JoinedAt:        m.JoinedAt,
		}
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
		Role string `json:"role" binding:"required"`
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

	target.Role = req.Role
	if _, err := h.DB.NewUpdate().Model(&target).WherePK().Column("role").Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update role"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": target.ID, "role": target.Role})
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

	_, err = h.DB.NewUpdate().Model((*pkg.OrgMember)(nil)).
		Set("encrypted_org_key = ?", req.EncryptedOrgKey).
		Where("id = ? AND org_id = ?", memberID, orgID).
		Exec(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set member key"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "key updated"})
}
