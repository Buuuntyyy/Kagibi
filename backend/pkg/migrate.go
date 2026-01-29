// internal/migrate.go
package pkg

import (
	"context"
	"fmt"
	"log"

	"github.com/uptrace/bun"
)

func Migrate(db *bun.DB) error {
	ctx := context.Background()

	// Crée les tables si elles n'existent pas
	models := []interface{}{
		(*User)(nil),
		(*File)(nil),
		(*Folder)(nil),
		(*Tag)(nil),
		(*ShareLink)(nil),
		(*ShareFileKey)(nil),
		(*ImportedShare)(nil),
		(*FileShare)(nil),
		(*FolderShare)(nil),
		(*FolderFileKey)(nil),
		(*FolderFolderKey)(nil),
		(*RecentActivity)(nil),
	}

	for _, model := range models {
		// Try to create table if not exists
		_, err := db.NewCreateTable().Model(model).IfNotExists().Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	// Manually add the 'path' column to 'share_links' if it doesn't exist
	_, err := db.ExecContext(ctx, `ALTER TABLE "share_links" ADD COLUMN IF NOT EXISTS "path" VARCHAR NOT NULL DEFAULT '';`)
	if err != nil {
		return fmt.Errorf("failed to add path column to share_links: %w", err)
	}

	// Add encrypted_key to files
	_, err = db.ExecContext(ctx, `ALTER TABLE "files" ADD COLUMN IF NOT EXISTS "encrypted_key" VARCHAR;`)
	if err != nil {
		return fmt.Errorf("failed to add encrypted_key column to files: %w", err)
	}

	// Add encrypted_key to folders
	_, err = db.ExecContext(ctx, `ALTER TABLE "folders" ADD COLUMN IF NOT EXISTS "encrypted_key" VARCHAR;`)
	if err != nil {
		return fmt.Errorf("failed to add encrypted_key column to folders: %w", err)
	}

	// Add encrypted_key to share_links
	_, err = db.ExecContext(ctx, `ALTER TABLE "share_links" ADD COLUMN IF NOT EXISTS "encrypted_key" VARCHAR;`)
	if err != nil {
		return fmt.Errorf("failed to add encrypted_key column to share_links: %w", err)
	}

	// Add encrypted_key to folder_shares
	_, err = db.ExecContext(ctx, `ALTER TABLE "folder_shares" ADD COLUMN IF NOT EXISTS "encrypted_key" VARCHAR;`)
	if err != nil {
		return fmt.Errorf("failed to add encrypted_key column to folder_shares: %w", err)
	}

	// Folder sizes table (with FK to folders)
	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS "folder_sizes" (
		"folder_id" BIGINT PRIMARY KEY REFERENCES "folders"("id") ON DELETE CASCADE,
		"user_id" VARCHAR NOT NULL,
		"size_bytes" BIGINT NOT NULL DEFAULT 0,
		"updated_at" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`)
	if err != nil {
		return fmt.Errorf("failed to create folder_sizes table: %w", err)
	}

	// Update existing rows to have a default path if it's empty
	_, err = db.ExecContext(ctx, `UPDATE "share_links" SET "path" = '' WHERE "path" IS NULL;`)
	if err != nil {
		return fmt.Errorf("failed to update existing paths in share_links: %w", err)
	}

	// --- ADD INDICES FOR PERFORMANCE ---

	// Index for listing files (User + Path)
	_, err = db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_files_user_path ON files (user_id, path);`)
	if err != nil {
		log.Printf("Warning: failed to create idx_files_user_path: %v", err)
	}

	// Index for listing folders (User + Path)
	_, err = db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_folders_user_path ON folders (user_id, path);`)
	if err != nil {
		log.Printf("Warning: failed to create idx_folders_user_path: %v", err)
	}

	// Index for folder sizes (User)
	_, err = db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_folder_sizes_user ON folder_sizes (user_id);`)
	if err != nil {
		log.Printf("Warning: failed to create idx_folder_sizes_user: %v", err)
	}

	// Index for finding shared folders (SharedWithUserID)
	_, err = db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_folder_shares_user ON folder_shares (shared_with_user_id);`)
	if err != nil {
		log.Printf("Warning: failed to create idx_folder_shares_user: %v", err)
	}

	return nil
}
