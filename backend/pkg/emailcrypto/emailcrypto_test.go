// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package emailcrypto

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Setenv("EMAIL_ENCRYPTION_KEY", "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff")
	Init()
	os.Exit(m.Run())
}

func TestEmailRoundTrip(t *testing.T) {
	const email = "User@Example.com"
	ct, err := Encrypt(email)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	pt, err := Decrypt(ct)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if pt != email {
		t.Fatalf("round-trip mismatch: got %q want %q", pt, email)
	}
}

func TestSecretRoundTrip(t *testing.T) {
	const secret = "JBSWY3DPEHPK3PXP" // sample base32 TOTP secret
	ct, err := EncryptSecret(secret)
	if err != nil {
		t.Fatalf("EncryptSecret: %v", err)
	}
	pt, err := DecryptSecret(ct)
	if err != nil {
		t.Fatalf("DecryptSecret: %v", err)
	}
	if pt != secret {
		t.Fatalf("round-trip mismatch: got %q want %q", pt, secret)
	}
}

// The email and secret subkeys must be independent: a ciphertext produced under one
// must not decrypt under the other.
func TestKeySeparation(t *testing.T) {
	ct, err := EncryptSecret("JBSWY3DPEHPK3PXP")
	if err != nil {
		t.Fatalf("EncryptSecret: %v", err)
	}
	if _, err := Decrypt(ct); err == nil {
		t.Fatal("email key unexpectedly decrypted a secret-key ciphertext")
	}
}

// A legacy plaintext TOTP secret is not valid ciphertext, so DecryptSecret must fail —
// this is what lets the provider fall back to treating the stored value as plaintext.
func TestDecryptSecretRejectsPlaintext(t *testing.T) {
	if _, err := DecryptSecret("JBSWY3DPEHPK3PXP"); err == nil {
		t.Fatal("expected error decrypting a plaintext value")
	}
}
