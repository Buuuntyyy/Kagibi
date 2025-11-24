// backend/handlers/files/download.go
package files

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"safercloud/backend/pkg"
	"github.com/uptrace/bun"
)

func DownloadFileHandler(c *gin.Context, db *bun.DB) {
	fileID, err := strconv.ParseInt(c.Param("fileID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de fichier invalide"})
		return
	}

	userID := c.GetInt64("userID")
	file, err := pkg.GetFile(db, fileID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Fichier non trouvé"})
		return
	}

	// Vérifie que l'utilisateur est propriétaire du fichier
	if file.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès interdit"})
		return
	}

	// Envoie le fichier
	c.File(file.Path)
}
