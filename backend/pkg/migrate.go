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
	models := []interface{}{(*User)(nil), (*File)(nil), (*Folder)(nil)}

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

	_, err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_files_user_id ON files(user_id);
		CREATE INDEX IF NOT EXISTS idx_folders_user_id ON folders(user_id);
	`)
	return err
}
