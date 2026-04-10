// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"kagibi/backend/pkg"
	"kagibi/backend/pkg/authprovider"
	"kagibi/backend/pkg/s3storage"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

const whereUserID = "user_id = ?"

type DeleteAccountRequest struct {
	Confirmation string `json:"confirmation" binding:"required"` // Must be "SUPPRIMER"
}

// DeleteAccount supprime le compte utilisateur (RGPD Article 17 - Droit à l'effacement)
// Stratégie: SUPPRESSION IMMÉDIATE ET COMPLÈTE - IRRÉVERSIBLE
func DeleteAccount(db *bun.DB, provider authprovider.AuthProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
			return
		}

		var req DeleteAccountRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Confirmation requise"})
			return
		}

		if req.Confirmation != "SUPPRIMER" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Veuillez confirmer en tapant 'SUPPRIMER'"})
			return
		}

		ctx := context.Background()

		var user pkg.User
		err := db.NewSelect().Model(&user).Where("id = ?", userID).Scan(ctx)
		if err != nil {
			log.Printf("[RGPD] User not found for deletion: %s", userID)
			c.JSON(http.StatusNotFound, gin.H{"error": "Compte non trouvé"})
			return
		}

		log.Printf("[RGPD] Starting IMMEDIATE deletion for user: %s (email: %s, provider: %s)", userID, user.Email, provider.Name())

		// Delete from auth provider (Supabase or PocketBase)
		if err := provider.DeleteUser(userID); err != nil {
			log.Printf("[RGPD] Failed to delete from auth provider (%s): %v (continuing with local deletion)", provider.Name(), err)
		} else {
			log.Printf("[RGPD] User deleted from auth provider (%s)", provider.Name())
		}

		// Delete S3 files asynchronously
		go deleteUserFilesFromS3(ctx, db, userID)

		// Hard delete all user data from PostgreSQL
		deleteUserData(ctx, db, userID, &user)

		log.Printf("[RGPD] COMPLETE DELETION: Account PERMANENTLY deleted: %s (email: %s)", userID, user.Email)
		c.JSON(http.StatusOK, gin.H{
			"success":      true,
			"message":      "Votre compte a été DÉFINITIVEMENT supprimé. Toutes vos données sont IRRÉCUPÉRABLES.",
			"deleted_at":   time.Now().Format(time.RFC3339),
			"irreversible": true,
			"warning":      "Vos fichiers, clés de chiffrement et toutes vos données ont été supprimés. Cette action est IRRÉVERSIBLE.",
		})
	}
}

func deleteUserData(ctx context.Context, db *bun.DB, userID string, user *pkg.User) {
	exec := func(label string, q *bun.DeleteQuery) {
		res, err := q.Exec(ctx)
		if err != nil {
			log.Printf("[RGPD] Failed to delete %s for user %s: %v", label, userID, err)
		} else if n, _ := res.RowsAffected(); n > 0 {
			log.Printf("[RGPD] Deleted %d row(s) from %s for user %s", n, label, userID)
		}
	}

	// 1. Realtime / P2P
	exec("RealtimeEvent", db.NewDelete().Model((*pkg.RealtimeEvent)(nil)).Where(whereUserID, userID))
	exec("P2PSignal", db.NewDelete().Model((*pkg.P2PSignal)(nil)).Where("sender_id = ? OR target_id = ?", userID, userID))

	// 2. Recent activity (references files/folders — must precede their deletion)
	exec("RecentActivity", db.NewDelete().Model((*pkg.RecentActivity)(nil)).Where(whereUserID, userID))

	// 3. Keys referencing user's folders/files (must precede File/Folder deletion)
	exec("FolderFolderKey", db.NewDelete().Model((*pkg.FolderFolderKey)(nil)).
		Where("parent_folder_id IN (SELECT id FROM folders WHERE user_id = ?) OR sub_folder_id IN (SELECT id FROM folders WHERE user_id = ?)", userID, userID))
	exec("FolderFileKey", db.NewDelete().Model((*pkg.FolderFileKey)(nil)).
		Where("folder_id IN (SELECT id FROM folders WHERE user_id = ?) OR file_id IN (SELECT id FROM files WHERE user_id = ?)", userID, userID))
	exec("ShareFileKey", db.NewDelete().Model((*pkg.ShareFileKey)(nil)).
		Where("share_id IN (SELECT id FROM share_links WHERE owner_id = ?) OR file_id IN (SELECT id FROM files WHERE user_id = ?)", userID, userID))

	// 4. Imported shares (user imported others' links, or user's links were imported by others)
	exec("ImportedShare", db.NewDelete().Model((*pkg.ImportedShare)(nil)).
		Where("user_id = ? OR share_link_id IN (SELECT id FROM share_links WHERE owner_id = ?)", userID, userID))

	// 5. File/folder shares (as recipient and as owner via owned files/folders)
	exec("FileShare", db.NewDelete().Model((*pkg.FileShare)(nil)).
		Where("shared_with_user_id = ? OR file_id IN (SELECT id FROM files WHERE user_id = ?)", userID, userID))
	exec("FolderShare", db.NewDelete().Model((*pkg.FolderShare)(nil)).
		Where("shared_with_user_id = ? OR folder_id IN (SELECT id FROM folders WHERE user_id = ?)", userID, userID))

	// 6. Share links owned by user
	exec("ShareLink", db.NewDelete().Model((*pkg.ShareLink)(nil)).Where("owner_id = ?", userID))

	// 7. Files, folders, and folder size cache
	exec("File", db.NewDelete().Model((*pkg.File)(nil)).Where(whereUserID, userID))
	exec("Folder", db.NewDelete().Model((*pkg.Folder)(nil)).Where(whereUserID, userID))
	exec("FolderSize", db.NewDelete().Model((*pkg.FolderSize)(nil)).Where(whereUserID, userID))

	// 8. Remaining user data
	exec("Tag", db.NewDelete().Model((*pkg.Tag)(nil)).Where(whereUserID, userID))
	exec("UserPlan", db.NewDelete().Model((*pkg.UserPlan)(nil)).Where(whereUserID, userID))
	exec("Friendship", db.NewDelete().Model((*pkg.Friendship)(nil)).Where("user_id_1 = ? OR user_id_2 = ?", userID, userID))

	// 9. Hard delete the profile (IRRÉVERSIBLE)
	if _, err := db.NewDelete().Model(user).Where("id = ?", userID).ForceDelete().Exec(ctx); err != nil {
		log.Printf("[RGPD] Failed to hard delete user %s: %v", userID, err)
	}
}

func deleteUserFilesFromS3(ctx context.Context, db *bun.DB, userID string) {
	if s3storage.Client == nil {
		log.Printf("[RGPD] S3 client not available, skipping S3 cleanup for user %s", userID)
		return
	}

	var files []pkg.File
	err := db.NewSelect().Model(&files).Where(whereUserID, userID).Scan(ctx)
	if err != nil {
		log.Printf("[RGPD] Failed to fetch files for S3 cleanup (user %s): %v", userID, err)
		return
	}

	successCount := 0
	for _, file := range files {
		if file.Path == "" {
			continue
		}
		s3Key := fmt.Sprintf("users/%s%s", userID, file.Path)
		_, err := s3storage.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(s3storage.BucketName),
			Key:    aws.String(s3Key),
		})
		if err == nil {
			successCount++
		}
	}

	// Clean up any orphaned files
	prefix := fmt.Sprintf("users/%s/", userID)
	paginator := s3.NewListObjectsV2Paginator(s3storage.Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s3storage.BucketName),
		Prefix: aws.String(prefix),
	})
	orphanCount := 0
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			break
		}
		for _, obj := range page.Contents {
			if _, err := s3storage.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
				Bucket: aws.String(s3storage.BucketName),
				Key:    obj.Key,
			}); err == nil {
				orphanCount++
			}
		}
	}

	log.Printf("[RGPD] S3 cleanup completed for user %s: %d files deleted, %d orphans removed", userID, successCount, orphanCount)
}
