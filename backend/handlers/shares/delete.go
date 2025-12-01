// backend/handlers/shares/delete.go
package shares

import (
	"net/http"

	"safercloud/backend/pkg"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// DeleteShareLinkHandler supprime un lien de partage
func DeleteShareLinkHandler(c *gin.Context, db *bun.DB) {
	userIDInterface, _ := c.Get("user_id")
	userID := userIDInterface.(string)

	shareID := c.Param("shareID")

	// Vérifier que le lien de partage appartient bien à l'utilisateur
	res, err := db.NewDelete().Model((*pkg.ShareLink)(nil)).
		Where("id = ? AND owner_id = ?", shareID, userID).
		Exec(c.Request.Context())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la suppression du lien"})
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la vérification de la suppression"})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lien de partage non trouvé ou vous n'avez pas la permission de le supprimer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lien de partage supprimé avec succès"})
}
