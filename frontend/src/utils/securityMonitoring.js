/**
 * Monitoring de sécurité côté client
 * Détecte et log les événements de sécurité suspects
 */

import { supabase } from '../supabase'

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

    // Log en console
    const color = this.getSeverityColor(severity);
    console.log(
      `%c[SecurityMonitor] ${type}`,
      `color: ${color}; font-weight: bold;`,
      details
    );

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
   * Couleur de sévérité pour console
   */
  getSeverityColor(severity) {
    const colors = {
      low: '#FFA500',      // Orange
      medium: '#FF6B6B',   // Rouge clair
      high: '#FF0000',     // Rouge
      critical: '#8B0000'  // Rouge foncé
    };
    return colors[severity] || '#000000';
  }

  /**
   * Rapporte un événement au backend
   */
  async reportToBackend(event) {
    try {
      // Essayer d'obtenir le token de Supabase
      let token = null;
      try {
        const { data: { session } } = await supabase.auth.getSession();
        token = session?.access_token;
      } catch (error) {
        // Silencieusement ignorer si Supabase n'est pas disponible
      }
      
      // Fallback à localStorage
      if (!token) {
        token = localStorage.getItem('safercloud_token');
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
        body: JSON.stringify(event)
      });

      if (response.ok) {
        console.log('[SecurityMonitor] Event reported to backend');
      } else if (response.status === 401) {
        console.debug('[SecurityMonitor] Not authenticated yet, event cached locally');
      } else {
        console.warn(`[SecurityMonitor] Backend returned ${response.status}`);
      }
    } catch (error) {
      console.debug('[SecurityMonitor] Cannot reach backend (normal before auth):', error.message);
    }
  }

  /**
   * Gère les activités suspectes (trop d'événements en peu de temps)
   */
  handleSuspiciousActivity() {
    console.error('[SecurityMonitor] Suspicious activity detected!');
    
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

  console.log('[SecurityMonitoring] Initialized');
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
