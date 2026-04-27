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

// DeleteFolderFromSharedFolderHandler deletes a folder and all its contents from a shared folder.
// Requires perm_delete on the share and can_delete not overridden to false on the folder or its ancestors.
func DeleteFolderFromSharedFolderHandler(c *gin.Context, db *bun.DB) {
	token := c.Param("token")
	folderID, err := strconv.ParseInt(c.Param("folder_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid folder ID"})
		return
	}

	var shareLink pkg.ShareLink
	if err := db.NewSelect().Model(&shareLink).Where("token = ?", token).Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Share not found"})
		return
	}

	if shareLink.ExpiresAt != nil && shareLink.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusGone, gin.H{"error": "Share link expired"})
		return
	}

	if !shareLink.PermDelete {
		c.JSON(http.StatusForbidden, gin.H{"error": "Delete not permitted on this share"})
		return
	}

	var folder pkg.Folder
	if err := db.NewSelect().Model(&folder).
		Where("id = ? AND user_id = ?", folderID, shareLink.OwnerID).
		Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}

	// Verify the folder is within the share path (but not the share root itself)
	if !strings.HasPrefix(folder.Path, shareLink.Path+"/") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Folder not within shared directory"})
		return
	}

	// Load overrides and check can_delete with ancestor cascade
	var overrides []pkg.ShareItemOverride
	_ = db.NewSelect().Model(&overrides).
		Where("share_id = ?", shareLink.ID).
		Scan(c.Request.Context())

	overrideMap := make(map[string]pkg.ShareItemOverride, len(overrides))
	for _, o := range overrides {
		overrideMap[o.ItemPath] = o
	}

	if !effectiveCanDelete(overrideMap, folder.Path, true) {
		c.JSON(http.StatusForbidden, gin.H{"error": "This folder is protected from deletion"})
		return
	}

	// Find all files recursively within this folder
	var files []pkg.File
	if err := db.NewSelect().Model(&files).
		Where("user_id = ?", shareLink.OwnerID).
		Where("path LIKE ?", folder.Path+"/%").
		Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list folder contents"})
		return
	}

	// Delete each file from S3
	var totalSize int64
	for _, f := range files {
		s3Key := fmt.Sprintf("users/%s%s", shareLink.OwnerID, f.Path)
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

	// Delete all subfolders and files from DB in a transaction
	tx, err := db.BeginTx(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction error"})
		return
	}
	defer tx.Rollback()

	// Delete all files under the folder path
	if _, err := tx.NewDelete().Model((*pkg.File)(nil)).
		Where("user_id = ? AND path LIKE ?", shareLink.OwnerID, folder.Path+"/%").
		Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete files"})
		return
	}

	// Delete all subfolders under the folder path (deepest first via ORDER BY path DESC)
	if _, err := tx.NewDelete().Model((*pkg.Folder)(nil)).
		Where("user_id = ? AND path LIKE ?", shareLink.OwnerID, folder.Path+"/%").
		Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete subfolders"})
		return
	}

	// Delete the folder itself
	if _, err := tx.NewDelete().Model((*pkg.Folder)(nil)).
		Where("id = ?", folderID).
		Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete folder"})
		return
	}

	// Update storage quota
	if totalSize > 0 {
		if _, err := tx.NewUpdate().Model((*pkg.UserPlan)(nil)).
			Set("storage_used = GREATEST(storage_used - ?, 0)", totalSize).
			Where("user_id = ?", shareLink.OwnerID).
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
