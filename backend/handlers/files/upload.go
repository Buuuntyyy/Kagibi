// backend/handlers/files/upload.go
package files

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"kagibi/backend/pkg"
	"kagibi/backend/pkg/monitoring"
	"kagibi/backend/pkg/workers"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/uptrace/bun"
)

type UploadRequest struct {
	UserID       string
	Path         string
	EncryptedKey string
	ShareKeys    string
	ChunkIndex   int
	TotalChunks  int
	IsChunked    bool
	TotalSize    int64
	PreviewID    *int64
	IsPreview    bool
}

// validatePath validates and sanitizes file paths to prevent path traversal
func validatePath(inputPath string) (string, error) {
	// Normalize separators and check traversal early
	rawPath := strings.ReplaceAll(inputPath, "\\", "/")
	if strings.Contains(rawPath, "..") {
		return "", fmt.Errorf("path traversal detected")
	}

	// 1. Clean the path using POSIX rules (virtual paths)
	cleanPath := path.Clean(rawPath)

	// 2. Check if it starts with ".."
	if strings.HasPrefix(cleanPath, "..") {
		return "", fmt.Errorf("path traversal detected")
	}

	// 3. Check if it contains ".."
	if strings.Contains(cleanPath, "..") {
		return "", fmt.Errorf("path traversal detected")
	}

	// 4. Ensure it starts with "/"
	if !strings.HasPrefix(cleanPath, "/") {
		cleanPath = "/" + cleanPath
	}

	// 5. Check for forbidden characters
	invalidChars := []string{"\x00", "\n", "\r"}
	for _, char := range invalidChars {
		if strings.Contains(cleanPath, char) {
			return "", fmt.Errorf("invalid characters in path")
		}
	}

	return cleanPath, nil
}

func UploadHandler(c *gin.Context, db *bun.DB, redisClient *redis.Client) {
	userID := c.GetString("user_id")

	req, err := parseUploadRequest(c, userID)
	if err != nil {
		log.Printf("SECURITY: Invalid upload request - UserID: %s, Error: %v", userID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Early Storage Quota Check
	if req.ChunkIndex == 0 && req.TotalSize > 0 {
		if err := checkStorageQuota(c.Request.Context(), db, userID, req.TotalSize); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	tempFilePath, err := handleChunkAssembly(fileHeader, userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	isLastChunk := !req.IsChunked || (req.ChunkIndex == req.TotalChunks-1)
	if !isLastChunk {
		c.JSON(http.StatusOK, gin.H{"message": "Chunk uploaded successfully", "chunk_index": req.ChunkIndex})
		return
	}

	// Finalize Upload
	fileRecord, err := finalizeUpload(c.Request.Context(), db, redisClient, req, fileHeader, tempFilePath)
	if err != nil {
		// Attempt cleanup
		os.Remove(tempFilePath)
		// Return appropriate error code based on error type? For now 500 or 403
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Increment uploaded files counter
	monitoring.FileUploadsTotal.Inc()

	c.JSON(http.StatusCreated, gin.H{"message": "Upload en cours de traitement", "file": fileRecord})
}

func parseUploadRequest(c *gin.Context, userID string) (UploadRequest, error) {
	path := c.PostForm("path")
	if path == "" {
		path = "/"
	}

	// CRITICAL: Validate path to prevent traversal
	validPath, err := validatePath(path)
	if err != nil {
		return UploadRequest{}, err
	}

	chunkIndexStr := c.PostForm("chunk_index")
	totalChunksStr := c.PostForm("total_chunks")
	chunkIndex := 0
	totalChunks := 1
	isChunked := chunkIndexStr != "" && totalChunksStr != ""

	if isChunked {
		chunkIndex, _ = strconv.Atoi(chunkIndexStr)
		totalChunks, _ = strconv.Atoi(totalChunksStr)
	}

	var totalSize int64
	if totalSizeStr := c.PostForm("total_file_size"); totalSizeStr != "" {
		totalSize, _ = strconv.ParseInt(totalSizeStr, 10, 64)
	}

	previewIDStr := c.PostForm("preview_id")
	var previewID *int64
	if previewIDStr != "" {
		pid, _ := strconv.ParseInt(previewIDStr, 10, 64)
		previewID = &pid
	}

	return UploadRequest{
		UserID:       userID,
		Path:         validPath, // PATH VALIDÉ
		EncryptedKey: c.PostForm("encrypted_key"),
		ShareKeys:    c.PostForm("share_keys"),
		ChunkIndex:   chunkIndex,
		TotalChunks:  totalChunks,
		IsChunked:    isChunked,
		TotalSize:    totalSize,
		PreviewID:    previewID,
		IsPreview:    c.PostForm("is_preview") == "true",
	}, nil
}

func checkStorageQuota(ctx context.Context, db *bun.DB, userID string, size int64) error {
	planState, err := pkg.FindUserPlanByUserID(db, userID)
	if err != nil {
		return nil // Skip check if user load fails (handled later)
	}
	if planState.StorageUsed+size > planState.StorageLimit {
		return fmt.Errorf("Storage limit exceeded")
	}
	return nil
}

func handleChunkAssembly(fileHeader *multipart.FileHeader, userID string, req UploadRequest) (string, error) {
	tempDir := filepath.Join(os.TempDir(), "kagibi_uploads", userID)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("Failed to create temp directory")
	}

	tempFilePath := filepath.Join(tempDir, fileHeader.Filename+"_partial")

	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	if req.IsChunked && req.ChunkIndex > 0 {
		flags = os.O_WRONLY | os.O_APPEND
	}

	dst, err := os.OpenFile(tempFilePath, flags, 0644)
	if err != nil {
		return "", fmt.Errorf("Failed to open temp file")
	}
	defer dst.Close()

	src, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("Failed to open uploaded file")
	}
	defer src.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("Failed to write chunk to temp file")
	}

	return tempFilePath, nil
}

func finalizeUpload(ctx context.Context, db *bun.DB, redisClient *redis.Client, req UploadRequest, fileHeader *multipart.FileHeader, tempFilePath string) (*pkg.File, error) {
	// 1. Verify file size
	fi, err := os.Stat(tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to stat final file")
	}
	fileSize := fi.Size()

	// 2. Transaction for DB updates
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("Database error")
	}
	defer tx.Rollback()

	// 3. Quota Check
	if err := checkStorageQuota(ctx, db, req.UserID, fileSize); err != nil {
		return nil, err
	}

	// 4. Enqueue S3 Task
	fullPathDB := path.Join(req.Path, fileHeader.Filename)
	fullPathDB = path.Clean(fullPathDB)
	if !strings.HasPrefix(fullPathDB, "/") {
		fullPathDB = "/" + fullPathDB
	}

	s3Key := fmt.Sprintf("users/%s%s", req.UserID, fullPathDB)

	task := workers.S3Task{
		Type:        workers.TaskUpload,
		UserID:      req.UserID,
		SrcKey:      tempFilePath,
		DestKey:     s3Key,
		ContentType: fileHeader.Header.Get("Content-Type"),
	}

	if err := workers.EnqueueTask(redisClient, task); err != nil {
		log.Printf("Upload Handler ERROR: Failed to enqueue task: %v", err)
		return nil, fmt.Errorf("Failed to enqueue upload task")
	}

	// 5. Update DB
	fileRecord := &pkg.File{
		Name:         fileHeader.Filename,
		Path:         fullPathDB,
		Size:         fileSize,
		MimeType:     fileHeader.Header.Get("Content-Type"),
		UserID:       req.UserID,
		EncryptedKey: req.EncryptedKey,
		PreviewID:    req.PreviewID,
		IsPreview:    req.IsPreview,
	}

	delta, err := upsertFileInDB(ctx, tx, fileRecord, fileSize)
	if err != nil {
		return nil, err
	}

	// 6. Handle Share Keys
	if err := processShareKeys(ctx, tx, req.ShareKeys, fileRecord); err != nil {
		fmt.Printf("Error inserting share keys: %v\n", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("Transaction commit failed")
	}

	if !req.IsPreview && delta != 0 {
		if err := pkg.UpdateFolderSizesForFile(ctx, db, req.UserID, fileRecord.Path, delta); err != nil {
			log.Printf("Failed to update folder sizes: %v", err)
		}
	}

	// 7. Notify Storage Update via Supabase Realtime
	notifyStorageUpdate(ctx, db, req.UserID)

	return fileRecord, nil
}

func upsertFileInDB(ctx context.Context, tx bun.Tx, file *pkg.File, size int64) (int64, error) {
	log.Printf("[UpsertFile] Attempting to upsert file: path=%s, user_id=%s, size=%d", file.Path, file.UserID, size)

	exists, _ := tx.NewSelect().Model((*pkg.File)(nil)).
		Where("user_id = ? AND path = ?", file.UserID, file.Path).
		Exists(ctx)

	if exists {
		log.Printf("[UpsertFile] File exists, updating: path=%s", file.Path)
		var oldFile pkg.File
		if err := tx.NewSelect().Model(&oldFile).Where("user_id = ? AND path = ?", file.UserID, file.Path).Scan(ctx); err == nil {
			// Update user storage: remove old, add new
			_, _ = tx.NewUpdate().Model((*pkg.UserPlan)(nil)).
				Set("storage_used = storage_used - ? + ?", oldFile.Size, size).
				Where("user_id = ?", file.UserID).Exec(ctx)
			file.ID = oldFile.ID
			// Return delta for folder sizes
			delta := size - oldFile.Size
			_, err := tx.NewUpdate().Model(file).Where("user_id = ? AND path = ?", file.UserID, file.Path).Exec(ctx)
			if err != nil {
				log.Printf("[UpsertFile] ERROR updating file: %v", err)
			} else {
				log.Printf("[UpsertFile] Successfully updated file with ID: %d", file.ID)
			}
			return delta, err
		}
		_, err := tx.NewUpdate().Model(file).Where("user_id = ? AND path = ?", file.UserID, file.Path).Exec(ctx)
		if err != nil {
			log.Printf("[UpsertFile] ERROR updating file: %v", err)
		} else {
			log.Printf("[UpsertFile] Successfully updated file with ID: %d", file.ID)
		}
		return 0, err
	}

	log.Printf("[UpsertFile] File doesn't exist, inserting new: path=%s", file.Path)
	_, err := tx.NewInsert().Model(file).Exec(ctx)
	if err != nil {
		log.Printf("[UpsertFile] ERROR inserting file: %v", err)
		return 0, err
	}
	log.Printf("[UpsertFile] Successfully inserted file with ID: %d", file.ID)

	_, err = tx.NewUpdate().Model((*pkg.UserPlan)(nil)).
		Set("storage_used = storage_used + ?", size).
		Where("user_id = ?", file.UserID).Exec(ctx)
	if err != nil {
		log.Printf("[UpsertFile] ERROR updating storage_used: %v", err)
	}
	return size, err
}

func processShareKeys(ctx context.Context, tx bun.Tx, shareKeysJSON string, file *pkg.File) error {
	if shareKeysJSON == "" {
		return nil
	}
	var shareKeysMap map[string]string
	if err := json.Unmarshal([]byte(shareKeysJSON), &shareKeysMap); err != nil {
		return err
	}

	var shareFileKeys []pkg.ShareFileKey
	for sIDStr, key := range shareKeysMap {
		sID, _ := strconv.ParseInt(sIDStr, 10, 64)
		if sID > 0 {
			shareFileKeys = append(shareFileKeys, pkg.ShareFileKey{
				ShareID:      sID,
				FileID:       file.ID,
				EncryptedKey: key,
			})
		}
	}

	if len(shareFileKeys) > 0 {
		_, err := tx.NewInsert().Model(&shareFileKeys).
			On("CONFLICT (share_id, file_id) DO UPDATE").
			Set("encrypted_key = EXCLUDED.encrypted_key").
			Exec(ctx)
		return err
	}
	return nil
}

func notifyStorageUpdate(ctx context.Context, db *bun.DB, userID string) {
	planState, err := pkg.FindUserPlanByUserID(db, userID)
	if err == nil {
		payload := map[string]interface{}{
			"storage_used": planState.StorageUsed,
		}
		if err := pkg.EmitRealtimeEvent(ctx, db, userID, "storage_update", payload); err != nil {
			log.Printf("Failed to emit storage_update event: %v", err)
		}
	}
}
