package shares

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"safercloud/backend/pkg"
	"safercloud/backend/pkg/ws"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type SharedWithMeResponse struct {
	ID           int64     `json:"id"`
	ShareLinkID  int64     `json:"share_link_id,omitempty"`
	Token        string    `json:"token,omitempty"`
	ResourceType string    `json:"type"`
	Name         string    `json:"name"`
	OwnerName    string    `json:"owner_name"`
	SharedAt     time.Time `json:"shared_at"`
	Size         int64     `json:"size"`
	Link         string    `json:"link,omitempty"`
	FileID       int64     `json:"file_id,omitempty"`
	FolderID     int64     `json:"folder_id,omitempty"`
	EncryptedKey string    `json:"encrypted_key"`
}

// ListImportedSharesHandler lists shares that have been shared with the user (imported)
func ListImportedSharesHandler(c *gin.Context, db *bun.DB) {
	userID := c.GetString("user_id")

	linkShares := fetchLinkShares(c.Request.Context(), db, userID)
	fileShares := fetchDirectFileShares(c.Request.Context(), db, userID)
	folderShares := fetchDirectFolderShares(c.Request.Context(), db, userID)

	response := append(linkShares, fileShares...)
	response = append(response, folderShares...)

	c.JSON(http.StatusOK, response)
}

func fetchLinkShares(ctx context.Context, db *bun.DB, userID string) []SharedWithMeResponse {
	var importedShares []pkg.ImportedShare
	err := db.NewSelect().Model(&importedShares).
		Relation("ShareLink").
		Where("ish.user_id = ?", userID).
		Order("ish.created_at DESC").
		Scan(ctx)

	if err != nil {
		log.Printf("Error fetching imported shares: %v", err)
		return []SharedWithMeResponse{}
	}

	var results []SharedWithMeResponse
	for _, is := range importedShares {
		if is.ShareLink == nil {
			continue
		}
		sl := is.ShareLink
		name, size := getResourceDetails(ctx, db, sl.ResourceType, sl.ResourceID)
		ownerName := getOwnerName(ctx, db, sl.OwnerID)

		results = append(results, SharedWithMeResponse{
			ID:           is.ID,
			ShareLinkID:  sl.ID,
			Token:        sl.Token,
			ResourceType: sl.ResourceType,
			Name:         name,
			OwnerName:    ownerName,
			SharedAt:     is.CreatedAt,
			Size:         size,
			Link:         "/s/" + sl.Token,
		})
	}
	return results
}

func fetchDirectFileShares(ctx context.Context, db *bun.DB, userID string) []SharedWithMeResponse {
	var fileShares []pkg.FileShare
	if err := db.NewSelect().Model(&fileShares).Where("shared_with_user_id = ?", userID).Scan(ctx); err != nil {
		return []SharedWithMeResponse{}
	}

	var results []SharedWithMeResponse
	for _, fs := range fileShares {
		var file pkg.File
		if err := db.NewSelect().Model(&file).Where("id = ?", fs.FileID).Scan(ctx); err == nil {
			results = append(results, SharedWithMeResponse{
				ID:           fs.ID,
				ResourceType: "file",
				Name:         file.Name,
				OwnerName:    getOwnerName(ctx, db, file.UserID),
				SharedAt:     fs.CreatedAt,
				Size:         file.Size,
				FileID:       file.ID,
				EncryptedKey: fs.EncryptedKey,
			})
		}
	}
	return results
}

func fetchDirectFolderShares(ctx context.Context, db *bun.DB, userID string) []SharedWithMeResponse {
	var folderShares []pkg.FolderShare
	if err := db.NewSelect().Model(&folderShares).Where("shared_with_user_id = ?", userID).Scan(ctx); err != nil {
		return []SharedWithMeResponse{}
	}

	var results []SharedWithMeResponse
	for _, fs := range folderShares {
		var folder pkg.Folder
		if err := db.NewSelect().Model(&folder).Where("id = ?", fs.FolderID).Scan(ctx); err == nil {
			results = append(results, SharedWithMeResponse{
				ID:           fs.ID,
				ResourceType: "folder",
				Name:         folder.Name,
				OwnerName:    getOwnerName(ctx, db, folder.UserID),
				SharedAt:     fs.CreatedAt,
				Size:         0,
				FolderID:     folder.ID,
				EncryptedKey: fs.EncryptedKey,
			})
		}
	}
	return results
}

func getResourceDetails(ctx context.Context, db *bun.DB, resType string, resID int64) (string, int64) {
	if resType == "file" {
		var f pkg.File
		if err := db.NewSelect().Model(&f).Where("id = ?", resID).Scan(ctx); err == nil {
			return f.Name, f.Size
		}
	} else if resType == "folder" {
		var f pkg.Folder
		if err := db.NewSelect().Model(&f).Where("id = ?", resID).Scan(ctx); err == nil {
			return f.Name, 0
		}
	}
	return "Unknown", 0
}

func getOwnerName(ctx context.Context, db *bun.DB, userID string) string {
	var owner pkg.User
	if err := db.NewSelect().Model(&owner).Where("id = ?", userID).Scan(ctx); err == nil {
		return owner.Name
	}
	return "Unknown"
}

// ImportShareHandler adds a share link to the user's "Shared With Me" list
func ImportShareHandler(c *gin.Context, db *bun.DB) {
	userID := c.GetString("user_id")

	var req struct {
		Token string `json:"token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Find the share link
	var shareLink pkg.ShareLink
	err := db.NewSelect().Model(&shareLink).Where("token = ?", req.Token).Scan(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Share link not found"})
		return
	}

	// Check if already imported
	exists, err := db.NewSelect().Model((*pkg.ImportedShare)(nil)).
		Where("user_id = ? AND share_link_id = ?", userID, shareLink.ID).
		Exists(c.Request.Context())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Already imported"})
		return
	}

	// Import it
	importedShare := &pkg.ImportedShare{
		UserID:      userID,
		ShareLinkID: shareLink.ID,
	}

	_, err = db.NewInsert().Model(importedShare).Exec(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to import share"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Share imported successfully"})
}

// RemoveImportedShareHandler removes a share from the "Shared With Me" list
func RemoveImportedShareHandler(c *gin.Context, db *bun.DB, wsManager *ws.Manager) {
	userID := c.GetString("user_id")
	id := sanitizeID(c.Param("id"))
	shareType := sanitizeID(c.Query("type"))

	var err error
	var ownerIDToNotify string

	switch shareType {
	case "direct_file":
		ownerIDToNotify, err = removeDirectFileShare(c.Request.Context(), db, id, userID)
	case "direct_folder":
		ownerIDToNotify, err = removeDirectFolderShare(c.Request.Context(), db, id, userID)
	default:
		err = removeImportedShareLink(c.Request.Context(), db, id, userID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if ownerIDToNotify != "" {
		wsManager.SendToUser(ownerIDToNotify, ws.MsgStorageUpdate, map[string]interface{}{
			"action": "share_revoked_by_recipient",
		})
	}

	// Also notify self
	wsManager.SendToUser(userID, ws.MsgStorageUpdate, map[string]interface{}{
		"action": "share_removed_from_imported",
	})

	c.JSON(http.StatusOK, gin.H{"message": "Removed from shared with me"})
}

func sanitizeID(id string) string {
	return strings.ReplaceAll(strings.ReplaceAll(id, "\n", "_"), "\r", "_")
}

func removeDirectFileShare(ctx context.Context, db *bun.DB, id, userID string) (string, error) {
	var fs pkg.FileShare
	if err := db.NewSelect().Model(&fs).Where("id = ? AND shared_with_user_id = ?", id, userID).Scan(ctx); err != nil {
		return "", fmt.Errorf("Share not found")
	}

	res, err := db.NewDelete().Model((*pkg.FileShare)(nil)).Where("id = ? AND shared_with_user_id = ?", id, userID).Exec(ctx)
	if err != nil {
		return "", err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return "", fmt.Errorf("Share not found")
	}

	var file pkg.File
	if err := db.NewSelect().Model(&file).Where("id = ?", fs.FileID).Scan(ctx); err == nil {
		return file.UserID, nil
	}
	return "", nil
}

func removeDirectFolderShare(ctx context.Context, db *bun.DB, id, userID string) (string, error) {
	var fs pkg.FolderShare
	if err := db.NewSelect().Model(&fs).Where("id = ? AND shared_with_user_id = ?", id, userID).Scan(ctx); err != nil {
		return "", fmt.Errorf("Share not found")
	}

	res, err := db.NewDelete().Model((*pkg.FolderShare)(nil)).Where("id = ? AND shared_with_user_id = ?", id, userID).Exec(ctx)
	if err != nil {
		return "", err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return "", fmt.Errorf("Share not found")
	}

	var folder pkg.Folder
	if err := db.NewSelect().Model(&folder).Where("id = ?", fs.FolderID).Scan(ctx); err == nil {
		return folder.UserID, nil
	}
	return "", nil
}

func removeImportedShareLink(ctx context.Context, db *bun.DB, id, userID string) error {
	res, err := db.NewDelete().Model((*pkg.ImportedShare)(nil)).
		Where("id = ? AND user_id = ?", id, userID).
		Exec(ctx)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return fmt.Errorf("Share not found")
	}
	return nil
}
