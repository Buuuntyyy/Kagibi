// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package organizations

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
)

const orgPresignTTL = 180 * time.Second

type OrgInitiateRequest struct {
	FileName     string `json:"file_name" binding:"required"`
	FilePath     string `json:"file_path"` // destination folder path within org (e.g. "/documents")
	ContentType  string `json:"content_type"`
	TotalSize    int64  `json:"total_size" binding:"required,min=1"`
	TotalParts   int    `json:"total_parts" binding:"required,min=1"`
	EncryptedKey string `json:"encrypted_key" binding:"required"`
}

type OrgInitiateResponse struct {
	UploadID      string             `json:"upload_id"`
	Key           string             `json:"key"`
	PresignedURLs []orgPresignedPart `json:"presigned_urls"`
}

type orgPresignedPart struct {
	PartNumber int    `json:"part_number"`
	URL        string `json:"url"`
}

type OrgCompleteRequest struct {
	UploadID     string    `json:"upload_id" binding:"required"`
	Key          string    `json:"key" binding:"required"`
	Parts        []orgPart `json:"parts" binding:"required,min=1"`
	FileName     string    `json:"file_name" binding:"required"`
	FilePath     string    `json:"file_path"`
	TotalSize    int64     `json:"total_size" binding:"required"`
	ContentType  string    `json:"content_type"`
	EncryptedKey string    `json:"encrypted_key" binding:"required"`
	// GroupID is non-nil when the file key was wrapped with a group key instead of the org key.
	GroupID      *int64    `json:"group_id,omitempty"`
}

type orgPart struct {
	PartNumber int    `json:"part_number"`
	ETag       string `json:"etag"`
}

type OrgAbortRequest struct {
	UploadID string `json:"upload_id" binding:"required"`
	Key      string `json:"key" binding:"required"`
}

// InitiateOrgMultipart starts a multipart S3 upload for an org file.
func (h *OrgHandler) InitiateOrgMultipart(c *gin.Context) {
	userID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}

	var req OrgInitiateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !h.checkMFAEnforcement(c, orgID, userID) {
		return
	}

	folderPath := normPath(req.FilePath)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// Permission check
	perm, err := h.resolvePermission(ctx, orgID, userID, folderPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check permissions"})
		return
	}
	if perm < PermWrite {
		c.JSON(http.StatusForbidden, gin.H{"error": "write access denied"})
		return
	}

	// Org-level quota check
	var org pkg.Organization
	if err := h.DB.NewSelect().Model(&org).Where("id = ?", orgID).Scan(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch organization"})
		return
	}
	quotaBytes := org.StorageQuotaMB * 1024 * 1024
	if org.StorageUsedBytes+req.TotalSize > quotaBytes {
		c.JSON(http.StatusForbidden, gin.H{"error": "organization storage quota exceeded"})
		return
	}

	// Per-member quota check (quota_bytes=0 means unlimited)
	var member pkg.OrgMember
	if err := h.DB.NewSelect().Model(&member).
		Where("org_id = ? AND user_id = ?", orgID, userID).
		Scan(ctx); err == nil && member.QuotaBytes > 0 {
		var memberUsed int64
		_ = h.DB.NewSelect().
			TableExpr("org_files").
			ColumnExpr("COALESCE(SUM(size), 0)").
			Where("org_id = ? AND uploaded_by = ? AND deleted_at IS NULL", orgID, userID).
			Scan(ctx, &memberUsed)
		if memberUsed+req.TotalSize > member.QuotaBytes {
			c.JSON(http.StatusForbidden, gin.H{"error": "member storage quota exceeded"})
			return
		}
	}

	fullPath := normPath(path.Join(folderPath, req.FileName))
	s3Key := orgS3Key(orgID, fullPath)

	contentType := req.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	createOut, err := s3storage.Client.CreateMultipartUpload(ctx, &s3.CreateMultipartUploadInput{
		Bucket:      aws.String(s3storage.BucketName),
		Key:         aws.String(s3Key),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		log.Printf("[OrgMultipart] CreateMultipartUpload error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initiate upload"})
		return
	}

	uploadID := *createOut.UploadId
	presigner := s3.NewPresignClient(s3storage.Client)

	urls, err := orgPresignParts(ctx, presigner, s3Key, uploadID, req.TotalParts, req.TotalSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate presigned URLs"})
		return
	}

	c.JSON(http.StatusOK, OrgInitiateResponse{
		UploadID:      uploadID,
		Key:           s3Key,
		PresignedURLs: urls,
	})
}

// CompleteOrgMultipart finalises the S3 upload and records the file in the DB.
func (h *OrgHandler) CompleteOrgMultipart(c *gin.Context) {
	userID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}

	var req OrgCompleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !h.checkMFAEnforcement(c, orgID, userID) {
		return
	}

	folderPath := normPath(req.FilePath)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	// Validate S3 key belongs to this org (prevent key injection)
	expectedPrefix := fmt.Sprintf("orgs/%d/", orgID)
	if !strings.HasPrefix(req.Key, expectedPrefix) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid upload key"})
		return
	}

	// Final quota check (re-validate in case concurrent uploads hit the limit)
	var org pkg.Organization
	if err := h.DB.NewSelect().Model(&org).Where("id = ?", orgID).Scan(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch organization"})
		return
	}
	if org.StorageUsedBytes+req.TotalSize > org.StorageQuotaMB*1024*1024 {
		// Abort the dangling S3 upload
		_, _ = s3storage.Client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
			Bucket: aws.String(s3storage.BucketName), Key: aws.String(req.Key), UploadId: aws.String(req.UploadID),
		})
		c.JSON(http.StatusForbidden, gin.H{"error": "organization storage quota exceeded"})
		return
	}

	sort.Slice(req.Parts, func(i, j int) bool { return req.Parts[i].PartNumber < req.Parts[j].PartNumber })
	completedParts := make([]types.CompletedPart, len(req.Parts))
	for i, p := range req.Parts {
		etag := p.ETag
		completedParts[i] = types.CompletedPart{
			PartNumber: aws.Int32(int32(p.PartNumber)),
			ETag:       aws.String(etag),
		}
	}

	if _, err := s3storage.Client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(s3storage.BucketName),
		Key:      aws.String(req.Key),
		UploadId: aws.String(req.UploadID),
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: completedParts,
		},
	}); err != nil {
		log.Printf("[OrgMultipart] CompleteMultipartUpload error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to complete upload"})
		return
	}

	fullPath := normPath(path.Join(folderPath, req.FileName))
	contentType := req.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// If a group_id is provided, verify the caller is a member of that group.
	if req.GroupID != nil {
		var count int
		if err := h.DB.NewSelect().TableExpr("org_group_members").
			ColumnExpr("COUNT(*)").
			Where("group_id = ? AND user_id = ?", *req.GroupID, userID).
			Scan(ctx, &count); err != nil || count == 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "you are not a member of the specified group"})
			return
		}
	}

	file := &pkg.OrgFile{
		OrgID:        orgID,
		Name:         req.FileName,
		Path:         fullPath,
		FolderPath:   folderPath,
		Size:         req.TotalSize,
		MimeType:     contentType,
		UploadedBy:   userID,
		EncryptedKey: req.EncryptedKey,
		GroupID:      req.GroupID,
		TagIDs:       []int64{},
	}
	if _, err := h.DB.NewInsert().Model(file).
		On("CONFLICT (org_id, path) WHERE deleted_at IS NULL DO UPDATE").
		Set("size = EXCLUDED.size, mime_type = EXCLUDED.mime_type, encrypted_key = EXCLUDED.encrypted_key, group_id = EXCLUDED.group_id, uploaded_by = EXCLUDED.uploaded_by, updated_at = NOW(), deleted_at = NULL").
		Exec(ctx); err != nil {
		log.Printf("[OrgMultipart] DB insert error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record file"})
		return
	}

	_, _ = h.DB.NewUpdate().Model((*pkg.Organization)(nil)).
		Set("storage_used_bytes = storage_used_bytes + ?", req.TotalSize).
		Where("id = ?", orgID).
		Exec(ctx)

	h.logAudit(ctx, orgID, userID, "file_uploaded", strconv.FormatInt(file.ID, 10), "file", file.Name)
	c.JSON(http.StatusCreated, gin.H{"file": file})
}

// AbortOrgMultipart cancels an in-progress S3 multipart upload.
func (h *OrgHandler) AbortOrgMultipart(c *gin.Context) {
	userID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}

	var req OrgAbortRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	expectedPrefix := fmt.Sprintf("orgs/%d/", orgID)
	if !strings.HasPrefix(req.Key, expectedPrefix) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid upload key"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	// Derive the folder path from the S3 key and enforce write permission.
	// The key format is orgs/{orgID}{fullFilePath}, e.g. orgs/7/documents/report.pdf
	filePath := strings.TrimPrefix(req.Key, fmt.Sprintf("orgs/%d", orgID))
	folderPath := path.Dir(normPath(filePath))
	perm, err := h.resolvePermission(ctx, orgID, userID, folderPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check permissions"})
		return
	}
	if perm < PermWrite {
		c.JSON(http.StatusForbidden, gin.H{"error": "write access required to abort an upload"})
		return
	}

	if _, err := s3storage.Client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(s3storage.BucketName),
		Key:      aws.String(req.Key),
		UploadId: aws.String(req.UploadID),
	}); err != nil {
		log.Printf("[OrgMultipart] AbortMultipartUpload error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to abort upload"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "upload aborted"})
}

func orgPresignParts(ctx context.Context, presigner *s3.PresignClient, key, uploadID string, totalParts int, totalSize int64) ([]orgPresignedPart, error) {
	const maxPartSize int64 = 100 * 1024 * 1024
	const minPartSize int64 = 5 * 1024 * 1024

	partSize := (totalSize + int64(totalParts) - 1) / int64(totalParts)
	if partSize < minPartSize && totalParts > 1 {
		partSize = minPartSize
	}
	if partSize > maxPartSize {
		partSize = maxPartSize
	}

	remaining := totalSize
	parts := make([]orgPresignedPart, 0, totalParts)
	for i := 1; i <= totalParts; i++ {
		thisSize := partSize
		if remaining < partSize {
			thisSize = remaining
		}
		remaining -= thisSize

		presigned, err := presigner.PresignUploadPart(ctx, &s3.UploadPartInput{
			Bucket:     aws.String(s3storage.BucketName),
			Key:        aws.String(key),
			UploadId:   aws.String(uploadID),
			PartNumber: aws.Int32(int32(i)),
		}, func(o *s3.PresignOptions) { o.Expires = orgPresignTTL })
		if err != nil {
			_, _ = s3storage.Client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
				Bucket: aws.String(s3storage.BucketName), Key: aws.String(key), UploadId: aws.String(uploadID),
			})
			return nil, err
		}
		parts = append(parts, orgPresignedPart{PartNumber: i, URL: presigned.URL})
	}
	return parts, nil
}
