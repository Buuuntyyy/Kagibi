// internal/database.go
package pkg

import (
	"context"
	"database/sql"
	"log"
	"os"
	"sync"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func NewDB() *bun.DB {
	// Récupère l'URL de la base de données depuis les variables d'environnement
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Valeur par défaut pour le développement local
		dsn = "postgresql://user:password@127.0.0.1:5432/mydb?sslmode=disable"
	}

	// Options de connexion
	opts := []pgdriver.Option{
		pgdriver.WithDSN(dsn),
		// Force l'utilisation de l'IPv4 pour éviter les problèmes de timeout IPv6 avec Supabase
		pgdriver.WithNetwork("tcp4"),
	}

	// Si on est en local (ou si explicitement demandé), on peut désactiver SSL au niveau du driver
	// Note: Pour Supabase, il ne faut PAS utiliser WithInsecure(true)
	if dsn == "postgresql://user:password@127.0.0.1:5432/mydb?sslmode=disable" {
		opts = append(opts, pgdriver.WithInsecure(true))
	}

	// Ouvre la connexion SQL
	sqldb := sql.OpenDB(pgdriver.NewConnector(opts...))

	// Configuration du Connection Pool
	// Important pour les performances sur une connexion distante (évite de refaire le handshake SSL à chaque requête)
	sqldb.SetMaxOpenConns(20)           // Maximum de connexions ouvertes
	sqldb.SetMaxIdleConns(5)            // Garder 5 connexions inactives prêtes à l'emploi
	sqldb.SetConnMaxLifetime(time.Hour) // Recycler les connexions toutes les heures

	// Crée une instance Bun
	db := bun.NewDB(sqldb, pgdialect.New())

	return db
}

func ListUsers(db *bun.DB) ([]User, error) {
	ctx := context.Background()
	var users []User
	err := db.NewSelect().Model(&users).Scan(ctx)
	return users, err
}

func FindUserByEmail(db *bun.DB, email string) (*User, error) {
	ctx := context.Background()
	var user User
	err := db.NewSelect().Model(&user).Where("email = ?", email).Scan(ctx)
	return &user, err
}

func FindUserByID(db *bun.DB, userID string) (*User, error) {
	var user User
	err := db.NewSelect().Model(&user).Where("id = ?", userID).Scan(context.Background())
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func CreateUser(db *bun.DB, user *User) error {
	ctx := context.Background()
	_, err := db.NewInsert().Model(user).Exec(ctx)
	return err
}

func CreateFile(db *bun.DB, file *File) error {
	ctx := context.Background()
	_, err := db.NewInsert().Model(file).Exec(ctx)
	return err
}

func CreateFolderDB(db *bun.DB, folder *Folder) error {
	ctx := context.Background()
	_, err := db.NewInsert().Model(folder).Exec(ctx)
	return err
}

// Lister les fichier d'un utilisateur
func ListItemsByUser(db *bun.DB, userID string, path string) ([]FileWithShare, []FolderWithShare, error) {
	start := time.Now()
	ctx := context.Background()
	var wg sync.WaitGroup

	var filesWithShare []FileWithShare
	var foldersWithShare []FolderWithShare
	var errFiles, errFolders error

	wg.Add(2)

	// --- 1. Files Goroutine ---
	go func() {
		defer wg.Done()
		var filesPlain []File
		var err error

		// 1.1 Fetch Files
		if path == "/" {
			err = db.NewSelect().Model(&filesPlain).
				Relation("Preview").
				Where("?TableAlias.user_id = ?", userID).
				Where("?TableAlias.is_preview = ?", false).
				Where("?TableAlias.path LIKE '/%' AND ?TableAlias.path NOT LIKE '%/%/%'").
				Scan(ctx)
		} else {
			searchPrefix := path + "/"
			err = db.NewSelect().Model(&filesPlain).
				Relation("Preview").
				Where("?TableAlias.user_id = ?", userID).
				Where("?TableAlias.is_preview = ?", false).
				Where("?TableAlias.path LIKE ? AND ?TableAlias.path NOT LIKE ?", searchPrefix+"%", searchPrefix+"%/%").
				Scan(ctx)
		}

		if err != nil {
			errFiles = err
			return
		}

		if len(filesPlain) == 0 {
			filesWithShare = []FileWithShare{}
			return
		}

		// 1.2 Fetch Shares (Parallel)
		fileIds := make([]int64, len(filesPlain))
		for i, f := range filesPlain {
			fileIds[i] = f.ID
		}

		var fileLinks []ShareLink
		var directFileShares []FileShare
		var errLink, errDirect error
		var wgShares sync.WaitGroup

		wgShares.Add(2)

		go func() {
			defer wgShares.Done()
			errLink = db.NewSelect().Model(&fileLinks).
				Where("resource_type = ?", "file").
				Where("resource_id IN (?)", bun.In(fileIds)).
				Scan(ctx)
		}()

		go func() {
			defer wgShares.Done()
			errDirect = db.NewSelect().Model(&directFileShares).
				Where("file_id IN (?)", bun.In(fileIds)).
				Scan(ctx)
		}()

		wgShares.Wait()

		if errLink != nil {
			errFiles = errLink
			return
		}
		if errDirect != nil {
			errFiles = errDirect
			return
		}

		// 1.3 Construct Result
		fileLinkMap := make(map[int64]ShareLink)
		for _, l := range fileLinks {
			if _, ok := fileLinkMap[l.ResourceID]; !ok {
				fileLinkMap[l.ResourceID] = l
			}
		}

		directShareMap := make(map[int64]bool)
		for _, s := range directFileShares {
			directShareMap[s.FileID] = true
		}

		filesWithShare = make([]FileWithShare, len(filesPlain))
		for i, f := range filesPlain {
			fw := FileWithShare{File: f}
			if l, ok := fileLinkMap[f.ID]; ok {
				fw.Shared = true
				if l.OwnerID == userID {
					tok := l.Token
					fw.ShareToken = &tok
					id := l.ID
					fw.ShareID = &id
					fw.ExpiresAt = l.ExpiresAt
				}
			}
			if directShareMap[f.ID] {
				fw.Shared = true
			}
			filesWithShare[i] = fw
		}
	}()

	// --- 2. Folders Goroutine ---
	go func() {
		defer wg.Done()
		var foldersPlain []Folder
		var err error

		// 2.1 Fetch Folders
		if path == "/" {
			err = db.NewSelect().Model(&foldersPlain).
				Where("user_id = ?", userID).
				Where("path LIKE '/%' AND path NOT LIKE '%/%/%'").
				Scan(ctx)
		} else {
			searchPrefix := path + "/"
			err = db.NewSelect().Model(&foldersPlain).
				Where("user_id = ?", userID).
				Where("path LIKE ? AND path NOT LIKE ?", searchPrefix+"%", searchPrefix+"%/%").
				Scan(ctx)
		}

		if err != nil {
			errFolders = err
			return
		}

		if len(foldersPlain) == 0 {
			foldersWithShare = []FolderWithShare{}
			return
		}

		// 2.2 Fetch Shares (Parallel)
		folderIds := make([]int64, len(foldersPlain))
		for i, f := range foldersPlain {
			folderIds[i] = f.ID
		}

		var folderLinks []ShareLink
		var directFolderShares []FolderShare
		var errLink, errDirect error
		var wgShares sync.WaitGroup

		wgShares.Add(2)

		go func() {
			defer wgShares.Done()
			errLink = db.NewSelect().Model(&folderLinks).
				Where("resource_type = ?", "folder").
				Where("resource_id IN (?)", bun.In(folderIds)).
				Scan(ctx)
		}()

		go func() {
			defer wgShares.Done()
			errDirect = db.NewSelect().Model(&directFolderShares).
				Where("folder_id IN (?)", bun.In(folderIds)).
				Scan(ctx)
		}()

		wgShares.Wait()

		if errLink != nil {
			errFolders = errLink
			return
		}
		if errDirect != nil {
			errFolders = errDirect
			return
		}

		// 2.3 Construct Result
		folderLinkMap := make(map[int64]ShareLink)
		for _, l := range folderLinks {
			if _, ok := folderLinkMap[l.ResourceID]; !ok {
				folderLinkMap[l.ResourceID] = l
			}
		}

		directFolderMap := make(map[int64]bool)
		for _, s := range directFolderShares {
			directFolderMap[s.FolderID] = true
		}

		foldersWithShare = make([]FolderWithShare, len(foldersPlain))
		for i, f := range foldersPlain {
			fw := FolderWithShare{Folder: f}
			if l, ok := folderLinkMap[f.ID]; ok {
				fw.Shared = true
				if l.OwnerID == userID {
					tok := l.Token
					fw.ShareToken = &tok
					id := l.ID
					fw.ShareID = &id
					fw.ExpiresAt = l.ExpiresAt
				}
			}
			if directFolderMap[f.ID] {
				fw.Shared = true
			}
			foldersWithShare[i] = fw
		}
	}()

	wg.Wait()

	log.Printf("ListItemsByUser total time: %v", time.Since(start))

	if errFiles != nil {
		return nil, nil, errFiles
	}
	if errFolders != nil {
		return nil, nil, errFolders
	}

	return filesWithShare, foldersWithShare, nil
}

// GetAllFilesRecursive retrieves all files in a folder and its subfolders
func GetAllFilesRecursive(db *bun.DB, userID string, rootPath string) ([]File, error) {
	ctx := context.Background()
	var files []File

	searchPrefix := rootPath
	if searchPrefix != "/" {
		searchPrefix += "/"
	}

	err := db.NewSelect().Model(&files).
		Where("user_id = ?", userID).
		Where("path LIKE ?", searchPrefix+"%").
		Scan(ctx)

	return files, err
}

// supprimer un fichier
func DeleteFile(db *bun.DB, fileID int64, userID string) error {
	ctx := context.Background()

	// Delete associated share links
	_, err := db.NewDelete().Model((*ShareLink)(nil)).
		Where("resource_type = ? AND resource_id = ?", "file", fileID).
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = db.NewDelete().Model((*File)(nil)).Where("id = ? AND user_id = ?", fileID, userID).Exec(ctx)
	return err
}

func DeleteFolder(db *bun.DB, folderID int64, userID string) error {
	ctx := context.Background()

	// Delete associated share links
	_, err := db.NewDelete().Model((*ShareLink)(nil)).
		Where("resource_type = ? AND resource_id = ?", "folder", folderID).
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = db.NewDelete().Model((*Folder)(nil)).Where("id = ? AND user_id = ?", folderID, userID).Exec(ctx)
	return err
}

// Trouver un fichier par son ID
func GetFile(db *bun.DB, fileID int64, userID string) (*File, error) {
	ctx := context.Background()
	var file File
	err := db.NewSelect().Model(&file).Where("id = ? AND user_id = ?", fileID, userID).Scan(ctx)
	return &file, err
}
