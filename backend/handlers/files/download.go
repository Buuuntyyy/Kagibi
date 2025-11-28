// backend/handlers/files/download.go
package files

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"safercloud/backend/pkg"
	"safercloud/backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func DownloadFileHandler(c *gin.Context, db *bun.DB) {
	userIDInterface, _ := c.Get("user_id")
	userID := userIDInterface.(string)

	fileIDStr := c.Param("fileID")
	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	file, err := pkg.GetFile(db, fileID, userID)
	if err != nil {
		// Log l'erreur pour le débogage côté serveur
		log.Printf("Error getting file from DB. FileID: %d, UserID: %s, Error: %v", fileID, userID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found or permission denied"})
		return
	}

	// Construit le chemin physique vers le fichier dans le dossier "uploads"
	// Utilise le chemin stocké en base de données (file.Path) et l'ID utilisateur
	userRoot := filepath.Join("uploads", userID)
	filePath, err := utils.SecureJoin(userRoot, file.Path)
	if err != nil {
		log.Printf("Security Alert: Path traversal in download: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Chemin de fichier invalide"})
		return
	}

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
