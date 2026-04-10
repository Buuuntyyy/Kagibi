// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package tags

import (
	"kagibi/backend/pkg"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func CreateTagHandler(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDInterface, _ := c.Get("user_id")
		userID, _ := userIDInterface.(string)

		var input struct {
			Name  string `json:"name"`
			Color string `json:"color"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Check if tag already exists for this user
		exists, err := db.NewSelect().Model((*pkg.Tag)(nil)).
			Where("user_id = ? AND name = ?", userID, input.Name).
			Exists(c.Request.Context())

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "Tag already exists"})
			return
		}

		tag := &pkg.Tag{
			UserID: userID,
			Name:   input.Name,
			Color:  input.Color,
		}

		_, err = db.NewInsert().Model(tag).Exec(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tag"})
			return
		}

		c.JSON(http.StatusOK, tag)
	}
}
