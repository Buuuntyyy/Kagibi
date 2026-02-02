package files

import (
	"net/http"
	"safercloud/backend/pkg"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func SearchFilesHandler(c *gin.Context, db *bun.DB) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDInterface.(string)

	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	// Recherche dans la base de données (Supabase)
	// On cherche dans les fichiers ET les dossiers
	var files []pkg.File
	err := db.NewSelect().
		Model(&files).
		Where("user_id = ?", userID).
		Where("name ILIKE ?", "%"+query+"%"). // ILIKE pour insensible à la casse (PostgreSQL)
		Where("is_preview = ?", false).       // Exclure les fichiers preview
		Order("created_at DESC").
		Scan(c.Request.Context())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search files"})
		return
	}

	var folders []pkg.Folder
	err = db.NewSelect().
		Model(&folders).
		Where("user_id = ?", userID).
		Where("name ILIKE ?", "%"+query+"%").
		Order("created_at DESC").
		Scan(c.Request.Context())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search folders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"files":   files,
		"folders": folders,
	})
}
