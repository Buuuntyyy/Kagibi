// backend/handlers/files/list.go
package files

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"safercloud/backend/pkg"
	"github.com/uptrace/bun"
)

func ListFilesHandler(c *gin.Context, db *bun.DB) {
	userID := c.GetInt64("userID")
	path := c.Param("path")
	if path == "" {
		path = "/"
	}

	files, folders, err := pkg.ListItemsByUser(db, userID, path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"files": files, "folders": folders})
}
