package shares

import (
	"net/http"
	"kagibi/backend/pkg"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// GetActiveSharesForPathHandler returns all active share links that cover the given path
func GetActiveSharesForPathHandler(c *gin.Context, db *bun.DB) {
	userIDInterface, _ := c.Get("user_id")
	userID := userIDInterface.(string)

	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path is required"})
		return
	}

	// Find shares where the share path is a prefix of the current path
	// e.g. Share Path: /docs, Current Path: /docs/work/project
	// We need to find shares where 'path' LIKE share.path + '%'
	// But SQL LIKE is usually 'column LIKE pattern'. Here we want 'pattern LIKE column + %' ?
	// No, we want: current_path LIKE share_path || '%'

	var shares []pkg.ShareLink
	err := db.NewSelect().Model(&shares).
		Where("owner_id = ?", userID).
		Where("resource_type = ?", "folder").
		Where("? LIKE path || '%'", path).
		Scan(c.Request.Context())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"shares": shares})
}
