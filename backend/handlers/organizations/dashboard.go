// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package organizations

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// MemberStorageStat aggregates storage and file count for a single org member.
type MemberStorageStat struct {
	UserID       string `json:"user_id"`
	Name         string `json:"name"`
	AvatarURL    string `json:"avatar_url"`
	FileCount    int64  `json:"file_count"`
	StorageBytes int64  `json:"storage_bytes"`
}

// OrgStats is the payload returned by GET /orgs/:orgID/stats.
type OrgStats struct {
	FileCount       int64               `json:"file_count"`
	FolderCount     int64               `json:"folder_count"`
	Activity7d      int64               `json:"activity_7d"`
	MembersNoKey    int64               `json:"members_no_key"`
	StorageByMember []MemberStorageStat `json:"storage_by_member"`
}

// GetOrgStats returns aggregate statistics for an organisation dashboard.
func (h *OrgHandler) GetOrgStats(c *gin.Context) {
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
	if callerRole == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member"})
		return
	}

	var stats OrgStats

	// Total file count.
	if err := h.DB.NewSelect().
		TableExpr("org_files").
		ColumnExpr("COUNT(*)").
		Where("org_id = ? AND deleted_at IS NULL", orgID).
		Scan(ctx, &stats.FileCount); err != nil {
		stats.FileCount = 0
	}

	// Total folder count (exclude root).
	if err := h.DB.NewSelect().
		TableExpr("org_folders").
		ColumnExpr("COUNT(*)").
		Where("org_id = ? AND deleted_at IS NULL", orgID).
		Scan(ctx, &stats.FolderCount); err != nil {
		stats.FolderCount = 0
	}

	// Audit events in the last 7 days.
	if err := h.DB.NewSelect().
		TableExpr("org_audit_logs").
		ColumnExpr("COUNT(*)").
		Where("org_id = ? AND created_at > NOW() - INTERVAL '7 days'", orgID).
		Scan(ctx, &stats.Activity7d); err != nil {
		stats.Activity7d = 0
	}

	// Members without an encrypted org key (need key provisioning).
	if err := h.DB.NewSelect().
		TableExpr("org_members").
		ColumnExpr("COUNT(*)").
		Where("org_id = ? AND (encrypted_org_key IS NULL OR encrypted_org_key = '')", orgID).
		Scan(ctx, &stats.MembersNoKey); err != nil {
		stats.MembersNoKey = 0
	}

	// Storage breakdown by uploader — aggregate first, then resolve all profiles
	// in a single batch query instead of one per uploader.
	type uploaderRow struct {
		UserID       string `bun:"uploaded_by"`
		FileCount    int64  `bun:"file_count"`
		StorageBytes int64  `bun:"storage_bytes"`
	}
	var uploaderRows []uploaderRow
	if err := h.DB.NewSelect().
		TableExpr("org_files").
		ColumnExpr("uploaded_by, COUNT(*) AS file_count, COALESCE(SUM(size), 0) AS storage_bytes").
		Where("org_id = ? AND deleted_at IS NULL", orgID).
		GroupExpr("uploaded_by").
		OrderExpr("storage_bytes DESC").
		Scan(ctx, &uploaderRows); err == nil {

		type profileRow struct {
			ID        string `bun:"id"`
			Name      string `bun:"name"`
			AvatarURL string `bun:"avatar_url"`
		}
		profileMap := make(map[string]profileRow, len(uploaderRows))
		if len(uploaderRows) > 0 {
			userIDs := make([]string, len(uploaderRows))
			for i, r := range uploaderRows {
				userIDs[i] = r.UserID
			}
			var profiles []profileRow
			if err := h.DB.NewSelect().
				TableExpr("profiles").
				ColumnExpr("id, name, avatar_url").
				Where("id IN (?)", bun.In(userIDs)).
				Scan(ctx, &profiles); err == nil {
				for _, p := range profiles {
					profileMap[p.ID] = p
				}
			}
		}

		for _, row := range uploaderRows {
			stat := MemberStorageStat{
				UserID:       row.UserID,
				FileCount:    row.FileCount,
				StorageBytes: row.StorageBytes,
			}
			if p, ok := profileMap[row.UserID]; ok {
				stat.Name = p.Name
				stat.AvatarURL = p.AvatarURL
			}
			stats.StorageByMember = append(stats.StorageByMember, stat)
		}
	}
	if stats.StorageByMember == nil {
		stats.StorageByMember = []MemberStorageStat{}
	}

	c.JSON(http.StatusOK, stats)
}
