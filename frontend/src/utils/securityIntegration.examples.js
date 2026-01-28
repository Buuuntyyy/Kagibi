/**
 * Exemple d'intégration du monitoring de sécurité dans une action Pinia
 * À adapter selon votre structure actuelle
 */

import { logSecurityEvent } from '@/utils/securityMonitoring'
import { encryptWithRateLimit, decryptWithRateLimit } from '@/utils/secureCrypto'

/**
 * Exemple: Action pour uploader un fichier
 * Intègre la détection de rate limiting
 */
export async function uploadFileWithSecurityMonitoring(store, file, parentFolderId) {
  try {
    // 1. Log de sécurité - début upload
    logSecurityEvent('FILE_UPLOAD_STARTED', 'low', {
      fileName: file.name,
      fileSize: file.size,
      parentFolderId,
      timestamp: new Date().toISOString()
    });

    // 2. Chiffrer le fichier avec rate limit
    let encrypted;
    try {
      const { encrypted: enc, iv } = await encryptWithRateLimit(
        await file.arrayBuffer(),
        store.masterKey
      );
      encrypted = { data: enc, iv };
    } catch (error) {
      if (error.message.includes('rate limit')) {
        logSecurityEvent('CRYPTO_RATE_LIMIT_EXCEEDED', 'high', {
          operation: 'file_upload_encryption',
          fileName: file.name
        });
        throw new Error('Trop d\'uploads simultanés, veuillez réessayer');
      }
      throw error;
    }

    // 3. Uploader le fichier
    const formData = new FormData();
    formData.append('file', new Blob([encrypted.data]));
    formData.append('iv', btoa(String.fromCharCode(...encrypted.iv)));
    formData.append('parentFolderId', parentFolderId);

    const response = await fetch('/api/v1/files/upload', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${store.token}`
      },
      body: formData
    });

    if (!response.ok) {
      throw new Error('Upload failed');
    }

    const result = await response.json();

    // 4. Log de sécurité - upload réussi
    logSecurityEvent('FILE_UPLOAD_COMPLETED', 'low', {
      fileName: file.name,
      fileSize: file.size,
      fileId: result.fileId
    });

    return result;
  } catch (error) {
    // 5. Log de sécurité - erreur
    logSecurityEvent('FILE_UPLOAD_FAILED', 'medium', {
      fileName: file.name,
      error: error.message
    });
    throw error;
  }
}

/**
 * Exemple: Action pour télécharger un fichier
 * Intègre la détection de rate limiting
 */
export async function downloadFileWithSecurityMonitoring(store, fileId) {
  try {
    // 1. Log de sécurité - début download
    logSecurityEvent('FILE_DOWNLOAD_STARTED', 'low', {
      fileId,
      timestamp: new Date().toISOString()
    });

    // 2. Télécharger le fichier
    const response = await fetch(`/api/v1/files/download/${fileId}`, {
      headers: {
        'Authorization': `Bearer ${store.token}`
      }
    });

    if (!response.ok) {
      logSecurityEvent('FILE_DOWNLOAD_UNAUTHORIZED', 'high', {
        fileId,
        status: response.status
      });
      throw new Error('Unauthorized');
    }

    const encrypted = await response.arrayBuffer();
    const ivHeader = response.headers.get('X-File-IV');

    // 3. Déchiffrer avec rate limit
    try {
      const iv = new Uint8Array(atob(ivHeader).split('').map(c => c.charCodeAt(0)));
      const decrypted = await decryptWithRateLimit(
        encrypted,
        iv,
        store.masterKey
      );
      
      // 4. Log de sécurité - download réussi
      logSecurityEvent('FILE_DOWNLOAD_COMPLETED', 'low', {
        fileId,
        decryptedSize: decrypted.byteLength
      });

      return decrypted;
    } catch (error) {
      if (error.message.includes('rate limit')) {
        logSecurityEvent('CRYPTO_RATE_LIMIT_EXCEEDED', 'high', {
          operation: 'file_download_decryption',
          fileId
        });
        throw new Error('Trop de téléchargements simultanés, veuillez réessayer');
      }
      throw error;
    }
  } catch (error) {
    // 5. Log de sécurité - erreur
    logSecurityEvent('FILE_DOWNLOAD_FAILED', 'medium', {
      fileId,
      error: error.message
    });
    throw error;
  }
}

/**
 * Exemple: Action de changement de password
 * Intègre le monitoring
 */
export async function changePasswordWithSecurityMonitoring(store, newPassword) {
  try {
    logSecurityEvent('PASSWORD_CHANGE_INITIATED', 'medium', {
      timestamp: new Date().toISOString()
    });

    // Appel au backend (ajuster selon votre implémentation)
    const response = await fetch('/api/v1/users/change-password', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${store.token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ newPassword })
    });

    if (!response.ok) {
      logSecurityEvent('PASSWORD_CHANGE_FAILED', 'medium', {
        status: response.status
      });
      throw new Error('Password change failed');
    }

    logSecurityEvent('PASSWORD_CHANGE_COMPLETED', 'medium', {
      timestamp: new Date().toISOString()
    });

    return true;
  } catch (error) {
    logSecurityEvent('PASSWORD_CHANGE_ERROR', 'high', {
      error: error.message
    });
    throw error;
  }
}

/**
 * Exemple: Handler pour les événements de sécurité
 */
export function setupSecurityEventHandlers() {
  // XSS Attack Detected
  window.addEventListener('xss-attack-detected', (event) => {
    console.error('[SECURITY] XSS Attack Detected!', event.detail);
    // Notifier l'utilisateur
    alert('⚠️ Tentative de sécurité détectée. Veuillez rafraîchir la page.');
  });

  // Session Expired
  window.addEventListener('crypto-session-expired', (event) => {
    console.warn('[SECURITY] Session expired', event.detail);
    // Rediriger vers login
    window.location.href = '/login?reason=session_expired';
  });

  // Suspicious Activity Detected
  window.addEventListener('suspicious-activity-detected', (event) => {
    console.error('[SECURITY] Suspicious Activity!', event.detail);
    // Notifier l'utilisateur
    alert('⚠️ Activité anormale détectée. Votre compte a été sécurisé.');
  });

  console.log('[Security] Event handlers registered');
}

export default {
  uploadFileWithSecurityMonitoring,
  downloadFileWithSecurityMonitoring,
  changePasswordWithSecurityMonitoring,
  setupSecurityEventHandlers
}
