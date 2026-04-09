package shares

import (
	"context"
	"database/sql"
	"fmt"
	"kagibi/backend/pkg"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

const (
	logStorageUpdateFailed = "Failed to emit storage_update event"
	errInvalidResourceType = "Invalid resource type"
	queryIDEq              = "id = ?"
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

func CreateDirectShareHandler(c *gin.Context, db *bun.DB) {
	var req CreateDirectShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	exists, err := checkUserExists(c.Request.Context(), db, req.FriendID)
	if err != nil || !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Friend user not found"})
		return
	}

	switch req.ResourceType {
	case "file":
		err = handleFileShare(c.Request.Context(), db, req)
	case "folder":
		err = handleFolderShare(c.Request.Context(), db, req)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": errInvalidResourceType})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to share resource: " + err.Error()})
		return
	}

	sendShareNotifications(c, db, req.FriendID)
	c.JSON(http.StatusCreated, gin.H{"message": "Resource shared successfully"})
}

func checkUserExists(ctx context.Context, db *bun.DB, userID string) (bool, error) {
	return db.NewSelect().Model((*pkg.User)(nil)).Where(queryIDEq, userID).Exists(ctx)
}

func handleFileShare(ctx context.Context, db *bun.DB, req CreateDirectShareRequest) error {
	share := &pkg.FileShare{
		FileID:           req.ResourceID,
		SharedWithUserID: req.FriendID,
		EncryptedKey:     req.EncryptedKey,
		Permission:       req.Permission,
		CreatedAt:        time.Now(),
	}
	_, err := db.NewInsert().Model(share).Exec(ctx)
	return err
}

func handleFolderShare(ctx context.Context, db *bun.DB, req CreateDirectShareRequest) error {
	share := &pkg.FolderShare{
		FolderID:         req.ResourceID,
		SharedWithUserID: req.FriendID,
		EncryptedKey:     req.EncryptedKey,
		Permission:       req.Permission,
		CreatedAt:        time.Now(),
	}

	if _, err := db.NewInsert().Model(share).Exec(ctx); err != nil {
		return err
	}

	if err := upsertFolderFileKeys(ctx, db, req.ResourceID, req.FolderFileKeys); err != nil {
		log.Print("[shares] FolderFileKeys upsert error")
	}

	if err := upsertFolderSubFolderKeys(ctx, db, req.ResourceID, req.FolderFolderKeys); err != nil {
		log.Print("[shares] FolderFolderKeys upsert error")
	}
	return nil
}

func upsertFolderFileKeys(ctx context.Context, db *bun.DB, folderID int64, keysMap map[int64]string) error {
	if len(keysMap) == 0 {
		return nil
	}
	var keys []pkg.FolderFileKey
	for fileID, key := range keysMap {
		keys = append(keys, pkg.FolderFileKey{
			FolderID:     folderID,
			FileID:       fileID,
			EncryptedKey: key,
			CreatedAt:    time.Now(),
		})
	}
	_, err := db.NewInsert().Model(&keys).
		On("CONFLICT (folder_id, file_id) DO UPDATE").
		Set("encrypted_key = EXCLUDED.encrypted_key").
		Exec(ctx)
	return err
}

func upsertFolderSubFolderKeys(ctx context.Context, db *bun.DB, parentID int64, keysMap map[int64]string) error {
	if len(keysMap) == 0 {
		return nil
	}
	var folderKeys []pkg.FolderFolderKey
	for subFolderID, key := range keysMap {
		folderKeys = append(folderKeys, pkg.FolderFolderKey{
			ParentFolderID: parentID,
			SubFolderID:    subFolderID,
			EncryptedKey:   key,
			CreatedAt:      time.Now(),
		})
	}
	_, err := db.NewInsert().Model(&folderKeys).
		On("CONFLICT (parent_folder_id, sub_folder_id) DO UPDATE").
		Set("encrypted_key = EXCLUDED.encrypted_key").
		Exec(ctx)
	return err
}

func sendShareNotifications(c *gin.Context, db *bun.DB, friendID string) {
	// Notify friend via Supabase Realtime
	if err := pkg.EmitRealtimeEvent(c.Request.Context(), db, friendID, "storage_update", map[string]interface{}{
		"action": "share_received",
	}); err != nil {
		log.Print(logStorageUpdateFailed)
	}
	userID := c.GetString("user_id")
	if err := pkg.EmitRealtimeEvent(c.Request.Context(), db, userID, "storage_update", map[string]interface{}{
		"action": "share_created",
	}); err != nil {
		log.Print(logStorageUpdateFailed)
	}
}

// RemoveDirectShareHandler revokes a direct share with a friend
func RemoveDirectShareHandler(c *gin.Context, db *bun.DB) {
	currentUserID := c.GetString("user_id")
	shareIDStr := sanitizeInput(c.Query("id"))
	resourceType := c.Query("resource_type")
	friendID := sanitizeInput(c.Query("friend_id"))
	resourceIDStr := c.Query("resource_id")
	resourceIDStr = sanitizeInput(resourceIDStr)

	if shareIDStr != "" {
		if err := deleteShareByID(c.Request.Context(), db, shareIDStr, resourceType); err != nil {
			handleRemoveError(c, err)
			return
		}
	} else {
		if resourceIDStr == "" || friendID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing parameters"})
			return
		}
		if err := deleteShareByResource(c.Request.Context(), db, resourceIDStr, resourceType, friendID); err != nil {
			handleRemoveError(c, err)
			return
		}
		// Notify friend via Supabase Realtime
		if err := pkg.EmitRealtimeEvent(c.Request.Context(), db, friendID, "storage_update", map[string]interface{}{
			"action": "share_revoked_by_owner",
		}); err != nil {
			log.Print(logStorageUpdateFailed)
		}
	}

	// Notify self via Supabase Realtime
	if err := pkg.EmitRealtimeEvent(c.Request.Context(), db, currentUserID, "storage_update", map[string]interface{}{
		"action": "share_revoked",
	}); err != nil {
		log.Print(logStorageUpdateFailed)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Share revoked"})
}

func sanitizeInput(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\n", "_"), "\r", "_")
}

func deleteShareByID(ctx context.Context, db *bun.DB, shareIDStr, resourceType string) error {
	shareID, _ := strconv.ParseInt(shareIDStr, 10, 64)
	var res sql.Result
	var err error

	switch resourceType {
	case "file":
		res, err = db.NewDelete().Model((*pkg.FileShare)(nil)).Where(queryIDEq, shareID).Exec(ctx)
	case "folder":
		res, err = db.NewDelete().Model((*pkg.FolderShare)(nil)).Where(queryIDEq, shareID).Exec(ctx)
	default:
		return fmt.Errorf("Invalid resource type for ID deletion")
	}

	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return fmt.Errorf("Share not found")
	}
	return nil
}

func deleteShareByResource(ctx context.Context, db *bun.DB, resourceIDStr, resourceType, friendID string) error {
	resID, _ := strconv.ParseInt(resourceIDStr, 10, 64)
	var err error

	switch resourceType {
	case "file":
		_, err = db.NewDelete().Model((*pkg.FileShare)(nil)).
			Where("file_id = ? AND shared_with_user_id = ?", resID, friendID).
			Exec(ctx)
	case "folder":
		_, err = db.NewDelete().Model((*pkg.FolderShare)(nil)).
			Where("folder_id = ? AND shared_with_user_id = ?", resID, friendID).
			Exec(ctx)
	default:
		return fmt.Errorf(errInvalidResourceType)
	}
	return err
}

func handleRemoveError(c *gin.Context, err error) {
	if err.Error() == "Share not found" {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	} else if strings.Contains(err.Error(), errInvalidResourceType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke share: " + err.Error()})
	}
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
