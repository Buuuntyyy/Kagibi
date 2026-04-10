// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package users

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// SecuritySettings represents user security preferences
type SecuritySettings struct {
	bun.BaseModel                  `bun:"table:user_security_settings"`
	UserID                         string `bun:"user_id,pk" json:"user_id"`
	MFAEnabled                     bool   `bun:"mfa_enabled" json:"mfa_enabled"`
	MFAVerified                    bool   `bun:"mfa_verified" json:"mfa_verified"`
	RequireMFAOnLogin              bool   `bun:"require_mfa_on_login" json:"require_mfa_on_login"`
	RequireMFAOnDestructiveActions bool   `bun:"require_mfa_on_destructive_actions" json:"require_mfa_on_destructive_actions"`
	RequireMFAOnDownloads          bool   `bun:"require_mfa_on_downloads" json:"require_mfa_on_downloads"`
}

// GetSecuritySettingsHandler retrieves user security settings
func GetSecuritySettingsHandler(c *gin.Context, db *bun.DB) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var settings SecuritySettings
	err := db.NewSelect().
		Model(&settings).
		Where("user_id = ?", userID).
		Scan(c.Request.Context())

	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("[security-settings] select failed for user_id=%s: %v", userID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch settings", "details": err.Error()})
			return
		}

		// If settings don't exist, return defaults
		settings = SecuritySettings{
			UserID:                         userID,
			MFAEnabled:                     false,
			MFAVerified:                    false,
			RequireMFAOnLogin:              false,
			RequireMFAOnDestructiveActions: false,
			RequireMFAOnDownloads:          false,
		}
	}

	c.JSON(http.StatusOK, settings)
}

// UpdateSecuritySettingsRequest represents the request body
type UpdateSecuritySettingsRequest struct {
	MFAEnabled                     *bool `json:"mfa_enabled"`
	MFAVerified                    *bool `json:"mfa_verified"`
	RequireMFAOnLogin              *bool `json:"require_mfa_on_login"`
	RequireMFAOnDestructiveActions *bool `json:"require_mfa_on_destructive_actions"`
	RequireMFAOnDownloads          *bool `json:"require_mfa_on_downloads"`
}

// UpdateSecuritySettingsHandler updates user security settings
func UpdateSecuritySettingsHandler(c *gin.Context, db *bun.DB) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req UpdateSecuritySettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Fetch existing settings
	var settings SecuritySettings
	err := db.NewSelect().
		Model(&settings).
		Where("user_id = ?", userID).
		Scan(c.Request.Context())

	// If settings don't exist, create them
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("[security-settings] select failed for user_id=%s: %v", userID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch settings", "details": err.Error()})
			return
		}

		settings = SecuritySettings{
			UserID: userID,
		}
	}

	// Update only provided fields
	if req.MFAEnabled != nil {
		settings.MFAEnabled = *req.MFAEnabled
	}
	if req.MFAVerified != nil {
		settings.MFAVerified = *req.MFAVerified
	}
	if req.RequireMFAOnLogin != nil {
		settings.RequireMFAOnLogin = *req.RequireMFAOnLogin
	}
	if req.RequireMFAOnDestructiveActions != nil {
		settings.RequireMFAOnDestructiveActions = *req.RequireMFAOnDestructiveActions
	}
	if req.RequireMFAOnDownloads != nil {
		settings.RequireMFAOnDownloads = *req.RequireMFAOnDownloads
	}

	// Upsert settings
	_, err = db.NewInsert().
		Model(&settings).
		On("CONFLICT (user_id) DO UPDATE").
		Set("mfa_enabled = EXCLUDED.mfa_enabled").
		Set("mfa_verified = EXCLUDED.mfa_verified").
		Set("require_mfa_on_login = EXCLUDED.require_mfa_on_login").
		Set("require_mfa_on_destructive_actions = EXCLUDED.require_mfa_on_destructive_actions").
		Set("require_mfa_on_downloads = EXCLUDED.require_mfa_on_downloads").
		Exec(c.Request.Context())

	if err != nil {
		log.Printf("[security-settings] upsert failed for user_id=%s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}
