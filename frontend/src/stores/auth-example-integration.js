// Exemple d'intégration dans auth.js
// Modifier votre store Pinia existant

import { defineStore } from 'pinia';
import { 
  initSecureCryptoWorker,
  generateNonExtractableMasterKey,
  makeKeyNonExtractable,
  storeMasterKeyInSW,
  getMasterKeyFromSW,
  clearMasterKeyFromSW,
  setupUserActivityTracking,
  checkSessionStatus
} from '@/utils/secureCrypto';

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null,
    token: null,
    masterKey: null, // CryptoKey non-extractable en RAM
    isAuthenticated: false,
  }),

  actions: {
    /**
     * Initialisation au démarrage de l'app
     */
    async initCrypto() {
      // Initialiser le Service Worker
      const swReady = await initSecureCryptoWorker();
      
      if (swReady) {
        // Setup tracking d'activité
        setupUserActivityTracking();

        // Écouter l'expiration de session
        window.addEventListener('crypto-session-expired', () => {
          this.handleSessionExpired();
        });

        // Tenter de restaurer la session après F5
        await this.tryRestoreSession();
      }
    },

    /**
     * Tentative de restauration après F5
     */
    async tryRestoreSession() {
      const status = await checkSessionStatus();
      
      if (status.hasKey && !status.isExpired) {
        // Récupérer la MasterKey du SW
        const masterKey = await getMasterKeyFromSW();
        
        if (masterKey) {
          this.masterKey = masterKey;
          //console.log('[Auth] Session restored from Service Worker');
          
          // Vérifier si le token est encore valide
          const tokenValid = localStorage.getItem('kagibi_token');
          if (tokenValid) {
            this.token = tokenValid;
            await this.checkAuth();
          }
        }
      } else {
        //console.log('[Auth] No valid session to restore');
      }
    },

    /**
     * Login modifié pour générer une clé non-extractable
     */
    async login(email, password) {
      try {
        // 1. Login Supabase
        const response = await fetch('/api/auth/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ email, password })
        });

        if (!response.ok) throw new Error('Login failed');

        const data = await response.json();
        
        // 2. Récupérer/Dériver la MasterKey
        // Selon votre architecture, vous avez 2 options:
        
        // Option A: Générer une nouvelle clé (si pas d'encrypted_master_key en BDD)
        // const masterKey = await generateNonExtractableMasterKey();
        
        // Option B: Dériver depuis le password et rendre non-extractable
        const kek = await this.deriveKEK(password); // Votre fonction existante
        const extractableMasterKey = await this.decryptMasterKey(data.encrypted_master_key, kek);
        const masterKey = await makeKeyNonExtractable(extractableMasterKey);

        // 3. Stocker en RAM et SW
        this.masterKey = masterKey;
        await storeMasterKeyInSW(masterKey);

        // 4. Stocker le reste normalement
        this.token = data.token;
        this.user = data.user;
        this.isAuthenticated = true;
        localStorage.setItem('kagibi_token', data.token);

        // NE PLUS FAIRE: sessionStorage.setItem("kagibi_mk", ...)
        
        //console.log('[Auth] Login successful with non-extractable MasterKey');
        return true;
      } catch (error) {
        console.error('[Auth] Login error:', error);
        return false;
      }
    },

    /**
     * Logout modifié
     */
    async logout() {
      try {
        // 1. Nettoyer le Service Worker
        await clearMasterKeyFromSW();

        // 2. Nettoyer le state
        this.masterKey = null;
        this.token = null;
        this.user = null;
        this.isAuthenticated = false;

        // 3. Nettoyer localStorage
        localStorage.removeItem('kagibi_token');
        
        //console.log('[Auth] Logout successful');
        return true;
      } catch (error) {
        console.error('[Auth] Logout error:', error);
        return false;
      }
    },

    /**
     * Gestion de l'expiration de session
     */
    handleSessionExpired() {
      console.warn('[Auth] Session expired, forcing logout');
      
      // Afficher une alerte
      alert('Votre session a expiré pour des raisons de sécurité. Veuillez vous reconnecter.');
      
      // Logout forcé
      this.logout();
      
      // Redirection vers login
      window.location.href = '/login';
    },

    /**
     * Chiffrement avec la clé non-extractable
     */
    async encryptFile(fileData) {
      if (!this.masterKey) {
        throw new Error('No MasterKey available');
      }

      const iv = window.crypto.getRandomValues(new Uint8Array(12));
      
      const encrypted = await window.crypto.subtle.encrypt(
        { name: "AES-GCM", iv },
        this.masterKey,
        fileData
      );

      return { encrypted, iv };
    },

    /**
     * Déchiffrement avec la clé non-extractable
     */
    async decryptFile(encryptedData, iv) {
      if (!this.masterKey) {
        throw new Error('No MasterKey available');
      }

      return await window.crypto.subtle.decrypt(
        { name: "AES-GCM", iv },
        this.masterKey,
        encryptedData
      );
    }
  }
});
