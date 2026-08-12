// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package files

import (
	"fmt"
	"kagibi/backend/pkg"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// Note: queryUserIDEq and s3UserPathFormat are defined in delete.go (same package)

// Structure pour recevoir les IDs depuis le corps de la requête JSON
type BulkDeleteRequest struct {
	FileIDs []int64 `json:"file_ids" binding:"required"`
}

// bulkDeleteInTx soft-deletes file records (personal trash) and decrements
// storage quota within a transaction. S3 objects, versions and shares are kept
// until the items are permanently deleted from the trash.
func bulkDeleteInTx(c *gin.Context, tx bun.Tx, userID string, fileIDs []int64, files []pkg.File) error {
	if _, err := tx.NewUpdate().Model((*pkg.File)(nil)).
		Set("deleted_at = ?, delete_root = TRUE", time.Now().UTC()).
		Where("id IN (?)", bun.In(fileIDs)).
		Where(queryUserIDEq, userID).
		Exec(c); err != nil {
		return fmt.Errorf("Impossible de déplacer les fichiers dans la corbeille")
	}
	var totalSize int64
	for _, f := range files {
		totalSize += f.Size
	}
	if _, err := tx.NewUpdate().Model((*pkg.UserPlan)(nil)).
		Set("storage_used = GREATEST(storage_used - ?, 0)", totalSize).
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
	for _, file := range filesToDelete {
		notifyFileEvent(c.Request.Context(), db, userID, "file_deleted", file.ID, file.Path)
	}
	c.JSON(http.StatusOK, gin.H{"message": "Fichiers supprimés avec succès"})
}
