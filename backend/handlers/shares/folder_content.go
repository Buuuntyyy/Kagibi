package shares

import (
	"context"
	"kagibi/backend/pkg"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type FolderContentResponse struct {
	Folders      []FolderWithKey `json:"folders"`
	Files        []FileWithKey   `json:"files"`
	RootShareID  int64           `json:"root_share_id"`
	RootFolderID int64           `json:"root_folder_id"`
}

type FileWithKey struct {
	pkg.File
	EncryptedKey string `json:"encrypted_key"`
}

type FolderWithKey struct {
	pkg.Folder
	EncryptedKey string `json:"encrypted_key"`
}

// GetSharedFolderContentHandler lists files and subfolders within a shared folder
func GetSharedFolderContentHandler(c *gin.Context, db *bun.DB) {
	userID := c.GetString("user_id")
	targetFolderID, _ := strconv.ParseInt(c.Param("folderID"), 10, 64)

	targetFolder, err := getTargetFolder(c.Request.Context(), db, targetFolderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}

	share, isAuthorized := checkFolderShareAccess(c.Request.Context(), db, userID, targetFolder)
	if !isAuthorized {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	files, folders, err := listContentItems(c.Request.Context(), db, targetFolder.UserID, targetFolder.Path)
	if err != nil {
		log.Printf("Error listing content: %v", err)
		// Continue with empty lists or return error? Original continued?
		// Original handled errors separately but didn't return.
	}

	rootSharedFolderID := share.FolderID
	if rootSharedFolderID == 0 {
		// Should not happen if isAuthorized is true, unless share struct is empty?
		// share is *pkg.FolderShare.
		rootSharedFolderID = targetFolderID
	}

	resp := enrichContentWithKeys(c.Request.Context(), db, rootSharedFolderID, files, folders)
	resp.RootShareID = share.ID
	resp.RootFolderID = rootSharedFolderID

	c.JSON(http.StatusOK, resp)
}

func getTargetFolder(ctx context.Context, db *bun.DB, folderID int64) (*pkg.Folder, error) {
	var folder pkg.Folder
	err := db.NewSelect().Model(&folder).Where("id = ?", folderID).Scan(ctx)
	return &folder, err
}

func checkFolderShareAccess(ctx context.Context, db *bun.DB, userID string, targetFolder *pkg.Folder) (*pkg.FolderShare, bool) {
	// 1. Direct Share Check
	var directShare pkg.FolderShare
	err := db.NewSelect().Model(&directShare).
		Where("folder_id = ? AND shared_with_user_id = ?", targetFolder.ID, userID).
		Scan(ctx)

	if err == nil {
		return &directShare, true
	}

	// 2. Recursive Share Check
	type FolderShareWithFolder struct {
		bun.BaseModel `bun:"table:folder_shares"`
		pkg.FolderShare
		Folder *pkg.Folder `bun:"rel:belongs-to,join:folder_id=id"`
	}

	var sharesWithFolders []FolderShareWithFolder
	err = db.NewSelect().
		Model(&sharesWithFolders).
		Relation("Folder").
		Where("shared_with_user_id = ?", userID).
		Scan(ctx)

	if err != nil {
		return nil, false
	}

	for _, s := range sharesWithFolders {
		if s.Folder == nil {
			continue
		}
		prefix := s.Folder.Path
		if !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		if strings.HasPrefix(targetFolder.Path, prefix) {
			return &s.FolderShare, true
		}
	}

	return nil, false
}

func listContentItems(ctx context.Context, db *bun.DB, ownerID string, currentPath string) ([]pkg.File, []pkg.Folder, error) {
	var subFolders []pkg.Folder
	err := db.NewSelect().Model(&subFolders).
		Where("user_id = ?", ownerID).
		Where("path LIKE ?", currentPath+"/%").
		Where("path NOT LIKE ?", currentPath+"/%/%").
		Scan(ctx)
	if err != nil {
		return nil, nil, err
	}

	var files []pkg.File
	err = db.NewSelect().Model(&files).
		Where("user_id = ?", ownerID).
		Where("path LIKE ?", currentPath+"/%").
		Where("path NOT LIKE ?", currentPath+"/%/%").
		Where("is_preview = ?", false).
		Scan(ctx)

	return files, subFolders, err
}

func enrichContentWithKeys(ctx context.Context, db *bun.DB, rootFolderID int64, files []pkg.File, folders []pkg.Folder) FolderContentResponse {
	filesWithKeys := enrichFilesWithKeys(ctx, db, rootFolderID, files)
	foldersWithKeys := enrichFoldersWithKeys(ctx, db, rootFolderID, folders)
	return FolderContentResponse{Files: filesWithKeys, Folders: foldersWithKeys}
}

func enrichFilesWithKeys(ctx context.Context, db *bun.DB, folderID int64, files []pkg.File) []FileWithKey {
	var fileIDs []int64
	for _, f := range files {
		fileIDs = append(fileIDs, f.ID)
	}
	keyMap := make(map[int64]string)
	if len(fileIDs) > 0 {
		var keys []pkg.FolderFileKey
		if err := db.NewSelect().Model(&keys).
			Where("folder_id = ?", folderID).
			Where("file_id IN (?)", bun.In(fileIDs)).
			Scan(ctx); err == nil {
			for _, k := range keys {
				keyMap[k.FileID] = k.EncryptedKey
			}
		}
	}
	result := make([]FileWithKey, len(files))
	for i, f := range files {
		result[i] = FileWithKey{File: f, EncryptedKey: keyMap[f.ID]}
	}
	return result
}

func enrichFoldersWithKeys(ctx context.Context, db *bun.DB, parentFolderID int64, folders []pkg.Folder) []FolderWithKey {
	var folderIDs []int64
	for _, f := range folders {
		folderIDs = append(folderIDs, f.ID)
	}
	keyMap := make(map[int64]string)
	if len(folderIDs) > 0 {
		var keys []pkg.FolderFolderKey
		if err := db.NewSelect().Model(&keys).
			Where("parent_folder_id = ?", parentFolderID).
			Where("sub_folder_id IN (?)", bun.In(folderIDs)).
			Scan(ctx); err == nil {
			for _, k := range keys {
				keyMap[k.SubFolderID] = k.EncryptedKey
			}
		}
	}
	result := make([]FolderWithKey, len(folders))
	for i, f := range folders {
		result[i] = FolderWithKey{Folder: f, EncryptedKey: keyMap[f.ID]}
	}
	return result
}
