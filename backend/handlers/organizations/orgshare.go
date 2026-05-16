// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package organizations

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"mime"
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

type createOrgShareRequest struct {
	EncryptedKey string     `json:"encrypted_key"`
	ExpiresAt    *time.Time `json:"expires_at"`
}

// CreateOrgFileShare creates a public share link for an org file.
// POST /orgs/:orgID/fs/file/:fileID/share
func (h *OrgHandler) CreateOrgFileShare(c *gin.Context) {
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

	var req createOrgShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if req.EncryptedKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "encrypted_key is required"})
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
		c.JSON(http.StatusForbidden, gin.H{"error": "read access denied"})
		return
	}

	token, err := generateOrgShareToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	share := &pkg.ShareLink{
		ResourceID:   fileID,
		ResourceType: "org_file",
		Path:         file.Path,
		OwnerID:      userID,
		Token:        token,
		EncryptedKey: req.EncryptedKey,
		ExpiresAt:    req.ExpiresAt,
		PermDownload: true,
		OrgID:        &orgID,
	}

	if _, err := h.DB.NewInsert().Model(share).Exec(ctx); err != nil {
		log.Printf("CreateOrgFileShare insert error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create share"})
		return
	}

	h.logAudit(ctx, orgID, userID, "file_shared_public", strconv.FormatInt(fileID, 10), "file", file.Name)

	c.JSON(http.StatusCreated, gin.H{
		"token": share.Token,
		"link":  fmt.Sprintf("/s/org/%s", share.Token),
	})
}

// GetOrgShare is a public (unauthenticated) endpoint returning file metadata + encrypted_key.
// GET /api/v1/public/org-share/:token
func GetOrgShare(c *gin.Context, db *bun.DB) {
	token := c.Param("token")

	share, err := fetchValidOrgShare(c.Request.Context(), db, token)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var file pkg.OrgFile
	if err := db.NewSelect().Model(&file).
		Where("id = ? AND org_id = ?", share.ResourceID, *share.OrgID).
		Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	go func() {
		_, _ = db.NewUpdate().Model((*pkg.ShareLink)(nil)).
			Set("views = views + 1").
			Where("id = ?", share.ID).
			Exec(context.Background())
	}()

	c.JSON(http.StatusOK, gin.H{
		"resource_name": file.Name,
		"file_size":     file.Size,
		"mime_type":     file.MimeType,
		"encrypted_key": share.EncryptedKey,
		"expires_at":    share.ExpiresAt,
	})
}

// DownloadOrgShare is a public (unauthenticated) endpoint that streams the encrypted file from S3.
// GET /api/v1/public/org-share/:token/download
func DownloadOrgShare(c *gin.Context, db *bun.DB) {
	token := c.Param("token")

	share, err := fetchValidOrgShare(c.Request.Context(), db, token)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var file pkg.OrgFile
	if err := db.NewSelect().Model(&file).
		Where("id = ? AND org_id = ?", share.ResourceID, *share.OrgID).
		Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	s3Key := orgS3Key(*share.OrgID, file.Path)
	output, err := s3storage.Client.GetObject(c.Request.Context(), &s3.GetObjectInput{
		Bucket: aws.String(s3storage.BucketName),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		log.Printf("DownloadOrgShare S3 error token=%s: %v", token, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve file"})
		return
	}
	defer output.Body.Close()

	contentType := file.MimeType
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	cd := mime.FormatMediaType("attachment", map[string]string{"filename": file.Name})
	if cd == "" {
		cd = "attachment"
	}
	c.Header("Content-Type", contentType)
	c.Header("Content-Length", strconv.FormatInt(file.Size, 10))
	c.Header("Content-Disposition", cd)

	if _, err := io.Copy(c.Writer, output.Body); err != nil {
		log.Printf("DownloadOrgShare stream error token=%s: %v", token, err)
	}
}

// ── helpers ──────────────────────────────────────────────────────────────────

func generateOrgShareToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func fetchValidOrgShare(ctx context.Context, db *bun.DB, token string) (*pkg.ShareLink, error) {
	var share pkg.ShareLink
	if err := db.NewSelect().Model(&share).
		Where("token = ? AND resource_type = 'org_file'", token).
		Scan(ctx); err != nil {
		return nil, fmt.Errorf("share not found")
	}
	if share.OrgID == nil {
		return nil, fmt.Errorf("share not found")
	}
	if share.ExpiresAt != nil && share.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("link expired")
	}
	return &share, nil
}
