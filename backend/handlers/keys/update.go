package keys

import (
	"net/http"
	"kagibi/backend/pkg"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type UpdateKeysRequest struct {
	PublicKey           string `json:"public_key"`
	EncryptedPrivateKey string `json:"encrypted_private_key"`
}

func UpdateKeysHandler(c *gin.Context, db *bun.DB) {
	var req UpdateKeysRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userID := c.GetString("user_id")

	_, err := db.NewUpdate().
		Model((*pkg.User)(nil)).
		Set("public_key = ?", req.PublicKey).
		Set("encrypted_private_key = ?", req.EncryptedPrivateKey).
		Where("id = ?", userID).
		Exec(c.Request.Context())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update keys"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Keys updated"})
}
