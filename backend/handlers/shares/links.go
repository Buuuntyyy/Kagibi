package shares

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"safercloud/backend/pkg"
	"safercloud/backend/pkg/s3storage"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func generateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// CreateShareLinkHandler creates a public link for a file or folder
func CreateShareLinkHandler(c *gin.Context, db *bun.DB) {
	userIDInterface, _ := c.Get("user_id")
	userID := userIDInterface.(string)

	var req struct {
		ResourceID   int64            `json:"resource_id"`
		ResourceType string           `json:"resource_type"` // "file" or "folder"
		ExpiresAt    *time.Time       `json:"expires_at"`
		Password     string           `json:"password"` // Optional, not implemented yet in this iteration
		EncryptedKey string           `json:"encrypted_key"`
		Token        string           `json:"token"`
		FileKeys     map[int64]string `json:"file_keys"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.ExpiresAt != nil && req.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Expiration date must be in the future"})
		return
	}

	log.Printf("CreateShareLink: ResourceID=%d Type=%s FileKeys=%d", req.ResourceID, req.ResourceType, len(req.FileKeys))

	// Verify ownership and get path
	var resourcePath string
	if req.ResourceType == "file" {
		var file pkg.File
		err := db.NewSelect().Model(&file).
			Where("id = ? AND user_id = ?", req.ResourceID, userID).
			Scan(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found or permission denied"})
			return
		}
		resourcePath = file.Path
	} else if req.ResourceType == "folder" {
		var folder pkg.Folder
		err := db.NewSelect().Model(&folder).
			Where("id = ? AND user_id = ?", req.ResourceID, userID).
			Scan(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found or permission denied"})
			return
		}
		resourcePath = folder.Path
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resource type"})
		return
	}

	// Check if a share link already exists for this resource
	var existingShare pkg.ShareLink
	err := db.NewSelect().Model(&existingShare).
		Where("resource_id = ? AND resource_type = ? AND owner_id = ?", req.ResourceID, req.ResourceType, userID).
		Scan(c.Request.Context())

	if err == nil {
		// A share link already exists, return it instead of creating a new one.
		c.JSON(http.StatusConflict, gin.H{
			"error": "A share link for this resource already exists",
			"token": existingShare.Token,
			"link":  fmt.Sprintf("/s/%s", existingShare.Token),
		})
		return
	}

	token := req.Token
	if token == "" {
		var err error
		token, err = generateToken()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}
	}

	shareLink := &pkg.ShareLink{
		ResourceID:   req.ResourceID,
		ResourceType: req.ResourceType,
		Path:         resourcePath,
		OwnerID:      userID,
		Token:        token,
		ExpiresAt:    req.ExpiresAt,
		EncryptedKey: req.EncryptedKey,
		// PasswordHash: ... (TODO if password provided)
	}

	_, err = db.NewInsert().Model(shareLink).Exec(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create share link"})
		return
	}

	// Insert file keys if any
	if len(req.FileKeys) > 0 {
		var shareFileKeys []pkg.ShareFileKey
		for fileID, key := range req.FileKeys {
			shareFileKeys = append(shareFileKeys, pkg.ShareFileKey{
				ShareID:      shareLink.ID,
				FileID:       fileID,
				EncryptedKey: key,
			})
		}
		if len(shareFileKeys) > 0 {
			if _, err := db.NewInsert().Model(&shareFileKeys).Exec(c.Request.Context()); err != nil {
				log.Printf("Error saving share keys: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file keys"})
				return
			}
			log.Printf("Saved %d share keys", len(shareFileKeys))
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Link created",
		"token":   token,
		"link":    fmt.Sprintf("/s/%s", token), // Frontend URL format
	})
}

// GetShareLinkHandler retrieves info about a shared resource
func GetShareLinkHandler(c *gin.Context, db *bun.DB) {
	token := c.Param("token")

	var shareLink pkg.ShareLink
	err := db.NewSelect().Model(&shareLink).
		Where("token = ?", token).
		Scan(c.Request.Context())

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
		return
	}

	// Check expiration
	if shareLink.ExpiresAt != nil && shareLink.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusGone, gin.H{"error": "Link expired"})
		return
	}

	// Increment views
	_, _ = db.NewUpdate().Model(&shareLink).
		Set("views = views + 1").
		Where("id = ?", shareLink.ID).
		Exec(c.Request.Context())

	// Fetch owner info
	var owner pkg.User
	err = db.NewSelect().Model(&owner).
		Where("id = ?", shareLink.OwnerID).
		Scan(c.Request.Context())

	ownerEmail := "Unknown"
	if err == nil {
		ownerEmail = owner.Email
	}

	if shareLink.ResourceType == "file" {
		var file pkg.File
		err := db.NewSelect().Model(&file).
			Where("id = ?", shareLink.ResourceID).
			Scan(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"resource_type": "file",
			"resource_name": file.Name,
			"file_size":     file.Size,
			"mime_type":     file.MimeType,
			"owner_email":   ownerEmail,
			"expires_at":    shareLink.ExpiresAt,
			"encrypted_key": shareLink.EncryptedKey,
		})
	} else {
		// Folder logic
		var folder pkg.Folder
		err := db.NewSelect().Model(&folder).
			Where("id = ?", shareLink.ResourceID).
			Scan(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"resource_type": "folder",
			"resource_name": folder.Name,
			"owner_email":   ownerEmail,
			"expires_at":    shareLink.ExpiresAt,
		})
	}
}

// DownloadSharedFileHandler downloads a file via a share link
func DownloadSharedFileHandler(c *gin.Context, db *bun.DB) {
	token := c.Param("token")

	var shareLink pkg.ShareLink
	err := db.NewSelect().Model(&shareLink).
		Where("token = ?", token).
		Scan(c.Request.Context())

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
		return
	}

	if shareLink.ResourceType != "file" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not a file"})
		return
	}

	if shareLink.ExpiresAt != nil && shareLink.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusGone, gin.H{"error": "Link expired"})
		return
	}

	var file pkg.File
	err = db.NewSelect().Model(&file).
		Where("id = ?", shareLink.ResourceID).
		Scan(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// S3 Key construction
	s3Key := fmt.Sprintf("users/%s%s", shareLink.OwnerID, file.Path)

	// Get object from S3
	output, err := s3storage.Client.GetObject(c.Request.Context(), &s3.GetObjectInput{
		Bucket: aws.String(s3storage.BucketName),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		log.Printf("Error getting shared file from S3. Key: %s, Error: %v", s3Key, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve file from storage"})
		return
	}
	defer output.Body.Close()

	// Headers
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename=\""+file.Name+"\"")
	c.Header("Content-Type", "application/octet-stream")
	if file.MimeType != "" {
		c.Header("Content-Type", file.MimeType)
	}
	c.Header("Content-Length", strconv.FormatInt(file.Size, 10))

	// Stream
	if _, err := io.Copy(c.Writer, output.Body); err != nil {
		log.Printf("Error streaming shared file to client: %v", err)
	}
}

// GetShareForResourceHandler allows the owner to retrieve share link(s) for a given file
func GetShareForResourceHandler(c *gin.Context, db *bun.DB) {
	userIDInterface, _ := c.Get("user_id")
	userID := userIDInterface.(string)

	fileIDParam := c.Param("fileID")
	fileID, err := strconv.ParseInt(fileIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	var links []pkg.ShareLink
	err = db.NewSelect().Model(&links).
		Where("resource_type = ?", "file").
		Where("resource_id = ?", fileID).
		Where("owner_id = ?", userID).
		Scan(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(links) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No share links found for this file"})
		return
	}

	resp := make([]gin.H, 0, len(links))
	for _, l := range links {
		resp = append(resp, gin.H{
			"id":         l.ID,
			"token":      l.Token,
			"link":       fmt.Sprintf("/s/%s", l.Token),
			"expires_at": l.ExpiresAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"links": resp})
}
