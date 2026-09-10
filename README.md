# Kagibi

**End-to-end encrypted cloud storage, without compromise.**  
**Stockage cloud chiffré de bout en bout, sans compromis.**

---

Read the documentation in your language:

- [English](./docs/README.en.md)
- [Français](./docs/README.fr.md)

---

## Features

### Security & Privacy
- **Zero-knowledge encryption** — files are encrypted on your device before upload; the server never sees plaintext content
- **AES-256-GCM** file encryption, split into 10 MB chunks with unique nonces
- **Argon2id** key derivation (64 MB memory, 4 iterations)
- **RSA-OAEP 4096-bit** key pair per user for asymmetric sharing
- **Optional filename encryption** — file and folder names can be encrypted client-side at registration
- **MFA (TOTP)** two-factor authentication support
- **Account recovery code** — password-independent recovery without server access to your keys

### File Management
- Upload, download, organize files and folders
- Streaming decryption — large files are never fully loaded into memory
- Recently opened files
- File search (disabled when filename encryption is active)

### Sharing
- **P2P transfer** — direct device-to-device file transfer over WebRTC DataChannel, encrypted with AES-GCM; a TURN relay (**no intermediate storage**, only relay) is used when a direct connection is not possible, without compromising encryption.
- **Public link sharing** — optional password protection and expiration (1–30 days)
- **Friend sharing (user-to-user)** — end-to-end encrypted via RSA; only the recipient can decrypt

### Social
- Friend system via unique friend codes (e.g. `#A7KD92XZ`)
- Real-time online presence
- P2P transfer notifications with sound and browser push

### Account & Settings
- Light / dark theme
- Multi-language interface (i18n)
- Usage dashboard (storage, P2P quota)
- GDPR-compliant data export and account deletion (30-day grace period)

---

## Roadmap

> This section lists planned and considered features. It is not a commitment to a release date.

<!-- Add roadmap items below. Example format:
- [ ] Feature name — short description
- [x] Already shipped feature
-->

---

## Installation & self-hosting

**Everything you need — local development, and production deployment with external S3 or fully local storage (Garage) — starts here: [`docs/en/self-hosting/prerequisites.md`](./docs/en/self-hosting/prerequisites.md)** (French: [`docs/self-hosting/prerequisites.md`](./docs/self-hosting/prerequisites.md)), which links onward to storage setup and the reverse-proxy guide.

## License

AGPLv3 — see [`LICENSE`](./LICENSE).
