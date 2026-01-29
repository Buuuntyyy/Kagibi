package files

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"safercloud/backend/pkg"
	"safercloud/backend/pkg/s3storage"
	"safercloud/backend/pkg/ws"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func DeleteFileHandler(c *gin.Context, db *bun.DB, wsManager *ws.Manager) {
	fileID, err := strconv.ParseInt(c.Param("fileID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de fichier invalide"})
		return
	}

	userIDInterface, _ := c.Get("user_id")
	userID := userIDInterface.(string)

	// 1. Récupérer les infos du fichier
	file, err := pkg.GetFile(db, fileID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Fichier introuvable"})
		return
	}

	// 2. Supprimer de S3
	s3Key := fmt.Sprintf("users/%s%s", userID, file.Path)
	log.Printf("Attempting to delete S3 object. Bucket: %s, Key: %s", s3storage.BucketName, s3Key)

	_, err = s3storage.Client.DeleteObject(c.Request.Context(), &s3.DeleteObjectInput{
		Bucket: aws.String(s3storage.BucketName),
		Key:    aws.String(s3Key),
	})

	if err != nil {
		log.Printf("Error deleting file from S3: %v", err)
	} else {
		log.Printf("Successfully deleted S3 object: %s", s3Key)
	}

	// 3. Mettre à jour la taille des dossiers (sauf preview)
	if !file.IsPreview {
		if err := pkg.UpdateFolderSizesForFile(c.Request.Context(), db, userID, file.Path, -file.Size); err != nil {
			log.Printf("Failed to update folder sizes on delete: %v", err)
		}
	}

	// 4. Supprimer de la BDD
	if err := pkg.DeleteFile(db, fileID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la suppression en base"})
		return
	}

	// Notify WebSocket about storage update
	var user pkg.User
	if err := db.NewSelect().Model(&user).Where("id = ?", userID).Scan(c); err == nil {
		wsManager.SendToUser(userID, ws.MsgStorageUpdate, map[string]interface{}{
			"storage_used": user.StorageUsed,
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Fichier supprimé avec succès"})
}

func DeleteFolderHandler(c *gin.Context, db *bun.DB, wsManager *ws.Manager) {
	folderID, err := strconv.ParseInt(c.Param("folderID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de dossier invalide"})
		return
	}

	userIDInterface, _ := c.Get("user_id")
	userID := userIDInterface.(string)

	// Note: Pour supprimer un dossier, il faudrait idéalement récupérer son chemin
	// et supprimer récursivement. pkg.DeleteFolder ne fait que le DELETE SQL.
	// Pour l'instant, on sécurise au moins l'appel DB, mais la suppression physique
	// nécessiterait une fonction GetFolder similaire à GetFile.

	// TODO: Implémenter la suppression physique récursive sécurisée pour les dossiers

	folder, err := pkg.GetFolder(db, folderID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dossier introuvable"})
		return
	}

	// Mettre à jour la taille des dossiers parents en fonction de la taille du dossier supprimé
	if folderSize, err := pkg.GetFolderSize(c.Request.Context(), db, folderID); err == nil && folderSize > 0 {
		parentPath := filepath.Dir(folder.Path)
		if parentPath == "." {
			parentPath = "/"
		}
		if err := pkg.UpdateFolderSizesForFolderPath(c.Request.Context(), db, userID, parentPath, -folderSize); err != nil {
			log.Printf("Failed to update folder sizes on folder delete: %v", err)
		}
		_ = pkg.DeleteFolderSize(c.Request.Context(), db, folderID)
	}

	if err := pkg.DeleteFolder(db, folderID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la suppression en base"})
		return
	}

	// Notify WebSocket about storage update
	var user pkg.User
	if err := db.NewSelect().Model(&user).Where("id = ?", userID).Scan(c); err == nil {
		wsManager.SendToUser(userID, ws.MsgStorageUpdate, map[string]interface{}{
			"storage_used": user.StorageUsed,
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Dossier supprimé avec succès"})
}
