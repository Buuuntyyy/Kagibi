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

// SupabaseProvider validates Supabase HS256 JWTs and uses the Supabase Admin API
// for user management operations.
type SupabaseProvider struct {
	url      string
	adminKey string
	secret   []byte
}

func NewSupabaseProvider() *SupabaseProvider {
	adminKey := os.Getenv("SUPABASE_ADMIN_KEY")
	if adminKey == "" {
		adminKey = os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	}
	return &SupabaseProvider{
		url:      os.Getenv("SUPABASE_URL"),
		adminKey: adminKey,
		secret:   []byte(os.Getenv("SUPABASE_JWT_SECRET")),
	}
}

func (p *SupabaseProvider) Name() string           { return "supabase" }
func (p *SupabaseProvider) GetUserIDClaim() string { return "sub" }
func (p *SupabaseProvider) GetJWTSecret() []byte   { return p.secret }

func (p *SupabaseProvider) DeleteUser(userID string) error {
	if p.url == "" || p.adminKey == "" {
		log.Printf("[Supabase] URL or admin key not configured, skipping user deletion for %s", userID)
		return nil
	}

	url := fmt.Sprintf("%s/auth/v1/admin/users/%s", p.url, userID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.adminKey))
	req.Header.Set("apikey", p.adminKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute DELETE request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("supabase admin API returned status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (p *SupabaseProvider) UpdateUserPassword(userID, newPassword string) error {
	if p.url == "" || p.adminKey == "" {
		return fmt.Errorf("supabase credentials not configured")
	}

	payload := map[string]string{"password": newPassword}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/auth/v1/admin/users/%s", p.url, userID)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.adminKey))
	req.Header.Set("apikey", p.adminKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("supabase admin API returned status %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}
