// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package dbimport

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// GetDropboxConfig returns the Dropbox App key to the authenticated frontend.
// This value is a public identifier — Dropbox's own JS SDK sends it directly from
// browser code for PKCE-based public clients. It is NOT a secret.
//
// Like OneDrive (and unlike Google), Dropbox supports PKCE for public clients with no
// client_secret: the browser and the desktop app each exchange the authorization code
// directly against Dropbox's token endpoint. There is no server-side token exchange
// endpoint here.
//
// DROPBOX_APP_KEY: OAuth app key (Dropbox App Console, "Scoped access" API)
func GetDropboxConfig(c *gin.Context) {
	appKey := os.Getenv("DROPBOX_APP_KEY")

	if appKey == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "dropbox_import_not_configured",
			"message": "Dropbox import is not enabled on this instance",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"app_key": appKey,
	})
}
