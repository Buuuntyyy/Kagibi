// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package shares

import (
	"context"
	"fmt"
	"kagibi/backend/pkg"
	"kagibi/backend/pkg/s3storage"
	"log"
	"net/http"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// resolvePublicShare looks up a share link by token, verifies it hasn't expired,
// and confirms perm_create is set. Returns (shareLink, ownerID, ok).
func resolvePublicShareForWrite(c *gin.Context, db *bun.DB, permCheck func(*pkg.ShareLink) bool, permName string) (*pkg.ShareLink, bool) {
	token := c.Param("token")

	var shareLink pkg.ShareLink
	if err := db.NewSelect().Model(&shareLink).Where("token = ?", token).Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Share not found"})
		return nil, false
	}

	if shareLink.ExpiresAt != nil && shareLink.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusGone, gin.H{"error": "Share link expired"})
		return nil, false
	}

	if shareLink.ResourceType != "folder" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not a folder share"})
		return nil, false
	}

	if !permCheck(&shareLink) {
		c.JSON(http.StatusForbidden, gin.H{"error": permName + " not permitted on this share"})
		return nil, false
	}

	return &shareLink, true
}

// CreateFolderInPublicShareHandler creates a subfolder inside a publicly shared folder.
// Requires perm_create on the share link.
func CreateFolderInPublicShareHandler(c *gin.Context, db *bun.DB) {
	shareLink, ok := resolvePublicShareForWrite(c, db, func(s *pkg.ShareLink) bool { return s.PermCreate }, "Create")
	if !ok {
		return
	}

	var req struct {
		Name       string `json:"name"`
		ParentPath string `json:"parent_path"` // path relative to share root, e.g. "/" or "/subdir"
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	req.Name = strings.ReplaceAll(strings.ReplaceAll(req.Name, "\n", "_"), "\r", "_")

	// Resolve absolute parent path
	parentPath := shareLink.Path
	if req.ParentPath != "" && req.ParentPath != "/" {
		candidate := shareLink.Path + "/" + strings.Trim(req.ParentPath, "/")
		candidate = strings.ReplaceAll(strings.ReplaceAll(candidate, "\n", "_"), "\r", "_")
		if !strings.HasPrefix(candidate, shareLink.Path) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Parent path not within shared directory"})
			return
		}
		parentPath = candidate
	}

	newPath := parentPath + "/" + req.Name
	folder := &pkg.Folder{
		Name:      req.Name,
		Path:      newPath,
		UserID:    shareLink.OwnerID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if _, err := db.NewInsert().Model(folder).Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create folder"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": folder.ID, "name": folder.Name, "path": folder.Path})
}

// RenameInPublicShareHandler renames a file or folder within a publicly shared folder.
// Requires perm_move on the share link.
func RenameInPublicShareHandler(c *gin.Context, db *bun.DB) {
	shareLink, ok := resolvePublicShareForWrite(c, db, func(s *pkg.ShareLink) bool { return s.PermMove }, "Rename")
	if !ok {
		return
	}

	var req struct {
		ID      int64  `json:"id"`
		Type    string `json:"type"`
		NewName string `json:"new_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ID == 0 || strings.TrimSpace(req.NewName) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id, type, and new_name are required"})
		return
	}
	req.NewName = strings.ReplaceAll(strings.ReplaceAll(req.NewName, "\n", "_"), "\r", "_")

	switch req.Type {
	case "file":
		var file pkg.File
		if err := db.NewSelect().Model(&file).
			Where("id = ? AND user_id = ?", req.ID, shareLink.OwnerID).
			Scan(c.Request.Context()); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}
		if !strings.HasPrefix(file.Path, shareLink.Path) {
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
			Where("id = ? AND user_id = ?", req.ID, shareLink.OwnerID).
			Scan(c.Request.Context()); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
			return
		}
		if !strings.HasPrefix(folder.Path, shareLink.Path) || folder.Path == shareLink.Path {
			c.JSON(http.StatusForbidden, gin.H{"error": "Folder not within shared directory or is the share root"})
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

		if _, err := tx.NewUpdate().Model((*pkg.Folder)(nil)).
			Set("path = REPLACE(path, ?, ?)", oldPath+"/", newPath+"/").
			Where("user_id = ? AND path LIKE ?", shareLink.OwnerID, oldPath+"/%").
			Exec(c.Request.Context()); err != nil {
			log.Printf("Warning: failed to update subfolder paths: %v", err)
		}
		if _, err := tx.NewUpdate().Model((*pkg.File)(nil)).
			Set("path = REPLACE(path, ?, ?)", oldPath+"/", newPath+"/").
			Where("user_id = ? AND path LIKE ?", shareLink.OwnerID, oldPath+"/%").
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

type publicInitiateRequest struct {
	FileName     string `json:"file_name" binding:"required"`
	FilePath     string `json:"file_path"`
	ContentType  string `json:"content_type"`
	TotalSize    int64  `json:"total_size" binding:"required,min=1"`
	TotalParts   int    `json:"total_parts" binding:"required,min=1"`
	EncryptedKey string `json:"encrypted_key" binding:"required"`
}

type publicCompleteRequest struct {
	UploadID     string              `json:"upload_id" binding:"required"`
	Key          string              `json:"key" binding:"required"`
	Parts        []sharedCompletePart `json:"parts" binding:"required,min=1"`
	FileName     string              `json:"file_name" binding:"required"`
	FilePath     string              `json:"file_path"`
	TotalSize    int64               `json:"total_size" binding:"required"`
	ContentType  string              `json:"content_type"`
	EncryptedKey string              `json:"encrypted_key" binding:"required"`
}

type publicAbortRequest struct {
	UploadID string `json:"upload_id" binding:"required"`
	Key      string `json:"key" binding:"required"`
}

// resolveAbsolutePath computes the absolute folder path from the share root and a client-supplied
// relative path (e.g. "/" or "/subdir"). Returns (absolutePath, ok).
func resolveAbsolutePath(c *gin.Context, shareLink *pkg.ShareLink, relativePath string) (string, bool) {
	absPath := shareLink.Path
	if relativePath != "" && relativePath != "/" {
		candidate := shareLink.Path + "/" + strings.Trim(relativePath, "/")
		candidate = strings.ReplaceAll(strings.ReplaceAll(candidate, "\n", "_"), "\r", "_")
		if !strings.HasPrefix(candidate, shareLink.Path) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Path not within shared directory"})
			return "", false
		}
		absPath = candidate
	}
	return absPath, true
}

// InitiatePublicShareUploadHandler starts a multipart upload within a publicly shared folder.
// Requires perm_create on the share link.
func InitiatePublicShareUploadHandler(c *gin.Context, db *bun.DB) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	shareLink, ok := resolvePublicShareForWrite(c, db, func(s *pkg.ShareLink) bool { return s.PermCreate }, "Create")
	if !ok {
		return
	}

	var req publicInitiateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	absPath, ok := resolveAbsolutePath(c, shareLink, req.FilePath)
	if !ok {
		return
	}

	fileName := strings.ReplaceAll(strings.ReplaceAll(req.FileName, "\n", "_"), "\r", "_")
	fullPath := path.Join(absPath, fileName)
	if !strings.HasPrefix(fullPath, "/") {
		fullPath = "/" + fullPath
	}
	s3Key := fmt.Sprintf("users/%s%s", shareLink.OwnerID, fullPath)

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
		log.Printf("S3 CreateMultipartUpload (public share) error: %v", err)
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

// CompletePublicShareUploadHandler completes a multipart upload within a publicly shared folder.
// Creates the File record (owned by the share owner) and a ShareFileKey entry.
func CompletePublicShareUploadHandler(c *gin.Context, db *bun.DB) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	shareLink, ok := resolvePublicShareForWrite(c, db, func(s *pkg.ShareLink) bool { return s.PermCreate }, "Create")
	if !ok {
		return
	}

	var req publicCompleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Security: S3 key must be under the owner's namespace
	expectedPrefix := fmt.Sprintf("users/%s/", shareLink.OwnerID)
	if !strings.HasPrefix(req.Key, expectedPrefix) {
		log.Printf("SECURITY: public share upload key %s not under owner %s", req.Key, shareLink.OwnerID)
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

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

	if _, err := s3storage.Client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(s3storage.BucketName),
		Key:      aws.String(req.Key),
		UploadId: aws.String(req.UploadID),
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: completedParts,
		},
	}); err != nil {
		log.Printf("S3 CompleteMultipartUpload (public share) error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete upload"})
		return
	}

	ownerPrefix := fmt.Sprintf("users/%s", shareLink.OwnerID)
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
		UserID:       shareLink.OwnerID,
		EncryptedKey: "", // key is stored in share_file_keys
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
		log.Printf("upsertSharedFile (public) error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file record"})
		return
	}

	shareFileKey := &pkg.ShareFileKey{
		ShareID:      shareLink.ID,
		FileID:       fileRecord.ID,
		EncryptedKey: req.EncryptedKey,
	}
	if _, err := tx.NewInsert().Model(shareFileKey).
		On("CONFLICT (share_id, file_id) DO UPDATE").
		Set("encrypted_key = EXCLUDED.encrypted_key").
		Exec(ctx); err != nil {
		log.Printf("Warning: failed to store ShareFileKey (public upload): %v", err)
	}

	if delta != 0 {
		if _, err := tx.NewUpdate().Model((*pkg.UserPlan)(nil)).
			Set("storage_used = storage_used + ?", delta).
			Where("user_id = ?", shareLink.OwnerID).
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

// AbortPublicShareUploadHandler cancels an in-progress public share multipart upload.
func AbortPublicShareUploadHandler(c *gin.Context, db *bun.DB) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	shareLink, ok := resolvePublicShareForWrite(c, db, func(s *pkg.ShareLink) bool { return s.PermCreate }, "Create")
	if !ok {
		return
	}

	var req publicAbortRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	expectedPrefix := fmt.Sprintf("users/%s/", shareLink.OwnerID)
	if !strings.HasPrefix(req.Key, expectedPrefix) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if _, err := s3storage.Client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(s3storage.BucketName),
		Key:      aws.String(req.Key),
		UploadId: aws.String(req.UploadID),
	}); err != nil {
		log.Printf("S3 AbortMultipartUpload (public share) error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to abort upload"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Upload aborted"})
}
