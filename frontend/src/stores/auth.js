import { defineStore } from 'pinia'
import api from '../api'
import router from '../router'
import { deriveKeyFromPassword, generateSalt, generateMasterKey, wrapMasterKey, unwrapMasterKey, generateRecoveryCode, deriveKeyFromRecoveryCode, hashRecoveryCode } from '../utils/crypto'
import sodium from 'libsodium-wrappers-sumo'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    isAuthenticated: false,
    user: null,
  }),
  actions: {
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
          
        } else {
          this.masterKey = null;
          console.warn("No salt received from server during login.");
        }

        this.isAuthenticated = true;

        await this.fetchUser();
        router.push({ name: 'Dashboard' });
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

      const payload = {
        name: username,
        email: email,
        password: password,
        salt: saltHex,
        encrypted_master_key: wrappedMasterKey,
        encrypted_master_key_recovery: wrappedMasterKeyRecovery,
        recovery_hash: recoveryHash,
        recovery_salt: saltHex // Use same salt initially
      };
      await api.post('/auth/register', payload)
      
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
        router.push({ name: 'Login' });
      }
    },
    async checkAuth() {
      try {
        const response = await api.get('/users/me');
        this.isAuthenticated = true;
        this.user = response.data;
        // Note: Au refresh de la page, masterKey est perdu (c'est voulu pour la sécurité).
        // L'utilisateur devra peut-être se reconnecter ou retaper son mot de passe pour déchiffrer.
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
