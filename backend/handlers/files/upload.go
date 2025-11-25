// backend/handlers/files/upload.go
package files

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"safercloud/backend/pkg"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func UploadHandler(c *gin.Context, db *bun.DB) {
	userIDInterface, _ := c.Get("user_id")
    userIDStr := userIDInterface.(string)
    userID, _ := strconv.ParseInt(userIDStr, 10, 64)

	path := c.PostForm("path") // Chemin virtuel où le fichier doit être stocké
	if path == "" {
		path = "/"
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	fullPathDB := filepath.ToSlash(filepath.Join(path, fileHeader.Filename))

	userUploadDir := filepath.Join("uploads", userIDStr, path)
	if err := os.MkdirAll(userUploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user upload directory"})
		return
	}

	diskPath := filepath.Join(userUploadDir, fileHeader.Filename)

	if err := c.SaveUploadedFile(fileHeader, diskPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save uploaded file"})
		return
	}


	fileRecord := &pkg.File{
		Name:   fileHeader.Filename,
		Path:   fullPathDB,
		Size:   fileHeader.Size,
		MimeType: fileHeader.Header.Get("Content-Type"),
		UserID: userID,
	}

	if err := pkg.CreateFile(db, fileRecord); err != nil {
		os.Remove(diskPath) // Nettoie le fichier sur le disque en cas d'erreur DB
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create file record"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "File uploaded successfully", "file": fileRecord})
}