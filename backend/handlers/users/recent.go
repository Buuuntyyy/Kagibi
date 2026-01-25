package users

import (
	"log"
	"net/http"
	"safercloud/backend/pkg"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type AddRecentRequest struct {
	ID   int64  `json:"id" binding:"required"`
	Type string `json:"type" binding:"required,oneof=file folder"`
}

// AddRecentActivityHandler adds a file or folder to the user's recent history
func AddRecentActivityHandler(c *gin.Context, db *bun.DB) {
	userID := c.MustGet("user_id").(string)

	var req AddRecentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	// Upsert approach: Check if exists, update AccessedAt, or Insert
	activity := &pkg.RecentActivity{
		UserID:     userID,
		AccessedAt: time.Now(),
	}

	if req.Type == "file" {
		activity.FileID = &req.ID
	} else {
		activity.FolderID = &req.ID
	}

	// Logic to keep only one entry per resource per user
	// We delete any existing entry for this resource first properly to "move it to top"
	var err error
	if req.Type == "file" {
		_, err = db.NewDelete().Model((*pkg.RecentActivity)(nil)).
			Where("user_id = ? AND file_id = ?", userID, req.ID).
			Exec(c.Request.Context())
	} else {
		_, err = db.NewDelete().Model((*pkg.RecentActivity)(nil)).
			Where("user_id = ? AND folder_id = ?", userID, req.ID).
			Exec(c.Request.Context())
	}

	if err != nil {
		log.Printf("Error clearing previous recent activity: %v", err)
		// Continue anyway
	}

	_, err = db.NewInsert().Model(activity).Exec(c.Request.Context())
	if err != nil {
		log.Printf("Failed to insert recent activity: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record activity"})
		return
	}

	// Cleanup old entries (keep last 20)
	// This can be done async or here. For simplicity, here.
	// Subquery delete is tricky in Bun sometimes, simpler to just run it.
	// DELETE FROM recent_activities WHERE id NOT IN (SELECT id FROM recent_activities WHERE user_id = ? ORDER BY accessed_at DESC LIMIT 20)
	// Implementing this safely in code:
	// Find the 20th item timestamp
	/*
		var recentIds []int64
		err = db.NewSelect().Model((*pkg.RecentActivity)(nil)).
			Column("id").
			Where("user_id = ?", userID).
			Order("accessed_at DESC").
			Limit(20).
			Scan(c.Request.Context(), &recentIds)
	*/

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// GetRecentActivityHandler retrieves the user's recent history
func GetRecentActivityHandler(c *gin.Context, db *bun.DB) {
	userID := c.MustGet("user_id").(string)

	var activities []pkg.RecentActivity
	err := db.NewSelect().Model(&activities).
		Where("?TableAlias.user_id = ?", userID). // Specify table name to avoid ambiguous column error
		Relation("File").
		Relation("Folder").
		Order("accessed_at DESC").
		Limit(10).
		Scan(c.Request.Context())

	if err != nil {
		log.Printf("Failed to fetch recent activity: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch history"})
		return
	}

	// Format response to match what frontend expects or generic structure
	var result []gin.H = []gin.H{} // Initialize as empty array to avoid null in JSON
	for _, act := range activities {
		if act.File != nil {
			result = append(result, gin.H{
				"type":        "file",
				"id":          act.File.ID,
				"name":        act.File.Name,
				"path":        act.File.Path,
				"size":        act.File.Size,
				"mime_type":   act.File.MimeType,
				"updated_at":  act.File.UpdatedAt,
				"accessed_at": act.AccessedAt,
				"file":        act.File, // Include full object if needed
			})
		} else if act.Folder != nil {
			result = append(result, gin.H{
				"type":        "folder",
				"id":          act.Folder.ID,
				"name":        act.Folder.Name,
				"path":        act.Folder.Path,
				"tags":        act.Folder.Tags,
				"updated_at":  act.Folder.UpdatedAt,
				"accessed_at": act.AccessedAt,
				"folder":      act.Folder,
			})
		}
	}

	c.JSON(http.StatusOK, result)
}
