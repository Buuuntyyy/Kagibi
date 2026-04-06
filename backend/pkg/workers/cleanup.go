package workers

import (
	"context"
	"log"
	"time"

	"kagibi/backend/pkg"

	"github.com/uptrace/bun"
)

// StartCleanupWorker lance une routine en arrière-plan pour nettoyer les données expirées
func StartCleanupWorker(db *bun.DB) {
	// Exécuter immédiatement au démarrage
	go func() {
		log.Println("Running initial cleanup...")
		cleanupExpiredShares(db)
	}()

	go func() {
		// Vérifier toutes les minutes (plus réactif pour l'utilisateur)
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				cleanupExpiredShares(db)
			}
		}
	}()
	log.Println("Cleanup Worker started (interval: 1m)")
}

func cleanupExpiredShares(db *bun.DB) {
	ctx := context.Background()

	// Supprimer les liens de partage dont la date d'expiration est passée
	res, err := db.NewDelete().
		Model((*pkg.ShareLink)(nil)).
		Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).
		Exec(ctx)

	if err != nil {
		log.Printf("Error cleaning up expired shares: %v", err)
		return
	}

	count, _ := res.RowsAffected()
	if count > 0 {
		log.Printf("Cleanup: Removed %d expired share links", count)
	}
}
