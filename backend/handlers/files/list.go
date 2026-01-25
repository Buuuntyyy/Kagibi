// backend/handlers/files/list.go
package files

import (
	"log"
	"net/http"

	"safercloud/backend/pkg"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"strings"
)

func ListFilesHandler(c *gin.Context, db *bun.DB) {
	userIDInterface, _ := c.Get("user_id")
	userID := userIDInterface.(string)

	path := c.Param("path")
	path = strings.ReplaceAll(strings.ReplaceAll(path, "\n", "_"), "\r", "_")
	if path == "" {
		path = "/"
	}

	log.Printf("ListFilesHandler: userID=%s path=%s", userID, path)

	files, folders, err := pkg.ListItemsByUser(db, userID, path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Fetch current folder meta (for ID)
	var currentFolderID int64 = 0
	if path != "/" && path != "" {
		currentFolder := new(pkg.Folder)
		err := db.NewSelect().Model(currentFolder).
			Column("id").
			Where("user_id = ? AND path = ?", userID, path).
			Limit(1).
			Scan(c.Request.Context())
		if err == nil {
			currentFolderID = currentFolder.ID
		}
	}

	c.JSON(http.StatusOK, gin.H{"files": files, "folders": folders, "current_folder_id": currentFolderID})
}
