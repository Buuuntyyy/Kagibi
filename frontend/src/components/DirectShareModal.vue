<template>
  <div v-if="isOpen" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
    <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-lg overflow-hidden flex flex-col max-h-[90vh]">
      
      <!-- Header -->
      <div class="px-6 py-4 border-b dark:border-gray-700 flex justify-between items-center bg-gray-50 dark:bg-gray-800">
        <h3 class="text-lg font-medium text-gray-900 dark:text-white">
          Partager "{{ file?.name }}" avec des amis
        </h3>
        <button @click="close" class="text-gray-400 hover:text-gray-500 dark:hover:text-gray-300">
          <span class="text-2xl">&times;</span>
        </button>
      </div>

      <!-- Body -->
      <div class="p-6 overflow-y-auto flex-1">
        <div v-if="loading" class="flex justify-center p-4">
           <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
        </div>
        
        <div v-if="!file?.encrypted_key && file?.type === 'folder'" class="p-4 mb-4 bg-yellow-50 dark:bg-yellow-900/20 text-yellow-700 dark:text-yellow-300 rounded-md text-sm">
          <strong>Note :</strong> Le partage de dossier complet (fichiers inclus) est en cours de développement. 
          Pour l'instant, seul l'accès au dossier vide sera partagé. Veuillez partager les fichiers individuellement pour qu'ils soient accessibles.
        </div>

        <div v-else-if="friends.length === 0" class="text-center text-gray-500 py-4">
          Vous n'avez pas encore d'amis à qui envoyer ce fichier.
          <br>
          <router-link to="/friends" class="text-primary-600 hover:underline">Ajouter des amis</router-link>
        </div>

        <div v-else class="space-y-4">
             <div v-for="friend in friends" :key="friend.id" 
                class="flex items-center justify-between p-3 rounded-lg border dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-750 transition-colors">
                
                <div class="flex items-center space-x-3">
                   <div class="h-10 w-10 rounded-full bg-primary-100 flex items-center justify-center text-primary-700 font-bold">
                     {{ friend.name.charAt(0).toUpperCase() }}
                   </div>
                   <div>
                     <p class="font-medium text-gray-900 dark:text-white">{{ friend.name }}</p>
                     <p class="text-xs text-gray-500">{{ friend.email }}</p>
                   </div>
                </div>

                <div v-if="!friend.public_key" class="text-xs text-red-500 flex items-center" title="Cet ami n'a pas encore configuré ses clés de sécurité">
                   <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path></svg>
                   Clé manquante
                </div>

                <button v-else 
                    @click="shareWith(friend)"
                    :disabled="sharing[friend.id]"
                    class="px-3 py-1.5 rounded-md text-sm font-medium transition-colors focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
                    :class="[
                      isShared(friend.id) 
                        ? 'bg-green-100 text-green-700 cursor-default' 
                        : 'bg-primary-600 text-white hover:bg-primary-700 disabled:opacity-50'
                    ]">
                    <span v-if="sharing[friend.id]">Chiffrement...</span>
                    <span v-else-if="isShared(friend.id)">Partagé</span>
                    <span v-else>Envoyer</span>
                </button>
             </div>
        </div>
      </div>
      
      <!-- Footer -->
      <div class="px-6 py-4 bg-gray-50 dark:bg-gray-800 border-t dark:border-gray-700 flex justify-end">
         <button @click="close" class="px-4 py-2 bg-white dark:bg-gray-700 border dark:border-gray-600 rounded-md shadow-sm text-sm font-medium text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-none">
           Fermer
         </button>
      </div>

    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, getCurrentInstance } from 'vue'
import { useFriendStore } from '../stores/friends'
import { useAuthStore } from '../stores/auth'
import { useFileStore } from '../stores/files'
import api from '../api'

// Crypto helpers
import { decryptKeyWithPrivateKey, importKeyFromPEM, encryptKeyWithPublicKey } from '../utils/crypto'
import sodium from 'libsodium-wrappers-sumo'

const props = defineProps({
  isOpen: Boolean,
  file: Object // The file object to share
})

const emit = defineEmits(['close'])

const friendStore = useFriendStore()
const authStore = useAuthStore()
const fileStore = useFileStore()

const sharing = ref({}) // Map [friendId] -> boolean (loading state)
const sharedStatus = ref({}) // Map [friendId] -> boolean (success state)

const friends = computed(() => friendStore.acceptedFriends)
const loading = computed(() => friendStore.loading)

const isShared = (friendId) => {
    // Check local tracking or maybe future API check
    return sharedStatus.value[friendId];
}

const close = () => {
    emit('close')
    sharedStatus.value = {}
}

const shareWith = async (friend) => {
    if (!props.file || !friend.public_key) return;
    
    sharing.value[friend.id] = true;
    try {
        await sodium.ready;
        
        let encryptedKeyForFriend = "";
        const resourceType = (props.file.type === 'folder' || props.file.is_dir) ? 'folder' : 'file';

        if (resourceType === 'file') {
             if (!props.file.encrypted_key) {
                  throw new Error("Clé du fichier manquante. Impossible de partager.");
             }
             
             // 1. Decrypt user's Master Key (Available in AuthStore)
             if (!authStore.privateKey) {
                  throw new Error("Clé privée non disponible. Veuillez vous reconnecter.");
             }

             const fileKeyEncryptedBytes = sodium.from_base64(props.file.encrypted_key);
             const iv = fileKeyEncryptedBytes.slice(0, 12);
             const data = fileKeyEncryptedBytes.slice(12);
             
             const fileKeyRawBuffer = await window.crypto.subtle.decrypt(
                 { name: "AES-GCM", iv: iv },
                 authStore.masterKey,
                 data
             );

             // 3. Import Friend's Public Key
             const friendPublicKey = await importKeyFromPEM(friend.public_key, 'spki');

             // 4. Encrypt the FileKey with Friend's Public Key
             encryptedKeyForFriend = await encryptKeyWithPublicKey(fileKeyRawBuffer, friendPublicKey);
        } else {
             // For folders, we don't share a key yet (or recursive logic needed)
        }
        
        // 5. Send to Server
        await api.post('/shares/direct', {
            resource_id: props.file.ID || props.file.id,
            resource_type: resourceType,
            friend_id: friend.id,
            encrypted_key: encryptedKeyForFriend,
            permission: 'read'
        });

        sharedStatus.value[friend.id] = true;
        // Optional: Toast success

    } catch (e) {
        console.error("Partage échoué:", e);
        alert("Erreur: " + e.message);
    } finally {
        sharing.value[friend.id] = false;
    }
}

onMounted(() => {
    if (friends.value.length === 0) {
        friendStore.fetchFriends();
    }
})

</script>
