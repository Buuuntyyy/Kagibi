// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package shares

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"kagibi/backend/pkg"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type ShareResponse struct {
	ID           int64      `json:"id"`
	Token        string     `json:"token"`
	Link         string     `json:"link"`
	ResourceType string     `json:"resource_type"`
	ResourceName string     `json:"resource_name"`
	ResourceID   int64      `json:"resource_id"`
	ResourcePath string     `json:"resource_path"`
	Views        int64      `json:"views"`
	ExpiresAt    *time.Time `json:"expires_at"`
	CreatedAt    time.Time  `json:"created_at"`
	PermDownload bool       `json:"perm_download"`
	PermCreate   bool       `json:"perm_create"`
	PermDelete   bool       `json:"perm_delete"`
	PermMove     bool       `json:"perm_move"`
}

// ListSharesHandler lists all active share links created by the user
func ListSharesHandler(c *gin.Context, db *bun.DB) {
	userID := c.GetString("user_id")

	links := fetchAndProcessShareLinks(c.Request.Context(), db, userID)
	directFiles := fetchAndProcessDirectFileShares(c.Request.Context(), db, userID)
	directFolders := fetchAndProcessDirectFolderShares(c.Request.Context(), db, userID)

	response := append(links, directFiles...)
	response = append(response, directFolders...)

	c.JSON(http.StatusOK, gin.H{"shares": response})
}

func fetchAndProcessShareLinks(ctx context.Context, db *bun.DB, userID string) []ShareResponse {
	var links []pkg.ShareLink
	err := db.NewSelect().Model(&links).
		Where("owner_id = ?", userID).
		Order("created_at DESC").
		Scan(ctx)

	if err != nil {
		return []ShareResponse{}
	}

	var response []ShareResponse
	for _, l := range links {
		name, exists := checkResourceExists(ctx, db, l.ResourceType, l.ResourceID)
		if !exists {
			go deleteOrphanedShareLink(l.ID, db)
			continue
		}

		response = append(response, ShareResponse{
			ID:           l.ID,
			Token:        l.Token,
			Link:         fmt.Sprintf("/s/%s", l.Token),
			ResourceType: l.ResourceType,
			ResourceName: name,
			ResourceID:   l.ResourceID,
			ResourcePath: l.Path,
			Views:        l.Views,
			ExpiresAt:    l.ExpiresAt,
			CreatedAt:    l.CreatedAt,
			PermDownload: l.PermDownload,
			PermCreate:   l.PermCreate,
			PermDelete:   l.PermDelete,
			PermMove:     l.PermMove,
		})
	}
	return response
}

func fetchAndProcessDirectFileShares(ctx context.Context, db *bun.DB, userID string) []ShareResponse {
	var fileShares []pkg.FileShare
	err := db.NewSelect().Model(&fileShares).
		Join("JOIN files ON files.id = file_share.file_id").
		Where("files.user_id = ?", userID).
		Scan(ctx)

	if err != nil {
		return []ShareResponse{}
	}

	var response []ShareResponse
	for _, fs := range fileShares {
		var f pkg.File
		if err := db.NewSelect().Model(&f).Where(queryIDEq, fs.FileID).Scan(ctx); err != nil {
			continue
		}

		friendName := getFriendName(ctx, db, fs.SharedWithUserID)

		response = append(response, ShareResponse{
			ID:           fs.ID,
			Token:        "DIRECT",
			Link:         "Shared with " + friendName,
			ResourceType: "file",
			ResourceName: f.Name,
			ResourceID:   f.ID,
			Views:        0,
			CreatedAt:    fs.CreatedAt,
			PermDownload: fs.PermDownload,
		})
	}
	return response
}

func fetchAndProcessDirectFolderShares(ctx context.Context, db *bun.DB, userID string) []ShareResponse {
	var folderShares []pkg.FolderShare
	err := db.NewSelect().Model(&folderShares).
		Join("JOIN folders ON folders.id = folder_share.folder_id").
		Where("folders.user_id = ?", userID).
		Scan(ctx)

	if err != nil {
		return []ShareResponse{}
	}

	var response []ShareResponse
	for _, fs := range folderShares {
		var f pkg.Folder
		if err := db.NewSelect().Model(&f).Where(queryIDEq, fs.FolderID).Scan(ctx); err != nil {
			continue
		}

		friendName := getFriendName(ctx, db, fs.SharedWithUserID)

		response = append(response, ShareResponse{
			ID:           fs.ID,
			Token:        "DIRECT",
			Link:         "Shared with " + friendName,
			ResourceType: "folder",
			ResourceName: f.Name,
			ResourceID:   f.ID,
			ResourcePath: f.Path,
			Views:        0,
			CreatedAt:    fs.CreatedAt,
			PermDownload: fs.PermDownload,
			PermCreate:   fs.PermCreate,
			PermDelete:   fs.PermDelete,
			PermMove:     fs.PermMove,
		})
	}
	return response
}

func checkResourceExists(ctx context.Context, db *bun.DB, resType string, resID int64) (string, bool) {
	if resType == "file" {
		var f pkg.File
		if err := db.NewSelect().Model(&f).Where(queryIDEq, resID).Scan(ctx); err == nil {
			return f.Name, true
		}
	} else if resType == "folder" {
		var f pkg.Folder
		if err := db.NewSelect().Model(&f).Where(queryIDEq, resID).Scan(ctx); err == nil {
			return f.Name, true
		}
	}
	return "Unknown", false
}

func deleteOrphanedShareLink(id int64, db *bun.DB) {
	db.NewDelete().Model((*pkg.ShareLink)(nil)).Where(queryIDEq, id).Exec(context.Background())
}

func getFriendName(ctx context.Context, db *bun.DB, userID string) string {
	var friend pkg.User
	if err := db.NewSelect().Model(&friend).Where(queryIDEq, userID).Scan(ctx); err == nil {
		return friend.Name
	}
	return "Unknown"
}
