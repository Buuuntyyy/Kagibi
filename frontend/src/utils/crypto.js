import sodium from 'libsodium-wrappers-sumo';

// Configuration OWASP / Libsodium pour Argon2id
const SALT_LENGTH = 16; // 16 octets minimum
const IV_LENGTH = 12;   // 12 octets pour AES-GCM

// Paramètres Argon2id ajustés pour le navigateur (compromis sécurité/UX)
// OWASP recommande m=64 à 128MB, t=4, p=4 pour le côté client si possible.
const ARGON2_MEMLIMIT = 64 * 1024 * 1024; // 64 MB de RAM
const ARGON2_OPSLIMIT = 4; // 4 passes, recommandation OSWASP

export const CHUNK_SIZE = 1 * 1024 * 1024; // 1 MB par chunk pour le traitement en worker
export const ENCRYPTED_CHUNK_SIZE = CHUNK_SIZE + SALT_LENGTH + IV_LENGTH; // Taille estimée après chiffrement

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
        false,
        ["encrypt", "decrypt"]
    );
}

/**
 * Fonction helper pour traiter un chunk via le Worker
 */
function processChunkInWorker(type, chunk, key, chunkIndex) {
    return new Promise((resolve, reject) => {
        const worker = new Worker(new URL('../workers/crypto.worker.js', import.meta.url), { type: 'module' });

        const timeoutId = setTimeout(() => {
            worker.terminate();
            reject(new Error('Le traitement du chunk a expiré.'));
        }, 5000); // 5 secondes timeout

        worker.onmessage = (e) => {
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
            reject(err);
            worker.terminate();
        }

        worker.postMessage({ type, fileChunk: chunk, key, chunkIndex }, [chunk]);
    });
}

/**
 * Chiffre un morceau de fichier via Web Worker
 */
export async function encryptChunkWorker(chunkArrayBuffer, key, chunkIndex) {
    const encryptedBuffer = await processChunkInWorker('ENCRYPT', chunkArrayBuffer, key, chunkIndex);
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