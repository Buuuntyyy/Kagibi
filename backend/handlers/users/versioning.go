// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package users

import (
	"net/http"

	"kagibi/backend/pkg"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type UpdateVersioningRequest struct {
	Enabled bool `json:"enabled"`
}

// GetVersioningHandler returns the current versioning preference for the authenticated user.
// GET /users/versioning
func GetVersioningHandler(c *gin.Context, db *bun.DB) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user, err := pkg.FindUserByID(db, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	plan, err := pkg.FindUserPlanByUserID(db, userID)
	if err != nil {
		plan = &pkg.UserPlan{}
	}

	c.JSON(http.StatusOK, gin.H{
		"versioning_enabled":    user.VersioningEnabled,
		"version_storage_bytes": plan.VersionStorageBytes,
		"max_versions":          pkg.GetMaxVersions(plan.Plan),
	})
}

// UpdateVersioningHandler toggles file versioning for the authenticated user.
// PUT /users/versioning
func UpdateVersioningHandler(c *gin.Context, db *bun.DB) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req UpdateVersioningRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if _, err := db.NewUpdate().Model((*pkg.User)(nil)).
		Set("versioning_enabled = ?", req.Enabled).
		Where("id = ?", userID).
		Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update versioning setting"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"versioning_enabled": req.Enabled})
}
