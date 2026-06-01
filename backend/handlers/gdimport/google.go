// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package gdimport

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// GetGoogleConfig returns the Google OAuth client ID to the authenticated frontend.
// The client ID is a public identifier — Google explicitly designed it to be exposed in
// browser-side code. It is NOT a secret and must not be treated as one.
//
// Production note: the drive.readonly scope used by the frontend requires Google app
// verification (OAuth consent screen review). Until verified, only test users added in
// Google Cloud Console can authenticate. See:
// https://developers.google.com/identity/protocols/oauth2/production-readiness/oauth-app-verification
func GetGoogleConfig(c *gin.Context) {
	clientID := os.Getenv("GOOGLE_OAUTH_CLIENT_ID")
	if clientID == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "google_import_not_configured",
			"message": "Google Drive import is not enabled on this instance",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"client_id": clientID})
}
