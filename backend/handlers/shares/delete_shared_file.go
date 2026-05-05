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

// DeleteFileFromSharedFolderHandler deletes a file from a shared folder.
// Requires perm_delete on the share and can_delete not overridden to false.
func DeleteFileFromSharedFolderHandler(c *gin.Context, db *bun.DB) {
	token := c.Param("token")
	fileID, err := strconv.ParseInt(c.Param("file_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
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

	var file pkg.File
	if err := db.NewSelect().Model(&file).
		Where("id = ? AND user_id = ?", fileID, shareLink.OwnerID).
		Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Verify the file is within the share path
	if !strings.HasPrefix(file.Path, shareLink.Path) {
		c.JSON(http.StatusForbidden, gin.H{"error": "File not within shared directory"})
		return
	}

	// Load all overrides and check can_delete with ancestor cascade
	var overrides []pkg.ShareItemOverride
	_ = db.NewSelect().Model(&overrides).
		Where("share_id = ?", shareLink.ID).
		Scan(c.Request.Context())

	overrideMap := make(map[string]pkg.ShareItemOverride, len(overrides))
	for _, o := range overrides {
		overrideMap[o.ItemPath] = o
	}

	if !effectiveCanDelete(overrideMap, file.Path, true) {
		c.JSON(http.StatusForbidden, gin.H{"error": "This file is protected from deletion"})
		return
	}

	// Delete from S3
	s3Key := fmt.Sprintf("users/%s%s", shareLink.OwnerID, file.Path)
	if _, err := s3storage.Client.DeleteObject(c.Request.Context(), &s3.DeleteObjectInput{
		Bucket: aws.String(s3storage.BucketName),
		Key:    aws.String(s3Key),
	}); err != nil {
		log.Printf("S3 delete error (continuing): %v", err)
	}

	// Update folder sizes
	if !file.IsPreview {
		if err := pkg.UpdateFolderSizesForFile(c.Request.Context(), db, shareLink.OwnerID, file.Path, -file.Size); err != nil {
			log.Printf("Failed to update folder sizes: %v", err)
		}
	}

	// Delete from DB + update quota
	tx, err := db.BeginTx(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction error"})
		return
	}
	defer tx.Rollback()

	if _, err := tx.NewUpdate().Model((*pkg.UserPlan)(nil)).
		Set("storage_used = GREATEST(storage_used - ?, 0)", file.Size).
		Where("user_id = ?", shareLink.OwnerID).
		Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update storage quota"})
		return
	}

	if err := pkg.DeleteFile(tx, fileID, shareLink.OwnerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Commit failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File deleted"})
}
