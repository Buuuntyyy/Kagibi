// internal/migrate.go
package pkg

import (
	"context"
	"fmt"
	"github.com/uptrace/bun"
)

func Migrate(db *bun.DB) error {
	ctx := context.Background()

	// Crée les tables si elles n'existent pas
	models := []interface{}{(*User)(nil), (*File)(nil)}

	for _, model := range models {
		_, err := db.NewSelect().Model(model).Exec(ctx)
		if err != nil {
			_, err = db.NewCreateTable().Model(model).IfNotExists().Exec(ctx)
			if err != nil {
				return fmt.Errorf("failed to create table: %w", err)
			}
			fmt.Printf("Table created: %T\n", model)
		}
	}

	return nil
}
