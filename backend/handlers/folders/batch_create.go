// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package folders

import (
	"context"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"kagibi/backend/pkg"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// BatchCreateFolderItem is a single folder to create in a batch request.
type BatchCreateFolderItem struct {
	Name         string `json:"name" binding:"required"`
	Path         string `json:"path" binding:"required"`
	EncryptedKey string `json:"encrypted_key,omitempty"`
}

// BatchCreateFolderRequest requests the creation of many folders at once.
// An item whose path already exists is treated as a successful no-op
// ("existed") — conflicts are always auto-merged, never rejected.
type BatchCreateFolderRequest struct {
	Folders []BatchCreateFolderItem `json:"folders" binding:"required,min=1,max=500,dive"`
}

// BatchCreateFolderResult is the per-item outcome of a batch create.
type BatchCreateFolderResult struct {
	Path   string      `json:"path"`
	Status string      `json:"status"` // "created" | "existed" | "invalid"
	Error  string      `json:"error,omitempty"`
	Folder *pkg.Folder `json:"folder,omitempty"`
}

// BatchCreateFolderResponse aggregates the results of a batch create.
type BatchCreateFolderResponse struct {
	Results      []BatchCreateFolderResult `json:"results"`
	SuccessCount int                       `json:"success_count"` // created + existed
	ErrorCount   int                       `json:"error_count"`   // invalid
}

// maxBatchFolderConcurrency bounds how many folders are created in parallel per request.
// Each item does 2-3 DB round trips (existence check + insert(s)) plus a MkdirAll; run
// sequentially, a few hundred items can push request latency well past a reverse proxy's
// read timeout. Matches the pattern (and constant) already used by batch_presign.go.
const maxBatchFolderConcurrency = 10

// BatchCreateHandler creates many folders in a single request.
// POST /api/v1/folders/batch-create
func BatchCreateHandler(c *gin.Context, db *bun.DB) {
	var req BatchCreateFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	userIDInterface, _ := c.Get("user_id")
	userID, _ := userIDInterface.(string)

	// Longer timeout for batch, same reasoning as BatchPresignDownloadHandler.
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	results := make([]BatchCreateFolderResult, len(req.Folders))
	semaphore := make(chan struct{}, maxBatchFolderConcurrency)
	var wg sync.WaitGroup

	for i, item := range req.Folders {
		wg.Add(1)
		go func(index int, it BatchCreateFolderItem) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// emitEvent=false: paying for an extra DB insert + WS push per item here is
			// exactly what made large batches slow enough to trip a gateway timeout. The
			// web frontend has no listener for "folder_created" and loses nothing, but the
			// desktop app's sync engine does listen for it (sync.js) to trigger fast
			// reconciliation — a single aggregated event is emitted below instead, once per
			// batch rather than once per folder.
			folder, status, errMsg := createOneFolder(ctx, db, userID, it.Name, it.Path, it.EncryptedKey, false, false)
			results[index] = BatchCreateFolderResult{
				Path:   filepath.ToSlash(filepath.Join(it.Path, it.Name)),
				Status: string(status),
				Error:  errMsg,
				Folder: folder,
			}
		}(i, item)
	}

	wg.Wait()

	var successCount, errorCount, createdCount int
	for _, r := range results {
		switch r.Status {
		case string(folderStatusInvalid):
			errorCount++
		case string(folderStatusCreated):
			successCount++
			createdCount++
		default:
			successCount++
		}
	}

	// One aggregated event instead of one per folder — see the comment above the
	// per-item createOneFolder call for why. Only fired when something actually changed
	// (a batch that only touched already-existing folders has nothing new to reconcile).
	if createdCount > 0 {
		if err := pkg.EmitRealtimeEvent(ctx, db, userID, "folder_created", map[string]any{
			"count": createdCount,
		}); err != nil {
			log.Printf("Failed to emit batch folder_created event: %v", err)
		}
	}

	c.JSON(http.StatusOK, BatchCreateFolderResponse{
		Results:      results,
		SuccessCount: successCount,
		ErrorCount:   errorCount,
	})
}
