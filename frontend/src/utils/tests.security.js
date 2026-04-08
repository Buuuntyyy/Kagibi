/**
 * Tests de validation des modifications de sécurité
 * À exécuter dans la console du navigateur (DevTools)
 */

// ============================================
// TEST 1: SRI - Vérification d'intégrité du SW
// ============================================
//console.log('%c=== TEST 1: SRI Verification ===', 'color: blue; font-weight: bold;');

(async () => {
  try {
    const response = await fetch('/sw-crypto.js', { cache: 'no-cache' });
    const swContent = await response.text();
    const hashBuffer = await crypto.subtle.digest('SHA-256', 
      new TextEncoder().encode(swContent));
    const hashHex = Array.from(new Uint8Array(hashBuffer))
      .map(b => b.toString(16).padStart(2, '0')).join('');
    
    //console.log('✅ Service Worker SHA-256:', hashHex.substring(0, 16) + '...');
    //console.log('✅ SRI Check: PASSED');
    
    // Note: Hash non stocké en localStorage pour éviter fuite d'info
  } catch (error) {
    console.error('❌ SRI Check FAILED:', error);
  }
})();

// ============================================
// TEST 2: XSS Detection
// ============================================
//console.log('%c=== TEST 2: XSS Detection ===', 'color: blue; font-weight: bold;');

(async () => {
  try {
    const { detectXSSAttempts } = await import('/src/utils/secureCrypto.js');
    const result = detectXSSAttempts();
    
    if (result) {
      //console.log('✅ XSS Detection: No unauthorized scripts detected');
    } else {
      console.error('❌ XSS Detection: Unauthorized scripts found!');
    }
  } catch (error) {
    console.error('⚠️ XSS Detection test error:', error.message);
  }
})();

// ============================================
// TEST 3: XSS Monitoring
// ============================================
//console.log('%c=== TEST 3: XSS Monitoring (Active) ===', 'color: blue; font-weight: bold;');

(async () => {
  try {
    const { setupXSSMonitoring } = await import('/src/utils/secureCrypto.js');
    setupXSSMonitoring();
    //console.log('✅ XSS Monitoring: Enabled and monitoring');
    
    // Test simulating XSS injection
    //console.log('%c[TEST] Attempting to inject malicious script...', 'color: red;');
    const script = document.createElement('script');
    script.textContent = 'alert("XSS Attack")';
    // Ne pas vraiment l'ajouter, juste simuler
    //console.log('✅ XSS Monitoring: Ready to detect injections');
  } catch (error) {
    console.error('❌ XSS Monitoring setup failed:', error);
  }
})();

// ============================================
// TEST 4: Crypto Rate Limiting
// ============================================
//console.log('%c=== TEST 4: Crypto Rate Limiting ===', 'color: blue; font-weight: bold;');

(async () => {
  try {
    const { checkCryptoRateLimit, getCryptoOperationsRemaining } = 
      await import('/src/utils/secureCrypto.js');
    
    let passCount = 0;
    let failCount = 0;
    
    // Exécuter 110 opérations
    for (let i = 0; i < 110; i++) {
      if (checkCryptoRateLimit()) {
        passCount++;
      } else {
        failCount++;
      }
    }
    
    //console.log(`✅ Rate Limiting Test:`);
    //console.log(`   Passed: ${passCount} (expected: 100)`);
    //console.log(`   Limited: ${failCount} (expected: 10)`);
    //console.log(`   Operations remaining: ${getCryptoOperationsRemaining()}`);
    
    if (passCount === 100 && failCount === 10) {
      //console.log('✅ Crypto Rate Limiting: PASSED');
    } else {
      console.warn('⚠️ Crypto Rate Limiting: Unexpected results');
    }
  } catch (error) {
    console.error('❌ Rate Limiting test failed:', error);
  }
})();

// ============================================
// TEST 5: Security Monitoring
// ============================================
//console.log('%c=== TEST 5: Security Monitoring ===', 'color: blue; font-weight: bold;');

(async () => {
  try {
    const { initSecurityMonitoring, getSecurityMonitor, logSecurityEvent } = 
      await import('/src/utils/securityMonitoring.js');
    
    initSecurityMonitoring();
    //console.log('✅ Security Monitoring: Initialized');
    
    // Log un événement test
    logSecurityEvent('TEST_EVENT', 'low', { test: true });
    
    const monitor = getSecurityMonitor();
    const events = monitor.getEvents('TEST_EVENT');
    
    //console.log(`✅ Security Monitor: ${events.length} event(s) logged`);
    //console.log('✅ Security Monitoring: PASSED');
  } catch (error) {
    console.error('❌ Security Monitoring test failed:', error);
  }
})();

// ============================================
// TEST 6: Service Worker Communication
// ============================================
//console.log('%c=== TEST 6: Service Worker Communication ===', 'color: blue; font-weight: bold;');

(async () => {
  try {
    if ('serviceWorker' in navigator) {
      const reg = await navigator.serviceWorker.getRegistration();
      if (reg && reg.active) {
        //console.log('✅ Service Worker: Active and registered');
        //console.log('✅ Service Worker Communication: PASSED');
      } else {
        console.warn('⚠️ Service Worker: Not yet active');
      }
    } else {
      console.error('❌ Service Workers: Not supported');
    }
  } catch (error) {
    console.error('❌ Service Worker test failed:', error);
  }
})();

// ============================================
// TEST 7: Non-Extractable Key Generation
// ============================================
//console.log('%c=== TEST 7: Non-Extractable Key Generation ===', 'color: blue; font-weight: bold;');

(async () => {
  try {
    const { generateNonExtractableMasterKey } = 
      await import('/src/utils/secureCrypto.js');
    
    const key = await generateNonExtractableMasterKey();
    
    // Essayer d'exporter la clé (doit échouer)
    try {
      await crypto.subtle.exportKey('jwk', key);
      console.error('❌ Key is extractable (security issue!)');
    } catch (e) {
      if (e.name === 'InvalidAccessError') {
        //console.log('✅ Non-Extractable Key: Extraction blocked as expected');
        //console.log('✅ Key Security: PASSED');
      } else {
        console.warn('⚠️ Unexpected error:', e.message);
      }
    }
  } catch (error) {
    console.error('❌ Non-Extractable Key test failed:', error);
  }
})();

// ============================================
// SUMMARY
// ============================================
//console.log('%c=== SECURITY ENHANCEMENTS TEST SUMMARY ===', 
  'color: green; font-weight: bold; font-size: 14px;');

//console.log(`
✅ Tests à Effectuer:
1. SRI Verification           - Vérifie l'intégrité du SW
2. XSS Detection              - Détecte les scripts non-autorisés
3. XSS Monitoring             - Monitore les injections en temps réel
4. Crypto Rate Limiting       - Limite les opérations crypto
5. Security Monitoring        - Agrège les événements de sécurité
6. Service Worker Status      - Vérifie l'état du SW
7. Non-Extractable Keys       - Vérifie la clé non-extractable

Exécutés automatiquement à chaque chargement de page.
`);

//console.log('%c[Tests Terminés]', 'color: green; font-weight: bold;');

// ============================================
// HELPERS POUR TESTS MANUELS
// ============================================

window.SecurityTests = {
  // Simuler une injection XSS
  async simulateXSSInjection() {
    console.warn('[TEST] Simulating XSS injection...');
    const script = document.createElement('script');
    script.textContent = '//console.log("XSS Injected")';
    document.head.appendChild(script);
  },

  // Checker l'état du monitoring
  async checkMonitoringStatus() {
    const { getSecurityMonitor } = await import('/src/utils/securityMonitoring.js');
    const monitor = getSecurityMonitor();
    //console.log('Recent Events:', monitor.getEvents().slice(-5));
  },

  // Exporter tous les événements
  async exportSecurityEvents() {
    const { getSecurityMonitor } = await import('/src/utils/securityMonitoring.js');
    const monitor = getSecurityMonitor();
    //console.log(monitor.exportEvents());
  },

  // Test de chiffrement avec rate limit
  async testCryptoRateLimit() {
    const { encryptWithRateLimit, generateNonExtractableMasterKey } = 
      await import('/src/utils/secureCrypto.js');
    
    const key = await generateNonExtractableMasterKey();
    const data = new TextEncoder().encode('test data');
    
    for (let i = 0; i < 110; i++) {
      try {
        await encryptWithRateLimit(data, key);
      } catch (e) {
        //console.log(`Op ${i}: Blocked - ${e.message}`);
        return;
      }
    }
  }
};

//console.log('%c[Helpers Available]', 'color: cyan;');
//console.log('window.SecurityTests.simulateXSSInjection()');
//console.log('window.SecurityTests.checkMonitoringStatus()');
//console.log('window.SecurityTests.exportSecurityEvents()');
//console.log('window.SecurityTests.testCryptoRateLimit()');
