// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package folders

import (
	"net/http"
	"os"
	"path/filepath"

	"kagibi/backend/pkg"

	"kagibi/backend/utils"
	"log"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// \p{L} matches any Unicode letter (covers accented and non-Latin characters).
// \p{N} matches any Unicode number. The remaining chars are safe punctuation.
var validNameRegex = regexp.MustCompile(`^[\p{L}\p{N}\s\-\._]+$`)

type CreateFolderRequest struct {
	Name string `json:"name" binding:"required" validate:"required,foldername"`
	Path string `json:"path" binding:"required"`
}

func CreateHandler(c *gin.Context, db *bun.DB) {
	var req CreateFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Validation du nom (Injection XSS)
	if !validNameRegex.MatchString(req.Name) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nom de dossier invalide (caractères interdits)"})
		return
	}

	userIDInterface, _ := c.Get("user_id")
	userID, _ := userIDInterface.(string)

	logicalPath := filepath.ToSlash(filepath.Join(req.Path, req.Name))

	// Vérifier si un dossier existe déjà avec ce path pour cet utilisateur
	exists, err := pkg.FolderExistsByPath(db, userID, logicalPath)
	if err != nil {
		log.Printf("Error checking folder existence: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la vérification du dossier"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Un dossier avec ce nom existe déjà à cet emplacement"})
		return
	}

	userRoot := filepath.Join("uploads", userID)

	diskPath, err := utils.SecureJoin(userRoot, logicalPath)
	if err != nil {
		log.Printf("Security Alert: Path traversal attempt by user %s: %v", userID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Chemin invalide"})
		return
	}

	if err := os.MkdirAll(diskPath, 0755); err != nil {
		// 3. Log serveur détaillé, erreur client générique
		log.Printf("Error creating directory %s: %v", diskPath, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur interne lors de la création"})
		return
	}

	folder := &pkg.Folder{
		Name:   req.Name,
		Path:   logicalPath,
		UserID: userID,
	}

	if err := pkg.CreateFolderDB(db, folder); err != nil {
		os.RemoveAll(diskPath) // Nettoie le dossier créé sur le disque en cas d'erreur DB
		log.Printf("DB Error creating folder: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create folder"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Dossier créé avec succès", "folder": folder})
}
