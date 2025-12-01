package files

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"safercloud/backend/pkg"
	"safercloud/backend/pkg/ws"
	"safercloud/backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// Structure pour recevoir les IDs depuis le corps de la requête JSON
type BulkDeleteRequest struct {
	FileIDs []int64 `json:"file_ids" binding:"required"`
}

func BulkDeleteHandler(c *gin.Context, db *bun.DB, wsManager *ws.Manager) {
	var req BulkDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Liste d'IDs invalide"})
		return
	}

	if len(req.FileIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Aucun fichier à supprimer"})
		return
	}

	userIDInterface, _ := c.Get("user_id")
	userID := userIDInterface.(string)

	// Débuter une transaction pour s'assurer que tout réussit ou tout échoue
	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de démarrer la transaction"})
		return
	}
	defer tx.Rollback() // Rollback si quelque chose se passe mal

	var filesToDelete []pkg.File
	// Récupérer les infos des fichiers pour obtenir leurs chemins
	err = tx.NewSelect().Model(&filesToDelete).
		Where("id IN (?)", bun.In(req.FileIDs)).
		Where("user_id = ?", userID).
		Scan(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de récupérer les fichiers"})
		return
	}

	if len(filesToDelete) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Aucun fichier trouvé à supprimer"})
		return
	}

	// Supprimer les fichiers du système de fichiers
	userRoot := filepath.Join("uploads", userID)
	for _, file := range filesToDelete {
		diskPath, err := utils.SecureJoin(userRoot, file.Path)
		if err != nil {
			log.Printf("Security Alert: Path traversal in bulk delete: %v", err)
			continue // On saute ce fichier suspect
		}
		os.Remove(diskPath) // On ignore l'erreur si le fichier n'existe déjà plus
	}

	// Supprimer les enregistrements de la base de données
	_, err = tx.NewDelete().Model((*pkg.File)(nil)).
		Where("id IN (?)", bun.In(req.FileIDs)).
		Where("user_id = ?", userID).
		Exec(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de supprimer les fichiers de la base de données"})
		return
	}

	// Mettre à jour l'espace de stockage utilisé
	var totalSize int64
	for _, file := range filesToDelete {
		totalSize += file.Size
	}

	_, err = tx.NewUpdate().Model((*pkg.User)(nil)).
		Set("storage_used = storage_used - ?", totalSize).
		Where("id = ?", userID).
		Exec(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de mettre à jour le quota de stockage"})
		return
	}

	// Si tout s'est bien passé, on commit la transaction
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de valider la transaction"})
		return
	}

	// Notify WebSocket about storage update
	var user pkg.User
	if err := db.NewSelect().Model(&user).Where("id = ?", userID).Scan(c); err == nil {
		wsManager.SendToUser(userID, ws.MsgStorageUpdate, map[string]interface{}{
			"storage_used": user.StorageUsed,
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Fichiers supprimés avec succès"})
}
