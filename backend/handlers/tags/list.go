package tags

import (
	"net/http"
	"safercloud/backend/pkg"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func ListTagsHandler(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("userID")
		var tags []pkg.Tag

		err := db.NewSelect().Model(&tags).Where("user_id = ?", userID).Scan(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list tags"})
			return
		}

		c.JSON(http.StatusOK, tags)
	}
}
