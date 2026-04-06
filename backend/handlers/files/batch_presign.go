// backend/handlers/files/batch_presign.go
package files

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	"kagibi/backend/pkg"
	"kagibi/backend/pkg/s3storage"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// BatchPresignRequest represents a batch presign request
type BatchPresignRequest struct {
	FileIDs []int64 `json:"file_ids" binding:"required,min=1,max=500"` // Max 500 files per batch
}

// BatchPresignItem represents a single presigned URL item
type BatchPresignItem struct {
	FileID       int64  `json:"file_id"`
	URL          string `json:"url"`
	ExpiresIn    int    `json:"expires_in"`
	FileName     string `json:"file_name"`
	FileSize     int64  `json:"file_size"`
	MimeType     string `json:"mime_type"`
	EncryptedKey string `json:"encrypted_key"`
	Error        string `json:"error,omitempty"`
}

// BatchPresignResponse contains all presigned URLs
type BatchPresignResponse struct {
	URLs         []BatchPresignItem `json:"urls"`
	TotalSize    int64              `json:"total_size"`
	SuccessCount int                `json:"success_count"`
	ErrorCount   int                `json:"error_count"`
}

// BatchPresignDownloadHandler generates presigned URLs for multiple files at once
// POST /api/v1/files/batch-presign
func BatchPresignDownloadHandler(c *gin.Context, db *bun.DB) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second) // Longer timeout for batch
	defer cancel()

	userID := c.GetString("user_id")
	clientIP := c.ClientIP()
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req BatchPresignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Rate limiting: batch counts as multiple downloads
	if len(req.FileIDs) > 50 && !checkDownloadRateLimit(userID, clientIP) {
		log.Printf("SECURITY: Batch presign rate limit exceeded - UserID: %s, IP: %s, Files: %d",
			userID, clientIP, len(req.FileIDs))
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
		return
	}

	// Fetch all files in a single query
	var files []pkg.File
	if err := db.NewSelect().Model(&files).
		Where("id IN (?)", bun.In(req.FileIDs)).
		Scan(ctx); err != nil {
		log.Printf("Failed to fetch files for batch presign: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch files"})
		return
	}

	// Create a map for quick lookup
	fileMap := make(map[int64]*pkg.File, len(files))
	for i := range files {
		fileMap[files[i].ID] = &files[i]
	}

	// Generate presigned URLs in parallel (with concurrency limit)
	const maxConcurrency = 10
	semaphore := make(chan struct{}, maxConcurrency)

	var mu sync.Mutex
	results := make([]BatchPresignItem, len(req.FileIDs))
	var wg sync.WaitGroup
	var totalSize int64
	var successCount, errorCount int

	presigner := s3.NewPresignClient(s3storage.Client)

	for i, fileID := range req.FileIDs {
		wg.Add(1)
		go func(index int, fID int64) {
			defer wg.Done()

			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			file, exists := fileMap[fID]
			if !exists {
				mu.Lock()
				results[index] = BatchPresignItem{
					FileID: fID,
					Error:  "File not found",
				}
				errorCount++
				mu.Unlock()
				return
			}

			// Verify ownership or share access
			if file.UserID != userID {
				// Check share access
				hasAccess := checkShareAccess(ctx, db, fID, userID)
				if !hasAccess {
					mu.Lock()
					results[index] = BatchPresignItem{
						FileID: fID,
						Error:  "Access denied",
					}
					errorCount++
					mu.Unlock()
					return
				}
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

			// Generate presigned URL
			presignReq, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
				Bucket:                     aws.String(s3storage.BucketName),
				Key:                        aws.String(s3Key),
				ResponseContentDisposition: aws.String(fmt.Sprintf("attachment; filename=\"%s\"", file.Name)),
				ResponseCacheControl:       aws.String("no-store"),
			}, func(opts *s3.PresignOptions) {
				opts.Expires = DownloadPresignTTL
			})

			if err != nil {
				log.Printf("Presign error for file %d: %v", fID, err)
				mu.Lock()
				results[index] = BatchPresignItem{
					FileID: fID,
					Error:  "Failed to generate URL",
				}
				errorCount++
				mu.Unlock()
				return
			}

			mu.Lock()
			results[index] = BatchPresignItem{
				FileID:       fID,
				URL:          presignReq.URL,
				ExpiresIn:    int(DownloadPresignTTL.Seconds()),
				FileName:     file.Name,
				FileSize:     file.Size,
				MimeType:     file.MimeType,
				EncryptedKey: file.EncryptedKey,
			}
			totalSize += file.Size
			successCount++
			mu.Unlock()
		}(i, fileID)
	}

	wg.Wait()

	log.Printf("Batch presign completed - UserID: %s, Requested: %d, Success: %d, Errors: %d, TotalSize: %d",
		userID, len(req.FileIDs), successCount, errorCount, totalSize)

	c.JSON(http.StatusOK, BatchPresignResponse{
		URLs:         results,
		TotalSize:    totalSize,
		SuccessCount: successCount,
		ErrorCount:   errorCount,
	})
}

// checkShareAccess verifies if user has access to a file through sharing
func checkShareAccess(ctx context.Context, db *bun.DB, fileID int64, userID string) bool {
	// Check direct file share
	var count int
	count, err := db.NewSelect().
		Model((*pkg.FileShare)(nil)).
		Where("file_id = ? AND shared_with_user_id = ?", fileID, userID).
		Count(ctx)
	if err == nil && count > 0 {
		return true
	}

	// Check folder share (file's parent folder is shared)
	var file pkg.File
	if err := db.NewSelect().Model(&file).Where("id = ?", fileID).Scan(ctx); err != nil {
		return false
	}

	// Check if any parent folder is shared with this user
	var folders []pkg.Folder
	if err := db.NewSelect().Model(&folders).
		Where("user_id = ?", file.UserID).
		Scan(ctx); err != nil {
		return false
	}

	for _, folder := range folders {
		if strings.HasPrefix(file.Path, folder.Path) {
			count, err := db.NewSelect().
				Model((*pkg.FolderShare)(nil)).
				Where("folder_id = ? AND shared_with_user_id = ?", folder.ID, userID).
				Count(ctx)
			if err == nil && count > 0 {
				return true
			}
		}
	}

	return false
}

// BatchPresignByPathRequest for downloading files by their paths (for shares)
type BatchPresignByPathRequest struct {
	Paths  []string `json:"paths" binding:"required,min=1,max=500"`
	UserID string   `json:"owner_user_id"` // Owner of the files (for shared content)
}

// BatchPresignByPathHandler generates presigned URLs for files by path (useful for shared folders)
// POST /api/v1/files/batch-presign-path
func BatchPresignByPathHandler(c *gin.Context, db *bun.DB) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req BatchPresignByPathRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// If owner_user_id provided, verify share access
	ownerID := userID
	if req.UserID != "" && req.UserID != userID {
		// TODO: Verify share access to the owner's files
		ownerID = req.UserID
	}

	// Fetch files by paths
	var files []pkg.File
	if err := db.NewSelect().Model(&files).
		Where("user_id = ?", ownerID).
		Where("path IN (?)", bun.In(req.Paths)).
		Scan(ctx); err != nil {
		log.Printf("Failed to fetch files by path: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch files"})
		return
	}

	// Convert to file IDs and call the batch handler logic
	fileIDs := make([]int64, len(files))
	for i, f := range files {
		fileIDs[i] = f.ID
	}

	// Reuse the batch presign logic
	batchReq := BatchPresignRequest{FileIDs: fileIDs}
	c.Set("batch_request", batchReq)

	// Generate URLs directly here (simplified)
	presigner := s3.NewPresignClient(s3storage.Client)
	results := make([]BatchPresignItem, len(files))
	var totalSize int64

	for i, file := range files {
		normalizedPath := path.Clean(strings.ReplaceAll(file.Path, "\\", "/"))
		if !strings.HasPrefix(normalizedPath, "/") {
			normalizedPath = "/" + normalizedPath
		}
		s3Key := fmt.Sprintf("users/%s%s", file.UserID, normalizedPath)

		presignReq, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(s3storage.BucketName),
			Key:    aws.String(s3Key),
		}, func(opts *s3.PresignOptions) {
			opts.Expires = DownloadPresignTTL
		})

		if err != nil {
			results[i] = BatchPresignItem{
				FileID: file.ID,
				Error:  "Failed to generate URL",
			}
			continue
		}

		results[i] = BatchPresignItem{
			FileID:       file.ID,
			URL:          presignReq.URL,
			ExpiresIn:    int(DownloadPresignTTL.Seconds()),
			FileName:     file.Name,
			FileSize:     file.Size,
			MimeType:     file.MimeType,
			EncryptedKey: file.EncryptedKey,
		}
		totalSize += file.Size
	}

	c.JSON(http.StatusOK, BatchPresignResponse{
		URLs:         results,
		TotalSize:    totalSize,
		SuccessCount: len(files),
		ErrorCount:   0,
	})
}

// GetMultipleFilesTreeHandler returns tree for multiple selected items
// POST /api/v1/files/selection-tree
type SelectionTreeRequest struct {
	FileIDs   []int64 `json:"file_ids"`
	FolderIDs []int64 `json:"folder_ids"`
}

func GetSelectionTreeHandler(c *gin.Context, db *bun.DB) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req SelectionTreeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	type FileItem struct {
		ID           int64  `json:"id"`
		Name         string `json:"name"`
		RelativePath string `json:"relative_path"`
		Size         int64  `json:"size"`
		MimeType     string `json:"mime_type"`
		EncryptedKey string `json:"encrypted_key"`
	}

	var allFiles []FileItem
	var totalSize int64

	// Fetch direct files
	if len(req.FileIDs) > 0 {
		var files []pkg.File
		if err := db.NewSelect().Model(&files).
			Where("id IN (?) AND user_id = ?", bun.In(req.FileIDs), userID).
			Scan(ctx); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch files"})
			return
		}

		for _, f := range files {
			allFiles = append(allFiles, FileItem{
				ID:           f.ID,
				Name:         f.Name,
				RelativePath: f.Name, // At root level
				Size:         f.Size,
				MimeType:     f.MimeType,
				EncryptedKey: f.EncryptedKey,
			})
			totalSize += f.Size
		}
	}

	// Fetch files from folders (recursively)
	if len(req.FolderIDs) > 0 {
		var folders []pkg.Folder
		if err := db.NewSelect().Model(&folders).
			Where("id IN (?) AND user_id = ?", bun.In(req.FolderIDs), userID).
			Scan(ctx); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch folders"})
			return
		}

		for _, folder := range folders {
			var files []pkg.File
			if err := db.NewSelect().Model(&files).
				Where("user_id = ?", userID).
				Where("path LIKE ?", folder.Path+"%").
				Where("is_preview = ?", false).
				Scan(ctx); err != nil {
				continue
			}

			for _, f := range files {
				dir := path.Dir(f.Path)
				relativePath := strings.TrimPrefix(dir, folder.Path)
				relativePath = strings.TrimPrefix(relativePath, "/")

				var fullRelativePath string
				if relativePath == "" {
					fullRelativePath = path.Join(folder.Name, f.Name)
				} else {
					fullRelativePath = path.Join(folder.Name, relativePath, f.Name)
				}

				allFiles = append(allFiles, FileItem{
					ID:           f.ID,
					Name:         f.Name,
					RelativePath: fullRelativePath,
					Size:         f.Size,
					MimeType:     f.MimeType,
					EncryptedKey: f.EncryptedKey,
				})
				totalSize += f.Size
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"files":       allFiles,
		"total_size":  totalSize,
		"total_files": len(allFiles),
	})
}
