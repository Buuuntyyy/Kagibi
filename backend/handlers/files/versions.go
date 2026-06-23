// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package files

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"kagibi/backend/pkg"
	"kagibi/backend/pkg/s3storage"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// ListVersionsHandler returns the version history of a personal file.
// GET /files/:fileID/versions
func ListVersionsHandler(c *gin.Context, db *bun.DB) {
	userID := c.GetString("user_id")
	fileID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	// Verify ownership
	if _, err := pkg.GetFile(db, fileID, userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	var versions []pkg.FileVersion
	if err := db.NewSelect().Model(&versions).
		Where("file_id = ? AND user_id = ?", fileID, userID).
		OrderExpr("version_number DESC").
		Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch versions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"versions": versions})
}

// RestoreVersionHandler sets an old version as the current file content.
// POST /files/:fileID/versions/:versionID/restore
// This is a pointer swap: the version's S3 key becomes the active key and a new version
// entry is created for what was previously the current file — no re-upload needed.
func RestoreVersionHandler(c *gin.Context, db *bun.DB) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	userID := c.GetString("user_id")
	fileID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}
	versionID, err := strconv.ParseInt(c.Param("versionID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid version ID"})
		return
	}

	// Load current file
	file, err := pkg.GetFile(db, fileID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Load the target version
	var version pkg.FileVersion
	if err := db.NewSelect().Model(&version).
		Where("id = ? AND file_id = ? AND user_id = ?", versionID, fileID, userID).
		Scan(ctx); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Version not found"})
		return
	}

	currentS3Key := fmt.Sprintf(s3UserPathFormat, userID, file.Path)

	// Copy old version content to the main S3 key
	copySource := fmt.Sprintf("%s/%s", s3storage.BucketName, version.S3Key)
	if _, err := s3storage.Client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(s3storage.BucketName),
		Key:        aws.String(currentS3Key),
		CopySource: aws.String(copySource),
	}); err != nil {
		log.Printf("[RestoreVersion] S3 CopyObject failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restore version in storage"})
		return
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer tx.Rollback()

	// Save current state as a new version before overwriting
	var maxVer int
	_ = tx.NewSelect().TableExpr("file_versions").
		ColumnExpr("COALESCE(MAX(version_number), 0)").
		Where("file_id = ?", fileID).
		Scan(ctx, &maxVer)

	savedS3Key := fmt.Sprintf("users/%s%s~v%d", userID, file.Path, time.Now().UnixMilli())
	// Copy current main S3 key to a version key before it gets replaced
	if _, copyErr := s3storage.Client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(s3storage.BucketName),
		Key:        aws.String(savedS3Key),
		CopySource: aws.String(fmt.Sprintf("%s/%s", s3storage.BucketName, currentS3Key)),
	}); copyErr != nil {
		log.Printf("[RestoreVersion] S3 copy of current to version failed (non-fatal): %v", copyErr)
		savedS3Key = "" // won't save this version record
	}

	if savedS3Key != "" {
		newVer := &pkg.FileVersion{
			FileID:        fileID,
			UserID:        userID,
			VersionNumber: maxVer + 1,
			S3Key:         savedS3Key,
			EncryptedKey:  file.EncryptedKey,
			Size:          file.Size,
			ChunkSize:     file.ChunkSize,
			MimeType:      file.MimeType,
		}
		if _, err := tx.NewInsert().Model(newVer).Exec(ctx); err != nil {
			log.Printf("[RestoreVersion] Failed to save current as version: %v", err)
		} else {
			_, _ = tx.NewUpdate().Model((*pkg.UserPlan)(nil)).
				Set("version_storage_bytes = version_storage_bytes + ?", file.Size).
				Where("user_id = ?", userID).Exec(ctx)
		}
	}

	// Update the files record to reflect the restored version content
	sizeDelta := version.Size - file.Size
	if _, err := tx.NewUpdate().Model((*pkg.File)(nil)).
		Set("encrypted_key = ?", version.EncryptedKey).
		Set("size = ?", version.Size).
		Set("chunk_size = ?", version.ChunkSize).
		Set("mime_type = ?", version.MimeType).
		Set("updated_at = NOW()").
		Where("id = ? AND user_id = ?", fileID, userID).
		Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update file record"})
		return
	}

	// Update primary storage quota with size delta
	if sizeDelta != 0 {
		if _, err := tx.NewUpdate().Model((*pkg.UserPlan)(nil)).
			Set("storage_used = GREATEST(storage_used + ?, 0)", sizeDelta).
			Where("user_id = ?", userID).Exec(ctx); err != nil {
			log.Printf("[RestoreVersion] Failed to update storage_used: %v", err)
		}
	}

	// Remove the restored version record — it is now the current file
	if _, err := tx.NewDelete().Model((*pkg.FileVersion)(nil)).
		Where("id = ?", versionID).Exec(ctx); err != nil {
		log.Printf("[RestoreVersion] Failed to delete restored version record: %v", err)
	} else {
		// Free its space from version storage
		_, _ = tx.NewUpdate().Model((*pkg.UserPlan)(nil)).
			Set("version_storage_bytes = GREATEST(version_storage_bytes - ?, 0)", version.Size).
			Where("user_id = ?", userID).Exec(ctx)
	}

	// Also delete the version's S3 object (it's now the main S3 object)
	go func() {
		bgCtx, bgCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer bgCancel()
		_, _ = s3storage.Client.DeleteObject(bgCtx, &s3.DeleteObjectInput{
			Bucket: aws.String(s3storage.BucketName),
			Key:    aws.String(version.S3Key),
		})
	}()

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit failed"})
		return
	}

	// Update folder sizes
	if sizeDelta != 0 {
		go func() {
			bgCtx := context.Background()
			_ = pkg.UpdateFolderSizesForFile(bgCtx, db, userID, file.Path, sizeDelta)
		}()
	}

	c.JSON(http.StatusOK, gin.H{"message": "Version restored successfully"})
}

// DeleteVersionHandler permanently removes one historical version.
// DELETE /files/:fileID/versions/:versionID
func DeleteVersionHandler(c *gin.Context, db *bun.DB) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	userID := c.GetString("user_id")
	fileID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}
	versionID, err := strconv.ParseInt(c.Param("versionID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid version ID"})
		return
	}

	// Verify ownership
	if _, err := pkg.GetFile(db, fileID, userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	var version pkg.FileVersion
	if err := db.NewSelect().Model(&version).
		Where("id = ? AND file_id = ? AND user_id = ?", versionID, fileID, userID).
		Scan(ctx); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Version not found"})
		return
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer tx.Rollback()

	if _, err := tx.NewDelete().Model((*pkg.FileVersion)(nil)).
		Where("id = ?", versionID).Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete version"})
		return
	}

	_, _ = tx.NewUpdate().Model((*pkg.UserPlan)(nil)).
		Set("version_storage_bytes = GREATEST(version_storage_bytes - ?, 0)", version.Size).
		Where("user_id = ?", userID).Exec(ctx)

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit failed"})
		return
	}

	// Delete from S3 after commit
	go func() {
		bgCtx, bgCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer bgCancel()
		if _, err := s3storage.Client.DeleteObject(bgCtx, &s3.DeleteObjectInput{
			Bucket: aws.String(s3storage.BucketName),
			Key:    aws.String(version.S3Key),
		}); err != nil {
			log.Printf("[DeleteVersion] S3 delete failed (non-fatal): %v", err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "Version deleted"})
}

// GetVersionPresignedDownloadHandler generates a presigned GET URL for a specific version.
// GET /files/:fileID/versions/:versionID/presigned
func GetVersionPresignedDownloadHandler(c *gin.Context, db *bun.DB) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	userID := c.GetString("user_id")
	fileID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}
	versionID, err := strconv.ParseInt(c.Param("versionID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid version ID"})
		return
	}

	file, err := pkg.GetFile(db, fileID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	var version pkg.FileVersion
	if err := db.NewSelect().Model(&version).
		Where("id = ? AND file_id = ? AND user_id = ?", versionID, fileID, userID).
		Scan(ctx); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Version not found"})
		return
	}

	presigner := s3.NewPresignClient(s3storage.Client)
	presignReq, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket:                     aws.String(s3storage.BucketName),
		Key:                        aws.String(version.S3Key),
		ResponseContentDisposition: aws.String(fmt.Sprintf("attachment; filename=\"%s\"", file.Name)),
		ResponseCacheControl:       aws.String("no-store, no-cache, must-revalidate"),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = DownloadPresignTTL
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate download URL"})
		return
	}

	chunkSize := version.ChunkSize
	if chunkSize <= 0 {
		chunkSize = 10 * 1024 * 1024
	}

	c.JSON(http.StatusOK, gin.H{
		"url":            presignReq.URL,
		"expires_in":     int(DownloadPresignTTL.Seconds()),
		"file_name":      file.Name,
		"file_size":      version.Size,
		"mime_type":      version.MimeType,
		"encrypted_key":  version.EncryptedKey,
		"encryption_alg": "AES-GCM-256",
		"chunk_size":     chunkSize,
		"iv_length":      12,
		"tag_length":     16,
	})
}

// enforceVersionLimitForFile deletes the oldest versions if the plan limit is exceeded.
// Runs in a background goroutine after each versioned upload.
func enforceVersionLimitForFile(ctx context.Context, db *bun.DB, fileID int64, userID, plan string) {
	maxVersions := pkg.GetMaxVersions(plan)

	var versions []pkg.FileVersion
	if err := db.NewSelect().Model(&versions).
		Where("file_id = ? AND user_id = ?", fileID, userID).
		OrderExpr("version_number ASC").
		Scan(ctx); err != nil || len(versions) <= maxVersions {
		return
	}

	toDelete := versions[:len(versions)-maxVersions]
	for _, v := range toDelete {
		if _, err := db.NewDelete().Model((*pkg.FileVersion)(nil)).Where("id = ?", v.ID).Exec(ctx); err != nil {
			log.Printf("[VersionLimit] Failed to delete old version %d: %v", v.ID, err)
			continue
		}
		_, _ = db.NewUpdate().Model((*pkg.UserPlan)(nil)).
			Set("version_storage_bytes = GREATEST(version_storage_bytes - ?, 0)", v.Size).
			Where("user_id = ?", userID).Exec(ctx)

		if _, err := s3storage.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(s3storage.BucketName),
			Key:    aws.String(v.S3Key),
		}); err != nil {
			log.Printf("[VersionLimit] S3 delete failed (non-fatal) for key %s: %v", v.S3Key, err)
		}
	}
}

// deleteAllVersionsForFile removes all versions of a file (called when the file itself is deleted).
func deleteAllVersionsForFile(ctx context.Context, db *bun.DB, fileID int64, userID string) {
	var versions []pkg.FileVersion
	if err := db.NewSelect().Model(&versions).
		Where("file_id = ? AND user_id = ?", fileID, userID).
		Scan(ctx); err != nil || len(versions) == 0 {
		return
	}

	var totalVersionSize int64
	for _, v := range versions {
		totalVersionSize += v.Size
		go func(key string) {
			bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			_, _ = s3storage.Client.DeleteObject(bgCtx, &s3.DeleteObjectInput{
				Bucket: aws.String(s3storage.BucketName),
				Key:    aws.String(key),
			})
		}(v.S3Key)
	}

	// versions rows are deleted by ON DELETE CASCADE from files.id
	// but we still need to free the version_storage_bytes quota
	if totalVersionSize > 0 {
		_, _ = db.NewUpdate().Model((*pkg.UserPlan)(nil)).
			Set("version_storage_bytes = GREATEST(version_storage_bytes - ?, 0)", totalVersionSize).
			Where("user_id = ?", userID).Exec(ctx)
	}
}
