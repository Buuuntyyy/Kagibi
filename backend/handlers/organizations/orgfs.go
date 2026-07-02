// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package organizations

import (
	"database/sql"
	"log"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"kagibi/backend/pkg"

	"github.com/gin-gonic/gin"
)

var forbiddenOrgFolderChars = regexp.MustCompile(`[/\\\x00-\x1f<>]`)

// ListOrgItems returns files and folders at a given path within the org.
func (h *OrgHandler) ListOrgItems(c *gin.Context) {
	userID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}

	if !h.checkMFAEnforcement(c, orgID, userID) {
		return
	}

	rawPath := c.Param("path")
	if rawPath == "" {
		rawPath = "/"
	}
	folderPath := normPath(rawPath)
	ctx := c.Request.Context()

	perm, err := h.resolvePermission(ctx, orgID, userID, folderPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check permissions"})
		return
	}
	if perm < PermRead {
		c.JSON(http.StatusForbidden, gin.H{"error": "read access denied"})
		return
	}

	var folders []pkg.OrgFolder
	if err := h.DB.NewSelect().Model(&folders).
		Where("org_id = ? AND parent_path = ?", orgID, folderPath).
		OrderExpr("name ASC").
		Scan(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list folders"})
		return
	}

	// For each sub-folder check whether the caller can read it. If not, mark it
	// locked so the frontend can show it as inaccessible (with a request-access CTA)
	// rather than hiding it entirely.
	if len(folders) > 0 {
		// Fetch caller's pending access requests in one query.
		var pendingPaths []struct {
			FolderPath string `bun:"folder_path"`
		}
		_ = h.DB.NewSelect().
			TableExpr("org_access_requests").
			ColumnExpr("folder_path").
			Where("org_id = ? AND user_id = ? AND status = 'pending'", orgID, userID).
			Scan(ctx, &pendingPaths)
		pendingSet := make(map[string]bool, len(pendingPaths))
		for _, p := range pendingPaths {
			pendingSet[p.FolderPath] = true
		}

		for i := range folders {
			subPerm, err := h.resolvePermission(ctx, orgID, userID, folders[i].Path)
			if err == nil && subPerm < PermRead {
				folders[i].Locked = true
				folders[i].AccessRequestPending = pendingSet[folders[i].Path]
			}
		}
	}

	var files []pkg.OrgFile
	if err := h.DB.NewSelect().Model(&files).
		Where("org_id = ? AND folder_path = ?", orgID, folderPath).
		OrderExpr("name ASC").
		Scan(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list files"})
		return
	}

	// Compute recursive folder sizes: one query fetches all file sizes under the
	// current path, then we sum per-folder in Go using path-prefix matching.
	if len(folders) > 0 {
		type pathSize struct {
			FolderPath string `bun:"folder_path"`
			Total      int64  `bun:"total"`
		}
		var sizes []pathSize
		if err := h.DB.NewSelect().
			TableExpr("org_files").
			ColumnExpr("folder_path, SUM(size) AS total").
			Where("org_id = ? AND folder_path LIKE ?", orgID, folderPath+"%").
			GroupExpr("folder_path").
			Scan(ctx, &sizes); err == nil {
			// Build a map of folder_path → direct-file total
			sizeMap := make(map[string]int64, len(sizes))
			for _, ps := range sizes {
				sizeMap[ps.FolderPath] = ps.Total
			}
			// Assign recursive totals to each sub-folder
			for i := range folders {
				prefix := folders[i].Path
				var total int64
				for p, s := range sizeMap {
					if len(p) >= len(prefix) && p[:len(prefix)] == prefix {
						total += s
					}
				}
				folders[i].TotalSize = total
			}
		}
	}

	// Resolve the group_id of the current folder (if any) so the client knows
	// which key to use when encrypting new uploads into this path.
	var currentGroupID *int64
	if folderPath != "/" {
		var curFolder pkg.OrgFolder
		if err := h.DB.NewSelect().Model(&curFolder).
			Column("group_id").
			Where("org_id = ? AND path = ?", orgID, folderPath).
			Scan(ctx); err == nil {
			currentGroupID = curFolder.GroupID
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"folders":          folders,
		"files":            files,
		"current_path":     folderPath,
		"current_group_id": currentGroupID,
	})
}

// CreateOrgFolder creates a new folder inside the org's shared storage.
func (h *OrgHandler) CreateOrgFolder(c *gin.Context) {
	userID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}

	var req struct {
		Name         string  `json:"name" binding:"required"`
		ParentPath   string  `json:"parent_path"`
		EncryptedKey string  `json:"encrypted_key"`
		GroupID      *int64  `json:"group_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	trimmedName := strings.TrimSpace(req.Name)
	if trimmedName == "" || trimmedName == "." || trimmedName == ".." || forbiddenOrgFolderChars.MatchString(req.Name) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid folder name"})
		return
	}

	if !h.checkMFAEnforcement(c, orgID, userID) {
		return
	}

	parentPath := normPath(req.ParentPath)
	folderPath := normPath(path.Join(parentPath, req.Name))
	ctx := c.Request.Context()

	perm, err := h.resolvePermission(ctx, orgID, userID, parentPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check permissions"})
		return
	}
	if perm < PermWrite {
		c.JSON(http.StatusForbidden, gin.H{"error": "write access denied"})
		return
	}

	// Explicit existence check so the error is never ambiguous.
	var existing pkg.OrgFolder
	err = h.DB.NewSelect().Model(&existing).
		Where("org_id = ? AND path = ?", orgID, folderPath).
		Limit(1).Scan(ctx)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "folder already exists at this path"})
		return
	}

	// If no group_id was provided, inherit it from the parent folder so sub-folders
	// inside a group-encrypted folder automatically belong to the same group (B1+B2 fix).
	if req.GroupID == nil && parentPath != "/" {
		var parentFolder pkg.OrgFolder
		if err := h.DB.NewSelect().Model(&parentFolder).
			Column("group_id").
			Where("org_id = ? AND path = ?", orgID, parentPath).
			Scan(ctx); err == nil {
			req.GroupID = parentFolder.GroupID
		}
	}

	folder := &pkg.OrgFolder{
		OrgID:        orgID,
		Name:         req.Name,
		Path:         folderPath,
		ParentPath:   parentPath,
		CreatedBy:    userID,
		EncryptedKey: req.EncryptedKey,
		GroupID:      req.GroupID,
		TagIDs:       []int64{},
	}
	if _, err := h.DB.NewInsert().Model(folder).Exec(ctx); err != nil {
		log.Printf("[CreateOrgFolder] insert error orgID=%d path=%q: %v", orgID, folderPath, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create folder"})
		return
	}
	c.JSON(http.StatusCreated, folder)
}

// DeleteOrgFolder soft-deletes a folder and all its contents.
func (h *OrgHandler) DeleteOrgFolder(c *gin.Context) {
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

	perm, err := h.resolvePermission(ctx, orgID, userID, folder.ParentPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check permissions"})
		return
	}
	if perm < PermWrite {
		c.JSON(http.StatusForbidden, gin.H{"error": "delete access denied"})
		return
	}

	// Compute total size of non-deleted files under this path for quota adjustment
	var totalSize struct{ Sum int64 }
	_ = h.DB.NewSelect().
		TableExpr("org_files").
		ColumnExpr("COALESCE(SUM(size), 0) AS sum").
		Where("org_id = ? AND (folder_path = ? OR folder_path LIKE ?) AND deleted_at IS NULL", orgID, folder.Path, folder.Path+"/%").
		Scan(ctx, &totalSize)

	now := time.Now().UTC()
	// Soft-delete the target folder itself (delete_root = true)
	_, _ = h.DB.NewUpdate().Model((*pkg.OrgFolder)(nil)).
		Set("deleted_at = ?, deleted_by = ?, delete_root = TRUE", now, userID).
		Where("id = ?", folderID).
		Exec(ctx)
	// Soft-delete nested sub-folders (cascaded, delete_root = false)
	_, _ = h.DB.NewUpdate().Model((*pkg.OrgFolder)(nil)).
		Set("deleted_at = ?, deleted_by = ?, delete_root = FALSE", now, userID).
		Where("org_id = ? AND path LIKE ? AND deleted_at IS NULL", orgID, folder.Path+"/%").
		Exec(ctx)
	// Soft-delete nested files (cascaded, delete_root = false)
	_, _ = h.DB.NewUpdate().Model((*pkg.OrgFile)(nil)).
		Set("deleted_at = ?, deleted_by = ?, delete_root = FALSE", now, userID).
		Where("org_id = ? AND (folder_path = ? OR folder_path LIKE ?) AND deleted_at IS NULL", orgID, folder.Path, folder.Path+"/%").
		Exec(ctx)

	if totalSize.Sum > 0 {
		_, _ = h.DB.NewUpdate().Model((*pkg.Organization)(nil)).
			Set("storage_used_bytes = GREATEST(0, storage_used_bytes - ?)", totalSize.Sum).
			Where("id = ?", orgID).
			Exec(ctx)
	}

	h.logAudit(ctx, orgID, userID, "folder_deleted", strconv.FormatInt(folderID, 10), "folder", folder.Name)
	c.JSON(http.StatusOK, gin.H{"message": "folder deleted", "freed_bytes": totalSize.Sum})
}
