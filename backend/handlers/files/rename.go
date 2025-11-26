package files

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"safercloud/backend/pkg"
	"safercloud/backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"log"
	"regexp"
)

var validNameRegex = regexp.MustCompile(`^[a-zA-Z0-9\s\-\._]+$`)

type RenameRequest struct {
	ID      int64  `json:"id" binding:"required"`
	Type    string `json:"type" binding:"required,oneof=file folder"`
	NewName string `json:"new_name" binding:"required"`
}

func RenameHandler(c *gin.Context, db *bun.DB) {
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
	userIDStr, _ := userIDInterface.(string)
	userID, _ := strconv.ParseInt(userIDStr, 10, 64)

	var oldPath string
	var parentPath string

	if req.Type == "file" {
		file, err := pkg.GetFile(db, req.ID, userID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}
		oldPath = file.Path
		parentPath = filepath.ToSlash(filepath.Dir(oldPath))
		if parentPath == "." {
			parentPath = ""
		}
	} else {
		var folder pkg.Folder
		err := db.NewSelect().Model(&folder).Where("id = ? AND user_id = ?", req.ID, userID).Scan(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
			return
		}
		oldPath = folder.Path
		parentPath = filepath.ToSlash(filepath.Dir(oldPath))
		if parentPath == "." {
			parentPath = ""
		}
	}

	var newPath string
	if parentPath == "/" || parentPath == "" {
		newPath = "/" + req.NewName
	} else {
		newPath = parentPath + "/" + req.NewName
	}

	if req.Type == "file" {
		exists, _ := db.NewSelect().Model((*pkg.File)(nil)).Where("path = ? AND user_id = ?", newPath, userID).Exists(c.Request.Context())
		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "A file with this name already exists"})
			return
		}
	} else {
		exists, _ := db.NewSelect().Model((*pkg.Folder)(nil)).Where("path = ? AND user_id = ?", newPath, userID).Exists(c.Request.Context())
		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "A folder with this name already exists"})
			return
		}
	}
	userRoot := filepath.Join("uploads", userIDStr)
	oldDiskPath, err := utils.SecureJoin(userRoot, oldPath)
	newDiskPath, err := utils.SecureJoin(userRoot, newPath)

	if err != nil {
		log.Printf("Security Alert: Path traversal in rename: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Chemin invalide"})
		return
	}

	if err := os.Rename(oldDiskPath, newDiskPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to rename item on disk: " + err.Error()})
		return
	}

	ctx := c.Request.Context()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		os.Rename(newDiskPath, oldDiskPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database transaction error"})
		return
	}

	if req.Type == "file" {
		_, err = tx.NewUpdate().Model((*pkg.File)(nil)).Set("name = ?", req.NewName).Set("path = ?", newPath).Where("id = ?", req.ID).Exec(ctx)
		if err != nil {
			tx.Rollback()
			os.Rename(newDiskPath, oldDiskPath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update file"})
			return
		}
	} else {
		_, err = tx.NewUpdate().Model((*pkg.Folder)(nil)).Set("name = ?", req.NewName).Set("path = ?", newPath).Where("id = ?", req.ID).Exec(ctx)
		if err != nil {
			tx.Rollback()
			os.Rename(newDiskPath, oldDiskPath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update folder"})
			return
		}

		startIdx := len(oldPath) + 1

		_, err = tx.NewRaw("UPDATE files SET path = ? || SUBSTRING(path, ?) WHERE path LIKE ? AND user_id = ?",
			newPath, startIdx, oldPath+"/%", userID).Exec(ctx)

		if err != nil {
			tx.Rollback()
			os.Rename(newDiskPath, oldDiskPath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update children files"})
			return
		}

		_, err = tx.NewRaw("UPDATE folders SET path = ? || SUBSTRING(path, ?) WHERE path LIKE ? AND user_id = ?",
			newPath, startIdx, oldPath+"/%", userID).Exec(ctx)

		if err != nil {
			tx.Rollback()
			os.Rename(newDiskPath, oldDiskPath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update children folders"})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		os.Rename(newDiskPath, oldDiskPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item renamed successfully", "newName": req.NewName, "newPath": newPath})
}
