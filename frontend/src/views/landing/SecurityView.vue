<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="security-page">
    <LandingNav />

    <!-- Hero -->
    <section class="sec-hero">
      <div class="container">
        <div class="hero-eyebrow">{{ t('landing.security.eyebrow') }}</div>
        <h1>
          {{ t('landing.security.title1') }}<br>
          <span class="highlight">{{ t('landing.security.titleHighlight') }}</span>
        </h1>
        <p class="hero-sub">{{ t('landing.security.heroSub') }}</p>
        <div class="hero-pill-row">
          <span class="pill green">AES-256-GCM</span>
          <span class="pill blue">Argon2id</span>
          <span class="pill purple">RSA-OAEP 4096</span>
          <span class="pill orange">WebCrypto API</span>
          <span class="pill teal">Service Worker</span>
        </div>
      </div>
    </section>

    <!-- Sommaire flottant -->
    <aside class="toc-sidebar">
      <div class="toc-panel" :class="{ folded: !tocNavOpen }">
        <div class="toc-title" @click="tocNavOpen = !tocNavOpen">
          {{ t('landing.security.toc') }}
          <svg class="toc-chevron" :class="{ open: tocNavOpen }" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="6 9 12 15 18 9"/></svg>
        </div>
        <div class="toc-nav-wrapper" :class="{ hidden: !tocNavOpen }">
          <nav>
            <a
              v-for="item in tocItems"
              :key="item.id"
              class="toc-item"
              :class="{ active: activeSection === item.id }"
              @click="scrollTo(item.id)"
            >
              <span class="toc-num">{{ item.num }}</span>
              <span class="toc-label">{{ item.label }}</span>
            </a>
          </nav>
        </div>
      </div>
    </aside>

    <!-- Principe ZK -->
    <section id="principe" class="section">
      <div class="container">
        <div class="section-label">{{ t('landing.security.sec1Label') }}</div>
        <h2>{{ t('landing.security.sec1Title') }}</h2>
        <p class="section-sub">{{ t('landing.security.sec1Sub') }}</p>

        <div class="zk-comparison">
          <div class="cmp-card bad">
            <div class="cmp-header">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>
              {{ t('landing.security.cmpBadCloud') }}
            </div>
            <div class="cmp-flow">
              <div class="cmp-node user-node">{{ t('landing.security.cmpYourFile') }}</div>
              <div class="cmp-arrow bad-arrow">
                <span class="arrow-label">{{ t('landing.security.cmpFilePlain') }}</span>
                <svg viewBox="0 0 60 20"><path d="M0 10 L50 10" stroke="#ef4444" stroke-width="2"/><polygon points="50,5 60,10 50,15" fill="#ef4444"/></svg>
              </div>
              <div class="cmp-node server-bad">{{ t('landing.security.cmpServerBad') }}<br><small>{{ t('landing.security.cmpServerBadSub') }}</small></div>
              <div class="cmp-arrow bad-arrow">
                <svg viewBox="0 0 60 20"><path d="M0 10 L50 10" stroke="#ef4444" stroke-width="2"/><polygon points="50,5 60,10 50,15" fill="#ef4444"/></svg>
              </div>
              <div class="cmp-node">{{ t('landing.security.cmpStorage') }}</div>
            </div>
            <p class="cmp-note">{{ t('landing.security.cmpBadNote') }}</p>
          </div>

          <div class="cmp-card good">
            <div class="cmp-header">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 11.08V12a10 10 0 11-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
              {{ t('landing.security.cmpGoodCloud') }}
            </div>
            <div class="cmp-flow">
              <div class="cmp-node user-node">{{ t('landing.security.cmpYourFile') }}</div>
              <div class="cmp-arrow good-arrow">
                <span class="arrow-label arrow-label-top">{{ t('landing.security.cmpLocalEnc') }}</span>
                <svg viewBox="0 0 60 20"><path d="M0 10 L50 10" stroke="#22c55e" stroke-width="2"/><polygon points="50,5 60,10 50,15" fill="#22c55e"/></svg>
              </div>
              <div class="cmp-node encrypted-node">{{ t('landing.security.cmpEncrypted') }}<br><small>(AES-256)</small></div>
              <div class="cmp-arrow good-arrow">
                <span class="arrow-label">{{ t('landing.security.cmpOpaqueBytes') }}</span>
                <svg viewBox="0 0 60 20"><path d="M0 10 L50 10" stroke="#22c55e" stroke-width="2"/><polygon points="50,5 60,10 50,15" fill="#22c55e"/></svg>
              </div>
              <div class="cmp-node server-good">{{ t('landing.security.cmpServerGood') }}<br><small>{{ t('landing.security.cmpServerGoodSub') }}</small></div>
            </div>
            <p class="cmp-note">{{ t('landing.security.cmpGoodNote') }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- Hiérarchie des clés -->
    <section id="hierarchie" class="section dark-section">
      <div class="container">
        <div class="section-label">{{ t('landing.security.sec2Label') }}</div>
        <h2>{{ t('landing.security.sec2Title') }}</h2>
        <p class="section-sub">{{ t('landing.security.sec2Sub') }}</p>

        <div class="key-hierarchy">
          <!-- Niveau 0: Mot de passe -->
          <div class="kh-level">
            <div class="kh-node kh-password" :class="{ active: activeKey === 'password' }" @mouseenter="activeKey = 'password'" @mouseleave="activeKey = null">
              <div class="kh-icon">🔑</div>
              <div class="kh-label">{{ t('landing.security.khPasswordLabel') }}</div>
              <div class="kh-sub">{{ t('landing.security.khPasswordSub') }}</div>
              <div class="kh-tooltip">
                <strong>{{ t('landing.security.khPasswordLabel') }}</strong> — {{ t('landing.security.khPasswordTooltipBody') }}
              </div>
            </div>
            <div class="kh-connector">
              <div class="kh-algo-badge">Argon2id · 64 MB · 4 passes</div>
              <div class="kh-line animated-line"></div>
            </div>
          </div>

          <!-- Niveau 1: KEK -->
          <div class="kh-level">
            <div class="kh-node kh-kek" :class="{ active: activeKey === 'kek' }" @mouseenter="activeKey = 'kek'" @mouseleave="activeKey = null">
              <div class="kh-icon">🗝️</div>
              <div class="kh-label">KEK</div>
              <div class="kh-sub">{{ t('landing.security.khKekSub') }}</div>
              <div class="kh-tooltip">
                <strong>KEK (Key Encryption Key)</strong> — {{ t('landing.security.khKekTooltipBody') }}
              </div>
            </div>
            <div class="kh-connector">
              <div class="kh-algo-badge">AES-256-GCM (unwrap)</div>
              <div class="kh-line animated-line"></div>
            </div>
          </div>

          <!-- Niveau 2: Master Key -->
          <div class="kh-level">
            <div class="kh-node kh-master" :class="{ active: activeKey === 'master' }" @mouseenter="activeKey = 'master'" @mouseleave="activeKey = null">
              <div class="kh-icon">🏛️</div>
              <div class="kh-label">Master Key</div>
              <div class="kh-sub">{{ t('landing.security.khMasterSub') }}</div>
              <div class="kh-tooltip">
                <strong>Master Key</strong> — {{ t('landing.security.khMasterTooltipBody1') }}<code>extractable: false</code>{{ t('landing.security.khMasterTooltipBody2') }}
              </div>
            </div>
            <div class="kh-branches">
              <div class="kh-branch-line"></div>
              <div class="kh-branch-row">
                <div class="kh-branch-col">
                  <div class="kh-algo-badge small">AES-GCM wrap</div>
                  <div class="kh-line-short animated-line"></div>
                  <div class="kh-node kh-rsa" :class="{ active: activeKey === 'rsa' }" @mouseenter="activeKey = 'rsa'" @mouseleave="activeKey = null">
                    <div class="kh-icon small">🔐</div>
                    <div class="kh-label small">{{ t('landing.security.khRsaLabel') }}</div>
                    <div class="kh-sub">{{ t('landing.security.khRsaSub1') }}<br>{{ t('landing.security.khRsaSub2') }}</div>
                    <div class="kh-tooltip kh-tooltip-left">
                      <strong>{{ t('landing.security.khRsaTooltipTitle') }}</strong> — {{ t('landing.security.khRsaTooltipBody') }}
                    </div>
                  </div>
                </div>
                <div class="kh-branch-col">
                  <div class="kh-algo-badge small">AES-GCM wrap</div>
                  <div class="kh-line-short animated-line"></div>
                  <div class="kh-node kh-folder" :class="{ active: activeKey === 'folder' }" @mouseenter="activeKey = 'folder'" @mouseleave="activeKey = null">
                    <div class="kh-icon small">📁</div>
                    <div class="kh-label small">{{ t('landing.security.khFolderLabel') }}</div>
                    <div class="kh-sub">{{ t('landing.security.khFolderSub1') }}<br>{{ t('landing.security.khFolderSub2') }}</div>
                    <div class="kh-tooltip">
                      <strong>{{ t('landing.security.khFolderTooltipTitle') }}</strong> — {{ t('landing.security.khFolderTooltipBody') }}
                    </div>
                  </div>
                  <div class="kh-line-short animated-line"></div>
                  <div class="kh-node kh-file" :class="{ active: activeKey === 'file' }" @mouseenter="activeKey = 'file'" @mouseleave="activeKey = null">
                    <div class="kh-icon small">📄</div>
                    <div class="kh-label small">{{ t('landing.security.khFileLabel') }}</div>
                    <div class="kh-sub">{{ t('landing.security.khFileSub1') }}<br>{{ t('landing.security.khFileSub2') }}</div>
                    <div class="kh-tooltip">
                      <strong>{{ t('landing.security.khFileTooltipTitle') }}</strong> — {{ t('landing.security.khFileTooltipBody') }}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Inscription animée -->
    <section id="inscription" class="section">
      <div class="container">
        <div class="section-label">{{ t('landing.security.sec3Label') }}</div>
        <h2>{{ t('landing.security.sec3Title') }}</h2>
        <p class="section-sub">{{ t('landing.security.sec3Sub') }}</p>

        <div class="flow-container">
          <div class="flow-steps">
            <div
              v-for="(step, i) in registrationSteps"
              :key="i"
              class="flow-step"
              :class="{ active: regStep === i, done: regStep > i }"
              @click="regStep = i"
            >
              <div class="step-num">{{ i + 1 }}</div>
              <div class="step-content">
                <div class="step-title">{{ step.title }}</div>
                <div class="step-desc">{{ step.desc }}</div>
              </div>
            </div>
          </div>

          <div class="flow-visual">
            <!-- Step 0: password → argon2 -->
            <div v-if="regStep === 0" class="anim-scene">
              <div class="anim-node input-node">
                <div class="node-icon">⌨️</div>
                <div class="node-label">{{ t('landing.security.rfInputLabel') }}</div>
                <div class="node-sub">{{ t('landing.security.rfInputSub') }}</div>
              </div>
              <div class="anim-pipe">
                <div class="pipe-label">Web Worker</div>
                <div class="anim-packet pw-packet" :class="{ flowing: regStep === 0 }">pw</div>
                <div class="pipe-algo">Argon2id<br>64 MB · 4 passes<br>{{ t('landing.security.rfArgonSalt') }}</div>
              </div>
              <div class="anim-node kek-node">
                <div class="node-icon">🗝️</div>
                <div class="node-label">KEK</div>
                <div class="node-sub">{{ t('landing.security.rfKekSub') }}</div>
              </div>
            </div>

            <!-- Step 1: generate master key -->
            <div v-if="regStep === 1" class="anim-scene">
              <div class="anim-node rng-node">
                <div class="node-icon">🎲</div>
                <div class="node-label">CSPRNG</div>
                <div class="node-sub">crypto.getRandomValues()</div>
              </div>
              <div class="anim-pipe">
                <div class="anim-packet mk-packet" :class="{ flowing: regStep === 1 }">256b</div>
                <div class="pipe-algo">{{ t('landing.security.rfMkGenLine1') }}<br>non-extractable</div>
              </div>
              <div class="anim-node master-node">
                <div class="node-icon">🏛️</div>
                <div class="node-label">Master Key</div>
                <div class="node-sub">{{ t('landing.security.rfMkSub') }}</div>
              </div>
            </div>

            <!-- Step 2: wrap master key -->
            <div v-if="regStep === 2" class="anim-scene wrap-scene">
              <div class="wrap-inputs">
                <div class="anim-node small-node kek-node">
                  <div class="node-icon">🗝️</div>
                  <div class="node-label">KEK</div>
                </div>
                <div class="anim-node small-node master-node">
                  <div class="node-icon">🏛️</div>
                  <div class="node-label">Master Key</div>
                </div>
              </div>
              <div class="wrap-arrow">
                <div class="anim-packet enc-packet" :class="{ flowing: regStep === 2 }">AES-GCM<br>wrap</div>
                <svg viewBox="0 0 80 40" class="wrap-svg"><path d="M0 20 Q40 5 80 20" stroke="var(--primary-color)" stroke-width="2" fill="none" stroke-dasharray="4 2"/></svg>
              </div>
              <div class="anim-node encrypted-master-node">
                <div class="node-icon">🔒</div>
                <div class="node-label">encrypted_master_key</div>
                <div class="node-sub">{{ t('landing.security.rfEncMkSub') }}</div>
              </div>
            </div>

            <!-- Step 3: RSA pair -->
            <div v-if="regStep === 3" class="anim-scene">
              <div class="anim-node rng-node">
                <div class="node-icon">🎲</div>
                <div class="node-label">CSPRNG</div>
              </div>
              <div class="anim-pipe">
                <div class="anim-packet rsa-packet" :class="{ flowing: regStep === 3 }">4096b</div>
                <div class="pipe-algo">RSA-OAEP<br>SHA-256</div>
              </div>
              <div class="rsa-split">
                <div class="anim-node pub-node">
                  <div class="node-icon">🔓</div>
                  <div class="node-label">{{ t('landing.security.rfPubKeyLabel') }}</div>
                  <div class="node-sub">{{ t('landing.security.rfPubKeySub') }}</div>
                </div>
                <div class="anim-node priv-node">
                  <div class="node-icon">🔐</div>
                  <div class="node-label">{{ t('landing.security.rfPrivKeyLabel') }}</div>
                  <div class="node-sub">{{ t('landing.security.rfPrivKeySub1') }}<br>{{ t('landing.security.rfPrivKeySub2') }}</div>
                </div>
              </div>
            </div>

            <!-- Step 4: recovery -->
            <div v-if="regStep === 4" class="anim-scene">
              <div class="anim-node rng-node">
                <div class="node-icon">🎲</div>
                <div class="node-label">{{ t('landing.security.rfRecoveryLabel') }}</div>
                <div class="node-sub">{{ t('landing.security.rfRecoverySub') }}</div>
              </div>
              <div class="recovery-split">
                <div class="rec-branch">
                  <div class="branch-label">SHA-256 (hash)</div>
                  <div class="anim-node hash-node">
                    <div class="node-icon">🔏</div>
                    <div class="node-label">recovery_hash</div>
                    <div class="node-sub">{{ t('landing.security.rfRecoveryHashSub') }}</div>
                  </div>
                </div>
                <div class="rec-branch">
                  <div class="branch-label">Argon2id → KEK recovery → AES-GCM wrap</div>
                  <div class="anim-node enc-rec-node">
                    <div class="node-icon">🔒</div>
                    <div class="node-label">encrypted_master_key_recovery</div>
                    <div class="node-sub">{{ t('landing.security.rfRecoveryEncSub') }}</div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Step 5: what goes to server -->
            <div v-if="regStep === 5" class="anim-scene server-scene">
              <div class="server-sends">
                <div class="send-item safe">
                  <span class="send-icon">✅</span>
                  <span>email</span>
                </div>
                <div class="send-item safe">
                  <span class="send-icon">✅</span>
                  <span>{{ t('landing.security.rfSaltLabel') }}</span>
                </div>
                <div class="send-item safe">
                  <span class="send-icon">✅</span>
                  <span>encrypted_master_key</span>
                </div>
                <div class="send-item safe">
                  <span class="send-icon">✅</span>
                  <span>public_key RSA (public)</span>
                </div>
                <div class="send-item safe">
                  <span class="send-icon">✅</span>
                  <span>encrypted_private_key</span>
                </div>
                <div class="send-item safe">
                  <span class="send-icon">✅</span>
                  <span>recovery_hash</span>
                </div>
                <div class="send-item never">
                  <span class="send-icon">🚫</span>
                  <span>{{ t('landing.security.rfNeverPassword') }}</span>
                </div>
                <div class="send-item never">
                  <span class="send-icon">🚫</span>
                  <span>{{ t('landing.security.rfNeverMk') }}</span>
                </div>
                <div class="send-item never">
                  <span class="send-icon">🚫</span>
                  <span>{{ t('landing.security.rfNeverKek') }}</span>
                </div>
              </div>
              <div class="server-box">
                <div class="server-icon">🖥️</div>
                <div class="server-label">{{ t('landing.security.rfServerLabel') }}</div>
                <div class="server-note">{{ t('landing.security.rfServerNote1') }}<br>{{ t('landing.security.rfServerNote2') }}</div>
              </div>
            </div>
          </div>
        </div>

        <div class="step-nav">
          <button class="step-btn" @click="regStep = Math.max(0, regStep - 1)" :disabled="regStep === 0">{{ t('landing.security.regPrev') }}</button>
          <span class="step-counter">{{ regStep + 1 }} / {{ registrationSteps.length }}</span>
          <button class="step-btn" @click="regStep = Math.min(registrationSteps.length - 1, regStep + 1)" :disabled="regStep === registrationSteps.length - 1">{{ t('landing.security.regNext') }}</button>
        </div>
      </div>
    </section>

    <!-- Upload fichier animé -->
    <section id="upload" class="section dark-section">
      <div class="container">
        <div class="section-label">{{ t('landing.security.sec4Label') }}</div>
        <h2>{{ t('landing.security.sec4Title') }}</h2>
        <p class="section-sub">{{ t('landing.security.sec4Sub') }}</p>

        <div class="upload-demo">
          <div class="upload-file-bar">
            <div class="file-icon">📄</div>
            <div class="file-info">
              <div class="file-name">document.pdf <span class="file-size">35 Mo</span></div>
              <div class="chunk-bar">
                <div
                  v-for="(chunk, i) in chunks"
                  :key="i"
                  class="chunk"
                  :class="{ encrypting: uploadStep === i && uploadPhase === 'encrypting', encrypted: uploadStep === i && uploadPhase === 'uploading', uploaded: uploadedChunks.includes(i) }"
                >
                  <span class="chunk-num">{{ i + 1 }}</span>
                </div>
              </div>
              <div class="chunk-legend">
                <span class="legend-dot plain"></span>{{ t('landing.security.uploadWaiting') }}
                <span class="legend-dot encrypting"></span>{{ t('landing.security.uploadEncrypting') }}
                <span class="legend-dot encrypted"></span>{{ t('landing.security.uploadEncrypted') }}
                <span class="legend-dot uploaded"></span>{{ t('landing.security.uploadUploaded') }}
              </div>
            </div>
          </div>

          <div class="upload-pipeline">
            <div class="pipeline-stage" :class="{ active: uploadPhase === 'reading' || uploadPhase === 'encrypting' }">
              <div class="stage-icon">💾</div>
              <div class="stage-label">{{ t('landing.security.uploadLocalRead') }}</div>
              <div class="stage-sub">FileReader API</div>
            </div>
            <div class="pipeline-arrow" :class="{ active: uploadPhase === 'encrypting' }">→</div>
            <div class="pipeline-stage" :class="{ active: uploadPhase === 'encrypting' }">
              <div class="stage-icon">⚙️</div>
              <div class="stage-label">{{ t('landing.security.uploadWorker') }}</div>
              <div class="stage-sub">crypto.worker.js</div>
              <div class="nonce-box" v-if="uploadPhase === 'encrypting'">
                <span class="nonce-label">Nonce chunk {{ uploadStep + 1 }}</span>
                <span class="nonce-val">{{ currentNonce }}</span>
              </div>
            </div>
            <div class="pipeline-arrow" :class="{ active: uploadPhase === 'upload' }">→</div>
            <div class="pipeline-stage" :class="{ active: uploadPhase === 'upload' }">
              <div class="stage-icon">☁️</div>
              <div class="stage-label">OVH S3</div>
              <div class="stage-sub">{{ t('landing.security.uploadOvh') }}</div>
            </div>
          </div>

          <div class="chunk-detail" v-if="uploadPhase === 'encrypting'">
            <div class="chunk-detail-title">{{ t('landing.security.uploadChunkDetail', { n: uploadStep + 1 }) }}</div>
            <div class="chunk-bytes">
              <div class="byte-block nonce-block">Nonce<br>{{ t('landing.security.uploadNonceSize') }}</div>
              <div class="byte-block cipher-block">Ciphertext AES-256-GCM<br>{{ t('landing.security.uploadCipherSize') }}</div>
              <div class="byte-block tag-block">Auth Tag<br>{{ t('landing.security.uploadTagSize') }}</div>
            </div>
          </div>

          <div class="upload-controls">
            <button class="step-btn primary" @click="advanceUpload" :disabled="uploadPhase === 'done'">
              {{ uploadBtnLabel }}
            </button>
            <button class="step-btn" @click="resetUpload" :disabled="uploadPhase === 'idle'">{{ t('landing.security.uploadReset') }}</button>
          </div>
          <div class="upload-progress-hint" v-if="uploadPhase !== 'idle' && uploadPhase !== 'done'">
            {{ t('landing.security.uploadChunkProgress', { n: uploadStep + 1, total: chunks.length }) }}
            <span v-if="uploadPhase === 'encrypting'">{{ t('landing.security.uploadEncProgress') }}</span>
            <span v-else>{{ t('landing.security.uploadS3Progress') }}</span>
          </div>
          <div class="upload-done-msg" v-if="uploadPhase === 'done'">
            {{ t('landing.security.uploadDone', { n: chunks.length }) }}
          </div>
        </div>
      </div>
    </section>

    <!-- Partage RSA -->
    <section id="partage" class="section">
      <div class="container">
        <div class="section-label">{{ t('landing.security.sec5Label') }}</div>
        <h2>{{ t('landing.security.sec5Title') }}</h2>
        <p class="section-sub">{{ t('landing.security.sec5Sub') }}</p>

        <div class="share-flow">
          <div class="share-actor alice">
            <div class="actor-avatar">👩</div>
            <div class="actor-name">Alice</div>
            <div class="actor-device">{{ t('landing.security.actorBrowser') }}</div>
          </div>

          <div class="share-steps-col">
            <div
              v-for="(step, i) in shareSteps"
              :key="i"
              class="share-step"
              :class="{ active: shareStep === i, done: shareStep > i }"
            >
              <div class="share-step-num">{{ i + 1 }}</div>
              <div class="share-step-body">
                <div class="share-step-actor" :class="step.actor">{{ step.actor === 'alice' ? t('landing.security.actorAlice') : step.actor === 'server' ? t('landing.security.actorServer') : t('landing.security.actorBob') }}</div>
                <div class="share-step-action">{{ step.action }}</div>
                <div class="share-step-data">{{ step.data }}</div>
              </div>
            </div>

            <div class="share-nav">
              <button class="step-btn" @click="shareStep = Math.max(0, shareStep - 1)" :disabled="shareStep === 0">{{ t('landing.security.sharePrev') }}</button>
              <span class="step-counter">{{ shareStep + 1 }} / {{ shareSteps.length }}</span>
              <button class="step-btn" @click="shareStep = Math.min(shareSteps.length - 1, shareStep + 1)" :disabled="shareStep === shareSteps.length - 1">{{ t('landing.security.shareNext') }}</button>
            </div>
          </div>

          <div class="share-actor bob">
            <div class="actor-avatar">👨</div>
            <div class="actor-name">Bob</div>
            <div class="actor-device">{{ t('landing.security.actorBrowser') }}</div>
          </div>
        </div>

        <div class="share-visual-box">
          <div class="share-visual" v-if="shareStep === 0">
            <div class="sv-node">
              <div class="sv-icon">📄</div>
              <div class="sv-label">{{ t('landing.security.svFileCrypted') }}</div>
            </div>
            <div class="sv-plus">+</div>
            <div class="sv-node highlight-node">
              <div class="sv-icon">🔑</div>
              <div class="sv-label">{{ t('landing.security.svAliceKey') }}</div>
              <div class="sv-sub">{{ t('landing.security.svAliceKeyMem') }}</div>
            </div>
          </div>
          <div class="share-visual" v-else-if="shareStep === 1">
            <div class="sv-node">
              <div class="sv-icon">🖥️</div>
              <div class="sv-label">{{ t('landing.security.svServer') }}</div>
            </div>
            <div class="sv-arrow">→</div>
            <div class="sv-node highlight-node">
              <div class="sv-icon">🔓</div>
              <div class="sv-label">{{ t('landing.security.svBobPubKey') }}</div>
              <div class="sv-sub">RSA-OAEP 4096 bits</div>
            </div>
          </div>
          <div class="share-visual" v-else-if="shareStep === 2">
            <div class="sv-node">
              <div class="sv-icon">🔑</div>
              <div class="sv-label">{{ t('landing.security.svFileKey') }}</div>
            </div>
            <div class="sv-pipe">
              <div class="sv-algo">RSA-OAEP encrypt<br>avec clé publique Bob</div>
              <div class="sv-packet rsa-anim"></div>
            </div>
            <div class="sv-node highlight-node">
              <div class="sv-icon">🔒</div>
              <div class="sv-label">{{ t('landing.security.svEncForBob') }}</div>
              <div class="sv-sub">{{ t('landing.security.svEncForBobSub') }}</div>
            </div>
          </div>
          <div class="share-visual" v-else-if="shareStep === 3">
            <div class="sv-node">
              <div class="sv-icon">🖥️</div>
              <div class="sv-label">{{ t('landing.security.svServerStores') }}</div>
              <div class="sv-sub">{{ t('landing.security.svServerStoresSub') }}</div>
            </div>
            <div class="sv-note">{{ t('landing.security.svServerNote') }}</div>
          </div>
          <div class="share-visual" v-else-if="shareStep === 4">
            <div class="sv-node">
              <div class="sv-icon">🔒</div>
              <div class="sv-label">{{ t('landing.security.svEncRsa') }}</div>
            </div>
            <div class="sv-pipe">
              <div class="sv-algo">RSA-OAEP decrypt<br>clé privée de Bob</div>
              <div class="sv-packet rsa-anim blue-anim"></div>
            </div>
            <div class="sv-node highlight-node success-node">
              <div class="sv-icon">🔑</div>
              <div class="sv-label">{{ t('landing.security.svBobDecrypts') }}</div>
              <div class="sv-sub">{{ t('landing.security.svBobDecryptsSub') }}</div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Service Worker -->
    <section id="session" class="section dark-section">
      <div class="container">
        <div class="section-label">{{ t('landing.security.sec6Label') }}</div>
        <h2>{{ t('landing.security.sec6Title') }}</h2>
        <p class="section-sub">{{ t('landing.security.sec6Sub') }}</p>

        <div class="sw-diagram">
          <div class="sw-zone page-zone">
            <div class="sw-zone-label">{{ t('landing.security.swPageZone') }}</div>
            <div class="sw-items">
              <div class="sw-item">{{ t('landing.security.swAppCode') }}</div>
              <div class="sw-item bad-item">{{ t('landing.security.swExtensions') }}</div>
              <div class="sw-item bad-item">{{ t('landing.security.swXss') }}</div>
              <div class="sw-item">
                <div class="sw-request">{{ t('landing.security.swRequest') }}</div>
              </div>
            </div>
          </div>

          <div class="sw-barrier">
            <div class="barrier-line"></div>
            <div class="barrier-label">{{ t('landing.security.swBarrier') }}</div>
            <div class="barrier-line"></div>
          </div>

          <div class="sw-zone worker-zone">
            <div class="sw-zone-label">{{ t('landing.security.swWorkerZone') }}</div>
            <div class="sw-items">
              <div class="sw-item key-item">
                🏛️ Master Key
                <div class="sw-badge">extractable: false</div>
              </div>
              <div class="sw-item">
                <div class="sw-response">{{ t('landing.security.swWorkerResponse') }}</div>
              </div>
              <div class="sw-item timeout-item">
                {{ t('landing.security.swTimeout') }}
              </div>
            </div>
          </div>
        </div>

        <div class="sw-props">
          <div class="sw-prop">
            <div class="prop-icon">🔒</div>
            <div class="prop-title">{{ t('landing.security.swProp1Title') }}</div>
            <div class="prop-desc">{{ t('landing.security.swProp1Desc') }}</div>
          </div>
          <div class="sw-prop">
            <div class="prop-icon">⏱️</div>
            <div class="prop-title">{{ t('landing.security.swProp2Title') }}</div>
            <div class="prop-desc">{{ t('landing.security.swProp2Desc') }}</div>
          </div>
          <div class="sw-prop">
            <div class="prop-icon">👁️</div>
            <div class="prop-title">{{ t('landing.security.swProp3Title') }}</div>
            <div class="prop-desc">{{ t('landing.security.swProp3Desc') }}</div>
          </div>
        </div>
      </div>
    </section>

    <!-- Ce que le serveur voit/ne voit pas -->
    <section id="serveur" class="section">
      <div class="container">
        <div class="section-label">{{ t('landing.security.sec7Label') }}</div>
        <h2>{{ t('landing.security.sec7Title') }}</h2>

        <div class="visibility-grid">
          <div class="vis-col vis-sees">
            <div class="vis-header sees-header">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>
              {{ t('landing.security.seesTitle') }}
            </div>
            <div class="vis-items">
              <div class="vis-item neutral">{{ t('landing.security.sees1') }}</div>
              <div class="vis-item neutral">{{ t('landing.security.sees2') }}</div>
              <div class="vis-item neutral">{{ t('landing.security.sees3') }}</div>
              <div class="vis-item neutral">{{ t('landing.security.sees4') }}</div>
              <div class="vis-item neutral">{{ t('landing.security.sees5') }}</div>
              <div class="vis-item neutral">{{ t('landing.security.sees6') }}</div>
              <div class="vis-item neutral">{{ t('landing.security.sees7') }}</div>
              <div class="vis-item neutral">{{ t('landing.security.sees8') }}</div>
              <div class="vis-item neutral">{{ t('landing.security.sees9') }}</div>
            </div>
          </div>

          <div class="vis-col vis-not">
            <div class="vis-header not-header">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M17.94 17.94A10.07 10.07 0 0112 20c-7 0-11-8-11-8a18.45 18.45 0 015.06-5.94M9.9 4.24A9.12 9.12 0 0112 4c7 0 11 8 11 8a18.5 18.5 0 01-2.16 3.19m-6.72-1.07a3 3 0 11-4.24-4.24"/><line x1="1" y1="1" x2="23" y2="23"/></svg>
              {{ t('landing.security.neverTitle') }}
            </div>
            <div class="vis-items">
              <div class="vis-item never-item">{{ t('landing.security.never1') }}</div>
              <div class="vis-item never-item">{{ t('landing.security.never2') }}</div>
              <div class="vis-item never-item">{{ t('landing.security.never3') }}</div>
              <div class="vis-item never-item">{{ t('landing.security.never4') }}</div>
              <div class="vis-item never-item">{{ t('landing.security.never5') }}</div>
              <div class="vis-item never-item">{{ t('landing.security.never6') }}</div>
              <div class="vis-item never-item">{{ t('landing.security.never7') }}</div>
              <div class="vis-item never-item">{{ t('landing.security.never8') }}</div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Footer CTA -->
    <section class="sec-cta">
      <div class="container">
        <h2>{{ t('landing.security.ctaTitle') }}</h2>
        <p>{{ t('landing.security.ctaSubtitle') }}</p>
        <div class="cta-row">
          <a href="https://github.com/Bunnntyyy/SaferCloud" target="_blank" rel="noopener" class="btn btn-primary">
            <svg viewBox="0 0 24 24" fill="currentColor" width="18"><path d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12"/></svg>
            {{ t('landing.security.ctaViewCode') }}
          </a>
          <router-link to="/dashboard" class="btn btn-secondary">{{ t('landing.security.ctaCreate') }}</router-link>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import LandingNav from '../../components/landing/LandingNav.vue'

const { t } = useI18n()

// ── TOC ────────────────────────────────────────────────────────────
const tocNavOpen = ref(true)
const activeSection = ref('')

const tocItems = computed(() => [
  { id: 'principe',   num: '01', label: t('landing.security.toc01') },
  { id: 'hierarchie', num: '02', label: t('landing.security.toc02') },
  { id: 'inscription',num: '03', label: t('landing.security.toc03') },
  { id: 'upload',     num: '04', label: t('landing.security.toc04') },
  { id: 'partage',    num: '05', label: t('landing.security.toc05') },
  { id: 'session',    num: '06', label: t('landing.security.toc06') },
  { id: 'serveur',    num: '07', label: t('landing.security.toc07') },
])

function scrollTo(id) {
  const el = document.getElementById(id)
  if (!el) return
  const page = document.querySelector('.security-page')
  if (page) {
    page.scrollTo({ top: el.offsetTop - 80, behavior: 'smooth' })
  }
}

let observer = null

onMounted(() => {
  observer = new IntersectionObserver(
    entries => {
      for (const entry of entries) {
        if (entry.isIntersecting) {
          activeSection.value = entry.target.id
        }
      }
    },
    { root: document.querySelector('.security-page'), rootMargin: '-30% 0px -60% 0px', threshold: 0 }
  )
  tocItems.value.forEach(item => {
    const el = document.getElementById(item.id)
    if (el) observer.observe(el)
  })
})

onUnmounted(() => {
  if (observer) observer.disconnect()
})

const activeKey = ref(null)
const regStep = ref(0)
const shareStep = ref(0)

const registrationSteps = computed(() => [
  { title: t('landing.security.regStep1Title'), desc: t('landing.security.regStep1Desc') },
  { title: t('landing.security.regStep2Title'), desc: t('landing.security.regStep2Desc') },
  { title: t('landing.security.regStep3Title'), desc: t('landing.security.regStep3Desc') },
  { title: t('landing.security.regStep4Title'), desc: t('landing.security.regStep4Desc') },
  { title: t('landing.security.regStep5Title'), desc: t('landing.security.regStep5Desc') },
  { title: t('landing.security.regStep6Title'), desc: t('landing.security.regStep6Desc') },
])

const shareSteps = computed(() => [
  { actor: 'alice', action: t('landing.security.shareStep1Action'), data: t('landing.security.shareStep1Data') },
  { actor: 'server', action: t('landing.security.shareStep2Action'), data: t('landing.security.shareStep2Data') },
  { actor: 'alice', action: t('landing.security.shareStep3Action'), data: t('landing.security.shareStep3Data') },
  { actor: 'server', action: t('landing.security.shareStep4Action'), data: t('landing.security.shareStep4Data') },
  { actor: 'bob', action: t('landing.security.shareStep5Action'), data: t('landing.security.shareStep5Data') },
])

// Upload demo — step by step
const chunks = ref(Array(4).fill(null).map((_, i) => ({ id: i })))
const uploadStep = ref(-1)
const uploadPhase = ref('idle')   // 'idle' | 'encrypting' | 'uploading' | 'done'
const uploadedChunks = ref([])
const currentNonce = ref('')

function generateNonce(chunkIndex) {
  const base = 'a3f8c2d1e5'
  return `${base}${String(chunkIndex).padStart(8, '0')}`
}

const uploadBtnLabel = computed(() => {
  if (uploadPhase.value === 'idle') return t('landing.security.uploadStart')
  if (uploadPhase.value === 'encrypting') return t('landing.security.uploadSendS3')
  if (uploadPhase.value === 'uploading') return uploadStep.value < chunks.value.length - 1 ? t('landing.security.uploadNextChunk') : t('landing.security.uploadFinish')
  return t('landing.security.uploadComplete')
})

function advanceUpload() {
  if (uploadPhase.value === 'idle') {
    uploadStep.value = 0
    uploadPhase.value = 'encrypting'
    currentNonce.value = generateNonce(0)
    return
  }
  if (uploadPhase.value === 'encrypting') {
    uploadPhase.value = 'uploading'
    return
  }
  if (uploadPhase.value === 'uploading') {
    uploadedChunks.value.push(uploadStep.value)
    const next = uploadStep.value + 1
    if (next < chunks.value.length) {
      uploadStep.value = next
      uploadPhase.value = 'encrypting'
      currentNonce.value = generateNonce(next)
    } else {
      uploadStep.value = -1
      uploadPhase.value = 'done'
    }
  }
}

function resetUpload() {
  uploadStep.value = -1
  uploadPhase.value = 'idle'
  uploadedChunks.value = []
  currentNonce.value = ''
}
</script>

<style scoped>
/* ── Global layout ─────────────────────────────────────────────────── */
.security-page {
  height: 100vh;
  overflow-y: auto;
  background: var(--background-color);
}

.container {
  max-width: 1100px;
  margin: 0 auto;
  padding: 0 2rem;
}

.section {
  padding: 5rem 0;
}

.dark-section {
  background: var(--card-color);
}

.section-label {
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--primary-color);
  margin-bottom: 0.75rem;
}

.section-sub {
  color: var(--secondary-text-color);
  font-size: 1.05rem;
  max-width: 680px;
  line-height: 1.7;
  margin-bottom: 2.5rem;
}

/* ── TOC sidebar flottant ─────────────────────────────────────────── */
.toc-sidebar {
  position: fixed;
  left: 1.25rem;
  top: 50%;
  transform: translateY(-50%);
  z-index: 200;
}

.toc-panel {
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 0.75rem 0.75rem;
  box-shadow: 0 4px 20px rgba(0,0,0,0.2);
  overflow: hidden;
  max-width: 210px;
  max-height: 80vh;
  transition: max-width 0.25s ease;
}

.toc-panel.folded {
  max-width: 115px;
}

.toc-title {
  font-size: 0.65rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--secondary-text-color);
  margin-bottom: 0.5rem;
  white-space: nowrap;
  padding: 0 0.25rem;
  display: flex;
  align-items: center;
  gap: 0.35rem;
  cursor: pointer;
  user-select: none;
}

.toc-title:hover {
  color: var(--primary-color);
}

.toc-chevron {
  width: 12px;
  height: 12px;
  flex-shrink: 0;
  transition: transform 0.2s ease;
  transform: rotate(-90deg);
}

.toc-chevron.open {
  transform: rotate(0deg);
}

.toc-nav-wrapper {
  overflow: hidden;
  max-height: 500px;
  transition: max-height 0.25s ease, opacity 0.2s ease;
  opacity: 1;
}

.toc-nav-wrapper.hidden {
  max-height: 0;
  opacity: 0;
}

.toc-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.4rem 0.5rem;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s;
  margin-bottom: 2px;
  white-space: nowrap;
}

.toc-item:hover {
  background: var(--hover-background-color);
  color: var(--primary-color);
}

.toc-item.active {
  background: rgba(250, 114, 104, 0.12);
  color: var(--primary-color);
}

.toc-item .toc-num {
  font-size: 0.6rem;
  font-weight: 700;
  font-family: monospace;
  color: var(--primary-color);
  opacity: 0.6;
  flex-shrink: 0;
}

.toc-item.active .toc-num {
  opacity: 1;
}

.toc-label {
  font-size: 0.78rem;
  font-weight: 500;
  color: var(--secondary-text-color);
}

.toc-item:hover .toc-label,
.toc-item.active .toc-label {
  color: var(--primary-color);
}

@media (max-width: 1100px) {
  .toc-sidebar { display: none; }
}

/* ── Hero ─────────────────────────────────────────────────────────── */
.sec-hero {
  padding: 7rem 0 4rem;
  text-align: center;
}

.hero-eyebrow {
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--primary-color);
  margin-bottom: 1rem;
}

.sec-hero h1 {
  font-size: clamp(2rem, 5vw, 3.5rem);
  line-height: 1.15;
  margin-bottom: 1.5rem;
}

.highlight {
  background: linear-gradient(135deg, var(--primary-color), var(--secondary-color));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.hero-sub {
  color: var(--secondary-text-color);
  font-size: 1.1rem;
  max-width: 600px;
  margin: 0 auto 2rem;
  line-height: 1.7;
}

.hero-pill-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  justify-content: center;
}

.pill {
  padding: 0.3rem 0.9rem;
  border-radius: 999px;
  font-size: 0.78rem;
  font-weight: 600;
  font-family: monospace;
}

.pill.green { background: rgba(34,197,94,0.15); color: #22c55e; border: 1px solid rgba(34,197,94,0.3); }
.pill.blue { background: rgba(59,130,246,0.15); color: #60a5fa; border: 1px solid rgba(59,130,246,0.3); }
.pill.purple { background: rgba(168,85,247,0.15); color: #c084fc; border: 1px solid rgba(168,85,247,0.3); }
.pill.orange { background: rgba(249,115,22,0.15); color: #fb923c; border: 1px solid rgba(249,115,22,0.3); }
.pill.teal { background: rgba(20,184,166,0.15); color: #2dd4bf; border: 1px solid rgba(20,184,166,0.3); }

/* ── ZK Comparison ────────────────────────────────────────────────── */
.zk-comparison {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.5rem;
}

@media (max-width: 768px) {
  .zk-comparison { grid-template-columns: 1fr; }
}

.cmp-card {
  border-radius: 16px;
  padding: 1.5rem;
  border: 1px solid var(--border-color);
}

.cmp-card.bad { border-color: rgba(239,68,68,0.3); background: rgba(239,68,68,0.04); }
.cmp-card.good { border-color: rgba(34,197,94,0.3); background: rgba(34,197,94,0.04); }

.cmp-header {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  font-weight: 600;
  margin-bottom: 1.25rem;
  font-size: 0.95rem;
}

.cmp-card.bad .cmp-header { color: #ef4444; }
.cmp-card.good .cmp-header { color: #22c55e; }
.cmp-header svg { width: 18px; height: 18px; flex-shrink: 0; }

.cmp-flow {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
  margin-bottom: 1rem;
}

.cmp-node {
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 0.5rem 0.75rem;
  font-size: 0.78rem;
  text-align: center;
  min-width: 70px;
}

.user-node { border-color: var(--primary-color); color: var(--primary-color); }
.server-bad { border-color: #ef4444; background: rgba(239,68,68,0.1); color: #ef4444; }
.server-good { border-color: #22c55e; background: rgba(34,197,94,0.1); color: #22c55e; }
.encrypted-node { border-color: #3b82f6; color: #60a5fa; }

.cmp-arrow {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  flex: 1;
  min-width: 50px;
}

.arrow-label {
  font-size: 0.65rem;
  color: var(--secondary-text-color);
  white-space: nowrap;
}

.arrow-label-top { margin-bottom: 2px; }

.cmp-arrow svg { width: 100%; height: 20px; }

.cmp-note {
  font-size: 0.82rem;
  color: var(--secondary-text-color);
  line-height: 1.5;
  margin: 0;
}

/* ── Key Hierarchy ────────────────────────────────────────────────── */
.key-hierarchy {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0;
  margin: 2.5rem 0 1rem;
}

.kh-level {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
}

.kh-node {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 1rem 1.5rem;
  border-radius: 14px;
  border: 2px solid var(--border-color);
  background: var(--background-color);
  cursor: pointer;
  transition: all 0.2s ease;
  text-align: center;
  min-width: 220px;
}

.kh-node:hover, .kh-node.active {
  transform: scale(1.04);
  box-shadow: 0 0 0 3px var(--primary-color);
}

.kh-password { border-color: var(--primary-color); }
.kh-kek { border-color: #f59e0b; }
.kh-master { border-color: #3b82f6; }
.kh-rsa { border-color: #8b5cf6; }
.kh-folder { border-color: #f59e0b; }
.kh-file { border-color: #22c55e; }

.kh-node.active.kh-password { background: rgba(250,114,104,0.1); }
.kh-node.active.kh-kek { background: rgba(245,158,11,0.1); }
.kh-node.active.kh-master { background: rgba(59,130,246,0.1); }
.kh-node.active.kh-rsa { background: rgba(139,92,246,0.1); }
.kh-node.active.kh-folder { background: rgba(245,158,11,0.1); }
.kh-node.active.kh-file { background: rgba(34,197,94,0.1); }

.kh-icon { font-size: 1.8rem; margin-bottom: 0.3rem; }
.kh-icon.small { font-size: 1.3rem; }
.kh-label { font-weight: 700; font-size: 1rem; }
.kh-label.small { font-size: 0.85rem; }
.kh-sub { font-size: 0.72rem; color: var(--secondary-text-color); margin-top: 0.2rem; }

.kh-connector {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.3rem;
  padding: 0.4rem 0;
}

.kh-algo-badge {
  font-size: 0.68rem;
  color: var(--secondary-text-color);
  background: var(--hover-background-color);
  padding: 0.2rem 0.6rem;
  border-radius: 4px;
  font-family: monospace;
  white-space: nowrap;
}

.kh-algo-badge.small { font-size: 0.62rem; }

.kh-line {
  width: 2px;
  height: 30px;
  background: linear-gradient(to bottom, var(--primary-color), var(--border-color));
  position: relative;
  overflow: hidden;
}

.kh-line-short {
  width: 2px;
  height: 20px;
  background: linear-gradient(to bottom, var(--primary-color), var(--border-color));
  position: relative;
  overflow: hidden;
  margin: 0 auto;
}

.animated-line::after {
  content: '';
  position: absolute;
  top: -100%;
  left: 0;
  width: 100%;
  height: 50%;
  background: var(--primary-color);
  animation: flow-down 1.5s linear infinite;
}

@keyframes flow-down {
  to { top: 200%; }
}

.kh-branches {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
}

.kh-branch-line {
  width: 2px;
  height: 20px;
  background: var(--border-color);
  margin: 0 auto;
}

.kh-branch-row {
  display: flex;
  gap: 4rem;
  justify-content: center;
  position: relative;
}

.kh-branch-row::before {
  content: '';
  position: absolute;
  top: 0;
  left: calc(50% - 2rem);
  width: calc(4rem + 4px);
  height: 2px;
  background: var(--border-color);
}

.kh-branch-col {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0;
}

.kh-node {
  position: relative;
}

.kh-tooltip {
  position: absolute;
  left: calc(100% + 14px);
  top: 50%;
  transform: translateY(-50%);
  width: 240px;
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 10px;
  padding: 0.75rem 1rem;
  font-size: 0.82rem;
  line-height: 1.55;
  color: var(--main-text-color);
  box-shadow: 0 6px 24px rgba(0,0,0,0.2);
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.18s ease, transform 0.18s ease;
  transform: translateY(-50%) translateX(-4px);
  z-index: 100;
}

.kh-tooltip::before {
  content: '';
  position: absolute;
  right: 100%;
  top: 50%;
  transform: translateY(-50%);
  border: 6px solid transparent;
  border-right-color: var(--border-color);
}

.kh-tooltip::after {
  content: '';
  position: absolute;
  right: calc(100% - 1px);
  top: 50%;
  transform: translateY(-50%);
  border: 6px solid transparent;
  border-right-color: var(--card-color);
}

.kh-tooltip-left {
  left: auto;
  right: calc(100% + 14px);
  transform: translateY(-50%) translateX(4px);
}

.kh-tooltip-left::before {
  right: auto;
  left: 100%;
  border-right-color: transparent;
  border-left-color: var(--border-color);
}

.kh-tooltip-left::after {
  right: auto;
  left: calc(100% - 1px);
  border-right-color: transparent;
  border-left-color: var(--card-color);
}

.kh-node:hover .kh-tooltip {
  opacity: 1;
  transform: translateY(-50%) translateX(0);
  pointer-events: auto;
}

.kh-node:hover .kh-tooltip-left {
  transform: translateY(-50%) translateX(0);
}

.kh-tooltip strong {
  display: block;
  margin-bottom: 0.3rem;
  color: var(--primary-color);
  font-size: 0.8rem;
}

.kh-tooltip code {
  background: var(--hover-background-color);
  padding: 0.1rem 0.35rem;
  border-radius: 4px;
  font-size: 0.78rem;
}

/* ── Registration flow ────────────────────────────────────────────── */
.flow-container {
  display: grid;
  grid-template-columns: 280px 1fr;
  gap: 2rem;
  align-items: start;
}

@media (max-width: 768px) {
  .flow-container { grid-template-columns: 1fr; }
}

.flow-steps {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.flow-step {
  display: flex;
  gap: 0.75rem;
  padding: 0.75rem;
  border-radius: 10px;
  border: 1px solid var(--border-color);
  cursor: pointer;
  transition: all 0.2s;
  opacity: 0.6;
}

.flow-step:hover { opacity: 0.8; background: var(--hover-background-color); }
.flow-step.active { opacity: 1; border-color: var(--primary-color); background: rgba(250,114,104,0.08); }
.flow-step.done { opacity: 0.7; border-color: #22c55e; }

.step-num {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: var(--border-color);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.72rem;
  font-weight: 700;
  flex-shrink: 0;
}

.flow-step.active .step-num { background: var(--primary-color); color: white; }
.flow-step.done .step-num { background: #22c55e; color: white; }

.step-title { font-weight: 600; font-size: 0.85rem; margin-bottom: 0.2rem; }
.step-desc { font-size: 0.75rem; color: var(--secondary-text-color); line-height: 1.4; }

.flow-visual {
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 14px;
  padding: 2rem;
  min-height: 220px;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* Animation nodes */
.anim-scene {
  display: flex;
  align-items: center;
  gap: 1.5rem;
  flex-wrap: wrap;
  justify-content: center;
  width: 100%;
  animation: scene-in 0.3s ease;
}

@keyframes scene-in {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}

.anim-node {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 1rem 1.25rem;
  border-radius: 12px;
  border: 1px solid var(--border-color);
  background: var(--background-color);
  text-align: center;
  min-width: 110px;
}

.anim-node.small-node { min-width: 80px; padding: 0.6rem 0.8rem; }

.node-icon { font-size: 1.5rem; margin-bottom: 0.3rem; }
.node-label { font-weight: 600; font-size: 0.82rem; }
.node-sub { font-size: 0.68rem; color: var(--secondary-text-color); margin-top: 0.2rem; }

.input-node { border-color: var(--primary-color); }
.kek-node { border-color: #f59e0b; }
.master-node { border-color: #3b82f6; }
.rng-node { border-color: #8b5cf6; }
.encrypted-master-node { border-color: #22c55e; }
.pub-node { border-color: #22c55e; }
.priv-node { border-color: #8b5cf6; }
.hash-node { border-color: #f59e0b; }
.enc-rec-node { border-color: #22c55e; }

.anim-pipe {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  flex: 1;
  min-width: 100px;
}

.pipe-label {
  font-size: 0.65rem;
  color: var(--secondary-text-color);
  background: var(--hover-background-color);
  padding: 0.15rem 0.5rem;
  border-radius: 4px;
}

.pipe-algo {
  font-size: 0.7rem;
  color: var(--secondary-text-color);
  text-align: center;
  line-height: 1.4;
}

.anim-packet {
  padding: 0.3rem 0.75rem;
  border-radius: 6px;
  font-size: 0.72rem;
  font-weight: 700;
  font-family: monospace;
  animation: none;
}

.anim-packet.flowing {
  animation: packet-flow 1s ease-in-out infinite alternate;
}

.pw-packet { background: rgba(250,114,104,0.2); color: var(--primary-color); }
.mk-packet { background: rgba(59,130,246,0.2); color: #60a5fa; }
.enc-packet { background: rgba(34,197,94,0.2); color: #22c55e; }
.rsa-packet { background: rgba(139,92,246,0.2); color: #c084fc; }

@keyframes packet-flow {
  from { transform: translateX(-8px); opacity: 0.7; }
  to { transform: translateX(8px); opacity: 1; }
}

/* Wrap scene */
.wrap-scene { flex-direction: column; }
.wrap-inputs { display: flex; gap: 1rem; justify-content: center; }
.wrap-arrow { display: flex; flex-direction: column; align-items: center; gap: 0.3rem; }
.wrap-svg { width: 80px; height: 40px; }

/* RSA split */
.rsa-split {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

/* Recovery split */
.recovery-split { display: flex; gap: 1.5rem; flex-wrap: wrap; justify-content: center; }
.rec-branch { display: flex; flex-direction: column; align-items: center; gap: 0.5rem; }
.branch-label { font-size: 0.65rem; color: var(--secondary-text-color); text-align: center; font-family: monospace; }

/* Server scene */
.server-scene { flex-direction: row; align-items: flex-start; gap: 2rem; flex-wrap: wrap; }
.server-sends { display: flex; flex-direction: column; gap: 0.4rem; flex: 1; min-width: 200px; }
.send-item { display: flex; align-items: center; gap: 0.5rem; font-size: 0.8rem; padding: 0.3rem 0.6rem; border-radius: 6px; }
.send-item.safe { background: rgba(34,197,94,0.1); color: #22c55e; }
.send-item.never { background: rgba(239,68,68,0.1); color: #ef4444; }
.send-icon { font-size: 0.9rem; flex-shrink: 0; }

.server-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 1.5rem;
  border-radius: 12px;
  border: 1px solid var(--border-color);
  background: var(--background-color);
  min-width: 150px;
}

.server-icon { font-size: 2rem; margin-bottom: 0.5rem; }
.server-label { font-weight: 700; font-size: 0.9rem; }
.server-note { font-size: 0.72rem; color: var(--secondary-text-color); text-align: center; margin-top: 0.4rem; line-height: 1.4; }

/* Step nav */
.step-nav {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 1.5rem;
  margin-top: 1.5rem;
}

.step-counter { font-size: 0.85rem; color: var(--secondary-text-color); }

.step-btn {
  background: var(--card-color);
  border: 1px solid var(--border-color);
  color: var(--main-text-color);
  padding: 0.5rem 1.25rem;
  border-radius: 8px;
  cursor: pointer;
  font-size: 0.85rem;
  transition: all 0.2s;
}

.step-btn:hover:not(:disabled) { border-color: var(--primary-color); color: var(--primary-color); }
.step-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.step-btn.primary { background: var(--primary-color); color: white; border-color: var(--primary-color); }
.step-btn.primary:hover:not(:disabled) { opacity: 0.9; }

/* ── Upload demo ──────────────────────────────────────────────────── */
.upload-demo {
  background: var(--background-color);
  border: 1px solid var(--border-color);
  border-radius: 16px;
  padding: 2rem;
}

.upload-file-bar {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  margin-bottom: 2rem;
}

.file-icon { font-size: 2.5rem; }

.file-info { flex: 1; }

.file-name { font-weight: 600; font-size: 0.95rem; margin-bottom: 0.75rem; }
.file-size { color: var(--secondary-text-color); font-weight: 400; margin-left: 0.5rem; }

.chunk-bar {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
  margin-bottom: 0.5rem;
}

.chunk {
  width: 60px;
  height: 40px;
  border-radius: 6px;
  border: 2px solid var(--border-color);
  background: var(--hover-background-color);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.75rem;
  font-weight: 700;
  color: var(--secondary-text-color);
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.chunk.encrypting {
  border-color: var(--primary-color);
  background: rgba(250,114,104,0.15);
  color: var(--primary-color);
  animation: encrypting-pulse 0.6s ease infinite alternate;
}

.chunk.encrypted {
  border-color: #f59e0b;
  background: rgba(245,158,11,0.1);
  color: #f59e0b;
}

.chunk.uploaded {
  border-color: #22c55e;
  background: rgba(34,197,94,0.1);
  color: #22c55e;
}

@keyframes encrypting-pulse {
  from { box-shadow: 0 0 0 0 rgba(250,114,104,0.4); }
  to { box-shadow: 0 0 0 6px rgba(250,114,104,0); }
}

.chunk-legend {
  display: flex;
  align-items: center;
  gap: 1rem;
  font-size: 0.72rem;
  color: var(--secondary-text-color);
  flex-wrap: wrap;
}

.legend-dot {
  display: inline-block;
  width: 10px;
  height: 10px;
  border-radius: 2px;
  border: 2px solid var(--border-color);
  margin-right: 2px;
}

.legend-dot.plain { background: var(--hover-background-color); }
.legend-dot.encrypting { background: rgba(250,114,104,0.3); border-color: var(--primary-color); }
.legend-dot.encrypted { background: rgba(245,158,11,0.2); border-color: #f59e0b; }
.legend-dot.uploaded { background: rgba(34,197,94,0.2); border-color: #22c55e; }

.upload-pipeline {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 1.5rem;
  flex-wrap: wrap;
}

.pipeline-stage {
  flex: 1;
  min-width: 120px;
  padding: 0.75rem 1rem;
  border-radius: 10px;
  border: 1px solid var(--border-color);
  background: var(--card-color);
  text-align: center;
  transition: all 0.3s;
  opacity: 0.5;
}

.pipeline-stage.active {
  opacity: 1;
  border-color: var(--primary-color);
  box-shadow: 0 0 12px rgba(250,114,104,0.2);
}

.stage-icon { font-size: 1.5rem; margin-bottom: 0.25rem; }
.stage-label { font-weight: 600; font-size: 0.82rem; }
.stage-sub { font-size: 0.7rem; color: var(--secondary-text-color); margin-top: 0.15rem; }

.pipeline-arrow {
  font-size: 1.5rem;
  color: var(--border-color);
  transition: color 0.3s;
}

.pipeline-arrow.active { color: var(--primary-color); }

.nonce-box {
  margin-top: 0.5rem;
  background: var(--hover-background-color);
  border-radius: 4px;
  padding: 0.2rem 0.4rem;
  font-size: 0.62rem;
  font-family: monospace;
}

.nonce-label { color: var(--secondary-text-color); margin-right: 0.3rem; }
.nonce-val { color: var(--primary-color); }

.chunk-detail {
  background: var(--hover-background-color);
  border-radius: 10px;
  padding: 1rem 1.25rem;
  margin-bottom: 1.5rem;
  animation: scene-in 0.2s ease;
}

.chunk-detail-title {
  font-size: 0.8rem;
  font-weight: 600;
  margin-bottom: 0.75rem;
  color: var(--secondary-text-color);
}

.chunk-bytes {
  display: flex;
  gap: 4px;
  align-items: stretch;
  height: 50px;
  border-radius: 6px;
  overflow: hidden;
}

.byte-block {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.65rem;
  font-weight: 600;
  text-align: center;
  line-height: 1.3;
  padding: 0 0.5rem;
}

.nonce-block { background: rgba(139,92,246,0.3); color: #c084fc; width: 80px; flex-shrink: 0; }
.cipher-block { background: rgba(59,130,246,0.2); color: #60a5fa; flex: 1; }
.tag-block { background: rgba(34,197,94,0.2); color: #22c55e; width: 80px; flex-shrink: 0; }

.upload-controls {
  display: flex;
  gap: 1rem;
  justify-content: center;
}

.upload-progress-hint {
  text-align: center;
  margin-top: 0.75rem;
  font-size: 0.8rem;
  color: var(--secondary-text-color);
}

.upload-progress-hint span { color: var(--primary-color); font-weight: 600; }

.upload-done-msg {
  text-align: center;
  margin-top: 0.75rem;
  font-size: 0.85rem;
  font-weight: 600;
  color: #22c55e;
  padding: 0.5rem 1rem;
  background: rgba(34,197,94,0.1);
  border-radius: 8px;
  border: 1px solid rgba(34,197,94,0.25);
}

/* ── Share flow ───────────────────────────────────────────────────── */
.share-flow {
  display: grid;
  grid-template-columns: 120px 1fr 120px;
  gap: 1.5rem;
  align-items: start;
  margin-bottom: 2rem;
}

@media (max-width: 768px) {
  .share-flow { grid-template-columns: 1fr; }
  .share-actor { flex-direction: row; justify-content: center; gap: 1rem; }
}

.share-actor {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.4rem;
  padding: 1rem 0.5rem;
  border-radius: 14px;
  border: 1px solid var(--border-color);
  text-align: center;
  position: sticky;
  top: 2rem;
}

.actor-avatar { font-size: 2rem; }
.actor-name { font-weight: 700; font-size: 0.9rem; }
.actor-device { font-size: 0.7rem; color: var(--secondary-text-color); }

.share-steps-col { display: flex; flex-direction: column; gap: 0.75rem; }

.share-step {
  display: flex;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  border-radius: 10px;
  border: 1px solid var(--border-color);
  opacity: 0.55;
  cursor: pointer;
  transition: all 0.2s;
}

.share-step:hover { opacity: 0.75; }
.share-step.active { opacity: 1; border-color: var(--primary-color); background: rgba(250,114,104,0.06); }
.share-step.done { opacity: 0.7; border-color: #22c55e; }

.share-step-num {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background: var(--border-color);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.7rem;
  font-weight: 700;
  flex-shrink: 0;
}

.share-step.active .share-step-num { background: var(--primary-color); color: white; }
.share-step.done .share-step-num { background: #22c55e; color: white; }

.share-step-actor {
  font-size: 0.7rem;
  font-weight: 700;
  padding: 0.1rem 0.4rem;
  border-radius: 4px;
  display: inline-block;
  margin-bottom: 0.2rem;
}

.share-step-actor.alice { background: rgba(250,114,104,0.15); color: var(--primary-color); }
.share-step-actor.server { background: rgba(59,130,246,0.15); color: #60a5fa; }
.share-step-actor.bob { background: rgba(34,197,94,0.15); color: #22c55e; }

.share-step-action { font-weight: 600; font-size: 0.83rem; margin-bottom: 0.2rem; }
.share-step-data { font-size: 0.72rem; color: var(--secondary-text-color); font-family: monospace; line-height: 1.4; }

.share-nav {
  display: flex;
  align-items: center;
  gap: 1rem;
  justify-content: center;
  margin-top: 0.5rem;
}

.share-visual-box {
  background: var(--card-color);
  border: 1px solid var(--border-color);
  border-radius: 14px;
  padding: 2rem;
  min-height: 130px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.share-visual {
  display: flex;
  align-items: center;
  gap: 1.25rem;
  flex-wrap: wrap;
  justify-content: center;
  animation: scene-in 0.25s ease;
}

.sv-node {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 0.75rem 1.25rem;
  border-radius: 10px;
  border: 1px solid var(--border-color);
  background: var(--background-color);
  text-align: center;
  min-width: 120px;
}

.sv-node.highlight-node { border-color: var(--primary-color); }
.sv-node.success-node { border-color: #22c55e; background: rgba(34,197,94,0.08); }

.sv-icon { font-size: 1.4rem; margin-bottom: 0.25rem; }
.sv-label { font-weight: 600; font-size: 0.82rem; }
.sv-sub { font-size: 0.68rem; color: var(--secondary-text-color); margin-top: 0.2rem; line-height: 1.4; }

.sv-arrow { font-size: 1.5rem; color: var(--secondary-text-color); }
.sv-plus { font-size: 1.5rem; color: var(--secondary-text-color); }

.sv-note {
  font-size: 0.8rem;
  color: var(--secondary-text-color);
  background: var(--hover-background-color);
  padding: 0.75rem 1rem;
  border-radius: 8px;
  max-width: 320px;
  text-align: center;
  line-height: 1.5;
}

.sv-pipe {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.4rem;
  flex: 1;
  min-width: 120px;
}

.sv-algo {
  font-size: 0.68rem;
  color: var(--secondary-text-color);
  text-align: center;
  line-height: 1.4;
  font-family: monospace;
}

.sv-packet {
  width: 40px;
  height: 12px;
  border-radius: 6px;
  background: var(--primary-color);
  animation: sv-flow 1s ease infinite alternate;
}

.sv-packet.blue-anim { background: #22c55e; }

@keyframes sv-flow {
  from { transform: translateX(-15px); opacity: 0.5; }
  to { transform: translateX(15px); opacity: 1; }
}

/* ── Service Worker diagram ───────────────────────────────────────── */
.sw-diagram {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  gap: 1rem;
  align-items: center;
  margin-bottom: 2.5rem;
}

@media (max-width: 768px) {
  .sw-diagram { grid-template-columns: 1fr; }
  .sw-barrier { flex-direction: row; }
}

.sw-zone {
  border-radius: 14px;
  padding: 1.5rem;
  border: 1px solid var(--border-color);
}

.page-zone { border-color: #f59e0b; background: rgba(245,158,11,0.04); }
.worker-zone { border-color: #22c55e; background: rgba(34,197,94,0.04); }

.sw-zone-label {
  font-size: 0.72rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  margin-bottom: 1rem;
}

.page-zone .sw-zone-label { color: #f59e0b; }
.worker-zone .sw-zone-label { color: #22c55e; }

.sw-items { display: flex; flex-direction: column; gap: 0.6rem; }

.sw-item {
  background: var(--background-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 0.5rem 0.75rem;
  font-size: 0.8rem;
}

.bad-item { border-color: rgba(239,68,68,0.4); background: rgba(239,68,68,0.06); color: #f87171; }
.key-item { border-color: #22c55e; background: rgba(34,197,94,0.1); color: #22c55e; font-weight: 600; }
.timeout-item { border-color: #f59e0b; background: rgba(245,158,11,0.08); color: #f59e0b; font-size: 0.78rem; }

.sw-request {
  font-size: 0.72rem;
  color: var(--secondary-text-color);
  font-family: monospace;
}

.sw-response {
  font-size: 0.72rem;
  line-height: 1.5;
}

.sw-badge {
  display: inline-block;
  background: rgba(34,197,94,0.2);
  color: #22c55e;
  font-size: 0.6rem;
  padding: 0.1rem 0.4rem;
  border-radius: 4px;
  font-family: monospace;
  margin-top: 0.3rem;
}

.sw-barrier {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
}

.barrier-line {
  width: 2px;
  height: 40px;
  background: linear-gradient(to bottom, transparent, var(--primary-color), transparent);
}

.barrier-label {
  font-size: 0.65rem;
  color: var(--primary-color);
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  writing-mode: vertical-rl;
  text-orientation: mixed;
  transform: rotate(180deg);
}

@media (max-width: 768px) {
  .barrier-line { width: 40px; height: 2px; }
  .barrier-label { writing-mode: horizontal-tb; transform: none; }
}

.sw-props {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
}

@media (max-width: 768px) {
  .sw-props { grid-template-columns: 1fr; }
}

.sw-prop {
  background: var(--background-color);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 1.25rem;
}

.prop-icon { font-size: 1.5rem; margin-bottom: 0.5rem; }
.prop-title { font-weight: 700; font-size: 0.9rem; margin-bottom: 0.4rem; }
.prop-desc { font-size: 0.78rem; color: var(--secondary-text-color); line-height: 1.5; }
.prop-desc code { background: var(--card-color); padding: 0.1rem 0.3rem; border-radius: 3px; font-size: 0.75rem; }

/* ── Visibility grid ──────────────────────────────────────────────── */
.visibility-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.5rem;
}

@media (max-width: 768px) {
  .visibility-grid { grid-template-columns: 1fr; }
}

.vis-col {
  border-radius: 16px;
  border: 1px solid var(--border-color);
  overflow: hidden;
}

.vis-header {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 1rem 1.25rem;
  font-size: 0.9rem;
}

.vis-header svg { width: 18px; height: 18px; flex-shrink: 0; }
.sees-header { background: rgba(59,130,246,0.08); border-bottom: 1px solid rgba(59,130,246,0.2); color: #60a5fa; }
.not-header { background: rgba(239,68,68,0.08); border-bottom: 1px solid rgba(239,68,68,0.2); color: #f87171; }

.vis-items { padding: 0.75rem; display: flex; flex-direction: column; gap: 0.4rem; }

.vis-item {
  padding: 0.5rem 0.75rem;
  border-radius: 7px;
  font-size: 0.8rem;
}

.vis-item.neutral { background: var(--hover-background-color); color: var(--main-text-color); }
.vis-item.never-item { background: rgba(239,68,68,0.08); color: #f87171; font-weight: 500; }

/* ── CTA ──────────────────────────────────────────────────────────── */
.sec-cta {
  padding: 5rem 0;
  text-align: center;
  background: var(--card-color);
}

.sec-cta h2 { margin-bottom: 1rem; }
.sec-cta p { color: var(--secondary-text-color); font-size: 1rem; margin-bottom: 2rem; }

.cta-row { display: flex; gap: 1rem; justify-content: center; flex-wrap: wrap; }

.btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1.75rem;
  border-radius: 10px;
  font-weight: 600;
  font-size: 0.9rem;
  text-decoration: none;
  transition: all 0.2s;
}

.btn-primary {
  background: var(--primary-color);
  color: white;
}

.btn-primary:hover { opacity: 0.9; transform: translateY(-1px); }

.btn-secondary {
  background: transparent;
  color: var(--main-text-color);
  border: 1px solid var(--border-color);
}

.btn-secondary:hover { border-color: var(--primary-color); color: var(--primary-color); }
</style>
