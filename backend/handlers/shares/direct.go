package shares

import (
	"fmt"
	"net/http"
	"safercloud/backend/pkg"
	"safercloud/backend/pkg/ws"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type CreateDirectShareRequest struct {
	ResourceID   int64  `json:"resource_id"`   // File or Folder ID
	ResourceType string `json:"resource_type"` // "file" or "folder"
	FriendID     string `json:"friend_id"`
	EncryptedKey string `json:"encrypted_key"` // Key encrypted with friend's PUBLIC key
	Permission   string `json:"permission"`
}

func CreateDirectShareHandler(c *gin.Context, db *bun.DB, wsManager *ws.Manager) {
	var req CreateDirectShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

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
			Permission:       req.Permission,
			CreatedAt:        time.Now(),
		}

		fmt.Printf("Inserting FolderShare: %+v\n", share)

		_, err := db.NewInsert().Model(share).Exec(c.Request.Context())
		if err != nil {
			fmt.Printf("FolderShare Insert Error: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to share folder: " + err.Error()})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resource type"})
		return
	}

	// Notify friend via WebSocket
	wsManager.SendToUser(req.FriendID, ws.MsgStorageUpdate, map[string]interface{}{
		"action": "share_received",
	})

	c.JSON(http.StatusCreated, gin.H{"message": "Resource shared successfully"})
}
