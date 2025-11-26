package folders

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"safercloud/backend/pkg"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"regexp"
	"safercloud/backend/utils"
	"log"
)

var validNameRegex = regexp.MustCompile(`^[a-zA-Z0-9\s\-\._]+$`)	
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
	userIDStr, _ := userIDInterface.(string)
	userID, _ := strconv.ParseInt(userIDStr, 10, 64)

	logicalPath := filepath.ToSlash(filepath.Join(req.Path, req.Name))

	userRoot := filepath.Join("uploads", userIDStr)


	diskPath, err := utils.SecureJoin(userRoot, logicalPath)
	if err != nil {
        log.Printf("Security Alert: Path traversal attempt by user %s: %v", userIDStr, err)
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
