package files

import (
	"fmt"
	"log"
	"net/http"
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
	_, err = s3storage.Client.DeleteObject(c.Request.Context(), &s3.DeleteObjectInput{
		Bucket: aws.String(s3storage.BucketName),
		Key:    aws.String(s3Key),
	})

	if err != nil {
		log.Printf("Error deleting file from S3: %v", err)
		// On continue quand même pour supprimer de la BDD, ou on peut retourner une erreur
		// c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file from storage"})
		// return
	}

	// 3. Supprimer de la BDD
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
