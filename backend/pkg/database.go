// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

// internal/database.go
package pkg

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"kagibi/backend/pkg/emailcrypto"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

const (
	queryPathEq      = "path = ?"
	queryUserIDEq    = "user_id = ?"
	queryTableUserID = "?TableAlias.user_id = ?"
	queryIDAndUserID = "id = ? AND user_id = ?"
	queryPathLike    = "path LIKE ?"
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

// DecryptUserEmail decrypts user.EmailEncrypted and stores the result in user.Email.
// Must be called after any DB load that needs the plaintext email.
func DecryptUserEmail(u *User) error {
	if u.EmailEncrypted == "" {
		return nil
	}
	plain, err := emailcrypto.Decrypt(u.EmailEncrypted)
	if err != nil {
		return fmt.Errorf("DecryptUserEmail: %w", err)
	}
	u.Email = plain
	return nil
}

func FindUserByEmail(db *bun.DB, email string) (*User, error) {
	ctx := context.Background()
	var user User
	hash := emailcrypto.Hash(email)
	err := db.NewSelect().Model(&user).Where("email_hash = ?", hash).Scan(ctx)
	if err != nil {
		return nil, err
	}
	if err := DecryptUserEmail(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func FindUserByID(db *bun.DB, userID string) (*User, error) {
	var user User
	err := db.NewSelect().Model(&user).Where("id = ?", userID).Scan(context.Background())
	if err != nil {
		return nil, err
	}
	if err := DecryptUserEmail(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func CreateUser(db *bun.DB, user *User) error {
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.NewInsert().Model(user).Exec(ctx); err != nil {
		return err
	}

	planState := &UserPlan{
		UserID:           user.ID,
		Plan:             PlanFree,
		StorageLimit:     StorageFree,
		StorageUsed:      0,
		P2PMaxExchanges:  P2PLimitFree,
		P2PExchangesUsed: 0,
	}
	if _, err = tx.NewInsert().Model(planState).Exec(ctx); err != nil {
		return err
	}

	err = tx.Commit()
	return err
}

func FindUserPlanByUserID(db *bun.DB, userID string) (*UserPlan, error) {
	var plan UserPlan
	err := db.NewSelect().Model(&plan).Where(queryUserIDEq, userID).Scan(context.Background())
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func CountUserActiveP2PExchanges(db *bun.DB, userID string) (int, error) {
	return db.NewSelect().TableExpr("file_shares fs").
		Join("JOIN files f ON f.id = fs.file_id").
		Where("f.user_id = ?", userID).
		Count(context.Background())
}

func UpsertUserPlan(db *bun.DB, plan *UserPlan) error {
	_, err := db.NewInsert().Model(plan).
		On("CONFLICT (user_id) DO UPDATE").
		Set("plan = EXCLUDED.plan").
		Set("storage_limit = EXCLUDED.storage_limit").
		Set("storage_used = EXCLUDED.storage_used").
		Set("p2p_max_exchanges = EXCLUDED.p2p_max_exchanges").
		Set("p2p_exchanges_used = EXCLUDED.p2p_exchanges_used").
		Set("updated_at = CURRENT_TIMESTAMP").
		Exec(context.Background())
	return err
}

func CreateFile(db *bun.DB, file *File) error {
	ctx := context.Background()
	_, err := db.NewInsert().Model(file).Exec(ctx)
	return err
}

// FolderExistsByPath vérifie si un dossier existe déjà pour cet utilisateur avec ce path exact
func FolderExistsByPath(db *bun.DB, userID string, path string) (bool, error) {
	ctx := context.Background()
	count, err := db.NewSelect().
		Model((*Folder)(nil)).
		Where(queryUserIDEq, userID).
		Where(queryPathEq, path).
		Count(ctx)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func CreateFolderDB(db *bun.DB, folder *Folder) error {
	ctx := context.Background()
	_, err := db.NewInsert().Model(folder).Returning("id").Exec(ctx)
	if err != nil {
		return err
	}

	fs := &FolderSize{
		FolderID:  folder.ID,
		UserID:    folder.UserID,
		SizeBytes: 0,
		UpdatedAt: time.Now(),
	}
	_, err = db.NewInsert().Model(fs).
		Column("folder_id", "user_id", "size_bytes", "updated_at").
		On("CONFLICT (folder_id) DO NOTHING").
		Exec(ctx)
	return err
}

// fetchFilesInPath fetches plain files for a given user and directory path.
func fetchFilesInPath(ctx context.Context, db *bun.DB, userID, path string) ([]File, error) {
	var filesPlain []File
	var err error
	if path == "/" {
		err = db.NewSelect().Model(&filesPlain).
			Relation("Preview").
			Where(queryTableUserID, userID).
			Where("?TableAlias.is_preview = ?", false).
			Where("?TableAlias.path LIKE '/%' AND ?TableAlias.path NOT LIKE '%/%/%'").
			Scan(ctx)
	} else {
		searchPrefix := path + "/"
		err = db.NewSelect().Model(&filesPlain).
			Relation("Preview").
			Where(queryTableUserID, userID).
			Where("?TableAlias.is_preview = ?", false).
			Where("?TableAlias.path LIKE ? AND ?TableAlias.path NOT LIKE ?", searchPrefix+"%", searchPrefix+"%/%").
			Scan(ctx)
	}
	return filesPlain, err
}

// fetchFileShareData fetches share links and direct file shares for a set of file IDs in parallel.
func fetchFileShareData(ctx context.Context, db *bun.DB, fileIds []int64) ([]ShareLink, []FileShare, error) {
	var fileLinks []ShareLink
	var directFileShares []FileShare
	var errLink, errDirect error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		errLink = db.NewSelect().Model(&fileLinks).
			Where("resource_type = ?", "file").
			Where("resource_id IN (?)", bun.In(fileIds)).
			Scan(ctx)
	}()
	go func() {
		defer wg.Done()
		errDirect = db.NewSelect().Model(&directFileShares).
			Where("file_id IN (?)", bun.In(fileIds)).
			Scan(ctx)
	}()
	wg.Wait()
	if errLink != nil {
		return nil, nil, errLink
	}
	return fileLinks, directFileShares, errDirect
}

// buildFilesWithShare annotates plain files with share metadata.
func buildFilesWithShare(userID string, filesPlain []File, fileLinks []ShareLink, directFileShares []FileShare) []FileWithShare {
	fileLinkMap := make(map[int64]ShareLink, len(fileLinks))
	for _, l := range fileLinks {
		if _, ok := fileLinkMap[l.ResourceID]; !ok {
			fileLinkMap[l.ResourceID] = l
		}
	}
	directShareMap := make(map[int64]bool, len(directFileShares))
	for _, s := range directFileShares {
		directShareMap[s.FileID] = true
	}
	result := make([]FileWithShare, len(filesPlain))
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
		result[i] = fw
	}
	return result
}

// listFilesWithShares fetches files for a user/path and attaches share info.
func listFilesWithShares(ctx context.Context, db *bun.DB, userID, path string) ([]FileWithShare, error) {
	filesPlain, err := fetchFilesInPath(ctx, db, userID, path)
	if err != nil {
		return nil, err
	}
	if len(filesPlain) == 0 {
		return []FileWithShare{}, nil
	}
	fileIds := make([]int64, len(filesPlain))
	for i, f := range filesPlain {
		fileIds[i] = f.ID
	}
	fileLinks, directFileShares, err := fetchFileShareData(ctx, db, fileIds)
	if err != nil {
		return nil, err
	}
	return buildFilesWithShare(userID, filesPlain, fileLinks, directFileShares), nil
}

// fetchFoldersInPath fetches plain folders for a given user and directory path.
func fetchFoldersInPath(ctx context.Context, db *bun.DB, userID, path string, includeSizes bool) ([]Folder, error) {
	var foldersPlain []Folder
	q := db.NewSelect().Model(&foldersPlain)
	if includeSizes {
		q = q.
			ColumnExpr("?TableAlias.*").
			ColumnExpr("COALESCE(fs.size_bytes, 0) AS size_bytes").
			Join("LEFT JOIN folder_sizes AS fs ON fs.folder_id = ?TableAlias.id")
	}
	var err error
	if path == "/" {
		err = q.Where(queryTableUserID, userID).
			Where("?TableAlias.path LIKE '/%' AND ?TableAlias.path NOT LIKE '%/%/%'").
			Scan(ctx)
	} else {
		searchPrefix := path + "/"
		err = q.Where(queryTableUserID, userID).
			Where("?TableAlias.path LIKE ? AND ?TableAlias.path NOT LIKE ?", searchPrefix+"%", searchPrefix+"%/%").
			Scan(ctx)
	}
	return foldersPlain, err
}

// fetchFolderShareData fetches share links and direct folder shares for a set of folder IDs in parallel.
func fetchFolderShareData(ctx context.Context, db *bun.DB, folderIds []int64) ([]ShareLink, []FolderShare, error) {
	var folderLinks []ShareLink
	var directFolderShares []FolderShare
	var errLink, errDirect error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		errLink = db.NewSelect().Model(&folderLinks).
			Where("resource_type = ?", "folder").
			Where("resource_id IN (?)", bun.In(folderIds)).
			Scan(ctx)
	}()
	go func() {
		defer wg.Done()
		errDirect = db.NewSelect().Model(&directFolderShares).
			Where("folder_id IN (?)", bun.In(folderIds)).
			Scan(ctx)
	}()
	wg.Wait()
	if errLink != nil {
		return nil, nil, errLink
	}
	return folderLinks, directFolderShares, errDirect
}

// buildFoldersWithShare annotates plain folders with share metadata.
func buildFoldersWithShare(userID string, foldersPlain []Folder, folderLinks []ShareLink, directFolderShares []FolderShare) []FolderWithShare {
	folderLinkMap := make(map[int64]ShareLink, len(folderLinks))
	for _, l := range folderLinks {
		if _, ok := folderLinkMap[l.ResourceID]; !ok {
			folderLinkMap[l.ResourceID] = l
		}
	}
	directFolderMap := make(map[int64]bool, len(directFolderShares))
	for _, s := range directFolderShares {
		directFolderMap[s.FolderID] = true
	}
	result := make([]FolderWithShare, len(foldersPlain))
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
		result[i] = fw
	}
	return result
}

// listFoldersWithShares fetches folders for a user/path and attaches share info.
func listFoldersWithShares(ctx context.Context, db *bun.DB, userID, path string, includeSizes bool) ([]FolderWithShare, error) {
	foldersPlain, err := fetchFoldersInPath(ctx, db, userID, path, includeSizes)
	if err != nil {
		return nil, err
	}
	if len(foldersPlain) == 0 {
		return []FolderWithShare{}, nil
	}
	folderIds := make([]int64, len(foldersPlain))
	for i, f := range foldersPlain {
		folderIds[i] = f.ID
	}
	folderLinks, directFolderShares, err := fetchFolderShareData(ctx, db, folderIds)
	if err != nil {
		return nil, err
	}
	return buildFoldersWithShare(userID, foldersPlain, folderLinks, directFolderShares), nil
}

// Lister les fichier d'un utilisateur
func ListItemsByUser(db *bun.DB, userID string, path string, includeFolderSizes bool) ([]FileWithShare, []FolderWithShare, error) {
	start := time.Now()
	ctx := context.Background()
	var wg sync.WaitGroup

	var filesWithShare []FileWithShare
	var foldersWithShare []FolderWithShare
	var errFiles, errFolders error

	wg.Add(2)

	go func() {
		defer wg.Done()
		filesWithShare, errFiles = listFilesWithShares(ctx, db, userID, path)
	}()

	go func() {
		defer wg.Done()
		foldersWithShare, errFolders = listFoldersWithShares(ctx, db, userID, path, includeFolderSizes)
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
		Where(queryUserIDEq, userID).
		Where(queryPathLike, searchPrefix+"%").
		Scan(ctx)

	return files, err
}

// GetFolderContentRecursive retrieves all files AND folders in a folder and its subfolders
func GetFolderContentRecursive(db *bun.DB, userID string, rootPath string) ([]File, []Folder, error) {
	ctx := context.Background()
	var files []File
	var folders []Folder

	// Logic correction: searchPrefix + "%" only matches subdirectories if it ends with /
	// We want direct children (path = rootPath) AND recursive children (path LIKE rootPath/%)

	// Files
	qFiles := db.NewSelect().Model(&files).Where(queryUserIDEq, userID)
	if rootPath == "/" {
		qFiles.Where(queryPathLike, "/%")
	} else {
		// Parenthesis are important for OR
		qFiles.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where(queryPathEq, rootPath).
				WhereOr(queryPathLike, rootPath+"/%")
		})
	}
	err := qFiles.Scan(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Folders
	qFolders := db.NewSelect().Model(&folders).Where(queryUserIDEq, userID)
	if rootPath == "/" {
		qFolders.Where(queryPathLike, "/%")
	} else {
		qFolders.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where(queryPathEq, rootPath).
				WhereOr(queryPathLike, rootPath+"/%")
		})
	}
	err = qFolders.Scan(ctx)

	return files, folders, err
}

// supprimer un fichier
func DeleteFile(db bun.IDB, fileID int64, userID string) error {
	ctx := context.Background()

	// Delete associated share links
	_, err := db.NewDelete().Model((*ShareLink)(nil)).
		Where("resource_type = ? AND resource_id = ?", "file", fileID).
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = db.NewDelete().Model((*File)(nil)).Where(queryIDAndUserID, fileID, userID).Exec(ctx)
	return err
}

func DeleteFolder(db bun.IDB, folderID int64, userID string) error {
	ctx := context.Background()

	// Delete associated share links
	_, err := db.NewDelete().Model((*ShareLink)(nil)).
		Where("resource_type = ? AND resource_id = ?", "folder", folderID).
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = db.NewDelete().Model((*Folder)(nil)).Where(queryIDAndUserID, folderID, userID).Exec(ctx)
	return err
}

// Trouver un fichier par son ID
func GetFile(db *bun.DB, fileID int64, userID string) (*File, error) {
	ctx := context.Background()
	var file File
	err := db.NewSelect().Model(&file).Where(queryIDAndUserID, fileID, userID).Scan(ctx)
	return &file, err
}

// Trouver un dossier par son ID
func GetFolder(db *bun.DB, folderID int64, userID string) (*Folder, error) {
	ctx := context.Background()
	var folder Folder
	err := db.NewSelect().Model(&folder).Where(queryIDAndUserID, folderID, userID).Scan(ctx)
	return &folder, err
}
