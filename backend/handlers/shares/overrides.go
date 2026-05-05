// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package shares

import (
	"context"
	"kagibi/backend/pkg"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// ListShareItemOverridesHandler returns all per-item overrides for a share (owner only).
func ListShareItemOverridesHandler(c *gin.Context, db *bun.DB) {
	userID := c.GetString("user_id")
	shareID, err := strconv.ParseInt(c.Param("shareID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid share ID"})
		return
	}

	if !isShareOwner(c.Request.Context(), db, shareID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var overrides []pkg.ShareItemOverride
	if err := db.NewSelect().Model(&overrides).Where("share_id = ?", shareID).Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load overrides"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"overrides": overrides})
}

type upsertOverrideRequest struct {
	ItemPath    string `json:"item_path" binding:"required"`
	ItemType    string `json:"item_type" binding:"required"`
	AccessLevel string `json:"access_level"` // "full" | "readonly" | "none"
	CanDelete   *bool  `json:"can_delete"`
	CanDownload *bool  `json:"can_download"`
}

// UpsertShareItemOverrideHandler creates or updates a per-item override (owner only).
func UpsertShareItemOverrideHandler(c *gin.Context, db *bun.DB) {
	userID := c.GetString("user_id")
	shareID, err := strconv.ParseInt(c.Param("shareID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid share ID"})
		return
	}

	if !isShareOwner(c.Request.Context(), db, shareID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var req upsertOverrideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	accessLevel := req.AccessLevel
	if accessLevel == "" {
		accessLevel = "full"
	}
	if accessLevel != "full" && accessLevel != "readonly" && accessLevel != "none" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid access_level"})
		return
	}

	canDelete := true
	if req.CanDelete != nil {
		canDelete = *req.CanDelete
	}

	canDownload := true
	if req.CanDownload != nil {
		canDownload = *req.CanDownload
	}

	override := &pkg.ShareItemOverride{
		ShareID:     shareID,
		ItemPath:    req.ItemPath,
		ItemType:    req.ItemType,
		AccessLevel: accessLevel,
		CanDelete:   canDelete,
		CanDownload: canDownload,
	}

	_, err = db.NewInsert().Model(override).
		On("CONFLICT (share_id, item_path) DO UPDATE").
		Set("access_level = EXCLUDED.access_level").
		Set("can_delete = EXCLUDED.can_delete").
		Set("can_download = EXCLUDED.can_download").
		Exec(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save override"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"override": override})
}

// DeleteShareItemOverrideHandler removes a per-item override (owner only), resetting to default.
func DeleteShareItemOverrideHandler(c *gin.Context, db *bun.DB) {
	userID := c.GetString("user_id")
	shareID, err := strconv.ParseInt(c.Param("shareID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid share ID"})
		return
	}

	if !isShareOwner(c.Request.Context(), db, shareID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	itemPath := c.Query("item_path")
	if itemPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "item_path is required"})
		return
	}

	_, err = db.NewDelete().Model((*pkg.ShareItemOverride)(nil)).
		Where("share_id = ? AND item_path = ?", shareID, itemPath).
		Exec(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete override"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Override removed"})
}

// BrowseShareTreeHandler lets the share owner browse their share tree (for setting overrides).
func BrowseShareTreeHandler(c *gin.Context, db *bun.DB) {
	userID := c.GetString("user_id")
	shareID, err := strconv.ParseInt(c.Param("shareID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid share ID"})
		return
	}

	var shareLink pkg.ShareLink
	if err := db.NewSelect().Model(&shareLink).Where("id = ? AND owner_id = ?", shareID, userID).Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	subpath := c.Param("subpath")
	if subpath == "" {
		subpath = "/"
	}

	requestedPath := subpath
	if subpath == "/" {
		requestedPath = shareLink.Path
	} else {
		requestedPath = shareLink.Path + subpath
	}

	files, folders, err := pkg.GetSharedFolderContent(db, requestedPath, shareLink.OwnerID, shareLink.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list items"})
		return
	}

	// Load overrides
	var overrides []pkg.ShareItemOverride
	_ = db.NewSelect().Model(&overrides).Where("share_id = ?", shareID).Scan(c.Request.Context())
	overrideMap := make(map[string]pkg.ShareItemOverride, len(overrides))
	for _, o := range overrides {
		overrideMap[o.ItemPath] = o
	}

	type folderWithOverride struct {
		pkg.Folder
		AccessLevel string `json:"access_level"`
		CanDelete   bool   `json:"can_delete"`
		CanDownload bool   `json:"can_download"`
	}
	type fileWithOverride struct {
		pkg.File
		CanDelete   bool `json:"can_delete"`
		CanDownload bool `json:"can_download"`
	}

	foldersOut := make([]folderWithOverride, 0, len(folders))
	for _, f := range folders {
		al := "full"
		cd := true
		dl := true
		if ov, ok := overrideMap[f.Path]; ok {
			al = ov.AccessLevel
			cd = ov.CanDelete
			dl = ov.CanDownload
		}
		foldersOut = append(foldersOut, folderWithOverride{Folder: f, AccessLevel: al, CanDelete: cd, CanDownload: dl})
	}

	filesOut := make([]fileWithOverride, 0, len(files))
	for _, f := range files {
		cd := true
		dl := true
		if ov, ok := overrideMap[f.Path]; ok {
			cd = ov.CanDelete
			dl = ov.CanDownload
		}
		filesOut = append(filesOut, fileWithOverride{File: f, CanDelete: cd, CanDownload: dl})
	}

	c.JSON(http.StatusOK, gin.H{
		"folders":    foldersOut,
		"files":      filesOut,
		"share_path": shareLink.Path,
	})
}

type bulkOverrideRequest struct {
	ItemType    string `json:"item_type" binding:"required"` // "folder" | "file"
	AccessLevel string `json:"access_level"`                 // for folders: "full"|"readonly"|"none"
	CanDelete   *bool  `json:"can_delete"`                   // for files and folders
	CanDownload *bool  `json:"can_download"`                 // for files and folders
}

// BulkOverrideHandler applies a permission setting to ALL items of a given type under the share (owner only).
func BulkOverrideHandler(c *gin.Context, db *bun.DB) {
	userID := c.GetString("user_id")
	shareID, err := strconv.ParseInt(c.Param("shareID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid share ID"})
		return
	}

	var shareLink pkg.ShareLink
	if err := db.NewSelect().Model(&shareLink).Where("id = ? AND owner_id = ?", shareID, userID).Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var req bulkOverrideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	ctx := c.Request.Context()
	searchPrefix := shareLink.Path
	if searchPrefix == "/" {
		searchPrefix = ""
	}

	var overrides []pkg.ShareItemOverride

	switch req.ItemType {
	case "folder":
		if req.AccessLevel != "full" && req.AccessLevel != "readonly" && req.AccessLevel != "none" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid access_level"})
			return
		}
		canDelete := true
		if req.CanDelete != nil {
			canDelete = *req.CanDelete
		}
		canDownload := true
		if req.CanDownload != nil {
			canDownload = *req.CanDownload
		}
		var folders []pkg.Folder
		if err := db.NewSelect().Model(&folders).Column("id", "path").
			Where("user_id = ?", shareLink.OwnerID).
			Where("path LIKE ?", searchPrefix+"/%").
			Scan(ctx); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list folders"})
			return
		}
		for _, f := range folders {
			overrides = append(overrides, pkg.ShareItemOverride{
				ShareID: shareID, ItemPath: f.Path, ItemType: "folder",
				AccessLevel: req.AccessLevel, CanDelete: canDelete, CanDownload: canDownload,
			})
		}
	case "file":
		canDelete := true
		if req.CanDelete != nil {
			canDelete = *req.CanDelete
		}
		canDownload := true
		if req.CanDownload != nil {
			canDownload = *req.CanDownload
		}
		var files []pkg.File
		if err := db.NewSelect().Model(&files).Column("id", "path").
			Where("user_id = ?", shareLink.OwnerID).
			Where("is_preview = ?", false).
			Where("path LIKE ?", searchPrefix+"/%").
			Scan(ctx); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list files"})
			return
		}
		for _, f := range files {
			overrides = append(overrides, pkg.ShareItemOverride{
				ShareID: shareID, ItemPath: f.Path, ItemType: "file",
				AccessLevel: "full", CanDelete: canDelete, CanDownload: canDownload,
			})
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item_type"})
		return
	}

	if len(overrides) == 0 {
		c.JSON(http.StatusOK, gin.H{"updated": 0})
		return
	}

	_, err = db.NewInsert().Model(&overrides).
		On("CONFLICT (share_id, item_path) DO UPDATE").
		Set("access_level = EXCLUDED.access_level").
		Set("can_delete = EXCLUDED.can_delete").
		Set("can_download = EXCLUDED.can_download").
		Exec(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save overrides"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"updated": len(overrides)})
}

func isShareOwner(ctx context.Context, db *bun.DB, shareID int64, userID string) bool {
	var sl pkg.ShareLink
	err := db.NewSelect().Model(&sl).Column("id").
		Where("id = ? AND owner_id = ?", shareID, userID).
		Scan(ctx)
	return err == nil
}
