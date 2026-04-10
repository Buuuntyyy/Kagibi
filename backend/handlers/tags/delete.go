// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package tags

import (
	"kagibi/backend/pkg"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func DeleteTagHandler(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDInterface, _ := c.Get("user_id")
		userID, _ := userIDInterface.(string)

		tagID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tag ID"})
			return
		}

		_, err = db.NewDelete().Model((*pkg.Tag)(nil)).Where("id = ? AND user_id = ?", tagID, userID).Exec(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete tag"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Tag deleted"})
	}
}
