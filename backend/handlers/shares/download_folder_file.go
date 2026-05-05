// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package shares

import (
	"fmt"
	"io"
	"kagibi/backend/pkg"
	"kagibi/backend/pkg/monitoring"
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

func DownloadFileFromSharedFolderHandler(c *gin.Context, db *bun.DB) {
	token := c.Param("token")
	fileIDStr := c.Param("file_id")
	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	var shareLink pkg.ShareLink
	err = db.NewSelect().Model(&shareLink).
		Where("token = ?", token).
		Scan(c.Request.Context())

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
		return
	}

	if shareLink.ResourceType != "folder" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not a folder share"})
		return
	}

	if shareLink.ExpiresAt != nil && shareLink.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusGone, gin.H{"error": "Link expired"})
		return
	}

	if shareLink.SingleUse && shareLink.UsedAt != nil {
		c.JSON(http.StatusGone, gin.H{"error": "Link already used"})
		return
	}

	if !checkSharePassword(c, &shareLink) {
		return
	}

	if !shareLink.PermDownload {
		c.JSON(http.StatusForbidden, gin.H{"error": "Download not permitted on this share"})
		return
	}

	var file pkg.File
	err = db.NewSelect().Model(&file).
		Where("id = ?", fileID).
		Scan(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Security check: Ensure the file belongs to the owner AND is within the shared folder path
	if file.UserID != shareLink.OwnerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if !strings.HasPrefix(file.Path, shareLink.Path) {
		c.JSON(http.StatusForbidden, gin.H{"error": "File is not in the shared folder"})
		return
	}

	// Load all overrides for the share to check access and download permissions
	var overrides []pkg.ShareItemOverride
	_ = db.NewSelect().Model(&overrides).
		Where("share_id = ?", shareLink.ID).
		Scan(c.Request.Context())

	overrideMap := make(map[string]pkg.ShareItemOverride, len(overrides))
	for _, o := range overrides {
		overrideMap[o.ItemPath] = o
	}

	// Reject if the file itself or any ancestor folder is set to 'none'
	for _, o := range overrides {
		if o.AccessLevel == "none" {
			if strings.HasPrefix(file.Path, o.ItemPath+"/") || file.Path == o.ItemPath {
				c.JSON(http.StatusForbidden, gin.H{"error": "Access to this file is restricted"})
				return
			}
		}
	}

	// Reject if the file itself or any ancestor folder has can_download = false
	if !effectiveCanDownload(overrideMap, file.Path, true) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Download not permitted for this file"})
		return
	}

	// For single-use links, atomically mark as used before streaming.
	if shareLink.SingleUse {
		marked, err := markShareLinkUsed(c.Request.Context(), db, shareLink.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process link"})
			return
		}
		if !marked {
			c.JSON(http.StatusGone, gin.H{"error": "Link already used"})
			return
		}
	}

	// S3 Key construction
	s3Key := fmt.Sprintf("users/%s%s", shareLink.OwnerID, file.Path)

	// Get object from S3
	output, err := s3storage.Client.GetObject(c.Request.Context(), &s3.GetObjectInput{
		Bucket: aws.String(s3storage.BucketName),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		log.Printf("Error getting shared file from S3. Key: %s, Error: %v", s3Key, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve file from storage"})
		return
	}
	defer output.Body.Close()

	monitoring.FileDownloadsTotal.Inc()

	// Headers
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")

	disposition := "attachment"
	if c.Query("inline") == "true" {
		disposition = "inline"
	}
	c.Header("Content-Disposition", disposition+"; filename=\""+file.Name+"\"")

	c.Header("Content-Type", "application/octet-stream")
	if file.MimeType != "" {
		c.Header("Content-Type", file.MimeType)
	}
	c.Header("Content-Length", strconv.FormatInt(file.Size, 10))

	// Stream
	_, err = io.Copy(c.Writer, output.Body)
	if err != nil {
		log.Printf("Error streaming file: %v", err)
	}
}
