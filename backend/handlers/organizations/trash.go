// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package organizations

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"kagibi/backend/pkg"
	"kagibi/backend/pkg/s3storage"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3svc "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

// TrashItem is the DTO returned by ListTrash.
type TrashItem struct {
	ID        int64      `json:"id"`
	ItemType  string     `json:"item_type"`
	Name      string     `json:"name"`
	Path      string     `json:"path"`
	DeletedAt *time.Time `json:"deleted_at"`
	DeletedBy string     `json:"deleted_by"`
	Size      int64      `json:"size,omitempty"`
	MimeType  string     `json:"mime_type,omitempty"`
}

func s3DeleteOrgFile(ctx context.Context, orgID int64, filePath string) {
	s3Key := fmt.Sprintf("orgs/%d%s", orgID, filePath)
	if _, err := s3storage.Client.DeleteObject(ctx, &s3svc.DeleteObjectInput{
		Bucket: aws.String(s3storage.BucketName),
		Key:    aws.String(s3Key),
	}); err != nil {
		log.Printf("Warning: s3 delete org file %s: %v", s3Key, err)
	}
}

// ListTrash returns all top-level soft-deleted items (delete_root = true) for this org.
func (h *OrgHandler) ListTrash(c *gin.Context) {
	userID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid org id"})
		return
	}
	ctx := c.Request.Context()

	role, err := h.memberRole(ctx, orgID, userID)
	if err != nil || role == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member"})
		return
	}

	items := make([]TrashItem, 0, 16)

	var folders []pkg.OrgFolder
	if err := h.DB.NewSelect().Model(&folders).
		WhereAllWithDeleted().
		Where("org_id = ? AND deleted_at IS NOT NULL AND delete_root = TRUE", orgID).
		OrderExpr("deleted_at DESC").
		Scan(ctx); err == nil {
		for _, f := range folders {
			items = append(items, TrashItem{ID: f.ID, ItemType: "folder", Name: f.Name, Path: f.Path, DeletedAt: f.DeletedAt, DeletedBy: f.DeletedBy})
		}
	}

	var files []pkg.OrgFile
	if err := h.DB.NewSelect().Model(&files).
		WhereAllWithDeleted().
		Where("org_id = ? AND deleted_at IS NOT NULL AND delete_root = TRUE", orgID).
		OrderExpr("deleted_at DESC").
		Scan(ctx); err == nil {
		for _, f := range files {
			items = append(items, TrashItem{ID: f.ID, ItemType: "file", Name: f.Name, Path: f.Path, DeletedAt: f.DeletedAt, DeletedBy: f.DeletedBy, Size: f.Size, MimeType: f.MimeType})
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

// RestoreItem un-deletes a single trash item (and its cascaded contents for folders).
func (h *OrgHandler) RestoreItem(c *gin.Context) {
	userID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid org id"})
		return
	}
	itemType := c.Param("itemType")
	itemID, err := strconv.ParseInt(c.Param("itemID"), 10, 64)
	if err != nil || (itemType != "file" && itemType != "folder") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid params"})
		return
	}
	ctx := c.Request.Context()

	role, err := h.memberRole(ctx, orgID, userID)
	if err != nil || role == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member"})
		return
	}

	if itemType == "file" {
		var file pkg.OrgFile
		if err := h.DB.NewSelect().Model(&file).WhereAllWithDeleted().
			Where("id = ? AND org_id = ? AND deleted_at IS NOT NULL AND delete_root = TRUE", itemID, orgID).
			Scan(ctx); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "item not found in trash"})
			return
		}
		if _, err := h.DB.NewUpdate().Model((*pkg.OrgFile)(nil)).
			Set("deleted_at = NULL, deleted_by = '', delete_root = FALSE").
			Where("id = ?", itemID).
			Exec(ctx); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to restore"})
			return
		}
		_, _ = h.DB.NewUpdate().Model((*pkg.Organization)(nil)).
			Set("storage_used_bytes = storage_used_bytes + ?", file.Size).
			Where("id = ?", orgID).
			Exec(ctx)
		h.logAudit(ctx, orgID, userID, "file_restored", strconv.FormatInt(itemID, 10), "file", file.Name)
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	var folder pkg.OrgFolder
	if err := h.DB.NewSelect().Model(&folder).WhereAllWithDeleted().
		Where("id = ? AND org_id = ? AND deleted_at IS NOT NULL AND delete_root = TRUE", itemID, orgID).
		Scan(ctx); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found in trash"})
		return
	}
	if _, err := h.DB.NewUpdate().Model((*pkg.OrgFolder)(nil)).
		Set("deleted_at = NULL, deleted_by = '', delete_root = FALSE").
		Where("id = ?", itemID).
		Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to restore folder"})
		return
	}
	_, _ = h.DB.NewUpdate().Model((*pkg.OrgFolder)(nil)).
		Set("deleted_at = NULL, deleted_by = '', delete_root = FALSE").
		WhereAllWithDeleted().
		Where("org_id = ? AND path LIKE ? AND delete_root = FALSE AND deleted_at IS NOT NULL", orgID, folder.Path+"/%").
		Exec(ctx)
	_, _ = h.DB.NewUpdate().Model((*pkg.OrgFile)(nil)).
		Set("deleted_at = NULL, deleted_by = '', delete_root = FALSE").
		WhereAllWithDeleted().
		Where("org_id = ? AND (folder_path = ? OR folder_path LIKE ?) AND delete_root = FALSE AND deleted_at IS NOT NULL", orgID, folder.Path, folder.Path+"/%").
		Exec(ctx)

	var restored struct{ Sum int64 }
	_ = h.DB.NewSelect().TableExpr("org_files").
		ColumnExpr("COALESCE(SUM(size), 0) AS sum").
		Where("org_id = ? AND (folder_path = ? OR folder_path LIKE ?) AND deleted_at IS NULL", orgID, folder.Path, folder.Path+"/%").
		Scan(ctx, &restored)
	if restored.Sum > 0 {
		_, _ = h.DB.NewUpdate().Model((*pkg.Organization)(nil)).
			Set("storage_used_bytes = storage_used_bytes + ?", restored.Sum).
			Where("id = ?", orgID).
			Exec(ctx)
	}
	h.logAudit(ctx, orgID, userID, "folder_restored", strconv.FormatInt(itemID, 10), "folder", folder.Name)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// PermanentDeleteItem permanently removes a trash item and its S3 data. Admin/owner only.
func (h *OrgHandler) PermanentDeleteItem(c *gin.Context) {
	userID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid org id"})
		return
	}
	itemType := c.Param("itemType")
	itemID, err := strconv.ParseInt(c.Param("itemID"), 10, 64)
	if err != nil || (itemType != "file" && itemType != "folder") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid params"})
		return
	}
	ctx := c.Request.Context()

	role, err := h.memberRole(ctx, orgID, userID)
	if err != nil || !isAdminOrOwner(role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	if itemType == "file" {
		var file pkg.OrgFile
		if err := h.DB.NewSelect().Model(&file).WhereAllWithDeleted().
			Where("id = ? AND org_id = ? AND deleted_at IS NOT NULL", itemID, orgID).
			Scan(ctx); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "item not found in trash"})
			return
		}
		s3DeleteOrgFile(ctx, orgID, file.Path)
		_, _ = h.DB.NewDelete().Model((*pkg.OrgFile)(nil)).
			WhereAllWithDeleted().Where("id = ?", itemID).ForceDelete().Exec(ctx)
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	var folder pkg.OrgFolder
	if err := h.DB.NewSelect().Model(&folder).WhereAllWithDeleted().
		Where("id = ? AND org_id = ? AND deleted_at IS NOT NULL", itemID, orgID).
		Scan(ctx); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found in trash"})
		return
	}
	var files []pkg.OrgFile
	_ = h.DB.NewSelect().Model(&files).WhereAllWithDeleted().
		Where("org_id = ? AND (folder_path = ? OR folder_path LIKE ?) AND deleted_at IS NOT NULL", orgID, folder.Path, folder.Path+"/%").
		Scan(ctx)
	for _, f := range files {
		s3DeleteOrgFile(ctx, orgID, f.Path)
	}
	_, _ = h.DB.NewDelete().Model((*pkg.OrgFile)(nil)).
		WhereAllWithDeleted().
		Where("org_id = ? AND (folder_path = ? OR folder_path LIKE ?) AND deleted_at IS NOT NULL", orgID, folder.Path, folder.Path+"/%").
		ForceDelete().Exec(ctx)
	_, _ = h.DB.NewDelete().Model((*pkg.OrgFolder)(nil)).
		WhereAllWithDeleted().
		Where("org_id = ? AND (id = ? OR (path LIKE ? AND deleted_at IS NOT NULL))", orgID, itemID, folder.Path+"/%").
		ForceDelete().Exec(ctx)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// EmptyTrash permanently deletes all soft-deleted items for this org. Admin/owner only.
func (h *OrgHandler) EmptyTrash(c *gin.Context) {
	userID := c.GetString("user_id")
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid org id"})
		return
	}
	ctx := c.Request.Context()

	role, err := h.memberRole(ctx, orgID, userID)
	if err != nil || !isAdminOrOwner(role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	var files []pkg.OrgFile
	_ = h.DB.NewSelect().Model(&files).WhereAllWithDeleted().
		Where("org_id = ? AND deleted_at IS NOT NULL", orgID).
		Scan(ctx)
	for _, f := range files {
		s3DeleteOrgFile(ctx, orgID, f.Path)
	}
	_, _ = h.DB.NewDelete().Model((*pkg.OrgFile)(nil)).
		WhereAllWithDeleted().Where("org_id = ? AND deleted_at IS NOT NULL", orgID).ForceDelete().Exec(ctx)
	_, _ = h.DB.NewDelete().Model((*pkg.OrgFolder)(nil)).
		WhereAllWithDeleted().Where("org_id = ? AND deleted_at IS NOT NULL", orgID).ForceDelete().Exec(ctx)

	h.logAudit(ctx, orgID, userID, "trash_emptied", "", "org", "")
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func isAdminOrOwner(role string) bool {
	return role == "admin" || role == "owner"
}
