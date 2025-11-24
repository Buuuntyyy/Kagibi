// internal/database.go
package pkg

import (
	"context"
	"database/sql"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func NewDB() *bun.DB {
	// Remplace les valeurs par celles de ton docker-compose ou de ta configuration locale
	dsn := "postgresql://user:password@localhost:5432/mydb?sslmode=disable"

	// Ouvre la connexion SQL
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

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

// Lister les fichier d'un utilisateur
func ListItemsByUser(db *bun.DB, userID int64, path string) ([]File, []Folder, error) {
	ctx := context.Background()
	var files []File
	var folders []Folder
	err := db.NewSelect().Model(&files).Where("user_id = ? AND path = ?", userID, path).Scan(ctx)
	if err != nil {
		return nil, nil, err
	}
	err = db.NewSelect().Model(&folders).Where("user_id = ? and path = ?", userID, path).Scan(ctx)
	if err != nil {
		return nil, nil, err
	}

	return files, folders, err
}

// supprimer un fichier
func DeleteFile(db *bun.DB, fileID int64, userID int64) error {
	ctx := context.Background()
	_, err := db.NewDelete().Model((*File)(nil)).Where("id = ? AND user_id = ?", fileID, userID).Exec(ctx)
	return err
}

func DeleteFolder(db *bun.DB, folderID int64, userID int64) error {
	ctx := context.Background()
	_, err := db.NewDelete().Model((*Folder)(nil)).Where("id = ? AND user_id = ?", folderID, userID).Exec(ctx)
	return err
}

// Trouver un fichier par son ID
func GetFile(db *bun.DB, fileID int64, userID int64) (*File, error) {
	ctx := context.Background()
	var file File
	err := db.NewSelect().Model(&file).Where("id = ? AND user_id = ?", fileID, userID).Scan(ctx)
	return &file, err
}
