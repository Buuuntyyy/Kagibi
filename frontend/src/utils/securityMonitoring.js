/**
 * Monitoring de sécurité côté client
 * Détecte et log les événements de sécurité suspects
 */

import { authClient } from '../auth-client'

class SecurityMonitor {
  constructor() {
    this.events = [];
    this.maxEvents = 100; // Garder les 100 derniers événements
    this.suspiciousThreshold = 5; // Nombre d'événements suspects avant alerte
    this.suspiciousCount = 0;
  }

  /**
   * Log un événement de sécurité
   */
  logSecurityEvent(type, severity, details) {
    const event = {
      timestamp: new Date().toISOString(),
      type,
      severity, // 'low', 'medium', 'high', 'critical'
      details,
      userAgent: navigator.userAgent
    };

    this.events.push(event);

    // Limiter la taille du buffer
    if (this.events.length > this.maxEvents) {
      this.events.shift();
    }

    // Envoyer au backend si severity élevée
    if (severity === 'critical' || severity === 'high') {
      this.reportToBackend(event);
    }

    // Compter les événements suspects
    if (severity === 'high' || severity === 'critical') {
      this.suspiciousCount++;
      if (this.suspiciousCount >= this.suspiciousThreshold) {
        this.handleSuspiciousActivity();
      }
    }
  }

  /**
   * Sanitize les détails avant envoi (ZK: ne jamais envoyer de contenu potentiellement sensible)
   */
  sanitizeEventForBackend(event) {
    return {
      timestamp: event.timestamp,
      type: event.type,
      severity: event.severity,
      // Ne pas envoyer: details, userAgent (fingerprinting)
    };
  }

  /**
   * Rapporte un événement au backend
   */
  async reportToBackend(event) {
    const sanitizedEvent = this.sanitizeEventForBackend(event);
    try {
      // Get token from the current auth provider (Supabase or PocketBase)
      let token = null;
      try {
        token = await authClient.getToken();
      } catch (error) {
        // Silently ignore if auth provider is not available
      }
      
      // Fallback à localStorage
      if (!token) {
        token = localStorage.getItem('kagibi_token');
      }
      
      const headers = {
        'Content-Type': 'application/json'
      };
      
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }
      
      const response = await fetch('/api/v1/security/report', {
        method: 'POST',
        headers,
        body: JSON.stringify(sanitizedEvent)
      });

      if (!response.ok) {
        if (response.status !== 401) {
          console.warn('[SecurityMonitor] Backend returned non-OK status:', response.status)
        }
        // 401 = not authenticated yet, event is cached locally
      }
    } catch (error) {
      // Cannot reach backend (normal before auth)
    }
  }

  /**
   * Gère les activités suspectes (trop d'événements en peu de temps)
   */
  handleSuspiciousActivity() {
    this.logSecurityEvent(
      'SUSPICIOUS_ACTIVITY_THRESHOLD_EXCEEDED',
      'critical',
      {
        count: this.suspiciousCount,
        threshold: this.suspiciousThreshold
      }
    );

    // Notifier l'utilisateur
    window.dispatchEvent(new CustomEvent('suspicious-activity-detected', {
      detail: { count: this.suspiciousCount }
    }));

    // Réinitialiser après notification
    this.suspiciousCount = 0;
  }

  /**
   * Obtient les événements de sécurité
   */
  getEvents(type = null, severity = null) {
    let filtered = this.events;

    if (type) {
      filtered = filtered.filter(e => e.type === type);
    }

    if (severity) {
      filtered = filtered.filter(e => e.severity === severity);
    }

    return filtered;
  }

  /**
   * Exporte les événements pour debugging
   */
  exportEvents() {
    return JSON.stringify(this.events, null, 2);
  }

  /**
   * Réinitialise les événements
   */
  clearEvents() {
    this.events = [];
    this.suspiciousCount = 0;
  }
}

// Instance globale
const securityMonitor = new SecurityMonitor();

/**
 * Initialise le monitoring de sécurité
 */
export function initSecurityMonitoring() {
  // Détecter les accès non-autorisés à la MasterKey
  window.addEventListener('crypto-session-expired', (event) => {
    securityMonitor.logSecurityEvent(
      'SESSION_EXPIRED',
      'medium',
      { timestamp: event.detail.timestamp }
    );
  });

  // Détecter les tentatives XSS
  window.addEventListener('xss-attack-detected', (event) => {
    securityMonitor.logSecurityEvent(
      'XSS_ATTACK_DETECTED',
      'critical',
      { 
        script: event.detail.script,
        timestamp: new Date().toISOString()
      }
    );
  });

  // Détecter les activités suspectes
  window.addEventListener('suspicious-activity-detected', (event) => {
    securityMonitor.logSecurityEvent(
      'SUSPICIOUS_ACTIVITY_THRESHOLD_EXCEEDED',
      'critical',
      { count: event.detail.count }
    );
  });

  // Monitorer les erreurs non capturées
  window.addEventListener('error', (event) => {
    // Ne logger que les erreurs potentiellement liées à la sécurité
    const errorMsg = event.message.toLowerCase();
    const suspiciousPatterns = ['crypto', 'key', 'decrypt', 'encrypt', 'auth', 'token'];
    
    if (suspiciousPatterns.some(pattern => errorMsg.includes(pattern))) {
      securityMonitor.logSecurityEvent(
        'RUNTIME_ERROR',
        'medium',
        {
          message: event.message,
          filename: event.filename,
          lineno: event.lineno
        }
      );
    }
  });

  // Monitorer les promesses rejetées non gérées
  window.addEventListener('unhandledrejection', (event) => {
    const reason = String(event.reason).toLowerCase();
    const suspiciousPatterns = ['crypto', 'key', 'decrypt', 'encrypt', 'auth', 'token'];
    
    if (suspiciousPatterns.some(pattern => reason.includes(pattern))) {
      securityMonitor.logSecurityEvent(
        'UNHANDLED_REJECTION',
        'medium',
        { reason: event.reason }
      );
    }
  });
}

/**
 * Obtient l'instance du moniteur
 */
export function getSecurityMonitor() {
  return securityMonitor;
}

/**
 * Log un événement de sécurité manuel
 */
export function logSecurityEvent(type, severity, details) {
  securityMonitor.logSecurityEvent(type, severity, details);
}

export default securityMonitor;
