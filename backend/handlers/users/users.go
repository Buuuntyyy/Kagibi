// backend/handlers/users.go
package users

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"safercloud/backend/pkg"
	"github.com/uptrace/bun"
)

func ListUsersHandler(c *gin.Context, db *bun.DB) {
	users, err := pkg.ListUsers(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

