// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package files

import (
	"fmt"
	"kagibi/backend/pkg"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func ListAllFilesRecursiveHandler(c *gin.Context, db *bun.DB) {
	userIDInterface, _ := c.Get("user_id")
	userID := userIDInterface.(string)

	path := c.Query("path")
	path = strings.ReplaceAll(strings.ReplaceAll(path, "\n", "_"), "\r", "_")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path is required"})
		return
	}

	files, folders, err := pkg.GetFolderContentRecursive(db, userID, path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log for debugging
	fmt.Printf("RecursiveList: Path=%s UserID=%s Found=%d files, %d folders\n", path, userID, len(files), len(folders))
	if len(files) > 0 {
		fmt.Printf("First file: ID=%d Name=%s EncryptedKeyLen=%d\n", files[0].ID, files[0].Name, len(files[0].EncryptedKey))
	}
	if len(folders) > 0 {
		fmt.Printf("First folder: ID=%d Name=%s EncryptedKeyLen=%d\n", folders[0].ID, folders[0].Name, len(folders[0].EncryptedKey))
	}

	c.JSON(http.StatusOK, gin.H{"files": files, "folders": folders})
}
