// ============================================================================
// AES-GCM Crypto Worker - NIST SP 800-38D Compliant
// ============================================================================

// Configuration - NIST SP 800-38D Section 5.2.1.1
const NONCE_LENGTH = 12;      // 96 bits - recommended for AES-GCM
const TAG_LENGTH_BITS = 128;  // 128 bits - maximum security
const TAG_LENGTH_BYTES = 16;

// Per-worker nonce tracking to detect reuse attempts (defense in depth)
const usedNonces = new Set();
const MAX_NONCE_CACHE = 10000; // Prevent memory exhaustion

self.onmessage = async (e) => {
  const { type, fileChunk, key, chunkIndex, baseNonce } = e.data;

  try {
    if (type === 'ENCRYPT') {
      const result = await encryptChunk(fileChunk, key, chunkIndex, baseNonce); // NOSONAR - async function returns a Promise
      self.postMessage({
        type: 'ENCRYPT_SUCCESS',
        encryptedChunk: result,
        chunkIndex
      }, [result]);

    } else if (type === 'DECRYPT') {
      const result = await decryptChunk(fileChunk, key); // NOSONAR - async function returns a Promise
      self.postMessage({ 
        type: 'DECRYPT_SUCCESS', 
        decryptedChunk: result, 
        chunkIndex 
      }, [result]);
    }
  } catch (error) {
    self.postMessage({ type: 'ERROR', error: error.message, chunkIndex });
  }
};

/**
 * Generates a deterministic nonce from base nonce + chunk index.
 * Structure: [8 bytes base nonce] + [4 bytes chunk index (LE)]
 * 
 * @param {Uint8Array} baseNonce - 8-byte random base
 * @param {number} chunkIndex - Chunk counter
 * @returns {Uint8Array} 12-byte nonce
 */
function generateChunkNonce(baseNonce, chunkIndex) {
  const nonce = new Uint8Array(NONCE_LENGTH);
  
  if (baseNonce && baseNonce.length === 8) {
    // Deterministic: base nonce + counter
    nonce.set(new Uint8Array(baseNonce), 0);
    const view = new DataView(nonce.buffer, 8, 4);
    view.setUint32(0, chunkIndex, true);
  } else {
    // Fallback: pure random (legacy compatibility)
    crypto.getRandomValues(nonce);
  }
  
  return nonce;
}

/**
 * Encrypts a chunk using AES-GCM with NIST-compliant nonce.
 * Output format: [Nonce (12B)] + [Ciphertext] + [Tag (16B)]
 * 
 * @param {ArrayBuffer} chunk - Plaintext chunk
 * @param {CryptoKey} key - AES-GCM key
 * @param {number} chunkIndex - Chunk index for deterministic nonce
 * @param {Uint8Array} baseNonce - 8-byte base nonce (optional)
 * @returns {ArrayBuffer} Encrypted chunk with nonce prepended
 */
async function encryptChunk(chunk, key, chunkIndex, baseNonce) {
  // Generate nonce
  const nonce = generateChunkNonce(baseNonce, chunkIndex);
  
  // Defense in depth: check for nonce reuse within this worker session
  const nonceHex = Array.from(nonce).map(b => b.toString(16).padStart(2, '0')).join('');
  if (usedNonces.has(nonceHex)) {
    throw new Error('CRITICAL: Nonce reuse detected! Aborting encryption.');
  }
  
  // Track nonce (with memory limit)
  if (usedNonces.size >= MAX_NONCE_CACHE) {
    usedNonces.clear(); // Reset on overflow (acceptable for defense-in-depth)
  }
  usedNonces.add(nonceHex);

  // Encrypt with AES-GCM
  const encryptedContent = await crypto.subtle.encrypt(
    { 
      name: "AES-GCM", 
      iv: nonce, 
      tagLength: TAG_LENGTH_BITS 
    },
    key,
    chunk
  );

  // Serialize: [Nonce (12B)] + [Ciphertext + Tag]
  const combined = new Uint8Array(NONCE_LENGTH + encryptedContent.byteLength);
  combined.set(nonce, 0);
  combined.set(new Uint8Array(encryptedContent), NONCE_LENGTH);

  return combined.buffer;
}

/**
 * Decrypts a chunk using AES-GCM.
 * Input format: [Nonce (12B)] + [Ciphertext] + [Tag (16B)]
 * 
 * @param {ArrayBuffer} encryptedChunk - Encrypted chunk with nonce
 * @param {CryptoKey} key - AES-GCM key
 * @returns {ArrayBuffer} Decrypted plaintext
 */
async function decryptChunk(encryptedChunk, key) {
  const data = new Uint8Array(encryptedChunk);
  
  // Validate minimum size: nonce + tag (AES-GCM is valid with 0 bytes of ciphertext)
  const MIN_SIZE = NONCE_LENGTH + TAG_LENGTH_BYTES;
  if (data.length < MIN_SIZE) {
    throw new Error(`Invalid encrypted chunk: too short (${data.length} < ${MIN_SIZE})`);
  }

  // Extract nonce and ciphertext
  const nonce = data.slice(0, NONCE_LENGTH);
  const ciphertextWithTag = data.slice(NONCE_LENGTH);

  // Decrypt with AES-GCM (tag verification is automatic)
  const decryptedContent = await crypto.subtle.decrypt(
    { 
      name: "AES-GCM", 
      iv: nonce, 
      tagLength: TAG_LENGTH_BITS 
    },
    key,
    ciphertextWithTag
  );

  return decryptedContent;
}