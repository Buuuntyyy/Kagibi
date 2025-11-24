// backend/handlers/files/upload.go
package files

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"safercloud/backend/pkg"
	"github.com/uptrace/bun"
)

func UploadHandler(c *gin.Context, db *bun.DB) {
	userID := c.GetInt64("userID")
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	files := form.File["files"]

	for _, file := range files {
		// Crée le dossier "uploads" s'il n'existe pas
		if err := os.MkdirAll("uploads", 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Chemin de destination
		dstPath := filepath.Join("uploads", file.Filename)
		dst, err := os.Create(dstPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer dst.Close()

		// Ouvre le fichier source
		src, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer src.Close()

		// Copie le fichier
		if _, err = io.Copy(dst, src); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Enregistre en base de données
		fileRecord := &pkg.File{
			Name:     file.Filename,
			Path:     "/uploads/" + file.Filename,
			Size:     file.Size,
			MimeType: file.Header.Get("Content-Type"),
			UserID:   userID,
		}
		if err := pkg.CreateFile(db, fileRecord); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Fichiers uploadés avec succès"})
}
