// backend/handlers/files/download.go
package files

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"safercloud/backend/pkg"
	"github.com/uptrace/bun"
)

func DownloadFileHandler(c *gin.Context, db *bun.DB) {
	userID := c.GetInt64("userID")
	fileIDStr := c.Param("fileID")
	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	file, err := pkg.GetFile(db, fileID, userID)
	if err != nil {
		// Log l'erreur pour le débogage côté serveur
		log.Printf("Error getting file from DB. FileID: %d, UserID: %d, Error: %v", fileID, userID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found or permission denied"})
		return
	}

	// Construit le chemin physique vers le fichier dans le dossier "uploads"
	filePath := filepath.Join("uploads", file.Name)

	// Vérifie si le fichier existe physiquement avant de tenter de le servir
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("File not found on disk. Path: %s", filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "File record exists but file is missing from disk"})
		return
	}

	// Définit les en-têtes pour forcer le téléchargement
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename=\""+file.Name+"\"")
	c.Header("Content-Type", "application/octet-stream")

	c.File(filePath)
}
