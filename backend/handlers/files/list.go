// backend/handlers/files/list.go
package files

import (
	"log"
	"net/http"

	"safercloud/backend/pkg"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func ListFilesHandler(c *gin.Context, db *bun.DB) {
	userIDInterface, _ := c.Get("user_id")
	userID := userIDInterface.(string)

	path := c.Param("path")
	if path == "" {
		path = "/"
	}

	log.Printf("ListFilesHandler: userID=%s path=%s", userID, path)

	files, folders, err := pkg.ListItemsByUser(db, userID, path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"files": files, "folders": folders})
}
