// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package organizations

import (
	"net/http"
	"strconv"

	"kagibi/backend/pkg"

	"github.com/gin-gonic/gin"
)

// GetOrgActivity returns the 50 most recent audit entries for any org member.
// Unlike ListAuditLog (admin-only, paginated), this is a lightweight feed for all members.
func (h *OrgHandler) GetOrgActivity(c *gin.Context) {
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid org id"})
		return
	}
	userID := c.GetString("user_id")
	role, err := h.memberRole(c.Request.Context(), orgID, userID)
	if err != nil || role == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member"})
		return
	}

	var entries []pkg.OrgAuditLog
	if err := h.DB.NewSelect().Model(&entries).
		Where("oal.org_id = ?", orgID).
		OrderExpr("oal.created_at DESC").
		Limit(50).
		Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if entries == nil {
		entries = []pkg.OrgAuditLog{}
	}
	c.JSON(http.StatusOK, entries)
}
