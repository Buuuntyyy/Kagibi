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

// OrgShareItem is the DTO returned by ListOrgShares.
type OrgShareItem struct {
	ID         int64      `json:"id"`
	FileID     int64      `json:"file_id"`
	FileName   string     `json:"file_name"`
	FilePath   string     `json:"file_path"`
	OwnerID    string     `json:"owner_id"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	Views      int64      `json:"views"`
	SingleUse  bool       `json:"single_use"`
	Token      string     `json:"token"`
}

// ListOrgShares returns all active share links for this org's files.
// Admins/owners see all shares; members see only their own.
func (h *OrgHandler) ListOrgShares(c *gin.Context) {
	userID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid org id"})
		return
	}
	ctx := c.Request.Context()

	role, err := h.memberRole(ctx, orgID, userID)
	if err != nil || role == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member"})
		return
	}

	q := h.DB.NewSelect().
		TableExpr("share_links AS sl").
		ColumnExpr("sl.id, sl.resource_id AS file_id, sl.path AS file_path, sl.owner_id, sl.created_at, sl.expires_at, sl.views, sl.single_use, sl.token").
		Where("sl.org_id = ? AND sl.resource_type = 'org_file'", orgID).
		OrderExpr("sl.created_at DESC")

	if !isAdminOrOwner(role) {
		q = q.Where("sl.owner_id = ?", userID)
	}

	type shareRow struct {
		ID        int64      `bun:"id"`
		FileID    int64      `bun:"file_id"`
		FilePath  string     `bun:"file_path"`
		OwnerID   string     `bun:"owner_id"`
		CreatedAt time.Time  `bun:"created_at"`
		ExpiresAt *time.Time `bun:"expires_at"`
		Views     int64      `bun:"views"`
		SingleUse bool       `bun:"single_use"`
		Token     string     `bun:"token"`
	}
	var rows []shareRow
	if err := q.Scan(ctx, &rows); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list shares"})
		return
	}

	items := make([]OrgShareItem, 0, len(rows))
	for _, row := range rows {
		item := OrgShareItem{
			ID:        row.ID,
			FileID:    row.FileID,
			FilePath:  row.FilePath,
			OwnerID:   row.OwnerID,
			CreatedAt: row.CreatedAt,
			ExpiresAt: row.ExpiresAt,
			Views:     row.Views,
			SingleUse: row.SingleUse,
			Token:     row.Token,
		}
		var file pkg.OrgFile
		if err := h.DB.NewSelect().Model(&file).
			Where("ofile.id = ? AND ofile.org_id = ?", row.FileID, orgID).
			Scan(ctx); err == nil {
			item.FileName = file.Name
		}
		items = append(items, item)
	}

	c.JSON(http.StatusOK, items)
}

// RevokeOrgShare deletes an org share link. Admins/owners can revoke any; members only their own.
func (h *OrgHandler) RevokeOrgShare(c *gin.Context) {
	userID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid org id"})
		return
	}
	shareID, err := strconv.ParseInt(c.Param("shareID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid share id"})
		return
	}
	ctx := c.Request.Context()

	role, err := h.memberRole(ctx, orgID, userID)
	if err != nil || role == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member"})
		return
	}

	var share pkg.ShareLink
	if err := h.DB.NewSelect().Model(&share).
		Where("sl.id = ? AND sl.org_id = ? AND sl.resource_type = 'org_file'", shareID, orgID).
		Scan(ctx); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "share not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch share"})
		return
	}

	if !isAdminOrOwner(role) && share.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot revoke another member's share"})
		return
	}

	if _, err := h.DB.NewDelete().Model((*pkg.ShareLink)(nil)).
		Where("id = ?", shareID).
		Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke share"})
		return
	}

	h.logAudit(ctx, orgID, userID, "file_share_revoked", strconv.FormatInt(share.ResourceID, 10), "file", "")
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
