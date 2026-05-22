// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package organizations

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"path"
	"strconv"

	"kagibi/backend/pkg"
	"kagibi/backend/pkg/workers"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/uptrace/bun"
)

type orgRenameRequest struct {
	EncryptedName string `json:"encrypted_name" binding:"required"`
}

// RenameOrgFile renames an org file (encrypted name sent from client, DB + S3 updated).
// PATCH /orgs/:orgID/fs/file/:fileID/rename
func (h *OrgHandler) RenameOrgFile(c *gin.Context) {
	userID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}
	fileID, err := strconv.ParseInt(c.Param("fileID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file id"})
		return
	}

	var req orgRenameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if !h.checkMFAEnforcement(c, orgID, userID) {
		return
	}

	ctx := c.Request.Context()

	var file pkg.OrgFile
	if err := h.DB.NewSelect().Model(&file).
		Where("id = ? AND org_id = ?", fileID, orgID).
		Scan(ctx); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch file"})
		return
	}

	perm, err := h.resolvePermission(ctx, orgID, userID, file.FolderPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check permissions"})
		return
	}
	if perm < PermWrite {
		c.JSON(http.StatusForbidden, gin.H{"error": "write access denied"})
		return
	}

	newPath := path.Join(path.Dir(file.Path), req.EncryptedName)
	if newPath == file.Path {
		c.JSON(http.StatusOK, gin.H{"message": "no change", "new_path": file.Path})
		return
	}

	exists, _ := h.DB.NewSelect().Model((*pkg.OrgFile)(nil)).
		Where("org_id = ? AND path = ?", orgID, newPath).
		Exists(ctx)
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "a file with this name already exists"})
		return
	}

	oldS3Key := orgS3Key(orgID, file.Path)
	newS3Key := orgS3Key(orgID, newPath)

	if _, err := h.DB.NewUpdate().Model(&file).
		Set("name = ?", req.EncryptedName).
		Set("path = ?", newPath).
		WherePK().
		Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to rename file"})
		return
	}

	enqueueOrgS3Rename(h.RedisClient, orgID, oldS3Key, newS3Key, false)
	h.logAudit(ctx, orgID, userID, "file_renamed", strconv.FormatInt(fileID, 10), "file", req.EncryptedName)

	c.JSON(http.StatusOK, gin.H{"message": "file renamed", "new_path": newPath})
}

// RenameOrgFolder renames an org folder and cascades the path change to all children.
// PATCH /orgs/:orgID/fs/folder/:folderID/rename
func (h *OrgHandler) RenameOrgFolder(c *gin.Context) {
	userID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}
	folderID, err := strconv.ParseInt(c.Param("folderID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid folder id"})
		return
	}

	var req orgRenameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if !h.checkMFAEnforcement(c, orgID, userID) {
		return
	}

	ctx := c.Request.Context()

	var folder pkg.OrgFolder
	if err := h.DB.NewSelect().Model(&folder).
		Where("id = ? AND org_id = ?", folderID, orgID).
		Scan(ctx); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "folder not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch folder"})
		return
	}

	perm, err := h.resolvePermission(ctx, orgID, userID, folder.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check permissions"})
		return
	}
	if perm < PermWrite {
		c.JSON(http.StatusForbidden, gin.H{"error": "write access denied"})
		return
	}

	newPath := path.Join(folder.ParentPath, req.EncryptedName)
	if newPath == folder.Path {
		c.JSON(http.StatusOK, gin.H{"message": "no change", "new_path": folder.Path})
		return
	}

	exists, _ := h.DB.NewSelect().Model((*pkg.OrgFolder)(nil)).
		Where("org_id = ? AND path = ?", orgID, newPath).
		Exists(ctx)
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "a folder with this name already exists"})
		return
	}

	oldPath := folder.Path
	oldS3Key := orgS3Key(orgID, oldPath)
	newS3Key := orgS3Key(orgID, newPath)

	if err := executeOrgFolderRenameTransaction(ctx, h.DB, orgID, folderID, oldPath, newPath, req.EncryptedName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	enqueueOrgS3Rename(h.RedisClient, orgID, oldS3Key, newS3Key, true)
	h.logAudit(ctx, orgID, userID, "folder_renamed", strconv.FormatInt(folderID, 10), "folder", req.EncryptedName)

	c.JSON(http.StatusOK, gin.H{"message": "folder renamed", "new_path": newPath})
}

// executeOrgFolderRenameTransaction renames the folder and cascades path changes to all
// descendant folders and files within a single transaction.
// Uses LEFT(col, n) = prefix instead of LIKE to avoid '_' and '%' wildcard collisions
// with encrypted path segments.
func executeOrgFolderRenameTransaction(ctx context.Context, db *bun.DB, orgID, folderID int64, oldPath, newPath, newName string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction")
	}

	if _, err := tx.NewUpdate().Model((*pkg.OrgFolder)(nil)).
		Set("name = ?", newName).
		Set("path = ?", newPath).
		Set("parent_path = ?", path.Dir(newPath)).
		Where("id = ? AND org_id = ?", folderID, orgID).
		Exec(ctx); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update folder")
	}

	prefixLen := len(oldPath) + 1
	childPrefix := oldPath + "/"

	// Child folder paths (deep descendants).
	if _, err := tx.NewRaw(
		`UPDATE org_folders SET path = ? || SUBSTRING(path, ?) WHERE org_id = ? AND LEFT(path, ?) = ?`,
		newPath, prefixLen, orgID, prefixLen, childPrefix,
	).Exec(ctx); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update child folder paths")
	}

	// Child folder parent_paths (deep descendants).
	if _, err := tx.NewRaw(
		`UPDATE org_folders SET parent_path = ? || SUBSTRING(parent_path, ?) WHERE org_id = ? AND LEFT(parent_path, ?) = ?`,
		newPath, prefixLen, orgID, prefixLen, childPrefix,
	).Exec(ctx); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update deep child folder parent_paths")
	}

	// Direct child folders' parent_path.
	if _, err := tx.NewUpdate().Model((*pkg.OrgFolder)(nil)).
		Set("parent_path = ?", newPath).
		Where("org_id = ? AND parent_path = ?", orgID, oldPath).
		Exec(ctx); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update direct child folder parent_paths")
	}

	// Child file paths (deep descendants).
	if _, err := tx.NewRaw(
		`UPDATE org_files SET path = ? || SUBSTRING(path, ?) WHERE org_id = ? AND LEFT(path, ?) = ?`,
		newPath, prefixLen, orgID, prefixLen, childPrefix,
	).Exec(ctx); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update child file paths")
	}

	// Child file folder_paths (deep descendants).
	if _, err := tx.NewRaw(
		`UPDATE org_files SET folder_path = ? || SUBSTRING(folder_path, ?) WHERE org_id = ? AND LEFT(folder_path, ?) = ?`,
		newPath, prefixLen, orgID, prefixLen, childPrefix,
	).Exec(ctx); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update deep child file folder_paths")
	}

	// Direct child files' folder_path.
	if _, err := tx.NewUpdate().Model((*pkg.OrgFile)(nil)).
		Set("folder_path = ?", newPath).
		Where("org_id = ? AND folder_path = ?", orgID, oldPath).
		Exec(ctx); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update direct child file folder_paths")
	}

	return tx.Commit()
}

func enqueueOrgS3Rename(rc *redis.Client, orgID int64, srcKey, destKey string, isFolder bool) {
	if rc == nil {
		return
	}
	task := workers.S3Task{
		Type:     workers.TaskRename,
		UserID:   fmt.Sprintf("org_%d", orgID),
		SrcKey:   srcKey,
		DestKey:  destKey,
		IsFolder: isFolder,
	}
	if err := workers.EnqueueTask(rc, task); err != nil {
		log.Printf("Failed to enqueue org S3 rename: %v", err)
	}
}
