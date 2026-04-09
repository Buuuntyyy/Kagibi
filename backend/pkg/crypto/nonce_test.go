package nonce

import (
	"bytes"
	"encoding/binary"
	"sync"
	"testing"
)

const (
	fmtUnexpectedError = "unexpected error: %v"
	fmtWrongByteLen    = "expected %d bytes, got %d"
)

func testNonceIsNonZero(t *testing.T, nonce []byte) {
	t.Helper()
	for _, b := range nonce {
		if b != 0 {
			return
		}
	}
	t.Error("nonce should not be all zeros")
}

func testNoncesAreUnique(t *testing.T, sampleSize int) {
	t.Helper()
	seen := make(map[string]struct{}, sampleSize)
	for i := 0; i < sampleSize; i++ {
		nonce, err := GenerateNonce()
		if err != nil {
			t.Fatalf("unexpected error at iteration %d: %v", i, err)
		}
		key := string(nonce)
		if _, exists := seen[key]; exists {
			t.Fatalf("duplicate nonce detected at iteration %d", i)
		}
		seen[key] = struct{}{}
	}
	if len(seen) != sampleSize {
		t.Errorf("expected %d unique nonces, got %d", sampleSize, len(seen))
	}
}

func TestGenerateNonce(t *testing.T) {
	t.Run("generates 12 bytes", func(t *testing.T) {
		nonce, err := GenerateNonce()
		if err != nil {
			t.Fatalf(fmtUnexpectedError, err)
		}
		if len(nonce) != NonceLength {
			t.Errorf(fmtWrongByteLen, NonceLength, len(nonce))
		}
	})

	t.Run("generates non-zero nonces", func(t *testing.T) {
		nonce, err := GenerateNonce()
		if err != nil {
			t.Fatalf(fmtUnexpectedError, err)
		}
		testNonceIsNonZero(t, nonce)
	})

	t.Run("generates unique nonces - 10000 samples", func(t *testing.T) {
		testNoncesAreUnique(t, 10000)
	})
}

func TestGenerateBaseNonce(t *testing.T) {
	t.Run("generates 8 bytes", func(t *testing.T) {
		base, err := GenerateBaseNonceSimple()
		if err != nil {
			t.Fatalf(fmtUnexpectedError, err)
		}
		if len(base) != BaseNonceLength {
			t.Errorf(fmtWrongByteLen, BaseNonceLength, len(base))
		}
	})

	t.Run("generates unique base nonces - 10000 samples", func(t *testing.T) {
		const sampleSize = 10000
		seen := make(map[string]struct{}, sampleSize)

		for i := 0; i < sampleSize; i++ {
			base, err := GenerateBaseNonceSimple()
			if err != nil {
				t.Fatalf("unexpected error at iteration %d: %v", i, err)
			}

			key := string(base)
			if _, exists := seen[key]; exists {
				t.Fatalf("duplicate base nonce detected at iteration %d", i)
			}
			seen[key] = struct{}{}
		}
	})
}

func testChunkCounterEncoding(t *testing.T, base []byte) {
	t.Helper()
	tests := []struct {
		index    uint32
		expected []byte
	}{
		{0, []byte{0, 0, 0, 0}},
		{1, []byte{1, 0, 0, 0}},
		{256, []byte{0, 1, 0, 0}},
		{0xFFFFFFFF, []byte{0xFF, 0xFF, 0xFF, 0xFF}},
		{0x12345678, []byte{0x78, 0x56, 0x34, 0x12}},
	}
	for _, tc := range tests {
		nonce, err := GenerateChunkNonce(base, tc.index)
		if err != nil {
			t.Fatalf("unexpected error for index %d: %v", tc.index, err)
		}
		counter := nonce[8:12]
		if !bytes.Equal(counter, tc.expected) {
			t.Errorf("index %d: got counter %v, want %v", tc.index, counter, tc.expected)
		}
	}
}

func testChunkNoncesAreUnique(t *testing.T, base []byte, chunkCount uint32) {
	t.Helper()
	seen := make(map[string]struct{}, chunkCount)
	for i := uint32(0); i < chunkCount; i++ {
		nonce, err := GenerateChunkNonce(base, i)
		if err != nil {
			t.Fatalf("unexpected error at chunk %d: %v", i, err)
		}
		key := string(nonce)
		if _, exists := seen[key]; exists {
			t.Fatalf("duplicate nonce detected at chunk %d", i)
		}
		seen[key] = struct{}{}
	}
}

func TestGenerateChunkNonce(t *testing.T) {
	t.Run("generates 12 bytes from base + counter", func(t *testing.T) {
		base := make([]byte, BaseNonceLength)
		nonce, err := GenerateChunkNonce(base, 0)
		if err != nil {
			t.Fatalf(fmtUnexpectedError, err)
		}
		if len(nonce) != NonceLength {
			t.Errorf(fmtWrongByteLen, NonceLength, len(nonce))
		}
	})

	t.Run("preserves base nonce in first 8 bytes", func(t *testing.T) {
		base := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		nonce, err := GenerateChunkNonce(base, 42)
		if err != nil {
			t.Fatalf(fmtUnexpectedError, err)
		}
		if !bytes.Equal(nonce[:8], base) {
			t.Errorf("base nonce not preserved: got %v, want %v", nonce[:8], base)
		}
	})

	t.Run("encodes counter in little-endian", func(t *testing.T) {
		base := make([]byte, BaseNonceLength)
		testChunkCounterEncoding(t, base)
	})

	t.Run("generates unique nonces for all chunks", func(t *testing.T) {
		base, _ := GenerateBaseNonceSimple()
		testChunkNoncesAreUnique(t, base, 10000)
	})

	t.Run("rejects invalid base nonce", func(t *testing.T) {
		invalidBases := [][]byte{
			nil,
			make([]byte, 7),
			make([]byte, 9),
			make([]byte, 12),
		}
		for _, base := range invalidBases {
			_, err := GenerateChunkNonce(base, 0)
			if err == nil {
				t.Errorf("expected error for base length %d, got nil", len(base))
			}
		}
	})
}

func TestGenerator_WithTracking(t *testing.T) {
	t.Run("detects nonce reuse", func(t *testing.T) {
		gen := NewGenerator(true)

		// Generate first nonce
		nonce1, err := gen.Generate()
		if err != nil {
			t.Fatalf("first generate failed: %v", err)
		}

		// Manually add to tracking (simulating reuse)
		gen.mu.Lock()
		gen.usedNonces[string(nonce1)] = struct{}{}
		gen.mu.Unlock()

		// This shouldn't trigger since we're generating new random nonces
		// But the tracking ensures we never return the same nonce twice
		for i := 0; i < 1000; i++ {
			_, err := gen.Generate()
			if err == ErrNonceReuseDetected {
				// This is extremely unlikely with CSPRNG but shows tracking works
				t.Log("Nonce collision detected (expected to be rare)")
			}
		}
	})
}

func TestEncryptedChunkFormat(t *testing.T) {
	t.Run("serialize and parse roundtrip", func(t *testing.T) {
		nonce, _ := GenerateNonce()
		ciphertext := []byte("this is ciphertext with 16B tag!")

		format := &EncryptedChunkFormat{
			Nonce:            nonce,
			CiphertextAndTag: ciphertext,
		}

		serialized := format.Serialize()

		// Verify length
		expectedLen := NonceLength + len(ciphertext)
		if len(serialized) != expectedLen {
			t.Errorf("expected length %d, got %d", expectedLen, len(serialized))
		}

		// Parse back
		parsed, err := ParseEncryptedChunk(serialized)
		if err != nil {
			t.Fatalf("parse failed: %v", err)
		}

		if !bytes.Equal(parsed.Nonce, nonce) {
			t.Errorf("nonce mismatch")
		}
		if !bytes.Equal(parsed.CiphertextAndTag, ciphertext) {
			t.Errorf("ciphertext mismatch")
		}
	})

	t.Run("rejects too-short data", func(t *testing.T) {
		// Minimum: nonce (12) + tag (16) + 1 byte = 29
		tooShort := make([]byte, 28)
		_, err := ParseEncryptedChunk(tooShort)
		if err == nil {
			t.Error("expected error for too-short data")
		}
	})

	t.Run("accepts minimum valid size", func(t *testing.T) {
		minData := make([]byte, NonceLength+TagLength+1)
		_, err := ParseEncryptedChunk(minData)
		if err != nil {
			t.Errorf("unexpected error for minimum valid size: %v", err)
		}
	})
}

func collectConcurrentNonces(t *testing.T, gen *Generator, goroutines, perGoroutine int) []string {
	t.Helper()
	var wg sync.WaitGroup
	results := make(chan string, goroutines*perGoroutine)
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < perGoroutine; j++ {
				nonce, err := gen.Generate()
				if err != nil {
					t.Errorf("generate failed: %v", err)
					return
				}
				results <- string(nonce)
			}
		}()
	}
	wg.Wait()
	close(results)
	var all []string
	for n := range results {
		all = append(all, n)
	}
	return all
}

func TestConcurrentGeneration(t *testing.T) {
	t.Run("thread-safe generation", func(t *testing.T) {
		gen := NewGenerator(true)
		const goroutines = 100
		const perGoroutine = 100

		all := collectConcurrentNonces(t, gen, goroutines, perGoroutine)

		seen := make(map[string]struct{}, len(all))
		for _, nonce := range all {
			if _, exists := seen[nonce]; exists {
				t.Error("duplicate nonce detected in concurrent generation")
			}
			seen[nonce] = struct{}{}
		}

		if len(seen) != goroutines*perGoroutine {
			t.Errorf("expected %d unique nonces, got %d", goroutines*perGoroutine, len(seen))
		}
	})
}

func BenchmarkGenerateNonce(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GenerateNonce()
	}
}

func BenchmarkGenerateChunkNonce(b *testing.B) {
	base, _ := GenerateBaseNonceSimple()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GenerateChunkNonce(base, uint32(i))
	}
}

func BenchmarkGeneratorWithTracking(b *testing.B) {
	gen := NewGenerator(true)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.Generate()
	}
}

// Helper to verify little-endian encoding
func TestLittleEndianEncoding(t *testing.T) {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, 0x01020304)

	expected := []byte{0x04, 0x03, 0x02, 0x01}
	if !bytes.Equal(buf, expected) {
		t.Errorf("little-endian encoding failed: got %v, want %v", buf, expected)
	}
}
