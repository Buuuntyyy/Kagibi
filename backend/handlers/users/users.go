// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

// backend/handlers/users.go
package users

import (
	"net/http"

	"kagibi/backend/pkg"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func ListUsersHandler(c *gin.Context, db *bun.DB) {
	users, err := pkg.ListUsers(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des utilisateurs"})
		return
	}

	var safeUsers []gin.H
	for _, u := range users {
		safeUsers = append(safeUsers, gin.H{
			"id":    u.ID,
			"name":  u.Name,
			"email": u.Email,
		})
	}

	c.JSON(http.StatusOK, safeUsers)
}
