// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package organizations

import (
	"net/http"
	"strconv"
	"time"

	"kagibi/backend/pkg"

	"github.com/gin-gonic/gin"
)

// OrgItemResult is a lightweight representation of a file or folder returned by the search endpoint.
type OrgItemResult struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`        // encrypted (client decrypts)
	Path       string    `json:"path"`        // full virtual path of the item
	ParentPath string    `json:"parent_path"` // path of the containing folder
	Type       string    `json:"type"`        // "file" | "folder"
	Size       int64     `json:"size,omitempty"`
	MimeType   string    `json:"mime_type,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// GetAllOrgItems returns a flat list of all files and folders within an org so
// that the client can decrypt names and perform local full-text search.
// Capped at 500 folders + 500 files to keep response size reasonable.
func (h *OrgHandler) GetAllOrgItems(c *gin.Context) {
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

	results := make([]OrgItemResult, 0, 64)

	var folders []pkg.OrgFolder
	if err := h.DB.NewSelect().Model(&folders).
		Where("org_id = ? AND deleted_at IS NULL", orgID).
		OrderExpr("path ASC").
		Limit(500).
		Scan(ctx); err == nil {
		for _, f := range folders {
			results = append(results, OrgItemResult{
				ID:         f.ID,
				Name:       f.Name,
				Path:       f.Path,
				ParentPath: f.ParentPath,
				Type:       "folder",
				CreatedAt:  f.CreatedAt,
			})
		}
	}

	var files []pkg.OrgFile
	if err := h.DB.NewSelect().Model(&files).
		Where("org_id = ? AND deleted_at IS NULL", orgID).
		OrderExpr("name ASC").
		Limit(500).
		Scan(ctx); err == nil {
		for _, f := range files {
			results = append(results, OrgItemResult{
				ID:         f.ID,
				Name:       f.Name,
				Path:       f.Path,
				ParentPath: f.FolderPath,
				Type:       "file",
				Size:       f.Size,
				MimeType:   f.MimeType,
				CreatedAt:  f.CreatedAt,
			})
		}
	}

	c.JSON(http.StatusOK, results)
}
