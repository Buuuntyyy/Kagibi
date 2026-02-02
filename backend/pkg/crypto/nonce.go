// Package nonce provides NIST SP 800-38D compliant nonce generation for AES-GCM.
//
// This implementation follows:
// - NIST SP 800-38D Section 8.2 (Nonce Generation)
// - ANSSI recommendations for symmetric encryption
//
// Nonce Format: 96 bits (12 bytes) - optimal for AES-GCM without GHASH overhead
package nonce

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
)

const (
	// NonceLength is the recommended nonce size for AES-GCM (96 bits)
	// NIST SP 800-38D Section 8.2.1
	NonceLength = 12

	// BaseNonceLength is the random portion for chunk-based encryption
	BaseNonceLength = 8

	// CounterLength is the chunk counter portion (32 bits = 4 bytes)
	CounterLength = 4

	// TagLength is the authentication tag size (128 bits = 16 bytes)
	TagLength = 16
)

var (
	ErrInvalidBaseNonce   = errors.New("base nonce must be exactly 8 bytes")
	ErrInvalidChunkIndex  = errors.New("chunk index exceeds 32-bit limit")
	ErrNonceReuseDetected = errors.New("CRITICAL: nonce reuse detected")
)

// Generator provides cryptographically secure nonce generation.
// Thread-safe with optional reuse detection.
type Generator struct {
	mu         sync.Mutex
	usedNonces map[string]struct{}
	trackUsage bool
	maxTracked int
}

// NewGenerator creates a new nonce generator.
// If trackUsage is true, generates nonces are tracked to detect reuse.
func NewGenerator(trackUsage bool) *Generator {
	g := &Generator{
		trackUsage: trackUsage,
		maxTracked: 100000, // Prevent memory exhaustion
	}
	if trackUsage {
		g.usedNonces = make(map[string]struct{})
	}
	return g
}

// Generate creates a cryptographically random 96-bit nonce.
// Uses crypto/rand which is backed by the OS CSPRNG.
func (g *Generator) Generate() ([]byte, error) {
	nonce := make([]byte, NonceLength)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("CSPRNG failed: %w", err)
	}

	if g.trackUsage {
		g.mu.Lock()
		defer g.mu.Unlock()

		key := string(nonce)
		if _, exists := g.usedNonces[key]; exists {
			return nil, ErrNonceReuseDetected
		}

		// Memory management
		if len(g.usedNonces) >= g.maxTracked {
			g.usedNonces = make(map[string]struct{})
		}
		g.usedNonces[key] = struct{}{}
	}

	return nonce, nil
}

// GenerateBaseNonce creates an 8-byte random base for chunk-based encryption.
func (g *Generator) GenerateBaseNonce() ([]byte, error) {
	base := make([]byte, BaseNonceLength)
	if _, err := rand.Read(base); err != nil {
		return nil, fmt.Errorf("CSPRNG failed: %w", err)
	}
	return base, nil
}

// GenerateChunkNonce creates a deterministic nonce from base + chunk index.
// Format: [8 bytes random base] + [4 bytes chunk counter (little-endian)]
//
// This ensures:
// - Uniqueness across files (random base)
// - Uniqueness within file (counter)
// - No nonce reuse with up to 2^32 chunks per file
func GenerateChunkNonce(baseNonce []byte, chunkIndex uint32) ([]byte, error) {
	if len(baseNonce) != BaseNonceLength {
		return nil, ErrInvalidBaseNonce
	}

	nonce := make([]byte, NonceLength)
	copy(nonce[:BaseNonceLength], baseNonce)
	binary.LittleEndian.PutUint32(nonce[BaseNonceLength:], chunkIndex)

	return nonce, nil
}

// EncryptedChunkFormat represents the wire format for encrypted chunks.
// Format: [Nonce (12B)] + [Ciphertext] + [Tag (16B)]
type EncryptedChunkFormat struct {
	Nonce            []byte
	CiphertextAndTag []byte
}

// Serialize combines nonce and ciphertext into wire format.
func (f *EncryptedChunkFormat) Serialize() []byte {
	result := make([]byte, NonceLength+len(f.CiphertextAndTag))
	copy(result[:NonceLength], f.Nonce)
	copy(result[NonceLength:], f.CiphertextAndTag)
	return result
}

// ParseEncryptedChunk extracts nonce and ciphertext from wire format.
func ParseEncryptedChunk(data []byte) (*EncryptedChunkFormat, error) {
	minSize := NonceLength + TagLength + 1 // nonce + tag + at least 1 byte
	if len(data) < minSize {
		return nil, fmt.Errorf("encrypted data too short: %d bytes (minimum: %d)", len(data), minSize)
	}

	return &EncryptedChunkFormat{
		Nonce:            data[:NonceLength],
		CiphertextAndTag: data[NonceLength:],
	}, nil
}

// Quick helper functions for simple use cases

// GenerateNonce creates a single random 96-bit nonce.
func GenerateNonce() ([]byte, error) {
	nonce := make([]byte, NonceLength)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return nonce, nil
}

// GenerateBaseNonceSimple creates a single 8-byte random base nonce.
func GenerateBaseNonceSimple() ([]byte, error) {
	base := make([]byte, BaseNonceLength)
	if _, err := rand.Read(base); err != nil {
		return nil, err
	}
	return base, nil
}
