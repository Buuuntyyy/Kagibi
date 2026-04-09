package authprovider

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/uptrace/bun"
	"golang.org/x/crypto/bcrypt"
)

// authUser maps to the auth_users table managed exclusively by LocalProvider.
type authUser struct {
	bun.BaseModel      `bun:"table:auth_users,alias:au"`
	ID                 string     `bun:"id,pk"`
	Email              string     `bun:"email,unique,notnull"`
	PasswordHash       string     `bun:"password_hash,notnull"`
	TOTPSecret         string     `bun:"totp_secret"`
	TOTPEnabled        bool       `bun:"totp_enabled,notnull,default:false"`
	TOTPFactorID       string     `bun:"totp_factor_id"`
	TOTPFriendlyName   string     `bun:"totp_friendly_name"`
	TOTPLastCode       string     `bun:"totp_last_code"`
	TOTPLastCodeAt     *time.Time `bun:"totp_last_code_at"`
	TOTPFailedAttempts int        `bun:"totp_failed_attempts,notnull,default:0"`
	TOTPLockedUntil    *time.Time `bun:"totp_locked_until"`
	CreatedAt          time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp"`
}

// LocalProvider authenticates users against the backend PostgreSQL database.
// No external auth service required — JWTs are signed locally with JWT_SECRET.
type LocalProvider struct {
	secret []byte
	db     *bun.DB
}

func NewLocalProvider(db *bun.DB) *LocalProvider {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		b := make([]byte, 32)
		_, _ = rand.Read(b)
		secret = hex.EncodeToString(b)
		log.Printf("[LocalAuth] WARNING: JWT_SECRET not set — using random secret. Tokens will be invalidated on restart.")
	}
	return &LocalProvider{secret: []byte(secret), db: db}
}

func (p *LocalProvider) Name() string           { return "local" }
func (p *LocalProvider) GetUserIDClaim() string { return "sub" }
func (p *LocalProvider) GetJWTSecret() []byte   { return p.secret }

// GenerateToken creates a signed HS256 JWT valid for 7 days with aal1.
func (p *LocalProvider) GenerateToken(userID, email string) (string, error) {
	return p.GenerateTokenWithAAL(userID, email, "aal1")
}

// GenerateTokenWithAAL creates a signed HS256 JWT with an explicit AAL claim.
// aal should be "aal1" (password only) or "aal2" (password + TOTP verified).
func (p *LocalProvider) GenerateTokenWithAAL(userID, email, aal string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"aal":   aal,
		"exp":   time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
	})
	return token.SignedString(p.secret)
}

// HashPassword hashes a plaintext password with bcrypt (cost 12).
func (p *LocalProvider) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(hash), err
}

// CheckPassword verifies a plaintext password against a bcrypt hash.
func (p *LocalProvider) CheckPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// CreateAuthUser inserts a new auth_users record and returns the generated UUID.
func (p *LocalProvider) CreateAuthUser(email, password string) (string, error) {
	hash, err := p.HashPassword(password)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	idBytes := make([]byte, 16)
	if _, err := rand.Read(idBytes); err != nil {
		return "", fmt.Errorf("failed to generate user ID: %w", err)
	}
	idBytes[6] = (idBytes[6] & 0x0f) | 0x40
	idBytes[8] = (idBytes[8] & 0x3f) | 0x80
	userID := fmt.Sprintf("%x-%x-%x-%x-%x",
		idBytes[0:4], idBytes[4:6], idBytes[6:8], idBytes[8:10], idBytes[10:16])

	au := &authUser{
		ID:           userID,
		Email:        email,
		PasswordHash: hash,
	}
	_, err = p.db.NewInsert().Model(au).Exec(context.Background())
	if err != nil {
		return "", err
	}
	return userID, nil
}

// FindAuthUserByEmail looks up an auth_users record by email.
func (p *LocalProvider) FindAuthUserByEmail(email string) (*authUser, error) {
	var au authUser
	err := p.db.NewSelect().Model(&au).Where("email = ?", email).Scan(context.Background())
	if err != nil {
		return nil, err
	}
	return &au, nil
}

// DeleteUser removes the auth_users record for the given user ID.
func (p *LocalProvider) DeleteUser(userID string) error {
	_, err := p.db.NewDelete().Model((*authUser)(nil)).Where("id = ?", userID).Exec(context.Background())
	return err
}

// UpdateUserPassword updates the bcrypt hash for the given user ID.
func (p *LocalProvider) UpdateUserPassword(userID, newPassword string) error {
	hash, err := p.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	_, err = p.db.NewUpdate().Model((*authUser)(nil)).
		Set("password_hash = ?", hash).
		Where("id = ?", userID).
		Exec(context.Background())
	return err
}

// UpdateUserPasswordWithVerification verifies the old password before updating.
// Returns a typed error string when the old password is wrong so callers can return 401.
func (p *LocalProvider) UpdateUserPasswordWithVerification(userID, oldPassword, newPassword string) error {
	var au authUser
	if err := p.db.NewSelect().Model(&au).Where("id = ?", userID).Scan(context.Background()); err != nil {
		return fmt.Errorf("user not found")
	}
	if err := p.CheckPassword(au.PasswordHash, oldPassword); err != nil {
		return fmt.Errorf("invalid current password")
	}
	return p.UpdateUserPassword(userID, newPassword)
}

// GetAuthUserByID returns the full auth_users row for the given user ID.
func (p *LocalProvider) GetAuthUserByID(userID string) (*authUser, error) {
	var au authUser
	if err := p.db.NewSelect().Model(&au).Where("id = ?", userID).Scan(context.Background()); err != nil {
		return nil, err
	}
	return &au, nil
}

// StartTOTPEnrollment generates a new TOTP secret, stores it (unverified), and returns the
// factor ID, OTP URI (for QR code), and raw base32 secret.
func (p *LocalProvider) StartTOTPEnrollment(userID, email, friendlyName string) (factorID, otpURI, secret string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Kagibi",
		AccountName: email,
		Period:      30,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate TOTP: %w", err)
	}

	idBytes := make([]byte, 16)
	if _, err := rand.Read(idBytes); err != nil {
		return "", "", "", fmt.Errorf("failed to generate factor ID: %w", err)
	}
	idBytes[6] = (idBytes[6] & 0x0f) | 0x40
	idBytes[8] = (idBytes[8] & 0x3f) | 0x80
	factorID = fmt.Sprintf("%x-%x-%x-%x-%x",
		idBytes[0:4], idBytes[4:6], idBytes[6:8], idBytes[8:10], idBytes[10:16])

	_, err = p.db.NewUpdate().Model((*authUser)(nil)).
		Set("totp_secret = ?, totp_factor_id = ?, totp_friendly_name = ?, totp_enabled = false",
			key.Secret(), factorID, friendlyName).
		Where("id = ?", userID).
		Exec(context.Background())
	if err != nil {
		return "", "", "", fmt.Errorf("failed to store TOTP secret: %w", err)
	}

	return factorID, key.URL(), key.Secret(), nil
}

// ValidateTOTPCode checks a 6-digit code against the user's stored TOTP secret.
// Enforces replay protection (30s window) and per-user lockout (5 failures → 15 min lock).
func (p *LocalProvider) ValidateTOTPCode(userID, code string) error {
	au, err := p.GetAuthUserByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}
	if au.TOTPSecret == "" {
		return fmt.Errorf("MFA not configured")
	}

	// Lockout check
	if au.TOTPLockedUntil != nil && time.Now().Before(*au.TOTPLockedUntil) {
		return fmt.Errorf("MFA temporarily locked due to too many failed attempts")
	}

	// Replay attack prevention — same code can't be used twice within the same 30s window
	if au.TOTPLastCode == code && au.TOTPLastCodeAt != nil && time.Since(*au.TOTPLastCodeAt) < 30*time.Second {
		return fmt.Errorf("TOTP code already used")
	}

	if !totp.Validate(code, au.TOTPSecret) {
		newAttempts := au.TOTPFailedAttempts + 1
		upd := p.db.NewUpdate().Model((*authUser)(nil)).Where("id = ?", userID)
		if newAttempts >= 5 {
			lockedUntil := time.Now().Add(15 * time.Minute)
			upd = upd.Set("totp_failed_attempts = ?, totp_locked_until = ?", newAttempts, lockedUntil)
		} else {
			upd = upd.Set("totp_failed_attempts = ?", newAttempts)
		}
		_, _ = upd.Exec(context.Background())
		return fmt.Errorf("invalid TOTP code")
	}

	// Valid — reset counters and record the used code to prevent replay
	now := time.Now()
	_, _ = p.db.NewUpdate().Model((*authUser)(nil)).
		Set("totp_failed_attempts = 0, totp_locked_until = NULL, totp_last_code = ?, totp_last_code_at = ?", code, now).
		Where("id = ?", userID).
		Exec(context.Background())
	return nil
}

// ActivateTOTP marks the TOTP factor as verified (totp_enabled = true).
func (p *LocalProvider) ActivateTOTP(userID string) error {
	_, err := p.db.NewUpdate().Model((*authUser)(nil)).
		Set("totp_enabled = true").
		Where("id = ?", userID).
		Exec(context.Background())
	return err
}

// SyncMFAStatus upserts user_security_settings to reflect the current TOTP state.
// Called by MFA handlers after activating or disabling TOTP so both tables stay consistent.
// Only touches mfa_enabled and mfa_verified — never overwrites the user's require_mfa_* preferences.
func (p *LocalProvider) SyncMFAStatus(userID string, enabled bool) error {
	_, err := p.db.ExecContext(context.Background(), `
		INSERT INTO user_security_settings (user_id, mfa_enabled, mfa_verified)
		VALUES (?, ?, ?)
		ON CONFLICT (user_id) DO UPDATE
		  SET mfa_enabled  = EXCLUDED.mfa_enabled,
		      mfa_verified = EXCLUDED.mfa_verified
	`, userID, enabled, enabled)
	return err
}

// DisableTOTP removes all TOTP data for the user.
func (p *LocalProvider) DisableTOTP(userID string) error {
	_, err := p.db.NewUpdate().Model((*authUser)(nil)).
		Set("totp_enabled = false, totp_secret = NULL, totp_factor_id = NULL, totp_friendly_name = NULL").
		Where("id = ?", userID).
		Exec(context.Background())
	return err
}
