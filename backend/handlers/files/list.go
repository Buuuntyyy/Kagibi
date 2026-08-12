// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

// backend/handlers/files/list.go
package files

import (
	"log"
	"net/http"
	"strings"

	"kagibi/backend/pkg"
	"kagibi/backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func ListFilesHandler(c *gin.Context, db *bun.DB) {
	userIDInterface, _ := c.Get("user_id")
	userID := userIDInterface.(string)

	pathParam := c.Param("path")
	if pathParam == "" {
		pathParam = "/"
	}
	// SanitizeVirtualPath handles URL-decoding (including double-encoded variants)
	// and rejects any ".." component before path.Clean normalization.
	cleanPath, err := utils.SanitizeVirtualPath(pathParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Chemin invalide"})
		return
	}
	// Sanitize newlines/CR that could appear in Gin-decoded path params
	cleanPath = strings.ReplaceAll(strings.ReplaceAll(cleanPath, "\n", "_"), "\r", "_")

	log.Printf("ListFilesHandler: userID=%s path=%s", userID, cleanPath)

	includeFolderSizes := c.Query("include_folder_sizes") == "1"

	files, folders, err := pkg.ListItemsByUser(db, userID, cleanPath, includeFolderSizes)
	if err != nil {
		log.Printf("ListFilesHandler ERROR: Failed to list items - %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("ListFilesHandler: Found %d files and %d folders", len(files), len(folders))

	// Fetch current folder meta (for ID)
	var currentFolderID int64 = 0
	if cleanPath != "/" && cleanPath != "" {
		currentFolder := new(pkg.Folder)
		err := db.NewSelect().Model(currentFolder).
			Column("id").
			Where("user_id = ? AND path = ?", userID, cleanPath).
			Limit(1).
			Scan(c.Request.Context())
		if err == nil {
			currentFolderID = currentFolder.ID
		}
	}

	c.JSON(http.StatusOK, gin.H{"files": files, "folders": folders, "current_folder_id": currentFolderID})
}
