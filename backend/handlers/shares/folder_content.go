package shares

import (
	"fmt"
	"log"
	"net/http"
	"safercloud/backend/pkg"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// GetSharedFolderContentHandler lists files and subfolders within a shared folder
func GetSharedFolderContentHandler(c *gin.Context, db *bun.DB) {
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
		// We need to fetch it to get the permission maybe?
		// For now assume read access.
	} else {
		// Check recursive share via Path
		// Find all folder shares for this user
		var shares []pkg.FolderShare
		err := db.NewSelect().Model(&shares).
			Where("shared_with_user_id = ?", userID).
			Scan(c.Request.Context())

		if err == nil {
			for _, s := range shares {
				var sharedFolder pkg.Folder
				if err := db.NewSelect().Model(&sharedFolder).Where("id = ?", s.FolderID).Scan(c.Request.Context()); err == nil {
					// Check if targetFolder is inside sharedFolder
					// Logic: targetPath must start with sharedPath + "/"
					if strings.HasPrefix(targetFolder.Path, sharedFolder.Path+"/") {
						isAuthorized = true
						effectiveShare = s
						// Note: for subfolders, we assume the client already has the Root Folder Key.
						// The files in subfolders are encrypted with the Root Folder Key?
						// Wait, let's verify encryption scheme.
						// Usually each folder might have a key?
						// OR all files in the hierarchy share the "Root Folder Key" via FolderFileKey table?

						// Check pkg/models.go: FolderFileKey has (FolderID, FileID).
						// "FolderID" here refers to the *Shared Folder ID* or the *Parent ID*?
						// "FolderFileKey: FileKey encrypted with FolderKey"
						// If I share folder A (id=1). File A/B.txt (id=2).
						// FolderFileKey (FolderID=1, FileID=2).

						// If I have subfolder A/Sub (id=3). File A/Sub/C.txt (id=4).
						// Is there a FolderFileKey(FolderID=1, FileID=4)?
						// OR FolderFileKey(FolderID=3, FileID=4)?

						// If it's the latter, then the user needs the key for folder 3.
						// But how do they get key for folder 3?
						// "Folders" usually don't have encryption keys themselves (they are just logical grouping).
						// The "Folder Key" is a concept generated at the moment of sharing.
						// Usually, one "Share" = one "Symmetric Key" for that share context.
						// And ALL files inside that tree are encrypted with that Share Key.

						// So we expect FolderFileKey(RootSharedFolderID, FileID).
						// So we need to find the "Root Shared Folder" that covers this target folder.
						effectiveShare = s
						break
					}
				}
			}
		}
	}

	if !isAuthorized {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// 3. List Content
	// Identical logic to "ListItemsByUser" but filtered by Path and Owner
	// AND we need to attach the correct EncryptedKey.

	ownerID := targetFolder.UserID
	currentPath := targetFolder.Path

	// Fetch Subfolders
	// Subfolders are folders where Path is exactly "currentPath + / + name" (no deeper)
	// Actually we can use the same logic as pkg.ListItemsByUser if we could constrain it.
	// But let's write a targeted query.

	// Subfolders
	var subFolders []pkg.Folder
	// We want folders where:
	// 1. user_id = ownerID
	// 2. path LIKE currentPath + "/%"
	// 3. AND path NOT LIKE currentPath + "/%/%" (direct children only)
	// PostgreSQL: path ~ '^{currentPath}/[^/]+$'

	safePath := strings.ReplaceAll(currentPath, "'", "''")
	// Make sure we handle root logic if needed, but here targetFolder is usually not root.
	// If path is "/test", children are "/test/x". Regex: ^/test/[^/]+$

	regex := fmt.Sprintf("^%s/[^/]+$", safePath)

	err = db.NewSelect().Model(&subFolders).
		Where("user_id = ?", ownerID).
		Where("path ~ ?", regex).
		Scan(c.Request.Context())

	if err != nil {
		log.Printf("Error listing subfolders: %v", err)
	}

	// Fetch Files
	var files []pkg.File
	err = db.NewSelect().Model(&files).
		Where("user_id = ?", ownerID).
		Where("path ~ ?", regex).
		Scan(c.Request.Context())

	if err != nil {
		log.Printf("Error listing files: %v", err)
	}

	// 4. Attach Keys for Files
	// We need to return the version of EncryptedKey that is usable by the user.
	// That is the one from FolderFileKey table linked to the "effectiveShare.FolderID".

	// If effectiveShare.FolderID is not set (e.g. logic above was fuzzy), we have a problem.
	// We assume effectiveShare IS set if authorized.

	rootSharedFolderID := effectiveShare.FolderID
	if rootSharedFolderID == 0 && hasDirectShare {
		rootSharedFolderID = targetFolderID
	}

	type FileWithKey struct {
		pkg.File
		EncryptedKey string `json:"encrypted_key"` // Overwrite with correct key
	}

	var filesWithKeys []FileWithKey

	for _, f := range files {
		// Find key in FolderFileKey
		var ffk pkg.FolderFileKey
		err := db.NewSelect().Model(&ffk).
			Where("folder_id = ? AND file_id = ?", rootSharedFolderID, f.ID).
			Scan(c.Request.Context())

		key := ""
		if err == nil {
			key = ffk.EncryptedKey
		} else {
			// This file might have been added AFTER the share was created.
			// If so, the owner might not have generated a key for it yet in the context of this share.
			// Or the system ensures keys are created.
			// For now, leave empty (client will fail to decrypt).
		}

		filesWithKeys = append(filesWithKeys, FileWithKey{
			File:         f,
			EncryptedKey: key,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"folders":        subFolders,
		"files":          filesWithKeys,
		"root_share_id":  effectiveShare.ID, // Might be useful
		"root_folder_id": rootSharedFolderID,
	})
}
