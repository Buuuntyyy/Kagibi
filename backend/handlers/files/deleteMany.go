package files

import (
	"fmt"
	"kagibi/backend/pkg"
	"kagibi/backend/pkg/s3storage"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// Note: queryUserIDEq and s3UserPathFormat are defined in delete.go (same package)

// Structure pour recevoir les IDs depuis le corps de la requête JSON
type BulkDeleteRequest struct {
	FileIDs []int64 `json:"file_ids" binding:"required"`
}

// bulkDeleteS3Objects deletes all S3 objects for the given files.
func bulkDeleteS3Objects(c *gin.Context, userID string, files []pkg.File) {
	for _, file := range files {
		s3Key := fmt.Sprintf(s3UserPathFormat, userID, file.Path)
		log.Printf("BulkDelete: Deleting S3 object. Bucket: %s, Key: %s", s3storage.BucketName, s3Key)
		if _, err := s3storage.Client.DeleteObject(c.Request.Context(), &s3.DeleteObjectInput{
			Bucket: aws.String(s3storage.BucketName),
			Key:    aws.String(s3Key),
		}); err != nil {
			log.Printf("Error deleting file from S3: %v", err)
		}
	}
}

// bulkDeleteInTx removes file records and decrements storage quota within a transaction.
func bulkDeleteInTx(c *gin.Context, tx bun.Tx, userID string, fileIDs []int64, files []pkg.File) error {
	if _, err := tx.NewDelete().Model((*pkg.File)(nil)).
		Where("id IN (?)", bun.In(fileIDs)).
		Where(queryUserIDEq, userID).
		Exec(c); err != nil {
		return fmt.Errorf("Impossible de supprimer les fichiers de la base de données")
	}
	var totalSize int64
	for _, f := range files {
		totalSize += f.Size
	}
	if _, err := tx.NewUpdate().Model((*pkg.UserPlan)(nil)).
		Set("storage_used = storage_used - ?", totalSize).
		Where(queryUserIDEq, userID).
		Exec(c); err != nil {
		return fmt.Errorf("Impossible de mettre à jour le quota de stockage")
	}
	return nil
}

func BulkDeleteHandler(c *gin.Context, db *bun.DB) {
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

	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de démarrer la transaction"})
		return
	}
	defer tx.Rollback()

	var filesToDelete []pkg.File
	if err = tx.NewSelect().Model(&filesToDelete).
		Where("id IN (?)", bun.In(req.FileIDs)).
		Where(queryUserIDEq, userID).
		Scan(c); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de récupérer les fichiers"})
		return
	}

	if len(filesToDelete) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Aucun fichier trouvé à supprimer"})
		return
	}

	bulkDeleteS3Objects(c, userID, filesToDelete)

	if err := bulkDeleteInTx(c, tx, userID, req.FileIDs, filesToDelete); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de valider la transaction"})
		return
	}

	for _, file := range filesToDelete {
		if file.IsPreview {
			continue
		}
		if err := pkg.UpdateFolderSizesForFile(c.Request.Context(), db, userID, file.Path, -file.Size); err != nil {
			log.Printf("Failed to update folder sizes on bulk delete: %v", err)
		}
	}

	notifyStorageUpdate(c.Request.Context(), db, userID)
	c.JSON(http.StatusOK, gin.H{"message": "Fichiers supprimés avec succès"})
}
