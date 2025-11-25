import sodium from 'libsodium-wrappers-sumo';

// Configuration OWASP / Libsodium pour Argon2id
const SALT_LENGTH = 16; // 16 octets minimum
const IV_LENGTH = 12;   // 12 octets pour AES-GCM

// Paramètres Argon2id ajustés pour le navigateur (compromis sécurité/UX)
// OWASP recommande m=64 à 128MB, t=4, p=4 pour le côté client si possible.
const ARGON2_MEMLIMIT = 64 * 1024 * 1024; // 64 MB de RAM
const ARGON2_OPSLIMIT = 4; // 4 passes, recommandation OSWASP

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