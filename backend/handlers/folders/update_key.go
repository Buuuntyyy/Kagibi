package folders

import (
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"log"
	"net/http"
	"kagibi/backend/pkg"
	"strings"
)

type UpdateFolderKeyRequest struct {
	EncryptedKey string `json:"encrypted_key" binding:"required"`
}

func UpdateFolderKeyHandler(c *gin.Context, db *bun.DB) {
	userIDInterface, _ := c.Get("user_id")
	userID := userIDInterface.(string)
	folderID := c.Param("id")
	folderID = strings.ReplaceAll(strings.ReplaceAll(folderID, "\n", "_"), "\r", "_")

	var req UpdateFolderKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	log.Printf("UpdateFolderKeyHandler: userID=%s folderID=%s", userID, folderID)

	// Check if folder exists and belongs to user
	var folder pkg.Folder
	err := db.NewSelect().Model(&folder).
		Where("id = ?", folderID).
		Where("user_id = ?", userID).
		Scan(c.Request.Context())

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found or access denied"})
		return
	}

	// Update the key
	folder.EncryptedKey = req.EncryptedKey
	_, err = db.NewUpdate().Model(&folder).
		Column("encrypted_key").
		Where("id = ?", folderID).
		Exec(c.Request.Context())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update folder key"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Folder key updated"})
}
