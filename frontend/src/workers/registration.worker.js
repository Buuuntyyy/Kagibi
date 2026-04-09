// Polyfill: make window.crypto available in worker context (used by crypto.js)
globalThis.window = globalThis

import sodium from 'libsodium-wrappers-sumo'
import {
  generateSalt,
  generateMasterKey,
  deriveKeyFromPassword,
  wrapMasterKey,
  generateRecoveryCode,
  hashRecoveryCode,
  deriveKeyFromRecoveryCode,
  generateRSAKeyPair,
  exportKeyToPEM,
  encryptPrivateKey,
} from '../utils/crypto.js'

self.onmessage = async ({ data }) => {
  if (data.type !== 'REGISTER_KEYS') return

  try {
    await sodium.ready

    // 1. Generate salt (16-byte random)
    const salt = generateSalt()
    const saltHex = sodium.to_hex(salt)

    // 2. Generate master key (AES-GCM 256-bit)
    const masterKey = await generateMasterKey()

    // 3. Derive KEK from password via Argon2id (the slow step — ~1s)
    const kek = await deriveKeyFromPassword(data.password, salt)

    // 4. Wrap master key with password-derived KEK
    const wrappedMasterKey = await wrapMasterKey(masterKey, kek)

    // 5. Generate recovery code and hash
    const recoveryCode = generateRecoveryCode()
    const recoveryHash = await hashRecoveryCode(recoveryCode)

    // 6. Derive recovery KEK and wrap master key with it
    const recoveryKek = await deriveKeyFromRecoveryCode(recoveryCode, salt)
    const wrappedMasterKeyRecovery = await wrapMasterKey(masterKey, recoveryKek)

    // 7. Generate RSA-OAEP 4096-bit keypair
    const keyPair = await generateRSAKeyPair()
    const publicKeyPEM = await exportKeyToPEM(keyPair.publicKey, 'spki')
    const encryptedPrivateKey = await encryptPrivateKey(keyPair.privateKey, masterKey)

    // 8. Export master key as raw bytes so main thread can re-import as CryptoKey.
    //    CryptoKey structured-clone is not reliable across worker boundaries in all environments.
    const masterKeyRaw = await crypto.subtle.exportKey('raw', masterKey)

    postMessage({
      type: 'REGISTER_KEYS_RESULT',
      payload: {
        saltHex,
        wrappedMasterKey,
        wrappedMasterKeyRecovery,
        recoveryHash,
        recoveryCode,     // plaintext — for display only, NEVER sent to backend
        publicKeyPEM,
        encryptedPrivateKey,
        masterKeyRaw,     // raw bytes; main thread re-imports as CryptoKey
      },
    })
  } catch (err) {
    // Log full error internally; send only a generic message to the main thread
    // to avoid leaking internal state or partial crypto data via error messages.
    console.error('[registration.worker] Key generation failed:', err)
    postMessage({ type: 'REGISTER_KEYS_ERROR', error: 'Erreur lors de la génération des clés.' })
  }
}
