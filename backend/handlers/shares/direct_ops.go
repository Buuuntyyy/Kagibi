// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package shares

import (
	"fmt"
	"kagibi/backend/pkg"
	"kagibi/backend/pkg/s3storage"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// resolveDirectShare fetches a FolderShare by ID and verifies that the calling user
// is the recipient (shared_with_user_id). Returns the share and the root folder.
func resolveDirectShare(c *gin.Context, db *bun.DB, shareID int64) (*pkg.FolderShare, *pkg.Folder, bool) {
	userID := c.GetString("user_id")

	var share pkg.FolderShare
	if err := db.NewSelect().Model(&share).Where("id = ?", shareID).Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Share not found"})
		return nil, nil, false
	}
	if share.SharedWithUserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return nil, nil, false
	}

	var rootFolder pkg.Folder
	if err := db.NewSelect().Model(&rootFolder).Where("id = ?", share.FolderID).Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Shared folder not found"})
		return nil, nil, false
	}
	return &share, &rootFolder, true
}

// isWithinShare checks that itemPath is strictly under rootFolder.Path.
func isWithinShare(rootPath, itemPath string) bool {
	prefix := rootPath
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	return strings.HasPrefix(itemPath, prefix)
}

// CreateFolderInDirectShareHandler creates a subfolder inside a directly-shared folder.
// Requires perm_create on the share.
func CreateFolderInDirectShareHandler(c *gin.Context, db *bun.DB) {
	shareID, err := strconv.ParseInt(c.Param("share_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid share ID"})
		return
	}

	share, _, ok := resolveDirectShare(c, db, shareID)
	if !ok {
		return
	}
	if !share.PermCreate {
		c.JSON(http.StatusForbidden, gin.H{"error": "Create not permitted on this share"})
		return
	}

	var req struct {
		Name           string `json:"name"`
		ParentFolderID int64  `json:"parent_folder_id"`
		EncryptedKey   string `json:"encrypted_key"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" || req.ParentFolderID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and parent_folder_id are required"})
		return
	}

	// Verify parent folder belongs to owner and is within share
	var parentFolder pkg.Folder
	if err := db.NewSelect().Model(&parentFolder).Where("id = ?", req.ParentFolderID).Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Parent folder not found"})
		return
	}

	// Verify root folder for path comparison
	var rootFolder pkg.Folder
	if err := db.NewSelect().Model(&rootFolder).Where("id = ?", share.FolderID).Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Root folder not found"})
		return
	}

	if parentFolder.ID != share.FolderID && !isWithinShare(rootFolder.Path, parentFolder.Path) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Parent folder not within shared directory"})
		return
	}

	newPath := parentFolder.Path + "/" + req.Name
	folder := &pkg.Folder{
		Name:         req.Name,
		Path:         newPath,
		UserID:       rootFolder.UserID,
		EncryptedKey: req.EncryptedKey,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if _, err := db.NewInsert().Model(folder).Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create folder"})
		return
	}

	c.JSON(http.StatusCreated, folder)
}

// DeleteFileFromDirectShareHandler deletes a file from a directly-shared folder.
// Requires perm_delete on the share.
func DeleteFileFromDirectShareHandler(c *gin.Context, db *bun.DB) {
	shareID, err := strconv.ParseInt(c.Param("share_id"), 10, 64)
	fileID, err2 := strconv.ParseInt(c.Param("file_id"), 10, 64)
	if err != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid IDs"})
		return
	}

	share, rootFolder, ok := resolveDirectShare(c, db, shareID)
	if !ok {
		return
	}
	if !share.PermDelete {
		c.JSON(http.StatusForbidden, gin.H{"error": "Delete not permitted on this share"})
		return
	}

	var file pkg.File
	if err := db.NewSelect().Model(&file).
		Where("id = ? AND user_id = ?", fileID, rootFolder.UserID).
		Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	if !isWithinShare(rootFolder.Path, file.Path) {
		c.JSON(http.StatusForbidden, gin.H{"error": "File not within shared directory"})
		return
	}

	s3Key := fmt.Sprintf("users/%s%s", rootFolder.UserID, file.Path)
	if _, err := s3storage.Client.DeleteObject(c.Request.Context(), &s3.DeleteObjectInput{
		Bucket: aws.String(s3storage.BucketName),
		Key:    aws.String(s3Key),
	}); err != nil {
		log.Printf("S3 delete error for %s (continuing): %v", s3Key, err)
	}

	if _, err := db.NewDelete().Model((*pkg.File)(nil)).Where("id = ?", fileID).Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file"})
		return
	}

	if !file.IsPreview {
		if _, err := db.NewUpdate().Model((*pkg.UserPlan)(nil)).
			Set("storage_used = GREATEST(storage_used - ?, 0)", file.Size).
			Where("user_id = ?", rootFolder.UserID).
			Exec(c.Request.Context()); err != nil {
			log.Printf("Failed to update storage quota: %v", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "File deleted"})
}

// DeleteFolderFromDirectShareHandler deletes a folder and all its contents from a directly-shared folder.
// Requires perm_delete on the share.
func DeleteFolderFromDirectShareHandler(c *gin.Context, db *bun.DB) {
	shareID, err := strconv.ParseInt(c.Param("share_id"), 10, 64)
	folderID, err2 := strconv.ParseInt(c.Param("folder_id"), 10, 64)
	if err != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid IDs"})
		return
	}

	share, rootFolder, ok := resolveDirectShare(c, db, shareID)
	if !ok {
		return
	}
	if !share.PermDelete {
		c.JSON(http.StatusForbidden, gin.H{"error": "Delete not permitted on this share"})
		return
	}

	var folder pkg.Folder
	if err := db.NewSelect().Model(&folder).
		Where("id = ? AND user_id = ?", folderID, rootFolder.UserID).
		Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}

	if !isWithinShare(rootFolder.Path, folder.Path) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Folder not within shared directory"})
		return
	}

	// Find all files recursively
	var files []pkg.File
	if err := db.NewSelect().Model(&files).
		Where("user_id = ?", rootFolder.UserID).
		Where("path LIKE ?", folder.Path+"/%").
		Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list folder contents"})
		return
	}

	var totalSize int64
	for _, f := range files {
		s3Key := fmt.Sprintf("users/%s%s", rootFolder.UserID, f.Path)
		if _, err := s3storage.Client.DeleteObject(c.Request.Context(), &s3.DeleteObjectInput{
			Bucket: aws.String(s3storage.BucketName),
			Key:    aws.String(s3Key),
		}); err != nil {
			log.Printf("S3 delete error for %s (continuing): %v", s3Key, err)
		}
		if !f.IsPreview {
			totalSize += f.Size
		}
	}

	tx, err := db.BeginTx(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction error"})
		return
	}
	defer tx.Rollback()

	if _, err := tx.NewDelete().Model((*pkg.File)(nil)).
		Where("user_id = ? AND path LIKE ?", rootFolder.UserID, folder.Path+"/%").
		Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete files"})
		return
	}

	if _, err := tx.NewDelete().Model((*pkg.Folder)(nil)).
		Where("user_id = ? AND path LIKE ?", rootFolder.UserID, folder.Path+"/%").
		Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete subfolders"})
		return
	}

	if _, err := tx.NewDelete().Model((*pkg.Folder)(nil)).
		Where("id = ?", folderID).
		Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete folder"})
		return
	}

	if totalSize > 0 {
		if _, err := tx.NewUpdate().Model((*pkg.UserPlan)(nil)).
			Set("storage_used = GREATEST(storage_used - ?, 0)", totalSize).
			Where("user_id = ?", rootFolder.UserID).
			Exec(c.Request.Context()); err != nil {
			log.Printf("Failed to update storage quota: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Commit failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Folder deleted"})
}

// RenameInDirectShareHandler renames a file or folder within a directly-shared folder.
// Requires perm_move on the share.
func RenameInDirectShareHandler(c *gin.Context, db *bun.DB) {
	shareID, err := strconv.ParseInt(c.Param("share_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid share ID"})
		return
	}

	share, rootFolder, ok := resolveDirectShare(c, db, shareID)
	if !ok {
		return
	}
	if !share.PermMove {
		c.JSON(http.StatusForbidden, gin.H{"error": "Rename not permitted on this share"})
		return
	}

	var req struct {
		ID      int64  `json:"id"`
		Type    string `json:"type"`
		NewName string `json:"new_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ID == 0 || req.NewName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id, type, and new_name are required"})
		return
	}

	switch req.Type {
	case "file":
		var file pkg.File
		if err := db.NewSelect().Model(&file).
			Where("id = ? AND user_id = ?", req.ID, rootFolder.UserID).
			Scan(c.Request.Context()); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}
		if !isWithinShare(rootFolder.Path, file.Path) {
			c.JSON(http.StatusForbidden, gin.H{"error": "File not within shared directory"})
			return
		}
		if _, err := db.NewUpdate().Model((*pkg.File)(nil)).
			Set("name = ?", req.NewName).
			Where("id = ?", req.ID).
			Exec(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to rename file"})
			return
		}

	case "folder":
		var folder pkg.Folder
		if err := db.NewSelect().Model(&folder).
			Where("id = ? AND user_id = ?", req.ID, rootFolder.UserID).
			Scan(c.Request.Context()); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
			return
		}
		if !isWithinShare(rootFolder.Path, folder.Path) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Folder not within shared directory"})
			return
		}
		oldPath := folder.Path
		newPath := folder.Path[:strings.LastIndex(folder.Path, "/")+1] + req.NewName

		tx, err := db.BeginTx(c.Request.Context(), nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction error"})
			return
		}
		defer tx.Rollback()

		if _, err := tx.NewUpdate().Model((*pkg.Folder)(nil)).
			Set("name = ?, path = ?", req.NewName, newPath).
			Where("id = ?", req.ID).
			Exec(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to rename folder"})
			return
		}

		// Update paths of all descendants
		if _, err := tx.NewUpdate().Model((*pkg.Folder)(nil)).
			Set("path = REPLACE(path, ?, ?)", oldPath+"/", newPath+"/").
			Where("user_id = ? AND path LIKE ?", rootFolder.UserID, oldPath+"/%").
			Exec(c.Request.Context()); err != nil {
			log.Printf("Warning: failed to update subfolder paths: %v", err)
		}
		if _, err := tx.NewUpdate().Model((*pkg.File)(nil)).
			Set("path = REPLACE(path, ?, ?)", oldPath+"/", newPath+"/").
			Where("user_id = ? AND path LIKE ?", rootFolder.UserID, oldPath+"/%").
			Exec(c.Request.Context()); err != nil {
			log.Printf("Warning: failed to update file paths: %v", err)
		}

		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Commit failed"})
			return
		}

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "type must be 'file' or 'folder'"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Renamed successfully"})
}
