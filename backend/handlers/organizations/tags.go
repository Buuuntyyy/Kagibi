// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package organizations

import (
	"database/sql"
	"net/http"
	"strconv"

	"kagibi/backend/pkg"

	"github.com/gin-gonic/gin"
)

type createTagRequest struct {
	EncryptedName string `json:"encrypted_name" binding:"required"`
	Color         string `json:"color" binding:"required"`
}

type updateTagRequest struct {
	EncryptedName string `json:"encrypted_name"`
	Color         string `json:"color"`
}

type setItemTagsRequest struct {
	TagIDs []int64 `json:"tag_ids"`
}

func validHexColor(s string) bool {
	if len(s) != 7 || s[0] != '#' {
		return false
	}
	for _, c := range s[1:] {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// ListOrgTags returns all tags for an org. Any member may call this.
func (h *OrgHandler) ListOrgTags(c *gin.Context) {
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid org id"})
		return
	}
	userID := c.GetString("user_id")
	role, err := h.memberRole(c.Request.Context(), orgID, userID)
	if err != nil || role == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member"})
		return
	}

	var tags []pkg.OrgTag
	if err := h.DB.NewSelect().Model(&tags).
		Where("ot.org_id = ?", orgID).
		OrderExpr("ot.created_at ASC").
		Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if tags == nil {
		tags = []pkg.OrgTag{}
	}
	c.JSON(http.StatusOK, tags)
}

// CreateOrgTag creates a new tag. Any member may create tags.
func (h *OrgHandler) CreateOrgTag(c *gin.Context) {
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid org id"})
		return
	}
	userID := c.GetString("user_id")
	role, err := h.memberRole(c.Request.Context(), orgID, userID)
	if err != nil || role == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member"})
		return
	}

	var req createTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !validHexColor(req.Color) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "color must be a 7-char hex value like #ff0000"})
		return
	}

	tag := &pkg.OrgTag{OrgID: orgID, EncryptedName: req.EncryptedName, Color: req.Color}
	if _, err := h.DB.NewInsert().Model(tag).Returning("*").Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, tag)
}

// UpdateOrgTag renames or recolors a tag. Any member may update tags.
func (h *OrgHandler) UpdateOrgTag(c *gin.Context) {
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid org id"})
		return
	}
	tagID, err := strconv.ParseInt(c.Param("tagID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag id"})
		return
	}
	userID := c.GetString("user_id")
	role, err := h.memberRole(c.Request.Context(), orgID, userID)
	if err != nil || role == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member"})
		return
	}

	var req updateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var tag pkg.OrgTag
	if err := h.DB.NewSelect().Model(&tag).
		Where("ot.id = ? AND ot.org_id = ?", tagID, orgID).
		Scan(c.Request.Context()); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "tag not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	q := h.DB.NewUpdate().Model(&tag).Where("id = ? AND org_id = ?", tagID, orgID)
	if req.EncryptedName != "" {
		tag.EncryptedName = req.EncryptedName
		q = q.Set("encrypted_name = ?", req.EncryptedName)
	}
	if req.Color != "" {
		if !validHexColor(req.Color) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "color must be a 7-char hex value"})
			return
		}
		tag.Color = req.Color
		q = q.Set("color = ?", req.Color)
	}
	if _, err := q.Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tag)
}

// DeleteOrgTag deletes a tag and removes it from all items. Admins only.
func (h *OrgHandler) DeleteOrgTag(c *gin.Context) {
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid org id"})
		return
	}
	tagID, err := strconv.ParseInt(c.Param("tagID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag id"})
		return
	}
	userID := c.GetString("user_id")
	role, err := h.memberRole(c.Request.Context(), orgID, userID)
	if err != nil || role == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member"})
		return
	}
	if !canManage(role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only admins can delete tags"})
		return
	}

	ctx := c.Request.Context()
	if _, err := h.DB.ExecContext(ctx,
		`UPDATE org_files SET tag_ids = array_remove(tag_ids, ?) WHERE org_id = ?`, tagID, orgID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if _, err := h.DB.ExecContext(ctx,
		`UPDATE org_folders SET tag_ids = array_remove(tag_ids, ?) WHERE org_id = ?`, tagID, orgID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if _, err := h.DB.NewDelete().TableExpr("org_tags").
		Where("id = ? AND org_id = ?", tagID, orgID).
		Exec(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

// SetFileTags sets the tag list on a file. Any org member may tag items.
func (h *OrgHandler) SetFileTags(c *gin.Context) {
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid org id"})
		return
	}
	fileID, err := strconv.ParseInt(c.Param("fileID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file id"})
		return
	}
	userID := c.GetString("user_id")
	role, err := h.memberRole(c.Request.Context(), orgID, userID)
	if err != nil || role == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member"})
		return
	}

	var req setItemTagsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.TagIDs == nil {
		req.TagIDs = []int64{}
	}

	var file pkg.OrgFile
	if err := h.DB.NewSelect().Model(&file).
		Where("ofile.id = ? AND ofile.org_id = ?", fileID, orgID).
		Scan(c.Request.Context()); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	file.TagIDs = req.TagIDs
	if _, err := h.DB.NewUpdate().Model(&file).Column("tag_ids").WherePK().Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tag_ids": req.TagIDs})
}

// SetFolderTags sets the tag list on a folder. Any org member may tag items.
func (h *OrgHandler) SetFolderTags(c *gin.Context) {
	orgID, err := strconv.ParseInt(c.Param("orgID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid org id"})
		return
	}
	folderID, err := strconv.ParseInt(c.Param("folderID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid folder id"})
		return
	}
	userID := c.GetString("user_id")
	role, err := h.memberRole(c.Request.Context(), orgID, userID)
	if err != nil || role == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member"})
		return
	}

	var req setItemTagsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.TagIDs == nil {
		req.TagIDs = []int64{}
	}

	var folder pkg.OrgFolder
	if err := h.DB.NewSelect().Model(&folder).
		Where("of.id = ? AND of.org_id = ?", folderID, orgID).
		Scan(c.Request.Context()); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "folder not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	folder.TagIDs = req.TagIDs
	if _, err := h.DB.NewUpdate().Model(&folder).Column("tag_ids").WherePK().Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tag_ids": req.TagIDs})
}
