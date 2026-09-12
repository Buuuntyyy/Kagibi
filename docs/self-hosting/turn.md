# Serveur TURN (P2P)

## Est-ce que vous en avez besoin ?

Le transfert P2P de Kagibi utilise WebRTC : dans la majorité des cas, une connexion directe entre les deux appareils suffit, ou à défaut un simple serveur **STUN** (déjà configuré par défaut, gratuit, fourni par Google). Un serveur **TURN** — qui relaie tout le trafic au lieu de simplement aider à établir la connexion — n'est nécessaire que lorsque **les deux** appareils sont derrière un NAT strict ou symétrique que STUN ne peut pas traverser (fréquent sur certains réseaux d'entreprise, certains opérateurs mobiles, ou un double NAT/CGNAT chez le fournisseur d'accès).

**Ne déployez ce service que si vous avez déjà constaté des transferts P2P qui échouent à se connecter** — sinon, ce n'est pas nécessaire.

## Deux options

| | Effort | Quand |
|---|---|---|
| **A — Utiliser un TURN existant** | Minimal — juste 3 variables à renseigner | Vous avez déjà un serveur TURN (Twilio, un coturn déjà opéré ailleurs...) |
| **B — Déployer coturn** (cette page) | Un service Docker de plus, ports à ouvrir | Vous voulez tout héberger vous-même |

### Option A — TURN existant

Dans `.env` :
```bash
TURN_SERVER_URL=turn:turn.votre-domaine.example:3478
TURN_USERNAME=...
TURN_CREDENTIAL=...
```
Rien d'autre à faire — passez directement à la section [Vérifier](#vérifier) plus bas.

### Option B — Déployer coturn

**Linux uniquement.** `docker-compose.turn.yml` utilise `network_mode: host` — c'est l'approche que coturn documente lui-même pour Docker : un relais TURN alloue un port UDP par connexion active dans une large plage, ce que le mapping de ports habituel de Docker ne gère pas à cette échelle. Le mode `host` n'est pas fiablement supporté par Docker Desktop sur Mac/Windows ; ça tombe bien, un service qui doit relayer du vrai trafic a de toute façon sa place sur un serveur Linux dédié, pas un poste de développement.

#### 1. Renseigner `.env`

```bash
# Obligatoire — IP PUBLIQUE de ce serveur (pas une IP LAN)
TURN_EXTERNAL_IP=203.0.113.10

# Obligatoire — un seul couple d'identifiants partagé par tous les utilisateurs
TURN_USERNAME=...   # openssl rand -hex 24
TURN_CREDENTIAL=...  # openssl rand -hex 24

# Optionnel
# TURN_REALM=kagibi
# TURN_MIN_PORT=49160
# TURN_MAX_PORT=49200
```
Ne renseignez **pas** `TURN_SERVER_URL`/`TURN_USERNAME`/`TURN_CREDENTIAL` de l'option A en plus — `docker-compose.turn.yml` les dérive automatiquement pour le backend à partir des variables ci-dessus (mêmes identifiants utilisés à la fois par coturn et par le backend, un seul endroit à changer).

#### 2. Ouvrir le pare-feu

| Port | Protocole | Usage |
|---|---|---|
| 3478 | UDP + TCP | Signalisation TURN |
| `TURN_MIN_PORT`–`TURN_MAX_PORT` (défaut 49160–49200) | UDP | Relais effectif du trafic média |

Sans ça, coturn démarre normalement mais aucun client externe ne peut réellement l'atteindre.

#### 3. Lancer

```bash
docker compose -f docker-compose.yaml -f docker-compose.turn.yml up -d
```

## Durcissement inclus par défaut

Un relais TURN mal configuré peut devenir un **relais ouvert vers votre propre réseau interne** (SSRF) — un client malveillant demande à relayer du trafic non pas vers un autre pair WebRTC, mais vers une IP interne de votre réseau, voire vers l'endpoint de métadonnées cloud (`169.254.169.254`, une façon classique de voler les identifiants IAM d'une VM cloud). `docker-compose.turn.yml` bloque déjà ça par défaut :

- **Plages IP privées/spéciales interdites en relais** (`--denied-peer-ip`) : RFC1918 (`10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`), loopback, link-local (`169.254.0.0/16`, donc les métadonnées cloud), CGNAT (`100.64.0.0/10`), et quelques autres plages spéciales. Le loopback est de toute façon refusé par coturn par défaut.
- **`--no-software-attribute`** : masque la version de coturn dans ses réponses (évite de faciliter le ciblage d'une CVE connue pour une version précise).
- **`--no-cli`** : désactive l'interface d'administration telnet de coturn (jamais nécessaire pour cet usage).
- **Quotas** (`TURN_USER_QUOTA`/`TURN_TOTAL_QUOTA`, défaut 50) : limite le nombre de sessions relayées simultanées. Important à comprendre : Kagibi utilise un **seul couple d'identifiants partagé par tous les utilisateurs** (pas de credentials temporaires par session) — avec `--lt-cred-mech`, ce quota "par utilisateur" est donc en pratique un plafond global sur ce relais. Il limite les dégâts si ces identifiants fuitent ou sont utilisés en dehors de Kagibi, sans empêcher un usage normal.

**À savoir sur les identifiants partagés** : `TURN_USERNAME`/`TURN_CREDENTIAL` sont envoyés au navigateur de tout utilisateur Kagibi authentifié (via `GET /ice-config`, visibles dans les outils de développement du navigateur) — ce ne sont pas des secrets aussi confidentiels qu'un mot de passe serveur. Traitez-les comme "accessibles à tout utilisateur connecté", pas comme un secret d'infrastructure, et régénérez-les périodiquement (`openssl rand -hex 24`, puis mettre à jour `.env` et relancer) si vous voulez limiter la fenêtre d'exposition.

## Vérifier

Depuis deux réseaux différents (ex. votre connexion domicile + le partage de connexion de votre téléphone, pour être sûr de passer par deux NAT distincts), lancez un transfert P2P dans l'interface Kagibi. S'il aboutit alors qu'il échouait avant, le TURN fonctionne.

Pour un test plus direct, [trickle-ice de webrtc.github.io](https://webrtc.github.io/samples/src/content/peerconnection/trickle-ice/) permet de renseigner votre serveur TURN (`turn:votre-ip:3478`, avec vos identifiants) et de vérifier qu'un candidat ICE de type `relay` est bien généré.

## Aller plus loin

`docker-compose.turn.yml` couvre le cas courant (relais mono-nœud, identifiants statiques partagés, durcissement de base). Pour toute configuration plus avancée — TLS/TURNS, IPv6, authentification par credentials temporaires (REST API), cluster multi-nœuds, tuning fin des performances — reportez-vous à la documentation officielle du projet coturn : **[github.com/coturn/coturn](https://github.com/coturn/coturn)** (README + [wiki](https://github.com/coturn/coturn/wiki)), qui référence l'intégralité des options du `turnserver`.

## Dépannage

| Symptôme | Cause probable |
|---|---|
| Aucun candidat `relay` généré (test trickle-ice) | Pare-feu bloquant le port 3478 ou la plage de relais — vérifier depuis l'extérieur du réseau, pas juste en local |
| coturn démarre puis s'arrête immédiatement | Une des variables obligatoires (`TURN_EXTERNAL_IP`, `TURN_USERNAME`, `TURN_CREDENTIAL`) est absente de `.env` — `docker compose logs turn` affiche l'erreur exacte |
| Ça fonctionne en local mais pas depuis l'extérieur | `TURN_EXTERNAL_IP` doit être l'IP **publique** du serveur, pas une IP LAN (ex. `192.168.x.x`) |
