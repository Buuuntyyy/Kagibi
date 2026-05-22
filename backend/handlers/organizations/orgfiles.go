// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package organizations

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"path"
	"strconv"
	"time"

	"kagibi/backend/pkg"
	"kagibi/backend/pkg/s3storage"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

// orgS3Key returns the S3 object key for an org file.
func orgS3Key(orgID int64, filePath string) string {
	return fmt.Sprintf("orgs/%d%s", orgID, filePath)
}

// DownloadOrgFile streams an org file from S3.
func (h *OrgHandler) DownloadOrgFile(c *gin.Context) {
	userID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}
	fileID, err := strconv.ParseInt(c.Param("fileID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file id"})
		return
	}

	if !h.checkMFAEnforcement(c, orgID, userID) {
		return
	}

	ctx := c.Request.Context()

	var file pkg.OrgFile
	if err := h.DB.NewSelect().Model(&file).
		Where("id = ? AND org_id = ?", fileID, orgID).
		Scan(ctx); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch file"})
		return
	}

	perm, err := h.resolvePermission(ctx, orgID, userID, file.FolderPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check permissions"})
		return
	}
	if perm < PermRead {
		log.Printf("SECURITY: org file download denied - user=%s org=%d file=%d", userID, orgID, fileID)
		c.JSON(http.StatusForbidden, gin.H{"error": "read access denied"})
		return
	}

	allowed, err := h.resolveDownloadAllowed(ctx, orgID, userID, file.FolderPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check download permission"})
		return
	}
	if !allowed {
		log.Printf("SECURITY: org file download blocked by perm_download=false - user=%s org=%d file=%d", userID, orgID, fileID)
		c.JSON(http.StatusForbidden, gin.H{"error": "download access denied"})
		return
	}

	s3Key := orgS3Key(orgID, file.Path)
	output, err := s3storage.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s3storage.BucketName),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		log.Printf("S3 GetObject error for org file %d: %v", fileID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve file from storage"})
		return
	}
	defer output.Body.Close()

	disposition := "attachment"
	if c.Query("inline") == "true" {
		disposition = "inline"
	}
	contentType := file.MimeType
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	cd := mime.FormatMediaType(disposition, map[string]string{"filename": file.Name})
	if cd == "" {
		cd = disposition
	}
	c.Header("Content-Disposition", cd)
	c.Header("Content-Type", contentType)
	c.Header("Content-Length", strconv.FormatInt(file.Size, 10))

	h.logAudit(ctx, orgID, userID, "file_downloaded", strconv.FormatInt(fileID, 10), "file", file.Name)
	if _, err := io.Copy(c.Writer, output.Body); err != nil {
		log.Printf("Error streaming org file %d: %v", fileID, err)
	}
}

// DeleteOrgFile soft-deletes a file and adjusts the org quota.
func (h *OrgHandler) DeleteOrgFile(c *gin.Context) {
	userID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}
	fileID, err := strconv.ParseInt(c.Param("fileID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file id"})
		return
	}

	if !h.checkMFAEnforcement(c, orgID, userID) {
		return
	}

	ctx := c.Request.Context()

	var file pkg.OrgFile
	if err := h.DB.NewSelect().Model(&file).
		Where("id = ? AND org_id = ?", fileID, orgID).
		Scan(ctx); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch file"})
		return
	}

	perm, err := h.resolvePermission(ctx, orgID, userID, file.FolderPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check permissions"})
		return
	}
	if perm < PermWrite {
		c.JSON(http.StatusForbidden, gin.H{"error": "delete access denied"})
		return
	}

	now := time.Now().UTC()
	if _, err := h.DB.NewUpdate().Model((*pkg.OrgFile)(nil)).
		Set("deleted_at = ?, deleted_by = ?, delete_root = TRUE", now, userID).
		Where("id = ?", fileID).
		Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete file"})
		return
	}

	_, _ = h.DB.NewUpdate().Model((*pkg.Organization)(nil)).
		Set("storage_used_bytes = GREATEST(0, storage_used_bytes - ?)", file.Size).
		Where("id = ?", orgID).
		Exec(ctx)

	h.logAudit(ctx, orgID, userID, "file_deleted", strconv.FormatInt(fileID, 10), "file", file.Name)
	c.JSON(http.StatusOK, gin.H{"message": "file deleted", "freed_bytes": file.Size})
}

// GetOrgFileKey returns the encrypted_key for a specific org file.
// Used by the frontend to decrypt a file after downloading.
func (h *OrgHandler) GetOrgFileKey(c *gin.Context) {
	userID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}
	fileID, err := strconv.ParseInt(c.Param("fileID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file id"})
		return
	}

	if !h.checkMFAEnforcement(c, orgID, userID) {
		return
	}

	ctx := c.Request.Context()

	var file pkg.OrgFile
	if err := h.DB.NewSelect().Model(&file).
		Where("id = ? AND org_id = ?", fileID, orgID).
		Scan(ctx); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	perm, err := h.resolvePermission(ctx, orgID, userID, path.Dir(file.Path))
	if err != nil || perm < PermRead {
		c.JSON(http.StatusForbidden, gin.H{"error": "read access denied"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"encrypted_key": file.EncryptedKey})
}
