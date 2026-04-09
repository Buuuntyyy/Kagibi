package pkg

import (
	"context"
	"log"

	"github.com/uptrace/bun"
)

// replaceFileKeysFromShare overwrites each file's EncryptedKey with the
// share-specific key stored in share_file_keys. Files without a share key
// have their EncryptedKey cleared so the owner's key is never leaked.
func replaceFileKeysFromShare(ctx context.Context, db *bun.DB, shareID int64, files []File) {
	fileIDs := make([]int64, len(files))
	for i, f := range files {
		fileIDs[i] = f.ID
	}

	var shareKeys []ShareFileKey
	if err := db.NewSelect().Model(&shareKeys).
		Where("share_id = ?", shareID).
		Where("file_id IN (?)", bun.In(fileIDs)).
		Scan(ctx); err != nil {
		log.Printf("Error fetching share keys: %v", err)
		return
	}

	log.Printf("Found %d share keys for %d files", len(shareKeys), len(files))
	keyMap := make(map[int64]string, len(shareKeys))
	for _, k := range shareKeys {
		keyMap[k.FileID] = k.EncryptedKey
	}

	for i := range files {
		if key, ok := keyMap[files[i].ID]; ok {
			files[i].EncryptedKey = key
		} else {
			log.Printf("File %d (%s) missing share key", files[i].ID, files[i].Name)
			files[i].EncryptedKey = "" // Hide owner's key
		}
	}
}

// GetSharedFolderContent retrieves files and folders within a shared path
func GetSharedFolderContent(db *bun.DB, basePath string, ownerID string, shareID int64) ([]File, []Folder, error) {
	ctx := context.Background()
	var files []File
	var folders []Folder

	log.Printf("GetSharedFolderContent: Path=%s Owner=%s ShareID=%d", basePath, ownerID, shareID)

	// Search for items that are direct children of the basePath
	searchPrefix := basePath
	if searchPrefix == "/" {
		searchPrefix = ""
	}

	// Files directly in the folder
	err := db.NewSelect().Model(&files).
		Where("user_id = ?", ownerID).
		Where("path LIKE ? AND path NOT LIKE ?", searchPrefix+"/%", searchPrefix+"/%/%").
		Scan(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Folders directly in the folder
	err = db.NewSelect().Model(&folders).
		Where("user_id = ?", ownerID).
		Where("path LIKE ? AND path NOT LIKE ?", searchPrefix+"/%", searchPrefix+"/%/%").
		Scan(ctx)
	if err != nil {
		return nil, nil, err
	}

	if len(files) > 0 {
		replaceFileKeysFromShare(ctx, db, shareID, files)
	}

	return files, folders, nil
}
