// backend/handlers/files/list.go
package files

import (
	"log"
	"net/http"
	"strconv"

	"safercloud/backend/pkg"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func ListFilesHandler(c *gin.Context, db *bun.DB) {
	userIDStr := c.GetString("user_id")
	userID, _ := strconv.ParseInt(userIDStr, 10, 64)

	path := c.Param("path")
	if path == "" {
		path = "/"
	}
	
    log.Printf("ListFilesHandler: userID=%d path=%s", userID, path)

	files, folders, err := pkg.ListItemsByUser(db, userID, path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"files": files, "folders": folders})
}
