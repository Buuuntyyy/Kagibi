// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package organizations

import (
	"database/sql"
	"net/http"
	"path"
	"strconv"

	"kagibi/backend/pkg"

	"github.com/gin-gonic/gin"
)

// MoveOrgFile moves a file to a different folder within the org.
// PATCH /orgs/:orgID/fs/file/:fileID/move
func (h *OrgHandler) MoveOrgFile(c *gin.Context) {
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

	var req struct {
		NewFolderPath string `json:"new_folder_path" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	destPath := normPath(req.NewFolderPath)
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

	if file.FolderPath == destPath {
		c.JSON(http.StatusOK, gin.H{"message": "no change"})
		return
	}

	// Validate destination folder exists (skip for root).
	if destPath != "/" {
		exists, _ := h.DB.NewSelect().Model((*pkg.OrgFolder)(nil)).
			Where("org_id = ? AND path = ?", orgID, destPath).
			Exists(ctx)
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "destination folder not found"})
			return
		}
	}

	// Write access required on both source and destination.
	srcPerm, err := h.resolvePermission(ctx, orgID, userID, file.FolderPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check source permissions"})
		return
	}
	if srcPerm < PermWrite {
		c.JSON(http.StatusForbidden, gin.H{"error": "write access denied on source folder"})
		return
	}
	dstPerm, err := h.resolvePermission(ctx, orgID, userID, destPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check destination permissions"})
		return
	}
	if dstPerm < PermWrite {
		c.JSON(http.StatusForbidden, gin.H{"error": "write access denied on destination folder"})
		return
	}

	newPath := path.Join(destPath, file.Name)

	exists, _ := h.DB.NewSelect().Model((*pkg.OrgFile)(nil)).
		Where("org_id = ? AND path = ?", orgID, newPath).
		Exists(ctx)
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "a file with this name already exists at destination"})
		return
	}

	oldS3Key := orgS3Key(orgID, file.Path)
	newS3Key := orgS3Key(orgID, newPath)

	if _, err := h.DB.NewUpdate().Model(&file).
		Set("path = ?", newPath).
		Set("folder_path = ?", destPath).
		WherePK().
		Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to move file"})
		return
	}

	enqueueOrgS3Rename(h.RedisClient, orgID, oldS3Key, newS3Key, false)
	h.logAudit(ctx, orgID, userID, "file_moved", strconv.FormatInt(fileID, 10), "file", destPath)

	c.JSON(http.StatusOK, gin.H{"message": "file moved", "new_path": newPath, "new_folder_path": destPath})
}

// MoveOrgFolder moves a folder (and all its contents) to a new parent within the org.
// PATCH /orgs/:orgID/fs/folder/:folderID/move
func (h *OrgHandler) MoveOrgFolder(c *gin.Context) {
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

	var req struct {
		NewParentPath string `json:"new_parent_path" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	newParentPath := normPath(req.NewParentPath)
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

	if folder.ParentPath == newParentPath {
		c.JSON(http.StatusOK, gin.H{"message": "no change"})
		return
	}

	// Prevent moving a folder into itself or its own subtree.
	if newParentPath == folder.Path ||
		len(newParentPath) > len(folder.Path) && newParentPath[:len(folder.Path)+1] == folder.Path+"/" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot move a folder into itself or its own subtree"})
		return
	}

	// Validate destination parent exists (skip for root).
	if newParentPath != "/" {
		exists, _ := h.DB.NewSelect().Model((*pkg.OrgFolder)(nil)).
			Where("org_id = ? AND path = ?", orgID, newParentPath).
			Exists(ctx)
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "destination folder not found"})
			return
		}
	}

	srcPerm, err := h.resolvePermission(ctx, orgID, userID, folder.ParentPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check source permissions"})
		return
	}
	if srcPerm < PermWrite {
		c.JSON(http.StatusForbidden, gin.H{"error": "write access denied on source parent"})
		return
	}
	dstPerm, err := h.resolvePermission(ctx, orgID, userID, newParentPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check destination permissions"})
		return
	}
	if dstPerm < PermWrite {
		c.JSON(http.StatusForbidden, gin.H{"error": "write access denied on destination"})
		return
	}

	newPath := path.Join(newParentPath, folder.Name)

	exists, _ := h.DB.NewSelect().Model((*pkg.OrgFolder)(nil)).
		Where("org_id = ? AND path = ?", orgID, newPath).
		Exists(ctx)
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "a folder with this name already exists at destination"})
		return
	}

	oldPath := folder.Path
	oldS3Key := orgS3Key(orgID, oldPath)
	newS3Key := orgS3Key(orgID, newPath)

	if err := executeOrgFolderRenameTransaction(ctx, h.DB, orgID, folderID, oldPath, newPath, folder.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	enqueueOrgS3Rename(h.RedisClient, orgID, oldS3Key, newS3Key, true)
	h.logAudit(ctx, orgID, userID, "folder_moved", strconv.FormatInt(folderID, 10), "folder", newParentPath)

	c.JSON(http.StatusOK, gin.H{"message": "folder moved", "new_path": newPath, "new_parent_path": newParentPath})
}
