// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package organizations

import (
	"net/http"

	"kagibi/backend/pkg"

	"github.com/gin-gonic/gin"
)

func (h *OrgHandler) CreateOrg(c *gin.Context) {
	userID := c.GetString("user_id")
	if !h.hasOrgAccess(userID) {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "organizations require a paid plan on the cloud"})
		return
	}
	var req struct {
		Name            string `json:"name" binding:"required"`
		Description     string `json:"description"`
		EncryptedOrgKey string `json:"encrypted_org_key"`
		StorageQuotaMB  int64  `json:"storage_quota_mb"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.StorageQuotaMB <= 0 {
		req.StorageQuotaMB = 10240
	}

	ctx := c.Request.Context()
	org := &pkg.Organization{
		Name:           req.Name,
		Description:    req.Description,
		OwnerID:        userID,
		StorageQuotaMB: req.StorageQuotaMB,
	}
	if _, err := h.DB.NewInsert().Model(org).Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create organization"})
		return
	}

	member := &pkg.OrgMember{
		OrgID:           org.ID,
		UserID:          userID,
		Role:            "owner",
		EncryptedOrgKey: req.EncryptedOrgKey,
	}
	if _, err := h.DB.NewInsert().Model(member).Exec(ctx); err != nil {
		h.DB.NewDelete().Model(org).WherePK().Exec(ctx) //nolint:errcheck
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initialize membership"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"organization": org, "my_role": "owner"})
}
