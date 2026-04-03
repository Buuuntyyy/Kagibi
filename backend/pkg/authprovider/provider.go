package authprovider

// AuthProvider defines the interface for authentication providers (Supabase or PocketBase).
// The backend uses this to validate JWTs and perform admin operations (delete/update users).
type AuthProvider interface {
	// Name returns the provider identifier ("supabase" or "pocketbase")
	Name() string
	// GetUserIDClaim returns the JWT claim that holds the user ID ("sub" for Supabase, "id" for PocketBase)
	GetUserIDClaim() string
	// GetJWTSecret returns the HMAC secret used to validate tokens
	GetJWTSecret() []byte
	// DeleteUser removes a user from the auth provider (called on profile creation failure or RGPD deletion)
	DeleteUser(userID string) error
	// UpdateUserPassword updates a user's password via the provider admin API (used in account recovery)
	UpdateUserPassword(userID, newPassword string) error
}
