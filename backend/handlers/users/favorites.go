// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package users

import (
	"kagibi/backend/pkg"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type AddFavoriteRequest struct {
	ID   int64  `json:"id" binding:"required"`
	Type string `json:"type" binding:"required,oneof=file folder"`
}

func ListUserFavoritesHandler(c *gin.Context, db *bun.DB) {
	userID := c.MustGet("user_id").(string)

	var favs []pkg.UserFavorite
	err := db.NewSelect().Model(&favs).
		Where("?TableAlias.user_id = ?", userID).
		Relation("File").
		Relation("Folder").
		Order("created_at ASC").
		Scan(c.Request.Context())

	if err != nil {
		log.Printf("Failed to fetch user favorites: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch favorites"})
		return
	}

	result := []gin.H{}
	for _, fav := range favs {
		if fav.File != nil {
			result = append(result, gin.H{
				"type":          "file",
				"id":            fav.File.ID,
				"name":          fav.File.Name,
				"path":          fav.File.Path,
				"size":          fav.File.Size,
				"mime_type":     fav.File.MimeType,
				"updated_at":    fav.File.UpdatedAt,
				"encrypted_key": fav.File.EncryptedKey,
				"file":          fav.File,
			})
		} else if fav.Folder != nil {
			result = append(result, gin.H{
				"type":       "folder",
				"id":         fav.Folder.ID,
				"name":       fav.Folder.Name,
				"path":       fav.Folder.Path,
				"tags":       fav.Folder.Tags,
				"updated_at": fav.Folder.UpdatedAt,
				"folder":     fav.Folder,
			})
		}
	}

	c.JSON(http.StatusOK, result)
}

func AddUserFavoriteHandler(c *gin.Context, db *bun.DB) {
	userID := c.MustGet("user_id").(string)

	var req AddFavoriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	fav := &pkg.UserFavorite{UserID: userID}
	if req.Type == "file" {
		fav.FileID = &req.ID
	} else {
		fav.FolderID = &req.ID
	}

	_, err := db.NewInsert().Model(fav).
		On("CONFLICT DO NOTHING").
		Exec(c.Request.Context())
	if err != nil {
		log.Printf("Failed to add user favorite: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add favorite"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func RemoveUserFavoriteHandler(c *gin.Context, db *bun.DB) {
	userID := c.MustGet("user_id").(string)
	itemType := c.Param("type")
	itemIDStr := c.Param("id")

	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	q := db.NewDelete().Model((*pkg.UserFavorite)(nil)).Where("user_id = ?", userID)
	if itemType == "file" {
		q = q.Where("file_id = ?", itemID)
	} else {
		q = q.Where("folder_id = ?", itemID)
	}

	res, err := q.Exec(c.Request.Context())
	if err != nil {
		log.Printf("Failed to remove user favorite: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove favorite"})
		return
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Favorite not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
