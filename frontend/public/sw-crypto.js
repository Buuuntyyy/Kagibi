// Service Worker pour gestion sécurisée de la MasterKey
// Survit aux F5 mais avec timeout de sécurité

const CACHE_NAME = 'kagibi-crypto-v1';
const SESSION_TIMEOUT = 30 * 60 * 1000; // 30 minutes

// État du Service Worker
let masterKey = null;
let sessionExpiry = null;
let lastActivity = Date.now();

// Validation de l'origine
const ALLOWED_ORIGINS = [
  self.location.origin,
  'http://localhost:5173',
  'http://localhost:3000'
];

function isValidOrigin(origin) {
  return ALLOWED_ORIGINS.includes(origin);
}

// Reset du timeout d'inactivité
function resetSessionTimeout() {
  lastActivity = Date.now();
  sessionExpiry = lastActivity + SESSION_TIMEOUT;
}

// Vérification de l'expiration de session
function isSessionExpired() {
  return sessionExpiry && Date.now() > sessionExpiry;
}

// Nettoyage de la clé
function clearMasterKey() {
  masterKey = null;
  sessionExpiry = null;
  lastActivity = null;
  console.log('[SW-Crypto] MasterKey cleared from memory');
}

// Installation du Service Worker
self.addEventListener('install', (event) => {
  console.log('[SW-Crypto] Installing...');
  self.skipWaiting();
});

// Activation
self.addEventListener('activate', (event) => {
  console.log('[SW-Crypto] Activating...');
  event.waitUntil(self.clients.claim());
});

// Heartbeat pour vérifier l'expiration
setInterval(() => {
  if (isSessionExpired() && masterKey) {
    console.warn('[SW-Crypto] Session expired, clearing MasterKey');
    clearMasterKey();
    
    // Notifier tous les clients
    self.clients.matchAll().then(clients => {
      clients.forEach(client => {
        client.postMessage({
          type: 'SESSION_EXPIRED',
          timestamp: Date.now()
        });
      });
    });
  }
}, 60000); // Vérifier chaque minute

// Gestion des messages du thread principal
self.addEventListener('message', (event) => {
  const { origin } = event;
  
  // SÉCURITÉ: Valider l'origine
  if (!isValidOrigin(origin)) {
    console.error('[SW-Crypto] Rejected message from invalid origin:', origin);
    return;
  }

  const { type, data } = event.data;

  switch (type) {
    case 'STORE_MASTER_KEY':
      // Stocker la MasterKey (déjà non-extractable)
      masterKey = data.masterKey;
      resetSessionTimeout();
      console.log('[SW-Crypto] MasterKey stored in SW memory');
      
      event.ports[0].postMessage({
        success: true,
        message: 'MasterKey stored successfully'
      });
      break;

    case 'GET_MASTER_KEY':
      // Vérifier expiration
      if (isSessionExpired()) {
        clearMasterKey();
        event.ports[0].postMessage({
          success: false,
          error: 'SESSION_EXPIRED'
        });
        return;
      }

      // Retourner la clé
      resetSessionTimeout();
      event.ports[0].postMessage({
        success: true,
        masterKey: masterKey
      });
      break;

    case 'CLEAR_MASTER_KEY':
      // Nettoyage explicite
      clearMasterKey();
      event.ports[0].postMessage({
        success: true,
        message: 'MasterKey cleared'
      });
      break;

    case 'RESET_TIMEOUT':
      // Réinitialiser le timeout (activité utilisateur)
      resetSessionTimeout();
      event.ports[0].postMessage({
        success: true,
        message: 'Timeout reset'
      });
      break;

    case 'CHECK_SESSION':
      // Vérifier si la session est encore valide
      const isExpired = isSessionExpired();
      const hasKey = masterKey !== null;
      
      event.ports[0].postMessage({
        success: true,
        isExpired,
        hasKey,
        expiresIn: sessionExpiry ? sessionExpiry - Date.now() : 0
      });
      break;

    default:
      console.warn('[SW-Crypto] Unknown message type:', type);
      event.ports[0].postMessage({
        success: false,
        error: 'Unknown message type'
      });
  }
});

// Fetch handler (pas de cache pour l'instant)
self.addEventListener('fetch', (event) => {
  // Laisser passer toutes les requêtes normalement
  event.respondWith(fetch(event.request));
});
