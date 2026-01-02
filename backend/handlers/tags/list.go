package tags

import (
	"log"
	"net/http"
	"safercloud/backend/pkg"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func ListTagsHandler(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		userIDInterface, _ := c.Get("user_id")
		userID, _ := userIDInterface.(string)

		var tags []pkg.Tag

		err := db.NewSelect().Model(&tags).Where("user_id = ?", userID).Scan(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list tags"})
			return
		}
		log.Printf("Tags query took: %v", time.Since(start))

		c.JSON(http.StatusOK, tags)
	}
}
