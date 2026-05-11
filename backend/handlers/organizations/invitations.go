// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package organizations

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"

	"kagibi/backend/pkg"

	"github.com/gin-gonic/gin"
)

func generateInviteToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// InvitationInfoResponse is returned to a user who fetches an invite by token.
// It reveals only what is needed to show a join confirmation screen.
type InvitationInfoResponse struct {
	OrgID   int64  `json:"org_id"`
	OrgName string `json:"org_name"`
	Role    string `json:"role"`
	Token   string `json:"token"`
}

func (h *OrgHandler) CreateInvitation(c *gin.Context) {
	callerID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}

	var req struct {
		Role            string     `json:"role"`
		MaxUses         int        `json:"max_uses"`
		ExpiresAt       *time.Time `json:"expires_at"`
		TargetUserID    *string    `json:"target_user_id"`    // nil = link invite, set = direct invite
		EncryptedOrgKey string     `json:"encrypted_org_key"` // pre-encrypted for direct invites
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Role == "" {
		req.Role = "member"
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
	if req.Role == "admin" && !isOwner(callerRole) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only owners can invite admins"})
		return
	}

	if req.TargetUserID != nil {
		var targetUser pkg.User
		if err := h.DB.NewSelect().Model(&targetUser).Where("id = ?", *req.TargetUserID).Scan(ctx); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "target user not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to look up target user"})
			return
		}
		existing, _ := h.memberRole(ctx, orgID, *req.TargetUserID)
		if existing != "" {
			c.JSON(http.StatusConflict, gin.H{"error": "user is already a member"})
			return
		}
	}

	token, err := generateInviteToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate invitation token"})
		return
	}

	inv := &pkg.OrgInvitation{
		OrgID:           orgID,
		InvitedBy:       callerID,
		Token:           token,
		TargetUserID:    req.TargetUserID,
		EncryptedOrgKey: req.EncryptedOrgKey,
		Role:            req.Role,
		MaxUses:         req.MaxUses,
		ExpiresAt:       req.ExpiresAt,
		Status:          "active",
	}
	if _, err := h.DB.NewInsert().Model(inv).Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create invitation"})
		return
	}
	targetDesc := "link invite"
	if req.TargetUserID != nil {
		targetDesc = "direct invite for " + *req.TargetUserID
	}
	h.logAudit(ctx, orgID, callerID, "invitation_created", strconv.FormatInt(inv.ID, 10), "invitation",
		req.Role+" / "+targetDesc)
	c.JSON(http.StatusCreated, inv)
}

func (h *OrgHandler) ListInvitations(c *gin.Context) {
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

	var invitations []pkg.OrgInvitation
	if err := h.DB.NewSelect().Model(&invitations).
		Where("org_id = ? AND status = 'active'", orgID).
		OrderExpr("created_at DESC").
		Scan(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list invitations"})
		return
	}
	c.JSON(http.StatusOK, invitations)
}

func (h *OrgHandler) RevokeInvitation(c *gin.Context) {
	callerID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}
	invID, err := strconv.ParseInt(c.Param("invID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invitation id"})
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

	res, err := h.DB.NewUpdate().Model((*pkg.OrgInvitation)(nil)).
		Set("status = ?", "revoked").
		Where("id = ? AND org_id = ? AND status = 'active'", invID, orgID).
		Exec(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke invitation"})
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "invitation not found or already revoked"})
		return
	}
	h.logAudit(ctx, orgID, callerID, "invitation_revoked", strconv.FormatInt(invID, 10), "invitation", "")
	c.JSON(http.StatusOK, gin.H{"message": "invitation revoked"})
}

// GetInvitation returns org info for a token — used to show a join confirmation screen.
func (h *OrgHandler) GetInvitation(c *gin.Context) {
	token := c.Param("token")
	ctx := c.Request.Context()

	var inv pkg.OrgInvitation
	if err := h.DB.NewSelect().Model(&inv).Where("token = ?", token).Scan(ctx); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "invitation not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch invitation"})
		return
	}
	if err := validateInvitation(inv); err != "" {
		c.JSON(http.StatusGone, gin.H{"error": err})
		return
	}

	var org pkg.Organization
	_ = h.DB.NewSelect().Model(&org).Where("id = ?", inv.OrgID).Scan(ctx)

	c.JSON(http.StatusOK, InvitationInfoResponse{
		OrgID:   inv.OrgID,
		OrgName: org.Name,
		Role:    inv.Role,
		Token:   inv.Token,
	})
}

// AcceptInvitation adds the authenticated caller as a member of the org.
func (h *OrgHandler) AcceptInvitation(c *gin.Context) {
	userID := c.GetString("user_id")
	token := c.Param("token")

	var req struct {
		EncryptedOrgKey string `json:"encrypted_org_key"` // provided by user for link invites
	}
	_ = c.ShouldBindJSON(&req) // body is optional

	ctx := c.Request.Context()

	var inv pkg.OrgInvitation
	if err := h.DB.NewSelect().Model(&inv).Where("token = ?", token).Scan(ctx); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "invitation not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch invitation"})
		return
	}
	if errMsg := validateInvitation(inv); errMsg != "" {
		c.JSON(http.StatusGone, gin.H{"error": errMsg})
		return
	}
	if inv.TargetUserID != nil && *inv.TargetUserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "this invitation is for another user"})
		return
	}

	existingRole, err := h.memberRole(ctx, inv.OrgID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check membership"})
		return
	}
	if existingRole != "" {
		c.JSON(http.StatusConflict, gin.H{"error": "already a member of this organization"})
		return
	}

	// For direct invites the key is pre-encrypted; for link invites the user may provide it.
	encryptedKey := req.EncryptedOrgKey
	if encryptedKey == "" {
		encryptedKey = inv.EncryptedOrgKey
	}

	member := &pkg.OrgMember{
		OrgID:           inv.OrgID,
		UserID:          userID,
		Role:            inv.Role,
		EncryptedOrgKey: encryptedKey,
	}
	if _, err := h.DB.NewInsert().Model(member).Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to join organization"})
		return
	}

	// When an admin-provisioned org is claimed, promote the accepting user to owner_id.
	if inv.Role == "owner" {
		_, _ = h.DB.NewUpdate().Model((*pkg.Organization)(nil)).
			Set("owner_id = ?", userID).
			Where("id = ? AND owner_id = 'pending'", inv.OrgID).
			Exec(ctx)
	}

	_, _ = h.DB.NewUpdate().Model((*pkg.OrgInvitation)(nil)).
		Set("uses = uses + 1").
		Where("id = ?", inv.ID).
		Exec(ctx)

	// Revoke if the invite is now exhausted or was direct (single-target)
	if inv.TargetUserID != nil || (inv.MaxUses > 0 && inv.Uses+1 >= inv.MaxUses) {
		_, _ = h.DB.NewUpdate().Model((*pkg.OrgInvitation)(nil)).
			Set("status = ?", "revoked").
			Where("id = ?", inv.ID).
			Exec(ctx)
	}

	h.logAudit(ctx, inv.OrgID, userID, "member_joined", userID, "user", inv.Role)

	// Notify all existing org members so they can refresh the member list in real time.
	var existingMembers []pkg.OrgMember
	if err := h.DB.NewSelect().Model(&existingMembers).
		Column("user_id").
		Where("org_id = ? AND user_id != ?", inv.OrgID, userID).
		Scan(ctx); err == nil {
		payload := map[string]any{"org_id": inv.OrgID, "action": "member_joined", "user_id": userID}
		for _, m := range existingMembers {
			_ = pkg.EmitRealtimeEvent(ctx, h.DB, m.UserID, "org_update", payload)
		}
	}

	c.JSON(http.StatusCreated, gin.H{"org_id": inv.OrgID, "role": inv.Role})
}

func validateInvitation(inv pkg.OrgInvitation) string {
	if inv.Status != "active" {
		return "invitation has been revoked"
	}
	if inv.ExpiresAt != nil && time.Now().After(*inv.ExpiresAt) {
		return "invitation has expired"
	}
	if inv.MaxUses > 0 && inv.Uses >= inv.MaxUses {
		return "invitation has reached its usage limit"
	}
	return ""
}
