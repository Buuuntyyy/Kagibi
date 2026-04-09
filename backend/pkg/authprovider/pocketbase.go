package authprovider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	headerContentType  = "Content-Type"
	headerAppJSON      = "application/json"
	headerBearerPrefix = "Bearer "
)

// PocketBaseProvider validates PocketBase HS256 JWTs and uses the PocketBase Admin API
// for user management operations.
//
// PocketBase tokens use the "id" claim (not "sub") for the user identifier.
// The JWT secret must match the recordAuthToken.secret configured in PocketBase settings.
//
// To configure PocketBase to use a specific JWT secret:
//  1. Start PocketBase
//  2. The backend will call SetupJWTSecret() at startup if POCKETBASE_JWT_SECRET is set
//  3. Or set it manually in the PocketBase admin UI: Settings → Application → Auth token signing key
type PocketBaseProvider struct {
	url            string
	adminEmail     string
	adminPassword  string
	secret         []byte
	collectionName string // default: "users"
}

func NewPocketBaseProvider() *PocketBaseProvider {
	collection := os.Getenv("POCKETBASE_COLLECTION")
	if collection == "" {
		collection = "users"
	}
	return &PocketBaseProvider{
		url:            os.Getenv("POCKETBASE_URL"),
		adminEmail:     os.Getenv("POCKETBASE_ADMIN_EMAIL"),
		adminPassword:  os.Getenv("POCKETBASE_ADMIN_PASSWORD"),
		secret:         []byte(os.Getenv("POCKETBASE_JWT_SECRET")),
		collectionName: collection,
	}
}

func (p *PocketBaseProvider) Name() string           { return "pocketbase" }
func (p *PocketBaseProvider) GetUserIDClaim() string { return "id" }
func (p *PocketBaseProvider) GetJWTSecret() []byte   { return p.secret }

// getAdminToken authenticates with PocketBase admin API and returns a short-lived admin JWT.
// Tries the v0.22+ endpoint first (/api/collections/_superusers/auth-with-password),
// then falls back to the v0.21 endpoint (/api/admins/auth-with-password).
func (p *PocketBaseProvider) getAdminToken() (string, error) {
	if p.url == "" || p.adminEmail == "" || p.adminPassword == "" {
		return "", fmt.Errorf("pocketbase admin credentials not configured (POCKETBASE_URL, POCKETBASE_ADMIN_EMAIL, POCKETBASE_ADMIN_PASSWORD)")
	}

	payload := map[string]string{
		"identity": p.adminEmail,
		"password": p.adminPassword,
	}
	payloadBytes, _ := json.Marshal(payload)

	// v0.22+ uses _superusers collection; v0.21 used /api/admins
	endpoints := []string{
		p.url + "/api/collections/_superusers/auth-with-password",
		p.url + "/api/admins/auth-with-password",
	}

	client := &http.Client{Timeout: 10 * time.Second}
	for _, endpoint := range endpoints {
		req, err := http.NewRequest("POST", endpoint, bytes.NewReader(payloadBytes))
		if err != nil {
			continue
		}
		req.Header.Set(headerContentType, headerAppJSON)

		resp, err := client.Do(req)
		if err != nil {
			return "", fmt.Errorf("failed to authenticate with PocketBase admin: %w", err)
		}
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			continue // endpoint not available in this PocketBase version, try next
		}
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("PocketBase admin auth failed (status %d): %s", resp.StatusCode, string(respBody))
		}

		var result struct {
			Token string `json:"token"`
		}
		if err := json.Unmarshal(respBody, &result); err != nil {
			return "", fmt.Errorf("failed to parse admin token: %w", err)
		}
		return result.Token, nil
	}

	return "", fmt.Errorf("PocketBase admin auth failed: no compatible endpoint found (tried v0.22+ and v0.21 APIs)")
}

// SetupJWTSecret configures PocketBase to use the JWT secret from POCKETBASE_JWT_SECRET.
// Called once at backend startup to ensure tokens are signed with the expected secret.
func (p *PocketBaseProvider) SetupJWTSecret() error {
	if len(p.secret) == 0 {
		log.Printf("[PocketBase] POCKETBASE_JWT_SECRET not set, using PocketBase's auto-generated secret. " +
			"Copy the secret from PocketBase admin UI (Settings → Application) to POCKETBASE_JWT_SECRET.")
		return nil
	}

	adminToken, err := p.getAdminToken()
	if err != nil {
		return fmt.Errorf("cannot configure JWT secret: %w", err)
	}

	// Update PocketBase recordAuthToken secret via settings API
	settings := map[string]interface{}{
		"tokens": map[string]interface{}{
			"recordAuthToken": map[string]interface{}{
				"secret":   string(p.secret),
				"duration": 1209600, // 14 days
			},
		},
	}
	body, _ := json.Marshal(settings)

	req, err := http.NewRequest("PATCH", p.url+"/api/settings", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create settings request: %w", err)
	}
	req.Header.Set(headerContentType, headerAppJSON)
	req.Header.Set("Authorization", headerBearerPrefix+adminToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update PocketBase settings: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("PocketBase settings API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	log.Printf("[PocketBase] ✓ JWT secret configured successfully")
	return nil
}

func (p *PocketBaseProvider) DeleteUser(userID string) error {
	if p.url == "" {
		log.Printf("[PocketBase] URL not configured, skipping user deletion for %s", userID)
		return nil
	}

	adminToken, err := p.getAdminToken()
	if err != nil {
		return fmt.Errorf("failed to get admin token: %w", err)
	}

	url := fmt.Sprintf("%s/api/collections/%s/records/%s", p.url, p.collectionName, userID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", headerBearerPrefix+adminToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete PocketBase user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("PocketBase admin API returned status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (p *PocketBaseProvider) UpdateUserPassword(userID, newPassword string) error {
	if p.url == "" {
		return fmt.Errorf("pocketbase URL not configured")
	}

	adminToken, err := p.getAdminToken()
	if err != nil {
		return fmt.Errorf("failed to get admin token: %w", err)
	}

	payload := map[string]string{
		"password":        newPassword,
		"passwordConfirm": newPassword,
	}
	body, _ := json.Marshal(payload)

	url := fmt.Sprintf("%s/api/collections/%s/records/%s", p.url, p.collectionName, userID)
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set(headerContentType, headerAppJSON)
	req.Header.Set("Authorization", headerBearerPrefix+adminToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update PocketBase user password: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("PocketBase admin API returned status %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}
