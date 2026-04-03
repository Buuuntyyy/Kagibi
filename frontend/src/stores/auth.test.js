import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'

// ── Mock key material returned by the registration worker ───────────────────

const mockRecoveryCode = 'deadbeef'.repeat(8)
const mockMasterKeyRaw = new Uint8Array(32).buffer // 32-byte ArrayBuffer of zeros

const mockKeyMaterial = {
  saltHex: 'aabbccddeeff00112233445566778899',
  wrappedMasterKey: 'wrapped-master-key-base64',
  wrappedMasterKeyRecovery: 'wrapped-master-key-recovery-base64',
  recoveryHash: 'recovery-hash-hex',
  recoveryCode: mockRecoveryCode,
  publicKeyPEM: '-----BEGIN PUBLIC KEY-----\nMOCK\n-----END PUBLIC KEY-----',
  encryptedPrivateKey: 'encrypted-private-key-base64',
  masterKeyRaw: mockMasterKeyRaw,
}

// ── Mock Worker — simulates registration.worker.js ──────────────────────────

const postMessageSpy = vi.fn()

class MockRegistrationWorker {
  constructor() {
    this.onmessage = null
    this.onerror = null
  }
  postMessage({ type }) {
    postMessageSpy({ type })
    if (type === 'REGISTER_KEYS') {
      setTimeout(() => {
        this.onmessage?.({ data: { type: 'REGISTER_KEYS_RESULT', payload: mockKeyMaterial } })
      }, 0)
    }
  }
  terminate() {}
}

// ── Module mocks ─────────────────────────────────────────────────────────────

vi.mock('../api', () => ({
  default: {
    post: vi.fn().mockResolvedValue({ data: {} }),
    get: vi.fn().mockResolvedValue({ data: { id: 'user-1', name: 'Test' } }),
  },
}))

vi.mock('../auth-client', () => ({
  authClient: {
    signUp: vi.fn().mockResolvedValue({
      data: { session: { access_token: 'mock-access-token' } },
      error: null,
    }),
    signIn: vi.fn(),
    signOut: vi.fn(),
    getSession: vi.fn().mockResolvedValue({ data: { session: null } }),
    getToken: vi.fn().mockResolvedValue('mock-token'),
    isMFASupported: false,
  },
  IS_POCKETBASE: false,
}))

vi.mock('../router', () => ({ default: { push: vi.fn() } }))
vi.mock('./friends', () => ({ useFriendStore: () => ({ cleanup: vi.fn() }) }))
vi.mock('../utils/useMFA', () => ({ useMFA: () => ({ isMFARequired: vi.fn().mockResolvedValue(false) }) }))

vi.mock('libsodium-wrappers-sumo', () => ({
  default: {
    ready: Promise.resolve(),
    to_hex: vi.fn(() => 'mockhex'),
    from_hex: vi.fn(() => new Uint8Array(16)),
    to_base64: vi.fn(() => 'mockbase64'),
    from_base64: vi.fn(() => new Uint8Array(28)),
    from_string: vi.fn(() => new Uint8Array(4)),
    crypto_hash_sha256: vi.fn(() => new Uint8Array(32)),
    crypto_pwhash: vi.fn(() => new Uint8Array(32)),
    crypto_pwhash_ALG_ARGON2ID13: 2,
  },
}))

vi.mock('../utils/crypto', () => ({
  deriveKeyFromPassword: vi.fn().mockResolvedValue({}),
  generateSalt: vi.fn(() => new Uint8Array(16)),
  wrapMasterKey: vi.fn().mockResolvedValue('wrapped-key'),
  unwrapMasterKey: vi.fn().mockResolvedValue({}),
  hashRecoveryCode: vi.fn().mockResolvedValue('hash'),
  deriveKeyFromRecoveryCode: vi.fn().mockResolvedValue({}),
  generateRSAKeyPair: vi.fn().mockResolvedValue({ publicKey: {}, privateKey: {} }),
  exportKeyToPEM: vi.fn().mockResolvedValue('-----BEGIN PUBLIC KEY-----\nMOCK\n-----END PUBLIC KEY-----'),
  importKeyFromPEM: vi.fn().mockResolvedValue({}),
  encryptPrivateKey: vi.fn().mockResolvedValue('encrypted-priv'),
  decryptPrivateKey: vi.fn().mockResolvedValue({}),
}))

// ── Tests ─────────────────────────────────────────────────────────────────────

describe('auth store — register()', () => {
  let originalWorker
  let importKeySpy

  beforeEach(() => {
    setActivePinia(createPinia())
    originalWorker = globalThis.Worker
    globalThis.Worker = MockRegistrationWorker
    postMessageSpy.mockClear()

    // Spy on crypto.subtle.importKey so we can mock the raw→CryptoKey re-import
    importKeySpy = vi.spyOn(globalThis.crypto.subtle, 'importKey').mockResolvedValue(
      { type: 'secret', algorithm: { name: 'AES-GCM', length: 256 }, extractable: true }
    )
  })

  afterEach(() => {
    globalThis.Worker = originalWorker
    importKeySpy.mockRestore()
    vi.clearAllMocks()
  })

  it('dispatches REGISTER_KEYS to the worker and returns the recovery code', async () => {
    const { useAuthStore } = await import('./auth')
    const store = useAuthStore()

    const code = await store.register('TestUser', 'test@example.com', 'TestPassword!1Secure')

    // Worker received REGISTER_KEYS message
    expect(postMessageSpy).toHaveBeenCalledWith(expect.objectContaining({ type: 'REGISTER_KEYS' }))

    // Returns recovery code from worker payload
    expect(code).toBe(mockRecoveryCode)
  })

  it('sends only opaque key blobs to the backend — no plaintext secrets', async () => {
    const api = (await import('../api')).default
    const { useAuthStore } = await import('./auth')
    const store = useAuthStore()

    await store.register('TestUser', 'test@example.com', 'TestPassword!1Secure')

    const registerCall = api.post.mock.calls.find(([url]) => url === '/auth/register')
    expect(registerCall).toBeTruthy()

    const body = registerCall[1]

    // Key blobs present
    expect(body.salt).toBe(mockKeyMaterial.saltHex)
    expect(body.encrypted_master_key).toBe(mockKeyMaterial.wrappedMasterKey)
    expect(body.encrypted_master_key_recovery).toBe(mockKeyMaterial.wrappedMasterKeyRecovery)
    expect(body.recovery_hash).toBe(mockKeyMaterial.recoveryHash)
    expect(body.public_key).toBe(mockKeyMaterial.publicKeyPEM)
    expect(body.encrypted_private_key).toBe(mockKeyMaterial.encryptedPrivateKey)

    // Zero-knowledge: no plaintext secrets
    expect(body).not.toHaveProperty('password')
    expect(body).not.toHaveProperty('recovery_code')
    expect(body).not.toHaveProperty('master_key')
  })

  it('throws when the auth provider signup returns an error', async () => {
    const { authClient } = await import('../auth-client')
    authClient.signUp.mockResolvedValueOnce({
      data: { session: null },
      error: new Error('Email already registered'),
    })

    const { useAuthStore } = await import('./auth')
    const store = useAuthStore()

    await expect(
      store.register('TestUser', 'existing@example.com', 'TestPassword!1Secure')
    ).rejects.toThrow('Email already registered')
  })
})
