package folders

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"safercloud/backend/pkg"
	"github.com/uptrace/bun"
)

type CreateFolderRequest struct {
	Name string `json:"name" binding:"required"`
	Path string `json:"path" binding:"required"`
}

func CreateHandler(c *gin.Context, db *bun.DB) {
	userID := c.GetInt64("userID")

	var req CreateFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	folder := &pkg.Folder{
		Name:   req.Name,
		Path:   req.Path,
		UserID: userID,
	}

	if err := pkg.CreateFolderDB(db, folder); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create folder"})
		return
	}

	c.JSON(http.StatusCreated, folder)
}
