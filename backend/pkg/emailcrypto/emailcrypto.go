// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package emailcrypto provides server-side email encryption at rest.
//
// Design:
//   - Hash(email)    → HMAC-SHA256 of normalised email → used as a stable lookup index (deterministic)
//   - Encrypt(email) → AES-256-GCM with a random nonce → stored in email_encrypted column
//   - Decrypt(ct)    → recovers the plaintext email when the backend needs it
//
// A single EMAIL_ENCRYPTION_KEY (32-byte hex) is split into two independent subkeys so
// that compromising one operation does not compromise the other.
//
// Protection: a DB dump alone (without the server secret) reveals no email addresses.
// Limitation: a full server compromise (code + env vars + DB) defeats the encryption,
// as with any server-side encryption scheme.
package emailcrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

var hashKey []byte
var encKey []byte

// Init reads EMAIL_ENCRYPTION_KEY from the environment and derives two subkeys.
// Must be called once at startup, before any Hash/Encrypt/Decrypt call.
func Init() {
	raw := os.Getenv("EMAIL_ENCRYPTION_KEY")
	if raw == "" {
		log.Fatal("[emailcrypto] EMAIL_ENCRYPTION_KEY is not set. Generate one with: openssl rand -hex 32")
	}
	keyBytes, err := hex.DecodeString(raw)
	if err != nil || len(keyBytes) != 32 {
		log.Fatalf("[emailcrypto] EMAIL_ENCRYPTION_KEY must be exactly 64 hex characters (32 bytes). Got len=%d, err=%v", len(raw)/2, err)
	}
	hashKey = deriveSubkey(keyBytes, "email-hash-v1")
	encKey = deriveSubkey(keyBytes, "email-enc-v1")
}

func deriveSubkey(master []byte, context string) []byte {
	mac := hmac.New(sha256.New, master)
	mac.Write([]byte(context))
	return mac.Sum(nil)
}

// Hash returns the HMAC-SHA256 of the lower-cased, trimmed email, encoded as 64 hex chars.
// Deterministic — same input always yields the same output — safe to use as a UNIQUE index.
func Hash(email string) string {
	mac := hmac.New(sha256.New, hashKey)
	mac.Write([]byte(strings.ToLower(strings.TrimSpace(email))))
	return hex.EncodeToString(mac.Sum(nil))
}

// Encrypt encrypts the email with AES-256-GCM using a fresh random 12-byte nonce.
// The returned string is base64(nonce || ciphertext || auth_tag).
// Each call produces a different ciphertext even for the same input.
func Encrypt(email string) (string, error) {
	block, err := aes.NewCipher(encKey)
	if err != nil {
		return "", fmt.Errorf("emailcrypto: create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("emailcrypto: create GCM: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("emailcrypto: generate nonce: %w", err)
	}
	ct := gcm.Seal(nonce, nonce, []byte(email), nil)
	return base64.StdEncoding.EncodeToString(ct), nil
}

// Decrypt decrypts a ciphertext previously produced by Encrypt.
func Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("emailcrypto: base64 decode: %w", err)
	}
	block, err := aes.NewCipher(encKey)
	if err != nil {
		return "", fmt.Errorf("emailcrypto: create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("emailcrypto: create GCM: %w", err)
	}
	if len(data) < gcm.NonceSize() {
		return "", fmt.Errorf("emailcrypto: ciphertext too short")
	}
	nonce, ct := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", fmt.Errorf("emailcrypto: decrypt failed (wrong key or corrupted data): %w", err)
	}
	return string(plain), nil
}
