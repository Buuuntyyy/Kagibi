// backend/handlers/files/download.go
package files

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	
	"safercloud/backend/pkg"
	"safercloud/backend/pkg/s3storage"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func DownloadFileHandler(c *gin.Context, db *bun.DB) {
	userIDInterface, _ := c.Get("user_id")
	userID := userIDInterface.(string)

	fileIDStr := c.Param("fileID")
	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	file, err := pkg.GetFile(db, fileID, userID)
	if err != nil {
		// If not found as owner, check if shared with user (Direct File Share)
		var fileShare pkg.FileShare
		errShare := db.NewSelect().Model(&fileShare).
			Where("file_id = ? AND shared_with_user_id = ?", fileID, userID).
			Scan(c.Request.Context())

		if errShare == nil {
			// It is shared!
			err = db.NewSelect().Model(file).Where("id = ?", fileID).Scan(c.Request.Context())
		} else {
			// Check if file is inside a Shared Folder
			// Retrieve file info first to get the path
			var tempFile pkg.File
			if errFetch := db.NewSelect().Model(&tempFile).Where("id = ?", fileID).Scan(c.Request.Context()); errFetch == nil {
				// Search for any FolderShare that covers this file's path
				type FolderShareWithFolder struct {
					pkg.FolderShare
					Folder *pkg.Folder `bun:"rel:belongs-to,join:folder_id=id"`
				}
		
				var sharesWithFolders []FolderShareWithFolder
				errFolderShare := db.NewSelect().
					Model(&sharesWithFolders).
					Relation("Folder").
					Where("shared_with_user_id = ?", userID).
					Scan(c.Request.Context())
				
				if errFolderShare == nil {
					for _, s := range sharesWithFolders {
						if s.Folder != nil && len(tempFile.Path) > len(s.Folder.Path) && 
						   tempFile.Path[0:len(s.Folder.Path)] == s.Folder.Path &&
						   (tempFile.Path[len(s.Folder.Path)] == '/') {
							// Found a parent shared folder
							err = nil // Clear error
							*file = tempFile // Assign to main file variable
							break
						}
					}
				}
			}
		}
	}

	if err != nil {
		// Log l'erreur pour le débogage côté serveur
		log.Printf("Error getting file from DB. FileID: %d, UserID: %s, Error: %v", fileID, userID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found or permission denied"})
		return
	}

	// Double check: if it was a share, ensure we actually found the file
	if file.ID == 0 {
		log.Printf("File with ID %d not found (even after share check)", fileID)
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// S3 Key construction: Use file.UserID (Owner) instead of userID (Requester)
	s3Key := fmt.Sprintf("users/%s%s", file.UserID, file.Path)

	// Get object from S3
	output, err := s3storage.Client.GetObject(c.Request.Context(), &s3.GetObjectInput{
		Bucket: aws.String(s3storage.BucketName),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		log.Printf("Error getting file from S3. Key: %s, Error: %v", s3Key, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve file from storage"})
		return
	}
	defer output.Body.Close()

	// Définit les en-têtes pour forcer le téléchargement
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")

	disposition := "attachment"
	if c.Query("inline") == "true" {
		disposition = "inline"
	}
	c.Header("Content-Disposition", disposition+"; filename=\""+file.Name+"\"")

	c.Header("Content-Type", "application/octet-stream")
	if file.MimeType != "" {
		c.Header("Content-Type", file.MimeType)
	}
	c.Header("Content-Length", strconv.FormatInt(file.Size, 10))

	// Stream the body to the client
	if _, err := io.Copy(c.Writer, output.Body); err != nil {
		log.Printf("Error streaming file to client: %v", err)
	}
}
