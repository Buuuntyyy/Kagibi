<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div v-if='visible' class='p2p-notification-container' :class="{ minimized: isMinimized }">
    <div class='p2p-card' :class="{ shaking: shakeCard }">
      <div class='card-header' @click.self="isMinimized && (isMinimized = false)">
        <h3 class='header-title'>
          <svg xmlns='http://www.w3.org/2000/svg' width='18' height='18' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'><circle cx='12' cy='12' r='10'></circle><line x1='2' y1='12' x2='22' y2='12'></line><path d='M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z'></path></svg>
          {{ t('p2p.title') }}
          <!-- Compact progress shown when minimized -->
          <span v-if="isMinimized && p2pStore.activeTransfer" class="mini-progress">
            {{ p2pStore.activeTransfer.progress }}%
          </span>
        </h3>
        <div class="header-actions">
          <button class='minimize-icon' @click='isMinimized = !isMinimized' :title="isMinimized ? 'Agrandir' : 'Réduire'">
            <svg v-if="isMinimized" xmlns='http://www.w3.org/2000/svg' width='16' height='16' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='2.5' stroke-linecap='round' stroke-linejoin='round'><polyline points='18 15 12 9 6 15'></polyline></svg>
            <svg v-else xmlns='http://www.w3.org/2000/svg' width='16' height='16' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='2.5' stroke-linecap='round' stroke-linejoin='round'><polyline points='6 9 12 15 18 9'></polyline></svg>
          </button>
          <button v-if='canClose' @click='close' class='close-icon'>&times;</button>
        </div>
      </div>
      
      <!-- Card body hidden when minimized -->
      <template v-if="!isMinimized">

      <!-- INCOMING REQUEST -->
      <div v-if='p2pStore.incomingOffer' class='notification-body'>
         <p class='request-text'>
            {{ t('p2p.incomingRequest', { sender: senderName }) }}
         </p>
         <div class='file-preview'>
            <div class='file-icon-box'>
                <svg xmlns='http://www.w3.org/2000/svg' width='24' height='24' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'><path d='M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z'></path><polyline points='13 2 13 9 20 9'></polyline></svg>
            </div>
            <div class='file-info'>
                <div class='f-name' :title='p2pStore.incomingOffer.name'>{{ p2pStore.incomingOffer.name }}</div>
                <div class='f-size'>{{ formatSize(p2pStore.incomingOffer.size) }}</div>
            </div>
         </div>
         <div class='actions-grid'>
            <button @click='reject' class='btn btn-secondary'>{{ t('common.refuse') }}</button>
            <button @click='accept' class='btn btn-primary'>{{ t('common.accept') }}</button>
         </div>
      </div>

      <!-- REJECTED BY RECIPIENT (or timed out) -->
      <div v-else-if='p2pStore.rejectedTransfer' class='notification-body'>
        <div class='rejected-notice'>
          <svg xmlns='http://www.w3.org/2000/svg' width='20' height='20' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'><circle cx='12' cy='12' r='10'></circle><line x1='15' y1='9' x2='9' y2='15'></line><line x1='9' y1='9' x2='15' y2='15'></line></svg>
          <div>
            <p class='rejected-title'>{{ p2pStore.rejectedTransfer.timedOut ? t('p2p.transferTimedOut') : t('p2p.transferRejected') }}</p>
            <p class='rejected-file'>{{ p2pStore.rejectedTransfer.fileName }}</p>
          </div>
        </div>
        <div class='actions-grid single'>
          <button @click='p2pStore.rejectedTransfer = null' class='btn btn-secondary'>{{ t('common.close') }}</button>
        </div>
      </div>

      <!-- ACTIVE TRANSFER -->
      <div v-else-if='p2pStore.activeTransfer' class='notification-body'>
         <!-- Status row -->
         <div class='status-header'>
             <span class='status-label status-icon-row' :class='statusLabelClass'>
                 <!-- Spinner: connecting or reconnecting -->
                 <svg v-if='isConnecting || isReconnecting' class='spin-anim' xmlns='http://www.w3.org/2000/svg' width='13' height='13' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='2.5' stroke-linecap='round' stroke-linejoin='round'><path d='M21 12a9 9 0 1 1-6.219-8.56'/></svg>
                 <!-- Checkmark: done -->
                 <svg v-else-if='isDone' xmlns='http://www.w3.org/2000/svg' width='13' height='13' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='2.5' stroke-linecap='round' stroke-linejoin='round'><polyline points='20 6 9 17 4 12'/></svg>
                 <!-- Error X -->
                 <svg v-else-if='isError' xmlns='http://www.w3.org/2000/svg' width='13' height='13' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='2.5' stroke-linecap='round' stroke-linejoin='round'><line x1='18' y1='6' x2='6' y2='18'/><line x1='6' y1='6' x2='18' y2='18'/></svg>
                 {{ statusText }}
             </span>
             <span class='pct-badge' :class='{ "pct-done": isDone, "pct-error": isError }'>{{ p2pStore.activeTransfer.progress }}%</span>
         </div>

         <!-- Progress bar -->
         <div class='progress-track'>
             <div class='progress-fill' :class='progressClass' :style='{ width: p2pStore.activeTransfer.progress + "%" }'></div>
         </div>

         <!-- Reconnecting notice -->
         <div v-if='isReconnecting' class='state-notice notice-reconnect'>
             <svg class='spin-anim' xmlns='http://www.w3.org/2000/svg' width='14' height='14' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'><path d='M21 12a9 9 0 1 1-6.219-8.56'/></svg>
             <span>{{ t('p2p.statusReconnecting') }}</span>
             <span v-if='resumeAttempts > 0' class='attempt-badge'>{{ t('p2p.reconnectAttempt', { current: resumeAttempts, max: 3 }) }}</span>
         </div>

         <!-- Error notice -->
         <div v-if='isError' class='state-notice notice-error'>
             <svg xmlns='http://www.w3.org/2000/svg' width='14' height='14' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'><circle cx='12' cy='12' r='10'/><line x1='12' y1='8' x2='12' y2='12'/><line x1='12' y1='16' x2='12.01' y2='16'/></svg>
             <span>{{ t('p2p.transferError') }}</span>
         </div>

         <!-- Speed + ETA + elapsed -->
         <div v-if='transferSpeed !== null || (isDone && formattedElapsed)' class='speed-eta-row'>
             <span v-if='transferSpeed !== null' class='speed-value'>{{ formattedSpeed }}</span>
             <span v-if='transferETA !== null' class='eta-value'>{{ formattedETA }}</span>
             <span v-if='isDone && formattedElapsed' class='elapsed-value'>{{ t('p2p.elapsedTime') }}: {{ formattedElapsed }}</span>
         </div>

         <p class='filename-display' :title='p2pStore.activeTransfer.fileName'>{{ p2pStore.activeTransfer.fileName }}</p>

         <!-- Keep window active warning -->
         <div v-if='!isDone && !isError' class='keep-active-notice'>
           <svg xmlns='http://www.w3.org/2000/svg' width='14' height='14' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'><path d='M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z'></path><line x1='12' y1='9' x2='12' y2='13'></line><line x1='12' y1='17' x2='12.01' y2='17'></line></svg>
           {{ t('p2p.keepWindowActive') }}
         </div>

         <!-- Connection Info — hidden during reconnect (stage text is redundant there) -->
         <div v-if='p2pStore.activeTransfer.connectionInfo && !isReconnecting && !isError' class='connection-info'>
             <div class='info-row'>
                 <span class='info-label'>{{ t('common.state') }}:</span>
                 <span class='info-value'>{{ p2pStore.activeTransfer.connectionInfo.stage }}</span>
             </div>
             <div v-if='p2pStore.activeTransfer.connectionInfo.connectionType' class='info-row'>
                 <span class='info-label'>{{ t('common.type') }}:</span>
                 <span class='info-value' :class='{ "turn-relay": p2pStore.activeTransfer.connectionInfo.usingTurn }'>
                     {{ p2pStore.activeTransfer.connectionInfo.connectionType }}
                 </span>
             </div>
         </div>

         <div class='actions-grid single' v-if='isDone || isError'>
            <button @click='close' class='btn btn-primary'>{{ t('common.close') }}</button>
         </div>
         <div class='actions-grid single' v-else>
             <button @click='cancel' class='btn btn-danger-text'>{{ t('common.cancel') }}</button>
         </div>

         <!-- Re-notify while waiting for acceptance -->
         <div v-if='isWaitingForAcceptance' class='renotify-section'>
             <button
                 class='btn btn-renotify'
                 @click='sendPing'
                 :disabled='!canPing'
             >
                 <svg xmlns='http://www.w3.org/2000/svg' width='14' height='14' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'><path d='M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9'/><path d='M13.73 21a2 2 0 0 1-3.46 0'/></svg>
                 {{ pingBtnText }}
             </button>
             <span class='ping-count'>{{ t('p2p.pingsLeft', { count: pingsLeft }) }}</span>
         </div>
      </div>

      </template><!-- end v-if="!isMinimized" -->
    </div>
  </div>
</template>

<script setup>
import { computed, ref, watch, onMounted, onUnmounted, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';
import { useP2PStore } from '../stores/p2p';
import { useFriendStore } from '../stores/friends';

const { t } = useI18n();

const p2pStore = useP2PStore();
const friendStore = useFriendStore();

const senderName = computed(() => {
    const offer = p2pStore.incomingOffer;
    if (!offer) return '';
    const friend = friendStore.friends.find(f => f.id === offer.senderId);
    return friend?.name || offer.senderId.substring(0, 8);
});

const visible = computed(() => !!p2pStore.incomingOffer || !!p2pStore.activeTransfer || !!p2pStore.rejectedTransfer);
const isDone = computed(() => p2pStore.activeTransfer?.status === 'Done' || p2pStore.activeTransfer?.status === 'Complete');
const isConnecting = computed(() => p2pStore.activeTransfer?.status === 'Connecting...');
const isReconnecting = computed(() => p2pStore.activeTransfer?.status === 'Reconnecting...');
const isError = computed(() => p2pStore.activeTransfer?.status === 'Error');
const canClose = computed(() => isDone.value || isError.value || !!p2pStore.incomingOffer || !!p2pStore.rejectedTransfer);
const isWaitingForAcceptance = computed(() =>
    p2pStore.activeTransfer?.type === 'send' && p2pStore.activeTransfer?.status === 'Connecting...'
);

const STATUS_I18N = {
    'Connecting...':   'p2p.statusConnecting',
    'Sending...':      'p2p.statusSending',
    'Receiving...':    'p2p.statusReceiving',
    'Done':            'p2p.statusDone',
    'Complete':        'p2p.statusComplete',
    'Reconnecting...': 'p2p.statusReconnecting',
    'Error':           'p2p.statusError',
};
const statusText = computed(() => {
    const status = p2pStore.activeTransfer?.status;
    if (!status) return '';
    const key = STATUS_I18N[status];
    return key ? t(key) : status;
});

const statusLabelClass = computed(() => ({
    'status-reconnecting': isReconnecting.value,
    'status-error':        isError.value,
    'status-done':         isDone.value,
}));

const progressClass = computed(() => ({
    'progress-fill--reconnecting': isReconnecting.value,
    'progress-fill--error':        isError.value,
    'progress-fill--done':         isDone.value,
}));

const resumeAttempts = computed(() => p2pStore.activeTransfer?.resumeAttempts ?? 0);

const isMinimized = ref(false);

// --- Shake card ---
const shakeCard = ref(false);
function triggerShake() {
    shakeCard.value = false;
    nextTick(() => {
        shakeCard.value = true;
        setTimeout(() => { shakeCard.value = false; }, 700);
    });
}

// --- Sound ---
function playPingSound() {
    try {
        const ctx = new (window.AudioContext || window.webkitAudioContext)();
        const oscillator = ctx.createOscillator();
        const gainNode = ctx.createGain();
        oscillator.connect(gainNode);
        gainNode.connect(ctx.destination);
        oscillator.type = 'sine';
        oscillator.frequency.setValueAtTime(880, ctx.currentTime);
        oscillator.frequency.exponentialRampToValueAtTime(440, ctx.currentTime + 0.3);
        gainNode.gain.setValueAtTime(0.3, ctx.currentTime);
        gainNode.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.5);
        oscillator.start(ctx.currentTime);
        oscillator.stop(ctx.currentTime + 0.5);
    } catch (_) { /* Audio not supported */ }
}

// --- Browser notifications ---
function requestNotificationPermission() {
    if ('Notification' in window && Notification.permission === 'default') {
        Notification.requestPermission();
    }
}
function showBrowserNotification(title, body) {
    if ('Notification' in window && Notification.permission === 'granted') {
        new Notification(title, { body, icon: '/Logo.png' });
    }
}

// Trigger on incoming offer
watch(() => p2pStore.incomingOffer, (offer) => {
    if (offer) {
        isMinimized.value = false; // Always expand on new offer
        playPingSound();
        showBrowserNotification(t('p2p.title'), t('p2p.incomingRequest', { sender: senderName.value }));
        setTimeout(triggerShake, 300);
    }
});

// Trigger on p2p_ping signal from sender — pingSeq increments on every ping,
// so the watcher always fires regardless of any previous state.
watch(() => p2pStore.pingSeq, (seq) => {
    if (seq === 0) return; // initial mount, not an actual ping
    triggerShake();
    playPingSound();
    if (p2pStore.incomingOffer) {
        showBrowserNotification(t('p2p.pingNotification'), t('p2p.pingNotificationBody'));
    }
});

// --- Re-notify cooldown timer ---
const now = ref(Date.now());
let nowInterval = null;
onMounted(() => {
    nowInterval = setInterval(() => { now.value = Date.now(); }, 1000);
    requestNotificationPermission();
});
onUnmounted(() => {
    if (nowInterval) clearInterval(nowInterval);
});

const canPing = computed(() => {
    if (p2pStore.pingCount >= 3) return false;
    if (p2pStore.pingCooldownUntil && now.value < p2pStore.pingCooldownUntil) return false;
    return true;
});
const pingsLeft = computed(() => Math.max(0, 3 - p2pStore.pingCount));
const cooldownSecondsLeft = computed(() => {
    if (!p2pStore.pingCooldownUntil) return 0;
    return Math.max(0, Math.ceil((p2pStore.pingCooldownUntil - now.value) / 1000));
});
const pingBtnText = computed(() => {
    if (p2pStore.pingCount >= 3) return t('p2p.pingLimitReached');
    if (cooldownSecondsLeft.value > 0) return `${t('p2p.pingNotify')} (${cooldownSecondsLeft.value}s)`;
    return t('p2p.pingNotify');
});

function sendPing() {
    if (!p2pStore.activeTransfer) return;
    p2pStore.sendPing(p2pStore.activeTransfer.friendId, p2pStore.activeTransfer.transferId);
}

const formatSize = (bytes) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Number.parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

// Speed in bytes/sec, null if not yet meaningful
const transferSpeed = computed(() => {
    const t = p2pStore.activeTransfer;
    if (!t?.transferStartedAt) return null;
    const elapsed = (now.value - t.transferStartedAt) / 1000;
    if (elapsed < 1) return null;
    const transferred = t.transferredBytes ?? t.receivedSize ?? 0;
    return transferred / elapsed;
});

const transferETA = computed(() => {
    const speed = transferSpeed.value;
    if (!speed || speed < 512) return null;
    const t = p2pStore.activeTransfer;
    const total = t.totalBytes ?? t.fileSize ?? 0;
    const transferred = t.transferredBytes ?? t.receivedSize ?? 0;
    const remaining = total - transferred;
    if (remaining <= 0) return null;
    return remaining / speed; // seconds
});

const formattedSpeed = computed(() => {
    const s = transferSpeed.value;
    if (s === null) return '';
    if (s >= 1024 * 1024) return `${(s / (1024 * 1024)).toFixed(1)} MB/s`;
    if (s >= 1024) return `${(s / 1024).toFixed(0)} KB/s`;
    return `${Math.round(s)} B/s`;
});

const formattedETA = computed(() => {
    const secs = transferETA.value;
    if (secs === null) return '';
    if (secs < 5) return '< 5s';
    if (secs < 60) return `~${Math.round(secs)}s`;
    const m = Math.floor(secs / 60);
    const s = Math.round(secs % 60);
    return s > 0 ? `~${m}m ${s}s` : `~${m}m`;
});

// Elapsed time — shown once the transfer is complete
const formattedElapsed = computed(() => {
    const tr = p2pStore.activeTransfer;
    if (!tr?.transferStartedAt) return null;
    const secs = Math.floor((now.value - tr.transferStartedAt) / 1000);
    if (secs < 60) return `${secs}s`;
    const m = Math.floor(secs / 60);
    const s = secs % 60;
    return s > 0 ? `${m}m ${s}s` : `${m}m`;
});

const accept = () => p2pStore.acceptTransfer();
const reject = () => p2pStore.rejectTransfer();
const cancel = () => p2pStore.cancelTransfer();
const close = () => {
    if(p2pStore.incomingOffer) p2pStore.rejectTransfer();
    else if(p2pStore.rejectedTransfer) p2pStore.rejectedTransfer = null;
    else if(p2pStore.activeTransfer && isDone.value) p2pStore.activeTransfer = null;
    else p2pStore.cancelTransfer();
};
</script>

<style scoped>
.p2p-notification-container {
    position: fixed;
    bottom: 2rem;
    right: 2rem;
    z-index: 9999;
    /* Ensure no full height taking */
    height: auto;
    width: auto;
}

.p2p-card {
    background: var(--card-color, #ffffff);
    width: 340px;
    border-radius: 12px;
    box-shadow: 0 8px 30px rgba(0,0,0,0.12);
    border: 1px solid var(--border-color, #e0e0e0);
    overflow: hidden;
    font-family: inherit;
    animation: slideIn 0.3s ease-out;
}

@keyframes slideIn {
    from { transform: translateY(20px); opacity: 0; }
    to { transform: translateY(0); opacity: 1; }
}

@keyframes shake {
    0%, 100% { transform: translateX(0); }
    15%, 45%, 75% { transform: translateX(-6px); }
    30%, 60%, 90% { transform: translateX(6px); }
}

.p2p-card.shaking {
    animation: shake 0.65s cubic-bezier(.36,.07,.19,.97) both;
}

.card-header {
    background: var(--background-color, #f8f9fa);
    padding: 12px 16px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-bottom: 1px solid var(--border-color, #eaeaea);
}

.header-title {
    margin: 0;
    font-size: 0.95rem;
    font-weight: 600;
    color: var(--main-text-color, #333);
    display: flex;
    align-items: center;
    gap: 8px;
}
.header-title svg { color: var(--primary-color, #3498db); }

.close-icon {
    background: none;
    border: none;
    font-size: 1.4rem;
    line-height: 1;
    cursor: pointer;
    color: var(--secondary-text-color, #888);
    padding: 0;
}
.close-icon:hover { color: var(--main-text-color, #333); }

.notification-body {
    padding: 16px;
}

.request-text {
    margin: 0 0 12px 0;
    font-size: 0.9rem;
    color: var(--main-text-color, #444);
    line-height: 1.4;
}

.file-preview {
    display: flex;
    align-items: center;
    gap: 12px;
    background: var(--hover-background-color, #f4f6f8);
    padding: 10px;
    border-radius: 8px;
    margin-bottom: 16px;
}

.file-icon-box {
    color: var(--secondary-text-color, #666);
    display: flex;
}

.file-info {
    flex: 1;
    overflow: hidden;
}

.f-name {
    font-weight: 600;
    font-size: 0.9rem;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    color: var(--main-text-color, #333);
}

.f-size {
    font-size: 0.75rem;
    color: var(--secondary-text-color, #888);
}

.actions-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
}
.actions-grid.single {
    grid-template-columns: 1fr;
}

.btn {
    padding: 8px;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    font-size: 0.9rem;
    font-weight: 600;
    transition: all 0.2s;
}

.btn-primary {
    background: var(--primary-color, #3498db);
    color: white;
}
.btn-primary:hover {
    filter: brightness(1.1);
}

.btn-secondary {
    background: transparent;
    border: 1px solid var(--border-color, #ddd);
    color: var(--secondary-text-color, #666);
}
.btn-secondary:hover {
    background: var(--hover-background-color, #f5f5f5);
    color: var(--main-text-color, #333);
}

.btn-danger-text {
    background: none;
    color: var(--error-color, #e74c3c);
    text-decoration: underline;
}

/* Progress styles */
.progress-track {
    height: 6px;
    background: var(--border-color, #eee);
    border-radius: 3px;
    overflow: hidden;
    margin: 8px 0 12px 0;
}
.progress-fill {
    height: 100%;
    background: var(--success-color, #2ecc71);
    transition: width 0.3s ease;
}
.progress-fill--done {
    background: var(--primary-color, #3498db);
}
.progress-fill--reconnecting {
    background: var(--warning-color, #f39c12);
    animation: pulse-bar 1.4s ease-in-out infinite;
}
.progress-fill--error {
    background: var(--error-color, #e74c3c);
}
@keyframes pulse-bar {
    0%, 100% { opacity: 1; }
    50%       { opacity: 0.55; }
}

/* Spinner animation */
@keyframes spin {
    from { transform: rotate(0deg); }
    to   { transform: rotate(360deg); }
}
.spin-anim {
    animation: spin 0.9s linear infinite;
    flex-shrink: 0;
}

/* Status label variants */
.status-icon-row {
    display: flex;
    align-items: center;
    gap: 5px;
}
.status-reconnecting { color: var(--warning-color, #f39c12); }
.status-error        { color: var(--error-color, #e74c3c); }
.status-done         { color: var(--success-color, #2ecc71); }

.pct-done  { color: var(--success-color, #2ecc71); }
.pct-error { color: var(--error-color, #e74c3c); }

/* State notice banners */
.state-notice {
    display: flex;
    align-items: center;
    gap: 8px;
    border-radius: 6px;
    padding: 7px 10px;
    font-size: 0.8rem;
    font-weight: 500;
    margin-bottom: 10px;
    flex-wrap: wrap;
}
.notice-reconnect {
    background: var(--warning-bg-color, #fffbf0);
    border: 1px solid var(--warning-color, #f39c12);
    color: var(--warning-color, #f39c12);
}
.notice-error {
    background: var(--error-bg-color, #fdf0f0);
    border: 1px solid var(--error-color, #e74c3c);
    color: var(--error-color, #e74c3c);
}
.attempt-badge {
    margin-left: auto;
    font-size: 0.75rem;
    opacity: 0.85;
    white-space: nowrap;
}

.elapsed-value {
    color: var(--secondary-text-color, #888);
    font-size: 0.78rem;
    margin-left: auto;
}

.status-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 0.85rem;
    color: var(--secondary-text-color, #666);
    margin-bottom: 4px;
}
.pct-badge {
    font-weight: 700;
    color: var(--primary-color, #3498db);
}
.speed-eta-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 0.78rem;
    margin-bottom: 8px;
}

.speed-value {
    font-weight: 600;
    color: var(--primary-color, #3498db);
}

.eta-value {
    color: var(--secondary-text-color, #888);
}

.filename-display {
    font-size: 0.85rem;
    color: var(--main-text-color, #333);
    margin: 0 0 12px 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.connection-info {
    background: var(--hover-background-color, #f4f6f8);
    border-radius: 6px;
    padding: 8px 10px;
    margin-bottom: 12px;
    font-size: 0.8rem;
}

.info-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 3px 0;
}

.info-label {
    color: var(--secondary-text-color, #888);
    font-weight: 500;
}

.info-value {
    color: var(--main-text-color, #333);
    font-weight: 600;
}

.info-value.turn-relay {
    color: var(--warning-color, #f39c12);
}

.rejected-notice {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    background: var(--error-bg-color, #fdf0f0);
    border: 1px solid var(--error-color, #e74c3c);
    border-radius: 8px;
    padding: 12px;
    margin-bottom: 16px;
    color: var(--error-color, #e74c3c);
}
.rejected-notice svg { flex-shrink: 0; margin-top: 2px; }
.rejected-title {
    margin: 0 0 4px 0;
    font-weight: 600;
    font-size: 0.9rem;
}
.rejected-file {
    margin: 0;
    font-size: 0.8rem;
    color: var(--secondary-text-color, #888);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 240px;
}

.renotify-section {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-top: 10px;
    padding-top: 10px;
    border-top: 1px solid var(--border-color, #eee);
}

.btn-renotify {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 12px;
    background: transparent;
    border: 1px solid var(--primary-color, #3498db);
    color: var(--primary-color, #3498db);
    border-radius: 6px;
    cursor: pointer;
    font-size: 0.82rem;
    font-weight: 600;
    transition: all 0.2s;
    white-space: nowrap;
}
.btn-renotify:hover:not(:disabled) {
    background: var(--primary-color, #3498db);
    color: white;
}
.btn-renotify:disabled {
    opacity: 0.45;
    cursor: not-allowed;
}

.ping-count {
    font-size: 0.78rem;
    color: var(--secondary-text-color, #888);
    white-space: nowrap;
}

.keep-active-notice {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 0.78rem;
    color: var(--warning-color, #f39c12);
    background: var(--warning-bg-color, #fffbf0);
    border: 1px solid var(--warning-color, #f39c12);
    border-radius: 6px;
    padding: 6px 10px;
    margin-bottom: 10px;
}

.header-actions {
    display: flex;
    align-items: center;
    gap: 4px;
}

.minimize-icon {
    background: none;
    border: none;
    cursor: pointer;
    color: var(--secondary-text-color, #888);
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 2px 4px;
    border-radius: 4px;
}
.minimize-icon:hover { color: var(--main-text-color, #333); }

.mini-progress {
    font-size: 0.78rem;
    font-weight: 700;
    color: var(--primary-color, #3498db);
    margin-left: 8px;
}

@media (max-width: 768px) {
    .p2p-notification-container {
        bottom: 72px; /* above mobile nav */
        right: 8px;
        left: 8px;
    }

    .p2p-card {
        width: 100%;
    }

    /* Minimized: just a compact pill at bottom-right */
    .p2p-notification-container.minimized {
        left: auto;
        right: 8px;
        width: auto;
    }

    .p2p-notification-container.minimized .p2p-card {
        width: auto;
        min-width: 160px;
    }

    .p2p-notification-container.minimized .card-header {
        cursor: pointer;
    }
}
</style>
