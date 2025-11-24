package files

import (
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
	"safercloud/backend/pkg"
	"github.com/uptrace/bun"
)

func DeleteFileHandler(c *gin.Context, db *bun.DB) {
	fileID, err := strconv.ParseInt(c.Param("fileID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de fichier invalide"})
		return
	}

	userID := c.GetInt64("userID")
	if err := pkg.DeleteFile(db, fileID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Fichier supprimé avec succès"})
}

func DeleteFolderHandler(c *gin.Context, db *bun.DB) {
	folderID, err := strconv.ParseInt(c.Param("folderID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de dossier invalide"})
		return
	}

	userID := c.GetInt64("userID")
	if err := pkg.DeleteFolder(db, folderID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Dossier supprimé avec succès"})
}
