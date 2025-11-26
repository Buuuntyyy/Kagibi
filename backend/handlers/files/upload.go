// backend/handlers/files/upload.go
package files

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"safercloud/backend/pkg"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func UploadHandler(c *gin.Context, db *bun.DB) {
	userIDInterface, _ := c.Get("user_id")
    userIDStr := userIDInterface.(string)
    userID, _ := strconv.ParseInt(userIDStr, 10, 64)

	path := c.PostForm("path") // Chemin virtuel où le fichier doit être stocké
	if path == "" {
		path = "/"
	}

	// Paramètre de chunking
	chunkIndexStr := c.PostForm("chunk_index")
	totalChunksStr := c.PostForm("total_chunks")

	isChunked := chunkIndexStr != "" && totalChunksStr != ""
	chunkIndex := 0
	totalChunks := 1
	if isChunked {
		chunkIndex, _ = strconv.Atoi(chunkIndexStr)
		totalChunks, _ = strconv.Atoi(totalChunksStr)
	}

	// Récupération du fichier (morceau actuel)
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	// Préparation des chemins
	fullPathDB := filepath.ToSlash(filepath.Join(path, fileHeader.Filename))
	userUploadDir := filepath.Join("uploads", userIDStr, path)

	if err := os.MkdirAll(userUploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user upload directory"})
		return
	}
	diskPath := filepath.Join(userUploadDir, fileHeader.Filename)

	// LOGIQUE D'ASSEMBLAGE DES MORCEAUX
	var flags int
	if isChunked && chunkIndex > 0 {
		// Si ce n'est pas le premier morceau, on l'ajoute à la fin du fichier existant
		flags = os.O_WRONLY | os.O_APPEND
	} else {
		// Premier morceau ou fichier non morcelé : créer/tronquer le fichier
		flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	}

	// Ouverture du fichier sur le disque
	dst, err := os.OpenFile(diskPath, flags, 0644)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file on disk"})
		return
	}
	defer dst.Close()

	// Lecture du morceau téléchargé
	src, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file"})
		return
	}
	defer src.Close()

	// Copie du morceau dans le fichier sur le disque
	if _, err := io.Copy(dst, src); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write chunk to final file"})
		return
	}

	// Enregistrement en base de données
	isLastChunk := !isChunked || (chunkIndex == totalChunks-1)

	if isLastChunk {
		fi, err := os.Stat(diskPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stat final file"})
			return
		}

		fileRecord := &pkg.File{
			Name:   fileHeader.Filename,
			Path:   fullPathDB,
			Size:   fi.Size(),
			MimeType: fileHeader.Header.Get("Content-Type"),
			UserID: userID,
		}

		// Vérifier si le fichier existe deja
		existsInDB, _ := db.NewSelect().Model((*pkg.File)(nil)).
			Where("user_id = ? AND path = ?", userID, fullPathDB).
			Exists(c)
		
		if existsInDB {
			// Mettre à jour l'enregistrement existant
			_, err = db.NewUpdate().Model(fileRecord).Where("user_id = ? AND path = ?", userID, fullPathDB).Exec(c)
		} else {
			// Créer un nouvel enregistrement
			err = pkg.CreateFile(db, fileRecord)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur base de données"})
            return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Upload terminé et assemblé", "file": fileRecord})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Chunk uploaded successfully", "chunk_index": chunkIndex})
	}

}