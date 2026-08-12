// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package files

import (
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"kagibi/backend/pkg"
	"kagibi/backend/pkg/monitoring"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

const (
	queryUserIDEq      = "user_id = ?"
	queryIDAndUserIDEq = "id = ? AND user_id = ?"
	queryFolderIDIn    = "folder_id IN (?)"
	s3UserPathFormat   = "users/%s%s"
)

// DeleteFileHandler déplace un fichier dans la corbeille personnelle (soft delete,
// même modèle que la corbeille des orgs). L'objet S3, les versions et les partages
// sont conservés ; ils ne sont purgés qu'à la suppression définitive depuis la corbeille.
func DeleteFileHandler(c *gin.Context, db *bun.DB) {
	fileID, err := strconv.ParseInt(c.Param("fileID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de fichier invalide"})
		return
	}

	userIDInterface, _ := c.Get("user_id")
	userID := userIDInterface.(string)

	// 1. Récupérer les infos du fichier (exclut automatiquement la corbeille)
	file, err := pkg.GetFile(db, fileID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Fichier introuvable"})
		return
	}

	// 2. Mettre à jour la taille des dossiers (sauf preview)
	if !file.IsPreview {
		if err := pkg.UpdateFolderSizesForFile(c.Request.Context(), db, userID, file.Path, -file.Size); err != nil {
			log.Printf("Failed to update folder sizes on delete: %v", err)
		}
	}

	// 3. Soft delete + décrément du quota dans une transaction.
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

	if _, err := tx.NewUpdate().Model((*pkg.File)(nil)).
		Set("deleted_at = ?, delete_root = TRUE", time.Now().UTC()).
		Where(queryIDAndUserIDEq, fileID, userID).
		Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la mise à la corbeille"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la validation de la suppression"})
		return
	}

	monitoring.RecordFileDeleted()

	notifyStorageUpdate(c.Request.Context(), db, userID)
	notifyFileEvent(c.Request.Context(), db, userID, "file_deleted", fileID, file.Path)

	c.JSON(http.StatusOK, gin.H{"message": "Fichier déplacé dans la corbeille"})
}

// DeleteFolderHandler déplace un dossier (et tout son contenu) dans la corbeille
// personnelle. Le dossier lui-même est marqué delete_root=TRUE, son contenu est
// soft-deleted en cascade avec delete_root=FALSE (même modèle que les orgs).
func DeleteFolderHandler(c *gin.Context, db *bun.DB) {
	folderID, err := strconv.ParseInt(c.Param("folderID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de dossier invalide"})
		return
	}

	userIDInterface, _ := c.Get("user_id")
	userID := userIDInterface.(string)
	ctx := c.Request.Context()

	folder, err := pkg.GetFolder(db, folderID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dossier introuvable"})
		return
	}

	// Calculate size to subtract from parent before deletion
	if folderSize, err := pkg.GetFolderSize(ctx, db, folderID); err == nil && folderSize > 0 {
		parentPath := filepath.Dir(folder.Path)
		if parentPath == "." {
			parentPath = "/"
		}
		if err := pkg.UpdateFolderSizesForFolderPath(ctx, db, userID, parentPath, -folderSize); err != nil {
			log.Printf("Failed to update folder sizes on folder delete: %v", err)
		}
	}

	// Somme des fichiers vivants (hors previews) sous ce dossier pour le quota
	var totalSize struct{ Sum int64 }
	_ = db.NewSelect().TableExpr("files").
		ColumnExpr("COALESCE(SUM(size), 0) AS sum").
		Where("user_id = ? AND is_preview = false AND deleted_at IS NULL AND path LIKE ?", userID, folder.Path+"/%").
		Scan(ctx, &totalSize)

	now := time.Now().UTC()
	// Soft-delete du dossier cible (delete_root = true)
	_, _ = db.NewUpdate().Model((*pkg.Folder)(nil)).
		Set("deleted_at = ?, delete_root = TRUE", now).
		Where(queryIDAndUserIDEq, folderID, userID).
		Exec(ctx)
	// Sous-dossiers en cascade (delete_root = false)
	_, _ = db.NewUpdate().Model((*pkg.Folder)(nil)).
		Set("deleted_at = ?, delete_root = FALSE", now).
		Where("user_id = ? AND path LIKE ? AND deleted_at IS NULL", userID, folder.Path+"/%").
		Exec(ctx)
	// Fichiers en cascade (delete_root = false)
	_, _ = db.NewUpdate().Model((*pkg.File)(nil)).
		Set("deleted_at = ?, delete_root = FALSE", now).
		Where("user_id = ? AND path LIKE ? AND deleted_at IS NULL", userID, folder.Path+"/%").
		Exec(ctx)

	if totalSize.Sum > 0 {
		_, _ = db.NewUpdate().Model((*pkg.UserPlan)(nil)).
			Set("storage_used = GREATEST(storage_used - ?, 0)", totalSize.Sum).
			Where(queryUserIDEq, userID).
			Exec(ctx)
	}

	notifyStorageUpdate(ctx, db, userID)
	notifyFileEvent(ctx, db, userID, "folder_deleted", folder.ID, folder.Path)

	c.JSON(http.StatusOK, gin.H{"message": "Dossier déplacé dans la corbeille"})
}
