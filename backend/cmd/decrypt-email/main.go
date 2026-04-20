// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

// decrypt-email is a diagnostic CLI tool that decrypts an email_encrypted value
// from the database using the same emailcrypto package as the backend.
//
// Usage:
//
//	EMAIL_ENCRYPTION_KEY=<hex-key> go run ./cmd/decrypt-email <base64-ciphertext>
package main

import (
	"fmt"
	"os"

	"kagibi/backend/pkg/emailcrypto"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: decrypt-email <base64-email-encrypted>")
		fmt.Fprintln(os.Stderr, "       EMAIL_ENCRYPTION_KEY must be set in the environment")
		os.Exit(1)
	}

	emailcrypto.Init()

	plaintext, err := emailcrypto.Decrypt(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: decryption failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(plaintext)
}
