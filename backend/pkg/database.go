// internal/database.go
package pkg

import (
	"database/sql"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"context"
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

func CreateUser(db *bun.DB, user *User) error {
	ctx := context.Background()
	_, err := db.NewInsert().Model(user).Exec(ctx)
	return err
}
