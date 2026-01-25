package shares

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"safercloud/backend/pkg"
	"safercloud/backend/pkg/ws"
	"strings"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// ListImportedSharesHandler lists shares that have been shared with the user (imported)
func ListImportedSharesHandler(c *gin.Context, db *bun.DB) {
	userIDInterface, _ := c.Get("user_id")
	userID := userIDInterface.(string)

	var importedShares []pkg.ImportedShare
	err := db.NewSelect().Model(&importedShares).
		Relation("ShareLink").
		Where("ish.user_id = ?", userID).
		Order("ish.created_at DESC").
		Scan(c.Request.Context())

	if err != nil {
		log.Printf("Error fetching imported shares: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch shared with me items"})
		return
	}

	type SharedWithMeResponse struct {
		ID           int64     `json:"id"` // ImportedShare ID or Share ID
		ShareLinkID  int64     `json:"share_link_id,omitempty"`
		Token        string    `json:"token,omitempty"`
		ResourceType string    `json:"type"`
		Name         string    `json:"name"`
		OwnerName    string    `json:"owner_name"`
		SharedAt     time.Time `json:"shared_at"`
		Size         int64     `json:"size"`
		Link         string    `json:"link,omitempty"`
		FileID       int64     `json:"file_id,omitempty"`   // IMPORTANT: For Direct File Share
		FolderID     int64     `json:"folder_id,omitempty"` // IMPORTANT: For Direct Folder Share
		EncryptedKey string    `json:"encrypted_key"`       // IMPORTANT: For Direct File Share (Removed omitempty to debug)
	}

	var response []SharedWithMeResponse

	// 1. Fetch Imported Shares (Link based)
	for _, is := range importedShares {
		if is.ShareLink == nil {
			continue
		}

		sl := is.ShareLink
		name := "Unknown"
		size := int64(0)
		ownerName := "Unknown"

		// Fetch Owner Name
		var owner pkg.User
		if err := db.NewSelect().Model(&owner).Where("id = ?", sl.OwnerID).Scan(c.Request.Context()); err == nil {
			ownerName = owner.Name
		}

		// Fetch Resource Details
		if sl.ResourceType == "file" {
			var f pkg.File
			if err := db.NewSelect().Model(&f).Where("id = ?", sl.ResourceID).Scan(c.Request.Context()); err == nil {
				name = f.Name
				size = f.Size
			}
		} else if sl.ResourceType == "folder" {
			var f pkg.Folder
			if err := db.NewSelect().Model(&f).Where("id = ?", sl.ResourceID).Scan(c.Request.Context()); err == nil {
				name = f.Name
			}
		}

		response = append(response, SharedWithMeResponse{
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

	// 2. Fetch Direct File Shares
	var fileShares []pkg.FileShare
	err = db.NewSelect().Model(&fileShares).
		Where("shared_with_user_id = ?", userID).
		Scan(c.Request.Context())

	if err == nil {
		for _, fs := range fileShares {
			var file pkg.File
			// Need to fetch file to get owner and details
			if err := db.NewSelect().Model(&file).Where("id = ?", fs.FileID).Scan(c.Request.Context()); err == nil {
				var owner pkg.User
				ownerName := "Unknown"
				if err := db.NewSelect().Model(&owner).Where("id = ?", file.UserID).Scan(c.Request.Context()); err == nil {
					ownerName = owner.Name
				}

				// Debug Log
				// fmt.Printf("Direct Share Found: ID=%d, File=%s, KeyLen=%d\n", fs.ID, file.Name, len(fs.EncryptedKey))

				response = append(response, SharedWithMeResponse{
					ID:           fs.ID, // Use Share ID
					ResourceType: "file",
					Name:         file.Name,
					OwnerName:    ownerName,
					SharedAt:     fs.CreatedAt,
					Size:         file.Size,
					FileID:       file.ID,         // Needed for frontend detection
					EncryptedKey: fs.EncryptedKey, // Needed for decryption
					Link:         "",
				})
			}
		}
	}

	// 3. Fetch Direct Folder Shares
	var folderShares []pkg.FolderShare
	err = db.NewSelect().Model(&folderShares).
		Where("shared_with_user_id = ?", userID).
		Scan(c.Request.Context())

	if err == nil {
		for _, fs := range folderShares {
			fmt.Printf("DEBUG READ: ShareID: %d, FolderID: %d, KeyLen: %d\n", fs.ID, fs.FolderID, len(fs.EncryptedKey))
			var folder pkg.Folder
			if err := db.NewSelect().Model(&folder).Where("id = ?", fs.FolderID).Scan(c.Request.Context()); err == nil {
				var owner pkg.User
				ownerName := "Unknown"
				if err := db.NewSelect().Model(&owner).Where("id = ?", folder.UserID).Scan(c.Request.Context()); err == nil {
					ownerName = owner.Name
				}

				response = append(response, SharedWithMeResponse{
					ID:           fs.ID, // Use Share ID
					ResourceType: "folder",
					Name:         folder.Name,
					OwnerName:    ownerName,
					SharedAt:     fs.CreatedAt,
					Size:         0,               // Folders don't track size directly here
					FolderID:     folder.ID,       // Needed for frontend
					EncryptedKey: fs.EncryptedKey, // Needed for decryption
					Link:         "",
				})
			}
		}
	}

	c.JSON(http.StatusOK, response)
}

// ImportShareHandler adds a share link to the user's "Shared With Me" list
func ImportShareHandler(c *gin.Context, db *bun.DB) {
	userIDInterface, _ := c.Get("user_id")
	userID := userIDInterface.(string)

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
	userIDInterface, _ := c.Get("user_id")
	userID := userIDInterface.(string)
	id := c.Param("id")
	shareType := c.Query("type") // "imported" (default) or "direct_file" or "direct_folder"
	shareType = strings.ReplaceAll(strings.ReplaceAll(shareType, "\n", "_"), "\r", "_")

	log.Printf("DEBUG: RemoveImportedShareHandler - UserID: %s, ID: %s, Type: %s", userID, id, shareType)

	var ownerIDToNotify string

	if shareType == "direct_file" {
		// Fetch share to get FileID first
		var fs pkg.FileShare
		err := db.NewSelect().Model(&fs).Where("id = ? AND shared_with_user_id = ?", id, userID).Scan(c.Request.Context())
		if err == nil {
			// Get Owner of file
			var file pkg.File
			if err := db.NewSelect().Model(&file).Where("id = ?", fs.FileID).Scan(c.Request.Context()); err == nil {
				ownerIDToNotify = file.UserID
			}
		}

		// Delete from FileShare
		res, err := db.NewDelete().Model((*pkg.FileShare)(nil)).
			Where("id = ? AND shared_with_user_id = ?", id, userID).
			Exec(c.Request.Context())

		if err != nil {
			log.Printf("Error deleting file share: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove direct share"})
			return
		}
		rows, _ := res.RowsAffected()
		log.Printf("Rows deleted (file share): %d", rows)

		if rows == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Share not found or mismatch"})
			return
		}
	} else if shareType == "direct_folder" {
		// Fetch share
		var fs pkg.FolderShare
		err := db.NewSelect().Model(&fs).Where("id = ? AND shared_with_user_id = ?", id, userID).Scan(c.Request.Context())
		if err == nil {
			var folder pkg.Folder
			if err := db.NewSelect().Model(&folder).Where("id = ?", fs.FolderID).Scan(c.Request.Context()); err == nil {
				ownerIDToNotify = folder.UserID
			}
		}

		// Delete from FolderShare
		res, err := db.NewDelete().Model((*pkg.FolderShare)(nil)).
			Where("id = ? AND shared_with_user_id = ?", id, userID).
			Exec(c.Request.Context())

		if err != nil {
			log.Printf("Error deleting folder share: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove direct folder share"})
			return
		}
		rows, _ := res.RowsAffected()
		log.Printf("Rows deleted (folder share): %d", rows)

		if rows == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Share not found or mismatch"})
			return
		}
	} else {
		// Default to imported_shares
		// Not really needed for owner notification as it's just a local reference removal,
		// but maybe owner wants to know their link was removed from someone's list?
		// Usually public links don't notify owner on import/removal from list.

		// Delete
		res, err := db.NewDelete().Model((*pkg.ImportedShare)(nil)).
			Where("id = ? AND user_id = ?", id, userID).
			Exec(c.Request.Context())

		if err != nil {
			log.Printf("Error deleting imported share: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove imported share"})
			return
		}
		rows, _ := res.RowsAffected()
		log.Printf("Rows deleted (imported share): %d", rows)

		if rows == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Share not found or mismatch"})
			return
		}
	}

	// Send Notification
	if ownerIDToNotify != "" {
		wsManager.SendToUser(ownerIDToNotify, ws.MsgStorageUpdate, map[string]interface{}{
			"action": "share_revoked_by_recipient",
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Removed from shared with me"})
}
