// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package workers

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"kagibi/backend/pkg"
	"kagibi/backend/pkg/s3storage"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/uptrace/bun"
)

const whereUserID = "user_id = ?"

// StartAccountCleanupWorker démarre le worker de maintenance et nettoyage (RGPD)
// Nettoie les données orphelines et les comptes dont la suppression immédiate a échoué
// Exécute quotidiennement pour garantir l'intégrité des données
func StartAccountCleanupWorker(db *bun.DB) {
	go func() {
		// Exécuter une première fois au démarrage (après 5 minutes pour laisser le temps à l'app de démarrer)
		time.Sleep(5 * time.Minute)
		cleanupOrphansAndFailedDeletions(db)

		// Puis exécuter quotidiennement
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			cleanupOrphansAndFailedDeletions(db)
		}
	}()

	log.Println("[RGPD] Maintenance Worker started (orphan cleanup + failed deletions recovery)")
}

func cleanupOrphansAndFailedDeletions(db *bun.DB) {
	ctx := context.Background()

	log.Println("[RGPD] Starting maintenance: checking for orphans and failed deletions...")

	// 1. SÉCURITÉ : Vérifier les comptes avec deleted_at (ne devrait pas exister avec hard delete)
	//    Si trouvés = suppression immédiate a échoué, les nettoyer maintenant
	var users []pkg.User
	err := db.NewSelect().
		Model(&users).
		Where("deleted_at IS NOT NULL").
		Scan(ctx)

	if err != nil {
		log.Printf("[RGPD] Failed to check for accounts with deleted_at flag: %v", err)
	} else if len(users) > 0 {
		log.Printf("[RGPD] Found %d accounts with deleted_at flag (immediate deletion failed)", len(users))
		log.Println("[RGPD] Performing emergency hard delete now...")

		for _, user := range users {
			cleanupUserData(ctx, db, user)
		}
	}

	// 2. Nettoyer les fichiers DB orphelins (fichiers en DB sans utilisateur valide)
	cleanupOrphanFiles(ctx, db)

	// 3. TODO : Nettoyer les fichiers S3 orphelins (fichiers S3 sans entrée DB)
	// Note : Nécessite de scanner tout S3, opération coûteuse - à implémenter si nécessaire

	log.Println("[RGPD] Maintenance completed")
}

// deleteUserS3Files deletes all S3 objects for the user's files and then the entire user prefix.
func deleteUserS3Files(ctx context.Context, userID string, files []pkg.File) {
	for _, file := range files {
		if s3storage.Client == nil || file.Path == "" {
			continue
		}
		s3Key := fmt.Sprintf("users/%s%s", userID, file.Path)
		if _, err := s3storage.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(s3storage.BucketName),
			Key:    aws.String(s3Key),
		}); err != nil {
			log.Printf("[RGPD] Failed to delete S3 file %s: %v", s3Key, err)
		}
	}
	cleanupS3Prefix(ctx, userID)
}

// deleteUserDBRecords hard-deletes all DB rows belonging to a user.
func deleteUserDBRecords(ctx context.Context, db *bun.DB, user pkg.User) {
	userID := user.ID
	models := []struct {
		model   interface{}
		name    string
		extraFn func() error
	}{
		{(*pkg.File)(nil), "files", nil},
		{(*pkg.Folder)(nil), "folders", nil},
		{(*pkg.Tag)(nil), "tags", nil},
		{(*pkg.RecentActivity)(nil), "recent activities", nil},
	}
	for _, m := range models {
		if _, err := db.NewDelete().Model(m.model).Where(whereUserID, userID).Exec(ctx); err != nil {
			log.Printf("[RGPD] Failed to delete %s for user %s: %v", m.name, userID, err)
		}
	}
	if _, err := db.NewDelete().Model(&user).Where("id = ?", userID).Exec(ctx); err != nil {
		log.Printf("[RGPD] Failed to hard delete user %s: %v", userID, err)
	} else {
		log.Printf("[RGPD] Hard deleted account: %s", userID)
	}
}

// cleanupUserData effectue le hard delete complet d'un utilisateur et de toutes ses données
// Utilisé comme filet de sécurité si la suppression immédiate a échoué
func cleanupUserData(ctx context.Context, db *bun.DB, user pkg.User) {
	userID := user.ID
	log.Printf("[RGPD] Emergency cleanup for user: %s", userID)

	var files []pkg.File
	if err := db.NewSelect().Model(&files).Where(whereUserID, userID).Scan(ctx); err != nil {
		log.Printf("[RGPD] Failed to fetch files for user %s: %v", userID, err)
	} else {
		deleteUserS3Files(ctx, userID, files)
	}

	deleteUserDBRecords(ctx, db, user)
}

// cleanupOrphanFiles supprime les fichiers en DB dont l'utilisateur n'existe plus
func cleanupOrphanFiles(ctx context.Context, db *bun.DB) {
	// Compter les fichiers orphelins
	count, err := db.NewSelect().
		Model((*pkg.File)(nil)).
		Where("user_id NOT IN (SELECT id FROM profiles)").
		Count(ctx)

	if err != nil {
		log.Printf("[RGPD] Failed to count orphan files: %v", err)
		return
	}

	if count == 0 {
		log.Println("[RGPD] No orphan files found")
		return
	}

	log.Printf("[RGPD] Found %d orphan files (user deleted but files remain)", count)

	// Supprimer les fichiers orphelins
	result, err := db.NewDelete().
		Model((*pkg.File)(nil)).
		Where("user_id NOT IN (SELECT id FROM profiles)").
		Exec(ctx)

	if err != nil {
		log.Printf("[RGPD] Failed to delete orphan files: %v", err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("[RGPD] Deleted %d orphan files from database", rowsAffected)
}

func cleanupS3Prefix(ctx context.Context, userID string) {
	if s3storage.Client == nil {
		return
	}

	prefix := fmt.Sprintf("users/%s/", userID)

	// Lister tous les objets avec ce préfixe
	paginator := s3.NewListObjectsV2Paginator(s3storage.Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s3storage.BucketName),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			log.Printf("[RGPD] Failed to list S3 objects for prefix %s: %v", prefix, err)
			break
		}

		for _, obj := range page.Contents {
			_, err := s3storage.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
				Bucket: aws.String(s3storage.BucketName),
				Key:    obj.Key,
			})
			if err != nil && !strings.Contains(err.Error(), "NoSuchKey") {
				log.Printf("[RGPD] Failed to delete orphan S3 object %s: %v", *obj.Key, err)
			}
		}
	}
}
