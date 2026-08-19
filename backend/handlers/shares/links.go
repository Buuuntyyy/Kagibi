// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package shares

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"kagibi/backend/middleware"
	"kagibi/backend/pkg"
	"kagibi/backend/pkg/mailer"
	"kagibi/backend/pkg/monitoring"
	"kagibi/backend/pkg/s3storage"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"golang.org/x/crypto/bcrypt"
)

const errLinkExpired = "Link expired"
const errLinkAlreadyUsed = "Link already used"

// checkSharePassword validates the X-Share-Password header against the stored bcrypt hash.
// Returns true if the link has no password or if the provided password matches.
// Writes a 401 response and returns false when the password is missing or wrong.
func checkSharePassword(c *gin.Context, shareLink *pkg.ShareLink) bool {
	if shareLink.PasswordHash == "" {
		return true
	}
	password := c.GetHeader("X-Share-Password")
	if password == "" || bcrypt.CompareHashAndPassword([]byte(shareLink.PasswordHash), []byte(password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "password_required"})
		return false
	}
	return true
}

type createShareLinkRequest struct {
	ResourceID     int64            `json:"resource_id"`
	ResourceType   string           `json:"resource_type"` // "file" or "folder"
	ExpiresAt      *time.Time       `json:"expires_at"`
	Password       string           `json:"password"` // Optional
	EncryptedKey   string           `json:"encrypted_key"`
	Token          string           `json:"token"`
	FileKeys       map[int64]string `json:"file_keys"`
	SingleUse      bool             `json:"single_use"`
	PermDownload   bool             `json:"perm_download"`
	PermCreate     bool             `json:"perm_create"`
	PermDelete     bool             `json:"perm_delete"`
	PermMove       bool             `json:"perm_move"`
	UploadOnly     bool             `json:"upload_only"`     // file request : dépôt seul, pas de lecture
	RequestLabel   string           `json:"request_label"`   // message affiché au déposant
	RecipientEmail string           `json:"recipient_email"` // optionnel : destinataire à notifier par mail
	SendEmail      bool             `json:"send_email"`      // envoyer un mail de notification avec le lien
	EmailLang      string           `json:"email_lang"`      // "fr" ou "en"
}

// sharePathFormat returns the public URL path format for a share link: file
// requests (upload_only) are dropped off at /r/, regular shares at /s/.
func sharePathFormat(uploadOnly bool) string {
	if uploadOnly {
		return "/r/%s"
	}
	return "/s/%s"
}

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
	userID := c.GetString("user_id")

	var req createShareLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.ExpiresAt != nil && req.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Expiration date must be in the future"})
		return
	}

	if req.UploadOnly && req.ResourceType != "folder" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File requests are only supported on folders"})
		return
	}

	resourcePath, err := verifyOwnerAndGetPath(c.Request.Context(), db, userID, req.ResourceID, req.ResourceType)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	existingShare, err := checkExistingShareLink(c.Request.Context(), db, userID, req.ResourceID, req.ResourceType, req.UploadOnly)
	if err == nil {
		sendShareLinkEmail(db, userID, req, existingShare.Token, existingShare.RequestLabel)
		c.JSON(http.StatusConflict, gin.H{
			"error": "A share link for this resource already exists",
			"token": existingShare.Token,
			"id":    existingShare.ID,
			"link":  fmt.Sprintf(sharePathFormat(existingShare.UploadOnly), existingShare.Token),
		})
		return
	}

	shareLink, err := createNewShareLink(c.Request.Context(), db, userID, resourcePath, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create share link"})
		return
	}

	if len(req.FileKeys) > 0 {
		if err := saveShareFileKeys(c.Request.Context(), db, shareLink.ID, req.FileKeys); err != nil {
			log.Printf("Error saving share keys: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file keys"})
			return
		}
	}

	monitoring.RecordShareCreated()
	middleware.LogShareCreated(c.Request.Context(), userID, req.ResourceType, req.ResourceID, shareLink.Token, c.ClientIP(), c.Request.UserAgent())
	sendShareLinkEmail(db, userID, req, shareLink.Token, shareLink.RequestLabel)
	c.JSON(http.StatusCreated, gin.H{
		"message": "Link created",
		"token":   shareLink.Token,
		"id":      shareLink.ID,
		"link":    fmt.Sprintf(sharePathFormat(shareLink.UploadOnly), shareLink.Token),
	})
}

// sendShareLinkEmail fires (asynchronously) a notification email pointing the
// recipient at a file-request drop link, mirroring the P2P invite email flow.
// Only used for upload_only (file request) links — regular shares are already
// meant to be sent by the owner through their own channel.
func sendShareLinkEmail(db *bun.DB, ownerID string, req createShareLinkRequest, token, requestLabel string) {
	recipientEmail := strings.TrimSpace(req.RecipientEmail)
	if !req.UploadOnly || !req.SendEmail || recipientEmail == "" {
		return
	}

	lang := req.EmailLang
	if lang != "en" {
		lang = "fr"
	}

	go func() {
		bgCtx := context.Background()
		var owner pkg.User
		if err := db.NewSelect().Model(&owner).Where(queryIDEq, ownerID).Scan(bgCtx); err != nil {
			log.Printf("[FileRequest] fetch owner name: %v", err)
			return
		}

		appURL := os.Getenv("APP_URL")
		if appURL == "" {
			appURL = "https://kagibi.cloud"
		}
		link := appURL + fmt.Sprintf(sharePathFormat(true), token)

		if err := mailer.SendFileRequestInvite(recipientEmail, owner.Name, requestLabel, link, lang); err != nil {
			log.Printf("[FileRequest] invite email failed to %s: %v", recipientEmail, err)
		} else {
			log.Printf("[FileRequest] invite email sent to %s", recipientEmail)
		}
	}()
}

// GetShareLinkHandler retrieves info about a shared resource
func GetShareLinkHandler(c *gin.Context, db *bun.DB) {
	token := c.Param("token")

	shareLink, err := getValidShareLink(c.Request.Context(), db, token)
	if err != nil {
		switch err.Error() {
		case errLinkExpired:
			monitoring.RecordShareAccess("expired")
			c.JSON(http.StatusGone, gin.H{"error": err.Error()})
		case errLinkAlreadyUsed:
			monitoring.RecordShareAccess("already_used")
			c.JSON(http.StatusGone, gin.H{"error": err.Error()})
		default:
			monitoring.RecordShareAccess("not_found")
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		}
		return
	}

	if !checkSharePassword(c, shareLink) {
		return
	}

	monitoring.RecordShareAccess("success")
	go incrementShareViews(context.Background(), db, shareLink.ID)

	ownerEmail := getShareOwnerEmail(c.Request.Context(), db, shareLink.OwnerID)
	response, err := buildShareLinkResponse(c.Request.Context(), db, shareLink, ownerEmail)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// DownloadSharedFileHandler downloads a file via a share link
func DownloadSharedFileHandler(c *gin.Context, db *bun.DB) {
	token := c.Param("token")

	shareLink, err := getValidShareLink(c.Request.Context(), db, token)
	if err != nil {
		status := http.StatusNotFound
		if err.Error() == errLinkExpired || err.Error() == errLinkAlreadyUsed {
			status = http.StatusGone
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	if !checkSharePassword(c, shareLink) {
		return
	}

	if shareLink.ResourceType != "file" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not a file"})
		return
	}

	if !shareLink.PermDownload {
		c.JSON(http.StatusForbidden, gin.H{"error": "Download not permitted on this share"})
		return
	}

	// For single-use links, atomically mark as used before streaming.
	// If another request already consumed it, abort with 410 Gone.
	if shareLink.SingleUse {
		marked, err := markShareLinkUsed(c.Request.Context(), db, shareLink.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process link"})
			return
		}
		if !marked {
			c.JSON(http.StatusGone, gin.H{"error": errLinkAlreadyUsed})
			return
		}
	}

	file, err := getSharedFile(c.Request.Context(), db, shareLink.ResourceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	monitoring.FileDownloadsTotal.Inc()
	go func() {
		_, _ = db.NewUpdate().Model((*pkg.ShareLink)(nil)).
			Set("download_count = download_count + 1").
			Where("id = ?", shareLink.ID).
			Exec(context.Background())
	}()

	if err := streamFileFromS3(c, shareLink.OwnerID, file); err != nil {
		log.Printf("Error streaming file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve file from storage"})
	}
}

// UpdateSharePermissionsHandler updates the permission flags on an existing share link
func UpdateSharePermissionsHandler(c *gin.Context, db *bun.DB) {
	userID := c.GetString("user_id")
	shareID, err := strconv.ParseInt(c.Param("shareID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid share ID"})
		return
	}

	var body struct {
		PermDownload *bool `json:"perm_download"`
		PermCreate   *bool `json:"perm_create"`
		PermDelete   *bool `json:"perm_delete"`
		PermMove     *bool `json:"perm_move"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var share pkg.ShareLink
	if err := db.NewSelect().Model(&share).Where("id = ? AND owner_id = ?", shareID, userID).Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Share not found"})
		return
	}

	// Les permissions d'un lien de demande de fichiers sont figées (dépôt seul)
	if share.UploadOnly {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permissions of a file request link cannot be changed"})
		return
	}

	if body.PermDownload != nil {
		share.PermDownload = *body.PermDownload
	}
	if body.PermCreate != nil {
		share.PermCreate = *body.PermCreate
	}
	if body.PermDelete != nil {
		share.PermDelete = *body.PermDelete
	}
	if body.PermMove != nil {
		share.PermMove = *body.PermMove
	}

	_, err = db.NewUpdate().Model(&share).
		Column("perm_download", "perm_create", "perm_delete", "perm_move").
		Where("id = ?", shareID).
		Exec(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"perm_download": share.PermDownload,
		"perm_create":   share.PermCreate,
		"perm_delete":   share.PermDelete,
		"perm_move":     share.PermMove,
	})
}

// GetShareForResourceHandler allows the owner to retrieve share link(s) for a given file
func GetShareForResourceHandler(c *gin.Context, db *bun.DB) {
	userID := c.GetString("user_id")
	fileID, err := strconv.ParseInt(c.Param("fileID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	links, err := getUserFileShareLinks(c.Request.Context(), db, userID, fileID)
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

// GetFileRequestLinkHandler returns the existing file-request (upload_only) link for a
// folder, if one exists — used by the frontend to show the existing link and its
// settings immediately when reopening the "request files" dialog, instead of the user
// having to attempt creation first and being told it already exists via a 409.
func GetFileRequestLinkHandler(c *gin.Context, db *bun.DB) {
	userID := c.GetString("user_id")
	folderID, err := strconv.ParseInt(c.Query("folder_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid folder ID"})
		return
	}

	existing, err := checkExistingShareLink(c.Request.Context(), db, userID, folderID, "folder", true)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No file request link found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":            existing.ID,
		"token":         existing.Token,
		"link":          fmt.Sprintf(sharePathFormat(true), existing.Token),
		"request_label": existing.RequestLabel,
		"expires_at":    existing.ExpiresAt,
		"has_password":  existing.PasswordHash != "",
		"created_at":    existing.CreatedAt,
	})
}

// --- Helpers ---

func verifyOwnerAndGetPath(ctx context.Context, db *bun.DB, userID string, resID int64, resType string) (string, error) {
	if resType == "file" {
		var file pkg.File
		err := db.NewSelect().Model(&file).Where("id = ? AND user_id = ?", resID, userID).Scan(ctx)
		if err != nil {
			return "", fmt.Errorf("File not found or permission denied")
		}
		return file.Path, nil
	} else if resType == "folder" {
		var folder pkg.Folder
		err := db.NewSelect().Model(&folder).Where("id = ? AND user_id = ?", resID, userID).Scan(ctx)
		if err != nil {
			return "", fmt.Errorf("Folder not found or permission denied")
		}
		return folder.Path, nil
	}
	return "", fmt.Errorf("Invalid resource type")
}

// checkExistingShareLink : l'unicité est par (ressource, upload_only) — un dossier
// peut avoir un lien de partage classique ET un lien de demande de fichiers.
func checkExistingShareLink(ctx context.Context, db *bun.DB, userID string, resID int64, resType string, uploadOnly bool) (*pkg.ShareLink, error) {
	var existingShare pkg.ShareLink
	err := db.NewSelect().Model(&existingShare).
		Where("resource_id = ? AND resource_type = ? AND owner_id = ? AND upload_only = ?", resID, resType, userID, uploadOnly).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &existingShare, nil
}

func createNewShareLink(ctx context.Context, db *bun.DB, userID, path string, req createShareLinkRequest) (*pkg.ShareLink, error) {
	token := req.Token
	if token == "" {
		var err error
		token, err = generateToken()
		if err != nil {
			return nil, err
		}
	}

	shareLink := &pkg.ShareLink{
		ResourceID:   req.ResourceID,
		ResourceType: req.ResourceType,
		Path:         path,
		OwnerID:      userID,
		Token:        token,
		ExpiresAt:    req.ExpiresAt,
		EncryptedKey: req.EncryptedKey,
		SingleUse:    req.SingleUse,
		PermDownload: req.PermDownload,
		PermCreate:   req.PermCreate,
		PermDelete:   req.PermDelete,
		PermMove:     req.PermMove,
		UploadOnly:   req.UploadOnly,
		RequestLabel: req.RequestLabel,
	}

	// File request : dépôt seul — permissions verrouillées côté serveur quoi
	// qu'envoie le client.
	if req.UploadOnly {
		shareLink.PermCreate = true
		shareLink.PermDownload = false
		shareLink.PermDelete = false
		shareLink.PermMove = false
		shareLink.SingleUse = false
	}

	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
		if err != nil {
			return nil, err
		}
		shareLink.PasswordHash = string(hash)
	}

	_, err := db.NewInsert().Model(shareLink).Exec(ctx)
	return shareLink, err
}

func saveShareFileKeys(ctx context.Context, db *bun.DB, shareID int64, keys map[int64]string) error {
	var shareFileKeys []pkg.ShareFileKey
	for fileID, key := range keys {
		shareFileKeys = append(shareFileKeys, pkg.ShareFileKey{
			ShareID:      shareID,
			FileID:       fileID,
			EncryptedKey: key,
		})
	}
	if len(shareFileKeys) > 0 {
		_, err := db.NewInsert().Model(&shareFileKeys).Exec(ctx)
		return err
	}
	return nil
}

func getValidShareLink(ctx context.Context, db *bun.DB, token string) (*pkg.ShareLink, error) {
	var shareLink pkg.ShareLink
	err := db.NewSelect().Model(&shareLink).
		Where("token = ? AND resource_type IN ('file', 'folder')", token).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("Link not found")
	}
	if shareLink.ExpiresAt != nil && shareLink.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf(errLinkExpired)
	}
	if shareLink.SingleUse && shareLink.UsedAt != nil {
		return nil, fmt.Errorf(errLinkAlreadyUsed)
	}
	return &shareLink, nil
}

// markShareLinkUsed atomically sets used_at on a single-use link.
// Returns true if the mark succeeded (link was not yet consumed).
func markShareLinkUsed(ctx context.Context, db *bun.DB, id int64) (bool, error) {
	now := time.Now()
	res, err := db.NewUpdate().Model(&pkg.ShareLink{}).
		Set("used_at = ?", now).
		Where("id = ? AND single_use = true AND used_at IS NULL", id).
		Exec(ctx)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

func incrementShareViews(ctx context.Context, db *bun.DB, id int64) {
	_, _ = db.NewUpdate().Model(&pkg.ShareLink{}).
		Set("views = views + 1").
		Where(queryIDEq, id).
		Exec(ctx)
}

func getShareOwnerEmail(ctx context.Context, db *bun.DB, ownerID string) string {
	var owner pkg.User
	if err := db.NewSelect().Model(&owner).Where(queryIDEq, ownerID).Scan(ctx); err == nil {
		return owner.Email
	}
	return "Unknown"
}

func buildShareLinkResponse(ctx context.Context, db *bun.DB, sl *pkg.ShareLink, ownerEmail string) (gin.H, error) {
	if sl.ResourceType == "file" {
		var file pkg.File
		if err := db.NewSelect().Model(&file).Where(queryIDEq, sl.ResourceID).Scan(ctx); err != nil {
			return nil, fmt.Errorf("File not found")
		}
		return gin.H{
			"resource_type": "file",
			"resource_name": file.Name,
			"file_size":     file.Size,
			"mime_type":     file.MimeType,
			"owner_email":   ownerEmail,
			"expires_at":    sl.ExpiresAt,
			"encrypted_key": sl.EncryptedKey,
		}, nil
	}

	// Folder
	var folder pkg.Folder
	if err := db.NewSelect().Model(&folder).Where(queryIDEq, sl.ResourceID).Scan(ctx); err != nil {
		return nil, fmt.Errorf("Folder not found")
	}
	return gin.H{
		"resource_type": "folder",
		"resource_name": folder.Name,
		"owner_email":   ownerEmail,
		"expires_at":    sl.ExpiresAt,
		"upload_only":   sl.UploadOnly,
		"request_label": sl.RequestLabel,
	}, nil
}

func getSharedFile(ctx context.Context, db *bun.DB, resourceID int64) (*pkg.File, error) {
	var file pkg.File
	err := db.NewSelect().Model(&file).Where(queryIDEq, resourceID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func streamFileFromS3(c *gin.Context, ownerID string, file *pkg.File) error {
	s3Key := fmt.Sprintf("users/%s%s", ownerID, file.Path)
	s3Start := time.Now()
	output, err := s3storage.Client.GetObject(c.Request.Context(), &s3.GetObjectInput{
		Bucket: aws.String(s3storage.BucketName),
		Key:    aws.String(s3Key),
	})
	monitoring.RecordS3Duration("get", time.Since(s3Start))
	if err != nil {
		monitoring.RecordS3Request("get", false)
		return err
	}
	monitoring.RecordS3Request("get", true)
	defer output.Body.Close()

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")

	disposition := "attachment"
	if c.Query("inline") == "true" {
		disposition = "inline"
	}

	c.Header("Content-Disposition", fmt.Sprintf("%s; filename=\"%s\"", disposition, file.Name))
	c.Header("Content-Type", "application/octet-stream")
	if file.MimeType != "" {
		c.Header("Content-Type", file.MimeType)
	}
	c.Header("Content-Length", strconv.FormatInt(file.Size, 10))

	_, err = io.Copy(c.Writer, output.Body)
	return err
}

func getUserFileShareLinks(ctx context.Context, db *bun.DB, userID string, fileID int64) ([]pkg.ShareLink, error) {
	var links []pkg.ShareLink
	err := db.NewSelect().Model(&links).
		Where("resource_type = ?", "file").
		Where("resource_id = ?", fileID).
		Where("owner_id = ?", userID).
		Scan(ctx)
	return links, err
}
