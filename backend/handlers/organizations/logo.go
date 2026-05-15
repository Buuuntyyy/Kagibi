// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package organizations

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"kagibi/backend/pkg/s3storage"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const maxLogoBytes = 2 << 20 // 2 MB

// UploadOrgLogo handles PUT /:orgID/logo — multipart image upload.
// Only org owners and admins may update the logo.
func (h *OrgHandler) UploadOrgLogo(c *gin.Context) {
	userID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid org id"})
		return
	}
	ctx := c.Request.Context()

	role, err := h.memberRole(ctx, orgID, userID)
	if err != nil || !canManage(role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only org admins can update the logo"})
		return
	}

	if s3storage.Client == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "storage not configured"})
		return
	}

	// Limit body to avoid memory exhaustion before parsing
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxLogoBytes+512)
	if err := c.Request.ParseMultipartForm(maxLogoBytes + 512); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid multipart form or file too large"})
		return
	}

	f, fh, err := c.Request.FormFile("logo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "logo file required (field name: logo)"})
		return
	}
	defer f.Close()

	if fh.Size > maxLogoBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "logo must be under 2 MB"})
		return
	}

	// Validate content type
	ct := fh.Header.Get("Content-Type")
	if ct == "" {
		ct = "image/png"
	}
	if !isLogoMIME(ct) {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "logo must be jpeg, png, gif, webp or svg"})
		return
	}

	ext := mimeExt(ct, fh.Filename)
	key := fmt.Sprintf("org-logos/%d/%s%s", orgID, uuid.New().String(), ext)

	body, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read logo"})
		return
	}

	// Fetch old logo key to delete after successful upload
	var oldKey string
	_ = h.DB.NewSelect().TableExpr("organizations").
		ColumnExpr("logo_path").
		Where("id = ?", orgID).
		Scan(ctx, &oldKey)

	// Upload to S3
	_, err = s3storage.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s3storage.BucketName),
		Key:           aws.String(key),
		Body:          bytes.NewReader(body),
		ContentType:   aws.String(ct),
		ContentLength: aws.Int64(int64(len(body))),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload logo"})
		return
	}

	// Persist the new key
	if _, err := h.DB.NewUpdate().TableExpr("organizations").
		Set("logo_path = ?", key).
		Set("updated_at = now()").
		Where("id = ?", orgID).
		Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save logo"})
		return
	}

	// Remove the old logo object (best-effort; don't fail the request)
	if oldKey != "" {
		_, _ = s3storage.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(s3storage.BucketName),
			Key:    aws.String(oldKey),
		})
	}

	logoURL := presignLogoURL(ctx, key)
	h.logAudit(ctx, orgID, userID, "logo_updated", strconv.FormatInt(orgID, 10), "organization", "")

	c.JSON(http.StatusOK, gin.H{"logo_url": logoURL, "logo_path": key})
}

// DeleteOrgLogo handles DELETE /:orgID/logo — removes the custom logo.
func (h *OrgHandler) DeleteOrgLogo(c *gin.Context) {
	userID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid org id"})
		return
	}
	ctx := c.Request.Context()

	role, err := h.memberRole(ctx, orgID, userID)
	if err != nil || !canManage(role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only org admins can remove the logo"})
		return
	}

	if s3storage.Client == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "storage not configured"})
		return
	}

	var logoKey string
	if err := h.DB.NewSelect().TableExpr("organizations").
		ColumnExpr("logo_path").
		Where("id = ?", orgID).
		Scan(ctx, &logoKey); err != nil || logoKey == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "no logo to remove"})
		return
	}

	// Delete from S3 (best-effort)
	_, _ = s3storage.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s3storage.BucketName),
		Key:    aws.String(logoKey),
	})

	if _, err := h.DB.NewUpdate().TableExpr("organizations").
		Set("logo_path = ''").
		Set("updated_at = now()").
		Where("id = ?", orgID).
		Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to clear logo"})
		return
	}

	h.logAudit(ctx, orgID, userID, "logo_removed", strconv.FormatInt(orgID, 10), "organization", "")
	c.JSON(http.StatusOK, gin.H{"message": "logo removed"})
}

func isLogoMIME(ct string) bool {
	switch ct {
	case "image/jpeg", "image/jpg", "image/png", "image/gif", "image/webp", "image/svg+xml":
		return true
	}
	return false
}

func mimeExt(ct, filename string) string {
	switch ct {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/svg+xml":
		return ".svg"
	}
	// Fallback: use filename extension
	if idx := len(filename) - 1; idx >= 0 {
		for i := idx; i >= 0; i-- {
			if filename[i] == '.' {
				return filename[i:]
			}
		}
	}
	return ".png"
}
