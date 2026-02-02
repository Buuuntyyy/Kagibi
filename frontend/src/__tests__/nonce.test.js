/**
 * NIST SP 800-38D Nonce Generation Tests
 * Tests for AES-GCM nonce uniqueness and entropy
 */

import { describe, it, expect, beforeAll } from 'vitest';
import { 
    generateNonce, 
    generateBaseNonce, 
    generateChunkNonce,
    NONCE_LENGTH,
    serializeEncryptedChunk,
    deserializeEncryptedChunk,
    TAG_LENGTH_BYTES
} from '../utils/crypto';

describe('NIST SP 800-38D Nonce Generation', () => {
    
    describe('generateNonce', () => {
        it('should generate exactly 12 bytes (96 bits)', () => {
            const nonce = generateNonce();
            expect(nonce).toBeInstanceOf(Uint8Array);
            expect(nonce.length).toBe(NONCE_LENGTH);
            expect(nonce.length).toBe(12);
        });

        it('should use CSPRNG (non-zero entropy)', () => {
            const nonce = generateNonce();
            // At least some bytes should be non-zero
            const hasNonZero = Array.from(nonce).some(b => b !== 0);
            expect(hasNonZero).toBe(true);
        });

        it('should generate unique nonces (10,000 samples)', () => {
            const SAMPLE_SIZE = 10000;
            const nonceSet = new Set();
            
            for (let i = 0; i < SAMPLE_SIZE; i++) {
                const nonce = generateNonce();
                const nonceHex = Array.from(nonce)
                    .map(b => b.toString(16).padStart(2, '0'))
                    .join('');
                
                // Check for duplicates
                expect(nonceSet.has(nonceHex)).toBe(false);
                nonceSet.add(nonceHex);
            }
            
            expect(nonceSet.size).toBe(SAMPLE_SIZE);
        });

        it('should have sufficient entropy distribution', () => {
            const SAMPLE_SIZE = 1000;
            const byteCounts = new Array(256).fill(0);
            
            for (let i = 0; i < SAMPLE_SIZE; i++) {
                const nonce = generateNonce();
                for (const byte of nonce) {
                    byteCounts[byte]++;
                }
            }
            
            // Chi-squared test approximation: each byte value should appear
            // roughly equally (SAMPLE_SIZE * 12 / 256 ≈ 47 times expected)
            const totalBytes = SAMPLE_SIZE * NONCE_LENGTH;
            const expectedPerValue = totalBytes / 256;
            
            // Allow significant variance (±70%) for random distribution
            const minExpected = expectedPerValue * 0.3;
            const maxExpected = expectedPerValue * 1.7;
            
            let outliers = 0;
            for (let i = 0; i < 256; i++) {
                if (byteCounts[i] < minExpected || byteCounts[i] > maxExpected) {
                    outliers++;
                }
            }
            
            // Allow up to 5% outliers (statistical variation)
            expect(outliers).toBeLessThan(256 * 0.05);
        });
    });

    describe('generateBaseNonce', () => {
        it('should generate exactly 8 bytes', () => {
            const baseNonce = generateBaseNonce();
            expect(baseNonce).toBeInstanceOf(Uint8Array);
            expect(baseNonce.length).toBe(8);
        });

        it('should generate unique base nonces (10,000 samples)', () => {
            const SAMPLE_SIZE = 10000;
            const nonceSet = new Set();
            
            for (let i = 0; i < SAMPLE_SIZE; i++) {
                const nonce = generateBaseNonce();
                const nonceHex = Array.from(nonce)
                    .map(b => b.toString(16).padStart(2, '0'))
                    .join('');
                
                expect(nonceSet.has(nonceHex)).toBe(false);
                nonceSet.add(nonceHex);
            }
            
            expect(nonceSet.size).toBe(SAMPLE_SIZE);
        });
    });

    describe('generateChunkNonce', () => {
        it('should generate 12-byte nonce from 8-byte base + 4-byte counter', () => {
            const baseNonce = generateBaseNonce();
            const chunkNonce = generateChunkNonce(baseNonce, 0);
            
            expect(chunkNonce).toBeInstanceOf(Uint8Array);
            expect(chunkNonce.length).toBe(12);
        });

        it('should preserve base nonce in first 8 bytes', () => {
            const baseNonce = generateBaseNonce();
            const chunkNonce = generateChunkNonce(baseNonce, 42);
            
            for (let i = 0; i < 8; i++) {
                expect(chunkNonce[i]).toBe(baseNonce[i]);
            }
        });

        it('should encode chunk index in last 4 bytes (little-endian)', () => {
            const baseNonce = new Uint8Array(8).fill(0);
            
            const nonce0 = generateChunkNonce(baseNonce, 0);
            expect(nonce0[8]).toBe(0);
            expect(nonce0[9]).toBe(0);
            expect(nonce0[10]).toBe(0);
            expect(nonce0[11]).toBe(0);
            
            const nonce1 = generateChunkNonce(baseNonce, 1);
            expect(nonce1[8]).toBe(1);
            expect(nonce1[9]).toBe(0);
            expect(nonce1[10]).toBe(0);
            expect(nonce1[11]).toBe(0);
            
            const nonce256 = generateChunkNonce(baseNonce, 256);
            expect(nonce256[8]).toBe(0);
            expect(nonce256[9]).toBe(1);
            expect(nonce256[10]).toBe(0);
            expect(nonce256[11]).toBe(0);
            
            // Max 32-bit value
            const nonceMax = generateChunkNonce(baseNonce, 0xFFFFFFFF);
            expect(nonceMax[8]).toBe(0xFF);
            expect(nonceMax[9]).toBe(0xFF);
            expect(nonceMax[10]).toBe(0xFF);
            expect(nonceMax[11]).toBe(0xFF);
        });

        it('should generate unique nonces for all chunks of a file', () => {
            const baseNonce = generateBaseNonce();
            const CHUNK_COUNT = 10000;
            const nonceSet = new Set();
            
            for (let i = 0; i < CHUNK_COUNT; i++) {
                const nonce = generateChunkNonce(baseNonce, i);
                const nonceHex = Array.from(nonce)
                    .map(b => b.toString(16).padStart(2, '0'))
                    .join('');
                
                expect(nonceSet.has(nonceHex)).toBe(false);
                nonceSet.add(nonceHex);
            }
            
            expect(nonceSet.size).toBe(CHUNK_COUNT);
        });

        it('should reject invalid baseNonce', () => {
            expect(() => generateChunkNonce(null, 0)).toThrow();
            expect(() => generateChunkNonce(new Uint8Array(7), 0)).toThrow();
            expect(() => generateChunkNonce(new Uint8Array(9), 0)).toThrow();
            expect(() => generateChunkNonce([1,2,3,4,5,6,7,8], 0)).toThrow();
        });

        it('should reject invalid chunkIndex', () => {
            const baseNonce = generateBaseNonce();
            expect(() => generateChunkNonce(baseNonce, -1)).toThrow();
            expect(() => generateChunkNonce(baseNonce, 0x100000000)).toThrow(); // > 32-bit
            expect(() => generateChunkNonce(baseNonce, 1.5)).toThrow();
            expect(() => generateChunkNonce(baseNonce, '0')).toThrow();
        });
    });

    describe('Serialization/Deserialization', () => {
        it('should serialize nonce + ciphertext correctly', () => {
            const nonce = generateNonce();
            const ciphertext = new Uint8Array([1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17]);
            
            const serialized = serializeEncryptedChunk(nonce, ciphertext.buffer);
            const data = new Uint8Array(serialized);
            
            // Total size: 12 (nonce) + 17 (ciphertext) = 29
            expect(data.length).toBe(29);
            
            // First 12 bytes should be nonce
            for (let i = 0; i < 12; i++) {
                expect(data[i]).toBe(nonce[i]);
            }
            
            // Remaining bytes should be ciphertext
            for (let i = 0; i < ciphertext.length; i++) {
                expect(data[12 + i]).toBe(ciphertext[i]);
            }
        });

        it('should deserialize encrypted chunk correctly', () => {
            const nonce = generateNonce();
            const ciphertext = new Uint8Array(32); // 16 bytes ciphertext + 16 bytes tag
            crypto.getRandomValues(ciphertext);
            
            const serialized = serializeEncryptedChunk(nonce, ciphertext.buffer);
            const { nonce: extractedNonce, ciphertextWithTag } = deserializeEncryptedChunk(serialized);
            
            // Verify nonce
            expect(extractedNonce.length).toBe(12);
            for (let i = 0; i < 12; i++) {
                expect(extractedNonce[i]).toBe(nonce[i]);
            }
            
            // Verify ciphertext
            expect(ciphertextWithTag.length).toBe(32);
            for (let i = 0; i < 32; i++) {
                expect(ciphertextWithTag[i]).toBe(ciphertext[i]);
            }
        });

        it('should reject too-short data in deserialization', () => {
            // Minimum: 12 (nonce) + 16 (tag) + 1 (ciphertext) = 29 bytes
            const tooShort = new ArrayBuffer(28);
            expect(() => deserializeEncryptedChunk(tooShort)).toThrow();
        });

        it('should handle minimum valid size', () => {
            const minData = new Uint8Array(12 + 16 + 1); // nonce + tag + 1 byte
            crypto.getRandomValues(minData);
            
            const { nonce, ciphertextWithTag } = deserializeEncryptedChunk(minData.buffer);
            expect(nonce.length).toBe(12);
            expect(ciphertextWithTag.length).toBe(17);
        });
    });

    describe('Nonce Collision Probability Analysis', () => {
        it('should document birthday paradox limits', () => {
            // NIST SP 800-38D Section 8.3
            // For 96-bit random nonces:
            // - 2^32 encryptions: ~10^-9 collision probability (acceptable)
            // - 2^48 encryptions: ~50% collision probability (danger zone)
            
            // At 10MB chunks, 2^32 chunks = 40 Petabytes per key
            // This is acceptable for single-file encryption with key-per-file strategy
            
            const BITS = 96;
            const safeLimit = Math.pow(2, 32); // 4 billion encryptions
            const dangerLimit = Math.pow(2, 48);
            
            // Document the limits
            expect(BITS).toBe(NONCE_LENGTH * 8);
            expect(safeLimit).toBeGreaterThan(4e9);
            expect(dangerLimit).toBeGreaterThan(2.8e14);
        });
    });
});
