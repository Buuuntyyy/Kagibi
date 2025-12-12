package shares

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"safercloud/backend/pkg"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// ListSharesHandler lists all active share links created by the user
func ListSharesHandler(c *gin.Context, db *bun.DB) {
	userIDInterface, _ := c.Get("user_id")
	userID := userIDInterface.(string)

	var links []pkg.ShareLink
	err := db.NewSelect().Model(&links).
		Where("owner_id = ?", userID).
		Order("created_at DESC").
		Scan(c.Request.Context())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch shares"})
		return
	}

	type ShareResponse struct {
		ID           int64      `json:"id"`
		Token        string     `json:"token"`
		Link         string     `json:"link"`
		ResourceType string     `json:"resource_type"`
		ResourceName string     `json:"resource_name"`
		Views        int64      `json:"views"`
		ExpiresAt    *time.Time `json:"expires_at"`
		CreatedAt    time.Time  `json:"created_at"`
	}

	var response []ShareResponse

	for _, l := range links {
		name := "Unknown"
		exists := false

		if l.ResourceType == "file" {
			var f pkg.File
			if err := db.NewSelect().Model(&f).Where("id = ?", l.ResourceID).Scan(c.Request.Context()); err == nil {
				name = f.Name
				exists = true
			}
		} else if l.ResourceType == "folder" {
			var f pkg.Folder
			if err := db.NewSelect().Model(&f).Where("id = ?", l.ResourceID).Scan(c.Request.Context()); err == nil {
				name = f.Name
				exists = true
			}
		}

		if !exists {
			// Lazy cleanup: Delete the orphaned share link
			go func(id int64) {
				db.NewDelete().Model((*pkg.ShareLink)(nil)).Where("id = ?", id).Exec(context.Background())
			}(l.ID)
			continue
		}

		response = append(response, ShareResponse{
			ID:           l.ID,
			Token:        l.Token,
			Link:         fmt.Sprintf("/s/%s", l.Token),
			ResourceType: l.ResourceType,
			ResourceName: name,
			Views:        l.Views,
			ExpiresAt:    l.ExpiresAt,
			CreatedAt:    l.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"shares": response})
}
