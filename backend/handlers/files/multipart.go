// backend/handlers/files/multipart.go
package files

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"safercloud/backend/pkg"
	"safercloud/backend/pkg/s3storage"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

const (
	PresignTTL         = 180 * time.Second // 3 minutes max per part
	DownloadPresignTTL = 5 * time.Minute   // 5 minutes for downloads
	MaxPartSize        = 100 * 1024 * 1024 // 100MB max per part
	MinPartSize        = 5 * 1024 * 1024   // 5MB min (S3 requirement except last part)
)

// InitiateMultipartRequest represents the request body for initiating multipart upload
type InitiateMultipartRequest struct {
	FileName     string `json:"file_name" binding:"required"`
	FilePath     string `json:"file_path"`
	ContentType  string `json:"content_type"`
	TotalSize    int64  `json:"total_size" binding:"required,min=1"`
	TotalParts   int    `json:"total_parts" binding:"required,min=1"`
	EncryptedKey string `json:"encrypted_key" binding:"required"`
}

// InitiateMultipartResponse contains upload ID and presigned URLs
type InitiateMultipartResponse struct {
	UploadID      string          `json:"upload_id"`
	Key           string          `json:"key"`
	PresignedURLs []PresignedPart `json:"presigned_urls"`
}

// PresignedPart represents a presigned URL for a single part
type PresignedPart struct {
	PartNumber int    `json:"part_number"`
	URL        string `json:"url"`
}

// CompleteMultipartRequest contains the parts info to complete upload
type CompleteMultipartRequest struct {
	UploadID     string         `json:"upload_id" binding:"required"`
	Key          string         `json:"key" binding:"required"`
	Parts        []CompletePart `json:"parts" binding:"required,min=1"`
	FileName     string         `json:"file_name" binding:"required"`
	FilePath     string         `json:"file_path"`
	TotalSize    int64          `json:"total_size" binding:"required"`
	ContentType  string         `json:"content_type"`
	EncryptedKey string         `json:"encrypted_key" binding:"required"`
	ShareKeys    string         `json:"share_keys"`
	PreviewID    *int64         `json:"preview_id"`
	IsPreview    bool           `json:"is_preview"`
}

// CompletePart represents a completed part with ETag
type CompletePart struct {
	PartNumber int    `json:"part_number"`
	ETag       string `json:"etag"`
}

// AbortMultipartRequest contains info to abort an upload
type AbortMultipartRequest struct {
	UploadID string `json:"upload_id" binding:"required"`
	Key      string `json:"key" binding:"required"`
}

// InitiateMultipartHandler starts a multipart upload and returns presigned URLs
func InitiateMultipartHandler(c *gin.Context, db *bun.DB) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req InitiateMultipartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Validate user exists and check quota
	var user pkg.User
	if err := db.NewSelect().Model(&user).Where("id = ?", userID).Scan(ctx); err != nil {
		log.Printf("SECURITY: Multipart init for unknown user: %s", userID)
		c.JSON(http.StatusForbidden, gin.H{"error": "User not found"})
		return
	}

	// Strict quota check
	if user.StorageUsed+req.TotalSize > user.StorageLimit {
		c.JSON(http.StatusForbidden, gin.H{"error": "Storage quota exceeded"})
		return
	}

	// Validate and sanitize path
	filePath := req.FilePath
	if filePath == "" {
		filePath = "/"
	}
	validPath, err := validatePath(filePath)
	if err != nil {
		log.Printf("SECURITY: Invalid path in multipart init - UserID: %s, Path: %s", userID, filePath)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid path"})
		return
	}

	// Construct S3 key
	fullPathDB := path.Join(validPath, req.FileName)
	fullPathDB = path.Clean(fullPathDB)
	if !strings.HasPrefix(fullPathDB, "/") {
		fullPathDB = "/" + fullPathDB
	}
	s3Key := fmt.Sprintf("users/%s%s", userID, fullPathDB)

	// Determine content type
	contentType := req.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Initiate multipart upload on S3
	createOutput, err := s3storage.Client.CreateMultipartUpload(ctx, &s3.CreateMultipartUploadInput{
		Bucket:      aws.String(s3storage.BucketName),
		Key:         aws.String(s3Key),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		log.Printf("S3 CreateMultipartUpload error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate upload"})
		return
	}

	uploadID := *createOutput.UploadId

	// Generate presigned URLs for each part
	presigner := s3.NewPresignClient(s3storage.Client)
	presignedURLs := make([]PresignedPart, 0, req.TotalParts)

	// Calculate part sizes for Content-Length restriction
	remainingSize := req.TotalSize
	partSize := (req.TotalSize + int64(req.TotalParts) - 1) / int64(req.TotalParts)
	if partSize < MinPartSize && req.TotalParts > 1 {
		partSize = MinPartSize
	}
	if partSize > MaxPartSize {
		partSize = MaxPartSize
	}

	for i := 1; i <= req.TotalParts; i++ {
		// Calculate this part's expected size
		thisPartSize := partSize
		if remainingSize < partSize {
			thisPartSize = remainingSize
		}
		remainingSize -= thisPartSize

		presignReq, err := presigner.PresignUploadPart(ctx, &s3.UploadPartInput{
			Bucket:        aws.String(s3storage.BucketName),
			Key:           aws.String(s3Key),
			UploadId:      aws.String(uploadID),
			PartNumber:    aws.Int32(int32(i)),
			ContentLength: aws.Int64(thisPartSize), // Force Content-Length restriction
		}, func(opts *s3.PresignOptions) {
			opts.Expires = PresignTTL
		})

		if err != nil {
			log.Printf("S3 PresignUploadPart error for part %d: %v", i, err)
			// Abort the multipart upload on failure
			_, _ = s3storage.Client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
				Bucket:   aws.String(s3storage.BucketName),
				Key:      aws.String(s3Key),
				UploadId: aws.String(uploadID),
			})
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate presigned URLs"})
			return
		}

		presignedURLs = append(presignedURLs, PresignedPart{
			PartNumber: i,
			URL:        presignReq.URL,
		})
	}

	log.Printf("Multipart upload initiated - UserID: %s, Key: %s, UploadID: %s, Parts: %d",
		userID, s3Key, uploadID, req.TotalParts)

	c.JSON(http.StatusOK, InitiateMultipartResponse{
		UploadID:      uploadID,
		Key:           s3Key,
		PresignedURLs: presignedURLs,
	})
}

// CompleteMultipartHandler assembles parts and creates DB record
func CompleteMultipartHandler(c *gin.Context, db *bun.DB) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req CompleteMultipartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Security: Verify the key belongs to this user
	expectedPrefix := fmt.Sprintf("users/%s/", userID)
	if !strings.HasPrefix(req.Key, expectedPrefix) {
		log.Printf("SECURITY: User %s attempted to complete upload for key: %s", userID, req.Key)
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Final quota check
	if err := checkStorageQuota(ctx, db, userID, req.TotalSize); err != nil {
		// Abort the upload since quota exceeded
		_, _ = s3storage.Client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
			Bucket:   aws.String(s3storage.BucketName),
			Key:      aws.String(req.Key),
			UploadId: aws.String(req.UploadID),
		})
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// Build completed parts for S3
	completedParts := make([]types.CompletedPart, 0, len(req.Parts))
	for _, p := range req.Parts {
		etag := p.ETag
		// S3 ETags should be quoted, ensure proper format
		if !strings.HasPrefix(etag, "\"") {
			etag = "\"" + etag + "\""
		}
		completedParts = append(completedParts, types.CompletedPart{
			PartNumber: aws.Int32(int32(p.PartNumber)),
			ETag:       aws.String(etag),
		})
	}

	// Complete the multipart upload on S3
	_, err := s3storage.Client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(s3storage.BucketName),
		Key:      aws.String(req.Key),
		UploadId: aws.String(req.UploadID),
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: completedParts,
		},
	})
	if err != nil {
		log.Printf("S3 CompleteMultipartUpload error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete upload on storage"})
		return
	}

	// Validate path
	filePath := req.FilePath
	if filePath == "" {
		filePath = "/"
	}
	validPath, err := validatePath(filePath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid path"})
		return
	}

	fullPathDB := path.Join(validPath, req.FileName)
	fullPathDB = path.Clean(fullPathDB)
	if !strings.HasPrefix(fullPathDB, "/") {
		fullPathDB = "/" + fullPathDB
	}

	// Database transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer tx.Rollback()

	// Create file record
	contentType := req.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	fileRecord := &pkg.File{
		Name:         req.FileName,
		Path:         fullPathDB,
		Size:         req.TotalSize,
		MimeType:     contentType,
		UserID:       userID,
		EncryptedKey: req.EncryptedKey,
		PreviewID:    req.PreviewID,
		IsPreview:    req.IsPreview,
	}

	delta, err := upsertFileInDB(ctx, tx, fileRecord, req.TotalSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file record"})
		return
	}

	// Process share keys
	if err := processShareKeys(ctx, tx, req.ShareKeys, fileRecord); err != nil {
		log.Printf("Error processing share keys: %v", err)
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit failed"})
		return
	}

	// Update folder sizes (non-blocking)
	if !req.IsPreview && delta != 0 {
		go func() {
			bgCtx := context.Background()
			if err := pkg.UpdateFolderSizesForFile(bgCtx, db, userID, fileRecord.Path, delta); err != nil {
				log.Printf("Failed to update folder sizes: %v", err)
			}
		}()
	}

	log.Printf("Multipart upload completed - UserID: %s, File: %s, Size: %d",
		userID, fullPathDB, req.TotalSize)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Upload completed successfully",
		"file":    fileRecord,
	})
}

// AbortMultipartHandler cancels an in-progress multipart upload
func AbortMultipartHandler(c *gin.Context, db *bun.DB) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req AbortMultipartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Security: Verify the key belongs to this user
	expectedPrefix := fmt.Sprintf("users/%s/", userID)
	if !strings.HasPrefix(req.Key, expectedPrefix) {
		log.Printf("SECURITY: User %s attempted to abort upload for key: %s", userID, req.Key)
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	_, err := s3storage.Client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(s3storage.BucketName),
		Key:      aws.String(req.Key),
		UploadId: aws.String(req.UploadID),
	})
	if err != nil {
		log.Printf("S3 AbortMultipartUpload error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to abort upload"})
		return
	}

	log.Printf("Multipart upload aborted - UserID: %s, Key: %s", userID, req.Key)
	c.JSON(http.StatusOK, gin.H{"message": "Upload aborted successfully"})
}

// GetPresignedDownloadHandler generates a temporary presigned URL for streaming download
func GetPresignedDownloadHandler(c *gin.Context, db *bun.DB) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	userID := c.GetString("user_id")
	clientIP := c.ClientIP()
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	fileIDStr := c.Param("fileID")
	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	// Rate limiting
	if !checkDownloadRateLimit(userID, clientIP) {
		log.Printf("SECURITY: Presigned download rate limit exceeded - UserID: %s, IP: %s", userID, clientIP)
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
		return
	}

	// Verify access permission (owner, direct share, or folder share)
	file, err := getFileWithPermission(ctx, db, fileID, userID)
	if err != nil {
		log.Printf("SECURITY: Unauthorized presigned download - UserID: %s, FileID: %d, IP: %s", userID, fileID, clientIP)
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Construct S3 key
	normalizedPath := path.Clean(strings.ReplaceAll(file.Path, "\\", "/"))
	if normalizedPath == "." {
		normalizedPath = "/"
	}
	if !strings.HasPrefix(normalizedPath, "/") {
		normalizedPath = "/" + normalizedPath
	}
	s3Key := fmt.Sprintf("users/%s%s", file.UserID, normalizedPath)

	// Generate presigned GET URL with streaming-optimized headers
	presigner := s3.NewPresignClient(s3storage.Client)
	presignReq, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket:                     aws.String(s3storage.BucketName),
		Key:                        aws.String(s3Key),
		ResponseContentDisposition: aws.String(fmt.Sprintf("attachment; filename=\"%s\"", file.Name)),
		ResponseCacheControl:       aws.String("no-store, no-cache, must-revalidate"),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = DownloadPresignTTL
	})

	if err != nil {
		log.Printf("S3 PresignGetObject error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate download URL"})
		return
	}

	log.Printf("Presigned streaming download URL generated - UserID: %s, FileID: %d, Size: %d", userID, fileID, file.Size)

	// Return URL with decryption metadata
	c.JSON(http.StatusOK, gin.H{
		"url":            presignReq.URL,
		"expires_in":     int(DownloadPresignTTL.Seconds()),
		"file_name":      file.Name,
		"file_size":      file.Size,
		"mime_type":      file.MimeType,
		"encrypted_key":  file.EncryptedKey,
		"encryption_alg": "AES-GCM-256",
		"chunk_size":     10 * 1024 * 1024, // 10MB - must match frontend CHUNK_SIZE
		"iv_length":      12,
		"tag_length":     16,
	})
}

// RefreshPresignedURLsHandler generates new presigned URLs for remaining parts
func RefreshPresignedURLsHandler(c *gin.Context, db *bun.DB) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	type RefreshRequest struct {
		UploadID   string `json:"upload_id" binding:"required"`
		Key        string `json:"key" binding:"required"`
		PartNumber int    `json:"part_number" binding:"required,min=1"`
		PartSize   int64  `json:"part_size" binding:"required,min=1"`
	}

	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Security: Verify the key belongs to this user
	expectedPrefix := fmt.Sprintf("users/%s/", userID)
	if !strings.HasPrefix(req.Key, expectedPrefix) {
		log.Printf("SECURITY: User %s attempted to refresh URL for key: %s", userID, req.Key)
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	presigner := s3.NewPresignClient(s3storage.Client)
	presignReq, err := presigner.PresignUploadPart(ctx, &s3.UploadPartInput{
		Bucket:        aws.String(s3storage.BucketName),
		Key:           aws.String(req.Key),
		UploadId:      aws.String(req.UploadID),
		PartNumber:    aws.Int32(int32(req.PartNumber)),
		ContentLength: aws.Int64(req.PartSize),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = PresignTTL
	})

	if err != nil {
		log.Printf("S3 PresignUploadPart refresh error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh presigned URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"part_number": req.PartNumber,
		"url":         presignReq.URL,
	})
}
