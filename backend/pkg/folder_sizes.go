// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package pkg

import (
	"context"
	"database/sql"
	"errors"
	"path"
	"strings"
	"time"

	"github.com/uptrace/bun"
)

func BuildFolderAncestorPaths(folderPath string) []string {
	clean := path.Clean(strings.ReplaceAll(folderPath, "\\", "/"))
	if clean == "." || clean == "/" {
		return nil
	}
	if !strings.HasPrefix(clean, "/") {
		clean = "/" + clean
	}

	parts := strings.Split(strings.TrimPrefix(clean, "/"), "/")
	paths := make([]string, 0, len(parts))
	current := ""
	for _, p := range parts {
		if p == "" {
			continue
		}
		current += "/" + p
		paths = append(paths, current)
	}
	return paths
}

func UpdateFolderSizesForFile(ctx context.Context, db *bun.DB, userID, filePath string, delta int64) error {
	parentPath := path.Dir(filePath)
	return UpdateFolderSizesForFolderPath(ctx, db, userID, parentPath, delta)
}

func UpdateFolderSizesForFolderPath(ctx context.Context, db *bun.DB, userID, folderPath string, delta int64) error {
	if delta == 0 {
		return nil
	}
	folderPaths := BuildFolderAncestorPaths(folderPath)
	return UpdateFolderSizesForPaths(ctx, db, userID, folderPaths, delta)
}

func UpdateFolderSizesForPaths(ctx context.Context, db *bun.DB, userID string, folderPaths []string, delta int64) error {
	if delta == 0 || len(folderPaths) == 0 {
		return nil
	}

	var folders []Folder
	if err := db.NewSelect().Model(&folders).
		Column("id", "path").
		Where("user_id = ?", userID).
		Where("path IN (?)", bun.In(folderPaths)).
		Scan(ctx); err != nil {
		return err
	}

	for _, f := range folders {
		fs := &FolderSize{
			FolderID:  f.ID,
			UserID:    userID,
			SizeBytes: delta,
			UpdatedAt: time.Now(),
		}
		_, err := db.NewInsert().Model(fs).
			Column("folder_id", "user_id", "size_bytes", "updated_at").
			On("CONFLICT (folder_id) DO UPDATE").
			Set("size_bytes = GREATEST(fs.size_bytes + EXCLUDED.size_bytes, 0)").
			Set("updated_at = EXCLUDED.updated_at").
			Exec(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func GetFolderSize(ctx context.Context, db *bun.DB, folderID int64) (int64, error) {
	var fs FolderSize
	err := db.NewSelect().Model(&fs).
		Column("size_bytes").
		Where("folder_id = ?", folderID).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return fs.SizeBytes, nil
}

func DeleteFolderSize(ctx context.Context, db *bun.DB, folderID int64) error {
	_, err := db.NewDelete().Model((*FolderSize)(nil)).Where("folder_id = ?", folderID).Exec(ctx)
	return err
}

// RebuildFolderSizes recompute sizes for all folders based on current files.
func RebuildFolderSizes(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `
INSERT INTO folder_sizes (folder_id, user_id, size_bytes, updated_at)
SELECT f.id AS folder_id,
       f.user_id,
       COALESCE(SUM(fl.size), 0) AS size_bytes,
       NOW() AS updated_at
FROM folders AS f
LEFT JOIN files AS fl
  ON fl.user_id = f.user_id
 AND fl.is_preview = false
 AND fl.path LIKE (f.path || '/%')
GROUP BY f.id, f.user_id
ON CONFLICT (folder_id) DO UPDATE
SET size_bytes = EXCLUDED.size_bytes,
    updated_at = EXCLUDED.updated_at;
`)
	return err
}

// EnsureFolderSizesInitialized rebuilds sizes if the table is empty.
func EnsureFolderSizesInitialized(ctx context.Context, db *bun.DB) error {
	var count int
	if err := db.NewSelect().Model((*FolderSize)(nil)).ColumnExpr("COUNT(*)").Scan(ctx, &count); err != nil {
		return err
	}
	if count == 0 {
		return RebuildFolderSizes(ctx, db)
	}
	return nil
}
