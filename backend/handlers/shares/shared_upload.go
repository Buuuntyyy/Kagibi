// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package shares

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"kagibi/backend/pkg"
	"kagibi/backend/pkg/s3storage"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

const (
	sharedPresignTTL = 180 * time.Second
	sharedMinPart    = 5 * 1024 * 1024
	sharedMaxPart    = 100 * 1024 * 1024
)

// upsertSharedFile inserts a new file record or updates the existing one at the same
// (user_id, path). It populates file.ID in both cases and returns the storage delta
// (positive for new files, positive/negative for overwrites).
// This avoids ON CONFLICT which requires a UNIQUE constraint — the files table only
// has a regular index on (user_id, path).
func upsertSharedFile(ctx context.Context, tx bun.Tx, file *pkg.File) (int64, error) {
	var existing pkg.File
	err := tx.NewSelect().Model(&existing).
		Where("user_id = ? AND path = ?", file.UserID, file.Path).
		Scan(ctx)

	if err == nil {
		// File already exists — update in place, preserve ID for FK references
		file.ID = existing.ID
		delta := file.Size - existing.Size
		if _, err := tx.NewUpdate().Model(file).Where("id = ?", file.ID).Exec(ctx); err != nil {
			return 0, err
		}
		return delta, nil
	}

	// New file — plain insert
	if _, err := tx.NewInsert().Model(file).Exec(ctx); err != nil {
		return 0, err
	}
	return file.Size, nil
}

type SharedInitiateRequest struct {
	FileName    string `json:"file_name" binding:"required"`
	FilePath    string `json:"file_path"`
	ContentType string `json:"content_type"`
	TotalSize   int64  `json:"total_size" binding:"required,min=1"`
	TotalParts  int    `json:"total_parts" binding:"required,min=1"`
	EncryptedKey string `json:"encrypted_key" binding:"required"`
}

type SharedCompleteRequest struct {
	UploadID     string              `json:"upload_id" binding:"required"`
	Key          string              `json:"key" binding:"required"`
	Parts        []sharedCompletePart `json:"parts" binding:"required,min=1"`
	FileName     string              `json:"file_name" binding:"required"`
	FilePath     string              `json:"file_path"`
	TotalSize    int64               `json:"total_size" binding:"required"`
	ContentType  string              `json:"content_type"`
	EncryptedKey string              `json:"encrypted_key" binding:"required"`
}

type sharedCompletePart struct {
	PartNumber int    `json:"part_number"`
	ETag       string `json:"etag"`
}

type SharedAbortRequest struct {
	UploadID string `json:"upload_id" binding:"required"`
	Key      string `json:"key" binding:"required"`
}

// InitiateSharedMultipartHandler starts a multipart upload within a directly-shared folder.
// The file is stored under the OWNER's S3 namespace (not the caller's).
// Requires perm_create on the share.
func InitiateSharedMultipartHandler(c *gin.Context, db *bun.DB) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	shareID, err := strconv.ParseInt(c.Param("share_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid share ID"})
		return
	}

	share, rootFolder, ok := resolveDirectShare(c, db, shareID)
	if !ok {
		return
	}
	if !share.PermCreate {
		c.JSON(http.StatusForbidden, gin.H{"error": "Create not permitted on this share"})
		return
	}

	var req SharedInitiateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Validate file path — must be within share root
	filePath := req.FilePath
	if filePath == "" {
		filePath = rootFolder.Path
	}
	if !strings.HasPrefix(filePath, rootFolder.Path) {
		c.JSON(http.StatusForbidden, gin.H{"error": "File path not within shared directory"})
		return
	}

	// Sanitise name (no newlines)
	fileName := strings.ReplaceAll(strings.ReplaceAll(req.FileName, "\n", "_"), "\r", "_")

	fullPath := path.Join(filePath, fileName)
	if !strings.HasPrefix(fullPath, "/") {
		fullPath = "/" + fullPath
	}
	s3Key := fmt.Sprintf("users/%s%s", rootFolder.UserID, fullPath)

	contentType := req.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	createOutput, err := s3storage.Client.CreateMultipartUpload(ctx, &s3.CreateMultipartUploadInput{
		Bucket:      aws.String(s3storage.BucketName),
		Key:         aws.String(s3Key),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		log.Printf("S3 CreateMultipartUpload (shared) error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate upload"})
		return
	}

	uploadID := *createOutput.UploadId
	presigner := s3.NewPresignClient(s3storage.Client)
	presignedURLs, err := sharedGeneratePresignedParts(ctx, presigner, s3Key, uploadID, req.TotalParts, req.TotalSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate presigned URLs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"upload_id":      uploadID,
		"key":            s3Key,
		"presigned_urls": presignedURLs,
	})
}

// CompleteSharedMultipartHandler completes a multipart upload within a directly-shared folder.
// Creates the File record owned by the OWNER and a FolderFileKey entry.
func CompleteSharedMultipartHandler(c *gin.Context, db *bun.DB) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	shareID, err := strconv.ParseInt(c.Param("share_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid share ID"})
		return
	}

	share, rootFolder, ok := resolveDirectShare(c, db, shareID)
	if !ok {
		return
	}
	if !share.PermCreate {
		c.JSON(http.StatusForbidden, gin.H{"error": "Create not permitted on this share"})
		return
	}

	var req SharedCompleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Security: key must be under owner's namespace
	expectedPrefix := fmt.Sprintf("users/%s/", rootFolder.UserID)
	if !strings.HasPrefix(req.Key, expectedPrefix) {
		log.Printf("SECURITY: shared upload key %s not under owner %s", req.Key, rootFolder.UserID)
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Build completed parts
	completedParts := make([]types.CompletedPart, 0, len(req.Parts))
	for _, p := range req.Parts {
		etag := p.ETag
		if !strings.HasPrefix(etag, "\"") {
			etag = "\"" + etag + "\""
		}
		completedParts = append(completedParts, types.CompletedPart{
			PartNumber: aws.Int32(int32(p.PartNumber)),
			ETag:       aws.String(etag),
		})
	}
	sort.Slice(completedParts, func(i, j int) bool {
		return *completedParts[i].PartNumber < *completedParts[j].PartNumber
	})

	_, err = s3storage.Client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(s3storage.BucketName),
		Key:      aws.String(req.Key),
		UploadId: aws.String(req.UploadID),
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: completedParts,
		},
	})
	if err != nil {
		log.Printf("S3 CompleteMultipartUpload (shared) error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete upload"})
		return
	}

	// Derive the DB file path from the S3 key (strip "users/{ownerID}" prefix)
	ownerPrefix := fmt.Sprintf("users/%s", rootFolder.UserID)
	fullPathDB := strings.TrimPrefix(req.Key, ownerPrefix)
	if !strings.HasPrefix(fullPathDB, "/") {
		fullPathDB = "/" + fullPathDB
	}

	contentType := req.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	fileRecord := &pkg.File{
		Name:         req.FileName,
		Path:         fullPathDB,
		Size:         req.TotalSize,
		MimeType:     contentType,
		UserID:       rootFolder.UserID,
		EncryptedKey: "", // uploaded by friend — key is in FolderFileKey
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	tx, txErr := db.BeginTx(ctx, nil)
	if txErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction error"})
		return
	}
	defer tx.Rollback()

	delta, err := upsertSharedFile(ctx, tx, fileRecord)
	if err != nil {
		log.Printf("upsertSharedFile (direct) error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file record"})
		return
	}

	// Store file key encrypted with folder key
	folderFileKey := &pkg.FolderFileKey{
		FolderID:     share.FolderID,
		FileID:       fileRecord.ID,
		EncryptedKey: req.EncryptedKey,
	}
	if _, err := tx.NewInsert().Model(folderFileKey).
		On("CONFLICT (folder_id, file_id) DO UPDATE").
		Set("encrypted_key = EXCLUDED.encrypted_key").
		Exec(ctx); err != nil {
		log.Printf("Warning: failed to store FolderFileKey: %v", err)
	}

	// Update owner's storage quota by the effective delta (handles overwrites correctly)
	if delta != 0 {
		if _, err := tx.NewUpdate().Model((*pkg.UserPlan)(nil)).
			Set("storage_used = storage_used + ?", delta).
			Where("user_id = ?", rootFolder.UserID).
			Exec(ctx); err != nil {
			log.Printf("Warning: failed to update storage quota: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Commit failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Upload completed", "file": fileRecord})
}

// AbortSharedMultipartHandler cancels an in-progress shared multipart upload.
func AbortSharedMultipartHandler(c *gin.Context, db *bun.DB) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	shareID, err := strconv.ParseInt(c.Param("share_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid share ID"})
		return
	}

	_, rootFolder, ok := resolveDirectShare(c, db, shareID)
	if !ok {
		return
	}

	var req SharedAbortRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	expectedPrefix := fmt.Sprintf("users/%s/", rootFolder.UserID)
	if !strings.HasPrefix(req.Key, expectedPrefix) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if _, err := s3storage.Client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(s3storage.BucketName),
		Key:      aws.String(req.Key),
		UploadId: aws.String(req.UploadID),
	}); err != nil {
		log.Printf("S3 AbortMultipartUpload (shared) error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to abort upload"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Upload aborted"})
}

// GetSharedFolderTreeHandler returns all files recursively within a shared folder,
// with folder-key-encrypted file keys (from folder_file_keys). Used for ZIP download.
func GetSharedFolderTreeHandler(c *gin.Context, db *bun.DB) {
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
	if !share.PermDownload {
		c.JSON(http.StatusForbidden, gin.H{"error": "Download not permitted"})
		return
	}

	// Verify the requested folder is within the share
	var targetFolder pkg.Folder
	if err := db.NewSelect().Model(&targetFolder).
		Where("id = ? AND user_id = ?", folderID, rootFolder.UserID).
		Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}

	if targetFolder.ID != rootFolder.ID && !isWithinShare(rootFolder.Path, targetFolder.Path) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Folder not within shared directory"})
		return
	}

	// Fetch all files recursively
	var files []pkg.File
	if err := db.NewSelect().Model(&files).
		Where("user_id = ?", rootFolder.UserID).
		Where("path LIKE ?", targetFolder.Path+"/%").
		Where("is_preview = ?", false).
		Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list files"})
		return
	}

	// Enrich with folder-key-encrypted file keys
	filesWithKeys := enrichFilesWithKeys(c.Request.Context(), db, share.FolderID, files)

	// Strip keys for files that have none (security: do not leak owner keys)
	type fileEntry struct {
		ID           int64  `json:"id"`
		Name         string `json:"name"`
		Path         string `json:"path"`
		Size         int64  `json:"size"`
		MimeType     string `json:"mime_type"`
		EncryptedKey string `json:"encrypted_key"`
	}

	result := make([]fileEntry, 0, len(filesWithKeys))
	for _, f := range filesWithKeys {
		result = append(result, fileEntry{
			ID:           f.ID,
			Name:         f.Name,
			Path:         f.Path,
			Size:         f.Size,
			MimeType:     f.MimeType,
			EncryptedKey: f.EncryptedKey,
		})
	}

	// Relative path prefix for ZIP structure (strip the root folder path)
	c.JSON(http.StatusOK, gin.H{
		"files":       result,
		"root_path":   targetFolder.Path,
		"folder_name": targetFolder.Name,
	})
}

// sharedGeneratePresignedParts generates presigned PUT URLs for a shared multipart upload.
func sharedGeneratePresignedParts(ctx context.Context, presigner *s3.PresignClient, s3Key, uploadID string, totalParts int, totalSize int64) ([]map[string]interface{}, error) {
	partSize := (totalSize + int64(totalParts) - 1) / int64(totalParts)
	if partSize < sharedMinPart && totalParts > 1 {
		partSize = sharedMinPart
	}
	if partSize > sharedMaxPart {
		partSize = sharedMaxPart
	}

	remainingSize := totalSize
	result := make([]map[string]interface{}, 0, totalParts)

	for i := 1; i <= totalParts; i++ {
		thisPartSize := partSize
		if remainingSize < partSize {
			thisPartSize = remainingSize
		}
		remainingSize -= thisPartSize

		req, err := presigner.PresignUploadPart(ctx, &s3.UploadPartInput{
			Bucket:     aws.String(s3storage.BucketName),
			Key:        aws.String(s3Key),
			UploadId:   aws.String(uploadID),
			PartNumber: aws.Int32(int32(i)),
		}, func(opts *s3.PresignOptions) {
			opts.Expires = sharedPresignTTL
		})
		if err != nil {
			_, _ = s3storage.Client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
				Bucket:   aws.String(s3storage.BucketName),
				Key:      aws.String(s3Key),
				UploadId: aws.String(uploadID),
			})
			return nil, err
		}
		result = append(result, map[string]interface{}{
			"part_number": i,
			"url":         req.URL,
		})
	}
	return result, nil
}

