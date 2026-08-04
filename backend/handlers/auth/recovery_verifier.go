// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// recoveryVerifierPrefix marks a stored recovery verifier that has been hashed
// server-side (format "v2:<hex>"). Values without this prefix are legacy
// verifiers stored as the raw client hash and are upgraded on next successful use.
const recoveryVerifierPrefix = "v2:"

// hashRecoveryVerifier converts the client-supplied recovery hash (itself already
// a SHA-256 of the high-entropy recovery code) into an at-rest verifier by hashing
// it again server-side. Storing this form means a read-only DB leak no longer
// yields a value that can be replayed to reset an account.
func hashRecoveryVerifier(clientHash string) string {
	sum := sha256.Sum256([]byte(clientHash))
	return recoveryVerifierPrefix + hex.EncodeToString(sum[:])
}

// verifyRecoveryHash checks a client-supplied recovery hash against the stored
// verifier in constant time. It transparently supports legacy plaintext verifiers.
// needsUpgrade is true when a legacy verifier matched and should be rewritten in
// the hashed format.
func verifyRecoveryHash(stored, clientHash string) (ok bool, needsUpgrade bool) {
	if strings.HasPrefix(stored, recoveryVerifierPrefix) {
		return hmac.Equal([]byte(stored), []byte(hashRecoveryVerifier(clientHash))), false
	}
	if hmac.Equal([]byte(stored), []byte(clientHash)) {
		return true, true
	}
	return false, false
}
