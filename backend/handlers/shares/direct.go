package shares

import (
	"fmt"
	"net/http"
	"safercloud/backend/pkg"
	"safercloud/backend/pkg/ws"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type CreateDirectShareRequest struct {
	ResourceID       int64            `json:"resource_id"`   // File or Folder ID
	ResourceType     string           `json:"resource_type"` // "file" or "folder"
	FriendID         string           `json:"friend_id"`
	EncryptedKey     string           `json:"encrypted_key"` // FileKey (for file) OR FolderKey (for folder), encrypted with friend's PUBLIC key
	Permission       string           `json:"permission"`
	FolderFileKeys   map[int64]string `json:"folder_file_keys"`   // For folders: Map of fileID -> FileKey encrypted with FolderKey
	FolderFolderKeys map[int64]string `json:"folder_folder_keys"` // For folders: Map of subFolderID -> SubFolderKey encrypted with FolderKey
}

func CreateDirectShareHandler(c *gin.Context, db *bun.DB, wsManager *ws.Manager) {
	var req CreateDirectShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	fmt.Printf("DEBUG: CreateDirectShare Payload - Type: %s, ID: %d, Friend: %s, KeyLen: %d\n", req.ResourceType, req.ResourceID, req.FriendID, len(req.EncryptedKey))

	// currentUserID := c.GetString("user_id")

	// Verify friendship exists?
	// Check if the friend user exists to prevent FK violation
	var friendExists bool
	exists, err := db.NewSelect().Model((*pkg.User)(nil)).Where("id = ?", req.FriendID).Exists(c.Request.Context())
	if err != nil {
		fmt.Printf("Error checking user existence: %v\n", err)
	} else {
		friendExists = exists
	}

	if !friendExists {
		fmt.Printf("User %s does not exist in users table!\n", req.FriendID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Friend user not found in database"})
		return
	}

	if req.ResourceType == "file" {
		share := &pkg.FileShare{
			FileID:           req.ResourceID,
			SharedWithUserID: req.FriendID,
			EncryptedKey:     req.EncryptedKey,
			Permission:       req.Permission,
			CreatedAt:        time.Now(),
		}

		fmt.Printf("Inserting FileShare: %+v\n", share)

		_, err := db.NewInsert().Model(share).Exec(c.Request.Context())
		if err != nil {
			// Check for unique constraint (already shared) -> Update key?
			// For now, let's just return error
			fmt.Printf("FileShare Insert Error: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to share file: " + err.Error()})
			return
		}
	} else if req.ResourceType == "folder" {
		share := &pkg.FolderShare{
			FolderID:         req.ResourceID,
			SharedWithUserID: req.FriendID,
			EncryptedKey:     req.EncryptedKey,
			Permission:       req.Permission,
			CreatedAt:        time.Now(),
		}

		fmt.Printf("Inserting FolderShare: %+v, KeyLen: %d\n", share, len(share.EncryptedKey))

		_, err := db.NewInsert().Model(share).Exec(c.Request.Context())
		if err != nil {
			fmt.Printf("FolderShare Insert Error: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to share folder: " + err.Error()})
			return
		}

		// Insert/Upsert FolderFileKeys
		if len(req.FolderFileKeys) > 0 {
			var keys []pkg.FolderFileKey
			for fileID, key := range req.FolderFileKeys {
				keys = append(keys, pkg.FolderFileKey{
					FolderID:     req.ResourceID,
					FileID:       fileID,
					EncryptedKey: key,
					CreatedAt:    time.Now(),
				})
			}

			// We use OnConflict to ignore duplicates (if keys already exist for this folder)
			// Assuming the FolderKey doesn't change frequently.
			// Or we can simple overwrite.
			_, err := db.NewInsert().Model(&keys).
				On("CONFLICT (folder_id, file_id) DO UPDATE").
				Set("encrypted_key = EXCLUDED.encrypted_key").
				Exec(c.Request.Context())

			if err != nil {
				fmt.Printf("FolderFileKeys Insert Error: %v\n", err)
				// Log but don't fail the whole request? The share is created.
				// But without keys, it's useless.
				// Retrying might be needed.
			}
		}

		// Insert/Upsert FolderFolderKeys (Recursively shared subfolder keys)
		if len(req.FolderFolderKeys) > 0 {
			var folderKeys []pkg.FolderFolderKey
			for subFolderID, key := range req.FolderFolderKeys {
				folderKeys = append(folderKeys, pkg.FolderFolderKey{
					ParentFolderID: req.ResourceID,
					SubFolderID:    subFolderID,
					EncryptedKey:   key,
					CreatedAt:      time.Now(),
				})
			}

			_, err := db.NewInsert().Model(&folderKeys).
				On("CONFLICT (parent_folder_id, sub_folder_id) DO UPDATE").
				Set("encrypted_key = EXCLUDED.encrypted_key").
				Exec(c.Request.Context())

			if err != nil {
				fmt.Printf("Error inserting FolderFolderKeys: %v\n", err)
			}
		}

	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resource type"})
		return
	}

	// Notify friend via WebSocket
	wsManager.SendToUser(req.FriendID, ws.MsgStorageUpdate, map[string]interface{}{
		"action": "share_received",
	})

	// Notify self (sender) via WebSocket to refresh file list (show share icon)
	userIDInterface, _ := c.Get("user_id")
	currentUserID := userIDInterface.(string)
	wsManager.SendToUser(currentUserID, ws.MsgStorageUpdate, map[string]interface{}{
		"action": "share_created",
	})

	c.JSON(http.StatusCreated, gin.H{"message": "Resource shared successfully"})
}

// RemoveDirectShareHandler revokes a direct share with a friend
func RemoveDirectShareHandler(c *gin.Context, db *bun.DB, wsManager *ws.Manager) {
	userIDInterface, _ := c.Get("user_id")
	currentUserID := userIDInterface.(string)

	// Query params: resource_id, resource_type, friend_id
	// OR: id (share_id) + resource_type
	resourceIDStr := c.Query("resource_id")
	resourceType := c.Query("resource_type")
	friendID := c.Query("friend_id")
	shareIDStr := c.Query("id")

	fmt.Printf("DEBUG: RemoveDirectShareHandler called. ResID=%s Type=%s FriendID=%s ShareID=%s\n", resourceIDStr, resourceType, friendID, shareIDStr)

	// Mode 1: Delete by Share ID (PK)
	if shareIDStr != "" {
		shareID, _ := strconv.ParseInt(shareIDStr, 10, 64)
		if resourceType == "file" {
			res, err := db.NewDelete().Model((*pkg.FileShare)(nil)).
				Where("id = ?", shareID).
				Exec(c.Request.Context())
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file share by ID"})
				return
			}
			rows, _ := res.RowsAffected()
			if rows == 0 {
				c.JSON(http.StatusNotFound, gin.H{"error": "Share not found"})
				return
			}
		} else if resourceType == "folder" {
			res, err := db.NewDelete().Model((*pkg.FolderShare)(nil)).
				Where("id = ?", shareID).
				Exec(c.Request.Context())
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete folder share by ID"})
				return
			}
			rows, _ := res.RowsAffected()
			if rows == 0 {
				c.JSON(http.StatusNotFound, gin.H{"error": "Share not found"})
				return
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resource type for ID deletion"})
			return
		}

		// Notify self to update UI
		wsManager.SendToUser(currentUserID, ws.MsgStorageUpdate, map[string]interface{}{
			"action": "share_revoked",
		})

		c.JSON(http.StatusOK, gin.H{"message": "Share revoked by ID"})
		return
	}

	// Mode 2: Delete by Resource + Friend (Legacy / ContextMenu)
	if resourceIDStr == "" || friendID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing parameters"})
		return
	}

	resID, _ := strconv.ParseInt(resourceIDStr, 10, 64)

	if resourceType == "file" {
		res, err := db.NewDelete().Model((*pkg.FileShare)(nil)).
			Where("file_id = ? AND shared_with_user_id = ?", resID, friendID).
			Exec(c.Request.Context())
		if err != nil {
			fmt.Printf("Error deleting file share: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke file share"})
			return
		}
		rows, _ := res.RowsAffected()
		fmt.Printf("Deleted %d file share rows\n", rows)
	} else if resourceType == "folder" {
		res, err := db.NewDelete().Model((*pkg.FolderShare)(nil)).
			Where("folder_id = ? AND shared_with_user_id = ?", resID, friendID).
			Exec(c.Request.Context())
		if err != nil {
			fmt.Printf("Error deleting folder share: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke folder share"})
			return
		}
		rows, _ := res.RowsAffected()
		fmt.Printf("Deleted %d folder share rows\n", rows)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resource type"})
		return
	}

	// Notify self to update UI (Mode 2)
	wsManager.SendToUser(currentUserID, ws.MsgStorageUpdate, map[string]interface{}{
		"action": "share_revoked",
	})

	// Also notify friend that access is lost? (Optional but good UX)
	wsManager.SendToUser(friendID, ws.MsgStorageUpdate, map[string]interface{}{
		"action": "share_revoked_by_owner",
	})

	c.JSON(http.StatusOK, gin.H{"message": "Share revoked"})
}

// ListDirectSharesForResourceHandler returns a list of friend IDs with whom the resource is shared
func ListDirectSharesForResourceHandler(c *gin.Context, db *bun.DB) {
	resourceIDStr := c.Query("resource_id")
	resourceType := c.Query("resource_type")

	if resourceIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing resource_id"})
		return
	}

	var sharedWithIDs []string

	if resourceType == "file" || resourceType == "" {
		var shares []pkg.FileShare
		err := db.NewSelect().Model(&shares).
			Where("file_id = ?", resourceIDStr).
			Scan(c.Request.Context())
		if err == nil {
			for _, s := range shares {
				sharedWithIDs = append(sharedWithIDs, s.SharedWithUserID)
			}
		}
	} else if resourceType == "folder" {
		var shares []pkg.FolderShare
		err := db.NewSelect().Model(&shares).
			Where("folder_id = ?", resourceIDStr).
			Scan(c.Request.Context())
		if err == nil {
			for _, s := range shares {
				sharedWithIDs = append(sharedWithIDs, s.SharedWithUserID)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"shared_with": sharedWithIDs})
}
