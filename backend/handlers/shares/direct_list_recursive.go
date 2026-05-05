// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package shares

import (
	"net/http"
	"strconv"
	"strings"

	"kagibi/backend/pkg"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// DirectFolderFilesRecursiveHandler returns all files recursively under a directly-shared
// folder, including their keys encrypted with the root folder key.
// The caller must have been granted access to the root folder share.
func DirectFolderFilesRecursiveHandler(c *gin.Context, db *bun.DB) {
	userID := c.GetString("user_id")
	folderID, err := strconv.ParseInt(c.Param("folderID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid folder ID"})
		return
	}

	targetFolder, err := getTargetFolder(c.Request.Context(), db, folderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}

	share, isAuthorized := checkFolderShareAccess(c.Request.Context(), db, userID, targetFolder)
	if !isAuthorized {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if !share.PermDownload {
		c.JSON(http.StatusForbidden, gin.H{"error": "Download not permitted on this share"})
		return
	}

	rootFolderID := share.FolderID

	// Load root folder path to compute relative paths
	rootFolder, err := getTargetFolder(c.Request.Context(), db, rootFolderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load root folder"})
		return
	}

	// Get all files recursively under the target folder
	searchPrefix := targetFolder.Path
	if searchPrefix == "/" {
		searchPrefix = ""
	}

	var allFiles []pkg.File
	if err := db.NewSelect().Model(&allFiles).
		Where("user_id = ?", targetFolder.UserID).
		Where("is_preview = ?", false).
		Where("path LIKE ?", searchPrefix+"/%").
		Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list files"})
		return
	}

	if len(allFiles) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"files":              []interface{}{},
			"root_encrypted_key": share.EncryptedKey,
			"root_folder_id":     rootFolderID,
		})
		return
	}

	// Load file keys for all files (keyed to root folder)
	fileIDs := make([]int64, len(allFiles))
	for i, f := range allFiles {
		fileIDs[i] = f.ID
	}
	var folderFileKeys []pkg.FolderFileKey
	if err := db.NewSelect().Model(&folderFileKeys).
		Where("folder_id = ?", rootFolderID).
		Where("file_id IN (?)", bun.In(fileIDs)).
		Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load file keys"})
		return
	}
	keyMap := make(map[int64]string, len(folderFileKeys))
	for _, k := range folderFileKeys {
		keyMap[k.FileID] = k.EncryptedKey
	}

	type fileEntry struct {
		ID           int64  `json:"id"`
		Name         string `json:"name"`
		Size         int64  `json:"size"`
		MimeType     string `json:"mime_type"`
		EncryptedKey string `json:"encrypted_key"`
		RelativePath string `json:"relative_path"`
	}

	result := make([]fileEntry, 0, len(allFiles))
	rootPathPrefix := rootFolder.Path
	for _, f := range allFiles {
		encKey := keyMap[f.ID]
		if encKey == "" {
			continue // skip files without a share key
		}
		rel := strings.TrimPrefix(f.Path, rootPathPrefix+"/")
		if rel == "" || rel == f.Path {
			rel = f.Name
		}
		result = append(result, fileEntry{
			ID:           f.ID,
			Name:         f.Name,
			Size:         f.Size,
			MimeType:     f.MimeType,
			EncryptedKey: encKey,
			RelativePath: rel,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"files":              result,
		"root_encrypted_key": share.EncryptedKey,
		"root_folder_id":     rootFolderID,
	})
}
