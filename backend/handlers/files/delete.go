package files

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"safercloud/backend/pkg"
	"safercloud/backend/pkg/s3storage"
	"safercloud/backend/pkg/ws"
	"safercloud/backend/utils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
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

	folder, err := pkg.GetFolder(db, folderID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dossier introuvable"})
		return
	}

	// Calculate size to subtract from parent before deletion
	if folderSize, err := pkg.GetFolderSize(c.Request.Context(), db, folderID); err == nil && folderSize > 0 {
		parentPath := filepath.Dir(folder.Path)
		if parentPath == "." {
			parentPath = "/"
		}
		if err := pkg.UpdateFolderSizesForFolderPath(c.Request.Context(), db, userID, parentPath, -folderSize); err != nil {
			log.Printf("Failed to update folder sizes on folder delete: %v", err)
		}
	}

	if err := deleteFolderRecursive(c, db, userID, folder.Path); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

func deleteFolderRecursive(c *gin.Context, db *bun.DB, userID, folderPath string) error {
	ctx := c.Request.Context()

	files, folders, err := pkg.GetFolderContentRecursive(db, userID, folderPath)
	if err != nil {
		return fmt.Errorf("Erreur lors de la récupération du contenu du dossier")
	}

	// Delete files from S3
	for _, file := range files {
		s3Key := fmt.Sprintf("users/%s%s", userID, file.Path)
		_, err = s3storage.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(s3storage.BucketName),
			Key:    aws.String(s3Key),
		})
		if err != nil {
			log.Printf("Error deleting file from S3: %v", err)
		}
	}

	// Best-effort cleanup: remove any remaining objects under the folder prefix (including folder markers)
	if s3storage.Client != nil {
		prefix := fmt.Sprintf("users/%s%s", userID, folderPath)
		if !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		if err := deleteS3Prefix(ctx, prefix); err != nil {
			log.Printf("Error deleting folder prefix from S3: %v", err)
		}
	}

	// Delete DB records in a transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("Erreur de transaction")
	}
	defer tx.Rollback()

	if len(files) > 0 {
		fileIDs := make([]int64, 0, len(files))
		for _, f := range files {
			fileIDs = append(fileIDs, f.ID)
		}

		_, _ = tx.NewDelete().Model((*pkg.ShareLink)(nil)).
			Where("resource_type = ? AND resource_id IN (?)", "file", bun.In(fileIDs)).
			Exec(ctx)
		_, _ = tx.NewDelete().Model((*pkg.FileShare)(nil)).
			Where("file_id IN (?)", bun.In(fileIDs)).
			Exec(ctx)
		_, _ = tx.NewDelete().Model((*pkg.ShareFileKey)(nil)).
			Where("file_id IN (?)", bun.In(fileIDs)).
			Exec(ctx)

		_, err = tx.NewDelete().Model((*pkg.File)(nil)).
			Where("id IN (?)", bun.In(fileIDs)).
			Where("user_id = ?", userID).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("Erreur lors de la suppression des fichiers")
		}
	}

	// Delete all subfolders + parent folder itself
	var allFolderIDs []int64
	if len(folders) > 0 {
		for _, f := range folders {
			allFolderIDs = append(allFolderIDs, f.ID)
		}
	}

	// Get parent folder ID to include it in deletion
	var parentFolder pkg.Folder
	if err := db.NewSelect().Model(&parentFolder).Where("path = ? AND user_id = ?", folderPath, userID).Scan(ctx); err == nil {
		allFolderIDs = append(allFolderIDs, parentFolder.ID)
	}

	if len(allFolderIDs) > 0 {
		_, _ = tx.NewDelete().Model((*pkg.ShareLink)(nil)).
			Where("resource_type = ? AND resource_id IN (?)", "folder", bun.In(allFolderIDs)).
			Exec(ctx)
		_, _ = tx.NewDelete().Model((*pkg.FolderShare)(nil)).
			Where("folder_id IN (?)", bun.In(allFolderIDs)).
			Exec(ctx)
		_, _ = tx.NewDelete().Model((*pkg.FolderFileKey)(nil)).
			Where("folder_id IN (?)", bun.In(allFolderIDs)).
			Exec(ctx)
		_, _ = tx.NewDelete().Model((*pkg.FolderFolderKey)(nil)).
			Where("parent_folder_id IN (?) OR sub_folder_id IN (?)", bun.In(allFolderIDs), bun.In(allFolderIDs)).
			Exec(ctx)
		_, _ = tx.NewDelete().Model((*pkg.FolderSize)(nil)).
			Where("folder_id IN (?)", bun.In(allFolderIDs)).
			Exec(ctx)

		_, err = tx.NewDelete().Model((*pkg.Folder)(nil)).
			Where("id IN (?)", bun.In(allFolderIDs)).
			Where("user_id = ?", userID).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("Erreur lors de la suppression des dossiers")
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("Erreur lors de la validation de la suppression")
	}

	// Remove local filesystem folder (best effort)
	userRoot := filepath.Join("uploads", userID)
	logicalPath := strings.TrimPrefix(folderPath, "/")
	diskPath, err := utils.SecureJoin(userRoot, logicalPath)
	if err == nil {
		_ = os.RemoveAll(diskPath)
	}

	return nil
}

func deleteS3Prefix(ctx context.Context, prefix string) error {
	if s3storage.Client == nil {
		return nil
	}

	var continuationToken *string
	for {
		listOut, err := s3storage.Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:            aws.String(s3storage.BucketName),
			Prefix:            aws.String(prefix),
			ContinuationToken: continuationToken,
		})
		if err != nil {
			return err
		}

		if len(listOut.Contents) > 0 {
			objects := make([]s3types.ObjectIdentifier, 0, len(listOut.Contents))
			for _, obj := range listOut.Contents {
				objects = append(objects, s3types.ObjectIdentifier{Key: obj.Key})
			}
			_, err = s3storage.Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
				Bucket: aws.String(s3storage.BucketName),
				Delete: &s3types.Delete{Objects: objects, Quiet: aws.Bool(true)},
			})
			if err != nil {
				return err
			}
		}

		if !aws.ToBool(listOut.IsTruncated) {
			break
		}
		continuationToken = listOut.NextContinuationToken
	}

	// Also delete the folder marker object itself if present
	_, _ = s3storage.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s3storage.BucketName),
		Key:    aws.String(strings.TrimSuffix(prefix, "/")),
	})

	return nil
}
