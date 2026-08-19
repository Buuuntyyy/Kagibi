// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package odimport

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// GetOneDriveConfig returns the Microsoft OAuth client ID to the authenticated frontend.
// This value is a public identifier — Microsoft explicitly designed it to be exposed in
// browser-side and native-app code for public clients using PKCE. It is NOT a secret.
//
// Unlike Google (which splits OAuth clients into "Web application" and "Desktop app"
// types, the latter requiring a client_secret), a single Azure AD App Registration can
// carry both a "Single-page application" platform (used by the browser) and a "Mobile
// and desktop applications" platform (used by the desktop app's loopback PKCE flow)
// under one client ID. Both platforms support PKCE for public clients with no secret,
// so there is no server-side token exchange endpoint here — the browser and the desktop
// app each exchange the authorization code directly against Microsoft's token endpoint.
//
// MICROSOFT_OAUTH_CLIENT_ID: OAuth client (Azure AD App Registration, public client)
func GetOneDriveConfig(c *gin.Context) {
	clientID := os.Getenv("MICROSOFT_OAUTH_CLIENT_ID")

	if clientID == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "onedrive_import_not_configured",
			"message": "OneDrive import is not enabled on this instance",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"client_id": clientID,
	})
}
