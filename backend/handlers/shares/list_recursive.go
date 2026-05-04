// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package shares

import (
	"net/http"
	"path"
	"strings"
	"time"

	"kagibi/backend/pkg"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// ListSharedFilesRecursiveHandler returns all downloadable files (recursively) under a subpath
// of a folder share. Used by the client to assemble ZIP archives client-side.
func ListSharedFilesRecursiveHandler(c *gin.Context, db *bun.DB) {
	token := c.Param("token")

	var shareLink pkg.ShareLink
	if err := db.NewSelect().Model(&shareLink).Where("token = ?", token).Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
		return
	}

	if shareLink.ExpiresAt != nil && shareLink.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusGone, gin.H{"error": "Link expired"})
		return
	}
	if shareLink.SingleUse && shareLink.UsedAt != nil {
		c.JSON(http.StatusGone, gin.H{"error": "Link already used"})
		return
	}
	if shareLink.ResourceType != "folder" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not a folder share"})
		return
	}
	if !shareLink.PermDownload {
		c.JSON(http.StatusForbidden, gin.H{"error": "Download not permitted on this share"})
		return
	}

	subpath := c.Query("subpath")
	if subpath == "" {
		subpath = "/"
	}
	subpath = strings.ReplaceAll(strings.ReplaceAll(subpath, "\n", "_"), "\r", "_")

	requestedPath := path.Join(shareLink.Path, subpath)
	if !strings.HasPrefix(requestedPath, shareLink.Path) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access forbidden"})
		return
	}

	files, err := pkg.GetSharedFilesRecursive(db, requestedPath, shareLink.OwnerID, shareLink.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list files"})
		return
	}

	// Load all overrides once
	var overrides []pkg.ShareItemOverride
	_ = db.NewSelect().Model(&overrides).Where("share_id = ?", shareLink.ID).Scan(c.Request.Context())
	overrideMap := make(map[string]pkg.ShareItemOverride, len(overrides))
	for _, o := range overrides {
		overrideMap[o.ItemPath] = o
	}

	type fileEntry struct {
		ID           int64  `json:"id"`
		Name         string `json:"name"`
		Size         int64  `json:"size"`
		MimeType     string `json:"mime_type"`
		EncryptedKey string `json:"encrypted_key"`
		RelativePath string `json:"relative_path"`
	}

	result := make([]fileEntry, 0, len(files))
	var totalSize int64
	for _, f := range files {
		if ov, ok := overrideMap[f.Path]; ok && ov.AccessLevel == "none" {
			continue
		}
		if !effectiveCanDownload(overrideMap, f.Path, shareLink.PermDownload) {
			continue
		}
		if f.EncryptedKey == "" {
			continue // key missing — cannot decrypt
		}

		rel := strings.TrimPrefix(f.Path, requestedPath+"/")
		if rel == "" || rel == f.Path {
			rel = f.Name
		}

		totalSize += f.Size
		result = append(result, fileEntry{
			ID:           f.ID,
			Name:         f.Name,
			Size:         f.Size,
			MimeType:     f.MimeType,
			EncryptedKey: f.EncryptedKey,
			RelativePath: rel,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"files":      result,
		"total":      len(result),
		"total_size": totalSize,
	})
}
