// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package files

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"kagibi/backend/pkg"
	"kagibi/backend/pkg/s3storage"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3svc "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// TrashItem is the DTO returned by ListTrashHandler (same shape as the org trash).
type TrashItem struct {
	ID        int64      `json:"id"`
	ItemType  string     `json:"item_type"`
	Name      string     `json:"name"`
	Path      string     `json:"path"`
	DeletedAt *time.Time `json:"deleted_at"`
	Size      int64      `json:"size,omitempty"`
	MimeType  string     `json:"mime_type,omitempty"`
}

// s3DeleteUserFileIfUnreferenced deletes the S3 object backing a trashed file,
// unless a live file re-uploaded at the same path now shares that S3 key —
// deleting it would destroy the live file's content.
func s3DeleteUserFileIfUnreferenced(ctx context.Context, db *bun.DB, userID, filePath string) {
	if s3storage.Client == nil {
		return
	}
	exists, err := db.NewSelect().Model((*pkg.File)(nil)).
		Where("user_id = ? AND path = ?", userID, filePath).
		Exists(ctx)
	if err == nil && exists {
		return
	}
	s3Key := fmt.Sprintf(s3UserPathFormat, userID, filePath)
	if _, err := s3storage.Client.DeleteObject(ctx, &s3svc.DeleteObjectInput{
		Bucket: aws.String(s3storage.BucketName),
		Key:    aws.String(s3Key),
	}); err != nil {
		log.Printf("Warning: s3 delete trashed file %s: %v", s3Key, err)
	}
}

// ListTrashHandler returns all top-level trashed items (delete_root = TRUE).
func ListTrashHandler(c *gin.Context, db *bun.DB) {
	userIDInterface, _ := c.Get("user_id")
	userID := userIDInterface.(string)
	ctx := c.Request.Context()

	items := make([]TrashItem, 0, 16)

	var folders []pkg.Folder
	if err := db.NewSelect().Model(&folders).
		WhereAllWithDeleted().
		Where("user_id = ? AND deleted_at IS NOT NULL AND delete_root = TRUE", userID).
		OrderExpr("deleted_at DESC").
		Scan(ctx); err == nil {
		for _, f := range folders {
			items = append(items, TrashItem{ID: f.ID, ItemType: "folder", Name: f.Name, Path: f.Path, DeletedAt: f.DeletedAt})
		}
	}

	var files []pkg.File
	if err := db.NewSelect().Model(&files).
		WhereAllWithDeleted().
		Where("user_id = ? AND deleted_at IS NOT NULL AND delete_root = TRUE AND is_preview = false", userID).
		OrderExpr("deleted_at DESC").
		Scan(ctx); err == nil {
		for _, f := range files {
			items = append(items, TrashItem{ID: f.ID, ItemType: "file", Name: f.Name, Path: f.Path, DeletedAt: f.DeletedAt, Size: f.Size, MimeType: f.MimeType})
		}
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].DeletedAt == nil {
			return false
		}
		if items[j].DeletedAt == nil {
			return true
		}
		return items[i].DeletedAt.After(*items[j].DeletedAt)
	})

	c.JSON(http.StatusOK, items)
}

// RestoreTrashItemHandler un-deletes a single trash item (and its cascaded
// contents for folders). Returns 409 if a live item now occupies the same path.
func RestoreTrashItemHandler(c *gin.Context, db *bun.DB) {
	userIDInterface, _ := c.Get("user_id")
	userID := userIDInterface.(string)
	itemType := c.Param("itemType")
	itemID, err := strconv.ParseInt(c.Param("itemID"), 10, 64)
	if err != nil || (itemType != "file" && itemType != "folder") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Paramètres invalides"})
		return
	}
	ctx := c.Request.Context()

	if itemType == "file" {
		var file pkg.File
		if err := db.NewSelect().Model(&file).WhereAllWithDeleted().
			Where("id = ? AND user_id = ? AND deleted_at IS NOT NULL AND delete_root = TRUE", itemID, userID).
			Scan(ctx); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Élément introuvable dans la corbeille"})
			return
		}
		// Un fichier vivant occupe-t-il déjà ce chemin ?
		if exists, _ := db.NewSelect().Model((*pkg.File)(nil)).
			Where("user_id = ? AND path = ?", userID, file.Path).Exists(ctx); exists {
			c.JSON(http.StatusConflict, gin.H{"error": "Un fichier existe déjà à cet emplacement"})
			return
		}
		if _, err := db.NewUpdate().Model((*pkg.File)(nil)).
			WhereAllWithDeleted().
			Set("deleted_at = NULL, delete_root = FALSE").
			Where("id = ?", itemID).
			Exec(ctx); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Échec de la restauration"})
			return
		}
		if !file.IsPreview {
			_, _ = db.NewUpdate().Model((*pkg.UserPlan)(nil)).
				Set("storage_used = storage_used + ?", file.Size).
				Where(queryUserIDEq, userID).
				Exec(ctx)
			if err := pkg.UpdateFolderSizesForFile(ctx, db, userID, file.Path, file.Size); err != nil {
				log.Printf("Failed to update folder sizes on restore: %v", err)
			}
		}
		notifyStorageUpdate(ctx, db, userID)
		notifyFileEvent(ctx, db, userID, "file_restored", itemID, file.Path)
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	var folder pkg.Folder
	if err := db.NewSelect().Model(&folder).WhereAllWithDeleted().
		Where("id = ? AND user_id = ? AND deleted_at IS NOT NULL AND delete_root = TRUE", itemID, userID).
		Scan(ctx); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Élément introuvable dans la corbeille"})
		return
	}
	if exists, _ := db.NewSelect().Model((*pkg.Folder)(nil)).
		Where("user_id = ? AND path = ?", userID, folder.Path).Exists(ctx); exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Un dossier existe déjà à cet emplacement"})
		return
	}
	if _, err := db.NewUpdate().Model((*pkg.Folder)(nil)).
		WhereAllWithDeleted().
		Set("deleted_at = NULL, delete_root = FALSE").
		Where("id = ?", itemID).
		Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Échec de la restauration du dossier"})
		return
	}
	// Restaurer le contenu cascadé (delete_root = FALSE uniquement : les éléments
	// supprimés individuellement AVANT le dossier restent dans la corbeille)
	_, _ = db.NewUpdate().Model((*pkg.Folder)(nil)).
		WhereAllWithDeleted().
		Set("deleted_at = NULL, delete_root = FALSE").
		Where("user_id = ? AND path LIKE ? AND delete_root = FALSE AND deleted_at IS NOT NULL", userID, folder.Path+"/%").
		Exec(ctx)
	_, _ = db.NewUpdate().Model((*pkg.File)(nil)).
		WhereAllWithDeleted().
		Set("deleted_at = NULL, delete_root = FALSE").
		Where("user_id = ? AND path LIKE ? AND delete_root = FALSE AND deleted_at IS NOT NULL", userID, folder.Path+"/%").
		Exec(ctx)

	// Quota : somme des fichiers redevenus vivants sous ce chemin
	var restored struct{ Sum int64 }
	_ = db.NewSelect().TableExpr("files").
		ColumnExpr("COALESCE(SUM(size), 0) AS sum").
		Where("user_id = ? AND is_preview = false AND deleted_at IS NULL AND path LIKE ?", userID, folder.Path+"/%").
		Scan(ctx, &restored)
	if restored.Sum > 0 {
		_, _ = db.NewUpdate().Model((*pkg.UserPlan)(nil)).
			Set("storage_used = storage_used + ?", restored.Sum).
			Where(queryUserIDEq, userID).
			Exec(ctx)
	}
	// Tailles des dossiers ancêtres (miroir du décrément fait à la suppression)
	if folderSize, err := pkg.GetFolderSize(ctx, db, folder.ID); err == nil && folderSize > 0 {
		parentPath := filepath.Dir(folder.Path)
		if parentPath == "." {
			parentPath = "/"
		}
		if err := pkg.UpdateFolderSizesForFolderPath(ctx, db, userID, parentPath, folderSize); err != nil {
			log.Printf("Failed to update folder sizes on folder restore: %v", err)
		}
	}

	notifyStorageUpdate(ctx, db, userID)
	notifyFileEvent(ctx, db, userID, "folder_restored", itemID, folder.Path)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// hardDeleteTrashedFiles purges trashed file rows: guarded S3 delete, version
// cleanup, share cleanup, then ForceDelete. Quota was already decremented when
// the items entered the trash.
func hardDeleteTrashedFiles(ctx context.Context, db *bun.DB, userID string, files []pkg.File) {
	if len(files) == 0 {
		return
	}
	fileIDs := make([]int64, 0, len(files))
	for _, f := range files {
		s3DeleteUserFileIfUnreferenced(ctx, db, userID, f.Path)
		deleteAllVersionsForFile(ctx, db, f.ID, userID)
		fileIDs = append(fileIDs, f.ID)
	}
	_, _ = db.NewDelete().Model((*pkg.ShareLink)(nil)).
		Where("resource_type = ? AND resource_id IN (?)", "file", bun.In(fileIDs)).Exec(ctx)
	_, _ = db.NewDelete().Model((*pkg.FileShare)(nil)).
		Where("file_id IN (?)", bun.In(fileIDs)).Exec(ctx)
	_, _ = db.NewDelete().Model((*pkg.ShareFileKey)(nil)).
		Where("file_id IN (?)", bun.In(fileIDs)).Exec(ctx)
	_, _ = db.NewDelete().Model((*pkg.FolderFileKey)(nil)).
		Where("file_id IN (?)", bun.In(fileIDs)).Exec(ctx)
	if _, err := db.NewDelete().Model((*pkg.File)(nil)).
		WhereAllWithDeleted().ForceDelete().
		Where("id IN (?) AND user_id = ? AND deleted_at IS NOT NULL", bun.In(fileIDs), userID).
		Exec(ctx); err != nil {
		log.Printf("hardDeleteTrashedFiles: failed for user %s: %v", userID, err)
	}
}

// hardDeleteTrashedFolders purges trashed folder rows and their key/size records.
func hardDeleteTrashedFolders(ctx context.Context, db *bun.DB, userID string, folderIDs []int64) {
	if len(folderIDs) == 0 {
		return
	}
	_, _ = db.NewDelete().Model((*pkg.ShareLink)(nil)).
		Where("resource_type = ? AND resource_id IN (?)", "folder", bun.In(folderIDs)).Exec(ctx)
	_, _ = db.NewDelete().Model((*pkg.FolderShare)(nil)).Where(queryFolderIDIn, bun.In(folderIDs)).Exec(ctx)
	_, _ = db.NewDelete().Model((*pkg.FolderFileKey)(nil)).Where(queryFolderIDIn, bun.In(folderIDs)).Exec(ctx)
	_, _ = db.NewDelete().Model((*pkg.FolderFolderKey)(nil)).
		Where("parent_folder_id IN (?)", bun.In(folderIDs)).Exec(ctx)
	_, _ = db.NewDelete().Model((*pkg.FolderFolderKey)(nil)).
		Where("sub_folder_id IN (?)", bun.In(folderIDs)).Exec(ctx)
	_, _ = db.NewDelete().Model((*pkg.FolderSize)(nil)).Where(queryFolderIDIn, bun.In(folderIDs)).Exec(ctx)
	if _, err := db.NewDelete().Model((*pkg.Folder)(nil)).
		WhereAllWithDeleted().ForceDelete().
		Where("id IN (?) AND user_id = ? AND deleted_at IS NOT NULL", bun.In(folderIDs), userID).
		Exec(ctx); err != nil {
		log.Printf("hardDeleteTrashedFolders: failed for user %s: %v", userID, err)
	}
}

// PermanentDeleteTrashItemHandler permanently removes a trash item and its S3 data.
func PermanentDeleteTrashItemHandler(c *gin.Context, db *bun.DB) {
	userIDInterface, _ := c.Get("user_id")
	userID := userIDInterface.(string)
	itemType := c.Param("itemType")
	itemID, err := strconv.ParseInt(c.Param("itemID"), 10, 64)
	if err != nil || (itemType != "file" && itemType != "folder") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Paramètres invalides"})
		return
	}
	ctx := c.Request.Context()

	if itemType == "file" {
		var file pkg.File
		if err := db.NewSelect().Model(&file).WhereAllWithDeleted().
			Where("id = ? AND user_id = ? AND deleted_at IS NOT NULL", itemID, userID).
			Scan(ctx); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Élément introuvable dans la corbeille"})
			return
		}
		hardDeleteTrashedFiles(ctx, db, userID, []pkg.File{file})
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	var folder pkg.Folder
	if err := db.NewSelect().Model(&folder).WhereAllWithDeleted().
		Where("id = ? AND user_id = ? AND deleted_at IS NOT NULL", itemID, userID).
		Scan(ctx); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Élément introuvable dans la corbeille"})
		return
	}
	var files []pkg.File
	_ = db.NewSelect().Model(&files).WhereAllWithDeleted().
		Where("user_id = ? AND path LIKE ? AND deleted_at IS NOT NULL", userID, folder.Path+"/%").
		Scan(ctx)
	hardDeleteTrashedFiles(ctx, db, userID, files)

	folderIDs := []int64{folder.ID}
	var subFolders []pkg.Folder
	_ = db.NewSelect().Model(&subFolders).WhereAllWithDeleted().
		Column("id").
		Where("user_id = ? AND path LIKE ? AND deleted_at IS NOT NULL", userID, folder.Path+"/%").
		Scan(ctx)
	for _, sf := range subFolders {
		folderIDs = append(folderIDs, sf.ID)
	}
	hardDeleteTrashedFolders(ctx, db, userID, folderIDs)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// EmptyTrashHandler permanently deletes every trashed item of the current user.
func EmptyTrashHandler(c *gin.Context, db *bun.DB) {
	userIDInterface, _ := c.Get("user_id")
	userID := userIDInterface.(string)
	ctx := c.Request.Context()

	var files []pkg.File
	_ = db.NewSelect().Model(&files).WhereAllWithDeleted().
		Where("user_id = ? AND deleted_at IS NOT NULL", userID).
		Scan(ctx)
	hardDeleteTrashedFiles(ctx, db, userID, files)

	var folders []pkg.Folder
	_ = db.NewSelect().Model(&folders).WhereAllWithDeleted().
		Column("id").
		Where("user_id = ? AND deleted_at IS NOT NULL", userID).
		Scan(ctx)
	folderIDs := make([]int64, 0, len(folders))
	for _, f := range folders {
		folderIDs = append(folderIDs, f.ID)
	}
	hardDeleteTrashedFolders(ctx, db, userID, folderIDs)

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
