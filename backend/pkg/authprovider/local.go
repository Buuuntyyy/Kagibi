// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package authprovider

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"time"

	"kagibi/backend/pkg/emailcrypto"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/uptrace/bun"
	"golang.org/x/crypto/bcrypt"
)

const queryIDEq = "id = ?"

// authUser maps to the auth_users table managed exclusively by LocalProvider.
type authUser struct {
	bun.BaseModel      `bun:"table:auth_users,alias:au"`
	ID                 string     `bun:"id,pk"`
	EmailHash          string     `bun:"email_hash,notnull"`
	EmailEncrypted     string     `bun:"email_encrypted,notnull"`
	Email              string     `bun:"-"` // virtual: decrypted from EmailEncrypted after load
	PasswordHash       string     `bun:"password_hash,notnull"`
	TOTPSecret         string     `bun:"totp_secret"`
	TOTPEnabled        bool       `bun:"totp_enabled,notnull,default:false"`
	TOTPFactorID       string     `bun:"totp_factor_id"`
	TOTPFriendlyName   string     `bun:"totp_friendly_name"`
	TOTPLastCode       string     `bun:"totp_last_code"`
	TOTPLastCodeAt     *time.Time `bun:"totp_last_code_at"`
	TOTPLastStep       int64      `bun:"totp_last_step,notnull,default:0"`
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
		log.Fatalf("[LocalAuth] FATAL: JWT_SECRET environment variable is not set. " +
			"Set a stable secret (e.g. openssl rand -hex 32) so tokens survive restarts.")
	}
	return &LocalProvider{secret: []byte(secret), db: db}
}

func (p *LocalProvider) Name() string           { return "local" }
func (p *LocalProvider) GetUserIDClaim() string { return "sub" }
func (p *LocalProvider) GetJWTSecret() []byte   { return p.secret }

// GenerateToken creates a signed HS256 JWT valid for 7 days with aal1.
func (p *LocalProvider) GenerateToken(userID, email string) (string, error) {
	return p.GenerateTokenWithClaims(userID, email, "aal1", false, 0)
}

// GenerateTokenWithAAL creates a signed HS256 JWT with an explicit AAL claim.
// aal should be "aal1" (password only) or "aal2" (password + TOTP verified).
func (p *LocalProvider) GenerateTokenWithAAL(userID, email, aal string) (string, error) {
	return p.GenerateTokenWithClaims(userID, email, aal, false, 0)
}

// GenerateTokenWithClaims creates a signed HS256 JWT.
//   - mfaEnabled adds an "mfa":"enabled" claim so step-up middleware can require aal2
//     for MFA-enrolled users without a per-request database lookup.
//   - mfaVerifiedAt (unix seconds, 0 to omit) stamps an "mfa_at" claim marking when the
//     session last completed MFA, used to enforce per-action step-up freshness.
//
// Both claims are signed, so a client cannot forge or strip them.
func (p *LocalProvider) GenerateTokenWithClaims(userID, email, aal string, mfaEnabled bool, mfaVerifiedAt int64) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"aal":   aal,
		"exp":   time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
	}
	if mfaEnabled {
		claims["mfa"] = "enabled"
	}
	if mfaVerifiedAt > 0 {
		claims["mfa_at"] = mfaVerifiedAt
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
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
	pwHash, err := p.HashPassword(password)
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

	emailHash := emailcrypto.Hash(email)
	emailEnc, err := emailcrypto.Encrypt(email)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt email: %w", err)
	}

	au := &authUser{
		ID:             userID,
		EmailHash:      emailHash,
		EmailEncrypted: emailEnc,
		PasswordHash:   pwHash,
	}
	_, err = p.db.NewInsert().Model(au).Exec(context.Background())
	if err != nil {
		return "", err
	}
	return userID, nil
}

// FindAuthUserByEmail looks up an auth_users record by email using the HMAC hash index.
func (p *LocalProvider) FindAuthUserByEmail(email string) (*authUser, error) {
	hash := emailcrypto.Hash(email)
	var au authUser
	err := p.db.NewSelect().Model(&au).Where("email_hash = ?", hash).Scan(context.Background())
	if err != nil {
		return nil, err
	}
	plain, err := emailcrypto.Decrypt(au.EmailEncrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt auth user email: %w", err)
	}
	au.Email = plain
	return &au, nil
}

// DeleteUser removes the auth_users record for the given user ID.
func (p *LocalProvider) DeleteUser(userID string) error {
	_, err := p.db.NewDelete().Model((*authUser)(nil)).Where(queryIDEq, userID).Exec(context.Background())
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
		Where(queryIDEq, userID).
		Exec(context.Background())
	return err
}

// UpdateUserPasswordWithVerification verifies the old password before updating.
// Returns a typed error string when the old password is wrong so callers can return 401.
func (p *LocalProvider) UpdateUserPasswordWithVerification(userID, oldPassword, newPassword string) error {
	var au authUser
	if err := p.db.NewSelect().Model(&au).Where(queryIDEq, userID).Scan(context.Background()); err != nil {
		return fmt.Errorf("user not found")
	}
	if err := p.CheckPassword(au.PasswordHash, oldPassword); err != nil {
		return fmt.Errorf("invalid current password")
	}
	return p.UpdateUserPassword(userID, newPassword)
}

// UpdateUserEmailWithVerification verifies the current password before updating the email.
// Returns an error if the password is wrong or the new email is already taken.
func (p *LocalProvider) UpdateUserEmailWithVerification(userID, password, newEmail string) error {
	var au authUser
	if err := p.db.NewSelect().Model(&au).Where(queryIDEq, userID).Scan(context.Background()); err != nil {
		return fmt.Errorf("user not found")
	}
	if err := p.CheckPassword(au.PasswordHash, password); err != nil {
		return fmt.Errorf("invalid current password")
	}
	newEmailHash := emailcrypto.Hash(newEmail)
	newEmailEnc, err := emailcrypto.Encrypt(newEmail)
	if err != nil {
		return fmt.Errorf("failed to encrypt new email: %w", err)
	}
	_, err = p.db.NewUpdate().Model((*authUser)(nil)).
		Set("email_hash = ?, email_encrypted = ?", newEmailHash, newEmailEnc).
		Where(queryIDEq, userID).
		Exec(context.Background())
	return err
}

// ReissueToken verifies the password for the given email and, if correct, returns
// the user ID and a fresh JWT. Used during orphan-signup recovery so the caller
// never needs to access the unexported authUser type directly.
func (p *LocalProvider) ReissueToken(email, password string) (userID, token string, err error) {
	au, err := p.FindAuthUserByEmail(email)
	if err != nil {
		return "", "", err
	}
	if err := p.CheckPassword(au.PasswordHash, password); err != nil {
		return "", "", err
	}
	tok, err := p.GenerateToken(au.ID, email)
	return au.ID, tok, err
}

// GetAuthUserByID returns the full auth_users row for the given user ID.
func (p *LocalProvider) GetAuthUserByID(userID string) (*authUser, error) {
	var au authUser
	if err := p.db.NewSelect().Model(&au).Where(queryIDEq, userID).Scan(context.Background()); err != nil {
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

	// Encrypt the shared secret at rest so a read-only DB leak cannot be used to
	// generate valid TOTP codes. The plaintext is only returned to the enrolling
	// client (for the QR code), never persisted in the clear.
	encSecret, err := emailcrypto.EncryptSecret(key.Secret())
	if err != nil {
		return "", "", "", fmt.Errorf("failed to encrypt TOTP secret: %w", err)
	}

	_, err = p.db.NewUpdate().Model((*authUser)(nil)).
		Set("totp_secret = ?, totp_factor_id = ?, totp_friendly_name = ?, totp_enabled = false",
			encSecret, factorID, friendlyName).
		Where(queryIDEq, userID).
		Exec(context.Background())
	if err != nil {
		return "", "", "", fmt.Errorf("failed to store TOTP secret: %w", err)
	}

	return factorID, key.URL(), key.Secret(), nil
}

// decryptTOTPSecret returns the plaintext TOTP secret from its stored form,
// transparently handling legacy rows that predate at-rest encryption. legacy is
// true when the stored value was plaintext and should be upgraded.
func (p *LocalProvider) decryptTOTPSecret(stored string) (secret string, legacy bool) {
	if pt, err := emailcrypto.DecryptSecret(stored); err == nil {
		return pt, false
	}
	return stored, true
}

// DecryptTOTPSecret returns the plaintext TOTP secret for re-displaying a pending
// enrollment's QR code. Accepts both encrypted and legacy-plaintext stored values.
func (p *LocalProvider) DecryptTOTPSecret(stored string) string {
	secret, _ := p.decryptTOTPSecret(stored)
	return secret
}

// ValidateTOTPCode checks a 6-digit code against the user's stored TOTP secret.
// Enforces single-use-per-time-step replay protection (a code cannot be reused
// anywhere in its validity window) and per-user lockout (5 failures → 15 min lock).
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

	secret, legacy := p.decryptTOTPSecret(au.TOTPSecret)

	// Identify which 30s time-step the code matches (searching ±1 period, replicating
	// the previous skew) so replays can be rejected across the whole validity window.
	matchedStep, valid := matchTOTPStep(code, secret)
	if !valid {
		newAttempts := au.TOTPFailedAttempts + 1
		upd := p.db.NewUpdate().Model((*authUser)(nil)).Where(queryIDEq, userID)
		if newAttempts >= 5 {
			lockedUntil := time.Now().Add(15 * time.Minute)
			upd = upd.Set("totp_failed_attempts = ?, totp_locked_until = ?", newAttempts, lockedUntil)
		} else {
			upd = upd.Set("totp_failed_attempts = ?", newAttempts)
		}
		_, _ = upd.Exec(context.Background())
		return fmt.Errorf("invalid TOTP code")
	}

	// Replay prevention — a code is single-use across its entire validity window:
	// reject any code from a time-step that has already been consumed (or an older one).
	if matchedStep <= au.TOTPLastStep {
		return fmt.Errorf("TOTP code already used")
	}

	// Valid and fresh — reset counters, advance the last-used step, record the code.
	now := time.Now()
	upd := p.db.NewUpdate().Model((*authUser)(nil)).
		Set("totp_failed_attempts = 0, totp_locked_until = NULL, totp_last_step = ?, totp_last_code = ?, totp_last_code_at = ?", matchedStep, code, now).
		Where(queryIDEq, userID)
	// Opportunistically upgrade a legacy plaintext secret to the encrypted form now
	// that we have verified it (and thus hold the plaintext).
	if legacy {
		if enc, encErr := emailcrypto.EncryptSecret(secret); encErr == nil {
			upd = upd.Set("totp_secret = ?", enc)
		}
	}
	_, _ = upd.Exec(context.Background())
	return nil
}

// matchTOTPStep returns the 30-second time-step (unix/period) whose TOTP code equals
// the submitted code, searching the current step and its immediate neighbours (±1
// period, matching the previous default skew). valid is false when no step matches.
func matchTOTPStep(code, secret string) (step int64, valid bool) {
	const period = 30
	opts := totp.ValidateOpts{
		Period:    period,
		Skew:      0,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	}
	nowStep := time.Now().Unix() / period
	for _, s := range []int64{nowStep - 1, nowStep, nowStep + 1} {
		if ok, err := totp.ValidateCustom(code, secret, time.Unix(s*period, 0), opts); err == nil && ok {
			return s, true
		}
	}
	return 0, false
}

// ActivateTOTP marks the TOTP factor as verified (totp_enabled = true).
func (p *LocalProvider) ActivateTOTP(userID string) error {
	_, err := p.db.NewUpdate().Model((*authUser)(nil)).
		Set("totp_enabled = true").
		Where(queryIDEq, userID).
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

// GenerateGuestToken issues a short-lived HS256 JWT for a P2P guest session.
// The token expires at expiresAt (matches the invite expiry) and carries is_guest=true
// so middleware can distinguish guest sessions from full accounts.
func (p *LocalProvider) GenerateGuestToken(guestUserID string, expiresAt time.Time) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      guestUserID,
		"is_guest": true,
		"aal":      "guest",
		"exp":      expiresAt.Unix(),
		"iat":      time.Now().Unix(),
	})
	return token.SignedString(p.secret)
}

// DisableTOTP removes all TOTP data for the user.
func (p *LocalProvider) DisableTOTP(userID string) error {
	_, err := p.db.NewUpdate().Model((*authUser)(nil)).
		Set("totp_enabled = false, totp_secret = NULL, totp_factor_id = NULL, totp_friendly_name = NULL").
		Where(queryIDEq, userID).
		Exec(context.Background())
	return err
}
