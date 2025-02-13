#!/bin/bash
set -e

# Load shell variables
. ./network/hermes/variables.sh

### Configure the clients and connection
echo "Initiating connection handshake..."
$HERMES_BINARY --config $CONFIG_DIR create connection --a-chain test-1 --b-chain test-2

sleep 2
$HERMES_BINARY --config $CONFIG_DIR create channel --a-chain test-1 --a-connection connection-0 --a-port transfer --b-port transfer