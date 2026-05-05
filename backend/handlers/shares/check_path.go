// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package shares

import (
	"kagibi/backend/pkg"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// GetDirectSharesForPathHandler returns active direct folder shares owned by the user
// that cover the given path. Returns each share's ID, root folder ID, and the root
// folder's EncryptedKey so the owner can re-wrap file keys for friends.
func GetDirectSharesForPathHandler(c *gin.Context, db *bun.DB) {
	userID := c.GetString("user_id")
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path is required"})
		return
	}

	type ShareWithFolder struct {
		bun.BaseModel `bun:"table:folder_shares"`
		pkg.FolderShare
		Folder *pkg.Folder `bun:"rel:belongs-to,join:folder_id=id"`
	}

	var sharesWithFolders []ShareWithFolder
	if err := db.NewSelect().
		Model(&sharesWithFolders).
		Relation("Folder").
		Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	type shareInfo struct {
		ID                 int64  `json:"id"`
		RootFolderID       int64  `json:"root_folder_id"`
		FolderEncryptedKey string `json:"folder_encrypted_key"`
	}

	var result []shareInfo
	for _, s := range sharesWithFolders {
		if s.Folder == nil || s.Folder.UserID != userID {
			continue
		}
		folderPath := s.Folder.Path
		if folderPath != "/" && !strings.HasSuffix(folderPath, "/") {
			folderPath += "/"
		}
		// The upload path is within this shared folder
		if strings.HasPrefix(path, folderPath) || path == s.Folder.Path {
			result = append(result, shareInfo{
				ID:                 s.FolderShare.ID,
				RootFolderID:       s.Folder.ID,
				FolderEncryptedKey: s.Folder.EncryptedKey,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"shares": result})
}

// GetActiveSharesForPathHandler returns all active share links that cover the given path
func GetActiveSharesForPathHandler(c *gin.Context, db *bun.DB) {
	userIDInterface, _ := c.Get("user_id")
	userID := userIDInterface.(string)

	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path is required"})
		return
	}

	// Find shares where the share path is a prefix of the current path
	// e.g. Share Path: /docs, Current Path: /docs/work/project
	// We need to find shares where 'path' LIKE share.path + '%'
	// But SQL LIKE is usually 'column LIKE pattern'. Here we want 'pattern LIKE column + %' ?
	// No, we want: current_path LIKE share_path || '%'

	var shares []pkg.ShareLink
	err := db.NewSelect().Model(&shares).
		Where("owner_id = ?", userID).
		Where("resource_type = ?", "folder").
		Where("? LIKE path || '%'", path).
		Scan(c.Request.Context())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"shares": shares})
}
