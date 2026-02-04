package users

import (
	"net/http"
	"safercloud/backend/pkg"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type UpdateAvatarRequest struct {
	AvatarURL string `json:"avatar_url" binding:"required"`
}

func UpdateAvatarHandler(c *gin.Context, db *bun.DB) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req UpdateAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Validate avatar URL format (basic check)
	if req.AvatarURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Avatar URL cannot be empty"})
		return
	}

	// Update avatar in database
	_, err := db.NewUpdate().
		Model((*pkg.User)(nil)).
		Set("avatar_url = ?", req.AvatarURL).
		Where("id = ?", userID).
		Exec(c.Request.Context())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update avatar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Avatar updated successfully",
		"avatar_url": req.AvatarURL,
	})
}
