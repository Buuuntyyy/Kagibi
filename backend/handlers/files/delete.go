// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	"time"

	"kagibi/backend/pkg"
	"kagibi/backend/pkg/monitoring"
	"kagibi/backend/pkg/s3storage"
	"kagibi/backend/utils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

const (
	queryUserIDEq    = "user_id = ?"
	queryFolderIDIn  = "folder_id IN (?)"
	s3UserPathFormat = "users/%s%s"
)

func DeleteFileHandler(c *gin.Context, db *bun.DB) {
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
	s3Key := fmt.Sprintf(s3UserPathFormat, userID, file.Path)
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

	// 4. Supprimer de la BDD et décrémenter le quota dans une transaction
	tx, err := db.BeginTx(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur de transaction"})
		return
	}
	defer tx.Rollback()

	if _, err := tx.NewUpdate().Model((*pkg.UserPlan)(nil)).
		Set("storage_used = GREATEST(storage_used - ?, 0)", file.Size).
		Where(queryUserIDEq, userID).
		Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la mise à jour du quota de stockage"})
		return
	}

	if err := pkg.DeleteFile(tx, fileID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la suppression en base"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la validation de la suppression"})
		return
	}

	monitoring.RecordFileDeleted()

	// Notify via Supabase Realtime about storage update
	notifyStorageUpdate(c.Request.Context(), db, userID)

	c.JSON(http.StatusOK, gin.H{"message": "Fichier supprimé avec succès"})
}

func DeleteFolderHandler(c *gin.Context, db *bun.DB) {
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

	// Notify via Supabase Realtime about storage update
	notifyStorageUpdate(c.Request.Context(), db, userID)

	c.JSON(http.StatusOK, gin.H{"message": "Dossier supprimé avec succès"})
}

// deleteFolderS3Objects removes all S3 objects for the given files and folder prefix.
func deleteFolderS3Objects(ctx context.Context, userID, folderPath string, files []pkg.File) {
	for _, file := range files {
		s3Key := fmt.Sprintf(s3UserPathFormat, userID, file.Path)
		if _, err := s3storage.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(s3storage.BucketName),
			Key:    aws.String(s3Key),
		}); err != nil {
			log.Printf("Error deleting file from S3: %v", err)
		}
	}
	if s3storage.Client != nil {
		prefix := fmt.Sprintf(s3UserPathFormat, userID, folderPath)
		if !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		if err := deleteS3Prefix(ctx, prefix); err != nil {
			log.Printf("Error deleting folder prefix from S3: %v", err)
		}
	}
}

// deleteFolderFilesInTx deletes file records and their shares within a transaction.
// Returns total size of deleted files.
func deleteFolderFilesInTx(ctx context.Context, tx bun.Tx, userID string, files []pkg.File) (int64, error) {
	var totalSize int64
	for _, file := range files {
		totalSize += file.Size
	}
	if len(files) == 0 {
		return 0, nil
	}
	fileIDs := make([]int64, 0, len(files))
	for _, f := range files {
		fileIDs = append(fileIDs, f.ID)
	}
	_, _ = tx.NewDelete().Model((*pkg.ShareLink)(nil)).
		Where("resource_type = ? AND resource_id IN (?)", "file", bun.In(fileIDs)).Exec(ctx)
	_, _ = tx.NewDelete().Model((*pkg.FileShare)(nil)).
		Where("file_id IN (?)", bun.In(fileIDs)).Exec(ctx)
	_, _ = tx.NewDelete().Model((*pkg.ShareFileKey)(nil)).
		Where("file_id IN (?)", bun.In(fileIDs)).Exec(ctx)
	if _, err := tx.NewDelete().Model((*pkg.File)(nil)).
		Where("id IN (?)", bun.In(fileIDs)).Where(queryUserIDEq, userID).Exec(ctx); err != nil {
		return 0, fmt.Errorf("Erreur lors de la suppression des fichiers")
	}
	return totalSize, nil
}

// deleteFolderFoldersInTx deletes folder records and their shares within a transaction.
func deleteFolderFoldersInTx(ctx context.Context, db *bun.DB, tx bun.Tx, userID, folderPath string, folders []pkg.Folder) error {
	var allFolderIDs []int64
	for _, f := range folders {
		allFolderIDs = append(allFolderIDs, f.ID)
	}
	var parentFolder pkg.Folder
	if err := db.NewSelect().Model(&parentFolder).Where("path = ? AND user_id = ?", folderPath, userID).Scan(ctx); err == nil {
		allFolderIDs = append(allFolderIDs, parentFolder.ID)
	}
	if len(allFolderIDs) == 0 {
		return nil
	}
	_, _ = tx.NewDelete().Model((*pkg.ShareLink)(nil)).
		Where("resource_type = ? AND resource_id IN (?)", "folder", bun.In(allFolderIDs)).Exec(ctx)
	_, _ = tx.NewDelete().Model((*pkg.FolderShare)(nil)).Where(queryFolderIDIn, bun.In(allFolderIDs)).Exec(ctx)
	_, _ = tx.NewDelete().Model((*pkg.FolderFileKey)(nil)).Where(queryFolderIDIn, bun.In(allFolderIDs)).Exec(ctx)
	_, _ = tx.NewDelete().Model((*pkg.FolderFolderKey)(nil)).
		Where("parent_folder_id IN (?) OR sub_folder_id IN (?)", bun.In(allFolderIDs), bun.In(allFolderIDs)).Exec(ctx)
	_, _ = tx.NewDelete().Model((*pkg.FolderSize)(nil)).Where(queryFolderIDIn, bun.In(allFolderIDs)).Exec(ctx)
	if _, err := tx.NewDelete().Model((*pkg.Folder)(nil)).
		Where("id IN (?)", bun.In(allFolderIDs)).Where(queryUserIDEq, userID).Exec(ctx); err != nil {
		return fmt.Errorf("Erreur lors de la suppression des dossiers")
	}
	return nil
}

func deleteFolderRecursive(c *gin.Context, db *bun.DB, userID, folderPath string) error {
	ctx := c.Request.Context()

	files, folders, err := pkg.GetFolderContentRecursive(db, userID, folderPath)
	if err != nil {
		return fmt.Errorf("Erreur lors de la récupération du contenu du dossier")
	}

	go func() {
		s3Ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		deleteFolderS3Objects(s3Ctx, userID, folderPath, files)
	}()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("Erreur de transaction")
	}
	defer tx.Rollback()

	totalSize, err := deleteFolderFilesInTx(ctx, tx, userID, files)
	if err != nil {
		return err
	}

	if err := deleteFolderFoldersInTx(ctx, db, tx, userID, folderPath, folders); err != nil {
		return err
	}

	if totalSize > 0 {
		if _, err = tx.NewUpdate().Model((*pkg.UserPlan)(nil)).
			Set("storage_used = GREATEST(storage_used - ?, 0)", totalSize).
			Where(queryUserIDEq, userID).Exec(ctx); err != nil {
			return fmt.Errorf("Erreur lors de la mise à jour du quota de stockage")
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
