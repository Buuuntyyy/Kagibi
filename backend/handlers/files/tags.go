package files

import (
	"net/http"

	"safercloud/backend/pkg"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type UpdateTagsRequest struct {
	ID   int64    `json:"id" binding:"required"`
	Type string   `json:"type" binding:"required,oneof=file folder"`
	Tags []string `json:"tags" binding:"required"`
}

func UpdateTagsHandler(c *gin.Context, db *bun.DB) {
	var req UpdateTagsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	userIDInterface, _ := c.Get("user_id")
	userID, _ := userIDInterface.(string)

	ctx := c.Request.Context()

	if req.Type == "file" {
		// Verify ownership
		exists, err := db.NewSelect().Model((*pkg.File)(nil)).Where("id = ? AND user_id = ?", req.ID, userID).Exists(ctx)
		if err != nil || !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}

		file := &pkg.File{Tags: req.Tags}
		_, err = db.NewUpdate().Model(file).Column("tags").Where("id = ?", req.ID).Exec(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tags: " + err.Error()})
			return
		}
	} else {
		// Verify ownership
		exists, err := db.NewSelect().Model((*pkg.Folder)(nil)).Where("id = ? AND user_id = ?", req.ID, userID).Exists(ctx)
		if err != nil || !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
			return
		}

		folder := &pkg.Folder{Tags: req.Tags}
		_, err = db.NewUpdate().Model(folder).Column("tags").Where("id = ?", req.ID).Exec(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tags: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tags updated successfully", "tags": req.Tags})
}
