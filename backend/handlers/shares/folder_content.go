package shares

import (
	"log"
	"net/http"
	"safercloud/backend/pkg"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// GetSharedFolderContentHandler lists files and subfolders within a shared folder
func GetSharedFolderContentHandler(c *gin.Context, db *bun.DB) {
	startTotal := time.Now()

	userIDInterface, _ := c.Get("user_id")
	userID := userIDInterface.(string)

	folderIDStr := c.Param("folderID")
	targetFolderID, err := strconv.ParseInt(folderIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid folder ID"})
		return
	}

	// 1. Get the target folder details
	var targetFolder pkg.Folder
	err = db.NewSelect().Model(&targetFolder).Where("id = ?", targetFolderID).Scan(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}

	// 2. Verify Access
	startAuth := time.Now()
	// Access is granted if:
	// A) The folder is directly shared with the user
	// B) The folder is a subfolder of a folder shared with the user

	// First, check direct share
	hasDirectShare, err := db.NewSelect().Model((*pkg.FolderShare)(nil)).
		Where("folder_id = ? AND shared_with_user_id = ?", targetFolderID, userID).
		Exists(c.Request.Context())

	if err != nil {
		log.Printf("Error checking direct share: %v", err)
	}

	var effectiveShare pkg.FolderShare
	isAuthorized := hasDirectShare

	if hasDirectShare {
		// Use the direct share key
		_ = db.NewSelect().Model(&effectiveShare).
			Where("folder_id = ? AND shared_with_user_id = ?", targetFolderID, userID).
			Scan(c.Request.Context())
	} else {
		// Check recursive share via Path
		// Optimization: Fetch all folder shares AND their associated Folder definitions in one query.
		type FolderShareWithFolder struct {
			pkg.FolderShare
			Folder *pkg.Folder `bun:"rel:belongs-to,join:folder_id=id"`
		}

		var sharesWithFolders []FolderShareWithFolder
		err := db.NewSelect().
			Model(&sharesWithFolders).
			Relation("Folder").
			Where("shared_with_user_id = ?", userID).
			Scan(c.Request.Context())

		if err == nil {
			for _, s := range sharesWithFolders {
				if s.Folder == nil {
					continue
				}
				// Check if targetFolder is inside sharedFolder
				if strings.HasPrefix(targetFolder.Path, s.Folder.Path+"/") {
					isAuthorized = true
					effectiveShare = s.FolderShare
					break
				}
			}
		}
	}
	log.Printf("[Perf] Auth Check took %v", time.Since(startAuth))

	if !isAuthorized {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// 3. List Content
	startList := time.Now()
	ownerID := targetFolder.UserID
	currentPath := targetFolder.Path

	// Subfolders
	var subFolders []pkg.Folder
	// Optimized: Use LIKE instead of Regex for performance

	err = db.NewSelect().Model(&subFolders).
		Where("user_id = ?", ownerID).
		Where("path LIKE ?", currentPath+"/%").
		Where("path NOT LIKE ?", currentPath+"/%/%").
		Scan(c.Request.Context())

	if err != nil {
		log.Printf("Error listing subfolders: %v", err)
	}

	// Fetch Files
	var files []pkg.File
	err = db.NewSelect().Model(&files).
		Where("user_id = ?", ownerID).
		Where("path LIKE ?", currentPath+"/%").
		Where("path NOT LIKE ?", currentPath+"/%/%").
		Scan(c.Request.Context())

	if err != nil {
		log.Printf("Error listing files: %v", err)
	}
	log.Printf("[Perf] Content Listing took %v", time.Since(startList))

	// 4. Attach Keys for Files AND SUBFOLDERS
	startKeys := time.Now()
	rootSharedFolderID := effectiveShare.FolderID
	if rootSharedFolderID == 0 && hasDirectShare {
		rootSharedFolderID = targetFolderID
	}

	// Optimization: Bulk fetch keys for files
	var fileIDs []int64
	for _, f := range files {
		fileIDs = append(fileIDs, f.ID)
	}

	fileKeyMap := make(map[int64]string)

	if len(fileIDs) > 0 {
		var keys []pkg.FolderFileKey
		err := db.NewSelect().Model(&keys).
			Where("folder_id = ?", rootSharedFolderID).
			Where("file_id IN (?)", bun.In(fileIDs)).
			Scan(c.Request.Context())

		if err != nil {
			log.Printf("Error fetching file keys: %v", err)
		} else {
			for _, k := range keys {
				fileKeyMap[k.FileID] = k.EncryptedKey
			}
		}
	}

	type FileWithKey struct {
		pkg.File
		EncryptedKey string `json:"encrypted_key"`
	}

	var filesWithKeys []FileWithKey

	for _, f := range files {
		key := fileKeyMap[f.ID]
		filesWithKeys = append(filesWithKeys, FileWithKey{
			File:         f,
			EncryptedKey: key,
		})
	}

	// Optimization: Bulk fetch keys for subfolders
	var subFolderIDs []int64
	for _, f := range subFolders {
		subFolderIDs = append(subFolderIDs, f.ID)
	}

	folderKeyMap := make(map[int64]string)
	if len(subFolderIDs) > 0 {
		var keys []pkg.FolderFolderKey
		err := db.NewSelect().Model(&keys).
			Where("parent_folder_id = ?", rootSharedFolderID).
			Where("sub_folder_id IN (?)", bun.In(subFolderIDs)).
			Scan(c.Request.Context())
			
		if err != nil {
			log.Printf("Error fetching folder keys: %v", err)
		} else {
			for _, k := range keys {
				folderKeyMap[k.SubFolderID] = k.EncryptedKey
			}
		}
	}

	type FolderWithKey struct {
		pkg.Folder
		EncryptedKey string `json:"encrypted_key"`
	}

	var foldersWithKeys []FolderWithKey
	for _, f := range subFolders {
		key := folderKeyMap[f.ID]
		// Fallback: If no shared key found, and it's a direct share,
		// maybe the client can try decrypting the original key? 
		// But usually we need the explicit shared key.
		foldersWithKeys = append(foldersWithKeys, FolderWithKey{
			Folder:       f,
			EncryptedKey: key,
		})
	}

	log.Printf("[Perf] Key Processing took %v", time.Since(startKeys))
	log.Printf("[Perf] Total Handler took %v", time.Since(startTotal))

	c.JSON(http.StatusOK, gin.H{
		"folders":        foldersWithKeys,
		"files":          filesWithKeys,
		"root_share_id":  effectiveShare.ID,
		"root_folder_id": rootSharedFolderID,
	})
}
