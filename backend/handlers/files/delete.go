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

func DeleteFileHandler(c *gin.Context, db *bun.DB) {
	fileID, err := strconv.ParseInt(c.Param("fileID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de fichier invalide"})
		return
	}

	userIDStr := c.GetString("user_id")
	userID, _ := strconv.ParseInt(userIDStr, 10, 64)

	// 1. Récupérer les infos du fichier
	file, err := pkg.GetFile(db, fileID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Fichier introuvable"})
		return
	}

	// 2. Supprimer du disque de manière sécurisée
	userRoot := filepath.Join("uploads", userIDStr)
	diskPath, err := utils.SecureJoin(userRoot, file.Path)
	if err != nil {
		log.Printf("Security Alert: Path traversal in delete file: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Chemin invalide"})
		return
	}

	if err := os.Remove(diskPath); err != nil && !os.IsNotExist(err) {
		log.Printf("Error deleting file from disk: %v", err)
		// On continue quand même pour supprimer de la BDD
	}

	// 3. Supprimer de la BDD
	if err := pkg.DeleteFile(db, fileID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la suppression en base"})
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

	userIDStr := c.GetString("user_id")
	userID, _ := strconv.ParseInt(userIDStr, 10, 64)

	// Note: Pour supprimer un dossier, il faudrait idéalement récupérer son chemin
	// et supprimer récursivement. pkg.DeleteFolder ne fait que le DELETE SQL.
	// Pour l'instant, on sécurise au moins l'appel DB, mais la suppression physique
	// nécessiterait une fonction GetFolder similaire à GetFile.

	// TODO: Implémenter la suppression physique récursive sécurisée pour les dossiers

	if err := pkg.DeleteFolder(db, folderID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la suppression en base"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Dossier supprimé avec succès"})
}
