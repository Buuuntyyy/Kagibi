// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package organizations

import (
	"fmt"
	"net/http"
	"strconv"

	"kagibi/backend/pkg"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// ListAuditLog returns paginated audit events for an org (last 1 year only).
// Restricted to admins and owners.
func (h *OrgHandler) ListAuditLog(c *gin.Context) {
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

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	const pageSize = 50
	offset := (page - 1) * pageSize

	var entries []pkg.OrgAuditLog
	if err := h.DB.NewSelect().Model(&entries).
		Where("org_id = ? AND created_at >= NOW() - INTERVAL '1 year'", orgID).
		OrderExpr("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Scan(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch audit log"})
		return
	}

	c.JSON(http.StatusOK, entries)
}

// AuditSummary returns the count of audit entries per day for the past year.
// Used by the frontend to build the calendar-based delete UI.
func (h *OrgHandler) AuditSummary(c *gin.Context) {
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

	type row struct {
		Day   string `bun:"day"`
		Count int    `bun:"count"`
	}
	var rows []row
	if err := h.DB.NewRaw(`
		SELECT to_char(created_at, 'YYYY-MM-DD') AS day, COUNT(*)::int AS count
		FROM org_audit_logs
		WHERE org_id = ? AND created_at >= NOW() - INTERVAL '1 year'
		GROUP BY day ORDER BY day DESC
	`, orgID).Scan(ctx, &rows); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch audit summary"})
		return
	}

	days := make(map[string]int, len(rows))
	for _, r := range rows {
		days[r.Day] = r.Count
	}
	c.JSON(http.StatusOK, gin.H{"days": days})
}

// DeleteAuditLog removes audit log entries for an org.
// Modes:
//   - "all"    — delete every entry for this org
//   - "months" — delete entries whose month matches any of req.Months (format: "YYYY-MM")
//   - "days"   — delete entries whose date matches any of req.Days (format: "YYYY-MM-DD")
//
// Restricted to admins and owners.
func (h *OrgHandler) DeleteAuditLog(c *gin.Context) {
	callerID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}

	var req struct {
		Mode   string   `json:"mode" binding:"required"`
		Months []string `json:"months"`
		Days   []string `json:"days"`
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

	q := h.DB.NewDelete().Model((*pkg.OrgAuditLog)(nil)).Where("org_id = ?", orgID)

	switch req.Mode {
	case "all":
		// no additional filter — delete everything for this org
	case "months":
		if len(req.Months) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no months specified"})
			return
		}
		q = q.Where("to_char(created_at, 'YYYY-MM') IN (?)", bun.In(req.Months))
	case "days":
		if len(req.Days) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no days specified"})
			return
		}
		q = q.Where("created_at::date::text IN (?)", bun.In(req.Days))
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid mode: must be all, months, or days"})
		return
	}

	res, err := q.Exec(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete audit log entries"})
		return
	}
	n, _ := res.RowsAffected()
	h.logAudit(ctx, orgID, callerID, "audit_cleared", "", "org",
		fmt.Sprintf("%d entries removed (%s)", n, req.Mode))
	c.JSON(http.StatusOK, gin.H{"deleted": n})
}

// GetOrgAllFileKeys returns all {id, encrypted_key} pairs for non-deleted files in the org.
// Used by the frontend to re-wrap keys during OrgKey rotation.
// Restricted to admins and owners.
func (h *OrgHandler) GetOrgAllFileKeys(c *gin.Context) {
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

	type fileKeyEntry struct {
		ID           int64  `json:"id"`
		EncryptedKey string `json:"encrypted_key"`
	}

	var files []pkg.OrgFile
	if err := h.DB.NewSelect().Model(&files).
		Column("id", "encrypted_key").
		Where("org_id = ?", orgID).
		Scan(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch file keys"})
		return
	}

	result := make([]fileKeyEntry, len(files))
	for i, f := range files {
		result[i] = fileKeyEntry{ID: f.ID, EncryptedKey: f.EncryptedKey}
	}
	c.JSON(http.StatusOK, result)
}
