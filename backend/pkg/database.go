// internal/database.go
package pkg

import (
	"database/sql"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func NewDB() *bun.DB {
	// Remplace les valeurs par celles de ton docker-compose ou de ta configuration locale
	dsn := "postgres://user:password@localhost:5432/mon_drive?sslmode=disable"

	// Ouvre la connexion SQL
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

	// Crée une instance Bun
	db := bun.NewDB(sqldb, pgdialect.New())

	return db
}
