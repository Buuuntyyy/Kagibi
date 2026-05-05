// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package shares

import (
	"kagibi/backend/pkg"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// BrowseSharedFolderHandler handles browsing within a shared folder
func BrowseSharedFolderHandler(c *gin.Context, db *bun.DB) {
	token := c.Param("token")
	subpath := c.Param("subpath")
	subpath = strings.ReplaceAll(strings.ReplaceAll(subpath, "\n", "_"), "\r", "_")

	var shareLink pkg.ShareLink
	err := db.NewSelect().Model(&shareLink).Where("token = ?", token).Scan(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Share link not found"})
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

	if !checkSharePassword(c, &shareLink) {
		return
	}

	if shareLink.ResourceType != "folder" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This share link is not for a folder"})
		return
	}

	// Clean and validate the subpath to prevent directory traversal
	requestedPath := path.Join(shareLink.Path, subpath)
	requestedPath = strings.ReplaceAll(strings.ReplaceAll(requestedPath, "\n", "_"), "\r", "_")
	if !strings.HasPrefix(requestedPath, shareLink.Path) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access to this path is forbidden"})
		return
	}

	// Check if the requested subfolder is accessible (not overridden to 'none')
	if subpath != "" && subpath != "/" {
		var override pkg.ShareItemOverride
		overrideErr := db.NewSelect().Model(&override).
			Where("share_id = ? AND item_path = ?", shareLink.ID, requestedPath).
			Scan(c.Request.Context())
		if overrideErr == nil && override.AccessLevel == "none" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access to this folder is restricted"})
			return
		}
	}

	// List files and folders in the requested path
	files, folders, err := pkg.GetSharedFolderContent(db, requestedPath, shareLink.OwnerID, shareLink.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list items"})
		return
	}

	// Load all item overrides for the share (used for cascade permission computation)
	var overrides []pkg.ShareItemOverride
	_ = db.NewSelect().Model(&overrides).Where("share_id = ?", shareLink.ID).Scan(c.Request.Context())

	overrideMap := make(map[string]pkg.ShareItemOverride, len(overrides))
	for _, o := range overrides {
		overrideMap[o.ItemPath] = o
	}

	type folderEntry struct {
		pkg.Folder
		AccessLevel string `json:"access_level"`
		CanDelete   bool   `json:"can_delete"`
		CanDownload bool   `json:"can_download"`
	}
	type fileEntry struct {
		pkg.File
		CanDelete   bool `json:"can_delete"`
		CanDownload bool `json:"can_download"`
	}

	filteredFolders := make([]folderEntry, 0, len(folders))
	for _, f := range folders {
		if ov, ok := overrideMap[f.Path]; ok {
			if ov.AccessLevel == "none" {
				continue
			}
			filteredFolders = append(filteredFolders, folderEntry{
				Folder:      f,
				AccessLevel: ov.AccessLevel,
				CanDelete:   effectiveCanDelete(overrideMap, f.Path, shareLink.PermDelete),
				CanDownload: effectiveCanDownload(overrideMap, f.Path, shareLink.PermDownload),
			})
		} else {
			filteredFolders = append(filteredFolders, folderEntry{
				Folder:      f,
				AccessLevel: "full",
				CanDelete:   shareLink.PermDelete,
				CanDownload: shareLink.PermDownload,
			})
		}
	}

	filteredFiles := make([]fileEntry, 0, len(files))
	for _, f := range files {
		canDl := effectiveCanDownload(overrideMap, f.Path, shareLink.PermDownload)
		canDel := effectiveCanDelete(overrideMap, f.Path, shareLink.PermDelete)

		entry := fileEntry{File: f, CanDelete: canDel, CanDownload: canDl}

		// Access-level 'none' hides the file; also clear key if download is blocked
		if ov, ok := overrideMap[f.Path]; ok && ov.AccessLevel == "none" {
			continue
		}
		if !canDl {
			entry.EncryptedKey = ""
		}
		filteredFiles = append(filteredFiles, entry)
	}

	// Fetch owner info
	var owner pkg.User
	err = db.NewSelect().Model(&owner).Where("id = ?", shareLink.OwnerID).Scan(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve owner info"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"folders":       filteredFolders,
		"files":         filteredFiles,
		"owner_email":   owner.Email,
		"owner_name":    owner.Name,
		"resource_name": path.Base(shareLink.Path),
		"permissions": gin.H{
			"download": shareLink.PermDownload,
			"create":   shareLink.PermCreate,
			"delete":   shareLink.PermDelete,
			"move":     shareLink.PermMove,
		},
	})
}

// effectiveCanDelete returns whether an item at itemPath can be deleted,
// considering the share-level permission and all per-item/ancestor overrides.
func effectiveCanDelete(overrideMap map[string]pkg.ShareItemOverride, itemPath string, sharePermDelete bool) bool {
	if !sharePermDelete {
		return false
	}
	if ov, ok := overrideMap[itemPath]; ok && !ov.CanDelete {
		return false
	}
	return !hasAncestorFolderRestriction(overrideMap, itemPath, func(o pkg.ShareItemOverride) bool {
		return !o.CanDelete
	})
}

// effectiveCanDownload returns whether an item at itemPath can be downloaded,
// considering the share-level permission and all per-item/ancestor overrides.
func effectiveCanDownload(overrideMap map[string]pkg.ShareItemOverride, itemPath string, sharePermDownload bool) bool {
	if !sharePermDownload {
		return false
	}
	if ov, ok := overrideMap[itemPath]; ok && !ov.CanDownload {
		return false
	}
	return !hasAncestorFolderRestriction(overrideMap, itemPath, func(o pkg.ShareItemOverride) bool {
		return !o.CanDownload
	})
}

// hasAncestorFolderRestriction checks if any ancestor folder in overrideMap
// satisfies the given predicate.
func hasAncestorFolderRestriction(overrideMap map[string]pkg.ShareItemOverride, itemPath string, pred func(pkg.ShareItemOverride) bool) bool {
	parts := strings.Split(itemPath, "/")
	// parts[0] is always "" (leading slash), last element is the item itself
	for i := 1; i < len(parts)-1; i++ {
		ancestorPath := strings.Join(parts[:i+1], "/")
		if ov, ok := overrideMap[ancestorPath]; ok {
			if pred(ov) {
				return true
			}
		}
	}
	return false
}
