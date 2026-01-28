// backend/handlers/files/download.go
package files

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"safercloud/backend/pkg"
	"safercloud/backend/pkg/s3storage"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// Rate limiter for downloads
var downloadAttempts = make(map[string][]time.Time)
var downloadMutex sync.Mutex

func checkDownloadRateLimit(userID, ip string) bool {
	downloadMutex.Lock()
	defer downloadMutex.Unlock()

	key := userID + "_" + ip
	now := time.Now()

	// Clean old attempts (> 1 minute)
	if attempts, ok := downloadAttempts[key]; ok {
		var recent []time.Time
		for _, t := range attempts {
			if now.Sub(t) < time.Minute {
				recent = append(recent, t)
			}
		}
		downloadAttempts[key] = recent
	}

	// Check limit (max 200 downloads per minute for blob streaming)
	if len(downloadAttempts[key]) >= 200 {
		return false
	}

	downloadAttempts[key] = append(downloadAttempts[key], now)
	return true
}

func DownloadFileHandler(c *gin.Context, db *bun.DB) {
	userIDInterface, _ := c.Get("user_id")
	userID := userIDInterface.(string)
	clientIP := c.ClientIP()

	fileIDStr := c.Param("fileID")
	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	// Rate limiting specific to downloads
	if !checkDownloadRateLimit(userID, clientIP) {
		log.Printf("SECURITY: Download rate limit exceeded for user %s from IP %s", userID, clientIP)
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many download attempts"})
		return
	}

	// Timing attack mitigation: always take the same time
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		if elapsed < 100*time.Millisecond {
			time.Sleep(100*time.Millisecond - elapsed)
		}
	}()

	file, err := getFileWithPermission(c.Request.Context(), db, fileID, userID)
	if err != nil {
		// Security log
		log.Printf("SECURITY: Unauthorized file access attempt - UserID: %s, FileID: %d, IP: %s", userID, fileID, clientIP)

		// Generic response
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Double check
	if file.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Log legitimate access
	log.Printf("INFO: File download - UserID: %s, FileID: %d, FileName: %s", userID, fileID, file.Name)

	streamFileFromS3(c, file)
}

func getFileWithPermission(ctx context.Context, db *bun.DB, fileID int64, userID string) (*pkg.File, error) {
	// 1. Check Owner Access
	if file, err := pkg.GetFile(db, fileID, userID); err == nil {
		return file, nil
	}

	// 2. Check Direct Share Access
	if file, err := checkDirectShareAccess(ctx, db, fileID, userID); err == nil {
		return file, nil
	}

	// 3. Check Recursive Folder Share Access
	if file, err := checkFolderShareAccess(ctx, db, fileID, userID); err == nil {
		return file, nil
	}

	return nil, fmt.Errorf("permission denied")
}

func checkDirectShareAccess(ctx context.Context, db *bun.DB, fileID int64, userID string) (*pkg.File, error) {
	var fileShare pkg.FileShare
	err := db.NewSelect().Model(&fileShare).
		Where("file_id = ? AND shared_with_user_id = ?", fileID, userID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	file := new(pkg.File)
	if err := db.NewSelect().Model(file).Where("id = ?", fileID).Scan(ctx); err != nil {
		return nil, err
	}
	return file, nil
}

func checkFolderShareAccess(ctx context.Context, db *bun.DB, fileID int64, userID string) (*pkg.File, error) {
	// Retrieve file info first to get the path
	var tempFile pkg.File
	if err := db.NewSelect().Model(&tempFile).Where("id = ?", fileID).Scan(ctx); err != nil {
		return nil, err
	}

	type FolderShareWithFolder struct {
		bun.BaseModel `bun:"table:folder_shares"`
		pkg.FolderShare
		Folder *pkg.Folder `bun:"rel:belongs-to,join:folder_id=id"`
	}

	var sharesWithFolders []FolderShareWithFolder
	err := db.NewSelect().
		Model(&sharesWithFolders).
		Relation("Folder").
		Where("shared_with_user_id = ?", userID).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	for _, s := range sharesWithFolders {
		if s.Folder == nil {
			continue
		}
		if isFileInSharedFolder(tempFile.Path, s.Folder.Path) {
			log.Printf("Debug: Download allowed via recursive share. File: %s, SharedFolder: %s", tempFile.Path, s.Folder.Path)
			return &tempFile, nil
		}
	}

	return nil, fmt.Errorf("no access through folder shares")
}

func isFileInSharedFolder(filePath, folderPath string) bool {
	if folderPath != "/" && len(folderPath) > 0 && folderPath[len(folderPath)-1] != '/' {
		folderPath += "/"
	}
	if folderPath == "" {
		folderPath = "/"
	}
	// Case: Subfolder match or exact match
	return len(filePath) >= len(folderPath) && filePath[0:len(folderPath)] == folderPath
}

func streamFileFromS3(c *gin.Context, file *pkg.File) {
	// S3 Key construction: Use file.UserID (Owner) instead of userID (Requester)
	normalizedPath := path.Clean(strings.ReplaceAll(file.Path, "\\", "/"))
	if normalizedPath == "." {
		normalizedPath = "/"
	}
	if !strings.HasPrefix(normalizedPath, "/") {
		normalizedPath = "/" + normalizedPath
	}

	primaryKey := fmt.Sprintf("users/%s%s", file.UserID, normalizedPath)
	output, err := s3storage.Client.GetObject(c.Request.Context(), &s3.GetObjectInput{
		Bucket: aws.String(s3storage.BucketName),
		Key:    aws.String(primaryKey),
	})
	if err != nil {
		log.Printf("Error getting file from S3. Key: %s, Error: %v", primaryKey, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve file from storage"})
		return
	}
	defer output.Body.Close()

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

	if _, err := io.Copy(c.Writer, output.Body); err != nil {
		log.Printf("Error streaming file to client: %v", err)
	}
}
