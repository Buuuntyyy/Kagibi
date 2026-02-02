import sodium from 'libsodium-wrappers-sumo';

// ============================================================================
// NIST SP 800-38D / ANSSI Compliant AES-GCM Configuration
// ============================================================================

// Nonce (IV) Configuration - NIST SP 800-38D Section 8.2
// 96 bits (12 bytes) is the recommended size to avoid GHASH overhead
export const NONCE_LENGTH = 12;
export const IV_LENGTH = NONCE_LENGTH; // Alias for backward compatibility

// Authentication Tag - 128 bits as per NIST recommendation
export const TAG_LENGTH_BITS = 128;
export const TAG_LENGTH_BYTES = 16;

// Salt for Argon2id key derivation
const SALT_LENGTH = 16;

// Paramètres Argon2id ajustés pour le navigateur (compromis sécurité/UX)
// OWASP recommande m=64 à 128MB, t=4, p=4 pour le côté client si possible.
const ARGON2_MEMLIMIT = 64 * 1024 * 1024; // 64 MB de RAM
const ARGON2_OPSLIMIT = 4; // 4 passes, recommandation OSWASP

export const CHUNK_SIZE = 10 * 1024 * 1024; // 10 MB par chunk pour le traitement en worker
// Encrypted chunk: [Nonce 12B] + [Ciphertext] + [Tag 16B included in WebCrypto output]
export const ENCRYPTED_CHUNK_SIZE = CHUNK_SIZE + NONCE_LENGTH + TAG_LENGTH_BYTES;

// ============================================================================
// NONCE GENERATION - CSPRNG Only (NIST SP 800-38D Section 8.2.2)
// ============================================================================

/**
 * Generates a cryptographically secure 96-bit nonce using CSPRNG.
 * 
 * Strategy: Random-based approach (NIST SP 800-38D Option 2)
 * 
 * Justification:
 * - Birthday Paradox limit: 2^48 encryptions before 50% collision probability
 * - At 10MB chunks, this allows ~2.8 Exabytes per key before rotation needed
 * - For cloud storage: key-per-file strategy ensures we never approach this limit
 * - Random approach chosen over counter for stateless operation (no persistence needed)
 * 
 * @returns {Uint8Array} 12-byte cryptographically random nonce
 */
export function generateNonce() {
    // CRITICAL: Only use CSPRNG - never Math.random()
    const nonce = new Uint8Array(NONCE_LENGTH);
    crypto.getRandomValues(nonce);
    return nonce;
}

/**
 * Generates a deterministic nonce for chunk-based encryption.
 * Combines random base nonce with chunk index to guarantee uniqueness within a file.
 * 
 * Structure: [8 bytes random base] + [4 bytes chunk counter (little-endian)]
 * 
 * This ensures:
 * - Uniqueness across files (random base)
 * - Uniqueness across chunks within a file (counter)
 * - No nonce reuse even with 2^32 chunks per file (40 PB at 10MB chunks)
 * 
 * @param {Uint8Array} baseNonce - 8-byte random base nonce (generated once per file)
 * @param {number} chunkIndex - Chunk index (0-based)
 * @returns {Uint8Array} 12-byte deterministic nonce
 */
export function generateChunkNonce(baseNonce, chunkIndex) {
    if (!(baseNonce instanceof Uint8Array) || baseNonce.length !== 8) {
        throw new Error('baseNonce must be a Uint8Array of exactly 8 bytes');
    }
    if (!Number.isInteger(chunkIndex) || chunkIndex < 0 || chunkIndex > 0xFFFFFFFF) {
        throw new Error('chunkIndex must be a non-negative 32-bit integer');
    }
    
    const nonce = new Uint8Array(NONCE_LENGTH);
    // First 8 bytes: random base (unique per file/key)
    nonce.set(baseNonce, 0);
    // Last 4 bytes: chunk counter (little-endian for consistency)
    const counterView = new DataView(nonce.buffer, 8, 4);
    counterView.setUint32(0, chunkIndex, true); // little-endian
    
    return nonce;
}

/**
 * Generates an 8-byte base nonce for chunked file encryption.
 * @returns {Uint8Array} 8-byte random base nonce
 */
export function generateBaseNonce() {
    const baseNonce = new Uint8Array(8);
    crypto.getRandomValues(baseNonce);
    return baseNonce;
}

// ============================================================================
// ENCRYPTED DATA SERIALIZATION - NIST Compliant Structure
// ============================================================================

/**
 * Serializes encrypted chunk data into NIST-compliant format:
 * [Nonce (12 bytes)] + [Ciphertext + Auth Tag]
 * 
 * Note: WebCrypto AES-GCM appends the auth tag to ciphertext automatically
 * 
 * @param {Uint8Array} nonce - 12-byte nonce
 * @param {ArrayBuffer} ciphertextWithTag - Encrypted data with auth tag appended
 * @returns {ArrayBuffer} Serialized encrypted chunk
 */
export function serializeEncryptedChunk(nonce, ciphertextWithTag) {
    const cipherArray = new Uint8Array(ciphertextWithTag);
    const combined = new Uint8Array(NONCE_LENGTH + cipherArray.length);
    combined.set(nonce, 0);
    combined.set(cipherArray, NONCE_LENGTH);
    return combined.buffer;
}

/**
 * Deserializes encrypted chunk data.
 * 
 * @param {ArrayBuffer} encryptedData - Serialized encrypted chunk
 * @returns {{nonce: Uint8Array, ciphertextWithTag: Uint8Array}} Parsed components
 * @throws {Error} If data is too short to contain valid encrypted content
 */
export function deserializeEncryptedChunk(encryptedData) {
    const data = new Uint8Array(encryptedData);
    
    // Minimum size: nonce (12) + tag (16) + at least 1 byte ciphertext
    const MIN_SIZE = NONCE_LENGTH + TAG_LENGTH_BYTES + 1;
    if (data.length < MIN_SIZE) {
        throw new Error(`Encrypted data too short: ${data.length} bytes (minimum: ${MIN_SIZE})`);
    }
    
    return {
        nonce: data.slice(0, NONCE_LENGTH),
        ciphertextWithTag: data.slice(NONCE_LENGTH)
    };
}

/**
 * Génère une clé maître AES-GCM 256 bits.
 */
export async function generateMasterKey() {
    return window.crypto.subtle.generateKey(
        {
            name: "AES-GCM",
            length: 256
        },
        true,
        ["encrypt", "decrypt"]
    );
}

/**
 * Emballe (chiffre) la clé maître avec la KEK dérivée du mot de passe.
 */
export async function wrapMasterKey(masterKey, kek) {
    const rawKeyData = await window.crypto.subtle.exportKey("raw", masterKey);
    const iv = window.crypto.getRandomValues(new Uint8Array(IV_LENGTH));
    const encryptedKeyBuffer = await window.crypto.subtle.encrypt(
        {
            name: "AES-GCM",
            iv: iv,
        },
        kek,
        rawKeyData
    );

    const combined = new Uint8Array(iv.byteLength + encryptedKeyBuffer.byteLength);
    combined.set(iv);
    combined.set(new Uint8Array(encryptedKeyBuffer), iv.byteLength);

    return sodium.to_base64(combined);
}

/** 
 * Déballe (déchiffre) la clé maître avec la KEK dérivée du mot de passe.
 */
export async function unwrapMasterKey(wrappedKeyBase64, kek) {
    const combined = sodium.from_base64(wrappedKeyBase64);
    const iv = combined.slice(0, IV_LENGTH);
    const encryptedKeyData = combined.slice(IV_LENGTH);

    const rawKeyData = await window.crypto.subtle.decrypt(
        {
            name: "AES-GCM",
            iv: iv,
        },
        kek,
        encryptedKeyData
    );

    return window.crypto.subtle.importKey(
        "raw",
        rawKeyData,
        { name: "AES-GCM" },
        true,
        ["encrypt", "decrypt"]
    );
}

/**
 * Fonction helper pour traiter un chunk via le Worker
 * @param {string} type - 'ENCRYPT' ou 'DECRYPT'
 * @param {ArrayBuffer} chunk - Data to process
 * @param {CryptoKey} key - AES-GCM key
 * @param {number} chunkIndex - Chunk index
 * @param {Uint8Array} baseNonce - 8-byte base nonce (required for ENCRYPT)
 */
function processChunkInWorker(type, chunk, key, chunkIndex, baseNonce = null) {
    return new Promise((resolve, reject) => {
        const worker = new Worker(new URL('../workers/crypto.worker.js', import.meta.url), { type: 'module' });

        const timeoutId = setTimeout(() => {
            worker.terminate();
            reject(new Error('Le traitement du chunk a expiré.'));
        }, 30000); // 30 secondes timeout (increased for large chunks)

        worker.onmessage = (e) => {
            clearTimeout(timeoutId);
            const { type: msgType, encryptedChunk, decryptedChunk, error } = e.data;

            if (msgType === 'ERROR') {
                reject(new Error(error));
            } else if (msgType === 'ENCRYPT_SUCCESS' && type === 'ENCRYPT') {
                resolve(encryptedChunk);
            } else if (msgType === 'DECRYPT_SUCCESS' && type === 'DECRYPT') {
                resolve(decryptedChunk);
            }
            worker.terminate();// Important : tuer le worker après usage pour libérer la mémoire
        };

        worker.onerror = (err) => {
            clearTimeout(timeoutId);
            reject(new Error(err.message));
            worker.terminate();
        }

        // Transfer chunk ownership for zero-copy performance
        const message = { type, fileChunk: chunk, key, chunkIndex };
        if (baseNonce) {
            message.baseNonce = baseNonce;
        }
        worker.postMessage(message, [chunk]);
    });
}

/**
 * Chiffre un morceau de fichier via Web Worker avec NIST-compliant nonce.
 * @param {ArrayBuffer} chunkArrayBuffer - Plaintext chunk
 * @param {CryptoKey} key - AES-GCM key
 * @param {number} chunkIndex - Chunk index
 * @param {Uint8Array} baseNonce - 8-byte base nonce (generated once per file)
 * @returns {Promise<Blob>} Encrypted chunk as Blob
 */
export async function encryptChunkWorker(chunkArrayBuffer, key, chunkIndex, baseNonce = null) {
    // Generate base nonce if not provided (backward compatibility)
    const effectiveBaseNonce = baseNonce || generateBaseNonce();
    const encryptedBuffer = await processChunkInWorker('ENCRYPT', chunkArrayBuffer, key, chunkIndex, effectiveBaseNonce);
    return new Blob([encryptedBuffer], { type: 'application/octet-stream' });
}

/**
 * Déchiffre un morceau de fichier via Web Worker
 */
export async function decryptChunkWorker(encryptedChunkBuffer, key, chunkIndex) {
    const decryptedBuffer = await processChunkInWorker('DECRYPT', encryptedChunkBuffer, key, chunkIndex);
    return decryptedBuffer;
}

/**
 * Déchiffre un fichier complet (composé de chunks) via Worker
 * Utilisé pour le téléchargement final
 */
export async function decryptChunkedFileWorker(encryptedBlob, key, mimeType) {
    const totalSize = encryptedBlob.size;
    let offset = 0;
    const decryptedParts = [];
    let chunkIndex = 0;
    // On pourrait paralléliser pour lancer plusieurs workers en même temps
    while (offset < totalSize) {
        const currentChunkSize = Math.min(ENCRYPTED_CHUNK_SIZE, totalSize - offset);
        const chunkBlob = encryptedBlob.slice(offset, offset + currentChunkSize);
        const chunkBuffer = await chunkBlob.arrayBuffer();

        if (chunkBuffer.byteLength < IV_LENGTH) break; // Sécurité

        const decryptedPart = await decryptChunkWorker(chunkBuffer, key, chunkIndex);
        decryptedParts.push(decryptedPart);

        offset += currentChunkSize;
        chunkIndex ++;
    }

    return new Blob(decryptedParts, { type: mimeType || 'application/octet-stream' });
}


/**
 * Dérive une clé AES-GCM 256 bits à partir d'un mot de passe et d'un sel via Argon2id.
 */
export async function deriveKeyFromPassword(password, salt) {
  await sodium.ready;

  if (!salt || salt.length < SALT_LENGTH) {
    throw new Error(`Le sel doit faire au moins ${SALT_LENGTH} octets.`);
  }

  // Utilisation explicite de Argon2id via libsodium
  const keyBytes = sodium.crypto_pwhash(
    32, // Longueur de la clé (32 octets = 256 bits)
    password,
    salt,
    ARGON2_OPSLIMIT,
    ARGON2_MEMLIMIT,
    sodium.crypto_pwhash_ALG_ARGON2ID13 // Algorithme Argon2id v1.3
  );

  // Importation de la clé brute dans l'API Web Crypto pour AES-GCM
  return window.crypto.subtle.importKey(
    "raw",
    keyBytes,
    { name: "AES-GCM" },
    false,
    ["encrypt", "decrypt"]
  );
}

/**
 * Chiffre un fichier (Blob/File) avec AES-GCM.
 */
export async function encryptFile(file, key) {
  try {
    const iv = window.crypto.getRandomValues(new Uint8Array(IV_LENGTH));
    const arrayBuffer = await file.arrayBuffer();

    const encryptedContent = await window.crypto.subtle.encrypt(
      {
        name: "AES-GCM",
        iv: iv,
        tagLength: 128 // Tag d'authentification standard (128 bits)
      },
      key,
      arrayBuffer
    );

    // Concaténation : IV + Contenu Chiffré
    const combinedBuffer = new Uint8Array(iv.byteLength + encryptedContent.byteLength);
    combinedBuffer.set(iv);
    combinedBuffer.set(new Uint8Array(encryptedContent), iv.byteLength);

    return new Blob([combinedBuffer], { type: 'application/octet-stream' });
  } catch (error) {
    console.error("Erreur critique de chiffrement:", error);
    throw new Error("Le chiffrement du fichier a échoué. Opération annulée par sécurité.");
  }
}

/**
 * Déchiffre un Blob avec AES-GCM.
 */
export async function decryptFile(encryptedBlob, key, mimeType) {
  const safeMimeType = validateMimeType(mimeType);

  try {
    const arrayBuffer = await encryptedBlob.arrayBuffer();
    
    if (arrayBuffer.byteLength < IV_LENGTH) {
      throw new Error("Fichier corrompu ou trop court.");
    }

    const iv = arrayBuffer.slice(0, IV_LENGTH);
    const data = arrayBuffer.slice(IV_LENGTH);

    const decryptedContent = await window.crypto.subtle.decrypt(
      {
        name: "AES-GCM",
        iv: new Uint8Array(iv),
        tagLength: 128
      },
      key,
      data
    );

    return new Blob([decryptedContent], { type: safeMimeType });
  } catch (error) {
    console.error("Erreur de déchiffrement:", error);
    throw new Error("Échec du déchiffrement. Le fichier est peut-être corrompu, altéré, ou la clé est incorrecte.");
  }
}

function validateMimeType(mimeType) {
  const dangerousTypes = ['text/html', 'image/svg+xml', 'application/javascript', 'application/x-javascript'];
  if (dangerousTypes.includes(mimeType)) {
    return 'application/octet-stream';
  }
  return mimeType || 'application/octet-stream';
}

export function generateSalt() {
    return window.crypto.getRandomValues(new Uint8Array(SALT_LENGTH));
}

export function generateRecoveryCode() {
    // Generate a random 32-byte hex string
    const bytes = window.crypto.getRandomValues(new Uint8Array(32));
    return sodium.to_hex(bytes);
}

export async function deriveKeyFromRecoveryCode(recoveryCode, salt) {
    // Similar to password derivation but maybe with different params if we wanted
    // For simplicity, we use the same robust Argon2id derivation
    return deriveKeyFromPassword(recoveryCode, salt);
}

export async function hashRecoveryCode(recoveryCode) {
    await sodium.ready;
    // Simple SHA-256 hash for server-side verification (proof of possession)
    // We don't need salt here because the recovery code itself is high entropy
    const msg = sodium.from_string(recoveryCode);
    const hash = sodium.crypto_hash_sha256(msg);
    return sodium.to_hex(hash);
}

export async function deriveKeyFromToken(token) {
    const encoder = new TextEncoder();
    const data = encoder.encode(token);
    const hash = await window.crypto.subtle.digest("SHA-256", data);
    return window.crypto.subtle.importKey(
        "raw",
        hash,
        { name: "AES-GCM" },
        false,
        ["encrypt", "decrypt"]
    );
}

// --- Asymmetric Encryption (RSA-OAEP) Implementation ---

/**
 * Generate a new RSA Key Pair for user identity and sharing
 * @returns {Promise<CryptoKeyPair>}
 */
export async function generateRSAKeyPair() {
    return window.crypto.subtle.generateKey(
        {
            name: "RSA-OAEP",
            modulusLength: 4096, // High security
            publicExponent: new Uint8Array([1, 0, 1]),
            hash: "SHA-256",
        },
        true,
        ["encrypt", "decrypt"]
    );
}

/**
 * Export a key to PEM format string
 */
export async function exportKeyToPEM(key, type = 'spki') {
    const exported = await window.crypto.subtle.exportKey(type, key);
    const exportedAsString = String.fromCodePoint(...new Uint8Array(exported));
    const exportedAsBase64 = btoa(exportedAsString);
    const pemHeader = type === 'spki' ? '-----BEGIN PUBLIC KEY-----' : '-----BEGIN PRIVATE KEY-----';
    const pemFooter = type === 'spki' ? '-----END PUBLIC KEY-----' : '-----END PRIVATE KEY-----';
    
    return `${pemHeader}\n${exportedAsBase64}\n${pemFooter}`;
}

/**
 * Import a PEM formatted key string back to CryptoKey
 */
export async function importKeyFromPEM(pemData, type = 'spki') {
    // Remove headers and newlines
    const pemHeader = type === 'spki' ? '-----BEGIN PUBLIC KEY-----' : '-----BEGIN PRIVATE KEY-----';
    const pemFooter = type === 'spki' ? '-----END PUBLIC KEY-----' : '-----END PRIVATE KEY-----';
    
    // Simple basic cleanup, robust enough for our generated keys
    const pemContents = pemData.replaceAll(/-----BEGIN [A-Z ]+-----/g, "")
                               .replaceAll(/-----END [A-Z ]+-----/g, "")
                               .replaceAll(/\s/g, "");
                               
    const binaryDerString = atob(pemContents);
    const binaryDer = new Uint8Array(binaryDerString.length);
    for (let i = 0; i < binaryDerString.length; i++) {
        binaryDer[i] = binaryDerString.codePointAt(i);
    }

    return window.crypto.subtle.importKey(
        type,
        binaryDer.buffer,
        {
            name: "RSA-OAEP",
            hash: "SHA-256",
        },
        true,
        type === 'spki' ? ["encrypt"] : ["decrypt"]
    );
}

/**
 * Encrypt a symmetric key (raw bytes) with a recipient's Public RSA Key
 */
export async function encryptKeyWithPublicKey(symmetricKeyRaw, publicKey) {
    const encryptedBuffer = await window.crypto.subtle.encrypt(
        {
            name: "RSA-OAEP"
        },
        publicKey,
        symmetricKeyRaw
    );
    return sodium.to_base64(new Uint8Array(encryptedBuffer));
}

/**
 * Decrypt a symmetric key with the user's Private RSA Key
 */
export async function decryptKeyWithPrivateKey(encryptedKeyBase64, privateKey) {
    const encryptedKeyBuffer = sodium.from_base64(encryptedKeyBase64);
    
    const decryptedBuffer = await window.crypto.subtle.decrypt(
        {
            name: "RSA-OAEP"
        },
        privateKey,
        encryptedKeyBuffer
    );
    
    return new Uint8Array(decryptedBuffer);
}

/**
 * Encrypt the RSA Private Key with the User's Master Key (AES-GCM) for storage
 */
export async function encryptPrivateKey(privateKey, masterKey) {
    const exportedPrivate = await window.crypto.subtle.exportKey("pkcs8", privateKey);
    const iv = window.crypto.getRandomValues(new Uint8Array(IV_LENGTH));
    
    const encryptedContent = await window.crypto.subtle.encrypt(
        {
            name: "AES-GCM",
            iv: iv,
        },
        masterKey,
        exportedPrivate
    );

    const combined = new Uint8Array(iv.byteLength + encryptedContent.byteLength);
    combined.set(iv);
    combined.set(new Uint8Array(encryptedContent), iv.byteLength);

    return sodium.to_base64(combined);
}

/**
 * Decrypt the RSA Private Key using the User's Master Key
 */
export async function decryptPrivateKey(encryptedPrivateKeyBase64, masterKey) {
    const combined = sodium.from_base64(encryptedPrivateKeyBase64);
    const iv = combined.slice(0, IV_LENGTH);
    const data = combined.slice(IV_LENGTH);

    const decryptedBuffer = await window.crypto.subtle.decrypt(
        {
            name: "AES-GCM",
            iv: iv,
        },
        masterKey,
        data
    );

    return window.crypto.subtle.importKey(
        "pkcs8",
        decryptedBuffer,
        {
            name: "RSA-OAEP",
            hash: "SHA-256",
        },
        true,
        ["decrypt"]
    );
}
