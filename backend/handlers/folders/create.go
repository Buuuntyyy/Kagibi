// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package folders

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"kagibi/backend/pkg"

	"kagibi/backend/utils"
	"log"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// Block path separators, control characters, and XSS vectors.
// Path traversal is also caught by SecureJoin below.
var forbiddenNameChars = regexp.MustCompile(`[/\\\x00-\x1f<>]`)

type CreateFolderRequest struct {
	Name     string `json:"name" binding:"required" validate:"required,foldername"`
	Path     string `json:"path" binding:"required"`
	Synced   bool   `json:"synced"`    // true quand créé par la sync desktop
	IsSynced bool   `json:"is_synced"` // alias envoyé par certains clients desktop
}

// folderCreateStatus is the outcome of createOneFolder.
type folderCreateStatus string

const (
	folderStatusCreated folderCreateStatus = "created"
	folderStatusExisted folderCreateStatus = "existed"
	folderStatusInvalid folderCreateStatus = "invalid"
)

// createOneFolder contains the logic shared by CreateHandler and BatchCreateHandler:
// validate the name/path, ensure the disk directory exists, and insert the DB row
// if the folder doesn't already exist. An already-existing folder at the same path
// is treated as a no-op success (status "existed"), not an error.
// emitEvent controls the "folder_created" realtime insert+WS push: BatchCreateHandler
// passes false, since no frontend listener consumes this event type today and paying
// for an extra DB insert + WS send per item on a batch of hundreds of folders is what
// pushed a large import's request latency past the gateway's timeout (502).
func createOneFolder(ctx context.Context, db *bun.DB, userID, name, reqPath, encryptedKey string, synced, emitEvent bool) (*pkg.Folder, folderCreateStatus, string) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" || trimmed == "." || trimmed == ".." || forbiddenNameChars.MatchString(name) {
		return nil, folderStatusInvalid, "Nom de dossier invalide (caractères interdits)"
	}

	safePath, err := utils.SanitizeVirtualPath(reqPath)
	if err != nil {
		return nil, folderStatusInvalid, "Chemin invalide"
	}

	logicalPath := filepath.ToSlash(filepath.Join(safePath, name))

	exists, err := pkg.FolderExistsByPath(db, userID, logicalPath)
	if err != nil {
		log.Printf("Error checking folder existence: %v", err)
		return nil, folderStatusInvalid, "Erreur lors de la vérification du dossier"
	}
	if exists {
		return nil, folderStatusExisted, ""
	}

	userRoot := filepath.Join("uploads", userID)

	diskPath, err := utils.SecureJoin(userRoot, logicalPath)
	if err != nil {
		log.Printf("Security Alert: Path traversal attempt by user %s: %v", userID, err)
		return nil, folderStatusInvalid, "Chemin invalide"
	}

	if err := os.MkdirAll(diskPath, 0755); err != nil {
		log.Printf("Error creating directory %s: %v", diskPath, err)
		return nil, folderStatusInvalid, "Erreur interne lors de la création"
	}

	folder := &pkg.Folder{
		Name:         name,
		Path:         logicalPath,
		UserID:       userID,
		EncryptedKey: encryptedKey,
		Synced:       synced,
	}

	if err := pkg.CreateFolderDB(db, folder); err != nil {
		os.RemoveAll(diskPath) // Nettoie le dossier créé sur le disque en cas d'erreur DB
		log.Printf("DB Error creating folder: %v", err)
		return nil, folderStatusInvalid, "Failed to create folder"
	}

	if emitEvent {
		if err := pkg.EmitRealtimeEvent(ctx, db, userID, "folder_created", map[string]any{
			"id":   folder.ID,
			"path": folder.Path,
		}); err != nil {
			log.Printf("Failed to emit folder_created event: %v", err)
		}
	}

	return folder, folderStatusCreated, ""
}

func CreateHandler(c *gin.Context, db *bun.DB) {
	var req CreateFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	userIDInterface, _ := c.Get("user_id")
	userID, _ := userIDInterface.(string)

	folder, status, errMsg := createOneFolder(c.Request.Context(), db, userID, req.Name, req.Path, "", req.Synced || req.IsSynced, true)

	switch status {
	case folderStatusInvalid:
		if errMsg == "Erreur lors de la vérification du dossier" || errMsg == "Erreur interne lors de la création" || errMsg == "Failed to create folder" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
	case folderStatusExisted:
		c.JSON(http.StatusConflict, gin.H{"error": "Un dossier avec ce nom existe déjà à cet emplacement"})
	default:
		c.JSON(http.StatusCreated, gin.H{"message": "Dossier créé avec succès", "folder": folder})
	}
}
