package files

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"safercloud/backend/pkg"
	"safercloud/backend/pkg/workers"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/uptrace/bun"
)

type MoveRequest struct {
	ID              int64  `json:"id" binding:"required"`
	Type            string `json:"type" binding:"required,oneof=file folder"`
	DestinationPath string `json:"destinationPath"` // Can be empty for root, or "/"
}

func MoveHandler(c *gin.Context, db *bun.DB, redisClient *redis.Client) {
	var req MoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	userIDInterface, _ := c.Get("user_id")
	userID, _ := userIDInterface.(string)

	// 1. Get the item to move
	var oldPath string
	var itemName string

	if req.Type == "file" {
		file, err := pkg.GetFile(db, req.ID, userID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}
		oldPath = file.Path
		itemName = file.Name
	} else {
		var folder pkg.Folder
		err := db.NewSelect().Model(&folder).Where("id = ? AND user_id = ?", req.ID, userID).Scan(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
			return
		}
		oldPath = folder.Path
		itemName = folder.Name
	}

	// 2. Calculate new path
	destPath := filepath.ToSlash(filepath.Clean(req.DestinationPath))
	if destPath == "." || destPath == "/" || destPath == "\\" {
		destPath = "" // Root
	} else if !strings.HasPrefix(destPath, "/") {
		destPath = "/" + destPath
	}

	// If destPath is empty, it means root.
	// But we need to be careful. If destPath is "", newPath becomes "/ItemName".
	// If destPath is "/Folder", newPath becomes "/Folder/ItemName".

	newPath := destPath + "/" + itemName

	// Check if moving into itself or subdirectory (for folders)
	if req.Type == "folder" {
		if oldPath == destPath {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot move folder into itself"})
			return
		}
		if strings.HasPrefix(destPath, oldPath+"/") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot move folder into its own subdirectory"})
			return
		}
	}

	// Check if destination exists (unless it's root)
	if destPath != "" {
		exists, err := db.NewSelect().Model((*pkg.Folder)(nil)).Where("path = ? AND user_id = ?", destPath, userID).Exists(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error checking destination"})
			return
		}
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Destination folder does not exist"})
			return
		}
	}

	// Check if item already exists at destination
	if req.Type == "file" {
		exists, _ := db.NewSelect().Model((*pkg.File)(nil)).Where("path = ? AND user_id = ?", newPath, userID).Exists(c.Request.Context())
		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "A file with this name already exists in the destination"})
			return
		}
	} else {
		exists, _ := db.NewSelect().Model((*pkg.Folder)(nil)).Where("path = ? AND user_id = ?", newPath, userID).Exists(c.Request.Context())
		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "A folder with this name already exists in the destination"})
			return
		}
	}

	// 3. Update DB
	ctx := c.Request.Context()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database transaction error"})
		return
	}

	if req.Type == "file" {
		_, err = tx.NewUpdate().Model((*pkg.File)(nil)).Set("path = ?", newPath).Where("id = ?", req.ID).Exec(ctx)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update file path"})
			return
		}
	} else {
		// Update folder path
		_, err = tx.NewUpdate().Model((*pkg.Folder)(nil)).Set("path = ?", newPath).Where("id = ?", req.ID).Exec(ctx)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update folder path"})
			return
		}

		// Update children paths
		// We use raw SQL for string manipulation to be safe and efficient
		// oldPath length + 1 because we want to remove the old path prefix.
		// Example: oldPath="/A", newPath="/B/A". Child="/A/C".
		// SUBSTRING("/A/C", len("/A")+1) -> "/C".
		// newPath || "/C" -> "/B/A/C".

		// Note: Postgres SUBSTRING is 1-based.
		startIdx := len(oldPath) + 1

		_, err = tx.NewRaw("UPDATE files SET path = ? || SUBSTRING(path, ?) WHERE path LIKE ? AND user_id = ?",
			newPath, startIdx, oldPath+"/%", userID).Exec(ctx)

		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update children files"})
			return
		}

		_, err = tx.NewRaw("UPDATE folders SET path = ? || SUBSTRING(path, ?) WHERE path LIKE ? AND user_id = ?",
			newPath, startIdx, oldPath+"/%", userID).Exec(ctx)

		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update children folders"})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// 4. Enqueue S3 Task
	// S3 Keys: users/{userID}/{path}
	srcKey := fmt.Sprintf("users/%s%s", userID, oldPath)
	destKey := fmt.Sprintf("users/%s%s", userID, newPath)

	task := workers.S3Task{
		Type:     workers.TaskMove,
		UserID:   userID,
		SrcKey:   srcKey,
		DestKey:  destKey,
		IsFolder: req.Type == "folder",
	}

	if err := workers.EnqueueTask(redisClient, task); err != nil {
		// Log error but don't fail request as DB is updated
		// In a real system, we might want to rollback DB or have a retry mechanism
		fmt.Printf("Failed to enqueue S3 task: %v\n", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item moved successfully", "newPath": newPath})
}
