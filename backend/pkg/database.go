// internal/database.go
package pkg

import (
	"context"
	"database/sql"
	"log"
	"os"
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
	var filesPlain []File
	var foldersPlain []Folder
	var err error

	if path == "/" {
		// Pour la racine : on cherche les chemins qui commencent par '/' mais qui n'ont pas de deuxième '/'.
		// Ex: '/test' (OK), '/image.jpg' (OK), mais pas '/test/doc.pdf' (NON)
		err = db.NewSelect().Model(&filesPlain).
			Where("user_id = ?", userID).
			Where("path LIKE '/%' AND path NOT LIKE '%/%/%'").
			Scan(ctx)
		if err != nil {
			return nil, nil, err
		}
		log.Printf("Files query took: %v", time.Since(start))

		t2 := time.Now()
		err = db.NewSelect().Model(&foldersPlain).
			Where("user_id = ?", userID).
			Where("path LIKE '/%' AND path NOT LIKE '%/%/%'").
			Scan(ctx)
		if err != nil {
			return nil, nil, err
		}
		log.Printf("Folders query took: %v", time.Since(t2))
	} else {
		// Pour un sous-dossier (ex: /test) : on cherche les chemins qui commencent par '/test/'
		// mais qui n'ont pas de '/' supplémentaire après.
		searchPrefix := path + "/"
		err = db.NewSelect().Model(&filesPlain).
			Where("user_id = ?", userID).
			Where("path LIKE ? AND path NOT LIKE ?", searchPrefix+"%", searchPrefix+"%/%").
			Scan(ctx)
		if err != nil {
			return nil, nil, err
		}
		log.Printf("Files query took: %v", time.Since(start))

		t2 := time.Now()
		err = db.NewSelect().Model(&foldersPlain).
			Where("user_id = ?", userID).
			Where("path LIKE ? AND path NOT LIKE ?", searchPrefix+"%", searchPrefix+"%/%").
			Scan(ctx)
		if err != nil {
			return nil, nil, err
		}
		log.Printf("Folders query took: %v", time.Since(t2))
	}

	// --- Traitement des fichiers ---
	filesWithShare := make([]FileWithShare, 0, len(filesPlain))
	if len(filesPlain) > 0 {
		t3 := time.Now()
		fileIds := make([]int64, 0, len(filesPlain))
		for _, f := range filesPlain {
			fileIds = append(fileIds, f.ID)
		}

		var fileLinks []ShareLink
		err = db.NewSelect().Model(&fileLinks).
			Where("resource_type = ?", "file").
			Where("resource_id IN (?)", bun.In(fileIds)).
			Scan(ctx)
		if err != nil {
			return nil, nil, err
		}
		log.Printf("File shares query took: %v", time.Since(t3))

		fileLinkMap := make(map[int64]ShareLink)
		for _, l := range fileLinks {
			if _, ok := fileLinkMap[l.ResourceID]; !ok {
				fileLinkMap[l.ResourceID] = l
			}
		}

		for _, f := range filesPlain {
			fw := FileWithShare{File: f}
			if l, ok := fileLinkMap[f.ID]; ok {
				fw.Shared = true
				if l.OwnerID == userID {
					tok := l.Token
					fw.ShareToken = &tok
					id := l.ID
					fw.ShareID = &id
				}
			}
			filesWithShare = append(filesWithShare, fw)
		}
	}

	// --- Traitement des dossiers ---
	foldersWithShare := make([]FolderWithShare, 0, len(foldersPlain))
	if len(foldersPlain) > 0 {
		t4 := time.Now()
		folderIds := make([]int64, 0, len(foldersPlain))
		for _, f := range foldersPlain {
			folderIds = append(folderIds, f.ID)
		}

		var folderLinks []ShareLink
		err = db.NewSelect().Model(&folderLinks).
			Where("resource_type = ?", "folder").
			Where("resource_id IN (?)", bun.In(folderIds)).
			Scan(ctx)
		if err != nil {
			return nil, nil, err
		}
		log.Printf("Folder shares query took: %v", time.Since(t4))

		folderLinkMap := make(map[int64]ShareLink)
		for _, l := range folderLinks {
			if _, ok := folderLinkMap[l.ResourceID]; !ok {
				folderLinkMap[l.ResourceID] = l
			}
		}

		for _, f := range foldersPlain {
			fw := FolderWithShare{Folder: f}
			if l, ok := folderLinkMap[f.ID]; ok {
				fw.Shared = true
				if l.OwnerID == userID {
					tok := l.Token
					fw.ShareToken = &tok
					id := l.ID
					fw.ShareID = &id
				}
			}
			foldersWithShare = append(foldersWithShare, fw)
		}
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
