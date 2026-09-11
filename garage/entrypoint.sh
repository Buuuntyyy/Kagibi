#!/bin/sh
# Copyright (C) 2025-2026  Buuuntyyy
# SPDX-License-Identifier: AGPL-3.0-or-later
#
# Runs the Garage server and its one-time bootstrap (layout/bucket/key) in the
# SAME container/process namespace. A previous version tried to do this from a
# separate "garage-init" container talking to this one over RPC, but that
# requires knowing the node's RPC identity before being able to contact it —
# and with db_engine=sqlite, Garage doesn't expose that identity as a readable
# file before the server itself computes it. Running everything locally avoids
# the whole problem: no --rpc-host is needed for commands issued from here.
set -e

CREDENTIALS_FILE="/credentials/s3.env"
GARAGE="garage -c /etc/garage.toml"
BUCKET="${GARAGE_BUCKET:-kagibi}"
KEY_NAME="${GARAGE_KEY_NAME:-kagibi-key}"
CAPACITY="${GARAGE_CAPACITY:-10G}"

$GARAGE server &
GARAGE_PID=$!

if [ -s "$CREDENTIALS_FILE" ]; then
    echo "[garage] Cluster déjà bootstrapé (credentials présentes) — rien à refaire."
else
    echo "[garage] Attente que le nœud local réponde..."
    until $GARAGE status >/dev/null 2>&1; do
        sleep 1
    done

    # "node id -q" prints "<pubkey>@host:port" — layout/bucket commands below want
    # just the bare pubkey.
    NODE_ID=$($GARAGE node id -q | cut -d'@' -f1)
    echo "[garage] Nœud prêt (id=$NODE_ID)."

    # Best-effort : si un layout a déjà été appliqué pour ce nœud (ex. le fichier
    # de credentials a été supprimé manuellement après un premier bootstrap),
    # éviter de retenter assign/apply, qui échouerait sur une version déjà
    # committée. En cas de doute, "garage layout show" fait foi.
    if $GARAGE layout show 2>/dev/null | grep -q "$NODE_ID"; then
        echo "[garage] Layout déjà assigné pour ce nœud, on passe à la suite."
    else
        echo "[garage] Assignation du layout (nœud unique, capacité $CAPACITY)..."
        $GARAGE layout assign -z dc1 -c "$CAPACITY" "$NODE_ID"
        $GARAGE layout apply --version 1
    fi

    echo "[garage] Création du bucket '$BUCKET'..."
    $GARAGE bucket create "$BUCKET" 2>/dev/null || echo "[garage] Bucket déjà existant, on continue."

    echo "[garage] Création de la clé '$KEY_NAME'..."
    KEY_OUTPUT=$($GARAGE key create "$KEY_NAME")
    ACCESS_KEY=$(echo "$KEY_OUTPUT" | grep -i "Key ID" | awk '{print $NF}')
    SECRET_KEY=$(echo "$KEY_OUTPUT" | grep -i "Secret key" | awk '{print $NF}')

    if [ -z "$ACCESS_KEY" ] || [ -z "$SECRET_KEY" ]; then
        echo "[garage] ERREUR: impossible d'extraire les clés de la sortie de 'garage key create':" >&2
        echo "$KEY_OUTPUT" >&2
        kill "$GARAGE_PID"
        exit 1
    fi

    $GARAGE bucket allow --read --write --owner "$BUCKET" --key "$KEY_NAME"

    cat > "$CREDENTIALS_FILE" <<EOF
S3_ACCESS_KEY=$ACCESS_KEY
S3_SECRET_KEY=$SECRET_KEY
EOF
    echo "[garage] Bootstrap terminé — credentials écrites dans $CREDENTIALS_FILE."
fi

wait "$GARAGE_PID"
