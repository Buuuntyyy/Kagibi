// Gestion sécurisée de la MasterKey avec Service Worker
// extractable: false + SW storage = Sécurité + UX

let swRegistration = null;
let serviceWorkerReady = false;

/**
 * Initialise le Service Worker pour la gestion crypto
 */
export async function initSecureCryptoWorker() {
  if ('serviceWorker' in navigator) {
    try {
      swRegistration = await navigator.serviceWorker.register('/sw-crypto.js', {
        scope: '/'
      });

      // Attendre que le SW soit actif
      if (swRegistration.active) {
        serviceWorkerReady = true;
        console.log('[SecureCrypto] Service Worker is active');
      } else {
        await navigator.serviceWorker.ready;
        serviceWorkerReady = true;
        console.log('[SecureCrypto] Service Worker is ready');
      }

      // Écouter les messages du SW
      navigator.serviceWorker.addEventListener('message', handleServiceWorkerMessage);

      return true;
    } catch (error) {
      console.error('[SecureCrypto] Failed to register Service Worker:', error);
      return false;
    }
  } else {
    console.warn('[SecureCrypto] Service Worker not supported');
    return false;
  }
}

/**
 * Gestion des messages du Service Worker
 */
function handleServiceWorkerMessage(event) {
  const { type } = event.data;

  switch (type) {
    case 'SESSION_EXPIRED':
      console.warn('[SecureCrypto] Session expired by Service Worker');
      // Déclencher un événement personnalisé pour notifier l'app
      window.dispatchEvent(new CustomEvent('crypto-session-expired', {
        detail: { timestamp: event.data.timestamp }
      }));
      break;
  }
}

/**
 * Envoie un message au Service Worker et attend une réponse
 */
async function sendMessageToSW(type, data = {}) {
  if (!serviceWorkerReady || !swRegistration) {
    throw new Error('Service Worker not ready');
  }

  return new Promise((resolve, reject) => {
    const messageChannel = new MessageChannel();

    messageChannel.port1.onmessage = (event) => {
      if (event.data.success) {
        resolve(event.data);
      } else {
        reject(new Error(event.data.error || 'Unknown error'));
      }
    };

    // Envoyer au SW actif
    const sw = swRegistration.active || navigator.serviceWorker.controller;
    if (sw) {
      sw.postMessage({ type, data }, [messageChannel.port2]);
    } else {
      reject(new Error('No active Service Worker'));
    }
  });
}

/**
 * Génère une MasterKey NON-EXTRACTABLE
 * SÉCURITÉ: extractable: false empêche crypto.subtle.exportKey()
 */
export async function generateNonExtractableMasterKey() {
  const masterKey = await window.crypto.subtle.generateKey(
    { name: "AES-GCM", length: 256 },
    false, // extractable: false ← CRITIQUE pour sécurité
    ["encrypt", "decrypt"]
  );

  console.log('[SecureCrypto] Non-extractable MasterKey generated');
  return masterKey;
}

/**
 * Wrap une MasterKey existante pour la rendre non-extractable
 * Utilisé lors du login avec une clé dérivée du mot de passe
 */
export async function makeKeyNonExtractable(extractableKey) {
  // Exporter temporairement
  const jwk = await window.crypto.subtle.exportKey("jwk", extractableKey);
  
  // Ré-importer comme non-extractable
  const nonExtractableKey = await window.crypto.subtle.importKey(
    "jwk",
    jwk,
    "AES-GCM",
    false, // extractable: false
    ["encrypt", "decrypt"]
  );

  // Effacer le JWK de la mémoire (best effort)
  for (let key in jwk) {
    delete jwk[key];
  }

  console.log('[SecureCrypto] Key made non-extractable');
  return nonExtractableKey;
}

/**
 * Stocke la MasterKey dans le Service Worker
 */
export async function storeMasterKeyInSW(masterKey) {
  try {
    const response = await sendMessageToSW('STORE_MASTER_KEY', { masterKey });
    console.log('[SecureCrypto] MasterKey stored in SW:', response.message);
    return true;
  } catch (error) {
    console.error('[SecureCrypto] Failed to store MasterKey in SW:', error);
    return false;
  }
}

/**
 * Récupère la MasterKey depuis le Service Worker
 * Utilisé après un F5 pour restaurer la clé
 */
export async function getMasterKeyFromSW() {
  try {
    const response = await sendMessageToSW('GET_MASTER_KEY');
    
    if (response.masterKey) {
      console.log('[SecureCrypto] MasterKey retrieved from SW');
      return response.masterKey;
    }
    
    return null;
  } catch (error) {
    if (error.message === 'SESSION_EXPIRED') {
      console.warn('[SecureCrypto] Session expired, MasterKey not available');
      return null;
    }
    console.error('[SecureCrypto] Failed to get MasterKey from SW:', error);
    return null;
  }
}

/**
 * Nettoie la MasterKey du Service Worker
 */
export async function clearMasterKeyFromSW() {
  try {
    await sendMessageToSW('CLEAR_MASTER_KEY');
    console.log('[SecureCrypto] MasterKey cleared from SW');
    return true;
  } catch (error) {
    console.error('[SecureCrypto] Failed to clear MasterKey from SW:', error);
    return false;
  }
}

/**
 * Réinitialise le timeout de session (activité utilisateur)
 */
export async function resetSessionTimeout() {
  try {
    await sendMessageToSW('RESET_TIMEOUT');
    return true;
  } catch (error) {
    console.error('[SecureCrypto] Failed to reset timeout:', error);
    return false;
  }
}

/**
 * Vérifie l'état de la session
 */
export async function checkSessionStatus() {
  try {
    const response = await sendMessageToSW('CHECK_SESSION');
    return {
      isExpired: response.isExpired,
      hasKey: response.hasKey,
      expiresIn: response.expiresIn
    };
  } catch (error) {
    console.error('[SecureCrypto] Failed to check session:', error);
    return { isExpired: true, hasKey: false, expiresIn: 0 };
  }
}

/**
 * Chiffrement avec la clé non-extractable
 */
export async function encryptWithNonExtractableKey(data, masterKey) {
  const iv = window.crypto.getRandomValues(new Uint8Array(12));
  
  const encrypted = await window.crypto.subtle.encrypt(
    { name: "AES-GCM", iv },
    masterKey,
    data
  );

  return { encrypted, iv };
}

/**
 * Déchiffrement avec la clé non-extractable
 */
export async function decryptWithNonExtractableKey(encrypted, iv, masterKey) {
  return await window.crypto.subtle.decrypt(
    { name: "AES-GCM", iv },
    masterKey,
    encrypted
  );
}

/**
 * Configuration de l'activité utilisateur
 * Appeler sur les événements utilisateur pour reset le timeout
 */
export function setupUserActivityTracking() {
  const events = ['mousedown', 'keydown', 'scroll', 'touchstart'];
  
  let lastReset = Date.now();
  const RESET_INTERVAL = 60000; // Reset max toutes les 1 minute

  const handleActivity = () => {
    const now = Date.now();
    if (now - lastReset > RESET_INTERVAL) {
      resetSessionTimeout();
      lastReset = now;
    }
  };

  events.forEach(event => {
    window.addEventListener(event, handleActivity, { passive: true });
  });

  console.log('[SecureCrypto] User activity tracking enabled');
}

// ============================================
// DÉTECTION XSS ET INJECTION DE SCRIPTS
// ============================================

/**
 * Détecte les tentatives d'injection de scripts XSS
 * Vérifie que tous les scripts ont un nonce valide ou sont des scripts autorisés
 */
export function detectXSSAttempts() {
  const scriptTags = document.querySelectorAll('script');
  const unAuthorizedScripts = [];
  
  scriptTags.forEach((script) => {
    // Ignorer les scripts autorisés
    const hasNonce = script.hasAttribute('nonce');
    const isModule = script.type === 'module';
    const isSRI = script.hasAttribute('nonce') && script.textContent.includes('verifyServiceWorkerIntegrity');
    const isInlineEvent = script.hasAttribute('onload') || script.hasAttribute('onerror');
    
    // Scripts inline sans nonce sont suspects (sauf les scripts SRI avec nonce)
    if (script.textContent && !hasNonce && !isModule && !isSRI) {
      unAuthorizedScripts.push({
        reason: 'Inline script without nonce',
        hasContent: true,
        length: script.textContent.length
      });
    }
    
    if (isInlineEvent) {
      unAuthorizedScripts.push({
        reason: 'Inline event handler detected',
        handlerType: script.hasAttribute('onload') ? 'onload' : 'onerror'
      });
    }
  });
  
  if (unAuthorizedScripts.length > 0) {
    console.error('[XSS-Detection] Unauthorized scripts detected:', unAuthorizedScripts.length);
    unAuthorizedScripts.forEach((item, idx) => {
      console.error(`[XSS-Detection] Script ${idx + 1}:`, item);
    });
    return false;
  }
  
  console.log('[XSS-Detection] No unauthorized scripts detected');
  return true;
}

/**
 * Monitore les modifications du DOM pour détecter les injections
 */
export function setupXSSMonitoring() {
  // MutationObserver pour détecter l'injection de scripts
  const observer = new MutationObserver((mutations) => {
    mutations.forEach((mutation) => {
      if (mutation.type === 'childList') {
        mutation.addedNodes.forEach((node) => {
          // Vérifier les scripts ajoutés dynamiquement
          if (node.tagName === 'SCRIPT') {
            // Vérifier le nonce
            const nonce = node.getAttribute('nonce');
            if (!nonce && node.textContent && !node.type.includes('module')) {
              console.error('[XSS-Monitoring] Malicious script injection detected!');
              // Bloquer le script
              if (node.parentNode) {
                node.parentNode.removeChild(node);
              }
              
              // Notifier l'utilisateur (sans exposer le contenu du script)
              window.dispatchEvent(new CustomEvent('xss-attack-detected', {
                detail: { detected: true, scriptLength: node.textContent.length }
              }));
            }
          }
          
          // Vérifier les images avec onerror handlers
          if (node.tagName === 'IMG' && node.hasAttribute('onerror')) {
            console.error('[XSS-Monitoring] XSS attempt via img onerror handler');
            node.removeAttribute('onerror');
          }
          
          // Vérifier les event handlers inline
          if (node.hasAttribute && (node.hasAttribute('onclick') || node.hasAttribute('onerror') || node.hasAttribute('onload'))) {
            console.error('[XSS-Monitoring] Inline event handler detected');
            node.removeAttribute('onclick');
            node.removeAttribute('onerror');
            node.removeAttribute('onload');
          }
        });
      }
    });
  });
  
  // Observer le document
  observer.observe(document.documentElement, {
    childList: true,
    subtree: true
  });
  
  console.log('[XSS-Monitoring] XSS monitoring enabled');
  return observer;
}

// ============================================
// RATE LIMITING CÔTÉ CLIENT (CRYPTO OPS)
// ============================================

let cryptoOperations = 0;
let lastResetTime = Date.now();
const CRYPTO_RATE_LIMIT = 100; // 100 opérations par minute
const CRYPTO_RATE_WINDOW = 60000; // 1 minute

/**
 * Vérifie le rate limit pour les opérations cryptographiques
 * Limite les opérations crypto pour éviter les attaques DoS
 */
export function checkCryptoRateLimit() {
  const now = Date.now();
  
  // Réinitialiser la fenêtre si 1 minute est passée
  if (now - lastResetTime > CRYPTO_RATE_WINDOW) {
    cryptoOperations = 0;
    lastResetTime = now;
  }
  
  // Incrémenter le compteur
  cryptoOperations++;
  
  // Vérifier le limit
  if (cryptoOperations > CRYPTO_RATE_LIMIT) {
    console.error('[Crypto-RateLimit] Rate limit exceeded:', cryptoOperations, '>', CRYPTO_RATE_LIMIT);
    return false;
  }
  
  return true;
}

/**
 * Obtient le nombre d'opérations crypto restantes
 */
export function getCryptoOperationsRemaining() {
  return Math.max(0, CRYPTO_RATE_LIMIT - cryptoOperations);
}

/**
 * Wrapper pour encryptWithNonExtractableKey avec rate limit
 */
export async function encryptWithRateLimit(data, masterKey) {
  if (!checkCryptoRateLimit()) {
    throw new Error('Crypto operations rate limit exceeded');
  }
  
  return encryptWithNonExtractableKey(data, masterKey);
}

/**
 * Wrapper pour decryptWithNonExtractableKey avec rate limit
 */
export async function decryptWithRateLimit(encrypted, iv, masterKey) {
  if (!checkCryptoRateLimit()) {
    throw new Error('Crypto operations rate limit exceeded');
  }
  
  return decryptWithNonExtractableKey(encrypted, iv, masterKey);
}
