import { defineStore } from 'pinia'
import api from '../api'
import router from '../router'
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
        const authentication_response = await api.post('/auth/login', credentials, {
          headers: {
            'Content-Type': 'application/json'
          }
        });

        const { salt, encrypted_master_key } = authentication_response.data;

        if (salt && encrypted_master_key) {
          await sodium.ready;
          const saltBytes = sodium.from_hex(salt);
          const kek = await deriveKeyFromPassword(credentials.password, saltBytes);
          this.masterKey = await unwrapMasterKey(encrypted_master_key, kek);
          
          // Persist key for page reload (SessionStorage)
          try {
            const exportedKey = await window.crypto.subtle.exportKey("jwk", this.masterKey);
            sessionStorage.setItem("safercloud_mk", JSON.stringify(exportedKey));
          } catch (e) {
            console.error("Failed to persist master key", e);
          }

          // Generate/Load RSA Keys (New functionality)
          await this.fetchUser(); // Get latest user data including keys
          await this.ensureRSAKeys(this.masterKey);
          
        } else {
          this.masterKey = null;
          console.warn("No salt received from server during login.");
        }

        this.isAuthenticated = true;

        await this.fetchUser();
        router.push({ name: 'Home' });
        return true
      } catch (error) {
        console.error("Login failed:", error)
        this.isAuthenticated = false;
        this.user = null;
        return false
      }
    },
    async register(username, email, password) {
      await sodium.ready;
      const salt = generateSalt();
      const saltHex = sodium.to_hex(salt);

      const masterKey = await generateMasterKey();
      const kek = await deriveKeyFromPassword(password, salt);
      const wrappedMasterKey = await wrapMasterKey(masterKey, kek);

      // Generate Recovery Code
      const recoveryCode = generateRecoveryCode();
      const recoveryHash = await hashRecoveryCode(recoveryCode);
      
      // Encrypt Master Key with Recovery Code
      // We use the SAME salt for simplicity, or we could generate a specific one.
      // Using the same salt is fine as long as the recovery code is high entropy.
      const recoveryKek = await deriveKeyFromRecoveryCode(recoveryCode, salt);
      const wrappedMasterKeyRecovery = await wrapMasterKey(masterKey, recoveryKek);

      // Generate RSA Keys for new user
      const keyPair = await generateRSAKeyPair();
      const publicKeyPEM = await exportKeyToPEM(keyPair.publicKey, 'spki');
      const encryptedPrivateKey = await encryptPrivateKey(keyPair.privateKey, masterKey);

      const payload = {
        name: username,
        email: email,
        password: password,
        salt: saltHex,
        encrypted_master_key: wrappedMasterKey,
        encrypted_master_key_recovery: wrappedMasterKeyRecovery,
        recovery_hash: recoveryHash,
        recovery_salt: saltHex, // Use same salt initially
        public_key: publicKeyPEM,
        encrypted_private_key: encryptedPrivateKey
      };
      try {
        await api.post('/auth/register', payload)
      } catch (err) {
        // Surface backend error message if available
        if (err.response && err.response.data && err.response.data.error) {
          throw new Error(err.response.data.error)
        } else {
          throw new Error("Erreur lors de l'inscription")
        }
      }
      return recoveryCode; // Return code so UI can display it
    },
    async logout() {
      try {
        await api.post('/auth/logout');
      } catch (error) {
        console.error("Logout failed:", error)
      } finally {
        this.isAuthenticated = false;
        this.user = null;
        this.masterKey = null;
        sessionStorage.removeItem("safercloud_mk");
        router.push({ name: 'Login' });
      }
    },
    async checkAuth() {
      try {
        await sodium.ready; // Ensure sodium is ready for key restoration

        const response = await api.get('/users/me'); 
        // Load user data first
        this.user = response.data;
        
        // Restore master key from session if available
        if (!this.masterKey) {
          const storedKey = sessionStorage.getItem("safercloud_mk");
          if (storedKey) {
            try {
              const jwk = JSON.parse(storedKey);
              this.masterKey = await window.crypto.subtle.importKey(
                "jwk",
                jwk,
                { name: "AES-GCM" },
                true,
                ["encrypt", "decrypt"]
              );
            } catch (e) {
              console.error("Failed to restore master key from session", e);
            }
          }
        }

        // Restore RSA keys if master key is available
        if (this.masterKey) {
             await this.ensureRSAKeys(this.masterKey);
        }

        // Set authenticated only after keys are potentially restored
        this.isAuthenticated = true;

        return true

      } catch (error) {
        this.isAuthenticated = false;
        this.user = null;
        this.masterKey = null;
        return false
      }
    },
    async fetchUser() {
      try {
        const response = await api.get('/users/me');
        this.user = response.data;
      } catch (error) {
        console.error("Failed to fetch user:", error)
        this.logout();
      }
    },
    async updatePassword(currentPassword, newPassword) {
      if (!this.masterKey) {
        throw new Error("Master key not available. Please re-login.");
      }

      await sodium.ready;
      
      // 1. Generate new salt
      const newSalt = generateSalt();
      const newSaltHex = sodium.to_hex(newSalt);

      // 2. Derive new KEK
      const newKek = await deriveKeyFromPassword(newPassword, newSalt);

      // 3. Re-encrypt master key
      const newEncryptedMasterKey = await wrapMasterKey(this.masterKey, newKek);

      // 4. Send to backend
      await api.post('/users/change-password', {
        current_password: currentPassword,
        new_password: newPassword,
        new_salt: newSaltHex,
        new_encrypted_master_key: newEncryptedMasterKey
      });
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
        
        return true;
    }
  },
})
