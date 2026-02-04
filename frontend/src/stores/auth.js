import { defineStore } from 'pinia'
import api from '../api'
import router from '../router'
import { supabase } from '../supabase'
import { 
  deriveKeyFromPassword, generateSalt, generateMasterKey, wrapMasterKey, unwrapMasterKey, 
  generateRecoveryCode, deriveKeyFromRecoveryCode, hashRecoveryCode,
  generateRSAKeyPair, exportKeyToPEM, importKeyFromPEM, encryptPrivateKey, decryptPrivateKey 
} from '../utils/crypto'
import sodium from 'libsodium-wrappers-sumo'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    isAuthenticated: false,
    user: null,
    masterKey: null,
    privateKey: null, // RSA Private Key (Unwrapped)
    publicKey: null,  // RSA Public Key (CryptoKey)
    sessionTimeoutId: null, // Timeout handler for security (30 minutes)
  }),
  actions: {
    // --- Key Management Helpers ---
    async ensureRSAKeys(masterKey) {
       if (!this.user) return;
       
       await sodium.ready; // Ensure sodium is ready before base64 ops

       // Si l'utilisateur n'a pas encore de clés (migration ou nouvel utilisateur)
       if (!this.user.public_key || !this.user.encrypted_private_key) {
           console.log("Generating RSA keys for user...");
           const keyPair = await generateRSAKeyPair();
           
           const publicKeyPEM = await exportKeyToPEM(keyPair.publicKey, 'spki');
           // Encrypt private key with Master Key
           const encryptedPrivateKey = await encryptPrivateKey(keyPair.privateKey, masterKey);
           
           // Send to server
           await api.post('/users/keys', {
               public_key: publicKeyPEM,
               encrypted_private_key: encryptedPrivateKey
           });
           
           this.user.public_key = publicKeyPEM;
           this.user.encrypted_private_key = encryptedPrivateKey;
           this.privateKey = keyPair.privateKey;
           this.publicKey = keyPair.publicKey;
       } else {
           // Decrypt existing private key
           try {
               this.privateKey = await decryptPrivateKey(this.user.encrypted_private_key, masterKey);
               this.publicKey = await importKeyFromPEM(this.user.public_key, 'spki'); // Load public key object too
           } catch (e) {
               console.error("Failed to decrypt RSA Private Key:", e);
               // Handle error (maybe re-generate? Careful with data loss)
           }
       }
    },

    async login(credentials) {
      try {
        // 1. Authentification Supabase
        const { data, error } = await supabase.auth.signInWithPassword({
          email: credentials.email,
          password: credentials.password,
        })

        if (error) throw error

        // 2. Récupérer les clés de chiffrement depuis votre Backend
        // Le token Supabase est injecté automatiquement par l'intercepteur api.js
        let keysResponse;
        try {
          keysResponse = await api.get('/auth/keys');
        } catch (err) {
          console.error("Failed to fetch keys from backend:", err);
          throw new Error("Impossible de récupérer les clés de chiffrement du serveur.");
        }
        
        const { salt, encrypted_master_key } = keysResponse.data;

        if (salt && encrypted_master_key) {
          try {
            await sodium.ready;
            
            const saltBytes = sodium.from_hex(salt);
            
            // Le mot de passe sert toujours à déchiffrer la clé maître
            const kek = await deriveKeyFromPassword(credentials.password, saltBytes);
            
            this.masterKey = await unwrapMasterKey(encrypted_master_key, kek);
            
            // SECURITY: MasterKey stays in RAM only, NOT persisted to storage
            // Set up automatic session timeout for security (30 minutes)
            this.setupSessionTimeout();

            // Generate/Load RSA Keys (New functionality)
            await this.fetchUser(); // Get latest user data including keys
            await this.ensureRSAKeys(this.masterKey);
          } catch (decryptError) {
            console.error("Decryption failed during login:", decryptError);
            throw new Error("Impossible de déchiffrer vos clés. Votre mot de passe est-il correct ?");
          }
          
        } else {
          this.masterKey = null;
          throw new Error("Pas de clés de chiffrement trouvées sur le serveur. Veuillez contacter le support.");
        }

        this.isAuthenticated = true;
        // this.user is populated by fetchUser() inside the block above

        await this.fetchUser();
        // Persist user data to localStorage (non-sensitive fields only)
        this.persistUserToStorage();
        router.push({ name: 'Home' });
        return true
      } catch (error) {
        console.error("Login failed:", error);
        this.isAuthenticated = false;
        this.user = null;
        throw error; // Re-throw for the UI to handle
      }
    },
    async register(username, email, password, avatarUrl = '/avatars/default.png') {
      await sodium.ready;
      
      // 1. Préparation de la cryptographie locale
      const salt = generateSalt();
      const saltHex = sodium.to_hex(salt);

      const masterKey = await generateMasterKey();
      const kek = await deriveKeyFromPassword(password, salt);
      const wrappedMasterKey = await wrapMasterKey(masterKey, kek);

      // Generate Recovery Code
      const recoveryCode = generateRecoveryCode();
      const recoveryHash = await hashRecoveryCode(recoveryCode);
      
      // Encrypt Master Key with Recovery Code
      const recoveryKek = await deriveKeyFromRecoveryCode(recoveryCode, salt);
      const wrappedMasterKeyRecovery = await wrapMasterKey(masterKey, recoveryKek);

      // Generate RSA Keys for new user
      const keyPair = await generateRSAKeyPair();
      const publicKeyPEM = await exportKeyToPEM(keyPair.publicKey, 'spki');
      const encryptedPrivateKey = await encryptPrivateKey(keyPair.privateKey, masterKey);

      try {
        // 2. Création du compte Supabase Auth
        const { data, error } = await supabase.auth.signUp({
          email: email,
          password: password,
          options: {
            data: { name: username }
          }
        })

        if (error) throw error
        
        // Check for missing session (Email Confirmation enabled case)
        if (!data.session && data.user) {
            throw new Error("L'inscription nécessite que la confirmation d'email soit DÉSACTIVÉE dans Supabase. Les clés de chiffrement générées ne peuvent pas être sauvegardées sans session active.")
        }

        // 3. Création du profil chiffré sur votre Backend
        // La session est active après le signUp (si confirmation email désactivée)
        // L'intercepteur injectera le token, mais on force pour être sûr (race condition)
        const accessToken = data.session?.access_token;
        const config = {};
        if (accessToken) {
            config.headers = { Authorization: `Bearer ${accessToken}` };
        } else {
             // Should be caught by the check above, but as a safety:
             throw new Error("Erreur: Token d'accès manquant après l'inscription.")
        }

        const payload = {
          name: username,
          email: email,
          avatar_url: avatarUrl,
          // Pas de password envoyé à votre backend !
          salt: saltHex,
          encrypted_master_key: wrappedMasterKey,
          encrypted_master_key_recovery: wrappedMasterKeyRecovery,
          recovery_hash: recoveryHash,
          recovery_salt: saltHex,
          public_key: publicKeyPEM,
          encrypted_private_key: encryptedPrivateKey
        };

        await api.post('/auth/register', payload, config)

        // Initialize state for immediate usage (Auto-Login)
        this.masterKey = masterKey;
        this.isAuthenticated = true;
        
        // SECURITY: MasterKey stays in RAM only, NOT persisted to storage
        this.setupSessionTimeout();

        // Fetch user completely
        await this.fetchUser();

      } catch (err) {
        // En cas d'erreur backend, on essaie de nettoyer le compte Supabase ? 
        // Idéalement oui, mais pour l'instant on renvoie l'erreur
        if (err.response && err.response.data && err.response.data.error) {
          throw new Error(err.response.data.error)
        } else if (err.message) {
           throw new Error(err.message)
        } else {
          throw new Error("Erreur lors de l'inscription")
        }
      }
      return recoveryCode; // Return code so UI can display it
    },
    async logout() {
      try {
        await api.post('/auth/logout'); // Backend logout
        await supabase.auth.signOut();  // Supabase logout (clears localStorage)
      } catch (error) {
        console.error("Logout failed:", error)
      } finally {
        this.isAuthenticated = false;
        this.user = null;
        this.masterKey = null;
        localStorage.removeItem("safercloud_user");
        router.push({ name: 'Login' });
      }
    },
    // Old checkAuth removed to avoid duplication

    async fetchUser() {
      try {
        // Force token injection if session works but interceptor lags
        const { data: { session } } = await supabase.auth.getSession();
        const headers = {};
        if (session?.access_token) {
            headers.Authorization = `Bearer ${session.access_token}`;
        }

        // On passe les headers explicitement
        const response = await api.get('/users/me', { headers });
        this.user = response.data;
        this.persistUserToStorage();
      } catch (error) {
        console.error("Failed to fetch user:", error)
        // Ne pas logout immédiatement si c'est juste une erreur réseau // this.logout();
      }
    },
    async updatePassword(currentPassword, newPassword) {
      if (!this.masterKey) {
        throw new Error("Master key not available. Please re-login.");
      }

      await sodium.ready;
      
      try {
        // 1. Mise à jour via Supabase (auth.users)
        console.log("Updating password in Supabase...");
        const { error } = await supabase.auth.updateUser({ password: newPassword });
        if (error) {
          console.error("Supabase password update error:", error);
          throw new Error("Erreur Supabase: " + error.message);
        }

        // 2. Mise à jour des clés chiffrées sur votre Backend (profiles)
        // Car le "KEK" qui protège la MasterKey dépend du mot de passe !
        
        const newSalt = generateSalt(); // Nouveau sel pour la nouvelle clé crypto
        const newSaltHex = sodium.to_hex(newSalt);
        const newKek = await deriveKeyFromPassword(newPassword, newSalt);
        const newEncryptedMasterKey = await wrapMasterKey(this.masterKey, newKek);
        // On appelle votre API pour mettre à jour Salt + EncryptedMasterKey
        // Note: l'API ne vérifie plus 'current_password' car c'est Supabase qui gère l'auth.
        // Cependant, pour sécuriser cet appel critique, votre backend pourrait demander de re-confirmer
        // l'ancien mot de passe, mais avec Supabase c'est complexe.
        // Pour l'instant on fait confiance à la session active.
        
        await api.post('/users/change-password', {
          new_salt: newSaltHex,
          new_encrypted_master_key: newEncryptedMasterKey
        });

        // SECURITY: MasterKey stays in RAM only - no persistence needed
        // The key remains valid in memory for current session
      } catch (error) {
        console.error("Password update failed:", error);
        throw error;
      }
    },
    async updateUsername(newName) {
      if (!newName || newName.trim().length === 0) {
        throw new Error("Le nom d'utilisateur ne peut pas être vide.");
      }

      try {
        // 1. Mettre à jour Supabase Auth (user metadata)
        const { error: supabaseError } = await supabase.auth.updateUser({
          data: { name: newName.trim() }
        });

        if (supabaseError) {
          console.error("Supabase username update error:", supabaseError);
          throw new Error("Erreur lors de la mise à jour Supabase: " + supabaseError.message);
        }

        // 2. Mettre à jour notre backend (profiles table)
        const response = await api.put('/users/profile', {
          name: newName.trim()
        });

        this.user = response.data;
        this.persistUserToStorage();
        return response.data;
      } catch (error) {
        console.error("Username update failed:", error);
        throw error;
      }
    },
    async checkAuth() { 
      // Cette fonction devrait être appelée au chargement de l'app (App.vue)
      const { data: { session } } = await supabase.auth.getSession();
      
      if (session?.access_token) {
          // Session Supabase active !
          
          // SECURITY: MasterKey is NOT persisted. User must re-login after page reload.
          // This prevents XSS attacks from stealing the key from storage.
          if (!this.masterKey) {
             console.warn("Session found but MasterKey missing (page reload or new tab). Redirecting to login.");
             // Force re-authentication to derive the key from password
             return false;
          }

          this.isAuthenticated = true; // Only set true if we have KEY + SESSION

          // Optimization: Don't re-fetch if we already have the user and it matches the session
          if (!this.user || this.user.id !== session.user.id) {
             await this.fetchUser();
          }

          if (this.user && this.masterKey) {
             await this.ensureRSAKeys(this.masterKey);
          }
      } else {
          // Pas de session
          this.isAuthenticated = false;
          this.user = null;
          this.masterKey = null;
      }
      return this.isAuthenticated;
    },
    async recoverAccount(email, recoveryCode, newPassword) {
        await sodium.ready;
        
        // 1. Get encrypted blob from server
        const initResponse = await api.post('/auth/recovery/init', { email });
        const { encrypted_master_key_recovery, salt } = initResponse.data;
        
        if (!encrypted_master_key_recovery) {
            throw new Error("Recovery not available for this account.");
        }

        // 2. Derive Recovery KEK locally
        const saltBytes = sodium.from_hex(salt);
        const recoveryKek = await deriveKeyFromRecoveryCode(recoveryCode, saltBytes);

        // 3. Decrypt Master Key
        let masterKey;
        try {
            masterKey = await unwrapMasterKey(encrypted_master_key_recovery, recoveryKek);
        } catch (e) {
            throw new Error("Invalid recovery code.");
        }

        // 4. Prepare new password data
        const newSalt = generateSalt();
        const newSaltHex = sodium.to_hex(newSalt);
        const newKek = await deriveKeyFromPassword(newPassword, newSalt);
        const newEncryptedMasterKey = await wrapMasterKey(masterKey, newKek);
        
        // 5. Calculate recovery hash for proof
        const recoveryHash = await hashRecoveryCode(recoveryCode);

        // 6. Send reset request
        await api.post('/auth/recovery/finish', {
            email: email,
            recovery_hash: recoveryHash,
            new_password: newPassword,
            new_salt: newSaltHex,
            new_encrypted_master_key: newEncryptedMasterKey
        });
        
        // Set up security timeout for recovered session
        this.setupSessionTimeout();
        
        return true;
    },
    // --- Storage Persistence Helpers ---
    persistUserToStorage() {
      if (this.user) {
        try {
          localStorage.setItem("safercloud_user", JSON.stringify(this.user));
        } catch (e) {
          console.error("Failed to persist user data to localStorage", e);
        }
      }
    },
    restoreUserFromStorage() {
      try {
        const storedUser = localStorage.getItem("safercloud_user");
        if (storedUser) {
          this.user = JSON.parse(storedUser);
          return true;
        }
      } catch (e) {
        console.error("Failed to restore user data from localStorage", e);
      }
      return false;
    },
    setupSessionTimeout() {
      // Clear any existing timeout
      if (this.sessionTimeoutId) {
        clearTimeout(this.sessionTimeoutId);
      }
      
      // Set 30-minute timeout for security
      const THIRTY_MINUTES = 30 * 60 * 1000;
      this.sessionTimeoutId = setTimeout(() => {
        console.warn('[Security] Session timeout - logging out');
        this.logout();
      }, THIRTY_MINUTES);
      
      console.log('[Security] Session timeout set to 30 minutes');
    },
    async updateAvatar(avatarUrl) {
      try {
        await api.put('/users/avatar', { avatar_url: avatarUrl })
        
        // Update local user state
        if (this.user) {
          this.user.avatar_url = avatarUrl
          this.persistUserToStorage()
        }
      } catch (error) {
        console.error('Failed to update avatar:', error)
        throw error
      }
    }
  },
})
