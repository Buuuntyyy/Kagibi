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
	path := c.Request.FormValue("path") // Récupère le chemin virtuel

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	files := form.File["files"]

	for _, file := range files {
		// Le fichier est sauvegardé physiquement dans un dossier "uploads" à la racine.
		// Le chemin virtuel est géré par la base de données.
		uploadDir := "uploads"
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
			return
		}

		// Utilise un nom de fichier unique pour éviter les collisions
		// Pour la simplicité, nous utiliserons le nom original, mais dans une vraie app, un UUID serait mieux.
		dstPath := filepath.Join(uploadDir, file.Filename)
		dst, err := os.Create(dstPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create destination file"})
			return
		}
		defer dst.Close()

		src, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file"})
			return
		}
		defer src.Close()

		if _, err = io.Copy(dst, src); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}

		// Enregistre en base de données avec le chemin virtuel
		fileRecord := &pkg.File{
			Name:     file.Filename,
			Path:     path, // Utilise le chemin virtuel fourni par le frontend
			Size:     file.Size,
			MimeType: file.Header.Get("Content-Type"),
			UserID:   userID,
		}
		if err := pkg.CreateFile(db, fileRecord); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file record"})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Fichiers uploadés avec succès"})
}
