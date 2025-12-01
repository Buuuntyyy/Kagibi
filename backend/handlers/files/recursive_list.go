package files

import (
	"fmt"
	"net/http"
	"safercloud/backend/pkg"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func ListAllFilesRecursiveHandler(c *gin.Context, db *bun.DB) {
	userIDInterface, _ := c.Get("user_id")
	userID := userIDInterface.(string)

	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path is required"})
		return
	}

	files, err := pkg.GetAllFilesRecursive(db, userID, path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log for debugging
	fmt.Printf("RecursiveList: Path=%s UserID=%s Found=%d files\n", path, userID, len(files))
	if len(files) > 0 {
		fmt.Printf("First file: ID=%d Name=%s EncryptedKeyLen=%d\n", files[0].ID, files[0].Name, len(files[0].EncryptedKey))
	}

	c.JSON(http.StatusOK, gin.H{"files": files})
}
