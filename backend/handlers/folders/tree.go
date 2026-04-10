// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

// backend/handlers/folders/tree.go
package folders

import (
	"context"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"kagibi/backend/pkg"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// FileTreeItem represents a file in the folder tree response
type FileTreeItem struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Path         string `json:"path"`          // Full path in DB
	RelativePath string `json:"relative_path"` // Path relative to requested folder (for ZIP structure)
	Size         int64  `json:"size"`
	MimeType     string `json:"mime_type"`
	EncryptedKey string `json:"encrypted_key"` // For decryption
}

// FolderTreeItem represents a folder in the tree (for keys hierarchy)
type FolderTreeItem struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Path         string `json:"path"`
	RelativePath string `json:"relative_path"`
	EncryptedKey string `json:"encrypted_key"`
}

// TreeResponse contains the complete folder tree for ZIP download
type TreeResponse struct {
	RootFolder    string            `json:"root_folder"`    // Name of the root folder
	TotalSize     int64             `json:"total_size"`     // Sum of all file sizes (for progress calc)
	TotalFiles    int               `json:"total_files"`    // Number of files
	TotalFolders  int               `json:"total_folders"`  // Number of subfolders
	Files         []FileTreeItem    `json:"files"`          // All files with relative paths
	Folders       []FolderTreeItem  `json:"folders"`        // All folders (for key hierarchy)
	EncryptedKeys map[string]string `json:"encrypted_keys"` // Map of fileID -> encryptedKey (for batch decryption)
}

// GetFolderTreeHandler returns the complete tree of a folder for ZIP download
// GET /api/v1/folders/:id/tree
func GetFolderTreeHandler(c *gin.Context, db *bun.DB) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	folderIDStr := c.Param("id")
	folderID, err := strconv.ParseInt(folderIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid folder ID"})
		return
	}

	// Fetch the root folder
	rootFolder := new(pkg.Folder)
	if err := db.NewSelect().Model(rootFolder).
		Where("id = ? AND user_id = ?", folderID, userID).
		Scan(ctx); err != nil {
		log.Printf("Folder not found or no access: FolderID=%d, UserID=%s", folderID, userID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}

	// Base path for relative path calculation
	basePath := rootFolder.Path
	if !strings.HasSuffix(basePath, "/") {
		basePath = basePath + "/"
	}

	// Fetch all files recursively (files under this folder's path)
	var files []pkg.File
	if err := db.NewSelect().Model(&files).
		Where("user_id = ?", userID).
		Where("path LIKE ?", rootFolder.Path+"%").
		Where("is_preview = ?", false).
		Scan(ctx); err != nil {
		log.Printf("Failed to fetch files for folder tree: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch files"})
		return
	}

	// Fetch all subfolders recursively
	var folders []pkg.Folder
	if err := db.NewSelect().Model(&folders).
		Where("user_id = ?", userID).
		Where("path LIKE ?", rootFolder.Path+"/%"). // Only subfolders, not the root itself
		Scan(ctx); err != nil {
		log.Printf("Failed to fetch subfolders for folder tree: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch folders"})
		return
	}

	// Build response
	response := TreeResponse{
		RootFolder:    rootFolder.Name,
		TotalSize:     0,
		TotalFiles:    len(files),
		TotalFolders:  len(folders),
		Files:         make([]FileTreeItem, 0, len(files)),
		Folders:       make([]FolderTreeItem, 0, len(folders)+1),
		EncryptedKeys: make(map[string]string),
	}

	// Add root folder
	response.Folders = append(response.Folders, FolderTreeItem{
		ID:           rootFolder.ID,
		Name:         rootFolder.Name,
		Path:         rootFolder.Path,
		RelativePath: "",
		EncryptedKey: rootFolder.EncryptedKey,
	})

	// Process folders
	for _, f := range folders {
		relativePath := strings.TrimPrefix(f.Path, rootFolder.Path)
		relativePath = strings.TrimPrefix(relativePath, "/")

		response.Folders = append(response.Folders, FolderTreeItem{
			ID:           f.ID,
			Name:         f.Name,
			Path:         f.Path,
			RelativePath: relativePath,
			EncryptedKey: f.EncryptedKey,
		})
	}

	// Process files
	for _, f := range files {
		// Calculate relative path from root folder
		dir := path.Dir(f.Path)
		relativePath := strings.TrimPrefix(dir, rootFolder.Path)
		relativePath = strings.TrimPrefix(relativePath, "/")

		// Full relative path including filename
		var fullRelativePath string
		if relativePath == "" {
			fullRelativePath = f.Name
		} else {
			fullRelativePath = path.Join(relativePath, f.Name)
		}

		response.Files = append(response.Files, FileTreeItem{
			ID:           f.ID,
			Name:         f.Name,
			Path:         f.Path,
			RelativePath: fullRelativePath,
			Size:         f.Size,
			MimeType:     f.MimeType,
			EncryptedKey: f.EncryptedKey,
		})

		response.TotalSize += f.Size
		response.EncryptedKeys[strconv.FormatInt(f.ID, 10)] = f.EncryptedKey
	}

	log.Printf("Folder tree fetched: UserID=%s, FolderID=%d, Files=%d, Folders=%d, TotalSize=%d",
		userID, folderID, response.TotalFiles, response.TotalFolders, response.TotalSize)

	c.JSON(http.StatusOK, response)
}
