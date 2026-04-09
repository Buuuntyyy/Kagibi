package files

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"

	"kagibi/backend/pkg"
	"kagibi/backend/pkg/workers"

	"log"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/uptrace/bun"
)

var validNameRegex = regexp.MustCompile(`^[a-zA-Z0-9\s\-\._]+$`)

type RenameRequest struct {
	ID      int64  `json:"id" binding:"required"`
	Type    string `json:"type" binding:"required,oneof=file folder"`
	NewName string `json:"new_name" binding:"required"`
}

func RenameHandler(c *gin.Context, db *bun.DB, redisClient *redis.Client) {
	var req RenameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	if !validNameRegex.MatchString(req.NewName) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid name"})
		return
	}

	userIDInterface, _ := c.Get("user_id")
	userID, _ := userIDInterface.(string)

	oldPath, parentPath, err := getRenameItemInfo(c.Request.Context(), db, req.ID, req.Type, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	newPath, finalName := calculateNewPath(req, oldPath, parentPath)

	if newPath == oldPath {
		c.JSON(http.StatusOK, gin.H{"message": "No changes made", "newName": finalName, "newPath": newPath})
		return
	}

	if err := checkRenameConflict(c.Request.Context(), db, req.Type, newPath, userID); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	if err := executeRenameTransaction(c.Request.Context(), db, req.ID, req.Type, userID, oldPath, newPath, finalName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	enqueueS3RenameTask(redisClient, userID, oldPath, newPath, req.Type == "folder")

	c.JSON(http.StatusOK, gin.H{"message": "Item renamed successfully", "newName": finalName, "newPath": newPath})
}

func getRenameItemInfo(ctx context.Context, db *bun.DB, id int64, itemType, userID string) (string, string, error) {
	var oldPath string
	if itemType == "file" {
		file, err := pkg.GetFile(db, id, userID)
		if err != nil {
			return "", "", fmt.Errorf("File not found")
		}
		oldPath = file.Path
	} else {
		var folder pkg.Folder
		err := db.NewSelect().Model(&folder).Where("id = ? AND user_id = ?", id, userID).Scan(ctx)
		if err != nil {
			return "", "", fmt.Errorf("Folder not found")
		}
		oldPath = folder.Path
	}

	parentPath := filepath.ToSlash(filepath.Dir(oldPath))
	if parentPath == "." {
		parentPath = ""
	}
	return oldPath, parentPath, nil
}

func calculateNewPath(req RenameRequest, oldPath, parentPath string) (string, string) {
	finalName := req.NewName
	if req.Type == "file" {
		// Conserver l'extension originale
		ext := filepath.Ext(oldPath)
		// Si le nouveau nom n'a pas l'extension, on l'ajoute
		if filepath.Ext(finalName) != ext {
			finalName = finalName + ext
		}
	}

	var newPath string
	if parentPath == "/" || parentPath == "" {
		newPath = "/" + finalName
	} else {
		newPath = parentPath + "/" + finalName
	}
	return newPath, finalName
}

func checkRenameConflict(ctx context.Context, db *bun.DB, itemType, newPath, userID string) error {
	var exists bool
	if itemType == "file" {
		exists, _ = db.NewSelect().Model((*pkg.File)(nil)).Where("path = ? AND user_id = ?", newPath, userID).Exists(ctx)
		if exists {
			return fmt.Errorf("A file with this name already exists")
		}
	} else {
		exists, _ = db.NewSelect().Model((*pkg.Folder)(nil)).Where("path = ? AND user_id = ?", newPath, userID).Exists(ctx)
		if exists {
			return fmt.Errorf("A folder with this name already exists")
		}
	}
	return nil
}

func executeRenameTransaction(ctx context.Context, db *bun.DB, id int64, itemType, userID, oldPath, newPath, newName string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("Database transaction error")
	}

	if itemType == "file" {
		_, err = tx.NewUpdate().Model((*pkg.File)(nil)).Set("name = ?", newName).Set("path = ?", newPath).Where("id = ?", id).Exec(ctx)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("Failed to update file")
		}
	} else {
		// Update folder itself
		_, err = tx.NewUpdate().Model((*pkg.Folder)(nil)).Set("name = ?", newName).Set("path = ?", newPath).Where("id = ?", id).Exec(ctx)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("Failed to update folder")
		}

		// Update children
		if err := updateChildrenPaths(ctx, tx, userID, oldPath, newPath); err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("Failed to commit transaction")
	}
	return nil
}

func updateChildrenPaths(ctx context.Context, tx bun.Tx, userID, oldPath, newPath string) error {
	startIdx := len(oldPath) + 1

	_, err := tx.NewRaw("UPDATE files SET path = ? || SUBSTRING(path, ?) WHERE path LIKE ? AND user_id = ?",
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

func enqueueS3RenameTask(redisClient *redis.Client, userID, oldPath, newPath string, isFolder bool) {
	srcKey := fmt.Sprintf("users/%s%s", userID, oldPath)
	destKey := fmt.Sprintf("users/%s%s", userID, newPath)

	task := workers.S3Task{
		Type:     workers.TaskRename,
		UserID:   userID,
		SrcKey:   srcKey,
		DestKey:  destKey,
		IsFolder: isFolder,
	}

	if err := workers.EnqueueTask(redisClient, task); err != nil {
		log.Printf("Failed to enqueue S3 task: %v", err)
	}
}
