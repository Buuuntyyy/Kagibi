// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

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

const (
	queryUserIDAndPath = "user_id = ? AND path = ?"
	errPathTraversal   = "path traversal detected"
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
		return "", fmt.Errorf(errPathTraversal)
	}

	// 1. Clean the path using POSIX rules (virtual paths)
	cleanPath := path.Clean(rawPath)

	// 2. Check if it starts with ".."
	if strings.HasPrefix(cleanPath, "..") {
		return "", fmt.Errorf(errPathTraversal)
	}

	// 3. Check if it contains ".."
	if strings.Contains(cleanPath, "..") {
		return "", fmt.Errorf(errPathTraversal)
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
		// SÉCURITÉ : fail-closed. Autoriser l'upload quand le plan ne peut être chargé
		// permettrait de contourner le quota en provoquant l'erreur de chargement.
		return fmt.Errorf("quota check unavailable")
	}
	if planState.StorageLimit > 0 && planState.StorageUsed+size > planState.StorageLimit {
		return fmt.Errorf("Storage limit exceeded")
	}
	return nil
}

func handleChunkAssembly(fileHeader *multipart.FileHeader, userID string, req UploadRequest) (string, error) {
	tempDir := filepath.Join(os.TempDir(), "kagibi_uploads", userID)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("Failed to create temp directory")
	}

	// SÉCURITÉ : le nom de fichier multipart est contrôlé par le client. Neutraliser
	// toute traversée de répertoire avant de l'utiliser comme chemin sur disque.
	safeName := filepath.Base(filepath.FromSlash(fileHeader.Filename))
	if safeName == "." || safeName == ".." || safeName == "" || strings.ContainsAny(safeName, `/\`) {
		return "", fmt.Errorf("invalid filename")
	}
	tempFilePath := filepath.Join(tempDir, safeName+"_partial")
	// Défense en profondeur : confirmer que le chemin résolu reste dans tempDir.
	if rel, err := filepath.Rel(tempDir, tempFilePath); err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("invalid filename")
	}

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

	// 3. Quota Check (fast-fail avant tout travail ; l'enforcement autoritaire et
	//    atomique du quota se fait dans upsertFileInDB pour fermer la fenêtre TOCTOU).
	if err := checkStorageQuota(ctx, db, req.UserID, fileSize); err != nil {
		return nil, err
	}

	// 4. Compute destination path
	// SÉCURITÉ : ne conserver que le nom de base du fichier (le client contrôle Filename).
	safeName := path.Base(filepath.ToSlash(fileHeader.Filename))
	if safeName == "." || safeName == ".." || safeName == "" {
		return nil, fmt.Errorf("invalid filename")
	}
	fullPathDB := path.Join(req.Path, safeName)
	fullPathDB = path.Clean(fullPathDB)
	if !strings.HasPrefix(fullPathDB, "/") {
		fullPathDB = "/" + fullPathDB
	}

	s3Key := fmt.Sprintf("users/%s%s", req.UserID, fullPathDB)

	// 5. Update DB — réserve le quota de façon atomique. Doit précéder l'enqueue S3
	//    pour qu'un dépassement de quota n'engendre pas d'upload orphelin sur S3.
	fileRecord := &pkg.File{
		Name:         safeName,
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

	// 7. Enqueue S3 Task (après réservation du quota, avant commit)
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
	notifyFileEvent(ctx, db, req.UserID, "file_updated", fileRecord.ID, fileRecord.Path)

	return fileRecord, nil
}

// notifyFileEvent emits a granular realtime event for a file/folder change (create,
// update, delete, move) so that desktop sync clients can trigger a fast reconciliation
// instead of waiting for their periodic poll (cf. section 0.2 du plan de sync).
func notifyFileEvent(ctx context.Context, db *bun.DB, userID, eventType string, resourceID int64, path string) {
	if err := pkg.EmitRealtimeEvent(ctx, db, userID, eventType, map[string]any{
		"id":   resourceID,
		"path": path,
	}); err != nil {
		log.Printf("Failed to emit %s event: %v", eventType, err)
	}
}

// notifyMoveEvent emits a granular realtime event for a move/rename (old_path -> new_path).
func notifyMoveEvent(ctx context.Context, db *bun.DB, userID, eventType string, resourceID int64, oldPath, newPath string) {
	if err := pkg.EmitRealtimeEvent(ctx, db, userID, eventType, map[string]any{
		"id":       resourceID,
		"old_path": oldPath,
		"new_path": newPath,
	}); err != nil {
		log.Printf("Failed to emit %s event: %v", eventType, err)
	}
}

func upsertFileInDB(ctx context.Context, tx bun.Tx, file *pkg.File, size int64) (int64, error) {
	log.Printf("[UpsertFile] Attempting to upsert file: path=%s, user_id=%s, size=%d", file.Path, file.UserID, size)

	// Fetch old size before upsert so we can compute the storage delta.
	var oldSize int64
	_ = tx.NewSelect().Model((*pkg.File)(nil)).ColumnExpr("size").
		Where(queryUserIDAndPath, file.UserID, file.Path).
		Scan(ctx, &oldSize)

	// Atomic upsert: the UNIQUE index uq_files_user_path guarantees no duplicates
	// even under concurrent imports (CONCURRENT_FILES=3).
	_, err := tx.NewInsert().Model(file).
		On("CONFLICT (user_id, path) WHERE deleted_at IS NULL DO UPDATE").
		Set("name = EXCLUDED.name").
		Set("size = EXCLUDED.size").
		Set("mime_type = EXCLUDED.mime_type").
		Set("encrypted_key = EXCLUDED.encrypted_key").
		Set("chunk_size = EXCLUDED.chunk_size").
		Set("preview_id = EXCLUDED.preview_id").
		Set("is_preview = EXCLUDED.is_preview").
		Set("synced = EXCLUDED.synced").
		Set("updated_at = current_timestamp").
		Exec(ctx)
	if err != nil {
		log.Printf("[UpsertFile] ERROR upserting file: %v", err)
		return 0, err
	}
	log.Printf("[UpsertFile] Successfully upserted file with ID: %d", file.ID)

	delta := size - oldSize

	// SÉCURITÉ (TOCTOU) : réserver le quota de façon atomique. Pour un delta positif,
	// n'appliquer l'incrément que si le résultat reste sous la limite. Le verrou de ligne
	// Postgres sérialise les uploads concurrents : le second UPDATE ré-évalue la garde
	// contre la valeur déjà committée par le premier, empêchant tout dépassement collectif.
	if delta > 0 {
		res, uErr := tx.NewUpdate().Model((*pkg.UserPlan)(nil)).
			Set("storage_used = storage_used + ?", delta).
			Where("user_id = ?", file.UserID).
			Where("(storage_limit <= 0 OR storage_used + ? <= storage_limit)", delta).
			Exec(ctx)
		if uErr != nil {
			log.Printf("[UpsertFile] ERROR updating storage_used: %v", uErr)
			return 0, uErr
		}
		// 0 ligne affectée ⇒ la garde de quota a échoué (la présence de la ligne user_plans
		// est déjà garantie par checkStorageQuota, fail-closed, en amont).
		if n, _ := res.RowsAffected(); n == 0 {
			return 0, fmt.Errorf("Storage limit exceeded")
		}
		return delta, nil
	}

	// delta <= 0 : fichier de taille égale ou inférieure — toujours autorisé, sans underflow.
	_, err = tx.NewUpdate().Model((*pkg.UserPlan)(nil)).
		Set("storage_used = GREATEST(storage_used + ?, 0)", delta).
		Where("user_id = ?", file.UserID).Exec(ctx)
	if err != nil {
		log.Printf("[UpsertFile] ERROR updating storage_used: %v", err)
	}
	return delta, err
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

func processDirectShareKeys(ctx context.Context, tx bun.Tx, directShareKeysJSON string, file *pkg.File) error {
	if directShareKeysJSON == "" {
		return nil
	}
	var keysMap map[string]string
	if err := json.Unmarshal([]byte(directShareKeysJSON), &keysMap); err != nil {
		return err
	}
	var folderFileKeys []pkg.FolderFileKey
	for folderIDStr, encKey := range keysMap {
		folderID, _ := strconv.ParseInt(folderIDStr, 10, 64)
		if folderID > 0 && encKey != "" {
			folderFileKeys = append(folderFileKeys, pkg.FolderFileKey{
				FolderID:     folderID,
				FileID:       file.ID,
				EncryptedKey: encKey,
			})
		}
	}
	if len(folderFileKeys) > 0 {
		_, err := tx.NewInsert().Model(&folderFileKeys).
			On("CONFLICT (folder_id, file_id) DO UPDATE").
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
