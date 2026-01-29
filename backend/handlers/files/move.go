package files

import (
	"context"
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

const CLAUSE = "path = ? AND user_id = ?"

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

	oldPath, itemName, err := getItemInfo(c.Request.Context(), db, req.ID, req.Type, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	newPath, err := validateMove(c.Request.Context(), db, req, userID, oldPath, itemName)
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(err.Error(), "already exists") {
			status = http.StatusConflict
		} else if strings.Contains(err.Error(), "Database error") {
			status = http.StatusInternalServerError
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	if err := executeMoveTransaction(c.Request.Context(), db, req.ID, req.Type, userID, oldPath, newPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	enqueueS3MoveTask(redisClient, userID, oldPath, newPath, req.Type == "folder")

	c.JSON(http.StatusOK, gin.H{"message": "Item moved successfully", "newPath": newPath})
}

func getItemInfo(ctx context.Context, db *bun.DB, id int64, itemType, userID string) (string, string, error) {
	if itemType == "file" {
		file, err := pkg.GetFile(db, id, userID)
		if err != nil {
			return "", "", fmt.Errorf("File not found")
		}
		return file.Path, file.Name, nil
	}

	var folder pkg.Folder
	err := db.NewSelect().Model(&folder).Where("id = ? AND user_id = ?", id, userID).Scan(ctx)
	if err != nil {
		return "", "", fmt.Errorf("Folder not found")
	}
	return folder.Path, folder.Name, nil
}

func validateMove(ctx context.Context, db *bun.DB, req MoveRequest, userID, oldPath, itemName string) (string, error) {
	destPath := normalizeDestPath(req.DestinationPath)
	newPath := destPath + "/" + itemName

	if err := checkFolderConstraints(req.Type, oldPath, destPath); err != nil {
		return "", err
	}

	if err := checkDestinationExists(ctx, db, destPath, userID); err != nil {
		return "", err
	}

	if err := checkConflict(ctx, db, req.Type, newPath, userID); err != nil {
		return "", err
	}

	return newPath, nil
}

func normalizeDestPath(rawPath string) string {
	destPath := filepath.ToSlash(filepath.Clean(rawPath))
	if destPath == "." || destPath == "/" || destPath == "\\" {
		return ""
	}
	if !strings.HasPrefix(destPath, "/") {
		return "/" + destPath
	}
	return destPath
}

func checkFolderConstraints(itemType, oldPath, destPath string) error {
	if itemType == "folder" {
		if oldPath == destPath {
			return fmt.Errorf("Cannot move folder into itself")
		}
		if strings.HasPrefix(destPath, oldPath+"/") {
			return fmt.Errorf("Cannot move folder into its own subdirectory")
		}
	}
	return nil
}

func checkDestinationExists(ctx context.Context, db *bun.DB, destPath, userID string) error {
	if destPath != "" {
		exists, err := db.NewSelect().Model((*pkg.Folder)(nil)).Where(CLAUSE, destPath, userID).Exists(ctx)
		if err != nil {
			return fmt.Errorf("Database error checking destination")
		}
		if !exists {
			return fmt.Errorf("Destination folder does not exist")
		}
	}
	return nil
}

func checkConflict(ctx context.Context, db *bun.DB, itemType, newPath, userID string) error {
	var exists bool
	if itemType == "file" {
		exists, _ = db.NewSelect().Model((*pkg.File)(nil)).Where(CLAUSE, newPath, userID).Exists(ctx)
		if exists {
			return fmt.Errorf("A file with this name already exists in the destination")
		}
	} else {
		exists, _ = db.NewSelect().Model((*pkg.Folder)(nil)).Where(CLAUSE, newPath, userID).Exists(ctx)
		if exists {
			return fmt.Errorf("A folder with this name already exists in the destination")
		}
	}
	return nil
}

func executeMoveTransaction(ctx context.Context, db *bun.DB, id int64, itemType, userID, oldPath, newPath string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("Database transaction error")
	}

	if itemType == "file" {
		_, err = tx.NewUpdate().Model((*pkg.File)(nil)).Set("path = ?", newPath).Where("id = ?", id).Exec(ctx)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("Failed to update file path")
		}
	} else {
		if err := updateFolderAndChildren(ctx, tx, id, userID, oldPath, newPath); err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("Failed to commit transaction")
	}
	return nil
}

func updateFolderAndChildren(ctx context.Context, tx bun.Tx, id int64, userID, oldPath, newPath string) error {
	// Update folder path
	_, err := tx.NewUpdate().Model((*pkg.Folder)(nil)).Set("path = ?", newPath).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("Failed to update folder path")
	}

	startIdx := len(oldPath) + 1

	_, err = tx.NewRaw("UPDATE files SET path = ? || SUBSTRING(path, ?) WHERE path LIKE ? AND user_id = ?",
		newPath, startIdx, oldPath+"/%", userID).Exec(ctx)
	if err != nil {
		return fmt.Errorf("Failed to update children files")
	}

	_, err = tx.NewRaw("UPDATE folders SET path = ? || SUBSTRING(path, ?) WHERE path LIKE ? AND user_id = ?",
		newPath, startIdx, oldPath+"/%", userID).Exec(ctx)
	if err != nil {
		return fmt.Errorf("Failed to update children folders")
	}
	return nil
}

func enqueueS3MoveTask(redisClient *redis.Client, userID, oldPath, newPath string, isFolder bool) {
	srcKey := fmt.Sprintf("users/%s%s", userID, oldPath)
	destKey := fmt.Sprintf("users/%s%s", userID, newPath)

	task := workers.S3Task{
		Type:     workers.TaskMove,
		UserID:   userID,
		SrcKey:   srcKey,
		DestKey:  destKey,
		IsFolder: isFolder,
	}

	if err := workers.EnqueueTask(redisClient, task); err != nil {
		fmt.Printf("Failed to enqueue S3 task: %v\n", err)
	}
}
