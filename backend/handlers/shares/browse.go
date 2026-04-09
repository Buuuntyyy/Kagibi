package shares

import (
	"kagibi/backend/pkg"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// BrowseSharedFolderHandler handles browsing within a shared folder
func BrowseSharedFolderHandler(c *gin.Context, db *bun.DB) {
	token := c.Param("token")
	subpath := c.Param("subpath")
	subpath = strings.ReplaceAll(strings.ReplaceAll(subpath, "\n", "_"), "\r", "_")

	var shareLink pkg.ShareLink
	err := db.NewSelect().Model(&shareLink).Where("token = ?", token).Scan(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Share link not found"})
		return
	}

	if shareLink.ResourceType != "folder" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This share link is not for a folder"})
		return
	}

	// Clean and validate the subpath to prevent directory traversal
	requestedPath := path.Join(shareLink.Path, subpath)
	requestedPath = strings.ReplaceAll(strings.ReplaceAll(requestedPath, "\n", "_"), "\r", "_")
	if !strings.HasPrefix(requestedPath, shareLink.Path) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access to this path is forbidden"})
		return
	}

	// List files and folders in the requested path
	files, folders, err := pkg.GetSharedFolderContent(db, requestedPath, shareLink.OwnerID, shareLink.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list items"})
		return
	}

	// Fetch owner info
	var owner pkg.User
	err = db.NewSelect().Model(&owner).Where("id = ?", shareLink.OwnerID).Scan(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve owner info"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"folders":       folders,
		"files":         files,
		"owner_email":   owner.Email,
		"resource_name": path.Base(shareLink.Path),
	})
}
