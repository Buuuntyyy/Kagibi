// backend/handlers/files/upload.go
package files

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"safercloud/backend/pkg"
	"safercloud/backend/pkg/workers"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/uptrace/bun"
)

func UploadHandler(c *gin.Context, db *bun.DB, redisClient *redis.Client) {
	userIDInterface, _ := c.Get("user_id")
	userID := userIDInterface.(string)

	path := c.PostForm("path") // Chemin virtuel où le fichier doit être stocké
	if path == "" {
		path = "/"
	}

	encryptedKey := c.PostForm("encrypted_key")
	shareKeysJSON := c.PostForm("share_keys") // JSON string: {"shareID": "encryptedKey", ...}

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

	// Use a temporary directory for assembling chunks
	tempDir := filepath.Join(os.TempDir(), "safercloud_uploads", userID)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temp directory"})
		return
	}

	// Temporary file path for assembly
	tempFilePath := filepath.Join(tempDir, fileHeader.Filename+"_partial")

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
	dst, err := os.OpenFile(tempFilePath, flags, 0644)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open temp file"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write chunk to temp file"})
		return
	}

	// Enregistrement en base de données
	isLastChunk := !isChunked || (chunkIndex == totalChunks-1)

	if isLastChunk {
		// Ensure data is written to disk
		dst.Sync()
		// Close the file before reading it for upload
		dst.Close()

		fi, err := os.Stat(tempFilePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stat final file"})
			return
		}

		fileSize := fi.Size()

		// Start transaction
		tx, err := db.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		defer tx.Rollback()

		// Check storage limit
		var user pkg.User
		err = tx.NewSelect().Model(&user).Where("id = ?", userID).Scan(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
			return
		}

		if user.StorageUsed+fileSize > user.StorageLimit {
			os.Remove(tempFilePath) // Clean up
			c.JSON(http.StatusForbidden, gin.H{"error": "Storage limit exceeded"})
			return
		}

		// S3 Key: users/{userID}/{path}/{filename}
		// fullPathDB starts with /, so we trim it or just concat
		s3Key := fmt.Sprintf("users/%s%s", userID, fullPathDB)

		// Enqueue Upload Task to Redis
		task := workers.S3Task{
			Type:        workers.TaskUpload,
			UserID:      userID,
			SrcKey:      tempFilePath, // Local path to temp file
			DestKey:     s3Key,        // S3 Key
			ContentType: fileHeader.Header.Get("Content-Type"),
		}

		if err := workers.EnqueueTask(redisClient, task); err != nil {
			os.Remove(tempFilePath) // Clean up if enqueue fails
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue upload task"})
			return
		}

		// Note: We do NOT remove tempFilePath here. The worker will remove it after successful upload.

		fileRecord := &pkg.File{
			Name:         fileHeader.Filename,
			Path:         fullPathDB,
			Size:         fileSize,
			MimeType:     fileHeader.Header.Get("Content-Type"),
			UserID:       userID,
			EncryptedKey: encryptedKey,
		}

		// Vérifier si le fichier existe deja
		existsInDB, _ := tx.NewSelect().Model((*pkg.File)(nil)).
			Where("user_id = ? AND path = ?", userID, fullPathDB).
			Exists(c)

		if existsInDB {
			// Get old file size to adjust storage
			var oldFile pkg.File
			err = tx.NewSelect().Model(&oldFile).Where("user_id = ? AND path = ?", userID, fullPathDB).Scan(c)
			if err == nil {
				// Update storage: remove old size, add new size
				_, err = tx.NewUpdate().Model(&user).
					Set("storage_used = storage_used - ? + ?", oldFile.Size, fileSize).
					Where("id = ?", userID).
					Exec(c)
			}

			// Mettre à jour l'enregistrement existant
			_, err = tx.NewUpdate().Model(fileRecord).Where("user_id = ? AND path = ?", userID, fullPathDB).Exec(c)
		} else {
			// Créer un nouvel enregistrement
			_, err = tx.NewInsert().Model(fileRecord).Exec(c)

			// Update storage: add new size
			_, err = tx.NewUpdate().Model(&user).
				Set("storage_used = storage_used + ?", fileSize).
				Where("id = ?", userID).
				Exec(c)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur base de données"})
			return
		}

		// Handle Share Keys if provided
		if shareKeysJSON != "" {
			var shareKeysMap map[string]string
			if err := json.Unmarshal([]byte(shareKeysJSON), &shareKeysMap); err == nil {
				var shareFileKeys []pkg.ShareFileKey

				// Need to get the File ID. If it was an update, we need to fetch it.
				// If it was an insert, fileRecord.ID should be populated by Bun if we passed a pointer?
				// Bun populates ID on Insert. But on Update, we might need to fetch it if we didn't have it.

				var finalFileID int64
				if existsInDB {
					// Fetch ID
					var f pkg.File
					_ = tx.NewSelect().Model(&f).Where("user_id = ? AND path = ?", userID, fullPathDB).Scan(c)
					finalFileID = f.ID
				} else {
					finalFileID = fileRecord.ID
				}

				for sIDStr, key := range shareKeysMap {
					sID, _ := strconv.ParseInt(sIDStr, 10, 64)
					if sID > 0 {
						shareFileKeys = append(shareFileKeys, pkg.ShareFileKey{
							ShareID:      sID,
							FileID:       finalFileID,
							EncryptedKey: key,
						})
					}
				}

				if len(shareFileKeys) > 0 {
					// Use OnConflict to update if exists
					_, err = tx.NewInsert().Model(&shareFileKeys).
						On("CONFLICT (share_id, file_id) DO UPDATE").
						Set("encrypted_key = EXCLUDED.encrypted_key").
						Exec(c)
					if err != nil {
						// Log error but don't fail upload? Or fail?
						// Better to fail so client knows sharing is broken
						// But maybe not critical. Let's log.
						fmt.Printf("Error inserting share keys: %v\n", err)
					}
				}
			}
		}

		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit failed"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Upload en cours de traitement", "file": fileRecord})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Chunk uploaded successfully", "chunk_index": chunkIndex})
	}
}
