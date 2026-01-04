package shares

import (
	"log"
	"net/http"
	"time"

	"safercloud/backend/pkg"

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
		ID           int64     `json:"id"` // ImportedShare ID
		ShareLinkID  int64     `json:"share_link_id"`
		Token        string    `json:"token"`
		ResourceType string    `json:"type"`
		Name         string    `json:"name"`
		OwnerName    string    `json:"owner_name"`
		SharedAt     time.Time `json:"shared_at"`
		Size         int64     `json:"size"`
		Link         string    `json:"link"`
	}

	var response []SharedWithMeResponse

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
func RemoveImportedShareHandler(c *gin.Context, db *bun.DB) {
	userIDInterface, _ := c.Get("user_id")
	userID := userIDInterface.(string)
	id := c.Param("id")

	_, err := db.NewDelete().Model((*pkg.ImportedShare)(nil)).
		Where("id = ? AND user_id = ?", id, userID).
		Exec(c.Request.Context())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove imported share"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Removed from shared with me"})
}
