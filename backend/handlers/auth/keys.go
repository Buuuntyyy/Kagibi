package auth

import (
	"kagibi/backend/pkg"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func GetUserKeys(c *gin.Context, db *bun.DB) {
	userID := c.GetString("user_id") // Récupéré du JWT

	user, err := pkg.FindUserByID(db, userID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Profil introuvable"})
		return
	}

	// On renvoie UNIQUEMENT ce qui sert à déchiffrer la MasterKey
	c.JSON(200, gin.H{
		"salt":                  user.Salt,
		"encrypted_master_key":  user.EncryptedMasterKey,
		"public_key":            user.PublicKey,
		"encrypted_private_key": user.EncryptedPrivateKey,
	})
}
