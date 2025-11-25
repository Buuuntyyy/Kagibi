package files

import (
	"net/http"
	"os"
	"path/filepath"
	"safercloud/backend/pkg"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// Structure pour recevoir les IDs depuis le corps de la requête JSON
type BulkDeleteRequest struct {
    FileIDs []int64 `json:"file_ids" binding:"required"`
}

func BulkDeleteHandler(c *gin.Context, db *bun.DB) {
    var req BulkDeleteRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Liste d'IDs invalide"})
        return
    }

    if len(req.FileIDs) == 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Aucun fichier à supprimer"})
        return
    }

    userID, _ := c.Get("user_id")
    userIDInt, _ := strconv.ParseInt(userID.(string), 10, 64)

    // Débuter une transaction pour s'assurer que tout réussit ou tout échoue
    tx, err := db.Begin()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de démarrer la transaction"})
        return
    }
    defer tx.Rollback() // Rollback si quelque chose se passe mal

    var filesToDelete []pkg.File
    // Récupérer les infos des fichiers pour obtenir leurs chemins
    err = tx.NewSelect().Model(&filesToDelete).
        Where("id IN (?)", bun.In(req.FileIDs)).
        Where("user_id = ?", userIDInt).
        Scan(c)

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de récupérer les fichiers"})
        return
    }

    if len(filesToDelete) == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Aucun fichier trouvé à supprimer"})
        return
    }

    // Supprimer les fichiers du système de fichiers
    for _, file := range filesToDelete {
        cleanPath := strings.TrimPrefix(file.Path, "/")
        filePath := filepath.Join("uploads", strconv.FormatInt(userIDInt, 10), cleanPath)
        os.Remove(filePath) // On ignore l'erreur si le fichier n'existe déjà plus
    }

    // Supprimer les enregistrements de la base de données
    _, err = tx.NewDelete().Model((*pkg.File)(nil)).
        Where("id IN (?)", bun.In(req.FileIDs)).
        Where("user_id = ?", userIDInt).
        Exec(c)

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de supprimer les fichiers de la base de données"})
        return
    }

    // Si tout s'est bien passé, on commit la transaction
    if err := tx.Commit(); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de finaliser la suppression"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Fichiers supprimés avec succès"})
}