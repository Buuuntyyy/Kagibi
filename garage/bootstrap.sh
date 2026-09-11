#!/bin/sh
# Copyright (C) 2025-2026  Buuuntyyy
# SPDX-License-Identifier: AGPL-3.0-or-later
#
# Bootstrap idempotent d'un cluster Garage mono-nœud : attend que le nœud
# réponde, assigne le layout (obligatoire même à un seul nœud — Garage ne sert
# jamais l'API S3 sans layout appliqué), crée le bucket et une clé S3, puis
# écrit les identifiants pour le backend Kagibi. Ne fait rien si le fichier de
# credentials existe déjà (cluster déjà bootstrapé — évite de retenter un
# "layout apply" sur une version déjà committée, qui échouerait).
set -e

CREDENTIALS_FILE="/credentials/s3.env"
NODE_KEY_FILE="/var/lib/garage/meta/node_key.pub"
BUCKET="${GARAGE_BUCKET:-kagibi}"
KEY_NAME="${GARAGE_KEY_NAME:-kagibi-key}"
CAPACITY="${GARAGE_CAPACITY:-10G}"

if [ -s "$CREDENTIALS_FILE" ]; then
    echo "[garage-init] Cluster déjà bootstrapé (credentials présentes) — rien à faire."
    exit 0
fi

echo "[garage-init] Attente que le nœud Garage écrive sa clé..."
until [ -s "$NODE_KEY_FILE" ]; do
    sleep 1
done

# --rpc-host exige "<node-id>@host:port" (pas juste "host:port") — l'identifiant
# du nœud EST sa clé publique RPC, dérivable directement du fichier sans passer
# par une commande RPC (qui échouerait justement tant que --rpc-host est invalide).
# xxd n'est pas dans Alpine/busybox par défaut (fourni par vim) — od l'est.
NODE_ID=$(od -An -tx1 "$NODE_KEY_FILE" | tr -d ' \n')
GARAGE="garage -c /etc/garage.toml --rpc-host $NODE_ID@garage:3901"

echo "[garage-init] Attente que le nœud Garage réponde (id=$NODE_ID)..."
until $GARAGE status >/dev/null 2>&1; do
    sleep 2
done
echo "[garage-init] Nœud prêt."

# Best-effort : si un layout a déjà été appliqué pour ce nœud (ex. le fichier
# de credentials a été supprimé manuellement après un premier bootstrap),
# éviter de retenter assign/apply, qui échouerait sur une version déjà
# committée. En cas de doute, "garage layout show" fait foi.
if $GARAGE layout show 2>/dev/null | grep -q "$NODE_ID"; then
    echo "[garage-init] Layout déjà assigné pour ce nœud, on passe à la suite."
else
    echo "[garage-init] Assignation du layout (nœud unique, capacité $CAPACITY)..."
    $GARAGE layout assign -z dc1 -c "$CAPACITY" "$NODE_ID"
    $GARAGE layout apply --version 1
fi

echo "[garage-init] Création du bucket '$BUCKET'..."
$GARAGE bucket create "$BUCKET" 2>/dev/null || echo "[garage-init] Bucket déjà existant, on continue."

echo "[garage-init] Création de la clé '$KEY_NAME'..."
KEY_OUTPUT=$($GARAGE key create "$KEY_NAME")
ACCESS_KEY=$(echo "$KEY_OUTPUT" | grep -i "Key ID" | awk '{print $NF}')
SECRET_KEY=$(echo "$KEY_OUTPUT" | grep -i "Secret key" | awk '{print $NF}')

if [ -z "$ACCESS_KEY" ] || [ -z "$SECRET_KEY" ]; then
    echo "[garage-init] ERREUR: impossible d'extraire les clés de la sortie de 'garage key create':" >&2
    echo "$KEY_OUTPUT" >&2
    exit 1
fi

$GARAGE bucket allow --read --write --owner "$BUCKET" --key "$KEY_NAME"

cat > "$CREDENTIALS_FILE" <<EOF
S3_ACCESS_KEY=$ACCESS_KEY
S3_SECRET_KEY=$SECRET_KEY
EOF

echo "[garage-init] Terminé — credentials écrites dans $CREDENTIALS_FILE."
