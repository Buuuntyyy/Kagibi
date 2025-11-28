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
		// Log l'erreur pour le débogage côté serveur
		log.Printf("Error getting file from DB. FileID: %d, UserID: %s, Error: %v", fileID, userID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found or permission denied"})
		return
	}

	// S3 Key construction
	s3Key := fmt.Sprintf("users/%s%s", userID, file.Path)

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
	c.Header("Content-Disposition", "attachment; filename=\""+file.Name+"\"")
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
