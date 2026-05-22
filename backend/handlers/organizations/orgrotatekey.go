// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package organizations

import (
	"fmt"
	"net/http"
	"strconv"

	"kagibi/backend/pkg"

	"github.com/gin-gonic/gin"
)

type rotateKeyRequest struct {
	MemberKeys []struct {
		MemberID        int64  `json:"member_id"`
		EncryptedOrgKey string `json:"encrypted_org_key"`
	} `json:"member_keys" binding:"required"`
	FileKeys []struct {
		FileID       int64  `json:"file_id"`
		EncryptedKey string `json:"encrypted_key"`
	} `json:"file_keys" binding:"required"`
}

// RotateOrgKey atomically replaces the OrgKey for every member and re-wraps
// every file key with the new OrgKey. Only the owner may call this.
//
// All cryptographic operations happen on the client — this endpoint receives
// already-encrypted values and swaps them in a single DB transaction.
func (h *OrgHandler) RotateOrgKey(c *gin.Context) {
	callerID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}

	var req rotateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !h.checkMFAEnforcement(c, orgID, callerID) {
		return
	}

	ctx := c.Request.Context()

	callerRole, err := h.memberRole(ctx, orgID, callerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check membership"})
		return
	}
	if !isOwner(callerRole) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the organization owner can rotate the key"})
		return
	}

	tx, err := h.DB.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to begin transaction"})
		return
	}
	defer tx.Rollback()

	for _, mk := range req.MemberKeys {
		if _, err := tx.NewUpdate().Model((*pkg.OrgMember)(nil)).
			Set("encrypted_org_key = ?", mk.EncryptedOrgKey).
			Where("id = ? AND org_id = ?", mk.MemberID, orgID).
			Exec(ctx); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update member key"})
			return
		}
	}

	for _, fk := range req.FileKeys {
		if _, err := tx.NewUpdate().Model((*pkg.OrgFile)(nil)).
			Set("encrypted_key = ?", fk.EncryptedKey).
			Where("id = ? AND org_id = ?", fk.FileID, orgID).
			Exec(ctx); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update file key"})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit key rotation"})
		return
	}

	h.logAudit(ctx, orgID, callerID, "key_rotated", "", "org",
		fmt.Sprintf("%d member keys and %d file keys re-encrypted", len(req.MemberKeys), len(req.FileKeys)))

	c.JSON(http.StatusOK, gin.H{
		"message":     "org key rotated",
		"member_keys": len(req.MemberKeys),
		"file_keys":   len(req.FileKeys),
	})
}
