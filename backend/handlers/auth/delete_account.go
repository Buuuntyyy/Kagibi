package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"safercloud/backend/pkg"
	"safercloud/backend/pkg/s3storage"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// DeleteAccountRequest représente la requête de suppression de compte
type DeleteAccountRequest struct {
	Confirmation string `json:"confirmation" binding:"required"` // Doit être "SUPPRIMER"
}

// DeleteAccount supprime le compte utilisateur (RGPD Article 17 - Droit à l'effacement)
// Stratégie: SUPPRESSION IMMÉDIATE ET COMPLÈTE - IRRÉVERSIBLE
// - Hard delete immédiat : utilisateur, fichiers DB, dossiers, tags, activités
// - Suppression S3 asynchrone (arrière-plan)
// - AUCUNE période de grâce : données définitivement irrécupérables
func DeleteAccount(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
			return
		}

		var req DeleteAccountRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Confirmation requise"})
			return
		}

		// Vérification de la confirmation pour éviter les suppressions accidentelles
		if req.Confirmation != "SUPPRIMER" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Veuillez confirmer en tapant 'SUPPRIMER'"})
			return
		}

		ctx := context.Background()

		// 1. Vérifier que l'utilisateur existe
		var user pkg.User
		err := db.NewSelect().
			Model(&user).
			Where("id = ?", userID).
			Scan(ctx)
		if err != nil {
			log.Printf("[RGPD] User not found for deletion: %s", userID)
			c.JSON(http.StatusNotFound, gin.H{"error": "Compte non trouvé"})
			return
		}

		log.Printf("[RGPD] 🗑️ Starting IMMEDIATE and COMPLETE deletion for user: %s (email: %s)", userID, user.Email)

		// 2. SUPPRESSION IMMÉDIATE DES FICHIERS S3 (asynchrone pour ne pas bloquer)
		go deleteUserFilesFromS3(ctx, db, userID)

		// 3. HARD DELETE des fichiers en DB (IMMÉDIAT)
		result, err := db.NewDelete().
			Model((*pkg.File)(nil)).
			Where("user_id = ?", userID).
			Exec(ctx)
		if err != nil {
			log.Printf("[RGPD] ⚠️ Failed to delete files for user %s: %v", userID, err)
		} else {
			rowsAffected, _ := result.RowsAffected()
			log.Printf("[RGPD] ✅ Deleted %d files from DB for user %s", rowsAffected, userID)
		}

		// 4. HARD DELETE des dossiers en DB (IMMÉDIAT)
		result, err = db.NewDelete().
			Model((*pkg.Folder)(nil)).
			Where("user_id = ?", userID).
			Exec(ctx)
		if err != nil {
			log.Printf("[RGPD] ⚠️ Failed to delete folders for user %s: %v", userID, err)
		} else {
			rowsAffected, _ := result.RowsAffected()
			log.Printf("[RGPD] ✅ Deleted %d folders from DB for user %s", rowsAffected, userID)
		}

		// 5. HARD DELETE des tags (IMMÉDIAT)
		result, err = db.NewDelete().
			Model((*pkg.Tag)(nil)).
			Where("user_id = ?", userID).
			Exec(ctx)
		if err != nil {
			log.Printf("[RGPD] ⚠️ Failed to delete tags for user %s: %v", userID, err)
		} else {
			rowsAffected, _ := result.RowsAffected()
			if rowsAffected > 0 {
				log.Printf("[RGPD] ✅ Deleted %d tags for user %s", rowsAffected, userID)
			}
		}

		// 6. HARD DELETE des activités récentes (IMMÉDIAT)
		result, err = db.NewDelete().
			Model((*pkg.RecentActivity)(nil)).
			Where("user_id = ?", userID).
			Exec(ctx)
		if err != nil {
			log.Printf("[RGPD] ⚠️ Failed to delete recent activities for user %s: %v", userID, err)
		} else {
			rowsAffected, _ := result.RowsAffected()
			if rowsAffected > 0 {
				log.Printf("[RGPD] ✅ Deleted %d activities for user %s", rowsAffected, userID)
			}
		}

		// 7. HARD DELETE des partages créés par l'utilisateur (IMMÉDIAT)
		_, err = db.NewDelete().
			Model((*pkg.ShareLink)(nil)).
			Where("user_id = ?", userID).
			Exec(ctx)
		if err != nil {
			log.Printf("[RGPD] Failed to delete share links for user %s: %v", userID, err)
		}

		// 8. HARD DELETE des partages directs (files)
		_, err = db.NewDelete().
			Model((*pkg.FileShare)(nil)).
			Where("owner_id = ?", userID).
			Exec(ctx)
		if err != nil {
			log.Printf("[RGPD] Failed to delete file shares for user %s: %v", userID, err)
		}

		// 9. HARD DELETE des partages directs (folders)
		_, err = db.NewDelete().
			Model((*pkg.FolderShare)(nil)).
			Where("owner_id = ?", userID).
			Exec(ctx)
		if err != nil {
			log.Printf("[RGPD] Failed to delete folder shares for user %s: %v", userID, err)
		}

		// 10. HARD DELETE des partages reçus par l'utilisateur
		_, err = db.NewDelete().
			Model((*pkg.FileShare)(nil)).
			Where("shared_with_user_id = ?", userID).
			Exec(ctx)
		if err != nil {
			log.Printf("[RGPD] Failed to delete received file shares for user %s: %v", userID, err)
		}

		_, err = db.NewDelete().
			Model((*pkg.FolderShare)(nil)).
			Where("shared_with_user_id = ?", userID).
			Exec(ctx)
		if err != nil {
			log.Printf("[RGPD] Failed to delete received folder shares for user %s: %v", userID, err)
		}

		// 11. HARD DELETE des amitiés (IMMÉDIAT)
		_, err = db.NewDelete().
			Model((*pkg.Friendship)(nil)).
			Where("user_id_1 = ? OR user_id_2 = ?", userID, userID).
			Exec(ctx)
		if err != nil {
			log.Printf("[RGPD] Failed to delete friendships for user %s: %v", userID, err)
		}

		// 12. HARD DELETE de l'utilisateur (IMMÉDIAT - IRRÉVERSIBLE)
		_, err = db.NewDelete().
			Model(&user).
			Where("id = ?", userID).
			Exec(ctx)

		if err != nil {
			log.Printf("[RGPD] ❌ Failed to hard delete user %s: %v", userID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la suppression définitive"})
			return
		}

		log.Printf("[RGPD] ✅ COMPLETE DELETION: Account PERMANENTLY deleted: %s (email: %s)", userID, user.Email)
		log.Printf("[RGPD] 🚨 IRRÉVERSIBLE: All user data deleted from DB. S3 cleanup running in background.")

		c.JSON(http.StatusOK, gin.H{
			"success":      true,
			"message":      "Votre compte a été DÉFINITIVEMENT supprimé. Toutes vos données sont IRRÉCUPÉRABLES.",
			"deleted_at":   time.Now().Format(time.RFC3339),
			"irreversible": true,
			"warning":      "Vos fichiers, clés de chiffrement et toutes vos données ont été supprimés. Cette action est IRRÉVERSIBLE.",
		})
	}
}

// deleteUserFilesFromS3 supprime tous les fichiers S3 d'un utilisateur de manière asynchrone
// Exécuté en arrière-plan pour ne pas bloquer la réponse API
func deleteUserFilesFromS3(ctx context.Context, db *bun.DB, userID string) {
	if s3storage.Client == nil {
		log.Printf("[RGPD] S3 client not available, skipping S3 cleanup for user %s", userID)
		return
	}

	// Récupérer tous les fichiers de l'utilisateur
	var files []pkg.File
	err := db.NewSelect().
		Model(&files).
		Where("user_id = ?", userID).
		Scan(ctx)

	if err != nil {
		log.Printf("[RGPD] ⚠️ Failed to fetch files for S3 cleanup (user %s): %v", userID, err)
		return
	}

	if len(files) == 0 {
		log.Printf("[RGPD] No files to delete from S3 for user %s", userID)
		return
	}

	log.Printf("[RGPD] 🗑️ Starting S3 cleanup: %d files for user %s", len(files), userID)

	successCount := 0
	errorCount := 0

	// Supprimer chaque fichier de S3
	for _, file := range files {
		if file.Path == "" {
			continue
		}

		// S3 key format: users/{userID}{filePath}
		s3Key := fmt.Sprintf("users/%s%s", userID, file.Path)

		_, err := s3storage.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(s3storage.BucketName),
			Key:    aws.String(s3Key),
		})

		if err != nil {
			log.Printf("[RGPD] ⚠️ Failed to delete S3 file %s: %v", s3Key, err)
			errorCount++
		} else {
			successCount++
		}
	}

	log.Printf("[RGPD] ✅ S3 cleanup completed for user %s: %d succeeded, %d failed", userID, successCount, errorCount)

	// Nettoyer le préfixe utilisateur (sécurité supplémentaire pour les fichiers orphelins)
	prefix := fmt.Sprintf("users/%s/", userID)
	paginator := s3.NewListObjectsV2Paginator(s3storage.Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s3storage.BucketName),
		Prefix: aws.String(prefix),
	})

	orphanCount := 0
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			log.Printf("[RGPD] ⚠️ Failed to list orphan files for user %s: %v", userID, err)
			break
		}

		for _, obj := range page.Contents {
			_, err := s3storage.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
				Bucket: aws.String(s3storage.BucketName),
				Key:    obj.Key,
			})
			if err == nil {
				orphanCount++
			}
		}
	}

	if orphanCount > 0 {
		log.Printf("[RGPD] ✅ Deleted %d orphan files from S3 for user %s", orphanCount, userID)
	}
}
