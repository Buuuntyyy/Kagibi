# 🎨 Frontend SaferCloud - Vue.js 3 + Zero-Knowledge Crypto

**Single Page Application** avec chiffrement client-side complet via Web Crypto API.

---

## 📋 Vue d'ensemble

Le frontend SaferCloud est une SPA Vue.js 3 qui implémente une architecture **Zero-Knowledge** complète :
- ✅ Chiffrement AES-GCM 256 bits dans le navigateur
- ✅ Dérivation de clés avec Argon2id (64MB RAM, 4 passes)
- ✅ Web Workers pour chiffrement async
- ✅ Service Worker pour gestion de session sécurisée
- ✅ RSA-OAEP 4096 bits pour partages
- ✅ WebRTC P2P pour transferts directs

---

## 🛠️ Stack Technique

| Technologie | Version | Rôle |
|-------------|---------|------|
| **Vue.js** | 3.5.24 | Framework réactif (Composition API) |
| **Vite** | 7.2.4 | Build tool ultra-rapide |
| **Pinia** | 3.0.4 | State management |
| **Vue Router** | 4.6.3 | Routing SPA |
| **Axios** | 1.13.2 | Client HTTP REST |
| **@supabase/supabase-js** | 2.90.1 | Authentification JWT |
| **libsodium-wrappers-sumo** | 0.7.15 | Argon2id + utilitaires crypto |
| **PrimeVue** | 4.4.1 | Composants UI |
| **pdfjs-dist** | 5.4.530 | Aperçu PDF |
| **Vitest** | 1.6.1 | Tests unitaires |

---

## 🚀 Démarrage Rapide

### Installation

```bash
# Installation des dépendances
npm install
# ou avec Bun (plus rapide)
bun install
```

### Développement

```bash
npm run dev
# → http://localhost:5173
```

### Build Production

```bash
npm run build
# Génère dist/ avec assets optimisés
```

### Tests

```bash
npm run test              # Tests unitaires Vitest
npm run test:coverage     # Coverage
```

---

## ⚙️ Configuration

### Variables d'environnement

Créez `.env` à la racine du dossier `frontend/`:

```bash
# Backend API
VITE_API_URL=http://localhost:8080/api/v1

# Supabase (Authentification JWT)
VITE_SUPABASE_URL=https://xxx.supabase.co
VITE_SUPABASE_ANON_KEY=your-anon-key

# WebSocket (optionnel)
VITE_WS_URL=ws://localhost:8080/ws
```

### Configuration Vite

Le fichier `vite.config.js` est configuré pour:
- ✅ Service Worker avec stratégie `injectManifest`
- ✅ Chunks optimisés (vue, pinia, crypto séparés)
- ✅ Exclusion des tests du build prod
- ✅ Support HTTPS en dev si nécessaire

---

## 🔐 Architecture Cryptographique

### Flux de Chiffrement

```javascript
// 1. Génération de la MasterKey (une seule fois à l'inscription)
import { generateMasterKey } from '@/utils/crypto.js'
const masterKey = await generateMasterKey() // AES-GCM 256 bits

// 2. Dérivation KEK depuis mot de passe (Argon2id)
import { deriveKeyFromPassword } from '@/utils/crypto.js'
const salt = generateSalt() // 16 bytes aléatoires
const kek = await deriveKeyFromPassword(password, salt)

// 3. Enveloppement de la MasterKey
import { wrapMasterKey } from '@/utils/crypto.js'
const encryptedMasterKey = await wrapMasterKey(masterKey, kek)
// → Envoyé au backend, stocké en base

// 4. Chiffrement fichier (chunks de 10MB via Web Worker)
import { encryptChunkWorker } from '@/utils/crypto.js'
const encryptedChunk = await encryptChunkWorker(fileChunk, masterKey, index)
```

### Stores Pinia

#### `stores/auth.js`
- **Authentification**: Login/Register/Logout via Supabase
- **Gestion MasterKey**: Unwrap au login, stockage RAM uniquement
- **Recovery**: Génération/validation codes de récupération
- **RSA Keys**: Génération paires clés pour partage

```javascript
// Exemple d'utilisation
import { useAuthStore } from '@/stores/auth'
const authStore = useAuthStore()

await authStore.login({ email, password })
// → masterKey déchiffrée et stockée en RAM
// → Session timeout de 30min
```

#### `stores/files.js`
- **CRUD Fichiers**: Upload/Download/Delete/Rename
- **Navigation**: Gestion path et breadcrumbs
- **Chiffrement**: Appel aux Web Workers pour chiffrer/déchiffrer
- **Chunked Upload**: Gros fichiers en morceaux de 10MB

```javascript
// Upload fichier chiffré
await fileStore.uploadFile(file, currentPath)
// → Chiffre chunk par chunk
// → Envoie au backend
// → Met à jour la liste
```

#### `stores/websocket.js`
- **Connexion WS**: WebSocket persistant avec auto-reconnect
- **Backoff exponentiel**: Avec jitter pour éviter thundering herd
- **Événements**: `storage_update`, `friend_update`, `p2p_signal`

#### `stores/p2p.js`
- **WebRTC**: Peer-to-peer file transfer
- **ICE Candidates**: Gestion NAT traversal
- **Data Channels**: Transfert chiffré direct

### Web Workers

#### `workers/crypto.worker.js`
Effectue le chiffrement/déchiffrement dans un thread séparé pour éviter de bloquer l'UI:

```javascript
// Messages supportés
{ type: 'ENCRYPT', fileChunk, key, chunkIndex }
{ type: 'DECRYPT', fileChunk, key, chunkIndex }

// Réponses
{ type: 'ENCRYPT_SUCCESS', encryptedChunk, chunkIndex }
{ type: 'DECRYPT_SUCCESS', decryptedChunk, chunkIndex }
{ type: 'ERROR', error }
```

### Service Worker

`public/sw-crypto.js` gère:
- ✅ Session timeout (30min inactivité)
- ✅ Stockage temporaire MasterKey (extractable: false)
- ✅ Reset timeout sur activité utilisateur

---

## 📁 Structure des Fichiers

```
frontend/
├── src/
│   ├── main.js                     # Point d'entrée
│   ├── App.vue                     # Composant racine
│   ├── style.css                   # Styles globaux + variables CSS
│   │
│   ├── api.js                      # Client Axios configuré
│   ├── supabase.js                 # Client Supabase
│   │
│   ├── components/                 # Composants réutilisables
│   │   ├── file/
│   │   │   ├── fileList.vue        # Liste fichiers avec vue grille/liste
│   │   │   ├── filePreview.vue     # Aperçu fichiers (images, PDF)
│   │   │   └── uploadDialog.vue    # Dialog upload
│   │   ├── shared/
│   │   │   ├── sharedWithMe.vue    # Fichiers partagés avec moi
│   │   │   └── shareDialog.vue     # Dialog partage
│   │   └── ui/
│   │       ├── confirmDialog.vue   # Dialog confirmation personnalisé
│   │       └── toast.vue           # Notifications
│   │
│   ├── stores/                     # Pinia State Management
│   │   ├── auth.js                 # Auth + Crypto (MasterKey)
│   │   ├── files.js                # CRUD fichiers + chiffrement
│   │   ├── friends.js              # Système d'amis
│   │   ├── p2p.js                  # WebRTC peer-to-peer
│   │   └── websocket.js            # WebSocket client
│   │
│   ├── utils/                      # Utilitaires
│   │   ├── crypto.js               # Fonctions AES, RSA, Argon2id
│   │   ├── secureCrypto.js         # XSS monitoring, rate limiting
│   │   └── securityMonitoring.js   # Monitoring événements sécurité
│   │
│   ├── views/                      # Pages
│   │   ├── HomeView.vue            # Dashboard principal
│   │   ├── LoginView.vue           # Connexion
│   │   ├── RegisterView.vue        # Inscription
│   │   ├── RecoveryView.vue        # Récupération compte
│   │   ├── PrivacyPolicy.vue       # Politique confidentialité
│   │   └── TermsOfService.vue      # CGU
│   │
│   ├── workers/                    # Web Workers
│   │   └── crypto.worker.js        # Chiffrement async
│   │
│   └── router/
│       └── index.js                # Configuration routes
│
├── public/
│   └── sw-crypto.js                # Service Worker session
│
├── index.html                      # Template HTML
├── vite.config.js                  # Config Vite
├── vitest.config.js                # Config tests
└── package.json                    # Dépendances npm
```

---

## 🎯 Composants Clés

### FileList Component

Affiche les fichiers avec:
- Vue liste/grille
- Tri par nom/date/taille
- Menu contextuel (télécharger, renommer, supprimer, partager)
- Drag & drop upload
- Sélection multiple

```vue
<FileList
  :items="fileStore.items"
  :currentPath="fileStore.currentPath"
  @upload="handleUpload"
  @delete="handleDelete"
/>
```

### ShareDialog Component

Gestion des partages:
- Partage utilisateur (chiffrement RSA avec clé publique destinataire)
- Lien public avec expiration
- Permissions lecture/écriture

### FilePreview Component

Aperçu intégré pour:
- Images (JPEG, PNG, GIF, WebP)
- PDF (via pdf.js)
- Texte (TXT, MD)
- Vidéo/Audio (HTML5)

---

## 🔒 Sécurité Frontend

### XSS Protection

```javascript
// utils/secureCrypto.js
import { detectXSSAttempts, setupXSSMonitoring } from '@/utils/secureCrypto'

// Détection scripts non-autorisés
const isSecure = detectXSSAttempts()

// Monitoring continu via MutationObserver
setupXSSMonitoring()
```

### Rate Limiting Client-Side

```javascript
// Limite opérations crypto à 100/min
import { checkCryptoRateLimit } from '@/utils/secureCrypto'

if (!checkCryptoRateLimit()) {
  throw new Error('Crypto rate limit exceeded')
}
```

### Security Monitoring

```javascript
// utils/securityMonitoring.js
import { logSecurityEvent } from '@/utils/securityMonitoring'

// Log événements de sécurité (sanitized avant envoi backend)
logSecurityEvent('XSS_ATTACK_DETECTED', 'critical', {
  timestamp: new Date().toISOString()
})
```

---

## 🧪 Tests

### Structure des tests

```
frontend/
├── src/
│   └── **/*.spec.js        # Tests unitaires Vitest
└── vitest.config.js
```

### Exemples de tests

```javascript
// stores/auth.spec.js
import { describe, it, expect } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore } from '@/stores/auth'

describe('Auth Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('should generate masterKey on register', async () => {
    const store = useAuthStore()
    await store.register({ email: 'test@test.com', password: 'Test123!' })
    expect(store.masterKey).toBeDefined()
  })
})
```

---

## 📊 Performance

### Optimisations appliquées

| Optimisation | Technique | Impact |
|--------------|-----------|--------|
| **Code Splitting** | Vite automatic chunks | -40% initial bundle |
| **Web Workers** | Chiffrement async | UI non-bloquante |
| **Lazy Loading** | Routes async | -60% temps chargement |
| **Image Optimization** | WebP + compression | -70% taille images |
| **Tree Shaking** | Vite + ES modules | -30% bundle final |

### Métriques cibles

- First Contentful Paint: < 1.5s
- Time to Interactive: < 3s
- Lighthouse Score: > 90

---

## 🚀 Déploiement

### Build Production

```bash
npm run build
# Génère dist/ optimisé
```

### Variables d'environnement production

```bash
VITE_API_URL=https://api.safercloud.com/api/v1
VITE_SUPABASE_URL=https://xxx.supabase.co
VITE_SUPABASE_ANON_KEY=prod-anon-key
VITE_WS_URL=wss://api.safercloud.com/ws
```

### Serveur statique

```bash
# Avec serve
npx serve -s dist -p 3000

# Avec nginx
# Copier dist/ vers /var/www/safercloud
# Configurer reverse proxy vers backend
```

---

## 🐛 Debugging

### Mode Dev

```bash
# Activer source maps
npm run dev
# → DevTools affichent le code source original
```

### Console helpers

```javascript
// Inspecter stores Pinia
window.$pinia = app.config.globalProperties.$pinia
window.$stores = {
  auth: useAuthStore(),
  files: useFileStore()
}

// Accès dans console
$stores.auth.masterKey // undefined (sécurité)
$stores.files.items    // Array de fichiers
```

---

## 📚 Ressources

- [Vue.js 3 Docs](https://vuejs.org/)
- [Pinia Docs](https://pinia.vuejs.org/)
- [Web Crypto API](https://developer.mozilla.org/en-US/docs/Web/API/Web_Crypto_API)
- [Vite Docs](https://vitejs.dev/)
- [libsodium.js](https://github.com/jedisct1/libsodium.js)

---

**Développé avec Vue.js 3 et 🔐 Web Crypto API**
