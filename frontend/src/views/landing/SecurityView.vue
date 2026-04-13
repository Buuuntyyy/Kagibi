<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="security-page">
    <LandingNav />

    <!-- Hero -->
    <section class="sec-hero">
      <div class="container">
        <div class="hero-eyebrow">Architecture Zero-Knowledge</div>
        <h1>
          Pourquoi Kagibi<br>
          <span class="highlight">ne peut pas lire vos données</span>
        </h1>
        <p class="hero-sub">
          Ce n'est pas une promesse commerciale — c'est une contrainte mathématique.
          Voici, concrètement, ce qui se passe sur votre appareil et ce que le serveur ne voit jamais.
        </p>
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
          Sommaire
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
        <div class="section-label">Le principe</div>
        <h2>Le serveur stocke des données qu'il ne peut pas lire</h2>
        <p class="section-sub">
          Chaque fois que vous chiffrez un fichier, toute la cryptographie se passe dans votre navigateur,
          jamais côté serveur. Le serveur reçoit uniquement des octets chiffrés.
          Sans votre mot de passe — qu'il ne connaît pas — il ne peut rien en faire.
        </p>

        <div class="zk-comparison">
          <div class="cmp-card bad">
            <div class="cmp-header">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>
              Cloud classique (Google Drive, etc.)
            </div>
            <div class="cmp-flow">
              <div class="cmp-node user-node">Votre fichier</div>
              <div class="cmp-arrow bad-arrow">
                <span class="arrow-label">Fichier en clair</span>
                <svg viewBox="0 0 60 20"><path d="M0 10 L50 10" stroke="#ef4444" stroke-width="2"/><polygon points="50,5 60,10 50,15" fill="#ef4444"/></svg>
              </div>
              <div class="cmp-node server-bad">Serveur<br><small>(lit tout)</small></div>
              <div class="cmp-arrow bad-arrow">
                <svg viewBox="0 0 60 20"><path d="M0 10 L50 10" stroke="#ef4444" stroke-width="2"/><polygon points="50,5 60,10 50,15" fill="#ef4444"/></svg>
              </div>
              <div class="cmp-node">Stockage</div>
            </div>
            <p class="cmp-note">Le serveur chiffre lui-même → il détient la clé → il peut tout lire, tout livrer à la justice, tout pirater.</p>
          </div>

          <div class="cmp-card good">
            <div class="cmp-header">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 11.08V12a10 10 0 11-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
              Kagibi — Zero-Knowledge
            </div>
            <div class="cmp-flow">
              <div class="cmp-node user-node">Votre fichier</div>
              <div class="cmp-arrow good-arrow">
                <span class="arrow-label arrow-label-top">Chiffrement local</span>
                <svg viewBox="0 0 60 20"><path d="M0 10 L50 10" stroke="#22c55e" stroke-width="2"/><polygon points="50,5 60,10 50,15" fill="#22c55e"/></svg>
              </div>
              <div class="cmp-node encrypted-node">🔒 Chiffré<br><small>(AES-256)</small></div>
              <div class="cmp-arrow good-arrow">
                <span class="arrow-label">Octets opaques</span>
                <svg viewBox="0 0 60 20"><path d="M0 10 L50 10" stroke="#22c55e" stroke-width="2"/><polygon points="50,5 60,10 50,15" fill="#22c55e"/></svg>
              </div>
              <div class="cmp-node server-good">Serveur<br><small>(ne voit rien)</small></div>
            </div>
            <p class="cmp-note">Le serveur ne reçoit que du chiffré. Sans votre mot de passe, il est mathématiquement incapable de déchiffrer.</p>
          </div>
        </div>
      </div>
    </section>

    <!-- Hiérarchie des clés -->
    <section id="hierarchie" class="section dark-section">
      <div class="container">
        <div class="section-label">Architecture</div>
        <h2>La hiérarchie des clés</h2>
        <p class="section-sub">
          Toutes vos données sont protégées par une chaîne de clés imbriquées.
          Chaque clé en chiffre une autre — seul votre mot de passe permet de tout débloquer.
        </p>

        <div class="key-hierarchy">
          <!-- Niveau 0: Mot de passe -->
          <div class="kh-level">
            <div class="kh-node kh-password" :class="{ active: activeKey === 'password' }" @mouseenter="activeKey = 'password'" @mouseleave="activeKey = null">
              <div class="kh-icon">🔑</div>
              <div class="kh-label">Mot de passe</div>
              <div class="kh-sub">Jamais envoyé au serveur</div>
              <div class="kh-tooltip">
                <strong>Mot de passe</strong> — Passé à Argon2id dans un Web Worker séparé, il n'est jamais sérialisé, jamais envoyé au serveur. Le hash produit est la KEK.
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
              <div class="kh-sub">Key Encryption Key · 256 bits</div>
              <div class="kh-tooltip">
                <strong>KEK (Key Encryption Key)</strong> — Dérivée de votre mot de passe via Argon2id (64 MB RAM, 4 passes). Elle sert uniquement à chiffrer/déchiffrer la Master Key. Elle n'est jamais stockée.
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
              <div class="kh-sub">AES-256 · Générée aléatoirement · Service Worker</div>
              <div class="kh-tooltip">
                <strong>Master Key</strong> — Clé AES-256 générée aléatoirement à l'inscription. Stockée chiffrée (par la KEK) sur le serveur. En session dans le Service Worker avec <code>extractable: false</code> — le JS ne peut pas l'exporter.
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
                    <div class="kh-label small">Clé privée RSA</div>
                    <div class="kh-sub">RSA-OAEP 4096 bits<br>Pour le partage</div>
                    <div class="kh-tooltip kh-tooltip-left">
                      <strong>Clé privée RSA-OAEP 4096</strong> — Générée à l'inscription, chiffrée par la Master Key avant envoi au serveur. Utilisée pour le partage : l'expéditeur chiffre une clé de fichier avec votre clé publique.
                    </div>
                  </div>
                </div>
                <div class="kh-branch-col">
                  <div class="kh-algo-badge small">AES-GCM wrap</div>
                  <div class="kh-line-short animated-line"></div>
                  <div class="kh-node kh-folder" :class="{ active: activeKey === 'folder' }" @mouseenter="activeKey = 'folder'" @mouseleave="activeKey = null">
                    <div class="kh-icon small">📁</div>
                    <div class="kh-label small">Clé de dossier</div>
                    <div class="kh-sub">AES-256 par dossier<br>Chiffrée par la Master Key</div>
                    <div class="kh-tooltip">
                      <strong>Clé de dossier</strong> — Chaque dossier a sa propre clé AES-256, chiffrée par la Master Key. Partager un dossier = partager cette clé chiffrée par la clé publique RSA du destinataire.
                    </div>
                  </div>
                  <div class="kh-line-short animated-line"></div>
                  <div class="kh-node kh-file" :class="{ active: activeKey === 'file' }" @mouseenter="activeKey = 'file'" @mouseleave="activeKey = null">
                    <div class="kh-icon small">📄</div>
                    <div class="kh-label small">Clé de fichier</div>
                    <div class="kh-sub">AES-256 par fichier<br>Chiffrée par la clé dossier</div>
                    <div class="kh-tooltip">
                      <strong>Clé de fichier</strong> — Chaque fichier a sa propre clé AES-256, chiffrée par la clé de dossier. Permet la rotation des clés et la révocation granulaire par fichier.
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
        <div class="section-label">Flux 1 · Inscription</div>
        <h2>Ce qui se passe quand vous créez un compte</h2>
        <p class="section-sub">Toute la cryptographie est réalisée dans votre navigateur avant qu'un seul octet soit envoyé.</p>

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
                <div class="node-label">mot_de_passe</div>
                <div class="node-sub">Saisi dans le navigateur</div>
              </div>
              <div class="anim-pipe">
                <div class="pipe-label">Web Worker</div>
                <div class="anim-packet pw-packet" :class="{ flowing: regStep === 0 }">pw</div>
                <div class="pipe-algo">Argon2id<br>64 MB · 4 passes<br>sel 16 octets (CSPRNG)</div>
              </div>
              <div class="anim-node kek-node">
                <div class="node-icon">🗝️</div>
                <div class="node-label">KEK</div>
                <div class="node-sub">256 bits · en mémoire seulement</div>
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
                <div class="pipe-algo">Génération aléatoire<br>non-extractable</div>
              </div>
              <div class="anim-node master-node">
                <div class="node-icon">🏛️</div>
                <div class="node-label">Master Key</div>
                <div class="node-sub">AES-256-GCM · Service Worker</div>
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
                <div class="node-sub">Stocké sur le serveur · inutilisable sans KEK</div>
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
                  <div class="node-label">Clé publique</div>
                  <div class="node-sub">→ serveur (en clair)</div>
                </div>
                <div class="anim-node priv-node">
                  <div class="node-icon">🔐</div>
                  <div class="node-label">Clé privée</div>
                  <div class="node-sub">→ chiffrée par Master Key<br>→ serveur</div>
                </div>
              </div>
            </div>

            <!-- Step 4: recovery -->
            <div v-if="regStep === 4" class="anim-scene">
              <div class="anim-node rng-node">
                <div class="node-icon">🎲</div>
                <div class="node-label">Code de récupération</div>
                <div class="node-sub">32 octets aléatoires</div>
              </div>
              <div class="recovery-split">
                <div class="rec-branch">
                  <div class="branch-label">SHA-256 (hash)</div>
                  <div class="anim-node hash-node">
                    <div class="node-icon">🔏</div>
                    <div class="node-label">recovery_hash</div>
                    <div class="node-sub">→ serveur pour vérification</div>
                  </div>
                </div>
                <div class="rec-branch">
                  <div class="branch-label">Argon2id → KEK recovery → AES-GCM wrap</div>
                  <div class="anim-node enc-rec-node">
                    <div class="node-icon">🔒</div>
                    <div class="node-label">encrypted_master_key_recovery</div>
                    <div class="node-sub">→ serveur · déverrouillable par le code</div>
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
                  <span>sel Argon2 (public)</span>
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
                  <span>mot de passe → JAMAIS</span>
                </div>
                <div class="send-item never">
                  <span class="send-icon">🚫</span>
                  <span>Master Key en clair → JAMAIS</span>
                </div>
                <div class="send-item never">
                  <span class="send-icon">🚫</span>
                  <span>KEK → JAMAIS</span>
                </div>
              </div>
              <div class="server-box">
                <div class="server-icon">🖥️</div>
                <div class="server-label">Serveur Kagibi</div>
                <div class="server-note">Stocke des chiffrés.<br>Ne peut rien déchiffrer.</div>
              </div>
            </div>
          </div>
        </div>

        <div class="step-nav">
          <button class="step-btn" @click="regStep = Math.max(0, regStep - 1)" :disabled="regStep === 0">← Précédent</button>
          <span class="step-counter">{{ regStep + 1 }} / {{ registrationSteps.length }}</span>
          <button class="step-btn" @click="regStep = Math.min(registrationSteps.length - 1, regStep + 1)" :disabled="regStep === registrationSteps.length - 1">Suivant →</button>
        </div>
      </div>
    </section>

    <!-- Upload fichier animé -->
    <section id="upload" class="section dark-section">
      <div class="container">
        <div class="section-label">Flux 2 · Upload</div>
        <h2>Chiffrement d'un fichier en chunks</h2>
        <p class="section-sub">
          Les fichiers sont découpés en blocs de 10 Mo, chaque bloc est chiffré indépendamment
          avec un nonce unique (déterministe par numéro de chunk). Jamais de texte clair sur le réseau.
        </p>

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
                <span class="legend-dot plain"></span>En attente
                <span class="legend-dot encrypting"></span>Chiffrement AES-GCM
                <span class="legend-dot encrypted"></span>Chiffré
                <span class="legend-dot uploaded"></span>Uploadé sur S3
              </div>
            </div>
          </div>

          <div class="upload-pipeline">
            <div class="pipeline-stage" :class="{ active: uploadPhase === 'reading' || uploadPhase === 'encrypting' }">
              <div class="stage-icon">💾</div>
              <div class="stage-label">Lecture locale</div>
              <div class="stage-sub">FileReader API</div>
            </div>
            <div class="pipeline-arrow" :class="{ active: uploadPhase === 'encrypting' }">→</div>
            <div class="pipeline-stage" :class="{ active: uploadPhase === 'encrypting' }">
              <div class="stage-icon">⚙️</div>
              <div class="stage-label">Web Worker</div>
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
              <div class="stage-sub">Octets chiffrés uniquement</div>
            </div>
          </div>

          <div class="chunk-detail" v-if="uploadPhase === 'encrypting'">
            <div class="chunk-detail-title">Structure du chunk {{ uploadStep + 1 }} chiffré</div>
            <div class="chunk-bytes">
              <div class="byte-block nonce-block">Nonce<br>12 octets</div>
              <div class="byte-block cipher-block">Ciphertext AES-256-GCM<br>~10 Mo</div>
              <div class="byte-block tag-block">Auth Tag<br>16 octets</div>
            </div>
          </div>

          <div class="upload-controls">
            <button class="step-btn primary" @click="advanceUpload" :disabled="uploadPhase === 'done'">
              {{ uploadBtnLabel }}
            </button>
            <button class="step-btn" @click="resetUpload" :disabled="uploadPhase === 'idle'">Réinitialiser</button>
          </div>
          <div class="upload-progress-hint" v-if="uploadPhase !== 'idle' && uploadPhase !== 'done'">
            Chunk {{ uploadStep + 1 }} / {{ chunks.length }} —
            <span v-if="uploadPhase === 'encrypting'">chiffrement AES-256-GCM en cours…</span>
            <span v-else>upload vers OVH S3…</span>
          </div>
          <div class="upload-done-msg" v-if="uploadPhase === 'done'">
            ✓ {{ chunks.length }} chunks chiffrés et uploadés — aucun octet en clair n'a transité
          </div>
        </div>
      </div>
    </section>

    <!-- Partage RSA -->
    <section id="partage" class="section">
      <div class="container">
        <div class="section-label">Flux 3 · Partage</div>
        <h2>Partager un fichier sans que le serveur voit la clé</h2>
        <p class="section-sub">
          Quand Alice partage un fichier avec Bob, la clé de fichier est chiffrée avec la clé publique RSA de Bob.
          Le serveur ne peut pas la déchiffrer — seul Bob, avec sa clé privée, peut le faire.
        </p>

        <div class="share-flow">
          <div class="share-actor alice">
            <div class="actor-avatar">👩</div>
            <div class="actor-name">Alice</div>
            <div class="actor-device">Son navigateur</div>
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
                <div class="share-step-actor" :class="step.actor">{{ step.actor === 'alice' ? '👩 Alice' : step.actor === 'server' ? '🖥️ Serveur' : '👨 Bob' }}</div>
                <div class="share-step-action">{{ step.action }}</div>
                <div class="share-step-data">{{ step.data }}</div>
              </div>
            </div>

            <div class="share-nav">
              <button class="step-btn" @click="shareStep = Math.max(0, shareStep - 1)" :disabled="shareStep === 0">← Précédent</button>
              <span class="step-counter">{{ shareStep + 1 }} / {{ shareSteps.length }}</span>
              <button class="step-btn" @click="shareStep = Math.min(shareSteps.length - 1, shareStep + 1)" :disabled="shareStep === shareSteps.length - 1">Suivant →</button>
            </div>
          </div>

          <div class="share-actor bob">
            <div class="actor-avatar">👨</div>
            <div class="actor-name">Bob</div>
            <div class="actor-device">Son navigateur</div>
          </div>
        </div>

        <div class="share-visual-box">
          <div class="share-visual" v-if="shareStep === 0">
            <div class="sv-node">
              <div class="sv-icon">📄</div>
              <div class="sv-label">Fichier (chiffré)</div>
            </div>
            <div class="sv-plus">+</div>
            <div class="sv-node highlight-node">
              <div class="sv-icon">🔑</div>
              <div class="sv-label">Clé de fichier d'Alice</div>
              <div class="sv-sub">AES-256 · en mémoire</div>
            </div>
          </div>
          <div class="share-visual" v-else-if="shareStep === 1">
            <div class="sv-node">
              <div class="sv-icon">🖥️</div>
              <div class="sv-label">Serveur</div>
            </div>
            <div class="sv-arrow">→</div>
            <div class="sv-node highlight-node">
              <div class="sv-icon">🔓</div>
              <div class="sv-label">Clé publique de Bob</div>
              <div class="sv-sub">RSA-OAEP 4096 bits</div>
            </div>
          </div>
          <div class="share-visual" v-else-if="shareStep === 2">
            <div class="sv-node">
              <div class="sv-icon">🔑</div>
              <div class="sv-label">Clé de fichier</div>
            </div>
            <div class="sv-pipe">
              <div class="sv-algo">RSA-OAEP encrypt<br>avec clé publique Bob</div>
              <div class="sv-packet rsa-anim"></div>
            </div>
            <div class="sv-node highlight-node">
              <div class="sv-icon">🔒</div>
              <div class="sv-label">Clé chiffrée pour Bob</div>
              <div class="sv-sub">Seul Bob peut déchiffrer</div>
            </div>
          </div>
          <div class="share-visual" v-else-if="shareStep === 3">
            <div class="sv-node">
              <div class="sv-icon">🖥️</div>
              <div class="sv-label">Serveur stocke</div>
              <div class="sv-sub">Clé chiffrée RSA<br>+ Fichier chiffré AES</div>
            </div>
            <div class="sv-note">Le serveur ne peut rien déchiffrer — il ne détient ni la clé privée de Bob, ni la Master Key d'Alice.</div>
          </div>
          <div class="share-visual" v-else-if="shareStep === 4">
            <div class="sv-node">
              <div class="sv-icon">🔒</div>
              <div class="sv-label">Clé chiffrée RSA</div>
            </div>
            <div class="sv-pipe">
              <div class="sv-algo">RSA-OAEP decrypt<br>clé privée de Bob</div>
              <div class="sv-packet rsa-anim blue-anim"></div>
            </div>
            <div class="sv-node highlight-node success-node">
              <div class="sv-icon">🔑</div>
              <div class="sv-label">Clé de fichier récupérée</div>
              <div class="sv-sub">Bob déchiffre le fichier ✓</div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Service Worker -->
    <section id="session" class="section dark-section">
      <div class="container">
        <div class="section-label">Protection en session</div>
        <h2>La Master Key n'est jamais exposée au JavaScript</h2>
        <p class="section-sub">
          Même si une extension malveillante ou une faille XSS injecte du JavaScript dans la page,
          elle ne peut pas voler votre Master Key. Voici pourquoi.
        </p>

        <div class="sw-diagram">
          <div class="sw-zone page-zone">
            <div class="sw-zone-label">Page (contexte JS normal)</div>
            <div class="sw-items">
              <div class="sw-item">Code de l'application</div>
              <div class="sw-item bad-item">⚠️ Extensions navigateur</div>
              <div class="sw-item bad-item">⚠️ Potentiel XSS</div>
              <div class="sw-item">
                <div class="sw-request">Demande : "donne-moi la Master Key"</div>
              </div>
            </div>
          </div>

          <div class="sw-barrier">
            <div class="barrier-line"></div>
            <div class="barrier-label">Frontière Service Worker</div>
            <div class="barrier-line"></div>
          </div>

          <div class="sw-zone worker-zone">
            <div class="sw-zone-label">Service Worker (contexte isolé)</div>
            <div class="sw-items">
              <div class="sw-item key-item">
                🏛️ Master Key
                <div class="sw-badge">extractable: false</div>
              </div>
              <div class="sw-item">
                <div class="sw-response">✓ Utilise la clé pour chiffrer/déchiffrer<br>✗ Ne l'exporte JAMAIS</div>
              </div>
              <div class="sw-item timeout-item">
                ⏱️ Timeout 30 min → clé effacée
              </div>
            </div>
          </div>
        </div>

        <div class="sw-props">
          <div class="sw-prop">
            <div class="prop-icon">🔒</div>
            <div class="prop-title">Non-extractable</div>
            <div class="prop-desc">La WebCrypto API interdit l'export de la clé. <code>crypto.subtle.exportKey()</code> retourne une erreur.</div>
          </div>
          <div class="sw-prop">
            <div class="prop-icon">⏱️</div>
            <div class="prop-title">Timeout d'inactivité</div>
            <div class="prop-desc">30 minutes sans activité → Master Key effacée du Service Worker → déconnexion automatique.</div>
          </div>
          <div class="sw-prop">
            <div class="prop-icon">👁️</div>
            <div class="prop-title">Détection XSS</div>
            <div class="prop-desc">Un MutationObserver surveille les scripts injectés dans le DOM. Les handlers <code>onerror</code> sont bloqués.</div>
          </div>
        </div>
      </div>
    </section>

    <!-- Ce que le serveur voit/ne voit pas -->
    <section id="serveur" class="section">
      <div class="container">
        <div class="section-label">Résumé</div>
        <h2>Ce que le serveur voit — et ne voit pas</h2>

        <div class="visibility-grid">
          <div class="vis-col vis-sees">
            <div class="vis-header sees-header">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>
              Ce que le serveur <strong>voit</strong>
            </div>
            <div class="vis-items">
              <div class="vis-item neutral">Email / identifiant</div>
              <div class="vis-item neutral">Salt Argon2 (public par nature)</div>
              <div class="vis-item neutral">encrypted_master_key (chiffré, inutilisable)</div>
              <div class="vis-item neutral">encrypted_private_key (chiffré)</div>
              <div class="vis-item neutral">Clé publique RSA (publique)</div>
              <div class="vis-item neutral">recovery_hash (SHA-256, irréversible)</div>
              <div class="vis-item neutral">Métadonnées fichier (nom chiffré si activé, taille, MIME)</div>
              <div class="vis-item neutral">Fichiers chiffrés AES-256-GCM</div>
              <div class="vis-item neutral">Clés de fichier chiffrées</div>
            </div>
          </div>

          <div class="vis-col vis-not">
            <div class="vis-header not-header">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M17.94 17.94A10.07 10.07 0 0112 20c-7 0-11-8-11-8a18.45 18.45 0 015.06-5.94M9.9 4.24A9.12 9.12 0 0112 4c7 0 11 8 11 8a18.5 18.5 0 01-2.16 3.19m-6.72-1.07a3 3 0 11-4.24-4.24"/><line x1="1" y1="1" x2="23" y2="23"/></svg>
              Ce que le serveur ne voit <strong>jamais</strong>
            </div>
            <div class="vis-items">
              <div class="vis-item never-item">🚫 Votre mot de passe</div>
              <div class="vis-item never-item">🚫 La KEK dérivée</div>
              <div class="vis-item never-item">🚫 La Master Key en clair</div>
              <div class="vis-item never-item">🚫 La clé privée RSA en clair</div>
              <div class="vis-item never-item">🚫 Le code de récupération</div>
              <div class="vis-item never-item">🚫 Les clés de dossier / fichier en clair</div>
              <div class="vis-item never-item">🚫 Le contenu de vos fichiers</div>
              <div class="vis-item never-item">🚫 Les noms de fichiers (si chiffrement activé)</div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Footer CTA -->
    <section class="sec-cta">
      <div class="container">
        <h2>Une architecture auditée, un code ouvert</h2>
        <p>Le code source complet est disponible sous licence AGPLv3. Tout peut être vérifié, audité, forké.</p>
        <div class="cta-row">
          <a href="https://github.com/Bunnntyyy/SaferCloud" target="_blank" rel="noopener" class="btn btn-primary">
            <svg viewBox="0 0 24 24" fill="currentColor" width="18"><path d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12"/></svg>
            Voir le code source
          </a>
          <router-link to="/dashboard" class="btn btn-secondary">Créer un compte</router-link>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import LandingNav from '../../components/landing/LandingNav.vue'

// ── TOC ────────────────────────────────────────────────────────────
const tocNavOpen = ref(true)
const activeSection = ref('')

const tocItems = [
  { id: 'principe',   num: '01', label: 'Le principe' },
  { id: 'hierarchie', num: '02', label: 'Hiérarchie des clés' },
  { id: 'inscription',num: '03', label: 'Flux inscription' },
  { id: 'upload',     num: '04', label: 'Chiffrement fichier' },
  { id: 'partage',    num: '05', label: 'Partage RSA' },
  { id: 'session',    num: '06', label: 'Protection session' },
  { id: 'serveur',    num: '07', label: 'Ce que voit le serveur' },
]

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
  tocItems.forEach(item => {
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

const registrationSteps = [
  { title: 'Dérivation de clé (KEK)', desc: 'Votre mot de passe est passé à Argon2id dans un Web Worker. Le résultat (KEK) ne quitte jamais le navigateur.' },
  { title: 'Génération de la Master Key', desc: 'Une clé AES-256 aléatoire est générée via CSPRNG. Elle sera stockée dans le Service Worker avec extractable: false.' },
  { title: 'Wrapping de la Master Key', desc: 'La Master Key est chiffrée (wrapped) avec la KEK via AES-GCM. Ce chiffré est envoyé au serveur.' },
  { title: 'Paire de clés RSA', desc: 'Une paire RSA-OAEP 4096 bits est générée. La clé publique va au serveur. La clé privée est chiffrée par la Master Key.' },
  { title: 'Code de récupération', desc: '32 octets aléatoires. Un hash SHA-256 va au serveur pour vérification. La Master Key est aussi wrapped avec ce code.' },
  { title: 'Ce qui est envoyé au serveur', desc: 'Uniquement des données chiffrées ou publiques. Jamais votre mot de passe, la KEK ou la Master Key en clair.' },
]

const shareSteps = [
  { actor: 'alice', action: 'Récupère sa clé de fichier en mémoire', data: 'fileKey (AES-256) — non-extractable dans le SW' },
  { actor: 'server', action: 'Retourne la clé publique RSA de Bob', data: 'bob_public_key (RSA-OAEP 4096 bits, stockée en clair)' },
  { actor: 'alice', action: 'Chiffre la clé de fichier avec la clé publique de Bob', data: 'RSA-OAEP.encrypt(fileKey, bob_public_key) → encrypted_file_key_for_bob' },
  { actor: 'server', action: 'Stocke la clé chiffrée RSA', data: 'Ne peut pas la déchiffrer (pas de clé privée de Bob)' },
  { actor: 'bob', action: 'Déchiffre la clé de fichier avec sa clé privée RSA', data: 'RSA-OAEP.decrypt(encrypted_key, bob_private_key) → fileKey → déchiffre le fichier' },
]

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
  if (uploadPhase.value === 'idle') return '▶ Démarrer'
  if (uploadPhase.value === 'encrypting') return '→ Envoyer vers S3'
  if (uploadPhase.value === 'uploading') return uploadStep.value < chunks.value.length - 1 ? '→ Chunk suivant' : '→ Terminer'
  return '✓ Terminé'
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
