package folders

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"safercloud/backend/pkg"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type CreateFolderRequest struct {
	Name string `json:"name" binding:"required"`
	Path string `json:"path" binding:"required"`
}

func CreateHandler(c *gin.Context, db *bun.DB) {
	
	var req CreateFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}
	userIDInterface, _ := c.Get("user_id")
	userIDStr, _ := userIDInterface.(string)
	userID, _ := strconv.ParseInt(userIDStr, 10, 64)

	fullPath := filepath.ToSlash(filepath.Join(req.Path, req.Name))

	diskPath := filepath.Join("uploads", userIDStr, fullPath)
	if err := os.MkdirAll(diskPath, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create folder on disk"})
		return
	}

	folder := &pkg.Folder{
		Name:   req.Name,
		Path:   fullPath,
		UserID: userID,
	}

	if err := pkg.CreateFolderDB(db, folder); err != nil {
		os.RemoveAll(diskPath) // Nettoie le dossier créé sur le disque en cas d'erreur DB
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create folder"})
		return
	}

    c.JSON(http.StatusCreated, gin.H{"message": "Dossier créé avec succès", "folder": folder})
}
