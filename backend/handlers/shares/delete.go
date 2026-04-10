// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

// backend/handlers/shares/delete.go
package shares

import (
	"fmt"
	"kagibi/backend/pkg"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// DeleteShareLinkHandler supprime un lien de partage
func DeleteShareLinkHandler(c *gin.Context, db *bun.DB) {
	userIDInterface, _ := c.Get("user_id")
	userID := userIDInterface.(string)

	shareIDStr := c.Param("shareID")
	shareIDStr = strings.ReplaceAll(strings.ReplaceAll(shareIDStr, "\n", "_"), "\r", "_")

	// Convertir shareID en int64 pour être sûr
	// (Postgres gère généralement bien les chaînes pour les entiers, mais soyons explicites)
	// et ajoutons des logs pour le débogage.

	fmt.Printf("DEBUG: DeleteShareLinkHandler - ShareID: %s, UserID: %s\n", shareIDStr, userID)

	// Utilisation directe de la chaîne si conversion échoue ou conversion explicite ?
	// Conversion explicite est plus propre.
	// shareID, err := strconv.ParseInt(shareIDStr, 10, 64)

	// Si on utilise directement shareIDStr dans la requête, Bun/PG le gère.
	// Essayons de voir si on trouve le lien AVANT de supprimer, pour diagnotiquer le 404.

	var link pkg.ShareLink
	count, err := db.NewSelect().Model(&link).
		Where("id = ?", shareIDStr).
		Count(c.Request.Context())

	if err != nil {
		fmt.Printf("DEBUG: Error checking if link exists: %v\n", err)
	} else {
		fmt.Printf("DEBUG: Link found by ID only: %d\n", count)
	}

	// Vérifier que le lien de partage appartient bien à l'utilisateur
	res, err := db.NewDelete().Model((*pkg.ShareLink)(nil)).
		Where("id = ? AND owner_id = ?", shareIDStr, userID).
		Exec(c.Request.Context())

	if err != nil {
		fmt.Printf("DEBUG: Error checking deletion: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la suppression du lien"})
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la vérification de la suppression"})
		return
	}

	fmt.Printf("DEBUG: DeleteShareLinkHandler - RowsDeleted: %d\n", rowsAffected)

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lien de partage INTROUVABLE (Public Link Not Found) ou permission refusée"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lien de partage supprimé avec succès"})
}
