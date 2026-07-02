// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package gdimport

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetGoogleConfig returns the Google OAuth client IDs to the authenticated frontend.
// Both values are public identifiers — Google explicitly designed them to be exposed in
// browser-side code. They are NOT secrets and must not be treated as ones.
//
// GOOGLE_OAUTH_CLIENT_ID        : OAuth client of type "Web application" (used by the browser frontend)
// GOOGLE_OAUTH_DESKTOP_CLIENT_ID: OAuth client of type "Desktop app" (used by the desktop app via PKCE loopback)
//
// Production note: the drive.readonly scope requires Google app verification before
// non-test users can authenticate. See:
// https://developers.google.com/identity/protocols/oauth2/production-readiness/oauth-app-verification
func GetGoogleConfig(c *gin.Context) {
	webClientID := os.Getenv("GOOGLE_OAUTH_CLIENT_ID")
	desktopClientID := os.Getenv("GOOGLE_OAUTH_DESKTOP_CLIENT_ID")

	if webClientID == "" && desktopClientID == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "google_import_not_configured",
			"message": "Google Drive import is not enabled on this instance",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"client_id":         webClientID,
		"desktop_client_id": desktopClientID,
	})
}

// ExchangeDesktopToken échange un code d'autorisation PKCE contre un access_token Google.
// Le client_secret reste côté serveur — il n'est jamais transmis au desktop app.
//
// Requiert GOOGLE_OAUTH_DESKTOP_CLIENT_ID + GOOGLE_OAUTH_DESKTOP_CLIENT_SECRET dans l'env.
// POST /import/google/desktop-token
// body: {code, code_verifier, redirect_uri}
func ExchangeDesktopToken(c *gin.Context) {
	clientID := os.Getenv("GOOGLE_OAUTH_DESKTOP_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_OAUTH_DESKTOP_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "google_desktop_not_configured",
			"message": "GOOGLE_OAUTH_DESKTOP_CLIENT_ID or GOOGLE_OAUTH_DESKTOP_CLIENT_SECRET not set",
		})
		return
	}

	var req struct {
		Code         string `json:"code" binding:"required"`
		CodeVerifier string `json:"code_verifier" binding:"required"`
		RedirectURI  string `json:"redirect_uri" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Appel au token endpoint Google
	resp, err := http.PostForm("https://oauth2.googleapis.com/token", url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {req.Code},
		"code_verifier": {req.CodeVerifier},
		"redirect_uri":  {req.RedirectURI},
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to reach Google token endpoint: " + err.Error()})
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "invalid response from Google"})
		return
	}

	// Ne retourner que l'access_token (pas le refresh_token ni autres champs sensibles)
	accessToken, _ := result["access_token"].(string)
	if errCode, ok := result["error"].(string); ok {
		desc, _ := result["error_description"].(string)
		if desc == "" {
			desc = errCode
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": desc})
		return
	}
	if accessToken == "" {
		c.JSON(http.StatusBadGateway, gin.H{"error": "no access_token in Google response"})
		return
	}

	// Supprimer les champs sensibles avant de retourner
	delete(result, "refresh_token")
	delete(result, "id_token")
	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
		"expires_in":   result["expires_in"],
	})
}

// stripSensitive is kept for future use
var _ = strings.TrimSpace
