import { defineStore } from 'pinia'
import { useRealtimeStore } from './realtime'
import { useAuthStore } from './auth'
import { useUIStore } from './ui'
import sodium from 'libsodium-wrappers-sumo'
import { 
    generateMasterKey, 
    encryptKeyWithPublicKey, 
    decryptKeyWithPrivateKey,
    importKeyFromPEM
} from '../utils/crypto'
import axios from 'axios'
import { API_BASE_URL } from '../api'
import { authClient } from '../auth-client'

// Fetch ICE config from the backend (works with both Supabase and PocketBase)
async function fetchICEConfig() {
    //console.log("[P2P] Fetching ICE Config...");
    try {
        const token = await authClient.getToken();
        const response = await axios.get(`${API_BASE_URL}p2p/ice-config`, {
            headers: { Authorization: `Bearer ${token}` }
        });
        if (response.data && response.data.iceServers) {
            //console.log("[P2P] Retrieved ICE Servers:", response.data.iceServers);
            return { iceServers: response.data.iceServers };
        }
    } catch (e) {
        console.error("[P2P] Failed to fetch ICE config via API, falling back to STUN only", e);
    }
    // Fallback sécurité si l'API échoue
    return {
        iceServers: [
            { urls: 'stun:stun.l.google.com:19302' },
            { urls: 'stun:stun1.l.google.com:19302' }
        ]
    };
}

export const useP2PStore = defineStore('p2p', {
  state: () => ({
    incomingOffer: null, 
    activeTransfer: null, 
    candidateQueue: [], // Store candidates that arrive before acceptance
    heartbeatInterval: null // Interval for session maintenance
  }),
  actions: {
    startHeartbeat() {
        if (this.heartbeatInterval) return; // Already running
        
        //console.log('[P2P] Starting session heartbeat');
        // Send heartbeat every 2.5 minutes (well before 5min Redis TTL)
        this.heartbeatInterval = setInterval(async () => {
            if (!this.activeTransfer) {
                this.stopHeartbeat();
                return;
            }
            
            try {
                const token = localStorage.getItem('token');
                await axios.get(`${API_BASE_URL}/heartbeat`, {
                    headers: { Authorization: `Bearer ${token}` }
                });
                //console.log('[P2P] Session heartbeat sent');
            } catch (e) {
                console.error('[P2P] Heartbeat failed:', e);
            }
        }, 150000); // 2.5 minutes
    },
    
    stopHeartbeat() {
        if (this.heartbeatInterval) {
            //console.log('[P2P] Stopping session heartbeat');
            clearInterval(this.heartbeatInterval);
            this.heartbeatInterval = null;
        }
    },
    /** Builds the oniceconnectionstatechange handler for a peer connection. */
    _makeIceStateHandler(pc, transferId) {
        const uiStore = useUIStore();
        return () => {
            if (this.activeTransfer && this.activeTransfer.transferId === transferId) {
                this.activeTransfer.connectionInfo.iceState = pc.iceConnectionState;
                if (pc.iceConnectionState === 'checking') {
                    this.activeTransfer.connectionInfo.stage = 'Négociation de la connexion...';
                } else if (pc.iceConnectionState === 'connected') {
                    this.activeTransfer.connectionInfo.stage = 'Connecté';
                    this.detectConnectionType(pc, transferId);
                }
            }
            if (pc.iceConnectionState === 'failed' || pc.iceConnectionState === 'disconnected') {
                if (this.activeTransfer && this.activeTransfer.transferId === transferId) {
                    const isDone = this.activeTransfer.status === 'Done' || this.activeTransfer.status === 'Complete';
                    if (!isDone) {
                        console.error('ICE Connection Failed/Disconnected during active transfer');
                        uiStore.showError(
                            "La connexion P2P a échoué (ICE Failed). Un serveur TURN est nécessaire pour traverser les pare-feux et NAT stricts.",
                            'Échec de Connexion'
                        );
                        this.activeTransfer.status = 'Error';
                        this.activeTransfer.connectionInfo.stage = 'Échec de connexion';
                    }
                }
            }
        };
    },

    _handleOfferSignal(sender_id, data) {
        this.incomingOffer = { senderId: sender_id, ...data.meta, sdp: data.sdp, transferId: data.transferId };
        this.candidateQueue = [];
    },

    async _handleAnswerSignal(sender_id, data) {
        if (!this.activeTransfer || this.activeTransfer.friendId !== sender_id) return;
        if (data.transferId && this.activeTransfer.transferId && data.transferId !== this.activeTransfer.transferId) {
            console.warn('Ignoring answer for old transfer session');
            return;
        }
        const sdpInit = data.transferId ? data.sdp : data;
        await this.activeTransfer.pc.setRemoteDescription(new RTCSessionDescription(sdpInit));
    },

    async _handleCandidateSignal(sender_id, data) {
        const candidateTransferId = data.transferId;
        const candidatePayload = candidateTransferId ? data.candidate : data;

        if (this.activeTransfer && this.activeTransfer.friendId === sender_id) {
            if (this.activeTransfer.transferId && candidateTransferId && this.activeTransfer.transferId !== candidateTransferId) {
                console.warn('Ignoring candidate for old transfer session', candidateTransferId);
                return;
            }
            try {
                await this.activeTransfer.pc.addIceCandidate(new RTCIceCandidate(candidatePayload));
            } catch (e) {
                console.error('Error adding candidate', e);
            }
        } else if (this.incomingOffer && this.incomingOffer.senderId === sender_id) {
            if (this.incomingOffer.transferId && candidateTransferId && this.incomingOffer.transferId !== candidateTransferId) {
                console.warn('Ignoring queued candidate for mismatched session');
                return;
            }
            this.candidateQueue.push(candidatePayload);
        }
    },

    async handleSignal(payload) {
        const { sender_id, type, data } = payload;
        if (type === 'offer') {
            this._handleOfferSignal(sender_id, data);
        } else if (type === 'answer') {
            await this._handleAnswerSignal(sender_id, data);
        } else if (type === 'candidate') {
            await this._handleCandidateSignal(sender_id, data);
        }
    },

    async startTransfer(friend, file) {
         await sodium.ready;
         
         if (!friend.public_key) {
             alert("L'ami n'a pas de clé publique (Il doit se reconnecter une fois pour la publier)."); 
             return;
         }

         const fileKey = await generateMasterKey();
         const fileKeyRaw = await window.crypto.subtle.exportKey("raw", fileKey);
         
         const publicKey = await importKeyFromPEM(friend.public_key);
         const keyEncryptedBase64 = await encryptKeyWithPublicKey(fileKeyRaw, publicKey);

         // Fetch ICE Config from backend
         const rtcConfig = await fetchICEConfig();
         //console.log("Using RTC Config:", rtcConfig);

         const pc = new RTCPeerConnection(rtcConfig);
         const realtimeStore = useRealtimeStore();

         // Generate unique transfer ID (Polyfill for older browsers)
         const transferId = (crypto.randomUUID) ? crypto.randomUUID() : Math.random().toString(36).substring(2) + Date.now().toString(36);

         pc.oniceconnectionstatechange = this._makeIceStateHandler(pc, transferId);

         pc.onicecandidate = e => {
             if(e.candidate) {
                 realtimeStore.sendP2PSignal(friend.id, 'candidate', {
                     candidate: e.candidate,
                     transferId: transferId
                 });
             }
         };

         const dc = pc.createDataChannel("file");
         dc.binaryType = "arraybuffer";
         dc.onopen = () => this.sendFileData(file, fileKey, dc);
         
         this.activeTransfer = {
             friendId: friend.id,
             type: 'send',
             pc: pc,
             status: 'Connecting...',
             progress: 0,
             fileName: file.name,
             transferId: transferId,
             connectionInfo: {
                 stage: 'Initialisation...',
                 iceState: 'new',
                 connectionType: null,
                 usingTurn: false
             }
         };
         
         // Start heartbeat to prevent session timeout
         this.startHeartbeat();

         const offer = await pc.createOffer();
         await pc.setLocalDescription(offer);

         realtimeStore.sendP2PSignal(friend.id, 'offer', {
             sdp: offer,
             transferId: transferId,
             meta: {
                 name: file.name,
                 size: file.size,
                 type: file.type,
                 fileKeyEncrypted: keyEncryptedBase64
             }
         });
    },

    async acceptTransfer() {
        if (!this.incomingOffer) return;
        const offerData = this.incomingOffer;
        // Don't nullify yet, keep ref for transferId check if needed? 
        // Actually we copy it to activeTransfer.
        this.incomingOffer = null; 

        await sodium.ready;
        const authStore = useAuthStore();
        
        // --- GUARD: Check Keys ---
        if (!authStore.privateKey) {
            console.warn("Private key missing in store. Attempting re-restoration...");
            if (authStore.masterKey) {
                await authStore.ensureRSAKeys(authStore.masterKey);
            }
            // Double check
            if (!authStore.privateKey) {
                console.error("Critical: User has no decrypted private RSA key. Cannot accept transfer.");
                alert("Erreur de sécurité : Votre clé de chiffrement n'est pas disponible. Essayez de recharger la page ou de vous reconnecter.");
                return;
            }
        }
        // -------------------------

        let fileKey = null;
        try {
            const rawKey = await decryptKeyWithPrivateKey(offerData.fileKeyEncrypted, authStore.privateKey);
            fileKey = await window.crypto.subtle.importKey("raw", rawKey, "AES-GCM", true, ["decrypt"]);
        } catch(e) {
            console.error("Decryption failed", e);
            alert("Erreur de déchiffrement de la clé");
            return;
        }

        // Fetch ICE Config from backend
        const rtcConfig = await fetchICEConfig();
        const pc = new RTCPeerConnection(rtcConfig);
        const realtimeStore = useRealtimeStore();
        const uiStore = useUIStore();

        // We don't necessarily need to send transferId back for candidates, but good practice.
        // Or we assume candidates from receiver belong to the same session implicitly.
        // Let's attach the same ID.
        const transferId = offerData.transferId;

        pc.oniceconnectionstatechange = this._makeIceStateHandler(pc, transferId);

        pc.onicecandidate = e => {
             if(e.candidate) {
                 realtimeStore.sendP2PSignal(offerData.senderId, 'candidate', {
                     candidate: e.candidate,
                     transferId: transferId
                 });
             }
        };
        
        this.activeTransfer = {
             friendId: offerData.senderId,
             type: 'receive',
             pc: pc,
             status: 'Connecting...',
             progress: 0,
             fileName: offerData.name,
             fileSize: offerData.size,
             fileType: offerData.type,
             fileKey: fileKey,
             buffer: [],
             receivedSize: 0,
             transferId: transferId,
             connectionInfo: {
                 stage: 'Initialisation...',
                 iceState: 'new',
                 connectionType: null,
                 usingTurn: false
             }
        };
        
        // Start heartbeat to prevent session timeout
        this.startHeartbeat();


        pc.ondatachannel = (event) => {
            const dc = event.channel;
            dc.binaryType = "arraybuffer";
            dc.onmessage = (msgEvent) => this.handleReceiveMessage(msgEvent);
            dc.onopen = () => { this.activeTransfer.status = 'Receiving...'; };
        };

        await pc.setRemoteDescription(new RTCSessionDescription(offerData.sdp));
        
        // Flush queued candidates
        //console.log(`Flushing ${this.candidateQueue.length} queued candidates`);
        for (const candidateData of this.candidateQueue) {
            try {
                await pc.addIceCandidate(new RTCIceCandidate(candidateData));
            } catch(e) {
                console.error("Error flushing candidate", e);
            }
        }
        this.candidateQueue = []; // Clear queue

        const answer = await pc.createAnswer();
        await pc.setLocalDescription(answer);
        
        realtimeStore.sendP2PSignal(offerData.senderId, 'answer', {
            sdp: answer,
            transferId: transferId
        });
    },

    rejectTransfer() {
        this.incomingOffer = null;
        // Should send 'reject' signal ideally
    },

    cancelTransfer() {
        if (this.activeTransfer) {
            this.activeTransfer.pc.close();
            this.activeTransfer = null;
        }
        this.stopHeartbeat();
    },

    async sendFileData(file, key, dc) {
        const CHUNK_SIZE = 16 * 1024; // 16KB chunk size
        let offset = 0;
        let chunkIndex = 0;
        this.activeTransfer.status = 'Sending...';

        const waitForBuffer = () => {
             // Lower buffer threshold to ensures smoother flow
             if(dc.bufferedAmount > 8 * 1024 * 1024) { 
                 return new Promise(resolve => {
                     const listener = () => {
                         dc.removeEventListener('bufferedamountlow', listener);
                         resolve();
                     };
                     dc.addEventListener('bufferedamountlow', listener);
                 });
             }
             return Promise.resolve();
        };

        while(offset < file.size) {
            if (!this.activeTransfer) break; // Cancelled

            const chunk = file.slice(offset, offset + CHUNK_SIZE);
            const buffer = await chunk.arrayBuffer();
            
            const iv = window.crypto.getRandomValues(new Uint8Array(12));
            const encrypted = await window.crypto.subtle.encrypt(
                { name: "AES-GCM", iv: iv },
                key,
                buffer
            );
            
            // Packet structure: [Index (4 bytes)][IV (12 bytes)][Encrypted Data]
            const packet = new Uint8Array(4 + 12 + encrypted.byteLength);
            
            // Write Index (Big Endian)
            new DataView(packet.buffer).setUint32(0, chunkIndex, false);
            
            packet.set(iv, 4);
            packet.set(new Uint8Array(encrypted), 16);
            
            await waitForBuffer();
            dc.send(packet);
            
            offset += CHUNK_SIZE;
            chunkIndex++;
            this.activeTransfer.progress = Math.round((offset / file.size) * 100);
        }
        
         await waitForBuffer();
         dc.send(new TextEncoder().encode("EOF"));
         this.activeTransfer.status = 'Done';
         
         // Close DataChannel and PeerConnection after a short delay
         setTimeout(() => {
             if (dc && dc.readyState === 'open') {
                 dc.close();
             }
             if (pc && pc.connectionState !== 'closed') {
                 pc.close();
             }
         }, 500);
         
         setTimeout(() => {
             this.activeTransfer = null;
             this.stopHeartbeat();
         }, 2000);
    },

    async handleReceiveMessage(event) {
        const data = event.data;
        
        // Detect EOF
        if (data.byteLength === 3) {
             const text = new TextDecoder().decode(data);
             if (text === 'EOF') {
                 if (this.activeTransfer && this.activeTransfer.receivedSize >= this.activeTransfer.fileSize) {
                     this.finishReceive();
                 }
                 return;
             }
        }
        
        // Decrypt
        // Packet: [Index (4)][IV (12)][Cypher]
        if (data.byteLength < 16) return; // Too small

        const arrayBuffer = data; 
        const view = new DataView(arrayBuffer);
        const index = view.getUint32(0, false);
        
        const iv = arrayBuffer.slice(4, 16);
        const cypher = arrayBuffer.slice(16);
        
        try {
            const decrypted = await window.crypto.subtle.decrypt(
                { name: "AES-GCM", iv: iv },
                this.activeTransfer.fileKey,
                cypher
            );
            
            if (!this.activeTransfer) return;

            // Deduplication: Ignore if we already have this chunk
            if (this.activeTransfer.buffer[index]) return;

            // Store with index to handle out-of-order packets if they happen
            this.activeTransfer.buffer[index] = decrypted;
            
            this.activeTransfer.receivedSize += decrypted.byteLength;
            this.activeTransfer.progress = Math.round((this.activeTransfer.receivedSize / this.activeTransfer.fileSize) * 100);

            if (this.activeTransfer.receivedSize >= this.activeTransfer.fileSize) {
                this.finishReceive();
            }
        } catch(e) {
            console.error("Decrypt error", e);
        }
    },
    
    async detectConnectionType(pc, transferId) {
        try {
            const stats = await pc.getStats();
            let usingTurn = false;
            let connectionType = 'direct';
            
            stats.forEach(report => {
                if (report.type === 'candidate-pair' && report.state === 'succeeded') {
                    const localCandidate = stats.get(report.localCandidateId);
                    const remoteCandidate = stats.get(report.remoteCandidateId);
                    
                    if (localCandidate && localCandidate.candidateType === 'relay') {
                        usingTurn = true;
                        connectionType = 'relay (TURN)';
                    } else if (remoteCandidate && remoteCandidate.candidateType === 'relay') {
                        usingTurn = true;
                        connectionType = 'relay (TURN)';
                    } else if (localCandidate && localCandidate.candidateType === 'srflx') {
                        connectionType = 'via STUN (NAT traversal)';
                    } else if (localCandidate && localCandidate.candidateType === 'host') {
                        connectionType = 'connexion directe (LAN)';
                    }
                }
            });
            
            if (this.activeTransfer && this.activeTransfer.transferId === transferId) {
                this.activeTransfer.connectionInfo.usingTurn = usingTurn;
                this.activeTransfer.connectionInfo.connectionType = connectionType;
                
                if (usingTurn) {
                    this.activeTransfer.connectionInfo.stage = 'Connecté via serveur relais TURN';
                } else {
                    this.activeTransfer.connectionInfo.stage = `Connecté en ${connectionType}`;
                }
            }
            
            //console.log('[P2P] Connection type detected:', connectionType, 'Using TURN:', usingTurn);
        } catch (e) {
            console.error('[P2P] Failed to detect connection type:', e);
        }
    },
    
    finishReceive() {
        if (!this.activeTransfer || this.activeTransfer.status === 'Complete') return;

        //console.log("Finishing transfer. Expected:", this.activeTransfer.fileSize, "Received:", this.activeTransfer.receivedSize);
        // buffer is now a sparse array (map-like), we need to flatten it in order
        // Object.keys(buffer) handles sparse arrays but not guaranteed numeric sort.
        // But since we used numeric index assignment, we can iterate up to length.
        
        // However, a simple array with holes will work with Blob? No.
        // We need to filter empty slots or just iterate.
        // Since we trust we received everything (or mostly), let's just use the array.
        
        const blob = new Blob(this.activeTransfer.buffer, { type: this.activeTransfer.fileType });
        const url = URL.createObjectURL(blob);

        const a = document.createElement('a');
        a.href = url;
        a.download = this.activeTransfer.fileName;
        a.click();
        
        window.URL.revokeObjectURL(url);
        
        this.activeTransfer.status = 'Complete';
        
        // Close PeerConnection after successful reception
        setTimeout(() => {
            if (this.activeTransfer && this.activeTransfer.pc && this.activeTransfer.pc.connectionState !== 'closed') {
                this.activeTransfer.pc.close();
            }
        }, 500);
        
        setTimeout(() => {
            this.activeTransfer = null;
            this.stopHeartbeat();
        }, 2000);
    }
  }
})