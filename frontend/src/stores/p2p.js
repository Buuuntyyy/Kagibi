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
    rejectedTransfer: null, // Info shown to initiator when recipient rejects
    candidateQueue: [], // Store candidates that arrive before acceptance
    heartbeatInterval: null, // Interval for session maintenance
    pingCount: 0, // Number of pings sent for current transfer
    pingCooldownUntil: null, // Timestamp until which pings are throttled
    pingSeq: 0, // Incremented on every received ping — watcher always fires
    connectingTimeout: null, // Timeout that dismisses a stuck Connecting... state
    resumeTimeout: null,     // Timeout waiting for a resume offer on the receiver side
    // Invite system: sender keeps the file + key in memory while waiting for the recipient
    pendingInvite: null, // { transferId, file, fileKey, recipientId, recipientPublicKey, recipientName }
    inviteReady: false,  // true once the recipient accepted → sender can start transfer
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
                const token = await authClient.getToken();
                await axios.get(`${API_BASE_URL}heartbeat`, {
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
                } else if (pc.iceConnectionState === 'connected' || pc.iceConnectionState === 'completed') {
                    this._clearConnectingTimeout(); // Connection established — no longer waiting
                    this.activeTransfer.connectionInfo.wasConnected = true;
                    if (pc.iceConnectionState === 'connected') {
                        this.activeTransfer.connectionInfo.stage = 'Connecté';
                        this.detectConnectionType(pc, transferId);
                    }
                }
            }
            if (pc.iceConnectionState === 'disconnected') {
                // 'disconnected' is a transient state — WebRTC may recover automatically.
                // Only update the stage label; never show an error dialog here.
                if (this.activeTransfer && this.activeTransfer.transferId === transferId) {
                    const isDone = this.activeTransfer.status === 'Done' || this.activeTransfer.status === 'Complete';
                    if (!isDone) {
                        this.activeTransfer.connectionInfo.stage = 'Reconnexion en cours...';
                    }
                }
            } else if (pc.iceConnectionState === 'failed') {
                if (this.activeTransfer && this.activeTransfer.transferId === transferId) {
                    const isDone = this.activeTransfer.status === 'Done' || this.activeTransfer.status === 'Complete';
                    if (!isDone) {
                        const MAX_RESUME = 3;
                        const attempts = this.activeTransfer.resumeAttempts ?? 0;
                        if (attempts < MAX_RESUME) {
                            if (this.activeTransfer.type === 'receive') {
                                // Receiver drives resume negotiation
                                this._initiateResume(transferId);
                            } else {
                                // Sender waits for the receiver's resume request (30 s window)
                                this.activeTransfer.status = 'Reconnecting...';
                                this.activeTransfer.connectionInfo.stage = 'Reconnexion en cours...';
                                this._senderResumeTimeout = setTimeout(() => {
                                    if (this.activeTransfer && this.activeTransfer.transferId === transferId &&
                                        this.activeTransfer.status === 'Reconnecting...') {
                                        this.activeTransfer.status = 'Error';
                                        this.activeTransfer.connectionInfo.stage = 'Échec de connexion';
                                        uiStore.showError(
                                            "Le contact distant s'est déconnecté et n'a pas pu reprendre le transfert.",
                                            'Contact déconnecté'
                                        );
                                    }
                                    this._senderResumeTimeout = null;
                                }, 30000);
                            }
                        } else {
                            // Exhausted retries — give up
                            const wasEverConnected = this.activeTransfer.connectionInfo.wasConnected || attempts > 0;
                            if (wasEverConnected) {
                                console.error('ICE Connection Failed — remote peer likely disconnected');
                                uiStore.showError(
                                    "Le contact distant s'est déconnecté pendant le transfert. Le fichier n'a pas été transféré complètement.",
                                    'Contact déconnecté'
                                );
                            } else {
                                console.error('ICE Connection Failed — could not establish connection');
                                uiStore.showError(
                                    "La connexion P2P a échoué (ICE Failed). Un serveur TURN est nécessaire pour traverser les pare-feux et NAT stricts.",
                                    'Échec de Connexion'
                                );
                            }
                            this.activeTransfer.status = 'Error';
                            this.activeTransfer.connectionInfo.stage = 'Échec de connexion';
                        }
                    }
                }
            }
        };
    },

    _handleOfferSignal(sender_id, data) {
        this.incomingOffer = { senderId: sender_id, ...data.meta, sdp: data.sdp, transferId: data.transferId };
        this.candidateQueue = [];
    },

    _handlePingSignal() {
        this.pingSeq++;
    },

    sendPing(friendId, transferId) {
        const now = Date.now();
        if (this.pingCount >= 3) return;
        if (this.pingCooldownUntil && now < this.pingCooldownUntil) return;
        const realtimeStore = useRealtimeStore();
        realtimeStore.sendP2PSignal(friendId, 'p2p_ping', { transferId });
        this.pingCount++;
        this.pingCooldownUntil = now + 30000;
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

    _clearConnectingTimeout() {
        if (this.connectingTimeout) { clearTimeout(this.connectingTimeout); this.connectingTimeout = null; }
    },

    _handleRejectSignal(sender_id, data) {
        // Only the sender side has type='send'; the receiver has no activeTransfer yet.
        if (!this.activeTransfer || this.activeTransfer.type !== 'send') return;
        // If both sides carry a transferId it must match — prevents stale/wrong-session signals.
        if (data?.transferId && this.activeTransfer.transferId && data.transferId !== this.activeTransfer.transferId) return;
        // Sender must be our intended recipient
        if (this.activeTransfer.friendId !== sender_id) {
            console.warn('[P2P] Reject signal ignored — sender_id mismatch:', sender_id, 'expected:', this.activeTransfer.friendId);
            return;
        }

        const fileName = this.activeTransfer.fileName;
        // Null out all handlers before closing to prevent stale callbacks
        const pc = this.activeTransfer.pc;
        pc.onicecandidate = null;
        pc.oniceconnectionstatechange = null;
        pc.ondatachannel = null;
        pc.close();
        this._clearConnectingTimeout();
        this.activeTransfer = null;
        this.stopHeartbeat();

        this.rejectedTransfer = { friendId: sender_id, fileName };
        // Auto-dismiss after 6 seconds
        setTimeout(() => { this.rejectedTransfer = null; }, 6000);
    },

    async handleSignal(payload) {
        const { sender_id, type, data } = payload;
        if (type === 'offer') {
            this._handleOfferSignal(sender_id, data);
        } else if (type === 'answer') {
            await this._handleAnswerSignal(sender_id, data);
        } else if (type === 'candidate') {
            await this._handleCandidateSignal(sender_id, data);
        } else if (type === 'reject') {
            this._handleRejectSignal(sender_id, data);
        } else if (type === 'p2p_ping') {
            this._handlePingSignal();
        } else if (type === 'p2p_resume_request') {
            await this._handleResumeRequestSignal(sender_id, data);
        } else if (type === 'p2p_resume_offer') {
            await this._handleResumeOfferSignal(sender_id, data);
        } else if (type === 'invite_accepted') {
            await this._handleInviteAcceptedSignal(sender_id, data);
        }
    },

    // Called when the recipient opens the invite link and accepts.
    // If the pendingInvite matches the transfer_id, kick off the WebRTC handshake.
    async _handleInviteAcceptedSignal(sender_id, data) {
        if (!this.pendingInvite) return;
        if (data?.transfer_id && data.transfer_id !== this.pendingInvite.transferId) return;
        if (sender_id !== this.pendingInvite.recipientId) return;

        this.inviteReady = true;
        const { file, fileKey, recipientId, recipientName } = this.pendingInvite;

        // For guest invites, public_key comes from the signal payload (generated by the guest browser).
        // For account-based invites, fall back to the key stored at invite creation time.
        const publicKey = data?.public_key || this.pendingInvite.recipientPublicKey;

        const pseudoFriend = {
            id: recipientId,
            name: recipientName,
            public_key: publicKey,
        };
        await this.startTransfer(pseudoFriend, file, fileKey);
    },

    // Prepare an invite-based pending transfer. The file and pre-generated key are
    // held in memory until the recipient accepts (up to 24 h, page must stay open).
    setPendingInvite({ transferId, file, fileKey, recipientId, recipientPublicKey, recipientName }) {
        this.pendingInvite = { transferId, file, fileKey, recipientId, recipientPublicKey, recipientName };
        this.inviteReady = false;
    },

    clearPendingInvite() {
        this.pendingInvite = null;
        this.inviteReady = false;
    },

    // preGeneratedKey: optional CryptoKey already created (invite flow).
    async startTransfer(friend, file, preGeneratedKey = null) {
         await sodium.ready;

         if (!friend.public_key) {
             alert("L'ami n'a pas de clé publique (Il doit se reconnecter une fois pour la publier).");
             return;
         }

         const fileKey = preGeneratedKey ?? await generateMasterKey();
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
         dc.onopen = () => this.sendFileData(file, fileKey, dc, 0);
         
         this.pingCount = 0;
         this.pingCooldownUntil = null;
         if (this.connectingTimeout) { clearTimeout(this.connectingTimeout); this.connectingTimeout = null; }

         this.activeTransfer = {
             friendId: friend.id,
             type: 'send',
             pc: pc,
             status: 'Connecting...',
             progress: 0,
             fileName: file.name,
             transferId: transferId,
             totalBytes: file.size,
             transferredBytes: 0,
             transferStartedAt: null,
             file: file,       // kept for resume
             fileKey: fileKey, // kept for resume
             resumeAttempts: 0,
             sendGeneration: 0, // incremented on each resume to stop stale send loops
             connectionInfo: {
                 stage: 'Initialisation...',
                 iceState: 'new',
                 connectionType: null,
                 usingTurn: false,
                 wasConnected: false
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

         // Auto-dismiss the waiting state after 3 minutes with no response
         this.connectingTimeout = setTimeout(() => {
             if (this.activeTransfer && this.activeTransfer.status === 'Connecting...') {
                 const pc = this.activeTransfer.pc;
                 if (pc) { pc.onicecandidate = null; pc.oniceconnectionstatechange = null; pc.close(); }
                 this.activeTransfer = null;
                 this.stopHeartbeat();
                 this.rejectedTransfer = { friendId: friend.id, fileName: file.name, timedOut: true };
                 setTimeout(() => { this.rejectedTransfer = null; }, 8000);
             }
             this.connectingTimeout = null;
         }, 3 * 60 * 1000);
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
             totalBytes: offerData.size,
             transferredBytes: 0,
             transferStartedAt: null,
             resumeAttempts: 0,
             connectionInfo: {
                 stage: 'Initialisation...',
                 iceState: 'new',
                 connectionType: null,
                 usingTurn: false,
                 wasConnected: false
             }
        };

        // Start heartbeat to prevent session timeout
        this.startHeartbeat();


        pc.ondatachannel = (event) => {
            const dc = event.channel;
            dc.binaryType = "arraybuffer";
            dc.onmessage = (msgEvent) => this.handleReceiveMessage(msgEvent);
            dc.onopen = () => {
                this.activeTransfer.status = 'Receiving...';
                this.activeTransfer.transferStartedAt = Date.now();
            };
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

    async rejectTransfer() {
        if (!this.incomingOffer) return;
        const realtimeStore = useRealtimeStore();
        const senderId = this.incomingOffer.senderId;
        const transferId = this.incomingOffer.transferId;
        // Clear local state immediately so the UI updates right away
        this.incomingOffer = null;
        this.candidateQueue = [];
        // Send the reject signal — best-effort, errors are logged but don't block the UI
        try {
            await realtimeStore.sendP2PSignal(senderId, 'reject', { transferId });
        } catch (err) {
            console.error('[P2P] Failed to send reject signal:', err);
        }
    },

    cancelTransfer() {
        this._clearConnectingTimeout();
        this._clearResumeTimeout();
        if (this._senderResumeTimeout) { clearTimeout(this._senderResumeTimeout); this._senderResumeTimeout = null; }
        if (this.activeTransfer) {
            const pc = this.activeTransfer.pc;
            if (pc) { pc.onicecandidate = null; pc.oniceconnectionstatechange = null; pc.ondatachannel = null; pc.close(); }
            this.activeTransfer = null;
        }
        this.stopHeartbeat();
    },

    _clearResumeTimeout() {
        if (this.resumeTimeout) { clearTimeout(this.resumeTimeout); this.resumeTimeout = null; }
    },

    // Returns the index of the first missing chunk (= next chunk the sender should send).
    _getResumePoint(buffer) {
        let i = 0;
        while (buffer[i] !== undefined) i++;
        return i;
    },

    // Called by the receiver when ICE fails during an active transfer.
    _initiateResume(transferId) {
        if (!this.activeTransfer || this.activeTransfer.transferId !== transferId) return;
        const transfer = this.activeTransfer;
        transfer.resumeAttempts++;
        const resumeFrom = this._getResumePoint(transfer.buffer);
        transfer.status = 'Reconnecting...';
        transfer.connectionInfo.stage = `Reprise depuis le paquet ${resumeFrom}…`;

        const realtimeStore = useRealtimeStore();
        realtimeStore.sendP2PSignal(transfer.friendId, 'p2p_resume_request', {
            transferId,
            resumeFrom,
        });

        // Give up if the sender never replies with a resume offer
        this._clearResumeTimeout();
        this.resumeTimeout = setTimeout(() => {
            this.resumeTimeout = null;
            if (this.activeTransfer && this.activeTransfer.transferId === transferId &&
                this.activeTransfer.status === 'Reconnecting...') {
                this.activeTransfer.status = 'Error';
                this.activeTransfer.connectionInfo.stage = 'Échec de connexion';
                const uiStore = useUIStore();
                uiStore.showError(
                    "Le contact distant ne répond plus. Le transfert n'a pas pu reprendre.",
                    'Contact déconnecté'
                );
            }
        }, 30000);
    },

    // Sender: receiver asked to resume from a given chunk index.
    async _handleResumeRequestSignal(sender_id, data) {
        if (!this.activeTransfer || this.activeTransfer.type !== 'send') return;
        if (this.activeTransfer.friendId !== sender_id) return;
        if (data.transferId !== this.activeTransfer.transferId) return;
        if (this.activeTransfer.status === 'Error') return;

        const MAX_RESUME = 3;
        if (this.activeTransfer.resumeAttempts >= MAX_RESUME) return;
        this.activeTransfer.resumeAttempts++;

        // Cancel the sender-side give-up timeout — the receiver is alive
        if (this._senderResumeTimeout) { clearTimeout(this._senderResumeTimeout); this._senderResumeTimeout = null; }

        // Tear down the old PC cleanly before creating a new one
        const oldPc = this.activeTransfer.pc;
        if (oldPc) {
            oldPc.onicecandidate = null;
            oldPc.oniceconnectionstatechange = null;
            oldPc.ondatachannel = null;
            try { oldPc.close(); } catch (_) {}
        }

        // Increment generation so any still-running sendFileData loop will exit
        this.activeTransfer.sendGeneration++;

        await this._createResumeOffer(sender_id, data.resumeFrom);
    },

    // Sender: build a new RTCPeerConnection and send a resume offer.
    async _createResumeOffer(recipientId, fromChunk) {
        const transfer = this.activeTransfer;
        if (!transfer) return;

        const transferId = transfer.transferId;
        transfer.connectionInfo.stage = `Reprise depuis le paquet ${fromChunk}…`;
        transfer.connectionInfo.wasConnected = false;
        transfer.connectionInfo.iceState = 'new';

        const rtcConfig = await fetchICEConfig();
        const pc = new RTCPeerConnection(rtcConfig);
        transfer.pc = pc;

        const realtimeStore = useRealtimeStore();
        pc.oniceconnectionstatechange = this._makeIceStateHandler(pc, transferId);
        pc.onicecandidate = e => {
            if (e.candidate) {
                realtimeStore.sendP2PSignal(recipientId, 'candidate', {
                    candidate: e.candidate,
                    transferId,
                });
            }
        };

        const dc = pc.createDataChannel("file");
        dc.binaryType = "arraybuffer";
        const capturedGeneration = transfer.sendGeneration;
        dc.onopen = () => {
            // Only start sending if this DC still belongs to the current generation
            if (this.activeTransfer && this.activeTransfer.sendGeneration === capturedGeneration) {
                this.sendFileData(transfer.file, transfer.fileKey, dc, fromChunk);
            }
        };

        const offer = await pc.createOffer();
        await pc.setLocalDescription(offer);

        realtimeStore.sendP2PSignal(recipientId, 'p2p_resume_offer', {
            sdp: offer,
            transferId,
            resumeFrom: fromChunk,
        });
    },

    // Receiver: sender replied with a fresh offer to resume from a given chunk.
    async _handleResumeOfferSignal(sender_id, data) {
        if (!this.activeTransfer || this.activeTransfer.type !== 'receive') return;
        if (this.activeTransfer.friendId !== sender_id) return;
        if (data.transferId !== this.activeTransfer.transferId) return;
        if (this.activeTransfer.status !== 'Reconnecting...') return;

        this._clearResumeTimeout();

        const transfer = this.activeTransfer;
        const transferId = transfer.transferId;

        // Tear down the old PC
        const oldPc = transfer.pc;
        if (oldPc) {
            oldPc.onicecandidate = null;
            oldPc.oniceconnectionstatechange = null;
            oldPc.ondatachannel = null;
            try { oldPc.close(); } catch (_) {}
        }

        transfer.connectionInfo.wasConnected = false;
        transfer.connectionInfo.iceState = 'new';
        transfer.connectionInfo.stage = `Reprise depuis le paquet ${data.resumeFrom}…`;

        const rtcConfig = await fetchICEConfig();
        const pc = new RTCPeerConnection(rtcConfig);
        transfer.pc = pc;

        const realtimeStore = useRealtimeStore();
        pc.oniceconnectionstatechange = this._makeIceStateHandler(pc, transferId);
        pc.onicecandidate = e => {
            if (e.candidate) {
                realtimeStore.sendP2PSignal(sender_id, 'candidate', {
                    candidate: e.candidate,
                    transferId,
                });
            }
        };

        pc.ondatachannel = (event) => {
            const dc = event.channel;
            dc.binaryType = "arraybuffer";
            dc.onmessage = (msgEvent) => this.handleReceiveMessage(msgEvent);
            dc.onopen = () => {
                // Restore receiving status without resetting progress or transferStartedAt
                if (this.activeTransfer && this.activeTransfer.transferId === transferId) {
                    this.activeTransfer.status = 'Receiving...';
                }
            };
        };

        await pc.setRemoteDescription(new RTCSessionDescription(data.sdp));
        this.candidateQueue = []; // stale candidates from old session are irrelevant

        const answer = await pc.createAnswer();
        await pc.setLocalDescription(answer);

        realtimeStore.sendP2PSignal(sender_id, 'answer', {
            sdp: answer,
            transferId,
        });
    },

    async sendFileData(file, key, dc, startChunk = 0) {
        const CHUNK_SIZE = 16 * 1024; // 16KB chunk size
        let offset = startChunk * CHUNK_SIZE;
        let chunkIndex = startChunk;

        // Capture the generation at entry — if the caller creates a new PC/DC for a
        // resume, it increments sendGeneration before calling us again, so the old
        // invocation detects the mismatch and exits cleanly.
        const myGeneration = this.activeTransfer.sendGeneration;

        this.activeTransfer.status = 'Sending...';
        if (startChunk === 0) {
            this.activeTransfer.transferStartedAt = Date.now();
        }

        const waitForBuffer = () => {
             if (dc.bufferedAmount > 8 * 1024 * 1024) {
                 return new Promise(resolve => {
                     const done = () => {
                         dc.removeEventListener('bufferedamountlow', done);
                         dc.removeEventListener('close', done);
                         dc.removeEventListener('error', done);
                         resolve();
                     };
                     dc.addEventListener('bufferedamountlow', done);
                     // Resolve immediately if the channel closes so the loop can exit
                     dc.addEventListener('close', done);
                     dc.addEventListener('error', done);
                 });
             }
             return Promise.resolve();
        };

        while (offset < file.size) {
            if (!this.activeTransfer) break; // Cancelled
            if (this.activeTransfer.sendGeneration !== myGeneration) break; // Superseded by resume

            const chunk = file.slice(offset, offset + CHUNK_SIZE);
            const buffer = await chunk.arrayBuffer();

            if (!this.activeTransfer || this.activeTransfer.sendGeneration !== myGeneration) break;

            const iv = window.crypto.getRandomValues(new Uint8Array(12));
            const encrypted = await window.crypto.subtle.encrypt(
                { name: "AES-GCM", iv: iv },
                key,
                buffer
            );

            if (!this.activeTransfer || this.activeTransfer.sendGeneration !== myGeneration) break;

            // Packet structure: [Index (4 bytes big-endian)][IV (12 bytes)][Encrypted Data]
            const packet = new Uint8Array(4 + 12 + encrypted.byteLength);
            new DataView(packet.buffer).setUint32(0, chunkIndex, false);
            packet.set(iv, 4);
            packet.set(new Uint8Array(encrypted), 16);

            await waitForBuffer();

            if (!this.activeTransfer || this.activeTransfer.sendGeneration !== myGeneration) break;

            try {
                dc.send(packet);
            } catch (_) {
                break; // DataChannel closed under us
            }

            offset += CHUNK_SIZE;
            chunkIndex++;
            this.activeTransfer.progress = Math.round((offset / file.size) * 100);
            this.activeTransfer.transferredBytes = Math.min(offset, file.size);
        }

        // Only finalise if this invocation is still the active one
        if (!this.activeTransfer || this.activeTransfer.sendGeneration !== myGeneration) return;

        await waitForBuffer();
        try { dc.send(new TextEncoder().encode("EOF")); } catch (_) {}
        this.activeTransfer.status = 'Done';

        setTimeout(() => {
            if (dc && dc.readyState === 'open') dc.close();
            if (this.activeTransfer && this.activeTransfer.sendGeneration === myGeneration) {
                const pc = this.activeTransfer.pc;
                if (pc && pc.connectionState !== 'closed') pc.close();
            }
        }, 500);

        setTimeout(() => {
            if (this.activeTransfer && this.activeTransfer.sendGeneration === myGeneration) {
                this.activeTransfer = null;
                this.stopHeartbeat();
            }
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
            this.activeTransfer.transferredBytes = this.activeTransfer.receivedSize;
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